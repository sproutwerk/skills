# IPC contracts and lease lifetime

The provider registers `WrathCombo.<method>` through EzIPC. These are in-process
Dalamud gates, not HTTP endpoints. The consumer helper is the static
`WrathCombo.API.WrathIPCWrapper`; initialize it with the plugin interface before
its nested `RawMethod` fields are touched. ECommons discovery is a fallback in
`Wrapper.cs`, but explicit initialization avoids relying on reflection.

Source: [provider](source/WrathCombo/WrathCombo/Services/IPC/Provider.cs),
[wrapper signatures](source/WrathCombo.API/IPC/Settings.cs),
[raw gate types](source/WrathCombo.API/IPC/RawMethods/Settings.cs), and
[exception handling](source/WrathCombo.API/IPC/_Helper.cs).

## Core gates

Types below are provider signatures. Include every argument when using raw
Dalamud subscribers; C# optional arguments are a wrapper convenience.

| Gate suffix | Arguments | Result |
| --- | --- | --- |
| `IPCReady` | none | `bool` |
| `RegisterForLease` | `string internalPluginName, string pluginName` | `Guid?` |
| `RegisterForLeaseWithCallback` | same names, `string? ipcPrefixForCallback` | `Guid?` |
| `GetAutoRotationState` | none | `bool` |
| `SetAutoRotationState` | `Guid lease, bool enabled` | `SetResult` |
| `IsCurrentJobAutoRotationReady` | none | `bool` |
| `SetCurrentJobAutoRotationReady` | `Guid lease` | `SetResult` |
| `ReleaseControl` | `Guid lease` | `void` |

`internalPluginName` must match the loaded plugin's internal name: Wrath checks
loaded plugins periodically, around every 500 framework frames, and removes
leases whose owner is absent. For scripts, identify the actual hosting plugin
and coordinate one lease manager there; inventing a plugin name defeats the
owner check. Registration can return null when service is disabled or the owner
is temporarily blacklisted. Cache the returned ID. At this pin the deduplication
code compares an existing display `PluginName` with the incoming internal name,
so repeated registration is not a dependable idempotent operation.

For callbacks, register a callable action at
`<prefix>.WrathComboCallback(int reason, string additionalInfo)`. Passing null
as the callback prefix selects `internalPluginName`. The provider's overload
accepting `Action<int,string>` has no `[EzIPC]` attribute; it is not the public
callback gate. Use `RegisterForLeaseWithCallback` across plugin boundaries.

`IPCReady` becomes true after player-dependent cache construction. It does not
check the remotely controlled service switch. The setter checks that switch
and lease validity separately. A missing provider may throw even when the plugin
was present during an earlier check.

## Results and failures

Use the API's [SetResult enum](source/WrathCombo.API/Enum/SetResult.cs).

| Value | Meaning and consumer response |
| --- | --- |
| `IGNORED = -1` | Wrapper fallback after a suppressed failure; stop dependent work. |
| `Okay = 0` | Override accepted; observe effective state where needed. |
| `OkayWorking = 1` | Work accepted with deferred visibility; wait for the requested state with a finite deadline. |
| `IPCDisabled = 10` | Service disabled; stop and report it. |
| `InvalidLease = 11` | Forget the invalid ID; investigate cancellation before a new explicit start. |
| `BlacklistedLease = 12` | Respect user cancellation; do not evade it with another name or lease. |
| `Duplicate = 13` | This lease already has that registration/value; re-read state rather than repeatedly writing. |
| `PlayerNotAvailable = 14` | Defer until the player is available within the operation's deadline. |
| `InvalidConfiguration = 15` | Inspect the name, option, and target revision. |
| `InvalidValue = 16` | Inspect the option's declared type and value. |

A successful setter is not evidence that an action executed. Job setup can
return `OkayWorking` without propagating each internal combo/option result.
Stop on an unexpected result and re-read state before considering a retry.

The wrapper throws `UninitializedException` for a missing interface or provider,
`APIBehindException` for type/null/length mismatch, `IPCException` for other
Dalamud IPC errors, and `UnexpectedException` for other exceptions.
`WrathIPCWrapper.ErrorType` flags can suppress those groups, but the resulting
fallback cannot distinguish unavailable state from an ordinary false value.
The name `APIBehindException` alone does not prove which side is newer.

## Cancellation and coexistence

The [lease implementation](source/WrathCombo/WrathCombo/Services/IPC/Leasing.cs)
and [cancellation enum](source/WrathCombo.API/Enum/CancellationReason.cs) define:

| Value | Reason |
| --- | --- |
| `0` | `WrathUserManuallyCancelled` |
| `1` | `LeaseePluginDisabled` |
| `2` | `WrathPluginDisabled` |
| `3` | `LeaseeReleased` |
| `4` | `AllServicesSuspended` |
| `5` | `JobChanged` |

Invalidate the ID and queued work for every reason, including unknown future
values. Do not throw from the callback. `RemoveRegistration` calls it before
removing the registration, so recursive release or immediate registration can
observe stale state. Schedule any later, authorized restart after the callback
returns and job/readiness checks pass. Manual cancellation should stop the
consumer until the user chooses to start it again; the two-minute release-build
blacklist is not a retry timer. Debug builds skip adding that manual blacklist.

Multiple leases can coexist. For overlapping auto-rotation state, combo, option,
and auto-rotation configuration values, the provider selects by descending
lease-wide `LastUpdated`. Updating one setting can affect the priority of other
settings held by that lease when caches are refreshed. There is no exclusive
lock or transactional batch. Reasserting settings every frame can compete with
other clients. Use a narrow lease and explain conflicts in the consumer UI.

To finish, call `ReleaseControl` for your ID and clear local pending work. Do
not write a saved snapshot of global settings back: another client or the user
may have changed them. Release may reveal another lease or the saved enabled
state. Setting false and immediately releasing is therefore not a global stop.

## Consumer example

This example enables rotation only for an already configured job. It deliberately
does not call the broad job-setup setter. Add it to an existing plugin with the
matching WrathCombo.API dependency. Call `StartRequested` only from an explicit
user-authorized workflow, on the framework thread. Use one instance and one
stable callback prefix per host integration.

```csharp
using System;
using Dalamud.Plugin;
using Dalamud.Plugin.Ipc;
using WrathCombo.API;
using WrathCombo.API.Enum;

public sealed class WrathSession : IDisposable
{
    private readonly string internalName;
    private readonly string callbackPrefix;
    private readonly Action<string> log;
    private readonly ICallGateProvider<int, string, object> callback;
    private Guid? lease;
    private Guid? releasePending;
    private bool disposed;

    public bool Active { get; private set; }

    public WrathSession(IDalamudPluginInterface pi, Action<string> log)
    {
        this.log = log; // Supply a non-throwing logger.
        internalName = pi.InternalName;
        callbackPrefix = internalName + ".WrathSession";
        WrathIPCWrapper.Init(pi); // Keep the default exception reporting.
        callback = pi.GetIpcProvider<int, string, object>(
            callbackPrefix + ".WrathComboCallback");
        callback.RegisterAction(OnCancelled);
    }

    public bool StartRequested()
    {
        if (disposed || releasePending.HasValue) return false;
        if (lease.HasValue) return Active;
        try
        {
            if (!WrathIPCWrapper.IPCReady() ||
                !WrathIPCWrapper.IsCurrentJobAutoRotationReady()) return false;
            lease = WrathIPCWrapper.RegisterForLeaseWithCallback(
                internalName, internalName, callbackPrefix);
            if (lease is not Guid id) return false;
            var result = WrathIPCWrapper.SetAutoRotationState(id, true);
            if ((result == SetResult.Okay || result == SetResult.Duplicate) &&
                lease == id && WrathIPCWrapper.GetAutoRotationState())
                return Active = true;
            log($"Wrath start not confirmed: {result}");
        }
        catch (Exception ex) { log($"Wrath start failed: {ex.Message}"); }
        Stop();
        return false;
    }

    private void OnCancelled(int reason, string info)
    {
        Active = false;
        lease = null;
        releasePending = null;
        // Also cancel the host's dependent work; do not call Wrath here.
        log($"Wrath lease cancelled: {(CancellationReason)reason}; {info}");
    }

    public void Stop()
    {
        Active = false;
        releasePending ??= lease;
        lease = null; // Clear before the synchronous cancellation callback.
        if (releasePending is not Guid id) return;
        try
        {
            WrathIPCWrapper.ReleaseControl(id);
            releasePending = null;
        }
        catch (Exception ex) { log($"Wrath release unconfirmed: {ex.Message}"); }
    }

    public void Dispose()
    {
        if (disposed) return;
        disposed = true;
        try { Stop(); }
        finally { callback.UnregisterAction(); }
    }
}
```

This is source-derived code, not a compiled or live-tested consumer. `Active`
means this session's start was confirmed; competing clients and user edits can
change the effective state later. The host must stop its dependent queue on
cancellation, job change, timeout, and unload. If release fails, `releasePending`
prevents a new start; retry that specific release only after checking provider
availability. During teardown, report an unconfirmed release instead of assuming
it completed. The wrapper has no general `Dispose` method at this pin; release
leases and unregister your own callbacks/subscriptions explicitly.

# IPC contracts and examples

The contracts below describe the pins in [provenance](provenance.md).
Start with the [wrapper](source/AutoRetainerAPI/AutoRetainerAPI/AutoRetainerApi.cs),
[event dispatch](source/AutoRetainerAPI/AutoRetainerAPI/AutoRetainerApiInternal.cs),
and [legacy provider](source/AutoRetainer/AutoRetainer/Modules/IPC.cs).
For additional controls, read the
[PluginState provider](source/AutoRetainer/AutoRetainer/Modules/EzIPCManagers/IPC_PluginState.cs).

## Setup and lifetime

Initialize the consumer's ECommons services before constructing
`AutoRetainerAPI.AutoRetainerApi`. The constructor uses
`ECommons.DalamudServices.Svc.PluginInterface` and immediately subscribes to
events. Keep one instance for the integration's lifetime, attach named handlers,
then detach handlers and call `Dispose()` before disposing ECommons. Reuse the
consumer's existing ECommons ownership.

`new AutoRetainerApi(additionalSuffix: "VenturePlanner")` identifies this
consumer as the plugin's `InternalName` plus `_VenturePlanner`. The suffix
distinguishes multiple consumers in the same plugin, including hosted scripts.
Keep it stable and unique among those consumers. Without a suffix, identity is
just `InternalName`.

`Ready` invokes `AutoRetainer.Init` and catches exceptions. It proves the legacy
provider is registered, not that the player is logged in or every newer endpoint
exists. Handle provider disappearance between readiness and a subsequent call.
Dispose only unsubscribes the wrapper; it does not finish an outstanding
postprocessing request.

## Legacy channels

For raw Dalamud subscribers, the last generic argument is the return type.
Actions use an `object` placeholder and `InvokeAction`; functions use
`InvokeFunc`. Prefer the wrapper for its data lifetime checks and event filtering.

| Channel suffix after `AutoRetainer.` | Parameters | Result or effect |
| --- | --- | --- |
| `Init` | none | No-op action used as a readiness probe |
| `GetSuppressed` / `SetSuppressed` | none / `bool` | Read/write the shared suppression flag |
| `GetMultiModeEnabled` / `SetMultiModeEnabled` | none / `bool` | Read/write MultiMode; setter calls `OnMultiModeEnabled()` |
| `SetVenture` | `uint` | Set the current venture override |
| `GetRegisteredCIDs` | none | `List<ulong>` excluding blacklisted and uninitialized characters |
| `GetOfflineCharacterData` | `ulong cid` | `OfflineCharacterData`, possibly null |
| `WriteOfflineCharacterData` | `OfflineCharacterData` | Copy public fields into an existing record, or add a record |
| `GetAdditionalRetainerData` | `ulong cid, string name` | `AdditionalRetainerData`; provider can create a missing entry |
| `WriteAdditionalRetainerData` | `ulong cid, string name, AdditionalRetainerData` | Copy public fields into that entry |

The wrapper method for `GetRegisteredCIDs` is `GetRegisteredCharacters()`.
It has no `GetMultiModeEnabled()` method at this pin; use the raw channel or
PluginState. Keep spelling and capitalization from
[ApiConsts](source/AutoRetainerAPI/AutoRetainerAPI/ApiConsts.cs) when subscribing
to events; several event names differ from the wrapper's C# names.

## Venture override

`SchedulerMain.Tick()` resets `VentureOverride` to zero, fires
`OnSendRetainerToVenture`, then consumes a positive override before checking the
configured planner. Call `SetVenture(uint)` synchronously inside that callback.
Setting it before the event will be overwritten; awaiting work in the callback
misses the decision point. Resolve the venture ID before the callback.

Example helper owned by an existing consumer:

```csharp
public sealed class VentureOverride : System.IDisposable
{
    private readonly AutoRetainerAPI.AutoRetainerApi api;
    private readonly ulong targetCid;
    private readonly string targetRetainer;
    private readonly uint targetVentureId;

    public VentureOverride(ulong cid, string retainer, uint ventureId)
    {
        targetCid = cid;
        targetRetainer = retainer;
        targetVentureId = ventureId;
        api = new("VentureOverride");
        api.OnSendRetainerToVenture += OnSend;
    }

    private void OnSend(string retainerName)
    {
        if (ECommons.DalamudServices.Svc.ClientState.LocalContentId == targetCid
            && retainerName == targetRetainer && targetVentureId != 0)
            api.SetVenture(targetVentureId);
    }

    public void Dispose()
    {
        api.OnSendRetainerToVenture -= OnSend;
        api.Dispose();
    }
}
```

Construct this helper after ECommons initialization, passing the target CID,
retainer name, and a verified `RetainerTask` row ID from the user's configuration.
Dispose it before the consumer disposes ECommons. This example does not enable
scheduling. It checks both CID and name because retainer names alone are not a
cross-character key. The example was checked against source, not compiled.

## Postprocessing handshake

Retainer processing runs while the retainer is selected. Character processing
has a separate request and lock. The wrapper filters ready events to its own
consumer identity, so use that identity consistently.

| Stage | Retainer wrapper member | Character wrapper member |
| --- | --- | --- |
| Opportunity to request work | `OnRetainerPostprocessStep(string)` | `OnCharacterPostprocessStep()` |
| Request in that callback | `RequestRetainerPostprocess()` | `RequestCharacterPostprocess()` |
| Start when granted | `OnRetainerReadyToPostprocess(string)` | `OnCharacterReadyToPostProcess()` |
| Release after completion | `FinishRetainerPostProcess()` | `FinishCharacterPostProcess()` |

Request once during the opportunity event only when there is work to perform.
Duplicate requests from the same identity throw. Do not begin UI actions until
the ready event grants the turn. At that point, record ownership locally and
start a bounded task sequence that restores the entry UI state before finishing.

Call the finish method once when your sequence completes or successfully
recovers from failure. For asynchronous work, finish in its completion path,
not in a `finally` around the enqueue operation. Prevent late callbacks from an
aborted run from finishing a later run's lock. The finish endpoints take no
owner token and simply clear a shared boolean, so they cannot enforce ownership
for you.

The scheduler's wait uses `int.MaxValue` milliseconds instead of the normal
20-second task limit. A swallowed callback exception or missing finish can
therefore leave it waiting for a long time. Log the failure, stop your own queue,
and restore the expected UI before releasing your turn. If the state cannot be
restored, report the stalled operation and use the user's recovery procedure;
do not signal a successful completion or blindly resume actions.

Provider implementations:
[retainer tasks](source/AutoRetainer/AutoRetainer/Scheduler/Tasks/TaskPostprocessRetainerIPC.cs)
and [character tasks](source/AutoRetainer/AutoRetainer/Scheduler/Tasks/TaskPostprocessCharacterIPC.cs).
`ProcessIPCTaskFromOverlay()` belongs only inside
`OnRetainerListTaskButtonsDraw`; it requests that this consumer's tasks be run
through the retainer list. It is not a general background scheduler API.

## PluginState channels

Prefix these method names with `AutoRetainer.PluginState.`. Names come from
the provider's explicit EzIPC prefix and method names. Raw subscribers work
without initializing a second EzIPC owner in the consumer.

| Method | Signature | Behavior |
| --- | --- | --- |
| `IsBusy` | `() -> bool` | Main queue, AutoGCHandin operation, or Lifestream busy |
| `GetEnabledRetainers` | `() -> Dictionary<ulong, HashSet<string>>` | Selected names by character content ID |
| `AreAnyRetainersAvailableForCurrentChara` | `() -> bool` | Uses AutoRetainer's current-character availability logic |
| `GetInventoryFreeSlotCount` | `() -> int` | Plugin's inventory slot calculation |
| `GetMultiModeStatus` | `() -> bool` | Configured enabled flag, including while suppressed |
| `GetClosestRetainerVentureSecondsRemaining` | `(ulong CID) -> long?` | Earliest qualifying selected retainer; null when unavailable |
| `AreAnyEnabledVesselsNotDeployed` | `(ulong contentId) -> bool?` | Null for an unknown character |
| `AreAnyEnabledVesselsReady` | `(ulong contentId) -> bool?` | Uses workshop advance timer and wait-for-all settings; null for unknown character |
| `IsItemProtected` | `(uint itemId) -> bool` | Membership in the current inventory-management protection list |
| `AbortAllTasks` | `() -> void` | Abort the main task manager |
| `DisableAllFunctions` | `() -> void` | Disable MultiMode, retainer scheduler, and voyage scheduler |
| `EnableMultiMode` | `() -> void` | Dispatch `/autoretainer multi enable` |
| `EnableSingleMultiMode` | `(MultiModeType? multiModeType) -> void` | Enable MultiMode and initialize the single-pass visited-character set |
| `EnableMultiModeSingle` | `() -> void` | Dispatch `/autoretainer multi single`; see the command caveat below |
| `CanAutoLogin` | `() -> bool` | Check the plugin's auto-login prerequisites |
| `Relog` | `(string charaNameWithWorld) -> bool` | Find a cached `Name@World` and request a relog; true does not confirm login completion |

`EnableSingleMultiMode` and `EnableMultiModeSingle` are separate methods.
At this pin, the command used by the latter only enables MultiMode if disabled;
it does not initialize `SingleMultiMode`. The former initializes that set and
marks the current character when logged in. Read the target implementation
before promising a single pass.

The provider also exposes RetainerSense option getters/setters, `EnqueueHET`
with an `Action onFailure`, and `OnSingleMultiModeFinished`. Use the full provider
file for those signatures. The separate `AutoRetainer.GC` provider marks
`GetGCInfo` obsolete; do not build a new consumer around it without checking
the target version's replacement.

## Failures and shared state

For a missing provider, skip the requested integration and report unavailable
state; do not repeatedly enable AutoRetainer to probe it. For a type mismatch,
compare the exact API assembly and DTO definitions with the target plugin.
Keep null timers distinct from zero and negative values. Remaining venture
seconds can be negative when a venture is already complete.

`Suppressed` is one shared boolean, with no owner or reference count. Coordinate
temporary changes with the consumer's lifetime and other integrations. It does
not abort queued tasks or cover every subsystem; the API summary's broad wording
should not be treated as a universal emergency stop.

# EzIPC and EzSharedData

Covers: attribute-driven Dalamud IPC (`EzIPC`, `[EzIPC]`, `[EzIPCEvent]`, `SafeWrapper`,
`EzIPCExtensions.TryInvoke`, `EzIPCDisposalToken`), the separate `ECommons.IPC` package as far as it
is visible, and `EzSharedData` over Dalamud DataShare. Verified against ECommons `bc38dc7`
(2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are relative to the skill
root.

Underlying Dalamud call gates (`ICallGateProvider`/`ICallGateSubscriber`, `IpcError` types,
threading, which types may cross the plugin boundary) and DataShare are described in the
`dalamud` skill, `references/topic-ipc.md`; this file only covers the ECommons layer.

```csharp
using ECommons.EzIpcManager;        // EzIPC, EzIPCAttribute, EzIPCEventAttribute, SafeWrapper,
                                    // EzIPCExtensions, EzIPCDisposalToken
using ECommons.EzSharedDataManager; // EzSharedData, CreationMode
using Dalamud.Plugin.Ipc.Exceptions; // IpcNotReadyError etc., only if you catch them yourself
```

Source: `references/ecommons/ECommons/EzIpcManager/*.cs`,
`references/ecommons/ECommons/EzIpcManager/EzIPC.md` (upstream doc),
`references/ecommons/ECommons/EzSharedDataManager/*.cs`.

Requirements: `ECommonsMain.Init(...)` must have run first (everything goes through
`Svc.PluginInterface`). No `Module` flag is needed. No hooks are installed.

## What Init does, in one table

`EzIPC.Init` reflects over one class and performs four passes, in this order. The member kind
decides the role; the attribute decides function-vs-event.

| Attribute | On a | Role | Dalamud call made | Token returned |
|---|---|---|---|---|
| `[EzIPC]` | method | provider of a callable | `GetIpcProvider<...>(name).RegisterFunc(d)` or `.RegisterAction(d)` | yes, `IsEvent == false` |
| `[EzIPC]` | delegate field/property | subscriber (consumer): member is assigned a delegate that calls the other plugin | `GetIpcSubscriber<...>(name)`, member = its `InvokeFunc` / `InvokeAction` (optionally wrapped) | no |
| `[EzIPCEvent]` | method | event subscription (consumer) | `GetIpcSubscriber<...>(name).Subscribe(d)` | yes, `IsEvent == true` |
| `[EzIPCEvent]` | delegate field | event sender (provider): member is assigned a delegate that fires the event | `GetIpcProvider<...>(name)`, member = its `SendMessage` | no |

Rule to remember: a METHOD is something other plugins call into (you implement it); a
FIELD/PROPERTY of delegate type is something you call out through (ECommons fills it in).

## EzIPC (static class)

```csharp
public static event Action<Exception>? OnSafeInvocationException;

public static EzIPCDisposalToken[] Init(object instance, string? prefix = null,
    SafeWrapper safeWrapper = SafeWrapper.None);
public static EzIPCDisposalToken[] Init(Type staticType, string? prefix = null,
    SafeWrapper safeWrapper = SafeWrapper.None);

public static MethodInfo? FindIpcProvider(int numGenericArgs);
public static MethodInfo? FindIpcSubscriber(int numGenericArgs);

internal static void Dispose();   // called by ECommonsMain.Dispose()
```

- `Init(object instance, ...)`: scans `instance.GetType()` with
  `BindingFlags.Public | NonPublic | Instance`. Static members of that class are ignored.
- `Init(Type staticType, ...)`: scans with `Public | NonPublic | Static`. Use `typeof(MyStaticIpc)`.
  Instance members are ignored. (AutoRetainer: `EzIPC.Init(typeof(IPC));`.)
- Private members are included in both cases, so `[EzIPCEvent] void OnX()` without an access
  modifier works.
- `prefix`: `null` becomes `Svc.PluginInterface.InternalName` (your own plugin). When consuming
  another plugin's IPC you must pass that plugin's prefix explicitly, e.g.
  `EzIPC.Init(this, "AutoRetainer.PluginState")`.
- `safeWrapper`: default wrapper for `[EzIPC]` subscriber delegates of this class whose attribute
  has `wrapper: SafeWrapper.Inherit` (the attribute default). Passing `SafeWrapper.Inherit` here
  throws `InvalidOperationException`.
- There is no `reducedLogging` parameter. Logging is controlled by the global
  `ECommonsMain.ReducedLogging` (`public static bool`, default `false`); when `false`, each
  registration/unregistration writes `PluginLog.Debug` lines tagged `[EzIPC Provider]` /
  `[EzIPC Subscriber]`.
- Return value: one `EzIPCDisposalToken` per registered provider method and per event
  subscription method. Normal use discards it. Keep it only to unregister before plugin unload.
- Error handling: every member is processed in its own `try`/`catch`. A failure logs
  `PluginLog.Error("[EzIPC ...] Failed to ...")` plus the exception and continues with the next
  member; `Init` itself does not throw for a bad member. A subscriber field that failed stays
  `null` and throws `NullReferenceException` when invoked. A member with an attribute that is not a
  delegate type logs `PluginLog.Warning("Not a delegate: <name>")`.
- A `null` `instance` fails with `NullReferenceException` from `instance.GetType()` (the XML doc
  says `ArgumentNullException`; the code has no such check).
- `FindIpcProvider(n)` / `FindIpcSubscriber(n)`: reflection lookup of the generic
  `GetIpcProvider` / `GetIpcSubscriber` overload on `Svc.PluginInterface` with exactly `n` generic
  arguments (parameters + 1 for `TRet`); `null` if none. Public mostly for internal use;
  AutoRetainer uses `EzIPC.FindIpcProvider(1).MakeGenericMethod(fieldType)` to register
  `Func<T>` getters for config fields at runtime.

### Disposal

`ECommonsMain.Dispose()` calls `EzSharedData.Dispose` and then `EzIPC.Dispose` (each inside
`GenericHelpers.Safe`). `EzIPC.Dispose`:

1. calls `Dispose()` on every token ever created by any `Init` call: `UnregisterFunc` /
   `UnregisterAction` for providers, `Unsubscribe(sameDelegate)` for event subscriptions;
2. clears the token list;
3. sets `OnSafeInvocationException = null`.

So a plugin that calls `ECommonsMain.Dispose()` needs no IPC cleanup code. Subscriber delegates and
event-sender delegates hold nothing registered and need no cleanup.

### EzIPCDisposalToken

```csharp
public sealed class EzIPCDisposalToken
{
    public readonly string IpcTag;      // full channel name, prefix included
    public readonly bool IsEvent;       // true = event subscription, false = provider func/action
    public bool IsDisposed { get; private set; }
    public void Dispose();              // idempotent; exceptions are caught and logged
}
```

The constructor is internal. It does not implement `IDisposable` (no `using`). Disposing a token
early is safe; the later global dispose skips it because `IsDisposed` is already `true`.

## [EzIPC] attribute

```csharp
[AttributeUsage(AttributeTargets.Method | AttributeTargets.Field | AttributeTargets.Property,
    AllowMultiple = false)]
public class EzIPCAttribute : Attribute
{
    public string? IPCName;
    public bool ApplyPrefix;
    public Type ActionLastGenericType;
    public SafeWrapper Wrapper = SafeWrapper.Inherit;

    public EzIPCAttribute(string? iPCName = null, bool applyPrefix = true,
        Type? actionLastGenericType = null, SafeWrapper wrapper = SafeWrapper.Inherit);
}
```

| Parameter | Default | Meaning |
|---|---|---|
| `iPCName` | `null` | Channel name part. `null` = the method/field/property name. |
| `applyPrefix` | `true` | `true`: channel = `"{prefix}.{ipcName}"`. `false`: channel = `ipcName` verbatim. |
| `actionLastGenericType` | `null` -> `typeof(object)` | The dummy `TRet` generic argument used when the member returns `void`. |
| `wrapper` | `SafeWrapper.Inherit` | Only used for subscriber delegates (field/property). Ignored on provider methods. |

Note the parameter spelling `iPCName` (lower-case `i`, upper-case `PC`) when using named arguments;
`applyPrefix:false` is the named argument the upstream doc uses.

### Name derivation (identical in all four passes)

```
ipcName = attr.IPCName ?? member.Name
ipcName = ipcName.Replace("%m", member.Name)
ipcName = ipcName.Replace("%p", Svc.PluginInterface.InternalName)
channel = attr.ApplyPrefix ? $"{prefix}.{ipcName}" : ipcName
```

- `%m` = the annotated member's own name.
- `%p` = the CALLING plugin's `InternalName`, always, regardless of the `prefix` argument passed to
  `Init`. It is not "the prefix".
- The separator is a single `.` and is added by ECommons; do not end `prefix` with a dot.
- The placeholders are not mentioned in the upstream `EzIPC.md`; they exist in code.

### Provider methods: Func vs Action, arity

- `method.ReturnType == typeof(void)` -> registered with `RegisterAction`, generic arguments are
  `[parameter types..., attr.ActionLastGenericType]` (so `GetIpcProvider<T1, object>` by default).
- Any other return type -> `RegisterFunc`, generic arguments `[parameter types..., return type]`.
- The delegate is built by `ReflectionHelper.CreateDelegate(method, instance)`, which uses
  `Expression.GetActionType` / `Expression.GetFuncType`; keep parameters plain (no `ref`/`out`)
  *(by-ref parameter behaviour unverified)*.
- Maximum 8 parameters. Dalamud's largest overload is
  `GetIpcProvider<T1..T8, TRet>` (9 generic arguments). With more, `FindIpcProvider` returns
  `null` and the member fails with the logged message "... Did you called EzIPC.Init before
  ECommonsMain.Init or specified more than 9 arguments?" (the "9" counts generic arguments, not
  parameters).
- One channel name can carry both a Func and an Action in Dalamud, but two `[EzIPC]` methods
  resolving to the same channel and same kind overwrite each other. C# overloads with the same
  name therefore need distinct `iPCName`s.

### Subscriber delegates (field or property)

- The member type must derive from `Delegate` and have an `Invoke` method; in practice
  `Action`, `Action<...>`, `Func<...>`. Return type `void` -> bound to
  `ICallGateSubscriber<..., TLast>.InvokeAction`; otherwise bound to `InvokeFunc`.
- The kind must match the provider: a provider method returning `void` registers an Action, so the
  consumer member must be `Action<...>`; a provider returning a value needs `Func<..., TRet>`. A
  mismatch is reported by Dalamud as `IpcNotReadyError` at call time (or swallowed by a wrapper).
- Same 8-parameter limit.
- Properties need a setter reachable by reflection (`GetSetMethod(true)`, private is fine); a
  get-only auto-property fails with `InvalidOperationException("No accessible setter found ...")`,
  which is logged and the property stays `null`.
- The upstream doc declares instance fields as `public readonly Action ...`; assignment happens via
  reflection after construction. `static readonly` fields with the `Type` overload are
  *(unverified)*; use non-readonly static fields.
- Assignment happens at `Init` time and succeeds whether or not the provider plugin is loaded.
  Availability is only checked when the delegate is invoked.

## [EzIPCEvent] attribute

```csharp
[AttributeUsage(AttributeTargets.Method | AttributeTargets.Field, AllowMultiple = false)]
public class EzIPCEventAttribute : Attribute
{
    public string? IPCName;
    public bool ApplyPrefix;
    public Type ActionLastGenericType;

    public EzIPCEventAttribute(string? iPCName = null, bool applyPrefix = true,
        Type? actionLastGenericType = null);
}
```

No `wrapper` parameter; safe wrappers do not apply to events. Not valid on properties (compile
error from `AttributeUsage`), unlike `[EzIPC]`.

- On a METHOD: the method is subscribed to the channel via `Subscribe`. It must return `void`,
  otherwise `InvalidOperationException("Event method must have void return value")` is logged and
  the method is skipped. Generic arguments: `[parameter types..., ActionLastGenericType]`.
  Unsubscribed by the token / `ECommonsMain.Dispose()`.
- On a FIELD of type `Action` / `Action<...>`: the field is assigned the provider's `SendMessage`.
  Calling the field fires the event to all current subscribers. A delegate with a return value
  logs `InvalidOperationException("Only void-returning delegates may be registered as events")`.
- Name derivation and prefix rules are the same as for `[EzIPC]`. Event name = field name on the
  sender side and method name on the receiver side unless `iPCName` is given, so keep them equal
  or name the channel explicitly.
- Firing is synchronous on the caller's thread. Dalamud's `SendMessage` iterates subscriptions
  without a `try`/`catch`, so an exception in one subscriber propagates to the sender and the
  remaining subscribers are not called; wrap the call if that matters.

## SafeWrapper

```csharp
public enum SafeWrapper
{
    None,           // 0
    IPCException,   // 1
    AnyException,   // 2
    Inherit = 255
}
```

Applies only to `[EzIPC]` subscriber delegates. Effective wrapper =
`attr.Wrapper == Inherit ? Init's safeWrapper : attr.Wrapper`.

| Value | Delegate assigned to the member | On failure |
|---|---|---|
| `None` | The raw `InvokeAction` / `InvokeFunc` of the Dalamud subscriber. | Exceptions propagate: `IpcNotReadyError` when the provider is absent, other `IpcError`s, and whatever the provider threw. |
| `IPCException` | `SafeWrapperIPC.Wrapper<...>.InvokeAction/InvokeFunction` | Catches ONLY `IpcNotReadyError` (not the `IpcError` base class, despite the name). Other exceptions propagate. |
| `AnyException` | `SafeWrapperAny.Wrapper<...>.InvokeAction/InvokeFunction` | Catches every `Exception`. |
| `Inherit` | Valid only in the attribute. | In `Init` it throws `InvalidOperationException`. |

Behaviour common to both wrappers:

- Before calling, the wrapper checks the subscriber: `HasAction == false` (for actions) or
  `HasFunction == false` (for functions) returns immediately, `default` for functions. This path
  raises NO exception and does NOT fire `OnSafeInvocationException`. "Provider plugin not loaded"
  is therefore completely silent.
- A caught exception is passed to `EzIPC.OnSafeInvocationException` and then dropped. Nothing is
  logged by ECommons; subscribe to the event if you want logs. Actions return normally, functions
  return `default(TRet)`.
- Consequence: with a wrapper, `default` (`false`, `0`, `null`) is indistinguishable from a real
  result. Use `Func<..., bool?>`-style nullable results only if the provider declares them that
  way; otherwise use `SafeWrapper.None` plus `TryInvoke` when you need to know whether the call
  happened.

`SafeWrapperAny` and `SafeWrapperIPC` are `internal static` classes holding `Wrapper<TRet>` ...
`Wrapper<T1..T8, TRet>`; they are created by `EzIPC` through reflection and are not usable from
plugin code. The two files differ only in the caught exception type.

```csharp
// typical: log swallowed IPC failures once, globally (AutoRetainer does this right after Init)
ECommonsMain.Init(pi, this);
EzIPC.OnSafeInvocationException += e => PluginLog.Error(e.ToString());
```

The handler list is cleared by `ECommonsMain.Dispose()`.

## EzIPCExtensions (TryInvoke)

`public static class EzIPCExtensions`: extension methods on plain BCL delegates, 0 to 8 arguments.

```csharp
public static bool TryInvoke(this Action action);
public static bool TryInvoke<T1>(this Action<T1> action, T1 a1);
// ... up to Action<T1, ..., T8>
public static bool TryInvoke<TRet>(this Func<TRet> function, out TRet? ret);
public static bool TryInvoke<T1, TRet>(this Func<T1, TRet> function, T1 a1, out TRet? ret);
// ... up to Func<T1, ..., T8, TRet>; `out ret` is always the last argument
```

Invokes the delegate; returns `true` on success, `false` (and `ret = default`) if
`IpcNotReadyError` was thrown. Any other exception propagates. Intended for delegates obtained
with `SafeWrapper.None`. On a wrapped delegate it always returns `true`, because the wrapper
already swallowed the error. A `null` delegate throws `NullReferenceException`.

```csharp
if(IsBusy.TryInvoke(out var busy) && !busy) { /* provider present and idle */ }
```

## Worked example

Modelled on AutoRetainer (`AutoRetainer/Modules/EzIPCManagers/IPC_PluginState.cs`: instance class,
`EzIPC.Init(this, $"{Svc.PluginInterface.InternalName}.PluginState")`, `[EzIPC]` methods,
`[EzIPCEvent] public Action OnSingleMultiModeFinished;`) and YesAlready
(`YesAlready/IPC/YesAlreadyIPC.cs`: `public YesAlreadyIPC() => EzIPC.Init(this);`). PandorasBox at
`3a35634` has no EzIPC usage.

Provider plugin, `InternalName` = `Courier`:

```csharp
using ECommons.DalamudServices;
using ECommons.EzIpcManager;

public class CourierIpcProvider
{
    public CourierIpcProvider()
    {
        // prefix "Courier.State"; omit the argument to get plain "Courier"
        EzIPC.Init(this, $"{Svc.PluginInterface.InternalName}.State");
    }

    [EzIPC] public bool IsBusy() => busy;                         // Func<bool>
    [EzIPC] public void Abort() { /* ... */ }                     // Action
    [EzIPC("SetEnabled")] public void Toggle(bool on) { }         // Action<bool>
    [EzIPC("Courier.Legacy.%m", applyPrefix: false)]
    public int GetCount(uint itemId) => 0;                        // Func<uint, int>

    [EzIPCEvent] public Action<uint> OnDelivered;                 // event sender, filled by Init

    private bool busy;
    internal void Finish(uint itemId) => OnDelivered(itemId);     // fire
}
```

Created once after `ECommonsMain.Init` (e.g. `ipc = new CourierIpcProvider();`); nothing to do on
unload besides `ECommonsMain.Dispose()`.

Resulting channels:

| Member | Channel name | Dalamud registration |
|---|---|---|
| `IsBusy` | `Courier.State.IsBusy` | `GetIpcProvider<bool>(..).RegisterFunc` |
| `Abort` | `Courier.State.Abort` | `GetIpcProvider<object>(..).RegisterAction` |
| `Toggle` | `Courier.State.SetEnabled` | `GetIpcProvider<bool, object>(..).RegisterAction` |
| `GetCount` | `Courier.Legacy.GetCount` | `GetIpcProvider<uint, int>(..).RegisterFunc` |
| `OnDelivered` | `Courier.State.OnDelivered` | `GetIpcProvider<uint, object>(..)`, field = `SendMessage` |

Consumer class in a different plugin:

```csharp
using ECommons.EzIpcManager;

public class CourierIpcConsumer
{
    public CourierIpcConsumer()
    {
        // explicit prefix of the OTHER plugin; AnyException for every [EzIPC] delegate below
        EzIPC.Init(this, "Courier.State", SafeWrapper.AnyException);
    }

    [EzIPC] public Func<bool> IsBusy;                              // Courier.State.IsBusy
    [EzIPC] public Action Abort;                                   // Courier.State.Abort
    [EzIPC("SetEnabled")] public Action<bool> SetCourierEnabled;   // Courier.State.SetEnabled
    [EzIPC("Courier.Legacy.GetCount", applyPrefix: false, wrapper: SafeWrapper.None)]
    public Func<uint, int> GetCount;                               // unwrapped, may throw

    [EzIPCEvent]
    private void OnDelivered(uint itemId)                          // Courier.State.OnDelivered
    {
    }
}
```

Usage: `if(!ipc.IsBusy()) ipc.SetCourierEnabled(true);` (both silently no-op / return `false` if
Courier is not loaded) and `ipc.GetCount.TryInvoke(1234u, out var n)` for the unwrapped one.

Equivalent hand-written Dalamud code for the first consumer line:
`Svc.PluginInterface.GetIpcSubscriber<bool>("Courier.State.IsBusy").InvokeFunc()`. A plugin that
does not use ECommons can consume or provide the same channels with plain call gates; EzIPC adds
no wire format of its own. Real channel names following this scheme: `YesAlready.IsPluginEnabled`
(`Func<bool>`), `YesAlready.PausePlugin` (`Action<int>`), `AutoRetainer.PluginState.IsBusy`
(`Func<bool>`), `AutoRetainer.GC.EnqueueInitiation` (`Action`),
`AutoRetainer.PluginState.OnSingleMultiModeFinished` (event, no payload). These assume each
plugin's `InternalName` equals its name as written, which the hard-coded `"AutoRetainer.*"` strings
in AutoRetainer's `Modules/IPC.cs` support for AutoRetainer.

Several classes may call `EzIPC.Init` with different prefixes (AutoRetainer has one class per
prefix: `.PluginState`, `.GC`). One class may mix provider and consumer members, but they all
share the one `prefix`, so use `applyPrefix: false` with full names for the foreign ones.

## ECommons.IPC (separate package)

Not part of the ECommons repository. Repo `NightmareXIV/ECommons.IPC` (from AutoRetainer's
`.gitmodules`); AutoRetainer references it as a project:
`<ProjectReference Include="..\ECommons.IPC\ECommons.IPC\ECommons.IPC.csproj" />`. The submodule
directory is empty in the inspected clone, so nothing below was read from its source. It is also
published on nuget.org as `ECommons.IPC`; latest is 1.0.0.24, whose nuspec names commit `43cbcb0`
of that repo, targets `net10.0-windows7.0` and depends on `ECommons >= 3.2.0.11` (checked on
nuget.org, 2026-11-13).

What AutoRetainer's usage shows (names only; signatures *(unverified)*):

- Namespaces `ECommons.IPC`, `ECommons.IPC.Subscribers`, `ECommons.IPC.Subscribers.Vnavmesh`,
  `ECommons.IPC.Subscribers.LifestreamIPC`, `ECommons.IPC.Subscribers.AutoRetainer` (contains
  `MultiModeType`).
- A static class `ECommons.IPC.ECommonsIPC`, imported with
  `global using static ECommons.IPC.ECommonsIPC;`, exposing ready-made subscriber objects as
  members: `Lifestream`, `Vnavmesh`, `AutoRetainer`. Calls seen: `Lifestream.IsBusy()`,
  `Lifestream.Abort()`, `Lifestream.ExecuteCommand(string)`, `Lifestream.Logout()`,
  `Lifestream.ChangeCharacter(name, world)`, `Lifestream.EnqueueInnShortcut(null)`,
  `Lifestream.EnqueuePropertyShortcut(...)`, `Lifestream.GetHousePathData(cid)`,
  `Lifestream.HasPrivateHouse()` (compared with `!= false`, so nullable), `Vnavmesh.Stop()`,
  `Vnavmesh.Reload()`, `ECommonsIPC.AutoRetainer.EnableSingleMultiMode(null)`.
- `IPCBase.DefaultWrapper = SafeWrapper.AnyException;` set BEFORE `ECommonsMain.Init` in
  AutoRetainer's constructor. `IPCBase` is not defined in ECommons, so it belongs to this package;
  its exact namespace and semantics are *(unverified)*.

It is a catalogue of pre-written EzIPC subscriber classes for popular plugins. Read its source
before relying on any member.

## EzSharedData

`public static class EzSharedData` wraps Dalamud DataShare with a per-plugin cache and automatic
relinquish on unload.

```csharp
public static bool TryGet<T>(string Name, [NotNullWhen(true)] out T Data,
    CreationMode Mode = CreationMode.ReadOnly, T DefaultValue = null) where T : class;

public static T Get<T>(string Name, CreationMode Mode = CreationMode.ReadOnly,
    T DefaultValue = null) where T : class;

public static T GetOrCreate<T>(string Name, T DefaultValue = null) where T : class;
    // => Get(Name, CreationMode.CreateAndRelinquish, DefaultValue)

public enum CreationMode { ReadOnly, CreateAndKeep, CreateAndRelinquish, ReadAndKeep }
```

Parameter names are PascalCase (`Name`, `Data`, `Mode`, `DefaultValue`) if you use named arguments.

| Mode | Dalamud call | If tag absent | Relinquished by `ECommonsMain.Dispose()` |
|---|---|---|---|
| `ReadOnly` | `TryGetData<T>` | `TryGet` -> `false`; `Get` -> throws `SharedDataNotReadyException` | yes |
| `ReadAndKeep` | `TryGetData<T>` | `TryGet` -> `false`; `Get` -> throws `ArgumentOutOfRangeException` ALWAYS (mode not handled by `Get`) | no |
| `CreateAndRelinquish` | `GetOrCreateData<T>(Name, () => DefaultValue)` | created from `DefaultValue` | yes |
| `CreateAndKeep` | `GetOrCreateData<T>(Name, () => DefaultValue)` | created from `DefaultValue` | no |

Behaviour:

- Cache first: once a name was obtained, every later `TryGet`/`Get` for that name returns the cached
  reference `(T)cached` without touching Dalamud and ignores `Mode`/`DefaultValue`. A different `T`
  for the same name throws `InvalidCastException`. A failed lookup is not cached, so polling
  `TryGet` until another plugin creates the tag works.
- Because of the cache the plugin registers as a DataShare consumer only once per tag, no matter
  how often `TryGet` is called. It is still a dictionary lookup plus cast; hold the reference
  yourself in per-frame code (upstream advice).
- `T` must be a reference type from .NET, Dalamud or Dalamud's own libraries: no plugin-defined
  types, no types from NuGet packages or submodules. For a value use a one-element array; for
  structured data use tuples/collections. AutoRetainer:
  `EzSharedData.TryGet("AutoRetainer.Started", out TimeLaunched, CreationMode.CreateAndKeep,
  [DateTimeOffset.Now.ToUnixTimeMilliseconds()]);`
- `DefaultValue` must be non-null in the create modes when the tag does not exist yet: Dalamud
  throws `DataCacheCreationError` if the generator returns `null`. `TryGet` does not catch this.
  With create modes `TryGet` either returns `true` or throws (also `DataCacheTypeMismatchError` if
  the existing data has another type). With read modes a type mismatch just yields `false`.
- `SharedDataNotReadyException` is `internal` (message
  `"Shared data tag {name} is not ready yet"`); plugin code cannot name it in a `catch`. Prefer
  `TryGet`, or catch `Exception`.
- The shared object is the same instance in every plugin; mutation is visible to all and is not
  synchronized. Lock or use concurrent collections if touched off the framework thread.
- Name the tag `"<PluginName>.<Something>"` (upstream advice; names are global).

Dispose: `EzSharedData.Dispose()` (internal, called from `ECommonsMain.Dispose()` before
`EzIPC.Dispose`) calls `Svc.PluginInterface.RelinquishData(name)` for every cached name not in the
keep list. There is no public per-tag relinquish; call
`Svc.PluginInterface.RelinquishData(name)` yourself if needed (the EzSharedData cache entry stays,
so the cached reference keeps being returned). In Dalamud, when the last consumer relinquishes, the
cache entry is removed and the value is disposed if it is `IDisposable`.

"Keep" modes skip the relinquish, leaving the plugin registered as a consumer so the data outlives
the plugin instance. AutoRetainer uses this as a "was loaded earlier in this game session" marker:

```csharp
// at the point of first load
EzSharedData.TryGet<object>("AutoRetainer.WasLoaded", out _, CreationMode.CreateAndKeep, new());
// on startup: absent => fresh game start, present => plugin was reloaded
if(!EzSharedData.TryGet<object>("AutoRetainer.WasLoaded", out _)) { /* auto-start */ }
```

Consumer-side pattern (AutoRetainer pausing YesAlready/TextAdvance via a shared set owned by the
other plugin):

```csharp
if(EzSharedData.TryGet<HashSet<string>>("YesAlready.StopRequests", out var data))
{
    data.Add(Svc.PluginInterface.InternalName);   // remove it again when done
}
```

## Gotchas

- `EzIPC.Init` before `ECommonsMain.Init` fails: `Svc.PluginInterface` is not set.
- Forgetting to call `EzIPC.Init` at all leaves every delegate field `null`; attributes alone do
  nothing.
- Default prefix is YOUR plugin's `InternalName`. A consumer class without an explicit prefix
  silently binds to `YourPlugin.X`, which nobody provides; with a wrapper the calls then return
  `default` forever without any error.
- `SafeWrapper.IPCException` catches only `IpcNotReadyError`. Type/length mismatches and provider
  exceptions still throw. Use `AnyException` for "never throw".
- Wrapped calls to an absent provider do not raise `OnSafeInvocationException`; the event only
  sees exceptions that were actually thrown.
- `wrapper:` on a provider method and on anything `[EzIPCEvent]` has no effect / does not exist.
  Exceptions thrown inside your provider method travel to the caller.
- `Init(object)` skips static members, `Init(Type)` skips instance members. Passing
  `typeof(X)` always selects the `Type` overload.
- Creating two instances of a provider class (or calling `Init` twice) registers the same channels
  again. Dalamud keeps one Func and one Action per channel, so the later registration replaces
  the earlier one, and disposing ANY of the tokens sets the slot to `null` for everybody
  (`UnregisterFunc`/`UnregisterAction` take no delegate). Event subscriptions are a
  `HashSet<Delegate>`: a second instance adds a second subscription (its method also fires), the
  same instance twice does not. Create IPC holder objects once.
- `void` vs value-returning must match across plugins (Action vs Func are separate slots on a
  channel). Keep `actionLastGenericType` at its default unless the other side documents otherwise.
- Provider methods run on whatever thread the caller is on, synchronously. Marshal to the
  framework thread yourself if the method touches game state.
- Only types both plugins share through Dalamud should appear in signatures; see the `dalamud`
  skill, `references/topic-ipc.md`.
- The private arrays `FuncTypes` / `ActionTypes` in `EzIPC.cs` list 10 generic arities but are
  unused; the real limit is 8 parameters.
- `EzSharedData.Get(name, CreationMode.ReadAndKeep)` always throws `ArgumentOutOfRangeException`
  when the name is not cached; use `TryGet` for that mode.
- The upstream `EzIPC.md` examples contain `public class IPCProvider()` followed by an explicit
  constructor, which is not valid C#; take the attribute usage from it, not the class shape.

# Events, throttlers, schedulers, singletons, lazy data, loader

Covers: `EzEventManager`, `Events/ProperOnLogin`, `Throttlers`, `Schedulers`, `Singletons`,
`LazyDataHelpers` and `Loader`, plus what `ECommonsMain.Dispose()` does for each. Verified against
ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are
relative to the skill root.

Stock Dalamud services (`IFramework`, `IClientState`) are covered by the `dalamud` skill. Task
queues with conditions (`TaskManager`) are in the automation topic, not here.

```csharp
using ECommons.EzEventManager;                     // EzFrameworkUpdate, EzLogout, EzTerritoryChanged
using ECommons.Events;                             // ProperOnLogin
using ECommons.Throttlers;                         // EzThrottler, FrameThrottler (+ generic forms)
using ECommons.Schedulers;                         // TickScheduler, ExecuteForScheduler, IScheduler, ThreadPool
using ECommons.Schedulers.Sequential.TaskQueue;    // SequentialTaskQueue (experimental)
using ECommons.Schedulers.SingleThreadRunnerHelper;// SingleThreadRunner, SyncHelper (obsolete)
using ECommons.Singletons;                         // SingletonServiceManager
using ECommons.LazyDataHelpers;                    // LazyData<T>, Purgatory, PurgeableAttribute
using ECommons.Loader;                             // PluginLoader, ILoadable
```

Everything here except `ThreadPool`, `SequentialTaskQueue`, `SingleThreadRunner`, `LazyData<T>` and
`Purgatory` touches `Svc.*`, so `ECommonsMain.Init(pluginInterface, this)` must have run first. No
`Module` flag is needed for anything in this file.

## What ECommonsMain.Dispose() does for this area

Source: `references/ecommons/ECommons/ECommonsMain.cs`. Relevant calls, in execution order (each
wrapped in `GenericHelpers.Safe` unless noted):

| Order | Call | Effect |
|---|---|---|
| 1 | `SingletonServiceManager.DisposeAll` | Disposes and nulls singleton members; runs first, so services can still use throttlers and events in their `Dispose` |
| 2 | `PluginLoader.Dispose` | No-op unless `PluginLoader.Init` was called |
| later | `ProperOnLogin.Dispose` | Unsubscribes the two `ClientState.Login` handlers |
| later | `EzThrottler.Throttler = null; FrameThrottler.Throttler = null;` | Plain assignments, not wrapped |
| later | `EzEvent.DisposeAll` | Disposes every live `EzFrameworkUpdate`, `EzLogout`, `EzTerritoryChanged` |
| near end | `Purgatory.Purge` | Runs registered purge actions, nulls statics of registered types |

Not tracked, so not cleaned up by ECommons: `TickScheduler`, `ExecuteForScheduler`, `ThreadPool`,
`SequentialTaskQueue`, `SingleThreadRunner`. Dispose those yourself if they can outlive the plugin.

## EzEventManager

Source: `references/ecommons/ECommons/EzEventManager/`. Three public wrapper classes, each
`IDisposable`; constructing one subscribes immediately and adds the instance to an internal static
`Registered` list. There is no separate register call and no need to keep the reference.

| Class | Constructor | Wraps |
|---|---|---|
| `EzFrameworkUpdate` | `EzFrameworkUpdate(Action @delegate)` | `Svc.Framework.Update` |
| `EzLogout` | `EzLogout(Action @delegate)` | `Svc.ClientState.Logout` (the `int type, int code` arguments are dropped) |
| `EzTerritoryChanged` | `EzTerritoryChanged(Action<uint> @delegate)` | `Svc.ClientState.TerritoryChanged` (delegate is subscribed directly) |

- All three throw `ArgumentNullException` for a null delegate.
- `Dispose()` unsubscribes and removes the instance from `Registered`. Use it for early removal.
- `EzEvent` is an `internal static` class; `EzEvent.DisposeAll()` is internal and only called by
  `ECommonsMain.Dispose()`. It disposes all instances and then sets the three `Registered` lists
  to `null`, so constructing any of these classes after `ECommonsMain.Dispose()` throws
  `NullReferenceException` (after having already subscribed the handler).
- There is no exception handling around the delegate; an exception propagates into Dalamud's
  event invocation.
- Callbacks run on the framework thread (that is where Dalamud raises these events).
- The XML comment on `EzTerritoryChanged` says "Framework.Update"; that is a copy-paste error in
  the source, it wraps `TerritoryChanged`.

```csharp
// AutoRetainer, Modules/FCPointsUpdater.cs (private ctor, created by SingletonServiceManager)
private FCPointsUpdater()
{
    ProperOnLogin.RegisterInteractable(() => ScheduleUpdateIfNeeded(), true);
    new EzLogout(() => TaskManager.Abort());
    new EzTerritoryChanged((x) => ScheduleUpdateIfNeeded());
}
```

## ProperOnLogin

Source: `references/ecommons/ECommons/Events/ProperOnLogin.cs`. `public static class ProperOnLogin`
in `ECommons.Events`. Fires callbacks once the local player actually exists after login, instead
of at the raw `ClientState.Login` event.

| Member | Notes |
|---|---|
| `static bool PlayerPresent { get; }` | `Svc.Objects.LocalPlayer != null && Svc.PlayerState.ContentId != 0` |
| `static void RegisterAvailable(Action action, bool fireImmediately = false)` | Fires when `PlayerPresent` becomes true after login |
| `static void RegisterInteractable(Action action, bool fireImmediately = false)` | Fires when `Player.Interactable` becomes true after login (`Player.Object?.IsTargetable ?? false`, `ECommons.GameHelpers.Player`) |
| `static void Unregister(Action action)` | Removes from either set; logs a warning if not registered |
| `static void FireArtificially()` | Runs both checks once via `Svc.Framework.RunOnFrameworkThread` |
| `static void Register(Action action, bool fireImmediately = false)` | `[Obsolete("Use either RegisterAvailable or RegisterInteractable")]`; forwards to `RegisterAvailable` |

Behaviour:

- Detection: the first registration of each kind subscribes one master handler to
  `Svc.ClientState.Login`. On login that handler subscribes a `Svc.Framework.Update` poller; the
  poller checks the condition every frame and, once true, unsubscribes itself and invokes every
  registered action. Each action is wrapped in try/catch (`PluginLog.Error` plus `e.Log()`), so one
  failing callback does not stop the others.
- `fireImmediately: true` creates `new TickScheduler(() => { if(<condition>) action(); })`, so the
  action runs on the next framework update, and only if the condition is already true at that
  moment. It is the way to cover "plugin loaded while already logged in". It does not wait for the
  condition to become true.
- An action can be in only one of the two sets. Registering a delegate that is already registered
  (in either set) logs a warning and does nothing, including no immediate fire.
- Registration is keyed on delegate equality (`HashSet<Action>`). `Unregister` needs the same
  delegate instance; a fresh lambda will not match.
- `FireArtificially()` invokes all callbacks of a set whose condition is currently true. It does
  not check that a login happened.
- `ECommonsMain.Dispose()` calls the internal `ProperOnLogin.Dispose()`, which only unsubscribes the
  `Login` handlers. It does not clear the action sets.
- The register methods are not thread-safe (plain `HashSet`); call them during plugin
  construction or on the framework thread. Callbacks run on the framework thread.

```csharp
ProperOnLogin.RegisterAvailable(OnLogin, fireImmediately: true);
// ...
private void OnLogin() { /* Svc.Objects.LocalPlayer is non-null here */ }

// AutoRetainer uses PlayerPresent as a generic "logged in and loaded" guard:
if(!ProperOnLogin.PlayerPresent) return;
```

## Throttlers

Source: `references/ecommons/ECommons/Throttlers/`. Four types in `ECommons.Throttlers`:

| Type | Key | Unit | Clock |
|---|---|---|---|
| `static class EzThrottler` | `string` | milliseconds | `Environment.TickCount64` |
| `class EzThrottler<T>` | `T` | milliseconds | `Environment.TickCount64` |
| `static class FrameThrottler` | `string` | frames | `(long)Svc.PluginInterface.UiBuilder.FrameCount` |
| `class FrameThrottler<T>` | `T` | frames | same |

The static classes forward every call to an internal static instance
(`internal static EzThrottler<string> Throttler = new();`, same for `FrameThrottler`).

### EzThrottler (static) exact members

```csharp
public static IReadOnlyCollection<string> ThrottleNames { get; }
public static bool Throttle(string name, TimeSpan ts, bool reThrottle = false);
public static bool Throttle(string name, int miliseconds = 500, bool rethrottle = false);
public static void Reset(string name);
public static bool Check(string name);
public static long GetRemainingTime(string name, bool allowNegative = false);
public static void ImGuiPrintDebugInfo();
public static void Clear();
```

`EzThrottler<T>` has the same instance members with `T name` instead of `string name`. The
`TimeSpan` overload casts `ts.TotalMilliseconds` to `int`.

Parameter names are spelled exactly as above: `miliseconds` (one "l") and, in the `TimeSpan`
overload only, `reThrottle` with a capital T. There is no parameter called `ms`. Pass these
positionally unless you copy the spelling.

### FrameThrottler (static) exact members

```csharp
public static IReadOnlyCollection<string> ThrottleNames { get; }
public static bool Throttle(string name, int frames = 60, bool rethrottle = false);
public static void Reset(string name);
public static bool Check(string name);
public static long GetRemainingTime(string name, bool allowNegative = false);
public static void ImGuiPrintDebugInfo();
public static void Clear();
```

`FrameThrottler<T>` mirrors it with `T name`. There is no `TimeSpan` overload. `GetRemainingTime`
returns frames, despite the name.

### Semantics (identical for both)

| Call | Behaviour |
|---|---|
| `Throttle(name, n)` | Unknown name, or deadline already passed (`now > deadline`, strict): sets deadline to `now + n` and returns `true`. Otherwise returns `false` and leaves the deadline alone. |
| `Throttle(name, n, rethrottle: true)` | Same, but in the `false` case the deadline is also pushed to `now + n`. Repeated calls keep the throttle closed until calls stop for `n`. |
| `Check(name)` | `true` if the name is unknown or its deadline has passed. Never modifies state. |
| `Reset(name)` | Removes the entry; the next `Throttle`/`Check` returns `true`. |
| `GetRemainingTime(name)` | Remaining ms/frames, clamped to 0. Unknown name returns 0. With `allowNegative: true` returns the raw difference, and for an unknown name returns `-now` (a large negative number). |
| `ThrottleNames` | Live view of the dictionary keys. |
| `ImGuiPrintDebugInfo()` | One `ImGuiEx.Text` line per entry, green if open, red if throttled. Call inside an ImGui draw. |
| `Clear()` | Removes all entries. |

So `true` from `Throttle` means "go ahead, and the throttle has just been armed". The first call
for a new name always returns `true`.

- Storage is a plain `Dictionary<T, long>`: not thread-safe. Use from one thread (normally the
  framework or draw thread) or keep a private `EzThrottler<T>` per thread.
- `FrameThrottler` counts Dalamud UI frames through `UiBuilder.FrameCount`, so it needs
  `Svc.PluginInterface` (after `ECommonsMain.Init`).
- Names in the static throttlers are global to the plugin. Prefix them (`"MyFeature.Click"`), or
  use a private `new EzThrottler<T>()` with an enum or `uint` key.
- After `ECommonsMain.Dispose()` both static `Throttler` fields are `null`; any static call then
  throws `NullReferenceException`. Watch for draw or framework callbacks that are still subscribed
  after dispose. Own `EzThrottler<T>` instances are unaffected.

### Examples from consumers

```csharp
// Rate-limit a per-frame action (AutoRetainer.cs); default is 500 ms
if(EzThrottler.Throttle("CheckHTweaks")) Utils.EnsureEnhancedLoginIsOff();

// Rate-limit a repeated notification, one throttle per message (AutoRetainer FastAddition.cs)
if(EzThrottler.Throttle($"Error_{error}", 1000)) Notify.Error(error);

// Arm with rethrottle, query with Check (AutoRetainer EntrustManager.cs): button is disabled
// for 2 s after a successful import
if(ImGuiEx.IconButton(FontAwesomeIcon.Paste, enabled: EzThrottler.Check("ImportPlan")))
{
    // ...
    EzThrottler.Throttle("ImportPlan", 2000, true);
}

// "Wait until a condition has been stable for N frames" (AutoRetainer RetainerHandlers.cs)
const string thrName = "ClickResultReassign.WaitForButtonEnabled";
if(!addon->ReassignButton->IsEnabled)
{
    FrameThrottler.Throttle(thrName, 5, true);   // keep pushing the deadline while not ready
}
if(FrameThrottler.Check(thrName) && addon->ReassignButton->IsEnabled && Utils.GenericThrottle)
{
    new AddonMaster.RetainerTaskResult(addon).Reassign();
}

// Delay step inside a TaskManager queue (PandorasBox AutoSprint.cs): first task arms and
// returns true, second task blocks the queue until the throttle has expired
TaskManager.Enqueue(() => EzThrottler.Throttle("Sprinting", (int)(Config.ThrottleF * 1000)));
TaskManager.Enqueue(() => EzThrottler.Check("Sprinting"));
```

The last pattern only works if the throttle is open when the first task runs; if it is still
closed from an earlier run, `Throttle` returns `false` and that task is retried until it opens.

## Schedulers

Source: `references/ecommons/ECommons/Schedulers/`.

### IScheduler

`public interface IScheduler : IDisposable` with no members of its own. Implemented by
`TickScheduler` and `ExecuteForScheduler`.

### TickScheduler

```csharp
public class TickScheduler : IScheduler
{
    public bool Disposed { get; private set; }
    public TickScheduler(Action function, long delayMS = 0);
    public void Dispose();
}
```

- The constructor computes `Environment.TickCount64 + delayMS` and subscribes to
  `Svc.Framework.Update`. The action runs once, on the framework thread, on the first update at
  or after that time, then the scheduler disposes itself. With `delayMS = 0` that is the next
  framework update, never synchronously inside the constructor.
- Exceptions from the action are caught and logged with `PluginLog.Error` (message and stack
  trace); the scheduler still disposes.
- `Dispose()` before it fires cancels it. It is idempotent.
- Can be constructed from any thread; this is the usual ECommons way to hop onto the framework
  thread or to defer a collection mutation out of an ImGui loop.
- Instances are not tracked. `ECommonsMain.Dispose()` does not cancel pending ones; keep the
  reference and dispose it if a long delay could outlive the plugin.

```csharp
// Defer removal until after the current draw loop (AutoRetainer UI code)
new TickScheduler(() => C.WhitelistedAccounts.Remove(x));

// Delayed one-shot
var s = new TickScheduler(() => DoThing(), 1000);
s.Dispose(); // cancel
```

### ExecuteForScheduler

```csharp
public class ExecuteForScheduler : IScheduler
{
    public ExecuteForScheduler(Action function, long executeForMS);
    public void Dispose();
}
```

Runs `function` on every framework update, starting with the next one, until
`Environment.TickCount64 > start + executeForMS`; the check happens after the call, so the
function runs at least once. Exceptions are caught and logged per call. `Dispose()` stops it
early. No `Disposed` property. Not tracked by `ECommonsMain.Dispose()`.

### ThreadPool

`ECommons.Schedulers.ThreadPool`, an own small pool, unrelated to `System.Threading.ThreadPool`
(the name clashes if both namespaces are imported; qualify it).

```csharp
public class ThreadPool : IDisposable
{
    public ThreadPool();                 // MaxThreads = 8
    public ThreadPool(int maxThreads);
    public bool IsWorking { get; }       // BusyThreads > 0
    public (uint RunningThreads, uint BusyThreads, int TasksQueued) State { get; }
    public void Run(Action task, Action<Exception?>? onCompletion = null);
    public void Dispose();
}
```

- `Run` enqueues and starts a new `Thread` when fewer threads are running than
  `min(MaxThreads, queued + busy)`. Threads are created with `new Thread(...)` without
  `IsBackground`, so they are foreground threads.
- A worker exits after about 10 s idle (100 polls of 100 ms) or once `Dispose()` has been called.
  `Dispose()` only sets a flag: queued tasks that were not started are dropped silently, a running
  task is not interrupted.
- Without `onCompletion`, an exception in the task is logged. With `onCompletion`, nothing is
  logged; the callback receives the exception, or `null` on success, on the worker thread.
  Exceptions thrown by `onCompletion` itself are logged.
- The counters are `volatile uint` changed with `++`/`--` from several threads, so `State` and
  `IsWorking` are approximate.
- Tasks run off the framework thread: do not touch game memory or most Dalamud services there;
  hop back with `new TickScheduler(...)` or `Svc.Framework.RunOnFrameworkThread`.

### SequentialTaskQueue (experimental)

Source: `references/ecommons/ECommons/Schedulers/Sequential/TaskQueue/`. The class carries
`[Experimental("E1")]` (declared on the partial in `QueuedTask.cs`), so any use produces compiler
diagnostic `E1` as an error until suppressed (`#pragma warning disable E1` or `<NoWarn>`).

```csharp
public partial class SequentialTaskQueue : IDisposable
{
    public SequentialTaskQueue(Action<Exception> exceptionHandler, int idleTimeoutSeconds = 60);
    public TaskCompletionToken<T> Enqueue<T>(Func<T> task, bool clearQueue = false);
    public TaskCompletionToken Enqueue(Action task, Action<Exception> callback = null, bool clearQueue = false);
    public void Dispose();

    public class TaskCompletionToken
    {
        public bool IsCompleted { get; }
        public void Wait();
        public bool Wait(int millisecondsTimeout);
    }
    public class TaskCompletionToken<T> : TaskCompletionToken
    {
        public T Result { get; internal set; }
        public Exception Exception { get; internal set; }
    }
}
```

- One background worker thread (`"SequentialTaskQueue Worker"`) executes tasks strictly in order.
  It is created on demand and exits after `idleTimeoutSeconds` without work; the next `Enqueue`
  starts a new one.
- `exceptionHandler` is required (`ArgumentNullException` on null) and receives every task
  exception. For `Enqueue<T>` the exception is also stored in `token.Exception`. For the `Action`
  overload, `callback` runs after the task with the exception or `null`.
- `clearQueue: true` drops all not-yet-started tasks before adding the new one.
- Dropped tasks, tasks pending at `Dispose()`, and tasks enqueued after `Dispose()` never
  complete their token: `IsCompleted` stays `false` and `Wait()` without timeout blocks forever.
  Use `Wait(int)`.
- Never call `Wait()` on the framework thread for a task that itself needs the framework thread.
- The internal static `Queues` list is never read or written elsewhere in the library, so
  instances are not registered anywhere and `ECommonsMain.Dispose()` does not dispose them.
  Dispose your own queue.

### SingleThreadRunner

Source: `references/ecommons/ECommons/Schedulers/SingleThreadRunnerHelper/`.

```csharp
public sealed class SingleThreadRunner
{
    public SingleThreadRunner(Action action, string name = "SingleThreadRunner");
    public bool IsRunning { get; }
    public void Start();
    public void Join();
}

[Obsolete("Not used, rework pending")]
public static class SyncHelper
{
    public static void Sync(this Task task);
    public static T Sync<T>(this Task<T> task);
}
```

Runs `action` on a dedicated background thread that has a custom `SynchronizationContext`
(pumped by a second thread). Exceptions are logged. `SyncHelper.Sync` throws
`InvalidOperationException("No SingleThreadContext is active.")` outside such a thread.
`SingleThreadContext` and `ThreadGuard` are internal. `SyncHelper` is marked obsolete with
"rework pending" and nothing else in the library uses these types; avoid them in new code.

## SingletonServiceManager

Source: `references/ecommons/ECommons/Singletons/SingletonServiceManager.cs`.

```csharp
public static class SingletonServiceManager
{
    public static void Initialize(Type staticTypeMaster);

    [AttributeUsage(AttributeTargets.Field | AttributeTargets.Property | AttributeTargets.Class, AllowMultiple = false)]
    public sealed class PriorityAttribute : Attribute   // "Lower number = loaded earlier"
    {
        public int Priority;
        public PriorityAttribute(int priority);
    }
}
```

The attribute is nested: write `[SingletonServiceManager.Priority(n)]`. `DisposeAll()` is
internal.

`Initialize(typeof(Container))`:

1. Takes the container type plus its public nested types (`GetNestedTypes()`), and records them
   for later disposal.
2. Enumerates all static fields, then all static properties, public and non-public, of each type
   (`GetFieldPropertyUnions`). Compiler-generated backing fields of auto-properties are included
   in the field pass, so an auto-property is normally filled through its backing field; a
   `private set` or get-only auto-property works.
3. For every member whose current value is `null`, takes the member type's first constructor
   returned by reflection (public or non-public, so a private constructor is fine), passes
   `null` for each reference-type parameter and a default value for each value-type parameter,
   and assigns the new instance. Members that already hold a value are skipped.
4. Construction errors are caught and logged per member (`TargetInvocationException` logs the
   inner exception too); remaining members are still processed and the failed member stays
   `null`. A type without any instance constructor logs a warning.

Consequences:

- Services should have a parameterless constructor. They reach other services through the
  static container, so declaration order is construction order: a service may use members
  declared above it.
- Any `null` static member is instantiated, not only "services". A nullable `string` field would
  be attempted too (its first constructor with default arguments). Initialise or avoid unrelated
  nullable statics in the container.
- Interface or abstract member types cannot be constructed; the error is logged.
- Call it on the thread you want the constructors to run on; it is synchronous.

Ordering and `[Priority]`: the sort is
`OrderBy(class priority of DeclaringType).ThenBy(t => fieldPriorities.SafeSelect(t.Value).Key)`.
The first key is the numeric priority of the declaring class, so `[Priority]` on nested container
classes orders those classes. The second key is not the member's numeric priority: it uses the
priority value as an index into the member list and compares the `IFieldPropertyUnion` found
there. With all member priorities left at 0 this is a stable sort that preserves declaration
order. With differing member priorities inside one class the result is not numeric order and may
throw, since `IFieldPropertyUnion` does not implement `IComparable` *(unverified: derived from
reading the source, not executed)*. Neither AutoRetainer nor YesAlready nor PandorasBox uses the
attribute. Rely on declaration order, or on class-level priority with nested classes.

Disposal (`DisposeAll`, called first in `ECommonsMain.Dispose()`):

- Types in registration order; within each type, members in reverse enumeration order
  (properties last-to-first, then fields last-to-first). For field-only containers that is exact
  reverse declaration order. Priorities are not considered.
- Every value implementing `IDisposable` is disposed (exceptions caught and logged), then every
  member, disposable or not, is set to `null`. That includes non-service statics in the container
  (YesAlready's `Service.Quests` dictionary is nulled this way).
- The `SetValue(null, null)` call is outside the try block. A member that cannot be assigned
  `null` (a static value-type field, a `const`, a property without setter that is not an
  auto-property) throws and aborts the rest of `DisposeAll`; the exception is caught by
  `GenericHelpers.Safe`. Keep containers to reference-type fields and auto-properties
  *(unverified: derived from reading the source, not executed)*.
- Afterwards the internal type list is `null`; `Initialize` cannot be called again.

```csharp
// AutoRetainer: Services/!AutoRetainerServiceManager.cs
public static class AutoRetainerServiceManager
{
    public static NeoWindow NeoWindow;
    public static EzIPCManager EzIPCManager;
    public static FCPointsUpdater FCPointsUpdater;   // private parameterless ctor
    // ...
}
// in plugin load, after ECommonsMain.Init and config load:
SingletonServiceManager.Initialize(typeof(AutoRetainerServiceManager));

// YesAlready: properties with private setters work as well
public static class Service
{
    public static TaskManager TaskManager { get; private set; } = null!;
    public static Watcher Watcher { get; private set; } = null!;
}
SingletonServiceManager.Initialize(typeof(Service));
```

## LazyDataHelpers

Source: `references/ecommons/ECommons/LazyDataHelpers/`.

### Purgatory

```csharp
public static class Purgatory
{
    public static void Add(Action action);
    public static void Add(Type type);
}
```

Registry of clean-up work for plugin unload. `Purge()` is internal and called by
`ECommonsMain.Dispose()` near the end. It:

1. invokes every registered action in registration order, each in try/catch with `e.Log()`;
2. for every registered type, sets each static field and property whose type is a class to
   `null` (public and non-public; value types and interfaces are skipped since `Type.IsClass` is
   false for them), errors logged per member;
3. sets the action list to `null`. `Purgatory.Add(Action)` after that throws
   `NullReferenceException`.

Library users of it: `LazyData<T>`, `Ref<T>` (`Add(typeof(Ref<T>))`), `EzPatch`,
`WindowFunctions`, legacy `GameNetwork`. Plugin code can use it for the same purpose: register
a static cache reset without writing dispose plumbing.

```csharp
Purgatory.Add(() => MyStaticCache = null);
Purgatory.Add(typeof(MyStaticHolder));   // nulls all class-typed statics of that type
```

### PurgeableAttribute

`[AttributeUsage(AttributeTargets.Class)] public class PurgeableAttribute : Attribute` with no
members. Nothing in the library reads it at this commit: putting `[Purgeable]` on a class has no
effect. Use `Purgatory.Add(typeof(X))`.

### LazyData<T>

```csharp
public static class LazyData<T>
{
    public static T Get(string tag, Func<T> generator);
    public static T Get(Func<T> generator);
}
```

Per-`T` static cache keyed by `tag`: returns the cached value, or calls `generator`, stores and
returns the result. The cache is cleared at dispose through a `Purgatory` action registered in the
static constructor.

- `Get(Func<T>)` uses one fixed key per closed type `T` (a GUID generated once). All tagless
  calls for the same `T` share a single slot: the first generator wins and later calls with a
  different generator return the first value. Use the tagged overload when a type is cached in
  more than one place.
- A `null` result is cached like any other value.
- Plain `Dictionary`, not thread-safe.
- After `ECommonsMain.Dispose()` the cache is `null` and `Get` throws. First use of a new
  `LazyData<T>` after dispose fails in the static constructor.

```csharp
var names = LazyData<Dictionary<uint, string>>.Get("MyPlugin.ItemNames", BuildItemNames);
```

## Loader

Source: `references/ecommons/ECommons/Loader/`. Remote version-blacklist gate for plugin start.

```csharp
public interface ILoadable { bool IsDisposed { get; } }

public static class PluginLoader
{
    public static bool IsLoaded { get; internal set; }
    public static void Init(ILoadable plugin, IDalamudPluginInterface pluginInterface, string blacklistURL, Action success);
    public static void Dispose();
}
```

`MicroServices` is an internal class holding `[PluginService]` `IDalamudPluginInterface`,
`ICommandManager` and `IFramework`; `Init` fills it with `pluginInterface.Create<MicroServices>()`,
so `PluginLoader.Init` does not depend on `ECommonsMain.Init` for its services (its logging calls
go through `PluginLog`).

`Init`:

- Throws `Exception("Loader has been already called before")` on a second call.
- Registers the command `/load_<InternalName>` ("Load plugin skipping version check").
- On a `Task.Run` thread downloads `blacklistURL` (5 s timeout) and parses it line by line:
  `1.2.3.4` blocks exactly that version, `1.0.0.0, 1.2.0.0` blocks the inclusive range. Invalid
  lines are logged at debug level and ignored. The current version is read from the plugin
  assembly file.
- Download failure is logged and treated as "not blacklisted".
- Not blacklisted: `success` is invoked on the next framework update, then the loader disposes
  itself (removing the command).
- Blacklisted: draws a "version revoked" ImGui window with buttons to open the plugin installer,
  close the window, or load anyway.
- `success` is skipped when `plugin.IsDisposed` is true; exceptions from it are logged.

Notes:

- `IsLoaded` is never set to `true` anywhere in the library (internal setter, no assignment), so
  it always reads `false` and the "already loaded" guard never triggers.
- `PluginLoader.Dispose()` is public and is also called by `ECommonsMain.Dispose()`; it is a no-op
  when `Init` was never called.
- None of the three consumer plugins uses it at the pinned commits. AutoRetainer has only
  commented-out calls to a `PluginLoader.CheckAndLoad` method that does not exist in this
  checkout. Treat the loader as legacy.

## Gotchas

- `EzThrottler.Throttle` returning `true` means "allowed and re-armed", not "currently throttled".
  `Check` is the read-only query and has the same polarity (`true` = open).
- `rethrottle` only matters when the call returns `false`; it extends the deadline from now.
- Named arguments: `miliseconds`, `rethrottle` (int overload), `reThrottle` (`TimeSpan` overload),
  `frames`, `allowNegative`, `delayMS`, `executeForMS`, `fireImmediately`.
- `GetRemainingTime(name, allowNegative: true)` for an unknown name returns `-TickCount64` (or
  minus the frame count), not 0.
- Static throttlers, `Ez*` event wrappers, `LazyData<T>` and `Purgatory.Add(Action)` all throw
  `NullReferenceException` when used after `ECommonsMain.Dispose()`. Unsubscribe your own draw
  and update handlers before calling it, or make them return early on `ECommonsMain.Disposed`.
- `new TickScheduler(action)` never runs synchronously, even when already on the framework
  thread. Code after the constructor runs first.
- `fireImmediately` on `ProperOnLogin` is a one-shot check on the next frame. If the player is
  mid-zone at that moment, the callback is not retried until the next login.
- `SingletonServiceManager.DisposeAll` nulls every static in the container, disposable or not.
  Code that runs after `ECommonsMain.Dispose()` must not read container members.
- `SingletonServiceManager` constructors get `null`/default for every parameter; do not rely on
  constructor injection.
- `ECommons.Schedulers.ThreadPool` shadows `System.Threading.ThreadPool` when both namespaces are
  imported.
- `SequentialTaskQueue` needs the `E1` experimental diagnostic suppressed, and its tokens never
  complete for dropped tasks.
- `[Purgeable]` and `PluginLoader.IsLoaded` are inert at this commit.

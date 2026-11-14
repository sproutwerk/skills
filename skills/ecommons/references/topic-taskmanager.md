# TaskManager (NeoTaskManager and LegacyTaskManager)

Covers: the frame-driven task queue `ECommons.Automation.NeoTaskManager.TaskManager`, its
configuration, ready-made tasks, and the older `ECommons.Automation.LegacyTaskManager.TaskManager`
with a migration mapping. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Source: `references/ecommons/ECommons/Automation/NeoTaskManager/*.cs`,
`references/ecommons/ECommons/Automation/NeoTaskManager/Tasks/*.cs`,
`references/ecommons/ECommons/Automation/LegacyTaskManager/*.cs`.
Related topics: `references/topic-addonmaster.md` (AddonMaster), `references/topic-automation.md` (`Chat`, `Callback`).

## Namespaces

```csharp
using ECommons.Automation.NeoTaskManager;        // TaskManager, TaskManagerTask, TaskManagerConfiguration
using ECommons.Automation.NeoTaskManager.Tasks;  // NeoTasks, DelayTask, FrameDelayTask
using ECommons.Automation.LegacyTaskManager;     // old TaskManager (same type name: never import both)
```

Both namespaces declare a class named `TaskManager`. Import exactly one of them per file, or alias.

## Model

A `TaskManager` owns a `List<TaskManagerTask>`. Its constructor subscribes a private `Tick` to
`Svc.Framework.Update`, so `ECommonsMain.Init` must have run first and every task delegate executes
on the framework thread. Each tick runs at most one call of the current task's function:

| Function returns | Meaning |
|---|---|
| `true` | Task complete. `CurrentTask` is cleared; the next task starts on the next tick. |
| `false` | Not done. The same function is called again next tick, until the time limit expires. |
| `null` | Abort request. See "Returning null" below. |
| throws | Handled per `AbortOnError` and `OnTaskException`. |

A task is dequeued and executed in the same tick (the legacy manager spends one tick dequeuing).
Nothing in the class is synchronised; enqueue and abort from the framework thread. The source
comment on `Tasks` says to modify the list only from the `Framework.Update` event.

## TaskManager members

```csharp
public partial class TaskManager : IDisposable
public TaskManager(TaskManagerConfiguration? defaultConfiguration = null)
```

| Member | Notes |
|---|---|
| `public readonly TaskManagerConfiguration DefaultConfiguration` | Built-in defaults overlaid with the constructor argument (`.With(defaultConfiguration)`). Must never contain null values; checked with `AssertNotNull()` in the constructor and at each busy tick (throws `NullReferenceException`). Its properties are settable at runtime. |
| `public int MaxTasks { get; private set; }` | Highest `NumQueuedTasks` seen during the current run. Updated only inside the tick; reset to 0 on the first idle tick. |
| `public int NumQueuedTasks` | `Tasks.Count + (CurrentTask == null ? 0 : 1)`. |
| `public float Progress` | `MaxTasks == 0 ? 0 : (MaxTasks - NumQueuedTasks) / (float)MaxTasks`. |
| `public bool IsBusy` | `Tasks.Count != 0 \|\| CurrentTask != null`. |
| `public List<TaskManagerTask> Tasks { get; init; }` | Pending queue, index 0 runs next. |
| `public TaskManagerTask? CurrentTask { get; private set; }` | Task being executed (already removed from `Tasks`). |
| `public long RemainingTimeMS { get; set; }` | Milliseconds until the current task times out. Setting it moves the deadline; a task may extend its own deadline this way. It is overwritten with the task's `TimeLimitMS` whenever a new task starts. |
| `public bool StepMode { get; set; }` | `true` unsubscribes `Tick` from `Framework.Update`, `false` subscribes it again. Setting `false` while already `false` subscribes a second time. The doc comment says time limits are ignored in step mode; the tick code has no such exemption. |
| `public void Step()` | Runs one tick manually. Throws `InvalidOperationException` outside step mode. |
| `public void Abort()` | Clears `Tasks`, clears `CurrentTask`, resets the deadline, clears every open stack. |
| `public void AbortCurrent()` | Clears only `CurrentTask`; the queue continues with the next task. |
| `public void Dispose()` | Unsubscribes `Tick` and removes the instance from the static instance list. Does not clear the queue. |
| `internal static void DisposeAll()` | Called by `ECommonsMain.Dispose()`; disposes every live instance. Manual `Dispose()` is only needed to stop a manager earlier. |
| `public List<StackDetails> Stack { get; init; }`, `public bool IsStackActive => Stack.Count > 0;` | See "Stack mode". |

### Enqueue (append to the end)

```csharp
public void Enqueue(Func<bool?> func, string taskName, TaskManagerConfiguration? configuration = null)
public void Enqueue(Func<bool?> func, TaskManagerConfiguration? configuration = null)
public void Enqueue(Func<bool> func, string taskName, TaskManagerConfiguration? configuration = null)
public void Enqueue(Func<bool> func, TaskManagerConfiguration? configuration = null)
public void Enqueue(Action action, string taskName, TaskManagerConfiguration? configuration = null)
public void Enqueue(Action action, TaskManagerConfiguration? configuration = null)
public void EnqueueTask(TaskManagerTask task)
public void EnqueueMulti(params TaskManagerTask?[] tasks)
```

### Insert (prepend to the front)

```csharp
public void Insert(Func<bool?> func, string taskName, TaskManagerConfiguration? configuration = null)
public void Insert(Func<bool?> func, TaskManagerConfiguration? configuration = null)
public void Insert(Func<bool> func, string taskName, TaskManagerConfiguration? configuration = null)
public void Insert(Func<bool> func, TaskManagerConfiguration? configuration = null)
public void Insert(Action action, string taskName, TaskManagerConfiguration? configuration = null)
public void Insert(Action action, TaskManagerConfiguration? configuration = null)
public void InsertTask(TaskManagerTask task)
public void InsertMulti(params TaskManagerTask?[] tasks)
```

`InsertMulti` iterates the arguments in reverse and inserts each at index 0, so the given order is
preserved at the front of the queue. Consecutive single `Insert` calls end up in reverse call order;
use `InsertMulti` or `InsertStack` for a sequence. Null entries in `EnqueueMulti`/`InsertMulti` are
skipped. Every call logs `Enqueued task {Name}@{Location}` / `Inserted task ...`.

Delegate forms: `Func<bool?>` is stored as is; `Func<bool>` is wrapped (`() => function()`), so it
can only signal done/retry; `Action` is wrapped to run once and return `true`. A lambda whose
branches return `true`, `false` and `null` binds to `Func<bool?>`.

### Delays

```csharp
public void EnqueueDelay(int ms, bool isFrame = false, TaskManagerConfiguration configuration = null)
public void InsertDelay(int ms, bool isFrame = false, TaskManagerConfiguration configuration = null)
```

`isFrame == false` creates a `DelayTask` (milliseconds); `isFrame == true` creates a
`FrameDelayTask` and `ms` is a frame count. The XML comment "measured in seconds" is wrong.

### Stack mode

Source: `references/ecommons/ECommons/Automation/NeoTaskManager/TaskManager.Stack.cs`.

```csharp
public record struct StackDetails()
{
    public string Name { get; set; }
    public List<TaskManagerTask> Stack { get; set; } = [];
}

public void EnqueueStack(Action executeAction)                 // name defaults to "Unnamed stack"
public void EnqueueStack(string name, Action executeAction)
public void InsertStack(Action executeAction)
public void InsertStack(string name, Action executeAction)

// all [Obsolete("Use EnqueueStack with action")]
public void BeginStack()
public void BeginStack(string? name)
public void EnqueueStack()
public void InsertStack()
public void DiscardStack()
```

`Stack` is a list of open stacks, innermost last, and `IsStackActive` is true while it is not
empty. `BeginStack` pushes a new `StackDetails`. While a stack is open, every
`Enqueue*`/`Insert*`/delay call goes into the innermost stack instead of `Tasks`. `EnqueueStack()`
pops the innermost stack and appends its tasks to the queue; `InsertStack()` pops it and puts the
tasks at the front in order; `DiscardStack()` pops it and drops the tasks. An outer stack stays
open, so when stacks nest, the "queue" an inner stack is flushed into is the enclosing stack.
Calling the parameterless `EnqueueStack()` or `InsertStack()` with no open stack logs an error and
does nothing. `Abort()` clears every open stack.

The `Action` overloads wrap `BeginStack(name)` ... `EnqueueStack()`/`InsertStack()` around
`executeAction`; if it throws, the exception is logged and that stack is discarded. Use them
instead of the manual calls. This is the replacement for the legacy immediate queue: a running
task can build a sub-sequence with ordinary `Enqueue` calls and insert it ahead of the remaining
queue, to any nesting depth.

Flushing a stack rewrites each task's `Name` to
`[Stack (enqueue, <count>): <stack name>] <task name>` (or `insert`), so do not match queued tasks
by their original name after a stack flush.

Up to package 3.2.1.22 there was one stack: `Stack` was `List<TaskManagerTask>`, `IsStackActive`
had a private setter, `BeginStack()` while a stack was active logged a warning and cleared it, and
the named overloads did not exist. Nested stacks are in 3.2.1.23 and later. Code that reads
`TM.Stack` directly does not compile across that boundary; the `Action` overloads work on both
sides.

```csharp
TM.Enqueue(() => TM.InsertStack(() =>
{
    foreach(var slot in slots) TM.Enqueue(() => Sell(slot), $"Sell({slot})");
    TM.EnqueueDelay(333);
    TM.Enqueue(CloseInventory);
}));
```

## TaskManagerConfiguration

```csharp
[OverloadResolutionPriority(1)]
public TaskManagerConfiguration(int? timeLimitMS = null, bool? abortOnTimeout = null,
    bool? abortOnError = null, bool? timeoutSilently = null, bool? showDebug = null,
    bool? showError = null, bool? executeDefaultConfigurationEvents = null, Guid? discardGuid = null)

public TaskManagerConfiguration(TimeSpan? timeLimitMS = null, bool? abortOnTimeout = null,
    bool? abortOnError = null, bool? timeoutSilently = null, bool? showDebug = null,
    bool? showError = null, bool? executeDefaultConfigurationEvents = null)   // no discardGuid
```

All properties are nullable with `get; set;`. `null` means "inherit from the manager's
`DefaultConfiguration`"; resolution is per property at tick time
(`CurrentTask.Configuration?.X ?? DefaultConfiguration.X!.Value`).

| Property | Type | Built-in default | Meaning |
|---|---|---|---|
| `TimeLimitMS` | `int?` | `30000` | Time budget of one task, counted from its first tick. |
| `AbortOnTimeout` | `bool?` | `true` | On timeout clear the whole queue (`Abort()`); `false` drops only the timed-out task. |
| `AbortOnError` | `bool?` | `true` | Same choice for an exception thrown by the task. |
| `TimeoutSilently` | `bool?` | `false` | Suppresses the timeout warning log. |
| `ShowDebug` | `bool?` | `false` | `true` sends progress messages to `PluginLog.Debug`; `false` sends them to `InternalLog.Debug`. Prefix `[NeoTaskManager] `. |
| `ShowError` | `bool?` | `true` | Log task exceptions (`e.Log()`). |
| `ExecuteDefaultConfigurationEvents` | `bool?` | `true` | When a task has its own configuration, also fire the events of `DefaultConfiguration`. Tasks without a configuration always fire the default events. |
| `DiscardGuid` | `Guid?` | `null` | If set, a timeout-abort or a `null` result removes only queued tasks carrying the same GUID instead of the whole queue. Nothing in the library assigns it; set it yourself on a group of tasks. |

Event fields (public delegate fields, not C# events; exceptions thrown by handlers are caught and
logged):

```csharp
public delegate void OnTaskExceptionDelegate(TaskManagerTask task, Exception exception, ref bool @continue, ref bool? abort);
public OnTaskExceptionDelegate? OnTaskException;
public delegate void OnTaskTimeoutDelegate(TaskManagerTask task, ref long remainingTimeMS);
public OnTaskTimeoutDelegate? OnTaskTimeout;
public delegate void OnTaskCompletionDelegate(TaskManagerTask task, ref bool? isCompleted);
public OnTaskCompletionDelegate? OnTaskCompletion;
public delegate void CompanionActionDelegate(TaskManagerTask task);
public CompanionActionDelegate? CompanionAction;
```

- `OnTaskException`: `@continue = true` keeps the task alive (it is called again next tick);
  otherwise a non-null `abort` overrides `AbortOnError`.
- `OnTaskTimeout`: fired when the deadline has passed; writing a positive value to
  `remainingTimeMS` grants more time and prevents the timeout.
- `OnTaskCompletion`: fired when the function returned `true` or `null`; the handler may rewrite
  the result (`false` re-runs the task, `null` aborts).
- `CompanionAction`: fired at the end of every tick in which a task was processed, including the
  tick where it completed, timed out or threw. Not fired when the manager was aborted from inside
  the task.

Order for each event: default configuration first (if applicable), then the task's configuration.

### With

```csharp
public static class TaskManagerConfigurationExtensions   // C# 14 extension block
public TaskManagerConfiguration With(TaskManagerConfiguration? dominantConfiguration, bool copyEvents = true)
```

`a.With(b)` returns a new configuration where each property is `b`'s value if non-null, else
`a`'s. With `copyEvents: true` (default) the four event fields are taken from `b` only, even when
`b` or its fields are null, so `a`'s handlers are lost; with `false` they are taken from `a` only.
Events are never merged.

Construction styles seen in consumers, all valid against the signatures above:

```csharp
TaskManager TM = new(new(timeLimitMS: 20000, abortOnTimeout: true, showDebug: true));
TaskManager TM2 = new(new() { TimeoutSilently = true, ShowDebug = false });
TM.Enqueue(Func, "name", new(timeLimitMS: 10 * 60 * 1000));
```

## Timeout and error handling

Per tick, in order:

1. First tick of a task: `RemainingTimeMS = TimeLimitMS`, log `→Starting to execute task [...]`.
2. If `RemainingTimeMS < 0`: fire `OnTaskTimeout`. If still negative, log and throw
   `TaskTimeoutException` (a plain `Exception` subclass with no members) internally. The function
   is not called on that tick.
3. Catch of `TaskTimeoutException`: unless `TimeoutSilently`, `e.LogWarning()`. Then if
   `AbortOnTimeout`: tasks with the same `DiscardGuid` are removed when the task has one, otherwise
   `Abort()`. If not `AbortOnTimeout`: only the current task is dropped. The exception never
   reaches plugin code and does not go through `OnTaskException`.
4. Otherwise the function runs. If `CurrentTask` is null afterwards (the task called `Abort()` or
   `AbortCurrent()`), `PluginLog.Warning("[NeoTaskManager] Task manager was aborted from inside the
   task.")` is written and the tick ends. Tasks the function enqueued after aborting stay queued,
   so "abort, then enqueue a recovery sequence" works.
5. Any other exception: logged if `ShowError`; `OnTaskException` handlers run; then `Abort()` or
   drop the current task according to `AbortOnError` / the handler's `abort`.

### Returning null

`null` is documented as "cancel the current queue". The implementation reads
`CurrentTask.Configuration.DiscardGuid` without a null check in this branch:

- Task has a per-task configuration with `DiscardGuid == null`: `Abort()`.
- Task has a configuration with a `DiscardGuid`: only tasks with that GUID are removed; a warning
  `Removed {n} tasks from stacks with matching GUID ...` is logged unless `TimeoutSilently`.
- Task has no per-task configuration: a `NullReferenceException` is raised inside the tick and
  handled as a task error. With the default `AbortOnError = true` the queue is still aborted, but
  an error is logged (if `ShowError`) and `OnTaskException` fires. With `AbortOnError = false` only
  the current task is dropped and the rest of the queue keeps running.

To get a clean abort from a task that may return `null`, pass a configuration
(`new TaskManagerConfiguration()` is enough) or call `Abort()` on the manager and return `false`.

## TaskManagerTask

```csharp
public class TaskManagerTask
{
    public string Name { get; set; }
    public string Location { get; init; }
    public Func<bool?> Function { get; init; }
    public TaskManagerConfiguration? Configuration { get; init; }

    public TaskManagerTask(Func<bool?> function, TaskManagerConfiguration? configuration = null)
    public TaskManagerTask(Func<bool> function, TaskManagerConfiguration? configuration = null)
    public TaskManagerTask(Action action, TaskManagerConfiguration? configuration = null)
    public TaskManagerTask(Func<bool?> function, string taskName, TaskManagerConfiguration? configuration = null)
    public TaskManagerTask(Func<bool> function, string taskName, TaskManagerConfiguration? configuration = null)
    public TaskManagerTask(Action action, string taskName, TaskManagerConfiguration? configuration = null)
}
```

`Name` defaults to the delegate's method name and `Location` to its declaring type name (via
`GetMethodInfo()`); for lambdas these are compiler-generated names, so pass `taskName` for
readable logs. Target-typed `new(...)` works where a `TaskManagerTask` is expected:
`TM.InsertMulti(new(() => DoA()), new(() => IsDone(), new(timeLimitMS: 300000)));`.

## Ready-made tasks (`NeoTaskManager.Tasks`)

```csharp
public class DelayTask      { public DelayTask(int ms, TaskManagerConfiguration? configuration = null) }
public unsafe class FrameDelayTask { public FrameDelayTask(int ms, TaskManagerConfiguration? configuration = null) }
// both: public static implicit operator TaskManagerTask(...)
```

- `DelayTask`: waits `ms` milliseconds measured from its first tick (`Environment.TickCount64`).
  Own configuration `timeLimitMS: ms * 2 + 5000, abortOnTimeout: false`, overlaid with the
  argument. Name `Delay ({ms}ms)`.
- `FrameDelayTask`: waits `ms` frames using `Framework.Instance()->FrameCounter`. Own configuration
  `abortOnTimeout: false` only, so the manager's `TimeLimitMS` still applies; a frame delay that
  lasts longer times out (and the queue continues). Name `Delay ({ms} frames)`.

```csharp
public static unsafe class NeoTasks
public static TaskManagerTask ApproachObjectViaAutomove(Func<IGameObject> getObjectFunc, float distance = 4f, TaskManagerConfiguration? configuration = null)
public static TaskManagerTask InteractWithObject(Func<IGameObject> getObjectFunc, bool checkLos = false, TaskManagerConfiguration? configuration = null)
public static TaskManagerTask WaitForScreenAndPlayer(TaskManagerConfiguration? configuration = null)
public static TaskManagerTask WaitForNotOccupied(TaskManagerConfiguration? configuration = null)
```

| Task | Behaviour |
|---|---|
| `ApproachObjectViaAutomove` | Targets the object (`Svc.Targets.Target`), then `/lockon on` + `/automove on` through `Chat.ExecuteCommand`; when moving and closer than `distance` sends `/automove off`. Completes when not moving and within `distance`. Waits while `getObjectFunc()` returns null or `Player.Interactable` is false. |
| `InteractWithObject` | Targets the object, then `TargetSystem.Instance()->InteractWithObject(obj.Struct(), checkLos)` and completes. Requires targetable object, no animation lock, `Player.Interactable`. |
| `WaitForScreenAndPlayer` | `Player.Interactable && !Player.IsAnimationLocked && GenericHelpers.IsScreenReady()`. Time limit 10 minutes unless overridden. |
| `WaitForNotOccupied` | Same plus `!GenericHelpers.IsOccupied()`. Time limit 10 minutes. Its task name is also "Wait for screen fadeout complete". |

They use the global `EzThrottler` keys `TaskFunction.AutomoveOff`, `TaskFunction.Lockon`,
`TaskFunction.SetTarget`, `TaskFunction.Interact`. Enqueue with `EnqueueTask`/`InsertTask`/`*Multi`:

```csharp
TM.EnqueueTask(NeoTasks.ApproachObjectViaAutomove(() => FindNpc(), 6.3f));
TM.EnqueueTask(NeoTasks.InteractWithObject(() => FindNpc()));
```

## Worked example: wait for addons and drive them through AddonMaster

Names verified in `references/ecommons/ECommons/GenericHelpers/AddonHelpers.cs`
(`TryGetAddonMaster<T>(out T)`, `TryGetAddonByName<T>`, `IsAddonReady`),
`references/ecommons/ECommons/UIHelpers/AddonMasterImplementations/SelectString.cs`,
`SelectYesno.cs`, `!AddonMasterBase.cs` (same folder), and
`references/ecommons/ECommons/Throttlers/EzThrottler.cs`
(`public static bool Throttle(string name, int miliseconds = 500, bool rethrottle = false)`:
returns `true` and arms the throttle if it is not armed or has expired, else `false`).

```csharp
using ECommons.Automation.NeoTaskManager;
using ECommons.Automation.NeoTaskManager.Tasks;
using ECommons.Throttlers;
using ECommons.UIHelpers.AddonMasterImplementations;
using static ECommons.GenericHelpers;

public unsafe class ShopRun
{
    private readonly TaskManager TM = new(new(timeLimitMS: 20000, abortOnTimeout: true, showDebug: true));

    public void Start(string menuEntry, string confirmText)
    {
        if(TM.IsBusy) return;
        TM.EnqueueTask(NeoTasks.WaitForNotOccupied());
        TM.Enqueue(() => SelectEntry(menuEntry), $"SelectEntry({menuEntry})", new TaskManagerConfiguration());
        TM.Enqueue(() => ConfirmYesno(confirmText), "ConfirmYesno", new(timeLimitMS: 60000));
        TM.Enqueue(() => !TryGetAddonMaster<AddonMaster.SelectYesno>(out _), "WaitYesnoClosed");
        TM.EnqueueDelay(200);
        TM.EnqueueTask(NeoTasks.WaitForScreenAndPlayer());
    }

    private static bool? SelectEntry(string text)
    {
        if(!TryGetAddonMaster<AddonMaster.SelectString>(out var m) || !m.IsAddonReady) return false; // retry
        foreach(var entry in m.Entries)
        {
            if(entry.Text.Contains(text))
            {
                if(!EzThrottler.Throttle("ShopRun.SelectEntry", 500)) return false;
                entry.Select();          // Callback.Fire(addon, true, Index)
                return true;
            }
        }
        return null;                     // menu is open but the entry does not exist: abort the queue
    }

    private static bool ConfirmYesno(string text)
    {
        if(TryGetAddonMaster<AddonMaster.SelectYesno>(out var m) && m.IsAddonReady && m.Text.Contains(text))
        {
            if(EzThrottler.Throttle("ShopRun.Yes", 500))
            {
                m.Yes();
                return true;
            }
        }
        return false;
    }
}
```

Pattern notes, as used by AutoRetainer (`Scheduler/Tasks/*.cs`) and PandorasBox:

- A task is a polling function. Check state, act at most once per throttle window, return
  `false` until the game state shows the step is done. Throttle every click or callback; the
  function runs every frame.
- Two styles of completion: return `true` right after acting (above), or keep returning `false`
  after acting and return `true` only when the expected next state is visible (next addon ready,
  addon gone). The second is more robust when a click can be swallowed; it needs the throttle so
  the click is repeated at a sane rate.
- `TryGetAddonMaster<T>(out T)` derives the addon name from the type name (`SelectYesno`,
  `_TitleMenu`, ...) and creates a new wrapper through `Activator.CreateInstance` on every call;
  the source comment marks it slower than `TryGetAddonByName`. Always also check `IsAddonReady`.
- Throttle names are global to the plugin (`EzThrottler` is static); prefix them. Use
  `EzThrottler.Reset(name)` when the context changes and an immediate action is wanted.
- PandorasBox wraps the configuration in extension methods of its own
  (`tm.Enqueue(func, taskName, new() { TimeLimitMS = timeoutMs, AbortOnTimeout = true })`).

## Legacy TaskManager

Namespace `ECommons.Automation.LegacyTaskManager`. The upstream README carries this warning: as of
2024-04-15 (commit `b4be673`) `TaskManager`'s namespace has changed; add
`using ECommons.Automation.LegacyTaskManager;` as an immediate fix. Code written before that date
fails to resolve `TaskManager` until the `using` is added (the previous namespace was presumably
`ECommons.Automation` *(unverified)*). The legacy class is not marked `[Obsolete]`; only its
`Dispose()` carries an `[Obsolete]` attribute whose message says managers are disposed
automatically by `ECommonsMain.Dispose` and the method is only for earlier disposal.

```csharp
public partial class TaskManager : IDisposable
public TaskManager()
```

| Member | Notes |
|---|---|
| `public int TimeLimitMS = 10000` | Field. Captured into each task at enqueue time. |
| `public bool AbortOnTimeout = false` | Field. Captured at enqueue time. |
| `public bool TimeoutSilently = false` | Field. `true` routes timeout messages to `PluginLog.Verbose` instead of `Warning`. |
| `public bool ShowDebug = true` | Field. Debug logging is on by default. |
| `public int MaxTasks { get; private set; }` | Incremented on every enqueue, reset when idle or on `Abort()`. |
| `public long AbortAt { get; private set; }` | `Environment.TickCount64` deadline of the current task. |
| `public string CurrentTaskName` | Name of the current task or null. |
| `public List<string> TaskStack` | Names of immediate tasks then main tasks (`Union`, so duplicates collapse). |
| `public int NumQueuedTasks` | Main + immediate + current. |
| `public bool IsBusy` | Any of the above non-empty. |
| `public void Abort()` | Clears both queues and the current task. |
| `public void SetStepMode(bool enabled)`, `public void Step()` | Manual stepping; `Step()` has no step-mode check. |
| `public void Dispose()` | Unsubscribes the tick. |

Enqueue family; each of `Enqueue`, `EnqueueImmediate`, `Insert` has these eight overloads
(`name` is optional, default `null`):

```csharp
public void Enqueue(Func<bool?> task, string name = null)
public void Enqueue(Func<bool?> task, int timeLimitMs, string name = null)
public void Enqueue(Func<bool?> task, bool abortOnTimeout, string name = null)
public void Enqueue(Func<bool?> task, int timeLimitMs, bool abortOnTimeout, string name = null)
public void Enqueue(Action task, string name = null)
public void Enqueue(Action task, int timeLimitMs, string name = null)
public void Enqueue(Action task, bool abortOnTimeout, string name = null)
public void Enqueue(Action task, int timeLimitMs, bool abortOnTimeout, string name = null)
```

Delays:

```csharp
public void DelayNext(int delayMS, bool useFrameThrottler = false)
public void DelayNext(string uniqueName, int delayMS, bool useFrameThrottler = false)
public void DelayNextImmediate(int delayMS, bool useFrameThrottler = false)
public void DelayNextImmediate(string uniqueName, int delayMS, bool useFrameThrottler = false)
public void InsertDelayNext(int delayMS, bool useFrameThrottler = false)
public void InsertDelayNext(string uniqueName, int delayMS, bool useFrameThrottler = false)
```

Behavioural differences from the new manager:

- Two queues. `EnqueueImmediate` appends to a separate immediate queue that is always drained
  before the main queue. It exists so a running task can schedule follow-up steps ahead of the
  rest; there is only one level.
- There is no `Func<bool>` overload. A lambda returning plain `bool` converts to `Func<bool?>`.
- Picking the next task consumes a tick; the function first runs on the following tick.
- Timeout is checked only after the function returned `false`. The function therefore always runs
  at least once. A timeout drops the current task; with `AbortOnTimeout` both queues are cleared
  as well.
- An exception in a task is logged and only the current task is dropped. There is no
  `AbortOnError`; the queue always continues.
- `null` result: warning log and `Abort()`. No configuration object is involved, so the
  null-configuration problem of the new manager does not exist here.
- Delays are two tasks built on the static throttlers (`EzThrottler.Throttle(uniqueName, delayMS)`
  then `EzThrottler.Check(uniqueName)`, or the `FrameThrottler` equivalents). The default name
  `"ECommonsGenericDelay"` is shared by every legacy manager in the plugin; overlapping delays in
  two managers interfere unless `uniqueName` is given.
- `MaxTasks` is over-counted by delays (the two inner enqueues plus an extra `+= 2`).
- No `Progress`, no events, no per-task configuration object, no stack mode, no public task type
  (`TaskManagerTask` is `internal`).

## Migration: legacy -> new

| Legacy | New |
|---|---|
| `using ECommons.Automation.LegacyTaskManager;` | `using ECommons.Automation.NeoTaskManager;` |
| `new TaskManager() { TimeLimitMS = x, AbortOnTimeout = y, ShowDebug = z, TimeoutSilently = w }` | `new TaskManager(new(timeLimitMS: x, abortOnTimeout: y, showDebug: z, timeoutSilently: w))` |
| `tm.TimeLimitMS = x` later on | `tm.DefaultConfiguration.TimeLimitMS = x` (never assign null) |
| `Enqueue(f, "name")` | `Enqueue(f, "name")` |
| `Enqueue(f, 5000, "name")` | `Enqueue(f, "name", new(timeLimitMS: 5000))` |
| `Enqueue(f, true, "name")` | `Enqueue(f, "name", new(abortOnTimeout: true))` |
| `Enqueue(f, 5000, true, "name")` | `Enqueue(f, "name", new(timeLimitMS: 5000, abortOnTimeout: true))` |
| `Insert(f, ...)` | `Insert(f, ...)`; for several tasks use `InsertMulti` (single inserts reverse the order) |
| `EnqueueImmediate(...)` sequence | `InsertMulti(...)`, or `BeginStack()` + `Enqueue(...)` + `InsertStack()`, or `InsertStack(() => { ... })`. The source warns that a single `Insert` is not a replacement for `EnqueueImmediate`. |
| `DelayNext(ms)` / `DelayNext(n, true)` | `EnqueueDelay(ms)` / `EnqueueDelay(n, true)` ("direct replacement of DelayNext") |
| `DelayNextImmediate(ms)` | `InsertDelay(ms)`, or `EnqueueDelay(ms)` inside a stack |
| `InsertDelayNext(ms)` | `InsertDelay(ms)` |
| `DelayNext("name", ms)` | No named variant; delays no longer use the throttlers. |
| `SetStepMode(b)` | `StepMode = b` |
| `CurrentTaskName` | `CurrentTask?.Name` |
| `TaskStack` | `Tasks.Select(x => x.Name)` |
| `AbortAt` | `RemainingTimeMS` |

Defaults change when migrating: time limit 10 s -> 30 s, `AbortOnTimeout` false -> true, exceptions
continue -> abort (`AbortOnError` true), `ShowDebug` true -> false. Set them explicitly to keep old
behaviour.

## Gotchas

- Returning `null` from a task that has no per-task configuration goes through the exception path
  (`NullReferenceException` in the tick). See "Returning null".
- `With(..., copyEvents: true)` replaces event handlers rather than merging them. A handler set on
  the configuration passed to the constructor survives (it is the dominant one); handlers on the
  base of any later `With` call do not.
- `NeoTasks.ApproachObjectViaAutomove` and `InteractWithObject` call `getObjectFunc()` once while
  building the task name, that is at creation time. A delegate that throws when the object is
  absent (`Svc.Objects.First(...)`) throws at enqueue time, not inside the manager.
- A `RemainingTimeMS` value set in an object initialiser or while idle has no effect; it is
  overwritten when the next task starts. Use `TimeLimitMS` in a configuration.
- `MaxTasks`/`Progress` are only meaningful while the queue is running and are recomputed in the
  tick, so they lag enqueue calls by up to one frame.
- `Dispose()` does not clear the queue or the current task; `IsBusy` stays true on a disposed
  manager that still had tasks. Call `Abort()` first if the state is inspected later.
- Tasks run inside `Framework.Update`. Blocking calls stall the game; long work belongs elsewhere
  with the task polling for its result.
- Timeouts are wall-clock (`Environment.TickCount64`) and keep running while the game is loading
  or the task waits on a fade; use a generous `timeLimitMS` for zone changes (the ready-made wait
  tasks use 10 minutes).
- Do not import both TaskManager namespaces in one file; the type name is identical.

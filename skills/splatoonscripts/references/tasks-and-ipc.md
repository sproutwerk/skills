## TaskManager (ECommons.Automation.NeoTaskManager)

```csharp
TaskManager TaskManager;                       // own it; survives OnReset
public override void OnEnable()  => TaskManager = new(new(timeLimitMS: 30000, abortOnTimeout: true, showDebug: true));
public override void OnDisable() => TaskManager.Dispose();
```

- Tasks run on the framework thread, one per tick, in order. A task is
  `Func<bool>` (`true` = done, next; `false` = poll again next frame),
  `Func<bool?>` (`null` = abort the whole queue), or `Action` (done after
  one call).
- `Enqueue(func[, "name"][, config])`, `EnqueueDelay(ms)`,
  `EnqueueTask(TaskManagerTask)`, `EnqueueMulti(params ...)`.
  `Insert(...)` / `InsertDelay` / `InsertTask` / `InsertMulti` put tasks at
  the **front** — used inside a running task to add sub-steps, such as a
  "wait for the dialog to close" task inserted from within a "wait for
  dialog" task. `EnqueueStack(Action)` / `InsertStack(Action)` group several
  enqueues: calls made inside the action go to a stack, which is then added
  to the queue as a block, or discarded if the action throws.
  `EnqueueStack(string name, Action)` / `InsertStack(string name, Action)`
  name the stack, and stacks nest. `Stack` is a `List<StackDetails>` and
  `IsStackActive` is `Stack.Count > 0`. The manual `BeginStack()`,
  `DiscardStack()` and parameterless `EnqueueStack()` / `InsertStack()` are
  `[Obsolete]`. ECommons before October 2026 has one unnamed stack
  (`List<TaskManagerTask> Stack`) and no named overloads, so check the
  installed dependency before using either.
- `Abort()` clears everything, `AbortCurrent()` drops the running task,
  `IsBusy`, `NumQueuedTasks`, `MaxTasks`, `Progress`, `RemainingTimeMS`,
  `StepMode`/`Step()`.
- `TaskManagerConfiguration(timeLimitMS, abortOnTimeout, abortOnError,
  timeoutSilently, showDebug, showError)` — defaults 30000 / true / true /
  false / false / true. Per-task config overrides the default:
  `TaskManager.Enqueue(() => !IsBusy(), new(timeLimitMS: 300000,
  abortOnTimeout: false))`. `abortOnTimeout: false` means "discard this
  task and **run the rest of the queue anyway**"
  (`TaskManagerConfiguration.cs`), so it fits cleanup steps (close the
  shop, resume AutoDuty) and is wrong for a wait whose success the next
  step depends on — there, let the timeout abort (default) or make the
  next step re-check the state it needs. A `Func<bool?>` task returning
  `null` aborts the queue on purpose, but only reliably when the task was
  enqueued with a `TaskManagerConfiguration`: without one the tick
  dereferences `CurrentTask.Configuration.DiscardGuid`, throws, and the
  queue then aborts only because `AbortOnError` defaults to true (with
  `abortOnError: false` it keeps going). The `ecommons` skill's TaskManager topic has the details. Events: `OnTaskException`,
  `OnTaskTimeout`, `OnTaskCompletion`, `CompanionAction`.
- Ready-made tasks (`ECommons.Automation.NeoTaskManager.Tasks.NeoTasks`):
  `ApproachObjectViaAutomove(Func<IGameObject>, distance)`,
  `InteractWithObject(Func<IGameObject>, checkLos)`,
  `WaitForScreenAndPlayer()`, `WaitForNotOccupied()` — pass to
  `EnqueueTask`.
- Common wait predicates: `!GenericHelpers.IsOccupied()`,
  `GenericHelpers.IsScreenReady()` (false while NowLoading/FadeMiddle/
  FadeBack are visible), `Player.Interactable`, `!Player.IsAnimationLocked`,
  `!Player.IsCasting`, `!Svc.Condition[ConditionFlag.BoundByDuty56]`.
  Pattern for "wait for a transition": enqueue `() => !IsScreenReady()` then
  `() => IsScreenReady()`.

## EzIPC (calling other plugins)

```csharp
using ECommons.EzIpcManager;

// full names, no prefix
[EzIPC("Lifestream.IsBusy", false)] public Func<bool> LifestreamIsBusy;
[EzIPC("Lifestream.EnqueueCustomAliasFromString", false)] public Action<string, bool, int?, int?> EnqueueCustomAliasFromString;
[EzIPC("Questionable.StartSingleQuest", false)] private Func<string, bool> QuestionableStartSingleQuest;

public override void OnSetup() => EzIPC.Init(this);

// or one prefix for the whole class:
[EzIPC] public Func<bool> IsStopped;                       // -> "AutoDuty.IsStopped"
public override void OnSetup() => EzIPC.Init(this, "AutoDuty", SafeWrapper.AnyException);
```

- Field type is `Func<...>` (returns) or `Action<...>` (void); the generic
  arguments must match the provider's signature exactly (nullable ints as
  `int?`). Nullable/tuple/custom types work if both sides agree.
- Without a `SafeWrapper`, invoking a delegate whose provider is not loaded
  throws Dalamud's `IpcNotReadyError`; guard with `try/catch` or
  `SafeWrapper.AnyException` (call returns default). `EzIPC.Init` returns
  disposal tokens; scripts ignore them.
- `[EzIPCEvent] void OnSomething(int a)` subscribes to provider events.
- Reference in `ecommons/EzIPC.md`.


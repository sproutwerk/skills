## Pitfalls

Lifecycle

- **`ECommonsMain.Init` first, `ECommonsMain.Dispose` last.** Any `Svc.` access before `Init` is a
  `NullReferenceException`. Forgetting `Dispose` leaves hooks, TaskManagers, commands and IPC
  registrations alive in an unloading assembly.
- **Static state dies in `Dispose`.** `EzThrottler`, `FrameThrottler`, the `Ez*` event classes,
  `LazyData<T>.Get` and `Purgatory.Add` throw `NullReferenceException` afterwards. Background
  tasks and timers must check `ECommonsMain.Disposed`.
- **Not tracked by `Dispose`:** `TickScheduler`, `ExecuteForScheduler`, `ThreadPool`,
  `SequentialTaskQueue`, `SimpleLogger`, Splatoon dynamic elements. Clean them up yourself.
- **`Init` raises the log level to Verbose** for the plugin.
- **Module init errors are swallowed** (`GenericHelpers.Safe`) and only show in the log.

Versions

- **ECommons version must match the Dalamud API level** (3.2.x for API 15), else compile errors
  inside `ECommons/` or load failures.
- **Names change between versions without a major bump.** Examples: `Player` rework (old class in
  `LegacyPlayer`), TaskManager namespace split, `ExtractText` renamed `GetText`, `GetSafe` /
  list `GetOrDefault` replaced by `SafeSelect`, `HashSetCheckbox` replaced by `CollectionCheckbox`,
  `ImGuiLineCentered` replaced by `LineCentered`. Code copied from another plugin may target a
  different version: grep `references/all-public-members.txt` before using a name.
- **AutoRetainer code uses the legacy `Player`** through a global alias: `Player.Territory` is a
  `uint` there, a `RowRef` in the current class.
- **The README is outdated**: it writes `Modules.<Module>` (the enum is `Module`) and omits
  `VfxTracking`.

TaskManager

- **Returning `null` from a task without a per-task configuration throws** a
  `NullReferenceException` inside the tick (it reads `CurrentTask.Configuration.DiscardGuid`). With
  the default `AbortOnError = true` the queue still aborts but logs an error; with
  `abortOnError: false` only that task is dropped and the queue continues. Pass
  `new TaskManagerConfiguration()` to tasks that can return `null`.
- **Legacy and new TaskManager have different defaults** (legacy: 10 s limit, no abort on
  timeout, debug logging on, exceptions drop one task). Porting changes behaviour silently.
- **Tasks run every frame until they return true.** Throttle every click and callback in them.
- **`FrameDelayTask` inherits the manager's time limit**, so a long frame delay times out.
- **`NeoTasks.ApproachObjectViaAutomove` / `InteractWithObject` call the object getter once at
  creation time**; a throwing getter throws at enqueue time.

Addons and callbacks

- **`addon->IsReady` versus `addon->IsReady()`**: the first is the ClientStructs bitfield property,
  the second the ECommons extension. They are different checks; neither null-checks.
- **Pointers and AddonMaster wrappers are valid for the current frame only.** Fetch them again.
- **`AddonMaster.SelectYesno.Yes()` force-enables a disabled Yes button** unless
  `RespectDisabledButtons = true` is set on the wrapper. `RespectHoldButtons` is declared but
  never read at this pin.
- **Most AddonMaster click methods return `void`** and silently do nothing when the button is
  disabled or hidden.
- **`Callback.Fire` with an unsupported argument type** (anything but `uint`, `int`, `float`,
  `bool`, `string`, `AtkValue`), or a `null` argument, throws after partial conversion and may
  free uninitialised pointers. Cast literals explicitly: `3` is `int`, `3u` is `uint`, and the
  game distinguishes them.
- **`Chat.SendMessage` and `Chat.ExecuteCommand` send to the server.** `SendMessage` throws
  `ArgumentException` for empty text, more than 500 UTF-8 bytes or characters the sanitiser
  removes; `ExecuteCommand` throws `InvalidOperationException` without a leading slash. For
  Dalamud plugin commands use `Svc.Commands.ProcessCommand`.

IPC and hooks

- **`[EzIPC]` on a method makes a provider; on a delegate field or property a subscriber.**
  Subscriber fields are `null` until `EzIPC.Init(this)` ran, and stay `null` if registration of
  that member failed (errors are logged, not thrown).
- **Default `SafeWrapper` is `None`**: invoking a subscriber whose provider is not loaded throws
  Dalamud's `IpcNotReadyError`. `SafeWrapper.IPCException` catches only `IpcNotReadyError`;
  `AnyException` catches everything. Wrapped calls to an absent provider return `default`
  silently.
- **`prefix` defaults to the calling plugin's `InternalName`.** To consume another plugin's IPC,
  pass its prefix or use full names with `applyPrefix: false`.
- **Provider methods take at most 8 parameters.**
- **`[EzHookFromCS]` does not work at this pin**: `EzSignatureHelper` builds the hook but never
  assigns it to the field. Construct `new EzHook<T>(address, detour, autoEnable)` by hand.
- **`EzHook.Disable()` disposes the hook; `Pause()` is the temporary off switch.** `autoEnable`
  defaults to true. `EzSignatureHelper.Initialize` logs failures and leaves the field `null`.
  `[EzHook]` works on fields only.
- **`DirectorUpdate.Init`, `MapEffect.Init`, `SendAction.Init` accept one callback and throw on a
  second call.** `ActionEffect`, `ActorVfx`, `StaticVfx`, `GameObjectCtor` are events.
- **Exceptions in detours crash the game.** Wrap the body of every detour in try/catch.

UI and helpers

- **`EzConfigGui.Init` throws on a second call** ("ConfigGui already initialized").
- **Default `WindowType.Config` hooks only `OpenConfigUi`.** Use `WindowType.Both` or subscribe
  `OpenMainUi` yourself, as YesAlready does.
- **Save-on-close belongs to `ConfigWindow`.** A plain `Window` passed to
  `EzConfigGui.Init<T>(T window, ...)` never saves on close, and `nameOverride` is ignored by
  that overload.
- **`ThreadLoadImageHandler.TryGet*` returns false until the image is loaded**; skip drawing that
  frame. A failed load is not retried until `ClearAll()`.
- **`TryGetDalamudPlugin` with the default `ignoreCache: false` throws without
  `Module.DalamudReflector`**, outside its own try block. All three consumers read pass
  `ignoreCache: true`.
- **`GetFoP` returns `null` for a missing member, `SetFoP` throws.**
- **`ContainsIgnoreCase` is a collection-membership test**, not a substring test.
- **`GenericHelpers.UseAsyncKeyCheck` is inverted** relative to its name.
- **`Splatoon.AddDynamicElement(name, element, 5)`** binds the integer to the `long` overload
  (absolute tick count), so the element expires at once. Use `5f` for seconds.
- **`MathHelper` mixes degrees and radians** across methods → `references/topic-generic-helpers.md`.
- **Name clashes**: `Callback`, `Module`, `ThreadPool`, `UIColor`, `Player` (two classes),
  `GenericHelpers` and `EzConfig` (PunishLib has its own). Alias or qualify.
- **Banned symbols** (`ImRaii`, most `IPlayerState` members) apply only when ECommons itself is
  compiled, so only when you patch the submodule. They do not apply to plugin code.
- **Policy**: chat, callbacks and clicks are server-visible automation. What the official Dalamud
  repository accepts → skill `dalamud`, `references/topic-publishing.md`.

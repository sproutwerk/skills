## Pitfalls

- **`[PluginService]` on a field** fails to compile (property-only attribute). A property of a type
  that is not a service makes plugin creation fail at "property injection".
- **Constructor is not on the framework thread.** Reading `IObjectTable.LocalPlayer` or calling
  ClientStructs there throws or races. Defer with `framework.RunOnFrameworkThread(...)` or do it in
  the first `Update`.
- **`RunOnFrameworkThread` vs `RunOnTick`**: the first runs inline when already on the thread; the
  second always waits for a later tick. Use `RunOnTick` to break recursion or to delay (`delayTicks:
  1`).
- **Blocking on tasks from the framework thread** (`.Result`, `.Wait()`, `GetAwaiter().GetResult()`)
  freezes or deadlocks the game. Use `IAsyncDalamudPlugin` or fire-and-forget with error logging.
- **Leaked subscriptions**: every `+=`, `AddHandler`, `RegisterListener`, `Subscribe`,
  `RegisterFunc`, hook, DTR entry and `AddEvent` needs its counterpart in `Dispose`. A reload that
  doubles your output is this.
- **A hook does nothing until `Enable()`**, including hooks filled in by `[Signature]`.
- **Exceptions in detours crash the game.** Dalamud catches and logs exceptions from `Update`
  handlers, lifecycle listeners and command handlers; it cannot wrap your hook.
- **Non-nullable `[Signature]` member + broken sig = plugin fails to load** (`SignatureException`).
  Make members nullable and degrade gracefully; signatures break on patches.
- **Holding game objects**: `IGameObject`, `IStatus`, `IPartyMember`, `IFate`, `IAetheryteEntry` are
  views over game memory valid for the current frame. Store ids (`EntityId`, `GameObjectId`,
  `ContentId`).
- **`GameObjectId` vs `EntityId`**: `SearchById` takes the 64-bit `GameObjectId`; `SearchByEntityId`
  the 32-bit id. Empty id is `0xE0000000`. `ObjectId`/`DataId` are old names.
- **Logged-in checks**: `IPlayerState` is valid from login to logout and is not tied to the player
  object (v14 notes), so `IObjectTable.LocalPlayer` can be `null` while it is loaded. Check the one
  you actually use.
- **`TerritoryChanged` can fire before objects exist** *(community knowledge, unverified)*. Wait for
  `ICondition[ConditionFlag.BetweenAreas]` to clear before reading the object table.
- **Renaming the configuration class or assembly** orphans saved configs (`$type`). Keep the name or
  migrate by hand.
- **`Lumina.Excel.Sheets.Action` / `Status` clash** with `System.Action` and Dalamud's status types
  (Dalamud itself writes them fully qualified); alias with `using LuminaAction =
  Lumina.Excel.Sheets.Action;`.
- **`RowRef.Value` on an invalid ref throws**; use `ValueNullable` or check `IsValid`.
- **IPC across load contexts**: do not pass your own classes or enums; type mismatches throw
  `IpcTypeMismatchError`. Wrap subscriber calls in try/catch for `IpcError`; the provider may unload
  at any time.
- **Docs lag the source.** Known stale spots in the pinned docs: `faq/development.mdx`
  (`[RequiredVersion]`, concrete service classes), `how-tos/AddonEventManager.md` (old handler
  signature, `->ID`, `AtkStage.GetSingleton()`), `how-tos/AddonLifecycle.md` (`AtkValueSpan`). Trust
  `references/dalamud/` over `references/docs/`.
- **Do not bundle Dalamud-provided assemblies** (Dalamud, Lumina, ClientStructs, Newtonsoft.Json,
  ImGui bindings); the SDK references them with `Private=false`.

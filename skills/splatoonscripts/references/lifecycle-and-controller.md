## Lifecycle (verified in `ScriptingProcessor.cs` / `SplatoonScript.cs`)

Enable/disable is decided by `UpdateState()`, run after load, after the
blacklist check, and on every `TerritoryChanged`:

```
territoryIsValid = ValidTerritories == null
                || (Svc.ClientState.IsLoggedIn
                    && (ValidTerritories.Count == 0 || ValidTerritories.Contains(Svc.ClientState.TerritoryType)))
enabled = territoryIsValid && !user-disabled
```

So `null` = enabled always, even at the title screen / logged out (use for
login hooks, `ProperOnLogin`); `[]` = enabled in any zone while logged in;
a set = only those TerritoryType IDs. Moving between two valid zones does
not re-run OnEnable/OnDisable.

| Hook | When | Notes |
|---|---|---|
| `OnSetup()` | once, right after compile | register elements/layouts, `EzIPC.Init(this)`, `EzSignatureHelper.Initialize(this)`. **No matching teardown** — do not subscribe to Dalamud events here. |
| `OnEnable()` | territory becomes valid | subscribe events, create your own `TaskManager`, enable hooks. After it returns `IsEnabled = true` and `OnReset()` runs. |
| `OnDisable()` | territory becomes invalid, user disables, script reload, plugin unload | unsubscribe, dispose. Before it: `Controller.SaveConfig()`, `OnReset()`, and `Controller.TaskManager` (the host-owned one) is disposed. |
| `OnUpdate()` | every framework tick while enabled | the main loop. Exceptions are caught and logged per call; the script keeps running. |
| `OnReset()` | after enable; before disable; on combat start/end; on director commence/recommence/wipe; when `Controller.ScheduleReset()` fires; on `Controller.Reset()` | host also cancels queued commands, `Controller.Schedule` timers and **aborts `Controller.TaskManager`**. |
| `OnCombatStart()` / `OnCombatEnd()` | `Controller.InCombat` flips | `OnReset` runs first. |
| `OnSettingsDraw()` | Splatoon's script settings UI | draw ImGui; overriding it makes the settings section appear (`DoSettingsDraw`). |
| `OnScriptUpdated(uint previousVersion)` | after a reinstall with a version bump | |
| `OnMessage(string)` | every chat/log message Splatoon's trigger system sees | e.g. `AutoFateSync` matches "You are 5 or more levels above..." |
| `OnDirectorUpdate(DirectorUpdateCategory)` / `(nint directorPtr, uint targetId, DirectorUpdateCategory a3, uint a4, uint a5, int a6, int a7, int a8, int a9)` | duty director events (commence, wipe, ...) | |
| `OnPhaseChange(int)` | Splatoon phase counter changes | |
| `OnMapEffect(uint position, ushort data1, ushort data2)` | | duty scripts |
| `OnObjectEffect(uint target, uint entityId, uint actionId)` | | |
| `OnTetherCreate(uint source, uint target, uint data2, uint data3, uint data5)` / `OnTetherRemoval(uint source, uint data2, uint data3, uint data5)` | | |
| `OnVFXSpawn(uint target, string vfxPath)` | | |
| `OnStartingCast(uint source, uint castId)` / `unsafe OnStartingCast(uint sourceId, PacketActorCast* packet)` | hostile cast starts | |
| `OnObjectCreation(nint newObjectPtr)` | | |
| `OnActorControl(uint sourceId, uint command, uint p1..p8, ulong targetId, byte replaying)` | every ActorControl packet | volatile data |
| `OnActionEffectEvent(ActionEffectSet set)` | an ActionEffect packet arrives (the action is announced) | `OnActionEffect(...)` is obsolete |
| `OnActionEffectApplied(ActionEffectSet set, int targetIndex, ActionEffectResult result)` | each entry of an EffectResult / EffectResultBasic packet, when HP, shield and statuses actually change (usually 0.5-2 s after the announcement) | once per target, and once more if the source is affected (`result.IsSource`). Check `result.HasActionEffect` first: when false, `set` is empty (DoT tick, fall damage, HP sync, or an action seen before load). Read effects from `result.Effects`, never `set.TargetEffects`. Fields in `host/ActionEffectResult.cs`. Not in hosts older than 3.9.2.32. |
| `OnGainBuffEffect(uint sourceId, Status)` / `OnRemoveBuffEffect` / `OnUpdateBuffEffect` | a status is gained, removed or updated | found by comparing statuses each frame by default. The user can switch them to hook-based events (the experimental "Status events from hooks" setting, off by default and applied on plugin reload); those fire at the moment of change and a removed status then has no remaining time. |

Other overridable members:

- `Metadata Metadata` (abstract): `new(uint version)`, `new(version,
  author)`, `new(version, author, description)`, `new(version, author,
  description, website)`.
- `Dictionary<int,string>? Changelog` — shown to the user on update.
- `bool Safe` (default false) — declares the script only uses Splatoon /
  ECommons / Dalamud APIs.
- `TaskManagerConfiguration TaskManagerConfiguration` — default
  `new(timeLimitMS: 30000, showDebug: true)`; used for `Controller.TaskManager`.

Base-class helpers: `BasePlayer` (local player, replay-aware; the analyzer
in the official csproj flags `Player.Object` with RS0030 — add
`#pragma warning disable RS0030` for non-duty scripts, Splatoon compiles it
either way), `EzThrottler` / `FrameThrottler` (**per-script instance
properties** of type `EzThrottler<string>`; with `using
ECommons.Throttlers;` the bare name `EzThrottler.Throttle(...)` binds to the
instance property, official scripts still prefix keys with
`InternalData.FullName` — do the same), `Loc(en, jp, de, fr, cn)`,
`InternalData` (`FullName`, `Name`, `Namespace`, `Path`, `GUID`, and
`Overrides`, publicly readable since 3.9.2.32),
`IsEnabled`, `Controller`.

## Controller (`host/Controller.cs`)

| Member | Use |
|---|---|
| `TaskManager` | host-owned NeoTaskManager, created on first access from `TaskManagerConfiguration`, disposed on disable, **aborted on every OnReset (combat start/end!)**. Fine for short UI chains; for a multi-minute automation that spans combat, own a `TaskManager` field instead (AutoTeleport and QuestionableQuestQueue both do). |
| `GetConfig<T>()` / `SaveConfig()` | load `Foo.cs.json` into `T : new()` (Newtonsoft via EzConfig) / write it. Saved automatically on disable. Call `SaveConfig()` after important edits; do not rely on the automatic save if the game crashes. |
| `Schedule(Action, int delayMs)` / `CancelSchedulers()` | one-shot timers, cleared on reset. |
| `ScheduleReset(uint delayMs)` | reset the script after a delay. |
| `Reset()` | run OnReset + cleanup now. |
| `DangerousEnqueueCommand(string text, bool test)` / `CancelQueuedCommands()` | send chat lines to the **server** like a macro (multi-line allowed, each must start with `/`). Not sent during replays. |
| `DisplayAttentionWindowLine(string)` / `(Vector4? color, string, params string[] args)` / `(Action)` / `DisplayAttentionWindowRaw(Action)` | Splatoon's attention overlay. Must be called **every frame** to stay visible; args substitute `$1`, `$2`. |
| `InCombat`, `CombatSeconds`, `CombatMiliseconds`, `Phase`, `Scene`, `AttentionColor`, `RolePosition` | state |
| `GetPartyMembers()` | `IEnumerable<IPlayerCharacter>` (non-cross-world) |
| `GetMapEffect(uint index)` / `GetMapEffect<TEnum>(...)` | current map-effect state in the instance director |
| `TryRegisterElement(name, Element, overwrite)`, `RegisterElement(...)`, `TryRegisterElementFromCode(name, exportString, out Element)`, `RegisterElementFromCode(...)`, `RegisterElementsFromMultilineCode(...)`, `TryGetElementByName`, `GetElementByName`, `TryUnregisterElement`, same set for `Layout` (`TryRegisterLayout`, `TryRegisterLayoutFromCode`, `TryGetLayoutByName`, ...), `GetRegisteredElements()`, `GetRegisteredLayouts()`, `ClearRegisteredElements/Layouts()`, `Clear()`, `Hide(elements, layouts)`, `OriginalElements`, `OriginalLayouts`, `ApplyOverrides()`, `SaveOverrides()` | drawing — duty-script territory; register in `OnSetup`, toggle `Enabled` at runtime |
| `Plugin` | the `Splatoon` instance (`Splatoon.Splatoon.P` is the same); `P.NotificationMasterApi.DisplayTrayNotification(title, text)`, `.FlashTaskbarIcon()`, `.PlaySound(path)` |

`SplatoonScript<T>` is a convenience base that exposes `T C =>
Controller.GetConfig<T>()`. Most official scripts instead declare
`Config C => Controller.GetConfig<Config>();` themselves. `IEzConfig` is
obsolete ("no longer required") but harmless; a plain class with public
fields works.


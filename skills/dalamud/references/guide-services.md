## Services

All in namespace `Dalamud.Plugin.Services` (45 interfaces) except `IDragDropManager`
(`Dalamud.Interface.DragDrop`) and `IDalamudAssetManager` (`Dalamud.Storage.Assets`), which move
there in API 16. Full signatures: `references/dalamud/Dalamud/Plugin/Services/<Name>.cs`.

Compact catalog: name and purpose only. Exact members, signatures, events and per-service gotchas →
`references/topic-services.md`. Read it before writing a call you have not verified.

### Game state

| Service | Purpose |
|---|---|
| `IClientState` | Login state, territory / map / instance ids, client language, PvP and GPose flags, and their change events. `LocalPlayer` and `LocalContentId` were **removed in API 15**: use `IObjectTable.LocalPlayer` / `IPlayerState.ContentId`. |
| `IPlayerState` | Local character data (name, ids, worlds, job, levels, attributes). Valid from login to logout; gate on `IsLoaded`. |
| `IObjectTable` | Game objects, `LocalPlayer`, search by id. **Main thread only**: throws `InvalidOperationException` elsewhere. Do not keep `IGameObject` across frames. |
| `ITargetManager` | Get/set the target, focus target, soft target, mouse-over targets. Setting a target is a game action; mind the publishing rules. |
| `ICondition` | `ConditionFlag` lookup (`this[ConditionFlag]`) and the `ConditionChange` event. |
| `IPartyList` | Party members. |
| `IBuddyList` | Companion and pet buddies. |
| `IFateTable` | FATEs. |
| `IAetheryteList` | The aetherytes available in the Teleport window. |
| `IJobGauges` | Job gauge data via `Get<T>()`. |
| `IDutyState` | Duty started / wiped / recommenced / completed events. |
| `IUnlockState` | Unlock checks (mounts, items, quests, emotes, recipes …). Takes Lumina **rows**, not ids. |
| `IGameInventory` | Inventory contents and inventory change events. |
| `IGameConfig` | Read and write game settings. `Set` changes the user's real game settings. |
| `IKeyState` | Keyboard state by `VirtualKey`. Settable; writes change what the game reads. |
| `IGamepadState` | Gamepad buttons and sticks. |
| `IGameLifecycle` | Cancellation tokens (`DalamudUnloadingToken`, `GameShuttingDownToken`, `LogoutToken`) to pass to long-running tasks. |

Object model: `IGameObject` → `ICharacter` → `IBattleChara` → `IPlayerCharacter` / `IBattleNpc`.
Also `INpc`, `IEventObj`. Pattern-match to narrow: `if (obj is IBattleChara bc)`. Member lists →
`references/topic-services.md`

### Game UI and text

| Service | Purpose |
|---|---|
| `IChatGui` | Print to the chat log, chat events, chat link handlers. Prints to the local log only; it never sends chat to the server. → `references/topic-sestring-chat.md` |
| `IGameGui` | Addon and agent lookup (`GetAddonByName`, addon index is 1-based), world ↔ screen coordinates, open the map, hovered item/action. |
| `IAddonLifecycle` | Listen to native window events by addon name. → `references/topic-addons.md` |
| `IAddonEventManager` | Attach custom events to native nodes. → `references/topic-addons.md` |
| `IAgentLifecycle` | Listen to agent events (`RegisterListener` / `UnregisterListener`). |
| `IContextMenu` | Add items to game context menus. |
| `IDtrBar` | Server info bar entries. |
| `IToastGui` | Show and intercept game toasts. |
| `IFlyTextGui` | Add and intercept fly text. |
| `INamePlateGui` | Nameplate update events. |
| `IPartyFinderGui` | Party Finder listing event. |
| `ITitleScreenMenu` | Title screen menu entries. |
| `INotificationManager` | ImGui toast, not the game toast. |
| `ISeStringEvaluator` | Resolves macros the way the game would display them. |

### Data, interop, infrastructure

| Service | Purpose |
|---|---|
| `IDataManager` | Lumina Excel sheets and game files. See Lumina section. |
| `IGameInteropProvider` | Creates hooks and resolves `[Signature]` members. The `HookBackend` parameter is obsolete: never pass it (removed in API 16). |
| `ISigScanner` | Signature scanning. `Scan*` throw `KeyNotFoundException` on miss; use `Try*`. |
| `IFramework` | Framework tick and thread marshalling. See Threading. |
| `ICommandManager` | Chat commands. See Commands. |
| `IPluginLog` | Logging. See Logging. |
| `IReliableFileStorage` | File storage where every write is duplicated into a backup store; do not write large files. `SavePluginConfig` already goes through it. |
| `IMarketBoard` | Market board events. |
| `IConsole` | Console (`/xllog`) commands and variables, not chat commands. `[Experimental("Dalamud001")]`. |
| `ISelfTestRegistry` | Steps shown in Dalamud's self-test window. |
| `ITextureProvider`, `ITextureReadbackProvider`, `ITextureSubstitutionProvider`, `IDragDropManager`, `IDalamudAssetManager` | UI/texture services → `dalamudimgui`. |

Removed services: `IGameNetwork` (API 13; hook the handling functions instead), `ILibcFunction`,
static `PluginLog`, `SeStringManager`, `ChatHandlers` (API 9/10).


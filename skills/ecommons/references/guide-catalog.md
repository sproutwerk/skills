## Catalog

One line per area. Folders are under `references/ecommons/ECommons/`; topics under `references/topic-*.md`.

| Folder | Holds | Topic |
|---|---|---|
| `ECommonsMain.cs`, `Module.cs` | init, dispose, module flags | `setup.md` |
| `DalamudServices/` | `Svc`; `Legacy/` has `GameNetwork` and compatibility shims | `game-hooks.md` |
| `ImGuiMethods/ImGuiEx/` | the `ImGuiEx` partial class | `imguiex-core.md`, `imguiex-widgets.md` |
| `ImGuiMethods/` (rest), `Funding/` | `Notify`, `EColor`, `EzColor`, `ImGuiBB`, `ThreadLoadImageHandler`, overlay windows, `PatreonBanner` | `imgui-helpers.md` |
| `SimpleGui/`, `EzCmd.cs` | `EzConfigGui`, `ConfigWindow`, `PopupWindow`, `EzCmd` | `simplegui.md` |
| `Configuration/` | `EzConfig`, serialization factories, `EzCharaConfig<T>` | `configuration.md` |
| `Automation/NeoTaskManager/` | current `TaskManager`, `NeoTasks` | `taskmanager.md` |
| `Automation/LegacyTaskManager/` | old `TaskManager` (`DelayNext`, `EnqueueImmediate`) | `taskmanager.md` |
| `Automation/` (rest) | `Chat`, `Callback`, `UIInput`, `MacroManager`, `WindowsKeypress` | `automation.md` |
| `UIHelpers/` | `AddonMaster.*`, `AtkReader`, `AddonFinder` | `addonmaster.md` |
| `EzIpcManager/`, `EzSharedDataManager/` | `EzIPC`, attributes, `SafeWrapper`, `EzSharedData` | `ezipc.md` |
| `EzHookManager/` | `EzHook<T>`, `EzSignatureHelper`, `EzDelegate`, `EzPatch` | `ezhook.md` |
| `Hooks/`, `Opcodes/`, `ObjectLifeTracker/`, `GameHelpers/VfxManager.cs` | built-in game hooks, `OpcodeUpdater` (obsolete), `ObjectLife`, `VfxManager` | `game-hooks.md` |
| `EzEventManager/`, `Events/`, `Throttlers/`, `Schedulers/`, `Singletons/`, `LazyDataHelpers/`, `Loader/` | `Ez*` events, `ProperOnLogin`, throttlers, `TickScheduler`, `SingletonServiceManager`, `Purgatory` | `events-throttlers-schedulers.md` |
| `GameHelpers/`, `GameFunctions/`, `PartyFunctions/`, `CSExtensions/`, `Gamepad/` | `Player`, `Content`, object and character extensions, `ExtendedPronoun`, `UniversalParty` | `game.md` |
| `ExcelServices/`, `TerritoryName.cs` | `Excel*Helper`, `Job` and other enums, territory constants | `excel.md` |
| `GenericHelpers/` | the `GenericHelpers` partial class | `generic-helpers.md`; addon, Excel and log parts in `addonmaster.md`, `excel.md`, `logging-commands.md` |
| `MathHelpers/`, `CircularBuffers/`, `Collections/`, `Numeric/`, `StringHelpers/` | `MathHelper`, `Angle`, `BitHelper`, `ENullable<T>`, `CircularBuffer<T>`, `OneOrMany<T>`, `EqualStrings` (obsolete) | `generic-helpers.md` |
| `Logging/`, `Commands/`, `ChatMethods/`, `EzDTR/`, `EzContextMenu/`, `SimpleLogger.cs` | logs, `[Cmd]`, `ChatPrinter`, `EzDtr` | `logging-commands.md` |
| `Reflection/`, `Interop/`, `WindowsFormsReflector/`, `LanguageHelpers/`, `Networking/` | `DalamudReflector`, `ReflectionHelper`, Win32 helpers, `Keys`, `Localization`, proxy | `reflection-interop.md` |
| `SplatoonAPI/` | `Splatoon`, `Element` | `splatoonapi.md` |

No class named `ProcessPacket`, `ProcessZonePacket` or `ProcessActorControlPacket` exists at this
pin. Packet access is `Svc.GameNetwork`; ActorControl-style events are covered by
`DirectorUpdate` and the other `Hooks` classes.


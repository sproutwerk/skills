## Topic files

Deep material lives in `references/topic-*.md` (first-party, same pin). Open the file when its trigger applies;
the sections below are summaries and are not enough to write non-trivial code from.

| File | Read this when ... |
|---|---|
| `references/topic-setup.md` | adding ECommons, choosing NuGet or submodule, a build cannot find Dalamud inside `ECommons/`, picking a version for an API level, you need the exact effect of `Init`, a `Module` flag or `Dispose` |
| `references/topic-examples.md` | you want a complete skeleton plugin or need to locate a worked example |
| `references/topic-imguiex-core.md` | using `ImGuiEx` text, tooltips, centering, width and scale helpers, `EzTabBar`, `EzTable`, `BeginDefaultTable`, collapsing headers, `Pagination`, `DragDropRepopulate`, `Filtering*` inputs |
| `references/topic-imguiex-widgets.md` | using `ImGuiEx` buttons, checkbox variants, `CollectionCheckbox`, `EnumCombo` / `Combo`, numeric and text inputs, sliders, `ExcelSheetCombo`, `RealtimeDragDrop` |
| `references/topic-imgui-helpers.md` | using `Notify`, `EColor` / `EzColor`, `ImGuiBB`, `ThreadLoadImageHandler` (icons, images), overlay windows, `TerritorySelector`, `Ref<T>` |
| `references/topic-simplegui.md` | creating the plugin window with `EzConfigGui` / `ConfigWindow`, adding windows, registering commands with `EzCmd`, save-on-close behaviour |
| `references/topic-configuration.md` | loading or saving with `EzConfig`, migrating from Dalamud's config, custom serializers, per-character or extra config files, a config that is lost or not saved |
| `references/topic-taskmanager.md` | queuing multi-step automation, `TaskManagerConfiguration`, timeouts and aborts, stacks, `NeoTasks`, porting from the legacy TaskManager |
| `references/topic-automation.md` | sending chat or commands (`Chat`), firing addon callbacks (`Callback.Fire`), clicking nodes (`UIInput`), macros, key presses, cutscene skipping |
| `references/topic-addonmaster.md` | reading or driving a native game window: `TryGetAddonByName`, `TryGetAddonMaster`, `IsAddonReady`, any `AddonMaster.X` class, `AtkReader` |
| `references/topic-ezipc.md` | providing or consuming IPC with `[EzIPC]` / `[EzIPCEvent]`, `SafeWrapper`, channel naming, `EzSharedData`, the `ECommons.IPC` package |
| `references/topic-ezhook.md` | hooking with `EzHook<T>`, `[EzHook]`, `EzSignatureHelper`, `EzDelegate`, patching memory with `EzPatch` |
| `references/topic-game-hooks.md` | subscribing to `ActionEffect`, `DirectorUpdate`, `MapEffect`, `SendAction`, VFX hooks, `GameObjectCtor`, `VfxManager`, `ObjectLife`, or reading packets through `Svc.GameNetwork` |
| `references/topic-events-throttlers-schedulers.md` | throttler details, `EzFrameworkUpdate` / `EzLogout` / `EzTerritoryChanged`, `ProperOnLogin`, `TickScheduler`, `SingletonServiceManager`, `LazyData`, `Purgatory` |
| `references/topic-game.md` | `Player` and the legacy `Player`, `Content`, `CharacterFunctions`, `ObjectFunctions`, `ExtendedPronoun`, `UniversalParty`, `RenderDisableManager`, the banned-symbol list |
| `references/topic-excel.md` | sheet helpers (`GetRow`, `FindRow`, `GetSheet`, `GetRef`), `Excel*Helper` classes, the `Job` enum, territory constants |
| `references/topic-generic-helpers.md` | any `GenericHelpers` member or extension (strings, collections, `Safe`, clipboard, keys, `IsOccupied`), `MathHelper`, `Angle`, `BitHelper`, `CircularBuffer` |
| `references/topic-logging-commands.md` | `PluginLog`, `DuoLog`, `InternalLog` and its viewer, exception `.Log()` extensions, `[Cmd]`, `ChatPrinter`, `EzDtr` |
| `references/topic-reflection-interop.md` | `DalamudReflector`, `GetFoP` / `SetFoP` / `Call`, Win32 helpers, `OpenFileDialog`, `WindowsFormsReflector.Keys`, `Localization` |
| `references/topic-splatoonapi.md` | drawing Splatoon elements from a regular plugin through `ECommons.SplatoonAPI` |


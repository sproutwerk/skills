Dalamud topic: the full `IDalamudPluginInterface` member table. Verified against `goatcorp/Dalamud` `43b3411` (2026-11-10, `DalamudVersion 15.0.3.9`, API 15), `goatcorp/dalamud-docs` `3c7597d` (2026-11-13), `goatcorp/SamplePlugin` `b8477da` (2026-08-15) and `aers/FFXIVClientStructs` `80cb4a7` (2026-11-10). Reviewed 2026-11-13.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that holds `SKILL.md`), not to this file. Part of the `dalamud` skill; items marked *(unverified)* are not confirmed from the pinned checkouts.

## IDalamudPluginInterface

| Member | Use |
|---|---|
| `InternalName`, `Manifest` (`IPluginManifest`), `Reason`, `IsDev`, `IsTesting`, `IsInProfile`, `SourceRepository`, `WorkingPluginId`, `IsAutoUpdateComplete` | Identity and load context |
| `AssemblyLocation` (`FileInfo`), `Assembly` (`System.Reflection.Assembly`, added in 15.0.3.x) | Where the DLL is (use for bundled files) and the loaded plugin assembly. An older API 15 runtime lacks `Assembly`. |
| `ConfigFile` (`FileInfo`), `ConfigDirectory` (`DirectoryInfo`), `GetPluginConfigDirectory()` (creates it, returns path), `GetPluginLocDirectory()` | Storage |
| `GetPluginConfig()`, `SavePluginConfig(IPluginConfiguration?)` | Configuration, see `references/guide-configuration.md` |
| `UiBuilder` | ImGui hookup, see `references/guide-idalamudplugininterface.md` |
| `UiLanguage`, `event LanguageChanged(string langCode)` | Dalamud UI language |
| `InstalledPlugins` (`IEnumerable<IExposedPlugin>`), `GetPlugin(Assembly)`, `event ActivePluginsChanged(IActivePluginsChangedEventArgs)` | Other plugins: `Name`, `InternalName`, `IsLoaded`, `Version`, `IsThirdParty`, `HasMainUi`, `OpenMainUi()`, `OpenConfigUi()` |
| `GetIpcProvider<…>(name)`, `GetIpcSubscriber<…>(name)` | IPC, see `references/topic-ipc.md` |
| `GetOrCreateData<T>`, `TryGetData<T>`, `GetData<T>`, `RelinquishData(tag)` | DataShare, see `references/topic-ipc.md` |
| `GetDalamudVersion()` → `IDalamudVersionInfo` (`Version`, `BetaTrack`, `GitHash`, `GitHashClientStructs`, `ScmVersion`) | Replaces the removed `Util` version helpers (API 14) |
| `OpenPluginInstallerTo(kind, searchText)`, `OpenDalamudSettingsTo(kind, searchText)`, `OpenDeveloperMenu()` | Open Dalamud windows |
| `IsDevMenuOpen`, `IsDebugging`, `Sanitizer`, `GeneralChatType`, `AllowSeasonalEvents`, `LoadTime`, `LoadTimeUTC`, `LoadTimeDelta`, `DalamudAssetDirectory`, `CheckForUpdateAsync()` | Misc |

`IUiBuilder` hookup points are listed in `references/guide-idalamudplugininterface.md`; drawing itself is the `dalamudimgui` skill.

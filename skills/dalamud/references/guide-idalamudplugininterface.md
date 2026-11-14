## IDalamudPluginInterface

Most used: `InternalName`, `Reason`, `IsDev`, `AssemblyLocation` (`FileInfo`, use for bundled
files), `ConfigFile`, `ConfigDirectory`, `GetPluginConfig()`,
`SavePluginConfig(IPluginConfiguration?)`, `UiBuilder`, `GetIpcProvider<…>(name)`,
`GetIpcSubscriber<…>(name)`, `InstalledPlugins`, `GetDalamudVersion()` (replaces the removed `Util`
version helpers, API 14). Full member table → `references/topic-plugin-interface.md`

### IUiBuilder hookup points (drawing itself → `dalamudimgui`)

`pluginInterface.UiBuilder`: events `Draw`, `OpenConfigUi` (installer cog button), `OpenMainUi`
(installer "Open" button), `ShowUi`, `HideUi`, `ResizeBuffers`, `DefaultGlobalScaleChanged`,
`DefaultFontChanged`, `DefaultStyleChanged`. Flags: `DisableAutomaticUiHide`, `DisableUserUiHide`,
`DisableCutsceneUiHide`, `DisableGposeUiHide`, `OverrideGameCursor`. State: `CutsceneActive`,
`ShouldModifyUi`, `UiPrepared`, `FrameCount`, `PluginUISoundEffectsEnabled`. `WaitForUi()`,
`RunWhenUiPrepared<T>(…)`. Unsubscribe every event in `Dispose`. The plugin validator expects
`OpenConfigUi`/`OpenMainUi` handlers when the plugin has such windows.


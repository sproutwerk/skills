## IUiBuilder: hooking up drawing

`IDalamudPluginInterface.UiBuilder` (`IUiBuilder`). `Draw` fires once per
frame from Dalamud's Present hook on the game's main thread.

| Member | Notes |
|---|---|
| `event Action? Draw` | Draw here. Exceptions are caught, logged, and replace your UI with a "`<plugin>` Error" window until dismissed |
| `event Action? OpenConfigUi` / `OpenMainUi` | Subscribing is what enables the Settings / Open buttons in the plugin installer (`HasConfigUi => OpenConfigUi != null`). Plugin validation warns when missing *(unverified here; the validator is outside the copied files)* |
| `event Action? ShowUi` / `HideUi` | Fired when Dalamud starts/stops hiding your UI; may fire consecutively |
| `DefaultGlobalScaleChanged`, `DefaultFontChanged`, `DefaultStyleChanged` | Invalidate cached sizes here |
| `DisableAutomaticUiHide`, `DisableUserUiHide`, `DisableCutsceneUiHide`, `DisableGposeUiHide` | While hidden, `Draw` is **not invoked at all** |
| `CutsceneActive`, `ShouldModifyUi`, `UiPrepared`, `FrameCount` | State |
| `ShouldUseReducedMotion`, `PluginUISoundEffectsEnabled` | User preferences to respect |
| `OverrideGameCursor` | Whether ImGui's cursor replaces the game's |
| `FontAtlas`, `CreateFontAtlas(autoRebuildMode, isGlobalScaled = true, debugName)` | Plugin-private atlas, see `references/topic-fonts.md` |
| `DefaultFontHandle`, `IconFontHandle`, `IconFontFixedWidthHandle`, `MonoFontHandle` | `IFontHandle`, safe to hold |
| `FontDefault`, `FontIcon`, `FontIconFixedWidth`, `FontMono` | `ImFontPtr`, valid only inside `Draw` |
| `DefaultFontSpec`, `FontDefaultSizePt`, `FontDefaultSizePx` | Current default font settings |
| `LoadUld(string uldPath)` | `UldWrapper` for multi-icon ULD textures |
| `WaitForUi()`, `RunWhenUiPrepared<T>(func, runInFrameworkThread = false)` | Await UI readiness during load |
| `DeviceHandle`, `WindowHandlePtr` | D3D11 device / HWND as `nint` |

Static equivalents on the concrete class: `UiBuilder.DefaultFont`,
`UiBuilder.IconFont`, `UiBuilder.MonoFont`, `UiBuilder.IconFontFixedWidth`
(`ImFontPtr`, draw-time only), `UiBuilder.DefaultFontSizePx`,
`UiBuilder.DefaultFontSizePt`.


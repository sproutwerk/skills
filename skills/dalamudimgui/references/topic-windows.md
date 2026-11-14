# WindowSystem and Window in depth

Covers the per-frame window lifecycle order, every `Window` property (persistence, pinning,
title bar buttons, size constraints), `PreDraw`/`PostDraw` styling and drawing without
`WindowSystem`. Read the [window example](guide-windowsystem-and-window.md) for plugin integration.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

Per-frame order inside `WindowSystem.Draw()` for each window (from
`WindowHost.DrawInternal`):

1. `PreOpenCheck()`: always, even when closed.
2. If `!IsOpen`: on the transition call `OnClose()`, play the close sound,
   run the fade-out, then `OnSafeToRemove()`; return.
3. `Update()`: every frame while open, even when collapsed.
4. `DrawConditions()`: `false` skips this frame without triggering `OnClose`.
5. `ImGui.PushID(Namespace)` if set; on the transition `OnOpen()`.
6. `PreDraw()`, then `Position`/`Size`/`Collapsed`/`SizeConstraints`/`BgAlpha`
   are applied as `SetNextWindow*`.
7. `ImGui.Begin(WindowName, ref open, Flags)`; `Draw()` runs only if Begin
   returned true. An exception in `Draw()` is caught and logged, and the
   window then shows an error message instead of calling `Draw()` again.
8. Title bar buttons, focus/hover bookkeeping, Escape and gamepad close.
9. `ImGui.End()`, then `PostDraw()`.

| Property | Default | Notes |
|---|---|---|
| `WindowName` | ctor | Title and ID. `AddWindow` throws `ArgumentException` on duplicates within the system. Use `Title###StableId` when the title changes |
| `IsOpen` | `false` | `Toggle()` flips it; `BringToFront()` sets `RequestFocus` if open |
| `Flags` | ctor | `ImGuiWindowFlags` |
| `Size`, `SizeConstraints` | `null` | **Unscaled**; multiplied by `GlobalScale`. `MaximumSize` smaller than `MinimumSize` (e.g. left at zero) is treated as unlimited |
| `Position`, `PositionCondition`, `SizeCondition`, `Collapsed`, `CollapsedCondition`, `BgAlpha` | unset | Position is not scaled; with `ForceMainWindow` it is relative to the main viewport |
| `ForceMainWindow` | `false` | Keeps the window inside the game window (no separate viewport) |
| `RespectCloseHotkey` | `true` | Escape closes the focused window (if the user enabled focus management) |
| `ShowCloseButton` | `true` | |
| `AllowPinning`, `AllowClickthrough`, `AllowBackgroundBlur` | `true` | User-facing options in the title bar menu; state persists per window ID |
| `IsPinned`, `IsClickthrough`, `IsTopMost` | | Read as `false` when `AllowPinning` (pinned) or `AllowClickthrough` (the other two) is off |
| `DisableWindowSounds`, `OnOpenSfxId` (23), `OnCloseSfxId` (24), `DisableFadeInFadeOut` | | |
| `TitleBarButtons` | empty | `List<TitleBarButton>`: `Icon`, `IconColor`, `IconOffset`, `ShowTooltip`, `Click` (`Action<ImGuiMouseButton>`), `Priority`, `AvailableClickthrough` |
| `IsFocused`, `IsHovered` | | Updated by the system each frame |

`PreDraw`/`PostDraw` are the place for style that must wrap `Begin` itself
(window padding, rounding, title colors). Hold the `ImRaii` object in a field:

```csharp
private ImRaii.StyleDisposable? style;
public override void PreDraw()  => this.style = ImRaii.PushStyle(ImGuiStyleVar.WindowPadding, Vector2.Zero);
public override void PostDraw() => this.style?.Dispose();
```

`IWindow`/`IWindowSystem` exist if you need your own base class; `Window`
is the stock implementation. Popups, tooltips and child windows inside
`Draw()` are plain ImGui/ImRaii.

Without `WindowSystem` you can still call `ImGui.Begin`/`End` yourself from
`UiBuilder.Draw`, but you lose scaling of sizes, pin/click-through,
persistence, close hotkey handling, sounds and the exception guard.

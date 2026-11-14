# ImGuiHelpers and ImGuiComponents

Covers the full `ImGuiHelpers` table (scaling, centering, viewports, blur, key mapping, font
helpers), `ImGuiExtensions`, and every `ImGuiComponents` widget.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

## Global scale and ImGuiHelpers

`ImGuiHelpers.GlobalScale` is the user's Dalamud UI scale
(`ImGui.GetIO().FontGlobalScale`, or the configured value before ImGui is
up). Fonts scale automatically; **every pixel constant you write does not**.

| Helper | Use |
|---|---|
| `GlobalScale` | Multiply hard-coded widths, heights, offsets, radii |
| `ScaledVector2(x)`, `ScaledVector2(x, y)`, `ScaledVector4(x, y, z, w)` | Scaled vectors |
| `ScaledDummy(size)`, `ScaledDummy(x, y)`, `ScaledDummy(Vector2)` | Scaled spacer |
| `ScaledIndent(size)`, `ScaledRelativeSameLine(offset, spacing = -1)` | Scaled layout |
| `GetButtonSize(text)` | `CalcTextSize + FramePadding * 2` |
| `CenterCursorFor(itemWidth)`, `CenterCursorForText(text)`, `CenteredText(text)` | Horizontal centering |
| `ClickToCopyText(text, textCopy = default, color = null)` | Text that copies on click, with tooltip |
| `BeginHorizontalButtonGroup()` | Fluent `.Add(text, action)`, `.SetCentered`, `.WithHeight`, then `.Draw()` |
| `MainViewport`, `CheckIsWindowOnMainViewport()`, `ForceNextWindowMainViewport()` | Multi-viewport control |
| `SetNextWindowPosRelativeMainViewport(pos, cond, pivot)`, `SetWindowPosRelativeMainViewport(name, pos, cond)` | Position relative to the game window |
| `AddBlurBehind(drawList, min, max, ...)`, `PrependBlurBehind(...)` | Blur-behind region; use `Prepend` from inside a window |
| `SeStringWrapped(...)`, `CompileSeStringWrapped(...)`, `CenteredSeString(...)` | See `references/topic-sestring.md` |
| `DefaultColorPalette(swatchCount = 32)` | Palette for color pickers |
| `ImGuiKeyToVirtualKey`, `VirtualKeyToImGuiKey` | Key mapping |
| `IsImGuiInitialized` | Guard for code that may run before the first frame |
| `CreateImGuiRangesFrom(UnicodeRange...)`, `BuildRangesToArray`, `NewFontGlyphRangeBuilderPtrScoped`, `CopyGlyphsAcrossFonts`, `AdjustGlyphMetrics`, `IsNotNullAndLoaded`, `OrElse` | Font building helpers |

Rules of thumb: derive sizes from ImGui metrics where possible
(`ImGui.GetFrameHeight()`, `GetTextLineHeightWithSpacing()`,
`GetStyle().ItemSpacing`, `CalcTextSize`) since those already include the
scale; multiply everything else. `Window.Size`/`SizeConstraints` and the
`size` argument of `ImGuiComponents.IconButton*` are scaled for you, so do
not pre-multiply those. Values stored in config should be unscaled.

`ImGuiExtensions.AddTextClippedEx(this ImDrawListPtr, posMin, posMax, text,
textSizeIfKnown, align, clipRect)` is the only extension there.

## ImGuiComponents

`Dalamud.Interface.Components.ImGuiComponents` (all `string` parameters, not
`ImU8String`):

| Method | Notes |
|---|---|
| `IconButton(FontAwesomeIcon icon)` | Also `(int id, icon)`, `(string id, icon)`, `(string iconText)`; each with `Vector2 size` or `(Vector4? defaultColor, activeColor, hoveredColor, Vector2? size)` |
| `IconButtonWithText(icon, text, size)` / `(icon, text, defaultColor?, activeColor?, hoveredColor?, size?)` | Icon plus label in one button |
| `GetIconButtonWithTextWidth(icon, text)` | For right-aligning |
| `IconButtonSelect<T>(label, ref val, optionIcons, optionValues, columns = 0, buttonSize?, ...)` | Icon radio group; also a `KeyValuePair<FontAwesomeIcon, T>` overload |
| `HelpMarker(helpText)` / `(helpText, icon, color?)` | Calls `SameLine()` itself, then draws a disabled icon with a wrapped tooltip |
| `ToggleButton(string id, ref bool v)` | The toggle switch (file is `ToggleSwitch.cs`, there is no `ToggleSwitch` method). Returns true when changed |
| `DisabledToggleButton(id, v)` | Non-interactive version |
| `DisabledButton(labelWithId, ...)` / `(FontAwesomeIcon icon, int? id, ...)` | Only *looks* disabled (alpha); still returns clicks. Use `ImRaii.Disabled` for real disabling |
| `ColorPickerWithPalette(int id, string description, Vector4 originalColor[, flags])` | Returns the new color |
| `TextWithLabel(label, value, hint = "")` | `label: value`, tooltip when `hint` set |

Icon buttons with the same icon in one ID scope collide: use the `id`
overloads or wrap rows in `ImRaii.PushId(i)`.

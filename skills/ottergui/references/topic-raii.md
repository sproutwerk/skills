OtterGui topic: `OtterGui.Raii.ImRaii` (Color, Style, ColorStyle, Font, Id, Indent, EndObjects) and
exactly how it differs from Dalamud's built-in `Dalamud.Interface.Utility.Raii.ImRaii`. Verified
against `Limiana/OtterGui` `a08fda6` (2026-05-30, branch `main`; `Raii/` is identical to upstream
`79771ee`) and `goatcorp/Dalamud` `43b3411` (`Dalamud/Interface/Utility/Raii/`).

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. Nothing was compiled.

## The class

```csharp
using OtterGui.Raii;   // public static partial class ImRaii
```

Seven files in `references/src/Raii/`: `Color.cs`, `ColorStyle.cs`, `Font.cs`, `Id.cs`,
`Indent.cs`, `Style.cs`, `EndObjects.cs`. Every `string` parameter is a managed `string`; for
UTF-8 and interpolated strings use `ImUtf8` instead (`references/topic-imutf8.md`).

## Stacks

| Factory on `ImRaii` | Returns | Members of the returned object |
|---|---|---|
| `PushColor(ImGuiCol idx, uint color, bool condition = true)`, `PushColor(ImGuiCol idx, Vector4 color, bool condition = true)` | `ImRaii.Color` (sealed class) | `Push(idx, uint / Vector4, condition)` chainable, `Pop(int num = 1)`, `Dispose()` |
| `DefaultColors()` | `ImRaii.Color` | pushes, for every colour currently changed through this `ImRaii`, the value it had before the first change |
| `PushStyle(ImGuiStyleVar idx, float value, bool condition = true)`, `PushStyle(ImGuiStyleVar idx, Vector2 value, bool condition = true)` | `ImRaii.Style` (sealed class) | `Push(...)` chainable, `Pop(int num = 1)`, `Dispose()`, static `GetStyle(ImGuiStyleVar idx)` returning `Vector2` (`Y` is `NaN` for float variables) |
| `DefaultStyle()` | `ImRaii.Style` | same idea as `DefaultColors()` |
| `PushFrameBorder(float thickness, uint color, bool condition = true)` | `ImRaii.ColorStyle` (readonly struct) | pushes `ImGuiCol.Border` and `ImGuiStyleVar.FrameBorderSize`; `Push(ImGuiStyleVar, float / Vector2, condition)`, `Push(ImGuiCol, uint, condition)`, `Dispose()` |
| `PushFont(ImFontPtr font, bool condition = true)` | `ImRaii.Font` (sealed class) | `Push(font, condition)`, `Pop(int num = 1)`, `Dispose()` |
| `DefaultFont()` | `ImRaii.Font` | pushes the font that was current before the first `ImRaii.PushFont`, only if some font is currently pushed |
| `PushId(string id, bool enabled = true)`, `PushId(int id, bool enabled = true)`, `PushId(nint id, bool enabled = true)` | `ImRaii.Id` (sealed class) | `Push(string / int / nint, condition)`, `Pop(int num = 1)`, `Dispose()` |
| `PushIndent(float f, bool scaled = true, bool condition = true)`, `PushIndent(int i = 1, bool condition = true)` | `ImRaii.Indent` (sealed class) | `Indentation`, `Push(...)`, `Pop(float indent, bool scaled = true)`, `Pop(int i)`, `Dispose()` |

Notes read from the code:

- `PushIndent(float)` multiplies by `ImGuiHelpers.GlobalScale` when `scaled` is true;
  `PushIndent(int)` multiplies by `ImGui.GetStyle().IndentSpacing`.
- `Pop(num)` clamps to what this object pushed, so over-popping is harmless.
- `Style.Push` validates the value type against the style variable only in `DEBUG` builds
  (`[Conditional("DEBUG")] CheckStyleIdx`); in Release a `float` pushed to a `Vector2` variable is
  passed through to ImGui unchecked. The switch covers `Alpha` through `DisabledAlpha` (25
  variables); any other `ImGuiStyleVar` makes `GetStyle` throw `ArgumentOutOfRangeException`.
- Pattern used throughout the library: push, draw, `Pop`, draw more, push again. See
  `ItemSelector<T>.Draw` (`references/src/ItemSelector.cs`), which pops three style variables
  inside a child and re-pushes them before the child ends.

```csharp
using var style = ImRaii.PushStyle(ImGuiStyleVar.FrameRounding, 0)
    .Push(ImGuiStyleVar.ItemSpacing, Vector2.Zero);
using var color = ImRaii.PushColor(ImGuiCol.FrameBg, 0x803030A0, !allSelected);
```

## End objects (`Raii/EndObjects.cs`)

All return `ImRaii.IEndObject` (`interface IEndObject : IDisposable` with `bool Success`, operators
`true`, `false`, `!`, `&`, `|`, and `static readonly IEndObject Empty`).

| Function | End call | When |
|---|---|---|
| `Child(string strId)`, `(strId, Vector2 size)`, `(strId, size, bool border)`, `(strId, size, border, ImGuiWindowFlags flags)` | `EndChild` | always |
| `Group()`, `Tooltip()` | `EndGroup`, `EndTooltip` | always |
| `ItemWidth(float width)` | `PopItemWidth` | always |
| `CollapsingHeader(string label)` | `PopID` | always; pushes `label` as ID after drawing the header |
| `Disabled()`, `Disabled(bool disabled)` | `EndDisabled` | always when it began |
| `Enabled()` | re-begins the disabled scopes it ended | temporarily leaves every `ImRaii.Disabled` scope |
| `FramedGroup(string label, uint borderColor = 0, uint headerColor = 0, FontAwesomeIcon headerPreIcon = FontAwesomeIcon.None)`, `FramedGroup(string label, Vector2 minSize, string description = "", uint borderColor = 0, uint headerColor = 0, FontAwesomeIcon headerPreIcon = FontAwesomeIcon.None)` | `Widget.EndFramedGroup(borderColor)` | always |
| `DragDropTarget()`, `DragDropSource()`, `DragDropSource(ImGuiDragDropFlags flags)` | `EndDragDropTarget` / `EndDragDropSource` | on success |
| `Popup(string id)`, `Popup(id, ImGuiWindowFlags flags)` | `EndPopup` | on success |
| `ContextPopupItem()`, `ContextPopupItem(string label)`, `ContextPopupItem(label, ImGuiPopupFlags flags)` | `EndPopup` | on success (`BeginPopupContextItem`) |
| `ContextPopup(string id)`, `ContextPopup(id, ImGuiPopupFlags flags)` | `EndPopup` | on success (`BeginPopupContextWindow`) |
| `Combo(string label, string previewValue)`, `(label, previewValue, ImGuiComboFlags flags)` | `EndCombo` | on success |
| `ListBox(string label)`, `(label, Vector2 size)` | `EndListBox` | on success |
| `Table(string table, int numColumns)`, `(.., ImGuiTableFlags flags)`, `(.., flags, Vector2 outerSize)`, `(.., flags, outerSize, float innerWidth)` | `EndTable` | on success |
| `TabBar(string label)`, `(label, ImGuiTabBarFlags flags)` | `EndTabBar` | on success |
| `TabItem(string label)`, `TabItem(byte* label, ImGuiTabItemFlags flags)`, `(string label, ref bool open)`, `(label, ref open, ImGuiTabItemFlags flags)` | `EndTabItem` | on success |
| `TreeNode(string label)`, `(label, ImGuiTreeNodeFlags flags)` | `TreePop` | on success; nothing when `NoTreePushOnOpen` is set |

There is no `TabItem(string label, ImGuiTabItemFlags flags)` overload: flags without `ref open`
need the `byte*` overload (this is what `Widgets/TabBar.cs` does with `fixed`), or use
`ImUtf8.TabItem(label, flags)`.

## Differences from Dalamud's ImRaii

Same class name, different namespace, different implementation. Dalamud side read from
`Dalamud/Interface/Utility/Raii/ImGui.cs` and `Disposables/*.cs` at `43b3411`.

| Aspect | `OtterGui.Raii.ImRaii` | `Dalamud.Interface.Utility.Raii.ImRaii` |
|---|---|---|
| String parameters | `string` | `ImU8String` (accepts `string`, UTF-8 spans, interpolated strings) |
| Scope return type | `IEndObject` interface; the implementations are private structs, so each call boxes | one concrete type per scope (`ChildDisposable`, `TableDisposable`, `PopupDisposable`, ...) |
| Stack return types | `ImRaii.Color`, `Style`, `Font`, `Id`, `Indent` | `ColorDisposable`, `StyleDisposable`, `FontDisposable`, `IdDisposable`, `IndentDisposable` |
| `Pop` on Color/Style | returns `void` | returns the disposable (chainable) |
| Extra stack members | none | `Count`, `PushDefault(type)`, static `PopUnsafe(num)`, static `PushDefaultColors()` / `PushDefaultStyle()` |
| `PushColor` with `Vector4?` | no | yes |
| Only in OtterGui | `CollapsingHeader(string)`, `FramedGroup(...)`, `PushFrameBorder` / `ColorStyle`, parameterless `ContextPopupItem()` | |
| Only in Dalamud | | `ChildFrame`, `Columns`, `Header`, `PopupModal`, `PushClipRect` (returns `ClipRectDisposable`), `Menu`, `MenuBar`, `MainMenuBar`, `TextWrapPos`, `ItemWidth(width, condition)`, `TabItem(label, flags)`, the ImPlot scopes (`Plot`, `AlignedPlots`, `LegendPopup`, `Subplots`, `DragDropSource*`, `DragDropTarget*`, `PushStyle(ImPlotStyleVar, ...)`, `PushColor(ImPlotCol, ...)`) |
| Bookkeeping | private static lists inside the OtterGui assembly (`Color.Stack`, `Style.Stack`, `Font.FontPushCounter`, `_disabledCount`) | its own, inside Dalamud |

OtterGui's `ImUtf8` covers several of the Dalamud-only scopes under other names: `ImUtf8.Modal`,
`ImUtf8.Menu`, `ImUtf8.Columns`, `ImUtf8.TabItem(label, flags)`, `ImUtf8.PushClipRect`.

### What the separate bookkeeping means

`DefaultColors()`, `DefaultStyle()`, `DefaultFont()` and `Enabled()` only know about pushes made
through the **same** `ImRaii`. If a colour was pushed with Dalamud's `ImRaii` (or raw
`ImGui.PushStyleColor`), OtterGui's `DefaultColors()` does not revert it, and the other way round.
`ImRaii.Enabled()` from OtterGui ends only the disabled scopes OtterGui itself began; its counter
does not see `ImGui.BeginDisabled()` calls made elsewhere. Pick one `ImRaii` per window where
these functions matter. `Widgets/Tutorial.cs` relies on this: it draws its popup inside
`DefaultStyle()`, `DefaultColors()` and `DefaultFont()`.

### Name clash

With both namespaces imported, every `ImRaii.X` is ambiguous (CS0104). Options:

```csharp
using ImRaii = OtterGui.Raii.ImRaii;                  // alias one of them
using OtterGui.Raii;                                  // or import only one
using var style = Raii.ImRaii.PushStyle(...);         // partial qualification, as Table/ColumnString.cs does
```

OtterGui itself imports `Dalamud.Interface.Utility.Raii` in ten files (`Classes/MessageService.cs`,
`Widgets/HexViewer.cs`, `Text/ImUtf8.Text.cs` and seven files under `Text/Extended/`), so both
implementations are in use inside the library.

## ImUtf8 end objects versus ImRaii end objects

| | `ImRaii.IEndObject` | `OtterGui.Text.EndObjects.*` |
|---|---|---|
| Kind | interface over a struct (heap allocation per scope) | `ref struct` (no allocation) |
| Strings | `string` | UTF-8 bytes, chars or interpolated |
| Can be stored in a field or captured by a lambda | yes | no |

For new code in a plugin that already uses OtterGui, `ImUtf8` scopes are the cheaper choice; for
everything else Dalamud's own `ImRaii` is enough and avoids the dependency (sibling skill
`dalamudimgui`).

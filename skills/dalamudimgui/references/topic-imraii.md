# ImRaii catalog

Covers every `ImRaii` factory: Begin/End scopes, push stacks, ImPlot scopes, and their rules.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

`Dalamud.Interface.Utility.Raii.ImRaii`: factory methods returning disposable
scopes. Two families with different rules.

## Begin/End scopes (`ref struct`, stack only)

| Factory | Wraps | `if (x)` needed | End called |
|---|---|---|---|
| `Child(strId[, size[, border[, flags]]])` | `BeginChild` | yes | always |
| `ChildFrame(uint id, size[, flags])` | `BeginChildFrame` | yes | always |
| `Table(id, numColumns[, flags[, outerSize[, innerWidth]]])` | `BeginTable` | yes | on success |
| `TabBar(label[, flags])` | `BeginTabBar` | yes | on success |
| `TabItem(label[, flags])`, `TabItem(label, ref open[, flags])` | `BeginTabItem` | yes | on success |
| `TreeNode(label[, flags])` | `TreeNodeEx` | yes | `TreePop` on success, skipped with `NoTreePushOnOpen` |
| `Header(label, flags)`, `Header(label, ref visible, flags)` | `CollapsingHeader` + `PushID(label)` | yes | `PopID` on success |
| `Combo(label, previewValue[, flags])` | `BeginCombo` | yes | on success |
| `ListBox(label[, size])` | `BeginListBox` | yes | on success |
| `Popup(id[, flags])` | `BeginPopup` | yes | on success |
| `PopupModal(id[, flags])`, `PopupModal(id, ref open[, flags])` | `BeginPopupModal` | yes | on success |
| `ContextPopup(id[, flags])` | `BeginPopupContextWindow` | yes | on success |
| `ContextPopupItem(id[, flags])` | `BeginPopupContextItem` | yes | on success |
| `Menu(label[, enabled])`, `MenuBar()`, `MainMenuBar()` | `BeginMenu*` | yes | on success |
| `DragDropSource([flags])`, `DragDropTarget()` | `BeginDragDrop*` | yes | on success |
| `Group()` | `BeginGroup` | no bool conversion | always |
| `Tooltip()` | `BeginTooltip` | no bool conversion | always |
| `Columns(count, id, border = false)` | legacy `Columns` | no bool conversion | restores previous column count |

The pattern, including the early-out that keeps nesting flat:

```csharp
using var table = ImRaii.Table("items"u8, 3, ImGuiTableFlags.RowBg | ImGuiTableFlags.ScrollY);
if (!table)
    return;                      // Dispose still runs; EndTable is skipped because Begin failed

ImGui.TableSetupColumn("Name"u8, ImGuiTableColumnFlags.WidthStretch);
ImGui.TableSetupColumn("Qty"u8, ImGuiTableColumnFlags.WidthFixed, 60 * ImGuiHelpers.GlobalScale);
ImGui.TableSetupColumn("##actions"u8, ImGuiTableColumnFlags.WidthFixed, ImGui.GetFrameHeight());
ImGui.TableHeadersRow();
```

Nested form: `using (var tab = ImRaii.TabItem("General"u8)) if (tab) DrawGeneral();`

Every checked scope has `Success`, `Alive`, implicit `bool`, `!`, `&`, `|`
and a static `EndUnsafe()`/`PopUnsafe()` escape hatch. `Dispose()` is
idempotent, so calling it early to close a scope before the block ends is
fine (`ImGuiClip` does this with `Group`).

Skipping the check draws into the parent or asserts in native code. For
`Child` the body is merely wasted work when clipped, but check anyway.

## Push stacks (`sealed class`, chainable, storable)

| Factory | Returns | Chain with |
|---|---|---|
| `PushColor(ImGuiCol, uint \| Vector4 \| Vector4?, bool condition = true)` | `ColorDisposable` | `.Push(idx, color[, condition])`, `.Pop(n)` |
| `PushStyle(ImGuiStyleVar, float \| Vector2, bool condition = true)` | `StyleDisposable` | `.Push(idx, value[, condition])`, `.Pop(n)` |
| `PushFont(ImFontPtr, bool condition = true)` | `FontDisposable` | `.Push(font[, condition])`, `.Pop(n)` |
| `PushId(ImU8String \| int \| nint, bool enabled = true)` | `IdDisposable` | `.Push(id[, enabled])`, `.Pop(n)` |
| `PushIndent(float f, bool scaled = true, bool condition = true)` | `IndentDisposable` | `.Indent(...)`, `.Unindent(...)` |
| `PushIndent(int i = 1, bool condition = true)` | `IndentDisposable` | multiples of `IndentSpacing` |
| `ItemWidth(float width[, bool condition])` | `ItemWidthDisposable` | `.Push(width)`, `.Pop(n)` |
| `TextWrapPos(float pos[, bool condition])` | `TextWrapDisposable` | `.Push(localX)`, `.Pop(n)` |
| `Disabled()`, `Disabled(bool disabled)` | `DisabledDisposable` | `.Push([condition])`, `.Pop(n)` |
| `PushClipRect(in Rectangle \| in Vector2 topLeft, in Vector2 bottomRight, bool intersectWithCurrentClipRect = false[, bool condition])` | `ClipRectDisposable` | `.Push(...)`, `.Pop(n)` |
| `Enabled()` | `EnabledDisposable` | Temporarily ends every `ImRaii.Disabled` scope; restores on dispose |
| `DefaultColors()`, `DefaultStyle()`, `DefaultFont()` | | Push the values that were active before any ImRaii push |

```csharp
using var color = ImRaii.PushColor(ImGuiCol.Button, ImGuiColors.DalamudRed, isDangerous)
                        .Push(ImGuiCol.ButtonHovered, ImGuiColors.DalamudRed.Lighten(0.1f), isDangerous);
using var style = ImRaii.PushStyle(ImGuiStyleVar.FramePadding, new Vector2(8, 4) * ImGuiHelpers.GlobalScale)
                        .Push(ImGuiStyleVar.FrameRounding, 4f);
using var disabled = ImRaii.Disabled(!canDelete);
if (ImGui.Button("Delete"u8)) Delete();
```

- One object tracks its own count; `Dispose()` pops exactly what it pushed,
  including zero when every `condition` was false. `PushColor` with a `null`
  `Vector4?` pushes nothing.
- `StyleDisposable.Push` validates the value type against the style var and
  throws `ArgumentException` for a mismatch (`float` into `FramePadding`).
- These are heap objects: one small allocation per `ImRaii.Push*` call per
  frame. Chain pushes on one object; in very hot loops use raw
  `ImGui.PushStyleColor`/`PopStyleColor`.
- `PushIndent(float)` scales by `GlobalScale` unless `scaled: false`.
- `EnabledDisposable` only knows about `ImRaii.Disabled`, not raw
  `ImGui.BeginDisabled`.
- ImPlot equivalents live in the same class (`Raii/Plot.cs`): `Plot`,
  `AlignedPlots`, `LegendPopup`, `Subplots`, `DragDropSourceItem/Axis/Plot`,
  `DragDropTargetAxis/Legend/Plot`, `PushStyle(ImPlotStyleVar, ...)`,
  `PushColor(ImPlotCol, ...)`, `DefaultPlotStyle()`, `DefaultPlotColors()`.

`Header(label, ref visible, flags)` leaked its `PushID(label)` on Dalamud builds
before `15.0.3.8` (the constructor never set `Alive`, so `Dispose()` returned
early). The bundled revision has the fix. On an older runtime use
`Header(label, flags)` or a manual `ImGui.CollapsingHeader(label, ref visible)`.

## Clip rectangles

`ImRaii.PushClipRect` wraps `ImGui.PushClipRect` / `PopClipRect` and takes screen
coordinates, either as two corners or as a `Rectangle`:

```csharp
var min = ImGui.GetCursorScreenPos();
using (ImRaii.PushClipRect(min, min + size, intersectWithCurrentClipRect: true))
    ImGui.GetWindowDrawList().AddText(min, color, text);
```

`Dalamud.Interface.Utility.Rectangle` is a `readonly record struct (Vector2 Minimum,
Vector2 Maximum)` that converts implicitly to and from `ImRect`. Members: `Size`, `Width`,
`Height`, `Center`, `TopLeft` and `BottomRight` (the two corners not stored), `IsInverted`, `FromSize(...)`, `Contains`,
`Add`, `Expand`, `Overlaps`, `OverlapsLine`, `Orient()`. `LineEvaluator` in the same
namespace has `ClosestPointOnLine` and `Evaluate`. `PushClipRect`, `Rectangle` and `LineEvaluator` first shipped in `15.0.3.8`.

There is no `ImRaii.Window`; windows go through `WindowSystem` or raw
`Begin`/`End`.

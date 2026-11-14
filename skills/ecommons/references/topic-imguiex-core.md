# ImGuiEx core: text, tooltips, layout, tables, tabs, trees

Covers: the core half of `ECommons.ImGuiMethods.ImGuiEx` (files `ImGuiEx.cs`, `Text.cs`,
`Alignment.cs`, `EzTable.cs`, `Extensions.cs`, plus the two-member `ObsoleteOverloads.cs`). Verified
against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths
are relative to the skill root.

Not covered here: buttons, checkboxes, collection checkboxes, inputs/sliders, `EnumCombo`, `Combo`,
`EnumRadio`, `RadioButtonBool`, Excel combos, `RealtimeDragDrop` -> `references/topic-imguiex-widgets.md`.
`Ref<T>`, `EColor`, `EzColor`, `ImGuiDragDrop`, `Notify`, overlay windows, `ThreadLoadImageHandler`
-> `references/topic-imgui-helpers.md`. Stock Dalamud ImGui (bindings, `ImRaii`, `Window`) -> skill
`dalamudimgui`.

```csharp
using Dalamud.Bindings.ImGui;
using Dalamud.Interface;          // FontAwesomeIcon
using ECommons.ImGuiMethods;      // ImGuiEx, EColor, EzColor
using System.Numerics;
```

`ImGuiEx` is one `public static partial class` in namespace `ECommons.ImGuiMethods`. Source:
`references/ecommons/ECommons/ImGuiMethods/ImGuiEx/` (`ImGuiEx.cs`, `Text.cs`, `Alignment.cs`,
`EzTable.cs`, `Extensions.cs`, `ObsoleteOverloads.cs`).

General facts that apply to the whole class:

- Everything takes `string`, not `ImU8String`/interpolated handlers (exception: `DragDouble`). Text
  is always drawn with `ImGui.TextUnformatted`, so `%` in strings is safe.
- Colours are `Vector4` (RGBA floats) or `uint` (passed straight to `ImGui.PushStyleColor`, i.e.
  ImGui's packed ABGR, not 0xRRGGBB). Most `Vector4?` overloads treat `null` as "no colour push".
- All members must be called from the draw thread inside an ImGui frame. No `Module` flag and no
  hooks are involved.
- There are no disabled/ID/style RAII scopes in these files; use `ImRaii` (skill `dalamudimgui`).
  The only style helpers are the multi-push methods in "Style and ID helpers" below, and they need
  a manual pop.
- There are no hotkey capture widgets in these files; only the modifier properties and
  `IsKeyPressed` in "Keys and mouse" below.

## Static state and caches

| Member | Kind | Notes |
|---|---|---|
| `public static readonly ImGuiWindowFlags OverlayFlags` | constant | `NoNav \| NoMouseInputs \| NoDecoration \| NoBackground \| NoSavedSettings \| NoFocusOnAppearing` |
| `public static ImGuiTableFlags DefaultTableFlags` | mutable | `NoSavedSettings \| RowBg \| Borders \| SizingFixedFit`; used by `BeginDefaultTable`; process-wide for your copy of ECommons |
| `SelectedPages` (private `Dictionary<string,int>`) | cache | selected page per `Pagination` id |
| `RightFloatWidthCache`, `CenteredLineWidths`, `InputWithRightButtonsAreaValues` (private `Dictionary<string,float>`) | cache | previous-frame measured widths, keyed by id |
| `JobSelectorFilter` (private `string`) | shared | one filter string shared by every `JobSelector` |
| `Ref<T>` storage | cache | used by the `Filtering*` inputs and `DragDropRepopulateClass` |

The private dictionaries are never trimmed and are not touched by `ECommonsMain.Dispose()`. `Ref<T>`
registers itself with `Purgatory` (`Purgatory.Add(typeof(Ref<T>))`), and `ECommonsMain.Dispose()`
calls `Purgatory.Purge`, which nulls its static storage.

Overloads without an `id` parameter derive the key from `GenericHelpers.GetCallStackID()`, which
builds a `StackTrace` on every call (documented in source as "Expensive") and yields the same key
for every call made from the same method chain. Two id-less calls in one method, or one call in a
loop, share a key and corrupt each other's cached width. Pass an explicit id in anything that is
drawn per frame in a loop.

## Text (`Text.cs`)

| Signature | Behaviour |
|---|---|
| `void Text(string s)` | `ImGui.TextUnformatted(s)` |
| `void Text(Vector4 col, string s)` / `Text(Vector4? col, string text)` | pushes `ImGuiCol.Text`; `null` draws uncoloured |
| `void Text(uint col, string s)` | packed ImGui colour |
| `void Text(EzColor col, string s)` | uses `col.Vector4` |
| `void Text(ImFontPtr font, string s)` | push/pop font |
| `void Text(uint col, ImFontPtr font, string s)` | font and colour |
| `void Text(Vector4? col, ImFontPtr? font, string s)` | both optional |
| `void TextV(string s)` / `TextV(Vector4? col, string s)` | `ImGui.AlignTextToFramePadding()` first; use next to buttons/inputs and in table cells |
| `void TextWrapped(string s)` / `(Vector4? col, string s)` / `(uint col, string s)` | `PushTextWrapPos` + unformatted text; wraps at the content region edge |
| `void TextCentered(string text)` / `(Vector4 col, string text)` / `(Vector4? col, string text)` | offsets cursor X by `GetContentRegionAvail().X / 2 - CalcTextSize(text).X / 2`; single line only |
| `void CenterColumnText(string text, bool underlined = false)` / `(Vector4 colour, string text, bool underlined = false)` | centres using `ImGui.GetColumnWidth()`; for table/column cells |
| `void TextUnderlined(string text)` / `(uint color, string text)` / `(Vector4 color, string text)` | draws a line under the text via the window draw list |
| `void TextCopy(string displayText, string? copyText = null)` / `(Vector4 col, string text, string? copyText = null)` | hand cursor on hover; left click calls `GenericHelpers.Copy(copyText ?? displayText)` |
| `void TextWrappedCopy(string text)` / `(Vector4 col, string text)` | wrapped text, copy via `HoveredAndClicked()` |
| `void CodeText(string s)` | drawn with `UiBuilder.MonoFont` |
| `void Icon(FontAwesomeIcon icon)` / `Icon(Vector4 col, FontAwesomeIcon icon)` | icon in `UiBuilder.IconFont` |
| `void IconWithText(FontAwesomeIcon icon, string s)` / `(Vector4 col, FontAwesomeIcon icon, string s)` / `(Vector4 col, FontAwesomeIcon icon, string? s = null, string? tooltip = null)` | icon, `SameLine`, then text |
| `void IconWithTooltip(FontAwesomeIcon icon, string tooltip)` / `(Vector4 col, FontAwesomeIcon icon, string? tooltip = null)` | icon plus hover tooltip |

Notes:

- `GenericHelpers.Copy` uses the WinForms clipboard and shows a `Notify.Success("Text copied to
  clipboard")` toast, so `TextCopy` gives user feedback with no extra code.
- `TextCopy` reacts on `IsItemClicked` (mouse down); `TextWrappedCopy` reacts on mouse release
  through `HoveredAndClicked`.
- The underline in `TextUnderlined` is always stroked with `ImGuiColors.DalamudWhite`; the colour
  argument only affects the text.
- `IconWithText` always calls `ImGui.SameLine()` after the icon, even when `s` is `null` (that is
  the path `Icon` and `IconWithTooltip` take). The next item you draw lands on the same line;
  call `ImGui.NewLine()` or `ImGui.Dummy` if that is not wanted. The tooltip attaches to the last
  item drawn: the text when `s` is given, otherwise the icon.
- String trimming extensions in `ImGuiEx.cs`:
  `string ImGuiTrim(this string str, int minTrimLength = 5)` cuts the string so that it plus "..."
  fits `GetContentRegionAvail().X` (strings shorter than `minTrimLength` are returned untouched;
  cost is one `CalcTextSize` per character), and `string Trim(this string str, int trimToLength)`
  cuts to a character count and appends "...". The latter is an extension named `Trim` on
  `string`; it does not clash with `string.Trim()` because of the `int` argument.

```csharp
ImGuiEx.Text(EColor.RedBright, "Not configured");
ImGuiEx.TextV($"Retainers: {count}");                 // aligned with the button that follows
ImGui.SameLine();
ImGui.Button("Refresh");
ImGuiEx.TextWrapped(ImGuiColors.DalamudGrey, longDescription);
ImGuiEx.TextCopy($"{item.RowId}");                    // click to copy
```

## Tooltips and help markers

| Signature | Behaviour |
|---|---|
| `void Tooltip(string s)` / `Tooltip(Vector4? col, string s)` | if the previous item is hovered, shows a wrapped tooltip |
| `void SetTooltip(string text)` / `SetTooltip(Vector4? col, string text)` | unconditional `BeginTooltip` + text wrapped at `GetFontSize() * 35f` + `EndTooltip` |
| `void HelpMarker(string helpText, Vector4? color = null, string symbolOverride = null, bool sameLine = true, bool preserveCursor = false)` | forwards to `InfoMarker` |
| `void InfoMarker(string helpText, Vector4? color = null, string symbolOverride = null, bool sameLine = true, bool preserveCursor = false)` | FontAwesome `InfoCircle` with hover tooltip |
| `bool HoveredAndClicked(string tooltip = null, ImGuiMouseButton btn = ImGuiMouseButton.Left, bool requireCtrl = false)` | makes the previous item clickable |

Differences from stock:

- `ImGuiEx.Tooltip` replaces the usual `if(ImGui.IsItemHovered()) ImGui.SetTooltip(...)` pair and
  wraps long text. Stock `ImGui.SetTooltip` does not wrap.
- `HelpMarker`/`InfoMarker` call `ImGui.SameLine()` themselves by default (`sameLine: true`), so
  do not add your own. Default colour is `ImGuiColors.DalamudGrey3`. `symbolOverride` is a string
  rendered in the icon font (pass `FontAwesomeIcon.X.ToIconString()`). With `preserveCursor: true`
  the marker is drawn with `SameLine(0, 0)`, shifted left by its own width, and the cursor is put
  back, so it overlays the end of the previous item without consuming layout space. `helpText ==
  null` draws the icon without a tooltip.
- `HoveredAndClicked` returns `true` on the frame the mouse button is released over the item
  (`ImGui.IsMouseReleased(btn)`), only while the drag delta is below 2 px on both axes. While
  hovered it sets the hand cursor and shows `tooltip` if given. With `requireCtrl` the click only
  counts while CTRL is held.

```csharp
ImGui.Checkbox("Enable", ref C.Enabled);
ImGuiEx.HelpMarker("Explains the option.");          // drawn on the same line

ImGuiEx.Text(character.Name);
if(ImGuiEx.HoveredAndClicked("Click to relog")) Relog(character);
if(ImGuiEx.HoveredAndClicked("Right click to clear", ImGuiMouseButton.Right)) Clear(character);
```

## Layout, alignment, width and scale

### Scale (extension methods)

```csharp
public static float    Scale(this float f)
public static float?   Scale(this float? f)
public static Vector2  Scale(this Vector2 v)
public static Vector2? Scale(this Vector2? v)
```

Returns
`f * ImGuiHelpers.GlobalScale * (Svc.PluginInterface.UiBuilder.DefaultFontSpec.SizePt / 12f)`.
This is not the same as Dalamud's `ImGuiHelpers.GlobalScale` alone: it also follows the user's
default font size. Requires `ECommonsMain.Init` to have run (`Svc.PluginInterface`). Usage:
`new Vector2(24f.Scale())`, `150f.Scale()`.

### Width helpers

| Signature | Behaviour |
|---|---|
| `void SetNextItemFullWidth(int mod = 0)` | `SetNextItemWidth(GetContentRegionAvail().X + mod)`; pass a negative `mod` to leave room |
| `void SetNextItemWidth(float percent, int mod = 0)` | `GetContentRegionAvail().X * percent + mod`; `percent` is a fraction (`0.5f` = half), despite the doc comment saying "%" |
| `void SetNextItemWidthScaled(float width)` | `ImGui.SetNextItemWidth(width.Scale())` |
| `float GetWindowContentRegionWidth()` | `GetWindowContentRegionMax().X - GetWindowContentRegionMin().X` (replacement for the removed ImGui function) |
| `Vector2 CalcIconSize(string icon, bool isButton = false)` / `CalcIconSize(FontAwesomeIcon icon, bool isButton = false)` | text size in the icon font, plus `FramePadding * 2` when `isButton` |
| `float Measure(Action func, bool includeSpacing = true)` | runs `func`, returns the X distance the cursor travelled, plus `ItemSpacing.X` when `includeSpacing`; it really draws `func` and ends the line with `ImGui.NewLine()` |

`ImGuiEx.SetNextItemWidth(200f)` compiles and silently means "200 x available width". Use
`ImGui.SetNextItemWidth` or `SetNextItemWidthScaled` for pixel widths.

### Cursor helpers

| Signature | Behaviour |
|---|---|
| `void PushCursor(Vector2 vec)` / `PushCursor(float x, float y)` | `SetCursorPos(GetCursorPos() + vec)` |
| `void PushCursorX(float x)` / `void PushCursorY(float y)` | single axis |
| `void Spacing()` | calls `Spacing((Vector2?)null)`: moves the cursor by 10 px scaled on both X and Y |
| `void Spacing(Vector2? size = null, bool accountForScale = true)` | moves cursor X and Y by `size` (default `new Vector2(10f)`), scaled when `accountForScale` |
| `void Spacing(float pix = 10f, bool accountForScale = true)` | `[Obsolete("Use ImGuiEx.Spacing(Vector2, bool) instead")]`; X only; the flag is inverted in source (`accountForScale ? pix : pix.Scale()`) |

"Push" here means "shift"; nothing is pushed on a stack and there is no matching pop.
`ImGuiEx.Spacing` is unrelated to `ImGui.Spacing()`: it moves the cursor instead of adding a
vertical gap item.

### Line alignment (`Alignment.cs`)

```csharp
public static void LineCentered(Action func)
public static void LineCentered(string id, Action func)

public static void RightFloat(Action drawAction)
public static void RightFloat(Action drawAction, out float width)
public static void RightFloat(string id, Action drawAction)
public static void RightFloat(string id, Action drawAction, out float width, float? contentRegionAvailX = null)

public static void InputWithRightButtonsArea(Action inputAction, Action rightAction)
public static void InputWithRightButtonsArea(string id, Action inputAction, Action rightAction)
```

All three measure the drawn content and use the width on the next frame, so the first frame is
drawn unaligned, and a width change lags one frame.

- `LineCentered`: centres one line of items (use `ImGui.SameLine()` between them inside `func`).
  Ends with `ImGui.Dummy(Vector2.Zero)`, so the line is terminated for you.
- `RightFloat`: draws `drawAction` flush right in the available region, then restores the cursor
  to where it was before the call (from the second frame on). It therefore overlays the current
  line; draw it before the left-hand content of the same line or follow it with your own layout.
  `width` returns the measured width. `contentRegionAvailX` overrides the region width.
- `InputWithRightButtonsArea`: calls `ImGui.SetNextItemWidth(avail - cachedRightWidth)` and then
  `inputAction`, `SameLine`, `rightAction`. `inputAction` must draw exactly one item that honours
  `SetNextItemWidth` (input, combo, slider). `rightAction` is one line of items.

```csharp
ImGuiEx.LineCentered("AboutButtons", () =>
{
    if(ImGui.Button("Discord")) GenericHelpers.ShellStart(url);
    ImGui.SameLine();
    if(ImGui.Button("GitHub")) GenericHelpers.ShellStart(repo);
});

ImGuiEx.InputWithRightButtonsArea("PlanSelector", () =>
{
    ImGui.InputTextWithHint("##name", "Plan name", ref name, 100);
}, () =>
{
    if(ImGuiEx.IconButton(FontAwesomeIcon.Save)) Save();       // topics/imguiex-widgets.md
    ImGui.SameLine();
    if(ImGuiEx.IconButton(FontAwesomeIcon.Trash)) Delete();
});
```

`GenericHelpers.ShellStart(string s)` exists in `GenericHelpers.cs`; its behaviour is documented
in `references/topic-generic-helpers.md`.

## Tables

### BeginDefaultTable

```csharp
public static bool BeginDefaultTable(string[] headers, bool drawHeader = true,
    ImGuiTableFlags extraFlags = ImGuiTableFlags.None)
public static bool BeginDefaultTable(string id, string[] headers, bool drawHeader = true,
    ImGuiTableFlags extraFlags = ImGuiTableFlags.None, bool flagsOverride = false,
    ImGuiTableFlags? nullifyFlags = null)
public static void DefaultTableColumns(IEnumerable<string> headers, bool drawHeader = true)
```

- Returns the result of `ImGui.BeginTable`. When `true` you must call `ImGui.EndTable()` yourself;
  there is no scope object.
- Column count is `headers.Length`. Flags are `DefaultTableFlags | extraFlags`; `flagsOverride`
  uses `extraFlags` alone; `nullifyFlags` are removed at the end. `SortTristate` in `extraFlags`
  adds `Sortable` automatically.
- Header prefixes, parsed by `DefaultTableColumns` and stripped from the label: `~` =
  `WidthStretch`, `^` = `NoSort`. They can be combined in any order (`"~^Name"`).
- Use `##id` headers for columns without a visible title (`"##controls"`).
- The id-less overload uses the fixed id `"##ECommonsDefaultTable"`. Two such tables in the same
  ID scope share an ImGui id; with the default `NoSavedSettings` this is usually harmless for
  plain tables, but give sortable or differently-shaped tables an explicit id.

```csharp
ImGuiEx.SetNextItemFullWidth();
ImGuiEx.FilteringInputTextWithHint("##search", "Search...", out var filter);
if(ImGuiEx.BeginDefaultTable("Characters", ["~Character", "Plan", "##controls"]))
{
    foreach(var x in C.Characters)
    {
        if(filter != "" && !x.Name.Contains(filter, StringComparison.OrdinalIgnoreCase)) continue;
        ImGui.PushID(x.ID);
        ImGui.TableNextRow();
        ImGui.TableNextColumn();
        ImGuiEx.TextV(x.Name);
        ImGui.TableNextColumn();
        ImGuiEx.Text(x.PlanName);
        ImGui.TableNextColumn();
        if(ImGuiEx.IconButton(FontAwesomeIcon.Trash, enabled: ImGuiEx.Ctrl)) toDelete = x;
        ImGui.PopID();
    }
    ImGui.EndTable();
}
```

(Shape taken from AutoRetainer `CleanupCharacterConfiguration.cs`. `IconButton` and its `enabled`
parameter are documented in `references/topic-imguiex-widgets.md`.)

### Table helpers

| Signature | Behaviour |
|---|---|
| `void SimpleTableTextColumns(params string[] texts)` | for each string: `TableNextColumn()` + `ImGuiEx.Text(x ?? "")` |
| `void SimpleTableTextColumns(Action<string> action, params string[] texts)` | same, but calls `action` for each cell |
| `bool TryGetTableSortDirection(out bool isAscending)` | see below |
| `bool TryGetTableSortDirection(out bool isAscending, out int column)` | |
| `bool TryGetTableSortDirection(out bool isAscending, out int column, out bool isSortDirectionDirty)` | |

`TryGetTableSortDirection` reads `ImGui.TableGetSortSpecs()` inside a table. Returns `false` when
there are no sort specs or the direction is `None`; otherwise `true` with the first spec's column
index and direction. `isSortDirectionDirty` mirrors `SpecsDirty`; the method does not reset it.
Only the first sort spec is reported (no multi-sort).

### EzTable (`EzTable.cs`)

```csharp
public static void EzTable(IEnumerable<EzTableEntry> entries)
public static void EzTable(string? id, IEnumerable<EzTableEntry> entries)
public static void EzTable(ImGuiTableFlags? tableFlags, IEnumerable<EzTableEntry> entries)
public static void EzTable(string? ID, ImGuiTableFlags? tableFlags, IEnumerable<EzTableEntry> entries, bool header)

public struct EzTableEntry   // nested: ImGuiEx.EzTableEntry
{
    public string ColumnName;
    public Action Delegate;
    public ImGuiTableColumnFlags ColumnFlags;
    public EzTableEntry(string columnName, Action @delegate);
    public EzTableEntry(string columnName, bool stretch, Action @delegate);   // WidthStretch : WidthFixed
    public EzTableEntry(string columnName, ImGuiTableColumnFlags? columnFlags, Action @delegate);
}
```

A complete table (begin, columns, header, rows, end) from a flat list of cells in row-major order.

- Column count is inferred: `entries.Length / (number of entries whose ColumnName equals the first
  entry's ColumnName)`. Every row must therefore supply the same columns in the same order, and
  the first column's name must not be reused for another column.
- Column setup (name and flags) is taken from the first row's entries only.
- Default flags: `Borders | RowBg | SizingFixedFit` (no `NoSavedSettings`, unlike
  `DefaultTableFlags`). `header` defaults to `true` in the short overloads.
- An empty sequence draws the red text "Table contains no elements!" and returns.
- `ID == null` falls back to `GenericHelpers.GetCallStackID()` each frame; pass an id.
- Each cell delegate is wrapped in try/catch and logged with `e.Log()`.
- `EzTableEntry` constructors throw `ArgumentNullException` for a null name or delegate.

```csharp
List<ImGuiEx.EzTableEntry> entries = [];
foreach(var x in items)
{
    entries.Add(new("Item", true, () => ImGuiEx.TextCopy(x.Name)));
    entries.Add(new("ID", () => ImGuiEx.TextCopy($"{x.ID}")));
    entries.Add(new("Count", () => ImGuiEx.Text($"{x.Count}")));
}
ImGuiEx.EzTable("ItemTable", entries);
```

### EzTableColumns

```csharp
public static void EzTableColumns(string id, Action[] values, int? columns = null,
    ImGuiTableFlags extraFlags = ImGuiTableFlags.None)
```

Draws each action in its own cell of a `SizingStretchSame | NoSavedSettings` table: an equal-width
column layout without borders or resizing. `columns` forces the column count (actions then flow
into further rows); default is `values.Length`. With exactly one action no table is created and
the action is simply invoked. Actions run through `GenericHelpers.Safe`. AutoRetainer uses it as
a responsive grid:

```csharp
ImGuiEx.EzTableColumns("cols", draw, Math.Max(1, (int)ImGui.GetContentRegionAvail().X / 150));
```

### Pagination

```csharp
public static Action[] Pagination(Action[] actions, int perPage = 0, int maxPages = 0)
public static Action[] Pagination(string id, Action[] actions, int perPage = 0, int maxPages = 0)
public static Action[] Pagination(Action[] actions, out Action? paginator, int perPage = 0, int maxPages = 0)
public static Action[] Pagination(string id, Action[] actions, out Action? paginator, int perPage = 0, int maxPages = 0)
```

Splits an array of draw actions into pages and returns the slice for the selected page; you invoke
the returned actions yourself. The `out paginator` overloads hand you the page-button row to draw
where you want (it is `null` when everything fits on one page or `actions` is empty); the other
overloads draw it immediately, before returning.

- `perPage <= 0` requires `maxPages > 0` (otherwise `ArgumentOutOfRangeException`); `perPage`
  becomes `ceil(actions.Length / maxPages)`.
- With both set, `maxPages` caps the page count and `perPage` grows when needed.
- The page row is `maxPages` equal-width buttons spanning the available width.
- Selected page is kept in the private static `SelectedPages[id]`; `id` must be unique per plugin.
- As written in source the slice is `actions[rangeMin..rangeMax]` with
  `rangeMax = (page + 1) * perPage - 1` and a C# exclusive range end, so a full page returns
  `perPage - 1` actions and the element at index `(page + 1) * perPage - 1` is not returned on any
  page. Verify the item count if you depend on this helper.

```csharp
var page = ImGuiEx.Pagination("MyList", [.. actions], out var paginator, 100, 10);
paginator?.Invoke();
foreach(var a in page) a();
```

## Tabs

```csharp
public static void EzTabBar(string id, params (string name, Action function, Vector4? color, bool child)[] tabs)
public static void EzTabBar(string id, string KoFiTransparent, params (...)[] tabs)
public static void EzTabBar(string id, string KoFiTransparent, string openTabName, params (...)[] tabs)
public static void EzTabBar(string id, string KoFiTransparent, string openTabName, ImGuiTabBarFlags flags, params (...)[] tabs)
```

Draws a whole tab bar. Per tab tuple:

- `name`: tab label and ImGui id. A tuple with `name == null` is skipped, which is the supported
  way to hide a tab conditionally (`(cond ? "Debug" : null, DrawDebug, null, true)`).
- `function`: draw callback; exceptions are caught and logged with `e.Log()`.
- `color`: optional label text colour (applied to the tab label only).
- `child`: when `true` the content is wrapped in `ImGui.BeginChild(name + "child")` /
  `EndChild()`, giving the tab its own scroll region.
- The content runs inside `ImGui.PushID(name + "tab")`.

`openTabName`: while it equals a tab's name, that tab gets `ImGuiTabItemFlags.SetSelected` every
frame. Pass it for one frame and reset it to `null` afterwards, or the user cannot switch tabs.

`KoFiTransparent`: only tested for `null`. When non-null, `PatreonBanner.RightTransparentTab()` is
called with no argument, which adds an invisible, disabled trailing tab as wide as
`PatreonBanner.Text` to reserve room for the donation banner drawn by `PatreonBanner.DrawRight()`.
The string value itself is not used. `ECommons/Funding/*` is not in `references/`; read it in a
checkout.

```csharp
PatreonBanner.DrawRight();
ImGuiEx.EzTabBar("tabbar", PatreonBanner.Text,
    ("Retainers", MultiModeUI.Draw, null, true),
    ("Statistics", DrawStats, null, true),
    (C.Debug ? "Debug" : null, DrawDebug, ImGuiColors.DalamudGrey3, true));

// without the banner:
ImGuiEx.EzTabBar("tabs", ("General", DrawGeneral, null, true), ("About", DrawAbout, null, false));
```

`BeginTabItem(string label, ImGuiTabItemFlags flags)` is `[Obsolete(..., true)]` (compile error)
and throws; the message says to use `ImGui.BeginTabItem`.

## Trees, headers, selectables

| Signature | Behaviour |
|---|---|
| `bool TreeNode(string name, ImGuiTreeNodeFlags flags = ImGuiTreeNodeFlags.SpanFullWidth)` | `ImGui.TreeNodeEx` with `SpanFullWidth` always OR-ed in |
| `bool TreeNode(Vector4? color, string name, ImGuiTreeNodeFlags flags = ImGuiTreeNodeFlags.SpanFullWidth)` | plus label colour |
| `bool CollapsingHeader(string text)` / `CollapsingHeader(Vector4? col, string text)` | `ImGui.CollapsingHeader` with optional text colour |
| `bool CollapsingHeader(string text, Vector4? col)` | `[Obsolete("Move color parameter to be the first one")]` |
| `void TreeNodeCollapsingHeader(string name, Action action, ImGuiTreeNodeFlags extraFlags = ImGuiTreeNodeFlags.None)` | `usePadding: true` |
| `void TreeNodeCollapsingHeader(string name, bool usePadding, Action action, ImGuiTreeNodeFlags extraFlags = ImGuiTreeNodeFlags.None, Action? postHeaderAction = null)` | bordered, narrow collapsing section |
| `bool Selectable(string id, bool enabled = true, bool selected = false, ImGuiSelectableFlags flags = ImGuiSelectableFlags.None, Vector2 size = default, Vector2? alignment = null)` | selectable that can look disabled; `alignment` pushes `ImGuiStyleVar.SelectableTextAlign` for the call (parameter added in 3.2.1.23) |
| `bool Selectable(Vector4? color, string id, bool enabled = true, bool selected = false, ImGuiSelectableFlags flags = ImGuiSelectableFlags.None, Vector2 size = default, Vector2? alignment = null)` | plus text colour |
| `bool SelectableNode(string id, bool enabled = true)` / `SelectableNode(Vector4? color, string id, bool enabled = true)` | leaf tree node used as a clickable bullet row |
| `bool SelectableNode(string id, ref bool selected, bool enabled = true)` | toggling variant |
| `bool SelectableNode(Vector4? color, string id, ref bool selected, ImGuiTreeNodeFlags extraFlags = ImGuiTreeNodeFlags.Leaf, bool enabled = true)` | toggling variant, full signature |

Notes:

- `ImGuiEx.TreeNode` returns what `TreeNodeEx` returns. If it returns `true` and you did not pass
  `NoTreePushOnOpen`, you still owe an `ImGui.TreePop()`.
- `TreeNodeCollapsingHeader` is callback-based and needs no pop. It builds a one-column bordered
  table, draws a `Selected | NoTreePushOnOpen` tree node as the header, invokes `postHeaderAction`
  right after the header (same cell, for `SameLine` extras), and when open runs `action` in an
  inner table (with `PadOuterX` when `usePadding`). Exceptions in `action` are logged. Use
  `###id` in `name` when the visible text changes:
  `ImGuiEx.TreeNodeCollapsingHeader($"Items ({n} selected)###items", () => { ... });`
- `Selectable` with `enabled: false` draws at 60% alpha and always returns `false`; it does not
  use `BeginDisabled`, so hover highlighting still occurs.
- `SelectableNode` without `ref selected` returns the `TreeNodeEx` result combined with `enabled`
  (leaf + `NoTreePushOnOpen`, no pop needed). The `ref selected` variants return `true` when the
  node is left-clicked while enabled, and flip `selected`.

## Popups and windows

| Signature | Behaviour |
|---|---|
| `bool BeginPopupNextToElement(string popupId)` | `SameLine(0,0)`, takes the cursor screen position, `Dummy(Vector2.Zero)`, `SetNextWindowPos(pos, ImGuiCond.Appearing)`, then `ImGui.BeginPopup(popupId)` |
| `OverlayFlags` | window flags for click-through overlays (see "Static state") |

`BeginPopupNextToElement` opens the popup at the right edge of the previous item instead of at the
mouse position. You still call `ImGui.OpenPopup(popupId)` to open it and `ImGui.EndPopup()` when
it returns `true`. Window classes (`EzOverlayWindow` and friends) are in `references/topic-imgui-helpers.md`.

## Field-less filter inputs

```csharp
public static bool FilteringInputTextWithHint(string label, string hint, out string result,
    int maxLength = 200, ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
public static ref string GetFilteringInputTextString(string label)
public static bool FilteringInputInt(string label, out int result, int step = 1, int step_fast = 100,
    ImGuiInputTextFlags flags = ImGuiInputTextFlags.None)
[OverloadResolutionPriority(1)]
public static bool FilteringCheckbox(string label, out bool result)
public static bool FilteringCheckbox(string label, out bool? result)
```

Same widgets as `ImGui.InputTextWithHint` / `InputInt` / `Checkbox`, but the value is stored inside
ECommons, so no field is needed for throwaway search boxes. Return value: `true` on the frame the
value changed. `result` always receives the current stored value.

- Storage is `Ref<T>.Get($"{ImGui.GetID(label)}_filter")` (the `bool?` checkbox uses the suffix
  `_filter1`). The key depends on the ImGui ID stack at the call site, so the same label under
  different `PushID` scopes gives separate values, and the same label in the same scope shares
  one.
- `GetFilteringInputTextString(label)` returns a `ref` to the stored string. Call it under the
  same ID stack to read or reset the filter: `ImGuiEx.GetFilteringInputTextString("##search") = "";`
- The initial string value is `""`, not `null` (`Ref<string>` special-cases it).
- `FilteringCheckbox(label, out var x)` binds to the `bool` overload because of
  `OverloadResolutionPriority`. The `bool?` overload draws the tri-state `ImGuiEx.Checkbox`
  (`references/topic-imguiex-widgets.md`).

## Drag-to-copy (`DragDropRepopulate`)

These live in `ImGuiEx.cs`, not in `DragDrop.cs`. Call one directly after a widget drawn in a loop:
the user can then drag that widget onto the same widget in another row to copy its value.

```csharp
public static void DragDropRepopulate<T>(string dragDropIdentifier, T data, ref T field) where T : struct
public static void DragDropRepopulate<T>(string dragDropIdentifier, T? data, ref T? field) where T : struct
public static void DragDropRepopulate<T>(string dragDropIdentifier, T data, Action<T> callback) where T : struct
public static void DragDropRepopulate<T>(string dragDropIdentifier, T data, ICollection<T> dataCollection) where T : struct
public static void DragDropRepopulateClass<T>(string dragDropIdentifier, T data, Action<T> callback) where T : class
public static void DragDropRepopulateClass<T>(string dragDropIdentifier, T data, ICollection<T> dataCollection) where T : class
```

- `dragDropIdentifier`: payload type string, unique per plugin and short. `data`: this row's
  current value. The target receives the dragged row's value in `field` / `callback`.
- Every overload adds the hover tooltip "Drag this selector to other selectors to set their
  values to the same" to the previous item.
- The payload is accepted with `AcceptBeforeDelivery`, so targets change while the drag hovers
  over them, before the mouse is released. This is what allows sweeping across many rows.
- The `ICollection<T>` overloads are meant for `CollectionCheckbox`: when the dragged value is in
  the collection the current `data` is added, otherwise it is removed.
- `DragDropRepopulateClass` cannot put a reference in the payload; it keeps a
  `ConditionalWeakTable<T, Box<Guid>>` in `Ref<>` under the key
  `"__ECommons.DragDropRepopulateClass.ConditionalWeakTable"` and transfers the Guid.
- The nullable-struct overload wraps the payload in `ECommons.MathHelpers.ENullable<T>`.
- Exceptions are caught and logged.

```csharp
foreach(var chr in C.Characters)
{
    ImGui.PushID(chr.ID);
    ImGui.Checkbox("##fuel", ref chr.AutoFuelPurchase);
    ImGuiEx.DragDropRepopulate("EnFuel", chr.AutoFuelPurchase, ref chr.AutoFuelPurchase);
    ImGui.PopID();
}
```

`RealtimeDragDrop<T>` (row reordering) is a separate class, see `references/topic-imguiex-widgets.md`.

## Keys and mouse

| Signature | Behaviour |
|---|---|
| `bool Ctrl`, `bool Alt`, `bool Shift` (static properties) | `ImGui.GetIO().KeyCtrl` / `KeyAlt` / `KeyShift` |
| `bool IsKeyPressed(int key, bool repeat)` | `ImGui.IsKeyPressed((ImGuiKey)key, repeat)`; the int is an `ImGuiKey` value, not a Windows virtual key |
| `void ActivateIfDoubleClicked()` | put after a slider/drag: double-click gives it keyboard focus (`SetKeyboardFocusHere(-1)`) for typed input |

The modifier properties read ImGui IO, so they reflect the state ImGui sees during the frame.
Typical use is a confirmation guard: `ImGuiEx.IconButton(FontAwesomeIcon.Trash, enabled:
ImGuiEx.Ctrl)`. For Windows virtual keys use `GenericHelpers.IsKeyPressed` (other topic file).

## Style, ID and colour helpers

| Signature | Behaviour |
|---|---|
| `void PushStyleVar(IEnumerable<ImGuiStyleVar> styleVars, Vector2 value)` / `(..., float value)` | one `ImGui.PushStyleVar` per element |
| `void PushStyleColor(IEnumerable<ImGuiCol> col, Vector4 value)` / `(..., uint value)` | one `ImGui.PushStyleColor` per element |
| `ImGui.PushID<T>(T g)` (C# 14 extension block on `ImGui`, `Extensions.cs`) | `ImGui.PushID(g.ToString())` |
| `Vector4 Vector4FromRGB(this uint col, float alpha = 1.0f)` | `0xRRGGBB` -> `Vector4` |
| `Vector4 Vector4FromRGBA(this uint col)` | `0xRRGGBBAA` -> `Vector4` |
| `Vector4 MutateColor(ImGuiCol col, byte r, byte g, byte b)` | current style colour `col` with RGB replaced and its alpha kept |
| `void WithTextColor(Vector4 col, Action func)` | `[Obsolete("Just push style col")]` |
| `Vector4 GetParsedColor(int percent)` | `[Obsolete("Probably isn't used by anyone")]`; FFLogs-style parse colours |

- The multi-push helpers have no matching pop: call `ImGui.PopStyleVar(n)` /
  `ImGui.PopStyleColor(n)` with the number of elements you passed.
- `PushID<T>` is declared as `extension(ImGui) { public static void PushID<T>(T g) }`, so with
  `using ECommons.ImGuiMethods;` it is written `ImGui.PushID(guid)`. Being an extension, it is
  only considered when no real `ImGui.PushID` overload is applicable to the argument. It calls
  `ToString()` every frame; a null reference argument throws. Pop with `ImGui.PopID()`.

## Other widgets defined in `ImGuiEx.cs`

| Signature | Behaviour |
|---|---|
| `bool DragDouble(ImU8String label, scoped ref double v, float vSpeed = 1.0f, float vMin = 0.0f, float vMax = 0.0f, ImU8String format = default, ImGuiSliderFlags flags = ImGuiSliderFlags.None)` | `ImGui.DragFloat` on a `float` copy; value is written back only when changed, with float precision |
| `bool ArrowButton(string label, ImGuiDir direction)` | calls `ImGuiNative.ArrowButton`; null/empty label becomes `"ECommonsDefaultID"` |
| `bool JobSelector(string id, ICollection<Job> selectedJobs, JobSelectorOption[] options = null, int maxPreviewJobs = 3, string noJobSelectedPreview = "None", Func<Job, bool> jobDisplayFilter = null)` | multi-select job combo with icons and a filter box |
| `void PluginAvailabilityIndicator(IEnumerable<RequiredPluginInfo> pluginInfos, string? prependText = null, bool all = true)` | check/cross icon with dependency tooltip |
| `bool EnumOrderer<T>(string id, List<T> order) where T : IConvertible` | `[Obsolete("Use RealtimeDragDrop. Better user experience anyway.")]` |

### JobSelector

`Job` is `ECommons.ExcelServices.Job`. Returns `true` whenever `selectedJobs` was modified.
`public enum JobSelectorOption { None, IncludeBase, ClearFilterOnOpen, BulkSelectors }`:

- `IncludeBase`: also list base classes (otherwise entries with `IsUpgradeable()` are hidden).
  `Job.ADV` is always skipped.
- `BulkSelectors`: adds group checkboxes (DoW/DoM, DoL, Tanks, Healers, DPS, Melee DPS, Ranged
  DPS, Magical/Physical ranged DPS, Ranged jobs, Melee jobs).
- `ClearFilterOnOpen`: in this commit the option does not clear anything; the only statement it
  guards is `ImGui.SetNextItemWidth(150f)` on the frame the combo appears.
- The preview shows the job names up to `maxPreviewJobs`, then "`N selected`".
- Icons come from `ThreadLoadImageHandler.TryGetIconTextureWrap`, drawn at `24f.Scale()`.
- The filter text is one private static string shared by every `JobSelector` in the plugin.

### PluginAvailabilityIndicator

```csharp
public readonly record struct RequiredPluginInfo   // nested: ImGuiEx.RequiredPluginInfo
{
    public readonly string InternalName;
    public readonly string? VanityName;
    public readonly Version? MinVersion;
    public RequiredPluginInfo(string internalName);
    public RequiredPluginInfo(string internalName, string vanityName);
    public RequiredPluginInfo(string internalName, Version minVersion);
    public RequiredPluginInfo(string internalName, string vanityName, Version minVersion);
}
```

Calls `ImGui.SameLine()` first, then draws a green check or red cross. A plugin passes when
`Svc.PluginInterface.InstalledPlugins` has an entry with `IsLoaded`, matching `InternalName`, and
`Version >= MinVersion` (when given). `all: true` requires every listed plugin, `false` any one.
The hover tooltip lists each plugin as installed, outdated or not installed.

```csharp
ImGui.Checkbox("Use Lifestream for travel", ref C.UseLifestream);
ImGuiEx.PluginAvailabilityIndicator([new("Lifestream", new Version("2.5.3.0"))]);
```

## Obsolete members (complete list for these files)

| Member | Attribute message / replacement |
|---|---|
| `bool CollapsingHeader(string text, Vector4? col)` | "Move color parameter to be the first one" -> `CollapsingHeader(Vector4? col, string text)` |
| `void Spacing(float pix = 10f, bool accountForScale = true)` | "Use ImGuiEx.Spacing(Vector2, bool) instead" |
| `void WithTextColor(Vector4 col, Action func)` | "Just push style col" |
| `Vector4 GetParsedColor(int percent)` | "Probably isn't used by anyone" |
| `bool EnumOrderer<T>(string id, List<T> order)` | "Use RealtimeDragDrop. Better user experience anyway." |
| `bool BeginTabItem(string label, ImGuiTabItemFlags flags)` | error-level; "Switch to using ImGui.BeginTabItem. It now has version without \"close\"." |
| `bool HashSetCheckbox<T>(string label, T value, HashSet<T> collection)` | "Please switch to CollectionCheckbox" |
| `void ImGuiLineCentered(string id, Action func)` | "Please switch to LineCentered" (still used 4 times in the pinned consumers) |

`Allocate`, `Free` and `GetUtf8` are `internal` and error-level obsolete; they are not callable
from a plugin.

## Gotchas

- `ImGuiEx.SetNextItemWidth(float percent, ...)` takes a fraction of the available width, not
  pixels. `ImGuiEx.SetNextItemWidthScaled` and `ImGui.SetNextItemWidth` take pixels.
- `.Scale()` multiplies by global scale and by default font size / 12; it is not interchangeable
  with `* ImGuiHelpers.GlobalScale`.
- `uint` colour overloads take ImGui packed colours. For `0xRRGGBB` literals convert with
  `0xFF8800u.Vector4FromRGB()` or use `EColor`/`EzColor`.
- `HelpMarker`, `InfoMarker` and `PluginAvailabilityIndicator` call `ImGui.SameLine()` themselves.
  `Icon`/`IconWithText`/`IconWithTooltip` call `SameLine()` after the icon.
- Id-less overloads of `LineCentered`, `RightFloat`, `InputWithRightButtonsArea`, `Pagination`
  and `EzTable` compute a stack-trace key per call: slow, and identical for all calls from the
  same method. Always pass an id in loops and hot paths.
- The alignment helpers are one frame late by design. Content whose width changes every frame
  will jitter.
- `BeginDefaultTable` needs a manual `ImGui.EndTable()`; `EzTable`, `EzTableColumns`,
  `EzTabBar` and `TreeNodeCollapsingHeader` close themselves.
- `EzTabBar`'s `openTabName` forces selection for as long as it is non-null.
- `Pagination` drops one element per full page in this commit (see Tables > Pagination).
- `JobSelectorOption.ClearFilterOnOpen` has no clearing effect in this commit, and the filter is
  shared between all job selectors.
- The `Filtering*` values are keyed by ImGui id; changing the surrounding `PushID` scope loses
  the value. They live until `ECommonsMain.Dispose()`.
- `ImGuiEx.Trim(this string, int)` and `ImGuiTrim` become visible on every `string` once
  `using ECommons.ImGuiMethods;` is present.
- Callbacks given to `EzTabBar`, `EzTable`, `EzTableColumns` and `TreeNodeCollapsingHeader` are
  exception-guarded; callbacks given to `LineCentered`, `RightFloat`,
  `InputWithRightButtonsArea`, `Measure` and the actions returned by `Pagination` are not. An
  exception thrown between a push and a pop inside a guarded callback still leaves the ImGui
  stack unbalanced.

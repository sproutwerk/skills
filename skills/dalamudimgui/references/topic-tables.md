# Tables and large lists

Covers table sort specs, `ImGuiClip` / `ImGuiListClipper` clipping, `ImGuiTable` and the
`Table<T>` / `Column<T>` classes.
Verified against goatcorp/Dalamud `master` @ `43b3411` (2026-11-10, API level 15, Dear ImGui 1.88 docking).
Paths are relative to the skill root (`SKILL.md`, `references/topic-*.md`, `references/...`); a bare
`index/...` or `src/...` path means `references/index/...` or `references/src/...`.

Tables are stock ImGui (`ImRaii.Table`, `TableSetupColumn`,
`TableSetupScrollFreeze`, `TableHeadersRow`, `TableNextRow`,
`TableNextColumn`, `TableSetColumnIndex`, `TableGetSortSpecs`). Sort specs:

```csharp
var specs = ImGui.TableGetSortSpecs();
if (specs.SpecsDirty)
{
    var col = specs.Specs.ColumnIndex;
    var asc = specs.Specs.SortDirection == ImGuiSortDirection.Ascending;
    this.Sort(col, asc);
    specs.SpecsDirty = false;
}
```

Clipping, for anything over a few hundred rows:

```csharp
using var table = ImRaii.Table("rows"u8, 2, ImGuiTableFlags.ScrollY | ImGuiTableFlags.RowBg, new Vector2(-1, -1));
if (!table) return;

ImGuiClip.ClippedDraw(this.rows, row =>
{
    ImGui.TableNextRow();
    ImGui.TableNextColumn(); ImGui.Text(row.Name);
    ImGui.TableNextColumn(); ImGui.Text($"{row.Count}");
}, ImGui.GetTextLineHeightWithSpacing());
```

| `ImGuiClip` | |
|---|---|
| `ClippedDraw<T>(IReadOnlyList<T> data, Action<T> draw, float lineHeight)` | Random-access data, uniform row height, uses `ImGuiListClipper` |
| `ClippedDraw<T>(IReadOnlyList<T>, Action<T, int> draw, float lineHeight)` | Same, with index |
| `ClippedDraw<T>(IReadOnlyList<T>, Action<T> draw, int itemsPerLine, float lineHeight)` | Grid (icons) |
| `GetNecessarySkips(height)` + `ClippedDraw<T>(IEnumerable<T>, int skips, Action<T> draw, int? count = null, int startIndex = 0)` + `DrawEndDummy(remainder, height)` | Non-random-access data |
| `FilteredClippedDraw<T>(IEnumerable<T>, skips, checkFilter, draw, startIndex = 0)` | Filtered while drawing |

Row height must be constant and include item spacing. Pass a cached delegate
or method group; a lambda capturing locals allocates every frame. By hand:
`var c = ImGui.ImGuiListClipper(); c.Begin(count, rowHeight); while (c.Step())`
loop `c.DisplayStart..c.DisplayEnd`, then `c.End(); c.Destroy();` (natively
allocated, so always `Destroy`).

`ImGuiTable.DrawTable<T>(label, data, drawRow, flags, params string[] columnTitles)`
and `DrawTabbedTable` (inside a collapsing header) cover simple unclipped
tables. `Dalamud.Interface.Utility.Table.Table<T>` with `Column<T>`,
`ColumnString<T>`, `ColumnSelect<T, TItem>`, `ColumnFlags<T, TItem>` is a
sortable, filterable, clipped table: subclass the columns (`ToName`,
`DrawColumn`, `Compare`, `FilterFunc`, `Width`), then `table.Draw(itemHeight)`.

OtterGui topic: the generic `OtterGui.Table.Table<T>` with its `Column<T>` types (`ColumnString`,
`ColumnNumber`, `ColumnSelect`, `ColumnFlags`, `TriStateColumnFlags`, `YesNoColumn`), how
filtering and sorting work, and a complete example. Verified against `Limiana/OtterGui` `a08fda6`
(2026-05-30, branch `main`; `Table/` is identical to upstream `79771ee`). The Artisan pin
`32c4454` changes three members, flagged below.

Paths such as `references/...` and `references/topic-*.md` are relative to the skill root (the directory that
holds `SKILL.md`), not to this file. The example was written against the source and **not
compiled**.

## Model

`Table<T>` owns a read-only collection of items and an array of columns. Each frame `Draw`:

1. if `FilterDirty`, rebuilds `FilteredItems` with every item for which **all** columns'
   `FilterFunc` return true, then marks sort dirty;
2. begins an ImGui table named `"Table"` inside `ImRaii.PushId(Label)`, sized to the available
   content region minus `ExtraHeight * GlobalScale`;
3. calls `PreDraw()`, freezes one column and one row, sets up one ImGui column per `Column<T>`
   from `Label`, `Flags`, `Width`;
4. draws the header row: an empty `TableHeader` (the sort arrow) followed on the same line by the
   column's `DrawFilter()`; a `true` return sets `FilterDirty`;
5. sorts `FilteredItems` with `StableSort` (`Extensions/StableInsertionSortExtension.cs`; ties keep
   their previous order) using the clicked column's `Compare`
   (ascending) or `CompareInv` (descending), wrapped in `PreSort()` / `PostSort()`;
6. draws rows through `ImGuiClip.ClippedDraw(FilteredItems, DrawItem, ItemHeight)`, calling
   `DrawColumn(item, index)` for each visible cell. `index` is the item's position in the
   original collection.

Rows must all have the height you pass to `Draw`.

## `Table<T>` members (`references/src/Table/Table.cs`)

```csharp
public static class Table { public const float ArrowWidth = 10; }

public class Table<T>
{
    public Table(string label, IReadOnlyCollection<T> items, params Column<T>[] headers);
    public void Draw(float itemHeight);
    public bool WouldBeVisible(T value);           // all columns accept the value

    public ImGuiTableFlags Flags;                  // default below
    public float ExtraHeight { get; set; } = 0;    // space to leave under the table, unscaled
    public int TotalItems { get; }                 // Items.Count
    public int CurrentItems { get; }               // FilteredItems.Count
    public int TotalColumns { get; }
    public int VisibleColumns { get; }

    protected bool FilterDirty = true;
    protected bool SortDirty   = true;
    protected readonly IReadOnlyCollection<T> Items;
    internal  readonly List<(T, int)> FilteredItems;   // Artisan pin: public, not readonly
    protected readonly string Label;
    protected readonly Column<T>[] Headers;            // Artisan pin: public, not readonly
    protected float ItemHeight { get; set; }
    protected bool  Sortable { get; set; }             // toggles ImGuiTableFlags.Sortable in Flags
    protected int   SortIdx = -1;

    protected virtual void PreDraw() { }
    protected virtual void DrawFilters();              // throws NotImplementedException; never called
}
```

Default `Flags`: `RowBg | Sortable | BordersOuter | ScrollY | ScrollX | PreciseWidths |
SizingFixedFit | BordersInnerV | NoBordersInBodyUntilResize`.

The table does not watch the collection. After adding, removing or changing items from outside,
set `FilterDirty = true` (it is `protected`, so expose a method in your subclass).

## Column types

`Column<TItem>` (`references/src/Table/Column.cs`), all members public:

| Member | Default | Purpose |
|---|---|---|
| `string Label` | `""` | header text and ImGui column name |
| `ImGuiTableColumnFlags Flags` | `NoResize` | passed to `TableSetupColumn` |
| `virtual float Width` | `-1f` | passed as `init_width_or_weight` |
| `string FilterLabel` | `$"##{Label}Filter"` | ID for the filter widget |
| `virtual bool DrawFilter()` | draws the label as text, returns `false` | return `true` when the filter changed |
| `virtual bool FilterFunc(TItem item)` | `true` | row passes this column's filter |
| `virtual int Compare(TItem lhs, TItem rhs)` | `0` | sort order |
| `int CompareInv(TItem lhs, TItem rhs)` | `Compare(rhs, lhs)` | not virtual |
| `virtual void DrawColumn(TItem item, int idx)` | nothing | cell contents |
| `virtual void PreSort()`, `PostSort()` | nothing | cache setup/teardown around a sort |

| Type | Filter widget | Override |
|---|---|---|
| `ColumnString<TItem>` | text input with the label as hint (max 256); the text is compiled as a case-insensitive `Regex`, and if that fails falls back to a case-insensitive `Contains` | `virtual string ToName(TItem item)` (default `item.ToString()`). Fields `public string FilterValue`, `protected Regex? FilterRegex`. Constructor clears `NoResize`. Draws `ToName(item)` as text. Artisan pin only: `OnContextMenuRequest`, `InvokeContextMenu(TItem)` |
| `ColumnNumber<TItem> : ColumnString<TItem>` | same input; if the text parses as `int`, compares numerically by the `ComparisonMethod` given to the constructor, otherwise behaves like `ColumnString` | `virtual int ToValue(TItem item)` (default `item.GetHashCode()`). Constructor `ColumnNumber(ComparisonMethod method)`; `enum ComparisonMethod : byte { Equal, LessEqual, GreaterEqual }`. Draws right-aligned |
| `ColumnSelect<T, TItem>` (`T : struct, Enum, IEquatable<T>`) | combo with one selectable per enum value | constructor `ColumnSelect(T initialValue)`; `public T FilterValue`; `protected virtual IReadOnlyList<T> Values`, `string[] Names`, `void SetValue(T value)`. **`FilterFunc` and `DrawColumn` are not implemented**, you must override both |
| `ColumnFlags<T, TItem>` (`T : struct, Enum`) | combo with an "Enable All" checkbox and one checkbox per flag; right-click resets to `AllFlags`; frame tinted `0x803030A0` while not all flags are set | `public T AllFlags`; `public virtual T FilterValue` (default `default`); `protected virtual void SetValue(T value, bool enable)`; `protected virtual IReadOnlyList<T> Values`, `string[] Names`, `bool DrawCheckbox(int idx, out bool ret)`. You store the filter value yourself and implement `FilterFunc` |
| `TriStateColumnFlags<T, TItem>` (abstract) | combo of tri-state checkboxes | `protected abstract IReadOnlyList<(T On, T Off)> Values`, `protected abstract string[] Names`; `SetValue((T On, T Off) value, bool? enable)`, `SetValue(T flags, bool value)` |
| `YesNoColumn<TItem> : ColumnFlags<YesNoFlag, TItem>` | two checkboxes Yes / No | `protected virtual bool GetValue(TItem item)` (default `false`); `public string Tooltip`. Draws a centred FontAwesome check or cross. `Width` is `2 * GetFrameHeight()`. `[Flags] enum YesNoFlag { Yes = 0x01, No = 0x02 }` |

Every built-in filter widget sets its width with
`ImGui.SetNextItemWidth(-Table.ArrowWidth * ImGuiHelpers.GlobalScale)`, leaving room for the sort
arrow.

## Complete example

```csharp
using System;
using System.Collections.Generic;
using Dalamud.Bindings.ImGui;
using Dalamud.Interface.Utility;
using OtterGui.Table;

public sealed record Retainer(string Name, int Level, bool HasVenture, Job Job);

[Flags]
public enum Job { Miner = 1, Botanist = 2, Fisher = 4, Combat = 8 }

public sealed class RetainerTable : Table<Retainer>
{
    private static readonly NameColumn    Name    = new() { Label = "Name" };
    private static readonly LevelColumn   Level   = new() { Label = "Level" };
    private static readonly VentureColumn Venture = new() { Label = "Venture", Tooltip = "Has a venture assigned." };
    private static readonly JobColumn     Jobs    = new() { Label = "Job" };

    public RetainerTable(IReadOnlyCollection<Retainer> items)
        : base("RetainerTable", items, Name, Level, Venture, Jobs)
    {
        Flags |= ImGuiTableFlags.Resizable | ImGuiTableFlags.Hideable;
    }

    // The table does not notice changes of the collection.
    public void SetDirty()
        => FilterDirty = true;

    private sealed class NameColumn : ColumnString<Retainer>
    {
        public override float Width
            => 200 * ImGuiHelpers.GlobalScale;

        public override string ToName(Retainer item)
            => item.Name;
    }

    private sealed class LevelColumn : ColumnNumber<Retainer>
    {
        public LevelColumn()
            : base(ComparisonMethod.GreaterEqual)
        { }

        public override float Width
            => 60 * ImGuiHelpers.GlobalScale;

        public override int ToValue(Retainer item)
            => item.Level;
    }

    private sealed class VentureColumn : YesNoColumn<Retainer>
    {
        protected override bool GetValue(Retainer item)
            => item.HasVenture;
    }

    private sealed class JobColumn : ColumnFlags<Job, Retainer>
    {
        private Job _filter;

        public JobColumn()
        {
            AllFlags = Job.Miner | Job.Botanist | Job.Fisher | Job.Combat;
            _filter  = AllFlags;
        }

        public override float Width
            => 100 * ImGuiHelpers.GlobalScale;

        public override Job FilterValue
            => _filter;

        protected override void SetValue(Job value, bool enable)
            => _filter = enable ? _filter | value : _filter & ~value;

        public override bool FilterFunc(Retainer item)
            => _filter.HasFlag(item.Job);

        public override int Compare(Retainer lhs, Retainer rhs)
            => lhs.Job.CompareTo(rhs.Job);

        public override void DrawColumn(Retainer item, int idx)
            => ImGui.TextUnformatted(item.Job.ToString());
    }
}

// In a window's Draw():
//   _table.Draw(ImGui.GetTextLineHeightWithSpacing());
//   ImGui.TextUnformatted($"{_table.CurrentItems} / {_table.TotalItems}");
```

What the example relies on, each checked in the source:

- `Label`, `Tooltip` are public fields, so object initializers work.
- `ColumnString` and `YesNoColumn` clear `NoResize` in their constructors; `ColumnFlags` and
  `ColumnSelect` keep the `Column<T>` default `NoResize`.
- `ColumnFlags.DrawFilter` calls `SetValue(AllFlags, true)` on right-click and `SetValue(flag,
  checked)` per checkbox, then returns `true`, which makes the table refilter.
- `Enum.GetValues<Job>()` / `Enum.GetNames<Job>()` are the default `Values` / `Names`, so every
  declared enum member becomes a checkbox. Override both to hide or rename members.
- Columns are `static readonly` here only to pass them to `base(...)`; one table instance per
  column set, because columns hold filter state.

## How Artisan does it (Artisan pin only)

`Artisan/UI/Tables/RaphaelCacheTable.cs` constructs with `base("RaphaelCacheTable", cacheList)`
(no headers), then assigns `this.Headers = [.. headers];` and sets `Sortable = true`. It computes
column widths in `PreDraw()` when `ImGuiHelpers.GlobalScale` changes, and
`UI/RaphaelCacheUI.cs` reads `Table.FilteredItems` from outside and calls
`Table.Draw(ImGui.GetTextLineHeightWithSpacing() - 4f)`. Assigning `Headers` and reading
`FilteredItems` from another class need the Artisan pin (`references/topic-fork-delta.md`); on fork `main`
pass the columns to the constructor and use `CurrentItems`.

## Simple tables without the class

`OtterGui.ImGuiTable` (`references/src/ImGuiTable.cs`):

```csharp
public static void DrawTable<T>(string label, IEnumerable<T> data, Action<T> drawRow,
    ImGuiTableFlags flags = ImGuiTableFlags.None, params string[] columnTitles)
public static void DrawTabbedTable<T>(string label, IEnumerable<T> data, Action<T> drawRow,
    ImGuiTableFlags flags = ImGuiTableFlags.None, params string[] columnTitles)
```

The column count is `columnTitles.Length` (nothing is drawn when it is empty). Titles are drawn
with `TableNextColumn` + `TableHeader`, not `TableSetupColumn`. `drawRow` runs after
`TableNextRow()` and must call `ImGui.TableNextColumn()` itself for every cell.
`DrawTabbedTable` wraps the same table in `ImGui.CollapsingHeader(label)`. No clipping, no
filtering, no sorting.

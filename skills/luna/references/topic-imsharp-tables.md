# ImSharp tables and table classes

Covers `Im.Table` / `Im.TableDisposable`, `ImEx.Table.Draw`, and the class-based filterable,
sortable table (`DefaultTable<T>`, column types, table cache). Verified against Luna `52eb0fe` and
ImSharp `4d09327` (branch `gc1.88`, ImGui 1.88). Paths are relative to the skill root. Grep
`references/api-index.txt` before using a member that is not shown here.

## Tables

`Im.Table.Begin(Utf8LabelHandler id, int columns, TableFlags flags = TableFlags.None, Vector2 outerSize = default, float innerWidth = 0)`
returns `Im.TableDisposable`, whose instance methods are the table API:
`SetupColumn(label, TableColumnFlags flags = None, float initWidthOrWeight = 0, ImGuiId userId = default)`,
`SetupScrollFreeze(cols, rows)`, `HeaderRow()`, `Header(label)`, `NextRow(…)`,
`NextColumn()`, `GoToColumn(i)`, `DrawColumn(text)` (= `NextColumn()` then
`Im.Text`), `DrawFrameColumn(text)` (frame-aligned), `DrawDataPair(label, text)`,
`SetBackgroundColor(target, color, column)`, `SortSpecifications`,
`ColumnCount`, `CurrentColumn`, `CurrentRow`, `HoveredColumn`,
`GetColumnFlags`, `SetColumnEnabled`. Static equivalents on `Im.Table` exist
for code that does not hold the disposable.

```csharp
using var table = Im.Table.Begin("##stats"u8, 2, TableFlags.RowBackground | TableFlags.SizingFixedFit);
if (!table)
    return;
table.SetupColumn("Name"u8, TableColumnFlags.WidthStretch);
table.SetupColumn("Value"u8, TableColumnFlags.WidthFixed, 80 * Im.Style.GlobalScale);
table.HeaderRow();
foreach (var (name, value) in data)
{
    table.DrawColumn(name);            // no NextRow needed, NextColumn wraps
    table.DrawColumn($"{value}");
}
```

`ImEx.Table.Draw<T>(id, IEnumerable<T> data, Action<T> drawRow, TableFlags flags, params ReadOnlySpan<StringU8> columnTitles)`
(and `DrawTabbed`) is the one-call version.

## Filterable, sortable table (worked example)

The class-based table (`ImSharp/Extended/Classes/Table/`) draws header filters,
sorts on header click, clips rows and caches the filtered list in
`CacheManager`. Shape taken from `Penumbra/UI/AdvancedWindow/OverviewTable.cs`.

```csharp
using ImSharp;
using ImSharp.Table;

public sealed record Entry(StringPair Name, int Level, bool Enabled);

public sealed class EntryTable(IReadOnlyList<Entry> source)
    : DefaultTable<Entry>(new StringU8("##entries"u8),
        new NameColumn    { Label = new StringU8("Name"u8), Flags = TableColumnFlags.WidthStretch },
        new LevelColumn   { Label = new StringU8("Level"u8), UnscaledWidth = 70 },
        new EnabledColumn { Label = new StringU8("On"u8) })
{
    public override IEnumerable<Entry> GetItems()
        => source;

    private sealed class NameColumn : TextColumn<Entry>          // regex/text filter in the header
    {
        protected override string ComparisonText(in Entry item, int globalIndex) => item.Name;
        protected override StringU8 DisplayText(in Entry item, int globalIndex)  => item.Name;
    }

    private sealed class LevelColumn : NumberColumn<int, Entry>  // numeric filter, right-aligned
    {
        public override int ToValue(in Entry item, int globalIndex)               => item.Level;
        protected override StringU8 DisplayNumber(in Entry item, int globalIndex) => new($"{item.Level}");
        protected override string ComparisonText(in Entry item, int globalIndex)  => item.Level.ToString();
    }

    private sealed class EnabledColumn : YesNoColumn<Entry>      // tri-state yes/no filter
    {
        protected override bool GetValue(in Entry item, int globalIndex, int triEnumIndex) => item.Enabled;
    }
}
// In a Draw(): _table.Draw();   // TableBase.Draw pushes Id and fetches/creates the cache
```

- `TableData<T>.DefaultFlags` =
  `RowBackground | Sortable | BordersOuter | ScrollY | ScrollX | PreciseWidths | BordersInnerVertical | NoBordersInBodyUntilResize`;
  set `Flags` to change. `GetSize()` defaults to `Im.ContentRegion.Available`, `GetFrozenScroll()`
  to `(1, 1)`.
- `GetItems()` is only called when the cache is dirty with `IManagedCache.DirtyFlags.Custom`. To
  refresh, override `CreateCache()` and set `Dirty |= IManagedCache.DirtyFlags.Custom` from your
  change event (Penumbra does this), or call `CacheManager.Instance.SetCustomDirty(id)`.
- Build display strings once, in the item type, not in `DisplayText`; `DisplayNumber` above
  allocates per draw and is written that way only for brevity.
- Columns: `BasicColumn<T>` (`Label`, `Flags` default `NoResize`, `UnscaledWidth` default 100,
  `WidthDependsOnItems`; override `DrawColumn`, `Compare`, `DrawFilter`, `WouldBeVisible`,
  `ComputeWidth`), `TextColumn<T>`, `NumberColumn<TNumber, T>` (`NumberFilterMethod` ctor argument),
  `FlagColumn<TEnum, T>`, `TriStateFlagColumn<TEnum, T>`, `YesNoColumn<T>`.
  `LunaStyle.YesNoColumn<T>` draws FontAwesome check/cross icons instead.
- For a `WidthStretch` column `ComputeWidth` is a weight; Penumbra returns `3 / 8f`.

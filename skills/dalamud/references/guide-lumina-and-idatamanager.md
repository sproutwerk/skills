## Lumina and IDataManager

```csharp
using Lumina.Excel.Sheets;

var items = dataManager.GetExcelSheet<Item>();
var name  = items.GetRow(4).Name.ToString();                 // throws if the row is missing
if (items.TryGetRow(itemId, out var item)) { /* item is a struct */ }
var opt   = items.GetRowOrDefault(itemId)?.Name.ToString();   // Item? (nullable struct)

var territory = dataManager.GetExcelSheet<TerritoryType>().GetRow(clientState.TerritoryType);
var placeName = territory.PlaceName.ValueNullable?.Name.ToString();   // RowRef<PlaceName>
if (territory.ContentFinderCondition.IsValid) { }

var german = dataManager.GetExcelSheet<Item>(ClientLanguage.German);
var raw    = dataManager.GetExcelSheet<RawRow>(name: "GatheringType").GetRow(1).ReadStringColumn(0);
var tex    = dataManager.GetFile<Lumina.Data.Files.TexFile>("ui/icon/000000/000001.tex");
```

Lumina itself is a NuGet dependency (7.7.0) and not in the checkout; the members above are the ones
Dalamud's own source calls or the v11 Lumina 5 guide documents.

- Rows are `readonly struct`s created on access (Lumina 5, API 11). No caching needed, no `null`
  rows. `GetExcelSheet<T>` requires `T : struct, IExcelRow<T>`; subrow sheets use
  `GetSubrowExcelSheet<T>` (`T : IExcelSubrow<T>`).
- `RowRef<T>`: `RowId`, `IsValid`, `Value` (throws when invalid), `ValueNullable`. Untyped `RowRef`
  for polymorphic columns. `LazyRow` is gone.
- Arrays are `Collection<T>`. Strings are `ReadOnlySeString`; `.ToString()` is plain text,
  `ExtractText()` still exists and Dalamud uses it.
- Sheets are generated from a schema pinned at the start of a major patch; a sheet whose layout
  changed in a minor patch throws `MismatchedColumnHashException`. API 16 moves sheets to
  `Dalamud.Excel.Sheets`.
- Prefer Lumina over XIVAPI inside a plugin (technical-considerations doc).
- Helpers: `Dalamud.Utility.ItemUtil` (`GetBaseId(rawItemId)` → `(ItemId, ItemKind)`,
  `IsHighQuality`, `IsCollectible`, `IsEventItem`, `GetItemName`), `MapUtil` (`WorldToMap(Vector2 |
  Vector3, Map)`, `gameObject.GetMapCoordinates()`).


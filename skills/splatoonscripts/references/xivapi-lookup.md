# Looking up game IDs without the game (XIVAPI v2)

Covers: XIVAPI v2 query grammar, the lookup table a vendor-run script
needs (items, NPCs, shops, aetherytes, territories) and map ↔ world math.

Verified against: Splatoon 3.9.2.32 (commit f396ef6, 2026-11-13), ECommons
bc38dc7, Lifestream b363201 — the pins in `references/provenance.md`.

Paths beginning with `references/` are relative to the skill root.

## Looking up game IDs without the game (XIVAPI v2)

If the `xivapi` skill is installed, prefer its cookbook (larger, more
recently verified). The recipe below is self-contained and covers what a
vendor-run script needs.

Base `https://v2.xivapi.com/api`. Search:
`/search?sheets=<Sheet>&query=<q>&fields=<a,b.c>&limit=N`, row:
`/sheet/<Sheet>/<row>?fields=...`. Query grammar: `Field="exact"`,
`Field~"partial"`, `Field=123`, `Field.Sub~"x"`, prefix `+` to require /
`-` to exclude, space-separated terms otherwise OR-ish scored. Array fields
in `fields` use `Item[].Item[].Name` (`Item[0]` is rejected with HTTP 400).
`language=de|ja|fr|en`. URL-encode quotes/spaces (`curl -G
--data-urlencode`). An unknown field path is **silently ignored**: a
response full of `{}` objects means the path is wrong, not that the data
is absent — check the expected keys are present before reading "no stock"
into an empty result. The EXDSchema relation groupings `ReceiveItems` /
`ItemCosts` are not API field names; the flat names below are.

| Need | Query |
|---|---|
| Item id | `search?sheets=Item&query=Name~"Allagan Tomestone of"&fields=Name` |
| NPC id by name | `search?sheets=ENpcResident&query=Singular="Zircon"&fields=Singular` → row 1049079 (= DataId for `Interact`) |
| Shops an NPC opens | `sheet/ENpcBase/1049079?fields=ENpcData` → entries with `"sheet":"SpecialShop"` / `GilShop` / `Story` etc. |
| Shop name (SelectString text) | `sheet/SpecialShop/1770994?fields=Name` → "Allagan Tomestones of Mathematics (Other)"; `search?sheets=SpecialShop&query=Name~"Heliometry"` |
| Shop contents | `sheet/SpecialShop/1770994?fields=Name,UseCurrencyType,Item[].Item[].Name,Item[].ReceiveCount,Item[].ItemCost@as(raw),Item[].CurrencyCost` → one entry per slot: `Item[]` = received items, `ReceiveCount[]`, `CurrencyCost[]` = price, `ItemCost@as(raw)` = cost item id **unless** `UseCurrencyType=4` (tomestones), where it is a `TomestonesItem.Tomestones` slot (2 = Item 48 Mathematics). Empty slots have `Item[0].value = 0`. Item → shop: `search?sheets=SpecialShop&query=Item[].Item[]=45987&fields=Name` (paths recorded in the xivapi skill cookbook on SpecialShop 1770925) |
| Where an NPC stands | `search?sheets=Level&query=Object=1049079&fields=X,Y,Z,Territory.Name,Map.Id` → X/Y/Z are **world** coords usable directly as a `Move_to_point` `Point` (Y is height) |
| Aetheryte / shard ids | `search?sheets=Aetheryte&query=PlaceName.Name~"Solution Nine"&fields=PlaceName.Name,IsAetheryte,AethernetName.Name,AethernetGroup` — `IsAetheryte=true` rows are teleport targets (`Kind 0`), `false` rows are aethernet shards (`Kind 4`, name in `AethernetName`), same `AethernetGroup` = same network |
| Aetherytes in a zone | `search?sheets=Aetheryte&query=Territory=1186&fields=PlaceName.Name,AethernetName.Name,IsAetheryte` |
| Territory + map params | `sheet/TerritoryType/1186?fields=Name,PlaceName.Name,Map.OffsetX,Map.OffsetY,Map.SizeFactor` |
| World id | `search?sheets=World&query=Name="Phoenix"` |

Map ↔ world coordinates (`Map.SizeFactor`, `Map.OffsetX/OffsetY`; map Y
corresponds to world **Z**, world Y is height):

```csharp
static float ToMap(float world, int offset, int sizeFactor)
{ var c = sizeFactor / 100f; return ((world + offset) * c + 1024f) * 41f / (c * 2048f) + 1f; }
static float ToWorld(float map, int offset, int sizeFactor)
{ var c = sizeFactor / 100f; return ((map - 1f) * c * 2048f / 41f - 1024f) / c - offset; }
// mapX = ToMap(worldX, OffsetX, SizeFactor); mapY = ToMap(worldZ, OffsetY, SizeFactor)
```

This is Dalamud's `MapUtil` form (41/2048 ≈ 1/49.95); xivapi's datamining
doc uses 1/50 — the difference is under 0.1 map units. Prefer `Level` rows
over converting hand-read map coordinates. More in `references/xivapi.md`.

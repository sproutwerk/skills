# XIVAPI v2 lookups (verified live 2026-09-23, spot-checked 2026-11-13)

Re-run on 2026-11-13 with unchanged results: the Zircon `ENpcResident`
search, `ENpcBase/1049079` `ENpcData`, `SpecialShop/1770994` `Name`,
`Item/48`, `TerritoryType/1186` map parameters, `Aetheryte/235`, the `Level`
position search, the HTTP 400 for `Item[0]`, and the silently ignored
`ReceiveItems` path. The other examples were not re-run.

Base: `https://v2.xivapi.com/api`. No key. Data is current game data
(`version` in every response; older versions via `?version=<key>`, list
with `/version`). JSON only. Quote strings; URL-encode (`curl -G
--data-urlencode 'query=...'`). Interactive docs at `/api/docs` are a JS app
(not fetchable as text); the grammar below is what was verified by request.

## Endpoints

- `GET /search?sheets=<Sheet>[,<Sheet>]&query=<q>&fields=<f1,f2>&limit=<n>[&cursor=<next>]`
  → `{ next, schema, version, results: [{ score, sheet, row_id, fields }] }`
- `GET /sheet/<Sheet>/<row>?fields=<...>` → `{ row_id, fields, transient }`
- `GET /sheet/<Sheet>?rows=1,2,3&fields=...` (multiple rows; not re-verified here)
- `language=en|de|ja|fr` selects string language (`Item/47` → "Allagischer Stein der Heliometrie" with `de`).

## Query grammar (verified)

- `Name~"Allagan Tomestone of"` partial (substring, scored)
- `Singular="Zircon"` exact
- `Object=1049079`, `Territory=1186` numeric equality on row-ref fields
- `PlaceName.Name~"Solution Nine"` nested field
- `+Name~"Tomestone" +IsUntradable=true` — `+` = required clause; `-` = must not
- Comparison `>`/`<` exist in boilmaster's grammar (not verified here).

## Field selection

- Comma-separated; nested with `.`: `Map.OffsetX,Map.SizeFactor,PlaceName.Name`
- Arrays: `Item[].Item[].Name` (index syntax `Item[0]` is rejected with HTTP 400)
- Unknown field paths are silently ignored — the response contains `{}` where the
  field would be. EXDSchema `relations:` groupings (`ReceiveItems`, `ItemCosts`)
  are schema-only and behave this way; use the flat field names. Treat a result
  whose expected keys are missing as a wrong path, not as empty data.
- Omitting `fields` returns everything (large; `ENpcData` expands nested quests, hundreds of KB)

## Verified examples

```
/search?sheets=Item&query=Name~"Allagan Tomestone of"&fields=Name&limit=5
  → 30 Law, 32 Lore, 34 Verity, 38 Goetia, 45 Comedy ...  (47 = Heliometry, retired 7.4; 48 = Mathematics; 49 = Mnemonics)
/sheet/Item/45987?fields=Name,Icon,IsUntradable
  → "Cronopio Skin", Icon.id 21820, IsUntradable false
/search?sheets=ENpcResident&query=Singular="Zircon"&fields=Singular
  → row_id 1049079   (ENpcResident row id == ENpcBase row id == GameObject.DataId)
/sheet/ENpcBase/1049079?fields=ENpcData
  → ENpcData[0] = Story 1704255; [1..7] = SpecialShop 1770913, 1770911, 1770912, 1770994, 1770982, 1770980, 1770981; rest 0
/sheet/SpecialShop/1770994?fields=Name
  → "Allagan Tomestones of Mathematics (Other)"   (this is the SelectString entry text)
/sheet/SpecialShop/1770925?fields=Name,UseCurrencyType,Item[].Item[].Name,Item[].ReceiveCount,Item[].ItemCost@as(raw),Item[].CurrencyCost
  → per slot: Item [47750 "Arcanite", 0 ""], ReceiveCount [1,1], ItemCost@as(raw) [2,0,0], CurrencyCost [500,0,0]
    (from the xivapi skill cookbook transcript, not re-run here; UseCurrencyType 4 → ItemCost 2 is a
    TomestonesItem.Tomestones slot = Item 48 Mathematics, not Item 2)
/sheet/SpecialShop/1770925?fields=Item[].ReceiveItems[].Item.Name
  → "Item":[{},{},…]   (relation names are silently ignored — see Field selection)
/search?sheets=SpecialShop&query=Item[].Item[]=50977&fields=Name
  → 1770925   (item → shop)
/search?sheets=SpecialShop&query=Name~"Heliometry"&fields=Name
  → 1770770 "(Other)", 1770767 "(DoW, IL 720)", 1770768 "(DoM, IL 720)"  — retired shops, not opened by Zircon since 7.4
/search?sheets=Level&query=Object=1049079&fields=X,Y,Z,Territory.Name,Map.Id
  → X -187.699, Y 0.656, Z -30.901, Territory 1186 "x6t1", Map 856 "x6t1/00"
/sheet/Aetheryte/217?fields=PlaceName.Name,IsAetheryte,AethernetGroup,Territory.Name,AethernetName.Name
  → "Solution Nine", IsAetheryte true, AethernetGroup 15, Territory 1186, AethernetName "Solution Nine Aetheryte Plaza"
/sheet/Aetheryte/235?fields=...
  → PlaceName "", IsAetheryte false, AethernetGroup 15, AethernetName "Nexus Arcade", Territory 1186
/sheet/TerritoryType/1186?fields=Name,PlaceName.Name,Map.OffsetX,Map.OffsetY,Map.SizeFactor,Map.Id
  → "x6t1" / "Solution Nine", Map 856: OffsetX 0, OffsetY 90, SizeFactor 180, Id "x6t1/00"
```

## Sheets that matter for scripts

| Sheet | Use | Key fields |
|---|---|---|
| `Item` | item ids for inventory counts, shop purchases | `Name`, `Icon`, `IsUntradable`, `StackSize`, `ItemUICategory` |
| `ENpcResident` | NPC display name → id | `Singular` |
| `ENpcBase` | what an NPC opens | `ENpcData[]` (32 refs: `SpecialShop`, `GilShop`, `Story`, `CustomTalk`, `TopicSelect`, `DefaultTalk`, ...) |
| `SpecialShop` | currency/item exchange shops; `Name` is the menu entry | `Name`, `UseCurrencyType`, `Item[].Item[]` (received items), `Item[].ReceiveCount`, `Item[].ItemCost@as(raw)` (cost item id; a `TomestonesItem.Tomestones` slot when `UseCurrencyType=4`), `Item[].CurrencyCost`, `Item[].CostType`. Empty slots have `Item[0].value = 0`. Not `ReceiveItems`/`ItemCosts` (EXDSchema relation names, ignored by the API) |
| `GilShop` | gil shops | `Name`; items via `GilShopItem` sheet (subrows) |
| `Aetheryte` | teleport targets and aethernet shards | `PlaceName.Name`, `AethernetName.Name`, `IsAetheryte`, `AethernetGroup`, `Territory`, `Map` |
| `TerritoryType` | zone ids for `ValidTerritories`, map params | `Name` (internal, e.g. `x6t1`), `PlaceName.Name`, `Map.{OffsetX,OffsetY,SizeFactor}`, `TerritoryIntendedUse`, `ContentFinderCondition` |
| `Level` | placement of objects (NPCs, EObjs, aetherytes) | `X,Y,Z` (world coords), `Territory`, `Map`, `Object` (row id in the referenced sheet), `Type` |
| `World` | world ids/names for `Change_world` | `Name`, `DataCenter` |
| `Quest` | quest ids (`65536 + n`) | `Name`, `Id` |
| `Action` / `Status` / `Mount` | action ids, status ids | `Name` |
| `PlaceName` | names for `AethernetTeleportByPlaceNameId` | `Name` |

## Map coordinates ↔ world coordinates

`Map.SizeFactor` (e.g. 100, 200, 180), `Map.OffsetX`, `Map.OffsetY`.
World X → map X, world **Z** → map Y. World Y is height and does not
appear on the map.

```
c     = SizeFactor / 100
mapX  = ((worldX + OffsetX) * c + 1024) * 41 / (c * 2048) + 1
mapY  = ((worldZ + OffsetY) * c + 1024) * 41 / (c * 2048) + 1
worldX = ((mapX - 1) * c * 2048 / 41 - 1024) / c - OffsetX
worldZ = ((mapY - 1) * c * 2048 / 41 - 1024) / c - OffsetY
```

This is the form Dalamud's `MapUtil` uses (41/2048 ≈ 1/49.95).
xivapi/ffxiv-datamining `docs/MapCoordinates.md` gives the equivalent
`pixel = (world + offset)/100*SizeFactor + 1024; map = pixel/SizeFactor*2 + 1`
(i.e. 1/50); they differ by < 0.1 map units. The game shows map coordinates
truncated to one decimal, so hand-read map coordinates are only accurate to
about ±2.5 world units at SizeFactor 100 — use `Level` rows or in-game
`/loc`-style position dumps (`Player.Position`) for `Move_to_point`.

Check with the point (-184.95, z -28.18) in Solution Nine
(SizeFactor 180, OffsetX 0, OffsetY 90): c = 1.8; mapX =
((-184.95)*1.8 + 1024)*41/(1.8*2048) + 1 ≈ 8.7; mapY =
((-28.18+90)*1.8 + 1024)*41/3686.4 + 1 ≈ 13.6 → roughly (8.7, 13.6), which
is the Nexus Arcade area.

# Excel sheets: GenericHelpers sheet access, ExcelServices helpers, Job enum, territory lists

Covers: the sheet-access members of `GenericHelpers` (`GetRow`, `FindRow`, `GetSheet`, `GetRef`
...), the `ExcelServices` helper classes, enums and custom sheets, the `TerritoryEnumeration`
constants and `TerritoryName`. Verified against ECommons `bc38dc7` (2026-11-08, v3.2.1.26, Dalamud
API 15); nothing here was compiled. Paths are relative to the skill root.

Raw Lumina / `IDataManager` usage (`ExcelSheet<T>`, `RowRef<T>`, `ReadOnlySeString`) -> `dalamud`
skill. Looking up an ID without the game client -> `xivapi` skill. `Player.*` properties that return
`RowRef<T>` -> `references/topic-game.md`.

## Namespaces

```csharp
using ECommons;                                   // GenericHelpers (static partial class), TerritoryName
using static ECommons.GenericHelpers;             // optional: call GetRow<T>() etc. unqualified
using ECommons.ExcelServices;                     // Excel*Helper classes, Job, GrandCompany, ItemRarity, BaseParamEnum,
                                                  // EquipSlotCategoryEnum, TerritoryIntendedUseEnum, CommonlyUsedAttribute
using ECommons.ExcelServices.Enums;               // ItemUICategoryEnum (the only enum in this sub-namespace)
using ECommons.ExcelServices.Sheets;              // ExtendedBaseParam, QuestDialogueText, IRowExtension<,>
using ECommons.ExcelServices.TerritoryEnumeration;// Dungeons, Raids, Inns, ... TerritoryRegion
using Lumina.Excel;                               // RowRef<T>, ExcelSheet<T>
using Lumina.Excel.Sheets;                        // Item, TerritoryType, World, ClassJob ...
```

Everything goes through `Svc.Data` (`IDataManager`), so `ECommonsMain.Init` must have run. No module
flag, no hooks, nothing for `ECommonsMain.Dispose()` to clean up. Sheet reads are safe from any
thread; the only member here that touches game memory is `ExcelActionHelper.GetActionCooldown` (and
`ExcelItemHelper.GetCooldown`, which calls it).

Folder layout trap: files under `ExcelServices/Enums/` declare namespace `ECommons.ExcelServices`,
except `ItemUICategoryEnum.cs`, which declares `ECommons.ExcelServices.Enums`.

## Sheet access in `GenericHelpers`

Source: `references/ecommons/ECommons/GenericHelpers/ExcelHelpers.cs`. These are members of `public
static unsafe partial class GenericHelpers` in namespace `ECommons` (not `ECommons.GenericHelpers`;
that is only the folder name).

`language` is `Dalamud.Game.ClientLanguage?` unless stated otherwise. `null` means
`Svc.ClientState.ClientLanguage`.

### Sheets

```csharp
public static ExcelSheet<T> GetSheet<T>(ClientLanguage? language = null) where T : struct, IExcelRow<T>
public static ExcelSheet<T> GetSheet<T>(string sheetName, ClientLanguage? language = null) where T : struct, IExcelRow<T>
public static SubrowExcelSheet<T> GetSubrowSheet<T>(ClientLanguage? language = null) where T : struct, IExcelSubrow<T>
public static int GetRowCount<T>() where T : struct, IExcelRow<T>
public static IEnumerable<T> AllRows<T>(this SubrowExcelSheet<T> subrowSheet) where T : struct, IExcelSubrow<T>
```

`AllRows` flattens a subrow sheet into every subrow of every row.

### Rows

```csharp
public static T?   GetRow<T>(uint rowId, ClientLanguage? language = null) where T : struct, IExcelRow<T>
public static T?   FindRow<T>(Func<T, bool> predicate, ClientLanguage? language = null) where T : struct, IExcelRow<T>
public static T[]  FindRows<T>(Func<T, bool> predicate, ClientLanguage? language = null) where T : struct, IExcelRow<T>
public static bool TryGetRow<T>(uint rowId, out T row) where T : struct, IExcelRow<T>
public static bool TryGetRow<T>(uint rowId, ClientLanguage? language, out T row) where T : struct, IExcelRow<T>
public static bool TryGetRow<T>(string sheetName, uint rowId, out T row) where T : struct, IExcelRow<T>
public static bool TryGetRow<T>(string sheetName, uint rowId, ClientLanguage? language, out T row) where T : struct, IExcelRow<T>
public static bool TryFindRow<T>(Predicate<T> predicate, out T row) where T : struct, IExcelRow<T>
public static bool TryFindRow<T>(Predicate<T> predicate, ClientLanguage? language, out T row) where T : struct, IExcelRow<T>
```

- `GetRow<T>` returns `null` for a missing row (it wraps `GetRowOrDefault`); it never throws. This
  differs from Lumina's own `ExcelSheet<T>.GetRow`, which throws.
- `FindRow` / `FindRows` / `TryFindRow` enumerate the whole sheet on every call. Cache the result.
- Note the delegate types: `FindRow`/`FindRows` take `Func<T, bool>`, `TryFindRow` takes
  `Predicate<T>`.

### Subrows

```csharp
public static T?   GetRow<T>(uint rowId, ushort subRowId, ClientLanguage? language = null) where T : struct, IExcelSubrow<T>
public static SubrowCollection<T>? GetSubRow<T>(uint rowId, ClientLanguage? language = null) where T : struct, IExcelSubrow<T>
public static bool TryGetRow<T>(uint rowId, ushort subRowId, out T row) where T : struct, IExcelSubrow<T>
public static bool TryGetRow<T>(uint rowId, ushort subRowId, ClientLanguage? language, out T row) where T : struct, IExcelSubrow<T>
public static T?   FindSubrow<T>(Func<T, bool> predicate, ClientLanguage? language = null) where T : struct, IExcelSubrow<T>
```

### Raw rows

```csharp
public static bool TryGetRawRow(string sheetName, uint rowId, out RawRow rawRow)
public static bool TryGetRawRow(string sheetName, uint rowId, ClientLanguage? language, out RawRow rawRow)
```

### RowRef helpers

```csharp
public static RowRef<T> WithLanguage<T>(this RowRef<T> rowRef, ClientLanguage? language) where T : struct, IExcelRow<T>
public static RowRef<T> WithLanguage<T>(this RowRef<T> rowRef, Lumina.Data.Language? language) where T : struct, IExcelRow<T>
public static bool TryGetValue<T>(this RowRef<T> rowRef, out T value) where T : struct, IExcelRow<T>
public static TExtension GetExtension<TExtension, TBase>(this TBase row)
    where TExtension : struct, IExcelRow<TExtension>, IRowExtension<TExtension, TBase>
    where TBase : struct, IExcelRow<TBase>
```

- `WithLanguage(ClientLanguage?)` maps only `Japanese`, `German`, `French`, `English`; any other
  value (and `null`) produces a `RowRef` with `null` language, i.e. the default language.
- `TryGetValue` is `ValueNullable != null` plus the value. It is an extension on `RowRef<T>`; the
  source also calls `.TryGetValue(out var data)` on a `T?`
  (`ExcelItemHelper.Get(id).TryGetValue(...)`), which is a different extension defined elsewhere in
  ECommons.

### Static extension members on sheet row types

Declared with a C# 14 extension block, `extension<T>(IExcelRow<T> sheet) where T : struct,
IExcelRow<T>`, so they are called on the row **type**:

```csharp
public static RowRef<T> GetRef(uint id, Lumina.Data.Language? language = null)   // [OverloadResolutionPriority(1)]
public static RowRef<T> GetRef(int id, Lumina.Data.Language? language = null)
public static T         Get(uint id, Dalamud.Game.ClientLanguage? language = null)   // GetRow: throws on a missing row
public static bool      TryGet(uint id, out T value, Dalamud.Game.ClientLanguage? language = null)
public static ExcelSheet<T> Values { get; }                                          // the whole sheet, default language
```

Usage as it appears inside ECommons itself:

```csharp
TerritoryType.GetRef(Svc.ClientState.TerritoryType)        // Player.cs
ClassJob.GetRef((uint)job)                                  // Job.cs, JobExtensions.GetGameData
Aetheryte.GetRef(PlayerState.Instance()->HomeAetheryteId)   // Player.cs
```

Requirements: `using ECommons;` in scope and a compiler that supports extension blocks. `GetRef`
takes `Lumina.Data.Language?` while `Get`/`TryGet` take `Dalamud.Game.ClientLanguage?`. Only
`GetRef` is exercised by the ECommons source; `Get`, `TryGet` and `Values` are declared the same way
but have no in-repo caller.

### Text from sheet strings

Source: `references/ecommons/ECommons/GenericHelpers/StringHelpers.cs` (same `GenericHelpers` class).

```csharp
public static string GetText(this ReadOnlySeString s, bool onlyFirst = false)
public static string GetText(this Utf8String s, bool onlyFirst = false)
public static string GetText(this SeString seStr, bool onlyFirst = false)
```

Keeps only text and auto-translate payloads; `onlyFirst` stops after the first one. The
`ExtractText(...)` overloads with the same parameters are `[Obsolete]` ("Dalamud has added their own
ExtractText method for Lumina strings that is not compatible with ECommons ... ExtractText method is
renamed to GetText"). On a `ReadOnlySeString`, `s.ExtractText()` now binds to Dalamud/Lumina's
instance method, not ECommons'. The `Utf8String` overload ignores `onlyFirst` (it always passes
`false`).

### Examples

```csharp
// YesAlready Features/SelectYesNo.cs
if (GenericHelpers.GetRow<Item>(itemId) is { } item)
{
    if (GenericHelpers.TryGetRow<WKSItemInfo>(item.AdditionalData.RowId, out var wksItem)) { }
}

// YesAlready YesAlready.cs
TerritoryNames = GenericHelpers.FindRows<TerritoryType>(r => r.PlaceName.IsValid && !r.PlaceName.Value.Name.IsEmpty)
    .Select((r, n) => (r.RowId, PlaceName: r.PlaceName.Value.Name.ToString()))
    .ToDictionary(t => t.RowId, t => t.PlaceName);

// ECommons GameHelpers/Map.cs, with `using static ECommons.GenericHelpers;`
var map = GetRow<Lumina.Excel.Sheets.Map>(mapId);
var scale = (map?.SizeFactor ?? 100) * 0.01f;
```

## ExcelItemHelper

Source: `references/ecommons/ECommons/ExcelServices/ExcelItemHelper.cs`.

```csharp
public static Item? Get(int rowId)
public static Item? Get(uint rowId)                                   // null when missing
public static string GetName(uint id, bool includeID = false)
public static string GetName(this Item? item, bool includeID = false)
public static string GetName(this Item item, bool includeID = false)
public static float  GetCooldown(this Item item)
public static int    GetStat(this Item item, BaseParamEnum param, bool isHq = false)
public static int    GetStat(this InventoryItem item, BaseParamEnum param)
public static int    GetStatCap(this Item item, BaseParamEnum baseParam)
public static int    GetStatCap(this InventoryItem item, BaseParamEnum baseParam)
public static ItemRarity GetRarity(this Item item)
public static uint   GetGCSeals(this Item item)
```

- `GetName(uint)`: `"#<id>"` when the row does not exist. `GetName(Item?)`: `"? Unknown ?"` for
  null, `"#<id>"` when the name is empty. Names go through `GetText()` (payloads stripped) and are
  cached in two static dictionaries for the plugin's lifetime (one per `includeID` value), in the
  language of the first call.
- `includeID` formats as `"<name> #<id>"`. Bug at this commit: on a cache miss the method stores the
  formatted string but returns the bare name, so the **first** call with `includeID: true` for a
  given item returns the name without the id; later calls return it with the id.
- `Get`/`GetName` expect a base item id. For an id taken from an inventory slot strip the HQ offset
  first (`id % 1000000`), as `GetStat(InventoryItem)` does.
- `GetStat(Item, ...)`: sum of matching `BaseParam` entries, plus `BaseParamSpecial` when `isHq`.
- `GetStat(InventoryItem, ...)`: base stat (HQ detected by `GetItemId() > 1000000`) plus melded
  materia, clamped to `GetStatCap` when the cap is positive.
- `GetStatCap`: `LevelItem` value times the slot percentage from `BaseParam`, divided by 1000.
  Returns `-1` for an unsupported param (for example `ElementalBonus`), an unsupported equip slot,
  or a missing item.
- `GetCooldown`: remaining recast of the item's action in seconds, `0` when the item has no action
  or no cooldown group.
- `GetGCSeals`: expert-delivery seals; `0` when `Desynth == 0` or the rarity is not Green, Pink or
  Blue.

```csharp
// AutoRetainer Helpers/ItemLevel.cs:42 (slot is InventoryItem*)
gathering += slot->GetStat(BaseParamEnum.Gathering);

// AutoRetainer Modules/GcHandin/GCContinuation.cs:408
if (!data.Value.GetRarity().EqualsAny(ItemRarity.Green, ItemRarity.Pink, ItemRarity.Blue)) continue;

// AutoRetainer Scheduler/Tasks/TaskEntrustDuplicates.cs:75
InternalLog.Debug($"added item: {ExcelItemHelper.GetName(add.Item1, true)}");
```

## ExcelActionHelper

Source: `references/ecommons/ECommons/ExcelServices/ExcelActionHelper.cs`.

```csharp
public static float  GetActionCooldown(uint id)
public static string GetActionName(this Lumina.Excel.Sheets.Action? dataNullable, bool forceIncludeID = false)
public static string GetActionName(uint id, bool forceIncludeID = false)
```

- `GetActionCooldown`: remaining seconds on the action's recast group, or the larger of that and the
  additional cooldown group. Reads `ActionManager` (game memory; framework thread). Uses
  `GetRow(id)`, so an invalid action id throws.
- `GetActionName`: `"null"` for a null row, `"#<id>"` for a missing row or empty name, otherwise the
  name, prefixed with `"#<id> "` when `forceIncludeID` is true (id first, unlike item names).

## ExcelJobHelper and the `Job` enum

Source: `references/ecommons/ECommons/ExcelServices/ExcelJobHelper.cs`,
`references/ecommons/ECommons/ExcelServices/Enums/Job.cs`.

`public enum Job : byte` in namespace `ECommons.ExcelServices`; values equal `ClassJob` row ids:

| | | | |
|---|---|---|---|
| `ADV = 0` | `GSM = 11` | `DRG = 22` | `AST = 33` |
| `GLA = 1` | `LTW = 12` | `BRD = 23` | `SAM = 34` |
| `PGL = 2` | `WVR = 13` | `WHM = 24` | `RDM = 35` |
| `MRD = 3` | `ALC = 14` | `BLM = 25` | `BLU = 36` |
| `LNC = 4` | `CUL = 15` | `ACN = 26` | `GNB = 37` |
| `ARC = 5` | `MIN = 16` | `SMN = 27` | `DNC = 38` |
| `CNJ = 6` | `BTN = 17` | `SCH = 28` | `RPR = 39` |
| `THM = 7` | `FSH = 18` | `ROG = 29` | `SGE = 40` |
| `CRP = 8` | `PLD = 19` | `NIN = 30` | `VPR = 41` |
| `BSM = 9` | `MNK = 20` | `MCH = 31` | `PCT = 42` |
| `ARM = 10` | `WAR = 21` | `DRK = 32` | `BST = 43` |

`JobExtensions` (same file): `public static RowRef<ClassJob> GetGameData(this Job job)`.

`ExcelJobHelper`:

```csharp
public static readonly ReadOnlyDictionary<Job, Job> Upgrades;   // GLA->PLD, PGL->MNK, MRD->WAR, LNC->DRG, ARC->BRD,
                                                               // CNJ->WHM, THM->BLM, ACN->SMN, ROG->NIN
public static Job  GetUpgradedJob(this Job j)      // j itself when there is no upgrade
public static Job  GetDowngradedJob(this Job j)    // j itself when there is no base class
public static bool IsUpgradeable(this Job j)
public static bool IsCombat(this Job j)            // ClassJob.Role > 0
public static bool IsDol(this Job j)               // ClassJobCategory.RowId == 32
public static bool IsDoh(this Job j)               // == 33
public static bool IsDom(this Job j)               // == 31
public static bool IsDow(this Job j)               // == 30
public static bool IsTank(this Job j)              // Role == 1
public static bool IsHealer(this Job j)            // Role == 4
public static bool IsDps(this Job j)               // melee or ranged
public static bool IsMeleeDps(this Job j)          // Role == 2
public static bool IsRangedDps(this Job j)         // Role == 3
public static bool IsPhysicalRangedDps(this Job j) // ranged && DoW
public static bool IsMagicalRangedDps(this Job j)  // ranged && DoM
public static ClassJob GetData(this Job j)         // GetRow: throws on an invalid value
public static Job  GetJob(this ClassJob cj)
public static int  GetIcon(this Job j)             // 62143 for ADV, otherwise 62100 + (int)j
public static ClassJob? GetJobByName(string name)            // case-insensitive, client language
public static bool TryGetJobByName(string name, out ClassJob result)
public static ClassJob? GetJobByAbbreviation(string name)    // case-insensitive
public static bool TryGetJobByAbbreviation(string name, out ClassJob result)
public static ClassJob? GetJobById(uint id)
public static bool TryGetJobById(uint id, out ClassJob result)
public static string GetJobNameById(uint id)                 // null when missing
public static ClassJob[] GetCombatJobs()
public static bool IsJobInCategory(this ClassJobCategory cat, Job job)
```

- `GetData` returns the full `ClassJob` row (throws when invalid), `GetGameData` returns a
  `RowRef<ClassJob>` (lazy, check `ValueNullable`).
- `Upgrades` has no entry for ACN -> SCH; `GetUpgradedJob(Job.ACN)` is `SMN` and
  `GetDowngradedJob(Job.SCH)` is `SCH`.
- `GetCombatJobs()` filters `Role` 2 or 3 only. Despite the name it returns DPS classes and jobs and
  **excludes tanks and healers**. Use `IsCombat()` for "any battle job".
- `IsJobInCategory` checks the category's boolean column for the job; `BST` reads the column Lumina
  still names `Unknown0`.
- Every `Is*` call performs a sheet row lookup.

```csharp
// PandorasBox Features/Actions/AutoMeditation.cs
var isMonk = Player.Job == Job.MNK;

if (Player.Job.IsDoh() || Player.Job.IsDol()) { }
var icon = (uint)Player.Job.GetIcon();                 // icon id for ITextureProvider
var max = Player.GetLevel(Player.Job.GetUpgradedJob());
```

## ExcelTerritoryHelper

Source: `references/ecommons/ECommons/ExcelServices/ExcelTerritoryHelper.cs`.

```csharp
public static bool IsSanctuary(uint territoryType)
public static TerritoryIntendedUseEnum GetTerritoryIntendedUse(this TerritoryType value)
public static TerritoryIntendedUseEnum GetTerritoryIntendedUse(uint value)
public static bool NameExists(uint TerritoryType)
public static bool NameExists(this TerritoryType? t)
public static string GetName(int TerritoryType, bool includeID = false, ClientLanguage? language = null)
public static string GetName(uint TerritoryType, bool includeID = false, ClientLanguage? language = null) // [OverloadResolutionPriority(1)]
public static TerritoryType? Get(uint ID)
public static string GetBG(this TerritoryType? t)
public static string GetBG(uint ID)
```

- `IsSanctuary`: membership in the constants of `MainCities`, `Inns`, `ResidentalAreas` and `Houses`
  (collected by reflection on first call, one Verbose log line per constant). It is a fixed list,
  not the game's sanctuary/rested-XP flag.
- `GetTerritoryIntendedUse(uint)` returns `(TerritoryIntendedUseEnum)37707` for a missing row, a
  value that matches no enum member.
- `GetName`: duty name (`ContentFinderCondition.Name`) if present, else place name, else `"#<id>"`;
  first character upper-cased. With `includeID` the result is `"#<id> | <name>"`. Not cached.
- `GetBG`: the territory's `Bg` path string; null when the row is missing.

```csharp
// AutoRetainer UI/NeoUI/AdvancedEntries/DebugSection/DebugMulti.cs:79
ImGuiEx.Text($"Is in sanctuary: {ExcelTerritoryHelper.IsSanctuary(Svc.ClientState.TerritoryType)}");
```

## ExcelWorldHelper

Source: `references/ecommons/ECommons/ExcelServices/ExcelWorldHelper.cs`.

```csharp
public enum Region { JP = 1, NA = 2, EU = 3, OC = 4 }          // nested: ExcelWorldHelper.Region

public static bool   IsPublic(this World w)
public static Region GetRegion(this World world)
public static World? Get(string name, bool onlyPublic = false)
public static World? Get(uint id, bool onlyPublic = false)
public static World? Get(uint id)
public static World? Get(int id)
public static bool   TryGet(string name, out World result)
public static bool   TryGet(uint id, out World result)
public static string GetName(int id)
public static string GetName(uint id)                           // [OverloadResolutionPriority(1)]
public static string GetName(int id, bool includeId)            // 3.2.1.23+
public static string GetName(uint id, bool includeId)           // 3.2.1.23+; appends "#<id>"
public static World[] GetPublicWorlds(Region? region = null)
public static World[] GetPublicWorlds(uint dataCenter)          // 0 = all data centres
public static WorldDCGroupType[] GetDataCenters(Region? region = null, bool checkForPublicWorlds = false)
public static WorldDCGroupType[] GetDataCenters(IEnumerable<Region> regions, bool checkForPublicWorlds = false)
```

`GetName` for a missing world depends on the version. Up to 3.2.1.22 it returns `null`. From
3.2.1.23 every overload ends in a string concatenation, so a missing world gives `""`, or `"#<id>"`
with `includeId: true`. Test with `string.IsNullOrEmpty`, or use `TryGet`.

Obsolete, with the replacement named in the attribute:

| Obsolete | Replacement |
|---|---|
| `GetWorldByName(string name)` | `Get` |
| `TryGetWorldByName(string name, out World result)` | `TryGet` |
| `GetWorldById(uint id)` | `Get` |
| `GetWorldNameById(uint id)` | `GetName` |
| `GetPublicWorldById(uint id)` | `Get` (it calls `Get(id, true)`) |
| `GetPublicWorldNameById(uint id)` | `GetName` (it calls `Get(id, true)?.Name`) |

- `w.IsPublic()` (method) is the ECommons extension; `w.IsPublic` (property) is the sheet column.
  The extension returns the column when true and additionally accepts Korean and Chinese worlds when
  the client language value is 4 or higher (`RowId > 2000 && UserType == 201` with an `InternalName`
  starting with `Kr` other than `KrOmega`; `RowId > 1000 && UserType == 101`).
- `GetRegion` derives the region from the world's data centre row (`WorldDCGroupType.Region`), and
  returns `0` (no enum member) when the data centre row is missing. `World.Region` itself is banned
  inside ECommons as unreliable (see `references/topic-game.md`, BannedSymbols).
- `onlyPublic` in `Get(..., onlyPublic)` means "region is one of JP/NA/EU/OC". It does not call
  `IsPublic()`.
- `Get(string)` is case-insensitive, prefers rows 1180, 1183, 1186, 1192, 1200, 1201 when names
  collide, and caches hits by name only. A hit cached by a call with `onlyPublic: false` is returned
  to later calls with `onlyPublic: true`. Misses are not cached, so each miss scans and sorts the
  sheet.
- `Get(id)` with one argument binds to the overload without the optional parameter.
- `TryGet` sets `result` to `default` on failure.

```csharp
// AutoRetainer Helpers/Utils.cs:486
ExcelWorldHelper.Get(thisRef.World)?.GetRegion() != ExcelWorldHelper.Region.JP

var na = ExcelWorldHelper.GetPublicWorlds(ExcelWorldHelper.Region.NA);
var dcs = ExcelWorldHelper.GetDataCenters(ExcelWorldHelper.Region.EU, checkForPublicWorlds: true);
var name = ExcelWorldHelper.GetName(Player.HomeWorld.RowId);
```

## Enums

All in namespace `ECommons.ExcelServices` unless noted. Sources under
`references/ecommons/ECommons/ExcelServices/Enums/`.

| Enum | Underlying | Values |
|---|---|---|
| `Job` | `byte` | see table above |
| `GrandCompany` | `int` | `Unemployed = 0, Maelstrom = 1, TwinAdder = 2, ImmortalFlames = 3` |
| `ItemRarity` | `byte` | `White = 1, Green = 2, Blue = 3, Purple = 4, Pink = 7` |
| `BaseParamEnum` | `uint` | `Strength = 1, Dexterity = 2, Vitality = 3, Intelligence = 4, Mind = 5, Piety = 6, HP = 7, MP = 8, TP = 9, GP = 10, CP = 11, Delay = 14, Tenacity = 19, AttackPower = 20, Defense = 21, DirectHitRate = 22, Evasion = 23, MagicDefense = 24, CriticalHit = 27, AttackMagicPotency = 33, HealingMagicPotency = 34, ElementalBonus = 36, Determination = 44, SkillSpeed = 45, SpellSpeed = 46, Haste = 47, Craftsmanship = 70, Control = 71, Gathering = 72, Perception = 73` |
| `EquipSlotCategoryEnum` | `uint` | `WeaponMainHand = 1, OffHand = 2, Head = 3, Body = 4, Gloves = 5, Waist = 6, Legs = 7, Feet = 8, Ears = 9, Neck = 10, Wrists = 11, Ring = 12, WeaponTwoHand = 13, BodyHead = 15, BodyGlovesLegsFeet = 16, SoulCrystal = 17, LegsFeet = 18, BodyHeadGlovesLegsFeet = 19, BodyGlovesLegs = 20, BodyLegsFeet = 21, BodyGloves = 22` |
| `TerritoryIntendedUseEnum` | `uint` | see below |
| `ECommons.ExcelServices.Enums.ItemUICategoryEnum` | `int` | 113 members `None = 0` .. `Outfits = 112`, mirroring `ItemUICategory` row ids (`Medicine = 44`, `Meal = 46`, `Materia = 58`, `Crystal = 59`, `Minion = 81`, `Currency = 100` ...); open the source for the full list |

`TerritoryIntendedUseEnum` (members tagged `[CommonlyUsed]` in source are marked `*`): `City_Area* =
0, Open_World* = 1, Inn* = 2, Dungeon* = 3, Variant_Dungeon* = 4, Gaol = 5, Starting_Area = 6,
Quest_Area = 7, Alliance_Raid* = 8, Quest_Battle = 9, Trial* = 10, Quest_Area_2 = 12,
Residential_Area* = 13, Housing_Instances* = 14, Quest_Area_3 = 15, Raid = 16, Raid_2* = 17,
Frontline = 18, Chocobo_Square = 20, Restoration_Event = 21, Sanctum = 22, Gold_Saucer = 23,
Lord_of_Verminion = 25, Diadem = 26, Hall_of_the_Novice = 27, Crystalline_Conflict = 28,
Quest_Battle_2 = 29, Barracks = 30, Deep_Dungeon* = 31, Seasonal_Event = 32, Treasure_Map_Duty = 33,
Seasonal_Event_Duty = 34, Battlehall = 35, Crystalline_Conflict_2 = 37, Diadem_2 = 38, Rival_Wings =
39, Unknown_1 = 40, Eureka* = 41, Seasonal_Event_2 = 43, Leap_of_Faith = 44, Masked_Carnivale = 45,
Ocean_Fishing = 46, Diadem_3 = 47, Bozja* = 48, Island_Sanctuary* = 49, Battlehall_2 = 50,
Battlehall_3 = 51, Large_Scale_Raid* = 52, Large_Scale_Savage_Raid* = 53, Quest_Area_4 = 54,
Tribal_Instance = 56, Criterion_Duty* = 57, Criterion_Savage_Duty* = 58, Blunderville = 59,
Occult_Crescent* = 61`.

Note `Raid = 16` is not tagged commonly used while `Raid_2 = 17` is. FFXIVClientStructs has its own
`FFXIVClientStructs.FFXIV.Client.Enums.TerritoryIntendedUse`; `Player.CsTerritoryIntendedUseEnum`
returns that one.

`CommonlyUsedAttribute` (`references/ecommons/ECommons/ExcelServices/CommonlyUsedAttribute.cs`):
`public class CommonlyUsedAttribute : Attribute {}`. A bare marker with no members and no
`AttributeUsage`; it only tags `TerritoryIntendedUseEnum` members for UI code that wants a short
list.

```csharp
// PandorasBox Features/Chat/AutoOpenCoords.cs, with `using static ECommons.ExcelServices.TerritoryIntendedUseEnum;`
if (Player.TerritoryIntendedUseEnum is not (City_Area or Open_World or Inn or Starting_Area)) return;
```

## Custom sheets (`ECommons.ExcelServices.Sheets`)

Sources under `references/ecommons/ECommons/ExcelServices/Sheets/`.

```csharp
public interface IRowExtension<out TExtension, in TBase> : IExcelRow<TExtension>
    where TBase : struct, IExcelRow<TBase>
    where TExtension : struct, IExcelRow<TExtension>, IRowExtension<TExtension, TBase>
{
    static virtual TExtension GetExtended(IExcelRow<TBase> baseRow);   // default: GetExcelSheet<TExtension>().GetRow(baseRow.RowId)
}

[Sheet("BaseParam")]
public readonly unsafe struct ExtendedBaseParam(ExcelPage page, uint offset, uint row) : IRowExtension<ExtendedBaseParam, BaseParam>
{
    public ExcelPage ExcelPage { get; }
    public BaseParam BaseParam { get; }                       // the stock row at the same position
    public Collection<ushort> EquipSlotCategoryPct { get; }   // 23 entries; index 0 is always 0
    public uint RowOffset { get; }
    public uint RowId { get; }
    public static ExtendedBaseParam Create(ExcelPage page, uint offset, uint row);
}

[Sheet]
public struct QuestDialogueText(ExcelPage page, uint offset, uint row) : IExcelRow<QuestDialogueText>
{
    public ExcelPage ExcelPage { get; }
    public ReadOnlySeString Key { get; }
    public ReadOnlySeString Value { get; }
    public uint RowOffset { get; }
    public uint RowId { get; }
    public static QuestDialogueText Create(ExcelPage page, uint offset, uint row);
}
```

- `ExtendedBaseParam` (credited in source to Caraxi / SimpleTweaks) exposes the per-equip-slot
  percentage columns as one collection indexed by `EquipSlotCategory` row id. Obtain it with
  `baseParamRow.GetExtension<ExtendedBaseParam, BaseParam>()` or `GetSheet<ExtendedBaseParam>()`.
- `QuestDialogueText` has a bare `[Sheet]` attribute (no sheet name), so it must be loaded by name.
  It reads any two-column key/value text sheet such as the `custom/NNN/...` dialogue sheets. Row ids
  and sheet names have to be looked up (`xivapi` skill); ECommons does not list them.

```csharp
// AutoRetainer Services/CabinetManager.cs:103
var text = Svc.Data.GetExcelSheet<QuestDialogueText>(name: "custom/000/CmnDefCabinet_00082").GetRow(1).Value.GetText();

// equivalent through GenericHelpers
if (GenericHelpers.TryGetRow<QuestDialogueText>("custom/000/CmnDefCabinet_00082", 1, out var row))
    text = row.Value.GetText();
```

## TerritoryEnumeration

Namespace `ECommons.ExcelServices.TerritoryEnumeration`. These files are **not in the reference
copy** (`ExcelServices/TerritoryEnumeration/*` is excluded); read them in an ECommons checkout at
`ECommons/ECommons/ExcelServices/TerritoryEnumeration/`.

Each class is a `public static class` of `public const ushort <Name> = <TerritoryType row id>;` plus

```csharp
public static ushort[] List { get; }   // every constant of the class, built once by reflection
```

| Class | Constants | Content |
|---|---|---|
| `MainCities` | 19 | city territories, e.g. `Limsa_Lominsa_Upper_Decks = 128`, `Uldah_Steps_of_Nald = 130`, `New_Gridania = 132`, `Foundation = 418`, `Kugane = 628`, `The_Crystarium = 819`, `Old_Sharlayan = 962`, `Tuliyollal = 1185` |
| `Inns` | 8 | `Mizzenmast_Inn = 177`, `The_Hourglass = 178`, `The_Roost = 179`, `Cloud_Nine = 429`, `Bokairo_Inn = 629`, `The_Pendants_Personal_Suite = 843`, `Andron = 990`, `The_Forard_Cabins = 1205` |
| `ResidentalAreas` (sic) | 5 | `Mist = 339`, `The_Lavender_Beds = 340`, `The_Goblet = 341`, `Shirogane = 641`, `Empyreum = 979` |
| `Houses` | 41 | house interiors, private chambers, company workshops, apartment lobbies and apartments per district, e.g. `Private_Cottage_Mist = 282`, `Company_Workshop_Mist = 423`, `Topmast_Apartment = 608` |
| `Prisons` | 1 | `Mordion_Gaol = 176` |
| `OpenAreas` | 52 | overworld zones, e.g. `Middle_La_Noscea = 134`, `Mor_Dhona = 156`, `Wolves_Den_Pier = 250`, `Solution_Nine = 1186` |
| `Dungeons` | 114 | dungeons and some instanced training/guildhest territories, e.g. `the_Wanderers_Palace = 159`, `Amdapor_Keep = 167` |
| `Trials` | 90 | e.g. `the_Bowl_of_Embers_Hard = 292`, `the_Navel_Extreme = 296` |
| `Raids` | 127 | 8-player raids incl. savage, e.g. `the_Binding_Coil_of_Bahamut_Turn_1 = 241`, `Alexander_The_Fist_of_the_Father_Savage = 449` |
| `AllianceRaids` | 16 | `the_World_of_Darkness = 151` .. `Jueno_The_First_Walk = 1248` |
| `VariantDungeons` | 3 | `The_Sildihn_Subterrane = 1069`, `Mount_Rokkon = 1137`, `Aloalo_Island = 1176` |

`public enum TerritoryRegion : uint` (same namespace, no `List`): `AbalanthiasSpine = 497, Coerthas
= 25, Dravania = 498, GyrAbania = 2400, Hingashi = 2402, Ilsabard = 3703, LaNoscea = 22, MorDhona =
26, Norvrandt = 2950, Othard = 2401, Thanalan = 24, TheBlackShroud = 23, TheHighSeas = 3443,
TheNorthernEmpty = 3702, TheSeaOfStars = 3704, TheWorldUnsundered = 3705`. No ECommons code in the
assigned area consumes it; what the numbers index is not stated in source *(unverified: they look
like `PlaceName` row ids of `TerritoryType.PlaceNameRegion`)*.

Notes:

- Capitalisation is inconsistent (`the_Wanderers_Palace`, `The_Copied_Factory`); rely on completion
  or the source, do not guess names.
- The lists are hand-maintained and lag behind patches. For "is this a dungeon / inn / city" prefer
  `Player.TerritoryIntendedUseEnum` or `ExcelTerritoryHelper.GetTerritoryIntendedUse(id)`; use the
  constants when a specific territory id is needed.
- `MainCities.Solution_Nine = 1205` has the same value as `Inns.The_Forard_Cabins = 1205`, while
  `OpenAreas.Solution_Nine = 1186`. One of the first two is wrong in source; do not use
  `MainCities.Solution_Nine` without checking the id (`xivapi` skill).
- The constants are `ushort`; `Svc.ClientState.TerritoryType` and `Player.Territory.RowId` are
  `uint`. `==` against a constant compiles as is, `List.Contains(...)` needs a cast to `ushort`.

```csharp
// AutoRetainer AutoRetainer.cs:640
if (Svc.ClientState.TerritoryType == Prisons.Mordion_Gaol) { }

// AutoRetainer UI/NeoUI/AdvancedEntries/DebugSection/DebugMulti.cs:76
var inResidential = ResidentalAreas.List.Contains((ushort)Svc.ClientState.TerritoryType);

// With the new Player class
var inInn = Inns.List.Contains((ushort)Player.Territory.RowId);
```

## TerritoryName

Source: `references/ecommons/ECommons/TerritoryName.cs`. `public static class TerritoryName` in
namespace `ECommons`.

```csharp
public static string GetTerritoryName(uint id, Lumina.Data.Language? lang = null)
```

Returns a debug-style label and caches it per id:

| Case | Result |
|---|---|
| place name and duty name present | `"<id> \| <zone> (<duty>)"` |
| place name only | `"<id> \| <zone>"` |
| row missing or place name empty | `"<id>"` |

The cache key is the id alone, so the `lang` of the first call for an id wins for the rest of the
plugin's lifetime. `ECommons.GameHelpers.Content` calls it with `Language.English` and parses the
result, which means any earlier call for the same territory in another language changes what
`Content.TerritoryName` returns. For a display name use `ExcelTerritoryHelper.GetName` instead.

## Choosing between the territory-name helpers

| Need | Use |
|---|---|
| user-facing name, duty name preferred, optional language | `ExcelTerritoryHelper.GetName(id, includeID, language)` |
| log/debug label with id | `TerritoryName.GetTerritoryName(id)` |
| the row itself | `GetRow<TerritoryType>(id)` or `ExcelTerritoryHelper.Get(id)` |
| current territory row | `Player.Territory` (`RowRef<TerritoryType>`) |

## Gotchas

- `GenericHelpers.GetRow<T>` returns `T?` and never throws; the static extension `T.Get(id)`,
  `Job.GetData()` and `ExcelActionHelper.GetActionCooldown(id)` use Lumina's throwing `GetRow`.
- `TryGetRow<T>(string sheetName, uint rowId, ClientLanguage? language, out T row)` ignores its
  `language` argument at this commit (it calls `GetSheet<T>(sheetName)`), and so does
  `TryGetRawRow(sheetName, rowId, language, out rawRow)`. Both always read the client language. Use
  `GetSheet<T>(sheetName, language).TryGetRow(rowId, out row)` when the language matters.
- `GetSheet<T>()` defaults to the client language via `Svc.ClientState.ClientLanguage`, whereas the
  static extension `T.Get`/`T.Values` pass `null` straight to `IDataManager`. Both end up at the
  client language; only the mechanism differs.
- `GetPublicWorlds()` with no argument selects the `Region?` overload. Pass a typed argument (`0u`,
  `(uint)dcId` or a `Region`) to make the intended overload explicit.
- `ExcelItemHelper.GetName(..., includeID: true)` returns the name without the id on the first call
  per item (see above). If the id must be present, format it yourself.
- Item and territory-label caches are static and never invalidated; they keep the language of the
  first call.
- `GetCombatJobs()` returns DPS only.
- `ExcelWorldHelper.Get(name, onlyPublic: true)` can return a non-public world if the same name was
  cached earlier by a call without `onlyPublic`.
- `World.IsPublic` (column) and `World.IsPublic()` (ECommons) differ for Korean and Chinese clients.
- `FindRow`, `FindRows`, `TryFindRow`, `FindSubrow`, `GetJobByName`, `GetJobByAbbreviation`,
  `Get(string worldName)` (on a miss), `GetPublicWorlds` and `GetDataCenters` scan whole sheets.
  `GetDataCenters(..., checkForPublicWorlds: true)` scans the `World` sheet once per data centre.
  Call them at load or on demand, not per frame.
- Use `GetText()` rather than `ExtractText()` on sheet strings when you want ECommons' payload
  stripping; the ECommons `ExtractText` overloads are obsolete.
- `ItemUICategoryEnum` needs `using ECommons.ExcelServices.Enums;`; every other enum is in
  `ECommons.ExcelServices`.
- The spelling `ResidentalAreas` is the real class name.

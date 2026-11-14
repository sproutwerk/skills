# Game state helpers: Player, Content, Map, objects, party, gamepad

Covers: the static `Player` class (and the obsolete `LegacyPlayer.Player`), `Content`, `Map`,
`TradeDetectionManager`, everything in `GameFunctions/`, `PartyFunctions/`, `CSExtensions/`,
`Gamepad/GamePad.cs`, and the project's `BannedSymbols.txt`. Verified against ECommons `bc38dc7`
(2026-11-08, v3.2.1.26, Dalamud API 15); nothing here was compiled. Paths are relative to the skill
root.

Not here: `GameHelpers/VfxManager.cs` -> `references/topic-game-hooks.md`. Sheet/ID helpers and the `Job` enum
-> `references/topic-excel.md`. Stock Dalamud services (`IObjectTable`, `IPartyList`, `IPlayerState`,
`ICondition`) -> `dalamud` skill. Struct layouts (`Character`, `BattleChara`, `PlayerState`,
`InventoryManager`) -> `clientstructs` skill.

## Namespaces

```csharp
using ECommons.GameHelpers;                          // Player, Content, ContentType, ContentDifficulty, Map, TradeDetectionManager
using ECommons.GameHelpers.LegacyPlayer;             // obsolete Player, GetJob(this IPlayerCharacter)
using ECommons.GameFunctions;                        // CharacterFunctions, ObjectFunctions, PlayerFunctions, ExtendedPronoun,
                                                     // FakeParty, NameplateKind, RenderDisableManager, TetherInfo, CombatRole
using ECommons.GameFunctions.VirtualTableClassifier; // VTableClassifier, VObjectKind
using ECommons.PartyFunctions;                       // UniversalParty, UniversalPartyMember
using ECommons.CSExtensions;                         // extension properties on ClientStructs structs
using ECommons.Gamepad;                              // GamePad
using ECommons.ExcelServices;                        // Job, GrandCompany, TerritoryIntendedUseEnum (types Player returns)
```

Almost everything reads live game memory or `Svc.Objects` / `Svc.Party` / `Svc.Condition`. Call it
from the framework thread unless a member is listed as a pure pointer read. Several classes use C#
14 `extension(...)` blocks (extension properties); the consuming project needs a language version
that understands them (ECommons itself builds with `LangVersion preview`).

## Player (`ECommons.GameHelpers.Player`)

Source: `references/ecommons/ECommons/GameHelpers/Player.cs`. `public static unsafe class Player`. No
initialisation, no module flag, no hooks, nothing to dispose. All members are static getters
evaluated on every access (no caching).

### The rework and `LegacyPlayer`

The README opens with: "Having errors regarding `Player` class? Add this `using` directive to bypass
it if you don't feel like updating everything just yet":

```csharp
global using Player = ECommons.GameHelpers.LegacyPlayer.Player;
```

The old class still exists as `ECommons.GameHelpers.LegacyPlayer.Player`
(`references/ecommons/ECommons/GameHelpers/LegacyPlayer.cs`) and is marked `[Obsolete("Use
GameHelpers.Player")]`. AutoRetainer at `cc020df` uses exactly that alias
(`AutoRetainer/Global.cs:32`), so **AutoRetainer code shows the legacy member types**; PandorasBox
and YesAlready use the new class. Check which one a project aliases before copying a `Player.X`
expression.

### Members of the new `Player`

| Member | Type | Source / behaviour |
|---|---|---|
| `Object` | `IPlayerCharacter?` | `Svc.Objects.LocalPlayer` |
| `Character` | `Character*` | `Object?.Address ?? 0`; null pointer when not available |
| `BattleChara` | `BattleChara*` | same |
| `GameObject` | `GameObject*` | same |
| `Available` | `bool` | `GameObjectManager.Instance()->Objects.IndexSorted[0].Value != null` (pointer read, does not touch `Svc.Objects`) |
| `Interactable` | `bool` | `Object?.IsTargetable ?? false` |
| `IsBusy` | `bool` | `GenericHelpers.IsOccupied()` or casting or `IsMoving` or `IsAnimationLocked` or `ConditionFlag.InCombat` or `GameMain.Instance()->TerritoryLoadState != 2` |
| `Name` | `string?` | `Object.Name.ToString()`; **no null check, throws `NullReferenceException` when `Object` is null** |
| `NameWithWorld` | `string` | `GetNameWithWorld(Object) ?? string.Empty` |
| `GetNameWithWorld(this IPlayerCharacter? pc)` | `string` | `"Name@HomeWorldName"`; `string.Empty` for null |
| `CID` | `ulong` | `Svc.PlayerState.ContentId` |
| `Status` | `StatusList` | `Object?.StatusList ?? default!` (null when unavailable) |
| `Sex` | `Dalamud.Game.Player.Sex` | `Svc.PlayerState.Sex` |
| `Level` | `int` | `Object?.Level ?? 0`; documented as unsynced level |
| `MaxLevel` | `ECommons.MathHelpers.Number` | `PlayerState.Instance()->MaxLevel` |
| `IsLevelSynced` | `bool` | `PlayerState.Instance()->IsLevelSynced` |
| `SyncedLevel` | `int` | `PlayerState.Instance()->SyncedLevel` |
| `GetLevel(Job job)` | `int` | `Svc.PlayerState.GetClassJobLevel(job.GetGameData().Value)`; unsynced |
| `Race` | `RowRef<Race>` | `Svc.PlayerState.Race` |
| `Tribe` | `RowRef<Tribe>` | from `PlayerState.Instance()->Tribe` |
| `HomeWorld` | `RowRef<World>` | `Object?.HomeWorld ?? default` |
| `CurrentWorld` | `RowRef<World>` | `Object?.CurrentWorld ?? default` |
| `HomeDateCenter` | `RowRef<WorldDCGroupType>` | spelled `Date` in source (typo is the real name) |
| `CurrentDataCenter` | `RowRef<WorldDCGroupType>` | |
| `Territory` | `RowRef<TerritoryType>` | `TerritoryType.GetRef(Svc.ClientState.TerritoryType)` |
| `TerritoryIntendedUse` | `RowRef<TerritoryIntendedUse>` | `Territory.ValueNullable?.TerritoryIntendedUse ?? default` |
| `HomeAetheryteTerritory` | `RowRef<TerritoryType>` | `Aetheryte.GetRef(PlayerState.Instance()->HomeAetheryteId).Value.Territory` (uses `.Value`) |
| `ClassJob` | `RowRef<ClassJob>` | `Object?.ClassJob ?? default` |
| `OnlineStatus` | `RowRef<OnlineStatus>` | `Object?.OnlineStatus ?? default` |
| `ContentFinderCondition` | `RowRef<ContentFinderCondition>` | from `GameMain.Instance()->CurrentContentFinderConditionId` |
| `HomeWorldName`, `CurrentWorldName` | `string` | null when unavailable |
| `HomeDataCenterName`, `CurrentDataCenterName` | `string` | null when unavailable |
| `CsTerritoryIntendedUseEnum` | `FFXIVClientStructs.FFXIV.Client.Enums.TerritoryIntendedUse` | cast of `TerritoryIntendedUse.RowId` |
| `TerritoryIntendedUseEnum` | `ECommons.ExcelServices.TerritoryIntendedUseEnum` | cast of `TerritoryIntendedUse.RowId` |
| `IsInDuty` | `bool` | `CurrentContentFinderConditionId != 0` |
| `IsOnIsland` | `bool` | `MJIManager.Instance()->IsPlayerInSanctuary` |
| `IsInPvP` | `bool` | `GameMain.IsInPvPInstance()` |
| `IsPenalised` | `bool` | `InstanceContent.Instance()->GetPenaltyRemainingInMinutes(0) > 0` |
| `IsInHomeWorld` | `bool` | `Available && CurrentWorld.RowId == HomeWorld.RowId` |
| `IsInHomeDC` | `bool` | `Available &&` data centre row ids equal (uses `.Value`) |
| `Job` | `ECommons.ExcelServices.Job` | `(Job)ClassJob.RowId` |
| `GrandCompany` | `ECommons.ExcelServices.GrandCompany` | `PlayerState.Instance()->GrandCompany` |
| `Position` | `Vector3` | `Object?.Position ?? default` |
| `Rotation` | `float` | `Object?.Rotation ?? default` |
| `IsMoving` | `bool` | `Available && (AgentMap.Instance()->IsPlayerMoving \|\| IsJumping)` |
| `IsJumping` | `bool` | `Jumping` or `Jumping61` condition, or `Character->IsJumping()` |
| `Mounted` | `bool` | `ConditionFlag.Mounted` |
| `Mounting` | `bool` | `ConditionFlag.MountOrOrnamentTransition` |
| `CanMount` | `bool` | territory `Mount == true` and `PlayerState.Instance()->NumOwnedMounts > 0` |
| `CanFly` | `bool` | `Control.CanFly` (requires being mounted, per source comment) |
| `AnimationLock` | `float` | `ActionManager.Instance()->AnimationLock` |
| `IsAnimationLocked` | `bool` | `AnimationLock > 0` |
| `IsCasting` | `bool` | `Object?.IsCasting() ?? false` (the ECommons extension method, see CharacterFunctions) |
| `IsDead` | `bool` | `ConditionFlag.Unconscious` |
| `Revivable` | `bool` | `IsDead && AgentRevive.Instance()->ReviveState != 0` |
| `DistanceTo(Vector3 other)` | `float` | 3D distance from `Position` |
| `DistanceTo(Vector2 other)` | `float` | 2D distance using `Position.ToVector2()` (X/Z) |
| `DistanceTo(IGameObject other)` | `float` | 3D distance |
| extension property `pc.Job` on `IPlayerCharacter` | `Job` | `(Job)(pc?.GetJob() ?? 0)` |

### Exact differences, new vs legacy

| Member | New `ECommons.GameHelpers.Player` | `LegacyPlayer.Player` |
|---|---|---|
| `HomeWorld`, `CurrentWorld` | `RowRef<World>` | `string` (world name) |
| world names | `HomeWorldName`, `CurrentWorldName` | the properties above |
| world ids | `HomeWorld.RowId`, `CurrentWorld.RowId` | `uint HomeWorldId`, `uint CurrentWorldId` |
| `HomeDataCenter` | not present; `HomeDateCenter` is `RowRef<WorldDCGroupType>`, name via `HomeDataCenterName` | `string` |
| `CurrentDataCenter` | `RowRef<WorldDCGroupType>`, name via `CurrentDataCenterName` | `string` |
| `Territory` | `RowRef<TerritoryType>` | `uint` |
| `TerritoryIntendedUse` | `RowRef<TerritoryIntendedUse>`; enum via `TerritoryIntendedUseEnum` / `CsTerritoryIntendedUseEnum` | `TerritoryIntendedUseEnum` |
| `HomeAetheryteTerritory` | `RowRef<TerritoryType>` | `uint` |
| `OnlineStatus` | `RowRef<OnlineStatus>` | `uint` |
| `JobId` | not present; `ClassJob.RowId` or `(uint)Job` | `uint` |
| `MaxLevel` | property, read from `PlayerState` | `static readonly Number MaxLevel = 100` field |
| `Available` | pointer check on sorted object index 0 | `Object != null` |
| `AvailableThreadSafe` | not present (it became `Available`) | the pointer check |
| `IsBusy` | null-safe; also true while `TerritoryLoadState != 2` | `Object.IsCasting` dereferenced without null check; no load-state term |
| `Character`, `BattleChara`, `GameObject` | null pointer when no player | `Object.Address` without null check (throws) |
| `Name` | `Object.Name` (throws when null) | `Object?.Name` (null when null) |
| `NameWithWorld`, `GetNameWithWorld` | `string.Empty` for null | `null` for null |
| level by job | `GetLevel(Job)` | `GetUnsyncedLevel(Job)`; `UnsyncedLevel` is `[Obsolete("Use Level instead")]` |
| job of another character | extension property `pc.Job` | extension method `GetJob(this IPlayerCharacter pc)` |
| new only | `Tribe`, `ContentFinderCondition`, `IsPenalised`, `CsTerritoryIntendedUseEnum` | |
| legacy only | | `Dismounting`, `Jumping`: `[Obsolete("Use IsJumping", true)]` (compile error) |

Unchanged in name and type: `Object`, `Interactable`, `CID`, `Status`, `Race`, `Sex`, `Level`,
`IsLevelSynced`, `SyncedLevel`, `IsInHomeWorld`, `IsInHomeDC`, `ClassJob`, `Job`, `GrandCompany`,
`IsInDuty`, `IsOnIsland`, `IsInPvP`, `Position`, `Rotation`, `IsMoving`, `IsJumping`, `Mounted`,
`Mounting`, `CanMount`, `CanFly`, `AnimationLock`, `IsAnimationLocked`, `IsCasting`, `IsDead`,
`Revivable`, `DistanceTo` (3 overloads).

Porting rule of thumb: legacy `Player.Territory == 123` becomes `Player.Territory.RowId == 123`;
legacy `Player.HomeWorld` (string) becomes `Player.HomeWorldName`; legacy
`Player.TerritoryIntendedUse == TerritoryIntendedUseEnum.Inn` becomes
`Player.TerritoryIntendedUseEnum == TerritoryIntendedUseEnum.Inn`.

### Examples

```csharp
// New Player (PandorasBox Features/Chat/AutoOpenCoords.cs, with
// `using static ECommons.ExcelServices.TerritoryIntendedUseEnum;`)
if (Player.TerritoryIntendedUseEnum is not (City_Area or Open_World or Inn or Starting_Area))
    return;

// New Player (YesAlready UI/Tabs/YesNo.cs): RowRef chain
var zone = Player.Territory.Value.PlaceName.Value.Name.ToString();

// New Player (PandorasBox Features/Actions/AutoMeditation.cs)
var isMonk = Player.Job == Job.MNK;

// Guard pattern used everywhere (AutoRetainer, PandorasBox)
if (!Player.Available) return false;
if (Player.Interactable && !Player.IsBusy) { /* act */ }

// LEGACY alias only (AutoRetainer Scheduler/Tasks/TaskTeleportToProperty.cs:101, Territory is uint):
if (!Inns.List.Contains((ushort)Player.Territory)) { }
// the same with the new class:
if (!Inns.List.Contains((ushort)Player.Territory.RowId)) { }
```

## Content (`ECommons.GameHelpers.Content`)

Source: `references/ecommons/ECommons/GameHelpers/Content.cs`. Classifies the current territory for
PvE combat content. Every property re-reads the sheets on access; sheets are always loaded in
`Language.English`, because difficulty is parsed from the English duty name.

Enums (namespace `ECommons.GameHelpers`):

- `ContentType`: `Unknown, OverWorld, PVP, Dungeon, DeepDungeon, Variant, Criterion, Trial,
  FieldOperations, FieldRaid, ARaid, Raid`. Name clashes with `Lumina.Excel.Sheets.ContentType`.
- `ContentDifficulty`: `Unknown, Normal, Hard, Unreal, FieldRaidsSavage, Extreme, Chaotic,
  Criterion, Savage, CriterionSavage, Ultimate`.

| Member | Type | Notes |
|---|---|---|
| `TerritoryID` | `uint` | `Svc.ClientState.TerritoryType` |
| `TerritoryName` | `string?` | zone name parsed out of `ECommons.TerritoryName.GetTerritoryName`; null when unresolved |
| `TerritoryTypeRow` | `TerritoryType?` | |
| `MapID` | `uint` | `Svc.ClientState.MapId` |
| `MapRow` | `Lumina.Excel.Sheets.Map?` | |
| `TerritoryIntendedUse` | `TerritoryIntendedUseEnum?` | |
| `ContentFinderConditionRow` | `ContentFinderCondition?` | from the territory row |
| `ContentFinderConditionRowId` | `uint?` | |
| `ContentName` | `string?` | CFC name, else `TerritoryName`, else null |
| `AllowUndersized` | `bool?` | |
| `HighEndDuty` | `bool?` | |
| `ContentDifficultyFromName` | `string?` | text inside the last `(...)` of the CFC name; only when the name contains ` (hard`, ` (extreme` or ` (savage` |
| `InstanceContentRow` | `InstanceContent?` | |
| `TimeLimit` | `ushort?` | minutes (`TimeLimitmin`) |
| `ContentTypeRow` | `Lumina.Excel.Sheets.ContentType?` | |
| `ContentTypeRowId` | `uint?` | |
| `ContentTypeName` | `string?` | `"OverWorld"` when the row id is 0 |
| `ContentType` | `ContentType?` | result of the private `DetermineContentType()`; falls back to `OverWorld` |
| `ContentDifficulty` | `ContentDifficulty?` | result of the private `DetermineContentDifficulty()`; falls back to `Normal` |

`ContentType` is decided mainly from `TerritoryIntendedUseEnum`; Bozja raids are detected by
`PlaceName` row ids (3597, 3535, 3682, 5179) and Wolves' Den by territory 250. `ContentDifficulty`
is decided from the CFC row (`ContentType.RowId` 28 = Ultimate, 37 = Chaotic, 30 = Criterion with
`Unknown34` 0/1 for normal/savage, 4 + `HighEndDuty` = Unreal, and so on).

```csharp
if (Content.ContentType == ContentType.Raid && Content.ContentDifficulty == ContentDifficulty.Savage) { }
```

Inside a file that also has `using Lumina.Excel.Sheets;` write `ECommons.GameHelpers.ContentType` to
avoid the ambiguity.

## Map (`ECommons.GameHelpers.Map`)

Source: `references/ecommons/ECommons/GameHelpers/Map.cs`. Pure sheet math, no game memory. Name
clashes with `Lumina.Excel.Sheets.Map`.

```csharp
public static Vector3 PixelCoordsToWorldCoords(int x, int z, uint mapId)   // Y is always 0
public static float   PixelCoordToWorldCoord(float coord, float scale, short offset)
public static uint    FindClosestAetheryte(uint territoryTypeId, Vector3 worldPos) // aetheryte row id, 0 if none
public static Vector3 AetherytePosition(uint aetheryteId)
public static Vector3 AetherytePosition(Aetheryte a)
public static uint    FindPrimaryAetheryte(uint aetheryteId)
```

- `PixelCoordsToWorldCoords` takes map-marker pixel coordinates (as stored in `MapMarker.X/Y`), not
  the in-game "X: 12.3" map coordinates. `mapId` is a `Map` sheet row id; an unknown row falls back
  to scale 100 and offset 0.
- `FindClosestAetheryte` hard-codes territory 886 (the Firmament) to aetheryte 70. It considers
  every `Aetheryte` row whose `Territory` matches, aethernet shards included; pass the result
  through `FindPrimaryAetheryte` to get something teleportable.
- `AetherytePosition` uses `Level[0]` when present, otherwise the `MapMarker` with `DataType == 3`
  (aetheryte) or `DataType == 4` (aethernet). `AetherytePosition(uint)` and `FindPrimaryAetheryte`
  use `GetRow<Aetheryte>(id)!.Value` and throw on an invalid id; the marker fallback uses
  `First(...)` and throws when no marker exists.
- `FindPrimaryAetheryte(0)` returns 0; otherwise the first row in the same `AethernetGroup`.

## TradeDetectionManager (`ECommons.GameHelpers.TradeDetectionManager`)

Source: `references/ecommons/ECommons/GameHelpers/TradeDetectionManager.cs`. Detects player-to-player
trades by watching `ConditionFlag.TradeOpen` and diffing inventory snapshots. No hooks.

```csharp
public delegate void OnTradeStartDelegate(IPlayerCharacter? counterparty);
public delegate void OnTradeEndDelegate(IPlayerCharacter? counterparty, TradeDescriptor? result);
public static event OnTradeStartDelegate OnTradeStart;
public static event OnTradeEndDelegate   OnTradeEnd;

public static IPlayerCharacter GetTradePartner();            // null when not found
public static bool SelfReadyForTrade { get; }
public static bool PartnerReadyForTrade { get; }
public static readonly InventoryType[] ValidInventories;     // Inventory1-4, all Armory*, Crystals
public static Dictionary<uint, uint> GetInventorySnapshot(IEnumerable<InventoryType> validInventories);
public static TradeDescriptor GetTradeResult(ulong cid, Dictionary<uint, uint> invStart, Dictionary<uint, uint> invEnd);

public class TradeDescriptor { public ulong TradePartnerCID; public int ReceivedGil; public ItemWithQuantity[] ReceivedItems = []; }
public class ItemWithQuantity { public uint ItemID; public int Quantity; /* ctor (uint itemID, int quantity) */ }
```

- Lazy: the first subscription to either event attaches `Svc.Condition.ConditionChange`.
  `ECommonsMain.Dispose()` calls the internal `Dispose()`, which detaches it. Your own handlers are
  not cleared; unsubscribing them is optional once the plugin unloads.
- `OnTradeEnd` fires only when a snapshot exists and the partner content id is known. `result` is
  `null` when nothing was received (no positive or negative item delta and gil delta 0). "Received"
  is a signed difference: items or gil you gave away appear with negative `Quantity` /
  `ReceivedGil`.
- Snapshot keys are item ids as returned by `InventoryItem.GetItemId()` (HQ items are `id +
  1000000`); key `1` holds gil. `ItemWithQuantity.ToString()` prints the name with ` HQ` for ids
  above 1000000.
- `GetTradePartner`, `SelfReadyForTrade` and `PartnerReadyForTrade` read hard-coded offsets from
  `InventoryManager.Instance()` (8612, 8572, 8576). They are patch-sensitive; a game update can
  break them silently.
- A cancelled trade also raises `OnTradeEnd` when the conditions above hold; the descriptor is then
  `null` because nothing changed.

```csharp
TradeDetectionManager.OnTradeEnd += (partner, result) =>
{
    if (result != null) PluginLog.Information($"Trade with {partner?.Name}: {result}");
};
```

## GameFunctions

### CharacterFunctions

Source: `references/ecommons/ECommons/GameFunctions/CharacterFunctions.cs`. `public static unsafe
class CharacterFunctions`, written with `extension(...)` blocks.

On `IBattleChara chr`:

| Member | Type | Notes |
|---|---|---|
| `chr.CastInfo` | `FFXIVClientStructs.FFXIV.Client.Game.Character.CastInfo` (copy) | `default` when the game returns no cast info |
| `chr.RemainingCastTime` | `float` | `CastInfo.TotalCastTime - CastInfo.CurrentCastTime` |
| `chr.Struct()` | `BattleChara*` | `(BattleChara*)chr?.Address` on `master`: a null `chr` gives a null pointer. Packages up to 3.2.1.26 dereference `chr` and throw |
| `chr.Character()` | `Character*` | same |
| `chr.GameObject()` | `GameObject*` | same |
| `chr.HasStatus(uint id, float? lessThan = null, float? moreThan = null)` | `bool` | |
| `chr.HasStatus(uint id, out float time, float? lessThan = null, float? moreThan = null)` | `bool` | `time` = remaining time |
| `chr.HasStatus(IEnumerable<uint> id, float? lessThan = null, float? moreThan = null)` | `bool` | any of |
| `chr.HasStatus(IEnumerable<uint> id, out List<(uint ID, float Time)> foundStatus, float? lessThan = null, float? moreThan = null)` | `bool` | collects all matches |
| `chr.IsCasting(uint spellId = 0, ActionType? type = null)` | `bool` | `spellId == 0` matches any cast |
| `chr.IsCasting(params uint[] spellId)` | `bool` | |
| `chr.IsCasting(IEnumerable<uint> spellId)` | `bool` | |

`lessThan` / `moreThan` filter on `RemainingTime`: a status passes when its remaining time is `<=
lessThan` and `>= moreThan`. `IsCasting(...)` returns false when `CastInfo.ActionId == 0` even if
Dalamud's `IBattleChara.IsCasting` property is true. `ActionType` is
`FFXIVClientStructs.FFXIV.Client.Game.ActionType`.

On `IGameObject obj`: `obj.ObjectId` (`uint`, alias of `EntityId`), `obj.Position2` (`Vector2`,
X/Z).

On `ICharacter chr`:

| Member | Type | Notes |
|---|---|---|
| `chr.Health` | `float` | `CurrentHp / MaxHp` as a 0..1 fraction; NaN when `MaxHp` is 0 |
| `chr.MissingHp` | `uint` | |
| `chr.StatusLoop` | `uint` | `StatusLoopVfxId` |
| `chr.ModelId` | `int` | `ModelContainer.ModelCharaId` |
| `chr.Struct()` | `Character*` | null-safe on `master` only, as above |
| `chr.IGameObject()` | `GameObject*` | method named `IGameObject`; null-safe on `master` only |
| `chr.IsCharacterVisible()` | `bool` | bit 0 of `DrawObject + 136`; false when no draw object |
| `chr.GetTransformationID()` | `byte` | `Timeline.ModelState` |
| `chr.IsInWater()` | `bool` | raw offset `Address + 1452` (patch-sensitive) |
| `chr.GetRole()` | `CombatRole` | `ClassJob.Role` 1 = Tank, 2 and 3 = DPS, 4 = Healer, else NonCombat |
| `chr.GetTethers(bool onlySource = false)` | `List<TetherInfo>` | see TetherInfo |

Static: `public static ushort GetVFXId(void* VfxData)` (reads `ushort` at offset 8, 0 for null).

`enum CombatRole { NonCombat, Tank, Healer, DPS }`
(`references/ecommons/ECommons/GameFunctions/CombatRole.cs`).

Because `IBattleChara` derives from `ICharacter`, `chr.Struct()` on an `IBattleChara` returns
`BattleChara*` and on an `ICharacter` returns `Character*`; on an `IGameObject` (ObjectFunctions) it
returns `GameObject*`.

### TetherInfo

Source: `references/ecommons/ECommons/GameFunctions/TetherInfo.cs`.

```csharp
public class TetherInfo
{
    public uint Id { get; }               // RawInfo.Id
    public bool IsSource { get; init; }
    public Tether RawInfo { get; init; }  // FFXIVClientStructs...Character.VfxContainer.Tether
    public uint PairId { get; init; }     // entity id of the other end
    public IGameObject? Pair { get; }     // Svc.Objects lookup by ObjectId on every access
    public TetherInfo();
    public TetherInfo(Tether rawInfo, uint pairId, bool isSource);
}
```

`GetTethers()` returns the character's own tether slots (`IsSource = true`) and, unless `onlySource`
is true, scans every `ICharacter` in `Svc.Objects` for tethers that target this character (`IsSource
= false`). The full scan is an object-table walk per call.

### ObjectFunctions

Source: `references/ecommons/ECommons/GameFunctions/ObjectFunctions.cs`.

```csharp
public static GameObject*  Struct(this IGameObject o);
public static EventObject* Struct(this IEventObj o);
public static GameObject*  GameObject(this IEventObj o);
[Obsolete("Use IGameObject.IsTargetable")] public static bool IsTargetable(this IGameObject o);
public static bool IsHostile(this IGameObject a);
public static NameplateKind GetNameplateKind(this IGameObject o);
public static int  GetAttackableEnemyCountAroundPoint(Vector3 point, float radius);
public static bool TryGetPartyMemberObjectByObjectId(uint objectId, out IGameObject partyMemberObject);
public static bool TryGetPartyMemberObjectByAddress(IntPtr address, out IGameObject partyMemberObject);
```

Module flag. `ECommonsMain.Init` calls `ObjectFunctions.Init()` when the module list contains
`Module.All` or `Module.ObjectFunctions` (`references/ecommons/ECommons/ECommonsMain.cs`). `Init()`
only resolves the native nameplate-colour function by signature through
`EzDelegate.Get<GetNameplateColorDelegate>(sig)`. `IsHostile` and `GetNameplateKind` also do
`GetNameplateColorNative ??= EzDelegate.Get<...>(...)` themselves, so at this commit they work
without the flag; the flag only moves the signature scan to plugin load. Passing
`Module.ObjectFunctions` is still the documented way and costs nothing. It is a delegate to a game
function, not a hook; nothing is disposed.

- `IsHostile`: true for nameplate kinds 4, 5, 6 (PvP enemies), 7, 9, 10, 11. Kind 8 (dead) is not
  hostile. An enemy engaged by another party (10) counts as hostile.
- `GetNameplateKind`: logs at Debug when the object or delegate is null but then calls the delegate
  anyway; pass a non-null object.
- `GetAttackableEnemyCountAroundPoint`: counts targetable hostile battle NPCs with `distance <=
  radius + HitboxRadius`.
- `TryGetPartyMemberObjectBy*`: checks the local player first, then `Svc.Party`. `objectId` is
  compared with `GameObjectId` (a `ulong`), not `EntityId`.

`NameplateKind` (`references/ecommons/ECommons/GameFunctions/NameplateKind.cs`): `PlayerCharacterSelf
= 1, InDutyPartyMember = 2, InDutyPCNotInParty = 3, EnemyMalestromPVPPC = 4, EnemyAdderPVPPC = 5,
EnemyFlamesPVPPC = 6, HostileNotEngaged = 7, Dead = 8, HostileEngagedSelfDamaged = 9,
HostileEngagedOther = 10, HostileEngagedSelfUndamaged = 11, FriendlyBattleNPC = 12,
PlayerCharacterChocobo = 15, OtherPlayerCharacterChocobo = 17, OtherAlliancePlayerCharacter = 20,
AnyPlayerCharacterDead = 21, OutOfDutyandPartyPC = 22, InDutyPCInPartyTank = 27,
InDutyPCInPartyHealer = 28, InDutyPCInPartyDPS = 29`.

```csharp
// PandorasBox Features/Targets/FATETargeting.cs
if (tar == null || tar.IsDead || (tar.Struct()->FateId == 0 && tar.IsHostile())) { }
```

### PlayerFunctions

Source: `references/ecommons/ECommons/GameFunctions/PlayerFunctions.cs`.

```csharp
public static bool TryGetPlaceholder(this IGameObject pc, out int number, bool verbose = false);
public static BattleChara* BattleChara(this IPlayerCharacter o);
public static Character*   Character(this IPlayerCharacter o);
public static GameObject*  IGameObject(this IPlayerCharacter o);
```

`TryGetPlaceholder` resolves `<1>` .. `<8>` through the game's `PronounModule` and returns the party
slot number whose object address equals `pc`. `verbose` logs each slot at Debug.

### ExtendedPronoun

Source: `references/ecommons/ECommons/GameFunctions/ExtendedPronoun.cs`.

```csharp
public static GameObject* ExtendedPronoun.Resolve(string pronoun);
public static void        ExtendedPronoun.Extensions.Add(delegate*<string, string, string, GameObject*> function);
public static void        ExtendedPronoun.Extensions.Remove(delegate*<string, string, string, GameObject*> extension);
public static GameObject* ExtendedPronoun.Extensions.Invoke(string pronoun, string extraCondition, string modifier);
```

`FakePronoun.Resolve(string)` is `[Obsolete("Use ExtendedPronoun instead of FakePronoun")]` and
forwards to `ExtendedPronoun.Resolve`.

- Input includes the brackets, e.g. `"<t1>"`, `"<me>"`, `"<h2::target>"`. The first and last
  character are stripped unconditionally; strings shorter than 2 return null.
- Syntax inside the brackets is `key[:extraCondition[:modifier]]`. The only modifier implemented is
  `target` (returns that object's target). `extraCondition` is parsed but unused by the built-in
  keys.
- Built-in keys: `d1`..`d8` (aliases `m1`, `m2` = d1, d2; `r1`, `r2` = d3, d4), `t1`..`t8` (aliases
  `mt` = t1, `ot` = t2), `h1`..`h8`, `me`, `target`. Role keys pick the n-th `IPlayerCharacter` in
  `Svc.Objects` order whose `GetRole()` matches; this is object-table order, not party-list order.
- Unknown keys fall through to the game's `PronounModule->ResolvePlaceholder`, so stock placeholders
  such as `<2>`, `<focus>`, `<mo>` work.
- During `ConditionFlag.DutyRecorderPlayback` a numeric key (`<3>`) returns the n-th player in the
  object table, and there is no fallback to the game resolver.
- Exceptions are caught, logged and turned into null.
- `Extensions.Invoke` is public but `Resolve` does not call it at this commit; registered extension
  handlers only run if the caller invokes `Extensions.Invoke` itself. Handlers are unmanaged
  function pointers (`&StaticMethod`), not delegates, and are not cleared by
  `ECommonsMain.Dispose()`.

### FakeParty

Source: `references/ecommons/ECommons/GameFunctions/FakeParty.cs`.

```csharp
public static IEnumerable<IPlayerCharacter> FakeParty.Get();   // framework thread only (source comment)
```

In duty-recorder playback: every `IPlayerCharacter` in the object table. Otherwise: members of
`Svc.Party` whose `GameObject` is loaded; when the party list is empty, only the local player, which
is yielded without a null check (a single `null` element when not logged in). Cross-world parties
outside a duty are not covered; use `UniversalParty` for those.

### RenderDisableManager

Source: `references/ecommons/ECommons/GameFunctions/RenderDisableManager.cs`. Cross-plugin
arbitration of the game's "3D rendering disabled" flag.

```csharp
public static bool RenderDisableManager.Debug;      // verbose logging switch
public static void RenderDisableManager.PlaceRequest();
public static void RenderDisableManager.RemoveRequest();
[Obsolete("... Just remove this Init call.", true)] public static void Init();   // compile error
```

- Initialises itself on the first `PlaceRequest()`. State lives in Dalamud data share
  (`ECommons.RenderDisableRequests`, `ECommons.RenderDisableProcessingFramecount`), so every plugin
  that embeds ECommons shares one request set keyed by `ECommonsMain.InstanceUniqueId`.
- Both methods must run on the framework thread; otherwise they log an error and do nothing.
- Requests do not stack: calling `PlaceRequest()` every frame is fine. Rendering stays disabled
  while any plugin holds a request.
- `PlaceRequest()` subscribes a `Svc.Framework.Update` handler that writes
  `Manager.Instance()->Is3DRenderingDisabled`; it unsubscribes itself once the request set is empty.
- `ECommonsMain.Dispose()` calls the internal `Dispose()`: removes this plugin's request, runs one
  update, detaches the handler and relinquishes both data-share entries.

```csharp
// AutoRetainer Modules/Multi/MultiMode.cs:184, inside a framework update handler
if (shouldDisableRender) RenderDisableManager.PlaceRequest();
else                     RenderDisableManager.RemoveRequest();
```

### VirtualTableClassifier

Source: `references/ecommons/ECommons/GameFunctions/VirtualTableClassifier/VTableClassifier.cs`,
`VObjectKind.cs`. Classifies an object by comparing its vtable pointer with ClientStructs'
`StaticVirtualTablePointer`, which is stricter than a C# `is` check on Dalamud wrapper types.

`enum VObjectKind { Null, Unknown, GameObject, Character, BattleChara, EventObject }`

```csharp
public static unsafe VObjectKind Classify(void* gameObject);
public static unsafe VObjectKind Classify(this ref GameObject obj);   // also overloads for ref Character, BattleChara,
    // EventObject, AreaObject, Aetheryte, GatheringPointObject, HousingObject, HousingEventObject,
    // HousingCombinedObject, ReactionEventObject, Treasure
public static unsafe VObjectKind Classify(this IGameObject? obj);
public static bool IsBattleChara(this IGameObject? obj);
public static bool IsBattleChara(this IGameObject? obj, [NotNullWhen(true)] out IBattleChara? chr);
public static bool IsBattleNpc(this IGameObject? obj);
public static bool IsBattleNpc(this IGameObject? obj, [NotNullWhen(true)] out IBattleNpc? chr);
public static IBattleNpc?   AsBattleNpc(this IGameObject? obj);
public static IBattleChara? AsBattleChara(this IGameObject? obj);
public static IEnumerable<IBattleNpc>   OfTypeIBattleNpc<T>(this IEnumerable<T> objects) where T : IGameObject;
public static IEnumerable<IBattleChara> OfTypeIBattleChara<T>(this IEnumerable<T> objects) where T : IGameObject;
```

- `Classify` only ever returns `Null`, `GameObject`, `Character`, `BattleChara` or `Unknown`.
  `VObjectKind.EventObject` exists in the enum but no code path returns it; event objects, treasure,
  housing objects and so on come back as `Unknown`.
- `IsBattleNpc` = vtable is `BattleChara` and `ObjectKind == BattleNpc`.

```csharp
foreach (var npc in Svc.Objects.OfTypeIBattleNpc())
    if (npc.IsHostile() && npc.IsCasting(12345)) { }
```

## PartyFunctions

Source: `references/ecommons/ECommons/PartyFunctions/UniversalParty.cs`, `UniversalPartyMember.cs`.
One list that works for normal parties and cross-world parties.

```csharp
public static unsafe class UniversalParty
{
    public static bool IsCrossWorldParty { get; }   // ConditionFlag.ParticipatingInCrossWorldPartyOrAlliance
    public static bool IsAlliance { get; }          // IsCrossWorldParty && InfoProxyCrossRealm.Instance()->IsInAllianceRaid
    public static int  Length { get; }              // Members.Count
    public static int  LengthPlayback { get; }      // MembersPlayback.Count
    public static List<UniversalPartyMember> Members { get; }
    public static List<UniversalPartyMember> MembersPlayback { get; }
}

public class UniversalPartyMember
{
    public string Name { get; init; }
    public RowRef<World> HomeWorld { get; init; }
    public RowRef<World> CurrentWorld { get; init; }
    public string NameWithWorld { get; }            // "Name@HomeWorld"
    public ulong ContentID { get; init; }
    public Job ClassJob { get; init; }              // ECommons.ExcelServices.Job
    public IGameObject IGameObject { get; }         // may be null
}
```

- `Members` is rebuilt on every access (`Length` rebuilds it too). Read it once into a local.
- Empty list when `Player.Available` is false. The local player is always element 0.
- Cross-world party: members come from `InfoProxyCrossRealm` (all groups, so alliance members are
  included), and `IGameObject` is resolved lazily by name + home world from `Svc.Objects`; it is
  null for members who are not loaded nearby.
- Normal party: members come from `Svc.Party`; `CurrentWorld` is set to the local player's current
  world, `IGameObject` is `IPartyMember.GameObject` (null when out of range or in another zone).
- `MembersPlayback`: in duty-recorder playback returns every named `IPlayerCharacter` in the object
  table, otherwise the same as `Members`.

## CSExtensions

Extension properties on FFXIVClientStructs structs for fields ClientStructs does not name. All are
raw offsets and therefore patch-sensitive. The receiver is `ref`, so they are used on a dereferenced
pointer (`agent->IsFlagMarkerSet`).

| File (`references/ecommons/ECommons/CSExtensions/`) | Receiver | Member |
|---|---|---|
| `AgentMapExtensions.cs` | `ref AgentMap` | `bool IsFlagMarkerSet` (`FlagMarkerCount > 0`), `FlagMapMarker FlagMapMarker` (`FlagMapMarkers[0]`) |
| `AtkEventStateExtensions.cs` | `ref AtkEventState` | `byte UnkFlags3 { get; set; }` (offset 3) |
| `EventObjectExtensions.cs` | `ref EventObject` | `ushort AnimationId { get; set; }` (offset `sizeof(GameObject) + 18`) |
| `EventObjectExtensions.cs` | `IEventObj` | `ushort AnimationId { get; set; }` (forwards to the struct) |
| `InputTimerModuleExtensions.cs` | `ref InputTimerModule` | `float Unk1C { get; set; }` (offset 28) |

```csharp
var agent = AgentMap.Instance();
if (agent->IsFlagMarkerSet) { var flag = agent->FlagMapMarker; }
```

## GamePad (`ECommons.Gamepad.GamePad`)

Source: `references/ecommons/ECommons/Gamepad/GamePad.cs`. Thin wrapper over `Svc.GamepadState` and
`Svc.GameConfig`. The namespace is `ECommons.Gamepad`, the class is `GamePad`.

```csharp
public static Dictionary<GamepadButtons, string> ControllerButtons;   // display names, e.g. L2 -> "L2 / LT"
public static bool IsControllerEnabled();                             // SystemConfigOption.PadAvailable
public static bool IsButtonPressed(GamepadButtons button);            // first frame only
public static bool IsButtonHeld(GamepadButtons button);               // every frame while down (Raw)
public static bool IsButtonJustReleased(GamepadButtons button);       // first frame after release
```

`GamepadButtons` is `Dalamud.Game.ClientState.GamePad.GamepadButtons`. `ControllerButtons` has 17
entries (`None`, 4 d-pad, 4 face, `L1`-`L3`, `R1`-`R3`, `Start`, `Select`) and is a mutable public
field.

```csharp
// PandorasBox Features/Other/TurboController.cs
foreach (var btn in GamePad.ControllerButtons)
    if (GamePad.IsButtonHeld(btn.Key)) { }
```

## BannedSymbols.txt

Source: `references/ecommons/ECommons/BannedSymbols.txt` (upstream
`ECommons/ECommons/BannedSymbols.txt`). `ECommons.csproj` includes it as `<AdditionalFiles>` and
references `Microsoft.CodeAnalysis.BannedApiAnalyzers` with `PrivateAssets=all`.

Scope, stated carefully:

- The list is enforced when the **ECommons project itself** is compiled: for contributors, and for
  the ECommons compilation step of a plugin that builds ECommons from a submodule. It governs code
  inside ECommons.
- It does **not** apply to the consuming plugin's own code. The analyzer reference is private to the
  ECommons project and `AdditionalFiles` is per project. A plugin is only affected if it adds the
  analyzer and its own banned-symbols file.
- NuGet consumers never see it.
- Diagnostics are `RS0030`, a warning by default. ECommons suppresses it locally where needed
  (`#pragma warning disable RS0030` in `ExcelWorldHelper.IsPublic`). The file's own message invites
  removing an entry "if there is no other sane way to obtain desired data".

It is still a useful statement of what the maintainer considers unreliable, and what to use instead:

| Banned | Message in file | ECommons replacement |
|---|---|---|
| `T:Dalamud.Interface.Utility.Raii.ImRaii` | "Use native ImGui methods for compatibility" | plain `ImGui.Begin*/End*`, `Push*/Pop*` calls inside ECommons; `ImRaii` itself -> `dalamudimgui` skill |
| `M:Lumina.Text.ReadOnly.ReadOnlySeString.ExtractText` | "Use GetText extension method for consistency" | `GetText(this ReadOnlySeString s, bool onlyFirst = false)` (`references/topic-excel.md`) |
| `P:Lumina.Excel.Sheets.World.Region` | "Unreliable - use GetRegion() extension method" | `ExcelWorldHelper.GetRegion(this World world)` |
| `IPlayerState` members (38 entries) | "IPlayerState usage discouraged because of poor implementation." | `Player` members, see below |
| `IBattleChara.IsCastInterruptible`, `CastActionType`, `CastActionId`, `CastTargetObjectId`, `CurrentCastTime`, `BaseCastTime`, `TotalCastTime` | none | `chr.CastInfo` (ClientStructs `CastInfo` copy), `chr.RemainingCastTime`, `chr.IsCasting(...)` |

Banned `IPlayerState` members: properties `CharacterName`, `EntityId`, `CurrentWorld`, `HomeWorld`,
`Tribe`, `ClassJob`, `Level`, `IsLevelSynced`, `EffectiveLevel`, `GuardianDeity`, `BirthMonth`,
`BirthDay`, `FirstClass`, `StartTown`, `BaseStrength`, `BaseDexterity`, `BaseVitality`,
`BaseIntelligence`, `BaseMind`, `BasePiety`, `GrandCompany`, `HomeAetheryte`, `FavoriteAetherytes`,
`FreeAetheryte`, `BaseRestedExperience`, `PlayerCommendations`, `DeliveryLevel`, `MentorVersion`,
`IsMentor`, `IsBattleMentor`, `IsTradeMentor`, `IsNovice`, `IsReturner`; methods `GetAttribute`,
`GetGrandCompanyRank`, `GetClassJobLevel`, `GetClassJobExperience`, `GetDesynthesisLevel`. Not
banned and used by `Player`: `ContentId`, `Sex`, `Race`.

| `IPlayerState` member | `Player` equivalent |
|---|---|
| `CharacterName` | `Player.Name` |
| `EntityId` | `Player.Object?.EntityId` |
| `CurrentWorld`, `HomeWorld` | `Player.CurrentWorld`, `Player.HomeWorld` |
| `Tribe` | `Player.Tribe` |
| `ClassJob` | `Player.ClassJob`, `Player.Job` |
| `Level` | `Player.Level` |
| `IsLevelSynced` | `Player.IsLevelSynced` |
| `EffectiveLevel` | `Player.SyncedLevel` (reads `PlayerState.Instance()->SyncedLevel`) |
| `GrandCompany` | `Player.GrandCompany` |
| `HomeAetheryte` | `Player.HomeAetheryteTerritory` (territory only) |
| `GetClassJobLevel` | `Player.GetLevel(Job)` (which itself still calls `Svc.PlayerState.GetClassJobLevel`) |
| everything else | no ECommons wrapper; read `FFXIVClientStructs...UI.PlayerState.Instance()` directly |

## Gotchas

- Two classes are named `Player`. With both `using ECommons.GameHelpers;` and `using
  ECommons.GameHelpers.LegacyPlayer;` in scope, `Player` is ambiguous unless a `using` alias picks
  one. Both classes also declare `GetNameWithWorld(this IPlayerCharacter)`, so importing both
  namespaces makes `pc.GetNameWithWorld()` ambiguous as well.
- `GetJob(this IPlayerCharacter pc)` exists only on the legacy class. With only the new namespace
  imported use the extension property `pc.Job`.
- New `Player.Name` dereferences `Object` without a null check, despite the class comment about not
  throwing. Guard with `Player.Available` or use `Player.Object?.Name.ToString()`.
- New `Player.Available` is a raw pointer check, not `Object != null`. Do not treat it as proof that
  `Player.Object` is non-null: members built on `Object?.` are safe either way, `Name` and anything
  you dereference yourself (`Player.Object.X`) are not.
- `Player.Character`, `Player.BattleChara`, `Player.GameObject` return null pointers when there is
  no player. Check before `->`.
- `Player.HomeAetheryteTerritory` and `Player.IsInHomeDC` use `RowRef.Value`, which throws on an
  invalid row.
- `Player.Level` is the unsynced level. Use `Player.SyncedLevel` inside synced duties.
- `Player.HomeDateCenter` is the real spelling.
- `Player.Territory` compared with a number must go through `.RowId` on the new class.
- `ContentType`, `Map` and `GrandCompany` clash with Lumina sheet names; alias or qualify them.
- `Content` parses English duty names; do not pass its string outputs through localisation logic.
- `TradeDetectionManager`, `CharacterFunctions.IsInWater` and the `CSExtensions` members read
  hard-coded offsets. After a game patch verify them before trusting the result.
- `IsHostile()` is true for enemies engaged with another party (kind 10). Filter with
  `GetNameplateKind()` when that matters.
- `VTableClassifier` never returns `VObjectKind.EventObject`.
- `ExtendedPronoun` role placeholders follow object-table order, which is not guaranteed to match
  the party list order shown in game.
- `FakeParty.Get()` can yield a single null when not logged in; `UniversalParty.Members` returns an
  empty list in that case.
- `UniversalPartyMember.IGameObject` is frequently null (member in another zone, or cross-world
  member not loaded). Never dereference it unchecked.
- `RenderDisableManager` calls off the framework thread are silently ignored apart from an error
  log.

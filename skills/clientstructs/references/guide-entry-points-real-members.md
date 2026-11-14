## Entry points (real members)

Everything is `unsafe`, pointer-returning, and may be `null` when the game
state is not ready (login screen, zoning). Check every pointer.

```csharp
using FFXIVClientStructs.FFXIV.Client.System.Framework;
using FFXIVClientStructs.FFXIV.Client.UI;
using FFXIVClientStructs.FFXIV.Client.UI.Agent;
using FFXIVClientStructs.FFXIV.Client.Game;
using FFXIVClientStructs.FFXIV.Client.Game.UI;
using FFXIVClientStructs.FFXIV.Client.Game.Object;
using FFXIVClientStructs.FFXIV.Client.Game.Character;
using FFXIVClientStructs.FFXIV.Client.Game.Control;
using FFXIVClientStructs.FFXIV.Client.Game.Event;
using FFXIVClientStructs.FFXIV.Component.GUI;

var fw  = Framework.Instance();                 // Client.System.Framework; ->UIModule, ->ClientLanguage, ->FrameCounter, GetServerTime() (static), GameVersionString
var ui  = UIModule.Instance();                  // = fw->GetUIModule(); ->GetRaptureAtkModule(), GetAgentModule(), GetRaptureShellModule(), GetRaptureGearsetModule(), GetRaptureHotbarModule(), GetRaptureLogModule(), GetInfoModule()
var atk = RaptureAtkModule.Instance();          // ->RaptureAtkUnitManager, ->AgentModule, ->AtkArrayDataHolder (inherited from AtkModule)
var uim = RaptureAtkUnitManager.Instance();     // GetAddonByName("Name", index=1), GetAddonById(ushort), AllLoadedUnitsList, FocusedAddon
var am  = AgentModule.Instance();               // GetAgentByInternalId(AgentId), generated GetAgentMap() etc.
var stg = AtkStage.Instance();                  // ->RaptureAtkUnitManager, ->AtkArrayDataHolder, GetNumberArrayData(NumberArrayType), GetStringArrayData(StringArrayType), GetFocus()
```

### Player / world state

- `PlayerState.Instance()` (`Client.Game.UI`, static object): `IsLoaded`,
  `CharacterNameString`, `EntityId`, `ContentId`, `CurrentClassJobId`,
  `CurrentLevel`, `SyncedLevel`, `IsLevelSynced`, `ClassJobLevels` (Span<short>,
  35), `ClassJobExperience`, `GrandCompany`, `HomeAetheryteId`,
  `FavouriteAetherytes`, `Attributes` (Span<int>, 74; `GetAttributeByIndex`),
  `PlayerCommendations`, `CanFly`, `IsMountUnlocked(uint)`,
  `IsOrnamentUnlocked(uint)`, `IsFishCaught(uint)`, the unlock bitarrays
  (`UnlockedMountsBitArray`, `UnlockedAetherCurrentsBitArray`, …),
  `TrackedStatuses` (StdMap<uint,bool>).
- `UIState.Instance()`: `->PlayerState`, `->Hotbar`, `->Buddy`, `->Map`,
  `->InstanceContent`, `IsUnlockLinkUnlocked(uint)`,
  `IsUnlockLinkUnlockedOrQuestCompleted(uint)`, `IsAetheryteUnlocked(uint)`,
  static `IsInstanceContentCompleted/Unlocked(uint)`.
- `QuestManager.Instance()`: static `IsQuestComplete(ushort)`,
  static `GetQuestSequence(ushort)`, `IsQuestAccepted(ushort)`,
  `NormalQuests` (Span<QuestWork>, 30), `DailyQuests`, `LeveQuests`,
  `NumLeveAllowances`, static `IsRecipeComplete(uint)`, `GetBeastTribeAllowance()`.
- `Conditions.Instance()` (`Client.Game`, C++ `Client::Game::Condition`): one
  `bool` per condition flag at its index — `InCombat`, `Casting`, `Mounted`,
  `BoundByDuty`, `BetweenAreas`, `OccupiedInEvent`, `OccupiedInQuestEvent`,
  `WatchingCutscene`, `Jumping`, `InFlight`, `Diving`, … plus `Flags`
  (Span<bool>, 112) indexable by Dalamud's `ConditionFlag`.
- `TerritoryInfo.Instance()`: `InSanctuary`, `FlyingDisabled`,
  `AreaPlaceNameId`, `SubAreaPlaceNameId`, `MapIdOverride`.
- `Telepo.Instance()`: `UpdateAetheryteList()` then `TeleportList`
  (`StdVector<TeleportInfo>`: `AetheryteId`, `GilCost`, `TerritoryId`, `Ward`,
  `Plot`, `SubIndex`, `IsFavourite`, `IsFreeAetheryte`, `IsApartment`),
  `Teleport(aetheryteId, subIndex)`, `ActiveTeleportRequest`.
- `HousingManager.Instance()`: `IsInside()`, `IsOutside()`, `IsInWorkshop()`,
  `GetCurrentWard()`, `GetCurrentPlot()`, `GetCurrentDivision()`,
  `GetCurrentHouseId()`, `HasHousePermissions()`, `->IndoorTerritory`,
  `->OutdoorTerritory`.
- `Control.Instance()`: `->LocalPlayer` (`BattleChara*`), `->TargetSystem`,
  `->CameraManager`; static `Control.GetLocalPlayer()`, `Control.CanFly`.

### Objects and characters

```csharp
var om = GameObjectManager.Instance();
foreach (var p in om->Objects.IndexSorted) {          // Span<Pointer<GameObject>>, 819 slots, nulls allowed
    var obj = p.Value; if (obj == null) continue;
    obj->EntityId; obj->BaseId; obj->ObjectKind; obj->Position; obj->NameString; obj->GetGameObjectId();
    if (obj->IsCharacter()) { var c = (Character*)obj; c->Health; c->MaxHealth; c->ClassJob; c->Level; c->TargetId; c->Mode; c->HomeWorld; c->GetStatusManager(); }
}
om->Objects.GetObjectByEntityId(uint); om->Objects.GetObjectByGameObjectId(GameObjectId);
```

`Objects.IndexSorted` layout: 0–199 CharacterManager (100 BattleCharas at
2i, dependents at 2i+1), 200–448 ClientObjectManager, 449–488 EventObjects,
489–628 StandObjects, … `EntityIdSorted`/`GameObjectIdSorted` are the sorted
views with `*SortedCount`.

- `GameObject` (0x1A0): `Name`(64), `EntityId`, `LayoutId`, `BaseId`,
  `OwnerId`, `ObjectIndex`, `ObjectKind` (enum `ObjectKind`), `SubKind`,
  `TargetableStatus`, `Position`, `Rotation`, `HitboxRadius`, `EventId`,
  `FateId`, `DrawObject`, `EventHandler`; `GetIsTargetable()`, `IsDead()`,
  `GetName()` (CStringPointer), `GetHeight()`, `GetNameId()`.
- `CharacterData` (inherited into `Character`): `Health`, `MaxHealth`, `Mana`,
  `MaxMana`, `GatheringPoints`, `CraftingPoints`, `TitleId`, `ClassJob`,
  `Level`, `OnlineStatus`, `ShieldValue`, `Battalion`, `Flags`. Access
  `chara->Health` or `chara->CharacterData.Health`.
- `Character` (0x2370): `MoveController`, `Mount`, `CompanionData`, `DrawData`,
  `Timeline`, `Vfx`, `Effects`, `ChildObject`, `FreeCompanyTagString`,
  `TargetId`, `SoftTargetId`, `NameId`, `AccountId`, `ContentId`,
  `CurrentWorld`, `HomeWorld`, `Mode` (`CharacterModes`), `ModeParam`;
  `GetStatusManager()` (vf), `GetCastInfo()`, `HasStatus(uint)`,
  `IsInPvP()`, `IsJumping()`, `GetParentCharacter()`.
- `BattleChara` (0x3810, `Inherits<Character>`): `StatusManager`, `CastInfo`,
  `ActionEffectHandler`, `ForayInfo`, `RidePillion(seat)`.
- `StatusManager` (0x3E0): `Status` (Span<Status>, 60: `StatusId`, `Param`,
  `RemainingTime`, `SourceObject`), `NumValidStatuses`, `HasStatus(id,
  sourceId=0xE0000000)`, `GetStatusIndex`, `GetRemainingTime(idx)`,
  static `ExecuteStatusOff(id)` (= right-click remove buff).
- `TargetSystem.Instance()`: `Target`, `SoftTarget`, `FocusTarget`,
  `MouseOverTarget`, `GPoseTarget`; `GetTargetObject()`, `SetHardTarget(obj)`,
  `SetSoftTarget`, `InteractWithObject(obj, checkLos=true)`,
  `IsObjectInViewRange`, `IsObjectOnScreen`, `TargetableObjectsOnScreen`.

### Actions

`ActionManager.Instance()`:
```csharp
am->UseAction(ActionType.Action, actionId, targetId /*ulong, default 0xE000_0000*/);
am->UseActionLocation(ActionType.Action, id, target, &pos);
uint status = am->GetActionStatus(ActionType.Action, id);         // 0 = usable; else LogMessage row id
am->GetAdjustedActionId(id);                                     // combo/upgrade replacement
am->GetRecastTime(type, id); am->GetRecastTimeElapsed(type, id); am->IsRecastTimerActive(type, id); am->IsActionOffCooldown(type, id);
am->GetRecastGroup((int)type, id); am->GetRecastGroupDetail(group);   // RecastDetail: IsActive, ActionId, Elapsed, Total
am->GetCurrentCharges(id); ActionManager.GetMaxCharges(id, 0); ActionManager.GetAdjustedRecastTime(type, id); ActionManager.GetActionRange(id);
ActionManager.CanUseActionOnTarget(id, target); am->IsActionTargetInRange(type, id);
```
`ActionType` enum (`uint`, sequential): `None=0, Action=1, Item=2, EventItem=3, EventAction=4, GeneralAction=5, BuddyAction=6, MainCommand=7, Companion=8, CraftAction=9, Unk10, PetAction=11, Unk12, Mount=13, PvPAction=14, FieldMarker=15, ChocoboRaceAbility=16, ChocoboRaceItem=17, Unk18, BgcArmyAction=19, Ornament=20`. `RecastDetail`: `IsActive`, `ActionId`, `Elapsed`, `Total`.

### Inventory

```csharp
var inv = InventoryManager.Instance();
int n = inv->GetInventoryItemCount(itemId, isHq: false, checkEquipped: true, checkArmory: true, minCollectability: 0);
int m = inv->GetItemCountInContainer(itemId, InventoryType.Inventory1);
uint gil = inv->GetGil();  inv->GetTomestoneCount(itemId);  inv->GetEmptySlotsInBag();
InventoryContainer* c = inv->GetInventoryContainer(InventoryType.Inventory1);   // ->Items (InventoryItem*), ->Size, ->IsLoaded, GetInventorySlot(i)
for (int i = 0; i < c->Size; i++) { var it = c->GetInventorySlot(i); if (it->ItemId == 0) continue; it->Quantity; it->IsHighQuality(); it->GetItemId(); it->Container; it->Slot; it->Flags; it->SpiritbondOrCollectability; it->Condition; it->GlamourId; it->GetMateriaId(0); }
inv->MoveItemSlot(srcType, srcSlot, dstType, dstSlot); inv->DiscardItem(type, slot); inv->SplitItem(type, slot, qty);
```
`InventoryType`: `Inventory1..4 = 0..3`, `EquippedItems = 1000`, `Currency =
2000`, `Crystals = 2001`, `KeyItems = 2004`, `ArmoryMainHand = 3500`,
`ArmoryOffHand = 3200 … ArmoryRings = 3300`, `SaddleBag1/2 = 4000/4001`,
`PremiumSaddleBag1/2`, `RetainerPage1..7 = 10000..`, `RetainerMarket = 12002`,
`FreeCompanyPage1..5 = 20000..`, `Invalid = 9999`.

`AgentInventoryContext.Instance()->UseItem(itemId)` uses an item;
`->DiscardItem(item, type, slot, addonId)`.

### Events / duties

- `EventFramework.Instance()`: `GetContentDirector()`,
  `GetInstanceContentDirector()`, `GetPublicContentDirector()`,
  `GetEventHandlerById(uint)`, `->EventHandlerModule`, `->DirectorModule`;
  static `GetCurrentContentId()`, `GetCurrentContentType()`,
  `CanLeaveCurrentContent()`, `LeaveCurrentContent(forced)`.
- `Director` (`Client.Game.Event`, `Inherits<LuaEventHandler>`): `ContentId`,
  `Sequence`, `Title`, `Objective` (Utf8String), `DirectorTodos`,
  `EventHandlerInfo`. `ContentDirector : Director`;
  `InstanceContentDirector : ContentDirector` adds the InstanceContent sheet
  row (`InstanceContentType`, `ContentTimeMax`, `BossExp0..4`, …).
  `InstanceContentType` enum: `Raid=1, Dungeon=2, GuildOrder=3, Trial=4,
  DeepDungeon=9, VariantDungeon=19, CriterionDungeon=20, …`.

### UI modules

- `RaptureShellModule.Instance()`: `ExecuteMacro(RaptureMacroModule.Macro*)`,
  `CancelMacro()`, `ChangeChatChannel`, `MacroCurrentLine`, `MacroLocked`.
  To run a chat line/command: `UIModule.Instance()->ProcessChatBoxEntry(Utf8String*
  message, nint a4 = 0, bool saveToHistory = false)` (this is what ECommons
  `Chat.SendMessage` calls) or `ShellCommandModule.ExecuteCommandInner(Utf8String*,
  UIModule*)`.
- `RaptureGearsetModule.Instance()`: `Entries` (Span<GearsetEntry>, 100:
  `Id`, `NameString`, `ClassJob`, `ItemLevel`, `Flags`, `Items`),
  `CurrentGearsetIndex`, `NumGearsets`, `GetGearset(id)`, `IsValidGearset(id)`,
  `EquipGearset(id, glamourPlateId = 0)`, `FindGearsetIdByName(Utf8String*)`,
  `HasLinkedGlamourPlate(id)`.
- `AgentMap.Instance()`: `CurrentTerritoryId`, `CurrentMapId`,
  `SelectedMapId`, `SetFlagMapMarker(territoryId, mapId, x, y, iconId=0xEC91)`,
  `OpenMapByMapId(mapId)`, `OpenMap(OpenMapInfo*)`, `FlagMarkerCount` (> 0
  = a flag is set) with `FlagMapMarkers[0]` (`TerritoryId`, `MapId`,
  `XFloat`, `YFloat`), `MapMarkers` (132), `MiniMapMarkers` (100),
  `EventMarkers`.
- `NumberArrayData` / `StringArrayData` (`Component.GUI`): the shared
  HUD/addon backing arrays. `AtkStage.Instance()->GetNumberArrayData(NumberArrayType.Hud)`
  → `->IntArray` / `->Span` / `SetValue(i, v)`;
  `GetStringArrayData(StringArrayType.Hud)` → `->StringArray`
  (`CStringPointer*`) / `SetValue(i, str)`. Enum values in `AtkArrayData.cs`.
  `AtkArrayDataHolder` (`atk->AtkArrayDataHolder`) has the raw
  `NumberArrays`/`StringArrays` + counts.


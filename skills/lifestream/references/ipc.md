# IPC contracts

The [provider](source/Lifestream/Lifestream/IPC/IPCProvider.cs) registers
every public `[EzIPC]` method under `Lifestream.<MethodName>`. `void` methods
are actions (bind with an `object` return placeholder); everything else is a
function whose last generic argument is the return type. The single
`[EzIPCEvent]` is `Lifestream.OnHouseEnterError`, an `Action` with no
arguments. Provider methods are not wrapped: an exception thrown inside one
propagates to the caller.

Throughout, `P` is the plugin instance, `S` the service container,
`C` the configuration, `IsBusy()` the provider's own check
(`P.TaskManager.IsBusy || (P.followPath != null && P.FollowPath.Waypoints.Count > 0)`),
and `Utils.IsBusy()` the same test from
[Utils.cs](source/Lifestream/Lifestream/Utils.cs). Several gates instead test
only `P.TaskManager.IsBusy`, which ignores a running path; the tables say
which.

## State and control

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `Instance` | `IDalamudPlugin Instance()` | none | Returns the plugin object itself. Only useful through reflection; not a serializable contract. |
| `IsBusy` | `bool IsBusy()` | n/a | Task queue busy or waypoints pending. |
| `Abort` | `void Abort()` | none | `TaskManager.Abort()` and `followPath?.Stop()`. Clears everything, including tasks another consumer queued. |
| `GetRealTerritoryType` | `uint GetRealTerritoryType()` | none | `P.Territory`, from [TerritoryWatcher](source/Lifestream/Lifestream/Services/TerritoryWatcher.cs): inside a private house it maps the generic interior territory to the district-specific one using the last outdoor residential territory seen. Elsewhere it equals `Svc.ClientState.TerritoryType`. |
| `ExecuteCommand` | `void ExecuteCommand(string arguments)` | inside parser | `P.ProcessCommand("/li", arguments)`. See [commands](commands.md). |

## Aetheryte teleport

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `Teleport` | `bool Teleport(uint destination, byte subIndex)` | none | `S.TeleportService.TeleportToAetheryte(destination, subIndex)` with `wait: false`. [TeleportService](source/Lifestream/Lifestream/Services/TeleportService.cs) first runs `CanTeleport`: player interactable, `ActionManager.GetActionStatus(ActionType.Action, 5) == 0` (Teleport usable), not animation locked. It then scans `Svc.AetheryteList` for an entry with that `AetheryteId` and `SubIndex` and calls `Telepo.Instance()->Teleport(id, sub)`. Returns `false` when the action is unusable or the aetheryte is not in the unlocked list. Nothing is queued, so `IsBusy()` stays false; wait on `BetweenAreas` yourself. |
| `GetActiveAetheryte` | `uint GetActiveAetheryte()` | none | `P.ActiveAetheryte.Value.ID` or `0`. Set each frame from the aetheryte or shard within range (see [travel](travel.md#how-lifestream-identifies-the-aetheryte-you-stand-at)). |
| `GetActiveCustomAetheryte` | `uint GetActiveCustomAetheryte()` | none | Same for the custom aethernet systems (field zones), else `0`. |
| `GetActiveResidentialAetheryte` | `uint GetActiveResidentialAetheryte()` | none | Same for housing-district shards, else `0`. |

## Aethernet

All of these end in `TaskTryTpToAethernetDestination.Enqueue(name)` with
`allowPartial = false` and `allowTpFallback = false`, so the character must
already be at an aetheryte or shard, or within 30 m of a master aetheryte
whose group contains the destination. Otherwise the task logs
"Destination could not be found (3)" and aborts.

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `AethernetTeleport` | `bool AethernetTeleport(string destination)` | `IsBusy()` | Enqueues the name match. Matching is case-insensitive `Contains` against the aethernet name or the user's `C.Renames`. Returns `true` on enqueue, not on arrival. |
| `AethernetTeleportById` | `bool AethernetTeleportById(uint aethernetSheetRowId)` | via `AethernetTeleport` | `Utils.GetAethernetNameWithOverrides(id)`: the plugin's own name table (which adds "Firmament" under `uint.MaxValue`), else the `Aetheryte` sheet row's `AethernetName`. `false` when no name resolves. |
| `AethernetTeleportByPlaceNameId` | `bool AethernetTeleportByPlaceNameId(uint placeNameRowId)` | via `AethernetTeleport` | `PlaceName` sheet row name. |
| `HousingAethernetTeleportById` | `bool HousingAethernetTeleportById(uint housingAethernetSheetRow)` | via `AethernetTeleport` | `HousingAethernet` sheet row's `PlaceName`. Must stand at a housing shard. |
| `AethernetTeleportToFirmament` | `bool AethernetTeleportToFirmament()` | via `AethernetTeleport` | Name for `TaskAetheryteAethernetTeleport.FirmamentAethernetId` (`uint.MaxValue`). Must be within range of the Foundation aetheryte (ID 70) and `C.Firmament` must be enabled. |

## World and data center

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `CanVisitSameDC` | `bool CanVisitSameDC(string world)` | none | `S.Data.DataStore.Worlds.Contains(world)`: public worlds on the current data center, built on login. Exact, case-sensitive name. |
| `CanVisitCrossDC` | `bool CanVisitCrossDC(string world)` | none | `DataStore.DCWorlds`: public worlds on other data centers in the home region, plus region 4. |
| `ChangeWorld` | `bool ChangeWorld(string world)` | `IsBusy()` | Cross-DC if `CanVisitCrossDC`, else same-DC if `CanVisitSameDC`, via `P.TPAndChangeWorld(world, isDcTransfer)`. Returns `true` as soon as the call is made; the inner method can still refuse (see [travel](travel.md#tpandchangeworld)). |
| `ChangeWorldById` | `bool ChangeWorldById(uint worldId)` | via `ChangeWorld` | Looks up the `World` sheet name, then `ChangeWorld`. |
| `TPAndChangeWorld` | `void TPAndChangeWorld(string w, bool isDcTransfer, string secondaryTeleport, bool noSecondaryTeleport, int? gateway, bool? doNotify, bool? returnToGateway)` | inside (`TaskManager.IsBusy` only) | Direct call to `P.TPAndChangeWorld` with `gateway` cast to `WorldChangeAetheryte?`. `secondaryTeleport` is an aethernet name run on arrival; `noSecondaryTeleport` suppresses both it and the configured default. Null `gateway`/`doNotify`/`returnToGateway` take configuration defaults. |
| `GetWorldChangeAetheryteByTerritoryType` | `int? GetWorldChangeAetheryteByTerritoryType(uint territoryType)` | none | Maps 130 to `Uldah` (9), 129 to `Limsa` (8), 132 to `Gridania` (2). The implementation casts `(int)Utils.GetWorldChangeAetheryteByTerritoryType(territoryType)`; for any other territory the nullable enum is null and the cast throws `InvalidOperationException` instead of returning `null`. |

`WorldChangeAetheryte` values are aetheryte row IDs: `Uldah = 9`,
`Gridania = 2`, `Limsa = 8`
([enum](source/Lifestream/Lifestream/Enums/WorldChangeAetheryte.cs)).

## Housing, inn, and property shortcuts

`TaskPropertyShortcut.PropertyType` is `Auto, Home, FC, Apartment, Inn,
Shared_Estate`
([source](source/Lifestream/Lifestream/Tasks/Shortcuts/TaskPropertyShortcut.cs)).
`HouseEnterMode` is `None, Walk_to_door, Enter_house, Enter_workshop`.
`ResidentialAetheryteKind` is `Uldah = 9, Gridania = 2, Limsa = 8,
Foundation = 70, Kugane = 111`.

`TaskPropertyShortcut.Enqueue(...)` refuses with a chat error when
`P.TaskManager.IsBusy`, returns silently when no player is available, and,
unless `useSameWorld`, first returns to the home world through
`TPAndChangeWorld`. It then teleports with `ExecuteTPToAethernetDestination`,
waits for the screen fade and interactability (15 minute limit), and when a
`HousePathData` with a path exists and the enter mode asks for it, validates
that the player stands within 10 m of the recorded plot entrance. A failed
validation invokes `OnHouseEnterError` and throws.

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `TeleportToFC` / `TeleportToHome` / `TeleportToApartment` | `bool ()` | `P.TaskManager.IsBusy` | `TaskPropertyShortcut.Enqueue(FC / Home / Apartment)`. `false` only for the busy case. |
| `EnqueuePropertyShortcut` | `void EnqueuePropertyShortcut(TaskPropertyShortcut.PropertyType type, HouseEnterMode? mode)` | inside only | Same task with an explicit enter mode. Enum types belong to Lifestream. |
| `EnterApartment` | `void EnterApartment(bool enter)` | inside only | Apartment shortcut with `enterApartment` set. |
| `EnqueueInnShortcut` | `void EnqueueInnShortcut(int? innIndex)` | inside only | Inn shortcut; returns to the home world first. `innIndex` is zero-based into `TaskPropertyShortcut.InnData` (sorted by territory), `null` autodetects. |
| `EnqueueLocalInnShortcut` | `void EnqueueLocalInnShortcut(int? innIndex)` | inside only | Same with `useSameWorld: true`. |
| `HasApartment` / `HasPrivateHouse` / `HasSharedEstate` / `HasFreeCompanyHouse` | `bool? ()` | none | `null` when not on the home world (reads `Player.Object`, so a missing player throws). Otherwise scans `Svc.AetheryteList` for the housing aetheryte IDs (59, 60, 61, 97, 165 for estates; 56, 57, 58, 96, 164 for FC) with the matching `IsApartment`/`IsSharedHouse` flags. |
| `GetCurrentPlotInfo` | `(ResidentialAetheryteKind Kind, int Ward, int Plot)? GetCurrentPlotInfo()` | none | `UIHouseReg.TryGetCurrentPlotInfo`, else `null`. |
| `GetPlotEntrance` | `Vector3? GetPlotEntrance(uint territory, int plot)` | none | Last point of the generated housing path for that residential territory and zero-based plot, from `S.Data.ResidentialAethernet.HousingData`. `null` when no path exists. |
| `GetResidentialTerritory` | `uint GetResidentialTerritory(ResidentialAetheryteKind r)` | none | Residential district territory for the kind (Mist, Goblet, Lavender Beds, Shirogane, Empyreum). |
| `GetHousePathData` | `(HousePathData Private, HousePathData FC) GetHousePathData(ulong CID)` | none | The user's recorded private and FC paths for that content ID, or `null` entries. Empty when `P.DisableHousePathData`. |
| `GetSharedHousePathData` | `HousePathData GetSharedHousePathData()` | none | Path data for the preferred or first shared estate in `Svc.AetheryteList`, else `null`. |
| `CanMoveToWorkshop` | `bool CanMoveToWorkshop()` | none | FC path data exists with a workshop path and the current plot info matches it. |
| `MoveToWorkshop` | `void MoveToWorkshop()` | `IsBusy()` | `P.FollowPath.Move(data.PathToWorkshop, true)` when the same conditions hold. Silent otherwise. |

## Address book

`AddressBookEntryTuple` is a global alias for
`(string Name, int World, int City, int Ward, int PropertyType, int Plot, int Apartment, bool ApartmentSubdivision, bool AliasEnabled, string Alias)`
([source](source/Lifestream/Lifestream/Data/AddressBookEntry.cs)). `City` is
a `ResidentialAetheryteKind` value and `PropertyType` is `Lifestream.Enums.PropertyType`
(`House = 0, Apartment = 1`). Ward, plot and apartment are one-based.

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `BuildAddressBookEntry` | `AddressBookEntryTuple BuildAddressBookEntry(string worldStr, string cityStr, string wardNum, string plotApartmentNum, bool isApartment, bool isSubdivision)` | none | `Utils.BuildAddressBookEntry`: world by exact or prefix match on public worlds, city by substring (`mist`, `goblet`, `empy`, `shiro`, `lavender`/`beds`/`lb`). Returns the tuple of a `null` entry (default values) when parsing fails; check `World` and the numbers. |
| `IsHere` | `bool IsHere(AddressBookEntryTuple)` | none | Current ward, district and plot/apartment match the entry (strict apartment room check). |
| `IsQuickTravelAvailable` | `bool IsQuickTravelAvailable(AddressBookEntryTuple)` | none | Same world, same district and ward, standing at a residential shard, and the ward has generated path data. |
| `GoToHousingAddress` | `void GoToHousingAddress(AddressBookEntryTuple)` | none at IPC level | `AddressBookEntry.GoTo()`: refuses with a chat error when no player or the world is not visitable, returns early when already at the address, else enqueues `TaskTpAndGoToWard` (cross-world when needed). |
| `GetAddressBookEntries` | `List<AddressBookEntryTuple> GetAddressBookEntries()` | none | All entries across folders. |
| `GetAddressBookEntriesWithFolders` | `Dictionary<string, List<AddressBookEntryTuple>> GetAddressBookEntriesWithFolders()` | none | Keyed by folder path, `""` for unplaced folders. |

## Movement

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `Move` | `void Move(List<Vector3> path)` | none | `P.FollowPath.Move(path, ignoreDeltaY: true)` with `destTolerance 0` and `tolerance 0.25f`. Replaces any current waypoints. |
| `MoveEx` | `void MoveEx(List<Vector3> path, bool? ignoreDeltaY, float? destTolerance, float? tolerance)` | none | Same with defaults `true`, `0`, `0.25f`. `ignoreDeltaY: false` enables the walk-to-fly branch. |

Neither gate calls vnavmesh. The follower is Lifestream's own
[FollowPath](source/Lifestream/Lifestream/Movement/FollowPath.cs), and the
caller supplies the whole path. Details and stop conditions are in
[travel](travel.md#movement-move-and-moveex).

## Instances

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `CanChangeInstance` | `bool CanChangeInstance()` | includes `!P.TaskManager.IsBusy` | `C.ShowInstanceSwitcher`, not waiting to visit another world or jumping, not occupied, current instance is non-zero, and an aetheryte (or the zone's extra instance changer NPC) is within 11 m. |
| `GetNumberOfInstances` | `int GetNumberOfInstances()` | none | Count recorded for the current territory in `C.PublicInstances`, learned from the SelectString menu the first time it opens; `0` when unknown. |
| `ChangeInstance` | `void ChangeInstance(int number)` | none | `TaskRemoveAfkStatus` then `TaskChangeInstance.Enqueue(number)`: optionally drop flight, interact with the aetheryte, select travel, select the instance, and re-enqueue while `C.InstanceSwitcherRepeat` and the instance still differs. No bounds or busy check at this layer. |
| `GetCurrentInstance` | `int GetCurrentInstance()` | none | `UIState.Instance()->PublicInstance.InstanceId`. |

## Login and character switching

`ErrorCode` is `ECommons.ErrorCode`
([source](source/ECommons/ECommons/ErrorCode.cs)):
`Success, Player_is_not_logged_in, Invalid_world_specified, Plugin_is_busy`.
The type is compiled into each plugin's own ECommons copy; confirm how your
binding receives it before matching on it.

`Utils.CanAutoLogin()` requires: not logged in, no condition flags set, the
`_TitleMenu` addon ready with its fourth node fully opaque, and neither
`TitleDCWorldMap` nor `TitleConnect` open. In short, the idle title screen.

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `CanAutoLogin` | `bool CanAutoLogin()` | none | As above. |
| `ConnectAndOpenCharaSelect` | `bool ConnectAndOpenCharaSelect(string charaName, string charaHomeWorld)` | `IsBusy()` | `TaskConnectAndOpenCharaSelect.Enqueue`: resolves the service account (AutoRetainer data when enabled, else `C.ServiceAccounts`, default 0), throws when the world is unknown, returns `false` unless `CanAutoLogin()`. Connects to the data center and selects the character's world. |
| `InitiateTravelFromCharaSelectScreen` | `bool InitiateTravelFromCharaSelectScreen(string charaName, string charaHomeWorld, string destination, bool noLogin)` | `IsBusy()` | From an open character list: finds the character by name and home world, then logs in on `destination` (DC travel when needed). `false` when the list is not open or the character is absent. Throws on an invalid home world. |
| `CanInitiateTravelFromCharaSelectList` | `bool CanInitiateTravelFromCharaSelectList()` | none | A valid `_CharaSelectListMenu` is open. |
| `ConnectAndTravel` | `bool ConnectAndTravel(string charaName, string charaHomeWorld, string destination, bool noLogin)` | `IsBusy()` and `CanAutoLogin()` | `ConnectAndOpenCharaSelect` then a queued `InitiateTravelFromCharaSelectScreenInternal`. |
| `InitiateLoginFromCharaSelectScreen` | `bool InitiateLoginFromCharaSelectScreen(string charaName, string charaHomeWorld)` | `IsBusy()` | Plain login of the named character from an open list. |
| `ConnectAndLogin` | `bool ConnectAndLogin(string charaName, string charaHomeWorld)` | `IsBusy()` and `CanAutoLogin()` | `ConnectAndOpenCharaSelect` then queued login. |
| `ChangeCharacter` | `ErrorCode ChangeCharacter(string name, string world)` | `Utils.IsBusy()` | `Plugin_is_busy`; `Invalid_world_specified` when `world` is not a public world; when a player is present, queues logout, chara select and login and returns `Success`; when at the title screen with `CanAutoLogin()`, queues the same without logout; else `Player_is_not_logged_in`. |
| `ChangeCharacterAndTravel` | `ErrorCode ChangeCharacterAndTravel(string name, string world, string destinationWorld)` | `Utils.IsBusy()` | Same, logging in on `destinationWorld`. |
| `Logout` | `ErrorCode Logout()` | `Utils.IsBusy()` | `Player_is_not_logged_in` without a player; otherwise `TaskLogout.Enqueue()`: wait until not casting or occupied, `/logout`, confirm the yes/no, wait for `CanAutoLogin()` (2 minute limit). |

## Custom aliases

| Gate | Signature | Busy check | What it does |
| --- | --- | --- | --- |
| `EnqueueCustomAlias` | `void EnqueueCustomAlias(CustomAlias alias, bool force, int? inclusiveStart, int? inclusiveEnd)` | `Utils.IsBusy()` unless `force` | `alias.Enqueue(force, inclusiveStart, inclusiveEnd, bypassCheck: true)`. The IPC path bypasses each command's `CanExecute` unlock checks. A refusal prints "Lifestream is busy!" and returns. |
| `EnqueueCustomAliasFromString` | `void EnqueueCustomAliasFromString(string aliasString, bool force, int? inclusiveStart, int? inclusiveEnd)` | same | `JsonConvert.DeserializeObject<CustomAlias>`; a null result throws `NullReferenceException`. Despite the parameter name, the second index is used as an exclusive end (`cmds.Take(exclusiveEnd - inclusiveStart)`). |

`CustomAliasKind` values, by JSON number:
`Teleport_to_Aetheryte 0, Move_to_point 1, Navmesh_to_point 2, Change_world 3, Use_Aethernet 4, Circular_movement 5, Interact 6, Mount_Up 7, Select_Yes 8, Select_List_Option 9, Confirm_Contents_Finder 10, Wait_for_Transition 11, Return_to_Home_World 12, Close_UI 13`
([source](source/Lifestream/Lifestream/Data/CustomAliasKind.cs)). The
shipped [StaticAlias](source/Lifestream/Lifestream/Data/StaticAlias.cs) JSON
strings are working examples of the format; the `splatoonscripts` skill
documents the schema field by field.

## Event

`Lifestream.OnHouseEnterError` fires from `TaskPropertyShortcut` when a
house path exists but the character is more than 10 m from the recorded
entrance after teleporting. Subscribe with
`GetIpcSubscriber<object>("Lifestream.OnHouseEnterError").Subscribe(...)`.
No other event exists; there is no arrival, completion or failure broadcast.

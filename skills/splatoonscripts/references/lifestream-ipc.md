# Lifestream IPC

Covers: the full Lifestream IPC method list (travel, housing, movement,
instances, login, aliases) and the enqueue / wait / verify snippet.

Verified against: Splatoon 3.9.2.32 (commit f396ef6, 2026-11-13), ECommons
bc38dc7, Lifestream b363201 — the pins in `references/provenance.md`.

Paths beginning with `references/` are relative to the skill root.

## Lifestream IPC (`references/lifestream/IPCProvider.cs`, all `Lifestream.<Name>`)

Travel / world:
`bool IsBusy()`, `void Abort()`, `void ExecuteCommand(string args)` (= `/li args`),
`bool Teleport(uint aetheryteId, byte subIndex)`, `bool TeleportToFC()`,
`bool TeleportToHome()`, `bool TeleportToApartment()`,
`bool AethernetTeleport(string destinationName)`,
`bool AethernetTeleportById(uint aetheryteRowId)`,
`bool AethernetTeleportByPlaceNameId(uint placeNameRowId)`,
`bool HousingAethernetTeleportById(uint)`, `bool AethernetTeleportToFirmament()`,
`uint GetActiveAetheryte()`, `uint GetActiveCustomAetheryte()`, `uint GetActiveResidentialAetheryte()`,
`bool ChangeWorld(string world)`, `bool ChangeWorldById(uint worldId)`,
`bool CanVisitSameDC(string world)`, `bool CanVisitCrossDC(string world)`,
`void TPAndChangeWorld(string w, bool isDcTransfer, string secondaryTeleport, bool noSecondaryTeleport, int? gateway, bool? doNotify, bool? returnToGateway)`,
`int? GetWorldChangeAetheryteByTerritoryType(uint)`, `uint GetRealTerritoryType()`.

Housing / inn:
`void EnqueueInnShortcut(int? innIndex)` (cross-world OK), `void EnqueueLocalInnShortcut(int? innIndex)`,
`void EnqueuePropertyShortcut(PropertyType type, HouseEnterMode? mode)`, `void EnterApartment(bool enter)`,
`bool? HasApartment()`, `bool? HasPrivateHouse()`, `bool? HasSharedEstate()`, `bool? HasFreeCompanyHouse()`,
`(ResidentialAetheryteKind Kind, int Ward, int Plot)? GetCurrentPlotInfo()`, `Vector3? GetPlotEntrance(uint territory, int plot)`,
`uint GetResidentialTerritory(ResidentialAetheryteKind)`, `GetHousePathData(ulong CID)`, `GetSharedHousePathData()`,
address book: `BuildAddressBookEntry(...)`, `IsHere(entry)`, `IsQuickTravelAvailable(entry)`, `GoToHousingAddress(entry)`, `GetAddressBookEntries()`, `GetAddressBookEntriesWithFolders()`,
workshop: `bool CanMoveToWorkshop()`, `void MoveToWorkshop()`.

Movement: `void Move(List<Vector3> path)`, `void MoveEx(List<Vector3> path, bool? ignoreDeltaY, float? destTolerance, float? tolerance)`.

Instances: `bool CanChangeInstance()`, `int GetNumberOfInstances()`, `void ChangeInstance(int)`, `int GetCurrentInstance()`.

Login / characters: `bool CanAutoLogin()`, `bool ConnectAndOpenCharaSelect(string name, string homeWorld)`,
`bool InitiateTravelFromCharaSelectScreen(name, homeWorld, destination, bool noLogin)`, `bool ConnectAndTravel(...)`,
`bool InitiateLoginFromCharaSelectScreen(name, homeWorld)`, `bool ConnectAndLogin(name, homeWorld)`,
`ErrorCode ChangeCharacter(string name, string world)`, `ErrorCode ChangeCharacterAndTravel(name, world, destWorld)`, `ErrorCode Logout()`.

Aliases: `void EnqueueCustomAlias(CustomAlias alias, bool force, int? inclusiveStart, int? inclusiveEnd)`,
`void EnqueueCustomAliasFromString(string json, bool force, int? inclusiveStart, int? inclusiveEnd)`.
`force = false` refuses when Lifestream is busy (`Notify.Error("Lifestream is busy!")`);
the IPC path bypasses the per-command `CanExecute` unlock checks. Event: `OnHouseEnterError`.

Most methods return `false`/do nothing when `IsBusy()`, and a refused
enqueue leaves `IsBusy()` false — so "Lifestream is not busy" says nothing
about whether *your* request ran. These alias methods return `void`; a call
returning does not acknowledge acceptance or completion. Track that you
requested travel, wait for idle, then **verify the state the next step
needs** (territory, shop identity, inventory) before acting on it.

```csharp
bool travelRequested = false;
TaskManager.Enqueue(() =>
{
    if(LifestreamIsBusy()) { PluginLog.Warning("Lifestream busy, alias not requested"); return; }
    EnqueueCustomAliasFromString(Alias, false, null, null);
    travelRequested = true; // request attempted; the void IPC has no acceptance result
}, "request alias");
TaskManager.Enqueue(() => !LifestreamIsBusy(), "travel idle",
    new(timeLimitMS: 300000, abortOnTimeout: true));
TaskManager.Enqueue(() => travelRequested && !LifestreamIsBusy()
    && Player.Available && Player.Interactable && Svc.ClientState.TerritoryType == 1186,
    "verify territory", new(timeLimitMS: 10000, abortOnTimeout: true));
```

This only gates steps requiring that territory. A purchase additionally needs
verification of the expected shop, currency, item and transaction prompt.
Timeouts explicitly abort this chain even if the manager has different defaults.
Cleanup can be a separate bounded recovery path; resuming AutoDuty is a dependent
action, not unconditional cleanup. See [vendor-run.md](vendor-run.md).


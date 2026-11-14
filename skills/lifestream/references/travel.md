# Travel, waiting and pitfalls

These details describe Lifestream's public source at
[the recorded pin](provenance.md). Recheck the target source after an
update. Requesting, teleporting, arriving and being usable again are four
separate states; Lifestream reports none of them to an IPC caller except
through `IsBusy()`.

## How Lifestream identifies the aetheryte you stand at

[DataStore](source/Lifestream/Lifestream/Systems/Legacy/DataStore.cs) builds
`Aetherytes`, a dictionary from each master aetheryte (`Aetheryte` sheet rows
with `IsAetheryte` and a non-zero `AethernetGroup`) to its shards (rows in
the same group with `IsAetheryte == false`). Each becomes a `TinyAetheryte`
with a 2D position computed in `GetTinyAetheryte`:

1. `StaticData.CustomPositions` (shipped `StaticData.json`) wins when the
   row ID is present.
2. Otherwise the first `Map` row whose `TerritoryType` matches the
   aetheryte's `Territory` supplies `SizeFactor`, and the first `MapMarker`
   subrow with `DataType == 3` and `DataKey == aetheryte row` (or
   `DataType == 4` and `DataKey == AethernetName row` for shards) supplies
   `X`/`Y`. `Utils.ConvertMapMarkerToRawPosition(pos, scale)` is
   `(pos - 1024) / (scale / 100)`.

That 2D point is only used for identification: every frame
`UpdateActiveAetheryte` takes the nearest targetable aetheryte object within
11 m horizontally and 15 m in 3D (`Utils.GetValidAetheryte`), then finds the
`TinyAetheryte` in the current `P.Territory` whose stored position is within
10 units of it. `P.ActiveAetheryte` is the result; `GetActiveAetheryte()`
exposes its ID. Lifestream never uses these positions to choose a
destination, and it has no "nearest aetheryte to a point" query.

## Teleporting near a world position in another zone

Lifestream resolves destinations by name, not by coordinates. To land near a
`(territory, position)` you compute the aetheryte yourself:

1. From the `Aetheryte` sheet, take rows with `IsAetheryte` and
   `Territory == territory`. Zones with aethernet shards have one master
   aetheryte per zone; field zones can have several.
2. Get each candidate's world position. The sheet's `Level` reference gives
   a `Level` row with world `X`/`Y`/`Z`; Lifestream itself reads
   `MapMarker` instead, and `Map.SizeFactor` is needed either way to turn
   marker units into world units (the formula above). Which source your
   host already loads decides; both are game data lookups outside
   Lifestream's IPC.
3. Pick the nearest candidate and call `Teleport(aetheryteId, 0)`.
   `subIndex` is non-zero only for housing and apartment entries in the
   player's aetheryte list. The call returns `false` when the Teleport
   action is on cooldown or unusable, when the player is not interactable, or
   when the aetheryte is not unlocked.
4. Wait as described below, then compare `GetRealTerritoryType()` with the
   intended territory before moving.

For a destination that is a shard rather than an aetheryte,
`AethernetTeleportById(shardRowId)` works only while standing at a member of
that group. The `/li <name>` path with `allowPartial` does the two-step
(teleport to the master, then use the aethernet) for you; the IPC
`AethernetTeleport*` gates do not.

## Waiting semantics

Lifestream's own tasks wait on these predicates, in this order, and a
consumer should mirror them:

| Predicate | Source | Meaning |
| --- | --- | --- |
| `Svc.Condition[BetweenAreas] || Svc.Condition[BetweenAreas51]` | `TaskTpToAethernetDestination`, `TaskTpAndWaitForArrival` | The teleport was accepted and the zone transition has begun. Until this flips, a `Teleport` that returned `true` can still fail silently (another cast interrupted it). |
| `IsScreenReady()` | ECommons | None of `NowLoading`, `FadeMiddle`, `FadeBack` is visible. |
| `Player.Interactable` | ECommons | `Svc.ClientState.LocalPlayer` exists and is targetable. |
| `WorldChange.WaitUntilNotBusy` | [WorldChange.cs](source/Lifestream/Lifestream/Schedulers/WorldChange.cs) | In a territory Lifestream knows (`DataStore.Territories`), `CastInfo.ActionId == 0`, not occupied, not waiting to visit another world or jumping, targetable. |
| `P.Territory == gateway.GetTerritory()` | `TaskTpToAethernetDestination` | The expected territory, after the fade. Two minute limit. |

`TaskTpToAethernetDestination.BuildTasks` is the canonical sequence: optional
`Utils.WaitForScreen` (when `C.WaitForScreenReady`, default true), the
teleport call under `ExecuteTPToAethernetDestination` (which also requires
`AgentMap.IsPlayerMoving == false`, not occupied, not casting, and a one
second throttle), `BetweenAreas`, `WaitUntilNotBusy` (2 min),
interactable in the right territory (2 min). `TaskSettings`
([source](source/Lifestream/Lifestream/Tasks/TaskSettings.cs)) defines the
limits: 1, 2, 5, 15, 60 minutes and infinite.

`IsBusy()` goes false the moment the queue drains, which can be a frame after
`WaitUntilNotBusy` passed and before the next step is safe. Add your own
screen-ready and interactable checks after idle, as the shipped
`splatoonscripts` snippet does.

## `TPAndChangeWorld`

Release `2.6.0.13` is described upstream as "Updated for world visit changes
brought by 7.58". That change is not in the public source, so the sequence
below is the pre-7.58 flow; see
[release notes](provenance.md#release-notes-2609-to-26013). Verify the
resulting world instead of relying on the steps.

`Lifestream.TPAndChangeWorld(destinationWorld, isDcTransfer, secondaryTeleport, noSecondaryTeleport, gateway, doNotify, returnToGateway, skipChecks)`
in [Lifestream.cs](source/Lifestream/Lifestream/Lifestream.cs) is what
`ChangeWorld`, `ChangeWorldById`, the IPC `TPAndChangeWorld`, the `/li
<world>` command and the `Change_world` alias all reach. It returns nothing.
Every early exit below is a `Notify` or a log line:

1. `Utils.AssertCanTravel` throws when a configured travel ban matches the
   character, source world and destination; the exception is caught and
   logged.
2. `CharaSelectVisit.ApplyDefaults`: `returnToGateway ??= C.DCReturnToGateway`,
   `gateway ??= C.WorldChangeAetheryte` (default `Uldah`), `doNotify ??= true`.
3. Unless `skipChecks`: a DC transfer with `C.AllowDcTransfer` off, or
   `TaskManager.IsBusy`, refuses ("Another task is in progress"). Note that
   this is the task-manager flag alone; a running path does not block it.
4. No player, or destination equals `Player.CurrentWorld`, refuses.
5. A party of more than one only warns.

Same data center: `TaskRemoveAfkStatus`, then
[`TaskTPAndChangeWorld.Enqueue(world, gateway.AdjustGateway(), false)`](source/Lifestream/Lifestream/Tasks/CrossWorld/TaskTPAndChangeWorld.cs).
`AdjustGateway` swaps the gateway for the first unlocked city aetheryte when
the configured one is not in `Svc.AetheryteList`. If the active aetheryte is
already a world-change aetheryte (IDs 2, 8, 9) it goes straight to
`TaskChangeWorld`; otherwise, when no world-change aetheryte is within 30 m,
it teleports to the gateway, then targets, locks on and automoves to the
aetheryte until `P.ActiveAetheryte` is a master aetheryte, waits ten frames
and runs `TaskChangeWorld`. A desktop notification and the secondary
aethernet teleport follow when configured.

[`TaskChangeWorld.Enqueue`](source/Lifestream/Lifestream/Tasks/CrossWorld/TaskChangeWorld.cs)
optionally leaves the party (and closes a party-finder listing), then: close
`WorldTravelSelect` if open, wait until not occupied, target and interact
with the aetheryte, select "Visit Another World", pick the world in the
`WorldTravelSelect` list (read from the `WorldTranslate` string array,
indices 3 to 10, `Callback.Fire(addon, true, 0, index + 2)`), confirm the
yes/no. With `C.RetryWorldVisit` (default on) it then waits up to five
minutes for `Player.CurrentWorld == target`, re-running the sequence every
`C.RetryWorldVisitInterval` seconds (default 30, plus a random delta) when
the server refuses. Error toasts with `LogMessage` 9419, 9426 or 9427 set
`WVErrorDetected`, which closes the dialog and retries.

Cross data center: the type is decided from home, current and target DC
(`HomeToGuest`, `GuestToHome`, `GuestToGuest`). With
`C.TeleportToGatewayBeforeLogout` it first teleports to the gateway unless in
a sanctuary or a territory with instant logout; with
`C.LeavePartyBeforeLogout` it leaves the party. `HomeToGuest` returns to the
home world if needed, waits for it, logs out and relogs, then
[`CharaSelectVisit.HomeToGuest`](source/Lifestream/Lifestream/Systems/CharaSelectVisit.cs)
drives the title screen and character list, waits until in the destination
world, enforces it, and runs the gateway return, notification and secondary
teleport. The other two types start with the matching return or DC change
task. Expect minutes, a logout, and a character-select pass; the character
must have a service account Lifestream can resolve (AutoRetainer data when
`C.UseAutoRetainerAccounts`, else `C.ServiceAccounts`).

## Movement: `Move` and `MoveEx`

`Move(path)` and `MoveEx(path, ignoreDeltaY, destTolerance, tolerance)` call
`P.FollowPath.Move(...)` on Lifestream's own
[FollowPath](source/Lifestream/Lifestream/Movement/FollowPath.cs), whose
movement code [credits awgil](source/Lifestream/Lifestream/Movement/README.md).
They do not call vnavmesh and they do not compute a route; the list you pass
is followed point by point with the camera and movement overrides.

Where vnavmesh does appear in Lifestream, it is a pathfinding client only:
[`VnavmeshIPC`](source/Lifestream/Lifestream/IPC/VnavmeshIPC.cs) binds
`Nav.IsReady`, `Nav.Pathfind`, `SimpleMove.PathfindInProgress`,
`Path.IsRunning`, `Path.Stop` and two mesh queries with
`SafeWrapper.AnyException`, so a missing vnavmesh returns defaults instead of
throwing. `Nav.Pathfind` is consumed by the `Navmesh_to_point` alias kind,
`TaskGeneratePath` (housing paths), the island shortcut and the debug UI;
each then hands the resulting list to `FollowPath.Move`. The Movement README
and `CustomAliasKind.Navmesh_to_point` can suggest that the movement gates
need vnavmesh. They do not: `Move`/`MoveEx` never depend on vnavmesh, while
`FollowPath.Update` does poll it.

`FollowPath.Update`, each frame with a player present:

- If `VnavmeshIPC.IsRunning()` is true, clears all waypoints and logs
  "Detected vnavmesh movement, Lifestream will abort all tasks now."
  (Without vnavmesh the wrapper returns `false`, so this never fires.)
- A 60 second throttle `FollowPathTime` is armed when movement starts and
  reset every time a waypoint is consumed; when it expires the list is
  cleared with "Lifestream movement has timed out."
- With `destTolerance > 0`, being within that distance of the last point
  clears the list.
- A waypoint is consumed when the segment between the previous and current
  frame positions passes within `tolerance` (default 0.25) of it; with
  `ignoreDeltaY` the Y axis is zeroed for that test.
- With `ignoreDeltaY == false`, a next waypoint above the player while not
  in flight or diving either spams jump when mounted or disables movement
  and stalls when not mounted. `Move` passes `true`, so this branch is only
  reachable through `MoveEx`.
- `MovementConfig` defaults (`StopOnStuck false`, `CancelMoveOnUserInput
  false`, `AlignCameraToMovement false`) are not exposed over IPC.

`IsBusy()` is the only completion signal (`Waypoints.Count > 0`). A cleared
list after a timeout or a vnavmesh collision looks identical to arrival, so
check the player position against your last point. Lifestream also publishes
a shared Dalamud data entry `Lifestream.PathIsRunning` (`bool[]`, index 0)
through `GetOrCreateData`; it mirrors the same state and is not an IPC gate.

## Known pitfalls, from the source

- A refused request looks like success. `ChangeWorld` returns `true` before
  `TPAndChangeWorld` can refuse for `AllowDcTransfer`, a travel ban, a party
  or an already-current world. `AethernetTeleport` returns `true` before the
  task discovers the destination is not reachable. The `void` gates return
  nothing at all. Always poll `IsBusy()` shortly after the call; if it never
  went true, the request was dropped.
- Two different busy flags. `IsBusy()`, `Utils.IsBusy()`,
  `EnqueueCustomAlias` and `ChangeCharacter` include a running path.
  `TeleportToFC/Home/Apartment`, `TaskPropertyShortcut.Enqueue`,
  `TPAndChangeWorld` (without `skipChecks`), `/li mb` and `/li tp` test
  `TaskManager.IsBusy` only, so they will queue while `Move` is still
  walking.
- `Abort()` and `/li stop` clear the whole queue, including tasks queued
  by the user or other plugins. There is no per-consumer cancellation.
- `ExecuteCommand` chains with `,`: a world name followed by an alias word
  is rewritten into a second command that runs after arrival. A destination
  whose name ends with one of `Utils.LifestreamNativeCommands` (for example
  anything ending in ` is` or ` tp`) is split unexpectedly.
- `GetWorldChangeAetheryteByTerritoryType` throws for territories other than
  129, 130 and 132 instead of returning `null`.
- `HasApartment` and siblings dereference `Player.Object`; call them only
  while logged in.
- `Teleport` requires the aetheryte to be unlocked for the current character
  and the Teleport action to be usable; it does not wait, queue or retry.
  `TeleportService.ReliableTeleportToAetheryte` (used by the
  `Teleport_to_Aetheryte` alias kind) retries on a throttle, but is not
  exposed over IPC.
- `GetRealTerritoryType()` depends on `TerritoryWatcher` having seen a
  residential outdoor zone since load; inside a house loaded fresh it falls
  back to the raw territory and the plugin warns about data reliability.
- World names must match the `World` sheet exactly for `CanVisit*` and
  `ChangeWorld`; the `/li` text path accepts prefixes, the IPC gates do not.
- Everything runs on the framework thread through ECommons `TaskManager`.
  Calling into Lifestream from another thread is unsupported by the
  provider.

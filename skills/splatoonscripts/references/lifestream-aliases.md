# Lifestream custom alias JSON

Source: `Lifestream/Data/CustomAlias.cs`, `CustomAliasCommand.cs`,
`CustomAliasKind.cs` (copies in this directory). Consumed by IPC
`Lifestream.EnqueueCustomAliasFromString(string json, bool force, int?
inclusiveStart, int? inclusiveEnd)` or exported/imported in the Lifestream
UI ("Custom Aliases" tab).

## Envelope (`CustomAlias`)

```json
{"ExportedName":"Stone Exch","Alias":"stones","Enabled":true,"Commands":[ ... ]}
```

- `ExportedName` string — display name.
- `Alias` string — chat alias (`/li stones`); omitted when empty.
- `Enabled` bool — serialised only when false.
- `GUID` — serialised when non-empty; irrelevant for IPC.
- `Commands` — array of `CustomAliasCommand`, run in order.

`Enqueue(force, inclusiveStart, exclusiveEnd, bypassCheck)`: refuses with
"Lifestream is busy!" unless `force`; `inclusiveStart`/`exclusiveEnd`
select a sub-range of commands. Through IPC `bypassCheck = true`, so the
per-command `CanExecute` (aetheryte/shard unlocked, world visitable) is
skipped. Consecutive `Move_to_point` commands that are chained
(`ChainGroup`, editor-only, not serialised) are merged into one path.

## `CustomAliasKind` (numeric `Kind`)

| Kind | Name | Fields used | Behaviour |
|---|---|---|---|
| 0 | `Teleport_to_Aetheryte` | `Aetheryte` (Aetheryte row id, IsAetheryte=true), `SkipTeleport` (float, default 15) | wait for screen ready + interactable; if the nearest eligible aetheryte is within `SkipTeleport` and we are already in the aetheryte's territory, skip; else `ReliableTeleportToAetheryte(Aetheryte, wait: true)` |
| 1 | `Move_to_point` | `Point` {X,Y,Z}, `ExtraPoints` [{X,Y,Z}] (random pick among Point+ExtraPoints), `Territory` (uint; 0 = don't wait), `UseFlight` (bool), `Scatter` (float radius) | wait until in `Territory` (if set), screen ready; optionally fly; sprint; `FollowPath.Move([point, ...appended])` straight line; wait until path empty (5 min); stops early if `Occupied33` or territory changes |
| 2 | `Navmesh_to_point` | `Point`, `Territory`, `UseFlight`, `UseTA` (bool) | like 1 but pathfinds with vnavmesh (`Pathfind(Player.Position, Point, UseFlight)`), or TextAdvance 2D move when `UseTA` and TextAdvance is loaded |
| 3 | `Change_world` | `World` (World row id) | if not already on that world: DC travel if `CanVisitCrossDC`, else same-DC travel |
| 4 | `Use_Aethernet` | `Aetheryte` (row id of the aethernet shard, IsAetheryte=false) | must be near an aetheryte/shard of the same network; `TaskTryTpToAethernetDestination`; then waits for screen not-ready → ready |
| 5 | `Circular_movement` | `CenterPoint` {X,Y} (Vector2 = world X/Z), `CircularExitPoint` {X,Y,Z}, `Precision` (float, 20), `Tolerance` (int, 1), `Clamp` {Min,Max} or null, `WalkToExit` (bool, true) | walks an arc around `CenterPoint` from current position to the exit angle, optionally continuing to the exit point |
| 6 | `Interact` | `DataID` (uint), `InteractDistance` (float?, null) | dismount; selector = nearest targetable object with `DataId == DataID`; if `InteractDistance` set, `NeoTasks.ApproachObjectViaAutomove(selector, InteractDistance)`; then `NeoTasks.InteractWithObject(selector)` |
| 7 | `Mount_Up` | `MountUpConditional` (bool) | mounts (`TaskMount.MountIfCan`) unless conditional and the user disabled mounts in Lifestream |
| 8 | `Select_Yes` | `SelectOption` [string], `StopOnScreenFade` (bool), `ExcelOnlyFirst` (bool) | when a `SelectYesno` whose text contains any option is ready, press Yes (throttled 200 ms). 10 s timeout, non-aborting. Empty strings ignored |
| 9 | `Select_List_Option` | `SelectOption` [string], `StopOnScreenFade`, `ExcelOnlyFirst` | on `SelectString` or `SelectIconString`: pick the first entry that equals, else contains, any option; done once clicked and the menu is gone. 10 s, non-aborting |
| 10 | `Confirm_Contents_Finder` | `StopOnScreenFade` | clicks Commence on `ContentsFinderConfirm`; 20 s, non-aborting |
| 11 | `Wait_for_Transition` | `RequireTerritoryChange` (bool) | wait for screen not-ready, then ready + interactable (+ territory changed if required) |
| 12 | `Return_to_Home_World` | `World` (row id, used for the name) | no-op if already on home world; otherwise DC/world travel |
| 13 | `Close_UI` | `SelectOption` [exact addon names], `Timeout` (ms, 0 → 30000), `RequireUiOpen` (bool) | if `RequireUiOpen`, first waits until one of the addons is ready; then fires `Callback.Fire(addon, true, -1)` (throttled) on each until none remain |

`SelectOption` strings can be sheet patterns
`<SheetName:Path:Row:Column>` resolved by `Utils.ParseSheetPattern`
(e.g. `<QuestDialogueText:transport/AetheryteBestwaysBurrow:4:Value>`),
which makes an alias language-independent. `ExcelOnlyFirst` limits that
resolution to the first match.

## Other `CustomAliasCommand` fields

- `Precision`, `Tolerance`, `Clamp`, `WalkToExit` — circular movement only.
- `SkipTeleport` — teleport only.
- `UseTA` — navmesh only (TextAdvance movement).
- `NoDisableYesAlready` (bool) — serialised when true; reserved for YesAlready interaction.
- `Timeout`, `RequireUiOpen` — Close_UI.
- `ID`, `ChainGroup` — internal, not serialised.

Only fields differing from defaults are written (`ShouldSerializeX`), so a
minimal command like `{"Kind":6,"DataID":1049079}` is valid.

## Examples from `Lifestream/Data/StaticAlias.cs`

```json
{"ExportedName":"Cosmic","Alias":"cosmic","Enabled":false,"Commands":[
 {"Kind":0,"Aetheryte":175,"SkipTeleport":35.0},
 {"Kind":6,"DataID":175,"InteractDistance":10.0},
 {"Kind":9,"SelectOption":["<QuestDialogueText:transport/AetheryteBestwaysBurrow:4:Value>","<QuestDialogueText:transport/AetheryteBestwaysBurrow:3:Value>","<QuestDialogueText:transport/AetheryteBestwaysBurrow:2:Value>","<QuestDialogueText:transport/AetheryteBestwaysBurrow:1:Value>"],"StopOnScreenFade":true}]}
```

Solution Nine tomestone vendor:

```json
{"ExportedName":"Stone Exch","Commands":[{"Kind":0,"Aetheryte":217},{"Kind":4,"Aetheryte":235},{"Kind":1,"Point":{"X":-161.6475,"Y":0.6600001,"Z":1.0833879},"Territory":1186},{"Kind":1,"Point":{"X":-184.94943,"Y":0.6600001,"Z":-28.184694},"Territory":1186},{"Kind":6,"DataID":1049079},{"Kind":9,"SelectOption":["Allagan Tomestones of Mathematics (Other)"]}]}
```

Aetheryte 217 = Solution Nine (IsAetheryte, AethernetGroup 15); 235 =
shard "Nexus Arcade" (AethernetGroup 15, territory 1186); Level row for
Object 1049079 ("Zircon") is at world (-187.70, 0.66, -30.90), matching the
second move point.

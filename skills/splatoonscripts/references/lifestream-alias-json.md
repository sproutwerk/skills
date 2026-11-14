# Lifestream custom alias JSON

Covers: the `CustomAlias` JSON shape, the numeric `Kind` table, how
`Interact` and `SelectOption` behave, and a worked alias.

Verified against: Splatoon 3.9.2.32 (commit f396ef6, 2026-11-13), ECommons
bc38dc7, Lifestream b363201 — the pins in `references/provenance.md`.

Paths beginning with `references/` are relative to the skill root.

## Lifestream custom alias JSON

`CustomAlias` = `{"ExportedName": "...", "Commands": [ CustomAliasCommand, ... ]}`.
Commands run in order through Lifestream's own TaskManager. `Kind` is the
numeric `CustomAliasKind`:

```
0 Teleport_to_Aetheryte   Aetheryte (Aetheryte row id); SkipTeleport (default 15f: skip if already within this distance of an eligible aetheryte in the right territory)
1 Move_to_point           Point {X,Y,Z} (+ExtraPoints, random pick); Territory (wait until in it; 0 = any); UseFlight; Scatter; straight-line walk via Lifestream FollowPath, stops if Occupied33
2 Navmesh_to_point        Point; Territory; UseFlight; UseTA (TextAdvance instead of vnavmesh). Requires vnavmesh
3 Change_world            World (World row id); uses DC travel if needed
4 Use_Aethernet           Aetheryte (row id of the aethernet *shard*, IsAetheryte=false, same AethernetGroup); must be within range of an aetheryte/shard
5 Circular_movement       CenterPoint {X,Y} (Vector2), CircularExitPoint {X,Y,Z}, Precision (20), Tolerance (1), Clamp {Min,Max}?, WalkToExit (true)
6 Interact                DataID (ENpcBase/EObj data id); InteractDistance (float?, null = no approach)
7 Mount_Up                MountUpConditional (only if Lifestream's UseMount is on)
8 Select_Yes              SelectOption [strings]; StopOnScreenFade; ExcelOnlyFirst; 10 s, non-aborting
9 Select_List_Option      SelectOption [strings] matched against SelectString / SelectIconString entries (equals, then contains); 10 s, non-aborting
10 Confirm_Contents_Finder  clicks Commence on ContentsFinderConfirm; 20 s
11 Wait_for_Transition    waits for screen not-ready then ready; RequireTerritoryChange
12 Return_to_Home_World   World (row id); no-op when already home
13 Close_UI               SelectOption = exact addon names to close via Callback.Fire(addon, true, -1); Timeout (ms, 0 = 30000); RequireUiOpen
```

Only non-default fields are serialised. Full field-by-field table and
defaults: `references/lifestream-aliases.md`.

`Interact` (verified in `CustomAliasCommand.Enqueue`): selector is
`Svc.Objects.OrderBy(Player.DistanceTo).FirstOrDefault(x => x.IsTargetable && x.DataId == DataID)`
— nearest targetable object with that DataId. If `InteractDistance` is set,
`NeoTasks.ApproachObjectViaAutomove(selector, InteractDistance)` runs first
(target, `/lockon on` + `/automove on`, `/automove off` when within range),
then `NeoTasks.InteractWithObject(selector)` (target then
`TargetSystem.InteractWithObject`). It dismounts first. So a
`Move_to_point` that lands within ~4–8 y is enough; set `InteractDistance`
when the arrival point is farther away.

`SelectOption` strings may be `<Sheet:Path:Row:Column>` sheet patterns
(e.g. `<QuestDialogueText:transport/AetheryteBestwaysBurrow:4:Value>`)
that Lifestream resolves per language; plain text is matched as-is. Match
is "equals, else contains" — the shop menu entry must actually exist on
that NPC in the current game version (see XIVAPI `ENpcData` in
`references/xivapi-lookup.md`).

Worked alias (Solution Nine tomestone vendor):

```json
{"ExportedName":"Stone Exch","Commands":[
 {"Kind":0,"Aetheryte":217},                                   // Solution Nine aetheryte
 {"Kind":4,"Aetheryte":235},                                   // aethernet shard "Nexus Arcade" (AethernetGroup 15)
 {"Kind":1,"Point":{"X":-161.6475,"Y":0.6600001,"Z":1.0833879},"Territory":1186},
 {"Kind":1,"Point":{"X":-184.94943,"Y":0.6600001,"Z":-28.184694},"Territory":1186},
 {"Kind":6,"DataID":1049079},                                  // ENpcBase 1049079 = "Zircon"
 {"Kind":9,"SelectOption":["Allagan Tomestones of Mathematics (Other)"]}]}
```

Consecutive `Move_to_point` commands that are "chained" in the editor are
merged into one path; from JSON they simply run back-to-back.

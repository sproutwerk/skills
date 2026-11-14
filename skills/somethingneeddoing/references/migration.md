# Legacy scripts and the rewrite

Most Lua scripts in circulation were written for the plugin before its
rewrite. They do not run on current builds. The plugin says so itself in a
one-time notice (`Gui/Modals/FirstTimeWarningModal.cs`):

> SomethingNeedDoing has been fully rewritten to support the framework
> changes from API 12. Native macros should work much the same as before.
> Lua macros will not work at all. Scripts authors will need to write new
> scripts. There is a legacy macro importer located in the settings menu.

## Telling the two generations apart

| Sign | Generation |
| --- | --- |
| Bare global functions: `GetCharacterCondition(4)`, `IsAddonReady("X")`, `GetNodeText("X", 2, 3)`, `PathfindAndMoveTo(x, y, z)`, `GetItemCount(123)`, `LogInfo("...")` | Legacy |
| Module tables: `Player.`, `Entity.`, `Addons.GetAddon(...)`, `Inventory.`, `Excel.`, `IPC.vnavmesh.` | Current |
| A `[[SND Metadata]]` block, `Config.Get`, `TriggerData`, `OnAddonEvent_...` functions | Current |
| A one-word `/click name`, `/click "RecipeNote Synthesize"` in quotes, `<listindex.N>` | Older native syntax |

A legacy script on a current build fails at its first legacy call with a
Lua error of the form `attempt to call a nil value (global 'IsAddonReady')`.

The legacy function names above were confirmed in the upstream repository's
own history: they were public methods under
`SomethingNeedDoing/Old/Macros/LuaFunctions/` at commit
`7a44a440f5927bee304dae6b685a8eb9b8854566` (2025-04-20), and that directory
was deleted in `cfb0425` (2025-04-23), before the first rewrite tag
`v12.2`. None of those methods exists in the pinned tree. No legacy source
is bundled with this skill.

## What carried over

- `yield("/command")` is still how Lua issues commands, and
  `yield("/wait N")` is still how it waits.
- Native macro lines, `<wait.N>`, `<maxwait.N>`, `<condition...>`,
  `<unsafe>`, `<hq>`, `<echo>`, `/loop`, `/gate`, `/waitaddon`, `/require`,
  `/runmacro`, `/send`, `/recipe`, `/item`, `/target` kept their meaning.
- `/snd run "Name"` and `/pcraft` still work.

## What changed for native macros

- `/click` takes an addon name and a method name as two words
  (`/click RecipeNote Synthesize`). The single-token names of the old click
  library are gone, and quoting both words together breaks parsing. The
  config migration from version 1 to 2 rewrote the default craft-loop
  template for exactly this.
- `<listindex.N>` was renamed `<list.N>` in 12.64.
- `<errorif.X>` modifiers and the commands `/callback`, `/interact`,
  `/targetenemy`, `/equipitem`, `/keyitem`, `/hold` and `/release` belong to
  the rewrite. Check [native-macros.md](native-macros.md) before assuming an
  old command still exists: for example the legacy tree at
  `daemitus/SomethingNeedDoing` had `/requirequality` and `/requirestats`,
  and the pinned parser has neither. On a current build such a line falls
  through to the game as an unknown chat command.
- Crafting-loop settings moved from per-macro flags to metadata
  (`crafting_loop`, `craft_loop_count`); the legacy importer carries them
  over.

## Translating Lua calls

The right-hand column is this skill's mapping from reading both trees.
Upstream publishes no conversion table, and semantics differ in places
noted below. Treat each row as a starting point and confirm against the
Help window's Lua tab.

| Legacy global | Current equivalent |
| --- | --- |
| `IsAddonReady(name)` | `Addons.GetAddon(name).Ready` |
| `IsAddonVisible(name)` | `Addons.GetAddon(name).Ready`, or `.Exists` when visibility does not matter |
| `GetNodeText(name, a, b, ...)` | `Addons.GetAddon(name):GetNode(a, b, ...).Text`, with a new chain (see below) |
| `IsNodeVisible(name, ...)` | `Addons.GetAddon(name):GetNode(...).IsVisible` |
| `GetCharacterCondition(n)` | No module function. `Player.IsBusy`, `Player.Available`, `Player.IsMoving`, `Entity.Player.IsCasting`, `IsInCombat`, `IsMounted` cover common flags; otherwise `Svc.Condition` through the CLR bridge |
| `IsPlayerAvailable()` | `Player.Available` |
| `IsPlayerOccupied()` | `Player.IsBusy` (closest) |
| `IsPlayerCasting()` | `Entity.Player.IsCasting` |
| `IsMoving()` | `Player.IsMoving` |
| `GetZoneID()`, `IsInZone(id)` | `Svc.ClientState.TerritoryType` |
| `GetCharacterName()` | `Entity.Player.Name` |
| `GetClassJobId()` | `Player.Job.Id` |
| `GetLevel()` | `Player.Job.Level` |
| `GetHP()` | `Entity.Player.CurrentHp` |
| `GetTargetName()` | `Entity.Target and Entity.Target.Name` |
| `GetTargetHP()`, `GetTargetHPP()` | `Entity.Target.CurrentHp`, `.HealthPercent` |
| `GetPlayerRawXPos()` and siblings | `Entity.Player.Position.X` / `.Y` / `.Z` |
| `GetDistanceToPoint(x, y, z)` | Compute from `Entity.Player.Position`, or `entity.DistanceTo` for an object |
| `HasStatus(name)`, `GetStatusTimeRemaining(id)` | Iterate `Player.Status` and compare `StatusId`; read `RemainingTime` |
| `GetItemCount(id)` | `Inventory.GetItemCount(id) + Inventory.GetHqItemCount(id)` |
| `GetInventoryFreeSlotCount()` | `Inventory.GetFreeInventorySlots()` |
| `GetGil()` | `Inventory.GetItemCount(1)` |
| `IsQuestComplete(id)` | `Quests.IsQuestComplete(id)` |
| `GetActiveFates()` | `Fates.GetActiveFates()` (wrappers; read `.Id`) |
| `GetFlagXCoord()` | `Instances.Map.Flag.XFloat` |
| `LogInfo(text)` | `Dalamud.Log(text)` |
| `HasPlugin(name)` | `plugin_dependencies` in metadata, or `IPC.IsInstalled(name)` for wrapped plugins |
| `NavIsReady()` | `IPC.vnavmesh.IsReady()` |
| `PathfindAndMoveTo(x, y, z, fly)` | `IPC.vnavmesh.PathfindAndMoveTo(Vector3(x, y, z), fly)` |
| `PathfindInProgress()`, `PathIsRunning()`, `PathStop()` | `IPC.vnavmesh.PathfindInProgress()`, `.IsRunning()`, `.Stop()` |
| `LifestreamIsBusy()`, `LifestreamExecuteCommand(s)`, `LifestreamAbort()` | `IPC.Lifestream.IsBusy()`, `.ExecuteCommand(s)`, `.Abort()` |
| `ARIsBusy()`, `ARGetMultiModeEnabled()` | `IPC.AutoRetainer.IsBusy()`, `.GetMultiModeEnabled()` |
| `DeliverooIsTurnInRunning()` | `IPC.Deliveroo.IsTurnInRunning()` |
| `PauseYesAlready()` | `plugins_to_disable: [YesAlready]`, or `IPC.YesAlready.PausePlugin(ms)` |

Differences that bite during a port:

- **Node addressing.** Legacy `GetNodeText` walked `UldManager.NodeList` by
  position. `GetNode` follows a chain of node ids from the root. The
  numbers in an old script are positions and will read the wrong node or
  nothing. Rebuild each chain with the addon inspector.
- **Item counts.** Legacy `GetItemCount` included HQ by default. The
  current `GetItemCount` with a base item id counts normal quality only.
- **Return shapes.** Legacy calls returned numbers, strings and Lua-friendly
  lists. Current calls often return wrapper objects or .NET lists, which
  are 0-based and need `.Count` and `[i]` or `luanet.each`.
- **Nil instead of defaults.** Legacy helpers tended to return `""` or `0`
  with no target. `Entity.Target` is `nil`.
- **Vectors.** Position arguments are `System.Numerics.Vector3` objects,
  not three numbers.
- **Missing plugins.** Legacy IPC helpers existed as globals whether or not
  the plugin was installed. Now declare `plugin_dependencies`.
- **Conditions by number.** Scripts full of `GetCharacterCondition(26)`
  need each number mapped to a property or a `ConditionFlag` member; the
  numbers are the enum's values.

## Porting procedure

1. Import the old macros with Settings, "Legacy Macro Import", so the text
   is in the library for reference. Imported Lua keeps its type and name.
2. List every bare global call in the script. Anything not defined by the
   script itself is a legacy function.
3. Replace each with the table above, checking the member in the Help
   window's Lua tab.
4. Replace hand-rolled "wait for X" loops with a deadline helper
   ([patterns.md](patterns.md)); legacy scripts often loop forever.
5. Move user-tunable variables at the top of the script into `configs`.
6. Move "call this helper file" conventions (`require` of a shared
   functions file, or pasted helper blocks) into a `dependencies` entry or
   a `require` of a stored macro, and port the helper file first.
7. Add `plugin_dependencies` for every `IPC.` table the script touches.
8. Run it with the Dalamud log open and fix nil errors one at a time.

Shared "functions file" libraries from community repositories are the usual
blocker: a script that calls two hundred helper functions cannot be ported
until the helper file is. Check whether its maintainers publish a version
for current SND before translating it by hand.

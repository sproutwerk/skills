# Lua API at v15.13

Derived from `LuaMacro/` and `External/` at commit
`527ca83892506c74fea4d5e07b9b9436e50ef95d`. Bundled copies:
[LuaMacro/](source/SomethingNeedDoing/SomethingNeedDoing/LuaMacro/),
[External/](source/SomethingNeedDoing/SomethingNeedDoing/External/).

The plugin generates its own reference at runtime from the same attributes
this file was read from: Help, then the Lua tab, lists every module, wrapper
member and enum value of the installed build, and clicking an entry copies
its call path. Prefer that tab over this file when they disagree, since the
installed version may be newer than the pin. The plugin also writes
LuaLS-style stub files to a `stubs` directory inside its config directory
(`Services/StubGeneratorService.cs`) for editor completion.

## Execution model

Read this before anything else; most hard failures come from it.

`NLuaMacroEngine.StartMacro` builds a fresh `NLua.Lua` state per run, then
loads your script (minus its metadata block) inside this wrapper
(`LuaCodeSnippets.EntrypointTemplate`):

```lua

yield = coroutine.yield
--
function entrypoint()
<your script>
end
--
return coroutine.wrap(entrypoint)
```

Consequences:

- Your whole script is the body of a coroutine. `yield` is a global alias
  for `coroutine.yield`.
- The engine resumes the coroutine on the game's framework thread
  (`Svc.Framework.RunOnTick(() => generator.Call())`). Lua code between two
  yields runs to completion inside one frame. A loop that does not yield
  freezes the game, and Stop cannot interrupt it, because cancellation and
  pause are only checked between resumes.
- Each resume must yield a string or finish:
  - `yield("/some command")` runs that text as a child native macro, waits
    for it to complete, waits one more framework tick, then resumes you. The
    text goes through the [native parser](native-macros.md), so it may hold
    several lines, SND commands and modifiers, and it must start with `/`.
  - `yield(123)` or any non-string throws
    `Lua Macro yielded a non-string value`.
  - `yield()` with no value ends the script. The engine treats an empty
    result as "the function returned" and stops resuming.
- `yield("/wait 0.1")` is the idiom for "let the game run for a while".
  There is no cheaper frame yield at the pin.
- Line numbers in Lua errors are offset: the wrapper adds 4 lines before
  your first line, and the metadata block is removed before loading, so
  subtract 4 and add back the length of the metadata block if it sits above
  the failing code.
- Top-level `local` variables are locals of `entrypoint`. Top-level
  functions declared with `function name()` are globals, which is what
  trigger and cleanup discovery relies on.
- State does not persist between runs. Use `Config.Set` for values that
  must survive (see [metadata.md](metadata.md#configs)).

A child command that fails (for example a syntax error in the yielded text)
sets the child's state to `Error`; the scheduler copies `Error` to the root
parent and cancels it. The script ends at that yield without a Lua error.

## Globals set up for every script

From `NLuaMacroEngine.CreateEnvironment` and `Utils/LuaExtensions.cs`:

| Global | What it is |
| --- | --- |
| `yield` | `coroutine.yield` |
| `luanet`, `import`, `CLRPackage` | NLua's CLR bridge (`lua.LoadCLRPackage()`). `luanet.load_assembly`, `luanet.import_type`, `luanet.enum`, `luanet.make_array`, `luanet.each`, `luanet.ctype` and friends. |
| `f` | F-string helper: `f("HP {Entity.Player.CurrentHp}")` evaluates each `{...}` as a Lua expression, seeing locals and globals. |
| `snd.require.paths`, `snd.require.add_paths(...)` | Search path list for `require`; seeded from the "Lua require paths" setting, keeping only absolute directories that exist. |
| `print` | Replaced. Writes its arguments, tab-joined, to the Dalamud log at Information level. It does not write to chat. |
| `Svc` | ECommons `Svc`, registered with `luanet.import_type('ECommons.DalamudServices.Svc')()`. Gives access to Dalamud services such as `Svc.ClientState`, `Svc.Condition`, `Svc.Targets`, `Svc.Objects`, `Svc.Data`. |
| `TriggerData` | Only when the run or the current function call was started by a trigger. See [metadata.md](metadata.md#triggerdata). |
| `InternalGetMacroText` | Internal helper used by `require`. |
| Modules | `Actions`, `Addons`, `Dalamud`, `Entity`, `Excel`, `Fates`, `InstancedContent`, `Instances`, `Inventory`, `IPC`, `Player`, `Quests`, `System`, plus per-script `Engines` and `Config`. |
| Enums | Every enum type that appears in a module, wrapper or IPC signature is registered as a global under its short type name, for example `InventoryType`, `ActionType`, `ObjectKind`, `FateState`. |

The standard Lua libraries are open, including `os`, `io` and `debug`, and
the CLR bridge reaches any loaded .NET type. A script has the same access to
the machine as the plugin does. Read third-party scripts before running
them.

### Calling convention

- Module members are plain functions and properties on a table:
  `Player.IsBusy`, `Inventory.GetItemCount(5111)`. Use a dot.
- Objects returned by modules ("wrappers" and other .NET objects) follow
  NLua's rules: properties with a dot, methods with a colon.
  `Entity.Target:Interact()`, `Addons.GetAddon("Talk"):GetNode(1, 2)`.
- Module properties are read through a metatable `__index` and are
  recomputed on every access. A few have setters in C#, but assignment from
  Lua writes a plain table field and does not reach the setter. Wrapper
  properties with setters (for example `Node.Text`,
  `Instances.DutyFinder.IsUnrestrictedParty`) are real .NET properties and
  can be assigned.
- Wrappers hold names or ids and look the game object up on each access.
  `Addons.GetAddon("X")` is cheap and can be created before the window
  exists. Wrappers that hold raw pointers (`EntityWrapper`,
  `InventoryItemWrapper`, nodes) can go stale once the object despawns; do
  not keep them across long waits.
- Nullable returns arrive as `nil`. `Entity.Target` is `nil` with no
  target; `Excel.GetRow` is `nil` for a missing row.
- .NET lists and arrays are 0-based userdata, not Lua tables. Use
  `list.Count` and `list[i]`, or `for item in luanet.each(list) do`.

### `require`

`package.searchers` is replaced (`LuaCodeSnippets.PackageSearchersSnippet`).
The preload searcher is kept, then:

1. A file searcher, skipped for names ending in `.macro`. It tries the name
   as given, then `<path>\<name>` and `<path>\<name>.lua` for each entry in
   `snd.require.paths`.
2. A macro searcher: strips a trailing `.macro` and loads the stored macro
   with that name (metadata stripped) as a chunk.

So `require("Helpers")` loads a file named `Helpers.lua` from a require path
if one exists and otherwise the stored macro named `Helpers`, and
`require("Helpers.macro")` forces the stored macro. A required chunk is not
wrapped in the entrypoint template; it runs as ordinary Lua and may return a
table. Do not yield at the top level of a required chunk: `require` is a C
function, and Lua cannot yield across it. Functions the module defines can
yield when the script body calls them later. The same holds for metadata
`dependencies`, which are executed with `DoString` before the coroutine
exists.

### .NET types

`import("System.Numerics")` followed by `Vector3(x, y, z)` is the route the
stub generator documents for vectors. Other types follow NLua's usual
pattern:

```lua
-- Unverified pattern; type and assembly names must match the running game.
luanet.load_assembly("Dalamud")
local ConditionFlag = luanet.import_type("Dalamud.Game.ClientState.Conditions.ConditionFlag")
local mounted = Svc.Condition[ConditionFlag.Mounted]
```

No file in the pinned tree exercises `Svc.Condition` from Lua, so treat the
indexer form as untested here. Where a module property answers the question
(`Player.IsBusy`, `Player.Available`, `Player.IsMoving`,
`Entity.Player.IsCasting`, `Entity.Player.IsInCombat`,
`Entity.Player.IsMounted`), use it.

## Modules

Signatures are C# as written upstream; read `uint` and friends as Lua
numbers. `?` marks a value that can be `nil`.

### `Dalamud` and `System`

```
Dalamud.Log(msg)        Dalamud.LogDebug(msg)        Dalamud.LogVerbose(msg)
System.FlashIcon()      System.GetClipboardText()    System.SetClipboardText(text)
```

`Dalamud.Log*` call `msg.ToString()`, so passing `nil` throws. There is no
module function that prints to chat; use `yield("/echo text")` from the
script body or `Engines.Native.Run("/echo text")` from a trigger function.

### `Player`

Properties: `Available`, `IsBusy`, `IsMoving`, `IsInDuty`, `IsOnIsland`,
`CanMount`, `CanFly`, `Revivable`, `IsLevelSynced`, `SyncedLevel`,
`GrandCompany`, `GCRankMaelstrom`, `GCRankTwinAdders`,
`GCRankImmortalFlames`, `FishingBait`, `Entity` (an `EntityWrapper?` for the
local player), `Job` (`JobWrapper`), `FreeCompany`, `Gearset`, `Gearsets`,
`Status` (list of `StatusWrapper`), `Bingo` (Wondrous Tails).

Functions: `GetJob(classJobId)`, `GetGearset(id)`.

`JobWrapper`: `Id`, `Name`, `Abbreviation`, `Level` (unsynced, local player
only), and role booleans `IsCrafter`, `IsGatherer`, `IsTank`, `IsHealer`,
`IsMeleeDPS`, `IsRangedDPS`, `IsMagicDPS`, `IsDPS`, `IsDiscipleOfWar`,
`IsDiscipleOfMagic`, `IsBlu`, `IsLimited`.

`GearsetWrapper`: `IsValid`, `Name`, `ClassJob`, `ItemLevel`,
`GlamourSetLink`, `BannerIndex`, `Items`, `Equip()`, `Update()`.

`StatusWrapper`: `StatusId`, `Param`, `RemainingTime`, `SourceObject`.

### `Entity`

Properties: `Player`, `Target`, `FocusTarget`, `NearestDeadCharacter`,
`NearestOtherCharacter` (each an `EntityWrapper?`).

Functions: `GetEntityByName(name)` (exact name, case-insensitive, first
match in the object table), `GetPartyMember(index)`,
`GetAllianceMember(index)`.

`Entity[n]` returns the object-table entry at index `n`, or `nil`.

`EntityWrapper`: `Type` (`ObjectKind`), `Name`, `EntityId`, `Position`
(`Vector3` with `.X .Y .Z`), `DistanceTo` (from the local player),
`HitboxRadius`, `ContentId`, `AccountId`, `CurrentWorld`, `HomeWorld`,
`CurrentHp`, `MaxHp`, `HealthPercent`, `CurrentMp`, `MaxMp`, `Target`,
`IsCasting`, `IsCastInterruptible`, `IsTargetable`, `IsInCombat`,
`IsHostile`, `IsMounted`, `HuntRank`, `Status`, `FateId`, and methods
`SetAsTarget()`, `SetAsFocusTarget()`, `ClearTarget()`, `Interact()`.

### `Addons`

```
Addons.GetAddon(name) -> AddonWrapper
```

`AddonWrapper`: `Exists` (loaded, visible or not), `Ready` (ECommons
`IsAddonReady`: visible and fully loaded), `GetAtkValue(index)`,
`AtkValues`, `GetNode(id, id, ...)`, `Nodes`.

`NodeWrapper`: `Id`, `IsVisible`, `Text` (read and write, text nodes only),
`NodeType`. `AtkValueWrapper`: `ValueString`.

`GetNode` takes a chain of node ids from the root, as shown by Dalamud's
addon inspector (`/xldata`, Addon Inspector). `GetAtkValue`, `GetNode`,
`AtkValues` and `Nodes` dereference the addon without a null check; test
`Exists` or `Ready` first or the call throws.

There is no callback or click function on the wrapper. Fire UI events with
`yield("/callback Name true 0")` or `yield("/click Name Method")`.

### `Inventory`

Functions: `GetItemCount(itemId)`, `GetHqItemCount(itemId)`,
`GetCollectableItemCount(itemId, minimumCollectability)`,
`GetFreeInventorySlots()`, `GetInventoryItem(itemId)`,
`GetInventoryItemBySlot(container, slot)`,
`GetInventoryContainer(container)`, `GetItemsInNeedOfRepairs(durability)`,
`GetSpiritbondedItems()`.

`Inventory.Inventory1` and any other `InventoryType` member name return that
container through the module's index hook.

`InventoryContainerWrapper`: `Count`, `FreeSlots`, `Items`, `[index]`.

`InventoryItemWrapper`: `ItemId`, `BaseItemId`, `Count`, `IsHighQuality`,
`IsCollectable`, `IsEmpty`, `Condition`, `SpiritbondOrCollectability`,
`GlamourId`, `Container`, `Slot`, `ArmouryContainer`, `LinkedItem`, and
methods `Use()`, `Desynth()`, `MoveItemSlot(container)`,
`MoveItemSlotToSlot(container, slot)`, `LowerQuality()`, `Discard()`,
`SplitItem(quantity)`. `Discard()` calls the game's discard directly with no
confirmation.

### `Actions`

`ExecuteAction(actionId, actionType)` (type defaults to `ActionType.Action`),
`ExecuteGeneralAction(actionId)`, `Teleport(aetheryteId)` (throws for a row
that is not an aetheryte), `CancelCast()`,
`GetActionStatus(actionType, actionId)` (0 means usable, otherwise a
`LogMessage` row id), `GetActionInfo(actionId)` (`ActionWrapper` with
`RecastTime`, `RecastTimeElapsed`, `SpellCooldown`, `AdjustedActionId`,
their `Real*` variants and `GetActionStatus()`), and `LimitBreak`
(`CurrentUnits`, `BarUnits`, `BarCount`).

### `Excel`

```
Excel.GetSheet(name)                 -> sheet wrapper or nil
Excel.GetRow(name, rowId)            -> row wrapper or nil
Excel.GetSubRow(name, rowId, subId)  -> row wrapper or nil
Excel.<SheetName>                    -> same as GetSheet
```

Sheet names are Lumina `Lumina.Excel.Sheets` type names, case-insensitive
(`Item`, `TerritoryType`, `Aetheryte`). A sheet wrapper has `GetRow(id)`,
`GetSubRow(id, subId)`, `Count` and `[id]`. Sub-row sheets only answer
`GetSubRow`; ordinary sheets only answer `GetRow`.

On a row, `row.Name`, `row["Name"]` and `row:GetProperty("Name")` are
equivalent and case-insensitive. Returned values are converted:

- `ReadOnlySeString` becomes a Lua string.
- A typed `RowRef<T>` becomes another row wrapper (or `nil` when invalid),
  so `Excel.GetRow("TerritoryType", id).PlaceName.Name` works.
- An untyped `RowRef` becomes its numeric row id.
- A Lumina `Collection<T>` becomes a wrapper indexed from 0; out-of-range
  indexes give `nil`.

Only properties declared on the row type are visible. `RowId` works; there
is no enumeration of a sheet, so iterate ids from 0 and skip `nil`.

### `Fates`

`CurrentFate`, `GetFateById(id)`, `GetNearestFate()`, `GetActiveFates()`
(nearest first).

`FateWrapper`: `Id`, `Exists`, `InFate`, `State` (`FateState`), `Name`,
`Location`, `Radius`, `Progress`, `Duration`, `StartTimeEpoch`,
`HandInCount`, `IsBonus`, `Rule` (`FateRule`), `Level`, `MaxLevel`,
`FATEChain`, `EventItem`, `IconId`, `DistanceToPlayer`. `GetFateById`
always returns a wrapper; check `Exists` before reading anything else.
`InFate` dereferences the current fate and throws when there is none.

### `Instances`

Each property returns a wrapper:

- `DutyFinder`: `OpenRouletteDuty(id)`, `OpenRegularDuty(cfcId)`,
  `QueueDuty(cfcId)`, `QueueRoulette(id)`, `CancelQueue()`, `QueueState`,
  `GetPenaltyTimeRemainingInMinutes()`, `IsRouletteIncomplete(id)`, and
  settable `IsUnrestrictedParty`, `IsLevelSync`, `IsMinIL`,
  `IsSilenceEcho`, `IsExplorerMode`, `IsLimitedLevelingRoulette`.
- `FriendsList`: `Friends`, `GetFriendByName(name)`.
- `Map`: `IsFlagMarkerSet`, `Flag` (`TerritoryId`, `MapId`, `XFloat`,
  `YFloat`, `Vector2`, `Vector3` with Y fixed at 0,
  `SetFlagMapMarker(...)`).
- `Framework`: `EorzeaTime`, `ClientLanguage`, `Region`.
- `Telepo`: `Teleport(aetheryteId, subIndex)`,
  `GetAetherytePosition(id)`, `IsAetheryteUnlocked(id)`.
- `EnvManager`: `ActiveWeather`, `DayTimeSeconds`, transition times.
- `Buddy`: `BuddyMember`, `CompanionInfo`, `PetInfo`.

### `InstancedContent`

`ContentTimeLeft`, `GetCurrentContentId()`, `GetCurrentContentType()`,
`CanLeaveCurrentContent()`, `LeaveCurrentContent()`, `OceanFishing`,
`OccultCrescent` (`Events`, `MKDData`, `OccultCrescentState`,
`ChainTargets`), `PublicInstance` (`TerritoryTypeId`, `InstanceId`,
`IsInstancedArea`).

### `Quests`

`IsQuestAccepted(id)`, `IsQuestComplete(id)`, `GetQuestSequence(id)`,
`GetAcceptedQuests()`, `IsLeveAccepted(id)`,
`GetMonsterNoteRankInfo(index)`. Ids are the 16-bit quest id.

### `Engines`

```
Engines.Run(text)          -- native if text starts with "/", otherwise Lua
Engines.Native.Run(text)
Engines.NLua.Run(text)
```

These start a new temporary macro on a background task and return at once.
Nothing waits for it and errors are only logged. They are the way to send a
command from a trigger function, which cannot yield. In the script body,
prefer `yield(text)`, which waits.

### `Config`

Per-script settings declared in metadata. See
[metadata.md](metadata.md#configs).

## `IPC`: other plugins

`IPC.<Name>` exists for every class under `External/` that derives from
`IPC`. `<Name>` is the class's `Name` string, which is not always the class
or file name:

| `IPC.` name | Source file | Notes |
| --- | --- | --- |
| `vnavmesh` | `Navmesh.cs` | `IsReady`, `BuildProgress`, `Reload`, `Rebuild`, `Pathfind`, `PathfindAndMoveTo(Vector3, fly)`, `PathfindInProgress`, `MoveTo(List<Vector3>, fly)`, `Stop`, `IsRunning`, `NearestPoint`, `PointOnFloor` |
| `Lifestream` | `Lifestream.cs` | `ExecuteCommand`, `IsBusy`, `Abort`, world, aethernet, housing and instance calls |
| `AutoRetainer` | `AutoRetainer.cs` | MultiMode, `IsBusy`, `Relog`, offline character data, `SetSuppressed` |
| `Questionable` | `Questionable.cs` | `IsRunning`, quest state and priority calls |
| `BossMod` | `BossMod.cs` | Preset calls only (`Presets.*` gates) |
| `WrathCombo` | `Wrath.cs` | `Register(scriptName)` returns the lease; state and option setters take it |
| `InventoryTools` | `AllaganTools.cs` | Allagan Tools counts, filters and craft lists |
| `Artisan`, `AutoDuty`, `AutoHook`, `ARDiscard`, `Deliveroo`, `Dropbox`, `Henchman`, `PandorasBox`, `TextAdvance`, `Tippy`, `visland`, `YesAlready` | same-named files (`DeliverooIPC.cs`, `PandorasBoxIPC.cs`) | |
| `Automaton` | `CBT.cs` | `IsTweakEnabled`, `SetTweakState` |

Module-level helpers:

```
IPC.IsInstalled(name)        -> bool
IPC.GetAvailablePlugins()    -> string[]
```

How it is wired (`Core/Interfaces/IIPC.cs`, `LuaMacro/Modules/IPCModule.cs`):

- Each class is constructed once per Lua state and bound with
  `EzIPC.Init(this, Name)`. A field tagged `[EzIPC]` and `[LuaFunction]`
  becomes `IPC.<Name>.<Field>`; an `[EzIPC("Prefix.%m")]` field binds the
  gate `<Name>.Prefix.<Field>` but is still exposed without the prefix. For
  example `IPC.vnavmesh.IsReady()` calls the gate `vnavmesh.Nav.IsReady`.
- The exposed set is fixed by these files. There is no generic "call any
  gate by name" function in the module. A gate that upstream did not wrap
  is not reachable through `IPC`.
- Fields are assigned to the Lua table only when non-null. If binding a
  gate failed, that entry is absent and calling it is a Lua
  "attempt to call a nil value" error. What a bound entry does when the
  target plugin is not loaded is decided by ECommons EzIPC, which is not
  part of this snapshot; declare the plugin in `plugin_dependencies` so the
  script refuses to start without it.
- `IPC.IsInstalled(name)` is true when an installed, loaded plugin has
  `p.Name == name`, where `name` is the `IPC.` name above. It compares
  against the plugin's `Name` property, while the scheduler's
  `plugin_dependencies` check compares `InternalName`. For a plugin whose
  two names differ, `IsInstalled` can report false while the dependency
  check passes. Confirm on the installed build before gating logic on it.
- Signatures are whatever the wrapped plugin exposed when upstream wrote the
  file. They can lag the plugin. Take the contract (what a call means, when
  it is refused, what to wait for) from the sibling skills `vnavmesh`,
  `lifestream`, `autoretainer`, `questionable`, `bossmod`, `wrathcombo` and
  `allagantools`, and take the callable name and parameter list from the
  `External/` file.
- Arguments must convert to the delegate's .NET types. A `Vector3`
  parameter needs a real `System.Numerics.Vector3`, for example an entity's
  `Position` or one built after `import("System.Numerics")`. A
  `List<Vector3>` or an `Action` parameter cannot be produced from a plain
  Lua table or function without CLR interop.
- Calls that return `Task<T>` (`IPC.vnavmesh.Pathfind`) hand back the task
  object. Poll `task.IsCompleted` between yields and read `task.Result`;
  never block on it, since the script runs on the framework thread.

Two plugins also implement `IDisableable` (`TextAdvance`, `YesAlready`) and
are the only valid values for `plugins_to_disable` in metadata.

Trigger events from other plugins are limited to
`OnAutoRetainerCharacterPostProcess`, which is wired through
`AutoRetainerAPI` directly, not through `IPC`.

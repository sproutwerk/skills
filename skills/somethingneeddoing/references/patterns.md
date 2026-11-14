# Script patterns

Lua snippets here are written against the API in [lua-api.md](lua-api.md)
at SomethingNeedDoing `v15.13`
(`527ca83892506c74fea4d5e07b9b9436e50ef95d`). None of them has been run in
the game. They show structure; check names against the Help window's Lua
tab on the installed build.

## The rule everything else follows

Lua between two yields runs inside one game frame, and Stop is only honored
at a yield. Every loop that waits for the game must yield on each pass:

```lua
local function sleep(seconds)
    yield("/wait " .. seconds)
end
```

Never write `while not ready() do end`, never `yield()` without a string
(it ends the script), and never block on a .NET task.

## Wait for a condition with a deadline

```lua
local function waitUntil(predicate, timeoutSeconds, label)
    local deadline = os.time() + timeoutSeconds
    while not predicate() do
        if os.time() >= deadline then
            error("timed out waiting for " .. label)
        end
        sleep(0.1)
    end
end
```

`os.time()` has one-second resolution, which is enough for timeouts.
`os.clock()` measures CPU time and is the wrong clock. Every wait gets a
deadline; an unbounded wait is the usual way a script ends up "running"
forever in the status window.

Decide per wait whether a timeout is fatal. `error(...)` ends the script,
runs the cleanup function and prints an error to chat. To continue, return
a boolean instead:

```lua
local function waitFor(predicate, timeoutSeconds)
    local deadline = os.time() + timeoutSeconds
    while not predicate() do
        if os.time() >= deadline then return false end
        sleep(0.1)
    end
    return true
end
```

## Wait for the player to be free

```lua
waitUntil(function() return Player.Available and not Player.IsBusy end, 30, "player")
```

`Player.Available` and `Player.IsBusy` come from ECommons `Player`. After a
zone change, wait for `Player.Available` before touching `Entity.Player` or
inventory.

## Addon readiness and callbacks

```lua
local function addonReady(name)
    return Addons.GetAddon(name).Ready
end

waitUntil(function() return addonReady("SelectYesno") end, 10, "SelectYesno")
yield("/callback SelectYesno true 0")
waitUntil(function() return not Addons.GetAddon("SelectYesno").Exists end, 5, "dialog close")
```

- `Ready` is "visible and loaded". `/waitaddon Name <maxwait.10>` does the
  same wait natively but only warns on timeout unless Stop on error is set.
  The Lua form lets the script decide.
- `/callback` silently does nothing when the addon is missing or not ready.
  Wait for `Ready` first and confirm the result afterwards (the window
  closed, the next window opened, a node's text changed).
- Callback values are found by watching the window with a callback logger
  such as the one the wiki describes, not by guessing. Firing the wrong
  values at a window can crash the client.
- Read text with `Addons.GetAddon("Name"):GetNode(1, 5, 2).Text` after the
  `Ready` check. Node id chains change between game patches.
- Prefer `/click Addon Method` when ECommons has an `AddonMaster` entry for
  the window; it validates the addon and names the button.
- Plugins that auto-answer dialogs (YesAlready, TextAdvance) race with a
  script that wants to answer them. List them under `plugins_to_disable`
  so SND pauses them for the run.

Reacting to a dialog whenever it appears, without polling, is a
function-level trigger:

```lua
function OnAddonEvent_SelectYesno_PostSetup()
    Engines.Native.Run("/callback SelectYesno true 0")
end
```

The handler cannot wait, so this suits single-step answers only.

## Throttling

```lua
local lastRun = {}
local function throttle(key, seconds)
    local now = os.time()
    if lastRun[key] and now - lastRun[key] < seconds then return false end
    lastRun[key] = now
    return true
end

function OnUpdate()
    if not throttle("status", 5) then return end
    Dalamud.Log("still alive")
end
```

`OnUpdate` handlers run every frame while the script is alive. Do the
cheapest possible test first and return. For periodic work a main loop with
`sleep(1)` is simpler than an `OnUpdate` handler.

## Stopping cleanly

Put the work in a function so every exit path is a `return`, and release
what the script started in a cleanup function. Cleanup sees globals only.

```lua
local navStarted = false          -- local: invisible to cleanup
NavStarted = false                -- global: visible to cleanup

function OnStop()
    if NavStarted and IPC.vnavmesh.IsRunning() then
        IPC.vnavmesh.Stop()
    end
end

local function main()
    -- ... work, returning early on failure ...
end

main()
```

`OnStop` runs after a normal return, after an error and after the user
presses Stop. It can yield commands. Things worth undoing there: movement
(`IPC.vnavmesh.Stop()`), held keys (`/release KEY` for every `/hold`),
Lifestream tasks, leases or suppression flags set on other plugins.
`plugins_to_disable` restores itself.

To end from deep inside nested calls, `error()` with a message is the
supported route; the message reaches the log and cleanup still runs.

## Error handling

```lua
local ok, err = pcall(function()
    return IPC.Lifestream.AethernetTeleport("The Aftcastle")
end)
if not ok then
    Dalamud.Log("Lifestream call failed: " .. tostring(err))
end
```

- `pcall` catches Lua errors and exceptions thrown by the .NET method being
  called. It does not catch a failing yielded command: that marks the
  script as errored from the outside and the script simply stops at that
  yield.
- Check for `nil` before indexing: `Entity.Target`,
  `Entity.GetEntityByName`, `Excel.GetRow` and friends return `nil` freely,
  and indexing `nil` is the most common script error.
- A yielded command that can legitimately miss should carry its own
  modifier, for example `/target "Name" <errorif.targetnotfound>` when a
  miss must stop the script, or none when the script checks
  `Entity.Target` afterwards.

## Logging and user messages

| Goal | Call |
| --- | --- |
| Dalamud log (`/xllog`), Information | `Dalamud.Log(text)` or `print(a, b)` |
| Debug or Verbose level | `Dalamud.LogDebug(text)`, `Dalamud.LogVerbose(text)` |
| Chat echo from the script body | `yield("/echo " .. text)` |
| Chat echo from a trigger handler | `Engines.Native.Run("/echo " .. text)` |
| Taskbar flash when done | `System.FlashIcon()` |

`Dalamud.Log(nil)` throws; wrap values in `tostring`. `/echo <se.1>` style
sound tags pass through to the game unchanged.

## Coordinating with other plugins

The `IPC` table gives call names; it does not give semantics. For what a
call means, when it is refused and what to wait for, use the sibling skill
for that plugin: `vnavmesh`, `lifestream`, `autoretainer`, `questionable`,
`bossmod`, `wrathcombo`, `allagantools`. The shape is the same for all of
them:

1. Declare the plugin in `plugin_dependencies` by its internal name so the
   script refuses to start without it.
2. Check readiness (`IPC.vnavmesh.IsReady()`, `IPC.WrathCombo.IPCReady()`,
   `IPC.InventoryTools.IsInitialized()`), with a deadline.
3. Make the request and keep its return value. Many gates return `false`
   or nothing when they refuse.
4. Wait for the plugin's busy signal to clear, with a deadline.
5. Verify the game state the next step needs. "Not busy" does not prove the
   request ran.
6. Undo in cleanup what the script switched on.

### vnavmesh

```lua
local nav = IPC.vnavmesh
waitUntil(function() return nav.IsReady() end, 120, "navmesh build")

local npc = Entity.GetEntityByName("Summoning Bell")
if not npc then error("target not in range") end

NavStarted = nav.PathfindAndMoveTo(npc.Position, false)
if not NavStarted then error("vnavmesh refused the request") end

waitUntil(function()
    return not nav.PathfindInProgress() and not nav.IsRunning()
end, 60, "arrival")

local again = Entity.GetEntityByName("Summoning Bell")
if not again or again.DistanceTo > 4 then error("did not arrive") end
```

Re-fetch the entity for the distance check instead of holding the wrapper
across the walk. The yielded command form (`yield("/vnav moveto x y z")`)
goes through the game's chat parser and returns nothing; the IPC form gives
a result to check. Mesh readiness, flying and arrival tolerances are in the
`vnavmesh` skill.

### Lifestream

```lua
local li = IPC.Lifestream
waitUntil(function() return not li.IsBusy() end, 30, "Lifestream idle")
li.ExecuteCommand("gc")
sleep(1)
waitUntil(function() return not li.IsBusy() end, 120, "Lifestream travel")
waitUntil(function() return Player.Available end, 30, "zone load")
-- verify: li.GetRealTerritoryType() is the territory you expected
```

`ExecuteCommand` takes the text after `/li`. A refused request leaves
`IsBusy()` false, so always verify the destination. Command grammar and the
gates that skip the busy check are in the `lifestream` skill.

### AutoRetainer

Two integration points. A script can drive AutoRetainer
(`IPC.AutoRetainer.IsBusy()`, `SetMultiModeEnabled(bool)`,
`AbortAllTasks()`, `SetSuppressed(bool)`), or AutoRetainer can drive a
script through the macro-level trigger
`OnAutoRetainerCharacterPostProcess` described in
[metadata.md](metadata.md#macro-level). A post-process script must
terminate, because AutoRetainer waits for it; give every wait a deadline
and let errors end the script. Details of MultiMode and post-processing are
in the `autoretainer` skill.

### Questionable, BossMod, WrathCombo, Allagan Tools

- `IPC.Questionable.IsRunning()` and the quest-state calls take quest ids
  as strings. Starting and stopping Questionable is done with its chat
  commands through `yield`; see the `questionable` skill.
- `IPC.BossMod` wraps the preset gates only (`SetActive`, `ClearActive`,
  `GetActive`, ...). Clear the preset in cleanup. See the `bossmod` skill,
  and note that BossMod forks differ.
- `IPC.WrathCombo.Register(name)` returns a lease id that the setters
  require; call `ReleaseControl(lease)` in cleanup. The `SetResult` enum is
  registered as a global. See the `wrathcombo` skill for lease rules.
- `IPC.InventoryTools` is Allagan Tools. Character ids are 64-bit; get the
  current one from `CurrentCharacter()` and pass it back unchanged without
  doing arithmetic on it in Lua. See the `allagantools` skill.

## Native macro patterns

```
/target "Summoning Bell" <errorif.targetnotfound>
/interact <wait.1>
/waitaddon "RetainerList" <maxwait.10> <errorif.actiontimeout>
```

- Put `<errorif...>` on the lines where continuing after a miss would act on
  the wrong window or target.
- `/waitaddon` before every `/click` or `/callback`.
- Use `<wait.A-B>` instead of a fixed wait when a loop repeats many times.
- A native macro has no variables or branches. When a step needs "if",
  convert the macro to Lua and yield the same lines.

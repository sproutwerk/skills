# Example scripts

Four complete scripts written against SomethingNeedDoing `v15.13`
(`527ca83892506c74fea4d5e07b9b9436e50ef95d`). They were derived from the
bundled source and have not been compiled, loaded into the plugin or run in
the game. Treat them as starting structure: check each API name in the Help
window's Lua tab, each addon name and callback value in the addon inspector,
and each action name in the client's language before use.

Game data in the examples (NPC names, action names, addon names, callback
values) is illustrative and was not checked against a live client.

## 1. Native crafting macro

A native macro is a list of `/` lines. No comments, no variables.

```
/ac "Reflect" <wait.3>
/ac "Manipulation" <wait.2>
/ac "Innovation" <wait.2>
/ac "Preparatory Touch" <wait.3>
/ac "Preparatory Touch" <wait.3>
/ac "Tricks of the Trade" <condition.good,excellent> <wait.2>
/ac "Byregot's Blessing" <wait.3>
/ac "Veneration" <wait.2>
/ac "Groundwork" <wait.3>
/ac "Groundwork" <wait.3>
/echo Craft finished <se.1>
```

To repeat it, set "Crafting loop" and a count in the macro's settings panel
(metadata keys `crafting_loop` and `craft_loop_count`); the scheduler then
wraps these lines in the craft-loop steps described in
[native-macros.md](native-macros.md#loops). Notes:

- With "Skip craft actions when not crafting" on (the default), the lines
  after the craft completes are skipped along with their waits.
- The `<condition...>` line runs only on those Synthesis conditions and
  costs no wait otherwise.
- `<se.1>` is a game sound tag, not an SND modifier, and passes through.

## 2. Lua: walk to an object, interact, wait for a window

Uses metadata, a config, vnavmesh through `IPC`, deadlines on every wait and
a cleanup function.

```lua
--[=====[
[[SND Metadata]]
author: example
version: 1.0.0
description: Walk to a named object, interact with it and wait for a window.
plugin_dependencies:
- vnavmesh
configs:
  targetName:
    default: Summoning Bell
    description: Exact name of the object to walk to
  expectedAddon:
    default: RetainerList
    description: Window that should open after interacting
  walkTimeout:
    default: 60
    type: int
    min: 5
    max: 300
    description: Seconds allowed for the walk
[[End Metadata]]
--]=====]

NavStarted = false   -- global so the cleanup function can read it

function OnStop()
    if NavStarted and IPC.vnavmesh.IsRunning() then
        IPC.vnavmesh.Stop()
    end
end

local function sleep(seconds)
    yield("/wait " .. seconds)
end

local function waitUntil(predicate, timeoutSeconds, label)
    local deadline = os.time() + timeoutSeconds
    while not predicate() do
        if os.time() >= deadline then
            error("timed out waiting for " .. label)
        end
        sleep(0.1)
    end
end

local function main()
    local targetName = Config.GetString("targetName")
    local expectedAddon = Config.GetString("expectedAddon")
    local walkTimeout = Config.GetInt("walkTimeout")
    local nav = IPC.vnavmesh

    waitUntil(function() return Player.Available and not Player.IsBusy end, 30, "player")
    waitUntil(function() return nav.IsReady() end, 120, "navmesh")

    local object = Entity.GetEntityByName(targetName)
    if not object then
        error("no object named '" .. targetName .. "' is loaded")
    end

    if object.DistanceTo > 4 then
        NavStarted = nav.PathfindAndMoveTo(object.Position, false)
        if not NavStarted then
            error("vnavmesh refused the path request")
        end
        waitUntil(function()
            return not nav.PathfindInProgress() and not nav.IsRunning()
        end, walkTimeout, "arrival")
        NavStarted = false
    end

    -- Re-fetch: the wrapper from before the walk may be stale.
    object = Entity.GetEntityByName(targetName)
    if not object or object.DistanceTo > 5 then
        error("did not arrive at '" .. targetName .. "'")
    end

    object:SetAsTarget()
    sleep(0.3)
    if not Entity.Target then
        error("could not target '" .. targetName .. "'")
    end
    Entity.Target:Interact()

    waitUntil(function() return Addons.GetAddon(expectedAddon).Ready end, 10, expectedAddon)
    Dalamud.Log("opened " .. expectedAddon)
end

main()
```

Points to check on a real build: the internal name `vnavmesh` for
`plugin_dependencies` and the interaction range for the chosen object.
`PathfindAndMoveTo` is declared `Func<Vector3, bool, bool>` in
`External/Navmesh.cs`, so the boolean result matches the wrapper; what a
`false` means is the `vnavmesh` skill's subject, as are the arrival rules.

## 3. Lua: react to events with trigger functions

A script that stays alive, counts zone changes, answers a chat keyword and
stops on its own after a time limit. Handlers cannot yield, so they set
flags or use `Engines.Native.Run`.

```lua
--[=====[
[[SND Metadata]]
author: example
version: 1.0.0
description: Demonstrates function-level triggers and a keep-alive loop.
configs:
  minutes:
    default: 30
    type: int
    min: 1
    max: 600
    description: How long to stay running
function_chat_filters:
  OnChatMessage_Ping:
    channels: [Party]
    message_contains: ping
[[End Metadata]]
--]=====]

ZoneChanges = 0
PendingReply = false
Finished = false

function OnTerritoryChange()
    ZoneChanges = ZoneChanges + 1
    Dalamud.Log("territory is now " .. tostring(TriggerData.territoryType))
end

function OnChatMessage_Ping()
    -- No waiting here. Hand the work to the main loop.
    PendingReply = true
end

function OnLogout()
    Finished = true
end

function OnEnd()
    Dalamud.Log("zone changes seen: " .. tostring(ZoneChanges))
end

local deadline = os.time() + Config.GetInt("minutes") * 60

while not Finished and os.time() < deadline do
    if PendingReply then
        PendingReply = false
        yield("/p pong <wait.1>")
    end
    yield("/wait 0.5")
end
```

Notes:

- Every handler is a global `function Name(`; a `local function` would be
  matched by the name scan and never found at call time.
- Helper names avoid the prefixes `OnUpdate`, `OnLogin`, `Stop`, `End` and
  so on, which the plugin would treat as handlers or cleanup.
- `OnEnd` is one of the cleanup names. It runs after the loop ends, after an
  error and after Stop.
- The loop is what keeps the handlers registered. Remove it and the script
  finishes immediately.
- An immediate, unawaited reaction inside a handler would be
  `Engines.Native.Run("/p pong")`.

## 4. Lua: answer a dialog with `/callback`, verified

Wait for a confirmation window, read its text, answer it and confirm it
closed. The node id chain and callback values are placeholders to replace
with what the addon inspector and a callback logger show for the window in
question.

```lua
--[=====[
[[SND Metadata]]
author: example
version: 1.0.0
description: Answer one confirmation dialog if its text matches.
plugins_to_disable:
- YesAlready
configs:
  expectedText:
    default: Proceed
    description: Text the dialog must contain before it is confirmed
[[End Metadata]]
--]=====]

local ADDON = "SelectYesno"
local TEXT_NODE = { 1, 2 }          -- placeholder node id chain
local CONFIRM = "true 0"            -- placeholder callback values

local function waitFor(predicate, timeoutSeconds)
    local deadline = os.time() + timeoutSeconds
    while not predicate() do
        if os.time() >= deadline then return false end
        yield("/wait 0.1")
    end
    return true
end

local function main()
    local dialog = Addons.GetAddon(ADDON)

    if not waitFor(function() return dialog.Ready end, 15) then
        Dalamud.Log(ADDON .. " did not open; nothing to do")
        return
    end

    local node = dialog:GetNode(table.unpack(TEXT_NODE))
    local text = node and node.Text or ""
    local expected = Config.GetString("expectedText")

    if not string.find(text, expected, 1, true) then
        error("unexpected dialog text: " .. tostring(text))
    end

    yield("/callback " .. ADDON .. " " .. CONFIRM)

    if not waitFor(function() return not dialog.Exists end, 5) then
        error(ADDON .. " is still open after the callback")
    end
end

main()
```

Notes:

- `plugins_to_disable` pauses YesAlready for the run so it does not answer
  the dialog first; SND restores it afterwards.
- `/callback` is silent when the window is not ready, which is why the
  script waits before and verifies after.
- `dialog` is an `AddonWrapper`, which looks the window up by name on each
  access and is safe to keep. The node wrapper is not; fetch it right
  before reading.
- The error path stops the script without answering. That is deliberate: a
  script that confirms a dialog it did not expect can sell, discard or
  spend something.

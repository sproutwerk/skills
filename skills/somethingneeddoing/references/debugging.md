# Debugging SND scripts

Behavior described here is read from the source at commit
`527ca83892506c74fea4d5e07b9b9436e50ef95d` (`v15.13`). Messages are quoted
from that source; none was reproduced in a running game.

## Where to look

| Place | What it shows |
| --- | --- |
| Dalamud log (`/xllog`) | Everything. Chat only gets a one-line summary; the exception, Lua message and stack are in the log. SND lines start with `[FileName]`, for example `[NLuaMacroEngine]` or `[MacroScheduler]`. |
| Log level | Skips and trigger activity are logged at Debug and Verbose. Raise the plugin's log level in `/xllog` to see `Condition skip`, `Not crafting skip`, `Skipping function ... not yet defined`, trigger firings and the parser's `Extracted modifiers` lines. |
| Chat | Errors go to the "ErrorChatType" channel (default Urgent) prefixed `[SND]`; `<echo>` output and loop counts go to "ChatType" (default Debug). If nothing appears, check those two settings. |
| Macro Status window (`/snd status`) | Each running macro and its state, including child macros spawned by `yield`. A script stuck in a wait shows as `Running` with a `/wait` child flickering under it. |
| Help window | Lua tab: the live API of the installed build. Commands tab: native commands and modifiers. Clicks tab: valid `/click` targets. Keys tab: `/send` key names, with held keys highlighted. |
| Macro settings panel | Parsed metadata, config values, and "Registered Trigger Events" for a running script. If a handler is not listed there, it was not discovered. |
| Addon inspector (`/xldata`) | Addon names, node id chains and AtkValues for `Addons.GetAddon`. |

`print(...)` and `Dalamud.Log(...)` write to the Dalamud log, not to chat.

## Symptom to cause

### Nothing happens when starting

| Check | Cause |
| --- | --- |
| Chat says `Cannot run <name>. The following plugins need to be installed: ...` | A `plugin_dependencies` entry does not match the `InternalName` of a loaded plugin. Internal names are case-sensitive and often differ from display names. |
| Chat says `Cannot run <name>. One or more of its configs failed to validate.` | A config's `validation_pattern` failed. See the validation caveat in [metadata.md](metadata.md#configs). |
| Chat says `... dependencies are not available: ...` | A script dependency could not be read or downloaded. The message carries the source and the reason. |
| No message at all, started with `/snd run` | The name did not match a stored macro; `/snd run` is silent on a miss. |
| No message, macro-level trigger or rapid restarts | The macro is already running (start is skipped), or it failed validation in the last 5 seconds (start is suppressed). |
| The script appears and vanishes at once | It finished. A script that only defines trigger handlers needs a loop to stay alive. Also check for a bare `yield()`, which ends the script. |

### The game freezes

A Lua loop without a yield. The script runs on the framework thread and
Stop is only checked between yields, so the client hangs until the loop
ends. There is no recovery from inside the game; kill the client, then add
a `yield("/wait 0.1")` to the loop. Blocking calls have the same effect:
`task:Wait()`, `task.Result` on an unfinished task, large synchronous file
or network work.

### Lua errors

Chat shows `Error executing Lua function for macro <id>`; the log has the
real message.

| Log message contains | Meaning |
| --- | --- |
| `attempt to call a nil value (global 'X')` | `X` is not defined. A legacy function ([migration.md](migration.md)), a typo, or a helper file that was not loaded. |
| `attempt to call a nil value (field 'X')` | The module exists but has no such member, or an `IPC` entry was not bound. Compare with the Help Lua tab. |
| `attempt to index a nil value` | A nullable accessor returned `nil`: no target, entity not found, row missing, addon wrapper member used on a closed addon. |
| `Lua Macro yielded a non-string value [...]` | `yield` was given a number, table or boolean. |
| `attempt to yield from outside a coroutine` | `yield` inside a function-level trigger handler, or at the top level of a dependency. |
| `attempt to yield across a C-call boundary` | `yield` at the top level of a `require`d chunk, or inside a callback invoked by a C function such as `table.sort` or `string.gsub`. |
| `Failed to load Lua script: No function returned` | The script text broke the entrypoint wrapper, typically an unbalanced `end`. |
| A .NET exception name (`NullReferenceException`, `InvalidCastException`, ...) | A module call hit game state it does not guard, or an argument had the wrong type. `Config.GetInt` and friends throw `InvalidCastException` with the offending value in the message. |

Line numbers are relative to the wrapped script: subtract 4, then add the
number of lines the metadata block occupied if the error is below it.

Method-call syntax errors are easy to miss. `Entity.Target.Interact()` (dot)
passes no object and fails inside NLua; wrapper methods need a colon.
Module functions are the opposite and take a dot.

### A yielded command "did nothing"

- `/callback` skips silently when the addon is absent or not ready, and
  `/target` skips silently on no match. Add `<errorif.addonnotfound>` or
  `<errorif.targetnotfound>` while debugging to turn the skip into an
  error.
- `/ac` on a crafting action is skipped when not crafting, at maximum
  progress, or at maximum quality for a quality-only action. The log has a
  Debug line naming the skip.
- `/waitaddon` timed out and continued. The log has a Warning:
  `Addon '<name>' did not appear within the timeout period, continuing...`.
- The text did not start with `/`. That is a syntax error in the child
  macro, which sets the parent to `Error` and stops the script at that
  yield with no Lua-side error.
- An unknown command name is passed to the game, which answers in chat that
  the command does not exist.
- Text with SND modifier look-alikes is consumed. A chat line containing
  `<wait.1>` or `<echo>` loses that tag to the parser.

### Native macro errors

Chat shows `Error executing macro command` and `Macro execution failed`
without the line. The log entry `Error executing macro <name>` carries the
exception:

| Message | Cause |
| --- | --- |
| `Syntax error in command: ...` | A line without a leading `/`, a bad count on `/loop` or `/gate`, `/wait` with a non-number, a malformed `/callback`. The macro is rejected before any line runs. |
| `Invalid click command: ...` | `/click` with fewer than two words. |
| `Addon <name> not found` | `/click` with the window closed. |
| `Type <name> not found`, `Could not find method ...` | No `AddonMaster` entry or method of that name and argument count. |
| `Condition wait timed out` | `/require`, `/requirerepair` or `/requirespiritbond` did not pass in time. |
| `Item not found: ...`, `Key item not found: ...` | The name is not an exact sheet name in the client language. |
| `/recipe cannot be used while the Synthesis window is open.` | As stated. |
| `Could not find target`, `You do not have that item` | A miss on a line carrying `<errorif...>` or with Stop on error enabled. |

### Trigger handlers never fire

1. Is the script still running? Handlers are tied to a live script.
2. Is the function global and declared as `function Name(`? `local
   function` and `M.Name = function` are not found.
3. Does the name start with the exact event name? Addon handlers need
   `OnAddonEvent_<Addon>_<Event>` with a valid `AddonEvent` member.
4. Is the script paused? Handlers only run in state `Running`.
5. For chat handlers, does `function_chat_filters` use the exact function
   name, and do the channel names parse as `XivChatType`?
6. Verbose log shows `Requesting function execution for <name>` when the
   event reached the script and `Skipping function ...` with the reason
   when it was dropped.

Macro-level triggers that never fire: the name under `triggers` was not a
`TriggerEvent` member and was dropped silently, or the metadata block did
not parse (`Failed to parse metadata YAML` in the log), or for
`OnAddonEvent` the `addon_event` map is missing.

### The script will not stop

- It is between yields in a long computation: see the freeze section.
- It stopped, and a cleanup function is running. Cleanup commands appear as
  `<name> - Cleanup Command` in the status window.
- A child macro started with `Engines.Run` or `/runmacro` is a separate
  macro and keeps running. Stop it by name or use `/snd stop all`.
- A held key from `/hold` stays down after Stop. Send `/release KEY`.
- Another plugin is still acting on the script's last request. Stopping SND
  does not stop vnavmesh or Lifestream; do that in a cleanup function.

### Stale or odd values

- A wrapper kept across a zone change or despawn points at freed memory.
  Re-fetch after waits.
- `Config.Get` returned an empty string: the name is not declared in
  `configs`, or differs in case.
- Config edits in the script text did not change the value: user values
  are kept by name across re-parses. Reset the value in the settings panel
  or call `Config.Reset(name)`.
- A Git macro reverted an edit: it updated from its repository. Clear the
  Git data to keep local changes.
- Remote dependency changes are not picked up: the cache lasts one day.
  Delete `cache/dependencies` in the plugin's config directory to force a
  download.

## Reporting what was checked

State which of these a diagnosis rests on: the pinned source, the log from
the user's session, or a run in the game. A fix reasoned from source alone
should say so, since the installed build can differ from the pin and no
part of this skill was validated in a live client.

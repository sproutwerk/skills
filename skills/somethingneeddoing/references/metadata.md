# Metadata, configs, triggers and dependencies

Derived from `Core/MetadataParser.cs`, `Core/MacroMetadata.cs`,
`Core/MacroConfigItem.cs`, `Core/MacroFunctions.cs`,
`LuaMacro/Modules/ConfigModule.cs`, `Managers/TriggerEventManager.cs`,
`Managers/TriggerEventSources.cs` and `Core/Macros/Dependencies/` at commit
`527ca83892506c74fea4d5e07b9b9436e50ef95d` (`v15.13`), with the upstream
wiki pages under [source/SomethingNeedDoing.wiki/](source/SomethingNeedDoing.wiki/)
as secondary evidence. Where the wiki and the code disagree, this file
follows the code and says so.

## The metadata block

Metadata is stored with the macro in the plugin config and can be edited in
the macro's settings panel. A script can also carry it in its text, which
is what makes a GitHub-hosted script self-describing. The parser looks for
this shape (`MetadataParser.MetadataBlockRegex`):

```lua
--[=====[
[[SND Metadata]]
author: someone
version: 1.0.0
description: What the script does
[[End Metadata]]
--]=====]
```

- The opening must be `--[=====[` (exactly five `=`) or `/*`, and the
  close `]=====]` or `*/`. The markers `[[SND Metadata]]` and
  `[[End Metadata]]` are matched case-insensitively.
- The body between the markers is YAML, read by YamlDotNet with the
  underscored naming convention. Unknown keys are ignored.
- The block is removed before the script is run
  (`ContentSansMetadata`), which shifts Lua line numbers.
- Fields present in the block override the stored metadata; absent fields
  keep their stored value. A YAML error is logged as
  `Failed to parse metadata YAML` and leaves the stored metadata as it was.
- The block is re-read when the plugin loads and about half a second after
  the content changes.

"Write Metadata to Content" in the settings panel serializes the stored
metadata back into the text. For Lua it uses the comment form above. For
native macros `GetCommentSyntax` returns empty delimiters at the pin, so the
written block has no `/* */` around it and does not match the parser's
pattern; keep native-macro metadata in the settings panel, or wrap the block
in `/* */` by hand and confirm it is stripped.

### Keys

| Key | Type | Meaning |
| --- | --- | --- |
| `author`, `version`, `description` | scalar | Shown in the UI. `version` defaults to `1.0.0`. |
| `triggers` | list of names | Macro-level trigger events. Names are `TriggerEvent` members, case-insensitive; unknown names are dropped without a message. |
| `plugin_dependencies` | list | Plugin `InternalName`s that must be loaded or the macro refuses to start. |
| `plugins_to_disable` | list | Plugins to turn off while the macro runs and back on when it ends or is paused. Only `TextAdvance` and `YesAlready` are supported at the pin; anything else logs `Plugin <name> is not registered as disableable`. |
| `dependencies` | list of maps | Other scripts to load first. Only `source` is read. |
| `configs` | map | User-editable settings. |
| `crafting_loop` | bool | Native macros: wrap in the craft loop. |
| `craft_loop_count` | int | Iterations; `-1` for unlimited. |
| `addon_event` | map | `addon_name` and `event_type` for a macro-level `OnAddonEvent`. |
| `chat_message_filter` | map | Filter for a macro-level `OnChatMessage`. |
| `function_chat_filters` | map of maps | Filters for function-level `OnChatMessage*` handlers, keyed by function name. |

## Configs

```yaml
configs:
  loops: 5                      # scalar form: default value, type inferred
  itemName:
    default: Iron Ore
    description: Item to craft
    section: Crafting
  quality:
    default: 100
    type: int
    min: 0
    max: 100
  job:
    description: Class to use
    is_choice: true
    choices: ["Culinarian", "Alchemist"]
  targets:
    description: Names to visit
    is_choice: false
    choices: []
```

Per-entry keys read by `ParseConfigs`: `default`, `description`, `section`,
`type`, `min`, `max`, `validation_pattern`, `validation_message`,
`choices`, `is_choice`. The wiki's examples also show `required`; the
parser does not read it.

- `type` accepts `bool`/`boolean`, `int`/`integer`/`number`,
  `float`/`double`, `list`; anything else is a string. Without `type` it is
  inferred from `default` (a string that parses as a bool or number counts
  as one).
- An entry with `choices` and no `default` is a list. With
  `is_choice: true` the UI shows a dropdown and the value is one string;
  with `is_choice: false` the user builds a list of strings.
- The user's values live in the plugin config, keyed by name. Re-parsing
  the block, including after a Git update, keeps the value of every config
  whose name still exists.

Reading and writing from Lua (`Config` is created per script):

```
Config.Get(name, default?)   Config.GetInt(...)   Config.GetFloat(...)
Config.GetBool(...)          Config.GetString(...)
Config.Set(name, value)      Config.Reset(name)   Config.GetDefault(name)
Config.Exists(name)          Config.GetAllNames() Config.GetType(name)
Config.GetDescription(name)
```

- `Get` on an undeclared name returns the supplied default, or an empty
  string when none is given. It never returns `nil`, so test with
  `Config.Exists` if a missing declaration should be an error.
- `Get` validates on the way out: numbers are clamped to `min`/`max`, a
  value that does not parse falls back to the default, a choice not in
  `choices` becomes the first choice, and a string failing
  `validation_pattern` becomes the default with a warning in the log.
- Float configs come back as a .NET double, int configs as an int, a list
  as a .NET `List<string>` (0-based, `.Count`).
- `Set` and `Reset` only affect declared names and save the plugin config
  for stored macros. They are the supported way to persist state between
  runs.

Start-up validation deserves a test before you depend on it.
`HasValidConfigs` (`Utils/ConfigMacroExtensions.cs`) runs each
`validation_pattern` against `cfg.Value.ToString()`, where `cfg.Value` is
the `MacroConfigItem` object and not the user's value, and that class does
not override `ToString`. As written, the pattern is matched against the
type name, so a script with a restrictive `validation_pattern` may be
refused with `One or more of its configs failed to validate.` regardless of
what the user typed. This is a reading of the source, not an observed
failure.

## Script dependencies

```yaml
dependencies:
- source: git://owner/repo/main/lib/helpers.lua
- source: https://github.com/owner/repo/blob/main/lib/helpers.lua
- source: 7f3c2a8e-1111-2222-3333-444455556666   # id of a stored macro
- source: C:\Scripts\helpers.lua
- source: https://example.com/helpers.lua
```

`DependencyFactory.CreateDependency` classifies each `source`:

1. Starts with `git://` or contains `github.com`: a Git dependency. The
   `git://owner/repo/branch/path` form and GitHub `blob` URLs carry a branch
   and path; a bare repository URL parses with an empty path and is
   reported as unavailable.
2. A GUID that matches a stored macro's id: that macro's content. The
   settings panel is the practical way to add one, since ids are not shown
   elsewhere.
3. A path to a file that exists: read from disk.
4. Anything else: fetched over HTTP as given.

Remote content is cached for one day under `cache/dependencies` in the
plugin's config directory, keyed by a hash of source, branch and path.

At start the scheduler checks every dependency is available, downloading if
needed; any failure stops the macro with a chat message naming the
dependency. For Lua macros `LoadDependenciesIntoScope` then runs each
dependency with `lua.DoString(content)` in order, before the script, so
whatever globals it defines are visible. Dependencies run outside the
script's coroutine: defining functions is fine, yielding at their top level
is not. For native macros the availability check still runs but nothing is
loaded.

`require("Name")` from inside a script is the other way to pull in a stored
macro or a file; see [lua-api.md](lua-api.md#require).

## Trigger events

`TriggerEvent` (`Core/Enums.cs`): `OnAutoRetainerCharacterPostProcess`,
`OnUpdate`, `OnConditionChange`, `OnCombatStart`, `OnCombatEnd`, `OnLogin`,
`OnLogout`, `OnTerritoryChange`, `OnChatMessage`, `OnAddonEvent`,
`OnActivePluginsChanged`, `OnDutyStarted`, `OnDutyWiped`,
`OnDutyCompleted`, `OnEmote`.

### Macro-level

Listing an event under `triggers` starts the whole macro when the event
fires. If the macro is already running the start is skipped. Works for
native and Lua macros.

- `OnUpdate` fires every framework tick. A macro-level `OnUpdate` restarts
  the macro as soon as the previous run ends.
- `OnAddonEvent` needs `addon_event` with the addon name and a Dalamud
  `AddonEvent` member name such as `PostSetup`.
- `OnChatMessage` should always have a `chat_message_filter`; the wiki
  warns that an unfiltered chat trigger fires often enough to make the game
  unplayable.
- `OnAutoRetainerCharacterPostProcess` registers with AutoRetainer through
  `AutoRetainerAPI`. When AutoRetainer reaches the post-process step, SND
  requests it, runs the macro when AutoRetainer says the character is
  ready, and calls `FinishCharacterPostProcess()` when the macro reaches
  `Completed` or `Error`. Characters listed in the plugin's exclusion
  setting are skipped. A script that never ends holds AutoRetainer on that
  character. The AutoRetainer side of this contract is in the
  `autoretainer` skill.

### Function-level (Lua only)

While a Lua script is running, any global function whose name starts with a
trigger event name is called when that event fires
(`MacroFunctions`, `MacroScheduler.OnFunctionExecutionRequested`).

- Discovery is a regex over the script text: `function\s+(\w+)\s*\(`. The
  name must start with the event name, case-insensitively.
  `OnTerritoryChange_Diadem` is a valid `OnTerritoryChange` handler.
- The lookup at call time is `lua[name]`, a global lookup. A
  `local function OnUpdate()` matches the regex but is never found.
  Functions in dependencies or required files are not scanned.
- Addon handlers are named `OnAddonEvent_<AddonName>_<EventType>`, for
  example `OnAddonEvent_SelectYesno_PostSetup`. A bare `OnAddonEvent` is
  ignored. An addon whose name contains an underscore cannot be expressed.
- The prefix rule is greedy about your helpers too. A function named
  `OnUpdateStatusText` is an `OnUpdate` handler and will be called every
  frame.
- Handlers run only while the macro's state is `Running`, so not while
  paused, and not before the script body has executed far enough to define
  them.
- The script must still be alive. A script that defines handlers and then
  falls off the end has finished; keep it running with a loop that yields.
- Handlers are called directly with `function.Call()` in the script's Lua
  state, outside the coroutine. They cannot `yield`. Use
  `Engines.Native.Run("/command")` to issue a command; it starts
  immediately and is not awaited, so waits inside it do not delay the
  handler. Set a flag in the handler and act on it from the main loop when
  the reaction needs waiting.
- A handler runs on the thread that raised the event, between two resumes
  of the main coroutine or during one of its yields. Keep handlers short.
- Chat handlers take their filter from `function_chat_filters`, keyed by
  the exact function name.

### Chat filters

```yaml
chat_message_filter:
  channels: [Say, Party]
  message_contains: ready
  sender_contains: Somebody
  message_regex: "^ready( up)?$"
```

`channels` are Dalamud `XivChatType` member names, case-insensitive;
unknown names are dropped. `message_contains` and `sender_contains` are
case-insensitive substring tests, `message_regex` a case-insensitive .NET
regex. All given conditions must hold. An invalid regex logs a warning and
never matches.

### TriggerData

`SetTriggerEventData` creates a global table `TriggerData` with `eventType`
and `timestamp`, then copies the event payload in. For function-level
handlers the table is replaced before each call.

| Event | Extra fields |
| --- | --- |
| `OnTerritoryChange` | `territoryType` |
| `OnConditionChange` | `flag` (`ConditionFlag`), `value` (bool) |
| `OnChatMessage` | `type` (`XivChatType`), `sender`, `message`, `isHandled`, and `timestamp`, which here is the chat message's own timestamp and replaces the event time |
| `OnAddonEvent` | `type` (`AddonEvent`), `args` (Dalamud `AddonArgs`), `AddonName`, `EventType` (string) |
| `OnLogout` | `type`, `code` |
| `OnEmote` | `SourceId`, `SourceName`, `EmoteId`, `TargetId` |
| `OnAutoRetainerCharacterPostProcess` | `Id` (content id), `Name` |
| `OnCombatStart` | `data`, an object with `startTimestamp` and `opponents` |
| `OnCombatEnd` | `data`, an object with `endTimestamp` and `duration` (seconds) |
| `OnActivePluginsChanged` | `data`, an object whose `eventData` holds `changedPlugins`; polled once per second |
| `OnUpdate`, `OnLogin`, `OnDutyStarted`, `OnDutyWiped`, `OnDutyCompleted` | none |

The three `data` rows are payloads that are not dictionaries, so the engine
stores the whole object under `TriggerData.data`. Field casing differs
between events; copy it from the table.

## Cleanup functions

A Lua script may define a global function with one of these names,
case-insensitively: `OnCleanup`, `OnDispose`, `OnStop`, `OnEnd`, `Cleanup`,
`Dispose`, `Stop`, `End` (`MacroFunctions.CleanupNames`). After the script
ends for any reason (normal return, Lua error, Stop, plugin unload) the
engine calls each one found in the text.

- Each cleanup call is wrapped like a script body, so it can `yield`
  commands, for example `yield("/vnav stop")`. Those commands show in the
  status window as `<macro name> - Cleanup Command`.
- Cleanup is discovered from the script text and runs in the same Lua
  state, so it sees the script's globals but not its top-level locals.
- If the script failed before the function was defined, the cleanup call
  errors and is logged; nothing else happens.
- A helper that happens to be called `Stop` or `End` is treated as cleanup
  and will run at exit. Rename such helpers.
- On unload, runs are given 5 seconds in total. Keep cleanup short.

# Script kinds, storage and control

Everything here describes SomethingNeedDoing `v15.13` at commit
`527ca83892506c74fea4d5e07b9b9436e50ef95d`. Paths in parentheses are files
under [source/SomethingNeedDoing/](source/SomethingNeedDoing/).

## Two kinds of macro

`MacroType` has two values (`Core/Enums.cs`):

| Type | Engine | Content |
| --- | --- | --- |
| `Native` | `NativeMacroEngine` | One command per line, each starting with `/`. Game chat commands plus SND's own commands and `<modifier>` tags. See [native-macros.md](native-macros.md). |
| `Lua` | `NLuaMacroEngine` | A Lua script run by NLua 1.7.9. See [lua-api.md](lua-api.md). |

The type is a property of the stored macro. The editor can change it, and
the create dialog asks for it. Nothing infers the type from content, except
`Engines.Run` inside Lua, which picks the native engine when the text starts
with `/`.

Any number of macros can run at once. A Lua script that yields a command
string spawns a child native macro (a `TemporaryMacro`) for that string and
waits for it. Children are tracked by `MacroHierarchyManager`; with the
"Propagate Controls to Child Macros" setting on (the default), pausing or
stopping a parent applies to its children too.

## Where macros live

Macros are `ConfigMacro` objects in the plugin's configuration, not loose
files. The config is serialized by ECommons `EzConfig` with a custom factory
whose file name is `ezSomethingNeedDoing.json` in the plugin's config
directory (`Config.cs`, `ConfigFactory`). Each macro has:

- `Id`: a GUID string generated at creation. It never changes and is what
  local macro dependencies store.
- `Name`: unique across all macros, compared case-insensitively
  (`Config.IsValidMacroName`). A rename to a taken name becomes
  `Name (1)`, `Name (2)`, ...
- `Type`, `Content`, `FolderPath` (a flat string, used as a folder tree in
  the UI), `Metadata` and `GitInfo`.

A `FileSystemWatcher` on the config file raises `Config.ConfigFileChanged`
(debounced 500 ms), which the code editor uses to refresh its buffer. The
watcher only refreshes the editor. Editing the JSON by hand while the game
runs is not a supported way to load scripts.

## Getting a script in

- **Create**: the "Create a new macro" dialog takes a name and a type
  (`Gui/Modals/CreateMacroModal.cs`, not bundled).
- **Import from GitHub**: the same dialog has an optional GitHub URL field.
  `GitMacroManager.Import` requires the form
  `https://github.com/owner/repo/blob/branch/path` and rejects `git://`
  sources and URLs without a file path. The macro becomes a "Git macro":
  its `GitInfo` records repository URL, branch, file path and commit hash,
  it checks for updates on startup, and the settings panel offers a version
  history browser. Clearing the URL or "Reset Git Data" turns it back into a
  local macro (`Gui/Tabs/HelpGeneralTab.cs`).
- **Paste**: there is no file picker and no arbitrary-URL import for whole
  macros at the pin. Pasting into the editor is the local route.
- **Legacy import**: Settings has "Legacy Macro Import", which reads the old
  plugin's `SomethingNeedDoing.json` from the config directory or from the
  clipboard and copies its folder tree in. Imported Lua is kept for reference
  and will not run unchanged; see [migration.md](migration.md).
- **Dependencies** are separate from macro import and do accept local file
  paths and plain HTTP URLs; see
  [metadata.md](metadata.md#script-dependencies).

Macros fetched from GitHub get their settings from the metadata block in the
file, which is why the settings panel has a "Write Metadata to Content"
button.

## Starting, pausing, stopping

From the UI: the editor toolbar and the Macro Status window have Start,
Pause, Resume and Stop. `/snd status` toggles the status window; it can open
automatically on start (`AutoOpenStatusWindow`).

From chat or from another macro (`Services/CommandService.cs`). The main
command is `/somethingneeddoing`; `/snd` and `/pcraft` are aliases.

| Command | Effect |
| --- | --- |
| `/snd` | Toggle the main window |
| `/snd run <name>` | Start the macro with that name |
| `/snd run loop <count> <name>` | Start it with a loop count |
| `/snd pause <name>` | Pause it now |
| `/snd pause loop <name>` | Pause at the next loop point |
| `/snd pause all` | Pause every running macro |
| `/snd resume <name>` | Resume |
| `/snd resume all` | Resume every paused macro |
| `/snd stop <name>` | Stop it |
| `/snd stop loop <name>` | Stop at the next loop point |
| `/snd stop all` | Stop every running macro |
| `/snd status` | Toggle the running-macros window |
| `/snd changelog` | Toggle the changelog window |
| `/snd help` | Print the command list to chat |

Details that differ from what the help text suggests:

- Names may be wrapped in double quotes; the handler trims them. The lookup
  is case-insensitive and exact.
- `/snd run` with an unknown name does nothing and prints nothing. The pause,
  resume and stop handlers do print `Macro '<name>' not found.`
- An unknown first word (anything not in the table) prints the help.
- The help output ends with a line advertising `/snd id <macro name>`. No
  `id` subcommand is registered at the pin, so typing it prints the help
  again.
- `/snd run loop N name` passes `N` as `loopCount`. The native engine runs
  the whole macro `N` times (`loopCount == 0 ? 1 : loopCount`), so `2` means
  two passes, not one pass plus two repeats. The Lua engine ignores the
  count.
- "Loop point" means a `/loop` line or the boundary between passes of a
  looped run. Lua scripts have no loop points; `pause loop` and `stop loop`
  set a flag nothing reads for them.

### What the scheduler does on start

`MacroScheduler.RunMacro` (`Managers/MacroScheduler.cs`), in order:

1. If a macro with the same id is already running, return. Only a Verbose
   log line records it.
2. If the macro failed validation in the last 5 seconds, return silently.
   This stops an `OnUpdate`-triggered macro from printing an error every
   frame.
3. `CanRun`: every entry in `plugin_dependencies` must match the
   `InternalName` of a loaded plugin, and the config validation check must
   pass. Failure prints `Cannot run <name>. <reason>` to chat and the log.
4. Register function-level triggers found in the content.
5. In the background, check script dependencies, downloading remote ones if
   needed. A missing dependency prints
   `Cannot run <name>. The following dependencies are not available: ...`
   and stops the macro. Then disable any `plugins_to_disable`.
6. Set state `Running` and hand the macro to its engine.
7. When the engine returns, set `Completed` (or `Error` first if it threw),
   and re-enable disabled plugins.

States shown in the status window are `Ready`, `Running`, `Paused`,
`Completed` and `Error` (`MacroState`).

### What stop and pause do

- **Stop** cancels the macro's `CancellationToken`, sets the state to
  `Completed`, unregisters its function triggers, re-enables plugins it
  disabled and stops its children.
- **Pause** resets a `ManualResetEventSlim`; engines poll it every 25 ms
  (`Utils/MacroPause.cs`).
- The native engine checks the token and the pause event before every
  command, and `Task.Delay` waits observe the token, so a native macro
  stops mid-`/wait`.
- The Lua engine checks them only between resumes of the script's
  coroutine. A Lua script that never yields cannot be stopped or paused and
  blocks the game's framework thread while it runs. See
  [lua-api.md](lua-api.md#execution-model).
- On plugin unload the scheduler cancels everything and waits up to 5
  seconds for runs to finish, so cleanup functions still get a chance.

## SND's own IPC for other plugins

SND registers one call gate (`External/SomethingNeedDoing.cs`):

```csharp
// EzIPC.Init(this) with no prefix: gate name is
// "SomethingNeedDoing.IsAnyMacroRunning"
bool IsAnyMacroRunning();
```

It returns true when any tracked macro, including temporary child macros, is
in state `Running`. A paused macro does not count. There is no IPC to start,
stop or list macros at the pin; a consumer plugin that needs to start one
has to send `/snd run <name>` as a chat command.

A consumer can bind it with ECommons `[EzIPC]` under prefix
`SomethingNeedDoing`, or with raw Dalamud IPC:

```csharp
// Uncompiled example.
var running = pluginInterface
    .GetIpcSubscriber<bool>("SomethingNeedDoing.IsAnyMacroRunning");
bool busy;
try { busy = running.InvokeFunc(); }
catch (Exception) { busy = false; } // provider not loaded
```

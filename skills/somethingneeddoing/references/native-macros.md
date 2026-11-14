# Native macro reference

Derived from `NativeMacro/MacroParser.cs`, `NativeMacro/ModifierDefinition.cs`,
`NativeMacro/Modifiers/Modifiers.cs` and `NativeMacro/Commands/*.cs` at
commit `527ca83892506c74fea4d5e07b9b9436e50ef95d` (`v15.13`). The bundled
copies are under
[source/SomethingNeedDoing/SomethingNeedDoing/NativeMacro/](source/SomethingNeedDoing/SomethingNeedDoing/NativeMacro/).
The same parser runs every string a Lua script yields.

## How a line is parsed

1. The text is split on `\n`. Blank lines are dropped.
2. SND modifiers are removed from the line and remembered. Game placeholders
   such as `<t>`, `<me>`, `<1>` and `<se.1>` are not SND modifiers and stay
   in the text.
3. What is left must match `^/(\S+)(?:\s+(.*))?$`. A line that does not
   start with `/` throws `MacroSyntaxError`, and because the whole macro is
   parsed before anything runs, one bad line fails the macro before its
   first command. There is no comment syntax: a `#` or `--` line is a syntax
   error, and `//text` is sent to the game as a chat command.
4. The command name is lower-cased and looked up in the table below.
   Anything not in the table becomes a `NativeCommand`, which sends the
   line (without SND modifiers) to the game through ECommons
   `Chat.SendMessage`. That is how `/echo`, `/e`, `/p`, emotes and other
   plugins' commands such as `/vnav` or `/li` work.

Leftover modifier tags are applied after the command object exists. A
modifier a command does not read is accepted and ignored.

## Commands

| Command | Arguments | Behavior |
| --- | --- | --- |
| `/wait` | seconds, decimals allowed | `Task.Delay`. Parsed with invariant culture, so use `.` as the decimal separator. A range such as `/wait 1-2` fails to parse even though the in-plugin help lists it; use `<wait.1-2>` on a command for a random delay. |
| `/waitaddon` | addon name, quotes optional | Poll every 250 ms until the addon is visible and its `UldManager.LoadedState` is `Loaded`. Timeout 5 s, or `<maxwait.N>`. On timeout it logs a warning and continues, unless "Stop on error" is set or the line carries `<errorif.actiontimeout>`. |
| `/callback` | `addon updateState values...` | `Callback.Fire(addon, updateState, values)` if the addon exists and `IsAddonReady`. `updateState` must be `true` or `false`. Each value is parsed as bool, then int, then uint (a trailing `u`/`U` is allowed), else string; `"quoted strings"` may contain spaces. A missing addon is silently skipped unless `<errorif.addonnotfound>`. A present but not-ready addon is always skipped silently. |
| `/click` | `AddonName Method [args...]` | Reflection over ECommons `AddonMaster.<AddonName>`. See [Click](#click). |
| `/target` | name, quotes optional | Nearest targetable object whose name contains the text (case-insensitive). `<index.N>` restricts to object-table index `N`; `<list.N>` skips the `N` nearest matches. No match is silent unless `<errorif.targetnotfound>`. |
| `/targetenemy` | none | Nearest targetable, hostile, living object. |
| `/interact` | none | `TargetSystem.InteractWithObject` on the current target. Failure logs a warning only. |
| `/action`, `/ac` | action name, quotes optional | Sends `/ac "Name"` to the game, with crafting awareness. See [Actions](#actions). |
| `/item` | item name | Finds the `Item` row by exact name (case-insensitive) and uses it. `<hq>` selects the HQ stack. Zero in inventory is silent unless "Stop on error" or `<errorif.itemnotfound>`. An unknown name throws. |
| `/keyitem` | key item name | Same, against the `EventItem` sheet. |
| `/equipitem` | item id (number) | Opens the item's context menu from inventory or armoury and fires the Equip entry. The argument is `uint.Parse`d, so a name throws. The in-plugin help shows this as `/equip`; the parser only knows `equipitem`. |
| `/recipe` | recipe name | Opens the crafting log at that recipe. Throws if the Synthesis window is open. |
| `/require` | status name | Poll every 250 ms until the local player has a status with that name. Timeout 5 s or `<maxwait.N>`; on timeout it throws and the macro errors. |
| `/requirerepair` | optional durability percent | Passes when no equipped item is at or below the threshold; times out with an error when something needs repair. |
| `/requirespiritbond` | optional percent, default 100 | Passes when `Game.HasSpiritbondedItems(percent)` is false; times out with an error otherwise. Read `Utils/Game.cs` for the exact rule before relying on a partial percentage. |
| `/loop` | optional count | Restart the macro from its first line. See [Loops](#loops). |
| `/craft`, `/gate` | optional count, default 1 | Loop guard for crafting. See [Loops](#loops). |
| `/runmacro` | macro name | Ask the scheduler to start another stored macro. It does not wait for it. An unknown name prints `No macro found with name: ...` |
| `/send` | key or `MOD+MOD+KEY` | One key press through ECommons `WindowsKeypress`. Names are Dalamud `VirtualKey` members, case-insensitive (`ESCAPE`, `NUMPAD0`, `CONTROL`, `MENU` for Alt). The last `+` part is the key, the rest are held modifiers. |
| `/hold` | same | Key down, no release. |
| `/release` | same | Key up. |

The two "require gear" commands read oddly because their in-plugin
descriptions say the opposite of what the code checks. Go by the code: both
are guards that stop a crafting loop when gear needs attention.

### Click

`/click AddonName Method [args...]` splits on spaces into at most three
parts. Do not quote: `/click "RecipeNote Synthesize"` leaves a quote inside
the addon name and fails, which is the bug the v1 to v2 config migration
fixed in the default craft-loop template.

- The addon must exist (`Addon <name> not found` otherwise).
- `AddonName` must be a nested type of `ECommons.UIHelpers.AddonMasterImplementations.AddonMaster`.
- `Method` is a method on that type with the same number of parameters as
  arguments. Arguments are strings, converted with the parameter type's
  static `Parse(string)`.
- `Sub.Method` reaches a method on a property or field, and
  `Sub[i].Method` reaches the `i`-th element of an enumerable member.

The examples in the command's own help attribute (`/click SelectYesNo_Yes`)
do not match the parser at the pin: a single token is rejected with
`Invalid click command`. Help, then the Clicks tab, lists what the
installed ECommons actually exposes; copy from there. When no AddonMaster
entry exists for a window, use `/callback`.

### Actions

`/ac Name <placeholders>` is rebuilt as `/ac "Name" <placeholders>` and
`<party.N>` is rewritten to `<N>`. Then, on the framework thread:

1. If the line has `<condition...>` and the current Synthesis condition does
   not satisfy it, the action is skipped.
2. If the name is a crafting action for the current job:
   - with "Skip craft actions when not crafting" (`CraftSkip`, default on),
     skip when not crafting or when progress is already at maximum;
   - with `QualitySkip` (default on), skip quality-only actions at maximum
     quality;
   - with "Smart wait" (`SmartWait`, default off), no `<unsafe>` and no
     `<wait>`, wait for the `ExecutingCraftingAction` condition to clear,
     up to 5 seconds.
3. Send the command.
4. Apply the `<wait>` only if the action was actually sent.

A skipped action costs no wait, so a macro written for a full craft finishes
quickly when the craft completes early.

### Loops

- `/loop` with no count loops forever. `/loop N` restarts the macro `N`
  more times, or `N - 1` with the "Count /loop number as total iterations"
  setting (`LoopTotal`). The counter belongs to the parsed command, so it
  survives the restart.
- `/gate N` (alias `/craft N`) goes at the top of a macro that ends in an
  unconditional `/loop`. It passes `N` times and on the next
  visit ends the macro quietly by throwing `MacroGateCompleteException`, so
  the body runs `N` times.
- `<echo>` on either prints the remaining count; the "Always echo /loop
  commands" setting does the same globally.
- Pause-at-loop and stop-at-loop requests are honored at `/loop`.

When a native macro's metadata has `crafting_loop: true`, `CraftLoopBuilder`
wraps the content before parsing. With the template setting on, it
substitutes `{{macro}}` and `{{count}}` into the configured template
(default: `/craft {{count}}`, wait for `RecipeNote`, `/click RecipeNote
Synthesize`, wait for `Synthesis`, the macro, `/loop`). Otherwise it builds
equivalent steps depending on whether the recipe window is open at start.
`craft_loop_count: -1` means until materials run out or the macro is
stopped.

## Modifiers

All are matched case-insensitively anywhere on the line
(`ModifierDefinitions.SndModifiers`).

| Modifier | Meaning | Read by |
| --- | --- | --- |
| `<wait.N>` | Wait `N` seconds after the command. Decimals allowed. | Every command, through `PerformWait`. On `/wait` the modifier replaces the argument: `/wait 3 <wait.1>` waits one second. |
| `<wait.A-B>` | Random wait between `A` and `B` seconds. `A` greater than `B` is a syntax error. | Same |
| `<maxwait.N>` | Timeout in seconds for a waiting command | `/waitaddon`, `/require`, `/requirerepair`, `/requirespiritbond` |
| `<echo>` | Print remaining count | `/loop`, `/gate`, `/craft` |
| `<unsafe>` | Do not smart-wait for the craft action to finish | `/ac` |
| `<hq>` | Use the HQ item | `/item` |
| `<condition.a,b>` | Run only when the Synthesis condition is one of the listed names | `/ac` |
| `<condition.not.a>`, `<condition.!a>` | Negated | `/ac` |
| `<index.N>` | Object-table index | `/target` |
| `<list.N>` | Skip the `N` nearest matches (was `<listindex.N>` before 12.64) | `/target` |
| `<errorif.X>` | Turn a silent miss into an error | See below |
| `<distance.N>` | Parsed and stored | Nothing at the pin |

`<errorif.X>` values are the `ErrorCondition` enum, case-insensitive:
`actiontimeout` (`/ac` smart wait, `/waitaddon`), `itemnotfound` (`/item`,
`/keyitem`), `targetnotfound` (`/target`, `/targetenemy`),
`addonnotfound` (`/callback`), and `cantuseitem`, which is defined but not
checked by any command. An unknown value throws while parsing.

Condition names come from ECommons `AddonMaster.Synthesis.Condition` and are
compared as lower-cased enum names, so use the English names the enum
defines (`good`, `excellent`, `poor`, ...) even though the upstream
description says the text may be localized.

Things the source does not do, despite older documentation:

- A `PartyIndexModifier` class exists and `/target` checks for it, but no
  modifier definition ever creates one. `/target` cannot select by party
  slot at the pin; `<1>` to `<8>` stay in the text and work only for
  commands the game itself interprets, such as `/ac "Cure" <2>`.
- "Max Timeout Retries", "Noisy Errors" and the beep settings exist in the
  config and settings UI. No engine or command reads them at the pin.

## Errors and the "Stop on error" setting

`StopOnError` (default off) turns these soft failures into thrown errors:
`/ac` smart-wait timeout, `/item` and `/keyitem` with zero count, and
`/waitaddon` timeout. Errors that always throw: syntax errors, `/click`
failures, `/recipe` during a craft, unknown item names, `/require*`
timeouts.

A thrown error prints `Error executing macro command` and
`Macro execution failed` to chat in the configured error channel, and the
exception with its message goes to the Dalamud log. Chat does not show
which line failed; the log does.

## Threading

Commands that set `RequiresFrameworkThread` run inside
`Framework.RunOnTick`. `/wait`, `/loop`, `/gate`, `/send`, `/hold`,
`/release` and `/runmacro` do not. Waits are awaited, not slept, so a native
macro does not hold the framework thread between commands.

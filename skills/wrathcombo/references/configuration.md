# Configuration and troubleshooting

These contracts describe the [recorded revisions](provenance.md). Recheck the
provider, API enums, and runtime readers together when targeting another build.
An enum member or accepted setter alone does not establish runtime support.

## Combos and options

Use [Provider.cs](source/WrathCombo/WrathCombo/Services/IPC/Provider.cs) and
[Search.cs](source/WrathCombo/WrathCombo/Services/IPC/Search.cs) for discovery:

| `WrathCombo.` gate suffix | Arguments | Result |
| --- | --- | --- |
| `GetComboNamesForJob` | `uint jobID` | `List<string>?` |
| `GetComboOptionNamesForJob` | `uint jobID` | `Dictionary<string,List<string>>?` |
| `GetComboState` | `string comboInternalName` | `Dictionary<ComboStateKeys,bool>?` |
| `SetComboState` | `Guid lease, string comboInternalName, bool comboState, bool autoState` | `SetResult` |
| `GetComboOptionState` | `string optionName` | `bool` |
| `SetComboOptionState` | `Guid lease, string optionName, bool state` | `SetResult` |
| `IsCurrentJobConfiguredOn` | none | `Dictionary<ComboTargetTypeKeys,ComboSimplicityLevelKeys?>` |
| `IsCurrentJobAutoModeOn` | none | same dictionary type |

Job IDs are ClassJob IDs. Names returned by discovery are internal preset
identifiers; options are child presets, not arbitrary numeric slider settings.
Getters prefer lease-controlled values where present. Missing combo/job entries
can return null. An unknown option can throw in `GetComboOptionState` because
its fallback indexes a missing entry; discover the option first and retain
exception handling.

`ComboStateKeys` is `Enabled=0, AutoMode=1`. A combo being enabled for button
replacement does not imply its AutoMode flag is enabled. Neither flag establishes
that the global auto-rotation switch is on.

`ComboTargetTypeKeys` is `SingleTargetDPS=0, AoEDPS=1, SingleTargetHeals=2,
AoEHeals=3, Other=4`. `ComboSimplicityLevelKeys` is `Simple=0, Advanced=1,
Other=2`; a null value means no matching enabled mode. `IsCurrentJobConfiguredOn`
includes healing categories for healers/BLU and a single DPS category for BST.
The public `IsCurrentJobAutoModeOn` returns the two DPS categories at this pin.
Do not assume the two dictionaries have identical keys or that readiness checks
every healing AutoMode flag.

`SetComboState` also disables the opposite Simple/Advanced preset in that lease,
even when the requested state is false. `SetCurrentJobAutoRotationReady` is a
broader operation: for a ready job it registers currently enabled parent combos
and child options; otherwise it prefers Simple combos and falls back to Advanced
combos plus their options. It returns `OkayWorking`. The source comments describe
locking the user's setup, but the ready branch enumerates enabled presets; do
not promise that every disabled option is locked. Prefer leaving an already
usable setup alone when the task only needs the global rotation switch.

Variant helpers are `GetVariantParentComboName(uint jobID)`,
`GetVariantOptionNames(uint jobID)`, and
`SetVariantReadyForJob(Guid lease, uint jobID, bool enabled)`.
Occult helpers are `GetOccultParentComboName(uint phantomJobID)`,
`GetOccultOptionNames(uint phantomJobID)`, and
`SetOccultReadyForPhantomJob(Guid lease, uint phantomJobID, bool enabled)`.
Their getters return `string?` / `List<string>?`; setters return `SetResult`.
Keep Phantom Job IDs separate from ordinary ClassJob IDs. Inspect the matching
mapping in `Search.cs` before applying a numeric ID.

## Auto-rotation configuration

The raw provider exposes `GetAutoRotationConfigState(object passedOption)`
returning `object?`, and
`SetAutoRotationConfigState(Guid lease, object passedOption, object value)`
returning `SetResult`. The wrapper accepts `AutoRotationConfigOption` and typed
setter overloads for `bool`, `int`, `DPSRotationMode`, and `HealerRotationMode`.
The raw API helper's generic arguments are `object`, matching the provider.

Read the exact [option enum](source/WrathCombo.API/Enum/AutoRotationConfigOption.cs),
[conversion code](source/WrathCombo/WrathCombo/Services/IPC/ProvideAutoRotConfig.cs),
and [runtime readers](source/WrathCombo/WrathCombo/AutoRotation/AutoRotationConfigIPCWrapper.cs).

| ID | Option | Value type |
| --- | --- | --- |
| 0 | `InCombatOnly` | bool |
| 1 | `DPSRotationMode` | DPSRotationMode |
| 2 | `HealerRotationMode` | HealerRotationMode |
| 3 | `FATEPriority` | bool |
| 4 | `QuestPriority` | bool |
| 5 | `SingleTargetHPP` | int |
| 6 | `AoETargetHPP` | int |
| 7 | `SingleTargetRegenHPP` | int |
| 8 | `ManageKardia` | bool |
| 9 | `AutoRez` | bool |
| 10 | `AutoRezDPSJobs` | bool |
| 11 | `AutoCleanse` | bool |
| 12 | `IncludeNPCs` | bool |
| 13 | `OnlyAttackInCombat` | bool |
| 14 | `OrbwalkerIntegration` | bool |
| 15 | `AutoRezOutOfParty` | bool |
| 16 | `DPSAoETargets` | int |
| 17 | `SingleTargetExcogHPP` | int |
| 18 | `AutoRezDPSJobsHealersOnly` | bool; runtime caveat below |
| 19 | `DPSAlwaysHardTarget` | bool |
| 20 | `HealerAlwaysHardTarget` | bool |
| 21 | `BypassQuest` | bool |
| 22 | `BypassFATE` | bool |
| 23 | `IgnoreRangeInBoss` | bool |
| 24 | `UnTargetAndDisableForPenalty` | bool |
| 25 | `IncludeShields` | bool; incomplete wiring below |

`DPSRotationMode` values are `Manual=0, Highest_Max=1, Lowest_Max=2,
Highest_Current=3, Lowest_Current=4, Tank_Target=5, Nearest=6, Furthest=7`.
`HealerRotationMode` values are `Manual=0, Highest_Current=1, Lowest_Current=2`.
They describe target selection, not a choice of rotation algorithm.

Enum getters can return an enum from saved configuration or an integer from
lease state. Convert by the declared enum's numeric value after checking for
null; do not rely on one boxed runtime type. Setters use conversion and store
integer values. They do not comprehensively validate enum membership or useful
slider ranges. Validate values against the target UI/domain before sending them.

At this pin, `IncludeShields=25` has no saved-value getter switch case. A lease
can store it and make the getter return it, while the runtime wrapper still
reads `settings.IncludeShields` directly. `AutoRezDPSJobsHealersOnly=18` likewise
has getter/setter support but a direct saved-value runtime reader. Do not claim
these two overrides change rotation behavior without a newer source fix or a
runtime check. This is a source observation, not a live-tested bug report.

There is another ordering hazard in `AddRegistrationForAutoRotation`: its
Duplicate check tests whether any auto-rotation configuration overrides exist,
then indexes `AutoRotationControlled[0]`. Setting configuration before ever
setting global rotation on that lease can therefore throw. For an authorized
rotation-start workflow, make the global state call before configuration
setters, and handle a failure after that call by releasing the lease. Do not
turn rotation on just to edit an unrelated setting.

## Commands and coexistence

[Commands.cs](source/WrathCombo/WrathCombo/Commands.cs) defines `/wrath` controls:

| Command | Effect at this pin |
| --- | --- |
| `/wrath auto on` / `/wrath auto off` | Set the saved global auto-rotation flag. |
| `/wrath auto` / `/wrath auto toggle` | Invert the saved flag. |
| `/wrath combo on` / `/wrath combo off` | Set saved action replacement and enable/disable its hook. |
| `/wrath combo` / `/wrath combo toggle` | Invert saved action replacement. |
| `/wrath auto target damage nearest` | Set saved DPS targeting mode to `Nearest`. |
| `/wrath auto target healer lowest_current` | Set saved healer targeting mode to `Lowest_Current`. |
| `/wrath autosettings` | Open the auto-rotation settings UI. |
| `/wrath dbg` | Open the debug UI, including IPC diagnostics. |

Use explicit on/off for a requested state. The auto parser treats an unrecognized
second argument as false, so a typo can disable the saved flag. Target mode
arguments use the enum names above, case-insensitively. A loaded Wrath plugin
must have registered the command; a command dispatch does not return a typed
success result. Effective rotation still needs the player, suitable presets,
and runtime conditions described below.

`ToggleAutoRotation` updates the saved flag, queues a save, and clears its
internal pause state. The command calls it only when the saved flag changes;
repeating `auto on` is not a general unpause command. Target commands also queue
saved configuration changes. `SetActionChanging` queues a save when changed
and adjusts the action-replacement hook. Saves are processed later; debug-config
mode skips them. These are persistent user settings, whereas IPC leases provide
temporary overrides. Commands do not release leases, and a lease can keep the
effective value different from the saved command value. Read
`GetAutoRotationState` and inspect the UI for effective state; do not claim that
`auto off` overrides another client's active lease.

Disabling action replacement does not disable auto-rotation: the controller
still runs and selects its direct action path. Opening a UI changes neither
plugin loading nor execution. Wrath's lease API supplies no general plugin-load
or navigation control. Prefer a lease for temporary integration. If a task
explicitly requires persistent command changes, record the intended setting,
verify readback, and avoid blindly restoring an old snapshot over later edits.

The [conflict checks](source/WrathCombo/WrathCombo/Data/Conflicts/ConflictingPluginsChecks.cs)
inspect competing plugins' targeting, automatic action queues,
rotation, and retargeting, as well as overlaps with action-retargeting plugins.
Their [BossMod helper](source/WrathCombo/WrathCombo/Services/IPC_Subscriber/BossMod.cs)
uses a mixture of IPC and reflection. These checks show possible conflicts;
they do not guarantee compatibility or establish a stable API for controlling
those other plugins. Keep one intended automatic action controller and coordinate
targeting with the user's setup. Passive encounter displays can remain a separate
concern; inspect actual enabled features before changing another plugin.

Wrath's optional Orbwalker integration affects casting while moving. The
[helper](source/WrathCombo/WrathCombo/Services/IPC_Subscriber/Orbwalker.cs) checks
version, enabled state, mouse movement, and job eligibility, and the rotation
controller consults it before casting. Enabling the Wrath option alone is not
evidence that movement is controlled. This skill requires no companion
rotation-plugin skill; use Wrath's own API for its controls.

## Diagnosing a stalled integration

Start with the operation, lease ID, result or exception category, cancellation
reason, player job, and plugin/API versions. Log only the data needed for the
failure; keep personal character records out of reusable examples.

- If `IPCReady` is false, inspect player availability and the deferred cache
  construction in `Provider.BuildCaches`. Poll with a finite deadline rather
  than registering repeatedly.
- If setters return `IPCDisabled`, inspect `[Wrath IPC]` logs and
  `Helper.IPCEnabled`. That source checks an upstream status file, caches it for
  45 minutes, and defaults to enabled when the fetch fails. Readiness and remote
  service status are separate. Do not patch out suspension to force a call.
- If the lease was cancelled, use the reason table in [lifecycle](ipc.md).
  `WrathCombo.UpdateCaches` suspends leases on job changes and first-run cache
  setup. Reacquire only as part of a still-authorized operation after readiness.
- If a write succeeds but has no effect, inspect competing controllers in
  Wrath's IPC debug UI (`/wrath dbg`), lease timestamps, combo Enabled/AutoMode
  flags, and the actual runtime reader. UI caches can defer visible changes.
- If rotation is enabled but no action occurs, inspect
  [AutoRotationController](source/WrathCombo/WrathCombo/AutoRotation/AutoRotationController.cs).
  Player death, mounting, occupied state, queued actions, raise invincibility,
  throttling, pause state, combat-only settings, and targeting checks can prevent
  execution. Check the selected combo and its job-specific conditions next.

Use public getters for diagnosis. `Leasing`, `Search`, `UIHelper`, the action
replacer, and the rotation controller are implementation evidence, not stable
consumer APIs. Direct JSON edits or reflection writes bypass lease arbitration
and can overwrite user configuration.

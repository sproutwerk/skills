# Quest paths and diagnostics

Start with the target quest's existing JSON and the pinned
[quest schema](source/Questionable/QuestPaths/quest-v1.json). Its common schemas
are in [Questionable.Model](source/Questionable/Questionable.Model/). Register
those local schemas under their `$id` URIs when validating offline; fetching
the current hosted schema can mix versions. Upstream's
[ValidJsonFilesTest](source/Questionable/QuestPaths.JsonValidator/ValidJsonFilesTest.cs)
shows the schema registrations.

## File identity and sequence structure

[QuestRegistry](source/Questionable/Questionable/Controller/QuestRegistry.cs)
extracts the element ID before the first underscore in the filename, using
`ElementId.FromString`. Keep the conventional `<id>_<name>.json` filename.
The JSON root does not supply that ID.

[QuestRoot](source/Questionable/Questionable.Model/Questing/QuestRoot.cs)
contains `Author`, `QuestSequence`, optional comments, and path-wide settings.
Each sequence has a byte `Sequence` value and an ordered `Steps` list.
Sequence values follow the game's quest progress; step indices are zero-based
positions within a sequence. Do not renumber game sequences just to make them
consecutive. Acceptance commonly uses sequence `0`, and turn-in commonly uses
`255`; verify the quest's actual progression.

The bundled
[Close to Home path](<source/Questionable/QuestPaths/2.x - A Realm Reborn/MSQ-1/Gridania/85_Close to Home (Lancer).json>)
shows acceptance, quest-variable conditions, territory transitions, attunement,
and turn-in. Preserve its author and `LastChecked` records if adapting it.
Set new verification metadata only after the corresponding runtime check.

## Choosing step fields

Read [QuestStep](source/Questionable/Questionable.Model/Questing/QuestStep.cs)
alongside its
[interaction converter](source/Questionable/Questionable.Model/Questing/Converter/InteractionTypeConverter.cs).
The schema uses serialized spellings and interaction-specific requirements;
a C# enum name alone is not enough to author a valid step.

| Field or group | Check before editing |
| --- | --- |
| `DataId` | The intended game object data ID, separate from quest ID and live object/entity ID |
| `Position`, `TerritoryId` | World-space X/Y/Z and the intended territory, including floor height |
| `InteractionType` | The task that must complete at this step, plus its required fields |
| `TargetTerritoryId` | Whether the step must wait for a zone transition |
| `StopDistance`, `NpcWaitDistance` | The relevant task's distance check; a larger value can leave interaction out of range |
| `Fly`, `Mount`, `DisableNavmesh` | Movement behavior supported by the route and character; direct movement can cross obstacles |
| `AetheryteShortcut`, `AethernetShortcut` | Serialized names from the bundled common schemas and converters |
| `SkipConditions`, `CompletionQuestVariablesFlags` | The exact predicate and quest variables that allow the step to be skipped or considered complete |
| `DialogueChoices`, quest/item/action fields | Matching IDs and text/row references for the intended interaction |

`QuestStep.DefaultStopDistance` is `3f`; nullable movement fields can carry
different meaning from explicit false. Keep the distinction when editing.
Use measured or supplied coordinates and IDs. If those are missing, leave
the path incomplete and identify the missing data instead of inventing it.

The schema and the C# model can differ. At this pin the schema accepts a
boolean `Passive` under `SinglePlayerDutyOptions`, while
[SinglePlayerDutyOptions.cs](source/Questionable/Questionable.Model/Questing/SinglePlayerDutyOptions.cs)
has no such member, so check the target build before relying on it.
`EExtendedClassJob` gained `Zephyr`, and its group and configured-job members
now have fixed values from 247 to 254; use the serialized names from the
converter, never the numbers.

The schema permits fields for different phases of a step. Follow
[SkipCondition.cs](source/Questionable/Questionable/Controller/Steps/Shared/SkipCondition.cs)
before changing `StepIf`, shortcut skip conditions, or completion flags.
Those conditions are not interchangeable. For an interaction outside the
bundled handlers, inspect `Questionable/Controller/Steps/` in the full source
at the recorded revision.

## Loading and overrides

At this pin, `QuestRegistry.Reload` first reads
`<plugin config directory>/PathData/bundle.zip`. The archive needs a manifest
compatible with `PathDataFormat.CurrentVersion` (`1` here), and quest entries
under `QuestPaths/`. A missing or incompatible bundle can leave no baseline
quests: the assembly-loading call in `Reload` is commented out, despite
older fallback wording in the implementation.

When advanced debug mode, a development plugin, or a DEBUG build enables the
branch, the registry then loads project paths and
`<plugin config directory>/Quests/` recursively. Later entries overwrite
earlier entries with the same ID, so user overrides take precedence over
successfully loaded bundle/project entries. A normal release without that
branch does not load those local overrides.

Downloaded path data can change independently of the plugin assembly.
Record its manifest/data version when diagnosing a path problem. Check
[PathDataBundle](source/Questionable/Questionable/PathData/PathDataBundle.cs),
[PathDataManifest](source/Questionable/Questionable/PathData/PathDataManifest.cs),
and [PathDataUpdater](source/Questionable/Questionable/PathData/PathDataUpdater.cs)
for the target build's handling. Avoid replacing a user's bundle or enabling
debug settings as an incidental documentation check.

## Diagnose from the failing stage

| Symptom | Inspect |
| --- | --- |
| Start returns false | ElementId syntax, filename ID, registry presence, bundle loading errors |
| Edited path has no effect | Debug/development condition, actual config directory, reload, duplicate IDs and source precedence |
| Start returns true but no progress | Current quest/sequence/step, controller logs, eligibility, configured stop conditions, companion readiness |
| Walks near an NPC but stalls | Correct DataId and territory, height, stop distance, interaction task and dialogue choice |
| Skips needed work or repeats a step | Quest variables, skip predicates, sequence transitions, existing completion flags |
| Continues to other quests | `StartQuest` versus `StartSingleQuest`, automatic priorities and current quest selection |
| Appears idle during a wait | Both automation mode and queued work; treat missing IPC as an error, not idle |

The command handler exposes `/qst reload`, `/qst validation`, and `/qst help-all`
for user-directed diagnosis. Reload mutates the active registry; do not invoke
these commands on a live client merely to validate the skill.

Validate edited JSON against the matching local schema and inspect semantic
validators in the target source. Schema success does not prove the IDs,
coordinates, or game conditions are correct. Build changes in a full target
checkout when needed, and report separately whether any game testing occurred.

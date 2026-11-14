# Configuration and venture plans

Read the [plugin config](source/AutoRetainer/AutoRetainer/PluginData/Config.cs),
[API DTOs](source/AutoRetainerAPI/AutoRetainerAPI/Configuration/), and
[provider](source/AutoRetainer/AutoRetainer/Modules/IPC.cs) together. A field in
a DTO does not prove there is an endpoint that updates just that field.

## Character and retainer identity

Characters use a `ulong` content ID (`CID`). `GetRegisteredCharacters()` filters
out blacklisted and uninitialized entries; the returned IDs do not establish
that characters are enabled for processing. `GetOfflineCharacterData(cid)` can
return null. `Name@World` is used for relog selection, while selection and
additional retainer data use CID plus retainer name.

| Location | Meaning |
| --- | --- |
| `Config.SelectedRetainers` | `Dictionary<ulong, HashSet<string>>` of selected retainer names |
| `OfflineCharacterData.Enabled` | Character participates in retainer MultiMode |
| `OfflineCharacterData.WorkshopEnabled` | Character participates in workshop MultiMode |
| `OfflineCharacterData.RetainerData` | Cached retainer observations, including venture timing |
| `Config.AdditionalData` | Per-retainer settings keyed by `#` + CID as 16 uppercase hex digits + a space + name |
| `AdditionalRetainerData` | Planner, entrust-plan reference, gil transfer settings, and cached gear stats |

Use `GetAdditionalRetainerData(cid, name)` rather than manufacturing a dictionary
key in a consumer. This getter creates an entry if one is missing, so reading an
invented name can change configuration. Verify the name against the character's
retainer list first. `OfflineDataManager` refreshes cached observations when the
relevant game state is available; they are not a live inventory or character API.

## Data lifetime and writes

The wrapper requires read, change, and write within one frame. Both mutable DTOs
record `CreationFrame`; their write methods compare it with
`Svc.PluginInterface.UiBuilder.FrameCount` and throw on a mismatch. Keep the whole
sequence on the framework thread without `await`, queue delays, or retaining the
DTO for a later callback. Retain identifiers between callbacks and reacquire data.

The provider updates public fields on an existing record. Sending a partial DTO
filled with default values can overwrite unrelated settings. Read the complete
record, change only the requested fields, and immediately use the corresponding
write method. If the freshness check fails, inspect dependency versions and
reacquire the DTO; do not bypass the check through raw IPC or alter `CreationFrame`.

`AutoRetainerApi.Config` is a collection of getters. The
[config provider](source/AutoRetainer/AutoRetainer/Services/IpcConfigValuesProvider.cs)
registers `AutoRetainer.GetConfig.<fieldName>` for public instance fields of the
plugin config. It does not register a matching generic setter. Do not invent
`AutoRetainer.SetConfig.*`, or assume mutating a returned collection persists it.
The getter list can lag the provider's fields, so compare both files when a
property is missing.

For an explicitly requested on-disk edit, locate the target installation's
configuration and preserve a backup. Stop or unload the writer before editing,
preserve unrelated fields, and validate the JSON before reloading. Do not commit
the user's configuration as a fixture. An IPC write changes in-memory fields;
the shown setter itself does not call a disk-save method.

## Venture plans

`AdditionalRetainerData` contains `EnablePlanner`, `VenturePlan`,
`VenturePlanIndex`, and `LinkedVenturePlan`. A plan's `List` contains
`PlannedVenture` records with `uint ID` and `int Num`. `ListUnwrapped` expands
each entry `Num` times; the index refers to that expanded list. Planner activity
requires both `EnablePlanner` and a nonempty expanded list.

Venture IDs are `RetainerTask.RowId`, not item IDs or `RetainerTaskNormal` row
IDs. Confirm the task's job category, retainer level, gear or gathering
requirements, and relevant unlocks. The plugin's
[VentureUtils](source/AutoRetainer/AutoRetainer/Helpers/VentureUtils.cs) resolves
items and adjusts tasks for the retainer's job. For new IDs, query the matching
game sheets or use the target client's venture browser; do not guess a row from
the desired item's number.

When changing a plan, validate nonzero IDs and positive repeat counts, then
decide whether the user's requested edit should reset `VenturePlanIndex`.
Inspect `LinkedVenturePlan` before changing a linked plan; the
[planner UI](source/AutoRetainer/AutoRetainer/UI/Windows/VenturePlanner.cs) shows
how it identifies the referenced retainer's plan.

`PlanCompleteBehavior` has four values: `Restart_plan`, `Assign_Quick_Venture`,
`Do_nothing`, and `Repeat_last_venture`. Follow the
[scheduler](source/AutoRetainer/AutoRetainer/Scheduler/SchedulerMain.cs) and
`GetNextPlannedVentureIndex` in
[Utils](source/AutoRetainer/AutoRetainer/Helpers/Utils.cs) when interpreting the
completion boundary. The index can reach the expanded count; it is not always
an in-range array index.

An event-time `SetVenture` override takes precedence over the configured plan
for that scheduling decision. It is not a persistent plan edit.

## Workshop and inventory data

Workshop state lives in each `OfflineCharacterData`: enabled airship/submarine
name sets, cached vessel lists, and additional data dictionaries. Per-vessel
behavior, unlock plans, and point plans have separate types in the API's
configuration directory. Verify the vessel's name, character, and type before
editing; airship and submarine records are distinct.

Inventory cleanup plans and entrust plans can move or remove items. Preserve
protection settings and plan associations, and limit edits to the user's named
items and requested operation. A data or UI review alone does not authorize
running a cleanup plan.

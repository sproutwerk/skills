# IPC contracts

The provider is
[QuestionableIpc.cs](source/Questionable/Questionable/External/QuestionableIpc.cs).
Every channel below has the prefix `Questionable.` and is registered with
`RegisterFunc`. Invoke it with `InvokeFunc`, including `Stop` and `ClearQuestPriority`.
Generic arguments to `GetIpcSubscriber` put the return type last.

## State and quest selection

| Channel suffix | Arguments → result | Meaning at the pin |
| --- | --- | --- |
| `IsRunning` | none → `bool` | Automation mode is non-manual **or** the controller has unfinished tasks |
| `GetCurrentQuestId` | none → `string?` | Current progress quest ID; null when absent |
| `GetCurrentStepData` | none → nullable DTO | Quest ID, sequence, step, interaction, position, territory |
| `GetCurrentlyActiveEventQuests` | none → `List<string>` | Quest IDs of hardcoded seasonal events inside their date window; see below |
| `StartQuest` | `string questId` → `bool` | Select a registered quest and start automatic mode |
| `StartSingleQuest` | `string questId` → `bool` | Select a registered quest and start single-quest mode |
| `Stop` | `string label` → `bool` | Stop questing and movement, combat, and gathering controllers; returns true |
| `IsQuestLocked` | `string questId` → `bool` | Invalid or unregistered IDs return true |
| `IsQuestLockedReason` | `string questId` → `(bool, string)` | Lock result and comma-joined reasons; invalid/unregistered returns `(true, "")` |
| `IsQuestComplete` | `string questId` → `bool` | Completion query; invalid syntax returns false |
| `IsReadyToAcceptQuest` | `string questId` → `bool` | Acceptance eligibility, including level and lock checks |
| `IsQuestAccepted` | `string questId` → `bool` | Acceptance query; invalid syntax returns false |
| `IsQuestUnobtainable` | `string questId` → `bool` | Unobtainable-state query; invalid syntax returns false |

Start methods check ID parsing and registry presence, then call `SetNextQuest`
and `Start`/`StartSingleQuest`. They do not establish that prerequisites are
met, a task started successfully, or the quest finished. The controller can
stop or choose work according to its current state and settings.

The step DTO has `string QuestId`, `byte Sequence`, `int Step`,
`string InteractionType`, `Vector3? Position`, and `uint TerritoryId`.
It is null when progress or the indexed step is absent. The bundled upstream
`Questionable.IpcTest` uses an older `ushort TerritoryId` DTO; use the provider
as the contract and verify DTO conversion against the consumer's Dalamud
version. The sample is not evidence that mismatched types remain compatible.

`GetCurrentlyActiveEventQuests` reads a list compiled into
[EventInfoComponent](source/Questionable/Questionable/Windows/QuestComponents/EventInfoComponent.cs),
so its contents change with plugin releases, not with path data. An event is
included when it has started (or has no start date), has not ended, and at
least one of its quests is obtainable and, unless the user enabled showing
completed seasonal events, incomplete. The call then returns every quest ID of
that event, completed ones included. Builds before `15.756.3.57` required all
of an event's quests to pass that check. The debug command `/qst@ eventquests`
prints the same list.

`Questionable.ReloadData` is a notification emitted with `SendMessage()` by
[QuestRegistry](source/Questionable/Questionable/Controller/QuestRegistry.cs).
It is not a callable reload function. A listener uses `GetIpcSubscriber<object>`
with `Subscribe(Action)` and removes its subscription on disposal.

## IDs and query limits

[ElementId.cs](source/Questionable/Questionable.Model/Questing/ElementId.cs)
parses ordinary quest IDs as unsigned 16-bit decimal strings. A Lumina Quest
row ID is converted by `QuestId.FromRowId` using `rowId & 0xFFFF`; only apply
that conversion to known Quest row IDs. Object DataIds and item IDs are
different identifiers.

The parser also accepts `S` for satisfaction-supply NPC IDs, `U` for unlock
links, `N` for aethernet IDs, `C` for aether currents, and `A<society>x<rank>`
for allied-society daily groups (`A<society>` defaults rank to zero).
Prefixes are case-sensitive. Parsing successfully does not mean every IPC
query supports that element type.

[QuestFunctions](source/Questionable/Questionable/Functions/QuestFunctions.cs)
returns false for satisfaction-supply and daily-group completion, queries
unlock state for `U`, and can throw for other element types. Ordinary quest
completion uses `QuestManager.IsQuestComplete`. Daily completion used for
priority removal has separate logic. Do not use the ordinary completion bit
as a universal result for repeatable quests, gathering, or pseudo-elements.

## Priority list

| Channel suffix | Arguments → result |
| --- | --- |
| `ExportQuestPriority` | none → `string` |
| `ImportQuestPriority` | `string encodedQuestPriority` → `bool` |
| `ClearQuestPriority` | none → `bool` |
| `AddQuestPriority` | `string questId` → `bool` |
| `InsertQuestPriority` | `int index, string questId` → `bool` |

[PriorityWindow](source/Questionable/Questionable/Windows/PriorityWindow.cs)
exports `qst:priority:` followed by base64 of UTF-8 semicolon-separated
ElementId strings. Import also accepts the legacy `qst:v1:` prefix. It does
not accept plain JSON or raw semicolon-separated text. Invalid decoding
produces an empty list, and the IPC import still returns true.

[QuestPriorityManager](source/Questionable/Questionable/Controller/QuestPriorityManager.cs)
imports by appending known, nonduplicate quests. It does not replace the list.
`AddQuestPriority` and `InsertQuestPriority` also return true for invalid or
unknown IDs at the provider boundary. The ElementId overload of `Add` returns
true for duplicates. `Insert` returns false when its attempted insertion
throws, such as an out-of-range index. Verify the exported contents when the
caller needs to know what changed.

The export contains IDs only. It does not preserve internal accept-only
flags, and `Clear` removes those flags. An export/clear/import cycle therefore
cannot restore all priority state. Use a local consumer queue when possible;
obtain authorization for edits to the shared priority list.

## Gathering and New Game+

`StartGathering(uint npcId, uint itemId, byte classJob, int quantity) -> bool`
and `StartGatheringComplex(uint npcId, uint itemId, byte classJob, int quantity,
ushort collectability) -> bool` delegate to the controller. IPC callers pass
every argument even though the private implementation declares defaults.
At this pin, the controller looks up a satisfaction-supply NPC's quest,
modifies its switch-class and gather steps, and starts gathering-only mode.
Only MIN and BTN are handled by that switch. Bad NPCs, unsupported jobs, or
unexpected path structure can throw; this is not a general item-gathering API.

`RedoLookup(uint questId) -> string` and
`RedoLookupIndex(uint questId) -> Tuple<string, int>` query New Game+ chapters.
The latter uses `System.Tuple`, unlike the value tuple returned by
`IsQuestLockedReason`. Inputs at or above `131072` return `""` or `("", -1)`;
lower inputs are cast to `ushort`. Verify the intended quest ID before calling.

There is no provider version endpoint in this snapshot. Disposal also omits
explicit unregistration of `IsQuestLockedReason` and `RedoLookupIndex`; do not
treat a surviving call-gate registration as proof of a healthy loaded plugin.

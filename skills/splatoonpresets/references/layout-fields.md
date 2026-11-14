# Layout fields (Splatoon 3.9.2.32, `Splatoon/Serializables/Layout.cs`)

| Field | Type | Default | Meaning / notes |
|---|---|---|---|
| `Enabled` | bool | true | master switch. Web API / `/splatoon enable|disable|toggle <Name>` flip it |
| `Name` | string | `""` | unique per config; `,` and `~` are replaced by `_` on import/edit; duplicate names block import unless CTRL held |
| `InternationalName` | `{En,Jp,De,Fr,Other}` | empty | localized name; gate: non-empty. Chinese authors use `Name` in CN and `InternationalName.En` |
| `Description` / `InternationalDescription` | string | `""` | free text shown in UI |
| `Group` | string | `""` | folder; user can disable whole groups (`DisabledGroups`) |
| `ZoneLockH` | `ushort[]` | [] | TerritoryType ids; gate: non-empty. Empty = all zones |
| `IsZoneBlacklist` | bool | false | invert zone list |
| `Scenes` | `int[]` | [] | must contain `*Scene.ActiveScene` (a byte at `EnvManager+36`); gate: non-empty |
| `DCond` | int | 0 | 0 Always, 1 combat, 2 instance, 3 combat AND instance, 4 combat OR instance, 5 On trigger only, 6 outside combat, 7 outside instance, 8 outside combat AND instance, 9 outside combat OR instance |
| `Phase` | int | 0 | 0 any; else must equal `P.Phase` (1 on zone entry; +1 on the "limit gauge resets" log line, LogMessage row 2844; DSR special-cased by DataId) |
| `JobLockH` | `Job[]` (ClassJob row ids) | [] | gate: non-empty. PLD 19 MNK 20 WAR 21 DRG 22 BRD 23 WHM 24 BLM 25 SMN 27 SCH 28 NIN 30 MCH 31 DRK 32 AST 33 SAM 34 RDM 35 BLU 36 GNB 37 DNC 38 RPR 39 SGE 40 VPR 41 PCT 42 (base classes 1–7, 26, 29) |
| `JobLock` | ulong | 0 | obsolete bitmask, migrated to `JobLockH` by `DataMigrator.MigrateJobs` on legacy import only |
| `DisableInDuty` | bool | false | hidden while `BoundByDuty` |
| `DisableDisabling` | bool | false | Web API cannot disable this layout |
| `Nodraw` | bool | false | processed (conditions, captures) but not rendered |
| `UseTriggers` | bool | false | enables `Triggers`; also required for `DCond 5` to hide anything |
| `Triggers` | `Trigger[]` | [] | gate: UseTriggers && non-empty. See below |
| `UseDistanceLimit` | bool | false | enables distance limit |
| `DistanceLimitType` | int | 0 | 0 distance player→current target (needs a target; hidden otherwise), 1 distance player→each element's position |
| `MinDistance`,`MaxDistance` | float | 0 | `min <= d < max`; gate: UseDistanceLimit |
| `DistanceLimitMyHitbox`,`DistanceLimitTargetHitbox` | bool | false | subtract hitboxes (type 0 only) |
| `ConditionalAnd` | bool | false | consecutive `Conditional` elements combine with AND (true) or OR (false) |
| `Freezing` | bool | false | snapshot mode |
| `FreezeFor` | float s | 0.1 | how long a snapshot stays; gate: Freezing |
| `IntervalBetweenFreezes` | float s | 10 | cooldown before the next snapshot (min 0.1 s; UI warns < 0.5) |
| `FreezeDisplayDelay` | float s | 0 | delay before a snapshot becomes visible |
| `FreezeResetTerr` | bool | true | clear snapshots on zone change |
| `FreezeResetCombat` | bool | true | serialized, **not read** in 3.9.2.32 |
| `ElementsL` | `Element[]` | [] | the default element set |
| `Elements` | dict | {} | obsolete; only migrated on legacy (`Name~json`) import, ignored for `~Lv2~` |
| `Subconfigurations` | `[{Guid,Name,Elements[]}]` | [] | alternative element sets; gate: non-empty |
| `SelectedSubconfigurationID` | Guid | `00000000-…` | which set is active; `Guid.Empty` = `ElementsL`; gate: Subconfigurations non-empty |
| `DefaultConfigurationName` | string | `""` | display name for the default set |
| `BlacklistedProjectorActions` | `uint[]` | [] | action ids projection must never draw; gate: non-empty |
| `ForcedProjectorActions` | `uint[]` | [] | action ids projection always draws; gate: non-empty |
| `ProjectionState` | bool? | null | serialized, not read anywhere |

## Trigger (`Splatoon/Serializables/Trigger.cs`)

| Field | Default | Meaning |
|---|---|---|
| `Type` | 0 | 0 Show at time in combat, 1 Hide at time, 2 Show at log message, 3 Hide at log message |
| `UserDisabled` | false | skip this trigger |
| `TimeBegin` | 0 | (0/1) seconds after combat start (`P.CombatStarted`; also reset by the limit-gauge message) |
| `Duration` | 0 | seconds the effect lasts; 0 = infinite (time triggers end immediately in state 2 but their condition persists until reset) |
| `Match` / `MatchIntl` | `""` | (2/3) text to find, case-insensitive substring |
| `IsRegex` | false | `Regex.IsMatch(message, Match)` |
| `MatchDelay` | 0 | seconds between match and effect |
| `FireOnce` | false | disable after first match until reset |
| `ResetOnTChange` | true | reset on territory change **and** on combat end (the only reset flag that is read) |
| `ResetOnCombatExit` | true | serialized, not read |

Trigger lists can be copied/pasted as a JSON array from the layout header.

## LayoutSubconfiguration

```json
{"Guid":"6f1c2b7e-2d3a-4c58-9d1e-8a5b0c9e4f21","Name":"Strat B","Elements":[ ...Element... ]}
```

Duplicate Guids are regenerated on the fly by the editor. Name them
consistently across layouts and scripts so the Configurations tab can switch a
whole zone to one strategy.

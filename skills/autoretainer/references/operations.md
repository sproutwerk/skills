# Commands and troubleshooting

The [command handler](source/AutoRetainer/AutoRetainer/AutoRetainer.cs) is the
authority for this snapshot. `/ays` aliases `/autoretainer`. The commands below
are documentation, not diagnostic actions to run automatically.

## Common commands

| Command after `/autoretainer` | Behavior at this pin |
| --- | --- |
| no arguments | Toggle main window |
| `settings` / `s` | Open settings |
| `browser` / `b` | Toggle venture browser |
| `log` / `l` | Toggle plugin log window |
| `debug` | Toggle verbose/debug mode |
| `enable` / `e`, `disable` / `d`, `toggle` / `t` | Change retainer scheduler state |
| `multi` / `m` | Toggle MultiMode |
| `multi enable`, `multi disable` | Explicitly change MultiMode state |
| `multi single` | Enable MultiMode if disabled; does not initialize single-pass state here |
| `night enable` | Enable NightMode and enable MultiMode if needed |
| `night set` | Set NightMode without enabling MultiMode |
| `night disable` | Clear NightMode; does not itself disable MultiMode |
| `relog Character Name@WorldName` | Request a relog to an exact cached name/world match |
| `reset` | Abort main queue and clear the character postprocessing lock |

The plugin also has commands for selling, discarding, deliveries, house entry,
and scheduled game shutdown. Inspect the exact handler branch before using one.
Do not treat `disable`, `reset`, or `night disable` as interchangeable stop
commands. A relog request succeeding does not establish that loading finished.

## Stop controls

| Control | What the implementation changes |
| --- | --- |
| `AutoRetainer.SetSuppressed(true)` | Shared suppression flag; retainer scheduler and `MultiMode.Active` consult it |
| `/autoretainer disable` | Retainer scheduler enabled flag |
| `/autoretainer multi disable` | MultiMode enabled flag |
| `AutoRetainer.PluginState.DisableAllFunctions` | MultiMode, retainer scheduler, and voyage scheduler flags |
| `AutoRetainer.PluginState.AbortAllTasks` | Main task manager queue |
| `/autoretainer reset` | Main task manager queue and `CharacterPostProcessLocked` |

Disabling flags does not itself abort the main queue. Aborting the queue does
not itself disable the schedulers, so later ticks can schedule more work.
Suppression is not an abort mechanism; voyage processing has its own framework
callback, and the shown voyage path does not check `IPC.Suppressed`.
`reset` does not explicitly clear `RetainerPostProcessLocked`.

For a requested stop of scheduler-driven work, distinguish the flags from the
already queued actions and inspect both. Other plugins, a GC operation, or
Lifestream can still own activity. Do not announce that all automation stopped
just because one boolean changed.

## Diagnose a stalled run

Start with the reported command, expected next UI state, target version, and
the relevant log/queue entry. Use an existing log or state view; enabling
MultiMode or interacting with a bell is not a read-only test.

| Symptom | Inspect |
| --- | --- |
| API unavailable | `AutoRetainerApi.Ready`, plugin load state, and whether the requested newer channel exists at the target pin |
| Enabled but no retainer chosen | Suppression, visible `RetainerList`, main queue idle state, inventory space, selected names, cached character record, venture time and eligibility |
| Venture override ignored | Callback timing, `RetainerTask` ID, matching current character and name, and whether another consumer writes an override |
| Postprocessing never finishes | Last granted consumer, its queue/error log, and the matching lock; compare request, ready, and finish events |
| MultiMode does not switch | `Enabled` versus `Active`, cached enable/exclusion flags, selected mode, wait-for-all/advance timers, and auto-login prerequisites |
| Workshop processing stops | Workshop enable state, enabled vessel names, panel state, fuel/repair materials, inventory, and permissions reported by the plugin |
| Wrong or missing config value | Target provider fields versus API getters, DTO assembly revision, CID/name identity, and same-frame write checks |

`GetNextRetainerName()` requires a selected name and remaining venture time at
or below `Config.UnsyncCompensation` (default `-5` in this snapshot). It also
requires an existing venture, permission to assign quick exploration, or an
enabled nonempty planner. A visible finished timer alone does not explain all
selection conditions. `CanAssignQuickExploration` additionally checks the
configuration, `DontReassign`, and venture currency.

The main `NeoTaskManager` is constructed with a 20-second timeout and
`abortOnTimeout: true`. Offline-data work uses a separate 60-second task manager;
postprocessing waits explicitly override the normal timeout. Read the named
task's predicate and addon expectations before raising time limits or changing
callback arguments.

Workshop errors can change selection state: the
[voyage handler](source/AutoRetainer/AutoRetainer/Modules/Voyage/VoyageMain.cs)
disables workshop processing for inventory or permission errors and removes
an affected vessel from the enabled set when repair materials are exhausted.
Fix the reported condition before considering a requested retry.

## Code changes and validation

Trace a change through the
[source map](provenance.md#source-map) and the matching upstream checkout.
The bundled files are reference excerpts, not a buildable copy of the plugin.
For a consumer, run this from the directory containing its `.csproj`, using
the project's existing dependencies:

```sh
dotnet build -c Debug
```

For AutoRetainer itself, inspect its csproj and submodule
pins first; at this pin it targets .NET 10 on Windows and Dalamud API 15.
The default build uses Dalamud's ClientStructs assembly, while `CustomCS=true`
selects the source project.

On a missing member or assembly error, compare versions before changing a
signature. A successful compile checks the managed contract; it does not verify
native UI layouts against a running game patch. Report source review, compilation,
and live testing separately. Do not run live game actions just to validate this
documentation.

# Source pins and navigation

Reviewed on 2026-11-13. The technical notes describe these source revisions:

| Project | Revision | Version metadata |
| --- | --- | --- |
| [AutoRetainer](https://github.com/PunishXIV/AutoRetainer/tree/cc020df0efee3ea65a45117e2b89999108c702e8) | `cc020df0efee3ea65a45117e2b89999108c702e8` | Plugin `4.6.1.26`, Dalamud API 15 |
| [AutoRetainerAPI](https://github.com/PunishXIV/AutoRetainerAPI/tree/7ccf0f6b4c7923821a43ed1e92456c9d5d7132f2) | `7ccf0f6b4c7923821a43ed1e92456c9d5d7132f2` | Assembly `1.0.0.4` |

On the review date AutoRetainer's master branch still resolved to the recorded
revision, and AutoRetainerAPI's main branch matched the plugin's gitlink. The
newest upstream release tag, `4.5.2.6`, is older than this snapshot; version
`4.6.1.26` comes from the project file. The plugin commit date is 2026-08-28. That date and the review date
do not establish compatibility with the user's installed game build.

## Dependency context

Both projects target `net10.0-windows7.0`, x64, with preview C# and unsafe code.
AutoRetainer directly references Dalamud libraries and uses `DALAMUD_HOME` for
non-Windows/CI builds. Its default `CustomCS=false` references the ClientStructs
assembly from Dalamud; the recorded ClientStructs gitlink is used only by the
optional source build.

The plugin has both an ECommons project reference and a wildcard package
reference with excluded assets. AutoRetainerAPI declares ECommons `3.2.0.11`.
Do not infer dependency compatibility from the API assembly version alone.
Inspect the target project's full dependency graph when compiling a consumer.

AutoRetainer records these submodule pins:

| Submodule | Revision |
| --- | --- |
| `AutoRetainerAPI` | `7ccf0f6b4c7923821a43ed1e92456c9d5d7132f2` |
| `ECommons` | `e6be8f0fd7786a9e1781db2e71cb5b9146f04980` |
| `ECommons.FileWriter` | `3509b662f4effb21cbad4166cfebf8ce98b1ebb5` |
| `ECommons.IPC` | `90986f272fbc45b2d7db0771a9b884d5e69ae654` |
| `FFXIVClientStructs` | `0769d1f180f859688f47a7a99610e9ce10da946c` |
| `NightmareUI` | `aae510d183a6114baece32394a9359f002a33a5d` |
| `NightmareUI.OtterGuiWrapper` | `371b1f90633394805244b50f1c135e145c6aa98d` |
| `OtterGui` | `a08fda67dfb731f007234c78df32a2b3ddccc450` |
| `PunishLib` | `96c0cefa8ebf948b3d3d71d85994c61dc8430b5c` |

Only AutoRetainerAPI source is bundled from those submodules. ECommons at the
recorded pin was inspected to verify EzIPC's `prefix.methodName` registration:
[EzIPC.cs](https://github.com/NightmareXIV/ECommons/blob/e6be8f0fd7786a9e1781db2e71cb5b9146f04980/ECommons/EzIpcManager/EzIPC.cs).
That dependency source is not copied into this skill.

## Source map

Paths below are relative to `references/`. The curated snapshot supports
lookups; it omits most UI, native interop, task handlers, resources, and build
dependencies. Follow missing calls in a full checkout at the matching revision.

| Question | File or directory |
| --- | --- |
| Wrapper lifetime, readiness, data checks, and methods | [AutoRetainerApi.cs](source/AutoRetainerAPI/AutoRetainerAPI/AutoRetainerApi.cs) |
| Event filtering and callback signatures | [AutoRetainerApiInternal.cs](source/AutoRetainerAPI/AutoRetainerAPI/AutoRetainerApiInternal.cs), [Delegates.cs](source/AutoRetainerAPI/AutoRetainerAPI/Delegates.cs) |
| Exact legacy event strings | [ApiConsts.cs](source/AutoRetainerAPI/AutoRetainerAPI/ApiConsts.cs) |
| Legacy providers and write semantics | [IPC.cs](source/AutoRetainer/AutoRetainer/Modules/IPC.cs) |
| Busy, stop, MultiMode and relog controls | [IPC_PluginState.cs](source/AutoRetainer/AutoRetainer/Modules/EzIPCManagers/IPC_PluginState.cs) |
| Config getters and provider registration | [AutoRetainerConfig.cs](source/AutoRetainerAPI/AutoRetainerAPI/AutoRetainerConfig.cs), [IpcConfigValuesProvider.cs](source/AutoRetainer/AutoRetainer/Services/IpcConfigValuesProvider.cs) |
| Config fields and DTO definitions | [Config.cs](source/AutoRetainer/AutoRetainer/PluginData/Config.cs), [Configuration/](source/AutoRetainerAPI/AutoRetainerAPI/Configuration/) |
| Commands, initialization, teardown, and framework tick | [AutoRetainer.cs](source/AutoRetainer/AutoRetainer/AutoRetainer.cs) |
| Retainer eligibility, override ordering, and planner completion | [SchedulerMain.cs](source/AutoRetainer/AutoRetainer/Scheduler/SchedulerMain.cs) |
| Request/grant/wait postprocessing sequence | [TaskPostprocessRetainerIPC.cs](source/AutoRetainer/AutoRetainer/Scheduler/Tasks/TaskPostprocessRetainerIPC.cs), [TaskPostprocessCharacterIPC.cs](source/AutoRetainer/AutoRetainer/Scheduler/Tasks/TaskPostprocessCharacterIPC.cs) |
| Linked plan UI and index changes | [VenturePlanner.cs](source/AutoRetainer/AutoRetainer/UI/Windows/VenturePlanner.cs) |
| Retainer list IPC button behavior | [RetainerListOverlay.cs](source/AutoRetainer/AutoRetainer/UI/Overlays/RetainerListOverlay.cs) |
| Selection, relog, and single-pass state | [MultiMode.cs](source/AutoRetainer/AutoRetainer/Modules/Multi/MultiMode.cs) |
| Cached state refresh | [OfflineDataManager.cs](source/AutoRetainer/AutoRetainer/Modules/OfflineDataManager.cs) |
| Timing, additional-data keys, planner helpers, busy state | [Utils.cs](source/AutoRetainer/AutoRetainer/Helpers/Utils.cs) |
| Venture sheets, job adjustment, availability | [VentureUtils.cs](source/AutoRetainer/AutoRetainer/Helpers/VentureUtils.cs) |
| Workshop event handling, scheduling, and readiness | [VoyageMain.cs](source/AutoRetainer/AutoRetainer/Modules/Voyage/VoyageMain.cs), [VoyageScheduler.cs](source/AutoRetainer/AutoRetainer/Modules/Voyage/VoyageScheduler.cs), [VoyageUtils.cs](source/AutoRetainer/AutoRetainer/Modules/Voyage/VoyageUtils.cs) |

For example, from the skill directory, locate the override decision with:

```sh
rg -n 'VentureOverride|FireSendRetainerToVentureEvent' references/source/AutoRetainer/AutoRetainer/Scheduler/SchedulerMain.cs
```

Read the function around a match. A name appearing in source does not establish
that it is public, registered, or reachable in the current game state.

## Snapshot maintenance and verification limits

[The file manifest](source-manifest.json) records each copied file's upstream
path, repository, revision, and SHA-256. Files were copied unchanged, including
upstream README and license texts. Authoring guidance is separate from the
snapshot. [NOTICE.md](../NOTICE.md) explains the applicable licenses.

For a refresh, retrieve the intended plugin revision and its AutoRetainerAPI
gitlink, compare the provider, wrapper, DTOs and scheduler call sites, and update
the authored notes only for verified behavior. Update the file manifest and any
repository-level source record together. Preserve upstream notices and hashes of the
new snapshot; do not treat a dependency's newest branch as the plugin's pin.

This skill was reviewed against source. Its C# examples were not compiled or
run in game. No live game actions were used to validate it. Local checks cover
reference paths, manifest agreement, and copied-file integrity. No security
audit is claimed.

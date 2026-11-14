# Source pins and verification

Reviewed on 2026-11-13 against
[awgil/ffxiv_bossmod](https://github.com/awgil/ffxiv_bossmod/tree/c50e67f37aa22f86d064489913ec9887f116accf),
commit `c50e67f37aa22f86d064489913ec9887f116accf`, the `master` head on that
date and the commit tagged `v7.5.8.2`. [Common.targets](source/Common.targets)
declares version `7.5.8.2` for every project; the plugin project no longer
carries its own version. The manifest names `BossMod` / `Boss Mod` and author
`veyn`, and the SDK reference is `Dalamud.NET.Sdk/15.0.0`.

Upstream can change multiple times daily, including breaking changes. This
snapshot will lag and cannot promise exact parity with an installed plugin.
The skill shortcuts understanding of the integration architecture and shows
where to verify contracts. It is not continuously synchronized.

## Durable guidance and revision-specific contracts

Keep state ownership explicit, use provider implementations to verify IPC,
dispose subscriptions with their owner, and separate replay predictions from
live execution. Those practices apply across revisions. Channel availability,
generic types, return behavior, module constructors, scheduling, and saved-data
schemas are revision-specific; recheck the files below at the user's target
commit before relying on them.

| Change under consideration | Files to recheck |
| --- | --- |
| IPC call or lifecycle | [IpcService](source/BossMod/Services/IpcService.cs), [Plugin](source/BossMod/Framework/Plugin.cs) |
| Config write/persistence | [ConfigRoot](source/BossMod/Config/ConfigRoot.cs), [SaveConfigService](source/BossMod/Services/SaveConfigService.cs) |
| Preset names, JSON, writes | [PresetDatabase](source/BossMod/Autorotation/PresetDatabase.cs), [Preset](source/BossMod/Autorotation/Preset.cs), [PlanPresetConverter](source/BossMod/Autorotation/PlanPresetConverter.cs) |
| Automation selection/stop behavior | [RotationModuleManager](source/BossMod/Autorotation/RotationModuleManager.cs), [SlashCommandService](source/BossMod/Services/SlashCommandService.cs), [AIConfig](source/BossMod/AI/AIConfig.cs), [AIWindow](source/BossMod/AI/AIWindow.cs) |
| Encounter construction/registration | [BossModuleRegistry](source/BossMod/BossModule/BossModuleRegistry.cs), [BossModule](source/BossMod/BossModule/BossModule.cs), [ModulePackService](source/BossMod/Services/ModulePackService.cs) |
| Mechanic events and phase lifetime | [BossComponent](source/BossMod/BossModule/BossComponent.cs), [StateMachineBuilder](source/BossMod/BossModule/StateMachineBuilder.cs), [StateMachine](source/BossMod/BossModule/StateMachine.cs) |
| Live/update/replay behavior | [TickService](source/BossMod/Services/TickService.cs), [WorldState](source/BossMod/Data/WorldState.cs), [ReplayPlayer](source/BossMod/Replay/ReplayPlayer.cs), [ReplayDetailsWindow](source/BossMod/Replay/Visualization/ReplayDetailsWindow.cs) |
| Hint contracts | [AIHints](source/BossMod/BossModule/AIHints.cs), [AIHintsBuilder](source/BossMod/BossModule/AIHintsBuilder.cs), [GenericAOEs](source/BossMod/Components/GenericAOEs.cs), [Vector](source/BossMod/Util/Vector.cs), [Sdf](source/BossMod/Util/Sdf.cs) |
| Rotation definitions/registry | [RotationModule](source/BossMod/Autorotation/RotationModule.cs), [RotationModuleRegistry](source/BossMod/Autorotation/RotationModuleRegistry.cs) |
| Obstacle-map integration | [ObstacleMapManager](source/BossMod/Pathfinding/ObstacleMapManager.cs), plus the target vnavmesh provider |

These are selected source files, not a buildable checkout. Follow missing
calls in the upstream tree at the same commit. Paths retain upstream layout
under `references/source/`. [source-manifest.json](source-manifest.json)
records each copied path, upstream path/repository/revision, Git blob ID, and
SHA-256. Source files were copied directly from pinned Git objects, preserving
BOMs and line endings. Authored guides are separate from that snapshot.

## Dependencies

The upstream git tree and [.gitmodules](source/.gitmodules) record:

| Submodule | Repository | Gitlink revision |
| --- | --- | --- |
| `FFXIVClientStructs` | `https://github.com/aers/FFXIVClientStructs` | `fb26505058f33791e31c029e3678dbe70378b972` |
| `DalaMock` | `https://github.com/xanunderscore/DalaMock.git` | `814fc7c10adde7c496d2df0929b2c118bc29ff31` |

[Common.targets](source/Common.targets) sets `LocalCS=false` and
`LocalDalamock=false`. LocalCS enables source project references and custom
resolver initialization; the default build uses Dalamud's ClientStructs.
DalaMock.Core `6.2.2` is a Debug package dependency unless LocalDalamock is true.
The plugin also references DalaMock.Host/Shared `6.2.2`. These dependencies are
not bundled here. Do not substitute another skill's ClientStructs pin for the
libraries the target runtime supplies.

[global.json](source/global.json),
[BossMod.csproj](source/BossMod/BossMod.csproj), and
[Module.targets](source/Module.targets) preserve SDK and build-target evidence.
A source pin does not establish compatibility with a later game build or
Dalamud release.

## Refresh procedure

See [fork-boundary.md](fork-boundary.md) for provider identification and porting
checks. This skill requires no companion rotation-plugin skill.

For a refresh, retrieve the intended VBM commit, inspect its gitlinks and
project references, then compare registrations and called implementations.
Check component dispatch, state-machine fallback, preset selection, and data
serialization where the task touches them. Update the affected guide, full
pin/review date, hash manifest, and repository `sources.json` together. Preserve
unchanged source/license notices and inspect newly selected files for separate
attribution. Do not describe a documentation refresh as runtime validation.

## Validation limits

Examples and behavior notes were reviewed against source. C# examples were
not compiled; replay playback, plugin loading, and live IPC were not run. No
live game actions were used. Local validation covers frontmatter, names,
reference links and anchors, and copied file hashes against Git objects.

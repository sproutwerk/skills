# Source pins and navigation

Reviewed on 2026-11-13. Upstream can change multiple times daily, including
breaking changes. These pins identify the evidence inspected; they do not
establish continuous synchronization or exact parity with a user's build.
Use the architectural guidance to locate the contract, then recheck the
target revision before depending on an endpoint, signature or side effect.

The vnavmesh and DotRecast rows were refreshed on that date. Between `6fc80725`
(1.2.3.14) and this pin, the bundled files changed only in the version number
and a debug-only `NavmeshManager.Envelope` clipboard dump; `IPCProvider.cs` is
identical. The BossMod and Lifestream caller pins were last inspected on
2026-10-25 and are kept as fixed comparison points.

| Source | Revision | Role |
| --- | --- | --- |
| [awgil/ffxiv_navmesh](https://github.com/awgil/ffxiv_navmesh/tree/38d2a24c3306f24640e64efe13915a78d301e878) | `38d2a24c3306f24640e64efe13915a78d301e878` | vnavmesh `1.2.3.15`, main source snapshot |
| [xanunderscore/DotRecast](https://github.com/xanunderscore/DotRecast/tree/8002b30b9f196bcd9eb0a898e51abcea4177fb04) | `8002b30b9f196bcd9eb0a898e51abcea4177fb04` | Plugin's exact dependency gitlink |
| [awgil/ffxiv_bossmod](https://github.com/awgil/ffxiv_bossmod/tree/93b96ac6c0f75dfc97ee0148c5c5141c785a3ec9) | `93b96ac6c0f75dfc97ee0148c5c5141c785a3ec9` | Inspected caller/comparison, no source copied |
| [NightmareXIV/Lifestream](https://github.com/NightmareXIV/Lifestream/tree/b36320170faafc8315f7585edb45cfbcd34e7046) | `b36320170faafc8315f7585edb45cfbcd34e7046` | Inspected caller, no source copied |

## Dependency context

The [project](source/vnavmesh/vnavmesh/vnavmesh.csproj) specifies
`Dalamud.NET.Sdk/15.0.0`, plugin version `1.2.3.15`, SharpDX.D3DCompiler,
SharpDX.Direct3D11 and SharpDX.Mathematics `4.2.0`, and a non-private
`Dalamud.Common` assembly reference. It references DotRecast.Core, Detour,
Detour.Extras and Recast projects from its single
[submodule](source/vnavmesh/.gitmodules). DotRecast.Core's
[project metadata](source/DotRecast/src/DotRecast.Core/DotRecast.Core.csproj)
lists `netstandard2.1;net6.0;net7.0;net8.0`.

The vnavmesh project leaves target framework defaults to the SDK. It does not
pin a separate FFXIVClientStructs Git revision in this tree. Resolve the SDK's
actual runtime assemblies and game compatibility when building; a sibling
skill's ClientStructs pin is not this plugin's dependency pin. No package restore
or build was performed for this documentation task.

The seed URL in `FloodFill.cs` follows a mutable branch. Its runtime data and
the user's local seeds/cache are outside this snapshot. DotRecast's gitlink was
checked out and inspected at the recorded commit, not at the fork's current HEAD.

## Source map

Paths are relative to this reference directory. Each copied upstream file is
listed in [source-manifest.json](source-manifest.json), with its original path,
repository, full revision and SHA-256. Copies come directly from Git objects,
preserving bytes, line endings and notices. This is a selected reference set,
not a buildable plugin checkout.

| Contract to verify | Source |
| --- | --- |
| Channel spelling, generic argument order, function versus action | [IPCProvider.cs](source/vnavmesh/vnavmesh/IPCProvider.cs) |
| Readiness, cache/build lifecycle, serialized query queue, cancellation | [NavmeshManager.cs](source/vnavmesh/vnavmesh/NavmeshManager.cs) |
| Empty/partial paths, tolerance, mesh projections, avoidance | [NavmeshQuery.cs](source/vnavmesh/vnavmesh/NavmeshQuery.cs) |
| Acceptance versus later movement, pending-task guard, stuck retry | [AsyncMoveRequest.cs](source/vnavmesh/vnavmesh/AsyncMoveRequest.cs) |
| Waypoint consumption, client paths, takeoff, shared flags, stop | [FollowPath.cs](source/vnavmesh/vnavmesh/Movement/FollowPath.cs) |
| Movement hooks, user-input checks, disposal | [OverrideMovement.cs](source/vnavmesh/vnavmesh/Movement/OverrideMovement.cs) |
| Framework update order, commands, provider teardown | [Plugin.cs](source/vnavmesh/vnavmesh/Plugin.cs) |
| Default settings and persistence | [Config.cs](source/vnavmesh/vnavmesh/Config.cs) |
| World vector conversion and map-flag projection | [Extensions.cs](source/vnavmesh/vnavmesh/Extensions.cs), [MapUtils.cs](source/vnavmesh/vnavmesh/MapUtils.cs) |
| Mesh/volume creation and territory defaults | [NavmeshBuilder.cs](source/vnavmesh/vnavmesh/NavmeshBuilder.cs), [NavmeshCustomization.cs](source/vnavmesh/vnavmesh/NavmeshCustomization.cs) |
| Remote/local seed loading | [FloodFill.cs](source/vnavmesh/vnavmesh/FloodFill.cs) |

For unbundled implementations, use the matching full checkout. In particular,
avoidance edge cases in flight require `vnavmesh/NavVolume/VoxelPathfind.cs`;
territory overrides require the corresponding `vnavmesh/Customizations/` file.
An absent file in this selected set does not prove the API is absent upstream.

## Consumer evidence

These callers establish usage patterns, not a compatibility promise or an example
to copy without reviewing error handling:

- BossMod's [QuestBattle.cs](https://github.com/awgil/ffxiv_bossmod/blob/93b96ac6c0f75dfc97ee0148c5c5141c785a3ec9/BossMod/BossModule/QuestBattle.cs)
  binds readiness and asynchronous position queries, then consumes results in
  its own quest navigation. Its
  [MovementOverride.cs](https://github.com/awgil/ffxiv_bossmod/blob/93b96ac6c0f75dfc97ee0148c5c5141c785a3ec9/BossMod/Framework/MovementOverride.cs)
  reads `vnav.PathIsRunning`; that shared boolean is not an ownership protocol.
- BossMod's [ObstacleMapManager.cs](https://github.com/awgil/ffxiv_bossmod/blob/93b96ac6c0f75dfc97ee0148c5c5141c785a3ec9/BossMod/Pathfinding/ObstacleMapManager.cs)
  binds `Nav.BuildBitmapBounded`, a file-writing export rather than movement.
- Lifestream's [VnavmeshIPC.cs](https://github.com/NightmareXIV/Lifestream/blob/b36320170faafc8315f7585edb45cfbcd34e7046/Lifestream/IPC/VnavmeshIPC.cs)
  maps its EzIPC delegates to the provider. Its `PathfindInProgress` delegate
  uses the `SimpleMove` prefix, so it does not describe all raw manager queries.
- Lifestream's [TaskGeneratePath.cs](https://github.com/NightmareXIV/Lifestream/blob/b36320170faafc8315f7585edb45cfbcd34e7046/Lifestream/Tasks/Utility/TaskGeneratePath.cs)
  polls the returned task and checks `IsCompletedSuccessfully` before reading
  results. Its [CustomAliasCommand.cs](https://github.com/NightmareXIV/Lifestream/blob/b36320170faafc8315f7585edb45cfbcd34e7046/Lifestream/Data/CustomAliasCommand.cs)
  can pass queried positions into Lifestream's own follower. Its
  [FollowPath.cs](https://github.com/NightmareXIV/Lifestream/blob/b36320170faafc8315f7585edb45cfbcd34e7046/Lifestream/Movement/FollowPath.cs)
  clears its own waypoints when it detects vnavmesh movement; its
  [movement notice](https://github.com/NightmareXIV/Lifestream/blob/b36320170faafc8315f7585edb45cfbcd34e7046/Lifestream/Movement/README.md)
  credits awgil.

## Refresh and verification

Keep the durable consumer design separate from revision-specific observations.
Refresh the main pin and dependency gitlinks together; inspect provider bindings
and follow each changed call into its implementation. Recheck caller mappings
without assuming a wrapper uses the prefix its method name suggests. Review
file-level attribution and license changes before replacing source copies.

Validate copied hashes against the pinned Git objects, and check local links
and frontmatter. The C# snippet and movement state machine were reviewed
against source only. Compilation, game behavior, native signatures and arrival
were not tested. Structural validation does not establish runtime compatibility.
See [NOTICE.md](../NOTICE.md) for the unresolved vnavmesh source grant.

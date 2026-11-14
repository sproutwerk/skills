# Navigation and troubleshooting

These implementation details describe [vnavmesh 1.2.3.15](provenance.md).
Recheck the target source after an update. IPC availability, mesh availability,
route computation, movement and arrival are separate states.

## Loading and building

[NavmeshManager](source/vnavmesh/vnavmesh/NavmeshManager.cs) detects changes to
the active layout key during framework updates. The key includes background,
filter, festivals and enabled shared groups. A territory ID alone is not the
complete mesh identity. Loading waits for an initialized layout and can wait
through cutscenes. `AutoLoadNavmesh` defaults true.

`Reload` allows reading the matching cache under the plugin configuration's
`meshcache` directory. If cache loading fails, it builds from the active scene.
`Rebuild` bypasses reading that cache and writes the rebuilt result. Both clear
state asynchronously and return true without waiting for success; no current
key can mean no build was queued. Reloading with autoload disabled and an empty
current key does not establish a mesh. Do not change the user's autoload setting
as an availability probe.

`Nav.IsReady` checks only `Navmesh != null`. During a queued clear it can briefly
describe the old mesh. During loading it can become true before seed-based
pruning finishes and before the mesh-change event. `BuildProgress < 0` means
the load task is not running, which also occurs after failure; it does not mean
success. Use fresh character/layout context, readiness, progress, and bounded
query outcomes together. There is no exported current-mesh key or fully-loaded
event in this provider.

Generation uses scene geometry, territory customizations and DotRecast. The
builder optionally creates a flight volume. Cache deserialization checks the
customization version. [FloodFill](source/vnavmesh/vnavmesh/FloodFill.cs) combines
local `seeds-local.json` with remote `seeds/seeds.json` before marking unreachable
polygons. The remote seed branch is mutable and is not pinned by the plugin's
commit; identical code does not guarantee identical runtime seed data. Diagnose
the seed-fetch/build logs before treating a missing ready state as a movement
error. Loading or rebuilding the plugin can make network requests and write
cache files; reading this reference requires neither action.

## Coordinates and projections

IPC positions are `System.Numerics.Vector3` in game world coordinates: X and Z
are horizontal, Y is height. [Extensions](source/vnavmesh/vnavmesh/Extensions.cs)
copies X/Y/Z unchanged between System.Numerics and DotRecast. Do not pass map
display coordinates, swap Y/Z to match a drawing convention, or reinterpret
a sheet's map coordinates as world coordinates.

Copy a live object's position only while its player/territory context is valid.
For stored destinations, store the territory and intended height/floor with
the position. Projection can select a nearby surface rather than the intended
floor; reject an unexpectedly large projection displacement.

`FlagToPoint` reads the first flag marker's X/Z values, starts at Y=1024, and
projects down to a mesh floor. [MapUtils](source/vnavmesh/vnavmesh/MapUtils.cs)
does not compare the flag's territory with the loaded mesh. Validate that context
in the consumer, especially when the user sets a flag in another zone.

`/vnav moveto` and `/vnav flyto` parse raw X Y Z values. At this pin the
`movedir`/`flydir` implementation adds the supplied vector to the player's world
position without rotating it by facing, despite the command help describing
facing-relative motion. Prefer typed IPC and check implementation when command
help and behavior disagree. `/vnavmesh` is the legacy alias for `/vnav`.

## Ground and flight

Ground queries use the navigation mesh. Flight uses the voxel volume; a mesh
can be ready while a volume is absent, in which case the flight query returns
an empty route. The default customization enables volume generation for
territory intended-use values 1, 49 and 47; zone customizations can override it.
This is a build decision, not proof that the player has unlocked flight or can
fly in the present state.

Ground following sets `IgnoreDeltaY` true for ordinary waypoint consumption.
Flight following preserves height. When a flight waypoint is above the player
and the player is neither flying nor diving, the follower tries takeoff jumps
if mounted; if unmounted, it disables movement for that update. It does not
mount the character. Mount/flight eligibility belongs in the consumer's requested
workflow, with separate checks and deadlines.

`SimpleMove.PathfindAndMoveCloseTo` supplies range both to the ground query
heuristic and to the follower's destination tolerance. The latter compares 3D
distance to the final waypoint even on a ground path. Flight query generation
does not use the range, though SimpleMove's follower still does. `Path.SetTolerance`
changes ordinary waypoint tolerance and is a different setting.

`PathfindAvoid` has explicit center/radius inputs. Ground filtering approximates
intersecting polygons by discs in XZ and is only enabled when the start/end
segment enters the avoidance region. Flight has its own volume search. Do not
treat either as dynamic enemy/AOE avoidance or a complete exclusion guarantee;
inspect `SegmentEntersAvoid`, `AvoidRadiusFilter` and the target volume search
when the requested behavior depends on their edge cases.

## Failure handling

| Observation | Next check / response |
| --- | --- |
| Missing channel or type mismatch | Compare the exact provider signature and installed version; report unavailable state |
| Never ready | Check active layout, autoload setting, cutscene, cache/build and seed-fetch logs; expire the readiness budget |
| Reload returned true | Continue observing; it has not confirmed a usable mesh |
| Query task faults | Observe/log the exception, including wrapped cancellation; invalidate the run |
| Successful empty route | Check start/end mesh polygons or flight voxels/volume; do not start movement |
| Nonempty route stalls | Check partial routes, coordinates/floor, flight eligibility, client-path metadata and movement ownership |
| Path idle before arrival | Treat as stopped/failed; check user-input/stuck settings and pending SimpleMove |
| Stop followed by movement | Look for a pending SimpleMove result or stuck retry; apply owned cleanup from the consumer reference |

Do not repeatedly call SimpleMove every frame. Requests can replace another
path, false can mean an existing query is still pending, and automatic stuck
retry has no attempt cap in `AsyncMoveRequest`. Use the consumer's deadline and
a finite retry policy. Retry a corrected query only after its previous task
settles and context is revalidated. Do not fall back from an empty path to a
straight `Path.MoveTo([destination], ...)` route through unverified geometry.

## BossMod and other movement systems

vnavmesh computes routes through world geometry and follows waypoints. BossMod
has its own encounter hints, combat navigation and movement overrides. In the
reviewed BossMod pin, quest navigation calls `vnavmesh.Nav.Pathfind` and then
consumes positions through BossMod's own navigation. Obstacle-map code can call
vnavmesh's bitmap export functions. Neither use implies SimpleMove is active.

BossMod's `MovementOverride` checks shared `vnav.PathIsRunning` data to suppress
part of its movement override while a vnavmesh path runs. That cooperation is
revision-specific, not an ownership guarantee across every plugin or fork.
Lifestream also queries vnavmesh but can follow positions itself; its follower
checks for active vnavmesh movement and clears its own waypoints on a conflict.
See the [pinned caller links](provenance.md#consumer-evidence).

Do not run two movement owners concurrently based solely on their using the
same pathfinding provider. Coordinate ownership in the consumer workflow. This
skill requires no companion rotation-plugin skill.
The `vnavmesh.*` channels belong to the FFXIV plugin; general Recast/Detour,
Unity navigation APIs and unrelated navigation libraries do not export them.

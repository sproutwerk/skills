# IPC contracts

The [provider](source/vnavmesh/vnavmesh/IPCProvider.cs) prefixes every channel
below with `vnavmesh.`. Functions use `InvokeFunc`; actions use `InvokeAction`
and an `object` return placeholder in `GetIpcSubscriber`. The final generic
argument is the return type. Check `HasFunction` or `HasAction` as appropriate
and still handle exceptions if the provider unloads between check and call.
No initialization, shutdown or arrival broadcast is registered at this pin.

## Navigation state and queries

| Channel suffix | Parameters | Return / behavior |
| --- | --- | --- |
| `Nav.IsReady` | none | `bool`: manager's mesh is non-null |
| `Nav.BuildProgress` | none | `float`: negative when load task is not running; otherwise nominally 0..1 |
| `Nav.Reload` | none | `bool`: clear state, queue load allowing cache; returns true even if there is no current layout key |
| `Nav.Rebuild` | none | `bool`: clear state, queue build without reading cache |
| `Nav.Pathfind` | `Vector3 from, Vector3 to, bool fly` | `Task<List<Vector3>>` |
| `Nav.PathfindWithTolerance` | `Vector3 from, Vector3 to, bool fly, float range` | `Task<List<Vector3>>`; ground search heuristic uses range; flying query ignores this range |
| `Nav.PathfindAvoid` | `Vector3 from, Vector3 to, bool fly, Vector3 avoidCenter, float avoidRadius` | `Task<List<Vector3>>`; explicit avoidance parameters, not combat awareness |
| `Nav.PathfindCancelable` | `Vector3 from, Vector3 to, bool fly, CancellationToken cancel` | `Task<List<Vector3>>`; caller cancellation linked to mesh lifetime |
| `Nav.PathfindCancelAll` | none | Action: calls `Reload(true)`; global cancellation and mesh reload |
| `Nav.PathfindInProgress` | none | `bool`: active manager query count is positive |
| `Nav.PathfindNumQueued` | none | `int`: `max(activeCount - 1, 0)` |
| `Nav.IsAutoLoad` | none | `bool`: shared configuration |
| `Nav.SetAutoLoad` | `bool` | Action: persist shared configuration |

There is no single exported query accepting cancellation, tolerance and avoidance
together. Keep the exact overload's capabilities when designing a wrapper.
The manager serializes mesh work and queries. Its count includes pending work;
it is not a per-consumer queue or a reliable admission lock. Prefer one in-flight
query per consumer and observe that task's completion directly.

The task may fault, be canceled, or succeed with an empty list. Mesh loading can
fail after a query is accepted. Ground pathfinding can return a partial corridor,
and the implementation appends the requested destination even when the partial
path clamping code is commented out. A nonempty list ending at the requested
position therefore does not certify reachability. See
[NavmeshQuery](source/vnavmesh/vnavmesh/NavmeshQuery.cs).

## Following and SimpleMove

| Channel suffix | Parameters | Return / behavior |
| --- | --- | --- |
| `Path.MoveTo` | `List<Vector3> waypoints, bool fly` | Action: replaces the current path with default-type waypoints; no obstacle query |
| `Path.Stop` | none | Action: clears waypoints and stuck counter; pending SimpleMove is unaffected |
| `Path.IsRunning` | none | `bool`: waypoint count > 0 |
| `Path.NumWaypoints` | none | `int`: remaining waypoint count |
| `Path.ListWaypoints` | none | `List<Vector3>`: copy of current positions |
| `Path.GetMovementAllowed` / `Path.SetMovementAllowed` | none / `bool` | `bool` function / action controlling the shared movement flag |
| `Path.GetAlignCamera` / `Path.SetAlignCamera` | none / `bool` | `bool` function / action; setter persists configuration |
| `Path.GetTolerance` / `Path.SetTolerance` | none / `float` | `float` function / action; waypoint proximity tolerance |
| `SimpleMove.PathfindAndMoveTo` | `Vector3 dest, bool fly` | `bool`: accepted internal query |
| `SimpleMove.PathfindAndMoveCloseTo` | `Vector3 dest, bool fly, float range` | `bool`: accepted query and destination tolerance |
| `SimpleMove.PathfindInProgress` | none | `bool`: internal pending task exists, including a completed task not yet processed |

SimpleMove returns false when it already has a pending task. It can throw if no
mesh lifetime exists. If accepted, it queries from the current local player's
position (or zero if no player), then a later `Update` feeds the result to
`FollowPath.Move`. It does not stop an existing path while computing its replacement.
Check that a player exists before calling. An empty result clears the path;
faults are logged. Neither case has a per-request IPC error/result channel.

`Nav.PathfindInProgress` covers manager queries, including raw IPC and SimpleMove.
`SimpleMove.PathfindInProgress` covers only the internal SimpleMove task.
`Path.IsRunning` covers queued movement, even when movement is disabled. Keep
these three observations separate in status displays.

`SetMovementAllowed(false)` does not cancel the query, clear waypoints or stop
all side effects. Follower updates can still process waypoint state, camera
alignment, takeoff logic and stuck detection. Do not use it as an emergency stop
or a consumer ownership lock. `Stop` clears the path immediately; movement and
camera hooks are disabled on a later follower update with a player present.

## Mesh projection

| Channel suffix | Parameters | Return |
| --- | --- | --- |
| `Query.Mesh.NearestPoint` | `Vector3 p, float halfExtentXZ, float halfExtentY` | `Vector3?` |
| `Query.Mesh.NearestPointReachable` | `Vector3 p, float halfExtentXZ, float halfExtentY` | `Vector3?`, excluding polygons flagged unreachable |
| `Query.Mesh.IsPointOnMesh` | `Vector3 p, float halfExtentY, bool allowUnreachable` | `bool`; false also when no query exists |
| `Query.Mesh.PointOnFloor` | `Vector3 p, bool allowUnlandable, float halfExtentXZ` | `Vector3?` |
| `Query.Mesh.FlagToPoint` | none | `Vector3?` |

Nullable results mean no usable projection or no query object; readiness must
be checked separately. `NearestPointReachable` uses flood-fill flags, not a new
route from the player. `PointOnFloor` passes its `allowUnlandable` parameter to
the implementation's `allowUnreachable` argument. It chooses the highest
candidate at or below `p.Y`, with a vertical search half extent of 2048. It does
not prove the character can land there. These calls are synchronous.

## Bitmap and UI channels

`Nav.BuildBitmap(startingPos, filename, pixelSize)` returns
`(Vector3 min, Vector3 max)` and writes a bitmap to the supplied path.
`Nav.BuildBitmapBounded` adds `Vector3 minBounds, Vector3 maxBounds`.
`Nav.BuildBitmapMulti` takes `List<Vector3>` instead of a single starting
position; `Nav.BuildBitmapMultiBounded` combines both variations. Pixel size is
documented in the manager as a power of two. These are synchronous file-writing
functions and require a loaded mesh and query. They enumerate polygons connected
to the starting points. A bitmap result is not a path or a movement request.

`Window.IsOpen` / `Window.SetOpen(bool)` and `DTR.IsShown` / `DTR.SetShown(bool)`
are bool functions and setter actions. The DTR setter persists configuration.
These controls do not affect navigation readiness.

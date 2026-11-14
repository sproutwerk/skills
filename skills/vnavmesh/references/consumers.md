# Consumer lifecycle and bounded example

The lifecycle design applies across versions. The channel names and behavioral
details here were checked at the [recorded pin](provenance.md); inspect the
target provider and its implementations when the installed build differs.

## Query without moving

Call this from an existing plugin's framework-thread callback after copying
the player's world position and validating the destination/territory. It needs
only Dalamud and .NET, not a reference to vnavmesh or an ECommons initializer.
It returns positions and never calls a movement endpoint.

```csharp
using System;
using System.Collections.Generic;
using System.Numerics;
using System.Threading;
using System.Threading.Tasks;
using Dalamud.Plugin;

static async Task<List<Vector3>> QueryRouteAsync(
    IDalamudPluginInterface pi,
    Vector3 from, Vector3 destination, bool fly,
    CancellationToken lifetime)
{
    static bool Finite(Vector3 v) =>
        float.IsFinite(v.X) && float.IsFinite(v.Y) && float.IsFinite(v.Z);

    if (!Finite(from) || !Finite(destination))
        throw new ArgumentException("World positions must be finite.");

    lifetime.ThrowIfCancellationRequested();
    var ready = pi.GetIpcSubscriber<bool>("vnavmesh.Nav.IsReady");
    var find = pi.GetIpcSubscriber<
        Vector3, Vector3, bool, CancellationToken, Task<List<Vector3>>>(
        "vnavmesh.Nav.PathfindCancelable");

    if (!ready.HasFunction || !find.HasFunction)
        throw new InvalidOperationException("Required vnavmesh IPC unavailable.");
    if (!ready.InvokeFunc())
        throw new InvalidOperationException("Navigation mesh is not ready.");

    using var cancel = CancellationTokenSource.CreateLinkedTokenSource(lifetime);
    cancel.CancelAfter(TimeSpan.FromSeconds(10));
    var task = find.InvokeFunc(from, destination, fly, cancel.Token);

    // Observe a late provider fault if the caller's bounded wait ends first.
    _ = task.ContinueWith(t => { _ = t.Exception; },
        CancellationToken.None,
        TaskContinuationOptions.OnlyOnFaulted |
        TaskContinuationOptions.ExecuteSynchronously,
        TaskScheduler.Default);
    try
    {
        var path = await task.WaitAsync(cancel.Token).ConfigureAwait(false);
        if (path.Count == 0)
            throw new InvalidOperationException("No route returned.");
        return new List<Vector3>(path);
    }
    finally
    {
        cancel.Cancel();
    }
}
```

This is an authored, source-derived example, not a compiled or game-tested
consumer. Ten seconds is an example query budget, not an upstream guarantee.
`WaitAsync` bounds the caller's wait; provider cancellation is cooperative and
may settle later. A cancellation inside the manager can surface as a fault
wrapping cancellation because it reads an inner task's `Result`. Inspect/log
the exception and stop the operation; do not depend only on `IsCanceled`.

Keep one query per consumer. Observe its task, including exceptions, without
`.Wait()` or `.Result` on an incomplete task. Cancel the lifetime token when the
script resets, the character/territory changes, the plugin disables, or the
consumer disposes. Detach the host's update handlers during teardown. A late
task completion must not schedule movement.

Before applying a returned route, return to the framework thread and compare
the captured operation generation, character, territory and relevant scene
state with fresh values. Also recheck the source position if the player moved
while the query ran. Cancellation can race completion, so a token alone cannot
replace this stale-result check. No vnavmesh callback is required for this
query-only example; there are no provider event subscriptions to dispose.

## Choose the follower

`Nav.Pathfind*` calls `QueryPathBasic`, which strips `Waypoint.Type` and returns
positions. `Path.MoveTo` constructs default-type waypoints from those positions.
This pair cannot preserve the special `ClientPath`/`ClientPathEnd` handling in
the internal follower. Use it for routes whose plain positions suffice, or use
a consumer-owned follower that implements the required transitions. Do not
assume arbitrary special routes survive the conversion.

SimpleMove calls `QueryPath` directly, retaining typed waypoints. Its internal
follower watches game conditions for client-path start/end transitions. The
`Warp` branch still contains a teleport TODO, so typed waypoints do not establish
general teleport support either.

For a Splatoon script, bind raw subscribers through the plugin interface
available in that script's host. Start work in an explicitly requested operation,
poll from its update lifecycle, and invalidate it on reset/disable/disposal.
Use the host's existing ECommons lifetime if a wrapper requires it; do not
initialize another global ECommons owner just to call these channels. Confirm
the script lifecycle names against the installed Splatoon version.

## Bounded movement state machine

This is a source-derived design example, not an executable script. Budgets are
consumer choices. It assumes the user requested movement to a verified world
position and an external coordinator grants exclusive use for the whole run.
Idle flags alone cannot grant exclusivity.

| State | Framework-thread work | Exit |
| --- | --- | --- |
| Waiting | Capture destination, character, territory and generation; wait up to 30 seconds for player and mesh readiness | Unavailable/timeout ends the operation |
| Querying | Start one cancelable query with a 10-second budget; retain the task and token | Fault, cancellation or empty list ends the operation |
| Validating | Recheck context and exclusive use; validate that the route needs no special waypoint types; require no active path or pending SimpleMove | Any stale/conflicting state ends the operation without starting movement |
| Following | Call `Path.MoveTo` once, record that this consumer started the path, then observe position and path state each tick | Deadline of 60 seconds, lost context, cancellation, or idle away from destination is failure |
| Arrived | Require matching context, player within the chosen destination tolerances, no pending SimpleMove and no remaining path | Record success and release external ownership |

When already at the destination, no movement call is needed. For ground arrival,
check horizontal XZ distance and a separate allowed vertical difference; ignoring
height can accept the wrong floor. For flight use three-dimensional distance.
Choose tolerances for the requested interaction, not just `Path.GetTolerance`.
The follower's waypoint test uses distance to the segment traversed between
frames, so it can consume a point after passing it.

If the path becomes idle away from the destination, report stopped/failed.
Distinguish cancellation from success even when the final position happens to
be close. A successful path query is not proof that movement will finish;
continue monitoring through the movement deadline.
Every failure or cancellation after starting movement enters the owned cleanup
procedure below before releasing the coordinator. If cleanup cannot be confirmed,
report that state instead of marking the run complete.

For SimpleMove, skip the explicit query/follow pair and issue one request after
the same ownership and context checks. Track acceptance, wait for
`SimpleMove.PathfindInProgress` to clear, then inspect `Path.IsRunning` and the
player's position. Do not immediately accept `Path.IsRunning == false` while
the internal query is pending. A fast route may finish between polls, so
requiring a sampled true running flag is also insufficient.

## Cancellation and ownership

`Path.Stop` and `/vnav stop` only clear the current path. They do not cancel the
pending SimpleMove task. That task can later feed a new path to the follower.
There is no per-request SimpleMove cancellation channel in this pin.

For raw queries, cancel your own token and invalidate the generation. The query
has no movement side effect, so a late result can be discarded. Call `Path.Stop`
only if the external coordinator still assigns the active movement to this
consumer. Do not stop another plugin's route on every script update or unload.

For an owned SimpleMove run that must be canceled, coordinate the global
`Nav.PathfindCancelAll` reload, clear the active path, and continue observing
pending/running state for a bounded cleanup period. Reload cancels all manager
queries and eventually clears the follower through the mesh-change event. It
also invalidates other consumers' work. A task may already have completed, so
observe the result and clear a late path while ownership remains exclusive;
do not declare cleanup finished just because the reload call returned. If
exclusivity is lost or the provider disappears, report unresolved cleanup and
stop dependent actions instead of blindly retrying global controls.

The shared `vnav.PathIsRunning` `bool[]` is a coordination hint used by some
other plugins. It has no owner identity or compare-and-set semantics and can
differ transiently from `Path.IsRunning`, which reads the waypoint list.
Never write it to acquire ownership. Raw IPC exposes no lease, caller ID,
request handle, priority arbitration or per-route completion result.

User-input cancellation and stuck recovery are configuration dependent. The
default `CancelMoveOnUserInput` and `StopOnStuck` fields are false;
`RetryOnStuck` is true. When enabled, stuck recovery can start SimpleMove even
after a raw `Path.MoveTo` because `AsyncMoveRequest` subscribes to the follower's
stuck event. A raw-query consumer must therefore account for pending SimpleMove
during cleanup too. If a design needs guaranteed per-operation cancellation,
use a follower it owns or establish a compatible configuration with the user.
Do not silently change shared settings to make an example's assumptions hold.

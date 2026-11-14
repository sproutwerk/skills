# Consumer lifecycle

Use the provider contracts in [ipc.md](ipc.md) and the execution behavior in
[QuestController](source/Questionable/Questionable/Controller/QuestController.cs).
The following C# fragment illustrates binding in a host that supplies an
`IDalamudPluginInterface pluginInterface`. It has not been compiled or run.

```csharp
var running = pluginInterface.GetIpcSubscriber<bool>("Questionable.IsRunning");
var current = pluginInterface.GetIpcSubscriber<string?>("Questionable.GetCurrentQuestId");
var complete = pluginInterface.GetIpcSubscriber<string, bool>("Questionable.IsQuestComplete");
var start = pluginInterface.GetIpcSubscriber<string, bool>("Questionable.StartSingleQuest");
var stop = pluginInterface.GetIpcSubscriber<string, bool>("Questionable.Stop");
```

Bind without invoking anything during construction. Invoke on the framework
thread when the requested workflow needs it. The provider calls native quest
state and controller methods directly and does not marshal requests for you.
Catch the host's IPC exceptions at the integration boundary; preserve an
unavailable/error state instead of reporting a failed `IsRunning` call as idle.

## Drive one ordinary quest

1. Record the requested quest ID and character context. Check provider
   availability, player readiness, and whether another automation session is
   already running. Query completion first, then acceptance and eligibility as
   appropriate. An accepted quest need not be ready to accept again.
2. Call `StartSingleQuest` once. Record a local operation generation and
   deadline. A false result means parsing or registry lookup failed. A true
   result only confirms that the provider dispatched the start.
3. Observe on later framework updates. Query completion for the requested ID;
   also inspect the current ID and running state for interruption or a change
   of work. `IsRunning == false` alone does not prove completion. A non-null
   current ID can remain useful after tasks stop and is not a running flag.
4. End the local operation when the quest completes, the context changes,
   the provider disappears, the user cancels, or the deadline expires.
   On a timeout, record the last quest/sequence/step and stop retrying. Avoid
   repeatedly issuing start calls while waiting for progress.

For repeatables and pseudo-elements, define a result appropriate to that work
before implementing the loop. Read `QuestFunctions` for the exact query
semantics. The provider has no general completion event carrying a request ID.

## Stopping and concurrent callers

`Stop(label)` calls `StopAllDueToConditionFailed`, affecting questing, movement,
combat, and gathering. The label becomes a diagnostic string, not an ownership
token. The controller's stop behavior can also invoke the user's configured
post-stop command when automatic mode stops.

Keep a record of whether this consumer started the operation, but do not treat
that local record as a provider lock. Another caller or the user can change
the global state. Coordinate ownership in the surrounding application; only
stop shared work when that coordination and the requested task authorize it.
Disposal should always remove the consumer's callbacks and polling. It should
not unconditionally stop unrelated Questionable activity.

`StartQuest` selects automatic mode. Choose it only when continuation is wanted;
the caller's list of IDs is not automatically a boundary for Questionable's
own scheduler. A local queue can dispatch one quest at a time with
`StartSingleQuest`, checking each result before advancing. Keep the user's
priority list intact unless modifying it is part of the task.

For Splatoon scripts, connect polling and cancellation to the script's own
enable, reset, and disable lifecycle. The same IPC contracts apply; ECommons
wrappers are optional. Do not copy internal controller methods into a script
or rely on private fields found through reflection when a supported channel
already provides the needed operation.

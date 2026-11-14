# Action requests and positional hints

## Timed action requests

This is a separate IPC family from configuration leases. The prefix is
`WrathCombo.ActionRequest.`. See the
[provider](source/WrathCombo/WrathCombo/Services/ActionRequestIPC/ActionRequestIPCProvider.cs)
and [raw signatures](source/WrathCombo.API/IPC/RawMethods/ActionRequest.cs).
`ActionType` is `FFXIVClientStructs.FFXIV.Client.Game.ActionType`, not an action ID
or a Wrath preset enum.

| Gate suffix | Arguments | Result |
| --- | --- | --- |
| `RequestActionUse` | `ActionType actionType, uint actionID, int timeMs, bool? isGcd` | void |
| `ResetRequest` | `ActionType actionType, uint actionID` | void |
| `ResetAllRequests` | none | void |
| `RequestBlacklist` | `ActionType actionType, uint actionID, int timeMs` | void |
| `ResetBlacklist` | `ActionType actionType, uint actionID` | void |
| `ResetAllBlacklist` | none | void |
| `GetArtificialCooldown` | `ActionType actionType, uint actionID` | float seconds |

The wrapper spells the all-blacklist helper `ResetAllBlacklists()` (plural),
but the gate suffix is singular. These action gates have no lease argument or
per-client request ID. Reset by action removes every matching request, including
another client's; reset-all affects all clients. Use a short, positive lifetime
and let it expire when shared ownership cannot be established. Releasing a
configuration lease does not remove action requests. Do not use global resets
as ordinary consumer cleanup.

Requests append an entry with a deadline based on `Environment.TickCount64`.
The provider examines newest requests first, checks readiness, and currently
accepts only `ActionType.Action` for execution. `isGcd=true` skips weave windows;
`false` requires one; null skips that classification check. Returning an action
from the selector does not remove the request, so it is not an exactly-once
queue. It expires or is reset. A successful void IPC call only means the request
was recorded. Requests enter the combo invocation path in
[CustomCombo.cs](source/WrathCombo/WrathCombo/CustomCombo/CustomCombo.cs); do not
promise immediate execution or force auto-rotation on as an implicit side effect.

Blacklists contribute an artificial cooldown: the getter uses the longest
remaining matching deadline and returns seconds. Jobs/territory cache updates
clear action requests and blacklists in `WrathCombo.UpdateCaches`. Review those
call sites when designing lifetime behavior after a patch.

Read-only utility gates in [Utility.cs](source/WrathCombo/WrathCombo/Services/IPC/Utility.cs)
include `WrathCombo.CanWeave(float?)`, `CanDelayedWeave(float?,float?)`,
`ActionReady(uint,bool?,bool?)`, and `JustUsed(uint,float?)`, all returning bool.
The pinned API wrapper covers the first three, but not `JustUsed`. These checks
are transient observations, not reservations or guarantees that a subsequent
action will succeed.

## Read-only positional hints

`WrathCombo.GetUpcomingPositionalHint()` returns `uint[]?`. The event
`OnUpcomingPositionalHint` has no `WrathCombo.` prefix and carries no payload;
pull a new snapshot when it fires. Neither needs a lease. Use
`WrathIPCWrapper.GetUpcomingPositionalHint()`,
`SubscribeUpcomingPositionalHint(Action)`, and the matching unsubscribe helper.
Retain the same delegate for teardown and stop displaying a hint when it expires,
the provider disappears, or the returned snapshot is null.

The [snapshot decoder](source/WrathCombo.API/PositionalHintSnapshot.cs) accepts
at least seven fields:

| Index | Field | Meaning |
| --- | --- | --- |
| 0 | Direction | `None=0, Rear=1, Flank=2, Unknown=3` |
| 1 | ActionId | Expected action |
| 2 | GcdsUntil | Number of GCDs until it is expected |
| 3 | TargetObjectId | Target entity ID |
| 4 | ExpiresInMs | Remaining validity in milliseconds |
| 5 | CurrentAngle | Front=0, Flank=1, Rear=2, Unknown=3 |
| 6 | IsSatisfied | Nonzero means satisfied |

`PositionalHintSnapshot.IsActive` requires direction other than None, nonzero
action, positive GCD count, and positive remaining lifetime. A hint is a forecast;
it is not permission to move the player or evidence of navigation completion.
The provider and event implementation are in
[ProvidePositionalHint.cs](source/WrathCombo/WrathCombo/Services/IPC/ProvidePositionalHint.cs)
and [UpcomingPositionalHintService.cs](source/WrathCombo/WrathCombo/Services/IPC/UpcomingPositionalHintService.cs).
Recheck the wire layout and expiry logic before using a different revision.

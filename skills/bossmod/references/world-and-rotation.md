# World state, hints, and automation

## World-state ownership

[WorldState](source/BossMod/Data/WorldState.cs) is updated through operations.
`Execute` applies an operation, stamps its timestamp from the world clock,
and fires `Modified`. Actors, party membership, client state, waymarks, and
zone/CFC IDs live in this model. The game-sync layer populates the live world;
replays apply recorded operations to another world.

Use `Actor.InstanceID` for a particular spawn and `Actor.OID` for its data
identity. `PartyState` slots are a separate domain. A matching OID does not
identify one actor if several copies exist. Reacquire actors by instance ID
when needed, handle absent party slots and destroyed actors, and discard
references when the world is replaced. For encounter timestamps use
`WorldState.CurrentTime` or `FutureTime`, so seeking and replaying do not depend
on the wall clock.

The live sequence in [TickService](source/BossMod/Services/TickService.cs) is
registered on `IUiBuilder.Draw`: update camera/game sync, update boss and zone
modules, rebuild hints, gather manual actions, update rotation, finish action
gathering, record replay state, draw UI, and execute hints. This is the actual
pin's scheduling path; do not assume a separate public framework-update hook
exposes mutable hints to other plugins.

## Loaded, drawn, and active modules

[BossModuleManager](source/BossMod/BossModule/BossModuleManager.cs) owns module
instances and disposes them on unload/reset. `LoadedModules` includes modules
that have not pulled. `DrawnModule` is the selected module, with an optional
manual override. `ActiveModule` returns that drawn module only when its state
machine has an active phase. An OID registry hit proves none of these states.
The manager updates all loaded modules, catches update exceptions, and unloads
failed or completed ones.

The pinned [AIHintsBuilder](source/BossMod/BossModule/AIHintsBuilder.cs) clears
hints each update, collects enemies, then uses the drawn module when its phase
is active. With no active encounter it calculates generic hints and adds zone
module hints. With an active encounter that has zero components it also
calculates generic AOE hints. This fallback is inferred from available game
state; it does not establish that the encounter is fully implemented.

## Hint output

[AIHints](source/BossMod/BossModule/AIHints.cs) holds shared per-update output:
forbidden zones and directions, goal functions, target priorities, predicted
damage, movement requests, and queued actions. Contribute through the
appropriate component or rotation hook after clearing; do not keep a reference
to a previous frame's entries as persistent mechanic state.

Forbidden zones carry a shape/distance function and activation time. Goal
zones score desirable positions among safe candidates; a high goal value does
not remove a forbidden zone. Supply `AddForbiddenZone` with the matching shape
or signed distance contract. `GenericAOEs` already handles inversion and
risk filtering, so use it when the mechanic fits.

Both function kinds are SIMD-vectorized at this pin.
[Vector.cs](source/BossMod/Util/Vector.cs) aliases `DistFn` and `GoalFn` to
`Func<WVec, Vector<float>>` and `PointTest` to `Func<WVec, Vector<int>>`, where
a `WVec` carries a batch of X/Z sample points. `AddForbiddenZone` takes a
`DistFn`, an [Sdf](source/BossMod/Util/Sdf.cs), or an `AOEShape` with origin
and rotation. `GoalZones` is a `GoalList` of `GoalFn`; `AddSimple(test, value)`
wraps a `PointTest` that should score a constant. The `Goal*` helpers on
`AIHints` return `GoalFn`. The `GoalZonesEnabled` flag from earlier revisions
is gone. Code written against `Func<WPos, float>` needs porting, not a cast.

`PotentialTargets` contains enemies with priorities. Negative priorities carry
different meanings: pointless (-1), invincible (-2), undesirable (-3), and
forbidden (-4). `PriorityForbidden` can prevent actions when custom queuing is
enabled; do not flatten every negative value to an equivalent blacklist.
`PredictedDamage` contains party-slot masks, activation times, and damage
metadata. It is a prediction from modules, not a record of confirmed damage.

## Rotation selection and stop controls

[RotationModuleManager](source/BossMod/Autorotation/RotationModuleManager.cs)
uses active preset modules when the active list is nonempty, otherwise a
selected encounter plan if one is eligible. It rebuilds modules for player
class/level, roleplay compatibility, and PvP compatibility, then sorts by
`RotationModuleOrder`: high-level strategy, actions, movement.

The special `ForceDisable` preset has an empty name and no modules. It suppresses
plan execution while selected. `SetForceDisabled` clears other presets before
adding it. Activating a normal preset removes the sentinel; clearing the list
can resume a plan. Read `GetForceDisabled` separately from `GetActive`, because
an empty string and null have different meanings here.

AI configuration can add the default `VBM Multibox` preset while rebuilding
modules. It removes that preset first, then re-adds it if `AIConfig.Enabled`
is true and the force-disable sentinel is absent. The active list can therefore
change after a successful IPC write. Registry changes, death, combat-end,
countdown, and pull-safety settings can also clear or force-disable presets.
Read the relevant config and events when diagnosing an unexpected change.

| User intent | Pinned control | Scope |
| --- | --- | --- |
| Suppress automatic rotation, including plans | `BossMod.Presets.SetForceDisabled` or `/vbm ar disable` | Rotation preset/plan selection; not a universal cancellation of manual input or other plugins |
| Remove manual preset overrides and return to plan selection | `BossMod.Presets.ClearActive` or `/vbm ar clear` | An eligible plan can execute afterward |
| Disable AI mode | `/vbm ai off` or a requested `Configuration` write to `AIConfig Enabled false` | Changes AI config; other user presets can still contain movement/action modules |
| Limit AI movement/targeting options | `AIConfig.ForbidMovement` / `ForbidActions` | Modify transient tracks on the default VBM Multibox preset; not global movement/action blocks |

[AIWindow](source/BossMod/AI/AIWindow.cs) responds to config changes by
rewriting track 0 on the first/last modules of the default `VBM Multibox`
preset. `ForbidActions` controls auto-targeting there; its name does not mean
all action execution is blocked. Avoid writing competing transient strategies
to that shared compatibility preset. `SetFollowSlot` updates its follow track
and enables AI when the slot is nonnegative.

The command handlers are in
[SlashCommandService](source/BossMod/Services/SlashCommandService.cs).
Bare `/vbm ai` toggles its window. `/vbm ai follow` accepts a party member name
or `slot<N>` with a one-based input, then publishes `SetFollowSlot`; it is a
control action, not a discovery command. The alias `/vbmai` is marked deprecated
in source. The only `BossMod.AI.*` registrations in this VBM provider are the
read-only `AI.IsEnabled` and `AI.IsAutoTargetEnabled`; see
[IPC contracts](ipc.md#read-only-state-queries). No IPC channel sets the follow
slot or pauses AI movement.

## Writing a rotation module

[RotationModule](source/BossMod/Autorotation/RotationModule.cs) takes a
`RotationModuleManager` and player `Actor`. The registry requires a public
static `Definition()` returning `RotationModuleDefinition`. Its class/level
range, compatibility, and order control whether and when the manager runs it.
Implement `Execute(StrategyValues, ref Actor? primaryTarget, float
estimatedAnimLockDelay, bool isMoving)`, or use `TypedRotationModule<TValues>`
with the matching typed strategy definition.

Track indices, `InternalName`, and option names must agree with the definition.
They are used by presets and IPC. Before renaming them, inspect saved preset
serialization and converter migration support. Use the normal action queue
and hint contracts; a rotation module should not bypass the action manager by
calling native action functions merely to make an example work.

[StayCloseToTarget](source/BossMod.Autorotation/MiscAI/StayCloseToTarget.cs)
is a compact source example: it defines a range track and adds goal functions
around the target hitbox, using `GoalZones.AddSimple` for the on-hitbox donut
and `AIHints.GoalSingleTarget` otherwise. It demonstrates a positional
preference. A movement
module still has to consume hints and execute movement. Test strategy selection
and missing-target behavior independently of live execution.

## Coexistence and passive use

For a requested temporary handoff, first read `GetActiveList`,
`GetForceDisabled`, and the relevant `AIConfig` fields. Select the explicit
set/off control for the behavior being handed off; a toggle depends on prior
state. `/vbm ai on` and `off` assign `Enabled`, while `toggle` inverts it.
`/vbm ar disable` force-disables rotation; bare `/vbm ar toggle` toggles that
sentinel and can clear overrides. These controls do not unload the plugin or
its encounter modules. Bare `/vbm` opens settings and `/vbm ar` toggles the
rotation window, which is separate from execution.

The config service persists AI config; active presets and transient strategies
are manager/runtime state subject to the reset rules above. After a handoff,
read the state again on a later update because the manager can rebuild its
compatibility preset. Restore only state your integration changed, and only
if no user or other consumer has changed it since. The readbacks cannot make
this atomic; without an ownership token, competing writers need coordination.

As an integration design rule, assign one owner to automatic action selection
and one to movement for the requested mode. Multiple VBM presets can run
modules of the same type (`DuplicateModules` exposes those duplicates), and
another plugin may independently issue movement or actions. VBM's IPC offers
no cross-plugin arbitration or proof that those other systems are idle.
Do not infer compatibility with another executor from their being installed
together. Inspect each target's controls and lifecycle before arranging a
handoff; do not bypass its access restrictions.

Passive use remains useful when direct control is unavailable: inspect
encounter source and replays, or query VBM registry support, configuration, and
preset state. The read-only `Rotation.*` and `AI.*` queries also report which
rotation modules the active presets hold and a few action-tweak and AI
settings. This provider does not expose the live `WorldState`, complete
`AIHints`, or a generic timeline query. Explain that limit instead of claiming
an active encounter from `HasModuleByDataId` or obtaining private state through
an unreviewed reflection fallback.

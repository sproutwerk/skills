# Encounter modules and components

These contracts describe the recorded VBM revision. Start with
[BossModule](source/BossMod/BossModule/BossModule.cs),
[BossComponent](source/BossMod/BossModule/BossComponent.cs), and
[BossModuleRegistry](source/BossMod/BossModule/BossModuleRegistry.cs) when
porting an older module. Examples from older VBM versions or forks can use
different constructors and callback signatures.

## Registration and construction

At this pin a module derives from `BossModule` and accepts `ModuleInit`, which
contains `World`, `Primary`, and registry `Info`. Supply the arena center and
bounds to the base constructor. The registry discovers nonabstract module
classes from loaded assemblies, associates a state builder by
`<ModuleFullName>States` unless `StatesType` overrides it, and locates `OID`,
`AID`, `SID`, `TetherID`, and `IconID` in the module namespace unless overridden.

The builder must derive from `StateMachineBuilder` and have a constructor
compatible with the module type. Existing simple builders accept `BossModule`;
more specialized builders accept their concrete module. A primary OID must be
provided in `ModuleInfo` or by an `OID.Boss` enum entry. Keep OID enums backed
by `uint`. The registry indexes by that primary OID; duplicate registrations
are logged. Module metadata includes grouping, name IDs, and contributor
credits. Preserve those credits when adapting an encounter.

A minimal source example is
[D011Chopper.cs](source/BossMod.Modules/RealmReborn/Dungeon/D01Sastasha/D011Chopper.cs),
credited to Kagekazu under the bundled BSD license. These two declarations
show the constructor and component pattern; use the full linked file for its
OID/AID enums, state builder, and metadata:

```csharp
class ChargedWhisker(BossModule module)
    : Components.StandardAOEs(module, AID.ChargedWhisker, 6.15f);

public class D011Chopper(ModuleInit init)
    : BossModule(init, init.Primary.Position, new ArenaBoundsCircle(20));
```

The full file activates `ChargedWhisker` from `TrivialPhase()`. This is a
source-derived excerpt, not a newly compiled starter. Do not install another
copy of its OID in the same registry. For a new encounter, derive the OID,
action IDs, cast timing, center, and bounds from the requested encounter's
source or replay; retain uncertainty where evidence is missing.

## Component lifetime and events

Only one component of a type may be active. `ActivateComponent<T>()` replaces
an existing instance with an error log and constructs a fresh instance from
the module. It then seeds existing state: non-player actor creation and spell
casts, targetability, tethers, and statuses. Account for this when writing
counters or storing casts; activation during a cast is different from replaying
its historical cast event.

Use component callbacks for cast starts/finishes, cast events, status changes,
icons, actor state, and map effects. Distinguish telegraph timing from damage
resolution: `OnCastFinished` and `OnEventCast` are separate inputs. Cast
callbacks exclude player spells. `OnStatusGain` also receives status changes
without an intervening lose event. Compare the actual action or status ID,
not only the caster type.

`Update()` runs for active components while the module's state machine has an
active state. Rendering and hint collection should read the accumulated
mechanic state; they may be called separately from updates. Use
`AddHints`, `AddGlobalHints`, `AddMovementHints`, and `AddAIHints` for their
respective outputs. Drawing an arena shape alone does not supply an avoidance
constraint to AI.

After the components and `CalculateModuleAIHints` run, `BossModule` clears
`ForcedMovement` (when nonzero), `ForcedTarget`, and `ForcedFocusTarget` and
logs a warning naming the module. Encounter code at this pin cannot use those
low-level fields; express the intent through forbidden zones, goal zones, and
target priorities instead. Modules written for older revisions that set them
will compile and then silently lose the behavior.

`DeactivateComponent<T>()` removes entries; `BossComponent` has no general
`IDisposable` teardown contract. Prefer its routed callbacks over subscribing
directly to long-lived services. If a module owns additional subscriptions,
dispose them in its `Dispose(bool)` override and call the base implementation.
Base disposal resets the state machine, clears components, and disposes world
subscriptions. [Event.Subscribe](source/BossMod/Util/Event.cs) returns a
disposable subscription; retain it with its owner. Do not assume removing a component cleans up its own external
subscriptions or cancels work it started.

## State-machine choices

[StateMachineBuilder](source/BossMod/BossModule/StateMachineBuilder.cs) supports
`TrivialPhase` for mechanics handled entirely by persistent components,
`DeathPhase` for a fight ending on the primary actor's death/destruction or
zero HP, and `SimplePhase` for explicit phase conditions. The default pull
requires the primary actor to be targetable and in combat, and
`AllowedToActivate` also checks the player's active FATE ID against the actor's
FATE ID. Override these only for encounter-specific evidence.

Use observed cast or component conditions for mechanics, with expected timing
for display and overrun handling. `Timeout` is elapsed-time driven.
`ComponentCondition<T>` checks the active component and logs missing-component
or overdue conditions. Review its fallback before treating a timeline as an
exact event prediction. `TrivialPhase` uses a long single state named Enrage;
it is not a full mechanic timeline.

`ActivateOnEnter<T>` and `DeactivateOnExit<T>` scope components to states.
Phase exits automatically remove components unless `KeepOnPhaseChange` is
set. A retained component must have an explicit lifetime across the remaining
phases. Stable state IDs matter because cooldown plans refer to them; keep
the high-byte phase/subsequence and high-word group conventions documented in
the builder when changing timelines.

[Ex4IfritStates](source/BossMod.Modules/RealmReborn/Extreme/Ex4Ifrit/Ex4IfritStates.cs)
shows cast starts/ends, component conditions, and downtime/raidwide hints. Its
comments also preserve uncertainty about HP transitions and variable timings.
Use it to understand sequencing; do not copy its delays to another fight.

## Reuse AOE components

[GenericAOEs.cs](source/BossMod/Components/GenericAOEs.cs) links one
`ActiveAOEs(slot, actor)` stream to player text, arena drawing, and AI forbidden
zones. Its `AOEInstance` holds shape, origin, rotation, activation time, color,
risk, and inversion. `Risky = false` suppresses warning/AI avoidance while
allowing drawing; an inverted instance reverses the check and distance sign.

Distance evaluation is vectorized at this pin. `AOEInstance.Distance(WVec)`
returns a `VFloat` (`System.Numerics.Vector<float>`) covering several sample
points per call, `DistV` exposes it as a `DistFn` for
`AIHints.AddForbiddenZone`, and the scalar `Check(WPos)` / `Distance(WPos)`
wrappers remain for single points. The aliases are global usings in
[Vector.cs](source/BossMod/Util/Vector.cs). A custom zone written as
`Func<WPos, float>` for an older revision no longer matches `AddForbiddenZone`;
port it to a `DistFn`, or pass an `AOEShape` with its origin and rotation.

`StandardAOEs` tracks matching cast starts and removes casters at cast finish,
using `CastInfo.LocXZ`, cast rotation, and `Module.CastFinishAt` for instances.
The radius overload constructs an `AOEShapeCircle`. Confirm whether an action's
reported range includes the actor hitbox before choosing a radius. AOE shapes
and `WPos` use horizontal X/Z coordinates; convert 3D positions deliberately.

For delayed resolution after a cast ends, moving/rotating telegraphs, or
multiple damage events, implement the required state in a `GenericAOEs`
subclass instead of assuming `StandardAOEs` models it. Use
`WorldState.CurrentTime`/`FutureTime` and remove expired or resolved entries on
the appropriate event. Exercise interrupted casts and actor destruction in a
replay before relying on the result.

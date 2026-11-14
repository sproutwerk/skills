# Conditions, triggers, freeze, phases — how a preset is evaluated

Source: `Splatoon/Utility/LayoutUtils.cs`, `Splatoon/Splatoon.cs`
(`ProcessLayout`, `ProcessElementsOfLayout`, `ResetLayout`, `CombatEnded`),
`RenderEngines/CommonRenderUtils.cs`, `RenderEngines/DirectX11/DirectX11Renderer.cs`.

## Per-frame order

```
for layout in Config.LayoutsL:            ProcessLayout(layout)
for script in enabled scripts:            ProcessLayout(each Controller.Layouts)   # same rules
for script in enabled scripts:            ProcessElement(each Controller.Elements) # no layout rules
injected elements (projection, /sf, editor previews), then Web-API dynamic elements
```

`ProcessLayout(l)`:
1. `l.ConditionalStatus = null`
2. `IsLayoutVisible(l)` (below) — if false only the freeze snapshots (if any) are re-drawn.
3. `Freezing`? → if `FreezeInfo.CanDisplay()` (now > `AllowRefreezeAt`): render
   elements into a snapshot; if anything was drawn, store it with
   `ShowAt = now + FreezeDisplayDelay`, `ShowUntil = now + FreezeFor`,
   `AllowRefreezeAt = now + max(0.1s, IntervalBetweenFreezes)`.
   When `Freezing` is false, call `ProcessElementsOfLayout(l)` live instead.
4. Active snapshots are injected; expired ones removed. Snapshots are cleared
   on zone change (`FreezeResetTerr`) and on combat end (same flag —
   `FreezeResetCombat` is not read).

### IsLayoutEnabled
false if: `!Enabled`; group in `DisabledGroups`; `DisableInDuty && BoundByDuty`;
zone list mismatch (whitelist unless `IsZoneBlacklist`); `Scenes` non-empty and
active scene not in it; `Phase != 0 && Phase != P.Phase`; `JobLockH` non-empty
and job not in it; `DCond` 1/3 outside combat, 2/3 outside duty, 4 neither,
6/8 in combat, 7/8 in duty, 9 not (combat && duty). `DCond 5` is **not**
checked here.

### IsLayoutVisible
`IsLayoutEnabled` && distance-to-target rule (`DistanceLimitType 0`: needs a
target; `min <= dist - hitboxes < max`) && trigger rule:

```
if UseTriggers:
  for each trigger not UserDisabled and not FiredState==2:
    type 2/3 and not Disabled: for each message in CurrentChatMessages this frame:
        match (regex or ContainsIgnoreCase) ->
           Duration==0 ? FiredState=0 : FiredState=1, DisableAt += now+Duration+MatchDelay
           MatchDelay!=0 ? EnableAt += now+MatchDelay : TriggerCondition = (type==2 ? 1 : -1)
           FireOnce -> Disabled = true
    type 0/1 and FiredState==0 and in combat and now - CombatStarted > TimeBegin*1000:
           Duration==0 ? FiredState=2 : FiredState=1, DisableAt += now+Duration
           TriggerCondition = (type==0 ? 1 : -1)
    due EnableAt -> TriggerCondition = (type==2 ? 1 : -1)
    due DisableAt -> FiredState = (type 2/3 ? 0 : 2); TriggerCondition = 0
  hidden if TriggerCondition == -1 or (TriggerCondition == 0 and DCond == 5)
```

`TriggerCondition` is one value per layout, so the last trigger to fire wins.
`Duration 0` on a time trigger makes the state permanent (until reset) —
that is how "show from 0:45 onwards" works. `ResetLayout` (zone change and
combat end) resets triggers whose `ResetOnTChange` is true: `FiredState=0`,
`TriggerCondition=0`, `Disabled=false`, timers cleared.

`CombatStarted` is set when `InCombat` becomes true and again when the
"limit gauge resets" line arrives (phase change), so `TimeBegin` restarts per
phase.

### Trigger fields

`Trigger.Type` 0 shows at `TimeBegin` seconds after combat starts; 1 hides at
that time. Types 2 and 3 show or hide when a message matches `Match`.
`Duration:0` has no expiry. `MatchDelay` postpones the effect, and `FireOnce`
disables the trigger until reset. `IsRegex` selects regular-expression matching;
otherwise matching uses a case-insensitive substring.

### Messages available to type 2/3 triggers

Everything enqueued into `P.ChatMessageQueue` and drained into
`CurrentChatMessages` each frame:

| Source | Exact format |
|---|---|
| Every non-standard chat channel (battle log lines etc.), text payloads concatenated, control chars stripped | the log line text |
| Cast start (`AttachedInfo`) | `{Name} starts casting {actionId} ({nameId}>{actionId})` — with `({X, Y, Z})` after the name when *Log position* is on |
| Map effect | `MapEffect: {index}, {param1}, {param2}` |
| Tether create | `Tether create: TetherSource:{YOU|JOB|baseId} TetherParam:({p1}, {p2}, {p3}) TetherTarget:{YOU|JOB|dataId|oid}` |
| VFX spawn (characters) | `VFX {path} spawned on {me|JOB|dataId} npc id={nameId}, model id={modelId}, name npc id={nameId}, position={pos}, name={name}` |
| VFX spawn (other objects) | `VFX {path} spawned on {dataId} npc id={nameId}, position={pos}` |
| Status on you | `You gains the effect of {statusId} Param: {param} ([buff+]You:{statusId}:{JOB})` |
| Status on others | `{Name} ({pos}) gains the effect of {statusId} Param: {param} ([buff+]{nameId}:{statusId})` or `([buff+]PC:{statusId}:{classJobId})` |

Prefixes `buff+` add, `buff-` remove, `buff*` update; the formats are the same
with `Config.ExperimentalStatusHooks` on. Not matchable:
`ObjectEffect: …`, `Tether removal…`, `Director Update…` (log window only).
Match on ids, not names, for `[International]` presets: `"starts casting 36284"`,
`"MapEffect: 5, 4, 2"`, `"[buff+]You:3818:"`.

## Elements inside a layout

`ProcessElementsOfLayout` iterates `GetElementsWithSubconfiguration()` (the
selected subconfiguration's `Elements`, else `ElementsL`) in order:

```
if element.Conditional && element.ConditionalReset: status = null
skip if status == false && (ConditionalAnd || !element.Conditional)
result = renderer.ProcessElement(element, layout)      # true if anything drew
if element.Enabled && element.Conditional:
    r = ConditionalInvert ? !result : result
    status = status == null ? r : (ConditionalAnd ? status && r : status || r)
```

So a `Conditional` element (usually `Nodraw:true`, `radius:0`, `thicc:0`)
whose object/cast/status filter matches turns the following normal elements
on. With `ConditionalAnd:false` (OR, default) a later conditional element
that matches re-enables drawing; with AND all preceding conditionals must
hold. `ConditionalReset` starts a fresh group.

`ProcessElement` returns true only when at least one object matched, so a
relative element whose NPC is absent gates to false.

### Object matching (types 1/3/4, `refActorType 0`)

For each object in the object table:
`IsAttributeMatches` (comparison type or AND of all set attributes) &&
`IsElementObjectMatches`: targetability, `IsDead`, `LimitRotation`,
`UseHitboxRadius`, `refTargetYou`, `refActorObjectLife`, `LimitDistance`,
`ObjectKinds`, then `CheckCharacterAttributes`: `onlyVisible` (character and
visible), cast (`IsCastingMatches` xor `refActorCastReverse`), status
(`CheckEffect`), transformation, head mark, tether (`IsTetherMatches` xor
`refActorIsTetherInvert`), `AnimationIds`. Layout distance limit type 1 is
checked per object. Then `HandleEnumeration` orders/filters, and
`AlterTargetIfNeeded` redirects to tether partners / target / N-th closest
player. Radius gets `+HitboxRadius` if `includeHitbox`.

`refActorType 1` (self) and `2` (target) skip attribute matching but still run
`CheckCharacterAttributes` with visibility ignored.

### Cast matching detail
`refActorCastId` non-empty: `chr.IsCasting(ids)`; with `refActorUseCastTime`
also `CurrentCastTime in [Min,Max]`. When not casting and
`refActorUseOvercast`, Splatoon's own `CastInfos[addr]` (start time recorded at
cast start, kept until the object stops existing/next cast) gives an elapsed
time that must be in `[Min,Max]` — use `CastTimeMax` larger than the cast
length to keep drawing after the snapshot (e.g. `Max 9.4` for a 5 s cast that
resolves 4 s later). Empty `refActorCastId` = any cast.

### Tether matching detail
Historical mode: `AttachedInfo.TetherInfos[obj]` entries created by the tether
hook (`Param1=p1, Param2=p2, Param3=p3, Target`) with `AgeF in [TimeMin,
TimeMax]` and each non-null param equal; `refActorIsTetherSource` null/true
checks tethers owned by the object, null/false does a reverse lookup over all
other objects' tethers targeting it. `refActorTetherConnectedWithPlayer`
placeholders must resolve to the other end. Live mode reads
`Character.Vfx.Tethers` (`Id` = param 2 only). Entries are removed on the
tether-removal hook, so `TimeMax` bounds how long after creation you keep
drawing; `TimeMin/Max = 0/0` never matches.

## Phases

`P.Phase` = 1 on territory change; `++` when chat type 2105 equals LogMessage
row 2844 ("The limit gauge resets!" in the client language); DSR forces phase
1/2 from director/DataId checks. Scripts get `OnPhaseChange` and
`Controller.Phase`. A layout with `Phase: 2` therefore means "after the first
limit-gauge reset" (door-boss → main boss).

## Freeze timing

A capture at `t0` appears at `t0 + FreezeDisplayDelay` and expires at
`t0 + FreezeFor`. With delay 3 and lifetime 10, it remains visible for about
7 seconds; delay greater than or equal to lifetime means it never appears.
The next capture is allowed at `t0 + max(0.1, IntervalBetweenFreezes)`.
During that cooldown, `Freezing:true` draws only still-active snapshots;
there is no live-rendering fallback. A moving actor stays at its captured
position. Use `Freezing:false` for live per-frame rendering.

## Freeze usage pattern

Combine `Freezing` with a cast condition limited to the first fraction of the
cast (`refActorUseCastTime`, `CastTimeMax 0.4`) to capture where helpers were
when the cast began, then `FreezeFor` the mechanic length and
`IntervalBetweenFreezes` ≥ the time until the next occurrence (upstream EX1
example: `FreezeFor 10`, `Interval 60`, `DCond 5` + MapEffect trigger).

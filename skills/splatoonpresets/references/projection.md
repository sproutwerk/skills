# Projection and per-layout projection overrides

Source: `Splatoon/Services/Projection.cs`, `Gui/Tabs/TabProjection.cs`,
`Gui/Layouts/Header/LayoutDrawHeader.cs`, `Serializables/Configuration.cs`.

## What projection is

A service (`S.Projection`, framework tick) that, for every object in the
object table that `IsBattleNpc`, is not a friendly BattleNPC, is casting an
`ActionType.Action`, and whose `Action` row is known, guesses the AoE shape
from sheet data and injects a temporary `Element` (not part of any layout):

| `Action.CastType` | Shape | Size |
|---|---|---|
| 2 | circle | `EffectRange` (ground-targeted: drawn as type 0 at `TargetLocation`) |
| 5 | circle | `EffectRange + caster hitbox` |
| 3 | cone | length `EffectRange + hitbox`, half-angle from omen path `fan###` (180° if unknown) |
| 13 | cone | length `EffectRange` |
| 4 | rect (type 3 line, `radius = XAxisModifier/2`, `offY = range`) | `EffectRange + hitbox` |
| 12 | rect | `EffectRange` |
| 11 | cross | two rects (second at +90°) |
| 10 | donut | omen `sircle_OOII` → inner `II`, thickness `OO-II` |
| 6,7,8 | none | custom shapes / player ground circles / charges skipped |

Circles with `EffectRange >= 30` are treated as raidwides and skipped.
Rotation comes from the cast packet when Splatoon saw it (`LastCast`), else
from the actor's facing. Opposing cones/lines whose cast ends >0.2 s later are
dropped to reduce clutter. Style: `Config.ProjectionColor1/2` gradient with
`ProjectionPulseTime`, `ProjectionFillIntensity`, `ProjectionCastAnimation`
(default Fill) with `AnimationColor1/2`.

## User-side switches

`Config.EnableProjection` (off by default), `ProjectionBlacklistedActions`
(`{Action, DataId, NameId, ModelId, Territory}` records; `/splatoon pbl`
adds everything currently projected), `ProjectionBlacklistedDuties` +
`IsProjectionBlacklistWhitelist` (zone list mode). The Projection tab shows
each projectable cast with which layouts blacklist/whitelist/suppress it.

## Layout-level overrides (what a preset can ship)

Evaluated for every `IsLayoutEnabled` layout (zone, DCond, phase, job all
apply; the layout may have zero elements):

- `ForcedProjectorActions` (UI "Projection Whitelist"): action ids always
  projected, regardless of `EnableProjection`, the global blacklist, or
  suppression by elements.
- `BlacklistedProjectorActions` (UI "Projection Blacklist"): never projected.
- Automatic suppression: any **enabled** element of type 1/3/4 in the layout
  with `refActorRequireCast` and the action id in `refActorCastId`, whose
  `IsAttributeMatches` passes for the caster, marks the cast as "already
  processed" — the hand-made element replaces the projection.

Resolution: `hide = whitelisted==false || (suppressed && !whitelisted) ||
(globallyBlacklisted && !whitelisted) || (!whitelisted && !EnableProjection)`.
Blacklist beats whitelist inside one layout; across layouts the one later in
the list overrides (`showOverride` is overwritten in list order).

Declutter preset shape (upstream M11S):

```
~Lv2~{"Name":"M11S - Projection declutter","Group":"M11S (2)","ZoneLockH":[1325],"ElementsL":[],"BlacklistedProjectorActions":[46116,46099,46101,46100,46126,46145,46165,46167,46166],"ForcedProjectorActions":[46136]}
```

`Layout.ProjectionState` (bool?) exists in the class but nothing reads it.
Per-element "projection overrides" do not exist beyond the automatic
suppression above; per-mechanic-type style overrides live in the user's
`Config.StyleOverrides` (Render → global style overrides), which is why
`mechanicType` should be set on shipped elements.

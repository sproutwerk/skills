# Element fields (Splatoon 3.9.2.32, `Splatoon/Serializables/Element.cs`)

Serialized with `DefaultValueHandling.Ignore`: a field equal to its default is
omitted. "Gate" = a `ShouldSerializeX()` rule that also suppresses the field
on export (import never requires it). Colours are `uint` ABGR (`0xAABBGGRR`)
in decimal. Angles marked *rad* are radians. Coordinates: `refX`/`offX` world
X, `refY`/`offY` world Z (horizontal), `refZ`/`offZ` world Y (height).

## Identity / geometry

| Field | Type | Default | Meaning |
|---|---|---|---|
| `Name` | string | `""` | display name; used as key by `RegisterElementFromCode(code)` and by `layout~element` toggles |
| `InternationalName` | `{En,Jp,De,Fr,Other}` | empty | localized name; gate: non-empty |
| `type` | int | 0 (ctor arg) | 0 fixed circle, 1 object circle, 2 fixed line, 3 object line, 4 object cone, 5 fixed cone |
| `Enabled` | bool | true | element switch |
| `refX`,`refY`,`refZ` | float | 0 | reference position / line point A / (type 3) point A in object-local space; gate: `type != 1` |
| `offX`,`offY`,`offZ` | float | 0 | offset (types 0,1,5 added to ref; types 2,3 = point B) |
| `radius` | float | 0.35 | circle radius; line half-width (0 = 1-px line); cone length |
| `Donut` | float | 0 | > 0: ring of thickness `Donut` outside `radius`; gate: type 0/1/4/5 and > 0 |
| `coneAngleMin`,`coneAngleMax` | int deg | 0 | cone span around facing; must be min < max; gate: type 4/5 |
| `includeHitbox` | bool | false | add target hitbox radius to `radius` |
| `includeOwnHitbox` | bool | false | add player hitbox radius |
| `includeRotation` | bool | false | rotate offset/line/cone with the object's facing (forced true for cones in UI) |
| `AdditionalRotation` | float *rad* | 0 | extra angle added when `includeRotation` |
| `RotationOverride` | bool | false | replace object rotation |
| `RotationOverrideAngleOnlyMode` | bool | false | fixed angle: `(180 + RotationOverrideAddAngle)°`; gate: RotationOverride |
| `RotationOverrideFaceMode` | bool | false | face a placeholder instead of a point |
| `RotationOverrideFaceModePlaceholders` | string[] | [] | gate: FaceMode && non-empty |
| `RotationOverridePoint` | `{X,Y}` | 0,0 | point to face (X, world Z); gate: RotationOverride |
| `RotationOverrideAddAngle` | float deg | 0 | added angle |
| `FaceMe` | bool | false | (types 3,4,5) orient towards `faceplayer` |
| `faceplayer` | string | `"<1>"` | placeholder to face (`<me>`, `<target>`, `<tethered>`, `<1>`..`<8>`, `<t1>`, `<h1>`, `<d1>`.., `<objectid:N>`, `<element:Layout:Element>`) |
| `FaceInvert` | bool | false | placeholder faces this element instead |
| `UseCastRotation` | bool | false | use rotation from the cast packet (`Projection.LastCast`) |
| `UseCastPosition`,`UseCastTarget` | bool | false | present in class, editor UI commented out |
| `LineAddHitboxLength{X,Y,Z}` / `...{X,Y,Z}A` | bool | false | (type 3) add target hitbox to point B / A per axis |
| `LineAddPlayerHitboxLength{X,Y,Z}` / `...A` | bool | false | same with player hitbox |
| `LineEndA`,`LineEndB` | enum | 0 None | 1 Arrow; only for zero-radius lines or tethers |
| `tether` | bool | false | draw line from player to the shape (not for line types) |
| `ExtraTetherLength` | float | 0 | extend tether past the shape (knockback preview) |
| `EnablePointerLine` | bool | false | (type 2, radius 0) animated arrow line |
| `PointerLineStyle` | struct | defaults | see `serializables/PointerLineStyle.cs`; nulls fall back to user defaults; gate: EnablePointerLine |
| `UsePlaceholderAsRefPosition`,`PlaceholdersRefPosition[]` | bool,string[] | false,[] | (types 0,2,5) take ref position(s) from placeholders |
| `UsePlaceholderAsOffPosition`,`PlaceholdersOffPosition[]` | bool,string[] | false,[] | (type 2) take point B from placeholders |
| `PairingMode` | enum | 0 One_to_one | (type 2 with placeholders) 1 Every_to_every |

## Anchor object (types 1, 3, 4)

| Field | Type | Default | Meaning |
|---|---|---|---|
| `refActorType` | int | 0 | 0 object-table search, 1 self, 2 current target (BattleNpc only) |
| `refActorComparisonType` | int | 0 | 0 Name, 1 ModelID, 2 ObjectID, 3 DataID, 4 NPC ID, 5 Placeholder, 6 NPC Name ID, 7 VFX path, 8 Object effect, 9 Nameplate icon |
| `refActorComparisonAnd` | bool | false | AND every non-empty attribute instead of the single selected one |
| `refActorName` / `refActorNameIntl` | string | `""` | partial case-insensitive; `"*"` matches any object; gate: type 0 or And |
| `refActorModelID` | uint | 0 | `ModelContainer.ModelCharaId` (characters only) |
| `refActorObjectID` | uint | 0 | `EntityId`; per-instance, scripts set it at runtime |
| `refActorDataID` | uint | 0 | `DataId` (BaseId); stable per NPC/EventObj, best for invisible helpers |
| `refActorNPCID` | uint | 0 | `GameObject.GetNameId()` (works for non-character objects) |
| `refActorPlaceholder[]` | string[] | [] | e.g. `["<2>","<3>",…]`, `<me>`, `<t>`, `<mo>`, `<t1>`, `<h2>`, `<d4>` |
| `refActorNPCNameID` | uint | 0 | `ICharacter.NameId` (BNpcName row). Language-independent; preferred |
| `refActorVFXPath`,`refActorVFXMin`,`refActorVFXMax` | string,int ms | `""`,0,0 | VFX attached to object, age window in ms (UI shows seconds) |
| `refActorObjectEffectData1/2`,`refActorObjectEffectMin/Max`,`refActorObjectEffectLastOnly` | uint,int ms,bool | 0 | object-effect (EventObj animation) data pair, age window, or only the most recent |
| `refActorNamePlateIconID` | uint | 0 | nameplate icon (decimal) |
| `TargetAlteration` | enum | 0 None | 1 Tethered (draw on tether partners), 2 Targeted (on its target), 1100–1103 Closest_Player_1..4, 2100–2103 Furthest_Player_1..4 |
| `Enumeration` | enum | 0 None | 1 Clockwise, 2 Counter_Clockwise ordering of all matches |
| `EnumerationCenter`,`EnumerationStart` | `{X,Y}` | 0,0 | angle origin; gate: Enumeration != None |
| `EnumerationOrder[]` | int[] | [] | 1-based indices to keep; negative = from end |
| `ObjectKinds[]` | int[] (Dalamud `ObjectKind`) | [] | restrict kinds (1 Player, 2 BattleNpc, 3 EventNpc, 7 EventObj …) |
| `onlyTargetable` / `onlyUnTargetable` | bool | false | targetability filter |
| `onlyVisible` | bool | false | `IsCharacterVisible()`; also excludes non-characters |
| `IsDead` | bool? | null | null any, true only dead, false only alive |
| `UseHitboxRadius`,`HitboxRadiusMin/Max` | bool,float | false,0 | hitbox radius window; gate: Use && != 0 |
| `LimitRotation`,`RotationMin`,`RotationMax` | bool,float *rad* | false,0 | object rotation window (UI shows `180 - deg`) |
| `refActorObjectLife`,`refActorLifetimeMin/Max` | bool,float s | false,0 | seconds since the object spawned (ECommons ObjectLifeTracker) |
| `refActorUseTransformation`,`refActorTransformationID` | bool,int | false,0 | character transformation id |
| `AnimationIds[]`,`AnimationInverted` | uint[],bool | [],false | EventObj `AnimationId` in list (or NOT) |
| `MapEffects[{Effect,State}]`,`MapEffectInvert`,`MapEffectAnd` | list,bool,bool | [] | serialized; consumed by MapEffect tooling |

## Cast condition

| Field | Default | Meaning |
|---|---|---|
| `refActorRequireCast` | false | require the object to be casting |
| `refActorCastReverse` | false | invert: while NOT casting |
| `refActorCastId[]` | [] | action row ids; empty = any cast; gate: RequireCast && non-empty |
| `refActorUseCastTime`,`refActorCastTimeMin/Max` | false,0,0 | seconds into the cast, `min <= t <= max` |
| `refActorUseOvercast` | false | keep matching after the cast bar is gone, using Splatoon's cast clock (`AttachedInfo.CastInfos`), until `CastTimeMax` |
| `castAnimation` | 0 | 1 Pulse, 2 ColorShift, 3 Fill (DX11 only; gate: RequireCast) |
| `animationColor`,`pulseSize`,`pulseFrequency` | 0xC80000FF, 0.5, 1.5 | animation style |

Cast progress for animations = `CurrentCastTime / TotalCastTime`, or the
`CastTimeMin..Max` window when `refActorUseCastTime`.

## Status condition

| Field | Default | Meaning |
|---|---|---|
| `refActorRequireBuff` | false | require statuses |
| `refActorBuffId[]` | [] | status row ids; gate: RequireBuff && non-empty |
| `refActorRequireAllBuffs` | false | ALL listed present (else ANY) |
| `refActorRequireBuffsInvert` | false | invert result |
| `refActorUseBuffTime`,`refActorBuffTimeMin/Max` | false | remaining time window (s) |
| `refActorUseBuffParam`,`refActorBuffParam` | false,0 | require `Status.Param` (stack count / variant) |

## Tether condition

| Field | Default | Meaning |
|---|---|---|
| `refActorTether` | false | enable |
| `refActorIsTetherLive` | false | check live `Vfx.Tethers` (only param 2 = tether id, no times) instead of Splatoon's tether history |
| `refActorTetherTimeMin/Max` | 0,0 | age window (s) of the historical tether — **set Max > 0** or nothing matches |
| `refActorTetherParam1/2/3` | null | match `Tether create: TetherParam:(p1, p2, p3)`; null = any |
| `refActorIsTetherSource` | null | true = object is source, false = object is target, null = either |
| `refActorIsTetherInvert` | false | invert |
| `refActorTetherConnectedWithPlayer[]` | [] | placeholders the other end must resolve to; empty = any |

## Other conditions

| Field | Default | Meaning |
|---|---|---|
| `refMark`,`refMarkID` | false,0 | head marker: 0-4 attack1-5, 5-7 bind1-3, 8-9 stop1-2, 10 square, 11 circle, 12 cross, 13 triangle, 14-16 attack6-8 |
| `refTargetYou`,`refActorTargetingYou` | false,0 | 1 = object NOT targeting you, 2 = targeting you |
| `LimitDistance`,`LimitDistanceInvert` | false | distance of the object to a source |
| `DistanceSourceX/Y/Z` | 0 | fixed source (X, world Z, world Y) |
| `UseDistanceSourcePlaceholder`,`DistanceSourcePlaceholder[]` | false,[] | placeholder sources (OR) |
| `DistanceMin`,`DistanceMax` | 0 | `min <= d <= max` |

## Style / text / misc

| Field | Default | Meaning |
|---|---|---|
| `color` | 0xC80000FF (3355443455) | stroke ABGR |
| `thicc` | 2 | stroke thickness; 0 hides stroke |
| `Filled` | true | fill (legacy `false` exports auto-migrate to true for non-circles) |
| `fillIntensity` | null | 0..1 multiplier of stroke alpha for fill; null = computed from `color` alpha/`FillStep` |
| `overrideFillColor`,`originFillColor`,`endFillColor` | false,null,null | explicit gradient fill; gate: Filled && override |
| `FillStep`,`LegacyFill` | 0.5,false | ImGui-legacy renderer fill density |
| `RenderEngineKind` | 0 Unspecified | 1 DirectX11, 2 ImGui_Legacy (Unspecified = user default) |
| `mechanicType` | 0 | 1 Danger 2 Safe 3 Soak 4 Gaze 5 Knockback 6 Information; user style overrides apply per type (not for 6) |
| `overlayText` / `overlayTextIntl` | `""` | text at shape centre (types 0,1,4,5) |
| `overlayVOffset`,`overlayFScale` | 0,1 | vertical offset (world units), font scale |
| `overlayBGColor`,`overlayTextColor` | 0x70000000, 0xC8FFFFFF | text box colours |
| `overlayPlaceholders` | false | (type 1) expand `$NAME $OBJECTID $DATAID $MODELID $NAMEID $NPCID $HITBOXR $KIND $LIFE $DISTANCE $TETHER $TRANSFORM $ANIMATIONID $CAST $CASTNAME $CAST:fmt$ $STREM:statusId:fmt$ \n` |
| `Conditional`,`ConditionalInvert`,`ConditionalReset` | false | gate elements below (see conditions doc) |
| `IsCapturing` | false | store matched positions for `<element:…>` placeholders |
| `Nodraw` | false | evaluate but never render (invisible condition element) |

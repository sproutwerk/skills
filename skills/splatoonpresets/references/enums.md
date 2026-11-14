# Integer enumerations used in preset JSON

All enums serialize as integers (no `StringEnumConverter`).

## Element.type
| # | UI label | Notes |
|---|---|---|
| 0 | Circle at fixed coordinates | `ref*` + `off*`; `Donut`; overlay text |
| 1 | Circle relative to object position | anchor by `refActorType`; `off*` rotated if `includeRotation` |
| 2 | Line between two fixed coordinates | A `ref*`, B `off*`; `radius` half-width; pointer line |
| 3 | Line relative to object position | A/B in object space; +Y forward when rotated; hitbox-length flags |
| 4 | Cone relative to object position | `radius` length, `coneAngleMin/Max` deg, rotation forced |
| 5 | Cone at fixed coordinates | `FaceMe`/`AdditionalRotation` for direction |

Projection maps game `CastType` to: circle/donut→1, rect/cross→3, cone→4 (or
0 for ground-targeted circles).

## Element.refActorType
0 Game object with specific data (search) · 1 Self · 2 Targeted enemy

## Element.refActorComparisonType
0 Name (partial, `*`=any) · 1 Model ID · 2 Object ID (EntityId) · 3 Data ID ·
4 NPC ID (`GetNameId()`) · 5 Placeholder · 6 NPC Name ID (`NameId`) ·
7 VFX Path · 8 Object Effect · 9 Icon ID (nameplate)

## Element.refActorTargetingYou (with refTargetYou)
1 not targeting you · 2 targeting you

## Element.refMarkID
0 attack1 1 attack2 2 attack3 3 attack4 4 attack5 5 bind1 6 bind2 7 bind3
8 stop1 9 stop2 10 square 11 circle 12 cross 13 triangle 14 attack6 15 attack7
16 attack8

## MechanicType (`mechanicType`)
0 Unspecified · 1 Danger · 2 Safe · 3 Soak · 4 Gaze · 5 Knockback · 6 Information
Default DX11 colours (stroke, fill): Unspecified 0xC8006FFF, Danger 0xC80000FF,
Safe 0xC800D114, Soak 0xC8FF9000, Gaze 0xC8A000D6, Knockback 0xC800F7FF (fill =
same RGB with alpha 0x45). Information cannot be style-overridden by users.

## CastAnimationKind (`castAnimation`)
0 Unspecified · 1 Pulse (`pulseSize`, `pulseFrequency`) · 2 ColorShift · 3 Fill

## LineEnd (`LineEndA`, `LineEndB`)
0 None · 1 Arrow

## TargetAlteration
0 None · 1 Tethered · 2 Targeted · 1100..1103 Closest_Player_1..4 ·
2100..2103 Furthest_Player_1..4

## EnumerationType (`Enumeration`)
0 None · 1 Clockwise · 2 Counter_Clockwise

## PairingMode
0 One_to_one · 1 Every_to_every

## RenderEngineKind (`RenderEngineKind`)
0 Unspecified · 1 DirectX11 · 2 ImGui_Legacy (`RenderEngines/RenderEngineKind.cs`)

## Dalamud ObjectKind (`ObjectKinds[]`)
0 None 1 Player 2 BattleNpc 3 EventNpc 4 Treasure 5 Aetheryte 6 GatheringPoint
7 EventObj 8 MountType 9 Companion 10 Retainer 11 Area 12 Housing 13 Cutscene
14 CardStand 15 Ornament (from Dalamud, not in this checkout)

## Layout.DCond
0 Always shown · 1 Only in combat · 2 Only in instance · 3 combat AND instance ·
4 combat OR instance · 5 On trigger only · 6 Outside of combat · 7 Outside of
instance · 8 Outside of combat AND instance · 9 Outside of combat OR instance

## Layout.DistanceLimitType
0 Distance to current target · 1 Distance to element

## Trigger.Type
0 Show at time in combat · 1 Hide at time in combat · 2 Show at log message ·
3 Hide at log message

## Job (`JobLockH`, ClassJob row ids)
GLA 1 PGL 2 MRD 3 LNC 4 ARC 5 CNJ 6 THM 7 · PLD 19 MNK 20 WAR 21 DRG 22 BRD 23
WHM 24 BLM 25 ACN 26 SMN 27 SCH 28 ROG 29 NIN 30 MCH 31 DRK 32 AST 33 SAM 34
RDM 35 BLU 36 GNB 37 DNC 38 RPR 39 SGE 40 VPR 41 PCT 42

## Placeholders (strings)
`<me>` `<target>`/`<t>` `<mo>` `<tethered>` `<1>`..`<8>` (party slot) `<t1>`
`<t2>` `<h1>` `<h2>` `<d1>`..`<d4>` (role slots by party-list order)
`<objectid:N>` / `<objectid:0xN>` (position by EntityId)
`<element:ElementName>` / `<element:LayoutName:ElementName>` (positions captured
by an `IsCapturing` element). In duty-recorder playback `<me>`/`<target>`
follow the base-player override.

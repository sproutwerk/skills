# Worked examples

Each `.json` is the pretty-printed Layout; the matching `.import.txt` is the
one-line `~Lv2~` string ready to paste (regenerate with
`printf '~Lv2~%s\n' "$(jq -c . x.json)"`).
Fields left at their defaults are included for readability; Splatoon's own
export would omit them.

Every example carries `[EXAMPLE/UNTESTED]` in its `Name` and a `Description`
saying which values are real ids and which are placeholders/illustrative, so
the label survives in the pasted `~Lv2~` line. None of them was verified in
combat; treat them as format demonstrations. Remove the tag from `Name` once
you have replaced the placeholder values from a recording (and remember that
importing a `Name` that already exists fails unless CTRL is held).

## 01 — circle on a boss cast (`01-boss-cast-circle.*`)

Real ids from the upstream EX1 preset (Worqor Lar Dor, zone 1196, action
39261 "Ice Boulder", radius 13). Any visible object (`refActorName:"*"`,
`onlyVisible`) casting 39261 gets a red circle that fills over the cast
(`castAnimation:3`), tagged `mechanicType:1` (Danger). Layout is combat-only.
Because the element has `refActorRequireCast` + that cast id, projection
suppresses its own guess for 39261 automatically.

To anchor on a specific NPC instead: replace `refActorName` with
`"refActorNPCNameID":<NameId>,"refActorComparisonType":6`.

## 02 — donut with a tether condition (`02-donut-tether.*`)

Zone 1232 / NameId 13057 (Wicked Thunder, from the upstream R4S script) are
real; **`refActorTetherParam2: 84` is a placeholder** — read the real tether id
from the Splatoon log line `Tether create: … TetherParam:(p1, p2, p3) …` and
use `p2`. Element 1: a ring (`radius 5` hole, `Donut 10` thickness) around the
boss while a tether it *owns* (`refActorIsTetherSource:true`) with that param
has existed for 0–30 s and its other end resolves to `<me>`; `tether:true`
also draws a line from you to it. Element 2: overlay text under your own
character (placeholder `<me>` via comparison type 5) when any other object's
tether targets you (`refActorIsTetherSource:false`, reverse lookup).
`refActorTetherTimeMax` must be > 0 or nothing matches.

## 03 — raid preset: two layouts, one paste (`03-raid-layout-phases-*.*`)

Paste both `.import.txt` lines together (one per line) — that is how upstream
ships multi-layout presets. Zone 1325 / NameId 14305 / actions 46114, 46116,
46099 are real M11S ids; the coordinates and the `MapEffect: 5, 4, 2` line are
**illustrative** — take them from a recording.

- `…-towers`: `DCond:5` ("on trigger only") + `UseTriggers`. A type-2 trigger
  shows the layout for 12 s when the MapEffect line appears; a type-3
  `FireOnce` trigger hides it when 46116 starts casting. `Freezing` snapshots
  the two fixed tower circles (`type:0`, `mechanicType:3` Soak) for 12 s and
  refuses a refreeze for 60 s. `BlacklistedProjectorActions` stops projection
  from drawing 46116/46099 on top. `Phase:0` = any phase; set `Phase:2` for
  post-limit-gauge-reset only.
- `…-safespots`: combat-only. Element 1 is an invisible gate (`Conditional`,
  `Nodraw`, `radius 0`, `thicc 0`) that is true while the boss casts 46114 and
  for up to 9.4 s from cast start thanks to `refActorUseOvercast`; with
  `ConditionalAnd:true` everything below is drawn only then: a fixed green
  safe spot and a danger donut around the boss. `Subconfigurations` holds a
  second element list "Strat B (mirrored, no donut)" with a mirrored safe
  spot and, deliberately, **no danger donut**: a subconfiguration is a full
  replacement of `ElementsL` (`Layout.GetElementsWithSubconfiguration()`
  returns the selected list instead of, not merged with, the default), so
  anything you want in every strat must be repeated in every list. Copy the
  donut into Strat B if you want it there too. The Configurations tab can
  switch every layout/script that has a configuration of the same name.
  `SelectedSubconfigurationID` all-zeros = default set.

# Worked cases and client verification

These are authored examples and future acceptance cases, not captured FFXIV
results. Their structure and arithmetic were rechecked by hand on 2026-11-13;
game-specific event behavior still needs observation.

Notation follows [rules.md](rules.md). A bracketed group is an already
declared meld and is excluded from the concealed string. Assume no unrelated
event yaku unless specified. Scoring examples have no honba, deposits, or pao.

## Hand analysis examples

| Case | Position | Expected result |
| --- | --- | --- |
| Standard completion | `123m123p123s456s77z` | Four sequences plus pair; structural shanten -1 |
| Two-sided wait | `123m123p123s45s77z` | Structural shanten 0; waits exactly `3s,6s` |
| Whole-wait furiten | Previous hand; own historical river contains `3s` | Ron on `6s` is also blocked; tsumo remains possible |
| Dora without yaku | Open `[123m]`; concealed `23m456p789s55p`; ron `4m`; dora `5p` | Complete shape, three dora (the pair plus the `5p` in `456p`), no yaku: no legal win |
| Seven Pairs | `1122m3344p5566s77z` | Complete Seven Pairs; fixed 25 fu |
| Quad is not two pairs | `1111m2233p4455s6z` | Seven Pairs shanten 2; five pair types and six distinct types |
| Thirteen-sided Kokushi | `19m19p19s1234567z` | Tenpai; all thirteen terminal/honor types complete it |
| Ambiguous decomposition | `11223344556677m` | Both Seven Pairs and standard decompositions exist; compare scores separately |
| Concealed kan accounting | `[1111m]` ankan plus `123p123s45s77z` | Ten concealed tiles, fourteen physical tiles after discard; still closed |
| Red-five normalization | `405p` | Counts `4p:1,5p:2`; one of those fives is red |

The no-yaku case also lacks pinfu because it is open, lacks tanyao because
of terminals, and has no complete straight or common sequence across suits.
Drawing the winning tile would not add menzen tsumo to that open hand.

## Scoring examples

Let the closed pre-win hand be `123m456m234p67s55p`, with seat South and
round East. The winning tile is `8s`; there are no red fives or matching dora,
no riichi, and no special event yaku.

- Ron gives Pinfu only: 1 han, 30 fu, base 240, payment 1,000.
- Tsumo gives Pinfu plus Menzen Tsumo: 2 han, 20 fu, base 320.
  The dealer pays 700 and each other player pays 400, for a total of 1,500.

At 3 han 40 fu, base points are 1,280. Non-dealer ron is 5,200, dealer ron
is 7,700, non-dealer tsumo is 1,300/2,600, and dealer tsumo is 2,600 each.
Here the slash means each non-dealer payer/dealer payer, not a total.
Round each payment independently.

At 5 han, use base 2,000 regardless of fu: non-dealer ron 8,000,
dealer ron 12,000, non-dealer tsumo 2,000/4,000, dealer tsumo 4,000 each.
These ordinary cases avoid the unresolved kiriage boundary.

## Event cases

Build replay cases for these before describing a future engine as validated:

| Sequence or condition | Expected invariant or question |
| --- | --- |
| Pass a structural completion with no yaku | Temporary furiten must still be considered under the official FAQ |
| Pass a win after accepted riichi | Ron stays blocked for the hand; tsumo remains eligible |
| Claim `1m` using `2m3m` | Immediate discards of `1m` and `4m` are restricted |
| Ankan after riichi | Unchanged wait count alone is insufficient; compare sets and hand type |
| Kokushi would complete from an opponent's ankan | Doman forbids ron in this event |
| Two players ron; then a separate three-ron case | Double win settlement versus abortive draw |
| Claimed river tile also appears in a meld | One physical copy, two historical references |
| Dora display shows `5m` | Under the default display setting the bonus applies to `5m`, including the red five; under the Traditional setting it applies to `6m`. Record which setting was active |
| Score becomes zero versus negative | Distinct continuation behavior |
| Kan request is robbed | Do not also commit its replacement draw as if the kan succeeded |
| Late or duplicate snapshot | Do not repeat a state transition or an action |

## Record observations

For each unresolved item in [scoring.md](scoring.md), create the smallest
position that distinguishes the alternatives. Observe permitted practice or
recorded play when that work is authorized. This skill does not itself request
live actions or claim that a test environment exists.

Record game build, language, mode, rules options, full hand/melds, winning
tile, event order, scores before/after, dora, and the displayed result. Preserve
whether a tile was drawn, discarded, claimed, or merely represented by an icon.
Record the dora display setting too. Mark missing history explicitly. Store
sanitized fixtures with the relevant source citation in the solver project's
own test data; keep private client data outside Git.

A useful observation record has:

```text
case_id, game_build, mode, rules_profile_revision
initial_state, ordered_events, final_state
observed_legal_actions, observed_result
prediction, differing_fields, candidate_explanation
evidence_path, confidence, unresolved_fields
```

When a prediction differs, first check tile mapping, phase, duplicated
observations, and missing history. Then isolate a rule or arithmetic mismatch.
Retain both the old expectation and new evidence. A corrected expectation
should identify the affected profile version and the regression case.

## Validation layers

Check parsing and tile conservation separately from agari/shanten. Then check
yaku/fu, stateful legality, transfers, and recommendations. Compare a second
Riichi implementation only on rules whose settings match. Record its pinned
revision and disagreements; majority agreement among generic libraries does
not override the FFXIV client.

Useful invariants include suit-renaming symmetry for shape, red-five shape
equivalence, conservation of physical tiles, and conservation of points plus
the riichi pot. Green-tile yaku and seat winds intentionally break some
symmetries, so do not apply a shape invariant indiscriminately to scoring.

Do not measure strategy quality by whether one suggested discard won its
hand. Use repeated seeded comparisons with a declared opponent model, then
evaluate recorded FFXIV positions once the observation adapter is reliable.

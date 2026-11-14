# Defense and push/fold

Offense-only ranking (shanten, then improving copies) loses matches by
dealing in. A player policy needs an explicit deal-in estimate per discard
and a comparison against the value of continuing. This reference records
the published statistics and the comparison most solvers use. They are
Tenhou top-room numbers, not FFXIV measurements; treat them as priors and
re-estimate from recorded matches when enough outcomes exist.

## Threat detection from what the client shows

- A sideways river tile is positive riichi evidence for that river's owner.
  Its absence is not evidence of safety: the marker can be missing from a
  partial observation, and an undeclared tenpai hand shows nothing.
- Rivers are per-opponent. Genbutsu (a tile that opponent discarded) is safe
  against that opponent only, by their discard furiten. Any tile another
  player discarded after that opponent's riichi without being ronned is also
  safe against the riichi hand, because passing it would leave them furiten
  for the rest of the hand.
- Undeclared tenpai is invisible without melds and discard order. Model it
  as a prior that grows with the opponent's river length, never as zero.

## Deal-in rates against riichi, by tile category

Houou-room data (Tenhou, about 1.2 million games) published by Path of
Houou, percent chance that the discard is the riichi hand's winning tile:

| Status | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Non-suji | 6.6 | 8.2 | 9.6 | 14.3 | 14.5 | 14.4 | 9.6 | 8.3 | 6.7 |
| Half suji (4–6 only) | | | | 8.3 | 8.7 | 8.3 | | | |
| Riichi suji | 1.5 | 5.2 | 7.7 | 3.8 | 3.5 | 3.9 | 7.7 | 5.4 | 1.5 |
| Full suji | 1.5 | 3.0 | 4.5 | 2.2 | 2.2 | 2.1 | 4.6 | 3.1 | 1.5 |

Suji: a 1–3 is suji when the rank three higher is in that river; a 7–9
when the rank three lower is; a 4–6 needs both the rank three lower and the
rank three higher for full suji, and one of them makes it half suji. The
chart's "Riichi suji" column is read here as suji of the riichi declaration
tile; the article does not define it beyond the label.

Honors by copies visible to the discarder besides the discard itself:

| Tile | 0 visible | 1 | 2 | 3 |
| --- | ---: | ---: | ---: | ---: |
| Guest wind | 2.6 | 1.5 | 0.1 | 0 |
| Round / seat wind | 3.0–4.0 | 1.2–1.5 | 0.15 | 0 |
| Dragon | 3.2 | 1.35 | 0.14 | 0 |
| Dora honor | 5.2–6.2 | 2.9–3.2 | 0.9–1.6 | 0 |

Each dead suji after about five raises a non-suji tile's danger by roughly
0.5 points (1 for a double non-suji). Dora honors are about twice as dangerous
as plain ones in the table above; treat suited dora and red fives as riskier
too, though this article gives no figures for them. A tile with three other
copies visible cannot complete a triplet, so an honor in that state loses only
to a single-tile pair wait or Thirteen Orphans; a suited tile can still
complete a sequence. Kabe (four visible copies of a neighbour) removes the
two-sided wait through it. Source and charts:
<https://pathofhouou.blogspot.com/2020/05/analysis-tile-deal-in-rates-by-live-suji.html>.

## Push/fold comparison

For each legal discard `d`:

```text
EV(d) = P(win | shape after d) × V(win) − λ × Σ_opponents P(tenpai_j) × P(deal-in_j | d) × V(loss_j)
```

- `P(win)` from shanten, improving copies and remaining draws (a stage
  chain or a lookup calibrated by simulation), reduced by the chance that
  someone else ends the hand.
- `V(win)` from the scorer over the actual waits at tenpai; before tenpai a
  generic estimate from held bonuses and likely yaku.
- `P(tenpai_j)` is 1 with riichi evidence, otherwise a prior.
- `V(loss_j)` an average riichi hand, or less for undeclared hands. A commonly
  quoted rough figure is 5,000–6,000 points; it is not from the article
  above, so calibrate it. `λ > 1` expresses placement risk aversion.

A fold is simply the discard with the best EV when every efficient tile is
too dangerous. Sequence folds: keep enough safe tiles to survive several
turns, and prefer tiles safe against every live threat.

## Riichi versus staying undeclared

Compare the declared and undeclared expected values of the best tenpai
discard. Declaring adds the riichi han, ura dora and ippatsu chances, and
the deposit lost when the hand does not win; it also locks the hand so every
later draw is discarded regardless of danger. Staying undeclared keeps the
option to fold but a yakuless hand then wins only by self-draw. Common
outcomes of the comparison: declare early good waits and yakuless hands;
pass thin waits late; pass when an opponent's riichi makes a low-value
locked hand a losing bet, then decide push or fold on the discard.

## What to measure

Rate the policy by whole-match outcomes with paired seeds and seat rotation:
mean placement, last-place rate, deal-in rate, win rate, final score, with
confidence intervals. Draw-completion speed alone does not measure strength.
Record ranked results separately; simulated opponents are not players.

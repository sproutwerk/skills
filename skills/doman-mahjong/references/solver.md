# Solver design

This is a proposed design for a future solver. It has not been implemented or
benchmarked by this skill. A useful first milestone is a deterministic hand
analyzer that runs on saved positions; game observation and action execution
can be added independently.

## State contract

Keep a versioned rules profile separate from match state. Give each rule an
evidence status: official documentation, observed on a named game build,
generic assumption, or unknown. An unknown value must not become `false`
through a boolean default. Where the value changes an answer, return alternatives
or withhold that part of the result.

| State | Required information |
| --- | --- |
| Identity | Schema version, game build, match/hand ID, event sequence |
| Rules | Match length, kuitan, confirmed exceptions, unresolved scoring options |
| Phase | Actor, draw/discard/call/replacement/settlement phase, pending responses |
| Hand | Concealed physical tiles, drawn tile separately, declared melds |
| Meld | Kind, physical tiles, claimed tile, source seat, event ID, open/closed |
| Rivers | Ordered discard events with red identity, claimed flag, riichi marker |
| Declaration | Pending/accepted riichi, deposit, ippatsu eligibility, pass history |
| Round | Dealer, seat winds, round wind, scheduled hand, honba, riichi pot |
| Public state | Scores, revealed tiles, live-wall count if known, observed call options |
| Bonuses | Actual dora types, raw indicator types only if verified, revealed ura |
| Reliability | Missing history, uncertain tile mappings, stale or partial observations |

Preserve the distinction between not observed and known absent. Ordered
events are needed for furiten, ippatsu, last-tile wins, and interrupted first
turns. A screenshot of the current hand cannot reconstruct all of them.

## Shape, shanten, and waits

Define complete shape as shanten `-1`, tenpai as `0`, and one away from tenpai
as `1`. This definition measures structural distance; yaku and ron legality
are separate results.

For standard hands, recursively enumerate non-overlapping complete groups,
one head pair, incomplete groups (taatsu), and unused tiles. With `m` total
complete groups including declared melds, `t` taatsu, and `p` equal to 0 or 1
for a head, a candidate is `8 - 2*m - min(t, 4-m) - p`. Minimize over valid
assignments. Pairs can serve as taatsu when not used as the head. Greedily
extracting sequences first loses valid alternatives. Reject physical fifth
copies and account for fixed melds before evaluating concealed counts.

For hands without declared melds, also calculate:

```text
chiitoitsu = 6 - pair_types + max(0, 7 - distinct_types)
kokushi   = 13 - distinct_terminal_or_honor_types - has_orphan_pair
shanten   = min(standard, chiitoitsu, kokushi)
```

Here `pair_types` counts types with at least two copies, once per type;
`has_orphan_pair` is 0 or 1. The distinct-type term prevents a quad counting
as two Seven Pairs pairs. Declared kans also disable these special families.
For shape completion, still require the exact family predicate in
[rules.md](rules.md), not just a plausible-looking formula result.

For each legal discard, form the resulting settled hand `H`. Try adding each
of the 34 tile types that would not create a fifth physical copy. A tile is
an improvement when it lowers shanten; at tenpai, use explicit complete-shape
enumeration to produce waits. Return separate sets for structural completion,
yaku-bearing ron, legal ron after furiten, and yaku-bearing tsumo. A wait
with no unobserved copies can remain part of the structural wait set used
for furiten while contributing zero available improvements.

Use exhaustive decomposition or independently validated tables before
optimizing. Shape calculations can share memoization across red/non-red
fives; value calculations must retain red identity.

## Available improvements

Count each physical tile once across the player's tiles and public tiles.
A tile claimed from a river is historical evidence for furiten, but its
physical copy is now in a meld. Do not subtract it twice.

```text
unseen[type] = 4 - distinct_known_physical_copies[type]
ukeire      = sum(unseen[type] for improving tile types)
```

Do not clamp a negative count silently; it identifies bad state or duplicate
observations. Track red and ordinary remaining fives inside their common
four-copy limit when estimating value.

By default FFXIV displays actual bonus types. A per-player display setting
("Traditional") shows the indicator tile one below the bonus tile instead, so
the adapter has to know which mode the observed client uses before mapping the
icon to a bonus type. Under the default, a dora icon is not evidence that a physical
tile of that type is exposed in the dead wall. Do not subtract one just
because it is shown. If a raw physical indicator is independently observed,
count that physical tile and normalize its bonus mapping exactly once.

Unseen tiles include opponents' hands and inaccessible wall tiles. Ukeire is
a count under partial information, not an exact draw probability. A displayed
percentage needs an explicit hidden-tile model, sampling assumptions, and a
defined time horizon. Do not use unrevealed ura as known future value.

## Recommendations

Generate legal actions before ranking them. Calls require simulating the
whole continuation, including restricted discards. Riichi is a declaration
plus a qualifying discard, with a deposit and future hand restrictions.
Kan changes the public bonuses and turn sequence; it is not automatically
good because it increases fu.

An initial recommendation can compare shanten, ukeire, surviving yaku routes,
and likely hand value, then explain defense and placement considerations.
Expose a few meaningful alternatives when these objectives disagree.
Lowest shanten alone can recommend a fast hand that cannot win or a discard
that sacrifices a secure match position.

For defense, evaluate safety per opponent. A tile in one opponent's river is
protected by that opponent's discard-furiten constraint; that does not make
it safe against everyone. Suji only addresses some two-sided waits, and
visible-tile barriers do not eliminate pair, triplet, or unusual waits.
Label these as risk heuristics. Prefer a defensible fold when its expected
placement value exceeds a push, rather than always maximizing hand value.

Later search can sample hidden states consistent with visible tile counts,
simulate legal continuations, and estimate score or placement utility.
Record the seed, budget, opponent model, and confidence interval. Hidden
hands make this an imperfect-information policy problem; a search result
does not establish a guaranteed optimal move.

## Game adapter boundary

The adapter translates observations into the state contract. Compare
predicted legal actions with offered UI actions and retain disagreements as
evidence. The offered action list is useful corroboration, but omitted or
transient UI elements are not proof of a new rule.

Keep recommendation generation free of callbacks. If execution is requested
later, revalidate the hand ID, event sequence, action options, and tile identity
immediately before acting. Tile sort order can change after draws and calls.
An old array index must not be reused as a discard command. Start integration
with observation and replay so failures can be localized before live actions.

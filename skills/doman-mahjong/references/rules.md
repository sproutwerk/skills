# Rules and representation

Read [sources.md](sources.md) for evidence labels. FFXIV-specific settings below
come from the official guide; the representation is a proposed solver contract.

## Tiles and hand shape

Use `m` for characters, `p` for dots, `s` for bamboo, and `z` for honors.
Define `1z..7z` as East, South, West, North, White, Green, Red. Write red fives
as `0m`, `0p`, `0s`, but count each as a five when checking shape. For example,
`405p` contains three tiles, two of which have rank five; only `0p` is red.
These are notation conventions, not discovered FFXIV IDs.

A logical array can use `0..8 = 1m..9m`, `9..17 = 1p..9p`,
`18..26 = 1s..9s`, `27..33 = 1z..7z`. Keep red identity and physical tile
identity separately. The ordinary set has four copies of each of 34 types;
one five in each suit is red. A red five replaces an ordinary five.
The [official introduction](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/)
describes the tile set and red bonuses.

Evaluate three shape families:

| Family | Completion predicate |
| --- | --- |
| Standard | Four groups plus one pair; a group is a suited sequence, triplet, or declared quad |
| Seven Pairs | Seven distinct tile types appearing twice; no declared melds |
| Thirteen Orphans | Every terminal and honor type, with one duplicate; no declared melds |

Honors cannot form sequences. Sequences neither cross suits nor wrap at nine.
Declared groups are fixed. Undeclared tiles may have several valid
decompositions; enumerate them before choosing a score.

If `M` is the number of declared melds and `K` the number of declared kans,
the concealed count at a settled post-discard state is `13 - 3*M`.
After a draw or on completion it is `14 - 3*M`. Each kan contributes four
physical tiles but one structural group, so the corresponding whole-hand
physical counts are `13 + K` and `14 + K`. Intermediate call/kan states need
their own phase; do not reject them using a settled-state count.

## Calls and concealed status

| Action | Structural effect |
| --- | --- |
| Chi | Take the preceding player's discard plus two concealed tiles to form a sequence; then discard without drawing |
| Pon | Take any opponent's discard plus a matching pair; then discard without drawing |
| Daiminkan | Take a discard with three matching concealed tiles; await a replacement draw |
| Ankan | Declare four matching concealed tiles; await a replacement draw; the hand stays closed |
| Shouminkan / kakan | Add a concealed tile to an existing open pon; permit the robbery window before completing the kan |
| Ron | Complete a hand with an opponent's eligible tile; apply yaku and furiten checks |
| Tsumo | Complete a hand with one's own draw; apply yaku checks |

The official introduction says play passes to the player on the right, so the
preceding player sits on the left. Translate relative seats
using observed turn order before trusting any client array index.

Ron is not restricted by chi's seat rule or by pon's pair requirement. A
closed hand can win by ron. The triplet completed by ron is treated as open
for triplet fu/concealed-triplet yaku, while the hand still retains its closed
status for closed-ron fu and riichi. A concealed kan does not prevent riichi.
See the official [action terminology](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/terminology/)
and the baseline scoring reference in [scoring.md](scoring.md).

## FFXIV profile

These settings come from the official
[Advanced Rules](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/special_rule/),
plus the [introduction](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/),
[screen guide](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/navigate/),
and [Starting a Match](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/play_mahjong/)
pages where noted:

| Setting | Doman behavior |
| --- | --- |
| Starting score | 25,000 each (introduction) |
| Match length | Eight hands (East and South) in a full match, four in a quick match (introduction) |
| Dora display | Default "Doman Mahjong" setting shows the bonus tile itself; the "Traditional" setting shows the indicator one below it. Per player, fixed during a match (Starting a Match) |
| Red fives | One per suit, one han each (introduction) |
| Kuitan | Enabled in Novice/Advanced; selectable in four-player matches |
| Atozuke | Enabled |
| Kuikae | Forbidden |
| Final discard | No chi, pon, or kan |
| Riichi | Requires at least 1,000 points; furiten allowed; 1,000-point deposit |
| Post-riichi ankan | Only when the waits are unaffected |
| Ankan robbery | Thirteen Orphans cannot win on a tile used for a concealed kan |
| Multiple ron | Double allowed, and the dealer repeats if one winner is the dealer; triple aborts |
| Yakuman | Twelve listed families; stacking capped at four |
| Kazoe | 13+ han supported |
| First-draw yakuman | Blessing of Heaven/Earth are void after any chi, pon, kan, or concealed kan |
| Pao | Daisangen/Daisuushii; not Shousuushii. On tsumo the liable player pays everything; on another player's deal-in the two split it |
| Honba | 300 points per counter to the winner (screen guide) |
| Bankruptcy | Below zero; zero continues |
| Extension | No West round |
| Dealer repeat | Win or exhaustive-draw tenpai |
| Final leading dealer | Win/tenpai ends match |
| Abortive draws | Nine terminals/honors on the first turn with no call, triple ron, four kans, four same winds; the hand is repeated |
| Four-kan exception | Same player's four continue; fifth aborts |
| Nagashi | Dealer 12,000, non-dealer 8,000; void if any of the player's discards was claimed; the hand still counts as a draw |
| Leftover riichi deposits | To first place at game end |
| Ties | Earlier seat from the first East hand's dealer ranks higher |

The abortive-draw list is exhaustive as written: four riichi declarations are
not on it. The page does not say whether riichi needs a minimum number of
remaining draws. Treat both as documented silence, not as confirmed rules.

Four or eight scheduled hands do not cap the number of hands when repeats
occur. The official end-of-game conditions name South 4; verify the quick-match
final hand before reusing them. The time limits are 80 minutes before the next
hand becomes the last and 120 minutes before the game ends unrecorded; see
[sources.md](sources.md) for a wording conflict on the 80-minute rule.

## Furiten and restricted choices

Maintain the full structural wait set and the player's entire discard history,
including tiles later claimed. An intersection blocks ron on every wait.
Changing the hand can remove discard furiten. Track temporary furiten and
post-riichi pass furiten independently; tsumo remains possible.

FFXIV's [Points to Remember](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/faq/)
explicitly gives a no-yaku completion that still causes temporary furiten:
holding 4-5 where only the 6 gives a yaku, passing a 3 blocks ron on a 6 "for
the rest of the turn". Do not gate this state change solely on whether a Ron
button was offered. After riichi the same page makes the block last until the
hand ends. The exact draw/call/discard reset boundary needs a recorded event
trace; the page's description of a turn is not a callback specification.

Kuikae forbids discarding the claimed type and sequence substitutions: after
claiming `1m` using `2m3m`, an immediate `4m` discard is restricted. Generate
call-plus-discard candidates together. A call with no legal follow-up discard
must not appear as a usable recommendation.

Post-riichi ankan must preserve waits. Advanced Rules words this as no effect
on the "number of waits"; the FAQ words it as not changing the type of hand.
Compare the complete wait sets, not only their sizes;
whether the client checks further decomposition conditions remains open.

# Yaku and scoring

Use a score result containing shape, winning-tile assignment, yaku, bonus han,
fu breakdown, limit tier, and each payer's amount. This makes disagreements
with FFXIV's result screen diagnosable.

## Yaku evaluation

The official [Yaku List](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/yaku_list/)
is the FFXIV catalog. This selection covers common solver decisions:

| Yaku | Closed/open han | Predicate |
| --- | --- | --- |
| Riichi | 1/none | Accepted declaration |
| Menzen tsumo | 1/none | Closed self-draw |
| Yakuhai | 1/1 per value | Dragon/seat/round triplet |
| Tanyao | 1/1 with kuitan | Only suited 2..8 |
| Pinfu | 1/none | Sequences, valueless pair, two-sided completion |
| Iipeikou | 1/none | Identical sequence pair |
| Sanshoku doujun | 2/1 | Same sequence across suits |
| Ittsu | 2/1 | 123,456,789 in one suit |
| Toitoi | 2/2 | Four triplets/quads |
| Chiitoitsu | 2/none | Seven distinct pairs |
| Honitsu | 3/2 | One suit plus honors |
| Chinitsu | 6/5 | One suit, no honors |

Use the full catalog for event yaku, outside hands, concealed triplets, quads,
and yakuman. Do not infer missing entries from another game's scorer. In
particular, the catalog's Nagashi han label is not a normal four-han score:
handle it through the draw settlement rule.

Model each additional yaku as a predicate over a chosen decomposition and
event state, with closed/open values and explicit exclusions. Evaluate every
decomposition and winning-tile placement before selecting the highest legal
payment. For example, `11223344556677m` admits both pair and sequence
interpretations. Do not combine yaku from incompatible interpretations.

Common checks when implementing the full catalog:

- Double riichi replaces riichi; ryanpeikou replaces iipeikou.
- Junchan replaces chanta; both need a sequence, and the pair also qualifies.
- Honroutou can accompany Seven Pairs; Toitoi requires a triplet decomposition.
- A ron-completed triplet cannot supply the third concealed triplet for
  Sanankou. Suuankou on ron needs a pair completion.
- Chinitsu and Honitsu are exclusive. Actual yakuman use their own multiplier,
  without ordinary yaku or dora increasing it.

These are baseline Riichi implementation checks. Confirm the exact FFXIV
catalog and result screen before treating a newly implemented predicate as
validated.

## Fu and payment baseline

The [WRC 2025 rules](https://www.worldriichi.org/s/WRC-Rules-2025-42fx.pdf),
sections 11.2 and 11.3, provide an inspectable calculation reference. WRC is
not the FFXIV profile. In particular, its kiriage and double-wind-pair policy
must not silently become FFXIV defaults.

| Contribution | Fu |
| --- | --- |
| Base | 20 |
| Closed ron | +10 |
| Tsumo, except pinfu | +2 |
| Edge, middle, or pair completion | +2 |
| Value pair | +2; double wind requires profile |
| Open simple triplet | +2 |
| Closed simple triplet | +4 |
| Open simple quad | +8 |
| Closed simple quad | +16 |

Double the group contribution for terminals/honors. Round total fu upward
to ten, except Seven Pairs stays exactly 25. Pinfu tsumo is 20; closed pinfu
ron is 30. An open hand otherwise totaling 20 receives 30, but still needs
a yaku.

For ordinary hands, let `B = min(2000, fu * 2^(han+2))`. Limit bases:
5 han 2000; 6..7 3000; 8..10 4000; 11..12 6000; 13+ 8000.
Actual yakuman use `8000 * multiplier`.

Define `ceil100(x) = 100 * ceil(x / 100)`:

| Win | Each payment before extras |
| --- | --- |
| Non-dealer ron | Discarder pays `ceil100(4*B)` |
| Dealer ron | Discarder pays `ceil100(6*B)` |
| Non-dealer tsumo | Dealer pays `ceil100(2*B)`; others each `ceil100(B)` |
| Dealer tsumo | Each opponent pays `ceil100(2*B)` |

The official [screen guide](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/navigate/)
gives the winner 300 points per honba. The usual split is 300 from the
discarder on ron or 100 from each payer on tsumo; the guide does not spell out
the tsumo split. Riichi deposits go to the winner. Multiple winners and
liability need separate settlement.

The introduction's beginner table lists 1 han 1,000, 2 han 2,000, 3 han 3,900,
4–5 han 8,000, 6–7 han 12,000, 8–10 han 16,000, 11–12 han 24,000, and 13+ han
32,000. Those are non-dealer ron values at 30 fu and the limit tiers, so the
limit bases above match the official ones. The table ignores fu, so do not
use it to score a hand.

Keep scoring arithmetic separate from match transfers. A result needs to
explain why score changes differ from the displayed hand value: honba,
deposits, pao, multiple winners, and draw settlements are possible causes.
The baseline assumes ordinary settlement without those exceptions.

## FFXIV details still requiring evidence

Do not claim an exact score where an unresolved setting changes it. Return
the alternatives with the missing setting, or label the configured assumption.

| Unresolved detail | Useful discriminating observation |
| --- | --- |
| Kiriage mangan | 4 han 30 fu or 3 han 60 fu: non-dealer ron 7,700 versus 8,000. The beginner table groups 4 han with 5 han at 8,000, which hints at kiriage but may be a simplification |
| Double-wind pair fu | A hand whose rounded fu differs under +2 versus +4 |
| Special double yakuman | Pure Kokushi, pure Chuuren, Suuankou tanki, Daisuushii result. Advanced Rules list the pure and pair-wait forms as "included" in the single yakuman, which suggests no doubling |
| Honba split and noten transfers | Exhaustive draw and subsequent tsumo, all scores before/after |
| Double ron | Deposit recipient and honba allocation |
| Pao with stacked yakuman | Liability portion when only one yakuman is liable, honba |
| Nagashi | Who pays the 12,000/8,000, whether a player who called can qualify, simultaneous qualifiers, noten/riichi settlement |
| Kan timing | Robbed kan versus new dora and replacement draw events |

These are questions for FFXIV evidence, not claims that all options are used.

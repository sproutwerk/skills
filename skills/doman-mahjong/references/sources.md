# Sources and conflicts

Reviewed 2026-11-13 against the live pages listed below. This is documentary
research; no live FFXIV match was observed for this skill. Runtime mappings
and unconfirmed rules remain open.

## Authority

Prefer current in-game help and reproducible FFXIV observations for the target
build, then the detailed Lodestone rules. Use the beginner page for orientation.
Community summaries help find questions but can contradict official rules.
Generic Riichi references and implementations establish comparison baselines.

| Source | Use |
| --- | --- |
| [Lodestone introduction](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/) | Tiles, turn flow, starting score, match length, basic calls, red dora, beginner point table |
| [Advanced Rules](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/special_rule/) | FFXIV profile and match rules |
| [Points to Remember](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/faq/) | Furiten, restricted discards, post-riichi kan |
| [Yaku List](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/yaku_list/) | FFXIV yaku catalog |
| [Terminology](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/terminology/) | Calls, melds, tile naming |
| [Screen guide](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/navigate/) | UI labels to compare against observations; honba value |
| [Starting a match](https://na.finalfantasyxiv.com/lodestone/playguide/contentsguide/goldsaucer/doman-mahjong/play_mahjong/) | Modes, time limits, hints, dora indicator setting |
| [Community Doman Mahjong page](https://ffxiv.consolegameswiki.com/wiki/Doman_Mahjong) | Secondary overview; lists the quick-match duty variants |
| [Japanese mahjong](https://en.wikipedia.org/wiki/Japanese_mahjong) | General overview |
| [WRC rules index](https://www.worldriichi.org/wrc-rules) | Primary generic Riichi rules and scoring reference |
| [WRC 2025 rulebook](https://www.worldriichi.org/s/WRC-Rules-2025-42fx.pdf) | Sections 11.2/11.3 for scoring arithmetic |
| [WRC 2025 scoring table](https://www.worldriichi.org/s/WRC-Scoring-Table-2025.pdf) | Comparison values; includes WRC-specific rounding |
| [Path of Houou deal-in analysis](https://pathofhouou.blogspot.com/2020/05/analysis-tile-deal-in-rates-by-live-suji.html) | Tenhou deal-in rates used in [defense.md](defense.md) |
| [MahjongRepository/mahjong](https://github.com/MahjongRepository/mahjong) | Candidate independent hand-calculation reference |

The library has not been installed or evaluated for FFXIV compatibility.
Pin a revision and inspect its rule options before using it for differential
checks. A Tenhou-compatible scorer is not automatically Doman-compatible.
The WRC scoring table is an image-only PDF; its values were not re-read in
this review. The Wikipedia page was not rechecked and no rule here rests on it.

## Conflicts to preserve

- The community page allows Kokushi to rob an ankan. Official Advanced Rules
  explicitly forbid that in FFXIV. Use the official rule.
- Broad descriptions saying that any kan opens a hand omit the ankan
  exception. Preserve a separate concealed-kan type.
- Statements about temporary furiten expiring after a turn, draw, or discard
  do not establish the client's event ordering. Keep that boundary unresolved
  until a trace distinguishes the cases.
- The official beginner explanation of tsumo payments is abbreviated: it says
  the other three players "each pay an even share" and then that the dealer
  pays double. Calculate each payer's share; do not divide a ron total into
  equal thirds.
- The beginner point table has one value per han and puts 4 han with 5 han
  at 8,000. It ignores fu, so it neither confirms nor rules out kiriage mangan.
- Advanced Rules say the dora display shows the bonus tile itself. Starting
  a Match documents a per-player setting that switches to a traditional
  indicator. Both are official; the first describes the default.
- Advanced Rules say the next hand is the last once 80 minutes have passed
  since the match started. Starting a Match says this happens when "a single
  hand lasts more than 80 minutes". Treat the Advanced Rules wording as the
  likelier reading and leave it unconfirmed.
- The official yaku catalog gives Nagashi a four-han label, while its advanced
  rules specify a 12,000/8,000 payment and draw continuation. Use a settlement
  event, not ordinary four-han/fu arithmetic.
- WRC 2025 uses kiriage mangan and a 2-fu double-wind pair. Neither is
  documented for FFXIV.
- Official text says a fifth kan aborts after one player's four kans. Preserve
  that documented rule without inventing a fifth replacement tile or claiming
  that the client offers such an action.

## Terms

Original web material retains its owners' terms. This skill links to it and
does not bundle it. See [NOTICE.md](../NOTICE.md).

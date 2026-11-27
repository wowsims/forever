# EV-0222: Venom goes up before its Shank's damage

- Claim: Venom (1310703) goes up before the damage of the Shank that applies it, so the poison that Shank's hit procs
  already rolls at Venom's 20 points more chance and deals its 30% more damage.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 474 rogue streams (344 character names) that show
  Venom, in 465 of the 521 reports whose abilities list it, too many to list
- Game build: 1.60.1.70938 to 1.60.1.71139 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20
  18:23 UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU, 71139 from 2026-11-25
  19:14 UTC in the US and the EU); the hotfix each player ran is not recorded, and the reports from after the
  2026-11-23 hotfix read the same
- Date: 2026-11-13 to 2026-11-26
- Pull request: #1269

Each Shank's damage event is matched with the Venom apply or refresh logged within 50 ms of it. A poison hit is an
Instant Poison damage event with a normal hit (no crit, no partial resist) on the Shank's target, 0 to 50 ms after the
Shank's damage. A Shank is fresh when no Venom was up before it. Each poison hit is taken as a share of the median
normal hit of the same player and poison rank with Venom down.

| Venom's log time against its Shank's damage | Shanks |
| ------------------------------------------- | -----: |
| Before it                                   | 33,743 |
| The same millisecond                        | 36,359 |
| After it                                    |      1 |

| Instant Poison hits    |    Hits | Median share | Under 0.8 | 0.8 to 1.12 | 1.12 to 1.5 | 1.5 and over |
| ---------------------- | ------: | -----------: | --------: | ----------: | ----------: | -----------: |
| From a fresh Shank     |   6,589 |         1.24 |        82 |         896 |       5,554 |           57 |
| Other hits, Venom up   | 128,800 |         1.29 |     1,885 |       5,257 |     118,480 |        3,178 |
| Other hits, Venom down |  69,129 |         0.99 |     1,346 |      64,138 |       3,625 |           20 |

The poison a fresh Shank procs reads like a poison under Venom. The reports from after the 2026-11-23 hotfix (111
streams in 111 reports) give the same: 1.24 on 2,491 fresh-Shank poison hits, 1.28 under Venom and 0.98 with Venom
down. About one fresh-Shank poison hit in seven reads in the 0.8 to 1.12 band, against one in 24 under Venom.
`venom_order.py` counts them again (6,503 hits, 817 in that band): 161 of the 817 follow a Shank with no Venom apply
logged within 50 ms, against 11 of the 5,547 in the 1.12 to 1.5 band, and a white swing in the same 50 ms is no more
common among them (98 of 817) than among those (887 of 5,547). The rest are not explained.

The chance is counted per Shank. A stream's odd-numbered Shanks pick the streams whose off hand carries Instant Poison
(10 Shanks or more, an Instant Poison event on the Shank's target 0 to 50 ms after at least 0.15 of them), and its
even-numbered Shanks are counted: fresh Shanks in the streams that show Venom, every Shank in the rogue streams that
show none. Full resists and the Shanks that put no Venom up stay in the counts.

| Shanks counted                     | Streams | Shanks | With an Instant Poison | Share |
| ---------------------------------- | ------: | -----: | ---------------------: | ----: |
| Fresh Shanks, streams with Venom   |     119 |  9,244 |                  3,531 | 0.382 |
| Every Shank, streams with no Venom |     516 | 42,935 |                  8,349 | 0.194 |

Instant Poison's chance is 0.2 (EV-0087) and Improved Poisons adds at most 10 points, so a poison rolled without
Venom stays at 0.3 or under; 65 of the 117 Venom streams with 20 or more counted Shanks read 0.4 or more.

Scripts and output, in `forever-venom-1127`: `venom.py` (`venom_out.txt`, and `venom_out_posthotfix.txt` with
`SINCE=1795482000000`), `venom_order.py` (`venom_order_out.txt`), `rate.py` (`rate_out.txt`) and `control.py`
(`control_out.txt`).

# EV-0188: Exorcism crits from 2026-11-25

- Claim: Exorcism, at every rank (879, 5614, 5615, 10312, 10313, 10314), could not crit until a server change on
  2026-11-25 at about 20:03 UTC, and since then crits like the paladin's other magic Holy spells. Its row has the magic
  defense type and no cannot-crit attribute on both sides of the change, and no build or hotfix changed its rows, so
  the game files show neither state.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every report from 2026-11-05 to 2026-11-25 where a
  player deals Exorcism, Holy Wrath or Hammer of Wrath damage; 12,090 of them hold a landed Exorcism hit (US 5,879,
  EU 5,540, KR 346, CN 319, TW 6); all fights.
- Game build: 1.60.1.70790 to 1.60.1.71091 by event time against the time each region's CDN first listed each build,
  with hits on every build in between and no crit on any; the hotfix each player ran is not recorded
- Date: 2026-11-05 12:32 to 2026-11-25 05:59 UTC
- Pull request: #1211
- Later count: the change, §Count of the 2026-11-25 reports and combat log

## Method

Every damage event of the six Exorcism ranks was counted by hit type. A landed hit is a normal hit or a crit; misses,
immunities and other outcomes are left out. The 293,088 landed hits come from 9,924 characters in 14,827
player-report pairs. 7,866 of the hits are second copies of play logged in two reports, and counting each play once
leaves 285,222. None is a crit, on any day of the count.

| Build        | Landed hits | Crits |
| ------------ | ----------: | ----: |
| 1.60.1.70790 |       5,044 |     0 |
| 1.60.1.70840 |       1,462 |     0 |
| 1.60.1.70862 |      28,427 |     0 |
| 1.60.1.70888 |      17,801 |     0 |
| 1.60.1.70938 |      65,037 |     0 |
| 1.60.1.70959 |      48,191 |     0 |
| 1.60.1.71037 |     118,161 |     0 |
| 1.60.1.71091 |       8,965 |     0 |
| All          |     293,088 |     0 |

## Controls

The same paladin's other magic-defense Holy spells crit in the same reports. Each control counts the player-report
pairs with 20 or more landed control hits, and the expected Exorcism crits are each pair's Exorcism hits at that
pair's own control crit rate.

| Control                                          | Pairs | Control crit rate | Exorcism hits | Crits expected | Crits seen |
| ------------------------------------------------ | ----: | ----------------: | ------------: | -------------: | ---------: |
| Holy Wrath (2812, 10318)                         | 1,978 |            15.82% |       106,580 |         16,326 |          0 |
| Holy Shock damage (25902, 25911, 25912, 1311604) |   344 |            19.60% |         9,596 |          1,924 |          0 |

Over every report, Holy Wrath crits on 15.71% of 135,521 landed hits and Holy Shock damage on 18.51% of 74,146.
Hammer of Wrath is no control: its row has the ranged defense type, and it rolls a ranged crit.

## Count of the 2026-11-25 reports and combat log

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 337 reports with a landed Exorcism hit from
  2026-11-25 05:59 to 21:30 UTC (EU 8,576 hits, US 3,923, KR 344, CN 165); and an EU player's three combat logs of
  2026-11-25, holding the Exorcism hits of six paladins, all rank 5 (10313) but one rank 2 (5614).
- Game build: the reports 1.60.1.71091 (US, EU) and 1.60.1.71037 (CN, KR) by event time, and from 19:14 UTC, when US
  and EU listed 1.60.1.71139, either of 1.60.1.71091 and 1.60.1.71139 there. The first combat log ran 1.60.1.71091 to
  20:22 UTC with no relog, so it spans the change; the other two ran 1.60.1.71139 after a relog. The Exorcism rows of
  1.60.1.71139 with its hotfixes, read after the change, match those of 1.60.1.71091 in every field, so the change came
  from the server.
- Date: 2026-11-25 05:59 to 21:30 UTC
- Pull request: #1238

Landed Exorcism hits, with Holy Wrath's crit rate as the control, counted as in §Method; play logged in two reports
counts in both. The windows split at the first crit.

| Source                    | Window (UTC)         | Exorcism landed | Exorcism crits | Holy Wrath crits |
| ------------------------- | -------------------- | --------------: | -------------: | ---------------: |
| Reports                   | 05:59 to 20:03:44    |          12,146 |              0 |   15.9% of 9,131 |
| Reports                   | 20:03:45 to 21:30    |             862 |    132 (15.3%) |     17.5% of 799 |
| Combat log, 1.60.1.71091  | 18:20 to 20:03:12    |              45 |              0 |           0 of 2 |
| Combat log, 1.60.1.71091  | 20:04:19 to 20:22:40 |              22 |              3 |                  |
| Combat logs, 1.60.1.71139 | 20:29 to 21:08       |              42 |              3 |           1 of 5 |

In the reports the last Exorcism hit before the first crit lands at 20:03:39 (EU), the first crit at 20:03:45 (EU)
and the first US crit at 20:04:30. The 73 hits from 19:59 to 20:03:44 (US 41, EU 32) hold no crit, where a 15% chance would
miss all of them about 1 time in 140,000, so the change is not earlier than 20:03 and reached both regions together.
From then on Exorcism crits at about Holy Wrath's rate. The bug report, ClassicWoWCommunity/forever-bugs#631, was
closed at 20:39 UTC as fixed in a hotfix.

# EV-0188: Exorcism crits since the 2026-11-25 server change

- Claim: Exorcism crits at the caster's spell crit chance, as Holy Wrath does, at every rank (879, 5614, 5615, 10312,
  10313, 10314). Its row has the magic defense type and no cannot-crit attribute. Until a server-side change at about
  20:03 UTC on 2026-11-25 it did not crit at all, and the game files do not show that either.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every report from 2026-11-05 to 2026-11-25 where a
  player deals Exorcism, Holy Wrath or Hammer of Wrath damage; 12,090 of them hold a landed Exorcism hit (US 5,879,
  EU 5,540, KR 346, CN 319, TW 6); all fights.
- Game build: 1.60.1.70790 to 1.60.1.71091 by event time against the time each region's CDN first listed each build,
  with hits on every build in between and no crit on any; the hotfix each player ran is not recorded
- Date: 2026-11-05 12:32 to 2026-11-25 05:59 UTC
- Pull request: #1211
- Later count: the change and the crit rate after it, §Count of the 2026-11-25 reports

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

## Count of the 2026-11-25 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 244 reports starting 2026-11-25 16:00 UTC or later
  that show an Exorcism rank, indexed up to 20:51 UTC; every landed Exorcism and Holy Wrath damage event, a play
  logged in two reports counted once
- Game build: 1.60.1.71091; the change is server-side with no hotfix push id, and the owner saw an Exorcism crit in game
  at about 20:17 UTC
- Date: 2026-11-25 16:00 to 21:12 UTC
- Pull request: #1223

The last landed Exorcism hit of the no-crit period logs at 20:03:02 UTC and the first crit at 20:03:05 UTC. Before
that, none of 5,010 landed hits crit; from it on, 122 paladins crit with Exorcism at about the rate Holy Wrath crits
for them.

| Hour (UTC) | Exorcism landed | Exorcism crits | Holy Wrath landed | Holy Wrath crits |
| ---------- | --------------: | -------------: | ----------------: | ---------------: |
| 16         |             423 |              0 |                56 |                3 |
| 17         |           1,131 |              0 |               130 |               23 |
| 18         |           1,345 |              0 |               203 |               35 |
| 19         |           2,030 |              0 |               300 |               54 |
| 20         |           1,818 |            260 |               256 |               45 |
| 21         |             512 |             82 |                89 |               15 |

From 20:03:05 UTC on, Exorcism crits on 342 of 2,249 landed hits (15.2%) and Holy Wrath on 57 of 328 (17.4%). The 9
paladins with 10 or more Holy Wrath hits in that window crit 61 of 350 Exorcism hits, against 79 at each one's own
Holy Wrath rate.
Revelation names Holy Wrath and not Exorcism (EV-0149), so the control's rate carries the buff's crits.

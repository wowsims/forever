# EV-0212: Seal of Light Ranks 3 and 4, Seal of Wisdom Ranks 2 and 3 at 15 a minute, and their judgements at 50%

- Claim: Seal of Light Ranks 3 and 4 and Seal of Wisdom Ranks 2 and 3 proc 15 times a minute on the weapon's speed, and
  Judgement of Light Ranks 3 and 4 and Judgement of Wisdom Ranks 2 and 3 fire on a flat 50% of landed melee swings, as
  Rank 1 and 2 do (EV-0017, EV-0018).
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every paladin stream on disk whose report lists the
  rank's seal or judgement, 502 reports and 33,134 lone swings (the reports per rank are in the table); and 34 raw
  combat logs with other players' swings in them, 4,616 lone swings (the attackers per rank are in the second table)
- Game build: 1.60.1.70862 to 1.60.1.71091 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-10 to 2026-11-24
- Pull request: not yet opened

EV-0017's method on the raw events. A swing counts when it is a landed white swing (hit, crit, glance, crush) with no
other hit by the attacker and no other swing within 400 ms, inside the seal's 30 s aura or against a target carrying
exactly one rank of the judgement, whose 40 s aura is its window. It is matched to the first unused proc of that rank (a
heal or a mana gain by the attacker) from 20 ms before to 300 ms after the swing. The weapon speed is the item's speed
from the fight's gear, checked against the median swing gap. A seal's rate is the 15 a minute fit that maximizes the
likelihood over swings at their speeds (2% of procs allowed to go missing); a judgement's chance is the share of swings
that matched, with a Wilson interval. Most procs land at 0 ms after the swing; up to about 10% land 50 to 250 ms late.

## Warcraft Logs paladins

| Proc                  | Reports | Paladins | Lone swings | Procs         | Reading                       |
| --------------------- | ------- | -------- | ----------- | ------------- | ----------------------------- |
| Seal of Light 3       | 37      | 38       | 286         | 164 (57.3%)   | 14.45 a minute [13.10, 15.75] |
| Seal of Light 4       | 69      | 69       | 924         | 520 (56.3%)   | 14.75 a minute [14.00, 15.45] |
| Seal of Wisdom 2      | 131     | 131      | 5,993       | 3,560 (59.4%) | 14.75 a minute [14.50, 15.00] |
| Seal of Wisdom 3      | 337     | 341      | 11,074      | 6,836 (61.7%) | 15.05 a minute [14.80, 15.25] |
| Judgement of Light 3  | 36      | 37       | 300         | 146 (48.7%)   | 48.7% [43.1, 54.3]            |
| Judgement of Light 4  | 69      | 69       | 851         | 427 (50.2%)   | 50.2% [46.8, 53.5]            |
| Judgement of Wisdom 2 | 131     | 131      | 4,565       | 2,329 (51.0%) | 51.0% [49.6, 52.5]            |
| Judgement of Wisdom 3 | 326     | 332      | 9,141       | 4,678 (51.2%) | 51.2% [50.2, 52.2]            |

The seal rates fit on the swings with a known weapon speed (274, 889, 5,620 and 9,737 of them). 15 a minute predicts
165.4, 510.0, 3,440.9 and 6,085.7 procs for the four seals on the swings of the full counts, 16 a minute 176.4, 544.0,
3,668.4 and 6,480.0, and 161, 498, 3,346 and 6,021 were seen. Of the swings with a weapon of 3.5 s or slower, where a
15 a minute rate is nearly certain per swing, Seal of Wisdom 2 procced on 546 of 653 (593.3 predicted) and Seal of
Wisdom 3 on 1,098 of 1,255 (1,132.0).

By time of the logs (before the 2026-11-20 23:47 UTC push, to 2026-11-24 01:07 UTC, after), the seal rates read:

| Seal             | Before                       | Middle                       | After                      |
| ---------------- | ---------------------------- | ---------------------------- | -------------------------- |
| Seal of Light 3  | 14.50 [13.00, 15.85] (230)   | 13.30 [9.30, 17.00] (36)     |                            |
| Seal of Light 4  | 14.70 [13.65, 15.65] (459)   | 14.95 [13.75, 16.00] (393)   | 12.85 [8.65, 17.15] (37)   |
| Seal of Wisdom 2 | 15.45 [15.10, 15.80] (2,762) | 13.90 [13.50, 14.30] (2,118) | 15.10 [14.40, 15.65] (740) |
| Seal of Wisdom 3 | 15.05 [14.85, 15.25] (6,411) | 14.85 [14.55, 15.35] (2,576) | 15.00 [14.20, 15.70] (750) |

Seal of Wisdom 2's middle reading is below 15 and its first above, with no common rate between them; the three
periods pooled read 14.75.

## Raw combat logs of other players

| Proc                  | Attackers | Lone swings | Procs       | Reading                       |
| --------------------- | --------- | ----------- | ----------- | ----------------------------- |
| Seal of Light 3       | 10        | 156         | 98 (62.8%)  | 16.25 a minute [14.65, 17.90] |
| Seal of Light 4       | 10        | 101         | 65 (64.4%)  | 16.45 a minute [14.05, 18.70] |
| Seal of Wisdom 2      | 66        | 1,177       | 897 (76.2%) | 16.10 a minute [15.45, 16.45] |
| Seal of Wisdom 3      | 15        | 302         | 185 (61.3%) | 15.15 a minute [13.85, 16.30] |
| Judgement of Light 3  | 17        | 403         | 206 (51.1%) | 51.1% [46.2, 56.0]            |
| Judgement of Light 4  | 86        | 596         | 307 (51.5%) | 51.5% [47.5, 55.5]            |
| Judgement of Wisdom 2 | 66        | 1,313       | 663 (50.5%) | 50.5% [47.8, 53.2]            |
| Judgement of Wisdom 3 | 44        | 568         | 283 (49.8%) | 49.8% [45.7, 53.9]            |

The raw logs read Seal of Light 3 and 4 and Seal of Wisdom 3 with 15 a minute inside their intervals. Seal of Wisdom 2
reads 16.10 [15.45, 16.45] on 1,176 swings with a known speed, above 15 and above the paladins' 14.75; 66 attackers
carry it, one of them 269 of the swings.

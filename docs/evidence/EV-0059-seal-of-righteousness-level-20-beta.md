# EV-0059: Seal of Righteousness hand factors and spell power coefficients

- Claim: a Seal of Righteousness proc is its value times the weapon speed times a hand factor (1.2 for two-handers, 0.85
  for one-handers from 2.0 speed up, 0.017 less for each tenth of a second under 2.0), plus one spell power coefficient
  per hand type: 0.1 on one-handers and 0.12 on two-handers. The tooltip's 0.0765 fits none of the one-handers.
- Sources: test log on the beta at level 20, ranks 1 to 3, eighteen vendor weapons from 1.7 to 3.7 speed, with and
  without spell power and Judgement of the Crusader
- Game build: 1.60.1 beta
- Date: 2026-09-26
- Pull request: #110
- Later count: the spell power coefficients, §Count of the 2026-10-16 test
- Later count: the one-hander factor under 2.0 speed, §Count of the 2026-11-05 to 2026-11-26 reports

```
proc = value/100 x speed x hand(speed) + coefficient x (spell power + Judgement of the Crusader)
  two-hander:  hand = 1.2,                                coefficient 0.12
  one-hander:  hand = 0.85 - 0.17 x max(0, 2.0 - speed),  coefficient 0.10
```

| Reading                           | Measured                                                                                                                                    |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Two-hander factor                 | 1.2, exact from 2.6 to 3.7 speed (the pull request states it held at every speed and rank)                                                  |
| One-hander factor, 2.0 speed up   | 0.85                                                                                                                                        |
| One-hander factor under 2.0 speed | §Count of the 2026-11-05 to 2026-11-26 reports; measured from 1.6 to 1.9, extrapolated below                                                |
| Spell power coefficient           | §Count of the 2026-10-16 test; flat across ranks 1 to 3; the judgement dummy's per rank number (0.185 at rank 3, 0.058 at rank 1) is not it |
| Judgement of the Crusader         | counts as spell power at the same coefficient                                                                                               |
| Crits                             | double the unrounded total                                                                                                                  |

The full test log was kept in a write-up outside the repository.

## Count of the 2026-10-16 test

- Sources: test on the beta at level 30, rank 4 (judgement 20282, per hit value 541 + 31 x 4 = 665), three one-handers and
  two two-handers, each hit at 0 and 50 spell power with no talents, buffs or Judgement of the Crusader
- Game build: 1.60.1.70425 beta
- Date: 2026-10-16
- Pull request: #473

At 0 spell power the weapon part, 665/100 x speed x hand, matches every weapon, crits included. A hit rounds to the
integer either side of its value and a crit doubles the unrounded value before rounding, so each set of hits and crits at
50 spell power bounds the weapon's value and with it the coefficient.

| Weapon           | Hand | Speed | Base  | Hits at 0 SP | Crits at 0 SP | Hits at 50 SP | Crits at 50 SP | Value      | Coefficient    |
| ---------------- | ---- | ----- | ----- | ------------ | ------------- | ------------- | -------------- | ---------- | -------------- |
| Stock Shortsword | 1H   | 1.80  | 9.77  | 9, 10        | 19, 20        | 14, 15        | 30             | 14.5 to 15 | 0.095 to 0.105 |
| Cutlass          | 1H   | 2.20  | 12.44 | 12, 13       | 25            | 17, 18        | 34, 35         | 17 to 17.5 | 0.091 to 0.101 |
| Cruel Barb       | 1H   | 2.80  | 15.83 | 15, 16       | 31, 32        | 20, 21        | 41, 42         | 20.5 to 21 | 0.093 to 0.103 |
| Espadon          | 2H   | 2.90  | 23.14 | 23, 24       | 46, 47        | 29            | 58, 59         | 29 to 29.5 | 0.117 to 0.127 |
| Corpsemaker      | 2H   | 3.80  | 30.32 | 30, 31       | 60, 61        | 36, 37        | 72, 73         | 36 to 36.5 | 0.114 to 0.124 |

All three one-handers fit a single value from 0.095 to 0.101, and 0.10 sits in that range. One coefficient holds from 1.8
to 2.8: it does not grow with speed and does not take the fast weapon penalty. The 2.2 and 2.8 one-handers gain the same
5 damage from 50 spell power. With 0.0765 the Stock Shortsword would hit for about 13.6 and the Cutlass for about 16.3,
where they hit for 14 to 15 and 17 to 18. The two-handers overlap from 0.117 to 0.124, so 0.12 holds there at both
speeds. Below 1.8 speed the coefficient is not measured.

## Count of the 2026-11-05 to 2026-11-26 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000); §Segments under 2.0 speed lists each fast segment with
  its report and fights, §Control reports the rest
- Game build: 1.60.1.70790 to 1.60.1.71139 by report date (70840 from 2026-11-06 23:22 UTC, 70862 from 2026-11-07 04:43
  UTC, 70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC,
  70959 from 2026-11-17 15:19 UTC in the US and 2026-11-18 02:04 UTC in the EU, 71037 from 2026-11-20 18:23 UTC, 71091
  from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU, 71139 from 2026-11-25 19:14 UTC; Korea and
  China, which hold one report under 2.0 speed and 29 control reports, take 70888 on 2026-11-11, 70959 on 2026-11-18 and
  71139 on 2026-11-25 and list no 71091). Reports under 2.0 speed per build: 70790 5, 70862 7, 70888 2, 70938 11, 70959
  6, 71037 5, 71091 1. Of the 279 hotfix exports of these weeks 37 carry a spell diff, and they hold one change to these
  spells: the 2026-11-13 18:41 UTC push moved every judgement's defense type from melee to ranged and left the
  judgement's damage and its per hit number as they were. The controls read the same before and after 70862, 70938, that
  push and 71037, §Factor per speed. The hotfix each player ran is not recorded
- Date: 2026-11-05 to 2026-11-26
- Pull request: #1258

### Method

A report states neither the paladin's level nor their spell power, so the count keeps the hits where both can be read
some other way.

- Level: a Judgement of Righteousness hit rolls its row's damage at the caster's level. The range of a wearer's
  judgements cuts their hits into stretches of one level, and the judgements' mean picks the level: it has to sit
  within 3 standard errors of that level's roll and over 4 from every other level's. The seal's hits never choose the
  level.
- Spell power, first sample: wearers whose gear holds no spell damage and no Holy damage. Second sample: wearers whose
  gear holds 3 to 129 spell power in the sim's item database (4 to 84 under 2.0 speed). A piece the database lacks
  counts as none, and a wearer is left out when the game files row of such a piece, or of an enchant or random suffix on
  it, carries spell damage or a caster weapon flag. Their hits lose 0.1 times the gear's spell power (0.12 on a
  two-hander) and their judgements 0.429 times it; a segment whose judgements then sit over 3 standard errors off the
  level's roll is dropped, since they show spell power the gear does not.
- Left out: wearers with a point in Vengeance or Champion of the Light, hits on a target that holds Judgement of the
  Crusader, hits resisted, absorbed or blocked in part, hits whose amount differs from the unmitigated amount, and
  fights another report already holds.
- A segment is one wearer with one main hand, one rank, one level and one gear state, 30 hits or more. A hit logs the
  integer on either side of its value and a crit the integer on either side of twice it, so a segment in one state shows
  two adjacent amounts. A segment with over 3 % of its hits elsewhere is dropped. In a kept segment the hits off the
  pair leave the value and the hit count (22 hits in 12 segments under 2.0 speed), and a crit counts from one under
  twice the lower amount to three over it.
- A segment's factor is its value times 100 over (M x per hit number x speed). The value is the summed amounts over the
  count of hits plus twice the count of crits, the per hit number is the judgement dummy's at the level read, and M is 1
  plus 0.03 for each point in Improved Seals. A segment over 2.5 % from its speed's median is dropped, 6 of 47 under 2.0
  speed: a level read one off moves the factor 3 to 7 %. Two of the six sit further out than any level allows, 1.956 at
  1.9 and 0.683 at 1.7.
- A speed's factor is the one that fits its segments' hits best, each amount scored as the split it is: a hit of value
  v logs the higher integer with chance frac(v). The interval is where the log-likelihood falls 2 under the best, and
  the shapes are scored the same way over every segment under 2.0 speed.

### Factor per speed

| Speed | Factor | 95 % interval    | Segments | Wearers |  Hits | Gear spell power | 0.85 - 0.17 x (2.0 - speed) |
| ----- | ------ | ---------------- | -------: | ------: | ----: | ---------------- | --------------------------- |
| 1.6   | 0.7815 | 0.7785 to 0.7840 |        6 |       6 |   536 | none             | 0.782                       |
| 1.7   | 0.7985 | 0.7925 to 0.8050 |        4 |       4 |   567 | 4 to 22          | 0.799                       |
| 1.8   | 0.8135 | 0.8065 to 0.8200 |        2 |       1 |   352 | none             | 0.816                       |
| 1.9   | 0.8340 | 0.8325 to 0.8355 |       12 |      10 | 1,059 | none             | 0.833                       |

With the second sample added, 1.6 reads 0.7815 (7 segments, 748 hits), 1.8 reads 0.8140 (0.8080 to 0.8200, 3 segments,
410 hits) and 1.9 reads 0.8325 (0.8315 to 0.8330, 27 segments, 2,876 hits). Under 2.0 speed the count holds 41 segments
from 36 wearers in 37 reports, 4,601 hits.

Controls, weighted means per speed over both samples: one-handers from 2.0 to 2.9 speed read 0.848 to 0.854 (340
segments, 255 wearers, 37,190 hits; 2.0 to 2.7 scored as above gives 0.849), and two-handers from 2.3 to 3.8 speed read
1.198 to 1.203 (326 segments, 234 wearers, 27,839 hits). Cut at 2026-11-07 04:43, 2026-11-13 01:36, 2026-11-13 18:41
and 2026-11-20 18:23 UTC, the one-handers read 0.8492, 0.8504, 0.8489, 0.8485 and 0.8507 and the two-handers 1.2005,
1.1991, 1.2001, 1.2012 and 1.2005, each weighted by hits.

### Shape

Log-likelihood of each shape over the segments under 2.0 speed, less that of the first row:

| Shape                                   | No gear spell power (20 segments) | Both samples (41 segments) |
| --------------------------------------- | --------------------------------: | -------------------------: |
| 0.85 - 0.2 x (2.0 - speed)              |                                 0 |                          0 |
| min(0.85, 1.2 x (speed - 0.57) / speed) |                              -0.2 |                      -79.2 |
| 0.85 - 0.175 x (2.0 - speed)            |                             +22.2 |                      +25.3 |
| 0.85 - 0.17 x (2.0 - speed)             |                             +23.0 |                      +24.8 |
| 0.85 - (2.0 - speed) / 6                |                             +22.3 |                      +22.7 |
| 0.85 x (speed / 2)^0.385                |                             +22.9 |                      +23.3 |

The slope a of 0.85 - a x (2.0 - speed) fits at 0.170 (95 %: 0.165 to 0.177) without gear spell power and at 0.173
(0.168 to 0.178) over both samples; 0.2 sits 6.8 and 7.1 standard deviations off. The two Mace segments at 1.9 read
0.845 and 0.854, and without them the slope is 0.172 (0.166 to 0.179) and 0.174 (0.169 to 0.179).

What the count leaves open:

- The straight line and the power law fit equally well from 1.6 speed up. At 1.5 they read 0.765 and 0.761.
- Under 1.6 speed nothing is measured. The reports hold 15,500 hits with a 1.5 weapon, 14 of them from a wearer with no
  gear spell power, and the rest form no segment: gear with a random suffix the item database does not price, over 150
  spell power, judgements that show spell power the gear does not, a state that changes inside the stretch, or a target
  that holds Judgement of the Crusader.
- 1.7 comes from the second sample alone: the first sample's one 1.7 segment (Thief's Blade, rank 3, 54 hits,
  gQwb8cW2zfvCdXZT/1) reads 0.683 and falls to the 2.5 % cut. So 1.7 rests on the 0.1 coefficient, which is measured
  from 1.8 up. Priced with the spell power its judgements read in place of the gear's, it is 0.796.
- 1.8 without gear spell power is one wearer, and 1.6 without it is one weapon at rank 2, 3.97 a hit. The Crescent of
  Forlorn Spirits segment (rank 4, 6 spell power) reads 0.7804.

### Segments under 2.0 speed

Level is the one the judgements pick, where a rank's last level (16, 24, 32, 40, 48) stands for that level or higher.
Seals is the points in Improved Seals, and Hits and Crits list each amount with its count. A hit amount off the
segment's pair is listed and left out of its factor; the XZqD3fHLhnpT9RCG pair is 5 and 6. The factor of a segment with
gear spell power has 0.1 times it taken out. Fights are the ones that hold the segment's hits.

| Speed | Weapon                       | Rank | Level | Seals | Gear SP | Hits                               | Crits                            | Factor | Report and fights                  |
| ----- | ---------------------------- | ---- | ----- | ----- | ------- | ---------------------------------- | -------------------------------- | ------ | ---------------------------------- |
| 1.6   | Fist of the People's Militia | 2    | 16    | 0     | 0       | 3 (1), 4 (38)                      | 8 (3)                            | 0.7818 | rFfgx1DMjJLm8RVT/1                 |
| 1.6   | Fist of the People's Militia | 2    | 16    | 0     | 0       | 3 (1), 4 (30)                      | none                             | 0.7798 | XjAvxzwprm8MfkCV/1                 |
| 1.6   | Fist of the People's Militia | 2    | 16    | 0     | 0       | 3 (4), 4 (163)                     | 8 (8)                            | 0.7819 | fnWGQ1r82cZaxTDd/1-3               |
| 1.6   | Fist of the People's Militia | 2    | 16    | 0     | 0       | 3 (6), 4 (140)                     | 8 (6)                            | 0.7787 | kGh4q7b1F8XZmKPt/10                |
| 1.6   | Fist of the People's Militia | 2    | 15    | 0     | 0       | 3 (21), 4 (63)                     | 7 (2), 8 (1)                     | 0.7775 | 81DKPGmHxyjvgtfC/1                 |
| 1.6   | Fist of the People's Militia | 2    | 16    | 0     | 0       | 3 (1), 4 (46)                      | 8 (2)                            | 0.7823 | pJQ4mtrDqcaAzwBP/1                 |
| 1.6   | Crescent of Forlorn Spirits  | 4    | 32    | 0     | 6       | 9 (64), 10 (118), 13 (2)           | 18 (1), 19 (14), 20 (13), 21 (2) | 0.7804 | cm9xJBPRg7D8YjwX/1                 |
| 1.7   | Thief's Blade                | 3    | 23    | 0     | 5       | 6 (3), 7 (51)                      | 13 (4), 14 (3)                   | 0.8058 | 4AcZ2YNQtrvBCyWx/1,3               |
| 1.7   | Thief's Blade                | 4    | 27    | 0     | 22      | 9 (4), 10 (45)                     | 20 (5)                           | 0.7952 | R7fw48zhb9ACYtVq/1                 |
| 1.7   | Electrocutioner Leg          | 5    | 38    | 0     | 4       | 13 (344), 14 (34)                  | 26 (20), 27 (3)                  | 0.7999 | x7J3QTrjf1cXKk8v/1                 |
| 1.7   | Sang'thraze the Deflector    | 6    | 46    | 0     | 20      | 18 (14), 19 (33)                   | 37 (1), 38 (3)                   | 0.7897 | jJZHv8gGqrXz3AwY/1                 |
| 1.8   | Scalping Tomahawk            | 2    | 13    | 0     | 0       | 3 (16), 4 (174)                    | 7 (2), 8 (4)                     | 0.8138 | TqmjcRtWHr6dvCGk/1                 |
| 1.8   | Scalping Tomahawk            | 2    | 14    | 0     | 0       | 3 (2), 4 (128), 5 (24)             | 8 (3), 9 (1)                     | 0.8130 | TqmjcRtWHr6dvCGk/1-6               |
| 1.8   | Fist of Stone                | 6    | 48    | 0     | 14      | 19 (1), 20 (5), 21 (53)            | none                             | 0.8163 | Yv1Vp8KLkgMfWTRB/6,9               |
| 1.9   | Raider Shortsword            | 2    | 16    | 3     | 0       | 4 (1), 5 (24), 6 (21)              | 11 (1)                           | 0.8303 | 9MR2avp1gdQ7khLc/1                 |
| 1.9   | Raider Shortsword            | 2    | 16    | 3     | 0       | 4 (1), 5 (35), 6 (28)              | 10 (1), 11 (4)                   | 0.8258 | 9MR2avp1gdQ7khLc/1                 |
| 1.9   | Mace                         | 2    | 13    | 0     | 0       | 4 (60), 5 (26)                     | 8 (1), 9 (1), 10 (2)             | 0.8535 | CraQTxMWPvpFnf9D/1                 |
| 1.9   | Raider Shortsword            | 2    | 16    | 0     | 0       | 4 (1), 5 (225), 6 (7)              | 9 (1), 10 (34), 11 (3)           | 0.8324 | NvGyh8wgxAQfmZc9/1-2               |
| 1.9   | Raider Shortsword            | 2    | 14    | 0     | 0       | 4 (42), 5 (42)                     | 9 (1)                            | 0.8340 | B7Lj8xcA4dQvGN9h/2                 |
| 1.9   | Raider Shortsword            | 2    | 15    | 0     | 0       | 4 (13), 5 (62)                     | 7 (1), 10 (1)                    | 0.8389 | B7Lj8xcA4dQvGN9h/2                 |
| 1.9   | Mace                         | 2    | 16    | 0     | 0       | 5 (59), 6 (6)                      | 10 (4), 11 (2)                   | 0.8447 | pKY8tGdVP1ajFRrQ/1                 |
| 1.9   | Raider Shortsword            | 2    | 16    | 0     | 0       | 4 (1), 5 (42), 6 (1)               | 10 (1)                           | 0.8312 | XZqD3fHLhnpT9RCG/1                 |
| 1.9   | Raider Shortsword            | 2    | 16    | 0     | 0       | 5 (108), 6 (4)                     | 10 (3), 11 (1)                   | 0.8344 | DHcAgqKFT92yhBVL/1                 |
| 1.9   | Raider Shortsword            | 2    | 14    | 0     | 0       | 4 (53), 5 (43)                     | none                             | 0.8243 | qmMb9cyDCfVRj2B7/1-3               |
| 1.9   | Sword of Omen                | 5    | 40    | 0     | 0       | 15 (1), 16 (38)                    | 32 (1)                           | 0.8350 | GNc1DXLaZ3YKwbyd/4                 |
| 1.9   | Tainted Pierce               | 4    | 32    | 0     | 0       | 11 (26), 12 (27)                   | 23 (3)                           | 0.8332 | frGARNWFpvzMwkQx/1                 |
| 1.9   | Sacrificial Sword            | 6    | 46    | 0     | 69      | 26 (18), 27 (11)                   | 52 (2), 53 (1), 54 (2)           | 0.8241 | pKdBhnj8NgbZ9ay1/4,6               |
| 1.9   | Sacrificial Sword            | 5    | 38    | 0     | 62      | 21 (35)                            | 42 (5)                           | 0.8349 | vMrt6hfyXbdwaKzm/4,7,10            |
| 1.9   | Sacrificial Sword            | 5    | 40    | 0     | 48      | 20 (18), 21 (58)                   | 41 (7), 42 (5)                   | 0.8336 | xK63tMDhHNWaP8zQ/1                 |
| 1.9   | Sacrificial Sword            | 5    | 39    | 0     | 59      | 21 (93), 22 (23)                   | 42 (11), 43 (5)                  | 0.8296 | K2jLZwNMAvYyrm89/1                 |
| 1.9   | Sacrificial Sword            | 6    | 42    | 0     | 81      | 25 (65), 26 (18), 27 (1)           | none                             | 0.8326 | dv3cF6ygfawN2BL9/1                 |
| 1.9   | Sacrificial Sword            | 6    | 43    | 0     | 81      | 25 (93), 26 (158)                  | 50 (2), 51 (6), 52 (5)           | 0.8215 | dv3cF6ygfawN2BL9/1                 |
| 1.9   | Sacrificial Sword            | 5    | 40    | 0     | 58      | 21 (20), 22 (57)                   | 43 (1), 44 (2)                   | 0.8335 | P6FRN2hjatwky78A/1-3               |
| 1.9   | Toofbreaker                  | 6    | 44    | 0     | 9       | 19 (19), 20 (14), 21 (1)           | 39 (1)                           | 0.8378 | wgf3PCXdpbFzWxJ1/3-7               |
| 1.9   | Toofbreaker                  | 6    | 45    | 0     | 9       | 20 (67), 21 (1)                    | 40 (7), 43 (1)                   | 0.8363 | Mc16FCfkmzRQZ4TH/1                 |
| 1.9   | Sacrificial Sword            | 6    | 47    | 0     | 83      | 28 (17), 29 (39), 30 (1)           | 57 (2), 58 (1)                   | 0.8340 | aKQDHV4PLdmXJ7kc/1                 |
| 1.9   | Smite's Reaver               | 3    | 23    | 0     | 5       | 7 (5), 8 (51)                      | 16 (6)                           | 0.8370 | a4QkgPHBN7GLtdWn/1,4-5             |
| 1.9   | Sacrificial Sword            | 6    | 42    | 0     | 84      | 25 (38), 26 (53)                   | 51 (2)                           | 0.8356 | 9Apc8qL1WRF6Qn7x/1                 |
| 1.9   | Sacrificial Sword            | 6    | 48    | 0     | 60      | 12 (2), 26 (106), 27 (199), 28 (5) | 53 (7), 54 (18), 55 (1)          | 0.8198 | YwNT4VWfrH7n19LB/1,5-8,11-18,20-22 |
| 1.9   | Sword of Omen                | 6    | 46    | 0     | 4       | 20 (44), 21 (7)                    | 40 (1), 41 (1)                   | 0.8341 | NqvBaWc27tHKdjfr/1-3               |
| 1.9   | Sacrificial Sword            | 6    | 44    | 0     | 59      | 24 (252), 25 (102), 26 (2), 37 (1) | 48 (10), 49 (24)                 | 0.8319 | gTxCDJHRPB87G3tK/1-10              |

### Control reports

One-handers from 2.0 to 2.9 speed, code and fights, by report start:

1g7MmWLKZTqX6v3d/1,3, V7GL8HqnAh4wNxDj/1, fAyjMvn4C7X8VQgZ/1,3-9, ZxjKaRt3gCXD2FNv/1, Xr3DpGtJjbmCgPN7/1,
qbzcLNrh396pv4Mf/1, ghdqzjHfm6kBy19W/1-2,4-5, 9m3DwapGdVjAKxWf/1, At8nDdMZrXaHKThq/1, rTYFRZhJ1m2BLpdg/3,
hq9favY7kJFZxpA4/1, zTmjBrqGgwb1V9yJ/1, x3GLTHZwkr14bXmQ/1,3, 4ftmvrYpjc2qAFJC/1,3-5, 7CPbKgV1yXaxGdWJ/1,3-10,16,20,
WmKrZaxFfJ48hjAH/1, WgRf9QyPGCMcxLz6/1, zwnF3cCaLDvb4P91/1, jxWzPbpMhXnmgNJL/1, vGj9ZyzrdNmfwKTa/4, yAxhQ2bBmRG4ZYJa/1,
Nxm4BH6WzVwRQqc8/1, c3JZV9bnrfagjkpx/1, jdzv4JWCfZ29gD6h/1, dnbhjHBDWLk4MXcN/1-2, jbaLXmdkqD4RyZgr/1,
QZCgDaNyj7Rnb9rJ/1,3, w7MLDZhA3v6pYHnR/1, xFR918XdVmAGN4fw/1-8, J6xcKqmHWAhNDpMb/1-3, ZTY9pRtzv6QPLXwd/1,
VCRDbpy27dKLqvmN/1, ntwDbdXZPV9KNfL8/1,6-9, f9bkymNFnP3pVvZr/1,3-5, hJQVamtANHwb683f/1, N8tM3qaZPDJTWCFH/1,
kM3FYn6zWGqKgDaQ/1, H3KgGCTYjkP2LDz9/1-5, ya1bWvYhRKJ8kXnz/1, xjtHzrcP6bVT9agJ/1, qwT1DJXbQ68fxGyF/1,
CraQTxMWPvpFnf9D/1, DAGFNvnt72KYBMyT/1-2, aGr2xBY89AZjMHXy/1,3-13, aCcwGDtZgb1pnfyR/1, dJ6rA8ynvk27MY4Z/1,
GyHMLZTPpmDNd8WX/1, PwLATb8Zmx2KQ9dc/1, gpvQGXkLfCqxtaBD/7-10, h9jfxHpG2cynzCD8/1, qawZ7rBKXDJRzP4T/1,
BXT2xcHrNFGg7mjA/1, rbvt7kx8Pmn12MKJ/3,5, YmyQ8XgrBZf67p9K/1,3,5, 6mWtDdBpNG3JazCA/14,16-17,19,
KmhnPaYFzWNVCJ78/1-2,4-6, y9PbTgCF1tJ6RHDf/1, AxqKzPhRC17p64JV/1, X32HzqVx8ryNgAav/1-6, bLQm3FG4VtBWXgAk/1,
NvGyh8wgxAQfmZc9/1, yT7zXBZbjaFkmGYR/1, 71BtLG8JCdKXfMR2/1, zBxk3a4g1hmM68WD/1,3, cT6gCFpxKt4N2Rry/1,
68xHQtfwvpYDbkmJ/1, h4ZBDcntzPWbwXqF/1, NyBCk6jYGQb1dX9g/1-2, WZTY9tGPCpBDNKrQ/1, tqNcjBx6Yf2P7mpk/1,
znjX9Hbpd34WcDBt/1,3-4, DCTrdPA261fwmJcb/1-2, WqQch1Bf4b23dTXP/1, kZRFTpQ3gNqrPGvL/1,3-4, WmGY8ndQkKMNBFXH/1,3,
pwDB8NfHbq1txAcK/1,3-4, yz14D2NFQJh9BZ8K/1, LX3xj9YgN2zP8khp/1,3, Gh3jbtkYW2yNvgDf/1, jHXpwZ3hCRTNGt1F/6-8,
pTLcXfQ7YCqRDBG6/1, d2ZGhvbMmy8zqLxk/28,30, Jd9PKXgYaDZvLpn3/1,3-4, Y6wkNJAQV2rXDCHM/1, h29vrCNBAbJtazwH/2,
zQaCV4PH9ywGZbR8/11-12, QkDTnRA71rYCaF6c/1,3, PN1zfYMB8n64QHJG/5-6,11, JGTCD8nx1qz4aRFP/1, F1LmvGQqd4K36Nnk/1,3-4,
p2yJMDgFhGLrn4HW/4-6, QbFtmTrhcPC6paR3/1, MTga6qVmGwHxpcRP/1, HN2tbcr4AKpZGyL1/1,3, 1wzAmd3DFRQfrXhW/1,
Am14BxNVvC8dLaZy/1, 7Fk3DnJMBG6c1CVr/3, 9mnpaYr4FBL6P8vC/1, NyZF9PX4YRTbGAda/1,3-4, rHLRvwCWDZ68cpbn/3,5-7,
DQk6pfxL8jbN4Rht/1,3,7-8, fZ47Wb8RNJFBamv9/4, p31TP2rh4cWZJtjC/1,3, rkwG2zKFTY6ZMRd9/1,3-4,6-8,11, MnY7DprL3BfhkRXG/1,
cD8yjHJk9WdVPXgC/1-3, tKmNn2kGwbgyJr7h/1,3-4, dkM4FXnJ2q9pwWc1/1, yb7pCZHR3mA6QfrW/1, KvjxMyF3HwkbLVNm/1,3-7,
TMrPFXHzfdxJvB9w/1, 2CqjG6pJHWfmcA8Q/1,3-6, vPbKnx1RCpfXqV98/31-40, XZqD3fHLhnpT9RCG/1, Kn8t3zDANvhaFGgP/1,
wGRbacPYMq9f2AFL/1-2, TkQamZcJwb7v4dxg/1,3-4, TZYXnbNqW3tvrJRg/1, pTqJR6FtBaNAZVKm/1, 842pXDRAWPfwm6BH/1,
TcYJdbaLwWR9m7GN/1, 6gj9k4KGfNp8tPAM/1, C8AQFVd9fyqbDtJr/1,5, bxqTh6rNY37FcLZa/1, V1r3JtD7qnChXkR8/1,3-4,6-7,
CXrm1dagchbBYy63/1,3,5, WTDd8yJ2AL1pknj7/1,3, yT4bAw3vqn9PBWkj/1,4, HTbqLNPCcY6DWkJ8/1,3, RyDtHbTm2hcFCPBw/3,
kqTymHZfWPKv7Xpw/3,5-6, PV1JnpcyBKAh9j7C/1,3, GP6LFMrKAtpkzNXj/1, pfJ4aQ6ZK9NWbwGF/1, TP4GatMKF9XhrBHJ/4,
JvVn3Y1Nrt6TMXRH/1-2, mByK1Z8XhvrNgA6R/1-4, 3QKPzy9GNVcxHWDp/1, A3kzTC26prwf9jLQ/1,3, RLVMaqADHfKp8jrn/2-6,
TGqpwnN7vQPdWcbV/1, C9PygqXRzpAdjDxY/2,6-7,9-13, hDLzpJyZn9fPcCGV/1, Td4GCmfzFJckjZrN/1-2,4-5, rdHAcwYbmJ8DFWg4/6,
XGpzftr6C4WcmNMZ/1, gwLFGWXxN3yRcqCr/1, cCQA7vd1txHGMprK/1, r1ZP6fqnAbY49MD7/1-2,5-11, cLfGWqzYnr63By8h/1,7,
X2cRWHZh61vp9bxD/1, AM3d9HJBp7avXGCk/5, zrVMpcNkAxmtBP1w/1, FPRvaxdnqcLCQkt7/1-14, GLQnwDzWKA9hgYRT/1,
v41rYANj2zBWkb8n/1, v2gaAYKGq6th7Lk8/1, THqJCK3NG9XAvPDd/1,3-7, YWpTBLFRDMbdgCQJ/1,3-5, MtnCFxJz2jLk3NqP/1,3-5,
3fwmZYBhHFxQtLpM/1, KMNXJn3gfxPb6hcG/1,4-5, z4WwnT3jGKx1rL7M/1, 9yBj7VFxPv1ZRrDg/5, 1VzvJh3K6xgpG7wW/7,
NHQa9P38cwFRXjpT/1-2, GCfmMZbpkW2nATF1/1, HmkNx9gL7BzAdb6a/1, XLqQdpNwhCW8DV9B/3-8,10, VPy2jCATR1g3WfYx/1,5-6,
Tp6Df1ZGzhjdvQPK/1,3, 4gW1GvfjXCYNHtPV/1, 4gF2qTb6hXMvZD9n/1, yWj2cGPQtCRnwZp1/1, ZrM8cydqRt1BpLfx/1,4,
tnZKzQ3bJFTYdAq7/1, w8DFfTr1ZV6dQGYA/1, KW1RANMpzdBrkQtC/3, 39PRjkrgFM7NQYLa/1,9-10,12, hjwb6DB8RJHdyV7F/1,
MzBkDLFQvb62J3Wp/1, 1hL7gbTB8YFVz6m3/1,3, X4wjVGNfnLv9DHm8/1, TGBfm1rF2WR9Qdbk/1, FGPnDkzQJTtb38yj/1,
xDdjVFwHJ1M4nWyk/1, AcH6yD1Xj92BWfFk/1, GZzaJH1QVjMKd8qA/23, 6mq8yPzVKTLQFB7w/1,3-22, KfAmWF67gcjrMw2k/1,5,
TK6kxjyg1fvMLZXa/1,3-6, ZgDp46qWGy2nVXYP/1,3-9, VvfFMTn4pyKCBkcR/1, kzwxabP9jcRq1NpZ/1, 4rG2Nh6dAznLtymZ/1,
DKjFRTxMGrwLJaCz/1, CpWrY13VANwHg2Zd/12, ZkA3Vhz72n9Gy4dY/1-2, Cw2N8dXFbTDjHfka/2, HJ6MD7caWjv9VKmZ/4,
6HTcXknwxQmV714y/1, 2hQFjArHfLWYxgmq/1, kxWGHzr7cnFhRaNC/1, DzYQhtBq64naArmR/8, JnVckZDLTwAgGKaQ/5,9-16,19-26,28-33,
CaV3xMtYmfyGbr72/1-2, wgkXrVbdjRT78qz1/1, bqrzQtnH3LaVgyhj/1,4, VXLDK3P7WGrza41p/1, NLzDkMAj8n2TmbxV/1,
GbnV82tNm4aJzgQA/1, h94WFYKVxQRaJ7NX/1, kCPQ1TvqNtKrBFfW/1-2,4-5, fLWDJzmjdG8QNhag/1, wd3YJgy1VPHaTcCF/1,
hmp2aWfjVdNwtbyQ/5-8,11,13,15-16,18-23,25-26,28,30, gXNaFPB8cvDKtpY1/1, 2cW94Y617THKNVXD/2, PtZYhWG8Tw26jn3x/1,
QFGTWzAmkY6vf31d/1-3, 9RrQv7fabk4PhwL8/1, JdqCZQDwt1x2yk9X/5, hByjwtf1pNa8q2Ac/1, tLzGhTR6kH4cYraC/1,
KdBRYz9mGZDNtbxg/1, fNH2bq8aF6cnZB9G/1, btyR3qZFHp7dnL21/4, 14CFg32rqAt7bQYj/1-2, Ljr8XQmFWJ9n6q4K/1,3-5,
RKb6gxTJN1DypGCV/8-9, MVcaxDQwWLpnK6HJ/3, rVtLJA4kXNz917Kd/1, FKdXLngf3z7Z62A8/1-2, MBaXr1H9GcPhDJLm/1,
X8HTCw1btGAhZxnd/1, vTtbzncrBG6xqX78/1,3-4, 6yXzKjmxpWfG4B9A/1, dgLhwpKJ1ZMjvzWA/7, cYgfxhMQKZNRGaX9/17,19,
3HjYg9NhQLd6vzJ8/3-4,6, rZK2w6qxaLfApmQk/1-3, CzhNH72Fcj84qdV9/1, aJtpF3CH7LPTYnkd/1, ckPQ7GJhv8x3DKMz/1,4-5,
ycnkM9Lmdp3rQhH1/1, mphGqZBzCtcbv3LY/1, BCTwY72MhXNgQypK/1, 8XkGyRVPztv9ghnj/12,19, 7zVY4HtRK2APMcDr/192-196,
1RqYyDnKwTvxZCpF/1,3-14, d3J9RbTN16kLHcPv/1, ykA6bCZmwDN3n2hB/1, gGr4tp2XfLZWhPwA/1, tGR42J96gNdzj8XQ/12-14,23-24,29-30,
61JYnjgbFwmG7LrW/55, 6LGwgzTMHpmnh8Rj/1, Xwc1C36tZ2hAxHFW/1,3-4,6-9, XqhP28vV7kgMQanF/1, MvjyBAcwRkH47ZrD/1,
2zgQcv3AZRGm4MP7/3, KrL2mg4z91JywcaA/4, YWgHFy8LMKb2BRfZ/7, J4mG1bXwVRZYkAf3/5, qmAdgrfVK7JGBFT9/1,3,
CmQZdL2a7vfrbMh1/8, NACdtFBw1QH39Dfj/1, hYg1Jv7QpkVBD2wT/1, QYJkcmr4gt6bTfKB/4-5,7,9-10, MCbR1kLH4mYJNcGw/20,
CK2TFNJZ8G3HygLw/7, qYFCL1XmbaGvnfP3/1,4-6, mtG3Z7DJdx8YRqj4/1, ndfDtLVYvQkjA18W/1,4-5,7-12, tK8jr3kzpXCy6LQD/1,3-4,
yF2wVqRMbd7gjBDY/2-4,10, q3V4v8zcfaXw7rMW/18

Two-handers:

PdvwqZzcJj63HTfN/1, JZ7fmPHBhCrtMGjN/1,3-7, KXWAg3VhTBanFjGN/1,4-6, yTdDcYx4FwhjVWXQ/1, rgBzVZC9hTcFNmdR/1-8,
rf3JF4cCybBNZhwp/1-8, c8q4taVLnPpHwZF9/5, 6FpCtDyzqBV1dQNL/1, HBd9VjKrm8kXaMT7/1, KJmDj7pBZ3dx6QNX/4,6-8,
GWAYXcV7txQkvbm3/1,3-5, 43nr8TkFth1q9QcV/1, 7PvA2KYJpwcW8bnV/1, DQf9HnT6bzgX2kmL/1-2, MKzVgmNcbLR92Zfj/1-3,
dQqfmPcwk4avZ6Gg/1-2, K4gF8hZ2RDjzwMxd/1-2,5, QfNRpFVbHAJTanD8/7, XHCBmgj1wYRyarMQ/3,5-16, XdrGfD84JZtv6x32/1,3-12,
tCZPMzKD6r3J4yAF/1,3, JpcKtbM678Y3G2wN/1, cVXAG3BJ8vpmz7Tg/1-3, TmYJVyaDWM61PfvK/1, Xv9wKgTcM27NBbhR/1,
qftPhpYJa6dkvcF3/4, 8fH9Ga7cDPx6NRjv/3, P4DLvK8T1aqQ9z3d/1, wV2FBDX6mTPYxkvL/1-6, fQBpM9taR2bk86wY/1-3,
qjC7JVycv96phFwx/1,3-7, hKqBcZpLm2zF1D9f/1, jbaLXmdkqD4RyZgr/1, Dkj6ybpV3rLNJ2v9/1-6, N48QkYhFbH21WLPj/1,3-4,
YXNRgrGKyk37nBwv/1,3-6,9,11, VJKGBbCYvrwydtx4/1-6, 2xanm9DtFzh3TjHY/1,3-17, PbwtrgnFG2KMfzdp/1, f9bkymNFnP3pVvZr/1,3-5,
h79wjLNH4dDm8trF/1, 6kgZWzVYqv82aHnC/1,3-4, w3RfQyM9YHpDaKrq/1, dQRwHFpWBKcVb4hm/1,3, N8tM3qaZPDJTWCFH/1-4,
48yCmDWJMtvfKR16/5, qwT1DJXbQ68fxGyF/1, gHWwC7dPFG3Mkctb/2,4-7, 6AGY7QNhwPCdTmrj/1,4-5, L84Znj2h1GVrKC6x/1,
cym3rfDV1P7Mdp6N/1,3-8, 3KBQtXjRxvgh14am/1,3,5, Mram82BnN9hZHxFw/1-4, Tx42tQBFW9XYLvrp/1-6, afmcLPbwtxzR8Hkp/1-4,
46Zchqf7nGzkdbgm/4,8,10-12,14,16-17, Z7QgkAGq3Mhm1v2W/1, wmYXQFnWGJVARtpP/1,3-4,6-7, kx84KnTvVmcHq2RQ/1,
W78mqbAZNHLQkt1n/6, K6ca7Ch2xLtY8jFn/19,21-22, 8Yqy7ZcvQWKtXMpw/1, fGJxZb1WMn9gqtXP/1, dDynFxbtQRL1v2pV/1,3,
rbvt7kx8Pmn12MKJ/1, YmyQ8XgrBZf67p9K/6, pzZ6T7XdfnkvMayw/1,3, KmhnPaYFzWNVCJ78/1-2,4-6, g2z4bBvaAZr3NTxD/1,3-4,
p9BqMKXGkZJD4axz/1, 2MHXV8Y9jx43vhpk/10, whY72HpXM1NJqyfB/1-2, yBWKxPJFwZ1GH7gj/1,3-4, RcxGrBNFjWY4PM2t/1,
zG674q2Tn1gVLvMQ/3, FN9XwQvna4zq1hdm/1, tqNcjBx6Yf2P7mpk/1, yNQTZncd4Pj327fR/1,3-6, 3HnGyxm9pZXQC27k/8,
WYJ7AXa6f8cVg4Rz/1, cGnpQF4qZBtr9Nwm/1,3-9,11-12, H28y7zw9qKG6Xr3A/1, 4fNjzhLYRtgW3dx6/1, rdRHvhfj1AVB7wXY/1,
Nfhyt7Zb834aJznD/1-4, 2VhZXpbJyqFjWMCt/1,6, gFtDxh81c79fNdnb/7,9-12, WBX9adF24nhPK6VH/1,3-4, HNAz6LWFJkY9nR8b/1,3-5,
4nBkqCWR3cxVJN6d/1,3-8,10-11, Gh3jbtkYW2yNvgDf/1-2, gVhaNWzJP7fwrqc2/1-3, AqdMjcrfkt1n4ThK/1, A8w32bTJaLm1zYQq/1,
fnWGQ1r82cZaxTDd/1, pAmJ3T2DahWCrG4c/1, tJf1QbcmTdD2MB3N/30, 72rp1dcxJjzG8gbw/1,4-7, QkDTnRA71rYCaF6c/1,
JwTCAQ1tXPHV3fhv/1,3, BzZ1Qnmq9tAVD4vb/1, CaxqKjBYAt8rFfd3/1,3, WpQcwvMXaGKt87zA/1,3, pj8Zb7NY29V1hzgn/1,
fyawZPMRCLb1Bc96/1, 7mgzcHxd9WKv6DhC/1, GgMtdQWhCYJAbHD7/1-5, 2aApqvBW4Z7X3GnN/1,3, LnQvZV17KzpdrMhX/1,
YQm7WPrK8qMChn1T/4-5, gVmz1xC8JnRfNZ3v/1, ACZG6Y4QN2b1FvDW/1, XbW476TYdGfDnPJm/1-2, W7NVM2hcknF4z19Y/3-4,
KCwB2HDP6pkxjYZq/1, DNnkJ2tAZ6QKbrvY/1,5, HgrJvLqZdtWFTyM7/1,11, bq3XHZ2YJzGxhRLQ/1-4, fNmLVTGqBRyJ8ZD7/1,
19cWdBFDgMLTKAZh/1-3,6-7, F9hWPfavRVtr21Tk/1-2, QgWL8J2yXRtxAMn9/1,3-4, MQgJ13WZ64HahkDn/4,6-14, TMrPFXHzfdxJvB9w/1,
8mTQXA7JtwxFz4nM/1-5, 7gwjmDGCMrvy1bdT/2, b2cKjpaQAF8gC97G/1, bAFzkaPKHGrpW6R9/1, FcqT16H2hAnV4RQd/1-4,
WVwbLB3vHnazkKyF/1,4,7, JgfCQpFvrqDT7mhG/1, 8Tj1aADWcmy7knGx/1,3-8, 842pXDRAWPfwm6BH/1, 8a726RLBfN1mgDnK/3-10,
CfMAbx6nZ3WdyKP8/1-2, wzkW97dAvQ642ZGX/7-8, 7vRLXAbyhF2Jwn1D/1,3-9, vJFNVgCaKwxpdD34/1,3-4, mzgBX42aqYkvJtM1/3-5,
B9QMZrka86nVLpdK/1, TyXmzrQvcbBDRFZn/1-3, cbX7KRnyL6mYWrz3/1, 3GwdvLc1FNDK62CH/1, kqTymHZfWPKv7Xpw/3-6,
DvjCptYAc2G6XRgL/2,4-9,11-13, W6G9PJkpdvRtzb4F/1, K7TCvZVyk3wxpGac/1,4, wfkXMdcKVgpQntAB/3,5-11,13,15,
mT8p32HvLBJFM9Wx/1, qFXwjrNRyP1nzh3W/1, ad7TtCwYXkJRqncF/1, N38fyXHDtmALhKkd/1,8, LAVjnR7hqW31Trag/4,
7AmFgtHBhMnaZXDT/5,7-8, AM3d9HJBp7avXGCk/1,5, kM7HQKNW1mGJ8gtc/1, pwHfn1vmMF9yhXrk/1,3, BYPdWqwjJ4MmTkgK/1,3-11,13-14,
ZTARHXM9tWaFnhvj/1,3-4, 2PmTnx9C8Fvk4hNG/1, xh1JtjkqrFN27La4/1, kyZGYL7WHRDXCp13/1, nh4Q8zvM7tRaTA6C/1,
cVjGDxdPwaX2tZ9h/1,3,5-7, wfaZ2tmhTr3DWgx7/1-3, PrRzZg27FtK1Xbkn/1, XLqQdpNwhCW8DV9B/1, LfKFyjtVbAgq4QhW/7-8,10-12,
QRYjLtK4W1Pcd69n/12, TcrjQzwnAxaybRBC/1, Vc87mvFXRbLAq3hn/1, 1cY6t9Gvx8pJ7PMk/1-2, YbXPxwZ1pzA4Lk6n/2,
1AfDWYQT7Xjd24K3/4-8, 6wZatNDXYzcpVqKH/1, pqr4HxC3vWJhfG7B/11, nFJtd2mQgT7MBqCK/1,4-9, yAfxnFj8N4VLp9Qv/1,
47XpTywZYckfFHjP/1-2,4-9, VPB6pC9HQjJMF8hT/1, t3pKm4JTjk72bHVn/1, JQ3jpdcyzqg28m1t/7, rvgYV9KxyX4kHatd/1,
b6L3w27WfqkJgQVT/1-3,6, C9MyxFVtzcvjLgnH/5, 82YfmdLnzpAyHWwN/1,3-9,11-15, kxWGHzr7cnFhRaNC/1, hkgf3RCn94ABDFHb/1-3,
atcB91jgTqWpXy62/1,3, z8CJWHpN1nc4Lw6j/1-4, qGKZkhvcpXwYDR7f/1,6, phMXPfRgx817F2Dt/1,3-6, M8cCgJWHLTFmR49b/1,3,
PtZYhWG8Tw26jn3x/1, RKDTjJY4mNc9wbVk/1, Yj6qDR7nZ2tw1dTJ/1, qjxtbfXzW3A972Tg/1-5, fhP8gbn9DcVMG3dk/5-13,
87pJCtvLDhjMWVFG/6-7,10-13, 9WJHzhV8gLmGnZ6c/1, vYjL3q712CyTGkQh/41-42,44, 8Yn14jB2Hzm7QqWx/7, N7MZYmQFVbJBL1Dn/1,3-5,
dG9bfvPqpZnmBH6M/1, pa47tkW1CfzFc6yr/8, Fnj3HfRdh8p6VkMQ/2,4-5, L8wzRftW72FVvApB/1, x6RFYtdQNWJZc3hp/20,22,
hPzJfkLpX8RC4x71/1, nJRFNLWy4rX91Htk/1, 1dJQcXNTj8BnrkgK/3, Z6JfkdCMXpj1aRmb/1, 2zgQcv3AZRGm4MP7/3,7-8,
TD74HhX9zAxy8gKc/1,3,5-8, aGXdrkgJy8QFK2jD/6, BgaAM267rGqQcFYv/1-4,6, ajnqZH3tcXfwDyGJ/4, yF6JbgcMXBxkZVYL/6,8-14,
DJhAX3MHdrfwNg8G/1, qYFCL1XmbaGvnfP3/1,3-6, YnaMQc3zqw8bPNxy/1, 2RTQkMVPrZJGj4Xb/1,8,10-13,15-18, Rh8bJQwLvfxXNj32/14

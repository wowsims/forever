# EV-0177: Proc rates of five weapons and Red Whelp Gloves from the 2026-11-22 sweep

- Claim: Hookfang Shanker, Hurd Smasher, Shimmering Platinum Warhammer, Siegebreaker's Blaster and The Lobotomizer each
  proc at the rate its section gives, priced on the weapon base speed of each hit that rolls it; every rate is observed.
  Red Whelp Gloves cast Firebolt on 10% of landed melee swings and damaging melee specials, the chance its row states.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each item's section lists, from
  every wearer stream on disk for the six items; 51 reports in all
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 21 reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is 1.60.1.71037, push
  114326; the hotfix each player ran is not recorded
- Date: 2026-11-09 to 2026-11-22
- Pull request: #1168

## Method

The count is EV-0153's Method, run again by the automated sweep of 2026-11-22 over every wearer stream on disk, with the
same rolling hits, classes and forms left out, intervals, lone-unit checks, swap drop, credit windows and calls of
measured or observed. From hotfix push 114313 (2026-11-20 23:47 UTC) each Mutilate strike rolls its own hand once; before
it Mutilate is left out with its procs, as the two-roll abilities are. No two counted runs of one player overlap in time.

The rate is the all-hits rate, rounded to the nearest 0.05 from the full-precision read. A one-hand white rate (lone
Auto Shots on a ranged weapon) or a specials rate of 50 or more procs checks it where their intervals overlap; a
dual-wield white rate neither sets nor checks a rate. The run's posted `ppm` is the lone white rate wherever that has 50
procs from 3 wearers, so the rates here come from each result's `all_hits`, matched on item and `effect_spell`.

The run counts no Shank as a roll. From 2026-11-20 Shank rolls the off-hand weapon (EV-0155 §Shank from 2026-11-20), so
The Lobotomizer's section counts the off-hand wearers' Shanks from that day as rolls.

Scripts and inputs, in `forever-sweep-auto`: `analyze.py` (version 12 with the paladin check and the Mutilate split at
push 114313, cache `cache_v13`) for the rates, and the run `runs/2026-11-22-1122`, whose `results.json` holds each
weapon's counts.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon                        |   Item | Rate | Call     | All hits [95%]    |  Procs / hits | Wearers | One-hand white              | Specials  |
| ----------------------------- | -----: | ---: | -------- | ----------------- | ------------: | ------: | --------------------------- | --------- |
| Hookfang Shanker              |  11635 |  2.4 | observed | 2.40 [2.25, 2.64] |  565 / 10,079 |       5 | -                           | -         |
| Hurd Smasher                  |  13198 | 2.15 | observed | 2.14 [1.98, 2.28] |   265 / 4,122 |       7 | -                           | -         |
| Shimmering Platinum Warhammer |  15418 | 1.85 | observed | 1.85 [1.74, 1.90] |   487 / 5,102 |       3 | 1.59 (101)                  | -         |
| Siegebreaker's Blaster        | 282561 |  5.3 | observed | 5.29 [4.95, 5.67] | 1,013 / 4,423 |       5 | 4.92 (371), lone Auto Shots | 4.80 (42) |
| The Lobotomizer               |  19324 | 0.55 | observed | 0.53 [0.49, 0.57] |  166 / 10,503 |       5 | -                           | -         |

## Hookfang Shanker (item 11635, spell 13526): 2.4, observed

565 procs on 10,079 rolling hits, 5 wearers, 8 runs, in the off hand: all hits 2.40 [2.25, 2.64]. The proc's flight
reaches the 300 ms isolation, so no lone-unit rate is measured and none checks the all-hits rate. Observed: no lone-unit
rate checks it; every wearer is a rogue. Travel time: the proc lands a median 99 ms after its hit, 98th percentile 284
ms (101 isolated procs). Dropping the swapped stretches (35 active minutes, 7% of hits, in 2 runs) moved the read from
2.25 to 2.40. Reports from 2026-11-19 to 2026-11-22.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |       5 |   565 | 10,079 | 2.40 |

Reports and fights: 6TNJgAGhynFQqw2a 1-2; 74PAn1mMkWwYGcDd 1-44; CxbgQvRAL61MZdc4 1-4; JHF4zftRn3PaBbN1 1-6;
JPLDVkNF7Y43nyfp 1; LQfjhwmGWga9XBHZ 1-4; cPGDLT3wyWYAxZkt 1.

## Hurd Smasher (item 13198, spell 17308): 2.15, observed

265 procs on 4,122 rolling hits, 7 wearers, 7 runs, in the off hand: all hits 2.14 [1.98, 2.28]. Neither the one-hand
white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no lone-unit rate
checks it. One-hand white: none. Dual-wield white, which sets no rate: 1.74 [1.36, 2.03], 109 procs on 2,087 lone units,
7 wearers. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 61 isolated hits. Dropping the swapped
stretches (27 active minutes, 15% of hits, in 2 runs) moved the read from 1.82 to 2.14. Reports from 2026-11-16 to
2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       6 |   249 | 3,811 | 2.18 |
| Warrior |       1 |    16 |   310 | 1.72 |

Reports and fights: 1TFWH4Yy38KhXMwx 6; 2ZLD7Jvwfa1n3xyM 1-2; ACJhv8TgVcbpRM3f 1-3; Tf1YmVratWLwcjkp 3-4;
VqfFdmTwMzxrcH1N 1-2; daTq1vcjC3QBntb8 4-18, 20-27, 29-42; qZTd9F3pkMLJcA6V 1-3.

## Shimmering Platinum Warhammer (item 15418, spell 19874): 1.85, observed

487 procs on 5,102 rolling hits, 3 wearers, 4 runs, in the main hand: all hits 1.85 [1.74, 1.90]. The one-hand white
swings read below it and their interval does not overlap the all-hits one, so under EV-0153's method the rate would stay
open; the owner takes the all-hits rate as observed (2026-11-22). One-hand white: 1.59 [1.26, 1.72], 101 procs on 1,226
lone units, 3 wearers. Travel time: the proc lands a median 106 ms after its hit, 98th percentile 234 ms, window [24,
244] ms (42 isolated hits). Left out: 1 of 5 runs in Cat or Bear Form. Reports from 2026-11-18 to 2026-11-21.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       3 |   487 | 5,102 | 1.85 |

Reports and fights: DvrjQXqZ9HN8a74z 1-4; KNtHCAPd97ZMTLBf 1-2; jWfDCghpYJKy9G3N 1-8; ycMCt3n9ZqP2bhaH 1-2, 4-8.

## Siegebreaker's Blaster (item 282561, spell 1314040): 5.3, observed

1,013 procs on 4,423 rolling hits, 5 wearers, 6 runs, in the ranged slot: all hits 5.29 [4.95, 5.67]. The rate is all
hits, which the lone Auto Shot rate agrees with. Observed: every wearer is a hunter. Lone Auto Shots: 4.92 [4.59, 5.93],
371 procs on 1,739 lone units, 5 wearers. Specials: 4.80 [3.37, 5.53], 42 procs on 202 lone units, 5 wearers. Travel
time: the proc lands a median 70 ms after its hit, 98th percentile 181 ms, window [22, 191] ms (411 isolated hits).
Dropping the swapped stretches (21 active minutes, 12% of hits, in 2 runs) moved the read from 4.64 to 5.29. Reports
from 2026-11-09 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Hunter |       5 | 1,013 | 4,423 | 5.29 |

Reports and fights: 1LK3bmTvrNXhWcwk 1-20; 8ZBqNXCjPa7hKWdG 2; d1pRGTnZVvcfa7qK 1-8; pGBPWgMrL9KkQZRm 1-2;
pQ32h46zWCF9wZGj 1-2; prCfHbFPmDAgqd3V 3-4, 6, 8.

## The Lobotomizer (item 19324, spell 1290950): 0.55, observed

166 procs on 10,503 rolling hits, 5 wearers, 8 runs, in the main hand or the off hand: all hits 0.53 [0.49, 0.57]. The
off-hand wearers' Shanks from 2026-11-20 count as rolls (EV-0155 §Shank from 2026-11-20): the run's `results.json` counts
the procs that follow a Shank but not the Shank itself, 166 procs on 9,537 hits, 0.58 [0.52, 0.62]; the 966 Shanks
added here bring it to 0.53. The off-hand runs read 0.53, 100 procs on 6,318 hits, and the main-hand runs 0.53, 66
procs on 4,185. Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the
all-hits rate. Observed: no lone-unit rate checks it; every wearer is a rogue. One-hand white: none. Dual-wield white,
which sets no rate: 0.41 [0.27, 0.60], 38 procs on 3,118 lone units, 5 wearers. Direct: window [-11, 100] ms, median 0
ms, 98th percentile 0 ms, 31 isolated hits. Reports from 2026-11-19 to 2026-11-22.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |       5 |   166 | 10,503 | 0.53 |

Reports and fights: 2x7ZjvdbRwFMWKTP 3-45; 4rTfkGxMwKWPAXng 1, 3; GD26JWT9QRqhCyFg 3-36; GXR7nt6ZCL4DQzxP 1;
Gg9W1x7TFHNLyVck 1; WqfyDhVHxT93Q7Pt 25-30; ZfXrLy2MdY6kzah8 2-7; nrpTgLQdcNKAqk1R 1.

The count reruns the analyzer's `measure` on the run's cached records (`cache_v13`) with each off-hand wearer's Shanks
from 2026-11-20 00:00 UTC added as off-hand rolls. Every Shank in these runs falls before 2026-11-19 22:03 UTC or after
2026-11-21 05:00 UTC, so the cut moves no Shank.

## Red Whelp Gloves (item 7284, spell 9233): 10%, the row's chance

The row 9233 states a 10% chance on landed melee swings and melee specials of either hand, and casts Firebolt 9057;
the tooltip reads 5%. A lone unit is a landed hit with no other event of the wearer within 100 ms before or after it,
and it has a proc when Firebolt logs from 6 to 100 ms after it. The intervals are per-hit Wilson intervals.

| Lone units                     | With a Firebolt / units | Share [95%]         | Wearers |
| ------------------------------ | ----------------------: | ------------------- | ------: |
| White swings, one hand         |             207 / 2,226 | 9.30% [8.16, 10.58] |       9 |
| White swings, dual wield       |             598 / 6,259 | 9.55% [8.85, 10.31] |       8 |
| Damaging melee specials        |             225 / 2,399 | 9.38% [8.28, 10.61] |      13 |
| All three                      |          1,030 / 10,884 | 9.46% [8.93, 10.03] |      16 |
| Sunder Armor and Rend landings |                 0 / 442 | 0.00% [0.00, 0.86]  |       4 |
| Ranged hits                    |               0 / 2,218 | 0.00% [0.00, 0.17]  |       8 |

Left out: 7 of 27 runs in Cat or Bear Form, whose paw swings read 9.79% (575 / 5,874). The wearers are 5
rogues, 4 hunters, 4 warriors and 3 shamans. Reports from 2026-11-15 to 2026-11-22.

Reports and fights: 192ZN3dKFymLACY8 1; 1B6kfNZzCb4GqpFH 1-2; 4ChPnBrVK2fJLMaX 4; 8Xxb9LDGHP7vcJfh 1-5; KBzML9kdHFCXVpQc
1-2; Mj1gA4rpDRQh9yJm 1-3; QD8ypCXWJcv9Y2B3 6-11; ZGtjkf7naDdXLhbP 1-2; dP4ahnBM2fx8t3Zk 1-6; dVYAgntWbDqwPJvr 1-9;
k6Jrj8BbDqG1apW3 1-2; mzTCH7rgQnB6jqkV 1-10; n23GhKYqmPdwyg7A 1-5; qNyQvka41YhVmdHR 1-4; r2BYfmZW1wkLDdNx 1-2;
t2jHDfcvLkK96bdV 1-9; tcBj1NqHbxM7AaZp 1-4; wPrMvF2qYa61TCGB 1; y4fHqpA8Qt2m3Vaj 1.

The count reads the run's cached records (`cache_v13`), paladins left out.

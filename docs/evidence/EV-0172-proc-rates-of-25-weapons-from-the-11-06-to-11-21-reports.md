# EV-0172: Proc rates of 25 weapons from the 2026-11-06 to 2026-11-21 reports

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it; all 25 rates are measured.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the 25 weapons; 753 reports in all
- Game build: 1.60.1.70790 to 1.60.1.71037 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC,
  71037 from 2026-11-20 18:23 UTC); 6 reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database
  is 1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-21
- Pull request: #TBD

This replaces the rates EV-0136 gives for 20 of these weapons, EV-0153's rates of Bloodrazor, Sul'thraze the Lasher and
Wraith Scythe, EV-0126's Argent Avenger rate and EV-0132's Brain Hacker rate.

## Method

The count is EV-0153's Method on the hits EV-0155 finds rolling a weapon proc, run again over every wearer stream on
disk for these weapons, with more wearers fetched on 2026-11-21. The same rolling hits, classes and forms are left out:
paladins from every rate, druids in Cat or Bear Form, and the two-roll abilities with the procs around them. The
intervals, lone-unit checks, swap drop and credit windows are EV-0153's.

A rate is measured where a lone-unit rate of 50 or more procs (lone one-hand white swings or specials, or lone wand
shots for Firebelcher and Mana Channeling Wand) overlaps the all-hits interval, and 5 or more wearers of more than one
class give an all-hits interval within 10% of the rate either side. The rate is the all-hits rate rounded to the nearest
0.05. A lone-unit interval from under 5 wearers is a Poisson interval. On Brain Hacker, Kang the Decapitator, Sul'thraze
the Lasher, Supercharger Battle Axe and Wraith Scythe every class past the first is one wearer.

Each section's read by report date is the all-hits rate over the reports that start before 2026-11-18, from then to
push 114313, and after it.

Scripts and inputs, in `forever-notices-1121`: `results.json`, the counts of `forever-sweep-auto/analyze.py` (version
12, as EV-0157's, with the paladin check that sets no rate) run with `--items items.txt --census census_used.json`;
`report.py`, the call per weapon (`rep_final.json`); and `dates.py`, the report dates (`dates_final.json`).

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs;
for the two wands the one-hand white column holds lone wand shots. Was is the row's rate before this count, which every
row held as observed.

| Weapon                      |   Item | Rate | All hits [95%]    |    Procs / hits | Wearers | One-hand white | Specials   |  Was |
| --------------------------- | -----: | ---: | ----------------- | --------------: | ------: | -------------- | ---------- | ---: |
| Argent Avenger              |  13246 | 1.05 | 1.06 [1.02, 1.09] | 5,000 / 128,616 |      62 | 1.03 (615)     | 1.07 (432) |    1 |
| Blade of the Wretched       |  10803 |  1.9 | 1.92 [1.81, 2.04] |  1,840 / 27,357 |      19 | 1.58 (122)     | 2.10 (223) | 1.95 |
| Bloodrazor                  |    809 |    1 | 0.98 [0.93, 1.02] |  2,309 / 52,506 |      30 | 1.05 (167)     | 0.88 (263) |  0.9 |
| Bloodspiller                |   7753 | 0.95 | 0.94 [0.86, 1.03] |     385 / 9,076 |      15 | 0.95 (116)     | 0.99 (25)  | 0.95 |
| Boneslicer                  | 273819 |    3 | 3.02 [2.81, 3.23] |     447 / 4,438 |      25 | 3.08 (238)     | 3.33 (93)  |  2.8 |
| Brain Hacker                |   1263 | 0.95 | 0.93 [0.90, 0.96] |    990 / 18,309 |       9 | 0.95 (251)     | 8.57 (1)   |    1 |
| Cookie's Tenderizer         |   5197 | 1.95 | 1.95 [1.81, 2.08] |  1,464 / 16,079 |      36 | 1.98 (277)     | 1.93 (288) |  1.9 |
| The Cruel Hand of Timmy     |  13401 |  2.2 | 2.20 [2.06, 2.37] |  1,200 / 18,223 |      18 | 2.27 (124)     | 1.73 (32)  |  2.1 |
| Eater of the Dead           |  10805 | 1.75 | 1.74 [1.64, 1.83] |  1,865 / 29,229 |      26 | 1.86 (348)     | 1.75 (206) | 1.65 |
| Firebelcher                 |   5243 | 3.95 | 3.93 [3.74, 4.14] |  1,416 / 12,725 |      64 | 4.01 (1,257)   | -          | 3.85 |
| Hammer of the Northern Wind |    810 |  4.5 | 4.52 [4.40, 4.68] |  4,649 / 29,390 |      12 | 4.46 (366)     | 4.17 (338) | 4.55 |
| Kang the Decapitator        |   2291 | 0.95 | 0.95 [0.90, 1.01] |  1,200 / 21,014 |      17 | 0.88 (274)     | 0.68 (6)   |    1 |
| Mana Channeling Wand        |  18483 |    4 | 4.02 [3.76, 4.29] |   1,030 / 9,616 |      38 | 3.81 (840)     | -          | 4.25 |
| Pyromancer's Baton          | 286650 | 1.95 | 1.94 [1.82, 2.07] |    964 / 12,402 |      16 | 1.85 (98)      | 2.17 (114) | 2.05 |
| Scorpion Sting              |   1265 | 0.95 | 0.97 [0.93, 1.01] |  2,096 / 54,256 |      59 | 1.03 (144)     | 0.99 (306) | 0.95 |
| Searing Blade               |  12992 |  1.9 | 1.90 [1.78, 2.06] |     392 / 4,944 |      17 | 1.89 (169)     | 1.88 (60)  |  2.4 |
| Serpent Slicer              |  13035 |    1 | 1.02 [0.94, 1.08] |  1,105 / 26,110 |      24 | 1.23 (26)      | 1.05 (139) |  1.1 |
| Sliverblade                 |   5756 |    2 | 2.01 [1.93, 2.13] |  1,267 / 26,995 |      21 | 2.39 (30)      | 1.75 (62)  | 2.25 |
| Sul'thraze the Lasher       |   9372 | 1.45 | 1.45 [1.32, 1.56] |     625 / 9,926 |       8 | 1.28 (126)     | 1.72 (49)  |  1.4 |
| Supercharger Battle Axe     |   9486 |  1.9 | 1.89 [1.72, 2.00] |     258 / 2,923 |       9 | 2.08 (68)      | 3.21 (3)   | 1.95 |
| Sword of Corruption         |  13032 |    1 | 1.02 [0.96, 1.07] |  1,249 / 35,139 |      59 | 1.07 (144)     | 0.99 (123) |    1 |
| Twilight Forgehammer        | 281665 | 2.95 | 2.95 [2.84, 3.07] |  2,280 / 18,539 |      13 | 3.08 (311)     | 3.20 (290) | 3.15 |
| Wraith Scythe               |  11920 | 3.45 | 3.45 [3.17, 3.73] |   1,136 / 8,989 |       8 | 3.09 (218)     | 3.81 (141) | 3.35 |
| X'caliboar                  |  10758 |  1.8 | 1.81 [1.73, 1.97] |     715 / 7,174 |       8 | 1.90 (113)     | 1.61 (48)  |  2.1 |
| The Ziggler                 |   8006 |  3.6 | 3.61 [3.48, 3.78] |  3,916 / 38,305 |      49 | -              | 3.91 (147) |  3.6 |

## Argent Avenger (item 13246, spell 17352): 1.05, measured

5,000 procs on 128,616 rolling hits, 62 wearers, 92 runs, in the main hand or the off hand: all hits 1.06 [1.02, 1.09].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.03 [0.92, 1.13],
615 procs on 16,253 lone units, 13 wearers. Dual-wield white: 0.92 [0.85, 0.98], 744 procs on 22,074 lone units, 51
wearers. Specials: 1.07 [0.97, 1.14], 432 procs on 11,057 lone units, 33 wearers. Two-roll abilities, priced as one
roll: 1.78, 19 procs on 291 lone units. Direct: window [-20, 100] ms, median -1 ms, 98th percentile 0 ms, 1,215 isolated
hits. Dropping the swapped stretches (138 active minutes, 3% of hits, in 14 runs) moved the read from 1.03 to 1.06. By
report date: before 2026-11-18 1.08 (1,354 procs); from 2026-11-18 to push 114313 1.05 (3,646 procs). Reports from
2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      31 | 1,430 | 37,406 | 1.04 |
| Warrior |      23 | 3,459 | 88,691 | 1.06 |
| Hunter  |       8 |   111 |  2,519 | 1.20 |

Reports and fights: 1pJHLqQw297v3aF4 1-2; 1z4vRDqGdM2cVPXk 1-4; 38kdWHmLZKVgTPqb 1-2; 3HaPg2Yktxrnb7NL 1-9;
3qR2PmyNvQk8J4Bc 16-31, 33-55; 4av7FfKdZjBh9Yp6 1-7; 4hHBvakRnT7p1GVK 6-7; 4VwxHWJ3pgNmvQY2 1-3; 649MAP3BJxaLfj7K 1;
6CbKfDBm78nhyLjJ 1; 6hDJQWTXwBr419AM 1-4; 6kD7wbm3NaKW8RYP 1-7; 6mD4XcqZR2YnraGB 1; 6mpWRxtgXn8JhfkT 3; 7fy93LdcwQhvF2gP
8-9; 7fz9Kk8rY6jyQAbL 3-10; 8FBqfvLyWTdKxbZh 1-38; 8GFRaPBnKWf7Qrpb 1-3, 5-6; 8QHYRK6xTXB2gAtd 1; 8yPRmZdrq4QhcGn1 1-4;
93QNL64MnXKkxhvF 1-5; 9zxHaWbYT3gqDQvF 17-18; aANdrQRjKP7MJnCY 1-10; AvXNG1TpnyLJMxVt 1; B1qcTGpnHCrQ64Xa 1-4;
BHJMb9CGa1TgjtAd 1; BHvgJVTdZ9pfNPa6 1-3; bNZvk628gXLz7t3c 1; bQw4YVFNJZBt2THG 1; CgXG9bvFH6jJVZKk 1-4, 6-9, 11-15;
DCjMyH4x2m8Qaz3b 1-5; dzgfcTtLA1v7jDZ9 15; FN8krKfCb3BXWmwV 1-8; FrL4wayTHRtW1CKx 2-15; gAMnRkmdrKWzLJ4C 1-3;
gdBnpzCGaJ2THt3y 1; GM97YvP1hdgx8akn 1-2; gVHqdJ1Xwf3kBny9 1; GWMx4hcJFCYHR6dk 1; HfaAyZNt6rKWV7pY 1-2; Ht1CryL9kJDKFabv
1; JbQyngR8NWXKardc 3-4, 6-9, 11-16, 18-24, 27-44, 46-47, 49-60, 62-63; JctLn6NBFx2rwGjv 1-6, 8-11; JFv64qNc8yfWY9Qh
1-3; JKYV6CpmP2GA1f7Z 1-3, 5-13; jQGkxMF61cvmaLdB 1-4; K6VvbDypXFLQJW7d 1-27, 29; KAv7LjkFHTwpZbD3 1-2; kjbQpMhK81LqRgyB
1-3; kvCd8J41RrZ3zXNp 1, 3; LF9b7JDR36p284Hv 1; LQ9zMDqpmknBgAX8 1-2; m69XVCZpd7Dn8AyH 1-2, 4-8; m9rN6C1Z7RbWfx2K 1, 3;
mPTyvJ6XYKCMAdck 1; mW2TtC9YqHZz3vpn 1-2; MXHn319aGYhkAw7L 1; N3nHF2GBgdT6qx1V 1, 4; N9HJ4LQVFDYkXR7j 1-3;
NK6cCM4pJyRhaAFW 1-7; p4RKTftgHvzLrwXm 1-3; pG8TNJwybLPYWmxF 2; pnGKFx1vL82VWPy3 1; q31yFhx7TgpPfzGW 1; qJX72LBYW1fAx6DP
1-3; qMgtfcB2vrRdyma4 1; Qr9PmCnptRWMjN7k 1-28; qRL8HagchkrTVJmD 1-39; RHgxPaQbVWjDTc4K 1-7; rTKYMd4yDvaL6Gn8 1, 3;
T8JKMZh1WwyazjHY 2; Tdva3pHk4tmCfjZK 1-2; tG4CqkJQvyPT2A1Y 1-10; TLp1jza6PCrBAX94 1-7; vG3BpgdzWnmMrcfa 1-3;
VMBZcafW2JLbT6dK 1-2; wh2p6j3rZfHMaYzv 1-17, 19; Wn8wHhaxy31MLzjp 6-8; wqahKP7xGMNbc8mH 1-6; x2QH6gyDAb9ZrR3q 1, 3-7;
XD23HFnxbzvPjACV 1; xkKRJPqCLzBa9GdH 61-66, 68-89, 91, 93-96, 98-101; Yf79qHLZAjgnTdJw 1, 3-4; Z4VAmgDfJ9GPFjtc 1-6;
Z6Xm92K3zgfrqM84 1-4, 6-12; Zk1Gh4WA7KJDaMf2 1-2; ZRhnvMANGJDy1kBj 1; Zvr84V1Y3Gpg6n9H 5-20.

## Blade of the Wretched (item 10803, spell 18088): 1.9, measured

1,840 procs on 27,357 rolling hits, 19 wearers, 19 runs, in the main hand: all hits 1.92 [1.81, 2.04]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.58 [1.31, 1.88], 122 procs on
2,210 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 1.82 [1.65, 2.03], 294 procs on
4,612 lone units, 14 wearers. Specials: 2.10 [1.85, 2.29], 223 procs on 3,039 lone units, 18 wearers. Two-roll
abilities, priced as one roll: 3.05, 21 procs on 197 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 461 isolated hits. Dropping the swapped stretches (87 active minutes, 10% of hits, in 3 runs) moved the
read from 1.74 to 1.92. The paladin check, which sets no rate: lone white swings 2.37, 20 procs on 241 lone units, 5
wearers. By report date: before 2026-11-18 1.90 (1,402 procs); from 2026-11-18 to push 114313 2.00 (438 procs). Reports
from 2026-11-12 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 |   847 | 13,708 | 1.77 |
| Rogue   |       8 |   745 |  9,943 | 2.14 |
| Hunter  |       3 |   248 |  3,706 | 1.91 |

Reports and fights: 34gy17FXnaGhVC2M 2; 6kN78G2xfqy3vwTP 6; 7M3gGcH6bprxKqZY 1-3, 5-8; ATMmnV2tC3ZPb6G8 1, 3;
bGRfPhMWCrKg9ydV 2-3; c69WN31xbHBPtmhv 5; cbkDQfXNPGj3tnWT 1; dwcgGfFmAXWK7Mjz 1-2; hDGL1wmpV4WjxbHP 1-3;
jJYHChwVBRaDxNWk 16-32; kz4FVdBRqpJwj8nx 1-3; mApW49VaFD2TKyg1 1-2; MHL1xhpwC4dcVDZK 1; N2m8nPTGJQrj1DhH 4-5;
Ncp4bVY3x1DXRqvA 1; VPG6FyAdJ7NcC29X 1-5; VwyW3ZP1H6RTpdmf 1-5; ZD3cfJpHa4vGYjKF 1; ZfRYBTWb168N7mzQ 2.

## Bloodrazor (item 809, spell 17504): 1, measured

2,309 procs on 52,506 rolling hits, 30 wearers, 41 runs, in the main hand: all hits 0.98 [0.93, 1.02]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. EV-0153 left this rate open, its one-hand white
swings reading above all hits and its specials below; on these reports both lone-unit intervals overlap the all-hits
interval. One-hand white: 1.05 [0.81, 1.15], 167 procs on 3,530 lone units, 5 wearers. Dual-wield white: 0.60 [0.45,
0.73], 138 procs on 5,114 lone units, 25 wearers. Specials: 0.88 [0.79, 0.98], 263 procs on 6,651 lone units, 30
wearers. Two-roll abilities, priced as one roll: 2.83, 14 procs on 110 lone units. Direct: window [-11, 100] ms, median
0 ms, 98th percentile 1 ms, 403 isolated hits. Dropping the swapped stretches (6 active minutes, 0% of hits, in 1 run)
moved the read from 0.97 to 0.98. By report date: before 2026-11-18 1.01 (843 procs); from 2026-11-18 to push 114313
0.96 (1,466 procs). Reports from 2026-11-13 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      19 | 1,392 | 31,593 | 0.98 |
| Warrior |      11 |   917 | 20,913 | 0.97 |

Reports and fights: 3XCdM1LQRNmvY8kp 1-8, 10-12; 76zv93gbJQFrWKfd 1-30; 7mFKVrW6QLHy3fgk 1; 8Cx9vcnzpJwk34WA 1-11;
8GFRaPBnKWf7Qrpb 1-3, 5-6; a4P93RWJwdb1c2QY 1; aNh3Dt8kbZHL9cBM 1; AZchwX93DmyHLbv6 1-3; bGjk9YQT8DgdKFPf 1;
byzq9hxPvpDAJac2 1-12; d1AC7qZ8h2tNnFLK 1-2; DcVtKxdCXMga9QZf 1; FvqNn8XAhHD6M4pY 1-10, 13; hYGRt9H1NXJWMZbB 1;
j1xyYV7gcMH69ND8 1-20; jJWt7Dnm4qc9rzQd 1; jr7TAd93qLPxWzhp 1-2; k8mdVabt2CFphzJ7 1; KcN4tqGDJQxmgrvZ 1-5;
kJ3WdhMLqv7K2Nyf 1-3; MXZry8VfkWHhvnYc 3, 5-12, 14-19; NjgtYv4VMR26q7h9 1-4; NWvgFp3kfR2bHyYa 1-35; nzxZ4h87yGXJPgfp 1;
p34dnvkg7yDKL6cr 1-3; PmybzCa2j7NXVxng 1-2; Q8Vgbpw4dJrzNx29 1-4; qVnZ6bPMhWRJA397 1-3; RW1BNk6Ybyh3XjLQ 1-21;
tk3BD1bZafMgKxWX 1-4; vBWX7HQn9zpTMGd8 13-22; vyQrDZN7RX82ApnP 1-3; W1P3YqBptbnGM7da 1, 3-5; w4fzHNVk2dAhLKnv 3;
WAyrNRJdm3pqCLMK 1-13; wFBytQxvWKz1Xf2L 1-5; Wj3TyNxChH4nvpLD 1-2; xkN3nXgd8RpHybzT 1-4; yPY3Nhg9MB2TVGZm 1-3;
ZA8pM9gmXcG6avCT 1-3, 5-6, 8; ZLpcV6krn1wYHJvt 1-8.

## Bloodspiller (item 7753, spell 18200): 0.95, measured

385 procs on 9,076 rolling hits, 15 wearers, 17 runs, in the main hand: all hits 0.94 [0.86, 1.03]. The rate is all
hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 0.95 [0.82, 1.11],
116 procs on 2,725 lone units, 14 wearers. Specials: 0.99 [0.93, 1.68], 25 procs on 561 lone units, 5 wearers. Two-roll
abilities, priced as one roll: 0.82, 3 procs on 81 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th
percentile 0 ms, 132 isolated hits. Dropping the swapped stretches (15 active minutes, 4% of hits, in 2 runs) moved the
read from 0.91 to 0.94. The paladin check, which sets no rate: lone white swings 1.22, 23 procs on 420 lone units, 6
wearers. By report date: before 2026-11-18 0.91 (282 procs); from 2026-11-18 to push 114313 1.05 (103 procs). Reports
from 2026-11-14 to 2026-11-19. The sim does not simulate this proc yet (the generator lists it as a missing effect); the
rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      11 |   295 | 6,541 | 1.00 |
| Warrior |       3 |    41 | 1,295 | 0.70 |
| Hunter  |       1 |    49 | 1,240 | 0.88 |

Reports and fights: 1YQTjgZh2KPBGCbN 1-3; 2ywhMnJq4Y3zaWtg 1-9; 2ZYxCWcLwVJar1Ay 1; ARXJvZh6kwt3np2D 1-9;
BNajn8zgV49rZ7Fm 1-18; bqW1xQpFc9ag7NtM 1-2; fAKd8D7wkgQzM9GZ 1-3; Fx34Wc78qhnPmVz6 1-5; GFkKCvMYJwfjHxac 1-6;
kma6pFnDVW1CQ7Tv 1-5; LxdYWR47ZFbTNBhM 1-2; LzFXxt3Y1HyVK4nc 1-7; tCafRNMpHyAbgcmz 1-5; wdrYBPyVFcxMX7Q4 1-2;
wPD796Jv2htngRmT 1; YFnCf6cWNRpgVMHT 17-31; ZMn6BHdqtvpGLhcK 1.

## Boneslicer (item 273819, spell 1292573): 3, measured

447 procs on 4,438 rolling hits, 25 wearers, 28 runs, in the main hand: all hits 3.02 [2.81, 3.23]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.08 [2.81, 3.49], 238 procs on
2,321 lone units, 25 wearers. Specials: 3.33 [2.74, 3.93], 93 procs on 839 lone units, 22 wearers. Two-roll abilities,
priced as one roll: 4.35, 10 procs on 69 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms,
305 isolated hits. Left out: 48 of 76 runs in Cat or Bear Form. By report date: before 2026-11-18 3.03 (211 procs); from
2026-11-18 to push 114313 3.01 (236 procs). Reports from 2026-11-10 to 2026-11-20. The sim does not simulate this proc
yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |      23 |   320 | 3,149 | 3.05 |
| Warrior |       2 |   127 | 1,289 | 2.96 |

Reports and fights: 3rbcnjTMRfwxv1mh 1-10; 3yMkaTdQx8NJXGKb 1, 3-12; 79RVgZWM8dF2wkCp 1-6, 8; 7zcat1AyFnrJ4BQR 10-19;
8fwQv1dbWXxpkG3r 1; aqCndvgQbBXL3Jf8 9-13; BzGXfjMLJ2RyCwHm 1-7; cG82vnXYjAVaTN1z 1-4; FAf9V76BctbG3mR4 1;
g74R1Wf9FL83pMP2 1; gJkN6DCr8ZfWTn72 10-13; hGNVDdLFRkj7tfcv 1-7; hLNVaj9HdnCM2BYt 1; K9WGdrx248FvaDVY 1-3;
KbNyBzwZg1t9jMcC 1; KrH8VYbgyfaTXWh9 6-8; MW8HLRKDFN4GxTr7 3-4; nKNgy3VrZXkcdjBT 4-5; NrHbC9xfM7KzZt4d 1;
pLh28qFZjAYvrMT6 1-6; qMKJYxCXzcPRZ8nD 1-3; rpB2MKnbm7LxCqJZ 1; tVB2a6fMvYNCGdkH 1-5; TZkhGPt2fqAnNY74 1-3;
vnwtmMTpkR1aYP7G 1-9; y7qfkc91hA3DKQ4g 1-11; Z26ztpb3FkJhgTAf 1-8; z3YfqBgbt1GjFD4W 1-4.

## Brain Hacker (item 1263, spell 17148): 0.95, measured

990 procs on 18,309 rolling hits, 9 wearers, 12 runs, in the main hand: all hits 0.93 [0.90, 0.96]. The rate is all
hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 0.95 [0.88, 1.04],
251 procs on 4,540 lone units, 9 wearers. Specials: 8.57, 1 procs on 2 lone units, 1 wearer. Two-roll abilities, priced
as one roll: 2.45, 4 procs on 28 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 200
isolated hits. Dropping the swapped stretches (27 active minutes, 9% of hits, in 2 runs) moved the read from 0.84 to
0.93. The paladin check, which sets no rate: lone white swings 0.65, 10 procs on 262 lone units, 3 wearers. By report
date: before 2026-11-18 0.98 (289 procs); from 2026-11-18 to push 114313 0.91 (701 procs). Reports from 2026-11-16 to
2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |       8 |   957 | 17,771 | 0.92 |
| Warrior |       1 |    33 |    538 | 1.05 |

Reports and fights: 1Tf47h2MvyczZbVK 3, 5, 7, 10-11, 13-20, 22-29, 31-33; 97HMYjdamngzW61A 1-6; CN68qfBaY3MLXkGP 1-2;
CqPk8KgFWYTMnw3D 1-4, 6-9, 11-12; FHA7TKpb3ncBqgMC 2-5; JjYkyzampgwCVPNr 1-6; jKdvhBa6WmJ1zq4D 1; KcLtDQRdjbrxyqXJ 3-5;
pQxYvP8khVb7ZMwB 1-3; tNjpX1d9L4A8CJBw 2, 4; vQfM3qZDHdmnJTpV 1-2; xm4JpaRkNncvP7dw 1-14.

## Cookie's Tenderizer (item 5197, spell 1292008): 1.95, measured

1,464 procs on 16,079 rolling hits, 36 wearers, 37 runs, in the main hand or the off hand: all hits 1.95 [1.81, 2.08].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.98 [1.70, 2.20],
277 procs on 3,004 lone units, 18 wearers. Dual-wield white: 1.69 [1.44, 1.88], 322 procs on 4,080 lone units, 17
wearers. Specials: 1.93 [1.76, 2.11], 288 procs on 3,195 lone units, 20 wearers. Direct: window [-10, 100] ms, median 0
ms, 98th percentile 0 ms, 635 isolated hits. Dropping the swapped stretches (189 active minutes, 21% of hits, in 8 runs)
moved the read from 1.55 to 1.95. Left out: 5 of 56 runs in Cat or Bear Form. The paladin check, which sets no rate:
lone white swings 2.00, 63 procs on 676 lone units, 17 wearers. By report date: before 2026-11-18 1.87 (983 procs); from
2026-11-18 to push 114313 2.14 (481 procs). Reports from 2026-11-07 to 2026-11-20. The sim does not simulate this proc
yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |      16 |   789 | 8,838 | 1.91 |
| Warrior |      11 |   527 | 5,681 | 1.99 |
| Shaman  |       7 |   148 | 1,557 | 2.04 |
| Priest  |       2 |     0 |     3 | 0.00 |

Reports and fights: 2gjR3tVq4Zdw7kzK 1-3; 2htpkPNX3yZ8K74C 1-3; 2PWL9DB1k47p8MRd 1; 8tvAkdpfn2hNc9Qa 4; 91hVt6Tpm3kgvcyB
1-13; 9gx3fpD7PA1RhCYH 1-2; adj7kHPW3ypMKrNL 1-3; aRDTH6pqdyGgPkQ1 3, 5-6; B2D9zYahGJLKbVHx 1-4; cJm7HAtbgG2hqyXz 1;
D4T2RPhgAMvFYk1c 1-6; DMZPzKB91HJAqLQy 1-4; f297FtyjGHmRVqYw 1-4, 6-9; FWHahtQvqjL3wNdV 1-3; FZB3hwQLpjHn6829 1-12,
14-19; GDWLqtcJgbrz1P89 1-8; k1tqMpa6hdByfCAQ 1; kFMtgzB1cCvxJTfL 1-9; KWFdXvCfR6rVYLcj 1, 3-12; LdcVztBw6pFjNahM 1-3;
MaXHVZ2nWDJ8z6Cd 1-3; MLkQWdC6hNcm34gz 1-3; Mv73gTFaNhbGWLy9 1-6; MxFBA2RChVbm4Ykj 1; nTY7GDAHgZf8cbzr 1-4;
QmjBxrzvah2XTVR1 7-8; qRQfj9HCFYp2MrDT 1-2; RpkqFDaLQn6bxwt8 1-8; tBnv1kDbzHK6hAwg 3-7; TFnvh6k2P8zL4DCQ 1-16, 18;
tHfZywmLT6XVpqvb 1-3; w1WaVmyY8jBT3GF4 1-8; yCTRh3Hq7GPt6nAX 1; YfWXxBnqKv7HgyQ1 1-2; ZkTHQfR8F13AdaMq 1-3;
ZTJqNPfRLb3c8Bav 1-10; Zv6fc3GgWpkX9Bqa 3-4.

## The Cruel Hand of Timmy (item 13401, spell 17505): 2.2, measured

1,200 procs on 18,223 rolling hits, 18 wearers, 22 runs, in the main hand or the off hand: all hits 2.20 [2.06, 2.37].
The rate is all hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 2.27
[1.89, 2.70], 124 procs on 1,823 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 1.91
[1.71, 2.07], 329 procs on 5,748 lone units, 14 wearers. Specials: 1.73 [1.58, 2.17], 32 procs on 615 lone units, 4
wearers. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 318 isolated hits. Dropping the swapped
stretches (43 active minutes, 6% of hits, in 3 runs) moved the read from 2.06 to 2.20. 1 run logged in two reports
counts once. By report date: before 2026-11-18 2.17 (574 procs); from 2026-11-18 to push 114313 2.22 (626 procs).
Reports from 2026-11-13 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   642 | 9,903 | 2.16 |
| Warrior |       9 |   558 | 8,320 | 2.24 |

Reports and fights: 1ZyhaVDdFtTYWMcC 4-7, 9-12; 3BAz1TDGL9kjy8ZQ 1-2; 72q1FtWgTMHAcLxG 1, 3-5; 8jDqr3JcYRCH1pbK 1-4;
aZ8yxm3AGgrpTvYh 1; C1cyPbL3RQvGjq8p 48; H8AgW6afwBktp4nv 3; HKjBYG4kRJtnXyc3 1-2; HL1FjV9MzyNDrb48 1; hLxABfWVavkM8HGT
1-4; KtCPg2xhR416AvXN 1-6; p6mVRBzvWMqnQZFc 11-12; qa8wcv4jLYZpGARK 5-9; Qb1yWxa8vHCGL7Yr 2-7, 9-15; qcdyZAr2ganDtxNW
13-44; QcrC9VK64xXa1JWg 1-8; rRnVJhQFa9zf7NqK 4-5; vT1rMa2cNGXLfqp7 9-12; yCdRrMhYwm69xj1Z 1; zBqkZQ1bxp3cAhDa 15,
17-18; zy3NjCQP7bMp1W8H 1-2; Zz3YcQk69P7N8rAn 3-4.

## Eater of the Dead (item 10805, spell 1299869): 1.75, measured

1,865 procs on 29,229 rolling hits, 26 wearers, 30 runs, in the main hand: all hits 1.74 [1.64, 1.83]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.86 [1.49, 2.13], 348 procs on
5,097 lone units, 10 wearers. Dual-wield white: 1.63 [1.46, 1.81], 140 procs on 2,337 lone units, 14 wearers. Specials:
1.75 [1.51, 1.97], 206 procs on 3,210 lone units, 23 wearers. Two-roll abilities, priced as one roll: 2.84, 27 procs on
259 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 500 isolated hits. By report date:
before 2026-11-18 1.73 (1,084 procs); from 2026-11-18 to push 114313 1.76 (781 procs). Reports from 2026-11-13 to
2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      12 | 1,547 | 24,357 | 1.73 |
| Hunter  |      11 |   121 |  1,805 | 1.83 |
| Rogue   |       3 |   197 |  3,068 | 1.75 |

Reports and fights: 1PcrGTyatDAFx8Yk 1-2; 23CWz6vwhH9YdxN7 1-2; 2tCNaTJf49KWyVLk 16; 3f7wRc8vmxMHFa9n 1-3;
3hpaqBV4jbGdx2TX 12-13; 6fqDJhcVRap21w8W 3-10; 8yPRmZdrq4QhcGn1 1-4; af6xmcjndRMkZ83N 1; aVDjBvL8hyfCGmtq 1-2;
bDcBjXw82GvhYW4F 1; C834MAj1cKh2RZyd 1, 3-7; CJqgRYWTnm3jHGxP 1-2; g7ZWhALb3rT6ctfq 1-19; jLyXfQ2Ftp8RJmNd 1;
mdXkc6xwP8Y9pR47 1; mJDNgqYVGwb6A8tF 1-17, 19, 21-22; nc1zq8VBvZCTJtjX 1; NdbCqzpGA8Yr9ZfL 1, 3-6; pYATGQ6aCmzwBNcM 1-8,
10-13, 15-25; qQwFaXyTdhmnr2kP 1; RHgby76wQXp2nrxj 1-4; RJ82yLB7HcYrd39M 2; V8F7qLQXzpPAmcyG 1-4; vGdq3ZrKyMAPaTJD 5-8;
VYBjgJx8aZtRhmpy 10-46; WgQrVv2BTbPGHKFD 1; wk2zgGAXYTbnHQ9V 1, 3; WrBH1KZxJvVP8hQk 1-2; ya61L9CjgmMWAQTJ 1;
YXTfd7QPJV4Nqhcn 1-18.

## Firebelcher (item 5243, spell 1291686): 3.95, measured

1,416 procs on 12,725 rolling hits, 64 wearers, 70 runs, as a wand: all hits 3.93 [3.74, 4.14]. The rate is all hits,
which the lone wand shot rate agrees with. Lone wand shots: 4.01 [3.83, 4.24], 1,257 procs on 11,060 lone units, 63
wearers. Direct: window [-9, 100] ms, median 20 ms, 98th percentile 36 ms, 1,054 isolated hits. By report date: before
2026-11-18 3.78 (548 procs); from 2026-11-18 to push 114313 4.02 (868 procs). Reports from 2026-11-06 to 2026-11-20. The
sim does not simulate this proc yet (the generator lists it as a missing effect); the rate is recorded on the item's
effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Priest  |      23 |   697 | 6,389 | 3.85 |
| Warlock |      22 |   502 | 4,491 | 3.95 |
| Mage    |      19 |   217 | 1,845 | 4.15 |

Reports and fights: 1RrDAhp7ZvVYc9P3 1-6; 1RrFMXYLPJhT7xBn 1, 3-10; 1xDPwjtGBTzZWm7Y 1-4; 1ZJVMgcYWTyRQh2w 1-3;
287trgyqQLvjACK9 1-4; 2mW9A7wQ4bhv1YMF 1-3; 3PnWT1jGz4tagB6r 1; 4NWMr2dJPyH9v8az 1-2; 4P93K2QXbxJ68qrH 1-2;
4vca1FKMyTnP97Az 1; 6WK7XDRHBNCx9kwn 1-4; 6wxJrRWgPyQb21hm 1-7; 6WzAhcVgfLCt78nX 1-8; 7xdwabvCPyBXYRFq 1-4;
8rdDTxLMQPKqYyAz 1-3; 9pTK7ZwNabjMydzV 1, 3, 5, 7, 9; 9wKZ21fVDH3xpvcT 1-4; aNFQk87zxVRKA2nf 1-6; APhXWvd6CkrHL4N3 1-7;
aryNkPZX814F37pJ 1-2; Bm6KyjgHafc19xP3 1-3; c3NLR49a1VwfCQtG 1-7; CGkfgyRY8mQ1Aah2 1-2; d6bwAkLKjGq7pVMy 1;
Dh2tFcWaMVng6Cdj 1-7; DHv3m4KLCdntG7pV 1-3; DmkRhHYPg4wVJf3t 1; DNJ682zfmT7Zykv4 1-2; FADNxntvpJkfM9RC 1;
FWf1LcA7PCb6NQB4 1-4; Fwr6n9BDgLaGPmKY 1-10; gmTPbGL4cNDJ92fH 1-3; Jn4FXD6QKW71pRzh 1-2; kBzVYb3Cy7QwnKMD 1-6;
KDFtbX6ZVxMgATLh 1-2; KkCznWJf8yYHvGgX 1; kpNtDy4ZYjLTGAbd 1-6; KYD67nVqZT4jxWyN 5-6; LAP9T3KqpjhC4MD2 1-7;
LbmB2QdWnPT7JGAF 1-6; mbBT1kLrM8qXzdJF 1; MN6V1awdYqymj4vr 2; mPz2LqFWbhKcV7QN 1; MX1B9kCJ7GDTQ6za 1-3; NGhbmnMwKCHYfaX6
1-5; nT2Wz7mVQZbtBRG4 1-4; PCYpB6DhnQMmXbzN 1-2; PLKnY9R2BDgwbdWv 1-2; qFZN9PQwagyL3Xnz 1-3; QHqMJPfgbBCvZan7 23-28;
qL6waTVj3WhQYk8x 1-3, 5-21; Qvg9N1YBc36d8Rnw 1-10; r6dAtgyVpcMXkL3T 1; r9X8bgWqQfnPVDyB 1-2; rCnfgxYFVXLHvjZp 1;
RKdyjLbQCZgtzrXp 1-3; tAZ12pKHxQYLydgv 1-2, 4-5; TGVQg7A3KfJdpbkw 1; TkMHbgQNKBt7Vzy9 1-13; TQCphZqxGzRmkgjf 1-2;
tvDFpcQ4w3L7PBy1 1-3, 5; txZvdrQzpwn3V4kb 1-4; vVa1NxRf6rZnYhWH 1-2; vWGgArDXtc13Lx6J 1-9, 11-14; w1pq2CGTWMbX7dQZ 1-2;
wDqmkbxdTXj1rA9F 1; XPmJACH8RtqFd3jM 1-3; yAGCbvd2r4RZNDaK 1, 3; Z9qbhQrJg2Wfz1FM 1-2.

## Hammer of the Northern Wind (item 810, spell 13439): 4.5, measured

4,649 procs on 29,390 rolling hits, 12 wearers, 16 runs, in the main hand: all hits 4.52 [4.40, 4.68]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 4.46 [4.02, 4.81], 366 procs on
2,345 lone units, 5 wearers. Dual-wield white: 4.21 [3.20, 4.59], 159 procs on 1,080 lone units, 7 wearers. Specials:
4.17 [3.76, 4.86], 338 procs on 2,314 lone units, 12 wearers. Two-roll abilities, priced as one roll: 7.79, 3 procs on
11 lone units. Travel time: the proc lands a median 116 ms after its hit, 98th percentile 326 ms, window [-10, 336] ms
(146 isolated hits). Dropping the swapped stretches (20 active minutes, 1% of hits, in 1 run) moved the read from 4.47
to 4.52. By report date: before 2026-11-18 4.38 (1,925 procs); from 2026-11-18 to push 114313 4.62 (2,724 procs).
Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 | 2,685 | 17,517 | 4.38 |
| Rogue   |       4 | 1,964 | 11,873 | 4.73 |

Reports and fights: 4Mcm2YV8yDhPnwrb 1; 8qACFaGbm1jn3PV9 14-26, 28-29, 31-35; cQpq1A6BxVMZFmfr 1; dBNfgVDLMmnc8h4v 10;
Gak2MmQ34xn9ctPj 1-34; J38xbrFytgM9p1LT 7-19; jvqAx3thp6wTz9ZR 1; krA2h3ZCfBXMmT4F 1-8, 10; nJZvaYCj2brRc7Fw 1-24;
nyAx3XNDVmYWGdrh 1-6; P4bXMwRjD1zyhZdJ 1-2; pravyY1JCwDAxZLb 1-24; qR2rnzfgPakD879W 1-2; TF6dc3NCHbKrfxJt 1-6;
xbg9BypFcdVDwqMk 1-5, 7-14, 16-42; z74YTAg1Jc2QFWDC 1-2.

## Kang the Decapitator (item 2291, spell 17153): 0.95, measured

1,200 procs on 21,014 rolling hits, 17 wearers, 22 runs, in the main hand: all hits 0.95 [0.90, 1.01]. The rate is all
hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 0.88 [0.73, 1.05],
274 procs on 5,195 lone units, 17 wearers. Specials: 0.68 [0.00, 0.88], 6 procs on 146 lone units, 3 wearers. Two-roll
abilities, priced as one roll: 2.73, 9 procs on 55 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th
percentile 1 ms, 253 isolated hits. The paladin check, which sets no rate: lone white swings 0.90, 34 procs on 628 lone
units, 6 wearers. By report date: before 2026-11-18 0.94 (670 procs); from 2026-11-18 to push 114313 0.96 (530 procs).
Reports from 2026-11-12 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      15 | 1,067 | 18,766 | 0.95 |
| Warrior |       1 |   116 |  1,950 | 0.99 |
| Hunter  |       1 |    17 |    298 | 0.95 |

Reports and fights: 2FkYRHB1dWhra8vt 1-17; 38kdWHmLZKVgTPqb 2; 6kf3tMF4wCPNcRnh 1-33; 6wxzV4YyNCh1aqfd 1-9;
9Z473MfXLYqcyWPg 1-2; aLw2pgPCyFKTVGYm 1-5, 7; C8Qd1LXMrGT62kpY 8-9; CNqm64YkTthB7V2g 1; DL9nAWQwcN3rvZ4K 1;
F6AaCcJQvdP1nqRN 1-4; FMaAjHcwCQLkpV32 1-4; gALWqjdyncf8xVPa 1-2, 4-7; QA7aBZpK8hGJn9rf 1; qR2rnzfgPakD879W 1-2;
r9Yj7KWkaqL4B3mp 1-12; RfT7K4v2jhY9FAzD 1-3; RpFQn1ftYCAgMW47 1-17; tBk9vLmJgqwX84ar 1-7; W78tmfknBKz46hRZ 1-2;
xcPjNLDXa9nz1Vrd 1-7; YpxNGMXcf4PyTwt6 1-6; ZNyPLR2vkjMGptCJ 1-2.

## Mana Channeling Wand (item 18483, spell 1302348): 4, measured

1,030 procs on 9,616 rolling hits, 38 wearers, 55 runs, as a wand: all hits 4.02 [3.76, 4.29]. The rate is all hits,
which the lone wand shot rate agrees with. Lone wand shots: 3.81 [3.59, 4.09], 840 procs on 8,275 lone units, 38
wearers. Direct: window [-10, 100] ms, median 10 ms, 98th percentile 22 ms, 810 isolated hits. Dropping the swapped
stretches (5 active minutes, 1% of hits, in 1 run) moved the read from 3.99 to 4.02. 1 run logged in two reports counts
once. By report date: before 2026-11-18 3.98 (423 procs); from 2026-11-18 to push 114313 3.97 (486 procs); after push
114313 4.30 (120 procs). Reports from 2026-11-11 to 2026-11-21.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Priest |      29 |   999 | 9,420 | 3.98 |
| Mage   |       9 |    30 |   196 | 5.74 |

Reports and fights: 1BLCZMxRfyYmHXNn 5-6, 8; 1VpHn3kYq6B8JMWw 1, 3-4, 6-7; 26AYbFNQzMZCmfkh 12-16; 2GczY83pjnCHdTNA 1-2;
3zwbTGcyQBLm7DMg 1-7; 4thjWV7Xbz2yrxBF 6; 6hDJQWTXwBr419AM 4; 81rnYqNT3djhJFRy 3-4; 8HvfKkArXpmbhwTF 1; 8YVKbmp1y4rMPFdW
2-11; aCR8n63FzdTYGqj7 1-7, 9-10, 13-24, 27-29, 31-36, 38-40, 42-57, 59-63, 65-67; aYBjphvLn8xMK2TN 1-5;
c7M6PR9WHkyDJL2q 1-6; cav7VZrmPbNW3y29 13; DvrjQXqZ9HN8a74z 1-4; f2CmVRcTnthPW6aH 1-14; fjtT3JVxbGPz6RwX 6;
GAaNPKjbZkwLBCT2 2; gBDMLv8J1K7nqkpd 112-157; GFBCh3VZH7TmAkL4 1-3; gw8fpa1YVkv4WP32 1, 3; hfNjty7QRBrFCKm6 1-3;
Ht8VFJxmnCG9bQhM 1-3; hYjVWRNBGDKLxTpc 1-3; JvFPcBTz9rZMmW7h 1; KNtHCAPd97ZMTLBf 1-2; Lb7Zcx2JtgjCTwP8 1;
LN2Vk4nMtBc8vWbT 1-6; m3ctFRXJaZnrKzkC 1-3; mCn63gBZfLFHaWr9 6, 8-34, 36-39, 41-43, 46-53, 55-89; mLH64Yyjpbd3AKBW 1-4;
mnPXjHcWNFKBhDCq 1-2; MPfbhA7yLxpkdtZQ 5-14; n8VK43MDGChq6rgv 3; p8WXHnkF6m4rvGJL 1; PDwtabFTz3VnQqrH 1;
pG8TNJwybLPYWmxF 2; PKBd6Vz2Gkjwp9Zt 1-2; pqxm14QnyLk9tCGK 1-2; QFqKVZgc6vAJfbrW 1-19, 21-28, 30-31; Qqf8ZC1vRYhwG2cr
4-7; QXFd74zDW8JZMmKx 1-2, 4-13; RJXQC67jbzATWdx1 1-2, 5-9, 11-16; T7dtHBv9YR2jnLwg 1-7, 9-10, 12-19; tG4CqkJQvyPT2A1Y
1-10; tHzN3kKbaPpVW8TY 1-8, 10-11; TXGK16YwMmcyj9Cd 2-3, 5-10, 12-13, 15-17, 22-29, 31, 34-74, 76-86; V9ZpFqnkLrbaT6zy
1; VvFWzAgjYPxwN8fT 32-35, 37, 39-42, 44-49, 51-52, 54-55, 58-61, 63, 65-67, 69-72, 75-79, 82-85; wkJXWxD23VMqThLg
14-17; wR8KA9WTgHkD6n7q 1-10, 12-34, 36-47, 49-62, 64-78, 80-94, 96-111, 113-128, 130-146, 148-202; Wv4QYA6C7antxKXr 1;
XD23HFnxbzvPjACV 1; zA8JyGdnWRQ9vXmK 1.

## Pyromancer's Baton (item 286650, spell 1320925): 1.95, measured

964 procs on 12,402 rolling hits, 16 wearers, 17 runs, in the main hand or the off hand: all hits 1.94 [1.82, 2.07]. The
rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.85 [1.50, 2.26], 98
procs on 1,322 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 1.43 [1.04, 1.81], 147
procs on 2,562 lone units, 10 wearers. Specials: 2.17 [1.80, 2.54], 114 procs on 1,316 lone units, 7 wearers. Two-roll
abilities, priced as one roll: 1.92, 1 procs on 13 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 275 isolated hits. Dropping the swapped stretches (24 active minutes, 4% of hits, in 3 runs) moved the
read from 1.87 to 1.94. Left out: 1 of 18 runs in Cat or Bear Form. By report date: before 2026-11-18 1.94 (725 procs);
from 2026-11-18 to push 114313 1.96 (239 procs). Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   526 | 6,578 | 2.00 |
| Warrior |       6 |   382 | 5,029 | 1.90 |
| Shaman  |       1 |    56 |   795 | 1.76 |

Reports and fights: 4aZmt3y9znXjhRvV 1-2; 4NMcqLwQ2rB9amCH 1-3; 9hywC8jMTLV36Pxd 4-8; 9rPx2JNYLbCy1waX 1-9, 11-15;
ARqz9awMbfr6QD7Y 1-2; Brj9PhXHQM3kLN2f 1-4; HqwWnYGKdhpCgZVR 1-12; PamDnk2rVQhW7Cj6 1-5; Pw9kWLGcAj3VpmCn 13-18, 20-22;
q6VRad2JDnpxXYAL 1-13; QPmw1C9MNc4hqH3K 1-5; ryqgQphLxMBz4TmH 1-4; Tbh1yKNfPtqLYCak 1, 3; WhJ7vPzQnbFkmKyj 1-12;
Wv3xD7KBm1C2tXNL 1-8, 10, 12-19; ZAjYLxr6TB7KGmav 1-9; zHYaAjmvQBT2fFW7 4-15.

## Scorpion Sting (item 1265, spell 18208): 0.95, measured

2,096 procs on 54,256 rolling hits, 59 wearers, 64 runs, in the both hands or the main hand or the off hand: all hits
0.97 [0.93, 1.01]. The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white:
1.03 [0.86, 1.24], 144 procs on 3,492 lone units, 10 wearers. Dual-wield white: 0.81 [0.73, 0.92], 352 procs on 10,880
lone units, 48 wearers. Specials: 0.99 [0.90, 1.10], 306 procs on 7,693 lone units, 43 wearers. Two-roll abilities,
priced as one roll: 1.05, 15 procs on 357 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms,
559 isolated hits. Dropping the swapped stretches (81 active minutes, 3% of hits, in 5 runs) moved the read from 0.94 to
0.97. By report date: before 2026-11-18 0.97 (1,403 procs); from 2026-11-18 to push 114313 0.97 (693 procs). Reports
from 2026-11-10 to 2026-11-20. The sim does not simulate this proc yet (the generator lists it as a missing effect); the
rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      29 | 1,107 | 28,298 | 0.98 |
| Warrior |      26 |   970 | 25,418 | 0.95 |
| Hunter  |       4 |    19 |    539 | 0.88 |

Reports and fights: 1TCjB8HdypDcxbt2 1-3; 26tvRFXhzdqrLApN 1-3, 5-9; 4ta7nVwdhkXM389v 1-3; 64NVwXYTpFRBxJWz 9-11;
6NrFMaRQ9yAjxWC2 11-12; 7L198CkvgnJyKAHm 1-14; 7vxRAKHWC83tgPDz 1, 3-7; 8cj6fJ9kYvGmqx3W 1-3; AbnJ1XMLhwqtpx6F 1-4;
avrcxzCymh1DNqA8 17-19; Azq6hMZn2RyWwagv 1-6; bHkgvJcQaRpXYB3y 4-11; BM8NmpvwVdzA76f9 1-9; BMPCZ3a7y9HzJYfX 1-4;
bpM8VcnfdmyxZ1vg 10-17, 19-20, 22-23; bpV7JWX8B9xgrj3c 1-5; Br2pycwj4zLC1ZP7 1-3; CdgL43HfzxKMhtPv 1; CKLMhw2QbTVc8Hym
1-14; cPHdgRMCy7ZGDqFf 1-8; D8Jwdf1PkWm2y3YH 4-7; D9vw4QTYJZWMPnph 1-3; DKpBr8Y6RZt7yMca 1-8; Fdw3nRjaYcDbvyPC 1-9;
gk2jJvTHhW4zNDKB 1-2, 4, 7-11, 13-18; gn4HAC1RqMfLt8wN 16-21; h1TVXKmtQwWpad6C 1-17; HrJFDLXd1NVmbGv3 1-5;
j9x1hJfM43nNpzrv 1-8; kBb8HK6JDdaX2nZq 1, 3-9; KGgcP8Jxz12mNwfv 1-15; Kt8wFb3qThrVjNGY 1-5, 7, 9-11; LJMmvtPTzdCpG9k6
5-9; LMdRqJGhV17XwWyT 16-22; M4CvXLZz3qm6pDRc 2-3, 5; nBHjcV4Ydm7z2X9q 1-3; nJgV6Ma2Qhkzm8Bt 1-5; NmwLfZHTpdqPWAMC 5, 7;
Ntvp7VG8kdm6Ja4g 1, 3-18; nvcCpPrdZWGBba36 1-2; pgjc6w1CaJ472ybt 1; pLFZK3MWV2TRw6t7 1-2; pmKDwngMT2NFktdA 8;
PNnXcA9hFBx7Wtdj 9-11; PxfYmBXLqFg3GMvT 1-26; Q4GadwADtxbFBNW9 4-8; Qf2RhFMXYDwBCzHg 6-8; QqVNrzBdvxnwjGYb 1-7;
rhWaxQK8cv37nzXN 1-5, 7-9; tDraZ1AfWnGhNYvg 1-6; TFzR3fQkt2ap7K4Z 1-10; TndajWC3c76QymF4 1-5; V8bxkrXJqAK6v3cL 13-14;
Vk9QFCRcmdypPBvt 1-4; VYb6f27NazpmCch3 1-2; w4qRTLXNzP7a9cQ2 1-6; Wgk8whB6HQNZmrnq 1-5; xMw1vFZhCGJN9YQP 3;
XNry2TxRndjwJQBm 1; xwT6Z1CV7WQhBmgD 1-3; yjqVPg81hM6TCW97 1-2; ynGtWRk4cZX3TmYv 1; zbg1wFp2L8hWHZva 1-26.

## Searing Blade (item 12992, spell 1299850): 1.9, measured

392 procs on 4,944 rolling hits, 17 wearers, 17 runs, in the main hand: all hits 1.90 [1.78, 2.06]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.89 [1.72, 2.21], 169 procs on
2,149 lone units, 17 wearers. Specials: 1.88 [1.31, 2.38], 60 procs on 766 lone units, 15 wearers. Two-roll abilities,
priced as one roll: 2.48, 6 procs on 58 lone units. Travel time: the proc lands a median 98 ms after its hit, 98th
percentile 214 ms, window [14, 224] ms (74 isolated hits). Dropping the swapped stretches (36 active minutes, 14% of
hits, in 2 runs) moved the read from 1.63 to 1.90. The paladin check, which sets no rate: lone white swings 1.78, 60
procs on 807 lone units, 8 wearers. By report date: before 2026-11-18 1.93 (336 procs); from 2026-11-18 to push 114313
1.74 (56 procs). Reports from 2026-11-13 to 2026-11-19. The sim does not simulate this proc yet (the generator lists it
as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |       9 |    29 |   301 | 2.31 |
| Warrior |       8 |   363 | 4,643 | 1.88 |

Reports and fights: 1K6c48bWMN9nzCZA 1-3; 2L1XYNpZCTR4nG9t 1-4; 3tFdrmh8baCZpNn7 1-6; 3yM21KdFP7LnrBvx 1-3;
6Tyb2X1ZNA7CkaVx 1-2; 7ZxcPNTjqvfzQHgX 1-2; GdDn73hPYt2JQmHL 7-8, 10-12; GdVxzv2hHjNtWMcn 1-20; kNYgfyZF4LV8QmXq 2, 4-5;
nG4DqxBAmQgZLf8a 1; PNYpz2hCmGMRV1tx 1-3; RQPftYBgN74bAz1q 1-5; tBnv1kDbzHK6hAwg 3-7; V4Cr3TxcZM2kDqtG 1-5;
ycJpHf8KPAVG9x4F 1; Yd4r69MDqbQyJCFH 16-36; zjcGFyxpkKTnQXdw 1-11.

## Serpent Slicer (item 13035, spell 17511): 1, measured

1,105 procs on 26,110 rolling hits, 24 wearers, 25 runs, in the both hands or the main hand or the off hand: all hits
1.02 [0.94, 1.08]. The rate is all hits, which the specials rate agrees with. The one-hand white swings stay under 50
procs. One-hand white: 1.23 [1.03, 1.35], 26 procs on 506 lone units, 2 wearers. Dual-wield white: 0.91 [0.76, 1.05],
228 procs on 6,034 lone units, 22 wearers. Specials: 1.05 [0.86, 1.23], 139 procs on 3,163 lone units, 15 wearers.
Two-roll abilities, priced as one roll: 1.36, 14 procs on 247 lone units. Direct: window [-10, 100] ms, median 0 ms,
98th percentile 1 ms, 278 isolated hits. Dropping the swapped stretches (24 active minutes, 3% of hits, in 4 runs) moved
the read from 0.99 to 1.02. 1 run logged in two reports counts once. By report date: before 2026-11-18 1.05 (637 procs);
from 2026-11-18 to push 114313 0.98 (468 procs). Reports from 2026-11-13 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      11 |   633 | 14,857 | 1.02 |
| Hunter  |       7 |   160 |  3,815 | 1.01 |
| Warrior |       6 |   312 |  7,439 | 1.01 |

Reports and fights: 2MKaBZy3AJqLdjPF 1-2; 8L9G1c3bKt6BAxJW 1-10, 12; 9Jj1FDzGp7BRQ648 1-15; 9wtfbKA27FNQ6BaW 1;
C4TGJvzHF3aZ2NjR 1-13; g4Y2frPx8kBmj9Tq 1-2; Gh3vdkL7AJrm6fFj 1-11; gywYLPFQCJa9XfVG 1-6; hYHt6kLTgApm8RMw 1-2;
nfLmGxNDqJpyXB6k 24-38; NrJZ7avW13QBVTGw 1-7; nrKZtkQhgbBYf7R6 1-6, 8; PmM8nxa1p6QkhVLC 1, 3-7; rAJbdCG47n2hgwVB 1;
Td2y9DjVvn1YgaHr 1-7, 9-21; VgbFTznDNP4XdAh9 1-13; vk9HJZwrfhmCayMY 1-7, 9-10, 12-21; VKHmAfvPjBGLQW4Y 1-9;
vpTQqDkgNYLnjtZr 1-47; wnMHyXgW6KpAhYF4 1; XkhC2YtZ6cMBHGfz 1-2, 4-6; Y1vtT4JxK97fhQrL 1-3; ynGzYZ3cLrpvwj1N 1-9;
Z7qF4j3hazBTxHcw 8-10; zMqZ3BNKbFYfAjwp 1-12.

## Sliverblade (item 5756, spell 1293653): 2, measured

1,267 procs on 26,995 rolling hits, 21 wearers, 23 runs, in the main hand or the off hand: all hits 2.01 [1.93, 2.13].
The rate is all hits, which the specials rate agrees with. The one-hand white swings stay under 50 procs. One-hand
white: 2.39 [1.26, 6.35], 30 procs on 538 lone units, 3 wearers. Dual-wield white: 1.79 [1.63, 1.98], 299 procs on 7,139
lone units, 18 wearers. Specials: 1.75 [1.33, 2.94], 62 procs on 1,516 lone units, 10 wearers. Two-roll abilities,
priced as one roll: 5.36, 1 procs on 8 lone units. Travel time: the proc lands a median 89 ms after its hit, 98th
percentile 221 ms, window [21, 231] ms (348 isolated hits). Dropping the swapped stretches (22 active minutes, 2% of
hits, in 3 runs) moved the read from 1.96 to 2.01. 1 run logged in two reports counts once. By report date: before
2026-11-18 2.02 (704 procs); from 2026-11-18 to push 114313 1.92 (457 procs); after push 114313 2.43 (106 procs).
Reports from 2026-11-13 to 2026-11-21. The sim does not simulate this proc yet (the generator lists it as a missing
effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      16 | 1,066 | 22,585 | 2.02 |
| Warrior |       5 |   201 |  4,410 | 1.95 |

Reports and fights: 43wtgArG9vm8pb6N 1, 3-4, 6-16; 4PhL971zdKXZtnVr 1-2; 6RPtC7B1WwVbaH2r 1-4; 9g7NQBqDH6FGfP8L 4;
AFN6pDm8G9tq2Znd 1-6, 8-10; aPRXW1cvbptnDhj3 1-4; BPxvyfp6nT2NwjKX 1-8, 10-17; c69DL8gBmzZhHvdA 1-4; JHrkTPq23LpAcGxY
14-18; MTCvzYwVy4qrX8Gb 2-11; mYP6QAb2CJxFtknM 4-6, 8; nZ3kwNftM8h2L1zd 7-8; PkWRr8jZp9ATL7N4 1-23, 25-26;
PZKA7G2kQtw9H4Dx 4-9; QjJ6MpmFtXzdkRyL 1-2; rh3y7JwG9fVxq6DX 5; RnfbW8KTc4JHypP6 1-4, 6; TbB8q4RMK3LmYJNv 1-6;
Tm68LdcfN1rZYFap 1-13; xDcNRgK2YCA78yvd 8-10; yBdgRzfVhDPWAnbZ 1-7; Zd7QD3Lw1TBvmFCa 1-13; zMTFJZLpPfmXYv3W 1, 3-17.

## Sul'thraze the Lasher (item 9372, spell 1294430): 1.45, measured

625 procs on 9,926 rolling hits, 8 wearers, 9 runs, in the main hand: all hits 1.45 [1.32, 1.56]. The rate is all hits,
which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 1.28 [1.10, 1.48], 126
procs on 2,267 lone units, 8 wearers. Specials: 1.72 [1.39, 1.98], 49 procs on 659 lone units, 8 wearers. Two-roll
abilities, priced as one roll: 2.70, 16 procs on 137 lone units. Travel time: the proc lands a median 99 ms after its
hit, 98th percentile 189 ms, window [23, 199] ms (168 isolated hits). The paladin check, which sets no rate: lone white
swings 1.43, 98 procs on 1,586 lone units, 16 wearers. By report date: before 2026-11-18 1.46 (415 procs); from
2026-11-18 to push 114313 1.44 (210 procs). Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   613 | 9,677 | 1.46 |
| Hunter  |       1 |    12 |   249 | 1.11 |

Reports and fights: 2ZHDh8LX4mKqctbz 1-3; 8ZdnfAT79zPwmyVM 1, 3-6; Ag69GWdLRDhKCrcT 1-6; cL3KMWpnXyvDrqa2 1-2;
hCPLc2G7KAVX8Bby 1-6, 8-9; HT2qmWDJNfXVj1xA 1-2; mYaWhF4n8zRLDZtw 1-7, 11-12, 16-17, 19-20, 22, 24; tnyh4dxq6pNXQzLw
1-3; tVhnGRkHP7f2xarB 1-3.

## Supercharger Battle Axe (item 9486, spell 13527): 1.9, measured

258 procs on 2,923 rolling hits, 9 wearers, 9 runs, in the main hand: all hits 1.89 [1.72, 2.00]. The rate is all hits,
which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 2.08 [1.53, 2.50], 68 procs
on 702 lone units, 8 wearers. Specials: 3.21, 3 procs on 20 lone units, 1 wearer. Travel time: the proc lands a median
95 ms after its hit, 98th percentile 161 ms, window [17, 171] ms (71 isolated hits). Dropping the swapped stretches (16
active minutes, 9% of hits, in 1 run) moved the read from 1.71 to 1.89. By report date: before 2026-11-18 1.90 (187
procs); from 2026-11-18 to push 114313 1.88 (71 procs). Reports from 2026-11-13 to 2026-11-20. The sim does not simulate
this proc yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       8 |   251 | 2,863 | 1.88 |
| Hunter |       1 |     7 |    60 | 2.50 |

Reports and fights: 2axJHNwqQdtW1jnF 1-3; 7kxNnWGw2FzfLJtR 1-8; argXcZydQj9tTHD4 1-2; Cc32w7hgMJQ4Tny9 1-2, 4;
fDycAFZWt4NkRX1x 1-3; FRxmtWLHr3wvB8Q1 1-8; LjFr1v7pJKVcfhWB 1-11, 13-14; rNwJc31qXfzWthQD 1-5; tBnv1kDbzHK6hAwg 7.

## Sword of Corruption (item 13032, spell 17510): 1, measured

1,249 procs on 35,139 rolling hits, 59 wearers, 67 runs, in the both hands or the main hand or the off hand: all hits
1.02 [0.96, 1.07]. The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white:
1.07 [0.88, 1.36], 144 procs on 3,857 lone units, 9 wearers. Dual-wield white: 0.89 [0.80, 1.00], 440 procs on 14,093
lone units, 50 wearers. Specials: 0.99 [0.79, 1.20], 123 procs on 3,551 lone units, 24 wearers. Two-roll abilities,
priced as one roll: 0.84, 1 procs on 34 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 522
isolated hits. Dropping the swapped stretches (87 active minutes, 7% of hits, in 4 runs) moved the read from 0.95 to
1.02. 1 run logged in two reports counts once. By report date: before 2026-11-18 0.99 (944 procs); from 2026-11-18 to
push 114313 1.11 (305 procs). Reports from 2026-11-11 to 2026-11-20. The sim does not simulate this proc yet (the
generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      31 |   730 | 20,969 | 0.99 |
| Warrior |      17 |   486 | 13,169 | 1.05 |
| Hunter  |      11 |    33 |  1,000 | 0.94 |

Reports and fights: 1h9vznDdxyjC7KWZ 1; 2mr4qBpFYGxVMAnj 1-2; 3npHxB84DT2zRtmN 1-4, 6; 4ChPnBrVK2fJLMaX 4;
62WRbBjFQCpZGr3z 1-4; 6dDjtwh1HXFxg72L 1; 6XBqk2G4LZy83mQz 1, 3-7; 7CcF8GHPv4fagQYh 1-10; 7LY3ADmFM1wBarRn 1-4;
7tnLcDMYbdk6P9Jr 1; 87BwAPvWYCTnKJ6Q 1-7; 8dt4TNMHjYCKZ3Xr 1-7; 9bvQ2gKdAt1nFx8C 1-2; AY9aT3fZ4RvQKjrV 1;
bYKdfpLMCNQDHF7v 5; cBQbDVqmdMtPCwLg 1-3; CkcZ7wQRdhpjYPfb 1; D1qH8nWkjQBLxPbX 1; DfMnbv3KLYmp2aCX 1-4; dTc2jh81CHgYFRZ7
1-6, 8; fJxRkQjHap2ThDyd 1-8; g3BPNGHVcYkAwx72 1; G4yA72PjmHJvBKDh 1-2; GV4wtvzD8RcfTqNC 1-2, 4, 6-15; h9ac8AXNDbLwJCzy
1; hByjwtf1pNa8q2Ac 1-2; hXv27Jz1dw8tVNPA 1-3; Jay1j4Q6gWNLZxD8 1-5; JCFHXZxhBYjP2vt6 1-12; JGyv2ZTWV6C4R1wj 1;
K14qDRFZjrfzYCxt 2-8; K1XzMGFDwrt7RQbv 1-3; k6bWMqt1BpjXyZAJ 1-2; kAJd8bLqcM1nNyfh 1-6; kbPtGMLcJ2Brp9vN 1-2;
KfAmWF67gcjrMw2k 1-5; kHfNcJTYj3RtMFGL 1-3; kn8aQKN2HBhqjG9z 1-3; KNhjmWPtyVCkv6B4 1; kNhVJ9WF2LmXab4w 1-2;
L4ZWkKT3cGNCAqnX 1-2; mqVALKC7PcRFNvky 3-5; NfDXwFm37WHghq9d 1-5; nKCDhXNbHpJaBcAR 15; nYycdrvkhx8RHCDw 1-4, 6-13;
p3dMajX1JRAwVrLW 1-2; Pck4Hb83WgXxrfJF 1-3; PKpXc2n7WrBjxfCA 1-2; pKwNTxL79kXv2tha 1-2; pwNAZn643XgzcbYa 1-3, 5-6, 8-10;
pyAxZXhWjqkY6NGD 1-4; QZcRmKX4vMVh7f6j 1-14; tAZ12pKHxQYLydgv 1-2, 4-5; Tbh1yKNfPtqLYCak 1, 3; tYHypqn6LDjXfxGR 1-5;
TYqhJAWQ64zP7Zjw 1-2; v6Ljcx9QBTZNakgt 8; VxmDcHTK8LqgNCwQ 1-7; WvJtM32w6FALXbPK 8-9; x2LqDAP6WtnJvMKk 1-2;
xN4Jd3VCWBZpnM1F 1-6; ybvr1GhnNfc84wBH 1-3, 5-6, 8, 10; ycZLWXxQ4jk21w3q 4-28; zaV76QJxyhbwD4pR 1; zcK9qfp32VnQWTMP 1-2;
Zd8PXKqcNQygRjT1 1; ZHK32LtTCmqR4Pxf 1, 3.

## Twilight Forgehammer (item 281665, spell 1312176): 2.95, measured

2,280 procs on 18,539 rolling hits, 13 wearers, 16 runs, in the main hand or the off hand: all hits 2.95 [2.84, 3.07].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.08 [2.75, 3.44],
311 procs on 2,422 lone units, 3 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 2.37 [2.00, 2.77], 373
procs on 3,785 lone units, 10 wearers. Specials: 3.20 [2.97, 3.38], 290 procs on 2,174 lone units, 8 wearers. Two-roll
abilities, priced as one roll: 7.49, 34 procs on 109 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 716 isolated hits. Dropping the swapped stretches (8 active minutes, 1% of hits, in 1 run) moved the
read from 2.93 to 2.95. 1 run logged in two reports counts once. The paladin check, which sets no rate: lone white
swings 4.63, 39 procs on 202 lone units, 4 wearers. By report date: before 2026-11-18 2.91 (927 procs); from 2026-11-18
to push 114313 2.98 (1,353 procs). Reports from 2026-11-09 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       7 | 1,529 | 12,603 | 2.91 |
| Rogue   |       5 |   744 |  5,893 | 3.03 |
| Shaman  |       1 |     7 |     43 | 3.91 |

Reports and fights: 1qcv4MKB7R9bWxJZ 1-9; 6b4F7HrzNTnxRMXd 1-13; 84VDXdcKZhHMPBAm 1-5; 8G3bmgKTvBxqfJQL 4-20, 22-25;
b4hn6GgpFLtJ1N2X 1-13; bLmVHN7BaP8DvX9Z 1-16; C8z3gcnBYGZQXJv2 1, 3-6; GmjR9hMTvBDtkNQ6 1-4, 7; hLn1YZ63qHVx49TF 1-6;
hYHt6kLTgApm8RMw 1-2; m8RLWaY7bh9jHKwD 2; qCgr4VRyj8HcaznF 1-2; qYC7g4Har6pKz8QT 1; RaP2LJFGrWzZtkf4 1-2;
rTYJ6yWfNcbKXztR 1; Ztv8xkT6dmQ7n91C 1.

## Wraith Scythe (item 11920, spell 16414): 3.45, measured

1,136 procs on 8,989 rolling hits, 8 wearers, 8 runs, in the main hand: all hits 3.45 [3.17, 3.73]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.09 [2.75, 3.46], 218 procs on
1,922 lone units, 5 wearers. Dual-wield white: 2.54 [1.93, 2.80], 40 procs on 430 lone units, 3 wearers. Specials: 3.81
[2.88, 4.25], 141 procs on 1,008 lone units, 7 wearers. Two-roll abilities, priced as one roll: 4.70, 5 procs on 29 lone
units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 286 isolated hits. Dropping the swapped
stretches (19 active minutes, 7% of hits, in 1 run) moved the read from 3.22 to 3.45. By report date: before 2026-11-18
3.57 (443 procs); from 2026-11-18 to push 114313 3.37 (693 procs). Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   849 | 6,899 | 3.36 |
| Rogue   |       1 |   247 | 1,760 | 3.83 |
| Hunter  |       1 |    40 |   330 | 3.30 |

Reports and fights: B6JLrnWKF9htyjv2 1; DZjKqdLXkBztYJTC 1; p1VPMyG8jgR6Ymc7 1-4; PtKXrwfj8zBHm2Gy 7-8; PtmYHjkDrqhdR9MA
1-6, 9-10; XPgHryZdABkRDhwV 3; YnVFyQq3w28mr41R 1; ZQtv6AqYmdJFWwhy 7-8.

## X'caliboar (item 10758, spell 1293787): 1.8, measured

715 procs on 7,174 rolling hits, 8 wearers, 9 runs, in the main hand: all hits 1.81 [1.73, 1.97]. The rate is all hits,
which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 1.90 [1.72, 2.29], 113
procs on 1,079 lone units, 6 wearers. Specials: 1.61 [1.06, 2.01], 48 procs on 543 lone units, 6 wearers. Two-roll
abilities, priced as one roll: 4.17, 22 procs on 96 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 146 isolated hits. Dropping the swapped stretches (34 active minutes, 16% of hits, in 1 run) moved the
read from 1.52 to 1.81. The paladin check, which sets no rate: lone white swings 2.20, 195 procs on 1,610 lone units, 21
wearers. By report date: before 2026-11-18 1.81 (273 procs); from 2026-11-18 to push 114313 1.81 (442 procs). Reports
from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   710 | 7,130 | 1.81 |
| Hunter  |       2 |     5 |    44 | 2.07 |

Reports and fights: 1WHqx7bPa8mkGydJ 1, 3; aPDTGRxFmMhZAzYW 1; HRnmMtD7Fp41wb2g 1; K8rhAtLjcz7gdHkn 6-7;
maYP7WvxJ2TzGtDH 1-3; mM3pZD8H7CqBNYwJ 1; P1yKzX3Q8NgFYxZH 1-3; WYZRJjhp7LvDqKNB 1-7; X4jWcyZhzfmJKG9F 1.

## The Ziggler (item 8006, spell 13482): 3.6, measured

3,916 procs on 38,305 rolling hits, 49 wearers, 50 runs, in the both hands or the main hand or the off hand: all hits
3.61 [3.48, 3.78]. The rate is all hits, which the specials rate agrees with. Dual-wield white: 3.17 [2.94, 3.35], 1,178
procs on 13,126 lone units, 48 wearers. Specials: 3.91 [3.68, 4.13], 147 procs on 1,327 lone units, 21 wearers. Two-roll
abilities, priced as one roll: 11.76, 3 procs on 9 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 817 isolated hits. Dropping the swapped stretches (22 active minutes, 1% of hits, in 2 runs) moved the
read from 3.56 to 3.61. 1 run logged in two reports counts once. By report date: before 2026-11-18 3.57 (2,489 procs);
from 2026-11-18 to push 114313 3.67 (1,427 procs). Reports from 2026-11-13 to 2026-11-20. The sim does not simulate this
proc yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      41 | 3,502 | 33,940 | 3.64 |
| Hunter  |       5 |   140 |  1,363 | 3.63 |
| Warrior |       3 |   274 |  3,002 | 3.22 |

Reports and fights: 2bYvT8kWxjFfdwrm 1-2; 2K4Z1jLyGWzdnD3f 10, 12; 3DQy7BhjH1mnZPpq 9-12; 3HwJAWkhFz7gCaQy 8-14;
3QgzWpJKCNMRvAYT 3-8, 10; 4xnDCWwrGKAFdNHv 1; 74GT2KcnWCLq6Bby 1-2; 7dZt6b9Kh4F2CNAn 1-12; 8YzmGj6fAZqMBNtx 1-7;
9kRzG7P6AKj2pdaF 1-2; 9MmVG8crjzJZ24K7 1-4; 9WaCtMzNbFcr4qmv 1-4; A2FMrK3Y1xkvafQH 1-7; adJRDVZ1mhYP8vj2 1-4;
bLmVHN7BaP8DvX9Z 1-16; BmfFgAczHXN7b16M 9-24; BTRMkxmvfrwcZNaG 1-3; fCJMF2rZqAHjBw67 1-5; FR9jJ7b31XCkLmtQ 5-7;
FxndQv9RDXt47KA8 1-2; FzCJvnHfGyMLWVwb 1-11; gLcGzRyxhZ2akWjw 1; GyQ6cqtCxPXHaDFz 22; HAZ4a9BmGxyD8gV3 1-3;
kvZgPDVybd6YN4Hw 1-2; LH4aGRZt7wVhkpcD 5-21; mjbfcTLkM2PGKadw 4-7; mwvHGaXY3KJ1xjdh 1-7; nwh3zdqc4CjXa6MG 1-7, 9-10,
12-23, 25; NyC7zw4Pm8aBHbZ2 10-20; pmKDwngMT2NFktdA 8; Pn8QfjmVL46tWpXR 1-9; pZNqCry3wRxmhk2L 1-16; R9njq1AGDvmp36bM
1-8; RPnVYQcG9xTJh1ar 1-4; V4HkPzgr21bxRvFt 1-3; vcxFpTb6zP2WrVAj 1-8; vkM4HWNYydVbPGxf 1-4; VYBNbJx7G8kP6vXd 10-15;
x3PftMnXCBZgLAG2 1, 4-9; x7mkBJZvX8PnKcRf 1; xM9bRGyajmBFvCqt 1; y8wDKJvdL1Ncp9za 1-3; yb6BrpYZnzHajqmx 1;
yhVbHg7vzDkKfnRY 1-7; YTn9F1x6jLqN2fdQ 1-9; ZbY9WHdayD6FGhJ7 3; zfaGW68H1jp9LRVm 5-14; zhkQZTaX8cVg47Jv 1-5, 7-8;
ZkpTbqwFYMAaLP4r 1, 3-6.

# EV-0183: Proc rates of fifteen weapons from the second 2026-11-22 sweep

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it; Mug O' Hurt's rate is measured, and the others observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the 15 weapons; 193 reports in all for them, of the count's 3,486
- Game build: 1.60.1.70790 to 1.60.1.71037 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC,
  71037 from 2026-11-20 18:23 UTC); 705 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 55
  of them for these weapons. The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

## Method

The count is EV-0153's Method, run again over every wearer stream on disk for the 90 items it rates, the ones here and
those of EV-0120, EV-0121, EV-0125, EV-0133, EV-0136, EV-0146, EV-0150, EV-0153, EV-0157, EV-0177 and EV-0178, with the
same classes and forms left out, intervals, lone-unit checks, swap drop, credit windows and calls of measured or
observed. Paladins are left out of every rate. From hotfix push 114313 (2026-11-20 23:47 UTC) each Mutilate strike rolls
its own hand once; before it Mutilate is left out with its procs, as the two-roll abilities are. From 2026-11-20 a Shank
rolls the off-hand weapon once (EV-0155 §Shank from 2026-11-20), and the count takes it as an off-hand roll. A lone-unit
rate from fewer than 5 wearers takes a Poisson interval on its procs.

Rolls and procs follow EV-0155 with these rules. A debuff-only landing (Sunder Armor, Rend's landing, Cheap Shot, a
sting) rolls a weapon proc once, and an equip proc only where its own lone landings proc at half the white chance or
more. Sap and Death Mark landings roll nothing. Shield Slam and Shield Bash roll equip procs, never weapon procs. A proc
that lands on two targets is two procs. The hits within the window of a killing blow or a two-roll ability leave with
its procs (800 ms for a missile proc). A missile strike that lands late rolls its weapon again, and the sim adds that
chain itself, so all hits leave its chain procs out. A missile proc's own side effects do not break a lone unit. A dual
wielder with two weapons of one speed has its white swings labelled by damage size, not timing.

Scripts and inputs, in `forever-sweep-auto`: the run `runs/2026-11-22-2335-v16`, whose `results.json` (`analyze.py`
version 16, cache `cache_v16`) holds every item's count, rerun on the wearer streams of the 2026-11-22 23:35 sweep. The
rate is each result's `all_hits`, matched on item and `effect_spell` and rounded to the nearest 0.05 at full precision;
the result's `ppm`, a lone-unit rate where one has enough procs, sets no rate.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon                    |   Item | Rate | Call     | All hits [95%]    |   Procs / hits | Wearers | One-hand white | Specials  |
| ------------------------- | -----: | ---: | -------- | ----------------- | -------------: | ------: | -------------- | --------- |
| Bloodpike                 |  13057 |  0.9 | observed | 0.90 [0.77, 0.96] |    170 / 3,533 |      13 | 0.85 (57)      | 0.70 (30) |
| Darkmist Dirk             | 285102 | 3.65 | observed | 3.66 [3.18, 3.79] |    418 / 4,280 |       5 | -              | 3.59 (42) |
| Electrified Dagger        |  19100 | 1.25 | observed | 1.27 [1.17, 1.48] |    262 / 6,862 |       5 | -              | 1.05 (7)  |
| Flesh Piercer             |   3336 | 0.95 | observed | 0.95 [0.85, 1.04] |   519 / 16,313 |      20 | -              | 0.88 (65) |
| Frightskull Shaft         |  14531 |    3 | observed | 3.02 [2.93, 3.16] |  1,036 / 6,046 |       6 | 2.78 (254)     | -         |
| Gargoyle Shredder Talons  |  13399 | 2.05 | observed | 2.06 [1.82, 2.19] |    282 / 4,556 |       5 | -              | -         |
| Glacial Blade             |  19099 |  1.3 | observed | 1.29 [1.20, 1.44] |    299 / 7,720 |       6 | -              | 1.66 (11) |
| Hedgecutter               |  18498 | 1.75 | observed | 1.77 [1.64, 1.94] |    220 / 3,917 |       6 | 2.55 (30)      | 0.79 (2)  |
| Mug O' Hurt               |   4090 | 2.15 | measured | 2.13 [2.01, 2.23] |    590 / 9,774 |       9 | 1.90 (80)      | 1.72 (21) |
| Naraxis' Fang             |   4449 |  2.1 | observed | 2.11 [2.02, 2.31] |    357 / 6,337 |       8 | 2.68 (2)       | 3.07 (19) |
| Needletooth's Needletooth | 282703 |    4 | observed | 3.98 [3.86, 4.18] | 1,353 / 12,012 |      10 | -              | 2.41 (19) |
| Runic Darkblade           |   3822 |    2 | observed | 2.00 [1.68, 2.10] |    444 / 4,045 |      18 | 1.98 (119)     | 1.76 (38) |
| Searing Needle            |  12531 |  2.8 | observed | 2.80 [2.68, 2.94] |    594 / 7,059 |      13 | -              | 3.23 (6)  |
| The Unstoppable Force     |  19323 | 1.05 | observed | 1.06 [1.02, 1.12] | 1,470 / 23,174 |      18 | 1.14 (391)     | 0.00 (0)  |
| Venom Web Fang            |    899 |  1.1 | observed | 1.09 [0.96, 1.22] |    255 / 9,385 |      18 | 1.38 (2)       | 0.70 (11) |

## Bloodpike (item 13057, spell 18202): 0.9, observed

170 procs on 3,533 rolling hits, 13 wearers, 16 runs, in the main hand: all hits 0.90 [0.77, 0.96]. The rate is all
hits, which the one-hand white rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand white:
0.85 [0.50, 1.00], 57 procs on 1,256 lone units, 13 wearers. Specials: 0.70 [0.49, 1.20], 30 procs on 803 lone units, 12
wearers. Two-roll abilities, priced as one roll: 1.93, 7 procs on 68 lone units (Execute 2 on 27 and Mongoose Bite 5 on
41), left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms, 97 isolated hits. Dropping
the swapped stretches (7.7 active minutes, 4% of hits, in 1 run) moved the read from 0.86 to 0.90. Left out: 34 of 50
runs in Cat or Bear Form. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |      12 |   130 | 2,720 | 0.90 |
| Warrior |       1 |    40 |   813 | 0.92 |

Reports and fights: 3HFWCJzxwqVAgMmB 1-3; 4fdqGcQjnvYg839V 1-3; B3pPdXHwrJkbKzhx 1-2; BCfabHPZ2YkqK7c8 1-2;
FRnrthqDC29Gajgy 1-3; GHw4ZrQRTk9BczXF 6-12; L2CWacgHtKQ9NFRM 1-9; NxmwqXVA7JkpP8Qn 1-4; Tby1rLwFvkgRtxpn 15-27, 29-35;
aDhRj9cFYQLn6Kzk 1; bNadwz7k2rM1htLP 1; d1gNRY9bxLtmqQyJ 1-2; dT4RAM6tkfaGJmBP 1-5; h9Nv8qJwW3QbRYpD 1-2;
w1Mrf9TACycaXhvY 1-7; wdRgBVHMYcvX94fr 1-17.

## Darkmist Dirk (item 285102, spell 1317583): 3.65, observed

418 procs on 4,280 rolling hits, 5 wearers, 5 runs, in the main hand or the off hand: all hits 3.66 [3.18, 3.79].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; the interval is wider than 10% of the rate. One-hand white: none. Dual-wield
white, which sets no rate: 3.51 [2.96, 4.13], 143 procs on 1,529 lone units, 4 wearers (a Poisson interval: under 5
wearers). Specials: 3.59 [2.59, 4.85], 42 procs on 439 lone units, 2 wearers (a Poisson interval: under 5 wearers).
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 115 isolated hits. Reports from 2026-11-13 to
2026-11-20.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       3 |   397 | 4,054 | 3.67 |
| Hunter |       2 |    21 |   226 | 3.48 |

Reports and fights: NvrDh4t3a1zcyfBd 1-10; R8LnATFfcY4KZmv1 1-7; hf1pF42aPHMv9kLX 15-20, 22, 24-36; r1tJ39Xp7M8byN64
7-8, 10-11; wL9nrZqGN71QTcKk 1-3.

## Electrified Dagger (item 19100, spell 23592): 1.25, observed

262 procs on 6,862 rolling hits, 5 wearers, 7 runs, in the main hand or the off hand: all hits 1.27 [1.17, 1.48].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; every wearer is a rogue; the interval is wider than 10% of the rate. One-hand
white: none. Dual-wield white, which sets no rate: 1.42 [1.20, 1.74], 66 procs on 1,553 lone units, 5 wearers. Specials:
1.05 [0.42, 2.17], 7 procs on 222 lone units, 2 wearers (a Poisson interval: under 5 wearers). Direct: window [-10, 100]
ms, median 0 ms, 98th percentile 0 ms, 44 isolated hits. Dropping the swapped stretches (6.8 active minutes, 2% of hits,
in 1 run) moved the read from 1.24 to 1.27. Reports from 2026-11-17 to 2026-11-21.

| Class | Wearers | Procs |  Hits |  PPM |
| ----- | ------: | ----: | ----: | ---: |
| Rogue |       5 |   262 | 6,862 | 1.27 |

Reports and fights: B6JLrnWKF9htyjv2 1; JWwXvLm1C9TgGykB 2; Pz3XcdNk8xCRjaTA 78 fights from 11 to 97; QHqPwymAX4rLv6at
1-2; WqfyDhVHxT93Q7Pt 25-30; gRWbpVBwcJxPjKZ1 4; mr8DVQNXZ7CTJK6y 2-4.

## Flesh Piercer (item 3336, spell 18078): 0.95, observed

519 procs on 16,313 rolling hits, 20 wearers, 26 runs, in the main hand or the off hand: all hits 0.95 [0.85, 1.04]. The
rate is all hits, which the specials rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand
white: none. Dual-wield white, which sets no rate: 0.87 [0.77, 0.99], 161 procs on 5,581 lone units, 19 wearers.
Specials: 0.88 [0.53, 1.14], 65 procs on 2,217 lone units, 11 wearers; Sinister Strike gives 61 procs on 1,850 of them.
Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 147 isolated hits. Dropping the swapped stretches (59
active minutes, 9% of hits, in 4 runs) moved the read from 0.87 to 0.95. Reports from 2026-11-07 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      18 |   498 | 15,760 | 0.95 |
| Warrior |       1 |    20 |    535 | 1.12 |
| Hunter  |       1 |     1 |     17 | 1.73 |

Reports and fights: 12xHrDfyTcWRk9M4 1-2, 4-5; 1KtWbHDQdyZAwrfa 1-19; 1grJhFB7DP6Nd2Wj 1-8; 2PtRnKzqJDMvG6rX 1-2;
7kxNnWGw2FzfLJtR 5-8; 7qGfC4xYgJz91Qmw 1-6; 93ygXAq67nFGMxBh 1-3; AbBcj4dL7yXVzCft 1-3; B8APxj1fHGLRpz7k 1, 3;
Ccarqm7TZDg4kxJW 1-4; D2p7zZ86V4KBGgrM 1-10; DpGahT9jyNQW42BK 1-12; GcjZ39HkxPXBAq7v 1-2; QNT4bZ6HB71mdaAF 1-8;
RC9LfKAhdmV8J2nH 1-21; XGtYA7qMNQk39ZD1 1-7, 9-11; Z68xDjRV2H3JtWfq 1; b6QHjXT2GF7A1YPr 1-8; bKfP317hkVcdG6LH 1-3;
bRDCX4NtdJKaxWhH 1-2; cG82vnXYjAVaTN1z 1-4; d1ZzbXAgTBjLMNfY 1-16; dV1C9qMBN7bZkcna 1-2; n7fPc4YgR3jBKGqd 1-5;
vPwpdrLM7tn3f4Wh 1; x4hVmK9AH817JPbW 1-15.

## Frightskull Shaft (item 14531, spell 18633): 3, observed

1,036 procs on 6,046 rolling hits, 6 wearers, 6 runs, in the main hand: all hits 3.02 [2.93, 3.16]. The rate is all
hits, which the one-hand white rate agrees with. Observed: every wearer is a shaman. One-hand white: 2.78 [2.53, 3.08],
254 procs on 1,615 lone units, 6 wearers. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms, 217 isolated
hits. Dropping the swapped stretches (13 active minutes, 5% of hits, in 1 run) moved the read from 2.86 to 3.02. The
paladin check, which sets no rate: lone white swings 2.97, 33 procs on 196 lone units, 2 wearers. Left out: 1 of 7 runs
in Cat or Bear Form. Reports from 2026-11-16 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       6 | 1,036 | 6,046 | 3.02 |

Reports and fights: BCq3RcM6kYdFtGJj 1; BX8n2xTycaKr4CGL 3-15; MRJHKBWt2TCp17PG 1-2, 4-7; Nxznkwgt78HVyvPC 1-10;
WcCH7rnkQdaLZJ6j 1-3; XfrWGm96NY1BTMng 1-2.

## Gargoyle Shredder Talons (item 13399, spell 1320824): 2.05, observed

282 procs on 4,556 rolling hits, 5 wearers, 5 runs, in the off hand: all hits 2.06 [1.82, 2.19]. Neither the one-hand
white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no lone-unit rate
checks it; the interval is wider than 10% of the rate. One-hand white: none. Dual-wield white, which sets no rate: 1.99
[1.60, 2.45], 90 procs on 1,507 lone units, 4 wearers (a Poisson interval: under 5 wearers). Direct: window [-10, 100]
ms, median 0 ms, 98th percentile 0 ms, 61 isolated hits. Dropping the swapped stretches (37 active minutes, 18% of hits,
in 2 runs) moved the read from 1.69 to 2.06. Reports from 2026-11-16 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       3 |   144 | 2,391 | 2.01 |
| Warrior |       2 |   138 | 2,165 | 2.13 |

Reports and fights: Kjk8dnN3y7BDhYAg 20; Qr9PmCnptRWMjN7k 27-31; d1vn7jt4RH2ZmLAB 1-2, 4-9; fMLcHzyZvt7QTCmg 1-9;
ra2vD4ZnN79zkGbj 1.

## Glacial Blade (item 19099, spell 18398): 1.3, observed

299 procs on 7,720 rolling hits, 6 wearers, 7 runs, in the main hand or the off hand: all hits 1.29 [1.20, 1.44].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; every wearer with a proc is a rogue; the interval is wider than 10% of the rate.
One-hand white: none. Dual-wield white, which sets no rate: 1.05 [0.90, 1.23], 52 procs on 1,657 lone units, 5 wearers.
Specials: 1.66 [0.83, 2.97], 11 procs on 221 lone units, 3 wearers (a Poisson interval: under 5 wearers). Direct: window
[-10, 100] ms, median 0 ms, 98th percentile 1 ms, 35 isolated hits. Reports from 2026-11-20 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       5 |   299 | 7,716 | 1.29 |
| Hunter |       1 |     0 |     4 | 0.00 |

Reports and fights: BktCDhG4XAmf3Pc7 1-5, 8-12; GD26JWT9QRqhCyFg 3-36; Xjrt9FRMWqJyKdpa 2-3; bmNLMA1QdaPDkCp4 1-9;
n4q6fxmaLKB8wcXJ 1-2; wkrzY3mDh7RGfMc2 54; z1dafpqhmA8LgTB7 1.

## Hedgecutter (item 18498, spell 1302344): 1.75, observed

220 procs on 3,917 rolling hits, 6 wearers, 6 runs, in the main hand or the off hand: all hits 1.77 [1.64, 1.94].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 2.55 [1.72, 3.65], 30 procs on 371 lone units, 2 wearers (a
Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.57 [1.20, 2.00], 63 procs on 1,270 lone
units, 4 wearers (a Poisson interval: under 5 wearers). Specials: 0.79 [0.10, 2.85], 2 procs on 80 lone units, 2 wearers
(a Poisson interval: under 5 wearers). Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 59 isolated
hits. Dropping the swapped stretches (17 active minutes, 13% of hits, in 1 run) moved the read from 1.55 to 1.77.
Reports from 2026-11-16 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       3 |   144 | 2,605 | 1.75 |
| Warrior |       3 |    76 | 1,312 | 1.83 |

Reports and fights: DcykPCwKQNzq7btT 4; TG7ZMLJB1cYAhq4k 1-3, 5-6; WRta3xn1PLJqj2rp 3; XWbYtZ6mTHND2G3P 1-3;
gdBnpzCGaJ2THt3y 1; qFnZWz3fpbdBVQrg 1-2.

## Mug O' Hurt (item 4090, spell 13496): 2.15, measured

590 procs on 9,774 rolling hits, 9 wearers, 10 runs, in the main hand or the off hand: all hits 2.13 [2.01, 2.23]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 1.90 [1.51, 2.36], 80 procs on 1,486 lone
units, 2 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 2.04 [1.85, 2.19], 136
procs on 2,350 lone units, 7 wearers. Specials: 1.72 [1.06, 2.62], 21 procs on 432 lone units, 2 wearers (a Poisson
interval: under 5 wearers). Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 147 isolated hits. Dropping
the swapped stretches (16 active minutes, 3% of hits, in 2 runs) moved the read from 2.06 to 2.13. Reports from
2026-11-15 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   299 | 4,851 | 2.18 |
| Rogue   |       6 |   291 | 4,923 | 2.09 |

Reports and fights: Bmw1jxaLbfzRF2WP 1-3; JwCKxjf984AZ2QdM 1-19; Lz2fJ1CtaPgAk49W 1; Vxvw2gM8Cz4TAZFK 14-35, 37-38,
40-41; ZRtnF29rMkgVydmJ 1; g7LxM62qdYmVJhDA 1-16; krA2h3ZCfBXMmT4F 1-8, 10; n8qwgXyvNbFhJxrK 1-5; rp72ZYxmG68tMJ3Q 1;
wRWZtcVCk3QJKDa6 1, 3.

## Naraxis' Fang (item 4449, spell 16400): 2.1, observed

357 procs on 6,337 rolling hits, 8 wearers, 10 runs, in the main hand or the off hand: all hits 2.11 [2.02, 2.31].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 2.68 [0.32, 9.68], 2 procs on 28 lone units, 1 wearer (a Poisson
interval: under 5 wearers). Dual-wield white, which sets no rate: 2.12 [1.89, 2.32], 165 procs on 2,920 lone units, 7
wearers. Specials: 3.07 [1.85, 4.80], 19 procs on 232 lone units, 3 wearers (a Poisson interval: under 5 wearers).
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 124 isolated hits. Reports from 2026-11-13 to
2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       6 |   352 | 6,265 | 2.11 |
| Hunter |       2 |     5 |    71 | 2.63 |

Reports and fights: 2m7hJVxgNdM8DTyb 1-3; 7A3kJP2ncwj89hVx 1; 9DpJ4jLhmMrqNK38 1-3; KbhdJBtAm3McFpZr 1-8;
NDHgq2hvyaTMfBm7 1-15; NFZkg9qWJhxTDPBM 1-12; abmBWK2ZRrVvtXch 1-6; akfrAm74NYHVqKgP 1; bPqj8zkJN3YatpwD 1-2;
ydb2wTpBFPgfWrCm 1-3.

## Needletooth's Needletooth (item 282703, spell 1314305): 4, observed

1,353 procs on 12,012 rolling hits, 10 wearers, 15 runs, in the main hand or the off hand: all hits 3.98 [3.86, 4.18].
The count flags its travel window as maybe cut short (its 98th-percentile delay, 250 ms, is within 17% of the 300 ms
isolation); on hits isolated 700 ms Puncture lands up to a 98th percentile of 671 ms after its hit (`MEASUREMENT.md`),
past the 400 ms a lone-unit rate allows, so none checks the all-hits rate. Observed: no lone-unit rate checks it; every
wearer is a rogue. One-hand white: none. Dual-wield white, which sets no rate: 3.35 [2.83, 3.74], 248 procs on 2,611
lone units, 10 wearers. Specials: 2.41 [0.98, 3.42], 19 procs on 278 lone units, 6 wearers. Travel time: the proc lands
a median 134 ms after its hit, 98th percentile 250 ms, window [-9, 260] ms (327 isolated hits). Dropping the swapped
stretches (20 active minutes, 6% of hits, in 1 run) moved the read from 3.73 to 3.98. All hits leave out 47 chain procs:
a Puncture that lands late rolls the dagger again, and that proc lands with it on the same target within 2 ms, all on
main-hand wearers after a white swing. The sim adds the chain itself. With them all hits read 4.11 [3.93, 4.46], and the
main-hand runs 4.47. The 59 procs and 238 hits within 800 ms after a two-roll ability, Mutilate before push 114313 among
them, or a killing blow are left out; a 50 ms window had credited 51 procs after those Mutilates to other hits
(`MEASUREMENT.md`). Reports from 2026-11-13 to 2026-11-22.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |      10 | 1,353 | 12,012 | 3.98 |

Reports and fights: 6q3DL8zbTahFMYwR 1; 6wTrKxAH49QhkMyg 1-10; 9dr8fLZc7PkH6KNw 1; Atbk8RdhcYHZK2Lx 1-2, 4-7;
BAKxn2GMyVhdXFN9 1; C9tYRWgh3yrvZqQT 1-3, 5-13; CwXy1rQ6pHj2FfbW 1-10; DGxVHRvhb4tpgwf8 1-19; JqGf1CNAZhm4bMrW 1;
MQN29wtPzWYkKm6h 1-4; Ncg28vrtLzkDBw93 1-10; YTtFk7y8b6d92Gw4 1-5; haDR7cLrFYPZK1QT 1-14; pqR7YdvjkwXZyhC1 1;
vPwpdrLM7tn3f4Wh 1.

## Runic Darkblade (item 3822, spell 16409): 2, observed

444 procs on 4,045 rolling hits, 18 wearers, 22 runs, in the main hand: all hits 2.00 [1.68, 2.10]. The rate is all
hits, which the one-hand white rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand white:
1.98 [1.64, 2.10], 119 procs on 1,093 lone units, 18 wearers. Specials: 1.76 [0.80, 2.31], 38 procs on 392 lone units,
12 wearers. Two-roll abilities, priced as one roll: 1.27, 10 procs on 143 lone units (Execute 9 on 128, Mocking Blow 0
on 3 and Mongoose Bite 1 on 12), left out of the rate. Travel time: the proc lands a median 151 ms after its hit, 98th
percentile 338 ms, window [24, 348] ms (74 isolated hits). Dropping the swapped stretches (23 active minutes, 9% of
hits, in 2 runs) moved the read from 1.82 to 2.00. The paladin check, which sets no rate: lone white swings 1.93, 64
procs on 604 lone units, 11 wearers. Reports from 2026-11-10 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   410 | 3,605 | 2.07 |
| Hunter  |      13 |    34 |   440 | 1.40 |

Reports and fights: 1bxfCJMQhwnra8yF 1-2; 2AkBjF16x3rwVLXC 1; 2v3c1tVwAmWCpQJF 1-3; 3vVtMWRcg9H4FTPw 1; AkRKf9tvFmL4xXCj
1-2; CH2dG7zJjqtRWTLr 1, 3-12; DLjAJ7VfXna8YtgZ 1-20; GJpYk7V1CyaKNPLb 1; H1kAdBR79zM4TChG 1; JQH2y6NWBXchGRbg 1;
Lt6jZNqM4CBGv9Fh 1-2; QFxf2YycX4K1G67q 1; RVTNDtGcvLjFzqnX 1-4; bpM8VcnfdmyxZ1vg 1-7, 9-17, 19-20, 22-23;
gBqYT9Da3KkpXHQj 1-4; gnZqNkAG9JDT2pQW 1; gqZDft46RKC2NLF7 11-12; h8gm2HFjGaZN4kKQ 1; nQf7GH3av6mDA4wT 1, 3-11;
qrKPAtNvzfGpC7JR 1, 3-6; tv1PgkxVBcK74zM6 1-3; wATW3ryJnMP2BGjR 1-16.

## Searing Needle (item 12531, spell 16454): 2.8, observed

594 procs on 7,059 rolling hits, 13 wearers, 14 runs, in the main hand or the off hand: all hits 2.80 [2.68, 2.94].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white, which sets no rate: 2.74 [2.48, 3.06],
226 procs on 2,746 lone units, 11 wearers. Specials: 3.23 [1.18, 7.02], 6 procs on 62 lone units, 4 wearers (a Poisson
interval: under 5 wearers). Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 146 isolated hits. Dropping
the swapped stretches (35 active minutes, 15% of hits, in 2 runs) moved the read from 2.38 to 2.80. Reports from
2026-11-15 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   377 | 4,415 | 2.85 |
| Hunter  |       3 |   110 | 1,365 | 2.69 |
| Warrior |       1 |   107 | 1,279 | 2.79 |

Reports and fights: 2AvFk4PJmBhyb89n 1; 2FXvGj386wa1ptHA 1-6, 8-12; 3ZnAfaWCBgJhxbVY 8; 4D9aqWFLJVvn7bYx 1-2;
9HD3NG4X8fALbgCy 1-5; GZxDpHnRMJ2YdyqW 1; JfwkMdTjLaDtVpXz 1-2; PHDt27hA391NBRXM 1, 3; QFRazgkZ6DHY178w 1-5;
RXPaYhrcHBQfv6m4 1-4; X47TDyBdAh2nQ6LK 1-10; ZzL7AbmQjJkrHdGn 1; kv32zprHjRfnZPh9 4-17; zPAaF3j9RMWKfJ4d 1.

## The Unstoppable Force (item 19323, spell 23454): 1.05, observed

1,470 procs on 23,174 rolling hits, 18 wearers, 25 runs, in the main hand: all hits 1.06 [1.02, 1.12]. The rate is all
hits, which the one-hand white rate agrees with. Observed: every wearer is a shaman. One-hand white: 1.14 [1.08, 1.22],
391 procs on 5,740 lone units, 18 wearers. Specials: 0.00, 0 procs on 6 lone units, 3 wearers. Direct: window [-11, 100]
ms, median 0 ms, 98th percentile 1 ms, 304 isolated hits. Left out: 15 of 40 runs in Cat or Bear Form. Reports from
2026-11-17 to 2026-11-22.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Shaman |      18 | 1,470 | 23,174 | 1.06 |

Reports and fights: 1YNQGtxavZV8bn3f 1, 3-4; 2mXywBCLNJHRW1xc 47-51; 3LTJqR7mjXtGxA2B 33; 3TdF61x7gMQ28pKX 4-35;
7RzFd1VfWgjZ3wML 1-8, 10-11, 13, 15-16; BfYZrMW4xdLvk36F 2-5; G29y3QBWV7XDR6mj 1-11; GXR7nt6ZCL4DQzxP 1;
Grm31RaDY9H6JLz2 1; HDFXxRcPGpdn73j6 1-8; LK46AC1ytPNQm9cD 1; LW1pRgPwCmXBqk9a 1-11, 13-27; RkqGcKZ9TBv1NgpA 1;
RyZmJKC4D8ztQLxY 1; Tdva3pHk4tmCfjZK 1-2; ZfXrLy2MdY6kzah8 2-7; ZqtpzHLbAQKWMcwN 1-3; bGTptRMWJyg4N7HQ 1-2;
fjPTxNBrbkJ6v3Q1 14-15; gnBdKW9myaxzNP3b 1; mTrJ2BgzPbDf4vLN 2-5; mzW48RckNaC1Tgpn 1-3; nvLk92CcVYAwxtaB 1-4;
qbtZdThvBCXVcfAW 1-6; zM4Dd8QKwBATnJR9 1-33.

## Venom Web Fang (item 899, spell 18077): 1.1, observed

255 procs on 9,385 rolling hits, 18 wearers, 20 runs, in the main hand or the off hand: all hits 1.09 [0.96, 1.22].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; the interval is wider than 10% of the rate. One-hand white: 1.38 [0.17, 4.98], 2
procs on 58 lone units, 1 wearer (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.22
[1.03, 1.43], 172 procs on 5,638 lone units, 17 wearers. Specials: 0.70 [0.16, 1.03], 11 procs on 625 lone units, 10
wearers. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 126 isolated hits. Dropping the swapped
stretches (55 active minutes, 10% of hits, in 2 runs) moved the read from 0.97 to 1.09. Reports from 2026-11-08 to
2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |      15 |   249 | 9,194 | 1.08 |
| Hunter |       3 |     6 |   191 | 1.25 |

Reports and fights: 1q8H9N7AM4ztBLZa 1-3; 4y2tDjL3wXdGTVaM 1-4; 9PjzqQhvgkA7T6K3 1-9; BNDKL8gfy3wXpjHV 1-4;
BYWx47AydwZcrDhz 1-3; CxbNBR6zJFpH3nYv 1-3; WjnNbmcqYtkpK3TL 1-49; XTH4hjrCaFymLfV9 1-3; a1zJ9xNVBy7trP23 1;
ah4jpXRfKY7FrPdG 1-5; b1ZxqrmHWn48MYQ9 1-3; cfBNnYH4Xp18F3bk 1-15; dfpkK4jbyY97LwxP 1; mrC9TwxYgb8H4F1d 1-4;
qGzkgDbAXjpc8Qad 1-5; rFcbq4XAamKhZ8gk 1-3; tbJx8pVKDrMBCYaP 1-4; wrx8NJY1tZKFgkyq 2-6; xMB7vGKDqz6knCW3 1-5;
zD4VF1278AymkKXY 4-5.

# EV-0186: Proc rates of three weapons from the 2026-11-24 sweep

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it; Winter's Bite's rate is measured, and Ragehammer's and Whipfang's Skinsearer's observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the three weapons; 25 reports in all for them, of the count's 36
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 20 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 13 of them for these weapons, and
  none start after 2026-11-24 01:00 UTC. The sim's database is 1.60.1.71037, push 114366; the hotfix each player ran is
  not recorded
- Date: 2026-11-06 to 2026-11-23
- Pull request: #1198

## Method

The count is EV-0183 §Method, run again by the automated sweep of 2026-11-24 14:50 over every wearer stream on disk for
four weapons: the three here and Hedgecutter, whose home is EV-0183. Analyzer version 19 differs from version 18 only on
melee-table strike procs and on Torturing Poker, and touches none of these weapons. Paladins are left out of every rate
(owner ruling, 2026-11-19), and their lone white swings of one-hand wearers are a check that sets no rate; three
paladins wear Ragehammer, two Hedgecutter and two Winter's Bite in the report index, and the run gives a paladin check
on Ragehammer alone. Every report starts before 2026-11-24 01:00 UTC, the sweep's cut for the 2026-11-23 hotfix ("Fixed
an issue where a few spells and abilities were causing more weapon procs than intended"); one Hedgecutter report,
dCzLqgx2cBH31Jwb, runs on to 07:44 UTC.

Scripts and inputs, in `forever-sweep-auto`: the run `runs/2026-11-24-1450`, whose `results.json` (`analyze.py`
version 19, cache `cache_v18`) holds each weapon's count. The rate is each result's `all_hits`, matched on item and
`effect_spell` and rounded to the nearest 0.05 at full precision; the result's `ppm`, a lone-unit rate where one has
enough procs, sets no rate. A lone-unit rate from fewer than 5 wearers takes a Poisson interval on its procs.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon                |   Item | Rate | Call     | All hits [95%]    | Procs / hits | Wearers | One-hand white | Specials  |
| --------------------- | -----: | ---: | -------- | ----------------- | -----------: | ------: | -------------- | --------- |
| Ragehammer            |  10626 |    2 | observed | 2.02 [1.87, 2.16] |  536 / 4,311 |       5 | 2.01 (161)     | 3.56 (9)  |
| Whipfang's Skinsearer | 284386 | 0.75 | observed | 0.76 [0.53, 0.91] |  106 / 4,932 |      13 | 0.91 (33)      | 0.79 (16) |
| Winter's Bite         |  10623 | 2.95 | measured | 2.94 [2.69, 3.21] |  609 / 5,915 |       7 | 3.04 (65)      | 2.72 (42) |

## Ragehammer (item 10626, spell 12686): 2, observed

536 procs on 4,311 rolling hits, 5 wearers, 5 runs, in the main hand: all hits 2.02 [1.87, 2.16]. The rate is all hits,
which the one-hand white rate agrees with. Observed: every wearer is a shaman. One-hand white: 2.01 [1.70, 2.28], 161
procs on 1,301 lone units, 5 wearers. Specials: 3.56 [2.12, 5.72], 9 procs on 41 lone units, 5 wearers, every one a
Stormstrike. The paladin check, which sets no rate: lone white swings 1.93 [1.42, 2.57], 47 procs on 394 lone units, 3
wearers (a Poisson interval: under 5 wearers). Direct: window [-20, 100] ms, median 0 ms, 98th percentile 0 ms, 147
isolated hits. Left out: 2 of 7 runs in Cat or Bear Form. Reports from 2026-11-16 to 2026-11-23.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       5 |   536 | 4,311 | 2.02 |

Reports and fights: AtdvyFrPCWzMTKp2 1-3; P6RGB3A9FxNVQrkw 1-8; QN8pLwHmn19zCPkA 1-5, 7-8; mayrdRq4hwZBjCQN 1-6;
pr3mcRDhzXQHwkx1 1-7.

## Whipfang's Skinsearer (item 284386, spell 1316039): 0.75, observed

106 procs on 4,932 rolling hits, 13 wearers, 13 runs, in the main hand or the off hand: all hits 0.76 [0.53, 0.91].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; the interval is wider than 10% of the rate. One-hand white: 0.91 [0.63, 1.28], 33
procs on 1,277 lone units, 3 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 0.62
[0.41, 0.92], 20 procs on 1,136 lone units, 9 wearers. Specials: 0.79 [0.32, 1.03], 16 procs on 714 lone units, 7
wearers. Direct: window [6, 100] ms, median 33 ms, 98th percentile 38 ms, 53 isolated hits. Dropping the swapped
stretches (24 active minutes, 14% of hits, in 2 runs) moved the read from 0.65 to 0.76. Reports from 2026-11-06 to
2026-11-23.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       4 |    65 | 2,770 | 0.83 |
| Rogue   |       5 |    37 | 1,964 | 0.66 |
| Hunter  |       4 |     4 |   198 | 0.71 |

Reports and fights: 2mNyRawv7fGMDYVx 1; 72gPMatGLpwd93jW 1-12; BxwQZPMADJf3bYF9 1-3; Cr4Y2BKd9tmRQVZD 3;
HjPpLZkTBmdMVWD6 1-2; M4THD3Qq9dBpbKvg 1-2; WhwcvTCZFPDxanyr 22-65; ZVkWt4Fh9McBGy3x 1-7; hAGwPapbQFyVjKWk 1, 3-6;
hRVgWcLv7fAt4MFP 1; hcPvrAHMG69VJxBQ 1-2; yPwYXVdkJb6Km14z 1-6; yqjrvnhW8YNVdg4A 1-18.

## Winter's Bite (item 10623, spell 13439): 2.95, measured

609 procs on 5,915 rolling hits, 7 wearers, 7 runs, in the main hand: all hits 2.94 [2.69, 3.21]. The rate is all hits,
which the one-hand white rate agrees with. One-hand white: 3.04 [2.35, 3.88], 65 procs on 610 lone units, 4 wearers (a
Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 2.25 [1.60, 3.08], 39 procs on 495 lone units,
3 wearers (a Poisson interval: under 5 wearers). Specials: 2.72 [2.41, 3.09], 42 procs on 441 lone units, 7 wearers.
Two-roll abilities, priced as one roll: 2.04, 1 proc on 14 lone units (Execute), left out of the rate. Travel time: the
proc lands a median 109 ms after its hit, 98th percentile 282 ms, window [-20, 292] ms (90 isolated hits); lone units
are isolated on the hit's own target. Dropping the swapped stretches (19 active minutes, 10% of hits, in 1 run) moved
the read from 2.65 to 2.94. Reports from 2026-11-19 to 2026-11-23.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   505 | 4,708 | 3.06 |
| Rogue   |       1 |   100 | 1,158 | 2.47 |
| Hunter  |       1 |     4 |    49 | 2.34 |

Reports and fights: 6rBtPT2AZaFv3J4H 1; 6zHgax7vFKqLrdmZ 1-2, 4, 6-8; 7q6NKLWFTtjVv4Xw 1; M7fnK21qrx3FBhGZ 1-2;
WJh8kcgXT2CnfNPz 1; XRaybqFnGfpk6HVv 1; tBk9vLmJgqwX84ar 1-7.

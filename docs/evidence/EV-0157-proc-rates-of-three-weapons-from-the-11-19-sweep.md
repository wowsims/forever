# EV-0157: Proc rate of Blackvenom Blade from the 2026-11-19 sweep

- Claim: Blackvenom Blade procs at the rate its section gives, priced on the weapon base speed of each hit that rolls it;
  the rate is observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the three weapons; 248 reports in all
- Game build: 1.60.1.70862 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114281; the hotfix
  each player ran is not recorded
- Date: 2026-11-09 to 2026-11-19
- Pull request: #1126

## Method

The count is EV-0153's Method, run again by the automated sweep of 2026-11-19 over every wearer stream on disk, with the
same rolling hits, classes and forms left out, intervals, lone-unit checks, swap drop, credit windows and calls of
measured or observed.

Scripts and inputs, in `forever-sweep-auto`: `analyze.py` (version 12, as EV-0153's, with a paladin check that sets no
rate) for the rates, and the run `runs/2026-11-19-2309`, whose `results.json` holds each weapon's counts.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs;
Was is the row's rate before this count.

| Weapon           | Item | Rate | Call     | All hits [95%]    | Procs / hits | Wearers | One-hand white | Specials  | Was |
| ---------------- | ---: | ---: | -------- | ----------------- | -----------: | ------: | -------------- | --------- | --: |
| Blackvenom Blade | 4446 |  4.2 | observed | 4.22 [3.95, 4.37] |  412 / 3,258 |       5 | -              | 2.99 (28) |   - |

## Blackvenom Blade (item 4446, spell 13518): 4.2, observed

412 procs on 3,258 rolling hits, 5 wearers, 5 runs, in the main hand or the off hand: all hits 4.22 [3.95, 4.37].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white: 4.14 [3.52, 4.57], 206 procs on 1,658
lone units, 5 wearers. Specials: 2.99 [2.90, 3.29], 28 procs on 312 lone units, 2 wearers. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 0 ms, 165 isolated hits. Reports from 2026-11-13 to 2026-11-19.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       4 |   367 | 2,875 | 4.26 |
| Hunter |       1 |    45 |   383 | 3.91 |

Reports and fights: APBKD3j6wYWG7xa4 8-11; FfAC3tWBxbp2HV9j 1-3; Ky8wY7p9bAcxTWvr 1-11; XBWhJzwTAYRPfQ4L 1-2, 4;
h8aHYk3QyJz2DNmt 1-2.

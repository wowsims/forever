# EV-0157: Proc rate of Blackvenom Blade from the 2026-11-19 sweep

- Claim: Blackvenom Blade procs at the rate its section gives, priced on the weapon base speed of each hit that rolls it;
  the rate is observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the three weapons; 248 reports in all
- Game build: 1.60.1.70862 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114281; the hotfix
  each player ran is not recorded
- Date: 2026-11-09 to 2026-11-19
- Pull request: #1126
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

## Method

The count is EV-0153's Method, run again by the automated sweep of 2026-11-19 over every wearer stream on disk, with the
same rolling hits, classes and forms left out, intervals, lone-unit checks, swap drop, credit windows and calls of
measured or observed.

Scripts and inputs, in `forever-sweep-auto`: `analyze.py` (version 12, as EV-0153's, with a paladin check that sets no
rate) for the rates, and the run `runs/2026-11-19-2309`, whose `results.json` holds each weapon's counts.

## Blackvenom Blade (item 4446, spell 13518): 3.9, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

489 procs on 4,203 rolling hits, 6 wearers, 6 runs, in the main hand or the off hand: all hits 3.88 [3.23, 4.33].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; the interval is wider than 10% of the rate. One-hand white: none. Dual-wield
white, which sets no rate: 3.97 [3.32, 4.46], 235 procs on 1,971 lone units, 6 wearers. Specials: 2.14 [1.47, 3.02], 32
procs on 498 lone units, 3 wearers (a Poisson interval: under 5 wearers). Direct: window [-10, 100] ms, median 0 ms,
98th percentile 0 ms, 187 isolated hits. Reports from 2026-11-13 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       5 |   444 | 3,822 | 3.87 |
| Hunter |       1 |    45 |   382 | 3.93 |

Reports and fights: APBKD3j6wYWG7xa4 8-11; CgwDT6Ptpd38v2Nr 1; FfAC3tWBxbp2HV9j 1-3; Ky8wY7p9bAcxTWvr 1-11;
XBWhJzwTAYRPfQ4L 1-2, 4; h8aHYk3QyJz2DNmt 1-2.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights its section lists, from every
  wearer stream on disk for the blade; 3,637 reports in all for the count, 6 of them for the blade
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 735 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 1 of them for the blade. The
  sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on the blade.

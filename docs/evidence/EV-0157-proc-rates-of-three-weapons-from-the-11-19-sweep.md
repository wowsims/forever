# EV-0157: Proc rate of Blackvenom Blade

- Claim: Blackvenom Blade procs Poison at 3.9 PPM, priced on the weapon base speed of each hit that rolls it; the rate
  is observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights listed below, from every wearer
  stream on disk for the blade; 6 reports in all
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 1 report starts after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is 1.60.1.71037, push
  114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

## Method

The count is EV-0183 §Method, run on this weapon.

## Blackvenom Blade (item 4446, spell 13518): 3.9, observed

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

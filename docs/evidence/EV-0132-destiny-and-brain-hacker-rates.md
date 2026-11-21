# EV-0132: Destiny's and Brain Hacker's proc rates

EV-0153 replaces the Destiny rate with the reports up to 2026-11-19, counted on the hits EV-0155 finds rolling a weapon
proc.

- Claim: Destiny (647, +200 Strength from 17152) procs about 1.2 times a minute, priced on the weapon's base speed and
  rolled by every landed melee hit, triggered ones such as Execute and Judgement of Command included. Brain Hacker
  (1263, 17148) procs 0.95 times a minute, priced on the weapon base speed of each hit that rolls it.
- Sources: Warcraft Logs, the fights from the first combatant info listing the weapon on the wearer to the last.
  Destiny: 1ZyhaVDdFtTYWMcC fights 10 to 12; xbJBZj14NWCXrhAm fights 5 to 12; VLz78BfZyWa3Cxr2 fights 1 to 9;
  vZqVmLYQa6jwP3A9 fights 2 to 16 and 18 to 21; Qb1yWxa8vHCGL7Yr fights 12 to 15; nX1qbatTzGZ3CAwB fights 1 and 3 to 7;
  zFXxNAn69HBGRaqg fights 1 to 10 (two wearers). Brain Hacker's section, which names its count, lists its own.
- Game build: 1.60.1.70938; the hotfix each player ran is not recorded. Brain Hacker's count gives its own
- Date: 2026-11-13 to 2026-11-16; Brain Hacker's count gives its own
- Pull request: #1059; Brain Hacker's count is #1156

## Method

This Method is Destiny's; Brain Hacker's is its count's. A proc is an application or refresh of Destiny's 17152 on the
wearer. A hit is a landed, non-periodic damage event of the wearer with melee defense: white swings, melee abilities,
and melee hits logged under a spell id the wearer logs no cast of (Execute's 20647, Judgement of Command, Seal of
Righteousness, Sweeping Strikes' 26654, Windfury Totem strikes). The rate per minute is the chance per hit times 60
over the weapon's base speed of 3.6 s, so haste raises procs a minute without moving it.

Each proc is credited to a hit by log order. Destiny's buff logs just before the damage of the hit that rolled it (497
of 654 credited procs 0 to 1 ms before, the rest within 20 ms), so a buff goes to the first hit within 20 ms after it,
else to the last hit within 20 ms before it (8 procs); 16 procs found no hit and count only in the first row.

## Destiny

| Hits                                              | Procs / hits | Rate                  |
| ------------------------------------------------- | ------------ | --------------------- |
| Every landed melee hit                            | 670 / 9440   | 1.18 PPM [1.10, 1.27] |
| White swings                                      | 258 / 3676   | 1.17 PPM [1.04, 1.31] |
| Melee hits under an id the wearer logs a cast of  | 189 / 3290   | 0.96 PPM [0.83, 1.10] |
| Melee hits under an id the wearer logs no cast of | 207 / 2474   | 1.39 PPM [1.22, 1.59] |
| Execute (warriors)                                | 59 / 505     | 11.7% a hit           |
| Judgement of Command (paladins)                   | 60 / 611     | 9.8% a hit            |

Per run, on every landed melee hit: 1.11, 1.31 and 1.19 for the three warrior runs, 1.18, 1.16, 0.94, 1.25 and 1.15
for the five paladin runs. Intervals are 95% Wilson on the hits. Melee abilities that hit several targets at once
(Whirlwind) read low per target, since one roll's credit goes to one of the targets' hits. The chance per kind of hit
differs more than its interval allows (Execute 11.7% against Seal of Righteousness 4.0%), which crediting by log order
can cause when hits share a millisecond; the pooled rate does not depend on which hit a proc goes to.

The chance per white swing shows no clear fall with haste, and the rate is priced on the base speed, as every weapon
rate is. Each landed white swing is binned by the base speed over the time since the previous swing (0.5 to 1.1 times
the base apart):

| Haste        | White swings | Procs | Chance | 95%         |
| ------------ | ------------ | ----- | ------ | ----------- |
| below 1.05   | 1342         | 99    | 7.38%  | 6.1 to 8.9  |
| 1.05 to 1.15 | 745          | 49    | 6.58%  | 5.0 to 8.6  |
| 1.15 to 1.30 | 347          | 24    | 6.92%  | 4.7 to 10.1 |
| 1.30 and up  | 350          | 24    | 6.86%  | 4.7 to 10.0 |

A rate on the base speed (1.17 PPM, log-likelihood -709.0) fits these swings slightly better than one on the hasted
interval (1.29 PPM, -710.4).

## Brain Hacker

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Rate: 0.95 PPM, measured

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

## Count of the 2026-11-06 to 2026-11-21 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 12 reports in all
- Game build: by report date, 1.60.1.70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC and 71037 from
  2026-11-20 18:23 UTC; no report starts after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is
  1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-16 to 2026-11-20
- Pull request: #1156

The count is EV-0153's Method on the hits EV-0155 finds rolling a weapon proc, run again over every wearer stream on
disk for its weapon, with more wearers fetched on 2026-11-21. The same rolling hits, classes and forms are left out:
paladins from every rate, druids in Cat or Bear Form, and the two-roll abilities with the procs around them. The
intervals, lone-unit checks, swap drop and credit windows are EV-0153's.

A rate is measured where a lone-unit rate of 50 or more procs (lone one-hand white swings or specials) overlaps the
all-hits interval, and 5 or more wearers of more than one class give an all-hits interval within 10% of the rate either
side. The rate is the all-hits rate rounded to the nearest 0.05. A lone-unit interval from under 5 wearers is a Poisson
interval. On Brain Hacker every class past the first is one wearer.

Each section's read by report date is the all-hits rate over the reports that start before 2026-11-18, from then to push
114313, and after it.

Scripts and inputs, in `forever-notices-1121`: `results.json`, the counts of `forever-sweep-auto/analyze.py` (version
12, as EV-0157's, with the paladin check that sets no rate) run with `--items items.txt --census census_used.json`;
`report.py`, the call per weapon (`rep_final.json`); and `dates.py`, the report dates (`dates_final.json`).

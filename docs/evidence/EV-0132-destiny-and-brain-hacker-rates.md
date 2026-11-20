# EV-0132: Destiny and Brain Hacker rates per landed melee hit

EV-0153 replaces the Destiny rate with the reports up to 2026-11-19, counted on the hits EV-0155 finds rolling a weapon
proc.

- Claim: Destiny (647, +200 Strength from 17152) procs about 1.2 times a minute and Brain Hacker (1263, 17148) about
  1 time a minute, each priced on the weapon's base speed and rolled by every landed melee hit, triggered ones such as
  Execute and Judgement of Command included.
- Sources: Warcraft Logs, the fights from the first combatant info listing the weapon on the wearer to the last.
  Destiny: 1ZyhaVDdFtTYWMcC fights 10 to 12; xbJBZj14NWCXrhAm fights 5 to 12; VLz78BfZyWa3Cxr2 fights 1 to 9;
  vZqVmLYQa6jwP3A9 fights 2 to 16 and 18 to 21; Qb1yWxa8vHCGL7Yr fights 12 to 15; nX1qbatTzGZ3CAwB fights 1 and 3 to 7;
  zFXxNAn69HBGRaqg fights 1 to 10 (two wearers). Brain Hacker: yHRFJ8kn9vQA4YLr fights 1 to 4.
- Game build: 1.60.1.70938; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-16
- Pull request: #1059

## Method

A proc is an application or refresh of Destiny's 17152 on the wearer, or a hit of Brain Hacker's 17148 (any hit type).
A hit is a landed, non-periodic damage event of the wearer with melee defense: white swings, melee abilities, and melee
hits logged under a spell id the wearer logs no cast of (Execute's 20647, Judgement of Command, Seal of Righteousness,
Sweeping Strikes' 26654, Windfury Totem strikes). The rate per minute is the chance per hit times 60 over the weapon's
base speed (3.6 s for Destiny, 3.5 s for Brain Hacker), so haste raises procs a minute without moving it.

Each proc is credited to a hit by log order. Destiny's buff logs just before the damage of the hit that rolled it (497
of 654 credited procs 0 to 1 ms before, the rest within 20 ms), so a buff goes to the first hit within 20 ms after it,
else to the last hit within 20 ms before it (8 procs); 16 procs found no hit and count only in the first row. Brain
Hacker's damage logs just after its hit, so it goes to the last hit within 20 ms before it. Brain Hacker's wearer also
swapped to Iceblade Hacker; hits and procs within 60 s of an Iceblade Hacker proc are left out (2 procs).

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

One paladin, 76 procs on 1,222 landed melee hits, Brain Hacker's own 17148 strikes left out: 6.22% a hit, 1.07 PPM
[0.86, 1.32]; white swings 0.94 PPM [0.64, 1.38], Judgement of Command 15 / 124 and Holy Strike 14 / 138.

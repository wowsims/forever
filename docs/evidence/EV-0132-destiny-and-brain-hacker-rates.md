# EV-0132: Destiny and Brain Hacker rates per landed melee hit

- Claim: Destiny (647, +200 Strength from 17152) procs about 1.2 times a minute and Brain Hacker (1263, 17148) about
  1 time a minute, each priced on the weapon's base speed and rolled by every landed melee hit, triggered ones such as
  Execute and Judgement of Command included.
- Sources: Warcraft Logs, every player wearing the weapon in the reports. Destiny: 1ZyhaVDdFtTYWMcC, xbJBZj14NWCXrhAm,
  VLz78BfZyWa3Cxr2, vZqVmLYQa6jwP3A9, Qb1yWxa8vHCGL7Yr, nX1qbatTzGZ3CAwB, zFXxNAn69HBGRaqg (two wearers). Brain Hacker:
  yHRFJ8kn9vQA4YLr. All fights.
- Game build: 1.60.1.70938; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-16
- Pull request: not yet opened

## Method

A proc is an application or refresh of Destiny's 17152 on the wearer, or a hit of Brain Hacker's 17148 (any hit type).
A hit is a landed, non-periodic damage event of the wearer with melee defense: white swings, melee abilities, and melee
hits the wearer did not cast itself (Execute, Sweeping Strikes, seals, Judgement of Command, Windfury Totem strikes).
The rate per minute is the chance per hit times 60 over the weapon's base speed (3.6 s for Destiny, 3.5 s for Brain
Hacker), so haste raises procs a minute without moving it.

Each proc is credited to a hit by log order. Destiny's buff logs just before the damage of the hit that rolled it (497
of 654 credited procs 0 to 1 ms before, the rest within 20 ms), so a buff goes to the first hit within 20 ms after it.
Brain Hacker's damage logs just after its hit, so it goes to the last hit within 20 ms before it. Brain Hacker's wearer
also swapped to Iceblade Hacker; hits and procs within 60 s of an Iceblade Hacker proc are left out (2 procs).

## Destiny

| Hits                               | Procs / hits | Rate                  |
| ---------------------------------- | ------------ | --------------------- |
| Every landed melee hit             | 670 / 9440   | 1.18 PPM [1.10, 1.27] |
| White swings                       | 258 / 3676   | 1.17 PPM [1.04, 1.31] |
| Melee abilities the wearer cast    | 189 / 3290   | 0.96 PPM [0.83, 1.10] |
| Melee hits the wearer did not cast | 207 / 2474   | 1.39 PPM [1.22, 1.59] |
| Execute (warriors)                 | 59 / 505     | 11.7% a hit           |
| Judgement of Command (paladins)    | 60 / 611     | 9.8% a hit            |

Per run, on every landed melee hit: 1.11, 1.31 and 1.19 for the three warrior runs, 1.18, 1.16, 0.94, 1.25 and 1.15
for the five paladin runs. Intervals are 95% Wilson on the hits. Melee abilities that hit several targets at once
(Whirlwind) read low per target, since one roll's credit goes to one of the targets' hits.

The chance per white swing does not fall with haste, so the rate is priced on the base speed, not the hasted interval.
Each landed white swing is binned by the base speed over the time since the previous swing (0.5 to 1.1 times the
base apart):

| Haste        | White swings | Procs | Chance | 95%         |
| ------------ | ------------ | ----- | ------ | ----------- |
| below 1.05   | 1342         | 99    | 7.38%  | 6.1 to 8.9  |
| 1.05 to 1.15 | 745          | 49    | 6.58%  | 5.0 to 8.6  |
| 1.15 to 1.30 | 347          | 24    | 6.92%  | 4.7 to 10.1 |
| 1.30 and up  | 350          | 24    | 6.86%  | 4.7 to 10.0 |

A rate on the base speed (1.17 PPM, log-likelihood -709.0) fits these swings better than one on the hasted interval
(1.29 PPM, -710.4).

## Brain Hacker

One paladin, 76 procs on 1,222 landed melee hits: 6.22% a hit, 1.07 PPM [0.86, 1.32]; white swings 0.94 PPM [0.64,
1.38], Judgement of Command 15 / 124 and Holy Strike 14 / 138.

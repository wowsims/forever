# EV-0133: Heart of Wyrmthalak procs per minute on melee and ranged hits

- Claim: Heart of Wyrmthalak (22321) casts Flame Lash 2 times a minute on landed melee and ranged hits, each hit on
  its own weapon's speed and Auto Shot and ranged specials on the ranged weapon's, with no cooldown; Volley and Shank
  never proc it, and its damage is tripled against orcs.
- Sources: Warcraft Logs, the 7 players whose every gear snapshot shows the trinket in 813rPaLtcdpHy67q,
  Hwq97xLh6gv8ZPrY, KPwXNByAG2xQm9ra, NXMZV9xjvWpB7GtP (two players), bYaT9KXtGwqvV6d4 and dvkzFtPH3WR9C4Km, and the
  one hunter of KA9rqdFTWzt1mpxN; all fights
- Game build: 1.60.1.70938, as in EV-0131 for reports from 2026-11-13 on; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-16
- Pull request: #1061

## Method

Flame Lash (27655) lands 20 to 50 ms after the hit that procs it, so a proc counts for the latest landed melee or
ranged hit of the wearer on the same target within 100 ms before it. A landed hit is a hit, crit, glancing or blocked
hit. Left out as hits: poison, weapon enchant, Seal and damage shield hits, ticks, and Volley and Shank, whose rows
suppress weapon procs. Druids, whose forms swing at their own speeds, and the paladin, whose Seal hits land with the
swing, are left out.

Each kind of hit is priced as procs per minute on the speed of the weapon behind it: a one-weapon white swing and a
melee special on the main hand's, a dual wielder's white swing on both hands' mean chance by their swing share, and
Auto Shot and ranged specials on the ranged weapon's. KA9rqdFTWzt1mpxN has no gear snapshot; its hunter's ranged
weapon speed, 3.3 s, is read off its Auto Shots, which come every 2.9 s with no haste buff up behind a 13 to 15%
quiver. Each rate carries a 95% likelihood interval from the hits and one from 2,000 draws of whole wearers.

## Rates

| Hits                     | Wearers | Procs / landed hits | Rate     | 95%, hits    | 95%, wearers |
| ------------------------ | ------- | ------------------- | -------- | ------------ | ------------ |
| White swings, one weapon | 3       | 7 / 85              | 1.64 PPM | 0.72 to 3.04 | 0.50 to 3.16 |
| White swings, dual wield | 3       | 392 / 5,448         | 2.10 PPM | 1.90 to 2.29 | 2.06 to 2.14 |
| Melee specials           | 6       | 205 / 2,425         | 1.86 PPM | 1.63 to 2.12 | 1.77 to 1.95 |
| Auto Shot                | 3       | 295 / 2,820         | 2.06 PPM | 1.84 to 2.28 | 1.92 to 2.49 |
| Ranged specials          | 3       | 35 / 481            | 1.60 PPM | 1.15 to 2.17 | 1.42 to 1.90 |
| Every kind               | 8       | 934 / 11,259        | 2.00 PPM | 1.88 to 2.12 |              |
| Volley                   | 2       | 0 / 2,941           | 0        |              |              |
| Shank                    | 4       | 0 / 593             | 0        |              |              |

A flat chance per hit fits worse: the best one, 8.3%, has a log likelihood of -3,219.3 against -3,201.7 for 2 procs
per minute. Auto Shot alone shows the speed: 7.05% of landed shots at a 2.2 s bow, 12.9% at 3.1 s and 11.25% at 3.3 s.

Two procs land as little as 9 ms apart, and 54 of the 929 gaps between procs are under a second, so it has no
cooldown. Flame Lash's normal hits deal 111 to 218 (the row's 112 to 168 and the wearers' damage bonuses), and about
300 hits of 335 to 669 are the tripled damage against orcs.

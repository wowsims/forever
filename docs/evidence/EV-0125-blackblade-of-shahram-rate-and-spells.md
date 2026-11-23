# EV-0125: Blackblade of Shahram's rate and what Shahram casts

- Claim: Blackblade of Shahram 12592 summons Shahram (16602) 1.0 times a minute on landed main-hand hits. Each
  Shahram casts exactly one of his six spells, starts its hardcast 1.03 to 1.83 s after the summon, and every copy of
  his Will or Blessing runs on its own timer.
- Sources: public reports on forever.warcraftlogs.com, two paladins wearing the sword in the main hand:
  NXMZV9xjvWpB7GtP fight 13 and D463FZBX1C9maJ7b fight 7.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded.
- Date: 2026-11-13 (NXMZV9xjvWpB7GtP) and 2026-11-14 (D463FZBX1C9maJ7b)
- Pull request: #1051
- Later count: Blackblade of Shahram's rate, §Count of the 2026-11-06 to 2026-11-22 reports

A proc is one summon line of 16602 by the wearer. A hit is a landed damage line of the wearer on an enemy (hit, crit,
block, glance or crush), or a melee cast that lands with no damage line, counted at its debuff. The sword's chain and
the ticks of damage over time are left out. Hits that land within 100 ms of each other form one cluster, since a white
swing and the Seal of Righteousness hit it triggers share a millisecond. A proc belongs to the cluster whose first hit
lies between 400 ms before the proc and 25 ms after it: the summon is logged at the cast line of a melee cast, 17 ms before
its damage line, and about 17 ms after a white swing.

## Which hits roll

Ranged-type hits (Judgement of Righteousness) roll nothing, 0 of 95 isolated ones, as EV-0119 found. A Judgement of
Command rolls twice, as EV-0119 found. A Seal of Righteousness hit rolls the sword as well: clusters of a white swing
and its seal hit procced 12 times in 98 (12.2%), single white swings 10 times in 168 (6.0%), and a fit with the seal
rolling explains the clusters better than one without it (log-likelihood -277.6 against -282.0).

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 16602
- Rate: 1 PPM, measured

1,005 procs on 17,613 rolling hits, 10 wearers, 15 runs, in the main hand: all hits 0.98 [0.93, 1.05]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.13 [0.97, 1.39], 132 procs on 2,011 lone
units, 10 wearers. Specials: 0.86 [0.49, 1.02], 52 procs on 1,035 lone units, 10 wearers; Rend landings give 20 procs on
330 of them. Two-roll abilities, priced as one roll: 0.89, 10 procs on 192 lone units (Execute 10 on 186 and Mongoose
Bite 0 on 6), left out of the rate. Direct: window [-20, 100] ms, median 0 ms, 98th percentile 0 ms, 189 isolated hits.
The paladin check, which sets no rate: lone white swings 1.02, 59 procs on 989 lone units, 17 wearers. Report
NdTXJzkWHLZrVQYm logs the same run as 1JLq46yg7wYNhan9 and counts once. Reports from 2026-11-16 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 |   971 | 17,108 | 0.97 |
| Hunter  |       2 |    34 |    505 | 1.15 |

Reports and fights: 1JLq46yg7wYNhan9 17-28; 2x7ZjvdbRwFMWKTP 3-45; 4VtCK6Qqr1MD2hpc 4; ARwPXzLJk78HgVQq 1;
JRtbDP4L3BdazHrV 1-2; NdTXJzkWHLZrVQYm 1-3; WFMjbyKYzvfNkPhH 6-14; WdbKG4RkmjxA7YV3 1-3; aDHTPGq6htJFx8Nw 1-8;
bAJFX6m7ftYWCKHP 1-4; bR9LQtgmKwZ3r6kW 1-25; n2HkhcYFrLK1tfZj 1-2; nqDL2hKfQB4yvjgJ 1-18; pHdqPxhAYLQTwya7 1-2;
qtNCmzfYQwAnRMkZ 1-9; tmHdvqxYj1z28PFy 1-9.

## Shahram's spells

Every one of the 102 summons cast one spell and nothing else.

| Spell                     | Picks |
| ------------------------- | ----- |
| Flames of Shahram 16596   | 33    |
| Will of Shahram 16598     | 30    |
| Might of Shahram 16600    | 13    |
| Blessing of Shahram 16599 | 9     |
| Fist of Shahram 16601     | 9     |
| Curse of Shahram 16597    | 8     |

- The hardcast begins 1,029 to 1,825 ms after the summon, median 1,421 ms, and runs 984 to 1,028 ms, median 1,002 ms.
- Flames hit 0 to 7 enemies a cast. Of 87 hit lines none crit, 41 were partly resisted, and the largest amount before
  resistance was 158.
- Will lands on the wearer alone. Each copy is removed 19,979 to 20,028 ms after its own application, with no refresh
  and no stack count, and two copies were up at once.
- Blessing ticks 50 mana and 50 health every 5 s on the party around Shahram. Two Blessings cast 13.0 s apart both
  ticked, each on its own 5 s beat.
- Fist lands on Shahram himself, 9 of 9 times.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the sword; 3,486 reports in all for the count, 16 of them for Blackblade of Shahram
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 705 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 6 of them for Blackblade of
  Shahram. The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on Blackblade of Shahram.

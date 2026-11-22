# EV-0125: Blackblade of Shahram's rate and what Shahram casts

- Claim: Blackblade of Shahram 12592 summons Shahram (16602) about 1.15 times a minute on landed main-hand hits. Each
  Shahram casts exactly one of his six spells, starts its hardcast 1.03 to 1.83 s after the summon, and every copy of
  his Will or Blessing runs on its own timer.
- Sources: public reports on forever.warcraftlogs.com, two paladins wearing the sword in the main hand:
  NXMZV9xjvWpB7GtP fight 13 and D463FZBX1C9maJ7b fight 7.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded.
- Date: 2026-11-13 (NXMZV9xjvWpB7GtP) and 2026-11-14 (D463FZBX1C9maJ7b)
- Pull request: #1051
- Later count: Blackblade of Shahram's rate, §Count of the 2026-10-08 to 2026-11-16 reports

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

- Count: §Count of the 2026-10-08 to 2026-11-16 reports
- Proc: weapon proc 16602, logged as 16602; rolls on main hand
- Sim rate: 1.15 PPM, observed; measured 1.15 PPM [0.96, 1.36], 4 wearers, 4 runs, 240 procs on 3576 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights                   | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out  |
| ---------------- | ------------------------ | ------- | ---- | ---------- | ----: | ---: | ---: | --------- |
| 1JLq46yg7wYNhan9 | 17-28                    | Warrior | mh   |            |    46 |  906 | 0.87 |           |
| 4f2ygdYCzFAvqc6K | 10, 12-14, 16-18         | Paladin | mh   |            |    92 | 1291 | 1.22 |           |
| D463FZBX1C9maJ7b | 7-8                      | Paladin | mh   |            |    43 |  680 | 1.08 |           |
| NXMZV9xjvWpB7GtP | 13-107, 109-116, 118-122 | Paladin | mh   |            |    59 |  699 | 1.45 |           |
| NdTXJzkWHLZrVQYm | 1-3                      | Warrior | mh   |            |    46 |  907 | 0.87 | duplicate |

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

## Count of the 2026-10-08 to 2026-11-16 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate's table names
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-16
- Pull request: #1065

The count is EV-0136 §Method, run on Blackblade of Shahram.

# EV-0136: Proc rates of 103 items from the Warcraft Logs sweep

- Claim: each item below procs at the rate its section gives, priced on each landed hit's weapon base speed; a rate whose
  item has 5 or more wearers and a 95% interval within 10% of it either side is measured, every other one observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each item's table names; 715 reports
  in all
- Game build: 1.60.1.70425 to 1.60.1.70938 by report date; each item's section gives its date span, and the hotfix each
  player ran is not recorded
- Date: 2026-10-08 to 2026-11-16
- Pull request: #1065

## Method

A proc is an event of the item's proc spell on the wearer or the target (damage, an aura applied or refreshed, a heal,
extra attacks, energy), counted once when one proc logs several events within 50 ms; a proc that misses or is resisted
still counts. The hits are the wearer's landed, non-periodic hits the trigger's proc flags hear: white swings of the item's
hand and melee-defense hits on the main hand (abilities, seals, Judgement of Command, Windfury Totem strikes) for a melee
proc, Auto Shot, Shoot, Throw and ranged-defense hits on the ranged weapon for a ranged one. Left out: periodic ticks,
damage shields, and Volley, Shank, Shield Slam, Shield Bash and Thunder Clap. A dual wielder's white swings are split
between the hands on their damage before crits, or by the weapon speeds when the split fails. A druid run in Cat or Bear
Form is left out, a run logged in two reports counts once, and a stretch with no proc where several were due is dropped
as the item swapped away. The rate is the procs over the sum of each hit's weapon base speed over 60, so haste raises
procs a minute without moving it. Its interval is the wider of the per-hit Wilson interval and 2,000 bootstrap draws of
whole wearers.

The method reproduces Destiny (EV-0132, 1.18), Heart of Wyrmthalak (EV-0133, 1.92 against 2.00), Skullforge Reaver
(EV-0121, 1.98 against 1.95) and the Highborne Research Tablet's weapon hits (EV-0131, 2.70 against 2.71). It reads
Flurry Axe 1.79 against EV-0120's 1.65, so no item rolled mostly by the off hand counts as measured. Items with fewer
than 50 procs, or one wearer with fewer than 100, are left out.

Each table lists every run looked at; a run left out names why. PPM is the run's procs over its expected procs at 1 PPM.

## What the procs do beyond their rate

Read from the same wearers' streams, to check what the rows state before the sim models it.

- Mana Channeling Wand (18483): all 211 procs of 1302350 restore 28 mana, the row's amount; the tooltip's 65 is the
  trigger 1302348's.
- Hammer of Expertise (8708): Forging Strike (1317432) is Physical on the melee attack table, as its row states: of 614
  hits, 501 hit, 73 crit, 14 were dodged, 11 parried, 12 blocked and 3 missed, and 49 show armor mitigation.
- Sul'thraze the Lasher (9372), 7 wearers, 958 landed Rage hits: all 487 timed Shadow Lash hits landed 5.0 s after the
  Lash went up or refreshed; 46 of 511 Lash damage events crit and 2 were immune. Of 111 Rages landing while a Lash was
  running, 99 put up no new Lash, 83 of them followed by the running Lash's damage at its 5 s. No missed Rage (22) put
  up a Lash.
- Pendulum of Doom (9425), 7 wearers, 615 landed hits: 469 sit in the plain damage range and 146 at about 1.5 times it
  (per wearer 1.38 to 1.50); the larger ones land a median 1.7 s before the wearer's last hit on that target, against
  15.1 s for plain ones. The logs hold no target health, so the threshold is the row's.
- Flame Wrath (11809), 8 wearers, 421 procs: the fire shield went up 297 times and refreshed 130 times, 195 windows lasted
  exactly 15 s, and it dealt damage 2,292 times inside its windows and 47 outside, never a crit. 8 procs came 0 to 1 s
  apart, so the row's 5 s category cooldown does not lock out procs.
- Gahz'rilla Fang (9467), 14 wearers, 594 non-crit hits: the damage fits 85.8 plus 1.97% per Static Electricity stack
  (the row: 84 times 1 plus 2% a stack), and 95 procs at 50 stacks dealt a median 168. 581 stack removals left exactly 1
  stack, 564 of them at a proc's timestamp. Single-stack gains came 50 ms to over 3 s apart, with no steady cadence.
- The Chief's Enforcer (9477), 13 wearers: 416 strikes applied or refreshed the stun, which also landed on 20 of 26
  dodged strikes, 6 of 12 parried ones and 3 of 5 misses; 56 stuns were immune, 2 resisted, and 305 of 389 lasted the full
  1.5 s.

## Axe of Inevitable Ruin (274484)

- Proc: weapon proc 1294226, logged as 1294226; rolls on main hand
- Sim rate: 2.15 PPM, observed; measured 2.16 PPM [1.98, 2.35], 4 wearers, 4 runs, 469 procs on 4351 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| LMdRqJGhV17XwWyT | 16-22  | Shaman | mh   |   169 | 1581 | 2.14 |          |
| MpcVyhTHnqCJvDFx | 1-6    | Hunter | mh   |     9 |   62 | 2.90 |          |
| mqzgv4rPVxA3JHGf | 1-2    | Shaman | mh   |    79 |  713 | 2.22 |          |
| yhW9cLwxnHNDP2jF | 4-8    | Shaman | mh   |   212 | 1995 | 2.13 |          |

## Barbaric Crossbow (272999)

- Proc: weapon proc 1291551, logged as 1291551; rolls on ranged
- Sim rate: 3.45 PPM, observed; measured 3.44 PPM [3.17, 3.71], 10 wearers, 12 runs, 630 procs on 5491 hits
- Reports from 2026-11-07 to 2026-11-16

| Report           | Fights      | Class   | Slot   | Procs | Hits |   PPM | Left out |
| ---------------- | ----------- | ------- | ------ | ----: | ---: | ----: | -------- |
| 1AfDWYQT7Xjd24K3 | 13-14, 16   | Hunter  | ranged |    84 |  654 |  3.85 |          |
| 1RrDAhp7ZvVYc9P3 | 1-6         | Rogue   | ranged |     0 |    0 |       |          |
| 1W9N26dHfnPV4G7A | 1, 3-9      | Hunter  | ranged |    64 |  554 |  3.47 |          |
| 1y6GV7KLTZPrJH2w | 1-3         | Hunter  | ranged |    70 |  604 |  3.48 |          |
| 2GbXqD7KcWTfRnYa | 1-2         | Hunter  | ranged |   110 |  795 |  4.15 |          |
| 2MHXV8Y9jx43vhpk | 3-12, 14-18 | Warrior | ranged |     2 |   43 |  1.40 |          |
| 2tJkzh16X38wrQfP | 1-3         | Hunter  | ranged |   114 | 1052 |  3.25 |          |
| Td29Nzahj1cmGMWZ | 1-2         | Hunter  | ranged |    79 |  717 |  3.31 |          |
| YHh2t9agcvQjTAZq | 2-3         | Hunter  | ranged |    48 |  470 |  3.06 |          |
| kBHbV3DZjamvfJY4 | 1-3         | Rogue   | ranged |     1 |    3 | 10.00 |          |
| nALy2WrwHC9qdMFm | 1-2         | Hunter  | ranged |    58 |  599 |  2.90 |          |
| wmT1zqYBat3fHDLy | 1-5, 8-9    | Warrior | ranged |     0 |    0 |       |          |

## Bashguuder (13204)

- Proc: weapon proc 17315, logged as 17315; rolls on main hand, off hand
- Sim rate: 3.3 PPM, observed; measured 3.28 PPM [2.82, 4.15], 7 wearers, 7 runs, 955 procs on 9693 hits
- Reports from 2026-11-11 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| FHqYCNcKntvDAjxg | 34-36  | Warrior | mh   |    92 |  939 | 3.27 |          |
| JRQqNL6DkvWKYdPB | 1-7, 9 | Warrior | mh   |   367 | 4131 | 2.96 |          |
| VK83G7MdBhLRFAfa | 1-42   | Rogue   | oh   |    58 |  596 | 3.25 |          |
| VYBjgJx8aZtRhmpy | 38-46  | Warrior | mh   |   160 | 1041 | 5.12 |          |
| VYBjgJx8aZtRhmpy | 44-46  | Rogue   | oh   |    61 | 1006 | 2.02 |          |
| v8Zp2QF7An4d9xMj | 1-10   | Warrior | oh   |   117 | 1272 | 3.07 |          |
| xXWkB3fqAjt9NJGH | 1-5    | Warrior | oh   |   100 |  708 | 4.71 |          |

## Black Duskwood Staff (937)

- Proc: weapon proc 18138, logged as 18138; rolls on main hand
- Sim rate: 0.9 PPM, observed; measured 0.88 PPM [0.68, 1.13], 2 wearers, 3 runs, 57 procs on 1386 hits
- Reports from 2026-11-10 to 2026-11-15
- Notes: 1 of 4 runs are feral druids, left out (paw speed, not the weapon); few wearers (2): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| 4bqYVtkjL6aKdwQM | 1-2    | Shaman | mh   |    18 |  366 | 1.05 |          |
| 9RgXYmBfN7hCdznQ | 1-3    | Shaman | mh   |    19 |  510 | 0.80 |          |
| waZyxVKk43BzGLN6 | 1-11   | Shaman | mh   |    20 |  510 | 0.84 |          |
| 6GVg7FYcHWJmrbKw | 1-16   | Druid  | mh   |    14 | 1270 | 0.24 | feral    |

## Black Malice (3194)

- Proc: weapon proc 1292155, logged as 1292155; rolls on main hand
- Sim rate: 2.05 PPM, observed; measured 2.06 PPM [1.70, 2.43], 5 wearers, 5 runs, 246 procs on 2175 hits
- Reports from 2026-11-10 to 2026-11-15
- Notes: 1 of 6 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 4WrfjAqJYhxGHwPd | 1-3    | Warrior | mh   |    58 |  415 | 2.54 |          |
| 7APg28BmdtzHhQFV | 1-6    | Shaman  | mh   |    20 |  203 | 1.79 |          |
| GV14JHxQP3wcbg9X | 1-2    | Paladin | mh   |    73 |  544 | 2.44 |          |
| PAy6j9tqdkwMTQHn | 6-10   | Paladin | mh   |    73 |  803 | 1.65 |          |
| VLcP3WgkCfKZhB2v | 5-8    | Paladin | mh   |    22 |  210 | 1.90 |          |
| 84aWfvkV2qb3mAnL | 1      | Druid   | mh   |    23 | 1355 | 0.31 | feral    |

## Black Menace (6831)

- Proc: weapon proc 1293407, logged as 1293407; rolls on main hand, off hand
- Sim rate: 2.05 PPM, observed; measured 2.05 PPM [1.87, 2.24], 15 wearers, 15 runs, 792 procs on 15460 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights              | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 3GT4tz8gxnphcbkD | 1-7                 | Rogue   | oh   |    56 | 1016 | 2.20 |          |
| 3pCKd6jqgHyFc9kf | 1-17                | Rogue   | oh   |    42 |  877 | 1.92 |          |
| 3wjCThZxkVFWzG69 | 1-6                 | Rogue   | mh   |    70 | 1541 | 1.82 |          |
| 3zTJ2VXvQ8GCyFbc | 1-13                | Warrior | mh   |   103 | 1526 | 2.70 |          |
| 3zTJ2VXvQ8GCyFbc | 1-13                | Rogue   | oh   |    29 |  579 | 2.00 |          |
| 4g9w7aLvtz2RmcFx | 1-16                | Rogue   | oh   |    23 |  592 | 1.56 |          |
| 6Ax3QmHWJqnPzvhd | 1-3                 | Rogue   | oh   |    31 |  669 | 1.85 |          |
| 6RMV4ya9h1pWjYLn | 1-3                 | Rogue   | mh   |    76 | 1565 | 1.94 |          |
| 8akXKLNxCdDtpqfv | 1-4, 7, 9-22, 24-38 | Rogue   | mh   |    42 | 1098 | 1.53 |          |
| 9Z473MfXLYqcyWPg | 1-2                 | Rogue   | oh   |    36 |  709 | 2.03 |          |
| BTRMkxmvfrwcZNaG | 1-3                 | Rogue   | mh   |   103 | 1730 | 2.38 |          |
| cFRpkf1jWTaHrgL3 | 1-2                 | Rogue   | oh   |    34 |  524 | 2.59 |          |
| hfmVrK1Aw7B3yqWF | 1-2, 4              | Rogue   | oh   |    43 |  802 | 2.14 |          |
| tac7hXVRgkbZNMDy | 3                   | Rogue   | oh   |    35 |  593 | 2.36 |          |
| x3PftMnXCBZgLAG2 | 1, 4-9              | Rogue   | mh   |    69 | 1639 | 1.68 |          |

## Blackblade of Shahram (12592)

- Proc: weapon proc 16602, logged as 16602; rolls on main hand
- Sim rate: 1.15 PPM, observed; measured 1.15 PPM [0.96, 1.36], 4 wearers, 4 runs, 240 procs on 3576 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights                   | Class   | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | ------------------------ | ------- | ---- | ----: | ---: | ---: | --------- |
| 1JLq46yg7wYNhan9 | 17-28                    | Warrior | mh   |    46 |  906 | 0.87 |           |
| 4f2ygdYCzFAvqc6K | 10, 12-14, 16-18         | Paladin | mh   |    92 | 1291 | 1.22 |           |
| D463FZBX1C9maJ7b | 7-8                      | Paladin | mh   |    43 |  680 | 1.08 |           |
| NXMZV9xjvWpB7GtP | 13-107, 109-116, 118-122 | Paladin | mh   |    59 |  699 | 1.45 |           |
| NdTXJzkWHLZrVQYm | 1-3                      | Warrior | mh   |    46 |  907 | 0.87 | duplicate |

## Blade of the Wretched (10803)

- Proc: weapon proc 18088, logged as 18088; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.97 PPM [1.81, 2.20], 7 wearers, 7 runs, 1117 procs on 16160 hits
- Reports from 2026-11-12 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 6kN78G2xfqy3vwTP | 6        | Rogue   | mh   |     8 |  105 | 2.17 |          |
| 7M3gGcH6bprxKqZY | 1-3, 5-8 | Warrior | mh   |   140 | 1834 | 2.18 |          |
| 9H72ZkdBGfhCy6P3 | 1-21     | Paladin | mh   |   250 | 3517 | 2.03 |          |
| 9zMva8xTVhFkHPAt | 1-10     | Paladin | mh   |   271 | 4090 | 1.89 |          |
| D463FZBX1C9maJ7b | 4-8      | Paladin | mh   |    92 | 1068 | 2.46 |          |
| N2m8nPTGJQrj1DhH | 4-5      | Rogue   | mh   |   142 | 1828 | 2.22 |          |
| VwyW3ZP1H6RTpdmf | 1-5      | Warrior | mh   |   214 | 3718 | 1.64 |          |

## Bloodrazor (809)

- Proc: weapon proc 17504, logged as 17504; rolls on main hand
- Sim rate: 1.05 PPM, observed; measured 1.03 PPM [0.76, 1.31], 5 wearers, 5 runs, 361 procs on 7810 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| AZchwX93DmyHLbv6 | 1-3    | Warrior | mh   |    56 | 1909 | 0.65 |          |
| NWvgFp3kfR2bHyYa | 1-35   | Rogue   | mh   |    72 | 1328 | 1.20 |          |
| RW1BNk6Ybyh3XjLQ | 1-21   | Warrior | mh   |   143 | 2246 | 1.41 |          |
| jJWt7Dnm4qc9rzQd | 1      | Warrior | mh   |    55 | 1578 | 0.77 |          |
| kJ3WdhMLqv7K2Nyf | 1-3    | Rogue   | mh   |    35 |  748 | 1.04 |          |

## Bloodspiller (7753)

- Proc: weapon proc 18200, logged as 18200; rolls on main hand
- Sim rate: 0.95 PPM, observed; measured 0.94 PPM [0.80, 1.15], 6 wearers, 6 runs, 130 procs on 3065 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 0.77 -> 0.94 (29 active minutes, 18% of hits, in 2 runs)

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 1YQTjgZh2KPBGCbN | 1-3    | Hunter  | mh   |    53 | 1329 | 0.89 |          |
| 2ywhMnJq4Y3zaWtg | 1-9    | Warrior | mh   |    18 |  445 | 0.90 |          |
| Fx34Wc78qhnPmVz6 | 1-5    | Shaman  | mh   |     2 |   55 | 0.81 |          |
| LxdYWR47ZFbTNBhM | 1-2    | Shaman  | mh   |    28 |  530 | 1.17 |          |
| Qtca4MnwCRLKPdN8 | 1      | Paladin | mh   |     9 |  112 | 1.79 |          |
| YFnCf6cWNRpgVMHT | 17-31  | Warrior | mh   |    20 |  594 | 0.75 |          |

## Bonechill Hammer (14487)

- Proc: weapon proc 18276, logged as 18276; rolls on main hand
- Sim rate: 1.8 PPM, observed; measured 1.78 PPM [1.64, 2.04], 4 wearers, 4 runs, 511 procs on 7166 hits
- Reports from 2026-11-13 to 2026-11-14
- Notes: swap noise: the swap drop moved PPM 1.31 -> 1.78 (48 active minutes, 26% of hits, in 1 runs); few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 9zxHaWbYT3gqDQvF | 17-18  | Rogue   | mh   |    93 | 1111 | 2.09 |          |
| rphdM14FbnTBaj3V | 1-27   | Paladin | mh   |    50 |  739 | 1.69 |          |
| vwCXzhMQ2drgpRxD | 1-41   | Paladin | mh   |   311 | 4589 | 1.69 |          |
| zmDY4LVX7N8vh9aR | 1-3    | Rogue   | mh   |    57 |  727 | 1.96 |          |

## Boneslicer (273819)

- Proc: weapon proc 1292573, logged as 1292573; rolls on main hand
- Sim rate: 2.8 PPM, observed; measured 2.78 PPM [2.41, 3.38], 7 wearers, 7 runs, 180 procs on 1944 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 6 of 13 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 3rbcnjTMRfwxv1mh | 1-10   | Hunter  | mh   |     9 |  101 | 2.67 |          |
| 79RVgZWM8dF2wkCp | 1-6, 8 | Hunter  | mh   |     5 |   65 | 2.31 |          |
| 7zcat1AyFnrJ4BQR | 10-19  | Hunter  | mh   |    10 |   95 | 3.16 |          |
| BzGXfjMLJ2RyCwHm | 1-7    | Hunter  | mh   |     2 |   36 | 1.67 |          |
| FAf9V76BctbG3mR4 | 1      | Warrior | mh   |    13 |  104 | 3.75 |          |
| K9WGdrx248FvaDVY | 1-3    | Hunter  | mh   |     4 |   24 | 5.00 |          |
| ywDgt7KhbjmAL82V | 1-11   | Paladin | mh   |   137 | 1519 | 2.71 |          |
| 3rgax8hkZzJjXKp2 | 4-5    | Druid   | mh   |    22 |  480 | 1.38 | feral    |
| 8BjkHKdLY1NWxFVw | 1      | Druid   | mh   |    26 |  660 | 1.18 | feral    |
| AH4pXcWv2BGFZk8f | 1-16   | Druid   | mh   |    63 | 1780 | 1.06 | feral    |
| AzbN8GY9DnBrH4aW | 1-18   | Druid   | mh   |    11 | 1171 | 0.28 | feral    |
| JGFcm9zq48CArhgk | 3-6    | Druid   | mh   |    32 |  832 | 1.15 | feral    |
| XGJ38LhbnH7FPdKq | 1-3    | Druid   | mh   |    33 | 1114 | 0.89 | feral    |

## Bow of Searing Arrows (2825)

- Proc: weapon proc 29638, logged as 29638; rolls on ranged
- Sim rate: 5.4 PPM, observed; measured 5.42 PPM [4.94, 5.81], 5 wearers, 5 runs, 701 procs on 2873 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights    | Class  | Slot   | Procs | Hits |  PPM | Left out  |
| ---------------- | --------- | ------ | ------ | ----: | ---: | ---: | --------- |
| 6Krp8bkfx4cRjBaz | 1, 3-15   | Hunter | ranged |    97 |  429 | 5.02 |           |
| AhLc6aGKMCdvBgDW | 1-5       | Hunter | ranged |   233 |  940 | 5.51 |           |
| H2GvZtVYNy4QzqLw | 12-21     | Hunter | ranged |   150 |  538 | 6.20 |           |
| KGQHdbFCZpt1YX23 | 18, 20-23 | Hunter | ranged |    99 |  484 | 4.55 |           |
| cJhVzAdLx3BHqpkT | 1-4       | Hunter | ranged |   122 |  482 | 5.62 |           |
| 2AzZR61XcPdKgQrB | 6, 8-14   | Hunter | ranged |    99 |  484 | 4.55 | duplicate |

## Brush Cutter (274265)

- Proc: weapon proc 1293518, logged as 1293518; rolls on main hand, off hand
- Sim rate: 2.25 PPM, observed; measured 2.26 PPM [2.03, 2.52], 3 wearers, 3 runs, 301 procs on 2850 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: swap noise: the swap drop moved PPM 1.81 -> 2.26 (31 active minutes, 20% of hits, in 1 runs); few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| cr6qMtj1vCTPHVfw | 1-7    | Warrior | oh   |    48 |  479 | 2.15 |          |
| pjRYqGgB2rdvawNn | 1-10   | Warrior | mh   |   199 | 1851 | 2.30 |          |
| qcdyZAr2ganDtxNW | 13-42  | Rogue   | oh   |    54 |  520 | 2.22 |          |

## Burning War Axe (2299)

- Proc: weapon proc 18199, logged as 18199; rolls on main hand
- Sim rate: 1.05 PPM, observed; measured 1.04 PPM [0.92, 1.19], 8 wearers, 8 runs, 242 procs on 4251 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 3cKTmVyb1ZPwFXkf | 1, 3-13     | Shaman  | mh   |    25 |  456 | 1.00 |          |
| 7XwRKY4qkjDt3ryz | 2-3         | Paladin | mh   |    38 |  758 | 0.91 |          |
| Fgy2bt81NL9CAzaR | 1-13, 15    | Shaman  | mh   |    13 |  200 | 1.18 |          |
| KRCBMJvg72fHVwWL | 1-17        | Warrior | mh   |    21 |  394 | 0.97 |          |
| LwTzGJXrN3HWqCfD | 1-5         | Warrior | mh   |    35 |  443 | 1.44 |          |
| RBmtAMLxz78Pj2Wh | 7-10        | Shaman  | mh   |    23 |  485 | 0.86 |          |
| VQFaRpT6J2xLDbZg | 1, 3-5      | Warrior | mh   |    20 |  289 | 1.26 |          |
| Wf8ghC6TtQKrkXdv | 8-10, 12-15 | Shaman  | mh   |    67 | 1226 | 0.99 |          |

## Claw of Celebras (17738)

- Proc: weapon proc 21952, logged as 21952; rolls on off hand
- Sim rate: 2.4 PPM, observed; measured 2.42 PPM [1.91, 2.75], 5 wearers, 6 runs, 213 procs on 2931 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2taZz4vTdfXYAk3b | 3-9      | Rogue   | oh   |     5 |  111 | 1.51 |          |
| 4Qm6MgZWBTx78Lvp | 1-2, 5-6 | Warrior | oh   |    64 |  701 | 3.04 |          |
| 4XGyazkdvmP98r2L | 1-7      | Rogue   | oh   |     9 |  166 | 1.80 |          |
| 61Hx8DgwhdbWaN9C | 1-12     | Warrior | oh   |    50 |  641 | 2.60 |          |
| 6yrCJtHYmaRnZBM3 | 1-5      | Rogue   | oh   |    30 |  460 | 2.17 |          |
| 7AxaTpFZMqGb9wNd | 1-8      | Warrior | oh   |    55 |  852 | 2.15 |          |

## Claw of the Shadowmancer (2912)

- Proc: weapon proc 16409, logged as 16409; rolls on main hand, off hand
- Sim rate: 1 PPM, observed; measured 1.00 PPM [0.85, 1.15], 10 wearers, 10 runs, 195 procs on 6161 hits
- Reports from 2026-11-07 to 2026-11-16

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| 1jgtdNmDMRycbZrP | 1      | Rogue  | oh   |    13 |  285 | 1.44 |          |
| 9LgNW1jQ2Zh8XDkY | 1-12   | Rogue  | mh   |    51 | 1344 | 1.20 |          |
| 9vwmqxP4dLKhDC8Z | 1-23   | Rogue  | oh   |    11 |  446 | 0.78 |          |
| AHdpxG72rk6CzbMa | 1, 3   | Rogue  | oh   |     6 |  293 | 0.65 |          |
| FYmnawMf8gjB4NQt | 1-3    | Rogue  | mh   |    25 | 1093 | 0.72 |          |
| JK4pARZQzyLb2DMh | 1-3    | Rogue  | mh   |    29 |  864 | 1.06 |          |
| L8yaqgDXG26Nf9v1 | 1-2    | Rogue  | oh   |    20 |  532 | 1.19 |          |
| LbmB2QdWnPT7JGAF | 6      | Hunter | oh   |     1 |   61 | 0.52 |          |
| Tw4BdRP1HmZqJL9F | 1-3    | Rogue  | mh   |    26 |  873 | 0.94 |          |
| hcGCaDW9tRYFB6Ax | 12-18  | Rogue  | oh   |    13 |  371 | 1.11 |          |

## Cobalt Crusher (7730)

- Proc: weapon proc 18204, logged as 18204; rolls on main hand
- Sim rate: 2.25 PPM, observed; measured 2.24 PPM [1.88, 2.66], 2 wearers, 2 runs, 112 procs on 938 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 4 of 6 runs are feral druids, left out (paw speed, not the weapon); few wearers (2): the interval is about the per-wearer range

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 8frHF64GzAhBxDyb | 4-14, 16-22 | Paladin | mh   |    51 |  412 | 2.32 |          |
| g8b1MkJnKtyfzvPa | 4-5         | Shaman  | mh   |    61 |  526 | 2.17 |          |
| LN78yvQwgJpcAYCG | 1-2, 4-6    | Druid   | mh   |     4 |  436 | 0.17 | feral    |
| M6kJfn7a3r2KFRgd | 1-4         | Druid   | mh   |    21 |  621 | 0.63 | feral    |
| XzKCqkBbmY3W2Tc8 | 6-7         | Druid   | mh   |     9 |  716 | 0.24 | feral    |
| cvfBaDxQX7K2zFtk | 1-7         | Druid   | mh   |    42 | 1732 | 0.45 | feral    |

## Coldrage Dagger (10761)

- Proc: weapon proc 1293790, logged as 1293790; rolls on main hand, off hand
- Sim rate: 2.95 PPM, observed; measured 2.93 PPM [2.69, 3.16], 16 wearers, 16 runs, 1258 procs on 17182 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights    | Class  | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | --------- | ------ | ---- | ----: | ---: | ---: | --------- |
| 1jgTLvJQBMP6kxrm | 1-12      | Rogue  | oh   |    12 |  254 | 1.89 |           |
| 1w3vfpYKqbrBNZPQ | 1-2       | Rogue  | mh   |    76 | 1536 | 1.98 |           |
| 3PAwV6XFaB7hqDgz | 1-12      | Rogue  | oh   |    20 |  258 | 3.10 |           |
| 3THgCcxGJ4XYd6Kp | 3         | Rogue  | oh   |    43 |  601 | 2.86 |           |
| 4xnDCWwrGKAFdNHv | 1         | Rogue  | mh   |   101 | 1386 | 2.92 |           |
| 7dZrzYqjKHh28mvW | 1-2       | Rogue  | oh   |    70 |  834 | 3.36 |           |
| 987q1pjnYdVQbCmT | 1-6       | Rogue  | oh   |    84 | 1304 | 2.58 |           |
| 9QnWbJphKLFVgZ6z | 3-8       | Rogue  | oh   |    86 | 1209 | 2.84 |           |
| NfqRryF7gnMcX2ak | 1-2       | Rogue  | mh   |   118 | 1588 | 2.97 |           |
| QKg4z1amMYyBCp69 | 1-3, 5-30 | Rogue  | mh   |   107 | 1214 | 3.52 |           |
| R9njq1AGDvmp36bM | 9         | Rogue  | oh   |     5 |   24 | 8.32 |           |
| RW1BNk6Ybyh3XjLQ | 1-21      | Rogue  | mh   |   226 | 3058 | 2.96 |           |
| RvcrA3P9jznxhHVQ | 1-2, 4-7  | Hunter | mh   |    17 |  215 | 3.16 |           |
| TADvnpKkdwgZGJ6b | 1         | Rogue  | oh   |    93 | 1357 | 2.74 |           |
| g41xhtdFCXZ8Vr9z | 1-2       | Rogue  | oh   |    73 |  724 | 4.04 |           |
| pqfK2AgHZ7DnhvPF | 1         | Rogue  | mh   |   127 | 1618 | 3.14 |           |
| dRch7jNvCBJfgQt6 | 1         | Rogue  | oh   |    45 |  676 | 2.66 | duplicate |

## Cookie's Tenderizer (5197)

- Proc: weapon proc 1292008, logged as 1292008; rolls on main hand, off hand
- Sim rate: 1.9 PPM, observed; measured 1.90 PPM [1.58, 2.26], 7 wearers, 8 runs, 400 procs on 4502 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 1pDcTvkrCBhZJRgf | 1-14   | Paladin | mh   |   105 | 1281 | 1.76 |          |
| 2PWL9DB1k47p8MRd | 1      | Warrior | mh   |    63 |  464 | 2.91 |          |
| 2htpkPNX3yZ8K74C | 1-3    | Rogue   | oh   |    18 |  272 | 1.42 |          |
| 3BH2d16wtD9F78QY | 1-18   | Paladin | mh   |    84 |  898 | 2.00 |          |
| 3BYFMvdKPgLc4xDn | 4-5    | Paladin | mh   |    65 |  660 | 2.11 |          |
| FWHahtQvqjL3wNdV | 1-3    | Rogue   | mh   |    32 |  573 | 1.20 |          |
| QmjBxrzvah2XTVR1 | 7-8    | Warrior | mh   |    33 |  354 | 2.00 |          |
| pwK87DzQgRHYCPNb | 1-4    | Priest  | mh   |     0 |    0 |      |          |

## Cursed Felblade (14145)

- Proc: weapon proc 18381, logged as 18381; rolls on main hand
- Sim rate: 2.55 PPM, observed; measured 2.56 PPM [2.18, 3.02], 8 wearers, 8 runs, 666 procs on 6786 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 12VhLkPa98DYFH4g | 1-3         | Warrior | mh   |    88 |  774 | 2.97 |          |
| 1RrDAhp7ZvVYc9P3 | 1-6         | Rogue   | mh   |    39 |  572 | 1.78 |          |
| 1XFBL6hQPHcY4Zwr | 1-2         | Warrior | mh   |    78 |  591 | 3.44 |          |
| 1cf7ba34WkwGyzQY | 1-12, 14-16 | Warrior | mh   |    72 |  976 | 1.92 |          |
| 3CyQk69LzJjTn2X4 | 1-5         | Warrior | mh   |   100 | 1066 | 2.45 |          |
| 3rDXa6gPxYNZJjFH | 2           | Rogue   | mh   |   122 | 1470 | 2.17 |          |
| 9aCJ7GH6QRFjgqb4 | 1-11        | Warrior | mh   |    50 |  471 | 2.77 |          |
| txhaZrbjXR7Kf2TD | 1, 3        | Warrior | mh   |   117 |  866 | 3.52 |          |

## Darkwater Talwar (11121)

- Proc: weapon proc 16408, logged as 16408; rolls on main hand
- Sim rate: 3.3 PPM, observed; measured 3.31 PPM [2.96, 3.68], 17 wearers, 19 runs, 2148 procs on 17689 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 14ZTxcCRXhQLfYwP | 1-2, 5-6 | Paladin | mh   |    45 |  431 | 2.85 |          |
| 1CYTfP3pFDcaRrJm | 1-8      | Warrior | mh   |   154 |  950 | 4.42 |          |
| 1vJgfm9dGnkDzwXb | 1-2      | Warrior | mh   |   174 | 1348 | 3.52 |          |
| 23dT9fk1DqcaJmwn | 1-7      | Paladin | mh   |   130 | 1270 | 2.79 |          |
| 24KTy3bLGNQm1rdx | 1-20     | Rogue   | mh   |    21 |  173 | 3.31 |          |
| 2GYNtHzfKmWpTPVZ | 1-14     | Warrior | mh   |   326 | 1929 | 4.61 |          |
| 2Z3YxGdfPhMK97kQ | 1-4      | Warrior | mh   |    88 |  745 | 3.22 |          |
| DK72jqp46VtnYHfZ | 1        | Rogue   | mh   |    20 |  173 | 3.15 |          |
| DK72jqp46VtnYHfZ | 1        | Paladin | mh   |    15 |  178 | 2.30 |          |
| H3vdcCr6JX1gwYtq | 1-3      | Rogue   | mh   |    52 |  509 | 2.79 |          |
| LHRhtADNW8GfyYqT | 1        | Rogue   | mh   |    73 |  840 | 2.37 |          |
| NDf1mPJRCaTM8WkK | 2        | Warrior | mh   |    31 |  200 | 4.23 |          |
| QgR6kbHvY1CBhXFj | 1-2      | Rogue   | mh   |   231 | 2015 | 3.13 |          |
| Z1KYP3yWDVN7L6XH | 1-14     | Paladin | mh   |   202 | 1742 | 3.16 |          |
| khV3XMmJ4aRKzxBt | 1-3      | Rogue   | mh   |    37 |  333 | 3.03 |          |
| mRxb9L2wNkMJTWXY | 1-7      | Rogue   | mh   |   126 | 1274 | 2.70 |          |
| mRxb9L2wNkMJTWXY | 1-7      | Warrior | mh   |   245 | 1846 | 3.62 |          |
| vk2wGPDbQp8TZBR4 | 1        | Rogue   | mh   |    91 |  869 | 2.86 |          |
| yzaKX18ZAmpHvkjP | 1-16     | Paladin | mh   |    87 |  864 | 2.75 |          |
| 4NWMr2dJPyH9v8az | 1-2      | Hunter  | mh   |     0 |    0 |      | swap     |

## Darrowspike (13984)

- Proc: weapon proc 18276, logged as 18276; rolls on main hand
- Sim rate: 2.2 PPM, observed; measured 2.18 PPM [1.94, 2.44], 1 wearers, 1 runs, 282 procs on 5182 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: one wearer: no wearer interval

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 7fy93LdcwQhvF2gP | 1-7    | Warrior | mh   |   282 | 5182 | 2.18 |          |

## Dazzling Longsword (869)

- Proc: weapon proc 13752, logged as 13752; rolls on main hand
- Sim rate: 2.8 PPM, measured; measured 2.79 PPM [2.59, 2.97], 7 wearers, 7 runs, 887 procs on 7349 hits
- Reports from 2026-11-09 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 1FV3WyfXGCcPrT9b | 1, 3-8 | Hunter  | mh   |   139 | 1195 | 2.69 |          |
| 4C28YPXqgTMQJxB7 | 5-9    | Rogue   | mh   |    44 |  375 | 2.71 |          |
| 8X2Kc46faTnxVBLm | 1-3    | Rogue   | mh   |   110 | 1042 | 2.44 |          |
| BCZgcRxrh7dA4ayX | 1-4    | Rogue   | mh   |   277 | 2101 | 3.04 |          |
| BgKt3Qm2fJDAWLzh | 1-23   | Rogue   | mh   |    76 |  683 | 2.57 |          |
| KVdX1Wn4GpzAH9MN | 11-12  | Rogue   | mh   |    80 |  573 | 3.22 |          |
| b12KD98fndF4hMq3 | 10-18  | Paladin | mh   |   161 | 1381 | 2.69 |          |

## Deadhead Blade (274271)

- Proc: weapon proc 1293518, logged as 1293518; rolls on main hand, off hand
- Sim rate: 2 PPM, observed; measured 2.00 PPM [1.57, 2.87], 4 wearers, 5 runs, 238 procs on 3966 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| QRCgnhck3KVq6vBT | 1-2    | Hunter | oh   |    28 |  529 | 1.76 |          |
| VLz78BfZyWa3Cxr2 | 1-9    | Hunter | oh   |    11 |  314 | 1.17 |          |
| nX1qbatTzGZ3CAwB | 1, 3-7 | Hunter | oh   |    11 |  218 | 1.68 |          |
| pQ48t7ArW9aDdwNn | 1-4    | Rogue  | mh   |   147 | 2521 | 1.94 |          |
| zp9yVrtfTNZJCv7g | 9-10   | Rogue  | oh   |    41 |  383 | 3.57 |          |

## Destiny (647)

- Proc: weapon proc 17152, logged as 17152; rolls on main hand
- Sim rate: 1.2 PPM, measured; measured 1.18 PPM [1.07, 1.27], 7 wearers, 8 runs, 670 procs on 9440 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1ZyhaVDdFtTYWMcC | 10-12       | Warrior | mh   |    33 |  496 | 1.11 |          |
| Qb1yWxa8vHCGL7Yr | 12-15       | Paladin | mh   |   124 | 1754 | 1.18 |          |
| VLz78BfZyWa3Cxr2 | 1-9         | Warrior | mh   |   152 | 1930 | 1.31 |          |
| nX1qbatTzGZ3CAwB | 1, 3-7      | Warrior | mh   |   160 | 2236 | 1.19 |          |
| vZqVmLYQa6jwP3A9 | 2-16, 18-21 | Paladin | mh   |    51 |  734 | 1.16 |          |
| xbJBZj14NWCXrhAm | 5-12        | Paladin | mh   |    56 |  989 | 0.94 |          |
| zFXxNAn69HBGRaqg | 1-10        | Paladin | mh   |    53 |  709 | 1.25 |          |
| zFXxNAn69HBGRaqg | 1-10        | Paladin | mh   |    41 |  592 | 1.15 |          |

## Dwarven Hand Cannon (2099)

- Proc: weapon proc 29639, logged as 29639; rolls on ranged
- Sim rate: 4.3 PPM, measured; measured 4.31 PPM [4.03, 4.60], 5 wearers, 5 runs, 702 procs on 3372 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class  | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ------ | ----: | ---: | ---: | -------- |
| BfVjHYvWdKkP8Jz7 | 1-2    | Hunter | ranged |   108 |  518 | 4.31 |          |
| F3BTZGrW48Mp6n2L | 3-39   | Hunter | ranged |   157 |  809 | 4.02 |          |
| KPqgbZxhn6BXwCvJ | 1      | Hunter | ranged |   100 |  530 | 3.90 |          |
| T6VDdFhPxQ43pHM2 | 1-2    | Hunter | ranged |   164 |  722 | 4.70 |          |
| hH4YRGVFz3KnN6bw | 1-8    | Hunter | ranged |   173 |  793 | 4.51 |          |

## Eater of the Dead (10805)

- Proc: weapon proc 1299869, logged as 1299869; rolls on main hand
- Sim rate: 1.65 PPM, observed; measured 1.67 PPM [1.52, 2.04], 6 wearers, 6 runs, 409 procs on 6693 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights          | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1PcrGTyatDAFx8Yk | 1-2             | Hunter  | mh   |    16 |  219 | 1.99 |          |
| 8yPRmZdrq4QhcGn1 | 1-4             | Warrior | mh   |    77 | 1276 | 1.65 |          |
| RJ82yLB7HcYrd39M | 2               | Warrior | mh   |     1 |   71 | 0.38 |          |
| VYBjgJx8aZtRhmpy | 10-46           | Warrior | mh   |    61 |  795 | 2.09 |          |
| mJDNgqYVGwb6A8tF | 1-17, 19, 21-22 | Hunter  | mh   |     1 |    7 | 3.94 |          |
| mdXkc6xwP8Y9pR47 | 1               | Warrior | mh   |   253 | 4325 | 1.60 |          |

## Electrocutioner Leg (9446)

- Proc: weapon proc 1292851, logged as 1292851; rolls on main hand
- Sim rate: 4.25 PPM, observed; measured 4.27 PPM [4.03, 4.59], 15 wearers, 15 runs, 2336 procs on 19314 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights           | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1qpQ6FW8fxD9gC4G | 1-7, 9-11, 13-26 | Paladin | mh   |   392 | 3554 | 3.89 |          |
| 27dNt8HLBCPpAmYZ | 1-31             | Rogue   | mh   |   164 | 1387 | 4.17 |          |
| 2jYyAK1RrPvTC8WM | 1-5              | Rogue   | mh   |    64 |  490 | 4.61 |          |
| 2ywhMnJq4Y3zaWtg | 1-9              | Rogue   | mh   |    46 |  402 | 4.04 |          |
| 3B4wmh7JGkQq2Ndj | 1-2, 4-6         | Paladin | mh   |   272 | 2643 | 3.63 |          |
| 3HRvtBnAbfM9xTkK | 1-3              | Warrior | mh   |    75 |  570 | 4.64 |          |
| 6vKFBQNC1fMqR3Wa | 1-6              | Paladin | mh   |   190 | 1544 | 4.34 |          |
| AkBH7pF6XvZc3Gdw | 6-8              | Warrior | mh   |   100 |  628 | 5.62 |          |
| J2XL6AyZqgpjx8QK | 1-13             | Warrior | mh   |   205 | 1520 | 4.76 |          |
| NXgv93kr2G7zyTxm | 1, 3-9           | Rogue   | mh   |   293 | 2323 | 4.45 |          |
| NZn9hmTJHvcMRbfY | 1-2              | Paladin | mh   |   236 | 1856 | 4.49 |          |
| P1QTDgLwfytJKrBh | 1-8              | Rogue   | mh   |    88 |  695 | 4.47 |          |
| TGmKaXrQPp7czjx1 | 1-4              | Rogue   | mh   |   100 |  957 | 3.69 |          |
| aMNVKZpg8tRkhBP3 | 1-17             | Hunter  | mh   |     6 |   60 | 3.54 |          |
| pbvkXazLqcHFhgRf | 4-7              | Warrior | mh   |   105 |  685 | 5.41 |          |

## Fang of the Crystal Spider (13218)

- Proc: weapon proc 17331, logged as 17331; rolls on main hand, off hand
- Sim rate: 1.05 PPM, observed; measured 1.03 PPM [0.85, 1.27], 4 wearers, 4 runs, 91 procs on 3297 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ----- | ---- | ----: | ---: | ---: | -------- |
| QXFd74zDW8JZMmKx | 12-13  | Rogue | oh   |    21 |  602 | 1.31 |          |
| XPYV9J6DfrMKxGqh | 5-6    | Rogue | oh   |    25 | 1006 | 0.93 |          |
| kaH1G4ztmYpgMTJ6 | 1-2    | Rogue | oh   |    10 |  541 | 0.69 |          |
| xXWkB3fqAjt9NJGH | 5      | Rogue | mh   |    35 | 1148 | 1.14 |          |

## Felstriker (12590)

- Proc: weapon proc 16551, logged as 16551; rolls on main hand, off hand
- Sim rate: 0.5 PPM, observed; measured 0.48 PPM [0.35, 0.57], 6 wearers, 7 runs, 114 procs on 8441 hits
- Reports from 2026-11-15 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2wMznDRvmkpbtHxA | 5-7         | Rogue   | oh   |     8 |  867 | 0.33 |          |
| YFtvD87kznr1CThg | 6-7         | Warrior | oh   |    23 | 1145 | 0.71 |          |
| cgMGCdPJQ84wzaBV | 12-13       | Warrior | mh   |    12 |  989 | 0.43 |          |
| jXTVfy3ApWKPYwDG | 1-22        | Warrior | mh   |    39 | 2532 | 0.54 |          |
| nX1qbatTzGZ3CAwB | 4-7         | Rogue   | mh   |    13 | 1268 | 0.36 |          |
| vZqVmLYQa6jwP3A9 | 5-16, 18-21 | Rogue   | oh   |     5 |  623 | 0.28 |          |
| vZqVmLYQa6jwP3A9 | 5-16, 18-21 | Warrior | mh   |    14 | 1018 | 0.49 |          |

## Fiery War Axe (870)

- Proc: weapon proc 18796, logged as 18796; rolls on main hand
- Sim rate: 0.95 PPM, observed; measured 0.94 PPM [0.83, 1.12], 6 wearers, 6 runs, 283 procs on 5009 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2b3Tf7QGKHA1WMxD | 1-3         | Paladin | mh   |    55 |  747 | 1.23 |          |
| 4C28YPXqgTMQJxB7 | 1, 3-9      | Shaman  | mh   |    55 | 1096 | 0.84 |          |
| 4LrMDtG6T8AVRJzQ | 3-11, 14-28 | Shaman  | mh   |    46 |  786 | 0.98 |          |
| 7n6rKTfk1vhzgXyF | 1-2         | Shaman  | mh   |    45 | 1032 | 0.73 |          |
| C3cK4LTfvbhAQxyj | 1-5         | Paladin | mh   |    66 | 1180 | 0.93 |          |
| C8Hk1q3rpN27YWam | 1-3         | Paladin | mh   |    16 |  168 | 1.59 |          |

## Firebelcher (5243)

- Proc: equip proc 1291686, logged as 1291662; rolls on ranged auto attacks
- Sim rate: 3.85 PPM, observed; measured 3.85 PPM [3.46, 4.28], 14 wearers, 14 runs, 306 procs on 2808 hits
- Reports from 2026-11-07 to 2026-11-16

| Report           | Fights | Class   | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ------ | ----: | ---: | ---: | -------- |
| 1RrDAhp7ZvVYc9P3 | 1-6    | Warlock | ranged |    15 |  109 | 4.86 |          |
| 1ZJVMgcYWTyRQh2w | 1-3    | Mage    | ranged |    18 |  110 | 5.78 |          |
| 1xDPwjtGBTzZWm7Y | 1-4    | Priest  | ranged |    35 |  309 | 4.00 |          |
| 287trgyqQLvjACK9 | 1-4    | Priest  | ranged |    10 |   94 | 3.75 |          |
| 2mW9A7wQ4bhv1YMF | 1-3    | Priest  | ranged |    50 |  401 | 4.40 |          |
| 3PnWT1jGz4tagB6r | 1      | Priest  | ranged |    27 |  315 | 3.03 |          |
| 4NWMr2dJPyH9v8az | 1-2    | Priest  | ranged |    20 |  193 | 3.66 |          |
| 6WzAhcVgfLCt78nX | 1-8    | Mage    | ranged |     1 |   14 | 2.52 |          |
| 6wxJrRWgPyQb21hm | 1-7    | Priest  | ranged |    33 |  283 | 4.12 |          |
| LbmB2QdWnPT7JGAF | 1-6    | Priest  | ranged |     4 |   31 | 4.55 |          |
| QHqMJPfgbBCvZan7 | 23-28  | Priest  | ranged |     2 |   16 | 4.41 |          |
| Z9qbhQrJg2Wfz1FM | 1-2    | Warlock | ranged |     5 |   36 | 4.90 |          |
| d6bwAkLKjGq7pVMy | 1      | Warlock | ranged |    23 |  262 | 3.10 |          |
| kBzVYb3Cy7QwnKMD | 1-6    | Priest  | ranged |    63 |  635 | 3.50 |          |

## Firebreather (10797)

- Proc: weapon proc 16413, logged as 16413; rolls on main hand, off hand
- Sim rate: 2.1 PPM, observed; measured 2.10 PPM [1.78, 2.49], 4 wearers, 4 runs, 676 procs on 8800 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| HMdLhWY3pNbnXPtF | 1-16   | Rogue   | oh   |    45 |  749 | 1.64 |          |
| PcVWH9XhGRCvb1Y3 | 1-2    | Rogue   | mh   |   317 | 4267 | 2.03 |          |
| QH63NmDVdWyrqwak | 1-17   | Rogue   | mh   |   134 | 1989 | 1.84 |          |
| ypwYm41tKWZxV69q | 1-5    | Warrior | oh   |   180 | 1794 | 2.74 |          |

## Fist of the Damned (10804)

- Proc: weapon proc 18084, logged as 18084; rolls on main hand, off hand
- Sim rate: 3.45 PPM, observed; measured 3.47 PPM [2.96, 4.17], 7 wearers, 7 runs, 830 procs on 7549 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights           | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 498vhQAtCw2XgZJM | 11, 13-15        | Warrior | mh   |   109 |  922 | 3.73 |          |
| 9MCqcwjzLk4JBARG | 1-5, 7-14, 16-22 | Warrior | mh   |   162 | 1619 | 3.16 |          |
| 9WGcCxKHQ4pkX6Rj | 1-11             | Warrior | oh   |    54 |  323 | 5.28 |          |
| 9hRCG4AnXxvZB3mL | 1-20             | Paladin | mh   |   171 | 1987 | 2.72 |          |
| JMnDtLGxf4FqCV6w | 1-3              | Warrior | oh   |   188 | 1346 | 4.41 |          |
| KYmwz2q1tG8yAcdP | 1-6              | Warrior | mh   |    93 |  679 | 4.33 |          |
| zn64f9D8TVt3wFZq | 1-3              | Paladin | mh   |    53 |  673 | 2.49 |          |

## Flame Wrath (11809)

- Proc: weapon proc 16559, logged as 16559, 16560; rolls on main hand
- Sim rate: 5.3 PPM, observed; measured 5.30 PPM [4.71, 5.93], 1 wearers, 1 runs, 202 procs on 693 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 7 of 8 runs are feral druids, left out (paw speed, not the weapon); one wearer: no wearer interval

| Report           | Fights           | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------- | ------ | ---- | ----: | ---: | ---: | -------- |
| 6yHgkT8LqP42NmxY | 1-17             | Hunter | mh   |   202 |  693 | 5.30 |          |
| 1X37tZRzgkmd9rNP | 1                | Druid  | mh   |   493 | 2710 | 3.31 | feral    |
| 1nTQBK4hmcaLbFZV | 2                | Druid  | mh   |   746 | 3600 | 3.77 | feral    |
| 3R69PBjxZvcfbnX8 | 2-6              | Druid  | mh   |    49 |  433 | 2.06 | feral    |
| 3qfDVTZ1J29zPChn | 1-23             | Druid  | mh   |   775 | 4023 | 3.50 | feral    |
| 4f2ygdYCzFAvqc6K | 10, 12-14, 16-18 | Druid  | mh   |   151 | 1699 | 1.62 | feral    |
| 7ChG2VXaALzcw9pW | 1-2              | Druid  | mh   |   385 | 1254 | 5.58 | feral    |
| k1gAvx7zVaPpLhYC | 1                | Druid  | mh   |   112 | 2232 | 0.91 | feral    |

## Frightalon (14024)

- Proc: weapon proc 19755, logged as 19755; rolls on off hand
- Sim rate: 2.4 PPM, observed; measured 2.39 PPM [2.18, 2.70], 6 wearers, 7 runs, 551 procs on 9862 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------ | ---- | ----: | ---: | ---: | -------- |
| 16hJKBmq2V8tnzXR | 1-3      | Rogue  | oh   |   129 | 2535 | 2.18 |          |
| 3B8AcaGNRKmdgPDj | 5        | Rogue  | oh   |    49 |  705 | 2.98 |          |
| CtMF9BgnAmWkTbfd | 1-33     | Rogue  | oh   |   147 | 2768 | 2.28 |          |
| KG62VXF3a97CxtAg | 5, 7-9   | Rogue  | oh   |    70 | 1007 | 2.98 |          |
| Lxy1kqgVvFc24aPr | 6-11, 13 | Hunter | oh   |    59 | 1339 | 1.89 |          |
| MgmdJqxBWA3Tw2vz | 16-17    | Rogue  | oh   |    16 |  199 | 3.44 |          |
| nX1qbatTzGZ3CAwB | 1, 3-7   | Rogue  | oh   |    81 | 1309 | 2.65 |          |

## Gahz'rilla Fang (9467)

- Proc: weapon proc 3742, logged as 3742; rolls on main hand, off hand
- Sim rate: 2.15 PPM, observed; measured 2.16 PPM [2.01, 2.34], 14 wearers, 14 runs, 700 procs on 10827 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.73 -> 2.16 (144 active minutes, 20% of hits, in 5 runs)

| Report           | Fights     | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2BZkQ4tRcFfdL37H | 1-8, 10-11 | Rogue   | mh   |    69 | 1249 | 1.84 |          |
| 3jqfdPNV4Xwy7g1Q | 1-18       | Rogue   | oh   |    10 |  222 | 1.50 |          |
| 4PtKGL1zcRV8Chfn | 1-2        | Warrior | mh   |    20 |  286 | 2.33 |          |
| 4RVqrpXkh7dKftAC | 1-2        | Hunter  | oh   |    22 |  182 | 4.03 |          |
| 6DZv2FAgWq7YrV3j | 1-19       | Rogue   | oh   |    55 |  846 | 2.17 |          |
| 7F3hKPycMYjXVJkb | 1-12       | Rogue   | oh   |    64 |  922 | 2.31 |          |
| 8tBp2ZK7XT4QJCgD | 1-5        | Rogue   | oh   |    38 |  553 | 2.29 |          |
| A3hT8CQBPJXWmtZM | 1-4        | Warrior | oh   |    51 |  581 | 2.92 |          |
| Mjr1czR7AknXqLvd | 1-8        | Rogue   | oh   |    56 |  921 | 2.03 |          |
| QH63NmDVdWyrqwak | 1-17       | Rogue   | oh   |    32 |  504 | 2.12 |          |
| TbB8q4RMK3LmYJNv | 1-7        | Rogue   | mh   |    98 | 1497 | 2.18 |          |
| XHDjdLb7R9P6NQm1 | 1-2        | Hunter  | mh   |     0 |   14 | 0.00 |          |
| dt4M3ypRnTkwaVQY | 1, 3-10    | Rogue   | mh   |   122 | 2099 | 1.94 |          |
| tac7hXVRgkbZNMDy | 3          | Rogue   | mh   |    63 |  952 | 2.21 |          |

## Galgann's Fireblaster (9412)

- Proc: weapon proc 29644, logged as 29644; rolls on ranged
- Sim rate: 4.25 PPM, observed; measured 4.27 PPM [4.04, 4.51], 11 wearers, 12 runs, 1992 procs on 10778 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights    | Class  | Slot   | Procs | Hits |  PPM | Left out  |
| ---------------- | --------- | ------ | ------ | ----: | ---: | ---: | --------- |
| 24fAZdgMD6K9qx8k | 1-4       | Hunter | ranged |   217 | 1213 | 4.13 |           |
| 2xY1D3mCwagdy7Wc | 1-2, 4-6  | Hunter | ranged |   134 |  736 | 4.20 |           |
| 3LTnjm16YN2bDzCP | 1-6       | Hunter | ranged |     7 |   52 | 3.11 |           |
| 3MY1W4gBv9yXakD6 | 1-8       | Hunter | ranged |   203 | 1204 | 3.89 |           |
| 3WyYMqr8ZxQRN1vD | 1-7       | Hunter | ranged |   248 | 1415 | 4.04 |           |
| 9XQfBmZRaNWJkH3F | 1-20      | Hunter | ranged |   199 |  919 | 5.00 |           |
| AKXHxFdgavCh9fkY | 1-16      | Hunter | ranged |   229 | 1299 | 4.07 |           |
| CpDKr6AWJ4fw7xhg | 5-7       | Hunter | ranged |    67 |  390 | 3.96 |           |
| GYNx3nzHdpr7X1AW | 2-10      | Hunter | ranged |    71 |  370 | 4.43 |           |
| aANyFDgb69hxqY2w | 3         | Hunter | ranged |   148 |  894 | 3.82 |           |
| pJjZqyKxb81RYdvG | 1-3, 5-16 | Hunter | ranged |   181 |  844 | 4.95 |           |
| y3wkJZnA7gv4qzfX | 1-30      | Hunter | ranged |   288 | 1442 | 4.61 |           |
| DBQHpyFz7chZRM6j | 2-10      | Hunter | ranged |    63 |  321 | 4.53 | duplicate |

## Glutton's Cleaver (10772)

- Proc: weapon proc 18075, logged as 18075; rolls on main hand, off hand
- Sim rate: 2.3 PPM, observed; measured 2.28 PPM [2.06, 2.55], 8 wearers, 8 runs, 363 procs on 4779 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 6VJxcqnYa4wdBWQv | 1-9    | Rogue   | oh   |    32 |  445 | 2.16 |          |
| 7tCZB31vV2Jp6XWd | 1-2    | Rogue   | oh   |    43 |  617 | 2.09 |          |
| 8vDJ9jmHd4FA7Twt | 1-2    | Rogue   | oh   |    45 |  644 | 2.10 |          |
| CnNwpaYW83cJr6fR | 1-6    | Warrior | oh   |    60 |  569 | 3.16 |          |
| L7JpRfWtAZ6T1hGx | 3, 5   | Rogue   | oh   |    16 |  263 | 1.82 |          |
| P8Znjv1fgBD2CAVY | 2-12   | Warrior | mh   |    76 | 1009 | 2.26 |          |
| PDT2yACNzd4LMZ98 | 1-14   | Rogue   | oh   |    42 |  611 | 2.06 |          |
| qnyc1zwaB4FrvM9x | 1      | Warrior | oh   |    49 |  621 | 2.37 |          |

## Grim Reaper (13054)

- Proc: weapon proc 14126, logged as 14126; rolls on main hand
- Sim rate: 1.1 PPM, observed; measured 1.10 PPM [0.97, 1.24], 5 wearers, 5 runs, 233 procs on 4113 hits
- Reports from 2026-11-09 to 2026-11-16
- Notes: 1 of 6 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 4HC91qxpR6hzKWdF | 1-2    | Paladin | mh   |    48 |  810 | 1.15 |          |
| 8dZb4CmyvLaWMGf1 | 8-16   | Paladin | mh   |    59 | 1073 | 1.06 |          |
| D9vw4QTYJZWMPnph | 1-3    | Hunter  | mh   |     6 |   95 | 1.22 |          |
| NXgv93kr2G7zyTxm | 3-9    | Paladin | mh   |    94 | 1602 | 1.14 |          |
| YJR78AMPDNKGQWjF | 5-12   | Warrior | mh   |    26 |  533 | 0.94 |          |
| LBZqKhR7WjvA31NH | 1-4    | Druid   | mh   |    17 | 2190 | 0.15 | feral    |

## Grimclaw (1481)

- Proc: weapon proc 1292158, logged as 1292158; rolls on main hand, off hand
- Sim rate: 2.2 PPM, observed; measured 2.22 PPM [1.65, 2.48], 6 wearers, 6 runs, 292 procs on 3949 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights                 | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 6rgDRnVFX9a3pbGy | 1-4                    | Warrior | oh   |    27 |  490 | 1.65 |          |
| F1MPmrQNq4v6XJ23 | 1-2                    | Paladin | mh   |    26 |  543 | 1.44 |          |
| HJxqkQPAnmKtwNjd | 1-16, 18, 20-24, 26-31 | Hunter  | oh   |    16 |  195 | 2.46 |          |
| JgL3qFZA4aQcpXPC | 1, 3-9, 11-13          | Warrior | mh   |   172 | 2041 | 2.53 |          |
| KjLXqdCZANvQc9BT | 1-3                    | Rogue   | oh   |    44 |  575 | 2.29 |          |
| Lcy3V1wHPpqZQAG7 | 2, 4-12                | Rogue   | oh   |     7 |  104 | 2.02 |          |

## Gryphon Rider's Stormhammer (9651)

- Proc: weapon proc 18081, logged as 18081; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.95 PPM [1.64, 2.21], 5 wearers, 6 runs, 567 procs on 6458 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights                                                             | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------------------------------------------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 8bWqTVN2QCvK7HPR | 1-13                                                               | Warrior | mh   |   106 | 1057 | 2.23 |          |
| 9cQz2ymYMgfL7TwK | 132-140, 142-143, 146-155, 157-159, 161, 163-174, 176-181, 183-190 | Rogue   | mh   |    73 |  746 | 2.17 |          |
| A3hT8CQBPJXWmtZM | 1-4                                                                | Warrior | mh   |    68 | 1094 | 1.38 |          |
| Mjr1czR7AknXqLvd | 1-8                                                                | Rogue   | mh   |   136 | 1573 | 1.92 |          |
| cx8pPRvN9Ta2zHg1 | 1-10                                                               | Warrior | mh   |    79 |  683 | 2.57 |          |
| mNBAdW4ajKHzhMYZ | 1, 3-5                                                             | Rogue   | mh   |   105 | 1304 | 1.79 |          |

## Gut Ripper (2164)

- Proc: weapon proc 18107, logged as 18107; rolls on main hand, off hand
- Sim rate: 1.5 PPM, measured; measured 1.52 PPM [1.38, 1.68], 9 wearers, 9 runs, 383 procs on 8400 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights   | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------ | ---- | ----: | ---: | ---: | -------- |
| 8tBp2ZK7XT4QJCgD | 1-5      | Rogue  | mh   |    37 |  910 | 1.35 |          |
| Ag69GWdLRDhKCrcT | 1-6      | Rogue  | mh   |    74 | 1715 | 1.44 |          |
| H1AnQVTYJzcam79x | 1-8      | Rogue  | oh   |    22 |  642 | 1.14 |          |
| VG18w7rW3KRQa9B4 | 4-10     | Rogue  | mh   |    51 | 1054 | 1.61 |          |
| WaHMR9NKLrwbzptn | 1-2, 4-7 | Rogue  | mh   |    69 | 1458 | 1.58 |          |
| fz2HQxXk96ywVYv7 | 3-8      | Hunter | oh   |     7 |  103 | 2.27 |          |
| g41xhtdFCXZ8Vr9z | 1-2      | Rogue  | mh   |    62 | 1233 | 1.68 |          |
| qNyQvka41YhVmdHR | 1-4      | Hunter | oh   |    37 |  666 | 1.85 |          |
| v2aNxQ7mGTdkJryV | 1-7      | Rogue  | oh   |    24 |  619 | 1.29 |          |

## Hameya's Slayer (15814)

- Proc: weapon proc 16406, logged as 16406; rolls on main hand, off hand
- Sim rate: 1.05 PPM, observed; measured 1.07 PPM [0.88, 3.04], 3 wearers, 3 runs, 104 procs on 2926 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| 9zxHaWbYT3gqDQvF | 3-18   | Hunter | mh   |     2 |   20 | 3.04 |          |
| PcVWH9XhGRCvb1Y3 | 1-2    | Rogue  | oh   |    69 | 1952 | 1.06 |          |
| vFQhXgw42KmfjT8W | 1-3    | Rogue  | oh   |    33 |  954 | 1.04 |          |

## Hammer of Expertise (8708)

- Proc: weapon proc 1317432, logged as 1317432; rolls on main hand
- Sim rate: 2 PPM, observed; measured 2.00 PPM [1.75, 2.23], 6 wearers, 6 runs, 568 procs on 8126 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| BQnXb3hTC4rmJ2pd | 1-2    | Warrior | mh   |    82 | 1323 | 1.77 |          |
| J97GvmrRXCFZfQ8a | 1-2    | Warrior | mh   |   112 | 1333 | 2.40 |          |
| LaJzMkFfhmpc2DbH | 7-10   | Paladin | mh   |    95 | 1204 | 2.25 |          |
| YWZm7NKhFvqbMfgy | 1-6    | Paladin | mh   |   117 | 1618 | 2.07 |          |
| dGXqkVhwT17nHfBc | 1-2    | Warrior | mh   |    79 | 1465 | 1.54 |          |
| jQPc3Xy1hW87gfdD | 1-2    | Paladin | mh   |    83 | 1183 | 2.00 |          |

## Hammer of the Northern Wind (810)

- Proc: weapon proc 13439, logged as 13439; rolls on main hand
- Sim rate: 4.6 PPM, observed; measured 4.58 PPM [4.04, 4.87], 4 wearers, 5 runs, 1619 procs on 10100 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 6kf3tMF4wCPNcRnh | 1-33   | Paladin | mh   |   588 | 3481 | 4.83 |          |
| J38xbrFytgM9p1LT | 7-19   | Warrior | mh   |   122 | 1074 | 3.24 |          |
| P4bXMwRjD1zyhZdJ | 1-2    | Warrior | mh   |   598 | 3702 | 4.62 |          |
| TF6dc3NCHbKrfxJt | 1-6    | Warrior | mh   |   163 | 1010 | 4.61 |          |
| dBNfgVDLMmnc8h4v | 10     | Warrior | mh   |   148 |  833 | 5.08 |          |

## Hanzo Sword (8190)

- Proc: weapon proc 16405, logged as 16405; rolls on main hand, off hand
- Sim rate: 1.25 PPM, observed; measured 1.23 PPM [1.08, 1.40], 7 wearers, 7 runs, 222 procs on 7212 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights                | Class   | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | --------------------- | ------- | ---- | ----: | ---: | ---: | --------- |
| 4f2ygdYCzFAvqc6K | 6-8, 10, 12-14, 16-18 | Rogue   | oh   |    26 |  863 | 1.21 |           |
| 6xHzpKd8wCna37Lb | 1-14                  | Warrior | mh   |    41 | 1112 | 1.47 |           |
| 824kNhdFaALMBnvR | 1-4, 7-18, 20-45      | Warrior | oh   |    17 |  514 | 1.32 |           |
| DNc6QALhxmGbCH3Y | 1-6                   | Rogue   | oh   |    42 | 1428 | 1.18 |           |
| KPqgbZxhn6BXwCvJ | 1                     | Rogue   | oh   |    40 | 1282 | 1.25 |           |
| MK6XJDzfLpFjnRWc | 1-5                   | Rogue   | oh   |    36 | 1082 | 1.33 |           |
| qNyQvka41YhVmdHR | 4                     | Rogue   | oh   |    20 |  932 | 0.86 |           |
| 2DC7xQNm9wHJnVF4 | 6-7                   | Rogue   | oh   |    26 |  863 | 1.21 | duplicate |

## Iceblade Hacker (13952)

- Proc: weapon proc 1298416, logged as 1298416; rolls on main hand
- Sim rate: 1.95 PPM, measured; measured 1.95 PPM [1.75, 2.11], 5 wearers, 7 runs, 1299 procs on 19965 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights                                                                                                                                  | Class   | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------- | ---- | ----: | ---: | ---: | --------- |
| 4h1PjwkaKX9q8mgT | 1-2                                                                                                                                     | Paladin | mh   |   161 | 2424 | 1.99 |           |
| A1yxMTmKatkf3HRJ | 15-17                                                                                                                                   | Rogue   | mh   |    10 |  324 | 0.93 |           |
| CWhG6LyXk3PBHdMQ | 1-2                                                                                                                                     | Paladin | mh   |   198 | 3526 | 1.68 |           |
| DNc6QALhxmGbCH3Y | 1-6                                                                                                                                     | Rogue   | mh   |    34 |  550 | 1.85 |           |
| HxCQat38vmjnkpb6 | 1-7, 9-20, 22-33, 35-48, 50-67, 69-86, 88-157, 159-185, 187-212, 214-241, 243-458, 460-498, 500-587, 589-639, 641-692, 694-802, 804-830 | Paladin | mh   |   525 | 7982 | 1.97 |           |
| KYmFPtyJ3gkzHfVM | 102-160                                                                                                                                 | Rogue   | mh   |   188 | 2237 | 2.52 |           |
| VLz78BfZyWa3Cxr2 | 1-7                                                                                                                                     | Paladin | mh   |   183 | 2922 | 1.88 |           |
| JrqDZX4LT6NpxzM7 | 1-6                                                                                                                                     | Paladin | mh   |   498 | 7368 | 2.03 | duplicate |

## Iron Knuckles (2942)

- Proc: weapon proc 13491, logged as 13491; rolls on main hand, off hand
- Sim rate: 2.4 PPM, observed; measured 2.40 PPM [2.08, 2.66], 5 wearers, 6 runs, 327 procs on 4812 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 38D6JN4nZzjcGtHg | 1-4    | Warrior | mh   |    64 |  912 | 2.48 |          |
| 4QAtNfTXMyDGPqCr | 1-13   | Rogue   | mh   |    21 |  341 | 2.18 |          |
| AzbN8GY9DnBrH4aW | 19-31  | Warrior | mh   |    46 |  788 | 2.06 |          |
| DwWM9H3rpTnGPfbt | 1-14   | Warrior | mh   |    80 |  989 | 2.85 |          |
| GRTmQn3VDpwZKbCA | 1-6, 8 | Warrior | mh   |   100 | 1508 | 2.34 |          |
| H3vdcCr6JX1gwYtq | 1-3    | Rogue   | oh   |    16 |  274 | 2.06 |          |

## Ironforge Greathammer (279895)

- Proc: weapon proc 1322295, logged as 1322295; rolls on main hand
- Sim rate: 3.3 PPM, measured; measured 3.29 PPM [3.03, 3.63], 18 wearers, 18 runs, 1549 procs on 9109 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 11 of 29 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 17jHXDAWBRQdgqV6 | 1-4    | Shaman  | mh   |    41 |  296 | 2.68 |          |
| 1HK7gFbwy8aJPAh6 | 1-3    | Paladin | mh   |   157 | 1025 | 2.96 |          |
| 1HK7gFbwy8aJPAh6 | 1-3    | Warrior | mh   |    97 |  595 | 3.16 |          |
| 1fpaDyG3mFnB9KW8 | 1      | Shaman  | mh   |     7 |   71 | 1.91 |          |
| 1h9P4rYzWDvtHRxb | 1      | Warrior | mh   |   130 |  656 | 3.84 |          |
| 1pDcTvkrCBhZJRgf | 1-14   | Shaman  | mh   |   106 |  673 | 3.05 |          |
| 7kYG3K21FmyMVABD | 1-4    | Shaman  | mh   |    11 |  106 | 2.01 |          |
| CdH9vgqPbRX1JKM3 | 1      | Warrior | mh   |   132 |  519 | 4.92 |          |
| HKQW3wCmr4kJXz2j | 1-3    | Shaman  | mh   |    28 |  167 | 3.25 |          |
| L8yaqgDXG26Nf9v1 | 1-2    | Paladin | mh   |    56 |  386 | 2.81 |          |
| LVwmRx1Mqak7vp4y | 1-8    | Warrior | mh   |    50 |  308 | 3.14 |          |
| QmCHxRYD9tbjLfW1 | 1-4    | Warrior | mh   |    78 |  332 | 4.55 |          |
| QmCHxRYD9tbjLfW1 | 1-4    | Shaman  | mh   |    46 |  326 | 2.73 |          |
| RQxLGd1agBDt6hXP | 1-4    | Paladin | mh   |   216 | 1172 | 3.57 |          |
| ZADc3bTCKhn9jJdz | 1-3    | Shaman  | mh   |    39 |  223 | 3.38 |          |
| gXLj6knBf4QFvGD7 | 3-5    | Paladin | mh   |   126 |  714 | 3.42 |          |
| jr1xaAPwb4f2HpvK | 1-3    | Paladin | mh   |   156 | 1095 | 2.76 |          |
| q2h6pRvdBCKMabnL | 1-2    | Warrior | mh   |    73 |  445 | 3.18 |          |
| 1K6c48bWMN9nzCZA | 1-3    | Druid   | mh   |    28 |  330 | 1.64 | feral    |
| 1K6c48bWMN9nzCZA | 1-3    | Druid   | mh   |    38 | 1104 | 0.67 | feral    |
| 7kYG3K21FmyMVABD | 1-4    | Druid   | mh   |    48 | 1172 | 0.79 | feral    |
| 7kYG3K21FmyMVABD | 1-4    | Druid   | mh   |     8 | 1448 | 0.11 | feral    |
| JGFcm9zq48CArhgk | 1-6    | Druid   | mh   |   141 | 3753 | 0.73 | feral    |
| LKzDgyYcfakX38xP | 1      | Druid   | mh   |    29 |  997 | 0.56 | feral    |
| LKzDgyYcfakX38xP | 1      | Druid   | mh   |     4 |  133 | 0.58 | feral    |
| LVwmRx1Mqak7vp4y | 8      | Druid   | mh   |    20 |  343 | 1.13 | feral    |
| T7WjV2vKaydLkMzX | 1-5    | Druid   | mh   |    16 |  682 | 0.45 | feral    |
| dJb4v1ZQNgtak9nB | 1-4    | Druid   | mh   |    56 | 1548 | 0.70 | feral    |
| y2dDThApkr8MVRcj | 1-3, 5 | Druid   | mh   |    14 |  109 | 2.49 | feral    |

## Jom-jom's Gabbar (274952)

- Proc: weapon proc 1295736, logged as 1295736; rolls on main hand, off hand
- Sim rate: 1.75 PPM, observed; measured 1.76 PPM [1.45, 2.08], 7 wearers, 7 runs, 389 procs on 6972 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.47 -> 1.76 (49 active minutes, 17% of hits, in 2 runs)

| Report           | Fights    | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------ | ---- | ----: | ---: | ---: | -------- |
| 3TDwbg41YhvNpPAq | 1-7       | Rogue  | oh   |     1 |   35 | 0.90 |          |
| 987q1pjnYdVQbCmT | 1-6       | Rogue  | mh   |   105 | 1551 | 2.14 |          |
| BqdMTtCvN3fVJpRk | 29-35     | Rogue  | mh   |    79 | 1244 | 2.01 |          |
| HndJP9Yf4xc7VzKw | 1         | Hunter | oh   |     1 |   14 | 2.19 |          |
| NfqRryF7gnMcX2ak | 1-2       | Rogue  | oh   |    47 |  739 | 2.01 |          |
| Pjz8kK2XvxyQrTCB | 1-2, 4-16 | Rogue  | mh   |    89 | 2120 | 1.33 |          |
| cAYgpkdqXRDynzPK | 1-3       | Rogue  | mh   |    67 | 1269 | 1.67 |          |

## Julie's Dagger (6660)

- Proc: weapon proc 8348, logged as 8348; rolls on off hand
- Sim rate: 0.85 PPM, observed; measured 0.86 PPM [0.66, 1.12], 3 wearers, 3 runs, 54 procs on 2893 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ----- | ---- | ----: | ---: | ---: | -------- |
| 6kN78G2xfqy3vwTP | 6      | Rogue | oh   |     1 |   54 | 0.86 |          |
| JMnDtLGxf4FqCV6w | 1-3    | Rogue | oh   |    42 | 2247 | 0.86 |          |
| dJa2G7PHZg9wt6QX | 1-4    | Rogue | oh   |    11 |  592 | 0.86 |          |

## Kang the Decapitator (2291)

- Proc: weapon proc 17153, logged as 17153; rolls on main hand
- Sim rate: 1 PPM, observed; measured 1.00 PPM [0.91, 1.12], 8 wearers, 9 runs, 587 procs on 9771 hits
- Reports from 2026-11-11 to 2026-11-16

| Report           | Fights             | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 2FkYRHB1dWhra8vt | 1-17               | Shaman  | mh   |    35 |  785 | 0.74 |          |
| 6kf3tMF4wCPNcRnh | 1-33               | Shaman  | mh   |    78 | 1255 | 1.04 |          |
| 9Z473MfXLYqcyWPg | 1-2                | Shaman  | mh   |    28 |  646 | 0.72 |          |
| FHqYCNcKntvDAjxg | 4-20, 22-31, 33-36 | Paladin | mh   |    50 |  577 | 1.44 |          |
| FMaAjHcwCQLkpV32 | 1-4                | Shaman  | mh   |    83 | 1323 | 1.05 |          |
| RfT7K4v2jhY9FAzD | 1-3                | Shaman  | mh   |    60 | 1169 | 0.86 |          |
| RpFQn1ftYCAgMW47 | 1-17               | Shaman  | mh   |    88 | 1590 | 0.92 |          |
| W78tmfknBKz46hRZ | 1-2                | Shaman  | mh   |    85 | 1386 | 1.02 |          |
| ZNyPLR2vkjMGptCJ | 1-2                | Shaman  | mh   |    80 | 1040 | 1.28 |          |

## Killmaim (13016)

- Proc: weapon proc 13318, logged as 13318; rolls on main hand
- Sim rate: 0.95 PPM, observed; measured 0.97 PPM [0.82, 1.14], 7 wearers, 8 runs, 139 procs on 2692 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 0.66 -> 0.97 (73 active minutes, 32% of hits, in 3 runs)

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | --------- |
| 29A3dKwPgpRNnWvc | 7-11   | Shaman  | mh   |     6 |  125 | 0.90 |           |
| 2aVnTxFzfckdXpBy | 1-7    | Warrior | mh   |    17 |  351 | 0.91 |           |
| 3xgPkwfn8yQ9BqAW | 1-4    | Warrior | mh   |    29 |  449 | 1.21 |           |
| 6v4wTWQkjb3Xd7mM | 6      | Warrior | mh   |     4 |   49 | 1.53 |           |
| P16dx3VLmRyaDJZQ | 1-18   | Shaman  | mh   |    31 |  561 | 1.04 |           |
| P7cV2HjLgxGpmZQk | 1-7    | Shaman  | mh   |    14 |  313 | 0.84 |           |
| d7DFtXL2jfNa8Bw4 | 1-5    | Shaman  | mh   |    28 |  529 | 0.99 |           |
| zJ6C2148dTYfaXHb | 2-4    | Shaman  | mh   |    10 |  315 | 0.60 |           |
| 9mN62kzgMdKypTqf | 1-6    | Shaman  | mh   |    28 |  729 | 0.72 | duplicate |

## Linken's Sword of Mastery (11902)

- Proc: weapon proc 18089, logged as 18089; rolls on main hand
- Sim rate: 2.05 PPM, observed; measured 2.05 PPM [1.83, 2.25], 4 wearers, 4 runs, 431 procs on 7005 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights             | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 17ymZpJ6xFPf4Raj | 1                  | Rogue   | mh   |    98 | 1561 | 2.09 |          |
| CMGKZHRVPwzvFmqy | 1, 3-5             | Warrior | mh   |   105 | 1673 | 2.09 |          |
| DNfvpnLRcPJyMKQA | 1                  | Paladin | mh   |    82 | 1579 | 1.73 |          |
| vwYj6cK8Xt24QLaN | 1-22, 24-26, 29-32 | Rogue   | mh   |   146 | 2192 | 2.22 |          |

## Loksey's Training Stick (7710)

- Proc: weapon proc 1293482, logged as 1293482; rolls on main hand
- Sim rate: 2.55 PPM, observed; measured 2.53 PPM [1.79, 2.84], 14 wearers, 14 runs, 252 procs on 1928 hits
- Reports from 2026-11-10 to 2026-11-16
- Notes: 8 of 22 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights                                                                                                                             | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1aBhkzdDMZKGtb37 | 162-170, 172-176, 178-187, 189-195, 197-237, 239-286, 288-289, 291-295, 297-312, 314-323, 325-328, 331, 333-335, 337-371, 374, 376 | Hunter  | mh   |     1 |    2 | 9.68 |          |
| 1xqBcj4CZQXV38Pv | 1-2                                                                                                                                | Warrior | mh   |    89 |  644 | 2.67 |          |
| 214ndcFyjpPM8wah | 1-3, 6-9, 11-15, 19-28, 30                                                                                                         | Warrior | mh   |   125 |  936 | 2.58 |          |
| 27jZGhfHgp1LR8MB | 1-11                                                                                                                               | Hunter  | mh   |     3 |   46 | 1.26 |          |
| 2PpAjf1qRmTv6gDz | 1-17                                                                                                                               | Hunter  | mh   |     2 |   27 | 1.43 |          |
| 2WH4jLt3FADy8fnh | 1-2, 5-6                                                                                                                           | Hunter  | mh   |     1 |   22 | 0.88 |          |
| CpDKr6AWJ4fw7xhg | 5-7                                                                                                                                | Hunter  | mh   |     5 |   51 | 1.90 |          |
| GBDAWjKNh6Q9XnF7 | 1-4                                                                                                                                | Hunter  | mh   |     1 |   20 | 0.97 |          |
| PmR8Lpvazd4q61nA | 6-12                                                                                                                               | Hunter  | mh   |     2 |   11 | 3.52 |          |
| Vb3ZxfXtyDWr1ndP | 1                                                                                                                                  | Hunter  | mh   |     0 |    2 | 0.00 |          |
| WYrCyJkdfMXw84Ph | 13-21, 23-35, 37-39, 41-48                                                                                                         | Hunter  | mh   |     6 |   40 | 2.90 |          |
| bCzt1Fq24mYcPXkB | 1-2                                                                                                                                | Hunter  | mh   |     3 |   60 | 0.97 |          |
| pqfK2AgHZ7DnhvPF | 1                                                                                                                                  | Hunter  | mh   |    11 |   42 | 5.07 |          |
| rD8m4JW1bVyYPKGX | 1-3, 6                                                                                                                             | Hunter  | mh   |     3 |   25 | 2.32 |          |
| 127tXgbzdMBZYKp9 | 1-17, 19-20                                                                                                                        | Druid   | mh   |    25 |  407 | 1.19 | feral    |
| 1FV3WyfXGCcPrT9b | 1, 3-8                                                                                                                             | Druid   | mh   |    87 | 3858 | 0.44 | feral    |
| 3zxaGvgWKLFArhjk | 4-8                                                                                                                                | Druid   | mh   |    36 | 1810 | 0.38 | feral    |
| 6xd1aP4McnjKyqTG | 1-2, 4-10                                                                                                                          | Druid   | mh   |    10 |  824 | 0.23 | feral    |
| 7CmgWKkj3Hx1qd4c | 2, 4-7, 9-10                                                                                                                       | Druid   | mh   |    40 | 2078 | 0.37 | feral    |
| 8CzwkcFWrA91QaxG | 1-12                                                                                                                               | Druid   | mh   |    33 | 2361 | 0.27 | feral    |
| PFGvpDf9WZgcY6Rk | 1-5                                                                                                                                | Druid   | mh   |    50 | 2793 | 0.35 | feral    |
| Z9DMkXc2KqNAxVtr | 1-12                                                                                                                               | Druid   | mh   |    37 | 1552 | 0.46 | feral    |

## Lord General's Sword (11817)

- Proc: weapon proc 15602, logged as 15602; rolls on main hand
- Sim rate: 2.1 PPM, measured; measured 2.08 PPM [1.91, 2.19], 7 wearers, 9 runs, 1267 procs on 14069 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights                       | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 24kgnZAXbKTW89x6 | 1-2                          | Warrior | mh   |    14 |   93 | 3.47 |          |
| ARFXzbqHDpBLJdhy | 1, 3-4                       | Rogue   | mh   |   185 | 2040 | 2.09 |          |
| CDKVcP6ghZJtv8AW | 3-6, 8-16                    | Rogue   | mh   |    91 | 1107 | 1.90 |          |
| MwVH1KQmak73NPd8 | 8-16                         | Warrior | mh   |   100 | 1326 | 1.74 |          |
| NXMZV9xjvWpB7GtP | 10, 12-107, 109-116, 118-122 | Warrior | mh   |    97 | 1305 | 1.72 |          |
| NcxbzW4kJYQ7DGhj | 1-3                          | Rogue   | mh   |    99 | 1014 | 2.25 |          |
| RVkGfcrbZ13MWqQ2 | 1-3                          | Rogue   | mh   |   396 | 4242 | 2.15 |          |
| X8z9dgD6HRTZ24Vc | 23                           | Warrior | mh   |    33 |  252 | 3.02 |          |
| pgN9QwMLYVfrqGbv | 1-4                          | Rogue   | mh   |   252 | 2691 | 2.16 |          |

## Mana Channeling Wand (18483)

- Proc: equip proc 1302348, logged as 1302350; rolls on ranged auto attacks
- Sim rate: 4.25 PPM, observed; measured 4.26 PPM [3.00, 4.92], 5 wearers, 5 runs, 160 procs on 1408 hits
- Reports from 2026-11-11 to 2026-11-15

| Report           | Fights    | Class  | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------ | ------ | ----: | ---: | ---: | -------- |
| QXFd74zDW8JZMmKx | 1-2, 4-13 | Mage   | ranged |     6 |   48 | 4.69 |          |
| Wv4QYA6C7antxKXr | 1         | Priest | ranged |     2 |   22 | 3.41 |          |
| fjtT3JVxbGPz6RwX | 6         | Priest | ranged |     4 |   26 | 5.77 |          |
| gBDMLv8J1K7nqkpd | 112-157   | Priest | ranged |   121 |  960 | 4.73 |          |
| wkJXWxD23VMqThLg | 14-17     | Priest | ranged |    27 |  352 | 2.88 |          |

## Naga Heartpiercer (3078)

- Proc: weapon proc 1292647, logged as 1292647; rolls on ranged
- Sim rate: 2.2 PPM, observed; measured 2.22 PPM [2.05, 2.43], 13 wearers, 14 runs, 965 procs on 10882 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class  | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ------ | ----: | ---: | ---: | -------- |
| 1WZ8Dn6mpQdCNxAT | 1-8    | Hunter | ranged |    71 |  594 | 2.99 |          |
| 1a3pDtCcPgBkL6wm | 1-22   | Hunter | ranged |   114 | 1041 | 2.74 |          |
| 1fhNLGvKgMzt8Ac3 | 1, 3   | Hunter | ranged |    31 |  326 | 2.38 |          |
| 1h9P4rYzWDvtHRxb | 3-4    | Hunter | ranged |    68 |  902 | 1.88 |          |
| 1hpgVjWH3GZBLYKw | 1-12   | Hunter | ranged |    85 |  912 | 2.33 |          |
| 1x8aqzFcL7ydGNgP | 1-2    | Hunter | ranged |    29 |  267 | 2.72 |          |
| BKnf4rNwJdcqpQ18 | 1-6    | Hunter | ranged |   108 | 1311 | 2.06 |          |
| BzGXfjMLJ2RyCwHm | 1-7    | Hunter | ranged |    85 |  891 | 2.38 |          |
| MwA4Vhx7zDfCdY6B | 1-2    | Hunter | ranged |    56 |  703 | 1.99 |          |
| aMNVKZpg8tRkhBP3 | 1-17   | Hunter | ranged |    51 |  660 | 1.93 |          |
| aykDWnCGx4jJFTV6 | 1-6    | Hunter | ranged |    54 |  806 | 1.67 |          |
| hPWNkH67AYCBFLgM | 1, 3-8 | Hunter | ranged |    55 |  652 | 2.11 |          |
| hcGCaDW9tRYFB6Ax | 1-11   | Hunter | ranged |   125 | 1505 | 2.08 |          |
| r7cZyhBXWfNwj2GY | 1-10   | Hunter | ranged |    33 |  312 | 2.64 |          |

## Night Reaver (1318)

- Proc: weapon proc 13480, logged as 13480; rolls on main hand
- Sim rate: 1.9 PPM, observed; measured 1.90 PPM [1.54, 2.40], 5 wearers, 5 runs, 218 procs on 2091 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights            | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 82YfmdLnzpAyHWwN | 1-9, 11-31, 33-50 | Paladin | mh   |    77 |  774 | 1.81 |          |
| MZRL4VdF6XtNf8B3 | 1-2               | Paladin | mh   |    32 |  420 | 1.39 |          |
| QYP1JCm2FMydzqjx | 1                 | Paladin | mh   |    47 |  410 | 2.08 |          |
| T72D1mwvJnabPtqK | 1-3               | Shaman  | mh   |    19 |  239 | 1.45 |          |
| ac64wQx7yZ9tRNfW | 1-3               | Warrior | mh   |    43 |  248 | 3.15 |          |

## Nightblade (1982)

- Proc: weapon proc 18211, logged as 18211; rolls on main hand
- Sim rate: 2 PPM, measured; measured 2.00 PPM [1.91, 2.10], 8 wearers, 10 runs, 1530 procs on 13083 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1NBKq2mDWghwR69z | 4, 6      | Paladin | mh   |   110 | 1101 | 1.71 |          |
| 4PtKGL1zcRV8Chfn | 1-3       | Paladin | mh   |   156 | 1171 | 2.28 |          |
| 7BQKZXWzY9MaxAPq | 1-4       | Paladin | mh   |    39 |  375 | 1.78 |          |
| 7c21XafhWTAdqNGY | 1-10      | Paladin | mh   |   213 | 1691 | 2.16 |          |
| 8cVPaCpAMmv93bkD | 1-9       | Paladin | mh   |   121 | 1039 | 2.00 |          |
| LPQd7pfbjDkYZyRz | 1-2, 4-5  | Paladin | mh   |   106 |  959 | 1.89 |          |
| Lg82ZFDyQ4MvRm9p | 5-11, 13  | Paladin | mh   |   140 | 1317 | 1.82 |          |
| P1LGvYV9FdmR8x3n | 1         | Paladin | mh   |   111 |  864 | 2.20 |          |
| aANyFDgb69hxqY2w | 3         | Paladin | mh   |   122 | 1010 | 2.07 |          |
| ynVw7ZgMx4tjF3fQ | 1-6, 8-18 | Paladin | mh   |   412 | 3556 | 1.99 |          |

## Pendulum of Doom (9425)

- Proc: weapon proc 10373, logged as 10373; rolls on main hand
- Sim rate: 1.05 PPM, measured; measured 1.05 PPM [0.98, 1.15], 6 wearers, 7 runs, 645 procs on 9187 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 9pMcvgxrQNfLZ7hd | 1-15        | Warrior | mh   |     0 |  783 | 0.00 |          |
| CR6khYVnQAp3HaTy | 1-2, 4-6    | Paladin | mh   |     0 |  447 | 0.00 |          |
| Pjz8kK2XvxyQrTCB | 4-16        | Shaman  | mh   |     0 |  677 | 0.00 |          |
| WygY76wT4zqbvCM2 | 1-11, 13-16 | Paladin | mh   |     0 |  696 | 0.00 |          |
| X186Gx2gRBZChbVA | 1-7         | Shaman  | mh   |     0 |  410 | 0.00 |          |
| ga9Qk6DcjJmNTGVH | 2-3         | Warrior | mh   |     0 |  141 | 0.00 |          |
| qw3ZzkBmQYW7pXVH | 1-19        | Warrior | mh   |     0 |  792 | 0.00 |          |

## Phantom Blade (7961)

- Proc: weapon proc 9806, logged as 9806; rolls on main hand, off hand
- Sim rate: 2.25 PPM, observed; measured 2.25 PPM [1.91, 2.65], 4 wearers, 4 runs, 218 procs on 3631 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights  | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2YfdbFMgGhr7zHwQ | 3-4     | Rogue   | oh   |    47 |  882 | 2.00 |          |
| 6kN78G2xfqy3vwTP | 1-6     | Warrior | mh   |    69 | 1022 | 2.53 |          |
| XFdcWaPL8MthYH6A | 1-5     | Rogue   | oh   |    49 |  649 | 2.83 |          |
| dt4M3ypRnTkwaVQY | 1, 3-10 | Rogue   | oh   |    53 | 1077 | 1.84 |          |

## Poison-tipped Bone Spear (1726)

- Proc: weapon proc 16401, logged as 16401; rolls on main hand
- Sim rate: 0.85 PPM, observed; measured 0.87 PPM [0.72, 1.11], 4 wearers, 5 runs, 117 procs on 3517 hits
- Reports from 2026-11-15 to 2026-11-16
- Notes: 2 of 7 runs are feral druids, left out (paw speed, not the weapon); few wearers (4): the interval is about the per-wearer range

| Report           | Fights                    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| GgmB9Yhp8fMPWjVN | 3-11, 13-16, 18-22, 24-30 | Hunter  | mh   |     6 |  186 | 0.84 |          |
| MzwqjvXHFKZap19g | 1, 4                      | Hunter  | mh   |    32 |  869 | 0.96 |          |
| TCk3ZaghzP1wJdXK | 1                         | Hunter  | mh   |    28 | 1030 | 0.71 |          |
| acpjMDfJkLPABvxT | 1                         | Paladin | mh   |    27 |  593 | 1.19 |          |
| rwxQgM38abAcmRVv | 1-6, 8-9                  | Hunter  | mh   |    24 |  839 | 0.75 |          |
| bpM8VcnfdmyxZ1vg | 23                        | Druid   | mh   |    17 | 1206 | 0.37 | feral    |
| hPWNkH67AYCBFLgM | 1, 3-8                    | Druid   | mh   |    14 | 1073 | 0.34 | feral    |

## Puncturing Spear (285332)

- Proc: weapon proc 1318250, logged as 1318250; rolls on main hand
- Sim rate: 2.45 PPM, observed; measured 2.46 PPM [0.92, 2.91], 3 wearers, 3 runs, 117 procs on 920 hits
- Reports from 2026-11-07 to 2026-11-15
- Notes: 3 of 6 runs are feral druids, left out (paw speed, not the weapon); few wearers (3): the interval is about the per-wearer range

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1ZJVMgcYWTyRQh2w | 1-3      | Paladin | mh   |    12 |  134 | 1.73 |          |
| aykDWnCGx4jJFTV6 | 1-6      | Hunter  | mh   |     1 |   21 | 0.92 |          |
| hxYkG4fKMrPmjRCD | 1-8      | Paladin | mh   |   104 |  765 | 2.63 |          |
| Ybjk79mPGZWrq2xy | 1-2      | Druid   | mh   |    23 | 1037 | 0.43 | feral    |
| ZFBD3y2d8cfP46Ja | 1-2      | Druid   | mh   |    38 | 1056 | 0.70 | feral    |
| zyALHJFhZtYq2Gvn | 1-3, 5-7 | Druid   | mh   |    67 | 1890 | 0.69 | feral    |

## Pyromancer's Baton (286650)

- Proc: weapon proc 1320925, logged as 1320925; rolls on main hand, off hand
- Sim rate: 2.05 PPM, observed; measured 2.03 PPM [1.85, 2.54], 7 wearers, 7 runs, 395 procs on 4861 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights       | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 4NMcqLwQ2rB9amCH | 1-3          | Rogue   | mh   |    35 |  340 | 2.57 |          |
| 9hywC8jMTLV36Pxd | 4-8          | Shaman  | mh   |    56 |  805 | 1.74 |          |
| 9rPx2JNYLbCy1waX | 1-9, 11-15   | Rogue   | oh   |    13 |  166 | 1.96 |          |
| Brj9PhXHQM3kLN2f | 1-4          | Rogue   | oh   |    50 |  481 | 2.60 |          |
| HqwWnYGKdhpCgZVR | 1-12         | Rogue   | mh   |   171 | 2335 | 1.83 |          |
| Pw9kWLGcAj3VpmCn | 13-18, 20-22 | Warrior | mh   |    31 |  398 | 1.95 |          |
| QPmw1C9MNc4hqH3K | 1-5          | Warrior | oh   |    39 |  335 | 2.91 |          |

## Quillshooter (10567)

- Proc: weapon proc 29646, logged as 29646; rolls on ranged
- Sim rate: 2.9 PPM, observed; measured 2.91 PPM [2.62, 3.17], 6 wearers, 6 runs, 465 procs on 3421 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights       | Class  | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | ------------ | ------ | ------ | ----: | ---: | ---: | -------- |
| 1QgdaxJ2GBpmvyjL | 1-2, 4-8     | Hunter | ranged |    68 |  509 | 2.86 |          |
| 4bqYVtkjL6aKdwQM | 1-2          | Hunter | ranged |    63 |  472 | 2.86 |          |
| ByRxHQmnzfhbvtcV | 8-24         | Hunter | ranged |   123 |  813 | 3.24 |          |
| CpABMyt29JkdwjZf | 1-20         | Hunter | ranged |    72 |  503 | 3.07 |          |
| Hrt2wgZkdjTRBC4A | 12-13, 15-20 | Hunter | ranged |    86 |  771 | 2.39 |          |
| MfXjTPdwar9gyKvW | 1-2          | Hunter | ranged |    53 |  353 | 3.22 |          |

## Rivenspike (13286)

- Proc: weapon proc 17315, logged as 17315; rolls on main hand
- Sim rate: 1.65 PPM, observed; measured 1.65 PPM [1.25, 2.26], 9 wearers, 9 runs, 1419 procs on 17845 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 3RB9X8HVxgCJ4My2 | 1-4         | Rogue   | mh   |    76 |  673 | 2.34 |          |
| AdfMXDLW29k6pGy1 | 1-3         | Warrior | mh   |    82 |  815 | 2.08 |          |
| AyTmpNtxB9jbnJ2H | 10-45       | Rogue   | mh   |   378 | 3225 | 2.42 |          |
| CfHvwytKJ6P4q7ag | 1-24        | Warrior | mh   |    95 | 1114 | 1.76 |          |
| CfHvwytKJ6P4q7ag | 21-24       | Rogue   | mh   |   104 | 1029 | 2.09 |          |
| CqM4Vnh831tFgcYw | 1-2         | Rogue   | mh   |    63 | 1270 | 1.03 |          |
| DWfMg7hzPF4d1mQ6 | 1-11, 13-52 | Warrior | mh   |   396 | 7548 | 1.09 |          |
| kaH1G4ztmYpgMTJ6 | 1-2         | Rogue   | mh   |    51 |  615 | 1.71 |          |
| vny4RK9TkMzwtAQV | 1-5         | Rogue   | mh   |   174 | 1556 | 2.31 |          |

## Sang'thraze the Deflector (9379)

- Proc: weapon proc 1294243, logged as 1294243; rolls on main hand
- Sim rate: 2.35 PPM, observed; measured 2.34 PPM [2.15, 2.86], 8 wearers, 10 runs, 921 procs on 13895 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights                               | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 1CYAWTtmcGdFqbPN | 1-5                                  | Warrior | mh   |    34 |  774 | 1.55 |          |
| 28AhM3VGnvFcJ4rD | 1-4                                  | Warrior | mh   |   168 | 2383 | 2.49 |          |
| 4NqLGxvrZYX2Qwnf | 1-4                                  | Warrior | mh   |    41 |  510 | 2.84 |          |
| 8KPHjdAxZYFNLtgR | 1-5                                  | Warrior | mh   |   129 | 1818 | 2.50 |          |
| 987q1pjnYdVQbCmT | 6                                    | Warrior | mh   |    34 |  384 | 3.12 |          |
| AXn7cJ9VQ6YHLWmh | 1-5, 7-9, 11-13, 15-18, 20-21, 23-30 | Paladin | mh   |    45 |  293 | 5.42 |          |
| Lg82ZFDyQ4MvRm9p | 1-2                                  | Paladin | mh   |    53 | 1120 | 1.67 |          |
| WYFZHxvL7zg6ydnQ | 1-6                                  | Paladin | mh   |   210 | 3081 | 2.41 |          |
| XHDjdLb7R9P6NQm1 | 1-2                                  | Paladin | mh   |   137 | 2494 | 1.94 |          |
| zLacx86Thr3AMtqn | 1-15                                 | Warrior | mh   |    70 | 1038 | 2.38 |          |

## Satyr's Lash (17752)

- Proc: weapon proc 1292155, logged as 1292155; rolls on main hand, off hand
- Sim rate: 2.1 PPM, observed; measured 2.10 PPM [1.91, 2.31], 9 wearers, 9 runs, 393 procs on 6606 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 4NqLGxvrZYX2Qwnf | 5-8      | Rogue   | oh   |    31 |  525 | 2.08 |          |
| 4PtKGL1zcRV8Chfn | 3        | Warrior | mh   |    47 |  728 | 2.28 |          |
| 4YzXaGtg81WdqM2N | 7-15     | Rogue   | oh   |     1 |   45 | 0.78 |          |
| Lg82ZFDyQ4MvRm9p | 14-26    | Hunter  | oh   |    18 |  443 | 1.43 |          |
| R1acrQhyNvwXLmBj | 1-3      | Rogue   | oh   |    33 |  630 | 1.85 |          |
| T71XnfaF3cY9M2Rg | 1-13     | Rogue   | oh   |    57 |  920 | 2.19 |          |
| Yf91cBLN26jdqrxC | 1-4      | Hunter  | oh   |    21 |  329 | 2.25 |          |
| qm1Bvy29hGK6bMaQ | 1-2      | Warrior | mh   |   157 | 2553 | 2.17 |          |
| tdYLMmfgcAJCTWyr | 8, 10-12 | Rogue   | oh   |    28 |  432 | 2.29 |          |

## Scented Runewood Brooch (269741)

- Proc: equip proc 1318323, logged as 1318324; rolls on harmful spells, melee abilities, melee swings, ranged abilities, ranged auto attacks
- Only against Furbolg targets (its tooltip); counted on hits against Blackmaw creatures, melee, ranged and spell hits alike
- Sim rate: 1.2 PPM, observed; measured 1.13 PPM [0.79, 2.17], 5 wearers, 5 runs, 235 procs on 4140 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights               | Class   | Slot    | Procs | Hits |  PPM | Left out |
| ---------------- | -------------------- | ------- | ------- | ----: | ---: | ---: | -------- |
| Hwq97xLh6gv8ZPrY | all in Blackmaw Hold | Paladin | trinket |    61 | 1027 | 1.11 |          |
| MR8g1PhjbDN3tpkv | all in Blackmaw Hold | Priest  | trinket |    11 |   63 | 3.14 |          |
| Q7V24RA9HTg6F38p | all in Blackmaw Hold | Hunter  | trinket |    74 |  706 | 1.90 |          |
| fNMJcqDHb9ZzPLnk | all in Blackmaw Hold | Rogue   | trinket |    75 | 1484 | 1.22 |          |
| wa7j4vxdLRZtfDcT | all in Blackmaw Hold | Mage    | trinket |    14 |  225 | 1.24 |          |

## Scorpion Sting (1265)

- Proc: weapon proc 18208, logged as 18208; rolls on main hand, off hand
- Sim rate: 0.95 PPM, observed; measured 0.95 PPM [0.79, 1.17], 9 wearers, 9 runs, 249 procs on 6572 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 64NVwXYTpFRBxJWz | 9-11   | Rogue   | mh   |     4 |  211 | 0.47 |          |
| 6NrFMaRQ9yAjxWC2 | 11-12  | Rogue   | mh   |    37 | 1240 | 0.75 |          |
| AbnJ1XMLhwqtpx6F | 1-4    | Rogue   | oh   |    10 |  402 | 0.62 |          |
| Axp6DdjvyT1Gwh4Y | 1      | Paladin | mh   |    44 | 1334 | 0.82 |          |
| BM8NmpvwVdzA76f9 | 1-9    | Rogue   | mh   |    59 | 1272 | 1.16 |          |
| Br2pycwj4zLC1ZP7 | 1-3    | Warrior | oh   |    30 |  476 | 1.57 |          |
| CKLMhw2QbTVc8Hym | 1-14   | Rogue   | mh   |    37 | 1160 | 0.80 |          |
| VYb6f27NazpmCch3 | 1-2    | Warrior | oh   |    10 |  224 | 1.12 |          |
| zbg1wFp2L8hWHZva | 1-26   | Warrior | oh   |    18 |  252 | 1.78 |          |

## Searing Blade (12992)

- Proc: weapon proc 1299850, logged as 1299850; rolls on main hand
- Sim rate: 2.4 PPM, observed; measured 2.38 PPM [2.10, 2.74], 9 wearers, 9 runs, 621 procs on 6261 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights       | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 1K6c48bWMN9nzCZA | 1-3          | Warrior | mh   |    62 |  553 | 2.69 |          |
| 3yM21KdFP7LnrBvx | 1-3          | Warrior | mh   |   103 |  991 | 2.49 |          |
| 6wxJrRWgPyQb21hm | 1-7          | Paladin | mh   |   109 | 1216 | 2.15 |          |
| 7ZxcPNTjqvfzQHgX | 1-2          | Hunter  | mh   |     6 |   54 | 2.67 |          |
| C7xfzTbPdAr1w9cX | 1-12         | Paladin | mh   |    66 |  882 | 1.80 |          |
| CGXFqgATBZdW1n7V | 7-15         | Paladin | mh   |   102 |  727 | 3.37 |          |
| CHZbFGkRPyWcrd9a | 1-3, 6, 8-10 | Paladin | mh   |    85 |  793 | 2.57 |          |
| PNYpz2hCmGMRV1tx | 1-3          | Hunter  | mh   |     0 |   10 | 0.00 |          |
| zjcGFyxpkKTnQXdw | 1-11         | Warrior | mh   |    88 | 1035 | 2.04 |          |

## Serpent Slicer (13035)

- Proc: weapon proc 17511, logged as 17511; rolls on main hand, off hand
- Sim rate: 1.05 PPM, observed; measured 1.05 PPM [0.93, 1.19], 6 wearers, 6 runs, 232 procs on 5297 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: swap noise: the swap drop moved PPM 0.86 -> 1.05 (20 active minutes, 18% of hits, in 1 runs)

| Report           | Fights    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 9zxHaWbYT3gqDQvF | 8-18      | Paladin | mh   |    52 | 1169 | 1.07 |          |
| C4TGJvzHF3aZ2NjR | 1-13      | Rogue   | mh   |   100 | 2327 | 1.03 |          |
| PmM8nxa1p6QkhVLC | 1, 3-7    | Hunter  | mh   |    51 | 1173 | 1.04 |          |
| Td2y9DjVvn1YgaHr | 1-7, 9-21 | Hunter  | oh   |    15 |  293 | 1.23 |          |
| Z7qF4j3hazBTxHcw | 8-10      | Hunter  | oh   |     7 |  162 | 1.04 |          |
| gywYLPFQCJa9XfVG | 1-6       | Hunter  | oh   |     7 |  173 | 0.97 |          |

## Shadowblade (2163)

- Proc: weapon proc 18138, logged as 18138; rolls on main hand, off hand
- Sim rate: 1.05 PPM, observed; measured 1.04 PPM [0.86, 1.22], 5 wearers, 5 runs, 224 procs on 7187 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights             | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 3GT4tz8gxnphcbkD | 1-7                | Rogue   | mh   |    53 | 1528 | 1.16 |          |
| 3qfDVTZ1J29zPChn | 1-23               | Warrior | oh   |    64 | 1602 | 1.33 |          |
| 7fz9Kk8rY6jyQAbL | 6-10               | Rogue   | oh   |    17 |  775 | 0.73 |          |
| ATwq9YWpNbLC6Khc | 1-10, 12-17, 19-20 | Rogue   | oh   |    23 |  839 | 0.91 |          |
| xRkgQCj3TD2yb6Fn | 1-25               | Rogue   | mh   |    67 | 2444 | 0.91 |          |

## Shortsword of Vengeance (754)

- Proc: weapon proc 13519, logged as 13519; rolls on main hand, off hand
- Sim rate: 1 PPM, measured; measured 1.01 PPM [0.92, 1.10], 7 wearers, 8 runs, 441 procs on 10952 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights        | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2AvFk4PJmBhyb89n | 1             | Hunter  | mh   |    24 |  488 | 1.23 |          |
| 3T8GrWJRdBkvy4gz | 1-6           | Rogue   | mh   |    37 |  921 | 1.00 |          |
| NhHDL86y4PFrCW3G | 1-2, 4-11, 13 | Warrior | mh   |    32 | 1026 | 0.78 |          |
| gDtFwpGPLM2TyZKB | 1-6, 8-9      | Warrior | mh   |    74 | 1953 | 0.95 |          |
| kv32zprHjRfnZPh9 | 4-17          | Hunter  | mh   |    16 |  503 | 0.80 |          |
| q68xzRHYhCpFLjfJ | 1-4           | Paladin | mh   |    76 | 1745 | 1.09 |          |
| tYFp32ThgGPfnvQH | 1, 3          | Rogue   | mh   |    76 | 1896 | 1.00 |          |
| zX7thnkw1pR6WFyJ | 1-21          | Rogue   | mh   |   106 | 2420 | 1.10 |          |

## Skeletal Club (2256)

- Proc: weapon proc 13440, logged as 13440; rolls on main hand
- Sim rate: 1.05 PPM, observed; measured 1.06 PPM [0.87, 1.54], 7 wearers, 7 runs, 151 procs on 3293 hits
- Reports from 2026-11-08 to 2026-11-15

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 7rAN4vgZq3xjRhf1 | 1-2    | Paladin | mh   |    36 |  807 | 1.03 |          |
| AQd8hCMn7a4ymGBW | 1, 3-6 | Shaman  | mh   |    11 |  231 | 1.10 |          |
| F4rpyvWbxgfZCBXa | 1      | Warrior | mh   |    16 |  402 | 0.92 |          |
| G7BvNCDbc12nKVpJ | 1-57   | Warrior | mh   |    25 |  169 | 3.41 |          |
| M6XxqLHC8nmvB9wc | 1      | Paladin | mh   |    13 |  286 | 1.05 |          |
| P143Tb2X9rBWwQDq | 1-18   | Warrior | mh   |    24 |  696 | 0.80 |          |
| Qb7TJqYkKtgyz1ah | 1      | Rogue   | mh   |    26 |  702 | 0.85 |          |

## Sliverblade (5756)

- Proc: weapon proc 1293653, logged as 1293653; rolls on main hand, off hand
- Sim rate: 2.25 PPM, observed; measured 2.24 PPM [1.97, 2.57], 6 wearers, 6 runs, 295 procs on 5635 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights       | Class   | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | ------------ | ------- | ---- | ----: | ---: | ---: | --------- |
| 43wtgArG9vm8pb6N | 1, 3-4, 6-16 | Rogue   | oh   |    68 | 1308 | 2.23 |           |
| 9g7NQBqDH6FGfP8L | 4            | Rogue   | mh   |    14 |  307 | 1.96 |           |
| JHrkTPq23LpAcGxY | 14-18        | Warrior | mh   |    89 | 1596 | 2.39 |           |
| PZKA7G2kQtw9H4Dx | 4-9          | Rogue   | oh   |    33 |  766 | 1.85 |           |
| PkWRr8jZp9ATL7N4 | 1-23, 25-26  | Rogue   | mh   |    51 | 1148 | 1.90 |           |
| TbB8q4RMK3LmYJNv | 1-6          | Rogue   | oh   |    40 |  511 | 3.35 |           |
| RJ1nvMk7AxTYm6tb | 14-18        | Warrior | mh   |    89 | 1596 | 2.39 | duplicate |

## Spellsever Crossbow (274293)

- Proc: weapon proc 1293714, logged as 1293714; rolls on ranged
- Sim rate: 1 PPM, observed; measured 1.01 PPM [0.94, 1.09], 30 wearers, 34 runs, 734 procs on 18098 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights                    | Class   | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------- | ------- | ------ | ----: | ---: | ---: | -------- |
| 1DVFRby49Pg2tpzG | 1-19                      | Hunter  | ranged |     6 |   80 | 1.88 |          |
| 1KF2jNqTzc6VJ9fp | 6-8                       | Hunter  | ranged |    13 |  368 | 0.88 |          |
| 1fHnGhFabmPTq6Dv | 7-17                      | Hunter  | ranged |    28 |  589 | 1.19 |          |
| 1jgTLvJQBMP6kxrm | 1-12                      | Hunter  | ranged |    36 | 1098 | 0.82 |          |
| 1qpQ6FW8fxD9gC4G | 1-7, 9-11, 13-27          | Hunter  | ranged |    68 | 1727 | 0.98 |          |
| 1w9jNgyZv7YKAacb | 1-13                      | Hunter  | ranged |    50 | 1053 | 1.19 |          |
| 214ndcFyjpPM8wah | 3, 6-9, 11-15, 19-28, 30  | Rogue   | ranged |     0 |    2 | 0.00 |          |
| 2WH4jLt3FADy8fnh | 1-2, 5-6                  | Hunter  | ranged |    38 |  938 | 1.01 |          |
| 2WH4jLt3FADy8fnh | 5-6                       | Rogue   | ranged |     0 |    0 |      |          |
| 3LhpDbvyrtNFWdxH | 1-5                       | Hunter  | ranged |    38 | 1162 | 0.82 |          |
| 3pCKd6jqgHyFc9kf | 1-17                      | Rogue   | ranged |     2 |   28 | 1.79 |          |
| 4KWV2rZDJMqNLRmT | 1-3, 6-7                  | Hunter  | ranged |     2 |   43 | 1.16 |          |
| 4NMcqLwQ2rB9amCH | 1-3                       | Hunter  | ranged |    19 |  523 | 0.91 |          |
| 6VJxcqnYa4wdBWQv | 1-9                       | Rogue   | ranged |     0 |    0 |      |          |
| 6kN78G2xfqy3vwTP | 1-6                       | Warrior | ranged |     1 |    3 | 8.33 |          |
| 8CzwkcFWrA91QaxG | 1-12                      | Warrior | ranged |     1 |   10 | 2.50 |          |
| D8JF7rbYQRcWxqn3 | 1-2, 4-8                  | Hunter  | ranged |    39 |  733 | 1.33 |          |
| Fwcmz4r9aDnpRWAd | 1-2                       | Rogue   | ranged |     0 |    0 |      |          |
| Gyw2T3LjptRW6vH9 | 1-6, 8-10                 | Hunter  | ranged |    34 |  867 | 0.98 |          |
| HARvQ2hJWp8f3grD | 4-9                       | Hunter  | ranged |    18 |  481 | 0.94 |          |
| J2XL6AyZqgpjx8QK | 8-13                      | Warrior | ranged |     1 |    9 | 2.78 |          |
| K6rZRbwYD1Wyjhpx | 1-4                       | Rogue   | ranged |     0 |    0 |      |          |
| KyYwHQjxAL8tzvdJ | 1-7                       | Hunter  | ranged |    27 |  645 | 1.05 |          |
| MpcVyhTHnqCJvDFx | 1-5                       | Hunter  | ranged |    41 | 1003 | 1.02 |          |
| PFGvpDf9WZgcY6Rk | 1-12                      | Hunter  | ranged |    44 |  868 | 1.27 |          |
| Vb3ZxfXtyDWr1ndP | 1                         | Hunter  | ranged |    31 |  912 | 0.85 |          |
| WYrCyJkdfMXw84Ph | 1-21, 23-35, 37-39, 41-48 | Hunter  | ranged |     5 |  201 | 0.62 |          |
| XHDjdLb7R9P6NQm1 | 1-2                       | Hunter  | ranged |    45 | 1046 | 1.08 |          |
| b12KD98fndF4hMq3 | 1-18                      | Hunter  | ranged |    75 | 1764 | 1.06 |          |
| bCzt1Fq24mYcPXkB | 1-2                       | Hunter  | ranged |    44 | 1362 | 0.81 |          |
| jJWt7Dnm4qc9rzQd | 1                         | Hunter  | ranged |     7 |  164 | 1.07 |          |
| rD8m4JW1bVyYPKGX | 1-3, 6                    | Rogue   | ranged |     0 |    3 | 0.00 |          |
| rQ4dFakMnWXj89vz | 1-6                       | Warrior | ranged |     0 |    3 | 0.00 |          |
| vgrm7nPFN231GYRW | 1, 3-15                   | Hunter  | ranged |    21 |  413 | 1.27 |          |

## Stalvan's Reaper (934)

- Proc: weapon proc 13524, logged as 13524; rolls on main hand, off hand
- Sim rate: 1.35 PPM, observed; measured 1.37 PPM [1.09, 1.71], 6 wearers, 7 runs, 351 procs on 5298 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 2THWnvG8AXCYqfNg | 1-2       | Warrior | oh   |    70 |  737 | 1.97 |          |
| 4ZKFNBk6czpynG3W | 1-2, 5-10 | Rogue   | mh   |    46 |  777 | 1.23 |          |
| 4dVyt9cvbPrgpY2K | 1-2       | Hunter  | mh   |    38 |  364 | 2.16 |          |
| 61VcxvhfQF2DRAGd | 1-2       | Warrior | mh   |    68 | 1513 | 0.93 |          |
| BtcMA1CqVrnapQ2m | 1-8       | Warrior | mh   |    90 | 1062 | 1.75 |          |
| C8Hk1q3rpN27YWam | 3         | Warrior | mh   |    14 |  184 | 1.57 |          |
| MjxzVLAta8FgXR7y | 1-2       | Warrior | mh   |    25 |  661 | 0.78 |          |

## Strike of the Hydra (6909)

- Proc: weapon proc 1292657, logged as 1292657; rolls on main hand
- Sim rate: 2.25 PPM, observed; measured 2.23 PPM [1.92, 2.64], 14 wearers, 14 runs, 796 procs on 6479 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 12A94BJt3KCFZrY8 | 1-12, 14-21 | Hunter  | mh   |    19 |  214 | 1.61 |          |
| 1a3pDtCcPgBkL6wm | 1-22        | Warrior | mh   |   119 |  604 | 3.58 |          |
| 1jgtdNmDMRycbZrP | 1           | Paladin | mh   |    64 |  477 | 2.44 |          |
| 1mRKL7Tb8qwjkgv3 | 1-2         | Warrior | mh   |    58 |  510 | 2.07 |          |
| 1yjwF8MxXqvtD4bT | 1-3         | Warrior | mh   |    37 |  317 | 2.12 |          |
| 23ZyGvk7BXAVDT4F | 1-14        | Warrior | mh   |    83 |  504 | 2.99 |          |
| 9zabj3grQkGB2FKL | 1-17        | Paladin | mh   |   153 | 1613 | 1.72 |          |
| BKnf4rNwJdcqpQ18 | 1-6         | Paladin | mh   |    76 |  562 | 2.46 |          |
| HARvQ2hJWp8f3grD | 1-9         | Hunter  | mh   |    12 |  123 | 1.77 |          |
| HtQr7kYaNq9MGgvZ | 1-12        | Warrior | mh   |    39 |  335 | 2.12 |          |
| LHRhtADNW8GfyYqT | 1           | Paladin | mh   |    76 |  646 | 2.14 |          |
| VQaYAf1Wxvm6ndHK | 1-11        | Hunter  | mh   |     2 |   80 | 0.45 |          |
| dQMKjpxJBCz1wrDW | 1, 3-12     | Warrior | mh   |    58 |  470 | 2.24 |          |
| nJtB3F7cQy6NmzGZ | 1-12        | Hunter  | mh   |     0 |   24 | 0.00 |          |

## Sul'thraze the Lasher (9372)

- Proc: weapon proc 1294430, logged as 1294430, 1294431; rolls on main hand
- Sim rate: 1.5 PPM, measured; measured 1.52 PPM [1.36, 1.61], 7 wearers, 7 runs, 991 procs on 15041 hits
- Reports from 2026-11-12 to 2026-11-16

| Report           | Fights                                    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------------------------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 6kJNxp39PWqrghmG | 1-2                                       | Paladin | mh   |   242 | 3525 | 1.58 |          |
| Ag69GWdLRDhKCrcT | 1-6                                       | Warrior | mh   |    98 | 1445 | 1.57 |          |
| BFPtcmbyLr71Kj63 | 1-3                                       | Paladin | mh   |   155 | 2295 | 1.56 |          |
| BMJxG63gcbQX8tWh | 1                                         | Paladin | mh   |   202 | 2792 | 1.67 |          |
| Z78jbgTL9V1Yfahq | 1, 3-4, 6-10, 13-15, 17, 20-26, 29, 31-37 | Paladin | mh   |    55 | 1372 | 0.93 |          |
| fz2HQxXk96ywVYv7 | 1-8                                       | Paladin | mh   |   163 | 2520 | 1.49 |          |
| hCPLc2G7KAVX8Bby | 1-6, 8-9                                  | Warrior | mh   |    76 | 1092 | 1.61 |          |

## Supercharger Battle Axe (9486)

- Proc: weapon proc 13527, logged as 13527; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.97 PPM [1.69, 2.28], 4 wearers, 4 runs, 156 procs on 1698 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| 2axJHNwqQdtW1jnF | 1-3    | Hunter | mh   |    10 |   71 | 3.02 |          |
| 7kxNnWGw2FzfLJtR | 1-8    | Shaman | mh   |    80 |  856 | 2.00 |          |
| fDycAFZWt4NkRX1x | 1-3    | Shaman | mh   |    51 |  560 | 1.95 |          |
| rNwJc31qXfzWthQD | 1-5    | Shaman | mh   |    15 |  211 | 1.52 |          |

## Sword of Corruption (13032)

- Proc: weapon proc 17510, logged as 17510; rolls on main hand, off hand
- Sim rate: 0.95 PPM, observed; measured 0.96 PPM [0.80, 1.21], 8 wearers, 8 runs, 195 procs on 5778 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 3npHxB84DT2zRtmN | 1-4, 6 | Warrior | mh   |    31 |  505 | 1.75 |          |
| 62WRbBjFQCpZGr3z | 1-4    | Rogue   | mh   |    19 |  631 | 0.86 |          |
| 6XBqk2G4LZy83mQz | 1, 3-7 | Warrior | mh   |    34 |  796 | 1.22 |          |
| 6dDjtwh1HXFxg72L | 1      | Warrior | oh   |    20 |  563 | 1.01 |          |
| 7CcF8GHPv4fagQYh | 1-10   | Rogue   | mh   |    25 |  757 | 0.94 |          |
| 7LY3ADmFM1wBarRn | 1-4    | Warrior | oh   |    18 |  723 | 0.71 |          |
| QZcRmKX4vMVh7f6j | 1-14   | Rogue   | mh   |    37 | 1459 | 0.72 |          |
| ZHK32LtTCmqR4Pxf | 1, 3   | Warrior | oh   |    11 |  345 | 0.91 |          |

## Sword of Decay (1727)

- Proc: weapon proc 13528, logged as 13528; rolls on main hand
- Sim rate: 1.1 PPM, observed; measured 1.08 PPM [0.90, 1.25], 6 wearers, 6 runs, 239 procs on 4932 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 9nCtfP6kvHrzV4FL | 2      | Rogue   | mh   |    51 | 1451 | 0.78 |          |
| MGzd8mFZJgPLkAc2 | 1-2    | Paladin | mh   |    38 |  634 | 1.33 |          |
| MXjF1GhRmLfabNZV | 1-18   | Warrior | mh   |    69 | 1294 | 1.18 |          |
| ZHK32LtTCmqR4Pxf | 1, 3   | Warrior | mh   |    25 |  432 | 1.29 |          |
| ZqghBpkcy7xtPfmH | 1-9    | Warrior | mh   |    35 |  693 | 1.12 |          |
| nZ3NhbvpVXg649Bm | 4-10   | Rogue   | mh   |    21 |  427 | 1.09 |          |

## Tainted Pierce (8225)

- Proc: weapon proc 13530, logged as 13530; rolls on main hand, off hand
- Sim rate: 2.15 PPM, measured; measured 2.13 PPM [1.96, 2.31], 6 wearers, 7 runs, 510 procs on 7567 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights     | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 4na1LcCpP8QHTNVD | 1, 3-8     | Rogue   | mh   |    24 |  400 | 1.89 |          |
| 4wTtzhPfDqnQWNy7 | 1-4        | Paladin | mh   |    78 | 1198 | 2.06 |          |
| 8WfLRh1qTpFxJgra | 1-2        | Warrior | mh   |    84 | 1064 | 2.49 |          |
| J97GvmrRXCFZfQ8a | 1-2        | Rogue   | mh   |    84 | 1263 | 2.10 |          |
| N3Y1mFkJ9Dwr74bA | 1-9, 11-12 | Paladin | mh   |    90 | 1459 | 1.95 |          |
| NXgv93kr2G7zyTxm | 1, 3-9     | Rogue   | oh   |    82 | 1269 | 2.04 |          |
| qcrKJGPvkfhAtdjQ | 1-4        | Rogue   | oh   |    68 |  914 | 2.35 |          |

## Taran Icebreaker (2915)

- Proc: weapon proc 16415, logged as 16415; rolls on main hand
- Sim rate: 0.9 PPM, observed; measured 0.89 PPM [0.79, 1.00], 3 wearers, 3 runs, 256 procs on 4550 hits
- Reports from 2026-11-13 to 2026-11-14
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----: | ---: | ---: | -------- |
| Bz8CAjZnNyXhgTtH | 1-14   | Shaman | mh   |    83 | 1479 | 0.89 |          |
| aAcjGPbyZXNTzR1K | 1      | Shaman | mh   |    91 | 1602 | 0.90 |          |
| mg3MBw6rcNtZpjbA | 1-14   | Shaman | mh   |    82 | 1469 | 0.88 |          |

## The Blackrock Slicer (13285)

- Proc: weapon proc 17407, logged as 17407; rolls on main hand
- Sim rate: 2.05 PPM, observed; measured 2.04 PPM [1.92, 2.35], 3 wearers, 5 runs, 1032 procs on 7598 hits
- Reports from 2026-11-11 to 2026-11-16
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights                       | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| DW9YkdvyrjtMBQnf | 1-5                          | Paladin | mh   |   173 | 1405 | 1.85 |          |
| FHqYCNcKntvDAjxg | 25-31, 33-36                 | Paladin | mh   |   109 |  779 | 2.10 |          |
| NXMZV9xjvWpB7GtP | 10, 12-107, 109-116, 118-122 | Warrior | mh   |   222 | 1418 | 2.35 |          |
| gmKvNqHFJLZXfdnP | 1, 3-6                       | Paladin | mh   |   295 | 2230 | 1.98 |          |
| vwCXzhMQ2drgpRxD | 1-41                         | Shaman  | mh   |   233 | 1766 | 1.98 |          |

## The Chief's Enforcer (9477)

- Proc: weapon proc 1294339, logged as 1294339; rolls on main hand
- Sim rate: 1 PPM, measured; measured 1.01 PPM [0.92, 1.11], 7 wearers, 7 runs, 422 procs on 7354 hits
- Reports from 2026-11-10 to 2026-11-16
- Notes: 6 of 13 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights         | Class  | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------------- | ------ | ---- | ----: | ---: | ---: | -------- |
| 1P4ZKpMGVjvntRgm | 1-3            | Shaman | mh   |     8 |  148 | 0.95 |          |
| 28AhM3VGnvFcJ4rD | 1-4            | Shaman | mh   |   116 | 1733 | 1.18 |          |
| 3LTnjm16YN2bDzCP | 1-6            | Hunter | mh   |    14 |  344 | 0.72 |          |
| 3PAwV6XFaB7hqDgz | 1-12           | Shaman | mh   |    27 |  505 | 0.94 |          |
| 7rzAygVjwH2C69Tv | 1, 3-14, 16-62 | Shaman | mh   |   103 | 1777 | 1.02 |          |
| fywmzq86PRtjDJra | 1              | Shaman | mh   |   127 | 2378 | 0.94 |          |
| mCqj9fpbthFNGgkc | 1              | Shaman | mh   |    27 |  469 | 1.02 |          |
| 274WmyH18B6ZRgAz | 1              | Druid  | mh   |     7 | 1218 | 0.10 | feral    |
| 4h1PjwkaKX9q8mgT | 1-2            | Druid  | mh   |    14 | 1347 | 0.18 | feral    |
| 6f2mtzphbVCgKYJ3 | 1-17           | Druid  | mh   |    25 | 2487 | 0.18 | feral    |
| 7Yj1a2RW9JCvDbHp | 1-7, 9-11      | Druid  | mh   |     4 |  951 | 0.07 | feral    |
| GTwzB9Q3VPZW8jYy | 1-5            | Druid  | mh   |    33 | 2767 | 0.21 | feral    |
| fTn7FmzCwkJpZt4N | 1-3            | Druid  | mh   |    17 | 1545 | 0.19 | feral    |

## The Cruel Hand of Timmy (13401)

- Proc: weapon proc 17505, logged as 17505; rolls on main hand, off hand
- Sim rate: 2.1 PPM, observed; measured 2.10 PPM [1.86, 2.59], 6 wearers, 7 runs, 583 procs on 9243 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1ZyhaVDdFtTYWMcC | 4-7, 9-12 | Rogue   | oh   |    47 |  641 | 2.44 |          |
| 2Z4G3r9pfqJtCM6W | 7-17      | Paladin | mh   |   245 | 4534 | 1.80 |          |
| 3BAz1TDGL9kjy8ZQ | 1-2       | Warrior | oh   |     2 |   31 | 2.14 |          |
| C1cyPbL3RQvGjq8p | 48        | Warrior | oh   |    33 |  300 | 3.67 |          |
| Qb1yWxa8vHCGL7Yr | 2-7, 9-15 | Rogue   | oh   |    44 |  844 | 1.74 |          |
| qcdyZAr2ganDtxNW | 13-44     | Warrior | mh   |   167 | 2270 | 2.45 |          |
| rRnVJhQFa9zf7NqK | 4-5       | Rogue   | oh   |    45 |  623 | 2.41 |          |

## The Ziggler (8006)

- Proc: weapon proc 13482, logged as 13482; rolls on main hand, off hand
- Sim rate: 3.7 PPM, observed; measured 3.69 PPM [3.25, 4.12], 9 wearers, 9 runs, 853 procs on 8154 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ----- | ---- | ----: | ---: | ---: | -------- |
| 2bYvT8kWxjFfdwrm | 1-2    | Rogue | oh   |    46 |  575 | 2.82 |          |
| 3DQy7BhjH1mnZPpq | 9-12   | Rogue | mh   |    63 |  478 | 4.65 |          |
| 4xnDCWwrGKAFdNHv | 1      | Rogue | oh   |    67 |  726 | 3.26 |          |
| 9MmVG8crjzJZ24K7 | 1-4    | Rogue | oh   |    63 |  794 | 2.80 |          |
| A2FMrK3Y1xkvafQH | 1-7    | Rogue | mh   |   141 | 1262 | 3.94 |          |
| BTRMkxmvfrwcZNaG | 1-3    | Rogue | oh   |    77 |  837 | 3.25 |          |
| BmfFgAczHXN7b16M | 9-24   | Rogue | mh   |   208 | 1869 | 3.93 |          |
| R9njq1AGDvmp36bM | 1-8    | Rogue | oh   |    80 |  880 | 3.21 |          |
| x3PftMnXCBZgLAG2 | 1, 4-9 | Rogue | oh   |   108 |  733 | 5.20 |          |

## Thornspike (6681)

- Proc: weapon proc 1293183, logged as 1293183; rolls on main hand, off hand
- Sim rate: 3.85 PPM, observed; measured 3.87 PPM [3.07, 4.59], 13 wearers, 13 runs, 807 procs on 7368 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1VKQY2Dafc7PJzpR | 1-2      | Warrior | mh   |    55 |  690 | 2.82 |          |
| 1Zv6bBfgDP3RVFYa | 1-2, 4-7 | Rogue   | oh   |    12 |  132 | 3.20 |          |
| 1w8AbW4RBNLX6Kfy | 1        | Warrior | mh   |    15 |  333 | 1.59 |          |
| 2KkaByn6H9mzb4Yq | 1-4      | Rogue   | oh   |    76 |  810 | 3.31 |          |
| 2t6mDPLkfWBxcr4q | 1-18     | Hunter  | mh   |     5 |   47 | 3.73 |          |
| 39BFWKC2J8RTzyMx | 1-3      | Rogue   | oh   |   123 |  686 | 6.33 |          |
| 72gPMatGLpwd93jW | 13       | Rogue   | mh   |    24 |  197 | 4.30 |          |
| MdaKf7zpNgWHF1Jj | 1-4      | Hunter  | oh   |     0 |   22 | 0.00 |          |
| PFGvpDf9WZgcY6Rk | 1-12     | Rogue   | mh   |   203 | 1553 | 4.61 |          |
| VYb6f27NazpmCch3 | 1-2      | Shaman  | mh   |    51 |  394 | 4.57 |          |
| b12KD98fndF4hMq3 | 10-18    | Rogue   | mh   |    94 |  745 | 4.46 |          |
| fnMH4F2pkGC8T6Kz | 1-8      | Shaman  | mh   |    79 | 1103 | 2.53 |          |
| xJbw6RTvtHPncNzp | 1-5      | Warrior | mh   |    70 |  657 | 3.76 |          |

## Toofbreaker (286430)

- Proc: weapon proc 1320269, logged as 1320269; rolls on main hand, off hand
- Sim rate: 1.95 PPM, measured; measured 1.96 PPM [1.86, 2.06], 14 wearers, 16 runs, 1364 procs on 21976 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1BNXynxVbH4kGt8R | 1-2      | Warrior | mh   |    19 |  305 | 1.97 |          |
| 2PZCz9c1HgtrX7Th | 3-5      | Warrior | mh   |    65 |  891 | 2.30 |          |
| 2taZz4vTdfXYAk3b | 1-2      | Rogue   | oh   |    39 |  600 | 2.05 |          |
| 2xY1D3mCwagdy7Wc | 1-2, 4-6 | Warrior | mh   |    66 | 1086 | 1.92 |          |
| 2yqzTJfLN8cvwaDK | 1-2, 4-6 | Warrior | mh   |    60 |  912 | 2.08 |          |
| 3JbjLGXVzctAmfaD | 1-4      | Warrior | mh   |    80 | 1137 | 2.22 |          |
| 4YVTAqgG3dtmD8C9 | 1-23     | Warrior | mh   |   172 | 2636 | 2.06 |          |
| Cp3ND2J1xWhbFzP7 | 1-26     | Warrior | mh   |   126 | 2142 | 1.86 |          |
| LPQd7pfbjDkYZyRz | 1-2, 4-5 | Rogue   | oh   |    42 |  651 | 2.04 |          |
| LPQd7pfbjDkYZyRz | 1-2, 4-5 | Paladin | mh   |   117 | 1852 | 1.99 |          |
| YakBLWFpjnrdhGb3 | 1-2      | Paladin | mh   |   182 | 3211 | 1.79 |          |
| cFRpkf1jWTaHrgL3 | 1-2      | Paladin | mh   |   156 | 2667 | 1.85 |          |
| qnyc1zwaB4FrvM9x | 1        | Paladin | mh   |    55 | 1059 | 1.64 |          |
| tac7hXVRgkbZNMDy | 3        | Warrior | mh   |    54 |  892 | 1.91 |          |
| vtyhfFz9wPGJdHqM | 1-3      | Warrior | mh   |    50 |  666 | 2.37 |          |
| ynVw7ZgMx4tjF3fQ | 10-18    | Paladin | mh   |    81 | 1269 | 2.02 |          |

## Twilight Forgehammer (281665)

- Proc: weapon proc 1312176, logged as 1312176; rolls on main hand, off hand
- Sim rate: 3.15 PPM, observed; measured 3.16 PPM [2.56, 3.79], 6 wearers, 6 runs, 1344 procs on 10206 hits
- Reports from 2026-11-09 to 2026-11-15

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| 6WgpmHQrv9kn1ZXJ | 3-22   | Paladin | mh   |   460 | 2846 | 3.88 |          |
| 6b4F7HrzNTnxRMXd | 1-13   | Warrior | oh   |   177 |  841 | 5.05 |          |
| JRr2jLy8vNx147fX | 1-27   | Paladin | mh   |   128 | 1131 | 2.72 |          |
| Ztv8xkT6dmQ7n91C | 1      | Warrior | mh   |   295 | 2645 | 2.68 |          |
| b4hn6GgpFLtJ1N2X | 1-13   | Warrior | mh   |   132 | 1419 | 2.23 |          |
| m8RLWaY7bh9jHKwD | 2      | Rogue   | mh   |   152 | 1323 | 2.76 |          |

## Twilight Maul (273841)

- Proc: weapon proc 1292679, logged as 1292679; rolls on main hand
- Sim rate: 1.75 PPM, observed; measured 1.74 PPM [1.58, 1.96], 7 wearers, 7 runs, 373 procs on 3905 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 5 of 12 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1WKqM9x42jcLnYCN | 1-36      | Shaman  | mh   |   105 | 1102 | 1.73 |          |
| 1hpgVjWH3GZBLYKw | 13, 15-25 | Shaman  | mh   |    25 |  207 | 2.20 |          |
| 1nLR3XCxjcAdpKtB | 1-18      | Shaman  | mh   |    57 |  688 | 1.51 |          |
| 1q2tTDcaFMnfzvkB | 1         | Shaman  | mh   |    17 |  172 | 1.80 |          |
| 2NQC6YJtXRPbMzg1 | 1-13      | Shaman  | mh   |    72 |  803 | 1.63 |          |
| BzGXfjMLJ2RyCwHm | 1-7       | Paladin | mh   |    62 |  675 | 1.67 |          |
| RqpmbCcYr8BJZWwK | 1-2       | Shaman  | mh   |    35 |  258 | 2.47 |          |
| 2NBRf9AwdZTjDQYa | 7-13      | Druid   | mh   |     4 |  168 | 0.43 | feral    |
| 6WzAhcVgfLCt78nX | 1-8       | Druid   | mh   |    34 | 1139 | 0.54 | feral    |
| GAVFWPJBtmn2x6ry | 1-6       | Druid   | mh   |    20 |  960 | 0.38 | feral    |
| VYb6f27NazpmCch3 | 1-2       | Druid   | mh   |    13 |  558 | 0.42 | feral    |
| b69d1Dhw4JkxBHTt | 1-14      | Druid   | mh   |    40 | 1426 | 0.51 | feral    |

## Venomstrike (6469)

- Proc: weapon proc 29653, logged as 29653; rolls on ranged
- Sim rate: 2.05 PPM, observed; measured 2.05 PPM [1.88, 2.23], 19 wearers, 23 runs, 667 procs on 8127 hits
- Reports from 2026-11-06 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.53 -> 2.05 (113 active minutes, 26% of hits, in 4 runs)

| Report           | Fights  | Class   | Slot   | Procs | Hits |  PPM | Left out |
| ---------------- | ------- | ------- | ------ | ----: | ---: | ---: | -------- |
| 12VhLkPa98DYFH4g | 1-6     | Warrior | ranged |     1 |    6 | 4.17 |          |
| 1Rk8y7VKNWapF2dt | 3-4     | Rogue   | ranged |     0 |    4 | 0.00 |          |
| 1jKJfWybdV2vBQ8A | 1-2     | Hunter  | ranged |    22 |  172 | 3.20 |          |
| 1qNaRQCwLmDPcJfx | 1-2     | Hunter  | ranged |    25 |  358 | 1.75 |          |
| 23dT9fk1DqcaJmwn | 1-7     | Hunter  | ranged |    58 |  743 | 1.95 |          |
| 261hbMqpgzLRKxwY | 1-2     | Warrior | ranged |     3 |   21 | 3.57 |          |
| 2axJHNwqQdtW1jnF | 1-3     | Hunter  | ranged |    68 |  737 | 2.31 |          |
| 3CyQk69LzJjTn2X4 | 1-5     | Warrior | ranged |     0 |    0 |      |          |
| 3npHxB84DT2zRtmN | 1-4, 6  | Warrior | ranged |     0 |    0 |      |          |
| 4dvhFmKbqkTQ1Wpc | 1-3     | Rogue   | ranged |     0 |    1 | 0.00 |          |
| 7W3YTHD9kyPzFvtc | 1-4     | Warrior | ranged |     0 |    0 |      |          |
| 7ZxcPNTjqvfzQHgX | 1-2     | Hunter  | ranged |    13 |  124 | 2.62 |          |
| 7ypMG8JVjRgZ1xHQ | 1-9     | Hunter  | ranged |    44 |  668 | 1.65 |          |
| GPhAqKkF8tHBpfxT | 1-3     | Hunter  | ranged |    37 |  542 | 1.71 |          |
| GmZaJgjW2R97cVnL | 6       | Hunter  | ranged |    37 |  351 | 2.64 |          |
| RQxLGd1agBDt6hXP | 1-4     | Hunter  | ranged |    65 |  737 | 2.20 |          |
| WhwcvTCZFPDxanyr | 22-65   | Hunter  | ranged |    78 |  795 | 2.45 |          |
| YCqznXvDMQA4d718 | 1-7     | Hunter  | ranged |    96 | 1102 | 2.18 |          |
| bzWMAC2mf7Z6Q3LV | 1-9     | Hunter  | ranged |    66 |  971 | 1.70 |          |
| d6bwAkLKjGq7pVMy | 1       | Hunter  | ranged |    15 |  211 | 1.78 |          |
| jM3dtQ7vcGZ9fhWn | 1-5     | Hunter  | ranged |    35 |  509 | 1.72 |          |
| kBzVYb3Cy7QwnKMD | 1-6     | Hunter  | ranged |     4 |   71 | 1.41 |          |
| rXvTBhVgpZmc1kbP | 1, 3-13 | Warrior | ranged |     0 |    4 | 0.00 |          |

## Verigan's Fist (6953)

- Proc: weapon proc 1322218, logged as 1322218; rolls on main hand
- Sim rate: 2.2 PPM, measured; measured 2.21 PPM [2.04, 2.34], 10 wearers, 10 runs, 963 procs on 8186 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 1CYTfP3pFDcaRrJm | 1-8         | Paladin | mh   |   138 | 1213 | 2.13 |          |
| 1HTPQGZya7pAz3qh | 3-32, 34-36 | Paladin | mh   |    78 |  682 | 2.14 |          |
| 1pgjDMXka9ZFxy8f | 1-3         | Paladin | mh   |    40 |  302 | 2.48 |          |
| 1rYynxW6AGfcQz9t | 5-6, 8-17   | Paladin | mh   |   211 | 1679 | 2.36 |          |
| 26cnTVbFtvQfzANd | 1           | Paladin | mh   |   137 | 1041 | 2.47 |          |
| 2AxRVTac9WCmG6f3 | 1-6         | Paladin | mh   |    44 |  354 | 2.33 |          |
| 7V8ykCfL1htHwzKD | 1-3         | Paladin | mh   |    54 |  528 | 1.92 |          |
| QmjBxrzvah2XTVR1 | 3-8         | Paladin | mh   |   146 | 1174 | 2.33 |          |
| VQaYAf1Wxvm6ndHK | 6-11        | Paladin | mh   |    59 |  584 | 1.89 |          |
| hLHaBZ29kyA4MCpj | 1-12        | Paladin | mh   |    56 |  629 | 1.67 |          |

## Vibroblade (9485)

- Proc: weapon proc 1293005, logged as 1293005; rolls on main hand, off hand
- Sim rate: 1 PPM, observed; measured 1.00 PPM [0.69, 1.20], 6 wearers, 6 runs, 111 procs on 4161 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | -------- |
| BtN6VMJRjgp1Ay9G | 1-3    | Paladin | mh   |     4 |  287 | 0.52 |          |
| MKx2v8JfHVc6kmND | 1-6    | Rogue   | oh   |     9 |  669 | 0.50 |          |
| Pzk84bCLKYcvN9fx | 1-10   | Rogue   | oh   |    12 |  482 | 0.93 |          |
| QBCAa6JwknGvcMrP | 1-2    | Rogue   | oh   |    22 |  617 | 1.34 |          |
| TacY2wkjAgXynNHJ | 1-17   | Paladin | mh   |    48 | 1452 | 1.24 |          |
| fA1GLbc6gw9nKHM2 | 1-16   | Rogue   | oh   |    16 |  654 | 0.92 |          |

## Waveslicer (18324)

- Proc: weapon proc 1302166, logged as 1302166; rolls on main hand
- Sim rate: 1.9 PPM, observed; measured 1.92 PPM [1.83, 2.01], 4 wearers, 5 runs, 1500 procs on 15142 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Procs | Hits |  PPM | Left out  |
| ---------------- | ------ | ------- | ---- | ----: | ---: | ---: | --------- |
| AyTmpNtxB9jbnJ2H | 4-44   | Shaman  | mh   |   318 | 3245 | 1.90 |           |
| DpVmH4WRxL72nGwd | 3      | Shaman  | mh   |   127 | 1275 | 1.93 |           |
| NrpM7DJQBKWqbzVF | 1-7    | Paladin | mh   |   164 | 1784 | 1.78 |           |
| VYBjgJx8aZtRhmpy | 1-43   | Shaman  | mh   |   691 | 6951 | 1.92 |           |
| q23Tz8hnZGkvbLx6 | 1-18   | Shaman  | mh   |   200 | 1887 | 2.05 |           |
| jnWqct213HRGAM76 | 1      | Shaman  | mh   |   107 | 1106 | 1.87 | duplicate |

## Witchfury (13051)

- Proc: weapon proc 18214, logged as 18214; rolls on main hand
- Sim rate: 1 PPM, observed; measured 1.02 PPM [0.87, 1.16], 6 wearers, 6 runs, 215 procs on 4525 hits
- Reports from 2026-11-15 to 2026-11-16

| Report           | Fights    | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 6vFAMDQZ2RmaxC7n | 1-3       | Paladin | mh   |    37 |  714 | 1.11 |          |
| HXg2W4bTFDjcRJ87 | 1-3       | Paladin | mh   |    34 |  958 | 0.76 |          |
| RKvbj4hcCTJWNk3Q | 4-5       | Paladin | mh   |    48 | 1013 | 1.02 |          |
| cLkwmZbzFDKdPhYn | 11, 13-14 | Paladin | mh   |    18 |  443 | 0.87 |          |
| f43cC7V8BvrMXyQb | 1-4       | Paladin | mh   |    51 |  915 | 1.19 |          |
| fBrKy1M2T6q9LC3g | 1-7       | Warrior | mh   |    27 |  482 | 1.20 |          |

## Wolfsbane (267369)

- Proc: weapon proc 1282503, logged as 1282503; rolls on main hand
- Sim rate: 2.25 PPM, measured; measured 2.26 PPM [2.15, 2.36], 29 wearers, 29 runs, 1857 procs on 14529 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights   | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 18zavRLKVhycxjWZ | 1-2      | Paladin | mh   |    41 |  399 | 1.81 |          |
| 1FNGzPgAaX6wDpBt | 1-7      | Paladin | mh   |    54 |  488 | 1.95 |          |
| 1NBKq2mDWghwR69z | 1-4, 6   | Paladin | mh   |    37 |  380 | 1.72 |          |
| 1mAfJLYwZXHMzVtd | 1-2      | Paladin | mh   |    66 |  462 | 2.52 |          |
| 1mWTYf2wPgCGqkMp | 1-2      | Paladin | mh   |    52 |  480 | 1.91 |          |
| 1qjDRYwtHxyJPfQn | 1-5      | Paladin | mh   |   101 |  733 | 2.43 |          |
| 6GY7rBk9gHJfVMFm | 1-2      | Paladin | mh   |   111 |  911 | 2.15 |          |
| 6GY7rBk9gHJfVMFm | 1-2      | Paladin | mh   |   118 |  813 | 2.56 |          |
| 8DQBRhWyXrjHzktc | 1, 3-4   | Paladin | mh   |    81 |  610 | 2.34 |          |
| Br9cPVnJgCZFwpmK | 1-2      | Paladin | mh   |    75 |  484 | 2.73 |          |
| GPhAqKkF8tHBpfxT | 1-3      | Paladin | mh   |    71 |  497 | 2.52 |          |
| HtQr7kYaNq9MGgvZ | 1-12     | Paladin | mh   |    64 |  548 | 2.06 |          |
| RvcrA3P9jznxhHVQ | 4-7      | Paladin | mh   |    65 |  511 | 2.24 |          |
| TAgkJxp6dN9M2BtQ | 1        | Paladin | mh   |    40 |  359 | 1.97 |          |
| VYb6f27NazpmCch3 | 1-2      | Paladin | mh   |    62 |  421 | 2.60 |          |
| Vhk7TNKgMzdH2C8A | 1-3      | Paladin | mh   |    73 |  557 | 2.31 |          |
| YCqznXvDMQA4d718 | 1-7      | Paladin | mh   |    53 |  444 | 2.11 |          |
| b69d1Dhw4JkxBHTt | 1-14     | Paladin | mh   |   133 |  987 | 2.38 |          |
| b6C24Jd3V8WwpFx7 | 2-4      | Paladin | mh   |    14 |   76 | 3.25 |          |
| bnRJKL9FTMq2zkQG | 1        | Paladin | mh   |     4 |   33 | 2.14 |          |
| bxtk19P4qVQjmYFC | 1-2      | Paladin | mh   |    90 |  828 | 1.92 |          |
| bzWMAC2mf7Z6Q3LV | 1-9      | Paladin | mh   |    34 |  395 | 1.52 |          |
| dP8M1ZqnTWR4Brf6 | 1        | Paladin | mh   |    67 |  476 | 2.48 |          |
| df8Mh4tqkJxHnCa3 | 1-5      | Paladin | mh   |   103 |  769 | 2.36 |          |
| jLTw4rQd1BKgMkhy | 1-2, 4-6 | Paladin | mh   |    66 |  482 | 2.42 |          |
| nDrBjQ74hH2gpvTw | 2        | Paladin | mh   |    10 |  102 | 1.73 |          |
| tZQTyMXr3BY7jxHg | 1-3      | Paladin | mh   |    80 |  551 | 2.56 |          |
| yLjkTYJft9DV2FGM | 1-2      | Paladin | mh   |    43 |  352 | 2.16 |          |
| yLjkTYJft9DV2FGM | 1-2      | Paladin | mh   |    49 |  381 | 2.27 |          |

## X'caliboar (10758)

- Proc: weapon proc 1293787, logged as 1293787; rolls on main hand
- Sim rate: 2.1 PPM, observed; measured 2.11 PPM [1.94, 2.36], 6 wearers, 6 runs, 492 procs on 4231 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.66 -> 2.11 (34 active minutes, 21% of hits, in 1 runs)

| Report           | Fights         | Class   | Slot | Procs | Hits |  PPM | Left out |
| ---------------- | -------------- | ------- | ---- | ----: | ---: | ---: | -------- |
| 176QTFWR8PtdfGZB | 25-32          | Paladin | mh   |    52 |  425 | 2.22 |          |
| 1WHqx7bPa8mkGydJ | 1, 3           | Warrior | mh   |   144 | 1313 | 1.99 |          |
| 6QvVFqNYty8bzrcf | 1, 3-11, 13-21 | Paladin | mh   |   117 |  828 | 2.57 |          |
| K8rhAtLjcz7gdHkn | 6-7            | Warrior | mh   |    23 |  233 | 1.79 |          |
| P1yKzX3Q8NgFYxZH | 1-3            | Warrior | mh   |   127 | 1158 | 1.99 |          |
| V3L7kJDzTGg4QqCF | 1-3            | Paladin | mh   |    29 |  274 | 1.92 |          |

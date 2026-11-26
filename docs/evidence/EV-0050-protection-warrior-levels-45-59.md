# EV-0050: Protection warrior levels 45 to 59, threat, Enrage, Rock on Stick, rage and damage shields

- Claim: Shield Slam Ranks 2 and 3, Cleave Ranks 3 and 4 and Heroic Strike Rank 8 carry their own flat threat, Enrage
  rolls 25% on non-physical spell hits taken, Rock on Stick procs 1.85 times a minute, damage taken rage holds near 38%
  through attacker level 59, and Razorsteel Shoulders' shield and Thorns reflect almost nothing on a block that takes
  the whole hit.
- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at levels 45 to 59; Cleave Rank 3's count
  gives its own
- Game build: 1.60.1.70862 to 1.60.1.70938; Cleave Rank 3's count gives its own
- Date: 2026-11-09 to 2026-11-13 per the pull request; the override rows read 2026-11-14; Cleave Rank 3's count gives
  its own
- Pull request: #1008; Cleave Rank 3's count is #1157, Heroic Strike Rank 8's reading #1236

## Flat threat

- Count: §Count of the 2026-11-17 to 2026-11-20 logs

The ThreatLog reads each Shield Slam rank as its damage plus its own flat, added once on a crit, and Cleave Rank 4 as a
flat split over the targets it lands on:

| Spell                        | Measured                                                    |
| ---------------------------- | ----------------------------------------------------------- |
| Shield Slam Rank 1 (23922)   | 178.1 [177.8, 178.4], n 134                                 |
| Shield Slam Rank 2 (23923)   | 202.85 [202.76, 202.94], n 379                              |
| Shield Slam Rank 3 (23924)   | 229.25 [229.10, 229.39], n 36                               |
| Cleave Rank 4 (11609)        | 80.2 [79.5, 80.8], n 53; one target 80.3, two 79.8          |
| Heroic Strike Rank 8 (11567) | 145: whole-point flat 313 on 9 of 15 windows, median 144.72 |

178, 203 and 229 is not 4.45 a level and 62 and 80 is not 1.5 a level, so the ranks above them extend the measured
line, Shield Slam Rank 4 to 255 at level 60 and Cleave Rank 5 at Rank 4's 1.6 a level, 96, until a level 60 log
measures them.

Heroic Strike Rank 8 makes its damage plus 145 threat, times the stance, where Rank 4's 2.5 a spell level would
give 140. Its windows come from levels 58 (1.60.1.70888, 16 windows) and 59 (1.60.1.70938, one), in Defensive Stance
with Defiance. A window is the threat the warrior's threat log gained over one landed Heroic Strike Rank 8 hit,
matched by its label, on a mob it was engaged with. Every other term in the window is a rule already measured, each
event rounded to the nearest whole point: white swings, item procs and Rend damage times the stance, Bloodrage ticks,
Revenge, Thunder Clap, Shield Slam, Heroic Strike Rank 5 and Sunder Armor (EV-0211). Windows with a term no rule
covers (Taunt, a stance swap, Execute, Whirlwind and the like) and the first window of a pull are left out. What is
left is the strike's own whole-point threat y; F = y / 2.16 − damage is the flat it implies.

17 windows are kept; 15 lie within 10 of the median of F, 144.72 (mean 143.34 [141.86, 144.82]). On those 15:

| Flat | round(2.16 x (damage + flat)), exact | Within 1 | round(2.16 x damage) + round(2.16 x flat), exact | Within 1 |
| ---- | ------------------------------------ | -------- | ------------------------------------------------ | -------- |
| 139  | 0                                    | 2        | 2                                                | 2        |
| 140  | 0                                    | 0        | 0                                                | 0        |
| 142  | 1                                    | 1        | 1                                                | 1        |
| 145  | 6                                    | 11       | 9                                                | 11       |
| 146  | 0                                    | 0        | 0                                                | 2        |

Every other flat from 136 to 150 matches no window exactly. The whole-point flat, y − round(2.16 x damage), reads 313
on 9 windows, 314 on 2, 307 on 1, 300 on 2 and 295 on 1; round(2.16 x 145) is 313, and 140 would read 302, which no
window does. The 16 windows on a single mob, 14 of them within 10 of the median, read the same (median 144.71; 145
exact on 5 by the first form and 8 by the second).

## Enrage

Ranks 1 to 5 proc on 3,439 of 11,773 swings taken (29.2%) and 236 of 812 physical spell hits (29.1%), but 268 of
1,087 non-physical spell hits (24.7% [22.2, 27.3]); never on full blocks, avoided hits or ticks. Resists, absorbs,
Enrage already being up and the matching window do not explain it (unresisted hits alone 24.9%; 25.0 to 25.2% at 100 to
250 ms windows). The game files state 30% for all of them.

## Rock on Stick

Rock to the Face procs 399 times on 4,980 landed main-hand hits without extra attacks or haste, 1.85 procs a
minute [1.68, 2.03] at 2.6 speed. White swings, specials and Sunder Armor casts roll it; Shield Slam (3 of 1,664),
Shield Bash and Thunder Clap do not.

## Damage-taken rage

Here k is 2.5 x (1 - A), as in EV-0048, read from rage using attackers' physical swings on landed damage:

| Attacker level | n   | k                    | Sim before |
| -------------- | --- | -------------------- | ---------- |
| 55             | 281 | 1.550 [1.545, 1.555] | 1.547      |
| 56             | 233 | 1.545 [1.539, 1.551] | 1.542      |
| 57             | 186 | 1.551 [1.547, 1.556] | 1.536      |
| 58             | 180 | 1.559 [1.550, 1.573] | 1.530      |
| 59             | 38  | 1.544 [1.536, 1.551] | 1.524      |

Attackers at level 60 to 62 are covered by no log, and interpolate from 38% (60: 38.5% in place of 39.3%).

## Damage shields

Razorsteel Shoulders' shield answered 15,042 of 15,069 hits and 3,835 of 3,840 partial blocks, and 34 of
5,281 full blocks. Thorns answered 1 of 8,334 full blocks. Mithril Shield Spike and Scarlet Retribution, which fire on
every block in game, go through the block proc path.

## Count of the 2026-11-17 to 2026-11-20 logs

- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at levels 42 to 47
- Game build: 1.60.1.70938 on 2026-11-17, 1.60.1.70959 from the 2026-11-18 09:07 login; hotfixes 114140 to 114286
  each taken at a login once pushed; no reading moves between them
- Date: 2026-11-17 to 2026-11-20
- Pull request: #1157

The method is EV-0043's, run per log as EV-0174 describes, with Cleave read as one flat per cast split over the targets
it lands on.

| Spell                 | Measured                                                                                                                                                                                                     |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Cleave Rank 3 (11608) | 62 a cast: one target landed, median 62.00 (n 9, party n 21), within 1 of 62 on 6 of 9 and 15 of 21 against 1 of 9 and 1 of 21 for 60; two landed, 61.59 a cast (n 137, party 191); flat from level 42 to 47 |

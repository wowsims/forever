# EV-0121: Skullforge Reaver's rate from the order of the log lines

- Claim: Skullforge Reaver 13361 casts Skullforge Brand (17484) 1.9 times a minute on main-hand hits, white swings and
  melee casts alike, before the brand's own hit roll; the rate is measured.
- Sources: public reports on forever.warcraftlogs.com, zone 3000, with the fights each wearer had it on. Rogue:
  123jYCvkVGrHfXWt fights 1 to 30, 1z4vRDqGdM2cVPXk 1 to 8, 2wMznDRvmkpbtHxA 1 to 7, 9zxHaWbYT3gqDQvF 17 to 18,
  CfHvwytKJ6P4q7ag 1 to 24, ZfbY6RgWjx9hmpGz 1 to 14, vZqVmLYQa6jwP3A9 5 to 21, vwYF4Z2C38cfKTDB 1 to 9. Warrior:
  NXMZV9xjvWpB7GtP 28 to 122. Hunter: NXMZV9xjvWpB7GtP 107 to 122.
- Game build: 1.60.1.70938; the hotfix state and relogs were not recorded. No hotfix on this build changed 17484, the
  item's on-hit entry or its 2.5 s speed.
- Date: 2026-11-14 to 2026-11-15
- Pull request: #1040
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

A wearer is a player whose combatant info lists the item, and a run that two players logged counts once. A proc is a
brand that landed (its applydebuff or refreshdebuff line) or one that failed (a damage line of 17484 that is not a tick,
resisted, immune or missed: 95 of them). Either line follows the damage line of the hit that rolled it in the same
millisecond: of 2,121 attributed procs, 1,961 share the hit's millisecond and 132 come within 10 ms after it. Each proc
therefore goes to the last landed hit logged before it within 50 ms, or failing that the first hit after it within
30 ms; 32 procs matched no hit. A melee cast that lands with no damage line, such as Cheap Shot, counts at its debuff
line.

Every wearer held the sword in the main hand with a weapon in the off hand. White swings are split by hand for each
run on their damage: a two-group split on the logarithm of the damage before armor, which the combat log gives before
the critical strike bonus. A rate is the chance times 60 over the sword's 2.5 s speed. A proc on a target that
already carries the brand logs refreshdebuff, so it refreshes the brand. The brand's drain and heal ticks crit (about 11%
of each).

| Hit                   | Procs / hits | Chance | PPM [95%]         | Wearer bootstrap [95%] |
| --------------------- | ------------ | ------ | ----------------- | ---------------------- |
| Main-hand white swing | 955 / 11,599 | 8.23%  | 1.98 [1.86, 2.10] | [1.74, 2.06]           |
| Main-hand melee cast  | 906 / 11,490 | 7.89%  | 1.89 [1.78, 2.01] | [1.82, 2.05]           |
| Windfury Totem hit    | 248 / 3,731  | 6.65%  | 1.60 [1.41, 1.80] |                        |
| Off-hand white swing  | 12 / 12,870  | 0.09%  |                   |                        |
| Ranged-type hit       | 0 / 579      | 0%     |                   |                        |

The melee casts are Sinister Strike, Eviscerate, Whirlwind, Bloodthirst, Cleave and Cheap Shot's debuff. White swings
and melee casts roll at one rate. Windfury Totem hits roll it too, at a lower rate (1.60, interval 1.41 to 1.80); 248
procs do not settle whether the Windfury hit rolls at a rate of its own. Off-hand swings give 12 procs in 12,870 and
ranged-type hits none in 579, against 8.23% for main-hand swings, so neither rolls it. §Rate gives the rate. The counts
come from 6 wearers (4 rogues, 1 warrior and 1 hunter) over 9 runs.

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 17484
- Rate: 1.9 PPM, measured

2,050 procs on 25,868 rolling hits, 11 wearers, 13 runs, in the main hand: all hits 1.90 [1.76, 1.99]. The rate is all
hits, which the specials rate agrees with. One-hand white: 1.98 [1.13, 3.21], 16 procs on 194 lone units, 1 wearer (a
Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.68 [1.39, 1.80], 280 procs on 3,993 lone
units, 10 wearers. Specials: 2.07 [1.78, 2.44], 318 procs on 3,682 lone units, 11 wearers; Sinister Strike gives 250
procs on 2,800 of them. Two-roll abilities, priced as one roll: 2.04, 8 procs on 94 lone units (Execute 7 on 69 and
Mongoose Bite 1 on 25), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 330
isolated hits. Dropping the swapped stretches (8.5 active minutes, 1% of hits, in 1 run) moved the read from 1.88 to
1.90. Report vwYF4Z2C38cfKTDB logs the same run as CfHvwytKJ6P4q7ag and counts once. Reports from 2026-11-13 to
2026-11-21.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |       6 | 1,677 | 20,447 | 1.97 |
| Warrior |       4 |   323 |  4,783 | 1.62 |
| Hunter  |       1 |    50 |    638 | 1.88 |

Reports and fights: 123jYCvkVGrHfXWt 1-20, 22-23, 27-30; 1z4vRDqGdM2cVPXk 1-4, 7-8; 2wMznDRvmkpbtHxA 1-3, 5-7;
8jDqr3JcYRCH1pbK 1-4; 9JqKQBY68twcR2hT 1; 9zxHaWbYT3gqDQvF 17-18; BHvgJVTdZ9pfNPa6 1-3; CfHvwytKJ6P4q7ag 1-24;
GRnMHmK69BXfWVTj 1-4; NXMZV9xjvWpB7GtP 28-107, 109-116, 118-122; Zk1Gh4WA7KJDaMf2 1-2; vZqVmLYQa6jwP3A9 5-16, 18-21;
vwYF4Z2C38cfKTDB 1-9; z8QPNWpMBx4VgdDX 1.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the sword; 3,486 reports in all for the count, 14 of them for the sword
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 705 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 1 of them for the sword. The
  sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on the sword.

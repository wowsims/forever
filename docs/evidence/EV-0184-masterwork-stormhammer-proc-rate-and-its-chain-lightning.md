# EV-0184: Masterwork Stormhammer's proc rate and its Chain Lightning

- Claim: Masterwork Stormhammer (item 12794) casts Chain Lightning (16921) at 0.9 procs per minute, priced on the weapon
  base speed (1.6) of each hit that rolls it, observed on four wearers, one shaman and three paladins, with paladins
  counted by EV-0155's and EV-0156's rolls. The chain hits up to 3 targets, rolls each jump on its own from the row's
  105 to 145 at 70% a jump, crits for 1.5 times and has no cooldown.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights listed below
- Game build: 1.60.1.70959 and 1.60.1.71037 by event time. Every region's version file named 70959 from 2026-11-18 21:17
  UTC and 71037 from 2026-11-20 18:23 UTC. The shaman's two runs and the first paladin run end by 2026-11-20 14:01 UTC,
  on 70959; the two runs in China start 27 and 95 minutes after the 71037 listing, on either build; the last run, on
  2026-11-22, is on 71037. The hotfix each player ran is not recorded. The first and last raidbots hotfix cache files of
  70959 (2026-11-17 15:47 and 2026-11-20 14:54 UTC) and the newest of 71037 (2026-11-22 18:07 UTC, 6,242 pushes) hold no
  record of a row of the item or of spell 16921, so no push moved these rows
- Date: 2026-11-19 to 2026-11-22
- Pull request: #1182

## Method

Every player whose gear shows the mace: 6 runs of 4 characters, 1 shaman (2 runs) and 3 paladins (4 runs), 2 runs in
Korea, 2 in the EU and 2 in China. All hold it in the main hand with a shield, and every gear snapshot inside the 6 runs
shows it. A fifth player deals the proc's damage in two more reports (8 damage lines) with no gear snapshot of the mace
and is left out. No zone 3000 report read shows the mace or the spell before 2026-11-19.

The proc logs damage lines only, up to 3 a proc, and a jump is read off the amount before the target's resistance and
absorbs: the first jump rolls 105 to 145, the second 73.5 to 101.5 and the third 51.45 to 71.05, and the three ranges do
not overlap. A line of the first range is one proc. Of the 249 procs, 11 have no first jump among their damage lines: 5
start with a line on an immune target, which logs 186 to 340 ms before the jumps that land, and 6 show later jumps
alone. Lines inside one proc are not in jump order, and a third jump lands up to 900 ms after the first.

A roll is a landed hit of the wearer on the melee table, a white swing or a spell whose `DefenseType` is 2, with ticks
left out and each ability counted as EV-0155 counts it. A hit with a killing blow is no roll, as in EV-0153; 108 are
left out. The spell is a missile of speed 20, and its first jump lands 22 to 330 ms after a lone white swing (median
159 ms, 39 swings), so no proc is given to a hit. Each run's procs after its first are set against its rolls after the
first proc up to the last: 243 procs on 9,911 rolls. The longest stretch with no proc is 287 rolls.

A lone white swing is a white swing with no other roll of the wearer within 500 ms either side, and it has a proc when
a first jump lands 20 to 500 ms after it. Every wearer swings at the mace's 1.6 s, so these logs cannot tell procs per
minute from a flat chance per hit.

## Rates

| Count                          | Runs | Procs |  Rolls | Chance per roll | PPM [95%]         |
| ------------------------------ | ---: | ----: | -----: | --------------: | ----------------- |
| All rolls, both classes (Rate) |    6 |   243 |  9,911 |           2.45% | 0.92 [0.81, 1.04] |
| All rolls, the shaman          |    2 |    63 |  2,934 |           2.15% | 0.81 [0.63, 1.03] |
| All rolls, paladins            |    4 |   180 |  6,977 |           2.58% | 0.97 [0.84, 1.12] |
| Whole runs, both classes       |    6 |   249 | 10,374 |           2.40% | 0.90 [0.80, 1.02] |
| Lone white swings, both        |    6 |    39 |  1,479 |           2.64% | 0.99 [0.73, 1.34] |
| Lone white swings, the shaman  |    2 |    23 |  1,040 |           2.21% | 0.83 [0.55, 1.24] |
| Lone white swings, paladins    |    4 |    16 |    439 |           3.64% | 1.37 [0.85, 2.19] |

The intervals are Wilson intervals on the chance. The whole-run row counts every proc and every roll of a run. The rate
is the rate over all rolls of both classes, rounded to 0.05. Paladins are in the rate (owner ruling, 2026-11-23): they
count by EV-0155's and EV-0156's rolls wherever their rate does not read apart from the other classes'. Every run here
is after the hotfix of the Judgement rows on 2026-11-13, and no run casts Judgement of Command. The two classes' rates over all rolls have overlapping intervals, and across the 6 runs the chance per roll
gives a chi-square of 4.7 on 5 degrees of freedom. The lone white swings do not check the rate: the paladins' read 3.64%
on 16 procs where the shaman's read 2.21%, and both classes together have 39 procs, under the 50 EV-0153 asks of a
lone-unit rate. The rate is observed: 4 wearers give it, where a measured rate takes 5, and its interval reaches further
than 10% from the read. At 0.5 procs per minute the 9,911 rolls give 132 procs.

| Report           | Fight | Class   | Region | Build          | Procs | Rolls |  PPM | Longest stretch with no proc |
| ---------------- | ----: | ------- | ------ | -------------- | ----: | ----: | ---: | ---------------------------: |
| Z6kDfqLt9n38yY2a |     1 | Shaman  | KR     | 70959          |    36 | 1,615 | 0.84 |                          157 |
| aRfyqH7A3t9bZXdm |     4 | Shaman  | KR     | 70959          |    27 | 1,319 | 0.77 |                          150 |
| mGzTNLp3CdYafHM8 |     2 | Paladin | EU     | 70959          |    59 | 1,942 | 1.14 |                          206 |
| 9H72ZkdBGfhCy6P3 |    65 | Paladin | CN     | 70959 or 71037 |    50 | 1,934 | 0.97 |                          206 |
| 9H72ZkdBGfhCy6P3 |    69 | Paladin | CN     | 70959 or 71037 |    43 | 1,809 | 0.89 |                          119 |
| WdbKG4RkmjxA7YV3 |     1 | Paladin | EU     | 71037          |    28 | 1,292 | 0.81 |                          287 |

## Which hits roll

The rolls by ability:

| Class   | Ability               | Rolls |
| ------- | --------------------- | ----: |
| Shaman  | white swing           | 1,649 |
| Shaman  | Windfury Weapon       |   645 |
| Shaman  | Windfury Totem        |   324 |
| Shaman  | Stormstrike           |   316 |
| Paladin | white swing           | 3,503 |
| Paladin | Seal of Fury          | 2,744 |
| Paladin | Holy Strike           |   622 |
| Paladin | Seal of Righteousness |   108 |

A white swing by the rolls within 500 ms of it, with no second white swing there, and the swings with a proc, for each
group of 50 swings or more:

| Class   | Rolls beside the white swing         | Rolls | Swings | With a proc | Chance | Per roll |
| ------- | ------------------------------------ | ----: | -----: | ----------: | -----: | -------: |
| Shaman  | none                                 |     1 |  1,040 |          23 |  2.21% |    2.21% |
| Shaman  | 2 Windfury Weapon                    |     3 |    173 |          15 |  8.67% |    2.89% |
| Shaman  | Windfury Totem                       |     2 |    171 |           9 |  5.26% |    2.63% |
| Shaman  | Stormstrike                          |     2 |    122 |           7 |  5.74% |    2.87% |
| Shaman  | Windfury Totem and 2 Windfury Weapon |     4 |     59 |           1 |  1.69% |    0.42% |
| Paladin | none                                 |     1 |    439 |          16 |  3.64% |    3.64% |
| Paladin | Seal of Fury                         |     2 |  1,786 |          79 |  4.42% |    2.21% |
| Paladin | Holy Strike, Seal of Fury            |     3 |    259 |          15 |  5.79% |    1.93% |
| Paladin | Seal of Righteousness                |     2 |     77 |           3 |  3.90% |    1.95% |
| Paladin | Holy Strike                          |     2 |     60 |           4 |  6.67% |    3.33% |

The shaman's white swings alone would read 63 procs on 1,649 swings, 3.82% a swing, where the lone white swings read
2.21%. Stormstrike is one roll in EV-0155, and a white swing beside one has a proc on 7 of 122 here; with its 316 rolls
left out the shaman reads 0.90 [0.71, 1.15] and both classes 0.95 [0.84, 1.07]. Two procs start within 35 ms of each
other four times: off a white swing and its two Windfury Weapon hits; off a white swing and its Seal of Righteousness
hit; off a white swing, its Seal of Fury hit and a Holy Strike; and off two white swings 38 ms apart with their Seal of
Fury hits.

A Judgement of Fury (20414), which EV-0156 has rolling nothing, has no proc after it on the 163 landings with no roll
within 500 ms either side.

## The chain

Of the 530 damage lines 423 hit, 92 crit, 14 meet an immune target and 1 misses; none is logged as a resist, in part or
in full. The 515 lines that deal damage are 238 first jumps, 163 second jumps and 114 third jumps, and no line reads
under 51, where a fourth jump would fall. The sim rolls each jump on the spell hit table. Whether the proc rolls it is
not settled here: the logs give neither the wearer's level nor the mob's.

| Jump | Lines | Lowest | Highest | The row's range |
| ---- | ----: | -----: | ------: | --------------- |
| 1    |   238 |    105 |     153 | 105 to 145      |
| 2    |   163 |     73 |     102 | 73.5 to 101.5   |
| 3    |   114 |     51 |      72 | 51.45 to 71.05  |

Three first jumps read above 145: 146, and 148 and 153 on one mob. Each jump rolls on its own: of the 127 procs with no
other proc within 2 s and one first and one second jump on record, 12 have the second within 1 of 0.7 times the first,
and the second reads 0.53 to 0.94 times the first. A crit deals 1.5 times the amount: 88 of the 92 crits read 1.48 to
1.55 times the amount before the target.

The four pairs above start 0, 15, 31 and 32 ms apart, so the proc has no cooldown.

## Reports and fights

9H72ZkdBGfhCy6P3 65, 69; aRfyqH7A3t9bZXdm 4; mGzTNLp3CdYafHM8 2; WdbKG4RkmjxA7YV3 1; Z6kDfqLt9n38yY2a 1. The shaman's
runs are aRfyqH7A3t9bZXdm and Z6kDfqLt9n38yY2a. The player left out is in 2jL7TrZMxAR3khbt 198 and vc3myqk7DJdBjn14 32.

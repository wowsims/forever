# EV-0021: Dragon's Call, its whelp, the Highborne Research Tablet and weapon procs on shield hits, level 57

- Claim: Dragon's Call procs 4.5 times a minute behind a 45 s cooldown and summons a level 57 whelp that takes nothing from its owner; the Highborne Research Tablet procs off damaging hits only; and Fiery Blaze, Rock on Stick and Dragon's Call do not proc off Shield Slam, Shield Bash or Thunder Clap, while equip procs still do. Highborne Research Tablet is the exception: Thunder Clap never procs it (EV-0131).
- Sources: combat log, one level 57 Protection warrior with a 2.5 s main hand (Mount Hyjal, about 27 minutes in combat, StatLedger for gear and level); Warcraft Logs FHqYCNcKntvDAjxg, WRa9p2xhPnQKVG8C and t6bL9xfVY3hXyvHT for Fiery Blaze (3 players)
- Game build: not recorded
- Date: 2026-11-12 for the combat log; not recorded for the Warcraft Logs reports
- Pull request: #958
- Later count: the Emerald Dragon Whelp's level, swing and Acid Spit, §Count of the 2026-11-21 to 2026-11-26 reports

## Highborne Research Tablet (260205)

| Reading      | Measured                                |
| ------------ | --------------------------------------- |
| Procs        | 70, from 647 landed damaging melee hits |
| Shortest gap | 0.12 s, so no internal cooldown         |

The rate is the same on white hits, Shield Slam and Revenge. Rend and Concussion Blow applications, Rend ticks and damage
shields never proc it, so only a hit that deals damage rolls. An Auto Shot from a hunter procced it too. The chance on a
spell hit is unmeasured here; EV-0131 measures it, with the ranged and off-hand rates.

## Dragon's Call (10847)

| Reading             | Measured                                                                                                       |
| ------------------- | -------------------------------------------------------------------------------------------------------------- |
| Procs               | 41, from 220 landed melee hits outside the cooldown (Shield Slam left out)                                     |
| Per hit             | 18.6% (13.7 to 24.4)                                                                                           |
| Per minute at 2.5 s | 4.5 (3.3 to 5.9)                                                                                               |
| Cooldown            | 45 s: shortest gaps 44.94, 45.03 and 45.16 s, and none of 73 hits in the last 5 s of a cooldown procced        |
| Shield Slam         | 1 proc in 50 landed outside the cooldown, about 9.5 expected; that one shares its millisecond with a white hit |

Revenge, Heroic Strike, Cleave and Victory Rush proc it at the same rate as white hits, and so do Rend and Concussion Blow
applications. Misses, dodges and parries do not. In the sim, 400 runs of 300 s gave 18.98% per landed hit outside the
cooldown and a shortest gap of 45.000 s.

## Weapon procs and the shield hits

Fiery Blaze is counted on the three Warcraft Logs reports and Dragon's Call on the level 57 log; the pull request does
not say which logs hold the other rows.

| Proc                                | Kind           | Shield Slam                        | Shield Bash       | Thunder Clap | White hits  |
| ----------------------------------- | -------------- | ---------------------------------- | ----------------- | ------------ | ----------- |
| Fiery Blaze (enchant 36), 3 players | combat enchant | 0 of 617                           | 0 of 78           | 0 of 436     | 212 of 1517 |
| Rock on Stick (279852)              | chance on hit  | 0 of 157                           |                   |              | 41 of 528   |
| Dragon's Call (10847)               | chance on hit  | 1 of 50                            | 0 of 4            |              |             |
| Dreadfrost Saber (274603)           | equip aura     | 60 of 105, and 54 of 80 in EV-0012 | 7 of 8 in EV-0012 |              | 208 of 311  |
| Highborne Research Tablet (260205)  | equip aura     | about 11%                          |                   |              | about 10%   |

Dragon's Call's one Shield Slam proc shares its millisecond with a white hit. Non-damaging Sunder Armor, Concussion
Blow and Rend applications proc Fiery Blaze and Dragon's Call; Rend ticks proc neither.

## The Emerald Dragon Whelp (8776) and Acid Spit (9591)

- Count: §Count of the 2026-11-21 to 2026-11-26 reports

The log has 40 summons that read the same health, armor and attack power while the owner's attack power ran from 981 to
1383, and its health, armor and Acid Spit fit the creature values for level 57. The whelp never crits: one crit in 96
landed swings and none in 82 spits.

| Reading                            | Log (level 57 owner)                                                                                                                                                              | Sim                                                                                                                                              |
| ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Health                             | 5082, which is 1.3 times the creature health at 57                                                                                                                                | 1.3 times the level 57 creature health                                                                                                           |
| Armor                              | 1735, which is 0.51 of the creature armor at 57 (ExpectedStatMod 1's share in the game files)                                                                                     | 0.51 of the level 57 creature armor                                                                                                              |
| Attack power                       | 174, which the swing leaves out                                                                                                                                                   | none                                                                                                                                             |
| Glancing                           | 87 of 96 landed swings, 79% against targets at the whelp's own level, at about 57% of a hit; the swings were 13% of the whelp's damage                                            | 2% of landed swings for each defense point over 250, glancing blows by target level as §Count of the 2026-11-21 to 2026-11-26 reports reads them |
| Acid Spit mana                     | 5340 at level 57, more spits than a summon lasts                                                                                                                                  | no cost                                                                                                                                          |
| First spit                         | 0.00, 0.01, 0.89 and 2.59 s after the summon at the 0, 25, 50 and 75% marks over 41 summons                                                                                       | uniform 0 to 1.8 s                                                                                                                               |
| Wait from a cast end to next start | 0.20, 0.58, 0.99 and 1.88 s at the same marks over 69 waits, maximum 6.9 s, all with the target alive; mean 1.86 s over the 16 full summons that lasted 15 s with no target dying | uniform 0.2 to 3.5 s, mean 1.85 s                                                                                                                |
| Spits per summon                   | 2.31 (clean full summons)                                                                                                                                                         | 2.86                                                                                                                                             |
| Swings per summon                  | 2.94                                                                                                                                                                              | 2.04                                                                                                                                             |
| Damage per summon                  | 665 (clean full summons)                                                                                                                                                          | 873 when the sim built the whelp at level 60; not rerun at level 57                                                                              |

No swing lands inside a cast that reaches its end, so a spit's hardcast holds the swing and the whelp swings only as it
arrives and in the waits between casts.

The Dragon's Call rate is an override row on the trigger row 10847, and the Acid Spit row 9591 is added by hand, since a
creature's spells are on the server and no row links it.

## Count of the 2026-11-21 to 2026-11-26 reports

- Sources: Warcraft Logs, the most recent 80 reports of the index that show the whelp, 67 of them with every summoning
  owner at one level (CombatantInfo talent points plus 9), 3,464 summons, every fight: 1ACVrZzjdKRYJHht
  1GzyVFrPntN4wqT3 1apmwADn8fRjG3LC 2F86MRmLH9yXVYp7 2aYzn9ZPWgCb8yMw 2da7JFR8Ny9CY3Dz 2mXywBCLNJHRW1xc
  3dn7kP9VWJc4q18Q 6tng9xw4kTdJDvQP 6xdrXcamTAgJy3NW 7RxbGfy9ZrcdD3tk 7TcDbwAjrN3Rh9Pv 8hVw1CZ7jPfKWmBn
  9vMC3QXRyDzAJWqL 9x7GcFmz3KTN4rjY BNWmkL6H4T97Q1xF Bv1wAbztjpXh93Kf BwLYX8WGnczTrg6N DdfHPcYQzLFVrhW9
  FBzKJ8QpdtP6yY4V FZtQmAJ1NvMwgyBb HDwChqvd2bWKA3QP HjYzK8qwZxbRBpd6 JPjRwmcXGfZvnK9r NzH1DCfF8YXJALvW
  PcQN4MpmtyDHzK63 QJdHMt2h7Bx6pk3F Rcn2dBw6T7vZFbtx TCWgZG1HRthpncPY THQkMvRt3FnZ8wL7 TYL8Hyt6nFaxPcX9
  V4rhg9H1YcXCFtj2 VFWXfMThRQb1gBCr VdrX1nhPC7NpmbFj WkgQHdtn4x3ABFVZ X9grzJjpW4dhyCNK b8VvgCRTcf4xFN6q
  bJPj9mHRLkWAfNTD cPBMAvnKrjg2ta1p dKBnzJ3wmP24ktTQ dZL2m8NJhyHCTFVD dbDAWjkcHMFtVYvy dqnamBZYFTMx3XPy
  dx7pZB298K6DYgnh fM8JgtDNW3xdYayK gDFPnNdGWavmjtb2 gqjBfFyzVLhR4W78 jfNB6a8wchntJC7v jgvLQ34zd6mRkWaN
  kwqAhKdX3GJZ7pBj m2cVkDw9XYAqnZ6g nzcBtpgDa3dLrTmV p2D4K8RMktFVgT9Z pCagFqL6P72XKZrz pqrmA6VNzPaRHdCT
  qty2brxjkgP4LwRZ rhAc6qLKVn7JQ8GD tx6a7T4XqvJVjRMb v6bGnfAjX8DB9War vLa1NzdYWTJbjc7V vWd9mxRDpAcHZtwj
  xGJcKTnzPmAr1aCw xrz2mpQYjDFL1yw8 ymXZ2HDQV9JqAGzk zRTpHGjQXB6kKwb1 zarkGhTX3CtYHmxW zy4ZAcxwaNPpkvgV
- Game build: 1.60.1.71037 to 1.60.1.71139 by the report dates; hotfixes not recorded
- Date: 2026-11-21 to 2026-11-26
- Pull request: #1250

Normal hits before armor (`unmitigatedAmount`) on targets of level 60 and below, with no debuff that lowers the whelp's
damage on it, as EV-0217 counts them. The whelp's melee and Acid Spit read the same at every owner level, so the whelp
is level 57 whatever its owner's level.

| Owner level | Melee hits | Melee, 2nd percentile to highest, median | Acid Spit hits | Acid Spit mean |
| ----------- | ---------- | ---------------------------------------- | -------------- | -------------- |
| 55          | 452        | 71 to 94, 83                             | 529            | 249.4          |
| 56          | 405        | 71 to 94, 83                             | 647            | 250.2          |
| 57          | 213        | 71 to 94, 82                             | 354            | 249.7          |
| 58          | 201        | 71 to 94, 82                             | 233            | 250.2          |
| 59          | 210        | 71 to 94, 82.5                           | 608            | 248.6          |
| 60          | 543        | 71 to 94, 83                             | 3,275          | 249.4          |

| Reading   | Measured                                                                                                              | Sim                                                   |
| --------- | --------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| Level     | 57 at every owner level from 55 to 60                                                                                 | 57                                                    |
| Swing     | 2.0 s, 71 to 94 before armor; 0.73 of 2 s of the level 57 creature dps, 14% either side and rounded down, is 71 to 94 | 2.0 s at 0.73 of the level 57 creature dps, ±14%      |
| Acid Spit | 249.7 mean at owner level 57 and 248.6 to 250.2 at the others; the row's 438.5% of the level 57 creature dps is 249.7 | the row's percent of the level 57 creature dps, 249.7 |

The whelp's glancing follows the target's level alone; against targets 57 to 59 it reads 92.5%, 90.6% and 91.4% for
paladin, rogue and warrior owners, so the owner's weapon skill plays no part. A landed swing glances 2% for each point
of defense, 5 a level, the target has over 250, about a level 50 creature's skill. A glancing blow's damage is its
logged amount over `unmitigatedAmount` times the target's armor factor at attacker level 57, armor / (armor + 5245);
normal hits read 1.00 on the same factor at every target level from 48 to 60. The sim draws a glancing blow's share
uniformly from the range fitted to each level's 5th to 95th percentile, and level 63, with 20 glances, takes level 62's.

| Target | Landed swings | Glancing | 2% a point over 250 | Glances | Share of a hit, median | 5th to 95th    | Sim range    |
| ------ | ------------- | -------- | ------------------- | ------- | ---------------------- | -------------- | ------------ |
| 50     | 373           | 0%       | 0%                  | 0       |                        |                |              |
| 51     | 343           | 9.3%     | 10%                 | 32      | 0.797                  | 0.650 to 0.974 | 0.60 to 0.98 |
| 52     | 309           | 19.7%    | 20%                 | 61      | 0.803                  | 0.611 to 0.937 | 0.60 to 0.98 |
| 53     | 310           | 32.9%    | 30%                 | 102     | 0.729                  | 0.590 to 0.964 | 0.60 to 0.98 |
| 54     | 529           | 45.6%    | 40%                 | 241     | 0.789                  | 0.609 to 0.970 | 0.60 to 0.98 |
| 55     | 800           | 51.2%    | 50%                 | 410     | 0.788                  | 0.604 to 0.985 | 0.60 to 0.98 |
| 56     | 543           | 67.2%    | 60%                 | 365     | 0.802                  | 0.614 to 0.973 | 0.60 to 0.98 |
| 57     | 592           | 78.2%    | 70%                 | 463     | 0.746                  | 0.574 to 0.883 | 0.57 to 0.92 |
| 58     | 986           | 90.2%    | 80%                 | 889     | 0.546                  | 0.363 to 0.729 | 0.34 to 0.75 |
| 59     | 1,103         | 98.6%    | 90%                 | 1,088   | 0.338                  | 0.117 to 0.573 | 0.08 to 0.59 |
| 60     | 874           | 99.4%    | 100%                | 869     | 0.234                  | 0.038 to 0.426 | 0.02 to 0.45 |
| 61     | 302           | 99.7%    | 100%                | 301     | 0.147                  | 0.016 to 0.278 | 0.01 to 0.29 |
| 62     | 186           | 100%     | 100%                | 186     | 0.080                  | 0.000 to 0.180 | 0.01 to 0.18 |
| 63     | 20            | 100%     | 100%                | 20      | 0.080                  | 0.016 to 0.195 | 0.01 to 0.18 |

The rule reads 5 to 8 points under the logs against levels 56 to 59, the levels where the share of a hit starts to
fall.

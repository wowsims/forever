# EV-0051: The armor the damage-taken rage formula assumes, by attacker level

- Claim: A, the armor reduction in rage from damage taken (2.5 × (1 - A) × damage before armor / c), follows the attacker's level: 21.8% up to attacker level 38, rising to 37.7% at 47 and staying near 38% to 59.
- Sources: combat logs of a protection warrior at levels 41 to 59, a bear at levels 26 to 33 and a feral druid at levels 47 to 55 (the rows of #945 and #960), taking melee hits from rage-using mobs, with Defensive Stance divided out
- Game build: 1.60.1.70862 to 1.60.1.70938
- Date: 2026-11-08 to 2026-11-13
- Pull request: #932, with #945 (levels 47 to 49), #960 (levels 50 to 55) and #1008 (levels 55 to 59)

The measured k is 2.5 × (1 - A), from the damage each hit implies before armor (the logged amount divided by the armor factor), which removes per-mob damage multipliers. k follows the attacker's level, not the defender's: level 43 mobs read 1.816, 1.811 and 1.811 against the warrior at 43, 44 and 45.

Levels up to 45, #932:

| Attacker level | Measured k                                                                                    | Table that stood before | Table after                        |
| -------------- | --------------------------------------------------------------------------------------------- | ----------------------- | ---------------------------------- |
| 38 and below   | 1.956 [1.948, 1.965] (warrior, pooled); 1.956 [1.941, 1.970] (bear, attackers about 20 to 35) | 2.000                   | 1.955 (21.8%)                      |
| 39, 40, 41     | 1.937 [1.920, 1.954], 1.920 [1.908, 1.931], 1.922 [1.913, 1.930]                              | 2.000                   | 1.942, 1.929, 1.916 (interpolated) |
| 42             | 1.903 [1.893, 1.913]                                                                          | 2.000                   | 1.9025 (23.9%)                     |
| 43             | 1.812 [1.801, 1.823]                                                                          | 1.875                   | 1.814 (interpolated)               |
| 44             | 1.724 [1.715, 1.732]                                                                          | 1.750                   | 1.725 (31%)                        |
| 45             | 1.662 [1.646, 1.678]                                                                          | 1.675                   | 1.6625 (interpolated)              |

Warrior: protection, levels 41 to 45, 1,946 clean rage pairs from rage-type mobs, unblocked hits, Defensive Stance. Bear: levels 26 to 33, 479 pairs. The rows at 46 (36%) and 50 (38%) rest on a few dozen hits (level 50 had 3 new hits in this pass).

Levels 47 to 55:

| Attacker level | Measured A, n                                                   | Table that stood before          |
| -------------- | --------------------------------------------------------------- | -------------------------------- |
| 47             | .377 (359 pairs)                                                | interpolated .365                |
| 48             | .381 (325)                                                      | .370                             |
| 49             | .379 (293)                                                      | .375                             |
| 50 to 55       | .382, .381, .378, .379, .369, .381 (n 244, 225, 33, 15, 17, 28) | the ramp to 63 read .383 to .388 |

Levels 55 to 59, k from rage-using attackers' physical swings on landed damage:

| Attacker | n   | k                    | Sim before |
| -------- | --- | -------------------- | ---------- |
| 55       | 281 | 1.550 [1.545, 1.555] | 1.547      |
| 56       | 233 | 1.545 [1.539, 1.551] | 1.542      |
| 57       | 186 | 1.551 [1.547, 1.556] | 1.536      |
| 58       | 180 | 1.559 [1.550, 1.573] | 1.530      |
| 59       | 38  | 1.544 [1.536, 1.551] | 1.524      |

In summary, the readings are 21.8% at attackers 17 to 38, 23.9% at 42, 31% at 44, 36% at 46 and 37.7% at 47 on hundreds of hits each, and from 47 to 59 near 38%, on hundreds of hits a level to 51 and at 55 to 58, about 90 at 52 to 55 and 38 at 59. The 40% at 63 is the top of the 20% to 40% range the game build's notes give; no log covers 60 to 63, and attackers at 60 to 62 interpolate from 38% (60: 38.5% in place of 39.3%).

Attackers that use mana or energy give about half this rage in game (one mob each: warrior 0.519 times, bear 0.553 times, energy 0.501 times; see also EV-0052). The sim gives every attacker the full amount, since a target has no power type.

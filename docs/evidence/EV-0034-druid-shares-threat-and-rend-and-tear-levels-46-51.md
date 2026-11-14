# EV-0034: Rip, Rake, Pounce Bleed, Swipe and Ferocious Bite shares, bear threat and Rend and Tear, druid levels 46 to 51

- Claim: the Rank 4 shares, Dire Bear Form's Enrage share, Rend and Tear on any bleed, Faerie Fire's and Lacerate's threat, Primal Bite's threat and the armor damage-taken rage assumes at attacker levels 47 to 49 follow the readings below.
- Sources: combat logs and ThreatLog of a feral druid at levels 46 to 51; the Pounce Bleed and Ferocious Bite rows come from levels 49 to 51, and the pull request body gives levels 37 to 51 for the whole set
- Game build: 1.60.1.70862 or 1.60.1.70888
- Date: 2026-11-10
- Pull request: #945

| Reading                                                                                  | Measured                                                                                                                                                                                     |
| ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Rip Rank 4 (9752), 0.0114 per combo point up to 4                                        | 371 of 372 ticks; the share on all 5 combo points fits none of 263                                                                                                                           |
| Rake Rank 3 (1824), 0.065 on the hit and the tick                                        | 347 of 349 hits, 683 of 686 ticks                                                                                                                                                            |
| Pounce Bleed Rank 2 (9824), 0.025                                                        | 0.02482 to 0.02539 fits all 34 ticks                                                                                                                                                         |
| Swipe Rank 4 (9754), 0.03                                                                | 7,518 of 7,711 hits exact                                                                                                                                                                    |
| Ferocious Bite Rank 4 (22828), 0.037                                                     | 0.03671 to 0.03745 over 62 bites, 59 fit                                                                                                                                                     |
| Enrage in Dire Bear Form takes its 16% from the same share as Bear Form                  | item and bonus armor after the form's multiplier dropped by 16% on 153 of 153 sheets, the Agility armor on 145                                                                               |
| Rend and Tear takes any bleed on the target, a weapon proc's and other players' included | Swipe at ×1.15 on 861 of 874 hits with only other players' bleeds up, 0 of 874 at ×1.00; 20 of 6,107 with no bleed; white swings never take it                                               |
| Faerie Fire's threat, 2 × the caster's level times the form's multiplier                 | Bear Form 40 of 48 windows within 1 threat (ratio 1.0025 ± 0.0032), Dire Bear Form 54 of 71 (1.0041 ± 0.0052); Cat Form 0 on 43 of 45; measured at caster levels 38 to 50 on the lower ranks |
| Lacerate's threat, 10/3 of its damage, rank 2 in Bear Form                               | hit 3.3350 ± 0.0075 (33 of 40 windows), ticks 3.3296 ± 0.0025 (59 of 65); ×1 and ×3 fit none                                                                                                 |
| Lacerate's later ranks in Dire Bear Form                                                 | fit 10/3 on 37 of 80 hit windows and 46 of 74 tick windows, none at ×1 or ×3                                                                                                                 |
| Primal Bite's threat, 3 times its damage with no flat part                               | measured on the first two ranks                                                                                                                                                              |

Rank 4 in Dire Bear Form, with ranks 3 and 5 fitting, is the later reading of EV-0037.

Damage-taken rage, the armor A the formula assumes (EV-0048) by attacker level, from melee hits of rage-using mobs:

| Attacker level | Measured A, n    | Table that stood before |
| -------------- | ---------------- | ----------------------- |
| 47             | .377 (359 pairs) | interpolated .365       |
| 48             | .381 (325)       | .370                    |
| 49             | .379 (293)       | .375                    |

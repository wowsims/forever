# EV-0048: Protection warrior threat and attack power shares, levels 43 to 45, and the armor damage-taken rage assumes

- Claim: Revenge's threat is 2.25 times its damage, Revenge Rank 4 takes 0.25 of attack power, Sunder Armor Rank 3
  makes 117 + 0.05 x attack power threat, Heroic Strike Rank 5 adds about 80.5, Thunder Clap makes 1 x its damage as
  threat, Rend and Thunder Clap take their attack power shares on the ranks below, and Concussion Blow makes no threat;
  the armor the damage-taken rage formula assumes follows the attacker's level, 21.8% up to attacker level 38 and
  rising from 42.
- Sources: combat logs of a protection warrior at levels 41 to 45, 1,946 clean rage pairs from rage-type mobs for the
  rage table; for the rage table also 479 pairs from the logs of a bear at levels 26 to 33
- Game build: 1.60.1.70862
- Date: 2026-11-08 to 2026-11-09
- Pull request: #932

## Warrior rulings, levels 43 to 45

| Ruling                                                    | Measured                                                      |
| --------------------------------------------------------- | ------------------------------------------------------------- |
| Thunder Clap threat: 1 x damage, no flat bonus            | own threat over damage 0.9965 to 0.9994 at 1 to 4 targets     |
| Thunder Clap Rank 4 attack power share 0.07               | floor(0.99 x (55 + 0.07 x AP)), slope 0.0696 [0.0687, 0.0704] |
| Rend Ranks 4 and 5: 0.02 x attack power read at each tick | 259 of 264 at the tick's attack power                         |
| Revenge threat 2.25 on Ranks 3 and 4                      | 2.2500 (n 112), 2.2498 to 2.2509 (n 64)                       |
| Revenge Rank 4 attack power share 0.25                    | 0.2518 [0.2455, 0.2581], 601 hits                             |
| Sunder Armor Rank 3: 117 + 0.05 x sheet attack power      | 125 windows                                                   |
| Heroic Strike Rank 5 flat threat (2.5 a spell level)      | +80.5 [79.6, 81.5], n 6                                       |
| Concussion Blow makes no threat                           | -0.05 over 6 windows                                          |

Which ranks the pull request states as measured: Rend Ranks 4 and 5, with Ranks 1, 3, 6 and 7 keeping their TODO;
Thunder Clap Rank 4, with Ranks 2, 3, 5 and 6 keeping their TODO; Revenge Ranks 3 and 4 for the threat, with Ranks 5 and
6 keeping their TODO. Rend Rank 2 and Thunder Clap Rank 1 come from the levels 13 and 14 logs (EV-0019).

## Rage from damage taken

Rage from damage taken is 2.5 x (1 - A) x damage before armor / c, where A is the armor the game assumes in place of the
defender's own. Measured k is 2.5 x (1 - A), from the damage each hit implies before armor (the logged amount divided by
the armor factor), which removes per-mob damage multipliers. k follows the attacker's level, not the defender's: level
43 mobs read 1.816, 1.811 and 1.811 against the warrior at 43, 44 and 45.

| Attacker level | Measured k                                                                                    | Table that stood before | Table after                        |
| -------------- | --------------------------------------------------------------------------------------------- | ----------------------- | ---------------------------------- |
| 38 and below   | 1.956 [1.948, 1.965] (warrior, pooled); 1.956 [1.941, 1.970] (bear, attackers about 20 to 35) | 2.000                   | 1.955 (21.8%)                      |
| 39, 40, 41     | 1.937 [1.920, 1.954], 1.920 [1.908, 1.931], 1.922 [1.913, 1.930]                              | 2.000                   | 1.942, 1.929, 1.916 (interpolated) |
| 42             | 1.903 [1.893, 1.913]                                                                          | 2.000                   | 1.9025 (23.9%)                     |
| 43             | 1.812 [1.801, 1.823]                                                                          | 1.875                   | 1.814 (interpolated)               |
| 44             | 1.724 [1.715, 1.732]                                                                          | 1.750                   | 1.725 (31%)                        |
| 45             | 1.662 [1.646, 1.678]                                                                          | 1.675                   | 1.6625 (interpolated)              |

The warrior's pairs are unblocked hits in Defensive Stance. The rows at 46 (36%), 50 (38%) and 63 (40%) stay; they rest
on a few dozen hits, and level 50 had 3 new hits in this pass. Mana and energy using attackers give about half this rage
(warrior 0.519 x, bear 0.553 x, energy 0.501 x on one mob).

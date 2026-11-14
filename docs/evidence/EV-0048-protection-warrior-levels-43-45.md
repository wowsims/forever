# EV-0048: Protection warrior threat and attack power shares, levels 43 to 45

- Claim: Revenge's threat is 2.25 times its damage, Revenge Rank 4 takes 0.25 of attack power, Sunder Armor Rank 3
  makes 117 + 0.05 x attack power threat, Heroic Strike Rank 5 adds about 80.5, Thunder Clap makes 1 x its damage as
  threat, Rend and Thunder Clap take their attack power shares on the ranks below, and Concussion Blow makes no threat.
- Sources: combat log of a protection warrior at levels 43 to 45, 1,946 clean rage pairs from rage type mobs for the
  rage table
- Game build: 1.60.1.70862
- Date: 2026-11-09
- Pull request: #932

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
6 keeping their TODO. The comments at the Rend and Thunder Clap registrations read "measured in game on Ranks 2, 4 and
5" and "on Ranks 1 and 4": Rank 2 and Rank 1 come from the levels 13 and 14 logs (EV-0019). The override rows of Rend
and Thunder Clap cite only that earlier pass, so the Rend Ranks 4 and 5 and Thunder Clap Rank 4 readings above have no
override row of their own; the rows were not changed by the pull request.

Revenge's threat comment names Ranks 2, 3 and 4: the pull request states Ranks 3 and 4 only. Rank 2 is not in its body.

Rage from damage taken, the table the same logs refit (a warrior and a bear), where k = 2.5 x (1 - A) and A is the armor
the game assumes in place of the defender's own, by attacker level:

| Attacker level | Measured k                                                                                                              |
| -------------- | ----------------------------------------------------------------------------------------------------------------------- |
| 38 and below   | 1.956 [1.948, 1.965] (warrior, 2026-11-08 and 2026-11-09 pooled); 1.956 [1.941, 1.970] (bear, attackers about 20 to 35) |
| 39, 40, 41     | 1.937 [1.920, 1.954], 1.920 [1.908, 1.931], 1.922 [1.913, 1.930]                                                        |
| 42             | 1.903 [1.893, 1.913]                                                                                                    |
| 43             | 1.812 [1.801, 1.823]                                                                                                    |
| 44             | 1.724 [1.715, 1.732]                                                                                                    |
| 45             | 1.662 [1.646, 1.678]                                                                                                    |

Measured k uses the damage each hit implies before armor (the logged amount divided by the armor factor). The warrior's
pairs are from unblocked hits in Defensive Stance; the bear's are 479 pairs of a druid at levels 26 to 33. k follows the
attacker's level, not the defender's: level 43 mobs read 1.816, 1.811 and 1.811 against the warrior at 43, 44 and 45.
Mana and energy using attackers give about half this rage (warrior 0.519 x, bear 0.553 x, energy 0.501 x on one mob).

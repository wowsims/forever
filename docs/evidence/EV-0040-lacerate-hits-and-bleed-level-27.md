# EV-0040: Lacerate's hits grow with the stacks after the cast and its bleed takes no attack power

- Claim: Lacerate's hit grows 1:2:3:4:5 with the stacks after the cast, so a first Lacerate hits for its percent, and its bleed gains no attack power.
- Sources: an in-game test of a druid in Bear Form, Lacerate Rank 1, against a level 27 target, in two runs at log attack power 439 and 482; 21 casts, 15 landed
- Game build: not recorded
- Date: not recorded
- Pull request: #885

Hit before armor, by the stacks after the cast:

| Stacks after the cast | 1   | 2   | 3   | 4   | 5                      |
| --------------------- | --- | --- | --- | --- | ---------------------- |
| Attack power 439      | 8   | 16  | 24  | 28  | 41                     |
| Attack power 482      | 9   | 15  | 26  | 37  | 40, 45, 41, 41, 44, 45 |

Over each run, the total before armor divided by the total stacks is 7.80 and 8.57 a stack. Ten percent of the bear's weapon damage at the log's attack power (attack power / 14 × 2.5 × 0.1) predicts 7.84 and 8.61. Single hits fall at 0.87 to 1.07 of the prediction, inside the paw's 20% roll. At 5 stacks a refresh averages 8.53 a stack over 6 hits, so the multiplier stops at 5. The 10% is the aura row's effect 2 (base 10, the tooltip's $s2%). A flat 10 a stack would give 10, 20 and 30 and would not move with attack power, while the per-stack value rose ×1.099 with attack power ×1.098.

The bleed matched the sim: 4 a stack each tick at both attack powers (43 apart), no armor, crits for 2×.

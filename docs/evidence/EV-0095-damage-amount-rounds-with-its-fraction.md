# EV-0095: A damage or heal amount rounds up with a chance equal to its fraction

- Claim: the amount a hit, heal or tick deals is its exact value, power and multipliers included, rounded up with a chance equal to its fraction and down otherwise, so it averages the exact value.
- Sources: Forever combat logs; the shaman coefficient check covered 9,675 shaman events (Flame Shock, shocks, Lightning Bolt, totems)
- Game build: not recorded
- Date: not recorded
- Pull request: #281

The dealt amount is always the logged base amount (the floor) or one more. The chance of one more tracks the fraction of the exact value: 1% of hits in the 0 to 10% fraction bin, 86% in the 80 to 90% bin.

| Reading                                                  | Measured                                      |
| -------------------------------------------------------- | --------------------------------------------- |
| Flame Shock Rank 1 (per-level gain 0.8) at 0 spell power | rounds up on 32 of 50 hits                    |
| Flame Shock Rank 1 at level 11 (exact 20.8), one player  | landed 21 on 21 of 26 hits and 20 on the rest |
| Flame Shock Rank 2, whose base is a whole number         | rounds up on 0 of 75                          |

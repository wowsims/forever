# EV-0019: Rend and Thunder Clap attack power shares, warrior levels 13 and 14

- Claim: a Rend Rank 2 tick takes 0.02 of attack power read at each tick and a Thunder Clap Rank 1 hit takes 0.07, the
  hit rounded down before armor.
- Sources: combat logs of a warrior at levels 13 and 14, sheet attack power 99 to 154
- Game build: not recorded
- Date: 2026-10-01
- Pull request: #172

The game files state no coefficient for either spell. Only Rend Rank 2 and Thunder Clap Rank 1 were verified here.

| Reading                            | Measured                                                                                                                 |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Rend Rank 2 (6546), tick           | 7 + 0.02 x attack power; 271 non-crit ticks over nine attack power values fit 0.02013 (± 0.00014)                        |
| Rend Rank 2, every observed tick   | crits (a flat 2x) included, fits a share between 0.01974 and 0.02020                                                     |
| Rend Rank 2, attack power read     | on every tick: cancelling Battle Shout (+21 attack power) mid Rend lowered the very next ticks, three times out of three |
| Rend Rank 2 with Improved Rend 3/3 | ticks are 1.35 times the formula, which the talent already applies                                                       |
| Thunder Clap Rank 1 (6343), hit    | floor(10 + 0.07 x attack power) before armor, 66 hits; the amount before armor never varies at a given attack power      |
| Thunder Clap Rank 1, rounding      | rounding down fits only 0.06977 to 0.07071; rounding to nearest fits no share                                            |

Ticks land on the floor or ceiling of the exact value at random, so the mean tick tracks the formula. The sim does not
floor Thunder Clap's amount, so it runs about half a point per hit above the game.

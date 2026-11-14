# EV-0004: Heroic Strike, Slam and Shield Slam threat and Shield Block's last charge, warrior levels 24 to 43

- Claim: Heroic Strike Rank 4 adds 60 threat beyond its damage and Slam Rank 1 adds 30; Shield Slam Rank 1 makes its damage plus 178 threat, added once on a crit, and adds twice the block
  value; other attacks landing in the batch of Shield Block's last charge are blocked too.
- Sources: combat logs with a StatLedger and a ThreatLog of a warrior at levels 24 to 27 and of two protection warriors,
  one at levels 40 to 42 and one at levels 41 to 43
- Game build: 1.60.1.70862
- Date: 2026-11-08 to 2026-11-09
- Pull request: #914

## Heroic Strike and Slam, levels 24 to 26

Threat over damage, before the stance multiplier, read from clean ThreatLog windows.

| Reading                                    | Measured                                                      | Implemented                         |
| ------------------------------------------ | ------------------------------------------------------------- | ----------------------------------- |
| Heroic Strike Rank 4                       | 60.08 [60.05, 60.12], 104 clean windows; Rank 3 control 39.99 | + 60                                |
| Heroic Strike Ranks 5–9                    | not measured here                                             | 2.5 a spell level, as Ranks 3 and 4 |
| Slam Rank 1 (spell level 20)               | 30.0, 29.5, 30.4 and 34.5 (4 readings, none fully clean)      | + 30, as 1.5 a spell level          |
| Slam Ranks 2–5 and the Improved Slam ranks | not measured                                                  | 1.5 a spell level                   |

## Shield Slam Rank 1, levels 40 to 43

| Reading                                | Measured                                                                                                                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Shield Slam Rank 1 threat              | mean 176.9 [175.6, 178.2], median 177.8 (n 28), flat over damage 223 to 400; seven clean crit windows read 176.9 to 181.1                                                |
| Shield Slam adds twice the block value | at block value 28 the mean is 481.1 (n 28 and n 39) against 481.1 for 2 x and 453.4 for 1 x; the 27 to 28 step favours 2 x over block value + a flat 28 about ten to one |
| Shield Slam, bonuses                   | Bastion applies to it and attack power does not (0.003 [-0.023, 0.029] a point), checked against the sheet's block value and attack power                                |

The second protection warrior's logs over block values 40 to 48 confirm both rules: no hit above the 2 x block value
ceiling in 205 hits (block value + 28 would leave 164 above it), and threat 178.10 [177.83, 178.38] (n 16). The override
table summed the first log as 177.7 over damage from 223 to 400.

## Shield Block's last charge

| Reading                                                                       | Measured                                                                   |
| ----------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| Spent auras that blocked 3 or 4 times (the warriors at 24 to 27 and 40 to 42) | two auras took 4 and 5 blocks; 6 of 204 spent auras blocked 3 to 4 times   |
| Spent auras with extra blocks (the warrior at 41 to 43)                       | 20 of 556, the extra blocks 0 to 2 ms after the last charge (one at 24 ms) |

The timing of those extra blocks against the aura's removal is #919's reading (EV-0047). Holy Shield spends charges on
blocks the same way but is not measured.

# EV-0046: Shield Slam Rank 1 threat and block value, protection warrior levels 40 to 42

- Claim: Shield Slam Rank 1 makes its damage plus 178 threat, added once on a crit, and adds twice the block value.
- Sources: combat log and ThreatLog of a protection warrior at levels 40 to 42, with a StatLedger; a second protection
  warrior at levels 41 to 43 confirmed it
- Game build: 1.60.1.70862
- Date: 2026-11-08
- Pull request: #914

| Reading                                | Measured                                                                                                                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Shield Slam Rank 1 threat              | mean 176.9 [175.6, 178.2], median 177.8 (n 28), flat over damage 223 to 400; seven clean crit windows read 176.9 to 181.1                                                |
| Shield Slam adds twice the block value | at block value 28 the mean is 481.1 (n 28 and n 39) against 481.1 for 2 x and 453.4 for 1 x; the 27 to 28 step favours 2 x over block value + a flat 28 about ten to one |
| Shield Slam, bonuses                   | Bastion applies to it and attack power does not (0.003 [-0.023, 0.029] a point), checked against the sheet's block value and attack power                                |

The second warrior's logs over block values 40 to 48 confirm both rules: no hit above the 2 x block value ceiling in 205
hits (block value + 28 would leave 164 above it), and threat 178.10 [177.83, 178.38] (n 16). The override table summed the first log as 177.7 over damage from 223 to 400.

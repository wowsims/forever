# EV-0047: Shield Block's last charge covers the rest of its batch

- Claim: other attacks landing in the batch of Shield Block's last charge are blocked too and spend nothing.
- Sources: combat logs of protection warriors at levels 41 to 43 with a StatLedger and a ThreatLog, and of two other
  warriors
- Game build: 1.60.1.70862
- Date: 2026-11-08 to 2026-11-09
- Pull request: #919, restated in #914

| Reading                                                    | Measured                                                                           |
| ---------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Extra blocks after the last charge                         | 19 of 20 came 0 to 2 ms after it, the other at 24 ms (the aura's removal at 25 ms) |
| Aura removal                                               | a median 1 ms after the last charge                                                |
| Swings logged before the removal                           | blocked 16 of 16                                                                   |
| Swings within 5 ms after the removal                       | blocked 5 of 15                                                                    |
| Spent auras that blocked 3 or 4 times (two other warriors) | two auras took 4 and 5 blocks; 6 of 204 spent auras blocked 3 to 4 times           |
| Spent auras with extra blocks (protection warrior)         | 20 of 556, the extra blocks 0 to 2 ms after the last charge (one at 24 ms)         |

Holy Shield spends charges on blocks the same way but is not measured.

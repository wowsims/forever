# EV-0004: Heroic Strike and Slam threat, warrior levels 24 to 26

- Claim: Heroic Strike Rank 4 adds 60 threat beyond its damage and Slam Rank 1 adds 30; both scale to the other ranks
  per spell level.
- Sources: combat log and ThreatLog, a warrior at levels 24 to 26
- Game build: 1.60.1.70862
- Date: 2026-11-08
- Pull request: #914

Threat over damage, before the stance multiplier, read from clean ThreatLog windows.

| Reading                                    | Measured                                                      | Implemented                         |
| ------------------------------------------ | ------------------------------------------------------------- | ----------------------------------- |
| Heroic Strike Rank 4                       | 60.08 [60.05, 60.12], 104 clean windows; Rank 3 control 39.99 | + 60                                |
| Heroic Strike Ranks 5–9                    | not measured here                                             | 2.5 a spell level, as Ranks 3 and 4 |
| Slam Rank 1 (spell level 20)               | 30.0, 29.5, 30.4 and 34.5 (4 readings, none fully clean)      | + 30, as 1.5 a spell level          |
| Slam Ranks 2–5 and the Improved Slam ranks | not measured                                                  | 1.5 a spell level                   |

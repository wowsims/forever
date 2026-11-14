# EV-0036: Rip Rank 5, Rake Rank 4 and Pounce Bleed Rank 3 shares, druid levels 55 to 56

- Claim: Rip Rank 5 takes 0.0114 of attack power a tick per combo point up to 4, Rake Rank 4 0.065 on the hit and the tick, and Pounce Bleed Rank 3 0.025 a tick.
- Sources: combat logs and StatLedger of a druid at levels 55 to 56
- Game build: 1.60.1.70888
- Date: 2026-11-12
- Pull request: #987

| Spell                                | Value                   | Fit                                                                                                                              |
| ------------------------------------ | ----------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Rip Rank 5 (9894), periodic          | 0.0114                  | 190 of 191 ticks, k 0.011365 to 0.011405; 109 of the 191 ticks were 5-point Rips and the share on all 5 points fits none of them |
| Rake Rank 4 (9904), direct           | 0.065                   | 121 of 121 hits, c 0.06498 to 0.06524                                                                                            |
| Rake Rank 4 (9904), periodic         | 0.065                   | 237 of 238 ticks                                                                                                                 |
| Pounce Bleed Rank 3 (9826), periodic | 0.025                   | 24 of 24 ticks, c 0.0248 to 0.0258                                                                                               |
| Ferocious Bite Rank 5 (22829)        | 0.037 fits its 25 bites | range 0.0354 to 0.0372, which does not pin it                                                                                    |

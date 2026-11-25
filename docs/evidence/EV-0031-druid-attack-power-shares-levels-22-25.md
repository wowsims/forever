# EV-0031: Rip, Rake and Swipe attack power shares and the failed Ferocious Bite refund, druid levels 22 to 25

- Claim: Rip Rank 1 takes 0.0114 of attack power a tick per combo point, Rake Rank 1's hit and ticks about 0.065, Swipe Ranks 1 and 2 3% on each hit, and a failed Ferocious Bite keeps its extra energy and combo points and gets 80% of its base cost back.
- Sources: combat log with a character sheet export, a Tauren druid at levels 22 to 25, mostly in Cat Form
- Game build: 1.60.1.70862
- Date: 2026-11-07
- Pull request: #908

| Reading                                         | Measured                                                                                                         |
| ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Rip Rank 1, share per tick per combo point      | 0.0114: 17 ticks at 3 and 4 combo points, 192 to 239 attack power                                                |
| Rake Rank 1, direct hit                         | (27 + 0.0648 × attack power) × Savage Fury: 45 on 26 of 26 hits at 221 attack power                              |
| Rake Rank 1, tick share                         | 0.0652: 37 on 55 of 55 ticks at 221 attack power                                                                 |
| Swipe Rank 2                                    | 0.0301 on 457 hits                                                                                               |
| Swipe Rank 1                                    | 0.03                                                                                                             |
| Ferocious Bite, energy spent on an avoided bite | 6.1 to 6.7 energy on 5 avoided bites; the extra energy and combo points kept, 80% of the 35 base energy refunded |

One share fits Rake's hit and its tick.

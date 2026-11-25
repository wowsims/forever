# EV-0211: Sunder Armor Ranks 4 and 5 threat takes 0.05 of attack power, protection warrior levels 45 to 59

- Claim: Sunder Armor Rank 4 makes floor(158 + 0.05 x attack power) and Rank 5 floor(206 + 0.05 x attack power) threat,
  times the stance, when the attack power is the one on the cast's own log line; the sheet's attack power at the cast
  reads a lower share because it differs from the cast line's on about a quarter of the casts.
- Sources: combat logs and threat log of a protection warrior at levels 45 to 59, Defensive Stance with Defiance
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938
- Date: 2026-11-09 to 2026-11-13
- Pull request: #1236

A window is the threat the warrior's threat log gained over one Sunder Armor cast on a mob it was the only one
engaged with, with exactly one Sunder Armor aura event on that mob. Every other term in the window is a rule already
measured, and the table keeps only windows whose other terms are small ones: white swings, item procs and Rend damage
times the stance, and Bloodrage ticks. Each event's threat is rounded to the nearest whole point. What is
left is the Sunder's own whole-point threat. The stance multiplier is 2.16. The attack power is the cast line's log
attack power minus the main-hand offset (14 x the average main-hand damage / 2.6, plus the Sharpened weapon enchant's
bonus while it ran), set against the sheet attack power the stat ledger held at the cast. Windows next to a window edge
are left out. The base is the row's threat effect, 158 for Rank 4 and 206 for Rank 5.

| Reading                                                | Rank 4 (11596), 120 windows     | Rank 5 (11597), 42 windows     |
| ------------------------------------------------------ | ------------------------------- | ------------------------------ |
| round(2.16 x floor(base + 0.05 x cast-line AP)), exact | 98 of 120 (99 within 1)         | 38 of 42                       |
| The same without the floor, exact                      | 16 of 120                       | 7 of 42                        |
| Base and share both free (base +- 25, share 0.02-0.07) | best 98 at (158, 0.05)          | best 38 at (206, 0.05)         |
| Slope against cast-line AP                             | 0.0509 [0.0485, 0.0533] (n 106) | 0.0480 [0.0445, 0.0514] (n 39) |
| Intercept                                              | 156.8 [155.1, 158.4]            | 206.9 [204.2, 209.5]           |
| Slope within levels                                    | 0.0503 [0.0475, 0.0531]         | 0.0479 [0.0445, 0.0514]        |
| The row model with the sheet AP, exact                 | 67 of 120                       | 29 of 42                       |
| Slope against sheet AP                                 | 0.0410 [0.0354, 0.0465]         | 0.0402 [0.0316, 0.0489]        |
| Slope within levels against sheet AP                   | 0.0350 [0.0292, 0.0408]         | 0.0400 [0.0314, 0.0486]        |

The slopes use the windows whose residual against the row model is under 7 points. 20 of the Rank 4 windows and 4 of
Rank 5's sit more than 5 points off the row model. The sheet attack power equals the cast line's on 84 of 114 Rank 4
casts and differs by 42 to 174 on the other 30.

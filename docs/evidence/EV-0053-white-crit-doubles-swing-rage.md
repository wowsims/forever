# EV-0053: A white crit doubles the swing's whole rage

- Claim: a warrior's critical auto attack doubles the swing's whole rage, the two-hand factor included, and an off-hand crit doubles the off-hand's halved share.
- Sources: combat logs of two-hand warriors (no off-hand swings, a 3.2 to 3.8 s swing gap) and of one-hand warriors, from Forever
- Game build: not recorded
- Date: 2026-10-09 to 2026-10-14
- Pull request: #780

A swing's step is its rage field less the warrior's previous rage block.

| Warrior (log date)   | Non-crit     | Crit         | Crits | Ratio |
| -------------------- | ------------ | ------------ | ----- | ----- |
| Two-hand, 2026-10-12 | 16.2         | 32.4         | 13    | 2.00  |
| Two-hand, 2026-10-12 | 14.8         | 29.7         | 8     | 2.01  |
| Two-hand, 2026-10-13 | 14.8 to 15.1 | 29.7 to 30.0 | 3     | 2.0   |
| Two-hand, 2026-10-09 | 15.2         | 30.2         | 1     | 1.99  |
| Two-hand, 2026-10-14 | 14.8         | 29.7         | 1     | 2.01  |

One-hand crits double (5.5 to 11.1 on 85 crits; 9.7 to 19.5 on 17), and an off-hand crit doubles the off-hand's halved share. Logs from before 2026-10-09 show the old rule: a two-hand warrior at a 3.4 s swing (2026-10-07) logs 15.3 on crit and non-crit swing lines alike, and the next own event carries 11.8 more, 0.77 of the swing.

The pull request measured no bear swing; the bear's crit rage is the reading of EV-0037.

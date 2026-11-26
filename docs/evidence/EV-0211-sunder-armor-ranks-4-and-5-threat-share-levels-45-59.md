# EV-0211: Sunder Armor Ranks 4 and 5 threat takes 0.05 of attack power, protection warrior levels 45 to 59

- Claim: Sunder Armor Rank 4 makes floor(158 + 0.05 x attack power) and Rank 5 floor(206 + 0.05 x attack power) threat,
  times the stance, when the attack power is the one on the cast's own log line, which counts attack power against the
  target's creature type; the sheet's attack power at the cast reads a lower share because it differs from the cast
  line's on about a quarter of the casts.
- Sources: combat logs and threat log of a protection warrior at levels 45 to 59, Defensive Stance with Defiance
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938
- Date: 2026-11-09 to 2026-11-13
- Pull request: #1236
- Later count: Sunder Armor Rank 5, §Count of the 2026-11-24 to 2026-11-26 logs

A window is the threat the warrior's threat log gained over one Sunder Armor cast on a mob it was the only one
engaged with, with exactly one Sunder Armor aura event on that mob. Every other term in the window is a rule already
measured, and the table keeps only windows whose other terms are small ones: white swings, item procs and Rend damage
times the stance, and Bloodrage ticks. Each event's threat is rounded to the nearest whole point. What is
left is the Sunder's own whole-point threat. The stance multiplier is 2.16. The attack power is the cast line's log
attack power minus the main-hand offset (14 x the average main-hand damage / 2.6, plus the Sharpened weapon enchant's
bonus while it ran), set against the sheet attack power the stat ledger held at the cast. Windows next to a window edge
are left out. The base is the row's threat effect, 158 for Rank 4 and 206 for Rank 5.

| Reading                                                | Rank 4 (11596), 120 windows     |
| ------------------------------------------------------ | ------------------------------- |
| round(2.16 x floor(base + 0.05 x cast-line AP)), exact | 98 of 120 (99 within 1)         |
| The same without the floor, exact                      | 16 of 120                       |
| Base and share both free (base +- 25, share 0.02-0.07) | best 98 at (158, 0.05)          |
| Slope against cast-line AP                             | 0.0509 [0.0485, 0.0533] (n 106) |
| Intercept                                              | 156.8 [155.1, 158.4]            |
| Slope within levels                                    | 0.0503 [0.0475, 0.0531]         |
| The row model with the sheet AP, exact                 | 67 of 120                       |
| Slope against sheet AP                                 | 0.0410 [0.0354, 0.0465]         |
| Slope within levels against sheet AP                   | 0.0350 [0.0292, 0.0408]         |

The slopes use the windows whose residual against the row model is under 7 points. 20 of the Rank 4 windows sit more
than 5 points off the row model. The sheet attack power equals the cast line's on 84 of 114 Rank 4
casts and differs by 42 to 174 on the other 30.

## Count of the 2026-11-24 to 2026-11-26 logs

- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at level 60 in dungeon groups, Defensive
  Stance with Defiance 3
- Game build: 1.60.1.71091 on 2026-11-24 and 1.60.1.71139 from the 2026-11-25 18:17 login; the hotfix is not recorded;
  no reading moves between the two builds
- Date: 2026-11-24 to 2026-11-26
- Pull request: #1254

The method is EV-0220's, against the cast line's attack power. From 2026-11-26 00:23 the warrior wears Seal of the Dawn,
which gives 78 attack power against Undead; the windows before it and after it are counted apart.

| Reading                                                 | Rank 5 (11597)                                                                                                             |
| ------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| round(2.16 x floor(206 + 0.05 x cast-line AP)), exact   | 63 of 74 before the seal (65 within 1); 19 without the floor, 44 with the sheet AP; base 204, 205 or 208 on none, 207 on 5 |
| Slope against cast-line AP, windows within 5 of the row | 0.0510 [0.0495, 0.0526] (n 71, AP 684 to 989), intercept 204.8                                                             |
| After the seal, exact with 78 more AP on Undead targets | 29 of 37; the model without it exact on 4, neither on 4; 70, 75, 80 and 90 more AP exact on 15, 28, 27 and 15              |

# EV-0059: Seal of Righteousness hand factors, level 20 beta

- Claim: a Seal of Righteousness proc is its value times the weapon speed times a hand factor (1.2 for two-handers, 0.85
  for one-handers from 2.0 speed up, less below), plus one spell power coefficient per hand type.
- Sources: test log on the beta at level 20, ranks 1 to 3, eighteen vendor weapons from 1.7 to 3.7 speed, with and
  without spell power and Judgement of the Crusader
- Game build: 1.60.1 beta
- Date: 2026-09-26
- Pull request: #110

```
proc = value/100 x speed x hand(speed) + coefficient x (spell power + Judgement of the Crusader)
  two-hander:  hand = 1.2,                               coefficient 0.22
  one-hander:  hand = 0.85 - 0.2 x max(0, 2.0 - speed),  coefficient 0.20
```

| Reading                   | Measured                                                                                                                                                    |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Two-hander factor         | 1.2, exact from 2.6 to 3.7 speed (the pull request states it held at every speed and rank)                                                                  |
| One-hander factor         | 0.85 from 2.0 speed up; 0.83 at 1.9, 0.81 at 1.8, 0.79 at 1.7, which is as fast as the test could reach                                                     |
| Fast one-hander shape     | the straight line through those points is one of two shapes the data allows; they agree to a percent down to 1.5 speed                                      |
| Spell power coefficient   | 0.20 on one-handers and 0.22 on two-handers, flat across speed and rank; the judgement dummy's per rank number (0.185 at rank 3, 0.058 at rank 1) is not it |
| Judgement of the Crusader | counts as spell power at the same coefficient                                                                                                               |
| Crits                     | double the unrounded total                                                                                                                                  |

The full test log was kept in a write-up outside the repository.

The spell power coefficients here are replaced by the 0.1 and 0.12 that EV-0060 read at level 30 on game build
1.60.1.70425; the hand factors stand.

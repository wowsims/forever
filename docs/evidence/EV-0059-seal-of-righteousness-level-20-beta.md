# EV-0059: Seal of Righteousness hand factors and spell power coefficients

- Claim: a Seal of Righteousness proc is its value times the weapon speed times a hand factor (1.2 for two-handers, 0.85
  for one-handers from 2.0 speed up, less below), plus one spell power coefficient per hand type: 0.1 on one-handers and
  0.12 on two-handers. The tooltip's 0.0765 fits none of the one-handers.
- Sources: test log on the beta at level 20, ranks 1 to 3, eighteen vendor weapons from 1.7 to 3.7 speed, with and
  without spell power and Judgement of the Crusader
- Game build: 1.60.1 beta
- Date: 2026-09-26
- Pull request: #110
- Later count: the spell power coefficients, §Count of the 2026-10-16 test

```
proc = value/100 x speed x hand(speed) + coefficient x (spell power + Judgement of the Crusader)
  two-hander:  hand = 1.2,                               coefficient 0.12
  one-hander:  hand = 0.85 - 0.2 x max(0, 2.0 - speed),  coefficient 0.10
```

| Reading                   | Measured                                                                                                                                    |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Two-hander factor         | 1.2, exact from 2.6 to 3.7 speed (the pull request states it held at every speed and rank)                                                  |
| One-hander factor         | 0.85 from 2.0 speed up; 0.83 at 1.9, 0.81 at 1.8, 0.79 at 1.7, which is as fast as the test could reach                                     |
| Fast one-hander shape     | the straight line through those points is one of two shapes the data allows; they agree to a percent down to 1.5 speed                      |
| Spell power coefficient   | §Count of the 2026-10-16 test; flat across ranks 1 to 3; the judgement dummy's per rank number (0.185 at rank 3, 0.058 at rank 1) is not it |
| Judgement of the Crusader | counts as spell power at the same coefficient                                                                                               |
| Crits                     | double the unrounded total                                                                                                                  |

The full test log was kept in a write-up outside the repository.

## Count of the 2026-10-16 test

- Sources: test on the beta at level 30, rank 4 (judgement 20282, per hit value 541 + 31 x 4 = 665), three one-handers and
  two two-handers, each hit at 0 and 50 spell power with no talents, buffs or Judgement of the Crusader
- Game build: 1.60.1.70425 beta
- Date: 2026-10-16
- Pull request: #473

At 0 spell power the weapon part, 665/100 x speed x hand, matches every weapon, crits included. A hit rounds to the
integer either side of its value and a crit doubles the unrounded value before rounding, so each set of hits and crits at
50 spell power bounds the weapon's value and with it the coefficient.

| Weapon           | Hand | Speed | Base  | Hits at 0 SP | Crits at 0 SP | Hits at 50 SP | Crits at 50 SP | Value      | Coefficient    |
| ---------------- | ---- | ----- | ----- | ------------ | ------------- | ------------- | -------------- | ---------- | -------------- |
| Stock Shortsword | 1H   | 1.80  | 9.70  | 9, 10        | 19, 20        | 14, 15        | 30             | 14.5 to 15 | 0.096 to 0.106 |
| Cutlass          | 1H   | 2.20  | 12.44 | 12, 13       | 25            | 17, 18        | 34, 35         | 17 to 17.5 | 0.091 to 0.101 |
| Cruel Barb       | 1H   | 2.80  | 15.83 | 15, 16       | 31, 32        | 20, 21        | 41, 42         | 20.5 to 21 | 0.094 to 0.104 |
| Espadon          | 2H   | 2.90  | 23.14 | 23, 24       | 46, 47        | 29            | 58, 59         | 29 to 29.5 | 0.117 to 0.127 |
| Corpsemaker      | 2H   | 3.80  | 30.32 | 30, 31       | 60, 61        | 36, 37        | 72, 73         | 36 to 36.5 | 0.114 to 0.124 |

All three one-handers fit a single value from 0.096 to 0.101, and 0.10 sits in that range. One coefficient holds from 1.8
to 2.8: it does not grow with speed and does not take the fast weapon penalty. The 2.2 and 2.8 one-handers gain the same
5 damage from 50 spell power. With 0.0765 the Stock Shortsword would hit for about 13.5 and the Cutlass for about 16.3,
where they hit for 14 to 15 and 17 to 18. The two-handers overlap from 0.117 to 0.124, so 0.12 holds there at both
speeds. Below 1.8 speed the coefficient is not measured.

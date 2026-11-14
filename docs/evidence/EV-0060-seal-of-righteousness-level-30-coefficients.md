# EV-0060: Seal of Righteousness spell power coefficients, level 30

- Claim: a Seal of Righteousness proc takes 0.1 of spell power on one-handers and 0.12 on two-handers; the tooltip's
  0.0765 fits none of the one-handers.
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

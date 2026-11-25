# EV-0214: Weaponmaster (Sword) rolls on every landed Whirlwind strike and grants one extra attack a cast

- Claim: Every landed Whirlwind strike made with a sword rolls Weaponmaster (Sword), on every target and with either
  hand, at the chance a landed white swing rolls; a cast that wins more than one roll still grants one extra attack.
  Sweeping Strikes' hits beside the cast do not roll it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the warrior streams on disk with a sword in either hand
  in reports that log the extra attack 1257049: two-handed sword 87 streams of 72 warriors in the 87 reports under
  Reports, a sword in each hand 5 streams of 5 warriors in 5 reports
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-24
- Pull request: not yet opened

## Method

Each Whirlwind cast (1680) with the warrior's own events in the 2 ms on either side: the landed main-hand strikes
(1680) and off-hand strikes (462891), the Sweeping Strikes hits (26654, 462890), and the extra attack events 1257049
from 2 ms before to 25 ms after the cast. An event a report repeats under two fight ids counts once. The weapons are
the fight's gear.

Each warrior's chance is its own: the share of its landed white swings with the same weapons that an extra attack
follows within the same window. Only warriors with 100 or more such swings and a share of at least 2% are kept, which
leaves out those without the talent. A cast with a white swing in the 40 ms before it, or with more white swings in the
40 ms after it than extra attacks, is left out, since that extra attack may be the swing's.

Each model predicts the casts that grant an extra attack from the warrior's chance p and the landed sword strikes k of
the cast: one roll a cast predicts p, one roll a landed strike predicts 1 − (1 − p)^k.

## Numbers

| Weapons, report build | Casts | Warriors' streams | Casts with an extra attack | A roll a strike | One roll a cast |
| --------------------- | ----- | ----------------- | -------------------------- | --------------- | --------------- |
| Two-handed, 70938     | 1,519 | 28                | 216                        | 243.8           | 90.7            |
| Two-handed, 70959     | 1,409 | 24                | 191                        | 188.3           | 70.7            |
| Two-handed, 71037     | 2,305 | 36                | 334                        | 313.1           | 118.3           |
| Two-handed, all       | 5,233 | 87                | 741                        | 745.3           | 279.7           |
| Sword in each hand    | 282   | 5                 | 56                         | 55.2            | 13.6            |

The two-handed casts read 438 against 465.2 a strike before hotfix push 114313 (2026-11-20 23:47 UTC) and 303 against
280.1 after it. With a sword in each hand, by report build: 70938 reads 32 against 25.7 a strike, 70959 20 against 26.0
and 71037 4 against 3.4. By landed strikes in the two-handed cast:

| Landed strikes | Casts | With an extra attack | A roll a strike | One roll a cast |
| -------------- | ----- | -------------------- | --------------- | --------------- |
| 1              | 653   | 34                   | 33.5            | 33.5            |
| 2              | 1,298 | 130                  | 133.4           | 68.8            |
| 3              | 1,518 | 239                  | 230.6           | 81.9            |
| 4              | 1,764 | 338                  | 347.8           | 95.4            |

With a sword in each hand, one roll a hand predicts 25.4 and a roll on main-hand strikes alone 29.5, against 56 seen; a
roll on every landed strike of either hand predicts 55.2.

No cast logs two extra attack events: 0 of the 797 casts that grant one in the tables, and 0 of the 896 that grant one
among every cast with no white swing in the 40 ms before it, whoever the warrior. Of those 896, 739 log one white swing
in the 40 ms after the cast, 11 two and 146 none, so a cast that grants swings once.

Sweeping Strikes' hits: over all two-handed casts, a roll on each Sweeping Strikes hit as well predicts 839.9 against 741. Split, the casts with Sweeping Strikes hits read 373 against 333.8 for the strikes alone and 428.4 with the hits,
and the casts without read 368 against 411.5.

## Reports

Two-handed sword: 1rDVF8GtakMjHXfY, 1RBjZLNxdMknWGfJ, 1WHqx7bPa8mkGydJ, 2x7ZjvdbRwFMWKTP, 2YvMKFznqp3D9PaA,
2ZHDh8LX4mKqctbz, 2zXhygVYpkrMZGxT, 3jFLzH68Jnxy1cTK, 3X9ychCWZz1vTVxA, 4VtCK6Qqr1MD2hpc, 6Krp8bkfx4cRjBaz,
6QvVFqNYty8bzrcf, 6Y2mMtCzb7ZvQKWD, 8mhjr7afgtPT6DFQ, 8W1aADknmKzH436M, 8ZdnfAT79zPwmyVM, a2ZkBR4b6HcFf7dW,
acmr3KgRtMYD2hnW, aDHTPGq6htJFx8Nw, Ag69GWdLRDhKCrcT, aPDTGRxFmMhZAzYW, ARwPXzLJk78HgVQq, ATNG2ZtRgm1H7cxP,
aYyJ9hdFwXKgDk16, B6xYXRgnPWaMJFK8, BfYZrMW4xdLvk36F, BgGa4DXLy1vCnpMz, Bm94aWXPCyqfJ7kn, bR9LQtgmKwZ3r6kW,
CB437hRzJAktQN6V, CFYhApL7jHWvKn34, CMLxXY2AQdv9bgDp, cRZ2WdrbxNYtQ4kM, CV1DwP4Bct23AzGY, cW6gZwyJqaHd8CBD,
d12XjHmazMqCFRt7, DMhF1q6Kd7XkVLn4, DYV6MtCX24qZKn3h, fdj2MNckqQ3YDHmR, fQhgAL7Fnaw3cBpk, gxnFmGwCYzKLbhft,
hCPLc2G7KAVX8Bby, HRnmMtD7Fp41wb2g, HT2qmWDJNfXVj1xA, J6dCKaR1tq8TYZnk, maYP7WvxJ2TzGtDH, mcYgjy9xFP7VB8Qh,
MjfT3pQKRWcVG4gY, mjYGdKhwCQ4D6ATf, MwVH1KQmak73NPd8, mYaWhF4n8zRLDZtw, n2HkhcYFrLK1tfZj, N9PGBMt7faY6W2J8,
nhtYVjb7CAw9DvQR, nqDL2hKfQB4yvjgJ, nwh3zdqc4CjXa6MG, nzBqZ8camA92drQb, P1yKzX3Q8NgFYxZH, pazJDTkqbAW1f6wZ,
pHdqPxhAYLQTwya7, prgqw1cBA8kZLQxT, PyMC9aDNr4wQLbxX, QFNhp8VrBzyPtGqd, qtNCmzfYQwAnRMkZ, RFGVCxDvb8mf2dPT,
rHM6TPA7NVckxQmh, RkXgz24GLmfvxCNF, tkPZYqTRgCyWG3jM, tnyh4dxq6pNXQzLw, tVhnGRkHP7f2xarB, tvhPG4q2KydF6Qwj,
v23ytJHDz6QkbXnN, VLz78BfZyWa3Cxr2, VtcPT63wQ7WaACvj, WFMjbyKYzvfNkPhH, WwAp7qrmHYNDPzfg, WY7gZXMxFVjt6pbK,
x6dCXkvmWH8LDnA7, xd64GKNcb7VCqarf, xnf4K91LT7A2mNaw, XqgW7x3ym6cwY2C8, YAQ1TL6rbFck3JNC, YF7mAtMxfqL3w6JD,
YrVa1xmJvBMTHgp4, z8yCGH2gpjabZ7f6, zKB98PbWwqFf6cxp, zWPNCbpYTV1xJ2FH.

A sword in each hand: 782JNDc1pgGRZxVC, AvXNG1TpnyLJMxVt, gDtFwpGPLM2TyZKB, mPMpkAgjbtR7HBXD, y6FZRHQnckD8YjAC.

# EV-0182: Sweeping Strikes' copied hit takes the attacker's damage bonuses again

- Claim: Sweeping Strikes' copied hit (12723) is the hit it copies, crit included, scaled by the two targets' armor,
  and the attacker's damage bonuses apply to it again: Enrage's +10% physical damage done shows up twice.
- Sources: Warcraft Logs, the event streams of the 148 warriors in 148 reports that log 12723. The Enrage comparison
  comes from 40 of them: bgDPx6qQ4tYrmBKT, mcYgjy9xFP7VB8Qh, AKXn47BZWpyhJRv3, 8ZdnfAT79zPwmyVM, 96drQ4BYwaqKC7NH,
  Wn8wHhaxy31MLzjp, nhtYVjb7CAw9DvQR, 8GNt2hC7v1rWRXyc, bMF6ZAPf7BQLakgn, tkPZYqTRgCyWG3jM, 16GhDNRmVn8KWjBT,
  Ag69GWdLRDhKCrcT, aYyJ9hdFwXKgDk16, bR9LQtgmKwZ3r6kW, BRxGa38JnVhPKw4A, fQhgAL7Fnaw3cBpk, HT2qmWDJNfXVj1xA,
  J6dCKaR1tq8TYZnk, q9RQgCbP4GjYfv86, RjvDtZMPhLm7g6XQ, VLz78BfZyWa3Cxr2, cbkDQfXNPGj3tnWT, n2HkhcYFrLK1tfZj,
  pHdqPxhAYLQTwya7, tVhnGRkHP7f2xarB, 1ZyhaVDdFtTYWMcC, 2hG49JzynDkwMVCY, 2ZHDh8LX4mKqctbz, 9gJ1KcWkwn7dQpmP,
  aDHTPGq6htJFx8Nw, ARwPXzLJk78HgVQq, aZKTvVcCHLwhgXzM, GMrxRanAp9fJ46bQ, hCPLc2G7KAVX8Bby, MwVH1KQmak73NPd8,
  nqDL2hKfQB4yvjgJ, qtNCmzfYQwAnRMkZ, RFGVCxDvb8mf2dPT, tnyh4dxq6pNXQzLw, WFMjbyKYzvfNkPhH.
- Game build: 1.60.1, the builds of 2026-11-07 to 2026-11-22 (70888 to 70959); the hotfix each player ran is not
  recorded
- Date: 2026-11-07 to 2026-11-22
- Pull request: the one adding this file

## Method

Each copied hit is paired with the warrior's last damage event within 60 ms before it on another target, other than
12723 and 26654, with damage and no absorb: 20,650 pairs. Enrage (12880, +10% physical damage done) is tracked from the
warrior's own apply and remove events.

The ratio of a copy to its hit depends on the two targets' armor, which Sunder Armor and other debuffs change. So the
comparisons take two copies on the same pair of targets (the same target ids and instances, the hit on the first, the
copy on the second) at most 10 s apart, both from hits that did not crit, and compare their copy-to-hit ratios.

## The copy follows its hit

| Two copies on the same target pair, ≤ 10 s apart | Pairs | Ratio change, median | Middle half      | Unchanged |
| ------------------------------------------------ | ----- | -------------------- | ---------------- | --------- |
| Any hit, same Enrage state                       | 5,753 | +0.00%               | −0.61% to +0.43% | 29.7%     |
| One hit a crit, one not                          | 2,328 | −0.03%               | −0.33% to +0.20% |           |

The hits themselves differ by a median 5.66% from one to the next, so the copy is a copy of its hit, not a damage roll
of its own, and it carries the hit's crit. No copy crits itself (all 20,650 log hit type 1).

## Enrage counts twice

| Between the two copies | Pairs | Ratio change, median | Middle half       | Unchanged |
| ---------------------- | ----- | -------------------- | ----------------- | --------- |
| Enrage came up         | 96    | +9.85%               | +3.86% to +11.31% | 0%        |
| Enrage went down       | 19    | −6.88%               | −9.12% to −2.27%  | 0%        |
| No change (above)      | 5,753 | +0.00%               | −0.61% to +0.43%  | 29.7%     |

A hit under Enrage already carries its +10%; the copy moves by Enrage's bonus once more.

Death Wish is not in these logs (the warriors with Sweeping Strikes do not take it), and a bonus that is always on,
such as Two-Handed Weapon Specialization, cannot be told apart from the armor difference between the targets.

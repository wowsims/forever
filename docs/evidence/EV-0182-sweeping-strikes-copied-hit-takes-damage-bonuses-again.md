# EV-0182: Sweeping Strikes' copied hit takes the attacker's damage bonuses again

- Claim: Sweeping Strikes' copied hit (12723) is the hit it copies, crit included, priced again on the second target's
  armor, and the attacker's damage bonuses apply to it a second time, except one that needs a weapon: Enrage's +10%
  physical damage done counts twice, Two-Handed Weapon Specialization once.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every fight of the 148 reports under Reports, from the
  event streams of the 148 warriors in them that log 12723.
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-07 to 2026-11-22
- Pull request: #1177

## Method

Each copied hit is paired with the warrior's last damage event within 60 ms before it on another target, other than
12723 and 26654, with damage and no absorb: 20,650 pairs. Enrage (12880, a row of +10% physical damage done) is
tracked from the warrior's own apply and remove events; the talent rank behind it is not logged. Sunder Armor counts
only the warrior's own applications, since a stream holds no other player's debuffs.

The ratio of a copy to its hit depends on the two targets' armor, which Sunder Armor and other debuffs change. So the
comparisons take two copies on the same pair of targets (the same target ids and instances, the hit on the first, the
copy on the second) at most 10 s apart and compare their copy-to-hit ratios. Every row takes two hits that did not crit,
except the row that pairs a crit with a hit that did not.

## The copy follows its hit

| Two copies on the same target pair, ≤ 10 s apart | Pairs | Ratio change, median | Middle half      | Unchanged |
| ------------------------------------------------ | ----- | -------------------- | ---------------- | --------- |
| Same Enrage state, no Sunder by the warrior      | 5,670 | +0.00%               | −0.55% to +0.44% |           |
| Same Enrage state                                | 5,753 | +0.00%               | −0.61% to +0.43% | 29.7%     |
| One hit a crit, one not                          | 2,328 | −0.03%               | −0.33% to +0.20% |           |

The hits themselves differ by a median 5.66% from one to the next, so the copy is a copy of its hit, not a damage roll
of its own, and it carries the hit's crit. No copy crits itself (all 20,650 log hit type 1).

## Armor: the second target's

| Between the two copies, same Enrage state | Pairs | Ratio change, median | Middle half       |
| ----------------------------------------- | ----- | -------------------- | ----------------- |
| The warrior's Sunder on the hit's target  | 68    | −5.53%               | −9.69% to −4.59%  |
| The warrior's Sunder on the copy's target | 10    | +5.39%               | −6.39% to +12.29% |

A Sunder on the first target raises the hit and leaves the copy; one on the second raises the copy. The copy takes the
hit's damage before the first target's armor and pays the second target's.

## Enrage counts twice

| Between the two copies | Pairs | Ratio change, median | Middle half       | Unchanged |
| ---------------------- | ----- | -------------------- | ----------------- | --------- |
| Enrage came up         | 96    | +9.85%               | +3.86% to +11.31% | 0%        |
| Enrage went down       | 19    | −6.88%               | −9.12% to −2.27%  | 0%        |
| Same Enrage state      | 5,753 | +0.00%               | −0.61% to +0.43%  | 29.7%     |

A hit under Enrage already carries its +10%; the copy moves by Enrage's bonus once more. Taking a 10% bonus away
predicts −9.09%; the 19 pairs read −6.88%, and why is open.

## Two-Handed Weapon Specialization counts once

Of the 135 warriors whose main hand is logged, 120 wield a two-hander. Without Enrage, 400 of their 11,230 copies
(3.6%) equal the hit exactly, from 67 of them; the 15 with one-handers show 21 of 1,057 (2.0%). Two-Handed Weapon
Specialization (12163) names a two-handed weapon as its requirement and the copy has none. Were its bonus applied
again, a copy on a target with the hit's armor would come out above its hit, never equal to it. The logs hold no
talents, so this rests on the two-hander warriors with Sweeping Strikes taking it.

Death Wish does not show: none of these 148 warriors takes it. A flat physical damage bonus is not measured.

## Reports

16GhDNRmVn8KWjBT, 176QTFWR8PtdfGZB, 1J6j9t2qDYdKMwLp, 1WHqx7bPa8mkGydJ, 1ZyhaVDdFtTYWMcC, 214ndcFyjpPM8wah,
23ZyGvk7BXAVDT4F, 2hG49JzynDkwMVCY, 2qk4MYw7cr8KGFtm, 2RdDXxwkKcm3Mf6q, 2vMwVxLqdB3GRZn6, 2x7ZjvdbRwFMWKTP,
2YvMKFznqp3D9PaA, 2ZHDh8LX4mKqctbz, 2zXhygVYpkrMZGxT, 3vVtMWRcg9H4FTPw, 4nmJDcGNkZC2Phdw, 4VtCK6Qqr1MD2hpc,
68fgHQ74dRxrwac3, 6jHPJBnLy1pF9Rkf, 6q1mWrtM7zXFckvp, 6wxzV4YyNCh1aqfd, 6Y2mMtCzb7ZvQKWD, 782JNDc1pgGRZxVC,
7tL3AyRHfawBJF94, 8GHAfMvNmbRrWQKj, 8GNt2hC7v1rWRXyc, 8ZdnfAT79zPwmyVM, 96drQ4BYwaqKC7NH, 9gJ1KcWkwn7dQpmP,
9pMcvgxrQNfLZ7hd, aDHTPGq6htJFx8Nw, Ag69GWdLRDhKCrcT, AKXn47BZWpyhJRv3, aPDTGRxFmMhZAzYW, ARwPXzLJk78HgVQq,
aYyJ9hdFwXKgDk16, aZKTvVcCHLwhgXzM, b186j7df4hKxNPkV, b196ArgRdLNX42ym, bgDPx6qQ4tYrmBKT, BgGa4DXLy1vCnpMz,
bhyYq9HXd4xvAwtJ, bMF6ZAPf7BQLakgn, bPQCc3Nvq27HRWt6, bR9LQtgmKwZ3r6kW, BRxGa38JnVhPKw4A, bxrzpqBVNQFZwjmT,
c7gHWKFrZMnm2yb4, cav7VZrmPbNW3y29, cbkDQfXNPGj3tnWT, CFYhApL7jHWvKn34, CJFAZxNLH4BDV9Ka, CRDXNJwfamHWTx2K,
cZ9B6Ay3RDTWgq4K, D9frqPV8dam7B1AM, dAV7JrjMkGfKNZtY, DJ436qAXGnBHLdTa, DQ62GWmAajLh4zbN, dQMKjpxJBCz1wrDW,
fBrKy1M2T6q9LC3g, fdj2MNckqQ3YDHmR, FN8krKfCb3BXWmwV, FNkQ3BDhAfzM9aZP, fQhgAL7Fnaw3cBpk, G4hzybWYmgrXvL83,
ga9Qk6DcjJmNTGVH, gaDnjKvP6RkQTW2B, gBqYT9Da3KkpXHQj, gDtFwpGPLM2TyZKB, GMrxRanAp9fJ46bQ, Gv6NDzAmFjC4P3qL,
gvWZG1cjrChKwdtz, H3bRaDmzyjxTW2BJ, h9XqNAJk6aPntHpv, hCPLc2G7KAVX8Bby, HRnmMtD7Fp41wb2g, HT2qmWDJNfXVj1xA,
HZP6h9AtRNK4n3Ck, j3F9NHxqtagPYGkJ, J6craWFxPjTAGv7t, J6dCKaR1tq8TYZnk, jtn8ap4XY6PzwDQL, JZv9XDFLC1yKxrPY,
K6VvbDypXFLQJW7d, K8rhAtLjcz7gdHkn, Kfk36GzDwnPVbHaN, kJq81KxZyTHD3A6z, KrtPndVLH4BjJcfD, ky84M92RxVhwpfFB,
maYP7WvxJ2TzGtDH, mcYgjy9xFP7VB8Qh, MjfT3pQKRWcVG4gY, MwVH1KQmak73NPd8, mYaWhF4n8zRLDZtw, n2HkhcYFrLK1tfZj,
nhtYVjb7CAw9DvQR, nPLWN2VDTZMz7Qk6, nqDL2hKfQB4yvjgJ, NtakrWH8xvRmzpCg, nX1qbatTzGZ3CAwB, NXMZV9xjvWpB7GtP,
nZ3NhbvpVXg649Bm, P1yKzX3Q8NgFYxZH, Pc3gyq8AFL9CD64m, pHdqPxhAYLQTwya7, q9RQgCbP4GjYfv86, qfG6QWjd3wg1JVXv,
qprynNbXtRZV4DHP, qR2rnzfgPakD879W, qtNCmzfYQwAnRMkZ, qVtTPaDMHfhmG9gr, RFGVCxDvb8mf2dPT, RjvDtZMPhLm7g6XQ,
rXvTBhVgpZmc1kbP, t6dr1MY8zTJBWnkg, tDT4JfyvdZ2AQNgR, tkPZYqTRgCyWG3jM, tnyh4dxq6pNXQzLw, tVhnGRkHP7f2xarB,
txP9jAb7gGHhXckB, v4QYKcD7XymZq2JB, v8P6TmD34M9bLJtx, vAGVD7aQZxNWX8h1, Vd1NZCgLMqxhb9kP, vdwBDjmZRykqHaWL,
vjykNnZ7b6BmTVhJ, VLmAcpbQxWf1CtaR, VLz78BfZyWa3Cxr2, vyzbt6kgRhTq4fdr, w9jRArgMBvnYt7JX, WFMjbyKYzvfNkPhH,
Wn8wHhaxy31MLzjp, wrVYM4xbWLNaczhP, xbvpjkzYJ7hqGw4Z, xnrg9m6jDMR1kaLN, y6FZRHQnckD8YjAC, ybFmNGwXcnBQW1rz,
yC6t8PJx9VRLrj4F, YX61J8CKqc42paNZ, z3YfqBgbt1GjFD4W, zc1hygaCXjDAPxL6, ZJgFdtvGTYQVCqyb, ZK8YCvy9wba13cpk,
ZkA3Vhz72n9Gy4dY, ZQtv6AqYmdJFWwhy, ZTxYXrtHQkcNGjA1, zXAnmG7MqPwBcyjK.

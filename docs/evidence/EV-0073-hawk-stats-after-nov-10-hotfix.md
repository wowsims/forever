# EV-0073: Summon Hawk health, attack power and swing after the November 10 hotfix

- Claim: after the 2026-11-10 hotfix a hawk takes 80% of the owner's attack power as the log shows it, the ranged attack power plus 14 a point of the main hand's dps, and its swing before armor at level 60 is 0.153 of the creature dps over 2.5 s plus 2.5/60 of its attack power.
- Sources: three combat logs of one level 48 hunter taken after the 2026-11-10 hotfix (naked; geared; geared with Aspect of the Hawk), 357 swings in all; report codes not recorded
- Game build: not recorded
- Date: after 2026-11-10
- Pull request: #950
- Later count: the hawk's swing and the weapon part of the owner's attack power, §Count of the 2026-11-05 to 2026-11-16 reports

The hotfix raised the melee damage of the Summon Hawk guardian by about 80%. The row that scales a hawk, Hunter Guardian Hawk Pet Scaling 01 (1293586), is not in the store, so the swing is fitted to the logs.

| State                            | Sheet ranged attack power | Owner attack power in the log | Hawk attack power | Swing before armor |
| -------------------------------- | ------------------------- | ----------------------------- | ----------------- | ------------------ |
| Naked                            | 345                       | 352                           | 282               | 28-37              |
| Geared                           | 747                       | 1180                          | 944               | 56-64              |
| Geared, more ranged attack power | about 792                 | 1235                          | 988               | 58-66              |
| The same with Aspect of the Hawk | about 882                 | 1325                          | 1060              | 60-69              |

- The swing is a roll plus 2.5/60 of the hawk's attack power, 60 attack power a point of dps, with Unleashed Fury 5/5 and Focused Fire 2/2, which reach the hawks; §Hawk swing gives the roll's share of the creature dps. At level 48 the roll spanned 22% either side of its mean. Logs before the hotfix read 0.0223 a point of the hawk's attack power, so the hotfix made the attack power part 1.8 times what it was.
- The hawk's attack power is 80% of the owner's attack power as the log shows it. That number is the ranged attack power plus a weapon part (§The weapon part of the owner's attack power): 7 with no weapons, 433 with a 27.32 dps bow and a two-hander. One level 48 hunter logged 352 at 345 ranged attack power with no weapons, and 1180 at 747 with a 27.3 dps bow.
- Hawk health stays at 5% of creature health. The logs read 6.45% at 48; it does not move DPS.
- The 5% health, the no armor and attack power of its own, and the 30% of the owner's armor are the level 30 readings of #592 (EV-0074).

## Hawk swing

- Count: §Count of the 2026-11-05 to 2026-11-16 reports

Only Unleashed Fury 5/5 and Focused Fire 2/2 hunters, and only reports from 2026-11-12 on: on 2026-11-10 and
2026-11-11 some hunters' hawks swing on the formula before the hotfix and others on the one after it. A normal hit's
attack power is the one on its swing's cast event. The roll is the hit before armor less 2.5/60 of that attack power,
over the creature dps for the hunter's level (ExpectedStat) times 2.5 s.

| Levels | Reports | Swings | Roll share | 95%, swings    | 95%, reports   |
| ------ | ------- | ------ | ---------- | -------------- | -------------- |
| 20-29  | 15      | 3,608  | 0.136      | 0.135 to 0.138 | 0.131 to 0.141 |
| 30-39  | 45      | 21,336 | 0.146      | 0.146 to 0.146 | 0.144 to 0.147 |
| 40-49  | 30      | 11,612 | 0.150      | 0.150 to 0.150 | 0.149 to 0.151 |
| 50-59  | 5       | 2,201  | 0.150      | 0.149 to 0.151 | 0.149 to 0.153 |
| 60     | 24      | 7,352  | 0.153      | 0.152 to 0.153 | 0.152 to 0.155 |

At level 60 the 2nd to 98th percentile of the roll share is 0.117 to 0.191, 24% either side of the mean with the
attack power's noise in it. A free fit over 11,484 normal hits at level 60 gives 0.0416 a point of hawk attack power.

Reports (114):

19KWay63j2HvGdgQ, 1xzPrQWDC4pv9HhJ, 23d6yh4tcqTGNnKR, 2ACDTKkcpWbNfRJ9, 2WLXwQG7kfaBRbxn, 2gKNqpY4G8VdtLMf,
3tLzvVH2AK6aFDMP, 3wdx94CkMQVGmayv, 4WtzXrnhyZpcRVCY, 6KtcxCT4b7ZDvqMr, 7fz9Kk8rY6jyQAbL, 813rPaLtcdpHy67q,
8jzXTaqr732V9DxJ, 9YJM6y2PdRxZarK3, 9YZPHyRaK3FrQkNT, 9gRDVvwAQnC41PBc, AYwLfvqh9QyNj6zJ, AvMqTmDyG3VrZNY2,
AvNwTdyBbDncFf2h, BFPtcmbyLr71Kj63, BmDpWPqCrJfhZxtK, C3r6TRtDnxzVPF9j, CTPkZ8AWYwcXBzNn, D3Qx92znVrkXpvC7,
D9mj8dy41FxGfcV3, FyjZtHCT6zM2WLqb, G2QDqn9gc1mRXByC, Gb4aLkx6KNf98qgD, HL38fnFjNXGTmRCa, Hw4kJRqxZhfvapcd,
J4zpnVTLMaY18tq2, Jjmv6drycNXpDxW8, KLNhgMTDAV6xY1Wj, Kdkm8xan6CFZjvpQ, KfgYa2LbdCRGMhXv, LxfZwpgAyMFcC3NG,
NHLVxbT9hgPZnGwp, NRvfwJbCyQ3qh4dj, NWzbYnVjDaqx3gRC, Nby7FHqDhWcBLYXd, NcRqzah1TQjMr3nF, P2QkCnLwyB1cVA9z,
P6HcTpZXdqAkQBYh, PGVWZxJDC4AMQb8L, PLGKvft3bq8WgRYN, PdZtrjfBJqYMVzCy, QAgrxLyh1CVwcf6k, QF2qpKC9rbxV3ZvW,
Qb1yWxa8vHCGL7Yr, RCGmnkZ9PY1vtQXh, RVgGFynN3KCYJHcA, TA3pLxKFXQ9Ckm41, TNAXzyxRPQZW3pq2, Tfp1v6KhDwJntrRd,
TqZ39QgVkdF6WRDr, TxvwCp4YWqygQ1Pt, V4v6h23bHLaT9Kxj, VK83G7MdBhLRFAfa, VhqB8wZJzRP34cAM, WKFAG7kVZwxbR8QX,
XCam1TRQqwW6pkzf, XDZpkGWVRTfxb6Nn, XHk7W3h1NDbpL4PV, ZYxGWdqtLT7Fjg1n, aHwCztFpYhVXKdNR, aRFzTnCLAtqHNdB8,
bBcVpmN3ARdWD7Jx, bYaT9KXtGwqvV6d4, bavD2gBYZHkqK7Ct, bmZcqx1pW4LGFfKB, cb9gzXjCK3BfH7qM, dFcXZkRP9qnLb7mT,
dvkzFtPH3WR9C4Km, f2zK39XdhnNqZJgR, f36RaB12wyDcTCGY, fD1aRkzx4jYV3drh, fYdRHbPy4nA2k8Mq, gKk9M6qRwByWYPLj,
gPxrZVY4JQ2Fm1Nd, gzrAwTqCWPm7xvpD, h8aVABn6zKPyMwD4, hPNgqjA4vkTxXwG1, hmtTqH16pPGzZJax, j6XFRzHxTC1kvVpy,
jXhdfNz2g7WpBYCa, k4DvNfBJC3wL7MRc, kjRYLXvzgbJyFGKD, m9dYchPtxMFBJ2Gz, mT2F3XZgcnyf7wzA, mak9vPD1KCtHW3n4,
mtXxyZRfgAhLP7rc, n1WgqK3GMVLT7pA6, nFvHyDjrCxLdA6k7, nxpWmkQJXhtPHdF9, q7MvX6Vd42jLrbF8, qGywkYrdHKT8NgAQ,
qTnGKAYkRXHBMDQJ, qVvZxP3gNp12KWyb, qvbXATzmk6Gha1Ny, raBGY1z29kVctF37, rkzFZMhRByHDtT8q, tkHwWYF1C8LTfhQg,
vQZLCzhPNatk7r3d, vfQr1GhC8HjdNZ7L, vk9HJZwrfhmCayMY, vnLtQck7gF9PamD4, vwCXzhMQ2drgpRxD, wJCmzRy4GHgDTA8F,
wPYdtLC7f8Xbc2qy, wp7Wx3jJVmtyn9qY, x8kcqNDnBtG3TYM1, xgCF4fVbKcTmpaN6, yCA1rNbBHntcMqxF, z9XbfMRh21Fg7QjN

## The weapon part of the owner's attack power

- Count: §Count of the 2026-11-05 to 2026-11-16 reports

A hunter's logged attack power is the ranged attack power plus a weapon part. Battle Shout, which raises only melee
attack power, leaves it where it is, and Aspect of the Hawk moves it by its whole bonus. The weapon part is read on the
same hunter's gear snapshots inside one report: when the weapons differ between two snapshots, the change in the logged
attack power on the first cast after each, less the change the snapshots' own stats explain (2 a point of Agility, gear
ranged attack power, Careful Aim's share of Intellect, 2 a level) and with Aspect of the Hawk's state taken from the buff
events, is the weapon part's change. Only swaps between items the sim's database knows, with no level gained between.

| Weapon that changed | Swaps | Change in the weapon part, per 14 x the dps change | 95%, swaps    |
| ------------------- | ----- | -------------------------------------------------- | ------------- |
| Main hand alone     | 17    | 1.01                                               | 0.84 to 1.13  |
| Ranged weapon alone | 12    | 0.21                                               | 0.00 to 0.62  |
| Off hand alone      | 9     | -0.02                                              | -0.05 to 0.01 |

Swaps from a ranged weapon the database does not know (taken as 0 dps) move the weapon part by -10 to +10 for a new
ranged weapon of 10 to 21 dps. Across 19 main-hand swaps of hunters with a hawk out, the hawk's own attack power moves
by 0.97 times 0.8 x 14 x the main hand's dps change.

Reports (72):

1HTPQGZya7pAz3qh, 1aBhkzdDMZKGtb37, 1zMFQx6RYwkgDXTh, 3R69PBjxZvcfbnX8, 3wdx94CkMQVGmayv, 4WtzXrnhyZpcRVCY,
6NRLznh1jgfwqrvb, 7cpTkjaFB3mz6VgM, 7fz9Kk8rY6jyQAbL, 813rPaLtcdpHy67q, 8jzXTaqr732V9DxJ, 9wVj3Kgz6abqNDt1,
AGWhn4CV8Yrjbpf6, AvNwTdyBbDncFf2h, C3r6TRtDnxzVPF9j, DBRqygp13jPC6c82, DLZPJbMwAvTfY4BC, HL38fnFjNXGTmRCa,
JLfYA2drVCZ3aMnv, Jjmv6drycNXpDxW8, JmWKjCcgfXv4GVHP, JnVckZDLTwAgGKaQ, KcdpBWTJyX1xLFfh, MpcVyhTHnqCJvDFx,
NHLVxbT9hgPZnGwp, NXMZV9xjvWpB7GtP, P6HcTpZXdqAkQBYh, PB6zGcVNtTZrMKXb, Pjd9MWAXgCYTcn7Z, QAgrxLyh1CVwcf6k,
QDadybK3XVPRvwLh, QRCgnhck3KVq6vBT, QXFd74zDW8JZMmKx, TA3pLxKFXQ9Ckm41, V4v6h23bHLaT9Kxj, VK83G7MdBhLRFAfa,
WGgVMc3B69TrZwzN, XBbarKyLTcJfNxj3, YH7n2xZqPKzWjQrR, YLNxvdDn96TjA3R8, aMF1JcGxkDpjnYH9, bYaT9KXtGwqvV6d4,
bavD2gBYZHkqK7Ct, bmZcqx1pW4LGFfKB, cb9gzXjCK3BfH7qM, dvkzFtPH3WR9C4Km, f36RaB12wyDcTCGY, fVaTmnXZvPcA7gxy,
fVvLCrnhpykHjQJ4, g3dYGjzy19w2qfXB, gBDMLv8J1K7nqkpd, gPxrZVY4JQ2Fm1Nd, gzrAwTqCWPm7xvpD, mak9vPD1KCtHW3n4,
nFvHyDjrCxLdA6k7, qNVWc2yZJT6k8mvh, qQgj2MWBcD4Gd9an, qVvZxP3gNp12KWyb, qa3hCjJ1w9yzgZYc, rB41vNC2npYbz7Zy,
rfkDTHnR8CBAwhz4, t2W3kaqMxzFPBDjL, tbC9yvQdpY7FnXmJ, tfX7HL9PmDdwkGja, v41rYANj2zBWkb8n, vMn4tpwVxA3Q71dL,
vpFrJRb71fNAatmL, vwCXzhMQ2drgpRxD, yAYRmxqgXkTwr6Bz, z2BJcmAP6frjMV8w, z9XbfMRh21Fg7QjN, zmDY4LVX7N8vh9aR

## Count of the 2026-11-05 to 2026-11-16 reports

- Sources: Warcraft Logs, the reports of the hunters on the Forever servers that each section naming this count lists;
  all fights
- Game build: 1.60.1.70888 for the reports from 2026-11-10 to 2026-11-12 and 1.60.1.70938 for those from 2026-11-13 to
  2026-11-16, as in EV-0131; the build of the reports from 2026-11-05 to 2026-11-09 is not recorded. The hotfix each
  player ran and whether they relogged are not recorded.
- Date: 2026-11-05 to 2026-11-16
- Pull request: #1060

The count is EV-0134 §Method.

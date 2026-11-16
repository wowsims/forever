# EV-0127: Windfury Totem's strike rolls main-hand weapon procs at a white swing's chance

- Claim: a landed Windfury Totem strike rolls the main-hand weapon's chance-on-hit procs, Thrash Blade and Hack and
  Slash included, at the chance a main-hand white swing or melee ability has, and an avoided strike rolls none.
- Sources: public reports on forever.warcraftlogs.com (zone 3000), the fights each wearer had the item in the main hand,
  events with the strike in its own damage line (from about 2026-11-13 18:10 UTC); codes under `## Logs`.
- Game build: 1.60.1.70938 by report date; the strike set runs from 2026-11-13 18:10 UTC across the 2026-11-14 02:00 UTC
  hotfix, and the Hack and Slash set is after it. Hotfix state per player and relogs not recorded.
- Date: 2026-11-13 to 2026-11-16
- Pull request: #1054

## Method

A strike is a damage event of Windfury Totem (10610) from the wearer. Before about 2026-11-13 18:10 UTC the strike
logged as an `extraattacks` event and a plain white swing, which cannot be told from a natural swing, so each wearer's
events up to their last Windfury Totem `extraattacks` are left out. Druids are left out. Seal of Fury and Seal of
Righteousness hits are no candidates: they log in the batch of the weapon hit that carried them.

Each proc goes to one hit by log order:

| Proc logs                     | Procs                                                                         | Credited hit                                |
| ----------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------- |
| before its hit's damage       | extra attacks (Thrash Blade, Flurry Axe, Hack and Slash), auras on the wearer | the first landed hit after it, within 20 ms |
| right after its hit's damage  | debuffs on the target, instant damage                                         | the last landed hit before it, within 2 ms  |
| one server tick after its hit | Ironfoe, Fist of Stone                                                        | the last landed hit 8 to 25 ms before it    |

A proc the strike rolled logs 6 to 20 ms before the strike's damage, and one a white swing rolled 0 to 2 ms before the
swing's damage, so the window is 20 ms; a 5 ms window loses most strike procs. The strike's damage logs about 33 ms
after the hit that triggered it, so that hit's own procs sit 30 to 40 ms before the strike and stay with it.

| Item, procs credited to | 0 ms | 1-2 ms | 6-10 ms | 11-15 ms | 16-20 ms |
| ----------------------- | ---: | -----: | ------: | -------: | -------: |
| Thrash Blade, white     |  62% |    17% |      6% |       7% |       6% |
| Thrash Blade, strike    |  12% |     3% |     38% |      37% |      10% |
| Ravager, white          |  69% |    28% |      1% |       1% |       1% |
| Ravager, strike         |  14% |     5% |     32% |      34% |      16% |

Expected = landed strikes × the item's chance per landed hit, taken per class and weapon setup: white swings for a
single weapon, cast melee abilities for a dual wielder, whose white lines mix both hands.

## Every item whose log order can be read

22 items in 219 reports, 2026-11-11 to 2026-11-15: 2,215 procs credited to 41,765 landed strikes,
2,257 expected at each class's chance and about 121 if strikes never rolled. Paladins left out: 1,828 on 35,626 strikes
against 1,861.5. Avoided strikes: 9 credits in 2,963. A test with no crediting gives the same answer: a white swing or
melee ability with nothing else around it is followed by 1.98 times the procs when a landed strike follows it, and 1.00
times when an avoided strike does. EV-0119's test of swings followed by a strike, rerun with the strike placed 10 to 60
ms after its hit: 674 procs credited to 15,855 strikes against 673.6 expected. The table holds every item with 400 or
more landed strikes.

| Item                        | Proc                 | Landed strikes | Credited | Expected per class | If strikes never rolled |
| --------------------------- | -------------------- | -------------: | -------: | -----------------: | ----------------------: |
| Thrash Blade                | extra attack         |         18,059 |      628 |              621.8 |                    53.1 |
| Bigger Rock on Bigger Stick | debuff on the target |          6,270 |      499 |              551.9 |                    18.4 |
| Ravager                     | aura on the wearer   |          3,295 |      175 |              178.1 |                     9.7 |
| Skullforge Reaver           | aura on the wearer   |          3,145 |      244 |              236.1 |                     9.2 |
| Rock on Stick               | debuff on the target |          3,070 |      177 |              184.0 |                     9.0 |
| Ripsaw                      | damage               |          1,049 |       69 |               80.5 |                     3.1 |
| Stoneslayer                 | damage               |            820 |       89 |               85.7 |                     2.4 |
| The Hand of Antu'sul        | debuff on the target |            851 |       53 |               38.3 |                     2.5 |
| Toxic Revenger              | debuff on the target |            754 |       40 |               38.3 |                     2.2 |
| Fist of Stone               | mana                 |            653 |       44 |               55.7 |                     1.9 |
| Ironfoe                     | extra attack         |            616 |       20 |               14.6 |                     1.8 |
| Rivenspike                  | debuff on the target |            555 |       25 |               28.4 |                     1.6 |
| Stinging Viper              | debuff on the target |            508 |       35 |               37.4 |                     1.5 |
| Argent Avenger              | aura on the wearer   |            453 |       19 |               19.2 |                     1.3 |
| Bite of Serra'kis           | debuff on the target |            409 |       15 |               11.8 |                     1.2 |

Strikes roll on both sides of the 2026-11-14 02:00 UTC hotfix: before it, Thrash Blade 30 credited to 653 strikes and
Fist of Stone 39 to 590. Nightblade, Meteor Shard and The Black Knight, whose damage logs late, are not settled here.

## Thrash Blade by setup

| Class, setup           | Strike              | Avoided strike | White swing            | Cast melee ability     |
| ---------------------- | ------------------- | -------------- | ---------------------- | ---------------------- |
| Warrior, single weapon | 196 / 5,531 (3.54%) | 1 / 518        | 1,129 / 29,782 (3.79%) | 939 / 26,035 (3.61%)   |
| Paladin, single weapon | 79 / 1,681 (4.70%)  | 0 / 109        | 998 / 22,966 (4.35%)   | 172 / 6,051 (2.84%)    |
| Rogue, dual wield      | 235 / 7,361 (3.19%) | 0 / 220        | both hands             | 1,874 / 51,376 (3.65%) |
| Warrior, dual wield    | 102 / 2,942 (3.47%) | 0 / 196        | both hands             | 445 / 23,806 (1.87%)   |
| Hunter, dual wield     | 16 / 544 (2.94%)    | 0 / 23         | both hands             | 107 / 3,739 (2.86%)    |

Single weapon: 275 credited to 7,212 strikes, 282.7 expected at the white chance. Dual wield: 353 to 10,847, 339.1
expected at the cast chance. The warriors' dual-wield cast chance holds one of Whirlwind's two damage lines, which reads
4 of 10,647; without it their cast melee abilities read 439 of 13,060 (3.36%) and their strikes expect 98.9, which puts
dual wield at 353 against 383.0 and Thrash Blade at 628 against 665.7. Up to 24 more procs log before a landed strike
within 20 ms but go to a hit logged between them.

EV-0002 read 30 of 1,003 clean strikes (3.0%) against 106 of 7,884 clean white swings (1.3%). Its script, rerun with
the swings split by the gear at the time, gives two reasons:

- 6,345 of those white swings belong to wearers with an off-hand weapon, whose white lines mix the blade's swings with
  the off hand's, which never roll it. Those read 76 (1.2%); the 1,126 swings with no off-hand weapon read 30 (2.7%),
  and the strikes 11 of 305 (3.6%) for the same wearers.
- A swing counts only when it is the first line in its millisecond, and in a millisecond an avoided line sorts before a
  landed one. A swing whose granted swing was avoided in that millisecond drops out, and each of those swings procced.
  Taking the landed line first gives 131 of 7,947 (1.6%).

The same script on a later scan (36 more reports with a Thrash Blade wearer): strikes of single-weapon wearers 69 of
1,891 (3.6%), their white swings 293 of 7,973 (3.7%), or 318 of 7,998 (4.0%) with the landed line first.

## Hack and Slash by setup

Rogues, events after the 2026-11-14 02:00 UTC hotfix, log order as above; "sword" is a sword or an axe. The expected
count takes each rogue's own chance on cast melee abilities, which are main-hand hits, so it covers talent ranks; it
counts only rogues with 200 or more of them.

| Sword in       | Strike              | Avoided strike | Cast melee ability   | White swings, both hands | Credited / expected on strikes |
| -------------- | ------------------- | -------------- | -------------------- | ------------------------ | ------------------------------ |
| both hands     | 171 / 5,802 (2.95%) | 0 / 300        | 667 / 21,467 (3.11%) | 1,370 / 46,368 (2.95%)   | 164 / 162.8                    |
| main hand only | 189 / 6,774 (2.79%) | 0 / 296        | 834 / 25,577 (3.26%) | 839 / 59,833 (1.40%)     | 187 / 213.2                    |
| off hand only  | 0 / 2,428           | 0 / 72         | 0 / 10,542           | 458 / 23,923 (1.91%)     | -                              |

Dual-wield strikes read about 0.88 of the cast chance in each cut: Hack and Slash with the sword in the main hand only
187 against 213.2, Thrash Blade on these streams 160 against 185.5, and the rogues above 235 against 268.5.

EV-0002 read 60 of 2,208 clean strikes (2.7%) against 327 of 21,185 clean white swings (1.5%). By setup, 11,457 of
those swings belong to rogues with the sword in the main hand only, whose off-hand swings cannot roll it (155, 1.4%),
and 3,801 to rogues with it in the off hand only (62, 1.6%), whose strikes cannot roll it (0 of 424). With the sword in
both hands its script reads strikes 19 of 681 (2.8%) against white swings 110 of 5,927 (1.9%), and 142 of 5,975 (2.4%)
with the landed line first; on the later scan 80 of 2,049 (3.9%) against 471 of 19,767 (2.4%), and 608 of 19,979
(3.0%). Log order reads both at 2.95% for these rogues; what keeps the later scan's clean strikes higher is not known.

## Flurry Axe

Flurry Axe is an off-hand weapon for most wearers, and the strike rolls only main-hand procs. Main-hand wearers with
strikes: one paladin (kpy279HLf1VWgPdx/1-6) with Seal of Fury up, whose seal hits roll the axe themselves. Strikes
with no seal hit in their batch: 0 of 54, where 2.3 were expected at the paladin's white chance without a seal hit (65
of 1,509, 4.3%); that reading has a 10% chance if the strike rolls. Strikes with a seal hit: 17 of 190 (8.9%); white
swings with one: 32 of 140 (22.9%). The report index (5,800 reports to 2026-11-15 14:09 UTC) and 1,000 reports from the
2026-11-16 listing, up to 08:16 UTC, hold no other main-hand wearer with strikes whose gear the logs show, so the
axe's strike chance is not measured.

## Logs

Thrash Blade wearer runs: 1DfQyPFdzBv9ZjAY/1-5, 1z4vRDqGdM2cVPXk/7-8, 2Z4G3r9pfqJtCM6W/7-17, 3KQJHzLcBqVFXCmp/1-12,
3RB9X8HVxgCJ4My2/1-4, 49zJkdp2qBjQRa6M/1-8, 4nkW7JDxV12waMfr/1-74, 6G3gb8DwtJFXm4f1/1-2, 6xHzpKd8wCna37Lb/1-14,
6yHgkT8LqP42NmxY/1-17, 7XK1DyrJ2TfAQ48p/2-3, 9vcQpzdbZmG4kyTP/1-2, A8BnQfxX2maKvFbJ/1, BjAw1m39YGJzWvRa/1-3,
D463FZBX1C9maJ7b/4-8, G7KxnMVrJw19gv2q/1-5, H2T7NDqBgm6LxGYK/1-2, HGNC6QbhXcTqvZ93/1, J38xbrFytgM9p1LT/4-19,
KW4Ytd2VPbHLDJaQ/1-3, LpPyXn6tCJRGMV3d/1-9, LqYCDQ1nh9M3VFrw/1, LyjHDc4nCYmBMRNX/1-2, NQv1MAX6mrxjC2nb/1-31,
NXMZV9xjvWpB7GtP/20-27, NXMZV9xjvWpB7GtP/116-122, PtmYHjkDrqhdR9MA/5-10, QHXdBJRj2GPpC3n1/1-6, QJyfFRYdjbxhnTpg/1,
QRCgnhck3KVq6vBT/1-2, Qm9XJwGdqNzhRc42/1-27, Rcanq4T9FhDg1JMy/1-14, RkFL4TH2rJnx7y8C/1-8, RkYCfyngJrLvwxDc/1-4,
RtZaGw26kvVcAjK8/1-23, TkfYVFdKBth9C3mR/22-50, Va9QYdwAzkphvc7X/1-5, VjBTMHXxFvbthkw2/1-3, VvwJytpX7F9G2Ccd/1-24,
WpgMVC4w2LGy9HQT/1-4, XATkKjnzRMvprfyD/1-12, XvjpYan3qxF9WRgz/4-9, Z4w7HKFPvcrTWd3M/1-5, ZfbY6RgWjx9hmpGz/9-14,
bcQZaKq8rkMgn467/1-12, bzQdaJXx3gqHYWB7/1-9, cD6JqWvBhk19xL3r/1-2, cr6qMtj1vCTPHVfw/1-7, dJa2G7PHZg9wt6QX/1-4,
daT4FcQgCYw2WBq6/1-2, fFd4BgHwTaWXqJkp/1-28, fNMJcqDHb9ZzPLnk/5-10, fjq9P3dHBDCkyL4R/1-2, g6JdFkQKfR7YyP9B/1-3,
gPnfcxhGJARTwatz/1-26, gmXwaKBfrC7V3jFL/1-40, gtNvbhF9aZCXyzVk/1-6, h3gQBtwb9ypMzLN4/1-7, hH4YRGVFz3KnN6bw/1-8,
hYGBZtM6f9QcDWJy/1-9, mxbK4d1k96GzwVRh/1-3, nfLmGxNDqJpyXB6k/38, nrmvCfDxwagQXTAb/1-2, ntaAc893bkmBd1Cg/1-21,
pb7dQrAw8CGWjMk4/1-5, v4BCMay28n1fHpkq/1-3, vTrR93LjF4kW8Hg7/1-2, vZqVmLYQa6jwP3A9/1-21, vqQgaxHKDw6kh4PX/1-5,
z6LMfrpbwhyVCAJK/1-21, zmDY4LVX7N8vh9aR/1-3.

The other items' 219 reports: 123jYCvkVGrHfXWt, 1DfQyPFdzBv9ZjAY, 1PcrGTyatDAFx8Yk, 1z4vRDqGdM2cVPXk, 2FkYRHB1dWhra8vt, 2KkaByn6H9mzb4Yq, 2VAWrb67G8KMzgj9, 2Z4G3r9pfqJtCM6W, 2wMznDRvmkpbtHxA, 3DZdnycJp9W4b6gC, 3KQJHzLcBqVFXCmp, 3RB9X8HVxgCJ4My2, 49dLXZ3g7NhzqABb, 49zJkdp2qBjQRa6M, 4NYV7znpwqtKM3mc, 4RcnAhpLZ96KfNDP, 4nkW7JDxV12waMfr, 6G3gb8DwtJFXm4f1, 6xHzpKd8wCna37Lb, 6xz4k73CtjgrZATF, 6yHgkT8LqP42NmxY, 7XK1DyrJ2TfAQ48p, 7dWytj1K24G6ckJ9, 86xt9ykqKzAC4hMX, 8cVPaCpAMmv93bkD, 94Y7DJ2wCPRxXBqM, 9gjX3ZtpyNw2qfA7, 9vcQpzdbZmG4kyTP, A8BnQfxX2maKvFbJ, BDFj9VNrfXhxwQT3, BLXHnFRfrhPGYa76, BjAw1m39YGJzWvRa, C4TGJvzHF3aZ2NjR, C64x3BnYWMQZkA8T, CRgvLNQYd8y3c2m6, CfHvwytKJ6P4q7ag, CjLKJRNhnb7TBzAZ, Cqkfvy2Lrdtc47zJ, D463FZBX1C9maJ7b, DWfMg7hzPF4d1mQ6, DjqmtPNvZCTKrBzx, Dx3HYL7cwt1GCMbq, FMaAjHcwCQLkpV32, FPfGz7bjkvmyWXTB, FtcXATwrHDd9vQ83, Fx94JmkLncTXBPQy, G7KxnMVrJw19gv2q, G9PLgJ8rWcCnkmKF, GTJw7YXampLrWcgM, GcJqNaV3MBfRC7Yp, GjmZfc8aLPFQVXKN, GpnMqtzaQVCWjyr2, GrcTMNyY8p2f6xV9, GyZmK2b8c9LMtDdw, H1CMQLTxXbKyBAWw, H2T7NDqBgm6LxGYK, HGNC6QbhXcTqvZ93, J38xbrFytgM9p1LT, J9B3mTqZ6KPNjAFY, JHrkTPq23LpAcGxY, Jt6431MZTKmqQLYF, JxVKYdTaBRZ9GztX, K1ybYhmj47pWfG3D, KFLjdptc3gqCGByx, KW4Ytd2VPbHLDJaQ, LNkzZDmPG9CKw8vV, Lg82ZFDyQ4MvRm9p, LpPyXn6tCJRGMV3d, Lpr76Vd3DbKWZqnh, LqYCDQ1nh9M3VFrw, LyjHDc4nCYmBMRNX, M9y1GJVwW7j4ZAXg, MC2GmHchpfVkNqwD, MyRFjDzr1Kn4b2d9, N3Y1mFkJ9Dwr74bA, NQv1MAX6mrxjC2nb, NXMZV9xjvWpB7GtP, NaDp8bGmJZgKf63y, PCW1JNXVx2B8nDyh, PLXrhm98NKCnk3Tw, PpLZ7FAgkvTwc1Cr, PtmYHjkDrqhdR9MA, QHXdBJRj2GPpC3n1, QHjqK4G3pzgRWaN6, QJyfFRYdjbxhnTpg, QRCgnhck3KVq6vBT, QajbfyAcJdRCkP92, Qm9XJwGdqNzhRc42, QxRZFwrkKgJPdB82, RNCXFn4H2hgqAkp6, Rcanq4T9FhDg1JMy, RkFL4TH2rJnx7y8C, RkYCfyngJrLvwxDc, RmTM8XL4aBJY12Ak, RtZaGw26kvVcAjK8, Rxd8h1LgQp2FTB6j, TBm9b8kyPANYFvjK, TF6dc3NCHbKrfxJt, TFYgK8vGyQc1MzrL, TJ9hP23LykXNDfn8, TK6QARgFMk8Byf4V, TkfYVFdKBth9C3mR, V39PRKTHt8kahfxX, VJwzgFmrtN8dRQGf, Va9QYdwAzkphvc7X, VjBTMHXxFvbthkw2, VtGAb9HNFMYhRjXq, VvwJytpX7F9G2Ccd, W6GLfFa4NCbXqBJh, W8nBLvF4jyGtazDg, WLKn4vH3VrxfF1Pg, Wb79QjX8wN4GkgrV, WbfQZpaPAkL8tzdv, WpgMVC4w2LGy9HQT, XAD1zkMGQFgtjPTy, XATkKjnzRMvprfyD, XfQ8GLDCFN6KyTvr, XpnDCMx9BtbWa1FG, XvjpYan3qxF9WRgz, XyPY7tNcdb1fFDCZ, Y6bcWnJv9kFPyXmG, Y7ZntCL98FcmXRNd, YAVgyZp9FNMKCh4D, YXgt6fkMdBTKDrAn, Z4w7HKFPvcrTWd3M, ZH7zNkqT9j3G1Qm4, ZfbY6RgWjx9hmpGz, a2ZkBR4b6HcFf7dW, amTgyKJbBwWcxZ69, avhktV7dAyZPzDnb, b2YKZcFzWPtkxfpg, bK19VvyTzrY4aj7L, bcQZaKq8rkMgn467, bzQdaJXx3gqHYWB7, bzwY28XH6Vfc4Zpg, c6DWA1CV2tdmRFqL, cD6JqWvBhk19xL3r, cD6XhzYbjHQFgpG1, cdFkXVKH1xm2JhpW, cmnDzYT96jLbRyPV, cr6qMtj1vCTPHVfw, d3bDt8AVT1caMYWn, d6TMB1P784WZHVhF, dJa2G7PHZg9wt6QX, dYzgZkCNJFQAXtBP, daT4FcQgCYw2WBq6, dpVr9znc2yCBHRvX, fFd4BgHwTaWXqJkp, fLyrVbpZAWxzQFYk, fNMJcqDHb9ZzPLnk, fWgMdvDnKGRxAt4T, fXyK8kCNdGxtQ67c, fgw1ckK8HTAvQ3pq, fjq9P3dHBDCkyL4R, fn2Dc941K63CNvmj, g6JdFkQKfR7YyP9B, gBDMLv8J1K7nqkpd, gPnfcxhGJARTwatz, gapAZfyCTYnmQkKF, gmXwaKBfrC7V3jFL, gn7Wva1FKAm8HhGZ, grjxzfTdLQKbnBwZ, gtNvbhF9aZCXyzVk, gwdtc8fFVQaKGjq1, h3gQBtwb9ypMzLN4, hH4YRGVFz3KnN6bw, hYGBZtM6f9QcDWJy, hf1pF42aPHMv9kLX, hvdx4tJf9QnzgrHb, j2FY7nBLCcakV9TK, jzmMkJpRG2X1HNxr, k13frPZGVLQMtCzh, k26rdYnGDNZ3b81P, kDjcaHqQvWpXryKn, kK1rBNaf9Yzq7Cbm, kQZqRBDMaNbX2jdV, kYrgAKnb6GNpvcTh, kpy279HLf1VWgPdx, m3Fjgrt124A7advV, mRzw46vjLTHnXQPx, mxbK4d1k96GzwVRh, nfLmGxNDqJpyXB6k, ngAGhLYavywCM2F3, nrmvCfDxwagQXTAb, ntaAc893bkmBd1Cg, nvXGCNxKYDHrPgjc, pK7H2P8mqjyJ34QD, pb7dQrAw8CGWjMk4, pgPZv2kcYVyA87fW, qKzDTJjwvXM264Q7, qP2MRFvJ16NBZGgm, rXKd6hBGZkvYFyTc, rtQwP64YL7GqR1Hf, t9hT46QXNwaPq1YG, tYamjwLT8B4k12Fh, tpWza62JNFrTm9hH, v4BCMay28n1fHpkq, v6hC8nkRyQjbgJHX, v8DRwt2gJBXGxcVN, vTMxmyJYbZ6dVkf4, vTrR93LjF4kW8Hg7, vZqVmLYQa6jwP3A9, vqQgaxHKDw6kh4PX, vqZx76MDQLbwpg3m, vrX8N4PRwHFqgJYM, w2PLVMzXTNvB1pAW, wp7Wx3jJVmtyn9qY, x268nX3VfMqAvwtN, xCMngy9JrNtkGFVp, xQ6W8rHmCn7AcTwz, xVLgCnjDvAkwJmhW, xqnmW3BDpFb28cJ1, yDbnf2Y6MQWm7KNp, ydbJ42ZkTaHnN3Br, z6LMfrpbwhyVCAJK, zH3TmcCQRb4Pa8tk, zM2hwbVmgyJHRPFt, zmDY4LVX7N8vh9aR, zvkWdnQb3t7yCmKZ.

Hack and Slash and EV-0002's rerun, events after 2026-11-14 02:00 UTC: 123jYCvkVGrHfXWt, 369Yv2NcqrygTpwX, 3WgzAhrRXZGktLbd, 3hX6wHt7WLCrjq8R, 4DmjXGdbBQTKP6zr, 4nkW7JDxV12waMfr, 6h17av3QGCFAj4BR, 6jqwbMpRFxTmHKkX, 73bMfN9AtZDkjQpF, 7f6DBRAPMLGYzyQm, 8G6CH9MqtxyczFhA, 8X2Kc46faTnxVBLm, 8dT1VNMQfLgrj9vc, 9XQfBmZRaNWJkH3F, 9nJGwhXTd8pvraxb, A8TBJgd39akVwMmz, ACxnFt32myYd19Gw, AWj6g3GbV9HTaPr7, AvNwTdyBbDncFf2h, BQZcK6CPmMLNqf8a, BQnXb3hTC4rmJ2pd, CDB7JAfPVW1Rn8Kz, D6W2wLrzyKp7MCcj, DBRqygp13jPC6c82, FRjMJvX6NbrCPZLK, FTXVdBrHJZD9aw7P, FbNk9QwxY7nG618P, FtcXATwrHDd9vQ83, Fwr9cPXGt4Nqym6f, FxMADVJakv76RBmC, H7f28bdvhrFpwVBM, HjrkAa4tR9Pp76nK, Ht6wjv31akZDbgWy, HvD8G6rkFgpKXyML, Hz2RfmrYDh8LBcPa, J2XL6AyZqgpjx8QK, JCpQhr8ZnHXkc97L, JPyp4NdnFBtM6qCL, Jf1aWNHgyYktjGXM, JxvWA68g3jTkFdQV, KLNhgMTDAV6xY1Wj, KPkDVWCvjLw2BqNR, KwNTBcG3ypfqhj1r, KyhdLNqFDbA496vw, LRxP6QhvGFgdTr7Y, LmDb8njHfXvAgC79, Lxy1kqgVvFc24aPr, MC2GmHchpfVkNqwD, MV3AfgqjPFXGzB1y, MgmdJqxBWA3Tw2vz, MvgK14QZDxjWGBXp, NaDp8bGmJZgKf63y, Pd3LqCXcxMg92kA6, PxfYmBXLqFg3GMvT, QJwXLaVWzx9pMqky, QNwkdxbMvWmrqjAY, QtHnFxTVRCLPWqgj, VYBjgJx8aZtRhmpy, Vh3bzdRMP6ymNKq9, W3nLgjDvbPz46BdR, XDAL64ahHPZqdWTM, XNry2TxRndjwJQBm, XWfw2PzpNcBhbayY, Xv2QmnyK8RBxW9jZ, Xwp1ybvGjJHVadkZ, Z182HvNhFzdqgCrj, ZWX4ChPY7pa1GnDf, ZpgM4C1DbqVQ3WHX, bavD2gBYZHkqK7Ct, bnfTwVz6BZvPxd1G, f43Q2HLNTj8CxbXZ, fXVLPGmhMj27xnaJ, fXyK8kCNdGxtQ67c, gAm7d3xLQYC1RGFp, gBDMLv8J1K7nqkpd, hJHTtV3YMK28NAPv, jhmMAGq7aRn3F8JP, mJDNgqYVGwb6A8tF, mg3MBw6rcNtZpjbA, nDNT9zhcQ8trJbwf, nZPQBaFvkqC9btAJ, nfH1DVpR3mvgjhW4, pKn8wBZcQ62PyJH7, pTtwfyhVQAN6v1X9, pjgYTq7WMcm6vJRf, pvWfktJnRm8VDBhP, rAFM4QTbC8jHypZL, rYJADWNn9RdbzMVL, tA13MXZ8pLDjqcxG, tRafp9h7wkdKcD1m, trk1cxf24WwHy7Cn, tzRdk7fHGKpx61vJ, v8aCfPGQpZ61Ytx7, w8DFfTr1ZV6dQGYA, wJtPdpxFGQCM9f3c, x3njCmpkVYf7WDgh, xhynMHmp1RQcaAgb, xqBjXaQGTZk9hV8v, xqQNadDY9JRwG1hL, xqmMzTC4dDwGN8Wk, xypHFRD4acndmhqJ, ycdL4FWKHQnjbpYq, z9XbfMRh21Fg7QjN, zA1GprYM7fPaKFH2, zNWxq9r8mPkTGMaB, zm9cFxXADgfht8dL.

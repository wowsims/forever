# EV-0169: Violet Sorcerer's Vestments roll 1 PPM on a spell's cast time, per target

- Claim: The Violet Sorcerer's Vestments 4-piece bonus (1324318, Violet Meditation 1324315) procs at 1.0 PPM. Each
  spell landing rolls a chance of 1.0 × its cast time / 60, instants priced at 1.5 s, and an area spell rolls once per
  target hit. The 5-piece bonus (1324323, Summon Arcane Elemental 1324322) rolls the same way at 0.95 PPM, after its 15
  s cooldown. Both are the lone-unit rates; the rows' base of 1 is that PPM. The elemental only melees, a swing every
  2.0 s for 35 to 43 damage before armor.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the 270 reports under Reports, every fight with 4 or
  more pieces worn in, and for the 5-piece bonus the 146 of them with fights where all 5 are worn. Runs logged in two
  reports count once.
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-10 to 2026-11-20
- Pull request: #1146

## Method

Wearers: 223 players (272 runs: 177 mage, 47 warlock, 34 priest, 12 druid, 2 shaman), gear checked per fight from
CombatantInfo. One paladin is left out. Only events the wearer is the source of count: the elemental, other pets and
totems are left out. A run logged in several reports counts once: two runs of one player whose damage events share
over 30% of their seconds are one run. A Violet Meditation proc is an applybuff or refreshbuff of 1324315 (no two log together). A
summon is a summon event of 1324322. Both log 0 to 40 ms after the spell landing that rolled them, so a landing gets
credit for a proc 20 ms before to 60 ms after it.

- Units: a spell landing (damage that is not a tick, or a debuff applied), with every landing of one spell within 5 ms
  joined into one unit and its targets counted. A unit is lone when no other hostile cast, landing or debuff of the
  wearer comes within 400 ms. Its cast time is the time from begincast to cast that the log shows, haste included, and
  0 for an instant. Whether the game prices a spell on its base or its hasted cast time cannot be told from these logs.
- Fit: P(proc) = 1 − (1 − PPM × max(cast time, floor) / 60)^targets, set against 1 − (1 − c)^targets with one flat
  chance c, by likelihood, with a 1,000-draw bootstrap over wearers.
- Left out of the fit, as triggers whose roll is not a landing: wand Shoot, channelled area ticks (Blizzard, Rain of
  Fire, Hellfire, Hurricane), Arcane Missiles, Mind Flay, Drain Life, Drain Soul, Ignite, Chilled and Melee.
- All hits: procs over the summed max(cast time, 1.5 s) / 60 of every target hit by a spell in the fit. Procs from 60
  ms before to 20 ms after a trigger outside it (the casts of Blizzard, Rain of Fire, Arcane Missiles, Mind Flay, Drain
  Life, Counterspell, Polymorph, and any landing with no cast time) are dropped.
- The 5-piece bonus counts only units and hits more than 15 s after the wearer's last summon.

## 4-piece: Violet Meditation

| Count                         | Units / hits | Procs | Wearers | PPM [95%]            |
| ----------------------------- | -----------: | ----: | ------: | -------------------- |
| Lone units, instants at 1.5 s |       34,999 | 2,198 |     214 | 1.009 [0.968, 1.052] |
| All hits                      |      115,783 | 3,917 |     223 | 1.092 [0.994, 1.201] |

The two counts overlap. The rule for weapon procs would post the all-hits rate rounded, 1.10; the lone-unit rate, 1.01,
matches the row's base of 1 and has the narrower interval, and the sim takes 1.0.

A flat chance (3.08% per target) fits the lone units worse: the log-likelihood is 76.2 lower (bootstrap 95% from 41.6
to 116.1). Pricing instants at 1.0, 1.75 or 2.0 s gives 1.247, 0.917 or 0.839 PPM, with the log-likelihood 58.5, 1.1
or 12.2 below 1.5 s.

| Single-target landing, cast time | Units | Procs | Chance | PPM at max(cast, 1.5 s) |
| -------------------------------- | ----: | ----: | -----: | ----------------------: |
| instant                          | 4,016 |   119 |  2.96% |                    1.19 |
| 1.5 s                            | 2,398 |    65 |  2.71% |                    1.08 |
| 2.0 s                            |   858 |    33 |  3.85% |                    1.15 |
| 2.5 s                            | 8,460 |   355 |  4.20% |                    1.01 |
| 3.0 s                            | 2,849 |   139 |  4.88% |                    0.98 |
| 3.5 s                            | 1,802 |   100 |  5.55% |                    0.95 |

| Instant area spell, targets hit | Units | Procs | Any proc | Per target |
| ------------------------------- | ----: | ----: | -------: | ---------: |
| 1                               | 3,964 |   119 |    3.00% |      3.00% |
| 2                               | 2,999 |   150 |    5.00% |      2.53% |
| 3                               | 4,058 |   310 |    7.64% |      2.61% |
| 4                               | 2,639 |   273 |   10.34% |      2.69% |
| 5                               | 1,558 |   179 |   11.49% |      2.41% |
| 6 or more                       | 1,641 |   270 |   16.45% |      2.95% |

Triggers outside the fit, on lone units: each Arcane Missiles missile 297 of 8,296 (3.58%); the cast of Blizzard 391
of 3,828 (10.2%) and of Rain of Fire 130 of 1,521 (8.6%), while their damage ticks proc almost nothing (Blizzard 10 of
23,842 units, Rain of Fire 2 of 4,783); Mind Flay and Drain Life casts 2.4% and 2.5%; Counterspell 2.9% and Polymorph
3.0% of casts; Hellfire 0 of 465 casts and 0 of 2,757 ticks; Hurricane ticks 0 of 633; wand Shoot 1 of 14,865.

## 5-piece: Summon Arcane Elemental

146 reports, 125 players (147 runs: 97 mage, 25 warlock, 14 priest, 9 druid, 2 shaman), 2,722 summons in fights with 5
pieces. The shortest gap between two summons is 15,005 ms, and gaps from 15 to 20 s are as common as later ones, so the
15 s cooldown holds and a living elemental does not block a new one. No run shows a stretch with 3 or more summons due
and none seen. 33 summons fall in fights whose CombatantInfo shows only 4 pieces.

| Count                                        | Units / hits | Procs | Wearers | PPM [95%]            |
| -------------------------------------------- | -----------: | ----: | ------: | -------------------- |
| Lone units, isolated 400 ms, credit to 60 ms |       16,407 |   941 |     121 | 0.940 [0.868, 1.019] |
| Lone units, isolated 1 s, credit to 900 ms   |       12,980 |   698 |     119 | 0.950 [0.880, 1.030] |
| All hits                                     |       51,854 | 1,652 |     125 | 1.011 [0.884, 1.173] |

A flat chance fits the lone units worse (log-likelihood 30.7 lower, bootstrap 95% from 14.1 to 49.8). Single-target
landings read 2.61% for instants, 3.74% at 2.5 s and 4.69% at 3.0 s.

The Arcane Elemental, from its own events in 146 reports (2,722 summons): it casts nothing but melee. It swings every
2.0 s (median 2,004 ms) and lasts about 20 s (median last swing 18.3 s after the summon, 90th percentile 20.0 s), for 8
swings a summon (median; up to 11). Over 19,327 swings: 83.2% hits, 7.3% glancing, 4.1% misses, 3.2% dodged, 0.9%
parried, 0.7% blocked, 0.5% crits. A summon deals 195 damage (median).

Its damage before armor is each hit's `unmitigatedAmount`, over 16,124 normal hits:

| Grouping                                       | Groups | Lowest hit | Highest hit | Median hit | Highest / lowest |
| ---------------------------------------------- | -----: | ---------: | ----------: | ---------: | ---------------: |
| One elemental's hits in one fight, all summons |    296 |         35 |          43 |       39.5 |             1.19 |
| One summon, 3 or more normal hits              |  2,310 |         38 |          43 |       40.5 |             1.14 |

Each cell is the median over the groups. The sim's swing is 35 to 43, the per-fight row. Across groups the median hit
runs from 34 (10th percentile) to 58 or 59 (90th), per fight or per summon. The wearers' gear is level 27 to 30
throughout, so this spread is probably the owner's level and not the target's armor.

The logged damage is 0.71 of `unmitigatedAmount` (median), which is the target's armor: a normal hit logs 30 (median,
23 to 39 between the 10th and 90th percentiles) and a crit 61. A crit's `unmitigatedAmount` is 40 (median, 97 crits),
the same as a normal hit's, so the value is taken before the crit multiplier, which the sim applies itself.

## Reports

Every report with 4 or more pieces worn; the 146 marked * also hold fights with all 5, which the 5-piece count
reads.

1Fh4d67NarWQqCK8, 1NRQ9tBnTrq24D3k*, 1PbdHz436qCvkTAX, 1cCAZtmgVx6qdyM2, 1n9BvKXPpNfDGgVc*, 1rH3FkMXK47ypwBY*,
291tprnTRC3aVPzW*, 2BHVPMfv94cKNqY3*, 2RV8ZMJfHvW4hgP6, 2XKwjadtZQN6khRf, 2c19k4Q7nGYHb3Nm*, 39NYyDpaM6LFX7nv*,
3PDH9BbkMvYZQFAW, 3PXfhvmTp6NQdwRb*, 3YqRjw1ncKCFax8B, 3pNML2gGPXkb6mtf, 48BhtMR2TJYLkFpz, 48kNQjHFdP2xMCnz,
48qzyJ2cQvt1mkdp*, 4RH82jaf63yDQ7MX, 4pynYMkBdVgRDvKH, 4wPxkWC2tznKD6mG*, 61rRkP3g7tbWZ42B*, 63kLWDCqMdRghNZY*,
6WkrKt3RwHGmb4Zf*, 6ZDVxnWtQ81cmdky*, 76RDdfF4zVJwrLZp, 7Rq3bLdfPn8g9QF1, 7dZt6b9Kh4F2CNAn*, 7n82gx49BrD1qwPF*,
87DQtHwPNfVTBXh4, 8DG4jcMNbvRyfhk7*, 8HNW6RnLjm79Y2Av*, 8Nz7kVYMWXJbPGja, 8VQDf7JtwbTknMLz, 8YzmGj6fAZqMBNtx*,
8ghGqzDZvCN2rW1j*, 8w9taHCVG4mFNQ1c, 8xab2dkt1jPXfhNQ*, 8y1t7dTbY2JHMpVj*, 931c4TwCkJNzFZrq*, 9DCzLFVM8NtdQ6pK*,
9nxtzJ1bC42FrgWN*, A6dpT9jYahH7tm3N*, B4ak2F9hrztgQJH6*, B9TF2RJdrC34xpXW*, BMw8FLrdn46AzZQC, BZvXM7FA2nyQLVqN,
C8LFym4xnwKadQJ9, CF1GJR8xjAzb3Zdc, CLzKxRAnaH7Z3v9G, CQfJXhtLZrTmFYj6*, CVhvramqdMpHY6QR*, CZ2jNc8pq6yYT3vg,
CxpqN24FHadRj8ct, D4JKtv8PX6wGHzp9, DGmh6FQBya4KxvjX*, DVwbPz4Zgq6p9yK3*, DaH6LCWQbmjXAJFh, DcdNQZC19wanmrjq,
DdRz8jwpVr396WyF, DyC3KYB7wAZHz4f6, F6KYtxXhp1VHRbd2, F8t73ajQPWMNXwdx*, F93PjaRL2BNKMcnG, F9a7GtdqmXzLCBWQ*,
FJt1fT3rLNyBPKmW, FMaAjHcwCQLkpV32*, FWPkHAtQCLGfqKpx*, G428r7YFXQjPVwcp*, GRF4brLcWktdQ9zy*, GZ6FBtjMw4cTbpDq*,
GaKdbJPj2wnkZhgQ*, GxZ6hXCcW8pafg2D*, H6jVzQtL9X1kJAZm, H8FTwc7GVzLQBCdh*, HT3vr4zB8kWw1AKM, HbtNwgyMJWPjpzhk*,
J2XL6AyZqgpjx8QK*, J4Cbw23F6KQPTRpq, J8hmPTFWGNzwRacM*, JBPr6fjCmRW8Xcpx, JGgDXzZk4dRFQhPN*, JKavrBGqLD6hVcHZ,
JLtv24HKRrPZjQ9G, Jp6KnNvBgyarLqwc*, KAfmrqFHXh7TzdL1*, KFLjdptc3gqCGByx, KNdhWJ6rTvxwtHB8*, KWb3R7HLng2DvJMp*,
KmWZxNk3cPFqzgRG, LD8cqra9kRtAN7WF*, LFtHzc94KfrhJ2vX, LKFWxVqpwfY8JyPM*, LMdRqJGhV17XwWyT*, LNPk712GxaTqKXZf*,
LQbJn6gcWz2D3HCT, LWPJhnKrAqtDgxkc*, Ln2w8ZDmR34KrYjg*, LtPf3jzMg2aZ4mWh*, M7VKfPYJ2jqcX8mC*, M8NVkXhtqPb69jAx,
MKk2adJtjyANf8LW, Mh2rARNkqymtG8Cc, MhPVnKBGqLXZ7N9H*, MtxT78RD4NkCJd1q, Mv6K4mBHN18WDpLb, NaDp8bGmJZgKf63y,
NmnLGRq3Tcpx972Q, P6qAmjBhrNa8b9MR, P7pHb2VkAdhT1Ggx, PLXrhm98NKCnk3Tw, PtzhwaTHXjdnqRm9*, Q1arfj4WPBNXqmLk*,
QCKMXajfN3wDcHzv, QDfmd1xbJpPkcFyC, QMRVxpm8rhAKcwk9, QVpNH3WvmhYbaFwy*, QYK67bGdw9Cz8fva, QYwC28JgBkmjaKrx*,
QZ7vgK6xtcb1LkCX*, QmwTPYBpVXt6z1fd, QyHRbdM3YzXvDwL9, R2kbrZ4NWwAqLMtT*, RPcZmtp8H2bwzKgY*, RQjfLHpktThvGVc7*,
RdmgVNQ4XJMfnC1F*, Rqpv8MyhzrwA2FN4, T17F3ya8cRYtmWBC, TBA9wg8R1z7XWFMJ, TdNkDGLxM6BQvCyh, Tf7FkLgZcpQ1a4AB*,
TmZ1DwGj4C9KV7Md*, TxwAyD48VPab9CQv, V1b6xcMyJ4ngYrBq*, V2PHgjY31ZNMAvd8*, V8CGvRftMWzFLXAN*, VDAtRmdp7vHQLyX1*,
VYawMfnr36HyGChq*, VgvHznLqA3PhXQjN*, VyXwH1Wfq2bc3xZ6*, W7VPQB18nHfkMaAc*, W8nBLvF4jyGtazDg*, WM1kTRnXVr7hBGKL,
WQhG6P2JL38FfAKq, WbZ3FnBqc71vGzRm, XG2JqKDjdacwNkQT*, XPzk1aT4pCFVtd8N, XYtxKrB7dWJPHa2G*, Xcjfm3ndv6BxWqz1*,
Xf1FQLdn6yAmKcMP*, Xk7rhfCt1dy2QA8P*, Xvczhd3HytKV92TM, XzQgt2Tq1YyCvBFn, Y8vzmt37cRbXVaNW*, YAbvjTrfZD728PM9,
YJR78AMPDNKGQWjF*, YP7R3K9GLVNvfrFT*, YjA14kP6FyQwLnNJ, Z7wfcJWjLmD9pHMg*, a1knxd78DpAGMbmX, a2M3gKWpwq7PnDc8*,
a3xXpct4Pv7CmHZj*, aXrMN8hPypRZB6Tg, agDVFTwcYZHnRd1C*, avndmD69zTjCHBwW, b2YKZcFzWPtkxfpg*, b7PCQ1A8z36BDrTx,
bNpCk1gG7y3LYrdP*, bT3LyRBtax8FQnch*, bZ1rRapKh7JcNy9V, c8bZxV7qLjKQFMHf, cCBghykdHNAY8QFG*, cgGyVjBHkPa1F3tN,
dDvzgp37F9VLBMkY*, dKFcyJ72vAYpkqHr*, dLXJ21r9jzAmtgf6*, dZrWKyBanQXbvTJ8, dnXbDqMcL3pT1mr6*, f4QJBt7N86x2bkjZ*,
fA1GLbc6gw9nKHM2*, fMkvA2btVy8WwjhC, fWRgvP17z8yrL36h*, fkRMFgKNQ2zWXZxP*, fmQnCP3kFW6AqL9Z, g9ZNnfPkdHG2bKRt,
g9rJBqRa7F1b3nxK*, gALWqjdyncf8xVPa*, gh1AFZdcqL4MPbQf*, gn4HAC1RqMfLt8wN*, gx8qb9JmBrW2KDhw*, h1q4pMzRFvcBytfH*,
hDRnH69d1LbQN4VJ, hMdWYxyPG3rn9VQ7, hYHBQ9ZJX3vFAnkj, hfBDjZxwkYCV29z8, hn69TbWrRmBd2KJN*, hv3mAnKNRZwypX7M*,
jBvJX1Ltky2hz7xm*, jDA37XdHbaywYq1M, jL6pgnkdMfChKwPq, jRvABFgTGYx47cfP, jYVvXCw4QANH1Ffn*, kVd6ApJarwLRZQPf*,
khaLCvcjTZ4R6A8W*, ky2zDj3JCp48M6Fd*, m6F2WvMAJzQHa9nL*, mBavwVKf8rLgH3CJ, mPGTVFdZX8RCbkgY, mYRPrb2VLwzcQJ4N,
mrFfq91T3wbXKH8D*, n1LQ6r2NCZDMGfjR*, n7913jW8dcfLvYP4*, nRJgDW7q4khr3wNC*, nTpr2qjNFygzZDa8, ncQ4d2H1NRgAbBxh,
p7QgHbGqKC2B6ryz*, p8FzkTmbqc2KDyW7*, pQ3Nba6vBAjxhMPg*, pfxzGqDBYgWKnF73, phMXPfRgx817F2Dt, pmKDwngMT2NFktdA,
px3rGD8JbH7YTdLV, pxRc4rDyWFXtNaVP, qW4HRXA31pNm6jZg*, qcL6h7XmHFTRVPvK*, qn1xQW9kjKCT3GMf*, qnPKLcfBYtR3TFDQ*,
qwp68a2FxvzCrGZc, r2DZgwHbCkKcNFPt, r38LkvhBFVH4QqfD*, r8n6tLymRbzx7hjw, rK8mBk7ZHTqDnbFN*, rRCjG4FnyBZa2wTz,
rc6DJw7FGkjRKMxd*, rhXV28QZbzvynFLd, rmTK1jhgtaV7B6np*, rnKvDNWYagZLXA24, t4QGjmbWFxKPRLA9, tFJ1HRZnvy8DhpkG*,
tTx8QaYmB2J6NR3g, tYpAPcQ6HyrNdg2k*, tfKXTRgWAPMZvkdJ*, tgXLJCypHzh8cZd6*, thpcGknfvy49gTz8*, txn8Q1X2CLpcbwyA,
v34xDNZRfyCPrKBJ, v6Ljcx9QBTZNakgt, vBnYM4rZGP8TpxJk*, vLGwZxXBaVcpMT7H*, vLTrfbYJzHXKkpt1*, vTMxmyJYbZ6dVkf4*,
vwCYLZdVtGHJ4jTc, wCRkQYvLABp1XZ8b, wDyPtkh3FzT8LpNn, wKanqBZGFzhWrAT7, wXYW7VL6hQnqxk3C, wdJqg9Z6nRaYC8pT,
wr3tCAx6cF7YkmKG*, xdvQ6hbRDmTVaH78, xnrg9m6jDMR1kaLN, y7qfkc91hA3DKQ4g*, y8HvMVq7QdBTatGm, yDFdKBrtAnxcHLq4,
yGj8MbJ7YrxnZmWq*, yaYtkz3mNhTWwxAJ, ym6CzKL4B1hnYaJQ, ymAPWzKdB6Zkv24f*, yrLT3aAxhf7NYGWF*, zAhyL8K49vgpGJB1

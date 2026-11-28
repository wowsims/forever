# EV-0074: A hawk's level 30 stats, and its crit, hit and expertise from the tamed pet's

- Claim: at level 30 a hawk has 5% of the creature health and no armor or attack power of its own, and takes 80% of the
  owner's ranged attack power and 30% of the owner's armor; its crit is the tamed pet's, Ferocity included, and its
  melee hit is the tamed pet's.
- Sources: four level 30 combat logs (2862 hawk swings, owner ranged attack power 407 to 1023); report codes not
  recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #592, whose body is empty; the readings are in its commits 81651da664 and b89878e6f9
- Later count: the crit, §Count of the 2026-11-25 to 2026-11-26 reports

## Level 30

Read off 2862 hawk swings in four level 30 logs, a hawk has 5% of the creature health, no armor or attack
power of its own, 80% of the owner's ranged attack power and 30% of the owner's armor. A swing before armor was 3.3 to
6.7 plus 0.0179 a point of the owner's ranged attack power, the same line in every log.

## Crit

- Count: §Count of the 2026-11-25 to 2026-11-26 reports

## Hit

1% hit took a hawk from 1 miss in 32 swings to none in 515 against a target 8 levels below it.

The code takes the tamed pet's expertise along with its hit; the commits record no count for that part.

## Count of the 2026-11-25 to 2026-11-26 reports

- Sources: public reports on forever.warcraftlogs.com, zone 3000, the 407 that start between 2026-11-25 20:03 and
  2026-11-26 09:40 UTC and hold a hunter the count keeps, listed under §Reports
- Game build: 1.60.1.71091 or 1.60.1.71139; the hotfix each player ran is not recorded
- Date: 2026-11-25 to 2026-11-26
- Pull request: #1287

A hunter counts in a report where the tamed pet, the hawks and Auto Shot each land 50 or more white hits. The crit
share is crits over hits and crits, with glancing and blocked hits left out. Ferocity's rank is read off the hunter's
talents and the level is the talent points plus 9. A hunter in two reports that overlap in time counts once. The count
starts after the hotfix that raised a hawk's crit, which #1252 followed.

| Ferocity | Level    | Hunter logs | Tamed pet's crit | Hawks' crit      | Auto Shot's crit |
| -------- | -------- | ----------: | ---------------- | ---------------- | ---------------: |
| 0        | under 60 |          29 | 8.5% of 14,064   | 7.3% of 9,848    |             6.7% |
| 0        | 60       |           4 | 28.9% of 7,338   | 26.4% of 3,113   |            15.2% |
| 1        | under 60 |          18 | 10.6% of 7,793   | 9.6% of 5,664    |             7.6% |
| 2        | under 60 |          11 | 11.5% of 7,007   | 10.4% of 4,211   |             5.9% |
| 3        | under 60 |          20 | 15.8% of 7,807   | 14.2% of 7,468   |             7.3% |
| 4        | under 60 |          15 | 16.8% of 7,963   | 16.6% of 5,332   |             7.5% |
| 4        | 60       |           2 | 26.3% of 1,258   | 24.9% of 703     |            21.7% |
| 5        | under 60 |         237 | 23.8% of 208,214 | 22.8% of 119,911 |            10.4% |
| 5        | 60       |          90 | 28.9% of 109,840 | 27.7% of 47,051  |            16.0% |

The hawks read 0.2 to 2.5 points under the tamed pet at every Ferocity rank, 1.0 and 1.2 in the two rows with the most
hits, so a hawk takes the tamed pet's crit with Ferocity whole. The count does not explain the point that is left.

### Reports

12wA9hFLR7mYBNPX, 1A9qN3vFnLPM4akZ, 1AVrQKcBdLJmaF6N, 1G3NPXBkY7Vvx6jW, 1K7cr3fVaD4QRYHz, 1XC4jKvcAyZktzTH,
1XJvGy4K29CqxtQj, 1jAJ2pgtxkfLdhWq, 1nB9YbGvTAXLC7h3, 1qNVZ8XF2TC4g67R, 29MrYfxz8BdWqmhC, 2BkW8P6pL7JrztGZ,
2FVfCBXqNZgJnpjH, 2KD8HhRcfXJqYzpv, 2hBxGvrzXMnpJt4c, 2qpvPaVgtz7khRn3, 2rJxb1RpjYkmz7Pt, 3FGkBmMgWTQXZRxY,
3HJNrtwhxKynXm8z, 3MdKvpqbya8QYk4W, 3NCDjtpRFQbd9Avx, 3TNQMF8arDxj9vz1, 3bdmL1fQW2gkDCrw, 3bzCJLdDhwKmYa8j,
3dzYhpwMvVKHmX1F, 3jZvFa9pC8rLG6c2, 3nracqh2KjfWRQLg, 3qZtPxYMzNTDXvmh, 4AFt1c2VHhTRna9z, 4HQ1Ng9xkazLf8rq,
4QKanhBp38dfb1tL, 4bacBRLPdyQvWVn7, 4bqTjQhJrYA6kKWz, 4gqjzbCTP3nA2yD8, 4zkBMtWVxb6HLnZN, 63KpD7Qt1xLzBHJT,
64ntVcPZwXY7kvKb, 6Br3QtTMwk9JNg2X, 6BrpVZPh74cydqbM, 6KP7avMb1tZXfNH8, 6VD4zFYXx82qHyT7, 6Vmr89YJvHq7RzMf,
6ZHd4kRw932cYWNQ, 6f3d2nHhWMKGgJwZ, 6gbwpYXzM4Wm7nr3, 6qJMzH93PvkgmVNw, 7BMXw2afGkVW6Cd1, 7NyjFqJx3BngzWPR,
7tGFmN1zKv4knwQA, 7wqWbtGx96MXTa2V, 8AcW4yxtHkj9wXGb, 8FdVD7wgkJa9j6t4, 8Q7VFakGBjR1c3KH, 8W4rDm3YdcFk9QCa,
8ZvfLytrD7bWKARp, 8jZFPCGWvHQ26knA, 8jbtwJ7fnLRFBT6D, 8r31qnZNPADGCv72, 92BCG67FxpYdWyfQ, 93ZbFTJ4XGADnKMB,
9PY7nrxfzKXjDkVJ, 9bJHWn8LzhfxYCkR, 9ckn2Q1h6VjNfw7K, 9mQVf1WDyYJ7j3wr, 9nb8ArYjHpPmw3Qy, 9tZxB8WD2GzwhXQp,
A4Vw3PFdZRLhHybB, A92c1CMH4t3qDphr, AGFnytWBwrVaPYxm, AXcVCy812f9FGDNw, AYTwFzcqgG2Dm7K1, Aqg3P9fmpZrJ4CwN,
Av8mtj76rdLHqRgB, AvK3jpb1JBt2wyYQ, B32Krvw4P1fzD8ny, B7qR1kCm83TXNwvZ, BGH6gdY1zKyjpZMa, BGHKrvAgWTxnkRXd,
BHgXWV6NkM3bQF1q, Bvk12tMYTbAfHw7m, BzPp2M6aNvYtk19r, CFjgMAB2wVJHGX8N, CJH3YDmBK6pcgjVt, CNZ9JGFch2M3DbvX,
CT8QKw4ALXHRZmn3, CWq8ZHKawhRxmd2J, CfJXTA47Z293BaVv, ChbG2MpyJjDWPfV6, Cmf7VdNGWtM4a3J2, CwRq8LgZbADvXcHK,
DGa2nzvKcRwJ6CQZ, DJQryzw3gP9qHZNx, DYNjAwqfV9G12hyW, DmJcrQzhtqkLpRaY, DmyMJj2whV8knY9r, DqwAgZCvKRxpnBGY,
DvH9VxKZnagAPGL8, FAJ1Q3M7PcDLnTxH, FajwxThymGq98kLt, FcNAQLftpBW2rv6g, G4L8KgRm61qbnHty, G4ZpPYthm1gMKb9a,
GKjtXQ4aVknfxNCR, GLVNazXZdFA439Yq, GaKFhmvgVN4cCWQJ, GfaZ4ARcmFjpMtb2, GhDRyj9618bTdtn4, GmTKNL3VpC27PkfX,
Gv2FBjkLXPxK7YNq, Gz3Jn1xBHqWhMbwZ, H3PBcWNzjfC6KY7M, H7K31X4mQGCRVfnb, HBcC3an7vdfpbLTy, HGnjTM3NqDwh1CQf,
HNavVjYQmtyhGn8K, HW9hvx8J3NAdGYbt, HYFdBcTgjfnMRX4C, HtCdNDkPLhZf3g8a, J1KVMcjD3xgNX7Bk, JCfFta6br3Hw1vdM,
JMNmWt19q8QVvkpY, JTzgpPMqmr13BHRy, Ja28gyVqTrnAhmM1, JaFQcdYDj2pNvbLr, JfKBv1MgckRq76P3, JgRbDLF8vHfp729x,
JhCZx6RdXcT9L2yv, Jz14GmZkWaYdCPXr, KBnjRA8TXQxJwNtk, KH6hZQJz7YR4tNjm, KLRbwZfpmJPT4dCA, KMaJnmDyqGTVcv6H,
KdZVh9AqXkLM18jx, KkZB1VXfycWg28nx, KrNdChkVYp9GJ3aA, L1gGVavzmRA3BykK, L61BdtNT2CFWwaRG, L7Rx8HnrwWcFDfqB,
LDvyp24V6HgQZGmd, LTNFDrMnqYCtzXJK, LVvhpd136xZ8HWRT, La8TxdBfD4Gp1wKC, LkYjKP3HpcG7ZDVA, M7caPHFW2BJZjLRz,
MJ49zAcWhDa2Nnrw, MKn8Dp9zQhJtHWfG, MTa8XcZ9rH4gDYvn, MTgfk2YBLN9nV4R3, MYkpnHFaCLWvmh7V, MhZGjy4QXzaPpJWq,
N1ZgnHzx4JLbFPYw, N8hQafkMJx9tb7yF, NBJvV6WCg9ktZrHD, NRnK38v2WYAPayMm, NWTfzQnt3JXCGZ6v, NjnA7zPdmZhgCvyL,
P1Yz6t4kZLC98Wpw, P1aBqzJRcVZhCYQL, PHy143FknGgxL6AW, PWzFjKRGvQ4aJBAZ, Q43nKX78JvxRdh6r, Q4jVYqC2Bpd7FPNg,
QAB2W367jrpH8kYd, QAHG1vr3MPcV2y8b, QBdMrwfVtzZxqFYv, QHK6tcgXm9wFAk74, QkWrA3jRLmJD6aPp, R1cQHJbrNgTXhMFm,
R63P8NcWZxwt79Hv, R9C4DjrcmLBbgXWw, RACFcXrV8YD69j2K, RBqVt632hkyj9vaP, RFrtN71cmkTXM4jv, RJMK28P1nNLgcjxY,
RYABNJ4qGpFjXk9n, Rcp8j7r6hQtJbyNk, RgN81mBAXz2k3cGT, Rr91k3pbDtMnYVTG, RtWZ7d8XjCG1LKbV, RxjL14afKbdNh3tY,
RyfQmNZ7L98cnBpb, THCyNfR3JXj9vWcK, THkMBbWLvrPa9q6h, TNdmDnFgRHrtz4Jc, Tb7wK1X9Lz8MnaV3, TcKxbaz4A3WBZvH6,
TkY4CQ3h7WNJqtfA, TmBZwDgyXMC8GKPp, Tnc1bgGFpamJHD6r, Tv6aWwR4QVBnFPtp, V1nWCDhLQ6fpNw2F, V9k7jdXABJnCwfQZ,
VgQdNcXKa6B8LR4k, Vjx7dT9nXbRmkqQw, VpjMBqCdtYmWQ4HA, W1ACVjTzY7nHXy3R, W1F2q3XdN8H9LjVy, W4z82APj6qDwKchf,
WCTzNAKM2PgDdFJj, WGrXTRfcq3BMzPxk, WKbYMyPzqmkJTAGQ, WLD1jTrCgfVYFBvM, WLvzD7AYxw9pR6Vr, WYNtyXmCL1P4KBDZ,
WZLD8RYn6T9kh2mx, WZXgpDwmCfTv7Yk4, WdvB1JrXAkypxZhz, Whwg8QmTVK1JNXac, Wt2qbDgjPNafTZV9, X8Yb3PQHKMA97Bdt,
X941dcrn8tMLVkBf, XCB7wPzLkN9YnTmr, XD4FQ9HvWMbm1tBA, XLfvNcz8tKjD1yW7, XMPjtKTVnHkhb9ND, XbhTMQLfD6VJrYpx,
Xp3rhN4tKJkgFdPY, XyVKaMfxgZRp7rLm, Xz24LhYtvqDA6aPT, YDyr7FwtGkdgpA8V, YGdx8WZjRMhq6L7n, YJBd6QTy7xkDH19K,
YJrLKc1wDyTGkMjm, YLvhAk7wPj8bzX1J, Yfr6LwXkH9mJqN4A, YpBzGKLbAVq6Dr43, YxG3cZdArJ46HNXf, Z1PCKkAzH6QhmpyN,
Z2xy789wNfz4WFQD, Z64krt8GfAJDhdYm, Z7TVDw6p39GYLCj8, ZJvCD2X6M7yftYW1, ZKGcg2RbXDTnhCkj, ZLm2aGTkQbR7gpjd,
ZNd37MnxTQYGVbWq, ZYbm1LBP4HnfwQK6, ZcFaA1KDdnXxzRYB, Zp9hkgJfyjLWxAQq, ZqAnD4d6VHvGQN1x, Zr8AN2fKMWdFCg3R,
a1WCvFk7BXJ6DxM8, a8bVYw4zgT3NMpZG, aWtVg8wzv7bFyXHY, adPFLH1b8zqMWDZK, ang7zYdGNRtBh2AX, atqAH6WMjXm3YfQR,
ay7n6RDrXkmb2fYH, b9NdHQWfkC64B8yv, bBTt3RVzLPQCM1wK, bTRgBXGkyFd1D7An, bY2xt8ZrXGNcPvAf, bZYvAz8kPTmNVtKB,
bgGzTK8VDL7cCRHP, bqrQvcZ2FTxfn7Jp, c4wmyCQYZrxDNh3k, c6tkQTwJdZPb7aYR, c9vzf6LtjDaXbVN4, cC3YWwPrgVtp8Ddj,
cH3LJQ84q7WAdfgM, cTfkaBh9V4nFxmy7, cWb8VJ7fmGCZ2BL1, cb6hZm1kBwGWtgKa, cgaQrK6xF9yRPdD1, ch6jXq93FPDNvArH,
cphbT73dj8g6GAkz, cxkQ3K6tJpnhXZFf, d1hGCM2NLWpYQqbX, dQ7kXBZLzRDxrHbT, dxhpz1cFGwrVMC3A, f3nGB2Fqc9zgQaWD,
fBx6VJQ2PM7btGk1, fWTHCc7tQvLDA1xV, fWtmMyZgF3LjxRGQ, fXkwmh4bZGRg3zH8, fYpK9CJ27QN4AZWD, favQ8mbXkK9437cz,
fdZAkPjzyn19CHbL, fjrtQBagXvnbM3G7, g3LRZHbmVTDY18XM, g3dHRMAVT42XKqZp, g6ZRhL7TwK1BJWxb, gaHypcwQbCGRkdqV,
gcLXpAMJ2RDxWnBk, gmJ4dbnBNFA8rTMp, gvV8mTa2jxY3Hcnb, h2JqbQPZfGANzBd6, h2pKkLqx73vJrgGy, hL3KvFkVCtpRNfHw,
hTyBAVpZKLfqXQ7t, hWLT2CvqdQFp9PM8, hbw1zxZQnTAWvFyV, hfH1kBPwGCmy3VW8, hn8QGwB2FtHgRyAV, hnvD61cd79Q2GFBW,
htnX9jxB6wKCaLPq, j19bXkc2mT4dA8PR, j1KxzPDvwZt97gQy, j1p3k4acQBZWhznt, jH3JMLa2VwGN7YPy, jJd4Ry37YFkHhTN1,
jLgxrFZcVkCDaTHA, jMZw4HWfLyApRa8V, jPVq8WvDQa4JThLB, jZMD1mWncYwBkX9F, jZqGgBnVJ91aK3tY, jpnvDV3QFwC7gTcW,
jzLBZwqW2gNYhD8H, jztPrG261ZqFbwL9, k1j4wYvBAnZCHM9t, k3KJWf1dxhbqjMB2, k4rC39R1LZXcATNQ, k6YXFxHcv9DZNgwL,
kGzKj9yWx8gLMPFT, kM3jZcvARyzm8g2x, kW2VAcCKXaQgj6nB, kZXTfwhqCK7Q6PF3, kbhVApwc1gvYF7Gy, kbx2PYpK6qHjV79G,
kg7j4M9KJNPpHCqb, khjVy7fbxvYtz63a, m1GQWPaxnVAX4bg9, m7f83g6HndLMAakT, mC7qGbT26zkZHQaB, mMYKLQFGgvH6axrn,
md2chvkJRBgyQYPq, mrhWB6Mzg34wT1qH, nDyNz9BtcfVg2H8q, nHzc3m7h691VPGKv, nVfG8kqgDMrbXmwH, ngvNJjPTmLkdw98M,
nkNArLz9Q7FMwcKg, nmAGgfbYwTxcZ6B8, nmpJaTF61zPqWckG, p8XLhYkDTGfQFjAx, pGMDAc9Wv4V1m8qw, pVX934LNGKczMmZJ,
pgZM9rYdbH7kz6CN, pmtHqLd9ywK6BJPg, pn6B8K7C4xRyDQbV, prC4b1DB9TncPJXa, prh9jA6v37XP2cJK, q4khHdzQCGWV8cMK,
q68AFBTfRrJzjNYZ, qFdrMP2gVQaATnKC, qJxFhLga6mkAYNpw, qW3hcRgV8C4D7fAN, qfKwNhTtkzG1rZmb, qm9TLQKfh4XkxdtW,
qynGxgVrwmYhQ93c, rCbHBVYzFN6nDy7M, rDCbzvZnkPKAFJ36, rJf1AGxtBkPy9ZHD, rMCmp4hHbNaxdQTZ, rjMPpLXcgH31NCWa,
rqPpTAzV4Xc6NBtx, rtdGbMyCzjgR4Vwc, tJvhRTkQPNmgBM8c, tK1BbFh92rA6HXcv, tLfvwqZm8jcWBdaF, tbKqjNPy3nwBxap1,
tcAPa8XG3QY1DpMz, txHNb6dYVqwzjMJ8, tzN9JLRanpqFV8My, v6gbxMKjrP4FktT3, vA8kjaYKXCDTPxdy, vDfzy623whxAVnWY,
vXLHn8pqt4zrVK1b, va1fkYqjDAMbNBG4, vbXfBWDQk8jAPTr9, vfNVtYMa19647HJR, vfV4T9WyNFB1HwQb, vmJ16Kkq8zQ2ch7D,
wFgzpDZ6rxMT13Ry, wKYyhHrp8TZWFz9C, wKb3YyPNq4Q7CDVA, wV7bpgqaA1GLdRW4, wfdRa4DnPBpML976, wmQ3dpMzNPHaFgk9,
wyZgjpLPqcCWHVmG, wz46qKkG19hmBY8T, x6TJa7CXVmhZgd1W, x6V92tPXbK1jWvaL, xPJwXD2FQd613AWp, xXk8KAGhg3CF7RPL,
xmYdkaAbr4X1JRQq, y62Q7NJjgZHLpxcW, y6YpaBLZCdtTG43h, y8wCFmfk3BvLZjqR, yQZW3YLGMmhkX7NH, yV1M6C48b9czNZHW,
yZaNF3DG9bQRXph2, yaw819pZgLBxrGTz, ygtVvKPNGAQhBxq3, yhXdzvAYGjfwtgNa, yqCabnrPcQxDR8NZ, yqfkAMGzvpZ8XQF3,
yravfwPhCTXWFGJg, z7gjr82BtZCmLQvX, zATdNgKZC1kGDqtJ, zNMDVR1d3kfGCpw7, zQPHjYRkd9BDCZq6, zRTpHGjQXB6kKwb1,
zWFnYbDkx8dM3yLw, zXNTDmVdxPwLyKQY, zZ3nJWaCgydYb21N, zaNyk8w97ZvWtQcj, zkQvK71XcYpjtmV8

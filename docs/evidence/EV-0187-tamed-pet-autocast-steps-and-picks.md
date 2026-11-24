# EV-0187: Tamed pet autocast steps and picks in live logs

- Claim: a tamed pet's autocast decides in 0.8 s steps, its next cast two steps after the last at the earliest; it
  picks evenly among the abilities it can cast, not by their order on the bar; and it casts Savage Rend only about one
  time in nine that it could.
- Sources: Warcraft Logs, the reports below, the tamed pets of level 60 hunters on the Forever servers; all fights
- Game build: 1.60.1.70888 for the reports from 2026-11-11 to 2026-11-12 and 1.60.1.70938 for those from 2026-11-13 to
  2026-11-16, as in EV-0131; the build of the reports from 2026-11-17 to 2026-11-20 is not recorded. The hotfix each
  player ran is not recorded.
- Date: 2026-11-11 to 2026-11-20
- Pull request: #1201

The reports were found as in EV-0134 §Method, and a hunter counts at 51 talent points. 546 pets that cast Bite, Claw or
Savage Rend, from 468 reports, are read cast by cast: the pet's focus at each cast is the cast's logged focus, and
Bite's 10 s and Savage Rend's 60 s cooldowns run from the pet's own casts before it in the same fight. Raptors are
the pets that cast Savage Rend and cats the ones that cast Prowl.

## Picks

A pick is a cast of Bite or Claw while Bite is off cooldown and the pet has the 35 focus for it; a Savage Rend pick is
any cast while Savage Rend is off cooldown and the pet has its 50 focus.

| Pets   | Bite picks / picks | Bite  | Pets casting both: Bite in 90% or more of picks, under 10% |
| ------ | ------------------ | ----- | ---------------------------------------------------------- |
| Cat    | 6,055 / 12,847     | 47.1% | 1, 1 of 41                                                 |
| Raptor | 699 / 1,463        | 47.8% | 0, 0 of 5                                                  |
| Other  | 29,732 / 60,826    | 48.9% | 0, 0 of 257                                                |
| All    | 36,486 / 75,136    | 48.6% | 1, 1 of 303                                                |

Savage Rend: 179 of 1,589 Savage Rend picks, 11.3%, over the 5 raptors.

## Steps

The spacing of two casts where the pet already had the focus for the second when it made the first, in 100 ms bins
(the 1.5 bin holds 1.5 to 1.6 s):

| Pets   | Spacings | 1.5 s | 1.6 s | 2.0 s | 2.4 s | 3.2 s |
| ------ | -------- | ----- | ----- | ----- | ----- | ----- |
| Cat    | 25,908   | 5.5%  | 88.2% | 1.7%  | 2.1%  | 0.8%  |
| Raptor | 2,978    | 5.3%  | 86.9% | 4.8%  | 1.1%  | 0.4%  |
| Other  | 129,674  | 5.1%  | 86.6% | 1.7%  | 2.4%  | 1.3%  |

On one target without a break, 172 stretches of 45 s or longer and 104.4 minutes from 20 s into each, every spacing:
1.6 s 67.8%, 2.4 s 20.9%, 3.2 s 2.6%; Bite 2.10 and Claw 28.3 a minute, the focus at a cast 33 at the median, its
regeneration 12.4 a second at the median.

Reports (468):

16hJKBmq2V8tnzXR, 1AkGzHm3J4B2xXFD, 1DZt4gPHfpTCk9W6, 1h89JnyWjpZxfXdk, 1xzPrQWDC4pv9HhJ, 1z4vRDqGdM2cVPXk,
27jZGhfHgp1LR8MB, 2DmCQtJVrXTRBy31, 2YTQxaqzvAZPwKW3, 2ghMyWmCFVtzTGv9, 2v4xNzmMKDbhp18n, 2wMznDRvmkpbtHxA,
31TtLPDXrczfGwAC, 3B8AcaGNRKmdgPDj, 3BAz1TDGL9kjy8ZQ, 3H1fvFaqxRL79PY8, 3R69PBjxZvcfbnX8, 3RB9X8HVxgCJ4My2,
3fFatLkqGnmjQ8ZW, 3qfDVTZ1J29zPChn, 4BtbFKLZ23G67dy8, 4J7dhZxb3GakBQ8p, 4Mcm2YV8yDhPnwrb, 4R6dyPjBXmCr3Tq1,
4W6d8VFkgmCrqGvc, 4XMGR6aQ2hd9Km7q, 4Xfp9L3bKh2aPYAn, 4azX71FvCjQgmJkc, 4hQkw2BmKJGd1XZ3, 6BZ8AkvTNwLJP9tC,
6QMFGaq7Ptjmyn1C, 6TH3QvJWXahx1KqP, 6WG74An9RrCdqkwD, 6WwaxM19bkXqJTpc, 6gnAywbVMGq7vRxZ, 6kZMyGqj2PhpLdT4,
6yHgkT8LqP42NmxY, 783W4MKARDyQLNBa, 7BraqHpPmfyRjwc6, 7FBaHRhMVKjC1QbT, 7GcATyt6WRjBJvaQ, 7NhKDGwbdFqWB1La,
7Q6WxzYRwjNqXvgZ, 7TXcFNRzjwHdAybn, 7dDbtWFhA8BPn4vZ, 7fy93LdcwQhvF2gP, 7fz9Kk8rY6jyQAbL, 7pmhcYx1JNCKvWjn,
7tbWfYqJFd4vzL3p, 7yLVTagpRbjYMhnz, 7zvM3HLnwkcBP9FW, 813rPaLtcdpHy67q, 87Hd29BkF4Nyt3nJ, 8FAmTNqRVLb7zgH3,
8FCjtBVyMJPhQv27, 8G3K1TCaq4fLXhHz, 8JTGAcnFYvtmjLN7, 8TBnW4cr6HCKp1xz, 8YfKRvzTxHCVpGtQ, 8YhMBZCPKckdbfAn,
8bWVvtJRXFQNTAjc, 8jzXTaqr732V9DxJ, 8rAWbkX93chqKjFG, 97aD8RXPqZjg13Hn, 9H2chVMdn1ZtxXpA, 9Q7FYnBbyKd246WX,
9Zdv4NLm3FAwQKaV, 9b71qzjAg6Myt34K, 9nX7rz3h2bvPJBaH, 9vcQpzdbZmG4kyTP, A7ahGYbzQ8M1ZtF6, ACHkhxt7nXqjJpKZ,
AMRnwhKV6JvFQYpr, ARFXzbqHDpBLJdhy, ARwPXzLJk78HgVQq, AY62aT1fNFrQcqHD, AYwLfvqh9QyNj6zJ, ApJhHfRkMDqFj3rb,
AvNwTdyBbDncFf2h, BHJMb9CGa1TgjtAd, BXMpQqGY7caChPjZ, BYXH4tk9GMW2C6Z7, BfVjHYvWdKkP8Jz7, BkPpXvACntQMqdVh,
BkTfV9yAZJX8P7wF, ByP4fLgNnzj1RmVr, C8hLYZf7Tdgz46DN, CF7ztDwY8pJ9KZhn, CGxH3FBAbqZNkRzW, CMGKZHRVPwzvFmqy,
CtMF9BgnAmWkTbfd, CvJfN7naWYjMhgzV, CxTmQRX2ytGcwA9r, Cyn4arZ1gJpYRbQK, D2jmgvdkfTyMBpxw, DBRqygp13jPC6c82,
DBzk8qw4yFWtb9rd, DH3dT6tYLKmMrNz8, DJCRPcry8mjpb1nH, DKftp1VTZw9hBLrX, DL1NbkX9GPvBaWwd, DW9YkdvyrjtMBQnf,
Dbfd6avYpzwTCx3N, DwzYTd1QPHFf6qJZ, F2vN8aVjBz1PJ4Q3, F3BTZGrW48Mp6n2L, FG3XhVnPwZrtv9Ha, FHqYCNcKntvDAjxg,
FJT3mz9yRK2PgVkM, FJdrDz9w1Vxaf2cT, FTQ8g6cvbyZ1wB3m, FkBxY9zn4faM7jWL, Fz2wJmfGKPnvTy9h, Fz9b1xLP2NRpMJdk,
G29y3QBWV7XDR6mj, GByQ7g9ZPkqxFJXh, GQA8HWjFDN37CM9x, GRnMHmK69BXfWVTj, GbvNW3tAP68ZpmcY, Gc6ZR7WmDdhM19tA,
Gjd4M6kTqVPXY7cL, Gkb3rKHF6V9gz4pR, GzqDHgcPy4vQXBxZ, H84Fk7tgwWy9vX6h, HPnctDh3axdKw1MV, HcvJPTBZptmKXabh,
HdAz2cWg6B417rp9, Hk68XrDPfTMGYL4g, HqyPGnV9xLmtMK2D, Ht8VFJxmnCG9bQhM, HwZDJN7MqYQkd8pG, HxCQat38vmjnkpb6,
J38xbrFytgM9p1LT, J49Ynj67LCBFm3yN, J6xKtDYPr3hG2BFZ, JDNPdfVbtxTmH42r, JNTRK1qgDtkCz4rQ, JQgB3aWcjTDV9nkM,
JQwbjnVYyGcP4rg8, JRQqNL6DkvWKYdPB, JTKwjfrDBp2bAnm4, JbKBmGHwCcyVzkjt, JbQyngR8NWXKardc, Jjmv6drycNXpDxW8,
Jn9HfV4qmQrFgBXT, JnVckZDLTwAgGKaQ, JrqDZX4LT6NpxzM7, K6Tj1zNfVwJy4c2W, KG62VXF3a97CxtAg, KMpAYyN4rwazWnmx,
KcdpBWTJyX1xLFfh, KgqPQWkpz9FdNb1D, KjMFGvmkR7LDYz8y, KndmGyYFazPjHftg, Kphk63jtr1wb9NyL, KqZ6xWAfDLGrVjYJ,
LCXrjHcgnAJ2yBaZ, LFNZrg2Cpty63BnK, LaC9tdXyrKb2P8Mm, LahKC3yT6JQRwtMn, MRLXmFGyV268jWwt, MWrPDAVFCBXzh8Z6,
MXYFRz46Na2DZjcB, Mg2ay1xDjkYpBqRX, MgmdJqxBWA3Tw2vz, MtxafB13DwLnCFzq, MypmYhB3RX9FJ7Kx, N3nHF2GBgdT6qx1V,
NFcRw6HdA7g1BGm2, NPHc3azxL8bmgdpF, NTJzVdMZBQDaqG83, NbMzcKyJPVYBDpnh, NjWKPpHgymCT832A, NrpM7DJQBKWqbzVF,
NtxKh7PgTAaZnHJd, NyYBCtKgGQr8Mnpf, P3ZHAvVJTq8NBmFc, PDFVdp172Kwz8aqj, PMtHv3y76FCj8AWB, PZHRWv8Vb2BTAGcK,
PdNq3JxntCKgavkm, Pjd9MWAXgCYTcn7Z, Pr6vfwKL713bYZHC, Q2qrkZR6NnbzLmaw, Q3XYxAvj7yDzp4c8, Q7V24RA9HTg6F38p,
QAgrxLyh1CVwcf6k, QCRFhVqmHWMK4Gv7, QHXdBJRj2GPpC3n1, QNwkdxbMvWmrqjAY, QRCgnhck3KVq6vBT, QWPwD7agVH9NdBmr,
QXFd74zDW8JZMmKx, Qb1yWxa8vHCGL7Yr, QjHW1hvfbNGgcLBd, Qm9XJwGdqNzhRc42, QpfnBdxWJmX42rAY, Qyf4CGR1hgczwvp9,
R4rLFGNfpg7wxHYC, R6aLAPN83fB9mMkv, R78CFp1D24jAatwZ, RJQAMwcxHypGYbqV, RVma8DjNx3qnLHtw, RY1yA2PFCBwb4MZa,
RgMfxbJrY89nTpmq, Rhx2TCj8wQN1XFkZ, Rk1qQxgLcZPmpBaf, Rry6WVb2NGD1dXMx, RwPHNk3B6fMGvQcV, TA3pLxKFXQ9Ckm41,
TRn3gPfqD6Bj2FXz, TcLMDwAGYvQk67tC, Tg1Hcb7z2qAXWkCx, TmQyqnM6hN217LZC, TqDxtfJKXPCLk3zG, VCnTF69XAMabJQfm,
VK83G7MdBhLRFAfa, VLz78BfZyWa3Cxr2, VMfN97kLXxyTpCwW, VNpDMaKXy2LHAPZh, VYBjgJx8aZtRhmpy, Vdv8p3ZnTB6fGMQN,
VgjdJkmCTrtYcXLB, VjrAGFkTvNy2W98B, VkLX4CZnc68R2Gb1, VmTj1LcnghRpxrMy, VqAygFH69PTC8Y7w, VtcPT63wQ7WaACvj,
VvRNkLTq867HMwn3, WFHghVcLRtPpw4MZ, WcpXN8F7vyR4gDK3, WhPpmnzB6rR3gD4Z, WkrD6xYbcp8JZQBm, Wm6K4G7abfxLA1PX,
Wq3R6PJNajKzxfvb, X1N7q2tCL3wrxYBP, X8PH2CG1hMkcfxYJ, XHCPnJGcKfZrhQYA, XgGTyamxrWvtQ62w, Xhz8p6yB7bYqFaJ2,
Xv2QmnyK8RBxW9jZ, Xvt13yTM2rFRVwdP, Y7BVrgx69aRjK2dH, YFtvD87kznr1CThg, YGR7VPK1CQgc2Dnm, YH7n2xZqPKzWjQrR,
YHRv2c3QqWZaLnmh, YVWgj6ra27PvmNnx, YZ6NdFDRQKL2CTmv, Yf79qHLZAjgnTdJw, YnVFyQq3w28mr41R, Z6kDfqLt9n38yY2a,
Z72wdaBH3CkrfQWz, ZM2PjyvY7WDRAgzr, ZTGq13h7DH2bYrNw, ZcFKV7tLHbDR4qJB, ZfbY6RgWjx9hmpGz, ZhbCF1LvkQ9gBAMy,
ZwvyzH9bxRN3ghCX, a4ZdvYKyhcLGnFRP, aRfyqH7A3t9bZXdm, aZKTvVcCHLwhgXzM, apMdm4GJH32L1DVC, awFLxG8Jk3DH9zqB,
b3aG821WYQj9RVhK, b6WxZXCdjrMDyLmc, bDcBjXw82GvhYW4F, bF8hVwJNDT6HyRWt, bNZvk628gXLz7t3c, bYaT9KXtGwqvV6d4,
bavD2gBYZHkqK7Ct, bg169rGdXTzDhLnj, bhNT7dLXrkDF3JWz, byjaGMnmJRtLpH2g, bzwY28XH6Vfc4Zpg, cBGdRD4xLq3fW9FZ,
cFVPYj61bXfBvgh9, cL3KMWpnXyvDrqa2, cMnAXrh1fLByH6P8, cb9gzXjCK3BfH7qM, cgMGCdPJQ84wzaBV, cr6qMtj1vCTPHVfw,
d1vn7jt4RH2ZmLAB, dA2gmBnLHkjM1bxh, dBNfgVDLMmnc8h4v, dFL1Bj4mnPAxvXHc, dGh1Fn4rgkjB7Vtq, dGvZm7g9wBCAVHLk,
dMxyPfarzTc3tLbk, dZkY4NtKAJCDgTrM, dnDg7j9kRw6WtbML, dvkzFtPH3WR9C4Km, dwJNZnKW3ApzMCf6, dxCHgDbw7z8PV3TK,
dxZCXJvPLKAB3cnQ, f36RaB12wyDcTCGY, f3hyv2GtFW4b8wJV, fHDqyWNB3ghMabYv, fJnW8D7qYaRXcT2M, fNbLwMvKCrHc23gG,
fNxhtGQYCBnPbzcp, fTx6mHq2par1RPJX, fZdz8MyCHKFkrRwB, fa8LwHZNJ6XYAntQ, fn2Dc941K63CNvmj, fqcPvpYkFd6rgbHA,
frdmcpgKB6twT8jH, fwqdV8T2QLcY1Pj6, gA94CztvFpWYyJ26, gBDMLv8J1K7nqkpd, gRmxdjCHt6XDMBpa, gTnLQqam716J2hAB,
gXhWFKATBJRaf3cQ, gmKvNqHFJLZXfdnP, gmN7Vxy8tPaD1bhZ, gnY83dbNMh2GqQpB, gtNvbhF9aZCXyzVk, h8aVABn6zKPyMwD4,
h8vCa2P43nbH7JwX, hH4YRGVFz3KnN6bw, hMv4dJgY3pnyVXTk, hWat9D4qKrJkBZNj, hXLT6BWagD32pkZv, hmDYk8MKdpqx1FzJ,
hyjDfc3BmY6X27z1, j6XFRzHxTC1kvVpy, j7xyJdLXHZNA6PQg, jJca4GYpvxmbWR13, jQWwcktazNJ94M3V, jWnC8Z7xDvpRcw46,
jdDpgCLc7h4qHG6Z, jftPgLMa8JA7bkHv, jmZzbTWpqgNRGy7M, k1jVRK7g9APNTLm6, k8NhDFbQjrCPzvn4, k9fKT8ZCtMFn3vab,
kGpgavmFADBVRhq7, kH1dt4TpQjfYmaAv, kaH1G4ztmYpgMTJ6, kmxBzfV3LnR6NDb9, knyCYtGqvZDWzxg1, kyrMKb6tVRGNX7c3,
m7qPWtJdX1VHzr8T, mAW1KFR9YP8r3bNx, mJZWhHDAt19gaKdw, mR9k6PLdgMKqvXJy, mRnQ6xjzVDvw2PMA, mTw3kz4cgqrv8JpA,
mV7NcC8vQ3jkTHwB, mW2TtC9YqHZz3vpn, mahRZxnfrQ2N4jKL, mak9vPD1KCtHW3n4, mb8FgjhwVTLZ4q1d, mbwzgt3VpLXTj4MZ,
mhbXcFqf9xTaA1NP, mjrQyAZVTtxJvpN2, mnPXjHcWNFKBhDCq, mpgnDQbLWXY8jJP1, mwrDFNLdhgJxAvct, mxNL9dpv8RXjPrhb,
mzB4Wj3q6Ktv12nG, nFvHyDjrCxLdA6k7, nK1fCTqXYFJxbMgr, nL4RtmJTc9qW37xN, nR1WGzY8pJKDxhQj, nTfhtVKvmw8gZkLj,
nX1qbatTzGZ3CAwB, nbLKdYfXHDam79R6, ncZbGamMVrXYgkRN, ndXfgW7P4bFxJ38L, nh16A8B9RrXNCWwG, nmd7QgxNGjDpcFhT,
nrTbFPwmpv1cL4Wt, ntZ8vdyQ4LPa1xY9, nyMqbY64tL7HhKQJ, p7ATmV4HqZRdFwLX, pAgcQ3MwGkfP8vVL, pCDhtnTv3QjALWmg,
pG8TNJwybLPYWmxF, pNKJR39CwGnmXtbL, pNW7aGKnzHP1MftT, pPq48NbAZGhHkDfw, pWnPwNFGgyX7tkdq, pYATGQ6aCmzwBNcM,
pc7drLMKbFJ1nhGv, pdkwZJNqWKCarQx9, phmHcBvV39RMtjqF, praQ172yvmdZNbgn, ptvYCVPymJXH21kR, pwzCJQZXDcLT39g4,
q2ZdJ13LthwMmyDR, qF9L7VZHR1QG8r2m, qFbWHnA6PzydX7Nc, qNVWc2yZJT6k8mvh, qQ4BrAY1Xjh8Hfvy, qVnt8TGrHXY3B2Q6,
qVvZxP3gNp12KWyb, qXPhR7Jtd8jKTB9p, qbZBRHy9kPKTVC8Q, qcdyZAr2ganDtxNW, qgKwp4NVTZYJAzv8, qrwtb1GNA8W7L6Xm,
qzmMxvpaADw976XH, r9b1TFDzPk7qjQfc, rJb3LkQVwa2CGNDW, rP6X8twFHyRKk1N9, rQpg6YfW7cLRNAVM, rRnVJhQFa9zf7NqK,
rZ426NnhjP1QMcAd, rfz4pR6mx1w9JWKA, rnfRqHt9Vk8g1JB7, t2z1VPDRbwZqfcKN, t3fny7PvdGMRWhCT, tKwnpXND1d9PzvJG,
tRVwX3yb7Q869YZh, tW38cgRn6b2kr4dJ, tjcwR486HaA7Zr1h, twBaM4y7KGPdZD86, v23ytJHDz6QkbXnN, v8L6NkQTzMV2mWq3,
v8Zp2QF7An4d9xMj, vGFaKgLT9mkn1Jdy, vPHb3a6c1KhknLrY, vQZLCzhPNatk7r3d, vc9pjnRr4xQFGTzh, vg1kL4rN2KZVCaRd,
vjN9b4n8dCRz2YTp, vw3zybZTQjH8crG6, vwCXzhMQ2drgpRxD, vwqYr6Qby1fJzk8G, w8n6JQ4DM1GVyKcA, wHM7BJFnKRjdWGbX,
wR8KA9WTgHkD6n7q, wm9ZjJLCbavznpHA, wqBW2GRb1CLKHFJj, x4kKfzHQZ2CVjn1J, x9v27CNzhRgMkXG6, xL1RCknYWN4brG2j,
xWpq74Z8Nf63TK1a, xWwKYM27a6HqLZN3, xX7APTmCNb3qY4Q8, xbJBZj14NWCXrhAm, xdjLk6znJ8Ng1B9c, xfWNbt8yZKzcRPjC,
xhtLrB9MjHqWvVY7, xkKRJPqCLzBa9GdH, yAYRmxqgXkTwr6Bz, yGhgd9Zp6WQMaKkf, yHAmqcVTdCnGQp4j, yVp9KT3QkzhmZXDR,
yWFhcNBab8wYMQtG, yZvwm7pRYDkAPGL4, z9XbfMRh21Fg7QjN, zDpnmHLJdhZr3YRq, zNDcfg3x8R9ParwW, zRkn9GVAy4avdtgB,
zThfNY7CVKqr1MFd, zWPNCbpYTV1xJ2FH, zgqYjRFPMTV2LA31, zhNBgjHxQX7anZTf, zhNraQdpRDX23j7w, zkQabTrD4RHLZN8G,
zmDY4LVX7N8vh9aR, zmTNyXkfqtpgcM1H, zn64f9D8TVt3wFZq, zp9yVrtfTNZJCv7g, zpvfR1gAaKmbWTMy, zrxfPGmgb2YBCqd1

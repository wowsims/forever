# EV-0187: Tamed pet autocast steps and picks in live logs

- Claim: a tamed pet's autocast decides in 0.8 s steps, and the 1.5 s global cooldown puts its next cast two steps
  after the last at the earliest, where Growl, whose row has none, can go one step after; it picks evenly among the
  abilities it can cast, not by their order on the bar; it holds an ability whose row reaches up to 4 targets (Savage
  Rend, Swipe, Mine!) while only one enemy is up; and it casts Dust Cloud only onto an enemy without the pet's Dust
  Cloud, another enemy than its own target when its target has it.
- Sources: Warcraft Logs, the reports below, the tamed pets of level 60 hunters on the Forever servers; all fights
- Game build: 1.60.1.70888 for the reports from 2026-11-11 to 2026-11-12 and 1.60.1.70938 for those from 2026-11-13 to
  2026-11-16, as in EV-0131; the build of the reports from 2026-11-17 to 2026-11-20 is not recorded. The hotfix each
  player ran is not recorded.
- Date: 2026-11-11 to 2026-11-20
- Pull request: #1201
- Later count: Picks, per pet, Growl and Dust Cloud, §Count of the 2026-11-11 to 2026-11-26 reports

The reports were found as in EV-0134 §Method, and a hunter counts at 51 talent points. 546 pets that cast Bite, Claw or
Savage Rend, from 468 reports, are read cast by cast: the pet's focus at each cast is the cast's logged focus, and
Bite's 10 s and Savage Rend's 60 s cooldowns run from the pet's own casts before it in the same fight. Raptors are
the pets that cast Savage Rend and cats the ones that cast Prowl.

## Picks

- Count: §Count of the 2026-11-11 to 2026-11-26 reports

## Steps

The spacing of two casts where the pet already had the focus for the second when it made the first, in 100 ms bins
(the 1.5 bin holds 1.5 to 1.6 s):

| Pets   | Spacings | 1.5 s | 1.6 s | 2.0 s | 2.4 s | 3.2 s |
| ------ | -------- | ----- | ----- | ----- | ----- | ----- |
| Cat    | 25,908   | 5.5%  | 88.2% | 1.7%  | 2.1%  | 0.8%  |
| Raptor | 2,978    | 5.3%  | 86.9% | 4.8%  | 1.1%  | 0.4%  |
| Other  | 129,674  | 5.1%  | 86.6% | 1.7%  | 2.4%  | 1.3%  |

On one target without a break, 172 stretches of 45 s or longer and 104.4 minutes from 20 s into each, every spacing:
1.6 s 67.8%, 2.4 s 20.9%, 3.2 s 2.6%; Bite 2.10 and Claw 28.3 a minute, the focus at a cast 33 at the median.

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

## Count of the 2026-11-11 to 2026-11-26 reports

- Sources: Warcraft Logs, the reports below, the tamed pets of level 60 hunters on the Forever servers; all fights
- Game build: as in EV-0072 §Count of the 2026-11-11 to 2026-11-26 reports
- Date: 2026-11-11 to 2026-11-26
- Pull request: not recorded

The reports were found as in EV-0134 §Method, and a hunter counts at 51 talent points. A pet's kit is the abilities it
cast in the fight, less Dash, Dive, Prowl and Cower, which it casts only to close in or out of combat; the 258 pets from
229 reports whose kit holds two or more abilities and no Growl are read. A decision counts once the pet has cast every
ability of its kit in the fight, so each cooldown runs from a cast before it, when it came 1.45 to 6.5 s after the
pet's last cast, and when two or more abilities were castable: the logged focus covered the cost and the cooldown had
run. Its even share gives each castable ability one over their number. Enemies are those the hunter or the pet hit
within 10 s either side of the decision.

| Ability              | Picks / even share | One enemy     | Two enemies   | Three or more |
| -------------------- | ------------------ | ------------- | ------------- | ------------- |
| Claw                 | 21,483 / 20,484    | 7,733 / 6,953 | 6,957 / 6,751 | 6,784 / 6,768 |
| Bite                 | 18,765 / 18,707    | 6,132 / 6,141 | 6,244 / 6,215 | 6,380 / 6,343 |
| Swipe                | 631 / 1,219        | 80 / 510      | 241 / 376     | 306 / 330     |
| Savage Rend          | 117 / 376          | 18 / 252      | 38 / 70       | 61 / 55       |
| Demoralizing Screech | 330 / 333          | 90 / 100      | 117 / 118     | 123 / 116     |
| Mine!                | 56 / 194           | 7 / 87        | 20 / 64       | 29 / 44       |
| Dust Cloud           | 85 / 154           | 9 / 20        | 22 / 44       | 54 / 90       |
| Scorpid Poison       | 126 / 126          | 19 / 24       | 42 / 46       | 65 / 57       |

With Bite and Claw alone castable the pet casts Bite 18,174 times of 36,482, 49.8%; over the 200 pets with 40 or more
such picks the Bite share runs 44.6% to 55.6% from the 10th to the 90th percentile, one pet at 90% or more and none
under 10%. Swipe, Savage Rend and Mine!, whose rows reach up to 4 targets, go out at a tenth or less of their even
share with one enemy and near it with three or more; Claw, Bite, Demoralizing Screech and Scorpid Poison take their even
share at any number. The one-enemy picks of the three are as many as a second enemy no one hit within the 10 s would
give. Dust Cloud falls short of its even share at every number of enemies because the pet holds it while its target
has it, below.

Reports (229):

16kchwqKn3rpyGMz, 1Qxcn2mJpNHwatA7, 1XJvGy4K29CqxtQj, 1Y4NWLka9pAJbB6G, 1h6tv4PBXQgqb8Dw, 27jZGhfHgp1LR8MB,
28cLVtZ6D7Wp4jmF, 2FN4M7GHLWDmdkRJ, 3BAz1TDGL9kjy8ZQ, 3LXKpR8kYbWactCm, 3R69PBjxZvcfbnX8, 3fFatLkqGnmjQ8ZW,
3qfDVTZ1J29zPChn, 42XtM7Jwdk9gnvqW, 4BtbFKLZ23G67dy8, 4HZTp3WYcdxD1bQV, 4R6dyPjBXmCr3Tq1, 4Xfp9L3bKh2aPYAn,
4hQkw2BmKJGd1XZ3, 4whtjfzW91pbaLv7, 6BZ8AkvTNwLJP9tC, 6PN8t7yDdwrVncJb, 6gnAywbVMGq7vRxZ, 6yHgkT8LqP42NmxY,
7TXcFNRzjwHdAybn, 7dDbtWFhA8BPn4vZ, 7fy93LdcwQhvF2gP, 7yLVTagpRbjYMhnz, 813rPaLtcdpHy67q, 82BgQM3GHJcL7ydt,
8G3K1TCaq4fLXhHz, 8TQJcMxNa2jG7ZLf, 9TD8g1kVRfZ6yLGn, 9vcQpzdbZmG4kyTP, A1yxMTmKatkf3HRJ, A31fKY79DJCjM2Qz,
A6pfHymTn4qNMZd3, A6pmrCYMfGjWK12z, AMRnwhKV6JvFQYpr, AXjhYkfN13L6HCPv, Ag6nVjXKpcRHkw8y, ApJhHfRkMDqFj3rb,
AvNwTdyBbDncFf2h, BFm724ahJxPNZDVv, BWcP4pDq3xAkKF2r, BXMpQqGY7caChPjZ, BjyVTt6wG4YamQvd, CF7ztDwY8pJ9KZhn,
CMGKZHRVPwzvFmqy, ChbG2MpyJjDWPfV6, CtMF9BgnAmWkTbfd, CwRq8LgZbADvXcHK, CxTmQRX2ytGcwA9r, DBzk8qw4yFWtb9rd,
DH3dT6tYLKmMrNz8, DJCRPcry8mjpb1nH, DW9YkdvyrjtMBQnf, DmJcrQzhtqkLpRaY, FHJjATXtmKkRnPN4, FJdrDz9w1Vxaf2cT,
FkBxY9zn4faM7jWL, Fz2wJmfGKPnvTy9h, GByQ7g9ZPkqxFJXh, GbvNW3tAP68ZpmcY, H7K31X4mQGCRVfnb, Hdm9vZVz1Np8PwAr,
Ht8VFJxmnCG9bQhM, J27rzKkLAndj4m6g, J38xbrFytgM9p1LT, JG89gCcma1Lbxrfn, JLMk78RNzg34wTAH, JNTRK1qgDtkCz4rQ,
JQwbjnVYyGcP4rg8, JaFQcdYDj2pNvbLr, KndmGyYFazPjHftg, KpwNtQh1DXMAWcTm, LCXrjHcgnAJ2yBaZ, MgmdJqxBWA3Tw2vz,
Mh6CHzaPQZc23Tf7, MtxafB13DwLnCFzq, NCfWgvjhqMk8R7td, NTJzVdMZBQDaqG83, NjnA7zPdmZhgCvyL, NrpM7DJQBKWqbzVF,
NtxKh7PgTAaZnHJd, PBLXNA3MhxYWmTKp, PdNq3JxntCKgavkm, Pjd9MWAXgCYTcn7Z, Q2qrkZR6NnbzLmaw, QJkv3RGLCyNBzYxD,
QNKD3HnVZFApMajY, QNwkdxbMvWmrqjAY, Qb1yWxa8vHCGL7Yr, QjHW1hvfbNGgcLBd, QkwyM9RmnV68xdPH, Qr2hVGLaFDydjxNC,
QwJFzfMABRYX9vkL, RBWjPh9tT61q3xFr, RK4bDBcJ28Gnp6Tj, RgMfxbJrY89nTpmq, RrATya1w6ZhY42bB, Rry6WVb2NGD1dXMx,
TcnphZkX2fM9g7Q6, TqDxtfJKXPCLk3zG, VK83G7MdBhLRFAfa, VYBjgJx8aZtRhmpy, Vdj689qcv1F4ZQnh, Vdv8p3ZnTB6fGMQN,
VgjdJkmCTrtYcXLB, VkLX4CZnc68R2Gb1, VqwAHdP4XYm7cxBk, VtcPT63wQ7WaACvj, WFHghVcLRtPpw4MZ, WNDyzRCwd3jhJnkK,
WYTawk8z3LAPpmrx, WZLD8RYn6T9kh2mx, WhPpmnzB6rR3gD4Z, X8PH2CG1hMkcfxYJ, XqQYd9BpCGFmcjJK, Xxdpgy7KNcz1VkwT,
YFtvD87kznr1CThg, YVWgj6ra27PvmNnx, ZCPahyN2JmkYMGLQ, ZcFKV7tLHbDR4qJB, Zt4Rqa6zLWXbFB2m, a7VDZcCkGArwj9Nn,
aZKTvVcCHLwhgXzM, avnFMLCYq89B3KXw, b2HF8pcvrkAGZgdT, b9gAB3nH6qVwpMRC, bF8hVwJNDT6HyRWt, bavD2gBYZHkqK7Ct,
bfmh3jwygC4V9A68, bg169rGdXTzDhLnj, bzwY28XH6Vfc4Zpg, cL3KMWpnXyvDrqa2, cb9gzXjCK3BfH7qM, cgMGCdPJQ84wzaBV,
cgrN7Ywd2fZzHCna, crP79HDfXz2bkZ6M, d7xrp3LnhRQAb4Jq, dBNfgVDLMmnc8h4v, dGh1Fn4rgkjB7Vtq, dKnYzPTqv8JCM21t,
dMxyPfarzTc3tLbk, dnDg7j9kRw6WtbML, dvkzFtPH3WR9C4Km, f3hyv2GtFW4b8wJV, fJnW8D7qYaRXcT2M, gA94CztvFpWYyJ26,
gBDMLv8J1K7nqkpd, gXhWFKATBJRaf3cQ, gmKvNqHFJLZXfdnP, gnZmp1JWfPwBzhC8, h8aVABn6zKPyMwD4, hH4YRGVFz3KnN6bw,
hMv4dJgY3pnyVXTk, hXLT6BWagD32pkZv, hpd6BPmjGbQv8TyV, hzmvKQk4PyY9bjcB, j6XFRzHxTC1kvVpy, jDfKLTzaWywBAZMk,
jWnC8Z7xDvpRcw46, k8NhDFbQjrCPzvn4, k9fKT8ZCtMFn3vab, kGpgavmFADBVRhq7, kH1dt4TpQjfYmaAv, kL1g29NH7MhXvBRQ,
kTKH1qFZYJhWvXVQ, kpjxBPY8dgmGqbZF, m3GbaCBXw1dj6rQL, m4FHwDBYVNPtQhJZ, m7XPvgkLqH1ZnB9Q, mNTMK4J9PQcjCR6L,
mbwzgt3VpLXTj4MZ, mg8qKkahVpLnXZvW, mpgnDQbLWXY8jJP1, n4vQdtMkRTc1zxXV, nK1fCTqXYFJxbMgr, nX1qbatTzGZ3CAwB,
nzWQvwYgca6LBfyj, p9vmyLhPgCf4tN6D, pKJGZBx9wfTdQ1kq, pVNna7r8qH64dcRb, pW8CdNZtPrbQfqGV, pYATGQ6aCmzwBNcM,
pgZM9rYdbH7kz6CN, ptvYCVPymJXH21kR, pyhwdHcqRamYDXzK, qF9L7VZHR1QG8r2m, qFbWHnA6PzydX7Nc, qG68VWwXrPKvfFxh,
qVnt8TGrHXY3B2Q6, qcdyZAr2ganDtxNW, qp6P4GLHwa8mdVjM, r9b1TFDzPk7qjQfc, rJb3LkQVwa2CGNDW, rQpg6YfW7cLRNAVM,
rRnVJhQFa9zf7NqK, rXAdyCnkhqHN1ZDt, rfz4pR6mx1w9JWKA, rjMPpLXcgH31NCWa, rnfRqHt9Vk8g1JB7, t3fny7PvdGMRWhCT,
t9PLr4H8gdkXncRZ, tQ1ymLaFbAT8c6KV, tW38cgRn6b2kr4dJ, tjcwR486HaA7Zr1h, txRbXZwMJm4BW7q6, vGxLrFDkg823p9qK,
vRYfZ3h24XJQpyGt, vbXfBWDQk8jAPTr9, vwCXzhMQ2drgpRxD, wJB7m9F8bc6tgMZx, wfWDyFRcMdr2jZ7K, wz46qKkG19hmBY8T,
xN8bgQ7Gtf4p2dA9, xWpq74Z8Nf63TK1a, xkKRJPqCLzBa9GdH, yGhgd9Zp6WQMaKkf, yHAmqcVTdCnGQp4j, yLFbzXntM8dvQ3YC,
yWFhcNBab8wYMQtG, z9XbfMRh21Fg7QjN, zYPAwt4gc1GRDWVk, zbJQ28WLdkFt9D1G, zmDY4LVX7N8vh9aR, znXGBy6wajHh8bFQ,
zpvfR1gAaKmbWTMy

### Per pet

Each pet counts as one across its fights and reports, by its owner, its name and its creature: the decisions above with
Bite and Claw both castable, the cast within 120 ms of a whole number of 0.8 s steps after the pet's last, in fights
where the pet cast Claw 10 times or more and Bite 3 times or more, so that neither autocast was off. The owners of two
pets cast Bite and Claw by hand: 39% and 15% of those pets' Bites and 51% and 20% of their Claws came off the steps,
and 30% and 12% of their Claws within 50 ms after a cast of the owner's, against 3% to 4% of each for every pet. The
other 237 pets with 40 or more decisions, 63,425 in all, cast Bite at 50.3%; none at 90% or more or at 10% or less,
40% at the 1st percentile and 61% at the 99th. The shares scatter 1.21 times as widely as an even pick would (chi-square
285 on 236 degrees of freedom), and 4 pets fall outside it at p < 0.01, where chance alone puts 2.4. The 38 cats run
from 36% to 57%, 49.0% in all.

Reports (318):

16hJKBmq2V8tnzXR, 16kchwqKn3rpyGMz, 1HDQqzJhamgxkR6X, 1Kmv3YFG9P4DqZJx, 1Qxcn2mJpNHwatA7, 1VRbCQD4MkKtY97G,
1XJvGy4K29CqxtQj, 1h6tv4PBXQgqb8Dw, 1h89JnyWjpZxfXdk, 1z4vRDqGdM2cVPXk, 27jZGhfHgp1LR8MB, 28cLVtZ6D7Wp4jmF,
29PYjXLhFrMkwBvD, 2FN4M7GHLWDmdkRJ, 2R37ZrTWvjHnCJLw, 2wMznDRvmkpbtHxA, 3BAz1TDGL9kjy8ZQ, 3H1fvFaqxRL79PY8,
3LXKpR8kYbWactCm, 3MqmK9ngZvaNDj82, 3R69PBjxZvcfbnX8, 3RB9X8HVxgCJ4My2, 3fFatLkqGnmjQ8ZW, 3nwbdLQXmP4yAhvW,
3qfDVTZ1J29zPChn, 42XtM7Jwdk9gnvqW, 4BtbFKLZ23G67dy8, 4HZTp3WYcdxD1bQV, 4R6dyPjBXmCr3Tq1, 4Xfp9L3bKh2aPYAn,
4hQkw2BmKJGd1XZ3, 4whtjfzW91pbaLv7, 67gJ19PCVB8bXv4r, 6BZ8AkvTNwLJP9tC, 6G9qNRc7KL4bDYjQ, 6PN8t7yDdwrVncJb,
6gnAywbVMGq7vRxZ, 6prM1FBVKTgARmhL, 6yHgkT8LqP42NmxY, 7BraqHpPmfyRjwc6, 7TXcFNRzjwHdAybn, 7ZbkCtYLJ9Aqvhwf,
7dDbtWFhA8BPn4vZ, 7fy93LdcwQhvF2gP, 7kgR1WtDBXJbdAYF, 7tbWfYqJFd4vzL3p, 7xD3CJ4kyjBwdngz, 7yLVTagpRbjYMhnz,
813rPaLtcdpHy67q, 82BgQM3GHJcL7ydt, 8G3K1TCaq4fLXhHz, 8TQJcMxNa2jG7ZLf, 8jzXTaqr732V9DxJ, 97aD8RXPqZjg13Hn,
97wxhXLvAaGKYz18, 9AKatWHpNxnkhBQC, 9TD8g1kVRfZ6yLGn, 9THrFj3VwbzgafAv, 9vcQpzdbZmG4kyTP, A1yxMTmKatkf3HRJ,
A31fKY79DJCjM2Qz, A6pfHymTn4qNMZd3, A6pmrCYMfGjWK12z, AMRnwhKV6JvFQYpr, AXjhYkfN13L6HCPv, Ag6nVjXKpcRHkw8y,
AvNwTdyBbDncFf2h, B16hLXQkryWKFRcb, BFm724ahJxPNZDVv, BWcP4pDq3xAkKF2r, BXMpQqGY7caChPjZ, BfVjHYvWdKkP8Jz7,
BjyVTt6wG4YamQvd, CF7ztDwY8pJ9KZhn, CFNBhYD2Ap6XbVqt, CMGKZHRVPwzvFmqy, CQxWT8KqXjk2NG9J, ChbG2MpyJjDWPfV6,
CtMF9BgnAmWkTbfd, CwRq8LgZbADvXcHK, CxTmQRX2ytGcwA9r, DBzk8qw4yFWtb9rd, DH3dT6tYLKmMrNz8, DJCRPcry8mjpb1nH,
DRf6gP9bac1nrMzH, DW9YkdvyrjtMBQnf, Dbfd6avYpzwTCx3N, DmJcrQzhtqkLpRaY, DvH9VxKZnagAPGL8, DwzYTd1QPHFf6qJZ,
FDRCpv9WTk7VAJjL, FHJjATXtmKkRnPN4, FJdrDz9w1Vxaf2cT, FkBxY9zn4faM7jWL, Fz2wJmfGKPnvTy9h, GByQ7g9ZPkqxFJXh,
GbvNW3tAP68ZpmcY, GnZdPBVXFW6kwp78, H7K31X4mQGCRVfnb, Ht8VFJxmnCG9bQhM, HwyJMCLKfbY1PkmG, J27rzKkLAndj4m6g,
J38xbrFytgM9p1LT, JG89gCcma1Lbxrfn, JLMk78RNzg34wTAH, JNTRK1qgDtkCz4rQ, JQwbjnVYyGcP4rg8, JaFQcdYDj2pNvbLr,
KcdpBWTJyX1xLFfh, KcgXx6na97dmMfGH, KndmGyYFazPjHftg, KpwNtQh1DXMAWcTm, LAH7KPWxNp34nd6a, LCXrjHcgnAJ2yBaZ,
LkYjKP3HpcG7ZDVA, M2Hj9kfNC86rLAWD, Mg2ay1xDjkYpBqRX, MgmdJqxBWA3Tw2vz, Mh6CHzaPQZc23Tf7, MtxafB13DwLnCFzq,
NCfWgvjhqMk8R7td, NTJzVdMZBQDaqG83, NjnA7zPdmZhgCvyL, NrpM7DJQBKWqbzVF, NtxKh7PgTAaZnHJd, PBLXNA3MhxYWmTKp,
PdNq3JxntCKgavkm, Pjd9MWAXgCYTcn7Z, Q2qrkZR6NnbzLmaw, QAgrxLyh1CVwcf6k, QJkv3RGLCyNBzYxD, QNKD3HnVZFApMajY,
QNwkdxbMvWmrqjAY, Qb1yWxa8vHCGL7Yr, QjHW1hvfbNGgcLBd, QkwyM9RmnV68xdPH, Qr2hVGLaFDydjxNC, QwJFzfMABRYX9vkL,
R2GxVLZnrkh1m6Hp, RACFcXrV8YD69j2K, RBWjPh9tT61q3xFr, RK4bDBcJ28Gnp6Tj, Rf3wzmWBt9TAvbX8, RgMfxbJrY89nTpmq,
RgX8kbhG1CdWmYKv, Rhx2TCj8wQN1XFkZ, RrATya1w6ZhY42bB, Rry6WVb2NGD1dXMx, TA3pLxKFXQ9Ckm41, TaCQNVZFp4MX67Yw,
TcLMDwAGYvQk67tC, TcnphZkX2fM9g7Q6, TkxDmPpnjzh3MJ1H, V9k7jdXABJnCwfQZ, VK83G7MdBhLRFAfa, VYBjgJx8aZtRhmpy,
Vdj689qcv1F4ZQnh, Vdv8p3ZnTB6fGMQN, VgjdJkmCTrtYcXLB, VjrAGFkTvNy2W98B, VkLX4CZnc68R2Gb1, VqwAHdP4XYm7cxBk,
VtcPT63wQ7WaACvj, W6AkTfMNjYXxqra9, WFHghVcLRtPpw4MZ, WNDyzRCwd3jhJnkK, WYTawk8z3LAPpmrx, WZLD8RYn6T9kh2mx,
Wa3kfZb1pAncGRBT, WdR6JhPmXq2tk3GA, WhPpmnzB6rR3gD4Z, X1NhLgQrYynPBxWk, X7tBjHNL6WQf8Jbp, X8PH2CG1hMkcfxYJ,
XqQYd9BpCGFmcjJK, Xxdpgy7KNcz1VkwT, YVWgj6ra27PvmNnx, ZCPahyN2JmkYMGLQ, ZV9NHvY6Mf4tkGhD, ZcFKV7tLHbDR4qJB,
ZhbCF1LvkQ9gBAMy, Zt4Rqa6zLWXbFB2m, a7VDZcCkGArwj9Nn, aZKTvVcCHLwhgXzM, ab9tnkPM3D2HvpVh, apMdm4GJH32L1DVC,
atk24BMr1PfjVJRc, avnFMLCYq89B3KXw, b2HF8pcvrkAGZgdT, b9gAB3nH6qVwpMRC, bF8hVwJNDT6HyRWt, bYaT9KXtGwqvV6d4,
bavD2gBYZHkqK7Ct, bfmh3jwygC4V9A68, bg169rGdXTzDhLnj, bzwY28XH6Vfc4Zpg, cL3KMWpnXyvDrqa2, cb9gzXjCK3BfH7qM,
cgrN7Ywd2fZzHCna, crP79HDfXz2bkZ6M, d7xrp3LnhRQAb4Jq, dA2gmBnLHkjM1bxh, dBNfgVDLMmnc8h4v, dGh1Fn4rgkjB7Vtq,
dKnYzPTqv8JCM21t, dMxyPfarzTc3tLbk, dkzHtWm6rvhAZpVN, dnDg7j9kRw6WtbML, dvkzFtPH3WR9C4Km, f3hyv2GtFW4b8wJV,
fJnW8D7qYaRXcT2M, fTBkLPW4MY2zN8AJ, fra8FyNLkPjYVB2M, fzD2LmvpFXCTK7jg, gXhWFKATBJRaf3cQ, gm1dVZJCztNTwq8Q,
gmKvNqHFJLZXfdnP, gmN7Vxy8tPaD1bhZ, gnZmp1JWfPwBzhC8, h3nCT1KLgZwG6PcY, h8aVABn6zKPyMwD4, h8vCa2P43nbH7JwX,
hH4YRGVFz3KnN6bw, hMv4dJgY3pnyVXTk, hTXgqGVtp1abjC92, hpd6BPmjGbQv8TyV, hwJNG8k2a93g4nCK, hzmvKQk4PyY9bjcB,
j6XFRzHxTC1kvVpy, jDfKLTzaWywBAZMk, jWnC8Z7xDvpRcw46, k1jVRK7g9APNTLm6, k9fKT8ZCtMFn3vab, kH1dt4TpQjfYmaAv,
kL1g29NH7MhXvBRQ, kNFDnMJXZcyrpLBd, kTKH1qFZYJhWvXVQ, kmxBzfV3LnR6NDb9, kpGFAabzQhXyd7vn, kpjxBPY8dgmGqbZF,
m3GbaCBXw1dj6rQL, m4FHwDBYVNPtQhJZ, m7XPvgkLqH1ZnB9Q, m8GqNRa9DCgkbVKT, mNTMK4J9PQcjCR6L, mRnQ6xjzVDvw2PMA,
mak9vPD1KCtHW3n4, mbwzgt3VpLXTj4MZ, mg8qKkahVpLnXZvW, mpgnDQbLWXY8jJP1, mpx92zHCYLnTKfAv, mxGpkwbNA9f7vzYH,
myD7KYZ6TzC8aQck, n4vQdtMkRTc1zxXV, nFvHyDjrCxLdA6k7, nK1fCTqXYFJxbMgr, nR1WGzY8pJKDxhQj, nTfhtVKvmw8gZkLj,
nXkPaRgHQ8LGtAyC, nYLth6W2ZbByR1Ga, nbLKdYfXHDam79R6, p7ATmV4HqZRdFwLX, p9vmyLhPgCf4tN6D, pKJGZBx9wfTdQ1kq,
pVNna7r8qH64dcRb, pW8CdNZtPrbQfqGV, pYATGQ6aCmzwBNcM, pgZM9rYdbH7kz6CN, ptvYCVPymJXH21kR, pyhwdHcqRamYDXzK,
qFbWHnA6PzydX7Nc, qG68VWwXrPKvfFxh, qVnt8TGrHXY3B2Q6, qVvZxP3gNp12KWyb, qcdyZAr2ganDtxNW, qp6P4GLHwa8mdVjM,
r9a62VT7RXJCdWNn, r9b1TFDzPk7qjQfc, rJb3LkQVwa2CGNDW, rQpg6YfW7cLRNAVM, rXAdyCnkhqHN1ZDt, rYj7BxLKmAN8t2fv,
rfz4pR6mx1w9JWKA, rjMPpLXcgH31NCWa, rnfRqHt9Vk8g1JB7, t3fny7PvdGMRWhCT, t9PLr4H8gdkXncRZ, tLfvwqZm8jcWBdaF,
tQ1ymLaFbAT8c6KV, tW38cgRn6b2kr4dJ, tjcwR486HaA7Zr1h, txRbXZwMJm4BW7q6, vGxLrFDkg823p9qK, vQZLCzhPNatk7r3d,
vRYfZ3h24XJQpyGt, vbXfBWDQk8jAPTr9, vwCXzhMQ2drgpRxD, wC2pGtAnLyBcNDj7, wJB7m9F8bc6tgMZx, wfWDyFRcMdr2jZ7K,
wz46qKkG19hmBY8T, xN8bgQ7Gtf4p2dA9, xQy7YfRrB2mF3ndw, xkKRJPqCLzBa9GdH, yAYRmxqgXkTwr6Bz, yGhgd9Zp6WQMaKkf,
yHAmqcVTdCnGQp4j, yLFbzXntM8dvQ3YC, yPJ6pm4Q8M9qfLag, yWFhcNBab8wYMQtG, z9XbfMRh21Fg7QjN, zATdNgKZC1kGDqtJ,
zNMDVR1d3kfGCpw7, zXHqKnhPtVAWb7rM, zYPAwt4gc1GRDWVk, zhNraQdpRDX23j7w, znXGBy6wajHh8bFQ, zpvfR1gAaKmbWTMy

### Growl

The 327 pets from 297 reports that cast Growl, every cast, Dash, Dive, Prowl and Cower left out. Growl's row has no
global cooldown; every other ability the pets cast has 1.5 s.

| Cast, then the next one               | Pairs   | 0.6 to 1.0 s apart | 1.4 to 1.8 s apart |
| ------------------------------------- | ------- | ------------------ | ------------------ |
| Another ability, then Growl           | 31,066  | 44.1%              | 33.2%              |
| Growl, then another ability           | 31,195  | 57.3%              | 18.7%              |
| Another ability, then another ability | 150,457 | 6 pairs            | 76.4%              |

Reports (297):

16hJKBmq2V8tnzXR, 16kchwqKn3rpyGMz, 1HDQqzJhamgxkR6X, 1Kmv3YFG9P4DqZJx, 1Qxcn2mJpNHwatA7, 1VRbCQD4MkKtY97G,
1XJvGy4K29CqxtQj, 1Y4NWLka9pAJbB6G, 1h6tv4PBXQgqb8Dw, 1h89JnyWjpZxfXdk, 1z4vRDqGdM2cVPXk, 2R37ZrTWvjHnCJLw,
2YTQxaqzvAZPwKW3, 2wMznDRvmkpbtHxA, 3H1fvFaqxRL79PY8, 3MqmK9ngZvaNDj82, 3RB9X8HVxgCJ4My2, 3qR2PmyNvQk8J4Bc,
4J7dhZxb3GakBQ8p, 4R6dyPjBXmCr3Tq1, 4Xfp9L3bKh2aPYAn, 4jRc6D2M1XpgCdHV, 4whtjfzW91pbaLv7, 67gJ19PCVB8bXv4r,
69j8PA1zfKvVmpkM, 6BZ8AkvTNwLJP9tC, 6G9qNRc7KL4bDYjQ, 6WG74An9RrCdqkwD, 6prM1FBVKTgARmhL, 7BraqHpPmfyRjwc6,
7FBaHRhMVKjC1QbT, 7Q6WxzYRwjNqXvgZ, 7fz9Kk8rY6jyQAbL, 7kgR1WtDBXJbdAYF, 7tbWfYqJFd4vzL3p, 7xD3CJ4kyjBwdngz,
7yLVTagpRbjYMhnz, 8ALbrqD9MNJwF1P4, 8TQJcMxNa2jG7ZLf, 8YfKRvzTxHCVpGtQ, 8jzXTaqr732V9DxJ, 8rAWbkX93chqKjFG,
8zC3dW71xXrgaA2v, 97aD8RXPqZjg13Hn, 97wxhXLvAaGKYz18, 9AKatWHpNxnkhBQC, 9CgtkpAYqGjxP34Q, 9H2chVMdn1ZtxXpA,
9THrFj3VwbzgafAv, A1yxMTmKatkf3HRJ, A31fKY79DJCjM2Qz, A6pmrCYMfGjWK12z, A7ahGYbzQ8M1ZtF6, A7kG1nHC9aY6yRJP,
AMRnwhKV6JvFQYpr, AdfMXDLW29k6pGy1, Ag6nVjXKpcRHkw8y, B16hLXQkryWKFRcb, BFm724ahJxPNZDVv, BVJg8aWCbQmM6hfp,
BfVjHYvWdKkP8Jz7, CF7ztDwY8pJ9KZhn, CFNBhYD2Ap6XbVqt, CH4GyNxfn1Lg9a2V, CKL6q7rTtgYWDXk2, CQxWT8KqXjk2NG9J,
ChbG2MpyJjDWPfV6, CtMF9BgnAmWkTbfd, DBRqygp13jPC6c82, DH3dT6tYLKmMrNz8, DRf6gP9bac1nrMzH, Dbfd6avYpzwTCx3N,
DcXfK2MmZgyR6wpq, DvH9VxKZnagAPGL8, DwzYTd1QPHFf6qJZ, FDRCpv9WTk7VAJjL, FHqYCNcKntvDAjxg, FTQ8g6cvbyZ1wB3m,
Fz2wJmfGKPnvTy9h, GByQ7g9ZPkqxFJXh, GRPFkVj2bKhD4Npz, GRnMHmK69BXfWVTj, GaKFhmvgVN4cCWQJ, GbvNW3tAP68ZpmcY,
Gkb3rKHF6V9gz4pR, GnZdPBVXFW6kwp78, H7K31X4mQGCRVfnb, HdAz2cWg6B417rp9, HqyPGnV9xLmtMK2D, Ht8VFJxmnCG9bQhM,
HwyJMCLKfbY1PkmG, HxCQat38vmjnkpb6, JDNPdfVbtxTmH42r, JG89gCcma1Lbxrfn, JQwbjnVYyGcP4rg8, Jjmv6drycNXpDxW8,
JnVckZDLTwAgGKaQ, JrqDZX4LT6NpxzM7, KG62VXF3a97CxtAg, KcdpBWTJyX1xLFfh, KcgXx6na97dmMfGH, KktaH16Yrd9f7hWD,
KndmGyYFazPjHftg, KpwNtQh1DXMAWcTm, KqZ6xWAfDLGrVjYJ, KyRzjkdcPghabTFJ, LAH7KPWxNp34nd6a, LCXrjHcgnAJ2yBaZ,
LCqYwgcpM4N9vJtb, LTGbja6krzZYxcg1, LahKC3yT6JQRwtMn, LkYjKP3HpcG7ZDVA, M2Hj9kfNC86rLAWD, M7kjDhzcr3RNgXQx,
MWrPDAVFCBXzh8Z6, Mg2ay1xDjkYpBqRX, MgmdJqxBWA3Tw2vz, MtxafB13DwLnCFzq, MypmYhB3RX9FJ7Kx, NRkpzQVnYtrjmZaT,
NjWKPpHgymCT832A, NtxKh7PgTAaZnHJd, PBLXNA3MhxYWmTKp, Pjd9MWAXgCYTcn7Z, Q2qrkZR6NnbzLmaw, Q7V24RA9HTg6F38p,
QAgrxLyh1CVwcf6k, QCRFhVqmHWMK4Gv7, QHXdBJRj2GPpC3n1, QJkv3RGLCyNBzYxD, QNKD3HnVZFApMajY, QRCgnhck3KVq6vBT,
QXFd74zDW8JZMmKx, Qm9XJwGdqNzhRc42, R2GxVLZnrkh1m6Hp, R78CFp1D24jAatwZ, R92pdBxqLDrnMV8C, RACFcXrV8YD69j2K,
RJXQC67jbzATWdx1, RQFBGJLVmbYynrTh, RY1yA2PFCBwb4MZa, Rf3wzmWBt9TAvbX8, Rhx2TCj8wQN1XFkZ, Rry6WVb2NGD1dXMx,
TA3pLxKFXQ9Ckm41, TG7ZMLJB1cYAhq4k, TaCQNVZFp4MX67Yw, TcLMDwAGYvQk67tC, TkxDmPpnjzh3MJ1H, V1jLfaQtnGdHPB2z,
V9k7jdXABJnCwfQZ, VLz78BfZyWa3Cxr2, VNpDMaKXy2LHAPZh, VYBjgJx8aZtRhmpy, VjrAGFkTvNy2W98B, VkLX4CZnc68R2Gb1,
W6AkTfMNjYXxqra9, WYTawk8z3LAPpmrx, WZLD8RYn6T9kh2mx, WdR6JhPmXq2tk3GA, X1NhLgQrYynPBxWk, XwWTF6ZLcQnYNxH3,
Xxdpgy7KNcz1VkwT, YH7n2xZqPKzWjQrR, YZ6NdFDRQKL2CTmv, ZCPahyN2JmkYMGLQ, ZJhXYVjywzrDcnWv, ZV9NHvY6Mf4tkGhD,
ZhbCF1LvkQ9gBAMy, Zt4Rqa6zLWXbFB2m, a2MWRNwFLTPHAzD4, ab9tnkPM3D2HvpVh, apMdm4GJH32L1DVC, atk24BMr1PfjVJRc,
b2HF8pcvrkAGZgdT, b9gAB3nH6qVwpMRC, bRfPGLA3Qp4Dn6xa, bYaT9KXtGwqvV6d4, bavD2gBYZHkqK7Ct, bfmh3jwygC4V9A68,
bzwY28XH6Vfc4Zpg, c8R4bAnHpmFTzaG6, cf9LwRYktpgvZdH8, cgMGCdPJQ84wzaBV, cgrN7Ywd2fZzHCna, cr6qMtj1vCTPHVfw,
d1vn7jt4RH2ZmLAB, dA2gmBnLHkjM1bxh, dGh1Fn4rgkjB7Vtq, dKnYzPTqv8JCM21t, dMxyPfarzTc3tLbk, dkzHtWm6rvhAZpVN,
dxZCXJvPLKAB3cnQ, f18RG4BAVbKTLMwD, fNbLwMvKCrHc23gG, fNxhtGQYCBnPbzcp, fTBkLPW4MY2zN8AJ, fcnqgCZtXyY34Gbm,
fn2Dc941K63CNvmj, fra8FyNLkPjYVB2M, fzD2LmvpFXCTK7jg, gBDMLv8J1K7nqkpd, gRmxdjCHt6XDMBpa, gVF7L1nvY3bCqwcz,
gXhWFKATBJRaf3cQ, gmN7Vxy8tPaD1bhZ, gpXqDYBNbxKc1y7W, gtNvbhF9aZCXyzVk, h3nCT1KLgZwG6PcY, h8vCa2P43nbH7JwX,
hXLT6BWagD32pkZv, hmDYk8MKdpqx1FzJ, hpd6BPmjGbQv8TyV, hwJNG8k2a93g4nCK, jDfKLTzaWywBAZMk, k1jVRK7g9APNTLm6,
kH1dt4TpQjfYmaAv, kL1g29NH7MhXvBRQ, kNFDnMJXZcyrpLBd, kTKH1qFZYJhWvXVQ, kXAF71Wfy6Tx2hZd, kaH1G4ztmYpgMTJ6,
kmxBzfV3LnR6NDb9, kpjxBPY8dgmGqbZF, mLhwHyCBfVx2vgZc, mRnQ6xjzVDvw2PMA, mV7NcC8vQ3jkTHwB, mak9vPD1KCtHW3n4,
mb8FgjhwVTLZ4q1d, mg8qKkahVpLnXZvW, mhbXcFqf9xTaA1NP, mpx92zHCYLnTKfAv, mxGpkwbNA9f7vzYH, myD7KYZ6TzC8aQck,
nFvHyDjrCxLdA6k7, nK1fCTqXYFJxbMgr, nR1WGzY8pJKDxhQj, nTfhtVKvmw8gZkLj, nYLth6W2ZbByR1Ga, nbLKdYfXHDam79R6,
nyMqbY64tL7HhKQJ, nzWQvwYgca6LBfyj, p7ATmV4HqZRdFwLX, pNW7aGKnzHP1MftT, pWnPwNFGgyX7tkdq, phmHcBvV39RMtjqF,
ptvYCVPymJXH21kR, pwzCJQZXDcLT39g4, qJvgVnr8c9kZ2DTL, qNVWc2yZJT6k8mvh, qVnt8TGrHXY3B2Q6, qVvZxP3gNp12KWyb,
qcdyZAr2ganDtxNW, qrwtb1GNA8W7L6Xm, r9a62VT7RXJCdWNn, r9b1TFDzPk7qjQfc, rYj7BxLKmAN8t2fv, rjMPpLXcgH31NCWa,
rtPp1Bd6L8yQZCNh, ryM1qhHFwc3YaDzG, t9PLr4H8gdkXncRZ, tLfvwqZm8jcWBdaF, tMpK7wQXD6WfFcyN, v23ytJHDz6QkbXnN,
v8L6NkQTzMV2mWq3, vPHb3a6c1KhknLrY, vQZLCzhPNatk7r3d, vRYfZ3h24XJQpyGt, vT1rMa2cNGXLfqp7, vbXfBWDQk8jAPTr9,
vg1kL4rN2KZVCaRd, vwCXzhMQ2drgpRxD, wC2pGtAnLyBcNDj7, wFQXdZknKhAvfyaz, wJB7m9F8bc6tgMZx, wc9dvnta1bz8K3QP,
wz46qKkG19hmBY8T, x9v27CNzhRgMkXG6, xN8bgQ7Gtf4p2dA9, xQy7YfRrB2mF3ndw, xWpq74Z8Nf63TK1a, xbJBZj14NWCXrhAm,
xfWNbt8yZKzcRPjC, xhtLrB9MjHqWvVY7, yAYRmxqgXkTwr6Bz, yHAmqcVTdCnGQp4j, yLFbzXntM8dvQ3YC, yPJ6pm4Q8M9qfLag,
z9PqXTpDyVG3Fvn2, zATdNgKZC1kGDqtJ, zNMDVR1d3kfGCpw7, zgqYjRFPMTV2LA31, zhNraQdpRDX23j7w, zkQabTrD4RHLZN8G,
zmTNyXkfqtpgcM1H, znXGBy6wajHh8bFQ, zp9yVrtfTNZJCv7g

### Dust Cloud

The 6 tallstriders from 6 reports, every Dust Cloud they cast, 2,211, with enemies told apart by instance. 23 went onto
an enemy that had held the pet's Dust Cloud for 100 ms or more; 1,091 went onto another enemy than the one the pet
last swung at while that one held it, and 443 onto another enemy while it did not. At the decisions counted as above,
Dust Cloud and another ability castable and the pet's swing target without the debuff, the pet cast Dust Cloud 673
times against an even share of 639. 44 casts were dodged, parried or missed.

Reports (6):

6prM1FBVKTgARmhL, KktaH16Yrd9f7hWD, TqDxtfJKXPCLk3zG, nX1qbatTzGZ3CAwB, xWpq74Z8Nf63TK1a, zbJQ28WLdkFt9D1G

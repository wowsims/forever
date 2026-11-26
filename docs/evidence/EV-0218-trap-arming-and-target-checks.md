# EV-0218: A placed trap's arming time and its checks for a target

- Claim: a placed trap arms 0.8 s after its placement and then checks for a target every 0.4 s, its first check falling
  evenly within the 0.4 s after it arms; about one trap in four springs a check later.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), 531 reports of the hunters on the Forever servers, all
  fights; the codes are listed below
- Game build: 1.60.1.70862 to 1.60.1.70938 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC); the hotfix each player ran and whether they relogged are not
  recorded
- Date: 2026-11-06 to 2026-11-17
- Pull request: #1252

## Method

One reading is one placement: from the hunter's cast of a trap's placement spell, which logs the trap's creation at the
same timestamp, to the first event of that trap's effect. That is Explosive Trap's first hit that is not a tick, of any
outcome; Immolation and Freezing Trap's first debuff or miss; Ice Trap's first debuff. Nothing counts past the next
placement of the same trap or 60 s. The trap lies at the hunter's position on the cast.

A placement counts only when the trap's first target was within 4 yards of the trap at the placement, by its last
position logged at most 1.5 s before the cast, and had moved less than 1 yard by the hit. Such a trap waited on no
target walking in, so its delay is the arming time and the wait for a check. Positions are the x and y logged on
events. Hunters of every level count.

## Delay to the first effect

| Trap            | Placements | Before 0.80 s | 0.80 to 1.22 s | 1.22 to 1.64 s | After 1.64 s | Later window | Mean, 0.80 to 1.64 s | Median  |
| --------------- | ---------- | ------------- | -------------- | -------------- | ------------ | ------------ | -------------------- | ------- |
| Explosive Trap  | 3,303      | 9             | 2,609          | 681            | 4            | 20.7%        | 1.099 s              | 1.068 s |
| Immolation Trap | 1,841      | 3             | 1,270          | 560            | 8            | 30.6%        | 1.144 s              | 1.114 s |
| Ice Trap        | 508        | 1             | 388            | 109            | 10           | 21.9%        | 1.110 s              | 1.085 s |
| Freezing Trap   | 105        | 0             | 82             | 22             | 1            | 21.2%        | 1.077 s              | 1.027 s |
| Every trap      | 5,757      | 13            | 4,349          | 1,372          | 23           | 24.0%        | 1.114 s              | 1.084 s |

"Later window" is the share of the placements in the two windows that fall in the second. Within each window the
delays spread evenly: the first's quartiles are 0.92, 1.02 and 1.12 s, the second's 1.30, 1.41 and 1.52 s. The target's
distance from the trap changes little: 18.1% fall in the later window with the target 0 to 1 yard from it (323
placements), 26.0% at 3 to 4 yards (1,982). Immolation Trap springs a check later more often than the other traps.

Reports (531):

123jYCvkVGrHfXWt, 16hJKBmq2V8tnzXR, 1HTPQGZya7pAz3qh, 1VpHn3kYq6B8JMWw, 1ZyhaVDdFtTYWMcC, 1a3Qn7JpLk8AH2hV,
1aBhkzdDMZKGtb37, 1b7PdYR4HhDXBGT6, 1h89JnyWjpZxfXdk, 1xzPrQWDC4pv9HhJ, 1z4vRDqGdM2cVPXk, 1zdqMb3QnryZL2j7,
27jZGhfHgp1LR8MB, 28vgnypFaHmVkqDL, 2ACDTKkcpWbNfRJ9, 2CcwRgf9MhVLaPJH, 2DvWNfwZHqgXdGLM, 2YTQxaqzvAZPwKW3,
2YfdbFMgGhr7zHwQ, 2Z4G3r9pfqJtCM6W, 2gKNqpY4G8VdtLMf, 2kV73aphdFrCQq9D, 2wMznDRvmkpbtHxA, 31TtLPDXrczfGwAC,
3ABMTH9n8rqRb71z, 3B8AcaGNRKmdgPDj, 3C26abGgdcQKATJt, 3H1fvFaqxRL79PY8, 3R69PBjxZvcfbnX8, 3fFatLkqGnmjQ8ZW,
3grN8Rct9Xzknm1H, 3kwcDX7JjVPdT6Mf, 3qfDVTZ1J29zPChn, 3wdx94CkMQVGmayv, 4BtbFKLZ23G67dy8, 4CVrTcXtBf2yAYH9,
4L7DBHTV2jA6gkQW, 4QyZD9TnbLdvGzKp, 4Tc3yBNVYn7wW8RJ, 4Xfp9L3bKh2aPYAn, 4YbZc7VkWQ8HzvqF, 4dakVPwbnc6QKjvh,
4nkW7JDxV12waMfr, 63pnYPtNJk1X7HFB, 6AvVhawr1QjRxWBz, 6DctnwkGNJmMVXhW, 6Jn1dFwNQCkgYLqV, 6WG74An9RrCdqkwD,
6gnAywbVMGq7vRxZ, 6vWJarDyMYxpAt8T, 6yHgkT8LqP42NmxY, 74YybNVPDg1MpJG8, 7FBaHRhMVKjC1QbT, 7Q6WxzYRwjNqXvgZ,
7QtfGD3ArVhH8qdw, 7cpTkjaFB3mz6VgM, 7d9yw6VNfWhLTkFQ, 7dDbtWFhA8BPn4vZ, 7fy93LdcwQhvF2gP, 7fz9Kk8rY6jyQAbL,
7nDgKB3dMbW92HwR, 7pmhcYx1JNCKvWjn, 7qDZP6zcXy9G2Qvt, 8G3K1TCaq4fLXhHz, 8YfKRvzTxHCVpGtQ, 8jzXTaqr732V9DxJ,
8rAWbkX93chqKjFG, 97aD8RXPqZjg13Hn, 97pV3KgNWFkqaGtd, 9H2chVMdn1ZtxXpA, 9JxanDLbZqC18YPg, 9VDYZBhfkCwdqb86,
9YZPHyRaK3FrQkNT, 9Z6yrfgDLjPahGAV, 9Zdv4NLm3FAwQKaV, 9gRDVvwAQnC41PBc, 9p6xLAGZ2Cc7Vr3f, A1yxMTmKatkf3HRJ,
A7ahGYbzQ8M1ZtF6, A9JQgmPDf8TycqMB, AFf3nqybHJkzQG1x, AXn7cJ9VQ6YHLWmh, AdfMXDLW29k6pGy1, ApJhHfRkMDqFj3rb,
AvNwTdyBbDncFf2h, AwJbDaHfdx289Khv, BJvAYckHpKt2LQ7F, BR2gLD1wPbVvQ3HX, BVJLw6vAjg8rhZFt, BVgtPkXmAcJwM21Q,
BXMpQqGY7caChPjZ, BYyqMbfxpZ8nFhJg, Bcfrn3xbXV8CMzaR, Bg7J4jxckP6pdvrZ, BnPXHVWwTrCYNtMh, C3r6TRtDnxzVPF9j,
CJKLxmXbMq3F62ht, CLBtcNZgqYb8zQhd, CXvjmgaJ6YqKzD4k, CbVw9y3XZPnLvzR8, CrJY1g8HjTkc3dyX, CtMF9BgnAmWkTbfd,
CxTmQRX2ytGcwA9r, CyN68LDZPxkFtQYB, CykdWTB4ZpG2VXQF, DBRqygp13jPC6c82, DH3dT6tYLKmMrNz8, DJCRPcry8mjpb1nH,
DMv2cWLFgT13AyRG, DRgLHWcytG7NvPjM, DW9YkdvyrjtMBQnf, DWfMg7hzPF4d1mQ6, Daj7G8xFX3bA2CmW, DqkcQyp7HZWMGAz2,
DwzYTd1QPHFf6qJZ, Dxv3XZjp2nmwkWQB, F1Pjx7JKnGMYCr2R, F1qzrfMnGYy9Vb7j, F2vN8aVjBz1PJ4Q3, FHqYCNcKntvDAjxg,
FN8krKfCb3BXWmwV, FNyng7tvQ1T2bra8, FXCcfp4Y8ZwjAvgJ, FXcLvZV8nxQT96fd, FjVGh46kcNQmn3tZ, FkBxY9zn4faM7jWL,
Fz2wJmfGKPnvTy9h, G1JF39DABcrVvTHC, G4akXpft3gNhCJ19, GNp7bXWgfjr3yAcm, GRnMHmK69BXfWVTj, GZdbRHQ146gjMVAy,
Gbkh7r8FVnCPYgpy, GbvNW3tAP68ZpmcY, GdWXA3pmjFy8RLwz, Gkb3rKHF6V9gz4pR, GnrXRzQ2THKJ369V, HL38fnFjNXGTmRCa,
HPnctDh3axdKw1MV, HdAz2cWg6B417rp9, HmF9nyBdrvAL1wKt, HmRZkYcPafqK3rMv, Hwq97xLh6gv8ZPrY, HxCQat38vmjnkpb6,
J16aZwyRBNXT3QbK, J38xbrFytgM9p1LT, J4nGFVKWD1Aygkaj, JBxW16RqGXMPYH3d, JC6v7hNYf3XrFgHb, JDNPdfVbtxTmH42r,
JDtYTwZ12c864k7a, JLfYA2drVCZ3aMnv, JM7zNDKTxPfm2LV9, JNTRK1qgDtkCz4rQ, JQgB3aWcjTDV9nkM, JRQqNL6DkvWKYdPB,
JVy7a8hrgRMZv6Xz, Jcx8g4CXq1QBpzVP, Jd6WanvfBV4M2zAk, Jjmv6drycNXpDxW8, JmWKjCcgfXv4GVHP, JnVckZDLTwAgGKaQ,
JrqDZX4LT6NpxzM7, Jyw9hD1bmKWCN3XH, KG62VXF3a97CxtAg, KNAaDTmGZ69CFkjH, KPwXNByAG2xQm9ra, KX28wFnvJBak374t,
KYvNjyn1p2Lb7M4r, KcdpBWTJyX1xLFfh, KfgYa2LbdCRGMhXv, KjdWqMAvFVN7831w, KndmGyYFazPjHftg, KqZ6xWAfDLGrVjYJ,
KzwZMvnPtXcF1p3f, L6GvARYBpNQy1hk3, LCXrjHcgnAJ2yBaZ, LDngCPfhQcKpqTyN, LFNZrg2Cpty63BnK, LQDbHMzkYc1T2a4W,
LR9WfMTdbr3ZNBFH, LRqc8tNbC1Hf7Vwn, LZYnpwV6Jxdr4WaH, LhByjkW9tmKAaMZ6, LjvDknrWTVP1xAaz, Lxy1kqgVvFc24aPr,
M4aF6w9Pz7JmrXgW, MHvXc1w9jCdATF76, MWrPDAVFCBXzh8Z6, Mg2ay1xDjkYpBqRX, MjfT3pQKRWcVG4gY, Mw6q32gCaKB8Hjnf,
MypmYhB3RX9FJ7Kx, MzwcDrG3fdFBRy6a, NCHVzbKtjGYpaJZW, NHLVxbT9hgPZnGwp, NKM174PyfxjTXRv3, NRJG9M1wF7ng4BdC,
NRK61HZcDdJkrwXG, NRvfwJbCyQ3qh4dj, NXMZV9xjvWpB7GtP, NbmjagBQMYLC8t6z, Nby7FHqDhWcBLYXd, P2QkCnLwyB1cVA9z,
P4Nv9AfZa8j1DdGr, P6GLNM3n27dtcYJD, PA1XWx2zbVgnHQKa, PB6zGcVNtTZrMKXb, PDFVdp172Kwz8aqj, PGVWZxJDC4AMQb8L,
PLGKvft3bq8WgRYN, PdNq3JxntCKgavkm, Pjd9MWAXgCYTcn7Z, Pn7j3yhC1vFDKcNM, Ppd7C9MWXFkr2Tga, PtmYHjkDrqhdR9MA,
Q2qrkZR6NnbzLmaw, Q4Z9KNk1jhJpMx6R, Q7V24RA9HTg6F38p, QADVqHgcjLyvBWP6, QAgrxLyh1CVwcf6k, QDadybK3XVPRvwLh,
QF2qpKC9rbxV3ZvW, QHXdBJRj2GPpC3n1, QNwkdxbMvWmrqjAY, QRCgnhck3KVq6vBT, QXFd74zDW8JZMmKx, QYbwBT6tkmNWa1Gh,
QZ4wdChqH7WXBDPm, Qb1yWxa8vHCGL7Yr, QbM9kPFqycCTgzVx, QcrkDWdPGqZmv6jf, QjHW1hvfbNGgcLBd, QmNYMtxwG7Py9Ajv,
QqbhvMX12Y3ZPBVG, Qr9PmCnptRWMjN7k, QvDAW3czJBVa6Ngx, R1g7q2JAnwzdrPvc, R8wpcTxFbDPz37gr, R9fbLyJAW1tDQnFT,
RGnyYzMgqCNdVwkT, RVgGFynN3KCYJHcA, RY1yA2PFCBwb4MZa, Rhx2TCj8wQN1XFkZ, T1ZpdB9NAkLP76vG, TA3pLxKFXQ9Ckm41,
TG7ZMLJB1cYAhq4k, TH81CKfzvmZVBbNQ, TNAXzyxRPQZW3pq2, TcLMDwAGYvQk67tC, Tfp1v6KhDwJntrRd, TxKGHQ1tAcqP9NYp,
TxX4AqCgafhyn6zc, V6YjPXcQLfD2FBZk, V8CWdb79mMRDLB3t, VGwz9XY47JAvChQy, VK83G7MdBhLRFAfa, VLz78BfZyWa3Cxr2,
VNpDMaKXy2LHAPZh, VYBjgJx8aZtRhmpy, Vdv8p3ZnTB6fGMQN, VhqB8wZJzRP34cAM, VjrAGFkTvNy2W98B, VkLX4CZnc68R2Gb1,
VtcPT63wQ7WaACvj, VtjvpP8bwcgCYTXF, Vw3yrtjXgAhcx94m, W2GLAPRX7jw8BC3n, WBVwmaXxh1JfPvKC, WGgVMc3B69TrZwzN,
WMy9axzfg3jKLhYc, WQ3fqPhAVRrcn7Bb, WaHFDqN2c37hGx6v, Wy6TdamAk4ZxJ8MC, X6Hp4gDwv7zftxMC, X8PH2CG1hMkcfxYJ,
XBbarKyLTcJfNxj3, XFdDMhAQVxYtjWB8, XHk7W3h1NDbpL4PV, XMxrZ4ByKVd16Wkn, XPYV9J6DfrMKxGqh, XQkavwVbDrLKmhJ9,
XWfw2PzpNcBhbayY, XY4mQxn86JHyarz1, XdFYRDafB7cjhgJb, Xv2QmnyK8RBxW9jZ, XvhwZD31Vz46JWG8, Xwxna4t6pmvYFAyH,
YCm7yQkZM2wvTaP1, YFtvD87kznr1CThg, YH7n2xZqPKzWjQrR, YL6pQKrq7ZXCbHdV, YLNxvdDn96TjA3R8, YVWgj6ra27PvmNnx,
YZ6NdFDRQKL2CTmv, YaBQZ2tbVFJMrAhk, YakBLWFpjnrdhGb3, ZC3cMxKjYRpVP7L9, ZDRTLNaKyw8gJdcM, ZDWyQGqFBvagT7kA,
ZHpKVnhPDGyaTrvw, ZaFD24rvwRc7yd8b, Zcjan4PWJq3GLwYT, ZfbY6RgWjx9hmpGz, a4ZdvYKyhcLGnFRP, aDzbqXNJgCVBAy8c,
aFq1fMXx8mZ6H7pQ, aLYv7r1kDFpxXJVG, aMF1JcGxkDpjnYH9, aRkHpWzNmfVGqAth, aVX9hAZ2ryGCwqKz, aZKTvVcCHLwhgXzM,
apNB26JmYPyhFWjH, b1jA7LXFrZfmGMCR, b6nYARMxa2ZGfX9Q, bDjF6xfrQ8Zz9nL7, bDwx9gzy4F7Np1f8, bF8hVwJNDT6HyRWt,
bKJcjWDRzvHpVmwa, bYaT9KXtGwqvV6d4, bavD2gBYZHkqK7Ct, bg169rGdXTzDhLnj, bmZcqx1pW4LGFfKB, btAc341MaWNQH8RL,
byjaGMnmJRtLpH2g, bzwY28XH6Vfc4Zpg, c14qPbJdkhX6a3vK, cATBb6FrVznvMPGk, cDxNAW6Zr3hvC2QG, cGYvJnm1TfNVxqgA,
cGrfRLdNmqyKZpFP, cYjJMvzXhQKaGkFL, cYxhqgdnyRtwGb93, cb9gzXjCK3BfH7qM, cgMGCdPJQ84wzaBV, d2FLqCfQJV4XT1gD,
d7RvpH8nJtGXBxc6, d8pgmJxYh73cbyj6, dA2gmBnLHkjM1bxh, dBNfgVDLMmnc8h4v, dFcXZkRP9qnLb7mT, dNT6DXyhHFRjZLaq,
dP8LDZ7z4tTk3xA9, dcpLCfkBjhRq1FPw, dnDg7j9kRw6WtbML, dvkzFtPH3WR9C4Km, dwjtaQyAR98XB1nF, dxZCXJvPLKAB3cnQ,
f36RaB12wyDcTCGY, f3hyv2GtFW4b8wJV, f8TKwNhrJqyBjx73, fD1aRkzx4jYV3drh, fJzHPMRwL2mQgGVB, fLKYjJNVRCgxcwQT,
fNbLwMvKCrHc23gG, fVaTmnXZvPcA7gxy, fWF83zPgHX1RpLkv, fYdRHbPy4nA2k8Mq, fjtT3JVxbGPz6RwX, fn2Dc941K63CNvmj,
g3dYGjzy19w2qfXB, gA94CztvFpWYyJ26, gANGd1jMVfQ9r8xP, gAmv8kd3JtZcbyLP, gBDMLv8J1K7nqkpd, gPxrZVY4JQ2Fm1Nd,
gRmxdjCHt6XDMBpa, gRtBKzxVFjMJwQCm, gmKvNqHFJLZXfdnP, grBtAJ9vRTk2XGjx, gtNvbhF9aZCXyzVk, gxC3BP2NRYLGt7A9,
gzrAwTqCWPm7xvpD, h29zDjKXGxvJfpkH, h8aVABn6zKPyMwD4, h8vCa2P43nbH7JwX, hAb6yN8xnqBpWj3D, hH4YRGVFz3KnN6bw,
hLxABfWVavkM8HGT, hLzm4tKQAc8DCMpR, hPNgqjA4vkTxXwG1, hQq8gvGjzPcWKF9A, hmDYk8MKdpqx1FzJ, htxML9PnY2q6BHWV,
j6XFRzHxTC1kvVpy, jWnC8Z7xDvpRcw46, jXhdfNz2g7WpBYCa, jYVvXCw4QANH1Ffn, jryFckzCJax34KQb, jy4gRbHKGWXVP29T,
k1gAvx7zVaPpLhYC, k1jVRK7g9APNTLm6, k2CRPKTymB6rqdav, k7gmfZ69awYr3Dct, k8NhDFbQjrCPzvn4, kGpgavmFADBVRhq7,
kH1dt4TpQjfYmaAv, kJfv4pHtm26FjCxn, kW9V4CMKfmBbtrRH, mCYWgaFVAdrby9NX, mJDNgqYVGwb6A8tF, mRnQ6xjzVDvw2PMA,
mXvcxNnBrtgGyq8b, mak9vPD1KCtHW3n4, mb8FgjhwVTLZ4q1d, mbZcgPavhdG7fKDW, mbwzgt3VpLXTj4MZ, mbxRn8FNZk6vH329,
mhbXcFqf9xTaA1NP, mnPXjHcWNFKBhDCq, mpXk9g8PzctZNJda, mpgnDQbLWXY8jJP1, mqvYk6HbFaGLTWQ2, mxbK4d1k96GzwVRh,
mzB4Wj3q6Ktv12nG, n29MtQFKRzqDC3xN, nDNT9zhcQ8trJbwf, nFvHyDjrCxLdA6k7, nJDNV6Md3bqpXZT9, nR1WGzY8pJKDxhQj,
nTfhtVKvmw8gZkLj, nX1qbatTzGZ3CAwB, naJFzXZPWbR9QAgt, nkrMNqLAJczYx269, nr4qJHp3tQaxzyTW, nyMqbY64tL7HhKQJ,
p7ATmV4HqZRdFwLX, pAKLzByFXYPravgT, pNKJR39CwGnmXtbL, pNW7aGKnzHP1MftT, pQbkWJAnqFgHxDzV, pQdVgvJwFZNjfaMG,
pWnPwNFGgyX7tkdq, pdkwZJNqWKCarQx9, pmrFXf1ZY9N6qnWh, pwzCJQZXDcLT39g4, q7MvX6Vd42jLrbF8, q7hWLNjgfknX94Zy,
qGywkYrdHKT8NgAQ, qNVWc2yZJT6k8mvh, qQgj2MWBcD4Gd9an, qVnt8TGrHXY3B2Q6, qVvZxP3gNp12KWyb, qa3hCjJ1w9yzgZYc,
qhRXQ43wNmkcPLB1, qrwtb1GNA8W7L6Xm, qvbXATzmk6Gha1Ny, r3an7RhCT4NjVyfb, r9b1TFDzPk7qjQfc, rAFM4QTbC8jHypZL,
rB41vNC2npYbz7Zy, rC4yxgXTdHkQcB2m, rJb3LkQVwa2CGNDW, rRnVJhQFa9zf7NqK, rWa1qLwjXxK42Ffd, rWjb2mNRfwdM7pq1,
raBGY1z29kVctF37, rf8j1q4Dw2pGVMRJ, rfg8k3Nw1P7TAHx6, rfkDTHnR8CBAwhz4, rnfRqHt9Vk8g1JB7, t3fny7PvdGMRWhCT,
tCzP4jWxdNLHcgp9, tVWcPZ1YDMNFzdXr, tbC9yvQdpY7FnXmJ, tjcwR486HaA7Zr1h, tkHwWYF1C8LTfhQg, tkPZYqTRgCyWG3jM,
tmHdvqxYj1z28PFy, v23ytJHDz6QkbXnN, v41rYANj2zBWkb8n, v84Y93dCaQrGbwcy, v8JbfQNY1RpDFXqh, v8L6NkQTzMV2mWq3,
v9CxPFNqVjnbkZM8, vAMz47GFTwqpXgQf, vFQhXgw42KmfjT8W, vGFaKgLT9mkn1Jdy, vGJMahXtgTzFRmCQ, vLXN64YjZCrDWfVF,
vPHb3a6c1KhknLrY, vQZLCzhPNatk7r3d, vatW8VJnyX3GmMKC, vg1kL4rN2KZVCaRd, vgVRZY9ykdzPWBMb, vkrZLbmDRx3ydpHz,
vpFrJRb71fNAatmL, vwCXzhMQ2drgpRxD, wJCmzRy4GHgDTA8F, wPYdtLC7f8Xbc2qy, wZnqkKFbjRNgc7Md, wZrvyFPjMGg7dt6R,
wm9ZjJLCbavznpHA, wptKdcMH8fWY4JQP, wpxBzF2LMRqX8vVd, wyMJYRVm2Qf7dThv, x6g2PMwjvLdWFZfT, x8Bz4XqAQ1wfCtpP,
x9jFAkqb6PcWTfR2, x9v27CNzhRgMkXG6, xDdjVFwHJ1M4nWyk, xWnkZ38mATh2XHJd, xWpq74Z8Nf63TK1a, xbJBZj14NWCXrhAm,
xhtLrB9MjHqWvVY7, xqnmW3BDpFb28cJ1, xra2mpXtqAc9YdvN, xz7YCK2qgAnZjk61, yBCqrazHtRNXY91T, yCA1rNbBHntcMqxF,
yGb8A3VQvJfrCgTp, yHAmqcVTdCnGQp4j, yJ9crQ64tPmaFCzv, yLKXxQgrPWTnmdZA, yRChfAvDqYwxBkFr, yd8vFWNnTtH6xL9K,
yvw9JgXQ8WmfGHzT, yxqvH82LFm1nwAfR, z1VCfZFJYPwHagpL, z6LMfrpbwhyVCAJK, z9XbfMRh21Fg7QjN, zCyk3LFp4XZwaQBd,
zL8QNrFp1PHWacbv, zM4Dd8QKwBATnJR9, zgqYjRFPMTV2LA31, zhNraQdpRDX23j7w, zkQabTrD4RHLZN8G, zmDY4LVX7N8vh9aR,
zmTNyXkfqtpgcM1H, zn64f9D8TVt3wFZq, zp9yVrtfTNZJCv7g

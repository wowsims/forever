# EV-0132: Tamed pet and hawk crit, the hawk's swing at level 60, and hunter crit and trap rolls in live logs

- Claim: a tamed pet's white swings take none of the crit it inherits from its owner, while its abilities take all of
  it; after the 2026-11-10 hotfix a hawk's swing before armor at level 60 is 0.153 of the creature dps over 2.5 s plus
  2.5/60 of its attack power; Predator's Edge reaches Raptor Strike's crit damage bonus twice, compounded; Explosive
  Trap's first hit rolls its base once and its attack power share not at all.
- Sources: Warcraft Logs, the reports of the hunters on the Forever servers that the sections below name; all fights
- Game build: 1.60.1.70888 for the reports from 2026-11-10 to 2026-11-12 and 1.60.1.70938 for those from 2026-11-13 to
  2026-11-16, as in EV-0131; the build of the reports from 2026-11-05 to 2026-11-09 is not recorded. The hawk's swing
  counts only reports from 2026-11-12 on, after the 2026-11-10 hotfix had reached every realm. The hotfix each player
  ran and whether they relogged are not recorded.
- Date: 2026-11-05 to 2026-11-16
- Pull request: #1059

## Method

The reports were found through the hunters on the 20 Forever servers and each one's recent reports. A hunter counts in
a report only when every combatant snapshot of theirs in it shows one set of talents, read from its trait tree; their
level is the talent points plus 9.

A crit share is crits over landed hits and crits; misses, dodges, parries, glancing and blocked hits are left out. The
owner's crit is their Auto Shot crit share in the same report, so it meets the same targets. A slope on the owner's
crit is divided by the share of the spread in the reports' owner crit that is not the binomial noise of their Auto
Shots (about 0.87), and its interval comes from 2,000 draws of whole reports, as does every "95%, reports" column.

Damage before armor is the logged `unmitigatedAmount`. On a crit that field is also before the crit, so a crit
multiplier is the crit's amount over its `unmitigatedAmount` times the target's armor factor, the amount over
`unmitigatedAmount` of the same hunter's nearest normal physical hit on the same target within 10 s.

## Tamed pet crit

Pet damage credited from a paladin's Retribution Aura is left out. Abilities are every pet hit other than the white
swing: Claw, Bite, Bloody Screech, Swipe, Lightning Breath and the periodic Savage Rend and Scorpid Poison.

| Ferocity | White swings: reports, crits / landed | Crit  | 95%, hits    | 95%, reports | Slope on owner crit (95%) |
| -------- | ------------------------------------- | ----- | ------------ | ------------ | ------------------------- |
| 0        | 266, 857 / 154,361                    | 0.6%  | 0.5 to 0.6   | 0.4 to 0.7   | 0.06 (0.03 to 0.10)       |
| 1        | 21, 476 / 14,717                      | 3.2%  | 3.0 to 3.5   | 2.7 to 3.7   |                           |
| 2        | 19, 501 / 9,215                       | 5.4%  | 5.0 to 5.9   | 5.1 to 5.8   |                           |
| 3        | 14, 789 / 9,386                       | 8.4%  | 7.9 to 9.0   | 7.7 to 9.1   |                           |
| 4        | 20, 1,436 / 13,238                    | 10.8% | 10.3 to 11.4 | 10.1 to 11.5 |                           |
| 5        | 146, 21,628 / 168,932                 | 12.8% | 12.6 to 13.0 | 12.6 to 13.1 | -0.06 (-0.18 to 0.04)     |

| Ferocity | Abilities: reports, crits / landed | Crit  | 95%, hits    | 95%, reports | Slope on owner crit (95%) |
| -------- | ---------------------------------- | ----- | ------------ | ------------ | ------------------------- |
| 0        | 201, 5,309 / 81,036                | 6.6%  | 6.4 to 6.7   | 6.1 to 7.0   | 0.94 (0.82 to 1.07)       |
| 1        | 17, 568 / 7,212                    | 7.9%  | 7.3 to 8.5   | 6.9 to 8.8   |                           |
| 2        | 15, 596 / 5,867                    | 10.2% | 9.4 to 11.0  | 8.7 to 11.5  |                           |
| 3        | 14, 747 / 6,704                    | 11.1% | 10.4 to 11.9 | 10.0 to 12.4 |                           |
| 4        | 13, 1,115 / 7,265                  | 15.3% | 14.5 to 16.2 | 13.7 to 17.0 |                           |
| 5        | 126, 18,487 / 94,198               | 19.6% | 19.4 to 19.9 | 18.8 to 20.5 | 0.92 (0.77 to 1.09)       |

The ranks between have too little spread in owner crit for a slope. Across every rank, with Ferocity's 2% a rank taken off, the abilities' slope is 0.93 (0.85 to 1.02, 386 reports) and
the white swings' 0.13 (0.08 to 0.18, 486 reports), the latter from the white swings gaining a little over 2% a rank.

Reports (429):

19KWay63j2HvGdgQ, 1HTPQGZya7pAz3qh, 1NpPkArt7Cdn9W3w, 1aBhkzdDMZKGtb37, 1xzPrQWDC4pv9HhJ, 1zMFQx6RYwkgDXTh,
1zdqMb3QnryZL2j7, 23d6yh4tcqTGNnKR, 24baXzjGrVgvmWDx, 28vgnypFaHmVkqDL, 2ACDTKkcpWbNfRJ9, 2CcwRgf9MhVLaPJH,
2DvWNfwZHqgXdGLM, 2WLXwQG7kfaBRbxn, 2d34LGM8rbcgAYHQ, 2gKNqpY4G8VdtLMf, 2kV73aphdFrCQq9D, 2nVxKL17RY8FAwk4,
2q4vxWyfZnLKF6zC, 2xhVyXGW7gaDjA4J, 3ABMTH9n8rqRb71z, 3BQyvXG7rRdfh9pa, 3C26abGgdcQKATJt, 3CDRTVdmQN7y6gXY,
3nXD8WG21VmKpHrj, 3tLzvVH2AK6aFDMP, 3wdx94CkMQVGmayv, 41JLRvzkwTdZX9Y6, 468rLCt1gBdXzRQY, 47Qx1fcW69RwHMqN,
4L7DBHTV2jA6gkQW, 4WtzXrnhyZpcRVCY, 4YbZc7VkWQ8HzvqF, 4aQpr3LJXTN96AC1, 4dakVPwbnc6QKjvh, 4j6p9mM8ZnqDzytN,
4mVAwCvGatgHWnfP, 4pkf7QamrDyW9YBx, 4zXPcrTMjnNLZqv6, 6Jn1dFwNQCkgYLqV, 6KtcxCT4b7ZDvqMr, 6NRLznh1jgfwqrvb,
6QgzZWmDLHY2Afx3, 6vWJarDyMYxpAt8T, 74MAxargnGd2B8t3, 74YybNVPDg1MpJG8, 74zrGjnmBdYDCg92, 7Ag8kQBhtx9v26an,
7HMvYBzVhG8NmkyZ, 7Hkh8FawGxbgZ9yW, 7QtfGD3ArVhH8qdw, 7d9yw6VNfWhLTkFQ, 7fz9Kk8rY6jyQAbL, 813rPaLtcdpHy67q,
87hBYkHFvXf32TCd, 8jzXTaqr732V9DxJ, 97pV3KgNWFkqaGtd, 9BkpHDXjbmZhgPwW, 9JxanDLbZqC18YPg, 9RxBtJPw1kgD6phd,
9WdLMbYFfz6jNpHr, 9Y3JvDkrKBL8NMHQ, 9YJM6y2PdRxZarK3, 9YZPHyRaK3FrQkNT, 9Z6yrfgDLjPahGAV, 9d1bQX8qcPJ26VFZ,
9gRDVvwAQnC41PBc, 9gXLdDHvJf6jZAK2, A2ncKhLMVbTwYNvj, A6PGMqKRzybtTp1N, A7WRvYrB8ahMxwcK, AGWhn4CV8Yrjbpf6,
AYwLfvqh9QyNj6zJ, AvMqTmDyG3VrZNY2, AvNwTdyBbDncFf2h, AwJbDaHfdx289Khv, BFPtcmbyLr71Kj63, BPG9j1V6hHTwQdxy,
BVgtPkXmAcJwM21Q, Bcfrn3xbXV8CMzaR, BfVjHYvWdKkP8Jz7, BmDpWPqCrJfhZxtK, Bmz3GDv8WRbZ62dV, BnPXHVWwTrCYNtMh,
C3r6TRtDnxzVPF9j, CLBtcNZgqYb8zQhd, CTPkZ8AWYwcXBzNn, CXvjmgaJ6YqKzD4k, Cat9MZ2fyqxj8gJ3, CbVw9y3XZPnLvzR8,
CrJY1g8HjTkc3dyX, CyN68LDZPxkFtQYB, CykdWTB4ZpG2VXQF, D3Qx92znVrkXpvC7, D7Kbvy93gkT8RJHa, D9mj8dy41FxGfcV3,
D9tWGCxNHRawZ8Bz, DJx8N7ckzXZq9Agm, DMv2cWLFgT13AyRG, DVJCM8qRwmvPXFYn, DVMfnxkJLvHqPZaz, DWjgPMmLFrZ4RacV,
DmyKc6zWJ7MwHdNR, DqkcQyp7HZWMGAz2, FXcLvZV8nxQT96fd, FmXh8Z9TB6DnLHGx, FvRW23mY16cNVBkP, FxcpZfaP2LYvGH4N,
FyjZtHCT6zM2WLqb, G1JF39DABcrVvTHC, G2QDqn9gc1mRXByC, G4akXpft3gNhCJ19, GJx3tmAN2v476pBh, GNp7bXWgfjr3yAcm,
GWabBnwjVk16xTyf, Gb4aLkx6KNf98qgD, Gbkh7r8FVnCPYgpy, GgMtdQWhCYJAbHD7, GzqDHgcPy4vQXBxZ, H2LPzQA4a7Nyr3jb,
HL38fnFjNXGTmRCa, HMc3fnaA1bJmPt8W, HWG1a7n34wbCgdmv, Hw4kJRqxZhfvapcd, J16aZwyRBNXT3QbK, J1CPKR4Q8pfz3m2V,
J4nGFVKWD1Aygkaj, J4zpnVTLMaY18tq2, JC6v7hNYf3XrFgHb, JDtYTwZ12c864k7a, JVy7a8hrgRMZv6Xz, Jcx8g4CXq1QBpzVP,
Jd6WanvfBV4M2zAk, Jjmv6drycNXpDxW8, JnVckZDLTwAgGKaQ, Jyw9hD1bmKWCN3XH, KLNhgMTDAV6xY1Wj, KYvNjyn1p2Lb7M4r,
KcdpBWTJyX1xLFfh, Kdkm8xan6CFZjvpQ, KfgYa2LbdCRGMhXv, KjdWqMAvFVN7831w, KzwZMvnPtXcF1p3f, L6GvARYBpNQy1hk3,
LDngCPfhQcKpqTyN, LF3BRbd1fx9gnyk7, LKwbAMzfnRZtg6q8, LRqc8tNbC1Hf7Vwn, LZYnpwV6Jxdr4WaH, LgNZ7Q6mtHAd8RyM,
LhByjkW9tmKAaMZ6, LjvDknrWTVP1xAaz, LxfZwpgAyMFcC3NG, MGfjz4vFadZHxt9y, MHtZ8pdhF9CwWxNv, MWHL9h6fJRNdbPt8,
Md7V9kmZxBNgr8KQ, MjfT3pQKRWcVG4gY, MpKYVnkb1wraGZdz, NHLVxbT9hgPZnGwp, NKM174PyfxjTXRv3, NRJG9M1wF7ng4BdC,
NRvfwJbCyQ3qh4dj, NVJ26vZQ9KnkpmRB, NWzbYnVjDaqx3gRC, NXMZV9xjvWpB7GtP, NY6dCGJWTMfb3XVZ, NbmjagBQMYLC8t6z,
Nby7FHqDhWcBLYXd, NcRqzah1TQjMr3nF, Np89a1BDwGQnjJcx, NydpC9jWtq31DXmn, P2QkCnLwyB1cVA9z, P6HcTpZXdqAkQBYh,
PA1XWx2zbVgnHQKa, PB6zGcVNtTZrMKXb, PGVWZxJDC4AMQb8L, PLGKvft3bq8WgRYN, PN64hvHQLZRCptGg, PdZtrjfBJqYMVzCy,
Ppd7C9MWXFkr2Tga, PzQLwbTDy39mCVZr, QAgrxLyh1CVwcf6k, QF2qpKC9rbxV3ZvW, QLfF372VKaMPm91T, QXFd74zDW8JZMmKx,
Qb1yWxa8vHCGL7Yr, QcrkDWdPGqZmv6jf, QqbhvMX12Y3ZPBVG, QvDAW3czJBVa6Ngx, R8wpcTxFbDPz37gr, RBaWwNzPf81x3gcY,
RCGmnkZ9PY1vtQXh, RDdJFMxKyXtCzwnP, RTkzcFndwNYjVMP7, RVgGFynN3KCYJHcA, RWq13nP4gBdhCpmw, TA3pLxKFXQ9Ckm41,
TH81CKfzvmZVBbNQ, TNAXzyxRPQZW3pq2, Tfp1v6KhDwJntrRd, TgXQr3V4LGam1CMd, TkdGL62CnKxgh8A1, TqZ39QgVkdF6WRDr,
TxX4AqCgafhyn6zc, TxvwCp4YWqygQ1Pt, V4v6h23bHLaT9Kxj, V6YjPXcQLfD2FBZk, V8k6pvjqDRtLzNHM, VHRAcyF1ZWhzqC4v,
VK83G7MdBhLRFAfa, VLqM1kcTB9NdjbtA, VhLaJgcKpGyFZz1m, VhqB8wZJzRP34cAM, Vk2BXyHrQZ3tjTJz, Vw3yrtjXgAhcx94m,
W2GLAPRX7jw8BC3n, W2qHFNwj7BtbgQnD, WBVwmaXxh1JfPvKC, WKFAG7kVZwxbR8QX, WKqhakHcDtMmAdb8, WQ3fqPhAVRrcn7Bb,
WaHFDqN2c37hGx6v, WvcXQGBAqKfbydwj, WzN41Fn7cL8R9KpH, XBbarKyLTcJfNxj3, XCam1TRQqwW6pkzf, XDZpkGWVRTfxb6Nn,
XFdDMhAQVxYtjWB8, XHczYZgmf76t9Jrk, XHk7W3h1NDbpL4PV, XJ1V4FxwjZTzLW8A, XKxGMWahY9D2vmc6, XdFYRDafB7cjhgJb,
Xwxna4t6pmvYFAyH, YCm7yQkZM2wvTaP1, YH7n2xZqPKzWjQrR, YHFzM9jpyQmL7Ara, YL6pQKrq7ZXCbHdV, YaBQZ2tbVFJMrAhk,
YcRVPMyFfqAgbQ3N, ZC3cMxKjYRpVP7L9, ZDRTLNaKyw8gJdcM, ZYxGWdqtLT7Fjg1n, ZfDykLVKgBP4YQGA, ZgBryJzmMnxLf9w1,
a43dJxFhB1CktfWw, a4ZdvYKyhcLGnFRP, aA67j34DCwZztKfM, aHwCztFpYhVXKdNR, aLYv7r1kDFpxXJVG, aRFzTnCLAtqHNdB8,
aRkHpWzNmfVGqAth, aVX9hAZ2ryGCwqKz, aXYmb4znJjcA1Dq2, ad24TYKCAb8LwBxc, aftcNjVY3dgGAZQB, aw2tRTYN1ADzm8Jq,
b1jA7LXFrZfmGMCR, b2RMy8qGcCmpFPVH, b6Pkw3ZrLtBHnC9f, bBcVpmN3ARdWD7Jx, bYaT9KXtGwqvV6d4, bavD2gBYZHkqK7Ct,
bmZcqx1pW4LGFfKB, brvjnxhJ18mTDQL2, btAc341MaWNQH8RL, c14qPbJdkhX6a3vK, cATBb6FrVznvMPGk, cDxNAW6Zr3hvC2QG,
cGrfRLdNmqyKZpFP, cKChLPz39GTX74wm, cVtBTGnm6h9ZP3X7, cYjJMvzXhQKaGkFL, cb9gzXjCK3BfH7qM, ch3xqmgd28AzCyFY,
ctChBVNr23xnkfTY, d2FLqCfQJV4XT1gD, d31BQFCaAXKD8bJr, d7RvpH8nJtGXBxc6, d8pgmJxYh73cbyj6, dFcXZkRP9qnLb7mT,
dNT6DXyhHFRjZLaq, dP8LDZ7z4tTk3xA9, dPXqjcwTzFCHfYmb, dRGjYkN2xHC1QXz9, dvkzFtPH3WR9C4Km, f2zK39XdhnNqZJgR,
f36RaB12wyDcTCGY, fD1aRkzx4jYV3drh, fLKYjJNVRCgxcwQT, fVaTmnXZvPcA7gxy, fWF83zPgHX1RpLkv, fYdRHbPy4nA2k8Mq,
fn2Dc941K63CNvmj, g2TtjmJ8X1RFwQHb, g3dYGjzy19w2qfXB, gAmv8kd3JtZcbyLP, gBDMLv8J1K7nqkpd, gKk9M6qRwByWYPLj,
gPxrZVY4JQ2Fm1Nd, gRtBKzxVFjMJwQCm, gjAQxFmzMtq3d9Za, grBtAJ9vRTk2XGjx, gzrAwTqCWPm7xvpD, h8aVABn6zKPyMwD4,
hAb6yN8xnqBpWj3D, hHfr6ya17DNvb9Xn, hLzm4tKQAc8DCMpR, hPNgqjA4vkTxXwG1, hmtTqH16pPGzZJax, htKPAwxrDdXV9znQ,
j6XFRzHxTC1kvVpy, jMWpNJLtkVPHy93z, jXhdfNz2g7WpBYCa, jYVvXCw4QANH1Ffn, jacFgk7PvVTXxNd3, jbVWrL17CFtDcN3v,
jhGZM8bFmY6HtgVc, jrTKDZcxNBvtdAP6, jryFckzCJax34KQb, jrzfg8kh6M4tJFAq, jxwNCpqt6V3zmQMa, k4DvNfBJC3wL7MRc,
k7gmfZ69awYr3Dct, kbZYPcw3t9jNa4mz, kjRYLXvzgbJyFGKD, kvPydcGVa3Lh6gFA, m3YgJPcrwFVfzTkQ, m9Z1hnrfCFcpT6bX,
m9dYchPtxMFBJ2Gz, mHkvNhbc1Ja7XfTA, mP6KgncLvka8jrTh, mT2F3XZgcnyf7wzA, mXvcxNnBrtgGyq8b, mak9vPD1KCtHW3n4,
mbxRn8FNZk6vH329, mhAWgkRNjGBfF4DZ, mpXk9g8PzctZNJda, mtXxyZRfgAhLP7rc, mwkA7n9ZCJKWtVyd, n1WgqK3GMVLT7pA6,
n29MtQFKRzqDC3xN, nFvHyDjrCxLdA6k7, nRawKrGLyj1TYd2C, nr4qJHp3tQaxzyTW, ntrf7zv2JXmD9yPB, nxpWmkQJXhtPHdF9,
p1jqMBCc9tPG7YTv, pHnQx1LdM9b6tRXr, pQbkWJAnqFgHxDzV, pYCF1GjZ4nXbH3cm, pqba47xPhz3X1DCf, pqn8tzQBAHGW7K2v,
prYKwjPA1WRyQJM3, pwdKxmAar9MBZDtc, q7MvX6Vd42jLrbF8, q8kLnf63bz2GAPJW, qGywkYrdHKT8NgAQ, qNVWc2yZJT6k8mvh,
qQgj2MWBcD4Gd9an, qTnGKAYkRXHBMDQJ, qVvZxP3gNp12KWyb, qW9njb3hXp27gMRT, qWtC7YxLH2G1FRQj, qhRXQ43wNmkcPLB1,
qhx187GfR623yjZg, qvbXATzmk6Gha1Ny, r3an7RhCT4NjVyfb, r6mNXwVKZhpGjvbf, rB41vNC2npYbz7Zy, rC4yxgXTdHkQcB2m,
rWjb2mNRfwdM7pq1, raBGY1z29kVctF37, rdKzHqAk4BYZL82y, rf3JF4cCybBNZhwp, rf8j1q4Dw2pGVMRJ, rfkDTHnR8CBAwhz4,
rkzFZMhRByHDtT8q, rq4vtxVnZw21kLCX, t2W3kaqMxzFPBDjL, tCzP4jWxdNLHcgp9, tY2jp7cVGaQx8nDw, tbC9yvQdpY7FnXmJ,
tkHwWYF1C8LTfhQg, v41rYANj2zBWkb8n, v84Y93dCaQrGbwcy, v8JbfQNY1RpDFXqh, vAMz47GFTwqpXgQf, vQZLCzhPNatk7r3d,
vfQr1GhC8HjdNZ7L, vjh8G9aydTHJgXN7, vk9HJZwrfhmCayMY, vkrZLbmDRx3ydpHz, vnLtQck7gF9PamD4, vpFrJRb71fNAatmL,
vwCXzhMQ2drgpRxD, wFbHKCBdJj6Z8DzL, wJCmzRy4GHgDTA8F, wJDCrmcnPXja1QbW, wPYdtLC7f8Xbc2qy, wZK63XDc9JntB2Rx,
wZnqkKFbjRNgc7Md, wp7Wx3jJVmtyn9qY, wyMJYRVm2Qf7dThv, x6g2PMwjvLdWFZfT, x8Bz4XqAQ1wfCtpP, x8kcqNDnBtG3TYM1,
xDdjVFwHJ1M4nWyk, xZWLaA8ByPzq1XYK, xgCF4fVbKcTmpaN6, xpJdXyYCVm9zg16a, xra2mpXtqAc9YdvN, xz7YCK2qgAnZjk61,
yCA1rNbBHntcMqxF, yG3LRBbqHwkPfvJx, yGb8A3VQvJfrCgTp, yGhgd9Zp6WQMaKkf, yMcZ2H8LfNGVDJPC, yP46XLfMk2rTQxYj,
yRChfAvDqYwxBkFr, ytDBmMjX3q8TWbFJ, yx9ta7bcdv8Yjw1H, z1VCfZFJYPwHagpL, z2BJcmAP6frjMV8w, z9XbfMRh21Fg7QjN,
zWZgvGc8ByFV2qwN, zmDY4LVX7N8vh9aR, zpbwQnT2387xJYdr

## Hawk swing

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

## Crit multipliers of melee abilities

| Ability       | Predator's Edge | Reports | Crits | Multiplier (median) | 95%, reports   | 1 + (1 + share)^n |
| ------------- | --------------- | ------- | ----- | ------------------- | -------------- | ----------------- |
| Raptor Strike | 0               | 62      | 324   | 1.998               | 1.995 to 2.000 | 2.000, n = 2      |
| Raptor Strike | 3               | 3       | 16    | 2.393               | 2.346 to 2.394 | 2.392, n = 2      |
| Raptor Strike | 4               | 1       | 23    | 2.529               |                | 2.538, n = 2      |
| Raptor Strike | 5               | 15      | 265   | 2.690               | 2.689 to 2.692 | 2.690, n = 2      |
| Mongoose Bite | 5               | 8       | 65    | 2.301               | 2.299 to 2.301 | 2.300, n = 1      |
| Strider Kick  | 3               | 3       | 18    | 2.180               | 2.177 to 2.182 | 2.180, n = 1      |
| Strider Kick  | 5               | 18      | 295   | 2.300               | 2.299 to 2.301 | 2.300, n = 1      |

Savage Strikes changes none of these: Raptor Strike reads 2.00 with Savage Strikes 2/2 and no Predator's Edge. Wing
Clip hits for so little that rounding spreads its multiplier from 1.9 to 3.0, so it is left out.

Reports (87):

19KWay63j2HvGdgQ, 1zMFQx6RYwkgDXTh, 3C26abGgdcQKATJt, 3grN8Rct9Xzknm1H, 468rLCt1gBdXzRQY, 4WtzXrnhyZpcRVCY,
7QtfGD3ArVhH8qdw, 8jzXTaqr732V9DxJ, 9JxanDLbZqC18YPg, 9Y3JvDkrKBL8NMHQ, 9Z6yrfgDLjPahGAV, AvNwTdyBbDncFf2h,
BVgtPkXmAcJwM21Q, CLBtcNZgqYb8zQhd, CXvjmgaJ6YqKzD4k, D9tWGCxNHRawZ8Bz, G2QDqn9gc1mRXByC, J4nGFVKWD1Aygkaj,
Jd6WanvfBV4M2zAk, Jyw9hD1bmKWCN3XH, L6GvARYBpNQy1hk3, LjvDknrWTVP1xAaz, M4aF6w9Pz7JmrXgW, Mw6q32gCaKB8Hjnf,
NKM174PyfxjTXRv3, NRJG9M1wF7ng4BdC, NXMZV9xjvWpB7GtP, NbmjagBQMYLC8t6z, Nby7FHqDhWcBLYXd, QAgrxLyh1CVwcf6k,
QRCgnhck3KVq6vBT, QXFd74zDW8JZMmKx, Qb1yWxa8vHCGL7Yr, RBaWwNzPf81x3gcY, RTkzcFndwNYjVMP7, RVgGFynN3KCYJHcA,
TNAXzyxRPQZW3pq2, TxX4AqCgafhyn6zc, VK83G7MdBhLRFAfa, VhqB8wZJzRP34cAM, WQ3fqPhAVRrcn7Bb, WaHFDqN2c37hGx6v,
X6Hp4gDwv7zftxMC, Xam2g1pGyjvzM8ZP, XdFYRDafB7cjhgJb, YL6pQKrq7ZXCbHdV, YaBQZ2tbVFJMrAhk, YakBLWFpjnrdhGb3,
a4ZdvYKyhcLGnFRP, aLYv7r1kDFpxXJVG, aRFzTnCLAtqHNdB8, bmZcqx1pW4LGFfKB, btAc341MaWNQH8RL, c14qPbJdkhX6a3vK,
cDxNAW6Zr3hvC2QG, cGYvJnm1TfNVxqgA, cb9gzXjCK3BfH7qM, ctChBVNr23xnkfTY, dvkzFtPH3WR9C4Km, f8TKwNhrJqyBjx73,
fn2Dc941K63CNvmj, g3dYGjzy19w2qfXB, gBDMLv8J1K7nqkpd, gzrAwTqCWPm7xvpD, h8aVABn6zKPyMwD4, hPNgqjA4vkTxXwG1,
j6XFRzHxTC1kvVpy, jMWpNJLtkVPHy93z, jryFckzCJax34KQb, k7gmfZ69awYr3Dct, kvPydcGVa3Lh6gFA, mXvcxNnBrtgGyq8b,
mbZcgPavhdG7fKDW, pQbkWJAnqFgHxDzV, qVvZxP3gNp12KWyb, rB41vNC2npYbz7Zy, rq4vtxVnZw21kLCX, v41rYANj2zBWkb8n,
vk9HJZwrfhmCayMY, wFbHKCBdJj6Z8DzL, wyMJYRVm2Qf7dThv, x9jFAkqb6PcWTfR2, xhDAKtcGBJ1yfM4b, xra2mpXtqAc9YdvN,
xz7YCK2qgAnZjk61, zL8QNrFp1PHWacbv, zmDY4LVX7N8vh9aR

## Explosive Trap first hit

Level 60 only. The first hits are grouped by hunter and by the attack power logged at the hit, so each group shares one
attack power share: 42 groups of 30 hits or more, 2,369 hits, means 239 to 604. The median group spans 0.186 of its
mean from lowest to highest hit and 0.150 from the 5th to the 95th percentile. The base rolls 208 to 265, 0.24 of its
average, which over hits of 239 to 604 spans 0.09 to 0.24 of the hit; a second roll of 0.2445 on the whole hit would
widen every group by about 0.24 more.

Reports (21):

8jzXTaqr732V9DxJ, AvNwTdyBbDncFf2h, BfVjHYvWdKkP8Jz7, Hwq97xLh6gv8ZPrY, KPwXNByAG2xQm9ra, NXMZV9xjvWpB7GtP,
QRCgnhck3KVq6vBT, TA3pLxKFXQ9Ckm41, VK83G7MdBhLRFAfa, X6Hp4gDwv7zftxMC, a4ZdvYKyhcLGnFRP, gBDMLv8J1K7nqkpd,
h8aVABn6zKPyMwD4, j6XFRzHxTC1kvVpy, mak9vPD1KCtHW3n4, mbZcgPavhdG7fKDW, nFvHyDjrCxLdA6k7, qVvZxP3gNp12KWyb,
vwCXzhMQ2drgpRxD, yGhgd9Zp6WQMaKkf, z9XbfMRh21Fg7QjN

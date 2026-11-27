# EV-0072: Pet focus regenerates 10 a second and the summon keeps attack power and health

- Claim: a pet's focus comes back at 10 a second, 1 every 0.1 s, times 1.1 a rank of Bestial Discipline and times the haste of each of the hunter's ranged haste auras (Quick Shots, Rapid Fire, Berserking) while it lasts, the quiver's haste not counting, and never shows above 100; its attack power and health are taken at the summon and kept, while its armor follows the owner.
- Sources: combat logs of 2026-10-20 (a level 30 Cat), 2026-11-07 and 2026-11-08 (hunter pets at levels 5 to 48); report codes not recorded
- Game build: not recorded
- Date: 2026-10-20, 2026-11-07 and 2026-11-08
- Pull request: #955
- Later count: Focus, §Count of the 2026-11-11 to 2026-11-26 reports

## Focus

- Count: §Count of the 2026-11-11 to 2026-11-26 reports

## Summon

In the logs of 2026-11-07 and 2026-11-08 the pet's attack power holds through 20 Aspect of the Hawk toggles and a gear swap in one summon. Its health holds through the owner's Stamina changes, and its armor follows each gear swap.

A sim run locally at the summon gave the pet 51.6 attack power, 10% of the owner's 516 ranged attack power with Trueshot Aura.

## Count of the 2026-11-11 to 2026-11-26 reports

- Sources: Warcraft Logs, the reports below, the tamed pets of level 60 hunters on the Forever servers; all fights
- Game build: 1.60.1.70888 for the reports from 2026-11-11 to 2026-11-12 and 1.60.1.70938 for those from 2026-11-13 to
  2026-11-16, as in EV-0131; the build of the reports from 2026-11-17 to 2026-11-26 is not recorded. The hotfix each
  player ran is not recorded.
- Date: 2026-11-11 to 2026-11-26
- Pull request: not recorded

The reports were found as in EV-0134 §Method, and a hunter counts at 51 talent points. A Claw or Bite cast logs the
pet's focus before its cost, so the focus regained between two casts is the second cast's focus less the first's after
its cost. 490 pets from 411 reports give every pair of casts 1.4 to 3.4 s apart with nothing between them and the
second below 100 focus; each pair is read against 10 a second times 1.1 a rank of the owner's Bestial Discipline, by
the owner's buffs that cover the whole pair. A pair that a buff covers only in part is left out. The Melee entries a pet's
swings log carry a focus that lags its casts, so only the casts are read.

| The hunter's buffs over the pair | Pairs   | Focus regained over 10/s × Bestial Discipline, median |
| -------------------------------- | ------- | ----------------------------------------------------- |
| none                             | 263,724 | 0.991                                                 |
| Quick Shots                      | 58,017  | 1.290                                                 |
| Rapid Fire                       | 15,448  | 1.391                                                 |
| Quick Shots and Rapid Fire       | 2,669   | 1.797                                                 |
| Rapid Fire and Berserking        | 418     | 1.531                                                 |
| Berserking                       | 171     | 1.088                                                 |

Quick Shots states 30% ranged haste, Rapid Fire 40% and Berserking 10%, each as A_MOD_RANGED_HASTE, and the hastes
multiply with each other and with Bestial Discipline: with no buff the median regained over 1.6 s is 16 at rank 0, 18 at
rank 1 and 19 at rank 2, and Quick Shots reads 1.30 at rank 0 and 1.29 at rank 2. The quiver does not count: the pets
of 238 owners whose Auto Shot runs 1.10 to 1.16 times the bow's speed with none of the three buffs read 0.989, and the
pets of 98 whose Auto Shot runs 1.27 to 1.35 times it read 0.990. Neither does the pet's own Frenzy or Bestial Wrath:
0.990 under each. The haste takes
hold within about one cast of the buff: a pair that starts in the first 0.8 s of Quick Shots reads 1.16, and one from
0.8 s on 1.27 to 1.29. No pet logs more than 100 focus.

Reports (411):

16hJKBmq2V8tnzXR, 16kchwqKn3rpyGMz, 1Kmv3YFG9P4DqZJx, 1Qxcn2mJpNHwatA7, 1VRbCQD4MkKtY97G, 1XJvGy4K29CqxtQj,
1Y4NWLka9pAJbB6G, 1h6tv4PBXQgqb8Dw, 1h89JnyWjpZxfXdk, 1xzPrQWDC4pv9HhJ, 1z4vRDqGdM2cVPXk, 27jZGhfHgp1LR8MB,
28cLVtZ6D7Wp4jmF, 2FN4M7GHLWDmdkRJ, 2R37ZrTWvjHnCJLw, 2fNJjyTc3C4ABxVZ, 2wMznDRvmkpbtHxA, 3B8AcaGNRKmdgPDj,
3BAz1TDGL9kjy8ZQ, 3H1fvFaqxRL79PY8, 3LXKpR8kYbWactCm, 3MqmK9ngZvaNDj82, 3R69PBjxZvcfbnX8, 3RB9X8HVxgCJ4My2,
3fFatLkqGnmjQ8ZW, 3qfDVTZ1J29zPChn, 42XtM7Jwdk9gnvqW, 4BtbFKLZ23G67dy8, 4HZTp3WYcdxD1bQV, 4J7dhZxb3GakBQ8p,
4R6dyPjBXmCr3Tq1, 4Xfp9L3bKh2aPYAn, 4hQkw2BmKJGd1XZ3, 4nkW7JDxV12waMfr, 4whtjfzW91pbaLv7, 67gJ19PCVB8bXv4r,
69j8PA1zfKvVmpkM, 6BZ8AkvTNwLJP9tC, 6G9qNRc7KL4bDYjQ, 6PN8t7yDdwrVncJb, 6WG74An9RrCdqkwD, 6gnAywbVMGq7vRxZ,
6yHgkT8LqP42NmxY, 7BraqHpPmfyRjwc6, 7NhKDGwbdFqWB1La, 7Q6WxzYRwjNqXvgZ, 7TXcFNRzjwHdAybn, 7dDbtWFhA8BPn4vZ,
7fy93LdcwQhvF2gP, 7fz9Kk8rY6jyQAbL, 7kgR1WtDBXJbdAYF, 7pmhcYx1JNCKvWjn, 7tbWfYqJFd4vzL3p, 7yLVTagpRbjYMhnz,
813rPaLtcdpHy67q, 82BgQM3GHJcL7ydt, 8G3K1TCaq4fLXhHz, 8TQJcMxNa2jG7ZLf, 8YfKRvzTxHCVpGtQ, 8bWVvtJRXFQNTAjc,
8jzXTaqr732V9DxJ, 8rAWbkX93chqKjFG, 8zC3dW71xXrgaA2v, 97aD8RXPqZjg13Hn, 97wxhXLvAaGKYz18, 9AKatWHpNxnkhBQC,
9CgtkpAYqGjxP34Q, 9H2chVMdn1ZtxXpA, 9TD8g1kVRfZ6yLGn, 9vcQpzdbZmG4kyTP, A1yxMTmKatkf3HRJ, A31fKY79DJCjM2Qz,
A6pfHymTn4qNMZd3, A6pmrCYMfGjWK12z, A7ahGYbzQ8M1ZtF6, AMRnwhKV6JvFQYpr, ARFXzbqHDpBLJdhy, AXjhYkfN13L6HCPv,
AYwLfvqh9QyNj6zJ, Ag6nVjXKpcRHkw8y, ApJhHfRkMDqFj3rb, AvNwTdyBbDncFf2h, BFm724ahJxPNZDVv, BVJg8aWCbQmM6hfp,
BWcP4pDq3xAkKF2r, BXMpQqGY7caChPjZ, BfVjHYvWdKkP8Jz7, BjyVTt6wG4YamQvd, CF7ztDwY8pJ9KZhn, CMGKZHRVPwzvFmqy,
ChbG2MpyJjDWPfV6, CpkLVaQhXD7nMAzg, CtMF9BgnAmWkTbfd, CwRq8LgZbADvXcHK, CxTmQRX2ytGcwA9r, DBRqygp13jPC6c82,
DBzk8qw4yFWtb9rd, DH3dT6tYLKmMrNz8, DJCRPcry8mjpb1nH, DKftp1VTZw9hBLrX, DRf6gP9bac1nrMzH, DW9YkdvyrjtMBQnf,
Dbfd6avYpzwTCx3N, DcXfK2MmZgyR6wpq, DmJcrQzhtqkLpRaY, DvH9VxKZnagAPGL8, DwzYTd1QPHFf6qJZ, F2vN8aVjBz1PJ4Q3,
F3BTZGrW48Mp6n2L, FDRCpv9WTk7VAJjL, FHJjATXtmKkRnPN4, FHqYCNcKntvDAjxg, FJdrDz9w1Vxaf2cT, FTQ8g6cvbyZ1wB3m,
FkBxY9zn4faM7jWL, Fz2wJmfGKPnvTy9h, GByQ7g9ZPkqxFJXh, GRPFkVj2bKhD4Npz, GRnMHmK69BXfWVTj, GaKFhmvgVN4cCWQJ,
GbvNW3tAP68ZpmcY, Gkb3rKHF6V9gz4pR, GnZdPBVXFW6kwp78, Gph6gaZYMz92RbcJ, H7K31X4mQGCRVfnb, HPnctDh3axdKw1MV,
Ht8VFJxmnCG9bQhM, HxCQat38vmjnkpb6, J27rzKkLAndj4m6g, J38xbrFytgM9p1LT, JC2MqjcFA96gLpxh, JG89gCcma1Lbxrfn,
JLMk78RNzg34wTAH, JNTRK1qgDtkCz4rQ, JNTYBgF8W9tp2afq, JQgB3aWcjTDV9nkM, JQwbjnVYyGcP4rg8, JaFQcdYDj2pNvbLr,
JrqDZX4LT6NpxzM7, KG62VXF3a97CxtAg, KcdpBWTJyX1xLFfh, KcgXx6na97dmMfGH, KndmGyYFazPjHftg, KpwNtQh1DXMAWcTm,
KqZ6xWAfDLGrVjYJ, LCXrjHcgnAJ2yBaZ, LFNZrg2Cpty63BnK, LkYjKP3HpcG7ZDVA, M2Hj9kfNC86rLAWD, M7kjDhzcr3RNgXQx,
MWrPDAVFCBXzh8Z6, MXmdGkgfYvN2rWjq, Mg2ay1xDjkYpBqRX, MgmdJqxBWA3Tw2vz, Mh6CHzaPQZc23Tf7, MtxafB13DwLnCFzq,
MypmYhB3RX9FJ7Kx, NCfWgvjhqMk8R7td, NPHc3azxL8bmgdpF, NTJzVdMZBQDaqG83, NjnA7zPdmZhgCvyL, NrpM7DJQBKWqbzVF,
NtxKh7PgTAaZnHJd, PBLXNA3MhxYWmTKp, PDFVdp172Kwz8aqj, PFJwz4HkR2xTM8Zb, PdNq3JxntCKgavkm, Pjd9MWAXgCYTcn7Z,
Q2qrkZR6NnbzLmaw, Q7V24RA9HTg6F38p, QAgrxLyh1CVwcf6k, QJkv3RGLCyNBzYxD, QNKD3HnVZFApMajY, QNwkdxbMvWmrqjAY,
QXFd74zDW8JZMmKx, Qb1yWxa8vHCGL7Yr, QjHW1hvfbNGgcLBd, QkwyM9RmnV68xdPH, Qm9XJwGdqNzhRc42, Qr2hVGLaFDydjxNC,
QwJFzfMABRYX9vkL, R4rLFGNfpg7wxHYC, R92pdBxqLDrnMV8C, RACFcXrV8YD69j2K, RBWjPh9tT61q3xFr, RK4bDBcJ28Gnp6Tj,
RQFBGJLVmbYynrTh, RY1yA2PFCBwb4MZa, Rf3wzmWBt9TAvbX8, RgMfxbJrY89nTpmq, Rhx2TCj8wQN1XFkZ, RrATya1w6ZhY42bB,
Rry6WVb2NGD1dXMx, TA3pLxKFXQ9Ckm41, TY8hRktpNaCnDXM2, TaCQNVZFp4MX67Yw, TcLMDwAGYvQk67tC, TcnphZkX2fM9g7Q6,
TmQyqnM6hN217LZC, V1jLfaQtnGdHPB2z, V9k7jdXABJnCwfQZ, VK83G7MdBhLRFAfa, VLz78BfZyWa3Cxr2, VYBjgJx8aZtRhmpy,
Vdj689qcv1F4ZQnh, Vdv8p3ZnTB6fGMQN, VgjdJkmCTrtYcXLB, VjrAGFkTvNy2W98B, VkLX4CZnc68R2Gb1, VqwAHdP4XYm7cxBk,
VtcPT63wQ7WaACvj, WFHghVcLRtPpw4MZ, WNDyzRCwd3jhJnkK, WYTawk8z3LAPpmrx, WZLD8RYn6T9kh2mx, WhPpmnzB6rR3gD4Z,
X1NhLgQrYynPBxWk, X8PH2CG1hMkcfxYJ, XbJKk1n8crD3aAtz, XqQYd9BpCGFmcjJK, Xv2QmnyK8RBxW9jZ, Xxdpgy7KNcz1VkwT,
YFtvD87kznr1CThg, YH7n2xZqPKzWjQrR, YVWgj6ra27PvmNnx, ZCPahyN2JmkYMGLQ, ZcFKV7tLHbDR4qJB, ZfbY6RgWjx9hmpGz,
ZhbCF1LvkQ9gBAMy, Zt4Rqa6zLWXbFB2m, a2MWRNwFLTPHAzD4, a7VDZcCkGArwj9Nn, aZKTvVcCHLwhgXzM, apMdm4GJH32L1DVC,
atk24BMr1PfjVJRc, avnFMLCYq89B3KXw, b2HF8pcvrkAGZgdT, b9gAB3nH6qVwpMRC, bF8hVwJNDT6HyRWt, bRfPGLA3Qp4Dn6xa,
bYaT9KXtGwqvV6d4, bavD2gBYZHkqK7Ct, bfmh3jwygC4V9A68, bg169rGdXTzDhLnj, bzwY28XH6Vfc4Zpg, cL3KMWpnXyvDrqa2,
cb9gzXjCK3BfH7qM, cf9LwRYktpgvZdH8, cgMGCdPJQ84wzaBV, cgrN7Ywd2fZzHCna, cr6qMtj1vCTPHVfw, crP79HDfXz2bkZ6M,
d1vn7jt4RH2ZmLAB, d7xrp3LnhRQAb4Jq, dA2gmBnLHkjM1bxh, dBNfgVDLMmnc8h4v, dGh1Fn4rgkjB7Vtq, dKnYzPTqv8JCM21t,
dMxyPfarzTc3tLbk, dnDg7j9kRw6WtbML, dvkzFtPH3WR9C4Km, dxZCXJvPLKAB3cnQ, f36RaB12wyDcTCGY, f3hyv2GtFW4b8wJV,
fJnW8D7qYaRXcT2M, fNbLwMvKCrHc23gG, fTBkLPW4MY2zN8AJ, fcnqgCZtXyY34Gbm, fn2Dc941K63CNvmj, frdmcpgKB6twT8jH,
fzD2LmvpFXCTK7jg, gBDMLv8J1K7nqkpd, gFp9AyntNQVqJMmc, gVF7L1nvY3bCqwcz, gXhWFKATBJRaf3cQ, gmKvNqHFJLZXfdnP,
gmN7Vxy8tPaD1bhZ, gnZmp1JWfPwBzhC8, gtNvbhF9aZCXyzVk, h3nCT1KLgZwG6PcY, h8aVABn6zKPyMwD4, h8vCa2P43nbH7JwX,
hH4YRGVFz3KnN6bw, hMv4dJgY3pnyVXTk, hXLT6BWagD32pkZv, hmDYk8MKdpqx1FzJ, hpd6BPmjGbQv8TyV, hwJNG8k2a93g4nCK,
hzmvKQk4PyY9bjcB, j6XFRzHxTC1kvVpy, j7xyJdLXHZNA6PQg, jDfKLTzaWywBAZMk, jWnC8Z7xDvpRcw46, k1jVRK7g9APNTLm6,
k8NhDFbQjrCPzvn4, k9fKT8ZCtMFn3vab, kGpgavmFADBVRhq7, kH1dt4TpQjfYmaAv, kL1g29NH7MhXvBRQ, kNFDnMJXZcyrpLBd,
kTKH1qFZYJhWvXVQ, kmxBzfV3LnR6NDb9, kpjxBPY8dgmGqbZF, m3GbaCBXw1dj6rQL, m4FHwDBYVNPtQhJZ, m7XPvgkLqH1ZnB9Q,
mJZWhHDAt19gaKdw, mNTMK4J9PQcjCR6L, mRnQ6xjzVDvw2PMA, mak9vPD1KCtHW3n4, mb8FgjhwVTLZ4q1d, mbwzgt3VpLXTj4MZ,
mg8qKkahVpLnXZvW, mhbXcFqf9xTaA1NP, mnPXjHcWNFKBhDCq, mpgnDQbLWXY8jJP1, mpx92zHCYLnTKfAv, mxGpkwbNA9f7vzYH,
myD7KYZ6TzC8aQck, n4vQdtMkRTc1zxXV, nDNT9zhcQ8trJbwf, nFvHyDjrCxLdA6k7, nK1fCTqXYFJxbMgr, nR1WGzY8pJKDxhQj,
nTfhtVKvmw8gZkLj, nYLth6W2ZbByR1Ga, nbLKdYfXHDam79R6, nyMqbY64tL7HhKQJ, nzWQvwYgca6LBfyj, p7ATmV4HqZRdFwLX,
p9vmyLhPgCf4tN6D, pKJGZBx9wfTdQ1kq, pNW7aGKnzHP1MftT, pVNna7r8qH64dcRb, pW8CdNZtPrbQfqGV, pWnPwNFGgyX7tkdq,
pYATGQ6aCmzwBNcM, pgZM9rYdbH7kz6CN, ptvYCVPymJXH21kR, pwzCJQZXDcLT39g4, pyhwdHcqRamYDXzK, q8wm1hGCyHNnzT7v,
qF9L7VZHR1QG8r2m, qFbWHnA6PzydX7Nc, qG68VWwXrPKvfFxh, qVnt8TGrHXY3B2Q6, qVvZxP3gNp12KWyb, qcdyZAr2ganDtxNW,
qp6P4GLHwa8mdVjM, qz3bVhMgtQX81dyk, r9b1TFDzPk7qjQfc, rJb3LkQVwa2CGNDW, rQpg6YfW7cLRNAVM, rRnVJhQFa9zf7NqK,
rXAdyCnkhqHN1ZDt, rYj7BxLKmAN8t2fv, rfz4pR6mx1w9JWKA, rjMPpLXcgH31NCWa, rnfRqHt9Vk8g1JB7, rtPp1Bd6L8yQZCNh,
ryM1qhHFwc3YaDzG, t3fny7PvdGMRWhCT, t9PLr4H8gdkXncRZ, tLfvwqZm8jcWBdaF, tMpK7wQXD6WfFcyN, tQ1ymLaFbAT8c6KV,
tW38cgRn6b2kr4dJ, tjcwR486HaA7Zr1h, txRbXZwMJm4BW7q6, v23ytJHDz6QkbXnN, v8L6NkQTzMV2mWq3, v8Zp2QF7An4d9xMj,
vGFaKgLT9mkn1Jdy, vGxLrFDkg823p9qK, vPHb3a6c1KhknLrY, vQZLCzhPNatk7r3d, vRYfZ3h24XJQpyGt, vbXfBWDQk8jAPTr9,
vg1kL4rN2KZVCaRd, vwCXzhMQ2drgpRxD, w8n6JQ4DM1GVyKcA, wC2pGtAnLyBcNDj7, wFQXdZknKhAvfyaz, wJB7m9F8bc6tgMZx,
wfWDyFRcMdr2jZ7K, wz46qKkG19hmBY8T, x9v27CNzhRgMkXG6, xN8bgQ7Gtf4p2dA9, xWpq74Z8Nf63TK1a, xbJBZj14NWCXrhAm,
xfWNbt8yZKzcRPjC, xkKRJPqCLzBa9GdH, yAYRmxqgXkTwr6Bz, yGhgd9Zp6WQMaKkf, yHAmqcVTdCnGQp4j, yLFbzXntM8dvQ3YC,
yPJ6pm4Q8M9qfLag, yWFhcNBab8wYMQtG, z9PqXTpDyVG3Fvn2, z9XbfMRh21Fg7QjN, zATdNgKZC1kGDqtJ, zNMDVR1d3kfGCpw7,
zYPAwt4gc1GRDWVk, zhNBgjHxQX7anZTf, zhNraQdpRDX23j7w, zkQabTrD4RHLZN8G, zmDY4LVX7N8vh9aR, zmTNyXkfqtpgcM1H,
znXGBy6wajHh8bFQ, zp9yVrtfTNZJCv7g, zpvfR1gAaKmbWTMy

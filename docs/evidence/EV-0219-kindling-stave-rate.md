# EV-0219: Kindling Stave proc rate

- Claim: Kindling Stave 11750's Ignition 1300781 rolls once for each enemy a harmful spell the wearer casts lands on,
  at 1.85 procs per minute on the cast time of the cast behind the hit, at least 1.5 s, and its 10 ms cooldown lets
  one cast proc it once. A killing blow shows no proc.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every wearer stream of the item, listed with the fights
  each wearer had it on under §Reports; 389 reports, 397 wearers
- Game build: 1.60.1.70888 to 1.60.1.71139 by report date (70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19
  UTC, 71037 from 2026-11-20 18:23 UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU,
  71139 by 2026-11-25 22:38 UTC; its rollout time is not recorded). Reports per build: 70888 5, 70938 47, 70959 91,
  71037 170, 71091 60, 71091 or 71139 16; 232 start after hotfix push 114313 (2026-11-20 23:47 UTC). The hotfix each
  player ran is not recorded
- Date: 2026-11-11 to 2026-11-26
- Pull request: #1253; the harmful spells that deal no damage, #1255

The proc is 1300782's fire damage on the target of the hit that rolled it, logged 0 to 49 ms after that hit (median
26 ms). The game files state no rate: the column reads 100 and no procs per minute row exists, while the tooltip reads
"Harmful spell casts have a chance".

## Method

A roll is a landing on an enemy by an ability the wearer casts: direct damage or a debuff apply or refresh. Periodic
ticks, misses and resists, wand Shoot, Auto Shot and physical hits are left out, and so are killing blows and
Devouring Plague's applies within 50 ms after it drops from a dying target: neither ever shows an Ignition (Mind Blast
0 of 346 killing blows, Shadow Word: Death 0 of 1,234, Fire Blast 0 of 763; moved Devouring Plague 0 of 4,017), and
the proc lands on the enemy the hit struck, so the logs cannot tell a roll on a dead target from no roll. A wearer's
landings within 10 ms of the first form one roll group (the cooldown is the wearer's), which procs at most once, with
chance 1 − (1 − p)^N for N landings. A group procs when an Ignition lands on one of its targets at most 60 ms after it.
p is the procs per minute times the median cast time of the ability, at least 1.5 s, over 60. Runs logged in two
reports (same wearer, start within 2 minutes) keep one copy (6 dropped), and a wearer's last stretch with no proc where
8 or more were due is cut (1,196 groups).

## Rate

198,490 roll groups, 13,740 procs. The fit is 1.8373 a minute, which rounds to 1.85 at the nearest 0.05.

| Model                                | Fit                       | Log-likelihood |
| ------------------------------------ | ------------------------- | -------------: |
| Procs per minute on cast time        | 1.837 (95 %: 1.810–1.865) |       −47074.5 |
| Linear in cast time (two parameters) | 0.6 % + 2.70 % a second   |       −47069.0 |
| Flat chance per landing              | 5.05 %                    |       −47332.1 |

| Groups           |  Groups | Procs | Procs per minute    |
| ---------------- | ------: | ----: | ------------------- |
| Single target    | 167,283 | 8,693 | 1.870 (1.830–1.905) |
| Area (2 or more) |  31,207 | 5,047 | 1.785 (1.740–1.835) |
| Mage             | 103,500 | 8,928 | 1.845 (1.810–1.880) |
| Priest           |  43,067 | 2,188 | 1.865 (1.795–1.940) |
| Warlock          |  31,651 | 1,549 | 1.800 (1.715–1.885) |
| Druid            |  14,974 |   802 | 1.870 (1.750–1.995) |
| Shaman           |   5,298 |   273 | 1.515 (1.345–1.695) |

Single-target groups by cast time: 1.5 s or instant 4.74 % (6,610 of 139,588), 1.5 to 2.2 s 5.52 % (224 of 4,056), 2.2
to 2.8 s 7.55 % (1,395 of 18,469), 2.8 s and longer 8.97 % (464 of 5,170). Per landing, before the cooldown grouping,
the cut and the killing blow rule, the rate reads 1.65 before push 114313, 1.65 from it to 2026-11-23 and 1.69 after.

## Per ability

Per landing, killing blows and moved Devouring Plague applies out; each Ignition is shared equally among the cast
landings on its target in the 60 ms before it, so a debuff landing in the same millisecond as a hit (a queued Mind
Flay beside Mind Blast) takes no proc from it. Cast time is the median measured from each cast's start, talents and
cast speed included.

| Spell              | Cast time | Landings | Procs | Chance |
| ------------------ | --------- | -------: | ----: | -----: |
| Arcane Explosion   | instant   |   60,288 | 2,682 |  4.45% |
| Arcane Missiles    | instant   |   26,041 | 1,358 |  5.21% |
| Mind Flay          | instant   |   12,186 |   588 |  4.83% |
| Shadow Word: Pain  | instant   |    9,836 |   490 |  4.98% |
| Fire Blast         | instant   |    7,444 |   337 |  4.53% |
| Corruption         | instant   |    4,029 |   208 |  5.16% |
| Devouring Plague   | instant   |    2,413 |   104 |  4.31% |
| Shadow Word: Death | instant   |    1,725 |    68 |  3.94% |
| Arcane Blast       | 1.5 s     |    8,834 |   462 |  5.23% |
| Mind Blast         | 1.5 s     |    6,394 |   278 |  4.35% |
| Searing Pain       | 1.5 s     |    3,336 |   180 |  5.40% |
| Wrath              | 1.47 s    |    5,191 |   244 |  4.70% |
| Frostbolt          | 2.5 s     |   13,994 | 1,002 |  7.16% |
| Starfire           | 2.45 s    |    2,270 |   178 |  7.84% |
| Flamestrike        | 3.0 s     |    7,220 |   500 |  6.93% |
| Arcane Wrath       | 3.04 s    |    3,099 |   286 |  9.23% |

Each Arcane Missiles missile rolls, and Mind Flay rolls once a cast, on the landing that starts its channel. Cone of
Cold (3.55 %) and Frost Nova (3.24 %) read below the instants on several targets and are not explained.

## What else rolls in game

- Blizzard, Rain of Fire and Hurricane casts proc it on an enemy in their area at the cast (about 840 procs), while their
  pulses roll nothing (Blizzard 5 of 127,274, Hurricane 2 of 43,516, Rain of Fire 2 of 24,817).
- Counterspell's interrupt procs it (about 190 procs).
- Harmful spells that deal no damage roll it like any other landing. Per landing, as §Per ability counts them before
  the killing blow rule: Curse of the Elements 4.65 % (77 of 1,655), Curse of Recklessness 5.10 % (23 of 451) and
  4.35 % (5 of 115) on its lower rank, Curse of Tongues 6.47 % (9 of 139), Silence 4.11 % (12 of 292), Faerie Fire
  4.03 % (10 of 248), and Polymorph 3.85 % (7 of 182) on rank 12825 and 6.16 % (13 of 211) on rank 12826.
- Flamestrike's pulses 1279990 and 1279989 roll at 3.4 % a landing (818 of 23,905), about what an instant-priced roll
  on several targets under the cooldown gives (Arcane Explosion 4.15 %).
- Arcane Missiles rolls on its missiles' landings alone. The game logs one cast for each missile (10274 and 10273,
  31,752 casts), and 13 Ignitions land on a cast's target within 60 ms of the cast with no missile landing before them,
  0.04 % of the casts against the 4.6 % a roll at the cast would give.
- Hellfire Effect rolls nothing (0 of 39,595), nor do the wearer's spells fired as a proc (Ignite, Shadow Weaving,
  Winter's Chill, Frostbite).

The channel casts, Counterspell and Flamestrike's pulses fall outside the fit above.

## Reports

Code and fights, by report start:

NHa2RqtjxAvPFwLg/10, 8V9FC2NtwW3yLmBp/1-28, kH1dt4TpQjfYmaAv/1-6, ngdvX1HhYcbN6RCr/1-14, YaKJbmPtANW3L6nZ/1-17,
AyTmpNtxB9jbnJ2H/10-45, NXMZV9xjvWpB7GtP/122-123, 7XK1DyrJ2TfAQ48p/1-3, Zh9LTdGmgbW1BXw3/3-4, Xv2QmnyK8RBxW9jZ/27-38,
tCnKQrPVv9Z6qdGX/1-5, GzqDHgcPy4vQXBxZ/7-27, tbgKX9CkwQ7YMxmT/1-3, zWA8JhVxZKaprwTc/8-9, 941apA2h6kRL8DcW/1-2,
kaH1G4ztmYpgMTJ6/1-2, 8GNt2hC7v1rWRXyc/6-10, QAgrxLyh1CVwcf6k/3-6, vZqVmLYQa6jwP3A9/2-21, bzwY28XH6Vfc4Zpg/1-16,
daT4FcQgCYw2WBq6/1-2, 7mnVXCgpyxkbtZGq/1-12, FxmYM7VcjZgL4XNJ/1-4, dvkzFtPH3WR9C4Km/8-16, PFY6KBxymcbDqjT7/1-19,
XQkavwVbDrLKmhJ9/1, HbvWQBKhRn9AX7Z3/7-26, MC7jkfvrpcbyTQ4x/2-21, h43GgypRxfbcznLD/1-5, YZy7L3P6AVhWJN1M/1-22,
v23ytJHDz6QkbXnN/1-3, 7fz9Kk8rY6jyQAbL/1-10, zFXxNAn69HBGRaqg/9-10, QD8ypCXWJcv9Y2B3/3-11, 7ChG2VXaALzcw9pW/1-2,
9H72ZkdBGfhCy6P3/1-21, M7cz2xknrqXVAKjD/1-3, VNpDMaKXy2LHAPZh/3-13, m6htAf2g7FvD3WVj/9-18, bP2CFzkZTLJ8q6Dn/4-16,
aZKTvVcCHLwhgXzM/6-7, bkDMVFPT6a91AXZq/1-39, Xh67CyAm1rcWGPM4/1-21, pMkPWzaC7wf9XZNt/1-7, NB4ag2CcKryDT91x/1-5,
H3DB9yJKA4hqw6Lg/6-8, 1bgXr3HRqG2cTnjQ/1-23, RW7kfjbYghJyx1dK/25-26, 4pynYMkBdVgRDvKH/1-10, p6mVRBzvWMqnQZFc/11-12,
tkvPjpQGY9Cf2zhT/1-3, v8XzFwL9pbHhBCDg/1-7, WfKv2TphtRrzNA3d/7-10, Ya9kKCWRwqjFpgz1/1-4, B4QqHnGXwzpRfaYW/3-4,
bkm7pBz4N9JtC23L/1-4, mtYDAGM9p3bqV8Wx/1-2, Xa3zRAWjGTvL4m9K/1-2, c8faRh2FPHkCt46b/1-9, LN2Vk4nMtBc8vWbT/1-6,
JZDFdhkvn3NP6xLm/1-7, xXGCDmpRakYWJBFy/3-26, yYPpk4TqAhjX3xC8/1-12, fCB4wQt3NWm6k8M9/1-7, MRJHKBWt2TCp17PG/6-7,
GP4zxdtkFDcA7LhW/1-5, b6PMNzGgRt1HwcVv/1-7, xWDkmwXKrRcdM7Jv/1, Xfbc1Pa24xzTBwnN/1-2, LYw76XG1CFHQNdJz/1-10,
kJZaP7HqRvF3xyKY/1-2, fjYFz7Q6PZRWNa3C/1-6, pMgWD6Yhm1Zvr8nw/1-10, 3aXTp97wjCWNxMhd/6-13, Xxm9BJHVWKjRvfd7/1,
Ycq2twGVavTpMn13/5, n9cJ3qy86bXrKfvw/1-11, QwYrJ4VPjyBkRbft/1-2, DnCBXW7ybdxtVprw/1, vyQrDZN7RX82ApnP/1-3,
RVhbZDtzWP6c1jNf/2-5, V8BGhv6yKrjtzCMA/1-2, 2tCNaTJf49KWyVLk/16, yW82C7hYvDmZkbc4/2-8, bVACzhMqpGfjPyY2/26-41,
zjLkFw6hrPvgnm8x/1, nYMZcpQ8WhRTt97q/1-4, 8Gw7m39ZXFz1y2fv/1-32, KTcgdHnN2ajkFfXP/1-2, pKCQwXNvnRBA84DF/1,
yjV6m7kpfgaPzXxt/8-9, RJXQC67jbzATWdx1/5-16, k1taHbwL76Cnxv4m/4-10, rVh6AXQvKgZJyPFz/1-2, CYbmtxQ1THrG248Z/1-2,
Z6Xm92K3zgfrqM84/1-12, 74ybmLP1cBR28KnJ/1-2, DMdzh7YqFjg3BrV1/3-17, Zw93D7qyL6zcRWkQ/1-2, F1X2mG4Zfdjv7B96/1,
qKtgYkmZMVjnLBfr/1, adYfCFcrgv231VRL/1-9, x2QH6gyDAb9ZrR3q/7, yb2dVDXtBQ4q9jLw/1-2, xyGf9QY4dwT8JHMp/5,
wq6XtpnDBzgFLf3Z/7-10, pARakYHTbQdWJjZM/1-2, bnMvHkFjxQLJ4VRT/1-5, WwDGjRtHYNZB7k36/1-3, RVma8DjNx3qnLHtw/1-3,
w1T3Z2hQHpkPcAxY/1, PpXafk1jrB8LATt6/1-2, qXPhR7Jtd8jKTB9p/1-3, z3VLyCTfvwRrdKbN/1-3, M6qak4W89xgcRCwX/1-2,
KHb2ktR6MgVLxrYc/1-6, GXrPqfDM2ptbandg/1-10, pHdLtwQrYFzMjb6R/1, dZpnGbHyTFaRVQ3k/1-3, fqcPvpYkFd6rgbHA/1-2,
z9wcdxZaRLhWk1Qf/1-3, knyCYtGqvZDWzxg1/1-9, NW7Fx8QTHLD2v3fY/1-9, 1bkyjpY4nN7JFqfH/2, q9K4Zv2mk3dnGaAP/1-2,
G29y3QBWV7XDR6mj/11, mXh4w3Pnz2BLW1xd/1-5, nyk7Dq8t3Jj4bhAT/1, pc7drLMKbFJ1nhGv/1-2, vYjL3q712CyTGkQh/3-15,
JQrpCnRxygXP4KL7/1-11, C8hLYZf7Tdgz46DN/1, 7GcATyt6WRjBJvaQ/1-4, Nc4MByg2LAd8VnDx/1-4, gTnLQqam716J2hAB/1-2,
FycQRpgMvhJ14wDn/1-2, vHp9QagZGFLBM7PC/1-2, 13WH2NcCdqJD7L6X/1-2, z7dxgt4mMVGLThCf/1, gRWbpVBwcJxPjKZ1/1-4,
y6hmbtDYwnQFpMx4/1, YkRwjX92cTAP7G1v/1-2, Jhr4QdVGtK3wZpLR/1-4, JAqxj8zanf7Zk1gX/1-2, yLJdAW4G2K93MQhC/1-2,
bPWwqdBY3vn8Kz6H/1-2, yzhRr8WJqTQb9dfP/1-8, 4cDqBtZ6XGYCnP7y/1, kdfANxKrp12BLWyQ/1-2, V1C2gjfm8kbhzpWL/8,
MDk4wyGC1dbrY8cH/3-5, FzTdnXhJBvVjMGqk/6, HvtmnFTPNZQRWg86/3, tMv7DFn2H6Gwmb9a/1-3, FJXjBnPcNr7dRkwb/1-2,
XkD2CdJ7QbaMAhVz/2-5, gWZ4YyRrCkzpAP8w/2-5, HhpZmgKMfXjzWwrk/1-2, WCF6TXxtDNMcna2k/1-5, hWYGMDpr1Bbc6Zzj/1-85,
GFQKCHpt62M8rRdg/1-51, MJ8a7QkXxtYRqGDH/1-3, 2H1VAy3MDrRndPvw/6, nfC91mcXzG4Dq8rd/8, TQYALMf39F2jWh6a/2,
tph1mCvBaPr8GNZQ/1-3, t9QVZBx2CN7Frmk4/2, gPay93qMJcFrDKZX/1-3, pdqNgvKcJAjBDakP/4-28, AZB4DFJG73yvqVjn/1-3,
7tyMxH1Tk2RzGFra/1-2, Zf2c1PF49rVDgjw8/1-2, ZGAnKLQq7pxHrdX4/1-2, CApKmW3knZr8x4HR/34, Y7NwHLrTpAtgPVZG/2-3,
prHqxRVvGFAX3TwN/1-3, zXFDJWQcKfrg9Nyp/1-4, jNRzcwDFf2nXYQTG/1-4, H61xVDhyrTLj9aGR/1-62, KDCn47kMjpyxAaFw/1-2,
XJf6ZABVhHaxgmdy/43, 3PKT1pk9DjA2QrJ6/3, y3Y8t6hwkQA1B2cG/1-8, TFARWN4Pc1JzGwZx/1-2, DyfNWcpwQRBxaKZk/1-3,
Z4HjaBMVPmkzTAh9/2, jxpKB9Gy1wgkh4NH/1-6, RBQF9L7myjb8aYh4/1, BCnafpPbAGY3cRNL/35-43, JLTamqrjZpWGv3Xy/5,
3TdF61x7gMQ28pKX/4-34, 61VCpZLY7TAqWBzr/1-3, bLqTWNdmzf3g2DFt/1, yJDXtQmMB8Va16Fb/1-2, r3KNJDjz69FXBngC/1-22,
3R47bTBvAqgQCrDt/1, aRfkTMrmBpztghy1/2, XTHZdtfGqP2b387B/2, 6dBHnM12zDfPZGW8/1-4, QvW6VCHNcMrDFpXG/1-5,
yhq2JbVLM1H4CfmF/1-5, rzfgt6FmyxLWvqGh/1-3, ZVf9tDamcR1hk6AP/1, wgdLfnmDvNrTFk1Q/1-3, H1LfWDazFVc7Jt6A/1-3,
QKgr63NjZMz7hHXb/1-5, pnxRG1zCyA4mHrB2/1-4, Ba4bzVmX2nDrRN7k/1-3, 7ZWFDG4XBPKqf8zR/33-36, x9nbrAvFwfmK3YWV/3-8,
mHLWpCvFd1VYj8gB/1, nyq6aJAGXj7MvtQT/1-6, AZ9ndhf4bkyaKMcr/9-55, c3qNjGmB7x9akpgr/14, TM3ZzXkJ8qjCfYG4/7-11,
hR7dDX6bPtwrWjf8/2-3, z2wZ3q9hpMnbxfWK/8-37, W2d6nPMQ9tV1BjZx/2-3, h2BF8zNyvKfMgZV7/1-2, tM6K41v2FpbArGQk/1-2,
YKqp6cRhg9zAfPJ8/1-2, RxNdCcDaG1AyF63M/45, 2MTtDmgLfxKJGdpP/3-14, hpkqztBQvZgJrAY4/45, Kg2da1Jfz76mRTDZ/3-42,
thqTyPArmkxfMnK7/5-38, 4XLqQjchbAzvVG9k/53-54, 4naTbcmP6zyXtAKx/1, Dv2XMZwbPY7ma6Hc/2, R6B94qfAjChQnrF1/3,
wxc28Qvf3V49WYJz/1, YfNAPwrMyqKjpk28/1-2, 6prM1FBVKTgARmhL/1-4, fvAcGJ68rRPtmaZV/1-2, rJaTCp796xwvqL1g/1-4,
ykvAGjz9dmwB6hYL/1-4, nyJTqgd7P2cQvpND/1-2, 1YHGxh7v3KLq2BPQ/4-35, wr3M9DfXFQZTqj72/1-4, KyAFadbJ6DLf1x92/1-8,
gwcLavGNXH1T9Crq/1-2, n4Nvcm8jWtVqKHZd/1-2, 9XFWN1zxfgLpZCK7/1, yMYVXpTNcxLJGBCz/1-3, T3a8ZkW2dJ6jrMvN/1-6,
gc3TvPVDFGYdkRNJ/50-52, LjMK9QJkaAf2m1w7/1, CV1DwP4Bct23AzGY/8-42, v2p9aYZ4b13NHLhf/1-5, B1WXm7QGV4hdHLbK/32-36,
Tgd8A79Z1mfbJqkB/1, dvhR6rWnDqxGab9g/1-4, v4ynHhcF829VprmA/10, dZL2m8NJhyHCTFVD/1-3, TYAFc4pG6WNRrKxQ/1-2,
zyGm48Cb9cjN7ZtK/50, BFtRGWmNv8JV1Tbq/2-5, 28q1cPCZXBktnmxh/1-2, T2aWdxMQ4Bw9tLGv/5-43, TXzd6Z7F84nb1R3M/44,
trYaMNzmHFG1gXBk/1-18, Df8YB9wqMdrQRKWt/9-44, tB6pZ1YcqAJ9NjRH/6-35, AJZQMFgTGxnmkV1f/1, HLtZjNGzq9bQc6rd/2-3,
YMNZW3T71phBmXQ6/1-2, 2r8NRy1VDWLTCbdq/5-8, D71Mx6ZRyYVTWnzw/5-82, 6ht4rB7GNxMFRbCJ/2-5, zmb3Ddp2ftWgXZvk/1-2,
hq6Tzf4Mbwpv8VjF/1-2, yaWvJ4Fj7YCQZT8N/1-2, xXVgrwAKTc2GJnkt/1-2, 3jQ7gMqr8m9dLWRF/1, wWgtJyPBfH6T2brD/1-3,
zaPX4DQ8yYRJjwAb/1-3, br4AMRdq1Wv7jpcB/1-3, KrWLQhNymAx4b6HY/24, YqvcHzFbKwpt7Bnx/1-3, FWXTJcHxnwRZakDj/1-3,
ChAHD6rfYm23KJFd/3, RZAWcjPyC4mNgbH3/1-3, 8h21yZTpzqFcBRxV/1-5, X2BHzGg93h1vbdM7/4-6, H3Pta4JbGg2fMFcB/2-7,
3Pa4C6wWQZJyjpxG/1, za176d29mpBkMnZh/1, VmZALY9PKwHJ7fxq/2-4, WG9ZxQbwn7jRqa6p/6-7, 2TvqNHmpnGhrBk7b/2-3,
Z63rfHgGkBjzxDnM/2-3, bHgRrBwFjqyPLAKY/3-4, vhTwnZPb1gBVMF6j/1-4, b6M8QqVjYpTNa1Dm/1-10, bRXTDdAWxkn4hYzG/3-19,
fr4dzjT3Dx2WkRZB/3-38, prGRAQNfTdxFBk72/2-7, j3pvLFZ7rmBgk8xD/2-17, 6cKNnwLvTGDWqpVb/1, TkxDmPpnjzh3MJ1H/2-4,
3jFLzH68Jnxy1cTK/1, 2NMx7HcKnXvtbDkd/1, H8XbkRvDQZ4p3Ywa/1-5, PbhY8aBcKv6RXr4g/1-2, 1aADhF39HKtjvLwx/2-3,
JC3waXhZjQG67c8g/1-8, HgBCnFKprVhL3J2Y/3-4, Vp9yHCmRkatcAgMq/1, 3rDCV19fnMKY2ZkX/1-4, gbZhtNpCMT4zwK9m/1-2,
FH6jLBmJTbNWwzrt/1-4, xQy7YfRrB2mF3ndw/8-37, JyxTCZaVf2KznHw1/3, 3w2gL6Wj1cqXApfT/2-3, r4hnqTNgQP2zJ7cp/1,
td3M7FJqG94fawQv/1-3, fgnRqwYhWcZK9vjC/2-8, hdGQmRZ4qMwJbatV/3, NJnx8ktqvg6aZY9R/1-3, gxh3JraGBmF1f2RP/2-6,
3dn7kP9VWJc4q18Q/1, 29ZWyVxcrX1ha6Av/1, x2awXN7ZVgz3yDMA/1-4, H3TqYxR1k4phg2yw/1-14, QJkv3RGLCyNBzYxD/6-8,
R4bh7219AjnrNv6G/1-2, N2vLpDafkjbmZzgq/2-5, tHcV4fNgxaMqhGT2/1-5, Ngcb48qvWpQRMtfh/1-4, y72Cf4pTxQ1WFbaV/35-36,
ZHhbDAfgNPGw7jKn/1-3, hkHDY7bf1ctRBp9L/1, Ww7g2kXh3ZMYHLcd/1-5, MyZGhvYbCJVtdj7H/2-3, Pz1mrwty3JaAdT9n/2,
bWH41fw8CDQAch3K/3-6, 7rbPgnDH8J4mTjwV/4-6, wYAFrKfHRgdTCWyB/2-5, 8QkFGYPqcJyTdwzf/1, RB1fCw2AbhVnGgHY/1-3,
pVNna7r8qH64dcRb/1, Hg96pT4BY3L8Gdm7/2-3, aBJQKkgj127X9Z3y/2-5, D7hPTJ182zvmVxjH/2-5, XKJkTp93czjy1mYM/1-3,
nMXQ9ypvK4jhWc6P/1-3, DV89ATQcFdKCMn1b/2-7, pH4kvW9rKy12fg3t/1, krT7dnCM4KmZ2yLD/1-2, jZKRVvhCXWwMcATb/2-30,
atrgRnDV9T8W64KM/2, 4PnxXQHtdL1ARN6Y/1-5, frQPY8HN4zqv7pna/2-3, D6qKf1YdgznJhMXr/1-3, 7P42rTp1z8t9JAqb/1,
yjq4XnAxt6WzBrkQ/1, 439wqZz6dnTAgDyJ/1-4, PycKXkp1va4L6j2Y/2-3, rdJ6mfkL2KZGzbgW/2-4, cWqCwAzkrNLyh6Va/1-2,
MXkKgPFhByr4TjZx/2-3, dNZ2GJqaQWK8YDCp/2, yQHNFcXabDRmzTAV/6, mKHrRz98jX3vBDp7/6, zYAqrG6V2DZwWvkX/1,
ZLMNzY7KTFDmbawW/1-3, Yhfxq86rjRdMzg4y/1, KrNdChkVYp9GJ3aA/1-4, Wq4XVPGA6CvpZwxR/14, Rc7GKfNxZ82tPqdv/3-6,
ZDC9VMWpvBFb4q61/1, V4dBhLwD3YZpgPJk/11-17, 71nbzXNW2aFBrQjw/1-13, LkYjKP3HpcG7ZDVA/1-4, WtVKNCjMfHYp6L3A/1,
38rbjAPwy1hak6GT/1-3, abVGnzLr3M4QhkwK/2-3, JBtCDYzLrFfx7And/1-3, Gz3Jn1xBHqWhMbwZ/37, GBxPQhTArwWnpFYK/2-3,
B7wJytbHDT9FpYkx/1-16, RjKZgWQXzCDy96fB/4-19, wM6tJNAgXHK1mZ2r/1-4, jBmVP6JK84hcXnAQ/1-3, gvB82YpHJRF67ack/1-2,
P8hC7DK3wgbdH4FM/1, tCRgfnBTbPxGD2kz/1, DTpqWL4GgkJ6mX7M/2-5, 98gKMZXRxqaHL4zN/1

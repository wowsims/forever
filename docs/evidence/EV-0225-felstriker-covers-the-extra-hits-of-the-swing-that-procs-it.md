# EV-0225: Felstriker covers the Windfury Totem hit of the swing that procs it

- Claim: Felstriker 16551 goes up as the swing that procs it lands, so the Windfury Totem hit that swing sets off
  mostly lands under it and crits; the swing that procs it mostly rolls without it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), 154 streams of Felstriker wearers, one per wearer and
  report (91 rogue, 60 warrior, 3 hunter), in the 152 reports under Reports, all their fights
- Game build: 1.60.1.70938 for 15 reports, 1.60.1.70959 for 49, 1.60.1.71037 for 78, 1.60.1.71091 for 6 and
  1.60.1.71139 for 4, by report date against the US times (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20
  18:23 UTC, 71091 from 2026-11-24 15:21 UTC and in the EU from 2026-11-25 02:19 UTC, 71139 from 2026-11-25 19:14 UTC);
  the hotfix each player ran is not recorded
- Date: 2026-11-15 to 2026-11-26
- Pull request: #

## Method

Each fresh Felstriker apply is a 16551 apply or refresh while it is not up, 3,313 in all. For each one the count takes
the wearer's own Melee damage events (white swings of either hand) and Windfury Totem damage events within 50 ms of
the apply. They are split by whether the log lists them before or after the apply. The crit share is crits over hits
and crits; avoided, glancing and blocked swings are left out. The baselines are the same events with Felstriker up for
at least 100 ms, and with it down.

The clean sequence counts one Windfury Totem hit per apply. The apply has no Windfury Totem hit logged before it
within the 50 ms. The counted hit is the first one logged after the apply, and a Melee swing within the 50 ms is
logged before that hit. That swing is most likely the one that set off the Windfury Totem hit, though it can be an
off-hand swing in the same instant.

Scripts and outputs: `requests/forever-aura-order-1127` in the log parser's folder. `extract.py` builds the stream
cache, `felstriker.py` counts, and `felstriker_out.txt` is its output.

## Numbers

| Events                                            | Melee crits              | Windfury Totem crits   |
| ------------------------------------------------- | ------------------------ | ---------------------- |
| Felstriker down                                   | 81,803 / 271,421 = 0.301 | 5,634 / 25,366 = 0.222 |
| Felstriker up for 100 ms or more                  | 10,057 / 11,126 = 0.904  | 831 / 953 = 0.872      |
| Within 50 ms of a fresh apply, logged after it    | 808 / 2,201 = 0.367      | 251 / 365 = 0.688      |
| Within 50 ms of a fresh apply, logged before it   | 42 / 152 = 0.276         | 1 / 12 = 0.083         |
| Clean sequence: apply, Melee, then Windfury Totem |                          | 110 / 170 = 0.647      |

The log lists the apply before the swing's damage in most cases: 2,201 Melee events within 50 ms come after an apply
against 152 before it. These are damage events, not applies.

The Windfury Totem hit of the clean sequence crits 0.647, far above the 0.222 with Felstriker down and below the 0.872
with it up. So most of those hits, though not all, roll under Felstriker. The count is thin at 170 hits, but 0.647 is
about 13 standard errors above the share with Felstriker down.

The Melee events logged after an apply crit 0.367, near the 0.301 with Felstriker down and far from the 0.904 with it
up. That window does not isolate the swing that procced the buff: it also holds off-hand swings and any other swing
within 50 ms. Even so, the swing that procs it does not crit under it as a rule.

Only Felstriker was counted. Neither the extra swings of Hand of Justice and the like nor the other weapon buffs were
counted.

## Reports

1ADrPJVmwgaxLqMc, 2aZ6C1vRF7DAHMBn, 2jJ1L8FPHzTB4fKG, 2qAWx1L63KfkcJaD, 2wMznDRvmkpbtHxA, 2wPQmVqbdCcZW9gv,
3bdcFvQtDZhnXHfg, 3cabDLxwK71Gzq2Q, 3DPvXKRnTJWxZCt8, 43R1XHKpfLgybZqm, 4TxMfmFpzBJ6aN1j, 4ypMXT8RDZgcKVwf,
67fqCHN9ZManpxrk, 6gnM4ThfQPLHkG83, 6ht4rB7GNxMFRbCJ, 6jvXHDwNAakZGBxF, 6vAwndj1FxBcDKq8, 7gW8baTzVZ9pYXFm,
7HpBD1Z9fJcbA6Ga, 7tbWfYqJFd4vzL3p, 7ZWFDG4XBPKqf8zR, 8cqCXZHgpw9Px2hK, 8Gw7m39ZXFz1y2fv, 8JTGAcnFYvtmjLN7,
9JqKQBY68twcR2hT, A6qcQ9jtTNXKbvYh, aFtMHKAG1vjTCwcq, AfZHhcnwNaRJYpgT, aL7VmXZxPdgcBtjG, anAZd29rK3vCJ4wQ,
B6bvH1fQ4kAcyRzt, BfMQJT7g4mZL6taF, bP2CFzkZTLJ8q6Dn, BPVxqXWLc3FJTpMh, bPWwqdBY3vn8Kz6H, bV7Tcp3DBzFLrwtk,
bxWNdRVg4C3Jwaj2, CApKmW3knZr8x4HR, cgMGCdPJQ84wzaBV, cKvwPrNmjRCFzgpZ, cPkXNvDW4hnKqMJV, CxbgQvRAL61MZdc4,
D2nQcj6wFbmBgVyJ, dF3zbWfawRkZ2j6K, dnDg7j9kRw6WtbML, F4JAb8gjwrQDCP2T, F9hmfgQHZbYTKyz4, FBLTNkwYgtPzjyfh,
FDmvgpZLhVbcHqxK, FhG3AxYL9ZNP4RvV, fMPL97aCA1ngvKQk, fq8RFwhABxm2GY7Q, fRbxm7XVTNw8CYhK, Ft1rJVHX3kz7yLBZ,
ftqKjgaPmbNvM21n, G29y3QBWV7XDR6mj, G7CL9NkZtpXWrBy4, G9NmY7ARvHKyb8Cn, G9TVWxhmHjdc3Ptw, gBrvVXRcGKFP6j89,
GhHZQDn4zgckNf2R, gHmyrJdbAB2nGkvw, gnBdKW9myaxzNP3b, Gqvf8dNR3mpJQg2b, gRpyT8V4KkBc6fMw, gv9V6Ty1Qa3MBZW7,
Gx9NyLPCjwDbV3qc, GXw2BAjbtRHv9Fr6, gYm2HAtf4bnC8QZT, h4m98xKRPayTnW2X, HLtZjNGzq9bQc6rd, HvtmnFTPNZQRWg86,
HzKL2XB3rx1hk8Wy, hZLykw9dRnYbtvKP, JbQyngR8NWXKardc, jC6FzGLJ2yn3kvw9, JK6dgtv2N4QCqjnZ, jn3dwzWPZ6r2ky1K,
jNVAY2Pav8GcbFWn, jXTVfy3ApWKPYwDG, JyxTCZaVf2KznHw1, KkRW7Am81LaJqHYz, kMftb1JGnDLjzPwm, kX7FtxzjfYnw69TB,
L6fHWjbn3NVqgpkr, LF63qnYybWDpCc12, LrNcfVvw3PqBQ9CA, M1tPx9DBZGYgCa4f, mCNxAVQ4bj3tvqyJ, MnFgXa87WRKPmADy,
mpgnDQbLWXY8jJP1, MqzhQN9FPHL3fBvT, n7hQW3NYzJwmDaG4, NjWKPpHgymCT832A, Nmv3nfLwY91hZy86, nX1qbatTzGZ3CAwB,
pdKXzcQAHr3MNhCa, Pk6V914wLHJpxA7f, PKBd6Vz2Gkjwp9Zt, PMtHv3y76FCj8AWB, pRA73YvkarFVLNCZ, qGMZVBwJfdtRxr6Q,
QkVRc329w8nLTzKp, QMLY8q2XNVry7bpj, QNfaZ9MTnWpFrqb4, qpYKczLm83WFN7ZR, QwmphLYJXT7GazMd, QZ3hRyAMbDrj9cHz,
rdcZKVykDFH4f7mn, rFRYL1d4yZDcHgq7, t1yGxzP7L3Aa6CQr, T2aWdxMQ4Bw9tLGv, t2z1VPDRbwZqfcKN, t4bYnDj8QMyacA2m,
T7dtHBv9YR2jnLwg, Tg1Hcb7z2qAXWkCx, thqTyPArmkxfMnK7, TmxZ31fRPBhjApqg, TNak4bKY8JvLqXDF, TNW9rHBt3MD1X8Ag,
tPbh63Rc4J8xdv1Q, TYnZqwA2av9z6jG7, tyrGHfnp3ghKNMz7, vazTdHkmWgA17pLc, VL7Yxg3ZmJkNFM1j, VY1qcrJ4HxjCFnfX,
VZAm6yQkGwD3XPKn, vZqVmLYQa6jwP3A9, wfV2DJZaNtY8rAHM, wT6AKYHXqb3tMvGd, x64Wd3Z8A2TKCDXL, xbc7XGytaKACYVT9,
XhVMPfc7mqNJt2Wn, xRPN2jwhCg6abJVH, XWaTVPCnmNAzD7FJ, Y31f7vWzA4nVF6rd, YAZ7thCmypcGBTN8, yfKb1MthLFDc4ANG,
YFtvD87kznr1CThg, ygMZKDTYd4qCQ9mP, YHRv2c3QqWZaLnmh, YmTFGMK1dg7XBz32, yPVkKpqtfJM27WFB, yZvwm7pRYDkAPGL4,
z9wcdxZaRLhWk1Qf, ZBXTKqHYa3QfkNxM, ZhyVJcb2v1zaLx3j, ZLqKhFbkAypnaY9t, ZpNAtvBLPMKmf7YF, zqXGMvdTDmQaZgtL,
ZVf9tDamcR1hk6AP, zVhC3bma8v4jTBp9.

# EV-0022: Weapon rates, paladin procs and Two-Handed Weapon Specialization's Holy reach from a levelling log corpus

- Claim: the procs per minute of Ripsaw, The Hand of Antu'sul, Jang'thraze the Protector, Thrash Blade and Bigger Rock on
  Bigger Stick, and that a druid's form specials roll no chance-on-hit weapon; Redoubt rolls on every melee swing taken,
  Shieldsworn's spell block comes with every Holy Strike while its mana and Holy Shield damage need a landed strike, Seal
  of Fury's shield is a flat share in the open world, and Touch of the Grave has a 15 s cooldown; Two-Handed Weapon
  Specialization's Holy part reaches every Holy damage spell counted: Retribution Aura, Consecration, Exorcism, Hammer
  of Wrath and Judgement of Righteousness.
- Sources: combat log corpus of 44 logs on 1.60.1, 18 levelling logs with StatLedger and 26 from other players; logs from
  the beta realms (11 to 17 October) counted apart
- Game build: 1.60.1
- Date: 2026-11-12 for the weapon rates; not recorded for the rest
- Pull request: #982
- Later count: Bigger Rock on Bigger Stick, §Count of the 2026-11-14 to 2026-11-15 reports
- Later count: Two-Handed Weapon Specialization's Holy part, §Count of the 2026-11-21 to 2026-11-28 reports

The #958 log of the same date is a different pass (EV-0021).

## Weapon rates

Each rate is the procs over the landed hits of the weapon, white swings and specials alike, at the weapon's speed, as the
measured rows already in the override table count.

| Weapon                                        | Procs / hits | PPM (95%)           | Row | Players           |
| --------------------------------------------- | ------------ | ------------------- | --- | ----------------- |
| Ripsaw 9478, Wound 1292573                    | 185 / 2018   | 2.04 (1.77 to 2.34) | 2   | 1, main-hand hits |
| The Hand of Antu'sul 9639, Thunder Clap 13532 | 91 / 991     | 2.04 (1.64 to 2.44) | 2   | 1                 |
| Jang'thraze the Protector 11086, absorb 11657 | 181 / 1972   | 2.90 (2.50 to 3.30) | 2.9 | 1                 |
| Thrash Blade 17705                            | 61 / 1425    | 0.95 (0.74 to 1.21) | 1   | 1, main-hand hits |

Bigger Rock on Bigger Stick answers whether an unstated rate is procs per minute or a flat chance: it reads 11.7% a
swing at its own 3.4 speed, 8.6% at a bear's 2.5 and 2.5% at a cat's 1.0. Procs per minute fits (deviance 2.9 on 5
degrees of freedom) and a flat chance does not (49 on 5).

Long sims with the weapons equipped fire the registered procs as measured. The Hand of Antu'sul on the default
Protection paladin follows 8.0% of white swings, 8.3% of Reckoning's swings, 9.7% of Holy Strikes, 9.4% of Judgement of
Fury hits and 9.5% of Seal of Fury hits (2 PPM at 2.7 speed is 9.0%).

A druid's form specials roll no chance-on-hit weapon: 2 procs in 7736 hits, from 3 druids and 2 weapons.

| Druid                                | Maul    | Swipe (target hits) | Claw    | Lacerate |
| ------------------------------------ | ------- | ------------------- | ------- | -------- |
| One druid                            | 0 / 441 | 0 / 5401            | 0 / 248 | 0 / 309  |
| Two more druids, with another weapon | 1 / 869 | 0 / 238             | 1 / 180 | 0 / 50   |

Their white swings roll at the weapon's rate. Combat enchants on those hits are unmeasured.

## Redoubt, Shieldsworn, Seal of Fury and Touch of the Grave

| Reading                                                                              | Measured                                                                                             |
| ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| Redoubt on avoided swings (dodged, parried, missed, fully blocked or fully absorbed) | 152 of 1394, 10.9%, about forty paladins                                                             |
| Redoubt on swings that dealt damage                                                  | 364 of 3501, 10.4%                                                                                   |
| Redoubt on melee abilities                                                           | 17 of 138                                                                                            |
| Redoubt on ranged attacks                                                            | 0 of 62                                                                                              |
| Shieldsworn's spell block on avoided Holy Strikes                                    | 46 of 46 on four paladins (15 of 15 on the two source logs), and 61 of 61 strikes made with a shield |
| Shield Specialization mana and Holy Shield damage on avoided strikes                 | 0 of 31 and 0 of 12                                                                                  |
| The same on landed strikes                                                           | 141 of 143 and 50 of 52                                                                              |
| Seal of Fury's shield in the open world                                              | 4077 of 4078 logged shields at the flat share                                                        |
| Seal of Fury's shield in a dungeon                                                   | scaled by expected creature health, see below                                                        |
| Touch of the Grave cooldown                                                          | 57 players, 707 drains, none closer than 14.99 s                                                     |

Inside a dungeon Seal of Fury's shield scales by expected creature health at a base level over the mob's level, the dungeon's
lowest mob level being the base in five dungeons: Scarlet Monastery, base 30 on two players, and Shadowfang Keep, the
Stockade, Gnomeregan and Ruins of Lordaeron on one player each, their base levels not recorded. Against mobs of the paladin's own level it stays flat, 118 of 118, three cases on two players.

The sim reproduces two of these readings. Redoubt rolls after 41 of 410 dodged, parried or missed boss swings, where the
sim before the change gave it no roll (0 of 444). Shieldsworn's spell block comes with 409 of 409 dodged or missed
strikes.

## Two-Handed Weapon Specialization's Holy part

- Count: §Count of the 2026-11-21 to 2026-11-28 reports

## Count of the 2026-11-14 to 2026-11-15 reports

- Sources: 76 public reports on forever.warcraftlogs.com (zone 3000 listing, 2026-11-14 17:49 to 2026-11-15 07:00 UTC),
  every fight Bigger Rock on Bigger Stick was worn in: 17RQWNTLJAgGBhrx, 1PcrGTyatDAFx8Yk, 49dLXZ3g7NhzqABb,
  6H1QkhV2qANXvKFc, 7AmNxqPcC8LdZWtD, 7dWytj1K24G6ckJ9, 8ndg4zGq9mfVWvQk, 9rPwVnazMdWYxj8X, BLXHnFRfrhPGYa76,
  BcDaTm9KbgyXZJLC, C64x3BnYWMQZkA8T, CRgvLNQYd8y3c2m6, FXaN14xP7VhbDpQM, Fx94JmkLncTXBPQy, G7KxnMVrJw19gv2q,
  GTJw7YXampLrWcgM, GjmZfc8aLPFQVXKN, GyZmK2b8c9LMtDdw, L7VBmjgyHRCzqYa2, LQvTgHXDf3Vq1zYA, LyjHDc4nCYmBMRNX,
  PCW1JNXVx2B8nDyh, PDT2yACNzd4LMZ98, QHjqK4G3pzgRWaN6, QajbfyAcJdRCkP92, QxRZFwrkKgJPdB82, RmTM8XL4aBJY12Ak,
  RtZaGw26kvVcAjK8, TFYgK8vGyQc1MzrL, TkfYVFdKBth9C3mR, VThFn29CGbX713cy, WbfQZpaPAkL8tzdv, XvjpYan3qxF9WRgz,
  XyPY7tNcdb1fFDCZ, YAVgyZp9FNMKCh4D, YBDwzFPf9WH7pgv8, YHXAcJxWG89pMFgf, YXgt6fkMdBTKDrAn, Z4w7HKFPvcrTWd3M,
  ZH7zNkqT9j3G1Qm4, Za21hDrqWXn6Rz3M, avm98fMJj3Zc7hP1, b2YKZcFzWPtkxfpg, bcLyFrwNCaH4f9DZ, c6DWA1CV2tdmRFqL,
  cHMQhVq76BR9YFNm, cdFkXVKH1xm2JhpW, cmnDzYT96jLbRyPV, d6TMB1P784WZHVhF, dJa2G7PHZg9wt6QX, dYzgZkCNJFQAXtBP,
  fLyrVbpZAWxzQFYk, fgw1ckK8HTAvQ3pq, garVmPTCWnAh2HfG, grjxzfTdLQKbnBwZ, hRZr7VPfazMjxJym, hatJzXGTk7qjmLPf,
  hf1pF42aPHMv9kLX, j2FY7nBLCcakV9TK, j39RX7DxTqgzWHmB, k26rdYnGDNZ3b81P, m3Fjgrt124A7advV, mpRWCjLgVDxZb9Tr,
  nfLmGxNDqJpyXB6k, ngAGhLYavywCM2F3, ntaAc893bkmBd1Cg, nvXGCNxKYDHrPgjc, pb7dQrAw8CGWjMk4, qKzDTJjwvXM264Q7,
  qP2MRFvJ16NBZGgm, qyVW9rhmwbnpfc4x, tpWza62JNFrTm9hH, v6hC8nkRyQjbgJHX, yYnGv1tHqWDV8bMA, zH3TmcCQRb4Pa8tk,
  zvkWdnQb3t7yCmKZ
- Game build: 1.60.1.70938. The hotfix pushes inside the window (2026-11-14 19:52 to 2026-11-15 06:22 UTC) add 29
  items and change housing and profession spells only, none of the items or proc spells measured here; whether the
  players relogged is not recorded.
- Date: 2026-11-14 and 2026-11-15
- Pull request: #1032

The count is EV-0119 §Method, run on Bigger Rock on Bigger Stick. The rate is the pooled rate of the table, all its
rows, rounded to the nearest 0.05 step, which the owner sets to 1.5 (pooled 1.55). Its proc spell is Bigger Rock to the
Face.

| Item                        | Class   | Hits   | Procs / isolated hits | Chance | PPM [95%]         | Wearers |
| --------------------------- | ------- | ------ | --------------------- | ------ | ----------------- | ------- |
| Bigger Rock on Bigger Stick | Paladin | cast   | 350 / 3862            | 9.1%   | 1.60 [1.45, 1.77] | 28      |
| Bigger Rock on Bigger Stick | Paladin | white  | 530 / 6487            | 8.2%   | 1.44 [1.33, 1.56] | 28      |
| Bigger Rock on Bigger Stick | Shaman  | cast   | 590 / 6061            | 9.7%   | 1.72 [1.59, 1.85] | 41      |
| Bigger Rock on Bigger Stick | Shaman  | white  | 1496 / 17002          | 8.8%   | 1.55 [1.48, 1.63] | 42      |
| Bigger Rock on Bigger Stick | Warrior | cast   | 159 / 1880            | 8.5%   | 1.49 [1.29, 1.73] | 7       |
| Bigger Rock on Bigger Stick | Warrior | white  | 270 / 3474            | 7.8%   | 1.37 [1.22, 1.54] | 7       |
| Bigger Rock on Bigger Stick | all     | pooled | 3395 / 38766          | 8.76%  | 1.55 [1.50, 1.60] | –       |

## Count of the 2026-11-21 to 2026-11-28 reports

- Sources: public reports on forever.warcraftlogs.com, zone 3000, the paladin streams of the 580 reports listed under
  §Reports, all starting at or after 2026-11-21 00:00 UTC
- Game build: 1.60.1.71037 to 1.60.1.71139 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-21 to 2026-11-28
- Pull request: #1287

The count starts after hotfix push 114094, which gave 20111 the attribute that lets its aura reach beyond the weapon.
A paladin's points in the talent are read off the talents in its combatant info, and the two-hander off its main hand.
The count sets paladins with 3/3 and a two-hander against paladins with no point in the talent. A run logged in two
reports counts once.

Retribution Aura's damage needs no spell power estimate. For paladins with no spell power on their gear and no
Champion of the Light, on hits that do not crit with no Vengeance up, the lowest mean hit of a paladin over the
amount the game files state reads:

| Retribution Aura | No points          | 3/3 with a two-hander     |
| ---------------- | ------------------ | ------------------------- |
| 7294             | 1.000, 18 paladins | 1.050, 1 paladin, 84 hits |
| 10298            | 1.000, 12 paladins | 1.054, 5 paladins         |
| 10299            | 0.93 to 1.00, 12   | 1.054, 1 paladin          |
| 10300            | 0.94 to 1.00, 7    | 1.074, 1 paladin          |

With up to 10 spell power on gear, none of 14 paladins with the talent reads under 1.045, against 55 of 100 without
it. Two points read 1.021 to 1.054 at 10298 and one point 1.008 at 7294.

Exorcism, Hammer of Wrath and Judgement of Righteousness scale with spell power, which the reports do not carry. Each
paladin's spell power is taken from its own Retribution Aura 10301 hits, and the hit is set against the amount the
game files give at that spell power:

| Spell                            | 3/3 with a two-hander    | Over paladins with no points                                  |
| -------------------------------- | ------------------------ | ------------------------------------------------------------- |
| Exorcism 10314                   | 75 paladins, 1,430 hits  | 1.055 [1.041, 1.069], 102 paladins                            |
| Exorcism 10313                   | 25 paladins, 352 hits    | 1.065 [1.008, 1.124], 13 paladins                             |
| Exorcism 10312                   | 6 paladins, 128 hits     | 1.070 [1.010, 1.106], 11 paladins                             |
| Hammer of Wrath 24239            | 154 paladins, 4,232 hits | 1.026 [0.969, 1.075], 24 paladins; 1.061 [1.043, 1.093] alone |
| Hammer of Wrath 24274            | 48 paladins, 1,119 hits  | 1.049 [1.026, 1.070], 8 paladins                              |
| Judgement of Righteousness 20286 | 161 paladins, 5,586 hits | 1.145 [1.016, 1.250], 34 paladins; 1.031 [1.007, 1.054] alone |

"Alone" is the same reading in reports with one paladin. The estimate assumes Retribution Aura 10301 takes the bonus
as the four ranks above do. Without that the three spells read about 1.04, and neither reading gives 1.00.
Consecration's flat tick 1280349, which took the bonus before the push as well, reads 1.062 [1.058, 1.066] over 69
paladins with the talent and 236 without.

### Reports

17qJwWpzKx93dNQF, 1DphBamzGfHvMnC9, 1GzyVFrPntN4wqT3, 1Kn7qmzVA8jtMHyk, 1LdQV84zhktHAaTp, 1NxTCjDZWd2gqkAn,
1TKGMnB8ctmLpP7N, 1Tpq8grAPKV7FRt6, 1ZfhArNnJX3gtPcB, 1xQMcH6aTWzCJ8XZ, 1xqRkXMpVBW82CHY, 1xvcXyZ8q2KhrjtP,
2NdqPDAnRkcYKymT, 2PkVvHQCKbyTfAJZ, 2RTQkMVPrZJGj4Xb, 2W6hMTrYqyRtL8wJ, 2fHzBmjCqXwyJLxQ, 2hcGL8FdDgK9xbrn,
2kvTfQHXr7VM1ZzR, 2q1MandA4HXxL7Ym, 2tnVfmyxNJY7rjPG, 2ztdfwHnv4PgYAxh, 34VCpxDkA67yXMRh, 37YWanvb6KqcxPh1,
38brBWmH9RLxYGva, 3BtPwr19KCMq6dfm, 3CXr8BZMjQV2P6zA, 3GDkvFaVHyPwZBbT, 3XmnkY9GfarbgHWV, 3njQ1qNxzC4g7HfW,
3r4pzgfc172HvbwC, 3vqd7j8h9rktfnMP, 47fTkRKLxMVBQrb2, 4Jyxhd3YkT9ABa2q, 4KAvaGncpZV2XxMz, 4Ktph9HWDPzj3c7v,
4MDFjtr2ngHbWK7d, 4XLqQjchbAzvVG9k, 4YhFbCdznkj9f6yX, 4ZLqKxbPXVpTafGM, 4awATLXZNJf2rkB6, 4cA7dB1LRTJnCQVx,
4nNpQ8LTWJVmK9zC, 4rmpzxaNB1nVLGvX, 67fqCHN9ZManpxrk, 6AQy17cKmDxqRLj8, 6GXLaVnCxj9HZN74, 6RNYaMxnb3mQ2XBt,
6RqzaWpYKx7XgjNr, 6VHGFCyBrvNw4AQc, 6X9NgbRChJYmprA7, 6bc8TJd7rCkamNKG, 6hRXTC1Wg2bp8xmA, 6qKXWBx1QvcpkTDt,
6qtBhwN9grpafnDC, 72Aj1QBxcKCtLYNp, 78XZRAyGrMYjwFDL, 7CrXyLht63MGfVb1, 7bntpkR6acHT4f2g, 7cWCr8jm4z2Ld9DF,
7dgQ1JGLAwNMc6rn, 7gCbcw9nrFGZJW83, 7mbkajTnAdPWBXCF, 7pN6RJXVqnFBfkaW, 7xD3CJ4kyjBwdngz, 81DqGMFYnZyRaVtH,
82aGWzBbKRJrjN7c, 86TqPwkhaybCJXWD, 8ACfQgLnHmTMNFxh, 8FKtkJ6ZQvMAdChx, 8GhFACrmbZv6fyzd, 8QBr62LATKxn3CvZ,
8VhQt4NWJ1mrMPRd, 8ZKPH6nTG7RDr3Vb, 8mMb4WKCLkrdpRNy, 8nQDR2Nq1GkLatHh, 8nafFyTwpvHVAc2G, 91QqvPjTdBg8hFLH,
96cP7tanHACXTdJ3, 96hcF3zPjVJK74Xp, 9AnYja6GJBqp7hkW, 9CgtkpAYqGjxP34Q, 9FR6h1P7gXBw8DL3, 9HTwnpy1qjJ4rDVx,
9T1NCP4vxBztFYZH, 9Tythdxr8B163mfc, 9cKfby8FBVZ6jA4J, 9jxRWGhpcFbZ14A6, A2GQnyktwCM6xd8T, A3fPr6hC8yNjz4TL,
A6qcQ9jtTNXKbvYh, A86pf24tnjvyBLqF, A9tLGgVcCBbDhJ7v, AB9mrk6C8vathbVH, ABYfC7d1zrxbwRhv, AG8BtLMJDganhZqC,
AMz3V19cYJDPRGxd, ARMrB9JYb2D1TLZC, AcHqVj39zMG2tBvh, AckpFX7QzHGMPmWT, AgLGDpdPrCNbXQ1m, AgQXzkmVdFcDCJjR,
AjWBvdLHzwRMT1kQ, Ap49agwB2DjCqTKX, ApdMH6bGhQWkjwV8, ApkFaxd9VJ2WKHC1, Axc8F6bfLaqg7BM2, B1WTnYLQqAX87caN,
B1WXm7QGV4hdHLbK, B2Cr1pw93vkndx6K, B7xXw4CjR6mDTp39, B8YhdtQbP7pGLnqM, BA4vwHzdK9CY2ZVN, BJKCH3pRvYwfLQA1,
BK9Jc4DgLQPxAf1C, BPVxqXWLc3FJTpMh, BTGkdCFrYaMXcK6j, BcXxkd4J8ACDvgQa, Bm94aWXPCyqfJ7kn, BrqpdcJZWn87kzKQ,
Btxk2TjrPZ7XYhwz, BvnzVNcQdKkyr31H, By24Afp8wQZn3aTJ, C4Jk2BbWPL9cnQYq, C6VJxRjXQb4FDvqf, C7BVnK4jJNyZMPzg,
CARgtYzynZd4w8Tc, CDpJm842TqvMHWK1, CTAQ8wN2gmDd4bKk, CV1DwP4Bct23AzGY, CWXBLFfMDAcxN8Jh, CXxJY7qTBQj6zrb8,
CdD1GTmxtWzKfjZY, ChAHD6rfYm23KJFd, ChXdtwa6PfVnj72H, CmytZpRw1ghTQDjB, Crphga8WmQ1FkfPt, D1hQVa9WdRwYgT7n,
DBtbX3zxp4NGTwm2, DMcXAzH4Ltq7KFV9, DRyzwVQch9aGtJFZ, DVJNHq2rBQnRTcxY, DVXTdwtYgJRNZBnG, DWjrdMCGpVYQN61a,
DXYJ2tmBKnqvwjAg, Da4mrGzPCLV3fRYp, DfMFk4WnxwYQ9Jjy, DfmZ1A8HwytVKJPx, Dk9nht8jaqMC7JBV, Dr8kfAv94CKQdtXm,
Dxmw1fJXY2ancgW3, F4XV6AnYhjK1zfqR, F4tPJ1naGrxXw32Z, FQ8z7pvRdACD6Vkm, FgkCy4tnvcdDPTZJ, FmCKyHA8aJxG1QVk,
Fr2YcVMyWAgQv3XH, FrtNA7VgxfGHjdKY, FzXGfVJr679TcNRB, G3QzKA8m4dFNv9gW, G6WAPnwKBRC2dkjX, G6fHbyCrcm4RnMLK,
G7YxnWBjNwHaJyr2, G8V6HwM1cTR9nWYk, G9NmY7ARvHKyb8Cn, GKw3dVycHTp4Rz8x, GRATFpytvL61qZWw, GTpK3MPycgkV9Rh2,
GX6Q1VbJ43YNpgK8, GXR7nt6ZCL4DQzxP, Gh6TaJX7DpHPW84r, Gh7apbAxTcXmCdJ8, Ghq9A8TmRBX1d7Pg, GhrqkVf4NKXBdL7w,
Gp4FkNwgRtQm9Jav, Gqvf8dNR3mpJQg2b, GtR3T8NAhXxWdQZ6, H3Pta4JbGg2fMFcB, H3TqYxR1k4phg2yw, H6hnfmvc39wqtC1D,
H74fPNCzKbvtWTGJ, HBTaRJQbMr2xpcN8, HNDXaqgxpjQznYL9, HPVRpGm31KQTbrNd, HTxwDrGLQFZ63h14, HYgxLwn9ZfqyVBN7,
HdZVFxP8BNrA9LKm, HfAcRBLNbtGK38rq, Hfr4C7dw8QK6P9qa, Hk2aWJDcB7zYAPC1, HmzAt2rjDVpw7G6v, HqN89W1GxZmwKCAp,
HqwZG7rfy6dg9RL2, HvKhpmJaD63Ytzgk, Hz8rx32AM97yKPcC, JAPL169xftdn2RGk, JFZ1yXwVfNpPjvmg, JFvtwZkq6yA1TzCB,
JNQHXxFyP8GRbMjB, JWk8ZvcHtRAML2dz, JctKLwGFbVd8QR2v, JfRHQ6XCjhz3b4Mr, Jqck2vGzfa7Ldp6C, JqznAcxDtY4TrGH8,
K1LZ6RcbPfztrdaY, K3fTgQYRjdJAqacz, K3xmjRCnMpV2c4qr, K69R4GZ3xdnwcMFf, K74YMcdmbN3X1qLg, K7GkFZhAdPvBW1g8,
K8RQY6qLxA7aVNhC, K8kwZPmy9LNxfGYM, KA9GJXfvPqdZ7Btg, KBnjRA8TXQxJwNtk, KNYqn8bcp42rgfJD, KNkCx8pqdvFrmLhJ,
KPz6dmYkQM1B7tNC, KTPmpLFv2arCZQM1, Kafm7RvyBTpnM1QZ, Kg71YzrL29FaDQwV, KnbcNJh9LVMQyYB8, KpGVLtw2m4abgj9y,
KrWLQhNymAx4b6HY, KvZ7f3GLJQ9MVmPB, L4GbPrC2hv6TRckQ, L4GgDdhCHJzx7qFZ, L7fPJM4Fq9Y8mdVR, L9BdJqGgZFpkCtnr,
L9fXHdaK3cPt7VvZ, LaxT1XFh98HAzpPR, Lfh9P4J61VytWGkw, LwTNrdJ6zjnmCKZg, M7rbCABmXtVTZH3v, MH9gFnGK1bvaYhNA,
MKjXLWxn7t2maqd9, MXDKc6LGpHq3k1vf, Mdy1jhwxF8AtRTKQ, MpaBFLAfqTb7dzcY, MqGj1HJfX7CgrP32, MqmRkWz6ZFwAfKxp,
MyZGhvYbCJVtdj7H, N8wyX9FR3TnPBpx2, NAK8CyqwtV2RbB97, NBJvV6WCg9ktZrHD, NC2m46DxFhBz1vwd, ND2ZM8dCmFBXKWnz,
NGDTVLr7vzQjnCYt, NZ8jY4wW6vTAcPCJ, NZHxnkw4RQ2T7Wg8, NfVtCBWXdb2c9r1Y, Nn2GzpaC6rh93Bqf, P9jxQmbVBwKYdqCM,
PC1M9hHgNxzn7VjF, PD74RWzd1vAmYV23, PR1XaHMzYL6m47Gq, PR8dn7gT932ZbK6H, PVHbqvaXgcWGF6hf, PcJyYBT7WGMqpLm8,
PdFTZN1kDpWxtyfa, PdFynjQZ8aA6hgqJ, PrbwXCWMymjL8Hak, PvTg8dk24FwzX1hZ, Q4qpAfNLdWx3KT6h, QBdJpmVM31WY72fH,
QHZCpDWdmPJBRFLt, QRGNbq3pH8ZKjJAB, QWGm4XvafMDhkVJY, QawpRxqVhrA4NGXZ, QdwJD6mHNVfBxrFp, Qn4dj1wYXBqD3m7L,
QqLkgPf3Z4dznBtr, QtgCbHWcFXN6jMAn, QvXKRxd7GBh4VjwP, R6ybrhfFWG7gvtJ1, R9JmTLqfBQwXhH6b, RQJTXpABKkaFxhjt,
RQLP96W41TZpMDxc, RQdWNhKTP2qBw7nA, RT1Nz9qP7b3LJdW6, RT416mPWkvJFgM2w, RWBahpJkAgVPGNFx, RWjNGhpVvJ36c74n,
Rfkdb89YQNxyL62M, RjKZgWQXzCDy96fB, RkqGcKZ9TBv1NgpA, RrqyGjMmAxHJZL6g, RyDHdZQYVB2PcLqf, T1zFLdXyHm8xWcDZ,
TDgxaVBAWQ96HdXv, TK24w8Fvpc1xDbRr, TLz2dBaAhGmHfgFD, TN2bP8zQdyxVBDgR, TXwZBQHj1Y2LPM4C, TY6PgvCzwMt8HJkD,
Ta4fLJj32kWzmXgw, TcNp6AFaCRXkdBVL, TdZh3tDzaQ9Kx2VW, Tm42j9AJcV6hrbC8, Tqwtgd2NYv8WrDba, TrARcLPHq61MVg4n,
TrGWBmHZ6p9DwvaJ, V1XQN8DmnyzRakKW, V26p8tfmBF1KdgN9, V4dBhLwD3YZpgPJk, VDHGAnJ4FLvw3TqW, VGzkZ8tdYFRwfX2n,
VHWwATjhrxtnzCyL, VY1qcrJ4HxjCFnfX, VmgqPWGND3RztJBc, Vnv8MY4q3ZKmypg9, Vtb8Qkyr7A1zpCKw, VtvCpR9TK4jHhNwc,
Vvhkgc1rpfDFjLZX, Vy7hfAMHLgkcWrqw, W28BmZap9hvFkt1R, W6NMxX7BkTYQdr1v, W9RBymHfwz7TK8Gq, WHYFqtknwJ4Xg2fK,
WQDN17aAkMT2tf4q, WRXxpjZcGbt4C16d, WVj7gKrz2tT6yfwa, WX9qwgrB3xGph7mJ, WaptR789mXANCh3v, WjDJXCHz8kwxTpQA,
WkQXqbMK7JjV4ABd, Wq4XVPGA6CvpZwxR, WwPYbgvAQr6pNVMd, X6krbCcwV3ZHLp9J, X73w8hANRMqvam4k, XJf6ZABVhHaxgmdy,
XKfM2vAnWPHtq9G4, XQDYTCwN8JkqgVWa, XQmMHk8y1JxC7Kq3, XqQYd9BpCGFmcjJK, XqpfVay26LGdQmkh, XwPJHVmAjQh3YTF2,
XytmY8TQFjpC7fKP, Y29j4KZCAcvXy7Hx, Y6LGBHaq8AVfjQrn, YFkDQLjn13b8vyfr, YTyBWAPjJVN9gmMF, YVdTBKcmMqGNF1yQ,
YZ1wNWbFA9H68Xrp, YabqrnBWdgQLpx3t, Ydbat6hRBJgDy9G2, Yg1xfCnHTqzbJNyv, Yyrb2WHMXnjg4qhk, ZDaJNBWT69jHfGX8,
ZNYqwcFVjxz63ABk, ZcC6xT3QhDgrHdKL, ZdQAV3Dy12W8vB6R, ZdngxT71CyYaFRMv, ZrYG8vXWTC6F4yQh, Ztd2myTWRM8BYNQ1,
a3XBGfYHjFAwZNdy, aMZY9nWkmKByVTAv, aNYA9PzjCMbpGtfv, aRLP2JdQntKzTx3f, anAt8MrDB3kpQ6Lw, ax8YgMTZpWn134fJ,
b6M8QqVjYpTNa1Dm, bFBk93mWvrMYTzP1, bXBgzWNVT96xjKaG, bZ8fqNxvWBwMnkXa, bdK1Xvg36T9Lrmfc, bfWDTQPjnt9Ygd2K,
bgycvf8pzGCqNBDm, bkvZtGC3WaAmw9rg, bnGFMvrLfDYqBN1z, bp8jCcfTa7g2P3xX, bpAFaRQj4BJ2CwZ3, c1p3Cm6bx79GgZjq,
c2HF3NWbaGxjfD47, c8zBgFj4raCKkGL2, cPn3g4AVrzFC28Zp, cQmtM3HFqA9h6vRK, cVF2KvqXwt37rfJz, cWnGN8dJmBPR4FwD,
cXL1Zp2FjNDnPdTH, cbnDh3GYMVX29xCd, cfJ82BVhmCWjwGKk, cmjTnH2apGryXR1Z, cvRt64Ldnx9wjXFa, d1JDaQ8q39CTp2gN,
d69tmxLbpG7Wq3kh, d92kGH46DPbmJAFW, d9zp84MZF32VcjCa, dKX2RrGT3Y9hnF4M, dKf8Nhr4MJkQPzqw, dL4trnyq9paPGKJc,
dZ9XapFTVWH14MJ8, dgxVz1N6JjLk87Hm, dh91QyRWVkN2aZXF, djZxw6JQNvF3p9tb, dk84NYFCVJaqL2BZ, dtCm3xkDnZ69GAzy,
dtqaDCLhXYb6BM1r, dwXv9MZWLKGFJBry, f3nGB2Fqc9zgQaWD, fGN4YpAdHJgVTy8D, fN7tCbBLrch1mWFv, fNB1YKqMvkmnw49L,
fVaQLr7wXTDZRdHn, fZQT3VFYG8LMjpNW, fh7tXKrzyw12dLW4, fjqGhPJdTmDxwH4V, fmkdpwVvgcntZxQM, fxpmV6DvZFQP8Hdt,
fytcnzZ3q8wJMp6X, g1YV283cxqG4pWm6, g1jbdaw47PcBWfCp, gCY8xcNAV9JkFL1Q, gDhPdH67mBMwaG3Z, gKYxMJ891D4CcrTn,
gaQPAhN7zfcFd2XB, gaTQH7Jq6fR3XVGk, gdN1cBza69wJGKZf, gm1dVZJCztNTwq8Q, gtpw7dDWYBfZacz8, gv9V6Ty1Qa3MBZW7,
gzXV4ArFDqyxG7Mf, h2FHr9mngGdYNwf8, h78XbgTK2zvrjfkH, hCm9xAQktZgq27dc, hDwjHp3BzvGNKQ47, hFd38mrtGvzTXMWV,
hHVqt8Wjv2wk4bmg, hR9BDKbcPFNxqVka, hRA8aFDXMVrw4gbq, haAMFPjKd3xvNpJr, hfdjyxV9GLrBQYcA, hvDbwf81aPNry2pc,
hxp4F8HyNrQnBG6g, jTK2ZJGvdg1Qfn7D, jXRnGCBtVYDqfQZK, jY9zAXQnLkGqwh8P, jaVWhxJM7LF28dR6, k6qa2JPBHGvYdWnT,
k8tr6TpNAzRdanWX, kBQNP7Rmxwvr3tY1, kBQdFHLZDJarNMjT, kBq6hmt8xCAvZrJy, kCAvWahQ6n7yd3rZ, kTKH1qFZYJhWvXVQ,
kX7FtxzjfYnw69TB, kXCv8JGrBKRj6QNg, kXtwmgzJ8hT4rK7d, kaV6Qg34v2xYRCT1, kj7ch2vw63DrdFLz, kpwtR3yJmZMVCYrf,
kqZhtBzNx4cVRYwW, kr83aFhpMDfRyzvT, krybD6pn23qctVAv, m1APDyT73GFKd2wL, m1GhaMXgbHPLn387, m4dCL16tqnN2D3hW,
mFNr9v8BCQh3f24g, mGaj83FNknJHpBvX, mTk2BJK7FLQWwcC8, mY9aApqQ6NdHjGLW, mdwkArzCj27X41DZ, mfA6MCQkPbFdRrz4,
mhA2QR4BP6LN8FYf, mj1kGdT2tJNrqR8K, mpPxkJzbQyLfrX96, nRawyr8fVmPg1z2K, nTz1vgfdJYW8XpKF, nW8CDGVNP3xhtAa6,
nYaVHwrBD3Ffbv2y, nYtvxcMjAaDLpzf9, nw2xkAQPabFB8Rgd, nwzB76QmRjhX1TLC, nzcBtpgDa3dLrTmV, p27aftcMZRrQWd1K,
p2D4K8RMktFVgT9Z, p8jYznr1xbMakgLV, pBDLzZTdxGQmqNCR, pBxZRYvLNfVDn42H, pHxtWjvhNAGq4Dc6, pKf7NdLngQxP13k2,
pPQD8rgT2tdaYxG4, pPqJfKV2k748Ldgb, pQR7HzLhDNWTqKZ4, pgqkzDfhVT3mtndJ, pkGMLWf49anFx7yA, pkPK1vjc2t7rTgZH,
pqKWtfN2n4J1TGwg, prGRAQNfTdxFBk72, pt1N84zWcXAZnCfY, q67Y8Pb9WKNzAyh3, qLMkzwp97H8rAJWQ, qQJPwYDvVcmWnh4z,
qTBFXxkNMavY92b6, qTNVfgKrcaR2BH6h, qbRGkr2J4yQgaDxY, qcN8ZYVRbvG4k7AK, qfhXJ2rAxyCkTvHG, qmYP9J6HByRKFWvX,
qrCmT4HYWA7KLRkB, r2dNzwXQ8nfkjgBq, rApMHbqFCyLh87GV, rDQvKCTYfVxWj1Xq, rHLfPxvRCX83G79k, rJaTCp796xwvqL1g,
rakK3JgWVztdnMfR, rdKvh7BR63YzwNWg, rdq9znPJg486tAXj, rdtjnfKcwkvN4Zhb, rfgDYKah8wpGbxnX, rvPn71AMTwajKNdG,
rvc39NCA1dpaGKFL, t1yGxzP7L3Aa6CQr, t7mDkLy2G6fVhXHJ, tBVKZzdATj8kx3fF, tC3vwdKyajnPxcfr, tGRZHM9wFgNYmf1L,
tHmvawdXAzRC72qD, tKWDPrJAxYvNLqm9, tLbxT8vzGVdh2jXJ, tMgf3AKWjBChp9X7, tPB1XdqVKxQyTRw2, tQ1ymLaFbAT8c6KV,
tZY2KRFTGfyh9WMV, tmbaz3Z1JxMA7RYH, tpVb8fFJM1aTNQcw, tpgAXh6aLnFDKyQ9, tqxDCdPfNkXKQjyp, tqxpQJPZDHkLA2Ka,
v2t1VkGxK4Dm9Yp3, v4ynHhcF829VprmA, vBkH8W6NVg17DnPX, vGW4kA6pNVwn1Z8c, vJmHxwgrWjR2Kfnk, vNJaB1t9d2hTPxAC,
vTW2MZKhFA9GcyaV, vgxq8Hkm1X79MaGJ, w2F1fVA9cY3djhNJ, wADJ74mV2NdcCLpX, wBL8JWmnbtr3zyNa, wCFPYVcykNZQKL2t,
wFgnz9tGRTMYK4Pr, wRBLd8bjVnvMqKZ7, wWX2379x8aY4C6zf, wXRCbqdLt6mAg8aV, wa1LDhZMzWyd7NCn, wafLQr31X6VYCmNZ,
wr3M9DfXFQZTqj72, wvtyBRzc7mXfbZV6, wxR6KAYyncqvf8HP, x1WVhMmGHqrXgJK4, xFY1LmP4twqnyGTR, xGyATg1HbPpRDLqW,
xY1MJ7WcVCGHQmgZ, xbRZTD2aYhHFXmGB, xdZBDCg6aVFfhr1M, xnRmFTKtYdJCGPaq, xvHCBa46VGXD2L8w, y72Cf4pTxQ1WFbaV,
y7TJp42XGtDcRWbq, y7qaY9Hrc4QxgCFW, yMbqGLNh6X2fHC3D, yPJ6pm4Q8M9qfLag, yTF7wAJ8DXRW6YnZ, yW4DYmfQzHwXtRV1,
yaLC4tG2WN9wr8JT, yaWvJ4Fj7YCQZT8N, ygMZKDTYd4qCQ9mP, z4DmgHWGP8RFr36q, z7gjr82BtZCmLQvX, zDjbNBLmTKqy27QP,
zKNw1PR6Bg84ZqjT, zPGf2QWgj4rhZXvL, zQV8vpGTDNr7j6YC, zRHK2GChqwxZW9VB, zW8FZL4Q9gjJT2X1, zXpWRQjwqZ1xn946,
zY7HNRDbc8ACwLp6, zdXvb4R2NtpVJ3fk, zmb3Ddp2ftWgXZvk, zqDZF1raYfbp8m47

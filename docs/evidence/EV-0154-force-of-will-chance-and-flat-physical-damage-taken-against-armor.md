# EV-0154: Force of Will's chance and cooldown, and flat Physical damage taken against armor

- Claim: Force of Will (item 11810, equip aura 15594) rolls on each melee hit taken that deals damage, a white swing or a
  melee ability, at 6.2% once its 30 s cooldown is over; a hit that deals no damage, a miss, a dodge, a parry, a spell, a
  ranged shot and a periodic tick roll nothing. A flat Physical damage taken amount lands before the target's armor: its
  buff 15595 takes 51 times the share of a swing that gets through armor off the swing, and Gift of Arthas 11374,
  Tenderized 1292008 and Lacerations 10370 add their amount times that share to a swing on the enemy that carries them.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports the Reports section lists; all fights.
- Game build: 1.60.1.70862, 1.60.1.70888, 1.60.1.70938 and 1.60.1.70959 for Force of Will, by event time against the
  time the CDN first listed each build; the Game builds section has the procs on each. The Tenderized reports start on
  1.60.1.70790. The hotfix each player ran is not recorded.
- Date: 2026-11-05 to 2026-11-19
- Pull request: #1114

## Method

Force of Will's wearers are the characters with item 11810 in a combatant info snapshot or with an application of
15595: 278 characters, a character being a name within a region, in 370 reports, with 19,160 procs. A wearer in one
report is a warrior 179 times, a paladin 159 times, a druid 78 times and a mage twice. A fight logged in two reports
counts once (105 dropped). A proc is an application of 15595 on the wearer; the buff lasts 10 s inside a 30 s cooldown
and never logs a refresh.

The buff logs after the hit that procced it: 16,080 procs have their nearest hit on the wearer 1 to 45 ms before them,
2,802 at their millisecond and 60 up to 5 ms after. A tank takes several hits inside that window, so the chance is
read on isolated hits: a hit with no other hit on the wearer within 60 ms, counted as a proc when the buff goes up
within 45 ms after it. A hit is exposed when the wearer's last proc is 30 s or more before it and a later proc follows
in the report, which keeps the trinket worn on both sides and leaves out the stretch after a wearer's last proc.
Counting that stretch up to the last snapshot that shows the trinket moves the chance by 0.01 points.

A white swing is ability 1 from an NPC. Its hit types are the report's: a normal hit, a crit, a block, a crushing
blow and a glancing blow land, and a miss, a dodge and a parry do not. An ability's kind is its
`SpellCategories.DefenseType` in the game files (1 magic, 2 melee, 3 ranged). The game files rows are read on
1.60.1.70959 at hotfix push 114196.

## Game builds

The CDN first listed 1.60.1.70862 for product wowf on 2026-11-07 04:43 UTC, 70888 on 2026-11-10 15:10, 70938 on
2026-11-13 01:36 and 70959 on 2026-11-17 15:19 (wago.tools' build history). When each region's realms took a build is
not recorded, so the events after a listing ran on that build or the one before it. The reports come from the EU (180),
US (143), CN (44) and KR (3) regions. The first proc is on 2026-11-10 07:19 UTC and the last on 2026-11-19 08:12.

| Build window | Reports | Characters |  Procs | Procs on damaging white hits | Chance |
| ------------ | ------: | ---------: | -----: | ---------------------------: | -----: |
| 70862        |       2 |          1 |    163 |                  98 of 1,955 |  5.01% |
| 70888        |      11 |          7 |    604 |                 392 of 6,279 |  6.24% |
| 70938        |     205 |        152 | 10,899 |             6,592 of 105,360 |  6.26% |
| 70959        |     158 |        149 |  7,494 |              4,832 of 78,912 |  6.12% |

## Cooldown

Of the 18,756 gaps between a wearer's consecutive procs, 4 are under 30.0 s (6.8, 6.9, 18.8 and 29.3 s), 283 lie from
30.0 to 30.5 s and 1,939 from 30.5 to 35 s. The row states 30 s (`ICDMs` 30000).

## Which hits roll, and the chance

Isolated exposed hits on the wearer:

| Hit                                           |  Procs |    Hits | Chance |
| --------------------------------------------- | -----: | ------: | -----: |
| White swing that lands and deals damage       | 11,914 | 192,506 |  6.19% |
| White swing that lands and deals none         |      0 |  31,658 |     0% |
| White swing missed, dodged or parried         |      1 |  85,746 |  0.00% |
| Melee ability that lands and deals damage     |    845 |  13,780 |  6.13% |
| Melee ability that lands and deals none       |      0 |   2,154 |     0% |
| Non-Physical melee swing that deals damage    |    549 |   8,712 |  6.30% |
| Magic spell that lands or is resisted in part |      0 |  33,480 |     0% |
| Ranged shot that lands                        |      0 |   3,684 |     0% |
| Periodic tick                                 |      0 |  20,908 |     0% |

The 95% interval of the first row is 6.08 to 6.30%. A swing that lands and deals no damage is one blocked or absorbed
in full. With those counted the chance reads 5.31% (11,914 of 224,164), blocked swings read 4.99% against 5.50% for
unblocked ones, and paladins read below warriors; on damaging swings a paladin's blocked swings read 6.09% (2,041 of
33,496) and its unblocked ones 6.12% (2,850 of 46,531). The non-Physical melee swings are the report's ability ids -4
and -64, a mob's swing in a school of its own; 248 of their 549 procs are on swings resisted in part.

The chance is flat per hit and no rate per minute. On every landed white swing, the damaging and the others together,
it reads 5.0 to 5.8% in each 5 s bin of time since the cooldown ended, and by the attacker's own swing interval:

| Attacker's swing interval | Procs |   Hits | Chance |
| ------------------------- | ----: | -----: | -----: |
| 1.0 s                     |   801 | 14,903 |  5.37% |
| 2.0 s                     | 4,090 | 78,248 |  5.23% |
| 3.0 s                     |   289 |  4,710 |  6.14% |
| 4.0 s                     |   267 |  5,124 |  5.21% |

The override on 15594 is 6%, the whole percent nearest 6.19%, since the field takes whole percents.

## What the chance moves with, cause open

On damaging white swings the chance differs by attacker more than chance allows: over the 221 NPCs with 200 or more
such swings the chi-square against one pooled chance is 497 on 220 degrees of freedom.

| NPC (id)                      | Procs |  Hits | Chance |
| ----------------------------- | ----: | ----: | -----: |
| Bloodaxe Worg (9696)          |    45 | 1,146 |  3.93% |
| Spire Spiderling (10375)      |    79 | 2,003 |  3.94% |
| Bloodaxe Veteran (9583)       |   121 | 2,790 |  4.34% |
| Chromatic Dragonspawn (10447) |    82 | 1,890 |  4.34% |
| Shadowforge Senator (8904)    |   164 | 2,238 |  7.33% |
| Blackhand Elite (10317)       |   129 | 1,747 |  7.38% |
| Anvilrage Footman (8892)      |   134 | 1,782 |  7.52% |
| Firebrand Grunt (9259)        |   162 | 1,871 |  8.66% |

By dungeon it reads 5.69% in Stratholme (1,083 of 19,032), 6.07% in Blackrock Spire (3,893 of 64,183), 6.10% in
Scholomance (1,621 of 26,559) and 6.70% in Blackrock Depths (1,995 of 29,794). By the swing's size it reads 5.93 to
6.35% under 400 damage and 7.17% on the 12,560 swings of 400 or more. By class a druid's damaging swings taken read
5.96%, a paladin's 6.11% and a warrior's 6.39%. A report carries no mob level and no wearer level, so whether a level
term sits behind these is not read here.

## The buff's 51 against armor

Unblocked white swings that are no crit and deal damage, one NPC on one wearer in one report, with 15595 up against
down, over the 1,147 such pairs with 20 or more swings on each side (58,639 swings with the buff up, 143,699 with it
down). The swing's `unmitigatedAmount` is the same either way (1.5 apart), and its `amount` is 20.2 lower with the buff
up. The share is the swing's amount over its unmitigated amount with the buff down:

| Share that gets through | Swings in the smaller side |  Cut | 51 x the bin's middle |
| ----------------------- | -------------------------: | ---: | --------------------: |
| 0.3 to 0.4              |                     30,090 | 16.9 |                  17.9 |
| 0.4 to 0.5              |                     25,622 | 23.5 |                  23.0 |
| 0.5 to 0.6              |                      2,120 | 29.2 |                  28.1 |

A line through the origin fits the pairs at 51.6 times the share. A cut taken after armor would read 51 at every
share. With the NPC held fixed the cut still rises with the wearer's share (a slope of 71 inside the NPCs, 1,043
pairs), so it follows the wearer's armor and not the mob. A free line fits at 80.8 times the share less 11.7, steeper
than 51; what bends it is open.

## Flat Physical damage taken debuffs on an enemy

One enemy's life is its events chained while 30 s apart or less. Inside a life and for one attacker, the normal white
swings with the debuff off give the share that gets through armor, k = sum(amount) / sum(unmitigatedAmount) over at
least 3 swings. A swing with the debuff on then leaves amount - k x unmitigatedAmount over: the debuff's amount if it
lands after armor, k times it if it lands before. Lives with an armor debuff event on the enemy (a name holding Armor,
Faerie Fire, Recklessness or Shatter) are left out, and so are swings within 100 ms of the debuff's own events.

| Debuff                    | Row's amount | Swings | Left over, 95% | Mean k | After armor | Before armor |
| ------------------------- | -----------: | -----: | -------------- | -----: | ----------: | -----------: |
| Gift of Arthas 11374      |            8 |    444 | 5.18 +- 0.06   |  0.648 |        8.00 |         5.18 |
| Tenderized 1292008        |            1 |  6,354 | 0.71 +- 0.02   |  0.739 |        1.00 |         0.74 |
| Lacerations 10370         |    1 a stack |    222 | 1.28 +- 0.14   |  0.641 |        1.99 |         1.27 |
| Gift of Arthas, Auto Shot |            8 |     21 | 4.28 +- 0.92   |  0.650 |        8.00 |         5.20 |

The left over follows k inside each debuff. Gift of Arthas leaves 5.11 over at a k of 0.6 to 0.7 (405 swings, 5.10
predicted before armor) and 5.96 at 0.7 and above (38 swings, 6.11). Tenderized leaves 0.38 at a k under 0.5 (56
swings, 0.31), 0.56 at 0.5 to 0.6 (60 swings, 0.54), 0.63 at 0.6 to 0.7 (379 swings, 0.65) and 0.72 above (5,859
swings, 0.75). Lacerations leaves 0.68, 1.27, 1.80, 2.62 and 3.19 over at 1 to 5 stacks, against 0.65, 1.29, 1.89, 2.58
and 2.97 before armor. The Auto Shot row is ability 75 from one hunter in one report.

Gift of Arthas' swings come from 19 attackers in 9 reports (2026-11-11 to 2026-11-17), Tenderized's from 483 attackers
in 245 reports (2026-11-05 to 2026-11-19) and Lacerations' from 19 attackers in 8 reports (2026-11-11 to 2026-11-18).
Spell hits against a flat spell damage taken amount, and a flat ranged damage taken amount, were not measured.

## Reports

Force of Will, the 370 reports with a proc:

kW9V4CMKfmBbtrRH, 1b7PdYR4HhDXBGT6, BYyqMbfxpZ8nFhJg, LQG6jWCnAP1gwvdt, Z9y8b6WCft3R1nzh, jT7NDMWFfhKk1bdz,
NPHc3azxL8bmgdpF, NRK61HZcDdJkrwXG, 27jZGhfHgp1LR8MB, HWC2c1Bq3kMXz6Ar, YJPVatA7TqpZjb6B, ZMBDHxvV68jPdQk7,
PRrfKLbYXQdqzDjv, 1hRZ2v8PwKWt7DkX, c6tMPWv4fG2Vb8HQ, Mg2ay1xDjkYpBqRX, FMzkb2KrnWCYt3dZ, F3BTZGrW48Mp6n2L,
VT81WfMQFg4z6Cyj, Hwq97xLh6gv8ZPrY, dfAPgkrq3mJNcVvt, rPpHYfq3tNnwJRMd, byvnpga9tMq3BfX6, 7XK1DyrJ2TfAQ48p,
8jzXTaqr732V9DxJ, 1TZvYG2tfR9wVPNH, M2RHxFDymYWLnpPh, bg169rGdXTzDhLnj, 4nkW7JDxV12waMfr, JArH1Ky9zdhVk8RN,
CWhG6LyXk3PBHdMQ, RTfqrCdVvLGYzj1A, tA13MXZ8pLDjqcxG, z9XbfMRh21Fg7QjN, x3njCmpkVYf7WDgh, 3R69PBjxZvcfbnX8,
QNwkdxbMvWmrqjAY, 4DmjXGdbBQTKP6zr, 9zxHaWbYT3gqDQvF, 9h3cXRZ7LfM6gPwb, FVqQab31p7KXAJzt, aFMwqH32y4DtGLTN,
HxNMb7JVw4pGP6dn, CVfpgcNPrMqXh68b, pdkwZJNqWKCarQx9, zCyk3LFp4XZwaQBd, nmwM719fpABxvjdz, K3FqtCdPbLYxBzM8,
PtmYHjkDrqhdR9MA, D463FZBX1C9maJ7b, KG62VXF3a97CxtAg, tCnKQrPVv9Z6qdGX, J38xbrFytgM9p1LT, 6G3gb8DwtJFXm4f1,
xqnmW3BDpFb28cJ1, VYBjgJx8aZtRhmpy, zmDY4LVX7N8vh9aR, RtZaGw26kvVcAjK8, BDFj9VNrfXhxwQT3, GZzaJH1QVjMKd8qA,
zrALwGH82WM6Jcbx, NXMZV9xjvWpB7GtP, 8GNt2hC7v1rWRXyc, h3VmpYZ4Gt6xNJC7, QAgrxLyh1CVwcf6k, vZqVmLYQa6jwP3A9,
QczFgrTbG4CPpB3n, 7mnVXCgpyxkbtZGq, 2wMznDRvmkpbtHxA, Z3qdRnrYyXtkhLPM, hZGvF2bkQ37gK4wL, hHZTNk2zgLjDBMx7,
xTdW6NpRYKrM1Lyb, MR8g1PhjbDN3tpkv, nJqw7xa3BHptg2hX, AdfMXDLW29k6pGy1, D39jWq2CaRgbh7Jv, TA3pLxKFXQ9Ckm41,
T79Cvazwp2V8GjQX, 38bwApgqTf2rh9Pc, 1a3Qn7JpLk8AH2hV, FTQ8g6cvbyZ1wB3m, X6Hp4gDwv7zftxMC, PdNq3JxntCKgavkm,
jTRbQzV4qyMvxp73, ZRHAwVNzx7MkyTtK, 7ChZtckBQRfp3KAD, bF8hVwJNDT6HyRWt, gcxh69fvXWGtbRaV, 17ymZpJ6xFPf4Raj,
bcp2G1PHy3KWNM9D, Wv4QYA6C7antxKXr, pwmkX48DKCGFZfvn, VjrAGFkTvNy2W98B, FkBxY9zn4faM7jWL, tFm4dXCGDwznQ61a,
8zGhrb1FcV2xnBQW, XQkavwVbDrLKmhJ9, j6XFRzHxTC1kvVpy, dvkzFtPH3WR9C4Km, Mxd3G1mcXFv4RfN2, Daj7G8xFX3bA2CmW,
4QJzfYwDkXFrbxgB, h9nNL4WmZfjHzvB6, 3mwdt9jJF6Tyb1q2, GbvNW3tAP68ZpmcY, yHRFJ8kn9vQA4YLr, GRnMHmK69BXfWVTj,
2VpbrfHatzjZFkGP, FHqYCNcKntvDAjxg, HKY8mdFptfNkDgwb, ZDWyQGqFBvagT7kA, YCbcQHW378fnMvXA, HMdLhWY3pNbnXPtF,
MC7jkfvrpcbyTQ4x, q9dHhDJMk1btXpg3, RVkGfcrbZ13MWqQ2, vGJMahXtgTzFRmCQ, X8PH2CG1hMkcfxYJ, v23ytJHDz6QkbXnN,
VK83G7MdBhLRFAfa, aHKFrZvJgBypW24m, gmKvNqHFJLZXfdnP, JavZtXcdp9WgLQ8x, CDKVcP6ghZJtv8AW, DW9YkdvyrjtMBQnf,
HkfzanMQcVptAB19, B3MA6RjphHCXKQmW, CqM4Vnh831tFgcYw, YVWgj6ra27PvmNnx, 1JLq46yg7wYNhan9, MB1yTdQ9xmRXPaK8,
nJDNV6Md3bqpXZT9, 1z6PXMGdQBDvrCqx, A2YGVnmxNFPqk7Ka, ARFXzbqHDpBLJdhy, 2Cg8R4ZMz7HV3rfN, M7cz2xknrqXVAKjD,
Xy3tw1gBKbmqZjn9, tGBMqP9JQDj1vCr3, DH3dT6tYLKmMrNz8, ZG367WYvCXqhr2ky, DNc6QALhxmGbCH3Y, jrPWFdYfQNz3xG6C,
CxTmQRX2ytGcwA9r, fNxhtGQYCBnPbzcp, BkH7c8WTzJVCQqyZ, QD8ypCXWJcv9Y2B3, gpXqDYBNbxKc1y7W, Zxvmb47dcjWCQAF6,
aqZnYb67BPjLD1mg, dCFnA2kaP9MVYNhc, 71aKdZW3xFHLCT8P, P6GLNM3n27dtcYJD, 2hG49JzynDkwMVCY, hamTvL4wFpx6nYNj,
pDKQwPRxqjzNkbc1, naJFzXZPWbR9QAgt, tTRm37A4YwgM8nCG, qprynNbXtRZV4DHP, xzhQVR3dHa7vWZ1y, jBwLZh7p8QVNkX2T,
tHzN3kKbaPpVW8TY, j7xyJdLXHZNA6PQg, zNTQ4WnfhpcYk9MH, gBK1ZQAP9aDTMhJ7, a6GB8AKFtcMLTrhP, hWJLtqnXbzwG1TMD,
N9PGBMt7faY6W2J8, tkPZYqTRgCyWG3jM, wqaZjCkvGPMDt28V, 8xjZmDw2LKv9bckh, 8rAWbkX93chqKjFG, FJdrDz9w1Vxaf2cT,
8bWVvtJRXFQNTAjc, zvjKa8hknGcLy7ZF, 1TBKn6pMtbQHfDdw, 4CVrTcXtBf2yAYH9, R1g7q2JAnwzdrPvc, WhPpmnzB6rR3gD4Z,
9GJNhYFB1QnKqzty, 3H1fvFaqxRL79PY8, C2xQXzmFg6P1Vaq7, pMkPWzaC7wf9XZNt, 4y6Qnzkw7GHKcVDb, 1VpHn3kYq6B8JMWw,
ykz7Bdp9avqbAjRV, Wy6TdamAk4ZxJ8MC, P3QtFzTLVp1aWBbA, CfdHgJAvtMRhXNGQ, WMy9axzfg3jKLhYc, wDGP2BrqpYvgQxRc,
rDYW2wpQ4bBncAy1, 26AYbFNQzMZCmfkh, q38PjpNH9WmL6zCd, KvQYAx8bW9FpVXdB, 4hQkw2BmKJGd1XZ3, ZKphvLAPfVxNCX6z,
rqjzTHacGgmF27Xx, 4pynYMkBdVgRDvKH, NphMF846ajqQWAZg, ht9mXk7AM24JdGRH, 2WkFQGcDVYwpTMJ9, 1Y4NWLka9pAJbB6G,
jkcbfRxrPgvWqwyM, 1GWa2FTw7BDnhA9v, DBzk8qw4yFWtb9rd, 1bgXr3HRqG2cTnjQ, p6mVRBzvWMqnQZFc, JQwbjnVYyGcP4rg8,
ZhbCF1LvkQ9gBAMy, ztby6Z2W3nJmT1NA, pLYdFf8zbqwGrcZK, p4waHBZgjkA9mDLn, pFH4amZPxkBhyTtV, R78CFp1D24jAatwZ,
ybf3pCJXnTQD2V9G, g4KqJfjzTN8AFQCr, nbLKdYfXHDam79R6, DmjBrYd7M9VZwCHF, B1qcTGpnHCrQ64Xa, 1MA3qWLGa8v2hdXP,
DyY2txcNwvGk4qpJ, 4ChPnBrVK2fJLMaX, xdFjKnGbC2caDhPJ, tmd2HZRwKyMFPQqz, kz4FVdBRqpJwj8nx, 16GhDNRmVn8KWjBT,
KcLtDQRdjbrxyqXJ, GHK2gB78FMVCmPfW, fqNBazArpjygDXhZ, 8mJtz4F6gTfpw1Wq, ZJgFdtvGTYQVCqyb, jKQDqkwmx46ZLBFg,
RGDd1M9QCkjFpymY, 7KZzkbdXjBHGWQJa, zwDb3yF98agvj1kA, HNtn34KXCMyDrQJj, pzBvhjtCD9VRf2mP, yCphrBFqdjgb8Nm9,
Mmp1crXk2BwRb97D, v19wAVhZBPacHQFL, k7j48fCpbzDKXwLc, F4kZ8RhXvzx1CQDG, wj7k1zy3QLmBNY2t, B4QqHnGXwzpRfaYW,
jfHwWDKtgdFxYznh, LQ9zMDqpmknBgAX8, dyNbRtmWjgT8xfMB, aYyJ9hdFwXKgDk16, pQWjBPM2XZ98CA3n, Rr1GThF93wHadWcx,
VJfPj3WbCzvLmQnA, 1Wc4yTf2MQPxwXDa, R2GxVLZnrkh1m6Hp, 4Tb63aXJykLnKWgN, aLwy1zjnrYFhvpDq, QXWcwaGZvCKjJFM4,
rf4VFwdKH2CjNPJh, 4vcn3WF1RxXf8VZt, 2fFYW3BmAjDRQxCg, 26j47DAXZCgwdxKB, fCB4wQt3NWm6k8M9, AKQPn8Ha7wG1XJVv,
LXPyfNtDC4mhVQ7R, zpw3WxckYDjXBJRy, 9R8rhcAwz6V17fpF, nAHx4Qv6ZaJ1VNz7, xWDkmwXKrRcdM7Jv, RJGVrNbKj1vXyF46,
tq89RC2Tv4jy1WcA, TzQa7YAycKWhrtCb, Xfbc1Pa24xzTBwnN, JcmNfbzKkq4prQg3, WYTNrpf3VQZt7Xqw, LYw76XG1CFHQNdJz,
J2tdqRB9HykWhGgn, hDJKcgjxvWQfHt9M, 1Tf47h2MvyczZbVK, qcWFJdjk3Mx6YK8B, tpKhjLGfwCzb6dA2, FHVZr9P48Wp1QjyL,
KtCPg2xhR416AvXN, 4D9aqWFLJVvn7bYx, RaTGbvyP3KxH8YwN, LvDw6BjHCVGgzJyb, HD8mBq62ZgkK1yRT, Wn8wHhaxy31MLzjp,
MgDw2bXrKTmWaLzY, Zz3YcQk69P7N8rAn, Q2PbD38mAWa7TzCk, Xn7WZbwdcp6jFP8B, PtKXrwfj8zBHm2Gy, f1WqtFkXGn4pBzT6,
8dxYbPNzVmFXvWjC, qbrGgBZfzLHFW12k, 8GFRaPBnKWf7Qrpb, dzgfcTtLA1v7jDZ9, aT2GjdRqgB9HtNL1, JnQNajVrBKq9tR17,
Ycq2twGVavTpMn13, 3aXTp97wjCWNxMhd, QwYrJ4VPjyBkRbft, vF3fwb7aQ9BgGmWt, Zvr84V1Y3Gpg6n9H, a2m7tvgYdbw3jVTq,
9RqNnvBcJp4fdX2h, 9tjaL4MdpWx6rmhT, M71jtPDgcNR4fq9T, tg9xyavcrR2GqhLz, CA9rN6hjpPHFdMvB, GcbndBCyf1N9m7YT,
vyQrDZN7RX82ApnP, a6t8yQpMdJRwDvhj, qJX72LBYW1fAx6DP, XM9mHgbhYAzWTjGy, VJ7chPDkjLx4AN8m, rKaNzcYX94fkTCmJ,
qTQYLaD9yjWG4Nnd, tJNYAnRT6pgv9xdz, wBZCxaXWn7TAyFGV, zFJ71HAQKYbv623V, rTDcQk4gGxfmjN1q, Pp4dVW6y1m7bfATF,
h67XZzf2CwTr4VNy, qbtZdThvBCXVcfAW, MWTNLc6jZdC7AXRH, ZDkaQjxTVm2dMgpW, Zb9vrBGFf2a8zgPd, 6hf8JKg9dzxcWL3q,
x9Jty6qTazhGHLKB, xcGfXYvTD6bpg3qt, MtWFJv1bPL26KTdx, g3zGbn4VD2fPycNv, Y74hG9vfj8MbcyCQ, DCWdZLkm2cPJbfz9,
RJXQC67jbzATWdx1, 3jARkacpG1LJgCTN, k1taHbwL76Cnxv4m, fvKDkG9td3jzXT2L, 8YGwk2vfrBpZDx4V, TtnNvmqR7GwpZaPM,
KfcWYGqM8hykpN12, HDmZ4zdFK6xgPCT8, ZRnjHLXNgyMFqpmW, T92NqcgBXaZrKvCm, NAnmwq6MHRdFQ1t9, tw8ZP2krfqXvAB4Y,
CBGfNpnjv2ALw9Kz, Xqkcn3CgZY2ryG1w, 82F6RVYwhapt4zLP, GM97YvP1hdgx8akn, f6Q2vmDkJ7YhAy9H, 3Ra7Dtjhc9AXLCzJ,
rajyY8pFtKfd9QDn, yfN361HMZk4vp2hm, QFqKVZgc6vAJfbrW, Z6Xm92K3zgfrqM84, RHgby76wQXp2nrxj, VKfYJ82cMTktxQzL,
cav7VZrmPbNW3y29, rfvtaC13gNqRx4Az, 7Lk3abCgNpVZjM41, 4zHJh1aptbByGFnj, 6CbKfDBm78nhyLjJ, n7hQW3NYzJwmDaG4,
qKtgYkmZMVjnLBfr, qkctgyRMaFxXVnmz, dK9TAcBfjVmD2n1J, gWpPrhyZ42qkJ8CR

Gift of Arthas:

ZXJDkzHTvMCpFbV1, hDGL1wmpV4WjxbHP, Mz7B1TyjCDahQY9x, 9tFBp3r6JvCkywqT, Gkc8Y1zZDNb7wWVg, yqFPH1MKv746dQhB,
xg7f81w2qvXGL6Fm, 7qPHR6cLn3zGptvN, tcBRkNWrqp29PzJj

Lacerations:

AnPHcV64wg8DThzZ, q62VRcw1m8fyHN4T, t38h6AJQWzRCHwZy, Xa2MfbLKQCPYH6nr, nPLWN2VDTZMz7Qk6, nP4N7AqTvChFQ2GY,
D1dxzYbNMKcpZAa2, 6Y2mMtCzb7ZvQKWD

Tenderized:

7dHJyYKgf6WB9TpD, pxMq28ANjCrWvcgw, jZhb8VYzvndpKPaJ, zFAVYtjCwbvRdyZn, 2aPGztJXAh974Fbm, 67yMCKbB91fdQtHw,
G6MdmNQW8YqaCZf4, 2qcmHx7QGDz1w8r9, BywJca3vKQPAM7L9, 86t7zaDPNxpRGrcT, Zy3P1DpzhALvR2Gq, KGnL8PNF7Hptzgm3,
B9dxNyPrnGTvDh2t, xmbNFzkq7cPAXJB3, vZ3x7bWPjfwHCgXh, tfX7HL9PmDdwkGja, JncG9jpRwQHhK7tb, 62bNYTtAkfDhZwzc,
HncmhA9J68wYv3gP, T6bnQy7fZrGVMcJv, zr6QacY1HTKLMWBJ, pMHYBm1VPADtKvg2, WNa6pZmGhrMtf19Q, VRvFB8wYK6Qg2myr,
qzvkgcVfWrm9npBy, h9jfxHpG2cynzCD8, GJbzHkj7yWhZ286x, Rbt46n8XhJaWFHkQ, rT3NBYyK1PFc62ax, fhRQALy7dFKwbznN,
YWdVfcBnjRQ9G3gz, 4fTZ768GxtrPNqzC, b3KV8BNhZv7cgjFz, zBxk3a4g1hmM68WD, 4jDvrdmYytRHZKqg, y2YcmB4qXnKCg18Q,
JZPaGbYCX26N4tgm, 8ZzhKBxdkVqjpXtr, J8YZxmQ6GB14yMr9, cTDB8RavNHbCK1Ln, fywxP7rCh6tGnBWa, g3kLNqK82twBy1nC,
A3vDFHYbw1KdNWpq, LNDkVQWKAgHhw6qB, 7nDYbQBv2RcP9Xpw, WcVwazXJKH9gCvpY, zQRvXaZ4ftdMbDq8, A7trQ8ZRvKDz3pgF,
f297FtyjGHmRVqYw, Py6HN97KpFYZQTD1, kvDKAcPqCgr8jn7F, HNAz6LWFJkY9nR8b, Zb6PGx7kaJHAyT1p, L2qKn81pm7jYkT9t,
jHXpwZ3hCRTNGt1F, hmc4Q9jtrgybfx68, Q827z1cwdtMnC9Dm, yjwLb3N497AVkRZf, H9VbtcRZJdaKM4QD, jxpN7FbH3Vz69QyW,
9L4hGr1bwmqAMnWf, bFVtHL9wGnxDKk87, KaVy7MjgvAYkBqFr, CKMFxqyLv9Gz731g, 6JKtQY2rRh3Nnjxg, mcATkq7RYvrP623t,
4MFWJ1gva3p7CjZ6, bhHyGVrP7MkJvtf6, rQMWqbxzVcp2Gn9Y, jtgZTL92cpQV7Kam, VAZk1B4wn2M9bDPg, KJ3xfRDvHLdtYcPG,
YkCwMAadhF48yzQD, 9Mw8TfWtvDNm4jF2, YVPX8gDCpb7B9QF4, 6JPk2WaZ9mcdvxYq, bq3XHZ2YJzGxhRLQ, ycDwGZPHhgQkYFd2,
X2BHnWa94wz7gcfP, MvAcd78yhXC4HYRb, ckxfJYL4DbGRz6hN, 8vC1fQH4cdYZqwhn, MQgJ13WZ64HahkDn, Wb1FC9Km7z3wdaTv,
KWZqJv6cCwhXBAkQ, g9mx7bNwpJPGWLk2, CyVnkj2p1tLAFrH8, pVNAk71xnYgDwqRF, 7ZDxRNanGAMz43cC, AmLpq9xyPrQYNagC,
Vk2BXyHrQZ3tjTJz, BPgkwpnQVfmvr78c, HpFdqkTJYLGNW1KR, Tp3CbBzgPWjKh7x4, xTQ1qz8DZ9KmgbdV, jkKx8tW69wg3VXh1,
8GL31HwqngaC4fzM, 3wK4c7n9avYdykVT, JF4x6wzCgGbpjvKA, jHxJyDKtz4vdgXZR, h4n18FVZftmMTKRG, kg2BK6YhJHwNPAQW,
Md7V9kmZxBNgr8KQ, 6L43dhP8DbTYrWK1, PMBZjgXqpFQVrb6t, pnmZ8qDtdar9b1KH, p6MKRN74X21z9CHW, fpj6NVGQ94BkqMdm,
aVXmGbxtLp3F8BPR, B9C8TVwHJZk6G4F7, B2D9zYahGJLKbVHx, q4ZYmVXk817FcGfa, yN3Rd1gPvFXkhjBG, BMPxQwDt4NhpnLZ1,
NmgTp6h9djVQvq2K, 1YKV2DhvtmbFfHLM, 76P9n4jkJtfpB1Gv, tMZQk8gya1NPhwzG, AxhZmXMvQTdfnDNF, c6YD2PNAJ3pWq1Fk,
KMynaR6bF2kzhxZW, ZH8xt94FYMLWJqbc, BTP4tm9cynVGvLXq, gY6ZhQxm27aHTk9C, dhfTFcVNnGHvamk2, B69h7FP4HTRbWmaV,
68QFLTtpACX2v9ym, HkfprWKGjZx4LR2F, Jx4GFqzVNn823ybt, T1YnWkgc7VLGdPJD, MkgT6r2q7ZvPdVnj, 174v6wGxA3bBCTzj,
xjvgTAhdzmJQNK1k, aMwchZFXtBdfz6Pg, 63qChvr8Py1WjZM9, zh71R8wP2HaJnjLx, db2awx9Wtym6PYBk, L1VXxcMARG9KD2dh,
j6d1mRwzMp4HVArW, Df7nRMWzaJAbFrZd, Vbmnq8CgAFK1M2t6, yLDtWzQq8NaJRbnM, 4fgDAM8PL2wpG6XY, 27Fr3MJTRdhNBHDt,
nf369aFPXxVmcYWy, Td4GCmfzFJckjZrN, nYwjhN643XFRykKW, DWjBLyVrC12hAmYq, HN9rfJtaAZLGmVT8, j1ZJ2xbYV9d7qh3A,
aTktVmBCM4g8vdYx, QhfFH38GWPJ6y2Yc, kzN31TQBth4L7cvx, gkPw6VGdWYFxqpK4, 8y4nDX7NHvQV6qKZ, CN4FVyDKvYkQHaX6,
1HcD8PyWkFbawzCJ, b74wA3g92NxtPvKH, mbM7QABp49cfwhaP, Z2RmbDLxXGHgNFJY, Mx8Az2XhNR9PjJQ4, xjMGtyV4pw2fDrqY,
c6By9C8PbVqt34vN, FrH3b8XJZQjLDfG9, wVA1zYmrNKCpvQn8, CFZ7bqQcWTXLA6fj, Anp2WTCjHa4YLBKd, 4LYCmjDGAbF9dHQ3,
hdL9NZfJzB1yjkvx, yDY2brPG3jQH9Kv4, FXQynRhmqdpGJrcH, nQd4V2KTCfMLrmRt, 2whN8nQCX9PH6VGz, zQydq3xRAZnkh9JN,
z2qnbhTND4wctxAK, xv19R2gVDqrpNXzf, FbZxpzjwGHDfLR37, AQ7X1kcHrNzDwKhj, kHVMLR2dtwrjxWD7, pMwG8ZDbJ9AFW7nT,
7gZq4n2M8YXGPfjh, B8AKvLZ6bQpaXktV, xDbJdz2KatXqfNpM, MLkQWdC6hNcm34gz, Q1TkNaDV8BXf39dK, zFrnw4yAmMxLkZ61,
yx9ta7bcdv8Yjw1H, NfK46gDLj8wrJz39, Bm4ZTvKwQWtaD7nd, FqVQN1ca6x7Z4tmk, 3z2aQRWk86GmtDLK, pzhdrVck7vGj8BCN,
GLgDZY82zW9Mha16, m9A2MvGDfPFRCWj6, vRPQ237fWm1tFYDV, LdcVztBw6pFjNahM, 2PWL9DB1k47p8MRd, Mv73gTFaNhbGWLy9,
rgXGC8LkTYVKMfAD, RbDfMjcTwdCBQKWJ, pJac1fgd8jQk9BbT, Y6bcWnJv9kFPyXmG, pzvJ72fcan6gkKNb, yCFgNJHXdZqDPKMz,
tJAFnvQby9TXc8jf, xPhaCzTjybrZ48gV, rmL73jwRtzQ6HdyP, 2xWwaBRLgQ9D46FM, qDKRc3AxpVkgCWwf, tdMgqKCmfLPJ8bX7,
3BH2d16wtD9F78QY, r6fZjG2Hx314KJTq, ZTJqNPfRLb3c8Bav, vYbjafXydK2z8chx, DMZPzKB91HJAqLQy, tH8kRGA3bph7V9vN,
cYazq49fdRZQv1bH, QpL8yfWmxY6A3Hwd, FHz1nBG8rhgMqaxW, AXqW386NctdMmQK4, MaXHVZ2nWDJ8z6Cd, Zv6fc3GgWpkX9Bqa,
TFnvh6k2P8zL4DCQ, 3BYFMvdKPgLc4xDn, 8VBhYx6tdm7GMvHK, WjnNbmcqYtkpK3TL, wtzQKj8GP4cJ2W6H, 8RWqvN4tjFHXz1DY,
FWHahtQvqjL3wNdV, ahRLFQNzAYfwT7jH, 1xXd2kGF3BQfAzJV, rWpDA8fghB1n4Xqx, KWFdXvCfR6rVYLcj, qRQfj9HCFYp2MrDT,
6J4ANdTDcpPK18M9, 2kFjvxn3MzRLQWTD, zKmW1p8PgRQtY6jq, GDWLqtcJgbrz1P89, xJRQt2TN3AFmPr6C, 8TKryAhdPfpW7t4V,
adj7kHPW3ypMKrNL, AVFatXcx7yBzN3hZ, ycK8T6mJWr43PpNd, fKGFtND74VBpCXbR, tZADbm8zqr9jVgJK

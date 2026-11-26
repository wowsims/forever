# EV-0115: Draconic Infused Emblem, Zealot Plate and Moonshadow Stave proc rates

- Claim: Draconic Infused Emblem 22268 procs 8% on each landed harmful spell hit, Zealot Plate 274651 procs 5% on
  each landed melee swing taken outside its 15 s cooldown, and Moonshadow Stave 22458 procs 1.5 times a minute on the
  cast time of each landed harmful spell, at least 1.5 s.
- Sources: Forever dungeon reports on Warcraft Logs, listed per item below with the fights each wearer had the item on
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938, by report date (counts per item below); hotfix state not
  recorded. No hotfix changed 1318930, 1318931, 1294666, 1294668, 1300435, 1300437 or 1300438 on these builds.
- Date: 2026-11-09 to 2026-11-14
- Pull request: #1016
- Later count: Moonshadow Stave's rate, §Count of the 2026-11-11 to 2026-11-26 reports

Wearers come from a search of 3,226 zone 3000 reports for each proc's spell ids. Combatant info gives the gear, once per
dungeon run; a wearer's fights run from the first combatant info that shows the item to the first that does not. The
proc column of every trigger reads 100, a sentinel the tooltips ("Chance on harmful spellcast", "When struck in combat,
has a chance", "Harmful spell casts have a chance") contradict.

## Draconic Infused Emblem

Reports: 1ZyhaVDdFtTYWMcC/1-12 (two mages), NRK61HZcDdJkrwXG/7-9 (warlock), NXMZV9xjvWpB7GtP/13 (warlock),
T79gRkJz4QVGbyYW/14-15 (priest). RXzv1k3V4PgaWhqj is a second upload of the T79gRkJz4QVGbyYW run and is left out.
Builds: 70888 (2 reports), 70938 (2).

A proc counts each apply or refresh of 1318930. The trigger's 0x10000 rolls once for each target a harmful spell lands
on: mages proc 17 to 20% a cast from Arcane Explosion and Flamestrike and 8% a hit, like the other classes.

| Wearer  | Procs | Counted against                   |  Rate |
| ------- | ----: | --------------------------------- | ----: |
| Mage    |   231 | 2,725 direct hits                 |  8.5% |
| Mage    |   135 | 1,682 direct hits                 |  8.0% |
| Warlock |    20 | 188 hits                          | 10.6% |
| Warlock |    12 | 148 hits and debuff applications  |  8.1% |
| Priest  |    37 | 491 harmful casts (single target) |  7.5% |
| Pooled  |   435 | about 5,200                       |  8.3% |

The pooled interval is ±0.75%. Per ability, hardcasts roll like instants: Frostbolt 9.1% (22 of 241), Arcane Blast 5.3%
(12 of 228), Pyroblast 4.7% (4 of 85), Arcane Explosion 7.9% (164 of 2,079), Arcane Missiles missiles 7.5% (37 of 493).
Procs a minute differ threefold by class (mages 2.1 to 2.4, the priest 1.1, a warlock 0.8), so the rate is no procs per
minute. Mages show two procs in one millisecond on area hits, so there is no cooldown. Hellfire Effect never procs it
(0 of 1,991 hits); Bane of Havoc copies proc it 4 times in 558. Misses are undetermined.

## Zealot Plate

Reports, 40 paladins: 2HMyR6kfbWKzBhtv/1, 2bdPF6mYxvJgZXK7/1-5, 369Yv2NcqrygTpwX/1-5, 3LwjMABn6RCd1Kz8/1-2,
4YzXaGtg81WdqM2N/1-15 (two wearers), 7vacy8gRZM9CHm16/1-11, 8vJjYft73m9Chzkc/14, CF7ztDwY8pJ9KZhn/1-12,
CGF4WpJPjbhvKg9y/1-2, JFqRThp6Z4HrnBwf/1, MC2GmHchpfVkNqwD/14-47, NXMZV9xjvWpB7GtP/10-99, NgRFKH1PWzykqC9Q/12-15,
PB6zGcVNtTZrMKXb/1-6, RPFnkAX7pH6qQ2Mm/7, VCaNv2TFXx6y1dYZ/1-9, VDCA7xJ8GyNfrtcW/1, VyXwH1Wfq2bc3xZ6/1-21,
WrBf9NpqGhTcm21x/1, XxqdrnVN9DCZcL8T/1-9, bP6qxJTBpdHf47MA/1-4, fAThJd7KGmX3Pkcg/1-2, fXyK8kCNdGxtQ67c/1-2,
g6JdFkQKfR7YyP9B/1-3, gBDMLv8J1K7nqkpd/10-70 (two wearers), gJtby1TP3nmA2DFp/1-29, hR6pzT1HkaLym2nd/1-2,
rBWb9Nyt4YpPFGH7/1-23, rDJL7nqX1VZHm9zk/1-8, rtQwP64YL7GqR1Hf/1-12, tY2jp7cVGaQx8nDw/1-9, wJtPdpxFGQCM9f3c/1-2,
wa9fNCxj8JZLMDcq/5-7, x378BTAPqhCJryVW/2-12, xWKHB8qbYAcgJVRn/1-6, xbfy4DQZ6PWCFXYr/1-10, ymcTt8vprBnJFC9H/1-23,
zptFCDJMLT7QGxvK/1-3. Builds: 70862 (5 reports), 70888 (8), 70938 (25).

A proc counts each mana gain of 1294666. The denominator is landed auto swings taken (hit, crit, block, glancing,
crushing) at least 15 s after the wearer's last proc; mob melee specials taken are left out of it, while their procs
stay in, so the rate is a little high.

| Procs | Landed melee swings taken outside the cooldown | All landed melee swings taken | Rate        |
| ----: | ---------------------------------------------: | ----------------------------: | ----------- |
| 1,661 |                                         31,232 |                        51,802 | 5.3% ± 0.2% |

The shortest gap between two procs of one wearer is 14.97 s, so the row's 15 s cooldown holds.

## Moonshadow Stave

- Count: §Count of the 2026-11-11 to 2026-11-26 reports

Reports, 10 druids: 24kgnZAXbKTW89x6/1-2, 7cpTkjaFB3mz6VgM/272-492, VqZ38vXydB7gn1J9/1-5, WKgT9mPJHnYa14Zy/1,
cvRTFb1qhGBKNkCD/1-24, f36RaB12wyDcTCGY/1-3, hCfYgdqVk7QDRZGN/1-3, rPpHYfq3tNnwJRMd/5-19, x3njCmpkVYf7WDgh/1-39,
z1kHPnyGtX7g4QqC/3. Builds: 70888 (4 reports), 70938 (6).

A proc counts each hit of 1300437 (Shadow) or 1300438 (Arcane), 359 in all: 165 Shadow and 194 Arcane, an even split
within chance (z = −1.5). Both are projectiles, so a proc lands one
flight after the spell that rolled it; each is matched to its spell by target and Wrath's flight time to that target.
The logs carry no positions. Two wearer windows end in long stretches with no proc (959 and 487 casts) and are cut at
their last proc.

The chance is procs per minute times the cast time of the launching cast, at least 1.5 s, divided by 60, rolled when a
harmful spell lands on a target; the rate is in §Count of the 2026-11-11 to 2026-11-26 reports.

| Model                                      | Log-likelihood |
| ------------------------------------------ | -------------: |
| Procs per minute on cast time, per landing |        −3881.3 |
| Flat 4.05% per landing                     |        −3887.7 |
| Procs per minute per cast, misses rolling  |        −3886.6 |
| Procs per minute, Hurricane ticks rolling  |          −4226 |
| Hurricane cast rolling once                |        −3932.6 |

Nothing else rolls it: Hurricane ticks 0 of 16,615, damage over time ticks 0.7 of 10,625, misses and resists 0 of 159,
helpful casts 0 of 3,103, Thorns and melee 0 of 1,122. A Hurricane cast procs it on 85 of 591 casts and never twice, so
it rolls once for each target until one procs. 18 procs stay unmatched.

## Count of the 2026-11-11 to 2026-11-26 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every wearer stream of Moonshadow Stave, listed below
  with the fights each wearer had it on; 387 reports, 390 wearers, all druids
- Game build: 1.60.1.70888 to 1.60.1.71139 by report date (70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19
  UTC, 71037 from 2026-11-20 18:23 UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU,
  71139 by 2026-11-25 22:38 UTC; its rollout time is not recorded). Reports per build: 70888 4, 70938 51, 70959 87,
  71037 160, 71091 68, 71091 or 71139 17; 235 start after hotfix push 114313 (2026-11-20 23:47 UTC). The hotfix each
  player ran is not recorded
- Date: 2026-11-11 to 2026-11-26
- Pull request: #1253

The count runs EV-0219's method (§Method there): roll groups of the wearer's landings within the row's 10 ms cooldown,
duplicate runs dropped (11), and each wearer's last stretch with no proc where 8 or more were due cut (13,168 groups).
Both bolts fly at 20 yards a second, so a group procs when a bolt lands on one of its targets at most a travel window
after it, and the window decides how many procs from other rolls (Hurricane casts) fall in by chance. Kindling Stave's
druid wearers, whose proc lands at once, put that share at 0.3 % of procs for a 1.6 s window and 1.5 % for 2 s.

179,149 groups, 172,687 of them single target; Hurricane's cast and ticks are no roll here.

| Window | Procs | Procs per minute on cast time | Flat chance, log-likelihood against it |
| -----: | ----: | ----------------------------- | -------------------------------------- |
|  1.6 s | 8,134 | 1.515 (95 %: 1.485–1.545)     | 4.32 %, 217.4 lower                    |
|  2.0 s | 8,398 | 1.565 (1.535–1.595)           | 4.47 %, 212.9 lower                    |
|  2.5 s | 8,636 | 1.610 (1.580–1.640)           | 4.60 %, 188.8 lower                    |

With the chance share taken out the rate is 1.51 to 1.54 a minute. Single-target groups at a 1.6 s window: 1.5 s or
instant 4.03 % (5,401 of 133,969), 2.2 to 2.8 s 6.26 % (2,413 of 38,543).

Reports: cvRTFb1qhGBKNkCD/1-24, p4R6gacHrbwtZPh1/1-4, hCfYgdqVk7QDRZGN/1-3, WKgT9mPJHnYa14Zy/1, f36RaB12wyDcTCGY/1-3,
rPpHYfq3tNnwJRMd/5-19, NXMZV9xjvWpB7GtP/103-123, 7cpTkjaFB3mz6VgM/272-492, 24kgnZAXbKTW89x6/1-2,
VqZ38vXydB7gn1J9/1-5, 7rGZbTJtpAqnBh2Y/1-11, z6LMfrpbwhyVCAJK/12-21, 8YR7TpW4qXbnMKfh/1-3, 2Z4G3r9pfqJtCM6W/7-17,
WrBH1KZxJvVP8hQk/1-2, Lpr76Vd3DbKWZqnh/3, bcp2G1PHy3KWNM9D/1-11, Wv4QYA6C7antxKXr/1, 7fy93LdcwQhvF2gP/1-9,
HKY8mdFptfNkDgwb/1-3, qcdyZAr2ganDtxNW/1-44, qVkdG2t67FhNgw3A/1-14, GRnMHmK69BXfWVTj/2-4, YwhQtdMbNARyqL6B/1-5,
4f2ygdYCzFAvqc6K/16-18, Hxy4dFY3qfXWZ9Bp/1-11, tW38cgRn6b2kr4dJ/1-11, fz2HQxXk96ywVYv7/1-8, aHKFrZvJgBypW24m/1,
cMx7Ny2q4RrFXJzg/1-7, 7cpmjq6DTBNLwFhG/2-7, HkfzanMQcVptAB19/1-3, kJ3WdhMLqv7K2Nyf/1-3, PYQf86cTrZ1GjdKx/1-3,
1z6PXMGdQBDvrCqx/1-2, rfg8k3Nw1P7TAHx6/1-3, P6GLNM3n27dtcYJD/1-38, Qr9PmCnptRWMjN7k/27-31, KCYXdTfNgGHczbLw/1-6,
zHWaxXt4jyCNnk6m/1-6, tkPZYqTRgCyWG3jM/22-23, VtkPTAqxCv1X8cJ3/1-5, j7xyJdLXHZNA6PQg/1-4, x9v27CNzhRgMkXG6/3-8,
ykz7Bdp9avqbAjRV/14-21, pqVHWtXCYvdc8y1B/1-2, ZcFKV7tLHbDR4qJB/1-13, JBxW16RqGXMPYH3d/10-13, C2xQXzmFg6P1Vaq7/1-4,
rDYW2wpQ4bBncAy1/1-12, XQNa9CdV317rygRt/1-5, zVAHZcQtTdbhMfLN/1-8, 4pynYMkBdVgRDvKH/1-10, ChNdrnFtXcqfvwBJ/1,
rfz4pR6mx1w9JWKA/6-7, H2Zg1jRm93PpTqK4/1-3, PLc9fmwZdkzRgYnT/1, HNtn34KXCMyDrQJj/1-2, LJzACmpf3F9cTrXN/1-2,
hHwbja8rcN6Lx4Qy/1-3, GFBCh3VZH7TmAkL4/2-3, bkm7pBz4N9JtC23L/1-4, P8YxXZwmqfzyCMtd/14-15, dK7tvz3P8qGaCN1H/1-13,
NjR4CYAZ1bcXtK38/1-29, fCB4wQt3NWm6k8M9/5-7, MqJFdTLDaVH9yWj7/34-37, G7X1CkT8z6PrfjWM/4-871, HRYM8C3jxmPGdbf2/20-64,
QNfaZ9MTnWpFrqb4/4-199, pMgWD6Yhm1Zvr8nw/1-10, vnKj7bWLgwJqm3dP/10, ZaPqVJn1bQ894pxT/5-52, dzgfcTtLA1v7jDZ9/15,
2kNtTqKPAf81CVx6/1, 9z4Xc8LHDWjtA2np/1-28, 8xMAnXgVpP9KbW4L/1, CA9rN6hjpPHFdMvB/4-5, pkQtTVjwCzv7H3AD/1-15,
tMxXdgvhQL8ZNP2b/1, 2tCNaTJf49KWyVLk/1-16, qa8wcv4jLYZpGARK/3-4, ZRhnvMANGJDy1kBj/1, 2CRfAV31QPXNw78G/1-4,
bVACzhMqpGfjPyY2/26-41, tj8NzJTyBFcH7kmq/1-2, vpq9gA6dfFB1CJjr/1-3, Tx7MPH9f8Ga6hdrF/4-20, 73xcRhZpPmNvC6fw/1-4,
rWcKvP3Y64tVDBxw/1-4, MPzHZF4RcJfrTahx/1-40, yfN361HMZk4vp2hm/8-11, GM97YvP1hdgx8akn/1-2, KdQtvTWfGpD97NMc/1,
J2ThDq8RCPKNZdrB/1-11, pt9DBx2nCKPMdvLz/4-6, vmyLhM9Pp6gQRrWK/4-6, rQ2bTtw8fHyRNGj4/1, DMdzh7YqFjg3BrV1/1-17,
rfvtaC13gNqRx4Az/1-3, mJxTr8d2FkKzNZB6/1-3, 6CbKfDBm78nhyLjJ/1, BdVWNAypw8kT1Rfn/1-2, t3g6TX91AGmz8yhr/1-3,
GXmjgzxb2hwqtW9J/1-2, D2nQcj6wFbmBgVyJ/1-40, mGc7gWtYbk6n2N4d/1-2, pnGKFx1vL82VWPy3/1, G9WjCtnrDK3hdZRf/1-8,
GXrPqfDM2ptbandg/1-10, Fc2YBt34nCHRr7Gd/1-32, 1Y2XNrVgGfKvpWMj/1, P3ZHAvVJTq8NBmFc/1, mW2TtC9YqHZz3vpn/1-2,
GcrpRyMfx7bqd3g2/1, 2TytDnRqwPXHa3gJ/1-3, h3ckbw9ZCRKnXL6g/1-3, bgVLFd2wKGTW9AP8/1, zPKMcn8JN9GkT7vh/1-7,
VRDxkwdmjzZrbCtQ/1-4, Ypx8qtdwyNfHam9z/1, 8FAmTNqRVLb7zgH3/1, vYjL3q712CyTGkQh/1-2, 71vkdgjap6CMJ2xL/1-3,
QrFmNx7k6DnwAHbj/1, 2JryaLXZKA9qGFcP/1, 6fXVrbgKR9P7DmaB/1-2, VW6NT73MXFw8r4CQ/8, LNbjwpJhRTZzGP3C/8,
Ajwk2cLmXKpt6Gfr/1, hBTn9D2CbwGNPfLq/1, r29YVtq7mkfNZBcF/1-2, ztyQjYFA8BT914fC/7-12, ZmCAR8bDny9kpKJ6/1,
63QrzvTALnZKgx7Y/1, DW4wrpGfc6Xzkt71/1-5, jJca4GYpvxmbWR13/1-7, hFmCtf4KAvdrP21X/1-8, GxgMyQYRHX7f1cFT/1-12,
W74wN6fJq2kybTjK/1-4, x1QpJTH2aDZLnq9h/1-6, 6CGNvtpXkVn3bxFT/1, 81MthAknL32HWKpq/6, 9QKHCPLay1fMvtpc/1-8,
39jqBFgQcwzJ6Z4K/1, aVA6R2pvy37rYZdt/1-2, mAwJybNRMq24VFBg/4, XLcKwVBC2DjfgrkP/2, QwCaA91PD3npqrFB/1-3,
ynmaFdBVKq8WXQAT/4-25, 68HNbmvFWk971wr2/1-8, PWzkYdMm9qxw3gaC/3-6, FwGCgx3YAcNnrjBL/3, hMKNVdYbX483ZvHw/1-19,
Bhm1dHcrJtR38MDL/1-2, By24Afp8wQZn3aTJ/1, K46NXazBqL2pMVcw/1-2, pBFcvGPyjmAXg41Q/4, FYKRbLxDcjVBa3WN/1-2,
TWH3vYBFCaAQND6p/2-3, Wc61XyHwnvRqZgJk/1-6, Vxj68RbmkJh9TcZX/1-8, FcxyC1BM7NKwf3DA/32-66, CH7XKxqGjwTbp4LJ/1,
14kfqVMz8TCpwmb7/2-3, Jc6AGf2bhLyqWRFj/1-4, n2HkhcYFrLK1tfZj/1-2, Z26yqvgBtjbHMPLQ/3, KA9GJXfvPqdZ7Btg/1-3,
AB9mrk6C8vathbVH/44-47, BqrLfbNgtGDTcAhk/1, aP6J2RH1jNfLZmA7/1-2, gWVp64hkqHdCP72b/1, VGzkZ8tdYFRwfX2n/1-6,
XtGMhqwAx9FanKLH/1-3, kcP9CaWv4p7KANFg/1-3, 89nh47QpfkMcRFa3/1, bFtqNmQg7wMLDA4W/1, DjLHGMgcNWY3ztCk/1-3,
HnC4whkxKqjL9mDy/1-2, q1X6Y3NrvTHRVa7Q/1-5, Y96q1TVjdmtDrL3a/1, APkrd31L79hQDjKb/9, DKhtLa2GqWFxfgnT/5-61,
vMgWz38rpDHmV9N1/1, JFZ1yXwVfNpPjvmg/1-2, pagXmwN7J96DF8hf/1-6, JP2vpcFz6bT8139k/1, NDXJYCGbQpqP9gVT/1-2,
tAqaDFBJ4HcLjdmk/1, qz21yTxvc4D7ZfAp/1, vCj4BFZmhx8dX1kA/1-2, tJP39ak7znBAYbwy/1, G18qvyXC9hgWKzTk/1-2,
rBbd4Q6qf3n19pYz/1-9, VjFXGq16923cYgAw/1-2, hxfCGZaDKvFyp72L/1, djc34BfAbG1rPgwv/1-9, aMJztGwK4rjvnPTN/8-43,
h63kRLJvgWztFfwX/1, T9Vjh2FWfNAxXb1J/1, QvW6VCHNcMrDFpXG/1-5, yVHK8nBdjZ6DxNMc/1, rW9hYmCpqXV4dA8B/1-2,
BDvPhzna6WN3Xgwr/1, PkVG6NDnLRFqhMzp/1-4, 3phFAzQPtmKgGYM7/1-2, NA7RJtkWGCFjXPYM/1, NVWhRCpkt9dP3JjH/1-9,
b6N4mkqYvLMdQBgD/1, ZVbRyxhG67kTYaPQ/1, vDmfBHPRT72y48CF/1, L78RzW6DBnqMJaC2/1-5, mbMkayH1FDqTQ3dr/1,
TM3ZzXkJ8qjCfYG4/1-6, yPLkwGYbDW37QXhK/6, hR7dDX6bPtwrWjf8/2-6, kYhH4d13pTFMLGBX/2-7, z2wZ3q9hpMnbxfWK/8-37,
Gd1bXFkygt2zwWaT/7-37, rvYCKmDXywQhfd1c/2-3, wHCkq6KnMDAL98mY/1-4, fpCPTBkQ4hJajzdM/11-49, ADxF6TNdhRb2amJZ/1,
Ztd2myTWRM8BYNQ1/1-7, bGgTqRaf2KY7hxLc/1-2, a7BgwJZv8VhRkYL1/1-3, wzayHVhr31LFQX7G/1, 9FAvgmy4fdCq2kMG/1-2,
dF3zbWfawRkZ2j6K/6-31, xrz2mpQYjDFL1yw8/1-6, wyA68kvKN7JqtgnX/1-2, jgvLQ34zd6mRkWaN/1-2, pcYnzMbDGkBTKgAW/2,
GtkKwPqHpBvXW7Rn/1-2, jm8GyZvRWAh3MQrK/1-3, 6prM1FBVKTgARmhL/1-4, Tk7cRZy6VQnD9hgd/1, 4gKrbypjJNta2Qzk/1,
fvAcGJ68rRPtmaZV/1-2, HM6cX2C4Kzh1frTk/1-4, vJAZWG1Cpahrmw9x/1-3, MyDhB6YA1VN4gTKz/1-4, 6yTMhGQPFN7mtLcV/1-3,
ZhyVJcb2v1zaLx3j/50, JbGTYWDxQyvFk9N6/4, vM6kRdNBmZYA3XrV/3, wTpN2bMCjdWyA4Jg/5-34, A2GQnyktwCM6xd8T/4,
X7APq6bphFMGT1Df/1-2, T3a8ZkW2dJ6jrMvN/1-6, gc3TvPVDFGYdkRNJ/15-52, PxKcfCHMgDZpbmqt/4-6, tZY2KRFTGfyh9WMV/1-2,
B1WXm7QGV4hdHLbK/4-31, 4v9LbMyPzRrtVdkW/1, haAMFPjKd3xvNpJr/1-3, YmTFGMK1dg7XBz32/1, Vvhkgc1rpfDFjLZX/1-9,
zy4ZAcxwaNPpkvgV/1-8, RWjNGhpVvJ36c74n/1-3, Tvyd137VtK8BrMQh/1-3, pwLcrnZ4FAYKRPq7/1-5, f6D78m2LpWv3bzgy/1-2,
AJZQMFgTGxnmkV1f/1, X7BT9Nph1tkdW4y6/35-41, Y7LzFwD98HmQNkBP/2, c48NByWfP2dMqmw3/3-4, 9WNwMp4QmchzHCk2/1,
hw42mCHpj7RAYXFa/1, tnxkwz6W8XKNfJB9/1-2, MfkWdwPAKVapznJ8/1, rdqgPfC7bwTyaWkp/8-78, z4Vm2hMCap9GNj8Q/8-78,
vFRJLdmpCQg3W49w/1-2, K1LZ6RcbPfztrdaY/1-4, PBV16XqGDnyvFYhp/1, hDjwH4fJPmNLpg72/2-3, RT1Nz9qP7b3LJdW6/2,
HavT3JbWpZQPgfLX/8, br4AMRdq1Wv7jpcB/3, vpmnQy1gjYD89JtB/1-12, YvZ3CGFnzPk4rJhV/1, gv9V6Ty1Qa3MBZW7/1-36,
TYA4L2zgKCtpq7cH/1, 2MvKXxZr8DjAmbqB/1-3, Pk2mFQLyAvxRXVcb/2-3, YTyBWAPjJVN9gmMF/1-13, pmM1KxRgVAd7XN2z/1-5,
KCvH3NFzqB6GdcWA/2, FrpBcP9M7mNYb2tf/1-3, PTzGK7Nma3DZ2W1q/1-2, 61ptfZJvWymL8QMT/1-24, 91NQzDJrYFGXy2Rm/4-14,
Tnpc1hvJZ84RybWC/1-2, qA8cmDH7P26J4QgY/1-2, mhMHaCzj1WvBtVw8/1-4, w6tWcJKGLqXrYgkx/1-2, HwgJBXfYM3L7VQtF/1-2,
n2FYaQD1gcMWfhNJ/1-5, 2qXQFf7pdaVvbJKN/1, Bt2mX4Gbk1rTvcDh/1-2, pWG3KqtARD7Jcdj8/4, AdncDbNmRQqKfFyP/1-15,
VC1fp3AbtknZ2w8K/1, kXtwmgzJ8hT4rK7d/5, mYT1dtFbhAMZBVW2/38-79, crbyqkzwnQCJRhDB/1-9, ZWdf9LAB3hc1RwQ4/1-2,
c9NXwA7mMbCL1ZJD/1-14, 4JFQAPcmTD1kfBR8/1-7, P1xTA3jMXHybdLNp/1-8, waN34KDtpkBgcX1h/1-8, nXN3vw2hyj6TLRPt/2-6,
8PWpYzaDMxwHN2QB/1-49, TCWgZG1HRthpncPY/2-5, h9txYKraHwfn26MB/2-6, N2vLpDafkjbmZzgq/2-5, fvbt6kJpyazFXK1n/1,
dLWzb32mAyh1TV6H/1-2, TRDHw9q17a3dPBmg/2, TyGawFf7MhLnB3Dv/2-7, 4ncxK1HmgCj63z8T/1-2, znBtTYGHyCgfWcQk/1-8,
wtGDFX91YWm8Q6KA/1-2, 7XRjxY6wkPbpBzrv/1, CkAWbp9gYdJwMQTn/1, kL1g29NH7MhXvBRQ/2-8, gZM2vjAfpdt3YwKk/1-4,
2ZK3wq1d7MWmvbhC/2-5, F8PLBprjC1xdztQG/2-3, cyWDrPwkJg8Y1bHv/2-7, pwkhWHQqRm8CGX6T/1-2, Kdw6P2fQYBnNcFaq/1-2,
v1qh7nCdNw8QH4Gp/1-6, vWd9mxRDpAcHZtwj/16, 9K8tJGrwBM3ZyVAC/2-3, xYvZt3LdKCwz8H7T/1-3, DV89ATQcFdKCMn1b/2-7,
yjQ4JrR9bXZ1aWPk/1, D8yCPRqt7XYmQjaA/2, HxJqMcGwYX29NArF/1-2, nzYCARkVHK7wM4P1/2-3, rVhy4TwgCX9jMQ7k/1-10,
xfQF29BkGTtjdRKh/1, 7ZwQdyLpNbRv6FDh/1, zFJDqtGyHZA9r7kj/2, TC7kgcFh1bvwPKYW/1-4, CcrfYL4k3pqnmy91/1-2,
k9VyMhnXLHBWdv4p/1-7, RjVpx4YM28XhWBm9/1, rNbmdgnBLkGqfVhW/2-4, z9nW86RpNLYZH2PC/1-2, LwNngyc6DQJKPt3G/2-3,
mM3Rt6wFcrBWxXkZ/1-16, tQ1ymLaFbAT8c6KV/1, 62rZKVMxLzk7aRyq/1, nmAGgfbYwTxcZ6B8/1, QkpdWYBFLyxbT8GM/1,
qGh6RTyPrtdZcvDX/8-23, v6bGnfAjX8DB9War/2, vpRD61MkQH2gPGJK/1-4, MTymYDhkzBbL1pnx/1, DypVQXYBwH6nfJah/1,
4hy92m31Jk7n6M8t/1, CbmhaRgpvDHNX4fA/1-11, 4rxZTmvWzyVaFYnw/2-3, wdp12RVFzXbD7yBA/1-6, cbxZvP84732hY6BL/1-13,
7NFWphAPYz3g9CJv/1, CT8QKw4ALXHRZmn3/1, VjXcwFDWzP9nQCJL/4, KH6hZQJz7YR4tNjm/11-12, kP7FWR6LpgVdJ1rB/1,
yHDbQJwLVzpqW1fn/1-5, Whwg8QmTVK1JNXac/1, 8naXPAcmwJMZzWYN/1-4, 3DjRbB6Z8w7kmghq/1, zbJQ28WLdkFt9D1G/2-7,
akbvGwcL21ZAhCjf/2, r2HNVk7QDGvXqKTn/1, ZcC6xT3QhDgrHdKL/1-2, MDmqgxJC689cwQdf/1, 9CgtkpAYqGjxP34Q/3,
H8hLvnkbazAcB16Y/1-2, yP8V12WGtDpjdcA4/1-5, BYaWg2t341pDvFTj/1-2

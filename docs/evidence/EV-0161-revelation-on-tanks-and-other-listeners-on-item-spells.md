# EV-0161: Revelation on paladin and warrior tanks, and the other listeners that hear an item's spell

- Claim: On a paladin tank, Shieldsworn's strike (Holy Shield 456544, spell level 30) rolls Enchant Weapon - Revelation
  (8217) at the level ramp's share and Blessing of Sacrifice's moved damage rolls it; shield spikes, Holy Shield's
  damage on a block, Retribution Aura and Consecration roll nothing, and no warrior gains a Revelation buff. An item's
  magic spell also rolls Insight (enchant 8216) and the paladin's Vengeance, as it rolls Revelation (EV-0149).
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000): the paladins and warriors wearing 8217 in either hand
  or gaining a Revelation buff in 110 reports, the paladins with Crest of Retribution damage and Vengeance in 4 reports,
  and the wearers of 8216 in 6 reports; one report is in the first two sets, so the Reports section lists 119. All
  fights.
- Game build: 1.60.1.70888, 1.60.1.70938 and 1.60.1.70938 or 70959 by event time against the time the CDN first
  listed each build (70959 on 2026-11-17 15:19 UTC), as in EV-0149; the Revelation section has the procs on each. The
  hotfix each player ran is not recorded.
- Date: 2026-11-11 to 2026-11-20
- Pull request: #1135

The proc, the roll and the level ramp are EV-0149's: a proc is an application or a refresh of the class's Revelation
buff on the wearer, and a roll a landed hit of the wearer that does not crit.

## Revelation on paladin tanks

Every paladin and warrior who wears 8217 in either hand or gains a Revelation buff, in the Warcraft Logs reports up to
2026-11-20 09:36 UTC. A paladin counts as a tank in a report where they cast Holy Shield or carry Righteous Fury:
2,922 procs in 110 reports, events from 2026-11-11 12:10 to 2026-11-20 09:36 UTC, US, EU and CN regions. Counted apart
here means a landed non-crit hit of the wearer with no hit of a spell known to roll (the cast spells, Seal of Light's
heal, Crest of Retribution, Krol'dok Resolve, the other item spells with the magic defense type and Shieldsworn's
strike) from 400 ms before it to 400 ms after, and a proc within 60 ms after it.

| Hit                                              | Hits counted apart | Procs |
| ------------------------------------------------ | -----------------: | ----: |
| Blessing of Sacrifice 20729, the damage it moves |                195 |    19 |
| Shield Spike 1308777                             |             13,204 |     0 |
| Mithril Shield Spike 9782                        |             14,235 |     0 |
| Thorium Shield Spike 16624                       |             14,021 |     9 |
| Holy Shield on a block 1239570                   |             39,537 |     6 |
| Retribution Aura 10301                           |             88,145 |     8 |
| Consecration 1280349                             |            144,973 |     5 |
| Holy Strike 10333                                |             17,299 |     0 |
| Seal of Fury 20418                               |             69,857 |     0 |
| White swings                                     |            105,835 |     1 |

Blessing of Sacrifice's moved damage logs as the paladin's damage on the blessed player; its rows have the magic
defense type and spell level 46 and 54. The spikes' rows deal their damage from the aura's own trigger and have no
defense type; the Thorium spike's 9 sit at the damage shields' rate.

Shieldsworn's strike 456544 is a spell of its own that a Holy Strike casts while Holy Shield is up. Counted apart from
the other rolling spells, by the Holy Strike rank beside it, a stand-in for the wearer's level:

| Holy Strike rank | Strikes | Procs | Chance |
| ---------------- | ------: | ----: | -----: |
| 2495             |     156 |    10 |  6.41% |
| 5569             |     495 |    35 |  7.07% |
| 10332            |     549 |    10 |  1.82% |
| 10333            |  13,838 |   114 |  0.82% |

At the top rank that is the tenth of the chance the level ramp gives a spell of level 30 at 60.

Procs by source, each put to the first of these with a hit of the wearer from 400 ms before it to 50 ms after: cast
spells 605, Seal of Light 55, Crest of Retribution 1,440, Krol'dok Resolve 444, other item spells 147, Shieldsworn's
strike 184, none of them 47. Procs by build window: 70888 6, 70938 1,426, 70938 or 70959 1,490.

Warriors: 7 wear 8217, one of them a tank, and none gains a Revelation buff of any class.

## Other listeners on an item's spell

Two more listeners that can proc from procs and hear spell hits roll on an item's magic spell.

- Vengeance (talent 20049, buff 20050, logged as Conviction) on the 4 paladins in 222 reports with Crest of Retribution
  damage who have the talent. Only an application and a new stack are logged, so the count keeps crits with the buff
  at 1 to 4 stacks and no other crit of the wearer within 300 ms, and a stack within 60 ms after. White swings 77 of
  77, Flame Wrath 16559 14 of 14, Shadow Bolt 17483 8 of 8, Flame Lash 27655 4 of 4, Smokey's Lighter 17283 3 of 3,
  Consecration 0 of 319. No Crest of Retribution crit falls with the buff at 1 to 4 stacks.
- Insight (enchant 8216, trigger 1248758, buff 1299796) on its wearers in 847 reports: 16 gains have only an item
  spell's hit of the wearer in the 400 ms before them: Flame Lash 27655 6 on hunters, Fireball 18796 7 on shamans and
  Firebolt 18083 3 on priests. A per-hit chance read the same way gave 5.4% on the wearers' own damage spells and 9.5%
  on their heals against the row's 35%, so only the windows are used.
- Flametongue and Frostbrand Attack: 37 Insight gains have only their hits in the 400 ms before them, and lone hits
  read 8 of 1,093. Insight's 45 s lockout thins the lone-hit reading. The sim lets an imbue's hit roll Insight from its
  row's Can Proc From Procs, as Lightning Shield's orb rolls Revelation (EV-0149) and Flametongue Totem's crits give
  Vengeance (EV-0226; R-0051); Insight itself is unmeasured.

## Reports

1DVFRby49Pg2tpzG, 1MA3qWLGa8v2hdXP, 1QRqZbYKXAmfD4gc, 1gzMYWQmFdtTJXcb, 1z6PXMGdQBDvrCqx, 26NKMABVWZ1CY3HP,
2Cg8R4ZMz7HV3rfN, 2PfTacjwkD9xJtzd, 2ghMyWmCFVtzTGv9, 3pBbkwHZcVYJdavj, 4Tb63aXJykLnKWgN, 4azX71FvCjQgmJkc,
4hHBvakRnT7p1GVK, 6CqXJpKNPtB2YhL7, 6QMFGaq7Ptjmyn1C, 7KZzkbdXjBHGWQJa, 7mFKVrW6QLHy3fgk, 8HxyMGJQwXNFPBAZ,
9R8rhcAwz6V17fpF, 9h3cXRZ7LfM6gPwb, AKQPn8Ha7wG1XJVv, ARwPXzLJk78HgVQq, C8hLYZf7Tdgz46DN, CA9rN6hjpPHFdMvB,
CqPk8KgFWYTMnw3D, D9frqPV8dam7B1AM, DZCc2gtwhW3FBKyx, FaBKy41kwpxCRGfr, GDzTwLKkX42ZpVvn, GP4zxdtkFDcA7LhW,
GWMx4hcJFCYHR6dk, HwyJMCLKfbY1PkmG, JctLn6NBFx2rwGjv, JfwkMdTjLaDtVpXz, JpMACkXNqrQjbBaF, KNtHCAPd97ZMTLBf,
KfjzWvhFNCDMLPwx, KjMFGvmkR7LDYz8y, Ky6k1d7mP2zVXWpt, Mg2ay1xDjkYpBqRX, N9PGBMt7faY6W2J8, NTJzVdMZBQDaqG83,
NXMZV9xjvWpB7GtP, NbMzcKyJPVYBDpnh, NtkdG1bPK9TBCDfH, Nxznkwgt78HVyvPC, P28Rgzc6kWVXNhwj, P3QtFzTLVp1aWBbA,
QczFgrTbG4CPpB3n, RJXQC67jbzATWdx1, Rk1qQxgLcZPmpBaf, RwPHNk3B6fMGvQcV, TQtaGPBgpC1KZy7c, TqDxtfJKXPCLk3zG,
V4b63zWRYL8yZmtc, VYBjgJx8aZtRhmpy, VgvHznLqA3PhXQjN, VqAygFH69PTC8Y7w, X4wDxLvF3YgfyKW7, XAMgrj3C9JLh1Kzc,
XG8yAaRFZnYbJCth, Xa6zFyNWmYdp3BZL, Xagb2QHKCf6rDwA9, Xhz8p6yB7bYqFaJ2, Z4VAmgDfJ9GPFjtc, Z6Xm92K3zgfrqM84,
ZAcmWFYBpjPJ7VKL, ZTGq13h7DH2bYrNw, ZkgVjD9fcKFP8bnL, aYtRFrf3kGAwxq1c, bNZvk628gXLz7t3c, bQw4YVFNJZBt2THG,
bahZczXqfLk3pQR2, bx9rg8PMj4ACT6Dp, dCFnA2kaP9MVYNhc, dwJNZnKW3ApzMCf6, f6mp4yJZwxAR7NLz, fFd4BgHwTaWXqJkp,
fTx6mHq2par1RPJX, gBDMLv8J1K7nqkpd, hDJKcgjxvWQfHt9M, hVJ6Km2AGkLPdt8B, hYjVWRNBGDKLxTpc, hZPWBMjdQgVJp4AK,
hfJzLYkNGarbVvWK, kVY4TxntBWArNjyK, kaPpXgt1bWJBr4z2, kph1Df7CqFVxQ2PW, mAW1KFR9YP8r3bNx, mahRZxnfrQ2N4jKL,
mbpQYGL8jaqgdrAz, mhZ8WvgabzLwD9FK, nLAb7KXamThYCp6d, pHqbZzVRTWy3N2nv, pzK2aRDNyrBG3AWX, q9K4Zv2mk3dnGaAP,
qgKwp4NVTZYJAzv8, qkctgyRMaFxXVnmz, rRG9MzWL4dQJ38hc, rpGZKzCgPdfm6RWM, rwxL1aVmWyBYA89g, tgd6xBMhfQvDq4K2,
tpKhjLGfwCzb6dA2, vdD1pLX6tkMamzqn, wvDFQB4cjG1dLyTh, x3njCmpkVYf7WDgh, ynL7Gb2MpkZT46j3, z7jmPBaWv4TQJ2AF,
zFXxNAn69HBGRaqg, zpw3WxckYDjXBJRy, DrnNChtLJb8pmAVd, GJwfBzhCqj6HtLKn, M3C87Yx2fKTvzgJ1, Tx7MPH9f8Ga6hdrF,
c1WPVLQ6Xw29RTCZ, vGJMahXtgTzFRmCQ, 6CN9XfdkHq148yZQ, Hwq97xLh6gv8ZPrY, qQ4BrAY1Xjh8Hfvy.

# EV-0136: Proc rates of 101 items and a set bonus from the Warcraft Logs sweep

- Claim: each item and set bonus below procs at the rate its section gives, priced on each landed hit's weapon base speed;
  a rate whose item has 5 or more wearers and a 95% interval within 10% of it either side is measured, every other one
  observed, except in the sections that name a later count, which take that count's test; and Windfury Totem's strike
  counts as a melee swing for the procs whose rows hear melee swings, Unsealed Wrath excepted. In these reports, up to
  2026-11-16, the strike grants no rage and spends a warrior Flurry charge; from push 114115 it grants a main-hand white
  hit's rage and spends no Flurry charge.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section's table names; 600
  reports in all for the rates, and the 209 reports the Windfury Totem section lists
- Game build: 1.60.1.70425 to 1.60.1.70938 by report date; each item's section gives its date span, and the hotfix each
  player ran is not recorded
- Date: 2026-10-08 to 2026-11-16
- Pull request: #1065
- Later count: Julie's Dagger and Waveslicer, §Count of the 2026-11-13 to 2026-11-17 reports
- Later count: Flame Wrath, §Count of the 2026-11-12 to 2026-11-18 reports
- Later count: Barbaric Crossbow, Bow of Searing Arrows, Burning War Axe, Coldrage Dagger, Cursed Felblade, Darkwater
  Talwar, Dazzling Longsword, Electrocutioner Leg, Fang of the Crystal Spider, Felstriker, Fiery War Axe, Fist of the
  Damned, Glutton's Cleaver, Grimclaw, Hammer of Expertise, Hanzo Sword, Iron Knuckles, Ironforge Greathammer, Killmaim,
  Lord General's Sword, Nightblade, Pendulum of Doom, Phantom Blade, Sang'thraze the Deflector, Satyr's Lash, Spellsever
  Crossbow, Strike of the Hydra, Sword of Decay, Tainted Pierce, Taran Icebreaker, Thornspike and Toofbreaker, §Count of
  the 2026-11-05 to 2026-11-19 reports
- Later count: Iceblade Hacker, Loksey's Training Stick and The Blackrock Slicer, §Count of the 2026-11-09 to 2026-11-19
  reports
- Later count: Blade of the Wretched, Bloodrazor, Bloodspiller, Boneslicer, Cookie's Tenderizer, Eater of the Dead,
  Firebelcher, Hammer of the Northern Wind, Kang the Decapitator, Mana Channeling Wand, Pyromancer's Baton, Scorpion
  Sting, Searing Blade, Serpent Slicer, Sliverblade, Sul'thraze the Lasher, Supercharger Battle Axe, Sword of
  Corruption, The Cruel Hand of Timmy, The Ziggler, Twilight Forgehammer and X'caliboar, §Count of the 2026-11-06 to
  2026-11-21 reports

## Method

A proc is an event of the item's proc spell on the wearer or the target (damage, an aura applied or refreshed, a heal,
extra attacks, energy), counted once when one proc logs several events within 50 ms; a proc that misses or is resisted
still counts. The hits are the wearer's landed, non-periodic hits the trigger's proc flags hear: white swings of the item's
hand and melee-defense hits on the main hand (abilities, seals, Judgement of Command, Windfury Totem strikes) for a melee
proc, Auto Shot, Shoot, Throw and ranged-defense hits on the ranged weapon for a ranged one. Left out: periodic ticks,
damage shields, and Volley, Shank, Shield Slam, Shield Bash and Thunder Clap. A dual wielder's white swings are split
between the hands by swing timing: each hand swings at its weapon's base speed over the haste of the moment, so the
swings form two interleaved chains, and the few a chain cannot place fall back to the damage before crits or the weapon
speeds (each run's table row names its split). A druid run in Cat or Bear
Form is left out, a run logged in two reports counts once, and a stretch with no proc where several were due is dropped
as the item swapped away. The rate is the procs over the sum of each hit's weapon base speed over 60, so haste raises
procs a minute without moving it. Its interval is the wider of the per-hit Wilson interval and 2,000 bootstrap draws of
whole wearers.

The method reproduces Heart of Wyrmthalak (EV-0133, 1.92 against 2.00), Skullforge Reaver
(EV-0121, 1.98 against 1.95) and the Highborne Research Tablet's weapon hits (EV-0131, 2.70 against 2.71). It reads
Flurry Axe 1.79 against EV-0120's 1.65: the warriors' off-hand Whirlwind hits roll the axe and are not counted as hits
here, which raises their runs to 2.17 and 2.83 (the others read 1.57 to 1.76), so no item rolled mostly by the off hand
counts as measured, except Satyr's Lash, whose hits fall half in each hand, by the owner's ruling. Against clean damage
splits the timing split agrees on 86% of swings and puts the main hand's share a median 0.025 from the weapon speeds'
share. Items with fewer than 50 procs, or one wearer with fewer than 100, are left out.

Each table lists every run looked at; a run left out names why. PPM is the run's procs over its expected procs at 1 PPM.

## What the procs do beyond their rate

Read from the same wearers' streams, to check what the rows state before the sim models it.

- Mana Channeling Wand (18483): all 211 procs of 1302350 restore 28 mana, the row's amount; the tooltip's 65 is the
  trigger 1302348's.
- Hammer of Expertise (8708): Forging Strike (1317432) is Physical on the melee attack table, as its row states: of 614
  hits, 501 hit, 73 crit, 14 were dodged, 11 parried, 12 blocked and 3 missed, and 49 show armor mitigation.
- Sul'thraze the Lasher (9372), 7 wearers, 958 landed Rage hits: all 487 timed Shadow Lash hits landed 5.0 s after the
  Lash went up or refreshed; 46 of 511 Lash damage events crit and 2 were immune. Of 111 Rages landing while a Lash was
  running, 99 put up no new Lash, 83 of them followed by the running Lash's damage at its 5 s. No missed Rage (22) put
  up a Lash.
- Pendulum of Doom (9425), 7 wearers, 615 landed hits: 469 sit in the plain damage range and 146 at about 1.5 times it
  (per wearer 1.38 to 1.50); the larger ones land a median 1.7 s before the wearer's last hit on that target, against
  15.1 s for plain ones. The logs hold no target health, so the threshold is the row's.
- Flame Wrath (11809), 8 wearers: the fire shield went up 297 times and refreshed 130 times, 195 windows lasted exactly
  15 s, and it dealt damage 2,292 times inside its windows and 47 outside, never a crit.
- Gahz'rilla Fang (9467), 14 wearers, 594 non-crit hits: the damage fits 85.8 plus 1.97% per Static Electricity stack
  (the row: 84 times 1 plus 2% a stack), and 95 procs at 50 stacks dealt a median 168. 581 stack removals left exactly 1
  stack, 564 of them at a proc's timestamp. Single-stack gains came 50 ms to over 3 s apart, with no steady cadence.
- The Chief's Enforcer (9477), 13 wearers: 416 strikes applied or refreshed the stun, which also landed on 20 of 26
  dodged strikes, 6 of 12 parried ones and 3 of 5 misses; 56 stuns were immune, 2 resisted, and 305 of 389 lasted the full
  1.5 s.

## Windfury Totem's strike as a swing

- Reports from 2026-11-11 to 2026-11-16; the reports are listed at the end of the section
- Scripts: `chain.py` for procs and `charges.py` for spent charges, run by `run_all.sh`; `reckoning.py` for Reckoning's
  extra attacks, `swingtimer.py` for the swing timer, and `paired.py` for rage

A strike counted here is a landed Windfury Totem strike (10610 and the lower ranks 10608 and 8516) that follows a landed
melee ability (melee DefenseType) within 60 ms, where the wearer lands no other melee hit in the 100 ms before the
ability or in the 100 ms after it. Its proc is an event of the proc's spell in the 100 ms after the ability. A melee
ability alone almost never procs these rows, so the rate after an ability and a strike is the strike's own; the
ability-alone rate is in the table beside it. The white swing rate is the same count after a landed white swing with no
other hit within 100 ms. A charge spend is counted only while the charged aura is up, and for both Flurries only where
neither the hit nor the strike crit, since a crit sets the charges again. Intervals are Wilson intervals per hit.

| Proc                                      | Row     | Wearers | Strikes | With proc | Per strike        | After an ability alone | Per white swing   |
| ----------------------------------------- | ------- | ------: | ------: | --------: | ----------------- | ---------------------- | ----------------- |
| Unbridled Wrath, rage                     | 12322   |      78 |  11,785 |     6,418 | 54.5% (53.6–55.4) | 0.26%                  | 55.1% (54.8–55.4) |
| Flurry (warrior), a charge spent          | 12966   |      36 |   1,880 |     1,861 | 99.0% (98.4–99.4) | 0.56%                  | 98.3% (98.1–98.5) |
| Reckless Swings, rage and a charge        | 1329555 |      46 |   1,475 |     1,469 | 99.6% (99.1–99.8) | 0.47%                  | 97.5% (97.1–97.8) |
| Quick Strikes (Aspect of the Beast)       | 1299447 |      16 |     727 |        62 | 8.5% (6.7–10.8)   | 0.24%                  | 8.6% (8.1–9.1)    |
| Rogue Armor Energize (rogues)             | 27787   |       5 |     737 |        15 | 2.0% (1.2–3.3)    | 0.05%                  | 2.5% (2.2–2.8)    |
| Rogue Armor Energize (druid)              | 27787   |       1 |      79 |         1 | 1.3% (0.2–6.8)    | 0 of 334               | 1.8% (1.1–3.1)    |
| Reckoning, the stored charges spent       | 1326084 |      18 |      25 |        25 | 100% (86.7–100)   | 0.36%                  | 80.4% (74.4–85.3) |
| Battlegear of Valor, heal                 | 450587  |       9 |   1,299 |        37 | 2.85% (2.04–3.87) | 0.03%                  | 2.68% (2.28–3.14) |
| Flurry (shaman), a charge spent (EV-0010) | 16257   |      55 |     357 |       286 | 80.1% (75.7–83.9) | 4.1%                   | 94.3% (93.9–94.8) |
| Seal of Command                           | seal    |      37 |   1,043 |       313 | 30.0% (27.3–32.9) | 0.73%                  | 25.7% (24.9–26.4) |
| Seal of Fury                              | seal    |      50 |   1,450 |     1,095 | 75.5% (73.2–77.7) | 1.5%                   | 69.1% (68.6–69.6) |
| Seal of Righteousness                     | seal    |      36 |     821 |       255 | 31.1% (28.0–34.3) | 0.89%                  | 39.3% (38.6–40.1) |
| Unsealed Wrath, an Echo stored            | 1310735 |      26 |     519 |        10 | 1.9% (1.1–3.5)    | 1.0%                   | 66.6% (65.5–67.7) |
| Bloodthrill                               | 1289682 |       9 |   1,633 |        10 | 0.61% (0.33–1.12) | 0.74%                  | 3.5% (3.1–4.0)    |
| Lightforge Armor, Crusader's Wrath        | 450625  |       2 |      33 |         0 | 0% (0–10.4)       | 0 of 375               | 3.4% (2.1–5.4)    |

Every row whose proc flags hear melee swings and nothing else rolls the strike as a main-hand swing. Unbridled Wrath,
the warrior's Flurry, Reckless Swings, Quick Strikes, Rogue Armor Energize and Battlegear of Valor roll it at the rate
a white swing does; the shaman's Flurry and Reckoning hear it too. A seal's row is each rank's own. Unsealed Wrath is the
exception: a strike stores an Echo no more often than the ability before it does alone. Bloodthrill, whose row also hears abilities but which procs on Heroic Strike and Cleave alone (EV-0056),
does not hear it either. Lightforge Armor's two wearers are too few to say. The seals hear the strike as the sim has
them.

A strike spends Reckoning's stored charges, landed (101 spends) or avoided (87), but it brings no extra attack: none of
the 26 spends by a strike that followed an ability logs Reckoning's extra attacks 20178. A strike that follows a white
swing, which may spend the charges itself, logs them on 74 of 160 spends, so those are left out.

A white swing that a strike follows spends 1.97 warrior Flurry charges and 1.96 Reckless Swings charges, against 0.99
and 0.98 for a white swing alone.

The strike does not reset the swing timer: the gap between the swings before and after an ability and a strike is a
median 1.002 of the wearer's median swing gap (3,314 such gaps), against 1.001 after an ability alone (26,063).

In these reports the strike grants no rage. Cast events carry the warrior's rage before the cast's cost, so the gain
between two casts is read straight off them. Over gaps under 600 ms with no damage taken and no resource change, 88 of
90 strikes from 13 warriors added nothing over the same warrior's gaps opened by the same ability without a strike. The
strikes dealt 1 to 509 damage, median 204, where a white swing of about 200 damage gives 10 to 19 rage. From push
114115 the strike grants a main-hand white hit's rage. Those runs are
DGFd8KtQk19hXabN fight 5, rfg8k3Nw1P7TAHx6 6, AfJ6DYbh9dmkgHZL 7, cgMGCdPJQ84wzaBV 5 and 183, xRK37aCvtVfj26kL 4,
FN8krKfCb3BXWmwV 3, JRQqNL6DkvWKYdPB 2, 4Tc3yBNVYn7wW8RJ 63, 7fy93LdcwQhvF2gP 5, 1JLq46yg7wYNhan9 1,
jXTVfy3ApWKPYwDG 3 and 4YVTAqgG3dtmD8C9 4.

The reports of the proc and charge counts: 123jYCvkVGrHfXWt 176QTFWR8PtdfGZB 1CYTfP3pFDcaRrJm 1DbXf9pPdZWKQTam
1FNGzPgAaX6wDpBt 1FV3WyfXGCcPrT9b 1HTPQGZya7pAz3qh 1JLq46yg7wYNhan9 1NBKq2mDWghwR69z 1PcrGTyatDAFx8Yk 1QpP6DrkacvCjdYy
1TZvYG2tfR9wVPNH 1WHqx7bPa8mkGydJ 1nLR3XCxjcAdpKtB 1q2tTDcaFMnfzvkB 1qjDRYwtHxyJPfQn 1qpQ6FW8fxD9gC4G 1rYynxW6AGfcQz9t
1z4vRDqGdM2cVPXk 214ndcFyjpPM8wah 28AhM3VGnvFcJ4rD 2AvFk4PJmBhyb89n 2FkYRHB1dWhra8vt 2WH4jLt3FADy8fnh 2Z4G3r9pfqJtCM6W
2b3Tf7QGKHA1WMxD 2wMznDRvmkpbtHxA 3B8AcaGNRKmdgPDj 3JbjLGXVzctAmfaD 3PAwV6XFaB7hqDgz 3mwdt9jJF6Tyb1q2 3qfDVTZ1J29zPChn
3zTJ2VXvQ8GCyFbc 49dLXZ3g7NhzqABb 4C28YPXqgTMQJxB7 4HC91qxpR6hzKWdF 4LrMDtG6T8AVRJzQ 4Qm6MgZWBTx78Lvp 4RVqrpXkh7dKftAC
4Tc3yBNVYn7wW8RJ 4bqYVtkjL6aKdwQM 4f2ygdYCzFAvqc6K 4h1PjwkaKX9q8mgT 4wTtzhPfDqnQWNy7 61Hx8DgwhdbWaN9C 61VcxvhfQF2DRAGd
6kN78G2xfqy3vwTP 6kf3tMF4wCPNcRnh 6vFAMDQZ2RmaxC7n 6vKFBQNC1fMqR3Wa 6xHzpKd8wCna37Lb 6yHgkT8LqP42NmxY 7AxaTpFZMqGb9wNd
7M3gGcH6bprxKqZY 7c21XafhWTAdqNGY 7cpmjq6DTBNLwFhG 7fy93LdcwQhvF2gP 7fz9Kk8rY6jyQAbL 7kxNnWGw2FzfLJtR 7n6rKTfk1vhzgXyF
7rGZbTJtpAqnBh2Y 7rzAygVjwH2C69Tv 7tLY1pVK9wRWPABX 8WfLRh1qTpFxJgra 8cVPaCpAMmv93bkD 8frHF64GzAhBxDyb 8n6w19TGyhZtmKfd
8yPRmZdrq4QhcGn1 9MwvzA3QN7CtHcyg 9Z473MfXLYqcyWPg 9hywC8jMTLV36Pxd 9nJGwhXTd8pvraxb 9zMva8xTVhFkHPAt 9zxHaWbYT3gqDQvF
A3hT8CQBPJXWmtZM APgTZrbzNvykx1JG ARFXzbqHDpBLJdhy AdfMXDLW29k6pGy1 AfJ6DYbh9dmkgHZL AvNwTdyBbDncFf2h B9gTmLcFW3R7GAHt
BMJxG63gcbQX8tWh BQnXb3hTC4rmJ2pd Br2pycwj4zLC1ZP7 Bz8CAjZnNyXhgTtH C8Hk1q3rpN27YWam CWhG6LyXk3PBHdMQ CfHvwytKJ6P4q7ag
CtMF9BgnAmWkTbfd D463FZBX1C9maJ7b DGFd8KtQk19hXabN DH3dT6tYLKmMrNz8 DNfvpnLRcPJyMKQA DW9YkdvyrjtMBQnf DWfMg7hzPF4d1mQ6
Dfa6GMQ2pmX8cBNz DpVmH4WRxL72nGwd FHqYCNcKntvDAjxg FHz1nBG8rhgMqaxW FMaAjHcwCQLkpV32 FN8krKfCb3BXWmwV FVqQab31p7KXAJzt
GRnMHmK69BXfWVTj Gyw2T3LjptRW6vH9 H2GvZtVYNy4QzqLw H2T7NDqBgm6LxGYK HXg2W4bTFDjcRJ87 HmF9nyBdrvAL1wKt HwqzG1dpaPXNQAFg
HxCQat38vmjnkpb6 JMnDtLGxf4FqCV6w JRQqNL6DkvWKYdPB KYmwz2q1tG8yAcdP LMdRqJGhV17XwWyT LPQd7pfbjDkYZyRz Lg82ZFDyQ4MvRm9p
LxdYWR47ZFbTNBhM Lxy1kqgVvFc24aPr MWrPDAVFCBXzh8Z6 MZRgJXdTb9mKLqcv MwVH1KQmak73NPd8 N3Y1mFkJ9Dwr74bA NXMZV9xjvWpB7GtP
NdTXJzkWHLZrVQYm NrpM7DJQBKWqbzVF P1LGvYV9FdmR8x3n PYQf86cTrZ1GjdKx Pjz8kK2XvxyQrTCB PtmYHjkDrqhdR9MA QRCgnhck3KVq6vBT
QTt29xVHrZa8mDBw Qb1yWxa8vHCGL7Yr RfT7K4v2jhY9FAzD RpFQn1ftYCAgMW47 RvGHZ6wnXpCyq3dB RvcrA3P9jznxhHVQ TacY2wkjAgXynNHJ
Td2y9DjVvn1YgaHr V3L7kJDzTGg4QqCF VYBjgJx8aZtRhmpy VYb6f27NazpmCch3 WYFZHxvL7zg6ydnQ Wf8ghC6TtQKrkXdv WygY76wT4zqbvCM2
X186Gx2gRBZChbVA X4QvgN2caZBbfnwF XWfw2PzpNcBhbayY YFtvD87kznr1CThg YWZm7NKhFvqbMfgy YZ6NdFDRQKL2CTmv Z78jbgTL9V1Yfahq
ZDWyQGqFBvagT7kA ZNyPLR2vkjMGptCJ ZqghBpkcy7xtPfmH aAcjGPbyZXNTzR1K acpjMDfJkLPABvxT b12KD98fndF4hMq3 bnRJKL9FTMq2zkQG
cFRpkf1jWTaHrgL3 cLkwmZbzFDKdPhYn cgMGCdPJQ84wzaBV cr6qMtj1vCTPHVfw dBNfgVDLMmnc8h4v dV2cFGtPM8ZajTYX daT4FcQgCYw2WBq6
fywmzq86PRtjDJra fz2HQxXk96ywVYv7 g8b1MkJnKtyfzvPa gBDMLv8J1K7nqkpd gmKvNqHFJLZXfdnP gtNvbhF9aZCXyzVk hCPLc2G7KAVX8Bby
jJWt7Dnm4qc9rzQd jLTw4rQd1BKgMkhy jXTVfy3ApWKPYwDG kv32zprHjRfnZPh9 mCqj9fpbthFNGgkc mRxb9L2wNkMJTWXY mdXkc6xwP8Y9pR47
mg3MBw6rcNtZpjbA mqzgv4rPVxA3JHGf nCKwNLGVhPMFjq9Y nJDNV6Md3bqpXZT9 nX1qbatTzGZ3CAwB nyMqbY64tL7HhKQJ q23Tz8hnZGkvbLx6
q68xzRHYhCpFLjfJ qcrKJGPvkfhAtdjQ rAFM4QTbC8jHypZL rfg8k3Nw1P7TAHx6 rnfRqHt9Vk8g1JB7 rphdM14FbnTBaj3V rwxQgM38abAcmRVv
tpNByDzWPvrVwYx6 v8Zp2QF7An4d9xMj vwCXzhMQ2drgpRxD wJtPdpxFGQCM9f3c waZyxVKk43BzGLN6 xRK37aCvtVfj26kL xVCyPt4A9HYL7MRc
xXWkB3fqAjt9NJGH yHRFJ8kn9vQA4YLr yhW9cLwxnHNDP2jF ynVw7ZgMx4tjF3fQ zFXxNAn69HBGRaqg zLacx86Thr3AMtqn zbg1wFp2L8hWHZva
zmDY4LVX7N8vh9aR

## Axe of Inevitable Ruin (274484)

- Proc: weapon proc 1294226, logged as 1294226; rolls on main hand
- Sim rate: 2.15 PPM, observed; measured 2.16 PPM [1.98, 2.35], 4 wearers, 4 runs, 469 procs on 4351 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ---------- | ----: | ---: | ---: | -------- |
| LMdRqJGhV17XwWyT | 16-22  | Shaman | mh   |            |   169 | 1581 | 2.14 |          |
| MpcVyhTHnqCJvDFx | 1-6    | Hunter | mh   |            |     9 |   62 | 2.90 |          |
| mqzgv4rPVxA3JHGf | 1-2    | Shaman | mh   |            |    79 |  713 | 2.22 |          |
| yhW9cLwxnHNDP2jF | 4-8    | Shaman | mh   |            |   212 | 1995 | 2.13 |          |

## Barbaric Crossbow (272999)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1291551
- Rate: 2.95 PPM, measured

3,794 procs on 38,444 rolling hits, 58 wearers, 72 runs, in the ranged slot: all hits 2.96 [2.89, 3.03]. The rate is all
hits, which the specials rate and the lone Auto Shot rate agree with. Lone Auto Shots: 3.00 [2.88, 3.11], 2,070 procs on
20,711 lone units, 58 wearers. Specials: 3.11 [2.96, 3.30], 1,002 procs on 9,653 lone units, 45 wearers. Direct: window
[-10, 100] ms, median 19 ms, 98th percentile 44 ms, 2,432 isolated hits. Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Hunter  |      45 | 3,786 | 38,317 | 2.96 |
| Warrior |      17 |     6 |     96 | 1.87 |
| Rogue   |       7 |     2 |     31 | 1.94 |

Reports and fights: 1AfDWYQT7Xjd24K3 13-14, 16; 1Gg92cW6XBhkwYb4 1-5; 1RrDAhp7ZvVYc9P3 1-6; 1W9N26dHfnPV4G7A 1, 3-9;
1y6GV7KLTZPrJH2w 1-3; 2GbXqD7KcWTfRnYa 1-2; 2MHXV8Y9jx43vhpk 3-12, 14-18; 2RygmFXpHCbnazVA 1-5; 2tJkzh16X38wrQfP 1-3;
3pTj2FQg6VvJz4DP 4, 6-7; 4ChPnBrVK2fJLMaX 4; 72XWYanybGdwPg41 1; 7VkrJpR8XyFtL6Hm 1-2; 7vyJCbH4kw8ZdfFV 1-10;
8Xxb9LDGHP7vcJfh 1-5; 92JNtax7RTkQ6cHv 1; CVFcbDGj6n8WMKyL 1-2; DACv8aKzbkxfGygr 1, 3-4; DZvkgfwrVLbpmWyQ 1-2;
DaNQ6WPf7GqTpKVk 1-8; Ghp9gxWFkBNacLfn 1-5; GkLDK8WpcdvPatfB 1-2; Gm7xJ3W81tHyTfXc 1-3; H9QxwyGt4z8MmqCj 1;
HwqhFkWZCBRTQVDf 5-7, 9-13; J4mG8yjgVCALMfRw 1-2; KBzML9kdHFCXVpQc 1-2; KYD67nVqZT4jxWyN 5-6; L7QHtkhacypgd3D1 1-2;
LnPRx96hKZN3CgTF 1; MfD1n7XFZrG4dcbH 1516 fights from 1 to 1743; NQyGkApg7zMc18La 1-3; NyG7zfKj2Mvk1FLJ 1-6;
P7MBDvyCRQtLkJHz 1, 3-8; QtGYMWzJZPNpmjvB 1; T6MXt73janNJgmyY 1-3; Td29Nzahj1cmGMWZ 1-2; TdXmWDR3QFJwpYfq 1;
VXcPMWHwZJftbCrm 1-9; VcDy8ZTRgQpWrqXz 1-3; WN1vpX6t7JkVDYZf 1-6; XV7Cf348hazGFg1t 1; YHh2t9agcvQjTAZq 2-3;
Z9CQ6fqYjzcB4tG2 1-3; aMQ63JtDnpmFz9WR 1-5; akDmAcjZrpgTFCQM 1-13; bLd41TDnmVYJZj2x 1-3; cAdCm1NLGaQnXrHW 1-4;
dc3Y84hWFnrA2wqp 1-3; dtLy89QkrcHx3pMf 1-2; fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; gBxakRwhQDvnLXp1 1-2;
hMdFnHqgVRDTyKPY 1; hRzbpMrDY3cgQX7n 1-7; jRytczTZF6pBJgYL 1-2; kBHbV3DZjamvfJY4 1-3; nALy2WrwHC9qdMFm 1-2;
p9HgV7zdCKqbrwJ3 1-3; ph6FqnCZaRQLvxbM 1-2; pw8aYCczKgjMbvAd 1-4; r2BYfmZW1wkLDdNx 1-2; rKXWhHqLza39gkmA 1-2;
tbWaYgLTw8XQAzdC 1-3; v9fnrCd8TLaPDYH4 1-2; vqpcV7yDGCkbftmP 1; vrZD4X7Vb3pW9FNK 1, 3-4; vxkFd3MfgJX7NyAc 1, 3-8;
wmT1zqYBat3fHDLy 1-5, 8-9; xrGFH2DzcNBZq3Cv 1-3; zcL8MvJnPXxYhNyC 1.

## Bashguuder (13204)

- Proc: weapon proc 17315, logged as 17315; rolls on main hand, off hand
- Sim rate: 3.2 PPM, observed; measured 3.21 PPM [2.72, 4.18], 7 wearers, 7 runs, 955 procs on 9926 hits
- Reports from 2026-11-11 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                          | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ----------------------------------- | ----: | ---: | ---: | -------- |
| FHqYCNcKntvDAjxg | 34-36  | Warrior | mh   |                                     |    92 |  939 | 3.27 |          |
| JRQqNL6DkvWKYdPB | 1-7, 9 | Warrior | mh   |                                     |   367 | 4131 | 2.96 |          |
| VK83G7MdBhLRFAfa | 1-42   | Rogue   | oh   | timing + otsu for 76 swings         |    58 |  629 | 3.08 |          |
| VYBjgJx8aZtRhmpy | 38-46  | Warrior | mh   |                                     |   160 | 1041 | 5.12 |          |
| VYBjgJx8aZtRhmpy | 44-46  | Rogue   | oh   | timing + speed ratio for 244 swings |    61 |  867 | 2.35 |          |
| v8Zp2QF7An4d9xMj | 1-10   | Warrior | oh   | timing + otsu for 71 swings         |   117 | 1662 | 2.35 |          |
| xXWkB3fqAjt9NJGH | 1-5    | Warrior | oh   | timing + otsu for 129 swings        |   100 |  659 | 5.06 |          |

## Battlegear of Valor (set 189)

- Proc: 5-piece set bonus proc 450587, logged as 450589; rolls on melee swings
- Sim rate: 1 PPM, observed; measured 0.88 PPM [0.74, 1.00], 9 wearers, 9 runs, 201 procs on 5895 hits
- Reports from 2026-11-14 to 2026-11-16, game build 1.60.1.70938
- Counted apart from the sweep, on the same pricing: the hits are landed white swings of either hand and Windfury Totem
  strikes, each on its hand's weapon base speed; a dual wielder's swings are split between the hands by damage (Otsu),
  and the interval is 4,000 bootstrap draws of whole wearers
- The Battlegear of Heroism reissue's 4-piece bonus is the same row; no wearer with it active was measured, and the sim
  gives it this rate

Landed white swings and Windfury Totem strikes roll it, the strike as a melee swing (§Windfury Totem's strike as a
swing). Of the strikes that
follow a melee ability with no other hit within 100 ms, 37 of 1,299 procced it (2.85%, 95% 2.04 to 3.87), against 2.68%
(2.28 to 3.14) on isolated white swings. None of 230 landed Heroic Strikes or 1,864 landed Cleaves is
followed by a proc, and Whirlwind, Shield Slam and Revenge are followed by one about 0.1% of the time, as chance would
put a proc next to them. The heal logs about 20 ms after its swing, and two procs logged in the same millisecond, so
there is no cooldown. The chance per swing rises with the weapon's speed: 2.5% on the single-weapon wearers' 1.8 to
2.2 s main hands, 4.3% on their 2.7 and 2.8 s ones, so the rate is per minute. The six wearers who mixed the two sets
(four Valor pieces and one Heroism piece, or three and two either way) never procced.

| Report           | Fights            | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out           |
| ---------------- | ----------------- | ------- | ---- | ---------- | ----: | ---: | ---: | ------------------ |
| 1JLq46yg7wYNhan9 | 1, 3-17           | Warrior | set  |            |     0 |      |      | 4 Valor, 1 Heroism |
| 2wMznDRvmkpbtHxA | 7                 | Warrior | set  |            |    15 |  437 | 1.14 |                    |
| 7fz9Kk8rY6jyQAbL | 1-2               | Warrior | set  | otsu       |    30 |  667 | 1.20 |                    |
| ARFXzbqHDpBLJdhy | 1                 | Warrior | set  |            |     0 |      |      | 2 Valor, 3 Heroism |
| AfJ6DYbh9dmkgHZL | 1                 | Warrior | set  | otsu       |    43 | 1281 | 0.91 |                    |
| D463FZBX1C9maJ7b | 7                 | Warrior | set  |            |     5 |  316 | 0.43 |                    |
| DGFd8KtQk19hXabN | 1-2               | Warrior | set  |            |    33 |  741 | 0.95 |                    |
| FN8krKfCb3BXWmwV | 1, 3-4, 9         | Warrior | set  |            |    24 |  987 | 0.66 |                    |
| NdTXJzkWHLZrVQYm | 1                 | Warrior | set  |            |     0 |      |      | 4 Valor, 1 Heroism |
| VtjvpP8bwcgCYTXF | 3, 5-18           | Warrior | set  |            |     0 |      |      | 4 Valor, 1 Heroism |
| YFtvD87kznr1CThg | 1, 4              | Warrior | set  |            |     0 |      |      | 3 Valor, 2 Heroism |
| cgMGCdPJQ84wzaBV | 3, 5-6, 8, 10, 12 | Warrior | set  |            |     0 |      |      | 4 Valor, 1 Heroism |
| dBNfgVDLMmnc8h4v | 10                | Warrior | set  |            |    12 |  520 | 0.66 |                    |
| rfg8k3Nw1P7TAHx6 | 1, 3              | Warrior | set  |            |    28 |  635 | 0.98 |                    |
| rnfRqHt9Vk8g1JB7 | 1-2               | Warrior | set  |            |    11 |  311 | 0.76 |                    |

## Black Duskwood Staff (937)

- Proc: weapon proc 18138, logged as 18138; rolls on main hand
- Sim rate: 0.9 PPM, observed; measured 0.88 PPM [0.68, 1.13], 2 wearers, 3 runs, 57 procs on 1386 hits
- Reports from 2026-11-10 to 2026-11-15
- Notes: 1 of 4 runs are feral druids, left out (paw speed, not the weapon); few wearers (2): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ---------- | ----: | ---: | ---: | -------- |
| 4bqYVtkjL6aKdwQM | 1-2    | Shaman | mh   |            |    18 |  366 | 1.05 |          |
| 9RgXYmBfN7hCdznQ | 1-3    | Shaman | mh   |            |    19 |  510 | 0.80 |          |
| waZyxVKk43BzGLN6 | 1-11   | Shaman | mh   |            |    20 |  510 | 0.84 |          |
| 6GVg7FYcHWJmrbKw | 1-16   | Druid  | mh   |            |    14 | 1270 | 0.24 | feral    |

## Black Malice (3194)

- Proc: weapon proc 1292155, logged as 1292155; rolls on main hand
- Sim rate: 2.05 PPM, observed; measured 2.06 PPM [1.70, 2.43], 5 wearers, 5 runs, 246 procs on 2175 hits
- Reports from 2026-11-10 to 2026-11-15
- Notes: 1 of 6 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 4WrfjAqJYhxGHwPd | 1-3    | Warrior | mh   |            |    58 |  415 | 2.54 |          |
| 7APg28BmdtzHhQFV | 1-6    | Shaman  | mh   |            |    20 |  203 | 1.79 |          |
| GV14JHxQP3wcbg9X | 1-2    | Paladin | mh   |            |    73 |  544 | 2.44 |          |
| PAy6j9tqdkwMTQHn | 6-10   | Paladin | mh   |            |    73 |  803 | 1.65 |          |
| VLcP3WgkCfKZhB2v | 5-8    | Paladin | mh   |            |    22 |  210 | 1.90 |          |
| 84aWfvkV2qb3mAnL | 1      | Druid   | mh   |            |    23 | 1355 | 0.31 | feral    |

## Black Menace (6831)

- Proc: weapon proc 1293407, logged as 1293407; rolls on main hand, off hand
- Sim rate: 2 PPM, measured; measured 2.01 PPM [1.82, 2.21], 15 wearers, 15 runs, 792 procs on 15762 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights              | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------- | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 3GT4tz8gxnphcbkD | 1-7                 | Rogue   | oh   | timing + otsu for 20 swings |    56 | 1016 | 2.20 |          |
| 3pCKd6jqgHyFc9kf | 1-17                | Rogue   | oh   | timing + otsu for 35 swings |    42 |  895 | 1.88 |          |
| 3wjCThZxkVFWzG69 | 1-6                 | Rogue   | mh   | timing + otsu for 7 swings  |    70 | 1527 | 1.83 |          |
| 3zTJ2VXvQ8GCyFbc | 1-13                | Warrior | mh   |                             |   103 | 1526 | 2.70 |          |
| 3zTJ2VXvQ8GCyFbc | 1-13                | Rogue   | oh   | timing + otsu for 59 swings |    29 |  669 | 1.73 |          |
| 4g9w7aLvtz2RmcFx | 1-16                | Rogue   | oh   | timing + otsu for 9 swings  |    23 |  602 | 1.53 |          |
| 6Ax3QmHWJqnPzvhd | 1-3                 | Rogue   | oh   | timing + otsu for 18 swings |    31 |  688 | 1.80 |          |
| 6RMV4ya9h1pWjYLn | 1-3                 | Rogue   | mh   | timing + otsu for 24 swings |    76 | 1739 | 1.75 |          |
| 8akXKLNxCdDtpqfv | 1-4, 7, 9-22, 24-38 | Rogue   | mh   | timing + otsu for 20 swings |    42 | 1095 | 1.53 |          |
| 9Z473MfXLYqcyWPg | 1-2                 | Rogue   | oh   | timing + otsu for 93 swings |    36 |  749 | 1.92 |          |
| BTRMkxmvfrwcZNaG | 1-3                 | Rogue   | mh   | timing + otsu for 14 swings |   103 | 1728 | 2.38 |          |
| cFRpkf1jWTaHrgL3 | 1-2                 | Rogue   | oh   | timing + otsu for 25 swings |    34 |  488 | 2.79 |          |
| hfmVrK1Aw7B3yqWF | 1-2, 4              | Rogue   | oh   | timing + otsu for 52 swings |    43 |  840 | 2.05 |          |
| tac7hXVRgkbZNMDy | 3                   | Rogue   | oh   | timing + otsu for 4 swings  |    35 |  601 | 2.33 |          |
| x3PftMnXCBZgLAG2 | 1, 4-9              | Rogue   | mh   | timing + otsu for 3 swings  |    69 | 1599 | 1.73 |          |

## Blade of the Wretched (10803)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 18088
- Rate: 1.9 PPM, measured

1,840 procs on 27,357 rolling hits, 19 wearers, 19 runs, in the main hand: all hits 1.92 [1.81, 2.04]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.58 [1.31, 1.88], 122 procs on
2,210 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 1.82 [1.65, 2.03], 294 procs on
4,612 lone units, 14 wearers. Specials: 2.10 [1.85, 2.29], 223 procs on 3,039 lone units, 18 wearers. Two-roll
abilities, priced as one roll: 3.05, 21 procs on 197 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 461 isolated hits. Dropping the swapped stretches (87 active minutes, 10% of hits, in 3 runs) moved the
read from 1.74 to 1.92. The paladin check, which sets no rate: lone white swings 2.37, 20 procs on 241 lone units, 5
wearers. By report date: before 2026-11-18 1.90 (1,402 procs); from 2026-11-18 to push 114313 2.00 (438 procs). Reports
from 2026-11-12 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 |   847 | 13,708 | 1.77 |
| Rogue   |       8 |   745 |  9,943 | 2.14 |
| Hunter  |       3 |   248 |  3,706 | 1.91 |

Reports and fights: 34gy17FXnaGhVC2M 2; 6kN78G2xfqy3vwTP 6; 7M3gGcH6bprxKqZY 1-3, 5-8; ATMmnV2tC3ZPb6G8 1, 3;
bGRfPhMWCrKg9ydV 2-3; c69WN31xbHBPtmhv 5; cbkDQfXNPGj3tnWT 1; dwcgGfFmAXWK7Mjz 1-2; hDGL1wmpV4WjxbHP 1-3;
jJYHChwVBRaDxNWk 16-32; kz4FVdBRqpJwj8nx 1-3; mApW49VaFD2TKyg1 1-2; MHL1xhpwC4dcVDZK 1; N2m8nPTGJQrj1DhH 4-5;
Ncp4bVY3x1DXRqvA 1; VPG6FyAdJ7NcC29X 1-5; VwyW3ZP1H6RTpdmf 1-5; ZD3cfJpHa4vGYjKF 1; ZfRYBTWb168N7mzQ 2.

## Bloodrazor (809)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 17504
- Rate: 1 PPM, measured

2,309 procs on 52,506 rolling hits, 30 wearers, 41 runs, in the main hand: all hits 0.98 [0.93, 1.02]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.05 [0.81, 1.15], 167 procs on
3,530 lone units, 5 wearers. Dual-wield white: 0.60 [0.45, 0.73], 138 procs on 5,114 lone units, 25 wearers. Specials:
0.88 [0.79, 0.98], 263 procs on 6,651 lone units, 30 wearers. Two-roll abilities, priced as one roll: 2.83, 14 procs on
110 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 403 isolated hits. Dropping the swapped
stretches (6 active minutes, 0% of hits, in 1 run) moved the read from 0.97 to 0.98. By report date: before 2026-11-18
1.01 (843 procs); from 2026-11-18 to push 114313 0.96 (1,466 procs). Reports from 2026-11-13 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      19 | 1,392 | 31,593 | 0.98 |
| Warrior |      11 |   917 | 20,913 | 0.97 |

Reports and fights: 3XCdM1LQRNmvY8kp 1-8, 10-12; 76zv93gbJQFrWKfd 1-30; 7mFKVrW6QLHy3fgk 1; 8Cx9vcnzpJwk34WA 1-11;
8GFRaPBnKWf7Qrpb 1-3, 5-6; a4P93RWJwdb1c2QY 1; aNh3Dt8kbZHL9cBM 1; AZchwX93DmyHLbv6 1-3; bGjk9YQT8DgdKFPf 1;
byzq9hxPvpDAJac2 1-12; d1AC7qZ8h2tNnFLK 1-2; DcVtKxdCXMga9QZf 1; FvqNn8XAhHD6M4pY 1-10, 13; hYGRt9H1NXJWMZbB 1;
j1xyYV7gcMH69ND8 1-20; jJWt7Dnm4qc9rzQd 1; jr7TAd93qLPxWzhp 1-2; k8mdVabt2CFphzJ7 1; KcN4tqGDJQxmgrvZ 1-5;
kJ3WdhMLqv7K2Nyf 1-3; MXZry8VfkWHhvnYc 3, 5-12, 14-19; NjgtYv4VMR26q7h9 1-4; NWvgFp3kfR2bHyYa 1-35; nzxZ4h87yGXJPgfp 1;
p34dnvkg7yDKL6cr 1-3; PmybzCa2j7NXVxng 1-2; Q8Vgbpw4dJrzNx29 1-4; qVnZ6bPMhWRJA397 1-3; RW1BNk6Ybyh3XjLQ 1-21;
tk3BD1bZafMgKxWX 1-4; vBWX7HQn9zpTMGd8 13-22; vyQrDZN7RX82ApnP 1-3; W1P3YqBptbnGM7da 1, 3-5; w4fzHNVk2dAhLKnv 3;
WAyrNRJdm3pqCLMK 1-13; wFBytQxvWKz1Xf2L 1-5; Wj3TyNxChH4nvpLD 1-2; xkN3nXgd8RpHybzT 1-4; yPY3Nhg9MB2TVGZm 1-3;
ZA8pM9gmXcG6avCT 1-3, 5-6, 8; ZLpcV6krn1wYHJvt 1-8.

## Bloodspiller (7753)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 18200
- Rate: 0.95 PPM, measured

385 procs on 9,076 rolling hits, 15 wearers, 17 runs, in the main hand: all hits 0.94 [0.86, 1.03]. The rate is all
hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 0.95 [0.82, 1.11],
116 procs on 2,725 lone units, 14 wearers. Specials: 0.99 [0.93, 1.68], 25 procs on 561 lone units, 5 wearers. Two-roll
abilities, priced as one roll: 0.82, 3 procs on 81 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th
percentile 0 ms, 132 isolated hits. Dropping the swapped stretches (15 active minutes, 4% of hits, in 2 runs) moved the
read from 0.91 to 0.94. The paladin check, which sets no rate: lone white swings 1.22, 23 procs on 420 lone units, 6
wearers. By report date: before 2026-11-18 0.91 (282 procs); from 2026-11-18 to push 114313 1.05 (103 procs). Reports
from 2026-11-14 to 2026-11-19. The sim does not simulate this proc yet (the generator lists it as a missing effect); the
rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      11 |   295 | 6,541 | 1.00 |
| Warrior |       3 |    41 | 1,295 | 0.70 |
| Hunter  |       1 |    49 | 1,240 | 0.88 |

Reports and fights: 1YQTjgZh2KPBGCbN 1-3; 2ywhMnJq4Y3zaWtg 1-9; 2ZYxCWcLwVJar1Ay 1; ARXJvZh6kwt3np2D 1-9;
BNajn8zgV49rZ7Fm 1-18; bqW1xQpFc9ag7NtM 1-2; fAKd8D7wkgQzM9GZ 1-3; Fx34Wc78qhnPmVz6 1-5; GFkKCvMYJwfjHxac 1-6;
kma6pFnDVW1CQ7Tv 1-5; LxdYWR47ZFbTNBhM 1-2; LzFXxt3Y1HyVK4nc 1-7; tCafRNMpHyAbgcmz 1-5; wdrYBPyVFcxMX7Q4 1-2;
wPD796Jv2htngRmT 1; YFnCf6cWNRpgVMHT 17-31; ZMn6BHdqtvpGLhcK 1.

## Bonechill Hammer (14487)

- Proc: weapon proc 18276, logged as 18276; rolls on main hand
- Sim rate: 1.8 PPM, observed; measured 1.79 PPM [1.64, 2.05], 4 wearers, 4 runs, 511 procs on 7155 hits
- Reports from 2026-11-13 to 2026-11-14
- Notes: swap noise: the swap drop moved PPM 1.31 -> 1.78 (48 active minutes, 26% of hits, in 1 runs); few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 9zxHaWbYT3gqDQvF | 17-18  | Rogue   | mh   | timing + otsu for 7 swings  |    93 | 1104 | 2.11 |          |
| rphdM14FbnTBaj3V | 1-27   | Paladin | mh   |                             |    50 |  739 | 1.69 |          |
| vwCXzhMQ2drgpRxD | 1-41   | Paladin | mh   |                             |   311 | 4589 | 1.69 |          |
| zmDY4LVX7N8vh9aR | 1-3    | Rogue   | mh   | timing + otsu for 23 swings |    57 |  723 | 1.97 |          |

## Boneslicer (273819)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1292573
- Rate: 3 PPM, measured

447 procs on 4,438 rolling hits, 25 wearers, 28 runs, in the main hand: all hits 3.02 [2.81, 3.23]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.08 [2.81, 3.49], 238 procs on
2,321 lone units, 25 wearers. Specials: 3.33 [2.74, 3.93], 93 procs on 839 lone units, 22 wearers. Two-roll abilities,
priced as one roll: 4.35, 10 procs on 69 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms,
305 isolated hits. Left out: 48 of 76 runs in Cat or Bear Form. By report date: before 2026-11-18 3.03 (211 procs); from
2026-11-18 to push 114313 3.01 (236 procs). Reports from 2026-11-10 to 2026-11-20. The sim does not simulate this proc
yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |      23 |   320 | 3,149 | 3.05 |
| Warrior |       2 |   127 | 1,289 | 2.96 |

Reports and fights: 3rbcnjTMRfwxv1mh 1-10; 3yMkaTdQx8NJXGKb 1, 3-12; 79RVgZWM8dF2wkCp 1-6, 8; 7zcat1AyFnrJ4BQR 10-19;
8fwQv1dbWXxpkG3r 1; aqCndvgQbBXL3Jf8 9-13; BzGXfjMLJ2RyCwHm 1-7; cG82vnXYjAVaTN1z 1-4; FAf9V76BctbG3mR4 1;
g74R1Wf9FL83pMP2 1; gJkN6DCr8ZfWTn72 10-13; hGNVDdLFRkj7tfcv 1-7; hLNVaj9HdnCM2BYt 1; K9WGdrx248FvaDVY 1-3;
KbNyBzwZg1t9jMcC 1; KrH8VYbgyfaTXWh9 6-8; MW8HLRKDFN4GxTr7 3-4; nKNgy3VrZXkcdjBT 4-5; NrHbC9xfM7KzZt4d 1;
pLh28qFZjAYvrMT6 1-6; qMKJYxCXzcPRZ8nD 1-3; rpB2MKnbm7LxCqJZ 1; tVB2a6fMvYNCGdkH 1-5; TZkhGPt2fqAnNY74 1-3;
vnwtmMTpkR1aYP7G 1-9; y7qfkc91hA3DKQ4g 1-11; Z26ztpb3FkJhgTAf 1-8; z3YfqBgbt1GjFD4W 1-4.

## Bow of Searing Arrows (2825)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 29638
- Rate: 4.9 PPM, measured

2,876 procs on 13,102 rolling hits, 19 wearers, 20 runs, in the ranged slot: all hits 4.88 [4.68, 5.09]. No lone-unit
rate applies, since its travel reaches past the isolation: the rate is all hits. A separate count reads its lone Auto
Shots at 4.98 [4.62, 5.31], which agrees. Specials: none. Travel time: the proc lands a median 553 ms after its hit,
98th percentile 958 ms, which reaches the 1,000 ms isolation, so no window was measured (674 isolated procs). Reports
from 2026-11-13 to 2026-11-19.

| Class  | Wearers | Procs |   Hits |   PPM |
| ------ | ------: | ----: | -----: | ----: |
| Hunter |      18 | 2,875 | 13,101 |  4.88 |
| Rogue  |       1 |     1 |      1 | 22.22 |

Reports and fights: 1XGPdhkb28MAagN4 1-7, 9-10; 4CBckLvG7RaxFzYA 1-4; 6Krp8bkfx4cRjBaz 1, 3-15; 6WG74An9RrCdqkwD 3-8;
7BjkRJr2DZp1g8TL 1-4; AhLc6aGKMCdvBgDW 1-5; ApJhHfRkMDqFj3rb 4-5; G6yaZfLzMtN13hcA 1; H2GvZtVYNy4QzqLw 12-21;
KGQHdbFCZpt1YX23 18, 20-23; KtDJH283f4bk7R9P 1-4; Pf4Y2xdhVz1vq9rn 1-12; XqBgvYQZFw4H2czt 1-2; YWGNyaDcLwT1QR6J 1-8;
cJhVzAdLx3BHqpkT 1-4; fHbJKd4LnVXDW2gj 1-2; krpvb6WaLmFnV1KG 1-2; mtcXBg7ZkVxFHQnh 1-4; qHgyawJNVb7126xL 1;
tk86znZXxMCdwBKb 2-18.

## Brush Cutter (274265)

- Proc: weapon proc 1293518, logged as 1293518; rolls on main hand, off hand
- Sim rate: 2.35 PPM, observed; measured 2.35 PPM [2.11, 2.61], 3 wearers, 3 runs, 301 procs on 2746 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: swap noise: the swap drop moved PPM 1.81 -> 2.26 (31 active minutes, 20% of hits, in 1 runs); few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Hand split                           | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ------------------------------------ | ----: | ---: | ---: | -------- |
| cr6qMtj1vCTPHVfw | 1-7    | Warrior | oh   | timing + otsu for 84 swings          |    48 |  409 | 2.52 |          |
| pjRYqGgB2rdvawNn | 1-10   | Warrior | mh   |                                      |   199 | 1851 | 2.30 |          |
| qcdyZAr2ganDtxNW | 13-42  | Rogue   | oh   | timing + speed ratio for 2032 swings |    54 |  486 | 2.38 |          |

## Burning War Axe (2299)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 18199
- Rate: 0.95 PPM, measured

627 procs on 11,799 rolling hits, 25 wearers, 25 runs, in the main hand: all hits 0.97 [0.91, 1.03]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.06 [0.90, 1.25], 168 procs on 2,881 lone units, 23
wearers. Specials: 1.27 [0.89, 1.46], 25 procs on 359 lone units, 5 wearers. Two-roll abilities, priced as one roll:
2.07, 15 procs on 132 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 159 isolated hits.
Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      20 |   535 | 10,030 | 0.97 |
| Warrior |       4 |    92 |  1,763 | 0.95 |
| Hunter  |       1 |     0 |      6 | 0.00 |

Reports and fights: 3cKTmVyb1ZPwFXkf 1, 3-13; 93NcnZPAhCfQ4wHz 1-4; BqjTQP3Ghbt9JyYn 1-11; FMcvKqTzn7m8Va3P 1-15;
Fgy2bt81NL9CAzaR 1-13, 15; GJHgk6MDpxfVdacv 1-23, 25-29, 32-33; KRCBMJvg72fHVwWL 1-17; LwTzGJXrN3HWqCfD 1-5;
MBaXr1H9GcPhDJLm 1-2; PZA4xq2WvQJ3amjH 1-5; Pc1RkCWp8wTNfFr2 3-4; QvKPgyHDkBMbcp7T 1-10, 12-21; RBmtAMLxz78Pj2Wh 7-10;
VQFaRpT6J2xLDbZg 1, 3-5; Wf8ghC6TtQKrkXdv 8-10, 12-15; bkMdQTwZmqNPvhrR 1-5; gFZ7BcKt9JhaY2dQ 1, 3-9; gGtkavy2n8bqADQ9
1-2; gJkrfBtbp7YM3PCd 6, 8-21; h2FAdwQfcJLHpzR3 1-16; j68nZp1gcrGChaRW 1-12; qMD3CVapwb6nAgx2 16-40, 42-43;
rMyzaJhG4k8wHPj6 1-10; ryqgQphLxMBz4TmH 1-4; t6HbjRZmkyx3pQwK 1-14.

## Claw of Celebras (17738)

- Proc: weapon proc 21952, logged as 21952; rolls on off hand
- Sim rate: 2.45 PPM, observed; measured 2.44 PPM [1.74, 2.80], 5 wearers, 6 runs, 213 procs on 2915 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split                          | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----------------------------------- | ----: | ---: | ---: | -------- |
| 2taZz4vTdfXYAk3b | 3-9      | Rogue   | oh   | timing + otsu for 4 swings          |     5 |  127 | 1.31 |          |
| 4Qm6MgZWBTx78Lvp | 1-2, 5-6 | Warrior | oh   | timing + otsu for 102 swings        |    64 |  677 | 3.15 |          |
| 4XGyazkdvmP98r2L | 1-7      | Rogue   | oh   | timing + otsu for 45 swings         |     9 |  181 | 1.66 |          |
| 61Hx8DgwhdbWaN9C | 1-12     | Warrior | oh   | timing + otsu for 79 swings         |    50 |  611 | 2.73 |          |
| 6yrCJtHYmaRnZBM3 | 1-5      | Rogue   | oh   | timing + otsu for 5 swings          |    30 |  492 | 2.03 |          |
| 7AxaTpFZMqGb9wNd | 1-8      | Warrior | oh   | timing + speed ratio for 251 swings |    55 |  827 | 2.22 |          |

## Claw of the Shadowmancer (2912)

- Proc: weapon proc 16409, logged as 16409; rolls on main hand, off hand
- Sim rate: 1 PPM, observed; measured 1.00 PPM [0.85, 1.15], 10 wearers, 10 runs, 195 procs on 6163 hits
- Reports from 2026-11-07 to 2026-11-16

| Report           | Fights | Class  | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 1jgtdNmDMRycbZrP | 1      | Rogue  | oh   | timing + otsu for 15 swings        |    13 |  287 | 1.43 |          |
| 9LgNW1jQ2Zh8XDkY | 1-12   | Rogue  | mh   | timing + otsu for 23 swings        |    51 | 1333 | 1.21 |          |
| 9vwmqxP4dLKhDC8Z | 1-23   | Rogue  | oh   | timing + otsu for 22 swings        |    11 |  453 | 0.77 |          |
| AHdpxG72rk6CzbMa | 1, 3   | Rogue  | oh   | timing + otsu for 6 swings         |     6 |  288 | 0.66 |          |
| FYmnawMf8gjB4NQt | 1-3    | Rogue  | mh   | timing + otsu for 1 swings         |    25 | 1092 | 0.72 |          |
| JK4pARZQzyLb2DMh | 1-3    | Rogue  | mh   | timing + otsu for 27 swings        |    29 |  884 | 1.04 |          |
| L8yaqgDXG26Nf9v1 | 1-2    | Rogue  | oh   | timing + otsu for 2 swings         |    20 |  518 | 1.22 |          |
| LbmB2QdWnPT7JGAF | 6      | Hunter | oh   | timing + speed ratio for 19 swings |     1 |   61 | 0.52 |          |
| Tw4BdRP1HmZqJL9F | 1-3    | Rogue  | mh   | timing + otsu for 10 swings        |    26 |  865 | 0.95 |          |
| hcGCaDW9tRYFB6Ax | 12-18  | Rogue  | oh   | timing + otsu for 9 swings         |    13 |  381 | 1.08 |          |

## Cobalt Crusher (7730)

- Proc: weapon proc 18204, logged as 18204; rolls on main hand
- Sim rate: 2.25 PPM, observed; measured 2.24 PPM [1.88, 2.66], 2 wearers, 2 runs, 112 procs on 938 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 4 of 6 runs are feral druids, left out (paw speed, not the weapon); few wearers (2): the interval is about the per-wearer range

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 8frHF64GzAhBxDyb | 4-14, 16-22 | Paladin | mh   |            |    51 |  412 | 2.32 |          |
| g8b1MkJnKtyfzvPa | 4-5         | Shaman  | mh   |            |    61 |  526 | 2.17 |          |
| LN78yvQwgJpcAYCG | 1-2, 4-6    | Druid   | mh   |            |     4 |  436 | 0.17 | feral    |
| M6kJfn7a3r2KFRgd | 1-4         | Druid   | mh   |            |    21 |  621 | 0.63 | feral    |
| XzKCqkBbmY3W2Tc8 | 6-7         | Druid   | mh   |            |     9 |  716 | 0.24 | feral    |
| cvfBaDxQX7K2zFtk | 1-7         | Druid   | mh   |            |    42 | 1732 | 0.45 | feral    |

## Coldrage Dagger (10761)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1293790
- Rate: 2.95 PPM, measured

5,110 procs on 69,251 rolling hits, 58 wearers, 64 runs, in the main hand or the off hand: all hits 2.95 [2.85, 3.04].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.62 [1.95, 3.58],
62 procs on 945 lone units, 2 wearers. Dual-wield white: 2.63 [2.38, 2.85], 900 procs on 13,691 lone units, 54 wearers.
Specials: 2.87 [2.38, 3.34], 94 procs on 1,308 lone units, 18 wearers. Travel time: the proc lands a median 100 ms after
its hit, 98th percentile 257 ms, window [41, 267] ms (70 isolated hits). Reports from 2026-11-14 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      47 | 4,592 | 62,481 | 2.94 |
| Warrior |       5 |   465 |  5,970 | 3.12 |
| Hunter  |       6 |    53 |    799 | 2.65 |

Reports and fights: 1RVnhHGv2bwDW9Yr 1-2; 1jgTLvJQBMP6kxrm 1-12; 1qcv4MKB7R9bWxJZ 1-9; 1w3vfpYKqbrBNZPQ 1-2;
36h91AnYPjTkbxcM 1-12; 3PAwV6XFaB7hqDgz 1-12; 3THgCcxGJ4XYd6Kp 3; 3ZnAfaWCBgJhxbVY 8; 3bvPyGC6HmRNnVjz 1-11;
4xnDCWwrGKAFdNHv 1; 6734nzHBJ8yGVjFb 1-15; 6DJfd2yhL8ZNpkXv 1-4; 7YKFdxrDHLGcybjT 1-4; 7dZrzYqjKHh28mvW 1-2;
987q1pjnYdVQbCmT 1-6; 9QnWbJphKLFVgZ6z 3-8; ALhN4BDjHGv7P82V 1-7; ANWTMG9KQ3mXYh72 2; AfwXCyb2WRDMQmzt 1;
BPf7rVXwgjLF4yR9 1-2; FCXJL6RHZjzQY8d7 1-5; GCTbWRf9mrHqxAX6 1-3, 5; GyQ6cqtCxPXHaDFz 22; H9FXjaBWRpYVMq4r 1;
KF87fDdVjrhgbMpB 1-6; LYPNJyTX4rHv6cxk 1-11; Lr6VTwchNAxWqzvK 1-5; NZXhmPJaAzCDBb4d 1-25; NfqRryF7gnMcX2ak 1-2;
QKg4z1amMYyBCp69 1-3, 5-30; QYvzMfCm3pkRXPr1 1-10, 12-13; QyqcgFJZA6mLH7wb 1; R1jPGYZyTtBkKr8J 1-4; R9njq1AGDvmp36bM 9;
RW1BNk6Ybyh3XjLQ 1-21; Rfwtj4ZF1nTx2mJp 1; RvcrA3P9jznxhHVQ 1-2, 4-7; TADvnpKkdwgZGJ6b 1; VQ36t9NzpFjnmvka 1-4;
W3mNJr96cDpqYQLy 1-2; XgB2qrQDx34ZMjH9 1-3; YCwpT4hDKgZB2JFA 1; YXmAnhFPdarR48qw 1; ZHxr2qha3V4QBvLN 1-12;
acxJ2gRWPwtpDCZh 1; d6CYVkHJhGXjByTP 1, 3-11; dn62GTvrBYFQbWfV 7-20; fX6JWBtcRQrmx7gD 1-5; g41xhtdFCXZ8Vr9z 1-2;
gcnGw3DaJdrFCAYX 1-5; jVhAWnaHJgf3FPmy 1-10, 12-13; kXG9NmQ7rJKbj1vM 1-7; mD3vxz7paH4jKgTF 1-5; mNzF1MgYJypDfnPb 1-14;
nYMZcpQ8WhRTt97q 1-4; pKn8wBZcQ62PyJH7 8-24; pqfK2AgHZ7DnhvPF 1; rcCT7JVyW893mPQk 1-2; rtbzWXcByFwk7G8n 1-17;
tyFZPdH39kGq21An 3-25; w6J3Ykh4cWzq7j8L 1; x3bWXZa1KDCmFvtn 1; yjqVPg81hM6TCW97 1-2; zYBvPW8dqghTAaZp 15-22, 24, 26-30.

## Cookie's Tenderizer (5197)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1292008
- Rate: 1.95 PPM, measured

1,464 procs on 16,079 rolling hits, 36 wearers, 37 runs, in the main hand or the off hand: all hits 1.95 [1.81, 2.08].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.98 [1.70, 2.20],
277 procs on 3,004 lone units, 18 wearers. Dual-wield white: 1.69 [1.44, 1.88], 322 procs on 4,080 lone units, 17
wearers. Specials: 1.93 [1.76, 2.11], 288 procs on 3,195 lone units, 20 wearers. Direct: window [-10, 100] ms, median 0
ms, 98th percentile 0 ms, 635 isolated hits. Dropping the swapped stretches (189 active minutes, 21% of hits, in 8 runs)
moved the read from 1.55 to 1.95. Left out: 5 of 56 runs in Cat or Bear Form. The paladin check, which sets no rate:
lone white swings 2.00, 63 procs on 676 lone units, 17 wearers. By report date: before 2026-11-18 1.87 (983 procs); from
2026-11-18 to push 114313 2.14 (481 procs). Reports from 2026-11-07 to 2026-11-20. The sim does not simulate this proc
yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |      16 |   789 | 8,838 | 1.91 |
| Warrior |      11 |   527 | 5,681 | 1.99 |
| Shaman  |       7 |   148 | 1,557 | 2.04 |
| Priest  |       2 |     0 |     3 | 0.00 |

Reports and fights: 2gjR3tVq4Zdw7kzK 1-3; 2htpkPNX3yZ8K74C 1-3; 2PWL9DB1k47p8MRd 1; 8tvAkdpfn2hNc9Qa 4; 91hVt6Tpm3kgvcyB
1-13; 9gx3fpD7PA1RhCYH 1-2; adj7kHPW3ypMKrNL 1-3; aRDTH6pqdyGgPkQ1 3, 5-6; B2D9zYahGJLKbVHx 1-4; cJm7HAtbgG2hqyXz 1;
D4T2RPhgAMvFYk1c 1-6; DMZPzKB91HJAqLQy 1-4; f297FtyjGHmRVqYw 1-4, 6-9; FWHahtQvqjL3wNdV 1-3; FZB3hwQLpjHn6829 1-12,
14-19; GDWLqtcJgbrz1P89 1-8; k1tqMpa6hdByfCAQ 1; kFMtgzB1cCvxJTfL 1-9; KWFdXvCfR6rVYLcj 1, 3-12; LdcVztBw6pFjNahM 1-3;
MaXHVZ2nWDJ8z6Cd 1-3; MLkQWdC6hNcm34gz 1-3; Mv73gTFaNhbGWLy9 1-6; MxFBA2RChVbm4Ykj 1; nTY7GDAHgZf8cbzr 1-4;
QmjBxrzvah2XTVR1 7-8; qRQfj9HCFYp2MrDT 1-2; RpkqFDaLQn6bxwt8 1-8; tBnv1kDbzHK6hAwg 3-7; TFnvh6k2P8zL4DCQ 1-16, 18;
tHfZywmLT6XVpqvb 1-3; w1WaVmyY8jBT3GF4 1-8; yCTRh3Hq7GPt6nAX 1; YfWXxBnqKv7HgyQ1 1-2; ZkTHQfR8F13AdaMq 1-3;
ZTJqNPfRLb3c8Bav 1-10; Zv6fc3GgWpkX9Bqa 3-4.

## Cursed Felblade (14145)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 18381
- Rate: 2.05 PPM, measured

7,913 procs on 99,778 rolling hits, 112 wearers, 127 runs, in the main hand: all hits 2.07 [2.02, 2.12]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.06 [1.97, 2.14], 2,518 procs on
31,918 lone units, 70 wearers. Dual-wield white: 1.91 [1.79, 2.03], 992 procs on 13,579 lone units, 41 wearers.
Specials: 2.06 [1.97, 2.15], 1,860 procs on 23,579 lone units, 109 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 4 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 3,915 isolated hits.
Reports from 2026-11-05 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      68 | 5,206 | 66,167 | 2.05 |
| Rogue   |      35 | 2,656 | 33,044 | 2.10 |
| Hunter  |       8 |    51 |    567 | 2.35 |
| Mage    |       1 |     0 |      1 | 0.00 |

Reports and fights: 12VhLkPa98DYFH4g 1-3; 1RrDAhp7ZvVYc9P3 1-6; 1cf7ba34WkwGyzQY 1-12, 14-16; 3CyQk69LzJjTn2X4 1-5;
3rDXa6gPxYNZJjFH 2; 6P4VND3w2r1LGQab 1-2; 6jfqGYcR9AWdJvFh 1-3; 6mrVNdqBkwgvGxW3 4-5; 6tANqvfKQP1Fkdzg 1-3;
7M48dqLp2yYHaA6Z 1; 7NqcJQvyKpCd3YAt 4-7; 7RyBGr4dYVFqJHDb 4; 7Yt1bzjCBg9JycdT 1-3; 7bch4MKwzf9HdT3v 1-3;
7qLWf2j9a8tvCNwH 1-7; 8NTz6vHAcj1W2ZmQ 1-7; 8Xxb9LDGHP7vcJfh 1-5; 8aymDGwFnWZAx7YR 1-3; 8tmqwrpyARjY6hnc 23-34;
9NbWjdLhagPZn1X7 1-2; 9aCJ7GH6QRFjgqb4 1-11; 9mKt61MCAyfP3GRN 1-4; ABX72zJxvr6YVpk1 4-5; APNwHKda8WZQRt1G 1-2;
AT1kyCR8NhMpcbjm 1-3; AdNRV84rQFWcxk6q 1-2; AtH8ZJm36y74V1d2 1-6; BwtcDCWfMLdTYFqN 1-7; CyJ13pMra2W6vdQH 1-2;
D8qzgc9h1YLf6krp 1-6; DaNQ6WPf7GqTpKVk 1-8; DmFcCT4QpH3AbhPR 1-2; G6M9XwZ1gQCTFbPa 1-2; GbrjBnmfgvphqcPZ 1-11;
GgTjx7XKkMvtwnQ3 1-5; Gv74gVkXcyPHaYzj 1-12; GzbK6XYDf1vQH4W9 1; HwqhFkWZCBRTQVDf 5-7, 9-13; HxVGdyMjDB6JCPKQ 1-2, 4-6;
KAvPryL1zCdtWTxV 2, 4; KYRtHcN6L7nmq3Cz 1-2; KgkThB9FPJv3MrZw 5-10; KpD3hj9wBk1bNL7d 1-5; KpTQqRdvf71n2Wm4 12-21;
LcaMApJGDWjg4Tbf 1-10; LwYXzRFCP8GvVKfA 6; MLTmgKZDkJaAcwYR 1-3; Mt8kPGravCBTjdA7 1, 3; MzwcDrG3fdFBRy6a 1-5;
N4tMyW9qJcBKhvY8 1-3; NZrPmMxy3LXtj9G8 1-3; Nkndb4wD6WgAK2Lh 3; PW7k29jyJ3C4chqQ 5-7; PtzakwgYVNR783An 1-2;
QDGMAJr2K3Lb91kX 1-7; Qmw4FNtdvqnkPcKy 1; RF61yp4dkXNCZTwh 1-2; T7HKGnYfbMLwWCJp 1-7; TMj2XLPvzgq3ZNJ9 1-5;
TbPzqABaXyWmLxD9 1-8; TdXmWDR3QFJwpYfq 1; VKxkQv9dyNAgpjPh 1-3; VXcPMWHwZJftbCrm 1-9; VYHAwTB2pP1r7mgJ 1-5;
W416FHmftD9rPRgA 1-2; Wk1G3TqNcZzPvRHD 1-8; XBQMrcVdkvqGCJF1 1-8, 10-11; XKc78ka3dWNgGRDr 1, 3; Y6p9C7nNx1KBWGdM 1-13,
15-33; YaBQZ2tbVFJMrAhk 1-4; YbxV4LgjKt2BfTCp 1, 3-4; ZrM8cydqRt1BpLfx 4; ZwHMphjVvdtxTY9C 1-4; a9fR1QjqFvnhgxW7 1-5;
aDRLCWyFkdqVc2Z6 1-2; aVyXCj1NpnRTFKxt 1-2; ad24TYKCAb8LwBxc 1-4; agA7pcDMCb9qJFKv 1; bLd41TDnmVYJZj2x 3;
bXYpRJw87M62gKar 1-4; cAdCm1NLGaQnXrHW 1-4; dwq8K9CRNBAQtLjp 1; dzTmC6F7xwLPJrq9 1-17; fTpYLw71QVR2qky8 2-13;
fkGrDPXmVWazJxpH 1-7; fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; hYDQPX1rjn7wZLxz 1-3; jGkWVBDa7FPmpwn4 2-9;
jKfkzadV3BLp1cbr 3; jRytczTZF6pBJgYL 1-2; k9wzA3ZbgnX7j6x2 1-4; kB7JaQmdX8RZjcnT 1, 3; kHnFMcBXRrx9VajJ 1-3;
kcwp7a2jqCF91ZWK 1-4; kgFaYjKwPH1CAx2m 1-2; mTPgn7Kf2XB6ykpt 1-10; mzTZRY4Hg9aGdvxP 1-13; nPZbN3Ymk1LVtDQK 1-2;
nZ1gTJbNYPaLkx92 1; nfCzby7RrmYFxTXc 1-3; nkVby8fgJ1PFTBLj 1-2; pL3Fbdfc2vNKCxtJ 1-4; ph6FqnCZaRQLvxbM 1-2;
ptWXb1v93rkZR462 1-4, 6-9; q6fAZYgapznjxPK9 4; qcyGBRWh2xVZgKAD 1-2; qhFK6CrAVTH932kt 9, 11-19; rzchPRZJy4qNp3mW 1-3;
txhaZrbjXR7Kf2TD 1, 3; v2y9BTtzLhYr1XHg 4; v9fnrCd8TLaPDYH4 1-2; vBCYwTV1D9gxHPMQ 1-9; vba4w71K9dVTXZ3r 1-2;
vrZD4X7Vb3pW9FNK 1, 3-4; vxkFd3MfgJX7NyAc 1, 3-8; xCQDhmMVT4qdRWYb 1-13; xPmgn423azWr8F9L 9-11; zCjDnVmyRZPg2J7Q 1-3;
zLK3mTNayZMgR6vh 1-7; zPJ2w1BrKA3Dv968 1-2; zRd3nAXrVQqg9ZH6 1, 3-5; zbV7pwYGARa4Tqgk 1-2; zhj69Dt2X1b3fvMT 1;
zk682jCMJ4AdKfV9 1-2.

## Darkwater Talwar (11121)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 16408
- Rate: 2.95 PPM, measured

12,936 procs on 119,119 rolling hits, 138 wearers, 154 runs, in the main hand: all hits 2.96 [2.90, 3.03]. No lone-unit
rate applies, since its travel window passes 400 ms: the rate is all hits. One-hand white: 3.09 [2.83, 3.32], 684 procs
on 6,046 lone units, 50 wearers. Dual-wield white: 2.70 [2.43, 2.95], 549 procs on 5,550 lone units, 72 wearers.
Specials: 2.92 [2.71, 3.15], 582 procs on 5,429 lone units, 118 wearers. Two-roll abilities, priced as one roll: 4.98,
19 procs on 104 lone units. Travel time: the proc lands a median 146 ms after its hit, 98th percentile 601 ms, window
[39, 611] ms (701 isolated hits). Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      66 | 7,790 | 71,391 | 2.98 |
| Rogue   |      47 | 4,838 | 45,007 | 2.93 |
| Hunter  |      24 |   308 |  2,721 | 3.09 |
| Warlock |       1 |     0 |      0 | 0.00 |

Reports and fights: 1CYTfP3pFDcaRrJm 1-8; 1h9vznDdxyjC7KWZ 1; 1vJgfm9dGnkDzwXb 1-2; 1w4hbgxXHpf3N6M9 2-3;
24KTy3bLGNQm1rdx 1-20; 24TVGCfvHpjtYzrD 1-9; 28GbNhyfwQ1DcRaP 8; 2GYNtHzfKmWpTPVZ 1-14; 2Z3YxGdfPhMK97kQ 1-4;
2pGYFgHwLDC17zZ9 1-4; 4CRDaYFWAqt1JnmV 2; 4LrMDtG6T8AVRJzQ 1-11, 14-28; 4ktfmxnqcbv61B2L 1-18; 4wJgRGx3rCKjQZmD 1-12;
6FZtkRcQHAKxNfrC 1-19; 6X1L4nB7zCfqk8vV 1-2; 6Zzwx92Wh4qFVLBp 1-8; 6prtaJW4g7ZTVQRK 1-6; 6wXA2CY3Bdcp7gDj 1-4;
6xhBGvgcM9H48tyK 1-8; 7BKLx2tGCwqQHr63 8-13; 7J9jfrb2LaHxgv6w 1-13; 7NHAcVG9nZCyBMjv 1-6; 7xY1zPCyKnk8DGQj 1, 3;
84c12nJmjaKCFr7H 1-3; 8BG41brwTQkLRjqJ 1; 8Bgm9rLTF67QvNwn 8-11; 8tmqwrpyARjY6hnc 1-10, 12-21, 23-34; 9GMN2DQLhrB64Atv
1-2; 9yNQP3FWHgcVvYDq 1-14; A1WnB2b3jvaJKyfp 1-3; AL1DxTaHGfVygm3p 1-15; ARwG2htz3VgjDd7v 1, 3-5; Ahv9342TJmG7bCzr
13-15, 17; B3vtVgmqMTHDAxNY 1-10; BmVrHWZvtj9GJTxD 1-13; C8J2PmQbz9dwcKXH 1-4; C8QWj29pvm46kMTF 1-3; D3Kbtnm2yaC4Tqwc
1-2; DK72jqp46VtnYHfZ 1; Dtph4q9QHVGPXyva 1-7; FR9jJ7b31XCkLmtQ 2, 4-7; Fw6bq2BvXnJmaRk9 1-3; GDWLqtcJgbrz1P89 1-8;
GntwzVb68vMyWjBd 1-2; GxWL9RvzyJC76d13 1-2; H1MQjnkcT6GraYbh 1-3; H3bRaDmzyjxTW2BJ 1; H3vdcCr6JX1gwYtq 1-3;
HG4YQy7Jf3xnFpLM 1, 3-9; HPxv4aJ1WgcTzMCd 1-13; HhFaG7mzPvZWpTtK 1-2; JzRmL9FYywDQgNcW 1-2; KqZaH876NnQV2CJw 1, 3-21,
23; Kz8FhGRY6kVTBjwb 5; KzctnTrXN64fmWPv 1-2; LHD7ZAYVNm2zjdXP 1-17; LHRhtADNW8GfyYqT 1; LK3jvm28bhWQadyr 1-4;
LhMNaKbHFvTt1Ax7 1; LzFXxt3Y1HyVK4nc 1-7; NDf1mPJRCaTM8WkK 2; NfLb7C1qDZ8cHxFp 19-39; NfjK7rVnPvkCB3XD 4-9;
NxmwqXVA7JkpP8Qn 1-4; PrNcBn64hjYw7KQ2 2-8; Q73HMJW9AkRPj4aF 1-10; QGLzd2PDcm763VRr 2-3; QRnGzYBNkWyCD96g 1-7;
QVg7ANTY4B8RhDJM 2-9; QgR6kbHvY1CBhXFj 1-2; R2zHVBktcw8j6rhM 1-3, 5-7; R7Fw1Kdcab3VYTx6 1, 3-9; RgjLQAhp3HTCwkca 1-15;
T2zBbgrnJdK7QNPc 1-8; T87bpgKckHzNDAfY 1, 3-6; TQ4PXCqWA3KGYzh2 1-5; V64yC1A3DZdWfGBP 7-25; VWF2YryN8j4qdK9H 1-51;
VaFNJB32Ah1WGLrn 1-11; VdwjgLC7qNzvPFpT 1-18; VqybFQPLtxDMnv2h 1-2; Vr12d8kxGHpKazBX 1, 4; XAyLYahcRwMWQfF7 1-4;
XWnJCQ1Mc38fRtwF 1-3; XqhkLtanYFGgx9jP 1-14; Y6wnvNGfpPVrR7XJ 3; YC214LNaFPwkRrfm 10-12; Z19fCFxJnyLc6dPN 1-20;
ZjgDb23AFJQkvpWL 1-3; Zvbh2XrFYqQPzDnJ 1-5, 7-20, 22-27; aKwXBJkG2bW6tVNY 1-5; aTHRgr6nQqVF7DC9 1-9, 11-20, 23-25;
azpj2hg1DVwnMCkZ 19-21; bH6GZhykAPqgQ3mt 1-4; c7gHWKFrZMnm2yb4 1; cfDYLdNC1t6kHQg7 1-5; dkAMWqBczZtbHTaV 1-2;
drvpchNFDMPXB8t2 1-10; f297FtyjGHmRVqYw 6-9; fLa4hw8d6pNFYMng 1; fXk6pdvPNF9wZCjc 4-5; g3nkvH9atfKRdVAh 1-11;
gAmJdBx7QyTcR3YK 1-17; ghCVqbxDmnJH6P9w 1-18; hfJZb8HL41KmtdMA 1-4; hvrBLpKmakJ8nG3W 1-19; j4HkVFBWdQvMR3cn 1;
jFDWcdYtPzkRhqB8 1, 3-12; kFxDzj7pGtTqCZg8 1-3; kHfNcJTYj3RtMFGL 1-3; khV3XMmJ4aRKzxBt 1-3; kp4FHxfa9CL7WArB 15-30;
mKHgTzDZXkBbxarG 1-12; mRxb9L2wNkMJTWXY 1-7; mj379AhvbKZVpWHC 1-7; mkGcRA9M37QyJfL6 34-55; n8ZQCDKLjHP9rdkR 1-2;
nKCDhXNbHpJaBcAR 1-15; nmFLdPWf8agkC7y2 1-3; pHZ86RKXPNY4nF1V 1-4; pKRWfNGZg96VBQTx 1-20; pwNAZn643XgzcbYa 1-3, 5-6,
8-10; qavTP8NJtW4hcQLH 5; qhGZDM3ygKAHtjbN 1-12; qjryBx72WJPmA4hT 22; qpXkvRdWxKzfNr2G 1; rTzVg1BxmZh6pFKC 1-11;
ryL9Dfkm6jtzFXgG 1-17; tD3pKRxTyNrqGchC 1-14, 16-18; tbT31mz6ahFMfnCW 1-2; tdxJaDzc8Qv2jFbq 1; tjN8BgxX7k4GyAfd 1-4;
tvrgBcqMzCRk1wF8 1-2; v12HBpQDLWa4XZ9t 1-4; vk2wGPDbQp8TZBR4 1; vtHaBZF2VAqJ19xY 15-16; wFn3hXfqxRHpQcbB 1-4;
wQAj94dkrc3MvXpD 1-6; wYmr1kfWPpVz7jyZ 6-13; wqCht6J1RkcKfVpy 1-4; wxdb7vaQp4CATFyj 1-7, 9; x1YCtNkmg2XLRbGr 1;
xQtGrv7cg6V8Z4BT 19-21; yX2kx6TbwHn4mcBq 1-5; yqntQdrPj7pvw8Fm 1-3; zcK9qfp32VnQWTMP 1-2; zhZcTYBNvwr48gP6 1-6;
zq6DmwtPJkNZy9KL 1-9.

## Darrowspike (13984)

- Proc: weapon proc 18276, logged as 18276; rolls on main hand
- Sim rate: 2.2 PPM, observed; measured 2.18 PPM [1.94, 2.44], 1 wearers, 1 runs, 282 procs on 5182 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: one wearer: no wearer interval

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 7fy93LdcwQhvF2gP | 1-7    | Warrior | mh   |            |   282 | 5182 | 2.18 |          |

## Dazzling Longsword (869)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 13752
- Rate: 2.7 PPM, measured

4,541 procs on 38,572 rolling hits, 41 wearers, 41 runs, in the main hand: all hits 2.72 [2.58, 2.85]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.49 [2.06, 2.99], 154 procs on
1,426 lone units, 6 wearers. Dual-wield white: 2.23 [1.98, 2.46], 582 procs on 6,029 lone units, 35 wearers. Specials:
2.83 [2.62, 3.03], 726 procs on 5,914 lone units, 40 wearers. Two-roll abilities, priced as one roll: 4.21, 27 procs on
148 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,030 isolated hits. Reports from
2026-11-09 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      28 | 3,280 | 26,778 | 2.83 |
| Warrior |      12 | 1,131 | 10,635 | 2.45 |
| Hunter  |       1 |   130 |  1,158 | 2.59 |

Reports and fights: 1FV3WyfXGCcPrT9b 1, 3-8; 4C28YPXqgTMQJxB7 5-9; 4qZXFGfc1wCMQKL7 1-8; 8F4MhkmNBxRrf6Zw 55-68, 71-83,
85-89, 92, 94-105, 108-123, 125-141, 143-151; 8X2Kc46faTnxVBLm 1-3; 9cajk6Jqd3NTQngf 6-8; A1FKkYpPxrX8tbNv 1-2, 4-5;
BCZgcRxrh7dA4ayX 1-4; BgKt3Qm2fJDAWLzh 1-23; CrR9c6VtXG4bzJhx 1-17; D4A3QNCrX19W7cqJ 1-10; FR9jJ7b31XCkLmtQ 5-7;
FzRmWck7KgT9yt6D 1; GzZH1WhTLQ3JaDXt 1-11; JTztbhZyaWHC6kxf 1-5; KVdX1Wn4GpzAH9MN 11-12; Lg82ZFDyQ4MvRm9p 1-2, 5-11,
13-26; M4CvXLZz3qm6pDRc 1-3, 5; MVcaxDQwWLpnK6HJ 7-9, 11-12; PKRTfgw2Q7ZBxt4r 1-26; Q7BxMTcG1ZbgX2v8 1-2;
QftYMR2haLcXZ7Nz 1; T71XnfaF3cY9M2Rg 1-13; THqFxLkCP6ayd2tZ 5, 7-9, 11; Thf9agqHJyRmtdpD 1-5; VLmAcpbQxWf1CtaR 1-17;
XCJdxAF17zqHMhLY 1-2; ZtndFBMraKvcLfw9 7-9; dmKgVGCJMQpq3t17 1-14; dzXT7JngpxRPwG32 1-17; fCrvDxhZXqJAg62m 4-10;
fkVmJwWZjMy3LBX4 1-3; gTPNZcJ9mG4dRC2X 1-4; h6wkBb9Z7jFtTrnp 1; nF7RcakdVTQD2tvH 1-3; nLk2TKDJWP6vBgam 1;
pKj3bDrf1h2qtNMF 9-15; qW3KvYDpw6ARxZbc 1-7; tkQKL9bGT6JVR3na 14-16; xM9bRGyajmBFvCqt 1; xqy2GfPj67XCwrT3 3-4.

## Deadhead Blade (274271)

- Proc: weapon proc 1293518, logged as 1293518; rolls on main hand, off hand
- Sim rate: 2 PPM, observed; measured 2.00 PPM [1.63, 2.70], 4 wearers, 5 runs, 238 procs on 3970 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Hand split                          | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ----------------------------------- | ----: | ---: | ---: | -------- |
| QRCgnhck3KVq6vBT | 1-2    | Hunter | oh   | timing + otsu for 142 swings        |    28 |  515 | 1.81 |          |
| VLz78BfZyWa3Cxr2 | 1-9    | Hunter | oh   | timing + speed ratio for 199 swings |    11 |  299 | 1.23 |          |
| nX1qbatTzGZ3CAwB | 1, 3-7 | Hunter | oh   | timing + speed ratio for 138 swings |    11 |  208 | 1.76 |          |
| pQ48t7ArW9aDdwNn | 1-4    | Rogue  | mh   | timing + otsu for 8 swings          |   147 | 2520 | 1.94 |          |
| zp9yVrtfTNZJCv7g | 9-10   | Rogue  | oh   | timing + otsu for 41 swings         |    41 |  427 | 3.20 |          |

## Dwarven Hand Cannon (2099)

- Proc: weapon proc 29639, logged as 29639; rolls on ranged
- Sim rate: 4.3 PPM, measured; measured 4.31 PPM [4.03, 4.60], 5 wearers, 5 runs, 702 procs on 3372 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class  | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ------ | ---------- | ----: | ---: | ---: | -------- |
| BfVjHYvWdKkP8Jz7 | 1-2    | Hunter | ranged |            |   108 |  518 | 4.31 |          |
| F3BTZGrW48Mp6n2L | 3-39   | Hunter | ranged |            |   157 |  809 | 4.02 |          |
| KPqgbZxhn6BXwCvJ | 1      | Hunter | ranged |            |   100 |  530 | 3.90 |          |
| T6VDdFhPxQ43pHM2 | 1-2    | Hunter | ranged |            |   164 |  722 | 4.70 |          |
| hH4YRGVFz3KnN6bw | 1-8    | Hunter | ranged |            |   173 |  793 | 4.51 |          |

## Eater of the Dead (10805)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1299869
- Rate: 1.75 PPM, measured

1,865 procs on 29,229 rolling hits, 26 wearers, 30 runs, in the main hand: all hits 1.74 [1.64, 1.83]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.86 [1.49, 2.13], 348 procs on
5,097 lone units, 10 wearers. Dual-wield white: 1.63 [1.46, 1.81], 140 procs on 2,337 lone units, 14 wearers. Specials:
1.75 [1.51, 1.97], 206 procs on 3,210 lone units, 23 wearers. Two-roll abilities, priced as one roll: 2.84, 27 procs on
259 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 500 isolated hits. By report date:
before 2026-11-18 1.73 (1,084 procs); from 2026-11-18 to push 114313 1.76 (781 procs). Reports from 2026-11-13 to
2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      12 | 1,547 | 24,357 | 1.73 |
| Hunter  |      11 |   121 |  1,805 | 1.83 |
| Rogue   |       3 |   197 |  3,068 | 1.75 |

Reports and fights: 1PcrGTyatDAFx8Yk 1-2; 23CWz6vwhH9YdxN7 1-2; 2tCNaTJf49KWyVLk 16; 3f7wRc8vmxMHFa9n 1-3;
3hpaqBV4jbGdx2TX 12-13; 6fqDJhcVRap21w8W 3-10; 8yPRmZdrq4QhcGn1 1-4; af6xmcjndRMkZ83N 1; aVDjBvL8hyfCGmtq 1-2;
bDcBjXw82GvhYW4F 1; C834MAj1cKh2RZyd 1, 3-7; CJqgRYWTnm3jHGxP 1-2; g7ZWhALb3rT6ctfq 1-19; jLyXfQ2Ftp8RJmNd 1;
mdXkc6xwP8Y9pR47 1; mJDNgqYVGwb6A8tF 1-17, 19, 21-22; nc1zq8VBvZCTJtjX 1; NdbCqzpGA8Yr9ZfL 1, 3-6; pYATGQ6aCmzwBNcM 1-8,
10-13, 15-25; qQwFaXyTdhmnr2kP 1; RHgby76wQXp2nrxj 1-4; RJ82yLB7HcYrd39M 2; V8F7qLQXzpPAmcyG 1-4; vGdq3ZrKyMAPaTJD 5-8;
VYBjgJx8aZtRhmpy 10-46; WgQrVv2BTbPGHKFD 1; wk2zgGAXYTbnHQ9V 1, 3; WrBH1KZxJvVP8hQk 1-2; ya61L9CjgmMWAQTJ 1;
YXTfd7QPJV4Nqhcn 1-18.

## Electrocutioner Leg (9446)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1292851
- Rate: 3.95 PPM, measured

32,359 procs on 290,009 rolling hits, 207 wearers, 250 runs, in the main hand: all hits 3.94 [3.89, 3.98]. The rate is
all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.97 [3.87, 4.06], 7,736 procs
on 68,738 lone units, 119 wearers. Dual-wield white: 3.36 [3.07, 3.62], 2,439 procs on 25,594 lone units, 87 wearers.
Specials: 3.97 [3.86, 4.08], 4,305 procs on 38,306 lone units, 201 wearers. Two-roll abilities, priced as one roll:
7.88, 136 procs on 609 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 10,318 isolated
hits. Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Warrior |     132 | 23,244 | 210,509 | 3.90 |
| Rogue   |      58 |  8,748 |  75,983 | 4.06 |
| Hunter  |      17 |    367 |   3,516 | 3.68 |

Reports and fights: 19DdXA3BwjZHFbVJ 1; 249KNqjY61fcwdMn 1-9; 27dNt8HLBCPpAmYZ 1-31; 2N4yhcrz8GW7mRnt 1-6;
2jYyAK1RrPvTC8WM 1-5; 2ywhMnJq4Y3zaWtg 1-9; 3HRvtBnAbfM9xTkK 1-3; 3HZNbTwqdtnLFg4K 1-4; 3wCRWQcyNrgKvHz9 1-9;
46bPRkzVHgafvMwA 1-12; 48jfdvMD9HQbP1tC 1; 48pmMdNLnqxWj2Fz 1-12; 4C28YPXqgTMQJxB7 1, 3-9; 4M8gQTJhZvRPWf2B 1-5;
4Ppkndm9fWvTh1Gr 1-5, 7-12; 4Wq6Kcgnz3r29fTh 1, 3-5; 4ZNwHaLtygF1k3xC 1-3; 4jCpGrhx3NvnDmZ1 8-11; 4nbDkxaLmgcwGYQv 1-9;
4ynKGp8XCQPNmJLF 1-4; 629cLJxdjGkVBDTa 1; 69atW41MxnTY2cXN 1-26, 28-37; 6Pwh4WBYyAbMGD3c 3-6; 6QvVFqNYty8bzrcf 11,
13-21; 6WkrKt3RwHGmb4Zf 1-3, 5-7, 9-15; 6aBqgmpQhVAvn94Y 1-5, 7-9; 6jVXd48tLqWgQvK3 3-10; 6z9hqHARQr3vwLXd 3-4;
7f6DBRAPMLGYzyQm 1-2; 7kVKd6JMNTCj3P4z 1-10, 12-16; 7n6rKTfk1vhzgXyF 1-2; 7tYjam2PXvfcGz61 1-14; 7vxRAKHWC83tgPDz 1,
3-7; 7x3ycAVKT9Ba4dMZ 1, 3-7; 8VQDf7JtwbTknMLz 1-3, 5; 8dT1VNMQfLgrj9vc 15-17; 8rCvtag1mFYqBHM3 1-6, 8; 94FmbDrtqgPRKyH1
1-5, 7-13, 15-36, 38-57; 97CVmB4zqyrcWpjT 1-7; 9T6Zbt2QxPXhmJpk 1, 3-27; 9bd6VQvJR2ZqGW3k 1-6; 9g7NQBqDH6FGfP8L 1-4;
9v23hzNwDpWM67Xy 1-3, 5-18, 21-22; AGjFnqDY42gfQBTx 1-2; ARqz9awMbfr6QD7Y 1-2; AhxGmJv3nYDzTW9q 1-13; AkBH7pF6XvZc3Gdw
6-8; AxCwvBkjqYVtLgnc 1-13; Azq6hMZn2RyWwagv 1-6; B6jfw4mcnPxWCKYD 2-18; B7wf2rAVZJLQgT3m 1-5; B9RcfwgQpTHL8m31 1;
BTAM7pf2VtYCnJ8Z 1-16; BydpaqJVW29G8QAc 1-9; C4TgLGnkKZcPqF8W 1-12; CBK8LRX4tZzWANqF 11-13; CDZG21zLfmMdh6FN 1-5, 7,
9-10, 12-15, 17-21; CLXHZ3c2GyN6mYBD 1-4; CcAYgrB3Q1w8KJvV 1; CfFMHPahckGpR8Wd 1-3, 5; Ct8XcxY26Mf34NjR 1-9;
DF6Vy2mcQBbrNp7n 1-6; DTrLRQbfXny12B6t 1, 3-8, 10-11; DVdNpYmZfz9MGh1X 5-17; DcyRfjPp3YzHMkra 1-18, 20-22;
DqHT31fVnh4rjaky 3-7; DzWf9A8B1MNnrm2L 2-5; F9DyXdCazKmPqWfN 3-5; FBQzxArfjdCZp6Lg 2-11, 13-18; FarzJZ1mH2ydBWxh 1-17;
GCTbWRf9mrHqxAX6 1-3, 5; GZzaJH1QVjMKd8qA 21-34; GxTn6w2dD9MmXypP 1-9; H7PBRD9CqnGxkF4t 1; HRJA3nGpVzDaXyYP 4;
HZCD4zaBgtKh63rb 1-10; HmnhWYy8BtCGv6R7 1-3; HqzVaYrZLQxAPKhF 4-10; J2XL6AyZqgpjx8QK 1-13; JBj14VkRgmcTqDFd 1-7;
JHDxh1BWZFGTfQ2n 1-4; JWbdPL3FRphQcM6z 1-10; JZRnFQDrmYqBycvG 1-7; JcanfA1v3rXN6bBW 18-22; Jcx8g4CXq1QBpzVP 1, 3-12;
JzRmL9FYywDQgNcW 1-2; K3jWRx2qQ1aXy9PA 1, 3, 6-13; K47ntNhqVQz83H9g 1-2, 4-5, 7-9; KAfmrqFHXh7TzdL1 13-20;
KFdN4BrZP9X6tbcL 1-9; KGLhazAPkWfd9rw3 1-6; KTz6Gaym8fdCxhkb 3-6, 8-13, 15, 18-23, 25-26, 28-45; KZzq7NTcBXj4Ln9r 8;
LBYKWV4xgfkTvHRt 1-6; LJ43xqm7ATnCH8MX 1-13; LKkr1pxAtT7CvcZN 2-17; LMdRqJGhV17XwWyT 4-22; LQbJn6gcWz2D3HCT 1-2;
LWtG286RrDcJZ1xw 1, 3-5, 7-11; LXNz7Ym8rvwqcHax 1-8; LayrxDGdQb712wpC 1-10; LjFr1v7pJKVcfhWB 1-11, 13-14;
LwTzGJXrN3HWqCfD 1-5; LzXRwTc3YCfFkG89 16-21; ML1xyRfnNVbqkhAX 1-7; MmTXHCN4KDfpaFwz 1, 3-12; Mr7nxYapkv693wRh 1-10;
MvgK14QZDxjWGBXp 3-14; N9qCv8pYmPQRb7a3 1-18; NKnQBL2MrwCyAHzd 1-6; NTcKDtjPwZk8QaXg 1; NXAxdr9JVKLzFDWg 1-4;
NXgv93kr2G7zyTxm 1, 3-9; NcRqzah1TQjMr3nF 19-25; NgVPn7M3yX6AB8rT 1, 3-8; Nh9wMxkpzqJTPA2f 1-8; NmwLfZHTpdqPWAMC 1-5, 7;
NnmBjCcpKrxkg9GW 1-4; Nr8n1KpjcYwk2CvH 1-5; P1QTDgLwfytJKrBh 1-8; P3JYyjrc1FWgfN6q 1-4; PNnXcA9hFBx7Wtdj 7-11;
PpLZ7FAgkvTwc1Cr 1-11; Q6hKWVt9H47vNwnb 1-8, 10-12, 14-16, 18-23; QP32XzqKRHTvZ6FN 1-4; QVg7ANTY4B8RhDJM 2-9;
QwL7NjYPXAmzakRp 1-9; RD6bnVNTgPhGYCvy 1-4; RLnWB4FrJgpMyNZj 1-5; RPMNvqn7tXwJjak9 8; Rfwtj4ZF1nTx2mJp 1;
T13Kx4nhGJW9vR8Y 1-16; TC29cvWmBMhfZJzV 1-2; TGmKaXrQPp7czjx1 1-4; TJ8fyKCp4BHtWdLm 1-2; TdP4nFMvCf2mW6rB 1-18;
TgVP6hxZKqY7MGQ4 1-5; TnbzrVaxtHcpMKLY 1-3; V3L7kJDzTGg4QqCF 12-18, 20-27; V4LaQWhtxHqXDpwT 2; VQbrqK3D1RHXWpa4 2-20,
22-27; VaFNJB32Ah1WGLrn 11; VaYkQ91Tt74yKbrc 1-2, 4-8; Vm9wxAyCHQX7q8Bk 1, 3-4; VxmDcHTK8LqgNCwQ 1-7; W2rLkBmgb864JNDh
1-6; WA9GBTPQRKvazHmy 5-9; WBZ7gMLqXFcQrzDJ 1-2, 4-20; WCQGyn1J9q4Kdzbk 1-5, 7-15; Wfw3CbM28cyT7Nn9 1-9;
WjgGnJtpyKzVY3b1 1-21; WrHQ38tb4KTnAy7h 1-3, 5-11; X3RtbZzTdYAxfKV4 1-4; XhZAV37vMFbYWRt6 1, 3, 5; XpnkZCf7j6qDcmy9 1-4;
Y2n7kAxTCch1mgf3 1, 3-17; YZyAfq6XN9k4a8md 1-5; Ydx67HwgajJKpbNr 18, 20-22; Yq9kmzBHDP3jM2Z7 5-6; YzMA1vd3BrqLWxhH 1-3;
Z182HvNhFzdqgCrj 1-5; ZD7YLw2qy4mWRhJ1 1-4; ZKzP2YvapHfqM8WC 1-3; ZQ2NaCptbjmWKJ6P 1-18; ZaPRprzfmqY4CtF6 3-5;
ZrD7NdhjWX4zG2PL 1-5; aDXfbQzV1Bwr9n2x 7-8; aKwXBJkG2bW6tVNY 4-5; aMNVKZpg8tRkhBP3 1-17; apqZQxfDMJB4LPTb 1-7;
avnKJBPL6ANFTC3V 1-2; avrcxzCymh1DNqA8 1-19; bjcHnakF34w6YLvC 1-6; btxm3BCN4jX61VhY 1-7; bvzqmdxkA4KVQaZh 3-10, 12;
c7XrRMZWYmzwgfq1 1-3; cBX7WxzLNYrmMdyQ 4-6; ctRfwLFrZ8BpaTJY 1-7; d2wFYHqnNcJpA1gt 1-10; dZYgMabwJXnD1rfK 1-4;
fWRgvP17z8yrL36h 45-55, 57-62, 64-68, 70-73; g7b3Jf29CVhKtYMD 1-2; gdMNAjD2Kyrtqh97 1-8; gn4HAC1RqMfLt8wN 1-10, 12-21;
gnMWYZa9ArC8jbNc 5-12; h3bCaBdyYJMPzvRk 9-19; hAJa2PcBmdnNbGFq 1-16; hMaTJ3ZtDzqPWHfx 1-7; hZvn9rLzYyNTaVBR 1-5;
hatAfwb9QHydP62c 1-2; hf1pF42aPHMv9kLX 26-36; jDHpNnbqf68kZxA9 1-3; jLtGKrwPDBMmF3Xc 1-3, 5-11; jPh97VNpz3Ark2YQ 1-4;
jVRTr4dWa8xfZHgv 1-7; jgByJqcWN6HFdLY1 1-7; jz3qpYLQ4PwTyHtr 1-12; kB3gRfF4hYJ2nAQH 1-5; kX94Vmn8tpJzGTD2 1-11;
kmBptFTG68Dcq3jx 9-10; kr13fAvjzZXMdNDT 1-2; m2jhcywg86TDGMZn 1, 3-6; m3dN8Wkg4LyqXZD2 1-11; m9PX3r2NY8jLQ7Dw 1;
mN7ZTVPw8WdJrhA1 1, 3-5; mR9Ch8QVfX1pJkT4 1-2; mwAFfdjYy3chWMxb 12-15; ndHRLVB9Kaz7xpqT 1-2; nqmgcYrPz9B1fLDx 1-17;
p4MR7G2atyhQwV6k 1-6; p92NRBfL3QCYrAgy 10-11; pbhgCm6PjYwZ3fRd 1-3; pbvkXazLqcHFhgRf 4-7; pkJPxXb3aYLmG8qK 1-4;
pwNAZn643XgzcbYa 9-10; pwd6RvtM2NYqmfGT 1; pyY6Kg9bPRkQfMHL 1-22; pzx4LKwq6PB8FrWy 1-10, 15-17, 19-25, 27-32, 35-36,
38-41, 43-46, 48-49, 52, 54-63, 65-70, 72-80; q6VRad2JDnpxXYAL 1-13; q7xagw1ZkdvnLQmc 1-6; q8wrpJkfFWdVg6Tc 1-3;
qRJ8DvjZfh3Wzwgx 1-3; qT76ZPycRHrL41p9 1-4; qbDZa2XcdrM1npAF 1-6; qhGZDM3ygKAHtjbN 1-12; r9FD3YQGHcqdjC7g 1;
rBy3ctF2NpKPQGWb 1, 3; rykhNAvxCwXG4m3Y 1; tAFwQbTHpJMCxZXP 13; tCj8cKdQr47MN6qb 1-14; tc1qGLjpgVrJ8WdQ 4-5, 7-15;
tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19; v8VbYrxBCyMdJhN6 1-4; vkG2m6pNJ8PWrQh1 1-4; wV9frJcGT4nzm8vK 1, 3; wYmTW817qdpJHRcK
1-2; wrVYM4xbWLNaczhP 1-3; xR6rvTQLBwZ9nkzy 1-12; xmvMVYpyRjX3n8d7 3; yFWHKnAfCQpjNk96 11-12; yqVzktK2B7Lgp9Yb 10;
z9mYPDLpcM81JCFA 1-3; zD9hTZPfwC38bgq1 1-2; ztamxJTvkp82GK6h 1-18.

## Fang of the Crystal Spider (13218)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 17331
- Rate: 1 PPM, observed

365 procs on 13,746 rolling hits, 13 wearers, 14 runs, in the main hand or the off hand: all hits 1.00 [0.92, 1.08].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white: 0.90 [0.72, 1.08], 103 procs on 4,312
lone units, 13 wearers. Specials: 0.00, 0 procs on 27 lone units, 1 wearer. Travel time: the proc lands a median 97 ms
after its hit, 98th percentile 237 ms, window [22, 247] ms (96 isolated hits). Reports from 2026-11-14 to 2026-11-18.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Rogue  |      12 |   349 | 13,113 | 1.00 |
| Hunter |       1 |    16 |    632 | 0.95 |

Reports and fights: AaKzvcY1wmH9pBbq 1; D46Xm8ncR9WbTFrL 1-31; G7X1CkT8z6PrfjWM 144 fights from 695 to 871;
KNTYZJ3Xg9d8HftG 1; QXFd74zDW8JZMmKx 12-13; Tx7MPH9f8Ga6hdrF 1-20; XPYV9J6DfrMKxGqh 5-6; XgQYbwGBZcHLKRxV 4;
c8mjK4TdzV2M7XLF 1-5, 7-46; h3P9VjGMZCJWfTcm 7-10; k1taHbwL76Cnxv4m 4-10; kaH1G4ztmYpgMTJ6 1-2; xXWkB3fqAjt9NJGH 5;
y6D3jqgWt8TvpNYM 1.

## Felstriker (12590)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 16551
- Rate: 0.55 PPM, measured

864 procs on 57,996 rolling hits, 21 wearers, 35 runs, in the main hand or the off hand: all hits 0.53 [0.49, 0.56]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 0.53 [0.38, 0.67], 58 procs on 3,873 lone
units, 6 wearers. Dual-wield white: 0.44 [0.40, 0.53], 222 procs on 17,783 lone units, 15 wearers. Specials: 0.67 [0.42,
0.97], 34 procs on 1,804 lone units, 11 wearers. Two-roll abilities, priced as one roll: 2.52, 1 proc on 14 lone units.
Direct: window [-13, 100] ms, median 0 ms, 98th percentile 0 ms, 212 isolated hits. Reports from 2026-11-15 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       9 |   551 | 36,551 | 0.53 |
| Rogue   |      11 |   302 | 20,914 | 0.51 |
| Hunter  |       1 |    11 |    530 | 0.73 |

Reports and fights: 2jJ1L8FPHzTB4fKG 1-5, 7-12; 2wMznDRvmkpbtHxA 5-7; 2wPQmVqbdCcZW9gv 528 fights from 1 to 629;
4TxMfmFpzBJ6aN1j 5-7; 7tbWfYqJFd4vzL3p 1-2; 8Gw7m39ZXFz1y2fv 1-32; 8cqCXZHgpw9Px2hK 1-28; B6bvH1fQ4kAcyRzt 1-2;
FhG3AxYL9ZNP4RvV 1-15; KkRW7Am81LaJqHYz 1-3; M1tPx9DBZGYgCa4f 1-3; MqzhQN9FPHL3fBvT 1-2; NjWKPpHgymCT832A 1-19;
PKBd6Vz2Gkjwp9Zt 1-2; QNfaZ9MTnWpFrqb4 151 fights from 1 to 199; QkVRc329w8nLTzKp 1-3; TNW9rHBt3MD1X8Ag 1-2;
YFtvD87kznr1CThg 6-7; bP2CFzkZTLJ8q6Dn 1-8, 10-16; cgMGCdPJQ84wzaBV 12-13; dnDg7j9kRw6WtbML 61, 63-69; gYm2HAtf4bnC8QZT
6-22; hZLykw9dRnYbtvKP 1-2; jXTVfy3ApWKPYwDG 1-22; jn3dwzWPZ6r2ky1K 1-3; mpgnDQbLWXY8jJP1 1; n7hQW3NYzJwmDaG4 36;
nX1qbatTzGZ3CAwB 4-7; pRA73YvkarFVLNCZ 1-3; rdcZKVykDFH4f7mn 1-4; tkPZYqTRgCyWG3jM 22-23; vZqVmLYQa6jwP3A9 5-16, 18-21;
yPVkKpqtfJM27WFB 1-7; yfKb1MthLFDc4ANG 1-7.

## Fiery War Axe (870)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 18796
- Rate: 0.85 PPM, observed

583 procs on 11,685 rolling hits, 14 wearers, 15 runs, in the main hand: all hits 0.83 [0.74, 0.89]. The rate is all
hits, which the one-hand white rate agrees with. Observed: its interval reaches further than 10% from the read. One-hand
white: 0.94 [0.81, 1.12], 130 procs on 2,299 lone units, 13 wearers. Specials: 0.90 [0.71, 1.03], 9 procs on 167 lone
units, 2 wearers. Two-roll abilities, priced as one roll: 1.25, 3 procs on 40 lone units. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 1 ms, 110 isolated hits. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      11 |   518 | 10,015 | 0.86 |
| Warrior |       1 |    50 |  1,040 | 0.80 |
| Hunter  |       2 |    15 |    630 | 0.40 |

Reports and fights: 1L73Gh9FVMm2J6jg 1-4; 3A1rzFqNBxTQykwC 1-2; 4C28YPXqgTMQJxB7 1, 3-9; 4LrMDtG6T8AVRJzQ 3-11, 14-28;
7n6rKTfk1vhzgXyF 1-2; Kzn8AkXFr3yvMZHQ 1-9; WAxmzHdKLVabk6nR 1-8; Wb9tmLq2cGKFjw3v 1-6; ZTxYXrtHQkcNGjA1 1-8, 10-18;
hVxH72qr8PYB3DNm 1-11; pravyY1JCwDAxZLb 7-24; rTLVqFpfa9GJQgHh 1-2, 4-5; tvBjG6TXDZq8gdrz 4-5; vkayWxntqQJ6pfjB 6-9;
zDKZYwnyML7WFHXB 1-6, 8-17.

## Firebelcher (5243)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1291686
- Rate: 3.95 PPM, measured

1,416 procs on 12,725 rolling hits, 64 wearers, 70 runs, as a wand: all hits 3.93 [3.74, 4.14]. The rate is all hits,
which the lone wand shot rate agrees with. Lone wand shots: 4.01 [3.83, 4.24], 1,257 procs on 11,060 lone units, 63
wearers. Direct: window [-9, 100] ms, median 20 ms, 98th percentile 36 ms, 1,054 isolated hits. By report date: before
2026-11-18 3.78 (548 procs); from 2026-11-18 to push 114313 4.02 (868 procs). Reports from 2026-11-06 to 2026-11-20. The
sim does not simulate this proc yet (the generator lists it as a missing effect); the rate is recorded on the item's
effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Priest  |      23 |   697 | 6,389 | 3.85 |
| Warlock |      22 |   502 | 4,491 | 3.95 |
| Mage    |      19 |   217 | 1,845 | 4.15 |

Reports and fights: 1RrDAhp7ZvVYc9P3 1-6; 1RrFMXYLPJhT7xBn 1, 3-10; 1xDPwjtGBTzZWm7Y 1-4; 1ZJVMgcYWTyRQh2w 1-3;
287trgyqQLvjACK9 1-4; 2mW9A7wQ4bhv1YMF 1-3; 3PnWT1jGz4tagB6r 1; 4NWMr2dJPyH9v8az 1-2; 4P93K2QXbxJ68qrH 1-2;
4vca1FKMyTnP97Az 1; 6WK7XDRHBNCx9kwn 1-4; 6wxJrRWgPyQb21hm 1-7; 6WzAhcVgfLCt78nX 1-8; 7xdwabvCPyBXYRFq 1-4;
8rdDTxLMQPKqYyAz 1-3; 9pTK7ZwNabjMydzV 1, 3, 5, 7, 9; 9wKZ21fVDH3xpvcT 1-4; aNFQk87zxVRKA2nf 1-6; APhXWvd6CkrHL4N3 1-7;
aryNkPZX814F37pJ 1-2; Bm6KyjgHafc19xP3 1-3; c3NLR49a1VwfCQtG 1-7; CGkfgyRY8mQ1Aah2 1-2; d6bwAkLKjGq7pVMy 1;
Dh2tFcWaMVng6Cdj 1-7; DHv3m4KLCdntG7pV 1-3; DmkRhHYPg4wVJf3t 1; DNJ682zfmT7Zykv4 1-2; FADNxntvpJkfM9RC 1;
FWf1LcA7PCb6NQB4 1-4; Fwr6n9BDgLaGPmKY 1-10; gmTPbGL4cNDJ92fH 1-3; Jn4FXD6QKW71pRzh 1-2; kBzVYb3Cy7QwnKMD 1-6;
KDFtbX6ZVxMgATLh 1-2; KkCznWJf8yYHvGgX 1; kpNtDy4ZYjLTGAbd 1-6; KYD67nVqZT4jxWyN 5-6; LAP9T3KqpjhC4MD2 1-7;
LbmB2QdWnPT7JGAF 1-6; mbBT1kLrM8qXzdJF 1; MN6V1awdYqymj4vr 2; mPz2LqFWbhKcV7QN 1; MX1B9kCJ7GDTQ6za 1-3; NGhbmnMwKCHYfaX6
1-5; nT2Wz7mVQZbtBRG4 1-4; PCYpB6DhnQMmXbzN 1-2; PLKnY9R2BDgwbdWv 1-2; qFZN9PQwagyL3Xnz 1-3; QHqMJPfgbBCvZan7 23-28;
qL6waTVj3WhQYk8x 1-3, 5-21; Qvg9N1YBc36d8Rnw 1-10; r6dAtgyVpcMXkL3T 1; r9X8bgWqQfnPVDyB 1-2; rCnfgxYFVXLHvjZp 1;
RKdyjLbQCZgtzrXp 1-3; tAZ12pKHxQYLydgv 1-2, 4-5; TGVQg7A3KfJdpbkw 1; TkMHbgQNKBt7Vzy9 1-13; TQCphZqxGzRmkgjf 1-2;
tvDFpcQ4w3L7PBy1 1-3, 5; txZvdrQzpwn3V4kb 1-4; vVa1NxRf6rZnYhWH 1-2; vWGgArDXtc13Lx6J 1-9, 11-14; w1pq2CGTWMbX7dQZ 1-2;
wDqmkbxdTXj1rA9F 1; XPmJACH8RtqFd3jM 1-3; yAGCbvd2r4RZNDaK 1, 3; Z9qbhQrJg2Wfz1FM 1-2.

## Firebreather (10797)

- Proc: weapon proc 16413, logged as 16413; rolls on main hand, off hand
- Sim rate: 2.1 PPM, observed; measured 2.12 PPM [1.82, 2.61], 4 wearers, 4 runs, 676 procs on 8679 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| HMdLhWY3pNbnXPtF | 1-16   | Rogue   | oh   | timing + otsu for 136 swings |    45 |  750 | 1.64 |          |
| PcVWH9XhGRCvb1Y3 | 1-2    | Rogue   | mh   | timing + otsu for 287 swings |   317 | 4297 | 2.01 |          |
| QH63NmDVdWyrqwak | 1-17   | Rogue   | mh   | timing + otsu for 158 swings |   134 | 1925 | 1.90 |          |
| ypwYm41tKWZxV69q | 1-5    | Warrior | oh   | timing + otsu for 156 swings |   180 | 1707 | 2.88 |          |

## Fist of the Damned (10804)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 18084
- Rate: 3.1 PPM, measured

3,141 procs on 32,009 rolling hits, 22 wearers, 25 runs, in the main hand or the off hand: all hits 3.10 [3.01, 3.20].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.18 [2.93, 3.53],
764 procs on 7,598 lone units, 13 wearers. Dual-wield white: 2.86 [2.63, 3.27], 297 procs on 3,284 lone units, 9
wearers. Specials: 3.11 [2.75, 3.51], 257 procs on 2,608 lone units, 14 wearers. Direct: window [-10, 100] ms, median 0
ms, 98th percentile 1 ms, 856 isolated hits. Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      19 | 2,994 | 30,656 | 3.08 |
| Rogue   |       3 |   147 |  1,353 | 3.43 |

Reports and fights: 1GN2LMtwvY4JQHba 8-11; 498vhQAtCw2XgZJM 11, 13-15; 4vcn3WF1RxXf8VZt 6-17; 6MYqkGpyJxDVCvRg 1-3;
8xjZmDw2LKv9bckh 1-15, 17-23; 9MCqcwjzLk4JBARG 1-5, 7-14, 16-22; 9WGcCxKHQ4pkX6Rj 1-11; B1qcTGpnHCrQ64Xa 1-4;
By41zJVmM3cdPQkG 1; JMnDtLGxf4FqCV6w 1-3; KYmwz2q1tG8yAcdP 1-6; NcwFg2bZjkrnXfQV 3; RTfqrCdVvLGYzj1A 2-3;
Tx7MPH9f8Ga6hdrF 1-20; VjT3aMZghKCyzGrt 1-6; WMBFLQ43XvqTK7rG 8; Wb8w1TtxA3kLF9M4 1-8; Z4w7HKFPvcrTWd3M 5;
fNMJcqDHb9ZzPLnk 5-10; j8bQnLtWwRMBXkDz 1-4; mbZcgPavhdG7fKDW 1-20, 22-24; nmwM719fpABxvjdz 5-6; qprynNbXtRZV4DHP 1-7;
y6D3jqgWt8TvpNYM 1; yYPpk4TqAhjX3xC8 12.

## Flame Wrath (11809)

- Count: §Count of the 2026-11-12 to 2026-11-18 reports
- Proc: weapon proc 16559
- Rate: 1.9 PPM, measured

| Wearer           | Swings | With a proc | Chance (95%)        | Swing speed |  PPM |
| ---------------- | -----: | ----------: | ------------------- | ----------: | ---: |
| Druid, Bear Form | 16,248 |       1,272 | 7.8% (7.4 to 8.2)   |       2.5 s | 1.88 |
| Druid, Cat Form  |  5,899 |         197 | 3.3% (2.9 to 3.7)   |       1.0 s | 2.00 |
| Druid, no form   |    147 |          17 | 11.6% (6.1 to 18.5) |       3.3 s | 2.10 |
| Paladin          |    934 |          94 | 10.1% (8.7 to 11.9) |       3.3 s | 1.83 |
| Warrior          |    250 |          17 | 6.8% (4.7 to 11.2)  |       3.3 s | 1.24 |
| All              | 23,478 |       1,597 |                     |             | 1.88 |

The chance per swing grows with the swing speed from Cat Form to Bear Form to the weapon, so the rate is procs per
minute. Without the wearer who swung most in Bear Form, Bear Form reads 8.0% (7.3 to 8.5), 1.91. Over every landed
melee hit of the paladins (abilities, seals and Windfury Totem strikes included, priced at 3.3 s), one proc to each
Fire Shield apply or refresh, the rate is 2.05 (1.96 to 2.17). The rate holds across builds: Bear Form 7.5% on 345
swings on 70888, 7.7% on 13,118 on 70938 and 8.0% on 2,889 after 70959 was listed.

One hunter in 6yHgkT8LqP42NmxY, left out of the table, read 20 procs on 170 isolated white swings, 11.8%, 2.14.

Fire Shield applies came 100 ms to 1 s apart 50 times and 1 to 5 s apart 508 times, so the 5 s category cooldown
on 16559 locks nothing out.

## Frightalon (14024)

- Proc: weapon proc 19755, logged as 19755; rolls on off hand
- Sim rate: 2.35 PPM, observed; measured 2.36 PPM [2.17, 2.66], 6 wearers, 7 runs, 551 procs on 9994 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class  | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------ | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 16hJKBmq2V8tnzXR | 1-3      | Rogue  | oh   | timing + otsu for 160 swings |   129 | 2641 | 2.09 |          |
| 3B8AcaGNRKmdgPDj | 5        | Rogue  | oh   | timing + otsu for 157 swings |    49 |  743 | 2.83 |          |
| CtMF9BgnAmWkTbfd | 1-33     | Rogue  | oh   | timing + otsu for 29 swings  |   147 | 2783 | 2.26 |          |
| KG62VXF3a97CxtAg | 5, 7-9   | Rogue  | oh   | timing + otsu for 21 swings  |    70 | 1022 | 2.94 |          |
| Lxy1kqgVvFc24aPr | 6-11, 13 | Hunter | oh   | timing + otsu for 86 swings  |    59 | 1289 | 1.96 |          |
| MgmdJqxBWA3Tw2vz | 16-17    | Rogue  | oh   | timing                       |    16 |  195 | 3.52 |          |
| nX1qbatTzGZ3CAwB | 1, 3-7   | Rogue  | oh   | timing + otsu for 17 swings  |    81 | 1321 | 2.63 |          |

## Gahz'rilla Fang (9467)

- Proc: weapon proc 3742, logged as 3742; rolls on main hand, off hand
- Sim rate: 2.1 PPM, measured; measured 2.12 PPM [1.97, 2.27], 14 wearers, 14 runs, 700 procs on 11026 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.73 -> 2.16 (144 active minutes, 20% of hits, in 5 runs)

| Report           | Fights     | Class   | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | ---------- | ------- | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 2BZkQ4tRcFfdL37H | 1-8, 10-11 | Rogue   | mh   | timing + otsu for 35 swings        |    69 | 1244 | 1.85 |          |
| 3jqfdPNV4Xwy7g1Q | 1-18       | Rogue   | oh   | timing + otsu for 21 swings        |    10 |  257 | 1.30 |          |
| 4PtKGL1zcRV8Chfn | 1-2        | Warrior | mh   |                                    |    20 |  286 | 2.33 |          |
| 4RVqrpXkh7dKftAC | 1-2        | Hunter  | oh   | timing + otsu for 176 swings       |    22 |  230 | 3.19 |          |
| 6DZv2FAgWq7YrV3j | 1-19       | Rogue   | oh   | timing + otsu for 10 swings        |    55 |  889 | 2.06 |          |
| 7F3hKPycMYjXVJkb | 1-12       | Rogue   | oh   | timing + otsu for 7 swings         |    64 |  951 | 2.24 |          |
| 8tBp2ZK7XT4QJCgD | 1-5        | Rogue   | oh   | timing + otsu for 11 swings        |    38 |  529 | 2.39 |          |
| A3hT8CQBPJXWmtZM | 1-4        | Warrior | oh   | timing + otsu for 16 swings        |    51 |  642 | 2.65 |          |
| Mjr1czR7AknXqLvd | 1-8        | Rogue   | oh   | timing + otsu for 10 swings        |    56 |  937 | 1.99 |          |
| QH63NmDVdWyrqwak | 1-17       | Rogue   | oh   | timing + otsu for 158 swings       |    32 |  553 | 1.93 |          |
| TbB8q4RMK3LmYJNv | 1-7        | Rogue   | mh   | timing + otsu for 32 swings        |    98 | 1467 | 2.23 |          |
| XHDjdLb7R9P6NQm1 | 1-2        | Hunter  | mh   | timing + speed ratio for 47 swings |     0 |   13 | 0.00 |          |
| dt4M3ypRnTkwaVQY | 1, 3-10    | Rogue   | mh   | timing + otsu for 28 swings        |   122 | 2083 | 1.95 |          |
| tac7hXVRgkbZNMDy | 3          | Rogue   | mh   | timing + otsu for 4 swings         |    63 |  944 | 2.22 |          |

## Galgann's Fireblaster (9412)

- Proc: weapon proc 29644, logged as 29644; rolls on ranged
- Sim rate: 4.25 PPM, measured; measured 4.27 PPM [4.04, 4.51], 11 wearers, 12 runs, 1992 procs on 10778 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights    | Class  | Slot   | Hand split | Procs | Hits |  PPM | Left out  |
| ---------------- | --------- | ------ | ------ | ---------- | ----: | ---: | ---: | --------- |
| 24fAZdgMD6K9qx8k | 1-4       | Hunter | ranged |            |   217 | 1213 | 4.13 |           |
| 2xY1D3mCwagdy7Wc | 1-2, 4-6  | Hunter | ranged |            |   134 |  736 | 4.20 |           |
| 3LTnjm16YN2bDzCP | 1-6       | Hunter | ranged |            |     7 |   52 | 3.11 |           |
| 3MY1W4gBv9yXakD6 | 1-8       | Hunter | ranged |            |   203 | 1204 | 3.89 |           |
| 3WyYMqr8ZxQRN1vD | 1-7       | Hunter | ranged |            |   248 | 1415 | 4.04 |           |
| 9XQfBmZRaNWJkH3F | 1-20      | Hunter | ranged |            |   199 |  919 | 5.00 |           |
| AKXHxFdgavCh9fkY | 1-16      | Hunter | ranged |            |   229 | 1299 | 4.07 |           |
| CpDKr6AWJ4fw7xhg | 5-7       | Hunter | ranged |            |    67 |  390 | 3.96 |           |
| GYNx3nzHdpr7X1AW | 2-10      | Hunter | ranged |            |    71 |  370 | 4.43 |           |
| aANyFDgb69hxqY2w | 3         | Hunter | ranged |            |   148 |  894 | 3.82 |           |
| pJjZqyKxb81RYdvG | 1-3, 5-16 | Hunter | ranged |            |   181 |  844 | 4.95 |           |
| y3wkJZnA7gv4qzfX | 1-30      | Hunter | ranged |            |   288 | 1442 | 4.61 |           |
| DBQHpyFz7chZRM6j | 2-10      | Hunter | ranged |            |    63 |  321 | 4.53 | duplicate |

## Glutton's Cleaver (10772)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 18075
- Rate: 2 PPM, measured

2,632 procs on 39,055 rolling hits, 47 wearers, 51 runs, in the main hand or the off hand: all hits 2.02 [1.95, 2.10].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.07 [1.86, 2.27],
279 procs on 4,039 lone units, 12 wearers. Dual-wield white: 1.69 [1.58, 1.79], 552 procs on 9,816 lone units, 34
wearers. Specials: 1.73 [1.33, 2.13], 77 procs on 1,339 lone units, 15 wearers. Two-roll abilities, priced as one roll:
5.82, 13 procs on 67 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 704 isolated hits.
Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      18 | 1,630 | 24,533 | 1.99 |
| Rogue   |      17 |   706 | 10,336 | 2.05 |
| Hunter  |      10 |   229 |  3,198 | 2.15 |
| Shaman  |       2 |    67 |    989 | 2.03 |

Reports and fights: 1RVnhHGv2bwDW9Yr 1-2; 6VJxcqnYa4wdBWQv 1-9; 7NgdQzG6VcRaAJ9X 1-6; 7tCZB31vV2Jp6XWd 1-2;
8vDJ9jmHd4FA7Twt 1-2; 92rYfXaKn1CckRM4 1-2; 9VacfmDpXwCgzZWq 1-9; CR4DHc2wFgjd7yrp 11-13; CnNwpaYW83cJr6fR 1-6;
Gx3AXVk9cCaR18Df 9-10, 12-16; Jad2wbYTXQjBc1n7 1-2; JcwBmYGVgnMQH9xd 3-19; K1zQgCcqbnt3kHXT 5; KzNcw4X2JPRCgVx7 11-19;
L7JpRfWtAZ6T1hGx 3, 5; M7BnDF4VPXcWRaZ1 1-5; MkrFhTq2HPnzmCVJ 3; NrJZ7avW13QBVTGw 1-7; P8Znjv1fgBD2CAVY 2-12;
PDT2yACNzd4LMZ98 1-14; PJXwqxpANGtya937 7-11; PkWRr8jZp9ATL7N4 1-23, 25-26; W7VPQB18nHfkMaAc 6-13, 15-31;
ZDnyzq6XAxdk3K7g 1-5; b2YKZcFzWPtkxfpg 1-3, 6; bCcFJj4aQwP1rt8m 1-4; d8c6LnPYbKkDhgVw 1-5; dn62GTvrBYFQbWfV 7-20;
f8bZ6ytgj7mMGHCk 6-12; gBDMLv8J1K7nqkpd 1-25, 27-70, 72-74, 76-157; gBYCvhH6MFmbVwkp 1-3, 5-6; gaF6jK78BpnfZqxD 1-4;
gnM2vVF7zWXhm9KP 1-4; h9XqNAJk6aPntHpv 12; hbnaypTcAWmqDMt6 10-13; kpPdVKFmw72nCyM4 1-3; mQJtAfacYp7wGryx 7;
nX9kCRqh3WGa8g4d 5-6; p4zvjxb3QZJY2d9a 1-2; pjZHqhYvkLKTAQCJ 1-2, 4-5; q1ytzBDkYgd6jZMH 1, 3-9; qRMCzhyNYfmHAcZk 1-3;
qnyc1zwaB4FrvM9x 1; rzvB21j7yD9hpaVF 7-20, 22-30; t681v9FWrLq4Tkxp 1-3, 5-7; wZGf4Y6tjLDyTP92 1, 3-7; wvbng2zVTYRrc9Zd
1-2; xtv8KX4ZzfAmLFPw 3-12; yRQMz8LfGjHNBx1T 1; zwPKLTQ4W6YkyRqd 3-8.

## Grim Reaper (13054)

- Proc: weapon proc 14126, logged as 14126; rolls on main hand
- Sim rate: 1.1 PPM, observed; measured 1.10 PPM [0.97, 1.24], 5 wearers, 5 runs, 233 procs on 4113 hits
- Reports from 2026-11-09 to 2026-11-16
- Notes: 1 of 6 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 4HC91qxpR6hzKWdF | 1-2    | Paladin | mh   |            |    48 |  810 | 1.15 |          |
| 8dZb4CmyvLaWMGf1 | 8-16   | Paladin | mh   |            |    59 | 1073 | 1.06 |          |
| D9vw4QTYJZWMPnph | 1-3    | Hunter  | mh   |            |     6 |   95 | 1.22 |          |
| NXgv93kr2G7zyTxm | 3-9    | Paladin | mh   |            |    94 | 1602 | 1.14 |          |
| YJR78AMPDNKGQWjF | 5-12   | Warrior | mh   |            |    26 |  533 | 0.94 |          |
| LBZqKhR7WjvA31NH | 1-4    | Druid   | mh   |            |    17 | 2190 | 0.15 | feral    |

## Grimclaw (1481)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1292158
- Rate: 2.05 PPM, measured

783 procs on 11,492 rolling hits, 15 wearers, 15 runs, in the main hand or the off hand: all hits 2.04 [1.91, 2.19]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 2.28 [2.04, 2.45], 81 procs on 1,067 lone
units, 3 wearers. Dual-wield white: 1.90 [1.47, 2.21], 161 procs on 2,539 lone units, 12 wearers. Specials: 1.63 [1.39,
1.79], 44 procs on 810 lone units, 5 wearers. Two-roll abilities, priced as one roll: 2.61, 2 procs on 23 lone units.
Travel time: the proc lands a median 163 ms after its hit, 98th percentile 249 ms, window [72, 259] ms (36 isolated
hits). Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   467 | 7,279 | 1.92 |
| Rogue   |       6 |   300 | 4,018 | 2.24 |
| Hunter  |       2 |    16 |   196 | 2.45 |

Reports and fights: 1QpaxjXWFDA7vPV3 1-3; 6rgDRnVFX9a3pbGy 1-4; 7awbvWHZXYF8Kxgk 1-5; 8WgnxAakbzDZdFBX 2-6;
AwYnpMhJDxPFavjr 1-10; Fy81NHrkKD3W97qw 1-2; HJxqkQPAnmKtwNjd 1-16, 18, 20-24, 26-31; JgL3qFZA4aQcpXPC 1, 3-9, 11-13;
KjLXqdCZANvQc9BT 1-3; Lcy3V1wHPpqZQAG7 2, 4-12; PJGwRWQYyA3nzXpv 1-12, 14-19; h9pRvCYG8fzNDrcX 4-11; k6XaY3Dy4r7jfTK9
1-3; m9JVLBndtZ63vhDW 1-2; mBRL3hY1zjaKqkX4 1, 3, 5, 7-25.

## Gryphon Rider's Stormhammer (9651)

- Proc: weapon proc 18081, logged as 18081; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.97 PPM [1.67, 2.22], 5 wearers, 6 runs, 567 procs on 6383 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights                                                             | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------------------------------------------------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 8bWqTVN2QCvK7HPR | 1-13                                                               | Warrior | mh   |                             |   106 | 1057 | 2.23 |          |
| 9cQz2ymYMgfL7TwK | 132-140, 142-143, 146-155, 157-159, 161, 163-174, 176-181, 183-190 | Rogue   | mh   | timing + otsu for 68 swings |    73 |  871 | 1.86 |          |
| A3hT8CQBPJXWmtZM | 1-4                                                                | Warrior | mh   | timing + otsu for 16 swings |    68 | 1055 | 1.43 |          |
| Mjr1czR7AknXqLvd | 1-8                                                                | Rogue   | mh   | timing + otsu for 10 swings |   136 | 1557 | 1.94 |          |
| cx8pPRvN9Ta2zHg1 | 1-10                                                               | Warrior | mh   |                             |    79 |  683 | 2.57 |          |
| mNBAdW4ajKHzhMYZ | 1, 3-5                                                             | Rogue   | mh   | timing + otsu for 23 swings |   105 | 1160 | 2.01 |          |

## Gut Ripper (2164)

- Proc: weapon proc 18107, logged as 18107; rolls on main hand, off hand
- Sim rate: 1.55 PPM, measured; measured 1.54 PPM [1.40, 1.70], 9 wearers, 9 runs, 383 procs on 8294 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights   | Class  | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------ | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 8tBp2ZK7XT4QJCgD | 1-5      | Rogue  | mh   | timing + otsu for 11 swings  |    37 |  934 | 1.32 |          |
| Ag69GWdLRDhKCrcT | 1-6      | Rogue  | mh   | timing + otsu for 9 swings   |    74 | 1635 | 1.51 |          |
| H1AnQVTYJzcam79x | 1-8      | Rogue  | oh   | timing + otsu for 8 swings   |    22 |  641 | 1.14 |          |
| VG18w7rW3KRQa9B4 | 4-10     | Rogue  | mh   | timing + otsu for 5 swings   |    51 | 1042 | 1.63 |          |
| WaHMR9NKLrwbzptn | 1-2, 4-7 | Rogue  | mh   | timing + otsu for 3 swings   |    69 | 1426 | 1.61 |          |
| fz2HQxXk96ywVYv7 | 3-8      | Hunter | oh   | timing + otsu for 109 swings |     7 |  117 | 2.00 |          |
| g41xhtdFCXZ8Vr9z | 1-2      | Rogue  | mh   | timing + otsu for 14 swings  |    62 | 1248 | 1.66 |          |
| qNyQvka41YhVmdHR | 1-4      | Hunter | oh   | timing + otsu for 29 swings  |    37 |  630 | 1.96 |          |
| v2aNxQ7mGTdkJryV | 1-7      | Rogue  | oh   | timing + otsu for 49 swings  |    24 |  622 | 1.29 |          |

## Hameya's Slayer (15814)

- Proc: weapon proc 16406, logged as 16406; rolls on main hand, off hand
- Sim rate: 1.1 PPM, observed; measured 1.08 PPM [0.89, 2.76], 3 wearers, 3 runs, 104 procs on 2888 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 9zxHaWbYT3gqDQvF | 3-18   | Hunter | mh   | timing + speed ratio for 21 swings |     2 |   22 | 2.76 |          |
| PcVWH9XhGRCvb1Y3 | 1-2    | Rogue  | oh   | timing + otsu for 287 swings       |    69 | 1924 | 1.08 |          |
| vFQhXgw42KmfjT8W | 1-3    | Rogue  | oh   | timing + otsu for 209 swings       |    33 |  943 | 1.05 |          |

## Hammer of Expertise (8708)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1317432
- Rate: 1.85 PPM, observed

539 procs on 8,403 rolling hits, 7 wearers, 7 runs, in the main hand: all hits 1.83 [1.72, 1.95]. The rate is all hits,
which the one-hand white rate and the specials rate agree with. Observed: every wearer is a warrior. One-hand white:
2.04 [1.94, 2.14], 149 procs on 2,085 lone units, 5 wearers. Dual-wield white: 1.85 [1.49, 2.70], 37 procs on 571 lone
units, 2 wearers. Specials: 1.90 [1.37, 2.31], 70 procs on 1,050 lone units, 7 wearers. Two-roll abilities, priced as
one roll: 3.57, 11 procs on 88 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 186 isolated
hits. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   539 | 8,403 | 1.83 |

Reports and fights: 89yPg6z4AjZGfWRm 1-12; BQnXb3hTC4rmJ2pd 1-2; J97GvmrRXCFZfQ8a 1-2; JtvQWAmT7HCDV8h4 4-15;
ZgTqAy2xaHXL8djt 1-3; dGXqkVhwT17nHfBc 1-2; ybFmNGwXcnBQW1rz 1-4.

## Hammer of the Northern Wind (810)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 13439
- Rate: 4.5 PPM, measured

4,649 procs on 29,390 rolling hits, 12 wearers, 16 runs, in the main hand: all hits 4.52 [4.40, 4.68]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 4.46 [4.02, 4.81], 366 procs on
2,345 lone units, 5 wearers. Dual-wield white: 4.21 [3.20, 4.59], 159 procs on 1,080 lone units, 7 wearers. Specials:
4.17 [3.76, 4.86], 338 procs on 2,314 lone units, 12 wearers. Two-roll abilities, priced as one roll: 7.79, 3 procs on
11 lone units. Travel time: the proc lands a median 116 ms after its hit, 98th percentile 326 ms, window [-10, 336] ms
(146 isolated hits). Dropping the swapped stretches (20 active minutes, 1% of hits, in 1 run) moved the read from 4.47
to 4.52. By report date: before 2026-11-18 4.38 (1,925 procs); from 2026-11-18 to push 114313 4.62 (2,724 procs).
Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 | 2,685 | 17,517 | 4.38 |
| Rogue   |       4 | 1,964 | 11,873 | 4.73 |

Reports and fights: 4Mcm2YV8yDhPnwrb 1; 8qACFaGbm1jn3PV9 14-26, 28-29, 31-35; cQpq1A6BxVMZFmfr 1; dBNfgVDLMmnc8h4v 10;
Gak2MmQ34xn9ctPj 1-34; J38xbrFytgM9p1LT 7-19; jvqAx3thp6wTz9ZR 1; krA2h3ZCfBXMmT4F 1-8, 10; nJZvaYCj2brRc7Fw 1-24;
nyAx3XNDVmYWGdrh 1-6; P4bXMwRjD1zyhZdJ 1-2; pravyY1JCwDAxZLb 1-24; qR2rnzfgPakD879W 1-2; TF6dc3NCHbKrfxJt 1-6;
xbg9BypFcdVDwqMk 1-5, 7-14, 16-42; z74YTAg1Jc2QFWDC 1-2.

## Hanzo Sword (8190)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 16405
- Rate: 1 PPM, observed

492 procs on 19,895 rolling hits, 17 wearers, 20 runs, in the main hand or the off hand: all hits 0.99 [0.91, 1.07].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 0.75, 10 procs on 530 lone units, 1 wearer. Dual-wield white:
0.98 [0.87, 1.12], 231 procs on 9,417 lone units, 16 wearers. Specials: 0.33, 1 proc on 123 lone units, 1 wearer.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 164 isolated hits. Reports from 2026-11-14 to
2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      11 |   309 | 12,384 | 1.00 |
| Warrior |       3 |   103 |  4,180 | 0.99 |
| Hunter  |       3 |    80 |  3,331 | 0.96 |

Reports and fights: 4f2ygdYCzFAvqc6K 6-8, 10, 12-14, 16-18; 6xHzpKd8wCna37Lb 1-14; 824kNhdFaALMBnvR 1-4, 7-18, 20-45;
DNc6QALhxmGbCH3Y 1-6; GHK2gB78FMVCmPfW 1-9; J2ThDq8RCPKNZdrB 7-11; KPqgbZxhn6BXwCvJ 1; MK6XJDzfLpFjnRWc 1-5;
bZaY8JyWxpmwQDjB 1; hwrmtk6aA7zJ8N1Q 1; jn3dwzWPZ6r2ky1K 1-3; kaPpXgt1bWJBr4z2 1-3; mtDjVd6a9rTgWBnC 1; pRA73YvkarFVLNCZ
1-3; qNyQvka41YhVmdHR 4; qbtZdThvBCXVcfAW 1-6; rTDcQk4gGxfmjN1q 1-7; x4TPAZXnW8kbv19j 1-6, 8; yPVkKpqtfJM27WFB 6-7;
zqpJDYHZC4Pa76bV 1-2.

## Iceblade Hacker (13952)

- Count: §Count of the 2026-11-09 to 2026-11-19 reports
- Proc: weapon proc 1298416
- Rate: 2.05 PPM, measured

813 procs on 11,764 rolling hits, 8 wearers, 11 runs, in the main hand: all hits 2.07 [1.94, 2.21]. The rate is all
hits, which the specials rate agrees with; the one-hand white swings stay under 50 procs. One-hand white: 2.07 [2.01,
2.11], 43 procs on 623 lone units, 2 wearers. Dual-wield white: 1.54 [1.03, 2.27], 63 procs on 1,231 lone units, 6
wearers. Specials: 1.93 [1.52, 2.21], 70 procs on 1,088 lone units, 8 wearers. Two-roll
abilities, priced as one roll: 6.82, 5 procs on 22 lone units. Travel time: the proc lands a median 99 ms after its hit,
98th percentile 206 ms, window [-10, 216] ms (176 isolated hits). Dropping the swapped stretches (75 active minutes, 19%
of hits, in 4 runs) moved the read from 1.67 to 2.07. Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       3 |   363 | 5,588 | 1.95 |
| Hunter  |       3 |   268 | 3,559 | 2.26 |
| Warrior |       2 |   182 | 2,617 | 2.09 |

Reports and fights: A1yxMTmKatkf3HRJ 15-17; B4wWqM2TbRmdQy7Y 1-2; DNc6QALhxmGbCH3Y 1-6; JQwbjnVYyGcP4rg8 3-13, 15-17;
KYmFPtyJ3gkzHfVM 102-160; Kphk63jtr1wb9NyL 1; fZdz8MyCHKFkrRwB 1, 3-7; gw8fpa1YVkv4WP32 1, 3; rTKYMd4yDvaL6Gn8 1, 3;
xWDkmwXKrRcdM7Jv 1; xyGf9QY4dwT8JHMp 5.

## Iron Knuckles (2942)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 13491
- Rate: 2 PPM, measured

373 procs on 6,508 rolling hits, 7 wearers, 8 runs, in the main hand or the off hand: all hits 2.02 [1.91, 2.12]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 2.14 [1.87, 2.26], 150 procs on 2,475 lone
units, 3 wearers. Dual-wield white: 0.95 [0.59, 1.56], 23 procs on 858 lone units, 4 wearers. Specials: 1.78 [1.39,
2.27], 33 procs on 655 lone units, 4 wearers. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms, 140
isolated hits. Reports from 2026-11-10 to 2026-11-17.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   289 | 5,002 | 2.04 |
| Rogue   |       4 |    84 | 1,506 | 1.97 |

Reports and fights: 38D6JN4nZzjcGtHg 1-4; 4QAtNfTXMyDGPqCr 1-13; AzbN8GY9DnBrH4aW 19-31; D8vykhYFL14KH7xX 1-3;
DwWM9H3rpTnGPfbt 1-14; GRTmQn3VDpwZKbCA 1-6, 8; H3vdcCr6JX1gwYtq 1-3; ZPyMjD1FkVGJH8Lb 1-4.

## Ironforge Greathammer (279895)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1322295
- Rate: 2.9 PPM, measured

12,165 procs on 81,815 rolling hits, 227 wearers, 259 runs, in the main hand: all hits 2.88 [2.82, 2.93]. No lone-unit
rate applies, since its travel window passes 400 ms: the rate is all hits. One-hand white: 2.85 [2.73, 2.98], 2,475
procs on 16,796 lone units, 182 wearers. Specials: 2.92 [2.71, 3.12], 624 procs on 4,136 lone units, 47 wearers.
Two-roll abilities, priced as one roll: 4.58, 9 procs on 38 lone units. Travel time: the proc lands a median 137 ms
after its hit, 98th percentile 601 ms, window [32, 611] ms (2,202 isolated hits). Left out: 180 of 440 runs in Cat or
Bear Form. Reports from 2026-11-05 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |     178 | 8,146 | 54,859 | 2.87 |
| Warrior |      47 | 3,992 | 26,709 | 2.89 |
| Druid   |       5 |    27 |    247 | 2.12 |

Reports and fights: 17jHXDAWBRQdgqV6 1-4; 1HK7gFbwy8aJPAh6 1-3; 1fpaDyG3mFnB9KW8 1; 1h9P4rYzWDvtHRxb 1; 1pDcTvkrCBhZJRgf
1-14; 1rH7mXJAxcqkRGt2 1-4; 21j6cGMKRFLpzqrf 1-8; 2FytXTANpbh6dRx1 1-3; 2HahmvNTK89f3qxQ 5-9; 2gjR3tVq4Zdw7kzK 1-3;
2kFjvxn3MzRLQWTD 4-10; 2mW9A7wQ4bhv1YMF 1-3; 3YRkhNP8BZbtDWgx 1-8; 4279fmhA1HPGpDcX 1-2; 48QbknMZw2xt1R79 1-4;
63KPqwWGkh29Abmj 1; 6A24xBzrQkqRFpKh 1-9; 6GcDXN8kWnb4P9mh 1-3; 6HTpWDrd8LKGXwjf 1-5; 6NdR8TAqXMyLmkYg 1-3;
6Y8AjXnH4daWgKym 1-6; 7WLGVBNDXqcar6Ad 1-2; 7kYG3K21FmyMVABD 1-4; 7qTnPC4NDM6Ypc2W 1-3; 7tahbgBF1jQGRKMq 1-3;
7zK9fMmcNVxnLQA2 1-3; 7zcat1AyFnrJ4BQR 1-19; 8134cDXJahtPZB9W 1; 8PD6Lc3j2qQ91RC7 1; 8Z1TNzMHLCXyGj7k 1-2;
8c1GqjXaPHTpZthY 1-2; 8c7DafNhKtndMBzm 1-3; 8xcD6XhLdNrGq3bP 1; 91YHvCrBN7atKcAq 1-4, 6; A4dNxBP7bY2XkMnc 1-3;
ACvNpyX4DafTgcZw 4-18; AGLkhzpCJdVnPDF1 1-3; AL1DxTaHGfVygm3p 6-15; AQJdTrDPn8qyK4Zt 1-6; AfntVm7bWdMvrP1w 1;
B8AKvLZ6bQpaXktV 15-17; Bbd2CDQZyrVv7gFw 8-10; BrTM6CbayRvXgmc7 1-3; ByrG7LFC9mH1N8gD 4-6; C1fWAmVtPFBGK6ky 1-3;
CD1QAmVqHrLyY2Fc 1-8; CQ1hBPYpTF9jZyak 1-2; CXqJYaRfBjnMFxD6 1-15; CYNKBypF2g17j6tA 10-14; CdH9vgqPbRX1JKM3 1;
Cn6rjJpgNBDP7Vxa 1; DmyKc6zWJ7MwHdNR 1-4; DrCRwbFGX3dmtHVW 1-5; FHpqmBdPGYZ3JaDy 1-5; FTw7k9hdbYKVfH8C 1-5;
FVPg91ybaXmGtAYp 2; FkNW7XzKjdVxbqfJ 4-21; FqVXJryknZYWTvKH 1-2; G3KCZD9NbzVpH1Xk 1-4; GDpvwrBfZ2yYCxT1 1;
GVqw3aXB9g7tnZ8W 1-8; GaynFPTvZf4Xq2b3 1-2; GdDn73hPYt2JQmHL 1-5, 7-8, 10-12; GwtNT4JLhd8AVWkD 1-2; Gyw6BrhCDx42qcHT
1-4; HBXA4RjC91yxbNQV 1-2; HKQW3wCmr4kJXz2j 1-3; J4LX7AMRmxVh96P1 1; J8YVraRnGTbtyq6L 1-2; JPDfnQGCH9Xwcbv8 1-4;
JnWQzq1m82fjg6Vk 1-6; KGgyJXnVzvMxc28h 1-4; KHqR9VAYTzZ24wxJ 1-11; KVh6PxG3fHTDWw8A 1-3; KYDWXztmZHQNCVkF 4-5, 7;
KtzqRMDGk27NwYyp 1-3; LJ8NCMHzpym4dVqk 5-10; LPQaWZVAkzpbJGFw 1-5; LVwmRx1Mqak7vp4y 1-8; LmhQHWJDvcV1RdPj 1-4;
LthNmKqbQFaAZ2JM 1; MBR3dy4azD1L78Qq 1-3; MQjCX4kG28BxztgT 1-4; MWqZ1xaVht9TkXJA 1-6; MhfdwPABCaxn9b82 1;
NFWnKbtVw8TqCcZp 1-4; NH8YCdhDgtXG6bZK 1-4; NJhx3D4knajQTYcP 1; NLbKRCyqrYPJdmFg 1-6; NRH63zjvtwpxn1Wg 1-3;
NfK46gDLj8wrJz39 1-3; NngHmf2D46Gw7bcL 1-3; PBkXcr2JGgx4NR9Q 5-8; PJVn4jX8WA1TQzg6 1-3; PK4Z2NcFnQdpLwYg 1-4, 6-7;
PN7jfn93M6HKwpGX 1-3; Pck4Hb83WgXxrfJF 1-3; PrQ4yZdgFbGTcpWv 1-4; QFCfmhc2yt4zR168 1-10; QMH23Y1hqCxamFLj 1-8;
QmCHxRYD9tbjLfW1 1-4; QnNkg8ATL4Dbv1hf 1-4; R76N48HMyW9V3ZYc 1-5; RGw7haYADHg8yC2m 1-3; RHDbqGV3fgjMrLCX 1-9;
RHYb2XPvpDdf17LA 1-5; RKDTjJY4mNc9wbVk 1-5; TAwnY8WFB4HmQZ9X 1; TBYt3rwJqhGxD1py 1-3; TGvNBcgzCQYD7hya 1-12;
TQ4PXCqWA3KGYzh2 1-5; V1L8nj6gMpqRB4WN 1; VkPnGcCgJ17b2YvH 1-4; VxcH4X3W8da1jYLM 1-4; VyjQ1bDHvPZFdxcN 1-10;
W3wGpxR4tQFcak9K 6-9; WnjRtrhBVyYwXx3M 1-4; XR78naNGyAmZhMKB 1-5; XWgjPGZD3ryACTv8 1-3; Y3cP8mqfA7F9RjKk 1-4;
YC214LNaFPwkRrfm 1-12; YFBpbDdgxG17CXhL 1-4; YM7tXFPkLZhTgj6z 1, 3-7; YPaFmdJbpXt1W8ZN 1-4; Ynw8cNZv4GJRBC1L 1;
YqnVZLX68FRmctKQ 1-25; YxKjwhRDNyJ4z9cF 1-7; ZADc3bTCKhn9jJdz 1-3; ZPBTJjpyxndNg1Lm 1-2; Zfwc2Q1bKRpMnWGg 1-7;
ZmQq1dyC8VJRcpAM 1-4; ZwBW1aVvkDAfcN3z 1-4; aD6nv3CRjrFqXfd2 1-3; aMVzJNgDwpbG143A 1-3; adR7kpKcMDmtHThG 1-8;
agztmnCXVNRGqwvx 1-4; anKHZjtTkdVv239J 1-4; arf1LxvVJGycRFAm 1-2; b2rJTDtLjC7BxgGv 1-3, 5-6; b6zYpvJBq2CVLyfj 4;
b9xGzmYPRWNCQcXr 1-3; bDrG4Vpyc9amg1KR 1-3; bTFC3LZRKaHgq8mJ 1-13; bVNRgda481CqpmfG 1-3; brAQWYqMfywd1JPt 1-3;
c7zvhTJGCLaQPZKp 1-5; cH3w7AN4VZXTfr6W 1-3; cT941jZhxvWM2aKd 5, 7-8; cnz1wmVkXtgGJjZA 1-5; dAgmNB1rwGy8f76Z 1-2;
dM46X2j1BZGhYTmx 1-2; dQqfmPcwk4avZ6Gg 3-9; dafrWBRmTJpnPVYv 1-5; dyTRnv6MZQ7BKaGC 1-4; dzBLDHJTfc13yAWQ 12-27;
fCbHhaK7jxv9FkYq 1-2; fDycAFZWt4NkRX1x 1-3; fTHKyQa6YFVj7JWL 1; fcCgK1R6GpBz8tPX 1-4; fjac46MdLgvHprZq 1;
fpGWB7am4PzYnMyN 1-3; fzyDKTP7x8B2hRNw 1; g9xAZ7jBQmRN6TGn 1-5; gNMJVXAKF7hH6WQx 3-5; gPbhJ73prMtxXcjn 1, 3-7;
gkFbNhdxDc384Pyp 5-6; h28ZpnT3kCm6NHMb 1-3; h2CgmZyNBarVjtYP 1, 3-5; h9L8pBqnVdJ1vDrb 1-12; hBWtTfC1rk8nMQmX 1;
hM67TLJGypgwvrzP 1; hVFtnpRGby9azXg8 1-2, 4-26, 28-42; hdtzWJKAZ8cHpGRx 1-3; hfJZb8HL41KmtdMA 1-4; hjkVZTFrGbRv7cH8 1;
hwcWQn8VLN7kJxmT 1; j1CGayX43bMdF8hP 1-3; j23K74gP1LNyYz6d 1, 3-6, 8-17; j3vLB41G8pkctPMH 1-2; jQP1aWnwH8C2T7Lt 1-4;
jpF1ML4ZTkcwGQ8W 1-3; jvtkpbcQCYFz83mT 1-3; k2rGgh6TVDXzc934 1-2, 4; kY3DG8tzBr2vX1Rm 1-2; kczT3CJVBXa9Gj84 1-16;
kyLDgPjrQYdHAc1a 1-2, 4; kzpRQ2gqGJ3yKh9Y 1-4; m9A2MvGDfPFRCWj6 1; n4xHq7mpR2htyXBr 1, 3-11; nFHZX8TMVG61PRpW 146-149;
nWhPF7JrzNRLmx2v 1-5; nadqr4JWQ82ZTkv1 1-3; pG2cZKTAjfYzv8yX 1-10; pJN1cYKwfgaZ8jt4 1-5; pJQ4mtrDqcaAzwBP 1-4;
pRrGZhvDw83LJHTy 2-4, 6; pT6jGtaZFDr92K7R 1; pqBwzhZQy7d9jFnY 1-8; pqPBFkXgnjDvN9ZG 1; pwNxagqK7HBZ1b9X 1-5;
q2h6pRvdBCKMabnL 1-2; q6bKHQCLRJaWTwGy 1-2; qCNzdMA63yLWmhGk 1-3; qFbMHLVtz94cXnCy 1-4; qNBhWJka9TtLPwV2 1-4;
qh8Lgfy31XnQF4jH 1-3; qxY87bFJjG6Tz14A 1-3; qyp6QA92RLJNKxTd 1-14; qzBa8HcG7xAfPyT3 1-14; r2c7bMKxfd3AHVYT 1-4;
r7RQ8FJpnBVgAy6f 1; r8paZAQhXF24Wdw6 1-2; rC3VbJjLQ6YBp8vM 1-3; rL2DzHGy7t4wxh3N 1-3; rLmJfNtq7v6BFcxy 1-2;
rQdnmBHXCpKgJqw4 1; rQt1k2LGcwVfFad7 11-14; rcYRpQA1aZVhPz9B 1-3; t9DRvJmjap3BrcCw 1-4; tH8kRGA3bph7V9vN 1-3;
tNxJXzB3q7rYfRK8 1-5; tWV1d42zgK7H3qNM 1-4; tZK2jg7HqTGAWnRd 1-3; tbQBwaTD7R2rxpkP 1-3; tgqAx3jr6vGW8nck 1-2;
thJy8Y43TPRnHvaK 1-5; v3WBbzYNjaqQJ67c 1-5; vCD8yMjK4WL1ZnqY 1-2; vRmX3tw8dbkx1HTV 4-8; vVCHXQgNFhqMnbjp 1;
vXnHY1xFwVzDBa2k 1-4; vf3QxCLyhckPdbM2 1-5; vk4gw8GBahr6VMxb 1-2; vm7QzHZCjFGtJ6DM 1-2; w3jgGmAkt7Rfnxah 1-3;
w4JfLXqF7HDK32ZV 1-3; wH9M8XGJN1K4h3tk 1-7; wTYjbBFDZN2zkLyV 8; wfMAD38qZRybWmn6 7-9; wnJHrLF14XD6d3RT 6-8;
wtzQKj8GP4cJ2W6H 1; yQ624zGVT18w9nbd 1-10; yVATBnQjpz1MLZPh 1-2; ykcwLtjHFQT87Mv9 1-3; zfW2RHdjLqKA1yD8 1-5;
zr6Z1GFPkpwQVmX3 1-5; zyFZC9kAVDB3Mvj1 1-3.

## Jom-jom's Gabbar (274952)

- Proc: weapon proc 1295736, logged as 1295736; rolls on main hand, off hand
- Sim rate: 1.75 PPM, observed; measured 1.74 PPM [1.45, 2.04], 7 wearers, 7 runs, 389 procs on 7049 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.47 -> 1.76 (49 active minutes, 17% of hits, in 2 runs)

| Report           | Fights    | Class  | Slot | Hand split                        | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------ | ---- | --------------------------------- | ----: | ---: | ---: | -------- |
| 3TDwbg41YhvNpPAq | 1-7       | Rogue  | oh   | timing + speed ratio for 5 swings |     1 |   27 | 1.16 |          |
| 987q1pjnYdVQbCmT | 1-6       | Rogue  | mh   | timing + otsu for 11 swings       |   105 | 1582 | 2.10 |          |
| BqdMTtCvN3fVJpRk | 29-35     | Rogue  | mh   | timing + otsu for 5 swings        |    79 | 1284 | 1.94 |          |
| HndJP9Yf4xc7VzKw | 1         | Hunter | oh   | timing + speed ratio for 8 swings |     1 |   15 | 2.04 |          |
| NfqRryF7gnMcX2ak | 1-2       | Rogue  | oh   | timing + otsu for 14 swings       |    47 |  738 | 2.01 |          |
| Pjz8kK2XvxyQrTCB | 1-2, 4-16 | Rogue  | mh   | timing + otsu for 62 swings       |    89 | 2130 | 1.32 |          |
| cAYgpkdqXRDynzPK | 1-3       | Rogue  | mh   | timing + speed ratio for 3 swings |    67 | 1272 | 1.66 |          |

## Julie's Dagger (6660)

- Count: §Count of the 2026-11-13 to 2026-11-17 reports
- Proc: weapon proc 8348
- Rate: 1.05 PPM, observed

218 procs on 9,524 landed hits, 9 wearers, 9 runs, the warrior run left out (its off-hand Whirlwind rolls the dagger):
1.06 [0.92, 1.20], the per-hit interval. With the warrior run the rate reads 1.10.

| Report           | Fights  | Class              | Hand     | Procs |  Hits | Rate |
| ---------------- | ------- | ------------------ | -------- | ----: | ----: | ---: |
| JMnDtLGxf4FqCV6w | 1-3     | Rogue              | off hand |    42 | 2,252 | 0.86 |
| nyAx3XNDVmYWGdrh | 1-6     | Rogue              | off hand |    40 | 1,672 | 1.10 |
| 4nNqbfmrCT2Q9PpH | 1-2     | Rogue              | off hand |    29 | 1,345 | 0.99 |
| N3mrMbX8TJyz1pxq | 1-4     | Warrior (left out) | off hand |    38 | 1,212 | 1.45 |
| nCDZQNWKX3twHhyY | 1-8, 10 | Rogue              | off hand |    28 | 1,089 | 1.19 |
| 7qPHR6cLn3zGptvN | 1-3     | Rogue              | off hand |    26 | 1,052 | 1.14 |
| ApJhHfRkMDqFj3rb | 4-5     | Rogue              | off hand |    22 |   929 | 1.09 |
| dJa2G7PHZg9wt6QX | 1-4     | Rogue              | off hand |    11 |   604 | 0.84 |
| DmjBrYd7M9VZwCHF | 22      | Rogue              | off hand |    19 |   520 | 1.69 |
| 6kN78G2xfqy3vwTP | 6       | Rogue              | off hand |     1 |    57 | 0.80 |

## Kang the Decapitator (2291)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 17153
- Rate: 0.95 PPM, measured

1,200 procs on 21,014 rolling hits, 17 wearers, 22 runs, in the main hand: all hits 0.95 [0.90, 1.01]. The rate is all
hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 0.88 [0.73, 1.05],
274 procs on 5,195 lone units, 17 wearers. Specials: 0.68 [0.00, 0.88], 6 procs on 146 lone units, 3 wearers. Two-roll
abilities, priced as one roll: 2.73, 9 procs on 55 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th
percentile 1 ms, 253 isolated hits. The paladin check, which sets no rate: lone white swings 0.90, 34 procs on 628 lone
units, 6 wearers. By report date: before 2026-11-18 0.94 (670 procs); from 2026-11-18 to push 114313 0.96 (530 procs).
Reports from 2026-11-12 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      15 | 1,067 | 18,766 | 0.95 |
| Warrior |       1 |   116 |  1,950 | 0.99 |
| Hunter  |       1 |    17 |    298 | 0.95 |

Reports and fights: 2FkYRHB1dWhra8vt 1-17; 38kdWHmLZKVgTPqb 2; 6kf3tMF4wCPNcRnh 1-33; 6wxzV4YyNCh1aqfd 1-9;
9Z473MfXLYqcyWPg 1-2; aLw2pgPCyFKTVGYm 1-5, 7; C8Qd1LXMrGT62kpY 8-9; CNqm64YkTthB7V2g 1; DL9nAWQwcN3rvZ4K 1;
F6AaCcJQvdP1nqRN 1-4; FMaAjHcwCQLkpV32 1-4; gALWqjdyncf8xVPa 1-2, 4-7; QA7aBZpK8hGJn9rf 1; qR2rnzfgPakD879W 1-2;
r9Yj7KWkaqL4B3mp 1-12; RfT7K4v2jhY9FAzD 1-3; RpFQn1ftYCAgMW47 1-17; tBk9vLmJgqwX84ar 1-7; W78tmfknBKz46hRZ 1-2;
xcPjNLDXa9nz1Vrd 1-7; YpxNGMXcf4PyTwt6 1-6; ZNyPLR2vkjMGptCJ 1-2.

## Killmaim (13016)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 13318
- Rate: 0.95 PPM, measured

872 procs on 17,458 rolling hits, 39 wearers, 43 runs, in the main hand: all hits 0.94 [0.87, 1.00]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 0.82 [0.72, 0.92], 210 procs on 4,820 lone units, 37
wearers. Specials: 1.08 [0.82, 1.29], 34 procs on 590 lone units, 12 wearers. Two-roll abilities, priced as one roll:
0.96, 6 procs on 117 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 766 ms, 160 isolated hits.
Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      28 |   607 | 11,657 | 0.98 |
| Warrior |      10 |   259 |  5,604 | 0.87 |
| Hunter  |       2 |     6 |    197 | 0.57 |

Reports and fights: 1g7hAY6CBNtW84HX 1-2; 29A3dKwPgpRNnWvc 7-11; 2aVnTxFzfckdXpBy 1-7; 3vDf6TMYBX9LK7bz 1-2;
3xgPkwfn8yQ9BqAW 1-4; 6hrcXFkgDamVB8jT 1-3; 6v4wTWQkjb3Xd7mM 6; 7hNgRwMfkQZ2Yv4b 1-3; 8GW16QY9BmFHJXLb 1;
8RFv6Xyr43dLpJ7h 1-2; DaXWGJMR2bnc9YtV 1-2; Dx2tAm7YawZjV6hG 1-4; G3VKdgR2aLAkNqWy 1-3; M4THD3Qq9dBpbKvg 1-2;
MrW6JCFZTDzdtYpH 1-4; P16dx3VLmRyaDJZQ 1-18; P7cV2HjLgxGpmZQk 1-7; PN79G4Lfv3KBrW8j 22-323, 325-361; QhYmjJ4bHZXcgfk1
1-2, 4-27; QqKcMJwHjxNPkgRT 1-3; T6MXt73janNJgmyY 1-3; Vvfz8wYbjTBFaxRq 1; XVJtpHkwZy4ancAQ 1-7; aQwNVCPHWmzDLXqk 1-2;
acHy1fNktP2m47pd 1-2; bdL8j9KGMR6r2xmk 1-3; d7DFtXL2jfNa8Bw4 1-5; fn2Dc941K63CNvmj 3-4; ftNBwvGgPcMqh8dJ 1-2;
gdKNPjVvcQt2LJnq 1-3; ghzjGKVBvJbRZnMr 1-2; grFA3BHxtCNyzR7v 1-2; hbW1zgMXyRv8JAVL 1-6; jKJPF16ZnWLkrR2d 1-2;
nT2Wz7mVQZbtBRG4 1-4; nyhMJLqBYxPpRaXf 1-2; pDRxq6wJynmhgdjL 1-9; pqr4HxC3vWJhfG7B 13-20; tMxrNWC6qPVzgwAv 7-8;
vJGT19YL26XFQpNB 1-2; zJ6C2148dTYfaXHb 2-4; zhj69Dt2X1b3fvMT 1.

## Linken's Sword of Mastery (11902)

- Proc: weapon proc 18089, logged as 18089; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.95 PPM [1.78, 2.14], 4 wearers, 4 runs, 431 procs on 7368 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights             | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------ | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 17ymZpJ6xFPf4Raj | 1                  | Rogue   | mh   | timing + otsu for 107 swings |    98 | 1718 | 1.90 |          |
| CMGKZHRVPwzvFmqy | 1, 3-5             | Warrior | mh   |                              |   105 | 1673 | 2.09 |          |
| DNfvpnLRcPJyMKQA | 1                  | Paladin | mh   |                              |    82 | 1579 | 1.73 |          |
| vwYj6cK8Xt24QLaN | 1-22, 24-26, 29-32 | Rogue   | mh   | timing + otsu for 117 swings |   146 | 2397 | 2.03 |          |

## Loksey's Training Stick (7710)

- Count: §Count of the 2026-11-09 to 2026-11-19 reports
- Proc: weapon proc 1293482
- Rate: 1.9 PPM, measured

3,520 procs on 35,405 rolling hits, 189 wearers, 235 runs, in the main hand: all hits 1.92 [1.86, 1.99]. The rate is all
hits, which the specials rate agrees with. The one-hand white rate reads above all hits, and the two intervals do not
overlap; the owner keeps the row at 1.9 (2026-11-19). One-hand white: 2.13 [2.02, 2.26], 1,102 procs on 10,013 lone
units, 177 wearers. Specials: 1.94 [1.62, 2.18], 180 procs on 1,798 lone units, 99 wearers. Two-roll abilities, priced
as one roll: 4.14, 68 procs on 318 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 1,169
isolated hits. Left out: 178 of 415 runs in Cat or Bear Form. Reports from 2026-11-09 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      26 | 1,529 | 15,616 | 1.90 |
| Warrior |       8 |   981 |  9,950 | 1.91 |
| Hunter  |     152 | 1,005 |  9,791 | 1.99 |
| Mage    |       4 |     4 |     43 | 1.80 |
| Warlock |       1 |     1 |      5 | 3.87 |
| Priest  |       2 |     0 |      0 |    - |

Reports and fights: 1QXAJwhcYL6M7TDk 8-30, 32-33; 1YLbjwRVPWr96gzd 1-8; 1aBhkzdDMZKGtb37 198 fights from 162 to 376;
1g2QAZPJrLTjCz3m 10, 12-14; 1hpF2AyzjRdrL78N 1, 3-12; 1xqBcj4CZQXV38Pv 1-2; 1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 1-3,
6-9, 11-15, 19-28, 30; 21gWyqM8QBJwrH4z 2-6; 27jZGhfHgp1LR8MB 1-11; 2PpAjf1qRmTv6gDz 1-17; 2TNQqFYdjRLB4Zga 1-10;
2WH4jLt3FADy8fnh 1-2, 5-6; 2aG3Ld8KVARCzrh1 1-7; 2vMwVxLqdB3GRZn6 1-4, 6-7; 39BFWKC2J8RTzyMx 1-3; 3HJrg1BGh8xR6Y2z 1-13;
3PXfhvmTp6NQdwRb 10-12; 3TgpnwxMD71qfLm4 1; 4ARKtCTVhrgBwadG 7-10; 4BhykDjm2p1ncLGK 10-11; 4JYa2mQgpVNMF89R 10-21,
23-30, 32-82, 84-103, 105-110; 4V2RLGtF9anbCKWD 1-23; 4wHnDaRqY9Z2zv1F 1-14, 16; 6NrFMaRQ9yAjxWC2 2, 4-12;
6WG74An9RrCdqkwD 3-8; 6wmLW1N4tBXpTVqy 1-3; 7Vywm9BAZgcMfbT3 14; 7d6p4TBxcjFyPNVr 1-4, 6; 7fgdkyv1njDKa8Gc 6;
7g9DpfNFBA3PatdZ 8-14; 7nCPmLqHd8jMfQk3 4-5, 7-10, 12-13; 7tL3AyRHfawBJF94 1-12; 8VQDf7JtwbTknMLz 6-8; 8X2Kc46faTnxVBLm
3; 8ZdnfAT79zPwmyVM 1, 3-6; 8frHF64GzAhBxDyb 13-14, 16-22; 8tgNprk9LJwW47MK 1-5; 91XtwNrLRzvMkQ62 1-3; 91ZzNbJMRtjrLW6K
1-2; 97FVRX1WLHqDmjAQ 5-12; 9TrdzatyJbQV3vmc 1-7; 9g7NQBqDH6FGfP8L 1-4; 9nxtzJ1bC42FrgWN 1-3; 9rPx2JNYLbCy1waX 11-15;
A8NWyY7dhpLDt3zX 12-30, 32-38; A8tQY9p3Dzn6TLRJ 1; BHk3rNLb9PCd4pYJ 1-6; BTRMkxmvfrwcZNaG 3; Bmk9chwYjd2F3zpZ 1, 3-6;
BydpaqJVW29G8QAc 1-9; C97aQyTJ4fPcrmjt 1; CMxBvHAN37YRqbGg 5-6; CPNY7nxkfXpvQgH1 1-3; CXvjmgaJ6YqKzD4k 1-2;
Ccvk1yDbBzFZAVqG 1-2, 5-9; CpDKr6AWJ4fw7xhg 5-7; D6bxHwnpyvc2VNR7 1, 3-8, 10; DWJhjnmAM6GZdbKv 1-2; DcgdHfFZjb87qRJa
1-4, 6-9, 11-12; FDybHvQ3pPqBracN 1-4; FQnMCW8kRqVv4mT2 8-13; FTXVdBrHJZD9aw7P 1-7, 9-23; FzGMchfD4PkdmJT1 1-4;
G6XMnQyqxTdJa2mb 4-15, 17-18; GAdW4fkq7bx9BPmL 1; GBDAWjKNh6Q9XnF7 1-4; GJgyrn6t1vABDQaW 1-3; GKZJBtWAvfPL81bc 1;
GrW7bVHhKwJgF4QT 7-9, 11-12; Gv6NDzAmFjC4P3qL 1-15; H6h87LGmw3gKRc9f 1; HL9zDvk2wWTXcP3N 1-6; HQTANmjP8FB29vYW 1-2;
Hp9ThMznR74PwYKb 5-8; J1QrDzNc4ydjV8ak 1-10; JZv9XDFLC1yKxrPY 5-10; JcwBmYGVgnMQH9xd 14-19; JtrVj2cqL3znKNXh 1;
KLNhgMTDAV6xY1Wj 1-3; KapmH2bcWfPqnrY7 1-8; KmjAV91JBhgQLrfP 3-4; L6jJNDYaATph4cfK 1-3; LJMmvtPTzdCpG9k6 5-9;
M4G1VJQdXbvyYLBD 1-2; MVcaxDQwWLpnK6HJ 1, 3-9, 11-12; MgcQ6txHr9R1VjPv 1-5; MpjtZc2zqHaR9hC3 1-19; MyzjakDC9gBbpq6K 2,
4; MzQJ9fgkwjhW1Lv6 12-14; NHvp73cXPQRC8agz 7-11; NJ7vtRpCzDTfk2W8 4; NWvgFp3kfR2bHyYa 1-35; Nc2grh6ByLvJCkRY 4-6;
NkTjKdbwYAtfJQBn 1-4; NkXjqd43AFVYpzGK 5-8; NkyKCr2XmgLt1YWv 83-87; NzgdqKHnfLJWkFVM 1, 3-12; PAmg9FX8fr23xHt4 1-5;
PDT2yACNzd4LMZ98 1-14; PLGKvft3bq8WgRYN 1-2, 4; PQDRkypFY94rzKwv 1-4, 6-7; PWhMrYaD4Aw8HcCL 1, 3-7; PmR8Lpvazd4q61nA
6-12; Pn2LYHXpjZMDywgA 7-9; PqBvJ94Nhb7ax1Vz 4; Q1CHmPvpYaxM49qR 4; Q1arfj4WPBNXqmLk 13; Q2xTrDcZqjB9pA7F 9-18;
Q8GxhTmH9f6ajcNr 6-10, 12-39, 41, 43-51; QDadybK3XVPRvwLh 30; QJafvkCx4yhVXBmT 1-2; QPmw1C9MNc4hqH3K 1-5;
QYTW7Hza9nBjJFwP 1-3; QnjR1t9Y3WHpG6Vd 1-2; Qq6CvRP12rAGwngY 1-2; QtB8qHZ4X2pyrDPY 1-10, 12-13; R6xvn3yVf4DGJpka 1-3;
RNpVm3CcDJbxjB6k 11-21; RZhVjDGx1vLPkAag 1; RmTM8XL4aBJY12Ak 11-16; T6F89dBtbNG3MjpZ 1-3; T9cJG8bq1HaNyLQC 1-3, 5-11;
TQAKbWq7vVJHLgBR 6-18; TnLAmyRdXvz1cBhC 1-18; TvMmZw8nc4hKjd1x 12-13; V39PRKTHt8kahfxX 1-6; V4HDKvxmYCkR1JnB 1-3;
V4aAZ2PGgkHw6mvB 2-3; VF1b79fRAWhtGrwa 1-3; VZdxq6yNBrW48ztG 1-3; Vb3ZxfXtyDWr1ndP 1; VmQK1zM4hFHBxXgJ 6-15;
WYrCyJkdfMXw84Ph 13-21, 23-35, 37-39, 41-48; Wf26RdmcnLDTXwV3 1-2; WgtPVTmY4XQ2DrGB 26-33; WqKd6nTV1GX8Qka4 5-10;
Wr2mNV3McRqf1Zwg 1-6; YH3c4zvFrpLtR1M7 5-9; YJR78AMPDNKGQWjF 12; YLHkP9qZWz2b4AJt 1-3; YZyAfq6XN9k4a8md 12-15;
ZNQyKft4zAGCgcDY 1-8; ZjdqP6bg7tTR894y 1, 3-5; a6QpkvcR1YHLBqPh 1-6; a8bQ4fNyGVnFcZAH 1-10, 12-25; aDKxMy2WYAfcnvGh 1-2;
aGWKDzhkCNJFXZgt 1-5; aj46hBQX1AfRFHtm 1-15; ajLmndyrzNJfwGCV 11-12; amTJhvjMAyCLdbXR 1-2; bCzt1Fq24mYcPXkB 1-2;
bT3LyRBtax8FQnch 1; bvqJAwG8D4KycjhC 1-18; byYP67TH4KtZqm31 11-13, 15-17; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2;
crtFfPbdavN6QHnX 1-9; cwtKD6v8xmJG3jNY 1-20; d37CTyKPbHaV1Yzc 1-15; dBPjxVMQmgZ9h2vr 10-26; dDKQgLW3MFPatynJ 1-25;
dnvYkhpKZQaqM2Hw 1, 3-5, 7; f3HYhjJkZdtDgzM6 1-5; fAXbZvgpmLFxWw4z 1-6; fd4MgwavrqWcYHZG 1-14; fjqQNyMbD1Y7R92h 16-21;
fnYThW3BpxKZv7jw 1-18; ftkqY3cwR7dBZK1J 1, 3-5; fzKwkZLcbYraJdF1 6-9, 11-16; g7Bh9F2b1kz6Cnar 1-2; g7RXrQTKBCNMxqJc 1;
g8BWDmknXJpvFbdf 1-4; gAdCTkV1LBxNYFca 1-20, 22-24, 26-32; gqFJrKhdG4QvaP3x 1, 3-7; hBr7xLW6nkVdMKF2 1-2, 4, 6-7;
hCzY42vqXrc1mG7w 1-4; hrFwMYgVGq41fNjQ 1-2; j2RNAQYzZK6DyT17 1; j7V46tC2GkwNYTrz 1-2; jGL3RyT4fYMtwPrn 12-14;
jGhpmnHrtyc1VD8W 1-4; jXLYwBW7bZ3nQdfJ 3-4; kH2bjrXh4BF39GmC 1-2; kQhZqY4DwGBKLdag 1-4; kjhVWgA3Gw7ypLtn 2;
kqVZPMHyFB4m9pK7 1-2, 4-5; mDbPCdX9AvwYJxGa 16-23; mH68YTDck3dKV4fx 1-2; mqjv1nLcRh6289yk 1-9; mqnZ7bkf3v86prGV 4-13;
n7qPAd4rN1JyHfpR 1-3, 5-9; nGctFJyB8jwX9WHL 1; nKhkfVbmGcg38CP6 1-2; nZ3NhbvpVXg649Bm 1-10; naCtKF4L3qzpGPmf 4-5;
p1zYDag2NHXPyRt3 1-4; p64RG9PVZcDg1XCk 4-6; pJ4AKDqYPvfwbTkM 5-8; pqfK2AgHZ7DnhvPF 1; pxAgrJDCXHc9tbB6 1-19;
qGywkYrdHKT8NgAQ 1-2, 4-8; qMLQt41BXrpR9hkA 1-4; qVtTPaDMHfhmG9gr 1-10; qj6gnCAM8kB2RNPy 1-2, 4; qzv6W9fkDaP1MZYH 1-6;
r1xm2Tt3bjgBahZF 1-17; rD8m4JW1bVyYPKGX 1-3, 6; rPAKqpWdFTR6atmN 1-2, 4; tBQwPjhFqYAyCv1L 1-16; tHz9M3arKp6PxA1W 1-14;
tLkym2AVWdJ7bM9H 1-8; tXFAxTLmg3bBNZrv 12-13; taFcCyAx9BD1KWXp 1-16; tkQKL9bGT6JVR3na 5-7, 9-16; twRbTJHfrDXBLj97 1-5;
v9Hf4bPtGq1LVykC 1-5, 7-8; vfFGXn4mbLAqkhWR 14-17, 19-20; vnrGwA8Z2RYtPc9g 1-4; wXNczvgHAqT7L8nV 1-13; xQ2J8AbFN19pPD6X
1-15; xRTpHyY7qC1GmghM 1-3; xYTCXa9myPJr6R8M 1-13; xk367aQTJVzfrKXq 1; xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6;
y9hXfMV6mWvNtnQC 1; yKvCQdzpgqY1WwbA 1-7; yR7t4NnKzkgph1bm 1-3; ythJ7TZDXk6AjHMq 9-10; yzFXjCY3WVbgRPx9 1-2;
zjgkHTcZDJNv13bF 40-85, 87, 89-98, 100-104, 106-158, 160-166, 168-174.

## Lord General's Sword (11817)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 15602
- Rate: 2 PPM, measured

7,168 procs on 83,360 rolling hits, 25 wearers, 39 runs, in the main hand: all hits 1.98 [1.90, 2.04]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.94 [1.39, 2.32], 167 procs on
1,991 lone units, 4 wearers. Dual-wield white: 1.52 [1.29, 1.69], 876 procs on 13,322 lone units, 22 wearers. Specials:
2.04 [1.88, 2.16], 1,081 procs on 12,234 lone units, 25 wearers. Two-roll abilities, priced as one roll: 2.87, 113 procs
on 908 lone units. Direct: window [-20, 100] ms, median 0 ms, 98th percentile 0 ms, 1,488 isolated hits. Reports from
2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      12 | 4,716 | 53,845 | 2.02 |
| Warrior |      11 | 2,173 | 26,303 | 1.91 |
| Hunter  |       2 |   279 |  3,212 | 2.00 |

Reports and fights: 24kgnZAXbKTW89x6 1-2; 26AYbFNQzMZCmfkh 1-16; 6c8P4BLaVAj3bpK7 9; 7TXcFNRzjwHdAybn 3-19;
7dgfrZ3bvhVqxtY8 3-14; 7tbWfYqJFd4vzL3p 3-6; ARFXzbqHDpBLJdhy 1, 3-4; CDKVcP6ghZJtv8AW 3-6, 8-16; D9frqPV8dam7B1AM
35-50, 53-55, 57, 59-61; HwyJMCLKfbY1PkmG 1-8, 10-38; JfNv1pWML9ZA7ghR 1-4; K6VvbDypXFLQJW7d 1-27, 29; MgC1BXh8fyArkYqa
1-30; MwVH1KQmak73NPd8 8-16; NK6cCM4pJyRhaAFW 1-7; NXMZV9xjvWpB7GtP 10, 12-107, 109-116, 118-122; NcxbzW4kJYQ7DGhj 1-3;
QFqKVZgc6vAJfbrW 26-28, 30-31; RVkGfcrbZ13MWqQ2 1-3; T92NqcgBXaZrKvCm 1-5; TLp1jza6PCrBAX94 1-7; Wn8wHhaxy31MLzjp 1-10;
X8z9dgD6HRTZ24Vc 23; XaF296wZvArq7Qxh 2-6, 8-19, 21; fCB4wQt3NWm6k8M9 5-7; gvWZG1cjrChKwdtz 1-2; hF7dxZ8nJDpk4wNG 1,
3-5; jn3dwzWPZ6r2ky1K 1-3; pPbM8VAdmWCTvLB3 1; pQPwFg8rf6VWhRNc 1-3; pgN9QwMLYVfrqGbv 1-4; q21PmQwWRfMLvKXc 1;
qbtZdThvBCXVcfAW 1-6; rfvtaC13gNqRx4Az 1-3; v1bPdgDtXypHncfN 1-4; yPVkKpqtfJM27WFB 6-7; yjV6m7kpfgaPzXxt 5-9;
zKgGQA6BqvcJtYZ9 1-3.

## Mana Channeling Wand (18483)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1302348
- Rate: 4 PPM, measured

1,030 procs on 9,616 rolling hits, 38 wearers, 55 runs, as a wand: all hits 4.02 [3.76, 4.29]. The rate is all hits,
which the lone wand shot rate agrees with. Lone wand shots: 3.81 [3.59, 4.09], 840 procs on 8,275 lone units, 38
wearers. Direct: window [-10, 100] ms, median 10 ms, 98th percentile 22 ms, 810 isolated hits. Dropping the swapped
stretches (5 active minutes, 1% of hits, in 1 run) moved the read from 3.99 to 4.02. 1 run logged in two reports counts
once. By report date: before 2026-11-18 3.98 (423 procs); from 2026-11-18 to push 114313 3.97 (486 procs); after push
114313 4.30 (120 procs). Reports from 2026-11-11 to 2026-11-21.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Priest |      29 |   999 | 9,420 | 3.98 |
| Mage   |       9 |    30 |   196 | 5.74 |

Reports and fights: 1BLCZMxRfyYmHXNn 5-6, 8; 1VpHn3kYq6B8JMWw 1, 3-4, 6-7; 26AYbFNQzMZCmfkh 12-16; 2GczY83pjnCHdTNA 1-2;
3zwbTGcyQBLm7DMg 1-7; 4thjWV7Xbz2yrxBF 6; 6hDJQWTXwBr419AM 4; 81rnYqNT3djhJFRy 3-4; 8HvfKkArXpmbhwTF 1; 8YVKbmp1y4rMPFdW
2-11; aCR8n63FzdTYGqj7 1-7, 9-10, 13-24, 27-29, 31-36, 38-40, 42-57, 59-63, 65-67; aYBjphvLn8xMK2TN 1-5;
c7M6PR9WHkyDJL2q 1-6; cav7VZrmPbNW3y29 13; DvrjQXqZ9HN8a74z 1-4; f2CmVRcTnthPW6aH 1-14; fjtT3JVxbGPz6RwX 6;
GAaNPKjbZkwLBCT2 2; gBDMLv8J1K7nqkpd 112-157; GFBCh3VZH7TmAkL4 1-3; gw8fpa1YVkv4WP32 1, 3; hfNjty7QRBrFCKm6 1-3;
Ht8VFJxmnCG9bQhM 1-3; hYjVWRNBGDKLxTpc 1-3; JvFPcBTz9rZMmW7h 1; KNtHCAPd97ZMTLBf 1-2; Lb7Zcx2JtgjCTwP8 1;
LN2Vk4nMtBc8vWbT 1-6; m3ctFRXJaZnrKzkC 1-3; mCn63gBZfLFHaWr9 6, 8-34, 36-39, 41-43, 46-53, 55-89; mLH64Yyjpbd3AKBW 1-4;
mnPXjHcWNFKBhDCq 1-2; MPfbhA7yLxpkdtZQ 5-14; n8VK43MDGChq6rgv 3; p8WXHnkF6m4rvGJL 1; PDwtabFTz3VnQqrH 1;
pG8TNJwybLPYWmxF 2; PKBd6Vz2Gkjwp9Zt 1-2; pqxm14QnyLk9tCGK 1-2; QFqKVZgc6vAJfbrW 1-19, 21-28, 30-31; Qqf8ZC1vRYhwG2cr
4-7; QXFd74zDW8JZMmKx 1-2, 4-13; RJXQC67jbzATWdx1 1-2, 5-9, 11-16; T7dtHBv9YR2jnLwg 1-7, 9-10, 12-19; tG4CqkJQvyPT2A1Y
1-10; tHzN3kKbaPpVW8TY 1-8, 10-11; TXGK16YwMmcyj9Cd 2-3, 5-10, 12-13, 15-17, 22-29, 31, 34-74, 76-86; V9ZpFqnkLrbaT6zy
1; VvFWzAgjYPxwN8fT 32-35, 37, 39-42, 44-49, 51-52, 54-55, 58-61, 63, 65-67, 69-72, 75-79, 82-85; wkJXWxD23VMqThLg
14-17; wR8KA9WTgHkD6n7q 1-10, 12-34, 36-47, 49-62, 64-78, 80-94, 96-111, 113-128, 130-146, 148-202; Wv4QYA6C7antxKXr 1;
XD23HFnxbzvPjACV 1; zA8JyGdnWRQ9vXmK 1.

## Naga Heartpiercer (3078)

- Proc: weapon proc 1292647, logged as 1292647; rolls on ranged
- Sim rate: 2.2 PPM, measured; measured 2.22 PPM [2.05, 2.43], 13 wearers, 14 runs, 965 procs on 10882 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class  | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ------ | ---------- | ----: | ---: | ---: | -------- |
| 1WZ8Dn6mpQdCNxAT | 1-8    | Hunter | ranged |            |    71 |  594 | 2.99 |          |
| 1a3pDtCcPgBkL6wm | 1-22   | Hunter | ranged |            |   114 | 1041 | 2.74 |          |
| 1fhNLGvKgMzt8Ac3 | 1, 3   | Hunter | ranged |            |    31 |  326 | 2.38 |          |
| 1h9P4rYzWDvtHRxb | 3-4    | Hunter | ranged |            |    68 |  902 | 1.88 |          |
| 1hpgVjWH3GZBLYKw | 1-12   | Hunter | ranged |            |    85 |  912 | 2.33 |          |
| 1x8aqzFcL7ydGNgP | 1-2    | Hunter | ranged |            |    29 |  267 | 2.72 |          |
| BKnf4rNwJdcqpQ18 | 1-6    | Hunter | ranged |            |   108 | 1311 | 2.06 |          |
| BzGXfjMLJ2RyCwHm | 1-7    | Hunter | ranged |            |    85 |  891 | 2.38 |          |
| MwA4Vhx7zDfCdY6B | 1-2    | Hunter | ranged |            |    56 |  703 | 1.99 |          |
| aMNVKZpg8tRkhBP3 | 1-17   | Hunter | ranged |            |    51 |  660 | 1.93 |          |
| aykDWnCGx4jJFTV6 | 1-6    | Hunter | ranged |            |    54 |  806 | 1.67 |          |
| hPWNkH67AYCBFLgM | 1, 3-8 | Hunter | ranged |            |    55 |  652 | 2.11 |          |
| hcGCaDW9tRYFB6Ax | 1-11   | Hunter | ranged |            |   125 | 1505 | 2.08 |          |
| r7cZyhBXWfNwj2GY | 1-10   | Hunter | ranged |            |    33 |  312 | 2.64 |          |

## Night Reaver (1318)

- Proc: weapon proc 13480, logged as 13480; rolls on main hand
- Sim rate: 1.9 PPM, observed; measured 1.90 PPM [1.54, 2.40], 5 wearers, 5 runs, 218 procs on 2091 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights            | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 82YfmdLnzpAyHWwN | 1-9, 11-31, 33-50 | Paladin | mh   |            |    77 |  774 | 1.81 |          |
| MZRL4VdF6XtNf8B3 | 1-2               | Paladin | mh   |            |    32 |  420 | 1.39 |          |
| QYP1JCm2FMydzqjx | 1                 | Paladin | mh   |            |    47 |  410 | 2.08 |          |
| T72D1mwvJnabPtqK | 1-3               | Shaman  | mh   |            |    19 |  239 | 1.45 |          |
| ac64wQx7yZ9tRNfW | 1-3               | Warrior | mh   |            |    43 |  248 | 3.15 |          |

## Nightblade (1982)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 18211
- Rate: 2.1 PPM, observed

86 procs on 705 rolling hits, 2 wearers, 2 runs, in the main hand: all hits 2.09 [2.02, 2.11]. With 2 wearers the wearer
bootstrap means little; the Poisson interval on the 86 procs is [1.67, 2.58]. Neither the one-hand white swings nor the
specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: fewer than 5 wearers. One-hand white:
2.17 [1.77, 3.96], 18 procs on 142 lone units, 2 wearers. Specials: 1.73 [0.00, 1.84], 12 procs on 119 lone units, 2
wearers. Two-roll abilities, priced as one roll: 5.71, 1 proc on 3 lone units. Travel time: the proc lands a median 137
ms after its hit, 98th percentile 254 ms, window [41, 264] ms (34 isolated hits). Reports from 2026-11-14 to 2026-11-15.

| Class   | Wearers | Procs | Hits |  PPM |
| ------- | ------: | ----: | ---: | ---: |
| Hunter  |       1 |    72 |  586 | 2.11 |
| Warrior |       1 |    14 |  119 | 2.02 |

Reports and fights: nNmBkKp3J96FTGfg 1; vdwBDjmZRykqHaWL 7.

## Pendulum of Doom (9425)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 10373
- Rate: 0.9 PPM, measured

641 procs on 10,733 rolling hits, 8 wearers, 10 runs, in the main hand: all hits 0.90 [0.83, 0.93]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.02 [0.81, 1.16], 158 procs on 2,333 lone units, 8
wearers. Specials: 1.00 [0.60, 1.10], 37 procs on 553 lone units, 3 wearers. Two-roll abilities, priced as one roll:
2.34, 12 procs on 77 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 178 isolated hits.
Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   393 | 6,633 | 0.89 |
| Shaman  |       5 |   248 | 4,100 | 0.91 |

Reports and fights: 9pMcvgxrQNfLZ7hd 1-15; AdD17hyxFtqfrQmC 1-4; Pjz8kK2XvxyQrTCB 4-16; WBfhVMYkvLHmPQjx 6;
X186Gx2gRBZChbVA 1-7; ga9Qk6DcjJmNTGVH 2-3; qw3ZzkBmQYW7pXVH 1-19; t6dr1MY8zTJBWnkg 1-5, 7; wTakYgvxcBQV9qRr 1-19;
zc1hygaCXjDAPxL6 1-2.

## Phantom Blade (7961)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 9806
- Rate: 2.05 PPM, observed

780 procs on 14,264 rolling hits, 16 wearers, 17 runs, in the main hand or the off hand: all hits 2.05 [1.88, 2.23].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 2.24, 25 procs on 419 lone units, 1 wearer. Dual-wield white:
1.97 [1.75, 2.19], 284 procs on 5,413 lone units, 15 wearers. Specials: 2.50, 8 procs on 120 lone units, 1 wearer.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 219 isolated hits. Reports from 2026-11-15 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      12 |   541 | 10,224 | 1.98 |
| Hunter  |       3 |   170 |  2,792 | 2.28 |
| Warrior |       1 |    69 |  1,248 | 2.07 |

Reports and fights: 2YfdbFMgGhr7zHwQ 3-4; 6kN78G2xfqy3vwTP 1-6; 8L9G1c3bKt6BAxJW 1-10, 12; Ap2fHzmtJLhv83DN 2-15;
D2d3hNAbwkPgL1vF 1-10; GzZH1WhTLQ3JaDXt 1-11; QAHXb9zy4jfgKPVJ 1-6; XFdcWaPL8MthYH6A 1-5; cRtHw9DWZMrKnyvQ 1-4;
dt4M3ypRnTkwaVQY 1, 3-10; hrV3QgyzvfdZKJwF 1-30; kQ2gDVr4h7NPm1pa 1; pLtJjx8DCPbHRacv 1-21; twcjmZNGMQP2FABR 1-12;
y2dFcNR1rMmKTPAj 1-25; ynGzYZ3cLrpvwj1N 1-9; yqFPH1MKv746dQhB 4-7.

## Poison-tipped Bone Spear (1726)

- Proc: weapon proc 16401, logged as 16401; rolls on main hand
- Sim rate: 0.85 PPM, observed; measured 0.87 PPM [0.72, 1.11], 4 wearers, 5 runs, 117 procs on 3517 hits
- Reports from 2026-11-15 to 2026-11-16
- Notes: 2 of 7 runs are feral druids, left out (paw speed, not the weapon); few wearers (4): the interval is about the per-wearer range

| Report           | Fights                    | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| GgmB9Yhp8fMPWjVN | 3-11, 13-16, 18-22, 24-30 | Hunter  | mh   |            |     6 |  186 | 0.84 |          |
| MzwqjvXHFKZap19g | 1, 4                      | Hunter  | mh   |            |    32 |  869 | 0.96 |          |
| TCk3ZaghzP1wJdXK | 1                         | Hunter  | mh   |            |    28 | 1030 | 0.71 |          |
| acpjMDfJkLPABvxT | 1                         | Paladin | mh   |            |    27 |  593 | 1.19 |          |
| rwxQgM38abAcmRVv | 1-6, 8-9                  | Hunter  | mh   |            |    24 |  839 | 0.75 |          |
| bpM8VcnfdmyxZ1vg | 23                        | Druid   | mh   |            |    17 | 1206 | 0.37 | feral    |
| hPWNkH67AYCBFLgM | 1, 3-8                    | Druid   | mh   |            |    14 | 1073 | 0.34 | feral    |

## Puncturing Spear (285332)

- Proc: weapon proc 1318250, logged as 1318250; rolls on main hand
- Sim rate: 2.45 PPM, observed; measured 2.46 PPM [0.92, 2.91], 3 wearers, 3 runs, 117 procs on 920 hits
- Reports from 2026-11-07 to 2026-11-15
- Notes: 3 of 6 runs are feral druids, left out (paw speed, not the weapon); few wearers (3): the interval is about the per-wearer range

| Report           | Fights   | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1ZJVMgcYWTyRQh2w | 1-3      | Paladin | mh   |            |    12 |  134 | 1.73 |          |
| aykDWnCGx4jJFTV6 | 1-6      | Hunter  | mh   |            |     1 |   21 | 0.92 |          |
| hxYkG4fKMrPmjRCD | 1-8      | Paladin | mh   |            |   104 |  765 | 2.63 |          |
| Ybjk79mPGZWrq2xy | 1-2      | Druid   | mh   |            |    23 | 1037 | 0.43 | feral    |
| ZFBD3y2d8cfP46Ja | 1-2      | Druid   | mh   |            |    38 | 1056 | 0.70 | feral    |
| zyALHJFhZtYq2Gvn | 1-3, 5-7 | Druid   | mh   |            |    67 | 1890 | 0.69 | feral    |

## Pyromancer's Baton (286650)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1320925
- Rate: 1.95 PPM, measured

964 procs on 12,402 rolling hits, 16 wearers, 17 runs, in the main hand or the off hand: all hits 1.94 [1.82, 2.07]. The
rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.85 [1.50, 2.26], 98
procs on 1,322 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 1.43 [1.04, 1.81], 147
procs on 2,562 lone units, 10 wearers. Specials: 2.17 [1.80, 2.54], 114 procs on 1,316 lone units, 7 wearers. Two-roll
abilities, priced as one roll: 1.92, 1 proc on 13 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 275 isolated hits. Dropping the swapped stretches (24 active minutes, 4% of hits, in 3 runs) moved the
read from 1.87 to 1.94. Left out: 1 of 18 runs in Cat or Bear Form. By report date: before 2026-11-18 1.94 (725 procs);
from 2026-11-18 to push 114313 1.96 (239 procs). Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   526 | 6,578 | 2.00 |
| Warrior |       6 |   382 | 5,029 | 1.90 |
| Shaman  |       1 |    56 |   795 | 1.76 |

Reports and fights: 4aZmt3y9znXjhRvV 1-2; 4NMcqLwQ2rB9amCH 1-3; 9hywC8jMTLV36Pxd 4-8; 9rPx2JNYLbCy1waX 1-9, 11-15;
ARqz9awMbfr6QD7Y 1-2; Brj9PhXHQM3kLN2f 1-4; HqwWnYGKdhpCgZVR 1-12; PamDnk2rVQhW7Cj6 1-5; Pw9kWLGcAj3VpmCn 13-18, 20-22;
q6VRad2JDnpxXYAL 1-13; QPmw1C9MNc4hqH3K 1-5; ryqgQphLxMBz4TmH 1-4; Tbh1yKNfPtqLYCak 1, 3; WhJ7vPzQnbFkmKyj 1-12;
Wv3xD7KBm1C2tXNL 1-8, 10, 12-19; ZAjYLxr6TB7KGmav 1-9; zHYaAjmvQBT2fFW7 4-15.

## Quillshooter (10567)

- Proc: weapon proc 29646, logged as 29646; rolls on ranged
- Sim rate: 2.9 PPM, measured; measured 2.91 PPM [2.62, 3.17], 6 wearers, 6 runs, 465 procs on 3421 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights       | Class  | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------------ | ------ | ------ | ---------- | ----: | ---: | ---: | -------- |
| 1QgdaxJ2GBpmvyjL | 1-2, 4-8     | Hunter | ranged |            |    68 |  509 | 2.86 |          |
| 4bqYVtkjL6aKdwQM | 1-2          | Hunter | ranged |            |    63 |  472 | 2.86 |          |
| ByRxHQmnzfhbvtcV | 8-24         | Hunter | ranged |            |   123 |  813 | 3.24 |          |
| CpABMyt29JkdwjZf | 1-20         | Hunter | ranged |            |    72 |  503 | 3.07 |          |
| Hrt2wgZkdjTRBC4A | 12-13, 15-20 | Hunter | ranged |            |    86 |  771 | 2.39 |          |
| MfXjTPdwar9gyKvW | 1-2          | Hunter | ranged |            |    53 |  353 | 3.22 |          |

## Rivenspike (13286)

- Proc: weapon proc 17315, logged as 17315; rolls on main hand
- Sim rate: 1.6 PPM, observed; measured 1.60 PPM [1.20, 2.25], 9 wearers, 9 runs, 1419 procs on 18330 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split                    | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----------------------------- | ----: | ---: | ---: | -------- |
| 3RB9X8HVxgCJ4My2 | 1-4         | Rogue   | mh   | timing + otsu for 9 swings    |    76 |  679 | 2.32 |          |
| AdfMXDLW29k6pGy1 | 1-3         | Warrior | mh   | timing + otsu for 14 swings   |    82 |  867 | 1.96 |          |
| AyTmpNtxB9jbnJ2H | 10-45       | Rogue   | mh   | timing + otsu for 1055 swings |   378 | 3163 | 2.47 |          |
| CfHvwytKJ6P4q7ag | 1-24        | Warrior | mh   | timing + otsu for 15 swings   |    95 | 1145 | 1.72 |          |
| CfHvwytKJ6P4q7ag | 21-24       | Rogue   | mh   | timing + otsu for 38 swings   |   104 | 1045 | 2.06 |          |
| CqM4Vnh831tFgcYw | 1-2         | Rogue   | mh   | timing + otsu for 80 swings   |    63 | 1268 | 1.03 |          |
| DWfMg7hzPF4d1mQ6 | 1-11, 13-52 | Warrior | mh   | timing + otsu for 224 swings  |   396 | 7863 | 1.04 |          |
| kaH1G4ztmYpgMTJ6 | 1-2         | Rogue   | mh   | timing + otsu for 34 swings   |    51 |  766 | 1.38 |          |
| vny4RK9TkMzwtAQV | 1-5         | Rogue   | mh   | timing + otsu for 131 swings  |   174 | 1533 | 2.35 |          |

## Sang'thraze the Deflector (9379)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1294243
- Rate: 2.05 PPM, measured

2,246 procs on 38,451 rolling hits, 29 wearers, 33 runs, in the main hand or the off hand: all hits 2.06 [1.97, 2.17].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.03 [1.89, 2.20],
653 procs on 11,343 lone units, 17 wearers. Dual-wield white: 1.69 [1.43, 1.90], 175 procs on 3,649 lone units, 12
wearers. Specials: 2.13 [1.91, 2.37], 212 procs on 3,521 lone units, 18 wearers. Direct: window [-19, 100] ms, median 0
ms, 98th percentile 0 ms, 723 isolated hits. Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      18 | 1,898 | 32,465 | 2.06 |
| Rogue   |       8 |   293 |  5,011 | 2.06 |
| Hunter  |       3 |    55 |    975 | 1.99 |

Reports and fights: 1CYAWTtmcGdFqbPN 1-5; 28AhM3VGnvFcJ4rD 1-4; 4NqLGxvrZYX2Qwnf 1-4; 6bM1k938LGNghacx 1-9;
8FBqfvLyWTdKxbZh 1-38; 8KPHjdAxZYFNLtgR 1-5; 987q1pjnYdVQbCmT 6; Axv7w1D9JW8ykLhY 1-9, 11; C3zW7BjkJXgRtq8m 1-11, 13-18;
FAL4Hv1D7nWf8VNh 1, 3-8; FHmqabcjK2RAM9vZ 1-3; FVv7P3a2KqJrdG4g 1-3; JDZFAK1z7QbckdHN 1-3; KfvWRDrFXQ1TMq9P 1-8;
LXWzFKcBmvyMZaH2 1-2, 4-9; NbwyAxHKzhYXV6L2 1; VQCRJWfq3KX7M1Nh 1-2; XCJdxAF17zqHMhLY 1-2; Xj23FkTPgrHawv7J 1;
ZLpHRCgGFDBk9TPf 1-16; ZLpcV6krn1wYHJvt 1-8; abWvgzBMf2rJXhcL 1-2; bXza3gZq7nRTGVfW 1-4; brf4Pzj8BkyMmCHJ 1;
cPLzXv4nTtNaYWF3 4-20; hgmcV2xfkQFbWMN8 1-3; qVJHr3mzbkAvxgyd 1; tCprA1RmvTcJkq3D 1-4; tnyh4dxq6pNXQzLw 1-3;
w63W4VhTgCbt9knB 1, 3-10; yQnph9W1VB4z6a2J 1; z2QFnM7tAjJmqgvw 1-4; zLacx86Thr3AMtqn 1-15.

## Satyr's Lash (17752)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1292155
- Rate: 1.95 PPM, observed

1,105 procs on 19,794 rolling hits, 27 wearers, 28 runs, in the main hand or the off hand: all hits 1.97 [1.80, 2.12].
No lone-unit rate applies, since its travel reaches past the isolation: the rate is all hits. Observed: no lone-unit
rate checks it. One-hand white: none. Specials: none. Travel time: the proc lands a median 133 ms after its hit, 98th
percentile 265 ms, which reaches the 300 ms isolation, so no window was measured (239 isolated procs). Reports from
2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      20 |   854 | 15,466 | 1.95 |
| Warrior |       2 |   203 |  3,500 | 2.05 |
| Hunter  |       5 |    48 |    828 | 2.05 |

Reports and fights: 4NqLGxvrZYX2Qwnf 5-8; 4PtKGL1zcRV8Chfn 3; 4YzXaGtg81WdqM2N 7-15; 6zJNTjMwWD31vrgy 1;
96tBVjg8RYJNMa7r 1; Bfk6ChZgAr8yWtXv 1-7; C8Qd1LXMrGT62kpY 8-9; L2c7jBZNwFPpb1tV 1-4; Lg82ZFDyQ4MvRm9p 14-26;
MKmntyr82Q4v7jDh 1-5; R1acrQhyNvwXLmBj 1-3; Rc4dCWA6N8PwpBvT 1-8, 10-11; T71XnfaF3cY9M2Rg 1-13; ThKD4jr8MfawcCPv 1-19;
V7Xna6bcWLRzZ3GN 1-6, 8-11; W1zqgCZVPAtTM27b 1-5; YamkzNdPTZFLAnGq 1-8; Yf91cBLN26jdqrxC 1-4; ZwxX9G3bRPFYfjng 1, 3-4;
cKyZr8MFNHBT7w9t 1-5; d4LFCkmQYhPRDx2y 1-6; ghWq2C84BJj9bQHd 2-4; hyCdAYx8GjgKfZHp 1-5; qm1Bvy29hGK6bMaQ 1-2;
rgKPxdF6G3NbMzmj 1-3; tCprA1RmvTcJkq3D 4; tdYLMmfgcAJCTWyr 8, 10-12; yJmtnzPNDVK42TQR 5-8.

## Scented Runewood Brooch (269741)

- Proc: equip proc 1318323, logged as 1318324; rolls on harmful spells, melee abilities, melee swings, ranged abilities, ranged auto attacks
- Only against Furbolg targets (its tooltip); counted on hits against Blackmaw creatures, melee, ranged and spell hits alike
- Sim rate: 1.2 PPM, observed; measured 1.13 PPM [0.79, 2.17], 5 wearers, 5 runs, 235 procs on 4140 hits
- Spell hits are priced on the main-hand weapon's speed by the owner's ruling, in the table as in the sim: the Mage's and
  the Priest's rows read 3.0 s and 3.3 s a hit, their staves' speeds. These logs do not settle how a spell hit is priced
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights               | Class   | Slot    | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | -------------------- | ------- | ------- | ---------- | ----: | ---: | ---: | -------- |
| Hwq97xLh6gv8ZPrY | all in Blackmaw Hold | Paladin | trinket |            |    61 | 1027 | 1.11 |          |
| MR8g1PhjbDN3tpkv | all in Blackmaw Hold | Priest  | trinket |            |    11 |   63 | 3.14 |          |
| Q7V24RA9HTg6F38p | all in Blackmaw Hold | Hunter  | trinket |            |    74 |  706 | 1.90 |          |
| fNMJcqDHb9ZzPLnk | all in Blackmaw Hold | Rogue   | trinket |            |    75 | 1484 | 1.22 |          |
| wa7j4vxdLRZtfDcT | all in Blackmaw Hold | Mage    | trinket |            |    14 |  225 | 1.24 |          |

## Scorpion Sting (1265)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 18208
- Rate: 0.95 PPM, measured

2,096 procs on 54,256 rolling hits, 59 wearers, 64 runs, in either hand: all hits
0.97 [0.93, 1.01]. The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white:
1.03 [0.86, 1.24], 144 procs on 3,492 lone units, 10 wearers. Dual-wield white: 0.81 [0.73, 0.92], 352 procs on 10,880
lone units, 48 wearers. Specials: 0.99 [0.90, 1.10], 306 procs on 7,693 lone units, 43 wearers. Two-roll abilities,
priced as one roll: 1.05, 15 procs on 357 lone units. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms,
559 isolated hits. Dropping the swapped stretches (81 active minutes, 3% of hits, in 5 runs) moved the read from 0.94 to
0.97. By report date: before 2026-11-18 0.97 (1,403 procs); from 2026-11-18 to push 114313 0.97 (693 procs). Reports
from 2026-11-10 to 2026-11-20. The sim does not simulate this proc yet (the generator lists it as a missing effect); the
rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      29 | 1,107 | 28,298 | 0.98 |
| Warrior |      26 |   970 | 25,418 | 0.95 |
| Hunter  |       4 |    19 |    539 | 0.88 |

Reports and fights: 1TCjB8HdypDcxbt2 1-3; 26tvRFXhzdqrLApN 1-3, 5-9; 4ta7nVwdhkXM389v 1-3; 64NVwXYTpFRBxJWz 9-11;
6NrFMaRQ9yAjxWC2 11-12; 7L198CkvgnJyKAHm 1-14; 7vxRAKHWC83tgPDz 1, 3-7; 8cj6fJ9kYvGmqx3W 1-3; AbnJ1XMLhwqtpx6F 1-4;
avrcxzCymh1DNqA8 17-19; Azq6hMZn2RyWwagv 1-6; bHkgvJcQaRpXYB3y 4-11; BM8NmpvwVdzA76f9 1-9; BMPCZ3a7y9HzJYfX 1-4;
bpM8VcnfdmyxZ1vg 10-17, 19-20, 22-23; bpV7JWX8B9xgrj3c 1-5; Br2pycwj4zLC1ZP7 1-3; CdgL43HfzxKMhtPv 1; CKLMhw2QbTVc8Hym
1-14; cPHdgRMCy7ZGDqFf 1-8; D8Jwdf1PkWm2y3YH 4-7; D9vw4QTYJZWMPnph 1-3; DKpBr8Y6RZt7yMca 1-8; Fdw3nRjaYcDbvyPC 1-9;
gk2jJvTHhW4zNDKB 1-2, 4, 7-11, 13-18; gn4HAC1RqMfLt8wN 16-21; h1TVXKmtQwWpad6C 1-17; HrJFDLXd1NVmbGv3 1-5;
j9x1hJfM43nNpzrv 1-8; kBb8HK6JDdaX2nZq 1, 3-9; KGgcP8Jxz12mNwfv 1-15; Kt8wFb3qThrVjNGY 1-5, 7, 9-11; LJMmvtPTzdCpG9k6
5-9; LMdRqJGhV17XwWyT 16-22; M4CvXLZz3qm6pDRc 2-3, 5; nBHjcV4Ydm7z2X9q 1-3; nJgV6Ma2Qhkzm8Bt 1-5; NmwLfZHTpdqPWAMC 5, 7;
Ntvp7VG8kdm6Ja4g 1, 3-18; nvcCpPrdZWGBba36 1-2; pgjc6w1CaJ472ybt 1; pLFZK3MWV2TRw6t7 1-2; pmKDwngMT2NFktdA 8;
PNnXcA9hFBx7Wtdj 9-11; PxfYmBXLqFg3GMvT 1-26; Q4GadwADtxbFBNW9 4-8; Qf2RhFMXYDwBCzHg 6-8; QqVNrzBdvxnwjGYb 1-7;
rhWaxQK8cv37nzXN 1-5, 7-9; tDraZ1AfWnGhNYvg 1-6; TFzR3fQkt2ap7K4Z 1-10; TndajWC3c76QymF4 1-5; V8bxkrXJqAK6v3cL 13-14;
Vk9QFCRcmdypPBvt 1-4; VYb6f27NazpmCch3 1-2; w4qRTLXNzP7a9cQ2 1-6; Wgk8whB6HQNZmrnq 1-5; xMw1vFZhCGJN9YQP 3;
XNry2TxRndjwJQBm 1; xwT6Z1CV7WQhBmgD 1-3; yjqVPg81hM6TCW97 1-2; ynGtWRk4cZX3TmYv 1; zbg1wFp2L8hWHZva 1-26.

## Searing Blade (12992)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1299850
- Rate: 1.9 PPM, measured

392 procs on 4,944 rolling hits, 17 wearers, 17 runs, in the main hand: all hits 1.90 [1.78, 2.06]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.89 [1.72, 2.21], 169 procs on
2,149 lone units, 17 wearers. Specials: 1.88 [1.31, 2.38], 60 procs on 766 lone units, 15 wearers. Two-roll abilities,
priced as one roll: 2.48, 6 procs on 58 lone units. Travel time: the proc lands a median 98 ms after its hit, 98th
percentile 214 ms, window [14, 224] ms (74 isolated hits). Dropping the swapped stretches (36 active minutes, 14% of
hits, in 2 runs) moved the read from 1.63 to 1.90. The paladin check, which sets no rate: lone white swings 1.78, 60
procs on 807 lone units, 8 wearers. By report date: before 2026-11-18 1.93 (336 procs); from 2026-11-18 to push 114313
1.74 (56 procs). Reports from 2026-11-13 to 2026-11-19. The sim does not simulate this proc yet (the generator lists it
as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |       9 |    29 |   301 | 2.31 |
| Warrior |       8 |   363 | 4,643 | 1.88 |

Reports and fights: 1K6c48bWMN9nzCZA 1-3; 2L1XYNpZCTR4nG9t 1-4; 3tFdrmh8baCZpNn7 1-6; 3yM21KdFP7LnrBvx 1-3;
6Tyb2X1ZNA7CkaVx 1-2; 7ZxcPNTjqvfzQHgX 1-2; GdDn73hPYt2JQmHL 7-8, 10-12; GdVxzv2hHjNtWMcn 1-20; kNYgfyZF4LV8QmXq 2, 4-5;
nG4DqxBAmQgZLf8a 1; PNYpz2hCmGMRV1tx 1-3; RQPftYBgN74bAz1q 1-5; tBnv1kDbzHK6hAwg 3-7; V4Cr3TxcZM2kDqtG 1-5;
ycJpHf8KPAVG9x4F 1; Yd4r69MDqbQyJCFH 16-36; zjcGFyxpkKTnQXdw 1-11.

## Serpent Slicer (13035)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 17511
- Rate: 1 PPM, measured

1,105 procs on 26,110 rolling hits, 24 wearers, 25 runs, in either hand: all hits
1.02 [0.94, 1.08]. The rate is all hits, which the specials rate agrees with. The one-hand white swings stay under 50
procs. One-hand white: 1.23 [1.03, 1.35], 26 procs on 506 lone units, 2 wearers. Dual-wield white: 0.91 [0.76, 1.05],
228 procs on 6,034 lone units, 22 wearers. Specials: 1.05 [0.86, 1.23], 139 procs on 3,163 lone units, 15 wearers.
Two-roll abilities, priced as one roll: 1.36, 14 procs on 247 lone units. Direct: window [-10, 100] ms, median 0 ms,
98th percentile 1 ms, 278 isolated hits. Dropping the swapped stretches (24 active minutes, 3% of hits, in 4 runs) moved
the read from 0.99 to 1.02. 1 run logged in two reports counts once. By report date: before 2026-11-18 1.05 (637 procs);
from 2026-11-18 to push 114313 0.98 (468 procs). Reports from 2026-11-13 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      11 |   633 | 14,857 | 1.02 |
| Hunter  |       7 |   160 |  3,815 | 1.01 |
| Warrior |       6 |   312 |  7,439 | 1.01 |

Reports and fights: 2MKaBZy3AJqLdjPF 1-2; 8L9G1c3bKt6BAxJW 1-10, 12; 9Jj1FDzGp7BRQ648 1-15; 9wtfbKA27FNQ6BaW 1;
C4TGJvzHF3aZ2NjR 1-13; g4Y2frPx8kBmj9Tq 1-2; Gh3vdkL7AJrm6fFj 1-11; gywYLPFQCJa9XfVG 1-6; hYHt6kLTgApm8RMw 1-2;
nfLmGxNDqJpyXB6k 24-38; NrJZ7avW13QBVTGw 1-7; nrKZtkQhgbBYf7R6 1-6, 8; PmM8nxa1p6QkhVLC 1, 3-7; rAJbdCG47n2hgwVB 1;
Td2y9DjVvn1YgaHr 1-7, 9-21; VgbFTznDNP4XdAh9 1-13; vk9HJZwrfhmCayMY 1-7, 9-10, 12-21; VKHmAfvPjBGLQW4Y 1-9;
vpTQqDkgNYLnjtZr 1-47; wnMHyXgW6KpAhYF4 1; XkhC2YtZ6cMBHGfz 1-2, 4-6; Y1vtT4JxK97fhQrL 1-3; ynGzYZ3cLrpvwj1N 1-9;
Z7qF4j3hazBTxHcw 8-10; zMqZ3BNKbFYfAjwp 1-12.

## Shadowblade (2163)

- Proc: weapon proc 18138, logged as 18138; rolls on main hand, off hand
- Sim rate: 1.05 PPM, observed; measured 1.03 PPM [0.83, 1.25], 5 wearers, 5 runs, 224 procs on 7253 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights             | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------ | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 3GT4tz8gxnphcbkD | 1-7                | Rogue   | mh   | timing + otsu for 20 swings  |    53 | 1528 | 1.16 |          |
| 3qfDVTZ1J29zPChn | 1-23               | Warrior | oh   | timing + otsu for 52 swings  |    64 | 1542 | 1.38 |          |
| 7fz9Kk8rY6jyQAbL | 6-10               | Rogue   | oh   | timing + otsu for 76 swings  |    17 |  825 | 0.69 |          |
| ATwq9YWpNbLC6Khc | 1-10, 12-17, 19-20 | Rogue   | oh   | timing + otsu for 112 swings |    23 |  879 | 0.87 |          |
| xRkgQCj3TD2yb6Fn | 1-25               | Rogue   | mh   | timing + otsu for 29 swings  |    67 | 2479 | 0.90 |          |

## Shortsword of Vengeance (754)

- Proc: weapon proc 13519, logged as 13519; rolls on main hand, off hand
- Sim rate: 1 PPM, measured; measured 1.01 PPM [0.92, 1.10], 7 wearers, 8 runs, 441 procs on 10949 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights        | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 2AvFk4PJmBhyb89n | 1             | Hunter  | mh   | timing + otsu for 14 swings  |    24 |  498 | 1.21 |          |
| 3T8GrWJRdBkvy4gz | 1-6           | Rogue   | mh   | timing + otsu for 179 swings |    37 |  898 | 1.03 |          |
| NhHDL86y4PFrCW3G | 1-2, 4-11, 13 | Warrior | mh   |                              |    32 | 1026 | 0.78 |          |
| gDtFwpGPLM2TyZKB | 1-6, 8-9      | Warrior | mh   | timing + otsu for 104 swings |    74 | 1953 | 0.95 |          |
| kv32zprHjRfnZPh9 | 4-17          | Hunter  | mh   | timing + otsu for 23 swings  |    16 |  512 | 0.78 |          |
| q68xzRHYhCpFLjfJ | 1-4           | Paladin | mh   |                              |    76 | 1745 | 1.09 |          |
| tYFp32ThgGPfnvQH | 1, 3          | Rogue   | mh   | timing + otsu for 188 swings |    76 | 1880 | 1.01 |          |
| zX7thnkw1pR6WFyJ | 1-21          | Rogue   | mh   | timing + otsu for 24 swings  |   106 | 2437 | 1.09 |          |

## Skeletal Club (2256)

- Proc: weapon proc 13440, logged as 13440; rolls on main hand
- Sim rate: 1.05 PPM, observed; measured 1.04 PPM [0.84, 1.54], 7 wearers, 7 runs, 151 procs on 3338 hits
- Reports from 2026-11-08 to 2026-11-15

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 7rAN4vgZq3xjRhf1 | 1-2    | Paladin | mh   |                             |    36 |  807 | 1.03 |          |
| AQd8hCMn7a4ymGBW | 1, 3-6 | Shaman  | mh   |                             |    11 |  231 | 1.10 |          |
| F4rpyvWbxgfZCBXa | 1      | Warrior | mh   |                             |    16 |  402 | 0.92 |          |
| G7BvNCDbc12nKVpJ | 1-57   | Warrior | mh   |                             |    25 |  169 | 3.41 |          |
| M6XxqLHC8nmvB9wc | 1      | Paladin | mh   |                             |    13 |  286 | 1.05 |          |
| P143Tb2X9rBWwQDq | 1-18   | Warrior | mh   | timing + otsu for 47 swings |    24 |  754 | 0.73 |          |
| Qb7TJqYkKtgyz1ah | 1      | Rogue   | mh   | timing + otsu for 7 swings  |    26 |  689 | 0.87 |          |

## Sliverblade (5756)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1293653
- Rate: 2 PPM, measured

1,267 procs on 26,995 rolling hits, 21 wearers, 23 runs, in the main hand or the off hand: all hits 2.01 [1.93, 2.13].
The rate is all hits, which the specials rate agrees with. The one-hand white swings stay under 50 procs. One-hand
white: 2.39 [1.26, 6.35], 30 procs on 538 lone units, 3 wearers. Dual-wield white: 1.79 [1.63, 1.98], 299 procs on 7,139
lone units, 18 wearers. Specials: 1.75 [1.33, 2.94], 62 procs on 1,516 lone units, 10 wearers. Two-roll abilities,
priced as one roll: 5.36, 1 proc on 8 lone units. Travel time: the proc lands a median 89 ms after its hit, 98th
percentile 221 ms, window [21, 231] ms (348 isolated hits). Dropping the swapped stretches (22 active minutes, 2% of
hits, in 3 runs) moved the read from 1.96 to 2.01. 1 run logged in two reports counts once. By report date: before
2026-11-18 2.02 (704 procs); from 2026-11-18 to push 114313 1.92 (457 procs); after push 114313 2.43 (106 procs).
Reports from 2026-11-13 to 2026-11-21. The sim does not simulate this proc yet (the generator lists it as a missing
effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      16 | 1,066 | 22,585 | 2.02 |
| Warrior |       5 |   201 |  4,410 | 1.95 |

Reports and fights: 43wtgArG9vm8pb6N 1, 3-4, 6-16; 4PhL971zdKXZtnVr 1-2; 6RPtC7B1WwVbaH2r 1-4; 9g7NQBqDH6FGfP8L 4;
AFN6pDm8G9tq2Znd 1-6, 8-10; aPRXW1cvbptnDhj3 1-4; BPxvyfp6nT2NwjKX 1-8, 10-17; c69DL8gBmzZhHvdA 1-4; JHrkTPq23LpAcGxY
14-18; MTCvzYwVy4qrX8Gb 2-11; mYP6QAb2CJxFtknM 4-6, 8; nZ3kwNftM8h2L1zd 7-8; PkWRr8jZp9ATL7N4 1-23, 25-26;
PZKA7G2kQtw9H4Dx 4-9; QjJ6MpmFtXzdkRyL 1-2; rh3y7JwG9fVxq6DX 5; RnfbW8KTc4JHypP6 1-4, 6; TbB8q4RMK3LmYJNv 1-6;
Tm68LdcfN1rZYFap 1-13; xDcNRgK2YCA78yvd 8-10; yBdgRzfVhDPWAnbZ 1-7; Zd7QD3Lw1TBvmFCa 1-13; zMTFJZLpPfmXYv3W 1, 3-17.

## Spellsever Crossbow (274293)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1293714
- Rate: 0.95 PPM, measured

4,593 procs on 119,039 rolling hits, 143 wearers, 199 runs, in the ranged slot: all hits 0.96 [0.94, 0.99]. No lone-unit
rate applies, since its travel reaches past the isolation: the rate is all hits. A separate count reads its lone Auto
Shots at 0.94 [0.90, 0.97], which agrees. Specials: none. Travel time: the proc lands a median 515 ms after its hit,
98th percentile 880 ms, which reaches the 1,000 ms isolation, so no window was measured (985 isolated procs). Reports
from 2026-11-09 to 2026-11-19.

| Class   | Wearers | Procs |    Hits |  PPM |
| ------- | ------: | ----: | ------: | ---: |
| Hunter  |     125 | 4,582 | 118,881 | 0.96 |
| Warrior |      24 |     7 |     119 | 1.47 |
| Rogue   |      27 |     4 |      39 | 2.56 |

Reports and fights: 1DVFRby49Pg2tpzG 1-19; 1KF2jNqTzc6VJ9fp 6-8; 1VCHKmNy8FgrLnAh 1-5; 1fHnGhFabmPTq6Dv 7-17;
1hpF2AyzjRdrL78N 1, 3-12; 1jgTLvJQBMP6kxrm 1-12; 1qpQ6FW8fxD9gC4G 1-7, 9-11, 13-27; 1w9jNgyZv7YKAacb 1-13;
1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 3, 6-9, 11-15, 19-28, 30; 2WH4jLt3FADy8fnh 1-2, 5-6; 2aG3Ld8KVARCzrh1 1-7;
3LhpDbvyrtNFWdxH 1-5; 3PaV6RHdnzw7Jpk8 1; 3pCKd6jqgHyFc9kf 1-17; 3vn6yQxN8AWmpKaM 1; 4BhykDjm2p1ncLGK 10-11;
4KWV2rZDJMqNLRmT 1-3, 6-7; 4MJWTYdRLrtPjDN1 1; 4NMcqLwQ2rB9amCH 1-3; 6VJxcqnYa4wdBWQv 1-9; 6gvfVjrpZR8mcYxk 1;
6kN78G2xfqy3vwTP 1-6; 84VDXdcKZhHMPBAm 1-5; 8CzwkcFWrA91QaxG 1-12; 8R1c427qCbhPKgzx 1-3; 8WgnxAakbzDZdFBX 2-9, 11-17,
19-20, 22, 24-25, 27-28, 30-31; 8frHF64GzAhBxDyb 13-14, 16-22; 8rNkjZ2B6WdzFcKq 1-3, 5-7; 94fVcDCNBLztXh1k 1-12;
97FVRX1WLHqDmjAQ 5-12; 9DVWQNAfHXJBLy61 1-2; 9RrTvMpkamJPVw7g 4-15; 9g7NQBqDH6FGfP8L 1-4; 9zfKMnRxtJGZQXBL 1-4;
A8tQY9p3Dzn6TLRJ 1; BTRMkxmvfrwcZNaG 3; Bmk9chwYjd2F3zpZ 1, 3-6; CLXHZ3c2GyN6mYBD 1-4; CR4DHc2wFgjd7yrp 11-13;
CXvjmgaJ6YqKzD4k 1-2; CdD48cWzqZGQYmp6 1-4; D8JF7rbYQRcWxqn3 1-2, 4-8; DWJhjnmAM6GZdbKv 1-2; DtAhCGwd3cmQ4P2g 1;
DyYJRTv2XBfGcwC1 1; FDybHvQ3pPqBracN 1-4; FTXVdBrHJZD9aw7P 1-7, 9-23; FbkAhDQcP4KgpXGw 1-2; Fwcmz4r9aDnpRWAd 1-2;
GALhdpX84abJWPFm 1-4; GKZJBtWAvfPL81bc 1; GXmnxzAfNabcMB6Z 1; GZzaJH1QVjMKd8qA 23-34; GfwAMF6LZDm7TVWz 1, 3;
Gv6NDzAmFjC4P3qL 5-15; Gyw2T3LjptRW6vH9 1-6, 8-10; H9FXjaBWRpYVMq4r 1; HARvQ2hJWp8f3grD 4-9; HLrCnVfNMgYAJ1jd 1;
HN7bnjM6mrKWqDRT 1, 3-4; HnC7N2YVAdg3wtrk 1-4; Hp9ThMznR74PwYKb 5-8; J2XL6AyZqgpjx8QK 8-13; K6rZRbwYD1Wyjhpx 1-4;
K7T21rXDGxcWabtQ 1-3; KCYXdTfNgGHczbLw 1-6; KLNhgMTDAV6xY1Wj 1-3; Ky6k1d7mP2zVXWpt 1; KyYwHQjxAL8tzvdJ 1-7;
LKkr1pxAtT7CvcZN 2-17; M4CvXLZz3qm6pDRc 1-3, 5; M7BnDF4VPXcWRaZ1 1-5; MNZwjf63JzdrYgTC 1-3; MPkVah87LYAcnFdZ 1;
MgcQ6txHr9R1VjPv 1-5; MkrFhTq2HPnzmCVJ 3; MnX2ZVD8bxTFfjdG 1-7; MpcVyhTHnqCJvDFx 1-5; MtxT78RD4NkCJd1q 1-3, 5-7;
MyzjakDC9gBbpq6K 2, 4; Nc2grh6ByLvJCkRY 4-6; NzgdqKHnfLJWkFVM 1, 3-12; P8jXWFgJ7GVvwZ6m 1-4; PDT2yACNzd4LMZ98 1-14;
PFGvpDf9WZgcY6Rk 1-12; PLGKvft3bq8WgRYN 1-2, 4; PkCJZ74Lmc3HyW2d 1-8; PkWRr8jZp9ATL7N4 1-23, 25-26; Pn2LYHXpjZMDywgA 1,
3-9; Q1arfj4WPBNXqmLk 13; Q4GadwADtxbFBNW9 1-8; Q8GxhTmH9f6ajcNr 18-39, 41, 43-51; QJafvkCx4yhVXBmT 1-2;
QtVv7GfBDHLyn1bN 1; TFzR3fQkt2ap7K4Z 1-10; THqFxLkCP6ayd2tZ 8-9, 11; TQAKbWq7vVJHLgBR 6-18; TZnCMWqK9vDgHjBa 1;
Tc9DkwQFmvJr1NCt 1-5; V4HDKvxmYCkR1JnB 1-3; VZdxq6yNBrW48ztG 1-3; Vb3ZxfXtyDWr1ndP 1; VkJvpT6FhYwx4q9G 1, 3-11;
WYrCyJkdfMXw84Ph 1-21, 23-35, 37-39, 41-48; Wf26RdmcnLDTXwV3 1-2; XHDjdLb7R9P6NQm1 1-2; XYT1F4RgJLh9pyb7 1, 3-5;
XzxyRLg8QW9rqPGn 1-5; YH3c4zvFrpLtR1M7 5-9; YZyAfq6XN9k4a8md 12-15; Ya9kKCWRwqjFpgz1 3-4; Z4w7HKFPvcrTWd3M 5;
ZkVJAmC8vjaM6QTK 1-3; ZrD7NdhjWX4zG2PL 1-5; ZxWz2nQqXRcYNAbV 1-5, 7-10; a42r3p6ZxtBTRqfQ 1-6, 8-9; a6QpkvcR1YHLBqPh 1-6;
aVXmtyDj48AWwYQr 1-2; aj46hBQX1AfRFHtm 1-15; b12KD98fndF4hMq3 1-18; bCzt1Fq24mYcPXkB 1-2; bT3LyRBtax8FQnch 1;
bpV7JWX8B9xgrj3c 1-5; byYP67TH4KtZqm31 11-13, 15-17; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2; cY6xjKJN7QfdbHtZ 1;
d2qfGRQw8Pr3NaVA 1-8; dDKQgLW3MFPatynJ 1-25; dgnB4QvPHR8mhtWz 1-3; dn62GTvrBYFQbWfV 7-20; dnvYkhpKZQaqM2Hw 1, 3-5, 7;
ftTngaZr73hXKHLq 1-11; g7Bh9F2b1kz6Cnar 1-2; gk2jJvTHhW4zNDKB 10-11, 13-18; h6wkBb9Z7jFtTrnp 1; hCzY42vqXrc1mG7w 1-4;
hYHBQ9ZJX3vFAnkj 7-14; hrV3QgyzvfdZKJwF 1-30; j7V46tC2GkwNYTrz 1-2; jJWt7Dnm4qc9rzQd 1; jLRdP7ah9pzfgQ4r 169-178, 180,
183-184; k9dcpgjKfyXGwA17 1-4; kWBdFxYXLaN1vryA 1-6; mDbPCdX9AvwYJxGa 16-23; mDndY9rhZbAtWNPB 1-6; mYaWhF4n8zRLDZtw 1-7,
11-12, 16-17, 19-20, 22, 24; n1CBNaJq4WhAfw8Y 1-2; n7qPAd4rN1JyHfpR 1-3, 5-9; nhtYVjb7CAw9DvQR 1; pCjdMQw32FtPkYGN 1-5;
pKn8wBZcQ62PyJH7 12-24; qTHXf9y6ag1CMtdp 1-2; qg8YWNaCdDvV7XZM 1-8, 10; qprynNbXtRZV4DHP 1-7; qzv6W9fkDaP1MZYH 1-6;
rD8m4JW1bVyYPKGX 1-3, 6; rQ4dFakMnWXj89vz 1-6; rY4H13hP6MKZfXWR 20-23, 25, 27-31, 33-62, 64-66; rpLZ287nfKWyJGtb 4-6, 8;
rzvB21j7yD9hpaVF 1-20, 22-30; t681v9FWrLq4Tkxp 1-3, 5-7; tCprA1RmvTcJkq3D 1-4; tD7Rm3GBjdqPzhwC 1-4, 6-8, 10-14, 16-17;
tDK9h6nGz4cLTRq8 1-9, 11-16; tDhCZyVrn84xBaFR 1-7; tHz9M3arKp6PxA1W 1-14; tXFAxTLmg3bBNZrv 1-13; taFcCyAx9BD1KWXp 16;
tkQKL9bGT6JVR3na 14-16; tyFZPdH39kGq21An 3-25; v9Hf4bPtGq1LVykC 1-5, 7-8; vdNjgtCXJVM4PT9Y 1-3; vgrm7nPFN231GYRW 1,
3-15; vqCJK6rZNzHX2Ymd 1-9; w6A8bVm7vTatMCYr 1-12; wNqHMTvYmCQKXV3W 1-10; x3VKzQDvwLtCgbrj 1-6; x3p7mzNhFcarw1fk 1;
x7ybf4LjBD6GJqWn 1-9, 11-15; x89ayWwRt7DTvYCr 1-2; xRTpHyY7qC1GmghM 1-3; xYTCXa9myPJr6R8M 1-13; xnZRBaDMKtNHGQfj 1-4;
xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6; y2dFcNR1rMmKTPAj 1-25; yKvCQdzpgqY1WwbA 1-7; ynL7Gb2MpkZT46j3 1-9;
ypQ9d13tMFJPNfvw 1, 3; ythJ7TZDXk6AjHMq 9-10; z4ybMrFWNGa1L2xJ 1-3; zD9hTZPfwC38bgq1 1-2; zgG84wjLNY1Pv9tp 1-2.

## Stalvan's Reaper (934)

- Proc: weapon proc 13524, logged as 13524; rolls on main hand, off hand
- Sim rate: 1.4 PPM, observed; measured 1.38 PPM [1.08, 1.71], 6 wearers, 7 runs, 351 procs on 5262 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights    | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 2THWnvG8AXCYqfNg | 1-2       | Warrior | oh   | timing + otsu for 17 swings |    70 |  646 | 2.24 |          |
| 4ZKFNBk6czpynG3W | 1-2, 5-10 | Rogue   | mh   | timing + otsu for 25 swings |    46 |  779 | 1.22 |          |
| 4dVyt9cvbPrgpY2K | 1-2       | Hunter  | mh   | timing + otsu for 91 swings |    38 |  373 | 2.11 |          |
| 61VcxvhfQF2DRAGd | 1-2       | Warrior | mh   | timing + otsu for 33 swings |    68 | 1513 | 0.93 |          |
| BtcMA1CqVrnapQ2m | 1-8       | Warrior | mh   |                             |    90 | 1062 | 1.75 |          |
| C8Hk1q3rpN27YWam | 3         | Warrior | mh   | timing + otsu for 19 swings |    14 |  184 | 1.57 |          |
| MjxzVLAta8FgXR7y | 1-2       | Warrior | mh   | timing + otsu for 7 swings  |    25 |  704 | 0.73 |          |

## Strike of the Hydra (6909)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1292657
- Rate: 2.1 PPM, measured

4,070 procs on 35,041 rolling hits, 95 wearers, 110 runs, in the main hand: all hits 2.11 [2.04, 2.18]. No lone-unit
rate applies, since its travel reaches past the isolation: the rate is all hits. A separate count, leaving the Hydra
Strike events beside each proc out of the isolation, reads its lone white swings at 1.99 [1.88, 2.09], which agrees.
One-hand white: none. Specials: none. Travel time: the proc lands a median 133 ms after its hit, 98th percentile 283 ms,
which reaches the 300 ms isolation, so no window was measured (43 isolated procs). Reports from 2026-11-07 to
2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      56 | 3,736 | 31,979 | 2.12 |
| Hunter  |      40 |   334 |  3,062 | 1.98 |

Reports and fights: 12A94BJt3KCFZrY8 1-12, 14-21; 14caBjX2QdynLkq3 1-9; 1AxvQpfrXyTg3dCW 24-42; 1a3pDtCcPgBkL6wm 1-22;
1mRKL7Tb8qwjkgv3 1-2; 1yjwF8MxXqvtD4bT 1-3; 23ZyGvk7BXAVDT4F 1-14; 3BH2d16wtD9F78QY 11-18; 3tRNmGkZB2Jhnv6b 5, 7-9;
4nmJDcGNkZC2Phdw 1-6; 4y8nq6Xjg93zfrc7 1-9; 68fgHQ74dRxrwac3 1-3; 6DYqj7B2WHxPcabQ 128-134; 6X1L4nB7zCfqk8vV 1-2;
6jHPJBnLy1pF9Rkf 1; 6pjNwdr12qzBAKhW 1-2; 7GNK4Y9bkH1awZxL 1-10; 7yRD39nwhmMWYjbX 7-8; 8Zfz4JNLxhHcdDtX 2, 4-5;
8apqXKQg6mbC91L3 14-18, 20, 22-28, 30-36, 38, 40-46, 48-49, 51-71; AqxY6kh43jc2RPaT 16-17; AzMD4BWYKT6NGX2d 2-3;
B2k3Xcxz1hqQwJ6P 1-12; G4hzybWYmgrXvL83 1-2; GZzaJH1QVjMKd8qA 23-34; Gv6NDzAmFjC4P3qL 5-15; H3bRaDmzyjxTW2BJ 1;
HARvQ2hJWp8f3grD 1-9; HpftBG4KPCZrMxbF 1-3, 5-9, 11; HpyNQCMnDYFbJtP6 1-18; HtQr7kYaNq9MGgvZ 1-12; JQH2y6NWBXchGRbg 1-3;
KZYNzqfJFdW34DVr 1; KrtPndVLH4BjJcfD 1-4; L9k3ND6tw8VvcgjP 1-7, 9-18; LAm6bcJw47YBh2kF 1-2; MLhC3KjWrngq2vtZ 1-2;
MQwzXcfxb4v7hnFy 1-2; McYJrBbXAfKj82W4 2; MjfT3pQKRWcVG4gY 8-10; MvrY6aBgftZxj1QX 1-9; N178cBjPtChmbk4w 1-7;
NJLbaF8mGnk9yYAq 1; NKXBATdWYqmz78CV 1-2; NvnJVbtwK8z7THyx 1-2; PCfQFpjYdm3rD8qy 1-7; Pc3gyq8AFL9CD64m 2;
QCLPznqJMZ8vH3mf 1-12; QJ42CwNWLjFyvqfR 1-3; QhYmjJ4bHZXcgfk1 12-27; TCzRJ82rWcPZmp94 1, 3-7; TDqbRJWAN3x1FGhv 1, 3-9;
TFVvAaY6HLj2NqQp 1-3; TpxWY1MakPNb72H9 1-9; VHth6gR7xLbjQmnr 4-14; VQaYAf1Wxvm6ndHK 1-11; Vd1NZCgLMqxhb9kP 1-28, 30-39;
XF1faW9mcYtKrqMC 1-18; XVvTwBdtzAcKhPg8 1-14; XpQW1hmwHYDAJ8at 1-12; YX61J8CKqc42paNZ 1-4; Yt6CVhy34zjNHMcm 1, 3;
ZFLd4hkVbrJ68vnK 1-13; Zf21kBmtDM3AyRjF 1-13; ZkA3Vhz72n9Gy4dY 2-7; aRmjxFbtnGYAKCqk 1, 3-4; adXPnhT3kjHp9VcC 1-13, 15;
bPQCc3Nvq27HRWt6 1-3; bdYgfJk7Mh8GqZ4t 1-5; bnx7WwavBpKrQjNh 1-4; btfTZQLYD4hwvzqx 1-4, 6-9, 11; c7gHWKFrZMnm2yb4 1;
d6f294aV7vYnWmFj 1-5; dAV7JrjMkGfKNZtY 1-14; dQMKjpxJBCz1wrDW 1, 3-12; dZVp6gHA92mbDnPk 1-3; fnYThW3BpxKZv7jw 1-18;
gaDnjKvP6RkQTW2B 1-10; hG8tvQjLBJgf4WHY 1-2; hgd1qH4QYFyjz6AK 1, 3-6; jbhGzfrnJRWX978L 1-10; jrwhRBWdbZPK1Xka 1-3;
jtn8ap4XY6PzwDQL 1-11; jwDg8KGcd3QJ9kNZ 1-3; nB7qNVpyZvLPXCJY 1-20; nJtB3F7cQy6NmzGZ 1-12; nKdyat38zLA7fNJq 1-18;
njaHZVpvYFtrTLNh 1-10; nqjwyVJp17PDBYTX 1; pLFZK3MWV2TRw6t7 1-2; qfG6QWjd3wg1JVXv 1-18; rY4H13hP6MKZfXWR 20-23, 25,
27-31, 33-62, 64-66; t1rjBxRdP8DzVHhT 1-10; tVFZybKHRWp8xMcf 1, 3-10; txP9jAb7gGHhXckB 1-6; v4QYKcD7XymZq2JB 1-12;
vAGVD7aQZxNWX8h1 1-17; vnwtmMTpkR1aYP7G 1-9; vpdJ3cMhwTKNHX7B 1-2; vqBtZRrXwKQb6xHj 1, 3-7; vyzbt6kgRhTq4fdr 1-4, 6-9;
wYmr1kfWPpVz7jyZ 5-13; xnrg9m6jDMR1kaLN 25, 27-29; xwGgV9zJ3FNayZf1 1-2; yDMwzc49TJKmFfCr 1, 3; ycZLWXxQ4jk21w3q 2-28;
z4ybMrFWNGa1L2xJ 1-3; zWrR937TBH1FgvfP 1-2; ztD4pKHA6jGvfcVr 1-2.

## Sul'thraze the Lasher (9372)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1294430
- Rate: 1.45 PPM, measured

625 procs on 9,926 rolling hits, 8 wearers, 9 runs, in the main hand: all hits 1.45 [1.32, 1.56]. The rate is all hits,
which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 1.28 [1.10, 1.48], 126
procs on 2,267 lone units, 8 wearers. Specials: 1.72 [1.39, 1.98], 49 procs on 659 lone units, 8 wearers. Two-roll
abilities, priced as one roll: 2.70, 16 procs on 137 lone units. Travel time: the proc lands a median 99 ms after its
hit, 98th percentile 189 ms, window [23, 199] ms (168 isolated hits). The paladin check, which sets no rate: lone white
swings 1.43, 98 procs on 1,586 lone units, 16 wearers. By report date: before 2026-11-18 1.46 (415 procs); from
2026-11-18 to push 114313 1.44 (210 procs). Reports from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   613 | 9,677 | 1.46 |
| Hunter  |       1 |    12 |   249 | 1.11 |

Reports and fights: 2ZHDh8LX4mKqctbz 1-3; 8ZdnfAT79zPwmyVM 1, 3-6; Ag69GWdLRDhKCrcT 1-6; cL3KMWpnXyvDrqa2 1-2;
hCPLc2G7KAVX8Bby 1-6, 8-9; HT2qmWDJNfXVj1xA 1-2; mYaWhF4n8zRLDZtw 1-7, 11-12, 16-17, 19-20, 22, 24; tnyh4dxq6pNXQzLw
1-3; tVhnGRkHP7f2xarB 1-3.

## Supercharger Battle Axe (9486)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 13527
- Rate: 1.9 PPM, measured

258 procs on 2,923 rolling hits, 9 wearers, 9 runs, in the main hand: all hits 1.89 [1.72, 2.00]. The rate is all hits,
which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 2.08 [1.53, 2.50], 68 procs
on 702 lone units, 8 wearers. Specials: 3.21, 3 procs on 20 lone units, 1 wearer. Travel time: the proc lands a median
95 ms after its hit, 98th percentile 161 ms, window [17, 171] ms (71 isolated hits). Dropping the swapped stretches (16
active minutes, 9% of hits, in 1 run) moved the read from 1.71 to 1.89. By report date: before 2026-11-18 1.90 (187
procs); from 2026-11-18 to push 114313 1.88 (71 procs). Reports from 2026-11-13 to 2026-11-20. The sim does not simulate
this proc yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       8 |   251 | 2,863 | 1.88 |
| Hunter |       1 |     7 |    60 | 2.50 |

Reports and fights: 2axJHNwqQdtW1jnF 1-3; 7kxNnWGw2FzfLJtR 1-8; argXcZydQj9tTHD4 1-2; Cc32w7hgMJQ4Tny9 1-2, 4;
fDycAFZWt4NkRX1x 1-3; FRxmtWLHr3wvB8Q1 1-8; LjFr1v7pJKVcfhWB 1-11, 13-14; rNwJc31qXfzWthQD 1-5; tBnv1kDbzHK6hAwg 7.

## Sword of Corruption (13032)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 17510
- Rate: 1 PPM, measured

1,249 procs on 35,139 rolling hits, 59 wearers, 67 runs, in either hand: all hits
1.02 [0.96, 1.07]. The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white:
1.07 [0.88, 1.36], 144 procs on 3,857 lone units, 9 wearers. Dual-wield white: 0.89 [0.80, 1.00], 440 procs on 14,093
lone units, 50 wearers. Specials: 0.99 [0.79, 1.20], 123 procs on 3,551 lone units, 24 wearers. Two-roll abilities,
priced as one roll: 0.84, 1 proc on 34 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 522
isolated hits. Dropping the swapped stretches (87 active minutes, 7% of hits, in 4 runs) moved the read from 0.95 to
1.02. 1 run logged in two reports counts once. By report date: before 2026-11-18 0.99 (944 procs); from 2026-11-18 to
push 114313 1.11 (305 procs). Reports from 2026-11-11 to 2026-11-20. The sim does not simulate this proc yet (the
generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      31 |   730 | 20,969 | 0.99 |
| Warrior |      17 |   486 | 13,169 | 1.05 |
| Hunter  |      11 |    33 |  1,000 | 0.94 |

Reports and fights: 1h9vznDdxyjC7KWZ 1; 2mr4qBpFYGxVMAnj 1-2; 3npHxB84DT2zRtmN 1-4, 6; 4ChPnBrVK2fJLMaX 4;
62WRbBjFQCpZGr3z 1-4; 6dDjtwh1HXFxg72L 1; 6XBqk2G4LZy83mQz 1, 3-7; 7CcF8GHPv4fagQYh 1-10; 7LY3ADmFM1wBarRn 1-4;
7tnLcDMYbdk6P9Jr 1; 87BwAPvWYCTnKJ6Q 1-7; 8dt4TNMHjYCKZ3Xr 1-7; 9bvQ2gKdAt1nFx8C 1-2; AY9aT3fZ4RvQKjrV 1;
bYKdfpLMCNQDHF7v 5; cBQbDVqmdMtPCwLg 1-3; CkcZ7wQRdhpjYPfb 1; D1qH8nWkjQBLxPbX 1; DfMnbv3KLYmp2aCX 1-4; dTc2jh81CHgYFRZ7
1-6, 8; fJxRkQjHap2ThDyd 1-8; g3BPNGHVcYkAwx72 1; G4yA72PjmHJvBKDh 1-2; GV4wtvzD8RcfTqNC 1-2, 4, 6-15; h9ac8AXNDbLwJCzy
1; hByjwtf1pNa8q2Ac 1-2; hXv27Jz1dw8tVNPA 1-3; Jay1j4Q6gWNLZxD8 1-5; JCFHXZxhBYjP2vt6 1-12; JGyv2ZTWV6C4R1wj 1;
K14qDRFZjrfzYCxt 2-8; K1XzMGFDwrt7RQbv 1-3; k6bWMqt1BpjXyZAJ 1-2; kAJd8bLqcM1nNyfh 1-6; kbPtGMLcJ2Brp9vN 1-2;
KfAmWF67gcjrMw2k 1-5; kHfNcJTYj3RtMFGL 1-3; kn8aQKN2HBhqjG9z 1-3; KNhjmWPtyVCkv6B4 1; kNhVJ9WF2LmXab4w 1-2;
L4ZWkKT3cGNCAqnX 1-2; mqVALKC7PcRFNvky 3-5; NfDXwFm37WHghq9d 1-5; nKCDhXNbHpJaBcAR 15; nYycdrvkhx8RHCDw 1-4, 6-13;
p3dMajX1JRAwVrLW 1-2; Pck4Hb83WgXxrfJF 1-3; PKpXc2n7WrBjxfCA 1-2; pKwNTxL79kXv2tha 1-2; pwNAZn643XgzcbYa 1-3, 5-6, 8-10;
pyAxZXhWjqkY6NGD 1-4; QZcRmKX4vMVh7f6j 1-14; tAZ12pKHxQYLydgv 1-2, 4-5; Tbh1yKNfPtqLYCak 1, 3; tYHypqn6LDjXfxGR 1-5;
TYqhJAWQ64zP7Zjw 1-2; v6Ljcx9QBTZNakgt 8; VxmDcHTK8LqgNCwQ 1-7; WvJtM32w6FALXbPK 8-9; x2LqDAP6WtnJvMKk 1-2;
xN4Jd3VCWBZpnM1F 1-6; ybvr1GhnNfc84wBH 1-3, 5-6, 8, 10; ycZLWXxQ4jk21w3q 4-28; zaV76QJxyhbwD4pR 1; zcK9qfp32VnQWTMP 1-2;
Zd8PXKqcNQygRjT1 1; ZHK32LtTCmqR4Pxf 1, 3.

## Sword of Decay (1727)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 13528
- Rate: 1 PPM, measured

404 procs on 8,932 rolling hits, 9 wearers, 9 runs, in the main hand: all hits 1.01 [0.92, 1.10]. The rate is all hits,
which the specials rate agrees with. One-hand white: 0.88 [0.84, 0.94], 42 procs on 1,059 lone units, 3 wearers.
Dual-wield white: 0.86 [0.50, 1.07], 66 procs on 1,714 lone units, 6 wearers. Specials: 1.10 [0.95, 1.34], 104 procs on
2,109 lone units, 9 wearers. Two-roll abilities, priced as one roll: 0.00, 0 procs on 8 lone units. Direct: window [-20,
100] ms, median 0 ms, 98th percentile 0 ms, 165 isolated hits. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   225 | 4,715 | 1.06 |
| Rogue   |       4 |   179 | 4,216 | 0.94 |

Reports and fights: 2mr4qBpFYGxVMAnj 1-2; 9nCtfP6kvHrzV4FL 2; LDg2vXctwnYM9jBq 1-6; MXjF1GhRmLfabNZV 1-18;
TaNXwvhfBLxdyQc9 1-2; ZHK32LtTCmqR4Pxf 1, 3; ZqghBpkcy7xtPfmH 1-9; nZ3NhbvpVXg649Bm 4-10; qg8YWNaCdDvV7XZM 1-8, 10.

## Tainted Pierce (8225)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 13530
- Rate: 2 PPM, measured

872 procs on 13,855 rolling hits, 17 wearers, 19 runs, in the main hand or the off hand: all hits 1.99 [1.81, 2.11]. The
rate is all hits, which the specials rate agrees with. One-hand white: 1.26, 8 procs on 200 lone units, 1 wearer.
Dual-wield white: 1.80 [1.46, 2.01], 313 procs on 5,489 lone units, 15 wearers. Specials: 1.88 [1.44, 2.28], 69 procs on
1,159 lone units, 8 wearers. Two-roll abilities, priced as one roll: 3.54, 22 procs on 196 lone units. Direct: window
[-10, 100] ms, median 0 ms, 98th percentile 1 ms, 277 isolated hits. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   430 | 6,503 | 2.09 |
| Warrior |       6 |   356 | 5,872 | 1.91 |
| Hunter  |       2 |    86 | 1,480 | 1.84 |

Reports and fights: 1zLtZcp3DRP97Hdx 1-6; 4na1LcCpP8QHTNVD 1, 3-8; 8WfLRh1qTpFxJgra 1-2; 8rNkjZ2B6WdzFcKq 1-3, 5-7;
J97GvmrRXCFZfQ8a 1-2; LGj3Kfvkz7cgZ1ya 1-5; NXgv93kr2G7zyTxm 1, 3-9; R3CDcTfm6bdNJrHx 1, 3-7; RcTKFQbGnmtJHWpa 1-2;
THqFxLkCP6ayd2tZ 8-9, 11; TnbzrVaxtHcpMKLY 1-3; VyZXRT7nrPJKBAt8 5-13; akPJ4AcBZQnb36f2 1-5; c9tKjzRyaYXQfnk8 1-3;
gn4HAC1RqMfLt8wN 16-21; pFNnChR7GzAX6aH3 1; qcrKJGPvkfhAtdjQ 1-4; xqy2GfPj67XCwrT3 1-4; zTMW1B2rRJxbm3w6 4.

## Taran Icebreaker (2915)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 16415
- Rate: 0.95 PPM, observed

694 procs on 11,509 rolling hits, 10 wearers, 11 runs, in the main hand: all hits 0.95 [0.90, 1.01]. The rate is all
hits, which the one-hand white rate agrees with. Observed: every wearer is a shaman. One-hand white: 1.10 [0.95, 1.27],
219 procs on 3,150 lone units, 10 wearers. Specials: none. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 2
ms, 188 isolated hits. Reports from 2026-11-13 to 2026-11-19.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Shaman |      10 |   694 | 11,509 | 0.95 |

Reports and fights: 1MA3qWLGa8v2hdXP 1-48; Bz8CAjZnNyXhgTtH 1-14; FRmWJKpCD8t7QGb2 1-3; PJHx4V21KhfCaryz 1-2;
Tx7MPH9f8Ga6hdrF 1-20; VhN7grtcvRYJxjBH 5; aAcjGPbyZXNTzR1K 1; aQAhMLzBbWJcdq6G 1-6; cLxk98bVfwpXWaqN 1-21, 23;
mg3MBw6rcNtZpjbA 1-14; nhtYVjb7CAw9DvQR 1.

## The Blackrock Slicer (13285)

- Count: §Count of the 2026-11-09 to 2026-11-19 reports
- Proc: weapon proc 17407
- Rate: 1.9 PPM, measured

2,154 procs on 17,055 rolling hits, 14 wearers, 20 runs, in the main hand: all hits 1.89 [1.81, 1.98]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.92 [1.80, 2.02], 395 procs on
3,082 lone units, 14 wearers. Specials: 1.97 [1.88, 2.13], 107 procs on 815 lone units, 7 wearers; Raptor Strike gives
77 procs on 608 of them, 1.90. Two-roll abilities, priced as one roll: 3.37, 38 procs on 169 lone units (Execute and
Mongoose Bite, left out of the rate). Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms, 458 isolated
hits. Dropping the swapped stretches (55 active minutes, 9% of hits, in 2 runs) moved the read from 1.72 to 1.89. The
paladin check, which sets no rate: lone white swings 2.59, 38 procs on 220 lone units, 2 wearers. Reports
QqKMnRmb7DrFXC36 and bGq3mN7JgCDzcyMZ log the same fights of one hunter, which the count takes twice (36 procs on 230
hits in each); taken once, all hits still read 1.89.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |       7 |   902 | 7,136 | 1.90 |
| Warrior |       3 |   760 | 6,322 | 1.80 |
| Hunter  |       4 |   492 | 3,597 | 2.05 |

Reports and fights: 2Dv3WbJrMTCFPVgH 1-4; 4yJ2cwCFqNnWYthD 1-2; 6FVmWYfq7DG2pzwJ 1-7, 9-10, 12; AKXn47BZWpyhJRv3 2;
AN4RP21ZMXLVpntm 1-3; CWzdaFnMfPbALgmJ 1; KtCPg2xhR416AvXN 1-6; MBRmC1TnXq86k2Nb 1; NXMZV9xjvWpB7GtP 10, 12-107,
109-116, 118-122; QqKMnRmb7DrFXC36 1, 3-6, 8-9, 11-12, 14, 17-19; ZaPqVJn1bQ894pxT 5-17, 19-52; ZqtpzHLbAQKWMcwN 1-3;
bGq3mN7JgCDzcyMZ 1, 3-5; bahZczXqfLk3pQR2 1, 3-5, 7-19; cav7VZrmPbNW3y29 1-13; cbTt1zxajFfNVPDA 1-4; kaPpXgt1bWJBr4z2
1-3; vwCXzhMQ2drgpRxD 1-41; wrzB6GtjPaqCKJZW 3-10; yVp9KT3QkzhmZXDR 1.

## The Chief's Enforcer (9477)

- Proc: weapon proc 1294339, logged as 1294339; rolls on main hand
- Sim rate: 1 PPM, measured; measured 1.01 PPM [0.92, 1.11], 7 wearers, 7 runs, 422 procs on 7354 hits
- Reports from 2026-11-10 to 2026-11-16
- Notes: 6 of 13 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights         | Class  | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | -------------- | ------ | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1P4ZKpMGVjvntRgm | 1-3            | Shaman | mh   |            |     8 |  148 | 0.95 |          |
| 28AhM3VGnvFcJ4rD | 1-4            | Shaman | mh   |            |   116 | 1733 | 1.18 |          |
| 3LTnjm16YN2bDzCP | 1-6            | Hunter | mh   |            |    14 |  344 | 0.72 |          |
| 3PAwV6XFaB7hqDgz | 1-12           | Shaman | mh   |            |    27 |  505 | 0.94 |          |
| 7rzAygVjwH2C69Tv | 1, 3-14, 16-62 | Shaman | mh   |            |   103 | 1777 | 1.02 |          |
| fywmzq86PRtjDJra | 1              | Shaman | mh   |            |   127 | 2378 | 0.94 |          |
| mCqj9fpbthFNGgkc | 1              | Shaman | mh   |            |    27 |  469 | 1.02 |          |
| 274WmyH18B6ZRgAz | 1              | Druid  | mh   |            |     7 | 1218 | 0.10 | feral    |
| 4h1PjwkaKX9q8mgT | 1-2            | Druid  | mh   |            |    14 | 1347 | 0.18 | feral    |
| 6f2mtzphbVCgKYJ3 | 1-17           | Druid  | mh   |            |    25 | 2487 | 0.18 | feral    |
| 7Yj1a2RW9JCvDbHp | 1-7, 9-11      | Druid  | mh   |            |     4 |  951 | 0.07 | feral    |
| GTwzB9Q3VPZW8jYy | 1-5            | Druid  | mh   |            |    33 | 2767 | 0.21 | feral    |
| fTn7FmzCwkJpZt4N | 1-3            | Druid  | mh   |            |    17 | 1545 | 0.19 | feral    |

## The Cruel Hand of Timmy (13401)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 17505
- Rate: 2.2 PPM, measured

1,200 procs on 18,223 rolling hits, 18 wearers, 22 runs, in the main hand or the off hand: all hits 2.20 [2.06, 2.37].
The rate is all hits, which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 2.27
[1.89, 2.70], 124 procs on 1,823 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 1.91
[1.71, 2.07], 329 procs on 5,748 lone units, 14 wearers. Specials: 1.73 [1.58, 2.17], 32 procs on 615 lone units, 4
wearers. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 318 isolated hits. Dropping the swapped
stretches (43 active minutes, 6% of hits, in 3 runs) moved the read from 2.06 to 2.20. 1 run logged in two reports
counts once. By report date: before 2026-11-18 2.17 (574 procs); from 2026-11-18 to push 114313 2.22 (626 procs).
Reports from 2026-11-13 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   642 | 9,903 | 2.16 |
| Warrior |       9 |   558 | 8,320 | 2.24 |

Reports and fights: 1ZyhaVDdFtTYWMcC 4-7, 9-12; 3BAz1TDGL9kjy8ZQ 1-2; 72q1FtWgTMHAcLxG 1, 3-5; 8jDqr3JcYRCH1pbK 1-4;
aZ8yxm3AGgrpTvYh 1; C1cyPbL3RQvGjq8p 48; H8AgW6afwBktp4nv 3; HKjBYG4kRJtnXyc3 1-2; HL1FjV9MzyNDrb48 1; hLxABfWVavkM8HGT
1-4; KtCPg2xhR416AvXN 1-6; p6mVRBzvWMqnQZFc 11-12; qa8wcv4jLYZpGARK 5-9; Qb1yWxa8vHCGL7Yr 2-7, 9-15; qcdyZAr2ganDtxNW
13-44; QcrC9VK64xXa1JWg 1-8; rRnVJhQFa9zf7NqK 4-5; vT1rMa2cNGXLfqp7 9-12; yCdRrMhYwm69xj1Z 1; zBqkZQ1bxp3cAhDa 15,
17-18; zy3NjCQP7bMp1W8H 1-2; Zz3YcQk69P7N8rAn 3-4.

## The Ziggler (8006)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 13482
- Rate: 3.6 PPM, measured

3,916 procs on 38,305 rolling hits, 49 wearers, 50 runs, in either hand: all hits
3.61 [3.48, 3.78]. The rate is all hits, which the specials rate agrees with. Dual-wield white: 3.17 [2.94, 3.35], 1,178
procs on 13,126 lone units, 48 wearers. Specials: 3.91 [3.68, 4.13], 147 procs on 1,327 lone units, 21 wearers. Two-roll
abilities, priced as one roll: 11.76, 3 procs on 9 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 817 isolated hits. Dropping the swapped stretches (22 active minutes, 1% of hits, in 2 runs) moved the
read from 3.56 to 3.61. 1 run logged in two reports counts once. By report date: before 2026-11-18 3.57 (2,489 procs);
from 2026-11-18 to push 114313 3.67 (1,427 procs). Reports from 2026-11-13 to 2026-11-20. The sim does not simulate this
proc yet (the generator lists it as a missing effect); the rate is recorded on the item's effect.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      41 | 3,502 | 33,940 | 3.64 |
| Hunter  |       5 |   140 |  1,363 | 3.63 |
| Warrior |       3 |   274 |  3,002 | 3.22 |

Reports and fights: 2bYvT8kWxjFfdwrm 1-2; 2K4Z1jLyGWzdnD3f 10, 12; 3DQy7BhjH1mnZPpq 9-12; 3HwJAWkhFz7gCaQy 8-14;
3QgzWpJKCNMRvAYT 3-8, 10; 4xnDCWwrGKAFdNHv 1; 74GT2KcnWCLq6Bby 1-2; 7dZt6b9Kh4F2CNAn 1-12; 8YzmGj6fAZqMBNtx 1-7;
9kRzG7P6AKj2pdaF 1-2; 9MmVG8crjzJZ24K7 1-4; 9WaCtMzNbFcr4qmv 1-4; A2FMrK3Y1xkvafQH 1-7; adJRDVZ1mhYP8vj2 1-4;
bLmVHN7BaP8DvX9Z 1-16; BmfFgAczHXN7b16M 9-24; BTRMkxmvfrwcZNaG 1-3; fCJMF2rZqAHjBw67 1-5; FR9jJ7b31XCkLmtQ 5-7;
FxndQv9RDXt47KA8 1-2; FzCJvnHfGyMLWVwb 1-11; gLcGzRyxhZ2akWjw 1; GyQ6cqtCxPXHaDFz 22; HAZ4a9BmGxyD8gV3 1-3;
kvZgPDVybd6YN4Hw 1-2; LH4aGRZt7wVhkpcD 5-21; mjbfcTLkM2PGKadw 4-7; mwvHGaXY3KJ1xjdh 1-7; nwh3zdqc4CjXa6MG 1-7, 9-10,
12-23, 25; NyC7zw4Pm8aBHbZ2 10-20; pmKDwngMT2NFktdA 8; Pn8QfjmVL46tWpXR 1-9; pZNqCry3wRxmhk2L 1-16; R9njq1AGDvmp36bM
1-8; RPnVYQcG9xTJh1ar 1-4; V4HkPzgr21bxRvFt 1-3; vcxFpTb6zP2WrVAj 1-8; vkM4HWNYydVbPGxf 1-4; VYBNbJx7G8kP6vXd 10-15;
x3PftMnXCBZgLAG2 1, 4-9; x7mkBJZvX8PnKcRf 1; xM9bRGyajmBFvCqt 1; y8wDKJvdL1Ncp9za 1-3; yb6BrpYZnzHajqmx 1;
yhVbHg7vzDkKfnRY 1-7; YTn9F1x6jLqN2fdQ 1-9; ZbY9WHdayD6FGhJ7 3; zfaGW68H1jp9LRVm 5-14; zhkQZTaX8cVg47Jv 1-5, 7-8;
ZkpTbqwFYMAaLP4r 1, 3-6.

## Thornspike (6681)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1293183
- Rate: 3.15 PPM, observed

17,132 procs on 191,806 rolling hits, 242 wearers, 279 runs, in the main hand or the off hand: all hits 3.15 [3.08,
3.23]. No lone-unit rate applies, since its travel reaches past the isolation: the rate is all hits. Observed: no
lone-unit rate checks it. One-hand white: none. Specials: none. Travel time: the proc lands a median 525 ms after its
hit, 98th percentile 948 ms, which reaches the 1,000 ms isolation, so no window was measured (342 isolated procs). Left
out: 2 of 282 runs in Cat or Bear Form. Reports from 2026-11-07 to 2026-11-19.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Rogue   |     169 | 12,619 | 140,616 | 3.17 |
| Warrior |      41 |  3,746 |  42,171 | 3.14 |
| Shaman  |       4 |    380 |   4,586 | 2.92 |
| Hunter  |      28 |    387 |   4,428 | 3.08 |
| Warlock |       1 |      0 |       6 | 0.00 |

Reports and fights: 12RB4fbmrp3YZjPF 1-5; 192ZN3dKFymLACY8 1; 1VKQY2Dafc7PJzpR 1-2; 1Zv6bBfgDP3RVFYa 1-2, 4-7;
1kWMb9cQ67dJhrZz 1-3; 1tCvNKHFM3c2WkmA 1-12; 1w8AbW4RBNLX6Kfy 1; 21PKRb9nk3HdYFm8 1-11; 2K4Z1jLyGWzdnD3f 1-8, 10, 12;
2KkaByn6H9mzb4Yq 1-4; 2t6mDPLkfWBxcr4q 1-18; 39BFWKC2J8RTzyMx 1-3; 3DdtwKbCf7g1XVyT 1-2; 3PXfhvmTp6NQdwRb 1, 3-12;
3ZFVrmzh4fanLM1B 1-8; 3bqWjVMPpDfdCgYz 1, 3; 3f4WnNCQJX9azxmV 1-5; 3hNwr9K71bz4xjCc 1; 3kqWrL1dwJfxMcX2 1-5;
3npdHy9Ag7rMTC6f 1-5; 4KCzQ8mFr9A1bkxG 1-8; 4L7DBHTV2jA6gkQW 3-19; 4WYavRTZm7QHxLgA 1-20; 4ZNwHaLtygF1k3xC 1-3;
4zYkPVvpb9KXLxA7 3-5, 7; 6CPXjKh3z1TDWtqM 1-10; 6tTZVHLYkbzrCXqD 1-8; 72gPMatGLpwd93jW 13; 76ZV8FQ1RHDhbq3z 1-4, 6-9;
7M1KwnDZkXvyRFB3 1-5; 7dZrzYqjKHh28mvW 1-2; 7dZt6b9Kh4F2CNAn 1-12; 7f6DBRAPMLGYzyQm 1-2; 7g9DpfNFBA3PatdZ 1-7;
7gPtBMrHwGyjWKRc 1; 7xGQWc4AbvD8CNBw 1-7; 81CznKf7taTFArLw 1-8, 10-16; 8MjcygF3zWK4d9NB 4362 fights from 1 to 4766;
8Z1rYqmCWKbLjvwz 1-2; 8ZTCpbxtFA1zfRNX 1-3; 8cyBqw6Cr3WjV9Ft 1-6; 8raZtT1Qq7WkYv2P 2-5; 9B8Wf16dHKjgaMDY 1;
9Nayz1f8BJnLcgQm 1; 9PxCyQfrdhXFmNGV 1-2; 9g7NQBqDH6FGfP8L 1-4; 9v6VdC7ymQfb4cWh 1; 9zfKMnRxtJGZQXBL 1-4;
B1xchF3f6L2Mkga4 198-205, 207-219, 221-238, 240-257, 259-269, 271-282, 284-287, 289-292, 294-306; B6jfw4mcnPxWCKYD 2-18;
BZxkqfarM97yHgLp 1-6; BdLc4AgFbrz8Tq2k 1-2; BxpTfqGZb34aJLc6 12-14; CPj7JpNFLDztgQRa 10-19; CbG9THJkRBjDmPX4 1, 3-6;
CjLKJRNhnb7TBzAZ 17-20; Ct8XcxY26Mf34NjR 1-9; CwYNtp2VQrJdbfx9 5-8; D8qzgc9h1YLf6krp 5-6; DHyYh18vLXAFTJqt 1-5;
DRv3TVCwrAPJ9zK8 1-2; FBQzxArfjdCZp6Lg 2-11, 13-18; FTrHGQmvDM2jzRn4 7-10; Fj9WzagMcZxw4mXY 1; FvdWRrT67kAPY43x 1-11;
FzCJvnHfGyMLWVwb 1-11; G6wYN314W9FfVXpr 1-2; GAgDtJmKHdBhnpwV 1-5; GCyghdxmrfYL8zcX 1; GJxyK8rTBamY7CvH 1-11;
GKmRMd7XzWtavJg8 1, 3-9; GNRbXQDvCkYpAta4 1-8, 10-12, 14-22; GRZrtByWvML6mzXA 1-11; GRdTC8kXNwVxpjKQ 1-2;
GV4wtvzD8RcfTqNC 2, 4, 6-15; GVzLqkhK4PMXjYmn 38-50, 52-70; GX7CDf4dvtLmR1Wh 1-3; GXmnxzAfNabcMB6Z 1; GYPbQpyCvVajRH4L
1-16; GYQB4c79w6TF1nWZ 1-6, 8, 10-11; Gb2W8Tga4VqvZy9c 5-11; GpdNkxVgzrtF49Xb 1-3; H7yaPANhDBKQ8Lmq 1-13, 15-17;
HTpFLJAq8CWmPvzr 1-2, 4; HWcyVPwhnNB9brYz 1-5; Hdnh1GJgD4TbQkXL 1-14; HqW4rjd6XmYgDM9F 1-2; J3tdCaXR8hAP9m2M 1-4, 6-9,
11; J9jRmkxvtXcAKfM3 1-4; JB7fzC2ZnP3Kwjxd 1, 3; JD1ypjL3xcCKfTGP 1-2, 4-6; JNZVwGWTBMYLz9an 1-13; JPB6NXkAGCQzmh38 1-5;
JZx3p6Nqdj9tCrMc 1-3, 5; JvdQP6VyY19kZDXm 1-9; K3jWRx2qQ1aXy9PA 1, 3, 6-13; KG4FcnrYWVHCfzvt 1-9; KNqTmpW6HZCnc378 1-4;
KmahZwCJcYFLj1Rq 1-7, 9-16; KmjAV91JBhgQLrfP 3-4; KvF7tTamGpn9kW34 15-24; LN78yvQwgJpcAYCG 4-6; LW4qmQykhj3p9atv 1-2;
LWx7bgKmpQ6ktYRr 5-12, 14; LaMmthYQK4HRWwXC 1-5; LaQztr3mDTwHyknB 5-8; LmDb8njHfXvAgC79 1, 3-42, 44-51, 53-63;
LxdYWR47ZFbTNBhM 1-2; M4CvXLZz3qm6pDRc 1-3, 5; MCdZTywRz3HYnvWg 1-10; MKk2adJtjyANf8LW 1-13; MW1Kw23ncjRDrxpA 1-14;
McNx4Dwk89vYJpdH 1-5; MdaKf7zpNgWHF1Jj 1-4; Mf1AyXnC6W3dpNJR 1-12, 14-16; MyZ8p6RaAFBWcJ1f 1-7; NLzDkMAj8n2TmbxV 1-6;
NWB8HZd1JvxzGjDr 1-3, 5; NZcYhqPxkrJMKz6a 1; NfLb7C1qDZ8cHxFp 19-29; NwKVWqHfgk1jLvAG 1-8, 10-13; NyZgJMVBXHn8vtWk 1-8;
PFGvpDf9WZgcY6Rk 1-12; PHqvNCxT1JVtpWFj 1-4; PN64hvHQLZRCptGg 1-5; PNnXcA9hFBx7Wtdj 7-11; PyH9Adag8VwRjKJh 1-16;
Q1arfj4WPBNXqmLk 11, 13; Q3kAgpdcGJ27VPvT 1-3, 5; Q4GadwADtxbFBNW9 4-8; QAc4wgTL7PfDphk6 1-5; QMkfqDhm3L69A7VY 1-11,
13-14, 16; QRbmZWh2yDGakt9Y 1-4, 7; QxJKAphaZt61kzvg 1; QyYtrwM6zKPXHDj3 11-15; RDPLkYamNgHKXFh2 3-7; RbLqYxNQj1MaWFdP
23-29; RdmgVNQ4XJMfnC1F 4, 6-10; RfxgwvQjA1pVy98k 1-6; T2MrDwA9hdgPXxRv 1-4; T4pxtrKv178PGNkA 9, 11-14, 16, 18-26,
29-41, 43-57, 59-120, 122-142, 144-164, 166-208; TgVP6hxZKqY7MGQ4 1-5; Twr9aYQFMWjnLPqR 1; VMP6Dv2pyTnYhaKJ 1-3;
VRM3KtXTh2W7kzJx 1-4; VYb6f27NazpmCch3 1-2; Vh3bzdRMP6ymNKq9 1-2; VhHLyw9CP76Wv3fX 1; VvmGw1CT7bZMNXp6 1-12, 14-19;
WAKwhqRtmPGT49ZQ 21-35; WFJf43apKLXHVzBh 1-4, 6-7; WqBa2AwcGHfgj8Lt 7; Wza8bdXVRhcgqGnZ 1-5; X16Yp9FzW87cNADq 1-5;
XYTRKMVFWdypLxtB 1-14; XhZAV37vMFbYWRt6 1, 3; XwqT7ZK48Df69tkh 14-22; YJzQRVL4r9DqFcpf 1-16; YcgyKZVCJvbW7Hj8 1;
ZCRLWJqmb7v3gjaN 1, 3-7; ZKLvG3Bmy9CAgjT1 1-7; ZKRmDGdV8FM1aYwt 1-13; ZPK6pnD7HzmyhvjT 1; ZPxzwLAmBpRMTNbC 1-3;
ZPyMjD1FkVGJH8Lb 1-4; avQywnDKt7RxcpfF 1-8; b12KD98fndF4hMq3 10-18; b4fNagK2jvXQJAGL 1-10; bAZJWP3hgxRav6Fm 1-7;
bdPKyBjRNDX9CHaQ 15; bpV7JWX8B9xgrj3c 1-5; bxywkpHBTzg2mtCQ 1-20; byYP67TH4KtZqm31 1-13, 15-17; cLkwmZbzFDKdPhYn 21-22;
cPHdgRMCy7ZGDqFf 1-8; cT8wJVq316kAWpn9 1-3; cVPtG1agnWhfX2xC 1-7; cXCfvnWtNrR63PVx 1-2; ctAnYvNLQK6F38PD 1-2;
d8KLZbmzWhwr9at7 1-4; dXR2BPqTYQ4aMxZj 1-5, 7; dtHTkZgMNvR3YBK4 1; f9vADt2CZ3M1ygwq 1-11; fRPcJYzWKXBCv32N 1-2;
fjpy1TBL4b3kgJhr 1-8, 10-15; fjqQNyMbD1Y7R92h 4-21; fkD9m3dBKcGAyqRY 1-6; fnMH4F2pkGC8T6Kz 1-8; fr2WyGqPVYNTXd91 1-4;
g2qW8NnDvCQ3TfVG 1-5; g8b1MkJnKtyfzvPa 1-5; gGn1vQ28YxpFrmaP 1-3; gT2mW39CV7DMwhzx 1-21; gbFJzMXnTKkxNQY7 1;
gdMNAjD2Kyrtqh97 1-8; ghbnmqx4Wz8Rj3Ck 1, 3; ghcTDmb871vtxyWM 1-5, 7-11; gn7Wva1FKAm8HhGZ 21-23; gwkN81FaqdAPQTfW 1-4,
6-8, 10-11; hXT9DqC7vpQW6ZLb 1-9; hf1pF42aPHMv9kLX 1-9, 11-20, 22, 24-36; hmtTqH16pPGzZJax 1, 3-17; hvgw9t1par7YCD3b
1-17; j7J4bKFHq6M2Wfvy 3-6; jD7whNgBrM4GxVvF 1-8; jGrv6B1hNMDJPc3W 5-10; jHXnNA128R4FmwYB 1-5; jQzWyR89avmgpd7H 1, 3-7;
jYAhJZ3BzPMHWx1q 1-5; jnxaGm43ctgQMFhf 1-8; jtdGCBNJq38zxcQ9 1, 3-9; jyxcbwYXG3pL9NnP 1-2; kHfNcJTYj3RtMFGL 1-3;
kJHLMmf3rGaj28d1 1-2; kX94Vmn8tpJzGTD2 1-11; kqXb3Tja4gZfxP7V 4-6; mDbPCdX9AvwYJxGa 7-30; mMFCR2atzKHWDcdQ 1-6;
mPMpkAgjbtR7HBXD 1-19; mc821XQyrpDjGJh9 1-3; n3hjGAFXxTaBfmRz 1-3; ndHRLVB9Kaz7xpqT 1-2; nmrNLB3j4pAPg8fQ 1-2, 4-15;
pKn8wBZcQ62PyJH7 8-24; prbPqjNgMz6FaBX7 1-7; pz7YaGwxXZgcbQPm 7-9; q9FHYznbZLDKtQTP 1-2; qLk1fVdN3rcaDxm6 1;
qMD3CVapwb6nAgx2 35-40, 42-43; qQhkt3L9yKfTFp2D 1-6, 8-13; qYRLwvVKDkTZGP86 20-43; qg8YWNaCdDvV7XZM 1-8, 10;
qyWNYFVZjLX1mtJA 1-19; r67HcKRnkYdzTQZP 1-19; rbg3wR9B7dKf8JmW 1-3; rpLZ287nfKWyJGtb 1-6, 8; t7432ga9HXvFBKzy 1-6;
tD7Rm3GBjdqPzhwC 1-4, 6-8, 10-14, 16-17; tLXnZDMCvJ34mzbF 1; tc1qGLjpgVrJ8WdQ 4-5, 7-15; th1nf93NzQjmVdrD 1-15;
tjN8BgxX7k4GyAfd 4; tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19; vCZk3N8n2dbTyMFV 1-34; vDg3LH1Zyfz49kKr 1-12; vFqn26gPR4NJBfjw
1-4; vKznWFRC7cBHTGwx 3-6, 8, 10-11; vpBwx4KbaWzRZ8V7 1-3, 5-10; vtRzBNFnW6KH3pjT 7-8; vzZPBtWcfm9HY8Ty 4-10, 12-18;
w3L1hH2ya4gCTKAq 1-13; wCGWfTxNgFrB8hDy 1-15; wHtnq3pgjxYVy9FZ 1-2; wRY9NtQp7bPZ1VDz 1-6, 8-12; waYzfQgyrcKV7H9J 1;
wfgxbHpNB2vZJYLW 1-6; wvbng2zVTYRrc9Zd 1-2; x7ybf4LjBD6GJqWn 1-9, 11-15; xDwACkbMZRyNcd1z 1-5; xJbw6RTvtHPncNzp 1-5;
xRNZwp7fWvTY29qV 1-15; xbCaBqNdAn3ZWKDw 1-3, 5; xmYzGavFHWpfbLkt 1-3; xnrg9m6jDMR1kaLN 21-23, 25, 27-29;
xrtBTqdwGDfzK72Q 1-2; xtFwVRc3gmdjLAyC 2; xw86k9LyMNF3YHjr 7-14, 16-21, 23-31; y6kYARn4dpCKavZ8 1-3, 5-6;
ycjXJnNkC8pGWH1Z 2-7; ypQ9d13tMFJPNfvw 1, 3; zDhrvnfPYdt9FXMj 1-4; zjxyZAgPDBMGndRK 1-16, 18-20, 22-24, 26;
zwPKLTQ4W6YkyRqd 3-8.

## Toofbreaker (286430)

- Count: §Count of the 2026-11-05 to 2026-11-19 reports
- Proc: weapon proc 1320269
- Rate: 2 PPM, measured

4,905 procs on 77,305 rolling hits, 49 wearers, 53 runs, in the main hand or the off hand: all hits 2.00 [1.95, 2.07].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.01 [1.90, 2.12],
1,411 procs on 22,187 lone units, 33 wearers. Dual-wield white: 1.99 [1.90, 2.09], 421 procs on 6,684 lone units, 16
wearers. Specials: 2.02 [1.85, 2.19], 480 procs on 7,519 lone units, 34 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 1 lone unit. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 1,631 isolated hits. Left
out: 1 of 54 runs in Cat or Bear Form. Reports from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      39 | 4,411 | 69,942 | 1.99 |
| Rogue   |      10 |   494 |  7,364 | 2.12 |

Reports and fights: 1BNXynxVbH4kGt8R 1-2; 1L9BQW2AcqdP8bfG 1-2; 1cm2k9qHaRdBMpFx 21-36; 2PZCz9c1HgtrX7Th 3-5;
2taZz4vTdfXYAk3b 1-2; 2xY1D3mCwagdy7Wc 1-2, 4-6; 2yqzTJfLN8cvwaDK 1-2, 4-6; 3JbjLGXVzctAmfaD 1-4; 3cWzXTLmvDh2bC7H 1-2;
4YVTAqgG3dtmD8C9 1-23; 4vZxJdFNqTXMRkD8 1-3; 4wWtcgbPhNdMJ2LQ 1-33, 35-37; 6aN7CQwbfTW182vp 1-5, 7-13; 72JjqrWXMdtwKx68
1-3; 84nxvZk9BmHf7Ldr 1-8; 9vWqJjhLGCpdQZKc 1-11; ALpDvxc3GtT7Bnkg 1-2; Cp3ND2J1xWhbFzP7 1-26; DavdRxwG1TZ6rzM7 1-4;
FLK4Vqw3cp6XkRAT 1-15; HkNB7A6tanYcGLWj 1-7; JX4afWrdcw8RkHDt 1-3; KGbxV6BjrZ3QkdDL 2-3; KJGzHwk2Xyfdc9RL 1-10;
LH9kgYvBqyVKadh3 1, 3; LPQd7pfbjDkYZyRz 1-2, 4-5; LcvAbQBGHk6MwgWR 1-3; Mc16FCfkmzRQZ4TH 1-2; P6qAmjBhrNa8b9MR 1-12;
PXbhqkFHpt47RmDy 1-5; PqNHCAjBt2RgrJwQ 1-41; R1jPGYZyTtBkKr8J 5-12; RyrkQYbMVXBFv23J 1-2; VdzPcGgaCnKHY4jF 1, 3-13;
YP3LzZWFTMDGXdyR 1-4; ZqBnHh1CtfDcpwK7 1-6; a7F96kX1mQcbKDvw 1-2, 4-12; ahyJfHkjWGV7P28M 1-26; bA8g6NZK4h9MPHRJ 1-10;
cRtHw9DWZMrKnyvQ 1-4; gTPNZcJ9mG4dRC2X 1-4; hRzbpMrDY3cgQX7n 7; jVhAWnaHJgf3FPmy 1-10, 12-13; qTHXf9y6ag1CMtdp 1-2;
qr7Vt9aW8TpFXJAD 1-2; rTYJ6yWfNcbKXztR 1; tLGanMDfw82FJymz 1-16; tac7hXVRgkbZNMDy 3; vtyhfFz9wPGJdHqM 1-3;
w6A8bVm7vTatMCYr 1-12; zTbHr3BPvKxRCMZg 9; zgG84wjLNY1Pv9tp 1-2; ztpB7m1kbLQ3KdqR 11-14.

## Twilight Forgehammer (281665)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1312176
- Rate: 2.95 PPM, measured

2,280 procs on 18,539 rolling hits, 13 wearers, 16 runs, in the main hand or the off hand: all hits 2.95 [2.84, 3.07].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 3.08 [2.75, 3.44],
311 procs on 2,422 lone units, 3 wearers (a Poisson interval: under 5 wearers). Dual-wield white: 2.37 [2.00, 2.77], 373
procs on 3,785 lone units, 10 wearers. Specials: 3.20 [2.97, 3.38], 290 procs on 2,174 lone units, 8 wearers. Two-roll
abilities, priced as one roll: 7.49, 34 procs on 109 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 716 isolated hits. Dropping the swapped stretches (8 active minutes, 1% of hits, in 1 run) moved the
read from 2.93 to 2.95. 1 run logged in two reports counts once. The paladin check, which sets no rate: lone white
swings 4.63, 39 procs on 202 lone units, 4 wearers. By report date: before 2026-11-18 2.91 (927 procs); from 2026-11-18
to push 114313 2.98 (1,353 procs). Reports from 2026-11-09 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       7 | 1,529 | 12,603 | 2.91 |
| Rogue   |       5 |   744 |  5,893 | 3.03 |
| Shaman  |       1 |     7 |     43 | 3.91 |

Reports and fights: 1qcv4MKB7R9bWxJZ 1-9; 6b4F7HrzNTnxRMXd 1-13; 84VDXdcKZhHMPBAm 1-5; 8G3bmgKTvBxqfJQL 4-20, 22-25;
b4hn6GgpFLtJ1N2X 1-13; bLmVHN7BaP8DvX9Z 1-16; C8z3gcnBYGZQXJv2 1, 3-6; GmjR9hMTvBDtkNQ6 1-4, 7; hLn1YZ63qHVx49TF 1-6;
hYHt6kLTgApm8RMw 1-2; m8RLWaY7bh9jHKwD 2; qCgr4VRyj8HcaznF 1-2; qYC7g4Har6pKz8QT 1; RaP2LJFGrWzZtkf4 1-2;
rTYJ6yWfNcbKXztR 1; Ztv8xkT6dmQ7n91C 1.

## Twilight Maul (273841)

- Proc: weapon proc 1292679, logged as 1292679; rolls on main hand
- Sim rate: 1.75 PPM, observed; measured 1.74 PPM [1.58, 1.96], 7 wearers, 7 runs, 373 procs on 3905 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 5 of 12 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights    | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1WKqM9x42jcLnYCN | 1-36      | Shaman  | mh   |            |   105 | 1102 | 1.73 |          |
| 1hpgVjWH3GZBLYKw | 13, 15-25 | Shaman  | mh   |            |    25 |  207 | 2.20 |          |
| 1nLR3XCxjcAdpKtB | 1-18      | Shaman  | mh   |            |    57 |  688 | 1.51 |          |
| 1q2tTDcaFMnfzvkB | 1         | Shaman  | mh   |            |    17 |  172 | 1.80 |          |
| 2NQC6YJtXRPbMzg1 | 1-13      | Shaman  | mh   |            |    72 |  803 | 1.63 |          |
| BzGXfjMLJ2RyCwHm | 1-7       | Paladin | mh   |            |    62 |  675 | 1.67 |          |
| RqpmbCcYr8BJZWwK | 1-2       | Shaman  | mh   |            |    35 |  258 | 2.47 |          |
| 2NBRf9AwdZTjDQYa | 7-13      | Druid   | mh   |            |     4 |  168 | 0.43 | feral    |
| 6WzAhcVgfLCt78nX | 1-8       | Druid   | mh   |            |    34 | 1139 | 0.54 | feral    |
| GAVFWPJBtmn2x6ry | 1-6       | Druid   | mh   |            |    20 |  960 | 0.38 | feral    |
| VYb6f27NazpmCch3 | 1-2       | Druid   | mh   |            |    13 |  558 | 0.42 | feral    |
| b69d1Dhw4JkxBHTt | 1-14      | Druid   | mh   |            |    40 | 1426 | 0.51 | feral    |

## Venomstrike (6469)

- Proc: weapon proc 29653, logged as 29653; rolls on ranged
- Sim rate: 2.05 PPM, measured; measured 2.05 PPM [1.88, 2.23], 19 wearers, 23 runs, 667 procs on 8127 hits
- Reports from 2026-11-06 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.53 -> 2.05 (113 active minutes, 26% of hits, in 4 runs)

| Report           | Fights  | Class   | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------- | ------- | ------ | ---------- | ----: | ---: | ---: | -------- |
| 12VhLkPa98DYFH4g | 1-6     | Warrior | ranged |            |     1 |    6 | 4.17 |          |
| 1Rk8y7VKNWapF2dt | 3-4     | Rogue   | ranged |            |     0 |    4 | 0.00 |          |
| 1jKJfWybdV2vBQ8A | 1-2     | Hunter  | ranged |            |    22 |  172 | 3.20 |          |
| 1qNaRQCwLmDPcJfx | 1-2     | Hunter  | ranged |            |    25 |  358 | 1.75 |          |
| 23dT9fk1DqcaJmwn | 1-7     | Hunter  | ranged |            |    58 |  743 | 1.95 |          |
| 261hbMqpgzLRKxwY | 1-2     | Warrior | ranged |            |     3 |   21 | 3.57 |          |
| 2axJHNwqQdtW1jnF | 1-3     | Hunter  | ranged |            |    68 |  737 | 2.31 |          |
| 3CyQk69LzJjTn2X4 | 1-5     | Warrior | ranged |            |     0 |    0 |      |          |
| 3npHxB84DT2zRtmN | 1-4, 6  | Warrior | ranged |            |     0 |    0 |      |          |
| 4dvhFmKbqkTQ1Wpc | 1-3     | Rogue   | ranged |            |     0 |    1 | 0.00 |          |
| 7W3YTHD9kyPzFvtc | 1-4     | Warrior | ranged |            |     0 |    0 |      |          |
| 7ZxcPNTjqvfzQHgX | 1-2     | Hunter  | ranged |            |    13 |  124 | 2.62 |          |
| 7ypMG8JVjRgZ1xHQ | 1-9     | Hunter  | ranged |            |    44 |  668 | 1.65 |          |
| GPhAqKkF8tHBpfxT | 1-3     | Hunter  | ranged |            |    37 |  542 | 1.71 |          |
| GmZaJgjW2R97cVnL | 6       | Hunter  | ranged |            |    37 |  351 | 2.64 |          |
| RQxLGd1agBDt6hXP | 1-4     | Hunter  | ranged |            |    65 |  737 | 2.20 |          |
| WhwcvTCZFPDxanyr | 22-65   | Hunter  | ranged |            |    78 |  795 | 2.45 |          |
| YCqznXvDMQA4d718 | 1-7     | Hunter  | ranged |            |    96 | 1102 | 2.18 |          |
| bzWMAC2mf7Z6Q3LV | 1-9     | Hunter  | ranged |            |    66 |  971 | 1.70 |          |
| d6bwAkLKjGq7pVMy | 1       | Hunter  | ranged |            |    15 |  211 | 1.78 |          |
| jM3dtQ7vcGZ9fhWn | 1-5     | Hunter  | ranged |            |    35 |  509 | 1.72 |          |
| kBzVYb3Cy7QwnKMD | 1-6     | Hunter  | ranged |            |     4 |   71 | 1.41 |          |
| rXvTBhVgpZmc1kbP | 1, 3-13 | Warrior | ranged |            |     0 |    4 | 0.00 |          |

## Verigan's Fist (6953)

- Proc: weapon proc 1322218, logged as 1322218; rolls on main hand
- Sim rate: 2.2 PPM, measured; measured 2.21 PPM [2.04, 2.34], 10 wearers, 10 runs, 963 procs on 8186 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1CYTfP3pFDcaRrJm | 1-8         | Paladin | mh   |            |   138 | 1213 | 2.13 |          |
| 1HTPQGZya7pAz3qh | 3-32, 34-36 | Paladin | mh   |            |    78 |  682 | 2.14 |          |
| 1pgjDMXka9ZFxy8f | 1-3         | Paladin | mh   |            |    40 |  302 | 2.48 |          |
| 1rYynxW6AGfcQz9t | 5-6, 8-17   | Paladin | mh   |            |   211 | 1679 | 2.36 |          |
| 26cnTVbFtvQfzANd | 1           | Paladin | mh   |            |   137 | 1041 | 2.47 |          |
| 2AxRVTac9WCmG6f3 | 1-6         | Paladin | mh   |            |    44 |  354 | 2.33 |          |
| 7V8ykCfL1htHwzKD | 1-3         | Paladin | mh   |            |    54 |  528 | 1.92 |          |
| QmjBxrzvah2XTVR1 | 3-8         | Paladin | mh   |            |   146 | 1174 | 2.33 |          |
| VQaYAf1Wxvm6ndHK | 6-11        | Paladin | mh   |            |    59 |  584 | 1.89 |          |
| hLHaBZ29kyA4MCpj | 1-12        | Paladin | mh   |            |    56 |  629 | 1.67 |          |

## Vibroblade (9485)

- Proc: weapon proc 1293005, logged as 1293005; rolls on main hand, off hand
- Sim rate: 1 PPM, observed; measured 1.01 PPM [0.69, 1.21], 6 wearers, 6 runs, 111 procs on 4128 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| BtN6VMJRjgp1Ay9G | 1-3    | Paladin | mh   |                             |     4 |  287 | 0.52 |          |
| MKx2v8JfHVc6kmND | 1-6    | Rogue   | oh   | timing + otsu for 59 swings |     9 |  678 | 0.50 |          |
| Pzk84bCLKYcvN9fx | 1-10   | Rogue   | oh   | timing                      |    12 |  475 | 0.95 |          |
| QBCAa6JwknGvcMrP | 1-2    | Rogue   | oh   | timing + otsu for 22 swings |    22 |  612 | 1.35 |          |
| TacY2wkjAgXynNHJ | 1-17   | Paladin | mh   |                             |    48 | 1452 | 1.24 |          |
| fA1GLbc6gw9nKHM2 | 1-16   | Rogue   | oh   | timing + otsu for 16 swings |    16 |  624 | 0.96 |          |

## Waveslicer (18324)

- Count: §Count of the 2026-11-13 to 2026-11-17 reports
- Proc: weapon proc 1302166
- Rate: 1.95 PPM, measured

1,688 procs on 16,901 landed hits, 5 wearers, 7 runs: 1.93 [1.84, 2.04].

| Report           | Fights | Class   | Hand      | Procs |  Hits | Rate |
| ---------------- | ------ | ------- | --------- | ----: | ----: | ---: |
| VYBjgJx8aZtRhmpy | 1-43   | Shaman  | main hand |   691 | 6,951 | 1.92 |
| AyTmpNtxB9jbnJ2H | 4-44   | Shaman  | main hand |   318 | 3,245 | 1.90 |
| q23Tz8hnZGkvbLx6 | 1-18   | Shaman  | main hand |   200 | 1,887 | 2.05 |
| NrpM7DJQBKWqbzVF | 1-7    | Paladin | main hand |   164 | 1,784 | 1.78 |
| DpVmH4WRxL72nGwd | 3      | Shaman  | main hand |   127 | 1,275 | 1.93 |
| H2Zg1jRm93PpTqK4 | 1-3    | Shaman  | main hand |   135 | 1,225 | 2.13 |
| KCYXdTfNgGHczbLw | 5-6    | Shaman  | main hand |    53 |   534 | 1.92 |

## Witchfury (13051)

- Proc: weapon proc 18214, logged as 18214; rolls on main hand
- Sim rate: 1 PPM, observed; measured 1.02 PPM [0.87, 1.16], 6 wearers, 6 runs, 215 procs on 4525 hits
- Reports from 2026-11-15 to 2026-11-16

| Report           | Fights    | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 6vFAMDQZ2RmaxC7n | 1-3       | Paladin | mh   |            |    37 |  714 | 1.11 |          |
| HXg2W4bTFDjcRJ87 | 1-3       | Paladin | mh   |            |    34 |  958 | 0.76 |          |
| RKvbj4hcCTJWNk3Q | 4-5       | Paladin | mh   |            |    48 | 1013 | 1.02 |          |
| cLkwmZbzFDKdPhYn | 11, 13-14 | Paladin | mh   |            |    18 |  443 | 0.87 |          |
| f43cC7V8BvrMXyQb | 1-4       | Paladin | mh   |            |    51 |  915 | 1.19 |          |
| fBrKy1M2T6q9LC3g | 1-7       | Warrior | mh   |            |    27 |  482 | 1.20 |          |

## Wolfsbane (267369)

- Proc: weapon proc 1282503, logged as 1282503; rolls on main hand
- Sim rate: 2.25 PPM, measured; measured 2.26 PPM [2.15, 2.36], 29 wearers, 29 runs, 1857 procs on 14529 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 18zavRLKVhycxjWZ | 1-2      | Paladin | mh   |            |    41 |  399 | 1.81 |          |
| 1FNGzPgAaX6wDpBt | 1-7      | Paladin | mh   |            |    54 |  488 | 1.95 |          |
| 1NBKq2mDWghwR69z | 1-4, 6   | Paladin | mh   |            |    37 |  380 | 1.72 |          |
| 1mAfJLYwZXHMzVtd | 1-2      | Paladin | mh   |            |    66 |  462 | 2.52 |          |
| 1mWTYf2wPgCGqkMp | 1-2      | Paladin | mh   |            |    52 |  480 | 1.91 |          |
| 1qjDRYwtHxyJPfQn | 1-5      | Paladin | mh   |            |   101 |  733 | 2.43 |          |
| 6GY7rBk9gHJfVMFm | 1-2      | Paladin | mh   |            |   111 |  911 | 2.15 |          |
| 6GY7rBk9gHJfVMFm | 1-2      | Paladin | mh   |            |   118 |  813 | 2.56 |          |
| 8DQBRhWyXrjHzktc | 1, 3-4   | Paladin | mh   |            |    81 |  610 | 2.34 |          |
| Br9cPVnJgCZFwpmK | 1-2      | Paladin | mh   |            |    75 |  484 | 2.73 |          |
| GPhAqKkF8tHBpfxT | 1-3      | Paladin | mh   |            |    71 |  497 | 2.52 |          |
| HtQr7kYaNq9MGgvZ | 1-12     | Paladin | mh   |            |    64 |  548 | 2.06 |          |
| RvcrA3P9jznxhHVQ | 4-7      | Paladin | mh   |            |    65 |  511 | 2.24 |          |
| TAgkJxp6dN9M2BtQ | 1        | Paladin | mh   |            |    40 |  359 | 1.97 |          |
| VYb6f27NazpmCch3 | 1-2      | Paladin | mh   |            |    62 |  421 | 2.60 |          |
| Vhk7TNKgMzdH2C8A | 1-3      | Paladin | mh   |            |    73 |  557 | 2.31 |          |
| YCqznXvDMQA4d718 | 1-7      | Paladin | mh   |            |    53 |  444 | 2.11 |          |
| b69d1Dhw4JkxBHTt | 1-14     | Paladin | mh   |            |   133 |  987 | 2.38 |          |
| b6C24Jd3V8WwpFx7 | 2-4      | Paladin | mh   |            |    14 |   76 | 3.25 |          |
| bnRJKL9FTMq2zkQG | 1        | Paladin | mh   |            |     4 |   33 | 2.14 |          |
| bxtk19P4qVQjmYFC | 1-2      | Paladin | mh   |            |    90 |  828 | 1.92 |          |
| bzWMAC2mf7Z6Q3LV | 1-9      | Paladin | mh   |            |    34 |  395 | 1.52 |          |
| dP8M1ZqnTWR4Brf6 | 1        | Paladin | mh   |            |    67 |  476 | 2.48 |          |
| df8Mh4tqkJxHnCa3 | 1-5      | Paladin | mh   |            |   103 |  769 | 2.36 |          |
| jLTw4rQd1BKgMkhy | 1-2, 4-6 | Paladin | mh   |            |    66 |  482 | 2.42 |          |
| nDrBjQ74hH2gpvTw | 2        | Paladin | mh   |            |    10 |  102 | 1.73 |          |
| tZQTyMXr3BY7jxHg | 1-3      | Paladin | mh   |            |    80 |  551 | 2.56 |          |
| yLjkTYJft9DV2FGM | 1-2      | Paladin | mh   |            |    43 |  352 | 2.16 |          |
| yLjkTYJft9DV2FGM | 1-2      | Paladin | mh   |            |    49 |  381 | 2.27 |          |

## X'caliboar (10758)

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Proc: weapon proc 1293787
- Rate: 1.8 PPM, measured

715 procs on 7,174 rolling hits, 8 wearers, 9 runs, in the main hand: all hits 1.81 [1.73, 1.97]. The rate is all hits,
which the one-hand white rate agrees with. The specials stay under 50 procs. One-hand white: 1.90 [1.72, 2.29], 113
procs on 1,079 lone units, 6 wearers. Specials: 1.61 [1.06, 2.01], 48 procs on 543 lone units, 6 wearers. Two-roll
abilities, priced as one roll: 4.17, 22 procs on 96 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 146 isolated hits. Dropping the swapped stretches (34 active minutes, 16% of hits, in 1 run) moved the
read from 1.52 to 1.81. The paladin check, which sets no rate: lone white swings 2.20, 195 procs on 1,610 lone units, 21
wearers. By report date: before 2026-11-18 1.81 (273 procs); from 2026-11-18 to push 114313 1.81 (442 procs). Reports
from 2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   710 | 7,130 | 1.81 |
| Hunter  |       2 |     5 |    44 | 2.07 |

Reports and fights: 1WHqx7bPa8mkGydJ 1, 3; aPDTGRxFmMhZAzYW 1; HRnmMtD7Fp41wb2g 1; K8rhAtLjcz7gdHkn 6-7;
maYP7WvxJ2TzGtDH 1-3; mM3pZD8H7CqBNYwJ 1; P1yKzX3Q8NgFYxZH 1-3; WYZRJjhp7LvDqKNB 1-7; X4jWcyZhzfmJKG9F 1.

## Count of the 2026-11-13 to 2026-11-17 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  names; 17 reports in all
- Game build: 1.60.1.70938 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-17
- Pull request: #1091

The count is EV-0146 §Method, run on these weapons.

## Count of the 2026-11-12 to 2026-11-18 reports

- Sources: Warcraft Logs reports (forever.warcraftlogs.com, zone 3000), every fight in which the wearer's gear
  snapshot holds 11809 in the main hand: 1X37tZRzgkmd9rNP, 1a3Qn7JpLk8AH2hV, 1nTQBK4hmcaLbFZV, 1z6PXMGdQBDvrCqx,
  2wFyAGLf1CqZKnaz, 2wZ3P6y9N4f7kXVq, 3R69PBjxZvcfbnX8, 3qfDVTZ1J29zPChn, 4BFrCYZajATmLfgn, 4f2ygdYCzFAvqc6K,
  6CN9XfdkHq148yZQ, 6VDthbQzAJxqpn7d, 6jHn3xcCZYW9GkMy, 6kJNxp39PWqrghmG, 6yHgkT8LqP42NmxY, 7ChG2VXaALzcw9pW,
  7mnVXCgpyxkbtZGq, 7tbWfYqJFd4vzL3p, 8FBqfvLyWTdKxbZh, 8LaNq1JXfZzy3Dj9, 9dmW4zX2YbnK6CGp, AYwLfvqh9QyNj6zJ,
  B6bvH1fQ4kAcyRzt, BHqXftrKD6AadWh1, Bv6mxRb1fKG9ZgYd, DQb2rzfnNyL148BR, Dagn1fC64vQRwK9N, FJdrDz9w1Vxaf2cT,
  FN8krKfCb3BXWmwV, FhG3AxYL9ZNP4RvV, FtcXATwrHDd9vQ83, HMdLhWY3pNbnXPtF, JKPwHm7fnCp3L14c, JQwbjnVYyGcP4rg8,
  JrqFhHP6wxV7n9yX, KfjzWvhFNCDMLPwx, LQ9zMDqpmknBgAX8, Lt3QCJGKXZYTBFkD, LvDw6BjHCVGgzJyb, M1tPx9DBZGYgCa4f,
  NQv1MAX6mrxjC2nb, Nxtj7m6zcFX9Aq3Z, PdNq3JxntCKgavkm, PtKXrwfj8zBHm2Gy, Pxn7LYfMX4Qd6vNr, TmQyqnM6hN217LZC,
  YCbcQHW378fnMvXA, YZy7L3P6AVhWJN1M, ZHpKVnhPDGyaTrvw, ZJgFdtvGTYQVCqyb, ZwkY3VX4H2aLWpDt, aLwy1zjnrYFhvpDq,
  aWLTAv36Kx9HVPkY, bg169rGdXTzDhLnj, byvnpga9tMq3BfX6, cb9gzXjCK3BfH7qM, gYm2HAtf4bnC8QZT, hQY9xzt7Mr4nG6Ap,
  jKQDqkwmx46ZLBFg, k1gAvx7zVaPpLhYC, kvhJGjQNAWRgBwrm, mbpQYGL8jaqgdrAz, mzVGhWHB3ZyATfC4, p4fMBXZaJz8ctTP6,
  p4waHBZgjkA9mDLn, pcKhkTd9wzArGvfm, pzBvhjtCD9VRf2mP, qF9L7VZHR1QG8r2m, qbrGgBZfzLHFW12k, rPpHYfq3tNnwJRMd,
  tTRm37A4YwgM8nCG, tmHdvqxYj1z28PFy, w6BX8kh4ADbRW2yg, yfKb1MthLFDc4ANG, zmDY4LVX7N8vh9aR; 52 wearers in 78
  wearer-reports (37 druids in 62, 9 paladins in 9, 5 hunters in 5, 1 warrior in 2)
- Game build: 1.60.1.70888 (6kJNxp39PWqrghmG, DQb2rzfnNyL148BR), 1.60.1.70938 (the reports up to 2026-11-17 15:18 UTC)
  and 1.60.1.70938 or 1.60.1.70959 after it (2wFyAGLf1CqZKnaz, 6CN9XfdkHq148yZQ, 8FBqfvLyWTdKxbZh, B6bvH1fQ4kAcyRzt,
  FhG3AxYL9ZNP4RvV, KfjzWvhFNCDMLPwx, LQ9zMDqpmknBgAX8, LvDw6BjHCVGgzJyb, M1tPx9DBZGYgCa4f, Nxtj7m6zcFX9Aq3Z,
  PtKXrwfj8zBHm2Gy, ZJgFdtvGTYQVCqyb, aLwy1zjnrYFhvpDq, aWLTAv36Kx9HVPkY, jKQDqkwmx46ZLBFg, mbpQYGL8jaqgdrAz,
  mzVGhWHB3ZyATfC4, p4fMBXZaJz8ctTP6, pzBvhjtCD9VRf2mP, qbrGgBZfzLHFW12k and part of yfKb1MthLFDc4ANG), dated by when
  the CDN first listed each build against each event's time; the US and EU listed 70959 from 2026-11-17 15:19 UTC
  while CN, KR and TW still listed 70938 on 2026-11-18. The hotfix each player ran is not recorded
- Date: 2026-11-12 14:54 to 2026-11-18 14:18 UTC
- Pull request: #1104

The count is EV-0151 §Method, run on Flame Wrath.

## Count of the 2026-11-05 to 2026-11-19 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every indexed wearer of the weapons; 1,856 reports in all
- Game build: 1.60.1.70790 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114196; the hotfix
  each player ran is not recorded
- Date: 2026-11-05 to 2026-11-19
- Pull request: #1108

The count is EV-0153 §Method, run on these weapons.

## Count of the 2026-11-09 to 2026-11-19 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 263 reports in all
- Game build: 1.60.1.70862 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114281; the hotfix
  each player ran is not recorded
- Date: 2026-11-09 to 2026-11-19
- Pull request: #1126 for Iceblade Hacker and Loksey's Training Stick, #1130 for The Blackrock Slicer

The count is EV-0157 §Method, run on these weapons by the same sweep, `runs/2026-11-19-2309` in
`forever-sweep-auto`.

## Count of the 2026-11-06 to 2026-11-21 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 652 reports in all
- Game build: 1.60.1.70790 to 1.60.1.71037 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC,
  71037 from 2026-11-20 18:23 UTC); 6 reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database
  is 1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-21
- Pull request: #1156

The count is EV-0153 §Count of the 2026-11-06 to 2026-11-21 reports, run on these weapons. Firebelcher's and Mana
Channeling Wand's lone units are wand shots. On Kang the Decapitator, Sul'thraze the Lasher and Supercharger Battle Axe
every class past the first is one wearer.

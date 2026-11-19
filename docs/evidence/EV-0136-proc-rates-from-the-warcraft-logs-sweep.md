# EV-0136: Proc rates of 103 items and a set bonus from the Warcraft Logs sweep

EV-0138 replaces what this file says of Windfury Totem's strike after push 114115: the strike grants rage and spends no
warrior Flurry charge. EV-0146 replaces the rates of Julie's Dagger, Killmaim and Waveslicer with the reports up to
2026-11-17 added. EV-0151 replaces the Flame Wrath rate, whose count took Fire Shield's hits on attackers for procs.
EV-0153 replaces the rates of Barbaric Crossbow, Bloodrazor, Bow of Searing Arrows, Burning War Axe, Coldrage Dagger,
Cursed Felblade, Darkwater Talwar, Dazzling Longsword, Destiny, Electrocutioner Leg, Fang of the Crystal Spider,
Felstriker, Fiery War Axe, Fist of the Damned, Glutton's Cleaver, Grimclaw, Hammer of Expertise, Hanzo Sword, Iron
Knuckles, Ironforge Greathammer, Killmaim, Loksey's Training Stick, Lord General's Sword, Nightblade, Pendulum of Doom,
Phantom Blade, Sang'thraze the Deflector, Satyr's Lash, Spellsever Crossbow, Strike of the Hydra, Sul'thraze the Lasher,
Sword of Decay, Tainted Pierce, Taran Icebreaker, Thornspike and Toofbreaker with the reports up to 2026-11-19, counted
on every hit the class census of EV-0154 finds rolling a weapon proc.

- Claim: each item and set bonus below procs at the rate its section gives, priced on each landed hit's weapon base speed;
  a rate whose item has 5 or more wearers and a 95% interval within 10% of it either side is measured, every other one
  observed; and Windfury Totem's strike counts as a melee swing for the procs whose rows hear melee swings, Unsealed
  Wrath excepted, while in game it grants no rage.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section's table names; 721
  reports in all for the rates, and the 209 reports the Windfury Totem section lists
- Game build: 1.60.1.70425 to 1.60.1.70938 by report date; each item's section gives its date span, and the hotfix each
  player ran is not recorded
- Date: 2026-10-08 to 2026-11-16
- Pull request: #1065

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

The method reproduces Destiny (EV-0132, 1.18), Heart of Wyrmthalak (EV-0133, 1.92 against 2.00), Skullforge Reaver
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
- Flame Wrath (11809), 8 wearers, 421 procs: the fire shield went up 297 times and refreshed 130 times, 195 windows lasted
  exactly 15 s, and it dealt damage 2,292 times inside its windows and 47 outside, never a crit. 8 procs came 0 to 1 s
  apart, so the row's 5 s category cooldown does not lock out procs.
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

In game the strike grants no rage. Cast events carry the warrior's rage before the cast's cost, so the gain between two
casts is read straight off them. Over gaps under 600 ms with no damage taken and no resource change, 88 of 90 strikes
from 13 warriors added nothing over the same warrior's gaps opened by the same ability without a strike. The strikes
dealt 1 to 509 damage, median 204, where a white swing of about 200 damage gives 10 to 19 rage. The owner calls this a
known game issue that is to be fixed, so the sim keeps the strike's rage as a main-hand swing's. Those runs are
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

- Proc: weapon proc 1291551, logged as 1291551; rolls on ranged
- Sim rate: 3.45 PPM, measured; measured 3.44 PPM [3.17, 3.71], 10 wearers, 12 runs, 630 procs on 5491 hits
- Reports from 2026-11-07 to 2026-11-16

| Report           | Fights      | Class   | Slot   | Hand split | Procs | Hits |   PPM | Left out |
| ---------------- | ----------- | ------- | ------ | ---------- | ----: | ---: | ----: | -------- |
| 1AfDWYQT7Xjd24K3 | 13-14, 16   | Hunter  | ranged |            |    84 |  654 |  3.85 |          |
| 1RrDAhp7ZvVYc9P3 | 1-6         | Rogue   | ranged |            |     0 |    0 |       |          |
| 1W9N26dHfnPV4G7A | 1, 3-9      | Hunter  | ranged |            |    64 |  554 |  3.47 |          |
| 1y6GV7KLTZPrJH2w | 1-3         | Hunter  | ranged |            |    70 |  604 |  3.48 |          |
| 2GbXqD7KcWTfRnYa | 1-2         | Hunter  | ranged |            |   110 |  795 |  4.15 |          |
| 2MHXV8Y9jx43vhpk | 3-12, 14-18 | Warrior | ranged |            |     2 |   43 |  1.40 |          |
| 2tJkzh16X38wrQfP | 1-3         | Hunter  | ranged |            |   114 | 1052 |  3.25 |          |
| Td29Nzahj1cmGMWZ | 1-2         | Hunter  | ranged |            |    79 |  717 |  3.31 |          |
| YHh2t9agcvQjTAZq | 2-3         | Hunter  | ranged |            |    48 |  470 |  3.06 |          |
| kBHbV3DZjamvfJY4 | 1-3         | Rogue   | ranged |            |     1 |    3 | 10.00 |          |
| nALy2WrwHC9qdMFm | 1-2         | Hunter  | ranged |            |    58 |  599 |  2.90 |          |
| wmT1zqYBat3fHDLy | 1-5, 8-9    | Warrior | ranged |            |     0 |    0 |       |          |

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

## Blackblade of Shahram (12592)

- Proc: weapon proc 16602, logged as 16602; rolls on main hand
- Sim rate: 1.15 PPM, observed; measured 1.15 PPM [0.96, 1.36], 4 wearers, 4 runs, 240 procs on 3576 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights                   | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out  |
| ---------------- | ------------------------ | ------- | ---- | ---------- | ----: | ---: | ---: | --------- |
| 1JLq46yg7wYNhan9 | 17-28                    | Warrior | mh   |            |    46 |  906 | 0.87 |           |
| 4f2ygdYCzFAvqc6K | 10, 12-14, 16-18         | Paladin | mh   |            |    92 | 1291 | 1.22 |           |
| D463FZBX1C9maJ7b | 7-8                      | Paladin | mh   |            |    43 |  680 | 1.08 |           |
| NXMZV9xjvWpB7GtP | 13-107, 109-116, 118-122 | Paladin | mh   |            |    59 |  699 | 1.45 |           |
| NdTXJzkWHLZrVQYm | 1-3                      | Warrior | mh   |            |    46 |  907 | 0.87 | duplicate |

## Blade of the Wretched (10803)

- Proc: weapon proc 18088, logged as 18088; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.94 PPM [1.76, 2.18], 7 wearers, 7 runs, 1117 procs on 16411 hits
- Reports from 2026-11-12 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 6kN78G2xfqy3vwTP | 6        | Rogue   | mh   | timing + otsu for 4 swings   |     8 |  101 | 2.25 |          |
| 7M3gGcH6bprxKqZY | 1-3, 5-8 | Warrior | mh   |                              |   140 | 1834 | 2.18 |          |
| 9H72ZkdBGfhCy6P3 | 1-21     | Paladin | mh   |                              |   250 | 3517 | 2.03 |          |
| 9zMva8xTVhFkHPAt | 1-10     | Paladin | mh   |                              |   271 | 4090 | 1.89 |          |
| D463FZBX1C9maJ7b | 4-8      | Paladin | mh   |                              |    92 | 1068 | 2.46 |          |
| N2m8nPTGJQrj1DhH | 4-5      | Rogue   | mh   | timing + otsu for 69 swings  |   142 | 1909 | 2.12 |          |
| VwyW3ZP1H6RTpdmf | 1-5      | Warrior | mh   | timing + otsu for 250 swings |   214 | 3892 | 1.57 |          |

## Bloodrazor (809)

- Proc: weapon proc 17504, logged as 17504; rolls on main hand
- Sim rate: 1.05 PPM, observed; measured 1.03 PPM [0.76, 1.31], 5 wearers, 5 runs, 361 procs on 7820 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| AZchwX93DmyHLbv6 | 1-3    | Warrior | mh   | timing + otsu for 135 swings |    56 | 1872 | 0.66 |          |
| NWvgFp3kfR2bHyYa | 1-35   | Rogue   | mh   | timing + otsu for 113 swings |    72 | 1343 | 1.19 |          |
| RW1BNk6Ybyh3XjLQ | 1-21   | Warrior | mh   |                              |   143 | 2246 | 1.41 |          |
| jJWt7Dnm4qc9rzQd | 1      | Warrior | mh   | timing + otsu for 52 swings  |    55 | 1626 | 0.75 |          |
| kJ3WdhMLqv7K2Nyf | 1-3    | Rogue   | mh   | timing + otsu for 81 swings  |    35 |  733 | 1.06 |          |

## Bloodspiller (7753)

- Proc: weapon proc 18200, logged as 18200; rolls on main hand
- Sim rate: 0.95 PPM, observed; measured 0.94 PPM [0.80, 1.15], 6 wearers, 6 runs, 130 procs on 3065 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 0.77 -> 0.94 (29 active minutes, 18% of hits, in 2 runs)

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1YQTjgZh2KPBGCbN | 1-3    | Hunter  | mh   |            |    53 | 1329 | 0.89 |          |
| 2ywhMnJq4Y3zaWtg | 1-9    | Warrior | mh   |            |    18 |  445 | 0.90 |          |
| Fx34Wc78qhnPmVz6 | 1-5    | Shaman  | mh   |            |     2 |   55 | 0.81 |          |
| LxdYWR47ZFbTNBhM | 1-2    | Shaman  | mh   |            |    28 |  530 | 1.17 |          |
| Qtca4MnwCRLKPdN8 | 1      | Paladin | mh   |            |     9 |  112 | 1.79 |          |
| YFnCf6cWNRpgVMHT | 17-31  | Warrior | mh   |            |    20 |  594 | 0.75 |          |

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

- Proc: weapon proc 1292573, logged as 1292573; rolls on main hand
- Sim rate: 2.8 PPM, observed; measured 2.78 PPM [2.41, 3.38], 7 wearers, 7 runs, 180 procs on 1944 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 6 of 13 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 3rbcnjTMRfwxv1mh | 1-10   | Hunter  | mh   |            |     9 |  101 | 2.67 |          |
| 79RVgZWM8dF2wkCp | 1-6, 8 | Hunter  | mh   |            |     5 |   65 | 2.31 |          |
| 7zcat1AyFnrJ4BQR | 10-19  | Hunter  | mh   |            |    10 |   95 | 3.16 |          |
| BzGXfjMLJ2RyCwHm | 1-7    | Hunter  | mh   |            |     2 |   36 | 1.67 |          |
| FAf9V76BctbG3mR4 | 1      | Warrior | mh   |            |    13 |  104 | 3.75 |          |
| K9WGdrx248FvaDVY | 1-3    | Hunter  | mh   |            |     4 |   24 | 5.00 |          |
| ywDgt7KhbjmAL82V | 1-11   | Paladin | mh   |            |   137 | 1519 | 2.71 |          |
| 3rgax8hkZzJjXKp2 | 4-5    | Druid   | mh   |            |    22 |  480 | 1.38 | feral    |
| 8BjkHKdLY1NWxFVw | 1      | Druid   | mh   |            |    26 |  660 | 1.18 | feral    |
| AH4pXcWv2BGFZk8f | 1-16   | Druid   | mh   |            |    63 | 1780 | 1.06 | feral    |
| AzbN8GY9DnBrH4aW | 1-18   | Druid   | mh   |            |    11 | 1171 | 0.28 | feral    |
| JGFcm9zq48CArhgk | 3-6    | Druid   | mh   |            |    32 |  832 | 1.15 | feral    |
| XGJ38LhbnH7FPdKq | 1-3    | Druid   | mh   |            |    33 | 1114 | 0.89 | feral    |

## Bow of Searing Arrows (2825)

- Proc: weapon proc 29638, logged as 29638; rolls on ranged
- Sim rate: 5.4 PPM, measured; measured 5.42 PPM [4.94, 5.81], 5 wearers, 5 runs, 701 procs on 2873 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights    | Class  | Slot   | Hand split | Procs | Hits |  PPM | Left out  |
| ---------------- | --------- | ------ | ------ | ---------- | ----: | ---: | ---: | --------- |
| 6Krp8bkfx4cRjBaz | 1, 3-15   | Hunter | ranged |            |    97 |  429 | 5.02 |           |
| AhLc6aGKMCdvBgDW | 1-5       | Hunter | ranged |            |   233 |  940 | 5.51 |           |
| H2GvZtVYNy4QzqLw | 12-21     | Hunter | ranged |            |   150 |  538 | 6.20 |           |
| KGQHdbFCZpt1YX23 | 18, 20-23 | Hunter | ranged |            |    99 |  484 | 4.55 |           |
| cJhVzAdLx3BHqpkT | 1-4       | Hunter | ranged |            |   122 |  482 | 5.62 |           |
| 2AzZR61XcPdKgQrB | 6, 8-14   | Hunter | ranged |            |    99 |  484 | 4.55 | duplicate |

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

- Proc: weapon proc 18199, logged as 18199; rolls on main hand
- Sim rate: 1.05 PPM, observed; measured 1.04 PPM [0.92, 1.19], 8 wearers, 8 runs, 242 procs on 4251 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 3cKTmVyb1ZPwFXkf | 1, 3-13     | Shaman  | mh   |            |    25 |  456 | 1.00 |          |
| 7XwRKY4qkjDt3ryz | 2-3         | Paladin | mh   |            |    38 |  758 | 0.91 |          |
| Fgy2bt81NL9CAzaR | 1-13, 15    | Shaman  | mh   |            |    13 |  200 | 1.18 |          |
| KRCBMJvg72fHVwWL | 1-17        | Warrior | mh   |            |    21 |  394 | 0.97 |          |
| LwTzGJXrN3HWqCfD | 1-5         | Warrior | mh   |            |    35 |  443 | 1.44 |          |
| RBmtAMLxz78Pj2Wh | 7-10        | Shaman  | mh   |            |    23 |  485 | 0.86 |          |
| VQFaRpT6J2xLDbZg | 1, 3-5      | Warrior | mh   |            |    20 |  289 | 1.26 |          |
| Wf8ghC6TtQKrkXdv | 8-10, 12-15 | Shaman  | mh   |            |    67 | 1226 | 0.99 |          |

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

- Proc: weapon proc 1293790, logged as 1293790; rolls on main hand, off hand
- Sim rate: 2.85 PPM, measured; measured 2.87 PPM [2.62, 3.09], 16 wearers, 16 runs, 1258 procs on 17529 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights    | Class  | Slot | Hand split                   | Procs | Hits |  PPM | Left out  |
| ---------------- | --------- | ------ | ---- | ---------------------------- | ----: | ---: | ---: | --------- |
| 1jgTLvJQBMP6kxrm | 1-12      | Rogue  | oh   | timing + otsu for 31 swings  |    12 |  276 | 1.74 |           |
| 1w3vfpYKqbrBNZPQ | 1-2       | Rogue  | mh   | timing + otsu for 10 swings  |    76 | 1647 | 1.85 |           |
| 3PAwV6XFaB7hqDgz | 1-12      | Rogue  | oh   | timing + otsu for 3 swings   |    20 |  269 | 2.98 |           |
| 3THgCcxGJ4XYd6Kp | 3         | Rogue  | oh   | timing + otsu for 56 swings  |    43 |  634 | 2.71 |           |
| 4xnDCWwrGKAFdNHv | 1         | Rogue  | mh   | timing + otsu for 18 swings  |   101 | 1361 | 2.97 |           |
| 7dZrzYqjKHh28mvW | 1-2       | Rogue  | oh   | timing + otsu for 54 swings  |    70 |  911 | 3.07 |           |
| 987q1pjnYdVQbCmT | 1-6       | Rogue  | oh   | timing + otsu for 11 swings  |    84 | 1273 | 2.64 |           |
| 9QnWbJphKLFVgZ6z | 3-8       | Rogue  | oh   | timing + otsu for 4 swings   |    86 | 1220 | 2.82 |           |
| NfqRryF7gnMcX2ak | 1-2       | Rogue  | mh   | timing + otsu for 14 swings  |   118 | 1589 | 2.97 |           |
| QKg4z1amMYyBCp69 | 1-3, 5-30 | Rogue  | mh   | timing + otsu for 155 swings |   107 | 1332 | 3.21 |           |
| R9njq1AGDvmp36bM | 9         | Rogue  | oh   | timing                       |     5 |   22 | 9.09 |           |
| RW1BNk6Ybyh3XjLQ | 1-21      | Rogue  | mh   | timing + otsu for 15 swings  |   226 | 3109 | 2.91 |           |
| RvcrA3P9jznxhHVQ | 1-2, 4-7  | Hunter | mh   | timing + otsu for 5 swings   |    17 |  235 | 2.89 |           |
| TADvnpKkdwgZGJ6b | 1         | Rogue  | oh   | timing + otsu for 1 swings   |    93 | 1309 | 2.84 |           |
| g41xhtdFCXZ8Vr9z | 1-2       | Rogue  | oh   | timing + otsu for 14 swings  |    73 |  709 | 4.12 |           |
| pqfK2AgHZ7DnhvPF | 1         | Rogue  | mh   | timing + otsu for 9 swings   |   127 | 1633 | 3.11 |           |
| dRch7jNvCBJfgQt6 | 1         | Rogue  | oh   | timing + otsu for 50 swings  |    45 |  715 | 2.52 | duplicate |

## Cookie's Tenderizer (5197)

- Proc: weapon proc 1292008, logged as 1292008; rolls on main hand, off hand
- Sim rate: 1.9 PPM, observed; measured 1.91 PPM [1.58, 2.26], 7 wearers, 8 runs, 400 procs on 4495 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 1pDcTvkrCBhZJRgf | 1-14   | Paladin | mh   |                             |   105 | 1281 | 1.76 |          |
| 2PWL9DB1k47p8MRd | 1      | Warrior | mh   |                             |    63 |  464 | 2.91 |          |
| 2htpkPNX3yZ8K74C | 1-3    | Rogue   | oh   | timing + otsu for 39 swings |    18 |  282 | 1.37 |          |
| 3BH2d16wtD9F78QY | 1-18   | Paladin | mh   |                             |    84 |  898 | 2.00 |          |
| 3BYFMvdKPgLc4xDn | 4-5    | Paladin | mh   |                             |    65 |  660 | 2.11 |          |
| FWHahtQvqjL3wNdV | 1-3    | Rogue   | mh   | timing + otsu for 46 swings |    32 |  556 | 1.23 |          |
| QmjBxrzvah2XTVR1 | 7-8    | Warrior | mh   |                             |    33 |  354 | 2.00 |          |
| pwK87DzQgRHYCPNb | 1-4    | Priest  | mh   |                             |     0 |    0 |      |          |

## Cursed Felblade (14145)

- Proc: weapon proc 18381, logged as 18381; rolls on main hand
- Sim rate: 2.55 PPM, observed; measured 2.56 PPM [2.17, 3.02], 8 wearers, 8 runs, 666 procs on 6786 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 12VhLkPa98DYFH4g | 1-3         | Warrior | mh   |                             |    88 |  774 | 2.97 |          |
| 1RrDAhp7ZvVYc9P3 | 1-6         | Rogue   | mh   | timing + otsu for 18 swings |    39 |  560 | 1.82 |          |
| 1XFBL6hQPHcY4Zwr | 1-2         | Warrior | mh   |                             |    78 |  591 | 3.44 |          |
| 1cf7ba34WkwGyzQY | 1-12, 14-16 | Warrior | mh   |                             |    72 |  976 | 1.92 |          |
| 3CyQk69LzJjTn2X4 | 1-5         | Warrior | mh   |                             |   100 | 1066 | 2.45 |          |
| 3rDXa6gPxYNZJjFH | 2           | Rogue   | mh   | timing + otsu for 36 swings |   122 | 1482 | 2.15 |          |
| 9aCJ7GH6QRFjgqb4 | 1-11        | Warrior | mh   |                             |    50 |  471 | 2.77 |          |
| txhaZrbjXR7Kf2TD | 1, 3        | Warrior | mh   |                             |   117 |  866 | 3.52 |          |

## Darkwater Talwar (11121)

- Proc: weapon proc 16408, logged as 16408; rolls on main hand
- Sim rate: 3.3 PPM, observed; measured 3.32 PPM [2.97, 3.68], 17 wearers, 19 runs, 2148 procs on 17662 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split                          | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ----------------------------------- | ----: | ---: | ---: | -------- |
| 14ZTxcCRXhQLfYwP | 1-2, 5-6 | Paladin | mh   |                                     |    45 |  431 | 2.85 |          |
| 1CYTfP3pFDcaRrJm | 1-8      | Warrior | mh   | timing + speed ratio for 150 swings |   154 |  967 | 4.35 |          |
| 1vJgfm9dGnkDzwXb | 1-2      | Warrior | mh   |                                     |   174 | 1348 | 3.52 |          |
| 23dT9fk1DqcaJmwn | 1-7      | Paladin | mh   |                                     |   130 | 1270 | 2.79 |          |
| 24KTy3bLGNQm1rdx | 1-20     | Rogue   | mh   | timing + otsu for 52 swings         |    21 |  183 | 3.13 |          |
| 2GYNtHzfKmWpTPVZ | 1-14     | Warrior | mh   |                                     |   326 | 1929 | 4.61 |          |
| 2Z3YxGdfPhMK97kQ | 1-4      | Warrior | mh   |                                     |    88 |  745 | 3.22 |          |
| DK72jqp46VtnYHfZ | 1        | Rogue   | mh   | timing                              |    20 |  169 | 3.23 |          |
| DK72jqp46VtnYHfZ | 1        | Paladin | mh   |                                     |    15 |  178 | 2.30 |          |
| H3vdcCr6JX1gwYtq | 1-3      | Rogue   | mh   | timing + otsu for 2 swings          |    52 |  494 | 2.87 |          |
| LHRhtADNW8GfyYqT | 1        | Rogue   | mh   | timing + otsu for 16 swings         |    73 |  833 | 2.39 |          |
| NDf1mPJRCaTM8WkK | 2        | Warrior | mh   |                                     |    31 |  200 | 4.23 |          |
| QgR6kbHvY1CBhXFj | 1-2      | Rogue   | mh   | timing + otsu for 84 swings         |   231 | 2024 | 3.11 |          |
| Z1KYP3yWDVN7L6XH | 1-14     | Paladin | mh   |                                     |   202 | 1742 | 3.16 |          |
| khV3XMmJ4aRKzxBt | 1-3      | Rogue   | mh   | timing + otsu for 30 swings         |    37 |  327 | 3.09 |          |
| mRxb9L2wNkMJTWXY | 1-7      | Rogue   | mh   | timing + otsu for 4 swings          |   126 | 1261 | 2.73 |          |
| mRxb9L2wNkMJTWXY | 1-7      | Warrior | mh   |                                     |   245 | 1846 | 3.62 |          |
| vk2wGPDbQp8TZBR4 | 1        | Rogue   | mh   | timing + otsu for 96 swings         |    91 |  851 | 2.92 |          |
| yzaKX18ZAmpHvkjP | 1-16     | Paladin | mh   |                                     |    87 |  864 | 2.75 |          |
| 4NWMr2dJPyH9v8az | 1-2      | Hunter  | mh   | timing + otsu for 26 swings         |     0 |    0 |      | swap     |

## Darrowspike (13984)

- Proc: weapon proc 18276, logged as 18276; rolls on main hand
- Sim rate: 2.2 PPM, observed; measured 2.18 PPM [1.94, 2.44], 1 wearers, 1 runs, 282 procs on 5182 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: one wearer: no wearer interval

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 7fy93LdcwQhvF2gP | 1-7    | Warrior | mh   |            |   282 | 5182 | 2.18 |          |

## Dazzling Longsword (869)

- Proc: weapon proc 13752, logged as 13752; rolls on main hand
- Sim rate: 2.8 PPM, measured; measured 2.78 PPM [2.57, 2.98], 7 wearers, 7 runs, 887 procs on 7362 hits
- Reports from 2026-11-09 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 1FV3WyfXGCcPrT9b | 1, 3-8 | Hunter  | mh   | timing + otsu for 38 swings  |   139 | 1210 | 2.65 |          |
| 4C28YPXqgTMQJxB7 | 5-9    | Rogue   | mh   | timing + otsu for 1 swings   |    44 |  381 | 2.66 |          |
| 8X2Kc46faTnxVBLm | 1-3    | Rogue   | mh   | timing + otsu for 2 swings   |   110 | 1046 | 2.43 |          |
| BCZgcRxrh7dA4ayX | 1-4    | Rogue   | mh   | timing + otsu for 864 swings |   277 | 2072 | 3.09 |          |
| BgKt3Qm2fJDAWLzh | 1-23   | Rogue   | mh   | timing + otsu for 31 swings  |    76 |  692 | 2.53 |          |
| KVdX1Wn4GpzAH9MN | 11-12  | Rogue   | mh   | timing + otsu for 13 swings  |    80 |  580 | 3.18 |          |
| b12KD98fndF4hMq3 | 10-18  | Paladin | mh   |                              |   161 | 1381 | 2.69 |          |

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

## Destiny (647)

- Proc: weapon proc 17152, logged as 17152; rolls on main hand
- Sim rate: 1.2 PPM, measured; measured 1.18 PPM [1.07, 1.27], 7 wearers, 8 runs, 670 procs on 9440 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1ZyhaVDdFtTYWMcC | 10-12       | Warrior | mh   |            |    33 |  496 | 1.11 |          |
| Qb1yWxa8vHCGL7Yr | 12-15       | Paladin | mh   |            |   124 | 1754 | 1.18 |          |
| VLz78BfZyWa3Cxr2 | 1-9         | Warrior | mh   |            |   152 | 1930 | 1.31 |          |
| nX1qbatTzGZ3CAwB | 1, 3-7      | Warrior | mh   |            |   160 | 2236 | 1.19 |          |
| vZqVmLYQa6jwP3A9 | 2-16, 18-21 | Paladin | mh   |            |    51 |  734 | 1.16 |          |
| xbJBZj14NWCXrhAm | 5-12        | Paladin | mh   |            |    56 |  989 | 0.94 |          |
| zFXxNAn69HBGRaqg | 1-10        | Paladin | mh   |            |    53 |  709 | 1.25 |          |
| zFXxNAn69HBGRaqg | 1-10        | Paladin | mh   |            |    41 |  592 | 1.15 |          |

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

- Proc: weapon proc 1299869, logged as 1299869; rolls on main hand
- Sim rate: 1.65 PPM, observed; measured 1.64 PPM [1.50, 2.05], 6 wearers, 6 runs, 409 procs on 6782 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights          | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | --------------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 1PcrGTyatDAFx8Yk | 1-2             | Hunter  | mh   | timing + otsu for 160 swings |    16 |  210 | 2.08 |          |
| 8yPRmZdrq4QhcGn1 | 1-4             | Warrior | mh   |                              |    77 | 1276 | 1.65 |          |
| RJ82yLB7HcYrd39M | 2               | Warrior | mh   |                              |     1 |   71 | 0.38 |          |
| VYBjgJx8aZtRhmpy | 10-46           | Warrior | mh   |                              |    61 |  795 | 2.09 |          |
| mJDNgqYVGwb6A8tF | 1-17, 19, 21-22 | Hunter  | mh   | speed ratio                  |     1 |    7 | 3.94 |          |
| mdXkc6xwP8Y9pR47 | 1               | Warrior | mh   | timing + otsu for 137 swings |   253 | 4423 | 1.56 |          |

## Electrocutioner Leg (9446)

- Proc: weapon proc 1292851, logged as 1292851; rolls on main hand
- Sim rate: 4.25 PPM, measured; measured 4.25 PPM [4.02, 4.57], 15 wearers, 15 runs, 2336 procs on 19381 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights           | Class   | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------- | ------- | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 1qpQ6FW8fxD9gC4G | 1-7, 9-11, 13-26 | Paladin | mh   |                                    |   392 | 3554 | 3.89 |          |
| 27dNt8HLBCPpAmYZ | 1-31             | Rogue   | mh   | timing + otsu for 85 swings        |   164 | 1319 | 4.39 |          |
| 2jYyAK1RrPvTC8WM | 1-5              | Rogue   | mh   | timing + otsu for 5 swings         |    64 |  526 | 4.29 |          |
| 2ywhMnJq4Y3zaWtg | 1-9              | Rogue   | mh   | timing                             |    46 |  402 | 4.04 |          |
| 3B4wmh7JGkQq2Ndj | 1-2, 4-6         | Paladin | mh   |                                    |   272 | 2643 | 3.63 |          |
| 3HRvtBnAbfM9xTkK | 1-3              | Warrior | mh   |                                    |    75 |  570 | 4.64 |          |
| 6vKFBQNC1fMqR3Wa | 1-6              | Paladin | mh   |                                    |   190 | 1544 | 4.34 |          |
| AkBH7pF6XvZc3Gdw | 6-8              | Warrior | mh   |                                    |   100 |  628 | 5.62 |          |
| J2XL6AyZqgpjx8QK | 1-13             | Warrior | mh   |                                    |   205 | 1520 | 4.76 |          |
| NXgv93kr2G7zyTxm | 1, 3-9           | Rogue   | mh   | timing + otsu for 11 swings        |   293 | 2351 | 4.40 |          |
| NZn9hmTJHvcMRbfY | 1-2              | Paladin | mh   |                                    |   236 | 1856 | 4.49 |          |
| P1QTDgLwfytJKrBh | 1-8              | Rogue   | mh   | timing + otsu for 5 swings         |    88 |  745 | 4.17 |          |
| TGmKaXrQPp7czjx1 | 1-4              | Rogue   | mh   | timing + speed ratio for 6 swings  |   100 |  969 | 3.64 |          |
| aMNVKZpg8tRkhBP3 | 1-17             | Hunter  | mh   | timing + speed ratio for 35 swings |     6 |   68 | 3.11 |          |
| pbvkXazLqcHFhgRf | 4-7              | Warrior | mh   |                                    |   105 |  685 | 5.41 |          |

## Fang of the Crystal Spider (13218)

- Proc: weapon proc 17331, logged as 17331; rolls on main hand, off hand
- Sim rate: 0.9 PPM, observed; measured 0.91 PPM [0.72, 1.11], 4 wearers, 4 runs, 91 procs on 3749 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ----- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| QXFd74zDW8JZMmKx | 12-13  | Rogue | oh   | timing + otsu for 29 swings |    21 |  625 | 1.26 |          |
| XPYV9J6DfrMKxGqh | 5-6    | Rogue | oh   | timing + otsu for 93 swings |    25 | 1015 | 0.92 |          |
| kaH1G4ztmYpgMTJ6 | 1-2    | Rogue | oh   | timing + otsu for 34 swings |    10 |  631 | 0.59 |          |
| xXWkB3fqAjt9NJGH | 5      | Rogue | mh   | timing + otsu for 3 swings  |    35 | 1478 | 0.89 |          |

## Felstriker (12590)

- Proc: weapon proc 16551, logged as 16551; rolls on main hand, off hand
- Sim rate: 0.5 PPM, observed; measured 0.48 PPM [0.40, 0.58], 6 wearers, 7 runs, 114 procs on 8343 hits
- Reports from 2026-11-15 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split                          | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ----------------------------------- | ----: | ---: | ---: | -------- |
| 2wMznDRvmkpbtHxA | 5-7         | Rogue   | oh   | timing + speed ratio for 166 swings |     8 |  724 | 0.39 |          |
| YFtvD87kznr1CThg | 6-7         | Warrior | oh   | timing + otsu for 99 swings         |    23 | 1084 | 0.75 |          |
| cgMGCdPJQ84wzaBV | 12-13       | Warrior | mh   | timing + speed ratio for 22 swings  |    12 | 1340 | 0.32 |          |
| jXTVfy3ApWKPYwDG | 1-22        | Warrior | mh   |                                     |    39 | 2532 | 0.54 |          |
| nX1qbatTzGZ3CAwB | 4-7         | Rogue   | mh   | timing + otsu for 6 swings          |    13 | 1268 | 0.36 |          |
| vZqVmLYQa6jwP3A9 | 5-16, 18-21 | Rogue   | oh   | timing                              |     5 |  377 | 0.47 |          |
| vZqVmLYQa6jwP3A9 | 5-16, 18-21 | Warrior | mh   |                                     |    14 | 1018 | 0.49 |          |

## Fiery War Axe (870)

- Proc: weapon proc 18796, logged as 18796; rolls on main hand
- Sim rate: 0.95 PPM, observed; measured 0.94 PPM [0.83, 1.12], 6 wearers, 6 runs, 283 procs on 5009 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 2b3Tf7QGKHA1WMxD | 1-3         | Paladin | mh   |            |    55 |  747 | 1.23 |          |
| 4C28YPXqgTMQJxB7 | 1, 3-9      | Shaman  | mh   |            |    55 | 1096 | 0.84 |          |
| 4LrMDtG6T8AVRJzQ | 3-11, 14-28 | Shaman  | mh   |            |    46 |  786 | 0.98 |          |
| 7n6rKTfk1vhzgXyF | 1-2         | Shaman  | mh   |            |    45 | 1032 | 0.73 |          |
| C3cK4LTfvbhAQxyj | 1-5         | Paladin | mh   |            |    66 | 1180 | 0.93 |          |
| C8Hk1q3rpN27YWam | 1-3         | Paladin | mh   |            |    16 |  168 | 1.59 |          |

## Firebelcher (5243)

- Proc: equip proc 1291686, logged as 1291662; rolls on ranged auto attacks
- Sim rate: 3.85 PPM, observed; measured 3.85 PPM [3.46, 4.28], 14 wearers, 14 runs, 306 procs on 2808 hits
- Reports from 2026-11-07 to 2026-11-16

| Report           | Fights | Class   | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ------ | ---------- | ----: | ---: | ---: | -------- |
| 1RrDAhp7ZvVYc9P3 | 1-6    | Warlock | ranged |            |    15 |  109 | 4.86 |          |
| 1ZJVMgcYWTyRQh2w | 1-3    | Mage    | ranged |            |    18 |  110 | 5.78 |          |
| 1xDPwjtGBTzZWm7Y | 1-4    | Priest  | ranged |            |    35 |  309 | 4.00 |          |
| 287trgyqQLvjACK9 | 1-4    | Priest  | ranged |            |    10 |   94 | 3.75 |          |
| 2mW9A7wQ4bhv1YMF | 1-3    | Priest  | ranged |            |    50 |  401 | 4.40 |          |
| 3PnWT1jGz4tagB6r | 1      | Priest  | ranged |            |    27 |  315 | 3.03 |          |
| 4NWMr2dJPyH9v8az | 1-2    | Priest  | ranged |            |    20 |  193 | 3.66 |          |
| 6WzAhcVgfLCt78nX | 1-8    | Mage    | ranged |            |     1 |   14 | 2.52 |          |
| 6wxJrRWgPyQb21hm | 1-7    | Priest  | ranged |            |    33 |  283 | 4.12 |          |
| LbmB2QdWnPT7JGAF | 1-6    | Priest  | ranged |            |     4 |   31 | 4.55 |          |
| QHqMJPfgbBCvZan7 | 23-28  | Priest  | ranged |            |     2 |   16 | 4.41 |          |
| Z9qbhQrJg2Wfz1FM | 1-2    | Warlock | ranged |            |     5 |   36 | 4.90 |          |
| d6bwAkLKjGq7pVMy | 1      | Warlock | ranged |            |    23 |  262 | 3.10 |          |
| kBzVYb3Cy7QwnKMD | 1-6    | Priest  | ranged |            |    63 |  635 | 3.50 |          |

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

- Proc: weapon proc 18084, logged as 18084; rolls on main hand, off hand
- Sim rate: 3.5 PPM, observed; measured 3.48 PPM [2.96, 4.20], 7 wearers, 7 runs, 830 procs on 7527 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights           | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 498vhQAtCw2XgZJM | 11, 13-15        | Warrior | mh   |                              |   109 |  922 | 3.73 |          |
| 9MCqcwjzLk4JBARG | 1-5, 7-14, 16-22 | Warrior | mh   |                              |   162 | 1619 | 3.16 |          |
| 9WGcCxKHQ4pkX6Rj | 1-11             | Warrior | oh   | timing + otsu for 19 swings  |    54 |  317 | 5.38 |          |
| 9hRCG4AnXxvZB3mL | 1-20             | Paladin | mh   |                              |   171 | 1987 | 2.72 |          |
| JMnDtLGxf4FqCV6w | 1-3              | Warrior | oh   | timing + otsu for 122 swings |   188 | 1330 | 4.46 |          |
| KYmwz2q1tG8yAcdP | 1-6              | Warrior | mh   |                              |    93 |  679 | 4.33 |          |
| zn64f9D8TVt3wFZq | 1-3              | Paladin | mh   |                              |    53 |  673 | 2.49 |          |

## Flame Wrath (11809)

- Proc: weapon proc 16559, logged as 16559, 16560; rolls on main hand
- Sim rate: 5.3 PPM, observed; measured 5.30 PPM [4.71, 5.93], 1 wearers, 1 runs, 202 procs on 693 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 7 of 8 runs are feral druids, left out (paw speed, not the weapon); one wearer: no wearer interval

| Report           | Fights           | Class  | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------- | ------ | ---- | ---------- | ----: | ---: | ---: | -------- |
| 6yHgkT8LqP42NmxY | 1-17             | Hunter | mh   |            |   202 |  693 | 5.30 |          |
| 1X37tZRzgkmd9rNP | 1                | Druid  | mh   |            |   493 | 2710 | 3.31 | feral    |
| 1nTQBK4hmcaLbFZV | 2                | Druid  | mh   |            |   746 | 3600 | 3.77 | feral    |
| 3R69PBjxZvcfbnX8 | 2-6              | Druid  | mh   |            |    49 |  433 | 2.06 | feral    |
| 3qfDVTZ1J29zPChn | 1-23             | Druid  | mh   |            |   775 | 4023 | 3.50 | feral    |
| 4f2ygdYCzFAvqc6K | 10, 12-14, 16-18 | Druid  | mh   |            |   151 | 1699 | 1.62 | feral    |
| 7ChG2VXaALzcw9pW | 1-2              | Druid  | mh   |            |   385 | 1254 | 5.58 | feral    |
| k1gAvx7zVaPpLhYC | 1                | Druid  | mh   |            |   112 | 2232 | 0.91 | feral    |

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

- Proc: weapon proc 18075, logged as 18075; rolls on main hand, off hand
- Sim rate: 2.4 PPM, observed; measured 2.39 PPM [2.14, 2.79], 8 wearers, 8 runs, 363 procs on 4555 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 6VJxcqnYa4wdBWQv | 1-9    | Rogue   | oh   | timing + speed ratio for 11 swings |    32 |  414 | 2.32 |          |
| 7tCZB31vV2Jp6XWd | 1-2    | Rogue   | oh   | timing + otsu for 108 swings       |    43 |  617 | 2.09 |          |
| 8vDJ9jmHd4FA7Twt | 1-2    | Rogue   | oh   | timing + otsu for 12 swings        |    45 |  616 | 2.19 |          |
| CnNwpaYW83cJr6fR | 1-6    | Warrior | oh   | timing + otsu for 13 swings        |    60 |  479 | 3.76 |          |
| L7JpRfWtAZ6T1hGx | 3, 5   | Rogue   | oh   | timing                             |    16 |  274 | 1.75 |          |
| P8Znjv1fgBD2CAVY | 2-12   | Warrior | mh   |                                    |    76 | 1009 | 2.26 |          |
| PDT2yACNzd4LMZ98 | 1-14   | Rogue   | oh   | timing + otsu for 46 swings        |    42 |  584 | 2.16 |          |
| qnyc1zwaB4FrvM9x | 1      | Warrior | oh   | timing + otsu for 16 swings        |    49 |  562 | 2.61 |          |

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

- Proc: weapon proc 1292158, logged as 1292158; rolls on main hand, off hand
- Sim rate: 2.15 PPM, observed; measured 2.15 PPM [1.64, 2.45], 6 wearers, 6 runs, 292 procs on 4078 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights                 | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------- | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 6rgDRnVFX9a3pbGy | 1-4                    | Warrior | oh   | timing                      |    27 |  466 | 1.74 |          |
| F1MPmrQNq4v6XJ23 | 1-2                    | Paladin | mh   |                             |    26 |  543 | 1.44 |          |
| HJxqkQPAnmKtwNjd | 1-16, 18, 20-24, 26-31 | Hunter  | oh   | timing + otsu for 29 swings |    16 |  194 | 2.48 |          |
| JgL3qFZA4aQcpXPC | 1, 3-9, 11-13          | Warrior | mh   |                             |   172 | 2041 | 2.53 |          |
| KjLXqdCZANvQc9BT | 1-3                    | Rogue   | oh   | timing + otsu for 53 swings |    44 |  754 | 1.75 |          |
| Lcy3V1wHPpqZQAG7 | 2, 4-12                | Rogue   | oh   | timing + otsu for 97 swings |     7 |   80 | 2.61 |          |

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

- Proc: weapon proc 1317432, logged as 1317432; rolls on main hand
- Sim rate: 2 PPM, observed; measured 1.99 PPM [1.73, 2.23], 6 wearers, 6 runs, 568 procs on 8157 hits
- Reports from 2026-11-13 to 2026-11-15

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| BQnXb3hTC4rmJ2pd | 1-2    | Warrior | mh   |                             |    82 | 1323 | 1.77 |          |
| J97GvmrRXCFZfQ8a | 1-2    | Warrior | mh   |                             |   112 | 1333 | 2.40 |          |
| LaJzMkFfhmpc2DbH | 7-10   | Paladin | mh   |                             |    95 | 1204 | 2.25 |          |
| YWZm7NKhFvqbMfgy | 1-6    | Paladin | mh   |                             |   117 | 1618 | 2.07 |          |
| dGXqkVhwT17nHfBc | 1-2    | Warrior | mh   | timing + otsu for 13 swings |    79 | 1496 | 1.51 |          |
| jQPc3Xy1hW87gfdD | 1-2    | Paladin | mh   |                             |    83 | 1183 | 2.00 |          |

## Hammer of the Northern Wind (810)

- Proc: weapon proc 13439, logged as 13439; rolls on main hand
- Sim rate: 4.55 PPM, observed; measured 4.57 PPM [4.00, 4.87], 4 wearers, 5 runs, 1619 procs on 10120 hits
- Reports from 2026-11-14 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 6kf3tMF4wCPNcRnh | 1-33   | Paladin | mh   |                             |   588 | 3481 | 4.83 |          |
| J38xbrFytgM9p1LT | 7-19   | Warrior | mh   | timing + otsu for 65 swings |   122 | 1094 | 3.19 |          |
| P4bXMwRjD1zyhZdJ | 1-2    | Warrior | mh   |                             |   598 | 3702 | 4.62 |          |
| TF6dc3NCHbKrfxJt | 1-6    | Warrior | mh   |                             |   163 | 1010 | 4.61 |          |
| dBNfgVDLMmnc8h4v | 10     | Warrior | mh   |                             |   148 |  833 | 5.08 |          |

## Hanzo Sword (8190)

- Proc: weapon proc 16405, logged as 16405; rolls on main hand, off hand
- Sim rate: 1.2 PPM, observed; measured 1.20 PPM [1.05, 1.36], 7 wearers, 7 runs, 222 procs on 7425 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights                | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out  |
| ---------------- | --------------------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | --------- |
| 4f2ygdYCzFAvqc6K | 6-8, 10, 12-14, 16-18 | Rogue   | oh   | timing + otsu for 121 swings |    26 |  872 | 1.19 |           |
| 6xHzpKd8wCna37Lb | 1-14                  | Warrior | mh   |                              |    41 | 1112 | 1.47 |           |
| 824kNhdFaALMBnvR | 1-4, 7-18, 20-45      | Warrior | oh   | timing + otsu for 59 swings  |    17 |  502 | 1.36 |           |
| DNc6QALhxmGbCH3Y | 1-6                   | Rogue   | oh   | timing + otsu for 43 swings  |    42 | 1602 | 1.05 |           |
| KPqgbZxhn6BXwCvJ | 1                     | Rogue   | oh   | timing + otsu for 143 swings |    40 | 1291 | 1.24 |           |
| MK6XJDzfLpFjnRWc | 1-5                   | Rogue   | oh   | timing + otsu for 328 swings |    36 | 1119 | 1.29 |           |
| qNyQvka41YhVmdHR | 4                     | Rogue   | oh   | timing + otsu for 13 swings  |    20 |  927 | 0.86 |           |
| 2DC7xQNm9wHJnVF4 | 6-7                   | Rogue   | oh   | timing + otsu for 121 swings |    26 |  872 | 1.19 | duplicate |

## Iceblade Hacker (13952)

- Proc: weapon proc 1298416, logged as 1298416; rolls on main hand
- Sim rate: 1.9 PPM, measured; measured 1.92 PPM [1.75, 2.02], 5 wearers, 7 runs, 1299 procs on 20342 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights                                                                                                                                  | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out  |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | --------- |
| 4h1PjwkaKX9q8mgT | 1-2                                                                                                                                     | Paladin | mh   |                              |   161 | 2424 | 1.99 |           |
| A1yxMTmKatkf3HRJ | 15-17                                                                                                                                   | Rogue   | mh   | timing + otsu for 826 swings |    10 |  342 | 0.88 |           |
| CWhG6LyXk3PBHdMQ | 1-2                                                                                                                                     | Paladin | mh   |                              |   198 | 3526 | 1.68 |           |
| DNc6QALhxmGbCH3Y | 1-6                                                                                                                                     | Rogue   | mh   | timing + otsu for 20 swings  |    34 |  572 | 1.78 |           |
| HxCQat38vmjnkpb6 | 1-7, 9-20, 22-33, 35-48, 50-67, 69-86, 88-157, 159-185, 187-212, 214-241, 243-458, 460-498, 500-587, 589-639, 641-692, 694-802, 804-830 | Paladin | mh   |                              |   525 | 7982 | 1.97 |           |
| KYmFPtyJ3gkzHfVM | 102-160                                                                                                                                 | Rogue   | mh   | timing + otsu for 213 swings |   188 | 2574 | 2.19 |           |
| VLz78BfZyWa3Cxr2 | 1-7                                                                                                                                     | Paladin | mh   |                              |   183 | 2922 | 1.88 |           |
| JrqDZX4LT6NpxzM7 | 1-6                                                                                                                                     | Paladin | mh   |                              |   498 | 7368 | 2.03 | duplicate |

## Iron Knuckles (2942)

- Proc: weapon proc 13491, logged as 13491; rolls on main hand, off hand
- Sim rate: 2.4 PPM, observed; measured 2.39 PPM [2.07, 2.66], 5 wearers, 6 runs, 327 procs on 4821 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                 | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | -------------------------- | ----: | ---: | ---: | -------- |
| 38D6JN4nZzjcGtHg | 1-4    | Warrior | mh   |                            |    64 |  912 | 2.48 |          |
| 4QAtNfTXMyDGPqCr | 1-13   | Rogue   | mh   | timing                     |    21 |  335 | 2.21 |          |
| AzbN8GY9DnBrH4aW | 19-31  | Warrior | mh   |                            |    46 |  788 | 2.06 |          |
| DwWM9H3rpTnGPfbt | 1-14   | Warrior | mh   |                            |    80 |  989 | 2.85 |          |
| GRTmQn3VDpwZKbCA | 1-6, 8 | Warrior | mh   |                            |   100 | 1508 | 2.34 |          |
| H3vdcCr6JX1gwYtq | 1-3    | Rogue   | oh   | timing + otsu for 2 swings |    16 |  289 | 1.96 |          |

## Ironforge Greathammer (279895)

- Proc: weapon proc 1322295, logged as 1322295; rolls on main hand
- Sim rate: 3.3 PPM, measured; measured 3.29 PPM [3.03, 3.63], 18 wearers, 18 runs, 1549 procs on 9109 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: 11 of 29 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 17jHXDAWBRQdgqV6 | 1-4    | Shaman  | mh   |            |    41 |  296 | 2.68 |          |
| 1HK7gFbwy8aJPAh6 | 1-3    | Paladin | mh   |            |   157 | 1025 | 2.96 |          |
| 1HK7gFbwy8aJPAh6 | 1-3    | Warrior | mh   |            |    97 |  595 | 3.16 |          |
| 1fpaDyG3mFnB9KW8 | 1      | Shaman  | mh   |            |     7 |   71 | 1.91 |          |
| 1h9P4rYzWDvtHRxb | 1      | Warrior | mh   |            |   130 |  656 | 3.84 |          |
| 1pDcTvkrCBhZJRgf | 1-14   | Shaman  | mh   |            |   106 |  673 | 3.05 |          |
| 7kYG3K21FmyMVABD | 1-4    | Shaman  | mh   |            |    11 |  106 | 2.01 |          |
| CdH9vgqPbRX1JKM3 | 1      | Warrior | mh   |            |   132 |  519 | 4.92 |          |
| HKQW3wCmr4kJXz2j | 1-3    | Shaman  | mh   |            |    28 |  167 | 3.25 |          |
| L8yaqgDXG26Nf9v1 | 1-2    | Paladin | mh   |            |    56 |  386 | 2.81 |          |
| LVwmRx1Mqak7vp4y | 1-8    | Warrior | mh   |            |    50 |  308 | 3.14 |          |
| QmCHxRYD9tbjLfW1 | 1-4    | Warrior | mh   |            |    78 |  332 | 4.55 |          |
| QmCHxRYD9tbjLfW1 | 1-4    | Shaman  | mh   |            |    46 |  326 | 2.73 |          |
| RQxLGd1agBDt6hXP | 1-4    | Paladin | mh   |            |   216 | 1172 | 3.57 |          |
| ZADc3bTCKhn9jJdz | 1-3    | Shaman  | mh   |            |    39 |  223 | 3.38 |          |
| gXLj6knBf4QFvGD7 | 3-5    | Paladin | mh   |            |   126 |  714 | 3.42 |          |
| jr1xaAPwb4f2HpvK | 1-3    | Paladin | mh   |            |   156 | 1095 | 2.76 |          |
| q2h6pRvdBCKMabnL | 1-2    | Warrior | mh   |            |    73 |  445 | 3.18 |          |
| 1K6c48bWMN9nzCZA | 1-3    | Druid   | mh   |            |    28 |  330 | 1.64 | feral    |
| 1K6c48bWMN9nzCZA | 1-3    | Druid   | mh   |            |    38 | 1104 | 0.67 | feral    |
| 7kYG3K21FmyMVABD | 1-4    | Druid   | mh   |            |    48 | 1172 | 0.79 | feral    |
| 7kYG3K21FmyMVABD | 1-4    | Druid   | mh   |            |     8 | 1448 | 0.11 | feral    |
| JGFcm9zq48CArhgk | 1-6    | Druid   | mh   |            |   141 | 3753 | 0.73 | feral    |
| LKzDgyYcfakX38xP | 1      | Druid   | mh   |            |    29 |  997 | 0.56 | feral    |
| LKzDgyYcfakX38xP | 1      | Druid   | mh   |            |     4 |  133 | 0.58 | feral    |
| LVwmRx1Mqak7vp4y | 8      | Druid   | mh   |            |    20 |  343 | 1.13 | feral    |
| T7WjV2vKaydLkMzX | 1-5    | Druid   | mh   |            |    16 |  682 | 0.45 | feral    |
| dJb4v1ZQNgtak9nB | 1-4    | Druid   | mh   |            |    56 | 1548 | 0.70 | feral    |
| y2dDThApkr8MVRcj | 1-3, 5 | Druid   | mh   |            |    14 |  109 | 2.49 | feral    |

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

- Proc: weapon proc 8348, logged as 8348; rolls on off hand
- Sim rate: 0.85 PPM, observed; measured 0.86 PPM [0.66, 1.11], 3 wearers, 3 runs, 54 procs on 2915 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ----- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 6kN78G2xfqy3vwTP | 6      | Rogue | oh   | timing + otsu for 4 swings   |     1 |   58 | 0.80 |          |
| JMnDtLGxf4FqCV6w | 1-3    | Rogue | oh   | timing + otsu for 20 swings  |    42 | 2252 | 0.86 |          |
| dJa2G7PHZg9wt6QX | 1-4    | Rogue | oh   | timing + otsu for 134 swings |    11 |  605 | 0.84 |          |

## Kang the Decapitator (2291)

- Proc: weapon proc 17153, logged as 17153; rolls on main hand
- Sim rate: 1 PPM, observed; measured 1.00 PPM [0.91, 1.12], 8 wearers, 9 runs, 587 procs on 9771 hits
- Reports from 2026-11-11 to 2026-11-16

| Report           | Fights             | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 2FkYRHB1dWhra8vt | 1-17               | Shaman  | mh   |            |    35 |  785 | 0.74 |          |
| 6kf3tMF4wCPNcRnh | 1-33               | Shaman  | mh   |            |    78 | 1255 | 1.04 |          |
| 9Z473MfXLYqcyWPg | 1-2                | Shaman  | mh   |            |    28 |  646 | 0.72 |          |
| FHqYCNcKntvDAjxg | 4-20, 22-31, 33-36 | Paladin | mh   |            |    50 |  577 | 1.44 |          |
| FMaAjHcwCQLkpV32 | 1-4                | Shaman  | mh   |            |    83 | 1323 | 1.05 |          |
| RfT7K4v2jhY9FAzD | 1-3                | Shaman  | mh   |            |    60 | 1169 | 0.86 |          |
| RpFQn1ftYCAgMW47 | 1-17               | Shaman  | mh   |            |    88 | 1590 | 0.92 |          |
| W78tmfknBKz46hRZ | 1-2                | Shaman  | mh   |            |    85 | 1386 | 1.02 |          |
| ZNyPLR2vkjMGptCJ | 1-2                | Shaman  | mh   |            |    80 | 1040 | 1.28 |          |

## Killmaim (13016)

- Proc: weapon proc 13318, logged as 13318; rolls on main hand
- Sim rate: 0.95 PPM, observed; measured 0.97 PPM [0.82, 1.14], 7 wearers, 8 runs, 139 procs on 2692 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 0.66 -> 0.97 (73 active minutes, 32% of hits, in 3 runs)

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out  |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | --------- |
| 29A3dKwPgpRNnWvc | 7-11   | Shaman  | mh   |            |     6 |  125 | 0.90 |           |
| 2aVnTxFzfckdXpBy | 1-7    | Warrior | mh   |            |    17 |  351 | 0.91 |           |
| 3xgPkwfn8yQ9BqAW | 1-4    | Warrior | mh   |            |    29 |  449 | 1.21 |           |
| 6v4wTWQkjb3Xd7mM | 6      | Warrior | mh   |            |     4 |   49 | 1.53 |           |
| P16dx3VLmRyaDJZQ | 1-18   | Shaman  | mh   |            |    31 |  561 | 1.04 |           |
| P7cV2HjLgxGpmZQk | 1-7    | Shaman  | mh   |            |    14 |  313 | 0.84 |           |
| d7DFtXL2jfNa8Bw4 | 1-5    | Shaman  | mh   |            |    28 |  529 | 0.99 |           |
| zJ6C2148dTYfaXHb | 2-4    | Shaman  | mh   |            |    10 |  315 | 0.60 |           |
| 9mN62kzgMdKypTqf | 1-6    | Shaman  | mh   |            |    28 |  729 | 0.72 | duplicate |

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

- Proc: weapon proc 1293482, logged as 1293482; rolls on main hand
- Sim rate: 2.55 PPM, observed; measured 2.53 PPM [1.79, 2.84], 14 wearers, 14 runs, 252 procs on 1928 hits
- Reports from 2026-11-10 to 2026-11-16
- Notes: 8 of 22 runs are feral druids, left out (paw speed, not the weapon)

| Report           | Fights                                                                                                                             | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------------------------------------------------------------------------------------------------------------------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1aBhkzdDMZKGtb37 | 162-170, 172-176, 178-187, 189-195, 197-237, 239-286, 288-289, 291-295, 297-312, 314-323, 325-328, 331, 333-335, 337-371, 374, 376 | Hunter  | mh   |            |     1 |    2 | 9.68 |          |
| 1xqBcj4CZQXV38Pv | 1-2                                                                                                                                | Warrior | mh   |            |    89 |  644 | 2.67 |          |
| 214ndcFyjpPM8wah | 1-3, 6-9, 11-15, 19-28, 30                                                                                                         | Warrior | mh   |            |   125 |  936 | 2.58 |          |
| 27jZGhfHgp1LR8MB | 1-11                                                                                                                               | Hunter  | mh   |            |     3 |   46 | 1.26 |          |
| 2PpAjf1qRmTv6gDz | 1-17                                                                                                                               | Hunter  | mh   |            |     2 |   27 | 1.43 |          |
| 2WH4jLt3FADy8fnh | 1-2, 5-6                                                                                                                           | Hunter  | mh   |            |     1 |   22 | 0.88 |          |
| CpDKr6AWJ4fw7xhg | 5-7                                                                                                                                | Hunter  | mh   |            |     5 |   51 | 1.90 |          |
| GBDAWjKNh6Q9XnF7 | 1-4                                                                                                                                | Hunter  | mh   |            |     1 |   20 | 0.97 |          |
| PmR8Lpvazd4q61nA | 6-12                                                                                                                               | Hunter  | mh   |            |     2 |   11 | 3.52 |          |
| Vb3ZxfXtyDWr1ndP | 1                                                                                                                                  | Hunter  | mh   |            |     0 |    2 | 0.00 |          |
| WYrCyJkdfMXw84Ph | 13-21, 23-35, 37-39, 41-48                                                                                                         | Hunter  | mh   |            |     6 |   40 | 2.90 |          |
| bCzt1Fq24mYcPXkB | 1-2                                                                                                                                | Hunter  | mh   |            |     3 |   60 | 0.97 |          |
| pqfK2AgHZ7DnhvPF | 1                                                                                                                                  | Hunter  | mh   |            |    11 |   42 | 5.07 |          |
| rD8m4JW1bVyYPKGX | 1-3, 6                                                                                                                             | Hunter  | mh   |            |     3 |   25 | 2.32 |          |
| 127tXgbzdMBZYKp9 | 1-17, 19-20                                                                                                                        | Druid   | mh   |            |    25 |  407 | 1.19 | feral    |
| 1FV3WyfXGCcPrT9b | 1, 3-8                                                                                                                             | Druid   | mh   |            |    87 | 3858 | 0.44 | feral    |
| 3zxaGvgWKLFArhjk | 4-8                                                                                                                                | Druid   | mh   |            |    36 | 1810 | 0.38 | feral    |
| 6xd1aP4McnjKyqTG | 1-2, 4-10                                                                                                                          | Druid   | mh   |            |    10 |  824 | 0.23 | feral    |
| 7CmgWKkj3Hx1qd4c | 2, 4-7, 9-10                                                                                                                       | Druid   | mh   |            |    40 | 2078 | 0.37 | feral    |
| 8CzwkcFWrA91QaxG | 1-12                                                                                                                               | Druid   | mh   |            |    33 | 2361 | 0.27 | feral    |
| PFGvpDf9WZgcY6Rk | 1-5                                                                                                                                | Druid   | mh   |            |    50 | 2793 | 0.35 | feral    |
| Z9DMkXc2KqNAxVtr | 1-12                                                                                                                               | Druid   | mh   |            |    37 | 1552 | 0.46 | feral    |

## Lord General's Sword (11817)

- Proc: weapon proc 15602, logged as 15602; rolls on main hand
- Sim rate: 2.1 PPM, measured; measured 2.09 PPM [1.92, 2.20], 7 wearers, 9 runs, 1267 procs on 14018 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights                       | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 24kgnZAXbKTW89x6 | 1-2                          | Warrior | mh   |                              |    14 |   93 | 3.47 |          |
| ARFXzbqHDpBLJdhy | 1, 3-4                       | Rogue   | mh   | timing + otsu for 40 swings  |   185 | 2044 | 2.09 |          |
| CDKVcP6ghZJtv8AW | 3-6, 8-16                    | Rogue   | mh   | timing + otsu for 49 swings  |    91 | 1022 | 2.06 |          |
| MwVH1KQmak73NPd8 | 8-16                         | Warrior | mh   | timing + otsu for 17 swings  |   100 | 1346 | 1.72 |          |
| NXMZV9xjvWpB7GtP | 10, 12-107, 109-116, 118-122 | Warrior | mh   |                              |    97 | 1305 | 1.72 |          |
| NcxbzW4kJYQ7DGhj | 1-3                          | Rogue   | mh   | timing + otsu for 51 swings  |    99 | 1034 | 2.21 |          |
| RVkGfcrbZ13MWqQ2 | 1-3                          | Rogue   | mh   | timing + otsu for 69 swings  |   396 | 4237 | 2.16 |          |
| X8z9dgD6HRTZ24Vc | 23                           | Warrior | mh   |                              |    33 |  252 | 3.02 |          |
| pgN9QwMLYVfrqGbv | 1-4                          | Rogue   | mh   | timing + otsu for 315 swings |   252 | 2686 | 2.17 |          |

## Mana Channeling Wand (18483)

- Proc: equip proc 1302348, logged as 1302350; rolls on ranged auto attacks
- Sim rate: 4.25 PPM, observed; measured 4.26 PPM [3.00, 4.92], 5 wearers, 5 runs, 160 procs on 1408 hits
- Reports from 2026-11-11 to 2026-11-15

| Report           | Fights    | Class  | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------ | ------ | ---------- | ----: | ---: | ---: | -------- |
| QXFd74zDW8JZMmKx | 1-2, 4-13 | Mage   | ranged |            |     6 |   48 | 4.69 |          |
| Wv4QYA6C7antxKXr | 1         | Priest | ranged |            |     2 |   22 | 3.41 |          |
| fjtT3JVxbGPz6RwX | 6         | Priest | ranged |            |     4 |   26 | 5.77 |          |
| gBDMLv8J1K7nqkpd | 112-157   | Priest | ranged |            |   121 |  960 | 4.73 |          |
| wkJXWxD23VMqThLg | 14-17     | Priest | ranged |            |    27 |  352 | 2.88 |          |

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

- Proc: weapon proc 18211, logged as 18211; rolls on main hand
- Sim rate: 2 PPM, measured; measured 2.00 PPM [1.91, 2.10], 8 wearers, 10 runs, 1530 procs on 13083 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights    | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1NBKq2mDWghwR69z | 4, 6      | Paladin | mh   |            |   110 | 1101 | 1.71 |          |
| 4PtKGL1zcRV8Chfn | 1-3       | Paladin | mh   |            |   156 | 1171 | 2.28 |          |
| 7BQKZXWzY9MaxAPq | 1-4       | Paladin | mh   |            |    39 |  375 | 1.78 |          |
| 7c21XafhWTAdqNGY | 1-10      | Paladin | mh   |            |   213 | 1691 | 2.16 |          |
| 8cVPaCpAMmv93bkD | 1-9       | Paladin | mh   |            |   121 | 1039 | 2.00 |          |
| LPQd7pfbjDkYZyRz | 1-2, 4-5  | Paladin | mh   |            |   106 |  959 | 1.89 |          |
| Lg82ZFDyQ4MvRm9p | 5-11, 13  | Paladin | mh   |            |   140 | 1317 | 1.82 |          |
| P1LGvYV9FdmR8x3n | 1         | Paladin | mh   |            |   111 |  864 | 2.20 |          |
| aANyFDgb69hxqY2w | 3         | Paladin | mh   |            |   122 | 1010 | 2.07 |          |
| ynVw7ZgMx4tjF3fQ | 1-6, 8-18 | Paladin | mh   |            |   412 | 3556 | 1.99 |          |

## Pendulum of Doom (9425)

- Proc: weapon proc 10373, logged as 10373; rolls on main hand
- Sim rate: 1.05 PPM, measured; measured 1.05 PPM [0.98, 1.15], 6 wearers, 7 runs, 645 procs on 9187 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 9pMcvgxrQNfLZ7hd | 1-15        | Warrior | mh   |            |   156 | 1755 | 1.33 |          |
| CR6khYVnQAp3HaTy | 1-2, 4-6    | Paladin | mh   |            |    83 | 1169 | 1.07 |          |
| Pjz8kK2XvxyQrTCB | 4-16        | Shaman  | mh   |            |    80 | 1249 | 0.96 |          |
| WygY76wT4zqbvCM2 | 1-11, 13-16 | Paladin | mh   |            |   119 | 1700 | 1.05 |          |
| X186Gx2gRBZChbVA | 1-7         | Shaman  | mh   |            |    58 |  919 | 0.95 |          |
| ga9Qk6DcjJmNTGVH | 2-3         | Warrior | mh   |            |    15 |  322 | 0.70 |          |
| qw3ZzkBmQYW7pXVH | 1-19        | Warrior | mh   |            |   134 | 2073 | 0.97 |          |

## Phantom Blade (7961)

- Proc: weapon proc 9806, logged as 9806; rolls on main hand, off hand
- Sim rate: 2.25 PPM, observed; measured 2.27 PPM [1.90, 2.71], 4 wearers, 4 runs, 218 procs on 3606 hits
- Reports from 2026-11-15 to 2026-11-15
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights  | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 2YfdbFMgGhr7zHwQ | 3-4     | Rogue   | oh   | timing + otsu for 155 swings |    47 |  878 | 2.01 |          |
| 6kN78G2xfqy3vwTP | 1-6     | Warrior | mh   |                              |    69 | 1022 | 2.53 |          |
| XFdcWaPL8MthYH6A | 1-5     | Rogue   | oh   | timing + otsu for 100 swings |    49 |  613 | 3.00 |          |
| dt4M3ypRnTkwaVQY | 1, 3-10 | Rogue   | oh   | timing + otsu for 28 swings  |    53 | 1093 | 1.82 |          |

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

- Proc: weapon proc 1320925, logged as 1320925; rolls on main hand, off hand
- Sim rate: 2.05 PPM, observed; measured 2.05 PPM [1.87, 2.59], 7 wearers, 7 runs, 395 procs on 4814 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights       | Class   | Slot | Hand split                          | Procs | Hits |  PPM | Left out |
| ---------------- | ------------ | ------- | ---- | ----------------------------------- | ----: | ---: | ---: | -------- |
| 4NMcqLwQ2rB9amCH | 1-3          | Rogue   | mh   | timing + otsu for 4 swings          |    35 |  336 | 2.60 |          |
| 9hywC8jMTLV36Pxd | 4-8          | Shaman  | mh   |                                     |    56 |  805 | 1.74 |          |
| 9rPx2JNYLbCy1waX | 1-9, 11-15   | Rogue   | oh   | timing + otsu for 2 swings          |    13 |  154 | 2.12 |          |
| Brj9PhXHQM3kLN2f | 1-4          | Rogue   | oh   | timing + speed ratio for 163 swings |    50 |  459 | 2.73 |          |
| HqwWnYGKdhpCgZVR | 1-12         | Rogue   | mh   | timing + otsu for 29 swings         |   171 | 2313 | 1.85 |          |
| Pw9kWLGcAj3VpmCn | 13-18, 20-22 | Warrior | mh   |                                     |    31 |  398 | 1.95 |          |
| QPmw1C9MNc4hqH3K | 1-5          | Warrior | oh   | timing + otsu for 5 swings          |    39 |  350 | 2.79 |          |

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

- Proc: weapon proc 1294243, logged as 1294243; rolls on main hand
- Sim rate: 2.35 PPM, observed; measured 2.34 PPM [2.15, 2.86], 8 wearers, 10 runs, 921 procs on 13895 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights                               | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------------------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1CYAWTtmcGdFqbPN | 1-5                                  | Warrior | mh   |            |    34 |  774 | 1.55 |          |
| 28AhM3VGnvFcJ4rD | 1-4                                  | Warrior | mh   |            |   168 | 2383 | 2.49 |          |
| 4NqLGxvrZYX2Qwnf | 1-4                                  | Warrior | mh   |            |    41 |  510 | 2.84 |          |
| 8KPHjdAxZYFNLtgR | 1-5                                  | Warrior | mh   |            |   129 | 1818 | 2.50 |          |
| 987q1pjnYdVQbCmT | 6                                    | Warrior | mh   |            |    34 |  384 | 3.12 |          |
| AXn7cJ9VQ6YHLWmh | 1-5, 7-9, 11-13, 15-18, 20-21, 23-30 | Paladin | mh   |            |    45 |  293 | 5.42 |          |
| Lg82ZFDyQ4MvRm9p | 1-2                                  | Paladin | mh   |            |    53 | 1120 | 1.67 |          |
| WYFZHxvL7zg6ydnQ | 1-6                                  | Paladin | mh   |            |   210 | 3081 | 2.41 |          |
| XHDjdLb7R9P6NQm1 | 1-2                                  | Paladin | mh   |            |   137 | 2494 | 1.94 |          |
| zLacx86Thr3AMtqn | 1-15                                 | Warrior | mh   |            |    70 | 1038 | 2.38 |          |

## Satyr's Lash (17752)

- Proc: weapon proc 1292155, logged as 1292155; rolls on main hand, off hand
- Sim rate: 2.1 PPM, measured; measured 2.09 PPM [1.90, 2.30], 9 wearers, 9 runs, 393 procs on 6621 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 4NqLGxvrZYX2Qwnf | 5-8      | Rogue   | oh   | timing + otsu for 47 swings        |    31 |  541 | 2.02 |          |
| 4PtKGL1zcRV8Chfn | 3        | Warrior | mh   |                                    |    47 |  728 | 2.28 |          |
| 4YzXaGtg81WdqM2N | 7-15     | Rogue   | oh   | timing + speed ratio for 5 swings  |     1 |   36 | 0.97 |          |
| Lg82ZFDyQ4MvRm9p | 14-26    | Hunter  | oh   | timing + speed ratio for 36 swings |    18 |  370 | 1.72 |          |
| R1acrQhyNvwXLmBj | 1-3      | Rogue   | oh   | timing + otsu for 96 swings        |    33 |  676 | 1.72 |          |
| T71XnfaF3cY9M2Rg | 1-13     | Rogue   | oh   | timing + otsu for 38 swings        |    57 |  946 | 2.13 |          |
| Yf91cBLN26jdqrxC | 1-4      | Hunter  | oh   | timing + otsu for 144 swings       |    21 |  334 | 2.22 |          |
| qm1Bvy29hGK6bMaQ | 1-2      | Warrior | mh   |                                    |   157 | 2553 | 2.17 |          |
| tdYLMmfgcAJCTWyr | 8, 10-12 | Rogue   | oh   | timing + otsu for 10 swings        |    28 |  437 | 2.26 |          |

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

- Proc: weapon proc 18208, logged as 18208; rolls on main hand, off hand
- Sim rate: 0.95 PPM, observed; measured 0.97 PPM [0.80, 1.25], 9 wearers, 9 runs, 249 procs on 6409 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 64NVwXYTpFRBxJWz | 9-11   | Rogue   | mh   | timing + otsu for 7 swings  |     4 |  211 | 0.47 |          |
| 6NrFMaRQ9yAjxWC2 | 11-12  | Rogue   | mh   | timing + otsu for 9 swings  |    37 | 1240 | 0.75 |          |
| AbnJ1XMLhwqtpx6F | 1-4    | Rogue   | oh   | timing + otsu for 21 swings |    10 |  380 | 0.66 |          |
| Axp6DdjvyT1Gwh4Y | 1      | Paladin | mh   |                             |    44 | 1334 | 0.82 |          |
| BM8NmpvwVdzA76f9 | 1-9    | Rogue   | mh   | timing + otsu for 30 swings |    59 | 1246 | 1.18 |          |
| Br2pycwj4zLC1ZP7 | 1-3    | Warrior | oh   | timing + otsu for 13 swings |    30 |  385 | 1.95 |          |
| CKLMhw2QbTVc8Hym | 1-14   | Rogue   | mh   | timing + otsu for 6 swings  |    37 | 1154 | 0.80 |          |
| VYb6f27NazpmCch3 | 1-2    | Warrior | oh   | timing + otsu for 17 swings |    10 |  212 | 1.18 |          |
| zbg1wFp2L8hWHZva | 1-26   | Warrior | oh   | timing + otsu for 10 swings |    18 |  246 | 1.83 |          |

## Searing Blade (12992)

- Proc: weapon proc 1299850, logged as 1299850; rolls on main hand
- Sim rate: 2.4 PPM, observed; measured 2.38 PPM [2.10, 2.74], 9 wearers, 9 runs, 621 procs on 6261 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights       | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------------ | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 1K6c48bWMN9nzCZA | 1-3          | Warrior | mh   |            |    62 |  553 | 2.69 |          |
| 3yM21KdFP7LnrBvx | 1-3          | Warrior | mh   |            |   103 |  991 | 2.49 |          |
| 6wxJrRWgPyQb21hm | 1-7          | Paladin | mh   |            |   109 | 1216 | 2.15 |          |
| 7ZxcPNTjqvfzQHgX | 1-2          | Hunter  | mh   |            |     6 |   54 | 2.67 |          |
| C7xfzTbPdAr1w9cX | 1-12         | Paladin | mh   |            |    66 |  882 | 1.80 |          |
| CGXFqgATBZdW1n7V | 7-15         | Paladin | mh   |            |   102 |  727 | 3.37 |          |
| CHZbFGkRPyWcrd9a | 1-3, 6, 8-10 | Paladin | mh   |            |    85 |  793 | 2.57 |          |
| PNYpz2hCmGMRV1tx | 1-3          | Hunter  | mh   |            |     0 |   10 | 0.00 |          |
| zjcGFyxpkKTnQXdw | 1-11         | Warrior | mh   |            |    88 | 1035 | 2.04 |          |

## Serpent Slicer (13035)

- Proc: weapon proc 17511, logged as 17511; rolls on main hand, off hand
- Sim rate: 1.1 PPM, observed; measured 1.08 PPM [0.95, 1.22], 6 wearers, 6 runs, 232 procs on 5156 hits
- Reports from 2026-11-14 to 2026-11-15
- Notes: swap noise: the swap drop moved PPM 0.86 -> 1.05 (20 active minutes, 18% of hits, in 1 runs)

| Report           | Fights    | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 9zxHaWbYT3gqDQvF | 8-18      | Paladin | mh   |                              |    52 | 1169 | 1.07 |          |
| C4TGJvzHF3aZ2NjR | 1-13      | Rogue   | mh   | timing + otsu for 117 swings |   100 | 2333 | 1.03 |          |
| PmM8nxa1p6QkhVLC | 1, 3-7    | Hunter  | mh   | timing + otsu for 85 swings  |    51 | 1005 | 1.22 |          |
| Td2y9DjVvn1YgaHr | 1-7, 9-21 | Hunter  | oh   | timing + otsu for 58 swings  |    15 |  308 | 1.17 |          |
| Z7qF4j3hazBTxHcw | 8-10      | Hunter  | oh   | timing + otsu for 49 swings  |     7 |  166 | 1.01 |          |
| gywYLPFQCJa9XfVG | 1-6       | Hunter  | oh   | timing + otsu for 116 swings |     7 |  175 | 0.96 |          |

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

- Proc: weapon proc 1293653, logged as 1293653; rolls on main hand, off hand
- Sim rate: 2.25 PPM, observed; measured 2.24 PPM [1.96, 2.56], 6 wearers, 6 runs, 295 procs on 5652 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights       | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out  |
| ---------------- | ------------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | --------- |
| 43wtgArG9vm8pb6N | 1, 3-4, 6-16 | Rogue   | oh   | timing + otsu for 2 swings  |    68 | 1313 | 2.22 |           |
| 9g7NQBqDH6FGfP8L | 4            | Rogue   | mh   | timing                      |    14 |  278 | 2.16 |           |
| JHrkTPq23LpAcGxY | 14-18        | Warrior | mh   |                             |    89 | 1596 | 2.39 |           |
| PZKA7G2kQtw9H4Dx | 4-9          | Rogue   | oh   | timing                      |    33 |  794 | 1.78 |           |
| PkWRr8jZp9ATL7N4 | 1-23, 25-26  | Rogue   | mh   | timing + otsu for 72 swings |    51 | 1146 | 1.91 |           |
| TbB8q4RMK3LmYJNv | 1-6          | Rogue   | oh   | timing + otsu for 29 swings |    40 |  525 | 3.27 |           |
| RJ1nvMk7AxTYm6tb | 14-18        | Warrior | mh   |                             |    89 | 1596 | 2.39 | duplicate |

## Spellsever Crossbow (274293)

- Proc: weapon proc 1293714, logged as 1293714; rolls on ranged
- Sim rate: 1 PPM, measured; measured 1.01 PPM [0.94, 1.09], 30 wearers, 34 runs, 734 procs on 18098 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights                    | Class   | Slot   | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------------------------- | ------- | ------ | ---------- | ----: | ---: | ---: | -------- |
| 1DVFRby49Pg2tpzG | 1-19                      | Hunter  | ranged |            |     6 |   80 | 1.88 |          |
| 1KF2jNqTzc6VJ9fp | 6-8                       | Hunter  | ranged |            |    13 |  368 | 0.88 |          |
| 1fHnGhFabmPTq6Dv | 7-17                      | Hunter  | ranged |            |    28 |  589 | 1.19 |          |
| 1jgTLvJQBMP6kxrm | 1-12                      | Hunter  | ranged |            |    36 | 1098 | 0.82 |          |
| 1qpQ6FW8fxD9gC4G | 1-7, 9-11, 13-27          | Hunter  | ranged |            |    68 | 1727 | 0.98 |          |
| 1w9jNgyZv7YKAacb | 1-13                      | Hunter  | ranged |            |    50 | 1053 | 1.19 |          |
| 214ndcFyjpPM8wah | 3, 6-9, 11-15, 19-28, 30  | Rogue   | ranged |            |     0 |    2 | 0.00 |          |
| 2WH4jLt3FADy8fnh | 1-2, 5-6                  | Hunter  | ranged |            |    38 |  938 | 1.01 |          |
| 2WH4jLt3FADy8fnh | 5-6                       | Rogue   | ranged |            |     0 |    0 |      |          |
| 3LhpDbvyrtNFWdxH | 1-5                       | Hunter  | ranged |            |    38 | 1162 | 0.82 |          |
| 3pCKd6jqgHyFc9kf | 1-17                      | Rogue   | ranged |            |     2 |   28 | 1.79 |          |
| 4KWV2rZDJMqNLRmT | 1-3, 6-7                  | Hunter  | ranged |            |     2 |   43 | 1.16 |          |
| 4NMcqLwQ2rB9amCH | 1-3                       | Hunter  | ranged |            |    19 |  523 | 0.91 |          |
| 6VJxcqnYa4wdBWQv | 1-9                       | Rogue   | ranged |            |     0 |    0 |      |          |
| 6kN78G2xfqy3vwTP | 1-6                       | Warrior | ranged |            |     1 |    3 | 8.33 |          |
| 8CzwkcFWrA91QaxG | 1-12                      | Warrior | ranged |            |     1 |   10 | 2.50 |          |
| D8JF7rbYQRcWxqn3 | 1-2, 4-8                  | Hunter  | ranged |            |    39 |  733 | 1.33 |          |
| Fwcmz4r9aDnpRWAd | 1-2                       | Rogue   | ranged |            |     0 |    0 |      |          |
| Gyw2T3LjptRW6vH9 | 1-6, 8-10                 | Hunter  | ranged |            |    34 |  867 | 0.98 |          |
| HARvQ2hJWp8f3grD | 4-9                       | Hunter  | ranged |            |    18 |  481 | 0.94 |          |
| J2XL6AyZqgpjx8QK | 8-13                      | Warrior | ranged |            |     1 |    9 | 2.78 |          |
| K6rZRbwYD1Wyjhpx | 1-4                       | Rogue   | ranged |            |     0 |    0 |      |          |
| KyYwHQjxAL8tzvdJ | 1-7                       | Hunter  | ranged |            |    27 |  645 | 1.05 |          |
| MpcVyhTHnqCJvDFx | 1-5                       | Hunter  | ranged |            |    41 | 1003 | 1.02 |          |
| PFGvpDf9WZgcY6Rk | 1-12                      | Hunter  | ranged |            |    44 |  868 | 1.27 |          |
| Vb3ZxfXtyDWr1ndP | 1                         | Hunter  | ranged |            |    31 |  912 | 0.85 |          |
| WYrCyJkdfMXw84Ph | 1-21, 23-35, 37-39, 41-48 | Hunter  | ranged |            |     5 |  201 | 0.62 |          |
| XHDjdLb7R9P6NQm1 | 1-2                       | Hunter  | ranged |            |    45 | 1046 | 1.08 |          |
| b12KD98fndF4hMq3 | 1-18                      | Hunter  | ranged |            |    75 | 1764 | 1.06 |          |
| bCzt1Fq24mYcPXkB | 1-2                       | Hunter  | ranged |            |    44 | 1362 | 0.81 |          |
| jJWt7Dnm4qc9rzQd | 1                         | Hunter  | ranged |            |     7 |  164 | 1.07 |          |
| rD8m4JW1bVyYPKGX | 1-3, 6                    | Rogue   | ranged |            |     0 |    3 | 0.00 |          |
| rQ4dFakMnWXj89vz | 1-6                       | Warrior | ranged |            |     0 |    3 | 0.00 |          |
| vgrm7nPFN231GYRW | 1, 3-15                   | Hunter  | ranged |            |    21 |  413 | 1.27 |          |

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

- Proc: weapon proc 1292657, logged as 1292657; rolls on main hand
- Sim rate: 2.25 PPM, observed; measured 2.23 PPM [1.92, 2.64], 14 wearers, 14 runs, 796 procs on 6479 hits
- Reports from 2026-11-10 to 2026-11-16

| Report           | Fights      | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 12A94BJt3KCFZrY8 | 1-12, 14-21 | Hunter  | mh   |            |    19 |  214 | 1.61 |          |
| 1a3pDtCcPgBkL6wm | 1-22        | Warrior | mh   |            |   119 |  604 | 3.58 |          |
| 1jgtdNmDMRycbZrP | 1           | Paladin | mh   |            |    64 |  477 | 2.44 |          |
| 1mRKL7Tb8qwjkgv3 | 1-2         | Warrior | mh   |            |    58 |  510 | 2.07 |          |
| 1yjwF8MxXqvtD4bT | 1-3         | Warrior | mh   |            |    37 |  317 | 2.12 |          |
| 23ZyGvk7BXAVDT4F | 1-14        | Warrior | mh   |            |    83 |  504 | 2.99 |          |
| 9zabj3grQkGB2FKL | 1-17        | Paladin | mh   |            |   153 | 1613 | 1.72 |          |
| BKnf4rNwJdcqpQ18 | 1-6         | Paladin | mh   |            |    76 |  562 | 2.46 |          |
| HARvQ2hJWp8f3grD | 1-9         | Hunter  | mh   |            |    12 |  123 | 1.77 |          |
| HtQr7kYaNq9MGgvZ | 1-12        | Warrior | mh   |            |    39 |  335 | 2.12 |          |
| LHRhtADNW8GfyYqT | 1           | Paladin | mh   |            |    76 |  646 | 2.14 |          |
| VQaYAf1Wxvm6ndHK | 1-11        | Hunter  | mh   |            |     2 |   80 | 0.45 |          |
| dQMKjpxJBCz1wrDW | 1, 3-12     | Warrior | mh   |            |    58 |  470 | 2.24 |          |
| nJtB3F7cQy6NmzGZ | 1-12        | Hunter  | mh   |            |     0 |   24 | 0.00 |          |

## Sul'thraze the Lasher (9372)

- Proc: weapon proc 1294430, logged as 1294430, 1294431; rolls on main hand
- Sim rate: 1.5 PPM, measured; measured 1.52 PPM [1.36, 1.61], 7 wearers, 7 runs, 991 procs on 15041 hits
- Reports from 2026-11-12 to 2026-11-16

| Report           | Fights                                    | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ----------------------------------------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 6kJNxp39PWqrghmG | 1-2                                       | Paladin | mh   |            |   242 | 3525 | 1.58 |          |
| Ag69GWdLRDhKCrcT | 1-6                                       | Warrior | mh   |            |    98 | 1445 | 1.57 |          |
| BFPtcmbyLr71Kj63 | 1-3                                       | Paladin | mh   |            |   155 | 2295 | 1.56 |          |
| BMJxG63gcbQX8tWh | 1                                         | Paladin | mh   |            |   202 | 2792 | 1.67 |          |
| Z78jbgTL9V1Yfahq | 1, 3-4, 6-10, 13-15, 17, 20-26, 29, 31-37 | Paladin | mh   |            |    55 | 1372 | 0.93 |          |
| fz2HQxXk96ywVYv7 | 1-8                                       | Paladin | mh   |            |   163 | 2520 | 1.49 |          |
| hCPLc2G7KAVX8Bby | 1-6, 8-9                                  | Warrior | mh   |            |    76 | 1092 | 1.61 |          |

## Supercharger Battle Axe (9486)

- Proc: weapon proc 13527, logged as 13527; rolls on main hand
- Sim rate: 1.95 PPM, observed; measured 1.97 PPM [1.69, 2.28], 4 wearers, 4 runs, 156 procs on 1698 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ---------- | ----: | ---: | ---: | -------- |
| 2axJHNwqQdtW1jnF | 1-3    | Hunter | mh   |            |    10 |   71 | 3.02 |          |
| 7kxNnWGw2FzfLJtR | 1-8    | Shaman | mh   |            |    80 |  856 | 2.00 |          |
| fDycAFZWt4NkRX1x | 1-3    | Shaman | mh   |            |    51 |  560 | 1.95 |          |
| rNwJc31qXfzWthQD | 1-5    | Shaman | mh   |            |    15 |  211 | 1.52 |          |

## Sword of Corruption (13032)

- Proc: weapon proc 17510, logged as 17510; rolls on main hand, off hand
- Sim rate: 1 PPM, observed; measured 0.99 PPM [0.83, 1.24], 8 wearers, 8 runs, 195 procs on 5621 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 3npHxB84DT2zRtmN | 1-4, 6 | Warrior | mh   |                             |    31 |  505 | 1.75 |          |
| 62WRbBjFQCpZGr3z | 1-4    | Rogue   | mh   | timing + otsu for 26 swings |    19 |  625 | 0.87 |          |
| 6XBqk2G4LZy83mQz | 1, 3-7 | Warrior | mh   |                             |    34 |  796 | 1.22 |          |
| 6dDjtwh1HXFxg72L | 1      | Warrior | oh   | timing + otsu for 40 swings |    20 |  552 | 1.03 |          |
| 7CcF8GHPv4fagQYh | 1-10   | Rogue   | mh   | timing + otsu for 25 swings |    25 |  742 | 0.96 |          |
| 7LY3ADmFM1wBarRn | 1-4    | Warrior | oh   | timing + otsu for 39 swings |    18 |  590 | 0.87 |          |
| QZcRmKX4vMVh7f6j | 1-14   | Rogue   | mh   | timing + otsu for 5 swings  |    37 | 1459 | 0.72 |          |
| ZHK32LtTCmqR4Pxf | 1, 3   | Warrior | oh   | timing + otsu for 21 swings |    11 |  352 | 0.89 |          |

## Sword of Decay (1727)

- Proc: weapon proc 13528, logged as 13528; rolls on main hand
- Sim rate: 1.1 PPM, observed; measured 1.09 PPM [0.94, 1.25], 6 wearers, 6 runs, 239 procs on 4887 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 9nCtfP6kvHrzV4FL | 2      | Rogue   | mh   | timing + otsu for 42 swings |    51 | 1332 | 0.85 |          |
| MGzd8mFZJgPLkAc2 | 1-2    | Paladin | mh   |                             |    38 |  634 | 1.33 |          |
| MXjF1GhRmLfabNZV | 1-18   | Warrior | mh   |                             |    69 | 1294 | 1.18 |          |
| ZHK32LtTCmqR4Pxf | 1, 3   | Warrior | mh   | timing + otsu for 21 swings |    25 |  425 | 1.31 |          |
| ZqghBpkcy7xtPfmH | 1-9    | Warrior | mh   | timing + otsu for 41 swings |    35 |  769 | 1.01 |          |
| nZ3NhbvpVXg649Bm | 4-10   | Rogue   | mh   | timing                      |    21 |  434 | 1.08 |          |

## Tainted Pierce (8225)

- Proc: weapon proc 13530, logged as 13530; rolls on main hand, off hand
- Sim rate: 2.15 PPM, measured; measured 2.13 PPM [1.96, 2.32], 6 wearers, 7 runs, 510 procs on 7555 hits
- Reports from 2026-11-14 to 2026-11-16

| Report           | Fights     | Class   | Slot | Hand split                   | Procs | Hits |  PPM | Left out |
| ---------------- | ---------- | ------- | ---- | ---------------------------- | ----: | ---: | ---: | -------- |
| 4na1LcCpP8QHTNVD | 1, 3-8     | Rogue   | mh   | timing + otsu for 65 swings  |    24 |  397 | 1.91 |          |
| 4wTtzhPfDqnQWNy7 | 1-4        | Paladin | mh   |                              |    78 | 1198 | 2.06 |          |
| 8WfLRh1qTpFxJgra | 1-2        | Warrior | mh   | timing + otsu for 53 swings  |    84 | 1065 | 2.49 |          |
| J97GvmrRXCFZfQ8a | 1-2        | Rogue   | mh   | timing + otsu for 15 swings  |    84 | 1276 | 2.08 |          |
| N3Y1mFkJ9Dwr74bA | 1-9, 11-12 | Paladin | mh   |                              |    90 | 1459 | 1.95 |          |
| NXgv93kr2G7zyTxm | 1, 3-9     | Rogue   | oh   | timing + otsu for 11 swings  |    82 | 1241 | 2.09 |          |
| qcrKJGPvkfhAtdjQ | 1-4        | Rogue   | oh   | timing + otsu for 196 swings |    68 |  918 | 2.34 |          |

## Taran Icebreaker (2915)

- Proc: weapon proc 16415, logged as 16415; rolls on main hand
- Sim rate: 0.9 PPM, observed; measured 0.89 PPM [0.79, 1.00], 3 wearers, 3 runs, 256 procs on 4550 hits
- Reports from 2026-11-13 to 2026-11-14
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights | Class  | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------ | ---- | ---------- | ----: | ---: | ---: | -------- |
| Bz8CAjZnNyXhgTtH | 1-14   | Shaman | mh   |            |    83 | 1479 | 0.89 |          |
| aAcjGPbyZXNTzR1K | 1      | Shaman | mh   |            |    91 | 1602 | 0.90 |          |
| mg3MBw6rcNtZpjbA | 1-14   | Shaman | mh   |            |    82 | 1469 | 0.88 |          |

## The Blackrock Slicer (13285)

- Proc: weapon proc 17407, logged as 17407; rolls on main hand
- Sim rate: 2.05 PPM, observed; measured 2.04 PPM [1.92, 2.35], 3 wearers, 5 runs, 1032 procs on 7598 hits
- Reports from 2026-11-11 to 2026-11-16
- Notes: few wearers (3): the interval is about the per-wearer range

| Report           | Fights                       | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | ---------------------------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| DW9YkdvyrjtMBQnf | 1-5                          | Paladin | mh   |            |   173 | 1405 | 1.85 |          |
| FHqYCNcKntvDAjxg | 25-31, 33-36                 | Paladin | mh   |            |   109 |  779 | 2.10 |          |
| NXMZV9xjvWpB7GtP | 10, 12-107, 109-116, 118-122 | Warrior | mh   |            |   222 | 1418 | 2.35 |          |
| gmKvNqHFJLZXfdnP | 1, 3-6                       | Paladin | mh   |            |   295 | 2230 | 1.98 |          |
| vwCXzhMQ2drgpRxD | 1-41                         | Shaman  | mh   |            |   233 | 1766 | 1.98 |          |

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

- Proc: weapon proc 17505, logged as 17505; rolls on main hand, off hand
- Sim rate: 2.1 PPM, observed; measured 2.08 PPM [1.85, 2.52], 6 wearers, 7 runs, 583 procs on 9345 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights    | Class   | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | --------- | ------- | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 1ZyhaVDdFtTYWMcC | 4-7, 9-12 | Rogue   | oh   | timing + otsu for 153 swings       |    47 |  687 | 2.28 |          |
| 2Z4G3r9pfqJtCM6W | 7-17      | Paladin | mh   |                                    |   245 | 4534 | 1.80 |          |
| 3BAz1TDGL9kjy8ZQ | 1-2       | Warrior | oh   | timing + speed ratio for 45 swings |     2 |   32 | 2.11 |          |
| C1cyPbL3RQvGjq8p | 48        | Warrior | oh   | timing + otsu for 5 swings         |    33 |  294 | 3.74 |          |
| Qb1yWxa8vHCGL7Yr | 2-7, 9-15 | Rogue   | oh   | timing + otsu for 15 swings        |    44 |  710 | 2.06 |          |
| qcdyZAr2ganDtxNW | 13-44     | Warrior | mh   |                                    |   167 | 2270 | 2.45 |          |
| rRnVJhQFa9zf7NqK | 4-5       | Rogue   | oh   | timing + otsu for 28 swings        |    45 |  818 | 1.83 |          |

## The Ziggler (8006)

- Proc: weapon proc 13482, logged as 13482; rolls on main hand, off hand
- Sim rate: 3.6 PPM, observed; measured 3.62 PPM [3.16, 4.03], 9 wearers, 9 runs, 853 procs on 8318 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights | Class | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ----- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 2bYvT8kWxjFfdwrm | 1-2    | Rogue | oh   | timing + otsu for 29 swings |    46 |  592 | 2.74 |          |
| 3DQy7BhjH1mnZPpq | 9-12   | Rogue | mh   | timing + otsu for 3 swings  |    63 |  468 | 4.75 |          |
| 4xnDCWwrGKAFdNHv | 1      | Rogue | oh   | timing + otsu for 18 swings |    67 |  751 | 3.15 |          |
| 9MmVG8crjzJZ24K7 | 1-4    | Rogue | oh   | timing + otsu for 10 swings |    63 |  810 | 2.75 |          |
| A2FMrK3Y1xkvafQH | 1-7    | Rogue | mh   | timing + otsu for 15 swings |   141 | 1273 | 3.91 |          |
| BTRMkxmvfrwcZNaG | 1-3    | Rogue | oh   | timing + otsu for 14 swings |    77 |  839 | 3.24 |          |
| BmfFgAczHXN7b16M | 9-24   | Rogue | mh   | timing + otsu for 28 swings |   208 | 1862 | 3.94 |          |
| R9njq1AGDvmp36bM | 1-8    | Rogue | oh   | timing + otsu for 85 swings |    80 |  949 | 2.98 |          |
| x3PftMnXCBZgLAG2 | 1, 4-9 | Rogue | oh   | timing + otsu for 3 swings  |   108 |  773 | 4.93 |          |

## Thornspike (6681)

- Proc: weapon proc 1293183, logged as 1293183; rolls on main hand, off hand
- Sim rate: 3.75 PPM, observed; measured 3.77 PPM [3.04, 4.40], 13 wearers, 13 runs, 807 procs on 7565 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split                         | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | ---------------------------------- | ----: | ---: | ---: | -------- |
| 1VKQY2Dafc7PJzpR | 1-2      | Warrior | mh   | timing + otsu for 114 swings       |    55 |  684 | 2.84 |          |
| 1Zv6bBfgDP3RVFYa | 1-2, 4-7 | Rogue   | oh   | timing                             |    12 |  134 | 3.16 |          |
| 1w8AbW4RBNLX6Kfy | 1        | Warrior | mh   |                                    |    15 |  333 | 1.59 |          |
| 2KkaByn6H9mzb4Yq | 1-4      | Rogue   | oh   | timing + otsu for 151 swings       |    76 |  859 | 3.12 |          |
| 2t6mDPLkfWBxcr4q | 1-18     | Hunter  | mh   | timing + speed ratio for 30 swings |     5 |   46 | 3.83 |          |
| 39BFWKC2J8RTzyMx | 1-3      | Rogue   | oh   | timing + otsu for 2 swings         |   123 |  760 | 5.71 |          |
| 72gPMatGLpwd93jW | 13       | Rogue   | mh   | timing                             |    24 |  192 | 4.41 |          |
| MdaKf7zpNgWHF1Jj | 1-4      | Hunter  | oh   | timing + speed ratio for 21 swings |     0 |   20 | 0.00 |          |
| PFGvpDf9WZgcY6Rk | 1-12     | Rogue   | mh   | timing + otsu for 39 swings        |   203 | 1624 | 4.41 |          |
| VYb6f27NazpmCch3 | 1-2      | Shaman  | mh   |                                    |    51 |  394 | 4.57 |          |
| b12KD98fndF4hMq3 | 10-18    | Rogue   | mh   | timing + otsu for 7 swings         |    94 |  758 | 4.38 |          |
| fnMH4F2pkGC8T6Kz | 1-8      | Shaman  | mh   |                                    |    79 | 1103 | 2.53 |          |
| xJbw6RTvtHPncNzp | 1-5      | Warrior | mh   |                                    |    70 |  657 | 3.76 |          |

## Toofbreaker (286430)

- Proc: weapon proc 1320269, logged as 1320269; rolls on main hand, off hand
- Sim rate: 1.95 PPM, measured; measured 1.96 PPM [1.86, 2.06], 14 wearers, 16 runs, 1364 procs on 21973 hits
- Reports from 2026-11-13 to 2026-11-16

| Report           | Fights   | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | -------- | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 1BNXynxVbH4kGt8R | 1-2      | Warrior | mh   |                             |    19 |  305 | 1.97 |          |
| 2PZCz9c1HgtrX7Th | 3-5      | Warrior | mh   |                             |    65 |  891 | 2.30 |          |
| 2taZz4vTdfXYAk3b | 1-2      | Rogue   | oh   | timing + otsu for 87 swings |    39 |  609 | 2.02 |          |
| 2xY1D3mCwagdy7Wc | 1-2, 4-6 | Warrior | mh   |                             |    66 | 1086 | 1.92 |          |
| 2yqzTJfLN8cvwaDK | 1-2, 4-6 | Warrior | mh   |                             |    60 |  912 | 2.08 |          |
| 3JbjLGXVzctAmfaD | 1-4      | Warrior | mh   |                             |    80 | 1137 | 2.22 |          |
| 4YVTAqgG3dtmD8C9 | 1-23     | Warrior | mh   |                             |   172 | 2636 | 2.06 |          |
| Cp3ND2J1xWhbFzP7 | 1-26     | Warrior | mh   |                             |   126 | 2142 | 1.86 |          |
| LPQd7pfbjDkYZyRz | 1-2, 4-5 | Rogue   | oh   | timing + otsu for 2 swings  |    42 |  639 | 2.08 |          |
| LPQd7pfbjDkYZyRz | 1-2, 4-5 | Paladin | mh   |                             |   117 | 1852 | 1.99 |          |
| YakBLWFpjnrdhGb3 | 1-2      | Paladin | mh   |                             |   182 | 3211 | 1.79 |          |
| cFRpkf1jWTaHrgL3 | 1-2      | Paladin | mh   |                             |   156 | 2667 | 1.85 |          |
| qnyc1zwaB4FrvM9x | 1        | Paladin | mh   |                             |    55 | 1059 | 1.64 |          |
| tac7hXVRgkbZNMDy | 3        | Warrior | mh   |                             |    54 |  892 | 1.91 |          |
| vtyhfFz9wPGJdHqM | 1-3      | Warrior | mh   |                             |    50 |  666 | 2.37 |          |
| ynVw7ZgMx4tjF3fQ | 10-18    | Paladin | mh   |                             |    81 | 1269 | 2.02 |          |

## Twilight Forgehammer (281665)

- Proc: weapon proc 1312176, logged as 1312176; rolls on main hand, off hand
- Sim rate: 3.15 PPM, observed; measured 3.16 PPM [2.54, 3.84], 6 wearers, 6 runs, 1344 procs on 10224 hits
- Reports from 2026-11-09 to 2026-11-15

| Report           | Fights | Class   | Slot | Hand split                  | Procs | Hits |  PPM | Left out |
| ---------------- | ------ | ------- | ---- | --------------------------- | ----: | ---: | ---: | -------- |
| 6WgpmHQrv9kn1ZXJ | 3-22   | Paladin | mh   |                             |   460 | 2846 | 3.88 |          |
| 6b4F7HrzNTnxRMXd | 1-13   | Warrior | oh   | timing + otsu for 58 swings |   177 |  809 | 5.25 |          |
| JRr2jLy8vNx147fX | 1-27   | Paladin | mh   |                             |   128 | 1131 | 2.72 |          |
| Ztv8xkT6dmQ7n91C | 1      | Warrior | mh   | timing + otsu for 19 swings |   295 | 2658 | 2.66 |          |
| b4hn6GgpFLtJ1N2X | 1-13   | Warrior | mh   | timing + otsu for 13 swings |   132 | 1457 | 2.17 |          |
| m8RLWaY7bh9jHKwD | 2      | Rogue   | mh   | timing + otsu for 2 swings  |   152 | 1323 | 2.76 |          |

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

- Proc: weapon proc 1302166, logged as 1302166; rolls on main hand
- Sim rate: 1.9 PPM, observed; measured 1.92 PPM [1.83, 2.01], 4 wearers, 5 runs, 1500 procs on 15142 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: few wearers (4): the interval is about the per-wearer range

| Report           | Fights | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out  |
| ---------------- | ------ | ------- | ---- | ---------- | ----: | ---: | ---: | --------- |
| AyTmpNtxB9jbnJ2H | 4-44   | Shaman  | mh   |            |   318 | 3245 | 1.90 |           |
| DpVmH4WRxL72nGwd | 3      | Shaman  | mh   |            |   127 | 1275 | 1.93 |           |
| NrpM7DJQBKWqbzVF | 1-7    | Paladin | mh   |            |   164 | 1784 | 1.78 |           |
| VYBjgJx8aZtRhmpy | 1-43   | Shaman  | mh   |            |   691 | 6951 | 1.92 |           |
| q23Tz8hnZGkvbLx6 | 1-18   | Shaman  | mh   |            |   200 | 1887 | 2.05 |           |
| jnWqct213HRGAM76 | 1      | Shaman  | mh   |            |   107 | 1106 | 1.87 | duplicate |

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

- Proc: weapon proc 1293787, logged as 1293787; rolls on main hand
- Sim rate: 2.1 PPM, observed; measured 2.11 PPM [1.94, 2.36], 6 wearers, 6 runs, 492 procs on 4231 hits
- Reports from 2026-11-13 to 2026-11-16
- Notes: swap noise: the swap drop moved PPM 1.66 -> 2.11 (34 active minutes, 21% of hits, in 1 runs)

| Report           | Fights         | Class   | Slot | Hand split | Procs | Hits |  PPM | Left out |
| ---------------- | -------------- | ------- | ---- | ---------- | ----: | ---: | ---: | -------- |
| 176QTFWR8PtdfGZB | 25-32          | Paladin | mh   |            |    52 |  425 | 2.22 |          |
| 1WHqx7bPa8mkGydJ | 1, 3           | Warrior | mh   |            |   144 | 1313 | 1.99 |          |
| 6QvVFqNYty8bzrcf | 1, 3-11, 13-21 | Paladin | mh   |            |   117 |  828 | 2.57 |          |
| K8rhAtLjcz7gdHkn | 6-7            | Warrior | mh   |            |    23 |  233 | 1.79 |          |
| P1yKzX3Q8NgFYxZH | 1-3            | Warrior | mh   |            |   127 | 1158 | 1.99 |          |
| V3L7kJDzTGg4QqCF | 1-3            | Paladin | mh   |            |    29 |  274 | 1.92 |          |

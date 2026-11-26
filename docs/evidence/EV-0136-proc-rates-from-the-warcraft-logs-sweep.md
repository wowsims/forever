# EV-0136: Proc rates of 101 items and a set bonus from the Warcraft Logs sweep

- Claim: each item and set bonus below procs at the rate its section gives, priced on each landed hit's weapon base speed;
  a rate whose item has 5 or more wearers and a 95% interval within 10% of it either side is measured, every other one
  observed, except in the sections that name a later count, which take that count's test; and a Windfury Totem hit
  counts as a melee swing for the procs whose rows hear melee swings, Unsealed Wrath excepted. In these reports, up to
  2026-11-16, the Windfury hit grants no rage and spends a warrior Flurry charge; from push 114115 it grants a main-hand
  white hit's rage and spends no Flurry charge (EV-0138).
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section's table names; 600
  reports in all for the rates, and the 209 reports the Windfury Totem section lists
- Game build: 1.60.1.70425 to 1.60.1.70938 by report date; each item's section gives its date span, and the hotfix each
  player ran is not recorded
- Date: 2026-10-08 to 2026-11-16
- Pull request: #1065
- Later count: Julie's Dagger and Waveslicer, §Count of the 2026-11-13 to 2026-11-17 reports
- Later count: Flame Wrath, §Count of the 2026-11-12 to 2026-11-18 reports
- Later count: Bow of Searing Arrows, Coldrage Dagger, Cursed Felblade, Darkwater Talwar, Electrocutioner Leg, Glutton's
  Cleaver, Grimclaw, Hammer of Expertise, Hanzo Sword, Ironforge Greathammer, Lord General's Sword, Sang'thraze the
  Deflector, Satyr's Lash, Taran Icebreaker and Toofbreaker, §Count of the 2026-11-05 to 2026-11-19 reports
- Later count: Bloodrazor, Cookie's Tenderizer, Kang the Decapitator, Pyromancer's Baton, Scorpion Sting, Searing Blade,
  Serpent Slicer, Sliverblade, Sul'thraze the Lasher, Sword of Corruption and Twilight Forgehammer, §Count of the
  2026-11-06 to 2026-11-21 reports
- Later count: Burning War Axe and Shortsword of Vengeance, §Count of the 2026-11-09 to 2026-11-22 reports
- Later count: Axe of Inevitable Ruin, Barbaric Crossbow, Bashguuder, Black Duskwood Staff, Black Malice, Blade of the
  Wretched, Bloodspiller, Bonechill Hammer, Boneslicer, Brush Cutter, Claw of Celebras, Claw of the Shadowmancer, Cobalt
  Crusher, Darrowspike, Dazzling Longsword, Deadhead Blade, Dwarven Hand Cannon, Eater of the Dead, Fang of the Crystal
  Spider, Fiery War Axe, Firebelcher, Firebreather, Fist of the Damned, Frightalon, Gahz'rilla Fang, Galgann's
  Fireblaster, Grim Reaper, Gryphon Rider's Stormhammer, Gut Ripper, Hameya's Slayer, Hammer of the Northern Wind,
  Iceblade Hacker, Iron Knuckles, Jom-jom's Gabbar, Killmaim, Linken's Sword of Mastery, Loksey's Training Stick, Mana
  Channeling Wand, Naga Heartpiercer, Night Reaver, Nightblade, Pendulum of Doom, Phantom Blade, Quillshooter,
  Rivenspike, Skeletal Club, Spellsever Crossbow, Stalvan's Reaper, Strike of the Hydra, Supercharger Battle Axe, Sword
  of Decay, Tainted Pierce, The Blackrock Slicer, The Ziggler, Thornspike, Twilight Maul, Venomstrike, Vibroblade and
  X'caliboar, §Count of the 2026-11-06 to 2026-11-22 reports
- Later count: Felstriker, Shadowblade and The Cruel Hand of Timmy, §Count of the 2026-11-13 to 2026-11-26 reports

## Method

A proc is an event of the item's proc spell on the wearer or the target (damage, an aura applied or refreshed, a heal,
extra attacks, energy), counted once when one proc logs several events within 50 ms; a proc that misses or is resisted
still counts. The hits are the wearer's landed, non-periodic hits the trigger's proc flags hear: white swings of the
item's hand and melee-defense hits on the main hand (abilities, seals, Judgement of Command, Windfury Totem hits) for a
melee proc, Auto Shot, Shoot, Throw and ranged-defense hits on the ranged weapon for a ranged one. Left out: periodic
ticks, damage shields, and Volley, Shank, Shield Slam, Shield Bash and Thunder Clap. A dual wielder's white swings are
split between the hands by swing timing: each hand swings at its weapon's base speed over the haste of the moment, so
the swings form two interleaved chains, and the few a chain cannot place fall back to the damage before crits or the
weapon speeds (each run's table row names its split). A druid run in Cat or Bear Form is left out, a run logged in two
reports counts once, and a stretch with no proc where several were due is dropped as the item swapped away. The rate is
the procs over the sum of each hit's weapon base speed over 60, so haste raises procs a minute without moving it. Its
interval is the wider of the per-hit Wilson interval and 2,000 bootstrap draws of whole wearers.

The method reproduces the Highborne Research Tablet's weapon hits (EV-0131, 2.70 against 2.71). On Flurry Axe the
warriors' off-hand Whirlwind hits roll the axe and are not counted as hits here, so no item rolled mostly by the off
hand counts as measured, except Satyr's Lash, whose hits fall half in each hand, by the owner's ruling. Against clean
damage splits the timing split agrees on 86% of swings and puts the main hand's share a median 0.025 from the weapon
speeds' share. Items with fewer than 50 procs, or one wearer with fewer than 100, are left out.

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

## Windfury Totem hit as a swing

- Reports from 2026-11-11 to 2026-11-16; the reports are listed at the end of the section
- Scripts: `chain.py` for procs and `charges.py` for spent charges, run by `run_all.sh`; `reckoning.py` for Reckoning's
  extra attacks, `swingtimer.py` for the swing timer, and `paired.py` for rage

A Windfury Totem hit counted here is a landed one (10610 and the lower ranks 10608 and 8516) that follows a landed
melee ability (melee DefenseType) within 60 ms, where the wearer lands no other melee hit in the 100 ms before the
ability or in the 100 ms after it. Its proc is an event of the proc's spell in the 100 ms after the ability. A melee
ability alone almost never procs these rows, so the rate after an ability and a Windfury hit is the Windfury hit's own;
the ability-alone rate is in the table beside it. The white swing rate is the same count after a landed white swing with
no other hit within 100 ms. A charge spend is counted only while the charged aura is up, and for both Flurries only
where neither the triggering hit nor the Windfury hit crit, since a crit sets the charges again. Intervals are Wilson
intervals per hit.

| Proc                                      | Row     | Wearers | Windfury hits | With proc | Per Windfury hit  | After an ability alone | Per white swing   |
| ----------------------------------------- | ------- | ------: | ------------: | --------: | ----------------- | ---------------------- | ----------------- |
| Unbridled Wrath, rage                     | 12322   |      78 |        11,785 |     6,418 | 54.5% (53.6–55.4) | 0.26%                  | 55.1% (54.8–55.4) |
| Flurry (warrior), a charge spent          | 12966   |      36 |         1,880 |     1,861 | 99.0% (98.4–99.4) | 0.56%                  | 98.3% (98.1–98.5) |
| Reckless Swings, rage and a charge        | 1329555 |      46 |         1,475 |     1,469 | 99.6% (99.1–99.8) | 0.47%                  | 97.5% (97.1–97.8) |
| Quick Strikes (Aspect of the Beast)       | 1299447 |      16 |           727 |        62 | 8.5% (6.7–10.8)   | 0.24%                  | 8.6% (8.1–9.1)    |
| Rogue Armor Energize (rogues)             | 27787   |       5 |           737 |        15 | 2.0% (1.2–3.3)    | 0.05%                  | 2.5% (2.2–2.8)    |
| Rogue Armor Energize (druid)              | 27787   |       1 |            79 |         1 | 1.3% (0.2–6.8)    | 0 of 334               | 1.8% (1.1–3.1)    |
| Reckoning, the stored charges spent       | 1326084 |      18 |            25 |        25 | 100% (86.7–100)   | 0.36%                  | 80.4% (74.4–85.3) |
| Battlegear of Valor, heal                 | 450587  |       9 |         1,299 |        37 | 2.85% (2.04–3.87) | 0.03%                  | 2.68% (2.28–3.14) |
| Flurry (shaman), a charge spent (EV-0010) | 16257   |      55 |           357 |       286 | 80.1% (75.7–83.9) | 4.1%                   | 94.3% (93.9–94.8) |
| Seal of Command                           | seal    |      37 |         1,043 |       313 | 30.0% (27.3–32.9) | 0.73%                  | 25.7% (24.9–26.4) |
| Seal of Fury                              | seal    |      50 |         1,450 |     1,095 | 75.5% (73.2–77.7) | 1.5%                   | 69.1% (68.6–69.6) |
| Seal of Righteousness                     | seal    |      36 |           821 |       255 | 31.1% (28.0–34.3) | 0.89%                  | 39.3% (38.6–40.1) |
| Unsealed Wrath, an Echo stored            | 1310735 |      26 |           519 |        10 | 1.9% (1.1–3.5)    | 1.0%                   | 66.6% (65.5–67.7) |
| Bloodthrill                               | 1289682 |       9 |         1,633 |        10 | 0.61% (0.33–1.12) | 0.74%                  | 3.5% (3.1–4.0)    |
| Lightforge Armor, Crusader's Wrath        | 450625  |       2 |            33 |         0 | 0% (0–10.4)       | 0 of 375               | 3.4% (2.1–5.4)    |

Every row whose proc flags hear melee swings and nothing else rolls the Windfury hit as a main-hand swing. Unbridled
Wrath, the warrior's Flurry, Reckless Swings, Quick Strikes, Rogue Armor Energize and Battlegear of Valor roll it at the
rate a white swing does; the shaman's Flurry and Reckoning hear it too. A seal's row is each rank's own. Unsealed
Wrath is the exception: a Windfury hit stores an Echo no more often than the ability before it does alone. Bloodthrill,
whose row also hears abilities but which procs on Heroic Strike and Cleave alone (EV-0056),
does not hear it either. Lightforge Armor's two wearers are too few to say. The seals hear the Windfury hit as the sim
has them.

A Windfury hit spends Reckoning's stored charges, landed (101 spends) or avoided (87), but it brings no extra attack:
none of the 26 spends by a Windfury hit that followed an ability logs Reckoning's extra attacks 20178. A Windfury hit
that follows a white swing, which may spend the charges itself, logs them on 74 of 160 spends, so those are left out.

A white swing that a Windfury hit follows spends 1.97 warrior Flurry charges and 1.96 Reckless Swings charges, against
0.99 and 0.98 for a white swing alone.

The Windfury hit does not reset the swing timer: the gap between the swings before and after an ability and a Windfury
hit is a median 1.002 of the wearer's median swing gap (3,314 such gaps), against 1.001 after an ability alone (26,063).

In these reports the Windfury hit grants no rage; from push 114115 it grants a main-hand white hit's rage (EV-0138).
Cast events carry the warrior's rage before the cast's cost, so the gain between two casts is read straight off them.
Over gaps under 600 ms with no damage taken and no resource change, 88 of 90 Windfury hits from 13 warriors added
nothing over the same warrior's gaps opened by the same ability without a Windfury hit. The Windfury hits dealt 1 to 509
damage, median 204, where a white swing of about 200 damage gives 10 to 19 rage. Those runs are DGFd8KtQk19hXabN fight
5, rfg8k3Nw1P7TAHx6 6, AfJ6DYbh9dmkgHZL 7, cgMGCdPJQ84wzaBV 5 and 183, xRK37aCvtVfj26kL 4, FN8krKfCb3BXWmwV 3,
JRQqNL6DkvWKYdPB 2, 4Tc3yBNVYn7wW8RJ 63, 7fy93LdcwQhvF2gP 5, 1JLq46yg7wYNhan9 1, jXTVfy3ApWKPYwDG 3
and 4YVTAqgG3dtmD8C9 4.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1294226
- Rate: 2.05 PPM, measured

667 procs on 6,517 rolling hits, 6 wearers, 7 runs, in the main hand: all hits 2.05 [2.00, 2.10]. The rate is all hits,
which the one-hand white rate agrees with. One-hand white: 2.24 [1.81, 3.22], 132 procs on 1,179 lone units, 6 wearers.
Specials: 5.71 [0.69, 20.64], 2 procs on 7 lone units, 1 wearer (a Poisson interval: under 5 wearers). Direct: window
[-11, 100] ms, median 0 ms, 98th percentile 1 ms, 122 isolated hits. Reports from 2026-11-13 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       4 |   658 | 6,451 | 2.04 |
| Hunter |       2 |     9 |    66 | 2.73 |

Reports and fights: LMdRqJGhV17XwWyT 16-22; MpcVyhTHnqCJvDFx 1-6; WFaZNV2mhkrHBPz6 1-2; WdqtNK9nVDmXPa2f 1;
fzR7DJAFNWbk6LTH 1; mqzgv4rPVxA3JHGf 1-2; yhW9cLwxnHNDP2jF 4-8.

## Barbaric Crossbow (272999)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1291551
- Rate: 3.05 PPM, measured

6,999 procs on 69,394 rolling hits, 97 wearers, 118 runs, in the ranged slot: all hits 3.03 [2.96, 3.09]. The rate is
all hits, which the lone Auto Shot and specials rates agree with. Lone Auto Shots: 3.04 [2.94, 3.14], 3,917 procs on
38,701 lone units, 97 wearers. Specials: 3.03 [2.91, 3.15], 2,370 procs on 23,475 lone units, 78 wearers; Arcane Shot
gives 768 procs on 7,496 of them. Direct: window [-10, 100] ms, median 19 ms, 98th percentile 38 ms, 5,285 isolated
hits. Dropping the swapped stretches (140 active minutes, 5% of hits, in 10 runs) moved the read from 2.88 to 3.03.
Reports from 2026-11-07 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Hunter  |      78 | 6,983 | 69,205 | 3.03 |
| Warrior |      12 |    12 |    141 | 2.55 |
| Rogue   |       7 |     4 |     48 | 2.50 |

Reports and fights: 1AfDWYQT7Xjd24K3 13-14, 16; 1Gg92cW6XBhkwYb4 1-5; 1RqPMwNghAb6rcmZ 1-5; 1RrDAhp7ZvVYc9P3 1-6;
1W9N26dHfnPV4G7A 1, 3-9; 1y6GV7KLTZPrJH2w 1-3; 2GbXqD7KcWTfRnYa 1-2; 2MHXV8Y9jx43vhpk 3-12, 14-18; 2MrQxAndPNYLVaKR 1-7;
2RygmFXpHCbnazVA 1-5; 2a3hGM7bC8c4yZYv 1-2; 2mNyRawv7fGMDYVx 1; 2tJkzh16X38wrQfP 1-3; 3pTj2FQg6VvJz4DP 4, 6-7;
47fTkRKLxMVBQrb2 1; 4ChPnBrVK2fJLMaX 4; 4RHbVv3t8BLYqyNz 1-2; 4pcnNWw3rFmKb82G 1-2; 6KjMDQfzRFq8ZwVp 1; 6wcRHajznKf79FYA
1, 3; 72XWYanybGdwPg41 1; 7VkrJpR8XyFtL6Hm 1-2; 7vyJCbH4kw8ZdfFV 1-10; 8LhHr9PGtAKzYa6X 1; 8Xxb9LDGHP7vcJfh 1-5;
8ZmTJ74jCYP3QGBq 1-8; 92JNtax7RTkQ6cHv 1; 9kGgJ3qN6RMbnhf1 1-2; 9qTHdZQfvhCVP7RB 1; CVFcbDGj6n8WMKyL 1-2;
DACv8aKzbkxfGygr 1, 3-4; DHv3m4KLCdntG7pV 1-3; DZvkgfwrVLbpmWyQ 1-2; DaNQ6WPf7GqTpKVk 1-8; Fqa2PczWH8YfXCd9 1;
Ghp9gxWFkBNacLfn 1-5; GkLDK8WpcdvPatfB 1-2; Gm7xJ3W81tHyTfXc 1-3; H9QxwyGt4z8MmqCj 1; HwqhFkWZCBRTQVDf 5-7, 9-13;
J4mG8yjgVCALMfRw 1-2; KBzML9kdHFCXVpQc 1-2; KYD67nVqZT4jxWyN 5-6; L7QHtkhacypgd3D1 1-2; LnPRx96hKZN3CgTF 1;
MfD1n7XFZrG4dcbH 1516 fights from 1 to 1743; My1pRBTfHgDqm2vQ 1-2; N1LFPWYgfxdr8BQG 1-4; NFXZAv4JRPTVGWcx 1, 3-9;
NGydRxVDPzJCnjg7 1-3; NL2z3kMT4wpXmAjr 1-3; NQyGkApg7zMc18La 1-3; NyG7zfKj2Mvk1FLJ 1-6; P7MBDvyCRQtLkJHz 1, 3-8;
PWfNR3ykhJ6DHcTr 1-3; PkG3v1AJpRFYdbqy 1-7; QJMhZN4j2ycBpzWP 1-2; QTRY1FPqxJmMZbNc 1; QmMr87NnWXjKZTVp 1-4;
QtGYMWzJZPNpmjvB 1; T6MXt73janNJgmyY 1-3; Td29Nzahj1cmGMWZ 1-2; TdXmWDR3QFJwpYfq 1; VXcPMWHwZJftbCrm 1-9;
VcDy8ZTRgQpWrqXz 1-3; WN1vpX6t7JkVDYZf 1-6; WqKvGDhb34Bj6mPL 1-2; XV7Cf348hazGFg1t 1; Xp6vKQcnYgbDh7Hm 1;
YHh2t9agcvQjTAZq 2-3; Z9CQ6fqYjzcB4tG2 1-3; ZyM7xXKHnrW1Y843 1; aMQ63JtDnpmFz9WR 1-5; aPLxyc4GrBXF2gDK 1-3;
akDmAcjZrpgTFCQM 1-13; b3Dyw2FXtGcLvkdY 1, 3; bLd41TDnmVYJZj2x 1-3; cAdCm1NLGaQnXrHW 1-4; dTh4GAWLxwRm9V6k 1-2, 4;
dc3Y84hWFnrA2wqp 1-3; dtLy89QkrcHx3pMf 1-2; fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; gBxakRwhQDvnLXp1 1-2;
hMdFnHqgVRDTyKPY 1; hRzbpMrDY3cgQX7n 1-7; hqyNfmTQKRMLaj8z 1; jRytczTZF6pBJgYL 1-2; kBHbV3DZjamvfJY4 1-3;
nALy2WrwHC9qdMFm 1-2; nrhX2cvmtxG36JQN 1-2; p9HgV7zdCKqbrwJ3 1-3; ph6FqnCZaRQLvxbM 1-2; pw8aYCczKgjMbvAd 1-4;
pydvC4YzqVMRgPAf 1-2; qYK34hWcn1CyRzfp 1-3; r2BYfmZW1wkLDdNx 1-2; r41hkm8BMfLRyxFH 1; rA4j69wmVfZk8NDJ 1;
rKXWhHqLza39gkmA 1-2; rLQHdFx3P4MfkwV8 1-3; tbWaYgLTw8XQAzdC 1-3; v9fnrCd8TLaPDYH4 1-2; vqpcV7yDGCkbftmP 1;
vrZD4X7Vb3pW9FNK 1, 3-4; vryhGHtbLz3wfpCB 1-4; vxkFd3MfgJX7NyAc 1, 3-8; wmT1zqYBat3fHDLy 1-5, 8-9; xXTm2jy7BZA689wP
1-19; xfqQ7bDrcwLaY39n 1; xrGFH2DzcNBZq3Cv 1-3; yFmQLHx3VC6Kb1at 1; yZc4vgDq1HwfQMTp 1-2; yqH3GwmL7vt14nY6 1;
zcL8MvJnPXxYhNyC 1.

## Bashguuder (13204)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 17315
- Rate: 3 PPM, observed

6,293 procs on 69,413 rolling hits, 39 wearers, 53 runs, in the main hand or the off hand: all hits 3.02 [2.85, 3.20].
Puncture Armor stacks on the target, and a refresh at full stacks can log nothing, so no lone-unit rate checks the
all-hits rate. Observed: no lone-unit rate checks it. One-hand white: 2.66 [2.43, 2.92], 794 procs on 9,941 lone units,
12 wearers. Dual-wield white, which sets no rate: 2.28 [1.99, 2.60], 1,067 procs on 15,575 lone units, 27 wearers.
Specials: 2.86 [2.62, 3.07], 879 procs on 10,227 lone units, 15 wearers; Cleave gives 550 procs on 6,323 of them.
Two-roll abilities, priced as one roll: 4.07, 15 procs on 123 lone units (Execute 15 on 123), left out of the rate.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 2,125 isolated hits. Dropping the swapped stretches (83
active minutes, 3% of hits, in 6 runs) moved the read from 2.94 to 3.02. Report xXWkB3fqAjt9NJGH logs the same run as
dnDg7j9kRw6WtbML and counts once. Reports from 2026-11-11 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      29 | 5,659 | 62,407 | 3.02 |
| Rogue   |      10 |   634 |  7,006 | 3.02 |

Reports and fights: 28xgKPbf6GXjzHZ3 4-7; 6mD4XcqZR2YnraGB 1; 6xnwM2qCN43Tv8Af 1-9; 7Kz3LtgfDmbBRGXP 1, 3-5;
7YZCHRjDh9wQMFy6 1-4; 92qYgyHrGKjfx74n 1; 9C2qkgtRP8AWTJav 1-4; 9hT126dqjwyx3g7r 1-5; 9pT4hkMKYaxcQCz6 2-5;
9x6B1wRDnVrZ3Kh7 1-4; 9zy7hfdYrKkPxvXV 10-52; A7wDCjF89JmfNkZa 1; BHvgJVTdZ9pfNPa6 1-3; C2M3X9bK7WYp4cgj 1;
CN68qfBaY3MLXkGP 1-2; FHqYCNcKntvDAjxg 34-36; Gkb3rKHF6V9gz4pR 1-5; H2dZ3f1rTC4JyntF 1; HkynYAjBCvN19rmh 3-4;
HwyJMCLKfbY1PkmG 5-8, 10-38; JRQqNL6DkvWKYdPB 1-7, 9; N4CjrdPD6FhKtaY2 1-9; QFCv6ZA4LjkNJdaq 1, 3; QN7nvZAP6Da8kGBF 1-2;
QZ3hRyAMbDrj9cHz 14; RBpWJ7zM9N3TDGQc 1; RJXQC67jbzATWdx1 1-2, 5-9, 11-16; RdaAqHPcWnNZ2Vhm 1-7, 9-10, 12-21;
TB6KHhwxyPnDQctg 5-13; Tmbn84VwC13A7kDr 8-14, 16-38; VK83G7MdBhLRFAfa 1-42; VKfYJ82cMTktxQzL 1-39; VYBjgJx8aZtRhmpy
38-46; VmTj1LcnghRpxrMy 1, 3; WfKv2TphtRrzNA3d 5-10; XWbYtZ6mTHND2G3P 1-3; Yt9VaKGvf6Ry3BMW 1-5; cBGdRD4xLq3fW9FZ 1;
dnDg7j9kRw6WtbML 1-21, 23-40, 42, 44-52, 54-61, 63-69; f6HGnWcxNgk7MZJa 1; g79pDyr6TLWzVmhX 1-2; jTmX9Dg4FCNWyd7q 1-6;
jgcf8Q94JHB1wkWn 1-2; kG98jRqpFfBryv3c 2-4; kNwn6HXKVhCByZr3 1-2; m4dCL16tqnN2D3hW 4; m76CKntQZWLYhTMg 1-6;
qw4YQ9kTJ3DxAgNt 42; v8Zp2QF7An4d9xMj 1-10; vjN9b4n8dCRz2YTp 1; xXWkB3fqAjt9NJGH 1-5; zFRnjkryLtYWVKqJ 1;
zfdVHZ26jRm473cn 1-2.

## Battlegear of Valor (set 189)

- Proc: 5-piece set bonus proc 450587, logged as 450589; rolls on melee swings
- Sim rate: 1 PPM, observed; measured 0.88 PPM [0.74, 1.00], 9 wearers, 9 runs, 201 procs on 5895 hits
- Reports from 2026-11-14 to 2026-11-16, game build 1.60.1.70938
- Counted apart from the sweep, on the same pricing: the hits are landed white swings of either hand and Windfury Totem
  hits, each on its hand's weapon base speed; a dual wielder's swings are split between the hands by damage (Otsu),
  and the interval is 4,000 bootstrap draws of whole wearers
- The Battlegear of Heroism reissue's 4-piece bonus is the same row; no wearer with it active was measured, and the sim
  gives it this rate

Landed white swings and Windfury Totem hits roll it, the Windfury hit as a melee swing (§Windfury Windfury hit as a
swing). Of the Windfury hits that
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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18138
- Rate: 1 PPM, measured

429 procs on 9,373 rolling hits, 13 wearers, 15 runs, in the main hand: all hits 0.98 [0.92, 1.04]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.04 [0.84, 1.20], 110 procs on 2,267 lone units, 13
wearers. Specials: 0.00, 0 procs on 1 lone unit, 1 wearer. Travel time: the proc lands a median 163 ms after its hit,
98th percentile 270 ms, window [-20, 280] ms (100 isolated hits). Dropping the swapped stretches (5.5 active minutes, 1%
of hits, in 1 run) moved the read from 0.97 to 0.98. Left out: 19 of 34 runs in Cat or Bear Form. Reports from
2026-11-10 to 2026-11-21.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |      12 |   414 | 9,092 | 0.98 |
| Druid  |       1 |    15 |   281 | 1.14 |

Reports and fights: 4bqYVtkjL6aKdwQM 1-2; 92YFDCN4AvBT7KqJ 1-3; 9RgXYmBfN7hCdznQ 1-3; Chcj6QkYyadWfT18 1-4, 6;
DJ1PR6j7LNAbm3yG 1; FJTtGa1hpZNncr2j 1-2; H7cZXGkY4xzMmL3W 1-2; HpftBG4KPCZrMxbF 1-3, 5-9, 11; JZPQ1mKC8MrHfVzF 1-4;
NwHAYtWgyXf7DRCp 1-14; TfZDt129xGWp4vFC 1; Y61cQXLJwHfRabN4 1-31; a6QpkvcR1YHLBqPh 4-6; mNPzBQtDyL1KHTpZ 1, 3;
waZyxVKk43BzGLN6 1-11.

## Black Malice (3194)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1292155
- Rate: 1.85 PPM, observed

340 procs on 3,323 rolling hits, 7 wearers, 8 runs, in the main hand: all hits 1.86 [1.70, 1.97]. Its lone-unit rates of
50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit rate
checks it. One-hand white: 2.25 [1.80, 2.70], 127 procs on 1,024 lone units, 6 wearers. Specials: 2.49 [1.73, 3.46], 35
procs on 256 lone units, 2 wearers (a Poisson interval: under 5 wearers). Travel time: the proc lands a median 167 ms
after its hit, 98th percentile 633 ms, window [-20, 643] ms (130 isolated hits). Dropping the swapped stretches (5.1
active minutes, 2% of hits, in 1 run) moved the read from 1.83 to 1.86. The paladin check, which sets no rate: lone
white swings 1.65, 16 procs on 176 lone units, 5 wearers. Left out: 3 of 11 runs in Cat or Bear Form. Reports from
2026-11-10 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       2 |   197 | 1,865 | 1.92 |
| Shaman  |       5 |   143 | 1,458 | 1.78 |

Reports and fights: 14KrcZhmb2WkGQLH 1, 3-7; 4WrfjAqJYhxGHwPd 1-3; 7APg28BmdtzHhQFV 1-6; f7LKGpNkHDF4qc2T 1-2;
rdhLY1mZjgwBW4TK 1-4; tAHYCaLM743mRJrc 1-3; tcJjKwdCFXH1nkaR 1-12; xNtYHKbFwVGD93BW 1-2.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18088
- Rate: 2 PPM, measured

3,563 procs on 50,513 rolling hits, 28 wearers, 31 runs, in the main hand: all hits 2.02 [1.93, 2.09]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.83 [1.60, 2.09], 289 procs on 4,516 lone
units, 8 wearers. Dual-wield white, which sets no rate: 1.82 [1.67, 1.97], 584 procs on 9,152 lone units, 18 wearers.
Specials: 1.91 [1.79, 2.03], 681 procs on 10,206 lone units, 26 wearers; Sinister Strike gives 306 procs on 4,530 of
them. Two-roll abilities, priced as one roll: 2.78, 23 procs on 236 lone units (Execute 21 on 199 and Mongoose Bite 2 on
37), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,186 isolated hits.
Dropping the swapped stretches (196 active minutes, 11% of hits, in 6 runs) moved the read from 1.79 to 2.02. The
paladin check, which sets no rate: lone white swings 2.40, 36 procs on 429 lone units, 6 wearers. Reports from
2026-11-12 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      13 | 2,035 | 27,671 | 2.10 |
| Warrior |      12 | 1,280 | 19,326 | 1.89 |
| Hunter  |       3 |   248 |  3,516 | 2.02 |

Reports and fights: 2BD91CaTbhmY8HNG 1; 34gy17FXnaGhVC2M 2; 6CGNvtpXkVn3bxFT 1; 6kN78G2xfqy3vwTP 6; 7M3gGcH6bprxKqZY
1-3, 5-8; ATMmnV2tC3ZPb6G8 1, 3; BPVxqXWLc3FJTpMh 7-10, 12-14; GmkC6RtH12vawYTL 1-5; HPJY87k2VhMrKwvT 1;
LFzpja9rZBXCxkNw 1; MHL1xhpwC4dcVDZK 1; N2m8nPTGJQrj1DhH 4-5; Ncp4bVY3x1DXRqvA 1; TDgxaVBAWQ96HdXv 1; VPG6FyAdJ7NcC29X
1-5; VwyW3ZP1H6RTpdmf 1-5; YjwcQ2f4xNXvGd81 1; ZD3cfJpHa4vGYjKF 1; ZfRYBTWb168N7mzQ 2; bGRfPhMWCrKg9ydV 2-3;
c69WN31xbHBPtmhv 5; cbkDQfXNPGj3tnWT 1; dwcgGfFmAXWK7Mjz 1-2; gZ4JNBfHzL6C8cPK 1-3, 5-8; hDGL1wmpV4WjxbHP 1-3;
hkADNqV7JWM19df4 1-8; jJYHChwVBRaDxNWk 16-32; jXTNmWCvcqzBpbYQ 3-4; kz4FVdBRqpJwj8nx 1-3; mApW49VaFD2TKyg1 1-2;
x1QpJTH2aDZLnq9h 1-6.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18200
- Rate: 0.9 PPM, measured

473 procs on 11,377 rolling hits, 19 wearers, 21 runs, in the main hand: all hits 0.92 [0.84, 1.01]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 0.93 [0.84, 1.05], 144 procs on 3,434 lone units, 17
wearers. Specials: 0.99 [0.93, 1.68], 25 procs on 561 lone units, 5 wearers. Two-roll abilities, priced as one roll:
0.82, 3 procs on 81 lone units (Execute 0 on 40 and Mongoose Bite 3 on 41), left out of the rate. Direct: window [-11,
100] ms, median 0 ms, 98th percentile 0 ms, 156 isolated hits. Dropping the swapped stretches (15 active minutes, 3% of
hits, in 2 runs) moved the read from 0.90 to 0.92. The paladin check, which sets no rate: lone white swings 1.17, 46
procs on 873 lone units, 9 wearers. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      15 |   383 | 8,856 | 0.96 |
| Hunter  |       1 |    49 | 1,236 | 0.88 |
| Warrior |       3 |    41 | 1,285 | 0.71 |

Reports and fights: 1YQTjgZh2KPBGCbN 1-3; 2ZYxCWcLwVJar1Ay 1; 2ywhMnJq4Y3zaWtg 1-9; ARXJvZh6kwt3np2D 1-9;
BNajn8zgV49rZ7Fm 1-18; FVp1fTnc2Y7wXPxK 1; Fx34Wc78qhnPmVz6 1-5; GFkKCvMYJwfjHxac 1-6; LxdYWR47ZFbTNBhM 1-2;
LzFXxt3Y1HyVK4nc 1-7; YFnCf6cWNRpgVMHT 17-31; ZMn6BHdqtvpGLhcK 1; bqW1xQpFc9ag7NtM 1-2; dm4taDwfNpHZgkVC 1;
fAKd8D7wkgQzM9GZ 1-3; kma6pFnDVW1CQ7Tv 1-5; tCafRNMpHyAbgcmz 1-5; wPD796Jv2htngRmT 1; wdrYBPyVFcxMX7Q4 1-2;
z3nAQTFjM4rg8XmH 1; zknXTJ6t3q4aGYyj 1-2, 4-5.

## Bonechill Hammer (14487)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18276
- Rate: 1.95 PPM, measured

2,353 procs on 30,221 rolling hits, 33 wearers, 37 runs, in the main hand or the off hand: all hits 1.95 [1.86, 2.03].
The rate is all hits, which the specials rate agrees with. One-hand white: 1.86 [1.37, 2.47], 48 procs on 645 lone
units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.60 [1.39, 1.77], 242
procs on 3,792 lone units, 27 wearers. Specials: 1.69 [1.43, 1.99], 159 procs on 2,349 lone units, 20 wearers; Sinister
Strike gives 86 procs on 1,283 of them. Two-roll abilities, priced as one roll: 3.91, 5 procs on 32 lone units (Execute
5 on 32), left out of the rate. Travel time: the proc lands a median 142 ms after its hit, 98th percentile 300 ms,
window [-20, 310] ms (331 isolated hits). Dropping the swapped stretches (54 active minutes, 5% of hits, in 5 runs)
moved the read from 1.85 to 1.95. Report Z4VAmgDfJ9GPFjtc logs the same run as KNtHCAPd97ZMTLBf and counts once. Reports
from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      16 | 1,339 | 17,804 | 1.88 |
| Warrior |      17 | 1,014 | 12,417 | 2.04 |

Reports and fights: 4jnhaMNdZczgvFBT 1; 6XwgYbxq2GycCFP1 3; 7Lk3abCgNpVZjM41 2; 8Qp4nrZNKaFzy79G 107-109;
9zxHaWbYT3gqDQvF 17-18; AmWBdq2xcHLv3aYt 1-3; BXDNYVpvMtACK93c 1-9, 11-12; BkTfV9yAZJX8P7wF 1; FHwByWtYngL3dDVp 1-71;
FdHnTy2wKqVbtvcj 1, 3-13; HkKrmn81pNzyj34D 1-12; JkcfCt4zVLn8badj 1-8; KNtHCAPd97ZMTLBf 1-2; Mkz7QrJmvA6ZN4qK 3-24,
26-33, 35; Mtqz1p49BfvnGyAd 8-44; P1BkzXfMtxDAYcra 1-2; TGtNZnyXfLzjqw7b 1; TJxaKCntRkGX3Dj2 1-4; Tjw2g6pHbWymaz7q 1;
XWbYtZ6mTHND2G3P 1-3; Y1qVpGRtMAxy4Pvk 3-4; Z3MBhbDTXJ6tPCdQ 1-4; Z4VAmgDfJ9GPFjtc 5-6; b2cX8fQ9GmMYkLwy 1-3, 5-9,
11-20; fMLcHzyZvt7QTCmg 1-9; j1dmtfcYpLNTgKrD 8-9; jAqXz8PN92ZbG3nc 1-3; kDWLVzmjQa3CAnxq 1-4; m48QxrDLwtqFPjJR 2;
rJmGPzj4QNTq1hg8 1-2; rQ2bTtw8fHyRNGj4 1; tF7GdWVJ4qh89wgT 1-2; vTW2MZKhFA9GcyaV 1-3; vYjL3q712CyTGkQh 15;
xbg9BypFcdVDwqMk 43-76; xc9YMbqVJmGgTvfD 1-4; yZvwm7pRYDkAPGL4 1-3; zmDY4LVX7N8vh9aR 1-3.

## Boneslicer (273819)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1292573
- Rate: 2.95 PPM, measured

529 procs on 5,344 rolling hits, 28 wearers, 35 runs, in the main hand: all hits 2.97 [2.80, 3.17]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 3.07 [2.80, 3.41], 279 procs on 2,727 lone
units, 28 wearers. Specials: 3.04 [2.56, 3.82], 104 procs on 1,026 lone units, 24 wearers; Raptor Strike gives 81 procs
on 816 of them. Two-roll abilities, priced as one roll: 4.29, 13 procs on 91 lone units (Execute 1 on 25 and Mongoose
Bite 12 on 66), left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 361 isolated
hits. The paladin check, which sets no rate: lone white swings 3.36, 41 procs on 366 lone units, 4 wearers. Left out: 53
of 88 runs in Cat or Bear Form. Reports from 2026-11-10 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |      26 |   402 | 4,061 | 2.97 |
| Warrior |       2 |   127 | 1,283 | 2.97 |

Reports and fights: 2x3qyFaGMJ1nHNhf 1-3; 3rbcnjTMRfwxv1mh 1-10; 3yMkaTdQx8NJXGKb 1, 3-12; 41mPNTRvncCKbjYz 1-2;
79RVgZWM8dF2wkCp 1-6, 8; 7zcat1AyFnrJ4BQR 10-19; 8fwQv1dbWXxpkG3r 1; BzGXfjMLJ2RyCwHm 1-7; FAf9V76BctbG3mR4 1;
JLGynNgvkBHPfd9w 1; K9WGdrx248FvaDVY 1-3; KbNyBzwZg1t9jMcC 1; KrH8VYbgyfaTXWh9 6-8; MW8HLRKDFN4GxTr7 3-4;
NY6JnFrwvyHhG1Pp 1-2; NrHbC9xfM7KzZt4d 1; RQpgbjTLn6wd8z2J 1, 3-4; TZkhGPt2fqAnNY74 1-3; Z26ztpb3FkJhgTAf 1-8;
aqCndvgQbBXL3Jf8 9-13; cG82vnXYjAVaTN1z 1-4; g74R1Wf9FL83pMP2 1; gJkN6DCr8ZfWTn72 10-13; hGNVDdLFRkj7tfcv 1-7;
hLNVaj9HdnCM2BYt 1; jHKfnckY8gbGTvyp 1-2; nKNgy3VrZXkcdjBT 4-5; pLh28qFZjAYvrMT6 1-6; qMKJYxCXzcPRZ8nD 1-3;
rpB2MKnbm7LxCqJZ 1; tVB2a6fMvYNCGdkH 1-5; vnwtmMTpkR1aYP7G 1-9; wy7HdTbGrvLXz9pt 1; y7qfkc91hA3DKQ4g 1-11;
z3YfqBgbt1GjFD4W 1-4.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293518
- Rate: 2 PPM, measured

723 procs on 7,823 rolling hits, 7 wearers, 7 runs, in the main hand or the off hand: all hits 1.98 [1.86, 2.10]. The
rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 2.33 [1.88, 2.86], 93 procs on
855 lone units, 1 wearer (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.56 [1.27, 1.91],
79 procs on 1,085 lone units, 5 wearers. Specials: 2.04 [1.67, 2.45], 110 procs on 1,158 lone units, 3 wearers (a
Poisson interval: under 5 wearers); Cleave gives 35 procs on 332 of them. Direct: window [-11, 100] ms, median 0 ms,
98th percentile 0 ms, 214 isolated hits. Dropping the swapped stretches (30 active minutes, 8% of hits, in 1 run) moved
the read from 1.82 to 1.98. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       4 |   475 | 5,254 | 1.94 |
| Rogue   |       3 |   248 | 2,569 | 2.07 |

Reports and fights: FLtpV1h8GTz6rgAc 2; LK46AC1ytPNQm9cD 1; cr6qMtj1vCTPHVfw 1-7; hF7dxZ8nJDpk4wNG 1, 3-5;
pjRYqGgB2rdvawNn 1-10; qcdyZAr2ganDtxNW 13-42; r7K3c94VNbyDAxLg 2.

## Burning War Axe (2299)

- Count: §Count of the 2026-11-09 to 2026-11-22 reports
- Proc: weapon proc 18199
- Rate: 0.95 PPM, measured

1,004 procs on 19,059 rolling hits, 36 wearers, 37 runs, in the main hand: all hits 0.96 [0.91, 1.00]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.09 [0.97, 1.21], 287 procs on
4,802 lone units, 34 wearers. Specials: 1.12 [0.95, 1.29], 76 procs on 1,229 lone units, 9 wearers; Rend landings give
49 procs on 747 of them. Two-roll abilities, priced as one roll: 2.00, 28 procs on 254 lone units (Execute 28 on 254),
left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 300 isolated hits. Dropping the
swapped stretches (65 active minutes, 6% of hits, in 4 runs) moved the read from 0.90 to 0.96. The paladin check, which
sets no rate: lone white swings 1.03, 79 procs on 1,399 lone units, 16 wearers. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      27 |   734 | 14,072 | 0.95 |
| Warrior |       8 |   270 |  4,981 | 0.99 |
| Hunter  |       1 |     0 |      6 | 0.00 |

Reports and fights: 3cKTmVyb1ZPwFXkf 1, 3-13; 3nhk8aFfrPqMATJ6 1; 93NcnZPAhCfQ4wHz 1-4; BqjTQP3Ghbt9JyYn 1-11;
FMcvKqTzn7m8Va3P 1-15; Fgy2bt81NL9CAzaR 1-13, 15; FkbzvXZadYtg6Nfp 2-7; GJHgk6MDpxfVdacv 1-23, 25-29, 32-33;
J6craWFxPjTAGv7t 1; KRCBMJvg72fHVwWL 1-17; Kdf1RavA6rZhk8LJ 1-13; L49QjcBKmtyVXkr6 1; LwTzGJXrN3HWqCfD 1-5;
MBaXr1H9GcPhDJLm 1-2; NtakrWH8xvRmzpCg 1; PZA4xq2WvQJ3amjH 1-5; Pc1RkCWp8wTNfFr2 3-4; QvKPgyHDkBMbcp7T 1-10, 12-21;
RBmtAMLxz78Pj2Wh 7-10; VQFaRpT6J2xLDbZg 1, 3-5; Wf8ghC6TtQKrkXdv 8-10, 12-15; Z7YLKv4Rfkq9cX6A 1; bkMdQTwZmqNPvhrR 1-5;
gFZ7BcKt9JhaY2dQ 1, 3-9; gGtkavy2n8bqADQ9 1-2; gJkrfBtbp7YM3PCd 6, 8-21; h2FAdwQfcJLHpzR3 1-16; j68nZp1gcrGChaRW 1-12;
kFMtgzB1cCvxJTfL 1-9; nDzw6MCg3jPxYyJN 1; qMD3CVapwb6nAgx2 16-40, 42-43; rMyzaJhG4k8wHPj6 1-10; rb7N9JZ6VHfjmknw 1-13;
ryqgQphLxMBz4TmH 1-4; t6HbjRZmkyx3pQwK 1-14; xGndPFNVX7Lg38Hj 1-4; zdyXxBwMPTQphqHa 1-2.

## Claw of Celebras (17738)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 21952
- Rate: 2.05 PPM, observed

2,123 procs on 34,556 rolling hits, 47 wearers, 52 runs, in the off hand: all hits 2.05 [1.93, 2.15]. Neither the
one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no
lone-unit rate checks it. One-hand white: none. Dual-wield white, which sets no rate: 1.87 [1.73, 2.01], 986 procs on
17,532 lone units, 47 wearers. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 772 isolated hits.
Dropping the swapped stretches (69 active minutes, 4% of hits, in 3 runs) moved the read from 1.97 to 2.05. Report
N7FQbLWcD3AZJkC2 logs the same run as Y8bALnZFcgK6R1X3 and counts once. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      15 | 1,064 | 16,293 | 2.18 |
| Rogue   |      26 |   977 | 16,929 | 1.92 |
| Hunter  |       6 |    82 |  1,334 | 2.05 |

Reports and fights: 14vzPR7JyKFLh6tb 1-5; 2taZz4vTdfXYAk3b 3-9; 4Qm6MgZWBTx78Lvp 1-2, 5-6; 4XGyazkdvmP98r2L 1-7;
4zyTBfqQpJ6MrCjv 1-2; 61Hx8DgwhdbWaN9C 1-12; 6j73x2mkhf18LJyH 1-7; 6yrCJtHYmaRnZBM3 1-5; 72gnLYcQ9ZkTwNqK 1-7;
7AxaTpFZMqGb9wNd 1-8; 7G1xqYWJa3A9fcgt 1-2; 7jwrWbDQ6C8VyBdp 1; 8TjkZQ9gB267bhaL 1-3; 9BNc8nbtQ2ZyWjkJ 6;
9ZTCYrFcDxQ8pk7H 1-7, 9, 11-13, 15-34; BGDd8mWpgRbwaCn2 5-8; CfwVjT24hFpMByDR 1; CnZmWjckYfLqdFDJ 1-2; GnPwRFpMLkd3tYca
1-6; LRH1kQK2J8frW6dt 1; LRvhYmZHFn96837g 1-9; LVdPJq3hKTFkBYNv 1; LzDT4PHN9rX8g7FA 5; N7FQbLWcD3AZJkC2 1-8;
PF763TBADCNg1YwV 1-4, 6-9; PTBj1azytM8RQrFX 1; QJpjNCgx8Kr7TfWD 1-16, 18; Rc4dCWA6N8PwpBvT 1-8, 10-11; RnYmQGbdzpfLyvVD
4-7, 9-13; Rvgx7kmXtdh83abr 1-6; Y8bALnZFcgK6R1X3 6-24, 26-27; YTn9F1x6jLqN2fdQ 1-5; ZJcLHaz2Q8V1XjNF 1-4;
aQAhMLzBbWJcdq6G 1-6; dXkatpwrWTRYP9hM 1-2; fPD87TvAkZ9XmpyW 1-11; gVwCaBW18Ry23dtj 1-3; gv68D7T4jwydx2Zn 1-8;
kJYq29CBFaLthXPR 1-7; kRB98ZcNgvDJj7at 1; kY28yDTZG6K9bFaJ 1-4; kY7VqD2CmyPMNAz3 1-4; mArH91f2PRavx4dt 5-7;
mHB2rNQ7aPh9kdWn 1-5; ntmraghfPTkdbqF6 1-14; pQPwFg8rf6VWhRNc 1-3; r16Zq7vmV3KDgdQR 1-12; rTJ2d6QwfypanWhv 1-4;
rmk2gCVqcp96aKtD 1-9, 11, 13, 15-16; vKGdpHx9P7agyhFj 1-11; wPJ21t6c4WvCrMaG 1, 3; x1gVkThZLdzG9a86 1; ydJnbZQkD2fNMhmt
1-14.

## Claw of the Shadowmancer (2912)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 16409
- Rate: 1.05 PPM, measured

1,413 procs on 43,498 rolling hits, 59 wearers, 67 runs, in the main hand or the off hand: all hits 1.03 [0.98, 1.08].
The rate is all hits, which the specials rate agrees with. One-hand white: none. Dual-wield white, which sets no rate:
0.98 [0.87, 1.09], 280 procs on 9,063 lone units, 53 wearers. Specials: 1.14 [0.81, 1.52], 68 procs on 1,882 lone units,
32 wearers; Sinister Strike gives 53 procs on 1,331 of them. Travel time: the proc lands a median 134 ms after its hit,
98th percentile 290 ms, window [-20, 300] ms (231 isolated hits). Dropping the swapped stretches (66 active minutes, 4%
of hits, in 8 runs) moved the read from 0.99 to 1.03. Report tWHjPvT9ZMfcAznY logs the same run as NkyKCr2XmgLt1YWv and
counts once. Reports from 2026-11-07 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      54 | 1,364 | 42,164 | 1.02 |
| Warrior |       1 |    20 |    416 | 1.52 |
| Shaman  |       1 |    15 |    357 | 1.33 |
| Hunter  |       3 |    14 |    561 | 0.79 |

Reports and fights: 1jgtdNmDMRycbZrP 1; 2cMhxrv31PBfy9Rz 1; 6G3dtKC2AwXD4fjM 1-3; 6jHPJBnLy1pF9Rkf 1; 76ZV8FQ1RHDhbq3z
1-4, 6-9; 8hWFBZNHYqAj4t9y 1-2; 8mM2R4ZDQfCxrbvh 1-2; 9LdQXKwv32Hf8pcz 1-10; 9LgNW1jQ2Zh8XDkY 1-12; 9nYmzB1HfDpbqkXR 1;
9vwmqxP4dLKhDC8Z 1-23; AHdpxG72rk6CzbMa 1, 3; AWfNm2BDqkpcCH7h 1; C71ArgnKwBqWf2Tk 1-13, 15-22; CTV9qxJNPyWcd6vA 1;
DrWqmBwPV9KMyLGZ 1-6; FR1W2Dgr4tcahZx9 1; FXdJPfWhc3wm2bvy 1-3, 5-7; FYmnawMf8gjB4NQt 1-3; FzRmWck7KgT9yt6D 1;
H3KAQPNjGC1xMdvY 1-2; HxPZVaBGLcKXg4bJ 1-2; JK4pARZQzyLb2DMh 1-3; KXmBtLjWnQ6zZa21 9-14; L8yaqgDXG26Nf9v1 1-2;
LbmB2QdWnPT7JGAF 6; N6bDPTdj3LaKhQZJ 1; NDvFdRrMVGZX2nCa 1-4, 6-13; NPrqRFH78JQyvgB1 1-13; NkyKCr2XmgLt1YWv 83-87;
PBjtDw1z8LXygxmR 1-5; Q3AaMx7GC9R1mXTp 1-2; RLnWB4FrJgpMyNZj 1-5; T1R3rhmW6BVqC9nQ 3-7; Tw4BdRP1HmZqJL9F 1-3;
Twr9aYQFMWjnLPqR 1; VapjftdQrYqCD4Ay 1-9; W1CdNAXcjaYwGhJP 1; WLN37gKYpGVXy4aq 3-9; WjY2BKR7khXtwqaG 1; XYDcxmK8A3at2wgN
1; YTjp1WwqH3gh9XDQ 1; YcVR36BKNkmgQLyb 4; aZdFD2AKRqt3ngPh 1-4; akfrAm74NYHVqKgP 1; cawCKTRvpA3LQ2JY 1-5;
cdBaV9ZJbLFhYnPp 1-6; cmaGQRC7PJnq2v6p 1-8; dM46X2j1BZGhYTmx 1-2; fGY9cXmQdyRKhpCj 1, 3-23; g1kdNFQh2c7DmAqt 1-11;
g3BPNGHVcYkAwx72 1; gxMP1YB28ApJDT4v 1; h7bAMpcfYrTjkm9z 1; hcGCaDW9tRYFB6Ax 12-18; jGVzyqWN2wgD9kcX 1; kn8a4mtMw6FcpXGg
2-8; nacNMLGj1X4YHdCV 1-6; p8FzkTmbqc2KDyW7 1-2; pYzrZAaCLfcM49vR 2-3; qh8cLxYGHF6bNXCQ 1; qy7j4dPf9aACJQcY 1;
tWHjPvT9ZMfcAznY 1-5; tvDFpcQ4w3L7PBy1 1-3, 5; wac3dD91PKCnqJt8 1-2; yL2Z3jpHdVqCarPF 1-2; yaWvCBQ2GZTkLM7j 1;
zafjtdQyBX2D6nmG 9-12.

## Cobalt Crusher (7730)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18204
- Rate: 2 PPM, observed

289 procs on 2,718 rolling hits, 6 wearers, 6 runs, in the main hand: all hits 1.99 [1.65, 2.28]. The rate is all hits,
which the one-hand white rate agrees with. Observed: every wearer is a shaman; the interval is wider than 10% of the
rate. One-hand white: 1.71 [1.38, 2.23], 60 procs on 656 lone units, 6 wearers. Travel time: the proc lands a median 73
ms after its hit, 98th percentile 191 ms, window [-20, 201] ms (50 isolated hits). Dropping the swapped stretches (6.1
active minutes, 5% of hits, in 1 run) moved the read from 1.89 to 1.99. The paladin check, which sets no rate: lone
white swings 2.11, 31 procs on 275 lone units, 9 wearers. Left out: 12 of 18 runs in Cat or Bear Form. Reports from
2026-11-14 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       6 |   289 | 2,718 | 1.99 |

Reports and fights: MyQ3KzBh2WVRnTPg 1-2; NaKpYVnWQ1LwrjCT 1-6; g8b1MkJnKtyfzvPa 4-5; nxHFpwZ8NvLP3r2z 1;
xy7rATKMdJkFhYaR 1, 3-14, 16; z7DdJBGj3T6YZQ8V 3-5.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18276
- Rate: 2 PPM, observed

1,394 procs on 27,908 rolling hits, 11 wearers, 18 runs, in the main hand or the off hand: all hits 2.00 [1.88, 2.07].
The rate is all hits; the one-hand white rate agrees with it and the specials rate reads below its interval. Observed:
the specials rate disagrees. One-hand white: 1.99 [1.57, 2.49], 76 procs on 1,526 lone units, 2 wearers (a Poisson
interval: under 5 wearers). Dual-wield white, which sets no rate: 1.82 [1.50, 2.13], 163 procs on 3,577 lone units, 10
wearers. Specials: 1.69 [0.81, 1.85], 60 procs on 1,416 lone units, 7 wearers; Cleave gives 39 procs on 984 of them.
Travel time: the proc lands a median 163 ms after its hit, 98th percentile 286 ms, window [-20, 296] ms (152 isolated
hits). Dropping the swapped stretches (43 active minutes, 4% of hits, in 5 runs) moved the read from 1.93 to 2.00.
Report qNVWc2yZJT6k8mvh logs the same run as JnVckZDLTwAgGKaQ and counts once. Reports from 2026-11-15 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |       9 | 1,004 | 19,962 | 2.01 |
| Warrior |       2 |   390 |  7,946 | 1.96 |

Reports and fights: 7fy93LdcwQhvF2gP 1-7; Ajg8VwL7CQFm6DWY 1, 3-12; C18pcNtwj4aRDbQk 3-15; HwyJMCLKfbY1PkmG 7-8, 10-38;
JChQqKVgLWykANRr 9-46; JZDFdhkvn3NP6xLm 1-7; JnVckZDLTwAgGKaQ 3-35; Qr9PmCnptRWMjN7k 29; RFGVCxDvb8mf2dPT 1-9;
RJXQC67jbzATWdx1 8-9, 11-16; XGBaYAqPbcgZCVk8 7-15; Yh8waABRynm9HfFk 1, 3-4; adzWfghGTA3pLnJX 1-2; mCn63gBZfLFHaWr9 6,
8-34, 36-39, 41-43, 46-53, 55-89; qNVWc2yZJT6k8mvh 1-2; r3dNjhavPz8LFVWY 1; t3RKbzMD8HY6pcfX 2-3; xCX8YjaRrvKq6tfg 4-21,
23-24; xK2PWFfnM9GwaC4A 6.

## Dazzling Longsword (869)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13752
- Rate: 2.75 PPM, measured

7,473 procs on 62,164 rolling hits, 58 wearers, 60 runs, in the main hand: all hits 2.77 [2.66, 2.88]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 2.70 [2.41, 2.96], 419 procs on 3,582 lone
units, 12 wearers. Dual-wield white, which sets no rate: 2.29 [2.09, 2.49], 940 procs on 9,492 lone units, 44 wearers.
Specials: 2.91 [2.78, 3.05], 1,581 procs on 12,525 lone units, 58 wearers; Sinister Strike gives 755 procs on 5,931 of
them. Two-roll abilities, priced as one roll: 3.66, 52 procs on 328 lone units (Execute 43 on 253 and Mongoose Bite 9 on
75), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 2,278 isolated hits.
Dropping the swapped stretches (23 active minutes, 2% of hits, in 1 run) moved the read from 2.73 to 2.77. The paladin
check, which sets no rate: lone white swings 2.49, 58 procs on 538 lone units, 9 wearers. Reports from 2026-11-09 to
2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      34 | 4,584 | 36,433 | 2.90 |
| Warrior |      22 | 2,655 | 23,441 | 2.61 |
| Hunter  |       2 |   234 |  2,290 | 2.36 |

Reports and fights: 1FV3WyfXGCcPrT9b 1, 3-8; 2DYRfGAgj3prvF7K 1-5; 3rxWD9jV1t7TgLJp 1; 4C28YPXqgTMQJxB7 5-9;
4qZXFGfc1wCMQKL7 1-8; 8F4MhkmNBxRrf6Zw 55-68, 71-83, 85-89, 92, 94-105, 108-123, 125-141, 143-151; 8X2Kc46faTnxVBLm 1-3;
8YQCGy97hpdx2Rk1 2; 9aXFMvhYPDbtCkp7 1; 9cajk6Jqd3NTQngf 6-8; A1FKkYpPxrX8tbNv 1-2, 4-5; BCZgcRxrh7dA4ayX 1-4;
BgKt3Qm2fJDAWLzh 1-23; CrR9c6VtXG4bzJhx 1-17; D4A3QNCrX19W7cqJ 1-10; D6jptzw9WnQxmyfq 2-3; FR9jJ7b31XCkLmtQ 5-7;
FmdgnGJ3YDPWr8RL 1; FwXMRQPHBmZcLWCa 1-9; FzRmWck7KgT9yt6D 1; GtmK2BRQ9PVkXgyv 1-3; GzZH1WhTLQ3JaDXt 1-11;
JTztbhZyaWHC6kxf 1-5; KVdX1Wn4GpzAH9MN 11-12; Lg82ZFDyQ4MvRm9p 1-2, 5-11, 13-26; M4CvXLZz3qm6pDRc 1-3, 5;
MVcaxDQwWLpnK6HJ 7-9, 11-12; PKRTfgw2Q7ZBxt4r 1-26; Q7BxMTcG1ZbgX2v8 1-2; QX2xnHapd6ZPhVzb 1; QftYMR2haLcXZ7Nz 1;
RacWqB34tLQT8P16 1-10; T71XnfaF3cY9M2Rg 1-13; THqFxLkCP6ayd2tZ 5, 7-9, 11; TV87BvKgRWGrPtx1 1-14; Thf9agqHJyRmtdpD 1-5;
VLmAcpbQxWf1CtaR 1-17; XCJdxAF17zqHMhLY 1-2; YXaWgCkPGAQn3R6N 1, 3-9; ZtndFBMraKvcLfw9 7-9; bf31yTwQmGt6gdP2 1-16;
dmKgVGCJMQpq3t17 1-14; dzXT7JngpxRPwG32 1-17; fCrvDxhZXqJAg62m 4-10; fkVmJwWZjMy3LBX4 1-3; gTPNZcJ9mG4dRC2X 1-4;
gkvTYj8Lz3dDQqrR 1-3; h6wkBb9Z7jFtTrnp 1; nF7RcakdVTQD2tvH 1-3; nLk2TKDJWP6vBgam 1; pKj3bDrf1h2qtNMF 9-15;
pMGyN9hFa8XRAHxV 1-5; qW3KvYDpw6ARxZbc 1-7; tkQKL9bGT6JVR3na 14-16; w6H9xLRJQCWfBK1a 1-2; wqJpvBhgTnyNf2cH 1-5, 7-17;
xM9bRGyajmBFvCqt 1; xqy2GfPj67XCwrT3 3-4; yrmQHh3CdM968NqX 1-6.

## Deadhead Blade (274271)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293518
- Rate: 1.9 PPM, observed

1,181 procs on 20,596 rolling hits, 18 wearers, 26 runs, in the main hand or the off hand: all hits 1.91 [1.77, 2.01].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white, which sets no rate: 1.65 [1.48, 1.80],
327 procs on 6,587 lone units, 16 wearers. Specials: 2.13 [1.73, 2.99], 32 procs on 500 lone units, 5 wearers. Direct:
window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 232 isolated hits. Reports from 2026-11-14 to 2026-11-22.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Rogue  |      11 | 1,098 | 18,690 | 1.96 |
| Hunter |       7 |    83 |  1,906 | 1.45 |

Reports and fights: 4VtCK6Qqr1MD2hpc 1, 4; 6DvQ2NMGgdj3cHk8 2; 8Gw7m39ZXFz1y2fv 1-32; JLt4PFa2qTYKh3XZ 1-2;
KVLgwAr9bB1dJyNR 1-2; NRAXymcxJfPCnbFK 23; Q8Vgbpw4dJrzNx29 1-4; QRCgnhck3KVq6vBT 1-2; QjwpA7DPVfCZKkMy 1-42;
VGhX29mL6ra38qd7 1-3; VLz78BfZyWa3Cxr2 1-9; WLGy4qAtFxmcrfw1 1; aNh3Dt8kbZHL9cBM 1; bGhg8M19xzB46QLY 1; ch87CPWVqaxZYfz3
1, 4; hMk4XPnp2wAT81bZ 7-8; mRjh2L3CyYzWvcPg 86 fights from 3 to 98; mxh9rJfdbvPkRtTX 1-2; nX1qbatTzGZ3CAwB 1, 3-7;
pQ48t7ArW9aDdwNn 1-4; tk3BD1bZafMgKxWX 1-4; wp8WXkj6nL7KTZ9d 1; xMLV8adPBDHYGF27 1-3; yFGXLWVAKxw61gDP 1-4;
zabF8pwjDWtvmn7J 1; zp9yVrtfTNZJCv7g 9-10.

## Dwarven Hand Cannon (2099)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 29639
- Rate: 4 PPM, observed

7,170 procs on 36,937 rolling hits, 43 wearers, 48 runs, in the ranged slot: all hits 4.02 [3.93, 4.10]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it; every wearer is a hunter. Lone Auto Shots: 3.92 [3.71, 4.10], 1,226 procs on 6,476 lone units, 42
wearers. Specials: 3.89 [3.66, 4.09], 1,172 procs on 6,238 lone units, 43 wearers; Multi-Shot gives 887 procs on 4,648
of them. Travel time: the proc lands a median 578 ms after its hit, 98th percentile 1032 ms, window [-20, 1042] ms
(1,577 isolated hits). Reports from 2026-11-13 to 2026-11-22.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Hunter |      43 | 7,170 | 36,937 | 4.02 |

Reports and fights: 2tWx1VCrv4PcgAHL 1-2; 3HaPg2Yktxrnb7NL 6-9; 4W6d8VFkgmCrqGvc 1-2; 7tbWfYqJFd4vzL3p 3-6;
91wXZG6pN8CDanmR 61-62; BXLbtpMR6j7fYyKq 2-3, 5-6; BfVjHYvWdKkP8Jz7 1-2; Ddx2RWnArCKf6mqP 46; DzGaRHQ89xL3Yfdk 1-2;
F3BTZGrW48Mp6n2L 3-39; FLtpV1h8GTz6rgAc 2; GJCpqnN4bVBhtrgQ 1-2; GpyT8hfkVXmWPvnA 1; HYmf1vTW6hVDPytK 6-9, 11-16;
JZDCKavqxz47AbYN 1-2; KPqgbZxhn6BXwCvJ 1; M1n84TjqH3admxtc 1; RJXQC67jbzATWdx1 1-2, 5-9, 11-16; T6VDdFhPxQ43pHM2 1-2;
TGtNZnyXfLzjqw7b 1; VMGwjvJD8cqNfPnx 1-2; VxPK6Nam2XwzL81y 1-5; W1P3YqBptbnGM7da 5; Y1qVpGRtMAxy4Pvk 3-4;
ZhbCF1LvkQ9gBAMy 13-38; c4C7HaT2WdfFnJNx 29-30; ch87CPWVqaxZYfz3 1, 4; fMPL97aCA1ngvKQk 8, 10-20; gXhQJLaTxK94WfHr 1;
gY7zpjW9cMk6nGK4 2; gvzCK9NqHhkBTwPX 1-2; hH4YRGVFz3KnN6bw 1-8; hKvX9x2Mr76gCmbp 1-3; hgv7XY4tJKbTWPzH 1-7, 9-16;
hn82jpvC4PM7wZY3 1-5; jkcbfRxrPgvWqwyM 1-13; kCf4QFqbVBTjmvrc 1; nWwyhgfZkJdpVxF3 1; nmd7QgxNGjDpcFhT 1-3;
pHLRfhraA4GXmvN2 2; pMvyNHtjKnkgR3X9 1-3; px3rGD8JbH7YTdLV 1-3; tGBNadZg7VYMnXKA 2; tNkVywdKhDc7jAZf 1-2;
tncbVBRhz7X8FP9C 1-2; vJF78jKHQaxmnDr6 1-3; zxLCmJbZA2pQhyHR 1-2, 4; zxtbh1BmNVC8W2FH 20-21.

## Eater of the Dead (10805)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1299869
- Rate: 1.8 PPM, measured

3,722 procs on 56,222 rolling hits, 40 wearers, 47 runs, in the main hand: all hits 1.81 [1.73, 1.89]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.80 [1.57, 1.99], 646 procs on 9,785 lone
units, 17 wearers. Dual-wield white, which sets no rate: 1.66 [1.47, 1.91], 275 procs on 4,529 lone units, 18 wearers.
Specials: 1.80 [1.61, 1.99], 337 procs on 5,104 lone units, 36 wearers; Revenge gives 150 procs on 2,293 of them.
Two-roll abilities, priced as one roll: 2.58, 45 procs on 475 lone units (Execute 45 on 448 and Mongoose Bite 0 on 27),
left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 954 isolated hits. Dropping the
swapped stretches (50 active minutes, 3% of hits, in 2 runs) moved the read from 1.76 to 1.81. The paladin check, which
sets no rate: lone white swings 2.85, 23 procs on 220 lone units, 5 wearers. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      23 | 3,339 | 50,426 | 1.81 |
| Rogue   |       4 |   258 |  3,973 | 1.77 |
| Hunter  |      13 |   125 |  1,823 | 1.87 |

Reports and fights: 1PcrGTyatDAFx8Yk 1-2; 1WBQ78fbdLVhPF4j 1-2; 23CWz6vwhH9YdxN7 1-2; 2tCNaTJf49KWyVLk 16;
3H7tGVbMTnacBdjz 1; 3f7wRc8vmxMHFa9n 1-3; 3hpaqBV4jbGdx2TX 12-13; 6fqDJhcVRap21w8W 3-10; 8yPRmZdrq4QhcGn1 1-4;
APvwzrLaGqQ72f1C 1-2; Af6Ghw38QdWpnCXT 1-2; C834MAj1cKh2RZyd 1, 3-7; CJqgRYWTnm3jHGxP 1-2; CbRP2KHXfTjmYxnL 1-2;
CmZh47GDdxbMPRqg 6; D7rxQ2dpGnXVPwa8 1; MKYHka4RgWp63BNc 2-11; NdbCqzpGA8Yr9ZfL 1, 3-6; Q8MGr7WxPzRhYm4q 1-4;
RHgby76wQXp2nrxj 1-4; RJ82yLB7HcYrd39M 2; V8F7qLQXzpPAmcyG 1-4; VYBjgJx8aZtRhmpy 10-46; WgQrVv2BTbPGHKFD 1;
WrBH1KZxJvVP8hQk 1-2; XGBaYAqPbcgZCVk8 15; YXTfd7QPJV4Nqhcn 1-18; ZRaHBDLTyqVCr7fk 1-2; aVDjBvL8hyfCGmtq 1-2;
af6xmcjndRMkZ83N 1; bDcBjXw82GvhYW4F 1; bvm6fx1tRHAD29Ch 1-2; cwgvKaxBbA4YWnVX 1, 3-21; g7ZWhALb3rT6ctfq 1-19;
jLyXfQ2Ftp8RJmNd 1; jTmX9Dg4FCNWyd7q 1-6; mJDNgqYVGwb6A8tF 1-17, 19, 21-22; mdXkc6xwP8Y9pR47 1; nc1zq8VBvZCTJtjX 1;
pYATGQ6aCmzwBNcM 1-8, 10-13, 15-25; qQwFaXyTdhmnr2kP 1; tBVKZzdATj8kx3fF 1; vGdq3ZrKyMAPaTJD 5-8; wk2zgGAXYTbnHQ9V 1, 3;
xdb8D9cpKmw2CA1G 2-4; ya61L9CjgmMWAQTJ 1; ybRtJnMWBfFDHakx 1-2.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 17331
- Rate: 1.05 PPM, observed

816 procs on 29,368 rolling hits, 33 wearers, 36 runs, in the main hand or the off hand: all hits 1.04 [0.98, 1.11].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 0.91 [0.41, 1.72], 9 procs on 372 lone units, 1 wearer (a Poisson
interval: under 5 wearers). Dual-wield white, which sets no rate: 0.93 [0.80, 1.05], 218 procs on 8,829 lone units, 32
wearers. Specials: 1.09 [0.45, 1.79], 5 procs on 172 lone units, 6 wearers. Travel time: the proc lands a median 87 ms
after its hit, 98th percentile 246 ms, window [-20, 256] ms (167 isolated hits). Dropping the swapped stretches (62
active minutes, 5% of hits, in 5 runs) moved the read from 0.99 to 1.04. Report DdKN3mXj6gH9yvpc logs the same run as
D46Xm8ncR9WbTFrL and counts once. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      30 |   770 | 27,664 | 1.04 |
| Warrior |       1 |    26 |    916 | 1.06 |
| Hunter  |       2 |    20 |    788 | 0.95 |

Reports and fights: 32MRYw6xgkGXrymv 2; 6hDJQWTXwBr419AM 1-4; AaKzvcY1wmH9pBbq 1; B8JHDYfAhtCRbFLc 1-2; BHJMb9CGa1TgjtAd
1; D46Xm8ncR9WbTFrL 1-31; DCAQBK7P3pbxYRyf 1-4; DdKN3mXj6gH9yvpc 1-31; DfTMrcFAzq9ayQKt 3; G7X1CkT8z6PrfjWM 144 fights
from 695 to 871; GQbTNkadrgxwyVMF 1; JcmNfbzKkq4prQg3 1-2; KNTYZJ3Xg9d8HftG 1; Kphk63jtr1wb9NyL 1; L1rD4aNkQdxzcPyT 4;
PrcWByj746kYq93v 3; QXFd74zDW8JZMmKx 12-13; Qban6jAYVkFHP3fW 4; RxNdCcDaG1AyF63M 42, 44-45; Tm42j9AJcV6hrbC8 10-11;
Tx7MPH9f8Ga6hdrF 1-20; VkW6CGtgqT9r2vxN 5-31, 33-41; XNQzBvfdcYb1CP3R 1-3; XPYV9J6DfrMKxGqh 5-6; XgQYbwGBZcHLKRxV 4;
aRLP2JdQntKzTx3f 1-2; c8mjK4TdzV2M7XLF 1-5, 7-46; h3P9VjGMZCJWfTcm 7-10; k1taHbwL76Cnxv4m 4-10; kaH1G4ztmYpgMTJ6 1-2;
m4dCL16tqnN2D3hW 1-4; vApRC4TdybkBDMG6 1-2; wBraTYbxCNgnpq4f 20; wKxWgZz6y7dqkRVD 3; xXWkB3fqAjt9NJGH 5;
y6D3jqgWt8TvpNYM 1; yMPtxfd1h84c7ART 5.

## Felstriker (12590)

- Count: §Count of the 2026-11-13 to 2026-11-26 reports
- Proc: weapon proc 16551
- Rate: 0.5 PPM, measured

3,197 procs on 217,803 rolling hits, 95 wearers, 153 runs, in the main hand or the off hand: all hits 0.52 [0.50, 0.53].
The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 0.51 [0.45, 0.57], 166
procs on 11,548 lone units, 18 wearers. Dual-wield white, which sets no rate: 0.46 [0.42, 0.49], 779 procs on 60,374
lone units, 78 wearers. Specials: 0.52 [0.43, 0.62], 128 procs on 8,703 lone units, 58 wearers. Two-roll abilities,
priced as one roll: 0.75, 1 proc on 47 lone units (Execute), left out of the rate. No paladin wears it in the count.
Direct: window [-20, 100] ms, median 0 ms, 98th percentile 0 ms, 678 isolated hits. Dropping the swapped stretches (220
active minutes, 4% of hits, in 19 runs) moved the read from 0.50 to 0.52. 2 runs logged in two reports count once.
Reports from 2026-11-15 to 2026-11-26.

Across the 2026-11-23 hotfix, which sets no rate: the reports before 2026-11-24 01:00 UTC read 0.52 [0.50, 0.54], 2,780
procs, 84 wearers; the reports from then on 0.49 [0.44, 0.55], 417 procs, 14 wearers. The analyzer run on the reports
from then on alone reads all hits 0.48 [0.43, 0.54], 417 procs, 14 wearers.

| Class   | Wearers | Procs |    Hits |  PPM |
| ------- | ------: | ----: | ------: | ---: |
| Rogue   |      57 | 1,831 | 124,443 | 0.52 |
| Warrior |      35 | 1,344 |  92,201 | 0.51 |
| Hunter  |       3 |    22 |   1,159 | 0.67 |

Reports and fights: 1ADrPJVmwgaxLqMc 1; 2aZ6C1vRF7DAHMBn 2-3; 2jJ1L8FPHzTB4fKG 1-5, 7-12; 2qAWx1L63KfkcJaD 1-9;
2wMznDRvmkpbtHxA 5-7; 2wPQmVqbdCcZW9gv 528 fights from 1 to 629; 3DPvXKRnTJWxZCt8 5-6; 3bdcFvQtDZhnXHfg 15-28;
3cabDLxwK71Gzq2Q 8-9, 11-12, 14-17; 43R1XHKpfLgybZqm 1-4; 4TxMfmFpzBJ6aN1j 5-7; 4ypMXT8RDZgcKVwf 1-12; 67fqCHN9ZManpxrk
4-20; 6gnM4ThfQPLHkG83 1-11; 6ht4rB7GNxMFRbCJ 2-5; 6jvXHDwNAakZGBxF 1; 6vAwndj1FxBcDKq8 3; 7HpBD1Z9fJcbA6Ga 1-2;
7ZWFDG4XBPKqf8zR 33-36; 7gW8baTzVZ9pYXFm 1; 7tbWfYqJFd4vzL3p 1-2; 8Gw7m39ZXFz1y2fv 1-32; 8JTGAcnFYvtmjLN7 5;
8cqCXZHgpw9Px2hK 1-28; 9JqKQBY68twcR2hT 1; A6qcQ9jtTNXKbvYh 6-45; AfZHhcnwNaRJYpgT 1-3; B6bvH1fQ4kAcyRzt 1-2;
BPVxqXWLc3FJTpMh 4-6; BfMQJT7g4mZL6taF 1-41, 43, 45-85; CApKmW3knZr8x4HR 34; CxbgQvRAL61MZdc4 1-4; D2nQcj6wFbmBgVyJ
1-37, 39; F4JAb8gjwrQDCP2T 41; F9hmfgQHZbYTKyz4 1-3; FBLTNkwYgtPzjyfh 1; FDmvgpZLhVbcHqxK 1-5; FhG3AxYL9ZNP4RvV 1-15;
Ft1rJVHX3kz7yLBZ 1-3; G1Q9MpkRF6BmhD2H 1-3; G29y3QBWV7XDR6mj 11; G7CL9NkZtpXWrBy4 2-19; G9NmY7ARvHKyb8Cn 51, 53-54;
G9TVWxhmHjdc3Ptw 1-3; GXw2BAjbtRHv9Fr6 2-5; GhHZQDn4zgckNf2R 2-4; Gqvf8dNR3mpJQg2b 1; HLtZjNGzq9bQc6rd 2;
HvtmnFTPNZQRWg86 1-3; HzKL2XB3rx1hk8Wy 1-4; JK6dgtv2N4QCqjnZ 1-2; JbQyngR8NWXKardc 60, 62-63; JyxTCZaVf2KznHw1 1-3;
KkRW7Am81LaJqHYz 1-3; L6fHWjbn3NVqgpkr 1-5; LF63qnYybWDpCc12 1-4; LrNcfVvw3PqBQ9CA 1-10; M1tPx9DBZGYgCa4f 1-3;
MnFgXa87WRKPmADy 1; MqzhQN9FPHL3fBvT 1-2; NjWKPpHgymCT832A 1-19; Nmv3nfLwY91hZy86 6; PKBd6Vz2Gkjwp9Zt 1-2;
PMtHv3y76FCj8AWB 2-6; Pk6V914wLHJpxA7f 1-3; QMLY8q2XNVry7bpj 42; QNfaZ9MTnWpFrqb4 151 fights from 1 to 199;
QZ3hRyAMbDrj9cHz 3-14; QkVRc329w8nLTzKp 1-3; QwmphLYJXT7GazMd 9-40; T2aWdxMQ4Bw9tLGv 43; TNW9rHBt3MD1X8Ag 1-2;
TNak4bKY8JvLqXDF 1-3; TYnZqwA2av9z6jG7 4, 6; Tg1Hcb7z2qAXWkCx 1; TmxZ31fRPBhjApqg 1-4; VL7Yxg3ZmJkNFM1j 2-3;
VY1qcrJ4HxjCFnfX 1-2; VZAm6yQkGwD3XPKn 1; XWaTVPCnmNAzD7FJ 12-13; XhVMPfc7mqNJt2Wn 1; Y31f7vWzA4nVF6rd 1-2;
YAZ7thCmypcGBTN8 11-33; YFtvD87kznr1CThg 6-7; YHRv2c3QqWZaLnmh 1-4; YmTFGMK1dg7XBz32 1; ZBXTKqHYa3QfkNxM 1-17;
ZLqKhFbkAypnaY9t 1-2; ZVf9tDamcR1hk6AP 1; ZhyVJcb2v1zaLx3j 50; ZpNAtvBLPMKmf7YF 1; aFtMHKAG1vjTCwcq 4-34;
aL7VmXZxPdgcBtjG 2-6, 8; anAZd29rK3vCJ4wQ 1, 3-11, 14; bP2CFzkZTLJ8q6Dn 1-8, 10-16; bPWwqdBY3vn8Kz6H 1-2;
bxWNdRVg4C3Jwaj2 1-2; cKvwPrNmjRCFzgpZ 1; cPkXNvDW4hnKqMJV 1-2; cgMGCdPJQ84wzaBV 12-13; dF3zbWfawRkZ2j6K 30-31;
dnDg7j9kRw6WtbML 61, 63-69; fMPL97aCA1ngvKQk 16-20; fRbxm7XVTNw8CYhK 25-29; fq8RFwhABxm2GY7Q 1-2; ftqKjgaPmbNvM21n 1-3;
gBrvVXRcGKFP6j89 1-2; gHmyrJdbAB2nGkvw 1-4; gRpyT8V4KkBc6fMw 1-2; gYm2HAtf4bnC8QZT 6-22; gnBdKW9myaxzNP3b 1;
gv9V6Ty1Qa3MBZW7 1-8, 10-36; h4m98xKRPayTnW2X 1-3; hZLykw9dRnYbtvKP 1-2; jC6FzGLJ2yn3kvw9 1-3; jNVAY2Pav8GcbFWn 1-2;
jXTVfy3ApWKPYwDG 1-22; jn3dwzWPZ6r2ky1K 1-3; kMftb1JGnDLjzPwm 1; kX7FtxzjfYnw69TB 1-3; mCNxAVQ4bj3tvqyJ 1-2;
mpgnDQbLWXY8jJP1 1; n7hQW3NYzJwmDaG4 36; nX1qbatTzGZ3CAwB 4-7; pRA73YvkarFVLNCZ 1-3; pdKXzcQAHr3MNhCa 1-4, 6;
qGMZVBwJfdtRxr6Q 1, 3-13; qpYKczLm83WFN7ZR 1; rFRYL1d4yZDcHgq7 1-5; rdcZKVykDFH4f7mn 1-4; t1yGxzP7L3Aa6CQr 1-5;
t2z1VPDRbwZqfcKN 2-16; t4bYnDj8QMyacA2m 1-2; tPbh63Rc4J8xdv1Q 1-9; thqTyPArmkxfMnK7 5-36, 38; tkPZYqTRgCyWG3jM 22-23;
tyrGHfnp3ghKNMz7 3; vZqVmLYQa6jwP3A9 5-16, 18-21; vazTdHkmWgA17pLc 1-2; wT6AKYHXqb3tMvGd 1-13; wfV2DJZaNtY8rAHM 1-2;
x64Wd3Z8A2TKCDXL 1-2, 4-5; xRPN2jwhCg6abJVH 1-5; xbc7XGytaKACYVT9 1-16, 18-19; yPVkKpqtfJM27WFB 1-7; yZvwm7pRYDkAPGL4
1-3; yfKb1MthLFDc4ANG 1-7; ygMZKDTYd4qCQ9mP 1-5; z9wcdxZaRLhWk1Qf 1-3; zVhC3bma8v4jTBp9 1; zqXGMvdTDmQaZgtL 1-2.

## Fiery War Axe (870)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18796
- Rate: 0.9 PPM, observed

1,343 procs on 25,187 rolling hits, 27 wearers, 35 runs, in the main hand: all hits 0.89 [0.85, 0.93]. The one-hand
white rate reads above the all-hits interval. Observed: the one-hand white rate disagrees. One-hand white: 1.06 [0.97,
1.18], 338 procs on 5,311 lone units, 26 wearers. Specials: 1.00 [0.61, 1.54], 20 procs on 334 lone units, 3 wearers (a
Poisson interval: under 5 wearers). Two-roll abilities, priced as one roll: 2.46, 9 procs on 61 lone units (Execute 9 on
51 and Mongoose Bite 0 on 10), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms,
316 isolated hits. Dropping the swapped stretches (30 active minutes, 3% of hits, in 2 runs) moved the read from 0.86 to
0.89. The paladin check, which sets no rate: lone white swings 1.18, 178 procs on 2,510 lone units, 24 wearers. Reports
from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      23 | 1,229 | 22,682 | 0.90 |
| Warrior |       2 |    99 |  1,880 | 0.88 |
| Hunter  |       2 |    15 |    625 | 0.40 |

Reports and fights: 1L73Gh9FVMm2J6jg 1-4; 2CKNRDMzhcwPnxQW 1-2; 34gY6KLVHPatQwMk 1-24; 3A1rzFqNBxTQykwC 1-2;
4C28YPXqgTMQJxB7 1, 3-9; 4LrMDtG6T8AVRJzQ 3-11, 14-28; 7n6rKTfk1vhzgXyF 1-2; CJFAZxNLH4BDV9Ka 1-2; DTaKC17dkxwcWn82 1-2;
GJwfBzhCqj6HtLKn 1-3; H8bX3GacPBh2ARCL 1-2; HKpGBT1PD6FtqxbR 1; Kzn8AkXFr3yvMZHQ 1-9; M3C87Yx2fKTvzgJ1 1;
NBZcz6JXg4GQhy72 1-7; Nr8n1KpjcYwk2CvH 1-5; TFhxVRgMXpzmBkY9 1-3, 5; WAxmzHdKLVabk6nR 1-8; Wb9tmLq2cGKFjw3v 1-6;
XgZdn3D9KR8cfLNY 9-10; ZTxYXrtHQkcNGjA1 1-8, 10-18; c1WPVLQ6Xw29RTCZ 2-4, 6-8; cD6XhzYbjHQFgpG1 1-3; fVvLCrnhpykHjQJ4
1-5; hVxH72qr8PYB3DNm 1-11; hvdx4tJf9QnzgrHb 1-2; hzCTANdXK1y68YnW 2; pravyY1JCwDAxZLb 7-24; rTLVqFpfa9GJQgHh 1-2, 4-5;
tvBjG6TXDZq8gdrz 4-5; vkayWxntqQJ6pfjB 6-9; w638C9zJtXWfHcQF 1-14; wJtPdpxFGQCM9f3c 1-2; xcRHnmfb7p1Ct48X 1-3, 5;
zDKZYwnyML7WFHXB 1-6, 8-17.

## Firebelcher (5243)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: equip proc 1291686
- Rate: 4.05 PPM, measured

1,932 procs on 16,748 rolling hits, 78 wearers, 92 runs, on wand shots: all hits 4.07 [3.89, 4.26]. The rate is all
hits, which the lone Auto Shot rate agrees with. One-hand white: 0.00, 0 procs on 273 lone units, 38 wearers. Lone Auto
Shots: 4.16 [3.98, 4.38], 1,699 procs on 14,409 lone units, 77 wearers. Direct: window [-9, 100] ms, median 20 ms, 98th
percentile 36 ms, 1,431 isolated hits. Dropping the swapped stretches (5.5 active minutes, 0.5% of hits, in 1 run) moved
the read from 4.05 to 4.07. Reports from 2026-11-06 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Priest  |      29 |   916 | 8,049 | 4.02 |
| Warlock |      26 |   741 | 6,376 | 4.10 |
| Mage    |      23 |   275 | 2,323 | 4.18 |

Reports and fights: 1FApk79HbtyDYWL8 1; 1RrDAhp7ZvVYc9P3 1-6; 1RrFMXYLPJhT7xBn 1, 3-10; 1ZJVMgcYWTyRQh2w 1-3;
1xDPwjtGBTzZWm7Y 1-4; 287trgyqQLvjACK9 1-4; 2DJrjC6WNLBxzgyh 1; 2fn6Cy1mJWzcgjAL 1; 2mNyRawv7fGMDYVx 1; 2mW9A7wQ4bhv1YMF
1-3; 3PnWT1jGz4tagB6r 1; 4NWMr2dJPyH9v8az 1-2; 4P93K2QXbxJ68qrH 1-2; 4vca1FKMyTnP97Az 1; 6WK7XDRHBNCx9kwn 1-4;
6WzAhcVgfLCt78nX 1-8; 6wxJrRWgPyQb21hm 1-7; 7xdwabvCPyBXYRFq 1-4; 8rdDTxLMQPKqYyAz 1-3; 9pTK7ZwNabjMydzV 1, 3, 5, 7, 9;
9wKZ21fVDH3xpvcT 1-4; APhXWvd6CkrHL4N3 1-7; B8Z4m1xL2dQNHgC9 1; Bm6KyjgHafc19xP3 1-3; CF1YrGNTBhQVp4tz 1;
CGkfgyRY8mQ1Aah2 1-2; D78yctPhbZzTFkCr 2; DHv3m4KLCdntG7pV 1-3; DNJ682zfmT7Zykv4 1-2; Dh2tFcWaMVng6Cdj 1-7;
DmkRhHYPg4wVJf3t 1; FADNxntvpJkfM9RC 1; FWf1LcA7PCb6NQB4 1-4; Fwr6n9BDgLaGPmKY 1-10; HKWTCMt2FqgnackY 1-5;
Jn4FXD6QKW71pRzh 1-2; KBy12aMkJxXgA3fm 1-2; KDFtbX6ZVxMgATLh 1-2; KYD67nVqZT4jxWyN 5-6; KkCznWJf8yYHvGgX 1;
LAP9T3KqpjhC4MD2 1-7; LbmB2QdWnPT7JGAF 1-6; MN6V1awdYqymj4vr 2; MX1B9kCJ7GDTQ6za 1-3; NGhbmnMwKCHYfaX6 1-5;
P4wbLW9ngQ2Kx1Ck 1-3, 5; PCYpB6DhnQMmXbzN 1-2; PLKnY9R2BDgwbdWv 1-2; QHqMJPfgbBCvZan7 23-28; QL67hAdbrGV23Jp9 306 fights
from 1 to 327; Qvg9N1YBc36d8Rnw 1-10; RKdyjLbQCZgtzrXp 1-3; TGVQg7A3KfJdpbkw 1; TQCphZqxGzRmkgjf 1-2; TkMHbgQNKBt7Vzy9
1-13; XPmJACH8RtqFd3jM 1-3; YB93KnL7VMdtJWH4 1-5; Z9qbhQrJg2Wfz1FM 1-2; ZhHrNFMTzqVR4xvw 1-3; aNFQk87zxVRKA2nf 1-6;
aryNkPZX814F37pJ 1-2; c3NLR49a1VwfCQtG 1-7; crGPAQ6ft4NwXHbx 1, 3-8, 10-11, 15-16, 18-19, 22-24, 26-27, 30;
d6bwAkLKjGq7pVMy 1; gmTPbGL4cNDJ92fH 1-3; kBzVYb3Cy7QwnKMD 1-6; kHyCnvQ2Pdc6VKfr 1-2; kpNtDy4ZYjLTGAbd 1-6;
mPz2LqFWbhKcV7QN 1; mbBT1kLrM8qXzdJF 1; mjZrgTPwCD4d21h3 1; nT2Wz7mVQZbtBRG4 1-4; nwFh1bBxDW6JGKpv 1-7, 9-16, 19-32;
qFZN9PQwagyL3Xnz 1-3; qJNfkaCxZF3DzBh6 1-5; qL6waTVj3WhQYk8x 1-3, 5-21; r6dAtgyVpcMXkL3T 1; r9X8bgWqQfnPVDyB 1-2;
rCnfgxYFVXLHvjZp 1; tAZ12pKHxQYLydgv 1-2, 4-5; tQc389CTMRrAvqFg 1; trkbKMT6maWx1HyQ 1-4; tvDFpcQ4w3L7PBy1 1-3, 5;
txZvdrQzpwn3V4kb 1-4; vVa1NxRf6rZnYhWH 1-2; vWGgArDXtc13Lx6J 1-9, 11-14; w1pq2CGTWMbX7dQZ 1-2; w7pg31PYHx2mra6Z 1-7;
wDqmkbxdTXj1rA9F 1; x61fvm9RLMjXyDqB 1-4; yAGCbvd2r4RZNDaK 1, 3.

## Firebreather (10797)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 16413
- Rate: 2.05 PPM, measured

2,870 procs on 38,536 rolling hits, 24 wearers, 28 runs, in the main hand or the off hand: all hits 2.03 [1.96, 2.10].
The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 2.05 [1.78, 2.34], 216
procs on 2,875 lone units, 1 wearer (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.76
[1.54, 1.96], 417 procs on 6,479 lone units, 23 wearers. Specials: 2.15 [1.94, 2.37], 390 procs on 4,947 lone units, 4
wearers (a Poisson interval: under 5 wearers); Cleave gives 207 procs on 2,561 of them. Travel time: the proc lands a
median 102 ms after its hit, 98th percentile 257 ms, window [-20, 267] ms (898 isolated hits). Dropping the swapped
stretches (29 active minutes, 2% of hits, in 2 runs) moved the read from 2.00 to 2.03. Reports from 2026-11-14 to
2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       7 | 1,499 | 19,676 | 2.08 |
| Rogue   |      15 | 1,340 | 18,451 | 1.98 |
| Hunter  |       2 |    31 |    409 | 2.06 |

Reports and fights: 1GhHmWBy27jg8n96 1; 2Wd93tJTkAYDyVQb 14, 16-66; 2pVHBkGvWDx7yrKq 2; 4XLqQjchbAzvVG9k 53-54;
CqDJ39VhZFB7pgnW 1-2; F9hmfgQHZbYTKyz4 1-3; GHK2gB78FMVCmPfW 1-9; HMdLhWY3pNbnXPtF 1-16; KpcADW3GwYqXzgM2 1;
LFzpja9rZBXCxkNw 1; Ncp4bVY3x1DXRqvA 1; PcVWH9XhGRCvb1Y3 1-2; QH63NmDVdWyrqwak 1-17; QMmL4NtaPgwZAVjp 1-2;
VafQDAvHNYnFJGyp 1; XNQzBvfdcYb1CP3R 1-3; XNhGAyQZTHcM64L3 1; bNvV2YCTJxrdw9Gk 1-2; dqBNabD23hC7YmtR 1; g4KqJfjzTN8AFQCr
1-17; m7qPWtJdX1VHzr8T 1-4; myLBa1DdHY8K7Ppr 12-59, 61-69; nzWQyZYJaRqvcTrm 1-3; qL9HjmZWxGKwkMr2 1-3; qLRNfvkW2DnHPbGc
2; t9rhYHPwgcv1fpZb 1; ybf3pCJXnTQD2V9G 1-2; ypwYm41tKWZxV69q 1-5.

## Fist of the Damned (10804)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18084
- Rate: 3.05 PPM, measured

5,955 procs on 61,469 rolling hits, 45 wearers, 48 runs, in the main hand or the off hand: all hits 3.06 [2.98, 3.14].
The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 2.99 [2.87, 3.16], 1,692
procs on 17,854 lone units, 30 wearers. Dual-wield white, which sets no rate: 2.86 [2.62, 3.24], 425 procs on 4,691 lone
units, 15 wearers. Specials: 2.88 [2.65, 3.12], 506 procs on 5,548 lone units, 31 wearers; Revenge gives 350 procs on
3,975 of them. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,742 isolated hits. Dropping the
swapped stretches (99 active minutes, 4% of hits, in 4 runs) moved the read from 2.92 to 3.06. The paladin check, which
sets no rate: lone white swings 3.04, 146 procs on 1,516 lone units, 27 wearers. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      38 | 5,578 | 57,742 | 3.05 |
| Rogue   |       6 |   279 |  2,578 | 3.42 |
| Shaman  |       1 |    98 |  1,149 | 2.69 |

Reports and fights: 12MCBnGNPTK9pwZD 1-16; 1GN2LMtwvY4JQHba 8-11; 498vhQAtCw2XgZJM 11, 13-15; 4vcn3WF1RxXf8VZt 6-17;
6MYqkGpyJxDVCvRg 1-3; 8xjZmDw2LKv9bckh 1-15, 17-23; 9MCqcwjzLk4JBARG 1-5, 7-14, 16-22; 9WGcCxKHQ4pkX6Rj 1-11;
AaLcfw3PT7hbmMxX 1-2; B1qcTGpnHCrQ64Xa 1-4; By41zJVmM3cdPQkG 1; DqmQrC2cnFKjy7Nb 1-2; JMnDtLGxf4FqCV6w 1-3;
JbQyngR8NWXKardc 4, 6-9, 11-16, 18-24, 27-44, 46-47, 49-60, 62-63; Jhyg1aZzNTAq2w4t 1-8; Jtp3TQPnv1aC9ZNj 1-20;
KYmwz2q1tG8yAcdP 1-6; LbPgnwx637h8JypD 1-2; NCHPzGaKRd76h1w8 17; NcwFg2bZjkrnXfQV 3; PrcWByj746kYq93v 3;
Q8MGr7WxPzRhYm4q 1-4; RTfqrCdVvLGYzj1A 2-3; TQhc7qAF6VJD384B 1; Tx7MPH9f8Ga6hdrF 1-20; V8F7qLQXzpPAmcyG 1-4;
VjT3aMZghKCyzGrt 1-6; WMBFLQ43XvqTK7rG 8; Wb8w1TtxA3kLF9M4 1-8; Z4w7HKFPvcrTWd3M 5; dJLcNXmGYr372A4P 1-6;
f9CamrL48AnykH6N 3; fNMJcqDHb9ZzPLnk 5-10; hc98g3LWa1wznbAf 4; j8bQnLtWwRMBXkDz 1-4; kARjmXZ3nwxC9vg6 1-2;
mbZcgPavhdG7fKDW 1-20, 22-24; nmwM719fpABxvjdz 5-6; pJWQ4namMhyTfcdX 1-3; qprynNbXtRZV4DHP 1-7; rf7pqCmzaxAJ9yVD 1-2;
tDbxQwTkFmYRfjLZ 1-2; tMwA9Zrfj7b8JGLy 1; vJWCyGPqBAZ6btF1 1; vw3zybZTQjH8crG6 1-2; w4fzHNVk2dAhLKnv 1, 3;
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
melee hit of the paladins (abilities, seals and Windfury Totem hits included, priced at 3.3 s), one proc to each
Fire Shield apply or refresh, the rate is 2.05 (1.96 to 2.17). The rate holds across builds: Bear Form 7.5% on 345
swings on 70888, 7.7% on 13,118 on 70938 and 8.0% on 2,889 after 70959 was listed.

One hunter in 6yHgkT8LqP42NmxY, left out of the table, read 20 procs on 170 isolated white swings, 11.8%, 2.14.

Fire Shield applies came 100 ms to 1 s apart 50 times and 1 to 5 s apart 508 times, so the 5 s category cooldown
on 16559 locks nothing out.

## Frightalon (14024)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 19755
- Rate: 2 PPM, observed

6,111 procs on 131,358 rolling hits, 100 wearers, 126 runs, in the main hand or the off hand: all hits 1.99 [1.94,
2.05]. Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white, which sets no rate: 1.93 [1.85, 2.02],
2,614 procs on 58,091 lone units, 100 wearers. Specials: 1.57 [0.77, 2.37], 34 procs on 929 lone units, 10 wearers.
Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,516 isolated hits. Dropping the swapped stretches
(118 active minutes, 3% of hits, in 12 runs) moved the read from 1.94 to 1.99. 2 runs logged in two reports count once.
Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |    Hits |  PPM |
| ------- | ------: | ----: | ------: | ---: |
| Rogue   |      88 | 5,598 | 120,618 | 1.99 |
| Hunter  |      11 |   452 |   9,485 | 2.04 |
| Warrior |       1 |    61 |   1,256 | 2.08 |

Reports and fights: 16hJKBmq2V8tnzXR 1-3; 1F8qRr72jxvZYdzy 1-4; 1aM8GxnBfXmWDJkV 1; 2H1VAy3MDrRndPvw 2, 4-6;
3B8AcaGNRKmdgPDj 5; 3BWcJRwaMgH92h81 5-12; 4av7FfKdZjBh9Yp6 1-7; 63QrzvTALnZKgx7Y 1; 6NYawKLDtkJWPR2H 52 fights from 1
to 63; 6XwgYbxq2GycCFP1 3; 6gDPKpbZAYN9XCxn 1-3; 7TXcFNRzjwHdAybn 1-19; 7ZWFDG4XBPKqf8zR 33-36; 82WLwYNzvjB3bXM6 1-2;
8VFd9gqxLY1DkhaK 1-15; 8cqCXZHgpw9Px2hK 1-28; 8hBd9Wz3AXHwZf21 1-2; 9QKHCPLay1fMvtpc 1-6, 8; 9TtNqcjaBvHFYQ61 2-3;
9b71qzjAg6Myt34K 1-3; 9x1Phv3ptzwqQbWa 1-2; AT2jv6mC1wL4KFkJ 2-3, 5-6; Ba4bzVmX2nDrRN7k 1-3; BkTfV9yAZJX8P7wF 1;
Bz2NMjHnY7PD6ykX 5; CBGfNpnjv2ALw9Kz 1, 3-18; CtMF9BgnAmWkTbfd 1-33; CxpHYfFXgjPchnyb 3; DAgp7nkzcLBXV3GM 1;
Dv3cMWKkhgYfVmJT 1; DyMmJxc4t9pw6PAz 1-3; FHwByWtYngL3dDVp 1-71; FrL4wayTHRtW1CKx 2-15; Fz9b1xLP2NRpMJdk 1;
G7X1CkT8z6PrfjWM 146 fights from 693 to 871; GKWHyVpXcPx8BCNb 1, 3; GpyT8hfkVXmWPvnA 1; HB7yCLD3wvVx4fj9 1-5, 7-15;
HQtGDN3P9g1Vvawr 1; HjTkmZRwD8PqWLv7 4; HkynYAjBCvN19rmh 1-4; HvtmnFTPNZQRWg86 1-2; HwyJMCLKfbY1PkmG 7-8, 10-38;
JChQqKVgLWykANRr 9-46; JQwbjnVYyGcP4rg8 3-13, 15-17; K6VvbDypXFLQJW7d 1-27, 29; K8VjmZgFCd1D39xc 12-42, 44;
KG62VXF3a97CxtAg 5, 7-9; Kphk63jtr1wb9NyL 1; L7bxWV42Z6cTnp3k 1-3; Lxy1kqgVvFc24aPr 6-11, 13; MRLXmFGyV268jWwt 1;
MgmdJqxBWA3Tw2vz 16-17; Mtqz1p49BfvnGyAd 45; NLftBVcrRhPwnv62 3-31; NTJzVdMZBQDaqG83 7-8; NWDCd6xf9AXK17yv 8-11;
NdntYGabAvy6LM1g 12; Nxtj7m6zcFX9Aq3Z 1-5; P3QtFzTLVp1aWBbA 16-19; PMjN9QGK7tw2CJWq 4-7; PMtHv3y76FCj8AWB 1-6;
PqK7NWmpYF1zA4fT 155; PtKXrwfj8zBHm2Gy 1-8; Q1pc9MH47qTnrJ3W 1-2, 4-5; QrjXL7FYdBf2VmWJ 34; RJXQC67jbzATWdx1 8-9, 11-16;
Rry6WVb2NGD1dXMx 56 fights from 69 to 137; T7dtHBv9YR2jnLwg 15-19; TQYALMf39F2jWh6a 2; TYnZqwA2av9z6jG7 4, 6;
TmxZ31fRPBhjApqg 1-4; V7cnHk1QB3FA8WYG 7-8; VGwz9XY47JAvChQy 1-7; VJxG2agMXqKLdQfy 1; VjBTMHXxFvbthkw2 1-3;
WNV2t9PaqjQGdBDc 37 fights from 4 to 51; WqfyDhVHxT93Q7Pt 19-30; XfrWGm96NY1BTMng 1-2; XrpxL4B8mPTWQkFK 1-3;
Y1qVpGRtMAxy4Pvk 3-4; Y74AgtbrWTvyNHQk 6-7; ZT82B6pDVFn1mRNt 1, 3-5; ZYatFp24x3mydH8j 1; ZvTrcNPQFKy92djf 1-7;
ZwvyzH9bxRN3ghCX 1-2; aFgwBnG3t6kjRPNC 1; b2cX8fQ9GmMYkLwy 1-3, 5-9, 11-20; bD4AfYM1QHdJpWac 3; bYdJCZ3RKnrj7Ly1 1-4;
c72QGjpCxhmwTrBa 2; cKG7yDArxfqhJPp6 1-9; cZf8ygFQTdAYNa1w 1; cbTt1zxajFfNVPDA 1-4; dnDg7j9kRw6WtbML 49-52, 54-61,
63-69; fkZgABGacHwdpLF1 1; gQT9G6Phq8bJ7vFd 1-2; h2BF8zNyvKfMgZV7 1-2; h4JXmPQDqWL61Cyc 2-3; hYjVWRNBGDKLxTpc 1-3;
jdRNZwApWYmKgLh7 1; k3ZGwNbF9W7DgzLQ 1-3; k9fKT8ZCtMFn3vab 1-3; kph1Df7CqFVxQ2PW 1-2; mwybgGpBr9k6aNHc 5;
myLBa1DdHY8K7Ppr 69; nX1qbatTzGZ3CAwB 1, 3-7; nzcBtpgDa3dLrTmV 1; pBRwT8xtckgrMFLC 1-3; pMKBcfmZNz1WJanV 1-15;
qbtZdThvBCXVcfAW 4-6; qmfjQ63TN81AnGkB 1-3; rFRYL1d4yZDcHgq7 1-5; rRG9MzWL4dQJ38hc 1; rfz4pR6mx1w9JWKA 1-2, 4-7;
tbZ6pxfrDvPzCnT7 1-2, 6, 10, 13; tkPZYqTRgCyWG3jM 22-23; tw8ZP2krfqXvAB4Y 1-11; xatzTjyc9GAh3MPC 1; yfN361HMZk4vp2hm 11;
z6LMfrpbwhyVCAJK 5-10, 12-21; z8QPNWpMBx4VgdDX 1; zjwVLbAaYgPHC6GN 3-4; zp9yVrtfTNZJCv7g 1-7, 9-10.

## Gahz'rilla Fang (9467)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 3742
- Rate: 2 PPM, measured

5,255 procs on 87,272 rolling hits, 97 wearers, 106 runs, in the main hand or the off hand: all hits 2.01 [1.95, 2.06].
The rate is all hits, which the specials rate agrees with. One-hand white: 2.16 [1.54, 2.95], 39 procs on 602 lone
units, 2 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.90 [1.79, 2.01], 1,724
procs on 30,180 lone units, 90 wearers. Specials: 1.91 [1.64, 2.16], 210 procs on 3,664 lone units, 26 wearers; Sinister
Strike gives 100 procs on 1,836 of them. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,480 isolated
hits. Dropping the swapped stretches (322 active minutes, 8% of hits, in 16 runs) moved the read from 1.84 to 2.01. One
run is dropped whole as swapped. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      76 | 4,551 | 76,406 | 1.99 |
| Warrior |       6 |   381 |  5,818 | 2.18 |
| Hunter  |      15 |   323 |  5,047 | 2.13 |

Reports and fights: 2BZkQ4tRcFfdL37H 1-8, 10-11; 39NVvfaxYwp1nb4q 1-11; 3jqfdPNV4Xwy7g1Q 1-18; 42LCADRGWqxkZctQ 1-9;
4PtKGL1zcRV8Chfn 1-2; 4RVqrpXkh7dKftAC 1-2; 4XDM1GpQzgPdk79v 1-5; 6CWvYhN1HVydfML2 1-2, 6-7; 6DZv2FAgWq7YrV3j 1-19;
6ZYhdMNTCGq9cVpB 1; 6aCfcqdLmAPWhQR4 1-11, 13-14; 6j73x2mkhf18LJyH 1-7; 6q1mWrtM7zXFckvp 1-3; 72JjqrWXMdtwKx68 1-3;
7DYwVnCaThxgmyJL 1-3; 7F3hKPycMYjXVJkb 1-12; 7zBWfJYckMgvrxtK 1-3; 8tBp2ZK7XT4QJCgD 1-5; 9YaF84QpvXJfGTRb 1;
9vWqJjhLGCpdQZKc 1-11; A3hT8CQBPJXWmtZM 1-4; A9J8jfRGY1yvZQnt 1-3; AKMJvHDZGq8fcmTz 1-2, 4-8; AwCa8WG1HjFKXkDp 1-9, 11;
BbHp1TnJzxk9RFNK 1-11; C8Qd1LXMrGT62kpY 8-9; CG6c4RDkPfHAgYNx 1-3; DmjBrYd7M9VZwCHF 14-22; FG6r9y2WNZAg3XQT 1-13;
FHnX82dpGtm6ANYh 9-19; FndJyRm1B9Dt8w6N 1; Fq4fzXtnj8PQHmLx 1-12; G6yaZfLzMtN13hcA 1; GJRTHLXkYr3dhC4j 1-3;
GNmhV1FHCPQx4rBt 1-2; Gzb3PZXB2FD1agYC 1; HbJrLPfk7xangFKc 1-3; HvCXkJATnhtLGpP3 1-11; HvrXCc2kYmnyZhVQ 1-8;
JvXt1qGpwxz7NDWk 1; KaHtpqfx2APZBQrw 1-12; KpWJMkH19dP8RGmb 1-2; LjqtT4FBgNkbdp1n 1-12; Lx1YbG9MgQ2ypc73 1-3;
MaQDxHvy7bc9XL2Z 1-3; Mjr1czR7AknXqLvd 1-8; MzFAfGZpt1nmRX7J 1-10; Pjkc2mBbhdYnQ9ZK 1-9; QH63NmDVdWyrqwak 1-17;
QPbNgVDfAMLmwC2X 1-9; QkLnHByzf1pDcJ63 1-10; RTb1ghjZCncVk64r 1-6; TbB8q4RMK3LmYJNv 1-7; VLmAcpbQxWf1CtaR 2-17;
VgmapW6THJ2hCrzY 1; W8y23dvYQgXLzxmC 1-3; X39gqwvYCQ87K6MH 1-3; X47TDyBdAh2nQ6LK 1-10; XHDjdLb7R9P6NQm1 1-2;
Y4znCFdR1QD6LZJj 1; ZLpcV6krn1wYHJvt 1-8; ZwxX9G3bRPFYfjng 1, 3-4; a3J47BPrHyxVQpnY 1; b3X8aj2Hxw7FAdcR 1-22;
bfvxDCF6qAZpQKVh 1-6; cN8Qdpg1BLy4RMam 2; cPLzXv4nTtNaYWF3 1-20; cQmtM3HFqA9h6vRK 1; cQpq1A6BxVMZFmfr 1;
cta2vw4CNDWrMRxQ 1-3; dKW8HMQmFNr1V9z6 1-11; dt4M3ypRnTkwaVQY 1, 3-10; f9gwFtJYnNyPAT8k 1; fA7MqjWypQ4hZJ1d 1-3;
gJhDBWtN6jCYLMxF 1; gfQkaTK4pyh3mD7X 2; gtpTCvBjWXdRhxnz 1-29; hMdWYxyPG3rn9VQ7 2-12; hfBDjZxwkYCV29z8 1-3;
j8WVkrG1mhYyfdCK 1-3; k2KxgByR4dTpLQtH 3-4; kFj1wxgGc4LMnbqD 1-3; kKQTqfnV2MCx6dNB 1-10; m9AhbXcTvr6gdp7M 1;
mxh9rJfdbvPkRtTX 1-2; mzgDjAhkWY8bdnxr 4; nJqcvxFVj83NTrKR 1-5; p34dnvkg7yDKL6cr 1-3; pGQn8NhC6y1gdt4T 1, 3-11;
pzZAwR9hgmGFcTKb 1-5; qYZPRjVJBQk3mnyN 1; qfA7gQpmnrLWC2zF 1-5; qwfRY7XxAy3MBcn1 1-21; rF2v3ndb4LWVzZKq 1;
rkwFCLMdVXn2D8tf 1-3, 5-8; tac7hXVRgkbZNMDy 3; vpTQqDkgNYLnjtZr 1-47; wFBytQxvWKz1Xf2L 1-5; wHt9bcP6dW2GxXAQ 8-17,
19-20; wXJGrhP2nVCQyctm 1-2; wnMHyXgW6KpAhYF4 1; x3bWXZa1KDCmFvtn 1; xYmak3JcyZQHMtVT 3; xcJvZFQ4X186mGdw 1;
y2dFcNR1rMmKTPAj 1-25.

## Galgann's Fireblaster (9412)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 29644
- Rate: 4 PPM, observed

15,598 procs on 89,641 rolling hits, 90 wearers, 101 runs, in the ranged slot: all hits 4.02 [3.94, 4.09]. The rate is
all hits, which the lone Auto Shot and specials rates agree with. Observed: every wearer with a proc is a hunter. Lone
Auto Shots: 4.02 [3.94, 4.10], 7,757 procs on 44,525 lone units, 89 wearers. Specials: 3.99 [3.88, 4.10], 4,979 procs on
28,806 lone units, 89 wearers; Multi-Shot gives 2,746 procs on 16,155 of them. Direct: window [-10, 100] ms, median 20
ms, 98th percentile 35 ms, 11,793 isolated hits. Dropping the swapped stretches (31 active minutes, 0.8% of hits, in 4
runs) moved the read from 3.98 to 4.02. Report DBQHpyFz7chZRM6j logs the same run as GYNx3nzHdpr7X1AW and counts once.
Reports from 2026-11-10 to 2026-11-22.

| Class  | Wearers |  Procs |   Hits |  PPM |
| ------ | ------: | -----: | -----: | ---: |
| Hunter |      89 | 15,598 | 89,640 | 4.02 |
| Rogue  |       1 |      0 |      1 | 0.00 |

Reports and fights: 1N379VLvJZXTy8rA 1; 1PVRtQZCwfTkmzx9 1-2; 24fAZdgMD6K9qx8k 1-4; 2PLVabYtpvxfk1Q6 1; 2xY1D3mCwagdy7Wc
1-2, 4-6; 39NVvfaxYwp1nb4q 1-11; 3LTnjm16YN2bDzCP 1-6; 3MY1W4gBv9yXakD6 1-8; 3WyYMqr8ZxQRN1vD 1-7; 4CdDbTrZqQyxcnhP
1-13; 4dBcqKv1zXAGw2NW 1-2, 4; 6Mw2czTXbL3DqfyF 1; 6fBaJK4PC3HhVGTN 1; 76RDdfF4zVJwrLZp 1; 7PmMArDFy2nbCYpR 1-6;
7yTVR6tJhkcAfGDr 1; 8FqZbR32daMz17cn 1, 3-5; 8zp6dHfbqhMZnRwX 1-2; 9XQfBmZRaNWJkH3F 1-20; 9n4adYrfk8RDzmAp 1-8;
A3LGf2R4yKbgjarC 1-3; AKXHxFdgavCh9fkY 1-16; B2Cr1pw93vkndx6K 1-12, 14-15; BA6gQRpLNtcMJXTD 2-21; BPywV1KYCgALhmfq 1-10;
Bmw1jxaLbfzRF2WP 1-3; CGT7KvcFj1XM42Bz 1-4; CpDKr6AWJ4fw7xhg 5-7; D9tmTWQndR2jvAX8 1-5; DBQHpyFz7chZRM6j 2-10;
FVy1paktLrxzqG73 1-2; Fb8cX76vJCYhZftV 1; G1kbnQ7vdZFBMgY2 1; GYNx3nzHdpr7X1AW 2-10; Gf2q7HLBRytjmTkz 1-5;
GvAQn76tacyqYPMw 1-3; HkNB7A6tanYcGLWj 1-7; KmawtZQ437GH1j6h 1-2, 4; M9GRNjTt2BvdbQDh 1-5; MHL1xhpwC4dcVDZK 1;
MNHPYcWrb4mDR7j8 1-2; MjzWnNZP4KX2c6kt 1; NTrM6ZyaGc8P9kC2 1-7; NfyQFG14xJRcvbLg 1-7; Nw8YFxQnPvCVMHmq 1;
RXDCK9PtZy3Jp2QA 1-3; TymBrt4wRG6cFQg8 1-7; V1wJ3t6YjPKMfDkX 1; V3cYyjhtTgbdx82a 1-2; VKjNQYHcrfJvq9DF 1-3;
VnczrWJL6YKZ2DXP 1-7; WbJV9LdhQ8jvK1Mw 1-3; XZ76NC8hT1pR9xKJ 1; XfP6Zp3W2nAVb8jT 1-10; YbwrZWhBXqAz9vyM 1-3;
ZBtF1q8GTKRHDNML 1-4; aANyFDgb69hxqY2w 3; aNhzDXYy7MwVb98g 1-4; b3aG821WYQj9RVhK 1-2; bDzAngQZfHFMxjvL 1-3, 5-7, 10-28;
bhW9XwPZTgD3mkfQ 1-4; c32gbnm8dW4vhLZA 1-2; cAGW6Farb9N3d2t1 17; d9CVthbzgQD8Tywr 1, 3; dMX79JcWz1FKPm6f 1-4;
dmkBCMpWa92Pj6hK 1-5, 7-13; fPD87TvAkZ9XmpyW 1-11; fYcnmzKRW73q6Ztk 1-7; fqYWmr3cC98AwXaJ 1-2; g7DjqhM61aLN8XFd 1-2;
gcnGw3DaJdrFCAYX 1-5; gjDxfW9TJBQydAh6 1; hRPnJDL2T1bNXc9q 1-10; hY2gt1GvZjnwW6N8 1-4; hdm9KTLDjZ6NJHWG 1-6;
hmyFp9aMP2zTJ3AC 1-9; hwxjbW2XDv9BTP1Q 1, 3-10; jFvXRqgmH4QwWBNP 1-13; kRLm8NqHrpcDtGdB 1-2; kdgMTYt3CPQzf4b9 1-14;
kgJRQK82qL1CFPGj 1-5; kvQ9DtmnTGpZzgJ6 1-3; mMHVBTWdPvYL7awh 1-3; mNzF1MgYJypDfnPb 1-14; pCrzdDcxXPF3BYgW 1-15;
pJjZqyKxb81RYdvG 1-3, 5-16; qhCRxrzfkDFMKAtL 1-23; rGbWtLZqknvwxRmg 1-11; t9gkf1YHbR2NLj6G 17-26; tLGanMDfw82FJymz 1-16;
tk86znZXxMCdwBKb 18; trmvKnMhZLypqwAz 1; vpTQqDkgNYLnjtZr 1-47; vyTBDPbRVW2m3NnM 1-11; w6A8bVm7vTatMCYr 1-12;
wDyPtkh3FzT8LpNn 1-13; wVT17rNdaBZPftyX 1-11; xwT6Z1CV7WQhBmgD 1-3; y3wkJZnA7gv4qzfX 1-30; zk9DnPBZqyXArLca 1;
ztpB7m1kbLQ3KdqR 1-7, 9-14.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 14126
- Rate: 0.9 PPM, observed

145 procs on 3,108 rolling hits, 10 wearers, 10 runs, in the main hand: all hits 0.90 [0.81, 1.03]. Neither the one-hand
white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no lone-unit rate
checks it; the interval is wider than 10% of the rate. One-hand white: 0.88 [0.72, 1.06], 40 procs on 882 lone units, 10
wearers. Specials: 0.81 [0.53, 1.24], 18 procs on 432 lone units, 8 wearers. Two-roll abilities, priced as one roll:
1.27, 5 procs on 76 lone units (Execute 3 on 26 and Mongoose Bite 2 on 50), left out of the rate. Direct: window [-10,
100] ms, median 0 ms, 98th percentile 1 ms, 54 isolated hits. Dropping the swapped stretches (5.3 active minutes, 4% of
hits, in 1 run) moved the read from 0.87 to 0.90. The paladin check, which sets no rate: lone white swings 1.20, 106
procs on 1,703 lone units, 18 wearers. Left out: 24 of 34 runs in Cat or Bear Form. One run logged in two reports counts
once. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       2 |    77 | 1,738 | 0.86 |
| Hunter  |       8 |    68 | 1,370 | 0.96 |

Reports and fights: Cp4J8Vj1T6hZgWP9 6-13, 15-28, 30-38, 40-47; D9vw4QTYJZWMPnph 1-3; MRpj61GaHP4LYcQq 1-5;
XMgyVQRCrfNmA43j 1-4, 6-10; YJR78AMPDNKGQWjF 5-12; j38paTrnGPbyZzV6 1-8; jMJHNGFWKRwh6Afr 1; nkBgQMXYmZ6PDF8a 2-32;
nmFLdPWf8agkC7y2 1-3; pHFDL6NabPXM2QAx 1-2; rzAnCVwGgjTDHp63 4-7.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18081
- Rate: 1.9 PPM, measured

1,480 procs on 17,206 rolling hits, 13 wearers, 15 runs, in the main hand: all hits 1.91 [1.81, 2.03]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 2.30 [1.91, 2.75], 120 procs on 1,158 lone
units, 2 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.62 [1.36, 1.99], 110
procs on 1,506 lone units, 9 wearers. Specials: 1.90 [1.69, 2.13], 252 procs on 2,945 lone units, 12 wearers; Sinister
Strike gives 90 procs on 1,075 of them. Two-roll abilities, priced as one roll: 2.40, 4 procs on 37 lone units (Execute
4 on 37), left out of the rate. Travel time: the proc lands a median 113 ms after its hit, 98th percentile 270 ms,
window [-20, 280] ms (452 isolated hits). Dropping the swapped stretches (32 active minutes, 6% of hits, in 2 runs)
moved the read from 1.79 to 1.91. Report h8MG9XnN3qAd6rtj logs the same run as vkbTXzRajd3tF9Bh and counts once. Reports
from 2026-11-14 to 2026-11-21.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       7 |   751 | 8,609 | 1.94 |
| Warrior |       6 |   729 | 8,597 | 1.88 |

Reports and fights: 8bWqTVN2QCvK7HPR 1-13; 9cQz2ymYMgfL7TwK 132-140, 142-143, 146-155, 157-159, 161, 163-174, 176-181,
183-190; A3hT8CQBPJXWmtZM 1-4; A71f4dnHQwCBtykF 1-5; DQ62GWmAajLh4zbN 3-8; FLK4Vqw3cp6XkRAT 1-15; MJAm9hKWXyHLz1bC 1;
Mjr1czR7AknXqLvd 1-8; at2VQ6ZDvAf9CFwc 1-5; cx8pPRvN9Ta2zHg1 1-10; gT4Q3Mpw8DcjmbtB 1; gjhQpzGZdy27awb8 1-5;
h8MG9XnN3qAd6rtj 6; mNBAdW4ajKHzhMYZ 1, 3-5; vkbTXzRajd3tF9Bh 1-3, 5-20; z8CDMnPH9tcBGwrd 1-9.

## Gut Ripper (2164)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18107
- Rate: 1.5 PPM, measured

2,392 procs on 53,519 rolling hits, 41 wearers, 48 runs, in the main hand or the off hand: all hits 1.49 [1.42, 1.57].
The rate is all hits, which the specials rate agrees with. One-hand white: 1.04 [0.54, 1.81], 12 procs on 386 lone
units, 1 wearer (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.35 [1.14, 1.55], 439
procs on 10,870 lone units, 32 wearers. Specials: 1.50 [1.24, 1.86], 92 procs on 2,050 lone units, 25 wearers; Sinister
Strike gives 33 procs on 797 of them. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 501 isolated
hits. Dropping the swapped stretches (86 active minutes, 5% of hits, in 5 runs) moved the read from 1.41 to 1.49.
Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      34 | 2,157 | 48,506 | 1.48 |
| Warrior |       2 |   140 |  3,295 | 1.42 |
| Hunter  |       5 |    95 |  1,718 | 1.84 |

Reports and fights: 1LPQHj7FTxRgqGpD 1-2; 1MA3qWLGa8v2hdXP 3-48; 2Jjb9fcXa3VH1MPd 1; 2Tr9gxXZDdhnc4BG 1-2;
3H7tGVbMTnacBdjz 1; 4Mxqdg8YvmnbtNpk 2-10; 63zKLgVXPTnMB7Jx 1; 6Bzkp3W4KACvRVt9 1-15; 72JjqrWXMdtwKx68 1-3;
79jK6VvL4ZYbF1Jz 2; 8tBp2ZK7XT4QJCgD 1-5; 9VzpHPJraCGZQdh4 1-12, 15-17; 9fvqtGwK7RgT31xF 1; A3FdRp4h8c2Zfkt9 1;
AN4RP21ZMXLVpntm 1-3; Ag69GWdLRDhKCrcT 1-6; D9tmTWQndR2jvAX8 1-5; H1AnQVTYJzcam79x 1-8; JHF4zftRn3PaBbN1 1-6;
NPkAjF7cTYW16RKw 10-33; NrXFbz62xyhDR1LB 1-2; T72a1vxJZMBNbAg4 1-4; VG18w7rW3KRQa9B4 4-10; VKHmAfvPjBGLQW4Y 1-9;
WaHMR9NKLrwbzptn 1-2, 4-7; Y4znCFdR1QD6LZJj 1; Y8JX1mLtHag7B3Fc 1; bQCaW3RyBxn9Lq84 1-8, 10-17; cPGDLT3wyWYAxZkt 1;
fz2HQxXk96ywVYv7 3-8; g41xhtdFCXZ8Vr9z 1-2; gNBZx1vDYcAKtr6H 1; kvQ9DtmnTGpZzgJ6 1-3; kxQJCbtq9wLHmjKR 1-6;
n8VK43MDGChq6rgv 3; nGctFJyB8jwX9WHL 1; njJ2Ab1N47RfDH6F 1; p7QgHbGqKC2B6ryz 1-7; qNyQvka41YhVmdHR 1-4; qyDxWPJrCtGzQBcv
1; v2aNxQ7mGTdkJryV 1-7; wCTBgJrL4dGYn2bt 1-4; wmjHY7rqfAZL9cB4 1; x7B8mTqhcDFCpHNn 1-4; xRCbgPhKfWZMVLwt 1-2;
xjZknJqzfC82b3VA 8-9; yrN2TKtxzjRcpbfh 1-4, 6-21; zMqZ3BNKbFYfAjwp 1-12.

## Hameya's Slayer (15814)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 16406
- Rate: 1.05 PPM, observed

355 procs on 10,335 rolling hits, 10 wearers, 13 runs, in the main hand or the off hand: all hits 1.03 [0.89, 1.20].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; the interval is wider than 10% of the rate. One-hand white: none. Dual-wield
white, which sets no rate: 1.04 [0.89, 1.22], 157 procs on 4,535 lone units, 10 wearers. Direct: window [-11, 100] ms,
median 0 ms, 98th percentile 1 ms, 115 isolated hits. Dropping the swapped stretches (38 active minutes, 8% of hits, in
4 runs) moved the read from 0.95 to 1.03. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       7 |   302 | 8,673 | 1.04 |
| Warrior |       1 |    48 | 1,510 | 0.95 |
| Hunter  |       2 |     5 |   152 | 0.98 |

Reports and fights: 3kfXLnqH1ZmjzFA8 1-3; 872v1HVANqbJGzar 1-5; 9zxHaWbYT3gqDQvF 3-18; CfdHgJAvtMRhXNGQ 1-2;
GKCq4gHBvM37FhZb 1-3, 5; KXG2knhLR1jJqPft 1-5; PcVWH9XhGRCvb1Y3 1-2; nMYk1BHyvNzQ6xPF 2-3; ptTk1cMFjxaAhnPK 1-10;
t6pbGN1rhkTX3wLj 1, 3-20; vFQhXgw42KmfjT8W 1-3; w12Qy37HWnhkJzXY 2-4; xyAvc8gWdqpjXb6Q 1-3, 5.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13439
- Rate: 4.6 PPM, observed

6,030 procs on 37,400 rolling hits, 20 wearers, 25 runs, in the main hand: all hits 4.61 [4.50, 4.75]. The rate is all
hits; the specials rate agrees with it and the one-hand white rate reads below its interval. Observed: the one-hand
white rate disagrees. One-hand white: 4.18 [3.55, 4.39], 474 procs on 3,243 lone units, 7 wearers. Dual-wield white,
which sets no rate: 4.27 [3.79, 4.65], 391 procs on 2,619 lone units, 13 wearers. Specials: 4.43 [4.28, 4.84], 1,115
procs on 7,190 lone units, 19 wearers; Cleave gives 501 procs on 3,144 of them. Two-roll abilities, priced as one roll:
6.95, 9 procs on 37 lone units (Execute 9 on 37), left out of the rate. Travel time: the proc lands a median 112 ms
after its hit, 98th percentile 271 ms, window [-20, 281] ms (1,684 isolated hits). Dropping the swapped stretches (20
active minutes, 0.8% of hits, in 1 run) moved the read from 4.57 to 4.61. The paladin check, which sets no rate: lone
white swings 4.13, 39 procs on 270 lone units, 10 wearers. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |       9 | 3,032 | 18,070 | 4.79 |
| Warrior |      10 | 2,991 | 19,272 | 4.43 |
| Shaman  |       1 |     7 |     57 | 3.51 |

Reports and fights: 2aZ6C1vRF7DAHMBn 2-3; 4Mcm2YV8yDhPnwrb 1; 4hRktPYfnpycGzwB 1; 8qACFaGbm1jn3PV9 14-26, 28-29, 31-35;
BnbwWtd19XKryYMJ 1-3; Gak2MmQ34xn9ctPj 1-34; J38xbrFytgM9p1LT 7-19; KXwCgWRbN49fn3Fz 1-8; M7pfDPzYBVkWZ819 1-7;
MHC1aAYWtLfGq6n9 1-2; P4bXMwRjD1zyhZdJ 1-2; QN7nvZAP6Da8kGBF 1-2, 5; TF6dc3NCHbKrfxJt 1-6; VXBMnjLrGabm9TdA 1;
adwfGCJAYTmKV8Hh 1-2; cQpq1A6BxVMZFmfr 1; dBNfgVDLMmnc8h4v 10; jvqAx3thp6wTz9ZR 1; krA2h3ZCfBXMmT4F 1-8, 10;
nJZvaYCj2brRc7Fw 1-24; nyAx3XNDVmYWGdrh 1-6; pravyY1JCwDAxZLb 1-24; qR2rnzfgPakD879W 1-2; xbg9BypFcdVDwqMk 1-5, 7-14,
16-42; z74YTAg1Jc2QFWDC 1-2.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1298416
- Rate: 2 PPM, measured

1,714 procs on 25,785 rolling hits, 19 wearers, 25 runs, in the main hand: all hits 1.99 [1.93, 2.06]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.92 [1.72, 2.07], 149 procs on 2,331 lone
units, 8 wearers. Dual-wield white, which sets no rate: 1.56 [1.28, 1.90], 118 procs on 2,273 lone units, 11 wearers.
Specials: 1.84 [1.64, 2.03], 200 procs on 3,259 lone units, 19 wearers; Cleave gives 55 procs on 991 of them. Two-roll
abilities, priced as one roll: 3.33, 6 procs on 54 lone units (Mongoose Bite 6 on 54), left out of the rate. Travel
time: the proc lands a median 112 ms after its hit, 98th percentile 246 ms, window [-20, 256] ms (369 isolated hits).
Dropping the swapped stretches (75 active minutes, 10% of hits, in 4 runs) moved the read from 1.80 to 1.99. The paladin
check, which sets no rate: lone white swings 2.00, 48 procs on 721 lone units, 13 wearers. Reports from 2026-11-10 to
2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       9 |   678 | 10,077 | 2.02 |
| Rogue   |       5 |   541 |  8,101 | 2.00 |
| Hunter  |       5 |   495 |  7,607 | 1.95 |

Reports and fights: 3r4pzgfc172HvbwC 2-19, 21; 8HvfKkArXpmbhwTF 1; 9TtNqcjaBvHFYQ61 2-3; A1yxMTmKatkf3HRJ 15-17;
AP4vBCgQj78qD6zW 3; B4wWqM2TbRmdQy7Y 1-2; BDfZtjC3Q4wmV8dX 1-3; CmZh47GDdxbMPRqg 3-6; DNc6QALhxmGbCH3Y 1-6;
DyMmJxc4t9pw6PAz 1-3; Gkmvn4M2gNxdCDFB 14-31, 33; Hn8yXmQgGJdZawL3 1; JQwbjnVYyGcP4rg8 3-13, 15-17; K3xmjRCnMpV2c4qr 42;
KYmFPtyJ3gkzHfVM 102-160; Kphk63jtr1wb9NyL 1; L6pXTqJAGPwD2tHc 3; QFktMKaqvbLWP89d 1-4; fZdz8MyCHKFkrRwB 1, 3-7;
gc2BTyJdxXAR7h4D 4-5; gw8fpa1YVkv4WP32 1, 3; rTKYMd4yDvaL6Gn8 1, 3; thqTyPArmkxfMnK7 5-36, 38; xWDkmwXKrRcdM7Jv 1;
xyGf9QY4dwT8JHMp 5.

## Iron Knuckles (2942)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13491
- Rate: 1.95 PPM, measured

473 procs on 8,457 rolling hits, 11 wearers, 12 runs, in the main hand or the off hand: all hits 1.97 [1.83, 2.08]. The
rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 2.17 [1.85, 2.53], 164 procs
on 2,668 lone units, 3 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.85 [1.56,
2.07], 106 procs on 2,021 lone units, 8 wearers. Specials: 2.06 [1.55, 2.70], 53 procs on 906 lone units, 4 wearers (a
Poisson interval: under 5 wearers); Revenge gives 27 procs on 492 of them. Direct: window [-11, 100] ms, median 0 ms,
98th percentile 1 ms, 230 isolated hits. Dropping the swapped stretches (34 active minutes, 9% of hits, in 1 run) moved
the read from 1.81 to 1.97. Report VqHQ1CtWbcYamAXf logs the same run as H3vdcCr6JX1gwYtq and counts once. Reports from
2026-11-10 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   290 | 4,994 | 2.05 |
| Rogue   |       7 |   181 | 3,412 | 1.87 |
| Hunter  |       1 |     2 |    51 | 1.38 |

Reports and fights: 38D6JN4nZzjcGtHg 1-4; 4QAtNfTXMyDGPqCr 1-13; AzbN8GY9DnBrH4aW 19-31; D8vykhYFL14KH7xX 1-3;
DW7Cf4cgH9RykPXJ 1-2; DwWM9H3rpTnGPfbt 1-14; FfGCMqbHYTyJAN2r 1-11; GRTmQn3VDpwZKbCA 1-6, 8; H3vdcCr6JX1gwYtq 1-3;
VqHQ1CtWbcYamAXf 1-4; XhWYGfBAbj4RrJy7 1-14; ZPyMjD1FkVGJH8Lb 1-4; b3ny2awfxjWHqBFQ 1-8.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1295736
- Rate: 1.85 PPM, observed

2,538 procs on 42,822 rolling hits, 39 wearers, 44 runs, in the main hand or the off hand: all hits 1.87 [1.77, 1.96].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 1.67 [0.67, 3.45], 7 procs on 132 lone units, 1 wearer (a Poisson
interval: under 5 wearers). Dual-wield white, which sets no rate: 1.56 [1.11, 2.13], 54 procs on 1,093 lone units, 33
wearers. Specials: 1.68 [0.57, 3.38], 17 procs on 319 lone units, 17 wearers. Travel time: the proc lands a median 320
ms after its hit, 98th percentile 749 ms, window [-20, 759] ms (30 isolated hits). Dropping the swapped stretches (156
active minutes, 9% of hits, in 9 runs) moved the read from 1.70 to 1.87. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      35 | 2,479 | 41,674 | 1.88 |
| Warrior |       1 |    57 |  1,123 | 1.60 |
| Hunter  |       2 |     2 |     25 | 2.50 |
| Warlock |       1 |     0 |      1 | 0.00 |

Reports and fights: 3TDwbg41YhvNpPAq 1-7; 46cPKy8fYXgTNAm2 1; 67AMbHctYhxwQnfj 1; 6ryPTtKvZ1QJYdGV 1-2; 7nbQazTHpvmhGdRA
1; 8MLzmqPrb46VT1wv 1-2; 94TYgrcdVnJFKj62 2-5; 987q1pjnYdVQbCmT 1-6; 9vfFckp37GHYz6Lj 1-16; BqdMTtCvN3fVJpRk 29-35;
DRxwh2TLNPgrHJd4 1-9; FqW7GA1Lrw8jCpTD 1; H49wNaZhf6qgk1MX 1; HRpLt7nqfXz4Z6Gr 1-9; HndJP9Yf4xc7VzKw 1; NYVHr2XmGfwQndPp
1-11; NfqRryF7gnMcX2ak 1-2; Nn16MQr7kB9X2Fj3 4, 6-12; Pjz8kK2XvxyQrTCB 1-2, 4-16; QgWvNdYJnMRByc93 1-4; V4HkPzgr21bxRvFt
3; VDAtRmdp7vHQLyX1 2-19; W7LqyBw93GKFxfta 1-15; WpDkZ6JHRdqLvb3Q 1-3; a7nBvdhcXQ3VjAYb 1-10, 12-17, 19, 22-27;
aF69WRYXAT1qNZjM 1-11; bxW3HFgVBwmyDLJ9 1; cAYgpkdqXRDynzPK 1-3; cBjY4znbPJNH3XZv 1; fT2knhqYQ8wGNVgC 1-12, 14-16,
18-23, 25-43; hRzbpMrDY3cgQX7n 7; j3k86BpwAbgMK4qx 1-3; jRvABFgTGYx47cfP 1; jVhAWnaHJgf3FPmy 1-10, 12-13;
mzJyZrRNBpFqcx71 4-5, 7-10; p3NHKaZrFGhDdwTg 1-2; pdb9yZKJ4WxvrQqV 1; rqxmc6TzHXWk2dVZ 1-2; rtbzWXcByFwk7G8n 1-17;
t6rWYq7yLgGnvNp4 1; tyFZPdH39kGq21An 3-25; wW6G8YxfRH2bhcgC 5, 8-11; xwKnW41M9cB8RF6T 1, 3; zaLxkDJGH8YpVZX9 1.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13318
- Rate: 1 PPM, measured

1,257 procs on 23,890 rolling hits, 53 wearers, 59 runs, in the main hand: all hits 0.99 [0.93, 1.05]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 0.88 [0.79, 0.97], 289 procs on 6,157 lone
units, 50 wearers. Specials: 0.91 [0.75, 1.08], 84 procs on 1,728 lone units, 15 wearers; Cleave gives 44 procs on 1,098
of them. Two-roll abilities, priced as one roll: 0.97, 7 procs on 136 lone units (Execute 7 on 130 and Mongoose Bite 0
on 6), left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 117 ms, 366 isolated hits.
Dropping the swapped stretches (162 active minutes, 10% of hits, in 11 runs) moved the read from 0.89 to 0.99. The
paladin check, which sets no rate: lone white swings 0.93, 21 procs on 425 lone units, 4 wearers. Report
9mN62kzgMdKypTqf logs the same run as d7DFtXL2jfNa8Bw4 and counts once. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      38 |   960 | 17,785 | 1.01 |
| Warrior |      12 |   291 |  5,876 | 0.93 |
| Hunter  |       3 |     6 |    229 | 0.49 |

Reports and fights: 1g7hAY6CBNtW84HX 1-2; 29A3dKwPgpRNnWvc 7-11; 2aVnTxFzfckdXpBy 1-7; 3vDf6TMYBX9LK7bz 1-2;
3xgPkwfn8yQ9BqAW 1-4; 6hrcXFkgDamVB8jT 1-3; 6v4wTWQkjb3Xd7mM 6; 7hNgRwMfkQZ2Yv4b 1-3; 8GW16QY9BmFHJXLb 1;
8Qb1mLGNZxdgvwcV 1-2; 8RFv6Xyr43dLpJ7h 1-2; 9mN62kzgMdKypTqf 1-6; A3QMFqkz1wVa9ZHC 1; CRBxmcyLvpMVnGPw 1;
DLApcY8P6y2xWqRh 1-4; DaXWGJMR2bnc9YtV 1-2; Dx2tAm7YawZjV6hG 1-4; G3VKdgR2aLAkNqWy 1-3; GyLWvrkxKVmNZgw3 1-7;
H2nMWrcCz8FGdV6j 1-10; H8hqxdF3Tbgywr9M 1; KTMx63Xn8WykrCQG 1-3; M4THD3Qq9dBpbKvg 1-2; MrW6JCFZTDzdtYpH 1-4;
P16dx3VLmRyaDJZQ 1-18; P7cV2HjLgxGpmZQk 1-7; PN79G4Lfv3KBrW8j 22-323, 325-361; PWAwZgV9LTDrFy6M 1-11; QGn9gwBK4MD237xa
1-4; QhYmjJ4bHZXcgfk1 1-2, 4-27; QqKcMJwHjxNPkgRT 1-3; QrqZXKL9kbnY31Hj 1-4; T6MXt73janNJgmyY 1-3; Vvfz8wYbjTBFaxRq 1;
XVJtpHkwZy4ancAQ 1-7; aQwNVCPHWmzDLXqk 1-2; acHy1fNktP2m47pd 1-2; bdL8j9KGMR6r2xmk 1-3; d7DFtXL2jfNa8Bw4 1-5;
fn2Dc941K63CNvmj 3-4; ftNBwvGgPcMqh8dJ 1-2; gdKNPjVvcQt2LJnq 1-3; ghzjGKVBvJbRZnMr 1-2; grFA3BHxtCNyzR7v 1-2;
hbW1zgMXyRv8JAVL 1-6; jKJPF16ZnWLkrR2d 1-2; nT2Wz7mVQZbtBRG4 1-4; nTNyhMwtgDCbvQBx 1; nyhMJLqBYxPpRaXf 1-2;
pDRxq6wJynmhgdjL 1-9; pqr4HxC3vWJhfG7B 13-20; qLJ3wZVXQdnybvK8 1; tMxrNWC6qPVzgwAv 7-8; tjLnhaw7G3CQDRkP 1-2;
vJGT19YL26XFQpNB 1-2; x61fvm9RLMjXyDqB 1-4; zFR17VGkQKBPjab9 1-5; zJ6C2148dTYfaXHb 2-4; zhj69Dt2X1b3fvMT 1.

## Linken's Sword of Mastery (11902)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18089
- Rate: 2 PPM, observed

1,526 procs on 25,184 rolling hits, 13 wearers, 14 runs, in the main hand: all hits 2.02 [1.97, 2.08]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it. One-hand white: 2.13 [1.83, 2.43], 141 procs on 2,205 lone units, 6 wearers. Dual-wield white, which
sets no rate: 1.87 [1.42, 2.20], 20 procs on 357 lone units, 6 wearers. Specials: 2.09 [1.83, 2.75], 161 procs on 2,570
lone units, 13 wearers; Cleave gives 95 procs on 1,494 of them. Travel time: the proc lands a median 225 ms after its
hit, 98th percentile 572 ms, window [-20, 582] ms (135 isolated hits). Reports from 2026-11-15 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       6 |   932 | 15,127 | 2.05 |
| Rogue   |       5 |   529 |  9,065 | 1.95 |
| Hunter  |       2 |    65 |    992 | 2.19 |

Reports and fights: 17ymZpJ6xFPf4Raj 1; 4zQakdHFXwDYyTNK 1; 6DvQ2NMGgdj3cHk8 2; 8qKTwcCyPgd4B3xZ 162-179;
CMGKZHRVPwzvFmqy 1, 3-5; DW4wrpGfc6Xzkt71 1-5; JLt4PFa2qTYKh3XZ 1-2; KVQAk9qhBwWvp26L 1; cJFDKjvMtrXyZY7p 1-7;
myLBa1DdHY8K7Ppr 69; vwYj6cK8Xt24QLaN 1-22, 24-26, 29-32; xatzTjyc9GAh3MPC 1; yDgVQJwn2vd73WtR 1; zQYBA9vdKVkHapg2 1-2.

## Loksey's Training Stick (7710)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293482
- Rate: 1.95 PPM, measured

3,965 procs on 39,071 rolling hits, 200 wearers, 252 runs, in the main hand: all hits 1.96 [1.92, 2.02]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 2.11 [2.01, 2.23], 1,263 procs on 11,569
lone units, 186 wearers. Specials: 1.92 [1.60, 2.11], 236 procs on 2,383 lone units, 106 wearers; Raptor Strike gives 73
procs on 809 of them. Two-roll abilities, priced as one roll: 3.86, 74 procs on 371 lone units (Execute 57 on 295 and
Mongoose Bite 17 on 76), left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 1,472
isolated hits. Dropping the swapped stretches (90 active minutes, 5% of hits, in 7 runs) moved the read from 1.87 to
1.96. Left out: 194 of 446 runs in Cat or Bear Form. 11 runs logged in two reports count once. 2 runs are dropped whole
as swapped. Reports from 2026-11-09 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      29 | 1,896 | 18,952 | 1.94 |
| Hunter  |     160 | 1,074 | 10,299 | 2.02 |
| Warrior |       8 |   990 |  9,772 | 1.96 |
| Mage    |       2 |     4 |     43 | 1.80 |
| Warlock |       1 |     1 |      5 | 3.87 |

Reports and fights: 1QXAJwhcYL6M7TDk 8-30, 32-33; 1YLbjwRVPWr96gzd 1-8; 1aBhkzdDMZKGtb37 198 fights from 162 to 376;
1g2QAZPJrLTjCz3m 10, 12-14; 1hpF2AyzjRdrL78N 1, 3-12; 1xqBcj4CZQXV38Pv 1-2; 1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 1-3,
6-9, 11-15, 19-28, 30; 21gWyqM8QBJwrH4z 2-6; 27jZGhfHgp1LR8MB 1-11; 2PpAjf1qRmTv6gDz 1-17; 2TNQqFYdjRLB4Zga 1-10;
2WH4jLt3FADy8fnh 1-2, 5-6; 2aG3Ld8KVARCzrh1 1-7; 2vMwVxLqdB3GRZn6 1-4, 6-7; 39BFWKC2J8RTzyMx 1-3; 3HJrg1BGh8xR6Y2z 1-13;
3PXfhvmTp6NQdwRb 10-12; 3TfzmHM2k9XF1d7R 1-3; 3TgpnwxMD71qfLm4 1; 3wnhGaBAmXpFk9Pz 1-4; 4ARKtCTVhrgBwadG 7-10;
4BhykDjm2p1ncLGK 10-11; 4JYa2mQgpVNMF89R 10-21, 23-30, 32-82, 84-103, 105-110; 4V2RLGtF9anbCKWD 1-23; 4wHnDaRqY9Z2zv1F
1-14, 16; 6NrFMaRQ9yAjxWC2 2, 4-12; 6WG74An9RrCdqkwD 3-8; 6wmLW1N4tBXpTVqy 1-3; 7Vywm9BAZgcMfbT3 14; 7d6p4TBxcjFyPNVr
1-4, 6; 7fgdkyv1njDKa8Gc 6; 7g9DpfNFBA3PatdZ 8-14; 7nCPmLqHd8jMfQk3 4-5, 7-10, 12-13; 7tL3AyRHfawBJF94 1-12;
8VQDf7JtwbTknMLz 6-8; 8X2Kc46faTnxVBLm 3; 8ZdnfAT79zPwmyVM 1, 3-6; 8frHF64GzAhBxDyb 13-14, 16-22; 8tgNprk9LJwW47MK 1-5;
91XtwNrLRzvMkQ62 1-3; 91ZzNbJMRtjrLW6K 1-2; 97FVRX1WLHqDmjAQ 5-12; 9AhxYCTgbLfc4zKj 1-4; 9TrdzatyJbQV3vmc 1-7;
9g7NQBqDH6FGfP8L 1-4; 9nxtzJ1bC42FrgWN 1-3; 9rPx2JNYLbCy1waX 11-15; A8NWyY7dhpLDt3zX 12-30, 32-38; A8tQY9p3Dzn6TLRJ 1;
BHk3rNLb9PCd4pYJ 1-6; BTRMkxmvfrwcZNaG 3; Bmk9chwYjd2F3zpZ 1, 3-6; BydpaqJVW29G8QAc 1-9; C97aQyTJ4fPcrmjt 1;
CMxBvHAN37YRqbGg 5-6; CPNY7nxkfXpvQgH1 1-3; CXvjmgaJ6YqKzD4k 1-2; Ccvk1yDbBzFZAVqG 1-2, 5-9; CpDKr6AWJ4fw7xhg 5-7;
D6bxHwnpyvc2VNR7 1, 3-8, 10; DWJhjnmAM6GZdbKv 1-2; DcgdHfFZjb87qRJa 1-4, 6-9, 11-12; FDybHvQ3pPqBracN 1-4;
FQnMCW8kRqVv4mT2 8-13; FTXVdBrHJZD9aw7P 1-7, 9-23; FzGMchfD4PkdmJT1 1-4; G6XMnQyqxTdJa2mb 4-15, 17-18; GAdW4fkq7bx9BPmL
1; GBDAWjKNh6Q9XnF7 1-4; GJgyrn6t1vABDQaW 1-3; GKZJBtWAvfPL81bc 1; GgcJL2zbrt6nYXTk 1-23; GrW7bVHhKwJgF4QT 7-9, 11-12;
Gv6NDzAmFjC4P3qL 1-15; H6h87LGmw3gKRc9f 1; H6hnfmvc39wqtC1D 1; HL9zDvk2wWTXcP3N 1-6; HQTANmjP8FB29vYW 1-2;
Hp9ThMznR74PwYKb 5-8; J1QrDzNc4ydjV8ak 1-10; JZv9XDFLC1yKxrPY 5-10; JcwBmYGVgnMQH9xd 14-19; JtrVj2cqL3znKNXh 1;
JxaN6LDpwF2RTbgX 1-5; KLNhgMTDAV6xY1Wj 1-3; KapmH2bcWfPqnrY7 1-8; KmjAV91JBhgQLrfP 3-4; L6jJNDYaATph4cfK 1-3;
LJMmvtPTzdCpG9k6 5-9; M4G1VJQdXbvyYLBD 1-2; MVcaxDQwWLpnK6HJ 1, 3-9, 11-12; MgcQ6txHr9R1VjPv 1-5; MjVrxLNwAJK7kcWy 1-2;
MpjtZc2zqHaR9hC3 1-19; MyzjakDC9gBbpq6K 2, 4; MzQJ9fgkwjhW1Lv6 12-14; N7KgTPw3kzJdtLZm 1; NHvp73cXPQRC8agz 7-11;
NJ7vtRpCzDTfk2W8 4; NWvgFp3kfR2bHyYa 1-35; Nc2grh6ByLvJCkRY 4-6; NkTjKdbwYAtfJQBn 1-4; NkXjqd43AFVYpzGK 5-8;
NkyKCr2XmgLt1YWv 83-87; NzgdqKHnfLJWkFVM 1, 3-12; P3JcY6gWyFMX1B7x 1; PAmg9FX8fr23xHt4 1-5; PDT2yACNzd4LMZ98 1-14;
PLGKvft3bq8WgRYN 1-2, 4; PQDRkypFY94rzKwv 1-4, 6-7; PWhMrYaD4Aw8HcCL 1, 3-7; PcxQzDy3fF46mNKG 1; PmR8Lpvazd4q61nA 6-12;
Pn2LYHXpjZMDywgA 7-9; PqBvJ94Nhb7ax1Vz 4; Q1CHmPvpYaxM49qR 4; Q1arfj4WPBNXqmLk 13; Q2xTrDcZqjB9pA7F 9-18;
Q8GxhTmH9f6ajcNr 6-10, 12-39, 41, 43-51; QDadybK3XVPRvwLh 30; QJafvkCx4yhVXBmT 1-2; QPmw1C9MNc4hqH3K 1-5;
QYTW7Hza9nBjJFwP 1-3; QnjR1t9Y3WHpG6Vd 1-2; Qq6CvRP12rAGwngY 1-2; QtB8qHZ4X2pyrDPY 1-10, 12-13; R6xvn3yVf4DGJpka 1-3;
RNpVm3CcDJbxjB6k 11-21; RZhVjDGx1vLPkAag 1; RmTM8XL4aBJY12Ak 11-16; T6F89dBtbNG3MjpZ 1-3; T9cJG8bq1HaNyLQC 1-3, 5-11;
TQAKbWq7vVJHLgBR 6-18; TnLAmyRdXvz1cBhC 1-18; TvMmZw8nc4hKjd1x 12-13; V39PRKTHt8kahfxX 1-6; V4HDKvxmYCkR1JnB 1-3;
V4aAZ2PGgkHw6mvB 2-3; VF1b79fRAWhtGrwa 1-3; VZdxq6yNBrW48ztG 1-3; Vb3ZxfXtyDWr1ndP 1; VmQK1zM4hFHBxXgJ 6-15;
WYrCyJkdfMXw84Ph 13-21, 23-35, 37-39, 41-48; WbJDLzNGXAt9Rafw 5-16, 18-23; Wf26RdmcnLDTXwV3 1-2; WgtPVTmY4XQ2DrGB 26-33;
WqKd6nTV1GX8Qka4 5-10; Wr2mNV3McRqf1Zwg 1-6; YH3c4zvFrpLtR1M7 5-9; YJR78AMPDNKGQWjF 12; YLHkP9qZWz2b4AJt 1-3;
YZyAfq6XN9k4a8md 12-15; ZNQyKft4zAGCgcDY 1-8; ZjdqP6bg7tTR894y 1, 3-5; a6QpkvcR1YHLBqPh 1-6; a8bQ4fNyGVnFcZAH 1-10,
12-25; aDKxMy2WYAfcnvGh 1-2; aGWKDzhkCNJFXZgt 1-5; acMtqjQW3NhmCHL2 1-5; aj46hBQX1AfRFHtm 1-15; ajLmndyrzNJfwGCV 11-12;
amTJhvjMAyCLdbXR 1-2; b7PCQ1A8z36BDrTx 164; bCzt1Fq24mYcPXkB 1-2; bT3LyRBtax8FQnch 1; br4P3JTAWgLd82KY 5-8;
bvqJAwG8D4KycjhC 1-18; byYP67TH4KtZqm31 11-13, 15-17; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2; crtFfPbdavN6QHnX 1-9;
cwtKD6v8xmJG3jNY 1-20; d37CTyKPbHaV1Yzc 1-15; dBPjxVMQmgZ9h2vr 10-26; dDKQgLW3MFPatynJ 1-25; dnvYkhpKZQaqM2Hw 1, 3-5, 7;
f3HYhjJkZdtDgzM6 1-5; fAXbZvgpmLFxWw4z 1-6; fd4MgwavrqWcYHZG 1-14; fjqQNyMbD1Y7R92h 16-21; fnYThW3BpxKZv7jw 1-18;
ftkqY3cwR7dBZK1J 1, 3-5; fzKwkZLcbYraJdF1 6-9, 11-16; g7Bh9F2b1kz6Cnar 1-2; g7RXrQTKBCNMxqJc 1; g8BWDmknXJpvFbdf 1-4;
gAdCTkV1LBxNYFca 1-20, 22-24, 26-32; gqFJrKhdG4QvaP3x 1, 3-7; hBr7xLW6nkVdMKF2 1-2, 4, 6-7; hCzY42vqXrc1mG7w 1-4;
hrFwMYgVGq41fNjQ 1-2; j2RNAQYzZK6DyT17 1; j7V46tC2GkwNYTrz 1-2; jGL3RyT4fYMtwPrn 12-14; jGhpmnHrtyc1VD8W 1-4;
jXLYwBW7bZ3nQdfJ 3-4; kH2bjrXh4BF39GmC 1-2; kQhZqY4DwGBKLdag 1-4; kcrCYgFmXhZAVpGw 1-2; kjhVWgA3Gw7ypLtn 2;
kqVZPMHyFB4m9pK7 1-2, 4-5; mDbPCdX9AvwYJxGa 16-23; mH68YTDck3dKV4fx 1-2; mKB1pycRAjPFZCzW 1-2; mqjv1nLcRh6289yk 1-9;
mqnZ7bkf3v86prGV 4-13; n7qPAd4rN1JyHfpR 1-3, 5-9; nGctFJyB8jwX9WHL 1; nKhkfVbmGcg38CP6 1-2; nZ3NhbvpVXg649Bm 1-10;
naCtKF4L3qzpGPmf 4-5; p1zYDag2NHXPyRt3 1-4; p64RG9PVZcDg1XCk 4-6; pJ4AKDqYPvfwbTkM 5-8; pqfK2AgHZ7DnhvPF 1;
pxAgrJDCXHc9tbB6 1-19; q94LM8HFPN21ytpk 1-4; qGywkYrdHKT8NgAQ 1-2, 4-8; qMLQt41BXrpR9hkA 1-4; qVtTPaDMHfhmG9gr 1-10;
qj6gnCAM8kB2RNPy 1-2, 4; qzv6W9fkDaP1MZYH 1-6; r1xm2Tt3bjgBahZF 1-17; rD8m4JW1bVyYPKGX 1-3, 6; rPAKqpWdFTR6atmN 1-2, 4;
rvJfLGckxztPWVha 7-11; tBQwPjhFqYAyCv1L 1-16; tHz9M3arKp6PxA1W 1-14; tLkym2AVWdJ7bM9H 1-8; tMg1aAHQGJkxpYcb 10-12;
tQJfv7r6CMThmRnK 17-29; tWHjPvT9ZMfcAznY 1-5; tXFAxTLmg3bBNZrv 12-13; taFcCyAx9BD1KWXp 1-16; tkQKL9bGT6JVR3na 5-7, 9-16;
twRbTJHfrDXBLj97 1-5; v9Hf4bPtGq1LVykC 1-5, 7-8; vfFGXn4mbLAqkhWR 14-17, 19-20; vnrGwA8Z2RYtPc9g 1-4; w1aNzbJn2kMPpc3G
1-5; w6ZLGgMjrDnKvFWN 1; wDmf6L8WrGvHqXdQ 1-7; wXNczvgHAqT7L8nV 1-13; xQ2J8AbFN19pPD6X 1-15; xRTpHyY7qC1GmghM 1-3;
xYTCXa9myPJr6R8M 1-13; xk367aQTJVzfrKXq 1; xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6; y9hXfMV6mWvNtnQC 1;
yKvCQdzpgqY1WwbA 1-7; yR7t4NnKzkgph1bm 1-3; ythJ7TZDXk6AjHMq 9-10; yzFXjCY3WVbgRPx9 1-2; zNmFC3BVvQLy4nX2 128-129;
zjgkHTcZDJNv13bF 40-85, 87, 89-98, 100-104, 106-158, 160-166, 168-174; zqMVDm1H7FTXf48B 1-4.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: equip proc 1302348
- Rate: 4.15 PPM, measured

1,376 procs on 12,482 rolling hits, 47 wearers, 75 runs, on wand shots: all hits 4.13 [3.86, 4.44]. The rate is all
hits, which the lone Auto Shot rate agrees with. One-hand white: 0.00, 0 procs on 123 lone units, 20 wearers. Dual-wield
white, which sets no rate: 0.00, 0 procs on 2 lone units, 1 wearer. Lone Auto Shots: 3.88 [3.67, 4.14], 1,103 procs on
10,657 lone units, 47 wearers. Direct: window [-10, 100] ms, median 10 ms, 98th percentile 25 ms, 1,054 isolated hits.
Dropping the swapped stretches (5.4 active minutes, 0.4% of hits, in 1 run) moved the read from 4.12 to 4.13. Report
ZTGq13h7DH2bYrNw logs the same run as wR8KA9WTgHkD6n7q and counts once. Reports from 2026-11-11 to 2026-11-22.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Priest |      34 | 1,325 | 12,130 | 4.10 |
| Mage   |      13 |    50 |    352 | 5.33 |

Reports and fights: 1BLCZMxRfyYmHXNn 5-6, 8; 1VpHn3kYq6B8JMWw 1, 3-4, 6-7; 26AYbFNQzMZCmfkh 12-16; 2GczY83pjnCHdTNA 1-2;
3zwbTGcyQBLm7DMg 1-7; 4CrWmPVNyv7JcaFh 1-2; 4thjWV7Xbz2yrxBF 6; 4z2Yy39WAVFDtjN1 1-7; 6hDJQWTXwBr419AM 4;
81rnYqNT3djhJFRy 3-4; 8HvfKkArXpmbhwTF 1; 8YVKbmp1y4rMPFdW 2-11; 9Y3Xbk6D28ThMHKd 1; B4QqHnGXwzpRfaYW 3-4;
DcykPCwKQNzq7btT 1-4; DvrjQXqZ9HN8a74z 1-4; FmKYfTwCLcVzHMDt 3-34, 36-38; GAaNPKjbZkwLBCT2 2; GFBCh3VZH7TmAkL4 1-3;
Ht8VFJxmnCG9bQhM 1-3; JZLXKHRYna3QMND7 1-50; JvFPcBTz9rZMmW7h 1; KNtHCAPd97ZMTLBf 1-2; LN2Vk4nMtBc8vWbT 1-6;
Lb7Zcx2JtgjCTwP8 1; MPfbhA7yLxpkdtZQ 5-14; P6GLNM3n27dtcYJD 38; PDwtabFTz3VnQqrH 1; PKBd6Vz2Gkjwp9Zt 1-2;
QFqKVZgc6vAJfbrW 1-19, 21-28, 30-31; QXFd74zDW8JZMmKx 1-2, 4-13; Qqf8ZC1vRYhwG2cr 4-7; RJXQC67jbzATWdx1 1-2, 5-9, 11-16;
T7dtHBv9YR2jnLwg 1-7, 9-10, 12-19; TXGK16YwMmcyj9Cd 2-3, 5-10, 12-13, 15-17, 22-29, 31, 34-74, 76-86; V9ZpFqnkLrbaT6zy
1; VvFWzAgjYPxwN8fT 40 fights from 32 to 85; W2d6nPMQ9tV1BjZx 2-3; Wv4QYA6C7antxKXr 1; XD23HFnxbzvPjACV 1;
XrpxL4B8mPTWQkFK 1-3; YKqp6cRhg9zAfPJ8 1-2; ZTGq13h7DH2bYrNw 1-7, 9-13, 15; aCR8n63FzdTYGqj7 57 fights from 1 to 67;
aYBjphvLn8xMK2TN 1-5; bX3A9KBha2QrVWLq 1; c7M6PR9WHkyDJL2q 1-6; cav7VZrmPbNW3y29 13; f2CmVRcTnthPW6aH 1-14;
fM4nQ2tKmpyWBLXj 6-10, 12-24; fjtT3JVxbGPz6RwX 6; gBDMLv8J1K7nqkpd 112-157; gw8fpa1YVkv4WP32 1, 3; h4JXmPQDqWL61Cyc 2-3;
hYjVWRNBGDKLxTpc 1-3; hdVbcpYXGPDaFwyR 12-17; hfNjty7QRBrFCKm6 1-3; m3ctFRXJaZnrKzkC 1-3; mCn63gBZfLFHaWr9 6, 8-34,
36-39, 41-43, 46-53, 55-89; mLH64Yyjpbd3AKBW 1-4; mnPXjHcWNFKBhDCq 1-2; n8VK43MDGChq6rgv 3; p8WXHnkF6m4rvGJL 1;
pG8TNJwybLPYWmxF 2; pMkPWzaC7wf9XZNt 6-7; pqxm14QnyLk9tCGK 1-2; tG4CqkJQvyPT2A1Y 1-10; tHzN3kKbaPpVW8TY 1-8, 10-11;
tfVvJR3aZFgz2CXn 2-3; v637zq9CWKLxcpA1 3-6; wR8KA9WTgHkD6n7q 193 fights from 1 to 202; wkJXWxD23VMqThLg 14-17;
xb49gaHAVryMZh2t 2, 4; yW82C7hYvDmZkbc4 2-8; zA8JyGdnWRQ9vXmK 1.

## Naga Heartpiercer (3078)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1292647
- Rate: 2.1 PPM, observed

6,691 procs on 79,717 rolling hits, 115 wearers, 135 runs, in the ranged slot: all hits 2.10 [2.05, 2.15]. The proc's
flight reaches the 3,500 ms isolation, so no lone-unit rate is measured and none checks the all-hits rate. Observed: no
lone-unit rate checks it. Travel time: the proc lands a median 3,244 ms after its hit, 98th percentile 3,452 ms (33
isolated procs). Dropping the swapped stretches (257 active minutes, 9% of hits, in 10 runs) moved the read from 1.92 to
2.10. Report zaCRx8Zqvw6ytNKg logs the same run as L6ZKBgy2FQDAHmkh and counts once. Reports from 2026-11-12 to
2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Hunter  |     107 | 6,689 | 79,665 | 2.10 |
| Warrior |       8 |     2 |     52 | 0.96 |

Reports and fights: 1FR3hDpwGHbjBnaZ 1-2; 1RPdFq7DLMNBgCZv 1-3; 1WZ8Dn6mpQdCNxAT 1-8; 1a3pDtCcPgBkL6wm 1-22;
1fAj2G9rZaKDc6v4 1; 1fhNLGvKgMzt8Ac3 1, 3; 1g2QAZPJrLTjCz3m 1-10, 12-14; 1gKMDwaxfPv68FH9 3-4; 1h9P4rYzWDvtHRxb 3-4;
1hpgVjWH3GZBLYKw 1-12; 1x8aqzFcL7ydGNgP 1-2; 2AkBjF16x3rwVLXC 1; 2cRHnkQgwKBxW4td 1-8; 3YMxWqZy84TcGJCK 1-2;
4kVDpBQtGAhHf2Ya 1-22; 6KrwJ1zc7NQjMbAV 1-5; 8DrWyzRCtY7vjx6d 1; 8XrvwZJFBNqYGmRz 1-15; 8x3NPqTmnw6tZ9vY 1;
8xGjCkqAhgXcwtfn 1; 9BbfyPxjcJkw1tYZ 1-7; 9yNQP3FWHgcVvYDq 1-14; ARXJvZh6kwt3np2D 1-9; ARwG2htz3VgjDd7v 1, 3-5;
AZjCQxXdvK7tw8Dn 2, 4-11; AjZXVRHFpcmyJDgC 1-4; B3vtVgmqMTHDAxNY 1-10; B4vtTCL6D17pf32m 1-3; BKnf4rNwJdcqpQ18 1-6;
BYxchfkT1Z9n7HV6 1-7; BbhmWkK7MtnQ89FA 1-3; BjPdctn9FDTG6rfk 1-4; BzGXfjMLJ2RyCwHm 1-7; CDxVJLG1WzKgNkRt 1;
CTzARQh8F36B1VP2 1; CcAYgrB3Q1w8KJvV 1; Cj9YcDd1hqbn6rkH 1; Cp8467ZnH2hGJDtw 1; DT8MFahL6VCWQdNA 1-15; F8X4dznrj7Jx3HQB
1-17; FJdxQgAGwDYW6MTh 1-11; GcrbT1JP3DqQBhLz 1; GfmL9CHXdWzJnMFA 1-5; H8azRKMDv9Cq1Qnh 2, 4-6; HJpXmBNg9fjv46Dr 1;
KXmBtLjWnQ6zZa21 9-14; Kc4MgjhBWZbCAN18 1; Kzn8AkXFr3yvMZHQ 1-9; L6ZKBgy2FQDAHmkh 1, 3; MwA4Vhx7zDfCdY6B 1-2;
Na6RyCPmj1tGLBhJ 1-3; NvnJVbtwK8z7THyx 1-2; Q4KA1zq7WkTHndwV 1-2, 4-8; QFxf2YycX4K1G67q 1; QkHGb6P92mJ1z4aC 1-16;
RVTNDtGcvLjFzqnX 1-4; TBCkGK4pwZv2f6cF 1; Tby1rLwFvkgRtxpn 15-27, 29-35; Td48VjbADm17Gwxn 1-2; VHth6gR7xLbjQmnr 4-14;
VpFhqbrGwB8YTXAt 1; VtpFGaT74JDMbwCk 1-4; VzLTxpcwAFNMd7jB 1-14; XBYHq64fPFNJMhQ3 1-2, 4-6; XKAja9wcpndHyZRJ 6;
XpQW1hmwHYDAJ8at 1-12; YNHv8GBrbc91FQ3C 1-6; YX61J8CKqc42paNZ 1-4; YcBZxvJVt18WLnrA 1-7; Yd4r69MDqbQyJCFH 16-36;
ZjTwPR1D2nVyC7xF 1-2; aL6dYTxw2J4Bqp7y 1; aMNVKZpg8tRkhBP3 1-17; aMn8XG7qHNAcLjhR 1-7; aNFQk87zxVRKA2nf 1-6;
akfrAm74NYHVqKgP 1; avQywnDKt7RxcpfF 1-8; aykDWnCGx4jJFTV6 1-6; bPqj8zkJN3YatpwD 1-2; bZ1rRapKh7JcNy9V 1-18;
bdYgfJk7Mh8GqZ4t 1-11; cNKrmFQY7A6f4Dhk 1; cT71Vr9tPDKkxWv3 1-2; chaXJYpVCrPyTm1N 1; dFTYwhpJGLHv7NAC 1-4;
dn4pDGQNgTCW1j2c 1-2; dycN4Qm9ChBtjV7z 1-2, 5-6; fHQ16pCP3XWmaLVG 1; gBbwv2qZP3t4LmVr 1-5; gzwRHnTC8VBNmcZF 1-5;
h4azBt3LkrgMQ6bJ 1-2; hKXDFmMTBRL2JwGW 1; hKvmpgJarkVqCZBw 1; hPWNkH67AYCBFLgM 1, 3-8; hWdz149pDn6Tm2jL 1;
hXx2mb3t4VzFCNfT 1; hcGCaDW9tRYFB6Ax 1-11; hgd1qH4QYFyjz6AK 1, 3-6; hxCm1nBtF9MZXwWq 1; j8nHCBbFmfPXkL72 1;
jRbzHJP47WCLVN9p 1; jWyF3qd7t1baYRPH 1, 4-6; jbhGzfrnJRWX978L 1-10; ka9LWwzxJbmKdnAP 1, 3-6; m1LvMTRFb9KVgh86 1-3;
m2gBvZCnrRVq7aNA 1-2; pF7jaH96PnycBNqJ 1; pMyY4zGNtLb2gxAX 1-10; qJAMNGwP2aKB9jDC 1; qVY3amnxyzcPhG6t 1;
qa9Vb3rkCD87z4Nw 1-10; qfG6QWjd3wg1JVXv 14-18; qgNdpJ9WbY6FHM2Q 1-10; qm3nkapzVMGgBQJb 1; qr1CbQRVcL67Nx2D 1-2;
qt4GCyZr6f8wJ7WQ 1; r7cZyhBXWfNwj2GY 1-10; rJXdfcmW6NY4MFKp 1-2; rZ1A2XtkJjLfQb9x 1-4; rf6C1QL4XG9yKAHF 1-8;
t4d9ApkXac6nVBCN 1; vGNdPWLc23BqCQVT 1-2; vPfBGprgtTdcjZWN 1-15; vtJLdmHfPMGV7j4Z 4-13, 15; w48cXFqxMbWTH2ky 1-3;
wJWpK2Fqd7z8QnT3 1-11; wYmr1kfWPpVz7jyZ 5-13; wdNZmfpCMQL7D8Vk 1; wdRZpaVcAtDkGPNx 1-6; xJga9vbpNtGhKRWz 1;
xQv1aRGzXdpcKZjC 15; xyGFtfYArzgn6CqZ 1-5; y8PLB1VYqpnrb32G 1; yC6t8PJx9VRLrj4F 1; ypR78QbwvYJhrtM2 1-2;
zaCRx8Zqvw6ytNKg 1, 4.

## Night Reaver (1318)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13480
- Rate: 1.7 PPM, observed

164 procs on 1,748 rolling hits, 6 wearers, 6 runs, in the main hand: all hits 1.71 [1.46, 2.04]. Neither the one-hand
white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no lone-unit rate
checks it; the interval is wider than 10% of the rate. One-hand white: 2.30 [1.57, 7.27], 27 procs on 213 lone units, 5
wearers. Specials: 2.38 [1.36, 3.87], 16 procs on 122 lone units, 2 wearers (a Poisson interval: under 5 wearers).
Travel time: the proc lands a median 137 ms after its hit, 98th percentile 613 ms, window [-20, 623] ms (34 isolated
hits). The paladin check, which sets no rate: lone white swings 1.70, 13 procs on 139 lone units, 6 wearers. Reports
from 2026-11-13 to 2026-11-19.

| Class   | Wearers | Procs | Hits |  PPM |
| ------- | ------: | ----: | ---: | ---: |
| Warrior |       2 |    91 |  860 | 1.92 |
| Shaman  |       4 |    73 |  888 | 1.49 |

Reports and fights: CaX6DnMGJd2Fgyqj 1; Mh8TkfRHy4w9Q67v 1; T72D1mwvJnabPtqK 1-3; ac64wQx7yZ9tRNfW 1-3; cZpT18gfxAPQ4CKY
1-5, 7-8; kz7AYv9wNjQDVKGH 1.

## Nightblade (1982)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 18211
- Rate: 1.8 PPM, observed

373 procs on 3,512 rolling hits, 5 wearers, 7 runs, in the main hand: all hits 1.82 [1.71, 2.01]. The rate is all hits,
which the one-hand white rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand white: 1.98
[1.75, 3.45], 111 procs on 961 lone units, 5 wearers. Specials: 2.01 [0.52, 2.26], 45 procs on 383 lone units, 5
wearers. Two-roll abilities, priced as one roll: 1.04, 2 procs on 33 lone units (Execute 1 on 6 and Mongoose Bite 1 on
27), left out of the rate. Travel time: the proc lands a median 140 ms after its hit, 98th percentile 315 ms, window
[-20, 325] ms (166 isolated hits). The paladin check, which sets no rate: lone white swings 2.33, 290 procs on 2,135
lone units, 38 wearers. The owner takes this paladin-free count for the row (2026-11-22). Reports from 2026-11-14 to
2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |       2 |   276 | 2,665 | 1.78 |
| Warrior |       3 |    97 |   847 | 1.96 |

Reports and fights: 6fdmqNGFJjbRgCcD 9-10, 12-23; FzRDWkfmyh47tTAV 1-4; HQZwFjDcmtnWyXMh 1-5; b196ArgRdLNX42ym 1-2;
m1JkqnD4gjfLwAvP 1; nNmBkKp3J96FTGfg 1; vdwBDjmZRykqHaWL 7.

## Pendulum of Doom (9425)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 10373
- Rate: 1 PPM, measured

1,119 procs on 17,111 rolling hits, 15 wearers, 17 runs, in the main hand: all hits 0.98 [0.93, 1.03]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.08 [0.95, 1.19], 317 procs on 4,383 lone
units, 15 wearers. Specials: 1.12 [0.77, 1.38], 108 procs on 1,447 lone units, 6 wearers; Rend landings give 60 procs on
785 of them. Two-roll abilities, priced as one roll: 1.78, 18 procs on 152 lone units (Execute 18 on 151 and Mocking
Blow 0 on 1), left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 443 isolated hits.
Dropping the swapped stretches (19 active minutes, 3% of hits, in 3 runs) moved the read from 0.95 to 0.98. The paladin
check, which sets no rate: lone white swings 1.20, 42 procs on 527 lone units, 4 wearers. Reports from 2026-11-13 to
2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   573 | 8,757 | 0.98 |
| Shaman  |      10 |   546 | 8,354 | 0.98 |

Reports and fights: 9pMcvgxrQNfLZ7hd 1-15; AdD17hyxFtqfrQmC 1-4; AmWdCznPR1ZxD3yV 1-10; Pjz8kK2XvxyQrTCB 4-16;
TQXGy2aJNdVvwCK7 1-9, 11; WBfhVMYkvLHmPQjx 6; X186Gx2gRBZChbVA 1-7; ga9Qk6DcjJmNTGVH 2-3; jC6FzGLJ2yn3kvw9 1-3;
qw3ZzkBmQYW7pXVH 1-19; t6dr1MY8zTJBWnkg 1-5, 7; tcgDCfvmA1aknRZq 1; tkDRJpKXPf63gvdL 1; w9jRArgMBvnYt7JX 1;
wTakYgvxcBQV9qRr 1-19; zXAnmG7MqPwBcyjK 1; zc1hygaCXjDAPxL6 1-2.

## Phantom Blade (7961)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 9806
- Rate: 2 PPM, measured

2,428 procs on 45,243 rolling hits, 44 wearers, 46 runs, in the main hand or the off hand: all hits 2.01 [1.92, 2.11].
The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 2.23 [1.82, 2.70], 105
procs on 1,766 lone units, 3 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.94
[1.82, 2.06], 940 procs on 18,148 lone units, 39 wearers. Specials: 1.64 [1.12, 2.11], 75 procs on 1,716 lone units, 5
wearers; Sinister Strike gives 26 procs on 621 of them. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms,
794 isolated hits. The paladin check, which sets no rate: lone white swings 1.73, 36 procs on 779 lone units, 6 wearers.
Reports from 2026-11-15 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      29 | 1,351 | 26,137 | 1.94 |
| Warrior |      10 |   854 | 14,827 | 2.16 |
| Hunter  |       5 |   223 |  4,278 | 1.95 |

Reports and fights: 2Gt18qjbgwPWzDTn 1-4; 2YfdbFMgGhr7zHwQ 3-4; 3hpaqBV4jbGdx2TX 12-13; 6kN78G2xfqy3vwTP 1-6;
7pZAGdWVxcF26wzR 1-6; 8L9G1c3bKt6BAxJW 1-10, 12; 9qQ1fFjyT2RHxhZA 1-2; 9wtfbKA27FNQ6BaW 1; Ap2fHzmtJLhv83DN 2-15;
CaWgXwKYjN2Qx9J8 1-5; D2d3hNAbwkPgL1vF 1-10; DX7BTrcHpdYgJtf2 1-5; F6pGhnJ7CrcKADH1 1-2; Gtpz8WLFQyfgxA7Y 1;
GzZH1WhTLQ3JaDXt 1-11; JtKmQZNYjhrHac6R 1; KXwCgWRbN49fn3Fz 1-8; PHDt27hA391NBRXM 1, 3-4; QAHXb9zy4jfgKPVJ 1-6;
RYXfhy4WLFgaCdN8 1; TvGqg219RpfFPZ3J 1; W37LQ1frPVgXmvzG 1-13, 15-20; XFdcWaPL8MthYH6A 1-5; XkhC2YtZ6cMBHGfz 1-2, 4-6;
Yzv1c7dBnLNhyXQT 1-9; amc3nxRgHhjyrqVb 1-9; bRXtBrFwDj7d4vHP 1, 3-9; bfAYmtpFLQ41dzy2 1; bw63h1prGDWZxyz9 1;
cRWDFxQykq47t9pV 1; cRtHw9DWZMrKnyvQ 1-4; cW2DhmMGFw34qnbP 8-10; ckBpjJrP48nQqKWy 1-4; dt4M3ypRnTkwaVQY 1, 3-10;
dt78h2VnRxJyqFXC 1-3; hrV3QgyzvfdZKJwF 1-30; kQ2gDVr4h7NPm1pa 1; pLtJjx8DCPbHRacv 1-21; q4HLhfGjYbWndJv1 1-5;
qL7cBTb6vKjXtpWJ 2; twcjmZNGMQP2FABR 1-12; w2cjN3nDL7Xrk8F4 1-4; wWr4kmdgv6B7CYcQ 1-2, 4-5; y2dFcNR1rMmKTPAj 1-25;
ynGzYZ3cLrpvwj1N 1-9; yqFPH1MKv746dQhB 4-7.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 29646
- Rate: 2.85 PPM, observed

1,604 procs on 12,028 rolling hits, 16 wearers, 16 runs, in the ranged slot: all hits 2.86 [2.74, 2.96]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it; every wearer is a hunter. Lone Auto Shots: 2.98 [2.78, 3.22], 331 procs on 2,384 lone units, 15 wearers.
Specials: 2.18 [1.43, 3.02], 29 procs on 285 lone units, 16 wearers. Travel time: the proc lands a median 603 ms after
its hit, 98th percentile 1003 ms, window [-20, 1013] ms (223 isolated hits). Dropping the swapped stretches (9.1 active
minutes, 2% of hits, in 1 run) moved the read from 2.79 to 2.86. Report VQCRJWfq3KX7M1Nh logs the same run as
rHM6TPA7NVckxQmh and counts once. Reports from 2026-11-13 to 2026-11-21.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Hunter |      16 | 1,604 | 12,028 | 2.86 |

Reports and fights: 1QgdaxJ2GBpmvyjL 1-2, 4-8; 38TMrdKDYzwc1nvZ 1-2; 4bqYVtkjL6aKdwQM 1-2; ByRxHQmnzfhbvtcV 8-24;
CpABMyt29JkdwjZf 1-20; FgMt3hxzyVBJp7bW 1-3, 5-7; Hrt2wgZkdjTRBC4A 12-13, 15-20; LkKn1mhzPwDJBVHC 1-6, 8, 10-15;
MfXjTPdwar9gyKvW 1-2; VQCRJWfq3KX7M1Nh 1-2; XDjZrbFMgfC48a19 1-2, 5-6; YyVjhcz81AHad3X4 1-3; hyzdP12689vwfrj3 1-11;
pbhgCm6PjYwZ3fRd 1-3; q167YKJGNhjpZcCz 1-2, 4-7; rHM6TPA7NVckxQmh 506 fights from 1 to 648; tQBgTykjJaq2xbXh 1-2.

## Rivenspike (13286)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 17315
- Rate: 1.95 PPM, observed

11,224 procs on 119,175 rolling hits, 73 wearers, 87 runs, in the main hand or the off hand: all hits 1.95 [1.82, 2.07].
Puncture Armor stacks on the target, and a refresh at full stacks can log nothing, so no lone-unit rate checks the
all-hits rate. Observed: no lone-unit rate checks it. One-hand white: 1.62 [1.31, 1.98], 94 procs on 1,204 lone units, 4
wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.24 [1.11, 1.37], 1,215 procs on
20,319 lone units, 68 wearers. Specials: 1.61 [1.50, 1.71], 1,497 procs on 19,290 lone units, 66 wearers; Sinister
Strike gives 825 procs on 10,993 of them. Two-roll abilities, priced as one roll: 1.96, 73 procs on 771 lone units
(Execute 73 on 747 and Mongoose Bite 0 on 24), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 2,130 isolated hits. Dropping the swapped stretches (142 active minutes, 3% of hits, in 10 runs) moved
the read from 1.88 to 1.95. Report vwYF4Z2C38cfKTDB logs the same run as CfHvwytKJ6P4q7ag and counts once. 2 runs are
dropped whole as swapped. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      42 | 7,071 | 71,171 | 2.06 |
| Warrior |      28 | 3,997 | 46,342 | 1.78 |
| Hunter  |       3 |   155 |  1,662 | 1.93 |

Reports and fights: 1DZt4gPHfpTCk9W6 1-4; 2LHApJKVNWybCqv6 1-2; 3MhvkdyQAt7bVpWJ 10-12; 3RB9X8HVxgCJ4My2 1-4;
4azX71FvCjQgmJkc 1-2; 6hDJQWTXwBr419AM 1-4; 6vaAyCJMWf4wHdN3 1-2; 87GwCAY4MtTg6aLB 1-2; 92qYgyHrGKjfx74n 1;
9jRHk463VwPQv7gr 1-2; 9nX7rz3h2bvPJBaH 1-2; AdfMXDLW29k6pGy1 1-3; AyTmpNtxB9jbnJ2H 10-45; B8JHDYfAhtCRbFLc 1-2;
BHvgJVTdZ9pfNPa6 1-3; C18pcNtwj4aRDbQk 1-15; CfHvwytKJ6P4q7ag 1-24; CqM4Vnh831tFgcYw 1-2; D2nQcj6wFbmBgVyJ 1-37, 39;
DWfMg7hzPF4d1mQ6 1-11, 13-52; FjfYbn9X3LgqydCk 1; Ft1rJVHX3kz7yLBZ 1-3; FtcXATwrHDd9vQ83 3-4; GWMx4hcJFCYHR6dk 1;
GnjD4gZfcd1F8RVh 1-47; Gzb3PZXB2FD1agYC 1; JRbYcyjAQ8Pvtdqp 1; JdYhqnzHRvL7t8ky 1; JtL37yw9AYNGWHQf 5-7;
KxrXLAQ2bfkc8dBF 8-14, 16-55; LV3Ny1CfK78Hw2aW 1-3; Lg8a3v6179fZAYRC 1-12; N3nHF2GBgdT6qx1V 4; N9PGBMt7faY6W2J8 10-12,
14-45; NXMZV9xjvWpB7GtP 20-107, 109-116, 118-122; NpVkX1JHYydCcjDT 1-2; PpTcnbZ4zqtkmgNf 6-37, 39-51; PwKXDAqJ4T7t8R9v
1-3; PwMavYLXV8GztqFd 3; Pxn7LYfMX4Qd6vNr 3-13; PzGdkxLV1rMXynpF 1-3; QHXdBJRj2GPpC3n1 4-6; QmbxcT8KPajn7ZFp 1-5;
R3j6zmXnDByHWJVr 1-2; RHgxPaQbVWjDTc4K 1-7; RP9bjvBAqThyndCc 2-3; T6VDdFhPxQ43pHM2 1-2; TB6KHhwxyPnDQctg 8-13;
TRn3gPfqD6Bj2FXz 1-2; Ttp9X82Vz1QBxGAD 2; V19NAGr7g26JymZ4 1; VzpjcgvbnLCaTZG2 1; WqfyDhVHxT93Q7Pt 9-30;
XyJPKWtCdNkVL4gB 1-3; YaF9bGvxM8NyDqkj 1; ZRhnvMANGJDy1kBj 1; ZmCAR8bDny9kpKJ6 1; aRLP2JdQntKzTx3f 1-2; b9B4aC82PXYypWLq
2-5, 7; dK9TAcBfjVmD2n1J 10-18; fM4nQ2tKmpyWBLXj 18-24; hQY9xzt7Mr4nG6Ap 1-2; hgv7XY4tJKbTWPzH 1-7, 9-16;
kNwn6HXKVhCByZr3 1-2; kaH1G4ztmYpgMTJ6 1-2; m6htAf2g7FvD3WVj 7-18; m76CKntQZWLYhTMg 1-6; mAW1KFR9YP8r3bNx 1-2;
mfA6MCQkPbFdRrz4 55-58; q6379W8YDwVcZ4Rp 1-2; q98JvNkWnKzyrwgM 1-6; qWhzDXTB1nKxk9JH 41-46; v8XzFwL9pbHhBCDg 1-7;
vGXa8D1FPyqTH7ZM 1; vHVNaQRrMdJmPpBC 1-3; vYjL3q712CyTGkQh 4-10, 12-15; vny4RK9TkMzwtAQV 1-5; vpq9gA6dfFB1CJjr 1-3;
vwYF4Z2C38cfKTDB 1-9; w4fzHNVk2dAhLKnv 1, 3; wRWZtcVCk3QJKDa6 1, 3; wtR76PQ8nxpANYkH 1-2; xt7BwRzWVvbKJTXp 42-44;
z9wcdxZaRLhWk1Qf 1-3; zFRnjkryLtYWVKqJ 1.

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

- Count: §Count of the 2026-11-13 to 2026-11-26 reports
- Proc: weapon proc 18138
- Rate: 1 PPM, observed

4,339 procs on 142,069 rolling hits, 84 wearers, 114 runs, in the main hand or the off hand: all hits 1.02 [0.99, 1.04].
The rate is all hits, which the specials rate agrees with. One-hand white: none. Dual-wield white, which sets no rate:
0.98 [0.91, 1.05], 699 procs on 23,811 lone units, 81 wearers. Specials: 1.02 [0.79, 1.35], 50 procs on 1,638 lone
units, 37 wearers. No paladin wears it in the count. Travel time: the proc lands a median 140 ms after its hit, 98th
percentile 300 ms, window [-20, 310] ms (482 isolated hits); lone units are isolated on the hit's own target. Dropping
the swapped stretches (155 active minutes, 3% of hits, in 18 runs) moved the read from 0.99 to 1.02. One run logged in
two reports counts once. Reports from 2026-11-13 to 2026-11-26.

Across the 2026-11-23 hotfix: the reports before 2026-11-24 01:00 UTC read 1.04 [1.01, 1.07], 3,567 procs on 114,838
rolling hits, 71 wearers, 94 runs; the reports from then on 0.95 [0.87, 1.01], 772 procs on 27,230 rolling hits, 17
wearers, 20 runs. The analyzer run on the reports from then on alone reads all hits 0.94 [0.85, 1.02], 772 procs, 17
wearers, which overlaps the whole count's 1.02 [0.99, 1.04]. The rate is the whole count's, observed by owner ruling of
2026-11-26 because the two halves read apart.

| Class   | Wearers | Procs |    Hits |  PPM |
| ------- | ------: | ----: | ------: | ---: |
| Rogue   |      71 | 4,012 | 131,659 | 1.02 |
| Warrior |       5 |   224 |   7,168 | 1.04 |
| Hunter  |       8 |   103 |   3,242 | 1.06 |

Reports and fights: 17bVXjMKCRqz3nFH 1-6; 2FXvGj386wa1ptHA 1-6, 8-12; 3GT4tz8gxnphcbkD 1-7; 3HJrg1BGh8xR6Y2z 1-13;
3qfDVTZ1J29zPChn 1-23; 42wYBypCXH1jVxZM 1-5; 43R1XHKpfLgybZqm 1-4; 72q1FtWgTMHAcLxG 1, 3-5; 7fz9Kk8rY6jyQAbL 6-10;
8JTGAcnFYvtmjLN7 1-4; 8TRwjBNPFDMqV6my 1; 9T1NCP4vxBztFYZH 1-6; 9fvqtGwK7RgT31xF 1; ATwq9YWpNbLC6Khc 1-10, 12-17, 19-20;
Ba4bzVmX2nDrRN7k 1-3; CDLFKa2Nw4td3rZk 1-4; D1HnX9MbVNaLRKBq 1; D46Xm8ncR9WbTFrL 32-71; DALjdvgcWTYBF283 7, 9-40, 42-52;
DCzyMhKTZQ98rmFw 1; DbQwz3av89BpZYfn 1-27; F8yVqYNM9HQC1aLn 1-6; F9bgjZYDQ6fqPaLk 10-21; FBLTNkwYgtPzjyfh 1;
FK2TQv9HJyhmLjBa 1-2; FrL4wayTHRtW1CKx 2-15; GB6fVLkvzPhxYJcZ 1-2; GmkC6RtH12vawYTL 1-5; J8YH1pCt7nWbcQha 1-11;
JbKBmGHwCcyVzkjt 1-6, 8-24, 26-29, 31-33, 35-40, 42-52, 54-59, 61-66; L6XGbAdMqBc19KQ7 1-9; LW1pRgPwCmXBqk9a 1-11,
13-27; LX4B8yPtVnhxDJvq 1; MDmqgxJC689cwQdf 1; NFcRw6HdA7g1BGm2 1; NTJzVdMZBQDaqG83 7-8; NjWKPpHgymCT832A 1-19;
Nxtj7m6zcFX9Aq3Z 5; PJpRmw7bf8WvDxA9 1-3; PQzCL76fmZF1TyN9 8, 10-12; PxgRcqaK9bdWjTGQ 2-5; PyFHbGW3t8gfXMrQ 3-14;
QHqPwymAX4rLv6at 1-2; QymjLY4XrbCJxKcP 1-2; R792cjGCzgx4NaTM 1-2; RWjNGhpVvJ36c74n 3; Rcn2dBw6T7vZFbtx 42-45;
T8JKMZh1WwyazjHY 2; TKDJLt2ABVFw8n9h 2; TWrFLYwN6tCK9fdm 2, 4-6; TdH1RpCqkK7BWw2c 1-2; TgQJ7t3wHrzPqM4C 2;
V7rWM1fCGc9jHQKt 1-3; VKkWz4Ar72pR1LyB 3; WrP34pNDfdyzhb2H 1-14; Xjrt9FRMWqJyKdpa 2-3; XwT7QqZVhp9gbmzc 1-2;
YGJ7Kv4m2gXHQrBf 1; YVMcL1DBkbNC9hxH 5-40; Ya9kKCWRwqjFpgz1 1-4; Zj2mcQnvKFPBLgCk 1; b6YgFwaVqjc7kyrP 1-3;
bWH41fw8CDQAch3K 6; bnMvHkFjxQLJ4VRT 5; br4AMRdq1Wv7jpcB 1-3; byzq9hxPvpDAJac2 1-12; cYXtDBjJZwKQzW2f 1;
d9VRhFPmDGkLH4vf 1-5; dnDg7j9kRw6WtbML 64-69; dtqaDCLhXYb6BM1r 1-4; f32xyDMjJvAdBHa9 3-4; f6Q2vmDkJ7YhAy9H 1;
fNZyXdzPRxhkjLcg 1-2; fcghxMTW4VXFQNHk 13-24, 26-53; fcnqgCZtXyY34Gbm 1; gRWbpVBwcJxPjKZ1 4; hAfKnz3MpdvNjQxk 1-5;
hy3mwRF9GAcfzJ4P 1-2; j3dqRfHCL86cN4Fg 1-2; jgL6xTGHDt19CvNF 2-4; jvqAx3thp6wTz9ZR 1; kxQJCbtq9wLHmjKR 1-6;
kxt1vq6DGM7YPA8n 1; m3ctFRXJaZnrKzkC 1-3; mDfBWP3bv1HtzChw 6-17, 19, 21, 23-25; mQAvYq6HaW2GJZKb 1-3; mdP46qvaFWALKM9B
1-2; n19QyaLzWJtCM23d 1-2; nM8ZvpbrDqNzFACd 5-7, 10-18; pB26nafTCqtMHPcD 1-2; pb3cw9Q62mzh7tWM 1-3; qKtgYkmZMVjnLBfr 1;
qXPhR7Jtd8jKTB9p 1-3; qmfjQ63TN81AnGkB 1-3; qn1cjM6CxJHyp7dr 1; qyDxWPJrCtGzQBcv 1; rf7pqCmzaxAJ9yVD 1-2;
rvPn71AMTwajKNdG 1; ry41NG3KdBnRzCAw 1-5; t3b8rm9h7CNXag26 1-9; tBgZqFKhaArDmw4d 1; twbKV4fgDyW3XHmn 1-2;
txBCajcb1MVJ6GPg 1-2; vHp9QagZGFLBM7PC 1-2; vyQrDZN7RX82ApnP 1-3; wXRCbqdLt6mAg8aV 2-3; wafLQr31X6VYCmNZ 1-3;
wkrzY3mDh7RGfMc2 54; xAHXJLYPCFcb7a4d 3; xRkgQCj3TD2yb6Fn 1-25; xdK1Pac8CynkFDYB 2-4; ycMCt3n9ZqP2bhaH 1-2, 4-8;
zJAHqrM1h6GKXBcm 1-2.

## Shortsword of Vengeance (754)

- Count: §Count of the 2026-11-09 to 2026-11-22 reports
- Proc: weapon proc 13519
- Rate: 1 PPM, measured

688 procs on 17,207 rolling hits, 16 wearers, 17 runs, in the main hand or the off hand: all hits 1.00 [0.93, 1.07]. The
rate is all hits, which the specials rate agrees with. One-hand white: 0.70 [0.33, 1.28], 10 procs on 359 lone units, 2
wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 0.78 [0.62, 0.95], 116 procs on
3,714 lone units, 14 wearers. Specials: 1.18 [1.02, 1.47], 102 procs on 2,152 lone units, 13 wearers; Sinister Strike
gives 65 procs on 1,449 of them. Two-roll abilities, priced as one roll: 1.50, 3 procs on 50 lone units (Execute 3 on 47
and Mongoose Bite 0 on 3), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 154
isolated hits. Dropping the swapped stretches (5.5 active minutes, 1% of hits, in 1 run) moved the read from 0.99 to
1.00. Reports from 2026-11-13 to 2026-11-22.

With the off-hand wearers' 129 Shanks from 2026-11-20 counted as rolls (EV-0155 §Shank from 2026-11-20), all hits read
0.99 [0.93, 1.06], the same rate rounded.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       7 |   370 | 8,699 | 1.06 |
| Warrior |       8 |   278 | 7,527 | 0.92 |
| Hunter  |       1 |    40 |   981 | 1.02 |

Reports and fights: 2AvFk4PJmBhyb89n 1; 3T8GrWJRdBkvy4gz 1-6; 6fvmrHP9dzy4Cp7b 3; 6zHgax7vFKqLrdmZ 1-2, 4, 6-8;
JKavrBGqLD6hVcHZ 50-59; NhHDL86y4PFrCW3G 1-2, 4-11, 13; a8R6AvpVJrH3XNCj 1; aNBtZyVqKMwfp4mL 1-5; akLZWv4X9VhHmRxz 1;
gDtFwpGPLM2TyZKB 1-6, 8-9; hY2gt1GvZjnwW6N8 1-4; kv32zprHjRfnZPh9 4-17; rjtG2MQYfqxJ4NcL 1-3; rp72ZYxmG68tMJ3Q 1;
tYFp32ThgGPfnvQH 1, 3; yw6Hzd1mphGk8vTJ 1-2; zX7thnkw1pR6WFyJ 1-21.

## Skeletal Club (2256)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13440
- Rate: 0.95 PPM, observed

1,072 procs on 26,629 rolling hits, 38 wearers, 42 runs, in the main hand: all hits 0.93 [0.89, 0.97]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it. One-hand white: 0.91 [0.77, 1.07], 153 procs on 3,875 lone units, 21 wearers. Dual-wield white, which
sets no rate: 0.92 [0.75, 1.08], 66 procs on 1,653 lone units, 17 wearers. Specials: 0.98 [0.87, 1.07], 151 procs on
3,552 lone units, 31 wearers; Sinister Strike gives 42 procs on 1,027 of them. Travel time: the proc lands a median 155
ms after its hit, 98th percentile 429 ms, window [-20, 439] ms (205 isolated hits). Dropping the swapped stretches (83
active minutes, 7% of hits, in 4 runs) moved the read from 0.87 to 0.93. The paladin check, which sets no rate: lone
white swings 0.71, 18 procs on 585 lone units, 18 wearers. Left out: 2 of 44 runs in Cat or Bear Form. Reports from
2026-11-11 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      16 |   545 | 13,508 | 0.93 |
| Rogue   |      15 |   443 | 10,909 | 0.94 |
| Shaman  |       6 |    84 |  2,204 | 0.88 |
| Priest  |       1 |     0 |      8 | 0.00 |

Reports and fights: 2TbWxFvzYy8gGVcj 1; 4tb63CTmMX2zxLDv 1-3; 6jYwBg18avQPmrbM 1-15; 9GAVBJNWDbmZk7tq 1-3;
9qTHdZQfvhCVP7RB 1; AQd8hCMn7a4ymGBW 1, 3-6; Dx2tAm7YawZjV6hG 1-4; F4LZcWy7VxjRfgrA 2-6; F4rpyvWbxgfZCBXa 1;
G7BvNCDbc12nKVpJ 1-57; GxFyfYrW31nbPpz7 1-5; H9XFTGLv8abwZdQ1 1-2; Ht2nzAM7dFbvY1wR 1-2; K4tq2TBgvb9R8wQm 1;
Mj1gA4rpDRQh9yJm 1-3; NDHgq2hvyaTMfBm7 1-15; P143Tb2X9rBWwQDq 1-18; PrN7Lagkfj6XA8vy 24; Qb7TJqYkKtgyz1ah 1;
QzNpGgjCMVxYR7BL 1-3; RnD2ych3aP6ZGXmV 1, 3; Rq7gBc4JQfGHAybC 1-5; RxcQmkVb8ZaACNw1 1; T93NgKpBfkx1qCLv 1-2;
WCFJ4kYHVXxLhwAZ 1-3; WFNrxP3XqbTBp427 6-7; Xd8v4PHrfjwDMxZq 1-3, 5; XfKMJj1BcQwP27zp 1-10; bdRPn6tTkhvQjBV7 1-9;
cj32G8gw9pvAZdxk 1-3; dLGpZ3YD1VfJTRWg 1-11; dRk86B4CyYw3p2qD 1; dT4RAM6tkfaGJmBP 1-5; dp9KVxZRnhAqP62F 1-15;
gnmCNT21xGd4q9AR 1-2; mF6k8yjVYKRtQ4zc 1; ndV9xfYhcFDaRwqG 1-3; pJxh4FZT9qgGVbXr 1-2; pydvC4YzqVMRgPAf 1-3;
rWPnwxALC4XQ1KVg 1; wJaf6LdbHcgR9Pxn 1-7; ycZLWXxQ4jk21w3q 13-28.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293714
- Rate: 1 PPM, measured

6,511 procs on 166,819 rolling hits, 201 wearers, 291 runs, in the ranged slot: all hits 0.98 [0.95, 1.00]. Its
lone-unit rates read on a travel window past 400 ms, so they check nothing; a separate count of its lone Auto Shots on
the reports up to 2026-11-19, 0.94 [0.90, 0.97], agrees with all hits and checks the rate. Lone Auto Shots: 0.96 [0.90,
1.02], 1,267 procs on 32,963 lone units, 195 wearers. Specials: 0.96 [0.90, 1.03], 841 procs on 21,842 lone units, 169
wearers; Multi-Shot gives 619 procs on 16,452 of them. Travel time: the proc lands a median 522 ms after its hit, 98th
percentile 1011 ms, window [-20, 1021] ms (1,429 isolated hits). Dropping the swapped stretches (219 active minutes, 4%
of hits, in 20 runs) moved the read from 0.94 to 0.98. 2 runs logged in two reports count once. Reports from 2026-11-09
to 2026-11-22.

| Class   | Wearers | Procs |    Hits |  PPM |
| ------- | ------: | ----: | ------: | ---: |
| Hunter  |     171 | 6,497 | 166,600 | 0.97 |
| Warrior |      19 |     9 |     175 | 1.29 |
| Rogue   |      11 |     3 |      44 | 1.70 |

Reports and fights: 1DVFRby49Pg2tpzG 1-19; 1KF2jNqTzc6VJ9fp 6-8; 1VCHKmNy8FgrLnAh 1-5; 1Y2XNrVgGfKvpWMj 1;
1fHnGhFabmPTq6Dv 7-17; 1hpF2AyzjRdrL78N 1, 3-12; 1jgTLvJQBMP6kxrm 1-12; 1qpQ6FW8fxD9gC4G 1-7, 9-11, 13-27;
1w9jNgyZv7YKAacb 1-13; 1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 3, 6-9, 11-15, 19-28, 30; 2WH4jLt3FADy8fnh 1-2, 5-6;
2aG3Ld8KVARCzrh1 1-7; 2x16Rt4gjcCFf3Gp 1-2; 38nNdjbYhAkpFagw 1-16; 3LhpDbvyrtNFWdxH 1-5; 3PaV6RHdnzw7Jpk8 1;
3X2paN4GkAdnLHj6 1-8; 3fzBnVRCTw1gjQGL 1; 3pCKd6jqgHyFc9kf 1-17; 3t2jQVPJFX48xCBm 1-3; 3vn6yQxN8AWmpKaM 1;
4BhykDjm2p1ncLGK 10-11; 4Jydm97DWFHARgPt 1-12, 14, 16; 4KWV2rZDJMqNLRmT 1-3, 6-7; 4MJWTYdRLrtPjDN1 1; 4NMcqLwQ2rB9amCH
1-3; 4XPJN7RDcM8BknFY 1, 3-8; 4hcXbHqV9dNDf1pt 1-2; 6Hpk7mLqw8YTFPGb 1; 6VJxcqnYa4wdBWQv 1-9; 6f8CLZy7VPMbqGKW 1-5, 7;
6fdmqNGFJjbRgCcD 9-10, 12-23; 6gvfVjrpZR8mcYxk 1; 6kN78G2xfqy3vwTP 1-6; 84VDXdcKZhHMPBAm 1-5; 8CzwkcFWrA91QaxG 1-12;
8R1c427qCbhPKgzx 1-3; 8RbAjm92ykw1pdCG 1-3; 8WgnxAakbzDZdFBX 2-9, 11-17, 19-20, 22, 24-25, 27-28, 30-31;
8YQCGy97hpdx2Rk1 2; 8bHBkarQFDqJPVAy 1-2; 8frHF64GzAhBxDyb 13-14, 16-22; 8rNkjZ2B6WdzFcKq 1-3, 5-7; 94fVcDCNBLztXh1k
1-12; 97FVRX1WLHqDmjAQ 5-12; 9DVWQNAfHXJBLy61 1-2; 9FKhWLbqTnXZpRD6 1-5; 9RrTvMpkamJPVw7g 4-15; 9g7NQBqDH6FGfP8L 1-4;
9nXALZFD462NGHTV 1-2; 9zfKMnRxtJGZQXBL 1-4; A8tQY9p3Dzn6TLRJ 1; Aa4fnqbyzhM8Xrpw 1; BTRMkxmvfrwcZNaG 3; Bmk9chwYjd2F3zpZ
1, 3-6; BrYJ1XhanMtCRjbD 1; C97aQyTJ4fPcrmjt 1; CLXHZ3c2GyN6mYBD 1-4; CR4DHc2wFgjd7yrp 11-13; CXvjmgaJ6YqKzD4k 1-2;
Ccvk1yDbBzFZAVqG 1-2, 5-9; CdD48cWzqZGQYmp6 1-4; D8JF7rbYQRcWxqn3 1-2, 4-8; DTQLfhZR9WGzBKjY 1; DWJhjnmAM6GZdbKv 1-2;
DaH6LCWQbmjXAJFh 1-5; DtAhCGwd3cmQ4P2g 1; DyYJRTv2XBfGcwC1 1; F6KYtxXhp1VHRbd2 1-5; FDybHvQ3pPqBracN 1-4;
FTXVdBrHJZD9aw7P 1-7, 9-23; FbkAhDQcP4KgpXGw 1-2; Fwcmz4r9aDnpRWAd 1-2; FzRDWkfmyh47tTAV 1-4; GALhdpX84abJWPFm 1-4;
GKZJBtWAvfPL81bc 1; GXmnxzAfNabcMB6Z 1; GZzaJH1QVjMKd8qA 23-34; GfwAMF6LZDm7TVWz 1, 3; GqLag3z6myTYrNb2 1-6;
Gv6NDzAmFjC4P3qL 5-15; Gyw2T3LjptRW6vH9 1-6, 8-10; H6hnfmvc39wqtC1D 1; H8bX3GacPBh2ARCL 1-2; H9FXjaBWRpYVMq4r 1;
HARvQ2hJWp8f3grD 4-9; HLrCnVfNMgYAJ1jd 1; HN7bnjM6mrKWqDRT 1, 3-4; HRFkxQLftWwcnzGq 1-2; HjBPLYwhmZprMQ12 1, 3;
HnC7N2YVAdg3wtrk 1-4; Hp9ThMznR74PwYKb 5-8; J2XL6AyZqgpjx8QK 8-13; JZNGm2rgxywWARPB 1-8; Jh7F2XTzVnL8Hyw1 1-2;
K6rZRbwYD1Wyjhpx 1-4; K7T21rXDGxcWabtQ 1-3; KCYXdTfNgGHczbLw 1-6; KLNhgMTDAV6xY1Wj 1-3; KpWJMkH19dP8RGmb 1-2;
Ky6k1d7mP2zVXWpt 1; KyYwHQjxAL8tzvdJ 1-7; LKkr1pxAtT7CvcZN 2-17; M4CvXLZz3qm6pDRc 1-3, 5; M7BnDF4VPXcWRaZ1 1-5;
MBg48hvnAyq2NDfx 1-2; MNZwjf63JzdrYgTC 1-3; MPkVah87LYAcnFdZ 1; MgcQ6txHr9R1VjPv 1-5; MkrFhTq2HPnzmCVJ 3;
MnX2ZVD8bxTFfjdG 1-7; MpcVyhTHnqCJvDFx 1-5; MtxT78RD4NkCJd1q 1-3, 5-7; MyzjakDC9gBbpq6K 2, 4; NFtR7vncTdwk92Qb 1;
NRATj7qzZ1myBtkK 1-5; Nc2grh6ByLvJCkRY 4-6; NzgdqKHnfLJWkFVM 1, 3-12; P3NRXmatYKH6zrvD 1-9; P8jXWFgJ7GVvwZ6m 1-4;
PDT2yACNzd4LMZ98 1-14; PFGvpDf9WZgcY6Rk 1-12; PLGKvft3bq8WgRYN 1-2, 4; Pj6DYZzG3JWQ4fRn 1-4; PkCJZ74Lmc3HyW2d 1-8;
PkWRr8jZp9ATL7N4 1-23, 25-26; Pn2LYHXpjZMDywgA 1, 3-9; Q1arfj4WPBNXqmLk 13; Q4GadwADtxbFBNW9 1-8; Q8GxhTmH9f6ajcNr
18-39, 41, 43-51; QAbWHfPXvRpZcN3j 1-3; QJafvkCx4yhVXBmT 1-2; QaTdmwLBtpy1gJbz 1-5; QtVv7GfBDHLyn1bN 1; R3j6zmXnDByHWJVr
1-2; RgmN46zPxHdJGcYh 1-3; TFzR3fQkt2ap7K4Z 1-10; THqFxLkCP6ayd2tZ 8-9, 11; TQAKbWq7vVJHLgBR 6-18; TZnCMWqK9vDgHjBa 1;
Tc9DkwQFmvJr1NCt 1-5; V4HDKvxmYCkR1JnB 1-3; V6zWwYR2K9FgqkG1 1-2; VBZDyjcPkmx9TN8z 1-2; VZdxq6yNBrW48ztG 1-3;
Vb3ZxfXtyDWr1ndP 1; VkJvpT6FhYwx4q9G 1, 3-11; WGRAgXb7txfpHPk9 1; WMV9yBYCj8nmQXwN 1-13; WV34XKG7MTQya1qj 1;
WYrCyJkdfMXw84Ph 1-21, 23-35, 37-39, 41-48; Wdmk6RxnLZYJMj4C 1-4; Wf26RdmcnLDTXwV3 1-2; WhbPNDgCT4G9zwYd 2-4;
XHDjdLb7R9P6NQm1 1-2; XNYqMaPdK1BJQpyg 1-6; XYT1F4RgJLh9pyb7 1, 3-5; XzxyRLg8QW9rqPGn 1-5; YFJncLm2ZPgM1GDh 1-8;
YH3c4zvFrpLtR1M7 5-9; YZyAfq6XN9k4a8md 12-15; Ya9kKCWRwqjFpgz1 3-4; Z4w7HKFPvcrTWd3M 5; ZkVJAmC8vjaM6QTK 1-3;
ZqaKFPJXQ18rWMck 1-5; ZrD7NdhjWX4zG2PL 1-5; ZxWz2nQqXRcYNAbV 1-5, 7-10; a42r3p6ZxtBTRqfQ 1-6, 8-9; a6QpkvcR1YHLBqPh 1-6;
aT8d19PF4pVznXtR 1; aVXmtyDj48AWwYQr 1-2; aW7JMwv9zYRTBLtN 1-13; aj46hBQX1AfRFHtm 1-15; amc3nxRgHhjyrqVb 1-9;
b12KD98fndF4hMq3 1-18; bCzt1Fq24mYcPXkB 1-2; bNadwz7k2rM1htLP 1; bT3LyRBtax8FQnch 1; bhW9XwPZTgD3mkfQ 1-4;
bpV7JWX8B9xgrj3c 1-5; byYP67TH4KtZqm31 11-13, 15-17; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2; cY6xjKJN7QfdbHtZ 1;
d2qfGRQw8Pr3NaVA 1-8; dDKQgLW3MFPatynJ 1-25; dgnB4QvPHR8mhtWz 1-3; dn62GTvrBYFQbWfV 7-20; dnvYkhpKZQaqM2Hw 1, 3-5, 7;
ftTngaZr73hXKHLq 1-11; g7Bh9F2b1kz6Cnar 1-2; gk2jJvTHhW4zNDKB 10-11, 13-18; h6wkBb9Z7jFtTrnp 1; hCzY42vqXrc1mG7w 1-4;
hDGL1wmpV4WjxbHP 1-3; hDJ6cbM4Hrz9Fwg7 1; hMdWYxyPG3rn9VQ7 2-12; hYHBQ9ZJX3vFAnkj 7-14; hkADNqV7JWM19df4 1-8;
hrV3QgyzvfdZKJwF 1-30; j7V46tC2GkwNYTrz 1-2; jJWt7Dnm4qc9rzQd 1; jLRdP7ah9pzfgQ4r 169-178, 180, 183-184;
jnGFZ1CcrVA2MgJ9 1-5; k9dcpgjKfyXGwA17 1-4; kWBdFxYXLaN1vryA 1-6; kgQpmaCcAhZYb39L 1-11; m1JkqnD4gjfLwAvP 1;
mDbPCdX9AvwYJxGa 16-23; mDndY9rhZbAtWNPB 1-6; mYaWhF4n8zRLDZtw 1-7, 11-12, 16-17, 19-20, 22, 24; maYP7WvxJ2TzGtDH 1-3;
n1CBNaJq4WhAfw8Y 1-2; n7qPAd4rN1JyHfpR 1-3, 5-9; nBHjcV4Ydm7z2X9q 1-2; nKhkfVbmGcg38CP6 1-2; nhtYVjb7CAw9DvQR 1;
p8ATX1Q3WFHmDKNj 1-2; pCjdMQw32FtPkYGN 1-5; pKn8wBZcQ62PyJH7 12-24; pMGyN9hFa8XRAHxV 1-5; pg7HYyM3qFXZP6n9 1-4;
q94LM8HFPN21ytpk 1-4; qMLQt41BXrpR9hkA 1-4; qMkZwCxQ9bLP74Y8 1-2; qTHXf9y6ag1CMtdp 1-2; qg8YWNaCdDvV7XZM 1-8, 10;
qprynNbXtRZV4DHP 1-7; qzv6W9fkDaP1MZYH 1-6; rD8m4JW1bVyYPKGX 1-3, 6; rQ4dFakMnWXj89vz 1-6; rV4NwFkC2hq7yHKW 1-12;
rY4H13hP6MKZfXWR 20-23, 25, 27-31, 33-62, 64-66; rh3y7JwG9fVxq6DX 1, 3, 5; rpLZ287nfKWyJGtb 4-6, 8; rz719QMkbCYvZ3Nt
1-4; rzvB21j7yD9hpaVF 1-20, 22-30; t681v9FWrLq4Tkxp 1-3, 5-7; tBZGVXz6RCqhHgj4 1, 3-6; tCprA1RmvTcJkq3D 1-4;
tD7Rm3GBjdqPzhwC 1-4, 6-8, 10-14, 16-17; tDK9h6nGz4cLTRq8 1-9, 11-16; tDhCZyVrn84xBaFR 1-7; tHz9M3arKp6PxA1W 1-14;
tXFAxTLmg3bBNZrv 1-13; taFcCyAx9BD1KWXp 16; tkQKL9bGT6JVR3na 14-16; tyFZPdH39kGq21An 3-25; v9Hf4bPtGq1LVykC 1-5, 7-8;
vGTRdWMxZYChHtzw 1-5; vTtbzncrBG6xqX78 1-4; vYmFJGLbwPWNcdxA 1-13; vdNjgtCXJVM4PT9Y 1-3; vfnQP3BCMkRmtGbX 1-6;
vgrm7nPFN231GYRW 1, 3-15; vj7g4Z2HN9JDTnBY 1-2; vqCJK6rZNzHX2Ymd 1-9; w1Mrf9TACycaXhvY 1-7; w1aNzbJn2kMPpc3G 1-5;
w6A8bVm7vTatMCYr 1-12; w6H9xLRJQCWfBK1a 1-2; wNqHMTvYmCQKXV3W 1-10; wXJGrhP2nVCQyctm 1-2; wmnqAZyLCxQ1hjN7 1;
x3VKzQDvwLtCgbrj 1-6; x3p7mzNhFcarw1fk 1; x7ybf4LjBD6GJqWn 1-9, 11-15; x89ayWwRt7DTvYCr 1-2; xRTpHyY7qC1GmghM 1-3;
xYTCXa9myPJr6R8M 1-13; xnZRBaDMKtNHGQfj 1-4; xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6; y2dFcNR1rMmKTPAj 1-25;
y4txQZad7FwJmLTB 1-2, 4-7, 10-12, 14; yKvCQdzpgqY1WwbA 1-7; yWahJbVH24Lwk68X 1-6; ynL7Gb2MpkZT46j3 1-9; ypQ9d13tMFJPNfvw
1, 3; ythJ7TZDXk6AjHMq 9-10; z4ybMrFWNGa1L2xJ 1-3; zD9hTZPfwC38bgq1 1-2; zgG84wjLNY1Pv9tp 1-2; zpJaRH6Gx7CZLq3d 1-4;
zxAPy9BLRTc4hvFf 1-11.

## Stalvan's Reaper (934)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13524
- Rate: 1.25 PPM, measured

5,827 procs on 96,337 rolling hits, 93 wearers, 116 runs, in the main hand or the off hand: all hits 1.25 [1.14, 1.37].
The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 1.28 [1.04, 1.55], 439
procs on 7,080 lone units, 16 wearers. Dual-wield white, which sets no rate: 1.11 [0.96, 1.27], 689 procs on 12,899 lone
units, 75 wearers. Specials: 1.22 [1.08, 1.36], 1,080 procs on 18,374 lone units, 70 wearers; Sinister Strike gives 401
procs on 6,197 of them. Two-roll abilities, priced as one roll: 1.08, 34 procs on 654 lone units (Execute 32 on 556,
Mocking Blow 0 on 1 and Mongoose Bite 2 on 97), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th
percentile 1 ms, 1,943 isolated hits. Dropping the swapped stretches (312 active minutes, 7% of hits, in 24 runs) moved
the read from 1.16 to 1.25. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      48 | 3,002 | 47,804 | 1.30 |
| Warrior |      36 | 2,721 | 46,318 | 1.22 |
| Hunter  |       8 |    97 |  2,021 | 0.99 |
| Shaman  |       1 |     7 |    194 | 0.75 |

Reports and fights: 1cGJrPCNvy7Fk3wX 1; 2TF1KbwpqLPRmNCn 1; 2THWnvG8AXCYqfNg 1-2; 2TNQqFYdjRLB4Zga 10; 2npcVzdGrXR6PZv9
4-5; 3XaH1Jdp4Amv2Bbx 1-3; 4ZKFNBk6czpynG3W 1-2, 5-10; 4dVyt9cvbPrgpY2K 1-2; 61VcxvhfQF2DRAGd 1-2; 6xNyB1ZawDC8gjK2 1-2,
4-7; 7QGzZNgTftwMWnrK 1-4; 7a61xpP8kGLzJAw2 1-9; 7an1DRwVTBzmtKLM 1-8; 82dvKxybP4Zfprh9 1-2; 8KH6htf4MATvC9ad 1;
9qtkJVNF7fwagTCG 1-4; 9v6VdC7ymQfb4cWh 1; A9bpD6tXZvmyMrn3 1-5, 7-11; AGXj4FCx83v2qLwy 1-2; AgTJ4ktjhrXQNKZW 17-18;
B9KaC3VWfxkbz7Lj 1-16; BjPdctn9FDTG6rfk 1-4; BtcMA1CqVrnapQ2m 1-8; C6XcAgmWVdbxkfKD 1; C71ArgnKwBqWf2Tk 1-13, 15-22;
C8Hk1q3rpN27YWam 3; CQfJXhtLZrTmFYj6 1; CaTWbfV38DjM4t2P 1-6; DTQLfhZR9WGzBKjY 1; F3qkWBzGXRcYP8rd 1-5; FTXVdBrHJZD9aw7P
1-7, 9-23; GZzaJH1QVjMKd8qA 21-34; GrcTMNyY8p2f6xV9 1-7, 9; GrqfZ4268BDWgyx3 1-14; HMfnLPQA8ra1xZtb 1-16;
HnC7N2YVAdg3wtrk 1-4; HrJFDLXd1NVmbGv3 1-5; J2HpXkgcB7WbfwNK 1-3; JHrkTPq23LpAcGxY 14-18; JZx3p6Nqdj9tCrMc 3, 5;
K1zQgCcqbnt3kHXT 1-5; KMfvLjt3ZDgYnV2m 1; MFtAJbv8W1L2jDV9 1-7; Ma37bzLRWnJpHvPC 1; MjxzVLAta8FgXR7y 1-2;
PpvQ9WzNTKgnVckA 1-6; PxfYmBXLqFg3GMvT 1-26; QYvzMfCm3pkRXPr1 1-10, 12-13; QrXYv7cC1NRDyWkg 1; RdmgVNQ4XJMfnC1F 4, 6-10;
RnfbW8KTc4JHypP6 1-4, 6; TfvPGqpL1VFhDNbn 1-21; VhkGaxQzZfMApr78 1-2; VmzWfc9FaKwdjGLr 7-11, 13-15; W1AbqkptzDRQTwJa 1,
3-5; WTjHX7kzFZt2cACx 1-3; WYpMaZPqdFX1K43B 1; WnHKVJgjY1pLX7mv 1; Zc94HvVzphwKJPgy 1-4; aBjNQDx4M8y3YX6L 1-2;
bHkgvJcQaRpXYB3y 4-11; bjBRJQvM1kCTVhdn 1-2; cTgakFj2W7xwHmJ8 1-2; cfmbxDLPZArF71wt 1-3; cxypmPhDvXwtVb67 1;
fLBZ8dpkzXPVmqh6 1; fdKHYz6Xjya9CGRQ 1, 3-5; ftTngaZr73hXKHLq 5-11; gKwj1ZPTzXFmrH3Q 1-2; gVtWYZm7Gp1wChxd 1-4, 6-11;
gk2jJvTHhW4zNDKB 8-11, 13-18; gxwj8BfpT6A2FMZy 1; h9XqNAJk6aPntHpv 4-12; hJ4Lvn9aN1fWzcHY 10; hYPvpAktGL6TZRFc 8-9;
jJ7xtfLqMQ9XKpY4 1; jQfz47GbWx9HZcPR 1-11; jgGH3A1KqvWZTcaF 4; k4QrLh1VCMzTN9Aj 1; kChTQHFwM2DY1fNV 1-10;
kHw4dmRbxpA9tW8N 1-3; kQd6CrMNGcHtDFjZ 1-6; kQhZqY4DwGBKLdag 1-4; kRPfh761abZ4MxGz 1-12, 14-15; mtBc1WZf3ydHVPFv 1-6;
nAPqCW1g46kadrTV 1-4; nPXGLarxKjRfpb2w 1-2; pLh28qFZjAYvrMT6 1-6; pNxkRa8jzG1yDbh7 1-2; pjP3BMVD9ftWzHcy 1-4, 6-10;
qPv2mdpY9T6H4RXA 1-2; qTFZ6Cn2w19ga4VW 7-9; qW3KvYDpw6ARxZbc 7; qWR71CkpDfrdHwGg 1-2, 5-22; rmnN9TZ1L7zRKBGh 1-2;
rwGYLaZMtQCmp9DN 1-2; rxKDLmg7wkZA6WPM 1, 3-9; tBf8NcDpAFLvZn3W 1-28; tZHf7DRXP2AW8v4G 6; tmwXD93aHW7VCZyA 1;
txn8Q1X2CLpcbwyA 1-2; vHjpcxGzbyMhAqJ6 1-8, 10; vdNjgtCXJVM4PT9Y 1-3; vjq9hRMPZG6Nw4HY 1-6; vtHaBZF2VAqJ19xY 1-16;
w31pKWNxGd9bPfMZ 1-4, 6; wrCK6pfGP49Bmy7j 1-9; xR7TzM9Zp1Pg6G2A 1-6; xnrg9m6jDMR1kaLN 28-29; yNvbPCQn4T1ZzrAd 1-2;
yX2GQL1zDK3xC6jn 1-5; ynxjYKNVbRBdGWfg 1-7; zBRj1JvVfYGL4ygP 5-8, 10-15; zK19ZY7WfDt4JC6k 1-3; zbYMFrgC1VD9kXRW 1-4;
zyTLWmK6gvX3x4NP 1.

## Strike of the Hydra (6909)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1292657
- Rate: 2 PPM, observed

4,796 procs on 43,914 rolling hits, 131 wearers, 151 runs, in the main hand: all hits 1.99 [1.93, 2.04]. The one-hand
white rate and the specials rate read below the all-hits interval. Observed: the one-hand white and the specials rates
disagree. One-hand white: 0.99 [0.90, 1.08], 686 procs on 12,572 lone units, 122 wearers. Specials: 1.17 [1.07, 1.26],
673 procs on 10,489 lone units, 112 wearers; Cleave gives 309 procs on 3,785 of them. Two-roll abilities, priced as one
roll: 1.65, 86 procs on 948 lone units (Execute 84 on 911, Mocking Blow 0 on 1 and Mongoose Bite 2 on 36), left out of
the rate. Travel time: the proc lands a median 81 ms after its hit, 98th percentile 247 ms, window [-20, 257] ms (1,176
isolated hits). Dropping the swapped stretches (142 active minutes, 6% of hits, in 7 runs) moved the read from 1.87 to
1.99. The paladin check, which sets no rate: lone white swings 1.12, 438 procs on 7,115 lone units, 93 wearers. 3 runs
logged in two reports count once. One run is dropped whole as swapped. Reports from 2026-11-07 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      72 | 4,359 | 39,801 | 1.99 |
| Hunter  |      59 |   437 |  4,113 | 1.93 |

Reports and fights: 12A94BJt3KCFZrY8 1-12, 14-21; 14caBjX2QdynLkq3 1-9; 1AxvQpfrXyTg3dCW 24-42; 1FR3hDpwGHbjBnaZ 1-2;
1a3pDtCcPgBkL6wm 1-22; 1mRKL7Tb8qwjkgv3 1-2; 1yjwF8MxXqvtD4bT 1-3; 23ZyGvk7BXAVDT4F 1-14; 2Jky4AMwBgct6aKd 1-3;
3BH2d16wtD9F78QY 11-18; 3M2GRBtqg4X6a7cA 1; 3tRNmGkZB2Jhnv6b 5, 7-9; 4nmJDcGNkZC2Phdw 1-6; 4y8nq6Xjg93zfrc7 1-9;
68fgHQ74dRxrwac3 1-3; 6DYqj7B2WHxPcabQ 128-134; 6X1L4nB7zCfqk8vV 1-2; 6jHPJBnLy1pF9Rkf 1; 6k2qACcn7WVzthrH 7-51;
6pjNwdr12qzBAKhW 1-2; 7GNK4Y9bkH1awZxL 1-10; 7yRD39nwhmMWYjbX 7-8; 8NtKzDpRBTgxhar4 1; 8Zfz4JNLxhHcdDtX 2, 4-5;
8apqXKQg6mbC91L3 14-18, 20, 22-28, 30-36, 38, 40-46, 48-49, 51-71; ALMJZm2dXhNKkPTq 1-9; AV1aRBfcGdyC8Qn3 1-8;
AqxY6kh43jc2RPaT 16-17; AzMD4BWYKT6NGX2d 2-3; B2k3Xcxz1hqQwJ6P 1-12; CXmbAKWLnyJ2tgQz 1-4; CyFhB96pgWwc4jbD 1;
F1v8TxmwrXp6knHR 1-32; FCq4t8GNAj7cRLMa 1, 3-9; FbhmQgZj9TH7f84t 1; G4hzybWYmgrXvL83 1-2; GZzaJH1QVjMKd8qA 23-34;
Gv6NDzAmFjC4P3qL 5-15; H3bRaDmzyjxTW2BJ 1; HARvQ2hJWp8f3grD 1-9; HpftBG4KPCZrMxbF 1-3, 5-9, 11; HpyNQCMnDYFbJtP6 1-18;
HtQr7kYaNq9MGgvZ 1-12; J9PZzNGkgBwWcf1C 1; JQH2y6NWBXchGRbg 1-3; JxNpTAqWtLyHzZG3 1; KQ4mGZ9NCdLF8byr 1-2;
KZYNzqfJFdW34DVr 1; KrtPndVLH4BjJcfD 1-4; L9k3ND6tw8VvcgjP 1-7, 9-18; LAm6bcJw47YBh2kF 1-2; LzcCRM4GNtK1HdJf 1;
MLhC3KjWrngq2vtZ 1-2; MQwzXcfxb4v7hnFy 1-2; McYJrBbXAfKj82W4 2; MjfT3pQKRWcVG4gY 8-10; MvrY6aBgftZxj1QX 1-9;
N178cBjPtChmbk4w 1-7; NJLbaF8mGnk9yYAq 1; NKXBATdWYqmz78CV 1-2; NMfHFAKRVXm8jyBz 1; NvnJVbtwK8z7THyx 1-2;
NzCyTLqbtAHM3Fm6 1-3; PCfQFpjYdm3rD8qy 1-7; Pc3gyq8AFL9CD64m 2; QCLPznqJMZ8vH3mf 1-12; QJ42CwNWLjFyvqfR 1-3;
QhYmjJ4bHZXcgfk1 12-27; RyaVqvxGnD8bKj3w 1-5; TCzRJ82rWcPZmp94 1, 3-7; TDqbRJWAN3x1FGhv 1, 3-9; TFVvAaY6HLj2NqQp 1-3;
Td48VjbADm17Gwxn 1-2; TpxWY1MakPNb72H9 1-9; VHth6gR7xLbjQmnr 4-14; VQaYAf1Wxvm6ndHK 1-11; Vd1NZCgLMqxhb9kP 1-28, 30-39;
XF1faW9mcYtKrqMC 1-18; XVvTwBdtzAcKhPg8 1-14; XpQW1hmwHYDAJ8at 1-12; YX61J8CKqc42paNZ 1-4; YcVR36BKNkmgQLyb 4;
Yt6CVhy34zjNHMcm 1, 3; ZFLd4hkVbrJ68vnK 1-13; ZK8YCvy9wba13cpk 1; ZLqQWyd1Bjr7HD42 1-9; Zf21kBmtDM3AyRjF 1-13;
ZkA3Vhz72n9Gy4dY 2-7; aRmjxFbtnGYAKCqk 1, 3-4; adXPnhT3kjHp9VcC 1-13, 15; b4FMw7XKgxWnvYN3 1; bPQCc3Nvq27HRWt6 1-3;
bapW9RQhFLvJmgtn 1-14; bdYgfJk7Mh8GqZ4t 1-5; bjxnchQ7qYRfVPtv 1-2; bmf23JNRp1w7WQKg 1-5; bnx7WwavBpKrQjNh 1-4;
btfTZQLYD4hwvzqx 1-4, 6-9, 11; c7gHWKFrZMnm2yb4 1; chaXJYpVCrPyTm1N 1; d6f294aV7vYnWmFj 1-5; dAV7JrjMkGfKNZtY 1-14;
dQMKjpxJBCz1wrDW 1, 3-12; dZVp6gHA92mbDnPk 1-3; dbXaAxv1cTVMh49j 2; fnYThW3BpxKZv7jw 1-18; gaDnjKvP6RkQTW2B 1-10;
hG8tvQjLBJgf4WHY 1-2; hRHtBCkm8c2d1nNK 1; hgd1qH4QYFyjz6AK 1, 3-6; hy4nZYdFGMDVf8j1 1-3; j4qGhxQ26AkKBwXy 1-2;
jbhGzfrnJRWX978L 1-10; jrwhRBWdbZPK1Xka 1-3; jtn8ap4XY6PzwDQL 1-11; jwDg8KGcd3QJ9kNZ 1-3; kNaPyJxKDvgrAf4w 1-3;
m2gBvZCnrRVq7aNA 1-2; nB7qNVpyZvLPXCJY 1-20; nJtB3F7cQy6NmzGZ 1-12; nKdyat38zLA7fNJq 1-18; nN8fPWtCDQdkmZhJ 2-4;
njaHZVpvYFtrTLNh 1-10; nqjwyVJp17PDBYTX 1; pLFZK3MWV2TRw6t7 1-2; qJjhQpNcr1tvVPfx 1-21; qfG6QWjd3wg1JVXv 1-18;
r2zKPQpkyjAY64Zm 1; rY4H13hP6MKZfXWR 20-23, 25, 27-31, 33-62, 64-66; rbxaknWH6TCB28jq 1; t1rjBxRdP8DzVHhT 1-10;
tVFZybKHRWp8xMcf 1, 3-10; txP9jAb7gGHhXckB 1-6; v4QYKcD7XymZq2JB 1-12; vAGVD7aQZxNWX8h1 1-17; vgXDVpPfw1QnWjYz 1-2;
vnwtmMTpkR1aYP7G 1-9; vpdJ3cMhwTKNHX7B 1-2; vqBtZRrXwKQb6xHj 1, 3-7; vyzbt6kgRhTq4fdr 1-4, 6-9; wYmr1kfWPpVz7jyZ 5-13;
wvmXrcnGp7YL918N 1; xnrg9m6jDMR1kaLN 25, 27-29; xwGgV9zJ3FNayZf1 1-2; yC6t8PJx9VRLrj4F 1; yDMwzc49TJKmFfCr 1, 3;
ycZLWXxQ4jk21w3q 2-28; yrMv6JBhWYdfpk87 1, 3-6; z4ybMrFWNGa1L2xJ 1-3; zWrR937TBH1FgvfP 1-2; zcMQN68Ava4JZgqT 1, 3-7;
ztD4pKHA6jGvfcVr 1-2.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13527
- Rate: 1.8 PPM, measured

384 procs on 4,532 rolling hits, 14 wearers, 14 runs, in the main hand: all hits 1.82 [1.66, 1.95]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 2.04 [1.71, 2.56], 111 procs on 1,164 lone units, 13
wearers. Specials: 3.21 [0.66, 9.39], 3 procs on 20 lone units, 1 wearer (a Poisson interval: under 5 wearers). Travel
time: the proc lands a median 100 ms after its hit, 98th percentile 196 ms, window [-20, 206] ms (108 isolated hits).
Dropping the swapped stretches (16 active minutes, 6% of hits, in 1 run) moved the read from 1.71 to 1.82. Reports from
2026-11-13 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |      13 |   377 | 4,473 | 1.81 |
| Hunter |       1 |     7 |    59 | 2.54 |

Reports and fights: 1ZC4fDT2xKnHX6Wj 1-4; 2axJHNwqQdtW1jnF 1-3; 7kxNnWGw2FzfLJtR 1-8; Cc32w7hgMJQ4Tny9 1-2, 4;
CyNXHdAzDWvQZ6aR 1-9; FRxmtWLHr3wvB8Q1 1-8; LK97QRMwtYNkxrX2 1-2; LjFr1v7pJKVcfhWB 1-11, 13-14; VhbQXPagDp81jtyY 1-7;
argXcZydQj9tTHD4 1-2; fDycAFZWt4NkRX1x 1-3; rNwJc31qXfzWthQD 1-5; tBnv1kDbzHK6hAwg 7; tpfAmB8HCvDJRPjw 1.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13528
- Rate: 1.05 PPM, measured

604 procs on 12,875 rolling hits, 18 wearers, 18 runs, in the main hand: all hits 1.04 [0.97, 1.12]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.06 [0.92, 1.33], 92 procs on 1,920 lone
units, 8 wearers. Dual-wield white, which sets no rate: 0.91 [0.74, 1.03], 83 procs on 2,033 lone units, 9 wearers.
Specials: 1.10 [0.98, 1.26], 197 procs on 3,964 lone units, 18 wearers; Sinister Strike gives 67 procs on 1,588 of them.
Two-roll abilities, priced as one roll: 3.51, 3 procs on 19 lone units (Execute 3 on 18 and Mocking Blow 0 on 1), left
out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 306 isolated hits. Dropping the
swapped stretches (9.3 active minutes, 2% of hits, in 1 run) moved the read from 1.02 to 1.04. Report APBKD3j6wYWG7xa4
logs the same run as qnCwWrxJdFVgHtvp and counts once. Reports from 2026-11-14 to 2026-11-21.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |      12 |   373 | 7,702 | 1.08 |
| Rogue   |       5 |   226 | 5,111 | 0.98 |
| Hunter  |       1 |     5 |    62 | 1.79 |

Reports and fights: 2mr4qBpFYGxVMAnj 1-2; 6qHfrKb4aJ1PxYjp 6-8; 9nCtfP6kvHrzV4FL 2; APBKD3j6wYWG7xa4 8-11;
CDZG21zLfmMdh6FN 19-21; LDg2vXctwnYM9jBq 1-6; M8NjYfpWZ1b3KRgd 1-3; MXjF1GhRmLfabNZV 1-18; NfDXwFm37WHghq9d 1-5;
QwzkrjDGHnW37qpv 1-8; TaNXwvhfBLxdyQc9 1-2; XYTRKMVFWdypLxtB 13-14; YFnbmvJgWVXzcBwh 1-7, 9; ZHK32LtTCmqR4Pxf 1, 3;
ZqghBpkcy7xtPfmH 1-9; kTaK6pzXtHZM4f9W 1-9; nZ3NhbvpVXg649Bm 4-10; qg8YWNaCdDvV7XZM 1-8, 10; qnCwWrxJdFVgHtvp 6-9.

## Tainted Pierce (8225)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13530
- Rate: 2.05 PPM, measured

1,346 procs on 20,986 rolling hits, 26 wearers, 32 runs, in the main hand or the off hand: all hits 2.03 [1.89, 2.12].
The rate is all hits, which the specials rate agrees with. One-hand white: 1.79 [1.23, 2.51], 33 procs on 582 lone
units, 2 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.90 [1.73, 2.05], 418
procs on 6,933 lone units, 22 wearers. Specials: 1.97 [1.71, 2.21], 128 procs on 2,047 lone units, 10 wearers; Sinister
Strike gives 31 procs on 427 of them. Two-roll abilities, priced as one roll: 3.58, 23 procs on 203 lone units (Execute
18 on 139 and Mongoose Bite 5 on 64), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1
ms, 412 isolated hits. Report Jh7F2XTzVnL8Hyw1 logs the same run as THqFxLkCP6ayd2tZ and counts once. Reports from
2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 |   672 | 10,697 | 1.98 |
| Rogue   |      15 |   588 |  8,816 | 2.11 |
| Hunter  |       3 |    86 |  1,473 | 1.84 |

Reports and fights: 1zLtZcp3DRP97Hdx 1-6; 34LbyqDBVnZJRYGa 1, 3; 4na1LcCpP8QHTNVD 1, 3-8; 7AnqNjw91W4L82yf 10-11;
8WfLRh1qTpFxJgra 1-2; 8rNkjZ2B6WdzFcKq 1-3, 5-7; 9WJHzhV8gLmGnZ6c 1-2; FrCdbcxM26VHRkym 14-17; J97GvmrRXCFZfQ8a 1-2;
JR4xwhVZ1jzngqyN 1-4; Jh7F2XTzVnL8Hyw1 1-2; KXmBtLjWnQ6zZa21 9-14; LGj3Kfvkz7cgZ1ya 1-5; NXgv93kr2G7zyTxm 1, 3-9;
R3CDcTfm6bdNJrHx 1, 3-7; RcTKFQbGnmtJHWpa 1-2; THqFxLkCP6ayd2tZ 8-9, 11; Tm68LdcfN1rZYFap 1-13; TnbzrVaxtHcpMKLY 1-3;
VyZXRT7nrPJKBAt8 5-13; akPJ4AcBZQnb36f2 1-5; c9tKjzRyaYXQfnk8 1-3; cNxrG8JKMXmw3PRH 4-7; czgKmt8H3NJC7Gxa 1;
gn4HAC1RqMfLt8wN 16-21; pFNnChR7GzAX6aH3 1; qcrKJGPvkfhAtdjQ 1-4; t3aPrJ8LCkb2zKhB 1; tQc389CTMRrAvqFg 1;
tb4N2WMhjXKxRVPp 1-4; xqy2GfPj67XCwrT3 1-4; xzqG9k6MvrgRnPJH 1; zTMW1B2rRJxbm3w6 4.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 17407
- Rate: 1.95 PPM, measured

4,629 procs on 35,259 rolling hits, 29 wearers, 40 runs, in the main hand: all hits 1.97 [1.91, 2.02]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.96 [1.86, 2.07], 1,017 procs on 7,773
lone units, 29 wearers. Specials: 2.15 [1.95, 2.37], 198 procs on 1,381 lone units, 11 wearers; Raptor Strike gives 94
procs on 726 of them. Two-roll abilities, priced as one roll: 3.71, 44 procs on 178 lone units (Execute 36 on 138 and
Mongoose Bite 8 on 40), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,172
isolated hits. Dropping the swapped stretches (55 active minutes, 5% of hits, in 2 runs) moved the read from 1.88 to
1.97. The paladin check, which sets no rate: lone white swings 2.07, 97 procs on 703 lone units, 7 wearers. Reports from
2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      19 | 3,286 | 25,178 | 1.96 |
| Warrior |       3 |   796 |  6,139 | 1.94 |
| Hunter  |       7 |   547 |  3,942 | 2.08 |

Reports and fights: 2Dv3WbJrMTCFPVgH 1-4; 4yJ2cwCFqNnWYthD 1-2; 6FVmWYfq7DG2pzwJ 1-7, 9-10, 12; 6jcxdQpGHvgTmVC2 1-2;
872v1HVANqbJGzar 1-5; 9QDMdpt1bmcAaR4q 11-44, 46-50; AKXn47BZWpyhJRv3 2; AN4RP21ZMXLVpntm 1-3; CWzdaFnMfPbALgmJ 1;
DKhtLa2GqWFxfgnT 39-61; GF12jgbN7Yk4APhB 1; GfyxbaQhZvL8dT9H 2-4; KtCPg2xhR416AvXN 1-6; MBRmC1TnXq86k2Nb 1;
NLftBVcrRhPwnv62 3-31; NXMZV9xjvWpB7GtP 10, 12-107, 109-116, 118-122; QqKMnRmb7DrFXC36 1, 3-6, 8-9, 11-12, 14, 17-19;
WkrD6xYbcp8JZQBm 1-2; WqfyDhVHxT93Q7Pt 9-29; ZaPqVJn1bQ894pxT 5-17, 19-52; ZqtpzHLbAQKWMcwN 1-3; aP6J2RH1jNfLZmA7 1-2;
aRLP2JdQntKzTx3f 1-2; bGq3mN7JgCDzcyMZ 1, 3-5; bahZczXqfLk3pQR2 1, 3-5, 7-19; bgycvf8pzGCqNBDm 1-3; c3qNjGmB7x9akpgr
3-12, 14; cav7VZrmPbNW3y29 1-13; cbTt1zxajFfNVPDA 1-4; dr1BMh6pNxqbFJLf 1, 3-6; h4JXmPQDqWL61Cyc 2-3; hdVbcpYXGPDaFwyR
12-17; j1Hzrkg6aRYdxvCT 2, 4-5; kNDzgqdc4nQB9vpA 1; kaPpXgt1bWJBr4z2 1-3; tgQ4FTkDX6NRyhmG 1-2, 4-5; vwCXzhMQ2drgpRxD
1-41; wrzB6GtjPaqCKJZW 3-10; yVp9KT3QkzhmZXDR 1; ztyQjYFA8BT914fC 2-7, 9-12.

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

- Count: §Count of the 2026-11-13 to 2026-11-26 reports
- Proc: weapon proc 17505
- Rate: 2.05 PPM, measured

3,278 procs on 52,949 rolling hits, 41 wearers, 54 runs, in the main hand or the off hand: all hits 2.06 [1.99, 2.14].
The rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 2.14 [1.94, 2.35], 376
procs on 5,864 lone units, 12 wearers. Dual-wield white, which sets no rate: 1.93 [1.76, 2.05], 825 procs on 14,253 lone
units, 28 wearers. Specials: 1.99 [1.56, 2.40], 110 procs on 1,844 lone units, 13 wearers. Two-roll abilities, priced as
one roll: 2.40, 12 procs on 167 lone units (Execute), left out of the rate. The paladin check, which sets no rate: lone
white swings 2.02 [1.46, 2.60], 34 procs on 562 lone units, 11 wearers, not apart from all hits; paladins in at that
rate, on as many hits per lone unit as the other wearers', read 2.06, so the rate is 2.05 either way. Direct: window
[-10, 100] ms, median 0 ms, 98th percentile 1 ms, 906 isolated hits. Dropping the swapped stretches (84 active minutes,
5% of hits, in 6 runs) moved the read from 1.97 to 2.06. One run logged in two reports counts once. Reports from
2026-11-13 to 2026-11-26.

Across the 2026-11-23 hotfix, which sets no rate: the reports before 2026-11-24 01:00 UTC read 2.09 [2.00, 2.19], 2,695
procs, 35 wearers; the reports from then on 1.94 [1.79, 2.06], 583 procs, 8 wearers.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      22 | 1,670 | 26,702 | 2.08 |
| Rogue   |      19 | 1,608 | 26,246 | 2.04 |

Reports and fights: 1ZyhaVDdFtTYWMcC 4-7, 9-12; 3BAz1TDGL9kjy8ZQ 1-2; 3RV6CW8rqMjbYT24 2-3; 72q1FtWgTMHAcLxG 1, 3-5;
7R9HYFpnN2XCq3bJ 3, 5; 8jDqr3JcYRCH1pbK 1-4; 9yWkxhQqKz1bR8fg 5; ARW6wQJHvNTVjbFG 1-3; BZ8Q1mHKCFJ7YTL6 33;
C1cyPbL3RQvGjq8p 48; DfKPt6rhYvCQFAy7 1-4, 6-7; FgPGfqyhC8j7rB4Q 11, 13-14; H8AgW6afwBktp4nv 3; HKjBYG4kRJtnXyc3 1-2;
HL1FjV9MzyNDrb48 1; HNyG1nBQfCqPZYbL 3; HTQP1MZrjha2p76V 1; Jth1dNXDLrjFm7yY 6-13, 15-38; KQZwHXz2TP1W4Mh6 1-9;
KtCPg2xhR416AvXN 1-6; N8RwqgTVYMvAxG6H 2; Qb1yWxa8vHCGL7Yr 2-7, 9-15; QcrC9VK64xXa1JWg 1-8; QdwJD6mHNVfBxrFp 4;
WmV1qKzjA42DkhHC 2-3; XfBLnTDgpdYWkmzJ 1; Zz3YcQk69P7N8rAn 3-4; aVXgmYF9Mv2A7LcH 1-13; aZ8yxm3AGgrpTvYh 1;
anAZd29rK3vCJ4wQ 14; gdN1cBza69wJGKZf 1-3; hLxABfWVavkM8HGT 1-4; hnWavgKQt3ZXLBdV 11-13; jq2PK7vmkVBcQNxH 4-14;
mNTMK4J9PQcjCR6L 1-3; nzXW6jJTC12w4v7m 1-2; p6mVRBzvWMqnQZFc 11-12; prxnh2JVAPwFdj9z 1-2, 5; qa8wcv4jLYZpGARK 5-9;
qcdyZAr2ganDtxNW 13-44; rRnVJhQFa9zf7NqK 4-5; rhAc6qLKVn7JQ8GD 1-2; tQvgam4P62MdjCGc 5; tr864mQnJgwHZvpM 1-4;
vT1rMa2cNGXLfqp7 9-12; vyna9XkWc2GpD1rV 3-4; wTpN2bMCjdWyA4Jg 5-34; wnBcvgTJaDPfCYxW 8-47; wr4CTRnNBDZa1jQM 34-35;
xGJcKTnzPmAr1aCw 1-4; yCdRrMhYwm69xj1Z 1; zBqkZQ1bxp3cAhDa 15, 17-18; zy3NjCQP7bMp1W8H 1-2.

## The Ziggler (8006)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 13482
- Rate: 3.55 PPM, measured

5,825 procs on 58,237 rolling hits, 63 wearers, 68 runs, in the main hand or the off hand: all hits 3.53 [3.42, 3.63].
The rate is all hits, which the specials rate agrees with. One-hand white: none. Dual-wield white, which sets no rate:
3.31 [3.11, 3.50], 1,533 procs on 16,330 lone units, 58 wearers. Specials: 3.39 [2.67, 3.88], 258 procs on 2,688 lone
units, 30 wearers; Sinister Strike gives 150 procs on 1,643 of them. Two-roll abilities, priced as one roll: 4.87, 4
procs on 29 lone units (Execute 2 on 25 and Mongoose Bite 2 on 4), left out of the rate. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 1 ms, 1,407 isolated hits. Dropping the swapped stretches (42 active minutes, 2% of hits,
in 3 runs) moved the read from 3.47 to 3.53. Report tMg1aAHQGJkxpYcb logs the same run as 7dZt6b9Kh4F2CNAn and counts
once. The sim does not simulate this proc yet (the generator lists it as a missing effect); the rate is recorded on the
item's effect. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      52 | 4,846 | 48,187 | 3.55 |
| Warrior |       5 |   838 |  8,555 | 3.46 |
| Hunter  |       6 |   141 |  1,495 | 3.33 |

Reports and fights: 2K4Z1jLyGWzdnD3f 10, 12; 2bYvT8kWxjFfdwrm 1-2; 3DQy7BhjH1mnZPpq 9-12; 3HwJAWkhFz7gCaQy 8-14;
3QgzWpJKCNMRvAYT 3-8, 10; 4xnDCWwrGKAFdNHv 1; 67AMbHctYhxwQnfj 1; 74GT2KcnWCLq6Bby 1-2; 7dGmwDN6qKnFxBLz 1-3;
7dZt6b9Kh4F2CNAn 1-12; 8YzmGj6fAZqMBNtx 1-7; 9MmVG8crjzJZ24K7 1-4; 9WaCtMzNbFcr4qmv 1-4; 9kRzG7P6AKj2pdaF 1-2;
A2FMrK3Y1xkvafQH 1-7; BTRMkxmvfrwcZNaG 1-3; BmfFgAczHXN7b16M 9-24; CcKvJ8BGNkV7wQ1n 8-9, 11-16; FR9jJ7b31XCkLmtQ 5-7;
FxndQv9RDXt47KA8 1-2; FzCJvnHfGyMLWVwb 1-11; GyQ6cqtCxPXHaDFz 22; HAZ4a9BmGxyD8gV3 1-3; HhgZrVXGJMyQjD4B 1-6;
KY97cpLDjBJ6GWfd 1-3; LH4aGRZt7wVhkpcD 5-21; MWqATtYvk6BprzfH 1-8; NyC7zw4Pm8aBHbZ2 10-20; Pn8QfjmVL46tWpXR 1-9;
QpGMmwvfJL9Cknab 1-4; R6tL9kh7nwNyGCVj 1; R9njq1AGDvmp36bM 1-8; RPnVYQcG9xTJh1ar 1-4; V4HkPzgr21bxRvFt 1-3;
VYBNbJx7G8kP6vXd 10-15; YTn9F1x6jLqN2fdQ 1-9; ZbY9WHdayD6FGhJ7 3; ZkpTbqwFYMAaLP4r 1, 3-6; aHzGnKtALM6PXVjD 1;
aWDj14ATX8qLwMNv 1-5; adJRDVZ1mhYP8vj2 1-4; bLmVHN7BaP8DvX9Z 1-16; fCJMF2rZqAHjBw67 1-5; fGTNQBq13vV9hLwC 1-3;
fw1JyYmMWRLQ7BDt 1, 3-5; gGWbZn94md3tDprC 1; gLcGzRyxhZ2akWjw 1; kvZgPDVybd6YN4Hw 1-2; mayrdRq4hwZBjCQN 1-6;
mjbfcTLkM2PGKadw 4-7; mwvHGaXY3KJ1xjdh 1-7; nwh3zdqc4CjXa6MG 1-7, 9-10, 12-23, 25; pZNqCry3wRxmhk2L 1-16;
pmKDwngMT2NFktdA 8; qd91JpfHmAwzFCQv 1-3; tMg1aAHQGJkxpYcb 1-12; vcxFpTb6zP2WrVAj 1-8; vgrGc2Zt6BzQ849V 1-7;
vkM4HWNYydVbPGxf 1-4; x3PftMnXCBZgLAG2 1, 4-9; x7mkBJZvX8PnKcRf 1; xM9bRGyajmBFvCqt 1; y8wDKJvdL1Ncp9za 1-3;
yCgqNx6LHWZdn3Pf 1-2, 4-11; yb6BrpYZnzHajqmx 1; yhVbHg7vzDkKfnRY 1-7; zLdkFAB1XmjRWhbw 1; zfaGW68H1jp9LRVm 5-14;
zhkQZTaX8cVg47Jv 1-5, 7-8.

## Thornspike (6681)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293183
- Rate: 3.1 PPM, observed

19,457 procs on 220,571 rolling hits, 283 wearers, 330 runs, in the main hand or the off hand: all hits 3.11 [3.05,
3.18]. Its lone-unit rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 2.03 [1.69, 2.32], 120 procs on 2,091 lone units, 30 wearers.
Dual-wield white, which sets no rate: 1.52 [1.03, 2.04], 29 procs on 673 lone units, 145 wearers. Specials: 2.08 [0.65,
3.55], 13 procs on 221 lone units, 45 wearers. Travel time: the proc lands a median 562 ms after its hit, 98th
percentile 1284 ms, window [-20, 1294] ms (68 isolated hits). Dropping the swapped stretches (501 active minutes, 5% of
hits, in 47 runs) moved the read from 2.97 to 3.11. Left out: 2 of 332 runs in Cat or Bear Form. 3 runs logged in two
reports count once. One run is dropped whole as swapped. Reports from 2026-11-07 to 2026-11-22.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Rogue   |     194 | 14,240 | 160,189 | 3.14 |
| Warrior |      49 |  4,228 |  48,725 | 3.06 |
| Shaman  |       5 |    519 |   6,309 | 2.90 |
| Hunter  |      34 |    470 |   5,337 | 3.11 |
| Warlock |       1 |      0 |      11 | 0.00 |

Reports and fights: 12RB4fbmrp3YZjPF 1-5; 1687brC3aBYKVF9X 3-8; 192ZN3dKFymLACY8 1; 1NLHzKfayFvtQVcG 1; 1Nnrd2wCHvXxBAtT
1-3; 1VKQY2Dafc7PJzpR 1-2; 1Zv6bBfgDP3RVFYa 1-2, 4-7; 1kWMb9cQ67dJhrZz 1-3; 1tCvNKHFM3c2WkmA 1-12; 1w8AbW4RBNLX6Kfy 1;
21PKRb9nk3HdYFm8 1-11; 2K4Z1jLyGWzdnD3f 1-8, 10, 12; 2KkaByn6H9mzb4Yq 1-4; 2ghYzpnG48dcxZJ9 1; 2t6mDPLkfWBxcr4q 1-18;
2xhZN4GMYwX9nTHL 2-27; 39BFWKC2J8RTzyMx 1-3; 3DdtwKbCf7g1XVyT 1-2; 3PXfhvmTp6NQdwRb 1, 3-12; 3ZFVrmzh4fanLM1B 1-8;
3bqWjVMPpDfdCgYz 1, 3; 3f4WnNCQJX9azxmV 1-5; 3hNwr9K71bz4xjCc 1; 3kqWrL1dwJfxMcX2 1-5; 3npdHy9Ag7rMTC6f 1-5;
4KCzQ8mFr9A1bkxG 1-8; 4L7DBHTV2jA6gkQW 3-19; 4NbKcpG8AxJW1Mg6 1; 4WYavRTZm7QHxLgA 1-20; 4XPJN7RDcM8BknFY 1, 3-7;
4ZNwHaLtygF1k3xC 1-3; 4hcXbHqV9dNDf1pt 1-2; 4zYkPVvpb9KXLxA7 3-5, 7; 6CPXjKh3z1TDWtqM 1-10; 6tTZVHLYkbzrCXqD 1-8;
72gPMatGLpwd93jW 13; 76ZV8FQ1RHDhbq3z 1-4, 6-9; 7M1KwnDZkXvyRFB3 1-5; 7dZrzYqjKHh28mvW 1-2; 7dZt6b9Kh4F2CNAn 1-12;
7f6DBRAPMLGYzyQm 1-2; 7g9DpfNFBA3PatdZ 1-7; 7gPtBMrHwGyjWKRc 1; 7xGQWc4AbvD8CNBw 1-7; 81CznKf7taTFArLw 1-8, 10-16;
8MjcygF3zWK4d9NB 4362 fights from 1 to 4766; 8Z1rYqmCWKbLjvwz 1-2; 8ZTCpbxtFA1zfRNX 1-3; 8cyBqw6Cr3WjV9Ft 1-6;
8raZtT1Qq7WkYv2P 2-5; 9B8Wf16dHKjgaMDY 1; 9Nayz1f8BJnLcgQm 1; 9PxCyQfrdhXFmNGV 1-2; 9g7NQBqDH6FGfP8L 1-4;
9v6VdC7ymQfb4cWh 1; 9zfKMnRxtJGZQXBL 1-4; ACWD9zRg8jaTbnt3 1; AvtKjphWbBxNH3rV 1-3; B1xchF3f6L2Mkga4 101 fights from 198
to 306; B6jfw4mcnPxWCKYD 2-18; BTgGta4c82yXYJvF 1-4; BZxkqfarM97yHgLp 1-6; BdLc4AgFbrz8Tq2k 1-2; BgGa4DXLy1vCnpMz 1-3;
BxpTfqGZb34aJLc6 12-14; CPj7JpNFLDztgQRa 10-19; CbG9THJkRBjDmPX4 1, 3-6; CjLKJRNhnb7TBzAZ 17-20; Ct8XcxY26Mf34NjR 1-9;
CwYNtp2VQrJdbfx9 5-8; D8qzgc9h1YLf6krp 5-6; DHyYh18vLXAFTJqt 1-5; DRv3TVCwrAPJ9zK8 1-2; FBQzxArfjdCZp6Lg 2-11, 13-18;
FTrHGQmvDM2jzRn4 7-10; Fj9WzagMcZxw4mXY 1; FvdWRrT67kAPY43x 1-11; FzCJvnHfGyMLWVwb 1-11; G6wYN314W9FfVXpr 1-2;
GAgDtJmKHdBhnpwV 1-5; GBKZhMtwamHAdpgk 1-2, 4-6; GCyghdxmrfYL8zcX 1; GJxyK8rTBamY7CvH 1-11; GKmRMd7XzWtavJg8 1, 3-9;
GNRbXQDvCkYpAta4 1-8, 10-12, 14-22; GRZrtByWvML6mzXA 1-11; GRdTC8kXNwVxpjKQ 1-2; GV4wtvzD8RcfTqNC 2, 4, 6-15;
GVzLqkhK4PMXjYmn 38-50, 52-70; GX7CDf4dvtLmR1Wh 1-3; GXmnxzAfNabcMB6Z 1; GYPbQpyCvVajRH4L 1-16; GYQB4c79w6TF1nWZ 1-6, 8,
10-11; Gb2W8Tga4VqvZy9c 5-11; GpdNkxVgzrtF49Xb 1-3; H7yaPANhDBKQ8Lmq 1-13, 15-17; HTpFLJAq8CWmPvzr 1-2, 4;
HWcyVPwhnNB9brYz 1-5; Hdnh1GJgD4TbQkXL 1-14; HqW4rjd6XmYgDM9F 1-2; HyjNrJQMCGA2tvgp 1-2; J3tdCaXR8hAP9m2M 1-4, 6-9, 11;
J9jRmkxvtXcAKfM3 1-4; JB7fzC2ZnP3Kwjxd 1, 3; JD1ypjL3xcCKfTGP 1-2, 4-6; JNZVwGWTBMYLz9an 1-13; JPB6NXkAGCQzmh38 1-5;
JZx3p6Nqdj9tCrMc 1-3, 5; JvdQP6VyY19kZDXm 1-9; K3jWRx2qQ1aXy9PA 1, 3, 6-13; KG4FcnrYWVHCfzvt 1-9; KNqTmpW6HZCnc378 1-4;
KXmBtLjWnQ6zZa21 6-14; KmahZwCJcYFLj1Rq 1-7, 9-16; KmjAV91JBhgQLrfP 3-4; KvF7tTamGpn9kW34 15-24; LN78yvQwgJpcAYCG 4-6;
LW4qmQykhj3p9atv 1-2; LWx7bgKmpQ6ktYRr 5-12, 14; LaMmthYQK4HRWwXC 1-5; LaQztr3mDTwHyknB 5-8; LmDb8njHfXvAgC79 1, 3-42,
44-51, 53-63; Lqg8NxGDmjX97bRv 1; Lqnpfy2wZNc3DACH 1; LxdYWR47ZFbTNBhM 1-2; M4CvXLZz3qm6pDRc 1-3, 5; M9VJ7nxjaZW8wN2G 1;
MCdZTywRz3HYnvWg 1-10; MKk2adJtjyANf8LW 1-13; MW1Kw23ncjRDrxpA 1-14; McNx4Dwk89vYJpdH 1-5; MdaKf7zpNgWHF1Jj 1-4;
Mf1AyXnC6W3dpNJR 1-12, 14-16; MjVNcrtDdRC3qFny 1-13; Mjbx1mPBCwFHgLWp 1, 3-4; Mxafv6rLGQw4TjWb 13; MyZ8p6RaAFBWcJ1f 1-7;
NLzDkMAj8n2TmbxV 1-6; NWB8HZd1JvxzGjDr 1-3, 5; NZcYhqPxkrJMKz6a 1; NfLb7C1qDZ8cHxFp 19-29; NwKVWqHfgk1jLvAG 1-8, 10-13;
NyZgJMVBXHn8vtWk 1-8; PFGvpDf9WZgcY6Rk 1-12; PHqvNCxT1JVtpWFj 1-4; PN64hvHQLZRCptGg 1-5; PNnXcA9hFBx7Wtdj 7-11;
PadwhAr6VnpkvZyN 1-11; PhBmDkGLadY4fHb1 1-2; PyH9Adag8VwRjKJh 1-16; Q1arfj4WPBNXqmLk 11, 13; Q3kAgpdcGJ27VPvT 1-3, 5;
Q4GadwADtxbFBNW9 4-8; QAc4wgTL7PfDphk6 1-5; QMkfqDhm3L69A7VY 1-11, 13-14, 16; QRbmZWh2yDGakt9Y 1-4, 7; QxJKAphaZt61kzvg
1; QyYtrwM6zKPXHDj3 11-15; RDPLkYamNgHKXFh2 3-7; RbLqYxNQj1MaWFdP 23-29; RdmgVNQ4XJMfnC1F 4, 6-10; RfxgwvQjA1pVy98k 1-6;
T2MrDwA9hdgPXxRv 1-4; T4pxtrKv178PGNkA 190 fights from 9 to 208; TDHPhpVmGJ63LR9y 1-5, 7-9, 11, 13, 15-16;
TgVP6hxZKqY7MGQ4 1-5; ThMBCkJYw76c3jKP 1, 3-4; Twr9aYQFMWjnLPqR 1; VL1kvAqnwrzKQfH3 1; VMP6Dv2pyTnYhaKJ 1-3;
VRM3KtXTh2W7kzJx 1-4; VYb6f27NazpmCch3 1-2; VftPp1r2mC9GvJZ7 1-2; Vh3bzdRMP6ymNKq9 1-2; VhHLyw9CP76Wv3fX 1;
VvmGw1CT7bZMNXp6 1-12, 14-19; VxL3N2FaHmypYjnb 1; WAKwhqRtmPGT49ZQ 21-35; WFJf43apKLXHVzBh 1-4, 6-7; WqBa2AwcGHfgj8Lt 7;
Wza8bdXVRhcgqGnZ 1-5; Wzn2yfYmrNj9xG31 1-6; X16Yp9FzW87cNADq 1-5; XYTRKMVFWdypLxtB 1-14; XhZAV37vMFbYWRt6 1, 3;
XwqT7ZK48Df69tkh 14-22; YJzQRVL4r9DqFcpf 1-16; YRb7P9QzAXTBdHyG 1-4; YcgyKZVCJvbW7Hj8 1; ZCRLWJqmb7v3gjaN 1, 3-7;
ZKLvG3Bmy9CAgjT1 1-7; ZKRmDGdV8FM1aYwt 1-13; ZPK6pnD7HzmyhvjT 1; ZPxzwLAmBpRMTNbC 1-3; ZPyMjD1FkVGJH8Lb 1-4;
ZjLXvY6RFdCgqp87 1, 3; ZqaKFPJXQ18rWMck 1-5; avQywnDKt7RxcpfF 1-8; b12KD98fndF4hMq3 10-18; b4fNagK2jvXQJAGL 1-10;
bAZJWP3hgxRav6Fm 1-7; bNpCk1gG7y3LYrdP 1, 3-5; bQytk7r1Txa84DMA 1; bdPKyBjRNDX9CHaQ 15; bpV7JWX8B9xgrj3c 1-5;
bxywkpHBTzg2mtCQ 1-20; byYP67TH4KtZqm31 1-13, 15-17; cBjY4znbPJNH3XZv 1; cLkwmZbzFDKdPhYn 21-22; cPHdgRMCy7ZGDqFf 1-8;
cT8wJVq316kAWpn9 1-3; cVPtG1agnWhfX2xC 1-7; cXCfvnWtNrR63PVx 1-2; ctAnYvNLQK6F38PD 1-2; d8KLZbmzWhwr9at7 1-4;
dXR2BPqTYQ4aMxZj 1-5, 7; dtHTkZgMNvR3YBK4 1; f9vADt2CZ3M1ygwq 1-11; fRPcJYzWKXBCv32N 1-2; fjpy1TBL4b3kgJhr 1-8, 10-15;
fjqQNyMbD1Y7R92h 4-21; fkD9m3dBKcGAyqRY 1-6; fnMH4F2pkGC8T6Kz 1-8; fr2WyGqPVYNTXd91 1-4; g2qW8NnDvCQ3TfVG 1-5;
g8b1MkJnKtyfzvPa 1-5; gFfpKhGR4yWknxAw 1-5; gGn1vQ28YxpFrmaP 1-3; gT2mW39CV7DMwhzx 1-21; gZmxq6CdtraHKW2D 1-3, 5-13;
gbFJzMXnTKkxNQY7 1; gdMNAjD2Kyrtqh97 1-8; ghbnmqx4Wz8Rj3Ck 1, 3; ghcTDmb871vtxyWM 1-5, 7-11; gn7Wva1FKAm8HhGZ 21-23;
gwkN81FaqdAPQTfW 1-4, 6-8, 10-11; hXT9DqC7vpQW6ZLb 1-9; hf1pF42aPHMv9kLX 1-9, 11-20, 22, 24-36; hmtTqH16pPGzZJax 1,
3-17; hvgw9t1par7YCD3b 1-17; j7J4bKFHq6M2Wfvy 3-6; jD7whNgBrM4GxVvF 1-8; jGrv6B1hNMDJPc3W 5-10; jHXnNA128R4FmwYB 1-5;
jQzWyR89avmgpd7H 1, 3-7; jYAhJZ3BzPMHWx1q 1-5; jnxaGm43ctgQMFhf 1-8; jtdGCBNJq38zxcQ9 1, 3-9; jyxcbwYXG3pL9NnP 1-2;
kHfNcJTYj3RtMFGL 1-3; kJHLMmf3rGaj28d1 1-2; kX94Vmn8tpJzGTD2 1-11; kqXb3Tja4gZfxP7V 4-6; kzLTMb41xj8WgpHX 1-5;
mDbPCdX9AvwYJxGa 7-30; mMFCR2atzKHWDcdQ 1-6; mPMpkAgjbtR7HBXD 1-19; mTCGfJXzxLrRMDkc 53-66, 68-88, 90-97, 99-101,
103-119; mc821XQyrpDjGJh9 1-3; mjbfcTLkM2PGKadw 4-7; n3hjGAFXxTaBfmRz 1-3; n9d2Fv4R8Ktz6BkC 1-16; ndHRLVB9Kaz7xpqT 1-2;
nmrNLB3j4pAPg8fQ 1-2, 4-15; p4ZxFNjQqXYCg7HP 1-7; pKn8wBZcQ62PyJH7 8-24; pPr6cRXnaC3B8WAZ 1; pfxzGqDBYgWKnF73 1-4;
prbPqjNgMz6FaBX7 1-7; pz7YaGwxXZgcbQPm 7-9; q3W8BmALpaj1zGR4 1-2; q9FHYznbZLDKtQTP 1-2; qLk1fVdN3rcaDxm6 1;
qMD3CVapwb6nAgx2 35-40, 42-43; qQhkt3L9yKfTFp2D 1-6, 8-13; qYKvbL9c6tg4zCnp 1-2; qYRLwvVKDkTZGP86 20-43;
qd4XPg1TWQNrRGBL 1-9; qg8YWNaCdDvV7XZM 1-8, 10; qyWNYFVZjLX1mtJA 1-19; r67HcKRnkYdzTQZP 1-19; rP3VYNn8mhAvaWpZ 1-9;
rbg3wR9B7dKf8JmW 1-3; rpLZ287nfKWyJGtb 1-6, 8; t7432ga9HXvFBKzy 1-6; t816jCfw42DHF7dv 1; tD7Rm3GBjdqPzhwC 1-4, 6-8,
10-14, 16-17; tLXnZDMCvJ34mzbF 1; tMg1aAHQGJkxpYcb 1-12; tc1qGLjpgVrJ8WdQ 4-5, 7-15; th1nf93NzQjmVdrD 1-15;
tjN8BgxX7k4GyAfd 4; tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19; vCZk3N8n2dbTyMFV 1-34; vDg3LH1Zyfz49kKr 1-12; vFqn26gPR4NJBfjw
1-4; vGTRdWMxZYChHtzw 1-5; vKznWFRC7cBHTGwx 3-6, 8, 10-11; vpBwx4KbaWzRZ8V7 1-3, 5-10; vtRzBNFnW6KH3pjT 7-8;
vzZPBtWcfm9HY8Ty 4-10, 12-18; w3L1hH2ya4gCTKAq 1-13; wCGWfTxNgFrB8hDy 1-15; wHtnq3pgjxYVy9FZ 1-2; wLvayNDdrJnGPWhZ 7-9;
wRY9NtQp7bPZ1VDz 1-6, 8-12; waYzfQgyrcKV7H9J 1; wfgxbHpNB2vZJYLW 1-6; wvbng2zVTYRrc9Zd 1-2; x7ybf4LjBD6GJqWn 1-9, 11-15;
xDwACkbMZRyNcd1z 1-5; xJbw6RTvtHPncNzp 1-5; xRNZwp7fWvTY29qV 1-15; xbCaBqNdAn3ZWKDw 1-3, 5; xmYzGavFHWpfbLkt 1-3;
xnrg9m6jDMR1kaLN 21-23, 25, 27-29; xrtBTqdwGDfzK72Q 1-2; xtFwVRc3gmdjLAyC 2; xw86k9LyMNF3YHjr 7-14, 16-21, 23-31;
y6kYARn4dpCKavZ8 1-3, 5-6; ycjXJnNkC8pGWH1Z 2-7; ypQ9d13tMFJPNfvw 1, 3; zDhrvnfPYdt9FXMj 1-4; zjxyZAgPDBMGndRK 1-16,
18-20, 22-24, 26; zwPKLTQ4W6YkyRqd 3-8.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1292679
- Rate: 1.8 PPM, measured

2,063 procs on 20,696 rolling hits, 42 wearers, 47 runs, in the main hand: all hits 1.81 [1.74, 1.89]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.86 [1.72, 2.04], 458 procs on 4,487 lone units, 39
wearers. Specials: 0.00, 0 procs on 1 lone unit, 1 wearer. Travel time: the proc lands a median 73 ms after its hit,
98th percentile 182 ms, window [-20, 192] ms (426 isolated hits). Dropping the swapped stretches (41 active minutes, 3%
of hits, in 3 runs) moved the read from 1.76 to 1.81. The paladin check, which sets no rate: lone white swings 2.24, 119
procs on 967 lone units, 10 wearers. Left out: 40 of 87 runs in Cat or Bear Form. Report L42wz7rMKXc8gHAx logs the same
run as bJm837KZrhMDzNwG and counts once. Reports from 2026-11-13 to 2026-11-22.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Shaman |      41 | 2,046 | 20,532 | 1.81 |
| Druid  |       1 |    17 |    164 | 1.88 |

Reports and fights: 1QaNtF3jmdkPn8wT 1; 1WKqM9x42jcLnYCN 1-36; 1hpgVjWH3GZBLYKw 13, 15-25; 1nLR3XCxjcAdpKtB 1-18;
1q2tTDcaFMnfzvkB 1; 29DARYyHxP7Gvp6T 1-12; 2FC3qQxzBbjpwAtG 1; 2NQC6YJtXRPbMzg1 1-13; 6r9Fzq4Gfa1Pjh2V 1-3;
9RpHrj4Ay8zmxvaB 2-7; 9RrQv7fabk4PhwL8 1-6; BzJTG7nxPmD13LbV 1-7; JA3cCVyvwX4kx2Q7 1-4; L42wz7rMKXc8gHAx 1-3;
LQAvZyWn8zw19bqp 1-5; LzcCRM4GNtK1HdJf 1; MfD1n7XFZrG4dcbH 1516 fights from 1 to 1743; NWQqXLPdxrZ3DG1B 1-16;
P2Zf7rRDHLhByVt8 1-17; PcyxnCJVgWQRpXtD 1-5; PqLkxdaC6mJDRH9p 1-4; QBdcYLvphWtTy3Rg 1-2; RqpmbCcYr8BJZWwK 1-2;
TC3Vv1WJBar9dXwQ 1-2; Txgtz1nP7RjkrFZX 1-2; VxL3N2FaHmypYjnb 1; X2wPBN4gvZQynAT9 58 fights from 1 to 73;
XPmJACH8RtqFd3jM 1-3; XqDBC3bGkvTNQfj6 1-5; Y7yP84t1TJZmwXLN 1; aPLxyc4GrBXF2gDK 1-3; acD7VhMJjkPpXbRr 1;
bJm837KZrhMDzNwG 1-3; djLgyY13PQ6W8Vth 1; dyTRnv6MZQ7BKaGC 5; fV4TkpzX8NgHLbZx 1-4; gFfpKhGR4yWknxAw 1-18;
hy4nZYdFGMDVf8j1 1-3; jRPy67dbq91HKCMJ 1-17; kY1d6fcDyWgrChmZ 1; m7Jbx2HtjdwWgkqK 1; mByrn4hLG8pFbK17 1-11;
p9HgV7zdCKqbrwJ3 1-3; pKWPQTCB317YgMdX 1-6; rcdBtkqxjy7zfTmX 1-7; tNBr2jAPvCzk1nRc 1-6; tbpKQy1rTjdJYL4V 1-17;
xQv1aRGzXdpcKZjC 15.

## Venomstrike (6469)

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 29653
- Rate: 1.85 PPM, observed

6,276 procs on 84,307 rolling hits, 123 wearers, 144 runs, in the ranged slot: all hits 1.86 [1.82, 1.90]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it. Lone Auto Shots: 1.89 [1.69, 2.10], 346 procs on 4,585 lone units, 121 wearers. Specials: 1.88 [1.73,
2.03], 540 procs on 7,193 lone units, 107 wearers; Multi-Shot gives 462 procs on 6,162 of them. Travel time: the proc
lands a median 1033 ms after its hit, 98th percentile 1919 ms, window [-20, 1929] ms (461 isolated hits). Dropping the
swapped stretches (546 active minutes, 15% of hits, in 22 runs) moved the read from 1.58 to 1.86. One run is dropped
whole as swapped. Reports from 2026-11-06 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Hunter  |     112 | 6,271 | 84,243 | 1.86 |
| Warrior |       9 |     5 |     60 | 2.08 |
| Rogue   |       2 |     0 |      4 | 0.00 |

Reports and fights: 12VhLkPa98DYFH4g 1-6; 17paTd8tYnM9zjxZ 1-10; 1BWnTkwG7aFvp6Hc 1-3; 1Rk8y7VKNWapF2dt 3-4;
1bxfCJMQhwnra8yF 1-2; 1gKMDwaxfPv68FH9 3-4; 1jKJfWybdV2vBQ8A 1-2; 1qNaRQCwLmDPcJfx 1-2; 1tLCfdvQnH4mT9BY 1-2;
23dT9fk1DqcaJmwn 1-7; 261hbMqpgzLRKxwY 1-2; 29pRqaycxtzTW8jC 1-3; 2axJHNwqQdtW1jnF 1-3; 3CyQk69LzJjTn2X4 1-5;
3npHxB84DT2zRtmN 1-4, 6; 42RqpLbKHhmryvTD 1-2; 4Brq3tnNJ6afk8Cm 1; 4JAHxYwDKQv3Nyjk 1; 4dvhFmKbqkTQ1Wpc 1-3;
4gaFjvryzRP1ZfhA 1; 4ntmx3RVAraQMh2C 1-2; 628JmtypwxXTK4MY 1-6; 7W3YTHD9kyPzFvtc 1-4; 7ZxcPNTjqvfzQHgX 1-2;
7dNXWFTq2R1zDa8m 1; 7ypMG8JVjRgZ1xHQ 1-9; 8kFvhK9NtCBP2aYg 1-5; 9GMN2DQLhrB64Atv 1-2; ALMXvaHbp2gn4zyG 1-3;
AWZXzaxq1dJ2BgPT 1, 3-5; At4HT7ZfYzFVwybh 1; AvtKjphWbBxNH3rV 1-3; D1qH8nWkjQBLxPbX 1; D2W68v13pkbF9Nmh 1-15;
F9DyXdCazKmPqWfN 1-7; FNkQ3BDhAfzM9aZP 1-7, 9-13; GPhAqKkF8tHBpfxT 1-3; GQFHDcXPtrgZxC2K 1-3; GTDwzbAMdXWZr7JP 1-2;
GmZaJgjW2R97cVnL 6; GyLWvrkxKVmNZgw3 1-7; HR1dVgK78CvFGaDA 1-5; HjBPLYwhmZprMQ12 1, 3; HxPZVaBGLcKXg4bJ 1-2;
JD2ag1wYV8qCxhtQ 2-9; JQH2y6NWBXchGRbg 1-3; JcwBmYGVgnMQH9xd 14-19; JdZYhLaPmtH4j1qK 1, 3-7; Jh4DQ8XH3LBv7CjW 1-4;
K1XzMGFDwrt7RQbv 1-3; KYpqanLMZ4Q2JtVF 1; LPTFCVQ4m6Yxn1cz 1, 3; LkjGgT6V9JbNqRQD 1; MY6ZgqyzPDk9tXN7 1;
Mj1gA4rpDRQh9yJm 1-3; NGhbmnMwKCHYfaX6 1-5; P6WhN28ZJbLAdtfm 1; PAH74TNnB9aYpm1C 1-4; PCYpB6DhnQMmXbzN 1-2;
PTGqdrR3Xz8gDaZ6 1-4; PdLaqgMzj1c4Kf7b 1-3; Q6PwqmyzMgv8AnGX 1-2; QDPGNp2JVHv9rhL4 1; QjLcYKby4Tk8ZDCt 1;
QrqZXKL9kbnY31Hj 1-3; QwzkrjDGHnW37qpv 1-8; RQxLGd1agBDt6hXP 1-4; RyaVqvxGnD8bKj3w 1-5; TxBYKw6QptvCLydr 1-2;
VHW3CYz2kjM9QTfm 1-7; VbxMRK92fYk4XH7g 1; VgxzmhZ72P4CNypJ 1-3; W2jZkgqyPYnNDXrx 1-2; W416FHmftD9rPRgA 1-2;
WhwcvTCZFPDxanyr 22-65; XKAja9wcpndHyZRJ 1-5; XyMzLZN3cf8JQBmT 1; YCqznXvDMQA4d718 1-7; YMJmhPLVFWHrQyxK 3;
ZA8dXczgaKMTt1pf 1; ZwBW1aVvkDAfcN3z 1-4; aFYgR2dpmqD86wMc 1-2; arQtRhA73GXckn8W 1; b8z7yMrKvVZCdPX9 1; bwByj2vLFnhC7cz1
1-2; bzWMAC2mf7Z6Q3LV 1-9; cPkR9NjHgVnJqMra 1; cQXqdZgpwNPytG6b 1-2; cVG3wHBmnqFN8jxJ 1-9; cWAqMk2jnJF4DvRm 1;
cYgfxhMQKZNRGaX9 1; d6bwAkLKjGq7pVMy 1; d93x2NC1X8FhLrfV 1-2; dTh4GAWLxwRm9V6k 1-2, 4; gAjMTrvJV9ywK7ZL 1-17;
gatpqM6JWKTdD2CH 1; hG8tvQjLBJgf4WHY 1-2; hX7P1FZKQmqTyfBn 1-4; hbt8kZM4zrFVXnjf 1; j6QM3BAp4GZxN8zY 1-8;
jM3dtQ7vcGZ9fhWn 1-5; jqpHX1N6Q9VyaMWd 1-5; jyDXpZRxJQf86q4w 1-2; kBzVYb3Cy7QwnKMD 1-6; kKjwaXP6vgCzW1fn 1-2;
mAv6knQrDRWc9PYt 13; mpckQL768Tgyv3jF 1-2; mqKLDBajcQxkVYpP 1-2; nA8bZqcCQYrV2BjP 1; nT2Wz7mVQZbtBRG4 1-4;
pbyJrkwLXGWg2VxN 1; pfDbnAPTwzLHxY4W 1-2; pwd6RvtM2NYqmfGT 1; pydvC4YzqVMRgPAf 3; q4xrRhnT2KVWJDLQ 1-3; qDvk6HmdP3QBgbN8
1-3; qJNfkaCxZF3DzBh6 1-5; qx4z6wYFNprn7M3V 1-2; qxACn7fMgDQL3Vw6 1-9; rXvTBhVgpZmc1kbP 1, 3-13; rZ1A2XtkJjLfQb9x 1-4;
tLbwmqhCrWv7BnyT 1; tLzGhTR6kH4cYraC 1-2; tWdPKc8M7nxpJTfk 8-9; tv1PgkxVBcK74zM6 1-3; tvrgBcqMzCRk1wF8 1-2;
txZvdrQzpwn3V4kb 1-4; vWGgArDXtc13Lx6J 1-9, 11-14; vfY9yp78NgAatjPX 1-14, 16; vgXDVpPfw1QnWjYz 1-2; vjhNJpcTVr4yYbqX
1-2, 4-5; vwpTh7B8kX16ynPz 1-12; w48cXFqxMbWTH2ky 1-3; w9khJAXpzqxGVmN6 1, 3-7; wAQjaFpY6Kbdtg1W 1-5; wW9ZKjH1qXPzaJnx
1-3; wYmr1kfWPpVz7jyZ 1-4; wkrz3qt6MQfZPFW7 1; xph9MnBHyqRvcYkt 1-12; yXvdkQ8ZYCTbxm2F 1; zLbDKmfkrhARcVJ1 1;
zbV7pwYGARa4Tqgk 1-2.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293005
- Rate: 0.95 PPM, observed

247 procs on 9,528 rolling hits, 15 wearers, 16 runs, in the main hand or the off hand: all hits 0.97 [0.85, 1.08].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it; the interval is wider than 10% of the rate. One-hand white: 1.08 [0.79, 1.43], 47
procs on 1,634 lone units, 4 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 0.72
[0.60, 0.87], 69 procs on 3,589 lone units, 11 wearers. Specials: 1.12 [0.61, 1.87], 14 procs on 470 lone units, 4
wearers (a Poisson interval: under 5 wearers). Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 106
isolated hits. Dropping the swapped stretches (16 active minutes, 4% of hits, in 1 run) moved the read from 0.94 to
0.97. One run is dropped whole as swapped. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       8 |   123 | 4,781 | 0.96 |
| Warrior |       6 |   122 | 4,706 | 0.97 |
| Hunter  |       1 |     2 |    41 | 1.85 |

Reports and fights: 4v9Y6CNdXyD7P2rq 1-8; 4zC92VDj7dYATRqv 1; HpCXWPnytdkvzqg2 1-8; JjXZWnxmwCa7HQDq 1-3;
MKx2v8JfHVc6kmND 1-6; Pzk84bCLKYcvN9fx 1-10; QBCAa6JwknGvcMrP 1-2; dNwLHJ3Pv4M279xr 1; fA1GLbc6gw9nKHM2 1-16;
fLRW43HDJY2TP7Nb 1-2, 4; ryPR7YjFN89CKXJg 1-11; tjGLcNVJXD6r4mCz 1-2; wHPxZ7ABrtJV1Qgq 1-3, 5-10; wbmNGq93nM6kXxaA 4,
6-8; xg7G3pFkR6dzwvJc 2-8.

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

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 1293787
- Rate: 1.95 PPM, measured

755 procs on 7,079 rolling hits, 11 wearers, 12 runs, in the main hand: all hits 1.94 [1.86, 2.11]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.88 [1.69, 2.18], 156 procs on 1,509 lone
units, 9 wearers. Specials: 1.74 [1.40, 2.04], 96 procs on 1,001 lone units, 8 wearers; Cleave gives 42 procs on 382 of
them. Two-roll abilities, priced as one roll: 4.08, 26 procs on 116 lone units (Execute 26 on 115 and Mongoose Bite 0 on
1), left out of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 333 isolated hits. Dropping
the swapped stretches (34 active minutes, 16% of hits, in 1 run) moved the read from 1.62 to 1.94. The paladin check,
which sets no rate: lone white swings 1.98, 332 procs on 3,047 lone units, 33 wearers. Reports from 2026-11-14 to
2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   740 | 6,944 | 1.94 |
| Hunter  |       5 |    15 |   135 | 2.02 |

Reports and fights: 1WHqx7bPa8mkGydJ 1, 3; D6xj8ZGpKTBHbNqX 1; HRnmMtD7Fp41wb2g 1; K8rhAtLjcz7gdHkn 6-7;
P1yKzX3Q8NgFYxZH 1-3; VAcf9Xxdwb7QCjBg 1-2, 6-7; WYZRJjhp7LvDqKNB 1-7; X4jWcyZhzfmJKG9F 1; aPDTGRxFmMhZAzYW 1;
bP9nWzTVjBC2NxA1 1-6; mM3pZD8H7CqBNYwJ 1; maYP7WvxJ2TzGtDH 1-3.

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

## Count of the 2026-11-06 to 2026-11-21 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 652 reports in all
- Game build: 1.60.1.70790 to 1.60.1.71037 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC,
  71037 from 2026-11-20 18:23 UTC); 6 reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database
  is 1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-21
- Pull request: #1156

The count is EV-0153 §Count of the 2026-11-06 to 2026-11-21 reports, run on these weapons. On Kang the Decapitator and
Sul'thraze the Lasher every class past the first is one wearer.

## Count of the 2026-11-09 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 321 reports in all for the count
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 74 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is
  1.60.1.71037, push 114330; the hotfix each player ran is not recorded
- Date: 2026-11-09 to 2026-11-22
- Pull request: #1168

The count is EV-0177 §Method, run on these weapons.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 3,714 reports in all for the count, 3,153 of them for these
  weapons
- Game build: 1.60.1.70790 to 1.60.1.71037 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC,
  71037 from 2026-11-20 18:23 UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 585
  of them for these weapons. The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on these weapons.

On Axe of Inevitable Ruin, Barbaric Crossbow, Black Duskwood Staff, Mana Channeling Wand, Spellsever Crossbow,
Supercharger Battle Axe, Twilight Maul and X'caliboar, measured, the classes past the first give under 3% of the rolling
hits.

## Count of the 2026-11-13 to 2026-11-26 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 353 reports in all for the count, 310 of them for these
  weapons
- Game build: 1.60.1.70938 to 1.60.1.71139 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU, 71139 by 2026-11-25 22:38 UTC; its
  rollout time is not recorded); 202 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 167 of
  them for these weapons, and 66 after 2026-11-24 01:00 UTC, 44 of them for these weapons. The sim's item database is
  1.60.1.71139, push 114448; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-26
- Pull request: #1262

The count is EV-0221 §Method, run on these weapons.

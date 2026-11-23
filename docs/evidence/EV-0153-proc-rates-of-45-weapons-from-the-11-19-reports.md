# EV-0153: Proc rates of 9 weapons from the Warcraft Logs reports up to 2026-11-19

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it. Angerforge's Battle Axe, Diabolic Skiver, Plaguefang and Wraith Scythe are measured, and Archeus, Blade of the
  Basilisk, Bleeding Crescent, Silent Fang and Wyvern Tailspike observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every indexed wearer of the weapons; 2,155 reports in all
- Game build: 1.60.1.70790 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114196; the hotfix
  each player ran is not recorded
- Date: 2026-11-05 to 2026-11-19
- Pull request: #1108
- Later count: Blade of the Basilisk, Diabolic Skiver, Wraith Scythe and Wyvern Tailspike, §Count of the 2026-11-06 to
  2026-11-22 reports

## Method

A proc is an event of the item's proc spell: a damage event, or an apply, refresh or stack of its aura. Events of one
proc within 50 ms count once, unless the same spell, target and kind repeat (two procs that landed together). A proc
that grants a damage shield counts once per apply or refresh of the shield. A target is keyed on its NPC id and its
instance together, so two mobs of one NPC id stay apart.

The rolling hits are the ones the class census (EV-0155) finds rolling a weapon proc, each on its own hand's weapon:

- landed white swings;
- every other landed hit on the melee table the wearer casts, on the main hand: the specials the census reads
  (Sinister Strike, Eviscerate, Hemorrhage, Kick, Revenge, Cleave, Heroic Strike, Overpower, Slam, Hamstring,
  Stormstrike, Raptor Strike, Wing Clip) and those it reads on thin samples or does not settle (Backstab, Mortal Strike,
  Bloodthirst, Gouge, Riposte, a two-hander's Whirlwind);
- every landing of a melee-table ability that logs a debuff and no damage, on the main hand (Sunder Armor, Rend,
  Rupture, Cheap Shot, Kidney Shot, Expose Armor, Disarm, Concussion Blow);
- Windfury Weapon and Windfury Totem hits, once each;
- a hunter's Multi-Shot, Arcane Shot, Aimed Shot and Summon Hawk hits, and the landings of Serpent Sting, Concussive
  Shot and Scorpid Sting, on the ranged weapon;
- a dual wielder's Whirlwind, once per hand: each target takes both weapons' hits at one millisecond, the larger hit
  rolls the main-hand weapon and the smaller the off-hand weapon;
- the strike chain (EV-0145): the first landed melee-table strike of a weapon proc within 30 ms after a main-hand white
  swing is one more roll. A magic proc adds none, and neither does a strike with its own cooldown.

Sap, Death Mark, Shield Slam, Shield Bash, Shank, Thunder Clap, Volley, periodic ticks, damage shields, Deep Wounds,
Sweeping Strikes hits, poisons, imbue strikes, pets and spells roll nothing.

Execute and Judgement of Command roll twice (EV-0155). They are left out, and so are Mongoose Bite and Mocking Blow,
whose few hits do not settle whether they roll once or twice: their hits count as no roll, and every proc from 20 ms
before one of them to 50 ms after it is dropped.

A proc that lands on the target logs almost never on a killing blow: over the 45 weapons #1108 counted, 3.7% of the
white swings that leave the target alive are followed by a proc and 0.7% of the killing ones. A proc that buffs the
wearer logs as usual. For a proc that lands on the target, a killing blow (a hit with overkill) is no roll, and a proc
from 20 ms before one to 50 ms after it is dropped; a killing blow is no lone unit either. Some reports list one
stretch of the log under several fight ids, and their events come several times over: a run whose white swings repeat,
to the millisecond and the point, keeps each event once.

Paladins are left out of every rate, since their rolls are not settled: they read above the other classes on some
weapons and below them on others (owner ruling, 2026-11-19). Druids in Cat, Bear or Dire Bear Form are left out, since
they swing the paw and not the weapon. A run logged in two reports counts once. A stretch of activity with no proc where
several were due is dropped as the item swapped away, from the hits and from the lone units alike.

The rate over all hits is the procs over the sum of each rolling hit's weapon base speed over 60. Its interval is 2,000
bootstrap draws of whole wearers; under 5 wearers the section gives a Poisson interval as well. Hits split between the
hands are fractional, so the class rows are rounded. A class row counts every wearer of the class, those with no rolling
hit included; a section's wearers count those with a rolling hit.

Three rates over lone units check it. A lone unit is a hit with no other event of the wearer (any target, ability, tick,
debuff or heal) within the proc's window before or after it, and a proc in that window counts once.

- One-hand white: lone landed white swings of wearers with nothing in the off hand. Dual wielders' lone swings are split
  between the hands by timing and damage; about 10% get the wrong hand, so the dual-wield white rate reads low and is
  listed but never compared. For a ranged weapon, lone Auto Shots take the place of white swings.
- Specials: the specials the census reads, listed above, and Rend's landing, on the main hand; Multi-Shot, Arcane Shot,
  Aimed Shot and Serpent Sting's landing on a ranged weapon. The other melee hits, Sunder Armor and Windfury hits are
  not in it.
- Two-roll abilities, priced as one roll: a true double roll reads about twice the white rate. This checks the roll
  table and sets no rate.

A direct proc logs within about [-10, 100] ms of its hit. A proc spell whose game files row has a missile speed lands
after its flight, and its window is the 2nd to 98th percentile of the delay after isolated hits, plus 10 ms.

The rate is the rate over all hits, rounded to 0.05, where the one-hand white rate or the specials rate, each with 50 or
more procs, agrees with it (their intervals overlap). Where they disagree, the rate is open; no weapon in this file is.
A stacking debuff refreshed at full stacks often logs nothing, and a proc whose travel window passes 400 ms or reaches
past the isolation has no usable lone unit: these take the rate over all hits as it stands, and so does a weapon whose
lone-unit rates stay under 50 procs; such a rate is observed, unless a separate count checks it (Spellsever Crossbow and
Bow of Searing Arrows, from #1108's review). A rate is measured when a lone-unit rate checks it and 5 or more wearers of
more than one class give it with an interval within 10% of the read either side, and observed otherwise.

Scripts and inputs, in `forever-ppm-review-1108/final-1119-v12`: `analyze.py` for the rates, `MEASUREMENT.md` for the
method, and `targets_final.json` for each weapon's rate and call.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon                  |   Item | Rate | Call     | All hits [95%]    |   Procs / hits | Wearers | One-hand white | Specials   |
| ----------------------- | -----: | ---: | -------- | ----------------- | -------------: | ------: | -------------- | ---------- |
| Angerforge's Battle Axe |  11816 |    2 | measured | 2.01 [1.93, 2.10] | 1,068 / 12,283 |      14 | 1.70 (298)     | 2.25 (27)  |
| Archeus                 |   2000 | 0.95 | observed | 0.94 [0.81, 1.08] |    221 / 4,268 |      27 | 0.87 (69)      | 0.87 (32)  |
| Bleeding Crescent       |   6738 |    1 | observed | 0.99 [0.90, 1.09] |   482 / 12,205 |      31 | 0.95 (69)      | 0.87 (50)  |
| Plaguefang              | 279876 |    2 | measured | 1.98 [1.87, 2.11] | 2,944 / 42,419 |      55 | 1.99 (560)     | 1.98 (324) |
| Silent Fang             |  13953 | 1.45 | observed | 1.47 [1.30, 1.61] |   466 / 11,915 |       9 | 1.50 (154)     | 1.49 (64)  |

## Angerforge's Battle Axe (item 11816, spell 1300783): 2, measured

1,068 procs on 12,283 rolling hits, 14 wearers, 14 runs, in the main hand: all hits 2.01 [1.93, 2.10]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 1.70 [1.41, 2.00], 298 procs on 4,035 lone units, 14
wearers. Specials: 2.25 [2.24, 2.56], 27 procs on 277 lone units, 2 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 2 lone units. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 0 ms, 257 isolated hits.
Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      13 |   965 | 11,204 | 1.99 |
| Warrior |       1 |   103 |  1,079 | 2.20 |

Reports and fights: 1GBhzTR8CZwcJfFV 20-32, 35, 37-38, 40-51, 53-63, 65-66, 68-82, 84-88, 90-99; 4rfgMCY9qjyvQhHD 1;
8CbGckB9gavMh6QW 1, 3-6; AfJ6DYbh9dmkgHZL 1-2; CYbmtxQ1THrG248Z 1-2; GWL9jZvR4rpaKftb 1-4, 6-8, 10, 12-13;
HZpdWaVtA8TM6bqX 1; Kfk36GzDwnPVbHaN 1-18; cgmhtvNdTGCn7f6L 73-84; dnDg7j9kRw6WtbML 55-61, 63-69; jKQDqkwmx46ZLBFg
17-24; jr7TAd93qLPxWzhp 1-2; zmDY4LVX7N8vh9aR 1-3; znBQHJTxXapkRLt3 1-11.

## Archeus (item 2000, spell 18091): 0.95, observed

221 procs on 4,268 rolling hits, 27 wearers, 30 runs, in the main hand: all hits 0.94 [0.81, 1.08]. The rate is all
hits, which the one-hand white rate agrees with. Observed: its interval reaches further than 10% from the read. One-hand
white: 0.87 [0.70, 1.09], 69 procs on 1,437 lone units, 26 wearers. Specials: 0.87 [0.60, 1.24], 32 procs on 670 lone
units, 23 wearers. Two-roll abilities, priced as one roll: 2.11, 8 procs on 69 lone units. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 1 ms, 108 isolated hits. Reports from 2026-11-10 to 2026-11-19.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   176 | 3,427 | 0.93 |
| Hunter  |      21 |    45 |   841 | 0.97 |

Reports and fights: 1NRQ9tBnTrq24D3k 1-11; 2qk4MYw7cr8KGFtm 1-8; 6z9hqHARQr3vwLXd 1-2; 76ZV8FQ1RHDhbq3z 1-4, 6-9;
FRxmtWLHr3wvB8Q1 1-8; FfAC3tWBxbp2HV9j 1-3; Hgp61KVdvPqYyr2N 1-11; KBZyYLg1WTqm6Chw 1-14; MmqC2hGvFY7x6Dgd 1-3;
NFD1RP4LdQrhfw3W 12-17; TDjrhyCkw3P9YApc 1-9; W6nhHydV1wpMmZ7a 4-5; XvWP1x3FkAzwLHZC 1-9; aQGPVW4TBcv8y3wD 1-2, 4, 6;
b186j7df4hKxNPkV 1-4; bPqj8zkJN3YatpwD 1-2; cvfBaDxQX7K2zFtk 1-7; gr9zjCAX3mY41aqf 8-12; kJq81KxZyTHD3A6z 1-3;
mtBc1WZf3ydHVPFv 1-6; pdqmMHchNGQjTwgk 1-3; qpXkvRdWxKzfNr2G 1; rNXvGcnZabhLJd9A 1-2; rQdnmBHXCpKgJqw4 1;
vN8WfZatBGgdCDHJ 1-4, 6-7; vRPQ237fWm1tFYDV 1-4; vTX3cKpNbY6jJVDa 1; wMRNLvP8CBrgd4x7 1-4, 6-13, 15-17, 19-26, 28-39;
z1APqCWaMYvxbkd7 1-3; z4WhK2CnLbaHgNVQ 1-6.

## Blade of the Basilisk (item 8223, spell 10351): 1.05, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

230 procs on 6,787 rolling hits, 6 wearers, 7 runs, in the main hand: all hits 1.07 [0.98, 1.20]. The rate is all hits,
which the one-hand white rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand white: 1.09
[0.98, 1.22], 67 procs on 1,943 lone units, 5 wearers. Dual-wield white, which sets no rate: 1.18 [0.66, 1.94], 15 procs
on 403 lone units, 1 wearer (a Poisson interval: under 5 wearers). Specials: 0.94 [0.76, 1.14], 32 procs on 1,080 lone
units, 6 wearers. Direct: window [-16, 100] ms, median 0 ms, 98th percentile 0 ms, 88 isolated hits. Dropping the
swapped stretches (23 active minutes, 9% of hits, in 2 runs) moved the read from 0.97 to 1.07. Reports from 2026-11-16
to 2026-11-21.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   193 | 5,664 | 1.08 |
| Rogue   |       1 |    37 | 1,123 | 1.04 |

Reports and fights: 7fgdkyv1njDKa8Gc 6; 8kFbLv24Dz9QYrdc 1; 96pBnGdAKJMaCb3N 1-12; hYHBQ9ZJX3vFAnkj 7-14;
qa1tKFnvZ98hwj6x 1-6; vYmFJGLbwPWNcdxA 1-13; xH8TDPQKb6mYragf 1.

## Bleeding Crescent (item 6738, spell 16403): 1, observed

482 procs on 12,205 rolling hits, 31 wearers, 32 runs, in the main hand: all hits 0.99 [0.90, 1.09]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. Observed: its interval reaches further than 10%
from the read. One-hand white: 0.95 [0.79, 1.16], 69 procs on 1,817 lone units, 10 wearers. Dual-wield white: 0.76
[0.53, 1.11], 62 procs on 2,041 lone units, 18 wearers. Specials: 0.87 [0.57, 1.13], 50 procs on 1,445 lone units, 22
wearers. Two-roll abilities, priced as one roll: 2.10, 11 procs on 131 lone units. Direct: window [-11, 100] ms, median
0 ms, 98th percentile 0 ms, 166 isolated hits. Dropping the swapped stretches (99 active minutes, 21% of hits, in 5
runs) moved the read from 0.78 to 0.99. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |      11 |   280 | 7,598 | 0.92 |
| Rogue   |       3 |   117 | 2,209 | 1.32 |
| Shaman  |       3 |    46 | 1,265 | 0.91 |
| Hunter  |      15 |    39 | 1,133 | 0.86 |

Reports and fights: 1WC3Bnw4A7RKpLDj 1-4; 1rdmQ9wptPWHfNG6 1-4, 6, 8-9, 11-23; 37gGDYxn4ZKLbMV1 1-7; 8KwxgLPtkpD7h2Ry
4-5, 8-32; 9CGYv1Jp24aWyhf8 1-2; 9mZgR7wqPxWMat6L 8-13; AwkvX3Tt1jqLhMVc 1, 3; B3vtVgmqMTHDAxNY 1-10; CLbmK261T3d9gAR4
1-6; CNWJgkfyAhjB3vRt 1-11; FbNk9QwxY7nG618P 1-19; H4NThcMa2dfR8gGr 1-2; NgAWV9MGwvbFLBC1 1; NwKVWqHfgk1jLvAG 1-5;
P8Znjv1fgBD2CAVY 6-12; QDVfcwHpjZRzL94q 1-6, 8-13; RPnVYQcG9xTJh1ar 1-4; WnvfNJK8dQG9xPMY 1-2; XgZdn3D9KR8cfLNY 1-10;
YhCcTtxfGJR82M7Q 1-11; ZAWJk6BdKhr9nC8T 1-6; ZyGV2gN1w6jc7zM9 1, 3-5, 7-18; f2zK39XdhnNqZJgR 1-2, 5; fjb2rRngTPa48WA9
1-4; h92ZHrfm3RFQtYvT 1-8; jC4HdxZGVNLz1Yvm 1-2, 4-6; jGrv6B1hNMDJPc3W 7-10; kmBptFTG68Dcq3jx 9-10; mQ4MqkD1nbaBXLdy 2,
4-6; npbfVjrQdDLhtRWY 20-22; pKn8wBZcQ62PyJH7 12-24; tD7Rm3GBjdqPzhwC 16-17.

## Diabolic Skiver (item 9475, spell 18206): 1.8, measured

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

545 procs on 6,252 rolling hits, 24 wearers, 26 runs, in the main hand: all hits 1.80 [1.65, 1.93]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 1.99 [1.73, 2.27], 202 procs on 2,102 lone
units, 23 wearers. Specials: 1.82 [1.31, 2.22], 68 procs on 772 lone units, 19 wearers; Raptor Strike gives 26 procs on
244 of them. Two-roll abilities, priced as one roll: 1.70, 6 procs on 73 lone units (Execute 4 on 37, Mocking Blow 0 on
1 and Mongoose Bite 2 on 35), left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 294
isolated hits. Dropping the swapped stretches (5.8 active minutes, 3% of hits, in 1 run) moved the read from 1.75 to
1.80. The paladin check, which sets no rate: lone white swings 2.95, 83 procs on 582 lone units, 9 wearers. Left out: 50
of 76 runs in Cat or Bear Form. One run logged in two reports counts once. Reports from 2026-11-10 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       4 |   363 | 4,065 | 1.85 |
| Hunter  |      20 |   182 | 2,187 | 1.72 |

Reports and fights: 1javmtD7nfNHYJZM 1-3; 2zXhygVYpkrMZGxT 12-15; 3HRN4JXQaDqGFAKp 1-2; 7gWZKLnBmty1cT3J 1;
94fVcDCNBLztXh1k 1-12; 9pJCQk8gz1byGr64 1; FvqNn8XAhHD6M4pY 1-10, 13; HGn9XrYNPvwKtC1m 1-5; HJFN3LKMXkWm7rab 1-18;
NR1k2zLFqKt7GXAD 6; RXDCK9PtZy3Jp2QA 1-3; VLmAcpbQxWf1CtaR 2-17; X47TDyBdAh2nQ6LK 1-10; YHXAcJxWG89pMFgf 1-9;
bhyYq9HXd4xvAwtJ 1-13, 15; ghWq2C84BJj9bQHd 2-4; j3F9NHxqtagPYGkJ 1; jZtdhVaHckGN72L9 1-11; p7ATmV4HqZRdFwLX 19-22;
pGBPWgMrL9KkQZRm 1-2; qcxrZQG46pzKRbfC 1; rDJL7nqX1VZHm9zk 1-8; tk86znZXxMCdwBKb 18; xjZknJqzfC82b3VA 1-9;
xwrkGVNA2Z7Hq4fW 1; yVcJHz7DjAWKdv8T 1; zTbHr3BPvKxRCMZg 9.

## Plaguefang (item 279876, spell 1309315): 2, measured

2,944 procs on 42,419 rolling hits, 55 wearers, 63 runs, in the main hand or the off hand: all hits 1.98 [1.87, 2.11].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.99 [1.82, 2.13],
560 procs on 8,058 lone units, 14 wearers. Dual-wield white: 1.70 [1.55, 1.82], 942 procs on 15,817 lone units, 40
wearers. Specials: 1.98 [1.69, 2.23], 324 procs on 4,674 lone units, 23 wearers. Two-roll abilities, priced as one roll:
0.00, 0 procs on 1 lone unit. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 1,317 isolated hits.
Reports from 2026-11-05 to 2026-11-19.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      29 | 1,348 | 20,892 | 1.84 |
| Warrior |      19 | 1,530 | 20,519 | 2.13 |
| Hunter  |       7 |    66 |  1,008 | 1.87 |

Reports and fights: 149bNrXGaKgQL38P 1, 3; 1Hgm9RqadvcjJh2w 1; 1HpaBdm9fFXDjYPy 1-4; 1XCND7YQJ86Ax32q 1, 3, 5;
1ZJVMgcYWTyRQh2w 1-3; 3tPjXKNACMHcn9b7 1-2; 4ChPnBrVK2fJLMaX 4; 4hNWVC2Kzj8RxpZL 1, 3-4; 6TQkbpNXgGZwYat4 1-3;
6tANqvfKQP1Fkdzg 1-2; 7DKRq39nf4yvFWLG 9-14; 7NbYtg6T2pjr1yL8 1-2; 8mBLvfVzCn4tWyTM 1-3; A4NykGFTwmXKj73Y 1-2;
ARBxdp9znVDMTwJ4 1-4; AZLYpdc6r7P2CMRm 1-5; AgFRCzn6X2KDdtMr 1-11, 13-18; BgCQ6Xzp3Nh2HWkP 1, 4; Bxk8b1hKMz6vt9Vd 1;
CJZhDp9FxwRMaWYy 1-2; CdXV4pYTKPwAqBmN 1-5; CfhkbPBKAZFRGp9g 2-3; FcC4vd2DkRqyMrVT 1-2; GAK2BMH8Wt3k9JFj 1-5;
HPRbL3N7p6wFhrqk 1-3; HqQTwzArBZpPN1y3 1-12; JDH7bFz23Qakdgnh 1, 3; KAvPryL1zCdtWTxV 2, 4; KWZaQjqPLJ7Ry8Gd 1-2;
KYD67nVqZT4jxWyN 1-6; L7QHtkhacypgd3D1 1-2; MzwcDrG3fdFBRy6a 3-5; PqjRDCnyaHt13QJb 1-2, 4-5, 7-8, 10-11;
Qb7TJqYkKtgyz1ah 1; RB47hLrMjdDG8Z6a 1; Rp4QCr31TVmgyqnz 1-2; TH63PWxNv9gRarCY 1-2; TnbVQWRJ7r26CqpK 1, 3;
W8YC9TA6bcXwLRMk 1-6; Z4GKmBf1HXdtkYDL 1-5; a8cL9vKhNmkQfBGd 1-3; aVyXCj1NpnRTFKxt 1-2, 4; afNcyKHdtY2qjPCp 1-4, 6;
bcATfGPv3DX1Jnxy 1-2; d7DFtXL2jfNa8Bw4 1-5; dzTmC6F7xwLPJrq9 5-17; fTpYLw71QVR2qky8 2-13; gKmcRvAdBYqrDnbt 1-3;
jKfkzadV3BLp1cbr 3; kLB1K4faPvhmGzqV 1-5; m2FqGpwzr9cQf7PK 1; n6DtHVNKp1wgymFL 1-2; nWxftVpvMycmw2FL 1-4;
q8DXKyGLJFb1CwVp 10-13; qtcFwCBmkXz7JMH8 21-22; rKXWhHqLza39gkmA 1-2; tm2FQVrcZzfB97v8 1-2; vYQqkXLm43RJZGy7 1-3;
x9vCJk6RYqFLhWGa 2-5; xXtbFrdR4qKAQVB1 7-8; yLKXxQgrPWTnmdZA 1-3; yMNq3krVBAHxG68w 1-3; yTDCtLmVK4kq7gQ8 1-5.

## Silent Fang (item 13953, spell 1293654): 1.45, observed

466 procs on 11,915 rolling hits, 9 wearers, 9 runs, in the main hand: all hits 1.47 [1.30, 1.61]. The rate is all hits,
which the one-hand white rate and the specials rate agree with. Observed: its interval reaches further than 10% from the
read. One-hand white: 1.50 [1.41, 1.81], 154 procs on 3,839 lone units, 6 wearers. Dual-wield white: 0.92 [0.48, 1.19],
16 procs on 651 lone units, 3 wearers. Specials: 1.49 [1.23, 2.11], 64 procs on 1,609 lone units, 9 wearers. Two-roll
abilities, priced as one roll: 0.00, 0 procs on 1 lone unit. Direct: window [-11, 100] ms, median 0 ms, 98th percentile
1 ms, 188 isolated hits. Reports from 2026-11-16 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       7 |   397 | 9,673 | 1.54 |
| Hunter  |       2 |    69 | 2,242 | 1.15 |

Reports and fights: 2qMm9WvRpzYx3GaB 1-2; 4hHBvakRnT7p1GVK 6-7; DmjBrYd7M9VZwCHF 14-22; YZkH8CA3wQxNc1hB 7;
Ya9kKCWRwqjFpgz1 3-4; fMLcHzyZvt7QTCmg 1-9; nyAx3XNDVmYWGdrh 1-6; pNKJR39CwGnmXtbL 1; y6FZRHQnckD8YjAC 5.

## Wraith Scythe (item 11920, spell 16414): 3.65, measured

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

2,009 procs on 14,996 rolling hits, 12 wearers, 13 runs, in the main hand: all hits 3.65 [3.42, 3.85]. The rate is all
hits, which the one-hand white and specials rates agree with. One-hand white: 3.29 [2.90, 3.82], 294 procs on 2,436 lone
units, 7 wearers. Dual-wield white, which sets no rate: 2.92 [2.43, 3.07], 155 procs on 1,450 lone units, 5 wearers.
Specials: 3.96 [3.45, 4.29], 218 procs on 1,501 lone units, 11 wearers; Revenge gives 84 procs on 616 of them. Two-roll
abilities, priced as one roll: 3.96, 18 procs on 124 lone units (Execute 18 on 123 and Mongoose Bite 0 on 1), left out
of the rate. Direct: window [-10, 100] ms, median 0 ms, 98th percentile 1 ms, 453 isolated hits. Dropping the swapped
stretches (69 active minutes, 10% of hits, in 3 runs) moved the read from 3.28 to 3.65. The paladin check, which sets no
rate: lone white swings 4.01, 51 procs on 347 lone units, 5 wearers. Reports from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       9 | 1,599 | 12,067 | 3.61 |
| Rogue   |       1 |   247 |  1,755 | 3.84 |
| Hunter  |       2 |   163 |  1,174 | 3.79 |

Reports and fights: B6JLrnWKF9htyjv2 1; DZjKqdLXkBztYJTC 1; JChQqKVgLWykANRr 9-46; PtKXrwfj8zBHm2Gy 7-8;
PtmYHjkDrqhdR9MA 1-6, 9-10; RBpWJ7zM9N3TDGQc 1; Ta4fLJj32kWzmXgw 1-29; XPgHryZdABkRDhwV 3; YnVFyQq3w28mr41R 1;
ZQtv6AqYmdJFWwhy 7-8; h4gWnT3HLBbYzMcK 1; p1VPMyG8jgR6Ymc7 1-4; zxtbh1BmNVC8W2FH 10-14, 16-21.

## Wyvern Tailspike (item 5752, spell 16400): 2.1, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

738 procs on 11,824 rolling hits, 29 wearers, 31 runs, in the main hand or the off hand: all hits 2.08 [1.90, 2.24].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: 2.41 [1.68, 3.35], 35 procs on 484 lone units, 4 wearers (a
Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.92 [1.72, 2.19], 277 procs on 4,809 lone
units, 21 wearers. Specials: 2.04 [1.33, 2.57], 41 procs on 671 lone units, 14 wearers. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 1 ms, 281 isolated hits. Dropping the swapped stretches (9.1 active minutes, 1% of hits, in
1 run) moved the read from 2.05 to 2.08. Reports from 2026-11-10 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |      13 |   622 | 9,874 | 2.10 |
| Hunter |      14 |    69 | 1,232 | 1.87 |
| Shaman |       2 |    47 |   718 | 2.18 |

Reports and fights: 2C6paPN987vMZWwd 1, 3; 2RdqgkHFNVMnKCh8 1-2; 4NWagKbPRwT81yh7 3-8; 7YcB9haxNGD463Cj 4;
8129XG6zfMVJrvmg 1-3; 8X94GnMwCpFcZ2Q6 1; 9h4H3bVCgyfPQWt7 1; 9nKGxYHVDhvApL83 1-2; Amh4kVqaFGB3xPZ9 1-3;
D4jMGTC7Q1cVtF6p 1; DMYyPqt27cGjL1K8 1; JD2ag1wYV8qCxhtQ 1; JPymD8BKctgqMFLH 1; L7rDtmvMqbH9xnVF 1-24; MLhC3KjWrngq2vtZ
1-2; PzbCdMALrYJh7vNV 1-2; QZyPVgTvLJnamk84 1-3; Wh7fgRNMmLP4YBnX 1; XYDcxmK8A3at2wgN 1; b6QHjXT2GF7A1YPr 1-8;
cT71Vr9tPDKkxWv3 1-2; dGhawD4rxCAH7bR2 1-4, 6-10, 12, 14-21, 23-30, 32-37, 39, 41-51; fyb2xc14zNRdvDMY 1-2;
hxZPKTfQYzdVjvN3 6-9, 11; qVY1prKMP6GjFdXm 1; qxACn7fMgDQL3Vw6 1-9; rqhGLfWyaAtZvjTC 3, 6; vnDM8K4RYx6kW2GH 1-2;
wtJqpy9ThGxFz4nZ 1-2; yF1hGHxMnrPJK83T 1; ynGtWRk4cZX3TmYv 1.

## Count of the 2026-11-06 to 2026-11-21 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; those sections are in EV-0126, EV-0132 and EV-0136
- Game build: 1.60.1.70790 to 1.60.1.71037 by report date (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11
  UTC in the US and 2026-11-11 02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC,
  71037 from 2026-11-20 18:23 UTC). The sim's database is 1.60.1.71037, push 114326; the hotfix each player ran is not
  recorded
- Date: 2026-11-06 to 2026-11-21
- Pull request: #1156

The count is §Method on the hits EV-0155 finds rolling a weapon proc, run again over every wearer stream on disk for
these weapons, with more wearers fetched on 2026-11-21. The same rolling hits, classes and forms are left out: paladins
from every rate, druids in Cat or Bear Form, and the two-roll abilities with the procs around them. The intervals,
lone-unit checks, swap drop and credit windows are this file's.

A rate is measured where a lone-unit rate of 50 or more procs (lone one-hand white swings or specials) overlaps the
all-hits interval, and 5 or more wearers of more than one class give an all-hits interval within 10% of the rate either
side. The rate is the all-hits rate rounded to the nearest 0.05. A lone-unit interval from under 5 wearers is a Poisson
interval.

Each section's read by report date is the all-hits rate over the reports that start before 2026-11-18, from then to push
114313, and after it.

Scripts and inputs, in `forever-notices-1121`: `results.json`, the counts of `forever-sweep-auto/analyze.py` (version
12, with the paladin check that sets no rate) run with `--items items.txt --census census_used.json`;
`report.py`, the call per weapon (`rep_final.json`); and `dates.py`, the report dates (`dates_final.json`).

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 3,714 reports in all for the count, 78 of them for these
  weapons
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 22 of them for these weapons.
  The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on these weapons.

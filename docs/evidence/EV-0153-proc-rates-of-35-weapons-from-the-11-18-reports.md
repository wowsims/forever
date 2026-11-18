# EV-0153: Proc rates of 35 weapons from the Warcraft Logs reports up to 2026-11-18

- Claim: each weapon below procs at the rate its section gives, priced on each landed hit's weapon base speed. A rate is
  measured when 5 or more wearers of more than one class give it and its 95% interval lies within 10% of it either side,
  and observed otherwise: 26 are measured and 9 observed. A melee ability that deals no direct damage (Sunder Armor,
  Rend) rolls weapon procs when it lands, as a white swing does, so its landing counts as a hit.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists; 904
  reports in all for the rates, and the 60 reports the section on melee abilities lists
- Game build: 1.60.1.70790 to 1.60.1.70959 by report date; the hotfix each player ran is not recorded
- Date: 2026-11-05 to 2026-11-18
- Pull request: #1108

This replaces the rates EV-0136 gives for 26 of these weapons and EV-0146's Killmaim rate.

## Method

EV-0136's method on the reports listed up to 2026-11-18. A proc is an event of the item's proc spell, counted once when
it logs several events within 50 ms. The hits are the wearer's landed, non-periodic hits the trigger's proc flags hear:
white swings of the item's hand and melee-defense hits on the main hand for a melee proc, Auto Shot and ranged-defense
hits on the ranged weapon for a ranged one. Periodic ticks, damage shields and the abilities EV-0136 names stay out. A
dual wielder's white swings are split between the hands as EV-0136 describes. A druid run in Cat or Bear Form is left
out, a run logged in two reports counts once, and a stretch with no proc where several were due is dropped as the item
swapped away. The rate is the procs over the sum of each hit's weapon base speed over 60, and its interval is 2,000
bootstrap draws of whole wearers. Each sim rate is the read rounded to 0.05. Hits split between the hands are
fractional, so the class rows are rounded.

Two things differ from EV-0136. The interval is the wearer bootstrap alone, where EV-0136 took the wider of it and the
per-hit Wilson interval. And a melee ability that deals no direct damage counts as a hit when it lands. Its landing logs
as its debuff applied, stacked or refreshed, with no damage event: Sunder Armor, and Rend's application. Each ability,
target and timestamp counts once.

Scripts: `analyze.py` for the rates, `targets.py` for the rounding and the measured or observed call.

## Melee abilities that deal no direct damage

A landed Sunder Armor is followed by a proc within 50 ms about as often as a landed white swing, so it rolls the weapon's
proc as a swing does. Warriors who wore each weapon, every fight of theirs in the reports listed below
(`nodamage_rolls.py`):

| Weapon          | After a landed white swing | After a landed Sunder Armor |
| --------------- | -------------------------- | --------------------------- |
| Toofbreaker     | 5.2% (1,770 / 34,108)      | 5.9% (219 / 3,686)          |
| Cursed Felblade | 7.1% (801 / 11,278)        | 7.4% (178 / 2,395)          |
| Argent Avenger  | 0.8% (241 / 30,543)        | 1.5% (60 / 3,980)           |

Without these landings warriors read higher than the other classes on the same weapon, since the procs that followed a
Sunder Armor had no hit counted for them. Cursed Felblade read 2.49 for warriors against 2.12 for rogues and 2.02 for
paladins, and reads 2.01, 2.06 and 2.02 with them counted. Plaguefang read 2.45 for warriors against about 1.85 for the
other classes, and 2.07 with them counted. `uncredited.py` lists the procs no counted hit explains, and
`trigger_hits.py` the hit each proc follows. The sim already gives Sunder Armor and Rend `ProcMaskMeleeMHSpecial`
(`sim/warrior/sunder_armor.go`, `sim/warrior/rend.go`), so only the measurement changes.

Toofbreaker (286430), 25 reports: 1BNXynxVbH4kGt8R 2PZCz9c1HgtrX7Th 2xY1D3mCwagdy7Wc 2yqzTJfLN8cvwaDK 3JbjLGXVzctAmfaD
3cWzXTLmvDh2bC7H 4YVTAqgG3dtmD8C9 4wWtcgbPhNdMJ2LQ 6aN7CQwbfTW182vp 72JjqrWXMdtwKx68 84nxvZk9BmHf7Ldr 9vWqJjhLGCpdQZKc
Cp3ND2J1xWhbFzP7 FLK4Vqw3cp6XkRAT HkNB7A6tanYcGLWj JX4afWrdcw8RkHDt P6qAmjBhrNa8b9MR PXbhqkFHpt47RmDy YP3LzZWFTMDGXdyR
ZqBnHh1CtfDcpwK7 gTPNZcJ9mG4dRC2X tLGanMDfw82FJymz tac7hXVRgkbZNMDy vtyhfFz9wPGJdHqM ztpB7m1kbLQ3KdqR.

Cursed Felblade (14145), 18 reports: 12VhLkPa98DYFH4g 1XFBL6hQPHcY4Zwr 1cf7ba34WkwGyzQY 3CyQk69LzJjTn2X4
7qLWf2j9a8tvCNwH 9aCJ7GH6QRFjgqb4 Gv74gVkXcyPHaYzj T7HKGnYfbMLwWCJp TdXmWDR3QFJwpYfq W416FHmftD9rPRgA dwq8K9CRNBAQtLjp
fxbtCJFLjv6WMBcy nfCzby7RrmYFxTXc txhaZrbjXR7Kf2TD zPJ2w1BrKA3Dv968 zRd3nAXrVQqg9ZH6 zhj69Dt2X1b3fvMT zk682jCMJ4AdKfV9.

Argent Avenger (13246), 17 reports: 3HaPg2Yktxrnb7NL 4VwxHWJ3pgNmvQY2 7fy93LdcwQhvF2gP 8FBqfvLyWTdKxbZh CgXG9bvFH6jJVZKk
FN8krKfCb3BXWmwV K6VvbDypXFLQJW7d NK6cCM4pJyRhaAFW Qr9PmCnptRWMjN7k RHgxPaQbVWjDTc4K Wn8wHhaxy31MLzjp Zvr84V1Y3Gpg6n9H
aANdrQRjKP7MJnCY dzgfcTtLA1v7jDZ9 kjbQpMhK81LqRgyB qRL8HagchkrTVJmD xkKRJPqCLzBa9GdH.

## Rates

| Weapon                     |   Item | Rate | Read [95%]        |   Procs / hits | Wearers | Call     |
| -------------------------- | -----: | ---: | ----------------- | -------------: | ------: | -------- |
| Angerforge's Battle Axe    |  11816 | 2.05 | 2.06 [1.97, 2.19] |   933 / 10,433 |      13 | measured |
| Archeus                    |   2000 |    1 | 0.99 [0.88, 1.09] |    390 / 7,188 |      29 | observed |
| Barbaric Crossbow          | 272999 | 3.25 | 3.26 [3.14, 3.39] | 2,622 / 24,147 |      36 | measured |
| Blade of the Basilisk      |   8223 | 1.05 | 1.06 [0.98, 1.16] |   335 / 10,016 |       7 | measured |
| Bleeding Crescent          |   6738 |    1 | 0.99 [0.89, 1.12] |   487 / 12,274 |      27 | observed |
| Bloodrazor                 |    809 |  0.9 | 0.91 [0.81, 1.01] | 1,250 / 30,627 |      15 | observed |
| Bow of Searing Arrows      |   2825 |  5.1 | 5.10 [4.89, 5.35] | 2,450 / 10,674 |      15 | measured |
| Burning War Axe            |   2299 | 0.95 | 0.96 [0.90, 1.03] |    504 / 9,494 |      15 | measured |
| Cursed Felblade            |  14145 |    2 | 2.02 [1.94, 2.11] | 2,469 / 31,808 |      32 | measured |
| Darkwater Talwar           |  11121 | 3.05 | 3.05 [2.91, 3.19] | 3,394 / 30,384 |      33 | measured |
| Dazzling Longsword         |    869 |  2.6 | 2.62 [2.45, 2.78] | 4,990 / 43,887 |      40 | measured |
| Diabolic Skiver            |   9475 |  2.1 | 2.12 [1.83, 2.49] |    675 / 6,604 |      13 | observed |
| Electrocutioner Leg        |   9446 | 3.95 | 3.93 [3.86, 4.00] | 8,714 / 78,287 |      53 | measured |
| Fang of the Crystal Spider |  13218 |    1 | 1.00 [0.92, 1.10] |   306 / 11,511 |      10 | observed |
| Felstriker                 |  12590 |  0.6 | 0.62 [0.54, 0.68] | 1,047 / 59,127 |      17 | observed |
| Fist of the Damned         |  10804 |  3.1 | 3.12 [2.97, 3.36] | 3,905 / 39,540 |      23 | measured |
| Glutton's Cleaver          |  10772 |  2.2 | 2.22 [2.09, 2.40] |   947 / 12,820 |      21 | measured |
| Grimclaw                   |   1481 |  1.9 | 1.90 [1.64, 2.05] |    292 / 4,618 |       6 | observed |
| Hanzo Sword                |   8190 |    1 | 1.01 [0.96, 1.08] |   695 / 27,532 |      16 | measured |
| Iron Knuckles              |   2942 | 2.05 | 2.06 [1.92, 2.16] |    349 / 5,988 |       6 | measured |
| Ironforge Greathammer      | 279895 | 2.85 | 2.87 [2.77, 2.98] | 3,561 / 24,016 |      45 | measured |
| Killmaim                   |  13016 | 0.95 | 0.93 [0.86, 0.99] |   810 / 16,414 |      34 | measured |
| Loksey's Training Stick    |   7710 | 2.15 | 2.14 [2.02, 2.26] |    902 / 8,161 |      35 | measured |
| Lord General's Sword       |  11817 | 1.95 | 1.96 [1.82, 2.05] | 6,172 / 72,714 |      20 | measured |
| Nightblade                 |   1982 |  2.1 | 2.09 [2.01, 2.18] | 3,535 / 29,012 |      22 | observed |
| Phantom Blade              |   7961 |    2 | 2.01 [1.85, 2.19] |   728 / 13,570 |      12 | measured |
| Plaguefang                 | 279876 | 1.95 | 1.93 [1.84, 2.02] | 3,901 / 57,860 |      64 | measured |
| Sang'thraze the Deflector  |   9379 |  2.1 | 2.09 [1.98, 2.22] | 2,787 / 47,139 |      27 | measured |
| Satyr's Lash               |  17752 | 1.95 | 1.94 [1.77, 2.09] |   918 / 16,691 |      22 | measured |
| Silent Fang                |  13953 | 1.55 | 1.55 [1.41, 1.70] |   816 / 19,726 |      11 | measured |
| Strike of the Hydra        |   6909 | 2.15 | 2.13 [2.04, 2.22] | 5,740 / 49,004 |      71 | measured |
| Tainted Pierce             |   8225 |  1.9 | 1.92 [1.69, 2.11] |   631 / 10,356 |       9 | observed |
| Taran Icebreaker           |   2915 | 0.95 | 0.94 [0.89, 1.00] |   710 / 11,908 |      11 | measured |
| Thornspike                 |   6681 | 3.45 | 3.43 [3.19, 3.69] | 3,654 / 37,647 |      45 | measured |
| Wyvern Tailspike           |   5752 |    2 | 2.01 [1.82, 2.19] |   607 / 10,078 |      14 | measured |

## Angerforge's Battle Axe (item 11816, spell 1300783): 2.05, measured

933 procs on 10,433 landed hits, 13 wearers, 13 runs, in the main hand: 2.06 [1.97, 2.19]. The sim had no rate. Reports
from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      10 |   697 | 8,001 | 2.01 |
| Paladin |       3 |   236 | 2,432 | 2.24 |

Reports and fights: 1GBhzTR8CZwcJfFV 20-32, 35, 37-38, 40-51, 53-63, 65-66, 68-82, 84-88, 90-99; 2WkFQGcDVYwpTMJ9 7-18;
4rfgMCY9qjyvQhHD 1; AfJ6DYbh9dmkgHZL 1-2; GWL9jZvR4rpaKftb 1-4, 6-8, 10, 12-13; W7VPQB18nHfkMaAc 11-13, 15-31;
cgmhtvNdTGCn7f6L 73-84; dnDg7j9kRw6WtbML 55-61, 63-69; hZLykw9dRnYbtvKP 1-2; jKQDqkwmx46ZLBFg 17-24;
jr7TAd93qLPxWzhp 1-2; zmDY4LVX7N8vh9aR 1-3; znBQHJTxXapkRLt3 1-11.

## Archeus (item 2000, spell 18091): 1, observed

390 procs on 7,188 landed hits, 29 wearers, 33 runs, in the main hand: 0.99 [0.88, 1.09]. Observed: its interval reaches
further than 10% from the rate. Its override row had 1, a vanilla carry-over. Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   175 | 3,345 | 0.95 |
| Paladin |       3 |   166 | 2,913 | 1.04 |
| Hunter  |      21 |    49 |   930 | 0.96 |

Reports and fights: 1NRQ9tBnTrq24D3k 1-11; 2qk4MYw7cr8KGFtm 1-8; 6z9hqHARQr3vwLXd 1-2; 76ZV8FQ1RHDhbq3z 1-4, 6-9;
FRxmtWLHr3wvB8Q1 1-8; FfAC3tWBxbp2HV9j 1-3; Hgp61KVdvPqYyr2N 1-11; KBZyYLg1WTqm6Chw 1-14; Lh96QfHTn7AaMJ3j 1-15;
MmqC2hGvFY7x6Dgd 1-3; NFD1RP4LdQrhfw3W 12-17; TDjrhyCkw3P9YApc 1-9; TkmGgNFMZvL9pHcP 1-5; TprqfR3jCg8HvLmP 1;
W6nhHydV1wpMmZ7a 4-5; WdXL8PgGtTZq3MwB 3-12, 14-39; XGrAwyaKghVdPkB7 1-2; XvWP1x3FkAzwLHZC 1-9; aQGPVW4TBcv8y3wD 1-2, 4,
6; b186j7df4hKxNPkV 1-4; bPqj8zkJN3YatpwD 1-2; cvfBaDxQX7K2zFtk 1-7; dg4ZD7xMfRBptJq9 1-4; gr9zjCAX3mY41aqf 8-12;
mtBc1WZf3ydHVPFv 1-6; qpXkvRdWxKzfNr2G 1; rNXvGcnZabhLJd9A 1-2; rQdnmBHXCpKgJqw4 1; vN8WfZatBGgdCDHJ 1-4, 6-7;
vRPQ237fWm1tFYDV 1-4; vTX3cKpNbY6jJVDa 1; wMRNLvP8CBrgd4x7 1-4, 6-13, 15-17, 19-26, 28-39; z1APqCWaMYvxbkd7 1-3.

## Barbaric Crossbow (item 272999, spell 1291551): 3.25, measured

2,622 procs on 24,147 landed hits, 36 wearers, 43 runs, on the ranged weapon: 3.26 [3.14, 3.39]. EV-0136 gave 3.45.
Reports from 2026-11-07 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |   PPM |
| ------- | ------: | ----: | -----: | ----: |
| Hunter  |      31 | 2,617 | 24,085 |  3.26 |
| Warrior |       4 |     4 |     59 |  2.03 |
| Rogue   |       1 |     1 |      3 | 10.00 |

Reports and fights: 1AfDWYQT7Xjd24K3 13-14, 16; 1Gg92cW6XBhkwYb4 1-5; 1RrDAhp7ZvVYc9P3 1-6; 1W9N26dHfnPV4G7A 1, 3-9;
1y6GV7KLTZPrJH2w 1-3; 2GbXqD7KcWTfRnYa 1-2; 2MHXV8Y9jx43vhpk 3-12, 14-18; 2tJkzh16X38wrQfP 1-3; 4ChPnBrVK2fJLMaX 4;
72XWYanybGdwPg41 1; 7VkrJpR8XyFtL6Hm 1-2; 8Xxb9LDGHP7vcJfh 1-5; 92JNtax7RTkQ6cHv 1; CVFcbDGj6n8WMKyL 1-2;
Ghp9gxWFkBNacLfn 1-5; GkLDK8WpcdvPatfB 1-2; H9QxwyGt4z8MmqCj 1; J4mG8yjgVCALMfRw 1-2; KBzML9kdHFCXVpQc 1-2;
KYD67nVqZT4jxWyN 5-6; L7QHtkhacypgd3D1 1-2; LnPRx96hKZN3CgTF 1; NyG7zfKj2Mvk1FLJ 1-6; P7MBDvyCRQtLkJHz 1, 3-8;
T6MXt73janNJgmyY 1-3; Td29Nzahj1cmGMWZ 1-2; TdXmWDR3QFJwpYfq 1; YHh2t9agcvQjTAZq 2-3; Z9CQ6fqYjzcB4tG2 1-3;
dc3Y84hWFnrA2wqp 1-3; dtLy89QkrcHx3pMf 1-2; fxbtCJFLjv6WMBcy 1-2; hMdFnHqgVRDTyKPY 1; hRzbpMrDY3cgQX7n 1-7;
kBHbV3DZjamvfJY4 1-3; nALy2WrwHC9qdMFm 1-2; p9HgV7zdCKqbrwJ3 1-3; pw8aYCczKgjMbvAd 1-4; r2BYfmZW1wkLDdNx 1-2;
tbWaYgLTw8XQAzdC 1-3; vqpcV7yDGCkbftmP 1; wmT1zqYBat3fHDLy 1-5, 8-9; zcL8MvJnPXxYhNyC 1.

## Blade of the Basilisk (item 8223, spell 10351): 1.05, measured

335 procs on 10,016 landed hits, 7 wearers, 8 runs, in the main hand: 1.06 [0.98, 1.16]. The sim had no rate. Reports
from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Paladin |       2 |   178 | 5,452 | 1.03 |
| Warrior |       5 |   157 | 4,564 | 1.09 |

Reports and fights: 7fgdkyv1njDKa8Gc 6; 8NMWGB1n6zdLtPrY 1-5; 96pBnGdAKJMaCb3N 1-12; ah8jAXTQFywKCf4x 1-7;
hYHBQ9ZJX3vFAnkj 7-14; qa1tKFnvZ98hwj6x 1-6; rfjTzc9QVZD3hxWq 1-3; xH8TDPQKb6mYragf 1.

## Bleeding Crescent (item 6738, spell 16403): 1, observed

487 procs on 12,274 landed hits, 27 wearers, 28 runs, in the main hand: 0.99 [0.89, 1.12]. Observed: its interval
reaches further than 10% from the rate. No rate was recorded. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       9 |   260 | 7,151 | 0.91 |
| Rogue   |       3 |   120 | 2,229 | 1.35 |
| Paladin |       3 |    48 | 1,230 | 0.98 |
| Hunter  |      10 |    34 |   944 | 0.90 |
| Shaman  |       2 |    25 |   721 | 0.87 |

Reports and fights: 1WC3Bnw4A7RKpLDj 1-4; 1rdmQ9wptPWHfNG6 1-4, 6, 8-9, 11-23; 9CGYv1Jp24aWyhf8 1-2; AwkvX3Tt1jqLhMVc 1,
3; B3vtVgmqMTHDAxNY 1-10; CLbmK261T3d9gAR4 1-6; CNWJgkfyAhjB3vRt 1-11; ChQpLdxH2Mfj3cVk 3-5; FbNk9QwxY7nG618P 1-19;
J9jRmkxvtXcAKfM3 1-4; NwKVWqHfgk1jLvAG 1-5; P8Znjv1fgBD2CAVY 6-12; QDVfcwHpjZRzL94q 1-6, 8-13; RPnVYQcG9xTJh1ar 1-4;
VF9aXbfLT68W4CdP 1-17; WnvfNJK8dQG9xPMY 1-2; XgZdn3D9KR8cfLNY 1-10; ZAWJk6BdKhr9nC8T 1-6; f2zK39XdhnNqZJgR 1-2, 5;
fjb2rRngTPa48WA9 1-4; h92ZHrfm3RFQtYvT 1-8; jC4HdxZGVNLz1Yvm 1-2, 4-6; jGrv6B1hNMDJPc3W 7-10; kmBptFTG68Dcq3jx 9-10;
mQ4MqkD1nbaBXLdy 2, 4-6; npbfVjrQdDLhtRWY 20-22; pKn8wBZcQ62PyJH7 12-24; tD7Rm3GBjdqPzhwC 16-17.

## Bloodrazor (item 809, spell 17504): 0.9, observed

1,250 procs on 30,627 landed hits, 15 wearers, 19 runs, in the main hand: 0.91 [0.81, 1.01]. Observed: its interval
reaches further than 10% from the rate. EV-0136 gave 1.05. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 |   777 | 19,621 | 0.88 |
| Rogue   |       7 |   473 | 11,006 | 0.96 |

Reports and fights: 3XCdM1LQRNmvY8kp 1-8, 10-12; 76zv93gbJQFrWKfd 1-30; 8GFRaPBnKWf7Qrpb 1-3, 5-6; AZchwX93DmyHLbv6 1-3;
NWvgFp3kfR2bHyYa 1-35; NjgtYv4VMR26q7h9 1-4; RW1BNk6Ybyh3XjLQ 1-21; ZLpcV6krn1wYHJvt 1-8; aNh3Dt8kbZHL9cBM 1;
j1xyYV7gcMH69ND8 1-20; jJWt7Dnm4qc9rzQd 1; jr7TAd93qLPxWzhp 1-2; kJ3WdhMLqv7K2Nyf 1-3; nzxZ4h87yGXJPgfp 1;
p34dnvkg7yDKL6cr 1-3; qVnZ6bPMhWRJA397 1-3; tk3BD1bZafMgKxWX 1-4; vBWX7HQn9zpTMGd8 13-22; vyQrDZN7RX82ApnP 1-3.

## Bow of Searing Arrows (item 2825, spell 29638): 5.1, measured

2,450 procs on 10,674 landed hits, 15 wearers, 15 runs, on the ranged weapon: 5.10 [4.89, 5.35]. Left out: 1 run logged
in a second report. EV-0136 gave 5.4. Reports from 2026-11-13 to 2026-11-18.

| Class  | Wearers | Procs |   Hits |   PPM |
| ------ | ------: | ----: | -----: | ----: |
| Hunter |      14 | 2,449 | 10,673 |  5.10 |
| Rogue  |       1 |     1 |      1 | 22.22 |

Reports and fights: 4CBckLvG7RaxFzYA 1-4; 6Krp8bkfx4cRjBaz 1, 3-15; 7BjkRJr2DZp1g8TL 1-4; AhLc6aGKMCdvBgDW 1-5;
ApJhHfRkMDqFj3rb 4-5; H2GvZtVYNy4QzqLw 12-21; KGQHdbFCZpt1YX23 18, 20-23; Pf4Y2xdhVz1vq9rn 1-12; YWGNyaDcLwT1QR6J 1-8;
cJhVzAdLx3BHqpkT 1-4; fHbJKd4LnVXDW2gj 1-2; krpvb6WaLmFnV1KG 1-2; mtcXBg7ZkVxFHQnh 1-4; qHgyawJNVb7126xL 1;
tk86znZXxMCdwBKb 2-18.

## Burning War Axe (item 2299, spell 18199): 0.95, measured

504 procs on 9,494 landed hits, 15 wearers, 15 runs, in the main hand: 0.96 [0.90, 1.03]. EV-0136 gave 1.05. Reports
from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      10 |   355 | 6,655 | 0.97 |
| Warrior |       4 |   111 | 2,081 | 0.97 |
| Paladin |       1 |    38 |   758 | 0.91 |

Reports and fights: 3cKTmVyb1ZPwFXkf 1, 3-13; 7XwRKY4qkjDt3ryz 2-3; FMcvKqTzn7m8Va3P 1-15; Fgy2bt81NL9CAzaR 1-13, 15;
KRCBMJvg72fHVwWL 1-17; LwTzGJXrN3HWqCfD 1-5; Pc1RkCWp8wTNfFr2 3-4; QvKPgyHDkBMbcp7T 1-10, 12-21; RBmtAMLxz78Pj2Wh 7-10;
VQFaRpT6J2xLDbZg 1, 3-5; Wf8ghC6TtQKrkXdv 8-10, 12-15; bkMdQTwZmqNPvhrR 1-5; j68nZp1gcrGChaRW 1-12;
qMD3CVapwb6nAgx2 16-40, 42-43; t6HbjRZmkyx3pQwK 1-14.

## Cursed Felblade (item 14145, spell 18381): 2, measured

2,469 procs on 31,808 landed hits, 32 wearers, 34 runs, in the main hand: 2.02 [1.94, 2.11]. EV-0136 gave 2.55. Reports
from 2026-11-05 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      16 | 1,323 | 17,161 | 2.01 |
| Rogue   |       9 |   708 |  8,985 | 2.06 |
| Paladin |       6 |   436 |  5,642 | 2.02 |
| Hunter  |       1 |     2 |     20 | 2.61 |

Reports and fights: 12VhLkPa98DYFH4g 1-3; 1RrDAhp7ZvVYc9P3 1-6; 1XFBL6hQPHcY4Zwr 1-2; 1cf7ba34WkwGyzQY 1-12, 14-16;
3CyQk69LzJjTn2X4 1-5; 3rDXa6gPxYNZJjFH 2; 4ChPnBrVK2fJLMaX 4; 7qLWf2j9a8tvCNwH 1-7; 8Xxb9LDGHP7vcJfh 1-5;
8aymDGwFnWZAx7YR 1-3; 9aCJ7GH6QRFjgqb4 1-11; CfhkbPBKAZFRGp9g 2-3; Gf4361AWjMR9Z2qY 1-13; Gv74gVkXcyPHaYzj 1-12;
KAvPryL1zCdtWTxV 2, 4; PKpXc2n7WrBjxfCA 1-2; T7HKGnYfbMLwWCJp 1-7; TdXmWDR3QFJwpYfq 1; W416FHmftD9rPRgA 1-2;
ZmCacYhPgANTkHjr 4; dwq8K9CRNBAQtLjp 1; dzTmC6F7xwLPJrq9 1-17; fxbtCJFLjv6WMBcy 1-2; hcfj4XqxKP8J6nCv 1-2;
jKfkzadV3BLp1cbr 3; nPZbN3Ymk1LVtDQK 1-2; nfCzby7RrmYFxTXc 1-3; txhaZrbjXR7Kf2TD 1, 3; zLK3mTNayZMgR6vh 1-7;
zPJ2w1BrKA3Dv968 1-2; zRd3nAXrVQqg9ZH6 1, 3-5; zhj69Dt2X1b3fvMT 1; zk682jCMJ4AdKfV9 1-2.

## Darkwater Talwar (item 11121, spell 16408): 3.05, measured

3,394 procs on 30,384 landed hits, 33 wearers, 36 runs, in the main hand: 3.05 [2.91, 3.19]. Left out: 1 run with the
item swapped away throughout. EV-0136 gave 3.3. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       9 | 1,525 | 12,778 | 3.25 |
| Rogue   |      12 |   918 |  8,688 | 2.88 |
| Paladin |       8 |   917 |  8,615 | 2.90 |
| Hunter  |       4 |    34 |    304 | 3.05 |

Reports and fights: 14ZTxcCRXhQLfYwP 1-2, 5-6; 1CYTfP3pFDcaRrJm 1-8; 1h9vznDdxyjC7KWZ 1; 1vJgfm9dGnkDzwXb 1-2;
1w4hbgxXHpf3N6M9 2-3; 23dT9fk1DqcaJmwn 1-7; 24KTy3bLGNQm1rdx 1-20; 2GYNtHzfKmWpTPVZ 1-14; 2Z3YxGdfPhMK97kQ 1-4;
4278z9CPcrFZKgHk 1, 3; 6xhBGvgcM9H48tyK 1-8; DK72jqp46VtnYHfZ 1; FR9jJ7b31XCkLmtQ 2, 4-7; Fw6bq2BvXnJmaRk9 1-3;
Gb86hWq9DjJazgTN 1-10; H3vdcCr6JX1gwYtq 1-3; LHRhtADNW8GfyYqT 1; LzFXxt3Y1HyVK4nc 1-7; NDf1mPJRCaTM8WkK 2;
QgR6kbHvY1CBhXFj 1-2; R2zHVBktcw8j6rhM 1-3, 5-7; Tz9nB8pHLD12XqVQ 1; XT4hNV3HgGDqdKPj 1; Z1KYP3yWDVN7L6XH 1-14;
g3nkvH9atfKRdVAh 1-11; khV3XMmJ4aRKzxBt 1-3; mRxb9L2wNkMJTWXY 1-7; mj379AhvbKZVpWHC 1-7; pwNAZn643XgzcbYa 1-3, 5-6,
8-10; rTzVg1BxmZh6pFKC 1-11; vk2wGPDbQp8TZBR4 1; wQAj94dkrc3MvXpD 1-6; yzaKX18ZAmpHvkjP 1-16; zhZcTYBNvwr48gP6 1-6.

## Dazzling Longsword (item 869, spell 13752): 2.6, measured

4,990 procs on 43,887 landed hits, 40 wearers, 40 runs, in the main hand: 2.62 [2.45, 2.78]. EV-0136 gave 2.8. Reports
from 2026-11-09 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      23 | 2,780 | 23,155 | 2.77 |
| Warrior |      11 | 1,092 | 11,383 | 2.21 |
| Paladin |       5 |   979 |  8,099 | 2.79 |
| Hunter  |       1 |   139 |  1,249 | 2.57 |

Reports and fights: 1FV3WyfXGCcPrT9b 1, 3-8; 4C28YPXqgTMQJxB7 5-9; 4qZXFGfc1wCMQKL7 1-8; 8F4MhkmNBxRrf6Zw 55-68, 71-83,
85-89, 92, 94-105, 108-123, 125-141, 143-151; 8X2Kc46faTnxVBLm 1-3; 9cajk6Jqd3NTQngf 6-8; A1FKkYpPxrX8tbNv 1-2, 4-5;
BCZgcRxrh7dA4ayX 1-4; BgKt3Qm2fJDAWLzh 1-23; CNvKjpmwaxfGcFzL 1-7; CrR9c6VtXG4bzJhx 1-17; D4A3QNCrX19W7cqJ 1-10;
FR9jJ7b31XCkLmtQ 5-7; GzZH1WhTLQ3JaDXt 1-11; JTztbhZyaWHC6kxf 1-5; KVdX1Wn4GpzAH9MN 11-12; Lg82ZFDyQ4MvRm9p 1-2, 5-11,
13-26; M4CvXLZz3qm6pDRc 1-3, 5; PKRTfgw2Q7ZBxt4r 1-26; Q7BxMTcG1ZbgX2v8 1-2; QDadybK3XVPRvwLh 25-30; QftYMR2haLcXZ7Nz 1;
T71XnfaF3cY9M2Rg 1-13; Thf9agqHJyRmtdpD 1-5; XCJdxAF17zqHMhLY 1-2; ZtndFBMraKvcLfw9 7-9; b12KD98fndF4hMq3 10-18;
dmKgVGCJMQpq3t17 1-14; fCrvDxhZXqJAg62m 4-10; fkVmJwWZjMy3LBX4 1-3; gTPNZcJ9mG4dRC2X 1-4; nF7RcakdVTQD2tvH 1-3;
nLk2TKDJWP6vBgam 1; pKj3bDrf1h2qtNMF 9-15; qW3KvYDpw6ARxZbc 1-7; tkQKL9bGT6JVR3na 14-16; vMdAbxpnTcFYfXZ9 1-22;
xM9bRGyajmBFvCqt 1; xqy2GfPj67XCwrT3 3-4; z639ZrPafCA1m7WT 1-3.

## Diabolic Skiver (item 9475, spell 18206): 2.1, observed

675 procs on 6,604 landed hits, 13 wearers, 14 runs, in the main hand: 2.12 [1.83, 2.49]. Observed: its interval reaches
further than 10% from the rate. Left out: 25 runs in Cat or Bear Form. The sim had no rate. Reports from 2026-11-10 to
2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Paladin |       4 |   490 | 4,525 | 2.24 |
| Hunter  |       8 |   132 | 1,413 | 1.93 |
| Warrior |       1 |    53 |   666 | 1.65 |

Reports and fights: 2zXhygVYpkrMZGxT 12-15; 7gWZKLnBmty1cT3J 1; 94fVcDCNBLztXh1k 1-12; DQhPGnHRqmTZ2M74 1-2;
NR1k2zLFqKt7GXAD 6; RXDCK9PtZy3Jp2QA 1-3; YHXAcJxWG89pMFgf 1-9; Z3Kg7BqpzGMjybTD 1; jZtdhVaHckGN72L9 1-11;
k2KxgByR4dTpLQtH 1-4; rBWb9Nyt4YpPFGH7 1-10, 12-23; rDJL7nqX1VZHm9zk 1-8; tPXDqNr13FY2abTc 1-3; tk86znZXxMCdwBKb 18.

## Electrocutioner Leg (item 9446, spell 1292851): 3.95, measured

8,714 procs on 78,287 landed hits, 53 wearers, 55 runs, in the main hand: 3.93 [3.86, 4.00]. EV-0136 gave 4.25. Reports
from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      23 | 3,720 | 33,545 | 3.91 |
| Paladin |      13 | 2,748 | 24,419 | 3.97 |
| Rogue   |      14 | 2,152 | 19,320 | 3.93 |
| Hunter  |       3 |    94 |  1,003 | 3.31 |

Reports and fights: 1qpQ6FW8fxD9gC4G 1-7, 9-11, 13-26; 27dNt8HLBCPpAmYZ 1-31; 2jYyAK1RrPvTC8WM 1-5;
2ywhMnJq4Y3zaWtg 1-9; 3B4wmh7JGkQq2Ndj 1-2, 4-6; 3HRvtBnAbfM9xTkK 1-3; 4M8gQTJhZvRPWf2B 1-5; 4jCpGrhx3NvnDmZ1 8-11;
6vKFBQNC1fMqR3Wa 1-6; 92YFDCN4AvBT7KqJ 1-6; 9Y41wzaRyxQ6CZHD 1; ARqz9awMbfr6QD7Y 1-2; AkBH7pF6XvZc3Gdw 6-8;
AxCwvBkjqYVtLgnc 1-13; Azq6hMZn2RyWwagv 1-6; B7wf2rAVZJLQgT3m 1-5; CcAYgrB3Q1w8KJvV 1; D9wn4C6gKBq8kcZd 1-2, 4-11;
F9DyXdCazKmPqWfN 3-5; GCTbWRf9mrHqxAX6 1-3, 5; J2XL6AyZqgpjx8QK 1-13; JZRnFQDrmYqBycvG 1-7; Kt8wFb3qThrVjNGY 9-11;
LJ43xqm7ATnCH8MX 1-13; LKkr1pxAtT7CvcZN 2-17; LMdRqJGhV17XwWyT 4-22; LjFr1v7pJKVcfhWB 1-11, 13-14;
MZ4cNxKYPzk62q3L 1-12; NXgv93kr2G7zyTxm 1, 3-9; NZn9hmTJHvcMRbfY 1-2; Nh9wMxkpzqJTPA2f 1-8; P1QTDgLwfytJKrBh 1-8;
P3JYyjrc1FWgfN6q 1-4; TGmKaXrQPp7czjx1 1-4; VxmDcHTK8LqgNCwQ 1-7; ZjTwPR1D2nVyC7xF 1-2; aMNVKZpg8tRkhBP3 1-17;
btxm3BCN4jX61VhY 1-7; bvzqmdxkA4KVQaZh 3-10, 12; fMkvA2btVy8WwjhC 1; ftLcwNqgjyvP68YB 2-14; g7b3Jf29CVhKtYMD 1-2;
gdMNAjD2Kyrtqh97 1-8; h3bCaBdyYJMPzvRk 9-19; hMaTJ3ZtDzqPWHfx 1-7; p4MR7G2atyhQwV6k 1-6; pbvkXazLqcHFhgRf 4-7;
pwNAZn643XgzcbYa 9-10; pwd6RvtM2NYqmfGT 1; pzx4LKwq6PB8FrWy 1-10, 15-17, 19-25, 27-32, 35-36, 38-41, 43-46, 48-49, 52,
54-63, 65-70, 72-80; q8wrpJkfFWdVg6Tc 1-3; qT76ZPycRHrL41p9 1-4; r9FD3YQGHcqdjC7g 1; wrLPGXTkp7bYnqV6 1-4;
xR6rvTQLBwZ9nkzy 1-12.

## Fang of the Crystal Spider (item 13218, spell 17331): 1, observed

306 procs on 11,511 landed hits, 10 wearers, 11 runs, in the main hand or the off hand: 1.00 [0.92, 1.10]. Observed:
every wearer is a rogue and its interval reaches further than 10% from the rate. Left out: 1 run logged in a second
report. EV-0136 gave 0.9. Reports from 2026-11-14 to 2026-11-18.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |      10 |   306 | 11,511 | 1.00 |

Reports and fights: AaKzvcY1wmH9pBbq 1; D46Xm8ncR9WbTFrL 1-31; KNTYZJ3Xg9d8HftG 1; QXFd74zDW8JZMmKx 12-13;
XPYV9J6DfrMKxGqh 5-6; XgQYbwGBZcHLKRxV 4; c8mjK4TdzV2M7XLF 1-5, 7-46; h3P9VjGMZCJWfTcm 7-10; kaH1G4ztmYpgMTJ6 1-2;
xXWkB3fqAjt9NJGH 5; y6D3jqgWt8TvpNYM 1.

## Felstriker (item 12590, spell 16551): 0.6, observed

1,047 procs on 59,127 landed hits, 17 wearers, 28 runs, in the main hand or the off hand: 0.62 [0.54, 0.68]. Observed:
its interval reaches further than 10% from the rate. EV-0136 gave 0.5. Reports from 2026-11-15 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |       9 |   553 | 32,917 | 0.59 |
| Warrior |       7 |   483 | 25,680 | 0.66 |
| Hunter  |       1 |    11 |    530 | 0.73 |

Reports and fights: 2jJ1L8FPHzTB4fKG 1-5, 7-12; 2wMznDRvmkpbtHxA 5-7; 2wPQmVqbdCcZW9gv 1-12, 14, 17-20, 22-42, 44-53,
55-64, 66-76, 78-88, 90-100, 102-113, 115, 118-121, 123-135, 137, 140-143, 145-158, 160-173, 175-188, 190-203, 205-218,
220-234, 237, 240-243, 245-259, 261, 263-278, 280, 282-297, 299, 301-307, 309-317, 319, 321-327, 329-337, 339, 341-347,
349-357, 359-361, 363, 366-369, 371-377, 379-387, 389-391, 393, 396-399, 401-407, 410-419, 421-423, 425-431, 434,
436-444, 446-448, 450-456, 459, 461-470, 472-474, 476-482, 485, 487-496, 498-500, 502-508, 511, 513-522, 524-527, 529,
532-535, 537-543, 546, 548, 550-558, 560-565, 567, 570-573, 575-581, 584, 586, 588-596, 598-603, 605, 608-611, 613-619,
622, 624, 626-629; 4TxMfmFpzBJ6aN1j 5-7; 7tbWfYqJFd4vzL3p 1-2; B6bvH1fQ4kAcyRzt 1-2; FhG3AxYL9ZNP4RvV 1-15;
M1tPx9DBZGYgCa4f 1-3; NjWKPpHgymCT832A 1-19; PKBd6Vz2Gkjwp9Zt 1-2; QNfaZ9MTnWpFrqb4 1-6, 9-17, 20-28, 31-40, 42, 45-48,
50, 52-53, 55-56, 59-61, 63, 65-72, 74, 77-80, 82, 84-85, 87-88, 91-93, 95, 97-105, 107, 110-113, 115, 117-118, 120-121,
124-126, 128, 130-140, 143-145, 147, 149-159, 162-164, 166, 168-178, 181-183, 185, 187-199; QkVRc329w8nLTzKp 1-3;
TNW9rHBt3MD1X8Ag 1-2; YFtvD87kznr1CThg 6-7; bP2CFzkZTLJ8q6Dn 1-8, 10-16; cgMGCdPJQ84wzaBV 12-13; dnDg7j9kRw6WtbML 61,
63-69; gYm2HAtf4bnC8QZT 6-22; hZLykw9dRnYbtvKP 1-2; jXTVfy3ApWKPYwDG 1-22; jn3dwzWPZ6r2ky1K 1-3; mpgnDQbLWXY8jJP1 1;
nX1qbatTzGZ3CAwB 4-7; tkPZYqTRgCyWG3jM 22-23; vZqVmLYQa6jwP3A9 5-16, 18-21; yPVkKpqtfJM27WFB 1-7; yfKb1MthLFDc4ANG 1-7.

## Fist of the Damned (item 10804, spell 18084): 3.1, measured

3,905 procs on 39,540 landed hits, 23 wearers, 26 runs, in the main hand or the off hand: 3.12 [2.97, 3.36]. EV-0136
gave 3.5. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Paladin |      10 | 2,031 | 22,149 | 2.90 |
| Warrior |      12 | 1,867 | 17,321 | 3.40 |
| Rogue   |       1 |     7 |     70 | 3.16 |

Reports and fights: 1GN2LMtwvY4JQHba 8-11; 29zqLfrTDthavjnx 2; 498vhQAtCw2XgZJM 11, 13-15; 4rfgMCY9qjyvQhHD 1;
4vcn3WF1RxXf8VZt 6-17; 6CN9XfdkHq148yZQ 1; 8ZrM3AyYtCKd6Pag 1; 9MCqcwjzLk4JBARG 1-5, 7-14, 16-22; 9WGcCxKHQ4pkX6Rj 1-11;
9hRCG4AnXxvZB3mL 1-20; AGpM6qwyRL2xbTzZ 1-2; B1qcTGpnHCrQ64Xa 1-4; Bfk6ChZgAr8yWtXv 1-7; By41zJVmM3cdPQkG 1;
JMnDtLGxf4FqCV6w 1-3; KYmwz2q1tG8yAcdP 1-6; NcwFg2bZjkrnXfQV 3; VjT3aMZghKCyzGrt 1-6; Wb8w1TtxA3kLF9M4 1-8;
aXg9mCq3GVDH8WBL 1-26, 28-39; j8bQnLtWwRMBXkDz 1-4; n9cJ3qy86bXrKfvw 1-11; nHNP84kfRTt2MbJQ 1-3; y6D3jqgWt8TvpNYM 1;
yYPpk4TqAhjX3xC8 12; zn64f9D8TVt3wFZq 1-3.

## Glutton's Cleaver (item 10772, spell 18075): 2.2, measured

947 procs on 12,820 landed hits, 21 wearers, 21 runs, in the main hand or the off hand: 2.22 [2.09, 2.40]. EV-0136 gave
2.4. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   441 | 5,704 | 2.32 |
| Rogue   |      11 |   397 | 5,563 | 2.14 |
| Shaman  |       2 |    69 |   964 | 2.15 |
| Hunter  |       2 |    40 |   590 | 2.03 |

Reports and fights: 1RVnhHGv2bwDW9Yr 1-2; 6VJxcqnYa4wdBWQv 1-9; 7NgdQzG6VcRaAJ9X 1-6; 7tCZB31vV2Jp6XWd 1-2;
8vDJ9jmHd4FA7Twt 1-2; CnNwpaYW83cJr6fR 1-6; JcwBmYGVgnMQH9xd 3-19; L7JpRfWtAZ6T1hGx 3, 5; M7BnDF4VPXcWRaZ1 1-5;
P8Znjv1fgBD2CAVY 2-12; PDT2yACNzd4LMZ98 1-14; PJXwqxpANGtya937 7-11; W7VPQB18nHfkMaAc 6-13, 15-31; f8bZ6ytgj7mMGHCk
6-12; h9XqNAJk6aPntHpv 12; kpPdVKFmw72nCyM4 1-3; pjZHqhYvkLKTAQCJ 1-2, 4-5; qnyc1zwaB4FrvM9x 1; wZGf4Y6tjLDyTP92 1, 3-7;
wvbng2zVTYRrc9Zd 1-2; zwPKLTQ4W6YkyRqd 3-8.

## Grimclaw (item 1481, spell 1292158): 1.9, observed

292 procs on 4,618 landed hits, 6 wearers, 6 runs, in the main hand or the off hand: 1.90 [1.64, 2.05]. Observed: its
interval reaches further than 10% from the rate. EV-0136 gave 2.15. Reports from 2026-11-14 to 2026-11-16.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       2 |   199 | 3,047 | 1.96 |
| Rogue   |       2 |    51 |   835 | 1.83 |
| Paladin |       1 |    26 |   543 | 1.44 |
| Hunter  |       1 |    16 |   194 | 2.48 |

Reports and fights: 6rgDRnVFX9a3pbGy 1-4; F1MPmrQNq4v6XJ23 1-2; HJxqkQPAnmKtwNjd 1-16, 18, 20-24, 26-31;
JgL3qFZA4aQcpXPC 1, 3-9, 11-13; KjLXqdCZANvQc9BT 1-3; Lcy3V1wHPpqZQAG7 2, 4-12.

## Hanzo Sword (item 8190, spell 16405): 1, measured

695 procs on 27,532 landed hits, 16 wearers, 20 runs, in the main hand or the off hand: 1.01 [0.96, 1.08]. Left out: 1
run logged in a second report. EV-0136 gave 1.2. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Paladin |       2 |   274 | 11,205 | 0.98 |
| Rogue   |       9 |   245 |  9,844 | 1.00 |
| Warrior |       3 |   104 |  3,425 | 1.21 |
| Hunter  |       2 |    72 |  3,058 | 0.94 |

Reports and fights: 4f2ygdYCzFAvqc6K 6-8, 10, 12-14, 16-18; 4hfTdvYwN7xA8Wa6 1-19; 6xHzpKd8wCna37Lb 1-14;
824kNhdFaALMBnvR 1-4, 7-18, 20-45; DNc6QALhxmGbCH3Y 1-6; GHK2gB78FMVCmPfW 1-9; JjYkyzampgwCVPNr 1-6; KPqgbZxhn6BXwCvJ 1;
MK6XJDzfLpFjnRWc 1-5; bZaY8JyWxpmwQDjB 1; bjpfAyh6Vc7DZ8zY 1-11; hwrmtk6aA7zJ8N1Q 1; jn3dwzWPZ6r2ky1K 1-3;
kaPpXgt1bWJBr4z2 1-3; mtDjVd6a9rTgWBnC 1; qNyQvka41YhVmdHR 4; qbtZdThvBCXVcfAW 1-6; x4TPAZXnW8kbv19j 1-6, 8;
yPVkKpqtfJM27WFB 6-7; zqpJDYHZC4Pa76bV 1-2.

## Iron Knuckles (item 2942, spell 13491): 2.05, measured

349 procs on 5,988 landed hits, 6 wearers, 7 runs, in the main hand or the off hand: 2.06 [1.92, 2.16]. EV-0136 gave
2.4. Reports from 2026-11-10 to 2026-11-17.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   290 | 4,926 | 2.08 |
| Rogue   |       3 |    59 | 1,062 | 1.96 |

Reports and fights: 38D6JN4nZzjcGtHg 1-4; 4QAtNfTXMyDGPqCr 1-13; AzbN8GY9DnBrH4aW 19-31; D8vykhYFL14KH7xX 1-3;
DwWM9H3rpTnGPfbt 1-14; GRTmQn3VDpwZKbCA 1-6, 8; H3vdcCr6JX1gwYtq 1-3.

## Ironforge Greathammer (item 279895, spell 1322295): 2.85, measured

3,561 procs on 24,016 landed hits, 45 wearers, 46 runs, in the main hand: 2.87 [2.77, 2.98]. Left out: 24 runs in Cat or
Bear Form. EV-0136 gave 3.3. Reports from 2026-11-12 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Paladin |      15 | 1,692 | 10,967 | 2.99 |
| Shaman  |      21 | 1,082 |  7,356 | 2.85 |
| Warrior |       9 |   787 |  5,693 | 2.68 |

Reports and fights: 17jHXDAWBRQdgqV6 1-4; 1HK7gFbwy8aJPAh6 1-3; 1fpaDyG3mFnB9KW8 1; 1h9P4rYzWDvtHRxb 1;
1pDcTvkrCBhZJRgf 1-14; 3pdnKMxtZcWkrDgb 1-3; 6HTpWDrd8LKGXwjf 1-5; 6NdR8TAqXMyLmkYg 1-3; 7kYG3K21FmyMVABD 1-4;
7tahbgBF1jQGRKMq 1-3; 8c7DafNhKtndMBzm 1-3; 9Vy4q1cxp2t6BQNd 1-5; A7hdvgfJbNj23Qwr 1; CdH9vgqPbRX1JKM3 1;
HKQW3wCmr4kJXz2j 1-3; J4LX7AMRmxVh96P1 1; L8yaqgDXG26Nf9v1 1-2; LVwmRx1Mqak7vp4y 1-8; LZjdNV9a84k7GPfx 1-5;
QMH23Y1hqCxamFLj 1-8; QmCHxRYD9tbjLfW1 1-4; RHDbqGV3fgjMrLCX 1-9; RHYb2XPvpDdf17LA 1-5; RQxLGd1agBDt6hXP 1-4;
TBYt3rwJqhGxD1py 1-3; VQNXy3G2KFH9RrgC 1; ZADc3bTCKhn9jJdz 1-3; ZPBTJjpyxndNg1Lm 1-2; arQwdZcPC8XYjgRV 1-5;
brAQWYqMfywd1JPt 1-3; fpGWB7am4PzYnMyN 1-3; gXLj6knBf4QFvGD7 3-5; hM67TLJGypgwvrzP 1; jr1xaAPwb4f2HpvK 1-3;
n8QFXC7YMdvBgJRt 1-13; nP1qcYhtNfWAj4DC 1-4; nWhPF7JrzNRLmx2v 1-5; pJN1cYKwfgaZ8jt4 1-5; pqPBFkXgnjDvN9ZG 1;
q2h6pRvdBCKMabnL 1-2; rZNG9mDwxBYnzkJh 4-5; thJy8Y43TPRnHvaK 1-5; yVATBnQjpz1MLZPh 1-2; z9rYLQ1png3TqvkA 4-6.

## Killmaim (item 13016, spell 13318): 0.95, measured

810 procs on 16,414 landed hits, 34 wearers, 37 runs, in the main hand: 0.93 [0.86, 0.99]. Left out: 1 run logged in a
second report. EV-0146 gave 1. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      22 |   467 | 9,382 | 0.93 |
| Warrior |      10 |   275 | 5,672 | 0.91 |
| Paladin |       1 |    62 | 1,146 | 1.01 |
| Hunter  |       1 |     6 |   214 | 0.53 |

Reports and fights: 1g7hAY6CBNtW84HX 1-2; 29A3dKwPgpRNnWvc 7-11; 2aVnTxFzfckdXpBy 1-7; 3vDf6TMYBX9LK7bz 1-2;
3xgPkwfn8yQ9BqAW 1-4; 6hrcXFkgDamVB8jT 1-3; 6v4wTWQkjb3Xd7mM 6; 7hNgRwMfkQZ2Yv4b 1-3; 8GW16QY9BmFHJXLb 1;
8RFv6Xyr43dLpJ7h 1-2; 9b1wvaKGpHdNJCZT 2-4; G3VKdgR2aLAkNqWy 1-3; MrW6JCFZTDzdtYpH 1-4; P16dx3VLmRyaDJZQ 1-18;
P7cV2HjLgxGpmZQk 1-7; QhYmjJ4bHZXcgfk1 1-2, 4-27; QqKcMJwHjxNPkgRT 1-3; T6MXt73janNJgmyY 1-3; Vvfz8wYbjTBFaxRq 1;
aQwNVCPHWmzDLXqk 1-2; acHy1fNktP2m47pd 1-2; bdL8j9KGMR6r2xmk 1-3; d7DFtXL2jfNa8Bw4 1-5; fn2Dc941K63CNvmj 3-4;
gdKNPjVvcQt2LJnq 1-3; ghzjGKVBvJbRZnMr 1-2; grFA3BHxtCNyzR7v 1-2; hbW1zgMXyRv8JAVL 1-6; jKJPF16ZnWLkrR2d 1-2;
nT2Wz7mVQZbtBRG4 1-4; nyhMJLqBYxPpRaXf 1-2; pDRxq6wJynmhgdjL 1-9; pqr4HxC3vWJhfG7B 13-20; tMxrNWC6qPVzgwAv 7-8;
vJGT19YL26XFQpNB 1-2; zJ6C2148dTYfaXHb 2-4; zhj69Dt2X1b3fvMT 1.

## Loksey's Training Stick (item 7710, spell 1293482): 2.15, measured

902 procs on 8,161 landed hits, 35 wearers, 36 runs, in the main hand: 2.14 [2.02, 2.26]. Left out: 16 runs in Cat or
Bear Form; 1 run logged in a second report. EV-0136 gave 2.55. Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   389 | 3,428 | 2.20 |
| Shaman  |       3 |   302 | 2,811 | 2.08 |
| Hunter  |      29 |   211 | 1,922 | 2.12 |

Reports and fights: 1aBhkzdDMZKGtb37 162-170, 172-176, 178-187, 189-195, 197-237, 239-286, 288-289, 291-295, 297-312,
314-323, 325-328, 331, 333-335, 337-371, 374, 376; 1xqBcj4CZQXV38Pv 1-2; 214ndcFyjpPM8wah 1-3, 6-9, 11-15, 19-28, 30;
27jZGhfHgp1LR8MB 1-11; 2PpAjf1qRmTv6gDz 1-17; 2WH4jLt3FADy8fnh 1-2, 5-6; 3HJrg1BGh8xR6Y2z 1-13; 7fgdkyv1njDKa8Gc 6;
8ZdnfAT79zPwmyVM 1, 3-6; 91ZzNbJMRtjrLW6K 1-2; CpDKr6AWJ4fw7xhg 5-7; GBDAWjKNh6Q9XnF7 1-4; J1QrDzNc4ydjV8ak 1-10;
JcwBmYGVgnMQH9xd 14-19; MgcQ6txHr9R1VjPv 1-5; NkyKCr2XmgLt1YWv 83-87; PmR8Lpvazd4q61nA 6-12; Pn2LYHXpjZMDywgA 7-9;
Q8GxhTmH9f6ajcNr 6-10, 12-39, 41, 43-51; TnLAmyRdXvz1cBhC 1-18; V4HDKvxmYCkR1JnB 1-3; Vb3ZxfXtyDWr1ndP 1;
WYrCyJkdfMXw84Ph 13-21, 23-35, 37-39, 41-48; Wf26RdmcnLDTXwV3 1-2; ZjdqP6bg7tTR894y 1, 3-5; bCzt1Fq24mYcPXkB 1-2;
bT3LyRBtax8FQnch 1; c2J4C8kzHaQxnwXW 1; cNhpPv28MHBJCwWm 1-2; ftkqY3cwR7dBZK1J 1, 3-5; jGhpmnHrtyc1VD8W 1-4;
kH2bjrXh4BF39GmC 1-2; p64RG9PVZcDg1XCk 4-6; pqfK2AgHZ7DnhvPF 1; rD8m4JW1bVyYPKGX 1-3, 6; tHz9M3arKp6PxA1W 1-14.

## Lord General's Sword (item 11817, spell 15602): 1.95, measured

6,172 procs on 72,714 landed hits, 20 wearers, 33 runs, in the main hand: 1.96 [1.82, 2.05]. EV-0136 gave 2.1. Reports
from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      10 | 4,091 | 46,177 | 2.04 |
| Warrior |       7 | 1,547 | 20,563 | 1.74 |
| Hunter  |       2 |   319 |  3,570 | 2.06 |
| Paladin |       1 |   215 |  2,404 | 2.06 |

Reports and fights: 24kgnZAXbKTW89x6 1-2; 26AYbFNQzMZCmfkh 1-16; 6c8P4BLaVAj3bpK7 9; 7Kz3LtgfDmbBRGXP 4-5;
7TXcFNRzjwHdAybn 3-19; 7dgfrZ3bvhVqxtY8 3-14; 7tbWfYqJFd4vzL3p 3-6; ARFXzbqHDpBLJdhy 1, 3-4; CDKVcP6ghZJtv8AW 3-6, 8-16;
HwyJMCLKfbY1PkmG 1-8, 10-38; JfNv1pWML9ZA7ghR 1-4; K6VvbDypXFLQJW7d 1-27, 29; MgC1BXh8fyArkYqa 1-30;
MwVH1KQmak73NPd8 8-16; NK6cCM4pJyRhaAFW 1-7; NXMZV9xjvWpB7GtP 10, 12-107, 109-116, 118-122; NcxbzW4kJYQ7DGhj 1-3;
RVkGfcrbZ13MWqQ2 1-3; RW7kfjbYghJyx1dK 1-25; TLp1jza6PCrBAX94 1-7; Wn8wHhaxy31MLzjp 1-10; X8z9dgD6HRTZ24Vc 23;
fCB4wQt3NWm6k8M9 5-7; jn3dwzWPZ6r2ky1K 1-3; pPbM8VAdmWCTvLB3 1; pQPwFg8rf6VWhRNc 1-3; pgN9QwMLYVfrqGbv 1-4;
q21PmQwWRfMLvKXc 1; qbtZdThvBCXVcfAW 1-6; v1bPdgDtXypHncfN 1-4; yPVkKpqtfJM27WFB 6-7; zKgGQA6BqvcJtYZ9 1-3.

## Nightblade (item 1982, spell 18211): 2.1, observed

3,535 procs on 29,012 landed hits, 22 wearers, 28 runs, in the main hand: 2.09 [2.01, 2.18]. Observed: every wearer is a
paladin. EV-0136 gave 2. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Paladin |      22 | 3,535 | 29,012 | 2.09 |

Reports and fights: 1NBKq2mDWghwR69z 4, 6; 1t43QDrG6pfbRAnH 1; 4PtKGL1zcRV8Chfn 1-3; 6HtbKfTNRYjq1cVy 1-4;
7BQKZXWzY9MaxAPq 1-4; 7c21XafhWTAdqNGY 1-10; 8cVPaCpAMmv93bkD 1-9; 9WAxvQbj4HMgJVfZ 1-3, 5-16, 18-21;
DXTzdJWyxk7bm1V3 1-5, 7; GvAQn76tacyqYPMw 1-3; HGKLB9Yvhbxdr81X 1-6; LPQd7pfbjDkYZyRz 1-2, 4-5; Lg82ZFDyQ4MvRm9p 5-11,
13; NcwFg2bZjkrnXfQV 3; P1LGvYV9FdmR8x3n 1; Q1jMALz72T4RPgpW 1; RQjfLHpktThvGVc7 1-8; TB8FMm9QKW2Htb6D 1-4;
V3cYyjhtTgbdx82a 1-2; ZqwpkCPmLMVYN8th 1; aANyFDgb69hxqY2w 3; aNh3Dt8kbZHL9cBM 1; bT3XwWamfrp7FV26 1;
dZ6CjB3hqarMzvbT 1-8; g4Y2frPx8kBmj9Tq 1-2; xJHzjrcgt1kvAWKn 5-8; ymkqDYNTxPjCgwnb 1-12; ynVw7ZgMx4tjF3fQ 1-6, 8-18.

## Phantom Blade (item 7961, spell 9806): 2, measured

728 procs on 13,570 landed hits, 12 wearers, 13 runs, in the main hand or the off hand: 2.01 [1.85, 2.19]. EV-0136 gave
2.25. Reports from 2026-11-15 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   418 | 8,074 | 1.94 |
| Hunter  |       1 |   145 | 2,330 | 2.33 |
| Paladin |       1 |    96 | 1,910 | 1.88 |
| Warrior |       1 |    69 | 1,256 | 2.06 |

Reports and fights: 2YfdbFMgGhr7zHwQ 3-4; 6kN78G2xfqy3vwTP 1-6; D2d3hNAbwkPgL1vF 1-10; GzZH1WhTLQ3JaDXt 1-11;
J23GjB7aNAKgrLqM 1-6; QAHXb9zy4jfgKPVJ 1-6; XFdcWaPL8MthYH6A 1-5; dt4M3ypRnTkwaVQY 1, 3-10; hrV3QgyzvfdZKJwF 1-30;
pLtJjx8DCPbHRacv 1-21; twcjmZNGMQP2FABR 1-12; y2dFcNR1rMmKTPAj 1-25; yqFPH1MKv746dQhB 4-7.

## Plaguefang (item 279876, spell 1309315): 1.95, measured

3,901 procs on 57,860 landed hits, 64 wearers, 72 runs, in the main hand or the off hand: 1.93 [1.84, 2.02]. Left out: 1
run logged in a second report. No rate was recorded. Reports from 2026-11-05 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      17 | 1,503 | 20,734 | 2.07 |
| Rogue   |      27 | 1,254 | 19,462 | 1.84 |
| Paladin |      13 | 1,078 | 16,620 | 1.85 |
| Hunter  |       7 |    66 |  1,044 | 1.81 |

Reports and fights: 149bNrXGaKgQL38P 1, 3; 1Hgm9RqadvcjJh2w 1; 1HpaBdm9fFXDjYPy 1-4; 1XCND7YQJ86Ax32q 1, 3, 5;
1ZJVMgcYWTyRQh2w 1-3; 2KRt1WLqNwGQ3Cpd 1-3; 3tPjXKNACMHcn9b7 1-2; 47Vj3cfMC1nNgbXw 2; 4ChPnBrVK2fJLMaX 4;
4hNWVC2Kzj8RxpZL 1, 3-4; 6TQkbpNXgGZwYat4 1-3; 6tANqvfKQP1Fkdzg 1-2; 7DKRq39nf4yvFWLG 9-14; 7NbYtg6T2pjr1yL8 1-2;
8cDyAtVnkHv3qfMb 1, 3; 8mBLvfVzCn4tWyTM 1-3; 93ygXAq67nFGMxBh 1-3; ARBxdp9znVDMTwJ4 1-4; AgFRCzn6X2KDdtMr 1-11, 13-18;
BgCQ6Xzp3Nh2HWkP 1, 4; Bxk8b1hKMz6vt9Vd 1; CJZhDp9FxwRMaWYy 1-2; CfhkbPBKAZFRGp9g 2-3; F2ZtrHh9fqaL84Rx 4-6, 8-16,
18-45; FcC4vd2DkRqyMrVT 1-2; GAK2BMH8Wt3k9JFj 1-5; GcjZ39HkxPXBAq7v 1-2; HPRbL3N7p6wFhrqk 1-3; HpjCAgJKaR9BN3Lm 1-13,
15-21, 23-27, 29; HqQTwzArBZpPN1y3 1-12; JDH7bFz23Qakdgnh 1, 3; KAvPryL1zCdtWTxV 2, 4; KWZaQjqPLJ7Ry8Gd 1-2;
KYD67nVqZT4jxWyN 1-6; L7QHtkhacypgd3D1 1-2; MzwcDrG3fdFBRy6a 3-5; NX6RBWQpm1ytnbZF 5-7; PqjRDCnyaHt13QJb 1-2, 4-5, 7-8,
10-11; QDNfAvVR9Z3CxaYp 1-2; Qb7TJqYkKtgyz1ah 1; RB47hLrMjdDG8Z6a 1; Rp4QCr31TVmgyqnz 1-2; TH63PWxNv9gRarCY 1-2;
TnbVQWRJ7r26CqpK 1, 3; W8YC9TA6bcXwLRMk 1-6; Z4GKmBf1HXdtkYDL 1-5; a8cL9vKhNmkQfBGd 1-3; a8cL9vKhNmkQfBGd 1-4;
aVyXCj1NpnRTFKxt 1-2, 4; afNcyKHdtY2qjPCp 1-4, 6; bcATfGPv3DX1Jnxy 1-2; d7DFtXL2jfNa8Bw4 1-5; dzTmC6F7xwLPJrq9 5-17;
fvGcC6WLJxHAkXb8 1-2; gKmcRvAdBYqrDnbt 1-3; jKfkzadV3BLp1cbr 3; kLB1K4faPvhmGzqV 1-5; m2FqGpwzr9cQf7PK 1;
mnaL3tDKH4QZGrWJ 12-13; n6DtHVNKp1wgymFL 1-2; nWxftVpvMycmw2FL 1-4; q8DXKyGLJFb1CwVp 10-13; qtcFwCBmkXz7JMH8 21-22;
ryqgQphLxMBz4TmH 1-4; tm2FQVrcZzfB97v8 1-2; vYQqkXLm43RJZGy7 1-3; x9vCJk6RYqFLhWGa 2-5; xXtbFrdR4qKAQVB1 7-8;
yLKXxQgrPWTnmdZA 1-3; yMNq3krVBAHxG68w 1-3; yTDCtLmVK4kq7gQ8 1-5.

## Sang'thraze the Deflector (item 9379, spell 1294243): 2.1, measured

2,787 procs on 47,139 landed hits, 27 wearers, 31 runs, in the main hand or the off hand: 2.09 [1.98, 2.22]. Left out: 1
run logged in a second report. EV-0136 gave 2.35. Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      15 | 1,518 | 25,601 | 2.09 |
| Paladin |       5 | 1,057 | 17,948 | 2.08 |
| Rogue   |       5 |   158 |  2,643 | 2.11 |
| Hunter  |       2 |    54 |    947 | 2.01 |

Reports and fights: 1CYAWTtmcGdFqbPN 1-5; 28AhM3VGnvFcJ4rD 1-4; 4NqLGxvrZYX2Qwnf 1-4; 6bM1k938LGNghacx 1-9;
8FBqfvLyWTdKxbZh 1-38; 8KPHjdAxZYFNLtgR 1-5; 987q1pjnYdVQbCmT 6; AXn7cJ9VQ6YHLWmh 1-5, 7-9, 11-13, 15-18, 20-21, 23-30;
C3zW7BjkJXgRtq8m 1-11, 13-18; FAL4Hv1D7nWf8VNh 1, 3-8; FVv7P3a2KqJrdG4g 1-3; KfvWRDrFXQ1TMq9P 1-8; LXWzFKcBmvyMZaH2 1-2,
4-9; Lg82ZFDyQ4MvRm9p 1-2; NbwyAxHKzhYXV6L2 1; R9bp4JGMD7xzchAf 1-3; VQCRJWfq3KX7M1Nh 1-2; WYFZHxvL7zg6ydnQ 1-6;
XCJdxAF17zqHMhLY 1-2; XHDjdLb7R9P6NQm1 1-2; XWzyRfva8ZK2T6pc 1-4; ZLpcV6krn1wYHJvt 1-8; bvzqmdxkA4KVQaZh 1-10, 12;
cPLzXv4nTtNaYWF3 4-20; jyD2wdc6FgZfHJqh 1-7, 10-11, 14-16, 18-20, 22-34, 36-41, 43-47; r4fBkAMgmdxvqYCD 1-2;
tCprA1RmvTcJkq3D 1-4; tnyh4dxq6pNXQzLw 1-3; w63W4VhTgCbt9knB 1, 3-10; yQnph9W1VB4z6a2J 1; zLacx86Thr3AMtqn 1-15.

## Satyr's Lash (item 17752, spell 1292155): 1.95, measured

918 procs on 16,691 landed hits, 22 wearers, 23 runs, in the main hand or the off hand: 1.94 [1.77, 2.09]. EV-0136 gave
2.1. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      16 |   668 | 12,277 | 1.92 |
| Warrior |       2 |   204 |  3,573 | 2.02 |
| Hunter  |       4 |    46 |    841 | 1.93 |

Reports and fights: 4NqLGxvrZYX2Qwnf 5-8; 4PtKGL1zcRV8Chfn 3; 4YzXaGtg81WdqM2N 7-15; 6zJNTjMwWD31vrgy 1;
Bfk6ChZgAr8yWtXv 1-7; C8Qd1LXMrGT62kpY 8-9; L2c7jBZNwFPpb1tV 1-4; Lg82ZFDyQ4MvRm9p 14-26; MKmntyr82Q4v7jDh 1-5;
R1acrQhyNvwXLmBj 1-3; T71XnfaF3cY9M2Rg 1-13; ThKD4jr8MfawcCPv 1-19; V7Xna6bcWLRzZ3GN 1-6, 8-11; W1zqgCZVPAtTM27b 1-5;
YamkzNdPTZFLAnGq 1-8; Yf91cBLN26jdqrxC 1-4; ZwxX9G3bRPFYfjng 1, 3-4; d4LFCkmQYhPRDx2y 1-6; hyCdAYx8GjgKfZHp 1-5;
qm1Bvy29hGK6bMaQ 1-2; tCprA1RmvTcJkq3D 4; tdYLMmfgcAJCTWyr 8, 10-12; yJmtnzPNDVK42TQR 5-8.

## Silent Fang (item 13953, spell 1293654): 1.55, measured

816 procs on 19,726 landed hits, 11 wearers, 11 runs, in the main hand: 1.55 [1.41, 1.70]. The sim had no rate. Reports
from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Paladin |       4 |   384 | 8,706 | 1.65 |
| Warrior |       5 |   359 | 8,660 | 1.55 |
| Hunter  |       2 |    73 | 2,360 | 1.16 |

Reports and fights: 2qMm9WvRpzYx3GaB 1-2; 4hHBvakRnT7p1GVK 6-7; Avb4ZFMdQYj6wfaN 1-3, 5-7; DmjBrYd7M9VZwCHF 14-22;
Hwq97xLh6gv8ZPrY 1-4, 6-8; Ya9kKCWRwqjFpgz1 3-4; cgmhtvNdTGCn7f6L 1-25, 27-35, 37-38, 41-50, 52-84;
fMLcHzyZvt7QTCmg 1-9; nyAx3XNDVmYWGdrh 1-6; pNKJR39CwGnmXtbL 1; zvbmjntP89KHyC4R 1-7.

## Strike of the Hydra (item 6909, spell 1292657): 2.15, measured

5,740 procs on 49,004 landed hits, 71 wearers, 83 runs, in the main hand: 2.13 [2.04, 2.22]. Left out: 1 run logged in a
second report. EV-0136 gave 2.25. Reports from 2026-11-08 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Paladin |      33 | 3,861 | 31,748 | 2.21 |
| Warrior |      24 | 1,752 | 15,899 | 2.00 |
| Hunter  |      14 |   127 |  1,357 | 1.70 |

Reports and fights: 12A94BJt3KCFZrY8 1-12, 14-21; 14caBjX2QdynLkq3 1-9; 1a3pDtCcPgBkL6wm 1-22; 1jgtdNmDMRycbZrP 1;
1mRKL7Tb8qwjkgv3 1-2; 1yjwF8MxXqvtD4bT 1-3; 23ZyGvk7BXAVDT4F 1-14; 3pqhtzPd9y6g8xvf 3-5; 3tRNmGkZB2Jhnv6b 5, 7-9;
4ZLvQP7jk3hAdGfq 1-8; 4y8nq6Xjg93zfrc7 1-9; 69atW41MxnTY2cXN 1-26, 28-37; 6X1L4nB7zCfqk8vV 1-2; 6pjNwdr12qzBAKhW 1-2;
7GNK4Y9bkH1awZxL 1-10; 8XFcyfRB6Tr3PJqn 1-6; 8apqXKQg6mbC91L3 14-18, 20, 22-28, 30-36, 38, 40-46, 48-49, 51-71;
9nCtfP6kvHrzV4FL 2; 9zabj3grQkGB2FKL 1-17; AqxY6kh43jc2RPaT 16-17; B3VmZw1jFDxL8pvQ 2, 4-13; BKnf4rNwJdcqpQ18 1-6;
BYjATckCnW8vNyFJ 2-6; G4hzybWYmgrXvL83 1-2; G8Chy3tNK6jFaApP 1-2; GMBKTFzhbpgDy3rN 1-15; GZzaJH1QVjMKd8qA 23-34;
Gb7JFkf2PypNRzQx 1-4; H3bRaDmzyjxTW2BJ 1; HARvQ2hJWp8f3grD 1-9; HpftBG4KPCZrMxbF 1-3, 5-9, 11; HqW4rjd6XmYgDM9F 1-2;
HtQr7kYaNq9MGgvZ 1-12; JBx4rPjFp2QbcHXA 1-12; L9k3ND6tw8VvcgjP 1-7, 9-18; LHRhtADNW8GfyYqT 1; LcjBAPhTgdnRCHQG 1-5;
LkdvJ8Dq69aMQwWF 1; MLhC3KjWrngq2vtZ 1-2; MQwzXcfxb4v7hnFy 1-2; McYJrBbXAfKj82W4 2; Mk1AY4Km7rvFanHj 1-7;
N178cBjPtChmbk4w 1-7; Pc3gyq8AFL9CD64m 2; QCLPznqJMZ8vH3mf 1-12; RzMcVwxLp4HZvk1Y 1-17; TDqbRJWAN3x1FGhv 1, 3-9;
TkG8PRj7hZXx92Nz 1-19; VPB6pC9HQjJMF8hT 1-2; VQaYAf1Wxvm6ndHK 1-11; XF1faW9mcYtKrqMC 1-18; YX61J8CKqc42paNZ 1-4;
ZDYjbLwfJTaV4GHq 1-6; Zf21kBmtDM3AyRjF 1-13; ZfcYB3RHzxJmajqv 1-2; ZkTHQfR8F13AdaMq 1-3; adXPnhT3kjHp9VcC 1-13, 15;
azpj2hg1DVwnMCkZ 1-21; bdYgfJk7Mh8GqZ4t 1-5; bfzwm387x4pRaNVd 1-5; d9Zjq674VMTh2Pat 1-3; dQMKjpxJBCz1wrDW 1, 3-12;
fnYThW3BpxKZv7jw 1-18; gXy73KBFQG8RCtcm 2, 4-7; haHcQwBDzK2MqJCn 2-5; hgd1qH4QYFyjz6AK 1, 3-6; j1CGayX43bMdF8hP 1-3;
jtn8ap4XY6PzwDQL 1-11; kQ1WR6vTKX2pwaLA 1-4; mFMLvknrhNyaARtB 1-3; mHVMN9yBdkFYCQZa 1-2; mfHcFPYJKLXdg9T7 1-2;
nJtB3F7cQy6NmzGZ 1-12; nKdyat38zLA7fNJq 1-18; rRLkva7gz9ymhb4V 1-4; tAnQr4PNCvW7zwX9 1, 3; tGqwYmXgrTMn89jc 1-7;
tQYfJ87KBMLXqzG9 2-3; tVFZybKHRWp8xMcf 1, 3-10; v4QYKcD7XymZq2JB 1-12; wbn98vT736Qat1Kk 1-6; yaYtkz3mNhTWwxAJ 1-11;
ztD4pKHA6jGvfcVr 1-2.

## Tainted Pierce (item 8225, spell 13530): 1.9, observed

631 procs on 10,356 landed hits, 9 wearers, 10 runs, in the main hand or the off hand: 1.92 [1.69, 2.11]. Observed: its
interval reaches further than 10% from the rate. EV-0136 gave 2.15. Reports from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       4 |   265 | 4,062 | 2.06 |
| Warrior |       3 |   198 | 3,637 | 1.72 |
| Paladin |       2 |   168 | 2,657 | 2.00 |

Reports and fights: 4na1LcCpP8QHTNVD 1, 3-8; 4wTtzhPfDqnQWNy7 1-4; 8WfLRh1qTpFxJgra 1-2; J97GvmrRXCFZfQ8a 1-2;
N3Y1mFkJ9Dwr74bA 1-9, 11-12; NXgv93kr2G7zyTxm 1, 3-9; R3CDcTfm6bdNJrHx 1, 3-7; gn4HAC1RqMfLt8wN 16-21;
pFNnChR7GzAX6aH3 1; qcrKJGPvkfhAtdjQ 1-4.

## Taran Icebreaker (item 2915, spell 16415): 0.95, measured

710 procs on 11,908 landed hits, 11 wearers, 12 runs, in the main hand: 0.94 [0.89, 1.00]. EV-0136 gave 0.9. Reports
from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |       8 |   583 | 9,637 | 0.96 |
| Paladin |       3 |   127 | 2,271 | 0.88 |

Reports and fights: 1MA3qWLGa8v2hdXP 1-48; 2wFyAGLf1CqZKnaz 1-12; AhVvc7rK9dRFQMD2 1-8; Bz8CAjZnNyXhgTtH 1-14;
FRmWJKpCD8t7QGb2 1-3; PJHx4V21KhfCaryz 1-2; aAcjGPbyZXNTzR1K 1; aLwy1zjnrYFhvpDq 1-16; aQAhMLzBbWJcdq6G 1-6;
cLxk98bVfwpXWaqN 1-21, 23; g2v7mTHfctMzhN39 1-9; mg3MBw6rcNtZpjbA 1-14.

## Thornspike (item 6681, spell 1293183): 3.45, measured

3,654 procs on 37,647 landed hits, 45 wearers, 49 runs, in the main hand or the off hand: 3.43 [3.19, 3.69]. Left out: 1
run in Cat or Bear Form. EV-0136 gave 3.75. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      24 | 2,370 | 24,893 | 3.36 |
| Warrior |      11 | 1,031 |  9,824 | 3.70 |
| Shaman  |       3 |   193 |  2,314 | 2.94 |
| Hunter  |       6 |    60 |    614 | 3.45 |
| Warlock |       1 |     0 |      2 | 0.00 |

Reports and fights: 1VKQY2Dafc7PJzpR 1-2; 1Zv6bBfgDP3RVFYa 1-2, 4-7; 1kWMb9cQ67dJhrZz 1-3; 1w8AbW4RBNLX6Kfy 1;
21PKRb9nk3HdYFm8 1-11; 2K4Z1jLyGWzdnD3f 1-8, 10, 12; 2KkaByn6H9mzb4Yq 1-4; 2t6mDPLkfWBxcr4q 1-18; 39BFWKC2J8RTzyMx 1-3;
3f4WnNCQJX9azxmV 1-5; 3npdHy9Ag7rMTC6f 1-5; 72gPMatGLpwd93jW 13; 76ZV8FQ1RHDhbq3z 1-4, 6-9; 7M1KwnDZkXvyRFB3 1-5;
9B8Wf16dHKjgaMDY 1; G6wYN314W9FfVXpr 1-2; GCyghdxmrfYL8zcX 1; GRZrtByWvML6mzXA 1-11; KNqTmpW6HZCnc378 1-4;
M4CvXLZz3qm6pDRc 1-3, 5; M4CvXLZz3qm6pDRc 2-3, 5; MdaKf7zpNgWHF1Jj 1-4; NZcYhqPxkrJMKz6a 1; NyZgJMVBXHn8vtWk 1-8;
PFGvpDf9WZgcY6Rk 1-12; Q4GadwADtxbFBNW9 4-8; QxJKAphaZt61kzvg 1; QyYtrwM6zKPXHDj3 11-15; RdmgVNQ4XJMfnC1F 4, 6-10;
VMP6Dv2pyTnYhaKJ 1-3; VYb6f27NazpmCch3 1-2; Wza8bdXVRhcgqGnZ 1-5; YJzQRVL4r9DqFcpf 1-16; ZCRLWJqmb7v3gjaN 1, 3-7;
ZPK6pnD7HzmyhvjT 1; b12KD98fndF4hMq3 10-18; fnMH4F2pkGC8T6Kz 1-8; gbFJzMXnTKkxNQY7 1; gdMNAjD2Kyrtqh97 1-8;
mDbPCdX9AvwYJxGa 7-30; qMD3CVapwb6nAgx2 40, 42-43; qg8YWNaCdDvV7XZM 1-8, 10; r67HcKRnkYdzTQZP 1-19;
rbg3wR9B7dKf8JmW 1-3; rpLZ287nfKWyJGtb 1-6, 8; waYzfQgyrcKV7H9J 1; wvbng2zVTYRrc9Zd 1-2; xJbw6RTvtHPncNzp 1-5;
y6kYARn4dpCKavZ8 1-3, 5-6.

## Wyvern Tailspike (item 5752, spell 16400): 2, measured

607 procs on 10,078 landed hits, 14 wearers, 16 runs, in the main hand or the off hand: 2.01 [1.82, 2.19]. No rate was
recorded. Reports from 2026-11-10 to 2026-11-17.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       8 |   386 | 6,419 | 2.00 |
| Shaman |       2 |   209 | 3,366 | 2.07 |
| Hunter |       4 |    12 |   293 | 1.37 |

Reports and fights: 2RdqgkHFNVMnKCh8 1-2; 4NWagKbPRwT81yh7 3-8; 8129XG6zfMVJrvmg 1-3; Amh4kVqaFGB3xPZ9 1-3;
JD2ag1wYV8qCxhtQ 1; JPymD8BKctgqMFLH 1; MLhC3KjWrngq2vtZ 1-2; QZyPVgTvLJnamk84 1-3; Wh7fgRNMmLP4YBnX 1;
b6QHjXT2GF7A1YPr 1-8; cT71Vr9tPDKkxWv3 1-2; dGhawD4rxCAH7bR2 1-4, 6-10, 12, 14-21, 23-30, 32-37, 39, 41-51;
fyb2xc14zNRdvDMY 1-2; hxZPKTfQYzdVjvN3 6-9, 11; rqhGLfWyaAtZvjTC 3, 6; vnDM8K4RYx6kW2GH 1-2.

# EV-0153: Proc rates of 35 weapons from the Warcraft Logs reports up to 2026-11-18

- Claim: each weapon below procs at the rate its section gives, priced on each landed hit's weapon base speed. A rate is
  measured when 5 or more wearers of more than one class give it and its 95% interval lies within 10% of the read either
  side, and observed otherwise: 23 are measured and 12 observed. Paladin wearers are left out of every rate but
  Nightblade's, which only paladins wore. A melee ability that deals no direct damage (Sunder Armor, Rend) rolls weapon
  procs when it lands, as a white swing does, so its landing counts as a hit.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every indexed wearer of the 35 weapons, their streams fetched on 2026-11-19; 1,861 reports in all for the rates, and
  the 60 reports the section on melee abilities lists
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

Three things differ from EV-0136. The interval is the wearer bootstrap alone, where EV-0136 took the wider of it and the
per-hit Wilson interval. And a melee ability that deals no direct damage counts as a hit when it lands. Its landing logs
as its debuff applied, stacked or refreshed, with no damage event: Sunder Armor, and Rend's application. Each ability,
target and timestamp counts once.

The third is that paladin wearers are left out of every rate, since their count of rolling hits is unsettled. They read
above the other classes on some weapons and below them on others: Strike of the Hydra 2.21 against 1.97, Ironforge
Greathammer 3.01 against 2.81, Diabolic Skiver 2.20 against 1.77 and Silent Fang 1.65 against 1.47, but Fist of the
Damned 2.89 against 3.33, Plaguefang 1.85 against 1.95 and Blade of the Basilisk 1.03 against 1.09. A rolling hit the
count misses would explain the high reads but not the low ones. One candidate is a Judgement debuff that a white swing
refreshes, which rolls on its own (EV-0145). Each section names the paladin wearers it leaves out and their read.
Nightblade is the exception: only paladins wore it, so its rate is their read, and it stays observed.

Scripts, in `forever-ppm-review-1108`: `analyze.py --skip-class Paladin` for the rates, `analyze.py` with every class
for the paladin reads, and `targets.py` for the rounding and the measured or observed call.

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

| Weapon                     |   Item | Rate | Read [95%]        |     Procs / hits | Wearers | Call     |
| -------------------------- | -----: | ---: | ----------------- | ---------------: | ------: | -------- |
| Angerforge's Battle Axe    |  11816 |    2 | 2.01 [1.91, 2.12] |      697 / 8,001 |      10 | observed |
| Archeus                    |   2000 | 0.95 | 0.95 [0.83, 1.12] |      224 / 4,273 |      26 | observed |
| Barbaric Crossbow          | 272999 |  3.3 | 3.29 [3.18, 3.41] |   3,276 / 29,847 |      48 | measured |
| Blade of the Basilisk      |   8223 |  1.1 | 1.09 [0.94, 1.24] |      157 / 4,564 |       5 | observed |
| Bleeding Crescent          |   6738 |    1 | 0.99 [0.88, 1.11] |     484 / 12,273 |      29 | observed |
| Bloodrazor                 |    809 |  0.9 | 0.91 [0.81, 1.02] |   1,310 / 31,842 |      15 | observed |
| Bow of Searing Arrows      |   2825 |  5.1 | 5.10 [4.86, 5.36] |   2,724 / 11,860 |      17 | measured |
| Burning War Axe            |   2299 | 0.95 | 0.96 [0.91, 1.02] |     654 / 12,350 |      24 | measured |
| Cursed Felblade            |  14145 |    2 | 2.02 [1.97, 2.06] |  7,759 / 100,366 |     108 | measured |
| Darkwater Talwar           |  11121 |  2.9 | 2.88 [2.81, 2.95] | 12,319 / 116,459 |     132 | measured |
| Dazzling Longsword         |    869 |  2.6 | 2.60 [2.43, 2.75] |   4,335 / 38,485 |      37 | measured |
| Diabolic Skiver            |   9475 | 1.75 | 1.77 [1.66, 1.93] |      333 / 3,886 |      13 | measured |
| Electrocutioner Leg        |   9446 |  3.9 | 3.90 [3.86, 3.95] | 32,143 / 290,507 |     202 | measured |
| Fang of the Crystal Spider |  13218 | 0.95 | 0.96 [0.86, 1.05] |     355 / 13,863 |      11 | observed |
| Felstriker                 |  12590 |  0.6 | 0.62 [0.53, 0.68] |   1,159 / 66,083 |      19 | observed |
| Fist of the Damned         |  10804 | 3.35 | 3.33 [3.09, 3.69] |   3,018 / 28,663 |      20 | observed |
| Glutton's Cleaver          |  10772 |  2.1 | 2.12 [2.01, 2.25] |   2,354 / 33,378 |      44 | measured |
| Grimclaw                   |   1481 | 2.05 | 2.05 [1.92, 2.21] |     754 / 11,031 |      14 | measured |
| Hanzo Sword                |   8190 | 1.05 | 1.03 [0.95, 1.11] |     470 / 18,281 |      16 | measured |
| Iron Knuckles              |   2942 | 2.05 | 2.04 [1.91, 2.14] |      374 / 6,461 |       7 | measured |
| Ironforge Greathammer      | 279895 |  2.8 | 2.81 [2.76, 2.86] |  12,393 / 85,376 |     223 | measured |
| Killmaim                   |  13016 |  0.9 | 0.92 [0.85, 0.99] |     748 / 15,268 |      33 | measured |
| Loksey's Training Stick    |   7710 |    2 | 2.02 [1.96, 2.09] |   3,474 / 33,273 |     177 | measured |
| Lord General's Sword       |  11817 | 1.95 | 1.94 [1.82, 2.03] |   7,100 / 84,624 |      21 | measured |
| Nightblade                 |   1982 |  2.1 | 2.09 [2.01, 2.18] |   3,577 / 29,368 |      23 | observed |
| Phantom Blade              |   7961 | 2.05 | 2.06 [1.87, 2.25] |     679 / 12,356 |      12 | measured |
| Plaguefang                 | 279876 | 1.95 | 1.95 [1.82, 2.08] |   2,929 / 43,023 |      53 | measured |
| Sang'thraze the Deflector  |   9379 |  2.1 | 2.09 [1.97, 2.26] |   1,997 / 33,705 |      25 | measured |
| Satyr's Lash               |  17752 |    2 | 1.99 [1.81, 2.14] |   1,060 / 18,825 |      25 | measured |
| Silent Fang                |  13953 | 1.45 | 1.47 [1.32, 1.63] |     439 / 11,187 |       8 | observed |
| Strike of the Hydra        |   6909 | 1.95 | 1.97 [1.90, 2.04] |   3,699 / 34,120 |      86 | measured |
| Tainted Pierce             |   8225 |    2 | 2.01 [1.75, 2.20] |     841 / 13,218 |      15 | observed |
| Taran Icebreaker           |   2915 | 0.95 | 0.97 [0.92, 1.03] |     685 / 11,141 |       8 | observed |
| Thornspike                 |   6681 |  3.2 | 3.22 [3.14, 3.32] | 17,449 / 191,171 |     233 | measured |
| Wyvern Tailspike           |   5752 |    2 | 2.01 [1.82, 2.19] |     607 / 10,078 |      14 | measured |

## Angerforge's Battle Axe (item 11816, spell 1300783): 2, observed

697 procs on 8,001 landed hits, 10 wearers, 10 runs, in the main hand: 2.01 [1.91, 2.12]. Observed: every wearer is a
shaman. Left out: 3 paladin wearers, who read 2.24. The sim had no rate. Reports from 2026-11-10 to 2026-11-18.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |      10 |   697 | 8,001 | 2.01 |

Reports and fights: 1GBhzTR8CZwcJfFV 20-32, 35, 37-38, 40-51, 53-63, 65-66, 68-82, 84-88, 90-99; 4rfgMCY9qjyvQhHD 1;
AfJ6DYbh9dmkgHZL 1-2; GWL9jZvR4rpaKftb 1-4, 6-8, 10, 12-13; cgmhtvNdTGCn7f6L 73-84; dnDg7j9kRw6WtbML 55-61, 63-69;
jKQDqkwmx46ZLBFg 17-24; jr7TAd93qLPxWzhp 1-2; zmDY4LVX7N8vh9aR 1-3; znBQHJTxXapkRLt3 1-11.

## Archeus (item 2000, spell 18091): 0.95, observed

224 procs on 4,273 landed hits, 26 wearers, 29 runs, in the main hand: 0.95 [0.83, 1.12]. Observed: its interval reaches
further than 10% from the read. Left out: 3 paladin wearers, who read 1.04; 2 runs logged in a second report. Its
override row had 1, a vanilla carry-over. Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   175 | 3,345 | 0.95 |
| Hunter  |      21 |    49 |   928 | 0.96 |

Reports and fights: 1NRQ9tBnTrq24D3k 1-11; 2qk4MYw7cr8KGFtm 1-8; 6z9hqHARQr3vwLXd 1-2; 76ZV8FQ1RHDhbq3z 1-4, 6-9;
FRxmtWLHr3wvB8Q1 1-8; FfAC3tWBxbp2HV9j 1-3; Hgp61KVdvPqYyr2N 1-11; KBZyYLg1WTqm6Chw 1-14; MmqC2hGvFY7x6Dgd 1-3;
NFD1RP4LdQrhfw3W 12-17; TDjrhyCkw3P9YApc 1-9; W6nhHydV1wpMmZ7a 4-5; XvWP1x3FkAzwLHZC 1-9; aQGPVW4TBcv8y3wD 1-2, 4, 6;
b186j7df4hKxNPkV 1-4; bPqj8zkJN3YatpwD 1-2; cvfBaDxQX7K2zFtk 1-7; gr9zjCAX3mY41aqf 8-12; mtBc1WZf3ydHVPFv 1-6;
pdqmMHchNGQjTwgk 1-3; qpXkvRdWxKzfNr2G 1; rNXvGcnZabhLJd9A 1-2; rQdnmBHXCpKgJqw4 1; vN8WfZatBGgdCDHJ 1-4, 6-7;
vRPQ237fWm1tFYDV 1-4; vTX3cKpNbY6jJVDa 1; wMRNLvP8CBrgd4x7 1-4, 6-13, 15-17, 19-26, 28-39; z1APqCWaMYvxbkd7 1-3;
z4WhK2CnLbaHgNVQ 1-6.

## Barbaric Crossbow (item 272999, spell 1291551): 3.3, measured

3,276 procs on 29,847 landed hits, 48 wearers, 60 runs, on the ranged weapon: 3.29 [3.18, 3.41]. EV-0136 gave 3.45.
Reports from 2026-11-07 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Hunter  |      39 | 3,270 | 29,752 | 3.30 |
| Warrior |       7 |     4 |     88 | 1.36 |
| Rogue   |       2 |     2 |      7 | 8.57 |

Reports and fights: 1AfDWYQT7Xjd24K3 13-14, 16; 1Gg92cW6XBhkwYb4 1-5; 1RrDAhp7ZvVYc9P3 1-6; 1W9N26dHfnPV4G7A 1, 3-9;
1y6GV7KLTZPrJH2w 1-3; 2GbXqD7KcWTfRnYa 1-2; 2MHXV8Y9jx43vhpk 3-12, 14-18; 2RygmFXpHCbnazVA 1-5; 2tJkzh16X38wrQfP 1-3;
3pTj2FQg6VvJz4DP 4, 6-7; 4ChPnBrVK2fJLMaX 4; 72XWYanybGdwPg41 1; 7VkrJpR8XyFtL6Hm 1-2; 8Xxb9LDGHP7vcJfh 1-5;
92JNtax7RTkQ6cHv 1; CVFcbDGj6n8WMKyL 1-2; DACv8aKzbkxfGygr 1, 3-4; DZvkgfwrVLbpmWyQ 1-2; DaNQ6WPf7GqTpKVk 1-8;
Ghp9gxWFkBNacLfn 1-5; GkLDK8WpcdvPatfB 1-2; H9QxwyGt4z8MmqCj 1; HwqhFkWZCBRTQVDf 5-7, 9-13; J4mG8yjgVCALMfRw 1-2;
KBzML9kdHFCXVpQc 1-2; KYD67nVqZT4jxWyN 5-6; L7QHtkhacypgd3D1 1-2; LnPRx96hKZN3CgTF 1; NQyGkApg7zMc18La 1-3;
NyG7zfKj2Mvk1FLJ 1-6; P7MBDvyCRQtLkJHz 1, 3-8; T6MXt73janNJgmyY 1-3; Td29Nzahj1cmGMWZ 1-2; TdXmWDR3QFJwpYfq 1;
VXcPMWHwZJftbCrm 1-9; YHh2t9agcvQjTAZq 2-3; Z9CQ6fqYjzcB4tG2 1-3; aMQ63JtDnpmFz9WR 1-5; cAdCm1NLGaQnXrHW 1-4;
dc3Y84hWFnrA2wqp 1-3; dtLy89QkrcHx3pMf 1-2; fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; gBxakRwhQDvnLXp1 1-2;
hMdFnHqgVRDTyKPY 1; hRzbpMrDY3cgQX7n 1-7; jRytczTZF6pBJgYL 1-2; kBHbV3DZjamvfJY4 1-3; nALy2WrwHC9qdMFm 1-2;
p9HgV7zdCKqbrwJ3 1-3; ph6FqnCZaRQLvxbM 1-2; pw8aYCczKgjMbvAd 1-4; r2BYfmZW1wkLDdNx 1-2; tbWaYgLTw8XQAzdC 1-3;
v9fnrCd8TLaPDYH4 1-2; vqpcV7yDGCkbftmP 1; vrZD4X7Vb3pW9FNK 1, 3-4; vxkFd3MfgJX7NyAc 1, 3-8; wmT1zqYBat3fHDLy 1-5, 8-9;
zcL8MvJnPXxYhNyC 1.

## Blade of the Basilisk (item 8223, spell 10351): 1.1, observed

157 procs on 4,564 landed hits, 5 wearers, 5 runs, in the main hand: 1.09 [0.94, 1.24]. Observed: every wearer is a
warrior, and its interval reaches further than 10% from the read. Left out: 2 paladin wearers, who read 1.03. The sim
had no rate. Reports from 2026-11-16 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       5 |   157 | 4,564 | 1.09 |

Reports and fights: 7fgdkyv1njDKa8Gc 6; 96pBnGdAKJMaCb3N 1-12; hYHBQ9ZJX3vFAnkj 7-14; qa1tKFnvZ98hwj6x 1-6;
xH8TDPQKb6mYragf 1.

## Bleeding Crescent (item 6738, spell 16403): 1, observed

484 procs on 12,273 landed hits, 29 wearers, 30 runs, in the main hand: 0.99 [0.88, 1.11]. Observed: its interval
reaches further than 10% from the read. Left out: 3 paladin wearers, who read 0.98; 1 run logged in a second report. No
rate was recorded. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |      11 |   301 | 8,174 | 0.92 |
| Rogue   |       3 |   120 | 2,229 | 1.35 |
| Hunter  |      13 |    38 | 1,150 | 0.83 |
| Shaman  |       2 |    25 |   721 | 0.87 |

Reports and fights: 1WC3Bnw4A7RKpLDj 1-4; 1rdmQ9wptPWHfNG6 1-4, 6, 8-9, 11-23; 37gGDYxn4ZKLbMV1 1-7;
8KwxgLPtkpD7h2Ry 4-5, 8-32; 9CGYv1Jp24aWyhf8 1-2; 9mZgR7wqPxWMat6L 8-13; AwkvX3Tt1jqLhMVc 1, 3; B3vtVgmqMTHDAxNY 1-10;
CLbmK261T3d9gAR4 1-6; CNWJgkfyAhjB3vRt 1-11; FbNk9QwxY7nG618P 1-19; NgAWV9MGwvbFLBC1 1; NwKVWqHfgk1jLvAG 1-5;
P8Znjv1fgBD2CAVY 6-12; QDVfcwHpjZRzL94q 1-6, 8-13; RPnVYQcG9xTJh1ar 1-4; WnvfNJK8dQG9xPMY 1-2; XgZdn3D9KR8cfLNY 1-10;
YhCcTtxfGJR82M7Q 1-11; ZAWJk6BdKhr9nC8T 1-6; f2zK39XdhnNqZJgR 1-2, 5; fjb2rRngTPa48WA9 1-4; h92ZHrfm3RFQtYvT 1-8;
jC4HdxZGVNLz1Yvm 1-2, 4-6; jGrv6B1hNMDJPc3W 7-10; kmBptFTG68Dcq3jx 9-10; mQ4MqkD1nbaBXLdy 2, 4-6;
npbfVjrQdDLhtRWY 20-22; pKn8wBZcQ62PyJH7 12-24; tD7Rm3GBjdqPzhwC 16-17.

## Bloodrazor (item 809, spell 17504): 0.9, observed

1,310 procs on 31,842 landed hits, 15 wearers, 20 runs, in the main hand: 0.91 [0.81, 1.02]. Observed: its interval
reaches further than 10% from the read. EV-0136 gave 1.05. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       8 |   777 | 19,621 | 0.88 |
| Rogue   |       7 |   533 | 12,222 | 0.97 |

Reports and fights: 3XCdM1LQRNmvY8kp 1-8, 10-12; 76zv93gbJQFrWKfd 1-30; 8GFRaPBnKWf7Qrpb 1-3, 5-6; AZchwX93DmyHLbv6 1-3;
NWvgFp3kfR2bHyYa 1-35; NjgtYv4VMR26q7h9 1-4; RW1BNk6Ybyh3XjLQ 1-21; WAyrNRJdm3pqCLMK 1-13; ZLpcV6krn1wYHJvt 1-8;
aNh3Dt8kbZHL9cBM 1; j1xyYV7gcMH69ND8 1-20; jJWt7Dnm4qc9rzQd 1; jr7TAd93qLPxWzhp 1-2; kJ3WdhMLqv7K2Nyf 1-3;
nzxZ4h87yGXJPgfp 1; p34dnvkg7yDKL6cr 1-3; qVnZ6bPMhWRJA397 1-3; tk3BD1bZafMgKxWX 1-4; vBWX7HQn9zpTMGd8 13-22;
vyQrDZN7RX82ApnP 1-3.

## Bow of Searing Arrows (item 2825, spell 29638): 5.1, measured

2,724 procs on 11,860 landed hits, 17 wearers, 18 runs, on the ranged weapon: 5.10 [4.86, 5.36]. Left out: 1 run logged
in a second report. EV-0136 gave 5.4. Reports from 2026-11-13 to 2026-11-18.

| Class  | Wearers | Procs |   Hits |   PPM |
| ------ | ------: | ----: | -----: | ----: |
| Hunter |      16 | 2,723 | 11,859 |  5.10 |
| Rogue  |       1 |     1 |      1 | 22.22 |

Reports and fights: 1XGPdhkb28MAagN4 1-7, 9-10; 4CBckLvG7RaxFzYA 1-4; 6Krp8bkfx4cRjBaz 1, 3-15; 6WG74An9RrCdqkwD 3-8;
7BjkRJr2DZp1g8TL 1-4; AhLc6aGKMCdvBgDW 1-5; ApJhHfRkMDqFj3rb 4-5; H2GvZtVYNy4QzqLw 12-21; KGQHdbFCZpt1YX23 18, 20-23;
Pf4Y2xdhVz1vq9rn 1-12; XqBgvYQZFw4H2czt 1-2; YWGNyaDcLwT1QR6J 1-8; cJhVzAdLx3BHqpkT 1-4; fHbJKd4LnVXDW2gj 1-2;
krpvb6WaLmFnV1KG 1-2; mtcXBg7ZkVxFHQnh 1-4; qHgyawJNVb7126xL 1; tk86znZXxMCdwBKb 2-18.

## Burning War Axe (item 2299, spell 18199): 0.95, measured

654 procs on 12,350 landed hits, 24 wearers, 24 runs, in the main hand: 0.96 [0.91, 1.02]. Left out: 6 paladin wearers,
who read 1.00. EV-0136 gave 1.05. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      19 |   543 | 10,263 | 0.96 |
| Warrior |       4 |   111 |  2,081 | 0.97 |
| Hunter  |       1 |     0 |      6 | 0.00 |

Reports and fights: 3cKTmVyb1ZPwFXkf 1, 3-13; 93NcnZPAhCfQ4wHz 1-4; BqjTQP3Ghbt9JyYn 1-11; FMcvKqTzn7m8Va3P 1-15;
Fgy2bt81NL9CAzaR 1-13, 15; GJHgk6MDpxfVdacv 1-23, 25-29, 32-33; KRCBMJvg72fHVwWL 1-17; LwTzGJXrN3HWqCfD 1-5;
PZA4xq2WvQJ3amjH 1-5; Pc1RkCWp8wTNfFr2 3-4; QvKPgyHDkBMbcp7T 1-10, 12-21; RBmtAMLxz78Pj2Wh 7-10; VQFaRpT6J2xLDbZg 1,
3-5; Wf8ghC6TtQKrkXdv 8-10, 12-15; bkMdQTwZmqNPvhrR 1-5; gFZ7BcKt9JhaY2dQ 1, 3-9; gGtkavy2n8bqADQ9 1-2;
gJkrfBtbp7YM3PCd 6, 8-21; h2FAdwQfcJLHpzR3 1-16; j68nZp1gcrGChaRW 1-12; qMD3CVapwb6nAgx2 16-40, 42-43;
rMyzaJhG4k8wHPj6 1-10; ryqgQphLxMBz4TmH 1-4; t6HbjRZmkyx3pQwK 1-14.

## Cursed Felblade (item 14145, spell 18381): 2, measured

7,759 procs on 100,366 landed hits, 108 wearers, 123 runs, in the main hand: 2.02 [1.97, 2.06]. Left out: 35 paladin
wearers, who read 2.07; 6 runs logged in a second report. EV-0136 gave 2.55. Reports from 2026-11-05 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      65 | 5,047 | 65,401 | 2.01 |
| Rogue   |      35 | 2,662 | 34,365 | 2.02 |
| Hunter  |       7 |    50 |    599 | 2.18 |
| Mage    |       1 |     0 |      1 | 0.00 |

Reports and fights: 12VhLkPa98DYFH4g 1-3; 1RrDAhp7ZvVYc9P3 1-6; 1cf7ba34WkwGyzQY 1-12, 14-16; 3CyQk69LzJjTn2X4 1-5;
3rDXa6gPxYNZJjFH 2; 6P4VND3w2r1LGQab 1-2; 6jfqGYcR9AWdJvFh 1-3; 6mrVNdqBkwgvGxW3 4-5; 6tANqvfKQP1Fkdzg 1-3;
7M48dqLp2yYHaA6Z 1; 7NqcJQvyKpCd3YAt 4-7; 7RyBGr4dYVFqJHDb 4; 7Yt1bzjCBg9JycdT 1-3; 7bch4MKwzf9HdT3v 1-3;
7qLWf2j9a8tvCNwH 1-7; 8NTz6vHAcj1W2ZmQ 1-7; 8Xxb9LDGHP7vcJfh 1-5; 8aymDGwFnWZAx7YR 1-3; 8tmqwrpyARjY6hnc 23-34;
9NbWjdLhagPZn1X7 1-2; 9aCJ7GH6QRFjgqb4 1-11; 9mKt61MCAyfP3GRN 1-4; ABX72zJxvr6YVpk1 4-5; APNwHKda8WZQRt1G 1-2;
AT1kyCR8NhMpcbjm 1-3; AdNRV84rQFWcxk6q 1-2; AtH8ZJm36y74V1d2 1-6; BwtcDCWfMLdTYFqN 1-7; CyJ13pMra2W6vdQH 1-2;
D8qzgc9h1YLf6krp 1-6; DaNQ6WPf7GqTpKVk 1-8; DmFcCT4QpH3AbhPR 1-2; G6M9XwZ1gQCTFbPa 1-2; GbrjBnmfgvphqcPZ 1-11;
GgTjx7XKkMvtwnQ3 1-5; Gv74gVkXcyPHaYzj 1-12; GzbK6XYDf1vQH4W9 1; HwqhFkWZCBRTQVDf 5-7, 9-13; HxVGdyMjDB6JCPKQ 1-2, 4-6;
KAvPryL1zCdtWTxV 2, 4; KYRtHcN6L7nmq3Cz 1-2; KgkThB9FPJv3MrZw 5-10; KpTQqRdvf71n2Wm4 12-21; LcaMApJGDWjg4Tbf 1-10;
LwYXzRFCP8GvVKfA 6; MLTmgKZDkJaAcwYR 1-3; Mt8kPGravCBTjdA7 1, 3; MzwcDrG3fdFBRy6a 1-5; N4tMyW9qJcBKhvY8 1-3;
NZrPmMxy3LXtj9G8 1-3; Nkndb4wD6WgAK2Lh 3; PW7k29jyJ3C4chqQ 5-7; PtzakwgYVNR783An 1-2; QDGMAJr2K3Lb91kX 1-7;
Qmw4FNtdvqnkPcKy 1; RF61yp4dkXNCZTwh 1-2; T7HKGnYfbMLwWCJp 1-7; TMj2XLPvzgq3ZNJ9 1-5; TbPzqABaXyWmLxD9 1-8;
TdXmWDR3QFJwpYfq 1; VKxkQv9dyNAgpjPh 1-3; VXcPMWHwZJftbCrm 1-9; VYHAwTB2pP1r7mgJ 1-5; W416FHmftD9rPRgA 1-2;
Wk1G3TqNcZzPvRHD 1-8; XBQMrcVdkvqGCJF1 1-8, 10-11; XKc78ka3dWNgGRDr 1, 3; Y6p9C7nNx1KBWGdM 1-13, 15-33;
YaBQZ2tbVFJMrAhk 1-4; YbxV4LgjKt2BfTCp 1, 3-4; ZrM8cydqRt1BpLfx 4; ZwHMphjVvdtxTY9C 1-4; a9fR1QjqFvnhgxW7 1-5;
aDRLCWyFkdqVc2Z6 1-2; aVyXCj1NpnRTFKxt 1-2; ad24TYKCAb8LwBxc 1-4; agA7pcDMCb9qJFKv 1; bXYpRJw87M62gKar 1-4;
cAdCm1NLGaQnXrHW 1-4; dwq8K9CRNBAQtLjp 1; dzTmC6F7xwLPJrq9 1-17; fTpYLw71QVR2qky8 2-13; fkGrDPXmVWazJxpH 1-7;
fxbtCJFLjv6WMBcy 1-2; gBvQqcCZaW9TPLnV 1-2; hYDQPX1rjn7wZLxz 1-3; jGkWVBDa7FPmpwn4 2-9; jKfkzadV3BLp1cbr 3;
jRytczTZF6pBJgYL 1-2; k9wzA3ZbgnX7j6x2 1-4; kB7JaQmdX8RZjcnT 1, 3; kHnFMcBXRrx9VajJ 1-3; kcwp7a2jqCF91ZWK 1-4;
kgFaYjKwPH1CAx2m 1-2; mzTZRY4Hg9aGdvxP 1-13; nPZbN3Ymk1LVtDQK 1-2; nZ1gTJbNYPaLkx92 1; nfCzby7RrmYFxTXc 1-3;
pL3Fbdfc2vNKCxtJ 1-4; ph6FqnCZaRQLvxbM 1-2; ptWXb1v93rkZR462 1-4, 6-9; q6fAZYgapznjxPK9 4; qcyGBRWh2xVZgKAD 1-2;
qhFK6CrAVTH932kt 9, 11-19; rzchPRZJy4qNp3mW 1-3; txhaZrbjXR7Kf2TD 1, 3; v2y9BTtzLhYr1XHg 4; v9fnrCd8TLaPDYH4 1-2;
vBCYwTV1D9gxHPMQ 1-9; vba4w71K9dVTXZ3r 1-2; vrZD4X7Vb3pW9FNK 1, 3-4; vxkFd3MfgJX7NyAc 1, 3-8; xCQDhmMVT4qdRWYb 1-13;
xPmgn423azWr8F9L 9-11; zCjDnVmyRZPg2J7Q 1-3; zLK3mTNayZMgR6vh 1-7; zPJ2w1BrKA3Dv968 1-2; zRd3nAXrVQqg9ZH6 1, 3-5;
zbV7pwYGARa4Tqgk 1-2; zhj69Dt2X1b3fvMT 1; zk682jCMJ4AdKfV9 1-2.

## Darkwater Talwar (item 11121, spell 16408): 2.9, measured

12,319 procs on 116,459 landed hits, 132 wearers, 147 runs, in the main hand: 2.88 [2.81, 2.95]. Left out: 25 paladin
wearers, who read 2.91; 6 runs logged in a second report; 1 run with the item swapped away throughout. EV-0136 gave 3.3.
Reports from 2026-11-07 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      63 | 7,471 | 70,723 | 2.88 |
| Rogue   |      44 | 4,538 | 42,908 | 2.88 |
| Hunter  |      24 |   310 |  2,827 | 2.99 |
| Warlock |       1 |     0 |      0 | 0.00 |

Reports and fights: 1CYTfP3pFDcaRrJm 1-8; 1h9vznDdxyjC7KWZ 1; 1vJgfm9dGnkDzwXb 1-2; 1w4hbgxXHpf3N6M9 2-3;
24KTy3bLGNQm1rdx 1-20; 24TVGCfvHpjtYzrD 1-9; 28GbNhyfwQ1DcRaP 8; 2GYNtHzfKmWpTPVZ 1-14; 2Z3YxGdfPhMK97kQ 1-4;
2pGYFgHwLDC17zZ9 1-4; 4CRDaYFWAqt1JnmV 2; 4LrMDtG6T8AVRJzQ 1-11, 14-28; 4ktfmxnqcbv61B2L 1-18; 4wJgRGx3rCKjQZmD 1-12;
6FZtkRcQHAKxNfrC 1-19; 6X1L4nB7zCfqk8vV 1-2; 6Zzwx92Wh4qFVLBp 1-8; 6prtaJW4g7ZTVQRK 1-6; 6wXA2CY3Bdcp7gDj 1-4;
6xhBGvgcM9H48tyK 1-8; 7BKLx2tGCwqQHr63 8-13; 7J9jfrb2LaHxgv6w 1-13; 7NHAcVG9nZCyBMjv 1-6; 7xY1zPCyKnk8DGQj 1, 3;
84c12nJmjaKCFr7H 1-3; 8BG41brwTQkLRjqJ 1; 8Bgm9rLTF67QvNwn 8-11; 8tmqwrpyARjY6hnc 1-10, 12-21, 23-34;
9GMN2DQLhrB64Atv 1-2; 9yNQP3FWHgcVvYDq 1-14; A1WnB2b3jvaJKyfp 1-3; AL1DxTaHGfVygm3p 1-15; ARwG2htz3VgjDd7v 1, 3-5;
Ahv9342TJmG7bCzr 13-15, 17; B3vtVgmqMTHDAxNY 1-10; BmVrHWZvtj9GJTxD 1-13; C8J2PmQbz9dwcKXH 1-4; C8QWj29pvm46kMTF 1-3;
D3Kbtnm2yaC4Tqwc 1-2; DK72jqp46VtnYHfZ 1; FR9jJ7b31XCkLmtQ 2, 4-7; Fw6bq2BvXnJmaRk9 1-3; GDWLqtcJgbrz1P89 1-8;
GntwzVb68vMyWjBd 1-2; GxWL9RvzyJC76d13 1-2; H1MQjnkcT6GraYbh 1-3; H3bRaDmzyjxTW2BJ 1; H3vdcCr6JX1gwYtq 1-3;
HG4YQy7Jf3xnFpLM 1, 3-9; HPxv4aJ1WgcTzMCd 1-13; HhFaG7mzPvZWpTtK 1-2; JzRmL9FYywDQgNcW 1-2; KqZaH876NnQV2CJw 1, 3-21,
23; Kz8FhGRY6kVTBjwb 5; KzctnTrXN64fmWPv 1-2; LHD7ZAYVNm2zjdXP 1-17; LHRhtADNW8GfyYqT 1; LK3jvm28bhWQadyr 1-4;
LhMNaKbHFvTt1Ax7 1; LzFXxt3Y1HyVK4nc 1-7; NDf1mPJRCaTM8WkK 2; NfLb7C1qDZ8cHxFp 19-39; NfLb7C1qDZ8cHxFp 32-39;
NfjK7rVnPvkCB3XD 4-9; NxmwqXVA7JkpP8Qn 1-4; PrNcBn64hjYw7KQ2 2-8; QGLzd2PDcm763VRr 2-3; QRnGzYBNkWyCD96g 1-7;
QVg7ANTY4B8RhDJM 2-9; QgR6kbHvY1CBhXFj 1-2; R2zHVBktcw8j6rhM 1-3, 5-7; R7Fw1Kdcab3VYTx6 1, 3-9; RgjLQAhp3HTCwkca 1-15;
T2zBbgrnJdK7QNPc 1-8; T87bpgKckHzNDAfY 1, 3-6; TQ4PXCqWA3KGYzh2 1-5; V64yC1A3DZdWfGBP 7-25; VWF2YryN8j4qdK9H 1-51;
VaFNJB32Ah1WGLrn 1-11; VdwjgLC7qNzvPFpT 1-18; Vr12d8kxGHpKazBX 1, 4; XAyLYahcRwMWQfF7 1-4; XAyLYahcRwMWQfF7 4;
XWnJCQ1Mc38fRtwF 1-3; XqhkLtanYFGgx9jP 1-14; Y6wnvNGfpPVrR7XJ 3; YC214LNaFPwkRrfm 10-12; Z19fCFxJnyLc6dPN 1-20;
ZjgDb23AFJQkvpWL 1-3; Zvbh2XrFYqQPzDnJ 1-5, 7-20, 22-27; aKwXBJkG2bW6tVNY 1-5; aTHRgr6nQqVF7DC9 1-9, 11-20, 23-25;
azpj2hg1DVwnMCkZ 19-21; bH6GZhykAPqgQ3mt 1-4; c7gHWKFrZMnm2yb4 1; cfDYLdNC1t6kHQg7 1-5; dkAMWqBczZtbHTaV 1-2;
drvpchNFDMPXB8t2 1-10; f297FtyjGHmRVqYw 6-9; fLa4hw8d6pNFYMng 1; fXk6pdvPNF9wZCjc 4-5; g3nkvH9atfKRdVAh 1-11;
gAmJdBx7QyTcR3YK 1-17; ghCVqbxDmnJH6P9w 1-18; hfJZb8HL41KmtdMA 1-4; hvrBLpKmakJ8nG3W 1-19; jFDWcdYtPzkRhqB8 1, 3-12;
kFxDzj7pGtTqCZg8 1-3; kHfNcJTYj3RtMFGL 1-3; khV3XMmJ4aRKzxBt 1-3; kp4FHxfa9CL7WArB 15-30; mRxb9L2wNkMJTWXY 1-7;
mj379AhvbKZVpWHC 1-7; mkGcRA9M37QyJfL6 34-55; n8ZQCDKLjHP9rdkR 1-2; nKCDhXNbHpJaBcAR 1-15; nmFLdPWf8agkC7y2 1-3;
pHZ86RKXPNY4nF1V 1-4; pKRWfNGZg96VBQTx 1-20; pwNAZn643XgzcbYa 1-3, 5-6, 8-10; qavTP8NJtW4hcQLH 5; qhGZDM3ygKAHtjbN 1-12;
qjryBx72WJPmA4hT 22; qpXkvRdWxKzfNr2G 1; rTzVg1BxmZh6pFKC 1-11; ryL9Dfkm6jtzFXgG 1-17; tD3pKRxTyNrqGchC 1-14, 16-18;
tbT31mz6ahFMfnCW 1-2; tdxJaDzc8Qv2jFbq 1; tjN8BgxX7k4GyAfd 1-4; tvrgBcqMzCRk1wF8 1-2; vk2wGPDbQp8TZBR4 1;
vtHaBZF2VAqJ19xY 15-16; wFn3hXfqxRHpQcbB 1-4; wQAj94dkrc3MvXpD 1-6; wYmr1kfWPpVz7jyZ 6-13; wqCht6J1RkcKfVpy 1-4;
wxdb7vaQp4CATFyj 1-7, 9; xQtGrv7cg6V8Z4BT 19-21; yX2kx6TbwHn4mcBq 1-5; yqntQdrPj7pvw8Fm 1-3; zcK9qfp32VnQWTMP 1-2;
zhZcTYBNvwr48gP6 1-6; zq6DmwtPJkNZy9KL 1-9.

## Dazzling Longsword (item 869, spell 13752): 2.6, measured

4,335 procs on 38,485 landed hits, 37 wearers, 37 runs, in the main hand: 2.60 [2.43, 2.75]. Left out: 6 paladin
wearers, who read 2.76. EV-0136 gave 2.8. Reports from 2026-11-09 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      25 | 3,104 | 25,852 | 2.77 |
| Warrior |      11 | 1,092 | 11,383 | 2.21 |
| Hunter  |       1 |   139 |  1,249 | 2.57 |

Reports and fights: 1FV3WyfXGCcPrT9b 1, 3-8; 4C28YPXqgTMQJxB7 5-9; 4qZXFGfc1wCMQKL7 1-8; 8F4MhkmNBxRrf6Zw 55-68, 71-83,
85-89, 92, 94-105, 108-123, 125-141, 143-151; 8X2Kc46faTnxVBLm 1-3; 9cajk6Jqd3NTQngf 6-8; A1FKkYpPxrX8tbNv 1-2, 4-5;
BCZgcRxrh7dA4ayX 1-4; BgKt3Qm2fJDAWLzh 1-23; CrR9c6VtXG4bzJhx 1-17; D4A3QNCrX19W7cqJ 1-10; FR9jJ7b31XCkLmtQ 5-7;
GzZH1WhTLQ3JaDXt 1-11; JTztbhZyaWHC6kxf 1-5; KVdX1Wn4GpzAH9MN 11-12; Lg82ZFDyQ4MvRm9p 1-2, 5-11, 13-26;
M4CvXLZz3qm6pDRc 1-3, 5; PKRTfgw2Q7ZBxt4r 1-26; Q7BxMTcG1ZbgX2v8 1-2; QftYMR2haLcXZ7Nz 1; T71XnfaF3cY9M2Rg 1-13;
Thf9agqHJyRmtdpD 1-5; XCJdxAF17zqHMhLY 1-2; ZtndFBMraKvcLfw9 7-9; dmKgVGCJMQpq3t17 1-14; dzXT7JngpxRPwG32 1-17;
fCrvDxhZXqJAg62m 4-10; fkVmJwWZjMy3LBX4 1-3; gTPNZcJ9mG4dRC2X 1-4; h6wkBb9Z7jFtTrnp 1; nF7RcakdVTQD2tvH 1-3;
nLk2TKDJWP6vBgam 1; pKj3bDrf1h2qtNMF 9-15; qW3KvYDpw6ARxZbc 1-7; tkQKL9bGT6JVR3na 14-16; xM9bRGyajmBFvCqt 1;
xqy2GfPj67XCwrT3 3-4.

## Diabolic Skiver (item 9475, spell 18206): 1.75, measured

333 procs on 3,886 landed hits, 13 wearers, 14 runs, in the main hand: 1.77 [1.66, 1.93]. Left out: 7 paladin wearers,
who read 2.20; 28 runs in Cat or Bear Form; 1 run logged in a second report. The sim had no rate. Reports from
2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Hunter  |      11 |   186 | 2,104 | 1.83 |
| Warrior |       2 |   147 | 1,782 | 1.71 |

Reports and fights: 2zXhygVYpkrMZGxT 12-15; 7gWZKLnBmty1cT3J 1; 94fVcDCNBLztXh1k 1-12; HGn9XrYNPvwKtC1m 1-5;
NR1k2zLFqKt7GXAD 6; RXDCK9PtZy3Jp2QA 1-3; X47TDyBdAh2nQ6LK 1-10; YHXAcJxWG89pMFgf 1-9; bhyYq9HXd4xvAwtJ 1-13, 15;
jZtdhVaHckGN72L9 1-11; rDJL7nqX1VZHm9zk 1-8; tk86znZXxMCdwBKb 18; xwrkGVNA2Z7Hq4fW 1; zTbHr3BPvKxRCMZg 9.

## Electrocutioner Leg (item 9446, spell 1292851): 3.9, measured

32,143 procs on 290,507 landed hits, 202 wearers, 243 runs, in the main hand: 3.90 [3.86, 3.95]. Left out: 69 paladin
wearers, who read 3.97; 4 runs logged in a second report. EV-0136 gave 4.25. Reports from 2026-11-07 to 2026-11-18.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Warrior |     131 | 23,526 | 214,204 | 3.88 |
| Rogue   |      55 |  8,229 |  72,603 | 4.00 |
| Hunter  |      16 |    388 |   3,700 | 3.70 |

Reports and fights: 19DdXA3BwjZHFbVJ 1; 249KNqjY61fcwdMn 1-9; 27dNt8HLBCPpAmYZ 1-31; 2N4yhcrz8GW7mRnt 1-6;
2jYyAK1RrPvTC8WM 1-5; 2ywhMnJq4Y3zaWtg 1-9; 3HRvtBnAbfM9xTkK 1-3; 3HZNbTwqdtnLFg4K 1-4; 3wCRWQcyNrgKvHz9 1-9;
46bPRkzVHgafvMwA 1-12; 48jfdvMD9HQbP1tC 1; 48pmMdNLnqxWj2Fz 1-12; 4C28YPXqgTMQJxB7 1, 3-9; 4M8gQTJhZvRPWf2B 1-5;
4Ppkndm9fWvTh1Gr 1-5, 7-12; 4Wq6Kcgnz3r29fTh 1, 3-5; 4ZNwHaLtygF1k3xC 1-3; 4jCpGrhx3NvnDmZ1 8-11; 4nbDkxaLmgcwGYQv 1-9;
4ynKGp8XCQPNmJLF 1-4; 629cLJxdjGkVBDTa 1; 69atW41MxnTY2cXN 1-26, 28-37; 6Pwh4WBYyAbMGD3c 3-6; 6QvVFqNYty8bzrcf 11,
13-21; 6WkrKt3RwHGmb4Zf 1-3, 5-7, 9-15; 6aBqgmpQhVAvn94Y 1-5, 7-9; 6jVXd48tLqWgQvK3 3-10; 6z9hqHARQr3vwLXd 3-4;
7f6DBRAPMLGYzyQm 1-2; 7kVKd6JMNTCj3P4z 1-10, 12-16; 7kVKd6JMNTCj3P4z 7-10, 12-16; 7n6rKTfk1vhzgXyF 1-2;
7tYjam2PXvfcGz61 1-14; 7vxRAKHWC83tgPDz 1, 3-7; 7x3ycAVKT9Ba4dMZ 1, 3-7; 8VQDf7JtwbTknMLz 1-3, 5;
8dT1VNMQfLgrj9vc 15-17; 8rCvtag1mFYqBHM3 1-6, 8; 94FmbDrtqgPRKyH1 1-5, 7-13, 15-36, 38-57; 97CVmB4zqyrcWpjT 1-7;
9T6Zbt2QxPXhmJpk 1, 3-27; 9bd6VQvJR2ZqGW3k 1-6; 9g7NQBqDH6FGfP8L 1-4; 9v23hzNwDpWM67Xy 1-3, 5-18, 21-22;
AGjFnqDY42gfQBTx 1-2; ARqz9awMbfr6QD7Y 1-2; AhxGmJv3nYDzTW9q 1-13; AkBH7pF6XvZc3Gdw 6-8; AxCwvBkjqYVtLgnc 1-13;
Azq6hMZn2RyWwagv 1-6; B6jfw4mcnPxWCKYD 2-18; B7wf2rAVZJLQgT3m 1-5; B9RcfwgQpTHL8m31 1; BTAM7pf2VtYCnJ8Z 1-16;
BydpaqJVW29G8QAc 1-9; C4TgLGnkKZcPqF8W 1-12; CBK8LRX4tZzWANqF 11-13; CDZG21zLfmMdh6FN 1-5, 7, 9-10, 12-15, 17-21;
CLXHZ3c2GyN6mYBD 1-4; CcAYgrB3Q1w8KJvV 1; CfFMHPahckGpR8Wd 1-3, 5; Ct8XcxY26Mf34NjR 1-9; DTrLRQbfXny12B6t 1, 3-8, 10-11;
DVdNpYmZfz9MGh1X 5-17; DcyRfjPp3YzHMkra 1-18, 20-22; DqHT31fVnh4rjaky 3-7; DzWf9A8B1MNnrm2L 2-5; F9DyXdCazKmPqWfN 3-5;
FBQzxArfjdCZp6Lg 2-11, 13-18; FarzJZ1mH2ydBWxh 1-17; GCTbWRf9mrHqxAX6 1-3, 5; GZzaJH1QVjMKd8qA 21-34;
GxTn6w2dD9MmXypP 1-9; H7PBRD9CqnGxkF4t 1; HRJA3nGpVzDaXyYP 4; HZCD4zaBgtKh63rb 1-10; HmnhWYy8BtCGv6R7 1-3;
HqzVaYrZLQxAPKhF 4-10; J2XL6AyZqgpjx8QK 1-13; JBj14VkRgmcTqDFd 1-7; JHDxh1BWZFGTfQ2n 1-4; JWbdPL3FRphQcM6z 1-10;
JZRnFQDrmYqBycvG 1-7; JcanfA1v3rXN6bBW 18-22; Jcx8g4CXq1QBpzVP 1, 3-12; JzRmL9FYywDQgNcW 1-2; K3jWRx2qQ1aXy9PA 1, 3,
6-13; K47ntNhqVQz83H9g 1-2, 4-5, 7-9; KAfmrqFHXh7TzdL1 13-20; KFdN4BrZP9X6tbcL 1-9; KGLhazAPkWfd9rw3 1-6;
KTz6Gaym8fdCxhkb 3-6, 8-13, 15, 18-23, 25-26, 28-45; KZzq7NTcBXj4Ln9r 8; LBYKWV4xgfkTvHRt 1-6; LJ43xqm7ATnCH8MX 1-13;
LKkr1pxAtT7CvcZN 2-17; LMdRqJGhV17XwWyT 4-22; LQbJn6gcWz2D3HCT 1-2; LWtG286RrDcJZ1xw 1, 3-5, 7-11; LXNz7Ym8rvwqcHax 1-8;
LayrxDGdQb712wpC 1-10; LjFr1v7pJKVcfhWB 1-11, 13-14; LwTzGJXrN3HWqCfD 1-5; LzXRwTc3YCfFkG89 16-21; MmTXHCN4KDfpaFwz 1,
3-12; Mr7nxYapkv693wRh 1-10; MvgK14QZDxjWGBXp 3-14; N9qCv8pYmPQRb7a3 1-18; NKnQBL2MrwCyAHzd 1-6; NTcKDtjPwZk8QaXg 1;
NXAxdr9JVKLzFDWg 1-4; NXgv93kr2G7zyTxm 1, 3-9; NcRqzah1TQjMr3nF 19-25; NgVPn7M3yX6AB8rT 1, 3-8; Nh9wMxkpzqJTPA2f 1-8;
NmwLfZHTpdqPWAMC 1-5, 7; NnmBjCcpKrxkg9GW 1-4; Nr8n1KpjcYwk2CvH 1-5; P1QTDgLwfytJKrBh 1-8; P3JYyjrc1FWgfN6q 1-4;
PNnXcA9hFBx7Wtdj 7-11; PpLZ7FAgkvTwc1Cr 1-11; Q6hKWVt9H47vNwnb 1-8, 10-12, 14-16, 18-23; QP32XzqKRHTvZ6FN 1-4;
QVg7ANTY4B8RhDJM 2-9; QwL7NjYPXAmzakRp 1-9; RD6bnVNTgPhGYCvy 1-4; RLnWB4FrJgpMyNZj 1-5; RPMNvqn7tXwJjak9 8;
Rfwtj4ZF1nTx2mJp 1; T13Kx4nhGJW9vR8Y 1-16; TC29cvWmBMhfZJzV 1-2; TGmKaXrQPp7czjx1 1-4; TJ8fyKCp4BHtWdLm 1-2;
TdP4nFMvCf2mW6rB 1-18; TgVP6hxZKqY7MGQ4 1-5; TnbzrVaxtHcpMKLY 1-3; V3L7kJDzTGg4QqCF 12-18, 20-27; V4LaQWhtxHqXDpwT 2;
VQbrqK3D1RHXWpa4 2-20, 22-27; VaFNJB32Ah1WGLrn 11; VaYkQ91Tt74yKbrc 1-2, 4-8; Vm9wxAyCHQX7q8Bk 1, 3-4;
VxmDcHTK8LqgNCwQ 1-7; W2rLkBmgb864JNDh 1-6; WA9GBTPQRKvazHmy 5-9; WBZ7gMLqXFcQrzDJ 1-2, 4-20; WCQGyn1J9q4Kdzbk 1-5,
7-15; WjgGnJtpyKzVY3b1 1-21; WrHQ38tb4KTnAy7h 1-3, 5-11; X3RtbZzTdYAxfKV4 1-4; XhZAV37vMFbYWRt6 1, 3, 5;
XpnkZCf7j6qDcmy9 1-4; Y2n7kAxTCch1mgf3 1, 3-17; YZyAfq6XN9k4a8md 1-5; Ydx67HwgajJKpbNr 18, 20-22; Yq9kmzBHDP3jM2Z7 5-6;
YzMA1vd3BrqLWxhH 1-3; Z182HvNhFzdqgCrj 1-5; ZD7YLw2qy4mWRhJ1 1-4; ZKzP2YvapHfqM8WC 1-3; ZQ2NaCptbjmWKJ6P 1-18;
ZaPRprzfmqY4CtF6 3-5; ZrD7NdhjWX4zG2PL 1-5; aDXfbQzV1Bwr9n2x 7-8; aKwXBJkG2bW6tVNY 4-5; aMNVKZpg8tRkhBP3 1-17;
apqZQxfDMJB4LPTb 1-7; avnKJBPL6ANFTC3V 1-2; avrcxzCymh1DNqA8 1-19; bjcHnakF34w6YLvC 1-6; btxm3BCN4jX61VhY 1-7;
bvzqmdxkA4KVQaZh 3-10, 12; c7XrRMZWYmzwgfq1 1-3; ctRfwLFrZ8BpaTJY 1-7; d2wFYHqnNcJpA1gt 1-10; dZYgMabwJXnD1rfK 1-4;
fWRgvP17z8yrL36h 45-55, 57-62, 64-68, 70-73; g7b3Jf29CVhKtYMD 1-2; gdMNAjD2Kyrtqh97 1-8; gn4HAC1RqMfLt8wN 1-10, 12-21;
gnMWYZa9ArC8jbNc 5-12; h3bCaBdyYJMPzvRk 9-19; hAJa2PcBmdnNbGFq 1-16; hMaTJ3ZtDzqPWHfx 1-7; hZvn9rLzYyNTaVBR 1-5;
hatAfwb9QHydP62c 1-2; hf1pF42aPHMv9kLX 26-36; jDHpNnbqf68kZxA9 1-3; jLtGKrwPDBMmF3Xc 1-3, 5-11; jPh97VNpz3Ark2YQ 1-4;
jVRTr4dWa8xfZHgv 1-7; jgByJqcWN6HFdLY1 1-7; jz3qpYLQ4PwTyHtr 1-12; kB3gRfF4hYJ2nAQH 1-5; kX94Vmn8tpJzGTD2 1-11;
kmBptFTG68Dcq3jx 9-10; kr13fAvjzZXMdNDT 1-2; m2jhcywg86TDGMZn 1, 3-6; m3dN8Wkg4LyqXZD2 1-11; m9PX3r2NY8jLQ7Dw 1;
mN7ZTVPw8WdJrhA1 1, 3-5; mR9Ch8QVfX1pJkT4 1-2; mwAFfdjYy3chWMxb 12-15; ndHRLVB9Kaz7xpqT 1-2; nqmgcYrPz9B1fLDx 1-17;
p4MR7G2atyhQwV6k 1-6; p92NRBfL3QCYrAgy 10-11; pbhgCm6PjYwZ3fRd 1-3; pbvkXazLqcHFhgRf 4-7; pkJPxXb3aYLmG8qK 1-4;
pwNAZn643XgzcbYa 9-10; pwd6RvtM2NYqmfGT 1; pyY6Kg9bPRkQfMHL 1-22; pzx4LKwq6PB8FrWy 1-10, 15-17, 19-25, 27-32, 35-36,
38-41, 43-46, 48-49, 52, 54-63, 65-70, 72-80; q6VRad2JDnpxXYAL 1-13; q7xagw1ZkdvnLQmc 1-6; q8wrpJkfFWdVg6Tc 1-3;
qT76ZPycRHrL41p9 1-4; qbDZa2XcdrM1npAF 1-6; qhGZDM3ygKAHtjbN 1-12; r9FD3YQGHcqdjC7g 1; rBy3ctF2NpKPQGWb 1, 3;
tAFwQbTHpJMCxZXP 13; tCj8cKdQr47MN6qb 1-14; tc1qGLjpgVrJ8WdQ 4-5, 7-15; tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19;
v8VbYrxBCyMdJhN6 1-4; vkG2m6pNJ8PWrQh1 1-4; wV9frJcGT4nzm8vK 1, 3; wYmTW817qdpJHRcK 1-2; wrVYM4xbWLNaczhP 1-3;
xR6rvTQLBwZ9nkzy 1-12; xmvMVYpyRjX3n8d7 3; yFWHKnAfCQpjNk96 11-12; yqVzktK2B7Lgp9Yb 10; z9mYPDLpcM81JCFA 1-3;
ztamxJTvkp82GK6h 1-18.

## Fang of the Crystal Spider (item 13218, spell 17331): 0.95, observed

355 procs on 13,863 landed hits, 11 wearers, 12 runs, in the main hand or the off hand: 0.96 [0.86, 1.05]. Observed:
every wearer is a rogue, and its interval reaches further than 10% from the read. Left out: 1 run logged in a second
report. EV-0136 gave 0.9. Reports from 2026-11-14 to 2026-11-18.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |      11 |   355 | 13,863 | 0.96 |

Reports and fights: AaKzvcY1wmH9pBbq 1; D46Xm8ncR9WbTFrL 1-31; G7X1CkT8z6PrfjWM 695, 697, 700-703, 705, 707-708,
710-711, 714-716, 718, 720-724, 727-737, 739-741, 743-760, 762-764, 767-769, 771, 773-777, 780-790, 792-794, 796-813,
815-817, 820-822, 824, 826-830, 833-843, 845-847, 849-866, 868-871; KNTYZJ3Xg9d8HftG 1; QXFd74zDW8JZMmKx 12-13;
XPYV9J6DfrMKxGqh 5-6; XgQYbwGBZcHLKRxV 4; c8mjK4TdzV2M7XLF 1-5, 7-46; h3P9VjGMZCJWfTcm 7-10; kaH1G4ztmYpgMTJ6 1-2;
xXWkB3fqAjt9NJGH 5; y6D3jqgWt8TvpNYM 1.

## Felstriker (item 12590, spell 16551): 0.6, observed

1,159 procs on 66,083 landed hits, 19 wearers, 32 runs, in the main hand or the off hand: 0.62 [0.53, 0.68]. Observed:
its interval reaches further than 10% from the read. EV-0136 gave 0.5. Reports from 2026-11-15 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      10 |   624 | 37,650 | 0.58 |
| Warrior |       8 |   524 | 27,903 | 0.66 |
| Hunter  |       1 |    11 |    530 | 0.73 |

Reports and fights: 2jJ1L8FPHzTB4fKG 1-5, 7-12; 2wMznDRvmkpbtHxA 5-7; 2wPQmVqbdCcZW9gv 1-12, 14, 17-20, 22-42, 44-53,
55-64, 66-76, 78-88, 90-100, 102-113, 115, 118-121, 123-135, 137, 140-143, 145-158, 160-173, 175-188, 190-203, 205-218,
220-234, 237, 240-243, 245-259, 261, 263-278, 280, 282-297, 299, 301-307, 309-317, 319, 321-327, 329-337, 339, 341-347,
349-357, 359-361, 363, 366-369, 371-377, 379-387, 389-391, 393, 396-399, 401-407, 410-419, 421-423, 425-431, 434,
436-444, 446-448, 450-456, 459, 461-470, 472-474, 476-482, 485, 487-496, 498-500, 502-508, 511, 513-522, 524-527, 529,
532-535, 537-543, 546, 548, 550-558, 560-565, 567, 570-573, 575-581, 584, 586, 588-596, 598-603, 605, 608-611, 613-619,
622, 624, 626-629; 4TxMfmFpzBJ6aN1j 5-7; 7tbWfYqJFd4vzL3p 1-2; 8Gw7m39ZXFz1y2fv 1-32; 8cqCXZHgpw9Px2hK 1-28;
B6bvH1fQ4kAcyRzt 1-2; FhG3AxYL9ZNP4RvV 1-15; KkRW7Am81LaJqHYz 1-3; M1tPx9DBZGYgCa4f 1-3; NjWKPpHgymCT832A 1-19;
PKBd6Vz2Gkjwp9Zt 1-2; QNfaZ9MTnWpFrqb4 1-6, 9-17, 20-28, 31-40, 42, 45-48, 50, 52-53, 55-56, 59-61, 63, 65-72, 74,
77-80, 82, 84-85, 87-88, 91-93, 95, 97-105, 107, 110-113, 115, 117-118, 120-121, 124-126, 128, 130-140, 143-145, 147,
149-159, 162-164, 166, 168-178, 181-183, 185, 187-199; QkVRc329w8nLTzKp 1-3; TNW9rHBt3MD1X8Ag 1-2; YFtvD87kznr1CThg 6-7;
bP2CFzkZTLJ8q6Dn 1-8, 10-16; cgMGCdPJQ84wzaBV 12-13; dnDg7j9kRw6WtbML 61, 63-69; gYm2HAtf4bnC8QZT 6-22;
hZLykw9dRnYbtvKP 1-2; jXTVfy3ApWKPYwDG 1-22; jn3dwzWPZ6r2ky1K 1-3; mpgnDQbLWXY8jJP1 1; nX1qbatTzGZ3CAwB 4-7;
pRA73YvkarFVLNCZ 1-3; tkPZYqTRgCyWG3jM 22-23; vZqVmLYQa6jwP3A9 5-16, 18-21; yPVkKpqtfJM27WFB 1-7; yfKb1MthLFDc4ANG 1-7.

## Fist of the Damned (item 10804, spell 18084): 3.35, observed

3,018 procs on 28,663 landed hits, 20 wearers, 23 runs, in the main hand or the off hand: 3.33 [3.09, 3.69]. Observed:
its interval reaches further than 10% from the read. Left out: 16 paladin wearers, who read 2.89. EV-0136 gave 3.5.
Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      18 | 2,906 | 27,652 | 3.32 |
| Rogue   |       2 |   112 |  1,011 | 3.50 |

Reports and fights: 1GN2LMtwvY4JQHba 8-11; 498vhQAtCw2XgZJM 11, 13-15; 4vcn3WF1RxXf8VZt 6-17; 8xjZmDw2LKv9bckh 1-15,
17-23; 9MCqcwjzLk4JBARG 1-5, 7-14, 16-22; 9WGcCxKHQ4pkX6Rj 1-11; B1qcTGpnHCrQ64Xa 1-4; By41zJVmM3cdPQkG 1;
JMnDtLGxf4FqCV6w 1-3; KYmwz2q1tG8yAcdP 1-6; NcwFg2bZjkrnXfQV 3; RTfqrCdVvLGYzj1A 2-3; VjT3aMZghKCyzGrt 1-6;
WMBFLQ43XvqTK7rG 8; Wb8w1TtxA3kLF9M4 1-8; Z4w7HKFPvcrTWd3M 5; fNMJcqDHb9ZzPLnk 5-10; j8bQnLtWwRMBXkDz 1-4;
mbZcgPavhdG7fKDW 1-20, 22-24; nmwM719fpABxvjdz 5-6; qprynNbXtRZV4DHP 1-7; y6D3jqgWt8TvpNYM 1; yYPpk4TqAhjX3xC8 12.

## Glutton's Cleaver (item 10772, spell 18075): 2.1, measured

2,354 procs on 33,378 landed hits, 44 wearers, 47 runs, in the main hand or the off hand: 2.12 [2.01, 2.25]. Left out: 1
paladin wearer, who read 1.72; 1 run logged in a second report. EV-0136 gave 2.4. Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      16 | 1,382 | 19,329 | 2.14 |
| Rogue   |      16 |   661 |  9,775 | 2.03 |
| Hunter  |      10 |   242 |  3,310 | 2.19 |
| Shaman  |       2 |    69 |    964 | 2.15 |

Reports and fights: 1RVnhHGv2bwDW9Yr 1-2; 6VJxcqnYa4wdBWQv 1-9; 7NgdQzG6VcRaAJ9X 1-6; 7tCZB31vV2Jp6XWd 1-2;
8vDJ9jmHd4FA7Twt 1-2; 92rYfXaKn1CckRM4 1-2; CR4DHc2wFgjd7yrp 11-13; CnNwpaYW83cJr6fR 1-6; Gx3AXVk9cCaR18Df 9-10, 12-16;
Jad2wbYTXQjBc1n7 1-2; JcwBmYGVgnMQH9xd 3-19; K1zQgCcqbnt3kHXT 5; KzNcw4X2JPRCgVx7 11-19; L7JpRfWtAZ6T1hGx 3, 5;
M7BnDF4VPXcWRaZ1 1-5; MkrFhTq2HPnzmCVJ 3; P8Znjv1fgBD2CAVY 2-12; PDT2yACNzd4LMZ98 1-14; PJXwqxpANGtya937 7-11;
PkWRr8jZp9ATL7N4 1-23, 25-26; W7VPQB18nHfkMaAc 6-13, 15-31; ZDnyzq6XAxdk3K7g 1-5; b2YKZcFzWPtkxfpg 1-3, 6;
bCcFJj4aQwP1rt8m 1-4; dn62GTvrBYFQbWfV 7-20; f8bZ6ytgj7mMGHCk 6-12; gBDMLv8J1K7nqkpd 1-25, 27-70, 72-74, 76-157;
gBYCvhH6MFmbVwkp 1-3, 5-6; gaF6jK78BpnfZqxD 1-4; gnM2vVF7zWXhm9KP 1-4; h9XqNAJk6aPntHpv 12; hbnaypTcAWmqDMt6 10-13;
kpPdVKFmw72nCyM4 1-3; mQJtAfacYp7wGryx 7; nX9kCRqh3WGa8g4d 5-6; p4zvjxb3QZJY2d9a 1-2; pjZHqhYvkLKTAQCJ 1-2, 4-5;
q1ytzBDkYgd6jZMH 1, 3-9; qnyc1zwaB4FrvM9x 1; rzvB21j7yD9hpaVF 12-20, 22-30; rzvB21j7yD9hpaVF 7-20, 22-30;
t681v9FWrLq4Tkxp 1-3, 5-7; wZGf4Y6tjLDyTP92 1, 3-7; wvbng2zVTYRrc9Zd 1-2; xtv8KX4ZzfAmLFPw 3-12; yRQMz8LfGjHNBx1T 1;
zwPKLTQ4W6YkyRqd 3-8.

## Grimclaw (item 1481, spell 1292158): 2.05, measured

754 procs on 11,031 landed hits, 14 wearers, 14 runs, in the main hand or the off hand: 2.05 [1.92, 2.21]. Left out: 1
paladin wearer, who read 1.44. EV-0136 gave 2.15. Reports from 2026-11-10 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   434 | 6,724 | 1.94 |
| Rogue   |       6 |   304 | 4,104 | 2.22 |
| Hunter  |       2 |    16 |   202 | 2.37 |

Reports and fights: 1QpaxjXWFDA7vPV3 1-3; 6rgDRnVFX9a3pbGy 1-4; 8WgnxAakbzDZdFBX 2-6; AwYnpMhJDxPFavjr 1-10;
Fy81NHrkKD3W97qw 1-2; HJxqkQPAnmKtwNjd 1-16, 18, 20-24, 26-31; JgL3qFZA4aQcpXPC 1, 3-9, 11-13; KjLXqdCZANvQc9BT 1-3;
Lcy3V1wHPpqZQAG7 2, 4-12; PJGwRWQYyA3nzXpv 1-12, 14-19; h9pRvCYG8fzNDrcX 4-11; k6XaY3Dy4r7jfTK9 1-3;
m9JVLBndtZ63vhDW 1-2; mBRL3hY1zjaKqkX4 1, 3, 5, 7-25.

## Hanzo Sword (item 8190, spell 16405): 1.05, measured

470 procs on 18,281 landed hits, 16 wearers, 19 runs, in the main hand or the off hand: 1.03 [0.95, 1.11]. Left out: 2
paladin wearers, who read 0.98; 1 run logged in a second report. EV-0136 gave 1.2. Reports from 2026-11-14 to
2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      10 |   284 | 11,490 | 0.99 |
| Warrior |       3 |   104 |  3,425 | 1.21 |
| Hunter  |       3 |    82 |  3,366 | 0.97 |

Reports and fights: 4f2ygdYCzFAvqc6K 6-8, 10, 12-14, 16-18; 6xHzpKd8wCna37Lb 1-14; 824kNhdFaALMBnvR 1-4, 7-18, 20-45;
DNc6QALhxmGbCH3Y 1-6; GHK2gB78FMVCmPfW 1-9; KPqgbZxhn6BXwCvJ 1; MK6XJDzfLpFjnRWc 1-5; bZaY8JyWxpmwQDjB 1;
hwrmtk6aA7zJ8N1Q 1; jn3dwzWPZ6r2ky1K 1-3; kaPpXgt1bWJBr4z2 1-3; mtDjVd6a9rTgWBnC 1; pRA73YvkarFVLNCZ 1-3;
qNyQvka41YhVmdHR 4; qbtZdThvBCXVcfAW 1-6; rTDcQk4gGxfmjN1q 1-7; x4TPAZXnW8kbv19j 1-6, 8; yPVkKpqtfJM27WFB 6-7;
zqpJDYHZC4Pa76bV 1-2.

## Iron Knuckles (item 2942, spell 13491): 2.05, measured

374 procs on 6,461 landed hits, 7 wearers, 8 runs, in the main hand or the off hand: 2.04 [1.91, 2.14]. Left out: 1 run
logged in a second report. EV-0136 gave 2.4. Reports from 2026-11-10 to 2026-11-17.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   290 | 4,926 | 2.08 |
| Rogue   |       4 |    84 | 1,535 | 1.93 |

Reports and fights: 38D6JN4nZzjcGtHg 1-4; 4QAtNfTXMyDGPqCr 1-13; AzbN8GY9DnBrH4aW 19-31; D8vykhYFL14KH7xX 1-3;
DwWM9H3rpTnGPfbt 1-14; GRTmQn3VDpwZKbCA 1-6, 8; H3vdcCr6JX1gwYtq 1-3; ZPyMjD1FkVGJH8Lb 1-4.

## Ironforge Greathammer (item 279895, spell 1322295): 2.8, measured

12,393 procs on 85,376 landed hits, 223 wearers, 255 runs, in the main hand: 2.81 [2.76, 2.86]. Left out: 190 paladin
wearers, who read 3.01; 176 runs in Cat or Bear Form; 6 runs logged in a second report; 1 run with the item swapped away
throughout. EV-0136 gave 3.3. Reports from 2026-11-05 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |     173 | 8,286 | 57,315 | 2.80 |
| Warrior |      46 | 4,078 | 27,808 | 2.84 |
| Druid   |       4 |    29 |    253 | 2.22 |

Reports and fights: 17jHXDAWBRQdgqV6 1-4; 1HK7gFbwy8aJPAh6 1-3; 1fpaDyG3mFnB9KW8 1; 1h9P4rYzWDvtHRxb 1;
1pDcTvkrCBhZJRgf 1-14; 1rH7mXJAxcqkRGt2 1-4; 21j6cGMKRFLpzqrf 1-8; 2FytXTANpbh6dRx1 1-3; 2HahmvNTK89f3qxQ 5-9;
2gjR3tVq4Zdw7kzK 1-3; 2kFjvxn3MzRLQWTD 4-10; 2mW9A7wQ4bhv1YMF 1-3; 3YRkhNP8BZbtDWgx 1-8; 4279fmhA1HPGpDcX 1-2;
48QbknMZw2xt1R79 1-4; 63KPqwWGkh29Abmj 1; 6A24xBzrQkqRFpKh 1-9; 6GcDXN8kWnb4P9mh 1-3; 6HTpWDrd8LKGXwjf 1-5;
6NdR8TAqXMyLmkYg 1-3; 6Y8AjXnH4daWgKym 1-6; 7WLGVBNDXqcar6Ad 1-2; 7kYG3K21FmyMVABD 1-4; 7tahbgBF1jQGRKMq 1-3;
7zK9fMmcNVxnLQA2 1-3; 7zcat1AyFnrJ4BQR 1-19; 8134cDXJahtPZB9W 1; 8PD6Lc3j2qQ91RC7 1; 8Z1TNzMHLCXyGj7k 1-2;
8c1GqjXaPHTpZthY 1-2; 8c7DafNhKtndMBzm 1-3; 8xcD6XhLdNrGq3bP 1; 91YHvCrBN7atKcAq 1-4, 6; A4dNxBP7bY2XkMnc 1-3;
ACvNpyX4DafTgcZw 10-18; ACvNpyX4DafTgcZw 4-18; AGLkhzpCJdVnPDF1 1-3; AL1DxTaHGfVygm3p 6-15; AQJdTrDPn8qyK4Zt 1-6;
AfntVm7bWdMvrP1w 1; B8AKvLZ6bQpaXktV 15-17; Bbd2CDQZyrVv7gFw 8-10; BrTM6CbayRvXgmc7 1-3; ByrG7LFC9mH1N8gD 4-6;
C1fWAmVtPFBGK6ky 1-3; CD1QAmVqHrLyY2Fc 1-8; CQ1hBPYpTF9jZyak 1-2; CXqJYaRfBjnMFxD6 1-15; CYNKBypF2g17j6tA 10-14;
CdH9vgqPbRX1JKM3 1; Cn6rjJpgNBDP7Vxa 1; DmyKc6zWJ7MwHdNR 1-4; DrCRwbFGX3dmtHVW 1-5; FHpqmBdPGYZ3JaDy 1-5;
FTw7k9hdbYKVfH8C 1-5; FVPg91ybaXmGtAYp 2; FkNW7XzKjdVxbqfJ 4-21; FqVXJryknZYWTvKH 1-2; G3KCZD9NbzVpH1Xk 1-4;
GDpvwrBfZ2yYCxT1 1; GVqw3aXB9g7tnZ8W 1-8; GVqw3aXB9g7tnZ8W 7-8; GaynFPTvZf4Xq2b3 1-2; GdDn73hPYt2JQmHL 1-5, 7-8, 10-12;
GwtNT4JLhd8AVWkD 1-2; Gyw6BrhCDx42qcHT 1-4; HBXA4RjC91yxbNQV 1-2; HKQW3wCmr4kJXz2j 1-3; J4LX7AMRmxVh96P1 1;
J8YVraRnGTbtyq6L 1-2; JPDfnQGCH9Xwcbv8 1-4; JnWQzq1m82fjg6Vk 1-6; KGgyJXnVzvMxc28h 1-4; KHqR9VAYTzZ24wxJ 1-11;
KVh6PxG3fHTDWw8A 1-3; KYDWXztmZHQNCVkF 4-5, 7; KtzqRMDGk27NwYyp 1-3; LJ8NCMHzpym4dVqk 5-10; LPQaWZVAkzpbJGFw 1-5;
LVwmRx1Mqak7vp4y 1-8; LmhQHWJDvcV1RdPj 1-4; LthNmKqbQFaAZ2JM 1; MBR3dy4azD1L78Qq 1-3; MQjCX4kG28BxztgT 1-4;
MWqZ1xaVht9TkXJA 1-6; MhfdwPABCaxn9b82 1; NH8YCdhDgtXG6bZK 1-4; NJhx3D4knajQTYcP 1; NLbKRCyqrYPJdmFg 1-6;
NRH63zjvtwpxn1Wg 1-2; NRH63zjvtwpxn1Wg 1-3; NfK46gDLj8wrJz39 1-3; NngHmf2D46Gw7bcL 1-3; PBkXcr2JGgx4NR9Q 5-8;
PJVn4jX8WA1TQzg6 1-3; PK4Z2NcFnQdpLwYg 1-4, 6-7; PN7jfn93M6HKwpGX 1-3; Pck4Hb83WgXxrfJF 1-3; PrQ4yZdgFbGTcpWv 1-4;
QFCfmhc2yt4zR168 1-10; QMH23Y1hqCxamFLj 1-8; QmCHxRYD9tbjLfW1 1-4; QnNkg8ATL4Dbv1hf 1-2; QnNkg8ATL4Dbv1hf 1-4;
R76N48HMyW9V3ZYc 1-5; RGw7haYADHg8yC2m 1-3; RHDbqGV3fgjMrLCX 1-9; RHYb2XPvpDdf17LA 1-5; RKDTjJY4mNc9wbVk 1-5;
TAwnY8WFB4HmQZ9X 1; TBYt3rwJqhGxD1py 1-3; TGvNBcgzCQYD7hya 1-12; TQ4PXCqWA3KGYzh2 1-5; V1L8nj6gMpqRB4WN 1;
VkPnGcCgJ17b2YvH 1-4; VxcH4X3W8da1jYLM 1-4; VyjQ1bDHvPZFdxcN 1-10; VyjQ1bDHvPZFdxcN 4-10; W3wGpxR4tQFcak9K 6-9;
WnjRtrhBVyYwXx3M 1-4; XR78naNGyAmZhMKB 1-5; XWgjPGZD3ryACTv8 1-3; Y3cP8mqfA7F9RjKk 1-4; YC214LNaFPwkRrfm 1-12;
YFBpbDdgxG17CXhL 1-4; YM7tXFPkLZhTgj6z 1, 3-7; YPaFmdJbpXt1W8ZN 1-4; Ynw8cNZv4GJRBC1L 1; YqnVZLX68FRmctKQ 1-25;
YxKjwhRDNyJ4z9cF 1-7; ZADc3bTCKhn9jJdz 1-3; ZPBTJjpyxndNg1Lm 1-2; Zfwc2Q1bKRpMnWGg 1-7; ZmQq1dyC8VJRcpAM 1-4;
aD6nv3CRjrFqXfd2 1-3; aMVzJNgDwpbG143A 1-3; adR7kpKcMDmtHThG 1-8; agztmnCXVNRGqwvx 1-4; anKHZjtTkdVv239J 1-4;
arf1LxvVJGycRFAm 1-2; b2rJTDtLjC7BxgGv 1-3, 5-6; b6zYpvJBq2CVLyfj 4; bDrG4Vpyc9amg1KR 1-3; bTFC3LZRKaHgq8mJ 1-13;
bVNRgda481CqpmfG 1-3; brAQWYqMfywd1JPt 1-3; c7zvhTJGCLaQPZKp 1-5; cH3w7AN4VZXTfr6W 1-3; cT941jZhxvWM2aKd 5, 7-8;
cnz1wmVkXtgGJjZA 1-5; dAgmNB1rwGy8f76Z 1-2; dM46X2j1BZGhYTmx 1-2; dQqfmPcwk4avZ6Gg 3-9; dafrWBRmTJpnPVYv 1-5;
dyTRnv6MZQ7BKaGC 1-4; dzBLDHJTfc13yAWQ 12-27; fCbHhaK7jxv9FkYq 1-2; fDycAFZWt4NkRX1x 1-3; fTHKyQa6YFVj7JWL 1;
fcCgK1R6GpBz8tPX 1-4; fjac46MdLgvHprZq 1; fpGWB7am4PzYnMyN 1-3; fzyDKTP7x8B2hRNw 1; g9xAZ7jBQmRN6TGn 1-5;
gNMJVXAKF7hH6WQx 3-5; gPbhJ73prMtxXcjn 1, 3-7; gkFbNhdxDc384Pyp 5-6; h28ZpnT3kCm6NHMb 1-3; h2CgmZyNBarVjtYP 1, 3-5;
h9L8pBqnVdJ1vDrb 1-12; hBWtTfC1rk8nMQmX 1; hM67TLJGypgwvrzP 1; hVFtnpRGby9azXg8 1-2, 4-26, 28-42; hdtzWJKAZ8cHpGRx 1-3;
hfJZb8HL41KmtdMA 1-4; hjkVZTFrGbRv7cH8 1; hwcWQn8VLN7kJxmT 1; j1CGayX43bMdF8hP 1-3; j23K74gP1LNyYz6d 1, 3-6, 8-17;
j3vLB41G8pkctPMH 1-2; jQP1aWnwH8C2T7Lt 1-4; jpF1ML4ZTkcwGQ8W 1-3; jvtkpbcQCYFz83mT 1-3; k2rGgh6TVDXzc934 1-2, 4;
kY3DG8tzBr2vX1Rm 1-2; kczT3CJVBXa9Gj84 1-16; kyLDgPjrQYdHAc1a 1-2, 4; kzpRQ2gqGJ3yKh9Y 1-4; m9A2MvGDfPFRCWj6 1;
n4xHq7mpR2htyXBr 1, 3-11; nFHZX8TMVG61PRpW 146-149; nWhPF7JrzNRLmx2v 1-5; nadqr4JWQ82ZTkv1 1-3; pG2cZKTAjfYzv8yX 1-10;
pJN1cYKwfgaZ8jt4 1-5; pJQ4mtrDqcaAzwBP 1-4; pRrGZhvDw83LJHTy 2-4, 6; pT6jGtaZFDr92K7R 1; pqBwzhZQy7d9jFnY 1-8;
pqPBFkXgnjDvN9ZG 1; pwNxagqK7HBZ1b9X 1-5; q2h6pRvdBCKMabnL 1-2; q6bKHQCLRJaWTwGy 1-2; qCNzdMA63yLWmhGk 1-3;
qFbMHLVtz94cXnCy 1-4; qNBhWJka9TtLPwV2 1-4; qh8Lgfy31XnQF4jH 1-3; qxY87bFJjG6Tz14A 1-3; qyp6QA92RLJNKxTd 1-14;
qzBa8HcG7xAfPyT3 1-14; r2c7bMKxfd3AHVYT 1-4; r7RQ8FJpnBVgAy6f 1; r8paZAQhXF24Wdw6 1-2; rC3VbJjLQ6YBp8vM 1-3;
rL2DzHGy7t4wxh3N 1-3; rLmJfNtq7v6BFcxy 1-2; rQdnmBHXCpKgJqw4 1; rQt1k2LGcwVfFad7 11-14; rcYRpQA1aZVhPz9B 1-3;
t9DRvJmjap3BrcCw 1-4; tH8kRGA3bph7V9vN 1-3; tNxJXzB3q7rYfRK8 1-5; tWV1d42zgK7H3qNM 1-4; tZK2jg7HqTGAWnRd 1-3;
tbQBwaTD7R2rxpkP 1-3; tgqAx3jr6vGW8nck 1-2; thJy8Y43TPRnHvaK 1-5; v3WBbzYNjaqQJ67c 1-5; vCD8yMjK4WL1ZnqY 1-2;
vRmX3tw8dbkx1HTV 4-8; vVCHXQgNFhqMnbjp 1; vXnHY1xFwVzDBa2k 1-4; vf3QxCLyhckPdbM2 1-5; vk4gw8GBahr6VMxb 1-2;
vm7QzHZCjFGtJ6DM 1-2; w3jgGmAkt7Rfnxah 1-3; w4JfLXqF7HDK32ZV 1-3; wH9M8XGJN1K4h3tk 1-7; wTYjbBFDZN2zkLyV 8;
wfMAD38qZRybWmn6 7-9; wnJHrLF14XD6d3RT 6-8; wtzQKj8GP4cJ2W6H 1; yQ624zGVT18w9nbd 1-10; yVATBnQjpz1MLZPh 1-2;
ykcwLtjHFQT87Mv9 1-3; zfW2RHdjLqKA1yD8 1-5; zr6Z1GFPkpwQVmX3 1-5; zyFZC9kAVDB3Mvj1 1-3.

## Killmaim (item 13016, spell 13318): 0.9, measured

748 procs on 15,268 landed hits, 33 wearers, 36 runs, in the main hand: 0.92 [0.85, 0.99]. Left out: 1 paladin wearer,
who read 1.01; 1 run logged in a second report. EV-0146 gave 1. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |      22 |   467 | 9,382 | 0.93 |
| Warrior |      10 |   275 | 5,672 | 0.91 |
| Hunter  |       1 |     6 |   214 | 0.53 |

Reports and fights: 1g7hAY6CBNtW84HX 1-2; 29A3dKwPgpRNnWvc 7-11; 2aVnTxFzfckdXpBy 1-7; 3vDf6TMYBX9LK7bz 1-2;
3xgPkwfn8yQ9BqAW 1-4; 6hrcXFkgDamVB8jT 1-3; 6v4wTWQkjb3Xd7mM 6; 7hNgRwMfkQZ2Yv4b 1-3; 8GW16QY9BmFHJXLb 1;
8RFv6Xyr43dLpJ7h 1-2; G3VKdgR2aLAkNqWy 1-3; MrW6JCFZTDzdtYpH 1-4; P16dx3VLmRyaDJZQ 1-18; P7cV2HjLgxGpmZQk 1-7;
QhYmjJ4bHZXcgfk1 1-2, 4-27; QqKcMJwHjxNPkgRT 1-3; T6MXt73janNJgmyY 1-3; Vvfz8wYbjTBFaxRq 1; aQwNVCPHWmzDLXqk 1-2;
acHy1fNktP2m47pd 1-2; bdL8j9KGMR6r2xmk 1-3; d7DFtXL2jfNa8Bw4 1-5; fn2Dc941K63CNvmj 3-4; gdKNPjVvcQt2LJnq 1-3;
ghzjGKVBvJbRZnMr 1-2; grFA3BHxtCNyzR7v 1-2; hbW1zgMXyRv8JAVL 1-6; jKJPF16ZnWLkrR2d 1-2; nT2Wz7mVQZbtBRG4 1-4;
nyhMJLqBYxPpRaXf 1-2; pDRxq6wJynmhgdjL 1-9; pqr4HxC3vWJhfG7B 13-20; tMxrNWC6qPVzgwAv 7-8; vJGT19YL26XFQpNB 1-2;
zJ6C2148dTYfaXHb 2-4; zhj69Dt2X1b3fvMT 1.

## Loksey's Training Stick (item 7710, spell 1293482): 2, measured

3,474 procs on 33,273 landed hits, 177 wearers, 219 runs, in the main hand: 2.02 [1.96, 2.09]. Left out: 165 runs in Cat
or Bear Form; 11 runs logged in a second report. EV-0136 gave 2.55. Reports from 2026-11-09 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Shaman  |      24 | 1,336 | 13,356 | 1.94 |
| Warrior |       8 | 1,106 | 10,485 | 2.04 |
| Hunter  |     143 | 1,031 |  9,427 | 2.12 |
| Warlock |       1 |     1 |      4 | 4.84 |
| Mage    |       1 |     0 |      1 | 0.00 |

Reports and fights: 1QXAJwhcYL6M7TDk 8-30, 32-33; 1aBhkzdDMZKGtb37 162-170, 172-176, 178-187, 189-195, 197-237, 239-286,
288-289, 291-295, 297-312, 314-323, 325-328, 331, 333-335, 337-371, 374, 376; 1g2QAZPJrLTjCz3m 10, 12-14;
1hpF2AyzjRdrL78N 1, 3-12; 1xqBcj4CZQXV38Pv 1-2; 1yMxz28Pm4NZbAJf 1-12; 214ndcFyjpPM8wah 1-3, 6-9, 11-15, 19-28, 30;
27jZGhfHgp1LR8MB 1-11; 2PpAjf1qRmTv6gDz 1-17; 2WH4jLt3FADy8fnh 1-2, 5-6; 2aG3Ld8KVARCzrh1 1-7; 2vMwVxLqdB3GRZn6 1-4,
6-7; 39BFWKC2J8RTzyMx 1-3; 3HJrg1BGh8xR6Y2z 1-13; 3PXfhvmTp6NQdwRb 10-12; 3TgpnwxMD71qfLm4 1; 4ARKtCTVhrgBwadG 7-10;
4BhykDjm2p1ncLGK 10-11; 4JYa2mQgpVNMF89R 10-21, 23-30, 32-82, 84-103, 105-110; 4V2RLGtF9anbCKWD 1-23;
4wHnDaRqY9Z2zv1F 1-14, 16; 6NrFMaRQ9yAjxWC2 2, 4-12; 6WG74An9RrCdqkwD 3-8; 6wmLW1N4tBXpTVqy 1-3; 7Vywm9BAZgcMfbT3 14;
7d6p4TBxcjFyPNVr 1-4, 6; 7fgdkyv1njDKa8Gc 6; 7g9DpfNFBA3PatdZ 8-14; 7nCPmLqHd8jMfQk3 4-5, 7-10, 12-13;
7tL3AyRHfawBJF94 1-12; 8VQDf7JtwbTknMLz 6-8; 8X2Kc46faTnxVBLm 3; 8ZdnfAT79zPwmyVM 1, 3-6; 8frHF64GzAhBxDyb 13-14, 16-22;
8tgNprk9LJwW47MK 1-5; 91XtwNrLRzvMkQ62 1-3; 91ZzNbJMRtjrLW6K 1-2; 97FVRX1WLHqDmjAQ 5-12; 9TrdzatyJbQV3vmc 1-7;
9g7NQBqDH6FGfP8L 1-4; 9nxtzJ1bC42FrgWN 1-3; 9rPx2JNYLbCy1waX 11-15; A8NWyY7dhpLDt3zX 12-30, 32-38; A8tQY9p3Dzn6TLRJ 1;
BTRMkxmvfrwcZNaG 3; BydpaqJVW29G8QAc 1-9; CMxBvHAN37YRqbGg 5-6; CPNY7nxkfXpvQgH1 1-3; CXvjmgaJ6YqKzD4k 1-2;
CpDKr6AWJ4fw7xhg 5-7; D6bxHwnpyvc2VNR7 1, 3-8, 10; DWJhjnmAM6GZdbKv 1-2; DcgdHfFZjb87qRJa 1-4, 6-9, 11-12;
FDybHvQ3pPqBracN 1-4; FQnMCW8kRqVv4mT2 8-13; FTXVdBrHJZD9aw7P 1-7, 9-23; FzGMchfD4PkdmJT1 1-4; G6XMnQyqxTdJa2mb 4-15,
17-18; GAdW4fkq7bx9BPmL 1; GBDAWjKNh6Q9XnF7 1-4; GJgyrn6t1vABDQaW 1-3; GKZJBtWAvfPL81bc 1; GrW7bVHhKwJgF4QT 7-9, 11-12;
Gv6NDzAmFjC4P3qL 1-15; H6h87LGmw3gKRc9f 1; HL9zDvk2wWTXcP3N 1-6; HQTANmjP8FB29vYW 1-2; Hp9ThMznR74PwYKb 5-8;
J1QrDzNc4ydjV8ak 1-10; JZv9XDFLC1yKxrPY 5-10; JcwBmYGVgnMQH9xd 14-19; JtrVj2cqL3znKNXh 1; KLNhgMTDAV6xY1Wj 1-3;
Kapf6NnH3Mzqyt49 1-2, 4-5; KapmH2bcWfPqnrY7 1-8; KmjAV91JBhgQLrfP 3-4; L6jJNDYaATph4cfK 1-3; LJMmvtPTzdCpG9k6 5-9;
M4G1VJQdXbvyYLBD 1-2; MgcQ6txHr9R1VjPv 1-5; MpjtZc2zqHaR9hC3 1-19; MyzjakDC9gBbpq6K 2, 4; MzQJ9fgkwjhW1Lv6 12-14;
NHvp73cXPQRC8agz 7-11; NJ7vtRpCzDTfk2W8 4; NWvgFp3kfR2bHyYa 1-35; Nc2grh6ByLvJCkRY 4-6; NkXjqd43AFVYpzGK 5-8;
NkyKCr2XmgLt1YWv 83-87; NzgdqKHnfLJWkFVM 1, 3-12; PAmg9FX8fr23xHt4 1-5; PDT2yACNzd4LMZ98 1-14; PLGKvft3bq8WgRYN 1-2, 4;
PQDRkypFY94rzKwv 1-4, 6-7; PWhMrYaD4Aw8HcCL 1, 3-7; PmR8Lpvazd4q61nA 6-12; Pn2LYHXpjZMDywgA 7-9; PqBvJ94Nhb7ax1Vz 4;
Q1CHmPvpYaxM49qR 4; Q1arfj4WPBNXqmLk 13; Q2xTrDcZqjB9pA7F 9-18; Q8GxhTmH9f6ajcNr 34-39, 41, 43-51;
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
fnYThW3BpxKZv7jw 1-18; ftkqY3cwR7dBZK1J 1, 3-5; fzKwkZLcbYraJdF1 6-9, 11-16; g7Bh9F2b1kz6Cnar 1-2; g8BWDmknXJpvFbdf 1-4;
gAdCTkV1LBxNYFca 1-20, 22-24, 26-32; hBr7xLW6nkVdMKF2 1-2, 4, 6-7; hCzY42vqXrc1mG7w 1-4; hrFwMYgVGq41fNjQ 1-2;
j2RNAQYzZK6DyT17 1; j7V46tC2GkwNYTrz 1-2; jGL3RyT4fYMtwPrn 12-14; jGhpmnHrtyc1VD8W 1-4; jXLYwBW7bZ3nQdfJ 3-4;
kH2bjrXh4BF39GmC 1-2; kQhZqY4DwGBKLdag 1-4; mDbPCdX9AvwYJxGa 16-23; mH68YTDck3dKV4fx 1-2; mqjv1nLcRh6289yk 1-9;
mqnZ7bkf3v86prGV 4-13; n7qPAd4rN1JyHfpR 1-3, 5-9; nGctFJyB8jwX9WHL 1; nZ3NhbvpVXg649Bm 1-10; naCtKF4L3qzpGPmf 4-5;
p1zYDag2NHXPyRt3 1-4; p64RG9PVZcDg1XCk 4-6; pJ4AKDqYPvfwbTkM 5-8; pqfK2AgHZ7DnhvPF 1; pxAgrJDCXHc9tbB6 1-19;
qGywkYrdHKT8NgAQ 1-2, 4-8; qVtTPaDMHfhmG9gr 1-10; qj6gnCAM8kB2RNPy 1-2, 4; qzv6W9fkDaP1MZYH 1-6; r1xm2Tt3bjgBahZF 1-17;
rD8m4JW1bVyYPKGX 1-3, 6; rPAKqpWdFTR6atmN 1-2, 4; tBQwPjhFqYAyCv1L 1-16; tHz9M3arKp6PxA1W 1-14; tLkym2AVWdJ7bM9H 1-8;
tXFAxTLmg3bBNZrv 12-13; taFcCyAx9BD1KWXp 1-16; tkQKL9bGT6JVR3na 5-7, 9-16; twRbTJHfrDXBLj97 1-5; v9Hf4bPtGq1LVykC 1-5,
7-8; vfFGXn4mbLAqkhWR 14-17, 19-20; vnrGwA8Z2RYtPc9g 1-4; wXNczvgHAqT7L8nV 1-13; xQ2J8AbFN19pPD6X 1-15;
xRTpHyY7qC1GmghM 1-3; xYTCXa9myPJr6R8M 1-13; xk367aQTJVzfrKXq 1; xnvgBCHkJf4Y19rV 1-4; xydMjQBYGHW91vTL 1, 3-6;
yKvCQdzpgqY1WwbA 1-7; ythJ7TZDXk6AjHMq 9-10; yzFXjCY3WVbgRPx9 1-2; zjgkHTcZDJNv13bF 40-85, 87, 89-98, 100-104, 106-158,
160-166, 168-174.

## Lord General's Sword (item 11817, spell 15602): 1.95, measured

7,100 procs on 84,624 landed hits, 21 wearers, 33 runs, in the main hand: 1.94 [1.82, 2.03]. Left out: 1 paladin wearer,
who read 2.06. EV-0136 gave 2.1. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      11 | 5,226 | 60,335 | 2.00 |
| Warrior |       8 | 1,555 | 20,718 | 1.73 |
| Hunter  |       2 |   319 |  3,570 | 2.06 |

Reports and fights: 24kgnZAXbKTW89x6 1-2; 26AYbFNQzMZCmfkh 1-16; 6c8P4BLaVAj3bpK7 9; 7TXcFNRzjwHdAybn 3-19;
7dgfrZ3bvhVqxtY8 3-14; 7tbWfYqJFd4vzL3p 3-6; ARFXzbqHDpBLJdhy 1, 3-4; CDKVcP6ghZJtv8AW 3-6, 8-16; HwyJMCLKfbY1PkmG 1-8,
10-38; JfNv1pWML9ZA7ghR 1-4; K6VvbDypXFLQJW7d 1-27, 29; MgC1BXh8fyArkYqa 1-30; MwVH1KQmak73NPd8 8-16;
NK6cCM4pJyRhaAFW 1-7; NXMZV9xjvWpB7GtP 10, 12-107, 109-116, 118-122; NcxbzW4kJYQ7DGhj 1-3; RVkGfcrbZ13MWqQ2 1-3;
TLp1jza6PCrBAX94 1-7; Wn8wHhaxy31MLzjp 1-10; X8z9dgD6HRTZ24Vc 23; fCB4wQt3NWm6k8M9 5-7; gvWZG1cjrChKwdtz 1-2;
hF7dxZ8nJDpk4wNG 1, 3-5; jn3dwzWPZ6r2ky1K 1-3; pPbM8VAdmWCTvLB3 1; pQPwFg8rf6VWhRNc 1-3; pgN9QwMLYVfrqGbv 1-4;
q21PmQwWRfMLvKXc 1; qbtZdThvBCXVcfAW 1-6; v1bPdgDtXypHncfN 1-4; yPVkKpqtfJM27WFB 6-7; zKgGQA6BqvcJtYZ9 1-3.

## Nightblade (item 1982, spell 18211): 2.1, observed

3,577 procs on 29,368 landed hits, 23 wearers, 29 runs, in the main hand: 2.09 [2.01, 2.18]. Observed: every wearer is a
paladin. With no other class to measure, the rate keeps their read. EV-0136 gave 2. Reports from 2026-11-13 to
2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Paladin |      23 | 3,577 | 29,368 | 2.09 |

Reports and fights: 1NBKq2mDWghwR69z 4, 6; 1h8ct2YXTFvdfJMQ 1-3; 1t43QDrG6pfbRAnH 1; 4PtKGL1zcRV8Chfn 1-3;
6HtbKfTNRYjq1cVy 1-4; 7BQKZXWzY9MaxAPq 1-4; 7c21XafhWTAdqNGY 1-10; 8cVPaCpAMmv93bkD 1-9; 9WAxvQbj4HMgJVfZ 1-3, 5-16,
18-21; DXTzdJWyxk7bm1V3 1-5, 7; GvAQn76tacyqYPMw 1-3; HGKLB9Yvhbxdr81X 1-6; LPQd7pfbjDkYZyRz 1-2, 4-5;
Lg82ZFDyQ4MvRm9p 5-11, 13; NcwFg2bZjkrnXfQV 3; P1LGvYV9FdmR8x3n 1; Q1jMALz72T4RPgpW 1; RQjfLHpktThvGVc7 1-8;
TB8FMm9QKW2Htb6D 1-4; V3cYyjhtTgbdx82a 1-2; ZqwpkCPmLMVYN8th 1; aANyFDgb69hxqY2w 3; aNh3Dt8kbZHL9cBM 1;
bT3XwWamfrp7FV26 1; dZ6CjB3hqarMzvbT 1-8; g4Y2frPx8kBmj9Tq 1-2; xJHzjrcgt1kvAWKn 5-8; ymkqDYNTxPjCgwnb 1-12;
ynVw7ZgMx4tjF3fQ 1-6, 8-18.

## Phantom Blade (item 7961, spell 9806): 2.05, measured

679 procs on 12,356 landed hits, 12 wearers, 13 runs, in the main hand or the off hand: 2.06 [1.87, 2.25]. Left out: 2
paladin wearers, who read 1.90. EV-0136 gave 2.25. Reports from 2026-11-15 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |      10 |   465 | 8,771 | 1.99 |
| Hunter  |       1 |   145 | 2,330 | 2.33 |
| Warrior |       1 |    69 | 1,256 | 2.06 |

Reports and fights: 2YfdbFMgGhr7zHwQ 3-4; 6kN78G2xfqy3vwTP 1-6; D2d3hNAbwkPgL1vF 1-10; GzZH1WhTLQ3JaDXt 1-11;
QAHXb9zy4jfgKPVJ 1-6; XFdcWaPL8MthYH6A 1-5; dt4M3ypRnTkwaVQY 1, 3-10; hrV3QgyzvfdZKJwF 1-30; kQ2gDVr4h7NPm1pa 1;
pLtJjx8DCPbHRacv 1-21; twcjmZNGMQP2FABR 1-12; y2dFcNR1rMmKTPAj 1-25; yqFPH1MKv746dQhB 4-7.

## Plaguefang (item 279876, spell 1309315): 1.95, measured

2,929 procs on 43,023 landed hits, 53 wearers, 61 runs, in the main hand or the off hand: 1.95 [1.82, 2.08]. Left out:
13 paladin wearers, who read 1.85; 1 run logged in a second report. No rate was recorded. Reports from 2026-11-05 to
2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      18 | 1,532 | 20,955 | 2.09 |
| Rogue   |      28 | 1,331 | 21,023 | 1.81 |
| Hunter  |       7 |    66 |  1,044 | 1.81 |

Reports and fights: 149bNrXGaKgQL38P 1, 3; 1Hgm9RqadvcjJh2w 1; 1HpaBdm9fFXDjYPy 1-4; 1XCND7YQJ86Ax32q 1, 3, 5;
1ZJVMgcYWTyRQh2w 1-3; 3tPjXKNACMHcn9b7 1-2; 4ChPnBrVK2fJLMaX 4; 4hNWVC2Kzj8RxpZL 1, 3-4; 6TQkbpNXgGZwYat4 1-3;
6tANqvfKQP1Fkdzg 1-2; 7DKRq39nf4yvFWLG 9-14; 7NbYtg6T2pjr1yL8 1-2; 8mBLvfVzCn4tWyTM 1-3; A4NykGFTwmXKj73Y 1-2;
ARBxdp9znVDMTwJ4 1-4; AZLYpdc6r7P2CMRm 1-5; AgFRCzn6X2KDdtMr 1-11, 13-18; BgCQ6Xzp3Nh2HWkP 1, 4; Bxk8b1hKMz6vt9Vd 1;
CJZhDp9FxwRMaWYy 1-2; CfhkbPBKAZFRGp9g 2-3; FcC4vd2DkRqyMrVT 1-2; GAK2BMH8Wt3k9JFj 1-5; HPRbL3N7p6wFhrqk 1-3;
HqQTwzArBZpPN1y3 1-12; JDH7bFz23Qakdgnh 1, 3; KAvPryL1zCdtWTxV 2, 4; KWZaQjqPLJ7Ry8Gd 1-2; KYD67nVqZT4jxWyN 1-6;
L7QHtkhacypgd3D1 1-2; MzwcDrG3fdFBRy6a 3-5; PqjRDCnyaHt13QJb 1-2, 4-5, 7-8, 10-11; Qb7TJqYkKtgyz1ah 1;
RB47hLrMjdDG8Z6a 1; Rp4QCr31TVmgyqnz 1-2; TH63PWxNv9gRarCY 1-2; TnbVQWRJ7r26CqpK 1, 3; W8YC9TA6bcXwLRMk 1-6;
Z4GKmBf1HXdtkYDL 1-5; a8cL9vKhNmkQfBGd 1-3; aVyXCj1NpnRTFKxt 1-2, 4; afNcyKHdtY2qjPCp 1-4, 6; bcATfGPv3DX1Jnxy 1-2;
d7DFtXL2jfNa8Bw4 1-5; dzTmC6F7xwLPJrq9 5-17; fTpYLw71QVR2qky8 2-13; gKmcRvAdBYqrDnbt 1-3; jKfkzadV3BLp1cbr 3;
kLB1K4faPvhmGzqV 1-5; m2FqGpwzr9cQf7PK 1; n6DtHVNKp1wgymFL 1-2; nWxftVpvMycmw2FL 1-4; q8DXKyGLJFb1CwVp 10-13;
qtcFwCBmkXz7JMH8 21-22; tm2FQVrcZzfB97v8 1-2; vYQqkXLm43RJZGy7 1-3; x9vCJk6RYqFLhWGa 2-5; xXtbFrdR4qKAQVB1 7-8;
yLKXxQgrPWTnmdZA 1-3; yMNq3krVBAHxG68w 1-3; yTDCtLmVK4kq7gQ8 1-5.

## Sang'thraze the Deflector (item 9379, spell 1294243): 2.1, measured

1,997 procs on 33,705 landed hits, 25 wearers, 26 runs, in the main hand or the off hand: 2.09 [1.97, 2.26]. Left out: 5
paladin wearers, who read 2.08; 1 run logged in a second report. EV-0136 gave 2.35. Reports from 2026-11-10 to
2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      16 | 1,696 | 28,638 | 2.09 |
| Rogue   |       7 |   247 |  4,120 | 2.12 |
| Hunter  |       2 |    54 |    947 | 2.01 |

Reports and fights: 1CYAWTtmcGdFqbPN 1-5; 28AhM3VGnvFcJ4rD 1-4; 4NqLGxvrZYX2Qwnf 1-4; 6bM1k938LGNghacx 1-9;
8FBqfvLyWTdKxbZh 1-38; 8KPHjdAxZYFNLtgR 1-5; 987q1pjnYdVQbCmT 6; C3zW7BjkJXgRtq8m 1-11, 13-18; FAL4Hv1D7nWf8VNh 1, 3-8;
FVv7P3a2KqJrdG4g 1-3; JDZFAK1z7QbckdHN 1-3; KfvWRDrFXQ1TMq9P 1-8; LXWzFKcBmvyMZaH2 1-2, 4-9; NbwyAxHKzhYXV6L2 1;
VQCRJWfq3KX7M1Nh 1-2; XCJdxAF17zqHMhLY 1-2; Xj23FkTPgrHawv7J 1; ZLpcV6krn1wYHJvt 1-8; bXza3gZq7nRTGVfW 1-4;
brf4Pzj8BkyMmCHJ 1; cPLzXv4nTtNaYWF3 4-20; tCprA1RmvTcJkq3D 1-4; tnyh4dxq6pNXQzLw 1-3; w63W4VhTgCbt9knB 1, 3-10;
yQnph9W1VB4z6a2J 1; zLacx86Thr3AMtqn 1-15.

## Satyr's Lash (item 17752, spell 1292155): 2, measured

1,060 procs on 18,825 landed hits, 25 wearers, 26 runs, in the main hand or the off hand: 1.99 [1.81, 2.14]. EV-0136
gave 2.1. Reports from 2026-11-13 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      18 |   808 | 14,406 | 1.98 |
| Warrior |       2 |   204 |  3,573 | 2.02 |
| Hunter  |       5 |    48 |    846 | 2.00 |

Reports and fights: 4NqLGxvrZYX2Qwnf 5-8; 4PtKGL1zcRV8Chfn 3; 4YzXaGtg81WdqM2N 7-15; 6zJNTjMwWD31vrgy 1;
96tBVjg8RYJNMa7r 1; Bfk6ChZgAr8yWtXv 1-7; C8Qd1LXMrGT62kpY 8-9; L2c7jBZNwFPpb1tV 1-4; Lg82ZFDyQ4MvRm9p 14-26;
MKmntyr82Q4v7jDh 1-5; R1acrQhyNvwXLmBj 1-3; Rc4dCWA6N8PwpBvT 1-8, 10-11; T71XnfaF3cY9M2Rg 1-13; ThKD4jr8MfawcCPv 1-19;
V7Xna6bcWLRzZ3GN 1-6, 8-11; W1zqgCZVPAtTM27b 1-5; YamkzNdPTZFLAnGq 1-8; Yf91cBLN26jdqrxC 1-4; ZwxX9G3bRPFYfjng 1, 3-4;
d4LFCkmQYhPRDx2y 1-6; hyCdAYx8GjgKfZHp 1-5; qm1Bvy29hGK6bMaQ 1-2; rgKPxdF6G3NbMzmj 1-3; tCprA1RmvTcJkq3D 4;
tdYLMmfgcAJCTWyr 8, 10-12; yJmtnzPNDVK42TQR 5-8.

## Silent Fang (item 13953, spell 1293654): 1.45, observed

439 procs on 11,187 landed hits, 8 wearers, 8 runs, in the main hand: 1.47 [1.32, 1.63]. Observed: its interval reaches
further than 10% from the read. Left out: 4 paladin wearers, who read 1.65. The sim had no rate. Reports from 2026-11-16
to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       6 |   366 | 8,827 | 1.55 |
| Hunter  |       2 |    73 | 2,360 | 1.16 |

Reports and fights: 2qMm9WvRpzYx3GaB 1-2; 4hHBvakRnT7p1GVK 6-7; DmjBrYd7M9VZwCHF 14-22; Ya9kKCWRwqjFpgz1 3-4;
fMLcHzyZvt7QTCmg 1-9; nyAx3XNDVmYWGdrh 1-6; pNKJR39CwGnmXtbL 1; y6FZRHQnckD8YjAC 5.

## Strike of the Hydra (item 6909, spell 1292657): 1.95, measured

3,699 procs on 34,120 landed hits, 86 wearers, 100 runs, in the main hand: 1.97 [1.90, 2.04]. Left out: 76 paladin
wearers, who read 2.21; 2 runs logged in a second report. EV-0136 gave 2.25. Reports from 2026-11-07 to 2026-11-18.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |      51 | 3,399 | 31,275 | 1.98 |
| Hunter  |      35 |   300 |  2,845 | 1.92 |

Reports and fights: 12A94BJt3KCFZrY8 1-12, 14-21; 14caBjX2QdynLkq3 1-9; 1AxvQpfrXyTg3dCW 24-42; 1a3pDtCcPgBkL6wm 1-22;
1mRKL7Tb8qwjkgv3 1-2; 1yjwF8MxXqvtD4bT 1-3; 23ZyGvk7BXAVDT4F 1-14; 3BH2d16wtD9F78QY 11-18; 3tRNmGkZB2Jhnv6b 5, 7-9;
4nmJDcGNkZC2Phdw 1-6; 4y8nq6Xjg93zfrc7 1-9; 68fgHQ74dRxrwac3 1-3; 6DYqj7B2WHxPcabQ 128-134; 6X1L4nB7zCfqk8vV 1-2;
6pjNwdr12qzBAKhW 1-2; 7GNK4Y9bkH1awZxL 1-10; 7yRD39nwhmMWYjbX 7-8; 8Zfz4JNLxhHcdDtX 2, 4-5; 8apqXKQg6mbC91L3 14-18, 20,
22-28, 30-36, 38, 40-46, 48-49, 51-71; AqxY6kh43jc2RPaT 16-17; AzMD4BWYKT6NGX2d 2-3; B2k3Xcxz1hqQwJ6P 1-12;
G4hzybWYmgrXvL83 1-2; GZzaJH1QVjMKd8qA 23-34; Gv6NDzAmFjC4P3qL 5-15; H3bRaDmzyjxTW2BJ 1; HARvQ2hJWp8f3grD 1-9;
HpftBG4KPCZrMxbF 1-3, 5-9, 11; HpyNQCMnDYFbJtP6 1-18; HtQr7kYaNq9MGgvZ 1-12; JQH2y6NWBXchGRbg 1-3; KrtPndVLH4BjJcfD 1-4;
L9k3ND6tw8VvcgjP 1-7, 9-18; LAm6bcJw47YBh2kF 1-2; MLhC3KjWrngq2vtZ 1-2; MQwzXcfxb4v7hnFy 1-2; McYJrBbXAfKj82W4 2;
MjfT3pQKRWcVG4gY 8-10; MvrY6aBgftZxj1QX 1-9; N178cBjPtChmbk4w 1-7; NJLbaF8mGnk9yYAq 1; NKXBATdWYqmz78CV 1-2;
PCfQFpjYdm3rD8qy 1-7; Pc3gyq8AFL9CD64m 2; QCLPznqJMZ8vH3mf 1-12; QJ42CwNWLjFyvqfR 1-3; QhYmjJ4bHZXcgfk1 12-27;
TDqbRJWAN3x1FGhv 1, 3-9; TFVvAaY6HLj2NqQp 1-3; VHth6gR7xLbjQmnr 4-14; VQaYAf1Wxvm6ndHK 1-11; Vd1NZCgLMqxhb9kP 1-28,
30-39; XF1faW9mcYtKrqMC 1-18; XVvTwBdtzAcKhPg8 1-14; XpQW1hmwHYDAJ8at 1-12; YX61J8CKqc42paNZ 1-4; Yt6CVhy34zjNHMcm 1, 3;
ZFLd4hkVbrJ68vnK 1-13; Zf21kBmtDM3AyRjF 1-13; ZkA3Vhz72n9Gy4dY 2-7; aRmjxFbtnGYAKCqk 1, 3-4; adXPnhT3kjHp9VcC 1-13, 15;
bPQCc3Nvq27HRWt6 1-3; bdYgfJk7Mh8GqZ4t 1-5; bnx7WwavBpKrQjNh 1-4; c7gHWKFrZMnm2yb4 1; d6f294aV7vYnWmFj 1-5;
dAV7JrjMkGfKNZtY 1-14; dQMKjpxJBCz1wrDW 1, 3-12; dZVp6gHA92mbDnPk 1-3; fnYThW3BpxKZv7jw 1-18; gaDnjKvP6RkQTW2B 1-10;
hG8tvQjLBJgf4WHY 1-2; hgd1qH4QYFyjz6AK 1, 3-6; jrwhRBWdbZPK1Xka 1-3; jtn8ap4XY6PzwDQL 1-11; jwDg8KGcd3QJ9kNZ 1-3;
nB7qNVpyZvLPXCJY 1-20; nJtB3F7cQy6NmzGZ 1-12; nKdyat38zLA7fNJq 1-18; njaHZVpvYFtrTLNh 1-10; pLFZK3MWV2TRw6t7 1-2;
rY4H13hP6MKZfXWR 20-23, 25, 27-31, 33-62, 64-66; t1rjBxRdP8DzVHhT 1-10; tVFZybKHRWp8xMcf 1, 3-10; txP9jAb7gGHhXckB 1-6;
v4QYKcD7XymZq2JB 1-12; vAGVD7aQZxNWX8h1 1-17; vnwtmMTpkR1aYP7G 1-9; vpdJ3cMhwTKNHX7B 1-2; vqBtZRrXwKQb6xHj 1, 3-7;
vyzbt6kgRhTq4fdr 1-4, 6-9; wYmr1kfWPpVz7jyZ 5-13; xnrg9m6jDMR1kaLN 25, 27-29; xwGgV9zJ3FNayZf1 1-2; yDMwzc49TJKmFfCr 1,
3; ycZLWXxQ4jk21w3q 2-28; z4ybMrFWNGa1L2xJ 1-3; zWrR937TBH1FgvfP 1-2; ztD4pKHA6jGvfcVr 1-2.

## Tainted Pierce (item 8225, spell 13530): 2, observed

841 procs on 13,218 landed hits, 15 wearers, 17 runs, in the main hand or the off hand: 2.01 [1.75, 2.20]. Observed: its
interval reaches further than 10% from the read. Left out: 3 paladin wearers, who read 1.91. EV-0136 gave 2.15. Reports
from 2026-11-14 to 2026-11-18.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       9 |   431 | 6,629 | 2.05 |
| Warrior |       5 |   352 | 5,547 | 2.00 |
| Hunter  |       1 |    58 | 1,043 | 1.76 |

Reports and fights: 1zLtZcp3DRP97Hdx 1-6; 4na1LcCpP8QHTNVD 1, 3-8; 8WfLRh1qTpFxJgra 1-2; 8rNkjZ2B6WdzFcKq 1-3, 5-7;
J97GvmrRXCFZfQ8a 1-2; NXgv93kr2G7zyTxm 1, 3-9; R3CDcTfm6bdNJrHx 1, 3-7; RcTKFQbGnmtJHWpa 1-2; TnbzrVaxtHcpMKLY 1-3;
VyZXRT7nrPJKBAt8 5-13; akPJ4AcBZQnb36f2 1-5; c9tKjzRyaYXQfnk8 1-3; gn4HAC1RqMfLt8wN 16-21; pFNnChR7GzAX6aH3 1;
qcrKJGPvkfhAtdjQ 1-4; xqy2GfPj67XCwrT3 1-4; zTMW1B2rRJxbm3w6 4.

## Taran Icebreaker (item 2915, spell 16415): 0.95, observed

685 procs on 11,141 landed hits, 8 wearers, 9 runs, in the main hand: 0.97 [0.92, 1.03]. Observed: every wearer is a
shaman. Left out: 4 paladin wearers, who read 0.90. EV-0136 gave 0.9. Reports from 2026-11-13 to 2026-11-18.

| Class  | Wearers | Procs |   Hits |  PPM |
| ------ | ------: | ----: | -----: | ---: |
| Shaman |       8 |   685 | 11,141 | 0.97 |

Reports and fights: 1MA3qWLGa8v2hdXP 1-48; Bz8CAjZnNyXhgTtH 1-14; FRmWJKpCD8t7QGb2 1-3; PJHx4V21KhfCaryz 1-2;
aAcjGPbyZXNTzR1K 1; aQAhMLzBbWJcdq6G 1-6; cLxk98bVfwpXWaqN 1-21, 23; mg3MBw6rcNtZpjbA 1-14; nhtYVjb7CAw9DvQR 1.

## Thornspike (item 6681, spell 1293183): 3.2, measured

17,449 procs on 191,171 landed hits, 233 wearers, 268 runs, in the main hand or the off hand: 3.22 [3.14, 3.32]. Left
out: 2 runs in Cat or Bear Form; 3 runs logged in a second report; 1 run with the item swapped away throughout. EV-0136
gave 3.75. Reports from 2026-11-07 to 2026-11-18.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Rogue   |     160 | 12,726 | 139,716 | 3.21 |
| Warrior |      40 |  3,919 |  42,224 | 3.28 |
| Hunter  |      28 |    406 |   4,537 | 3.16 |
| Shaman  |       4 |    398 |   4,692 | 2.99 |
| Warlock |       1 |      0 |       2 | 0.00 |

Reports and fights: 12RB4fbmrp3YZjPF 1-5; 1VKQY2Dafc7PJzpR 1-2; 1Zv6bBfgDP3RVFYa 1-2, 4-7; 1kWMb9cQ67dJhrZz 1-3;
1tCvNKHFM3c2WkmA 1-12; 1w8AbW4RBNLX6Kfy 1; 21PKRb9nk3HdYFm8 1-11; 2K4Z1jLyGWzdnD3f 1-8, 10, 12; 2KkaByn6H9mzb4Yq 1-4;
2t6mDPLkfWBxcr4q 1-18; 39BFWKC2J8RTzyMx 1-3; 3DdtwKbCf7g1XVyT 1-2; 3PXfhvmTp6NQdwRb 1, 3-12; 3ZFVrmzh4fanLM1B 1-8;
3bqWjVMPpDfdCgYz 1, 3; 3f4WnNCQJX9azxmV 1-5; 3hNwr9K71bz4xjCc 1; 3kqWrL1dwJfxMcX2 1-5; 3npdHy9Ag7rMTC6f 1-5;
4KCzQ8mFr9A1bkxG 1-8; 4L7DBHTV2jA6gkQW 3-19; 4WYavRTZm7QHxLgA 1-20; 4ZNwHaLtygF1k3xC 1-3; 4zYkPVvpb9KXLxA7 3-5, 7;
6tTZVHLYkbzrCXqD 1-8; 72gPMatGLpwd93jW 13; 76ZV8FQ1RHDhbq3z 1-4, 6-9; 7M1KwnDZkXvyRFB3 1-5; 7dZrzYqjKHh28mvW 1-2;
7dZt6b9Kh4F2CNAn 1-12; 7f6DBRAPMLGYzyQm 1-2; 7g9DpfNFBA3PatdZ 1-7; 7gPtBMrHwGyjWKRc 1; 7xGQWc4AbvD8CNBw 1-7;
81CznKf7taTFArLw 1-8, 10-16; 8MjcygF3zWK4d9NB 1-65, 67-83, 85-93, 95-107, 109-131, 133-142, 144-155, 157-179, 181-203,
205-227, 229-239, 241-252, 254-264, 266-277, 279-289, 291-302, 304-316, 318-322, 324-340, 342-367, 369-386, 388-415,
417-435, 437-466, 468-486, 488-512, 514-518, 520-538, 540-565, 567-574, 576-594, 596-621, 623-633, 635-653, 655-680,
682-686, 688-697, 699-718, 720-745, 747-751, 753-759, 761-765, 767-786, 788-813, 815-819, 821-830, 833-837, 839-858,
860-885, 887-891, 893-902, 904-906, 908-912, 914-933, 935-960, 962-966, 968-977, 979-983, 985-989, 991-1010, 1012-1037,
1039-1043, 1045-1054, 1056-1060, 1062-1066, 1068-1087, 1089-1114, 1116-1120, 1122-1131, 1133-1146, 1148-1167, 1169-1194,
1196-1200, 1202-1211, 1213-1233, 1235-1260, 1262-1266, 1268-1277, 1279-1299, 1301-1326, 1328-1332, 1334-1343, 1345-1365,
1367-1392, 1394-1398, 1400-1409, 1411-1431, 1433-1458, 1460-1464, 1466-1475, 1477-1497, 1499-1524, 1526-1530, 1532-1541,
1543-1563, 1565-1590, 1592-1596, 1598-1607, 1609-1629, 1631-1656, 1658-1662, 1664-1673, 1675-1683, 1685-1689, 1691-1714,
1716-1741, 1743-1747, 1749-1758, 1760-1770, 1773-1777, 1779-1802, 1804-1829, 1831-1835, 1837-1846, 1848-1858, 1860-1863,
1866-1870, 1872-1895, 1897-1922, 1924-1928, 1930-1939, 1941-1951, 1953-1956, 1958-1961, 1964-1969, 1971-1994, 1996-2021,
2023-2027, 2029-2038, 2040-2050, 2052-2055, 2057-2060, 2063-2069, 2071-2094, 2096-2121, 2123-2127, 2129-2138, 2140-2150,
2152-2155, 2157-2160, 2163-2164, 2166-2170, 2172-2195, 2197-2222, 2224-2228, 2230-2239, 2241-2251, 2253-2256, 2258-2261,
2264-2268, 2270-2274, 2276-2299, 2301-2326, 2328-2332, 2334-2343, 2345-2355, 2357-2360, 2362-2365, 2368-2373, 2375-2383,
2385-2408, 2410-2435, 2437-2441, 2443-2452, 2454-2464, 2466-2469, 2471-2474, 2477-2482, 2484-2490, 2492-2496, 2498-2521,
2523-2548, 2550-2554, 2556-2565, 2567-2577, 2579-2582, 2584-2587, 2590-2595, 2597-2606, 2608-2612, 2614-2637, 2639-2664,
2666-2670, 2672-2681, 2683-2693, 2695-2698, 2700-2703, 2706-2711, 2713-2722, 2724-2725, 2727-2733, 2735-2758, 2760-2785,
2787-2791, 2793-2802, 2804-2814, 2816-2819, 2821-2824, 2827-2832, 2834-2843, 2845-2846, 2848, 2850-2857, 2859-2883,
2885-2910, 2912-2916, 2918-2927, 2929-2939, 2941-2944, 2946-2949, 2952-2957, 2959-2968, 2970-2971, 2973, 2975-2976,
2978-2982, 2984-3008, 3010-3035, 3037-3041, 3043-3052, 3054-3064, 3066-3069, 3071-3074, 3077-3082, 3084-3093, 3095-3096,
3098, 3100-3104, 3106-3108, 3110-3114, 3116-3128, 3130-3141, 3143-3168, 3170-3174, 3176-3185, 3187-3197, 3199-3202,
3204-3207, 3210-3215, 3217-3226, 3228-3229, 3231, 3233-3237, 3239-3245, 3247-3251, 3253-3265, 3267-3278, 3280-3305,
3307-3311, 3313-3322, 3324-3334, 3336-3339, 3341-3344, 3347-3352, 3354-3363, 3365-3366, 3368, 3370-3374, 3376-3386,
3388-3392, 3394-3406, 3408-3419, 3421-3446, 3448-3452, 3454-3463, 3465-3475, 3477-3480, 3482-3485, 3488-3493, 3495-3504,
3506-3507, 3509, 3511-3515, 3517-3533, 3535-3539, 3541-3553, 3555-3566, 3568-3593, 3595-3599, 3601-3610, 3612-3622,
3624-3627, 3629-3632, 3635-3640, 3642-3651, 3653-3654, 3656, 3658-3662, 3664-3688, 3690-3694, 3696-3708, 3710-3721,
3723-3748, 3750-3754, 3756-3765, 3767-3777, 3779-3782, 3784-3787, 3790-3795, 3797-3806, 3808-3809, 3811, 3813-3817,
3819-3848, 3850-3854, 3856-3868, 3870-3881, 3883-3908, 3910-3914, 3916-3925, 3927-3937, 3939-3942, 3944-3947, 3950-3955,
3957-3966, 3968-3969, 3971, 3973-3977, 3979-4016, 4018-4022, 4024-4036, 4038-4049, 4051-4076, 4078-4082, 4084-4093,
4095-4105, 4107-4110, 4112-4115, 4118-4123, 4125-4134, 4136-4137, 4139, 4141-4145, 4147-4189, 4191-4195, 4197-4209,
4211-4222, 4224-4249, 4251-4255, 4257-4266, 4268-4278, 4280-4283, 4285-4288, 4291-4296, 4298-4307, 4309-4310, 4312,
4314-4318, 4320-4363, 4365-4369, 4371-4383, 4385-4396, 4398-4423, 4425-4429, 4431-4440, 4442-4452, 4454-4457, 4459-4462,
4465-4470, 4472-4481, 4483-4484, 4486, 4488-4492, 4494-4552, 4554-4558, 4560-4572, 4574-4585, 4587-4612, 4614-4618,
4620-4629, 4631-4641, 4643-4646, 4648-4651, 4654-4659, 4661-4670, 4672-4673, 4675, 4677-4681, 4683-4746, 4748-4752,
4754-4766; 8Z1rYqmCWKbLjvwz 1-2; 8ZTCpbxtFA1zfRNX 1-3; 8cyBqw6Cr3WjV9Ft 1-6; 8raZtT1Qq7WkYv2P 2-5; 9B8Wf16dHKjgaMDY 1;
9Nayz1f8BJnLcgQm 1; 9PxCyQfrdhXFmNGV 1-2; 9g7NQBqDH6FGfP8L 1-4; B1xchF3f6L2Mkga4 198-205, 207-219, 221-238, 240-257,
259-269, 271-282, 284-287, 289-292, 294-306; B1xchF3f6L2Mkga4 251-257, 259-269, 271-282, 284-287, 289-292, 294-306;
B6jfw4mcnPxWCKYD 2-18; BZxkqfarM97yHgLp 1-6; BdLc4AgFbrz8Tq2k 1-2; BxpTfqGZb34aJLc6 12-14; CPj7JpNFLDztgQRa 10-19;
CbG9THJkRBjDmPX4 1, 3-6; CjLKJRNhnb7TBzAZ 17-20; Ct8XcxY26Mf34NjR 1-9; CwYNtp2VQrJdbfx9 5-8; D8qzgc9h1YLf6krp 5-6;
DHyYh18vLXAFTJqt 1-5; DRv3TVCwrAPJ9zK8 1-2; FBQzxArfjdCZp6Lg 2-11, 13-18; FTrHGQmvDM2jzRn4 7-10; Fj9WzagMcZxw4mXY 1;
FvdWRrT67kAPY43x 1-11; FzCJvnHfGyMLWVwb 1-11; G6wYN314W9FfVXpr 1-2; GAgDtJmKHdBhnpwV 1-5; GCyghdxmrfYL8zcX 1;
GJxyK8rTBamY7CvH 1-11; GKmRMd7XzWtavJg8 1, 3-9; GNRbXQDvCkYpAta4 1-8, 10-12, 14-22; GRZrtByWvML6mzXA 1-11;
GRdTC8kXNwVxpjKQ 1-2; GV4wtvzD8RcfTqNC 2, 4, 6-15; GVzLqkhK4PMXjYmn 38-50, 52-70; GX7CDf4dvtLmR1Wh 1-3;
GYPbQpyCvVajRH4L 1-16; GYQB4c79w6TF1nWZ 1-6, 8, 10-11; Gb2W8Tga4VqvZy9c 5-11; GpdNkxVgzrtF49Xb 1-3;
H7yaPANhDBKQ8Lmq 1-13, 15-17; HTpFLJAq8CWmPvzr 1-2, 4; HWcyVPwhnNB9brYz 1-5; Hdnh1GJgD4TbQkXL 1-14;
HqW4rjd6XmYgDM9F 1-2; J3tdCaXR8hAP9m2M 1-4, 6-9, 11; J9jRmkxvtXcAKfM3 1-4; JB7fzC2ZnP3Kwjxd 1, 3; JD1ypjL3xcCKfTGP 1-2,
4-6; JNZVwGWTBMYLz9an 1-13; JPB6NXkAGCQzmh38 1-5; JZx3p6Nqdj9tCrMc 1-3, 5; JvdQP6VyY19kZDXm 1-9; K3jWRx2qQ1aXy9PA 1, 3,
6-13; KG4FcnrYWVHCfzvt 1-9; KNqTmpW6HZCnc378 1-4; KmahZwCJcYFLj1Rq 1-7, 9-16; KmjAV91JBhgQLrfP 3-4;
KvF7tTamGpn9kW34 15-24; LN78yvQwgJpcAYCG 4-6; LW4qmQykhj3p9atv 1-2; LWx7bgKmpQ6ktYRr 5-12, 14; LaMmthYQK4HRWwXC 1-5;
LaQztr3mDTwHyknB 5-8; LmDb8njHfXvAgC79 1, 3-42, 44-51, 53-63; LxdYWR47ZFbTNBhM 1-2; M4CvXLZz3qm6pDRc 1-3, 5;
M4CvXLZz3qm6pDRc 2-3, 5; MCdZTywRz3HYnvWg 1-10; MKk2adJtjyANf8LW 1-13; MW1Kw23ncjRDrxpA 1-14; McNx4Dwk89vYJpdH 1-5;
MdaKf7zpNgWHF1Jj 1-4; Mf1AyXnC6W3dpNJR 1-12, 14-16; MyZ8p6RaAFBWcJ1f 1-7; NLzDkMAj8n2TmbxV 1-6; NWB8HZd1JvxzGjDr 1-3, 5;
NZcYhqPxkrJMKz6a 1; NfLb7C1qDZ8cHxFp 19-29; NwKVWqHfgk1jLvAG 1-8, 10-13; NyZgJMVBXHn8vtWk 1-8; PFGvpDf9WZgcY6Rk 1-12;
PHqvNCxT1JVtpWFj 1-4; PN64hvHQLZRCptGg 1-5; PNnXcA9hFBx7Wtdj 7-11; PyH9Adag8VwRjKJh 1-16; Q1arfj4WPBNXqmLk 11, 13;
Q4GadwADtxbFBNW9 4-8; QAc4wgTL7PfDphk6 1-5; QRbmZWh2yDGakt9Y 1-4, 7; QxJKAphaZt61kzvg 1; QyYtrwM6zKPXHDj3 11-15;
RDPLkYamNgHKXFh2 3-7; RbLqYxNQj1MaWFdP 23-29; RdmgVNQ4XJMfnC1F 4, 6-10; RfxgwvQjA1pVy98k 1-6; T4pxtrKv178PGNkA 9, 11-14,
16, 18-26, 29-41, 43-57, 59-120, 122-142, 144-164, 166-208; TgVP6hxZKqY7MGQ4 1-5; Twr9aYQFMWjnLPqR 1;
VMP6Dv2pyTnYhaKJ 1-3; VRM3KtXTh2W7kzJx 1-4; VYb6f27NazpmCch3 1-2; Vh3bzdRMP6ymNKq9 1-2; VhHLyw9CP76Wv3fX 1;
VvmGw1CT7bZMNXp6 1-12, 14-19; WAKwhqRtmPGT49ZQ 21-35; WFJf43apKLXHVzBh 1-4, 6-7; WqBa2AwcGHfgj8Lt 7;
Wza8bdXVRhcgqGnZ 1-5; X16Yp9FzW87cNADq 1-5; XYTRKMVFWdypLxtB 1-14; XhZAV37vMFbYWRt6 1, 3; XwqT7ZK48Df69tkh 14-22;
YJzQRVL4r9DqFcpf 1-16; ZCRLWJqmb7v3gjaN 1, 3-7; ZKRmDGdV8FM1aYwt 1-13; ZPK6pnD7HzmyhvjT 1; ZPxzwLAmBpRMTNbC 1-3;
ZPyMjD1FkVGJH8Lb 1-4; avQywnDKt7RxcpfF 1-8; b12KD98fndF4hMq3 10-18; b4fNagK2jvXQJAGL 1-10; bAZJWP3hgxRav6Fm 1-7;
bdPKyBjRNDX9CHaQ 15; bpV7JWX8B9xgrj3c 1-5; bxywkpHBTzg2mtCQ 1-20; byYP67TH4KtZqm31 1-13, 15-17; cLkwmZbzFDKdPhYn 21-22;
cPHdgRMCy7ZGDqFf 1-8; cT8wJVq316kAWpn9 1-3; cVPtG1agnWhfX2xC 1-7; cXCfvnWtNrR63PVx 1-2; ctAnYvNLQK6F38PD 1-2;
d8KLZbmzWhwr9at7 1-4; dXR2BPqTYQ4aMxZj 1-5, 7; dtHTkZgMNvR3YBK4 1; f9vADt2CZ3M1ygwq 1-11; fRPcJYzWKXBCv32N 1-2;
fjpy1TBL4b3kgJhr 1-8, 10-15; fjqQNyMbD1Y7R92h 4-21; fkD9m3dBKcGAyqRY 1-6; fnMH4F2pkGC8T6Kz 1-8; fr2WyGqPVYNTXd91 1-4;
g2qW8NnDvCQ3TfVG 1-5; g8b1MkJnKtyfzvPa 1-5; gGn1vQ28YxpFrmaP 1-3; gT2mW39CV7DMwhzx 1-21; gbFJzMXnTKkxNQY7 1;
gdMNAjD2Kyrtqh97 1-8; ghcTDmb871vtxyWM 1-5, 7-11; gn7Wva1FKAm8HhGZ 21-23; gwkN81FaqdAPQTfW 1-4, 6-8, 10-11;
hXT9DqC7vpQW6ZLb 1-9; hf1pF42aPHMv9kLX 1-9, 11-20, 22, 24-36; hf1pF42aPHMv9kLX 18-20, 22, 24-36; hmtTqH16pPGzZJax 1,
3-17; hvgw9t1par7YCD3b 1-17; j7J4bKFHq6M2Wfvy 3-6; jD7whNgBrM4GxVvF 1-8; jGrv6B1hNMDJPc3W 5-10; jHXnNA128R4FmwYB 1-5;
jQzWyR89avmgpd7H 1, 3-7; jYAhJZ3BzPMHWx1q 1-5; jnxaGm43ctgQMFhf 1-8; jtdGCBNJq38zxcQ9 1, 3-9; jyxcbwYXG3pL9NnP 1-2;
kHfNcJTYj3RtMFGL 1-3; kJHLMmf3rGaj28d1 1-2; kX94Vmn8tpJzGTD2 1-11; kqXb3Tja4gZfxP7V 4-6; mDbPCdX9AvwYJxGa 7-30;
mMFCR2atzKHWDcdQ 1-6; mPMpkAgjbtR7HBXD 1-19; mc821XQyrpDjGJh9 1-3; n3hjGAFXxTaBfmRz 1-3; ndHRLVB9Kaz7xpqT 1-2;
nmrNLB3j4pAPg8fQ 1-2, 4-15; pKn8wBZcQ62PyJH7 8-24; prbPqjNgMz6FaBX7 1-7; pz7YaGwxXZgcbQPm 7-9; q9FHYznbZLDKtQTP 1-2;
qLk1fVdN3rcaDxm6 1; qMD3CVapwb6nAgx2 35-40, 42-43; qMD3CVapwb6nAgx2 40, 42-43; qQhkt3L9yKfTFp2D 1-6, 8-13;
qYRLwvVKDkTZGP86 20-43; qg8YWNaCdDvV7XZM 1-8, 10; qyWNYFVZjLX1mtJA 1-19; r67HcKRnkYdzTQZP 1-19; rbg3wR9B7dKf8JmW 1-3;
rpLZ287nfKWyJGtb 1-6, 8; t7432ga9HXvFBKzy 1-6; tD7Rm3GBjdqPzhwC 1-4, 6-8, 10-14, 16-17; tLXnZDMCvJ34mzbF 1;
tc1qGLjpgVrJ8WdQ 4-5, 7-15; th1nf93NzQjmVdrD 1-15; tjN8BgxX7k4GyAfd 4; tyAMjpZb9Yq2wkV7 1-4, 6-10, 12-19;
vCZk3N8n2dbTyMFV 1-34; vDg3LH1Zyfz49kKr 1-12; vFqn26gPR4NJBfjw 1-4; vKznWFRC7cBHTGwx 3-6, 8, 10-11;
vpBwx4KbaWzRZ8V7 1-3, 5-10; vtRzBNFnW6KH3pjT 7-8; vzZPBtWcfm9HY8Ty 4-10, 12-18; w3L1hH2ya4gCTKAq 1-13;
wCGWfTxNgFrB8hDy 1-15; wHtnq3pgjxYVy9FZ 1-2; wRY9NtQp7bPZ1VDz 1-6, 8-12; waYzfQgyrcKV7H9J 1; wfgxbHpNB2vZJYLW 1-6;
wvbng2zVTYRrc9Zd 1-2; x7ybf4LjBD6GJqWn 1-9, 11-15; xDwACkbMZRyNcd1z 1-5; xJbw6RTvtHPncNzp 1-5; xRNZwp7fWvTY29qV 1-15;
xbCaBqNdAn3ZWKDw 1-3, 5; xmYzGavFHWpfbLkt 1-3; xnrg9m6jDMR1kaLN 21-23, 25, 27-29; xrtBTqdwGDfzK72Q 1-2;
xtFwVRc3gmdjLAyC 2; xw86k9LyMNF3YHjr 10-13; xw86k9LyMNF3YHjr 7-14, 16-21, 23-31; y6kYARn4dpCKavZ8 1-3, 5-6;
ycjXJnNkC8pGWH1Z 2-7; ypQ9d13tMFJPNfvw 1, 3; zDhrvnfPYdt9FXMj 1-4; zjxyZAgPDBMGndRK 1-16, 18-20, 22-24, 26;
zwPKLTQ4W6YkyRqd 3-8.

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

# EV-0175: Proc rate of the Sword of Zeal

- Claim: Sword of Zeal procs Zeal at 1.95 PPM, priced on the weapon base speed of each hit that rolls it; the rate is
  measured.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights listed below, from every wearer
  stream on disk for the weapon; 25 reports in all
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date; 4 reports start after hotfix push 114313 (2026-11-20 23:47
  UTC). The sim's database is 1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-21
- Pull request: #1160

## Method

The count is EV-0153 §Method, run on 2026-11-21 by the sweep's analyzer on this weapon. Six wearer streams (reports
L7tdvhcYjgwHTrMz, PThANm6GzytHYDrd, 7YZCHRjDh9wQMFy6, 7tf9xk14DBqPXjWa, ARMrB9JYb2D1TLZC and YwJVCfGj3hnFpRy1) were
fetched for it; the others were already on disk.

## Rate

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs;
Was is the row's rate before this count.

| Weapon        | Item | Rate | Call     | All hits [95%]    |   Procs / hits | Wearers | One-hand white | Specials   | Was |
| ------------- | ---: | ---: | -------- | ----------------- | -------------: | ------: | -------------- | ---------- | --: |
| Sword of Zeal | 6622 | 1.95 | measured | 1.93 [1.90, 2.02] | 3,079 / 34,152 |      12 | 2.01 (302)     | 1.91 (355) | 1.8 |

## Sword of Zeal (item 6622, spell 8191): 1.95, measured

3,079 procs on 34,152 rolling hits, 12 wearers, 25 runs, in the main hand: all hits 1.93 [1.90, 2.02]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.01 [1.79, 2.25], 302 procs on
3,214 lone units, 2 wearers (a Poisson interval: under 5 wearers). Specials: 1.91 [1.80, 2.05], 355 procs on 3,989 lone
units, 12 wearers; Sinister Strike gives 181 procs on 1,866 of them. Dual-wield white, which sets no rate: 1.56 [1.49,
1.64], 227 procs on 3,114 lone units, 11 wearers. Two-roll abilities, priced as one roll: 3.23, 64 procs on 424 lone
units, left out of the rate; Execute gives 62 procs on 403 of them and Mongoose Bite 2 on 21. Direct: window [-20, 100]
ms, median -1 ms, 98th percentile 0 ms, 648 isolated hits. Dropping the swapped stretches (5.8 active minutes, 0.5% of
hits, in 1 run) moved the read from 1.92 to 1.93. Report PThANm6GzytHYDrd logs the same run as L7tdvhcYjgwHTrMz and
counts once. No paladin wears the weapon in these reports.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |       9 | 1,128 | 12,129 | 1.99 |
| Warrior |       2 | 1,899 | 21,481 | 1.89 |
| Hunter  |       1 |    52 |    542 | 2.05 |

Reports and fights: 2YMHxKd8h1zwrAZW 1; 36XqpAkDvTmjZrhR 1-30; 63QrzvTALnZKgx7Y 1; 6c8P4BLaVAj3bpK7 1-9;
7YZCHRjDh9wQMFy6 1-4; 7tf9xk14DBqPXjWa 1-3; 82WLwYNzvjB3bXM6 1-2; 8ATZQ31xabwdgf2y 1; 9zV8D3nxcAwQqhTa 1-3;
ARMrB9JYb2D1TLZC 7-15; GFBCh3VZH7TmAkL4 2-3; Ht1CryL9kJDKFabv 1; KAv7LjkFHTwpZbD3 1-2; L7tdvhcYjgwHTrMz 1;
NtxKh7PgTAaZnHJd 1-4, 6-7, 9-11; VMBZcafW2JLbT6dK 1-2; YwJVCfGj3hnFpRy1 1-2; fFNJ7HtK49hLDvCd 1-5, 7-9;
gVHqdJ1Xwf3kBny9 1; jNVAY2Pav8GcbFWn 1-2; m69XVCZpd7Dn8AyH 4-8; qgfYWFzjGtmJTpAX 1-7; rcM96QNt3CkxTWmg 1;
wFjBdZtXxvm39yzq 1-3; wh2p6j3rZfHMaYzv 1-17, 19.

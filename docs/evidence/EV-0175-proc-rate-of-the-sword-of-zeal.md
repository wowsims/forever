# EV-0175: Proc rate of the Sword of Zeal

- Claim: Sword of Zeal procs Zeal at 1.95 PPM, priced on the weapon base speed of each hit that rolls it; the rate is
  measured.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights listed below, from every wearer
  stream on disk for the weapon; 31 reports in all
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 10 reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is 1.60.1.71037, push
  114326; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-22
- Pull request: #1168

## Method

The count is EV-0177 §Method, run on this weapon.

## Rate

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon        | Item | Rate | Call     | All hits [95%]    |   Procs / hits | Wearers | One-hand white | Specials   |
| ------------- | ---: | ---: | -------- | ----------------- | -------------: | ------: | -------------- | ---------- |
| Sword of Zeal | 6622 | 1.95 | measured | 1.93 [1.91, 2.00] | 3,687 / 40,881 |      16 | 2.01 (302)     | 1.85 (441) |

## Sword of Zeal (item 6622, spell 8191): 1.95, measured

3,687 procs on 40,881 rolling hits, 16 wearers, 30 runs, in the main hand: all hits 1.93 [1.91, 2.00]. The rate is all
hits, which the one-hand white rate and the specials rate agree with. One-hand white: 2.01 [1.96, 2.41], 302 procs on
3,214 lone units, 2 wearers. Dual-wield white, which sets no rate: 1.60 [1.51, 1.73], 306 procs on 4,086 lone units, 15
wearers. Specials: 1.85 [1.74, 1.98], 441 procs on 5,097 lone units, 16 wearers; Sinister Strike gives 242 procs on
2,699 of them. Two-roll abilities, priced as one roll: 3.23, 64 procs on 425 lone units (Execute 62 on 403 and Mongoose
Bite 2 on 22), left out of the rate. Direct: window [-20, 100] ms, median -1 ms, 98th percentile 0 ms, 735 isolated
hits. Dropping the swapped stretches (5.8 active minutes, 0.5% of hits, in 1 run) moved the read from 1.92 to 1.93.
Report PThANm6GzytHYDrd logs the same run as L7tdvhcYjgwHTrMz and counts once. Reports from 2026-11-13 to 2026-11-22.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Warrior |       2 | 1,899 | 21,481 | 1.89 |
| Rogue   |      12 | 1,691 | 18,348 | 1.97 |
| Hunter  |       2 |    97 |  1,051 | 1.98 |

Reports and fights: 2YMHxKd8h1zwrAZW 1; 36XqpAkDvTmjZrhR 1-30; 63QrzvTALnZKgx7Y 1; 6c8P4BLaVAj3bpK7 1-9;
7YZCHRjDh9wQMFy6 1-4; 7faQXRhyZr6V32Fk 5-8, 10-13, 15-22, 24-28, 30-35; 7tf9xk14DBqPXjWa 1-3; 82WLwYNzvjB3bXM6 1-2;
8ATZQ31xabwdgf2y 1; 9zV8D3nxcAwQqhTa 1-3; ARMrB9JYb2D1TLZC 7-15; GFBCh3VZH7TmAkL4 2-3; H4nMT6FVBwXmJbP1 12-54;
Ht1CryL9kJDKFabv 1; KAv7LjkFHTwpZbD3 1-2; L7tdvhcYjgwHTrMz 1; NtxKh7PgTAaZnHJd 1-4, 6-7, 9-11; PThANm6GzytHYDrd 1;
VMBZcafW2JLbT6dK 1-2; YwJVCfGj3hnFpRy1 1-2; fFNJ7HtK49hLDvCd 1-5, 7-9; gVHqdJ1Xwf3kBny9 1; jNVAY2Pav8GcbFWn 1-2;
kGjNAHY2Mqz1dfCJ 1-4, 7; m69XVCZpd7Dn8AyH 4-8; qgfYWFzjGtmJTpAX 1-7; rcM96QNt3CkxTWmg 1; twgy41aHZFz6DkA7 1, 3-17;
vH9PnJDV81wrthRj 2-5; wFjBdZtXxvm39yzq 1-3; wh2p6j3rZfHMaYzv 1-17, 19.

# EV-0157: Proc rates of Blackvenom Blade, Iceblade Hacker and Loksey's Training Stick from the 2026-11-19 sweep

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it. Iceblade Hacker's and Loksey's Training Stick's rates are measured and Blackvenom Blade's observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the three weapons; 248 reports in all
- Game build: 1.60.1.70862 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114281; the hotfix
  each player ran is not recorded
- Date: 2026-11-09 to 2026-11-19
- Pull request: #1126

This replaces the Iceblade Hacker rate EV-0136 gives and the Loksey's Training Stick rate EV-0153 gives.

## Method

The count is EV-0153's Method, run again by the automated sweep of 2026-11-19 over every wearer stream on disk, with the
same rolling hits, classes and forms left out, intervals, lone-unit checks, swap drop, credit windows and calls of
measured or observed.

Scripts and inputs, in `forever-sweep-auto`: `analyze.py` (version 12, as EV-0153's, with a paladin check that sets no
rate) for the rates, and the run `runs/2026-11-19-2309`, whose `results.json` holds each weapon's counts.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs;
Was is the row's rate before this count.

| Weapon                  |  Item | Rate | Call     | All hits [95%]    |   Procs / hits | Wearers | One-hand white | Specials   | Was |
| ----------------------- | ----: | ---: | -------- | ----------------- | -------------: | ------: | -------------- | ---------- | --: |
| Blackvenom Blade        |  4446 |  4.2 | observed | 4.22 [3.95, 4.37] |    412 / 3,258 |       5 | -              | 2.99 (28)  |   - |
| Iceblade Hacker         | 13952 | 2.05 | measured | 2.07 [1.94, 2.21] |   813 / 11,764 |       8 | 2.07 (43)      | 1.93 (70)  | 1.9 |
| Loksey's Training Stick |  7710 |  1.9 | measured | 1.92 [1.86, 1.99] | 3,520 / 35,405 |     189 | 2.13 (1,102)   | 1.94 (180) | 1.9 |

## Blackvenom Blade (item 4446, spell 13518): 4.2, observed

412 procs on 3,258 rolling hits, 5 wearers, 5 runs, in the main hand or the off hand: all hits 4.22 [3.95, 4.37].
Neither the one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate.
Observed: no lone-unit rate checks it. One-hand white: none. Dual-wield white: 4.14 [3.52, 4.57], 206 procs on 1,658
lone units, 5 wearers. Specials: 2.99 [2.90, 3.29], 28 procs on 312 lone units, 2 wearers. Direct: window [-10, 100] ms,
median 0 ms, 98th percentile 0 ms, 165 isolated hits. Reports from 2026-11-13 to 2026-11-19.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Rogue  |       4 |   367 | 2,875 | 4.26 |
| Hunter |       1 |    45 |   383 | 3.91 |

Reports and fights: APBKD3j6wYWG7xa4 8-11; FfAC3tWBxbp2HV9j 1-3; Ky8wY7p9bAcxTWvr 1-11; XBWhJzwTAYRPfQ4L 1-2, 4;
h8aHYk3QyJz2DNmt 1-2.

## Iceblade Hacker (item 13952, spell 1298416): 2.05, measured

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

## Loksey's Training Stick (item 7710, spell 1293482): 1.9, measured

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

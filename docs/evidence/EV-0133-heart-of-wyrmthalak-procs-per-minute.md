# EV-0133: Heart of Wyrmthalak procs per minute on melee and ranged hits

- Claim: Heart of Wyrmthalak (22321) casts Flame Lash 1.85 times a minute on landed melee and ranged hits, each hit on
  its own weapon's speed and Auto Shot and ranged specials on the ranged weapon's, with no cooldown; Volley never
  procs it, nor Shank up to 2026-11-19, and its damage is tripled against orcs.
- Sources: Warcraft Logs, the 7 players whose every gear snapshot shows the trinket in 813rPaLtcdpHy67q,
  Hwq97xLh6gv8ZPrY, KPwXNByAG2xQm9ra, NXMZV9xjvWpB7GtP (two players), bYaT9KXtGwqvV6d4 and dvkzFtPH3WR9C4Km, and the
  one hunter of KA9rqdFTWzt1mpxN; all fights
- Game build: 1.60.1.70938, as in EV-0131 for reports from 2026-11-13 on; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-16
- Pull request: #1061
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

## Method

Flame Lash (27655) lands 20 to 50 ms after the hit that procs it, so a proc counts for the latest landed melee or
ranged hit of the wearer on the same target within 100 ms before it. A landed hit is a hit, crit, glancing or blocked
hit. Left out as hits: poison, weapon enchant, Seal and damage shield hits, ticks, and Volley and Shank, whose rows
suppress weapon procs. Druids, whose forms swing at their own speeds, and the paladin, whose Seal hits land with the
swing, are left out.

Each kind of hit is priced as procs per minute on the speed of the weapon behind it: a one-weapon white swing and a
melee special on the main hand's, a dual wielder's white swing on both hands' mean chance by their swing share, and
Auto Shot and ranged specials on the ranged weapon's. KA9rqdFTWzt1mpxN has no gear snapshot; its hunter's ranged
weapon speed, 3.3 s, is read off its Auto Shots, which come every 2.9 s with no haste buff up behind a 13 to 15%
quiver. Each rate carries a 95% likelihood interval from the hits and one from 2,000 draws of whole wearers.

## Rates

| Hits                     | Wearers | Procs / landed hits | Rate     | 95%, hits    | 95%, wearers |
| ------------------------ | ------- | ------------------- | -------- | ------------ | ------------ |
| White swings, one weapon | 3       | 7 / 85              | 1.64 PPM | 0.72 to 3.04 | 0.50 to 3.16 |
| White swings, dual wield | 3       | 392 / 5,448         | 2.10 PPM | 1.90 to 2.29 | 2.06 to 2.14 |
| Melee specials           | 6       | 205 / 2,425         | 1.86 PPM | 1.63 to 2.12 | 1.77 to 1.95 |
| Auto Shot                | 3       | 295 / 2,820         | 2.06 PPM | 1.84 to 2.28 | 1.92 to 2.49 |
| Ranged specials          | 3       | 35 / 481            | 1.60 PPM | 1.15 to 2.17 | 1.42 to 1.90 |
| Volley                   | 2       | 0 / 2,941           | 0        |              |              |
| Shank                    | 4       | 0 / 593             | 0        |              |              |

A flat chance per hit fits worse: the best one, 8.3%, has a log likelihood of -3,219.3 against -3,201.7 for procs per
minute. Auto Shot alone shows the speed: 7.05% of landed shots at a 2.2 s bow, 12.9% at 3.1 s and 11.25% at 3.3 s.

Two procs land as little as 9 ms apart, and 54 of the 929 gaps between procs are under a second, so it has no
cooldown. Flame Lash's normal hits deal 111 to 218 (the row's 112 to 168 and the wearers' damage bonuses), and about
300 hits of 335 to 669 are the tripled damage against orcs.

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: equip proc 27656
- Rate: 1.85 PPM, observed

37,800 procs on 514,815 rolling hits, 183 wearers, 262 runs, on melee and ranged hits: all hits 1.87 [1.84, 1.90]. The
rate is all hits; the lone Auto Shot rate and the specials rate agree with it and the one-hand white rate reads above
its interval. Observed: the one-hand white rate disagrees. One-hand white: 2.02 [1.92, 2.12], 1,932 procs on 19,772 lone
units, 81 wearers. Dual-wield white, which sets no rate: 1.95 [1.91, 2.00], 6,648 procs on 104,432 lone units, 103
wearers. Lone Auto Shots: 1.99 [1.85, 2.10], 1,054 procs on 10,805 lone units, 88 wearers. Specials: 1.85 [1.74, 1.94],
2,759 procs on 33,504 lone units, 158 wearers; Sinister Strike gives 1,288 procs on 14,004 of them. Two-roll abilities,
priced as one roll: 2.16, 204 procs on 2,072 lone units (Execute 157 on 1,655 and Mongoose Bite 47 on 417), left out of
the rate. Direct: window [-9, 100] ms, median 31 ms, 98th percentile 48 ms, 8,025 isolated hits. Dropping the swapped
stretches (246 active minutes, 2% of hits, in 13 runs) moved the read from 1.83 to 1.87. The paladin check, which sets
no rate: lone white swings 1.92, 117 procs on 1,160 lone units, 21 wearers. Left out: 11 of 273 runs in Cat or Bear
Form. 4 runs logged in two reports count once. From 2026-11-20 the count takes each Shank as an off-hand roll (EV-0155
§Shank from 2026-11-20). Shield Slam, which the count takes as no roll, logs a proc after 276 of its 3,479 lone units,
1.09 times the same classes' white swing chance. Reports from 2026-11-10 to 2026-11-22.

| Class   | Wearers |  Procs |    Hits |  PPM |
| ------- | ------: | -----: | ------: | ---: |
| Rogue   |      67 | 15,928 | 241,756 | 1.89 |
| Warrior |      49 | 11,472 | 158,545 | 1.81 |
| Hunter  |      41 |  5,448 |  70,587 | 1.85 |
| Shaman  |      24 |  4,951 |  43,913 | 1.98 |
| Mage    |       2 |      1 |      14 | 2.03 |

Reports and fights: 123jYCvkVGrHfXWt 1-3, 20, 22-23, 27-30; 1JLq46yg7wYNhan9 17-28; 1TFWH4Yy38KhXMwx 6; 1ZyhaVDdFtTYWMcC
1-7, 9-12; 1b7PdYR4HhDXBGT6 1-8, 10-14; 27jZGhfHgp1LR8MB 8-11; 2H1VAy3MDrRndPvw 6; 2RPhk6zxZyLjgn4B 1-3;
2aZ6C1vRF7DAHMBn 2-3; 2mXywBCLNJHRW1xc 47-51; 2q1MandA4HXxL7Ym 1-3; 2wPQmVqbdCcZW9gv 234, 237, 240-243, 245-250;
3B8AcaGNRKmdgPDj 1-5; 3TdF61x7gMQ28pKX 4-35; 4D9aqWFLJVvn7bYx 1-2; 4Ktph9HWDPzj3c7v 1; 4Tc3yBNVYn7wW8RJ 28-31;
4VwxHWJ3pgNmvQY2 1-3; 4av7FfKdZjBh9Yp6 1-5; 6c8P4BLaVAj3bpK7 9; 6kZMyGqj2PhpLdT4 1-3; 72q1FtWgTMHAcLxG 1, 3-5;
7J3gB2PzrMykZQvh 5; 7ZWFDG4XBPKqf8zR 33-36; 7cpmjq6DTBNLwFhG 5-9, 11-21, 24, 26-28; 7dDbtWFhA8BPn4vZ 1-2;
82WLwYNzvjB3bXM6 1-2; 8Qp4nrZNKaFzy79G 107-109; 8c9BMJ3CT7VhWgGz 1-3; 8hBd9Wz3AXHwZf21 1-2; 8n6w19TGyhZtmKfd 1-6;
9C2qkgtRP8AWTJav 1-4; 9QDMdpt1bmcAaR4q 50; 9QKHCPLay1fMvtpc 1-6, 8; 9TtNqcjaBvHFYQ61 2-3; 9Zdv4NLm3FAwQKaV 1-6;
9hT126dqjwyx3g7r 1-5; 9nJGwhXTd8pvraxb 1-8; 9vntJM67VbardhHD 1; A1yxMTmKatkf3HRJ 22-44; AB9mrk6C8vathbVH 44, 46-47;
AfJ6DYbh9dmkgHZL 1-2; AyTmpNtxB9jbnJ2H 10-45; B7hYPwdmMb49fa2A 1-30; BDFj9VNrfXhxwQT3 3-4; BPVxqXWLc3FJTpMh 7-10, 12-14;
BYyqMbfxpZ8nFhJg 1-15; Btxk2TjrPZ7XYhwz 1; C8WHXv24kqZgKb9t 1; CbzZjqcYvGRAnh8p 2-4; CfHvwytKJ6P4q7ag 3-9;
CmZh47GDdxbMPRqg 3-6; CtMF9BgnAmWkTbfd 33; D46Xm8ncR9WbTFrL 32-71; D9frqPV8dam7B1AM 35-50, 53-55, 57, 59-61;
DAgp7nkzcLBXV3GM 1; DH3dT6tYLKmMrNz8 1-4, 6; Daj7G8xFX3bA2CmW 1-23; DdKN3mXj6gH9yvpc 32-71; DyMmJxc4t9pw6PAz 1-3;
FHqYCNcKntvDAjxg 22-31, 33-36; FN8krKfCb3BXWmwV 9-10; FNyng7tvQ1T2bra8 1-71; FhG3AxYL9ZNP4RvV 1-15; FrL4wayTHRtW1CKx
2-15; Ft1rJVHX3kz7yLBZ 1-3; Fz2wJmfGKPnvTy9h 6-8; G29y3QBWV7XDR6mj 11; G6tcKk2rJvQnA3gP 1-3; G7X1CkT8z6PrfjWM 146 fights
from 693 to 871; G9NmY7ARvHKyb8Cn 51, 53-54; GB6fVLkvzPhxYJcZ 1-2; GFBCh3VZH7TmAkL4 1-3; GRnMHmK69BXfWVTj 2-4;
GXR7nt6ZCL4DQzxP 1; GfMQw7CnvJZ64zRg 1-7; Gkb3rKHF6V9gz4pR 1-5; Gx9NyLPCjwDbV3qc 1; H2dZ3f1rTC4JyntF 1; HQtGDN3P9g1Vvawr
1; Ht1CryL9kJDKFabv 1; J38xbrFytgM9p1LT 4-19; JChQqKVgLWykANRr 9-46; JK6dgtv2N4QCqjnZ 1-2; JRQqNL6DkvWKYdPB 7, 9;
JWwXvLm1C9TgGykB 1-2; JavZtXcdp9WgLQ8x 1-32; JbQyngR8NWXKardc 60, 62-63; K3xmjRCnMpV2c4qr 40; KAv7LjkFHTwpZbD3 1-2;
Kphk63jtr1wb9NyL 1; LF63qnYybWDpCc12 1-4; LFNZrg2Cpty63BnK 3-10; LW1pRgPwCmXBqk9a 1-11, 13-27; MB1yTdQ9xmRXPaK8 1-12;
MBRmC1TnXq86k2Nb 1; MWrPDAVFCBXzh8Z6 8-9; NLftBVcrRhPwnv62 3-31; NPHc3azxL8bmgdpF 1-44; NRK61HZcDdJkrwXG 7-9;
NTJzVdMZBQDaqG83 7-8; NXMZV9xjvWpB7GtP 107, 109-116, 118-122; NdTXJzkWHLZrVQYm 1-3; Nkq1WZc8gATJRHnF 1-3;
NtK3hmqgfzAHj4PR 2-5; PtKXrwfj8zBHm2Gy 1-8; PtmYHjkDrqhdR9MA 6, 9-10; PwKXDAqJ4T7t8R9v 1-3; PwMavYLXV8GztqFd 1, 3;
PxgRcqaK9bdWjTGQ 2-5; Pz3XcdNk8xCRjaTA 78 fights from 11 to 97; QMLY8q2XNVry7bpj 5-23, 25-42; QNfaZ9MTnWpFrqb4 151
fights from 1 to 199; QNwkdxbMvWmrqjAY 1-3, 5-8; QZ3hRyAMbDrj9cHz 14; R6B94qfAjChQnrF1 3; RJXQC67jbzATWdx1 1-2, 5-9,
11-16; RXzv1k3V4PgaWhqj 1, 3-9; RkqGcKZ9TBv1NgpA 1; RwMv31La6FBZ9HDP 1, 3-24; T79gRkJz4QVGbyYW 6-7, 9-14;
TNak4bKY8JvLqXDF 1-3; TQYALMf39F2jWh6a 2; Tg1Hcb7z2qAXWkCx 1; Tx7MPH9f8Ga6hdrF 1-20; V7cnHk1QB3FA8WYG 7-8;
VGhX29mL6ra38qd7 1-3; VL7Yxg3ZmJkNFM1j 2-3; VMBZcafW2JLbT6dK 1-2; VMfN97kLXxyTpCwW 1-10; VTGmWXnbAK9Q1jHp 2-10;
VYBjgJx8aZtRhmpy 16-46; VmTj1LcnghRpxrMy 1, 3; VtcPT63wQ7WaACvj 1-12; WGbpC3Qz84JPa7jN 1; WNV2t9PaqjQGdBDc 39-43, 46-51;
WcCH7rnkQdaLZJ6j 3; WdbKG4RkmjxA7YV3 1-3; WjDJXCHz8kwxTpQA 1-5; WkrD6xYbcp8JZQBm 1-2; WqfyDhVHxT93Q7Pt 9-13, 25-30;
Wv4QYA6C7antxKXr 1; X8PH2CG1hMkcfxYJ 10-12; XGBaYAqPbcgZCVk8 7-15; XWaTVPCnmNAzD7FJ 12-13; XWfw2PzpNcBhbayY 5, 7-9;
Xjrt9FRMWqJyKdpa 2-3; YFtvD87kznr1CThg 6-7; YHRv2c3QqWZaLnmh 1-4; YZ6NdFDRQKL2CTmv 1-2; Yt9VaKGvf6Ry3BMW 1-6;
YwJVCfGj3hnFpRy1 1-2; Z1d98D2CFxk3cyh6 1-3; Z3MBhbDTXJ6tPCdQ 1-4; Z3nAqFVzTrmcHQWN 1-2; Z4VAmgDfJ9GPFjtc 1-6;
Z6kDfqLt9n38yY2a 1-2; ZaPqVJn1bQ894pxT 7-17, 19-52; ZfXrLy2MdY6kzah8 2-7; ZpNAtvBLPMKmf7YF 1; ZqtpzHLbAQKWMcwN 1-3;
Zz3YcQk69P7N8rAn 3-4; aFWqLhAB6XbKQf7D 2-3; aNh3Dt8kbZHL9cBM 1; apNB26JmYPyhFWjH 1-37, 39-89; b1Lmt96A8wRxzKQT 1-2;
b9B4aC82PXYypWLq 2-5, 7; bAJFX6m7ftYWCKHP 1-4; bDLrR2fCmJd3VgZY 8; bGTptRMWJyg4N7HQ 1-2; bP2CFzkZTLJ8q6Dn 1-8, 10-16;
bg169rGdXTzDhLnj 6-8; bnMvHkFjxQLJ4VRT 5; c4C7HaT2WdfFnJNx 29-30; c8mjK4TdzV2M7XLF 1-4; cKG7yDArxfqhJPp6 7-9;
cZ9B6Ay3RDTWgq4K 4-34; cbTt1zxajFfNVPDA 1-4; cgMGCdPJQ84wzaBV 3-8, 10-13; dAkztWXJP1jhrgcG 2-13, 15-17; dF3zbWfawRkZ2j6K
6-31; dHBxA1mTK3jwtR7q 1-2; daTq1vcjC3QBntb8 21-27, 29-42; dwXv9MZWLKGFJBry 2; fM4nQ2tKmpyWBLXj 18-24; fMPL97aCA1ngvKQk
16-20; fZdz8MyCHKFkrRwB 1, 3-7; fjPTxNBrbkJ6v3Q1 14-15; g79pDyr6TLWzVmhX 1-2; gQnavNZ3Wk4DPJGF 1-2, 4-9, 11-26;
gVHqdJ1Xwf3kBny9 1; gcxh69fvXWGtbRaV 1-6; gtJRYfmNcKLj9XTv 1; h9nNL4WmZfjHzvB6 1-2, 4-5; hJ3vY1FtLg2P9XQa 8;
hMdxA2bRGqr8ncvV 5-8; hQq8gvGjzPcWKF9A 1-6, 8-12, 14-15; j1Hzrkg6aRYdxvCT 2, 4-5; j1dmtfcYpLNTgKrD 8-9; jNRzcwDFf2nXYQTG
1-4; jNVAY2Pav8GcbFWn 1-2; jTYbp6KV9fnBD2mA 1-2; jWnC8Z7xDvpRcw46 1, 3-13; jXTVfy3ApWKPYwDG 1-22; jdRNZwApWYmKgLh7 1;
kG98jRqpFfBryv3c 2-4; kYX6FqpHthxnKcTM 1-10; m2cVkDw9XYAqnZ6g 1; m7FyYPBwvMngHbDk 6-11; mfA6MCQkPbFdRrz4 55-58;
nqDL2hKfQB4yvjgJ 1-18; nzcBtpgDa3dLrTmV 1; nzxZ4h87yGXJPgfp 1; pMvyNHtjKnkgR3X9 1-3; pNKJR39CwGnmXtbL 1;
pVaLq4TbGH2QFtfM 1-7; q4Rrpj73ZMQH96b2 1-5; qA612gBzDRyHt7mP 5-37, 41; qJX72LBYW1fAx6DP 1-3; qQG1nBdhztTcwHaW 1;
qbtZdThvBCXVcfAW 1-6; qmfjQ63TN81AnGkB 1-3; rAFM4QTbC8jHypZL 9-11; rfz4pR6mx1w9JWKA 1-2, 4-7; rhAc6qLKVn7JQ8GD 1-2;
tG4CqkJQvyPT2A1Y 1-10; tfVvJR3aZFgz2CXn 2-3; tg9xyavcrR2GqhLz 1-2; thqTyPArmkxfMnK7 5-36, 38; v8L6NkQTzMV2mWq3 1-8;
v8Zp2QF7An4d9xMj 1-10; vGJMahXtgTzFRmCQ 1-8, 10-14; vH9PnJDV81wrthRj 2-5; vHVNaQRrMdJmPpBC 1-3; vTW2MZKhFA9GcyaV 1-3;
vYjL3q712CyTGkQh 12-15; vwYF4Z2C38cfKTDB 3-9; wFjBdZtXxvm39yzq 1-3; wRWZtcVCk3QJKDa6 1, 3; wh2p6j3rZfHMaYzv 1-16;
wkrzY3mDh7RGfMc2 54; wtR76PQ8nxpANYkH 1-2; x6h24RZfAkVNDjJG 1-3; xGJcKTnzPmAr1aCw 1-4; xMLV8adPBDHYGF27 1-3;
xRK37aCvtVfj26kL 1-12; xXWkB3fqAjt9NJGH 5; xatzTjyc9GAh3MPC 1; xzHZV7MFgNhWj9n4 1; yTpdtkZDg87BqKwC 1-6;
yVp9KT3QkzhmZXDR 1; z8QPNWpMBx4VgdDX 1; zBx9XMQPdpW36a2D 1-4; zWQLAnHTjq4RJCby 1-2; ztQ3jnA8gdL6BKpF 1; ztyQjYFA8BT914fC
2-7, 9-12.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the trinket; 1,703 reports in all for the count, 256 of them for the trinket
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 416 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 87 of them for the trinket. The
  sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #PRNUM

The count is EV-0183 §Method, run on the trinket.

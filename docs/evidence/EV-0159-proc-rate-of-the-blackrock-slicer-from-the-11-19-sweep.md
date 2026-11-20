# EV-0159: Proc rate of The Blackrock Slicer from the 2026-11-19 sweep

- Claim: The Blackrock Slicer procs Wound at 1.9 PPM, priced on the weapon base speed of each hit that rolls it; the rate
  is measured.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights the weapon's section lists, from
  every wearer stream on disk for the weapon; 20 reports in all
- Game build: 1.60.1.70938 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114281; the hotfix
  each player ran is not recorded
- Date: 2026-11-13 to 2026-11-19
- Pull request: #1130

This replaces The Blackrock Slicer rate EV-0136 gives.

## Method

The count is EV-0157's: EV-0153's Method, run by the same sweep, `runs/2026-11-19-2309` in `forever-sweep-auto`, whose
`results.json` holds the weapon's counts.

## Rate

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs;
Was is the row's rate before this count.

| Weapon               |  Item | Rate | Call     | All hits [95%]    |   Procs / hits | Wearers | One-hand white | Specials   |  Was |
| -------------------- | ----: | ---: | -------- | ----------------- | -------------: | ------: | -------------- | ---------- | ---: |
| The Blackrock Slicer | 13285 |  1.9 | measured | 1.89 [1.81, 1.98] | 2,154 / 17,055 |      14 | 1.92 (395)     | 1.97 (107) | 2.05 |

## The Blackrock Slicer (item 13285, spell 17407): 1.9, measured

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

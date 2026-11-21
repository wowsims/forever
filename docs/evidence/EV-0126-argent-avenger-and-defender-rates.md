# EV-0126: Argent Avenger's and Argent Defender's proc rates

- Claim: Argent Avenger's 17352 procs 1.05 times a minute, priced on the weapon base speed of each hit that rolls it.
  Argent Defender's 17350 procs about 0.5 times a minute on the attacker's swing interval, rolled by each landed melee
  hit taken that deals damage outside its 20 s cooldown, and about twice as often when an Undead strikes. The owner
  rules 0.5 procs per minute on the attacker's swing interval for the shield, twice that when an Undead strikes.
- Sources: public reports on forever.warcraftlogs.com, the fights whose combatant info lists the item on the wearer.
  Argent Defender, where a fight with no gear record counts too: VjBTMHXxFvbthkw2 fight 1; NXMZV9xjvWpB7GtP fight 107;
  7tLY1pVK9wRWPABX fights 15 to 18 and 20; MWrPDAVFCBXzh8Z6 fights 1 to 8. Argent Avenger's section, which names its
  count, lists its own.
- Game build: 1.60.1.70938. Which hotfix push each report ran on and whether players relogged is not recorded. Argent
  Avenger's count gives its own.
- Date: 2026-11-13 07:09 to 2026-11-15 09:41 UTC; Argent Avenger's count gives its own
- Pull request: #1051, written up in #1052; Argent Avenger's count is #1156

## Argent Avenger

- Count: §Count of the 2026-11-06 to 2026-11-21 reports
- Rate: 1.05 PPM, measured

The game files state no rate: 17352 has no proc row, and the item's chance-on-hit effect names it.

5,000 procs on 128,616 rolling hits, 62 wearers, 92 runs, in the main hand or the off hand: all hits 1.06 [1.02, 1.09].
The rate is all hits, which the one-hand white rate and the specials rate agree with. One-hand white: 1.03 [0.92, 1.13],
615 procs on 16,253 lone units, 13 wearers. Dual-wield white: 0.92 [0.85, 0.98], 744 procs on 22,074 lone units, 51
wearers. Specials: 1.07 [0.97, 1.14], 432 procs on 11,057 lone units, 33 wearers. Two-roll abilities, priced as one
roll: 1.78, 19 procs on 291 lone units. Direct: window [-20, 100] ms, median -1 ms, 98th percentile 0 ms, 1,215 isolated
hits. Dropping the swapped stretches (138 active minutes, 3% of hits, in 14 runs) moved the read from 1.03 to 1.06. By
report date: before 2026-11-18 1.08 (1,354 procs); from 2026-11-18 to push 114313 1.05 (3,646 procs). Reports from
2026-11-14 to 2026-11-20.

| Class   | Wearers | Procs |   Hits |  PPM |
| ------- | ------: | ----: | -----: | ---: |
| Rogue   |      31 | 1,430 | 37,406 | 1.04 |
| Warrior |      23 | 3,459 | 88,691 | 1.06 |
| Hunter  |       8 |   111 |  2,519 | 1.20 |

Reports and fights: 1pJHLqQw297v3aF4 1-2; 1z4vRDqGdM2cVPXk 1-4; 38kdWHmLZKVgTPqb 1-2; 3HaPg2Yktxrnb7NL 1-9;
3qR2PmyNvQk8J4Bc 16-31, 33-55; 4av7FfKdZjBh9Yp6 1-7; 4hHBvakRnT7p1GVK 6-7; 4VwxHWJ3pgNmvQY2 1-3; 649MAP3BJxaLfj7K 1;
6CbKfDBm78nhyLjJ 1; 6hDJQWTXwBr419AM 1-4; 6kD7wbm3NaKW8RYP 1-7; 6mD4XcqZR2YnraGB 1; 6mpWRxtgXn8JhfkT 3; 7fy93LdcwQhvF2gP
8-9; 7fz9Kk8rY6jyQAbL 3-10; 8FBqfvLyWTdKxbZh 1-38; 8GFRaPBnKWf7Qrpb 1-3, 5-6; 8QHYRK6xTXB2gAtd 1; 8yPRmZdrq4QhcGn1 1-4;
93QNL64MnXKkxhvF 1-5; 9zxHaWbYT3gqDQvF 17-18; aANdrQRjKP7MJnCY 1-10; AvXNG1TpnyLJMxVt 1; B1qcTGpnHCrQ64Xa 1-4;
BHJMb9CGa1TgjtAd 1; BHvgJVTdZ9pfNPa6 1-3; bNZvk628gXLz7t3c 1; bQw4YVFNJZBt2THG 1; CgXG9bvFH6jJVZKk 1-4, 6-9, 11-15;
DCjMyH4x2m8Qaz3b 1-5; dzgfcTtLA1v7jDZ9 15; FN8krKfCb3BXWmwV 1-8; FrL4wayTHRtW1CKx 2-15; gAMnRkmdrKWzLJ4C 1-3;
gdBnpzCGaJ2THt3y 1; GM97YvP1hdgx8akn 1-2; gVHqdJ1Xwf3kBny9 1; GWMx4hcJFCYHR6dk 1; HfaAyZNt6rKWV7pY 1-2; Ht1CryL9kJDKFabv
1; JbQyngR8NWXKardc 3-4, 6-9, 11-16, 18-24, 27-44, 46-47, 49-60, 62-63; JctLn6NBFx2rwGjv 1-6, 8-11; JFv64qNc8yfWY9Qh
1-3; JKYV6CpmP2GA1f7Z 1-3, 5-13; jQGkxMF61cvmaLdB 1-4; K6VvbDypXFLQJW7d 1-27, 29; KAv7LjkFHTwpZbD3 1-2; kjbQpMhK81LqRgyB
1-3; kvCd8J41RrZ3zXNp 1, 3; LF9b7JDR36p284Hv 1; LQ9zMDqpmknBgAX8 1-2; m69XVCZpd7Dn8AyH 1-2, 4-8; m9rN6C1Z7RbWfx2K 1, 3;
mPTyvJ6XYKCMAdck 1; mW2TtC9YqHZz3vpn 1-2; MXHn319aGYhkAw7L 1; N3nHF2GBgdT6qx1V 1, 4; N9HJ4LQVFDYkXR7j 1-3;
NK6cCM4pJyRhaAFW 1-7; p4RKTftgHvzLrwXm 1-3; pG8TNJwybLPYWmxF 2; pnGKFx1vL82VWPy3 1; q31yFhx7TgpPfzGW 1; qJX72LBYW1fAx6DP
1-3; qMgtfcB2vrRdyma4 1; Qr9PmCnptRWMjN7k 1-28; qRL8HagchkrTVJmD 1-39; RHgxPaQbVWjDTc4K 1-7; rTKYMd4yDvaL6Gn8 1, 3;
T8JKMZh1WwyazjHY 2; Tdva3pHk4tmCfjZK 1-2; tG4CqkJQvyPT2A1Y 1-10; TLp1jza6PCrBAX94 1-7; vG3BpgdzWnmMrcfa 1-3;
VMBZcafW2JLbT6dK 1-2; wh2p6j3rZfHMaYzv 1-17, 19; Wn8wHhaxy31MLzjp 6-8; wqahKP7xGMNbc8mH 1-6; x2QH6gyDAb9ZrR3q 1, 3-7;
XD23HFnxbzvPjACV 1; xkKRJPqCLzBa9GdH 61-66, 68-89, 91, 93-96, 98-101; Yf79qHLZAjgnTdJw 1, 3-4; Z4VAmgDfJ9GPFjtc 1-6;
Z6Xm92K3zgfrqM84 1-4, 6-12; Zk1Gh4WA7KJDaMf2 1-2; ZRhnvMANGJDy1kBj 1; Zvr84V1Y3Gpg6n9H 5-20.

## Argent Defender

17350 states a ProcChance of 100, a ProcTypeMask of 0x28 (melee hits taken) and a 20 s ProcCategoryRecovery, and its
tooltip says "Has a chance when struck in combat"; its second effect, a dummy of 2, is the tooltip's "Undead that
strike you are 2 times as likely". The 100 in the column is not the chance.

A trial is a melee hit the wearer takes outside the cooldown that deals damage after block and absorb: a white swing,
a school-melee swing (logged as ability -64 or -16) or a melee special. Partial blocks count. A fight counts unless
its combatant info lists gear without the shield; a fight with no gear record counts. The cooldown runs from the hit
a proc is credited to.

Each 17351 application logs 0 to 33 ms after the hit that rolled it (median 9 ms, 8 at 0 ms), the opposite of the
Avenger's order. A proc goes to the last landed melee hit in the 60 ms before it, and all 86 have one. Seven windows
held hits from more than one attacker, two of them with a different speed or creature type; crediting the earliest
moves nothing past the third decimal. The shortest gap between two procs is 20.76 s, and 806 hits inside a cooldown
drew none.

| Wearer  | Dungeons                     | Procs | Trials | From Undead |
| ------- | ---------------------------- | ----- | ------ | ----------- |
| Paladin | Blackrock Spire              | 8     | 403    | 0           |
| Warrior | Blackrock Spire              | 9     | 280    | 0           |
| Paladin | Blackrock Spire, Scholomance | 46    | 2,420  | 710         |
| Paladin | Dire Maul                    | 23    | 1,655  | 234         |
| All     |                              | 86    | 4,758  | 944         |

What rolls, outside the cooldown:

| Hits taken                                     | Procs / hits | Rate |
| ---------------------------------------------- | ------------ | ---- |
| White swings, plain                            | 37 / 1,662   | 2.2% |
| White swings, partly absorbed                  | 17 / 802     | 2.1% |
| White swings, partly blocked                   | 27 / 2,051   | 1.3% |
| White swings, partly resisted or crits         | 0 / 18       | 0    |
| Melee specials that deal damage                | 5 / 273      | 1.8% |
| Melee hits blocked or absorbed whole           | 0 / 567      | 0    |
| Melee misses, dodges, parries and immunes      | 0 / 2,623    | 0    |
| Spells, ranged hits and other non-melee damage | 0 / 985      | 0    |

Counted as trials, the 554 hits blocked or absorbed whole that the models see would draw 9.6 procs (P ≈ 1e-4).
Partial blocks proc at about 0.6 times the rate of unblocked hits (Mantel-Haenszel odds ratio 0.57 within wearer and
attacker, p ≈ 0.03), for no reason found; they stay in the trials.

### Attackers

Creature types come from Wowhead's classic NPC data for 191 NPCs, since the Forever pages could not be read; every
trial NPC has the same type in all three data environments read. An attacker's speed is its time between its own
swings at the wearer, measured from the logs: the mode of the gaps between its swings, refined to the median of the
gaps within 12% of it, with swings 100 ms apart merged as extra attacks. 81 of the 106 NPCs with a clean speed swing
every 2.0 s. Swings within 30 s of the warrior's Thunder Clap slow from 2.0 to 2.4 s and are left out of the speeds.
16 NPCs dual-wield, told apart by hits that alternate high and low and by gap pairs that sum to one hand's speed.
The 48 trials whose attacker has no speed, one proc among them, are left out of the wearer table's trials and every
fit; the wearer table's procs and the table of what rolls count them.

### Fits

Each model is fitted by maximum likelihood over the trials: chance = c × speed^β × k^u, with u marking an Undead
attacker. Flat is β = 0 and procs per minute β = 1, with PPM = 60c. Intervals are profile-likelihood 95% intervals.
Speed is per hand here; a dual wielder's per-swing gap is below.

| Model          | logL    | AIC    | Estimate [95%]                            |
| -------------- | ------- | ------ | ----------------------------------------- |
| Flat, k = 2    | -426.50 | 855.00 | 1.49% [1.20, 1.81]                        |
| PPM, k = 2     | -425.36 | 852.73 | 0.476 [0.384, 0.580]                      |
| Flat, k free   | -425.43 | 854.85 | 1.65% [1.28, 2.07], k 1.41 [0.87, 2.23]   |
| PPM, k free    | -424.23 | 852.46 | 0.530 [0.411, 0.664], k 1.40 [0.85, 2.17] |
| Speed^β, k = 2 | -425.31 | 854.62 | β 0.82 [-0.20, 1.94]                      |

The logs barely tell flat from procs per minute: 68% of the trials come from 2.0 s attackers, where both give the same
chance. On simulated data with the same attackers, the likelihood picks the true model 83 to 84% of the time.

| Trials                                | Trials / procs | Flat [95%]         | PPM [95%]         |
| ------------------------------------- | -------------- | ------------------ | ----------------- |
| All                                   | 4,758 / 85     | 1.49% [1.20, 1.81] | 0.48 [0.38, 0.58] |
| Single-weapon attackers               | 4,069 / 83     | 1.74% [1.40, 2.12] | 0.55 [0.44, 0.67] |
| One attacker on the wearer in the 5 s | 1,418 / 39     | 2.22% [1.61, 2.97] | 0.69 [0.50, 0.91] |
| Unique NPCs                           | 469 / 15       | 2.47% [1.44, 3.89] | 0.81 [0.47, 1.27] |
| Trash, two or more attackers          | 3,244 / 45     | 1.18% [0.87, 1.55] | 0.38 [0.28, 0.50] |
| Dual wielders at their per-swing gap  | 4,758 / 85     | 1.49%              | 0.52 [0.42, 0.63] |

The Undead factor is 1.75 [1.06, 2.72] on single-weapon attackers and 1.92 [1.14, 3.15] on those with a clean speed,
in line with the row's 2. Over all attackers it reads 1.41 [0.87, 2.23], pulled down by two dual-wielding Undead that
procced none of 249 hits.

Expected procs under each fitted model, k = 2:

| Attacker speed | Attacker | Mean speed | Trials | Procs | Flat | PPM  |
| -------------- | -------- | ---------- | ------ | ----- | ---- | ---- |
| Under 1.3 s    | Other    | 1.11       | 382    | 5     | 5.7  | 3.4  |
| Under 1.3 s    | Undead   | 1.01       | 77     | 0     | 2.3  | 1.2  |
| 1.3 to 1.7 s   | Other    | 1.49       | 639    | 6     | 9.5  | 7.6  |
| 1.3 to 1.7 s   | Undead   | 1.37       | 61     | 4     | 1.8  | 1.3  |
| 1.7 to 2.2 s   | Other    | 2.00       | 2,467  | 42    | 36.7 | 39.1 |
| 1.7 to 2.2 s   | Undead   | 2.00       | 788    | 17    | 23.5 | 25.0 |
| 2.2 to 2.7 s   | Other    | 2.41       | 290    | 10    | 4.3  | 5.5  |
| 2.2 to 2.7 s   | Undead   | 2.40       | 18     | 1     | 0.5  | 0.7  |
| 2.7 s and over | Other    | 3.61       | 36     | 0     | 0.5  | 1.0  |

### What neither model explains

- Dual-wielding attackers drew 2 procs in 689 trials. Flat expects 14.0 (P ≤ 1e-4), procs per minute on each hand's
  speed 12.6 (P = 3e-4), and procs per minute on the attacker's per-swing gap 6.8 (P = 0.03). The last reading fits best
  of the speed models (β = 1.43 [0.62, 2.40], logL 6.4 above flat), so the speed that matters is the attacker's time
  between its own swings. The shortfall holds at every damage size and every density of incoming hits.
- The chance per hit falls as more attackers hit the wearer: 2.75% with one attacker in the 5 s before the hit, 1.54%
  with two, 1.32% with three and 1.17% with four or more. As n^η it fits η = -0.50 [-0.90, -0.12] with speed in the
  model, and -0.35 [-0.78, 0.08] within trash. The time since the wearer's previous hit hardly matters (δ = 0.15 [0.00,
  0.30]), which rules out a cap on the wearer's procs per minute. The cause is open.
- Per-wearer rates against one pooled rate give p ≈ 0.11 (flat) and 0.22 (procs per minute). The wearers' own weapon
  speeds run against their rates, which rules out the wearer's swing speed.

The owner rules 0.5 procs per minute on the attacker's swing interval, from the pooled fit, with the Undead factor of
2 the row states and the 20 s cooldown.

## Count of the 2026-11-06 to 2026-11-21 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 88 reports in all
- Game build: by report date, 1.60.1.70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC and 71037 from
  2026-11-20 18:23 UTC; no report starts after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is
  1.60.1.71037, push 114326; the hotfix each player ran is not recorded
- Date: 2026-11-14 to 2026-11-20
- Pull request: #1156

The count is EV-0153's Method on the hits EV-0155 finds rolling a weapon proc, run again over every wearer stream on
disk for its weapon, with more wearers fetched on 2026-11-21. The same rolling hits, classes and forms are left out:
paladins from every rate, druids in Cat or Bear Form, and the two-roll abilities with the procs around them. The
intervals, lone-unit checks, swap drop and credit windows are EV-0153's.

A rate is measured where a lone-unit rate of 50 or more procs (lone one-hand white swings or specials) overlaps the
all-hits interval, and 5 or more wearers of more than one class give an all-hits interval within 10% of the rate either
side. The rate is the all-hits rate rounded to the nearest 0.05. A lone-unit interval from under 5 wearers is a Poisson
interval.

Each section's read by report date is the all-hits rate over the reports that start before 2026-11-18, from then to push
114313, and after it.

Scripts and inputs, in `forever-notices-1121`: `results.json`, the counts of `forever-sweep-auto/analyze.py` (version
12, as EV-0157's, with the paladin check that sets no rate) run with `--items items.txt --census census_used.json`;
`report.py`, the call per weapon (`rep_final.json`); and `dates.py`, the report dates (`dates_final.json`).

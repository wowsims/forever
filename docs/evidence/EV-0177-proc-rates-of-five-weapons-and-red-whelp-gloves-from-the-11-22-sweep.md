# EV-0177: Proc rates of five weapons and Red Whelp Gloves from the 2026-11-22 sweep

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it; every rate is observed but Hurd Smasher's, which a later count measures. Red Whelp Gloves cast Firebolt on 10% of
  landed melee swings and damaging melee specials, the chance their row states.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each item's section lists, from
  every wearer stream on disk for the six items; 51 reports in all
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 21 reports start after hotfix push 114313 (2026-11-20 23:47 UTC). The sim's database is 1.60.1.71037, push
  114330; the hotfix each player ran is not recorded
- Date: 2026-11-09 to 2026-11-22
- Pull request: #1168
- Later count: Hookfang Shanker, Hurd Smasher and Siegebreaker's Blaster, §Count of the 2026-11-06 to 2026-11-22 reports
- Later count: Shimmering Platinum Warhammer, §Count of the 2026-11-13 to 2026-11-26 reports

## Method

The count is EV-0153's Method, run again by the automated sweep of 2026-11-22 over every wearer stream on disk, with the
same rolling hits, classes and forms left out, intervals, lone-unit checks, swap drop, credit windows and calls of
measured or observed. Each Mutilate strike rolls its own hand once (EV-0155); the count takes each Mutilate from hotfix
push 114313 (2026-11-20 23:47 UTC) as those rolls and leaves out the ones before it with their procs, as it does the
two-roll abilities. No two counted runs of one player overlap in time.

The run counts no Shank as a roll, though Shank rolls the off-hand weapon (EV-0155). The Lobotomizer's read counts the
off-hand wearers' Shanks from 2026-11-20 as off-hand rolls.

Scripts and inputs, in `forever-sweep-auto`: the run `runs/2026-11-22-1122`, whose `results.json` (`analyze.py` version
12, cache `cache_v13`) holds the counts of the count's weapons in EV-0136 and EV-0175, and whose cached records give Red
Whelp Gloves' lone units. The Lobotomizer's count comes from `analyze.py` version 14 (cache `cache_v14`), which counts
Shank's off-hand roll, on the same wearer streams. The rate is each result's `all_hits`, matched on item and
`effect_spell`.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon          |  Item | Rate | Call     | All hits [95%]    | Procs / hits | Wearers | One-hand white | Specials |
| --------------- | ----: | ---: | -------- | ----------------- | -----------: | ------: | -------------- | -------- |
| The Lobotomizer | 19324 | 0.55 | observed | 0.53 [0.49, 0.57] | 166 / 10,503 |       5 | -              | -        |

## Hookfang Shanker (item 11635, spell 13526): 1.95, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

625 procs on 13,792 rolling hits, 8 wearers, 11 runs, in the off hand: all hits 1.94 [1.81, 2.20]. Neither the one-hand
white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no lone-unit rate
checks it; every wearer is a rogue; the interval is wider than 10% of the rate. One-hand white: none. Dual-wield white,
which sets no rate: 1.97 [1.56, 2.38], 116 procs on 2,519 lone units, 8 wearers. Travel time: the proc lands a median
100 ms after its hit, 98th percentile 314 ms, window [-20, 324] ms (76 isolated hits). Dropping the swapped stretches
(36 active minutes, 5% of hits, in 2 runs) moved the read from 1.84 to 1.94. Reports from 2026-11-19 to 2026-11-22.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |       8 |   625 | 13,792 | 1.94 |

Reports and fights: 6TNJgAGhynFQqw2a 1-2; 74PAn1mMkWwYGcDd 1-44; Bz2NMjHnY7PD6ykX 1-5; CxbgQvRAL61MZdc4 1-4;
JHF4zftRn3PaBbN1 1-6; JPLDVkNF7Y43nyfp 1; LQfjhwmGWga9XBHZ 1-4; QwmphLYJXT7GazMd 9-40; cPGDLT3wyWYAxZkt 1;
yFGXLWVAKxw61gDP 1-4.

## Hurd Smasher (item 13198, spell 17308): 2, measured

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

577 procs on 9,669 rolling hits, 10 wearers, 10 runs, in the main hand or the off hand: all hits 1.99 [1.92, 2.06]. The
rate is all hits, which the one-hand white and specials rates agree with. One-hand white: 1.92 [1.50, 2.43], 70 procs on
1,214 lone units, 2 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.94 [1.61,
2.16], 188 procs on 3,228 lone units, 8 wearers. Specials: 2.46 [1.90, 3.13], 65 procs on 881 lone units, 2 wearers (a
Poisson interval: under 5 wearers); Cleave gives 30 procs on 484 of them. Direct: window [-10, 100] ms, median 0 ms,
98th percentile 1 ms, 221 isolated hits. Dropping the swapped stretches (27 active minutes, 7% of hits, in 2 runs) moved
the read from 1.85 to 1.99. Reports from 2026-11-16 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       7 |   377 | 6,237 | 2.01 |
| Warrior |       3 |   200 | 3,432 | 1.94 |

Reports and fights: 1TFWH4Yy38KhXMwx 6; 2ZLD7Jvwfa1n3xyM 1-2; ACJhv8TgVcbpRM3f 1-3; N8wyX9FR3TnPBpx2 1-8;
Tf1YmVratWLwcjkp 3-4; VqfFdmTwMzxrcH1N 1-2; daTq1vcjC3QBntb8 4-18, 20-27, 29-42; hqp4yWv6fY7c8PwV 2-3; q98JvNkWnKzyrwgM
1-6; qZTd9F3pkMLJcA6V 1-3.

## Shimmering Platinum Warhammer (item 15418, spell 19874): 2.05, observed

- Count: §Count of the 2026-11-13 to 2026-11-26 reports

924 procs on 8,797 rolling hits, 6 wearers, 9 runs, in the main hand: all hits 2.03 [1.91, 2.16]. The rate is all hits,
which the one-hand white rate agrees with. Observed: every wearer is a shaman. One-hand white: 1.94 [1.72, 2.48], 210
procs on 2,100 lone units, 6 wearers. Specials: none. The paladin check, which sets no rate: lone white swings 2.58
[1.33, 4.51], 12 procs on 90 lone units, 2 wearers (a Poisson interval: under 5 wearers), not apart from all hits;
paladins in at that rate, on as many hits per lone unit as the shamans', read 2.06, so the rate is 2.05 either way.
Travel time: the proc lands a median 112 ms after its hit, 98th percentile 250 ms, window [-20, 260] ms (183 isolated
hits). Left out: 1 of 10 runs in Cat or Bear Form. Reports from 2026-11-18 to 2026-11-26.

Across the 2026-11-23 hotfix: the reports before 2026-11-24 01:00 UTC read 1.91 [1.78, 2.11] (Poisson [1.75, 2.09]:
under 5 wearers), 513 procs on 5,187 rolling hits, 3 wearers, 5 runs; the reports from then on 2.20 [2.13, 2.28]
(Poisson [2.00, 2.43]: under 5 wearers), 411 procs on 3,610 rolling hits, 4 wearers, 4 runs. The analyzer run on the
reports from then on alone reads all hits 2.20 [2.13, 2.28], 411 procs, 4 wearers, which overlaps the whole count's 2.03
[1.91, 2.16]. The rate is the whole count's; the owner ruling of 2026-11-26 keeps it observed while the two halves read
apart.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Shaman |       6 |   924 | 8,797 | 2.03 |

Reports and fights: 68B7qNxMZLtRAh9j 1-2, 4; 6qtBhwN9grpafnDC 1-3; ARVPq7wnrdzpWQkH 9-18, 20; DvrjQXqZ9HN8a74z 1-4;
KNtHCAPd97ZMTLBf 1-2; WKkRPDJb8njrtQp1 1-2; jWfDCghpYJKy9G3N 1-8; rvPn71AMTwajKNdG 1; ycMCt3n9ZqP2bhaH 1-2, 4-8.

## Siegebreaker's Blaster (item 282561, spell 1314040): 5.25, observed

- Count: §Count of the 2026-11-06 to 2026-11-22 reports

1,532 procs on 6,727 rolling hits, 7 wearers, 8 runs, in the ranged slot: all hits 5.26 [4.90, 5.76]. Its lone-unit
rates of 50 or more procs read on a travel window past 400 ms, so none checks the all-hits rate. Observed: no lone-unit
rate checks it; every wearer is a hunter. Lone Auto Shots: 5.25 [4.87, 5.97], 488 procs on 2,147 lone units, 7 wearers.
Specials: 3.96 [2.68, 5.33], 30 procs on 175 lone units, 7 wearers. Travel time: the proc lands a median 70 ms after its
hit, 98th percentile 1656 ms, window [-20, 1666] ms (157 isolated hits). Dropping the swapped stretches (21 active
minutes, 8% of hits, in 2 runs) moved the read from 4.82 to 5.26. Reports from 2026-11-09 to 2026-11-22.

| Class  | Wearers | Procs |  Hits |  PPM |
| ------ | ------: | ----: | ----: | ---: |
| Hunter |       7 | 1,532 | 6,727 | 5.26 |

Reports and fights: 1LK3bmTvrNXhWcwk 1-20; 8ZBqNXCjPa7hKWdG 2; d1pRGTnZVvcfa7qK 1-8; pGBPWgMrL9KkQZRm 1-2;
pQ32h46zWCF9wZGj 1-2; prCfHbFPmDAgqd3V 3-4, 6, 8; qfay2cPntA6brzWg 1-18; z2FDkRa8TbNBZmLc 1, 3-9.

## The Lobotomizer (item 19324, spell 1290950): 0.55, observed

166 procs on 10,503 rolling hits, 5 wearers, 8 runs, in the main hand or the off hand: all hits 0.53 [0.49, 0.57], the
off-hand wearers' 966 Shanks from 2026-11-20 among them; the run, which counts no Shank, reads 0.58 [0.52, 0.62] on
9,537 hits. The off-hand runs read 0.53 (100 procs on 6,318 hits) and the main-hand runs 0.53 (66 on 4,185). Neither the
one-hand white swings nor the specials reach 50 procs, so no lone-unit rate checks the all-hits rate. Observed: no
lone-unit rate checks it; every wearer is a rogue. One-hand white: none. Dual-wield white, which sets no rate: 0.41
[0.27, 0.60], 38 procs on 3,118 lone units, 5 wearers. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 0 ms,
31 isolated hits. Reports from 2026-11-19 to 2026-11-22.

| Class | Wearers | Procs |   Hits |  PPM |
| ----- | ------: | ----: | -----: | ---: |
| Rogue |       5 |   166 | 10,503 | 0.53 |

Reports and fights: 2x7ZjvdbRwFMWKTP 3-45; 4rTfkGxMwKWPAXng 1, 3; GD26JWT9QRqhCyFg 3-36; GXR7nt6ZCL4DQzxP 1;
Gg9W1x7TFHNLyVck 1; WqfyDhVHxT93Q7Pt 25-30; ZfXrLy2MdY6kzah8 2-7; nrpTgLQdcNKAqk1R 1.

## Red Whelp Gloves (item 7284, spell 9233): 10%, the row's chance

The row 9233 states a 10% chance on landed melee swings and melee specials of either hand, and casts Firebolt 9057;
the tooltip reads 5%. A lone unit is a landed hit with no other event of the wearer within 100 ms before or after it,
and it has a proc when Firebolt logs from 6 to 100 ms after it. The intervals are per-hit Wilson intervals.

| Lone units                     | With a Firebolt / units | Share [95%]         | Wearers |
| ------------------------------ | ----------------------: | ------------------- | ------: |
| White swings, one hand         |             207 / 2,226 | 9.30% [8.16, 10.58] |       9 |
| White swings, dual wield       |             598 / 6,259 | 9.55% [8.85, 10.31] |       8 |
| Damaging melee specials        |             225 / 2,399 | 9.38% [8.28, 10.61] |      13 |
| All three                      |          1,030 / 10,884 | 9.46% [8.93, 10.03] |      16 |
| Sunder Armor and Rend landings |                 0 / 442 | 0.00% [0.00, 0.86]  |       4 |
| Ranged hits                    |               0 / 2,218 | 0.00% [0.00, 0.17]  |       8 |

The wearers are 5 rogues, 4 hunters, 4 warriors and 3 shamans; paladins and 7 of 27 runs in Cat or Bear Form, whose paw
swings read 9.79% (575 / 5,874), are left out. Reports from 2026-11-15 to 2026-11-22.

Reports and fights: 192ZN3dKFymLACY8 1; 1B6kfNZzCb4GqpFH 1-2; 4ChPnBrVK2fJLMaX 4; 8Xxb9LDGHP7vcJfh 1-5; KBzML9kdHFCXVpQc
1-2; Mj1gA4rpDRQh9yJm 1-3; QD8ypCXWJcv9Y2B3 6-11; ZGtjkf7naDdXLhbP 1-2; dP4ahnBM2fx8t3Zk 1-6; dVYAgntWbDqwPJvr 1-9;
k6Jrj8BbDqG1apW3 1-2; mzTCH7rgQnB6jqkV 1-10; n23GhKYqmPdwyg7A 1-5; qNyQvka41YhVmdHR 1-4; r2BYfmZW1wkLDdNx 1-2;
t2jHDfcvLkK96bdV 1-9; tcBj1NqHbxM7AaZp 1-4; wPrMvF2qYa61TCGB 1; y4fHqpA8Qt2m3Vaj 1.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each section naming this count
  lists, from every wearer stream on disk for its weapon; 3,714 reports in all for the count, 28 of them for these
  weapons
- Game build: 1.60.1.70862 to 1.60.1.71037 by report date (70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11
  02:19 UTC in the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 16 of them for these weapons.
  The sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on these weapons.

## Count of the 2026-11-13 to 2026-11-26 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights the Shimmering Platinum
  Warhammer section lists, from every wearer stream on disk for it; 353 reports in all for the count, 9 of them for
  Shimmering Platinum Warhammer
- Game build: 1.60.1.70938 to 1.60.1.71139 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU, 71139 from 2026-11-25 19:14 UTC in
  the US and the EU); 202 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 6 of them for
  Shimmering Platinum Warhammer, and 66 after 2026-11-24 01:00 UTC, 4 of them for Shimmering Platinum Warhammer. The
  sim's item database is 1.60.1.71139, push 114448; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-26
- Pull request: #1262

The count is EV-0221 §Method, run on this weapon.

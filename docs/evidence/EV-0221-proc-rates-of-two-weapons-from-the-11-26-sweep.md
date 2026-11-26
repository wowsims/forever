# EV-0221: Proc rates of two weapons from the 2026-11-26 sweep

- Claim: each weapon below procs at the rate its section gives, priced on the weapon base speed of each hit that rolls
  it; Alcor's Sunrazor's rate is measured, and The Judge's Gavel's observed.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights each weapon's section lists, from
  every wearer stream on disk for the two weapons; 18 reports in all for them, of the count's 353
- Game build: 1.60.1.70938 to 1.60.1.71139 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC, 71091 from 2026-11-24 15:21 UTC in the US and 2026-11-25 02:19 UTC in the EU, 71139 by 2026-11-25 22:38 UTC; its
  rollout time is not recorded); 202 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 12 of
  them for these weapons, and 66 after 2026-11-24 01:00 UTC, 9 of them for these weapons. The sim's item database is
  1.60.1.71139, push 114448; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-26
- Pull request: not recorded

## Method

The count is EV-0183 §Method, run again by the automated sweep of 2026-11-26 17:57 over every wearer stream on disk for
seven weapons: the two here, Electrified Dagger, whose home is EV-0183, Shimmering Platinum Warhammer, whose home is
EV-0177, and Felstriker, Shadowblade and The Cruel Hand of Timmy, whose home is EV-0136. Analyzer version 19 differs
from version 18 only on melee-table strike procs and on Torturing Poker, and touches none of these weapons: no proc row
of theirs has DefenseType 2, the melee table. Paladins count in a rate by EV-0155's and EV-0156's rolls wherever their
rate does not read apart from the other classes' (owner ruling, 2026-11-23; EV-0184). The sweep's counts leave paladins
out and keep their lone white swings of one-hand wearers as a check, which each weapon's section gives; a check's
"paladins in" read adds the paladins' lone units at their rate, on as many rolling hits per lone unit as the other
wearers'. The count makes no cut for the 2026-11-23 hotfix ("Fixed an issue where a few spells and abilities were
causing more weapon procs than intended"), where EV-0186 cut at 2026-11-24 01:00 UTC: 66 of its reports start after that
time.

Scripts and inputs, in `forever-sweep-auto`: the run `runs/2026-11-26-1757`, whose `results.json` (`analyze.py`
version 19, cache `cache_v18`) holds each weapon's count. The rate is each result's `all_hits`, matched on item and
`effect_spell` and rounded to the nearest 0.05 at full precision; the result's `ppm`, a lone-unit rate where one has
enough procs, sets no rate. A lone-unit rate from fewer than 5 wearers takes a Poisson interval on its procs.

Each section also splits the count at 2026-11-24 01:00 UTC, by report start: each side's read is its runs' procs over
their expected procs at 1 PPM, with an interval of 2,000 bootstrap draws over runs, and a Poisson interval beside it
under 5 wearers. The analyzer run again on the reports from that time on alone (`analyze_posthotfix.py`, the run
`runs/2026-11-26-posthotfix`) gives each of the 103 items with 100 or more procs and 5 or more wearers there an all-hits
interval that overlaps the whole count's, so the hotfix moved no rate and each weapon takes the whole count's.
Shadowblade's and Shimmering Platinum Warhammer's halves read apart, and by owner ruling of 2026-11-26 both stay
observed.

## Rates

Rate is the override row's rate. The one-hand white and specials columns give the read and, in parentheses, the procs.

| Weapon            |  Item | Rate | Call     | All hits [95%]    | Procs / hits | Wearers | One-hand white | Specials  |
| ----------------- | ----: | ---: | -------- | ----------------- | -----------: | ------: | -------------- | --------- |
| Alcor's Sunrazor  | 14555 | 1.95 | measured | 1.95 [1.84, 1.99] | 595 / 14,084 |       6 | 1.92 (52)      | 2.52 (43) |
| The Judge's Gavel | 12528 | 1.05 | observed | 1.06 [1.00, 1.19] |  318 / 5,271 |       6 | 1.05 (95)      | 1.00 (9)  |

## Alcor's Sunrazor (item 14555, spell 18833): 1.95, measured

595 procs on 14,084 rolling hits, 6 wearers, 12 runs, in the main hand or the off hand: all hits 1.95 [1.84, 1.99]. The
rate is all hits, which the one-hand white rate agrees with. One-hand white: 1.92 [1.43, 2.52], 52 procs on 1,251 lone
units, 3 wearers (a Poisson interval: under 5 wearers). Dual-wield white, which sets no rate: 1.90 [1.57, 2.28], 117
procs on 2,842 lone units, 3 wearers (a Poisson interval: under 5 wearers). Specials: 2.52 [1.83, 3.40], 43 procs on 787
lone units, 4 wearers (a Poisson interval: under 5 wearers). Two-roll abilities, priced as one roll: 0.00, 0 procs on 3
lone units (Execute), left out of the rate. No paladin wears it in the count. Direct: window [-10, 100] ms, median 0 ms,
98th percentile 1 ms, 120 isolated hits. Reports from 2026-11-18 to 2026-11-26.

Across the 2026-11-23 hotfix, which sets no rate: the reports before 2026-11-24 01:00 UTC read 1.94 [1.78, 2.09]
(Poisson [1.72, 2.19]: under 5 wearers), 264 procs, 4 wearers; the reports from then on 1.96 [1.81, 2.15] (Poisson
[1.75, 2.18]: under 5 wearers), 331 procs, 3 wearers.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Rogue   |       2 |   345 | 8,064 | 1.97 |
| Warrior |       4 |   250 | 6,020 | 1.92 |

Reports and fights: 83B9JpGFmVtkThWL 2; 9wm8cNnWv7p2DCbV 1-3; CxbgQvRAL61MZdc4 4; HYgxLwn9ZfqyVBN7 1; HvKhpmJaD63Ytzgk
1; KL9mjTYDkwMztf8r 2-5; KZAyNqgzHQ8VwtdB 3-8; M1n84TjqH3admxtc 1; PafbczYWBMjgHCG7 77; TgQJ7t3wHrzPqM4C 2;
W2d6nPMQ9tV1BjZx 2-3; pQxYvP8khVb7ZMwB 1-3.

## The Judge's Gavel (item 12528, spell 56): 1.05, observed

318 procs on 5,271 rolling hits, 6 wearers, 6 runs, in the main hand: all hits 1.06 [1.00, 1.19]. The rate is all hits,
which the one-hand white rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand white: 1.05
[0.96, 1.13], 95 procs on 1,599 lone units, 6 wearers. Specials: 1.00 [0.46, 1.90], 9 procs on 159 lone units, 3 wearers
(a Poisson interval: under 5 wearers). Two-roll abilities, priced as one roll: 3.28, 8 procs on 43 lone units (Execute),
left out of the rate. Its 2 paladin wearers in the count give no lone-unit rate, so the count holds no paladin check.
Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 99 isolated hits. Left out: 3 of 9 runs in Cat or Bear
Form. The sim does not simulate the stun (the generator lists it as a missing effect); the rate is recorded on the
item's effect. Reports from 2026-11-20 to 2026-11-26.

Across the 2026-11-23 hotfix, which sets no rate: the reports before 2026-11-24 01:00 UTC read 1.00 [1.00, 1.01]
(Poisson [0.86, 1.15]: under 5 wearers), 188 procs, 3 wearers; the reports from then on 1.17 [1.04, 1.44] (Poisson
[0.98, 1.39]: under 5 wearers), 130 procs, 3 wearers.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |       5 |   266 | 4,349 | 1.08 |
| Warrior |       1 |    52 |   922 | 1.00 |

Reports and fights: 8GHAfMvNmbRrWQKj 1; 8ghPD6kpwR4LCxc1 1-4; TRBpca3wfFJAVyQh 1; bNfrjh3dkPDBc6QC 1-6; jDmbJA49cXdayRqv
1; vLa1NzdYWTJbjc7V 1-18, 20-33, 35-42.

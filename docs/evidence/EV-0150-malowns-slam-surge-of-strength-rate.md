# EV-0150: Malown's Slam puts Surge of Strength on the wearer with each proc, about 13% of landed weapon hits

- Claim: Malown's Slam 13393 procs on landed white swings and melee specials alike, at 2.0 procs per minute, about 13%
  of each landed hit on its 3.8 s speed, with no cooldown. Each proc casts 17500 on the target and puts
  Surge of Strength 17499 on the wearer, a stun the target is immune to or a killing hit included. A proc never rolls
  another, and spells roll none.
- Sources: Warcraft Logs reports (forever.warcraftlogs.com) HYmf1vTW6hVDPytK fights 15 and 16, NXMZV9xjvWpB7GtP
  fight 159, ZfbY6RgWjx9hmpGz fight 13 and pNW7aGKnzHP1MftT fights 1, 3 and 4, an Enhancement shaman in each; combat
  log, a level 60 Retribution paladin in Stratholme
- Game build: the reports 1.60.1.70938 (ZfbY6RgWjx9hmpGz, pNW7aGKnzHP1MftT and NXMZV9xjvWpB7GtP) and 1.60.1.70959
  (HYmf1vTW6hVDPytK), dated by when the CDN first listed each build against the log times; the combat log 1.60.1.70959
  by the same dating. The hotfix each log ran on is not recorded
- Date: the reports 2026-11-14 16:07 to 2026-11-18 05:08 UTC; the combat log 2026-11-18
- Pull request: #1101
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

## Method

The game files' on-hit spell for 13393 is 17500, which holds the stun alone and states no proc chance or rate.
Surge of Strength 17499 holds the Strength; no effect of 17500 casts it. A proc is an apply or refresh of 17499 on
the wearer.

A hit is one of the wearer's own damage lines that hit, crit or glanced. A hit is isolated when the wearer logs no
other hit within 60 ms of it (50 ms in the combat log). In the reports a hit has a proc when one logs from 20 ms before
it to 40 ms after it; in the combat log 17499 logs 1 ms before its hit and 17500 1 ms after, and a hit has a proc
from 20 ms before it to 5 ms after. The same window shifted 0.3 to 0.9 s later counts chance coincidence.

Two other shamans wore the mace for a few swings only and a druid tank swapped weapons through the log, so they are
left out. Every wearer counted swung the 3.8 s mace, so these hits fit a flat 13% chance as well as a rate in procs
per minute.

## Reports

Isolated landed hits:

| Hit            | With a proc / hits | Share | Shifted window |
| -------------- | ------------------ | ----: | -------------: |
| White swing    | 79 / 664           | 11.9% |           1.5% |
| Stormstrike    | 32 / 260           | 12.3% |           3.1% |
| Earth Shock    | 0 / 237            |     0 |                |
| Lightning Bolt | 0 / 143            |     0 |                |

By report, white swings and Stormstrikes:

| Report           | White swing | Stormstrike |
| ---------------- | ----------- | ----------- |
| HYmf1vTW6hVDPytK | 3 / 24      | 7 / 31      |
| NXMZV9xjvWpB7GtP | 22 / 162    | 7 / 67      |
| ZfbY6RgWjx9hmpGz | 27 / 249    | 9 / 78      |
| pNW7aGKnzHP1MftT | 27 / 229    | 9 / 84      |

Of 73 procs on a white swing with no other hit of the wearer within 150 ms, none has a second proc within 150 ms
after it, nor do the left-out druid's 53.

## Combat log

The paladin wore the mace from the Magistrate Barthilas fight on. 86 procs cast 17500 (65 applied, 10 refreshed, 11 the
target was immune to), and each of the 11 immune lines has a 17499 at its millisecond. 3 more procs logged 17499 with
no 17500, each on a hit that killed its target.

Isolated landed hits:

| Hit                  | With a proc / hits | Share |
| -------------------- | ------------------ | ----: |
| White swing          | 19 / 205           |  9.3% |
| Holy Strike          | 7 / 125            |  5.6% |
| Judgement of Command | 17 / 117           | 14.5% |

The count takes Judgement of Command as two rolls a cast. Two procs logged at one millisecond, and the next shortest
gap between procs on separate hits was 0.94 s. Every pair of procs within 200 ms is a white swing and a Seal of Command
strike.

The paladin's specials fall short of the rate: 2.0 procs per minute expects 16 of the 125 Holy Strikes (7 seen, a 0.8%
tail) and 28 of the 117 Judgements at two rolls (17 seen, 1%). The claim that specials roll it alike rests on the
reports' Stormstrikes.

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 17500
- Rate: 2 PPM, measured

1,175 procs on 9,244 rolling hits, 11 wearers, 12 runs, in the main hand: all hits 2.01 [1.88, 2.14]. The rate is all
hits, which the one-hand white rate agrees with. One-hand white: 2.10 [1.89, 2.34], 327 procs on 2,461 lone units, 11
wearers. Specials: 1.82 [1.13, 2.78], 21 procs on 182 lone units, 2 wearers (a Poisson interval: under 5 wearers).
Two-roll abilities, priced as one roll: 2.20, 11 procs on 79 lone units (Execute 11 on 79), left out of the rate.
Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 325 isolated hits. The paladin check, which sets no
rate: lone white swings 1.36, 43 procs on 499 lone units, 8 wearers. Left out: 1 of 13 runs in Cat or Bear Form. Reports
from 2026-11-14 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Shaman  |       9 |   925 | 7,201 | 2.03 |
| Warrior |       2 |   250 | 2,043 | 1.93 |

Reports and fights: HZP6h9AtRNK4n3Ck 1-2; XHCPnJGcKfZrhQYA 1-2; ZfbY6RgWjx9hmpGz 13-14; cZ9B6Ay3RDTWgq4K 4-24;
jx76r4NWPZGXDaBp 1-3, 5; rTKYMd4yDvaL6Gn8 1, 3; tfdhHTZx8B6azF24 1-3; tg9xyavcrR2GqhLz 1-2; vHVNaQRrMdJmPpBC 1-3;
vHp9QagZGFLBM7PC 1-2; xt4kjHLMnFvPCrhT 1-2; yjd7A18ghwWmqYzD 1, 3-4.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the mace; 3,714 reports in all for the count, 12 of them for the mace
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 748 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 5 of them for the mace. The
  sim's database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on the mace.

# EV-0150: Malown's Slam puts Surge of Strength on the wearer with each proc, about 12% of landed weapon hits

- Claim: Malown's Slam 13393 procs on landed white swings and melee specials alike, at about 12% of each landed hit
  on its 3.8 s speed (about 1.8 procs per minute), with no cooldown. Each proc casts 17500 on the target and puts
  Surge of Strength 17499 on the wearer, a stun the target is immune to or a killing hit included. A proc never rolls
  another, and spells roll none.
- Sources: Warcraft Logs reports (forever.warcraftlogs.com) HYmf1vTW6hVDPytK fights 15 and 16, NXMZV9xjvWpB7GtP
  fight 159, ZfbY6RgWjx9hmpGz fight 13 and pNW7aGKnzHP1MftT fights 1, 3 and 4, an Enhancement shaman in each; combat
  log, a level 60 Retribution paladin in Stratholme
- Game build: the reports 1.60.1.70938 (ZfbY6RgWjx9hmpGz, pNW7aGKnzHP1MftT and NXMZV9xjvWpB7GtP) and 1.60.1.70959
  (HYmf1vTW6hVDPytK), dated by when the CDN first listed each build against the log times; the combat log 1.60.1.70959
  by the same dating. The hotfix each log ran on is not recorded
- Date: the reports 2026-11-14 16:07 to 2026-11-18 05:08 UTC; the combat log 2026-11-18
- Pull request: not opened yet

## Method

The game files' on-hit spell for 13393 is 17500, which holds the stun alone and states no proc chance or rate.
Surge of Strength 17499 holds the Strength; no effect of 17500 casts it. A proc is an apply or refresh of 17499 on
the wearer.

A hit is one of the wearer's own damage lines that hit, crit or glanced. A hit is isolated when the wearer logs no
other hit within 60 ms of it (50 ms in the combat log). In the reports a hit has a proc when one logs from 20 ms before
it to 40 ms after it; in the combat log 17499 logs 1 ms before its hit and 17500 1 ms after, and a hit has a proc
from 20 ms before it to 5 ms after. The same window shifted 0.3 to 0.9 s later counts chance coincidence.

Two other shamans wore the mace for a few swings only and a druid tank swapped weapons through the log, so they are
left out. Every wearer counted swung the 3.8 s mace, so these hits fit a flat 12% chance as well as a rate in procs
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

Judgement of Command is two rolls a cast. Two procs logged at one millisecond, and the next shortest gap between
procs on separate hits was 0.94 s. Every pair of procs within 200 ms is a white swing and a Seal of Command strike.

## Rate

The white swings of the reports and the combat log pooled are 98 of 869 hits, 11.3%, which is 1.78 procs per minute
on the 3.8 s speed; the reports alone give 1.88. The owner set 1.8.

# EV-0178: Proc rate of Blackhand Doomsaw

- Claim: Blackhand Doomsaw (item 12583) casts Fatal Wound (16549) at 1.0 procs per minute, priced on the weapon base
  speed (3.5) of each hit that rolls it; the rate is observed. A feral druid's paw swings read about the same rate,
  0.94, when each is priced on the paw's speed (Cat Form 1.0, Dire Bear Form 2.5).
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the druid reports and fights listed below; one combat
  log of a level 60 feral druid, 2026-11-20, Blackrock Spire
- Game build: Warcraft Logs reports 1.60.1.70888 to 1.60.1.71037 by report date (70938 from 2026-11-13 01:36 UTC,
  70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23 UTC); the hotfix each player ran is not recorded. The
  combat log ran 1.60.1.70959 on hotfix push 114281, relogged
- Date: 2026-11-11 to 2026-11-22
- Pull request: #1172
- Later count: the rate, §Count of the 2026-11-06 to 2026-11-22 reports

## Method

The Warcraft Logs count is EV-0153's Method, run again by the automated sweep of 2026-11-22 (`forever-sweep-auto`,
`runs/2026-11-22-1122/results.json`, the run behind EV-0177) over every wearer stream on disk: 15 runs, of which 6 are
feral druids and are left out of the rate, since they swing the paw. Fatal Wound has no flight (`SpellMisc.Speed` 0) and
lands a median 0 ms and a 98th percentile 1 ms after its hit (63 isolated procs), so it is credited from 11 ms before to
100 ms after. The 16549 row has no `SpellAuraOptions` entry, so the game files state no rate.

A separate script counts the druids. It splits each druid's landed white swings by form: in the Warcraft Logs streams by
Cat Form (768) and Dire Bear Form (9634) applies and removes, in the combat log by the power type each swing logs
(energy for Cat Form, rage for Dire Bear Form). It keeps only the wearer's own events and drops procs from 20 ms before
to 50 ms after a killing blow; none fell there in either count. In the combat log, Warchief Rend Blackhand, who drops
Doomsaw, dies at 21:57:14 local time, the first proc lands at 22:04:30 and the gear first shows Doomsaw at 22:05:38.
Counted from 21:57:14 the combat log reads 0.83; counted from 22:04:00 it reads 0.90, and the tables below use that
start. 2 of its 31 procs missed and are counted. One proc, in Warcraft Logs, logs 16 ms before the next landed white
swing, outside its window; it is counted in the pooled rates and has no form. Every other proc has a landed white swing
in its window. Of the 8,712 swings, 5 are in caster form, with no proc.

## Rates

| Count                           | Wearers | Procs |  Hits | Priced on | PPM [95%]         |
| ------------------------------- | ------: | ----: | ----: | --------- | ----------------- |
| Druid paw swings, Warcraft Logs |       6 |   151 | 7,054 | paw       | 0.95              |
| Druid paw swings, combat log    |       1 |    31 | 1,658 | paw       | 0.83 to 0.90      |
| Druid paw swings, both          |       7 |   182 | 8,712 | paw       | 0.94 [0.82, 1.10] |

The intervals are bootstraps over wearers, and the combat log's range comes from its two start times.

The druid swings by form, priced both ways:

| Form           | Procs | Landed swings | Chance per swing | PPM on paw speed | PPM on weapon speed |
| -------------- | ----: | ------------: | ---------------: | ---------------: | ------------------: |
| Cat Form       |   106 |         6,796 |            1.56% |             0.94 |                0.27 |
| Dire Bear Form |    75 |         1,911 |            3.92% |             0.94 |                0.67 |

Reports and fights of the druids: cvRTFb1qhGBKNkCD 45-53; cZ9B6Ay3RDTWgq4K 4-42; f6HGnWcxNgk7MZJa 1; PrkmRY4gWJtvTf6D 1,
3; r4tJg3fYcBKnQkPx 1-6; X8PH2CG1hMkcfxYJ 10-12.

## Rate

- Count: §Count of the 2026-11-06 to 2026-11-22 reports
- Proc: weapon proc 16549
- Rate: 1 PPM, observed

249 procs on 4,376 rolling hits, 8 wearers, 10 runs, in the main hand: all hits 0.98 [0.87, 1.16]. The rate is all hits,
which the one-hand white rate agrees with. Observed: the interval is wider than 10% of the rate. One-hand white: 1.12
[0.93, 1.54], 56 procs on 860 lone units, 8 wearers. Specials: 1.42 [1.09, 2.04], 32 procs on 385 lone units, 7 wearers.
Two-roll abilities, priced as one roll: 1.62, 12 procs on 127 lone units (Execute 10 on 83 and Mongoose Bite 2 on 44),
left out of the rate. Direct: window [-11, 100] ms, median 0 ms, 98th percentile 1 ms, 89 isolated hits. Dropping the
swapped stretches (25 active minutes, 17% of hits, in 1 run) moved the read from 0.81 to 0.98. The paladin check, which
sets no rate: lone white swings 1.11, 146 procs on 2,245 lone units, 27 wearers. Left out: 6 of 16 runs in Cat or Bear
Form. Reports from 2026-11-16 to 2026-11-22.

| Class   | Wearers | Procs |  Hits |  PPM |
| ------- | ------: | ----: | ----: | ---: |
| Warrior |       3 |   178 | 3,315 | 0.92 |
| Hunter  |       5 |    71 | 1,061 | 1.15 |

Reports and fights: 17tkGC3naHzD2X8h 1-2; 2jJ1L8FPHzTB4fKG 10-12; 782JNDc1pgGRZxVC 1-5; APMHcT7wXdVn8FfL 1-2;
QXamf2TH1q43Arjw 1; T9Vjh2FWfNAxXb1J 1; aZKTvVcCHLwhgXzM 6-7; d69BJcNG4mwWThna 1-2; k3ZGwNbF9W7DgzLQ 1-3;
q4Rrpj73ZMQH96b2 1-5.

## Count of the 2026-11-06 to 2026-11-22 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights §Rate lists, from every wearer
  stream on disk for the axe; 3,637 reports in all for the count, 10 of them for the axe
- Game build: 1.60.1.70938 to 1.60.1.71037 by report date (70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23
  UTC); 735 reports of the count start after hotfix push 114313 (2026-11-20 23:47 UTC), 2 of them for the axe. The sim's
  database is 1.60.1.71037, push 114332; the hotfix each player ran is not recorded
- Date: 2026-11-06 to 2026-11-22
- Pull request: #1180

The count is EV-0183 §Method, run on the axe.

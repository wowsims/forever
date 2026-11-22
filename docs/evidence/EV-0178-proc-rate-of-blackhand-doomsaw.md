# EV-0178: Proc rate of Blackhand Doomsaw

- Claim: Blackhand Doomsaw (item 12583) casts Fatal Wound (16549) at 0.9 procs per minute, priced on the weapon base
  speed (3.5) of each hit that rolls it; the rate is observed. A feral druid's paw swings read the same rate when each
  is priced on the paw's speed (Cat Form 1.0, Dire Bear Form 2.5).
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights listed below and, for the
  paladin check, the run's paladin wearer streams; one combat log of a level 60 feral druid, 2026-11-20, Blackrock
  Spire
- Game build: Warcraft Logs reports 1.60.1.70888 to 1.60.1.71037 by report date (70938 from 2026-11-13 01:36 UTC,
  70959 from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23 UTC); the hotfix each player ran is not recorded. The
  combat log ran 1.60.1.70959 on hotfix push 114281, relogged
- Date: 2026-11-11 to 2026-11-22
- Pull request: #1172

## Method

The Warcraft Logs count is EV-0153's Method, run again by the automated sweep of 2026-11-22 (`forever-sweep-auto`,
`runs/2026-11-22-1122/results.json`, the run behind EV-0177) over every wearer stream on disk: 15 runs, of which 6 are
feral druids and are left out of the rate, since they swing the paw. Fatal Wound has no flight (`SpellMisc.Speed` 0) and
lands a median 0 ms and a 98th percentile 1 ms after its hit (63 isolated procs), so it is credited from 11 ms before to
100 ms after. Dropping the swapped stretches (25 active minutes, 18% of hits, in 1 run) moved the all-hits read from
0.75 to 0.91. The 16549 row has no `SpellAuraOptions` entry, so the game files state no rate.

A separate script counts the druids. It splits each druid's landed white swings by form: in the Warcraft Logs streams by
Cat Form (768) and Dire Bear Form (9634) applies and removes, in the combat log by the power type each swing logs
(energy for Cat Form, rage for Dire Bear Form). It keeps only the wearer's own events and drops procs from 20 ms before
to 50 ms after a killing blow; none fell there in either count. In the combat log, Warchief Rend Blackhand, who drops
Doomsaw, dies at 21:57:14 local time, the first proc lands at 22:04:30 and the gear first shows Doomsaw at 22:05:38.
Counted from 21:57:14 the combat log reads 0.83; counted from 22:04:00 it reads 0.90, and the tables below use that
start. 2 of its 31 procs missed and are counted. One proc, in Warcraft Logs, logs 53 ms before its swing; it is counted
in the pooled rates and has no form. Every other proc has a landed white swing in its window. Of the 8,712 swings, 5 are
in caster form, with no proc.

## Rates

| Count                           | Wearers | Procs |  Hits | Priced on    | PPM [95%]         |
| ------------------------------- | ------: | ----: | ----: | ------------ | ----------------- |
| All hits, no druids (Rate)      |       7 |   211 | 3,971 | weapon (3.5) | 0.91 [0.81, 1.14] |
| One-hand white                  |       7 |    42 |   676 | weapon (3.5) | 1.07 [0.90, 1.77] |
| Specials                        |       6 |    21 |   239 | weapon (3.5) | 1.51 [1.02, 2.16] |
| Paladin check (one-hand white)  |      26 |   137 | 2,074 | weapon (3.5) | 1.13 [0.83, 1.37] |
| Druid paw swings, Warcraft Logs |       6 |   151 | 7,054 | paw          | 0.95              |
| Druid paw swings, combat log    |       1 |    31 | 1,658 | paw          | 0.83 to 0.90      |
| Druid paw swings, both          |       7 |   182 | 8,712 | paw          | 0.94 [0.82, 1.10] |

The intervals are bootstraps over wearers, and the combat log's range comes from its two start times. The one-hand white
and specials rates hold fewer than 50 procs each, so neither checks the all-hits rate, and the all-hits interval reaches
past 10% of the rate on both sides (11% below, 25% above); the 0.9 rate stays observed.

By class, no druids: 4 hunter runs, 36 procs on 530 hits (1.16); 5 warrior runs, 175 procs on 3,441 hits (0.87).

The druid swings by form, priced both ways:

| Form           | Procs | Landed swings | Chance per swing | PPM on paw speed | PPM on weapon speed |
| -------------- | ----: | ------------: | ---------------: | ---------------: | ------------------: |
| Cat Form       |   106 |         6,796 |            1.56% |             0.94 |                0.27 |
| Dire Bear Form |    75 |         1,911 |            3.92% |             0.94 |                0.67 |

Reports and fights, no druids: 17tkGC3naHzD2X8h 1-2; 2jJ1L8FPHzTB4fKG 10-12; 782JNDc1pgGRZxVC 1-5; aZKTvVcCHLwhgXzM 6-7;
d69BJcNG4mwWThna 1-2; k3ZGwNbF9W7DgzLQ 1-3; q4Rrpj73ZMQH96b2 1-5; QXamf2TH1q43Arjw 1; T9Vjh2FWfNAxXb1J 1. Druids:
cvRTFb1qhGBKNkCD 45-53; cZ9B6Ay3RDTWgq4K 4-42; f6HGnWcxNgk7MZJa 1; PrkmRY4gWJtvTf6D 1, 3; r4tJg3fYcBKnQkPx 1-6;
X8PH2CG1hMkcfxYJ 10-12.

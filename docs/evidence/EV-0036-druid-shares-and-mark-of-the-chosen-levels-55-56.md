# EV-0036: Rip Rank 5, Rake Rank 4 and Pounce Bleed Rank 3 shares and Mark of the Chosen, druid levels 55 to 56

- Claim: Rip Rank 5 takes 0.0114 of attack power a tick per combo point up to 4, Rake Rank 4 0.065 on the hit and the
  tick, Pounce Bleed Rank 3 0.025 a tick and Ferocious Bite Rank 5 0.037 per combo point; Mark of the Chosen (item
  17774, trigger 21969) procs only on the melee, ranged and harmful spell hits taken, at about the row's 2%, and no
  helpful spell taken fired it.
- Sources: combat logs and StatLedger of a druid at levels 55 to 56
- Game build: 1.60.1.70888
- Date: 2026-11-12
- Pull request: #987
- Later count: Ferocious Bite Rank 5, §Count of the 2026-11-13 to 2026-11-21 logs

## Attack power shares

| Spell                                | Value  | Fit                                                                                                                              |
| ------------------------------------ | ------ | -------------------------------------------------------------------------------------------------------------------------------- |
| Rip Rank 5 (9894), periodic          | 0.0114 | 190 of 191 ticks, k 0.011365 to 0.011405; 109 of the 191 ticks were 5-point Rips and the share on all 5 points fits none of them |
| Rake Rank 4 (9904), direct           | 0.065  | 121 of 121 hits, c 0.06498 to 0.06524                                                                                            |
| Rake Rank 4 (9904), periodic         | 0.065  | 237 of 238 ticks                                                                                                                 |
| Pounce Bleed Rank 3 (9826), periodic | 0.025  | 24 of 24 ticks, c 0.0248 to 0.0258                                                                                               |

## Mark of the Chosen

The row's ProcTypeMask also names helpful spells taken, but every proc in the logs followed a hit taken.

| Reading                             | Measured                                                           |
| ----------------------------------- | ------------------------------------------------------------------ |
| Procs on eligible landed hits taken | 5 procs on 392 landed hits taken, 1.28%, 95% interval 0.55 to 2.95 |
| Procs off a helpful spell taken     | none                                                               |

## Count of the 2026-11-13 to 2026-11-21 logs

- Sources: raw combat logs of six cat druids at level 60, from the recordings of 2026-11-13, 2026-11-19, 2026-11-20 and
  2026-11-21: 228 Ferocious Bites
- Game build: by event time, 1.60.1.70938 for the US log of 2026-11-13 (22:26 UTC to 00:12 UTC on 2026-11-14),
  1.60.1.70959 for the EU logs of 2026-11-19 (15:26 to 16:49 UTC) and 2026-11-20 (08:46 to 10:15 UTC), and
  1.60.1.71037 for the EU log of 2026-11-21 (10:26 to 14:34 UTC) (70938 from 2026-11-13 01:36 UTC, 70959 from
  2026-11-17 15:19 UTC in the US and 2026-11-18 02:04 UTC in the EU, 71037 from 2026-11-20 18:23 UTC); the hotfix each
  player ran is not recorded
- Date: 2026-11-13 to 2026-11-21
- Pull request: #1236

Each druid is calibrated and its lines are kept as in EV-0210's method, and a Ferocious Bite under Clearcasting is left
out.

- Ferocious Bite 22829: floor((87 x (1 +- 0.35714) + 160 x combo points + 3.1 + k x attack power x combo points) x
  Naturalist x Rend and Tear), the attack power from the cast line. 3.1 is the row's 310 / 100 extra damage; 2.5 fits
  worse.

| Reading                          | Bites fitting                   |
| -------------------------------- | ------------------------------- |
| k 0.035                          | 193 of 228                      |
| k 0.036                          | 207 of 228                      |
| k 0.037                          | 218 of 228                      |
| k 0.038                          | 207 of 228                      |
| Best band, 0.00001 grid          | 0.03698 to 0.03709 (218 of 228) |
| Mean-roll estimate, uniform roll | 0.03719 [0.03652, 0.03786]      |

Per druid at 0.037: 3 of 3, 59 of 60, 126 of 134, 17 of 18, 8 of 8 and 5 of 5. The bites had 1 to 5 combo points,
110 of them 5. The mean-roll estimate is the sum of (logged base - 87 - 160 x combo points - 3.1) over the sum of
attack power x combo points, with the roll uniform over 87 x (1 +- 0.35714).

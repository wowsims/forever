# EV-0210: Rip Rank 6 and Ferocious Bite Rank 5 attack power shares, cat druids at level 60

- Claim: Rip Rank 6 takes 0.0114 of attack power a tick per combo point up to 4, not 0.01, and Ferocious Bite Rank 5
  takes 0.037 of attack power per combo point.
- Sources: raw combat logs of cat druids at level 60, from the recordings of 2026-11-13, 2026-11-19, 2026-11-20 and
  2026-11-21: four druids for Rip (76 ticks), six for Ferocious Bite (228 bites)
- Game build: 1.60.1.70938 for the logs of 2026-11-13 to 2026-11-20 and 1.60.1.71037 for those of 2026-11-21, by log
  date; the hotfix each player ran is not recorded
- Date: 2026-11-13 to 2026-11-21
- Pull request: not yet opened

No attack power ledger exists for these druids, so each druid is calibrated first from rules already measured (EV-0036).
The cat paw's offset in the log's attack power is a multiple of 7, and the druid's Naturalist, Savage Fury, Rend and
Tear and Genesis ranks come from fitting white swings and Rake hits and ticks; a druid whose best Rake fit is not unique
within 2% is left out. One druid's offset (Alterac Valley, a wide white-swing spread) is not confirmed by its whites, so
its ticks are also read apart from the pool. Only clean lines count: no resist, block, absorb, glance or crush, creature
targets only, and Ferocious Bite under Clearcasting is left out.

- Rip 9896: floor((19 + 31.9 x combo points + k x attack power x min(combo points, 4)) x Naturalist x Rend and Tear x
  Genesis), the attack power from the druid's newest own block within 3 s before the tick.
- Ferocious Bite 22829: floor((87 x (1 +- 0.35714) + 160 x combo points + 3.1 + k x attack power x combo points) x
  Naturalist x Rend and Tear), the attack power from the cast line. 3.1 is the row's 310 / 100 extra damage; 2.5 fits
  worse.

## Rip Rank 6

| Reading                           | Ticks fitting      |
| --------------------------------- | ------------------ |
| k 0.0114                          | 74 of 76           |
| k 0.01, the tester's sheet        | 0 of 76            |
| Best band, 0.00001 grid           | 0.01138 to 0.01140 |
| Intersection of the fitting ticks | [0.01137, 0.01141) |

Per druid at 0.0114: 34 of 35, 23 of 24, 5 of 5 and 12 of 12; at 0.01 none fits in any of them. The two misfits are
one tick each of the first two druids. 67 of the 76 ticks were 5-point Rips and 9 were 4-point. The druid with both
(3 and 9 ticks) separates the cap: its 9 ticks below 5 points read [0.01129, 0.01147), its 3 5-point ticks read 0.0091
if all 5 points counted and fit 0.0114 capped at 4. The pool without that druid (64 ticks) reads 62 of 64 at 0.0114 with
the same intersection.

## Ferocious Bite Rank 5

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

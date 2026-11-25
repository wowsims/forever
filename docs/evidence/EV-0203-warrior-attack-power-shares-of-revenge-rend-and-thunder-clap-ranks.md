# EV-0203: Attack power shares of Revenge Ranks 1, 2, 3 and 5, Rend Ranks 1, 3 and 6 and Thunder Clap Ranks 2, 3, 5 and 6

- Claim: Revenge Ranks 1, 2, 3 and 5 take 0.25 of the sheet's attack power in their damage (the stance and Improved Revenge
  multiply the sum, and 0.28 fits none of them); Rend Ranks 1, 3 and 6 take 0.02 of the attack power at each tick; Thunder
  Clap Ranks 2, 3, 5 and 6 take 0.07 of it in a damage the hit floors.
- Sources: combat logs of a Protection warrior at levels 14 to 38 and 38 to 39, with the addon's sheet ledger for the attack
  power of each hit (the log's attack power less the main hand's and the temporary enchant's offset), and combat logs of a
  second Protection warrior at levels 45 to 59 (Thunder Clap Ranks 5 and 6, Revenge Rank 5, Rend Rank 6); report codes not
  recorded
- Game build: 1.60.1.70790 (the Rend Rank 1 logs), 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938, each count's section
  naming its own; the hotfix each log ran on is not recorded
- Date: 2026-11-05 to 2026-11-14
- Pull request: not recorded

Each hit's damage is fitted against the sheet's attack power. A share counts as measured when the free slope's 95%
interval holds the value, or when the damage the row predicts matches the logged amount (a residual centred on 0 or an
exact match of the floored amount). Rend ticks read the attack power at the tick. The flat physical damage bonus applies
to Revenge and not to Thunder Clap or Rend ticks. Revenge's damage is `floor(0.9 x (1 + 0.2 x Improved Revenge rank) x (roll +
0.25 x AP + flat))` in Defensive Stance, the roll a whole number, and Thunder Clap's is `floor(multiplier x (base + 0.07 x AP))`.

## Revenge

| Rank      | Levels   | Date and build                                   | Slope [95%]             | Residual against the row | Hits |
| --------- | -------- | ------------------------------------------------ | ----------------------- | ------------------------ | ---- |
| 1 (6572)  | 20 to 24 | 2026-11-07, 70862                                | 0.2488 [0.2458, 0.2518] | -0.08 +- 0.11            | 447  |
| 2 (6574)  | 24 to 27 | 2026-11-08, 70862                                | 0.2514 [0.2445, 0.2582] | -0.10 +- 0.18            | 508  |
| 2         | 27 to 33 | 2026-11-09 to 2026-11-13, 70862 to 70938         | 0.2505 [0.2477, 0.2533] | +0.01 +- 0.13            | 957  |
| 3 (7379)  | 27 to 38 | 2026-11-09 to 2026-11-13, 70862 to 70938         | 0.2528 [0.2408, 0.2647] | -0.03 +- 0.24            | 570  |
| 3         | 38 to 39 | 2026-11-13 to 2026-11-14, 70938                  | 0.2535 [0.2473, 0.2598] | +0.27 +- 0.24            | 639  |
| 5 (11601) | 58 to 59 | 2026-11-13, 70888 and 70938 (split not recorded) | 0.2434 [0.2326, 0.2541] | not recorded             | 441  |

0.28 is rejected on Ranks 1, 2 and 3. The logs that read 0.28 (EV-0049) predate the tooltip's statement of 0.25.
The Rank 1 hits are the 447 EV-0045 fits the roll on. A Rank 1 count on 2026-11-08 (70862, 158 hits) reads +0.12 +- 0.16. Revenge Rank 4 is
in EV-0048.

## Rend

| Rank      | Levels                        | Date and build                                        | Reading                                                                                                                | Ticks     |
| --------- | ----------------------------- | ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | --------- |
| 1 (772)   | 14 to 19                      | 2026-11-05, 70790                                     | residual -0.07 +- 0.07 (10:09 to 10:30) and +0.04 +- 0.09 (19:41 to 20:54), 5 + 0.02 x AP                              | 92 and 98 |
| 3 (6547)  | 24 to 27                      | 2026-11-08, 70862                                     | residual +0.011 +- 0.028 against 9 + 0.02 x AP; crits x2.000 +- 0.005                                                  | 1,197     |
| 6 (11573) | not recorded, within 45 to 59 | 2026-11-09 to 2026-11-13, split by build not recorded | 1,159 of 1,176 ticks equal `floor(multiplier x (18 + 0.02 x AP at the tick))`; the attack power at the cast fits 1,022 | 1,176     |

Rank 1's attack power rests on an offset bounded by Thunder Clap Rank 1 that day. Rend Rank 7 (11574) is not measured.

## Thunder Clap

| Rank      | Levels                        | Date and build                                                  | Reading                                                                                            | Hits  |
| --------- | ----------------------------- | --------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ----- |
| 2 (8198)  | 24 to 27                      | 2026-11-08, 70862                                               | 88 of 89 exact against `floor(multiplier x (23 + 0.07 x AP))`, no crit in 89                       | 89    |
| 3 (8204)  | 27 to 38                      | 2026-11-09 to 2026-11-13, 70862 (108), 70888 (344), 70938 (300) | 752 of 752 exact against `floor(multiplier x (37 + 0.07 x AP))`; slope 0.0696 [0.0691, 0.0701]     | 752   |
| 5 (11580) | 48 to 59                      | 2026-11-10 to 2026-11-13, 70888 and 70938 (split not recorded)  | 1,346 of 1,363 exact against `floor(multiplier x (82 + 0.07 x AP))`; slope 0.0700 [0.0695, 0.0705] | 1,363 |
| 6 (11581) | not recorded, within 45 to 59 | 2026-11-10 to 2026-11-13, split by build not recorded           | 148 of 152 exact against `floor(multiplier x (103 + 0.07 x AP))`                                   | 152   |

Thunder Clap Rank 4 is in EV-0048.

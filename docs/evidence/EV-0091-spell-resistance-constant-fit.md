# EV-0091: The spell resistance constant fits 8L + 19

- Claim: the constant K in a = R / (R + K), L being the caster's level, fits 8L + 19; ExpectedStat's ArmorConstant × 0.088 fits nearly as well.
- Sources: Spell Piercing logs from level 20 to 50 and mobs near level 38 casting at players of known resistance (a tracking issue, #76, holds the per-level data and charts); at level 49 to 51 the logs of a mage with 15 Spell Piercing
- Game build: not recorded; every log is from before the 2026-11-19 hotfix (1.60.1.71037, hotfixes 114309, 114310, 114312, 114313 and 114314) after which binary spells take the Spell Piercing bonus (EV-0142)
- Date: 2026-11-11 and 2026-11-12 for the level 49 to 51 logs, the others not recorded; all before the 2026-11-19 hotfix
- Pull request: #961

The mage at levels 49 to 51 with 15 Spell Piercing had a bonus on 1,500 of 5,756 direct hits, which is 7.5a, so a = 0.0347 and K = 417 (399 to 435). 9L gives 450 there, 3.7 sigma off.

| Level                      | Measured K       | 8L + 19                 | 9L   | ArmorConstant × 0.088 |
| -------------------------- | ---------------- | ----------------------- | ---- | --------------------- |
| 20                         | 179 ± 18         | 179                     | 180  | 185                   |
| 24 and up                  | 209 ± 5          | 211                     | 216  | 215                   |
| 29                         | 251 ± 10         | 251                     | 261  | 252                   |
| 30 (pooled)                | 270 (247 to 299) | 259                     | 270  | 260                   |
| about 38 (mobs on players) | 315 (285 to 350) | 323                     | 342  | 319                   |
| 50                         | 417 (399 to 435) | 419                     | 450  | 409                   |
| chi-squared                |                  | 1.1 (fitted 7.97L + 19) | 18.9 | 2.9                   |

At level 60 the two candidates give 499 and 484, which at 15 piercing is 21.9% against 22.5% of hits binned with a bonus, so a level 60 log tells them apart only with some 15,000 direct hits.

# EV-0091: The spell resistance constant fits 8L + 19

- Claim: the constant K in a = R / (R + K), L being the caster's level, fits 8L + 19; ExpectedStat's ArmorConstant × 0.088 fits nearly as well, and at level 60 the reports do not tell the two apart.
- Sources: Spell Piercing logs from level 20 to 50 and mobs near level 38 casting at players of known resistance (a tracking issue, #76, holds the per-level data and charts); at level 49 to 51 the logs of a mage with 15 Spell Piercing
- Game build: not recorded; every log is from before the 2026-11-19 hotfix (1.60.1.71037, hotfixes 114309, 114310, 114312, 114313 and 114314) after which binary spells take the Spell Piercing bonus (EV-0142)
- Date: 2026-11-11 and 2026-11-12 for the level 49 to 51 logs, the others not recorded; all before the 2026-11-19 hotfix
- Pull request: #961
- Later count: K at level 60, §Count of the 2026-11-19 to 2026-11-27 reports

The mage at levels 49 to 51 with 15 Spell Piercing had a bonus on 1,500 of 5,756 direct hits, which is 7.5a, so a = 0.0347 and K = 417 (399 to 435). 9L gives 450 there, 3.7 sigma off. The 5,756 hits are those of the spells that took a bonus then: Arcane Explosion 2,284, Arcane Missiles 2,208, Arcane Blast 844, Fire Blast 337 and Shoot 83. The binary spells, which took none before the hotfix (Arcane Wrath 0 of 554, Flamestrike 0 of 506, Cone of Cold 0 of 80, Frostbolt 0 of 67), are left out, so the hotfix does not touch the fit.

| Level                      | Measured K       | 8L + 19                 | 9L   | ArmorConstant × 0.088 |
| -------------------------- | ---------------- | ----------------------- | ---- | --------------------- |
| 20                         | 179 ± 18         | 179                     | 180  | 185                   |
| 24 and up                  | 209 ± 5          | 211                     | 216  | 215                   |
| 29                         | 251 ± 10         | 251                     | 261  | 252                   |
| 30 (pooled)                | 270 (247 to 299) | 259                     | 270  | 260                   |
| about 38 (mobs on players) | 315 (285 to 350) | 323                     | 342  | 319                   |
| 50                         | 417 (399 to 435) | 419                     | 450  | 409                   |
| chi-squared                |                  | 1.1 (fitted 7.97L + 19) | 18.9 | 2.9                   |

## K at level 60

- Count: §Count of the 2026-11-19 to 2026-11-27 reports

At level 60 the candidates give 499 (8L + 19), 484 (ArmorConstant × 0.088) and 540 (9L). Mages without Arcane Subtlety,
by their gear piercing p from 6 to 21, read K = 509.5 (504.5 to 514.0) over 256,570 hits of Arcane Explosion and
Blizzard, but their 15 groups scatter more than their counts allow (chi-squared 289 on 14 degrees of freedom), which
widens it to 487 to 530. Mages with the talent at rank 2 read 499 if it gives 14 piercing and 522 if it gives 15. Read
against the median of the caster's other hits instead of the hit before resists, the same mages give 551 (545 to 557);
the hit before resists is the reading a combat log confirmed one to one (EV-0092).

8L + 19 lies inside the spread, ArmorConstant × 0.088 at its edge and 9L outside it. The 3% between 499 and 484 is under
the scatter and under the talent's 14 or 15, so the reports do not separate them. A level 60 combat log, whose negative
resisted amount reads every bonus, at known gear piercing tells 499 from 484 at P 0.01 on each side with about 60,600
direct hits at 20 piercing or 86,400 at 15, and 499 from 9L's 540 with about 9,700 at 20.

## Count of the 2026-11-19 to 2026-11-27 reports

- Sources: Warcraft Logs reports of Forever (zone 3000) started 2026-11-19 23:03 to 2026-11-27 05:50 UTC, as EV-0092
  lists them; the level 60 mages among them
- Game build: as EV-0092's: 1.60.1.70959 or 1.60.1.71037 until 2026-11-20 18:23 UTC, 1.60.1.71037 until 2026-11-24
  15:21 UTC (US) and 2026-11-25 02:19 UTC (EU), 1.60.1.71091 until 2026-11-25 19:14 UTC and 1.60.1.71139 from then. The
  hotfix a run was on and relogs are not recorded
- Date: 2026-11-19 23:03 to 2026-11-27 05:50 UTC
- Pull request: #1267

A bonus is a hit with nothing resisted whose damage is 1.1, 1.2 or 1.3 times the hit before resists (within 2.5%), on
mobs whose plain direct hits of the school show 3% or fewer partial resists; the rate of mages whose own bonus rate is
under 4% (1.45%) is taken off. K comes from the rate r as p (7.5 / r − 1), and over the groups from a maximum
likelihood fit; level and gear piercing are those of the hit's fight. Arcane Subtlety's rank comes from the fight's
talents.

Scripts and tables: `forever-piercing-posthotfix-1127` (RESULTS.md, `tables/kfit_abs.txt`, `tables/kfit.txt`).

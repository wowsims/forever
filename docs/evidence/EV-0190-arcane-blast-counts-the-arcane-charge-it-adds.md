# EV-0190: Arcane Blast counts the Arcane Charge it adds

- Claim: Arcane Blast (1239700) adds its Arcane Charge (400573) before its damage, so the hit counts the spell damage of
  that charge; its mana cost reads the charges held before the cast, 182 × (1 + charges), times 1.3 under Arcane Power,
  and 0 under Clearcasting.
- Sources: combat log `WoWCombatLog-112426_191754`, logged by a rogue in a level 60 Scholomance group, runs 2 and 3;
  the hits and costs of one level 60 Orc Arcane mage in that group (Arcane Mind 5, Arcane Instability 3, Arcane
  Concentration 5, Missile Barrage 2, Precision 3)
- Game build: 1.60.1.71091; hotfix push not recorded; relog unknown
- Date: 2026-11-24 19:17 to 23:25 UTC−5 (2026-11-25 00:17 to 04:25 UTC)
- Pull request: not recorded

## Order in the log

Each Arcane Blast that adds a charge logs the 400573 apply or dose first, then the cast line, then the damage.

## Damage

Only Arcane Blast hits with a `HIT` outcome count, and those cast under Arcane Power or Blood Fury are left out. The
damage is the hit's base amount, before the crit and any partial resist. The expected base is
(236 + 0.429 × spell power) × 1.03 Arcane Instability. The spell power field on the log line leaves charges out, so
each charge adds 12% of the mage's 353 Intellect (42.4); one charge is 0.429 × 42.4 × 1.03 = 18.7 damage. A hit
counts as adding a charge when a 400573 apply or dose falls in the 5 ms before it, and as at 4 charges when only a
refresh does.

Log base minus the expected base when the expected base counts only the charges held before the cast:

| Hits                                       |   n | Log − expected, mean [95% interval] |
| ------------------------------------------ | --: | ----------------------------------- |
| Added a charge                             | 198 | +20.7 [18.7, 22.6]                  |
| At 4 charges (refresh only)                |   5 | −0.7 [−8.8, 7.4]                    |
| No 400573 event in the 5 ms before the hit |   9 | −3.8 [−13.1, 5.5]                   |

The gap on hits that add a charge is one charge's 18.7. Counting that charge, log over expected is 1.004
[0.999, 1.009] on the same 212 hits. Over all 252 hits, with Arcane Power (×1.3 damage) and Blood Fury (×1.1 spell
power) included, it is 1.006 [1.001, 1.011], and 1.059 [1.054, 1.064] without the charge.

## Cost

The mana paid is the mage's mana on the Arcane Blast's cast line minus the mana on the next line carrying the mage's
own unit info within 5 ms. Base cost is 182, 15% of 1213 base mana. Mana drops by charges after the cast:

| Charges after the cast | No Arcane Power   | Arcane Power | Clearcasting, with or without Arcane Power |
| ---------------------: | ----------------- | ------------ | ------------------------------------------ |
|                      1 | 182 (39)          | 237 (6)      | 0 (3), −1 (1)                              |
|                      2 | 364 (23)          | 473 (3)      | 0 (4), −1 (1)                              |
|                      3 | 546 (15)          | 710 (4)      | 0 (3)                                      |
|                      4 | 728 (13), 910 (2) | 946 (3)      | 0 (1)                                      |

Counts are in brackets. A cast at 4 charges keeps 4, so the 910 drops are casts with 4 charges before, 182 × 5.

# EV-0077: Lash of Pain at level 30 shows no Sayaad multiplier

- Claim: Lash of Pain deals (16 + 0.429 × Spell Damage) × 1.10 under 5/5 Unholy Power, with no multiplier of the Sayaad's own.
- Sources: a combat log and sheets of a level 30 warlock (5/5 Unholy Power, +10% demon damage in Forever) with a level 30 Sayaad; report code not recorded
- Game build: not recorded
- Date: 2026-10-13
- Pull request: #418

| Reading                                  | Value                                                                                                     |
| ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| Warlock sheet Spell Damage               | 108 (Spell Healing 115)                                                                                   |
| Warlock spell damage, log advanced field | 126 (likely the shadow total)                                                                             |
| Sayaad sheet Spell Damage                | 42 = 126 / 3                                                                                              |
| Lash of Pain rank 1 (7814)               | 16 base points, 0.429 coefficient, no variance                                                            |
| Lash of Pain tooltip                     | 37                                                                                                        |
| Lash of Pain in the log                  | 37 in the base field of every non-crit hit (193); dealt 37 (118), 38 (73), 39 (2), resisted 0; mean 37.40 |

(16 + 0.429 × 42) × 1.10 = 37.42 fits the mean dealt. With the sheet's 1.05 on top it would be 39.3; with 10% spell damage (12.6), 23.5. The 38s are rounding of 37.42. The two 39s hit one yeti under another warlock's Curse of the Elements rank 1 (+4% damage taken): 37.42 × 1.04 = 38.9. None of them carry a negative resist, so none are Spell Piercing.

The pass covers no other spell damage value, and not the Imp, Voidwalker or Felhunter. The attack power share (0.17) stays unmeasured: the Sayaad had 117 attack power in every log. The Sayaad's melee shows no multiplier either way, since one attack power value cannot separate a multiplier from the weapon range.

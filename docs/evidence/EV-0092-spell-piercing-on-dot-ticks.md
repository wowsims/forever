# EV-0092: Spell Piercing on DoT ticks follows the spell

- Claim: Spell Piercing on a tick follows the spell, not a tenth: Shadow Word: Pain, Corruption, Agony, Immolate, Fireball's and Frostfire Bolt's ticks bin 3 of 6,232, while Devouring Plague's, Pyroblast's and Starshards' ticks bin at the caster's direct rate.
- Sources: casters with 7 to 15 Spell Piercing; 189 raw combat logs of casters with proven direct-hit piercing, and Warcraft Logs reports of Forever from 2026-11-07 to 2026-11-13 (about 420 API points)
- Game build: not recorded
- Date: 2026-10-09 to 2026-11-13
- Pull request: #1005

A hit binned by piercing carries a negative resisted amount: bin = round(-resisted / base × 10) × 10. Targets that take a positive resist in the school are left out. Warcraft Logs drops negative resisted (0 of 858,898 cached damage events carry one), but the bonus survives as the amount above the unmitigated amount by exactly 10% or 20%. One report was checked one to one against a raw log of 1,425 hits: every raw -10% or -20% hit lands in the matching bin and no unpierced hit does. Target debuffs (Stormstrike +20%, a sticky +10%, Shadow Weaving) look the same on one hit, so hits are binned against the median of the same caster's same-school hits on that target within 10 s, and only casters showing a random 0, 10 and 20 mix at about 2:1 count: 13 mages, one warlock and one priest.

Raw logs:

| Spell (aura)                       | Ticks | Binned        | Caster's direct rate         | Caster                                                                                          |
| ---------------------------------- | ----- | ------------- | ---------------------------- | ----------------------------------------------------------------------------------------------- |
| Devouring Plague (leech, 53)       | 216   | 23 (17 and 6) | 12.0%                        | a priest, 8 piercing, level 60; other priests in group logs bin too (6 of 27, 3 of 29, 2 of 11) |
| Pyroblast DoT (periodic damage, 3) | 239   | 41            | 4% to 19%                    | two mages                                                                                       |
| Shadow Word: Pain (3)              | 515   | 0             | 12.0%                        | the level 60 priest                                                                             |
| Mind Flay (channel)                | 320   | 0             | 12.0%                        | the level 60 priest                                                                             |
| Mind Flay (channel), 2026-10-09    | 390   | 67            |                              | a priest, 7 piercing, level 30                                                                  |
| Agony, Corruption, Immolate (3)    | 1,001 | 0             | 4.1%                         | a warlock, 7 piercing, level 30                                                                 |
| Fireball DoT (3)                   | 422   | 0             | 25% on Fireball's direct hit | a mage                                                                                          |
| Frostfire Bolt DoT (3)             | 74    | 0             | 14.8%                        | a mage                                                                                          |

Warcraft Logs:

| Spell                | Casters   | Ticks | Binned                  | Caster's direct rate |
| -------------------- | --------- | ----- | ----------------------- | -------------------- |
| Pyroblast DoT        | 8 mages   | 632   | 175 (27.7%, 125 and 50) | 27.3%                |
| Devouring Plague     | a priest  | 731   | at least 51 (about 2:1) | Mind Blast 8.4%      |
| Starshards (channel) | a priest  | 218   | 16                      | 8.4%                 |
| Shadow Word: Pain    | a priest  | 2,971 | 0                       | 8.4%                 |
| Mind Flay (channel)  | a priest  | 1,418 | 0                       | 8.4%                 |
| Corruption, Agony    | a warlock | 338   | 0                       | about 11.5%          |
| Immolate DoT         | a warlock | 911   | 3                       | about 11.5%          |

No row property separates the two groups: aura type (Pyroblast's DoT is aura 3 like Shadow Word: Pain), dispel type, the attribute words and the piercing auras (none carries a class mask) were all checked.

Mind Flay, a channel, binned at the full rate at 7 piercing on 2026-10-09 (67 of 390) and never at 8 piercing in two November logs (0 of 1,738). At 7 to 8 piercing a tenth rounded to whole points also gives no bins, so a DoT log at 10 or more piercing is needed before the tenth goes.

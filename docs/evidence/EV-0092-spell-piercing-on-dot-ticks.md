# EV-0092: Spell Piercing on DoT ticks follows the spell

- Claim: Spell Piercing on a tick follows the spell: Shadow Word: Pain's, Corruption's, Agony's, Immolate's and Fireball's ticks take no bonus, while Devouring Plague's, Pyroblast's and Starshards' ticks take it at the caster's rate on direct hits.
- Sources: casters with 7 to 15 Spell Piercing; 189 raw combat logs of casters with proven direct-hit piercing, and Warcraft Logs reports of Forever from 2026-11-07 to 2026-11-13 (about 420 API points), one of them, cvRTFb1qhGBKNkCD, checked against a raw log
- Game build: not recorded; every log is from before the 2026-11-19 hotfix (1.60.1.71037, hotfixes 114309, 114310, 114312, 114313 and 114314) after which binary spells take the Spell Piercing bonus (EV-0142)
- Date: 2026-10-09 to 2026-11-13, before the 2026-11-19 hotfix
- Pull request: #1005

A hit binned by piercing carries a negative resisted amount: bin = round(-resisted / base × 10) × 10. Targets that take a positive resist in the school are left out. Warcraft Logs drops negative resisted (0 of 858,898 cached damage events carry one), but the bonus survives as the amount above the unmitigated amount by exactly 10% or 20%. One report was checked one to one against a raw log of 1,425 hits: every raw -10% or -20% hit lands in the matching bin and no unpierced hit does. Target debuffs (Stormstrike +20%, a sticky +10%, Shadow Weaving) look the same on one hit, so hits are binned against the median of the same caster's same-school hits on that target within 10 s, and only casters showing a random 0, 10 and 20 mix at about 2:1 count: 13 mages, one warlock and one priest.

Raw logs:

| Spell (aura)                       | Ticks | Binned        | Caster's direct rate         | Caster                                                                                          |
| ---------------------------------- | ----- | ------------- | ---------------------------- | ----------------------------------------------------------------------------------------------- |
| Devouring Plague (leech, 53)       | 216   | 23 (17 and 6) | 12.0%                        | a priest, 8 piercing, level 60; other priests in group logs bin too (6 of 27, 3 of 29, 2 of 11) |
| Pyroblast DoT (periodic damage, 3) | 239   | 41            | 4% to 19%                    | two mages                                                                                       |
| Shadow Word: Pain (3)              | 515   | 0             | 12.0%                        | the level 60 priest                                                                             |
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
| Corruption, Agony    | a warlock | 338   | 0                       | about 11.5%          |
| Immolate DoT         | a warlock | 911   | 3                       | about 11.5%          |

No row property separates the two groups: aura type (Pyroblast's DoT is aura 3 like Shadow Word: Pain), dispel type, the attribute words and the piercing auras (none carries a class mask) were all checked.

Shadow Word: Pain's, Corruption's, Agony's, Immolate's, Fireball's and Frostfire Bolt's ticks bin 3 of 6,232 across both tables. Immolate and Frostfire Bolt are binary spells (EV-0142); Immolate's ticks take no bonus after the 2026-11-19 hotfix too (EV-0142), and Frostfire Bolt's ticks have no count after it, so whether they take the bonus now is not measured.

At 7 to 8 piercing, a tenth of the piercing rounded to whole points gives no bins, so these logs do not tell a tick at a tenth of the share from one that takes no bonus; EV-0142 tells them apart at 11 piercing.

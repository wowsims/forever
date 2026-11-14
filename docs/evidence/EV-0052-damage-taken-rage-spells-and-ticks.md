# EV-0052: Non-physical and periodic damage taken grants twice the rage of a physical hit

- Claim: damage of a school other than physical, and every periodic tick whatever its school, grants twice the damage-taken rage of a physical direct hit.
- Sources: combat logs of a warrior at levels 27 to 38, hits from rage-using mobs, Defensive Stance divided out; Cleave's reading comes from the level 36 druid logs of EV-0037
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938
- Date: 2026-11-09 to 2026-11-13
- Pull request: #993

The factor is the observed rage over damage / c(attacker level). The sim's physical melee rule is 2.5 × (1 - 0.218) = 1.955.

| Damage taken                              | Factor [95%]         | n     |
| ----------------------------------------- | -------------------- | ----- |
| Physical melee swings, attackers 19 to 38 | 1.960 [1.944, 1.979] | 3,702 |
| Physical Shoot (direct special)           | 2.08 [1.95, 2.20]    | 74    |
| Arcane melee swings                       | 3.85 [3.80, 3.89]    | 26    |
| Fire melee swings                         | 3.83                 | 9     |
| Poison ticks                              | 3.75 [3.65, 3.84]    | 116   |
| Localized Toxin ticks                     | 3.86                 | 41    |
| Rend ticks (physical, periodic)           | 3.76                 | 44    |

Mana-using mobs give about half of each of these (physical swings 1.058, arcane swings 2.06, Fireball 2.10, Lightning Bolt 2.08). The spell and periodic readings span 3.75 to 3.86, against 1.96 for physical melee swings and 2.08 for a physical Shoot.

A mob's physical Mortal Strike read about 1.2 times a melee swing (attacker power type and sample not recorded) and Cleave about 1.3 (11 hits at level 36, from the level 36 druid logs of EV-0037); a test in game is still needed.

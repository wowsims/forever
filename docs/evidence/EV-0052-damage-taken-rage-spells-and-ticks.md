# EV-0052: Non-physical and periodic damage taken grants twice the rage of a physical melee swing

- Claim: damage of a school other than physical, and every periodic tick whatever its school, grants twice the
  damage-taken rage of a physical melee swing.
- Sources: combat logs of a warrior at levels 27 to 38, hits from rage-using mobs, Defensive Stance divided out
- Game build: 1.60.1.70862, 1.60.1.70888 and 1.60.1.70938
- Date: 2026-11-09 to 2026-11-13
- Pull request: #993

The factor is the observed rage over damage / c(attacker level). The sim's physical melee rule is 2.5 × (1 - 0.218) = 1.955.

| Damage taken                              | Factor [95%]         | n     |
| ----------------------------------------- | -------------------- | ----- |
| Physical melee swings, attackers 19 to 38 | 1.960 [1.944, 1.979] | 3,702 |
| Arcane melee swings                       | 3.85 [3.80, 3.89]    | 26    |
| Fire melee swings                         | 3.83                 | 9     |
| Poison ticks                              | 3.75 [3.65, 3.84]    | 116   |
| Localized Toxin ticks                     | 3.86                 | 41    |
| Rend ticks (physical, periodic)           | 3.76                 | 44    |

Mana-using mobs give about half of each of these (physical swings 1.058, arcane swings 2.06, Fireball 2.10, Lightning Bolt 2.08). The spell and periodic readings span 3.75 to 3.86, against 1.96 for physical melee swings.

A physical direct hit that is not a melee swing, a Shoot or a Mortal Strike among them, grants twice the melee rule's
rage on the damage left after armor and the block (EV-0118). Cleave's reading above a melee swing is EV-0037's.

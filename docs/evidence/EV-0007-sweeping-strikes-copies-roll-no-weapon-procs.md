# EV-0007: Sweeping Strikes' copied hit rolls no chance-on-hit weapon

- Claim: Sweeping Strikes' copied hit 12723, whose row states no DefenseType, rolls no chance-on-hit weapon, while its
  normalized hit 26654 rolls them at about the white hit rate.
- Sources: Arms warrior combat logs; report codes not recorded
- Game build: not recorded
- Date: not recorded
- Pull request: #965

Weapons counted: Nightblade, Runic Darkblade, Sword of Dismemberment, Rock on Stick and Bigger Rock.

| Hit                    | Chance-on-hit procs                                         |
| ---------------------- | ----------------------------------------------------------- |
| 12723 (copied hit)     | 0 on 567 copies, where the white hit rate predicts about 61 |
| 26654 (normalized hit) | 0.83 of the white hit rate (95% interval 0.37 to 1.44)      |

Equip procs such as Hand of Justice and Dreadfrost Saber's were not counted on 12723.

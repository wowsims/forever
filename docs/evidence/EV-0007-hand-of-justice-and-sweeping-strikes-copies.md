# EV-0007: Hand of Justice against Dwarves, and Sweeping Strikes' copied hit rolls no chance-on-hit weapon

- Claim: Hand of Justice procs at 1% on a melee hit and at 3 times that against a Dwarf target, whatever the wearer's
  race; Sweeping Strikes' copied hit 12723, whose row states no DefenseType, rolls no chance-on-hit weapon, while its
  normalized hit 26654 rolls them at about the white hit rate.
- Sources: for Hand of Justice, Warcraft Logs 8c9BMJ3CT7VhWgGz fights 1 and 3; RXzv1k3V4PgaWhqj fights 1, 3, 5 and 8;
  T79gRkJz4QVGbyYW fights 6 and 7; VTfg4qtPDFz6wJvX fight 1; ngdvX1HhYcbN6RCr fights 1 and 3. For Sweeping Strikes, an
  Arms warrior's local combat log and Warcraft Logs gBqYT9Da3KkpXHQj, yxhmCYJ6dTbjF9Xg, fTn3gkRLv8tXcbyA,
  mbzaQBXf9W8NnJq3 and z9K6Wf4PJxVjB2Zq, fight 1 of each
- Game build: not recorded
- Date: not recorded
- Pull request: #965

## Hand of Justice (11815)

Procs over damaging melee hits outside the 2 s cooldown.

| Targets                                      | Procs / hits | Rate  | 95% CI         |
| -------------------------------------------- | ------------ | ----- | -------------- |
| Not Dwarves                                  | 106 / 10,977 | 0.97% | 0.80% to 1.17% |
| Dark Iron dwarves (one Blackrock Depths run) | 12 / 437     | 2.75% | 1.58% to 4.74% |

An orc wearer reads 1%, so the wearer's race does not matter. The shortest gap over 229 procs is 2,073 ms. A sim run gave
0.93% (0.84% to 1.03%) per landed damaging hit.

## Sweeping Strikes' copied hit

The weapons counted are Nightblade, Runic Darkblade, Sword of Dismemberment, Rock on Stick and Bigger Rock.

| Hit                    | Chance-on-hit procs                                         |
| ---------------------- | ----------------------------------------------------------- |
| 12723 (copied hit)     | 0 on 567 copies, where the white hit rate predicts about 61 |
| 26654 (normalized hit) | 0.83 of the white hit rate (95% interval 0.37 to 1.44)      |

Equip procs such as Hand of Justice and Dreadfrost Saber's were not counted on 12723.

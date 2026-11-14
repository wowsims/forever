# EV-0026: Hand of Justice triples against Dark Iron dwarves

- Claim: Hand of Justice procs at 1% on a melee hit and at 3 times that against a Dwarf target, whatever the wearer's race.
- Sources: Warcraft Logs 8c9BMJ3CT7VhWgGz fights 1 and 3; RXzv1k3V4PgaWhqj fights 1, 3, 5 and 8; T79gRkJz4QVGbyYW fights 6 and 7; VTfg4qtPDFz6wJvX fight 1; ngdvX1HhYcbN6RCr fights 1 and 3
- Game build: not recorded
- Date: not recorded
- Pull request: #965

Procs over damaging melee hits outside the 2 s cooldown.

| Targets                                      | Procs / hits | Rate  | 95% CI         |
| -------------------------------------------- | ------------ | ----- | -------------- |
| Not Dwarves                                  | 106 / 10,977 | 0.97% | 0.80% to 1.17% |
| Dark Iron dwarves (one Blackrock Depths run) | 12 / 437     | 2.75% | 1.58% to 4.74% |

An orc wearer reads 1%, so the wearer's race does not matter. The shortest gap over 229 procs is 2,073 ms. A sim run gave
0.93% (0.84% to 1.03%) per landed damaging hit.

# EV-0033: Rip, Rake, Pounce Bleed and Swipe shares and Lacerate's weapon hit, druid levels 37 to 41

- Claim: Rip Rank 3, Rake Rank 2, Pounce Bleed Rank 1 and Swipe Rank 3 take the measured attack power shares, and Lacerate's weapon hit 414647 follows a landed cast of the rank, which alone misses, is dodged, parried or immune. The hit's crits take Natural Instinct's crit damage, as the ticks and the specials do.
- Sources: combat log with StatLedger, a Tauren feral druid at levels 37 to 41
- Game build: 1.60.1.70862
- Date: 2026-11-08
- Pull request: #915
- Later count: Lacerate's hit crit damage, §Count of the 2026-11-10 logs

| Reading                                                           | Measured                                                                                                                                                                                      |
| ----------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Rake Rank 2 (1823) at 0.065, direct hit                           | 514 of 530 hits fit                                                                                                                                                                           |
| Rake Rank 2, ticks                                                | 1,157 of 1,232 ticks fit                                                                                                                                                                      |
| Rip Rank 3 (9493) at 0.0114 × attack power × min(combo points, 4) | 210 of 237 ticks fit; 0.0091 fits 154                                                                                                                                                         |
| Swipe Rank 3 (769) at 0.03                                        | 4,055 of 5,387 hits exact                                                                                                                                                                     |
| Pounce Bleed Rank 1 (9007), 28 casts, 118 ticks                   | 118 of 118 ticks match floor(19 + c × attack power) for c in [0.02496, 0.02536); shifting form mid-dot changes the next tick                                                                  |
| Rend and Tear on Pounce Bleed's ticks                             | without it every tick fits; with it no share fits more than 113 of 118; the level 37 ticks (Rend and Tear 3/5) and the caster form ticks reject it                                            |
| Pounce Bleed counts as a bleed for other attacks' Rend and Tear   | with only Pounce Bleed on the target, Claw 3 of 3 and the Rake hit 25 of 25 took the talent                                                                                                   |
| Pounce cast and bleed                                             | the stun landed 28 of 28 and the combo point was paid first every time; the bleed was avoided 3 of 28 (2 dodges, 1 miss) with the stun and combo point still landing, and no energy came back |
| Lacerate avoidance                                                | every SPELL_MISSED in the log names the rank, never 414647: MISS 49, DODGE 47, PARRY 39, IMMUNE 15                                                                                            |
| Lacerate's hit blocked                                            | 0 of 790 hits on 414647 blocked                                                                                                                                                               |

The Ferocious Bite Ranks 2 and 3 reading of the same log is #933's (EV-0032).

## Count of the 2026-11-10 logs

- Sources: combat logs of a feral druid at levels 46 to 51, the logs EV-0034 reads
- Game build: 1.60.1.70862 or 1.60.1.70888; the hotfix push each log ran on is not recorded
- Date: 2026-11-10
- Pull request: #1287

Each crit of 414647 is divided by the hit the same stacks, attack power and Rend and Tear give without a crit.

| Reading                                              | Measured                                                                                                                           |
| ---------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Lacerate's hit crit damage with Natural Instinct 2/2 | 122 of 150 hit crits at ×2.15 to ×2.25 and none at ×1.95 to ×2.05; Swipe crits ×2.2 on 922 of 922, white swings ×2.0 on 543 of 543 |

Natural Instinct's effect names the hit's class flag in the game files as of hotfix push 113890.

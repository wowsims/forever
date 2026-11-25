# EV-0008: Windfury Totem rolls on melee abilities, never on spells

- Claim: Windfury Totem rolls on a main-hand melee swing or melee ability that lands, Holy Strike, Stormstrike, Shred,
  Rake, Ferocious Bite, Claw and bear Swipe included, and never on avoided hits, off-hand swings, DoT ticks or spells;
  triggered hits whose rows carry Not a Proc all but never roll it.
- Sources: combat logs after the hotfix, ranks 1 to 3 of the totem, seven players; report codes not recorded
- Game build: not recorded
- Date: 2026-11-13
- Pull request: #1006

Combat logs after the hotfix of 2026-11-13. The rule that an ability rolled the totem only where its row
requires a main-hand weapon or goes off on the next swing came from logs before the hotfix.

| Hit                                              | Rolls the totem                                      |
| ------------------------------------------------ | ---------------------------------------------------- |
| Main-hand white swing                            | 851 of 4,574 (18.6%) at rank 3, 145 of 858 at rank 2 |
| Holy Strike                                      | 58 of 277, 22 of 73                                  |
| Stormstrike                                      | 58 of 295, 17 of 100                                 |
| Shred, Rake (first hit), Ferocious Bite, Claw    | 53 of 332, 33 of 213, 11 of 78, 3 of 31              |
| Bear Swipe, per target hit                       | 27 of 140                                            |
| Sinister Strike, Eviscerate                      | 25 of 107, 3 of 24                                   |
| Extra attack swings                              | 34 of 216                                            |
| Avoided hits, off-hand swings, DoT ticks, spells | never                                                |

Before the hotfix Holy Strike read 0 of 115 and Stormstrike 0 of 111. In the sim, off-hand autos
(rogue 0 of 2,406, Fury 0 of 1,318), Whirlwind's off-hand hit, ranged attacks, Thunder Clap, shouts, Consecration, seals
other than Command, shocks, Deep Wound and poisons never roll it.

Triggered hits whose rows carry Not a Proc read all but never in these logs: Windfury Weapon 1 of 78 and 0 of 194 after
the hotfix; Seal of Command 0 of 117, Judgement of Command 0 of 105, Retaliation 0 of 18 and Sweeping Strikes 0 of 166
before it. The sim lets them roll the totem anyway, as their rows allow, except Sweeping Strikes' copied hit, which has no
defense type and never rolls it (EV-0002).

Measured and left for logs after the deploy: the Windfury Totem hit ignores armor in every log, the 100 ms cooldown is
not seen, it gives no Maelstrom Weapon stack, Windfury Weapon and Totem hits spend a Flurry charge after a crit, and
rank 3 adds its 246 attack power while the logs of ranks 1 and 2 add none.

# EV-0028: Revelation's mage charge and Flamestrike's ticks at level 60

EV-0142 replaces what this file says of the ticks' partial resist: they roll it against a tenth of the target's
resistance.

- Claim: the mage's Revelation charge goes as the buffed cast completes, and Flamestrike's ground ticks are hits on each enemy that take no partial resist.
- Sources: combat log, one level 60 mage with 10% sheet crit (Stratholme, area pulls)
- Game build: not recorded
- Date: not recorded
- Pull request: #990

## The mage charge

81 of 84 mage charges go in the same millisecond as the buffed cast's success and the other 3 expire. Every
enemy of the spending area cast crits (79 of 79), and one charge covers one cast. Both buffed Pyroblasts in the log spend
the charge and crit, though the row's crit modifier names neither Pyroblast nor Fireball.

## Flamestrike's ground ticks

Each Flamestrike ground tick is a hit of the tick spell 1279990 on every enemy in the area. Revelation rolls on
20 of 230 non-crit tick hits (8.7% a hit; per tick it would be 22.5%), so it rolls per enemy. Other listeners against the
ticks:

| Listener                       | Log                                                                     |
| ------------------------------ | ----------------------------------------------------------------------- |
| Revelation                     | 20 of 230 tick hits                                                     |
| Ignite                         | applied or refreshed after 67 of 68 tick crits, 0 of 287 non-crit ticks |
| Arcane Concentration           | Clearcasting after 4 of 95 ticks                                        |
| Master of Elements, Heating Up | 0 of 41 tick crits                                                      |
| Revelation's crit and charge   | ticks under the buff crit 16 of 76, as without it; no tick spends it    |
| Combustion                     | no tick inside a Combustion window                                      |

Ticks take no partial resist: 2 of 370 ticks show a resist, against 52 of 189 Scorch hits on the same mobs. In the sim, a
Flamestrike-only rotation on 5 targets (60 s) goes from 3.0 to 8.6 Revelation procs a fight once the ticks hit each
enemy.

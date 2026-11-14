# EV-0076: Voidwalker shares and halved melee at level 30

- Claim: a Voidwalker's own stats are shares of the creature values, its average swing before attack power is 0.494 of the creature dps at 2.0 s spanning 26.4% either side, and its melee is halved.
- Sources: a naked Undead level 30 warlock's screenshots of the warlock and Voidwalker sheets, no buffs, talents `-00500032` (Demonic Embrace 5/5, Fel Vitality 3/3, Demonic Energies 2/2), warlock Stamina 44 and armor 62
- Game build: not recorded
- Date: 2026-10-15
- Pull request: #447

The Voidwalker sheet reads Health 2,379, Armor 1,660, Damage 16 to 22 (Main Attack 2.00 s, 9.5 dps), Attack Power 96.

| Own stat       | Level 30                                                      | Share of the creature value                            |
| -------------- | ------------------------------------------------------------- | ------------------------------------------------------ |
| Health         | 2,379 / 1.15 Fel Vitality less 7.5 × 44 = 1,738.7             | 1.400 of the creature health (1,242)                   |
| Armor          | 1,660 less 0.35 × 62 = 1,638.3                                | 1.352 of the creature armor (1,212)                    |
| Attack power   | 96                                                            | 3 × level + 6, the Sayaad's line                       |
| Weapon (2.0 s) | about 24.4 a swing before attack power and before the halving | 0.494 of the creature dps (24.69), plus or minus 26.4% |

## The halving

With spell damage gear the Voidwalker's attack power rose to about 108 and 116 and its sheet read 17 to 23 and 17 to 24. Unhalved, +20 attack power adds 20 × 2 / 14 = 2.86 a swing and lifts the minimum to at least 18. Halved it adds 1.43, and one weapon fits all three sheets (16 to 22, 17 to 23 and 17 to 24 at attack power 96, 108 and 116): best fit 18.0 to 30.9 a swing before attack power, times 0.5. No row in the game files states it (the demon passives 18727 to 18730 carry 0), so it is the server's. With the same gear the sheet's health (2,707) and armor (1,758) match the shares above at Stamina 82.

## Open points

It is read at level 30 only; level 60 is extrapolated, and the attack power line is the Sayaad's with the Voidwalker's one point on it. At the 103 spell damage sheet the Voidwalker's attack power rose by 20 and it showed 42 spell damage, more than 17% and 10% of 103 give. The sheet shows resistances inherited in full (Arcane 5 from the warlock's 5).

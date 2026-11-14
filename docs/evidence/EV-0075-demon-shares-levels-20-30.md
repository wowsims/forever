# EV-0075: Imp and Sayaad own stats as shares of the creature values at levels 20 and 30

- Claim: an Imp's and a Sayaad's own health, armor, weapon and attack power are fixed shares of the game files' creature values for their level, each demon with its own.
- Sources: sheets and tooltips of a naked level 20 warlock and a level 30 warlock (screenshots and two combat logs); report codes not recorded
- Game build: not recorded
- Date: 2026-10-13 and 2026-10-14
- Pull request: #433

The shares were taken after removing 7.5 health a point of the warlock's Stamina, 35% of its armor, Fel Vitality's 1.15 and attack power over 14 a second.

| Sayaad, own        | Level 20                  | Level 30                                 | Share                                                 |
| ------------------ | ------------------------- | ---------------------------------------- | ----------------------------------------------------- |
| Health             | 687 less 7.5 × 33 = 439.5 | 869                                      | 0.699 of the creature health (629 and 1,242)          |
| Armor              | 559 less 0.35 × 48 = 542  | 838 (Demon Armor left out)               | 0.630 and 0.692 of the creature armor (861 and 1,212) |
| Weapon dps (2.0 s) | 12.05                     | 17.1                                     | 0.704 and 0.692 of the creature dps (17.10 and 24.69) |
| Attack power       | 66                        | 96 (tooltip "Attack Power 103 (96 + 7)") | 3 × level + 6                                         |

The Sayaad's average swing before attack power is 0.698 of the creature dps at 2.0 s. Its sheet's damage range spans 22.8% either side of the average at level 20 and 20.7% at level 30, so 21.5%.

| Imp, own (level 20) | Value                                                                                                                       | Share        |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------ |
| Health              | 496 less 7.5 × 36 = 226 (Blood Pact's +3 Stamina on the warlock); a second sheet without it reads 473 less 7.5 × 33 = 225.5 | 0.359 of 629 |
| Armor               | 297 less 0.35 × 48 = 280.2                                                                                                  | 0.325 of 861 |

The same logs and tooltips confirm the warlock's shares the sim already used: attack power 17% of the warlock's spell damage, armor 35% without Demon Armor, health 7.5 a point of Stamina (times 1.15 Fel Vitality), spell damage 10%.

In a Sayaad log of the same pass, when Demon Armor fell off (warlock armor 543 to 333) the Sayaad kept 955 armor through 65 swings and casts, while its attack power went 101 to 102 at its next action, 1.8 s after the warlock's spell damage went 30 to 36, with no re-summon.

The shares are read at levels 20 and 30, and level 60 is extrapolated; the Sayaad's 186 attack power comes from a line through 66 and 96. The Imp's shares come from level 20 only. The level 30 armor share and the Demon Armor exclusion that this reading led to were replaced by the reading of EV-0078.

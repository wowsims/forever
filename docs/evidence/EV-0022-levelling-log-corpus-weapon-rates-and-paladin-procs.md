# EV-0022: Weapon rates, paladin procs and Two-Handed Weapon Specialization's Holy reach from a levelling log corpus

- Claim: the procs per minute of Ripsaw, The Hand of Antu'sul, Jang'thraze the Protector, Thrash Blade and Bigger Rock on
  Bigger Stick, and that a druid's form specials roll no chance-on-hit weapon; Redoubt rolls on every melee swing taken,
  Shieldsworn's spell block comes with every Holy Strike while its mana and Holy Shield damage need a landed strike, Seal
  of Fury's shield is a flat share in the open world, and Touch of the Grave has a 15 s cooldown; Two-Handed Weapon
  Specialization's Holy part skips Exorcism, Hammer of Wrath and Retribution Aura and reaches Consecration.
- Sources: combat log corpus of 44 logs on 1.60.1, 18 levelling logs with StatLedger and 26 from other players; logs from
  the beta realms (11 to 17 October) counted apart
- Game build: 1.60.1
- Date: 2026-11-12 for the weapon rates; not recorded for the rest
- Pull request: #982

The #958 log of the same date is a different pass (EV-0021).

## Weapon rates

Each rate is the procs over the landed hits of the weapon, white swings and specials alike, at the weapon's speed, as the
measured rows already in the override table count.

| Weapon                                            | Procs / hits | PPM (95%)           | Row | Players                                          |
| ------------------------------------------------- | ------------ | ------------------- | --- | ------------------------------------------------ |
| Ripsaw 9478, Wound 1292573                        | 185 / 2018   | 2.04 (1.77 to 2.34) | 2   | 1, main-hand hits                                |
| The Hand of Antu'sul 9639, Thunder Clap 13532     | 91 / 991     | 2.04 (1.64 to 2.44) | 2   | 1                                                |
| Jang'thraze the Protector 11086, absorb 11657     | 181 / 1972   | 2.90 (2.50 to 3.30) | 2.9 | 1                                                |
| Thrash Blade 17705                                | 61 / 1425    | 0.95 (0.74 to 1.21) | 1   | 1, main-hand hits                                |
| Bigger Rock on Bigger Stick 279480, bleed 1306816 | 255 / 3130   | 2.00 (1.77 to 2.24) | 2   | 4, a bear and a cat druid among them, white hits |

An earlier recount of the Antu'sul, Thrash Blade and Jang'thraze counts moved Jang'thraze's row from 2.8 to 2.9.

Bigger Rock on Bigger Stick answers whether an unstated rate is procs per minute or a flat chance: it reads 11.7% a
swing at its own 3.4 speed, 8.6% at a bear's 2.5 and 2.5% at a cat's 1.0. Procs per minute fits (deviance 2.9 on 5
degrees of freedom) and a flat chance does not (49 on 5).

Long sims with the weapons equipped fire the registered procs as measured. The Hand of Antu'sul on the default
Protection paladin follows 8.0% of white swings, 8.3% of Reckoning's swings, 9.7% of Holy Strikes, 9.4% of Judgement of
Fury hits and 9.5% of Seal of Fury hits (2 PPM at 2.7 speed is 9.0%).

A druid's form specials roll no chance-on-hit weapon: 2 procs in 7736 hits, from 3 druids and 2 weapons.

| Druid                                | Maul    | Swipe (target hits) | Claw    | Lacerate |
| ------------------------------------ | ------- | ------------------- | ------- | -------- |
| One druid                            | 0 / 441 | 0 / 5401            | 0 / 248 | 0 / 309  |
| Two more druids, with another weapon | 1 / 869 | 0 / 238             | 1 / 180 | 0 / 50   |

Their white swings roll at the weapon's rate. Combat enchants on those hits are unmeasured.

## Redoubt, Shieldsworn, Seal of Fury and Touch of the Grave

| Reading                                                                              | Measured                                                                                             |
| ------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| Redoubt on avoided swings (dodged, parried, missed, fully blocked or fully absorbed) | 152 of 1394, 10.9%, about forty paladins                                                             |
| Redoubt on swings that dealt damage                                                  | 364 of 3501, 10.4%                                                                                   |
| Redoubt on melee abilities                                                           | 17 of 138                                                                                            |
| Redoubt on ranged attacks                                                            | 0 of 62                                                                                              |
| Shieldsworn's spell block on avoided Holy Strikes                                    | 46 of 46 on four paladins (15 of 15 on the two source logs), and 61 of 61 strikes made with a shield |
| Shield Specialization mana and Holy Shield damage on avoided strikes                 | 0 of 31 and 0 of 12                                                                                  |
| The same on landed strikes                                                           | 141 of 143 and 50 of 52                                                                              |
| Seal of Fury's shield in the open world                                              | 4077 of 4078 logged shields at the flat share                                                        |
| Seal of Fury's shield in a dungeon                                                   | scaled by expected creature health, see below                                                        |
| Touch of the Grave cooldown                                                          | 57 players, 707 drains, none closer than 14.99 s                                                     |

Inside a dungeon Seal of Fury's shield scales by expected creature health at a base level over the mob's level, the dungeon's
lowest mob level being the base in five dungeons (Scarlet Monastery's base 30 on two players, the other four on one
each). Against mobs of the paladin's own level it stays flat, 118 of 118, three cases on two players.

The sim reproduces two of these readings. Redoubt rolls after 41 of 410 dodged, parried or missed boss swings, where the
sim before the change gave it no roll (0 of 444). Shieldsworn's spell block comes with 409 of 409 dodged or missed
strikes.

## Two-Handed Weapon Specialization's Holy part

Holy spell multipliers with the talent at 3/3 and a two-hander, once mobs with their own damage reduction are split out:

| Holy spell                 | Measured multiplier                                  |
| -------------------------- | ---------------------------------------------------- |
| Consecration               | 1.021, 1.036 and 1.049 to 1.062 for Ranks 1, 2 and 3 |
| Exorcism, Retribution Aura | 1.001 to 1.010 and 0.997 to 1.000 by rank            |
| Hammer of Wrath            | 0.991 (n = 77, one player)                           |

The talent at 3/3 gives 1.06. Holy Forgefire was measured without the bonus too.

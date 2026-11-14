# EV-0022: Chance-on-hit weapon rates from a levelling log corpus

- Claim: the procs per minute of Ripsaw, The Hand of Antu'sul, Jang'thraze the Protector, Thrash Blade and Bigger Rock on Bigger Stick, and that a druid's form specials roll no chance-on-hit weapon.
- Sources: combat log corpus of 44 logs on 1.60.1, 18 levelling logs with StatLedger and 26 from other players; logs from the beta realms (11 to 17 October) counted apart
- Game build: 1.60.1
- Date: 2026-11-12
- Pull request: #982

The same corpus backs EV-0023. The #958 log of the same date is a different pass (EV-0021).

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

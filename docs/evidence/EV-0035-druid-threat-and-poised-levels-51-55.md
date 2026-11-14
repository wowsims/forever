# EV-0035: Shifting Power and Blood Frenzy threat and Poised on avoided casts, druid levels 51 to 55

- Claim: Shifting Power's energy and Blood Frenzy's rage threaten without the form's multiplier, and an avoided cast spends Poised.
- Sources: StatLedger, ThreatLog and combat logs of a feral druid at levels 51 to 55 in Tanaris, Un'Goro and the Sunken Temple
- Game build: 1.60.1.70888
- Date: not recorded
- Pull request: #960

| Reading                                                     | Measured                                                                                                         |
| ----------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| Shifting Power's energy threatens 0.5 a point               | 40 energy read 20 threat at one target on 10 windows; Cat Form's ×0.71 rejected                                  |
| Blood Frenzy's rage threatens without the form's multiplier | 5 rage read 24.7 and 25.0 in Dire Bear Form; the form-scaled 53.75 never appears                                 |
| An avoided cast spends Poised                               | with Poised up, 27 of 27 avoided Claws, Rakes, Pounces and Ravages (dodge, parry, miss, immune) spent it         |
| Pounce's cast can miss (same logs)                          | 2 of 74 casts logged SPELL_MISSED MISS on Pounce 9823 right after the cast; 8 of the 40 cost spent               |
| A Lacerate dose keeps the bleed's tick grid (same logs)     | first tick 3.0 s after a fresh application 258 of 258; tick gaps across a dose or refresh 3.0 ± 0.1 s 212 of 212 |
| Damage-taken rage at attacker levels 50 to 55               | see EV-0051                                                                                                      |

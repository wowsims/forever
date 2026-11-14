# EV-0064: Two-Handed Weapon Specialization's Holy follows the defense type

- Claim: Two-Handed Weapon Specialization's Holy part reaches the melee-defense Holy hits (the seal hits, Holy Strike and
  Judgement of Command) and not the magic-defense or ranged-defense ones, except Consecration, whose ticks take it.
- Sources: logs; report codes not recorded
- Game build: not recorded
- Date: 2026-11-13 for the Judgement of Command and Seal of Righteousness readings, after the hotfix; the rest not
  recorded
- Pull request: #1009

Holy spell with the talent at 3/3 and a two-hander:

| Holy spell                                   | Defense | Measured multiplier                                                                            |
| -------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------- |
| Seal hits, Holy Strike, Judgement of Command | melee   | Judgement of Command 1.062 (n = 109)                                                           |
| Consecration                                 | magic   | 1.021, 1.036 and 1.049 to 1.062 by rank                                                        |
| The ranged judgements                        | ranged  | Righteousness 1.003 (0.999 to 1.008, n = 93) on 2026-11-13; 1.050 (n = 192) while it was melee |
| Exorcism, Retribution Aura                   | magic   | 1.001 to 1.010 and 0.997 to 1.000                                                              |
| Hammer of Wrath                              | ranged  | 0.991 (n = 77)                                                                                 |
| Holy Wrath, Holy Shock                       | magic   | not measured                                                                                   |

Holy Forgefire 1322218, which the logs read without the bonus, is magic-defense too. The Consecration, Exorcism,
Retribution Aura and Hammer of Wrath readings and the 1.050 reading come from the earlier logs of pull request #982,
whose counts are not recorded here.

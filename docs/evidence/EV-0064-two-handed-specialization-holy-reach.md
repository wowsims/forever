# EV-0064: Two-Handed Weapon Specialization's Holy follows the defense type

- Claim: Two-Handed Weapon Specialization's Holy part reaches Judgement of Command, a melee-defense Holy hit, and not
  Judgement of Righteousness once the judgements are ranged-defense.
- Sources: logs; report codes not recorded
- Game build: not recorded
- Date: 2026-11-13 for the Judgement of Command and Seal of Righteousness readings, after the hotfix; not recorded for
  the reading while the judgements were melee
- Pull request: #1009

Holy spell with the talent at 3/3 and a two-hander:

| Holy spell                                   | Defense | Measured multiplier                                                                            |
| -------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------- |
| Seal hits, Holy Strike, Judgement of Command | melee   | Judgement of Command 1.062 (n = 109)                                                           |
| The ranged judgements                        | ranged  | Righteousness 1.003 (0.999 to 1.008, n = 93) on 2026-11-13; 1.050 (n = 192) while it was melee |
| Holy Wrath, Holy Shock                       | magic   | not measured                                                                                   |

The Consecration, Exorcism, Retribution Aura and Hammer of Wrath readings the pull request rests on are those of #982
(EV-0022): Consecration takes the bonus although it is magic-defense, and the other three take none. Holy Forgefire
1322218, which the logs read without the bonus, is magic-defense too.

# EV-0018: Seal of Light, Seal of Wisdom and their judgements at level 41

- Claim: Seal of Light Rank 2 and Seal of Wisdom Rank 1 proc at 15 a minute, Judgement of Light Rank 2 and Judgement of
  Wisdom Rank 1 at 50%, and the judgements pay on landed hits only.
- Sources: combat log of one level 41 paladin (Retribution 31, Holy 1): open world and five Scarlet Monastery runs, a
  3.3 s two-hander and a 2.6 s one-hander with a shield
- Game build: not recorded
- Date: 2026-11-09
- Pull request: #925

| Reading                                                      | Measured                                                                 |
| ------------------------------------------------------------ | ------------------------------------------------------------------------ |
| Seal of Light Rank 2                                         | 90 of 129 landed melee hits at 2.6 s, 0.70 where 15 a minute gives 0.65  |
| Seal of Wisdom Rank 1                                        | 187 of 269 landed melee hits at 2.6 s, 0.70 where 15 a minute gives 0.65 |
| Judgement of Light Rank 2                                    | 59 of 123 landed melee swings; never on a ranged or spell hit            |
| Judgement of Wisdom Rank 1, landed melee                     | 164 of 336                                                               |
| Judgement of Wisdom Rank 1, direct spells                    | 59 of 116                                                                |
| Judgement of Wisdom Rank 1, ranged shots                     | 44 of 102                                                                |
| Judgement of Wisdom Rank 1, missed, dodged or parried swings | 1 of 52 (0 of 32 for the paladin, 1 of 20 for the party)                 |
| Judgement of Wisdom Rank 1, dot ticks                        | 3 of 401                                                                 |
| Judgement of Wisdom Rank 1, damage shield hits               | 0 of 33                                                                  |
| Judgement of Wisdom Rank 1, Consecration ticks               | 0 of 67                                                                  |

16 a minute would fit the two seal rates better but fails the 3.7 s Seal of Light Rank 1 log (EV-0017), so the rate
stays 15. Holy Strike rolls both seals. No log has measured a landed hit that deals no damage reaching the judgement.

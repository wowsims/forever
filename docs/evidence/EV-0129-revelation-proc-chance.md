# EV-0129: Revelation's proc chance per non-crit hit

- Claim: Revelation (enchant 8217, proc spell 1248806) rolls about 8% on each non-crit, non-periodic damage or heal hit
  at 8.5 to 10.1% sheet crit, and 6.8% in the level 30 logs at 13.42%; crits and aura-only applications never roll it,
  and it can roll while the buff is up.
- Sources: combat logs, a druid (ten logs) and a mage (one log).
- Game build: 1.60.1.70425 for the druid's logs and 1.60.1.70938 for the mage's
- Date: 2026-10-03 to 2026-10-06 (druid) and 2026-11-13 (mage)
- Pull request: #1051

A roll is one hit that deals damage or heals with no crit, from a direct spell or heal. A proc is a `SPELL_AURA_APPLIED`
or `SPELL_AURA_REFRESH` of the class's Revelation buff on the wearer within 250 ms after such a hit: druid 1323400,
mage 1248808 (the other classes' buffs are 1323377, 1323392, 1323410, 1323418 and 1323419). Crits, aura-only
applications (buffs, roots, dot applications) and periodic ticks do not count as hits. The mage log's Flamestrike
ground ticks (1279990) log as direct hits and count, one roll per target per tick.

| Log                   | Class, level | Sheet crit | Procs / hits | Rate  |
| --------------------- | ------------ | ---------- | ------------ | ----- |
| 2026-10-03            | druid, 22-23 | 10.1%      | 30/360       | 8.3%  |
| 2026-10-04            | druid, 22-23 | 10.0%      | 20/232       | 8.6%  |
| 2026-10-04            | druid, 22-23 | 9.66%      | 28/265       | 10.6% |
| 2026-10-04            | druid, 22-23 | 9.63%      | 23/311       | 7.4%  |
| 2026-10-05            | druid, 22-23 | 9.63%      | 24/289       | 8.3%  |
| 2026-10-05            | druid, 22-23 | 9.3%       | 36/404       | 8.9%  |
| 2026-10-05            | druid, 22-23 | 8.72%      | 14/178       | 7.9%  |
| 2026-10-05            | druid, 22-23 | 8.52%      | 14/194       | 7.2%  |
| Druid pooled          | druid, 22-23 | 8.52-10.1% | 189/2,233    | 8.5%  |
| 2026-10-06            | druid, 30    | 13.42%     | 85/1,391     | 6.1%  |
| 2026-10-06            | druid, 30    | 13.42%     | 53/652       | 8.1%  |
| Druid level 30 pooled | druid, 30    | 13.42%     | 138/2,043    | 6.8%  |
| 2026-11-13            | mage, 60     | 10%        | 94/1,277     | 7.4%  |

The mage's 10% is the owner's figure for the sheet; its level comes from the top spell ranks in the log. The 13.42% logs
differ by level and by Wrath rank as well as crit, so a crit falloff and a level effect are not told apart. The mage's
7.4% sits between the druid pool at 8.5% and the level 30 pool at 6.8%, and against the rolling spell's measured crit
(7.2% under 15%, 7.4% at 15 to 25%, 7.6% at 25% and above) it shows no trend.

Other readings, from the same logs:

- Crits never proc it: 0 of 967 crit hits in the ten druid logs.
- Aura-only applications never proc it: 0 of 999 in the ten druid logs; 0 of 186 mage batches whose every hit crit
  were followed by a proc.
- It can proc while the buff is up, as a `SPELL_AURA_REFRESH`: mage 6/66 non-crit hits while up (9.1%); the druid logs
  show 4 procs in 32 such hits.

The mage's 100 procs are 84 applications and 16 refreshes.

The sim uses 8%: the owner's round figure between the druid pool (8.5%) and the mage (7.4%), inside the 8.5 to 10% crit
band. How the chance falls as crit rises is not modelled, so the sim rolls the same 8% at any crit.

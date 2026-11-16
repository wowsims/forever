# EV-0127: Summon Hawk's hawk appears 1 sec after the cast and swings at once

- Claim: a hawk appears as the Summon Hawk dive lands, 1 sec after the cast at any distance, and swings as it appears, also when it replaces another hawk.
- Sources: five combat logs with advanced logging, hunters of levels 27 to 48: one group log of 2026-11-07, one log of 2026-11-08, and three logs of one level 48 Beast Mastery hunter of 2026-11-11; report codes not recorded
- Game build: 1.60.1 (the log headers name no build number); the 2026-11-07 and 2026-11-08 logs ran before the 2026-11-10 hotfix, the 2026-11-11 logs after it
- Date: 2026-11-07, 2026-11-08 and 2026-11-11
- Pull request: not opened yet

For each `SPELL_SUMMON` of the Hawk creature (264850), the gap back to its summoner's last `SPELL_CAST_SUCCESS` of
Summon Hawk (1293241, 1293525, 1293526 or 1293527), and the gap forward to the hawk's first `SWING_DAMAGE` or
`SWING_MISSED`. The distance runs from the caster's position at the cast to the hawk's position at its first
`SWING_DAMAGE`, which is beside the target. A hawk replaces another when its summoner summoned two hawks in the 18 sec
before it.

| Log                | Summons | Cast to summon, median | Cast to summon, range | First swing in the log | Of those, within 50 ms of the summon |
| ------------------ | ------- | ---------------------- | --------------------- | ---------------------- | ------------------------------------ |
| 2026-11-07         | 81      | 0.994 s                | 0.919-1.129 s         | 34                     | 21                                   |
| 2026-11-08         | 10      | 1.001 s                | 0.951-1.094 s         | 4                      | 3                                    |
| 2026-11-11, 1 of 3 | 37      | 0.994 s                | 0.912-1.037 s         | 36                     | 35                                   |
| 2026-11-11, 2 of 3 | 15      | 1.010 s                | 0.925-1.078 s         | 15                     | 15                                   |
| 2026-11-11, 3 of 3 | 17      | 1.002 s                | 0.917-1.041 s         | 17                     | 16                                   |
| All                | 160     |                        | 0.912-1.129 s         | 106                    | 90                                   |

- The gap does not grow with distance. Over 101 casts from 2 to 57 yd, 95 of them within the spell's 35 yd, the fitted
  line is 0.999 sec plus -0.0001 sec a yard. The row states a missile speed of 35 and a minimum flight of 1 sec
  (`MinDuration`), and within 35 yd the minimum is the longer of the two.
- The dive's damage lands with the summon: in the 51 casts that log both, within 1 ms.
- A hawk swings as it appears whether or not it replaces another: 19 of the 20 replacing hawks and 71 of the other 86
  swing within 50 ms of the summon.
- The other 16 first swings come 0.4 to 17 sec after the summon, 13 of them in the 2026-11-07 group log. The logs do
  not show why.

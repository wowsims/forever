# EV-0044: Bloodrage's instant rage threat and no rage threat before the pull

- Claim: Bloodrage's instant rage makes 5 threat a rage in Battle and Defensive Stance alike and Defiance does not
  scale it, while its rage over time takes the stance's multiplier, and rage gained before the pull makes no threat.
- Sources: ThreatLog addon and combat log of a warrior, day and evening sessions; the evening log holds 120 fights and
  942 threat windows. The reading under Defiance gives its own
- Game build: 1.60.1.70790; the reading under Defiance gives its own
- Date: 2026-11-06; the reading under Defiance gives its own
- Pull request: #882; the reading under Defiance is #1157

The method is that of EV-0043, applied to the evening log and re-checked on the day data.

- Count: §Count of the 2026-11-13 to 2026-11-20 logs, Bloodrage's instant rage under Defiance

| Reading                     | Measured                                                                                                                                                                               |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bloodrage instant 10 rage   | 50 threat in both stances: Defensive 5.00 [4.91, 5.09] per rage (n = 3) in the evening; day data Battle 5.00 three times and Defensive 5.00 twice; six mid fight casts read 50 by hand |
| Bloodrage ticks             | 5 per rage times the stance, Defensive 8.72 [7.82, 9.63] (n = 24)                                                                                                                      |
| Rage gained before the pull | 0 threat: counting from the pull moves fight totals from 86% to 95% within 5%                                                                                                          |
| Shield Bash                 | 1.5 x damage x stance; 4 to 9 damage hits fit 1.5 x and +2 alike                                                                                                                       |
| Hamstring                   | 1.25 x damage x stance (5 threat per 4 damage hit, day data n = 3); 4 damage hits fit 1.25 x and +1 alike                                                                              |

Together the Bloodrage changes close the 6 to 8% gap that fights with Bloodrage showed on both models in EV-0043: the
day data's 19 such fights go from 0.925 to 1.001 (actual over model). Only rage was measured before the pull; heals and
mana gains follow the same rule because no target is in combat yet.

## Count of the 2026-11-13 to 2026-11-20 logs

- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at levels 38 to 39 and 42 to 47, Defiance 3
  and Improved Bloodrage 2 throughout
- Game build: 1.60.1.70938 on 2026-11-13 and 2026-11-14 (hotfixes 114060, 114075 and 114091, each taken at a login
  inside the session) and on 2026-11-17; 1.60.1.70959 from the 2026-11-18 09:07 login, hotfixes 114140 to 114286 each
  taken at a login once pushed; no reading moves between them
- Date: 2026-11-13 to 2026-11-20
- Pull request: #1157

The method is EV-0043's, run per log as EV-0174 describes. All readings are in Defensive Stance.

| Reading                                    | Measured                                                                                                                                                                                                                                                                                                                                 |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bloodrage instant 15 rage under Defiance 3 | 75 threat, 5 a rage: levels 42 to 47, 75 ± 1.5 on 25 of 73 one-mob windows and 40 of 86 party, 37.5 on 25 and 14 more and 25 on 2 and 1 more inside the split stretches, the rest with an event near a window edge; levels 38 to 39, 75 once solo and once in a party; none of them reads 90 (x Defiance) or 162 (x stance and Defiance) |

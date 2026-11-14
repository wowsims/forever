# EV-0044: Bloodrage's instant rage threat and no rage threat before the pull

- Claim: Bloodrage's instant rage makes 5 threat a rage in Battle and Defensive Stance alike, while its rage over time
  takes the stance's multiplier, and rage gained before the pull makes no threat.
- Sources: ThreatLog addon and combat log of a warrior, day and evening sessions; the evening log holds 120 fights and
  942 threat windows
- Game build: 1.60.1.70790
- Date: 2026-11-06
- Pull request: #882

The method is that of EV-0043, applied to the evening log and re-checked on the day data.

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

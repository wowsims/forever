# EV-0206: A missed Auto Shot rolls no Naga Heartpiercer, a Mongoose Bite spends Expose Prey, and a missed Strider Kick still gives the speed

- Claim: an Auto Shot that misses never rolls the Naga Heartpiercer's Pierce Heart; a Mongoose Bite cast with Expose Prey
  up and a dodge in the last 5 s removes Expose Prey; and a Strider Kick that misses, is dodged or is parried puts the
  Strider Kick speed buff on the hunter as a landed kick does.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the hunters' own event streams in every report of the
  sweep's local index that holds the item or the ability (counts under Counts); report codes are not listed here. Report
  start dates run 2026-11-10 to 2026-11-24
- Game build: 1.60.1.70888 to 1.60.1.71037 by report start date, on the US times (70888 from 2026-11-10 15:11 UTC, 70938 from 2026-11-13
  01:36, 70959 from 2026-11-17 15:19, 71037 from 2026-11-20 18:23); the hotfix each player ran is not recorded
- Date: 2026-11-10 to 2026-11-24
- Pull request: not recorded

Two streams of each of the Mongoose Bite and Strider Kick counts date to build 70862 and hold no counted cast.

## Counts

| Item                                        | Reports                        | Streams | What was counted                                                                                                                                                                                          |
| ------------------------------------------- | ------------------------------ | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Naga Heartpiercer (3078 in the ranged slot) | 136 (2026-11-12 to 2026-11-24) | 136     | each Auto Shot (75) cast, matched to its damage or miss event within 3 s on the same target, and a Pierce Heart event (1292647) from -5 to +15 ms of the cast                                             |
| Mongoose Bite (Ranks 1 to 4)                | 363 (2026-11-10 to 2026-11-24) | 370     | each bite cast with Expose Prey (1310726) up for 5 ms or more before it and a dodge of a melee attack on the hunter in the 5 s before it; removed means the Expose Prey removal within 300 ms of the cast |
| Strider Kick (1317257)                      | 415 (2026-11-10 to 2026-11-24) | 424     | each cast's outcome within 300 ms and a speed buff applied or refreshed on the hunter from 50 ms before to 300 ms after the cast                                                                          |

## Naga Heartpiercer rolls on landed Auto Shots

| Build | Auto Shots that missed, with a Pierce Heart | Auto Shots that landed, with a Pierce Heart |
| ----- | ------------------------------------------- | ------------------------------------------- |
| 70888 | 0 of 0                                      | 37 of 312                                   |
| 70938 | 1 of 557                                    | 1,121 of 18,751                             |
| 70959 | 2 of 592                                    | 1,964 of 26,737                             |
| 71037 | 0 of 537                                    | 1,446 of 19,640                             |
| All   | 3 of 1,686 (0.18%)                          | 4,568 of 65,440 (7.0%)                      |

At the landed rate, 1,686 misses would give about 118. The 3 were not traced. The rate itself is in EV-0136. Only Auto
Shot was counted, not the ranged specials.

## Mongoose Bite spends Expose Prey

| Build | Expose Prey removed | Expose Prey kept |
| ----- | ------------------- | ---------------- |
| 70888 | 17                  | 0                |
| 70938 | 144                 | 2                |
| 70959 | 176                 | 0                |
| 71037 | 323                 | 0                |
| All   | 660                 | 2                |

Expose Prey lasts 10 s, the dodge's Defensive State (5302) 5 s with one charge, and a bite has a 5 s cooldown, so a second
bite on one dodge cannot happen; the logs do not show the Defensive State aura, and whether the bite spends it too is not
observable.

## Strider Kick's speed comes with every outcome

| Outcome | 70888    | 70938      | 70959      | 71037      | All            |
| ------- | -------- | ---------- | ---------- | ---------- | -------------- |
| Miss    | 3 of 3   | 153 of 153 | 210 of 210 | 272 of 272 | 638 of 638     |
| Dodge   | 30 of 30 | 476 of 476 | 599 of 599 | 891 of 891 | 1,996 of 1,996 |
| Parry   | 14 of 14 | 331 of 331 | 370 of 370 | 542 of 542 | 1,257 of 1,257 |

Every other outcome (46,568 casts) and the 231 casts with no outcome event carried the buff too.

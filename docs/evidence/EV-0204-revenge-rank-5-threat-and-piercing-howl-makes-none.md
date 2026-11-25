# EV-0204: Revenge Rank 5 makes 2.25 times its damage as threat, and Piercing Howl makes none

- Claim: Revenge's threat on Rank 5 is 2.25 times its damage, as on Ranks 3 and 4 (EV-0048), and a Piercing Howl makes no
  threat.
- Sources: the addon's threat log and combat logs of a Protection warrior at levels 45 to 59 (6,110 clean windows of the
  20,528 label-matched, over 1,561 fights); report codes not recorded
- Game build: 1.60.1.70862 to 1.60.1.70938 (70888 from 2026-11-10 12:44 and 70938 from 2026-11-13 17:09 in the log's
  own clock); the hotfix the player ran is not recorded
- Date: 2026-11-09 to 2026-11-13
- Pull request: not recorded

A window is the threat the addon logged around one ability, matched to the combat log's damage; the counts are clean
windows, and the stance's 2.16 in Defensive Stance is divided out of the threat.

| Reading                                   | Measured                                                                                                                                           |
| ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| Revenge Ranks 4 and 5, threat over damage | 2.2501 (742 windows) and 2.2500 (102 windows); the count does not say which number is which rank, and Rank 5 is cast only from 2026-11-13 01:44 on |
| Piercing Howl, threat                     | 0.00 over 4 windows (learned 2026-11-11 08:10)                                                                                                     |

The Revenge reading lists Ranks 4 and 5 together; Rank 4's own threat is in EV-0048. Rank 6 (25288) is in no report
and keeps its TODO.

# EV-0204: Revenge Rank 5 makes 2.25 times its damage as threat, and Piercing Howl makes none

- Claim: Revenge's threat on Rank 5 is 2.25 times its damage, as on Ranks 3 and 4 (EV-0048), and a Piercing Howl makes no
  threat.
- Sources: the addon's threat log and combat logs of a Protection warrior at levels 45 to 59 (6,110 clean windows of the
  20,528 label-matched, over 1,561 fights); report codes not recorded
- Game build: 1.60.1.70862 to 1.60.1.70938 (70888 from 2026-11-10 12:44 and 70938 from 2026-11-13 17:09 in the log's
  own clock); the hotfix the player ran is not recorded
- Date: 2026-11-09 to 2026-11-13
- Pull request: #1232
- Later count: Revenge Rank 5, §Count of the 2026-11-24 to 2026-11-26 logs

A window is the threat the addon logged around one ability, matched to the combat log's damage; the counts are clean
windows, and the stance's 2.16 in Defensive Stance is divided out of the threat.

| Reading               | Measured                                       |
| --------------------- | ---------------------------------------------- |
| Piercing Howl, threat | 0.00 over 4 windows (learned 2026-11-11 08:10) |

Revenge Rank 6 (25288) is not measured.

## Count of the 2026-11-24 to 2026-11-26 logs

- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at level 60 in dungeon groups, Defiance 3
  and Improved Revenge 3
- Game build: 1.60.1.71091 on 2026-11-24 and 1.60.1.71139 from the 2026-11-25 18:17 login; the hotfix is not recorded;
  no reading moves between the two builds
- Date: 2026-11-24 to 2026-11-26
- Pull request: #1254

The method is EV-0220's.

| Reading                            | Measured                                                                                                                            |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| Revenge Rank 5, threat over damage | 2.250 [2.250, 2.250] (n 53, strict; 19 on 1.60.1.71091, 34 on 71139); a flat bonus over 2.25 x damage would read 0.05 [-0.02, 0.17] |

# EV-0215: Heroic Strike Rank 8 adds 145 flat threat

- Claim: Heroic Strike Rank 8 (11567) makes its damage plus 145 threat, times the stance; Rank 4's 2.5 a spell level
  would give 140.
- Sources: combat logs and threat log of one protection warrior at levels 58 and 59, Defensive Stance with Defiance
- Game build: 1.60.1.70888 for the 16 windows at level 58, 1.60.1.70938 for the one at level 59; the hotfix is not
  recorded
- Date: 2026-11-13
- Pull request: not yet opened

A window is the threat the warrior's threat log gained over one landed Heroic Strike Rank 8 hit, matched by its label,
on a mob it was engaged with. Every other term in the window is a rule already measured, each event rounded to the
nearest whole point: white swings, item procs and Rend damage times the stance, Bloodrage ticks, Revenge, Thunder Clap,
Shield Slam, Heroic Strike Rank 5 and Sunder Armor (EV-0211). Windows with a term no rule covers (Taunt, a stance swap,
Execute, Whirlwind and the like) and the first window of a pull are left out. What is left is the strike's own
whole-point threat y; F = y / 2.16 − damage is the flat it implies.

17 windows are kept; 15 lie within 10 of the median of F, 144.72 (mean 143.34 [141.86, 144.82]). On those 15:

| Flat | round(2.16 x (damage + flat)), exact | Within 1 | round(2.16 x damage) + round(2.16 x flat), exact | Within 1 |
| ---- | ------------------------------------ | -------- | ------------------------------------------------ | -------- |
| 139  | 0                                    | 2        | 2                                                | 2        |
| 140  | 0                                    | 0        | 0                                                | 0        |
| 142  | 1                                    | 1        | 1                                                | 1        |
| 145  | 6                                    | 11       | 9                                                | 11       |
| 146  | 0                                    | 0        | 0                                                | 2        |

Every other flat from 136 to 150 matches no window exactly. The whole-point flat, y − round(2.16 x damage), reads 313
on 9 windows, 314 on 2, 307 on 1, 300 on 2 and 295 on 1; round(2.16 x 145) is 313, and 140 would read 302, which no
window does. The 14 windows on a single mob read the same (median 144.71, 8 exact at 145).

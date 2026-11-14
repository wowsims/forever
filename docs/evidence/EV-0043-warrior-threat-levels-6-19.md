# EV-0043: Heroic Strike and Overpower threat, warrior levels 6 to 19

- Claim: Heroic Strike Ranks 1 to 3 add flat threat over their damage (16.6, 27.6 and 40) and Overpower's threat is 0.75
  times its damage, each times the stance.
- Sources: ThreatLog addon and combat log of a warrior at levels 6 to 19
- Game build: 1.60.1.70790
- Date: 2026-11-05 and 2026-11-06
- Pull request: #876

The addon records the player's threat on the target at every threat list update (about 1.2 s apart), and each change is
matched to the combat log events since the previous one, with each event multiplied by the stance active at that moment.
The numbers come from clean windows: one kind of action plus white swings, no event within 0.3 s of a window edge, no
stance change, one mob. Intervals are 95%.

| Reading                                | Measured                                                                                    |
| -------------------------------------- | ------------------------------------------------------------------------------------------- |
| Heroic Strike Rank 1 (78), flat threat | +16.61 [16.44, 16.79], n = 53, Battle                                                       |
| Heroic Strike Rank 2 (284)             | +27.60 [27.52, 27.69], n = 61, Battle                                                       |
| Heroic Strike Rank 3 (285)             | +40.05 [39.97, 40.13], n = 32, Defensive                                                    |
| Heroic Strike Rank 4 (1608)            | 61, the 2026-10-07 measurement (EV-0042)                                                    |
| Overpower threat multiplier            | 0.76 [0.75, 0.76] clean (n = 4); 0.745 [0.668, 0.817] over 15 windows                       |
| Shield Bash Rank 1 (72), flat threat   | +2.07 [1.91, 2.23], n = 3, Defensive; +2 measured earlier                                   |
| Shield Specialization rage threat      | 25 per 5 rage proc on 8 of 8 clean single mob Defensive windows; 45 (5 x 1.8) never appears |
| Bloodrage rage ticks                   | Battle 3.97 [3.37, 4.57], n = 55; Defensive 9.00 [8.77, 9.23], n = 6                        |

16.6, 27.6, 40 and 61 at spell levels 1, 8, 16 and 24 fit no single share per spell level. Not changed: Hamstring (5
threat per 4 damage hit fits a factor of 1.25 and +1 alike, n = 3), and a halving of every rage source's threat from
about 14:05 on 2026-11-06 that nothing in gear, talents, group or zone explains. Defensive Stance's multiplier is 1.8
and Battle Stance's 0.8 (Defensive Stance Passive 7376 went from +30 to +80 at 70790); Revenge 2.25, Thunder Clap, Rend,
Battle Shout, Shield Block and Charge's rage all matched the sim already.

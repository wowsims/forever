# EV-0043: Heroic Strike and Overpower threat, warrior levels 6 to 19

- Claim: Heroic Strike Ranks 1 to 3 add flat threat over their damage (16.6, 27.6 and 40) and Overpower's threat is 0.75
  times its damage, each times the stance, and Battle Shout Ranks 1 to 5 make no threat.
- Sources: ThreatLog addon and combat log of a warrior at levels 6 to 19; Battle Shout's count gives its own
- Game build: 1.60.1.70790; Battle Shout's count gives its own
- Date: 2026-11-05 and 2026-11-06; Battle Shout's count gives its own
- Pull request: #876; Battle Shout's count is #1157

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

16.6, 27.6, 40 and 61 at spell levels 1, 8, 16 and 24 fit no single share per spell level. Two readings left the code unchanged: Hamstring (5
threat per 4 damage hit fits a factor of 1.25 and +1 alike, n = 3), and a halving of every rage source's threat from
about 14:05 on 2026-11-06 that nothing in gear, talents, group or zone explains. Defensive Stance's multiplier is 1.8
and Battle Stance's 0.8 (Defensive Stance Passive 7376 went from +30 to +80 at 70790); Revenge 2.25, Thunder Clap, Rend,
Shield Block and Charge's rage all matched the sim already.

- Count: §Count of the 2026-11-13 to 2026-11-20 logs, Battle Shout

## Count of the 2026-11-13 to 2026-11-20 logs

- Sources: combat logs, ThreatLog and StatLedger of one protection warrior at levels 38 to 39 and 42 to 47
- Game build: 1.60.1.70938 on 2026-11-13 and 2026-11-14 (hotfixes 114060, 114075 and 114091, each taken at a login
  inside the session) and on 2026-11-17; 1.60.1.70959 from the 2026-11-18 09:07 login, hotfixes 114140 to 114286 each
  taken at a login once pushed; no reading moves between them
- Date: 2026-11-13 to 2026-11-20
- Pull request: #1157

The method is the one above, run per log as EV-0174 describes. No window holds Battle Shout alone, so each reading is
the window's change less every other event in it. Ranks 1 to 3, cast at levels 6 to 19, read 0 too, with no counts
kept.

| Reading                     | Measured                                                                                                                                                                                                         |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Battle Shout Rank 4 (11549) | 0: within ±1 on 20 of 26 solo and 8 of 12 party windows at levels 38 to 39, and 5 of 6 at level 42                                                                                                               |
| Battle Shout Rank 5 (11550) | 0: within ±1 on 18 of 31 solo and 33 of 41 party windows, medians −0.12 and −0.07; Battle Shout and white swings only, 4 of 6 and 13 of 14; 10 of the 13 solo misses have an event within 0.3 s of a window edge |

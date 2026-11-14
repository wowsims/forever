# EV-0024: Sword of Dismemberment procs 0.6 times a minute

- Claim: Sword of Dismemberment 287005, whose Dismember row 1324687 states no rate, procs at 0.6 per minute without extra attacks or haste; seal and judgement hits roll it.
- Sources: Warcraft Logs reports of dungeons below the end game (Scarlet Monastery, Razorfen Downs, Uldaman, Zul'Farrak, Maraudon, Sunken Temple, Krol'dok Stronghold, Alcaz Prison and The Drowned City), 2791 reports searched, the proc found in 11; the three kept runs are hWHgAPGmrNpDxwCR, FMcQw7KgbLjhy3CN and VnD3b8WCJQfxr4XZ
- Game build: 1.60.1.70888 to 1.60.1.70938
- Date: 2026-11-13
- Pull request: #985

Procs over the weapon's landed hits, white swings and specials alike, at its 3.7 speed, as the other measured rows count.
Only fights where CombatantInfo shows the sword in the main hand count. Extra-attack swings proc it (10 of the 224 procs
in these reports), so runs with extra attacks are left out. Within the three runs kept, haste is cut as well: no hit made
while an attack speed aura is up on the wielder (Seal of the Crusader, +40%, is on all three), and no white swing that
came sooner than 3.6 s after the one before, nor a proc from one, since seal casts reset the swing timer with no extra
attack event and two swings can land 13 ms apart. Every white swing left is 3.67 s to 3.71 s after the last.

Every wielder found:

| Report           | Class   | Extra attacks                         | Procs / landed hits | PPM  | Used |
| ---------------- | ------- | ------------------------------------- | ------------------- | ---- | ---- |
| hWHgAPGmrNpDxwCR | paladin | none                                  | 32 / 889            | 0.58 | yes  |
| FMcQw7KgbLjhy3CN | paladin | none                                  | 34 / 881            | 0.63 | yes  |
| kq4Zc1f6GR3AgCKx | paladin | the same run uploaded again           |                     |      | no   |
| VnD3b8WCJQfxr4XZ | paladin | none                                  | 41 / 799            | 0.83 | yes  |
| KcCfFTjmbn1waBHN | paladin | Windfury 31                           | 48 / 1220           | 0.64 | no   |
| gw3TPrnfz6bAYLQj | warrior | Sword Specialization 21               | 9 / 334             | 0.44 | no   |
| pJRYTAHMthmqLVBb | warrior | Windfury 194, Sword Specialization 38 | 28 / 911            | 0.50 | no   |
| yxhmCYJ6dTbjF9Xg | warrior | Windfury 187, Sword Specialization 86 | 32 / 1112           | 0.47 | no   |
| levelling log    | warrior | could not be checked                  | 23 / 587            | 0.64 | no   |

After the cuts:

| Player     | Haste windows | Events cut | Procs / landed hits | Per hit | PPM                 |
| ---------- | ------------- | ---------- | ------------------- | ------- | ------------------- |
| First run  | 5             | 23         | 32 / 878            | 3.64%   | 0.59                |
| Second run | 17            | 317        | 30 / 797            | 3.76%   | 0.61                |
| Third run  | 29            | 603        | 29 / 658            | 4.41%   | 0.71                |
| Together   |               |            | 91 / 2333           | 3.90%   | 0.63 (0.50 to 0.76) |

White swings alone read 0.68 (39 / 926) and Holy Strike, seal and judgement hits 0.58 (50 / 1407); seal procs and
judgements roll it (Judgement of Command 13, Seal of Command 12). The row takes 0.6.

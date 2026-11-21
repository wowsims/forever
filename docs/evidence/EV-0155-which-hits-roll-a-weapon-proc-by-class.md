# EV-0155: Which hits roll a weapon proc, by class

EV-0156 splits the reading that a Judgement of the Crusader, Justice, Light or Wisdom landing rolls a weapon proc once:
it holds to 2026-11-13, and since that day's hotfix the landing rolls nothing. The section After hotfix push 114313
dates the Execute reading: Execute rolls twice up to that push (2026-11-20), and the first hours after it read
between one and two rolls on few wearers. Judgement of Command still rolls twice after it.

- Claim: Execute rolls a weapon proc twice, on both hand setups: with killing blows left out, Execute units hold 1.95
  times the procs one roll gives. Judgement of Command rolls twice, 1.85 [1.78, 1.93]. The
  landings of Serpent Sting, Concussive Shot and Scorpid Sting roll the ranged weapon once. A dual wielder's Whirlwind
  rolls each hand's weapon once. Sap and Death Mark roll nothing. Windfury Weapon and Windfury Totem hits roll once;
  imbue strikes and poisons add no roll. A weapon proc's melee-table strike that a main-hand white swing casts rolls the
  main-hand weapon once more, Loksey's Training Stick's Tough Love and the Rend procs among them; magic procs never do.
  Mongoose Bite and Mocking Blow are not settled: thin samples here read above one roll, a larger separate count reads
  one. Summon Hawk's hits roll the ranged weapon once.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every indexed wearer run of the 24 census weapons: 2,670
  runs in 2,361 reports, too many to list; the examples below name report, fight and time
- Game build: 1.60.1.70790 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114196; the hotfix
  each player ran is not recorded
- Date: 2026-11-05 to 2026-11-19; the section After hotfix push 114313 has its own build and dates
- Pull request: #1108; #1151 added the section After hotfix push 114313

## Method

The census reads the streams of every wearer of 24 weapons whose procs log within a known delay of their hit (Archeus,
Barbaric Crossbow, Bleeding Crescent, Bloodpike, Bloodrazor, Bloodspiller, Burning War Axe, Cursed Felblade, Darkwater
Talwar, Electrocutioner Leg, Fiery War Axe, Flesh Piercer, Galgann's Fireblaster, Glutton's Cleaver, Ironforge
Greathammer, Kang the Decapitator, Loksey's Training Stick, Plaguefang, Scorpion Sting, Strike of the Hydra, Sword of
Corruption, The Ziggler, Verigan's Fist and Wolfsbane). Five weapons stay out: Thornspike, Spellsever Crossbow and
Venomstrike, whose procs land up to 3 s late; Naga Heartpiercer, which rolls at the shot's cast; and Lord General's
Sword, a self buff with no target.

A unit is the wearer's and the wearer's pets' hits, debuff landings and ticks on one target (NPC id and instance), with
gaps of at most 100 ms, longer for a proc with a missile. A proc counts for a unit when it falls from 30 ms before the
unit's first event to 50 ms after its last (400 ms for the missile procs, 60 ms for the ranged ones). Units are grouped
by class and hand: main hand with no off-hand weapon, a dual wielder's main hand, off hand, ranged. Druids in a form
make a group of their own. A run logged in two reports counts once, a stretch with no proc where several were due is
dropped as a swap, and a run with 60 or more counted hits and no proc is dropped.

For each kind of unit the census compares the units with a proc against the count one roll would give at the white
swing's chance: the same class's on the same weapon and hand, or, where the class has too few lone white swings, the
chance pooled over every class but paladins and druids in a form. A ratio of 1 is one roll, about 2 is two rolls and 0
is none. The interval is a wearer bootstrap.

Scripts and output, in `forever-double-roll-1119`: `extract.py`, `units.py` and `report.py` (`report_out.md`, the full
table), `pooled.py` (`pooled_out.md`), `wf.py` (`wf_out.md`), `chain.py` (`chain_out.md`), `jud.py` (`jud_out.md`) and
`examples.py` (`examples_out.md`).

## Lone units of one event

W is a landed white swing, A an ability's damage event, D a debuff landing with no damage event, T a periodic tick. The
white swing rows are the reference each class's other rows are compared with. A row is thin where one roll would give
fewer than 50 units with a proc.

| Class        | Hand                 | Unit                      | Wearers |   Units | With a proc | Two or more | One roll would give | Ratio [95%]       | Reading            |
| ------------ | -------------------- | ------------------------- | ------: | ------: | ----------: | ----------: | ------------------: | ----------------- | ------------------ |
| Warrior      | main hand            | W                         |     394 | 157,683 |      15,856 |          55 |              15,816 | 1.00 [1.00, 1.00] | one roll           |
| Warrior      | main hand            | D: Sunder Armor           |     316 |  25,720 |       2,565 |           0 |               2,574 | 1.00 [0.96, 1.04] | one roll           |
| Warrior      | dual-wield main hand | D: Sunder Armor           |      34 |     781 |          54 |           0 |                  74 | 0.73 [0.59, 0.93] | below one roll     |
| Warrior      | main hand            | D: Rend                   |     340 |  19,908 |       2,031 |           0 |               2,026 | 1.00 [0.96, 1.05] | one roll           |
| Warrior      | main hand            | A: Heroic Strike          |     315 |  13,055 |       1,400 |           0 |               1,365 | 1.03 [0.97, 1.09] | one roll           |
| Warrior      | main hand            | A: Execute                |      79 |   1,388 |         234 |          13 |                 133 | 1.76 [1.54, 2.00] | two rolls          |
| Warrior      | dual-wield main hand | A: Execute                |      44 |   1,233 |         221 |           8 |                 104 | 2.12 [1.80, 2.41] | two rolls          |
| Warrior      | off hand             | A: Execute                |      17 |     385 |           0 |           0 |                  20 | 0.00 [0.00, 0.00] | none (thin)        |
| Warrior      | main hand            | A: Whirlwind (two-hander) |      11 |   1,227 |         108 |           0 |                 124 | 0.87 [0.63, 0.97] | below one roll     |
| Warrior      | dual-wield main hand | A: Whirlwind, both hits   |      19 |   3,607 |         220 |           0 |                 248 | 0.89 [0.77, 1.02] | one roll           |
| Warrior      | off hand             | A: Whirlwind, both hits   |      12 |   2,305 |         132 |           0 |                 125 | 1.05 [0.91, 1.19] | one roll           |
| Warrior      | main hand            | A: Mocking Blow           |      57 |     130 |          18 |           0 |                  13 | 1.39 [0.81, 1.99] | unresolved (thin)  |
| Warrior      | main hand            | A: Thunder Clap           |     382 |  96,683 |           2 |           0 |               9,855 | 0.00 [0.00, 0.00] | none               |
| Warrior      | main hand            | A: Shield Bash            |     215 |   3,093 |           0 |           0 |                 293 | 0.00 [0.00, 0.00] | none               |
| Warrior      | main hand            | A: Shield Slam            |      16 |   2,414 |           0 |           0 |                 203 | 0.00 [0.00, 0.00] | none               |
| Rogue        | dual-wield main hand | W                         |     176 |  37,329 |       3,237 |           3 |               3,237 | 1.00 [1.00, 1.00] | one roll           |
| Rogue        | dual-wield main hand | A: Sinister Strike        |     167 |  29,542 |       2,563 |           0 |               2,510 | 1.02 [0.96, 1.09] | one roll           |
| Rogue        | off hand             | A: Sinister Strike        |      59 |  14,175 |           0 |           0 |                 949 | 0.00 [0.00, 0.00] | none               |
| Rogue        | dual-wield main hand | A: Shank                  |     140 |  10,592 |           0 |           0 |                 908 | 0.00 [0.00, 0.00] | none               |
| Rogue        | dual-wield main hand | D: Sap                    |      39 |     102 |           0 |           0 |                   9 | 0.00 [0.00, 0.00] | none (thin)        |
| Rogue        | dual-wield main hand | D: Death Mark             |       4 |     127 |           0 |           0 |                  11 | 0.00 [0.00, 0.00] | none (thin)        |
| Paladin      | main hand            | A: Judgement of Command   |     274 |  17,148 |       3,524 |         202 |               1,906 | 1.85 [1.78, 1.93] | two rolls          |
| Hunter       | main hand            | A: Raptor Strike          |     143 |   1,921 |         141 |           0 |                 144 | 0.98 [0.80, 1.14] | one roll           |
| Hunter       | main hand            | A: Mongoose Bite          |      39 |     128 |          17 |           0 |                  10 | 1.75 [0.89, 2.51] | unresolved (thin)  |
| Hunter       | ranged               | A: Multi-Shot             |      79 |  11,639 |       1,723 |           0 |               1,748 | 0.99 [0.94, 1.03] | one roll           |
| Hunter       | ranged               | A: Arcane Shot            |      80 |   5,556 |         734 |           0 |                 702 | 1.05 [0.98, 1.12] | one roll           |
| Hunter       | ranged               | A: Aimed Shot             |      67 |   1,718 |         237 |           0 |                 236 | 1.00 [0.88, 1.11] | one roll           |
| Hunter       | ranged               | D: Serpent Sting          |      91 |   6,166 |         868 |           0 |                 805 | 1.08 [1.01, 1.15] | one roll (above 1) |
| Hunter       | ranged               | D: Concussive Shot        |      42 |     296 |          44 |           0 |                  38 | 1.15 [0.82, 1.47] | one roll (thin)    |
| Hunter       | ranged               | D: Scorpid Sting          |      11 |     209 |          22 |           0 |                  30 | 0.72 [0.00, 1.02] | one roll (thin)    |
| Hunter       | ranged               | T: Serpent Sting          |      91 |  15,289 |           0 |           0 |               2,026 | 0.00 [0.00, 0.00] | none               |
| Hunter       | ranged               | A: Volley                 |      44 |  31,681 |          66 |           0 |               5,501 | 0.01 [0.00, 0.04] | none               |
| Shaman       | main hand            | W                         |     207 |  24,297 |       2,540 |          43 |               2,540 | 1.00 [1.00, 1.00] | one roll           |
| Shaman       | main hand            | A: Stormstrike            |      82 |   5,250 |         382 |           0 |                 380 | 1.01 [0.88, 1.13] | one roll           |
| Druid (form) | main hand            | W                         |     301 |  70,330 |       8,018 |           2 |               8,018 | 1.00 [1.00, 1.00] | one roll           |
| Druid (form) | main hand            | A: Maul                   |     193 |   7,193 |           0 |           0 |                 868 | 0.00 [0.00, 0.00] | none               |
| Druid (form) | main hand            | A: Swipe (Bear)           |     126 |  28,568 |           2 |           0 |               2,985 | 0.00 [0.00, 0.00] | none               |

A dual wielder's Sunder Armor and a two-hander's Whirlwind read below one roll. Both are counted as one roll in
#1108's rates (EV-0153 §Method); neither is settled. Serpent Sting's landing reads 1.08 [1.01, 1.15]: a little above
one roll and far below two.

A Serpent Sting, Concussive Shot or Scorpid Sting landing rolls the ranged weapon with no damage event, so a count that
takes it for no roll credits its proc to nothing: over every hunter unit on a ranged weapon, the units with such a
landing hold 13.89 [11.74, 16.33] times the procs no roll would leave them (`pooled_out.md`). Execute's off-hand row
shows that both of its rolls fall on the main-hand weapon. Judgement of the Crusader, Justice, Light and Wisdom landings
roll once: 1,221, 44, 46 and 45 units with a proc against 1,160, 52, 40 and 34 for one roll and 2,189, 97, 77 and 65 for
two (`jud_out.md`).

## Killing blows

The census counts a killing blow as a unit like any other hit, but a proc that lands on the target logs almost never on
one: over the 45 weapons #1108 counted (EV-0153 §Method), 3.7% of the white swings that leave the target alive are
followed by a proc and 0.7% of the killing ones, while self-buff procs (Destiny, Lord General's Sword, Sang'thraze the
Deflector) log on a kill as usual. So a unit kind that often kills reads low here. Execute does most of all: its rows
above read 1.76 with no off-hand weapon and 2.12 dual wielding, and with killing blows left out, over #1108's weapons
(EV-0153 §Method), 3,683 lone Execute units hold 553 with a proc where one roll at the same weapon's white chance gives
284, 1.95 times. A separate count on the reports up to 2026-11-19, with killing blows left out (#1108's review), reads
Execute at 1.90 [1.70, 2.08] with no off-hand weapon and 1.94 [1.73, 2.16] dual wielding, Mocking Blow at 0.97 [0.66,
1.29] on 346 units, Mongoose Bite at 0.95 and 1.13 on 189 and 162 units, and Summon Hawk's hits at 1.05 [0.95, 1.13] on
4,140 units.

## After hotfix push 114313

Hotfix push 114313 sets Suppress Weapon Procs on Execute's cast rows 5308, 20658, 20660, 20661 and 20662, takes Suppress
Caster Procs off them, and sets Suppress Weapon Procs on Judgement 20271 and on Mutilate's cast rows 399956, 1241582 and 1241584. It changes no Judgement of Command row and no Mutilate strike row.

- Game build: 1.60.1.71037. The push is missing from the game files' hotfix cache of 2026-11-20 22:16 UTC and present
  in the one of 23:47 UTC. Report 8vY4qHxyMNWd36aK, one warrior's log with no gap, shows the switch between 23:35:49
  and 00:01:37 UTC, with no restart. The hotfix each player ran is not recorded.
- Date: before is 2026-11-20 up to 22:16 UTC; after is 2026-11-20 23:47 UTC to 2026-11-21 08:05 UTC, about eight hours
  of play in 1,390 reports. Events between 22:16 and 23:47 count on neither side.

The counts use this file's method, killing blows left out, and white chances from the same class in the same period.
Execute's rows take the census weapons and #1108's weapons (EV-0153 §Method) whose procs land on the target within a
known delay; Bigger Rock, whose strike chains, is a row of its own.

| Unit                                      | Period            | Wearers |  Units | With a proc | One roll would give | Ratio [95%]       | Reading   |
| ----------------------------------------- | ----------------- | ------: | -----: | ----------: | ------------------: | ----------------- | --------- |
| Paladin, A: Judgement of Command          | to 2026-11-19     |     274 | 17,148 |       3,524 |               1,903 | 1.85 [1.77, 1.93] | two rolls |
| Paladin, A: Judgement of Command          | 2026-11-20 before |      65 |  4,443 |         908 |                 462 | 1.96 [1.78, 2.15] | two rolls |
| Paladin, A: Judgement of Command          | after             |      43 |  2,400 |         431 |                 247 | 1.75 [1.55, 2.00] | two rolls |
| Paladin, D: Judgement of the Crusader + W | after             |      39 |  1,158 |         132 |                 125 | 1.06 [0.89, 1.24] | one roll  |
| Warrior, A: Execute                       | to 2026-11-19     |     122 |  2,831 |         455 |                 240 | 1.90 [1.72, 2.07] | two rolls |
| Warrior, A: Execute                       | 2026-11-20 before |      10 |    188 |          24 |                14.3 | 1.68 [1.10, 2.50] | thin      |
| Warrior, A: Execute                       | after             |       9 |     73 |           7 |                 5.8 | 1.20 [0.48, 2.00] | thin      |
| Warrior, A: Execute, Bigger Rock          | 2026-11-20 before |       5 |    148 |          23 |                17.0 | 1.35 [0.99, 1.63] | thin      |
| Warrior, A: Execute, Bigger Rock          | after             |       2 |     78 |           9 |                 7.6 | 1.19 [0.54, 2.25] | thin      |

Judgement of Command reads two rolls after the push in every hour from 00:00 to 07:00 UTC (1.6 to 1.9), and the 19
paladins with units on both sides read 2.06 [1.82, 2.43] after it. 25 lone Judgement of Command units after it hold two
procs, against 26.6 for two rolls; 0 of 4,131 lone white swings do. The Judgement of the Crusader row counts Bigger
Rock's wearers too. Execute's two rows after the push pool to 1.19 on 11 warriors, Poisson interval 0.68 to 1.94: one
roll fits and two sit at the edge. The readings before the push on 2026-11-20 already sit between one and two, so the
logs do not yet show Execute's change.

Windfury Totem off Execute, with the totem in range (another Windfury Totem hit of the warrior within 5 s) and no
overkill, against the same warrior's Heroic Strike, Bloodthirst, Mortal Strike, Revenge, Overpower, Slam and Hamstring.
Two rolls read about 1.81:

| Period            | Ranks | Warriors | Landed Executes struck | Ratio [95%]       | Avoided Executes struck |
| ----------------- | ----- | -------: | ---------------------: | ----------------- | ----------------------: |
| 2026-11-14 to 15  | 1–4   |       45 |                217/664 | 1.73 [1.53, 2.00] |                   17/77 |
| 2026-11-14 to 15  | 5     |       12 |                 33/203 | 0.99 [0.65, 1.46] |                    0/16 |
| 2026-11-20 before | 1–4   |       48 |                108/558 | 1.01 [0.82, 1.24] |                    0/83 |
| 2026-11-20 before | 5     |       64 |              227/1,266 | 1.05 [0.90, 1.19] |                   0/105 |
| after             | 1–4   |       39 |                153/480 | 1.68 [1.38, 1.98] |                   10/74 |
| after             | 5     |       50 |              507/1,350 | 2.09 [1.89, 2.31] |                   18/97 |

Every rank's cast rolls the totem again after the push, an avoided Execute included; leaving out one warrior at a time
keeps ranks 1–4 at 1.60 to 1.74. Shield Slam and Shield Bash drew no Windfury hit on 2026-11-20 on either side, so the
reference leaves them out.

Mutilate, rogues with a dagger in each hand, 2026-11-14 to 2026-11-20 22:15 UTC before the push and 23:47 UTC to
2026-11-21 08:49 UTC after it. A unit is one Mutilate's main-hand and off-hand strike, both landed, with no other hit or
debuff of the rogue on that target within 100 ms and no killing blow. Each poison is put on the hand whose white swings
roll it for that rogue, and the reference is the same rogue's lone white swing of that hand in the same period. Two rolls
read about 1.67 for Instant Poison, 1.6 for Deadly Poison and 1.9 for weapon procs; a refreshed Deadly Poison stack at
its maximum often logs nothing.

| Mutilate unit, hand       | Period | Rogues |  Units | With a proc | One roll would give | Ratio [95%]       | Reading   |
| ------------------------- | ------ | -----: | -----: | ----------: | ------------------: | ----------------- | --------- |
| Instant Poison, main hand | before |     90 | 14,713 |       7,843 |               4,680 | 1.68 [1.63, 1.72] | two rolls |
| Instant Poison, main hand | after  |     22 |  3,290 |       1,087 |               1,052 | 1.03 [1.00, 1.07] | one roll  |
| Deadly Poison, main hand  | before |     33 |  4,898 |       3,030 |               1,908 | 1.59 [1.53, 1.65] | two rolls |
| Deadly Poison, main hand  | after  |      2 |    249 |         109 |                91.3 | 1.19 [0.98, 1.44] | thin      |
| Deadly Poison, off hand   | before |     93 | 15,164 |       6,487 |               6,210 | 1.04 [1.02, 1.07] | one roll  |
| Deadly Poison, off hand   | after  |     22 |  3,166 |       1,320 |               1,293 | 1.02 [0.98, 1.07] | one roll  |
| Weapon procs, main hand   | before |      9 |  1,107 |         115 |                49.0 | 2.35 [1.96, 3.03] | two rolls |
| Weapon procs, main hand   | after  |      1 |     85 |           5 |                 6.1 | 0.82 [0.27, 1.92] | thin      |
| Weapon procs, off hand    | before |      9 |  1,119 |          57 |                54.9 | 1.04 [0.59, 1.57] | one roll  |
| Weapon procs, off hand    | after  |      2 |    434 |          31 |                29.7 | 1.04 [0.71, 1.48] | one roll  |

The 5 rogues with main-hand Instant Poison on both sides read 1.58 [1.45, 1.79] before the push and 1.05 [0.98, 1.08]
after it. Windfury Totem off a landed Mutilate, against the same rogue's other main-hand specials and Eviscerate, reads
one roll on both sides: 1.16 [1.04, 1.28] on 83 rogues before and 0.94 [0.79, 1.13] on 15 after, where two rolls read
about 1.8. A dodged, parried or missed Mutilate cast is followed by no Windfury hit, 0 of 925 before and 0 of 136 after,
so the cast never rolls the totem. Mutilate's reports are in `codes_mut_before.txt` (188), `codes_mut_after.txt` (29),
`codes_mutwf_before.txt` (121) and `codes_mutwf_after.txt` (24); its scripts are `mut.py`, `mut_report.py`
(`mut_out.md`) and `totem_mut.py` (`totem_out_mutwf.md`).

Reports after the push, Judgement of Command: 1RYWd3fA9tJMPL6G, 1wbzNyLhKQvMGTtm, 28AFzYhbjwZGqx1T, 2d69FD7QXBaVmNtH, 38jkDhQAVGYvpwqc, 3rVjZ4gfbQm12Y6F, 4VdfDNBWZvk8pJnH, 6YqkW3zmJLKNBMnj, 8Dd1whVHJcm4XnRx, 9AVbQBj3Ckq12cTL, APjaWhT6kfvRxZb8, AbBcj4dL7yXVzCft, BtJP21rynT4kQY7C, D1RGNCTjcJxvYd6a, G3TNcd2vzx89a7fj, GHn7VRAP9mjLDTc1, Gm4f9LFAZrWzTJdC, GvBdKFbyLQnH76Jz, HhpZmgKMfXjzWwrk, JdVwfvqRHcxyaF14, Lh87FKNJvRX6Ccpf, MPWQdq86nXbwk7tR, NKHzt9GyVJMd7x3j, NWVCbLJK3taTmkfQ, NxKhn6PDpmg8WF4G, PKVJGwaRHz1qfYZx, PT7AtvXWkrQgLCpn, PqFpXZx4NQ2Mm81v, RphZTjABx9cYDwJF, T9fhDpRaFzGHP4By, VKPvXq7kJwcZGW6L, W9nck8JraRvxBDF4, WBXnyNJ46VYjA2Qm, X93jK7N6aALMpWnD, YB93KnL7VMdtJWH4, YCrzxNT1hWZJQvBL, YdzQHMp4GhrJ2A7n, Z48VBnhdrxPTMYyz, Z8Fx4J1c6htrT7fq, ZC73QxMhdPrpTavb, aV4L8grZWvDRz3MF, bHkgvJcQaRpXYB3y, cj32G8gw9pvAZdxk, hMadZHPw4cA6XLJW, jAdhbnXrCDJfvmQF, jFxXVgMTftpPGDQw, jNLqmbwFYt16JT3v, k96y3KTD4qtpLNrC, ktRfG63rXBLdPcpA, mHFdMA8zXTaxLDyP, mQv7YCGNHwkAMJ6d, mprAaDbTyH7Pvj9R, nrNzAVCtHKfPXdpk, pA1nYfC3BF7vDxJL, pMGyN9hFa8XRAHxV, prw16bXnmtg8cLya, q167YKJGNhjpZcCz, rMkqF3LHcbPvZNwt, rqc2pdNPg4Xyh3nt, wf3Vdy7DxXYgKhMv, wrKpnNaxqj2ydFAM, yRK2MZNanDbAdq1Q, z1dHvNPABL9Kba8J, zCMpYhkZQt7PKanf, zLFgy8AV3JxbHh4w. Execute's weapon procs: 1RYWd3fA9tJMPL6G, 3ygCHfWnAqdmRbrF, 9HCnBrxmMAz8tvjW, DLApcY8P6y2xWqRh, FbhmQgZj9TH7f84t, HaqMK2rTjDn6AY1C, NTbRAz19k3DMZ7Ka, TwYLbd7Fj9NxgRMv, dbXaAxv1cTVMh49j, fmR1tDMXHzTpjPJw, k4v2tVp6gcJRw8fW. The totem
rows name 126 reports after the push, in `codes_post_totem_r2.txt`; fights are not listed.

Scripts and output, in `forever-double-roll-post114313`: `extract_r2.py`, `units_r2.py` and `report_r2.py`
(`report_r2_period.md`), `joc_extra_r2.py` (`joc_extra_out_r2.md`), and `totem_ratio.py` and `totem_ranks.py`
(`totem_ratio_out_r2.md`, `totem_ranks_out_r2.txt`).

## Windfury hits, imbues and poisons

Units of one white swing with one Windfury hit, against one roll per hit (`wf_out.md`):

| Class   | Unit                      | Wearers | Units | With a proc | One roll per hit | A Windfury hit rolling twice | Ratio [95%]       |
| ------- | ------------------------- | ------: | ----: | ----------: | ---------------: | ---------------------------: | ----------------- |
| Warrior | white + Windfury Totem    |     118 | 4,734 |         888 |            881.5 |                      1,255.4 | 1.01 [0.95, 1.07] |
| Rogue   | white + Windfury Totem    |      42 |   858 |         143 |            150.5 |                        214.5 | 0.95 [0.78, 1.11] |
| Shaman  | white + Windfury Totem    |      59 | 1,312 |         176 |            170.9 |                        246.6 | 1.03 [0.92, 1.14] |
| Shaman  | white + 2 Windfury Weapon |      71 | 1,835 |         321 |            335.3 |                        519.6 | 0.96 [0.84, 1.08] |
| Shaman  | white + 1 Windfury Weapon |      63 |   355 |          47 |             45.7 |                         65.9 | 1.03 [0.77, 1.30] |

Units holding an event that rolls nothing, against the count with that event as no roll (`pooled_out.md`; 1 adds no
roll):

| Class   | Hand      | Event            | Wearers |   Units | With a proc | Ratio [95%]       |
| ------- | --------- | ---------------- | ------: | ------: | ----------: | ----------------- |
| Shaman  | main hand | imbue strike     |     179 |  47,193 |       5,787 | 1.04 [0.97, 1.10] |
| Warrior | main hand | imbue strike     |      53 |   6,438 |         732 | 1.01 [0.93, 1.08] |
| Rogue   | main hand | poison           |     142 |  32,691 |       2,632 | 1.04 [0.94, 1.11] |
| Rogue   | off hand  | poison           |      44 |  12,708 |         452 | 1.12 [0.97, 1.27] |
| Warrior | main hand | Deep Wound hit   |      94 |  10,032 |         244 | 1.12 [0.97, 1.26] |
| Warrior | main hand | Sweeping Strikes |      33 |   3,612 |         101 | 0.92 [0.79, 1.13] |
| Warrior | main hand | damage shield    |     295 | 206,208 |       2,269 | 0.98 [0.92, 1.03] |
| Warrior | main hand | periodic tick    |     382 |  98,702 |       1,439 | 1.05 [1.00, 1.11] |
| Hunter  | main hand | pet hit          |     285 | 447,373 |         653 | 0.95 [0.86, 1.04] |

## The strike chain

Lone white swings of each weapon, the ones with two procs against the count a chained strike rolling once more at the
weapon's own chance would give (`chain_out.md`; paladins and druids in a form left out):

| Weapon                  | Proc             | Hand      | Wearers | Lone whites | With a proc | With two | A chain would give |
| ----------------------- | ---------------- | --------- | ------: | ----------: | ----------: | -------: | -----------------: |
| Loksey's Training Stick | Tough Love       | main hand |     168 |       9,473 |         993 |       82 |                104 |
| Glutton's Cleaver       | Rend             | main hand |      17 |       5,339 |         338 |       20 |                 21 |
| Kang the Decapitator    | Rend             | main hand |      14 |       4,618 |         246 |       10 |                 13 |
| Bloodrazor              | Rend             | main hand |      15 |       4,671 |         211 |        8 |                 10 |
| Bleeding Crescent       | Rend             | main hand |      19 |       3,373 |         113 |        6 |                  4 |
| Flesh Piercer           | Rend             | main hand |       8 |       2,833 |          72 |        3 |                  2 |
| Glutton's Cleaver       | Rend             | off hand  |      22 |       6,911 |         443 |        0 |                 28 |
| Electrocutioner Leg     | Lightning Bolt   | main hand |     188 |      93,846 |      10,202 |        0 |              1,109 |
| Ironforge Greathammer   | Firebolt         | main hand |     204 |      24,666 |       3,030 |        0 |                372 |
| Cursed Felblade         | Cripple          | main hand |     109 |      45,625 |       3,367 |        0 |                248 |
| Darkwater Talwar        | Shadow Bolt      | main hand |     115 |      18,337 |       1,901 |        2 |                197 |
| Strike of the Hydra     | Corrosive Poison | main hand |      88 |       8,658 |         854 |        0 |                 84 |

## Examples

Report, fight and time into the fight, class and weapon (`examples_out.md`):

- Serpent Sting landing alone, with a proc: 1AfDWYQT7Xjd24K3 13 2:15.921 hunter, Barbaric Crossbow;
  1PVRtQZCwfTkmzx9 1 1:29.386 hunter, Galgann's Fireblaster; 1W9N26dHfnPV4G7A 1 3:58.649 hunter, Barbaric Crossbow.
- Concussive Shot landing alone, with a proc: 1y6GV7KLTZPrJH2w 1 0:45.837 hunter, Barbaric Crossbow; 92JNtax7RTkQ6cHv
  1 6:33.773 hunter, Barbaric Crossbow; BA6gQRpLNtcMJXTD 9 0:29.559 hunter, Galgann's Fireblaster.
- Auto Shot and Serpent Sting landing, two procs: 9XQfBmZRaNWJkH3F 1 10:22.123, Fb8cX76vJCYhZftV 1 0:06.803 and
  RXDCK9PtZy3Jp2QA 1 5:10.184, hunter, Galgann's Fireblaster.
- Mongoose Bite alone, with a proc: 1YQTjgZh2KPBGCbN 1 9:22.725 hunter, Bloodspiller; fd4MgwavrqWcYHZG 13 7:46.182
  and 2PpAjf1qRmTv6gDz 1 2:21.257 hunter, Loksey's Training Stick.
- A dual wielder's Whirlwind, a proc on the off-hand weapon: 9VacfmDpXwCgzZWq 1 1:40.512 warrior, Glutton's Cleaver;
  Br2pycwj4zLC1ZP7 1 0:36.751 warrior, Scorpion Sting; CnNwpaYW83cJr6fR 1 1:08.762 warrior, Glutton's Cleaver.
- A lone white swing with two Tough Love: DWJhjnmAM6GZdbKv 1 6:16.203 hunter; 1xqBcj4CZQXV38Pv 1 22:50.566 warrior;
  tHz9M3arKp6PxA1W 7 21:30.891 hunter; each Loksey's Training Stick.
- A lone white swing with two Rend procs: Gx3AXVk9cCaR18Df 9 23:56.586 warrior, Jad2wbYTXQjBc1n7 1 47:28.241 warrior
  and f8bZ6ytgj7mMGHCk 11 15:38.790 shaman, each Glutton's Cleaver.
- Execute alone, two procs: 4ktfmxnqcbv61B2L 1 15:03.432 warrior, Darkwater Talwar; H3bRaDmzyjxTW2BJ 1 9:09.428
  warrior, Strike of the Hydra; 6wxzV4YyNCh1aqfd 1 28:30.342 warrior, Kang the Decapitator.
- A missed Execute alone, with a proc: 1xqBcj4CZQXV38Pv 1 8:26.401 warrior, Loksey's Training Stick; 1a3pDtCcPgBkL6wm
  1 18:31.732 warrior, Strike of the Hydra; nqmgcYrPz9B1fLDx 1 26:49.602 warrior, Electrocutioner Leg.

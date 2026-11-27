# EV-0155: Which hits roll a weapon proc, by class

- Claim: Execute and Judgement of Command each roll the main-hand weapon once and equip procs once. A Mutilate rolls the
  main-hand weapon once and the off-hand weapon once. The landings of Serpent Sting, Concussive Shot and Scorpid Sting
  roll the ranged weapon once. A dual wielder's Whirlwind rolls each hand's weapon once. Sap and Death Mark roll
  nothing. Windfury Weapon and Windfury Totem hits roll once; imbue strikes and poisons add no roll. A weapon proc's
  melee-table strike that a main-hand white swing casts rolls the main-hand weapon once more, Loksey's Training Stick's
  Tough Love and the Rend procs among them; magic procs never do. Mongoose Bite and Mocking Blow are not settled: thin
  samples here read above one roll, a larger separate count reads one. Summon Hawk's hits roll the ranged weapon once. A
  rogue's Shank rolls the off-hand weapon and equip procs once, the main-hand weapon never and Windfury Totem once;
  Torturing Poker, which procs only from its own hand's hits, never procs from it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), every indexed wearer run of the 24 census weapons: 2,670
  runs in 2,361 reports, too many to list; the examples below name report, fight and time
- Game build: 1.60.1.70790 to 1.60.1.70959 by report date; the sim's database is 1.60.1.70959, push 114196; the hotfix
  each player ran is not recorded
- Date: 2026-11-05 to 2026-11-19; the sections After hotfix push 114313, Judgement of Command after hotfix push 114399
  and Count of the 2026-11-20 to 2026-11-27 reports have their own build and dates
- Pull request: #1108; #1151 added the section After hotfix push 114313, #1208 the section Judgement of Command after
  hotfix push 114399
- Later count: Shank, Execute, Judgement of Command, Mutilate, §Count of the 2026-11-20 to 2026-11-27 reports

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
table), `pooled.py` (`pooled_out.md`), `wf.py` (`wf_out.md`), `chain.py` (`chain_out.md`) and `examples.py`
(`examples_out.md`).

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
| Rogue        | dual-wield main hand | D: Sap                    |      39 |     102 |           0 |           0 |                   9 | 0.00 [0.00, 0.00] | none (thin)        |
| Rogue        | dual-wield main hand | D: Death Mark             |       4 |     127 |           0 |           0 |                  11 | 0.00 [0.00, 0.00] | none (thin)        |
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
landing hold 13.89 [11.74, 16.33] times the procs no roll would leave them (`pooled_out.md`). A Judgement of the
Crusader, Justice, Light or Wisdom landing rolls nothing (EV-0156).

## Killing blows

The census counts a killing blow as a unit like any other hit, but a proc that lands on the target logs almost never on
one: over the 45 weapons #1108 counted (EV-0153 §Method), 3.7% of the white swings that leave the target alive are
followed by a proc and 0.7% of the killing ones, while self-buff procs (Destiny, Lord General's Sword, Sang'thraze the
Deflector) log on a kill as usual. So a unit kind that often kills, Execute most of all, reads low in a count that keeps
them. A separate count on the reports up to 2026-11-19, with killing blows left out (#1108's review), reads Mocking
Blow at 0.97 [0.66, 1.29] on 346 units, Mongoose Bite at 0.95 and 1.13 on 189 and 162 units, and Summon Hawk's hits at
1.05 [0.95, 1.13] on 4,140 units.

## After hotfix push 114313

Hotfix push 114313 sets Suppress Weapon Procs on Execute's cast rows 5308, 20658, 20660, 20661 and 20662, takes Suppress
Caster Procs off them, and sets Suppress Weapon Procs on Judgement 20271 and on Mutilate's cast rows 399956, 1241582 and 1241584. It changes no Judgement of Command row and no Mutilate strike row.

- Game build: 1.60.1.71037. The push is missing from the game files' hotfix cache of 2026-11-20 22:16 UTC and present
  in the one of 23:47 UTC. The hotfix each player ran is not recorded.
- Date: 2026-11-20 23:47 UTC to 2026-11-21 08:05 UTC, about eight hours of play in 1,390 reports.

The counts use this file's method, killing blows left out, and white chances from the same class in the same period. A
paladin's Judgement of the Crusader landing beside a white swing reads one roll, the swing's: 132 of 1,158 units hold a
proc on 39 paladins, Bigger Rock's wearers among them, where one roll gives 125, 1.06 [0.89, 1.24].

Windfury Totem off Execute, with the totem in range (another Windfury Totem hit of the warrior within 5 s) and no
overkill, against the same warrior's Heroic Strike, Bloodthirst, Mortal Strike, Revenge, Overpower, Slam and Hamstring.
Two rolls read about 1.81:

| Ranks | Warriors | Landed Executes struck | Ratio [95%]       | Avoided Executes struck |
| ----- | -------: | ---------------------: | ----------------- | ----------------------: |
| 1–4   |       39 |                153/480 | 1.68 [1.38, 1.98] |                   10/74 |
| 5     |       50 |              507/1,350 | 2.09 [1.89, 2.31] |                   18/97 |

Every rank's cast rolls the totem as well as its strike, an avoided Execute included; leaving out one warrior at a time
keeps ranks 1–4 at 1.60 to 1.74. Shield Slam and Shield Bash drew no Windfury hit on 2026-11-20, so the reference
leaves them out.

Mutilate, rogues with a dagger in each hand, 2026-11-20 23:47 UTC to 2026-11-21 08:49 UTC. A unit is one Mutilate's
main-hand and off-hand strike, both landed, with no other hit or debuff of the rogue on that target within 100 ms and no
killing blow. Each poison is put on the hand whose white swings roll it for that rogue, and the reference is the same
rogue's lone white swing of that hand in the same period. Two rolls read about 1.67 for Instant Poison and 1.6 for
Deadly Poison; a refreshed Deadly Poison stack at its maximum often logs nothing. A Mutilate's weapon procs are in the
section Count of the 2026-11-20 to 2026-11-27 reports.

| Mutilate unit, hand       | Rogues | Units | With a proc | One roll would give | Ratio [95%]       | Reading  |
| ------------------------- | -----: | ----: | ----------: | ------------------: | ----------------- | -------- |
| Instant Poison, main hand |     22 | 3,290 |       1,087 |               1,052 | 1.03 [1.00, 1.07] | one roll |
| Deadly Poison, main hand  |      2 |   249 |         109 |                91.3 | 1.19 [0.98, 1.44] | thin     |
| Deadly Poison, off hand   |     22 | 3,166 |       1,320 |               1,293 | 1.02 [0.98, 1.07] | one roll |

Windfury Totem off a landed Mutilate, against the same rogue's other main-hand specials and Eviscerate, reads one roll:
1.16 [1.04, 1.28] on 83 rogues from 2026-11-14 to 2026-11-20 22:15 UTC and 0.94 [0.79, 1.13] on 15 after the push, where
two rolls read about 1.8. A dodged, parried or missed Mutilate cast is followed by no Windfury hit, 0 of 925 and 0 of
136, so the cast never rolls the totem. Mutilate's reports are in `codes_mut_after.txt` (29), `codes_mutwf_before.txt`
(121) and `codes_mutwf_after.txt` (24); its scripts are `mut.py`, `mut_report.py` (`mut_out.md`) and `totem_mut.py`
(`totem_out_mutwf.md`).

The totem rows name 126 reports after the push, in `codes_post_totem_r2.txt`; fights are not listed.

Scripts and output, in `forever-double-roll-post114313`: `extract_r2.py`, `units_r2.py` and `report_r2.py`
(`report_r2_period.md`), and `totem_ratio.py` and `totem_ranks.py` (`totem_ratio_out_r2.md`, `totem_ranks_out_r2.txt`).

## Judgement of Command after hotfix push 114399

- Count: §Count of the 2026-11-20 to 2026-11-27 reports

Hotfix push 114399 sets Suppress Weapon Procs on the five Judgement of Command dummies, 20425, 20961, 20962, 20967
and 20968. It changes no Judgement of Command damage row (20467, 20963 to 20966) and no Seal of Command row.

- Game build: 1.60.1.71037 on EU realms and 1.60.1.71091 on US realms. The lone Judgement of Command units fit one
  switch best at 2026-11-25 01:15 UTC (95% profile interval 01:15 to 01:20). The hotfix each player ran is not
  recorded.
- Date: 2026-11-25 01:15 to 06:10 UTC.
- Sources: the reports with paladin units from 2026-11-24 12:00 UTC on that run past the switch; those with lone
  Judgement of Command units are listed below.

The units are this file's lone units on the census weapons, a paladin's main hand, killing blows left out. Rolls per hit
are a maximum likelihood fit: a unit with r rolls at the weapon's lone white chance p holds a proc with chance
1 − (1 − p)^r, with p taken over every paladin's lone white swings in the count. A proc counts whether or not it lands.
The interval is a paladin bootstrap. One roll reads about 1 and two about 2; a white swing beside the hit adds one.

| Unit                             | Paladins | Units | With a proc | Procs | One roll would give | Rolls per hit [95%] | Two or more |
| -------------------------------- | -------: | ----: | ----------: | ----: | ------------------: | ------------------- | ----------: |
| W                                |       37 | 2,880 |         283 |   289 |               280.0 | 1.01 [0.92, 1.13]   |           6 |
| A: Judgement of Command + W      |       26 |   122 |          30 |    32 |                12.9 | 2.53 [1.77, 3.63]   |           2 |
| A: Seal of Command + W           |       32 | 1,090 |         201 |   208 |               109.3 | 1.94 [1.65, 2.26]   |           7 |
| D: Judgement of the Crusader + W |       25 |   579 |          53 |    54 |                61.6 | 0.85 [0.56, 1.02]   |           1 |
| D: Judgement of the Crusader     |       23 |   259 |           0 |     0 |                27.9 | 0.00                |           0 |
| A: Judgement of Righteousness    |       15 |   518 |           0 |     0 |                52.4 | 0.00                |           0 |

The Seal of Command strike beside its white swing reads two rolls. A lone Judgement of Righteousness and a lone
Judgement of the Crusader landing, each with its Judgement cast, hold no proc, so neither the Judgement cast nor those
judgements roll.

Crusader (Holy Strength 20007) and Windfury Totem hits off a lone Judgement of Command after the switch, from the 589
report streams with a paladin from 2026-11-24 12:41 UTC on (760 paladin runs once a player's run logged twice counts
once), with no weapon filter, against each run's own lone white swings. Crusader takes runs with 100 lone white swings
and the enchant seen; Windfury Totem takes runs with 20 lone white swings with the totem in range (another Windfury
Totem hit of the paladin within 5 s outside the unit). The fit is the one above with p the run's own chance per white
swing.

| Proc           | Paladins | Units | With a proc | One roll would give | Rolls per hit [95%] | Reading        |
| -------------- | -------: | ----: | ----------: | ------------------: | ------------------- | -------------- |
| Crusader       |        3 |   347 |          16 |                17.1 | 0.93 [0.77, 1.11]   | one roll, thin |
| Windfury Totem |       28 |   758 |          93 |                79.4 | 1.17 [0.85, 1.51]   | one roll       |

Reports after the switch, Judgement of Command: 2kKA63MnVtd9ZpgC, 3k8pBhVyrtAnq74j, 7RT8QrzcyVxPWBmZ, 8hYCqLbWKwgnk3Nz,
B2RVjm8aFcyYrxP3, ByKpfCMF79xd2b6N, CXNKZaGRcHkg1VLx, DhRyLkKcN4B7pQtM, FzXGfVJr679TcNRB, GQfYzHnc6b3jhmpx,
HDYAPTJ9Cgw2BGvx, HvKhpmJaD63Ytzgk, MHQFABkWN8Rzxrtq, RFx8YV7v6qnG1c9H, T7XtHkAKM2WPB9YJ, VRCNP4y8TgAHqrzp,
W94a1YzLCMXdJ3wZ, X73w8hANRMqvam4k, XwPJHVmAjQh3YTF2, YNbjCvB14zAWymtT, brPZjRMJNWgqDfGh, dLMvNRZPbQkrFh6J,
gNc19PX3bDfM7QFL, gNn8G94YtABfWT2K, hZwnNYgLKGMbXkFD, kLbc6g8YfXwxHtTC, kbtPNGaFfWDVvAqX, v1q7y4LTR3JGtDpc,
vNJaB1t9d2hTPxAC, zRybPNv3hAcwtf7J, zV6waX37tYjnrABR; fights are not listed.

Scripts and output, in `forever-joc-suppress-1125`: `extract_r2.py` and `units_r2.py` (the units), `joc_bins.py`
(`joc_bins_out.md`) and `crusader_totem.py` (`crusader_totem_out.md`).

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

## Count of the 2026-11-20 to 2026-11-27 reports

- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the streams of every rogue, warrior and paladin wearer
  of the weapons and equip procs the sweep's roll watch follows, and of every rogue among the sweep's wearer streams
  with Windfury Totem hits, too many to list
- Game build: 1.60.1.70959 to 1.60.1.71139 by report date (71037 from 2026-11-20 18:23 UTC, 71091 from 2026-11-24 15:21,
  71139 from 2026-11-25 19:14); the hotfix each player ran is not recorded
- Date: 2026-11-20 00:00 UTC to 2026-11-27, by the time of each event, split at hotfix push 114313 (2026-11-21 00:00
  UTC), the 2026-11-23 hotfix (2026-11-24 01:00 UTC) and hotfix push 114399 (2026-11-25 02:00 UTC); the hour that holds
  each of the two pushes is left out. The Windfury Totem count takes the reports by their start, from 2026-11-20 00:00
  UTC
- Pull request: #1266, #1273

Hotfix push 114281 takes the flags that suppressed weapon procs off Shank's row. The first Shank followed by a weapon
proc in these logs is at 2026-11-20 01:10 UTC.

A unit is a wearer's lone Shank, Execute, Judgement of Command or Mutilate on one target, a Mutilate's main-hand and
off-hand strike together, with killing blows left out. One roll gives the procs the lone white swings in the same period
predict: the same class's on the same item and hand, or, where the class has fewer than 30, the chance pooled over every
class but paladins and druids in a form. An equip proc takes its rate on the weapon's speed, or its chance per swing
where it has a flat chance. The rolls are the procs counted over the procs one roll gives, and the interval comes from
300 draws of whole wearers. A run logged in two reports counts once.

| Proc            | Period                                 | Wearers | Lone Shanks | Procs | One roll gives | Rolls [95%]       |
| --------------- | -------------------------------------- | ------: | ----------: | ----: | -------------: | ----------------- |
| Off hand        | 2026-11-20 to hotfix push 114313       |      84 |       3,840 |   195 |            220 | 0.88 [0.69, 1.02] |
| Off hand        | hotfix push 114313 to the 11-23 hotfix |     291 |      18,395 |   899 |            898 | 1.00 [0.93, 1.08] |
| Off hand        | the 11-23 hotfix to hotfix push 114399 |      79 |       4,539 |   219 |            212 | 1.03 [0.88, 1.27] |
| Off hand        | after hotfix push 114399               |      74 |       4,063 |   205 |            201 | 1.02 [0.89, 1.18] |
| Main hand       | 2026-11-20 to hotfix push 114313       |      72 |       2,640 |     7 |            163 | 0.04 [0.00, 0.13] |
| Main hand       | hotfix push 114313 to the 11-23 hotfix |     262 |      13,417 |     6 |            820 | 0.01 [0.00, 0.02] |
| Main hand       | the 11-23 hotfix to hotfix push 114399 |     102 |       5,290 |     3 |            359 | 0.01 [0.00, 0.03] |
| Main hand       | after hotfix push 114399               |     102 |       5,066 |     0 |            258 | 0.00 [0.00, 0.00] |
| Equip           | 2026-11-20 to hotfix push 114313       |      11 |         799 |    44 |             50 | 0.87 [0.64, 1.09] |
| Equip           | hotfix push 114313 to the 11-23 hotfix |      74 |       5,367 |   308 |            331 | 0.93 [0.81, 1.05] |
| Equip           | the 11-23 hotfix to hotfix push 114399 |      40 |       2,297 |   173 |            160 | 1.08 [0.91, 1.25] |
| Equip           | after hotfix push 114399               |      30 |       1,351 |    71 |             90 | 0.79 [0.62, 1.00] |
| Torturing Poker | 2026-11-20 to hotfix push 114313       |       6 |         420 |     0 |            377 | 0.00 [0.00, 0.00] |
| Torturing Poker | hotfix push 114313 to the 11-23 hotfix |      13 |         624 |     0 |            551 | 0.00 [0.00, 0.00] |
| Torturing Poker | the 11-23 hotfix to hotfix push 114399 |       3 |         205 |     0 |            164 | 0.00 [0.00, 0.00] |

A Shank rolls the off-hand weapon and equip procs once and the main-hand weapon never, in every period since hotfix
push 114281; the main hand reads 0.00 to 0.04 rolls. Torturing Poker in the off hand, whose proc hears only its own
hand's hits, never procs from a Shank; after hotfix push 114399 two wearers hold no lone Shank.

Execute and Mutilate from hotfix push 114313, and Judgement of Command from hotfix push 114399:

| Ability              | Proc      | Period                                 | Wearers | Lone units | Procs | One roll gives | Rolls [95%]       |
| -------------------- | --------- | -------------------------------------- | ------: | ---------: | ----: | -------------: | ----------------- |
| Execute              | Main hand | hotfix push 114313 to the 11-23 hotfix |      87 |      2,166 |   188 |            175 | 1.08 [0.93, 1.31] |
| Execute              | Main hand | the 11-23 hotfix to hotfix push 114399 |      33 |        777 |    72 |             66 | 1.09 [0.78, 1.34] |
| Execute              | Main hand | after hotfix push 114399               |      28 |        751 |    63 |             62 | 1.01 [0.83, 1.21] |
| Execute              | Equip     | hotfix push 114313 to the 11-23 hotfix |      44 |      2,020 |   195 |            185 | 1.06 [0.92, 1.22] |
| Execute              | Equip     | the 11-23 hotfix to hotfix push 114399 |      25 |        712 |    74 |             71 | 1.04 [0.76, 1.27] |
| Execute              | Equip     | after hotfix push 114399               |       9 |        464 |    54 |             46 | 1.19 [0.80, 1.88] |
| Judgement of Command | Main hand | after hotfix push 114399               |      45 |      2,852 |   208 |            228 | 0.91 [0.73, 1.12] |
| Judgement of Command | Equip     | after hotfix push 114399               |       5 |        282 |    43 |             39 | 1.11 [0.60, 1.75] |
| Mutilate             | Main hand | hotfix push 114313 to the 11-23 hotfix |      63 |      3,550 |   153 |            160 | 0.95 [0.78, 1.13] |
| Mutilate             | Main hand | the 11-23 hotfix to hotfix push 114399 |      18 |        936 |    51 |             48 | 1.06 [0.87, 1.36] |
| Mutilate             | Main hand | after hotfix push 114399               |      30 |      1,498 |    51 |             56 | 0.90 [0.59, 1.20] |
| Mutilate             | Off hand  | hotfix push 114313 to the 11-23 hotfix |      34 |      1,938 |    86 |            101 | 0.85 [0.67, 1.01] |
| Mutilate             | Off hand  | the 11-23 hotfix to hotfix push 114399 |      11 |        492 |    24 |             24 | 0.98 [0.42, 1.39] |
| Mutilate             | Off hand  | after hotfix push 114399               |      15 |        611 |    33 |             27 | 1.21 [0.86, 1.68] |

Execute rolls the main-hand weapon and equip procs once in every period since hotfix push 114313, and Judgement of
Command since hotfix push 114399. A Mutilate rolls the main-hand weapon once and the off-hand weapon once.

Equip procs read one roll off both before those pushes too, in the same count: Judgement of Command 0.95 [0.70, 1.12]
on 14 wearers from 2026-11-06 to 2026-11-20, 1.04 [0.72, 1.36] on 19 from hotfix push 114313 to the 11-23 hotfix and
1.11 [0.75, 1.35] on 8 from that hotfix to hotfix push 114399; Execute 1.00 [0.83, 1.16] on 26 from 2026-11-06 to
2026-11-20 and 1.25 [0.73, 1.88] on 8 from 2026-11-20 to hotfix push 114313.

A Shank rolls Windfury Totem once, at a Sinister Strike's rate. Here a unit is a rogue's lone Shank, Sinister Strike or
white swing, with no other hit or cast of the rogue within 100 ms outside the same millisecond, in the streams of rogues
with Windfury Totem hits, and it counts when a Windfury Totem hit on the same target follows 0 to 60 ms later. The
baseline moves each lone Shank's time 2 s later where no event lies within 100 ms. The log names every white swing
Melee, so the white swings of both hands count together. Reports are taken by their start, and a run logged in two
reports counts in each.

| Lone unit                        |   Units | With a Windfury Totem hit |  Share |
| -------------------------------- | ------: | ------------------------: | -----: |
| Shank                            |  53,871 |                     5,486 | 10.18% |
| Sinister Strike                  | 118,554 |                    11,916 | 10.05% |
| White swing, either hand         | 843,355 |                    28,079 |  3.33% |
| Baseline, 2 s after a lone Shank |  26,981 |                         2 |  0.01% |

Both shares sit below the row's chance because the sample includes stretches without the totem and avoided hits, for
Shank and Sinister Strike alike.

Scripts and output, in `forever-sweep-auto`: `roll_watch.py` (its main-hand Shank row) and the run
`runs/2026-11-27-shank` (`roll_watch.md`, `shank_rows.json`, `all_rows.json`); in `forever-venom-1127`:
`shank_wf_ctl.py` (`shank_wf_ctl_out.txt`).

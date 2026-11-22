# EV-0179: Barovian Family Sword's proc rate and its heal

- Claim: Barovian Family Sword (item 14541) casts Siphon Health (18652) at 1.9 procs per minute, priced on the weapon
  base speed (2.1) of each hit that rolls it, with paladins counted by EV-0155's and EV-0156's rolls. The drain goes on
  with every proc, with no hit roll, and each of its ticks heals the wearer through Siphon Health (18653) for the 49 the
  drain's row states, whatever the tick dealt; the heal crits on a roll of its own.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports and fights listed below
- Game build: Warcraft Logs reports 1.60.1.70938 to 1.60.1.71037 by report date (70938 from 2026-11-13 01:36 UTC, 70959
  from 2026-11-17 15:19 UTC, 71037 from 2026-11-20 18:23 UTC); the hotfix each player ran is not recorded. The last
  raidbots hotfix cache file of each of the three builds (the newest from 2026-11-22 03:29 UTC, 6,241 pushes) holds two
  pushes with a record of the item or of its two spells: 112758, the drain's `SpellName` and `Spell` rows, out by
  2026-10-30, and 16791757, the item's `Item`, `ItemSparse` and `ItemSearchName` rows, out by 2026-11-11. Both are in
  all three files, so no push moved these rows between the first run and the last
- Date: 2026-11-14 to 2026-11-22
- Pull request: #1174

## Method

Every player whose gear shows the sword or who casts Siphon Health in a report: 15 runs of 12 characters, 11 paladins
(13 runs) and 1 warrior (2 runs), 11 runs in the US and 4 in the EU. Two more reports hold a run a second time, one
whole and one only its last 48 minutes, and count once; two reports with one proc each are left out.

A proc is an application or a refresh of the drain by the wearer; both log, at the millisecond of the hit that cast it
(1,689 of the 1,816 in the 15 runs at the same millisecond, 116 one later, 11 further off or with no hit near). A unit
is the wearer's hits on the melee table at one millisecond on one target, and a proc goes to the unit at its millisecond
on its target. Only units strictly between the first and last proc of a stretch count: a stretch is a wearer's procs
with no gear snapshot of another main hand between them, and the longest run with no proc is 102 units. A unit with a
killing blow is no roll, as in EV-0153. Of the 1,816 procs, the 32 that open or close one of the 16 stretches, the 2 on
the 351 killing-blow units and the 10 with no melee-table hit at their millisecond are outside the 1,772 of the rate.

The rolls are EV-0155's and EV-0156's: a white swing, a seal's hit, Holy Strike, a Windfury Totem hit and a melee
ability roll once, Judgement of Command twice and, in these warrior runs of 2026-11-17, Execute twice. The two-roll hits
stay in the count: without the units that hold only a Judgement of Command or an Execute (528 rolls, 35 procs) the
chance is 1,737 of 25,838, the same 6.72%. Every run starts after the Judgement's own roll ended on 2026-11-13. A lone
white swing is a white swing with no other melee-table hit of the wearer within 100 ms; it needs no roll table and
checks the rate.

Each tick of the drain is paired with the heal the wearer logs as Siphon Health 18653 within 5 ms of it.

## Rates

| Count                          | Runs | Procs |  Rolls | Chance per roll | PPM [95%]         |
| ------------------------------ | ---: | ----: | -----: | --------------: | ----------------- |
| All rolls, both classes (Rate) |   15 | 1,772 | 26,366 |           6.72% | 1.92 [1.84, 2.01] |
| All rolls, paladins            |   13 | 1,594 | 23,534 |           6.77% | 1.94 [1.85, 2.03] |
| All rolls, the warrior         |    2 |   178 |  2,832 |           6.29% | 1.80 [1.56, 2.07] |
| Lone white swings, both        |   15 |   142 |  2,122 |           6.69% | 1.91 [1.63, 2.24] |
| Lone white swings, paladins    |   13 |    75 |  1,109 |           6.76% | 1.93 [1.55, 2.40] |
| Lone white swings, the warrior |    2 |    67 |  1,013 |           6.61% | 1.89 [1.50, 2.38] |

The intervals are Wilson intervals on the chance. Across the 15 runs the chance per roll gives a chi-square of 20.3 on
14 degrees of freedom. The paladins' lone white swings read the warrior's chance and their roll count agrees with
both, so they are in the rate (owner ruling, 2026-11-22). Every wearer swings at the sword's 2.1 s, so these logs
cannot tell procs per minute from a flat chance per hit.

By what a unit holds, the units of one hit:

| Class   | Unit                  | Units | Procs | Chance |
| ------- | --------------------- | ----: | ----: | -----: |
| Paladin | white swing           | 8,235 |   528 |  6.41% |
| Paladin | Seal of Righteousness | 5,652 |   396 |  7.01% |
| Paladin | Holy Strike           | 1,851 |   142 |  7.67% |
| Paladin | Seal of Fury          |   578 |    42 |  7.27% |
| Paladin | Windfury Totem hit    |   550 |    43 |  7.82% |
| Paladin | Judgement of Command  |   178 |    25 | 14.04% |
| Warrior | white swing           | 1,268 |    86 |  6.78% |
| Warrior | Whirlwind             |   512 |    30 |  5.86% |
| Warrior | Windfury Totem hit    |   151 |     8 |  5.30% |
| Warrior | Execute               |    86 |    10 | 11.63% |

No unit of one roll holds two procs.

## The drain and the heal

No application fails: beside the 1,816 applications and refreshes the wearers log no miss, resist or immune line of
the drain, on bosses as on other mobs, and every damage line of it is a tick.

A proc onto a target that carries the drain refreshes it: the 5 s start again and the next tick comes 1 s after the
refresh (333 of 358 first ticks after a refresh, to the nearest 100 ms). An application that runs its 5 s with no
refresh ticks 5 times (635 of 649).

Of 5,956 ticks, 5,917 have a heal beside them. The heal reads 49 on 4,959 and 73 or 74, a crit of 1.5 times, on 921. The
other 37 read a quarter or a half of those, under a debuff on healing taken, and 7 of them are crits. The heal does not
follow the tick: 1,875 heals sit beside a tick that reads something other than 49, and 1,869 of them read 49, 73 or 74.
A tick that deals nothing still heals: all 28 such ticks, 16 absorbed and 12 on an immune target, have their heal.

The paladins' heals crit on 881 of 5,298 and the warrior's on 47 of 619; the ticks crit on 192 of 5,336 and on 1 of 620.
The heal crits 4.6 times as often as the tick on the paladins and 47 times as often on the warrior, so the two do not
roll one chance. Which chance each rolls is not measured, and the sim rolls both on the wearer's spell crit.

The tick's amount before the target's resistance and absorbs is 49 on 4,376 of the 5,751 ticks that neither crit nor met
an immune target. On the others it moves with the mob, to 48 and to 50 through 54 on all but 26: one mob reads one
amount or two that differ by 1, and the bosses read 50 to 54. The logs state no mob level, so what moves it is not
measured, and the sim ticks the row's 49.

## Reports and fights

2ca9MDYtNRHWP74j 22; aYyJ9hdFwXKgDk16 1, 6; DfmZ1A8HwytVKJPx 2; HTKhZkgwam7PLC16 5-7; J38xbrFytgM9p1LT 4-5;
KMR8X9qNQDY4rcF7 1; NXMZV9xjvWpB7GtP 57, 59-65, 67, 69-77, 79-82, 84-94, 96; pG8TNJwybLPYWmxF 2; qbtZdThvBCXVcfAW 8-9;
RFGVCxDvb8mf2dPT 1; Vtb8Qkyr7A1zpCKw 57, 60; WfKv2TphtRrzNA3d 9; XytmY8TQFjpC7fKP 1, 5; yKRFZQN9CXDw1rYG 7, 10, 13-14,
16; zDxgMqR83J2dXkWF 1-2. The warrior's runs are aYyJ9hdFwXKgDk16 and RFGVCxDvb8mf2dPT.

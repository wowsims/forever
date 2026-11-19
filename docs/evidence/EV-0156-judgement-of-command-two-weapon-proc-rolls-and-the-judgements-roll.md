# EV-0156: Judgement of Command's two weapon proc rolls, and the roll the other Judgements lost

- Claim: a landed Judgement of Command rolls a chance-on-hit weapon effect twice and Fiery Weapon a second time, and
  rolls Windfury Totem and Vindication once; until a hotfix on 2026-11-13 a Judgement that lands a debuff, a white swing's refresh of
  that debuff and Judgement of Righteousness each rolled a weapon proc once, and since that hotfix they roll nothing.
- Sources: the 1,139 Warcraft Logs reports of forever.warcraftlogs.com zone 3000 from 2026-11-06 to 2026-11-19 in
  which a paladin lands Vindication 440667 and procs a chance-on-hit weapon or Fiery Weapon 13897: 1,172 paladins. The
  Method says how the units are cut, and the last section names two places.
- Game build: 1.60.1.70790 to 1.60.1.70959, each report dated against the build its region's version file named at
  the time (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11 02:19 UTC in the
  EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC). The hotfix of the Judgement rows on
  2026-11-13 is read off the logs and matches the game files exports wowsims/gamedata#284 and wowsims/gamedata#285;
  which push each player ran is not recorded.
- Date: 2026-11-06 to 2026-11-19
- Pull request: #1119

EV-0145 reads a white swing's Judgement refresh as a roll of its own. Its refresh units come from two reports of
2026-11-11 and 2026-11-12, and this file dates that rule: it holds to 2026-11-13. EV-0155 reads a Judgement debuff's
landing as one roll over 2026-11-05 to 2026-11-19; this file splits that reading at the hotfix. What EV-0155 took as a
landing's unit is not counted again here.

## Method

A unit is everything a paladin logs on one enemy with gaps of 250 ms or less: damage that is no tick, a debuff
landing, a cast other than Melee. The weapon's own proc events and Vindication's events are left out of the unit and
of its name. A unit holds a proc when a proc event lies from 30 ms before its first event to 150 ms after its last.
Units count between the paladin's first and last proc of the report, with the weapon in the main hand and no off-hand
weapon, outside stretches with no proc where 8 were due. A unit whose target dies within 400 ms is left out, since a
killing blow logs no proc that lands on the target (EV-0153).

The chance per roll is the paladin's own share of lone white swings with a proc, per weapon. The rolls of a kind of
unit are the number of rolls at that chance that gives the observed count of units with a proc. Intervals are 95%,
from a bootstrap over paladins. "To 2026-11-13" is before 00:00 UTC of that day, and "from 2026-11-14" is from 12:00
UTC of that day, except where a table says otherwise.

The tables hold different numbers of units of one kind, for three reasons. A table by period leaves out the hours
between its two periods. The table by period under "The other Judgements" and the two period reads of a Judgement of
Command price a paladin only from lone white swings on a target that does not carry the paladin's Vindication, 30 or
more of them, where the table per weapon and the table by region also take the swings on a target that carries it and
so price more paladins. The table of where a proc logs counts instant weapon effects alone.

## Judgement of Command

Units that hold a Judgement cast, Judgement of Command's damage and nothing else:

| Weapon                      |  Units | With a proc | With two | One roll gives | Rolls |
| --------------------------- | -----: | ----------: | -------: | -------------: | ----: |
| Wolfsbane                   | 16,763 |       3,517 |      189 |          1,919 |  1.95 |
| Bigger Rock on Bigger Stick |  6,966 |       1,468 |       53 |            772 |  2.03 |
| Verigan's Fist              |  6,092 |       1,239 |       70 |            623 |  2.12 |
| Stoneslayer                 |  2,718 |         634 |       39 |            348 |  1.96 |

Over every weapon the cast reads 1.97 rolls on 11,598 units to 2026-11-13 and 1.88 on 15,445 from 2026-11-14. A Holy
Strike reads 0.95 and 1.04 on 11,913 and 17,620 units and never holds two procs of an instant weapon effect.

Fiery Weapon, on 18 paladins who wear it: 141 of 1,271 Judgement of Command units hold two strikes, where two rolls
give 154 (0.92 [0.79, 1.07]), and none of 1,670 Holy Strike units do, where two rolls would give 200. The units with a
strike are 610, which is 0.85 [0.77, 0.91] of what two full rolls give; why they fall short is open. Crusader and the
weapon oils are not counted.

Windfury Totem, on the 169 paladins with 5 or more of its strikes from 2026-11-14, when the log shows the strike. A
strike is a swing and a seal can proc off it, so a unit here is a Holy Strike or a Judgement of Command with no white
swing, whatever seal strikes ride along:

| Unit                 |  Units | With a Windfury Totem strike |
| -------------------- | -----: | ---------------------------: |
| Holy Strike          | 13,737 |                1,462 (10.6%) |
| Judgement of Command | 12,198 |                1,319 (10.8%) |

Compared within each paladin, the judgement's share is 1.01 [0.94, 1.08] of Holy Strike's, which rolls the totem once
(EV-0008).

Vindication, on targets that do not carry the paladin's debuff, at a weapon speed of 3.4:

| Unit                          |  Units | Vindication goes on |
| ----------------------------- | -----: | ------------------: |
| Lone white swing              | 20,148 |       5,403 (26.8%) |
| Judgement of Command          | 13,827 |       3,709 (26.8%) |
| Holy Strike                   | 14,636 |       4,030 (27.5%) |
| White swing + Seal of Command |  7,262 |       3,480 (47.9%) |

The judgement rolls Vindication once. The white swing with a Seal of Command strike shows what two rolls read.

In the game files the rank's dummy (20968 at rank 5) is on the melee table and lacks Not a Proc, and the damage spell
it names (20966) is on the melee table with Not a Proc. The counts fit a weapon proc that hears both landings and an
aura proc that hears the damage alone. Which of the two lands each weapon roll is not measured, since the damage
always follows a landed dummy.

## The other Judgements

Rolls per unit:

| Unit                                 | To 2026-11-13 | Units | From 2026-11-14 | Units |
| ------------------------------------ | ------------: | ----: | --------------: | ----: |
| White swing + Judgement refresh      |          1.89 | 4,424 |            0.97 | 5,782 |
| Judgement, a debuff and no damage    |          0.91 | 1,775 |            0.00 | 2,770 |
| Judgement of Righteousness           |          0.93 |   700 |            0.00 | 2,283 |
| White swing + Seal of Command strike |          2.02 | 7,114 |            2.05 | 9,516 |
| White swing + Seal of Righteousness  |          1.86 | 1,253 |            1.99 | 3,448 |

The refresh is a landed white swing refreshing the paladin's own Judgement of the Crusader, Wisdom, Light or Justice
on that target, logged 20 to 45 ms after the swing. An instant weapon effect logs at its swing's millisecond, and one
the refresh rolled logs at the refresh:

| White swing + Judgement refresh | Units | Proc at 0 to 10 ms | At 11 to 19 ms | At 20 to 45 ms | Later |
| ------------------------------- | ----: | -----------------: | -------------: | -------------: | ----: |
| To 2026-11-13 12:00 UTC         | 5,512 |                571 |             33 |            484 |    23 |
| From 2026-11-14 00:00 UTC       | 5,446 |                542 |              2 |              0 |     1 |

A Judgement with no melee hit in its unit (a debuff alone, or Judgement of Righteousness), by the report's region:

| Region | To 2026-11-13, with a proc / units | One roll gives | From 2026-11-14 00:00 UTC, with a proc / units | One roll gives |
| ------ | ---------------------------------: | -------------: | ---------------------------------------------: | -------------: |
| EU     |                        185 / 1,847 |            200 |                                      0 / 4,332 |            550 |
| US     |                        192 / 1,997 |            194 |                                      0 / 5,100 |            594 |
| CN     |                                    |                |                                        1 / 943 |            126 |
| KR, TW |                                    |                |                                        0 / 244 |             31 |

The one proc after the change is Demonshear's Shadow Bolt landing 15 ms before the Judgement, the missile of an
earlier swing.

### The hotfix

wowsims/gamedata#284, exported on 2026-11-13 at 18:41 UTC, turns the DefenseType of every Judgement row from melee to
ranged: the debuffs, Judgement of Righteousness, Judgement of Fury, and Judgement of Command's dummies and damage
spells. wowsims/gamedata#285, exported at 19:42 UTC, turns Judgement of Command's ten rows back to melee. The EU
reports show both pushes:

- The last proc on a Judgement with no melee hit is at 18:00:30 UTC, and none follows on about 900 such Judgements from
  20:00 UTC to the end of that night.
- Judgement of Command's last proc is at 18:18:37 UTC. The next 98 casts, by 3 paladins, hold none, where two rolls a
  cast give about 19. The next proc is at 19:43:10 UTC, and the cast reads two rolls from then on.

The US reports hold no Judgement of Command between 17:35 and 21:00 UTC that day, so the window is read off the EU
reports alone.

## A paladin's count of rolls

Units with a proc over what a count of rolls gives, where the count takes a landed white swing and a landed hit of a
melee-table spell as one roll each:

| Period                    | Units with a proc | Judgement of Command once | Twice | Twice, and the other Judgements' landings and refreshes |
| ------------------------- | ----------------: | ------------------------: | ----: | ------------------------------------------------------: |
| To 2026-11-13 12:00 UTC   |            22,150 |                     1.249 | 1.092 |                                                   0.993 |
| From 2026-11-14 00:00 UTC |            38,255 |                     1.139 | 1.022 |                                                   0.919 |

The other Judgements' landings and refreshes are 11% to 12% of the counted rolls. About 2% of a paladin's procs are
not placed after the hotfix.

## Places

- `Tcyn8kFJ2WwDLjgN`, fight 1, 31:25.7 into the fight, 2026-11-08: a white swing, then the Judgement of the Crusader
  refresh and Wolfsbane's Blazewind Blast together 34 ms later.
- `GnaKtvX31m9Z7jp8`, fight 3, 7:17.0 into the fight, 2026-11-10: the same shape on another paladin.

# EV-0156: What a Judgement of Command and the other Judgements roll

- Claim: a landed Judgement of Command rolls a chance-on-hit weapon effect once (EV-0155), and Windfury Totem and
  Vindication once; a Judgement that lands a debuff, a white swing's refresh of that debuff and Judgement of
  Righteousness roll no weapon proc.
- Sources: the 1,139 Warcraft Logs reports of forever.warcraftlogs.com zone 3000 from 2026-11-06 to 2026-11-19 in
  which a paladin lands Vindication 440667 and procs a chance-on-hit weapon or Fiery Weapon 13897: 1,172 paladins. The
  Method says how the units are cut.
- Game build: 1.60.1.70790 to 1.60.1.70959, each report dated against the build its region's version file named at
  the time (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11 02:19 UTC in the
  EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC). The hotfix of the Judgement rows on
  2026-11-13 is in the game files exports wowsims/gamedata#284 and wowsims/gamedata#285; which push each player ran is
  not recorded.
- Date: 2026-11-06 to 2026-11-19
- Pull request: #1119

## Method

A unit is everything a paladin logs on one enemy with gaps of 250 ms or less: damage that is no tick, a debuff
landing, a cast other than Melee. The weapon's own proc events and Vindication's events are left out of the unit and
of its name. A unit holds a proc when a proc event lies from 30 ms before its first event to 150 ms after its last.
Units count between the paladin's first and last proc of the report, with the weapon in the main hand and no off-hand
weapon, outside stretches with no proc where 8 were due. A unit whose target dies within 400 ms is left out, since a
killing blow logs no proc that lands on the target (EV-0153).

The chance per roll is the paladin's own share of lone white swings with a proc, per weapon. The rolls of a kind of
unit are the number of rolls at that chance that gives the observed count of units with a proc. Intervals are 95%,
from a bootstrap over paladins. "From 2026-11-14" is from 12:00 UTC of that day, except where a table says otherwise.

The tables hold different numbers of units of one kind. The table of rolls per unit under "The other Judgements"
prices a paladin only from lone white swings on a target that does not carry the paladin's Vindication, 30 or more of
them, where the table by region also takes the swings on a target that carries it and so prices more paladins. The
table of where a proc logs counts instant weapon effects alone.

## Judgement of Command

A landed Judgement of Command rolls a chance-on-hit weapon effect once (EV-0155).

Windfury Totem, on the 169 paladins with 5 or more of its hits from 2026-11-14, when the log shows the hit. A
Windfury Totem hit is a swing and a seal can proc off it, so a unit here is a Holy Strike or a Judgement of Command with
no white swing, whatever seal strikes ride along:

| Unit                 |  Units | With a Windfury Totem hit |
| -------------------- | -----: | ------------------------: |
| Holy Strike          | 13,737 |             1,462 (10.6%) |
| Judgement of Command | 12,198 |             1,319 (10.8%) |

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

In the game files the rank's dummy (20968 at rank 5) is on the melee table, lacks Not a Proc and carries Suppress
Weapon Procs (EV-0155), and the damage spell it names (20966) is on the melee table with Not a Proc.

## The other Judgements

Rolls per unit, from 2026-11-14:

| Unit                                 | Rolls | Units |
| ------------------------------------ | ----: | ----: |
| White swing + Judgement refresh      |  0.97 | 5,782 |
| Judgement, a debuff and no damage    |  0.00 | 2,770 |
| Judgement of Righteousness           |  0.00 | 2,283 |
| White swing + Seal of Command strike |  2.05 | 9,516 |
| White swing + Seal of Righteousness  |  1.99 | 3,448 |

The refresh is a landed white swing refreshing the paladin's own Judgement of the Crusader, Wisdom, Light or Justice
on that target, logged 20 to 45 ms after the swing. An instant weapon effect logs at its swing's millisecond, and one
the refresh rolled would log at the refresh; none does:

| White swing + Judgement refresh | Units | Proc at 0 to 10 ms | At 11 to 19 ms | At 20 to 45 ms | Later |
| ------------------------------- | ----: | -----------------: | -------------: | -------------: | ----: |
| From 2026-11-14 00:00 UTC       | 5,446 |                542 |              2 |              0 |     1 |

A Judgement with no melee hit in its unit (a debuff alone, or Judgement of Righteousness), by the report's region, from
2026-11-14 00:00 UTC:

| Region | With a proc / units | One roll gives |
| ------ | ------------------: | -------------: |
| EU     |           0 / 4,332 |            550 |
| US     |           0 / 5,100 |            594 |
| CN     |             1 / 943 |            126 |
| KR, TW |             0 / 244 |             31 |

The one proc is Demonshear's Shadow Bolt landing 15 ms before the Judgement, the missile of an earlier swing.

wowsims/gamedata#284, exported on 2026-11-13 at 18:41 UTC, turns the DefenseType of every Judgement row from melee to
ranged: the debuffs, Judgement of Righteousness, Judgement of Fury, and Judgement of Command's dummies and damage
spells. wowsims/gamedata#285, exported at 19:42 UTC, turns Judgement of Command's ten rows back to melee.

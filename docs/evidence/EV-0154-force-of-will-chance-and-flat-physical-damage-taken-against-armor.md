# EV-0154: Force of Will's chance and cooldown, and flat Physical damage taken against armor

- Claim: Force of Will (item 11810, equip aura 15594) rolls on each melee hit taken that deals damage, a white swing or
  a melee ability, at 6.2% once its 30 s cooldown is over; a hit that deals no damage, a miss, a dodge, a parry, a
  spell, a ranged shot and a periodic tick roll nothing. A flat Physical damage taken amount lands before the target's
  armor: its buff 15595 takes 51 times the share of a swing that gets through armor off the swing, and Gift of Arthas
  11374, Tenderized 1292008 and Lacerations 10370 add their amount times that share to a swing on the enemy that carries
  them. Stoneskin Totem's party aura (10403 to 10405) takes its amount off the same way. A flat reduction takes at most
  half of a normal white swing, and in the one fight with two reductions the second adds nothing once the first has
  reached its cap: under Force of Will a swing under 102 before armor lands at half of what it would without the buff,
  and Gold-plated 1291908 and Stoneform Shoulders 1298231 read the same on the hit they spend themselves on. A crit
  doubles the flat amount. A periodic Physical tick on an enemy takes none of Gift of Arthas or Lacerations, whether
  the tick ignores armor, meets a tenth of it or meets all of it.
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000), the reports the Reports section lists; all fights.
- Game build: 1.60.1.70862, 1.60.1.70888, 1.60.1.70938 and 1.60.1.70959 for Force of Will, by event time against the
  time the CDN first listed each build; the Game builds section has the procs on each. The Tenderized and Gold-plated
  reports start on 1.60.1.70790, and the first Stoneskin Totem summon counted is on 2026-11-07 06:52 UTC, two hours
  after 1.60.1.70862 was listed. The tick reports run on 1.60.1.70862 to 1.60.1.71091 by report date. The hotfix each
  player ran is not recorded.
- Date: 2026-11-05 to 2026-11-19; the ticks 2026-11-10 to 2026-11-25
- Pull request: #1114; the single applications, Stoneskin Totem, the cap at half a hit and the crits in #1116; the ticks
  in #1229

## Method

Force of Will's wearers are the characters with an application of 15595: 277 characters, a character being a name
within a region, in 370 reports, with 19,160 procs. A wearer in one report is a warrior 171 times, a paladin 152
times, a druid 77 times, a mage twice and of a class the report does not give twice. A fight logged in two reports
counts once (105 dropped). A proc is an application of 15595 on the wearer; the buff lasts 10 s inside a 30 s cooldown
and never logs a refresh.

The buff logs after the hit that procced it: 16,080 procs have their nearest hit on the wearer 1 to 45 ms before them,
2,802 at their millisecond and 60 up to 5 ms after. A tank takes several hits inside that window, so the chance is
read on isolated hits: a hit with no other hit on the wearer within 60 ms, counted as a proc when the buff goes up
within 45 ms after it. A hit is exposed when the wearer's last proc is 30 s or more before it and a later proc follows
in the report, which keeps the trinket worn on both sides and leaves out the stretch after a wearer's last proc.
Counting that stretch up to the last snapshot that shows the trinket adds 476 landed swings to the 224,164 and no
proc. Past that snapshot a report does not say whether the trinket stayed on, so the rest is not counted.

A white swing is ability 1 from an NPC. Its hit types are the report's: a normal hit, a crit, a block, a crushing
blow and a glancing blow land, and a miss, a dodge and a parry do not. An ability's kind is its
`SpellCategories.DefenseType` in the game files (1 magic, 2 melee, 3 ranged). The game files rows are read on
1.60.1.70959 at hotfix push 114196.

## Game builds

The CDN first listed 1.60.1.70862 for product wowf on 2026-11-07 04:43 UTC, 70888 on 2026-11-10 15:10, 70938 on
2026-11-13 01:36 and 70959 on 2026-11-17 15:19 (wago.tools' build history). When each region's realms took a build is
not recorded, so the events after a listing ran on that build or the one before it. The reports come from the EU (180),
US (143), CN (44) and KR (3) regions. The first proc is on 2026-11-10 07:19 UTC and the last on 2026-11-19 08:12.

| Build window | Reports | Characters |  Procs | Procs on damaging white hits | Chance |
| ------------ | ------: | ---------: | -----: | ---------------------------: | -----: |
| 70862        |       2 |          1 |    163 |                  98 of 1,955 |  5.01% |
| 70888        |      11 |          7 |    604 |                 392 of 6,279 |  6.24% |
| 70938        |     205 |        152 | 10,899 |             6,592 of 105,360 |  6.26% |
| 70959        |     158 |        149 |  7,494 |              4,832 of 78,912 |  6.12% |

## Cooldown

Of the 18,756 gaps between a wearer's consecutive procs, 4 are under 30.0 s (6.8, 6.9, 18.8 and 29.3 s), 283 lie from
30.0 to 30.5 s and 1,939 from 30.5 to 35 s. The row states 30 s (`ICDMs` 30000).

## Which hits roll, and the chance

Isolated exposed hits on the wearer:

| Hit                                           |  Procs |    Hits | Chance |
| --------------------------------------------- | -----: | ------: | -----: |
| White swing that lands and deals damage       | 11,914 | 192,506 |  6.19% |
| White swing that lands and deals none         |      0 |  31,658 |     0% |
| White swing missed, dodged or parried         |      1 |  85,746 |  0.00% |
| Melee ability that lands and deals damage     |    845 |  13,780 |  6.13% |
| Melee ability that lands and deals none       |      0 |   2,154 |     0% |
| Non-Physical melee swing that deals damage    |    549 |   8,712 |  6.30% |
| Magic spell that lands or is resisted in part |      0 |  33,480 |     0% |
| Ranged shot that lands                        |      0 |   3,684 |     0% |
| Periodic tick                                 |      0 |  20,908 |     0% |

The one proc beside an avoided swing, in x3njCmpkVYf7WDgh, logs at the millisecond of a dodge, 177 ms after a white
swing that dealt damage. The 95% interval of the first row is 6.08 to 6.30%. A swing that lands and deals no damage is
one blocked or absorbed in full. With those counted the chance reads 5.31% (11,914 of 224,164), blocked swings read
4.99% against 5.50% for unblocked ones, and paladins read below warriors; on damaging swings a paladin's blocked swings
read 6.09% (2,041 of 33,496) and its unblocked ones 6.12% (2,850 of 46,531). The non-Physical melee swings are the
report's ability ids -4, -16, -32 and -64, a mob's swing in a school of its own; 248 of their 549 procs are on swings
resisted in part.

The chance is flat per hit and no rate per minute. On every landed white swing, the damaging and the others together,
it reads 5.0 to 5.8% in each 5 s bin of time since the cooldown ended, and by the attacker's own swing interval:

| Attacker's swing interval | Procs |   Hits | Chance |
| ------------------------- | ----: | -----: | -----: |
| 0.5 s                     |   211 |  5,394 |  3.91% |
| 1.0 s                     |   801 | 14,903 |  5.37% |
| 1.5 s                     | 1,134 | 24,501 |  4.63% |
| 2.0 s                     | 4,090 | 78,248 |  5.23% |
| 2.5 s                     | 1,872 | 31,727 |  5.90% |
| 3.0 s                     |   289 |  4,710 |  6.14% |
| 3.5 s                     |   161 |  3,028 |  5.32% |
| 4.0 s                     |   267 |  5,124 |  5.21% |
| 4.5 s                     |    60 |  1,106 |  5.42% |

The override on 15594 is 6%, the whole percent nearest 6.19%, since the field takes whole percents.

## What the chance moves with, cause open

On damaging white swings the chance differs by attacker more than chance allows: over the 221 attackers with 200 or
more such swings, 220 NPCs and the swings from no NPC as one group, the chi-square against one pooled chance is 497 on
220 degrees of freedom. The four lowest and the four highest of the 58 NPCs with 1,000 or more such swings:

| NPC (id)                          | Procs |  Hits | Chance |
| --------------------------------- | ----: | ----: | -----: |
| Scholomance Student (10475)       |    36 | 1,054 |  3.42% |
| Splintered Skeleton (10478)       |    39 | 1,051 |  3.71% |
| Bloodaxe Worg (9696)              |    45 | 1,146 |  3.93% |
| Spire Spiderling (10375)          |    79 | 2,003 |  3.94% |
| Molten War Golem (8908)           |    78 | 1,036 |  7.53% |
| Smolderthorn Shadow Priest (9240) |    81 | 1,067 |  7.59% |
| Firebrand Grunt (9259)            |   162 | 1,871 |  8.66% |
| Gordok Brute (11441)              |   104 | 1,100 |  9.45% |

The 162 NPCs with 200 to 999 swings run wider: Cinder (261974) reads 2.99% (6 of 201), General Drakkisath (10363)
reads 11.22% (56 of 499) and Darkmaster Gandling (1853) 11.28% (30 of 266).

By dungeon it reads 5.69% in Stratholme (1,083 of 19,032), 6.07% in Blackrock Spire (3,893 of 64,183), 6.10% in
Scholomance (1,621 of 26,559) and 6.70% in Blackrock Depths (1,995 of 29,794). By the swing's size it reads 5.93 to
6.35% under 400 damage and 7.17% on the 12,560 swings of 400 or more. By class a druid's damaging swings taken read
5.96%, a paladin's 6.11% and a warrior's 6.39%. A report carries no mob level and no wearer level, so whether a level
term sits behind these is not read here.

## The buff's 51 against armor

Unblocked white swings that are no crit and deal damage, one NPC on one wearer in one report, with 15595 up against
down, over the 1,147 such pairs with 20 or more swings on each side (58,639 swings with the buff up, 143,699 with it
down). The swing's `unmitigatedAmount` is the same either way (1.5 apart), and its `amount` is 20.2 lower with the buff
up. The share is the swing's amount over its unmitigated amount with the buff down:

| Share that gets through | Swings in the smaller side |  Cut | 51 x the bin's middle |
| ----------------------- | -------------------------: | ---: | --------------------: |
| 0.3 to 0.4              |                     30,090 | 16.9 |                  17.9 |
| 0.4 to 0.5              |                     25,622 | 23.5 |                  23.0 |
| 0.5 to 0.6              |                      2,120 | 29.2 |                  28.1 |

A line through the origin fits the pairs at 51.6 times the share. A cut taken after armor would read 51 at every
share. With the NPC held fixed the cut still rises with the wearer's share (a slope of 71 inside the NPCs, 1,043
pairs), so it follows the wearer's armor and not the mob. A free line fits at 80.8 times the share less 11.7, steeper
than 51: the small swings bend it, which the next section reads one application at a time.

## Single applications of the buff, and its cap at half a swing

A window is one application of 15595 with 3 or more normal white swings from one NPC on the wearer in the 20 s before
it, 3 or more under it and 3 or more in the 20 s after it, inside one fight and clear of the wearer's neighbouring
applications. Swings within 100 ms of the buff going up or coming off are left out, and so is a swing that deals no
damage. The swings with the buff down, on both sides together, must hold one amount over unmitigated amount to rounding
(a spread of at most 3 over the smallest unmitigated amount among them), which keeps the wearer's armor fixed across the
window; k is their summed amount over their summed unmitigated amount. A swing under the buff then lands at some share
of k x unmitigatedAmount: (U - 51) / U when the 51 comes off before armor, U being its unmitigated amount. 259 windows
in 82 reports (2026-11-11 to 2026-11-19) hold 1,122 swings under the buff and 2,741 with it down.

| Unmitigated amount | Swings | Wearers in a report | Amount / (k x U) | (U - 51) / U |  Cut | 51 x k |
| ------------------ | -----: | ------------------: | ---------------: | -----------: | ---: | -----: |
| Under 65           |      5 |                   2 |            0.504 |        0.038 |  9.8 |   20.6 |
| 65 to 74           |     62 |                   4 |            0.485 |        0.304 | 12.9 |   17.5 |
| 75 to 84           |     25 |                   6 |            0.479 |        0.332 | 14.6 |   18.7 |
| 85 to 101          |     32 |                  12 |            0.499 |        0.465 | 20.2 |   21.5 |
| 102 to 149         |    104 |                  26 |            0.602 |        0.598 | 21.9 |   22.1 |
| 150 and above      |    894 |                  77 |            0.843 |        0.846 | 22.3 |   21.9 |

A swing of 102 or more loses 51 x k: 22.3 on the 998 of them against 21.9, and a free line through their 230 windows
fits at 46.6 times k plus 2.3. A swing under 102, twice the buff's amount, lands at half of k x unmitigatedAmount
whatever its size, where the whole 51 would leave 0.04 to 0.47 of it. The buff takes at most half of a swing, and
that cap bends a free line: the windows with a k under 0.35 are small swings, and they read 13.0 against 17.4.

The half is a cap and no smaller amount in a lower dungeon. Read on every swing under the buff, with k from 5 or more
of the NPC's swings with the buff down within 60 s of it that hold one ratio, the 3,869 swings under 102 land at 0.490
of k x unmitigatedAmount (a median of 0.499), and the 6,684 of 150 or more lose 56.0 times k (a median of 50.9). By
fight name, for the names with 25 or more swings under 102:

| Fight                | Swings under 102 | Amount / (k x U) | (U - 51) / U | Swings of 150+ | Cut over k |
| -------------------- | ---------------: | ---------------: | -----------: | -------------: | ---------: |
| Krol'dok Stronghold  |            3,317 |            0.486 |        0.308 |            623 |       57.6 |
| Dire Maul - East     |              107 |            0.519 |        0.000 |             27 |       87.3 |
| Blackrock Spire      |               89 |            0.500 |        0.469 |          1,566 |       55.9 |
| Stratholme           |               62 |            0.505 |        0.327 |            759 |       53.0 |
| Blackmaw Hold        |               55 |            0.500 |        0.396 |            328 |       48.6 |
| Dire Maul            |               52 |            0.641 |        0.270 |            484 |       66.6 |
| Scholomance          |               37 |            0.509 |        0.194 |            861 |       57.7 |
| Blackrock Depths     |               28 |            0.463 |        0.180 |            941 |       53.5 |
| The Shaper's Terrace |               27 |            0.501 |        0.423 |            534 |       50.3 |
| Plagued Insect       |               27 |            0.493 |        0.000 |              0 |            |

This wider reading keeps the swings absorbed in part, which the windows show at 4 of 124 and which read the share low.
Dire Maul's 0.641 is not explained. Outside Krol'dok Stronghold the 552 swings under 102 land at 0.518, against 0.266
for the whole 51 off.

Two windows, by time into the fight. In tTRm37A4YwgM8nCG, fight 1, the buff is up from 1:55:05.6 to 1:55:15.6: a
Bloodaxe Veteran's 14 swings on a druid at a k of 0.429 are cut by 21.4 to 22.6, against 21.9. In CxTmQRX2ytGcwA9r,
fight 1, it is up from 1:13:55.4 to 1:14:05.4: a Redridge Prisoner's swings of 73 to 75 before armor land on a paladin
for 25 or 26 with the buff down and for 13 under it, where 51 off before armor would leave 8.

## Stoneskin Totem

The totem's party aura (`A_MOD_DAMAGE_TAKEN` on Physical: 16, 22 and 30 off for 10403, 10404 and 10405, ranks IV to VI)
logs no event, so its cover is read from the totem's summon (10406, 10407, 10408). A summon counts when the report
holds no Stoneskin Totem summon of any rank in the 125 s before it. For one NPC on one player, k is the summed amount
over the summed unmitigated amount of the NPC's normal white swings in the 20 s before the summon: 3 or more swings
that deal damage, none blocked or absorbed in part, holding one ratio to rounding, and a k of 0.6 at most. Each such
swing from 2 s to 20 s after the summon, up to the shaman's next earth totem or the totem's death, leaves
k x unmitigatedAmount - amount over: nothing where the player stands out of range, the rank's amount times k where it
comes off before armor, and the rank's amount where it comes off after. A place is one summon, one player and one NPC
with 4 or more swings after the summon, read by the median of those swings over the rank's amount times k; a place
logged in two reports counts once. 1,849 places in 760 reports (2026-11-07 to 2026-11-19) hold 10,779 swings after a
summon.

| Rank | Row's amount | Places | Swings | Median place, quartiles | Mean k | After armor | Places at 0.9 to 1.1 | At the full amount | At none |
| ---- | -----------: | -----: | -----: | ----------------------- | -----: | ----------: | -------------------: | -----------------: | ------: |
| IV   |           16 |  1,272 |  7,460 | 1.00 (0.95 to 1.06)     |  0.503 |        1.99 |                  809 |                 27 |      32 |
| V    |           22 |    431 |  2,426 | 1.01 (0.95 to 1.19)     |  0.457 |        2.19 |                  218 |                 17 |       6 |
| VI   |           30 |    146 |    893 | 1.14 (0.99 to 1.86)     |  0.441 |        2.27 |                   48 |                 12 |       3 |

"After armor" is what a place would read with the whole amount off, 1 over the mean k. "At the full amount" counts the
places whose median swing loses the rank's amount to within a tenth, and "At none" those within a tenth of the amount
times k of nothing. Rank VI runs wide, a quarter of its places above 1.86: what else moved those players' armor inside
the window is not read, and neither is Guardian Totems (16258), which raises the amount.

The three rank VI places with the most swings among the 7 with 5 or more swings before the summon, 6 or more after it
and those within 3 damage of each other:

| Report, fight       | Summon  | NPC on                         | k     | Swings after | Left over           | 30 x k |
| ------------------- | ------- | ------------------------------ | ----- | -----------: | ------------------- | -----: |
| 4Xfp9L3bKh2aPYAn, 1 | 1:20:32 | Bloodaxe Veteran, a warrior    | 0.471 |           12 | 13.9 to 16.1 (15.1) |   14.1 |
| 1GBhzTR8CZwcJfFV, 1 | 0:07:33 | Alzzin the Wildshaper, a druid | 0.510 |            7 | 14.1 to 15.7 (14.7) |   15.3 |
| BkH7c8WTzJVCQqyZ, 1 | 0:43:19 | Anvilrage Officer, a druid     | 0.468 |            7 | 13.5 to 14.3 (13.9) |   14.0 |

The cap at half a swing shows here too, on few swings. Among the places that read the totem (0.8 to 1.3) or hold a swing
under 2.5 times the rank's amount, the 15 swings under 1.5 times the amount land at 0.45 to 0.49 of k x
unmitigatedAmount, where the amount off in full would leave 0 and 0.24 on average, and the 90 from 1.5 to 2 times it at
0.52 against 0.43.

## Flat Physical damage taken debuffs on an enemy

One enemy's life is its events chained while 30 s apart or less. Inside a life and for one attacker, the normal white
swings with the debuff off give the share that gets through armor, k = sum(amount) / sum(unmitigatedAmount) over at
least 3 swings. A swing with the debuff on then leaves amount - k x unmitigatedAmount over: the debuff's amount if it
lands after armor, k times it if it lands before. Lives with an armor debuff event on the enemy (a name holding Armor,
Faerie Fire, Recklessness or Shatter) are left out, and so are swings within 100 ms of the debuff's own events. A fight
logged in two reports counts once.

| Debuff                    | Row's amount | Swings | Left over, 95% | Mean k | After armor | Before armor |
| ------------------------- | -----------: | -----: | -------------- | -----: | ----------: | -----------: |
| Gift of Arthas 11374      |            8 |    440 | 5.18 +- 0.06   |  0.648 |        8.00 |         5.18 |
| Tenderized 1292008        |            1 |  6,187 | 0.71 +- 0.02   |  0.738 |        1.00 |         0.74 |
| Lacerations 10370         |    1 a stack |    222 | 1.28 +- 0.14   |  0.641 |        1.99 |         1.27 |
| Gift of Arthas, Auto Shot |            8 |     21 | 4.28 +- 0.92   |  0.650 |        8.00 |         5.20 |

The left over follows k inside each debuff. Gift of Arthas leaves 5.11 over at a k of 0.6 to 0.7 (401 swings, 5.10
predicted before armor) and 5.96 at 0.7 and above (38 swings, 6.11). Tenderized leaves 0.38 at a k under 0.5 (56
swings, 0.31), 0.56 at 0.5 to 0.6 (60 swings, 0.54), 0.64 at 0.6 to 0.7 (374 swings, 0.65) and 0.72 above (5,697
swings, 0.75). Lacerations leaves 0.68, 1.27, 1.80, 2.62 and 3.19 over at 1 to 5 stacks, against 0.65, 1.29, 1.89, 2.58
and 2.97 before armor. The Auto Shot row is ability 75 from one hunter in one report.

Gift of Arthas' swings come from 18 attackers in 8 reports (2026-11-11 to 2026-11-17), Tenderized's from 473 attackers
in 239 reports (2026-11-05 to 2026-11-19) and Lacerations' from 19 attackers in 8 reports (2026-11-11 to 2026-11-18).
Spell hits against a flat spell damage taken amount, and a flat ranged damage taken amount, were not measured.

## Ticks under a flat Physical damage taken debuff

The reports are every indexed one where Gift of Arthas 11374 shows beside a Physical bleed or item tick (10 reports),
and the 15 with Tenderized or Lacerations, a tick and no Gift of Arthas. Per report, caster and tick spell, the ticks
with every one of the three debuffs off give r, the median of amount / unmitigatedAmount over 5 or more ticks. A tick
with the debuff on leaves amount - r x unmitigatedAmount over; a tick that took the debuff's amount leaves the amount,
or r times it before armor. Only normal ticks count (no crit, absorb or overkill), and none within 50 ms of an event of
any of the three debuffs on its target.

| Tick                               | Armor (EV-0194) | Gift of Arthas: on / off | Median left over | Lacerations: on / off | Median left over |
| ---------------------------------- | --------------- | -----------------------: | ---------------: | --------------------: | ---------------: |
| Rake 1824, 9904                    | none            |                223 / 539 |             0.00 |             115 / 166 |             0.00 |
| Rip 9752, 9896                     | none            |                 132 / 96 |             0.00 |               56 / 82 |             0.00 |
| Pounce Bleed 9007, 9824, 9826      | none            |                 48 / 120 |             0.00 |                       |                  |
| Garrote 11289, 11290               | none            |                101 / 373 |             0.00 |                       |                  |
| Rupture 11274                      | none            |                113 / 550 |             0.16 |                       |                  |
| Rend 11572 to 11574 (warrior)      | none            |                142 / 302 |             0.00 |             335 / 488 |             0.00 |
| Rend 17153 (Kang the Decapitator)  | a tenth         |                   2 / 17 |            -1.50 |                       |                  |
| Wound 17407 (The Blackrock Slicer) | all             |                280 / 400 |             0.00 |                       |                  |
| Bigger Rock to the Face 1306816    | all             |                217 / 586 |            -0.31 |           182 / 1,581 |             0.00 |
| Rock to the Face 1309967           | all             |                  81 / 92 |             0.17 |                       |                  |
| All                                |                 |            1,339 / 3,075 |       mean -0.90 |           688 / 2,317 |        mean 0.18 |

Gift of Arthas adds 8 to a swing at a k near 0.65, so a tick that took it would leave 5 to 8 over; Lacerations leaves
0.65 to 3.2 on a swing at 1 to 5 stacks. On the ticks that meet all of the target's armor, unmitigatedAmount and amount
read the same with the debuff on and off: The Blackrock Slicer's Wound reads 13 before armor and 8 after it either way.
The same reports' normal white swings are the control: on the four Gift of Arthas reports fetched for them
(1TFWH4Yy38KhXMwx, 26KV1wfdJpnaZFhA, 2gZx9BrGYTNwXRk7, Gkc8Y1zZDNb7wWVg), 3,089 swings with the debuff on leave 5.15
over against 8 x 0.641 = 5.13 before armor, and 4,353 with it off leave 0.00. Tenderized shows on 5 Garrote ticks in
one report, 0.00 over at the median; no Barman Shanker wearer is in these reports. The flat reductions (Force of Will,
Stoneskin Totem) on a tick a player takes were not measured.

## The cap on other flat reductions, on the amounts together and on a crit

None of the 124 window swings under 102 holds a block, a second flat reduction that logs on the wearer or a Stoneskin
Totem summon in the 125 s before it, and 4 are absorbed in part. The game files hold 24 `A_MOD_DAMAGE_TAKEN` rows with
a negative amount on Physical, and 6 of them appear in the zone 3000 reports from 2026-10-08 to 2026-11-19.

Gold-plated (1291908) takes 15 and Stoneform Shoulders (1298231) 60 off the next Physical hit and then go. The swing
read is the one normal white swing within 30 ms of the buff's removal, alone on the player in those 30 ms and with
nothing blocked or absorbed. k is the summed amount over the summed unmitigated amount of the same NPC's such swings
with the buff down within 60 s of it: 3 or more, holding one ratio to rounding. A swing logged in two reports counts
once. Gold-plated's 2,597 swings come from 323 players in 322 reports (2026-11-05 to 2026-11-19) and Stoneform
Shoulders' 100 from 25 players in 25 reports (2026-11-14 to 2026-11-19). U is the swing's unmitigated amount and F the
buff's:

| Buff                | U over F    | Swings | Players | Amount / (k x U) | (U - F) / U |
| ------------------- | ----------- | -----: | ------: | ---------------: | ----------: |
| Gold-plated         | Under 0.5   |     45 |      40 |            0.545 |           0 |
| Gold-plated         | 0.5 to 1    |     94 |      62 |            0.564 |           0 |
| Gold-plated         | 1 to 2      |     31 |      21 |            0.537 |       0.378 |
| Gold-plated         | 2 to 3      |    307 |     166 |            0.604 |       0.588 |
| Gold-plated         | 3 to 6      |  1,725 |     295 |            0.790 |       0.776 |
| Gold-plated         | 6 and above |    395 |     166 |            0.878 |       0.864 |
| Stoneform Shoulders | Under 2     |      4 |       4 |            0.513 |       0.229 |
| Stoneform Shoulders | 2 to 3      |      6 |       6 |            0.571 |       0.554 |
| Stoneform Shoulders | 3 to 6      |     11 |       7 |            0.803 |       0.795 |
| Stoneform Shoulders | 6 and above |     79 |      23 |            0.926 |       0.903 |

A swing under twice the buff's amount keeps about half of itself, the 139 under Gold-plated's 15 among them, where the
whole amount off would leave none of those. From twice the amount up both follow (U - F) / U, 0.01 to 0.02 above it;
that offset is not explained.

One fight reads two flat reductions together: mb8FgjhwVTLZ4q1d, fight 7, where Xorothian Imps swing for 55 to 99 before
armor at players under Black March Blessing (23120: 100 off, every school, an area buff no unit casts). The Blessing is
down for 12 s with no Stoneskin Totem alive, which gives k for three players (17 to 22 swings each). With it up a swing
loses 0.24 of itself before armor on the shaman (105 swings, a line at 0.250 x U - 0.8) and on the hunter (66 swings,
0.259 x U - 2.0), and 0.22 to 0.30 on the druid (373 swings): a share of the swing, and neither 100 nor a half. The
shaman's rank VI Stoneskin Totem stands for the first 108 s of those swings and changes nothing: the shaman's swings of
85 or more lose 0.242 of themselves with it (17 swings) and 0.241 without (36), the hunter's 0.238 (11) and 0.236 (18).
Once the Blessing has taken its share a second flat reduction adds nothing. The fight cannot tell one cap on the summed
amounts from a rule where only the largest reduction counts, the totem's aura on these players is not logged, and no
swing in the reports carries two flat reductions below the cap, so whether two amounts add there is not read. Why that
cap is a quarter under the Blessing is open; the imps' swings get through the druid's armor at 0.267 where most of the
report's other NPCs read 0.37 to 0.42.

A report's `unmitigatedAmount` holds no crit multiplier: on the Gift of Arthas enemies a player's white crit lands at
2.005 times k x unmitigatedAmount with the debuff off (111 crits). With it on a crit leaves 10.71 +- 0.23 over twice
k x unmitigatedAmount (102 crits, a mean k of 0.655), against 10.48 for the 8 doubled and 5.24 for the 8 added once:
the crit doubles the flat amount. Under Force of Will 5 NPC crits land beside a steady k. Four of them, 157 to 410
before armor, land at 2 x (U - 51) / U of k x unmitigatedAmount to within 0.03, and one of 499 lands at 0.72 against
1.80. All 5 are above twice the buff's amount, so whether the cap sits before the doubling is not read.

How often the cap is reached on the wearers: 10.9% of their 566,734 normal and critical white swings that deal damage
are under 102 before armor, 2.2% of the unmitigated damage. By fight that is 1.5% of the swings in Blackrock Spire, 1.9%
in Blackrock Depths, 5.5% in Scholomance, 7.5% in Stratholme and 68.2% in Krol'dok Stronghold.

## Reports

Force of Will, the 370 reports with a proc:

kW9V4CMKfmBbtrRH, 1b7PdYR4HhDXBGT6, BYyqMbfxpZ8nFhJg, LQG6jWCnAP1gwvdt, Z9y8b6WCft3R1nzh, jT7NDMWFfhKk1bdz,
NPHc3azxL8bmgdpF, NRK61HZcDdJkrwXG, 27jZGhfHgp1LR8MB, HWC2c1Bq3kMXz6Ar, YJPVatA7TqpZjb6B, ZMBDHxvV68jPdQk7,
PRrfKLbYXQdqzDjv, 1hRZ2v8PwKWt7DkX, c6tMPWv4fG2Vb8HQ, Mg2ay1xDjkYpBqRX, FMzkb2KrnWCYt3dZ, F3BTZGrW48Mp6n2L,
VT81WfMQFg4z6Cyj, Hwq97xLh6gv8ZPrY, dfAPgkrq3mJNcVvt, rPpHYfq3tNnwJRMd, byvnpga9tMq3BfX6, 7XK1DyrJ2TfAQ48p,
8jzXTaqr732V9DxJ, 1TZvYG2tfR9wVPNH, M2RHxFDymYWLnpPh, bg169rGdXTzDhLnj, 4nkW7JDxV12waMfr, JArH1Ky9zdhVk8RN,
CWhG6LyXk3PBHdMQ, RTfqrCdVvLGYzj1A, tA13MXZ8pLDjqcxG, z9XbfMRh21Fg7QjN, x3njCmpkVYf7WDgh, 3R69PBjxZvcfbnX8,
QNwkdxbMvWmrqjAY, 4DmjXGdbBQTKP6zr, 9zxHaWbYT3gqDQvF, 9h3cXRZ7LfM6gPwb, FVqQab31p7KXAJzt, aFMwqH32y4DtGLTN,
HxNMb7JVw4pGP6dn, CVfpgcNPrMqXh68b, pdkwZJNqWKCarQx9, zCyk3LFp4XZwaQBd, nmwM719fpABxvjdz, K3FqtCdPbLYxBzM8,
PtmYHjkDrqhdR9MA, D463FZBX1C9maJ7b, KG62VXF3a97CxtAg, tCnKQrPVv9Z6qdGX, J38xbrFytgM9p1LT, 6G3gb8DwtJFXm4f1,
xqnmW3BDpFb28cJ1, VYBjgJx8aZtRhmpy, zmDY4LVX7N8vh9aR, RtZaGw26kvVcAjK8, BDFj9VNrfXhxwQT3, GZzaJH1QVjMKd8qA,
zrALwGH82WM6Jcbx, NXMZV9xjvWpB7GtP, 8GNt2hC7v1rWRXyc, h3VmpYZ4Gt6xNJC7, QAgrxLyh1CVwcf6k, vZqVmLYQa6jwP3A9,
QczFgrTbG4CPpB3n, 7mnVXCgpyxkbtZGq, 2wMznDRvmkpbtHxA, Z3qdRnrYyXtkhLPM, hZGvF2bkQ37gK4wL, hHZTNk2zgLjDBMx7,
xTdW6NpRYKrM1Lyb, MR8g1PhjbDN3tpkv, nJqw7xa3BHptg2hX, AdfMXDLW29k6pGy1, D39jWq2CaRgbh7Jv, TA3pLxKFXQ9Ckm41,
T79Cvazwp2V8GjQX, 38bwApgqTf2rh9Pc, 1a3Qn7JpLk8AH2hV, FTQ8g6cvbyZ1wB3m, X6Hp4gDwv7zftxMC, PdNq3JxntCKgavkm,
jTRbQzV4qyMvxp73, ZRHAwVNzx7MkyTtK, 7ChZtckBQRfp3KAD, bF8hVwJNDT6HyRWt, gcxh69fvXWGtbRaV, 17ymZpJ6xFPf4Raj,
bcp2G1PHy3KWNM9D, Wv4QYA6C7antxKXr, pwmkX48DKCGFZfvn, VjrAGFkTvNy2W98B, FkBxY9zn4faM7jWL, tFm4dXCGDwznQ61a,
8zGhrb1FcV2xnBQW, XQkavwVbDrLKmhJ9, j6XFRzHxTC1kvVpy, dvkzFtPH3WR9C4Km, Mxd3G1mcXFv4RfN2, Daj7G8xFX3bA2CmW,
4QJzfYwDkXFrbxgB, h9nNL4WmZfjHzvB6, 3mwdt9jJF6Tyb1q2, GbvNW3tAP68ZpmcY, yHRFJ8kn9vQA4YLr, GRnMHmK69BXfWVTj,
2VpbrfHatzjZFkGP, FHqYCNcKntvDAjxg, HKY8mdFptfNkDgwb, ZDWyQGqFBvagT7kA, YCbcQHW378fnMvXA, HMdLhWY3pNbnXPtF,
MC7jkfvrpcbyTQ4x, q9dHhDJMk1btXpg3, RVkGfcrbZ13MWqQ2, vGJMahXtgTzFRmCQ, X8PH2CG1hMkcfxYJ, v23ytJHDz6QkbXnN,
VK83G7MdBhLRFAfa, aHKFrZvJgBypW24m, gmKvNqHFJLZXfdnP, JavZtXcdp9WgLQ8x, CDKVcP6ghZJtv8AW, DW9YkdvyrjtMBQnf,
HkfzanMQcVptAB19, B3MA6RjphHCXKQmW, CqM4Vnh831tFgcYw, YVWgj6ra27PvmNnx, 1JLq46yg7wYNhan9, MB1yTdQ9xmRXPaK8,
nJDNV6Md3bqpXZT9, 1z6PXMGdQBDvrCqx, A2YGVnmxNFPqk7Ka, ARFXzbqHDpBLJdhy, 2Cg8R4ZMz7HV3rfN, M7cz2xknrqXVAKjD,
Xy3tw1gBKbmqZjn9, tGBMqP9JQDj1vCr3, DH3dT6tYLKmMrNz8, ZG367WYvCXqhr2ky, DNc6QALhxmGbCH3Y, jrPWFdYfQNz3xG6C,
CxTmQRX2ytGcwA9r, fNxhtGQYCBnPbzcp, BkH7c8WTzJVCQqyZ, QD8ypCXWJcv9Y2B3, gpXqDYBNbxKc1y7W, Zxvmb47dcjWCQAF6,
aqZnYb67BPjLD1mg, dCFnA2kaP9MVYNhc, 71aKdZW3xFHLCT8P, P6GLNM3n27dtcYJD, 2hG49JzynDkwMVCY, hamTvL4wFpx6nYNj,
pDKQwPRxqjzNkbc1, naJFzXZPWbR9QAgt, tTRm37A4YwgM8nCG, qprynNbXtRZV4DHP, xzhQVR3dHa7vWZ1y, jBwLZh7p8QVNkX2T,
tHzN3kKbaPpVW8TY, j7xyJdLXHZNA6PQg, zNTQ4WnfhpcYk9MH, gBK1ZQAP9aDTMhJ7, a6GB8AKFtcMLTrhP, hWJLtqnXbzwG1TMD,
N9PGBMt7faY6W2J8, tkPZYqTRgCyWG3jM, wqaZjCkvGPMDt28V, 8xjZmDw2LKv9bckh, 8rAWbkX93chqKjFG, FJdrDz9w1Vxaf2cT,
8bWVvtJRXFQNTAjc, zvjKa8hknGcLy7ZF, 1TBKn6pMtbQHfDdw, 4CVrTcXtBf2yAYH9, R1g7q2JAnwzdrPvc, WhPpmnzB6rR3gD4Z,
9GJNhYFB1QnKqzty, 3H1fvFaqxRL79PY8, C2xQXzmFg6P1Vaq7, pMkPWzaC7wf9XZNt, 4y6Qnzkw7GHKcVDb, 1VpHn3kYq6B8JMWw,
ykz7Bdp9avqbAjRV, Wy6TdamAk4ZxJ8MC, P3QtFzTLVp1aWBbA, CfdHgJAvtMRhXNGQ, WMy9axzfg3jKLhYc, wDGP2BrqpYvgQxRc,
rDYW2wpQ4bBncAy1, 26AYbFNQzMZCmfkh, q38PjpNH9WmL6zCd, KvQYAx8bW9FpVXdB, 4hQkw2BmKJGd1XZ3, ZKphvLAPfVxNCX6z,
rqjzTHacGgmF27Xx, 4pynYMkBdVgRDvKH, NphMF846ajqQWAZg, ht9mXk7AM24JdGRH, 2WkFQGcDVYwpTMJ9, 1Y4NWLka9pAJbB6G,
jkcbfRxrPgvWqwyM, 1GWa2FTw7BDnhA9v, DBzk8qw4yFWtb9rd, 1bgXr3HRqG2cTnjQ, p6mVRBzvWMqnQZFc, JQwbjnVYyGcP4rg8,
ZhbCF1LvkQ9gBAMy, ztby6Z2W3nJmT1NA, pLYdFf8zbqwGrcZK, p4waHBZgjkA9mDLn, pFH4amZPxkBhyTtV, R78CFp1D24jAatwZ,
ybf3pCJXnTQD2V9G, g4KqJfjzTN8AFQCr, nbLKdYfXHDam79R6, DmjBrYd7M9VZwCHF, B1qcTGpnHCrQ64Xa, 1MA3qWLGa8v2hdXP,
DyY2txcNwvGk4qpJ, 4ChPnBrVK2fJLMaX, xdFjKnGbC2caDhPJ, tmd2HZRwKyMFPQqz, kz4FVdBRqpJwj8nx, 16GhDNRmVn8KWjBT,
KcLtDQRdjbrxyqXJ, GHK2gB78FMVCmPfW, fqNBazArpjygDXhZ, 8mJtz4F6gTfpw1Wq, ZJgFdtvGTYQVCqyb, jKQDqkwmx46ZLBFg,
RGDd1M9QCkjFpymY, 7KZzkbdXjBHGWQJa, zwDb3yF98agvj1kA, HNtn34KXCMyDrQJj, pzBvhjtCD9VRf2mP, yCphrBFqdjgb8Nm9,
Mmp1crXk2BwRb97D, v19wAVhZBPacHQFL, k7j48fCpbzDKXwLc, F4kZ8RhXvzx1CQDG, wj7k1zy3QLmBNY2t, B4QqHnGXwzpRfaYW,
jfHwWDKtgdFxYznh, LQ9zMDqpmknBgAX8, dyNbRtmWjgT8xfMB, aYyJ9hdFwXKgDk16, pQWjBPM2XZ98CA3n, Rr1GThF93wHadWcx,
VJfPj3WbCzvLmQnA, 1Wc4yTf2MQPxwXDa, R2GxVLZnrkh1m6Hp, 4Tb63aXJykLnKWgN, aLwy1zjnrYFhvpDq, QXWcwaGZvCKjJFM4,
rf4VFwdKH2CjNPJh, 4vcn3WF1RxXf8VZt, 2fFYW3BmAjDRQxCg, 26j47DAXZCgwdxKB, fCB4wQt3NWm6k8M9, AKQPn8Ha7wG1XJVv,
LXPyfNtDC4mhVQ7R, zpw3WxckYDjXBJRy, 9R8rhcAwz6V17fpF, nAHx4Qv6ZaJ1VNz7, xWDkmwXKrRcdM7Jv, RJGVrNbKj1vXyF46,
tq89RC2Tv4jy1WcA, TzQa7YAycKWhrtCb, Xfbc1Pa24xzTBwnN, JcmNfbzKkq4prQg3, WYTNrpf3VQZt7Xqw, LYw76XG1CFHQNdJz,
J2tdqRB9HykWhGgn, hDJKcgjxvWQfHt9M, 1Tf47h2MvyczZbVK, qcWFJdjk3Mx6YK8B, tpKhjLGfwCzb6dA2, FHVZr9P48Wp1QjyL,
KtCPg2xhR416AvXN, 4D9aqWFLJVvn7bYx, RaTGbvyP3KxH8YwN, LvDw6BjHCVGgzJyb, HD8mBq62ZgkK1yRT, Wn8wHhaxy31MLzjp,
MgDw2bXrKTmWaLzY, Zz3YcQk69P7N8rAn, Q2PbD38mAWa7TzCk, Xn7WZbwdcp6jFP8B, PtKXrwfj8zBHm2Gy, f1WqtFkXGn4pBzT6,
8dxYbPNzVmFXvWjC, qbrGgBZfzLHFW12k, 8GFRaPBnKWf7Qrpb, dzgfcTtLA1v7jDZ9, aT2GjdRqgB9HtNL1, JnQNajVrBKq9tR17,
Ycq2twGVavTpMn13, 3aXTp97wjCWNxMhd, QwYrJ4VPjyBkRbft, vF3fwb7aQ9BgGmWt, Zvr84V1Y3Gpg6n9H, a2m7tvgYdbw3jVTq,
9RqNnvBcJp4fdX2h, 9tjaL4MdpWx6rmhT, M71jtPDgcNR4fq9T, tg9xyavcrR2GqhLz, CA9rN6hjpPHFdMvB, GcbndBCyf1N9m7YT,
vyQrDZN7RX82ApnP, a6t8yQpMdJRwDvhj, qJX72LBYW1fAx6DP, XM9mHgbhYAzWTjGy, VJ7chPDkjLx4AN8m, rKaNzcYX94fkTCmJ,
qTQYLaD9yjWG4Nnd, tJNYAnRT6pgv9xdz, wBZCxaXWn7TAyFGV, zFJ71HAQKYbv623V, rTDcQk4gGxfmjN1q, Pp4dVW6y1m7bfATF,
h67XZzf2CwTr4VNy, qbtZdThvBCXVcfAW, MWTNLc6jZdC7AXRH, ZDkaQjxTVm2dMgpW, Zb9vrBGFf2a8zgPd, 6hf8JKg9dzxcWL3q,
x9Jty6qTazhGHLKB, xcGfXYvTD6bpg3qt, MtWFJv1bPL26KTdx, g3zGbn4VD2fPycNv, Y74hG9vfj8MbcyCQ, DCWdZLkm2cPJbfz9,
RJXQC67jbzATWdx1, 3jARkacpG1LJgCTN, k1taHbwL76Cnxv4m, fvKDkG9td3jzXT2L, 8YGwk2vfrBpZDx4V, TtnNvmqR7GwpZaPM,
KfcWYGqM8hykpN12, HDmZ4zdFK6xgPCT8, ZRnjHLXNgyMFqpmW, T92NqcgBXaZrKvCm, NAnmwq6MHRdFQ1t9, tw8ZP2krfqXvAB4Y,
CBGfNpnjv2ALw9Kz, Xqkcn3CgZY2ryG1w, 82F6RVYwhapt4zLP, GM97YvP1hdgx8akn, f6Q2vmDkJ7YhAy9H, 3Ra7Dtjhc9AXLCzJ,
rajyY8pFtKfd9QDn, yfN361HMZk4vp2hm, QFqKVZgc6vAJfbrW, Z6Xm92K3zgfrqM84, RHgby76wQXp2nrxj, VKfYJ82cMTktxQzL,
cav7VZrmPbNW3y29, rfvtaC13gNqRx4Az, 7Lk3abCgNpVZjM41, 4zHJh1aptbByGFnj, 6CbKfDBm78nhyLjJ, n7hQW3NYzJwmDaG4,
qKtgYkmZMVjnLBfr, qkctgyRMaFxXVnmz, dK9TAcBfjVmD2n1J, gWpPrhyZ42qkJ8CR

Gift of Arthas:

ZXJDkzHTvMCpFbV1, hDGL1wmpV4WjxbHP, Mz7B1TyjCDahQY9x, Gkc8Y1zZDNb7wWVg, yqFPH1MKv746dQhB, xg7f81w2qvXGL6Fm,
7qPHR6cLn3zGptvN, tcBRkNWrqp29PzJj

Lacerations:

AnPHcV64wg8DThzZ, q62VRcw1m8fyHN4T, t38h6AJQWzRCHwZy, Xa2MfbLKQCPYH6nr, nPLWN2VDTZMz7Qk6, nP4N7AqTvChFQ2GY,
D1dxzYbNMKcpZAa2, 6Y2mMtCzb7ZvQKWD

Ticks under Gift of Arthas:

1TFWH4Yy38KhXMwx, 26KV1wfdJpnaZFhA, 2Dv3WbJrMTCFPVgH, 2gZx9BrGYTNwXRk7, C8pJQGXhK6cxgtT2, Gkc8Y1zZDNb7wWVg,
WmAYZ9MVRPqncTBJ, dYwmDcAyqTBGW61g, hwBvpVCGWtH7gKXf, yqFPH1MKv746dQhB

Ticks under Lacerations:

6Y2mMtCzb7ZvQKWD, 7RaLcAJVWH2nZPrj, BXhNKykZL7MDat8c, D1dxzYbNMKcpZAa2, Kdp4ArYMTxCZWHcX, nPLWN2VDTZMz7Qk6,
vtbdYWr47pPFKQ3m

Ticks under Tenderized: xMRpbWkrzv2AZTN3

Tenderized:

7dHJyYKgf6WB9TpD, pxMq28ANjCrWvcgw, jZhb8VYzvndpKPaJ, zFAVYtjCwbvRdyZn, 2aPGztJXAh974Fbm, 67yMCKbB91fdQtHw,
G6MdmNQW8YqaCZf4, 2qcmHx7QGDz1w8r9, BywJca3vKQPAM7L9, 86t7zaDPNxpRGrcT, Zy3P1DpzhALvR2Gq, KGnL8PNF7Hptzgm3,
B9dxNyPrnGTvDh2t, xmbNFzkq7cPAXJB3, vZ3x7bWPjfwHCgXh, tfX7HL9PmDdwkGja, JncG9jpRwQHhK7tb, 62bNYTtAkfDhZwzc,
HncmhA9J68wYv3gP, T6bnQy7fZrGVMcJv, zr6QacY1HTKLMWBJ, pMHYBm1VPADtKvg2, WNa6pZmGhrMtf19Q, VRvFB8wYK6Qg2myr,
qzvkgcVfWrm9npBy, h9jfxHpG2cynzCD8, GJbzHkj7yWhZ286x, Rbt46n8XhJaWFHkQ, rT3NBYyK1PFc62ax, fhRQALy7dFKwbznN,
YWdVfcBnjRQ9G3gz, 4fTZ768GxtrPNqzC, b3KV8BNhZv7cgjFz, zBxk3a4g1hmM68WD, 4jDvrdmYytRHZKqg, y2YcmB4qXnKCg18Q,
JZPaGbYCX26N4tgm, 8ZzhKBxdkVqjpXtr, J8YZxmQ6GB14yMr9, cTDB8RavNHbCK1Ln, fywxP7rCh6tGnBWa, g3kLNqK82twBy1nC,
A3vDFHYbw1KdNWpq, LNDkVQWKAgHhw6qB, 7nDYbQBv2RcP9Xpw, WcVwazXJKH9gCvpY, zQRvXaZ4ftdMbDq8, A7trQ8ZRvKDz3pgF,
f297FtyjGHmRVqYw, Py6HN97KpFYZQTD1, kvDKAcPqCgr8jn7F, HNAz6LWFJkY9nR8b, Zb6PGx7kaJHAyT1p, L2qKn81pm7jYkT9t,
jHXpwZ3hCRTNGt1F, hmc4Q9jtrgybfx68, Q827z1cwdtMnC9Dm, yjwLb3N497AVkRZf, H9VbtcRZJdaKM4QD, jxpN7FbH3Vz69QyW,
9L4hGr1bwmqAMnWf, bFVtHL9wGnxDKk87, KaVy7MjgvAYkBqFr, CKMFxqyLv9Gz731g, 6JKtQY2rRh3Nnjxg, mcATkq7RYvrP623t,
4MFWJ1gva3p7CjZ6, bhHyGVrP7MkJvtf6, rQMWqbxzVcp2Gn9Y, jtgZTL92cpQV7Kam, VAZk1B4wn2M9bDPg, KJ3xfRDvHLdtYcPG,
YkCwMAadhF48yzQD, 9Mw8TfWtvDNm4jF2, YVPX8gDCpb7B9QF4, 6JPk2WaZ9mcdvxYq, bq3XHZ2YJzGxhRLQ, ycDwGZPHhgQkYFd2,
X2BHnWa94wz7gcfP, MvAcd78yhXC4HYRb, ckxfJYL4DbGRz6hN, 8vC1fQH4cdYZqwhn, MQgJ13WZ64HahkDn, Wb1FC9Km7z3wdaTv,
KWZqJv6cCwhXBAkQ, g9mx7bNwpJPGWLk2, CyVnkj2p1tLAFrH8, pVNAk71xnYgDwqRF, 7ZDxRNanGAMz43cC, AmLpq9xyPrQYNagC,
Vk2BXyHrQZ3tjTJz, BPgkwpnQVfmvr78c, HpFdqkTJYLGNW1KR, Tp3CbBzgPWjKh7x4, jkKx8tW69wg3VXh1, 8GL31HwqngaC4fzM,
3wK4c7n9avYdykVT, JF4x6wzCgGbpjvKA, jHxJyDKtz4vdgXZR, h4n18FVZftmMTKRG, kg2BK6YhJHwNPAQW, Md7V9kmZxBNgr8KQ,
6L43dhP8DbTYrWK1, PMBZjgXqpFQVrb6t, pnmZ8qDtdar9b1KH, p6MKRN74X21z9CHW, fpj6NVGQ94BkqMdm, aVXmGbxtLp3F8BPR,
B9C8TVwHJZk6G4F7, B2D9zYahGJLKbVHx, q4ZYmVXk817FcGfa, yN3Rd1gPvFXkhjBG, BMPxQwDt4NhpnLZ1, NmgTp6h9djVQvq2K,
1YKV2DhvtmbFfHLM, 76P9n4jkJtfpB1Gv, tMZQk8gya1NPhwzG, AxhZmXMvQTdfnDNF, c6YD2PNAJ3pWq1Fk, KMynaR6bF2kzhxZW,
ZH8xt94FYMLWJqbc, BTP4tm9cynVGvLXq, gY6ZhQxm27aHTk9C, dhfTFcVNnGHvamk2, B69h7FP4HTRbWmaV, 68QFLTtpACX2v9ym,
HkfprWKGjZx4LR2F, Jx4GFqzVNn823ybt, T1YnWkgc7VLGdPJD, MkgT6r2q7ZvPdVnj, 174v6wGxA3bBCTzj, xjvgTAhdzmJQNK1k,
aMwchZFXtBdfz6Pg, 63qChvr8Py1WjZM9, zh71R8wP2HaJnjLx, db2awx9Wtym6PYBk, L1VXxcMARG9KD2dh, j6d1mRwzMp4HVArW,
Df7nRMWzaJAbFrZd, Vbmnq8CgAFK1M2t6, yLDtWzQq8NaJRbnM, 4fgDAM8PL2wpG6XY, 27Fr3MJTRdhNBHDt, nf369aFPXxVmcYWy,
Td4GCmfzFJckjZrN, nYwjhN643XFRykKW, DWjBLyVrC12hAmYq, HN9rfJtaAZLGmVT8, j1ZJ2xbYV9d7qh3A, aTktVmBCM4g8vdYx,
QhfFH38GWPJ6y2Yc, kzN31TQBth4L7cvx, gkPw6VGdWYFxqpK4, 8y4nDX7NHvQV6qKZ, CN4FVyDKvYkQHaX6, 1HcD8PyWkFbawzCJ,
b74wA3g92NxtPvKH, mbM7QABp49cfwhaP, Z2RmbDLxXGHgNFJY, Mx8Az2XhNR9PjJQ4, FrH3b8XJZQjLDfG9, wVA1zYmrNKCpvQn8,
CFZ7bqQcWTXLA6fj, Anp2WTCjHa4YLBKd, 4LYCmjDGAbF9dHQ3, hdL9NZfJzB1yjkvx, yDY2brPG3jQH9Kv4, FXQynRhmqdpGJrcH,
nQd4V2KTCfMLrmRt, 2whN8nQCX9PH6VGz, zQydq3xRAZnkh9JN, z2qnbhTND4wctxAK, xv19R2gVDqrpNXzf, AQ7X1kcHrNzDwKhj,
kHVMLR2dtwrjxWD7, pMwG8ZDbJ9AFW7nT, 7gZq4n2M8YXGPfjh, B8AKvLZ6bQpaXktV, xDbJdz2KatXqfNpM, MLkQWdC6hNcm34gz,
Q1TkNaDV8BXf39dK, zFrnw4yAmMxLkZ61, yx9ta7bcdv8Yjw1H, NfK46gDLj8wrJz39, Bm4ZTvKwQWtaD7nd, FqVQN1ca6x7Z4tmk,
3z2aQRWk86GmtDLK, pzhdrVck7vGj8BCN, GLgDZY82zW9Mha16, m9A2MvGDfPFRCWj6, vRPQ237fWm1tFYDV, LdcVztBw6pFjNahM,
2PWL9DB1k47p8MRd, Mv73gTFaNhbGWLy9, rgXGC8LkTYVKMfAD, RbDfMjcTwdCBQKWJ, pJac1fgd8jQk9BbT, pzvJ72fcan6gkKNb,
yCFgNJHXdZqDPKMz, tJAFnvQby9TXc8jf, xPhaCzTjybrZ48gV, rmL73jwRtzQ6HdyP, 2xWwaBRLgQ9D46FM, qDKRc3AxpVkgCWwf,
tdMgqKCmfLPJ8bX7, 3BH2d16wtD9F78QY, r6fZjG2Hx314KJTq, ZTJqNPfRLb3c8Bav, vYbjafXydK2z8chx, DMZPzKB91HJAqLQy,
tH8kRGA3bph7V9vN, cYazq49fdRZQv1bH, FHz1nBG8rhgMqaxW, AXqW386NctdMmQK4, MaXHVZ2nWDJ8z6Cd, Zv6fc3GgWpkX9Bqa,
TFnvh6k2P8zL4DCQ, 3BYFMvdKPgLc4xDn, 8VBhYx6tdm7GMvHK, WjnNbmcqYtkpK3TL, wtzQKj8GP4cJ2W6H, 8RWqvN4tjFHXz1DY,
FWHahtQvqjL3wNdV, ahRLFQNzAYfwT7jH, 1xXd2kGF3BQfAzJV, rWpDA8fghB1n4Xqx, KWFdXvCfR6rVYLcj, qRQfj9HCFYp2MrDT,
6J4ANdTDcpPK18M9, 2kFjvxn3MzRLQWTD, zKmW1p8PgRQtY6jq, GDWLqtcJgbrz1P89, xJRQt2TN3AFmPr6C, 8TKryAhdPfpW7t4V,
adj7kHPW3ypMKrNL, AVFatXcx7yBzN3hZ, ycK8T6mJWr43PpNd, fKGFtND74VBpCXbR, tZADbm8zqr9jVgJK

Stoneskin Totem, the 760 reports with a counted place:

NkcR19TQmhVLH4X8, Nx8YWHnr9hkqdwTJ, Djdh7LRMJZrTVgF9, F1rwvAXMKnWHqZgG, CdxGnj8Y7qrFyPRw, pFARQ4YtJqCP1aBk,
yQxCj74ztA21bnvw, 4CmQf7RxZTWXGP6y, D6dFkhMbm7tw1LCV, y3PRBMXYt9NfZDK6, 6Vcr8QYKH7hwAyDg, bD7WtNgT46PHQBkp,
zMn1jgVfJYxdXkLp, pJRYTAHMthmqLVBb, rkwG2zKFTY6ZMRd9, Wv12xNmaRkFypMCL, KQwZBprnckJ8bm3N, jkwWNBH9G3zyL8Ag,
V6AZd24WzYQPcxqb, dATmngxQrHfKpqhk, dzrKbhYRBa2Wy4mq, AV7JMpN3xL9Ywtrz, 43aHMAFCTwZbjnqm, Dw7PCT1RcqkJnvzY,
4DBhRn8Zzfdm3vT6, AxwDMvThyCmZzjgQ, PZACvYfMTyQa2FDn, RG8zj432mg6ZxTPf, AR9JMy8WBZDPXFqQ, 32TLH9vVcxNwzFQn,
2mr8jH3G7tCWLB14, JmaXQTGL9CVBrkD4, hC348bx7WkApLZJ6, BPhLF8yY2mJKtj3r, JCAj374pdrnqmhaM, 8AmJqPgwR2XM1kpz,
yxhmCYJ6dTbjF9Xg, Qcrwht1Xf3pjmYNJ, PDj4Xg1MWHFfnyrY, acPDKzvVdfxgwhTL, 1DcCdn4kfwJQpH7P, nBXvM31hf7FLWwGg,
ZcmAVRWMHXN6Ga1L, 7Az4bMNxJp2nRvfY, cqrAK8pgHNtz1mR7, mrFNg9RJBMtvhZKc, xDj6rh4Z12CMbLd7, YxFymKjwbNhLrMXJ,
wNQXy68CckKd9tRf, 1GBhzTR8CZwcJfFV, 6kCDd8zJKx1Ta4yZ, b19HGfTxL7R8Wp3V, 7mL3K8JcyCnTvX2h, bqkXHJ6K3NV8WM1v,
LY7gpHKAGchdjaZ1, cTgGLZ2fnYb9rpXF, nt1xXQpYyvGDqm8N, ftjDWaN8vT1XywZF, HXap1w9JYLqDAz2n, XMtL6zgdykpHGAT9,
xLYrjawpvF8yfd7Z, HgphfaTGdF8j6b9x, AGmXnhdJ9CcH8kzb, WbGCLXt8YxcwH2AB, HCvdbR2faypALKBw, 3k9rjMtPc4HTKpzm,
1Yj2ac7FMWQyH436, b1Yw7ckBhVja9FpN, Gqrc9N7MYbCyV6PQ, W38GAMfDbQhPLmN4, tNv61dYPCRFDnMm2, DAqYGd2WxZHVX3yb,
DdvXZH9nhaNByGjR, JWPgHQwmVXd3v2xt, kaMGJPmQLqByntjw, pLg6WDkjbzJaKh79, cfTk3YGvLwMrn8jg, kj1cNYTqGprDWRXJ,
WfxC2VYZHp6zkwh3, FAw1mt7b9MnCpNxa, TYmCntqj4MKG68Dc, zHZ4TX7Fnj9NKGtx, 2xCA4Qg31KV6bDyG, L9GjQd8AfHyRMZpq,
PzDfCG8mR4L3vtX1, T3ZGKnD7PcqdxkYW, 18rtPFkmAYq47Gag, ADznCRyQvmB4LqrW, Brtq69yb7cFXT8av, M1r2cNFGKBdVwghT,
v64Kcwk9bfXhQHRL, bjC8vZfh1c9M3JTk, CKYDdkHp2cnZL3AB, Mdn6xPJfHYh2RKW9, fhPzxX9QkDRAGC83, MdH67VBcJQgFTA2j,
WBXpDkcmynvC1hZK, 4TCPNvYm8MJXZp3w, pQvBYR1Z8LMXkg49, VLxFaKfQdC9mTqwD, WXM1hdLA4KbaYJHj, XfRDZr1MdNqJm6Gk,
RKmpPJhjZkf1VwcC, Nxp8wvcm7YTjRD2a, NPm3bZFztkBhLp9V, awT31rkNjQ7gtZB2, FMcQw7KgbLjhy3CN, 8WF4rmqzDpQgdPML,
GThRYwvzx26gNqF7, YZ1hKgH9b8DFzJWf, yPh1bAn6BapDkYW7, fNxCDkFjbqwh1Ldt, qa3hCjJ1w9yzgZYc, AfWn6NgLxbJ9P4XR,
aB7NcvpT9XQyrDjd, Yq7Anf3Fc1k46yaN, BNtHF4P19prM8Gxd, 8vZWVpdfD9AQcRHP, tBLbzrxyJ2C8ma7D, jhMTxHAnzWkPvd47,
JQVb4DAg3G1BK8Nm, ptKhnx9LrXJdANqg, 68b2wQyB3MRrHNgx, RM8t3VJF2BqG194X, K9fRd1Dwn48QX36H, RKQta7wZBJbv4fCY,
WjrY1pxNv38Fa6XZ, HpWmwnhGNFCk2jYd, Xv2kmKwxAZ3nYJgj, fQBxrTyqt26zX9Gw, 9FTaGAmpDYV6QqdK, q1rnP6YwMWD72dfQ,
rZP8qwmMQvDf2gjX, 42APtmwhNfaKx8nX, 4mgGvb2ZkNc79Pdw, FPLqKbdva4pY2gJ7, mjqgn2BXQbKLHzw6, AG1dKyakc9CvnfxJ,
bvZ6KTY3pGyjdmCA, f4LTA9BGkpR7YVxQ, raQxDgvCRXzGc3kN, b4YAPhmQk17Czn3q, qzpjJZY3ta1yBcf7, a8RcNxjmfVdyXp2w,
VAwGT2HLXmZNbhfQ, 8qj6wKcWbY4TX7Lf, GTVtNYz1bBAkJMQw, zjcabD2A6TtWHXC8, 8WQtqLj6wgfGmYny, VwyW3ZP1H6RTpdmf,
kh6JqDXxy9PRAHBb, rQ1D8PR6TnFYHydh, K9Fwn7CH81vbzkfM, 9JfHCPQVzyWxT734, qzbd2fW16PgpRcMQ, tpTvcBfJH6Cdz8Fj,
1pxF73a4nXqrHR9K, nYa7vm6rP3j4HLJK, ATygp8Gnq3hjYPCK, 3aG6Pf4pzDtJL9dT, 9jzZXJVwp6PQ148v, Nk6X4DhZAMxbKFym,
8PwLADkCTZYVHq4J, YPmbL2Ggfn1648HR, 2WH41MTLznPKRmvw, KT8XZkCD6nFdAgcm, MpFc4tj8Yhvnqx76, kw8a9C23KrDRj6XP,
QzPkRNnWT71rG2mg, 9KvzgHJqNMrAfhGp, VfZY7kQLwTpPXgKa, 6cJAVCrZ3FLT9zph, d479W2JqB68cQPzx, FMzkb2KrnWCYt3dZ,
4N8MrczThbGA6YXw, AP87dMpLm4axKyQG, ATwqPbhkavQCx2fz, vMJD9yVb2k3F6mXg, Hxdm7TKZqDJQFXzy, aDzbqXNJgCVBAy8c,
jRxQ32n6WNtGb4ZP, 4nNkF36BXh1GQzVC, 4X6VQynHdkDBF1K3, Lkz7TfZDtj1nmYMy, dptMfQHnJDxGAab8, 8FkLHnc1ZBgJd43v,
P4Nv9AfZa8j1DdGr, yJn9hAkMrQPYzDvK, jMbQaJ8mXkqPDGrF, xCnXPbMp3q4jagG7, BCx3qWw6dTA4ygcN, q8vTLWa9Bgpwz7KP,
wTAhM4aG9yYxLfq8, 8G3K1TCaq4fLXhHz, t2CLVPKwb6GrQhZ1, 8xP6yC7rLVHtdZfb, Fz2kv4hVDQ6cdWpG, FxcpZfaP2LYvGH4N,
jxZBAyvc8MmLqWC4, NbcV1BMRdY2k7Ja9, rPGVHTDmaK1kZ2dC, vpkbzJjPZhWMfYnC, YFDa7yhx8HbANTCv, rtQwP64YL7GqR1Hf,
btHDxFPjAgCyzLc7, NXMZV9xjvWpB7GtP, JtrVj2cqL3znKNXh, WXvYrG7p1Nw6Zgbd, PMgzmNcF8QXZdwjW, RxLjmanQC9XJyBvZ,
LXNz7Ym8rvwqcHax, 61VcxvhfQF2DRAGd, 7cpTkjaFB3mz6VgM, DQG8gWd1YvBPf2rh, 1YZn9rFqkLz3xVDb, ctAnYvNLQK6F38PD,
4LrMDtG6T8AVRJzQ, r1tJ39Xp7M8byN64, Z1VhLWGHNvQgD9z7, Tby1rLwFvkgRtxpn, 2mTBkH6At4qfc1v7, dnXbDqMcL3pT1mr6,
4ymYb9TKzHaDprhR, VYBjgJx8aZtRhmpy, q2bavD6kJPFMjcNA, tJTjwYNRHzPrW6Q4, LAQdRxqNBwKk3J8z, mWzXqK2kR9npr68t,
Lxy1kqgVvFc24aPr, RQ9C72kFLVNPdpMW, gVvTtLAWfj6mRHrZ, xYL9vR7wpmaMty2z, 7F6CMz3c4pqWbQ9R, DMLvZrTNa1xV4AX6,
nGt1RLYbdxTmFKwh, xk9bWKnJ1Zc3rydm, 3g24PkCNcApvHnh7, ACJ4XDNLKgkZdtjc, zP9LaH1dvGTKCFwr, XhAKfRr4GZYyzkT1,
aAcjGPbyZXNTzR1K, hJHTtV3YMK28NAPv, J2XL6AyZqgpjx8QK, ZpgM4C1DbqVQ3WHX, ra3pYvDdLKwFGB14, pKn8wBZcQ62PyJH7,
RhjnDZTwdGqxbgr8, 18Nf9mQz4RW2aYc6, rAFM4QTbC8jHypZL, BMJxG63gcbQX8tWh, Hdnh1GJgD4TbQkXL, bnfTwVz6BZvPxd1G,
aW6mC48kV1LxY7fh, 7Xx9fnkNvba2RHP1, JArH1Ky9zdhVk8RN, nzTN2Mc96KDtyqHY, KJgwjb9zDrnWBRH8, 28AhM3VGnvFcJ4rD,
q6tyF4ngNfMWTYZB, zA1GprYM7fPaKFH2, jy7Qk8vCNABT6m3Y, QvdrG4Vw8LBXAfN2, YF1tVDkdpBQaXvZL, Wfn9Lg6TGJPcaxK2,
Cqkfvy2Lrdtc47zJ, tgXLJCypHzh8cZd6, JKPwHm7fnCp3L14c, jZ2wqpGyv4zWT9xA, MC2GmHchpfVkNqwD, mZfvChVJXxKpB28W,
zNWxq9r8mPkTGMaB, nfH1DVpR3mvgjhW4, gAm7d3xLQYC1RGFp, yGhgd9Zp6WQMaKkf, Z3qdRnrYyXtkhLPM, Xwp1ybvGjJHVadkZ,
3wdx94CkMQVGmayv, NaDp8bGmJZgKf63y, A8TBJgd39akVwMmz, CzB8WTAZn36bgapQ, JVjbr91Zd7xnv62N, ZvQPn4LthVbAYw6T,
rTZYqFdvcCPyxWAR, 8PyaKJ2qHX4R6jkt, tpX42G3kzFZNHJBd, ZfbY6RgWjx9hmpGz, J9QrvLzMfB36Hagq, dXxjkL84WzNfw7CV,
FVwAz46LYh3nH9N8, qZFxvTQXVBWMJaCf, 9cQz2ymYMgfL7TwK, yKv4LMV1XwNFQj7b, JG9PjfzvTYADhqx1, ZLjg1tQ47a26W8fd,
PxR4q1fAHQp8wzYG, 6xz4k73CtjgrZATF, CxbcdQWmzMrt1qnf, mbTrtRa8JDhNMCkK, KFLjdptc3gqCGByx, qjy3rnkxZbN2tDvW,
tpWza62JNFrTm9hH, Td6bLPxyCFMgNKHX, KZzq7NTcBXj4Ln9r, D4rYVqW7ByaJLdmX, ntaAc893bkmBd1Cg, QFGbjqMxd1zfWAPV,
MT3fCDnYLRKXBJNA, FKmHWv9g17kphxQq, dYzgZkCNJFQAXtBP, dDKQgLW3MFPatynJ, xbJBZj14NWCXrhAm, 4RcnAhpLZ96KfNDP,
Ymk1av3chXyn7BJb, q1ytzBDkYgd6jZMH, PXdvLTcJwhzy78HM, CbG9THJkRBjDmPX4, d8tKg4mbQvkZhJnH, dzBLDHJTfc13yAWQ,
LjA2HFg3PTpZ84ND, hbnaypTcAWmqDMt6, gn7Wva1FKAm8HhGZ, zpDLH83PKwqxVfv7, 2VAWrb67G8KMzgj9, pnfQJ1tMhkwcRyxD,
tmVg9WB2hv3DT8AL, 7dWytj1K24G6ckJ9, nPLWN2VDTZMz7Qk6, k28NjXqp96v1HYAc, VRM3KtXTh2W7kzJx, a2ZkBR4b6HcFf7dW,
BLXHnFRfrhPGYa76, tvkf26qxyBNmJGPZ, Z4w7HKFPvcrTWd3M, PCW1JNXVx2B8nDyh, vatW8VJnyX3GmMKC, vfQr1GhC8HjdNZ7L,
grjxzfTdLQKbnBwZ, XznCM8awybq7xVr3, KLqzx6wJBdWHPZ4m, QTrc1HJj2d69hNMw, 6VJxcqnYa4wdBWQv, 1kdV2TvAWpBL7tqr,
fNwPCVrkyRZM6qFd, HgDdVqRj1YnBPw39, mb8FgjhwVTLZ4q1d, 4bZrDMz8cQqfp9P6, AdGxYrRMvywzb9pj, cJ3GnK21bC7tzj69,
FMaAjHcwCQLkpV32, YAjn4Ltv6KQdh2kp, FarzJZ1mH2ydBWxh, A3LXQbGck8Tpmgnx, yQapY1wXhDmMcZV2, wza3mdhrbgvLJjZf,
aKwXBJkG2bW6tVNY, LDj642vGyaF9wtZq, BCJvdpWVx7tK4Mcm, BCqWn73YfAmZPVN9, AVRZPtK3gJCrzwMj, xVLgCnjDvAkwJmhW,
KGQHdbFCZpt1YX23, rwxQgM38abAcmRVv, xQ6W8rHmCn7AcTwz, DNfvpnLRcPJyMKQA, YHXAcJxWG89pMFgf, 2AzZR61XcPdKgQrB,
KWmCxdQb18knPALV, 8mXaYLBcvgk6FnT9, BwAGFRkxhKpjWqnY, wnzh48QdvRMmYkq1, DpVmH4WRxL72nGwd, VJP9mHdj3YTWBkQa,
GhCgFan8yrxRATfJ, 2mTXwF9n1HW3LZpC, G4m9fBZrjt8gCdpW, dPmTc8p9vKMgtJkw, 8YHfcN1BTDCVvGgM, GZpcK3DfRvTX9dam,
rJb3LkQVwa2CGNDW, 8T9bVwnFLKxtYgRk, wM1d6qj9bGJn8Nhz, QADVqHgcjLyvBWP6, XmpdVHMF9W3aY8BZ, PWzVQpmK7Y28ZjqH,
bBKYFRrH1TnPVgcL, qKAn94haMcgjCwv1, VJZ4Q2KLGNqTx3PY, hx8gnMQ2waXVpLmB, gVZkC8ym74aRKHBj, 2gXTmYzt3dJ4PMwc,
MzXy2mnjLYqtcRa7, AtP7R1aN3vM4xjQq, mt43jMk97ZG2vyVW, 6tgdWjGnDL3N8mKq, 26CZDjqdH93v4bfA, zBr1JQc8KxVW9kRy,
Kr9cw3zjkApXbRyg, dZLYFXpjD4z231mk, dWHFMYj1zfRJAgLb, Jf9LwyAvhnp2KrVb, Vkxt7qH6fmDc9RCQ, Bg2HDytFChpKNP1V,
YMdVyaNtRLxkbPh8, 8frHF64GzAhBxDyb, mVNW7wGbzhMZ2vfd, gHL7ZnNBF4kyzd1f, tApFjznhLvCcg39T, Jk2Mjwf6K1Lpnyh4,
Tfp1v6KhDwJntrRd, vaBd6KrZx8HyGLRJ, 6VDthbQzAJxqpn7d, QnDdj3rkpTfBcmGz, 2aG3Ld8KVARCzrh1, TAyP1xXHrQRYFWLc,
FQnMCW8kRqVv4mT2, 12A94BJt3KCFZrY8, Qb1yWxa8vHCGL7Yr, 24fAZdgMD6K9qx8k, N2m8nPTGJQrj1DhH, YNLAyVFpwtcTWdrk,
2mTdHJaVjC4FYbhX, k7gmfZ69awYr3Dct, MfvNa4CQbRdJVAGW, BhyQMbdx1jcNpVP2, R7NLgmcVzbaZpktd, ByRxHQmnzfhbvtcV,
8r2XyARV6kcZaTFf, h9FNKMfq24Lp6JZv, VkJBAaKPFGwvq8CM, xTkKNdCZMbnXpmDh, 27qdfNjnPDKJT6W3, 97FVRX1WLHqDmjAQ,
h7aMpHbvcydmAPxN, hNLGm91wvzHKq43f, DLR1JAvjKnk7cVba, gqQTZC17pdHhBmAP, v8Zp2QF7An4d9xMj, q3KVJz9bZ6HmBtTg,
kAKR8fMtYHmp2Wg1, yD7rpdFXzxJgNHv1, TCztb2yhFq9mRZ4M, MwVH1KQmak73NPd8, hg8LCKr4cGp7nYkb, NcRqzah1TQjMr3nF,
w82TFzBGQyVHrNqc, 2XZNpQAPCcxYhtkH, AY4Kj1cmvGNq2d7z, wCvhA1WtFnVLB3mX, GMBKTFzhbpgDy3rN, PGXN8fK9FphR1n4v,
vTzn9cX2Nr3HCGkK, 1p8LXjBWrPtVazMQ, xKzXFQwq4LpBvD6h, TP91AFYK68WDgNVh, AdnrVKZNqHc2D3Ya, tKTFGHXcPBy1zbVg,
b8A3MkKZzj6N2PYG, wqtGCvbVByjRXmzp, gPxMKcJ7jBFtnGk9, 2f8Jc1ATWKqV3nFG, qGVkTWva8bcm7gr2, 2ypxfTd6vaMPFZj8,
2Whj4TMzbndBN8mR, zjxyZAgPDBMGndRK, JQaWFbjRfK8Mz46G, JPft9V7kQpaYnyMC, 8rkpCDc9hzKYVGnH, acw6RHFChKXdk1QP,
wLBjrZpV7v3CMFzt, XRJrztAfDK16HYFc, v9yDH6cFXwgAqbzZ, hc9VBMm83TdrRfWK, RkD1F2YMK36hGXBN, 8Ax6VHb3CMypWztX,
MPgTz8CHZKDjwcBq, CDKVcP6ghZJtv8AW, mzZNwpb3TjMFXGgH, dn62GTvrBYFQbWfV, CHxJrQMgPfynDGBp, ycjXJnNkC8pGWH1Z,
BhzZbxyKdrf4tGRM, xRK37aCvtVfj26kL, 6ZpTdQLxYNMF4fv8, qNWpaHdZLBRr6cxY, 4KWV2rZDJMqNLRmT, HqwQhrTj28a19Dfx,
8kjdAQ7mCrPBN9K2, jTYbp6KV9fnBD2mA, ynVw7ZgMx4tjF3fQ, PFGvpDf9WZgcY6Rk, HnGBKJvafjc6tQdp, BD8YvRFrymx4h61M,
tLAcdRaN8x1hrjKb, Pn7j3yhC1vFDKcNM, 7c21XafhWTAdqNGY, tpNByDzWPvrVwYx6, rj4pA3Cw1aNvMFfy, Aqjav1gfrDdXZwYW,
BkH7c8WTzJVCQqyZ, dfD1PZmkLayrtXB8, cQy3BzYNrFbZL8Pw, YFtvD87kznr1CThg, DBQHpyFz7chZRM6j, qcrKJGPvkfhAtdjQ,
2WH4jLt3FADy8fnh, 6wkfxZYRz4gHrGtA, LPQd7pfbjDkYZyRz, pqfK2AgHZ7DnhvPF, HtXh2PYD1VvwK84J, 4Xfp9L3bKh2aPYAn,
1tkBDhXngK4J23xL, zNTQ4WnfhpcYk9MH, Jcqz8grLfdPF9NDx, bvTaYKXPFQJqnB2V, mJZWhHDAt19gaKdw, FPTH8kbn6zvRgaKw,
GJhqzbPxTjRBCrVN, Cz7TcZ9D4x8tQMvy, DJCRPcry8mjpb1nH, rzvB21j7yD9hpaVF, 1h89JnyWjpZxfXdk, F67yxgT4hK8ZnCvp,
WY14kbF3Zqa7Gmnh, bpCg1YmBtKJvA7rT, Jm98RzHKFx3pdvWt, kz9YWBCb4vr6x7jc, hLxABfWVavkM8HGT, QhYmjJ4bHZXcgfk1,
6fvmrHP9dzy4Cp7b, tTRm37A4YwgM8nCG, 9v23hzNwDpWM67Xy, prCfHbFPmDAgqd3V, crtFfPbdavN6QHnX, 6nazf3JNH7DXtbBh,
bM67z9YCKwpJkQdj, JRHWTDakx4mdNtfg, axbF8PtgpH47v9wf, v8thQxgGN7ZFLXCW, wmCW7JvhcP6taxjT, j7xyJdLXHZNA6PQg,
zwT6Yyc9FN7WnhZx, T9aDyGtM6ZKA1Xcd, aWJqbmTz6fcB8HVx, d8T7NmC1tvFbZBVy, NW7A8vwpDybJkKX6, NwHAYtWgyXf7DRCp,
Wv46zL3F1Qc8Jayp, xGbBAMyFKq31648f, 9DVZT4NFcvaxrP2X, MwRP9LF4QNZ63dhY, 9MmXqYP3xz4vgZWH, KAfmrqFHXh7TzdL1,
AhxGmJv3nYDzTW9q, CFRzVcmKkYLpbdnh, mBRL3hY1zjaKqkX4, aBq2rjyJHQY3WvCL, L6h2Q3kMftCpjgVn, qfjNhkrtgabVCJpx,
DyC3KYB7wAZHz4f6, kCfXyg8GB3VADqpY, aL2ZJnf9pCgv6T1V, c8NL6YVXRvhgfkFB, faz6MxnBmhAtbdrZ, fDtYaBcQMK2dgn7N,
4zYkPVvpb9KXLxA7, PhB2bV6gv9mp4n1a, NkXjqd43AFVYpzGK, jad9HMkGz8cTCYwW, VyQ3pt7Hga1XbMjL, Lvh7GRfHXNAYnpV3,
2BHVPMfv94cKNqY3, 1wCQrckqDhYJvb49, nCDZQNWKX3twHhyY, mdbQ1GfTNZka4Jny, mnPXjHcWNFKBhDCq, TBA9wg8R1z7XWFMJ,
cCLTqrRHM4Pdk17g, RD6bnVNTgPhGYCvy, nRv1VCkBbY4APFMh, carnwgjTZ8V7bBJ2, 1mHkTJgX8G4NDMPx, L3yHtpv6hcTYFQfW,
6zmV2XcbfAT41Myn, ypQ9d13tMFJPNfvw, wPy3CjTdg49vKxmq, TNW9rHBt3MD1X8Ag, gDLz1ZAwkX2YPNK6, LwDfTRCWc691Maxd,
FRcWVytAbz3rBkK9, mbgjLf9CYp47RVrc, McP6DGYbrFh7zkjT, a9mp46BWQbwYvNLt, madL1Art92nRQFTC, QCRFhVqmHWMK4Gv7,
f7JajNKbQYLRZHpW, F4HWfxYQ7atgkJ81, H4C2q7tFvm96DJQc, 39NYyDpaM6LFX7nv, 1RVnhHGv2bwDW9Yr, 8TXMxn1VfPvdZtmq,
36h91AnYPjTkbxcM, jFvXRqgmH4QwWBNP, P3JYyjrc1FWgfN6q, jNz3BGqrmJfD9YbW, WJbgywhZnxzGNv7k, YAdaWntLCNyRVgDq,
cCBghykdHNAY8QFG, TnLAmyRdXvz1cBhC, pLtJjx8DCPbHRacv, wdYZbrDagm7cjpK3, Hk1mCnrfZ98hJq3w, AMRnwhKV6JvFQYpr,
a187rMQYxPBNAtvD, PqBvJ94Nhb7ax1Vz, vkG2m6pNJ8PWrQh1, KcLtDQRdjbrxyqXJ, XTfyZvKL9V1wCApr, 1J6j9t2qDYdKMwLp,
tHvCxwP9pDJByTzk, tk86znZXxMCdwBKb, 9bd6VQvJR2ZqGW3k, F4kZ8RhXvzx1CQDG, khtFfa9C6BnZGjrK, q46r2DwHaXmYdBAx,
RNtGgj9b8nm16ZHa, Kd6zWnZ9PFwbjMT1, YwNk3frmpRa4PbZB, 7mXHDLv4Yg9xKwp3, JgMKXYkv816Z7ABL, QNkVRHBbaXJYn6wC,
pzBvhjtCD9VRf2mP, 9HNMajzfDYLvxdt2, DC7kHQVWpayr1KTb, 6WnymAQYb7TJtdHr, rNqXQT1CwWgz7Lbd, pdNRGK6j9rkbyaV4,
VPgxZKWRAFDntLrQ, LKkr1pxAtT7CvcZN, Ax6CMHq9gNcVfnzp, wvbng2zVTYRrc9Zd, pQWjBPM2XZ98CA3n, GCTbWRf9mrHqxAX6,
8QHg7CqdNkJa4xGc, hAqC23VKb6JQkzRt, Azq6hMZn2RyWwagv, LcN7nmgBMF4CTHp9, LQDB9akd6rhc7bTz, RnaGBjV4xNkJTfZr,
Jk9bRtfDCmrhwAv7, 3XhFaCnrmKABQZHw, JZDFdhkvn3NP6xLm, fjZNTcyGxKQL4BhA, FAL4Hv1D7nWf8VNh, ZCRLWJqmb7v3gjaN,
bahZczXqfLk3pQR2, jYfTtdqGJ13CcW4X, HyzNBTmx9KqgAJ1h, B1jMTK283G4gtb6d, 87yxQDJMZfwgqrF3, DrnNChtLJb8pmAVd,
RfcDWKCHXvFZBy1x, ftTngaZr73hXKHLq, d9nFfyRzgZqcvBWx, X3bnZG1MkRT4xDAL, dPg8aMp21ckD9TCH, NqKayk8Q9rwWvjpz,
mCJFDgKxk9NfRGtL, gvTtPzKkmXnGCQyh, cbfAWCMDvQXj8LJP, 89yPg6z4AjZGfWRm, FVv7P3a2KqJrdG4g, 4ntRLypA6zd8YBQm,
QxT9aDgwJMcp3F4R, 9DtFQvLG6TzgnJH1, zmgHQXYjpcnC8MrK, rZG7RAbXYLQdD9nz, qTFxWz8QmJjwBdgy, GyaHhzDfZRJpwjWL,
WwV6DmHbYB7X8Fh3, cxgRYMkV9vZQaPWJ, thvrYzLKGB97kQZC, G7X1CkT8z6PrfjWM, WdXQa13P796xGF4k, y1pZvGw2HKBcfRTQ,
VnczrWJL6YKZ2DXP, AmQvw4LWKfCN71VX, hPAzHcMYGZ3w6mva, MgDw2bXrKTmWaLzY, Pbf124y7QWT6dGhr, jXxvrGqCFaJ8NAPh,
zfaGW68H1jp9LRVm, 3tRNmGkZB2Jhnv6b, 3jARkacpG1LJgCTN, n4KWZTbvRfXw3HPc, H84JRbhBNZTaLVQ3, dhG3xvZJTkRDycAr,
6kQyNWXpP8mTbR9x, NZXhmPJaAzCDBb4d, PkQBHzRa4YyCKTfb, JnQNajVrBKq9tR17, MKmntyr82Q4v7jDh, 6Y2mMtCzb7ZvQKWD,
KGbxV6BjrZ3QkdDL, PDjxyrdVvWHCtRMX, yp7wgNfmFRYdqjBc, qbtZdThvBCXVcfAW, krjVCfv19TbQgJH2, NdntYGabAvy6LM1g,
wKNnB9JDtv3RPVxW, Bmkht6z4cjAKa2WT, d4LFCkmQYhPRDx2y, Y1tBfF2c9yvRZWxN, 2AH37tCbvQfwJV8p, F68TMBPYkLwX7C3v,
ZDYjbLwfJTaV4GHq, WayYc9LQVPKmwAX1, XTZRWgGDrwYJvxM4, V47pD16qPCb9QJjM, Yb83B2kvHA9gRLGn, XaF63gxRH2v1TVBw,
BPf6K932MFQAxZkL, 4CNvhR7ZMrkxPwbQ, 1c2nhjYPARfVZBdv, vyzbt6kgRhTq4fdr, DRv3TVCwrAPJ9zK8, znj2Y79BCdbQgTJ1,
vYNVgrT9BQnyD1Hk, PwMavYLXV8GztqFd, 7gPtBMrHwGyjWKRc, Qx8JpNFC1hy7Rcrv, b461jPtVcmwyGAqN, BKLwZFjpybz984PN,
mzgDjAhkWY8bdnxr, rtbzWXcByFwk7G8n, Y479CJ6Hb3Zt2zPk, H4NThcMa2dfR8gGr, Ragx4ZHt98MG6qkr, gk2jJvTHhW4zNDKB,
4PCDJtkXhry9MW13, NDCn7Zy4dBzTPYGg, g1kdNFQh2c7DmAqt, Hr2k6NDLjKZA8xQg, qCgr4VRyj8HcaznF, rajyY8pFtKfd9QDn,
kBaRfmcVqbYypN4G, hBT4R7JfKgG6pNnt, LH4aGRZt7wVhkpcD, Pjkc2mBbhdYnQ9ZK, TJxaKCntRkGX3Dj2, 4AYDzX3qhLF86mbW,
6Ab23vqJ9z1pRmH8, RGa4ZVtzjbh3YQ6x, c9pH1QPFxA3mGtvd, nqDL2hKfQB4yvjgJ, cmqRf34jdgzTxWp1, B6xYXRgnPWaMJFK8,
bvfzYR3PaZ9HdynM, hAT34VWRB8aXjvk7, Ccvk1yDbBzFZAVqG, dMX79JcWz1FKPm6f

Gold-plated, the 322 reports with a counted swing:

7dHJyYKgf6WB9TpD, Wa2JYTq7tk9LjhRA, j342aPRr8Vg96mxv, FKY7R8vj3xTAMrHp, 8DgRxkGfPHCYQXnN, ZQvC4F7mPHxftAb3,
bX29KarfQRWkGB4y, 6ZkgW4DwMP8RdbBG, 4xdA1HG6ykcTbvPB, DBd1HZM3YQ7h8zpb, DPGXCjVdmvhyL2fn, 8hgcQY1RkB6MwFAa,
hvzCx1FQ8Bq6Zn3V, zQX4tA3FRYr1Kmwj, FfVHgkQpy4Y8KZ6n, 8fGaJA1zMHZFyWN3, 8rW9vbYx27AXf6nk, zXkwGd4jQfYNR9gJ,
3jRP9wApBrXxCyav, nwacBPXR3ZJqYQdv, 2qcmHx7QGDz1w8r9, Zy3P1DpzhALvR2Gq, wHbyf7XFxLRkCpAW, MHjhartn3AxwBpXL,
W2TRnwfz9cLFr7gv, Tm63MLtxhwpncqdN, xv6qP7Z4zkgyJdT2, 62qMKV9TgbvXzAfw, jbf2HXtNRzdArYpB, bAf9WdKnGpyY2aRB,
GNhPk8tby6wC3Vnd, tapwDKg13nmTXWPy, 2WDqG6H8hac17PZK, aQqpwnMrNmkyBZYH, 821Lbp3wRQrHmKNP, mH7LaXAzcrKk3Ffq,
bVmZBjTX68Nga9CG, qCr3MjH1LRwzTdWF, dRkYncbyFh27jr3t, kcLdxbB7aqXmjKTh, Zvkty2FVxwzCMbXd, 6Wna7r19qQKDP4yV,
ZqHf41jvV9GM7Bnz, J8YZxmQ6GB14yMr9, 9DJGKCWqvZf3apkX, F67dNtKPnmQb9fvw, tWp4w8b7TnxhJHfP, hPYpkxrVKRZJ6D7j,
Zq2GzC4Fvy3hbDnN, Y1NQrLjC6AZnkKVp, kyHYDjaT7F2zmQtA, XmtN7TFqRp1wykLZ, vdKwWbcHPFzh62a9, H6hMXfDybRFVApkG,
gnfPKV9DLC3qXmy8, d9gct8wCB3VGmrK6, 3WCzBtTFZKH9y7mb, 8Mv36mpFnJ4G2btY, Lphqk4fPn1BVKbW3, Dm4Cxkb1f8rAWM7v,
LHmG7FykZfA4tg6h, Jy7x384jfrgVQAMY, 1qa6bp8FvwZ7GRhD, 4A6VDmtjhcpB2YrL, QjrgXcZWRTqnMAFV, 83KD4dYAtWjrg9xn,
TnVQ3YHbqMDRrxv9, GQxCXNjrYbTBJ8K1, AjWpMQrnk681gTta, 7dcWCq6AKzrb92nN, FMapyHhwz7qPG3KB, dbKR72zmgNcLy1TB,
6vwYGgdFNXTrCfD8, Dkp9M2gHWXvfPFJK, gkNVCqXjhQt7DZB1, XwMvZqJ3Hm6NVDWh, YJN48Fh2QqRLkwaB, 1twaKYGX6FZnWV3y,
DJBnfy6d1Wz8KRqH, 9L4hGr1bwmqAMnWf, Kf3BgWCLjMr29QpA, FX3KrvQ2Dag8ZPbT, WJPkYxQhTZw1LXrd, CY9BNHhZ4D27Xc6T,
yZ39dLJ8McXgtTPw, pDCXzdPGY28WKagL, Hhv4QZCMDRyKPF6d, VacF6CqyjLRHp34m, NaMJGgXB6trWwyPb, bhHyGVrP7MkJvtf6,
KJ3xfRDvHLdtYcPG, MFtRpDVkZjWTHJwz, jtgZTL92cpQV7Kam, Gbrm19NqXQ2a8ndV, Z9jPDJvKRWVtHpCf, jcHC6LMPw4DpvkXY,
9vKbDJgYwkntAprm, C9jZwX2kJYBDW1xm, RMkDbFj7wnQJV3vT, D9Ljta1QzGAV8p3r, XA8gMNrqQJakKCFB, wxmvFq7dWj9KyBTD,
A2yDHYB18fKr6FJx, ktdh9VjHNp1R4P6B, FBVvjf1bLmGMDh6A, 4AHmGnP78vgdTFrf, fvVc2qzapFBGn6Jx, pmvLh8dcMFXTgfDN,
TCk49NrjMJ28GWpV, hk6yL9KjQBfgnAFJ, mZ8whFNxzBDWP1y2, rTKxLH3FyAgmNaXG, 3CAgGDrWfKF8PVBk, VC8hRDpv7gNrZabf,
kzBmJ1T3KA6nLvZQ, mpwZ68XKaAhBn1FH, DWjgPMmLFrZ4RacV, dwakQVz3YLNmftKj, XvwRpVFxgPNG3LbW, kDLC6Km3f4ghPwyX,
xcp2w7DCma3vN1Mg, Tp3CbBzgPWjKh7x4, fqMQK9daJZ47WLwz, VB6h49wfknpLjJZa, mBt3bGv8jX69dwpP, 9XwqjatdDGBMT8Wk,
2dMA1wpLcCG8fVkz, Qb4kRc6yKBT2f1XC, fn4jGwgdJCRTFr6D, M1GaftLbY38xTmRw, 3wK4c7n9avYdykVT, 2Jkcq3LjCKXB9MG6,
J6FCMqv7khrL3bpP, wPagd9YpZz72GWFJ, 91TVk6pJG2Afw7mC, DtFRCdLhjTVZ31Py, cQR2JVmj9b1n4L7p, bZXRqLpg6fxH9KkC,
thg4dwFG98qMTHNy, bNyvF1M2XC9gG3BD, 8gwWdRT1NDpBhVcJ, tVDRKYQapH9AkxwX, B9C8TVwHJZk6G4F7, hNtMzKGBkfvZXqRg,
L1YZF2PynCdbXzx3, PBkXcr2JGgx4NR9Q, HVqzTcj2aLgJPmB7, 8hTP3mzRM9XVbawW, a42r7kCB9p81DPvn, pNAzkK7MY4RvcVyQ,
KzYBRnrkaX3gCFbA, TFMYA6Cz3N9HKxX7, YAyD8grNcbFH7Zvm, BT416hmrQPtqx79a, Mx6A1bjq3dy7fmc2, dLX3JtnFRNymYZ9C,
7mxjzt4DfGMZ6TgX, ZG4W7AFmfzqKLhYC, B69h7FP4HTRbWmaV, nak6GtNdqJw3mcrj, ZgM4pwAFDjm8x9R7, pnAvmYVz7jw84fGR,
Y1Wqcg6w7BymQh2L, gTZdj4GM1yQvRYtN, dgaxVBLNyGKb7TQr, BRyNGMPY96zrnT8F, b6fYJmdBAHrxwvqW, aMwchZFXtBdfz6Pg,
9Vjt3BZw4cYD8zmd, R7fw48zhb9ACYtVq, 2vQcF8YKdHnLCrBN, HQC93xvpFwnrmyNV, A1M83b4jXrWRK6gz, m3rVcFxPgJDTRd9y,
BycYP1qn6bw7HCrV, 9Z2yTWHBMvatDKxb, QwH69ngLG7yfTkzr, 1VfBdPt6JDZQ3wX7, nhtA6zwvkQjqFKMm, zYmZ4ArjnWtvCqHP,
bdcvwjJrQmLVY9fP, g3yqM4DtzwkW7FGK, TGqpwnN7vQPdWcbV, mzcQ9vbkR4nTB7LH, x2WKajR61wQ4BdvF, ZKG3k4Lnhdmgr8j1,
tgzN739wYnGfW8rQ, 84W9MpqLQkDT3fJH, Trj4wqJR1BCk2tMH, bDMQn7pV28BP4HKL, YvBZ9Mc4ztyLFbR1, BnvaZrdQ3ygkhf9x,
MTgVc6w94AtKLQpk, aTktVmBCM4g8vdYx, JcNHWGxnTh9mwvBD, cjTv8ACq3y4XdxQa, 9nbzya6gFtvcTZPh, hdxJkZnMzvQ3KbtA,
Qm6RtXMDGBPvhxC8, qnpPwbNYBZMf491d, ptwbJA3achHmQxqB, MTwXJg3v8RdLNpY9, DhzvT6fPQjc4ygbk, Mx8Az2XhNR9PjJQ4,
jC3PWRVfTNQpLdaJ, nbDwkWVNhcpBGfXM, 2wg3NaAh4Qt8bJcr, RAgjxHhDLWVTJ2mv, jtZvgmBzfqXwdxM8, 6vWPz9xgLHr3Tpb7,
41Fxfz6Vc3ZJgXjY, hN9wYPyq7HbTdBVF, bgwMVyYd4LAZQvXh, HZTBKCW8RfYQxNDn, QxaZ3wVBvgpfr6MC, rF8yVmHXcQZb4tGA,
ntbCK7RZ8BLdvW1h, HbD6y7PaVKBLW2Cz, Dgmd4crXnyfapVNq, nN9XQhjfT7FxLtyB, xRAMwrk7GgdNqL19, NbwPAnvLaBZ8Trkf,
nWbRZ8GdptYkCjvK, KMNXJn3gfxPb6hcG, CMYR2Z4kXaPJFBG7, LcBTWAwMNvjq3Qda, 4mtapjhwPdZ9LWAn, kKqtP1XZT8YGbj6g,
1GtfnPp2k98Kv64F, RGpw8tJ7cgaj6bzf, ZqFcJdfa2VKzyvmW, pyY6Kg9bPRkQfMHL, XjLK4b67frpFnh1m, JVWykrTQB6PZzxHj,
n8d1pHcjXzrZTM7x, agztmnCXVNRGqwvx, ZtPATcqmdFxKvaBR, HBXA4RjC91yxbNQV, pVdmf31ztah8cNj6, Nc8VmvRdjKWhgDwQ,
LDpn3MJfr6v9my2F, nKNgy3VrZXkcdjBT, trgpWwRhD2GjVcK4, txD6pv3aBKVkHQGC, bVxfq4YPLFGpkT9X, QnbwKFxGraygkpAv,
aMF1JcGxkDpjnYH9, C4dnZJb7yYfARQGN, DmVkKqcQyvx7FRaN, T9dcWVK4kayGtnHY, qkAc48TbvKmJaLY2, VCaB6hMPrxA8bYKJ,
XnQjTd4bh6z7kPDB, bhY6kGNWtn7amcBJ, zf26hjagkZbFV1rJ, rcFvTxgkVp49jDCG, ARC6jdMJG4zNTtBV, DPZBhwmV4cHkyfab,
fxNam3ZBVbdJ4Hp6, t7bRZVYApzMGgJW8, 6RvQJGgkZqyW9Pm7, VGXx3tCWpNAFTrMh, KZ1DBj6kfmt4xLpX, 28fkDyHTm76YAgVL,
awxJTVtfzGC1hvD8, qrfRBgWAP76y8JFv, BfzWYwLADh1HnPk4, B6jfw4mcnPxWCKYD, aVqbfWYCMP8zrwFZ, 9yNQP3FWHgcVvYDq,
X7T93BGj2DcvNkLt, N3MXVw9xrJqdcyQt, MQzcrtmPn8pYJ4WF, qCpQc6T9FVDARfNg, xPhaCzTjybrZ48gV, bWfZPBHG17dxgv6y,
gGM6bHhJwTXm718W, zZvPT8KkQR4qftyj, 9TVG1QyMjZrhtPWY, xFCGzPZBL1MwaW8h, 9ZcCphv6nQ3wkd1P, 1HKqNrgGJpt6Tvnx,
zdv2nL6XxaBm1Z8A, mvLQz21GRrKnWfPd, rfGhbTRZKtjBxzvX, xPC78Y6JRpgLDWwq, j7tKNzr3YQXTVR6n, 9LBdTqpzGHyXWanm,
Phj961FCXx2KZWwk, 8z4xqtkQ6jZaPm3f, gGLmCaHknFdjTtJ9, 1h9P4rYzWDvtHRxb, nyWj34DhL9w1YFAP, PKRdnV4qapxMH1Gt,
GNp7bXWgfjr3yAcm, QD8ypCXWJcv9Y2B3, BYFrZ4jf1zJbVkvd, ngCkNVyZRPzM6hYX, 94j2QJYfxcR1gvdD, b6C24Jd3V8WwpFx7,
hjkVZTFrGbRv7cH8, 7vwKQyRBMt4kD3Td, bcqT8k3XQpKBCfJv, ACvNpyX4DafTgcZw, M6A1N8acHRJ29hVG, rFXNJymAgGtMh9bL,
4q3p1Vhb2N6Q7rFY, 2kFjvxn3MzRLQWTD, DMKd8QbWqYGyanV2, DGcAwmbJY2TpdB83, 1PhWaJAyYQRFHcVB, RBPX86Fg9CxqhTD4,
pxyvdRAQ4kZ2zLYa, FqVXJryknZYWTvKH, 6yfhjmC8tJYNk7Zd, 6CbctYwZzTX9qyjJ, Dcfmnp2VTQR7t8xw, LZVpRYxDC6aygfMF,
Qh6PtMBaFcR194fb, rWX7GbVdjg6vTfFY, AZLYpdc6r7P2CMRm, 1mJ9QRpLy68NkPfw

Stoneform Shoulders, the 25 reports with a counted swing:

FVqQab31p7KXAJzt, jN7MvwCfz164kYyZ, NXMZV9xjvWpB7GtP, zmDY4LVX7N8vh9aR, B9gTmLcFW3R7GAHt, vFQhXgw42KmfjT8W,
ZDWyQGqFBvagT7kA, dBNfgVDLMmnc8h4v, NrpM7DJQBKWqbzVF, v9CxPFNqVjnbkZM8, m43B9RfZpPMC21xr, N9PGBMt7faY6W2J8,
3GYRdvqbPFr8yp9K, 81LcPyjXVbm2xGa3, aYyJ9hdFwXKgDk16, mbpQYGL8jaqgdrAz, XbDfcTpMyZ986arw, 4Tb63aXJykLnKWgN,
QkVRc329w8nLTzKp, hDJKcgjxvWQfHt9M, Ck2vbA1pfqnKDgYV, 3jARkacpG1LJgCTN, Tnv3yPNtFKXQhgH9, cav7VZrmPbNW3y29,
wrzB6GtjPaqCKJZW

The Gift of Arthas crits come from the reports of its swings and zbpHhqxwNAZcj2Wa.

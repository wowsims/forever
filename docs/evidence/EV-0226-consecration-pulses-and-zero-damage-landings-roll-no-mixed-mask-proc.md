# EV-0226: Consecration's pulses and landings that deal no damage roll no proc of a melee, ranged and spell mask

- Claim: a hit of Consecration's pulse rows (1280345 to 1280349, Treat As Periodic and no Not A Proc) rolls none of
  Bolt's Charged Eyeball (trigger 1311505), Cursed Murloc Eye (1292674) and Vengeance (20049), and neither do Deep
  Wound and Ignite, the other Treat As Periodic hits counted. A periodic tick rolls the Eyeball, whose mask holds the
  periodic bit, also when a proc laid the dot (Deadly Poison, Wound, Immolation Trap), while the direct hit of a proc
  (Instant Poison, Lightning Shield, Fiery Weapon, Shadowguard, Touch of Weakness) rolls none of it. Flametongue
  Totem's crits set off Vengeance, which can proc from procs, as Fiery Weapon's do. A spell landing that deals no
  damage (a debuff, a dot's application, the Consecration cast) rolls none of the Eyeball and Murloc Eye, whose masks
  pair the harmful-spell bit with melee and ranged bits, as it rolls no Highborne Research Tablet (EV-0021).
- Sources: Warcraft Logs (forever.warcraftlogs.com, zone 3000). The Eyeball: every player whose gear lists item 280032
  or who deals 1315996, 1,014 players in 922 reports. Murloc Eye: wearers of 273840 in the reports from 2026-11-26 on,
  396 players in 329 reports. Vengeance: paladins who gain 20050 from 2026-11-24 on, 3,463 players in 2,601 reports. The report codes are in the
  scripts' inputs and are not listed here.
- Game build: 1.60.1.70862 to 1.60.1.71139 by each report's region and start, against the time the CDN first listed
  each build (70862 from 2026-11-07 04:43 UTC, 70888 from 2026-11-10 15:11 UTC in the US and 2026-11-11 02:19 UTC in
  the EU, 70938 from 2026-11-13 01:36 UTC, 70959 from 2026-11-17 15:19 UTC in the US, 71037 from 2026-11-20 18:23 UTC,
  71091 from 2026-11-24 15:21 UTC in the US, 71139 from 2026-11-25 19:14 UTC in the US and the EU). The Eyeball's
  reports by build: 2, 13, 135, 117, 279, 134 and 242; Murloc Eye's are all 71139, and Vengeance's 784, 706 and 1,111 from 71037 on. The first and the last raidbots hotfix cache of
  each build from the oldest file raidbots' list serves (70938, 2026-11-13 18:31 UTC, push 114060) to the newest
  (71139, 2026-11-27 11:25 UTC, push 114448) hold no record of the rows of 1311505, 1315996, 1292674, 1292675,
  1280345 to 1280349 or 20049, and the same records of the Consecration casts (push 113341), the
  ground's mark 1280353 (113792) and 20050 (113542). No cache covers 2026-11-10 to 2026-11-13; no hotfix export after
  2026-11-09 21:42 UTC lists any of these spells, though only some exports carry a spell diff. The hotfix each player
  ran is not recorded.
- Date: 2026-11-10 to 2026-11-27
- Pull request: not recorded

## Method

A roll is one of the player's own events: a landed hit by ability, a periodic tick, a debuff apply or refresh on another
unit with no damage of its own, or an avoided hit (miss, dodge, parry, immune, resist). It counts when no other landed
hit of the player falls within 300 ms of it, so the proc that follows has one candidate, and when the item's cooldown
has run out (the Eyeball's 3 s, with no proc in the 3.1 s before the event). It procs when the item's damage (1315996
or 1292675) by the same player follows within 75 ms. A landed hit is any hit type outside the
avoided ones, a partial resist included. A fight another report holds for the same character is counted once.

A Consecration cast counts as a roll on its cast event with no landed hit of the paladin within 300 ms either side.

Vengeance counts a gain (an apply, a refresh or a new stack of 20050 on the paladin) within 75 ms after a crit of the
paladin that is its only crit within 300 ms, split by whether the buff was up. A crit counts only when another gain
falls within 20 s of it and more than 1 s away, so the talent is in play: in 6,112 of 17,789 fights, counted once per
paladin, the paladin gains none at all, and 10,600 Seal of Fury crits fall in those. A crit at five stacks logs no
refresh (EV-0224), so the rate with the buff up reads low for every source alike.

The rates count every stretch a player is selected for, stretches with the item swapped away included, so they
read below the trigger's 15%. Within 60 s after a shock, an isolated white swing past the cooldown casts one 12,040
times in 84,290 (14.3%), from 14.7% at 3 to 4 s to 13.3% at 20 to 60 s, and 15 of the 48,906 gaps between two shocks
under 6 s are shorter than 2.75 s. The zero readings stand against the landed hits of the same stretches.

## The Eyeball

| Roll                                                                                                       | Procs  | Rolls   | Rate  |
| ---------------------------------------------------------------------------------------------------------- | ------ | ------- | ----- |
| Consecration pulse                                                                                         | 0      | 7,797   | 0%    |
| Deep Wound                                                                                                 | 0      | 9,160   | 0%    |
| Ignite                                                                                                     | 0      | 482     | 0%    |
| Consecration cast (20924 and 20923)                                                                        | 0      | 3,754   | 0%    |
| Periodic tick                                                                                              | 11,233 | 168,545 | 6.66% |
| Periodic tick of Deadly Poison                                                                             | 540    | 6,657   | 8.11% |
| Periodic tick of Immolation Trap                                                                           | 71     | 1,289   | 5.51% |
| Periodic tick of Wound, Skullforge Brand, Seeping Willow and Venomhide Poison, dots a weapon proc lays     | 272    | 4,413   | 6.16% |
| Landed direct hit, damage shields left out                                                                 | 30,376 | 473,890 | 6.41% |
| Damage shield (Retribution Aura, Thorns, Fire Shield, Holy Shield's block damage, spikes)                  | 0      | 42,907  | 0%    |
| Direct hit of a proc: Instant Poison, Lightning Shield, Fiery Weapon, Shadowguard, Touch of Weakness       | 0      | 1,328   | 0%    |
| Landing with no damage, all                                                                                | 2      | 118,402 | 0.00% |
| Of those: Hunter's Mark, Faerie Fire, Sunder Armor                                                         | 0      | 16,836  | 0%    |
| Of those: Demoralizing Shout and Roar                                                                      | 0      | 9,288   | 0%    |
| Of those: the application of Corruption, Shadow Word: Pain, Serpent Sting, Insect Swarm and Curse of Agony | 0      | 27,616  | 0%    |
| Avoided hit                                                                                                | 3      | 69,732  | 0.00% |

## Cursed Murloc Eye

| Roll                                       | Procs   | Rolls   |
| ------------------------------------------ | ------- | ------- |
| Landed direct hit, damage shields left out | 107,163 | 122,860 |
| Consecration pulse                         | 0       | 2,888   |
| Periodic tick                              | 1       | 71,954  |
| Landing with no damage, all                | 17      | 41,791  |
| Of those: Hunter's Mark and Faerie Fire    | 0       | 1,824   |
| Avoided hit                                | 17      | 23,095  |

Murloc Eye's tooltip reads "Spells and attacks deal" with no word about damage, and it needs damage as the Eyeball
does.

## Vengeance

| Crit                  | Gains with the buff down | Crits  | Gains with the buff up | Crits   |
| --------------------- | ------------------------ | ------ | ---------------------- | ------- |
| Consecration pulse    | 0                        | 38,843 | 0                      | 173,358 |
| Seal of Command       | 6,043                    | 6,146  | 17,119                 | 22,197  |
| Seal of Righteousness | 2,102                    | 2,194  | 13,076                 | 18,599  |
| Seal of Fury          | 135                      | 139    | 520                    | 750     |
| Hammer of Wrath       | 403                      | 436    | 3,593                  | 5,758   |
| Fiery Weapon          | 805                      | 818    | 2,301                  | 3,285   |
| Flametongue Totem     | 159                      | 163    | 472                    | 625     |
| Holy Strike           | 4,538                    | 4,613  | 11,539                 | 15,899  |
| White swing           | 24,058                   | 24,462 | 63,199                 | 83,051  |

Vengeance's row carries Can Proc From Procs, and Seal of Righteousness, Fiery Weapon and Flametongue Totem, procs
without Not A Proc, set it off; its mask holds no periodic bit.

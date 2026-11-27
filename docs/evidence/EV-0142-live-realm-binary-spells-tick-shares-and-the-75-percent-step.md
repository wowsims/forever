# EV-0142: Spell resistance on the live realms: binary spells, Spell Piercing, tick shares and the 75% step

- Claim: on the live realms a spell resists all or nothing when its damage row carries, beside the damage, a non-damage
  aura on the target, an interrupt, E_DUMMY, E_SCRIPT_EFFECT, E_ENERGIZE or E_CREATE_AREATRIGGER on any target; an
  A_DUMMY aura on the target does not count. Its hit roll resists it whole at about the target's average resist, and a
  landed hit or tick takes no partial resist. The Spell Piercing bonus is the negative side of the partial resist roll:
  a landed binary hit takes it at the full share, as do Insect Swarm's and Mind Flay's ticks, while Hurricane's,
  Entangling Roots', Bane of Agony's, Bane of Doom's (on few hits), Curse of Agony's, Immolate's, Immolation Trap's and Serpent Sting's DoT ticks and a hit or tick at a tenth of the share take none, and a hit or tick at the
  full share takes it whole. Devouring Plague's and Pyroblast's ticks meet the target's full resistance. A channel's ticks meet a tenth like a DoT's, Starshards' excepted, which meet the full resistance; the
  missiles of Arcane Missiles are hits and meet it in full. No partial resist takes more than 75% of a hit, and damage shields take none.
  Spell Piercing beyond the resistance gives no bonus while the share it makes is under 1%, and from 1% the bonus follows
  the share; Magma Totem's and Searing Totem's attacks take none of the shaman's Spell Piercing.
- Sources: 213 combat logs from 2026-09-27 to 2026-11-16, casters of every class; 36 live-realm logs have hits on mobs
  with a resistance. The player streams of 992 Warcraft Logs reports of Forever (zone 3000) from
  2026-10-14 to 2026-11-16, of which 146 live-realm reports have hits on mobs with a resistance: 2 on 70862, 16 on
  70888, 86 on 70938 and 42 after push 114115; one report was checked line by line against a combat log of the same run.
  For the binary rule, 2,158 further caster streams from 706 live-realm reports (2026-11-06 to 2026-11-17) that the first
  pass never read, with the rules written down before they were fetched. For the channels'
  tick shares, two studies whose predictions were written down first, both of 2026-11-17: the drain study
  (`forever-resist-drain-bombs-1117`, RESULTS.md) over the 1,322 live-realm reports with a control or test hit among the
  first pass's reports, the 706 held-out reports and 3 more, and the tick share study (`forever-resist-dispel-1117`, RESULTS.md) over the 706
  held-out reports and 866 more fetched for it. The items naming a count section list
  its own
- Game build: live realms from 2026-11-05 on 1.60.1.70862, 1.60.1.70888 from 2026-11-10 and
  1.60.1.70938 from 2026-11-13, the last reports after the 2026-11-16 hotfix push 114115 and possibly the 2026-11-17 push
  114122, all before the 2026-11-19 hotfix after which binary spells take the Spell Piercing bonus; the hotfix a run was
  on and relogs are not recorded. The items naming a count section have their own
- Date: 2026-09-27 to 2026-11-17; the items naming a count section have their own
- Pull request: #1080; each count section names its own
- Later count: the Incinerate, Shadow Bolt, Death Coil, Shadowburn and Frost Shock lines of §Spell Piercing and §Rows
  with few hits, §Count of the 2026-11-07 to 2026-11-24 reports
- Later count: the 5 and 6 piercing reading at level 60 of §The 1% cut, §Count of the 2026-11-27 test

## Method

A target's average resist a for a spell is its caster's own share on the same mob and school in the same log or report:
the damage resisted over the damage before the resist, on at least 3 hits of nukes whose rows deal damage and nothing
else (Wrath, Starfire, Moonfire, Lightning Bolt, Chain Lightning, Flame Shock, Shadow Bolt, Searing Pain, Soul Fire, Mind
Blast, Smite, Holy Fire, Fire Blast, Scorch, Fireball, Pyroblast, Arcane Blast, Arcane Explosion, Arcane Shot, Fire Nova,
Lightning Shield and Instant Poison). A pair of caster and mob counts from a = 0.1. Expected counts are those of the
partial resist roll (the triangle around a), at a for a hit and at a tenth of a for a DoT tick, and a full RESIST is expected at the rate a. Rows logged
twice, in overlapping combat logs or in two reports of one run, count once. Intervals are Wilson 95%.

Builds 70888 and 70938 and push 114115 move none of the resist counts below, and the share on the level 60 dungeon
mobs holds across them (Chromatic Dragonspawn Nature 0.512, 0.498 and 0.512).

## A landed binary spell takes no partial resist

- Count: §Count of the 2026-11-19 to 2026-11-21 reports

| Landed hits and ticks                                                          | Any target  | Pairs with a ≥ 0.1 (expected) |
| ------------------------------------------------------------------------------ | ----------- | ----------------------------- |
| Binary direct hits                                                             | 0 of 43,465 | 0 of 62 (60.5)                |
| Immolate, Shadow Word: Death, Arcane Wrath, Incinerate, Ice Lance, Flamestrike | 0 of 19,967 | 0 of 352 (347.4)              |
| Binary DoT ticks                                                               | 0 of 12,082 | 0 of 63 (22.1 at a tenth)     |
| Mind Flay ticks                                                                | 0 of 9,242  | 0 of 372 (127.5 at a tenth)   |
| Hurricane ticks                                                                | 0 of 28,930 | 0 of 223 (223.0)              |
| Other direct hits, for comparison                                              |             | 6,962 of 7,382 with a partial |

Full RESISTs of binary spells run at about the share: 203 of 617 casts on pairs with a ≥ 0.1 (226.1 expected), and
Hurricane 178 of 401 (183.9).

## Which spells are binary

Five rules were written down before the held-out streams were fetched: an aura or an interrupt on the target only; that,
or E_DUMMY or E_SCRIPT_EFFECT on any target; the same on an enemy target only; any non-aura, non-damage effect; and any
non-damage effect at all. Each tested spell needs at least 8 landed hits on pairs with a share. Shares are the caster's
own control nukes on the same mob and school in the same report; Ice Lance takes the caster's own Frostbolt RESIST rate on
the mob, since every other frost spell a mage casts is binary. Mobs whose resistance comes in windows are left out.

| Spell (row beside the damage)                                             | Landed with a partial (expected at the share) | Full RESIST over the share |
| ------------------------------------------------------------------------- | --------------------------------------------- | -------------------------- |
| Immolate (E_SCRIPT_EFFECT)                                                | 0 of 286 (283.8); ticks 0 of 758 (207.4)      | 1.07 [0.95, 1.19]          |
| Shadow Word: Death (E_SCRIPT_EFFECT, E_DUMMY)                             | 0 of 145 (144.9)                              | 0.93 [0.80, 1.06]          |
| Arcane Wrath (E_DUMMY)                                                    | 0 of 511 (486.1)                              | 0.94 [0.84, 1.04]          |
| Incinerate (E_DUMMY)                                                      | 0 of 25 (25.0)                                | 0.93 [0.60, 1.32]          |
| Ignition (E_DUMMY with no target)                                         | 0 of 17 (17.0)                                | 1.09 [0.65, 1.60]          |
| Ice Lance (three E_DUMMY with no target)                                  | 0 of 74 (72.7)                                | 1.01 [0.82, 1.21]          |
| Flamestrike's cast (E_CREATE_AREATRIGGER, A_PERIODIC_DUMMY on the caster) | 0 of 607 (599.5)                              | 0.89 [0.80, 0.99]          |
| Earth Shock (E_INTERRUPT_CAST)                                            | 0 of 310 (306.6)                              | 0.99 [0.90, 1.08]          |
| Mind Flay ticks (A_MOD_DECREASE_SPEED)                                    | 0 of 1,637 (570.3)                            |                            |
| Bane of Agony ticks (A_MOD_SPELL_DAMAGE_FROM_CASTER, A_DUMMY)             | 0 of 439 (149.0)                              |                            |
| Flame Shock (A_DUMMY on the target)                                       | 989 of 989; ticks 960 of 3,292 (987.1)        | 0 of 989 (392.4 if binary) |
| Wrack ticks (A_DUMMY on the target), shares of every caster on the mob    | 140 of 492 (build 71037)                      | 0.09                       |

The plain nukes stay partial and never resist whole: Arcane Explosion 15,065 of 16,264 with a partial and RESIST 0 of
16,264, Searing Pain 2,177 of 2,208 and 0 of 2,208, Lightning Bolt 770 of 774 and 0 of 774, Shadow Bolt 196 of 200 and 0
of 200, Fireball 246 of 272 and 0 of 272; likewise Scorch, Soul Fire, Mind Blast, Chain Lightning, Arcane Missiles and
Fire Nova. In the first pass, Mind Flay 87 of 205 casts (99.3), Earth Shock and Frost Shock 133
of 355 (149.4) and Bane of Agony 10 of 37 (12.0). Gahz'rilla Fang's 3742 showed 0 of 22 landed with a partial and RESIST
10 of 32 (16.0).

Two effects rest on one spell each: E_CREATE_AREATRIGGER on Flamestrike's cast, which also carries a dummy aura on the
caster, and E_ENERGIZE on Engulfing Shadows, read through Spell Piercing before the 2026-11-19 hotfix and not
re-measured since (below). E_NONE rows do not follow the rule: Smokey's Lighter took no resist of any kind in 3,300 hits
(0 of 179 with a partial and 0 of 179 RESIST against 68.8 at a mean share of 0.38), and Cloud of Chained Essence none in
4,246 applications and 15,460 ticks. E_TRIGGER_SPELL is not tested.

## Spell Piercing

- Count: §Count of the 2026-11-19 to 2026-11-21 reports
- Count: §Count of the 2026-11-22 combat log (the warlock's ticks)
- Count: §Count of the 2026-11-19 to 2026-11-23 reports (the Blast Wave and Curse of Agony lines)
- Count: §Count of the 2026-11-07 to 2026-11-24 reports (the Incinerate, Shadow Bolt, Death Coil, Shadowburn and Frost
  Shock lines)

The bonus goes where the partial resist roll goes, at the same share. E is the hits the caster's own rate on plain nukes
expects.

| Hits                                                            | With a piercing bonus (E), +10% : +20%           |
| --------------------------------------------------------------- | ------------------------------------------------ |
| Mages' nukes                                                    | 23,244 of 82,767, 15,747 : 7,432                 |
| Binary: Frostbolt                                               | 1,863 of 6,276 (1,802), 1,310 : 531              |
| Binary: Cone of Cold                                            | 1,111 of 4,675 (1,348), 738 : 366                |
| Binary: Frost Nova                                              | 418 of 1,781 (481), 287 : 131                    |
| Binary: Flamestrike's cast                                      | 994 of 3,235 (1,033), 666 : 327                  |
| Binary: Arcane Wrath                                            | 898 of 3,385 (872), 638 : 257                    |
| Binary: Ice Lance, Frostfire Bolt                               | 169 of 604 (181), 83 of 325 (91)                 |
| Binary: Immolate's hit (33 warlocks)                            | 194 of 1,827 (201)                               |
| Binary: Incinerate (22 warlocks)                                | 179 of 1,820 (195), 118 : 61                     |
| Binary: Shadow Word: Death (10 priests)                         | 19 of 315 (24)                                   |
| Shadow Bolt outside Improved Shadow Bolt windows (173 warlocks) | 619 of 5,980 (605), 420 : 199                    |
| Binary: Death Coil (144 warlocks)                               | 50 of 413 (46), 30 : 20                          |
| Binary: Blast Wave (10 mages)                                   | 40 of 156 (48), 27 : 13                          |
| Binary: Shadowburn (36 warlocks)                                | 73 of 639 (68), 49 : 24                          |
| Binary: Frost Shock (44 shamans)                                | 25 of 198 (22), 19 : 5, 1 at +30%                |
| Binary: Earth Shock, +10% bonuses only                          | 42 (32.8)                                        |
| Binary: Mind Flay ticks (10 priests)                            | 483 of 5,398 (388)                               |
| Binary: Insect Swarm ticks (13 druids)                          | 1,089 of 9,240 (1,051)                           |
| Binary: Hurricane ticks                                         | 19 of 7,577, all +10% (786)                      |
| Binary: Entangling Roots ticks                                  | 0 of 126 (12)                                    |
| Immolate's DoT ticks                                            | 22 of 5,953 (618)                                |
| Curse of Agony's ticks (157 warlocks)                           | 1,026 of 100,558 (10,825), 531 : 481, 14 at +30% |
| Tenth: Corruption, Siphon Life, Wrack ticks                     | 0.06 to 0.08 of E                                |
| Tenth: Shadow Word: Pain, Flame Shock, Moonfire ticks           | 0.01 to 0.04 of E                                |
| Tenth: Flamestrike's ground ticks, Rain of Fire                 | 7 of 2,454 (672), 21 of 4,655 (608)              |
| Drain Life, Drain Soul ticks                                    | 2 of 714                                         |
| Full share: Pyroblast's DoT ticks                               | 342 of 1,084 (297), 240 : 100                    |

A warlock with 11 piercing at level 60, on mobs with no Shadow resistance, took no bonus on any tick: Bane of Agony 0 of
281 (45.5 expected at the full share, 4.2 at a tenth), Corruption, Siphon Life and Wrack 0 of 512 (83 at the full share,
7.7 with a tenth of the piercing rounded down to 1 point), Drain Life 0 of 62 and Drain Soul 0 of 3. So binary Bane of
Agony's ticks take none, and a tick at a tenth of the share takes none at 10 piercing or more.

Mages without piercing take bonuses on binary spells at their fixed-modifier rate (Frostbolt 4.1%, Cone of Cold 4.6%,
Frost Nova 1.2%, Arcane Wrath 0 of 189, Flamestrike's cast 0 of 1,236), so the bonuses above come from piercing.

Shadow Bolt takes the bonus: warlocks without gear piercing took none in 212 hits. Improved Shadow Bolt's Shadow
Vulnerability (17794) is in the logged amount and not in the hit before resists. Until 2026-11-23 15:45 UTC, where
§Count of the 2026-11-19 to 2026-11-23 reports ends, 16.7% and 36.0% of the warlocks' direct Shadow hits inside its
windows log at 1.2 and 1.32 times the hit before resists (of 1,045 hits, against 2.0% and 1.1% outside); after it, 25.0%
and 39.3% of 336. With the windows' hits kept in the medians, Shadow Bolts inside a window read 2.24 times E before that
time (116 of 501, E 51.8) and 2.28 after (40 of 168, E 17.5). With the windows' hits left out before the count, as
§Count of the 2026-11-07 to 2026-11-24 reports does, Shadow Bolts outside the windows read 0.98 of E before (455 of
4,554, E 466.6) and 1.18 after (164 of 1,426, E 138.8, P 0.020).

Curse of Agony's ticks take none. Their rate stays flat across gear piercing, 1.03% of 82,285 ticks under 10 piercing
and 0.97% of 18,273 at 10 to 19, where a tenth of the piercing expects 0 and 274; their +10% and +20% bins come about
1:1, with +30% bins the piercing roll cannot give at these shares; and at 0.095 of E they read like the tenth-share
ticks above.

Incinerate, Shadowburn, Death Coil and Frost Shock take the bonus at about E:

- Incinerate reads 0.92 of E (P 0.13), and 0.97 against Searing Pain, Conflagrate and Soul Fire alone (130 of 1,260, E
  133.6).
- Shadowburn reads 1.07 (P 0.30) with Bane of Havoc's hits left out of its median. Bane of Havoc's hits log at 1.00 of the hit before resists, while the warlock's other
  hits carry Shadow and Flame's (426316) +10% for 20 s: Shadow (1293816) after a Conflagrate, Flame (426311) after a
  Shadowburn. With Bane of Havoc's hits in its median it reads 91 of 684 (E 73.5, 1.24, P 0.026). Inside Improved
  Shadow Bolt windows it reads 9 of 30 against 3.3, 6 of them at +20%.
- Death Coil reads 1.28 of E before 2026-11-23 15:45 UTC (44 of 311, E 34.4) and 0.52 after (6 of 102, E 11.6), apart at
  P 0.034 (Fisher), and 1.09 pooled (P 0.30). Its leech heal is not tested.
- Frost Shock reads 1.13 (P 0.29). Read against the hit before resists, the same Frost Shocks give 51 bonuses on 195 of
  the 198 (3 have no before-resists bin) against 29.1 with the floor: some shamans' hits of every school run at +10%
  over the hit before resists for stretches, from a source not identified.

All four are binary by their rows (Death Coil's E_HEALTH_LEECH is its damage, beside a fear aura on the target). On mobs
with a resistance, in the streams of §Count of the 2026-11-07 to 2026-11-24 reports, none of 386 Incinerate, 155
Shadowburn, 10 Death Coil and 26 Frost Shock landed hits takes a partial resist.

Open: E_ENERGIZE rests on Engulfing Shadows alone, measured before the 2026-11-19 hotfix, while a landed binary hit took
no Spell Piercing bonus, and not re-measured. Every piercing source row (aura 123) states no class mask, so the source
does not pick the spells.

### The 1% cut

- Count: §Count of the 2026-11-19 to 2026-11-23 reports
- Count: §Count of the 2026-11-27 test (one mage at 5 and 6 piercing, level 60)

A caster's gear piercing p at level L makes the share s = p / (p + 8L + 19). Under 1% it gives no bonus; from 1% the
bonus lands on about 7.5·s of hits, with no rounding. The rows are the hits of casters other than mages.

| Share s              | Hits   | With a bonus       | 7.5·s  |
| -------------------- | ------ | ------------------ | ------ |
| 0 (no gear piercing) | 16,370 | 0.49%              | 0      |
| 0.70–0.80%           | 19,515 | 0.11%              | 5.88%  |
| 0.95–1.00%           | 8,570  | 0.63%              | 7.44%  |
| 1.15–1.20%           | 37,324 | 8.01% [7.73, 8.29] | 8.91%  |
| 1.30–1.40%           | 27,348 | 8.81%              | 10.33% |
| 1.50–1.75%           | 21,941 | 11.17%             | 11.84% |
| 1.75–2.00%           | 12,556 | 12.51%             | 13.80% |
| 2.00–2.25%           | 3,796  | 14.23%             | 15.98% |
| 2.50–3.00%           | 1,710  | 18.77%             | 19.44% |

The rate climbs inside 1 to 2%, where a share rounded to whole percent would stay at 7.5%, so the share is not rounded.
The cut is no whole number of piercing points either: 5 piercing gives the bonus at levels 52, 56 and 57 (s 1.14%, 1.06%,
1.04%; 9.58%, 5.99% and 2.95% of hits, one priest each) and not at level 60 (s 0.992%: 0.64% [0.49, 0.83] over 20
casters), and 4 gives it at level 40 (3.64% [1.93, 6.78], 2 casters) and not from level 50 to 60. The cut lies between
s 0.992% and about 1.04%, so at level 60 between 5 and 6 piercing: one mage without Arcane Subtlety took no bonus in 265
hits at 5 (s 0.992%) and 17 in 231 at 6 (s 1.188%, 20.6 at 7.5·s). Between 1.00 and 1.05% casters read both ways: 3
casters other than mages took it on 2.97% [1.86, 4.71] of 572 hits, and 2 mages without the talent on 0 of 817 [0,
0.47%].

Mages: Arcane Subtlety at rank 2 adds about 14 piercing (the sim's row gives 15). Mages at rank 2 and level 60 with 0, 5,
11 and 14 gear piercing (s 2.9 to 5.5% with the talent at 15) took the bonus on 21.0, 28.3, 35.6 and 40.2% of 10,673,
11,310, 9,809 and 4,444 hits, against 7.5·s 21.9, 28.9, 37.1 and 41.2%. Mages without the talent show the same cut: 5
piercing at level 60 gives 0.58% [0.38, 0.88] over 11 mages.

Whether a resistance share under 1% resists anything is not measured. No mob in the logs has a steady small resistance:
against level 60 casters without piercing, 228 mob-schools show no partial resist and none a partial rate of 0.5 to 2%.

### Magma Totem and Searing Totem

- Count: §Count of the 2026-11-19 to 2026-11-23 reports

Their attacks take none of the shaman's Spell Piercing. The rows are the hits of shamans with piercing, and E is what
each shaman's own reference rate expects. Searing Totem's attack logs as "Attack" and as "Searing Bolt".

| Hits               | With a piercing bonus (E)           | Over E |
| ------------------ | ----------------------------------- | ------ |
| Magma Totem        | 186 of 26,257, 76 shamans (3,042)   | 0.06   |
| Searing Totem      | 3 of 3,160, 58 shamans (334)        | 0.01   |
| Lightning Bolt     | 1,468 of 13,823, 63 shamans (1,688) | 0.87   |
| Fire Nova          | 175 of 1,263, 56 shamans (161)      | 1.09   |
| Flametongue Attack | 345 of 2,702, 35 shamans (362)      | 0.95   |

Shamans without gear piercing read 201 of 63,498 on Magma Totem and 6 of 4,088 on Searing Totem.

### Rows with few hits

- Count: §Count of the 2026-11-07 to 2026-11-24 reports

E is 7.5·s summed over the hits at a share of 1% or more, and the floor is the bonus rate of casters of the class without
gear piercing on the same spell and reading. The median reading is that of §Count of the 2026-11-19 to 2026-11-21
reports; the spell median is the median of the caster's hits of the same spell in the report. The hits needed are those
for which a count separates the bonus from the floor at P 0.01 on each side, at the row's mean share and the median
reading's floor.

| Row                                       | Reading        | With a piercing bonus (E), +10% : +20%                                                  | Result                                                              |
| ----------------------------------------- | -------------- | --------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| Arcane Shot                               | spell median   | 177 of 2,100, 26 hunters at a mean share of 1.11% (174.9, and 84.8 the floor), 133 : 44 | takes the bonus                                                     |
| Immolation Trap ticks                     | median         | 0 of 191, 8 hunters (14.9)                                                              | take none                                                           |
| Serpent Sting ticks                       | median         | 54 of 2,114, 23 hunters (171.7, and 10.4 the floor), 53 : 1; +20% 1 against 57.5        | take none                                                           |
| Bane of Doom ticks                        | median         | 0 of 44, 16 warlocks (5.4, P 0.005); 22 of the 44 take another fight's gear or level 60 | take none, few hits                                                 |
| Frostbrand Attack (binary)                | before resists | 20 of 82, 8 shamans (9.6, and 0.4 the floor), 13 : 5, 2 at +30%; +20% 5 against 3.2     | takes the bonus, few hits (2.0 of E plus the floor on this reading) |
| Frostbrand Attack (binary)                | spell median   | 14 of 74, 4 shamans (8.5), 9 : 4, 1 at +30%                                             |                                                                     |
| Volley                                    | spell median   | 0 of 987, one hunter (84.4); +20% 0 against 28.1                                        | unsettled, leans none                                               |
| Explosive Trap, Instant and Deadly Poison | every reading  | no hit at a share of 1% or more; 80 to 155 needed                                       | unsettled                                                           |
| Touch of Weakness (binary)                | median         | 9 of 90, 17 priests (9.6), 8 : 1; the floor is not measured (1 hit without piercing)    | unsettled                                                           |

Arcane Shot takes the bonus by its +20% bins: 44, against a floor of about 1.5 from hunters without piercing, who put
0.07% of hits at +20%. Its rate is not settled: against E plus the 4.04% floor of all bonuses it reads 0.68 (P 3e-8), on
its +20% bins alone 44 against 59.7, and by share band (1.00–1.10%, 1.10–1.20%, above) 1.04, 0.83 and 1.18 of E.

Serpent Sting's bonuses over the floor, 53 : 1 at +10% and +20%, are not piercing: its +20% bins are 1 against 57.5, and
on the spell-median reading, where the same hunters' Arcane Shots read 161 of 1,910 against 160.0, 1 against 96.5.

## Tick shares

- Count: §Count of the 2026-11-19 to 2026-11-21 reports

| Ticks                                                   | With a partial resist on pairs with a ≥ 0.1 | Resisted share over the target's a                                                         |
| ------------------------------------------------------- | ------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Devouring Plague                                        | 256 of 256                                  | 0.95                                                                                       |
| Pyroblast's DoT                                         | 127 of 161                                  | 0.98                                                                                       |
| Seeping Willow                                          | 156 of 156                                  | 0.99                                                                                       |
| Rain of Fire                                            | 232 of 738                                  | 0.12                                                                                       |
| Other DoT ticks                                         | 1,267 of 4,309                              | Corruption 0.10, Flame Shock 0.09, Shadow Word: Pain 0.11, Siphon Life 0.09, Moonfire 0.12 |
| Curse of Agony, Immolate's DoT                          | 0 of 953, 0 of 558                          |                                                                                            |
| Hellfire Effect, Arcane Missiles' missiles (full share) | 2,740 of 2,810, 303 of 325                  | 0.99, 0.99                                                                                 |

### Channels

- Count: §Count of the 2026-11-19 to 2026-11-21 reports (the Wrack line)

Each line but the Corruption one takes landed ticks or hits on pairs of caster and mob with a ≥ 0.1, and a is the
caster's own share on the mob from control nukes of the school, else the share of every caster in the report, else the
mob's share over every live-realm report; Drain Life's RESIST count uses the mob's share alone. The tick lines take ticks of at least 10 damage before the resist. The Corruption line
needs no share: it sets each warlock's Drain Life ticks against that warlock's Corruption ticks on the same mob in the
same report, over ticks of any size. Expected counts are at a tenth of a and at a.

| Hits                         | With a partial resist (at a tenth, at the full share)                                                                                                                                                                  | Resisted share over the target's a |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Drain Life ticks             | 23 of 79 (25.3, 79); no full RESIST in 31 casts (13.0 if binary)                                                                                                                                                       | 0.087                              |
| Drain Life beside Corruption | on 1,462 sets of report, warlock and mob with both, 31 of 13,525 Drain Life ticks against 42.7 at the rate of the warlock's own Corruption, at least 170.9 in full; 31 of 13,496 against 42.7 with ticks of 10 or more | share free                         |
| Wrack ticks                  | 140 of 492 (build 71037)                                                                                                                                                                                               | 0.09                               |
| Starshards ticks             | 232 of 237 (64.2, 230); they take the Spell Piercing bonus at the full rate (EV-0092)                                                                                                                                  | 1.19                               |

Drain Life's and Wrack's ticks take a tenth and Starshards' the full share, and all three are channels, so being a
channel does not set a tick's share. As with the DoTs, no row field found tells the two groups apart, so the full share
is set per spell. Mind Flay (above) and Drain Soul are binary (Drain Soul 0 partials in 2,144 landed ticks, full RESIST
46 of 106 against 46.9), so their share does not matter.

## The 75% step

- Count: §Count of the 2026-11-19 to 2026-11-21 reports

Every partial resist on a landed hit with 50 or more damage before the resist lands on a 10% step up to 70% or on
exactly 75%; none lies between steps or above 75%.

| Step        | 10%   | 20%   | 30%   | 40%   | 50%   | 60%   | 70% | 75% |
| ----------- | ----- | ----- | ----- | ----- | ----- | ----- | --- | --- |
| Direct hits | 1,529 | 2,714 | 5,110 | 6,521 | 4,314 | 1,233 | 143 | 69  |
| Ticks       | 1,683 | 849   | 96    | 202   | 172   | 74    | 16  | 2   |

## Count of the 2026-11-19 to 2026-11-21 reports

- Sources: Warcraft Logs reports of Forever (zone 3000): every caster's damage in 120 random reports with a mage started
  2026-11-18 00:00 to 2026-11-19 21:00 UTC, for the reading before the switch, and the mages of 120 started from
  2026-11-21 00:00 UTC; the streams of the warlocks, priests, shamans and druids with 6 or more Spell Piercing on their
  gear, 178 in reports started 2026-11-19 23:00 UTC or later and 60 before; and 33 streams of 2026-11-20 reports of
  mages seen with piercing
- Game build: 1.60.1.71037 with hotfixes 114309, 114310, 114312, 114313 and 114314, live on the evening of 2026-11-19
  UTC; the logs place the switch of binary spells' Spell Piercing between 21:00 and 22:48 UTC. The hotfix a run was on
  and relogs are not recorded
- Date: 2026-11-19 23:00 to 2026-11-21 UTC
- Pull request: #1155

The resist counts run this file's Method again. Spell Piercing: WCL logs no negative resisted amount, so a bonus is a
normal hit with nothing resisted whose damage is 1.1, 1.2 or 1.3 times the damage before resists (within 2%). A mage
counts as piercing when 12% or more of at least 40 hits of its nukes that roll the partial (Arcane Explosion, Arcane
Missiles, Arcane Blast, Fire Blast, Blizzard, Fireball, Scorch, Pyroblast) take a bonus, a fifth of them or more at 20%;
65 mages. Talents and debuffs put a fixed +10% or +20% on whole spells, so the other classes' hits are read against the
median of the same caster's other hits on the same target and school within 10 seconds, which cancels such a modifier;
Stormstrike's +20% on Nature survives that on Earth Shock, which is read on its +10% bonuses alone.

Damage shields took no partial resist in 20,414 hits. The nukes with 50 or more hits on pairs with a ≥ 0.1 (Arcane
Blast, Arcane Explosion, Chain Lightning, Fire Blast, Fire Nova, Flame Shock, Flametongue Attack, Lightning Bolt, Mind
Blast, Moonfire, Pyroblast, Scorch, Searing Pain, Shadow Bolt, Soul Fire, Starfire, Touch of the Grave, Wrath) resist
0.94 to 1.09 of the share.

Scripts and tables: `forever-resist-hotfix-114309-1121` (RESULTS.md, `tables/`).

## Count of the 2026-11-22 combat log

- Sources: combat log of one Affliction warlock with 11 Spell Piercing, level 60 by its spell ranks, in Blackrock Spire
  (Firegut ogres, Ember Worgs); the source of the piercing is not recorded
- Game build: 1.60.1.71037 after hotfixes 114309, 114310, 114312, 114313 and 114314, two days after they went live; a
  relog is not recorded
- Date: 2026-11-21 23:18 to 23:32 in the player's local time, North America (2026-11-22 04:18 to 07:32 UTC)
- Pull request: #1155

Spell Piercing shows in the combat log as a negative resisted amount. The mobs took no partial resist from any of the
warlock's 858 ticks, so they have no Shadow resistance. Expected counts take the average share −11 / (11 + 8 × 60 + 19)
= −0.0216 and the triangle's 7.5 × 0.0216 = 16.2% chance of a bonus at the full share; a tenth of the piercing rounded
down to whole points is 1 point, 1.5% of ticks.

## Count of the 2026-11-19 to 2026-11-23 reports

- Sources: Warcraft Logs reports of Forever (zone 3000) started 2026-11-19 23:00 UTC or later, taken from the 2026-11-19
  to 2026-11-21 count's streams, the sweep index's streams and the players fetched for the Spell Piercing on their gear:
  2,937 shamans, mages, druids, priests and warlocks in 2,165 reports have hits counted here; the table of §The 1% cut
  rests on 900 casters other than mages in 806 reports, those with a level, their fight's gear and reference hits, once
  casters logged twice are dropped
- Game build: 1.60.1.71037 with the hotfixes up to 2026-11-23 (blizztrack lists 71037 from 2026-11-20 18:23 UTC; reports
  from 2026-11-19 23:00 may be on 70959); the hotfix a run was on and relogs are not recorded
- Date: 2026-11-19 23:00 to 2026-11-23 15:45 UTC
- Pull request: #1189

The Spell Piercing count runs as in §Count of the 2026-11-19 to 2026-11-21 reports, within 2.5% where that count took
2%: a bonus is a normal hit with nothing resisted, 50 or more damage before resists, at 1.1, 1.2 or 1.3 times the median of the same
caster's other same-school hits on the same mob within 10 seconds. Only mobs with no resistance in the school count:
every caster's plain direct hits of that school on the mob show 3% or fewer partial resists over 10 or more hits. Each
hit takes the gear piercing of its own fight from CombatantInfo, and the caster's level is its talent points plus 9.
Mages are counted apart by their Arcane Subtlety rank. A caster counts as having piercing when 5% or more of at least 30
of its reference hits take a bonus, 40 to 90% of them at +10% and at least one at +20%; its reference rate is that
share. Casters logged twice count once. Intervals are Wilson 95%.

Scripts and tables: `forever-piercing-rule-rows-1123` (`tables/edge.txt`, `tables/mage.txt`, `tables/tables.txt`,
`tables/resist.txt`).

## Count of the 2026-11-07 to 2026-11-24 reports

- Sources: Warcraft Logs reports of Forever (zone 3000) started 2026-11-23 15:45 to 2026-11-24 15:36 UTC: the sweep
  index's stream files of 701 reports (801 player streams) and 174 of the 267 streams fetched for the Spell Piercing on
  their gear. The Incinerate, Shadow Bolt, Death Coil, Shadowburn and Frost Shock lines pool them with the reports of
  §Count of the 2026-11-19 to 2026-11-23 reports, and the partial resist check of §Spell Piercing reads this count's
  streams and the other 93 fetched streams. §Rows with few hits takes reports of every date from 2026-11-07 22:31 UTC, the other 93 fetched
  streams among them, and Touch of Weakness and Frostbrand Attack only from 2026-11-19 23:00 UTC
- Game build: 1.60.1.71037, across the 2026-11-23 hotfix (EV-0134 puts it at about 2026-11-24 02:10 UTC); §Rows with
  few hits also 1.60.1.70862 from 2026-11-07,
  1.60.1.70888 from 2026-11-10, 1.60.1.70938 from 2026-11-13 and 1.60.1.70959 from 2026-11-17. The hotfix a run was on
  and relogs are not recorded
- Date: 2026-11-23 15:45 to 2026-11-24 15:12 UTC; §Rows with few hits from 2026-11-07 22:31 UTC
- Pull request: #1204

The count runs as in §Count of the 2026-11-19 to 2026-11-23 reports, with these changes:

- E comes from each caster's rate on the class's other nukes: Conflagrate, Hellfire Effect, Immolate's hit, Bane of
  Havoc, Searing Pain, Soul Fire and Shadow Bolt for warlocks; Lightning Bolt, Chain Lightning, Fire Nova, Lightning
  Shield, Flametongue Attack and Flame Shock's hit for shamans. A spell's own hits stay out of its E.
- A warlock's Shadow hits, direct or tick, from 0 to 12 s after the same warlock's Shadow Bolt crit on the same mob
  (Improved Shadow Bolt's window) are left out.
- Frost Shock is read against the median of the shaman's hits of every school on the same mob within 10 seconds.
- §Rows with few hits counts hits of 20 or more damage before resists. A hit outside every fight that logs gear takes
  the caster's most piercing in any fight of the report, and a caster without a level counts at level 60 when that gear
  holds 6 or more piercing, and is left out otherwise.
- P is Poisson on E plus the floor.

Scripts and tables: `forever-piercing-outliers-1124` (`tables/summary.txt`, `tables/recount_*_isb.txt`,
`tables/isb_*.txt`, `tables/shadowburn_checks.txt`, `tables/partials_new.txt`, `tables/rows_lowu.txt`,
`tables/needed.txt`, `tables/outlier_neighbours.txt`, `tables/ticks_abs.txt`).

## Count of the 2026-11-27 test

- Sources: combat logs of one level 60 Frost mage without Arcane Subtlety, EU, in Dire Maul, Blackrock Depths and the
  Stormwind Stockade, at 5 Spell Piercing on the character sheet (wrists off) and at 6 (wrists on, chest off); the
  StatLedger addon read the gear piercing at every gear change
- Game build: 1.60.1.71139 by StatLedger's login lines, on the latest hotfix push of 71139 at the time, relogged
- Date: 2026-11-27 07:59 to 09:27 UTC (08:59 to 10:27 in the player's local time, CET): 5 piercing from 07:59:05 to 08:35
  and from 09:20:48 to 09:24:46 UTC, 6 piercing from 09:24:48 to 09:26:56 UTC. StatLedger shows more time at 5 piercing
  (08:42 to 09:14 UTC) and at 6 (09:14:20 to 09:20:23 UTC) whose combat log was not provided
- Pull request: this branch's, not yet opened

The bonus is the negative resisted field of a SPELL_DAMAGE event in the raw combat log: −10, −20 or −30% of the hit
before resists, taken before the crit multiplier. Every SPELL_DAMAGE of the caster counts, and no hit took a partial
(positive) resist. Each target of an area spell counts as its own hit, since it rolls its own bonus and its own step: of
43 Blizzard waves on 2 or more targets, none gave every target a bonus, 15 gave some and 28 none, and one wave on 15
targets logged −8 and −4 on two of them and nothing on the other 13. The share is s = p / (p + 8 × 60 + 19).

| Hits                                 | 5 piercing (s 0.992%) | 6 piercing (s 1.188%) |
| ------------------------------------ | --------------------- | --------------------- |
| Arcane Explosion                     | 0 of 78               | 5 of 52               |
| Blizzard                             | 0 of 118              | 9 of 162              |
| Frost Nova                           | 0 of 23               | 3 of 15               |
| Cone of Cold                         | 0 of 19               | 0 of 2                |
| Frostbolt                            | 0 of 13               |                       |
| Arcane Missiles                      | 0 of 8                |                       |
| Ice Lance                            | 0 of 4                |                       |
| Fire Blast                           | 0 of 2                |                       |
| All                                  | 0 of 265              | 17 of 231 (7.4%)      |
| Expected at 7.5·s with no cut        | 19.7                  | 20.6                  |
| Expected at the 0.58% of §The 1% cut | 1.5                   | 1.3                   |

At 7.5·s with no cut, 0 of 265 has a chance of about 3·10⁻⁹. Blizzard's waves are hits of their own spell and take the
bonus at the full share: 9 of 162 at 6 piercing (14.4 at 7.5·s), where a tenth of the share, under the cut, gives none.

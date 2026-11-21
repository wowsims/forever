# EV-0142: Spell resistance on the live realms: binary spells, Spell Piercing, tick shares and the 75% step

- Claim: on the live realms a spell resists all or nothing when its damage row carries, beside the damage, a non-damage
  aura on the target, an interrupt, E_DUMMY, E_SCRIPT_EFFECT, E_ENERGIZE or E_CREATE_AREATRIGGER on any target; an
  A_DUMMY aura on the target does not count. Its hit roll resists it whole at about the target's average resist, and a
  landed hit or tick takes no partial resist. The Spell Piercing bonus is the negative side of the partial resist roll:
  a landed binary hit takes it at the full share, as do Insect Swarm's and Mind Flay's ticks, while Hurricane's,
  Entangling Roots' and Immolate's DoT ticks and a hit or tick at a tenth of the share take none, and a hit or tick at the
  full share takes it whole. Devouring Plague's and Pyroblast's ticks meet the target's full resistance. A channel's ticks meet a tenth like a DoT's, Starshards' excepted, which meet the full resistance; the
  missiles of Arcane Missiles are hits and meet it in full. No partial resist takes more than 75% of a hit, and damage shields take none.
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
  114122; the hotfix a run was on and relogs are not recorded. The items naming a count section have their own
- Date: 2026-09-27 to 2026-11-17; the items naming a count section have their own
- Pull request: #1080; the items naming a count section are #1155

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
caster, and E_ENERGIZE on Engulfing Shadows, not re-measured on build 71037 (below). E_NONE rows do not follow the
rule: Smokey's Lighter took no resist of any kind in 3,300 hits (0 of 179 with a partial and 0 of 179 RESIST against 68.8
at a mean share of 0.38), and Cloud of Chained Essence none in 4,246 applications and 15,460 ticks. E_TRIGGER_SPELL is
not tested.

## Spell Piercing

- Count: §Count of the 2026-11-19 to 2026-11-21 reports

The bonus goes where the partial resist roll goes, at the same share. E is the hits the caster's own rate on plain nukes
expects.

| Hits                                                  | With a piercing bonus (E), +10% : +20% |
| ----------------------------------------------------- | -------------------------------------- |
| Mages' nukes                                          | 23,244 of 82,767, 15,747 : 7,432       |
| Binary: Frostbolt                                     | 1,863 of 6,276 (1,802), 1,310 : 531    |
| Binary: Cone of Cold                                  | 1,111 of 4,675 (1,348), 738 : 366      |
| Binary: Frost Nova                                    | 418 of 1,781 (481), 287 : 131          |
| Binary: Flamestrike's cast                            | 994 of 3,235 (1,033), 666 : 327        |
| Binary: Arcane Wrath                                  | 898 of 3,385 (872), 638 : 257          |
| Binary: Ice Lance, Frostfire Bolt                     | 169 of 604 (181), 83 of 325 (91)       |
| Binary: Immolate's hit (33 warlocks)                  | 194 of 1,827 (201)                     |
| Binary: Incinerate (7 warlocks)                       | 71 of 898 (95)                         |
| Binary: Shadow Word: Death (10 priests)               | 19 of 315 (24)                         |
| Binary: Earth Shock, +10% bonuses only                | 42 (32.8)                              |
| Binary: Mind Flay ticks (10 priests)                  | 483 of 5,398 (388)                     |
| Binary: Insect Swarm ticks (13 druids)                | 1,089 of 9,240 (1,051)                 |
| Binary: Hurricane ticks                               | 19 of 7,577, all +10% (786)            |
| Binary: Entangling Roots ticks                        | 0 of 126 (12)                          |
| Immolate's DoT ticks                                  | 22 of 5,953 (618)                      |
| Tenth: Corruption, Siphon Life, Wrack ticks           | 0.06 to 0.08 of E                      |
| Tenth: Shadow Word: Pain, Flame Shock, Moonfire ticks | 0.01 to 0.04 of E                      |
| Tenth: Flamestrike's ground ticks, Rain of Fire       | 7 of 2,454 (672), 21 of 4,655 (608)    |
| Drain Life, Drain Soul ticks                          | 2 of 714                               |
| Full share: Pyroblast's DoT ticks                     | 342 of 1,084 (297), 240 : 100          |

Mages without piercing take bonuses on binary spells at their fixed-modifier rate (Frostbolt 4.1%, Cone of Cold 4.6%,
Frost Nova 1.2%, Arcane Wrath 0 of 189, Flamestrike's cast 0 of 1,236), so the bonuses above come from piercing.

Open: Incinerate reads low (P 0.007); Curse of Agony's ticks take 661 of 44,469 (0.14 of E) in a 2:1 mix, which may be a
tenth of the bonus; Frost Shock, Blast Wave, Bane of Agony and Shadowburn are not measured. E_ENERGIZE rests on Engulfing
Shadows alone, measured before build 71037 and not re-measured. Every piercing
source row (aura 123) states no class mask, so the source does not pick the spells. Under 10 piercing a tenth rounded
down to whole points also gives no bonus on a tick at a tenth. Shadow Bolt may be an exception: one caster's 44 hits
took none where the same caster's Hellfire Effect, Conflagrate and Bane of Havoc in the same reports took it, at P about
0.015 (4.2 expected).

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

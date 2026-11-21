# EV-0142: Spell resistance on the live realms: binary spells, Spell Piercing, tick shares and the 75% step

EV-0172 replaces the Spell Piercing rule for binary spells below: from the 2026-11-20 hotfixes a landed binary hit, and
Insect Swarm's and Mind Flay's ticks, take the bonus at the full share.

- Claim: on the live realms a spell resists all or nothing when its damage row carries, beside the damage, a non-damage
  aura on the target, an interrupt, E_DUMMY, E_SCRIPT_EFFECT, E_ENERGIZE or E_CREATE_AREATRIGGER on any target; an
  A_DUMMY aura on the target does not count. Its hit roll resists it whole at about the target's average resist, and a
  landed hit or tick takes no partial resist. The Spell Piercing bonus is the negative side of the partial resist roll:
  a binary spell takes none, nor does a hit or tick at a tenth of the share, while a hit or tick at the full share takes
  it whole. Devouring Plague's and Pyroblast's ticks meet the target's full resistance, and Flamestrike's ground ticks a
  tenth of it. A channel's ticks meet a tenth like a DoT's, Starshards' excepted, which meet the full resistance; the
  missiles of Arcane Missiles are hits and meet it in full. No partial resist takes more than 75% of a hit.
- Sources: 213 combat logs from 2026-09-27 to 2026-11-16, casters of every class; 17 beta-realm logs and 36 live-realm
  logs have hits on mobs with a resistance. The player streams of 992 Warcraft Logs reports of Forever (zone 3000) from
  2026-10-14 to 2026-11-16, of which 147 have hits on mobs with a resistance: 1 from the beta realms, 2 on 70862, 16 on
  70888, 86 on 70938 and 42 after push 114115; one report was checked line by line against a combat log of the same run.
  For the binary rule, 2,158 further caster streams from 706 live-realm reports (2026-11-06 to 2026-11-17) that the first
  pass never read, with the rules written down before they were fetched. For Spell Piercing, the combat logs and the
  report streams of 20 casters with piercing, on targets with no resistance in the spell's school. For the channels'
  tick shares, two studies whose predictions were written down first, both of 2026-11-17: the drain study
  (`forever-resist-drain-bombs-1117`, RESULTS.md) over the 1,322 live-realm reports with a control or test hit among the
  first pass's reports, the 706 held-out reports and 3 more, and the tick share study (`forever-resist-dispel-1117`, RESULTS.md) over the 706
  held-out reports and 866 more fetched for it
- Game build: beta realms to 2026-10-20; live realms from 2026-11-05 on 1.60.1.70862, 1.60.1.70888 from 2026-11-10 and
  1.60.1.70938 from 2026-11-13, the last reports after the 2026-11-16 hotfix push 114115 and possibly the 2026-11-17 push
  114122; the hotfix a run was on and relogs are not recorded
- Date: 2026-09-27 to 2026-11-17
- Pull request: #1080

## Method

A target's average resist a for a spell is its caster's own share on the same mob and school in the same log or report:
the damage resisted over the damage before the resist, on at least 3 hits of nukes whose rows deal damage and nothing
else (Wrath, Starfire, Moonfire, Lightning Bolt, Chain Lightning, Flame Shock, Shadow Bolt, Searing Pain, Soul Fire, Mind
Blast, Smite, Holy Fire, Fire Blast, Scorch, Fireball, Pyroblast, Arcane Blast, Arcane Explosion, Arcane Shot, Fire Nova,
Lightning Shield and Instant Poison). A pair of caster and mob counts from a = 0.1. In this first pass a binary spell's row deals damage
beside an aura or an interrupt on the target. Expected counts are those of the partial resist roll the sim used before
this change, at a for a hit and at a tenth of a for a DoT tick, and a full RESIST is expected at the rate a. Rows logged
twice, in overlapping combat logs or in two reports of one run, count once. Intervals are Wilson 95%.

The change sits between the beta and the live realms: the last combat logs with a partial resist on a landed binary
spell are from 2026-10-14, and from 2026-11-07 on no log has one. Builds 70888 and 70938 and push 114115 move none of
the counts below, and the share on the level 60 dungeon mobs holds across them (Chromatic Dragonspawn Nature 0.512,
0.498 and 0.512).

## A landed binary spell takes no partial resist

| Hits                                                                                            | Beta realms                      | Live realms, combat logs        | Live realms, reports                                             |
| ----------------------------------------------------------------------------------------------- | -------------------------------- | ------------------------------- | ---------------------------------------------------------------- |
| Binary direct hits (Earth Shock, Frostbolt, Frost Nova, Thunder Clap, Blast Wave, Frag Grenade) | 84 of 87 partial (86.7 expected) | 0 of 163 (156)                  | 0 of 235 (231): 0/70 on 70888, 0/106 on 70938, 0/59 after 114115 |
| Binary DoT ticks (Insect Swarm, Entangling Roots, Corrosive Poison, Bane of Agony)              | 70 of 84 (82.6)                  | 0 of 185 (183; 49.5 at a tenth) | 0 of 54                                                          |
| Mind Flay ticks                                                                                 | 68 of 78 (73.8)                  | 0 of 56                         | 0 of 267 (267)                                                   |
| Hurricane ticks                                                                                 | none cast                        | 0 of 39                         | 0 of 67; over every target 0 of 30,053                           |
| Other direct hits, for comparison                                                               | 816 of 849 (814)                 | 2,153 of 2,264 (2,215)          | 9,221 of 9,786 (9,627)                                           |

Full RESISTs of binary spells run at about the share: live combat logs 188 of 587 casts (212.1 expected), reports 381 of
986 (434.5); beta 108 of 282 (87.2).

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
| Wrack ticks (A_DUMMY on the target), shares of every caster on the mob    | 59 of 163 (62.2)                              |                            |

The plain nukes stay partial and never resist whole: Arcane Explosion 15,065 of 16,264 with a partial and RESIST 0 of
16,264, Searing Pain 2,177 of 2,208 and 0 of 2,208, Lightning Bolt 770 of 774 and 0 of 774, Shadow Bolt 196 of 200 and 0
of 200, Fireball 246 of 272 and 0 of 272; likewise Scorch, Soul Fire, Mind Blast, Chain Lightning, Arcane Missiles and
Fire Nova. In the first pass, Hurricane's ticks resisted whole at the share (logs 25 of 64 against 27.0, reports 64 of
131 against 68.8) with no partial on 30,053 landed ticks, Mind Flay 87 of 205 casts (99.3), Earth Shock and Frost Shock 133
of 355 (149.4) and Bane of Agony 10 of 37 (12.0). Gahz'rilla Fang's 3742 showed 0 of 22 landed with a partial and RESIST
10 of 32 (16.0).

Two effects rest on one spell each: E_CREATE_AREATRIGGER on Flamestrike's cast, which also carries a dummy aura on the
caster, and E_ENERGIZE on Engulfing Shadows, measured through Spell Piercing alone (below). E_NONE rows do not follow the
rule: Smokey's Lighter took no resist of any kind in 3,300 hits (0 of 179 with a partial and 0 of 179 RESIST against 68.8
at a mean share of 0.38), and Cloud of Chained Essence none in 4,246 applications and 15,460 ticks. E_TRIGGER_SPELL is
not tested.

## Spell Piercing

The bonus goes where the partial resist roll goes, at the same share. E is the hits the caster's own rate on plain nukes
in that log expects.

| Hits                                                 | With a piercing bonus (E)                            |
| ---------------------------------------------------- | ---------------------------------------------------- |
| Binary: Arcane Wrath, Frost Nova                     | 0 of 1,787 (539), 0 of 1,258 (361)                   |
| Binary: Cone of Cold, Frostbolt                      | 0 of 1,595 (420), 0 of 425 (112)                     |
| Binary: Flamestrike's cast                           | 0 of 1,145 (279)                                     |
| Binary: Engulfing Shadows (E_ENERGIZE)               | 0 of 252 (64)                                        |
| Binary: Blast Wave, Frostfire Bolt, Ice Lance        | 0 of 230 (42), 0 of 202 (72), 0 of 75 (19)           |
| Binary: Mind Flay ticks, at 8 piercing               | 0 of 320 (44)                                        |
| Tenth: Shadow Word: Pain ticks, at 8 piercing        | 0 of 515 (71 at the full share, 7.1 at a tenth)      |
| Tenth: Flamestrike's ground ticks                    | 0 of 4,177 (1,095 at the full share, 110 at a tenth) |
| Full share: Fire Blast                               | 439 of 1,686 (468)                                   |
| Full share: Blizzard's ticks                         | 543 of 2,123 (554)                                   |
| Full share: Pyroblast's and Devouring Plague's ticks | 56 of 280 (59), 23 of 215 (30)                       |

On the beta realms binary spells still took the bonus: one mage's Frost Nova on 87 of 173 hits from 2026-10-04 to
10-07, and 3 of 14 on 10-14. Every piercing source row (aura 123) states no class mask, so the source does not pick the
spells. Mind Flay's ticks took none in 1,738 at 8 piercing in November (EV-0092). Under 10 piercing a tenth rounded down
to whole points also gives no bonus on a tick at a tenth. Shadow Bolt is the exception: one caster's 44 hits took none
where the same caster's Hellfire Effect, Conflagrate and Bane of Havoc in the same reports took it, at P about 0.015 (4.2
expected).

## Tick shares

| Ticks                    | With a partial resist                                              | Resisted share over the target's a                         |
| ------------------------ | ------------------------------------------------------------------ | ---------------------------------------------------------- |
| Devouring Plague         | logs 45 of 45, reports 262 of 262; beta 8 of 8                     | 0.99 (0.98 on 70938, 1.02 after 114115)                    |
| Pyroblast's DoT          | logs 9 of 9, reports 20 of 20                                      | 1.09                                                       |
| Flamestrike ground ticks | reports 57 of 235, logs 12 of 109 (16.9 at a tenth)                | 0.09 on 70888 and on 70938; no RESIST in 235               |
| Other DoT ticks          | logs 435 of 1,495 (359 at a tenth), reports 1,522 of 4,464 (1,382) | Flame Shock 0.10 and 0.09, Shadow Word: Pain 0.11 and 0.08 |

### Channels

Each line but the Corruption one takes landed ticks or hits on pairs of caster and mob with a ≥ 0.1, and a is the
caster's own share on the mob from control nukes of the school, else the share of every caster in the report, else the
mob's share over every live-realm report; Drain Life's RESIST count uses the mob's share alone, and Arcane Missiles' line
the caster's own share alone. The tick lines take ticks of at least 10 damage before the resist. The Corruption line
needs no share: it sets each warlock's Drain Life ticks against that warlock's Corruption ticks on the same mob in the
same report, over ticks of any size. Expected counts are at a tenth of a and at a.

| Hits                                    | With a partial resist (at a tenth, at the full share)                                                                                                                                                                  | Resisted share over the target's a |
| --------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| Drain Life ticks                        | 23 of 79 (25.3, 79); no full RESIST in 31 casts (13.0 if binary)                                                                                                                                                       | 0.087                              |
| Drain Life beside Corruption            | on 1,462 sets of report, warlock and mob with both, 31 of 13,525 Drain Life ticks against 42.7 at the rate of the warlock's own Corruption, at least 170.9 in full; 31 of 13,496 against 42.7 with ticks of 10 or more | share free                         |
| Wrack ticks                             | 284 of 967 (336, 967)                                                                                                                                                                                                  | 0.082                              |
| Starshards ticks                        | 232 of 237 (64.2, 230); they take the Spell Piercing bonus at the full rate (EV-0092)                                                                                                                                  | 1.19                               |
| Arcane Missiles' missiles, own controls | 4,257 of 4,899 (1,166, 4,688); no full RESIST in 4,899                                                                                                                                                                 | 1.19                               |

Drain Life's and Wrack's ticks take a tenth and Starshards' the full share, and all three are channels, so being a
channel does not set a tick's share. As with the DoTs, no row field found tells the two groups apart, so the full share
is set per spell. Mind Flay (above) and Drain Soul are binary (Drain Soul 0 partials in 2,144 landed ticks, full RESIST
46 of 106 against 46.9), so their share does not matter.

## The 75% step

No hit on the live realms loses more than 75%. Hits that lost exactly 75%: 27 in the combat logs, and in the reports
3 on 70888, 48 on 70938 and 10 after push 114115, on seven mob families: Jade Ooze and Green Sludge, Molten War Golem,
Kirtonos, Phase Lasher, Morbent Fel and the Sunken Temple dragonkin. At 80% or more the triangle around a expects 16 hits
in the combat logs, which hold none, and 50 in the reports, which hold 18, all of them Instant Poison hits of about 35
damage on Thorn Eater Ghouls, where 75% rounds to between 76 and 79%. In one log of 2026-11-13, 16 of 28 hits on Jade
Oozes lost exactly 75% and the rest 60 or 70%; the triangle with its 80, 90 and 100% steps folded onto 75% fits them
better than the plain triangle (log-likelihood -27.2 against -31.7).

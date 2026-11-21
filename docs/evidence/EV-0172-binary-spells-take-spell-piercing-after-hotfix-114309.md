# EV-0172: Binary spells take the Spell Piercing bonus after hotfix 114309

- Claim: from the 2026-11-20 hotfixes on, a binary spell's landed hit takes the Spell Piercing bonus at the full share,
  as a nuke does, and Insect Swarm's and Mind Flay's ticks take it too. Hurricane's ticks, Immolate's DoT ticks and
  the ticks at a tenth of the share still take none, and a landed binary hit still takes no partial resist.
- Sources: Warcraft Logs reports of Forever (zone 3000). Every friendly DamageDone event of each caster class in 120
  random reports with a mage started 2026-11-18 00:00 to 2026-11-19 21:00 UTC and 120 started from 2026-11-21 00:00 UTC
  (the post side complete for mages); the streams of every warlock, priest, shaman and druid wearing 6 or more Spell
  Piercing in their gear in the post reports (48) and of 60 such casters in the pre reports; and 33 streams of 11-20
  reports of mages already seen with piercing, for the dating. Scripts and tables:
  `forever-resist-hotfix-114309-1121` (RESULTS.md)
- Game build: 1.60.1.71037 with hotfixes 114309, 114310, 114312, 114313 and 114314, dated 2026-11-20; in the logs the
  change starts between 2026-11-19 21:00 and 22:48 UTC. The hotfix a run was on and relogs are not recorded
- Date: 2026-11-18 to 2026-11-21
- Pull request: not recorded

## Method

Spell Piercing shows only as damage: WCL logs no negative resisted amount. A bonus is a normal hit with nothing resisted
whose damage is 1.1, 1.2 or 1.3 times the damage before resists (within 2%). A mage counts as piercing when 12% or more
of at least 40 hits of its nukes that roll the partial (Arcane Explosion, Arcane Missiles, Arcane Blast, Fire Blast,
Blizzard, Fireball, Scorch, Pyroblast) take a bonus, a fifth of them or more at 20%. The expected count for a spell is
its hits times its caster's own rate on those nukes. Talents and debuffs put a fixed +10% or +20% on whole spells of the
other classes, so their hits are compared with the median of the same caster's other hits on the same target and
school within 10 seconds, which cancels such a modifier; on that reading a binary spell scored 0 to 0.04 of the
expected rate before the hotfix. Resists use EV-0142's method: a from the caster's own nukes on the same mob and school,
a pair counted from a = 0.1.

## Spell Piercing, piercing casters

Ratio is the bonus count over the expected count.

| Spell | Before: bonus/hits (ratio) | After: bonus/hits (ratio) | After, +10% : +20% |
| --- | --- | --- | --- |
| Mages' nukes (43 and 55 mages) | 10,841/40,460 | 14,207/48,268 | 9,617 : 4,551 |
| Frostbolt | 137/2,632 (0.19, all +10%) | 1,330/3,813 (1.15) | 968 : 345 |
| Cone of Cold | 21/1,891 (0.05) | 798/3,067 (0.88) | 530 : 263 |
| Frost Nova | 2/843 (0.01) | 147/610 (0.89) | 101 : 46 |
| Flamestrike, the cast's hit | 2/2,124 (0.00) | 698/2,044 (1.01) | 463 : 234 |
| Arcane Wrath | 60/1,444 (0.16, all +10%) | 515/1,887 (1.03) | 363 : 152 |
| Ice Lance | 16/222 (0.27, all +10%) | 131/385 (1.07) | 104 : 26 |
| Frostfire Bolt | 0/53 (0.00) | 72/273 (0.95) | 42 : 29 |
| Immolate's hit (5 warlocks after) | 1/527 (0.02) | 45/421 (0.83) | 33 : 12 |
| Mind Flay ticks (3 priests after) | 3/1,812 (0.01) | 130/1,174 (1.52) | 101 : 29 |
| Insect Swarm ticks (3 druids after) | 0/2,251 (0.00) | 130/1,470 (0.94) | 85 : 45 |
| Hurricane ticks (the same druids) | 0/4,655 (expected 575) | 0/1,496 (expected 139) | |
| Immolate's DoT ticks | 3/1,557 (0.03) | 0/573 (expected 81) | |
| Corruption, Curse of Agony, Siphon Life, Wrack ticks | 0.03 to 0.07 | 0.05 to 0.08 | |
| Shadow Word: Pain, Flame Shock, Moonfire ticks | 0.00 to 0.01 | 0.00 to 0.04 | |
| Rain of Fire | 4/1,139 (0.04) | 2/897 (expected 186) | |
| Flamestrike's ground ticks | 12/2,980 (0.02) | 15/2,393 (0.02) | |
| Drain Life, Drain Soul ticks | 3/188 | 0/78 | |
| Pyroblast's DoT ticks (full share) | 48/253 (0.83) | 139/412 (1.15) | |

The bonuses before the hotfix on Frostbolt, Arcane Wrath and Ice Lance are all +10%: a fixed modifier on some casters.
Mages without piercing read the same before and after: Frostbolt 2.8% and 4.1%, Cone of Cold 2.4% and 4.6%, Frost Nova
0.1% and 1.2%, Arcane Wrath 0/43 and 0/189, Flamestrike's cast 0.7% and 0/1,236.

Not measured, under 25 hits from a piercing caster after the hotfix: Frost Shock, Blast Wave, Incinerate, Shadow Word:
Death (4/109, 0.45), Entangling Roots, Bane of Agony. Earth Shock keeps a talent's +20% that the relative reading does not
cancel.

## Resists

| | Before | After |
| --- | --- | --- |
| Landed binary hits and ticks with a partial resist, any target | 0 of 68,690 | 0 of 57,162 |
| The same on pairs with a ≥ 0.1, against the roll at a (binary ticks at a/10) | 0/1,029 (688.9 expected) | 0/166 (144.7 expected) |
| Binary full RESIST on those pairs | 159/408 (176.7 at a) | 61/203 (70.3 at a) |
| Non-binary direct hits on those pairs with a partial | 4,493/4,712 | 2,006/2,152 |
| Resisted over a, per nuke | 0.98 to 1.02 | 0.89 to 1.06 |
| Resisted over a, DoT ticks (Corruption, Flame Shock, Siphon Life, Wrack) | 0.08 to 0.09 | 0.09 to 0.12 |
| Resisted over a, Devouring Plague and Pyroblast's DoT ticks | 1.02, 1.02 | 0.93, 1.05 |

## When

+20% bonuses on binary Mage spells, every mage in the streams, per UTC hour on 2026-11-19: at most one an hour from 12:00
through the 20:00 hour (2,894 binary hits), no data in the 21:00 hour, the first two at 22:48 and 22:52, then 4, 10, 5,
18, 9, 14 and 23 an hour from 23:00 on. Each of the 33 streams of 11-20 takes the bonus from its first minutes.

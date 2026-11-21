# EV-0172: Binary spells take the Spell Piercing bonus after hotfix 114309

- Claim: from the hotfixes of 2026-11-19 evening UTC on, a binary spell's landed hit takes the Spell Piercing bonus at
  the full share, as a nuke does, and Insect Swarm's and Mind Flay's ticks take it too. Hurricane's ticks, Entangling
  Roots' ticks, Immolate's DoT ticks and the ticks at a tenth of the share still take none, and a landed binary hit
  still takes no partial resist.
- Sources: Warcraft Logs reports of Forever (zone 3000). Every friendly DamageDone event of each caster class in 120
  random reports with a mage started 2026-11-18 00:00 to 2026-11-19 21:00 UTC, and the mages of 120 started from
  2026-11-21 00:00 UTC; the streams of every warlock, priest, shaman and druid wearing 6 or more Spell Piercing in their
  gear in reports started 2026-11-19 23:00 UTC or later (178), and of 60 such casters in the earlier reports; and 33
  streams of 2026-11-20 reports of mages already seen with piercing. Scripts and tables:
  `forever-resist-hotfix-114309-1121` (RESULTS.md)
- Game build: 1.60.1.71037 with hotfixes 114309, 114310, 114312, 114313 and 114314, dated 2026-11-20 and live on the
  evening of 2026-11-19 UTC; in the logs the change starts between 21:00 and 22:48 UTC that evening. The hotfix a run
  was on and relogs are not recorded
- Date: 2026-11-18 to 2026-11-21
- Pull request: #1155

## Method

Spell Piercing shows only as damage: WCL logs no negative resisted amount. A bonus is a normal hit with nothing resisted
whose damage is 1.1, 1.2 or 1.3 times the damage before resists (within 2%). A mage counts as piercing when 12% or more
of at least 40 hits of its nukes that roll the partial (Arcane Explosion, Arcane Missiles, Arcane Blast, Fire Blast,
Blizzard, Fireball, Scorch, Pyroblast) take a bonus, a fifth of them or more at 20%. The expected count for a spell is
its hits times its caster's own rate on those nukes.

Talents and debuffs put a fixed +10% or +20% on whole spells, so for the other classes each hit is compared with the
median of the same caster's other hits on the same target and school within 10 seconds, which cancels such a modifier.
On that reading every binary spell scored 0 to 0.16 of the expected rate before the hotfix. Stormstrike's +20% on Nature
survives it on Earth Shock, so Earth Shock is read on its +10% bonuses alone against the share of +10% in its caster's
nukes. Resists use EV-0142's method: a from the caster's own nukes on the same mob and school, a pair counted from
a = 0.1, the partial expected from the roll at a (at a/10 for binary DoT and channel ticks).

## Spell Piercing, piercing casters

Ratio is the bonus count over the expected count. Mages: 43 casters before, 65 after.

| Spell                              | Before: bonus/hits (ratio) | After: bonus/hits (ratio) | After, +10% : +20% |
| ---------------------------------- | -------------------------- | ------------------------- | ------------------ |
| Mages' nukes                       | 10,841/40,460              | 23,244/82,767             | 15,747 : 7,432     |
| Frostbolt                          | 137/2,632 (0.19, all +10%) | 1,863/6,276 (1.03)        | 1,310 : 531        |
| Cone of Cold                       | 21/1,891 (0.05)            | 1,111/4,675 (0.82)        | 738 : 366          |
| Frost Nova                         | 2/843 (0.01)               | 418/1,781 (0.87)          | 287 : 131          |
| Flamestrike, the cast's hit        | 2/2,124 (0.00)             | 994/3,235 (0.96)          | 666 : 327          |
| Arcane Wrath                       | 60/1,444 (0.16, all +10%)  | 898/3,385 (1.03)          | 638 : 257          |
| Ice Lance                          | 16/222 (0.27, all +10%)    | 169/604 (0.93)            | 127 : 41           |
| Frostfire Bolt                     | 0/53 (0.00)                | 83/325 (0.91)             | 47 : 32            |
| Pyroblast's DoT ticks (full share) | 48/253 (0.83)              | 342/1,084 (1.15)          | 240 : 100          |

The bonuses before the hotfix on Frostbolt, Arcane Wrath and Ice Lance are all +10%: a fixed modifier on some casters.
Mages without piercing read the same before and after: Frostbolt 2.8% and 4.1%, Cone of Cold 2.4% and 4.6%, Frost Nova
0.1% and 1.2%, Arcane Wrath 0/43 and 0/189, Flamestrike's cast 0.7% and 0/1,236.

Read against each caster's own hits (casters after the hotfix in brackets):

| Spell                                          | Before: bonus/hits (ratio) | After: bonus/hits (ratio)               |
| ---------------------------------------------- | -------------------------- | --------------------------------------- |
| Immolate's hit                                 | 1/527 (0.02)               | 194/1,827 (0.96, 33)                    |
| Incinerate                                     | 1/220 (0.14)               | 71/898 (0.75, 7)                        |
| Shadow Word: Death                             | 0/106 (0.00)               | 19/315 (0.81, 10)                       |
| Earth Shock, +10% bonuses only                 | 6 against 15.3             | 42 against 32.8                         |
| Mind Flay ticks                                | 3/1,812 (0.01)             | 483/5,398 (1.25, 10)                    |
| Insect Swarm ticks                             | 0/2,251 (0.00)             | 1,089/9,240 (1.04, 13)                  |
| Hurricane ticks                                | 0/4,655 (575 expected)     | 19/7,577 (0.02, all +10%; 786 expected) |
| Entangling Roots ticks                         | 0/7                        | 0/126 (12 expected)                     |
| Immolate's DoT ticks                           | 3/1,557 (0.03)             | 22/5,953 (0.04)                         |
| Flamestrike's ground ticks (mages)             | 0/1,702 (0.00)             | 7/2,454 (0.01)                          |
| Corruption, Siphon Life, Wrack ticks           | 0.03 to 0.05               | 0.06 to 0.08                            |
| Curse of Agony ticks                           | 91/13,766 (0.07)           | 661/44,469 (0.14)                       |
| Shadow Word: Pain, Flame Shock, Moonfire ticks | 0.00 to 0.01               | 0.01 to 0.04                            |
| Rain of Fire                                   | 4/1,139 (0.04)             | 21/4,655 (0.03)                         |
| Drain Life, Drain Soul ticks                   | 3/188                      | 2/714                                   |

Incinerate reads low (P 0.007) but well clear of none. Curse of Agony's ticks rose from 0.07 to 0.14 of the full rate in
a 2:1 mix; whether they take a tenth of the bonus is open. Not measured, under 25 hits from a piercing caster after the
hotfix: Frost Shock, Blast Wave, Bane of Agony, Shadowburn.

## Resists

|                                                                                          | Before                   | After                    |
| ---------------------------------------------------------------------------------------- | ------------------------ | ------------------------ |
| Landed binary hits and ticks with a partial resist, any target                           | 0 of 68,690              | 0 of 113,686             |
| The same on pairs with a ≥ 0.1, against the roll at a (binary ticks at a/10)             | 0/1,029 (688.9 expected) | 0/1,072 (780.5 expected) |
| Binary full RESIST on those pairs (direct hits)                                          | 159/408 (176.7 at a)     | 203/617 (226.1 at a)     |
| Hurricane full RESIST on those pairs                                                     | 254/530 (237.7 at a)     | 178/401 (183.9 at a)     |
| Non-binary direct hits on those pairs with a partial                                     | 4,493/4,712              | 6,962/7,382              |
| Resisted over a, per nuke with 50+ hits                                                  | 0.95 to 1.01             | 0.94 to 1.09             |
| Resisted over a, DoT ticks (Corruption, Flame Shock, Siphon Life, Wrack, SW:P, Moonfire) | 0.08 to 0.11             | 0.09 to 0.12             |
| Resisted over a, Devouring Plague and Pyroblast's DoT ticks                              | 1.02, 1.02               | 0.95, 0.98               |
| Resisted over a, Flamestrike's ground ticks and Rain of Fire                             | 0.13, 0.12               | 0.10, 0.12               |

## When

+20% bonuses on binary Mage spells, every mage in the streams, per UTC hour on 2026-11-19: at most one an hour from 12:00
through the 20:00 hour (2,894 binary hits), no data in the 21:00 hour, the first two at 22:48 and 22:52, then 4, 10, 5,
18, 9, 14 and 23 an hour from 23:00 on. Each of the 33 streams of 11-20 takes the bonus from its first minutes.

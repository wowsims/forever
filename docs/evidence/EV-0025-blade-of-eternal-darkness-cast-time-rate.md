# EV-0025: Blade of Eternal Darkness procs on cast time

- Claim: Blade of Eternal Darkness 17780 procs 1.8 times a minute on the cast time of the cast that launched each landed
  harmful spell, at least 1.5 s, whether or not the spell deals damage; damage over time ticks, Ignite and wand shots
  never roll.
- Sources: Warcraft Logs Nq3C7bVGnvmyh2Pw, cvRTFb1qhGBKNkCD and JK1z96cCrj28tpBV, Forever dungeon reports of one mage;
  Pjd9MWAXgCYTcn7Z fights 244, 245, 247 and 248 and VRrgQJFbdXckaZq7 fights 1 and 2, of two warlocks;
  NTJzVdMZBQDaqG83 fight 2, of a priest; the combat log of the warlock in VRrgQJFbdXckaZq7, Maraudon, 2026-11-17, level
  not recorded
- Game build: the mage reports not recorded; the warlock and priest reports 1.60.1.70938 by report date; the combat log
  1.60.1, hotfix not recorded
- Date: 2026-11-14 for the mage reports and Pjd9MWAXgCYTcn7Z; 2026-11-17 for VRrgQJFbdXckaZq7, NTJzVdMZBQDaqG83 and the
  combat log
- Pull request: #1007; the harmful spells that deal no damage, #1097

On the three mage reports the blade is confirmed in the main hand by combatant info in 12 fights; the other fights with
procs are inferred from those procs. There are 554 procs, each tied to the event at its timestamp.

| Trigger                                                                                                   | Procs |
| --------------------------------------------------------------------------------------------------------- | ----: |
| AoE casts (Arcane Explosion 192, Flamestrike first hit 45, Frost Nova 11, Cone of Cold 6)                 |   256 |
| Single-target casts (Arcane Wrath 45, Fire Blast 34, Arcane Blast 32, Fireball 18, Pyroblast 5, Scorch 5) |   145 |
| Flamestrike's periodic spell 1279990                                                                      |    75 |
| Arcane Missiles missiles                                                                                  |    58 |
| Counterspell, Arcane Missiles channel starts, mixed                                                       |    20 |
| DoT ticks (0 of 488), Ignite (0 of 377), wand Shoot (0 of 132)                                            |     0 |

An area spell rolls once for each target. On short-tier AoE the per-landing rate rises with the targets hit: 3.8% at 1
target, 14.0% at 3, 20.0% at 4, 44.3% at 9 or more. Rolling 4.7% per target predicts 4.7%, 13.4%, 17.5% and 44.8%, where
one roll per landing predicts a flat 17.5%. Log-likelihood is -1420 per target against -1578 per landing.

The chance follows the cast time. Single-target landings proc 4.4% on instants and triggered spells (93 of 2119), 3.4% on 1.5 s
casts (29 of 842) and 9.3% on casts of about 2.95 s (60 of 644). Pyroblast cut to 1.48 s by Heating Up procs 2.1% (3 of
140). Arcane Wrath with Presence of Mind procs 4.0% (2 of 50), against 9.0% hardcast. The fit is 1.80 per minute
(1.65 to 1.96). Against one flat chance per target the likelihood-ratio statistic is 35.2 (p about 3e-9). The blade's
weapon speed is 1.5 s, so a manager on the main hand's speed would price every hit at 4.5%; the 3 s tier rules that out.
Fitted per tier, the rate is 4.5% (4.1 to 5.0) on instants, triggered spells and 1.5 s casts and 9.1% (7.5 to 10.9) on 3 s
casts.

Harmful spells that deal no damage also rolled in game (Counterspell 5 of 92, Arcane Missiles channel starts 6 of 219),
11 of the 554 procs; the section below has the other classes' casts. Misses are undetermined, since the sample has 2 true
misses or full resists, so only landed spells roll.

The cast-time fit rests on one player and mage spells only. Only two cast-time levels exist (about 1.5 s and 2.95 s), so a two-tier rule
fits as well as the proportional one. Haste is unmeasured. One low-level fight (Scarlet Monastery Library) has 2 procs
against 7.5 expected.

The mana each proc restores, a full resist included, is the reading of #1001 on the same reports (EV-0108).

## Harmful spells that deal no damage

Each proc in the warlock and priest reports is credited to the wearer's own event logged 0 to 45 ms before the proc's
mana gain, a damage event first where there is one. Most of the warlocks' procs come off a curse or a damage over time
effect landing with no direct damage.

| Trigger                      | Pjd9MWAXgCYTcn7Z | VRrgQJFbdXckaZq7 | NTJzVdMZBQDaqG83 |
| ---------------------------- | ---------------: | ---------------: | ---------------: |
| Corruption cast              |               11 |                  |                  |
| Bane of Agony cast           |                7 |                  |                  |
| Wrack cast                   |                5 |                2 |                  |
| Curse of Recklessness cast   |                4 |                  |                  |
| Siphon Life cast             |                2 |                  |                  |
| Curse of the Elements cast   |                2 |                  |                  |
| Rain of Fire cast            |                1 |                  |                  |
| Shadow Bolt damage           |                2 |                2 |                  |
| Immolate damage              |                1 |                1 |                  |
| Holy Nova damage             |                  |                  |                5 |
| Mind Vision cast on an enemy |                  |                  |                3 |
| Shadow Word: Pain cast       |                  |                  |                1 |
| Procs                        |               35 |                5 |                9 |

The combat log of the warlock in VRrgQJFbdXckaZq7 shows both Wrack procs, at 11:47:28.977 and 11:51:09.444, logged 16 ms
after the SPELL_CAST_SUCCESS and SPELL_AURA_APPLIED of Wrack 1316697 on the target, with no damage event between.

On the two warlock reports a landing is a cast of one of the spells below followed by its debuff applied or refreshed on
the target within 300 ms, so resists, immunities and evades are left out; it procs when the mana gain is logged 0 to
45 ms after the debuff. Every one of these casts is instant; Corruption's start and end are logged at most 1 ms
apart.

| Spell                 | Landings | Procs | Rate |
| --------------------- | -------: | ----: | ---: |
| Wrack                 |      280 |     7 | 2.5% |
| Corruption            |      268 |    12 | 4.5% |
| Bane of Agony         |      221 |     7 | 3.2% |
| Curse of Recklessness |       81 |     4 | 4.9% |
| Curse of the Elements |       58 |     2 | 3.4% |
| Siphon Life           |       55 |     2 | 3.6% |
| All                   |      963 |    34 | 3.5% |

The 34 procs in 963 landings, 3.5% (2.5 to 4.9), sit 1.5 standard deviations under the 4.5% that 1.8 per minute gives an
instant; the sample cannot tell the two apart. One proc counts for Corruption here and for Immolate's damage in the table
above, since both landed in the same millisecond.

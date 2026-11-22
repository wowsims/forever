# EV-0181: Magma Totem's pulse at its end and a level 60 Dwarf shaman's base attributes

- Claim: a Magma Totem that stands its full duration also pulses at the moment it ends, so a 40 s totem pulses 20 times
  (2 s to 40 s) and a 20 s one 10 times; a level 60 Dwarf shaman's base attributes are Strength 90, Agility 48, Stamina
  98, Intellect 86 and Spirit 99.
- Sources: combat logs of a level 60 Dwarf Enhancement shaman (Blackrock Spire and Dire Maul on 2026-11-11, Blackrock
  Spire on 2026-11-19 and 2026-11-20, Alterac Valley on 2026-11-21), of an Enhancement shaman at levels 41 and 42
  (Maraudon, 2026-11-15) and of an Elemental shaman at levels 45 and 46 (2026-11-16 and 2026-11-17); the level 60 Dwarf
  shaman's two StatLedger files, 2026-11-11 and 2026-11-15 to 2026-11-22
- Game build: the logs by date: 1.60.1.70888 (2026-11-11), 1.60.1.70938 (2026-11-15 to 2026-11-17), 1.60.1.70959
  (2026-11-19, hotfix pushes up to 114122, relogged) and 1.60.1.71037 (2026-11-21); the ledgers as each login states, 1.60.1.70888,
  1.60.1.70938, 1.60.1.70959 and 1.60.1.71037. The hotfix of the other logs is not recorded, and whether their players
  relogged after a hotfix is not known
- Date: 2026-11-11 to 2026-11-22
- Pull request: #1175

## Magma Totem's last pulse

Each Magma Totem drop is read from the summon: the offset of every pulse, the drop's full duration (20 s for the rank
alone, 40 s with Trial by Fire), and whether a Totemic Call, a Totemic Recall or another totem of the shaman's fell inside
it. A drop counts when no other fire totem and no Totemic Recall took it down before its end. Where a drop's last pulse
is 2 s before its end, the targets of that pulse are checked for a death before the end.

| Full duration | Inside the drop        | Pulse at the end | Last pulse 2 s early, a death of each target before the end | Last pulse 2 s early, no death of a target before the end |
| ------------- | ---------------------- | ---------------: | ----------------------------------------------------------: | --------------------------------------------------------: |
| 20 s          | nothing                |               14 |                                                           2 |                                                         0 |
| 20 s          | another totem, no Call |                6 |                                                           2 |                                                         0 |
| 40 s          | a Totemic Call         |               29 |                                                           0 |                                                         0 |
| 40 s          | nothing                |               11 |                                                           2 |                                                         4 |
| 40 s          | another totem, no Call |                1 |                                                           1 |                                                         0 |

Every 20 s drop pulsed at its end unless each target of its last pulse died before the end, 20 of 20, and every 40 s
drop with a Totemic Call inside pulsed at its end, 29 of 29. Of the 40 s drops without one, leaving out those whose
targets all died before the end, 12 of 16 pulsed at the end. The logs do not show why the other 4 did not: two of their
targets have no death line in the log, one dies 0.73 s after the end and one 21.55 s after it. A target leaving the
pulse's reach would explain them, which the logs cannot confirm.

The pulses at the end land 39.989 to 40.017 s after the summon, and 19.927 to 20.017 s for the 20 s drops. On the
2026-11-19 log the first pulse lands 2.0053 s [2.0041, 2.0064] after the summon (125 drops) and each next one 1.9999 s
[1.9995, 2.0004] after the last (1,205 gaps under 2.5 s; the 45 longer gaps had no target in reach).

## Base attributes

Each ledger line gives each attribute as total, effective, positive and negative, so the base is the total less the
positive and the negative. Read on every line's sheet, carried forward from the last change and reset at each login and
reload, the base is Strength 90, Agility 48, Stamina 98, Intellect 86 and Spirit 99 on all 11,756 lines that hold the
five attributes, the 25 lines whose sheet reads every stat as 0 left out.

| Ledger                   | Build        | Lines |
| ------------------------ | ------------ | ----: |
| 2026-11-11               | 1.60.1.70888 |   693 |
| 2026-11-15 to 2026-11-22 | 1.60.1.70938 | 4,176 |
| 2026-11-15 to 2026-11-22 | 1.60.1.70959 | 5,139 |
| 2026-11-15 to 2026-11-22 | 1.60.1.71037 | 1,748 |

The 5,805 out-of-combat sheets of the second ledger with a head, chest and main hand worn and the talents' melee hit
were rebuilt from the items worn, with no buffs. With the base above, 57 states (409 sheets) match the rebuild in all five
attributes, attack power, spell crit, spell power, mana and armor. On those, dodge differs by 0.04% for each point of
defense below 300, and melee crit by 0.04% for each point of two-hand axe skill below 300 or by Elemental Devastation's
crit. Health matches on 171 of the 409 and reads 300 or 616 higher on the rest, with Stamina matching on all of them. A
sheet built on Strength 87, Agility 51, Stamina 98, Intellect 89 and Spirit 99 matches none.

These sheets are of one Dwarf shaman. No level 60 sheet of another shaman race or of another Dwarf class has been read,
so the sheets do not show how these attributes split between the shaman's class attributes and the Dwarf's race offset.

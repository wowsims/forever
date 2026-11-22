# EV-0180: Clearcasting kept by Lightning Bolt and Chain Lightning at five Maelstrom Weapon stacks

- Claim: a Lightning Bolt that consumes five Maelstrom Weapon stacks keeps Elemental Focus's Clearcasting charge, and a
  Chain Lightning cast while Maelstrom Ready! (408501) is up costs nothing under Clearcasting but keeps the charge and
  takes off Maelstrom Ready! instead; Maelstrom Ready! goes up as the stacks reach five and comes off at the next
  Lightning Bolt, Chain Lightning, Healing Wave or Lesser Healing Wave, or with the stacks.
- Sources: combat logs of one level 60 Dwarf Enhancement shaman with Elemental Focus and Maelstrom Weapon 1/5: Lower
  and Upper Blackrock Spire on 2026-11-19, Alterac Valley on 2026-11-21. Maelstrom Ready!'s gain and removal also rest
  on three earlier combat logs: a level 60 Dwarf Enhancement shaman in Stratholme on 2026-11-13, and a level 44 to 49
  Enhancement shaman on the night of 2026-11-13 to 2026-11-14 and in Maraudon on 2026-11-15
- Game build: 1.60.1.70959 on 2026-11-19 and 1.60.1.71037 on 2026-11-21, hotfixes up to push 114122; relog not known.
  The earlier logs: 1.60.1 on 2026-11-13 and 2026-11-14, full build not recorded; 1.60.1.70938 on 2026-11-15
- Date: 2026-11-19 and 2026-11-21; the earlier logs 2026-11-13 to 2026-11-15
- Pull request: not opened yet

Each Lightning Bolt and Chain Lightning cast is read with the Maelstrom Weapon stacks, Maelstrom Ready! and
Clearcasting as they stood before the cast, the cost field of its cast line, the mana it paid where the shaman's power
fields allow, and the auras removed in its ms. A Lightning Bolt at five stacks is instant and free from Maelstrom
Weapon alone. The Alterac Valley log holds no Lightning Bolt or Chain Lightning cast with Clearcasting up.

Casts with Clearcasting up, 2026-11-19:

| Cast                                               | Clearcasting kept | Clearcasting spent |
| -------------------------------------------------- | ----------------: | -----------------: |
| Lightning Bolt, five stacks, Maelstrom Ready! up   |                 5 |                  0 |
| Lightning Bolt, five stacks, Maelstrom Ready! gone |                 2 |                  0 |
| Lightning Bolt, four stacks                        |                 0 |                  2 |
| Lightning Bolt, three stacks                       |                 1 |                  0 |
| Chain Lightning, five stacks, Maelstrom Ready! up  |                 5 |                  0 |
| Chain Lightning, zero to four stacks               |                 2 |                 35 |

The Chain Lightning casts at five stacks with Clearcasting up logged a cost of 0 and paid none, and each took off
Maelstrom Ready! with the stacks left in place. Without Clearcasting, 13 Chain Lightning casts with Maelstrom Ready! up
logged the full 495 (median paid 493 where readable, n 8) and took off Maelstrom Ready! alone. The Lightning Bolt at
three stacks that kept the charge, against both at four stacks spending it, is not explained.

Maelstrom Ready! comes off with every Lightning Bolt at five stacks, together with the stacks: 99 casts without
Clearcasting and 5 with it on 2026-11-19, 12 on 2026-11-21.

Earlier logs read where it goes up and what else takes it off. A level 60 Dwarf Enhancement shaman with Maelstrom
Weapon 5/5, Stratholme, 2026-11-13: it went up exactly at five stacks 179 of 179 times. A level 44 to 45 Enhancement
shaman with Maelstrom Weapon 5/5, the night of 2026-11-13 to 2026-11-14: up at five stacks 168 of 169 times; Healing
Wave and Lesser Healing Wave cast at five stacks took it off 9 of 9 times with the stacks left in place. The same shaman
at levels 48 to 49 in Maraudon, 2026-11-15 (1.60.1.70938): up at five stacks 122 of 123 times; taken off by Lightning
Bolt 92 times, Healing Wave 8 and Lesser Healing Wave 3, with the stacks left after the heals, and by the stacks running
out 20 times.

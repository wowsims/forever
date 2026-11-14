# EV-0104: Item tooltips that the database rules reproduce

- Claim: the in-game tooltip shows one bonus where an item's stats and its equip spell both state it, and the rules the database reads reproduce the tooltips of the weapons and items listed.
- Sources: in-game item tooltips; no log
- Game build: not recorded
- Date: not recorded
- Pull request: #974, #302, #4

Each row is a tooltip compared against the value the database computes.

| Reading                                                                                                       | Tooltip                                                                             | Pull request                   |
| ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ------------------------------ |
| Bloody Claws of Attack 280068 states +1 Fist Weapons as item stat 98 and again as equip spell 1297442         | one +1                                                                              | #974                           |
| Seal of the Dawn 13209 states +78 Attack Power against Undead as item stat 128 and again as equip spell 23930 | one +78                                                                             | #974's rule, no count recorded |
| Caster weapons flagged for implicit spell power or healing power                                              | all 150 changed weapons match the in-game tooltip exactly, healing and spell damage | #302                           |
| ItemModType 124 "All Resistances": 22197 Heavy Obsidian Belt and 22198 Jagged Obsidian Shield                 | "+5 All Resistances"                                                                | #4                             |
| ItemModType 124: 12252 Staff of Protection                                                                    | "+6 Arcane/Fire/Nature/Frost/Shadow"                                                | #4                             |

Three independent tooltips confirm the expansion of "All Resistances" in the last two rows.

# EV-0113: ItemModType 124 is all five resistances in the item tooltips

- Claim: item stat 124 ("All Resistances") raises the five resistances at once.
- Sources: three in-game item tooltips; no log
- Game build: not recorded
- Date: not recorded
- Pull request: #4; the reading is in its commit a586234102

| Item                                                       | Tooltip                              |
| ---------------------------------------------------------- | ------------------------------------ |
| 22197 Heavy Obsidian Belt and 22198 Jagged Obsidian Shield | "+5 All Resistances"                 |
| 12252 Staff of Protection                                  | "+6 Arcane/Fire/Nature/Frost/Shadow" |

The commit names these three independent tooltips as what identified index 124.

# EV-0032: Rip Rank 2 and Ferocious Bite Rank 1 shares and Enrage's armor, druid levels 26 to 33

- Claim: Rip Rank 2 takes 0.0114 × attack power × min(combo points, 4) a tick, Ferocious Bite Rank 1 about 0.037 per combo point, and Enrage in Bear Form takes 27% of 2.8 times the item and bonus armor plus the armor from Agility.
- Sources: combat log and character sheets, a Tauren druid at levels 26 to 33
- Game build: 1.60.1.70862
- Date: 2026-11-08 to 2026-11-09
- Pull request: #933

| Reading                                                   | Measured                                                                                 |
| --------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Rip Rank 2 (9492)                                         | 0.0114 × attack power × min(combo points, 4) per tick, 139 of 139 ticks                  |
| Ferocious Bite Rank 1 (1328618)                           | k in [0.03683, 0.03707), 157 of 157 clean landed bites; 0.0375 fits 152, 0.0360 fits 147 |
| Ferocious Bite Ranks 2 and 3, same build, levels 37 to 41 | [0.0367, 0.0376] over 24 bites                                                           |
| Rake ticks read the attack power at the tick              | 334 of 337 at the live attack power, 325 of 340 at the cast's                            |
| Blood Frenzy 2/2 in Bear Form                             | +5 rage on every non-periodic crit, once per Swipe target, 90 of 90                      |

For Enrage, the character sheet reads 60 bonus armor with one weapon and 100 after a cloak and trinket swap, taken from the jump at a shift (1.8 × bonus armor). The items' own bonus armor reads 45 and 76, so how the item's 45 becomes 60 is open. The game takes 0.27 × (2.8 × (item armor + bonus armor) + 2 × Agility), with buff armor untouched and no level term: 40 of 40 Enrage pairs on the sheet. The same rule explains the earlier "42 to 48 more from level 29" (147 and 179 armor) on the levels 37 to 41 log, which was a bonus armor item. Dire Bear Form's 16% was unmeasured here.

The Enrage sheet readings of an earlier pass at levels 17 to 22 are EV-0109.

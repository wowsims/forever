# EV-0114: Innervate multiplies Spirit regen and MP5 but not a periodic energize tick

- Claim: Innervate 29166 makes the regen five times the Spirit regen and MP5 together, while Blessing of Wisdom 19853
  ticks the same 30 mana under it as outside it.
- Sources: combat log and StatLedger addon ledger, a Tauren Balance druid, levels 45 to 47, partly grouped; the same
  druid's 2026-11-12 ledger for the pooled lines with MP5 up
- Game build: 1.60.1.70938, hotfix 114060
- Date: 2026-11-13, 07:55 to 19:59
- Pull request: #1018

The druid cast Innervate on himself. Blessing of Wisdom's ticks are its periodic energize lines on him, amount plus
overenergize, split by whether his Innervate was up. Innervate's regen is read from the ledger's regen fields on every
line inside an Innervate (0.5 s inside each edge) whose regen field was rewritten after the aura landed; the ledger
carries regen forward and rewrites it only on some events, so a line before the first rewrite holds the value from before
Innervate and is left out.

| Reading                                                       | Count                                                 |
| ------------------------------------------------------------- | ----------------------------------------------------- |
| Blessing of Wisdom ticks under Innervate at 30 mana           | 12 of 12                                              |
| Blessing of Wisdom ticks outside Innervate at 30 mana         | 1,237; the other 10 logged 9 to 29                    |
| Ledger lines with regen = 5 × (Spirit regen + MP5)            | 31 of 31, over the 18 of 30 Innervates with a rewrite |
| Of those, lines with MP5 up (Warchief's Blessing, 2 mana a s) | 2; 5 × Spirit regen + MP5 fits neither                |
| Lines with MP5 up, pooled with the 2026-11-12 ledger          | 10 of 10                                              |
| Ledger lines with the casting regen equal to the full regen   | 31 of 31                                              |

Mana Spring 10491 ticked 33 times on him, never under Innervate; the sim leaves it unscaled too, since it is a periodic
energize aura like Blessing of Wisdom.

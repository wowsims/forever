# EV-0097: Caster weapon damage ranges fit fixed shares of the damage table

- Claim: a weapon flagged REDUCED_CASTER_DPS, or carrying a random suffix whose bonus list has an ITEM_BONUS_REDUCED_CASTER_DPS row, deals 2/3 (one-hand) or 29/39 (two-hand) of its item level damage table in game, while the game files' caster tables equal the regular ones.
- Sources: in-game tooltip damage ranges of the flagged weapons, read against the generator's output; the tooltip is the measurement, so this file records a fit to tooltip ranges
- Game build: 1.60.1.70245 (the local database at hotfix push 112486)
- Date: not recorded
- Pull request: #298, with the suffix reading of #303

ItemSparse Flags[4] bit 0x100 marks 165 weapons (81 one-hand, 84 two-hand), all of them caster weapons. Under the
generator's min and max rounding, the shares reproduce the in-game damage range of all 165, including the four with a
negative QualityModifier (Benediction and three AQ20 maces and swords). 29/39 of the two-hand table is 29/30 of the
one-hand one. QualityModifier stays an additive per-second offset on top. Unflagged weapons are unaffected. Of the flagged
weapons, the 160 the sim ships all match the in-game ranges after the change.

| Item                          | Before the shares | In game and after |
| ----------------------------- | ----------------- | ----------------- |
| Death Speaker Scepter         | 42 to 78          | 28 to 52          |
| Simone's Cultivating Hammer   | 53 to 100         | 35 to 67          |
| Hammer of the Gathering Storm | 62 to 137         | 33 to 83          |
| Benediction                   | 134 to 222        | 89 to 155         |

The pull request left one check open: Death Speaker Scepter at 28 to 52 with 2.80 speed in game.

A bonus list with an ITEM_BONUS_REDUCED_CASTER_DPS row marks its suffix, and a weapon offering one carries the damage it
deals with it: Brutal War Axe (3210) shows 43 to 66 with "of the Physician" (12911) in game against 58 to 88 on its own.
A weapon offering none, and the armor the lower lists sit on, carry no such range.

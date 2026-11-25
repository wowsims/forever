# EV-0192: What Death Mark stores

- Claim: Death Mark stores the rogue's physical damage, Garrote's and Rupture's at twice the share, and none of the
  poisons', Fiery Weapon's or Touch of the Grave's.
- Sources: combat log `WoWCombatLog-112426_191754`, one level 60 rogue, three Scholomance runs
- Game build: 1.60.1.71091; the hotfix push and whether the player relogged after it not recorded
- Date: 2026-11-24 19:17 to 23:25 UTC−5 (2026-11-25 00:17 to 04:25 UTC)
- Pull request: not recorded

The rogue is a Forsaken wielding a sword in the main hand and an axe in the off hand, both with Fiery Weapon, Instant
Poison on the main hand and Deadly Poison on the off hand. The talent set is read from the casts each set alone has:
the first run is combat with Hack and Slash 5/5, the second subtlety with Death Mark and Coup de Grace 3/3, and the
third combat with Hack and Slash 1/5 until it turns subtlety again.

## Death Mark

Each mark's payout, the one Death Mark tick, is set against the damage the rogue dealt the marked target from the
mark's landing to that tick, the amount with its absorbed part, split into physical (white swings and specials),
bleeds (Garrote and Rupture) and magic (Instant Poison, Deadly Poison, Fiery Weapon and Touch of the Grave). The stored
amount is 12% of that damage per combo point times Coup de Grace's 1.2, with each kind weighted. Of 67 marks with a
payout, the 43 whose payout sits below 0.9 times the cap before crit are fitted, so the cap cannot bind.

| Weights (physical, magic, bleeds) | Payout / stored, median | Interquartile range |
| --------------------------------- | ----------------------- | ------------------- |
| 1, 0, 2                           | 1.000                   | 0.868 to 1.001      |
| 1, 1, 2                           | 0.818                   | 0.717 to 0.886      |
| 1, 1, 1                           | 0.968                   | 0.799 to 1.169      |
| 1, 0, 1                           | 1.237                   | 1.005 to 1.414      |

A free least-squares fit reads physical ×0.818, magic ×0.067 and bleeds ×2.153, with a median residual of 6.5%. With
the weights 1, 0, 2, 16 of the 43 payouts sit within 1% of the stored amount; with 1, 1, 2 and with 1, 1, 1 none does,
and with 1, 0, 1 two do. The payout's base amount equals its amount (median 0.999), so it takes no modifier of its own.
Under 1, 0, 2, 19 payouts read below 0.95 of the stored amount and 4 above 1.05; what leaves those low is not known.

# R-0025: Elixirs stack across groups but only the strongest elixir crit applies

- Skill: wowsims-consumables-items
- Code: `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.Exclusive(elixirCritCategory, false)` (the category `"ElixirCrit"`)
- Pull request: #209
- Evidence: pending

Elixirs stack their stats across groups, but only the strongest elixir crit applies.

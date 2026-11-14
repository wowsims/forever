# R-0029: Reduced caster DPS gives weapon DPS from fitted shares of the damage table

- Skill: wowsims-consumables-items
- Code: `tools/database/dbc/item_damage.go`, `reducedCasterDpsShare`
- Pull request: #298
- Evidence: none

ItemSparse `REDUCED_CASTER_DPS` gives weapon DPS = damage table × 2/3 one-hand or × 29/39 two-hand, with QualityModifier
an additive offset; the shares are fitted to in-game ranges, not read from the game files.

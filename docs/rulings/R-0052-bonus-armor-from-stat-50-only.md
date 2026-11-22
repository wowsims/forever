# R-0052: An item's bonus armor is its stat 50, never its QualityModifier

- Skill: wowsims-consumables-items
- Code: `tools/database/dbc/maps.go`, `MapBonusStatIndexToStat`; `tools/database/dbc/item.go`, `GetWeaponQualityModifier`
- Pull request: #PR
- Evidence: EV-0176

An item's bonus armor is the amount of its stat 50 (`ITEM_MOD_EXTRA_ARMOR`) and nothing else. The `QualityModifier` an
armor item carries gives no armor in game, so an item with only a `QualityModifier` has no bonus armor, and one with
both has the stat 50 amount, whatever its `QualityModifier`. The ruling covers every armor item, the ones no ledger has measured
included.

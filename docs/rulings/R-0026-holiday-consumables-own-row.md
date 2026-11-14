# R-0026: Holiday consumables sit in their own row, stack with food and are not Well Fed

- Skill: wowsims-consumables-items
- Code: `tools/database/gen_consumables.go`, `consumableHoliday`; `sim/common/shared/consumable_stats.go`, `ConsumableSpells`
- Pull request: #207
- Evidence: none

Holiday consumables (`Consumable.holiday`) sit in their own row above Miscellaneous, are not Well Fed and stack with
food; a Forever food's stats are the amount Nutritious Food's `A_PERIODIC_TRIGGER_SPELL_WITH_VALUE` hands Well Fed.

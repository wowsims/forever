# R-0030: Caster spell power and healing power follow the item budget formulas

- Skill: wowsims-consumables-items
- Code: `tools/database/dbc/item.go`, `addCasterSpellPower`
- Pull request: #302
- Evidence: pending

`CASTER_SPELL_POWER` gives spell power ⌊20000 × budget / 10000⌋ and `CASTER_HEALING_POWER` gives healing ⌊37600 × budget
/ 10000⌋ with spell damage ⌊healing / 3⌋, budget being RandPropPoints column 0.

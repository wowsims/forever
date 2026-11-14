# R-0028: A survival cooldown that adds no damage or mana disables autocast

- Skill: wowsims-consumables-items
- Code: `sim/core/major_cooldown.go`
- Pull request: #335
- Evidence: not needed: a sim design choice

A survival cooldown that adds no damage or mana has `DisableAutocast` set unless the character sits in a raid Tanks
slot, so Autocast Other Cooldowns uses it only at a cooldown timing; defensive on-use buffs count as Survival and MP5 or
Spirit ones as Mana; explicit APL casts are unaffected.

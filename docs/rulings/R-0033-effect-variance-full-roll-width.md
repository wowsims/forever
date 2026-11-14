# R-0033: EffectVariance is the full roll width

- Skill: wowsims-core-mechanics
- Code: `sim/core/spell_result.go`; `sim/core/spelldata/effect.go` `RollAt`
- Pull request: #267
- Evidence: none

`EffectVariance` is the full roll width: spell hits and heals roll the whole amount, power included, by ±V/2
(`Spell.DirectVariance` after power), and physical effects roll the base by ±V/2, rounded to a whole number, before
attack power.

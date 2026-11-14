# R-0021: During a hardcast the pick reads the cast's own state as landed

- Skill: wowsims-apl
- Code: `sim/core/apl_values_aura.go`
- Pull request: #640
- Evidence: none

During the hardcast the pick reads the cast's own dot, own-id aura, cooldown and mana as landed; the full reading and
the lines that need `Spell Is Casting` are `docs/rotation.md`.

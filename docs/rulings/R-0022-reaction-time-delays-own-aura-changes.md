# R-0022: Include Reaction Time delays every aura change, the unit's own casts included

- Skill: wowsims-apl
- Code: `sim/core/apl_values_aura.go`
- Pull request: #638, #640
- Evidence: none

Include Reaction Time delays every aura change the same way, the unit's own casts included; during a hardcast the cast's
own aura reads as landed instead.

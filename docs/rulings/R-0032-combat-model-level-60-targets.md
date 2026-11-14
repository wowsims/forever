# R-0032: The combat model is level 60 against targets of level 58 to 63

- Skill: wowsims-core-mechanics
- Code: `sim/core/target.go`
- Pull request: #70
- Evidence: none

The combat model is level 60 with target levels 58 to 63: attack tables come from the level difference, capped at +3,
and per-hand weapon skill.

# R-0023: An APL value reads the unit's reaction time when it runs

- Skill: wowsims-apl
- Code: `sim/core/gcd.go` `drawReactionTime`, `sim/core/apl_values_aura.go`
- Pull request: #848
- Evidence: none

An APL value reads `unit.ReactionTime` when it runs, never a copy taken when the rotation is built: with a reaction
range (`reaction_time_max_ms`) each rotation pass draws a new one.

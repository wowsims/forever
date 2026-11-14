# R-0036: Mana regenerates in 100 ms ticks and a paid channel holds regen

- Skill: wowsims-core-mechanics
- Code: `sim/core/mana.go`
- Pull request: #249
- Evidence: EV-0107

Mana regenerates in 100 ms ticks; a paid channel holds regen until max(start + 5 s, end) and a free channel does not.

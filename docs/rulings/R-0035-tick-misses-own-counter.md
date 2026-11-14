# R-0035: Tick misses count in tick_misses and not in Misses

- Skill: wowsims-core-mechanics
- Code: `sim/core/spell_outcome.go`
- Pull request: #570
- Evidence: none

Tick misses count in `TargetedActionMetrics.tick_misses`, not `Misses`, and Consecration's own cast roll does not count
(`OutcomeTickMagicHitNoHitCounter`).

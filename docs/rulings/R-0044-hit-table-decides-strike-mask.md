# R-0044: The hit table decides a strike's mask, not its school

- Skill: wowsims-procs
- Code: `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask` and `procDamageHandler`
- Pull request: #352
- Evidence: pending

The hit table decides a strike's mask, not its school (Bow of Searing Arrows is a ranged special); a strike's own
re-cast during its cast goes one batch later.

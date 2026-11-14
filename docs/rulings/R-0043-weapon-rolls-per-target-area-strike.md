# R-0043: A chance-on-hit weapon rolls once for each target an area strike lands on

- Skill: wowsims-procs
- Code: `sim/core/aura_helpers.go`, `matchesSpell`; `sim/warrior/talents_arms.go`, `core.OncePerCast`
- Pull request: #986
- Evidence: EV-0001

A chance-on-hit weapon rolls once for each target an area melee strike lands on, since the hit listener hears each
target's result. Weaponmaster's sword roll is the exception and rolls once a cast.

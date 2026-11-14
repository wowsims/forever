# R-0007: Deep Wounds bleeds a share of weapon damage and its ticks never crit

- Skill: wowsims-abilities-auras
- Code: `sim/warrior/talents_arms.go`
- Pull request: #231
- Evidence: none

Deep Wounds bleeds the talent share of average weapon damage without attack power; a new crit rolls the owed damage over
and extends the timer without resetting it, and the ticks never crit.

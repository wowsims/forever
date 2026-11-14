# R-0004: Intercept's stun is its own spell and level plus 3 targets are immune to it

- Skill: wowsims-abilities-auras
- Code: `sim/warrior/intercept.go`
- Pull request: #652
- Evidence: none

Intercept's stun is its own spell from the stun row, rolling `OutcomeMagicHitAndCrit` with an empty proc mask; targets
at character level + 3 are immune to its stun and damage, and nothing is dealt at fight-end teardown.

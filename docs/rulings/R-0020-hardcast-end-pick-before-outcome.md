# R-0020: At a hardcast end the rotation picks before the cast's outcome applies

- Skill: wowsims-apl
- Code: `sim/core/gcd.go`
- Pull request: #640
- Evidence: not needed: a sim design choice

At a player's hardcast end the rotation picks at the cast-end timestamp, before the cast's outcome applies, and a
pushback re-queues the pick; the major cooldowns that go ahead of it and the units that look after the landing are
`docs/rotation.md` §The flow.

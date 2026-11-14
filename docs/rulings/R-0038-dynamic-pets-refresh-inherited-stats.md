# R-0038: Dynamic pets refresh their inherited stats from the owner at the pet heartbeat

- Skill: wowsims-core-mechanics
- Code: `sim/core/pet.go`
- Pull request: #432
- Evidence: none

Dynamic pets (demons, hawks) refresh their inherited stats from the owner's totals at the pet heartbeat instead of
running stat deltas through the inheritance.

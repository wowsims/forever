# R-0031: An APL aura condition reads an aura only once it is active

- Skill: wowsims-core-mechanics
- Code: `sim/core/aura_helpers.go`
- Pull request: #638
- Evidence: none

An APL aura condition reads an aura only once it is active: a ProcTrigger declares no pending aura, and a proc that must
land in the same batch as its hit sets `TriggerImmediately`, only with log proof.

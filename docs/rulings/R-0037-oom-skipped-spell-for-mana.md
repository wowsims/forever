# R-0037: OOM runs while the highest-priority wanted spell is skipped for mana

- Skill: wowsims-core-mechanics
- Code: `sim/core/spell.go`, `meetsCost`
- Pull request: #371
- Evidence: not needed: a sim design choice

OOM runs while the highest-priority spell the APL wants (condition true, ready but for mana) is skipped for mana,
whatever casts instead; it starts at the first free rotation pass (GCD ready, no hardcast) that skips one and ends at
the first that skips none.

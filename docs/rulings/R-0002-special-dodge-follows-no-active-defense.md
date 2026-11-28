# R-0002: A special can be dodged, parried or blocked only when it lacks No Active Defense

- Skill: wowsims-abilities-auras
- Code: `sim/core/unit.go`, `GetTotalDodgeChanceAsDefender`
- Pull request: #712
- Evidence: pending

Whether a special can be dodged, parried or blocked follows the spell's No Active Defense (`Attributes[0] 0x200000`),
which `rowFlags` turns into `SpellFlagNoActiveDefense`; the outcome function never encodes it, so Backstab and Feint,
which lack the bit, can be dodged.

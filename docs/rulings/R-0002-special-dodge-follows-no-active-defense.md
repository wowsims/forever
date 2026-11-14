# R-0002: A special can be dodged, parried or blocked only when its row lacks No Active Defense

- Skill: wowsims-abilities-auras
- Code: `sim/core/unit.go`, `GetTotalDodgeChanceAsDefender`
- Pull request: #712
- Evidence: none

Whether a special can be dodged, parried or blocked follows the row's No Active Defense (`Attributes[0] 0x200000`),
which `rowFlags` turns into `SpellFlagNoActiveDefense`; the outcome function never encodes it, so Backstab and Feint,
whose rows lack the bit, can be dodged.

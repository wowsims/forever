# R-0024: Is Hit Capped and Is Expertise Capped are per spell on the current target

- Skill: wowsims-apl
- Code: `sim/core/apl_values_spell.go`
- Pull request: #608
- Evidence: not needed: a sim design choice

`Is Hit Capped` and `Is Expertise Capped` are per spell, current target only and yes or no: melee caps at the
special-attack soft cap without the dual-wield penalty, the table follows the spell's `DefenseType` (proc mask for
none), auto shots roll the ranged table, and the 1e-9 tolerance stays in the APL values.

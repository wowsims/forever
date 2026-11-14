# R-0042: No weapon proc hears a melee or ranged special with no DefenseType

- Skill: wowsims-procs
- Code: `sim/core/aura_helpers.go`, `canProcFrom`; `sim/warrior/talents_arms.go`, `registerSweepingStrikes`
- Pull request: #965
- Evidence: EV-0007

No weapon proc hears a melee or ranged special whose row states no DefenseType; auto attacks stay exempt. Sweeping
Strikes' copied hit 12723 has no SpellCategories row and procs no chance-on-hit weapon, while its normalized hit 26654,
with melee DefenseType, rolls them. Equip procs such as Hand of Justice and Dreadfrost Saber's still hear 12723, unmeasured there.

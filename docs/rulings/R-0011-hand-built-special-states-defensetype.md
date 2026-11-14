# R-0011: A hand-built melee or ranged special states its row's DefenseType

- Skill: wowsims-abilities-auras
- Code: `sim/warrior/talents_arms.go`, `registerSweepingStrikes`
- Pull request: #965
- Evidence: not needed: follows R-0042

A hand-built melee or ranged special states its row's DefenseType, since a special with none is no weapon hit and no
chance-on-hit item effect or combat enchant hears it. Sweeping Strikes' copied hit 12723 has no SpellCategories row and
registers with none; its normalized hit 26654 is melee (R-0042).

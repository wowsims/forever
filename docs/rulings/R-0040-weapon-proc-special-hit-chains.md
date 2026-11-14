# R-0040: A weapon proc with melee or ranged DefenseType is a special hit that chains to weapon procs

- Skill: wowsims-procs
- Code: `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask`
- Pull request: #352
- Evidence: pending

A weapon proc whose row has melee or ranged DefenseType (direct hits and DoT-only applications) is a special hit:
triggered by a main-hand or ranged hit it chains to weapon procs (its own weapon included), combat enchants and oils;
triggered by an off-hand hit it is heard by none (main hand chains, off hand nothing). Magic weapon procs stay
`ProcMaskEmpty`, equip and aura procs keep the `SpellFlagProc` gate, and Windfury never hears them.

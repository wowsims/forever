# R-0012: Casters register every rank only when the implementation requires it

- Skill: wowsims-abilities-auras
- Code: `sim/core/spelldata/resolve_aura.go`
- Pull request: #246
- Evidence: none

Casters register every rank only when the implementation requires it, a question asked during integration (melee
registers `Highest()` only, apart from the class book ranks below); each per-rank DoT or HoT is labelled with
`spelldata.RankLabel(rank, &owner.Unit)` so ranks never share one aura; category cooldowns already share one timer
across ranks, so no hand `NewTimer()`.

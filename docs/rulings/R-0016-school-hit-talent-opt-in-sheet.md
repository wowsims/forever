# R-0016: A school hit talent shows on the sheet only by opting in

- Skill: wowsims-abilities-auras
- Code: `sim/core/spelldata/parse_effects.go`
- Pull request: #285
- Evidence: not needed: a sim design choice

A school hit talent (SPELLMOD_RESIST_MISS_CHANCE on class flags) shows on the sheet by opting in with
`spelldata.ShowAsSchoolHit(schools)` and its tooltip's schools, never `SchoolBonusHitChance`; all-spell hit talents
(aura 55 misc 0) already show and must not opt in.

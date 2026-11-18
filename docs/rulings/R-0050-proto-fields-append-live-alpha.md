# R-0050: Saved settings carry proto field numbers, so a new field takes the next free number

- Skill: wowsims-core-mechanics
- Code: `proto/`, `tools/database/buffmanifest/buffs.go`
- Pull request: #1092
- Evidence: not needed: a repo process rule

The sim is live in alpha and saved settings carry proto field numbers: a new field takes the next free number at the end
of its message, and no field is dropped, moved or reused. Every `// NextIndex:` is the highest used number + 1. A buff
manifest row's place in its scope's slice is its field number, so a new row goes at the end of its slice.

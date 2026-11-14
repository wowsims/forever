# R-0039: The sim is not live, so breaking saved settings needs no migration

- Skill: wowsims-core-mechanics
- Code: `proto/common.proto`
- Pull request: #149
- Evidence: none

The sim is not live, so a change that breaks saved APLs or settings needs no migration, only a note in the PR body; a
removed proto field leaves no `reserved` line, new fields refill freed numbers and every `// NextIndex:` is the highest
used number + 1.

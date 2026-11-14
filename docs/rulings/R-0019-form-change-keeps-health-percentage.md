# R-0019: A druid's change of form keeps the current health percentage

- Skill: wowsims-abilities-auras
- Code: `sim/druid/forms.go`
- Pull request: #149
- Evidence: none

A druid's change of form keeps the current health percentage both ways and a Dire Bear guardian starts the pull at 100%;
any new form or max-health form effect sits inside `keepHealthPercentAcrossForms`.

# R-0003: A spell with magic DefenseType and no Always Hit rolls a magic hit even without damage

- Skill: wowsims-abilities-auras
- Code: `sim/warrior/taunt.go`
- Pull request: #649
- Evidence: none

A spell whose row has DefenseType magic and no Always Hit rolls `OutcomeMagicHit` even when it deals no damage (Taunt,
Challenging Shout, Intimidating Shout, Growl, Challenging Roar); taunts carry `ProcMaskEmpty`.

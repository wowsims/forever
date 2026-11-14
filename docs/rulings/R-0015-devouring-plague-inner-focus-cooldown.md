# R-0015: Devouring Plague is learnable by every race and Inner Focus restarts its cooldown

- Skill: wowsims-abilities-auras
- Code: `sim/priest/devouring_plague.go`, `sim/priest/talents_discipline.go`
- Pull request: #246
- Evidence: pending

Devouring Plague is learnable by every race, and Inner Focus restarts its cooldown when the buff is spent (`OnExpire`
calls `CD.Use`) though the row has no cooldown-on-event bit.

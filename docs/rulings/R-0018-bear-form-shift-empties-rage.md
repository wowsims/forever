# R-0018: A shift into Bear Form empties the rage bar before Furor's rage

- Skill: wowsims-abilities-auras
- Code: `sim/druid/forms.go`
- Pull request: #149
- Evidence: none

A shift into Bear or Dire Bear Form empties the rage bar in the shift spell's `ApplyEffects`, Furor's 10 rage comes
after, and the starting form at the pull keeps `StartingRage`.

# R-0009: Seal of Righteousness keeps hand coefficients over the rows' zero

- Skill: wowsims-abilities-auras
- Code: `sim/paladin/seal_of_righteousness.go`
- Pull request: #469; the one-hander coefficient in #473
- Evidence: EV-0059

Seal of Righteousness adds spell power at 0.12 on a two-hander and 0.1 on a one-hander, over the rows' 0: the damage
spells carry no coefficient. 0.12 is the tooltip's description variable `$2h`. The tooltip's `$1h`, 0.0765, fits no
one-hander, whose coefficient is measured.

# R-0006: DoTs do not snapshot and every tick uses the stats of the moment it lands

- Skill: wowsims-abilities-auras
- Code: `sim/core/spelldata/resolve_aura.go`, `DotConfig`
- Pull request: none
- Evidence: none

DoTs do not snapshot: every tick deals damage on the stats and modifiers of the moment it lands
(`CalcAndDealPeriodicDamage` in `OnTick`, no `OnSnapshot`), Deep Wounds' rollover pool being the one amount fixed at the
crit.

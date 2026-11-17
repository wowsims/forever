package buffs

import (
	"github.com/wowsims/forever/sim/core"
)

// The aura on the unit the shield protects. The external (individual-buff) copy takes the
// providing druid's Nature spell power as externalSpellPower.
//
// The druid's spell damage against the attacker's creature type adds nothing to Thorns, though it
// reaches every other spell and tick of the druid's (EV-0139). Only Thorns is measured, so Retribution
// Aura and the item shields still take it.
func NewSpellDataThornsAuraWithSpellPower(unit *core.Unit, isPlayer bool, externalSpellPower float64) *core.Aura {
	return newSpellPowerDamageShield(unit, thornsMeta, isPlayer, ThornsValue(0), externalSpellPower,
		core.SpellFlagNoSpellPowerVersus)
}

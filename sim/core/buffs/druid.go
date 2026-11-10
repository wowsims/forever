package buffs

import (
	"github.com/wowsims/forever/sim/core"
)

// The aura on the unit the shield protects. The external (individual-buff) copy takes the
// providing druid's Nature spell power as externalSpellPower.
func NewSpellDataThornsAuraWithSpellPower(unit *core.Unit, isPlayer bool, externalSpellPower float64) *core.Aura {
	return newSpellPowerDamageShield(unit, thornsMeta, isPlayer, ThornsValue(0), externalSpellPower)
}

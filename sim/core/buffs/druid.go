package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The rows whose aura amounts round to the nearest whole number in game, which the manifest has no
// column for. Demoralizing Roar's attack power reduction: rank 9747 read -round(144 + 1.4 × (level - 42))
// on 400 of 403 applications at levels 46 to 51, where the floor fits 229 and the ceiling 188.
// TODO: Ingame research needed -- the level 60 rank 9898, which this takes to round the same way.
var nearestAmountRows = map[*spelldata.Spell]bool{demoralizingRoarSpell: true}

// The aura on the unit the shield protects. The external (individual-buff) copy takes the
// providing druid's Nature spell power as externalSpellPower.
func NewSpellDataThornsAuraWithSpellPower(unit *core.Unit, isPlayer bool, externalSpellPower float64) *core.Aura {
	return newSpellPowerDamageShield(unit, thornsMeta, isPlayer, ThornsValue(0), externalSpellPower)
}

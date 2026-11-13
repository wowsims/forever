package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The aura on the unit the shield protects. The external (individual-buff) copy takes the
// providing druid's Nature spell power as externalSpellPower.
func NewSpellDataThornsAuraWithSpellPower(unit *core.Unit, isPlayer bool, externalSpellPower float64) *core.Aura {
	return newSpellPowerDamageShield(unit, thornsMeta, isPlayer, ThornsValue(0), externalSpellPower)
}

// Innervate's mana regen percent multiplies all of the regen, Spirit and MP5 alike, and its other
// effect keeps the Spirit regen running while casting, so the regen is the same inside the
// five-second rule as outside it. The other rows with a mana regen percent keep the Spirit reading.
func NewSpellDataInnervatesAuraWithMP5(unit *core.Unit, isPlayer bool) *core.Aura {
	return innervatesMeta.buff(unit, isPlayer, innervatesMeta.label(isPlayer), 0,
		append(innervatesMeta.Options(0), spelldata.AllManaRegen()))
}

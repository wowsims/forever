package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A priest's prayer reads its own row and bids as the raid's Power Word: Fortitude does, so the two
// count once.
func NewSpellDataPrayerOfFortitudeAura(unit *core.Unit, isPlayer bool, prayer *spelldata.Spell) *core.Aura {
	return newBuff(unit, powerWordFortitudeMeta.atState(prayer, prayer.Name), isPlayer, 0)
}

// A priest's prayer reads its own row and bids in the shadow school as the raid's Shadow Protection
// does, so the two count once.
func NewSpellDataPrayerOfShadowProtectionAura(unit *core.Unit, isPlayer bool, prayer *spelldata.Spell) *core.Aura {
	return newBuff(unit, shadowProtectionMeta.atState(prayer, prayer.Name), isPlayer, 0)
}

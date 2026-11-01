package buffs

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var prayerOfFortitudeRow = spelldata.MustFind(21564)
var prayerOfShadowProtectionRow = spelldata.MustFind(27683)

// A priest's prayer reads its own row and bids as the raid's Power Word: Fortitude does, so the two
// count once.
func NewSpellDataPrayerOfFortitudeAura(unit *core.Unit, isPlayer bool, talentPoints int32) *core.Aura {
	return newBuff(unit, powerWordFortitudeMeta.atState(prayerOfFortitudeRow, prayerOfFortitudeRow.Name), isPlayer, talentPoints)
}

// A priest's prayer reads its own row and bids in the shadow school as the raid's Shadow Protection
// does, so the two count once.
func NewSpellDataPrayerOfShadowProtectionAura(unit *core.Unit, isPlayer bool, talentPoints int32) *core.Aura {
	return newBuff(unit, shadowProtectionMeta.atState(prayerOfShadowProtectionRow, prayerOfShadowProtectionRow.Name), isPlayer, talentPoints)
}

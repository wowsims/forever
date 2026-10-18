package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var pvpGloveItemIDs = []int32{23279, 22862, 16463, 16571}

// The gloves' equip spell: Multi-Shot deals 4% more.
var pvpGloveMultiShot = spelldata.MustFind(28539)

func init() {
	for _, itemID := range pvpGloveItemIDs {
		core.NewItemEffect(itemID, func(_ core.Agent) {})
	}
}

func (hunter *Hunter) addPvpGloves() {
	hunter.RegisterPvPGloveMod(
		pvpGloveItemIDs,
		core.SpellModConfig{
			Kind:       core.SpellMod_DirectDamageDone_Pct,
			ClassFlags: multiShotRank.ClassFlags,
			FloatValue: pvpGloveMultiShot.EffectN(1).Percent(),
		})
}

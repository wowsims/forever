package priest

import (
	"github.com/wowsims/forever/sim/common/shared"
)

func init() {
	// Benediction
	// https://www.wowhead.com/forever/item=18608/benediction
	//
	// Increases the critical effect chance of your Holy spells by 2%.
	//
	// Reaches the Holy spells by class flags; Renew is not in its mask.
	// The staff's on-use 23041 turns it into Anathema and has no effect in the sim.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 23236},
		[]shared.ItemVariant{
			{ItemID: 18608, ItemName: "Benediction"},
		})
}

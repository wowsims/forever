package druid

import (
	"github.com/wowsims/forever/sim/common/shared"
)

func init() {
	// Pristine Enchanted South Seas Kelp
	// https://www.wowhead.com/forever/item=19613/pristine-enchanted-south-seas-kelp
	//
	// Increases the critical hit chance of Wrath and Starfire by 2%.
	//
	// Reaches Wrath and Starfire by class flags.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 24433},
		[]shared.ItemVariant{
			{ItemID: 19613, ItemName: "Pristine Enchanted South Seas Kelp"},
		})
}

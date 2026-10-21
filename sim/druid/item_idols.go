package druid

import (
	"github.com/wowsims/forever/sim/common/shared"
)

// The idols' equip spells reach their spells by class flags.
//
// TODO: Items without full game files data are not registered here:
//   - Idol of Shifting Tides 263411 (1270470, shapeshifting spells cost 40 less mana).
//   - Mark of Urs'endris 263435 (1270490, armor from items +4%, base and bonus).
func init() {
	// Idol of Health
	// https://www.wowhead.com/forever/item=22399/idol-of-health
	//
	// Reduces the casting time of your Healing Touch spell by 0.15 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27846},
		[]shared.ItemVariant{
			{ItemID: 22399, ItemName: "Idol of Health"},
		})

	// Howling Idol
	// https://www.wowhead.com/forever/item=272427/howling-idol
	//
	// Reduces the cooldown of your Shifting Power ability by 0.5 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291059},
		[]shared.ItemVariant{
			{ItemID: 272427, ItemName: "Howling Idol"},
		})

	// Enraged Idol
	// https://www.wowhead.com/forever/item=272428/enraged-idol
	//
	// Your Enrage ability generates an additional 10 Rage over its duration.
	//
	// TODO: 1291060 states the 10 as SPELLMOD_EFFECT1 on Enrage, which the parse reads as base
	// damage. The Enrage port has to add the rage off its effect 1.
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291060},
	// 	[]shared.ItemVariant{
	// 		{ItemID: 272428, ItemName: "Enraged Idol"},
	// 	})

	// Swarming Idol
	// https://www.wowhead.com/forever/item=272430/swarming-idol
	//
	// Increases the duration of your Insect Swarm ability by 2 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291061},
		[]shared.ItemVariant{
			{ItemID: 272430, ItemName: "Swarming Idol"},
		})

	// Mystic Mushroom
	//
	// Increases your Spirit by 5%.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1248751},
		[]shared.ItemVariant{
			{ItemID: 249396, ItemName: "Mystic Mushroom"},
		})

	// Idol of Swiftness
	// https://www.wowhead.com/forever/item=279250/idol-of-swiftness
	//
	// Reduces the cooldown of your Swiftmend spell by 3 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1306488},
		[]shared.ItemVariant{
			{ItemID: 279250, ItemName: "Idol of Swiftness"},
		})
}

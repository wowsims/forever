package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	// Atiesh, Greatstaff of the Guardian
	// https://www.wowhead.com/forever/item=22589/atiesh-greatstaff-of-the-guardian
	//
	// Increases the spell critical chance of all party members within 30 yards by 2%.
	//
	// 28142 is a party aura, and the wearer is in its own party, so the staff puts the party buff's own
	// copy on the wearer. The party buff's Atiesh - Mage count is the staves the other party members
	// carry.
	core.NewItemEffect(22589, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(buffs.NewSpellDataAtieshMageAura(&character.Unit, true, 0, 1))
		// Worn from the start, the staff is part of the gear the character sheet shows.
		if character.HasItemEquipped(22589, []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}) {
			aura.BuildPhase = core.CharacterBuildPhaseGear
		}
		character.ItemSwap.RegisterProc(22589, aura)
	})
}

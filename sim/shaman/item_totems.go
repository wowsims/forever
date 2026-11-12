package shaman

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Totem of the Storm's dummy 1291078: the share of Maelstrom Weapon's rate Lightning Bolt rolls at, 50.
var totemOfTheStormShare = spelldata.MustFind(1291078).EffectN(1).Percent()

func init() {
	// Polished Driftwood Icon
	// https://www.wowhead.com/forever/item=249398/polished-driftwood-icon
	//
	// Allows 8% of your Mana regeneration to continue while casting.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1248756},
		[]shared.ItemVariant{
			{ItemID: 249398, ItemName: "Polished Driftwood Icon"},
		})

	// Tidal Totem
	// https://www.wowhead.com/forever/item=272431/tidal-totem
	//
	// Reduces the Mana cost of your Healing Wave ability by 5%.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291076},
		[]shared.ItemVariant{
			{ItemID: 272431, ItemName: "Tidal Totem"},
		})

	// Totem of the Storm
	// https://www.wowhead.com/forever/item=272432/totem-of-the-storm
	//
	// Your Lightning Bolt ability can now also trigger the Maelstrom Weapon talent, but with a 50%
	// reduced chance.
	//
	// A landed Lightning Bolt rolls at half the talent's rate per minute, priced on the bolt's cast
	// time after talent reductions; cast speed and Maelstrom Weapon's own stacks do not change it.
	// Item effects apply before the talents, so the buff is looked up when the bolt lands.
	core.NewItemEffect(272432, func(agent core.Agent) {
		shaman := agent.(ShamanAgent).GetShaman()
		if shaman.Talents.MaelstromWeapon == 0 {
			return
		}

		ppm := maelstromWeaponPPM(spellData.MaelstromWeapon.Rank(shaman.Talents.MaelstromWeapon)) * totemOfTheStormShare
		var maelstromAura *core.Aura
		aura := shaman.MakeProcTriggerAura(core.ProcTrigger{
			Name:           "Totem of the Storm",
			ActionID:       core.ActionID{SpellID: 1291078},
			Callback:       core.CallbackOnSpellHitDealt,
			ProcMask:       core.ProcMaskSpellDamage,
			Outcome:        core.OutcomeLanded,
			ClassSpellMask: SpellMaskLightningBolt,
			Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				if !sim.Proc(ppm*spell.DefaultCast.CastTime.Minutes(), "Totem of the Storm") {
					return
				}
				if maelstromAura == nil {
					maelstromAura = shaman.GetAura(maelstromWeaponBuff.Name)
				}
				maelstromAura.Activate(sim)
				maelstromAura.AddStack(sim)
			},
		})

		shaman.ItemSwap.RegisterProc(272432, aura)
	})

	// Burning Totem
	// https://www.wowhead.com/forever/item=272433/burning-totem
	//
	// Increases the duration of your Flame Shock ability by 3 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291077},
		[]shared.ItemVariant{
			{ItemID: 272433, ItemName: "Burning Totem"},
		})

	// Totem of Urgency
	// https://www.wowhead.com/forever/item=279249/totem-of-urgency
	//
	// Increases the critical strike chance of your Lesser Healing Wave spell by 4%.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1306448},
		[]shared.ItemVariant{
			{ItemID: 279249, ItemName: "Totem of Urgency"},
		})
}

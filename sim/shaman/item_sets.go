package shaman

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var (
	stormcallerProc       = spelldata.MustFind(26119) // 3pc: the proc
	stormcallerWrath      = spelldata.MustFind(26121) // 3pc: Stormcaller's Wrath, the Nature damage
	stormcallerChainHeal  = spelldata.MustFind(26122) // 5pc: Chain Heal's cast time
	gatheringStormChain   = spelldata.MustFind(26123) // 3pc: Chain Lightning's jump multiplier
	earthshatterTotemCost = spelldata.MustFind(28818) // 2pc: the totems' cost
	earthshatterSpring    = spelldata.MustFind(29171) // 4pc: Mana Spring Totem's mana
	earthshatterShield    = spelldata.MustFind(28821) // 8pc: the trigger on Lightning Shield
	earthshatterShieldMP5 = spelldata.MustFind(28820) // 8pc: the mana per 5 sec
)

// Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=501/stormcallers-garb
var ItemSetStormcallersGarb = core.NewItemSet(core.ItemSet{
	ID:   501,
	Name: "Stormcaller's Garb",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		// Your Lightning Bolt, Chain Lightning, and Shock spells have a 20% chance to grant up to 50
		// Nature damage to spells for 8s.
		// The proc effect names class family 11 and none of its spells, so the listener keeps to the
		// spells the tooltip names; the copies Lightning Overload casts are not among them. The spell
		// hears harmful spells, which the decode reads as the cast, so a Chain Lightning rolls once
		// however many targets it hits.
		// TODO: Ingame test needed -- whether only Lightning Bolt, Chain Lightning and the shocks roll it,
		// once per cast.
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			shaman := agent.(ShamanAgent).GetShaman()

			wrath := shaman.RegisterAura(spelldata.AuraConfig(stormcallerWrath))
			spelldata.ParseEffects(&shaman.Character, wrath, stormcallerWrath)

			trigger := spelldata.ProcTrigger(&shaman.Character, stormcallerProc,
				func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
					wrath.Activate(sim)
				})
			trigger.Name = "Stormcaller's Garb - 3PC"
			trigger.ClassSpellMask = SpellMaskLightningBolt | SpellMaskChainLightning | SpellMaskShock

			setBonusAura.AttachProcTrigger(trigger).ExposeToAPL(stormcallerProc.ID)
		},
		5: shared.ParsedSetBonus(stormcallerChainHeal),
	},
})

// Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=502/gift-of-the-gathering-storm
var ItemSetGiftOfTheGatheringStorm = core.NewItemSet(core.ItemSet{
	ID:   502,
	Name: "Gift of the Gathering Storm",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotBack,
		proto.ItemSlot_ItemSlotFinger1,
		proto.ItemSlot_ItemSlotFinger2,
		proto.ItemSlot_ItemSlotMainHand,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		// Increases the chain target damage multiplier of your Chain Lightning spell by 5%.
		// The modifier (SPELLMOD_DAMAGE_MULTIPLIER, misc 20) scales the share of the damage each jump
		// keeps, the effect's ChainAmp. It is a percent modifier, so it multiplies that share.
		// TODO: Ingame test needed -- whether the 5% multiplies each jump's share or adds to it.
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			shaman := agent.(ShamanAgent).GetShaman()
			multiplier := 1 + gatheringStormChain.EffectN(1).Percent()

			setBonusAura.
				AttachMultiplicativePseudoStatBuff(&shaman.chainLightningJumpMultiplier, multiplier).
				ExposeToAPL(gatheringStormChain.ID)
		},
	},
})

// Tier 1 - Enhancement
// https://www.wowhead.com/forever/item-set=2110/the-spiritcallers-rage
//
// The game files ship four of the six pieces (shoulders, hands, legs, feet), so the 5-piece bonus
// waits on the head and chest the ItemSet row lists. The 3-piece bonus widens the earth and air
// totems' radius, which the sim does not model.
var ItemSetTheSpiritcallersRage = core.NewItemSet(core.ItemSet{
	ID:   2110,
	Name: "The Spiritcaller's Rage",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotHands,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1300957)), // +1% attack and casting speed
		4: shared.ParsedSetBonus(spelldata.MustFind(1301090)), // +36 melee attack power against Elementals
		5: shared.ParsedSetBonus(spelldata.MustFind(1301709)), // Stormstrike cooldown -0.5 sec
	},
})

// Tier 3
// https://www.wowhead.com/forever/item-set=527/the-earthshatterer
var ItemSetTheEarthshatterer = core.NewItemSet(core.ItemSet{
	ID:    527,
	Name:  "The Earthshatterer",
	Slots: shared.TierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(earthshatterTotemCost),
		4: shared.ParsedSetBonus(earthshatterSpring),
		// TODO: 6pc (28823) gives Healing Wave and Lesser Healing Wave a chance to imbue the target
		// with Totemic Power, which the sim does not model.

		// Your Lightning Shield spell also grants you 15 mana per 5 sec. while active.
		// registerEarthshattererShieldMP5 links the mana to Lightning Shield, which registers after the set bonuses.
		// TODO: Ingame test needed -- whether the mana per 5 sec applies only while Lightning Shield is up.
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			shaman := agent.(ShamanAgent).GetShaman()
			shaman.earthshattererAura = setBonusAura
			setBonusAura.ExposeToAPL(earthshatterShield.ID)
		},
	},
})

// 28821 puts 28820 on the shaman when Lightning Shield lands. Effect 2 of 28820 is a periodic dummy.
func (shaman *Shaman) registerEarthshattererShieldMP5() {
	// 28820 shares Lightning Shield's name.
	mp5 := shaman.RegisterAura(spelldata.AuraConfig(earthshatterShieldMP5, spelldata.Label("Lightning Shield - The Earthshatterer")))
	spelldata.ParseEffects(&shaman.Character, mp5, earthshatterShieldMP5, spelldata.Effects(1))

	follow := func(_ *core.Aura, sim *core.Simulation) {
		if shaman.earthshattererAura.IsActive() && shaman.LightningShieldAura.IsActive() {
			mp5.Activate(sim)
		} else {
			mp5.Deactivate(sim)
		}
	}
	shaman.LightningShieldAura.ApplyOnGain(follow).ApplyOnExpire(follow)
	shaman.earthshattererAura.ApplyOnGain(follow).ApplyOnExpire(follow)
}

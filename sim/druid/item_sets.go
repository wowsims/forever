package druid

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func parsedSetBonus(rows ...*spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		for _, row := range rows {
			spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
		}
	}
}

var sixPieceSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// Tier 1 - Balance
// https://www.wowhead.com/forever/item-set=2114/grovekeeper-eclipse
//
// The game files ship four of the six pieces (shoulders, hands, legs, feet), so the 5-piece bonus
// waits on the head and chest the ItemSet row lists. The 3-piece bonus takes 15 sec off Nature's
// Grasp's cooldown, which the sim does not cast.
var ItemSetGrovekeeperEclipse = core.NewItemSet(core.ItemSet{
	ID:    2114,
	Name:  "Grovekeeper Eclipse",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(spelldata.MustFind(1300940)), // +1% hit
		4: func(_ core.Agent, setBonusAura *core.Aura) {
			row := spelldata.MustFind(1301073)
			shared.AttachSpellDamageVersus(setBonusAura, row).ExposeToAPL(row.ID)
		},
		5: parsedSetBonus(spelldata.MustFind(1301242)), // Insect Swarm lasts 3 sec longer
	},
})

// Tier 1 - Feral
// https://www.wowhead.com/forever/item-set=2115/grovekeeper-ferocity
//
// The game files ship four of the six pieces (shoulders, hands, legs, feet), so the 5-piece bonus
// waits on the head and chest the ItemSet row lists. The 3-piece bonus lets Hibernate be cast in Cat
// Form, which the sim does not cast.
var ItemSetGrovekeeperFerocity = core.NewItemSet(core.ItemSet{
	ID:    2115,
	Name:  "Grovekeeper Ferocity",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1300934)), // +1% attack and casting speed
		4: shared.ParsedSetBonus(spelldata.MustFind(1301075)), // +36 melee attack power against Demons
		5: shared.ParsedSetBonus(spelldata.MustFind(1301247)), // Shifting Power cooldown -1 sec
	},
})

// Tier 1 - Guardian
// https://www.wowhead.com/forever/item-set=2113/grovekeeper-rage
//
// The game files ship four of the six pieces (shoulders, hands, legs, feet), so the 5-piece bonus
// waits on the head and chest the ItemSet row lists. The 3-piece bonus lets Rebirth be cast in Bear
// Form, and the 5-piece takes 15 sec off Berserk; the sim casts neither.
var ItemSetGrovekeeperRage = core.NewItemSet(core.ItemSet{
	ID:    2113,
	Name:  "Grovekeeper Rage",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: shared.ParsedSetBonus(spelldata.MustFind(1300942)), // +7 Defense skill
		4: shared.ParsedSetBonus(spelldata.MustFind(1301076)), // +12 expertise rating
	},
})

// Zul'Gurub
// https://www.wowhead.com/forever/item-set=479/haruspexs-garb
var ItemSetHaruspexsGarb = core.NewItemSet(core.ItemSet{
	ID:   479,
	Name: "Haruspex's Garb",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotNeck,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotWrist,
		proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotTrinket1,
		proto.ItemSlot_ItemSlotTrinket2,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(spelldata.MustFind(21618)), // 4 mana per 5 sec
		3: parsedSetBonus(spelldata.MustFind(24479)), // Faerie Fire lasts 5 sec longer
		5: parsedSetBonus(spelldata.MustFind(24480)), // Starfire crit +3%
	},
})

// Ruins of Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=494/symbols-of-unending-life
var ItemSetSymbolsOfUnendingLife = core.NewItemSet(core.ItemSet{
	ID:   494,
	Name: "Symbols of Unending Life",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotBack,
		proto.ItemSlot_ItemSlotFinger1,
		proto.ItemSlot_ItemSlotFinger2,
		proto.ItemSlot_ItemSlotMainHand,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		// TODO: Ingame test needed -- whether the 30 energy adds to the usual refund on Rip and Ferocious Bite.
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			// Your finishing moves now refund 30 energy on a Miss, Dodge, Block, or Parry.
			//
			// 26107 hears every melee ability and names none of the druid's spells, so the listener
			// keeps to the finishers, Rip and Ferocious Bite, and to the outcomes the tooltip names.
			// 26976's energy comes on top of the finisher's own miss refund.
			character := agent.GetCharacter()
			row := spelldata.MustFind(26107)
			setBonusAura.ExposeToAPL(row.ID)
			trigger := spelldata.ProcTrigger(character, row, nil)
			trigger.Handler = shared.FiredSpellHandler(character, trigger.ActionID, row)
			if trigger.Handler == nil {
				// The wearer has no energy bar.
				return
			}
			trigger.Name = "Symbols of Unending Life - 3PC"
			trigger.Outcome = core.OutcomeMiss | core.OutcomeDodge | core.OutcomeBlock | core.OutcomeParry
			trigger.RequireDamageDealt = false
			trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				return spellData.Rip.Contains(spell) || spellData.FerociousBite.Contains(spell)
			}

			setBonusAura.AttachProcTrigger(trigger)
		},
	},
})

// The PvP sets. Every bonus is a plain stat but movement speed in the animal forms (23218), which has
// no place in the sim.
var (
	pvpStamina  = spelldata.MustFind(14467)  // +20 Stamina
	pvpRefuge   = spelldata.MustFind(467550) // +44 healing, +15 spell damage
	pvpWildhide = spelldata.MustFind(14047)  // +23 spell damage and healing
)

var ItemSetChampionsRefuge = core.NewItemSet(core.ItemSet{
	ID:    539,
	Name:  "Champion's Refuge",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpRefuge),
		6: parsedSetBonus(pvpStamina),
	},
})

var ItemSetLieutenantCommandersRefuge = core.NewItemSet(core.ItemSet{
	ID:    551,
	Name:  "Lieutenant Commander's Refuge",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpRefuge),
		6: parsedSetBonus(pvpStamina),
	},
})

var ItemSetWarlordsRefuge = core.NewItemSet(core.ItemSet{
	ID:    2076,
	Name:  "Warlord's Refuge",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpStamina),
		6: parsedSetBonus(pvpRefuge),
	},
})

var ItemSetFieldMarshalsRefuge = core.NewItemSet(core.ItemSet{
	ID:    2077,
	Name:  "Field Marshal's Refuge",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpStamina),
		6: parsedSetBonus(pvpRefuge),
	},
})

var ItemSetChampionsWildhide = core.NewItemSet(core.ItemSet{
	ID:    2080,
	Name:  "Champion's Wildhide",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpWildhide),
		6: parsedSetBonus(pvpStamina),
	},
})

var ItemSetLieutenantCommandersWildhide = core.NewItemSet(core.ItemSet{
	ID:    2081,
	Name:  "Lieutenant Commander's Wildhide",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpWildhide),
		6: parsedSetBonus(pvpStamina),
	},
})

var ItemSetWarlordsWildhide = core.NewItemSet(core.ItemSet{
	ID:    2079,
	Name:  "Warlord's Wildhide",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpStamina),
		6: parsedSetBonus(pvpWildhide),
	},
})

var ItemSetFieldMarshalsWildhide = core.NewItemSet(core.ItemSet{
	ID:    2078,
	Name:  "Field Marshal's Wildhide",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpStamina),
		6: parsedSetBonus(pvpWildhide),
	},
})

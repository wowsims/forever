package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func parsedSetBonus(rows ...*spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		for _, row := range rows {
			spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
		}
	}
}

// TODO: Not registered: the Feral and Guardian T1 sets, 2115 (Grovekeeper Ferocity) and 2113 (Grovekeeper
// Rage), until their pieces state stats. Their four shipped pieces, like Grovekeeper Eclipse's, state armor
// and no other stat. The Symbols of Unending Life 3-piece bonus lacks full game files data.

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
			// A_MOD_DAMAGE_DONE_CREATURE has no parse row, so the amount and the creature type
			// mask come off the row by hand. Only the spell damage half is modelled.
			row := spelldata.MustFind(1301073)
			setBonusAura.
				AttachMobTypeBonusStats(stats.Stats{stats.SpellDamage: row.EffectN(1).BaseValue()}, spelldata.MobTypes(row.EffectN(1).Misc)...).
				ExposeToAPL(row.ID)
		},
		5: parsedSetBonus(spelldata.MustFind(1301242)), // Insect Swarm lasts 3 sec longer
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

// Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=493/genesis-raiment
//
// The 5-piece bonus takes 10 minutes off Rebirth's cooldown, which the sim does not cast.
var ItemSetGenesisRaiment = core.NewItemSet(core.ItemSet{
	ID:   493,
	Name: "Genesis Raiment",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		3: parsedSetBonus(spelldata.MustFind(18196), spelldata.MustFind(15693)), // +15 defense, +150 armor
	},
})

// The PvP sets. Every bonus is a plain stat but movement speed in the animal forms (23218), which has
// no place in the sim.
var (
	pvpStamina   = spelldata.MustFind(14467)  // +20 Stamina
	pvpRefuge    = spelldata.MustFind(467550) // +44 healing, +15 spell damage
	pvpWildhide  = spelldata.MustFind(14047)  // +23 spell damage and healing
	pvpSanctuary = spelldata.MustFind(14049)  // +40 attack power
)

var ItemSetLieutenantCommandersSanctuary = core.NewItemSet(core.ItemSet{
	ID:    381,
	Name:  "Lieutenant Commander's Sanctuary",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpSanctuary),
		6: parsedSetBonus(pvpStamina),
	},
})

var ItemSetChampionsSanctuary = core.NewItemSet(core.ItemSet{
	ID:    382,
	Name:  "Champion's Sanctuary",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpSanctuary),
		6: parsedSetBonus(pvpStamina),
	},
})

var ItemSetFieldMarshalsSanctuary = core.NewItemSet(core.ItemSet{
	ID:    397,
	Name:  "Field Marshal's Sanctuary",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpStamina),
		6: parsedSetBonus(pvpSanctuary),
	},
})

var ItemSetWarlordsSanctuary = core.NewItemSet(core.ItemSet{
	ID:    398,
	Name:  "Warlord's Sanctuary",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(pvpStamina),
		6: parsedSetBonus(pvpSanctuary),
	},
})

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

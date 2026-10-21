package warlock

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The rows ItemSetSpell names for each set's thresholds. None of them is a warlock class spell, so the
// generated table carries no reference to them.
var (
	pvpSpellDamage    = spelldata.MustFind(14047)   // spell damage
	pvpImmolateCast   = spelldata.MustFind(23047)   // Immolate's cast time and global cooldown
	pvpStamina        = spelldata.MustFind(14467)   // Stamina
	demoniacDamage    = spelldata.MustFind(9417)    // 2pc: spell damage
	demoniacCorrupt   = spelldata.MustFind(24486)   // 3pc: Corruption's damage
	demoniacDeathCoil = spelldata.MustFind(24487)   // 5pc: Death Coil's cooldown
	doomcallerImmo    = spelldata.MustFind(26116)   // 3pc: Immolate's damage
	doomcallerBolt    = spelldata.MustFind(26117)   // 5pc: Shadow Bolt's cost
	unspokenPets      = spelldata.MustFind(26118)   // 3pc: the demons' damage
	demonheartHit     = spelldata.MustFind(1300966) // 2pc: hit
	demonheartDemons  = spelldata.MustFind(1301094) // 4pc: spell damage against Demons
)

// The six slots of a cloth PvP set, and of Demonheart Raiment.
var sixPieceSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// A bonus that is the row's parse and nothing else.
func parsedSetBonus(row *spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
	}
}

// The PvP sets share their three rows; the sets differ only in the piece counts they grant them at.
func pvpSetBonuses(spellDamageAt, immolateAt, staminaAt int32) map[int32]core.ApplySetBonus {
	return map[int32]core.ApplySetBonus{
		spellDamageAt: parsedSetBonus(pvpSpellDamage),
		immolateAt:    parsedSetBonus(pvpImmolateCast),
		staminaAt:     parsedSetBonus(pvpStamina),
	}
}

var ItemSetChampionsThreads = core.NewItemSet(core.ItemSet{
	ID:      345,
	Name:    "Champion's Threads",
	Slots:   sixPieceSlots,
	Bonuses: pvpSetBonuses(2, 4, 6),
})

var ItemSetLieutenantCommandersThreads = core.NewItemSet(core.ItemSet{
	ID:      346,
	Name:    "Lieutenant Commander's Threads",
	Slots:   sixPieceSlots,
	Bonuses: pvpSetBonuses(2, 4, 6),
})

var ItemSetWarlordsThreads = core.NewItemSet(core.ItemSet{
	ID:      391,
	Name:    "Warlord's Threads",
	Slots:   sixPieceSlots,
	Bonuses: pvpSetBonuses(6, 3, 2),
})

var ItemSetFieldMarshalsThreads = core.NewItemSet(core.ItemSet{
	ID:      392,
	Name:    "Field Marshal's Threads",
	Slots:   sixPieceSlots,
	Bonuses: pvpSetBonuses(6, 3, 2),
})

var ItemSetChampionsDreadgear = core.NewItemSet(core.ItemSet{
	ID:      541,
	Name:    "Champion's Dreadgear",
	Slots:   sixPieceSlots,
	Bonuses: pvpSetBonuses(2, 4, 6),
})

var ItemSetLieutenantCommandersDreadgear = core.NewItemSet(core.ItemSet{
	ID:      547,
	Name:    "Lieutenant Commander's Dreadgear",
	Slots:   sixPieceSlots,
	Bonuses: pvpSetBonuses(2, 4, 6),
})

// Zul'Gurub
var ItemSetDemoniacsThreads = core.NewItemSet(core.ItemSet{
	ID:   481,
	Name: "Demoniac's Threads",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotNeck,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotWrist,
		proto.ItemSlot_ItemSlotTrinket1,
		proto.ItemSlot_ItemSlotTrinket2,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(demoniacDamage),
		3: parsedSetBonus(demoniacCorrupt),
		5: parsedSetBonus(demoniacDeathCoil),
	},
})

// Ahn'Qiraj
var ItemSetDoomcallersAttire = core.NewItemSet(core.ItemSet{
	ID:   499,
	Name: "Doomcaller's Attire",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		3: parsedSetBonus(doomcallerImmo),
		5: parsedSetBonus(doomcallerBolt),
	},
})

// Ahn'Qiraj
var ItemSetImplementsOfUnspokenNames = core.NewItemSet(core.ItemSet{
	ID:   500,
	Name: "Implements of Unspoken Names",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotBack,
		proto.ItemSlot_ItemSlotFinger1,
		proto.ItemSlot_ItemSlotFinger2,
		proto.ItemSlot_ItemSlotMainHand,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		// Effect 2 raises the demons' passive, which raises all they deal, by the tooltip's 5%. The store
		// does not carry the passive, so the share is the demons' damage multiplier.
		// TODO: Effect 1 raises Firebolt and Lash of Pain by another 5%, which the tooltip does not
		// state; left out until it is confirmed in game that the two stack.
		3: func(agent core.Agent, _ *core.Aura) {
			warlock := agent.(WarlockAgent).GetWarlock()
			petMult := 1 + unspokenPets.EffectN(2).Percent()
			for _, pet := range warlock.demons() {
				pet.PseudoStats.DamageDealtMultiplier *= petMult
			}
		},
	},
})

var (
	plagueCorruption = spelldata.MustFind(28829) // 4pc: Corruption's damage
	plagueThreat     = spelldata.MustFind(28746) // 6pc: threat
)

// Tier 3
var ItemSetPlagueheartRaiment = core.NewItemSet(core.ItemSet{
	ID:    529,
	Name:  "Plagueheart Raiment",
	Slots: shared.TierThreeSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		// 2pc (28831) heals the wearer off Shadow Bolt hits, which does not affect sims.
		4: shared.ParsedSetBonus(plagueCorruption),
		// Effect 2 lowers the threat of Corruption, Immolate, Curse of Agony and Siphon Life. Effect 1,
		// 25% less threat from spell crits, has no place in the sim's threat.
		6: shared.ParsedSetBonus(plagueThreat),
		// 8pc (28830) takes 12% off Life Tap's health cost, which does not affect sims.
	},
})

// Tier 1
// The ItemSet row lists six pieces, head and chest among them, but the game files ship only the
// shoulders, hands, legs and feet, so the 5-piece bonus cannot be worn yet. The 3-piece bonus
// shortens Banish, which the sim does not cast.
var ItemSetDemonheartRaiment = core.NewItemSet(core.ItemSet{
	ID:    2100,
	Name:  "Demonheart Raiment",
	Slots: sixPieceSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(demonheartHit),
		// A_MOD_DAMAGE_DONE_CREATURE has no row in the parse table, so the bonus is the spell damage
		// the sim keeps against the mob types the effect's mask names.
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			warlock := agent.(WarlockAgent).GetWarlock()
			effect := demonheartDemons.EffectN(1)
			bonus := stats.Stats{stats.SpellDamage: effect.BaseValue()}
			mobTypes := spelldata.MobTypes(effect.Misc)

			setBonusAura.
				ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
					warlock.AddMobTypeBonusStats(bonus, mobTypes...)
				}).
				ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
					warlock.AddMobTypeBonusStats(bonus.Invert(), mobTypes...)
				})
		},
		// TODO: 5pc (1301715) is a dummy holding Life Tap's 20% more mana at no more health, which Life
		// Tap does not read yet.
	},
})

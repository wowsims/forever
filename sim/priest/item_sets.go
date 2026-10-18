package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A bonus whose row the parse states whole, exposed to the APL under the row's id.
func parsedSetBonus(id int32) core.ApplySetBonus {
	row := spelldata.MustFind(id)
	return func(agent core.Agent, setBonusAura *core.Aura) {
		spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
		setBonusAura.ExposeToAPL(row.ID)
	}
}

// The six slots of a Conviction set and of a PvP set. A Conviction ItemSet row lists six pieces the
// game files do not ship; the shoulders, hands, legs and feet ship under other ids, and no head or
// chest piece ships.
var sixPieceSetSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// Zul'Gurub
// https://www.wowhead.com/forever/item-set=480/confessors-raiment
//
// The 5-piece bonus cuts Mind Control's cast time, which the sim does not cast.
var ItemSetConfessorsRaiment = core.NewItemSet(core.ItemSet{
	ID:   480,
	Name: "Confessor's Raiment",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotNeck,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotWrist,
		proto.ItemSlot_ItemSlotWaist,
		proto.ItemSlot_ItemSlotTrinket1,
		proto.ItemSlot_ItemSlotTrinket2,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(9408),  // Increases healing done by spells and effects by up to 22.
		3: parsedSetBonus(24482), // Increase the range of your Smite and Holy Fire spells by 5 yds.
	},
})

// Ruins of Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=508/finery-of-infinite-wisdom
var ItemSetFineryOfInfiniteWisdom = core.NewItemSet(core.ItemSet{
	ID:   508,
	Name: "Finery of Infinite Wisdom",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotMainHand,
		proto.ItemSlot_ItemSlotFinger1,
		proto.ItemSlot_ItemSlotFinger2,
		proto.ItemSlot_ItemSlotBack,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		// Increases the damage of your Shadow Word: Pain spell by 5%. The row's class flags name
		// Shadow Word: Pain alone.
		3: parsedSetBonus(26172),
	},
})

// Temple of Ahn'Qiraj
// https://www.wowhead.com/forever/item-set=507/garments-of-the-oracle
//
// The 3-piece bonus (26169) heals the priest for a share of the heals cast on others, which the sim
// does not model.
var ItemSetGarmentsOfTheOracle = core.NewItemSet(core.ItemSet{
	ID:   507,
	Name: "Garments of the Oracle",
	Slots: []proto.ItemSlot{
		proto.ItemSlot_ItemSlotHead,
		proto.ItemSlot_ItemSlotShoulder,
		proto.ItemSlot_ItemSlotChest,
		proto.ItemSlot_ItemSlotLegs,
		proto.ItemSlot_ItemSlotFeet,
	},
	Bonuses: map[int32]core.ApplySetBonus{
		5: parsedSetBonus(26171), // Increases the duration of your Renew spell by 3 sec.
	},
})

var convictionUndeadDamage = spelldata.MustFind(1301087) // Shadow 4pc: spell damage against Undead

// Tier 1 - Discipline and Holy
// https://www.wowhead.com/forever/item-set=2104/vestments-of-conviction
//
// The 3-piece bonus takes 60 sec off Fear Ward's cooldown, which the sim does not cast.
var ItemSetVestmentsOfConviction = core.NewItemSet(core.ItemSet{
	ID:    2104,
	Name:  "Vestments of Conviction",
	Slots: sixPieceSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(1300954), // Increased Spirit +10.
		4: parsedSetBonus(1301085), // Increases healing done by up to 26 and damage done by up to 9 for all magical spells and effects.
		5: parsedSetBonus(1301703), // Reduces the cooldown of your Penance and Prayer of Mending spells by 1 sec.
	},
})

// Tier 1 - Shadow
// https://www.wowhead.com/forever/item-set=2105/raiments-of-conviction
//
// The 3-piece bonus lengthens Shackle Undead, which the sim does not cast.
var ItemSetRaimentsOfConviction = core.NewItemSet(core.ItemSet{
	ID:    2105,
	Name:  "Raiments of Conviction",
	Slots: sixPieceSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: parsedSetBonus(1300955), // Improves your chance to hit by 1%.
		4: func(_ core.Agent, setBonusAura *core.Aura) {
			// Increases damage done by your spells and effects by up to 21 when fighting Undead.
			// A_MOD_DAMAGE_DONE_CREATURE has no row in the parse table, so the amount and the
			// creature types it names come off the effect by hand.
			effect := convictionUndeadDamage.EffectN(1)
			setBonusAura.
				AttachMobTypeBonusStats(stats.Stats{stats.SpellDamage: effect.BaseValue()}, spelldata.MobTypes(effect.Misc)...).
				ExposeToAPL(convictionUndeadDamage.ID)
		},
		5: parsedSetBonus(1301706), // Reduces the cooldown of your Devouring Plague spell by 60 sec.
	},
})

// The PvP sets. Each states a spell damage and healing row and a Stamina row, at thresholds that
// differ by set, and a bonus to Psychic Scream's duration (23044), which the sim does not cast.
func init() {
	for _, set := range []struct {
		id                 int32
		name               string
		twoPiece, sixPiece int32
	}{
		{342, "Champion's Raiment", 14047, 14462},
		{344, "Lieutenant Commander's Raiment", 14047, 14467},
		{389, "Field Marshal's Raiment", 14467, 14047},
		{390, "Warlord's Raiment", 14467, 14047},
		{540, "Champion's Investiture", 14047, 14467},
		{549, "Champion's Raiment", 14047, 14467},
		{2072, "Champion's Investiture", 467550, 14467},
		{2073, "Warlord's Investiture", 14467, 467550},
		{2074, "Field Marshal's Investiture", 14467, 467550},
		{2075, "Lieutenant Commander's Investiture", 467550, 14467},
	} {
		core.NewItemSet(core.ItemSet{ID: set.id, Name: set.name, Slots: sixPieceSetSlots, Bonuses: map[int32]core.ApplySetBonus{
			2: parsedSetBonus(set.twoPiece),
			6: parsedSetBonus(set.sixPiece),
		}})
	}
}

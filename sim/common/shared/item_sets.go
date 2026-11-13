package shared

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A bonus whose rows the parse states whole, exposed to the APL under the first row's id. A row's spell
// modifiers name its class's spells by class flags, so on another class they reach nothing.
func ParsedSetBonus(rows ...*spelldata.Spell) core.ApplySetBonus {
	return func(agent core.Agent, setBonusAura *core.Aura) {
		for _, row := range rows {
			spelldata.ParseEffects(agent.GetCharacter(), setBonusAura, row)
		}
		setBonusAura.ExposeToAPL(rows[0].ID)
	}
}

// A set gen_db registers from the game files, each piece count's bonus parsed from its rows. A set
// written by hand under the same id keeps it off.
func NewParsedItemSet(id int32, name string, slots []proto.ItemSlot, bonuses map[int32][]int32) {
	NewParsedProfessionItemSet(id, name, proto.Profession_ProfessionUnknown, slots, bonuses)
}

// A generated set whose ItemSet row gates the bonuses on a profession (RequiredSkill).
func NewParsedProfessionItemSet(id int32, name string, profession proto.Profession, slots []proto.ItemSlot, bonuses map[int32][]int32) {
	if slices.ContainsFunc(core.ItemSets(), func(set *core.ItemSet) bool { return set.ID == id }) {
		return
	}

	set := core.ItemSet{ID: id, Name: name, Slots: slots, RequiredProfession: profession, Bonuses: map[int32]core.ApplySetBonus{}}
	for pieces, spellIDs := range bonuses {
		set.Bonuses[pieces] = ParsedSetBonus(core.MapSlice(spellIDs, spelldata.MustFind)...)
	}
	core.NewItemSet(set)
}

// The eight armor slots of a crafted or dungeon set.
var EightPieceSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead,
	proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest,
	proto.ItemSlot_ItemSlotWrist,
	proto.ItemSlot_ItemSlotHands,
	proto.ItemSlot_ItemSlotWaist,
	proto.ItemSlot_ItemSlotLegs,
	proto.ItemSlot_ItemSlotFeet,
}

// A Naxxramas tier 3 set: the eight armor pieces and the ring, all worn only by the set's class.
var TierThreeSlots = slices.Concat(EightPieceSlots, []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2})

package dbc

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

func TestAddAreaStatsKeepsOneEntryPerArea(t *testing.T) {
	ap := int32(proto.Stat_StatAttackPower)
	sp := int32(proto.Stat_StatSpellDamage)
	source := map[int32]float64{ap: 29}
	opts := &proto.ScalingItemProperties{Stats: map[int32]float64{ap: 15}}

	addAreaStats(opts, proto.AreaType_AreaTypeForestGrassland, source)
	addAreaStats(opts, proto.AreaType_AreaTypeForestGrassland, map[int32]float64{ap: 1, sp: 8})
	addAreaStats(opts, proto.AreaType_AreaTypeMountainous, map[int32]float64{sp: 6})

	if got := opts.Stats[ap]; got != 15 {
		t.Errorf("the unconditional stats changed: attack power = %v, want 15", got)
	}
	if len(opts.AreaStats) != 2 {
		t.Fatalf("got %d area entries, want one per area", len(opts.AreaStats))
	}
	forest, mountain := opts.AreaStats[0], opts.AreaStats[1]
	if forest.AreaType != proto.AreaType_AreaTypeForestGrassland || forest.Stats[ap] != 30 || forest.Stats[sp] != 8 {
		t.Errorf("forest entry = %v, want attack power 30 and spell damage 8", forest)
	}
	if mountain.AreaType != proto.AreaType_AreaTypeMountainous || mountain.Stats[sp] != 6 {
		t.Errorf("mountain entry = %v, want spell damage 6", mountain)
	}
	if source[ap] != 29 {
		t.Errorf("the caller's map was written through: %v", source)
	}
}

// An equip spell raising only a weapon skill the item's own stats state leaves the item; the same spell on
// an item whose stats state no such skill, or a spell doing more than raise it, stays.
func TestEquipSpellRestatingAStatWeaponSkillLeavesTheItem(t *testing.T) {
	const fistLine, fistSpell, fistAndArmorSpell = 162, 1, 2
	skill := func(line int) SpellEffect {
		return SpellEffect{EffectAura: dbcenums.A_MOD_SKILL, EffectMiscValues: []int{line, 0}}
	}
	d := NewDBC()
	d.SpellEffects[fistSpell] = map[int]SpellEffect{0: skill(fistLine)}
	d.SpellEffects[fistAndArmorSpell] = map[int]SpellEffect{0: skill(fistLine), 1: {EffectAura: dbcenums.A_MOD_RESISTANCE}}

	for _, tc := range []struct {
		name  string
		stats []int
		spell int
		stays bool
	}{
		{"the stat and the spell on the same skill", []int{ITEM_MOD_FIST_WEAPONS_SKILL}, fistSpell, false},
		{"the spell with no weapon skill stat", []int{ITEM_MOD_STAMINA}, fistSpell, true},
		{"the stat on another skill", []int{ITEM_MOD_DAGGERS_SKILL}, fistSpell, true},
		{"a spell that does more than raise the skill", []int{ITEM_MOD_FIST_WEAPONS_SKILL}, fistAndArmorSpell, true},
	} {
		const itemID = 10
		d.Items[itemID] = Item{Id: itemID, BonusStat: tc.stats}
		d.ItemEffectsByParentID[itemID] = []ItemEffect{{SpellID: tc.spell, TriggerType: ITEM_SPELLTRIGGER_ON_EQUIP, ParentItemID: itemID}}
		d.dropRestatedWeaponSkills()
		if stays := len(d.ItemEffectsByParentID[itemID]) == 1; stays != tc.stays {
			t.Errorf("%s: the equip spell stays %v, want %v", tc.name, stays, tc.stays)
		}
	}
}

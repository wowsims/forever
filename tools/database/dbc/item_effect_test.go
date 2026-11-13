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

// An equip spell raising only a weapon skill or a stat against a creature type the item's own stats state
// leaves the item; the same spell on an item whose stats state no such bonus, or a spell doing more, stays.
func TestEquipSpellRestatingAnItemStatLeavesTheItem(t *testing.T) {
	const fistLine, undeadMask, demonAndUndeadMask = 162, 32, 36
	const fistSpell, fistAndArmorSpell, undeadAttackPowerSpell, undeadSpellDamageSpell, demonAndUndeadSpell = 1, 2, 3, 4, 5
	skill := func(line int) SpellEffect {
		return SpellEffect{EffectAura: dbcenums.A_MOD_SKILL, EffectMiscValues: []int{line, 0}}
	}
	versus := func(aura EffectAuraType, mask int) SpellEffect {
		return SpellEffect{EffectAura: aura, EffectMiscValues: []int{mask, 0}}
	}
	d := NewDBC()
	d.SpellEffects[fistSpell] = map[int]SpellEffect{0: skill(fistLine)}
	d.SpellEffects[fistAndArmorSpell] = map[int]SpellEffect{0: skill(fistLine), 1: {EffectAura: dbcenums.A_MOD_RESISTANCE}}
	d.SpellEffects[undeadAttackPowerSpell] = map[int]SpellEffect{
		0: versus(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, undeadMask),
		1: versus(dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS, undeadMask),
	}
	d.SpellEffects[undeadSpellDamageSpell] = map[int]SpellEffect{0: versus(dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS, undeadMask)}
	d.SpellEffects[demonAndUndeadSpell] = map[int]SpellEffect{0: versus(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, demonAndUndeadMask)}

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
		{"attack power against the stat's creature type", []int{ITEM_MOD_ATTACK_POWER_VS_UNDEAD}, undeadAttackPowerSpell, false},
		{"spell damage against the stat's creature type", []int{ITEM_MOD_SPELL_DAMAGE_VS_UNDEAD}, undeadSpellDamageSpell, false},
		{"attack power against another creature type", []int{ITEM_MOD_ATTACK_POWER_VS_DEMON}, undeadAttackPowerSpell, true},
		{"spell damage where the stat gives attack power", []int{ITEM_MOD_ATTACK_POWER_VS_UNDEAD}, undeadSpellDamageSpell, true},
		{"a spell against more creature types than the stats", []int{ITEM_MOD_ATTACK_POWER_VS_UNDEAD}, demonAndUndeadSpell, true},
		{"a spell against every creature type the stats state", []int{ITEM_MOD_ATTACK_POWER_VS_DEMON, ITEM_MOD_ATTACK_POWER_VS_UNDEAD}, demonAndUndeadSpell, false},
	} {
		const itemID = 10
		d.Items[itemID] = Item{Id: itemID, BonusStat: tc.stats}
		d.ItemEffectsByParentID[itemID] = []ItemEffect{{SpellID: tc.spell, TriggerType: ITEM_SPELLTRIGGER_ON_EQUIP, ParentItemID: itemID}}
		d.dropRestatedItemStats()
		if stays := len(d.ItemEffectsByParentID[itemID]) == 1; stays != tc.stays {
			t.Errorf("%s: the equip spell stays %v, want %v", tc.name, stays, tc.stays)
		}
	}
}

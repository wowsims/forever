package database

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// Test items and spells: a use that summons, an equip aura whose trigger summons, a use written by hand, a use that
// summons a totem, and a use whose summon the test adds to overrides.NotGuardianSummons.
const (
	summonUseItem, summonUseSpell                     = 990001, 990101
	summonProcItem, summonProcAura, summonProcTrigger = 990002, 990102, 990112
	summonHandItem, summonHandSpell                   = 990003, 990103
	summonTotemItem, summonTotemSpell                 = 990004, 990104
	summonExcludedItem, summonExcludedSpell           = 990005, 990105
)

func summonEffect(creature, properties int, count float64) map[int]dbc.SpellEffect {
	return map[int]dbc.SpellEffect{0: {EffectType: dbcenums.E_SUMMON, EffectMiscValues: []int{creature, properties}, EffectBasePoints: count}}
}

func summonInputs(t *testing.T) (*dbc.DBC, *WowDatabase) {
	overrides.NotGuardianSummons[summonExcludedSpell] = "a test summon that is no guardian"
	t.Cleanup(func() { delete(overrides.NotGuardianSummons, summonExcludedSpell) })
	instance := &dbc.DBC{
		ItemEffectsByParentID: map[int][]dbc.ItemEffect{
			summonUseItem:      {{TriggerType: dbc.ITEM_SPELLTRIGGER_ON_USE, SpellID: summonUseSpell}},
			summonProcItem:     {{TriggerType: dbc.ITEM_SPELLTRIGGER_ON_EQUIP, SpellID: summonProcAura}},
			summonHandItem:     {{TriggerType: dbc.ITEM_SPELLTRIGGER_ON_USE, SpellID: summonHandSpell}},
			summonTotemItem:    {{TriggerType: dbc.ITEM_SPELLTRIGGER_ON_USE, SpellID: summonTotemSpell}},
			summonExcludedItem: {{TriggerType: dbc.ITEM_SPELLTRIGGER_ON_USE, SpellID: summonExcludedSpell}},
		},
		SpellEffects: map[int]map[int]dbc.SpellEffect{
			summonUseSpell:      summonEffect(800001, 6853, 3),
			summonProcAura:      {0: {EffectType: dbcenums.E_APPLY_AURA, EffectTriggerSpell: summonProcTrigger}},
			summonProcTrigger:   summonEffect(800002, 6792, 1),
			summonHandSpell:     summonEffect(800003, 6698, 1),
			summonTotemSpell:    summonEffect(800004, int(totemSlotSummonProperties[2]), 5),
			summonExcludedSpell: summonEffect(800005, 6812, 1),
		},
	}
	db := NewWowDatabase()
	for _, id := range []int32{summonUseItem, summonProcItem, summonHandItem, summonTotemItem, summonExcludedItem} {
		db.Items[id] = &proto.UIItem{Id: id, Name: "Test Summoner"}
	}
	return instance, db
}

// A use, and an equip spell's trigger, each reach their summon's creature and count; a hand-written item is marked,
// and a totem and a summon NotGuardianSummons names stay out.
func TestItemSummonsFindsEveryGuardianSummon(t *testing.T) {
	instance, db := summonInputs(t)
	summons := itemSummons(instance, db, func(id int32) bool { return id == summonHandItem })

	want := []itemSummon{
		{itemID: summonUseItem, name: "Test Summoner", onUse: true, spellID: summonUseSpell, summonID: summonUseSpell, creature: 800001, count: 3},
		{itemID: summonProcItem, name: "Test Summoner", spellID: summonProcAura, summonID: summonProcTrigger, creature: 800002, count: 1},
		{itemID: summonHandItem, name: "Test Summoner", onUse: true, spellID: summonHandSpell, summonID: summonHandSpell, creature: 800003, count: 1, byHand: true},
	}
	if len(summons) != len(want) {
		t.Fatalf("found %d summons, want %d: %+v", len(summons), len(want), summons)
	}
	for i := range want {
		if summons[i] != want[i] {
			t.Errorf("summon %d is %+v, want %+v", i, summons[i], want[i])
		}
	}
}

// The file registers a use the spell store carries, and writes a use it does not carry, a proc and a hand-written
// summon as comments.
func TestRenderItemGuardiansRegistersOnlyTheUsesTheStoreCarries(t *testing.T) {
	instance, db := summonInputs(t)
	summons := itemSummons(instance, db, func(id int32) bool { return id == summonHandItem })
	stored := renderItemGuardians(summons, func(int32) bool { return true })
	missing := renderItemGuardians(summons, func(int32) bool { return false })

	for _, tc := range []struct {
		text, line string
	}{
		{stored, "\t\t{990001, 990101}, // Test Summoner: 990101 summons 3 × creature 800001\n"},
		{missing, "\t\t// not in the spell store: {990001, 990101}, // Test Summoner: 990101 summons 3 × creature 800001\n"},
		{stored, "\t// 990002 Test Summoner: 990112 summons 1 × creature 800002 on a proc, which needs its trigger wired by hand\n"},
		{stored, "\t// 990003 Test Summoner: 990103 summons 1 × creature 800003, written by hand\n"},
	} {
		if !strings.Contains(tc.text, tc.line) {
			t.Errorf("the file has no line %q:\n%s", tc.line, tc.text)
		}
	}
	if strings.Contains(stored, "{990003,") {
		t.Errorf("the file registers the hand-written item:\n%s", stored)
	}
}

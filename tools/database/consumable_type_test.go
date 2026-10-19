package database

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestAConsumableIsTypedByTheCategoryItsOnUseShares(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name     string
		id       int
		category int32
		want     proto.ConsumableType
	}{
		{"Major Mana Potion, the potion category on the item", 13444, dbc.POTION_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypePotion},
		{"Minor Spellblasting Potion, the potion category on its spell 1251919 alone", 250932, dbc.POTION_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypePotion},
		{"Potion of Demon Slaying, an Elixir-subclass item whose spell 11406 shares the potion category", 9224, dbc.POTION_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypePotion},
		{"Thistle Tea, the conjured category on the item over its spell's own", 7676, dbc.CONJURED_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypeConjured},
		{"Demonic Rune, not a Consumable-class item", 12662, dbc.CONJURED_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypeConjured},
		{"Keg of Thunderbrew, an Other-subclass drink", 1262, dbc.DRINK_SPELL_CATEGORY, proto.ConsumableType_ConsumableTypeFood},
		{"Mageblood Elixir, an Elixir-subclass item on a category that types nothing", 20007, 0, proto.ConsumableType_ConsumableTypeElixir},
		{"Gift of Arthas, an elixir on a category that types nothing", 9088, 0, proto.ConsumableType_ConsumableTypeElixir},
	} {
		catalogued, ok := instance.Consumables[tc.id]
		if !ok {
			t.Fatalf("%s: %d is not in the catalogue", tc.name, tc.id)
		}
		got := catalogued.ToProto()
		if got.Type != tc.want {
			t.Errorf("%s: type %v, want %v", tc.name, got.Type, tc.want)
		}
		if tc.category == 0 {
			if slices.Contains([]int32{dbc.POTION_SPELL_CATEGORY, dbc.CONJURED_SPELL_CATEGORY, dbc.FOOD_SPELL_CATEGORY,
				dbc.DRINK_SPELL_CATEGORY, dbc.EXPLOSIVES_SPELL_CATEGORY}, got.CategoryId) {
				t.Errorf("%s: category %d, want one that types nothing", tc.name, got.CategoryId)
			}
		} else if got.CategoryId != tc.category {
			t.Errorf("%s: category %d, want %d", tc.name, got.CategoryId, tc.category)
		}
	}
}

// The explosive picker offers what shares the explosives' category 24 and routes as damage.
func TestOnlyADamagingExplosiveIsTypedAnExplosive(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name      string
		id        int
		explosive bool
	}{
		{"Goblin Sapper Charge, 13241 deals damage", 10646, true},
		{"Bag of Marbles, category 24, 5917 lowers the target's hit chance", 1191, false},
		{"Large Rope Net, category 24, 8312 roots", 835, false},
		{"EZ-Splode Blasting Charge, a Trade Goods explosive on no category", 221325, false},
		{"Caustic Smog Potion, category 4, 1252525's cloud deals damage", 250951, false},
		{"Dragonfire Potion, category 4, 1252554 deals damage", 250953, false},
	} {
		catalogued, ok := instance.Consumables[tc.id]
		if !ok {
			t.Fatalf("%s: %d is not in the catalogue", tc.name, tc.id)
		}
		got := ShippedConsumable(&catalogued, RouteConsumable(&catalogued, instance), instance).Type
		if (got == proto.ConsumableType_ConsumableTypeExplosive) != tc.explosive {
			t.Errorf("%s: typed %v", tc.name, got)
		}
	}
}

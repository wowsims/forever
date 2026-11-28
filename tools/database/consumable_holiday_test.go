package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestAConsumableTakesItsHoliday(t *testing.T) {
	for _, tc := range []struct {
		name       string
		consumable dbc.Consumable
		holiday    string
	}{
		{"RequiredHoliday 335", dbc.Consumable{Id: 990001, RequiredHoliday: 335}, "Love is in the Air"},
		{"a RequiredHoliday with no known name", dbc.Consumable{Id: 990002, RequiredHoliday: 990}, "Holiday 990"},
		{"no RequiredHoliday is sold all year", dbc.Consumable{Id: 990003}, ""},
		{"Fire-toasted Bun, by HolidayConsumables", dbc.Consumable{Id: 23327}, "Midsummer Fire Festival"},
		{"Toasted Smorc, by HolidayConsumables over its RequiredHoliday", dbc.Consumable{Id: 23211, RequiredHoliday: 335}, "Midsummer Fire Festival"},
	} {
		if got := consumableHoliday(&tc.consumable); got != tc.holiday {
			t.Errorf("%s: holiday %q, want %q", tc.name, got, tc.holiday)
		}
	}

	withDBCInputs(t)
	instance := dbc.GetDBC()
	for _, id := range []int{23327, 23211} {
		if got, want := shippedConsumable(t, instance, id).Holiday, HolidayConsumables[int32(id)]; got != want {
			t.Errorf("%d: shipped with holiday %q, want %q", id, got, want)
		}
	}
}

// Nutritious Food hands the amount of its periodic trigger to Well Fed, which states another
// of its own.
func TestNutritiousFoodListsTheAmountItHandsWellFed(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name string
		id   int
		stat stats.Stat
	}{
		{"Smoked Desert Dumplings, 1248401 to Well Fed 1248422", 20452, stats.Strength},
		{"Swiftstrike Steak, 1248399 to Well Fed 1248420", 250072, stats.Agility},
		{"Lobster Stew, 1249516 to Well Fed 1249520", 13933, stats.SpellDamage},
	} {
		consumable := shippedConsumable(t, instance, tc.id)
		food := spelldata.MustFind(consumable.ItemEffects[0].BuffId)
		trigger := food.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE)
		want := trigger.BaseValue()
		if want == 0 {
			t.Fatalf("%s: %d states no amount it hands on", tc.name, food.ID)
		}
		wellFed := spelldata.MustFind(trigger.TriggerID)
		for i := range wellFed.Effects {
			if own := wellFed.Effects[i].BaseValue(); own == want {
				t.Fatalf("%s: Well Fed %d states the handed amount %v itself, so the test cannot tell the two apart", tc.name, wellFed.ID, own)
			}
		}
		if got := stats.FromProtoArray(consumable.Stats)[tc.stat]; got != want {
			t.Errorf("%s: %v %v, want %v", tc.name, got, tc.stat, want)
		}
	}
}

// A plain trigger keeps the triggered spell's own amounts, and a seated Drink lists nothing, whether
// it is the on-use spell or one a trigger puts up.
func TestATriggeredBuffListsItsOwnStats(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	firstTriggered := func(t *testing.T, row *spelldata.Spell) *spelldata.Spell {
		for i := range row.Effects {
			if e := &row.Effects[i]; e.Type == dbcenums.E_TRIGGER_SPELL {
				return e.Trigger()
			}
		}
		t.Fatalf("%d triggers no spell", row.ID)
		return nil
	}
	for _, tc := range []struct {
		name string
		id   int
		stat stats.Stat
	}{
		{"Blessed Sunfruit, 18124 to 18125", 13810, stats.Strength},
		{"Windblossom Berries, 26030 to 18191", 11950, stats.Stamina},
		{"Blessed Sunfruit Juice, 18140 to 18141", 13813, stats.Spirit},
	} {
		consumable := shippedConsumable(t, instance, tc.id)
		buff := firstTriggered(t, spelldata.MustFind(consumable.ItemEffects[0].BuffId))
		want := buff.FirstAura(dbcenums.A_MOD_STAT).BaseValue()
		if want == 0 {
			t.Fatalf("%s: %d states no stat", tc.name, buff.ID)
		}
		if got := stats.FromProtoArray(consumable.Stats)[tc.stat]; got != want {
			t.Errorf("%s: %v %v, want %v", tc.name, got, tc.stat, want)
		}
	}

	juice := shippedConsumable(t, instance, 13813)
	tea := shippedConsumable(t, instance, 249865)
	for _, tc := range []struct {
		name       string
		consumable *proto.Consumable
		drink      *spelldata.Spell
	}{
		{"Blessed Sunfruit Juice, its own 18140", juice, spelldata.MustFind(juice.ItemEffects[0].BuffId)},
		{"Peace Tea, 1249920 to 430", tea, firstTriggered(t, spelldata.MustFind(tea.ItemEffects[0].BuffId))},
	} {
		if !tc.drink.WhileSeated() || tc.drink.FirstAura(dbcenums.A_MOD_POWER_REGEN).BaseValue() == 0 {
			t.Fatalf("%s: %d is no seated mana restore, so the test cannot tell it is left out", tc.name, tc.drink.ID)
		}
		if mp5 := stats.FromProtoArray(tc.consumable.Stats)[stats.MP5]; mp5 != 0 {
			t.Errorf("%s: %v MP5 from the seated Drink %d, want none", tc.name, mp5, tc.drink.ID)
		}
	}
}

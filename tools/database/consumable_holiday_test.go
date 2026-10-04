package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestAConsumableTakesItsHoliday(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name    string
		id      int
		holiday string
	}{
		{"Dark Desire, RequiredHoliday 335", 22237, "Love is in the Air"},
		{"Very Berry Cream, RequiredHoliday 335", 22238, "Love is in the Air"},
		{"Fire-toasted Bun, by HolidayConsumables", 23327, "Midsummer Fire Festival"},
		{"Toasted Smorc, by HolidayConsumables", 23211, "Midsummer Fire Festival"},
		{"Smoked Desert Dumplings is sold all year", 20452, ""},
	} {
		if got := shippedConsumable(t, instance, tc.id).Holiday; got != tc.holiday {
			t.Errorf("%s: holiday %q, want %q", tc.name, got, tc.holiday)
		}
	}
}

// Nutritious Food hands its amount to Well Fed, whose own row states 1.
func TestNutritiousFoodListsTheAmountItHandsWellFed(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	for _, tc := range []struct {
		name  string
		id    int
		stat  stats.Stat
		value float64
	}{
		{"Smoked Desert Dumplings, 1248401 to Well Fed 1248422", 20452, stats.Strength, 20},
		{"Swiftstrike Steak, 1248399 to Well Fed 1248420", 250072, stats.Agility, 20},
		{"Lobster Stew, 1249516 to Well Fed 1249520", 13933, stats.SpellDamage, 28},
	} {
		if got := stats.FromProtoArray(shippedConsumable(t, instance, tc.id).Stats)[tc.stat]; got != tc.value {
			t.Errorf("%s: %v %v, want %v", tc.name, got, tc.stat, tc.value)
		}
	}
}

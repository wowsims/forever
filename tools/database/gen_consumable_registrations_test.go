package database

import (
	"testing"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

// A temporary enchant's stat slot naming a school's spell penetration pierces that school by the
// slot's points, whichever slot it sits in.
func TestEnchantPiercingReadsTheSchoolSlot(t *testing.T) {
	stat := dbc.ITEM_ENCHANTMENT_STAT
	for _, tc := range []struct {
		name    string
		enchant dbc.TemporaryEnchant
		want    shared.SchoolPiercing
	}{
		{"fire, beside a Stamina slot",
			dbc.TemporaryEnchant{Effects: [3]int{stat, stat}, Args: [3]int{dbc.ITEM_MOD_STAMINA, dbc.ITEM_MOD_FIRE_SPELL_PENETRATION}, Points: [3]int{5, 7}},
			shared.SchoolPiercing{School: stats.SchoolIndexFire, Amount: 7}},
		{"frost, alone",
			dbc.TemporaryEnchant{Effects: [3]int{stat}, Args: [3]int{dbc.ITEM_MOD_FROST_SPELL_PENETRATION}, Points: [3]int{4}},
			shared.SchoolPiercing{School: stats.SchoolIndexFrost, Amount: 4}},
		{"a penetration argument on a slot that is no stat",
			dbc.TemporaryEnchant{Effects: [3]int{dbc.ITEM_ENCHANTMENT_EQUIP_SPELL}, Args: [3]int{dbc.ITEM_MOD_FIRE_SPELL_PENETRATION}, Points: [3]int{9}},
			shared.SchoolPiercing{}},
		{"no penetration slot",
			dbc.TemporaryEnchant{Effects: [3]int{stat}, Args: [3]int{dbc.ITEM_MOD_STAMINA}, Points: [3]int{5}},
			shared.SchoolPiercing{}},
	} {
		if got := enchantPiercing(tc.enchant); got != tc.want {
			t.Errorf("%s: piercing %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

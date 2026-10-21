package dbc

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

// A_MOD_ARMOR_NO_MODS armor on misc 1 is ArmorNoMods, as an equip spell's aura stat and in a buff's
// parsed stats, and none of it is Armor. A mask naming any other school gives no stat.
func TestArmorNoModsAuraWritesArmorNoMods(t *testing.T) {
	if got := ConvertEffectAuraToStatIndex(dbcenums.A_MOD_ARMOR_NO_MODS, 1); got != proto.Stat_StatArmorNoMods {
		t.Errorf("misc 1 names %v, want StatArmorNoMods", got)
	}
	for _, misc := range []int{0, 2, 1 | 4, 126} {
		if got := ConvertEffectAuraToStatIndex(dbcenums.A_MOD_ARMOR_NO_MODS, misc); got != -1 {
			t.Errorf("misc %d names %v, want no stat", misc, got)
		}
	}

	effect := SpellEffect{EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_ARMOR_NO_MODS,
		EffectMiscValues: []int{1, 0}, EffectBasePoints: 300}
	got, ok := effect.ParseStatEffect(false, 0)
	if !ok || got[proto.Stat_StatArmorNoMods] != 300 || got[proto.Stat_StatArmor] != 0 {
		t.Errorf("misc 1 parses to %v ArmorNoMods and %v Armor (resolved %v), want 300 and none",
			got[proto.Stat_StatArmorNoMods], got[proto.Stat_StatArmor], ok)
	}

	effect.EffectMiscValues = []int{1 | 4, 0}
	if got, ok := effect.ParseStatEffect(false, 0); ok {
		t.Errorf("misc %d parses to %v, want no stats", 1|4, got)
	}
}

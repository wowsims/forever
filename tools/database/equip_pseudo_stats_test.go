package database

import (
	"maps"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestEquipSpellPseudoStats(t *testing.T) {
	withDBCInputs(t)
	dbc.GetDBC()

	type pseudo = proto.PseudoStat
	for _, tc := range []struct {
		name    string
		spellID int
		want    map[pseudo]float64
	}{
		{"22780 Biznicks 247x128 Accurascope, hit on bows, guns and crossbows", 22780,
			map[pseudo]float64{proto.PseudoStat_PseudoStatRangedHitPercent: 3}},
		{"24154 Falcon's Call, hit on any weapon", 24154,
			map[pseudo]float64{proto.PseudoStat_PseudoStatMeleeHitPercent: 1, proto.PseudoStat_PseudoStatRangedHitPercent: 1}},
		{"1220596 Enchant Shield - Critical Strike, weapon and spell crit", 1220596,
			map[pseudo]float64{
				proto.PseudoStat_PseudoStatMeleeCritPercent:  1,
				proto.PseudoStat_PseudoStatRangedCritPercent: 1,
				proto.PseudoStat_PseudoStatSpellCritPercent:  1,
			}},
		{"1310308 SAF-T Ultra Precision Scope, all crit on bows, guns and crossbows", 1310308,
			map[pseudo]float64{proto.PseudoStat_PseudoStatRangedCritPercent: 2}},
		{"24156 Presence of Sight, spell hit", 24156,
			map[pseudo]float64{proto.PseudoStat_PseudoStatSpellHitPercent: 1}},
		{"25071 Enchant Cloak - Dodge", 25071,
			map[pseudo]float64{proto.PseudoStat_PseudoStatDodgePercent: 1}},
		{"13690 Enchant Shield - Lesser Block", 13690,
			map[pseudo]float64{proto.PseudoStat_PseudoStatBlockPercent: 2}},
		{"28142 Power of the Guardian, a party aura", 28142, map[pseudo]float64{}},
		{"7217 Weapon Counterweight, melee speed", 7217,
			map[pseudo]float64{proto.PseudoStat_PseudoStatMeleeHastePercent: 3}},
		{"13928 Enchant Gloves - Minor Haste, attack and cast speed", 13928,
			map[pseudo]float64{
				proto.PseudoStat_PseudoStatMeleeHastePercent:  1,
				proto.PseudoStat_PseudoStatRangedHastePercent: 1,
				proto.PseudoStat_PseudoStatSpellHastePercent:  1,
			}},
		{"22841 Arcanum of Rapidity, melee and ranged speed", 22841,
			map[pseudo]float64{
				proto.PseudoStat_PseudoStatMeleeHastePercent:  1,
				proto.PseudoStat_PseudoStatRangedHastePercent: 1,
			}},
		{"1293881 Pendulum of Doom, a zero melee speed aura", 1293881, map[pseudo]float64{}},
		{"22755 Critical Weapon +2, crit on melee weapons", 22755, map[pseudo]float64{}},
		{"430391 Hit Weapon +2, hit on melee weapons", 430391, map[pseudo]float64{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := make([]float64, stats.PseudoStatsLen)
			_, added := dbc.AddEquipSpellStats(&stats.Stats{}, got, tc.spellID)

			want := make([]float64, stats.PseudoStatsLen)
			for pseudoStat, value := range tc.want {
				want[pseudoStat] = value
			}
			if !slices.Equal(got, want) {
				t.Errorf("pseudo stats %v, want %v", got, want)
			}
			if added != (len(tc.want) > 0) {
				t.Errorf("reported added=%v with %d pseudo stats expected", added, len(tc.want))
			}
		})
	}

	t.Run("430391 Hit Weapon +2 held to the off hand's slot, hit in the off hand", func(t *testing.T) {
		got := make([]float64, stats.PseudoStatsLen)
		if _, added := dbc.AddEquipSpellStats(&stats.Stats{}, got, offHandHitWeapon(t)); added {
			t.Errorf("pseudo stats %v, want none", got)
		}
	})
}

// The enchant carries what its equip spell states: 2523 applies 22780.
func TestEnchantCarriesEquipSpellPseudoStats(t *testing.T) {
	withDBCInputs(t)

	for _, enchant := range dbc.GetDBC().Enchants {
		if enchant.EffectId != 2523 {
			continue
		}
		pseudoStats := enchant.ToProto().PseudoStats
		if len(pseudoStats) <= int(proto.PseudoStat_PseudoStatRangedHitPercent) ||
			pseudoStats[proto.PseudoStat_PseudoStatRangedHitPercent] != 3 ||
			pseudoStats[proto.PseudoStat_PseudoStatMeleeHitPercent] != 0 {
			t.Errorf("enchant 2523 pseudo stats %v, want ranged hit 3 alone", pseudoStats)
		}
		return
	}
	t.Fatal("enchant 2523 is not in the enchant inputs")
}

// An equip spell folds into the stats of what carries it when it names no weapon, or only ranged ones
// for a hit or crit aura. Any other weapon requirement is left to run time, whole.
func TestEquipSpellFolds(t *testing.T) {
	withDBCInputs(t)
	dbc.GetDBC()

	for _, tc := range []struct {
		name    string
		spellID int
		want    bool
	}{
		{"24154 Falcon's Call, no weapon requirement", 24154, true},
		{"22780 Biznicks 247x128 Accurascope, bows, guns and crossbows", 22780, true},
		{"1310308 SAF-T Ultra Precision Scope, bows, guns and crossbows", 1310308, true},
		{"22755 Critical Weapon +2, melee weapons", 22755, false},
		{"430391 Hit Weapon +2, melee weapons", 430391, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := dbc.EquipSpellFolds(tc.spellID); got != tc.want {
				t.Errorf("folds %v, want %v", got, tc.want)
			}
		})
	}

	t.Run("430391 Hit Weapon +2 held to the off hand's slot", func(t *testing.T) {
		if dbc.EquipSpellFolds(offHandHitWeapon(t)) {
			t.Error("folds true, want false")
		}
	})

	t.Run("22780 as a hit rating on every channel, bows, guns and crossbows", func(t *testing.T) {
		if dbc.EquipSpellFolds(rangedWeaponHitRating(t)) {
			t.Error("folds true, want false")
		}
	})
}

// Each effect of an equip spell goes to the stats, to ranged attacks alone, or to run time.
func TestEquipEffectFold(t *testing.T) {
	withDBCInputs(t)
	dbc.GetDBC()

	for _, tc := range []struct {
		name    string
		spellID func(t *testing.T) int
		want    dbc.EquipFold
	}{
		{"24154 Falcon's Call, hit on any weapon", fixedSpell(24154), dbc.FoldStats},
		{"22780 Biznicks 247x128 Accurascope, hit on bows, guns and crossbows", fixedSpell(22780), dbc.FoldRanged},
		{"1310308 SAF-T Ultra Precision Scope, all crit on bows, guns and crossbows", fixedSpell(1310308), dbc.FoldRanged},
		{"22755 Critical Weapon +2, crit on melee weapons", fixedSpell(22755), dbc.FoldRuntime},
		{"430391 Hit Weapon +2 held to the off hand's slot", offHandHitWeapon, dbc.FoldRuntime},
		{"22780 as a hit rating on every channel, bows, guns and crossbows", rangedWeaponHitRating, dbc.FoldRuntime},
		{"24154 as a ranged hit rating on any weapon", func(t *testing.T) int {
			return asRatingAura(t, 24154, dbcenums.RATING_MOD_HIT_RANGED)
		}, dbc.FoldRanged},
	} {
		t.Run(tc.name, func(t *testing.T) {
			effect := dbc.GetDBC().SpellEffectsInOrder(tc.spellID(t))[0]
			if got := dbc.EquipEffectFold(effect); got != tc.want {
				t.Errorf("fold %v, want %v", got, tc.want)
			}
		})
	}
}

func fixedSpell(spellID int) func(t *testing.T) int {
	return func(*testing.T) int { return spellID }
}

// spellID with its first effect traded for a combat rating aura on mask, a shape no equip spell the store
// carries states.
func asRatingAura(t *testing.T, spellID int, mask int32) int {
	effect := &spelldata.Find(int32(spellID)).Effects[0]
	aura, misc := effect.Aura, effect.Misc
	effect.Aura, effect.Misc = dbcenums.A_MOD_RATING, mask
	t.Cleanup(func() { effect.Aura, effect.Misc = aura, misc })
	return spellID
}

// A combat rating of ranged bits alone is the ranged percent it converts to.
func TestRangedRatingPercents(t *testing.T) {
	type pseudo = proto.PseudoStat
	for _, tc := range []struct {
		name string
		mask int
		want map[pseudo]float64
	}{
		{"ranged hit", dbcenums.RATING_MOD_HIT_RANGED, map[pseudo]float64{proto.PseudoStat_PseudoStatRangedHitPercent: 20 / core.HitRatingPerHitPercent}},
		{"ranged crit", dbcenums.RATING_MOD_CRIT_RANGED, map[pseudo]float64{proto.PseudoStat_PseudoStatRangedCritPercent: 20 / core.CritRatingPerCritPercent}},
		{"ranged haste", dbcenums.RATING_MOD_HASTE_RANGED, map[pseudo]float64{proto.PseudoStat_PseudoStatRangedHastePercent: 20 / core.HasteRatingPerHastePercent}},
		{"ranged hit and crit", dbcenums.RATING_MOD_HIT_RANGED | dbcenums.RATING_MOD_CRIT_RANGED, map[pseudo]float64{
			proto.PseudoStat_PseudoStatRangedHitPercent:  20 / core.HitRatingPerHitPercent,
			proto.PseudoStat_PseudoStatRangedCritPercent: 20 / core.CritRatingPerCritPercent,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := dbc.RangedRatingPercents(tc.mask, 20); !maps.Equal(got, tc.want) {
				t.Errorf("percents %v, want %v", got, tc.want)
			}
		})
	}
}

// 430391 Hit Weapon +2 with its melee weapons traded for the off hand's inventory type, a requirement
// no equip spell the store carries states.
func offHandHitWeapon(t *testing.T) int {
	row := spelldata.Find(430391)
	subclass, invType := row.EquipSubclass, row.EquipInvType
	row.EquipSubclass, row.EquipInvType = 0, dbcenums.INVTYPE_BIT_WEAPONOFFHAND
	t.Cleanup(func() { row.EquipSubclass, row.EquipInvType = subclass, invType })
	return 430391
}

// 22780 Biznicks 247x128 Accurascope as hit rating on the melee, ranged and spell bits, under the same
// bows, guns and crossbows.
func rangedWeaponHitRating(t *testing.T) int {
	return asRatingAura(t, 22780, dbcenums.RATING_MOD_HIT_MELEE|dbcenums.RATING_MOD_HIT_RANGED|dbcenums.RATING_MOD_HIT_SPELL)
}

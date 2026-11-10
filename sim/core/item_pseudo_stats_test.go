package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// An item's and an enchant's percent PseudoStats reach the wearer as the back-end percent Stats, and
// the dodge they state lands on the same stat dodge rating converts into.
func TestItemAndEnchantPseudoStatsReachTheWearer(t *testing.T) {
	const testItemID, testEnchantID = 990101, 990102

	itemStats := make([]float64, stats.ProtoStatsLen)
	itemStats[stats.DodgeRating] = DodgeRatingPerDodgePercent

	itemPseudoStats := make([]float64, stats.PseudoStatsLen)
	itemPseudoStats[proto.PseudoStat_PseudoStatMeleeHitPercent] = 1
	itemPseudoStats[proto.PseudoStat_PseudoStatRangedHitPercent] = 1
	itemPseudoStats[proto.PseudoStat_PseudoStatDodgePercent] = 2
	itemPseudoStats[proto.PseudoStat_PseudoStatBlockPercent] = 3

	// Shaped like Biznicks 247x128 Accurascope (2523): ranged hit alone, and cut short the way a
	// database array may be.
	enchantPseudoStats := make([]float64, proto.PseudoStat_PseudoStatRangedHitPercent+1)
	enchantPseudoStats[proto.PseudoStat_PseudoStatRangedHitPercent] = 3

	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{
			Id:             testItemID,
			Name:           "Test Pseudo Stat Ranged",
			Type:           proto.ItemType_ItemTypeRanged,
			PseudoStats:    itemPseudoStats,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {Stats: map[int32]float64{int32(stats.DodgeRating): itemStats[stats.DodgeRating]}}},
		}},
		Enchants: []*proto.SimEnchant{{
			EffectId:    testEnchantID,
			Name:        "Test Pseudo Stat Scope",
			PseudoStats: enchantPseudoStats,
		}},
	})

	equipment := ProtoToEquipment(&proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: testItemID, Enchant: testEnchantID}}})
	equipStats := equipment.Stats(proto.Spec_SpecUnknown)

	for _, want := range []struct {
		stat  stats.Stat
		value float64
	}{
		{stats.PhysicalHitPercent, 1},
		{stats.RangedHitPercent, 3},
		{stats.DodgePercent, 2},
		{stats.BlockPercent, 3},
		{stats.DodgeRating, DodgeRatingPerDodgePercent},
	} {
		if got := equipStats[want.stat]; math.Abs(got-want.value) > 1e-9 {
			t.Errorf("%s from the item and its enchant: got %v, want %v", want.stat.StatName(), got, want.value)
		}
	}

	unit := &Unit{StatDependencyManager: stats.NewStatDependencyManager()}
	unit.addUniversalStatDependencies()
	unit.AddStatDependency(stats.Agility, stats.DodgeRating, 1/30.0*DodgeRatingPerDodgePercent)
	unit.FinalizeStatDeps()
	unit.stats = unit.ApplyStatDependencies(equipStats.Add(stats.Stats{stats.Agility: 30}))

	// 2% stated by the item, 1% from its rating and 1% from 30 Agility through the rating.
	if got := unit.GetDodgeFromRating(); math.Abs(got-0.04) > 1e-9 {
		t.Errorf("dodge: got %v, want 0.04", got)
	}
	if got := unit.GetBlockFromRating(); math.Abs(got-0.03) > 1e-9 {
		t.Errorf("block: got %v, want 0.03", got)
	}
}

// A random suffix adds its allocations times the item's random property points over 10000, each
// rounded to the nearest point, on top of the item's own stats, and its percent PseudoStats as they are.
func TestRandomSuffixStatsAddToTheItem(t *testing.T) {
	const testItemID, testSuffixID = 990103, 990104

	suffixStats := make([]float64, stats.ProtoStatsLen)
	suffixStats[stats.Intellect] = 1250    // 0.75 at 6 points.
	suffixStats[stats.Stamina] = 3300      // 1.98.
	suffixStats[stats.SpellDamage] = 47500 // 28.5, which the game shows as 29.
	suffixPseudoStats := make([]float64, stats.PseudoStatsLen)
	suffixPseudoStats[proto.PseudoStat_PseudoStatDodgePercent] = 1

	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{
			Id:   testItemID,
			Name: "Test Random Suffix Ring",
			Type: proto.ItemType_ItemTypeFinger,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {
				RandPropPoints: 6,
				Stats:          map[int32]float64{int32(stats.Intellect): 5},
			}},
		}},
		RandomSuffixes: []*proto.ItemRandomSuffix{{
			Id:          testSuffixID,
			Name:        "of the Test",
			Stats:       suffixStats,
			PseudoStats: suffixPseudoStats,
		}},
	})

	equipment := ProtoToEquipment(&proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: testItemID, RandomSuffix: testSuffixID}}})
	equipStats := equipment.Stats(proto.Spec_SpecUnknown)

	for _, want := range []struct {
		stat  stats.Stat
		value float64
	}{
		{stats.Intellect, 5 + 1},
		{stats.Stamina, 2},
		{stats.SpellDamage, 29},
		{stats.DodgePercent, 1},
	} {
		if got := equipStats[want.stat]; got != want.value {
			t.Errorf("%s: got %v, want %v", want.stat.StatName(), got, want.value)
		}
	}
}

// A random suffix that reduces caster DPS swaps in the damage the item carries for it; any other
// suffix leaves the weapon's own damage.
func TestReducedCasterDpsSuffixSwapsTheWeaponDamage(t *testing.T) {
	const testItemID, reducingSuffixID, plainSuffixID = 990105, 990106, 990107

	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{
			Id:          testItemID,
			Name:        "Test Random Suffix Axe",
			Type:        proto.ItemType_ItemTypeWeapon,
			WeaponType:  proto.WeaponType_WeaponTypeAxe,
			HandType:    proto.HandType_HandTypeTwoHand,
			WeaponSpeed: 3.6,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {
				RandPropPoints:            11,
				WeaponDamageMin:           58,
				WeaponDamageMax:           88,
				UnroundedDps:              20.483225,
				ReducedCasterWeaponDamage: &proto.ReducedCasterWeaponDamage{WeaponDamageMin: 43, WeaponDamageMax: 66, UnroundedDps: 15.231116},
			}},
		}},
		RandomSuffixes: []*proto.ItemRandomSuffix{
			{Id: reducingSuffixID, Name: "of the Physician", Stats: make([]float64, stats.ProtoStatsLen), ReducedCasterDps: true},
			{Id: plainSuffixID, Name: "of the Bear", Stats: make([]float64, stats.ProtoStatsLen)},
		},
	})

	for _, tc := range []struct {
		name     string
		suffix   int32
		min, max float64
		dps      float32
	}{
		{"no suffix", 0, 58, 88, 20.483225},
		{"a suffix that does not reduce caster DPS", plainSuffixID, 58, 88, 20.483225},
		{"a suffix that reduces caster DPS", reducingSuffixID, 43, 66, 15.231116},
	} {
		equipment := ProtoToEquipment(&proto.EquipmentSpec{Items: []*proto.ItemSpec{{Id: testItemID, RandomSuffix: tc.suffix}}})
		item := equipment.MainHand()
		if item.WeaponDamageMin != tc.min || item.WeaponDamageMax != tc.max || item.UnroundedDPS != tc.dps {
			t.Errorf("%s: damage %v-%v at %v DPS, want %v-%v at %v", tc.name, item.WeaponDamageMin, item.WeaponDamageMax, item.UnroundedDPS, tc.min, tc.max, tc.dps)
		}
	}
}

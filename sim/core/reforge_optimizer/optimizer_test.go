//go:build with_db

package reforgeoptimizer

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	protopkg "google.golang.org/protobuf/proto"
)

const (
	agilityWristID int32 = 910221 + iota
	strengthWristID
	agilityFeetID
	noStatsFeetID
	agilityWeaponID
	attackPowerShoulderID
	critShoulderID
	plainArmorEnchantID
	armorNoModsEnchantID
	armorScalingNeckID
	hitRatingWristID
	holyHitWristID
)

const (
	attackPowerShoulderAP = 30
	critShoulderAP        = 26 // Below the plain shoulder's, so the 1% crit has to carry it.
)

func addFixtureEnchants() {
	stat := func(which proto.Stat, value float64) []float64 {
		values := make([]float64, len(proto.Stat_name))
		values[which] = value
		return values
	}
	crit := make([]float64, len(proto.PseudoStat_name))
	crit[proto.PseudoStat_PseudoStatMeleeCritPercent] = 1
	crit[proto.PseudoStat_PseudoStatRangedCritPercent] = 1
	holyHit := make([]float64, len(proto.PseudoStat_name))
	holyHit[proto.PseudoStat_PseudoStatSchoolHitPercentHoly] = 1
	core.AddToDatabase(&proto.SimDatabase{Enchants: []*proto.SimEnchant{
		{EffectId: agilityWristID, Type: proto.ItemType_ItemTypeWrist, Stats: stat(proto.Stat_StatAgility, 7)},
		{EffectId: strengthWristID, Type: proto.ItemType_ItemTypeWrist, Stats: stat(proto.Stat_StatStrength, 7)},
		{EffectId: agilityFeetID, Type: proto.ItemType_ItemTypeFeet, Stats: stat(proto.Stat_StatAgility, 7)},
		{EffectId: noStatsFeetID, Type: proto.ItemType_ItemTypeFeet},
		{EffectId: agilityWeaponID, Type: proto.ItemType_ItemTypeWeapon, Stats: stat(proto.Stat_StatAgility, 15)},
		{EffectId: attackPowerShoulderID, Type: proto.ItemType_ItemTypeShoulder, Stats: stat(proto.Stat_StatAttackPower, attackPowerShoulderAP)},
		{EffectId: critShoulderID, Type: proto.ItemType_ItemTypeShoulder, Stats: stat(proto.Stat_StatAttackPower, critShoulderAP), PseudoStats: crit},
		{EffectId: hitRatingWristID, Type: proto.ItemType_ItemTypeWrist, Stats: stat(proto.Stat_StatHitRating, core.HitRatingPerHitPercent)},
		{EffectId: holyHitWristID, Type: proto.ItemType_ItemTypeWrist, PseudoStats: holyHit},
	}})
}

func enchantRequest(raid *proto.Raid, weights map[proto.Stat]float64, options ...*proto.BulkEnchantSlotOptions) *proto.ReforgeOptimizeRequest {
	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	for stat, w := range weights {
		ep.Stats[stat] = w
	}
	return &proto.ReforgeOptimizeRequest{Raid: raid, PreCapEpWeights: ep, Settings: &proto.ReforgeSettings{}, EnchantOptions: options}
}

func wrist(ids ...int32) *proto.BulkEnchantSlotOptions {
	return &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotWrist, EffectIds: ids}
}

// Makes the player's current Agility an upper bound, so any Agility on top
// of it breaks the cap. A cap the request does not list as an undershoot cap is a target instead.
func capAgilityAtTheSheet(t *testing.T, request *proto.ReforgeOptimizeRequest) {
	t.Helper()
	capAgilityBelowTheSheet(t, request, 0)
}

func capAgilityBelowTheSheet(t *testing.T, request *proto.ReforgeOptimizeRequest, below float64) {
	t.Helper()
	caps := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name))}
	caps.Stats[proto.Stat_StatAgility] = sheetStats(t, request.Raid).Stats[stats.Agility] - below
	request.Settings.StatCaps = caps
	upperBound := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name))}
	upperBound.Stats[proto.Stat_StatAgility] = 1
	request.UndershootCaps = upperBound
}

func optimize(t *testing.T, request *proto.ReforgeOptimizeRequest) *proto.ReforgeOptimizeResult {
	t.Helper()
	result := OptimizeAsync(request, simsignals.CreateSignals())
	if err := result.GetError(); err != nil {
		t.Fatalf("OptimizeAsync: %s", err.GetMessage())
	}
	return result
}

func slotEnchant(result *proto.ReforgeOptimizeResult, slot proto.ItemSlot) int32 {
	return result.GetOptimizedGear().Items[slot].GetEnchant()
}

var agilityOverStrength = map[proto.Stat]float64{proto.Stat_StatAgility: 2, proto.Stat_StatStrength: 1}

func TestOptimizePicksTheHigherEPEnchant(t *testing.T) {
	addFixtureEnchants()
	request := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(strengthWristID, agilityWristID))
	result := optimize(t, request)
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != agilityWristID {
		t.Errorf("wrist enchant %d, want %d (Agility)", got, agilityWristID)
	}
	if result.GetScore() <= 0 {
		t.Errorf("score %v, want above 0", result.GetScore())
	}
	if again := optimize(t, request); !protopkg.Equal(again.GetOptimizedGear(), result.GetOptimizedGear()) {
		t.Errorf("second run gave %v, the first %v", again.GetOptimizedGear(), result.GetOptimizedGear())
	}
}

func TestOptimizeRespectsAStatCap(t *testing.T) {
	addFixtureEnchants()
	request := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(strengthWristID, agilityWristID))
	capAgilityAtTheSheet(t, request)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotWrist); got != strengthWristID {
		t.Errorf("wrist enchant %d, want %d (Strength)", got, strengthWristID)
	}
}

// The Agility enchant breaks the Agility ceiling on the first solve, so a second one picks the
// Strength one.
func TestOptimizeReportsItsSolvePasses(t *testing.T) {
	addFixtureEnchants()
	uncapped := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(strengthWristID, agilityWristID))
	if got := optimize(t, uncapped).GetPassesDone(); got != 1 {
		t.Errorf("uncapped: %d passes, want 1", got)
	}
	capped := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(strengthWristID, agilityWristID))
	capAgilityAtTheSheet(t, capped)
	if got := optimize(t, capped).GetPassesDone(); got <= 1 {
		t.Errorf("capped: %d passes, want more than 1", got)
	}
}

func TestOptimizeLeavesASlotEmptyWhenEveryCandidateBreaksACap(t *testing.T) {
	addFixtureEnchants()
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotWrist].Enchant = strengthWristID
	request := enchantRequest(raid, map[proto.Stat]float64{proto.Stat_StatAgility: 2}, wrist(agilityWristID))
	capAgilityAtTheSheet(t, request)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotWrist); got != 0 {
		t.Errorf("wrist enchant %d, want none", got)
	}
}

// The baseline already exceeds the ceiling, so the ceiling only stops more Agility.
func TestOptimizeAllowsACeilingTheBaselineExceeds(t *testing.T) {
	addFixtureEnchants()
	request := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(strengthWristID, agilityWristID))
	capAgilityBelowTheSheet(t, request, 10)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotWrist); got != strengthWristID {
		t.Errorf("wrist enchant %d, want %d (Strength)", got, strengthWristID)
	}
}

// An enchant with no stats, such as one with a permanent effect, stays on the boots when the only
// candidate breaks the cap.
func TestOptimizeKeepsAPermanentEffectEnchantWhenNothingIsPicked(t *testing.T) {
	addFixtureEnchants()
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotFeet].Enchant = noStatsFeetID
	feet := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotFeet, EffectIds: []int32{agilityFeetID}}
	request := enchantRequest(raid, map[proto.Stat]float64{proto.Stat_StatAgility: 2}, feet)
	capAgilityAtTheSheet(t, request)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotFeet); got != noStatsFeetID {
		t.Errorf("feet enchant %d, want %d (no stats)", got, noStatsFeetID)
	}
}

// A 1% melee and ranged crit counts as its crit rating, which outweighs the plain shoulder's extra
// attack power.
func TestOptimizeValuesAPercentEnchantThroughItsRating(t *testing.T) {
	addFixtureEnchants()
	const critRatingEP = 4
	weights := map[proto.Stat]float64{proto.Stat_StatStrength: 2.2, proto.Stat_StatAttackPower: 1, proto.Stat_StatCritRating: critRatingEP}
	shoulder := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotShoulder, EffectIds: []int32{attackPowerShoulderID, critShoulderID}}
	result := optimize(t, enchantRequest(protectionWarriorRaid("preraid"), weights, shoulder))
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotShoulder); got != critShoulderID {
		t.Errorf("shoulder enchant %d, want %d (attack power and crit)", got, critShoulderID)
	}
	if got, want := result.GetScore(), critShoulderAP+core.CritRatingPerCritPercent*critRatingEP; math.Abs(got-want) > 1e-9 {
		t.Errorf("score %v, want %v", got, want)
	}
}

// A school hit weight is per percent, so a percent of Holy hit, reached through hit rating or carried
// by the enchant, is worth what its ten hit rating are worth under a hit rating weight a tenth as large.
func TestOptimizeValuesSchoolHitPerPercent(t *testing.T) {
	addFixtureEnchants()
	const holyHitEP = 23.2
	for _, tc := range []struct {
		name    string
		enchant int32
		weight  func(*proto.UnitStats)
	}{
		{"hit rating under a HitRating weight", hitRatingWristID, func(w *proto.UnitStats) {
			w.Stats[proto.Stat_StatHitRating] = holyHitEP / core.HitRatingPerHitPercent
		}},
		{"hit rating under a Holy hit weight", hitRatingWristID, func(w *proto.UnitStats) { w.PseudoStats[proto.PseudoStat_PseudoStatSchoolHitPercentHoly] = holyHitEP }},
		{"Holy hit under a Holy hit weight", holyHitWristID, func(w *proto.UnitStats) { w.PseudoStats[proto.PseudoStat_PseudoStatSchoolHitPercentHoly] = holyHitEP }},
	} {
		request := enchantRequest(protectionWarriorRaid("preraid"), nil, wrist(tc.enchant))
		tc.weight(request.PreCapEpWeights)
		result := optimize(t, request)
		if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != tc.enchant {
			t.Errorf("%s: wrist enchant %d, want %d", tc.name, got, tc.enchant)
		}
		if got := result.GetScore(); math.Abs(got-holyHitEP) > 1e-9 {
			t.Errorf("%s: score %v, want %v", tc.name, got, holyHitEP)
		}
	}
}

// An enchant's ArmorNoMods is worth its own weight per point, with no scale, and its Armor the Armor
// weight. A fixture neck triples the armor from items and then doubles the armor, a percent armor
// modifier, which leaves the ArmorNoMods whole in the cap space's armor as it does on the sheet.
func TestOptimizeValuesArmorNoModsAtItsOwnWeight(t *testing.T) {
	const armor = 120
	values := func(stat proto.Stat) []float64 {
		values := make([]float64, len(proto.Stat_name))
		values[stat] = armor
		return values
	}
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{Id: armorScalingNeckID, Name: "Test Armor Scaling Neck", Type: proto.ItemType_ItemTypeNeck,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}},
		Enchants: []*proto.SimEnchant{
			{EffectId: plainArmorEnchantID, Type: proto.ItemType_ItemTypeChest, Stats: values(proto.Stat_StatArmor)},
			{EffectId: armorNoModsEnchantID, Type: proto.ItemType_ItemTypeChest, Stats: values(proto.Stat_StatArmorNoMods)},
		},
	})
	if !core.HasItemEffect(armorScalingNeckID) {
		core.NewItemEffect(armorScalingNeckID, func(agent core.Agent) {
			character := agent.GetCharacter()
			character.ApplyEquipScaling(stats.Armor, 3)
			character.MultiplyStat(stats.Armor, 2)
		})
	}
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotNeck] = &proto.ItemSpec{Id: armorScalingNeckID}

	baseResult, sdm, sheet := computeReforgeStatsAndDeps(&proto.ComputeStatsRequest{Raid: protopkg.Clone(raid).(*proto.Raid)})
	if baseResult.ErrorResult != "" {
		t.Fatalf("ComputeStats: %s", baseResult.ErrorResult)
	}
	o := &reforgeOptimizer{statDeps: sdm, sheet: sheet, baseStats: protoToCoreUnitStats(baseResult.RaidStats.Parties[0].Players[0].FinalStats)}
	capCoeffs := o.resolveCapCoeffs(rawUnitStatsFromStats(stats.Stats{stats.ArmorNoMods: armor}))
	if got := capCoeffs[statCoeffKey(proto.Stat_StatArmor)]; math.Abs(got-armor) > 1e-9 {
		t.Errorf("%v ArmorNoMods moves the cap space's armor by %v, want all of it", armor, got)
	}

	const armorEP, armorNoModsEP = 2, 0.25
	weights := map[proto.Stat]float64{proto.Stat_StatArmor: armorEP, proto.Stat_StatArmorNoMods: armorNoModsEP}
	for _, tc := range []struct {
		enchant int32
		want    float64
	}{
		{plainArmorEnchantID, armor * armorEP},
		{armorNoModsEnchantID, armor * armorNoModsEP},
	} {
		chest := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotChest, EffectIds: []int32{tc.enchant}}
		result := optimize(t, enchantRequest(protopkg.Clone(raid).(*proto.Raid), weights, chest))
		if got := slotEnchant(result, proto.ItemSlot_ItemSlotChest); got != tc.enchant {
			t.Errorf("chest enchant %d, want %d", got, tc.enchant)
		}
		if got := result.GetScore(); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("enchant %d scores %v, want %v", tc.enchant, got, tc.want)
		}
	}
}

// Crusader is not a candidate, so the main hand keeps it while the wrist is optimized.
func TestOptimizeDropsProcCandidates(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 1900
	crusader := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotMainHand, EffectIds: []int32{1900}}
	addFixtureEnchants()
	result := optimize(t, enchantRequest(raid, agilityOverStrength, crusader, wrist(agilityWristID)))
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotMainHand); got != 1900 {
		t.Errorf("main hand enchant %d, want 1900 (Crusader)", got)
	}
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != agilityWristID {
		t.Errorf("wrist enchant %d, want %d (Agility)", got, agilityWristID)
	}
}

// A listed slot is optimized from its options even when its equipped enchant is a proc the sim
// implements.
func TestOptimizeReplacesAnImplementedProcEnchant(t *testing.T) {
	addFixtureEnchants()
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 1900
	agility := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotMainHand, EffectIds: []int32{agilityWeaponID}}
	result := optimize(t, enchantRequest(raid, agilityOverStrength, agility))
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotMainHand); got != agilityWeaponID {
		t.Errorf("main hand enchant %d, want %d (Agility)", got, agilityWeaponID)
	}
}

// The request's own database is merged before the candidates resolve, so an enchant only it
// carries is a candidate.
func TestOptimizeResolvesAnEnchantFromTheRequestDatabase(t *testing.T) {
	const effectID = 910201
	agility := make([]float64, len(proto.Stat_name))
	agility[proto.Stat_StatAgility] = 20
	raid := protectionWarriorRaid("preraid")
	player := raid.Parties[0].Players[0]
	player.Equipment.Items[proto.ItemSlot_ItemSlotWrist].Enchant = 0
	player.Database = &proto.SimDatabase{Enchants: []*proto.SimEnchant{{EffectId: effectID, Type: proto.ItemType_ItemTypeWrist, Stats: agility}}}
	if got := slotEnchant(optimize(t, enchantRequest(raid, agilityOverStrength, wrist(effectID))), proto.ItemSlot_ItemSlotWrist); got != effectID {
		t.Errorf("wrist enchant %d, want %d", got, effectID)
	}
}

// With no EP on either stat, the wrist keeps Greater Stamina rather than switching or ending empty.
func TestOptimizeKeepsTheEquippedEnchantOnATie(t *testing.T) {
	for _, options := range [][]int32{{929, 1887}, {929}} {
		raid := protectionWarriorRaid("preraid")
		raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotWrist].Enchant = 929
		result := optimize(t, enchantRequest(raid, nil, wrist(options...)))
		if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != 929 {
			t.Errorf("options %v: wrist enchant %d, want 929 (Greater Stamina)", options, got)
		}
		if got := result.GetScore(); got != 0 {
			t.Errorf("options %v: score %v, want 0", options, got)
		}
	}
}

func TestOptimizeLeavesFrozenAndUnlistedSlots(t *testing.T) {
	addFixtureEnchants()
	raid := protectionWarriorRaid("preraid")
	items := raid.Parties[0].Players[0].Equipment.Items
	items[proto.ItemSlot_ItemSlotHead].Enchant = 1506
	items[proto.ItemSlot_ItemSlotWrist].Enchant = strengthWristID
	feet := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotFeet, EffectIds: []int32{agilityFeetID}}
	request := enchantRequest(raid, agilityOverStrength, wrist(agilityWristID), feet)
	request.Settings = &proto.ReforgeSettings{FreezeItemSlots: true, FrozenItemSlots: []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}}

	result := optimize(t, request)
	for _, check := range []struct {
		slot proto.ItemSlot
		want int32
	}{
		{proto.ItemSlot_ItemSlotWrist, strengthWristID},
		{proto.ItemSlot_ItemSlotHead, 1506},
		{proto.ItemSlot_ItemSlotFeet, agilityFeetID},
	} {
		if got := slotEnchant(result, check.slot); got != check.want {
			t.Errorf("%s enchant %d, want %d", check.slot, got, check.want)
		}
	}
}

func TestOptimizeSkipsAnEmptySlot(t *testing.T) {
	addFixtureEnchants()
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{}
	shield := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotOffHand, EffectIds: []int32{929}}
	result := optimize(t, enchantRequest(raid, agilityOverStrength, wrist(agilityWristID), shield))
	if got := result.GetOptimizedGear().Items[proto.ItemSlot_ItemSlotOffHand]; got.GetId() != 0 || got.GetEnchant() != 0 {
		t.Errorf("off hand %v, want empty", got)
	}
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != agilityWristID {
		t.Errorf("wrist enchant %d, want %d (Agility)", got, agilityWristID)
	}
}

func TestOptimizeReturnsTheGearUnchanged(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	result := OptimizeAsync(&proto.ReforgeOptimizeRequest{Raid: raid}, simsignals.CreateSignals())
	if err := result.GetError(); err != nil {
		t.Fatalf("OptimizeAsync: %s", err.GetMessage())
	}

	if gear := result.GetOptimizedGear(); !protopkg.Equal(gear, raid.Parties[0].Players[0].Equipment) {
		t.Errorf("optimized gear %v, want the input gear", gear)
	}
	if result.GetScore() != 0 || result.GetPassesDone() != 0 {
		t.Errorf("score %v after %d passes, want 0 after 0", result.GetScore(), result.GetPassesDone())
	}
}

package database

import (
	"bytes"
	"fmt"
	"go/format"
	"maps"
	"slices"
	"strings"
	"testing"

	_ "github.com/wowsims/forever/sim" // Registers every item set the sim implements.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
	googleProto "google.golang.org/protobuf/proto"
)

// A measured Source needs no notice; a carry-over names its version, an observed fit says so, and anything else is an error.
func TestRateOriginReadsTheSource(t *testing.T) {
	for source, want := range map[string]string{
		"EV-0001":           "",
		"EV-0017, EV-0018":  "",
		"vanilla-carryover": "vanilla",
		"tbc-carryover":     "tbc",
		"observed":          "observed",
	} {
		got, err := rateOrigin(source)
		if err != nil || got != want {
			t.Errorf("source %q reads as origin %q with error %v, want %q and no error", source, got, err, want)
		}
	}
	for _, source := range []string{"EV-1", "estimate", "owner decision"} {
		if got, err := rateOrigin(source); err == nil {
			t.Errorf("source %q reads as origin %q, want an error", source, got)
		}
	}
}

// A stub database holding the items, enchants and sets the fixtures name.
func rateTestDB(itemIDs []int32, enchantIDs []int32, sets map[int32]int32) *WowDatabase {
	db := NewWowDatabase()
	for _, id := range itemIDs {
		db.Items[id] = &proto.UIItem{Id: id, Name: fmt.Sprintf("Item %d", id)}
	}
	for itemID, setID := range sets {
		db.Items[itemID] = &proto.UIItem{Id: itemID, Name: fmt.Sprintf("Item %d", itemID), SetId: setID, SetName: fmt.Sprintf("Set %d", setID)}
	}
	for _, id := range enchantIDs {
		db.Enchants[EnchantDBKey{EffectID: id}] = &proto.UIEnchant{EffectId: id, Name: fmt.Sprintf("Enchant %d", id)}
	}
	return db
}

// Each unmeasured rate lands on its piece of gear with the kind and origin of its row: a weapon rate on its
// item, a spell rate on every item, enchant and set bonus casting the spell, an enchant rate on its
// enchant. A measured rate, a row that is not a rate, and gear the database does not carry stay out. Each
// piece is listed once, in id order.
func TestUnvalidatedRatesListEachUnmeasuredRateOnItsGear(t *testing.T) {
	in := rateInputs{
		WeaponProcs: []overrides.WeaponProc{
			{ItemID: 30, SpellID: 1, PPM: 1, Source: "vanilla-carryover"},
			{ItemID: 30, SpellID: 2, PPM: 2, Source: "vanilla-carryover"},
			{ItemID: 10, SpellID: 3, PPM: 1, Source: "tbc-carryover"},
			{ItemID: 20, SpellID: 4, PPM: 1, Source: "EV-0001"},
			{ItemID: 99, SpellID: 5, PPM: 1, Source: "vanilla-carryover"},
			{ItemID: 60, SpellID: 12, PPM: 1.1, Source: "observed"},
		},
		Spells: []overrides.Override{
			{SpellID: 6, Field: overrides.PPM, Value: 1, Source: "tbc-carryover"},
			{SpellID: 7, Field: overrides.ProcChancePct, Value: 5, Source: "tbc-carryover"},
			{SpellID: 8, Field: overrides.PPM, Value: 1, Source: "EV-0002"},
			{SpellID: 9, Field: overrides.ICDMs, Value: 1000, Source: "tbc-carryover"},
		},
		EnchantProcs: []overrides.EnchantProc{
			{EnchantID: 100, SpellID: 10, Kind: overrides.RateChance, Value: 0.1, Source: "vanilla-carryover"},
			{EnchantID: 200, SpellID: 11, Kind: overrides.RatePPM, Value: 1, Source: "EV-0003"},
		},
		SpellItems:    map[int32][]int32{6: {40, 98}, 7: {45}, 8: {20}, 9: {45}},
		SpellEnchants: map[int32][]int32{6: {300}},
		SpellSets:     map[int32][][2]int32{6: {{7000, 6}, {9000, 2}}},
	}
	db := rateTestDB([]int32{10, 20, 30, 40, 45, 60}, []int32{100, 200, 300}, map[int32]int32{70: 7000})

	rates, err := unvalidatedRates(in, db)
	if err != nil {
		t.Fatal(err)
	}

	type entry struct {
		id, pieces int32
		kind       overrides.RateKind
		rate       float64
		origin     string
	}
	entries := func(list []UnvalidatedRate) []entry {
		var got []entry
		for _, rate := range list {
			got = append(got, entry{rate.ID, rate.Pieces, rate.Kind, rate.Rate, rate.Origin})
		}
		return got
	}
	for _, tc := range []struct {
		name string
		got  []UnvalidatedRate
		want []entry
	}{
		{"items", rates.Items, []entry{
			{10, 0, overrides.RatePPM, 1, "tbc"},
			{30, 0, overrides.RatePPM, 0, "vanilla"}, // Two procs at 1 and 2 per minute: no single rate.
			{40, 0, overrides.RatePPM, 1, "tbc"},
			{45, 0, overrides.RateChance, 5, "tbc"},
			{60, 0, overrides.RatePPM, 1.1, "observed"},
		}},
		{"enchants", rates.Enchants, []entry{
			{100, 0, overrides.RateChance, 10, "vanilla"}, // A chance of 0.1 is 10 percent.
			{300, 0, overrides.RatePPM, 1, "tbc"},
		}},
		{"set bonuses", rates.SetBonuses, []entry{
			{7000, 6, overrides.RatePPM, 1, "tbc"},
		}},
	} {
		if got := entries(tc.got); !slices.Equal(got, tc.want) {
			t.Errorf("%s listed %v, want %v", tc.name, got, tc.want)
		}
	}
	if rates.Items[0].Name != db.Items[10].Name || rates.SetBonuses[0].Name != db.Items[70].SetName {
		t.Errorf("entries carry the names %q and %q, want the database's", rates.Items[0].Name, rates.SetBonuses[0].Name)
	}
}

// Two rates of different origin on one piece of gear cannot share its notice, so the lists refuse them.
func TestUnvalidatedRatesRefuseTwoOriginsOnOneItem(t *testing.T) {
	in := rateInputs{WeaponProcs: []overrides.WeaponProc{
		{ItemID: 10, SpellID: 1, PPM: 1, Source: "vanilla-carryover"},
		{ItemID: 10, SpellID: 2, PPM: 1, Source: "tbc-carryover"},
	}}
	if _, err := unvalidatedRates(in, rateTestDB([]int32{10}, nil, nil)); err == nil {
		t.Error("an item with a vanilla and a TBC rate was listed, want an error")
	}
}

// A rate whose Source is no evidence id, carry-over or observed fit stops the lists.
func TestUnvalidatedRatesRefuseAnUnknownSource(t *testing.T) {
	in := rateInputs{WeaponProcs: []overrides.WeaponProc{{ItemID: 10, SpellID: 1, PPM: 1, Source: "estimate"}}}
	if _, err := unvalidatedRates(in, rateTestDB([]int32{10}, nil, nil)); err == nil {
		t.Error("a rate with the source \"estimate\" was listed, want an error")
	}
}

// Each rate goes onto its gear for the gear tab: an item's onto the item, an enchant's onto every enchant sharing
// its effect id, and a set bonus's into the database's list in set and piece order. A settled rate keeps its
// status and an unvalidated one takes its origin's, with a rate of 0 where it has no single one. The effects the sim
// rolls keep their own rates.
func TestStampProcRatesPutsEachRateOnItsGear(t *testing.T) {
	db := rateTestDB([]int32{1, 2, 3}, []int32{10, 11}, map[int32]int32{70: 7000})
	db.Enchants[EnchantDBKey{EffectID: 10, ItemID: 5}] = &proto.UIEnchant{EffectId: 10, ItemId: 5}
	effect := &proto.ProcEffect{ProcRate: &proto.ProcEffect_ProcChance{ProcChance: 0.03}}
	db.Items[1].ItemEffects = []*proto.ItemEffect{{Effect: &proto.ItemEffect_Proc{Proc: effect}}}
	known := KnownRates{
		Items:      []KnownRate{{ID: 1, Kind: overrides.RatePPM, Rate: 1.5, Status: proto.ProcRateStatus_ProcRateStatusMeasured}},
		Enchants:   []KnownRate{{ID: 10, Kind: overrides.RateChance, Rate: 10, Status: proto.ProcRateStatus_ProcRateStatusGameFiles}},
		SetBonuses: []KnownRate{{ID: 7000, Pieces: 4, Kind: overrides.RateChance, Rate: 4, Status: proto.ProcRateStatus_ProcRateStatusGameFiles}},
	}
	unvalidated := UnvalidatedRates{
		Items: []UnvalidatedRate{
			{ID: 2, Kind: overrides.RatePPM, Rate: 1.1, Origin: "vanilla"},
			{ID: 3, Kind: overrides.RateChance, Origin: "observed"},
		},
		Enchants:   []UnvalidatedRate{{ID: 11, Kind: overrides.RatePPM, Rate: 2, Origin: "tbc"}},
		SetBonuses: []UnvalidatedRate{{ID: 7000, Pieces: 2, Kind: overrides.RatePPM, Rate: 0.5, Origin: "tbc"}},
	}

	StampProcRates(db, known, unvalidated)

	rate := func(kind proto.ProcRateKind, value float64, status proto.ProcRateStatus) *proto.ProcRate {
		return &proto.ProcRate{Kind: kind, Rate: value, Status: status}
	}
	ppm, chance := proto.ProcRateKind_ProcRateKindPpm, proto.ProcRateKind_ProcRateKindChance
	for _, tc := range []struct {
		name string
		got  *proto.ProcRate
		want *proto.ProcRate
	}{
		{"item 1", db.Items[1].ProcRate, rate(ppm, 1.5, proto.ProcRateStatus_ProcRateStatusMeasured)},
		{"item 2", db.Items[2].ProcRate, rate(ppm, 1.1, proto.ProcRateStatus_ProcRateStatusVanillaCarryover)},
		{"item 3", db.Items[3].ProcRate, rate(chance, 0, proto.ProcRateStatus_ProcRateStatusObserved)},
		{"set item 70", db.Items[70].ProcRate, nil},
		{"enchant 10", db.Enchants[EnchantDBKey{EffectID: 10}].ProcRate, rate(chance, 10, proto.ProcRateStatus_ProcRateStatusGameFiles)},
		{"enchant 10 from item 5", db.Enchants[EnchantDBKey{EffectID: 10, ItemID: 5}].ProcRate, rate(chance, 10, proto.ProcRateStatus_ProcRateStatusGameFiles)},
		{"enchant 11", db.Enchants[EnchantDBKey{EffectID: 11}].ProcRate, rate(ppm, 2, proto.ProcRateStatus_ProcRateStatusTbcCarryover)},
	} {
		if !googleProto.Equal(tc.got, tc.want) {
			t.Errorf("%s carries %v, want %v", tc.name, tc.got, tc.want)
		}
	}
	wantSets := []*proto.SetBonusProcRate{
		{SetId: 7000, Pieces: 2, Rate: rate(ppm, 0.5, proto.ProcRateStatus_ProcRateStatusTbcCarryover)},
		{SetId: 7000, Pieces: 4, Rate: rate(chance, 4, proto.ProcRateStatus_ProcRateStatusGameFiles)},
	}
	if !slices.EqualFunc(db.SetBonusProcRates, wantSets, func(a, b *proto.SetBonusProcRate) bool { return googleProto.Equal(a, b) }) {
		t.Errorf("the set bonus rates are %v, want %v", db.SetBonusProcRates, wantSets)
	}
	if effect.GetProcChance() != 0.03 {
		t.Errorf("item 1's effect rolls a chance of %v, want its own 0.03", effect.GetProcChance())
	}
}

// Each settled rate lands on its piece of gear: the rate on an item's or an enchant's effects, a chance
// as a percent, a measured rate override on the spell an item, enchant or set bonus casts, an enchant's and
// a set bonus's row chance. Each proc row of hand-written gear takes, over the effects of its gear, the gear's
// own rate override for its spell, else the spell's rate override, else what the row states, and a gear rate
// override replaces them all: one with no kind leaves the gear with none, one with a kind states its rate with
// the status its Source gives. Gear with no single rate, a proc stating no rate that no override fills, a
// carried-over override, an unvalidated rate, an item or enchant missing an effect and a missing set bonus stay
// out, and so does a set the database lacks. A hand-written proc whose gear carries an unmeasured rate of its
// own for the spell gets none. Generated gear counts only where its procs register, and a row's rate only where
// they read the rows. Each piece is listed once, in id order.
func TestKnownRatesListEachSettledRateOnItsGear(t *testing.T) {
	procs := func(rates ...*proto.ProcEffect) []*proto.ItemEffect {
		var effects []*proto.ItemEffect
		for _, rate := range rates {
			effects = append(effects, &proto.ItemEffect{Effect: &proto.ItemEffect_Proc{Proc: rate}})
		}
		return effects
	}
	ppm := func(rate float64) *proto.ProcEffect {
		return &proto.ProcEffect{ProcRate: &proto.ProcEffect_Ppm{Ppm: rate}}
	}
	chance := func(rate float64) *proto.ProcEffect {
		return &proto.ProcEffect{ProcRate: &proto.ProcEffect_ProcChance{ProcChance: rate}}
	}
	db := rateTestDB([]int32{20, 21, 23, 24, 25, 26, 27, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50}, []int32{200, 201, 202, 203, 204, 205, 206, 207}, map[int32]int32{70: 7000, 71: 9000, 72: 9100})
	for id, effects := range map[int32][]*proto.ItemEffect{
		20: procs(ppm(1.5)),
		21: procs(chance(0.05)),
		23: procs(chance(0.05)),
		24: procs(ppm(1), ppm(2)),
		25: procs(ppm(1), chance(0.05)),
		26: procs(ppm(1)),
		27: procs(ppm(1), &proto.ProcEffect{}),
		32: procs(&proto.ProcEffect{}),
		34: procs(ppm(1)),
		35: procs(chance(0.04)),
		36: procs(&proto.ProcEffect{}),
		37: procs(ppm(1), ppm(1)),
		38: procs(chance(0.05)),
		39: procs(ppm(1)),
		41: procs(chance(0.05)),
	} {
		db.Items[id].ItemEffects = effects
	}
	db.Enchants[EnchantDBKey{EffectID: 200}].EnchantEffects = procs(chance(0.1))
	db.Enchants[EnchantDBKey{EffectID: 201}].EnchantEffects = procs(chance(0.08))
	db.Enchants[EnchantDBKey{EffectID: 204}].EnchantEffects = procs(chance(0.06))
	db.Enchants[EnchantDBKey{EffectID: 207}].EnchantEffects = procs(chance(0.1))
	in := rateInputs{
		Spells: []overrides.Override{
			{SpellID: 6, Field: overrides.PPM, Value: 2, Source: "EV-0001"},
			{SpellID: 40, Field: overrides.PPM, Value: 1.8, Source: "EV-0005"},
			{SpellID: 41, Field: overrides.ProcChancePct, Value: 6, Source: "EV-0011"},
			{SpellID: 42, Field: overrides.PPM, Value: 1, Source: "tbc-carryover"},
			{SpellID: 50, Field: overrides.PPM, Value: 1.4, Source: "EV-0007"},
			{SpellID: 52, Field: overrides.ProcChancePct, Value: 9, Source: "tbc-carryover"},
			{SpellID: 57, Field: overrides.PPM, Value: 3, Source: "EV-0010"},
			{SpellID: 58, Field: overrides.PPM, Value: 1.2, Source: "EV-0012"},
			{SpellID: 59, Field: overrides.PPM, Value: 1, Source: "tbc-carryover"},
		},
		WeaponProcs: []overrides.WeaponProc{
			{ItemID: 43, SpellID: 53, PPM: 2.2, Source: "EV-0008"},
			{ItemID: 45, SpellID: 57, PPM: 1, Source: "vanilla-carryover"}, // Its own rate for the spell comes first and is unmeasured.
		},
		EnchantProcs: []overrides.EnchantProc{{EnchantID: 206, SpellID: 55, Kind: overrides.RateChance, Value: 0.08, Source: "EV-0009"}},
		HandProcs: map[string]map[[2]int32][]HandProc{
			"item": {
				{40, 0}: {{SpellID: 50, Row: gameFilesChance(3)}},
				{41, 0}: {{SpellID: 51, Row: gameFilesChance(6)}},
				{42, 0}: {{SpellID: 52, Row: gameFilesChance(9)}},
				{43, 0}: {{SpellID: 53}},
				{44, 0}: {{SpellID: 54}},
				{45, 0}: {{SpellID: 57}},
				{46, 0}: {{SpellID: 60, Row: gameFilesChance(6)}},
				{47, 0}: {{SpellID: 61, Row: gameFilesChance(6)}},
				{48, 0}: {{SpellID: 62}},
				{49, 0}: {{SpellID: 63, Row: gameFilesChance(3)}, {SpellID: 64, Row: gameFilesChance(5)}}, // Two rows at two chances: no single rate.
			},
			"enchant":   {{206, 0}: {{SpellID: 55}}},
			"set bonus": {{7000, 8}: {{SpellID: 56, Row: gameFilesChance(5)}}},
		},
		GearProcRates: []overrides.GearProcRate{
			{Gear: overrides.GearItem, ID: 46, Source: "tooltip"},
			{Gear: overrides.GearItem, ID: 47, Kind: overrides.RateChance, Rate: 0.04, Source: "tooltip"},
			{Gear: overrides.GearItem, ID: 48, Kind: overrides.RatePPM, Rate: 2, Source: "EV-0013"},
			{Gear: overrides.GearItem, ID: 50, Kind: overrides.RatePPM, Rate: 1.5, Source: "observed"},
		},
		SpellSets:      map[int32][][2]int32{6: {{7000, 6}}, 58: {{9000, 4}}, 59: {{9100, 2}}},
		HandItems:      map[int32]bool{23: true, 31: true, 40: true, 41: true, 42: true, 43: true, 44: true, 45: true, 46: true, 47: true, 48: true, 49: true, 50: true},
		HandSets:       map[int32]bool{9000: true},
		HandEnchants:   map[int32]bool{201: true, 206: true},
		SetChances:     map[[2]int32][]float64{{7000, 4}: {4}, {7000, 2}: {3}, {8000, 2}: {5}},
		SpellItems:     map[int32][]int32{40: {30, 31, 35, 36}, 41: {32}, 42: {33}},
		EnchantChances: map[int32][]float64{202: {3}, 203: {5, 0}, 205: {9}},
		LiveItems: map[int32]LiveProcs{
			20: {1, true}, 21: {1, true}, 24: {2, true}, 25: {2, true}, 26: {1, true}, 27: {2, true}, 30: {1, true},
			32: {1, true}, 33: {1, true}, 35: {1, false}, 36: {1, false}, 37: {1, true}, 38: {1, true}, 39: {1, true},
		},
		LiveEnchants: map[int32]LiveProcs{200: {1, true}, 202: {1, true}, 203: {2, true}, 205: {1, false}, 207: {1, true}},
		TargetScaled: map[string]map[[2]int32]bool{"item": {{38, 0}: true}}, // Its rate varies by target type, so none is listed.
		Missing:      map[string]map[int32]bool{"item": {39: true}, "enchant": {207: true}},
	}
	unvalidated := UnvalidatedRates{Items: []UnvalidatedRate{{ID: 26, Kind: overrides.RatePPM, Rate: 1, Origin: "vanilla"}}}
	missingSets := []MissingSetBonus{{ID: 7000, Pieces: 2, Name: "Set 7000"}}

	rates := knownRates(in, db, unvalidated, missingSets)

	type entry struct {
		id, pieces int32
		kind       overrides.RateKind
		rate       float64
		status     proto.ProcRateStatus
	}
	entries := func(list []KnownRate) []entry {
		var got []entry
		for _, rate := range list {
			got = append(got, entry{rate.ID, rate.Pieces, rate.Kind, rate.Rate, rate.Status})
		}
		return got
	}
	const measured, gameFiles, observed = proto.ProcRateStatus_ProcRateStatusMeasured, proto.ProcRateStatus_ProcRateStatusGameFiles, proto.ProcRateStatus_ProcRateStatusObserved
	for _, tc := range []struct {
		name string
		got  []KnownRate
		want []entry
	}{
		{"items", rates.Items, []entry{
			{20, 0, overrides.RatePPM, 1.5, measured},   // An effect's procs per minute come from a rate override.
			{21, 0, overrides.RateChance, 5, gameFiles}, // A chance of 0.05 is 5 percent.
			{30, 0, overrides.RatePPM, 1.8, measured},
			{32, 0, overrides.RateChance, 6, measured},  // The override states the rate its effect leaves out.
			{35, 0, overrides.RateChance, 4, gameFiles}, // Its registration reads the effect, never the row's override.
			{40, 0, overrides.RatePPM, 1.4, measured},
			{41, 0, overrides.RateChance, 6, gameFiles}, // The row's chance, not the effect's 5.
			{43, 0, overrides.RatePPM, 2.2, measured},
			{47, 0, overrides.RateChance, 4, gameFiles}, // A tooltip's rate is the game files'.
			{48, 0, overrides.RatePPM, 2, measured},
			{50, 0, overrides.RatePPM, 1.5, observed},
		}},
		{"enchants", rates.Enchants, []entry{
			{200, 0, overrides.RateChance, 10, gameFiles},
			{202, 0, overrides.RateChance, 3, gameFiles},
			{206, 0, overrides.RateChance, 8, measured},
		}},
		{"set bonuses", rates.SetBonuses, []entry{
			{7000, 4, overrides.RateChance, 4, gameFiles},
			{7000, 6, overrides.RatePPM, 2, measured},
			{7000, 8, overrides.RateChance, 5, gameFiles},
		}},
	} {
		if got := entries(tc.got); !slices.Equal(got, tc.want) {
			t.Errorf("%s listed %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A set bonus is missing when no registered set covers its piece count: a registered set lists the
// counts it leaves out, an unregistered set lists them all, a set registered with ID 0 is matched by
// name, and a set with no items in the database stays out.
func TestMissingSetBonusesListUncoveredPieceCounts(t *testing.T) {
	registered := []*core.ItemSet{
		{ID: 100, Name: "Set 100", Bonuses: map[int32]core.ApplySetBonus{2: nil}},
		{Name: "Set 300", Bonuses: map[int32]core.ApplySetBonus{2: nil, 4: nil}},
	}
	setSpells := map[int][]dbc.ItemSetSpell{
		100: {{ItemSetID: 100, Threshold: 2, SpellID: 1}, {ItemSetID: 100, Threshold: 4, SpellID: 2}, {ItemSetID: 100, Threshold: 4, SpellID: 3}},
		200: {{ItemSetID: 200, Threshold: 3, SpellID: 4}, {ItemSetID: 200, Threshold: 6, SpellID: 5}},
		300: {{ItemSetID: 300, Threshold: 2, SpellID: 6}, {ItemSetID: 300, Threshold: 4, SpellID: 7}, {ItemSetID: 300, Threshold: 5, SpellID: 8}},
		400: {{ItemSetID: 400, Threshold: 2, SpellID: 9}},
	}
	db := rateTestDB(nil, nil, map[int32]int32{10: 100, 20: 200, 30: 300})

	got := missingSetBonuses(registered, setSpells, db, func(spellID int) string { return fmt.Sprintf("Tooltip %d", spellID) })

	want := []MissingSetBonus{
		{ID: 100, Pieces: 4, Name: "Set 100", Spells: []int32{2, 3}, Tooltips: []string{"Tooltip 2", "Tooltip 3"}},
		{ID: 200, Pieces: 3, Name: "Set 200", Spells: []int32{4}, Tooltips: []string{"Tooltip 4"}},
		{ID: 200, Pieces: 6, Name: "Set 200", Spells: []int32{5}, Tooltips: []string{"Tooltip 5"}},
		{ID: 300, Pieces: 5, Name: "Set 300", Spells: []int32{8}, Tooltips: []string{"Tooltip 8"}},
	}
	if !slices.EqualFunc(got, want, func(a, b MissingSetBonus) bool {
		return a.ID == b.ID && a.Pieces == b.Pieces && a.Name == b.Name && slices.Equal(a.Spells, b.Spells) && slices.Equal(a.Tooltips, b.Tooltips)
	}) {
		t.Errorf("listed %+v, want %+v", got, want)
	}
}

// A set's missing bonuses render under the set by piece count, one tooltip per spell.
func TestMissingSetBonusesRender(t *testing.T) {
	bonuses := []MissingSetBonus{
		{ID: 3, Pieces: 2, Name: "Set", Spells: []int32{11}, Tooltips: []string{`Says "hi".`}},
		{ID: 3, Pieces: 4, Name: "Set", Spells: []int32{12, 13}, Tooltips: []string{"One.", "Two."}},
	}
	var rendered bytes.Buffer
	if err := renderMissingEffects(&rendered, bonuses); err != nil {
		t.Fatal(err)
	}

	block := "export const MISSING_SET_BONUSES = new Map<number, Map<number, string[]>>([\n" +
		"\t[\n" +
		"\t\t3, // Set\n" +
		"\t\tnew Map([\n" +
		"\t\t\t[\n" +
		"\t\t\t\t2,\n" +
		"\t\t\t\t[\n" +
		"\t\t\t\t\t\"Says \\\"hi\\\".\", // 11 - https://www.wowhead.com/forever/spell=11\n" +
		"\t\t\t\t],\n" +
		"\t\t\t],\n" +
		"\t\t\t[\n" +
		"\t\t\t\t4,\n" +
		"\t\t\t\t[\n" +
		"\t\t\t\t\t\"One.\", // 12 - https://www.wowhead.com/forever/spell=12\n" +
		"\t\t\t\t\t\"Two.\", // 13 - https://www.wowhead.com/forever/spell=13\n" +
		"\t\t\t\t],\n" +
		"\t\t\t],\n" +
		"\t\t]),\n" +
		"\t],\n" +
		"])\n"
	if !strings.Contains(rendered.String(), block) {
		t.Errorf("the generated file holds no block\n%s\nin\n%s", block, rendered.String())
	}
}

// A set no registered set covers is generated with every piece count, and a bonus is supported only when
// each of its rows registers: it parses whole, or it is a proc trigger the set builds. A set covered by id, or by name for a set registered with ID 0, and a
// set with no items in the database stay out. The skill line the ItemSet row requires is the set's profession.
func TestGeneratedSetsTakeTheSetsNoRegisteredSetCovers(t *testing.T) {
	const hit, proc = 460230, 1321560 // +2% hit; a proc trigger whose mask states no listener
	const lightning = 18979           // a 10% chance on melee hits to cast 18980's Nature damage
	const leatherworking = 165
	registered := []*core.ItemSet{{ID: 100, Name: "Set 100"}, {Name: "Set 300"}}
	setSpells := map[int][]dbc.ItemSetSpell{
		100: {{ItemSetID: 100, Threshold: 2, SpellID: hit}},
		200: {
			{ItemSetID: 200, Threshold: 4, SpellID: hit, RequiredSkill: leatherworking},
			{ItemSetID: 200, Threshold: 4, SpellID: proc, RequiredSkill: leatherworking},
			{ItemSetID: 200, Threshold: 2, SpellID: hit, RequiredSkill: leatherworking},
			{ItemSetID: 200, Threshold: 3, SpellID: lightning, RequiredSkill: leatherworking},
		},
		300: {{ItemSetID: 300, Threshold: 2, SpellID: hit}},
		400: {{ItemSetID: 400, Threshold: 2, SpellID: hit}},
	}
	db := rateTestDB(nil, nil, map[int32]int32{10: 100, 20: 200, 30: 300})

	got := generatedSets(registered, setSpells, db, func(spellID int) string { return fmt.Sprintf("Tooltip %d", spellID) })

	if len(got) != 1 || got[0].ID != 200 || got[0].Name != "Set 200" {
		t.Fatalf("generated %+v, want set 200 alone", got)
	}
	if got[0].Profession != proto.Profession_Leatherworking {
		t.Errorf("set 200's row requires Leatherworking (165), generated with profession %v", got[0].Profession)
	}
	type bonus struct {
		pieces    int32
		spells    []int32
		supported bool
	}
	want := []bonus{{2, []int32{hit}, true}, {3, []int32{lightning}, true}, {4, []int32{hit, proc}, false}}
	if !slices.EqualFunc(got[0].Bonuses, want, func(a GeneratedSetBonus, b bonus) bool {
		return a.Pieces == b.pieces && slices.Equal(a.Spells, b.spells) && a.Supported() == b.supported
	}) {
		t.Errorf("generated bonuses %+v, want %+v", got[0].Bonuses, want)
	}
}

// A set with a supported bonus is a registration whose other bonuses are TODO comments; a set with none
// is commented out whole. A set with a profession registers gated on it.
func TestGeneratedSetsRender(t *testing.T) {
	sets := []GeneratedSet{
		{ID: 1, Name: "Set", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}, Bonuses: []GeneratedSetBonus{
			{Pieces: 2, Spells: []int32{11}, Tooltips: []string{"Hit."}, Unsupported: [][]string{nil}},
			{Pieces: 4, Spells: []int32{12}, Tooltips: []string{"Proc."}, Unsupported: [][]string{{"effect 1 is E_DUMMY"}}},
		}},
		{ID: 2, Name: "Other", Bonuses: []GeneratedSetBonus{
			{Pieces: 2, Spells: []int32{13}, Tooltips: []string{"Proc."}, Unsupported: [][]string{{"effect 1 is E_DUMMY"}}},
		}},
		{ID: 3, Name: "Crafted", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotChest}, Profession: proto.Profession_Tailoring, Bonuses: []GeneratedSetBonus{
			{Pieces: 2, Spells: []int32{14}, Tooltips: []string{"Crit."}, Unsupported: [][]string{nil}},
		}},
	}

	got, err := format.Source(renderSetBonuses(sets))
	if err != nil {
		t.Fatal(err)
	}

	for _, block := range []string{
		"\tshared.NewParsedItemSet(1, \"Set\", []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}, map[int32][]int32{\n" +
			"\t\t// Hit.\n" +
			"\t\t// https://www.wowhead.com/forever/spell=11\n" +
			"\t\t2: {11},\n" +
			"\n" +
			"\t\t// TODO: Manual implementation required\n" +
			"\t\t//\n" +
			"\t\t// Proc.\n" +
			"\t\t// https://www.wowhead.com/forever/spell=12\n" +
			"\t\t// unsupported: effect 1 is E_DUMMY\n" +
			"\t\t// 4: {12},\n" +
			"\t})\n",
		"\t// TODO: Manual implementation required\n" +
			"\t//\n" +
			"\t// Other\n" +
			"\t// https://www.wowhead.com/forever/item-set=2\n" +
			"\t// shared.NewParsedItemSet(2, \"Other\", []proto.ItemSlot{}, map[int32][]int32{\n" +
			"\t// \t// Proc.\n" +
			"\t// \t// https://www.wowhead.com/forever/spell=13\n" +
			"\t// \t// unsupported: effect 1 is E_DUMMY\n" +
			"\t// \t2: {13},\n" +
			"\t// })\n",
		"\tshared.NewParsedProfessionItemSet(3, \"Crafted\", proto.Profession_Tailoring, []proto.ItemSlot{proto.ItemSlot_ItemSlotChest}, map[int32][]int32{\n" +
			"\t\t// Crit.\n" +
			"\t\t// https://www.wowhead.com/forever/spell=14\n" +
			"\t\t2: {14},\n" +
			"\t})\n",
	} {
		if !strings.Contains(string(got), block) {
			t.Errorf("the generated file holds no block\n%s\nin\n%s", block, got)
		}
	}
}

// A registered set with an id grants its bonuses only at piece counts the game files give that set.
func TestRegisteredSetBonusesSitOnGameFilesThresholds(t *testing.T) {
	withDBCInputs(t)
	setSpells := dbc.GetDBC().ItemSetSpells

	if len(core.ItemSets()) == 0 {
		t.Fatal("no item set is registered")
	}
	for _, set := range core.ItemSets() {
		if set.ID == 0 {
			continue
		}
		thresholds := map[int32]bool{}
		for _, row := range setSpells[int(set.ID)] {
			thresholds[int32(row.Threshold)] = true
		}
		for _, pieces := range slices.Sorted(maps.Keys(set.Bonuses)) {
			if !thresholds[pieces] {
				t.Errorf("set %d (%s) registers a %d-piece bonus, which the game files do not grant", set.ID, set.Name, pieces)
			}
		}
	}
}

// What holds for any shipped rate: its Source is a carry-over, an observed fit or a measurement, every row
// names gear the game files carry, and no measured rate is listed.
func TestShippedRatesAreWellFormed(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()
	db := NewWowDatabase()
	in := shippedRateInputs(instance, db)

	knownSource := func(source string) bool {
		_, err := rateOrigin(source)
		return err == nil
	}
	var measuredItems, measuredEnchants []int32
	for _, row := range in.WeaponProcs {
		if !knownSource(row.Source) {
			t.Errorf("item %d's rate states source %q, want a carry-over, an observed fit or a measurement", row.ItemID, row.Source)
		}
		if _, ok := instance.Items[int(row.ItemID)]; !ok {
			t.Errorf("item %d's rate names an item the game files do not carry", row.ItemID)
		}
		if origin, _ := rateOrigin(row.Source); origin == "" {
			measuredItems = append(measuredItems, row.ItemID)
		}
		db.Items[row.ItemID] = &proto.UIItem{Id: row.ItemID}
	}
	for _, row := range in.EnchantProcs {
		if !knownSource(row.Source) {
			t.Errorf("enchant %d's rate states source %q, want a carry-over, an observed fit or a measurement", row.EnchantID, row.Source)
		}
		enchant, ok := instance.EnchantsByEffectId[int(row.EnchantID)]
		if !ok || !slices.Contains(enchant.EffectArgs, int(row.SpellID)) {
			t.Errorf("enchant %d's rate names spell %d, which no slot of an enchant the game files carry casts", row.EnchantID, row.SpellID)
		}
		if origin, _ := rateOrigin(row.Source); origin == "" {
			measuredEnchants = append(measuredEnchants, row.EnchantID)
		}
		db.Enchants[EnchantDBKey{EffectID: row.EnchantID}] = &proto.UIEnchant{EffectId: row.EnchantID}
	}
	for _, row := range in.Spells {
		if spellRateKind(row) != 0 && !knownSource(row.Source) {
			t.Errorf("spell %d's rate states source %q, want a carry-over, an observed fit or a measurement", row.SpellID, row.Source)
		}
		for _, itemID := range in.SpellItems[row.SpellID] {
			if origin, _ := rateOrigin(row.Source); spellRateKind(row) != 0 && origin == "" {
				measuredItems = append(measuredItems, itemID)
			}
			db.Items[itemID] = &proto.UIItem{Id: itemID}
		}
	}

	rates, err := unvalidatedRates(in, db)
	if err != nil {
		t.Fatal(err)
	}
	for _, rate := range rates.Items {
		if slices.Contains(measuredItems, rate.ID) {
			t.Errorf("item %d carries a measured rate and is listed as unvalidated", rate.ID)
		}
	}
	for _, rate := range rates.Enchants {
		if slices.Contains(measuredEnchants, rate.ID) {
			t.Errorf("enchant %d carries a measured rate and is listed as unvalidated", rate.ID)
		}
	}
}

// Each gear proc rate names gear the sim registers by hand, once, and states a Reason. A rate comes only with
// a kind and a Source the gear tab reads, a chance is at most 1, and a row with no kind may leave Source empty.
// A row whose gear's tooltip prints the chance as a fraction of the column states no kind, since that gear
// shows no rate. The check refuses Bonechill Hammer 14487, whose proc the generated files register.
func TestGearProcRatesNameHandWrittenGear(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()
	setNames := map[int32]string{}
	for _, item := range instance.Items {
		if item.ItemSetId != 0 {
			setNames[int32(item.ItemSetId)] = item.ItemSetName
		}
	}
	handWritten := func(o overrides.GearProcRate) bool {
		switch o.Gear {
		case overrides.GearItem:
			return o.Pieces == 0 && implementedByHand(o.ID)
		case overrides.GearEnchant:
			return o.Pieces == 0 && core.HasEnchantEffect(o.ID)
		case overrides.GearSetBonus:
			return slices.ContainsFunc(core.ItemSets(), func(set *core.ItemSet) bool {
				_, ok := set.Bonuses[o.Pieces]
				return ok && coversSet(set, o.ID, setNames[o.ID])
			})
		}
		return false
	}
	if handWritten(overrides.GearProcRate{Gear: overrides.GearItem, ID: 14487}) {
		t.Fatal("Bonechill Hammer 14487, whose proc the generated files register, reads as written by hand")
	}

	in := shippedRateInputs(instance, NewWowDatabase())
	seen := map[[3]int32]bool{}
	for _, o := range overrides.GearProcRates {
		key := [3]int32{int32(o.Gear), o.ID, o.Pieces}
		if seen[key] {
			t.Errorf("the %s %d (%d pieces) rate has two rows", o.Gear, o.ID, o.Pieces)
		}
		seen[key] = true
		if !handWritten(o) {
			t.Errorf("the %s %d (%d pieces) rate names gear the sim does not register by hand", o.Gear, o.ID, o.Pieces)
		}
		if _, err := gearProcRateStatus(o.Source); o.Reason == "" || (o.Kind != 0 || o.Source != "") && err != nil {
			t.Errorf("the %s %d rate states reason %q and source %q", o.Gear, o.ID, o.Reason, o.Source)
		}
		if o.Kind == 0 && o.Rate != 0 || o.Kind != 0 && (o.Kind > overrides.RateChance || o.Rate <= 0) || o.Kind == overrides.RateChance && o.Rate > 1 {
			t.Errorf("the %s %d rate is a %s of %v", o.Gear, o.ID, o.Kind, o.Rate)
		}
		if o.Kind != 0 && in.TargetScaled[o.Gear.String()][[2]int32{o.ID, o.Pieces}] {
			t.Errorf("the %s %d rate states a %s, which its gear never shows: its tooltip prints the chance as a fraction of the column", o.Gear, o.ID, o.Kind)
		}
	}
}

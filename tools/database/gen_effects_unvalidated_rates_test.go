package database

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// A measured Source needs no notice; a carry-over names its version, and anything else is an estimate.
func TestRateOriginReadsTheSource(t *testing.T) {
	for source, want := range map[string]string{
		"wcl:abc/1-2, 2026-01-01":             "",
		"issue #1":                            "",
		"combat log, Rend Rank 2, 2026-10-01": "",
		"vanilla-carryover":                   "vanilla",
		"tbc-carryover":                       "tbc",
		"estimate":                            "unknown",
		"owner decision":                      "unknown",
	} {
		if got := rateOrigin(source); got != want {
			t.Errorf("source %q reads as origin %q, want %q", source, got, want)
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

// Each unmeasured rate lands on its piece of gear with the kind and origin of its row: a weapon rate
// on its item, a spell rate on every item, enchant and set bonus casting the spell, an enchant rate on its
// enchant, a literal on its item, enchant or set bonus. A measured rate, a row that is not a rate, and
// gear the database does not carry stay out. Each piece is listed once, in id order.
func TestUnvalidatedRatesListEachUnmeasuredRateOnItsGear(t *testing.T) {
	in := rateInputs{
		WeaponProcs: []overrides.WeaponProc{
			{ItemID: 30, SpellID: 1, PPM: 1, Source: "vanilla-carryover"},
			{ItemID: 30, SpellID: 2, PPM: 2, Source: "vanilla-carryover"},
			{ItemID: 10, SpellID: 3, PPM: 1, Source: "tbc-carryover"},
			{ItemID: 20, SpellID: 4, PPM: 1, Source: "wcl:abc/1"},
			{ItemID: 99, SpellID: 5, PPM: 1, Source: "vanilla-carryover"},
		},
		Spells: []overrides.Override{
			{SpellID: 6, Field: overrides.PPM, Value: 1, Source: "tbc-carryover"},
			{SpellID: 7, Field: overrides.ProcChancePct, Value: 5, Source: "owner decision"},
			{SpellID: 8, Field: overrides.PPM, Value: 1, Source: "issue #1"},
			{SpellID: 9, Field: overrides.ICDMs, Value: 1000, Source: "tbc-carryover"},
		},
		EnchantProcs: []overrides.EnchantProc{
			{EnchantID: 100, SpellID: 10, Kind: overrides.RateChance, Value: 0.1, Source: "vanilla-carryover"},
			{EnchantID: 200, SpellID: 11, Kind: overrides.RatePPM, Value: 1, Source: "issue #2"},
		},
		Literals: []literalRate{
			{Gear: "item", ID: 50, Kind: overrides.RatePPM, Origin: "unknown"},
			{Gear: "set bonus", ID: 7000, Pieces: 4, Kind: overrides.RateChance, Origin: "tbc"},
			{Gear: "set bonus", ID: 7000, Pieces: 2, Kind: overrides.RatePPM, Origin: "vanilla"},
			{Gear: "set bonus", ID: 8000, Pieces: 2, Kind: overrides.RatePPM, Origin: "vanilla"},
		},
		SpellItems:    map[int32][]int32{6: {40, 98}, 7: {45}, 8: {20}, 9: {45}},
		SpellEnchants: map[int32][]int32{6: {300}},
		SpellSets:     map[int32][][2]int32{6: {{7000, 6}, {9000, 2}}},
	}
	db := rateTestDB([]int32{10, 20, 30, 40, 45, 50, 60}, []int32{100, 200, 300}, map[int32]int32{70: 7000})

	rates, err := unvalidatedRates(in, db)
	if err != nil {
		t.Fatal(err)
	}

	type entry struct {
		id, pieces int32
		kind       overrides.RateKind
		origin     string
	}
	entries := func(list []UnvalidatedRate) []entry {
		var got []entry
		for _, rate := range list {
			got = append(got, entry{rate.ID, rate.Pieces, rate.Kind, rate.Origin})
		}
		return got
	}
	for _, tc := range []struct {
		name string
		got  []UnvalidatedRate
		want []entry
	}{
		{"items", rates.Items, []entry{
			{10, 0, overrides.RatePPM, "tbc"},
			{30, 0, overrides.RatePPM, "vanilla"},
			{40, 0, overrides.RatePPM, "tbc"},
			{45, 0, overrides.RateChance, "unknown"},
			{50, 0, overrides.RatePPM, "unknown"},
		}},
		{"enchants", rates.Enchants, []entry{
			{100, 0, overrides.RateChance, "vanilla"},
			{300, 0, overrides.RatePPM, "tbc"},
		}},
		{"set bonuses", rates.SetBonuses, []entry{
			{7000, 2, overrides.RatePPM, "vanilla"},
			{7000, 4, overrides.RateChance, "tbc"},
			{7000, 6, overrides.RatePPM, "tbc"},
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

// Each list renders one entry per line with its kind, origin and name, and a set's bonuses render
// under the set by piece count.
func TestUnvalidatedRatesRender(t *testing.T) {
	rates := UnvalidatedRates{
		Items:    []UnvalidatedRate{{ID: 1, Name: "Blade", Kind: overrides.RatePPM, Origin: "vanilla"}},
		Enchants: []UnvalidatedRate{{ID: 2, Name: "Enchant", Kind: overrides.RateChance, Origin: "tbc"}},
		SetBonuses: []UnvalidatedRate{
			{ID: 3, Pieces: 2, Name: "Set", Kind: overrides.RateChance, Origin: "unknown"},
			{ID: 3, Pieces: 4, Name: "Set", Kind: overrides.RatePPM, Origin: "tbc"},
		},
	}
	var rendered bytes.Buffer
	if err := renderMissingEffects(&rendered, rates, nil); err != nil {
		t.Fatal(err)
	}

	for _, block := range []string{
		"export const UNVALIDATED_RATE_ITEMS = new Map<number, UnvalidatedRate>([\n" +
			"\t[1, { kind: 'ppm', origin: 'vanilla' }], // Blade\n" +
			"])\n",
		"export const UNVALIDATED_RATE_ENCHANTS = new Map<number, UnvalidatedRate>([\n" +
			"\t[2, { kind: 'chance', origin: 'tbc' }], // Enchant\n" +
			"])\n",
		"export const UNVALIDATED_RATE_SET_BONUSES = new Map<number, Map<number, UnvalidatedRate>>([\n" +
			"\t[\n" +
			"\t\t3, // Set\n" +
			"\t\tnew Map([\n" +
			"\t\t\t[2, { kind: 'chance', origin: 'unknown' }],\n" +
			"\t\t\t[4, { kind: 'ppm', origin: 'tbc' }],\n" +
			"\t\t]),\n" +
			"\t],\n" +
			"])\n",
	} {
		if !strings.Contains(rendered.String(), block) {
			t.Errorf("the generated file holds no block\n%s\nin\n%s", block, rendered.String())
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
	if err := renderMissingEffects(&rendered, UnvalidatedRates{}, bonuses); err != nil {
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

// What holds for any shipped rate: its Source is a carry-over, an estimate or a measurement, a literal
// names an origin the notices know, every row names gear the game files carry, and no measured rate is listed.
func TestShippedRatesAreWellFormed(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()
	in := shippedRateInputs(instance)

	knownSource := func(source string) bool {
		return source == "vanilla-carryover" || source == "tbc-carryover" || source == "estimate" || rateOrigin(source) == ""
	}
	db := NewWowDatabase()
	var measuredItems, measuredEnchants []int32
	for _, row := range in.WeaponProcs {
		if !knownSource(row.Source) {
			t.Errorf("item %d's rate states source %q, want a carry-over, an estimate or a measurement", row.ItemID, row.Source)
		}
		if _, ok := instance.Items[int(row.ItemID)]; !ok {
			t.Errorf("item %d's rate names an item the game files do not carry", row.ItemID)
		}
		if rateOrigin(row.Source) == "" {
			measuredItems = append(measuredItems, row.ItemID)
		}
		db.Items[row.ItemID] = &proto.UIItem{Id: row.ItemID}
	}
	for _, row := range in.EnchantProcs {
		if !knownSource(row.Source) {
			t.Errorf("enchant %d's rate states source %q, want a carry-over, an estimate or a measurement", row.EnchantID, row.Source)
		}
		enchant, ok := instance.EnchantsByEffectId[int(row.EnchantID)]
		if !ok || !slices.Contains(enchant.EffectArgs, int(row.SpellID)) {
			t.Errorf("enchant %d's rate names spell %d, which no slot of an enchant the game files carry casts", row.EnchantID, row.SpellID)
		}
		if rateOrigin(row.Source) == "" {
			measuredEnchants = append(measuredEnchants, row.EnchantID)
		}
		db.Enchants[EnchantDBKey{EffectID: row.EnchantID}] = &proto.UIEnchant{EffectId: row.EnchantID}
	}
	for _, row := range in.Spells {
		if spellRateKind(row) != 0 && !knownSource(row.Source) {
			t.Errorf("spell %d's rate states source %q, want a carry-over, an estimate or a measurement", row.SpellID, row.Source)
		}
		for _, itemID := range in.SpellItems[row.SpellID] {
			if spellRateKind(row) != 0 && rateOrigin(row.Source) == "" {
				measuredItems = append(measuredItems, itemID)
			}
			db.Items[itemID] = &proto.UIItem{Id: itemID}
		}
	}
	for _, literal := range in.Literals {
		_, isItem := instance.Items[int(literal.ID)]
		_, isEnchant := instance.EnchantsByEffectId[int(literal.ID)]
		if !slices.Contains([]string{"vanilla", "tbc", "unknown"}, literal.Origin) || literal.Reason == "" ||
			literal.Gear == "item" && !isItem || literal.Gear == "enchant" && !isEnchant {
			t.Errorf("the literal %+v names gear the game files do not carry, no reason or an origin the notices do not know", literal)
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

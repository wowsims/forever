package database

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// The rows a header reads beyond a spell's own, as loadTooltipHeaders gathers them, for these tests alone.
func fixtureHeaders() *tooltipHeaders {
	return &tooltipHeaders{
		ranges: map[int32]headerRange{
			1: {Flags: rangeFlagMelee, MaxFriendly: 5},
			4: {MaxFriendly: 25},
			5: {MaxFriendly: 50000},
			6: {MaxFriendly: 40},
		},
		optional:   map[int32][]int32{1: {0, 4}},
		forms:      map[int32]string{1: "Cat Form", 5: "Bear Form", 8: "Dire Bear Form"},
		items:      map[int32]string{17020: "Arcane Powder", 6265: "Soul Shard"},
		masks:      []itemSubclassMask{{Name: "Melee Weapon", Class: 2, Mask: 173555}, {Name: "Shield", Class: 4, Mask: 96}},
		kinds:      map[[2]int32]string{{2, 15}: "Daggers", {4, 6}: "Shields"},
		tools:      map[int32]spellTools{9: {Categories: []int32{2}}},
		categories: map[int32]string{2: "Earth Totem"},
	}
}

func onNextSwing() [17]uint32 {
	var attr [17]uint32
	attr[dbcenums.ATTR_INDEX_BASE] = dbcenums.ATTR_ON_NEXT_SWING
	return attr
}

func channelled() [17]uint32 {
	var attr [17]uint32
	attr[dbcenums.ATTR_INDEX_EX_1] = dbcenums.ATTR_EX_1_IS_CHANNELLED
	return attr
}

// Each line under a spell's name reads as the game's strings write it, for the costs, ranges, cast times,
// cooldowns, requirements, tools and reagents the game files rows state.
func TestHeaderLinesWriteTheGameStrings(t *testing.T) {
	var passive [17]uint32
	passive[dbcenums.ATTR_INDEX_BASE] = dbcenums.ATTR_PASSIVE
	var casterForm [17]uint32
	casterForm[dbcenums.ATTR_INDEX_EX_2] = dbcenums.ATTR_EX_2_CASTABLE_IN_CASTER_FORM

	for _, c := range []struct {
		name  string
		spell spelldata.Spell
		want  [][2]any
	}{
		{"a passive shows no header", spelldata.Spell{ID: 3, Attr: passive, MaxRange: 30}, nil},
		{"a finisher costs its energy and its combo points, from the least to the most", spelldata.Spell{
			ID: 1, MaxRange: 5, EquipClass: 2, EquipSubclass: 173555,
			Powers: []spelldata.Power{{Type: dbcenums.POWER_ENERGY, Cost: 35}, {Type: dbcenums.POWER_COMBO_POINTS, Cost: 1}},
		}, [][2]any{{"35 Energy / 1 to 5 Combo Points", "Melee Range"}, {"Instant", ""}, {"Requires Melee Weapon", ""}}},
		{"a next-swing ability goes off on the next melee swing, in its forms", spelldata.Spell{
			ID: 2, Attr: onNextSwing(), StanceMask: 1<<4 | 1<<7,
			Powers: []spelldata.Power{{Type: dbcenums.POWER_RAGE, Cost: 150}},
		}, [][2]any{{"15 Rage", ""}, {"Next melee", ""}, {"Requires Bear Form, Dire Bear Form", ""}}},
		{"a spell castable out of its forms requires none", spelldata.Spell{ID: 2, Attr: casterForm, StanceMask: 1}, [][2]any{{"Instant", ""}}},
		{"a share of health is of the maximum, a share of mana of the base", spelldata.Spell{ID: 2, Powers: []spelldata.Power{
			{Type: dbcenums.POWER_HEALTH, CostPct: 25}, {Type: dbcenums.POWER_MANA, CostPct: 13.9},
		}}, [][2]any{{"25% of Maximum Health / 13.9% of base mana", ""}, {"Instant", ""}}},
		{"a channel's drain is per second", spelldata.Spell{ID: 2, Attr: channelled(), Powers: []spelldata.Power{
			{Type: dbcenums.POWER_HEALTH, Cost: 11, PerSecond: 5}, {Type: dbcenums.POWER_MANA, PerSecond: 4},
		}}, [][2]any{{"11 Health, plus 5 per sec / 4 Mana per sec", ""}, {"Channeled", ""}}},
		{"a minimum range comes before the most", spelldata.Spell{ID: 4, MinRange: 8, MaxRange: 25, CooldownMs: 3600000},
			[][2]any{{"", "8 - 25 yd range"}, {"Instant", "1 hour cooldown"}}},
		{"50000 yards is unlimited", spelldata.Spell{ID: 5, MaxRange: 50000, CastTimeMs: 1500, CategoryCooldownMs: 90000},
			[][2]any{{"", "Unlimited range"}, {"1.5 sec cast", "1.5 min cooldown"}}},
		{"a friendly range of its own goes on a line below the enemy one", spelldata.Spell{ID: 6, MaxRange: 20, Attr: onNextSwing()},
			[][2]any{{"", "Enemy: 20 yd range"}, {"", "Friendly: 40 yd range"}, {"Next melee", ""}}},
		{"a negative cast time has no cast line", spelldata.Spell{ID: 2, CastTimeMs: -1000000, CooldownMs: 6000},
			[][2]any{{"", "6 sec cooldown"}}},
		{"a weapon kind no group names is written by its plural", spelldata.Spell{ID: 2, EquipClass: 2, EquipSubclass: 1 << 15},
			[][2]any{{"Instant", ""}, {"Requires Daggers", ""}}},
		{"an enchant's item is the one it goes on", spelldata.Spell{ID: 2, EquipClass: 4, EquipSubclass: 1 << 6,
			Effects: []spelldata.Effect{{Type: dbcenums.E_ENCHANT_ITEM}}}, [][2]any{{"Instant", ""}}},
		{"tools, then reagents on the line below their label, with a count past one", spelldata.Spell{ID: 9, CastTimeMs: 10000,
			Reagents: []spelldata.Reagent{{Item: 17020, Count: 1}, {Item: 6265, Count: 2}}},
			[][2]any{{"10 sec cast", ""}, {"Tools: Earth Totem", ""}, {"Reagents:", ""}, {"Arcane Powder, Soul Shard (2)", ""}}},
	} {
		if got := headerLines(&c.spell, fixtureHeaders(), nil); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %#v, want %#v", c.name, got, c.want)
		}
	}
}

// A talent that changes a cost, a range or a cast time still changes it in the header, on the cost of the first bar
// the spell spends and nothing after it.
func TestHeaderLinesKeepTheTalentModifiers(t *testing.T) {
	const improvedHeroicStrike, hawkEye = 12282, 19498
	flags := func(talent int32) (int32, []int) {
		effect := spelldata.MustFind(talent).EffectN(1)
		mask := make([]int, len(effect.ClassFlags.Mask))
		for i, word := range effect.ClassFlags.Mask {
			mask[i] = int(word)
		}
		return effect.ClassFlags.Family, mask
	}
	instance := dbc.NewDBC()
	family, mask := flags(improvedHeroicStrike)
	instance.Spells[501] = dbc.Spell{ID: 501, SpellClassSet: family, SpellClassMask: mask}
	family, mask = flags(hawkEye)
	instance.Spells[502] = dbc.Spell{ID: 502, SpellClassSet: family, SpellClassMask: mask}
	mods := tooltip.NewModifiers(instance, map[int32]int32{improvedHeroicStrike: 3, hawkEye: 3})

	headers := fixtureHeaders()
	headers.optional[501] = []int32{0, 4}
	headers.ranges[502] = headerRange{MaxFriendly: 35}
	cost := headerLines(&spelldata.Spell{ID: 501, Powers: []spelldata.Power{
		{Type: dbcenums.POWER_RAGE, Cost: 150}, {Type: dbcenums.POWER_COMBO_POINTS, Cost: 1},
	}}, headers, mods)[0][0]
	segments, ok := cost.([]tooltip.Segment)
	if !ok || len(segments) != 4 || segments[1] != " Rage" || segments[2] != " / " || segments[3] != "1 to 5 Combo Points" {
		t.Fatalf("the cost with Improved Heroic Strike is %s, want the rage modified and the combo points as they are", jsonOf(t, cost))
	}
	if got := jsonOf(t, segments[0]); !containsTalent(got, improvedHeroicStrike) {
		t.Errorf("the rage cost %s does not take Improved Heroic Strike", got)
	}

	reach := headerLines(&spelldata.Spell{ID: 502, MinRange: 8, MaxRange: 35}, headers, mods)[0][1]
	if got := jsonOf(t, reach); !containsTalent(got, hawkEye) {
		t.Errorf("the range %s does not take Hawk Eye", got)
	}
}

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func containsTalent(segmentJSON string, talent int) bool {
	var decoded any
	if json.Unmarshal([]byte(segmentJSON), &decoded) != nil {
		return false
	}
	var walk func(v any) bool
	walk = func(v any) bool {
		switch n := v.(type) {
		case map[string]any:
			if id, ok := n["t"].(float64); ok && int(id) == talent {
				return true
			}
			for _, child := range n {
				if walk(child) {
					return true
				}
			}
		case []any:
			for _, child := range n {
				if walk(child) {
					return true
				}
			}
		}
		return false
	}
	return walk(decoded)
}

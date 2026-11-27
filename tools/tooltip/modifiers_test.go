package tooltip

import (
	"encoding/json"
	"math/bits"
	"reflect"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The segment's JSON, as the site reads it.
func segmentJSON(t *testing.T, segment Segment) string {
	t.Helper()
	data, err := json.Marshal(segment)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func compileWith(t *testing.T, provider DBCTooltipDataProvider, spellId int64, text string) []Segment {
	t.Helper()
	segments, err := Compile(text, provider, spellId, nil)
	if err != nil {
		t.Fatal(err)
	}
	return segments
}

// A damage amount keeps its base, its spell power share, its kind and school and the talents' modifiers on it apart,
// for the UI to apply; one on its base points applies there, a damage one to the whole amount.
func TestCompiledDamageAmountCarriesItsKindSchoolAndModifiers(t *testing.T) {
	provider := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 101, SchoolMask: 4},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_SCHOOL_DAMAGE, EffectBasePoints: 100, EffectBonusCoefficient: 0.5}},
	})
	plain := compileWith(t, provider, 101, "$s1 Fire damage.")
	if got, want := segmentJSON(t, plain[0]), `{"abs":true,"n":{"fx":{"b":100,"c":4,"k":"dmg","s":[{"a":[0.5,{"v":"SPFI"}],"o":"*"}]}},"p":0}`; got != want {
		t.Errorf("a fire damage amount compiles to %s, want %s", got, want)
	}

	provider.Modifiers = Modifiers{101: {
		{op: dbcenums.SPELLMOD_DAMAGE, Modifier: Modifier{Talent: 900, Values: []float64{5, 10}}},
		{op: dbcenums.SPELLMOD_EFFECT1, Modifier: Modifier{Talent: 901, Values: []float64{2}, Flat: true}},
		{op: dbcenums.SPELLMOD_EFFECT2, Modifier: Modifier{Talent: 902, Values: []float64{3}, Flat: true}},
		{op: dbcenums.SPELLMOD_DOT, Modifier: Modifier{Talent: 903, Values: []float64{4}}},
	}}
	modified := compileWith(t, provider, 101, "$s1 Fire damage.")
	want := `{"abs":true,"n":{"fx":{"b":100,"c":4,"k":"dmg","m":[{"t":901,"v":[2],"f":true},{"t":900,"v":[5,10],"w":true}],"s":[{"a":[0.5,{"v":"SPFI"}],"o":"*"}]}},"p":0}`
	if got := segmentJSON(t, modified[0]); got != want {
		t.Errorf("with the first effect's and the damage modifiers, the amount compiles to %s, want %s", got, want)
	}
}

// An effect that is neither damage nor healing, and has no modifier on it, stays the number it is.
func TestCompiledOtherAmountStaysANumber(t *testing.T) {
	provider := fixtureProvider(fixtureRow{
		spell:   dbc.Spell{ID: 101},
		effects: []dbc.SpellEffect{{EffectIndex: 0, EffectType: dbcenums.E_APPLY_AURA, EffectAura: dbcenums.A_MOD_DECREASE_SPEED, EffectBasePoints: -40}},
	})
	if got := compileWith(t, provider, 101, "Slows by $s1%."); !reflect.DeepEqual(got, []Segment{"Slows by 40%."}) {
		t.Errorf("a slow compiles to %#v, want the text alone", got)
	}
}

// A duration the talents lengthen is written the way a fixed one is, in the units every value it can take reads in.
func TestCompiledDurationTakesTheDurationModifiers(t *testing.T) {
	durationOf := func(ms int32, flat ...float64) []Segment {
		provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 101, Duration: ms}})
		provider.Modifiers = Modifiers{101: {{op: dbcenums.SPELLMOD_DURATION, Modifier: Modifier{Talent: 900, Values: flat, Flat: true}}}}
		return compileWith(t, provider, 101, "Lasts $d.")
	}
	if got, want := segmentJSON(t, durationOf(10000, 1000, 2000)), `["Lasts ",{"n":{"fx":{"b":10,"m":[{"t":900,"v":[1,2],"f":true}]}},"p":1,"t":true}," sec."]`; got != want {
		t.Errorf("a 10 s duration with 1 and 2 more compiles to %s, want %s", got, want)
	}
	crossing := durationOf(50000, 20000)
	if len(crossing) < 2 {
		t.Fatalf("a duration that can reach a minute compiles to %#v", crossing)
	}
	if _, ok := crossing[1].(map[string]Expr)["if"]; !ok {
		t.Errorf("a duration that can reach a minute compiles to %s, want a branch on the minute", segmentJSON(t, crossing))
	}
}

// A cast time and a cooldown are written in the largest unit they reach, in two significant digits.
func TestHeaderTimesTakeTheLargestUnit(t *testing.T) {
	for _, c := range []struct {
		got, want string
	}{
		{CastTimeText(0), "Instant"},
		{CastTimeText(250), "0.25 sec cast"},
		{CastTimeText(1500), "1.5 sec cast"},
		{CastTimeText(60000), "1 min cast"},
		{CooldownText(0), ""},
		{CooldownText(12500), "12 sec cooldown"},
		{CooldownText(59000), "59 sec cooldown"},
		{CooldownText(90000), "1.5 min cooldown"},
		{CooldownText(3600000), "1 hour cooldown"},
		{CooldownText(7200000), "2 hour cooldown"},
		{CooldownText(172800000), "2 day cooldown"},
	} {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

// A cooldown or cast time the talents change reads at every rank as a fixed one of that length does, with one form
// where every rank takes the same unit and a branch where they do not.
func TestModifiedHeaderTimesReadAsTheFixedOnes(t *testing.T) {
	for _, c := range []struct {
		ms       float64
		flat     []float64
		cast     bool
		branches bool
	}{
		{ms: 6000, flat: []float64{-1000, -2000}},
		{ms: 15400, flat: []float64{-2000}},
		{ms: 12300, flat: []float64{-2600}},
		{ms: 180000, flat: []float64{-30000, -60000}},
		{ms: 70000, flat: []float64{-20000}, branches: true},
		{ms: 300000, flat: []float64{-45000}, branches: true},
		{ms: 1500, flat: []float64{-250}, cast: true, branches: true},
		{ms: 3000, flat: []float64{-500}, cast: true},
		{ms: 1500, flat: []float64{-1500}, cast: true, branches: true},
	} {
		mods := []Modifier{{Talent: 900, Values: c.flat, Flat: true}}
		segments, fixed := ModifiedCooldown(c.ms, mods), CooldownText
		if c.cast {
			segments, fixed = ModifiedCastTime(c.ms, mods), CastTimeText
		}
		first, _ := segments[0].(map[string]Expr)
		if _, ok := first["if"]; ok != c.branches {
			t.Errorf("%v ms less %v compiles to %s, branching %v, want %v", c.ms, c.flat, segmentJSON(t, segments), ok, c.branches)
		}
		for rank := 0; rank <= len(c.flat); rank++ {
			ms := c.ms
			if rank > 0 {
				ms += c.flat[rank-1]
			}
			if got, want := renderJSON(t, segments, player{points: map[int64]int{900: rank}}), fixed(ms); got != want {
				t.Errorf("%v ms less %v at rank %d reads %q, want %q", c.ms, c.flat, rank, got, want)
			}
		}
	}
}

// A passive talent's spell modifier reaches the spells its class flags name, at its ladder's value for each rank,
// and no spell of another family or outside the mask; a talent that is cast reaches none.
func TestNewModifiersReachTheSpellsTheClassFlagsName(t *testing.T) {
	const hawkEye, arcanePower = 19498, 12042
	talent := spelldata.MustFind(hawkEye)
	effect := talent.EffectN(1)
	if !talent.IsPassive() || effect.Aura != dbcenums.A_ADD_FLAT_MODIFIER || dbcenums.SpellModOp(effect.Misc) != dbcenums.SPELLMOD_RANGE {
		t.Fatalf("talent %d is not a passive flat range modifier", hawkEye)
	}
	cast := spelldata.MustFind(arcanePower)
	if cast.IsPassive() || cast.EffectN(1).Aura != dbcenums.A_ADD_PCT_MODIFIER {
		t.Fatalf("talent %d is not a cast percentage modifier", arcanePower)
	}

	mask := func(flags [4]uint32) []int {
		out := make([]int, len(flags))
		for i, word := range flags {
			out[i] = int(word)
		}
		return out
	}
	outside := [4]uint32{}
	for i, word := range effect.ClassFlags.Mask {
		if word != ^uint32(0) {
			outside[i] = 1 << bits.TrailingZeros32(^word)
			break
		}
	}
	instance := dbc.NewDBC()
	instance.Spells[501] = dbc.Spell{ID: 501, SpellClassSet: effect.ClassFlags.Family, SpellClassMask: mask(effect.ClassFlags.Mask)}
	instance.Spells[502] = dbc.Spell{ID: 502, SpellClassSet: effect.ClassFlags.Family, SpellClassMask: mask(outside)}
	instance.Spells[503] = dbc.Spell{ID: 503, SpellClassSet: effect.ClassFlags.Family + 1, SpellClassMask: mask(effect.ClassFlags.Mask)}
	instance.Spells[504] = dbc.Spell{ID: 504, SpellClassSet: cast.EffectN(1).ClassFlags.Family, SpellClassMask: mask(cast.EffectN(1).ClassFlags.Mask)}

	mods := NewModifiers(instance, map[int32]int32{hawkEye: 3, arcanePower: 1})
	ladder := spelldata.Talent(hawkEye, 3)
	want := []Modifier{{Talent: hawkEye, Flat: true}}
	for n := int32(1); n <= 3; n++ {
		want[0].Values = append(want[0].Values, ladder.Rank(n).EffectN(1).Average())
	}
	if got := mods.For(501, dbcenums.SPELLMOD_RANGE); !reflect.DeepEqual(got, want) {
		t.Errorf("the named spell's range takes %+v, want %+v", got, want)
	}
	for _, id := range []int64{502, 503} {
		if got := mods.For(id, dbcenums.SPELLMOD_RANGE); len(got) != 0 {
			t.Errorf("spell %d, which the class flags do not name, takes %+v", id, got)
		}
	}
	if got := mods[504]; len(got) != 0 {
		t.Errorf("the spell a cast talent's buff names takes %+v", got)
	}
}

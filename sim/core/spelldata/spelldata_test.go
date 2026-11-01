package spelldata

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Rows in the shape the game files state them, standing in for the generated store: a ranked
// direct-damage spell with a slow on it and the two ranks above it, a proc aura that fires a
// triggered spell, the spell it fires, and two carrying the fractional EffectBasePointsF about one
// effect in 55 has. Ids ascend, as the generated slice's do.
func fixture() []Spell {
	return []Spell{
		{
			ID:                    116,
			Name:                  "Frostbolt",
			Rank:                  "Rank 1",
			School:                16,
			SpellLevel:            4,
			MaxLevel:              8,
			CastTimeMs:            1500,
			GCDMs:                 1500,
			StartRecoveryCategory: 133,
			DurationMs:            5000,
			Effects: []Effect{
				{SpellID: 116, Index: 0, Type: dbcenums.E_APPLY_AURA, Aura: 33, BasePoints: -40},
				{SpellID: 116, Index: 1, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 19, PPL: 0.5,
					BaseLevel: 4, MaxLevel: 8, Variance: 0.105263, SPCoef: 0.407},
			},
			Powers: []Power{{Type: 0, Cost: 25}},
		},
		{ID: 205, Name: "Frostbolt", Rank: "Rank 2", School: 16},
		{
			ID:          324,
			Name:        "Lightning Shield",
			School:      8,
			DurationMs:  600000,
			ProcCharges: 3,
			ProcFlags:   [2]uint32{0x222A8, 0},
			ICDMs:       3500,
			Effects: []Effect{
				{SpellID: 324, Index: 0, Type: dbcenums.E_APPLY_AURA, Aura: 42, TriggerID: 26364},
			},
			Powers: []Power{{Type: 0, Cost: 750}},
		},
		{
			ID:         700,
			Name:       "Fractional Base",
			SpellLevel: 60,
			Effects: []Effect{
				{SpellID: 700, Index: 0, Type: dbcenums.E_APPLY_AURA, Aura: 13, BasePoints: -58.4697},
			},
		},
		{
			ID:         800,
			Name:       "Fractional Base Per Level",
			SpellLevel: 4,
			MaxLevel:   7,
			Effects: []Effect{
				{SpellID: 800, Index: 0, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 19.7, PPL: 0.5,
					BaseLevel: 4, MaxLevel: 7},
			},
		},
		{ID: 837, Name: "Frostbolt", Rank: "Rank 3", School: 16},
		{
			ID:     26364,
			Name:   "Lightning Shield",
			School: 8,
			Effects: []Effect{
				{SpellID: 26364, Index: 0, Type: dbcenums.E_SCHOOL_DAMAGE, BasePoints: 25},
			},
		},
	}
}

func TestMain(m *testing.M) {
	handTriggers = map[int32][]int32{}
	setSpells(fixture())
	curves = map[int32][][]float64{}
	os.Exit(m.Run())
}

// Rows for one test, with the package's own fixture put back afterwards.
func withRows(t *testing.T, rows []Spell) {
	t.Helper()
	setSpells(rows)
	t.Cleanup(func() { setSpells(fixture()) })
}

func TestFindMisses(t *testing.T) {
	if Find(0) != Nil {
		t.Fatalf("Find(0) = %v, want Nil", Find(0))
	}
	if Find(117) != Nil {
		t.Fatalf("Find(117) = %v, want Nil", Find(117))
	}
	if Find(116).ID != 116 {
		t.Fatalf("Find(116).ID = %d, want 116", Find(116).ID)
	}
}

// 19 base plus 0.5 a level over the four levels between the spell's own 4 and its cap of 8, which is
// where the scaling stops however high the caster is.
func TestEffectAverage(t *testing.T) {
	damage := Find(116).EffectN(2)
	if got := damage.AverageAt(60); got != 21 {
		t.Errorf("Average(60) = %v, want 21", got)
	}
	if got := damage.AverageAt(6); got != 20 {
		t.Errorf("Average(6) = %v, want 20", got)
	}
	if got := damage.AverageAt(2); got != 19 {
		t.Errorf("Average(2) = %v, want 19", got)
	}

	// 21 spread by 10.5% of the base 19 is 21 ± 1.
	if got := damage.MinAt(60); got != 20 {
		t.Errorf("Min(60) = %v, want 20", got)
	}
	if got := damage.MaxAt(60); got != 22 {
		t.Errorf("Max(60) = %v, want 22", got)
	}
}

// A damage amount keeps its fraction, in float32: 19.7 gaining 0.5 over three levels is 21.2, and its
// tooltip ends are the floor, 21. Any other amount is floored whole: the aura's -58.4697 is -59.
func TestEffectAverageKeepsTheFraction(t *testing.T) {
	if got := Find(700).EffectN(1).AverageAt(60); got != -59 {
		t.Errorf("Average(60) of a fractional negative aura base = %v, want -59", got)
	}
	damage := Find(800).EffectN(1)
	if got, want := damage.AverageAt(60), float64(float32(21.2)); got != want {
		t.Errorf("Average(60) of a fractional base gaining 1.5 = %v, want %v", got, want)
	}
	if got, want := damage.AverageAt(4), float64(float32(19.7)); got != want {
		t.Errorf("Average(4) at the spell's own level = %v, want %v", got, want)
	}
	if got, want := damage.MeanAt(60), float64(float32(21.2)); got != want {
		t.Errorf("Mean(60) of an effect that does not roll = %v, want %v", got, want)
	}
	if got, want := [2]float64{damage.MinAt(60), damage.MaxAt(60)}, [2]float64{21, 21}; got != want {
		t.Errorf("Min(60) and Max(60) = %v, want %v", got, want)
	}
}

func TestEffectOutOfRange(t *testing.T) {
	if Find(116).EffectN(3) != NilEffect {
		t.Error("EffectN(3) of a two-effect spell is not NilEffect")
	}
	if Find(116).EffectN(0) != NilEffect {
		t.Error("EffectN(0) is not NilEffect")
	}
	if got := NilEffect.AverageAt(60); got != 0 {
		t.Errorf("NilEffect.Average(60) = %v, want 0", got)
	}
	if got := Nil.CastTime(); got != 0 {
		t.Errorf("Nil.CastTime() = %v, want 0", got)
	}
}

func TestTriggerAndDrivers(t *testing.T) {
	if got := Find(324).EffectN(1).Trigger().ID; got != 26364 {
		t.Errorf("trigger of 324 = %d, want 26364", got)
	}

	drivers := Find(26364).Drivers()
	if len(drivers) != 1 || drivers[0].ID != 324 {
		t.Errorf("drivers of 26364 = %v, want [324]", ids(drivers))
	}

	triggered := Find(324).Triggered()
	if len(triggered) != 1 || triggered[0].ID != 26364 {
		t.Errorf("triggered by 324 = %v, want [26364]", ids(triggered))
	}
	if len(Find(116).Drivers()) != 0 {
		t.Errorf("drivers of 116 = %v, want none", ids(Find(116).Drivers()))
	}
}

func TestTimes(t *testing.T) {
	if got := Find(324).ICD(); got != 3500*time.Millisecond {
		t.Errorf("ICD of 324 = %v, want 3.5s", got)
	}
	if got := Find(116).Duration(); got != 5*time.Second {
		t.Errorf("duration of 116 = %v, want 5s", got)
	}
	if got := Find(116).CastTime(); got != 1500*time.Millisecond {
		t.Errorf("cast time of 116 = %v, want 1.5s", got)
	}

	// The game files' permanent aura.
	permanent := &Spell{DurationMs: -1}
	if got := permanent.Duration(); got != core.NeverExpires {
		t.Errorf("duration of a permanent aura = %v, want NeverExpires", got)
	}
}

func TestPowerCost(t *testing.T) {
	if got := Find(116).PowerCost(0); got != 25 {
		t.Errorf("mana cost of 116 = %v, want 25", got)
	}

	// Rage is on the game files' 0-1000 bar: 300 is 30 rage.
	rager := &Spell{Powers: []Power{{Type: 1, Cost: 300}}}
	if got := rager.PowerCost(1); got != 30 {
		t.Errorf("rage cost = %v, want 30", got)
	}
	if got := Find(116).PowerCost(3); got != 0 {
		t.Errorf("energy cost of a mana spell = %v, want 0", got)
	}
}

func TestCost(t *testing.T) {
	cases := []struct {
		name string
		s    *Spell
		want float64
	}{
		{"no power row", &Spell{}, 0},
		{"first of several powers, mana", &Spell{Powers: []Power{{Type: 0, Cost: 25}, {Type: 1, Cost: 300}}}, 25},
		{"first of several powers, rage", &Spell{Powers: []Power{{Type: 1, Cost: 300}, {Type: 0, Cost: 25}}}, 30},
	}
	for _, c := range cases {
		if got := c.s.Cost(); got != c.want {
			t.Errorf("%s: Cost() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestRankNumber(t *testing.T) {
	cases := []struct {
		rank string
		want int32
	}{
		{"Rank 1", 1},
		{"Rank 12", 12},
		{"", 0},
		{"Passive", 0},
	}
	for _, c := range cases {
		s := &Spell{Rank: c.rank}
		if got := s.RankNumber(); got != c.want {
			t.Errorf("RankNumber() of %q = %d, want %d", c.rank, got, c.want)
		}
	}
}

func TestRankedLadder(t *testing.T) {
	ladder := Ranked(116)
	if got := ladder.Highest().ID; got != 116 {
		t.Errorf("Highest().ID = %d, want 116", got)
	}
	if ladder.Rank(0) != Nil {
		t.Error("rank 0 is not Nil")
	}
	if ladder.Rank(2) != Nil {
		t.Error("a rank the ladder does not have is not Nil")
	}
	if got := ladder.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
	if got := ladder.EffectAt(2).ValueAt(1); got != 19 {
		t.Errorf("EffectAt(2).ValueAt(1) = %v, want 19", got)
	}
	if got := ladder.ValueAt(0); got != 0 {
		t.Errorf("ValueAt(0) = %v, want 0", got)
	}
}

func TestHighestTrained(t *testing.T) {
	const trained, book = 13555, 25295
	withRows(t, []Spell{{ID: 116}, {ID: trained}, {ID: book}})
	if IsBookRank(trained) || !IsBookRank(book) {
		t.Fatalf("IsBookRank says %v for %d and %v for %d, want false and true", IsBookRank(trained), trained, IsBookRank(book), book)
	}

	ladder := Ranked(trained, book)
	if got := ladder.HighestTrained().ID; got != trained {
		t.Errorf("HighestTrained().ID = %d with a book rank on top, want %d", got, trained)
	}
	if got := ladder.Highest().ID; got != book {
		t.Errorf("Highest().ID = %d, want the book rank %d", got, book)
	}
	if got := Ranked(116, trained).HighestTrained().ID; got != trained {
		t.Errorf("HighestTrained().ID = %d on a ladder with no book rank, want %d", got, trained)
	}
	if Ranked(book).HighestTrained() != Nil {
		t.Error("a ladder of only book ranks does not answer Nil")
	}
	if (Ladder{}).HighestTrained() != Nil {
		t.Error("an empty ladder does not answer Nil")
	}
}

// Each rank of a trait talent is a copy carrying the curve's value, so the store's own row keeps the
// game files'.
func TestTalentLadder(t *testing.T) {
	curves[116] = [][]float64{{-40, -45, -50}, {19, 20, 21}}
	defer delete(curves, 116)

	ladder := Talent(116, 3)
	if got := ladder.Rank(2).EffectN(2).BaseValue(); got != 20 {
		t.Errorf("rank 2 damage = %v, want 20", got)
	}
	if got := ladder.EffectAt(1).ValueAt(3); got != -50 {
		t.Errorf("rank 3 slow = %v, want -50", got)
	}
	if got := Find(116).EffectN(2).BaseValue(); got != 19 {
		t.Errorf("the store's own row reads %v, want the game files' 19", got)
	}
}

// An effect the curve has no row for keeps the spell's base value.
func TestTalentLadderWithoutCurve(t *testing.T) {
	ladder := Talent(116, 2)
	if got := ladder.Rank(2).EffectN(2).BaseValue(); got != 19 {
		t.Errorf("rank 2 damage without a curve = %v, want the base 19", got)
	}
}

// A curve that does not state every rank of the talent it belongs to.
func TestTalentCurveLengthPanics(t *testing.T) {
	curves[116] = [][]float64{{-40, -45, -50}, {19, 20}}
	defer delete(curves, 116)

	requirePanic(t, "spelldata: spell 116 curve row 1 has 2 ranks, want 3", func() { Talent(116, 3) })
}

func TestLadderContains(t *testing.T) {
	withRows(t, []Spell{{ID: 116}, {ID: 205}, {ID: 837}})

	ladder := Ranked(116, 205)
	if !ladder.Contains(&core.Spell{ActionID: core.ActionID{SpellID: 205}}) {
		t.Error("rank 2 is not contained")
	}
	if ladder.Contains(&core.Spell{ActionID: core.ActionID{SpellID: 837}}) {
		t.Error("a spell outside the ladder is contained")
	}
}

// Find binary searches, so the store refuses rows it could not search.
func TestOutOfOrderIDsPanic(t *testing.T) {
	defer setSpells(fixture())

	requirePanic(t, "spelldata: spell ids are out of order at 1: 116 after 324", func() {
		setSpells([]Spell{{ID: 324}, {ID: 116}})
	})
	requirePanic(t, "spelldata: spell ids are out of order at 1: 116 after 116", func() {
		setSpells([]Spell{{ID: 116}, {ID: 116}})
	})
}

func TestMustFindPanics(t *testing.T) {
	requirePanic(t, "spelldata: spell 1 is not in the store", func() { MustFind(1) })
}

func TestEffectPanics(t *testing.T) {
	ambiguous := &Spell{ID: 7, Effects: []Effect{{Aura: 42, Misc: 3}, {Aura: 42, Misc: 3}}}
	requirePanic(t, "has 2 effects with aura 42 misc 3", func() { ambiguous.Effect(42, 3) })
	requirePanic(t, "has no effect with aura 99 misc 0", func() { ambiguous.Effect(99, 0) })

	// Two effects and none of them named: the ladder refuses to pick.
	requirePanic(t, "name the one you mean", func() { Ranked(116).ValueAt(1) })
}

func TestAttributes(t *testing.T) {
	refunding := &Spell{Attr: [17]uint32{1: dbcenums.ATTR_EX_1_DISCOUNT_POWER_ON_MISS}}
	if !refunding.RefundsOnMiss() || refunding.MissRefund() != 0.8 {
		t.Error("a spell flagged Discount Power On Miss does not refund")
	}
	if Find(116).RefundsOnMiss() || Find(116).MissRefund() != 0 {
		t.Error("an unflagged spell refunds")
	}

	channel := &Spell{Attr: [17]uint32{1: dbcenums.ATTR_EX_1_IS_SELF_CHANNELLED}}
	if !channel.IsChanneled() {
		t.Error("a self-channelled spell does not read as channeled")
	}
	if Nil.IsPassive() || Nil.HasAttr(20, 1) {
		t.Error("the zero spell reads an attribute")
	}
}

// The three ticks the flags pick, without building a Dot to compare function pointers with.
func TestTickOutcomeKind(t *testing.T) {
	cases := []struct {
		canCrit, magic bool
		want           int
	}{
		{true, true, tickOutcomeMagicCrit},
		{true, false, tickOutcomePhysicalCrit},
		{false, true, tickOutcomePlain},
		{false, false, tickOutcomePlain},
	}
	for _, c := range cases {
		if got := tickOutcomeKind(c.canCrit, c.magic); got != c.want {
			t.Errorf("tickOutcomeKind(%v, %v) = %d, want %d", c.canCrit, c.magic, got, c.want)
		}
	}
}

func ids(spells []*Spell) []int32 {
	out := make([]int32, len(spells))
	for i, s := range spells {
		out[i] = s.ID
	}
	return out
}

func requirePanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic, want one mentioning %q", want)
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, want) {
			t.Fatalf("panic %v, want one mentioning %q", r, want)
		}
	}()
	fn()
}

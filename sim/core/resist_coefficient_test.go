package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestAverageResistIgnoresTargetLevel(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60}
	target := func(shadowResistance float64) *Unit {
		unit := &Unit{Type: EnemyUnit, Level: 63}
		unit.stats[stats.ShadowResistance] = shadowResistance
		return unit
	}

	for _, tc := range []struct {
		name       string
		resistance float64
		tick       bool
		flags      SpellFlag
		want       float64
	}{
		{"no resistance", 0, false, 0, 0},
		{"no resistance, tick", 0, true, 0, 0},
		{"100 resistance", 100, false, 0, 100.0 / 640},
		{"100 resistance, tick", 100, true, 0, 100.0 / 640 / 10},
		{"100 resistance, channel tick", 100, true, SpellFlagChanneled, 100.0 / 640},
		{"100 resistance, tick with full resistance", 100, true, SpellFlagFullResistanceTicks, 100.0 / 640},
	} {
		spell := &Spell{SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: tc.flags}
		if got := target(tc.resistance).averageResist(spell, attacker, tc.tick); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%s at +3: average resist %v, want %v", tc.name, got, tc.want)
		}
	}

	spell := &Spell{SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: SpellFlagBinary}
	if got, want := target(100).binaryHitChance(spell, attacker), 1-100.0/640; math.Abs(got-want) > 1e-12 {
		t.Errorf("binary hit chance against 100 resistance at +3 = %v, want %v", got, want)
	}
}

// Spell Piercing beyond the target's resistance is a negative average resist, which the roll pays out
// as bonus damage in 10% steps, and never as a resisted outcome.
func TestSpellPiercingBeyondResistanceAddsDamage(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attacker.stats[stats.SpellPiercing] = 15
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolArcane, SchoolIndex: stats.SchoolIndexArcane}

	a := 15.0 / (15 + 540)
	if got := defender.averageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("average resist with 15 piercing and no resistance = %v, want %v", got, -a)
	}

	debuffed := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	debuffed.stats[stats.ArcaneResistance] = -75
	if got := debuffed.averageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("average resist with 15 piercing against -75 resistance = %v, want %v as against 0", got, -a)
	}

	sim := &Simulation{rand: NewSplitMix(1)}
	attackTable := NewAttackTable(attacker, defender)
	counts := map[float64]int{}
	for range 100000 {
		multiplier, outcome := spell.ResistanceMultiplier(sim, false, attackTable)
		if outcome != OutcomeEmpty {
			t.Fatalf("a piercing bonus came with outcome %v", outcome)
		}
		counts[math.Round(multiplier*10)/10]++
	}
	for multiplier, want := range map[float64]float64{1: 1 - 7.5*a, 1.1: 5 * a, 1.2: 2.5 * a} {
		if got := float64(counts[multiplier]) / 100000; math.Abs(got-want) > 0.005 {
			t.Errorf("share of hits at x%v = %.4f, want %.4f", multiplier, got, want)
		}
	}
	if len(counts) != 3 {
		t.Errorf("multipliers rolled: %v, want only 1, 1.1 and 1.2", counts)
	}
}

// The roll spreads around the average resist in 10% steps: 1 - 7.5a, 5a and 2.5a on 0, 10 and 20% up
// to a = 10%, and 0.5 - 2.5|x - a| on each step x within 20% of a above it.
func TestPartialResistRollSteps(t *testing.T) {
	for _, tc := range []struct {
		a    float64
		want map[int]float64
	}{
		{0.06, map[int]float64{0: 0.55, 1: 0.30, 2: 0.15}},
		{0.25, map[int]float64{1: 0.125, 2: 0.375, 3: 0.375, 4: 0.125}},
		{0.30, map[int]float64{2: 0.25, 3: 0.5, 4: 0.25}},
	} {
		sim := &Simulation{rand: NewSplitMix(7)}
		counts := map[int]int{}
		for range 100000 {
			counts[rollResistTenths(sim, tc.a)]++
		}
		for tenths, want := range tc.want {
			if got := float64(counts[tenths]) / 100000; math.Abs(got-want) > 0.005 {
				t.Errorf("a = %v: share of %d0%% resists = %.4f, want %.4f", tc.a, tenths, got, want)
			}
		}
		for tenths := range counts {
			if _, ok := tc.want[tenths]; !ok {
				t.Errorf("a = %v: rolled %d0%%, outside the steps around the average", tc.a, tenths)
			}
		}
	}
}

func TestPartialResistOutcomeCarriesTheTenths(t *testing.T) {
	for tenths := 1; tenths <= 10; tenths++ {
		outcome := OutcomeHit | partialResistOutcome(tenths)
		if got := outcome.ResistedTenths(); got != tenths {
			t.Errorf("ResistedTenths of a %d0%% resist = %d", tenths, got)
		}
		if !outcome.Matches(OutcomePartial) || !outcome.Matches(OutcomeHit) {
			t.Errorf("a %d0%% resist outcome lost its bits: %b", tenths, outcome)
		}
	}
	if got := (OutcomeCrit | partialResistOutcome(3)).String(); got != "Crit (30% Resist)" {
		t.Errorf("a 30%% resisted crit reads %q", got)
	}
}

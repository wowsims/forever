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
		{"100 resistance", 100, false, 0, 100.0 / (100 + 499)},
		{"100 resistance, tick", 100, true, 0, 100.0 / (100 + 499) / 10},
		{"100 resistance, channel tick", 100, true, SpellFlagChanneled, 100.0 / (100 + 499) / 10},
		{"100 resistance, tick with full resistance", 100, true, SpellFlagFullResistanceTicks, 100.0 / (100 + 499)},
		{"100 resistance, channel tick with full resistance", 100, true, SpellFlagChanneled | SpellFlagFullResistanceTicks, 100.0 / (100 + 499)},
		{"100 resistance, hit flagged for a tenth", 100, false, SpellFlagTenthResistance, 100.0 / (100 + 499) / 10},
	} {
		spell := &Spell{SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: tc.flags}
		if got := target(tc.resistance).averageResist(spell, attacker, tc.tick); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%s at +3: average resist %v, want %v", tc.name, got, tc.want)
		}
	}

	spell := &Spell{SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: SpellFlagBinary}
	if got, want := target(100).binaryHitChance(spell, attacker), 1-100.0/(100+499); math.Abs(got-want) > 1e-12 {
		t.Errorf("binary hit chance against 100 resistance at +3 = %v, want %v", got, want)
	}
}

// Spell Piercing beyond the target's resistance is a negative average resist, which the roll pays out
// as bonus damage in 10% steps, with an outcome carrying the negative percent. A hit at a tenth of the
// share takes no bonus, and one at the full share takes it whole.
func TestSpellPiercingBeyondResistanceAddsDamage(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attacker.stats[stats.SpellPiercing] = 15
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolArcane, SchoolIndex: stats.SchoolIndexArcane}

	a := 15.0 / (15 + 499)
	if got := defender.averageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("average resist with 15 piercing and no resistance = %v, want %v", got, -a)
	}

	debuffed := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	debuffed.stats[stats.ArcaneResistance] = -75
	if got := debuffed.averageResist(spell, attacker, false); math.Abs(got+a) > 1e-12 {
		t.Errorf("average resist with 15 piercing against -75 resistance = %v, want %v as against 0", got, -a)
	}

	for _, tc := range []struct {
		name       string
		flags      SpellFlag
		isPeriodic bool
		want       float64
	}{
		{"a tick at a tenth", 0, true, 0},
		{"a hit flagged for a tenth", SpellFlagTenthResistance, false, 0},
		{"a tick at the full share", SpellFlagFullResistanceTicks, true, -a},
		{"a channel tick", SpellFlagChanneled, true, 0},
		{"a channel tick at the full share", SpellFlagChanneled | SpellFlagFullResistanceTicks, true, -a},
	} {
		spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolArcane, SchoolIndex: stats.SchoolIndexArcane, Flags: tc.flags}
		if got := defender.averageResist(spell, attacker, tc.isPeriodic); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("average resist of %s with 15 piercing and no resistance = %v, want %v", tc.name, got, tc.want)
		}
	}

	sim := &Simulation{rand: NewSplitMix(1)}
	attackTable := NewAttackTable(attacker, defender)
	counts := map[float64]int{}
	for range 100000 {
		multiplier, outcome := spell.ResistanceMultiplier(sim, false, attackTable)
		if want := int(math.Round((1 - multiplier) * 100)); outcome.ResistedPercent() != want {
			t.Fatalf("a hit at x%v came with a %d%% resist outcome, want %d%%", multiplier, outcome.ResistedPercent(), want)
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

// Spell Piercing beyond the resistance gives no bonus while its share is under 1%, binary spells
// included, and from 1% on the share takes the piercing in full.
func TestSpellPiercingUnderAOnePercentShareGivesNoBonus(t *testing.T) {
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attacker := func(piercing float64) *Unit {
		unit := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
		unit.stats[stats.SpellPiercing] = piercing
		return unit
	}

	under := attacker(5)
	for _, flags := range []SpellFlag{0, SpellFlagBinary} {
		spell := &Spell{Unit: under, SpellSchool: SpellSchoolArcane, SchoolIndex: stats.SchoolIndexArcane, Flags: flags}
		if got := defender.averageResist(spell, under, false); got != 0 {
			t.Errorf("flags %v: average resist with 5 piercing and no resistance = %v, want 0", flags, got)
		}
		sim := &Simulation{rand: NewSplitMix(1)}
		attackTable := NewAttackTable(under, defender)
		for range 10000 {
			if multiplier, outcome := spell.ResistanceMultiplier(sim, false, attackTable); multiplier != 1 || outcome != OutcomeEmpty {
				t.Fatalf("flags %v: a hit with 5 piercing and no resistance came at x%v with outcome %v, want x1 and none", flags, multiplier, outcome)
			}
		}
	}

	over := attacker(6)
	spell := &Spell{Unit: over, SpellSchool: SpellSchoolArcane, SchoolIndex: stats.SchoolIndexArcane}
	if got, want := defender.averageResist(spell, over, false), -6.0/(6+499); math.Abs(got-want) > 1e-12 {
		t.Errorf("average resist with 6 piercing and no resistance = %v, want %v", got, want)
	}
}

// A spell flagged SpellFlagNoSpellPiercing takes none of the caster's Spell Piercing: it gets no bonus
// where the target has no resistance and meets the whole resistance where it has some.
func TestSpellFlaggedNoSpellPiercingMeetsTheWholeResistance(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attacker.stats[stats.SpellPiercing] = 15
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolFire, SchoolIndex: stats.SchoolIndexFire, Flags: SpellFlagNoSpellPiercing}

	for _, resistance := range []float64{0, 100} {
		defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
		defender.stats[stats.FireResistance] = resistance
		if got, want := defender.averageResist(spell, attacker, false), resistance/(resistance+499); math.Abs(got-want) > 1e-12 {
			t.Errorf("average resist with 15 piercing against %v resistance = %v, want %v", resistance, got, want)
		}
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
			counts[rollResistTenths(sim, resistTenthsChances(tc.a))]++
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

func TestPartialResistOutcomeCarriesThePercent(t *testing.T) {
	for _, percent := range []int{10, 20, 30, 40, 50, 60, 70, 75} {
		outcome := OutcomeHit | partialResistOutcome(float64(percent)/100)
		if got := outcome.ResistedPercent(); got != percent {
			t.Errorf("ResistedPercent of a %d%% resist = %d", percent, got)
		}
		if !outcome.Matches(OutcomePartial) || !outcome.Matches(OutcomeHit) {
			t.Errorf("a %d%% resist outcome lost its bits: %b", percent, outcome)
		}
	}
	if got := (OutcomeCrit | partialResistOutcome(0.3)).String(); got != "Crit (30% Resist)" {
		t.Errorf("a 30%% resisted crit reads %q", got)
	}
	if got := (OutcomeHit | partialResistOutcome(0.75)).String(); got != "Hit (75% Resist)" {
		t.Errorf("a 75%% resisted hit reads %q", got)
	}
}

// A Spell Piercing bonus carries its percent as a negative resist and counts as a pierced hit, not a resisted one.
func TestPartialResistBonusOutcomeCarriesANegativePercent(t *testing.T) {
	for _, percent := range []int{10, 20, 30} {
		result := &SpellResult{Outcome: OutcomeHit | partialResistOutcome(-float64(percent)/100)}
		if got := result.Outcome.ResistedPercent(); got != -percent {
			t.Errorf("ResistedPercent of a %d%% bonus = %d, want %d", percent, got, -percent)
		}
		if !result.DidPierce() {
			t.Errorf("a %d%% bonus does not count as a pierced hit", percent)
		}
		if result.DidResist() {
			t.Errorf("a %d%% bonus counts as a resisted hit", percent)
		}
	}
	if got := (OutcomeCrit | partialResistOutcome(-0.2)).String(); got != "Crit (20% Pierced)" {
		t.Errorf("a crit with a 20%% bonus reads %q", got)
	}
}

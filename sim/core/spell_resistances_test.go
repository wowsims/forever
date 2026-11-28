package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Piercing for one school lowers the resistance a spell of that school meets, and leaves the other
// schools where SpellPiercing alone puts them.
func TestSchoolSpellPiercingLowersOnlyItsSchool(t *testing.T) {
	defender := Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.FireResistance] = 100
	defender.stats[stats.FrostResistance] = 100

	fire := &Spell{SpellSchool: SpellSchoolFire, SchoolIndex: stats.SchoolIndexFire}
	frost := &Spell{SpellSchool: SpellSchoolFrost, SchoolIndex: stats.SchoolIndexFrost}

	coeffs := func(generic, firePiercing float64) (float64, float64) {
		attacker := Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
		attacker.stats[stats.SpellPiercing] = generic
		attacker.PseudoStats.SchoolSpellPiercing[stats.SchoolIndexFire] = firePiercing
		return defender.averageResist(fire, &attacker, false), defender.averageResist(frost, &attacker, false)
	}

	_, frostBare := coeffs(0, 0)
	fireResist, frostResist := coeffs(0, 30)
	if want := 70.0 / (70 + 499); !WithinToleranceFloat64(want, fireResist, 1e-9) {
		t.Errorf("fire average resist with 30 fire piercing = %v, want %v", fireResist, want)
	}
	if frostResist != frostBare {
		t.Errorf("frost average resist with 30 fire piercing = %v, want %v as without it", frostResist, frostBare)
	}

	fireBoth, frostBoth := coeffs(20, 30)
	if want := 50.0 / (50 + 499); !WithinToleranceFloat64(want, fireBoth, 1e-9) {
		t.Errorf("fire average resist with 20 piercing and 30 fire piercing = %v, want %v", fireBoth, want)
	}
	if want := 80.0 / (80 + 499); !WithinToleranceFloat64(want, frostBoth, 1e-9) {
		t.Errorf("frost average resist with 20 piercing and 30 fire piercing = %v, want %v", frostBoth, want)
	}
}

// A Frostfire spell meets the lower of the target's fire and frost resistance after each school's
// piercing, so piercing beyond both makes the share negative; a mask with Holy meets none.
func TestMultiSchoolSpellMeetsItsLowestResistance(t *testing.T) {
	defender := Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.FireResistance] = 100
	defender.stats[stats.FrostResistance] = 40
	attacker := Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	frostfire := &Spell{SpellSchool: SpellSchoolFrostfire, SchoolIndex: stats.SchoolIndexFire}
	holyfire := &Spell{SpellSchool: SpellSchoolHoly | SpellSchoolFire, SchoolIndex: stats.SchoolIndexFire}

	cases := []struct {
		name                   string
		piercing, firePiercing float64
		want                   float64
	}{
		{"no piercing", 0, 0, 40.0 / (40 + 499)},
		{"fire after piercing below frost", 0, 70, 30.0 / (30 + 499)},
		{"piercing beyond both", 50, 0, -10.0 / (10 + 499)},
	}
	for _, c := range cases {
		attacker.stats[stats.SpellPiercing] = c.piercing
		attacker.PseudoStats.SchoolSpellPiercing[stats.SchoolIndexFire] = c.firePiercing
		if got := defender.averageResist(frostfire, &attacker, false); !WithinToleranceFloat64(c.want, got, 1e-9) {
			t.Errorf("%s: Frostfire average resist = %v, want %v", c.name, got, c.want)
		}
		if got := defender.averageResist(holyfire, &attacker, false); got != 0 {
			t.Errorf("%s: Holy and Fire average resist = %v, want 0", c.name, got)
		}
	}
}

// A binary spell's hits and ticks are never partially resisted once it lands, nor is a spell flagged
// SpellFlagNoPartialResist. Spell Piercing beyond the resistance gives a binary spell's hits and
// full-share ticks the bonus, and its tenth-share ticks and a spell flagged SpellFlagNoPartialResist none.
func TestBinaryAndNoPartialResistKeepTheWholeHit(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attackTable := NewAttackTable(attacker, defender)

	partials := func(flags SpellFlag, isPeriodic bool) int {
		sim := &Simulation{rand: NewSplitMix(1)}
		spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow, Flags: flags}
		n := 0
		for range 1000 {
			if multiplier, outcome := spell.ResistanceMultiplier(sim, isPeriodic, attackTable); multiplier != 1 || outcome != OutcomeEmpty {
				n++
			}
		}
		return n
	}

	defender.stats[stats.ShadowResistance] = 150
	if n := partials(0, false); n == 0 {
		t.Errorf("a plain shadow spell was never partially resisted in 1000 hits against 150 shadow resistance")
	}
	for _, tc := range []struct {
		name       string
		flags      SpellFlag
		isPeriodic bool
	}{
		{"a spell flagged SpellFlagNoPartialResist", SpellFlagNoPartialResist, false},
		{"a binary spell's hit", SpellFlagBinary, false},
		{"a binary spell's tick", SpellFlagBinary, true},
	} {
		if n := partials(tc.flags, tc.isPeriodic); n != 0 {
			t.Errorf("%s was partially resisted %d times in 1000 against 150 shadow resistance, want 0", tc.name, n)
		}
	}

	defender.stats[stats.ShadowResistance] = 0
	attacker.stats[stats.SpellPiercing] = 100
	if n := partials(0, false); n == 0 {
		t.Errorf("a plain shadow spell never gained a Spell Piercing bonus in 1000 hits at 100 piercing")
	}
	if n := partials(SpellFlagBinary, false); n == 0 {
		t.Errorf("a binary spell's hit never gained a Spell Piercing bonus in 1000 hits at 100 piercing")
	}
	if n := partials(SpellFlagBinary|SpellFlagFullResistanceTicks, true); n == 0 {
		t.Errorf("a binary spell's full-share tick never gained a Spell Piercing bonus in 1000 ticks at 100 piercing")
	}
	for _, tc := range []struct {
		name       string
		flags      SpellFlag
		isPeriodic bool
	}{
		{"a binary spell's tick at a tenth", SpellFlagBinary, true},
		{"a binary spell flagged SpellFlagNoPartialResist", SpellFlagBinary | SpellFlagNoPartialResist, false},
	} {
		if n := partials(tc.flags, tc.isPeriodic); n != 0 {
			t.Errorf("%s gained a Spell Piercing bonus %d times in 1000 at 100 piercing, want 0", tc.name, n)
		}
	}
}

// Against a resistance far above the caster's level a plain hit always lands and loses 75%, the top step,
// and its expected damage is a quarter of the unresisted one. A binary spell's hit roll resists it whole:
// a miss with no hit or crit counted, and no expected damage.
func TestPartialResistStopsAt75PercentAndBinaryResistsWhole(t *testing.T) {
	sim, fw := setupFakeWarlockSim(proto.Race_RaceHuman, &proto.IndividualBuffs{})
	target := fw.CurrentTarget
	spell := fw.Corruption
	expected := func() float64 {
		return spell.expectedDamage(sim, target, func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
			return spell.CalcDamage(sim, target, 100, spell.OutcomeExpectedMagicHitAndCrit)
		}, false)
	}

	unresisted := expected()
	target.stats[stats.ShadowResistance] = 1e12
	for range 200 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		if !result.Landed() || result.Outcome.ResistedPercent() != 75 || result.ArmorAndResistanceMultiplier != 0.25 {
			t.Fatalf("a plain hit against a resistance of 1e12 landed %v, resisted %d%% with multiplier %v, want a landed hit losing 75%%",
				result.Landed(), result.Outcome.ResistedPercent(), result.ArmorAndResistanceMultiplier)
		}
	}
	if got := expected(); !WithinToleranceFloat64(0.25*unresisted, got, 1e-9) {
		t.Errorf("the expected damage of a plain hit against a resistance of 1e12 is %v, want a quarter of %v", got, unresisted)
	}

	spell.Flags |= SpellFlagBinary
	before := spell.SpellMetrics[target.UnitIndex]
	for range 200 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
		if result.Landed() || result.Outcome != OutcomeMiss || result.Damage != 0 {
			t.Fatalf("a binary hit against a resistance of 1e12 landed %v with outcome %v and %v damage, want a miss with none",
				result.Landed(), result.Outcome, result.Damage)
		}
	}
	after := spell.SpellMetrics[target.UnitIndex]
	if after.Misses-before.Misses != 200 || after.Hits != before.Hits || after.Crits != before.Crits || after.ResistedHits != before.ResistedHits {
		t.Errorf("200 fully resisted binary hits added %d misses, %d hits, %d crits and %d resisted hits, want 200 misses only",
			after.Misses-before.Misses, after.Hits-before.Hits, after.Crits-before.Crits, after.ResistedHits-before.ResistedHits)
	}
	if got := expected(); got > 1e-6 {
		t.Errorf("the expected damage of a binary hit against a resistance of 1e12 is %v, want 0 up to the average resist left", got)
	}
}

// An expected-damage call takes the mean of the roll, the 75% top step included, and draws no random
// number, so the next real roll is the one it would have been without the call. A real hit still rolls.
func TestExpectedDamageUsesTheRollsMeanWithoutRolling(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attackTable := NewAttackTable(attacker, defender)
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow}

	for _, resistance := range []float64{150, 499, 1500, 5000, -100} {
		defender.stats[stats.ShadowResistance] = max(resistance, 0)
		attacker.stats[stats.SpellPiercing] = 0
		if resistance < 0 {
			attacker.stats[stats.SpellPiercing] = -resistance
		}

		withCall := &Simulation{rand: NewSplitMix(7)}
		without := &Simulation{rand: NewSplitMix(7)}

		spell.expecting = true
		multiplier, outcome := spell.ResistanceMultiplier(withCall, false, attackTable)
		spell.expecting = false
		if outcome != OutcomeEmpty {
			t.Errorf("resistance %v: an expected-damage call came with outcome %v, want none", resistance, outcome)
		}
		if got, want := withCall.RandomFloat("Partial Resist"), without.RandomFloat("Partial Resist"); got != want {
			t.Errorf("resistance %v: an expected-damage call moved the random stream, next roll %v, want %v", resistance, got, want)
		}

		sum, distinct := 0.0, map[float64]bool{}
		for range 100000 {
			real, _ := spell.ResistanceMultiplier(withCall, false, attackTable)
			sum += real
			distinct[real] = true
		}
		if mean := sum / 100000; math.Abs(mean-multiplier) > 0.002 {
			t.Errorf("resistance %v: expected multiplier %v, want the mean of the real rolls %v", resistance, multiplier, mean)
		}
		if len(distinct) < 2 && resistance < 1500 {
			t.Errorf("resistance %v: a real hit rolled only %v", resistance, distinct)
		}
	}
}

// Around an average resist of 74% the roll's 80 and 90% steps land on 75%, the top step, beside the
// 60 and 70% the triangle gives.
func TestPartialResistAbove70PercentLandsOn75(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.ShadowResistance] = 0.74 * 499 / 0.26
	attackTable := NewAttackTable(attacker, defender)
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow}

	sim := &Simulation{rand: NewSplitMix(3)}
	counts := map[int]int{}
	for range 100000 {
		_, outcome := spell.ResistanceMultiplier(sim, false, attackTable)
		counts[outcome.ResistedPercent()]++
	}
	want := map[int]float64{60: 0.15, 70: 0.40, 75: 0.45}
	for percent, share := range want {
		if got := float64(counts[percent]) / 100000; math.Abs(got-share) > 0.005 {
			t.Errorf("share of %d%% resists = %.4f, want %.4f", percent, got, share)
		}
	}
	for percent := range counts {
		if _, ok := want[percent]; !ok {
			t.Errorf("rolled a %d%% resist, outside 60, 70 and 75%%", percent)
		}
	}
}

package core

import (
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
	if want := 70.0 / (70 + 540); !WithinToleranceFloat64(want, fireResist, 1e-9) {
		t.Errorf("fire average resist with 30 fire piercing = %v, want %v", fireResist, want)
	}
	if frostResist != frostBare {
		t.Errorf("frost average resist with 30 fire piercing = %v, want %v as without it", frostResist, frostBare)
	}

	fireBoth, frostBoth := coeffs(20, 30)
	if want := 50.0 / (50 + 540); !WithinToleranceFloat64(want, fireBoth, 1e-9) {
		t.Errorf("fire average resist with 20 piercing and 30 fire piercing = %v, want %v", fireBoth, want)
	}
	if want := 80.0 / (80 + 540); !WithinToleranceFloat64(want, frostBoth, 1e-9) {
		t.Errorf("frost average resist with 20 piercing and 30 fire piercing = %v, want %v", frostBoth, want)
	}
}

// Against 150 resistance a plain spell is often partially resisted, as are a binary spell's hits and
// ticks once it lands, and one flagged SpellFlagNoPartialResist never is, binary or not.
func TestNoPartialResistKeepsTheWholeHit(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender.stats[stats.ShadowResistance] = 150
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

	if n := partials(0, false); n == 0 {
		t.Errorf("a plain shadow spell was never partially resisted in 1000 hits against 150 shadow resistance")
	}
	if n := partials(SpellFlagNoPartialResist, false); n != 0 {
		t.Errorf("a spell flagged SpellFlagNoPartialResist was partially resisted %d times in 1000 hits, want 0", n)
	}
	if n := partials(SpellFlagBinary, false); n == 0 {
		t.Errorf("a binary spell's direct hit was never partially resisted in 1000 against 150 shadow resistance")
	}
	if n := partials(SpellFlagBinary, true); n == 0 {
		t.Errorf("a binary spell's tick was never partially resisted in 1000 against 150 shadow resistance")
	}
	if n := partials(SpellFlagBinary|SpellFlagNoPartialResist, false); n != 0 {
		t.Errorf("a binary spell flagged SpellFlagNoPartialResist was partially resisted %d times in 1000 hits, want 0", n)
	}
}

// A hit the target resists in full is a miss: it is not landed, counts as a miss, and no hit or crit is counted.
func TestFullResistIsAMiss(t *testing.T) {
	sim, fw := setupFakeWarlockSim(proto.Race_RaceHuman, &proto.IndividualBuffs{})
	target := fw.CurrentTarget
	spell := fw.Corruption
	target.stats[stats.ShadowResistance] = 1e12

	for range 200 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
		if result.Landed() || result.Outcome != OutcomeMiss || result.Damage != 0 {
			t.Fatalf("a fully resisted hit landed %v with outcome %v and %v damage, want a miss with none",
				result.Landed(), result.Outcome, result.Damage)
		}
	}
	metrics := spell.SpellMetrics[target.UnitIndex]
	if metrics.Misses != 200 || metrics.Hits != 0 || metrics.Crits != 0 || metrics.ResistedHits != 0 {
		t.Errorf("200 fully resisted hits counted %d misses, %d hits, %d crits and %d resisted hits, want 200 misses only",
			metrics.Misses, metrics.Hits, metrics.Crits, metrics.ResistedHits)
	}

	expected := spell.expectedDamage(sim, target, func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
		return spell.CalcDamage(sim, target, 100, spell.OutcomeExpectedMagicHitAndCrit)
	}, false)
	if expected > 1e-6 {
		t.Errorf("the expected damage of a fully resisted hit is %v, want 0 up to the average resist left at 1e12", expected)
	}
	if got := spell.SpellMetrics[target.UnitIndex].Misses; got != 200 {
		t.Errorf("an expected-damage call against a full resist counted %d misses, want 200 and none added", got)
	}
}

// An expected-damage call takes the average resist as the multiplier and draws no random number, so
// the next real roll is the one it would have been without the call. A real hit still rolls.
func TestExpectedDamageUsesAverageResistWithoutRolling(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: EnemyUnit, Level: 60, PseudoStats: stats.NewPseudoStats()}
	attackTable := NewAttackTable(attacker, defender)
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolShadow, SchoolIndex: stats.SchoolIndexShadow}

	for _, resistance := range []float64{150, 540, 5000, -100} {
		defender.stats[stats.ShadowResistance] = max(resistance, 0)
		attacker.stats[stats.SpellPiercing] = 0
		if resistance < 0 {
			attacker.stats[stats.SpellPiercing] = -resistance
		}
		average := defender.averageResist(spell, attacker, false)

		withCall := &Simulation{rand: NewSplitMix(7)}
		without := &Simulation{rand: NewSplitMix(7)}

		spell.expecting = true
		multiplier, outcome := spell.ResistanceMultiplier(withCall, false, attackTable)
		spell.expecting = false
		if !WithinToleranceFloat64(1-average, multiplier, 1e-12) || outcome != OutcomeEmpty {
			t.Errorf("resistance %v: expected multiplier %v with outcome %v, want %v and none", resistance, multiplier, outcome, 1-average)
		}
		if got, want := withCall.RandomFloat("Partial Resist"), without.RandomFloat("Partial Resist"); got != want {
			t.Errorf("resistance %v: an expected-damage call moved the random stream, next roll %v, want %v", resistance, got, want)
		}

		distinct := map[float64]bool{}
		for range 200 {
			real, _ := spell.ResistanceMultiplier(withCall, false, attackTable)
			distinct[real] = true
		}
		if len(distinct) < 2 && average != 0 && average < 0.9 {
			t.Errorf("resistance %v: a real hit rolled only %v", resistance, distinct)
		}
	}
}

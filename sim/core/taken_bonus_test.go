package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// The target's flat school bonus lands after the caster's done multipliers, at the spell's
// coefficient, at the coefficient a spell states for it when it sets none, and at the dot's on a tick.
func TestTakenSchoolBonusLandsAfterDoneMultipliers(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	spell := fa.Spell
	target := sim.Encounter.ActiveTargetUnits[0]
	if spell.DamageMultiplier == 1 {
		t.Fatal("the fake spell carries no done multiplier to tell the two placements apart")
	}

	hit := func() float64 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		result.inUse = false
		return result.Damage
	}
	tick := func() float64 {
		fa.Dot.Snapshot(target, 100)
		result := fa.Dot.CalcSnapshotDamage(sim, target, fa.Dot.OutcomeTick)
		result.inUse = false
		return result.Damage
	}
	// The bonus's share of a hit: the same hit with the bonus on the target and off it.
	bonus := &target.PseudoStats.SchoolBonusSpellDamage[spell.SchoolIndex]
	share := func(roll func() float64) float64 {
		*bonus = 0
		without := roll()
		*bonus = 100
		return roll() - without
	}

	if got := share(hit); got != 0 {
		t.Errorf("with no coefficient the bonus adds %v, want 0", got)
	}
	spell.BonusCoefficient = 0.5
	if got, want := share(hit), 50.0; got != want {
		t.Errorf("at coefficient 0.5 the bonus adds %v, want %v", got, want)
	}
	spell.BonusCoefficient = 0
	spell.TakenBonusCoefficient = 0.25
	if got, want := share(hit), 25.0; got != want {
		t.Errorf("at a stated taken coefficient of 0.25 the bonus adds %v, want %v", got, want)
	}
	if got, want := share(tick), fa.Dot.BonusCoefficient*100; got != want {
		t.Errorf("a tick at dot coefficient %v takes %v of the bonus, want %v", fa.Dot.BonusCoefficient, got, want)
	}
}

// A spell that ignores the target's flat school bonus, as a damage shield does, takes none of it at
// its coefficient.
func TestIgnoreTargetSchoolBonusTakesNoneOfTheBonus(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	spell := fa.Spell
	target := sim.Encounter.ActiveTargetUnits[0]
	spell.BonusCoefficient = 0.5
	spell.Flags |= SpellFlagIgnoreTargetSchoolBonus

	hit := func(bonus float64) float64 {
		target.PseudoStats.SchoolBonusSpellDamage[spell.SchoolIndex] = bonus
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		result.inUse = false
		return result.Damage
	}
	if without, with := hit(0), hit(100); with != without {
		t.Errorf("a hit deals %v under the bonus, want the %v it deals without", with, without)
	}
}

// A flat physical reduction takes at most half of a hit. A flat bonus has no such limit.
func TestFlatPhysicalReductionTakesAtMostHalfTheHit(t *testing.T) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	spell := physicalSpecial(sim.Encounter.ActiveTargetUnits[0], ProcMaskMeleeMHSpecial)
	fw.stats[stats.Armor] = 0

	tests := []struct {
		name string
		flat float64
		base float64
		want float64
	}{
		{name: "a hit of four times the reduction", flat: -30, base: 120, want: 90},
		{name: "a hit of twice the reduction", flat: -30, base: 60, want: 30},
		{name: "a hit under the reduction", flat: -30, base: 20, want: 10},
		{name: "a bonus larger than the hit", flat: 30, base: 20, want: 50},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw.PseudoStats.BonusPhysicalDamageTaken = test.flat
			result := spell.calcDamageInternal(sim, &fw.Unit, test.base, 1, false, spell.OutcomeAlwaysHitNoHitCounter)
			if !WithinToleranceFloat64(test.want, result.Damage, 0.001) {
				t.Fatalf("Damage taken: Expected: %0.3f, Actual: %0.3f", test.want, result.Damage)
			}
			spell.DisposeResult(result)
		})
	}
}

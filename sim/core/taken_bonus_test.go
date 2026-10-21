package core

import "testing"

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

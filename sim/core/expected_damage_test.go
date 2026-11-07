package core

import "testing"

func TestExpectedDamageSkipsDirectVariance(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	spell := fa.Spell
	target := sim.Encounter.ActiveTargetUnits[0]

	spell.DirectVariance = 0.2
	spell.expectedInitialDamageInternal = func(sim *Simulation, target *Unit, spell *Spell, _ bool) *SpellResult {
		return spell.CalcDamage(sim, target, 100, spell.OutcomeExpectedMagicAlwaysHit)
	}

	expected := spell.ExpectedInitialDamage(sim, target)
	for range 20 {
		if again := spell.ExpectedInitialDamage(sim, target); again != expected {
			t.Fatalf("expected damage rolled: %.4f then %.4f", expected, again)
		}
	}

	rolled := map[float64]bool{}
	for range 20 {
		result := spell.CalcDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		rolled[result.Damage] = true
		result.inUse = false
	}
	if len(rolled) < 2 {
		t.Fatalf("a cast's damage should still roll over DirectVariance, got %v", rolled)
	}
}

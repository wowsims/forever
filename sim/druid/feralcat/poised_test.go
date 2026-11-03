//go:build with_db

package feralcat

import (
	"testing"
)

// Poised: a cat at maximum Energy when the encounter starts pays 10 Energy less for its first Energy ability,
// and the next one pays in full.
func TestPoisedLowersTheOpenersEnergyCost(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	fillResource(sim, cat)
	full := cat.Shred.DefaultCast.Cost
	cat.OnEncounterStart(sim)
	if !cat.PoisedAura.IsActive() {
		t.Fatal("Poised is not up at maximum Energy")
	}
	spend := func() float64 {
		before := cat.CurrentEnergy()
		if !cat.Shred.Cast(sim, cat.CurrentTarget) {
			t.Fatal("shred did not cast")
		}
		return before - cat.CurrentEnergy()
	}
	if got := spend(); got != full-10 {
		t.Errorf("first Shred spent %g Energy, want %g", got, full-10)
	}
	if cat.PoisedAura.IsActive() {
		t.Error("Poised is still up after an Energy ability")
	}
	fillResource(sim, cat)
	sim.CurrentTime += cat.GCD.TimeToReady(sim)
	if got := spend(); got != full {
		t.Errorf("second Shred spent %g Energy, want %g", got, full)
	}
}

// Poised needs maximum Energy when the encounter starts.
func TestPoisedNeedsMaximumEnergy(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	fillResource(sim, cat)
	cat.SpendEnergy(sim, 1, cat.Shred.EnergyMetrics())
	cat.OnEncounterStart(sim)
	if cat.PoisedAura.IsActive() {
		t.Fatal("Poised is up below maximum Energy")
	}
}

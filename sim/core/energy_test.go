package core

import (
	"testing"
	"time"
)

// Combo points sit on one unit: swapping away loses nothing, and a gain on another unit drops them.
func TestComboPointsLiveOnOneUnit(t *testing.T) {
	character := newPowerBarCharacter()
	character.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5})
	sim := &Simulation{}
	a, b := &Unit{Label: "A"}, &Unit{Label: "B"}
	metrics := character.NewComboPointMetrics(ActionID{SpellID: 1})

	character.AddComboPoints(sim, a, 3, metrics)
	if got := character.ComboPoints(a); got != 3 {
		t.Fatalf("A holds %d combo points after a gain of 3, want 3", got)
	}
	if got := character.ComboPoints(b); got != 0 {
		t.Fatalf("B holds %d combo points with none gained on it, want 0", got)
	}

	onB := character.NewComboPointMetrics(ActionID{SpellID: 2})
	character.AddComboPoints(sim, b, 2, onB)
	if got := character.ComboPoints(a); got != 0 {
		t.Errorf("A holds %d combo points after a gain on B, want 0", got)
	}
	if got := character.ComboPoints(b); got != 2 {
		t.Errorf("B holds %d combo points after a gain of 2, want 2", got)
	}
	if onB.Events != 2 || onB.Gain != -1 || onB.ActualGain != -1 {
		t.Errorf("the gain on B records %d events netting %v (%v actual), want 2 netting -1: 3 dropped, 2 gained", onB.Events, onB.Gain, onB.ActualGain)
	}

	character.SpendComboPoints(sim, a, metrics)
	if got := character.ComboPoints(b); got != 2 {
		t.Errorf("B holds %d combo points after a spend on A, want 2", got)
	}
}

// A target that dies takes the combo points on it, and comes back without them.
func TestComboPointsGoWithADyingTarget(t *testing.T) {
	sim, fw := setupTwoTargetRotationSim()
	fw.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5})
	second := sim.Encounter.AllTargetUnits[1]
	fw.AddComboPoints(sim, second, 3, fw.NewComboPointMetrics(ActionID{SpellID: 1}))

	sim.DisableTargetUnit(second, true)
	if got := fw.ComboPoints(second); got != 0 {
		t.Errorf("%s holds %d combo points after it died, want 0", second.Label, got)
	}
	sim.EnableTargetUnit(second)
	if got := fw.ComboPoints(second); got != 0 {
		t.Errorf("%s holds %d combo points after it came back, want 0", second.Label, got)
	}
}

func TestEnergyRegenTicksEveryTenthOfASecond(t *testing.T) {
	character := newPowerBarCharacter()
	character.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5})
	sim := &Simulation{}

	character.energyBar.currentEnergy = 0
	for tick := 1; tick <= 10; tick++ {
		sim.CurrentTime = character.energyBar.RunTask(sim)
		if want := time.Duration(tick) * time.Second / 10; sim.CurrentTime != want {
			t.Fatalf("tick %d scheduled at %v, want %v", tick, sim.CurrentTime, want)
		}
		if got := character.CurrentEnergy(); got != float64(tick) {
			t.Fatalf("after tick %d: %v energy, want %d", tick, got, tick)
		}
	}
}

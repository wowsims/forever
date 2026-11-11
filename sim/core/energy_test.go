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
	if got := character.ComboPoints(a); got != 3 {
		t.Fatalf("A holds %d combo points after B was read, want 3", got)
	}

	character.AddComboPoints(sim, b, 2, metrics)
	if got := character.ComboPoints(a); got != 0 {
		t.Errorf("A holds %d combo points after a gain on B, want 0", got)
	}
	if got := character.ComboPoints(b); got != 2 {
		t.Errorf("B holds %d combo points after a gain of 2, want 2", got)
	}
	lost := character.energyBar.comboPointsLostMetrics
	if lost.Events != 1 || lost.Gain != -3 || lost.ActualGain != -3 {
		t.Errorf("the loss metric records %d events of %v (%v actual), want 1 of -3", lost.Events, lost.Gain, lost.ActualGain)
	}

	character.SpendComboPoints(sim, a, metrics)
	if got := character.ComboPoints(b); got != 2 {
		t.Errorf("B holds %d combo points after a spend on A, want 2", got)
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

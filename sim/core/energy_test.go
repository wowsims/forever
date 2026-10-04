package core

import (
	"testing"
	"time"
)

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

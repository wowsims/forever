package core

import (
	"fmt"
	"time"
)

func (target *Target) nextExecuteAt() time.Duration {
	switch target.executePhase {
	case 0:
		return target.executeAt35
	case 35:
		return target.executeAt25
	case 25:
		return target.executeAt20
	}
	return NeverExpires
}

// Whether the unit is at or below 20% health. Players and pets never are.
func (unit *Unit) IsExecutePhase20() bool {
	return unit.executePhase == 20
}

// Whether the unit is at or below 35% health. Players and pets never are.
func (unit *Unit) IsExecutePhase35() bool {
	return unit.executePhase != 0
}

// Whether the unit is at or below the health percent, one of the execute phases' 35, 25 and 20.
// Players and pets never are.
func (unit *Unit) IsBelowExecuteHealth(percent int32) bool {
	switch percent {
	case 35, 25, 20:
		return unit.executePhase != 0 && unit.executePhase <= percent
	}
	panic(fmt.Sprintf("no execute phase starts at %d%% health", percent))
}

// Every target reaches 35% and 20% at the encounter's execute proportions of the fight. Its health
// falls at a steady rate between the two, so it reaches 25% two thirds of the way from the 35%
// crossing to the 20% one. The first advance moves the targets into the phases already crossed.
func (encounter *Encounter) resetExecutePhases(sim *Simulation) {
	for _, target := range encounter.AllTargets {
		target.executeAt35 = time.Duration((1 - encounter.ExecuteProportion_35) * float64(sim.Duration))
		target.executeAt20 = time.Duration((1 - encounter.ExecuteProportion_20) * float64(sim.Duration))
		target.executeAt25 = target.executeAt35 + time.Duration(float64(target.executeAt20-target.executeAt35)*(35-25)/(35-20))
		target.executePhase = 0
	}
	encounter.updateNextExecuteDuration(sim)
}

// Moves the target down into each execute phase it has crossed, 35% before 25% before 20%.
func (target *Target) advanceExecutePhase(sim *Simulation) {
	if target.executePhase == 0 && sim.CurrentTime >= target.executeAt35 {
		target.executePhase = 35
	}
	if target.executePhase == 35 && sim.CurrentTime >= target.executeAt25 {
		target.executePhase = 25
	}
	if target.executePhase == 25 && sim.CurrentTime >= target.executeAt20 {
		target.executePhase = 20
	}
}

// Moves each target whose execute phase crossing is due into its phase, in target order.
func (encounter *Encounter) advanceExecutePhases(sim *Simulation) {
	for sim.CurrentTime >= sim.nextExecuteDuration {
		for _, target := range encounter.AllTargets {
			target.advanceExecutePhase(sim)
		}
		encounter.updateNextExecuteDuration(sim)
	}
}

func (encounter *Encounter) updateNextExecuteDuration(sim *Simulation) {
	sim.nextExecuteDuration = NeverExpires
	for _, target := range encounter.AllTargets {
		sim.nextExecuteDuration = min(sim.nextExecuteDuration, target.nextExecuteAt())
	}
}

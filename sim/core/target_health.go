package core

import "time"

func (target *Target) nextExecuteAt() time.Duration {
	switch target.executePhase {
	case 0:
		return target.executeAt35
	case 35:
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

// Every target reaches 35% and 20% at the encounter's execute proportions of the fight. The first advance
// moves the targets into the phases already crossed.
func (encounter *Encounter) resetExecutePhases(sim *Simulation) {
	for _, target := range encounter.AllTargets {
		target.executeAt35 = time.Duration((1 - encounter.ExecuteProportion_35) * float64(sim.Duration))
		target.executeAt20 = time.Duration((1 - encounter.ExecuteProportion_20) * float64(sim.Duration))
		target.executePhase = 0
	}
	encounter.updateNextExecuteDuration(sim)
}

// Moves the target down into each execute phase it has crossed, 35% before 20%.
func (target *Target) advanceExecutePhase(sim *Simulation) {
	if target.executePhase == 0 && sim.CurrentTime >= target.executeAt35 {
		target.executePhase = 35
	}
	if target.executePhase == 35 && sim.CurrentTime >= target.executeAt20 {
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

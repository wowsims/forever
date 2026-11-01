package core

import (
	"fmt"
	"time"
)

// A point on a target's health curve: the health percent at a fraction of the curve's span.
type HealthKnot struct {
	At      float64 // Fraction of the span, 0 at its start and 1 at its end.
	Percent float64 // Health in percent, 0 to 100.
}

const maxHealthKnots = 16

// A piecewise-linear health curve over [start, start+span], with the times it first reaches 35% and 20%.
type healthCurve struct {
	knots       [maxHealthKnots]HealthKnot
	numKnots    int
	start       time.Duration
	span        time.Duration
	executeAt35 time.Duration
	executeAt20 time.Duration
}

// Sets the target's health curve, knots sorted by fraction, and puts the target in the phase the curve is in
// now. Set while the sim resets, the curve starts the target above the execute phases, and the first advance
// moves it into the ones already crossed.
func (target *Target) SetHealthCurve(sim *Simulation, start, span time.Duration, knots ...HealthKnot) {
	target.health.set(start, span, knots)
	if sim.isResetting {
		target.executePhase = 0
	} else {
		if sim.CurrentTime < target.health.executeAt20 && target.executePhase == 20 {
			target.executePhase = 35
		}
		if sim.CurrentTime < target.health.executeAt35 && target.executePhase == 35 {
			target.executePhase = 0
		}
		target.advanceExecutePhase(sim)
	}
	sim.Encounter.updateNextExecuteDuration(sim)
}

func (curve *healthCurve) set(start, span time.Duration, knots []HealthKnot) {
	if len(knots) == 0 || len(knots) > maxHealthKnots {
		panic(fmt.Sprintf("a health curve takes 1 to %d knots, got %d", maxHealthKnots, len(knots)))
	}
	curve.numKnots = copy(curve.knots[:], knots)
	curve.start = start
	curve.span = span
	curve.executeAt35 = curve.crossing(35)
	curve.executeAt20 = curve.crossing(20)
}

// The first time the curve is at or below pct, or NeverExpires. A knot exactly at pct gives its own time, so
// the plain curve's crossings come out as (1 - proportion) * duration with no rounding of their own.
func (curve *healthCurve) crossing(pct float64) time.Duration {
	for i, knot := range curve.knots[:curve.numKnots] {
		if knot.Percent > pct {
			continue
		}
		at := knot.At
		if i > 0 && knot.Percent < pct {
			prev := curve.knots[i-1]
			at = prev.At + (prev.Percent-pct)/(prev.Percent-knot.Percent)*(knot.At-prev.At)
		}
		return curve.start + time.Duration(at*float64(curve.span))
	}
	return NeverExpires
}

// The target's health in percent at the current time: the first knot's before the curve, the last knot's
// after it, and the line between the two knots around it in between.
func (target *Target) HealthPercent(sim *Simulation) float64 {
	curve := &target.health
	knots := curve.knots[:curve.numKnots]
	if curve.span <= 0 {
		if sim.CurrentTime < curve.start {
			return knots[0].Percent
		}
		return knots[len(knots)-1].Percent
	}

	at := float64(sim.CurrentTime-curve.start) / float64(curve.span)
	if at <= knots[0].At {
		return knots[0].Percent
	}
	for i := 1; i < len(knots); i++ {
		if next := knots[i]; at < next.At {
			prev := knots[i-1]
			return prev.Percent + (at-prev.At)/(next.At-prev.At)*(next.Percent-prev.Percent)
		}
	}
	return knots[len(knots)-1].Percent
}

func (target *Target) nextExecuteAt() time.Duration {
	switch target.executePhase {
	case 0:
		return target.health.executeAt35
	case 35:
		return target.health.executeAt20
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

// Gives every target the plain curve: full health at the pull, 35% and 20% at the encounter's execute
// proportions, 0 at the end of the fight. The first advance moves the targets into the phases already crossed.
func (encounter *Encounter) resetExecutePhases(sim *Simulation) {
	for _, target := range encounter.AllTargets {
		target.health.set(0, sim.Duration, []HealthKnot{
			{At: 0, Percent: 100},
			{At: 1 - encounter.ExecuteProportion_35, Percent: 35},
			{At: 1 - encounter.ExecuteProportion_20, Percent: 20},
			{At: 1, Percent: 0},
		})
		target.executePhase = 0
	}
	encounter.updateNextExecuteDuration(sim)
}

// Moves the target down into each execute phase its curve has crossed, 35% before 20%.
func (target *Target) advanceExecutePhase(sim *Simulation) {
	if target.executePhase == 0 && sim.CurrentTime >= target.health.executeAt35 {
		target.executePhase = 35
	}
	if target.executePhase == 35 && sim.CurrentTime >= target.health.executeAt20 {
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

package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// A main hand that came ready out of range swings on the first half-second retry after the arrival is
// seen, counted from the melee attack turning on: on the way in, or at the arrival when it starts there.
func TestReadySwingLandsOnTheNextRetryAfterArrival(t *testing.T) {
	near := func(a, b time.Duration) bool { return a-b < time.Millisecond && b-a < time.Millisecond }
	retry := func(from, arrival time.Duration) time.Duration {
		steps := (arrival + core.MeleeArrivalDelay - from + core.MeleeAutoRetryInterval - 1) / core.MeleeAutoRetryInterval
		return from + steps*core.MeleeAutoRetryInterval
	}

	sim, hunter, stepTo := newStartAttackSim(t, &proto.APLRotation{})
	stepTo(3 * time.Second)
	ready, _, _ := hunter.AutoAttacks.SwingAt()
	attackOn := sim.CurrentTime
	hunter.AutoAttacks.StartMeleeAttack(sim)
	arrival := walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
	want := retry(max(ready, attackOn), arrival)
	if got, _, _ := hunter.AutoAttacks.SwingAt(); !near(got, want) || got <= arrival+core.MeleeArrivalDelay {
		t.Errorf("attack on at %v, timer out at %v, arrival at %v: main hand swings at %v, want %v", attackOn, ready, arrival, got, want)
	}

	sim, hunter, stepTo = newStartAttackSim(t, &proto.APLRotation{})
	stepTo(3 * time.Second)
	arrival = walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
	want = arrival + core.MeleeAutoRetryInterval
	if got, _, _ := hunter.AutoAttacks.SwingAt(); !near(got, want) {
		t.Errorf("attack started on arrival at %v: main hand swings at %v, want %v", arrival, got, want)
	}

	stepTo(want + 10*time.Millisecond)
	ready, _, _ = hunter.AutoAttacks.SwingAt()
	walk(sim, hunter, stepTo, 8)
	stepTo(ready - 20*time.Millisecond - core.DurationFromSeconds(3/hunter.GetMovementSpeed()))
	arrival = walk(sim, hunter, stepTo, core.MaxMeleeRange-0.01)
	want = ready + core.MeleeAutoRetryInterval
	if got, _, _ := hunter.AutoAttacks.SwingAt(); !near(got, want) {
		t.Errorf("arrival at %v, %v before the timer runs out at %v: main hand swings at %v, want %v", arrival, ready-arrival, ready, got, want)
	}
}

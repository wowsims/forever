package core

import (
	"math"
	"testing"
	"time"
)

// An event during a walk reads where the unit is at that moment, with the rotation parked so nothing
// else asks for the position.
func TestWalkPositionIsCurrentAtEveryEvent(t *testing.T) {
	sim, fc := setupSwingResetSim(t) // at 0 yd, one second in
	fc.MoveTo(12, sim)
	fc.WaitUntil(sim, NeverExpires)

	seen := -1.0
	sim.AddPendingAction(NewDelayedAction(DelayedActionOptions{
		DoAt:     time.Millisecond * 1500,
		OnAction: func(*Simulation) { seen = fc.DistanceFromTarget },
	}))
	stepSwingSim(t, sim, time.Millisecond*1500)

	if expected := 3.5; math.Abs(seen-expected) > 1e-9 {
		t.Fatalf("expected the event to read %.2f yd, got %.4f", expected, seen)
	}
}

// A walk across melee range stops or starts the melee swings on the line. The rotation is parked, as
// under a GCD, so nothing else asks for the position on the way.
func TestWalkAcrossMeleeRange(t *testing.T) {
	speed := 7.0 // a player's run speed, yards per second
	toLine := func(yards float64) time.Duration { return DurationFromSeconds(yards / speed) }

	t.Run("no swing lands after the walk out crosses the line", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t) // at 0 yd, MH due at 2.6s, OH at 2.7s
		aa := &fc.AutoAttacks
		mhPrevious, ohPrevious := aa.MainhandPreviousSwing(), aa.OffhandPreviousSwing()

		fc.MoveTo(12, sim) // crosses 5 yd at 1.714s, arrives at 2.714s
		fc.WaitUntil(sim, NeverExpires)
		stepSwingSim(t, sim, sim.CurrentTime+toLine(12))

		if aa.MainhandPreviousSwing() != mhPrevious || aa.OffhandPreviousSwing() != ohPrevious {
			t.Fatalf("expected no swing from out of range, got MH at %v and OH at %v", aa.MainhandPreviousSwing(), aa.OffhandPreviousSwing())
		}
	})

	t.Run("a swing that is due lands where the walk in crosses the line", func(t *testing.T) {
		sim, fc := setupSwingResetSim(t)
		aa := &fc.AutoAttacks
		fc.MoveTo(12, sim)
		stepSwingSim(t, sim, time.Second*3) // out of range since 1.714s, the MH swing is due

		start := sim.CurrentTime
		fc.MoveTo(0, sim) // crosses 5 yd after 1s, arrives after 1.714s
		fc.WaitUntil(sim, NeverExpires)
		stepSwingSim(t, sim, start+toLine(12))

		onLine := start + toLine(7)
		if swing := aa.MainhandPreviousSwing(); swing < onLine || swing > onLine+time.Millisecond {
			t.Fatalf("expected the MH swing on the line at %v, got %v", onLine, swing)
		}
	})
}

// For a unit with a ranged auto, a main hand that came ready out of melee range with the melee attack on
// swings on the first half-second retry after the server sees the unit arrive, the retries counted from
// the later of the timer running out and the attack starting. An attack started on arrival swings at once.
func TestReadySwingAfterTheWalkIn(t *testing.T) {
	const walkInAt = 3200 * time.Millisecond // from 12 yd, on the line 1 s later
	const onLine = walkInAt + time.Second

	// Out of melee range from 1.714 s with the melee attack off, as an Auto Shot leaves it; the MH timer
	// runs out at 2.6 s. attackOnAt < 0 leaves the attack off; swingAt > 0 moves the timer's end.
	swingAfterWalkIn := func(t *testing.T, attackOnAt, swingAt time.Duration) time.Duration {
		sim, fc := setupSwingResetSim(t)
		aa := &fc.AutoAttacks
		aa.AutoSwingRanged = true
		aa.meleeAttackOn = false
		fc.MoveTo(12, sim)
		if attackOnAt >= 0 {
			stepSwingSim(t, sim, attackOnAt)
			aa.StartMeleeAttack(sim)
		}
		stepSwingSim(t, sim, walkInAt)
		fc.MoveTo(0, sim)
		fc.WaitUntil(sim, NeverExpires)
		if swingAt > 0 {
			aa.mh.swingAt = swingAt
		}
		stepSwingSim(t, sim, onLine+time.Second)
		return aa.MainhandPreviousSwing()
	}
	for _, c := range []struct {
		name                string
		attackOnAt, swingAt time.Duration
		want                time.Duration
	}{
		{"attack started after the timer ran out: retries from the attack, 3.7 s, 4.2 s, 4.7 s", walkInAt, 0, 4700 * time.Millisecond},
		{"attack started before the timer ran out: retries from the timer, 3.1 s ... 4.6 s", 2 * time.Second, 0, 4600 * time.Millisecond},
		{"attack started on arrival: swings at once", -1, 0, onLine},
		{"timer runs out after the arrival but before the server sees it: half a second late", 2 * time.Second, onLine + 20*time.Millisecond, onLine + 520*time.Millisecond},
		{"timer runs out after the server sees the arrival: on time", 2 * time.Second, onLine + 80*time.Millisecond, onLine + 80*time.Millisecond},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := swingAfterWalkIn(t, c.attackOnAt, c.swingAt); got < c.want || got > c.want+time.Millisecond {
				t.Fatalf("expected the MH swing at %v, got %v", c.want, got)
			}
		})
	}
}

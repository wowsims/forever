package core

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func designMoveRequest(moves ...*proto.DesignMove) *proto.RaidSimRequest {
	return designTestRequest(1, &proto.EncounterDesign{Moves: moves})
}

func TestDesignMoveWalksThePlayerForItsSeconds(t *testing.T) {
	sim, fw := setupDesignSim(designMoveRequest(&proto.DesignMove{At: &proto.DesignTime{At: 0.1}, Seconds: 2}), nil)

	if got := sim.Encounter.NextMoveAt(sim); got != designAt(0.1) {
		t.Errorf("next move before the pull: expected %s, got %s", designAt(0.1), got)
	}
	if got := sim.Encounter.NextMoveDuration(sim); got != 2*time.Second {
		t.Errorf("next move duration: expected 2s, got %s", got)
	}

	runLifecycleUntil(sim, designAt(0.1)-time.Millisecond)
	if fw.Moving {
		t.Fatal("before the move: expected the player standing")
	}

	runLifecycleUntil(sim, designAt(0.1))
	if !fw.Moving {
		t.Fatal("at the move's start: expected the player walking")
	}
	if got := sim.Encounter.NextMoveAt(sim); got != NeverExpires {
		t.Errorf("next move after the only one started: expected never, got %s", got)
	}
	if got := sim.Encounter.NextMoveDuration(sim); got != 0 {
		t.Errorf("next move duration without a move to come: expected 0, got %s", got)
	}

	runLifecycleUntil(sim, designAt(0.1)+2*time.Second-time.Millisecond)
	if !fw.Moving {
		t.Error("just before the walk's end: expected the player still walking")
	}
	runLifecycleUntil(sim, designAt(0.1)+2*time.Second+time.Millisecond)
	if fw.Moving {
		t.Error("after the walk's end: expected the player standing")
	}
}

func TestDesignMoveWalksYardsAtTheRunSpeed(t *testing.T) {
	sim, fw := setupDesignSim(designMoveRequest(&proto.DesignMove{At: &proto.DesignTime{At: 0.1}, Yards: 14, Seconds: 30}), nil)
	want := DurationFromSeconds(14 / fw.GetMovementSpeed())

	if got := sim.Encounter.NextMoveDuration(sim); got != want {
		t.Errorf("duration of a 14 yard move: expected %s, got %s", want, got)
	}
	runLifecycleUntil(sim, designAt(0.1)+want-time.Millisecond)
	if !fw.Moving {
		t.Error("just before the yards are walked: expected the player walking")
	}
	runLifecycleUntil(sim, designAt(0.1)+want+time.Millisecond)
	if fw.Moving {
		t.Error("after the yards are walked: expected the player standing")
	}
}

// A permanent speed aura comes back after the design's reset, and the lookahead still reads the walk it gives.
func TestDesignNextMoveDurationReadsTheRunSpeedNow(t *testing.T) {
	config := &proto.DesignMove{At: &proto.DesignTime{At: 0.5}, Yards: 7}
	var speed *Aura
	sim, fw := setupDesignSim(designMoveRequest(config), func(char *Character) {
		speed = char.NewPassiveMovementSpeedAura("Test Speed", ActionID{SpellID: 993200}, 1.4)
	})
	if !speed.IsActive() {
		t.Fatal("test setup: expected the speed aura up at the pull")
	}
	want := DurationFromSeconds(config.Yards / fw.GetMovementSpeed())
	if got := sim.Encounter.NextMoveDuration(sim); got != want {
		t.Errorf("next move duration at the pull: expected %s, got %s", want, got)
	}

	sim.Cleanup()
	sim.Reset()
	sim.PrePull()
	if got := sim.Encounter.NextMoveDuration(sim); got != want {
		t.Errorf("next move duration in the next iteration: expected %s, got %s", want, got)
	}
}

func TestDesignMoveWaitsForACastThatCannotMove(t *testing.T) {
	sim, fw := setupDesignSim(designMoveRequest(&proto.DesignMove{At: &proto.DesignTime{At: 0.1}, Seconds: 2}), nil)

	runLifecycleUntil(sim, designAt(0.1)-time.Second)
	fw.Hardcast = Hardcast{Expires: designAt(0.1) + 500*time.Millisecond}

	runLifecycleUntil(sim, designAt(0.1)+400*time.Millisecond)
	if fw.Moving {
		t.Error("during the cast: expected the walk to wait")
	}
	runLifecycleUntil(sim, designAt(0.1)+500*time.Millisecond)
	if !fw.Moving {
		t.Error("at the cast's end: expected the walk to start")
	}
	runLifecycleUntil(sim, designAt(0.1)+2*time.Second+400*time.Millisecond)
	if !fw.Moving {
		t.Error("the delayed walk lasts its full 2s: expected the player still walking")
	}
	runLifecycleUntil(sim, designAt(0.1)+2*time.Second+600*time.Millisecond)
	if fw.Moving {
		t.Error("after the delayed walk: expected the player standing")
	}
}

func TestDesignMoveStartsWithoutWaitingForACastThatCanMove(t *testing.T) {
	sim, fw := setupDesignSim(designMoveRequest(&proto.DesignMove{At: &proto.DesignTime{At: 0.1}, Seconds: 2}), nil)

	runLifecycleUntil(sim, designAt(0.1)-time.Second)
	fw.Hardcast = Hardcast{Expires: designAt(0.1) + 500*time.Millisecond, CanMove: true}

	runLifecycleUntil(sim, designAt(0.1))
	if !fw.Moving {
		t.Error("with a cast that can move: expected the walk at the move's start")
	}
}

func TestDesignMoveRepeatsFollowTheCount(t *testing.T) {
	for _, tc := range []struct {
		name   string
		repeat *proto.DesignRepeat
		want   []float64
	}{
		{"count 2", &proto.DesignRepeat{Interval: 0.3, Count: 2}, []float64{0.1, 0.4, 0.7}},
		{"until the fight ends", &proto.DesignRepeat{Interval: 0.25}, []float64{0.1, 0.35, 0.6, 0.85}},
		{"no repeat", nil, []float64{0.1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, _ := setupDesignSim(designMoveRequest(&proto.DesignMove{At: &proto.DesignTime{At: 0.1}, Seconds: 1, Repeat: tc.repeat}), nil)
			for i, fraction := range tc.want {
				if got := sim.Encounter.NextMoveAt(sim); got != designAt(fraction) {
					t.Errorf("next move before move %d: expected %s, got %s", i+1, designAt(fraction), got)
				}
				runLifecycleUntil(sim, designAt(fraction))
			}
			if got := sim.Encounter.NextMoveAt(sim); got != NeverExpires {
				t.Errorf("next move after the last: expected never, got %s", got)
			}
		})
	}
}

func TestDesignMoveLookaheadTakesTheEarliestMove(t *testing.T) {
	sim, _ := setupDesignSim(designMoveRequest(
		&proto.DesignMove{At: &proto.DesignTime{At: 0.5}, Seconds: 3},
		&proto.DesignMove{At: &proto.DesignTime{At: 0.2}, Seconds: 1},
	), nil)

	if at, duration := sim.Encounter.NextMoveAt(sim), sim.Encounter.NextMoveDuration(sim); at != designAt(0.2) || duration != time.Second {
		t.Errorf("first lookahead: expected the 1s move at %s, got %s at %s", designAt(0.2), duration, at)
	}
	runLifecycleUntil(sim, designAt(0.2))
	if at, duration := sim.Encounter.NextMoveAt(sim), sim.Encounter.NextMoveDuration(sim); at != designAt(0.5) || duration != 3*time.Second {
		t.Errorf("second lookahead: expected the 3s move at %s, got %s at %s", designAt(0.5), duration, at)
	}
}

func TestDesignMoveRunsEveryIteration(t *testing.T) {
	sim, _ := setupDesignSim(designMoveRequest(&proto.DesignMove{
		At:      &proto.DesignTime{At: 0.1, Jitter: 0.05},
		Seconds: 1,
		Repeat:  &proto.DesignRepeat{Interval: 0.3, Count: 2},
	}), nil)
	sim.Cleanup()
	walks := 0
	gained := "Aura gained: " + ActionID{OtherID: proto.OtherAction_OtherActionMove}.String()
	sim.Log = func(message string, vals ...interface{}) {
		if strings.Contains(fmt.Sprintf(message, vals...), gained) {
			walks++
		}
	}
	for i := range 3 {
		walks = 0
		sim.runOnce()
		if walks != 3 {
			t.Errorf("walks in iteration %d: expected 3, got %d", i+1, walks)
		}
	}
}

func TestDesignMoveWithoutDistanceIsSkipped(t *testing.T) {
	sim, _ := setupDesignSim(designMoveRequest(&proto.DesignMove{At: &proto.DesignTime{At: 0.1}}), nil)
	if got := sim.Encounter.NextMoveAt(sim); got != NeverExpires {
		t.Errorf("next move of a move that walks nothing: expected never, got %s", got)
	}
}

package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func executeTestSim(duration, p20, p35 float64, numTargets int) *Simulation {
	return setupFakeRageSimFrom(executeTestRequest(duration, p20, p35, numTargets))
}

func executeTestRequest(duration, p20, p35 float64, numTargets int) *proto.RaidSimRequest {
	request := fakeRageSimRequest()
	request.Encounter.Duration = duration
	request.Encounter.ExecuteProportion_20 = p20
	request.Encounter.ExecuteProportion_35 = p35
	request.Encounter.Targets = nil
	for range numTargets {
		request.Encounter.Targets = append(request.Encounter.Targets,
			&proto.Target{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	}
	return request
}

type executePhaseEntry struct {
	target int32
	phase  int32
	at     time.Duration
}

func recordExecutePhases(sim *Simulation) *[]executePhaseEntry {
	entries := &[]executePhaseEntry{}
	sim.RegisterExecutePhaseCallback(func(sim *Simulation, target *Unit, phase int32) {
		*entries = append(*entries, executePhaseEntry{target: target.Index, phase: phase, at: sim.CurrentTime})
	})
	return entries
}

func TestPlainExecuteCrossingsFollowTheProportions(t *testing.T) {
	for _, tc := range []struct {
		duration, p20, p35 float64
	}{
		{180, 0.2, 0.35},
		{187.3, 0.137, 0.291},
		{300, 0, 0},
		{300, 1, 1},
		{60, 0, 1},
		{91.7, 0.25, 0.25},
	} {
		sim := executeTestSim(tc.duration, tc.p20, tc.p35, 1)
		target := sim.Encounter.AllTargets[0]
		want35 := time.Duration((1 - tc.p35) * float64(sim.Duration))
		want20 := time.Duration((1 - tc.p20) * float64(sim.Duration))
		if got := target.ExecuteAt(35); got != want35 {
			t.Errorf("%v s with %v/%v: 35%% at %v, want %v", tc.duration, tc.p35, tc.p20, got, want35)
		}
		if got := target.ExecuteAt(20); got != want20 {
			t.Errorf("%v s with %v/%v: 20%% at %v, want %v", tc.duration, tc.p35, tc.p20, got, want20)
		}
	}
}

func TestHealthPercentInterpolatesAndClamps(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 1)
	target := sim.Encounter.AllTargets[0]
	target.SetHealthCurve(sim, 10*time.Second, 100*time.Second,
		HealthKnot{At: 0, Percent: 100}, HealthKnot{At: 0.5, Percent: 50}, HealthKnot{At: 1, Percent: 10})

	for _, tc := range []struct {
		at   time.Duration
		want float64
	}{
		{0, 100},
		{10 * time.Second, 100},
		{35 * time.Second, 75},
		{60 * time.Second, 50},
		{85 * time.Second, 30},
		{110 * time.Second, 10},
		{200 * time.Second, 10},
	} {
		sim.CurrentTime = tc.at
		if got := target.HealthPercent(sim); got != tc.want {
			t.Errorf("health at %v = %v%%, want %v%%", tc.at, got, tc.want)
		}
	}

	if got, want := target.ExecuteAt(35), 78750*time.Millisecond; got != want {
		t.Errorf("35%% at %v, want %v", got, want)
	}
	if got, want := target.ExecuteAt(20), 97500*time.Millisecond; got != want {
		t.Errorf("20%% at %v, want %v", got, want)
	}

	target.SetHealthCurve(sim, 0, 100*time.Second, HealthKnot{At: 0, Percent: 100}, HealthKnot{At: 1, Percent: 50})
	if got := target.ExecuteAt(35); got != NeverExpires {
		t.Errorf("a curve that stays above 35%% reaches it at %v", got)
	}
}

func TestExecutePhasesPerTarget(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 2)
	first, second := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	player := &sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Unit
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation
	entries := recordExecutePhases(sim)

	first.SetHealthCurve(sim, 0, 100*time.Second,
		HealthKnot{At: 0, Percent: 100}, HealthKnot{At: 0.5, Percent: 35}, HealthKnot{At: 0.8, Percent: 20}, HealthKnot{At: 1, Percent: 0})
	second.SetHealthCurve(sim, 0, 100*time.Second,
		HealthKnot{At: 0, Percent: 100}, HealthKnot{At: 0.2, Percent: 20}, HealthKnot{At: 1, Percent: 0})

	currentTarget20 := rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{Threshold: proto.APLValueIsExecutePhase_E20}, nil)
	secondTarget20 := rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{
		Threshold:  proto.APLValueIsExecutePhase_E20,
		TargetUnit: &proto.UnitReference{Type: proto.UnitReference_Target, Index: 1},
	}, nil)

	for _, step := range []struct {
		at                    time.Duration
		first, second         int32
		currentAPL, secondAPL bool
	}{
		{10 * time.Second, 0, 0, false, false},
		{17 * time.Second, 0, 35, false, false},
		{20 * time.Second, 0, 20, false, true},
		{50 * time.Second, 35, 20, false, true},
		{80 * time.Second, 20, 20, true, true},
	} {
		sim.advance(step.at)
		if first.executePhase != step.first || second.executePhase != step.second {
			t.Errorf("at %v the phases are %d and %d, want %d and %d", step.at, first.executePhase, second.executePhase, step.first, step.second)
		}
		if got := currentTarget20.GetBool(sim); got != step.currentAPL {
			t.Errorf("at %v the current target's execute phase 20 reads %t, want %t", step.at, got, step.currentAPL)
		}
		if got := secondTarget20.GetBool(sim); got != step.secondAPL {
			t.Errorf("at %v the second target's execute phase 20 reads %t, want %t", step.at, got, step.secondAPL)
		}
	}
	if player.IsExecutePhase35() {
		t.Error("the player is in the execute phase")
	}

	want := []executePhaseEntry{
		{1, 35, 17 * time.Second},
		{1, 20, 20 * time.Second},
		{0, 35, 50 * time.Second},
		{0, 20, 80 * time.Second},
	}
	if !slices.Equal(*entries, want) {
		t.Errorf("callbacks ran %v, want %v", *entries, want)
	}
}

func TestExecutePhaseCallbacksRunOncePerTargetAndPhase(t *testing.T) {
	sim := executeTestSim(180, 1, 1, 2)
	entries := recordExecutePhases(sim)

	if sim.Encounter.AllTargets[0].IsExecutePhase35() {
		t.Error("the target is in the execute phase before the first advance")
	}
	sim.advance(0)
	sim.advance(5 * time.Second)
	sim.advance(sim.Duration)

	want := []executePhaseEntry{{0, 35, 0}, {0, 20, 0}, {1, 35, 0}, {1, 20, 0}}
	if !slices.Equal(*entries, want) {
		t.Errorf("callbacks ran %v, want %v", *entries, want)
	}
}

func TestHealthCurveSetMidFightMovesThePhaseAtOnce(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 1)
	target := sim.Encounter.AllTargets[0]
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation
	isExecute20 := rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{Threshold: proto.APLValueIsExecutePhase_E20}, nil)
	entries := recordExecutePhases(sim)
	sim.advance(60 * time.Second)

	crossed := []HealthKnot{{At: 0, Percent: 100}, {At: 0.3, Percent: 20}, {At: 1, Percent: 0}}
	target.SetHealthCurve(sim, 0, 100*time.Second, crossed...)
	if !target.IsExecutePhase20() || !isExecute20.GetBool(sim) {
		t.Error("a curve set at 60 s that crossed 20% at 30 s leaves the target out of the execute phase at 60 s")
	}
	want := []executePhaseEntry{{0, 35, 60 * time.Second}, {0, 20, 60 * time.Second}}
	if !slices.Equal(*entries, want) {
		t.Errorf("callbacks ran %v, want %v", *entries, want)
	}

	target.SetHealthCurve(sim, 0, 100*time.Second, crossed...)
	target.SetHealthCurve(sim, 0, 50*time.Second, HealthKnot{At: 0, Percent: 10})
	if !target.IsExecutePhase20() || len(*entries) != 2 {
		t.Errorf("setting a curve again on a target at 20%% ran callbacks %v, want %v", *entries, want)
	}

	target.SetHealthCurve(sim, 60*time.Second, 100*time.Second, HealthKnot{At: 0, Percent: 100}, HealthKnot{At: 1, Percent: 0})
	if target.IsExecutePhase35() || len(*entries) != 2 {
		t.Errorf("a curve back at full health keeps the target in execute (%d) or runs callbacks %v", target.executePhase, *entries)
	}
	sim.advance(125 * time.Second)
	want = append(want, executePhaseEntry{0, 35, 125 * time.Second})
	if !slices.Equal(*entries, want) {
		t.Errorf("callbacks ran %v, want %v", *entries, want)
	}
}

func TestExecutePhaseCallbackSettingACrossedCurveEntersEachPhaseOnce(t *testing.T) {
	sim := executeTestSim(180, 1, 1, 1)
	target := sim.Encounter.AllTargets[0]
	entries := recordExecutePhases(sim)
	sim.RegisterExecutePhaseCallback(func(sim *Simulation, _ *Unit, _ int32) {
		target.SetHealthCurve(sim, 0, sim.Duration, HealthKnot{At: 0, Percent: 0})
	})

	sim.advance(0)
	sim.advance(time.Second)

	want := []executePhaseEntry{{0, 35, 0}, {0, 20, 0}}
	if !slices.Equal(*entries, want) {
		t.Errorf("callbacks ran %v, want %v", *entries, want)
	}
}

func TestHealthCurveSetInAResetEffectWaitsForTheFirstAdvance(t *testing.T) {
	sim := NewSim(executeTestRequest(180, 0.2, 0.35, 1), simsignals.CreateSignals())
	target := sim.Encounter.AllTargets[0]
	var entries *[]executePhaseEntry
	sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).RegisterResetEffect(func(sim *Simulation) {
		entries = recordExecutePhases(sim)
		target.SetHealthCurve(sim, 0, sim.Duration, HealthKnot{At: 0, Percent: 10})
	})

	sim.Reset()
	if len(*entries) != 0 || target.IsExecutePhase35() {
		t.Errorf("a curve set in a reset effect ran callbacks %v or moved the phase to %d before the first advance", *entries, target.executePhase)
	}
	sim.advance(0)
	want := []executePhaseEntry{{0, 35, 0}, {0, 20, 0}}
	if !slices.Equal(*entries, want) {
		t.Errorf("callbacks ran %v, want %v", *entries, want)
	}
}

func TestExecutePhasesAllocateNothing(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 3)
	sim.RegisterExecutePhaseCallback(func(*Simulation, *Unit, int32) {})
	allocs := testing.AllocsPerRun(100, func() {
		sim.Encounter.resetExecutePhases(sim)
		sim.CurrentTime = sim.Duration
		sim.Encounter.advanceExecutePhases(sim)
	})
	if allocs != 0 {
		t.Errorf("an iteration's execute phases allocate %v times", allocs)
	}
}

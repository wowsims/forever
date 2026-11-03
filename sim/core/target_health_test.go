package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
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
		if got := target.executeAt35; got != want35 {
			t.Errorf("%v s with %v/%v: 35%% at %v, want %v", tc.duration, tc.p35, tc.p20, got, want35)
		}
		if got := target.executeAt20; got != want20 {
			t.Errorf("%v s with %v/%v: 20%% at %v, want %v", tc.duration, tc.p35, tc.p20, got, want20)
		}
	}
}

func TestExecutePhasesPerTarget(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 2)
	first, second := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	player := &sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Unit
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation

	// A plain encounter gives every target the same crossing times, so the test gives the two their own.
	first.executeAt35, first.executeAt20 = 50*time.Second, 80*time.Second
	second.executeAt35, second.executeAt20 = 15*time.Second, 20*time.Second
	sim.Encounter.updateNextExecuteDuration(sim)

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
}

func TestExecutePhasesWaitForTheFirstAdvance(t *testing.T) {
	sim := executeTestSim(180, 1, 1, 2)
	targets := sim.Encounter.AllTargets

	if targets[0].IsExecutePhase35() {
		t.Error("the target is in the execute phase before the first advance")
	}
	sim.advance(0)
	for i, target := range targets {
		if !target.IsExecutePhase20() {
			t.Errorf("target %d is in phase %d after the first advance, want 20", i, target.executePhase)
		}
	}
}

func TestExecutePhasesAllocateNothing(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 3)
	allocs := testing.AllocsPerRun(100, func() {
		sim.Encounter.resetExecutePhases(sim)
		sim.CurrentTime = sim.Duration
		sim.Encounter.advanceExecutePhases(sim)
	})
	if allocs != 0 {
		t.Errorf("an iteration's execute phases allocate %v times", allocs)
	}
}

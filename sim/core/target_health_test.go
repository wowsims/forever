package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
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
		if got := target.health.executeAt35; got != want35 {
			t.Errorf("%v s with %v/%v: 35%% at %v, want %v", tc.duration, tc.p35, tc.p20, got, want35)
		}
		if got := target.health.executeAt20; got != want20 {
			t.Errorf("%v s with %v/%v: 20%% at %v, want %v", tc.duration, tc.p35, tc.p20, got, want20)
		}
	}
}

func TestHealthPercentInterpolatesAndClamps(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 1)
	target := sim.Encounter.AllTargets[0]
	target.health.set(10*time.Second, 100*time.Second,
		[]HealthKnot{{At: 0, Percent: 100}, {At: 0.5, Percent: 50}, {At: 1, Percent: 10}})

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

	if got, want := target.health.executeAt35, 78750*time.Millisecond; got != want {
		t.Errorf("35%% at %v, want %v", got, want)
	}
	if got, want := target.health.executeAt20, 97500*time.Millisecond; got != want {
		t.Errorf("20%% at %v, want %v", got, want)
	}

	target.health.set(0, 100*time.Second, []HealthKnot{{At: 0, Percent: 100}, {At: 1, Percent: 50}})
	if got := target.health.executeAt35; got != NeverExpires {
		t.Errorf("a curve that stays above 35%% reaches it at %v", got)
	}
}

func TestExecutePhasesPerTarget(t *testing.T) {
	sim := executeTestSim(180, 0.2, 0.35, 2)
	first, second := sim.Encounter.AllTargets[0], sim.Encounter.AllTargets[1]
	player := &sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Unit
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation

	// A plain encounter gives every target the same curve, so the test gives the two their own.
	first.health.set(0, 100*time.Second,
		[]HealthKnot{{At: 0, Percent: 100}, {At: 0.5, Percent: 35}, {At: 0.8, Percent: 20}, {At: 1, Percent: 0}})
	second.health.set(0, 100*time.Second,
		[]HealthKnot{{At: 0, Percent: 100}, {At: 0.2, Percent: 20}, {At: 1, Percent: 0}})
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

// Current Health and Current Health % on an enemy read its curve. The percent needs no Health stat, and a target
// without one warns only for the absolute value.
func TestCurrentHealthOfAnEnemyReadsItsCurve(t *testing.T) {
	request := executeTestRequest(180, 0.2, 0.35, 2)
	request.Encounter.Targets[0].Stats = stats.Stats{stats.Health: 1000}.ToProtoArray()
	sim := setupFakeRageSimFrom(request)
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation
	first := &proto.UnitReference{Type: proto.UnitReference_Target, Index: 0}
	second := &proto.UnitReference{Type: proto.UnitReference_Target, Index: 1}
	health := rot.newAPLValue(&proto.APLValue{Uuid: &proto.UUID{Value: "a"}, Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: first}}})
	percent := rot.newAPLValue(&proto.APLValue{Uuid: &proto.UUID{Value: "b"}, Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: first}}})
	unstatted := &proto.APLValue{Uuid: &proto.UUID{Value: "c"}, Value: &proto.APLValue_CurrentHealth{CurrentHealth: &proto.APLValueCurrentHealth{SourceUnit: second}}}
	unstattedPercent := &proto.APLValue{Uuid: &proto.UUID{Value: "d"}, Value: &proto.APLValue_CurrentHealthPercent{CurrentHealthPercent: &proto.APLValueCurrentHealthPercent{SourceUnit: second}}}
	if health == nil || percent == nil || rot.newAPLValue(unstatted) == nil || rot.newAPLValue(unstattedPercent) == nil {
		t.Fatal("expected every health value on an enemy kept")
	}
	if warnings := rot.uuidValidations[unstatted.Uuid]; len(warnings) != 1 {
		t.Errorf("Current Health of a target without a Health stat: validations = %v, want one warning", warnings)
	}
	if warnings := rot.uuidValidations[unstattedPercent.Uuid]; len(warnings) != 0 {
		t.Errorf("Current Health %% of a target without a Health stat: validations = %v, want none", warnings)
	}

	// 90 s is 90/117 of the way from 100% at the pull to 35% at 117 s, so the target is at 50%.
	sim.advance(90 * time.Second)
	if got := percent.GetFloat(sim); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("Current Health %% at 90 s = %v, want 0.5", got)
	}
	if got := health.GetFloat(sim); math.Abs(got-500) > 1e-6 {
		t.Errorf("Current Health at 90 s = %v, want 500", got)
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

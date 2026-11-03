package core

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func targetValueRef(index int32) *proto.UnitReference {
	return &proto.UnitReference{Type: proto.UnitReference_Target, Index: index}
}

func TestValueTargetActivity(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets,
		&proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2},
		&proto.Target{Name: "down target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2, DisabledAtStart: true},
	)
	sim := setupFakeRageSimFrom(request)
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation
	sim.CurrentTime = sim.Duration / 4

	isActive := func(index int32) APLValue {
		return rot.newValueTargetIsActive(&proto.APLValueTargetIsActive{TargetUnit: targetValueRef(index)}, nil)
	}
	untilActive := func(index int32) APLValue {
		return rot.newValueTimeUntilTargetActive(&proto.APLValueTimeUntilTargetActive{TargetUnit: targetValueRef(index)}, nil)
	}
	untilDies := func(index int32) APLValue {
		return rot.newValueTimeUntilTargetDies(&proto.APLValueTimeUntilTargetDies{TargetUnit: targetValueRef(index)}, nil)
	}
	adds := rot.newValueNumActiveAdds(&proto.APLValueNumActiveAdds{}, nil)

	for index, want := range []bool{true, true, false} {
		if got := isActive(int32(index)).GetBool(sim); got != want {
			t.Errorf("Target Is Active(%d) = %t, want %t", index, got, want)
		}
		wantUntilActive, wantUntilDies := time.Duration(0), sim.Duration-sim.CurrentTime
		if !want {
			wantUntilActive, wantUntilDies = NeverExpires, NeverExpires
		}
		if got := untilActive(int32(index)).GetDuration(sim); got != wantUntilActive {
			t.Errorf("Time Until Target Active(%d) = %s, want %s", index, got, wantUntilActive)
		}
		if got := untilDies(int32(index)).GetDuration(sim); got != wantUntilDies {
			t.Errorf("Time Until Target Dies(%d) = %s, want %s", index, got, wantUntilDies)
		}
	}
	if got := adds.GetInt(sim); got != 1 {
		t.Errorf("Num Active Adds = %d, want 1 for the second target alone", got)
	}

	sim.Encounter.AllTargets[2].Enable(sim)
	if got := adds.GetInt(sim); got != 2 {
		t.Errorf("Num Active Adds after the third target is enabled = %d, want 2", got)
	}
	if !isActive(2).GetBool(sim) {
		t.Error("Target Is Active(2) after it is enabled = false, want true")
	}

	bossDown := fakeRageSimRequest()
	bossDown.Encounter.Targets[0].DisabledAtStart = true
	bossDown.Encounter.Targets = append(bossDown.Encounter.Targets,
		&proto.Target{Name: "add", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	bossDownSim := setupFakeRageSimFrom(bossDown)
	bossDownRot := bossDownSim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation
	if got := bossDownRot.newValueNumActiveAdds(&proto.APLValueNumActiveAdds{}, nil).GetInt(bossDownSim); got != 1 {
		t.Errorf("Num Active Adds with the boss down = %d, want 1 for the add", got)
	}
}

func TestValueTargetActivityNeedsAnEnemyTarget(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	self := &proto.UnitReference{Type: proto.UnitReference_Self}
	uuids := []*proto.UUID{{Value: "a"}, {Value: "b"}, {Value: "c"}}
	wantNoValue(t, map[string]APLValue{
		"Target Is Active":         rot.newValueTargetIsActive(&proto.APLValueTargetIsActive{TargetUnit: self}, uuids[0]),
		"Time Until Target Active": rot.newValueTimeUntilTargetActive(&proto.APLValueTimeUntilTargetActive{TargetUnit: self}, uuids[1]),
		"Time Until Target Dies":   rot.newValueTimeUntilTargetDies(&proto.APLValueTimeUntilTargetDies{TargetUnit: self}, uuids[2]),
	})
	for _, uuid := range uuids {
		wantWarnings(t, "the player as the target", rot.uuidValidations[uuid], "is not an enemy target")
	}

	current := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}
	currentIsActive := rot.newValueTargetIsActive(&proto.APLValueTargetIsActive{TargetUnit: current}, nil)
	currentUntilActive := rot.newValueTimeUntilTargetActive(&proto.APLValueTimeUntilTargetActive{TargetUnit: current}, nil)
	currentUntilDies := rot.newValueTimeUntilTargetDies(&proto.APLValueTimeUntilTargetDies{TargetUnit: current}, nil)
	fw.CurrentTarget = &fw.Unit
	if currentIsActive.GetBool(sim) {
		t.Error("Target Is Active for a current target that is not an enemy = true, want false")
	}
	if got := currentUntilActive.GetDuration(sim); got != NeverExpires {
		t.Errorf("Time Until Target Active for a current target that is not an enemy = %s, want never", got)
	}
	if got := currentUntilDies.GetDuration(sim); got != NeverExpires {
		t.Errorf("Time Until Target Dies for a current target that is not an enemy = %s, want never", got)
	}
}

// In a fight that ends at a health, the fight's end is the one Remaining Time estimates from the damage taken.
func TestValueTimeUntilTargetDiesInAHealthFight(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.UseHealth = true
	request.Encounter.Targets[0].Stats = stats.Stats{stats.Health: 1000}.ToProtoArray()
	sim := setupFakeRageSimFrom(request)
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation
	sim.CurrentTime = 10 * time.Second
	sim.Encounter.DamageTaken = 250

	remaining := rot.newValueRemainingTime(&proto.APLValueRemainingTime{}, nil).GetDuration(sim)
	if remaining >= sim.Duration-sim.CurrentTime {
		t.Fatalf("Remaining Time = %s, want an estimate below the %s left of the fight's placeholder duration", remaining, sim.Duration-sim.CurrentTime)
	}
	untilDies := rot.newValueTimeUntilTargetDies(&proto.APLValueTimeUntilTargetDies{TargetUnit: targetValueRef(0)}, nil)
	if got := untilDies.GetDuration(sim); got != remaining {
		t.Errorf("Time Until Target Dies(0) = %s, want Remaining Time's %s", got, remaining)
	}
}

package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func changeTargetTo(index int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_ChangeTarget{ChangeTarget: &proto.APLActionChangeTarget{
		NewTarget: &proto.UnitReference{Type: proto.UnitReference_Target, Index: index},
	}}}
}

// A reset fake warrior sim with two targets and this priority list.
func setupTwoTargetRotationSim(priorityList ...*proto.APLListItem) (*Simulation, *FakeRageWarrior) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets, &proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
	sim := setupFakeRageSimFrom(request)
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	fw.Rotation = fw.newAPLRotation(&proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: priorityList})
	sim.Cleanup()
	sim.Reset()
	return sim, fw
}

// A Change Target to the unit already targeted is not ready; one to another unit is.
func TestChangeTargetToCurrentTargetIsNotReady(t *testing.T) {
	sim, fw := setupTwoTargetRotationSim(&proto.APLListItem{Action: changeTargetTo(0)}, &proto.APLListItem{Action: changeTargetTo(1)})

	first, second := sim.Encounter.AllTargetUnits[0], sim.Encounter.AllTargetUnits[1]
	toFirst, toSecond := fw.Rotation.priorityList[0], fw.Rotation.priorityList[1]
	if fw.CurrentTarget != first {
		t.Fatalf("current target is %q, want %q", fw.CurrentTarget.Label, first.Label)
	}
	if toFirst.IsReady(sim) {
		t.Errorf("Change Target to %q is ready while %q is the current target", first.Label, first.Label)
	}
	if !toSecond.IsReady(sim) {
		t.Fatalf("Change Target to %q is not ready while %q is the current target", second.Label, first.Label)
	}

	toSecond.Execute(sim)
	if fw.CurrentTarget != second {
		t.Fatalf("after Change Target to %q, current target is %q", second.Label, fw.CurrentTarget.Label)
	}
	if !toFirst.IsReady(sim) {
		t.Errorf("Change Target to %q is not ready while %q is the current target", first.Label, second.Label)
	}
}

// Inside a sequence, a Change Target to the unit already targeted is a step that does nothing, and the sequence
// goes on to its next step.
func TestSequenceStepChangeTargetToCurrentTargetPasses(t *testing.T) {
	sim, fw := setupTwoTargetRotationSim(&proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{
		Name:    "same then second",
		Actions: []*proto.APLAction{changeTargetTo(0), changeTargetTo(1)},
	}}}})

	sequence := fw.Rotation.priorityList[0]
	second := sim.Encounter.AllTargetUnits[1]
	for step := range 2 {
		if !sequence.IsReady(sim) {
			t.Fatalf("sequence is not ready at step %d", step+1)
		}
		sequence.Execute(sim)
	}
	if fw.CurrentTarget != second {
		t.Errorf("after the sequence, current target is %q, want %q", fw.CurrentTarget.Label, second.Label)
	}
}

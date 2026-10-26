package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Two pre-pull casts back to back, the second at the instant the first one ends. Both complete.
func TestPrepullCastAtPreviousCastEnd(t *testing.T) {
	castAt := func(doAt string) *proto.APLPrepullAction {
		return &proto.APLPrepullAction{
			DoAtValue: constValue(doAt),
			Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
				SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 1}},
			}}},
		}
	}
	request := fakeSwingCasterRequest()
	request.Raid.Parties[0].Players[0].Rotation.PrepullActions = []*proto.APLPrepullAction{castAt("-5s"), castAt("-3s")}

	sim := NewSim(request, simsignals.CreateSignals())
	sim.Reset()
	sim.PrePull()
	fc := sim.Raid.Parties[0].Players[0].(*FakeSwingCaster)
	stepSwingSim(t, sim, -time.Second/2)

	if casts := fc.Resetting.SpellMetrics[0].Casts; casts != 2 {
		t.Fatalf("expected both pre-pull casts complete by -1s, got %d", casts)
	}
}

func TestPrepullActionCondition(t *testing.T) {
	// A pre-pull Change Target to the second target, run only when the encounter has more targets than minTargets.
	runPrepull := func(minTargets string) (current *Unit, second *Unit) {
		request := fakeRageSimRequest()
		request.Encounter.Targets = append(request.Encounter.Targets, &proto.Target{Name: "second target", Level: 63, MobType: proto.MobType_MobTypeHumanoid, SwingSpeed: 2})
		request.Raid.Parties[0].Players[0].Rotation.PrepullActions = []*proto.APLPrepullAction{{
			DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "-1s"}}},
			Action: &proto.APLAction{
				Condition: &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
					Op:  proto.APLValueCompare_OpGt,
					Lhs: &proto.APLValue{Value: &proto.APLValue_NumberTargets{NumberTargets: &proto.APLValueNumberTargets{}}},
					Rhs: constValue(minTargets),
				}}},
				Action: &proto.APLAction_ChangeTarget{ChangeTarget: &proto.APLActionChangeTarget{
					NewTarget: &proto.UnitReference{Type: proto.UnitReference_Target, Index: 1},
				}},
			},
		}}

		sim := setupFakeRageSimFrom(request)
		fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
		for _, prepull := range sim.prepullActions {
			prepull.Action(sim)
		}
		return fw.CurrentTarget, sim.Encounter.AllTargetUnits[1]
	}

	if current, second := runPrepull("1"); current != second {
		t.Errorf("2 targets > 1: current target is %q, want %q", current.Label, second.Label)
	}
	if current, second := runPrepull("2"); current == second {
		t.Errorf("2 targets > 2: current target is %q, want the first target", current.Label)
	}
}

// A 'Do At' that reads the encounter, here the current time, or that is not a time leaves its prepull action out
// with a warning. One that is a whole number of seconds is read as that many seconds.
func TestPrepullDoAtValues(t *testing.T) {
	currentTime := &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}
	minusTwo := &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpMul, Lhs: constValue("-1"), Rhs: constValue("2")}}}
	cases := []struct {
		doAt *proto.APLValue
		want []string
	}{
		{currentTime, []string{"cannot be read before the pull"}},
		{yesNoValue(), []string{"Cannot read a yes/no value as a duration", "'Do At' needs a time"}},
		{minusTwo, nil},
	}
	request := fakeRageSimRequest()
	for _, tc := range cases {
		request.Raid.Parties[0].Players[0].Rotation.PrepullActions = append(request.Raid.Parties[0].Players[0].Rotation.PrepullActions, &proto.APLPrepullAction{
			DoAtValue: tc.doAt,
			Action: &proto.APLAction{Action: &proto.APLAction_ChangeTarget{ChangeTarget: &proto.APLActionChangeTarget{
				NewTarget: &proto.UnitReference{Type: proto.UnitReference_Target, Index: 0},
			}}},
		})
	}
	sim := setupFakeRageSimFrom(request)
	rot := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior).Rotation

	for i, tc := range cases {
		wantWarnings(t, fmt.Sprintf("Do At %d", i), rot.prepullValidations[i], tc.want...)
	}
	if len(rot.prepullActions) != 1 || len(sim.prepullActions) != 1 || sim.prepullActions[0].DoAt != -2*time.Second {
		t.Errorf("%d prepull actions, want the one at -2s; the sim runs %v", len(rot.prepullActions), sim.prepullActions)
	}
}

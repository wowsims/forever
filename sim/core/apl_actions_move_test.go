package core

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A walk whose move re-checks while moving stops where it is once the move's condition turns false, and
// turns to the move's range when that changes. A move that does not re-check walks on to the range it set
// out for.
func TestMoveRecheckWhileMoving(t *testing.T) {
	currentTime := &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}}
	firstSecond := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op: proto.APLValueCompare_OpLt, Lhs: currentTime, Rhs: constValue("1s"),
	}}}
	// 20 yd at the pull, closing in at 30 yd a second, never under 3 yd.
	shrinking := &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: []*proto.APLValue{
		constValue("3"),
		{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{
			Op:  proto.APLValueMath_OpSub,
			Lhs: constValue("20"),
			Rhs: &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{
				Op:  proto.APLValueMath_OpMul,
				Lhs: constValue("30"),
				Rhs: &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: proto.APLValueMath_OpDiv, Lhs: currentTime, Rhs: constValue("1s")}}},
			}}},
		}}},
	}}}}

	run := func(t *testing.T, condition, moveRange *proto.APLValue, recheck bool, until time.Duration) *FakeRageWarrior {
		sim := setupFakeRageSimFrom(fakeRageSimRequest())
		fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
		fw.Rotation = fw.newAPLRotation(&proto.APLRotation{
			Type: proto.APLRotation_TypeAPL,
			PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
				Condition: condition,
				Action:    &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: moveRange, RecheckWhileMoving: recheck}},
			}}},
		})
		sim.Cleanup()
		sim.Reset()
		sim.PrePull()
		sim.Step()
		fw.AutoAttacks.CancelAutoSwing(sim)

		ct := &channelTest{sim: sim, fw: fw, rot: fw.Rotation}
		ct.runUntil(until)
		fw.UpdatePosition(sim, false)
		return fw
	}

	t.Run("stops once its condition turns false", func(t *testing.T) {
		fw := run(t, firstSecond, constValue("20"), true, 3*time.Second)
		speed, reaction := fw.GetMovementSpeed(), fw.ReactionTime.Seconds()
		if fw.Moving {
			t.Fatalf("still walking at 3s")
		}
		if d := fw.DistanceFromTarget; d < speed*1 || d > speed*(1+reaction)+1e-6 {
			t.Errorf("stopped at %.3f yd, want %.3f to %.3f yd", d, speed, speed*(1+reaction))
		}
	})

	t.Run("walks on without the re-check", func(t *testing.T) {
		fw := run(t, firstSecond, constValue("20"), false, 3*time.Second)
		if d := fw.DistanceFromTarget; math.Abs(d-20) > 1e-6 {
			t.Errorf("ended at %.3f yd, want 20 yd", d)
		}
	})

	t.Run("turns to the range when it changes", func(t *testing.T) {
		fw := run(t, nil, shrinking, true, 2*time.Second)
		if d := fw.DistanceFromTarget; math.Abs(d-3) > 1e-6 {
			t.Errorf("at %.3f yd at 2s, want 3 yd", d)
		}
		without := run(t, nil, shrinking, false, 2*time.Second)
		if d := without.DistanceFromTarget; d < 10 {
			t.Errorf("without the re-check, at %.3f yd at 2s, want it still on the way out to 20 yd", d)
		}
	})
}

// A spell is in range from its minimum range to its maximum after the range mods.
func TestSpellInRange(t *testing.T) {
	const shotID = 910105
	sim := setupFakeRageSimFrom(fakeRageSimRequest())
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	sim.Environment.State = Constructed
	shot := fw.RegisterSpell(SpellConfig{
		ActionID:     ActionID{SpellID: shotID},
		Flags:        SpellFlagAPL,
		MinRange:     8,
		MaxRange:     30,
		ApplyEffects: func(*Simulation, *Unit, *Spell) {},
	})
	sim.Environment.State = Finalized
	shot.finalize()
	fw.Rotation = fw.newAPLRotation(&proto.APLRotation{Type: proto.APLRotation_TypeAPL})
	inRange := fw.Rotation.newAPLValue(&proto.APLValue{Value: &proto.APLValue_SpellInRange{SpellInRange: &proto.APLValueSpellInRange{SpellId: ActionID{SpellID: shotID}.ToProto()}}})

	for _, tc := range []struct {
		distance, rangeMultiplier float64
		want                      bool
	}{
		{5, 1, false},
		{10, 1, true},
		{30, 1, true},
		{35, 1, false},
		{35, 1.2, true},
	} {
		fw.DistanceFromTarget = tc.distance
		shot.RangeMultiplier = tc.rangeMultiplier
		if got := inRange.GetBool(sim); got != tc.want {
			t.Errorf("at %v yd with range x%v: in range %v, want %v", tc.distance, tc.rangeMultiplier, got, tc.want)
		}
	}
}

// A Move, a Move Duration, a Wait or a Wait Until without a value it can read is left out of the rotation with
// a warning saying why.
func TestMoveAndWaitWithoutValue(t *testing.T) {
	yesNo := yesNoValue()
	text := &proto.APLValue{Value: &proto.APLValue_VariableRef{VariableRef: &proto.APLValueVariableRef{Name: "text"}}}
	cases := []struct {
		action *proto.APLAction
		want   []string
	}{
		{&proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{}}}, []string{"Move needs a range"}},
		{&proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: text}}}, []string{"Cannot read text as a number", "Move needs a range"}},
		{&proto.APLAction{Action: &proto.APLAction_MoveDuration{MoveDuration: &proto.APLActionMoveDuration{}}}, []string{"Move Duration needs a duration"}},
		{&proto.APLAction{Action: &proto.APLAction_MoveDuration{MoveDuration: &proto.APLActionMoveDuration{Duration: yesNo}}}, []string{"Cannot read a yes/no value as a duration", "Move Duration needs a duration"}},
		{&proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{}}}, []string{"Wait needs a duration"}},
		{&proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: yesNo}}}, []string{"Cannot read a yes/no value as a duration", "Wait needs a duration"}},
		{&proto.APLAction{Action: &proto.APLAction_WaitUntil{WaitUntil: &proto.APLActionWaitUntil{}}}, []string{"Wait Until needs a condition"}},
	}
	config := &proto.APLRotation{
		Type:           proto.APLRotation_TypeAPL,
		ValueVariables: []*proto.APLValueVariable{{Name: "text", Value: constValue("far")}},
	}
	for _, tc := range cases {
		config.PriorityList = append(config.PriorityList, &proto.APLListItem{Action: tc.action})
	}
	sim := setupFakeRageSimFrom(fakeRageSimRequest())
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	rot := fw.newAPLRotation(config)

	if len(rot.priorityList) != 0 {
		t.Errorf("priority list has %d actions, want none", len(rot.priorityList))
	}
	for i, tc := range cases {
		wantWarnings(t, fmt.Sprintf("action %d", i), rot.priorityListValidations[i], tc.want...)
	}
}

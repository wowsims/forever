package core

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// A Strict Multidot on two targets dots both in one chain, unless a condition on its entry, or on a
// group reference it sits under at any depth, turns false before the second dot. An outer Sequence's
// condition only decides whether the sequence starts.
func TestStrictMultidotStopsWhenItsConditionTurnsFalse(t *testing.T) {
	const dotID = 910104
	firstSecond := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpLt,
		Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
		Rhs: constValue("1s"),
	}}}
	multidot := func(condition *proto.APLValue) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_StrictMultidot{StrictMultidot: &proto.APLActionStrictMultidot{SpellId: ActionID{SpellID: dotID}.ToProto(), MaxDots: 2}},
		}}
	}
	groupRef := func(name string, condition *proto.APLValue) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: name}},
		}}
	}
	group := func(item *proto.APLListItem) []*proto.APLGroup {
		return []*proto.APLGroup{{Name: "dots", Actions: []*proto.APLListItem{item}}}
	}
	nested := []*proto.APLGroup{
		{Name: "outer", Actions: []*proto.APLListItem{groupRef("dots", firstSecond)}},
		{Name: "dots", Actions: []*proto.APLListItem{multidot(nil)}},
	}
	sequence := &proto.APLListItem{Action: &proto.APLAction{
		Condition: firstSecond,
		Action:    &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{Actions: []*proto.APLAction{multidot(nil).Action}}},
	}}

	for _, tc := range []struct {
		name   string
		first  *proto.APLListItem
		groups []*proto.APLGroup
		castOn []int32
	}{
		{"without a condition", multidot(nil), nil, []int32{0, 1}},
		{"with a condition on its entry", multidot(firstSecond), nil, []int32{0}},
		{"with a condition on its entry in a group", groupRef("dots", nil), group(multidot(firstSecond)), []int32{0}},
		{"with a condition on the group reference", groupRef("dots", firstSecond), group(multidot(nil)), []int32{0}},
		{"with a condition on a nested group reference", groupRef("outer", nil), nested, []int32{0}},
		{"with a condition on an outer sequence", sequence, nil, []int32{0, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := fakeRageSimRequest()
			request.Encounter.Targets = append(request.Encounter.Targets, request.Encounter.Targets[0])
			sim := setupFakeRageSimFrom(request)
			fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
			var castOn []int32
			sim.Environment.State = Constructed
			dot := fw.RegisterSpell(SpellConfig{
				ActionID: ActionID{SpellID: dotID},
				Flags:    SpellFlagAPL,
				Cast:     CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
				Dot:      DotConfig{Aura: Aura{Label: "Test Dot"}, NumberOfTicks: 5, TickLength: 3 * time.Second},
				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					castOn = append(castOn, target.UnitIndex)
					spell.Dot(target).Apply(sim)
				},
			})
			sim.Environment.State = Finalized
			dot.finalize()
			fw.Rotation = fw.newAPLRotation(&proto.APLRotation{
				Type:         proto.APLRotation_TypeAPL,
				Groups:       tc.groups,
				PriorityList: []*proto.APLListItem{tc.first},
			})
			sim.Cleanup()
			sim.Reset()
			sim.PrePull()
			sim.Step()
			fw.AutoAttacks.CancelAutoSwing(sim)

			ct := &channelTest{sim: sim, fw: fw, rot: fw.Rotation}
			ct.runUntil(5 * time.Second)
			if !slices.Equal(castOn, tc.castOn) {
				t.Errorf("dotted targets %v, want %v", castOn, tc.castOn)
			}
		})
	}
}

// A cast whose range is shorter than the distance the character starts at warns, unless something in
// the rotation moves the character closer.
func TestCastOutOfReachWarns(t *testing.T) {
	const nova, charge, drain = 910101, 910102, 910103
	spellID := func(id int32) *proto.ActionID { return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}} }
	cast := func(id int32) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spellID(id)}}}}
	}
	channel := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{SpellId: spellID(drain), InterruptIf: constValue("true")}}}}
	sequence := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{Actions: []*proto.APLAction{cast(nova).Action}}}}}
	move := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: constValue("5")}}}}

	groupName := "reach"
	groupRef := &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: groupName}}}
	groupedSequence := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{Actions: []*proto.APLAction{groupRef}}}}}
	group := &proto.APLGroup{Name: groupName, Actions: []*proto.APLListItem{cast(nova)}}

	for _, tc := range []struct {
		name     string
		first    *proto.APLListItem
		start    float64
		extra    []*proto.APLListItem
		groups   []*proto.APLGroup
		warnings int
	}{
		{"from 20 yd", cast(nova), 20, nil, nil, 1},
		{"from 5 yd", cast(nova), 5, nil, nil, 0},
		{"from 20 yd with a move", cast(nova), 20, []*proto.APLListItem{move}, nil, 0},
		{"from 20 yd with a cast that moves to the target", cast(nova), 20, []*proto.APLListItem{cast(charge)}, nil, 0},
		{"from 20 yd in a channel", channel, 20, nil, nil, 1},
		{"from 20 yd in a sequence", sequence, 20, nil, nil, 1},
		{"from 20 yd in a group referenced from a sequence", groupedSequence, 20, nil, []*proto.APLGroup{group}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := NewSim(fakeSwingCasterRequest(), simsignals.CreateSignals())
			fc := sim.Raid.Parties[0].Players[0].(*FakeSwingCaster)
			sim.Environment.State = Constructed
			for _, s := range []struct {
				id    int32
				flags SpellFlag
			}{{nova, SpellFlagAPL}, {charge, SpellFlagAPL | SpellFlagMovesToTarget}, {drain, SpellFlagAPL | SpellFlagChanneled}} {
				fc.RegisterSpell(SpellConfig{
					ActionID:     ActionID{SpellID: s.id},
					Flags:        s.flags,
					MaxRange:     10,
					ApplyEffects: func(*Simulation, *Unit, *Spell) {},
				})
			}
			sim.Environment.State = Finalized
			fc.StartDistanceFromTarget = tc.start

			rot := fc.newAPLRotation(&proto.APLRotation{
				Type:         proto.APLRotation_TypeAPL,
				Groups:       tc.groups,
				PriorityList: append([]*proto.APLListItem{tc.first}, tc.extra...),
			})
			stats := rot.getStats()
			validations := stats.PriorityList[0].Validations
			for _, group := range stats.Groups {
				for _, action := range group.Actions {
					validations = append(validations, action.Validations...)
				}
			}
			warnings := 0
			for _, v := range validations {
				if v.LogLevel == proto.LogLevel_Warning && strings.Contains(v.Validation, "never cast") {
					warnings++
				}
			}
			if warnings != tc.warnings {
				t.Errorf("warnings = %d, want %d: %v", warnings, tc.warnings, validations)
			}
		})
	}
}

// A manual cast is a button press by the rotation. A spell that another spell casts keeps its own casts
// but adds no press.
func TestManualCastsCountOnlyWhatTheRotationPresses(t *testing.T) {
	const pressedID, triggeredID = 910105, 910106
	var pressed, triggered *Spell
	ct := newChannelTest(func(ct *channelTest, char *Character) *Spell {
		triggered = char.RegisterSpell(SpellConfig{
			ActionID:     ActionID{SpellID: triggeredID},
			Flags:        SpellFlagPassiveSpell | SpellFlagProc,
			ApplyEffects: func(*Simulation, *Unit, *Spell) {},
		})
		pressed = char.RegisterSpell(SpellConfig{
			ActionID: ActionID{SpellID: pressedID},
			Flags:    SpellFlagAPL,
			Cast:     CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
			ApplyEffects: func(sim *Simulation, target *Unit, _ *Spell) {
				triggered.Cast(sim, target)
			},
		})
		return fakeShoot(char, &ct.shots)
	})
	pressed.finalize()
	triggered.finalize()

	ct.start(ct.castOf(pressed))
	ct.runUntil(10 * time.Second)

	unit := ct.fw.CurrentTarget.UnitIndex
	got, trig := pressed.SpellMetrics[unit], triggered.SpellMetrics[unit]
	if got.Casts == 0 || got.ManualCasts != got.Casts {
		t.Errorf("pressed spell: ManualCasts = %d, Casts = %d, want both equal and above 0", got.ManualCasts, got.Casts)
	}
	if trig.Casts != got.Casts || trig.ManualCasts != 0 {
		t.Errorf("triggered spell: ManualCasts = %d, Casts = %d, want 0 and %d", trig.ManualCasts, trig.Casts, got.Casts)
	}
}

// A walk the rotation sets off from standing is one press of the move. A walk a spell starts, such as Charge,
// and a turn of a walk under way add none, though the first walk and the spell's own are both casts.
func TestManualCastsCountTheMovesTheRotationStarts(t *testing.T) {
	ct := newShootTest()
	move := ct.rot.newActionMove(&proto.APLActionMove{RangeFromTarget: constValue("20")})
	moveFor := ct.rot.newActionMoveDuration(&proto.APLActionMoveDuration{Duration: constValue("2s")})
	// Walks by hand and walks a spell causes record in separate splits of the move; the counts here are their sum.
	metrics := func() SpellMetrics {
		var sum SpellMetrics
		for _, split := range ct.fw.moveSpell.splitSpellMetrics {
			sum.Casts += split[ct.fw.CurrentTarget.UnitIndex].Casts
			sum.ManualCasts += split[ct.fw.CurrentTarget.UnitIndex].ManualCasts
		}
		return sum
	}

	move.Execute(ct.sim)
	if got := metrics(); got.ManualCasts != 1 || got.Casts != 1 {
		t.Errorf("a Move from standing: ManualCasts = %d, Casts = %d, want 1 and 1", got.ManualCasts, got.Casts)
	}

	moveFor.Execute(ct.sim)
	if got := metrics(); got.ManualCasts != 1 || got.Casts != 1 {
		t.Errorf("a MoveDuration that turns the walk under way: ManualCasts = %d, Casts = %d, want 1 and 1", got.ManualCasts, got.Casts)
	}

	ct.fw.StopMoving(ct.sim)
	ct.fw.MoveTo(5, ct.sim)
	if got := metrics(); got.ManualCasts != 1 || got.Casts != 2 {
		t.Errorf("a walk a spell starts: ManualCasts = %d, Casts = %d, want 1 and 2", got.ManualCasts, got.Casts)
	}

	ct.fw.StopMoving(ct.sim)
	moveFor.Execute(ct.sim)
	if got := metrics(); got.ManualCasts != 2 || got.Casts != 3 {
		t.Errorf("a MoveDuration from standing: ManualCasts = %d, Casts = %d, want 2 and 3", got.ManualCasts, got.Casts)
	}
}

package core

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func TestStrictMultidotStopsWhenItsConditionTurnsFalse(t *testing.T) {
	request := fakeRageSimRequest()
	request.Encounter.Targets = append(request.Encounter.Targets, request.Encounter.Targets[0])
	sim := setupFakeRageSimFrom(request)
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	var castOn []int32
	sim.Environment.State = Constructed
	dot := fw.RegisterSpell(SpellConfig{
		ActionID: ActionID{SpellID: 910104},
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
	sim.Cleanup()
	sim.Reset()

	ct := &channelTest{sim: sim, fw: fw, rot: fw.Rotation}
	action := ct.rot.newAPLAction(&proto.APLAction{
		Condition: &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
			Op:  proto.APLValueCompare_OpLt,
			Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
			Rhs: constValue("1s"),
		}}},
		Action: &proto.APLAction_StrictMultidot{StrictMultidot: &proto.APLActionStrictMultidot{SpellId: dot.ActionID.ToProto(), MaxDots: 2}},
	})
	action.Finalize(ct.rot)
	ct.rot.priorityList = []*APLAction{action}
	sim.PrePull()
	sim.Step()
	fw.AutoAttacks.CancelAutoSwing(sim)

	ct.runUntil(5 * time.Second)
	if !slices.Equal(castOn, []int32{0}) {
		t.Errorf("a condition true only for the first second should dot only the first target at 0s, dotted targets %v", castOn)
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

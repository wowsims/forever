package core

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

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

	for _, tc := range []struct {
		name  string
		first *proto.APLListItem
		start float64
		extra []*proto.APLListItem
		warns bool
	}{
		{"from 20 yd", nil, 20, nil, true},
		{"from 5 yd", nil, 5, nil, false},
		{"from 20 yd with a move", nil, 20, []*proto.APLListItem{move}, false},
		{"from 20 yd with a cast that moves to the target", nil, 20, []*proto.APLListItem{cast(charge)}, false},
		{"from 20 yd in a channel", channel, 20, nil, true},
		{"from 20 yd in a sequence", sequence, 20, nil, true},
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

			first := tc.first
			if first == nil {
				first = cast(nova)
			}
			rot := fc.newAPLRotation(&proto.APLRotation{
				Type:         proto.APLRotation_TypeAPL,
				PriorityList: append([]*proto.APLListItem{first}, tc.extra...),
			})
			validations := rot.getStats().PriorityList[0].Validations
			warnings := 0
			for _, v := range validations {
				if v.LogLevel == proto.LogLevel_Warning && strings.Contains(v.Validation, "never cast") {
					warnings++
				}
			}
			if want := map[bool]int{true: 1, false: 0}[tc.warns]; warnings != want {
				t.Errorf("warnings = %d, want %d: %v", warnings, want, validations)
			}
		})
	}
}

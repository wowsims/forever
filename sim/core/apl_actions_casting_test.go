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
	const nova, charge = 910101, 910102
	spellID := func(id int32) *proto.ActionID { return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}} }
	cast := func(id int32) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spellID(id)}}}}
	}
	move := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: constValue("5")}}}}

	for _, tc := range []struct {
		name  string
		start float64
		extra []*proto.APLListItem
		warns bool
	}{
		{"from 20 yd", 20, nil, true},
		{"from 5 yd", 5, nil, false},
		{"from 20 yd with a move", 20, []*proto.APLListItem{move}, false},
		{"from 20 yd with a cast that moves to the target", 20, []*proto.APLListItem{cast(charge)}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := NewSim(fakeSwingCasterRequest(), simsignals.CreateSignals())
			fc := sim.Raid.Parties[0].Players[0].(*FakeSwingCaster)
			sim.Environment.State = Constructed
			for _, s := range []struct {
				id    int32
				flags SpellFlag
			}{{nova, SpellFlagAPL}, {charge, SpellFlagAPL | SpellFlagMovesToTarget}} {
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
				PriorityList: append([]*proto.APLListItem{cast(nova)}, tc.extra...),
			})
			validations := rot.getStats().PriorityList[0].Validations
			warned := false
			for _, v := range validations {
				warned = warned || (v.LogLevel == proto.LogLevel_Warning && strings.Contains(v.Validation, "never cast"))
			}
			if warned != tc.warns {
				t.Errorf("warned = %v, want %v: %v", warned, tc.warns, validations)
			}
		})
	}
}

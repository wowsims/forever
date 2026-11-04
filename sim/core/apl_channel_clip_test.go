package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A channel whose interrupt condition turns true keeps running when the rotation's next action would
// start the same channel again, unless its channel action allows the recast. The guard finds that action
// behind a condition on the channel's own aura, which reads as down with the channel ended, inside a
// group, as the next step of the strict sequence the channel runs in, and as a multidot of the spell.
func TestChannelClipHoldsWhenTheNextActionRecastsIt(t *testing.T) {
	const channelID = 910109
	spellID := ActionID{SpellID: channelID}.ToProto()
	afterTwoTicks := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpGe,
		Lhs: &proto.APLValue{Value: &proto.APLValue_SpellChanneledTicks{SpellChanneledTicks: &proto.APLValueSpellChanneledTicks{SpellId: spellID}}},
		Rhs: constValue("2"),
	}}}
	auraDown := &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     spellID,
	}}}}}}
	firstSecond := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpLt,
		Lhs: &proto.APLValue{Value: &proto.APLValue_CurrentTime{CurrentTime: &proto.APLValueCurrentTime{}}},
		Rhs: constValue("1s"),
	}}}
	channel := func(condition *proto.APLValue, allowRecast bool) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{SpellId: spellID, InterruptIf: afterTwoTicks, AllowRecast: allowRecast}},
		}}
	}
	multidot := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Multidot{Multidot: &proto.APLActionMultidot{SpellId: spellID, MaxDots: 1}}}}
	strictSequence := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_StrictSequence{StrictSequence: &proto.APLActionStrictSequence{
		Actions: []*proto.APLAction{channel(nil, false).Action, channel(nil, false).Action},
	}}}}
	groupRef := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: "channel"}}}}
	group := func(item *proto.APLListItem) []*proto.APLGroup {
		return []*proto.APLGroup{{Name: "channel", Actions: []*proto.APLListItem{item}}}
	}

	for _, tc := range []struct {
		name   string
		rows   []*proto.APLListItem
		groups []*proto.APLGroup
		castAt []time.Duration
	}{
		{"in the priority list", []*proto.APLListItem{channel(nil, false)}, nil, seconds(0, 5)},
		{"behind a condition on its aura", []*proto.APLListItem{channel(auraDown, false)}, nil, seconds(0, 5)},
		{"in a group", []*proto.APLListItem{groupRef}, group(channel(nil, false)), seconds(0, 5)},
		{"in a group behind a condition on its aura", []*proto.APLListItem{groupRef}, group(channel(auraDown, false)), seconds(0, 5)},
		{"as the next step of its strict sequence", []*proto.APLListItem{strictSequence}, nil, seconds(0, 5)},
		{"as a multidot of the spell", []*proto.APLListItem{channel(firstSecond, false), multidot}, nil, seconds(0, 5)},
		{"with the recast allowed", []*proto.APLListItem{channel(auraDown, true)}, nil, seconds(0, 2, 4)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var castAt []time.Duration
			var spell *Spell
			sim, fw, _ := setupFakeRageRotationWith(func(char *Character) {
				spell = char.RegisterSpell(SpellConfig{
					ActionID: ActionID{SpellID: channelID},
					Flags:    SpellFlagAPL | SpellFlagChanneled,
					Cast:     CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
					Dot:      DotConfig{Aura: Aura{Label: "Test Channel"}, NumberOfTicks: 5, TickLength: time.Second},
					ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
						castAt = append(castAt, sim.CurrentTime)
						spell.Dot(target).Apply(sim)
					},
				})
			})
			spell.finalize()
			fw.Rotation = fw.newAPLRotation(&proto.APLRotation{
				Type:         proto.APLRotation_TypeAPL,
				Groups:       tc.groups,
				PriorityList: tc.rows,
			})
			sim.Cleanup()
			sim.Reset()
			sim.PrePull()
			sim.Step()
			fw.AutoAttacks.CancelAutoSwing(sim)

			ct := &channelTest{sim: sim, fw: fw, rot: fw.Rotation}
			ct.runUntil(5500 * time.Millisecond)
			if !slices.Equal(castAt, tc.castAt) {
				t.Errorf("channel cast at %v, want %v", castAt, tc.castAt)
			}
		})
	}
}

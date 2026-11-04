package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// Not(Aura Is Active) of the aura on the current target.
func notAuraActive(id *proto.ActionID) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
		AuraId:     id,
	}}}}}}
}

// A channel whose interrupt condition turns true keeps running when the rotation's next action would
// start the same channel again, unless its channel action allows the recast. The guard finds that action
// behind a condition on the channel's own aura, which reads as down with the channel ended, inside a
// group, as the current step of a sequence or of the strict sequence the channel runs in, and as a
// multidot of the spell.
func TestChannelClipHoldsWhenTheNextActionRecastsIt(t *testing.T) {
	const channelID = 910109
	spellID := ActionID{SpellID: channelID}.ToProto()
	afterTwoTicks := &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{
		Op:  proto.APLValueCompare_OpGe,
		Lhs: &proto.APLValue{Value: &proto.APLValue_SpellChanneledTicks{SpellChanneledTicks: &proto.APLValueSpellChanneledTicks{SpellId: spellID}}},
		Rhs: constValue("2"),
	}}}
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
	sequence := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{
		Name:    "channels",
		Actions: []*proto.APLAction{channel(nil, false).Action, channel(nil, false).Action},
	}}}}
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
		{"in the priority list", []*proto.APLListItem{channel(nil, false)}, nil, seconds(0, 5, 10)},
		{"behind a condition on its aura", []*proto.APLListItem{channel(notAuraActive(spellID), false)}, nil, seconds(0, 5, 10)},
		{"in a group", []*proto.APLListItem{groupRef}, group(channel(nil, false)), seconds(0, 5, 10)},
		{"in a group behind a condition on its aura", []*proto.APLListItem{groupRef}, group(channel(notAuraActive(spellID), false)), seconds(0, 5, 10)},
		{"as the current step of a sequence", []*proto.APLListItem{sequence}, nil, seconds(0, 5)},
		{"as the next step of its strict sequence, and its first once it restarts", []*proto.APLListItem{strictSequence}, nil, seconds(0, 5, 10)},
		{"as a multidot of the spell", []*proto.APLListItem{channel(firstSecond, false), multidot}, nil, seconds(0, 5, 10)},
		{"with the recast allowed", []*proto.APLListItem{channel(notAuraActive(spellID), true)}, nil, seconds(0, 2, 4, 6, 8, 10)},
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
			ct.runUntil(10500 * time.Millisecond)
			if !slices.Equal(castAt, tc.castAt) {
				t.Errorf("channel cast at %v, want %v", castAt, tc.castAt)
			}
		})
	}
}

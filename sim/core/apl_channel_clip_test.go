package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A channel whose interrupt condition turns true keeps running when the rotation's next action would
// start the same channel again, unless its channel action allows the recast. The guard finds that action
// behind a condition on the channel's own aura, which reads as down with the channel ended, and inside a
// group.
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
	channel := func(condition *proto.APLValue, allowRecast bool) *proto.APLListItem {
		return &proto.APLListItem{Action: &proto.APLAction{
			Condition: condition,
			Action:    &proto.APLAction_ChannelSpell{ChannelSpell: &proto.APLActionChannelSpell{SpellId: spellID, InterruptIf: afterTwoTicks, AllowRecast: allowRecast}},
		}}
	}
	groupRef := &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: &proto.APLActionGroupReference{GroupName: "channel"}}}}
	group := func(item *proto.APLListItem) []*proto.APLGroup {
		return []*proto.APLGroup{{Name: "channel", Actions: []*proto.APLListItem{item}}}
	}

	for _, tc := range []struct {
		name   string
		first  *proto.APLListItem
		groups []*proto.APLGroup
		castAt []time.Duration
	}{
		{"in the priority list", channel(nil, false), nil, seconds(0, 5)},
		{"behind a condition on its aura", channel(auraDown, false), nil, seconds(0, 5)},
		{"in a group", groupRef, group(channel(nil, false)), seconds(0, 5)},
		{"in a group behind a condition on its aura", groupRef, group(channel(auraDown, false)), seconds(0, 5)},
		{"with the recast allowed", channel(auraDown, true), nil, seconds(0, 2, 4)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim := setupFakeRageSimFrom(fakeRageSimRequest())
			fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
			var castAt []time.Duration
			sim.Environment.State = Constructed
			spell := fw.RegisterSpell(SpellConfig{
				ActionID: ActionID{SpellID: channelID},
				Flags:    SpellFlagAPL | SpellFlagChanneled,
				Cast:     CastConfig{DefaultCast: Cast{GCD: GCDDefault}},
				Dot:      DotConfig{Aura: Aura{Label: "Test Channel"}, NumberOfTicks: 5, TickLength: time.Second},
				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					castAt = append(castAt, sim.CurrentTime)
					spell.Dot(target).Apply(sim)
				},
			})
			sim.Environment.State = Finalized
			spell.finalize()
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
			ct.runUntil(5500 * time.Millisecond)
			if !slices.Equal(castAt, tc.castAt) {
				t.Errorf("channel cast at %v, want %v", castAt, tc.castAt)
			}
		})
	}
}

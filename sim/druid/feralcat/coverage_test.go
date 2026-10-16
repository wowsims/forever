//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Feral charge, shifting power, primal bite and berserk, plus furor and natural shapeshifter so the
// form spells can both be paid for. Prowl is castable only before the pull, so each opener is its own
// request: one prowls into Pounce, the other into Ravage. Both then leave cat form and come back so
// both form spells are casts.
const coverageTalents = "0000000000000000-00000000101100000001-0500300000000000"

func TestFeralCatCoverage(t *testing.T) {
	core.RunTestSuite(t, t.Name(), []core.TestGenerator{castsGen{
		{"Casts", catCoverageRequest(9827)},
		{"RavageCasts", catCoverageRequest(9867)},
	}})
}

type namedRequest struct {
	name string
	rsr  *proto.RaidSimRequest
}

type castsGen []namedRequest

func (g castsGen) NumTests() int { return len(g) }

func (g castsGen) GetTest(i int) (string, *proto.ComputeStatsRequest, *proto.StatWeightsRequest, *proto.RaidSimRequest) {
	return g[i].name, nil, nil, g[i].rsr
}

func catCoverageRequest(opener int32) *proto.RaidSimRequest {
	// The opener while prowling, and nothing else while prowling so an auto-replace does not spend the
	// stealth. The form spells sit under the builders.
	prowling := auraUp(9913)
	return coverageRequest(false, &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PrepullActions: []*proto.APLPrepullAction{{
			Action:    spellAction(9913),
			DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "-1s"}}},
		}},
		PriorityList: []*proto.APLListItem{
			castWhen(opener, prowling),
			castWhen(417141, not(prowling)),
			castWhen(1322605, not(prowling)),
			// Energy ticks jump over the gaps between these costs, so each builder and finisher
			// owns a slice of the fight. Claw also builds combo points for the finishers.
			castWhen(9830, and(not(prowling), timeBetween("0.2", "0.4"))),
			castWhen(9904, and(not(prowling), timeBetween("0.4", "0.6"))),
			castWhen(22829, and(not(prowling), timeBetween("0.6", "0.8"))),
			castWhen(9896, and(not(prowling), timeAtLeast("0.8"))),
			castWhen(9850, and(not(prowling), or(timeBelow("0.2"), timeAtLeast("0.6")))),
			// A late shift into dire bear spends the mana needed to shift back, so the form
			// spells only run while claw still owns the fight.
			castWhen(9634, and(not(prowling), timeBelow("0.2"))),
			castWhen(768, and(not(prowling), timeBelow("0.2"))),
			castWhen(9892, not(prowling)),
			castWhen(9821, not(prowling)),
			castWhen(9907, and(not(prowling), not(auraOnTarget(9907)))),
		},
	})
}

func coverageRequest(bear bool, rotation *proto.APLRotation) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	player := &proto.Player{
		Name: "Feral", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		TalentsString: coverageTalents,
		Equipment:     &proto.EquipmentSpec{Items: items},
		Consumables:   &proto.ConsumesSpec{},
		Buffs:         &proto.IndividualBuffs{},
		Rotation:      rotation,
	}
	if bear {
		player.Spec = &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
			Options: &proto.FeralBearDruid_Options{StartingRage: 100},
		}}
	} else {
		player.Spec = &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
			Options: &proto.FeralCatDruid_Options{},
		}}
	}
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}
}

func cast(id int32) *proto.APLListItem {
	return castWhen(id, nil)
}

func castWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	action := spellAction(id)
	action.Condition = condition
	return &proto.APLListItem{Action: action}
}

func auraUp(id int32) *proto.APLValue {
	return auraOn(id, nil)
}

func auraOnTarget(id int32) *proto.APLValue {
	return auraOn(id, &proto.UnitReference{Type: proto.UnitReference_CurrentTarget})
}

func auraOn(id int32, source *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: source,
		AuraId:     &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func not(value *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: value}}}
}

func and(values ...*proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: values}}}
}

func or(values ...*proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: values}}}
}

func timeBelow(fraction string) *proto.APLValue {
	return cmp(proto.APLValueCompare_OpLt, timePercent(), konst(fraction))
}

func timeAtLeast(fraction string) *proto.APLValue {
	return cmp(proto.APLValueCompare_OpGe, timePercent(), konst(fraction))
}

func timeBetween(low, high string) *proto.APLValue {
	return and(timeAtLeast(low), timeBelow(high))
}

func timePercent() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentTimePercent{CurrentTimePercent: &proto.APLValueCurrentTimePercent{}}}
}

func konst(val string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

func cmp(op proto.APLValueCompare_ComparisonOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: lhs, Rhs: rhs}}}
}

func sequence(ids ...int32) *proto.APLListItem {
	actions := make([]*proto.APLAction, len(ids))
	for i, id := range ids {
		actions[i] = spellAction(id)
	}
	return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_StrictSequence{
		StrictSequence: &proto.APLActionStrictSequence{Actions: actions},
	}}}
}

func spellAction(id int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

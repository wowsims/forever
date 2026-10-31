//go:build with_db

package feralbear

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The cat coverage talents with Lacerate in place of Shifting Power. The opener leaves dire bear and
// comes back so both form spells are casts, then the bear abilities run.
const coverageTalents = "0000000000000000-05000000100100000010-0000300000000000"

func TestFeralBearCoverage(t *testing.T) {
	core.RunTestSuite(t, t.Name(), []core.TestGenerator{castsGen{bearCoverageRequest()}})
}

type castsGen struct {
	rsr *proto.RaidSimRequest
}

func (g castsGen) NumTests() int { return 1 }

func (g castsGen) GetTest(int) (string, *proto.ComputeStatsRequest, *proto.StatWeightsRequest, *proto.RaidSimRequest) {
	return "Casts", nil, nil, g.rsr
}

func bearCoverageRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 1613
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: coverageTalents,
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 100},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation: &proto.APLRotation{
					Type: proto.APLRotation_TypeAPL,
					PriorityList: []*proto.APLListItem{
						sequence(9908, 1235827, 8983, 5209, 9898, 1238070),
						cast(9881),
						cast(5229),
						cast(22842),
						cast(6795),
						cast(16979),
						cast(22812),
						castUnlessOnTarget(9907),
						// Both shifts cost mana: shift out only with the mana to shift back.
						withCondition(sequence(768, 9634), cmp(proto.APLValueCompare_OpGe,
							&proto.APLValue{Value: &proto.APLValue_CurrentMana{CurrentMana: &proto.APLValueCurrentMana{}}},
							mathOp(proto.APLValueMath_OpAdd, spellCost(768), spellCost(9634)))),
					},
				},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}
}

func withCondition(item *proto.APLListItem, condition *proto.APLValue) *proto.APLListItem {
	item.Action.Condition = condition
	return item
}

func cast(id int32) *proto.APLListItem {
	return &proto.APLListItem{Action: spellAction(id)}
}

// A debuff the druid can refresh at any time, cast only while the target does not have it.
func castUnlessOnTarget(id int32) *proto.APLListItem {
	action := spellAction(id)
	action.Condition = &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: &proto.APLValue{
		Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
			SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
			AuraId:     &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
		}},
	}}}}
	return &proto.APLListItem{Action: action}
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

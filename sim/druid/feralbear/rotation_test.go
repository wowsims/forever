//go:build with_db

package feralbear

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

const bearTalents = DefaultTalents

func TestBearShippedAplRuns(t *testing.T) {
	req := bearNakedRequest()
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	for _, id := range []int32{1238073, 9881, 22812} {
		if bearCastsOf(player.Actions, id) == 0 {
			t.Fatalf("shipped bear APL never casts %d", id)
		}
	}
}

func bearNakedRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 1613
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: bearTalents,
				Equipment:     &proto.EquipmentSpec{Items: items},
				Consumables:   &proto.ConsumesSpec{},
				Buffs:         &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{},
				}},
				Rotation: core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation,
			}},
			Buffs: &proto.PartyBuffs{},
		}},
			Tanks: []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}
}

func bearCastsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

package healer

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterHealerPriest()
	common.RegisterAllEffects()
}

// Stats-only suite: this spec is a gear planner, it has no healing rotation.
// Pins the final character stats for each gear preset so the passives stay covered. The empty APL
// rotation and the fake prepull (no SkipRotation) make it exercise a full environment reset, the
// path the UI's stats request takes.
func TestHealerPriest(t *testing.T) {
	t.Skip("healer gear set JSON files referenced by this test are absent")
	var generators []core.TestGenerator
	for _, gearSet := range []string{"preraid", "p3_t5", "p3"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPriest,
				Race:          proto.Race_RaceDwarf,
				Equipment:     core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: StandardTalents,
				Profession1:   proto.Profession_Tailoring,
				Profession2:   proto.Profession_Enchanting,
				Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			},
			PlayerOptions,
		)
		generators = append(generators, &core.SingleCharacterStatsTestGenerator{
			Name: gearSet,
			Request: &proto.ComputeStatsRequest{
				Raid: core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
			},
		})
	}
	core.RunTestSuite(t, t.Name(), generators)
}

// Circle of Healing 20/41/0, wowhead's TBC raid build.
var StandardTalents = "50023011305-235050032002150520051"

var FullConsumes = &proto.ConsumesSpec{
	FoodId:        250066, // Soaring Pamplona
	ResourcePotId: 13444,  // Major Mana Potion
}

var PlayerOptions = &proto.Player_HealerPriest{
	HealerPriest: &proto.HealerPriest{
		Options: &proto.HealerPriest_Options{
			ClassOptions: &proto.PriestOptions{},
		},
	},
}

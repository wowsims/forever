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

// Pins the final character stats for each gear preset so the passives stay covered, and the healing
// of the Disc and Holy preset rotations on a raid-healing encounter. The stats rows' empty APL
// rotation and fake prepull (no SkipRotation) make them exercise a full environment reset, the path
// the UI's stats request takes.
func TestHealerPriest(t *testing.T) {
	var generators []core.TestGenerator
	for _, gearSet := range []string{"preraid"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPriest,
				Race:          proto.Race_RaceDwarf,
				Equipment:     core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: HolyTalents,
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
	for _, preset := range []struct{ name, build, talents string }{
		{"Disc", "disc_testing", DiscTalents},
		{"Holy", "holy_testing", HolyTalents},
	} {
		generators = append(generators, &core.SingleDpsTestGenerator{
			Name:    preset.name + "-RaidHealing10",
			Request: raidHealingRequest(preset.build, preset.talents),
		})
	}
	core.RunTestSuite(t, t.Name(), generators)
}

// A preset build's rotation, options, consumables and buffs on its encounter, in the pre-raid gear,
// with the Raid healing assignment of a 10-player group: nine target dummies, four in the priest's
// party and five in a second one.
func raidHealingRequest(buildFile string, talents string) *proto.RaidSimRequest {
	build := core.GetTestBuildFromJSON(proto.Class_ClassPriest, "../../../ui/specs/priest/healer/builds", buildFile,
		core.ItemFilter{}, nil, nil)
	player := core.WithSpec(
		&proto.Player{
			Class:              proto.Class_ClassPriest,
			Race:               build.Race,
			Equipment:          core.GetGearSet("../../../ui/specs/priest/healer/gear_sets", "preraid").GearSet,
			Consumables:        build.Consumables,
			Buffs:              build.IndividualBuffs,
			TalentsString:      talents,
			Profession1:        build.Profession1,
			Profession2:        build.Profession2,
			Rotation:           build.Rotation.Rotation,
			HealingModel:       build.HealingModel,
			DistanceFromTarget: build.StartingDistance,
			ReactionTimeMs:     build.ReactionTimeMs,
		},
		build.SpecOptions.SpecOptions,
	)
	raid := core.SinglePlayerRaidProto(player, build.PartyBuffs, build.RaidBuffs, build.Debuffs)
	raid.Parties = append(raid.Parties, &proto.Party{})
	raid.TargetDummies = 9
	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  build.Encounter.Encounter,
		SimOptions: core.DefaultSimTestOptions,
	}
}

// The Holy and Disc presets from ui/specs/priest/healer/presets.ts.
var HolyTalents = "00500502103-33505003030121531"
var DiscTalents = "005005031330011531-335020030301"

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

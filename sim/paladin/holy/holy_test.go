package holy

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterHolyPaladin()
	common.RegisterAllEffects()
}

// Stats-only suite: this spec is a gear planner, it has no healing rotation.
// Pins the final character stats for each gear preset so the passives stay covered. The empty APL
// rotation and the fake prepull (no SkipRotation) make it exercise a full environment reset, the
// path the UI's stats request takes.
func TestHolyPaladin(t *testing.T) {
	var generators []core.TestGenerator
	for _, gearSet := range []string{"default"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPaladin,
				Race:          proto.Race_RaceUndead,
				Equipment:     core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: StandardTalents,
				Profession1:   proto.Profession_Enchanting,
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

// 36/15: Illumination, Divine Favor, Holy Shock, Holy Power and Light's Vigil, with Toughness,
// Precision, Anticipation and Sacred Duty behind them.
var StandardTalents = "55320030025131051-503050002"

var FullConsumes = &proto.ConsumesSpec{
	FlaskId: 13512,  // Flask of Supreme Power
	FoodId:  250066, // Soaring Pamplona
	PotId:   250949, // Major Mender's Potion
}

var PlayerOptions = &proto.Player_HolyPaladin{
	HolyPaladin: &proto.HolyPaladin{
		Options: &proto.HolyPaladin_Options{
			ClassOptions: &proto.PaladinOptions{},
		},
	},
}

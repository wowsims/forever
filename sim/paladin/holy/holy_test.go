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

// 35/16: Illumination, Divine Favor, Oath of Light, Holy Shock, Boundless Light, Holy Power and
// Light's Vigil, with Divine Intellect, Precision, Sacred Duty, Toughness and Anticipation behind them.
var StandardTalents = "55320015022101251-503020005001"

var FullConsumes = &proto.ConsumesSpec{
	FlaskId:        13512,  // Flask of Supreme Power
	FoodId:         250066, // Soaring Pamplona
	OffensivePotId: 250949, // Major Mender's Potion
}

var PlayerOptions = &proto.Player_HolyPaladin{
	HolyPaladin: &proto.HolyPaladin{
		Options: &proto.HolyPaladin_Options{
			ClassOptions: &proto.PaladinOptions{},
		},
	},
}

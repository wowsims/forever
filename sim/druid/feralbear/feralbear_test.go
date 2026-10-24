package feralbear

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterFeralBearDruid()
	common.RegisterAllEffects()
}

func TestFeralBear(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},

			GearSet: core.GetGearSet("../../../ui/specs/druid/feralbear/gear_sets", "default"),

			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "Hybrid", Talents: HybridTalents},
			},

			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},
			OtherSpecOptions: []core.SpecOptionsCombo{
				{Label: "BleedAssumed", SpecOptions: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 25, AssumeBleedActive: true},
				}}},
			},

			Rotation: core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default"),

			Consumables: DefaultConsumables,

			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Enchanting,

			IndividualBuffs: core.FullTankIndividualBuffs,

			IsTank:          true,
			InFrontOfTarget: true,

			ItemFilter: core.ItemFilter{
				ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeDagger,
					proto.WeaponType_WeaponTypeFist,
					proto.WeaponType_WeaponTypeMace,
					proto.WeaponType_WeaponTypeStaff,
				},
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeIdol,
				},
			},

			EPReferenceStat: proto.Stat_StatAgility,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatHealth,
				proto.Stat_StatStamina,
				proto.Stat_StatAgility,
				proto.Stat_StatStrength,
				proto.Stat_StatAttackPower,
				proto.Stat_StatArmor,
				proto.Stat_StatBonusArmor,
				proto.Stat_StatDodgeRating,
				proto.Stat_StatDefenseRating,
				proto.Stat_StatHitRating,
				proto.Stat_StatExpertiseRating,
			},
		},
	}))
}

func BenchmarkSimulate(b *testing.B) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Class:         proto.Class_ClassDruid,
				Race:          proto.Race_RaceNightElf,
				TalentsString: DefaultTalents,
				Equipment:     core.GetGearSet("../../../ui/specs/druid/feralbear/gear_sets", "default").GearSet,
				Consumables:   DefaultConsumables,
				Spec:          DefaultSpecOptions,
				Rotation:      core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation,
			},
			nil, nil, nil,
		),
		Encounter: &proto.Encounter{
			Duration: 300,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: core.AverageDefaultSimTestOptions,
	}

	core.RaidBenchmark(b, rsr)
}

// Fixed allocations for the class benchmark.
const DefaultTalents = "050022-15230332121132012051-04"
const HybridTalents = "050022-05230332121131212051-04"

var DefaultSpecOptions = &proto.Player_FeralBearDruid{
	FeralBearDruid: &proto.FeralBearDruid{
		Options: &proto.FeralBearDruid_Options{
			StartingRage: 25,
		},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	DefensivePotId: 13455,                // Greater Stoneshield Potion
	ElixirIds:      []int32{13452, 9088}, // Elixir of the Mongoose, Gift of Arthas
	FoodId:         21023,                // Dirge's Kickin' Chimaerok Chops
	ConjuredId:     9421,                 // Major Healthstone
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV, Scroll of Protection IV
	ExtraIds: []int32{10646, 10309, 10310, 10305},
}

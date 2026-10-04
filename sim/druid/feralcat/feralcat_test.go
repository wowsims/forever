package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterFeralCatDruid()
	common.RegisterAllEffects()
}

func TestFeralCat(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the expected .results numbers cannot be meaningful until then")
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},

			GearSet: core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_realistic_6p"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "pre_raid"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_realistic_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_bis_6p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_bis_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_alt_6p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_alt_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p2_6p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p2_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p2_alt_6p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p2_alt_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p3_6p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p3_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p4_6p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p4_9p"),
				core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p5"),
			},

			Talents: DefaultTalents,
			OtherTalentSets: []core.TalentsCombo{
				{Label: "Monocat", Talents: MonocatTalents},
			},

			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: DefaultSpecOptions},

			Rotation: core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default"),
			OtherRotations: []core.RotationCombo{
				{
					Label: "Simple",
					Rotation: &proto.APLRotation{
						Type: proto.APLRotation_TypeSimple,
					},
				},
			},

			Consumables: DefaultConsumables,

			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Enchanting,

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

			EPReferenceStat: proto.Stat_StatAttackPower,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatAgility,
				proto.Stat_StatStrength,
				proto.Stat_StatAttackPower,
				proto.Stat_StatFeralAttackPower,
				proto.Stat_StatHitRating,
				proto.Stat_StatExpertiseRating,
				proto.Stat_StatCritRating,
				proto.Stat_StatHasteRating,
				proto.Stat_StatArmorPenetration,
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
				Equipment:     core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "p1_realistic_6p").GearSet,
				Consumables:   DefaultConsumables,
				Spec:          DefaultSpecOptions,
				Rotation: &proto.APLRotation{
					Type: proto.APLRotation_TypeSimple,
				},
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

const DefaultTalents = "-50303231320221050012-05503301"
const MonocatTalents = "-5530023132022105001-05503301"

var DefaultSpecOptions = &proto.Player_FeralCatDruid{
	FeralCatDruid: &proto.FeralCatDruid{
		Rotation: &proto.FeralCatDruid_Rotation{
			FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
			Biteweave:          true,
			RipMinComboPoints:  5,
			BiteMinComboPoints: 5,
			MangleTrick:        true,
			MaintainFaerieFire: false,
		},
		Options: &proto.FeralCatDruid_Options{},
	},
}

var DefaultConsumables = &proto.ConsumesSpec{
	PotId:      250943,                // Major Frenzy Potion
	ElixirIds:  []int32{13452, 13447}, // Elixir of the Mongoose, Elixir of the Sages
	FoodId:     250072,                // Swiftstrike Steak
	MhImbueId:  12643,                 // Dense Weightstone
	ConjuredId: 12662,                 // Demonic Rune
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV
	ExtraIds: []int32{10646, 10309, 10310},
}

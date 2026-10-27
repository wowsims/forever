package dps

import (
	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"

	"testing"
)

func init() {
	RegisterDpsWarrior()
	common.RegisterAllEffects()
}

func TestDpsWarrior(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{dpsWarriorSuite()}))
}

func BenchmarkDpsWarrior(b *testing.B) {
	fury := dpsWarriorSuite()
	arms := fury
	arms.GearSet = core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p1_arms")
	arms.Talents = DefaultArmsTalents
	arms.Rotation = core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "arms")

	b.Run("Fury", func(b *testing.B) { core.CharacterBenchmark(b, fury) })
	b.Run("Arms", func(b *testing.B) { core.CharacterBenchmark(b, arms) })
}

func dpsWarriorSuite() core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassWarrior,
		Race:       proto.Race_RaceOrc,
		OtherRaces: []proto.Race{proto.Race_RaceHuman},
		GearSet:    core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p1_fury"),
		OtherGearSets: []core.GearSetCombo{
			core.GetGearSet("../../../ui/specs/warrior/dps/gear_sets", "p1_arms"),
		},
		Talents: DefaultFuryTalents,
		OtherTalentSets: []core.TalentsCombo{
			{Label: "Arms", Talents: DefaultArmsTalents},
		},
		Consumables:      DefaultConsumables,
		SpecOptions:      core.SpecOptionsCombo{Label: "Fury", SpecOptions: DefaultOptions},
		StartingDistance: 0,
		Profession1:      proto.Profession_Engineering,
		Profession2:      proto.Profession_Blacksmithing,

		Rotation: core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "fury"),
		OtherRotations: []core.RotationCombo{
			core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "arms"),
			armsStanceDanceRotation(),
		},

		ItemFilter: core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypeLeather,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeFist,
				proto.WeaponType_WeaponTypeMace,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeAxe,
			},
			HandTypes: []proto.HandType{
				proto.HandType_HandTypeMainHand,
				proto.HandType_HandTypeOffHand,
				proto.HandType_HandTypeOneHand,
				proto.HandType_HandTypeTwoHand,
			},
		},
	}
}

func armsStanceDanceRotation() core.RotationCombo {
	rotation := core.GetAplRotation("../../../ui/specs/warrior/dps/apls", "arms").Rotation
	for _, item := range rotation.PriorityList {
		if item.GetAction().GetGroupReference().GetGroupName() == "Stance Dance ON/OFF" {
			item.Hide = false
		}
	}
	return core.RotationCombo{Label: "arms_stance_dance", Rotation: rotation}
}

var DefaultOptions = &proto.Player_DpsWarrior{
	DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				UseBattleShout: true,
				DefaultStance:  proto.WarriorStance_WarriorStanceBerserker,
			},
		},
	},
}

var DefaultFuryTalents = "30305213002-15053105102210501"
var DefaultArmsTalents = "30325213132515201-0505-2"

var DefaultConsumables = &proto.ConsumesSpec{
	ResourcePotId: 13442,  // Mighty Rage Potion
	FlaskId:       274274, // Flask of Natural Aggression
	FoodId:        20452,  // Smoked Desert Dumplings
	ExplosiveId:   18588,
	OhImbueId:     12404,
	// Goblin Sapper Charge, Scroll of Agility IV, Scroll of Strength IV
	ExtraIds: []int32{10646, 10309, 10310},
}

package dps

import (
	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"fmt"
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
			trainerRanksRotation("fury"),
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

// The preset with its class book ranks moved to the trainer's top ranks, as the simple rotation's toggles left
// off move them, so the trainer ranks reach the results.
func trainerRanksRotation(preset string) core.RotationCombo {
	rotation := core.GetAplRotation("../../../ui/specs/warrior/dps/apls", preset).Rotation
	if withSpellIds(rotation.ProtoReflect(), map[int32]int32{25286: 11567, 25289: 11551}) == 0 {
		panic(fmt.Sprintf("the %s preset names no book rank", preset))
	}
	return core.RotationCombo{Label: preset + "_trainer_ranks", Rotation: rotation}
}

// Swaps every spell id that ids maps for the id it maps to, except an external buff copy (tag -1), and returns
// how many it swapped.
func withSpellIds(message protoreflect.Message, ids map[int32]int32) int {
	if id, ok := message.Interface().(*proto.ActionID); ok {
		if to, ok := ids[id.GetSpellId()]; ok && id.Tag != -1 {
			id.RawId = &proto.ActionID_SpellId{SpellId: to}
			return 1
		}
		return 0
	}
	swapped := 0
	message.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		switch {
		case field.IsList() && field.Message() != nil:
			for i := 0; i < value.List().Len(); i++ {
				swapped += withSpellIds(value.List().Get(i).Message(), ids)
			}
		case field.Message() != nil && !field.IsMap():
			swapped += withSpellIds(value.Message(), ids)
		}
		return true
	})
	return swapped
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

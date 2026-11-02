package enhancement

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

func TestEnhancement(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:       proto.Class_ClassShaman,
			Race:        proto.Race_RaceOrc,
			OtherRaces:  []proto.Race{proto.Race_RaceTroll},
			SpecOptions: core.SpecOptionsCombo{Label: "Windfury", SpecOptions: imbueOptions(proto.ShamanImbue_WindfuryWeapon)},
			OtherSpecOptions: []core.SpecOptionsCombo{
				{Label: "Flametongue", SpecOptions: imbueOptions(proto.ShamanImbue_FlametongueWeapon)},
				{Label: "Frostbrand", SpecOptions: imbueOptions(proto.ShamanImbue_FrostbrandWeapon)},
				{Label: "Rockbiter", SpecOptions: imbueOptions(proto.ShamanImbue_RockbiterWeapon)},
			},
			GearSet:     testingBuild.GearSet,
			Talents:     DefaultTalents,
			Consumables: DefaultConsumables,
			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Leatherworking,
			Rotation:    testingBuild.Rotation,
			OtherRotations: []core.RotationCombo{
				core.RotationWithSpellIds("testing_book_ranks", testingBuild.Rotation, map[int32]int32{10442: 25361}),
			},
			ItemFilter: core.ItemFilter{
				WeaponTypes: []proto.WeaponType{
					proto.WeaponType_WeaponTypeAxe,
					proto.WeaponType_WeaponTypeMace,
				},
				HandTypes: []proto.HandType{proto.HandType_HandTypeTwoHand},
				ArmorType: proto.ArmorType_ArmorTypeMail,
				RangedWeaponTypes: []proto.RangedWeaponType{
					proto.RangedWeaponType_RangedWeaponTypeTotem,
				},
			},
		},
	}))
}

func imbueOptions(imbue proto.ShamanImbue) *proto.Player_EnhancementShaman {
	return &proto.Player_EnhancementShaman{
		EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{
				ClassOptions: &proto.ShamanOptions{
					ImbueMh: imbue,
				},
			},
		},
	}
}

// The Testing build's gear, talents and rotation.
var testingBuild = core.GetTestBuildFromJSON(proto.Class_ClassShaman, "../../../ui/specs/shaman/enhancement/builds", "testing",
	core.ItemFilter{}, nil, nil)

var DefaultTalents = testingBuild.Talents

// The UI's Enhancement defaults: Major Mana Potion, Flask of Supreme Power, Smoked Desert Dumplings,
// Demonic Rune, the Owl, Sages, Greater Mageblood, Nature Power, Greater Arcane, Cunning and Ferocity
// elixirs, Winterfall Firewater, EZ-Thro Dark Bomb, Goblin Sapper Charge, and Elemental Sharpening Stone
// on the main hand beside the shaman's imbue.
var DefaultConsumables = &proto.ConsumesSpec{
	ResourcePotId: 13444,
	FlaskId:       13512,
	FoodId:        20452,
	ConjuredId:    12662,
	ElixirIds:     []int32{250337, 250338, 250341, 250343, 13454, 12820, 250328, 250350},
	ExplosiveId:   260817,
	ExtraIds:      []int32{10646},
	MhImbueId:     18262,
}

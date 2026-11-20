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
			GearSet:     defaultBuild.GearSet,
			Talents:     DefaultTalents,
			Consumables: defaultBuild.Consumables,
			Profession1: proto.Profession_Engineering,
			Profession2: proto.Profession_Leatherworking,
			Rotation:    defaultBuild.Rotation,
			OtherRotations: []core.RotationCombo{
				core.RotationWithSpellIds("testing_book_ranks", defaultBuild.Rotation, map[int32]int32{10442: 25361}),
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

// The Default build's gear, talents, consumables and rotation, which the UI's defaults read too.
var defaultBuild = core.GetTestBuildFromJSON(proto.Class_ClassShaman, "../../../ui/specs/shaman/enhancement/builds", "default",
	core.ItemFilter{}, nil, nil)

var DefaultTalents = defaultBuild.Talents

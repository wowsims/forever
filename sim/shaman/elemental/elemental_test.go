package elemental

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

func init() {
	RegisterElementalShaman()
	common.RegisterAllEffects()
}

func TestElemental(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassShaman,
			Race:       proto.Race_RaceWindshaperSkyborne,
			OtherRaces: []proto.Race{proto.Race_RaceTroll, proto.Race_RaceOrc, proto.Race_RaceDwarf},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_ElementalShaman{
				ElementalShaman: &proto.ElementalShaman{
					Options: &proto.ElementalShaman_Options{
						ClassOptions: &proto.ShamanOptions{
							ShieldProcrate: 0.0,
						},
					},
				},
			}},
			GearSet:     core.GearSetCombo{Label: "default", GearSet: defaults.Equipment()},
			Talents:     DefaultTalents,
			Consumables: defaults.Settings.GetPlayer().GetConsumables(),
			Profession1: proto.Profession_Leatherworking,
			Profession2: proto.Profession_Enchanting,
			Rotation:    core.RotationCombo{Label: "default", Rotation: defaults.Rotation()},
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
		},
	}))
}

// The page's defaults, whose gear, talents, consumables and rotation the suite sims.
var defaults = readDefaults()

var DefaultTalents = defaults.Settings.GetPlayer().GetTalentsString()

func readDefaults() *siml.File {
	f, err := siml.ReadFile("../../../ui/specs/shaman/elemental/defaults.siml")
	if err != nil {
		panic(err)
	}
	return f
}

const DefaultArmorType = proto.ArmorType_ArmorTypeMail

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeAxe,
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeFist,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeShield,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeTotem,
}

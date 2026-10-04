package elemental

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterElementalShaman()
	common.RegisterAllEffects()
}

func TestElemental(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassShaman,
			Race:       proto.Race_RaceTroll,
			OtherRaces: []proto.Race{proto.Race_RaceOrc, proto.Race_RaceDwarf},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_ElementalShaman{
				ElementalShaman: &proto.ElementalShaman{
					Options: &proto.ElementalShaman_Options{
						ClassOptions: &proto.ShamanOptions{
							ShieldProcrate: 0.0,
						},
					},
				},
			}},
			GearSet:     core.GetGearSet("../../../ui/specs/shaman/elemental/gear_sets", "preraid"),
			Talents:     DefaultTalents,
			Consumables: DefaultConsumables,
			Profession1: proto.Profession_Leatherworking,
			Profession2: proto.Profession_Enchanting,
			Rotation:    core.GetAplRotation("../../../ui/specs/shaman/elemental/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
		},
	}))
}

const DefaultTalents = "4500301323123051-05-05305"

// The UI's Elemental defaults: Major Mana Potion, Flask of Supreme Power, Demonic Rune and Brilliant
// Wizard Oil.
var DefaultConsumables = &proto.ConsumesSpec{
	PotId:      13444,
	FlaskId:    13512,
	ConjuredId: 12662,
	MhImbueId:  20749,
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

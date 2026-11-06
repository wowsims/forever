package protection

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterProtectionPaladin()
	common.RegisterAllEffects()
}

func TestProtection(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{protectionSuite()}))
}

func BenchmarkProtection(b *testing.B) {
	core.CharacterBenchmark(b, protectionSuite())
}

func protectionSuite() core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:            proto.Class_ClassPaladin,
		Race:             proto.Race_RaceUndead,
		OtherRaces:       []proto.Race{proto.Race_RaceHuman},
		GearSet:          core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default"),
		Talents:          DefaultProtectionTalents,
		Consumables:      DefaultConsumables,
		SpecOptions:      core.SpecOptionsCombo{Label: "Protection", SpecOptions: DefaultOptions},
		StartingDistance: 5,
		Profession1:      proto.Profession_Engineering,
		Profession2:      proto.Profession_Enchanting,

		Rotation: core.GetAplRotation("../../../ui/specs/paladin/protection/apls", "default"),

		IndividualBuffs: core.FullTankIndividualBuffs,

		IsTank:          true,
		InFrontOfTarget: true,

		ItemFilter: core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypePlate,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeAxe,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeMace,
				proto.WeaponType_WeaponTypeShield,
			},
			HandTypes: []proto.HandType{
				proto.HandType_HandTypeMainHand,
				proto.HandType_HandTypeOffHand,
				proto.HandType_HandTypeOneHand,
			},
			RangedWeaponTypes: []proto.RangedWeaponType{
				proto.RangedWeaponType_RangedWeaponTypeLibram,
			},
		},
	}
}

var DefaultOptions = &proto.Player_ProtectionPaladin{
	ProtectionPaladin: &proto.ProtectionPaladin{
		Options: &proto.ProtectionPaladin_Options{
			ClassOptions: &proto.PaladinOptions{},
		},
	},
}

var DefaultProtectionTalents = "050050002-05302102213301531-502"

var DefaultConsumables = &proto.ConsumesSpec{
	ConjuredId: 12662, // Demonic Rune
}

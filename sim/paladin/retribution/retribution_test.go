package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterRetributionPaladin()
	common.RegisterAllEffects()
}

func TestRetribution(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{retributionSuite()}))
}

func BenchmarkRetribution(b *testing.B) {
	twist := retributionSuite()
	shock := twist
	shock.GearSet = core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "holy_shock")
	shock.Talents = HolyShockTalents
	shock.Rotation = core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "holy_shock")
	shock.ItemFilter.WeaponTypes = []proto.WeaponType{
		proto.WeaponType_WeaponTypeAxe,
		proto.WeaponType_WeaponTypeMace,
		proto.WeaponType_WeaponTypeSword,
		proto.WeaponType_WeaponTypeShield,
		proto.WeaponType_WeaponTypeOffHand,
	}
	shock.ItemFilter.HandTypes = []proto.HandType{
		proto.HandType_HandTypeMainHand,
		proto.HandType_HandTypeOneHand,
		proto.HandType_HandTypeTwoHand,
	}

	b.Run("Twist of Light", func(b *testing.B) { core.CharacterBenchmark(b, twist) })
	b.Run("Holy Shock", func(b *testing.B) { core.CharacterBenchmark(b, shock) })
}

func retributionSuite() core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassPaladin,
		Race:       proto.Race_RaceUndead,
		OtherRaces: []proto.Race{proto.Race_RaceHuman},
		SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: &proto.Player_RetributionPaladin{
			RetributionPaladin: &proto.RetributionPaladin{
				Options: &proto.RetributionPaladin_Options{
					ClassOptions: &proto.PaladinOptions{},
				},
			},
		}},
		Consumables: DefaultConsumables,
		Profession1: proto.Profession_Engineering,
		Profession2: proto.Profession_Blacksmithing,
		GearSet:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "twist_of_light"),
		Talents:     TwistOfLightTalents,
		Rotation:    core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "twist_of_light"),
		ItemFilter: core.ItemFilter{
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypePolearm,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeAxe,
				proto.WeaponType_WeaponTypeMace,
			},
			ArmorType: proto.ArmorType_ArmorTypePlate,
			HandTypes: []proto.HandType{proto.HandType_HandTypeTwoHand},
			RangedWeaponTypes: []proto.RangedWeaponType{
				proto.RangedWeaponType_RangedWeaponTypeLibram,
			},
		},
	}
}

var HolyShockTalents = "550030032101012--052252300012303"
var TwistOfLightTalents = "55003003--05225331101330321"

var DefaultConsumables = &proto.ConsumesSpec{
	ConjuredId: 12662,          // Demonic Rune
	FlaskId: 13512, // Flask of Supreme Power
	FoodId: 20452, // Smoked Desert Dumplings
	ElixirIds:  []int32{
		21546, // Elixir of Holy Power
		250338, // Elixir of Sages
		13452, // Elixir of The Mongoose
		250337, // Elixir of The Owl
		250351, // Elixir of The Grizzly
		13454, // Greater Arcane Elixir
		250341, // Greater Mageblood Elixir
		12460, // Juju Might
	},
	ExplosiveId: 18588, // Ez-Thro Dynamite II
	ExtraIds:   []int32{10646, 8410}, // Goblin Sapper Charge, R.O.I.D.S
	MhImbueId: 20749, // Brilliant Wizard Oil
	PotId: 250937, // Major Spellblasting Potion
}

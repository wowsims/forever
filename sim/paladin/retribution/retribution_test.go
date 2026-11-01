package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	goproto "google.golang.org/protobuf/proto"
)

func init() {
	RegisterRetributionPaladin()
	common.RegisterAllEffects()
}

func TestRetribution(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{retributionSuite(), shockadinSuite()}))
}

func BenchmarkRetribution(b *testing.B) {
	unsealed := retributionSuite()
	shock := shockadinSuite()

	b.Run("Unsealed Wrath", func(b *testing.B) { core.CharacterBenchmark(b, unsealed) })
	b.Run("Holy Shock", func(b *testing.B) { core.CharacterBenchmark(b, shock) })
}

func shockadinSuite() core.CharacterSuiteConfig {
	consumables := goproto.Clone(DefaultConsumables).(*proto.ConsumesSpec)
	consumables.ElixirIds[0] = 8423  // Cerebral Cortex Compound
	consumables.MhImbueId = 3829     // Frost Oil
	consumables.SwapMhImbueId = 3824 // Shadow Oil

	return core.CharacterSuiteConfig{
		Class:            proto.Class_ClassPaladin,
		Race:             proto.Race_RaceUndead,
		OtherRaces:       []proto.Race{proto.Race_RaceHuman},
		SpecOptions:      core.SpecOptionsCombo{Label: "HolyShock", SpecOptions: DefaultSpecOptions},
		StartingDistance: 25,
		Consumables:      consumables,
		Profession1:      proto.Profession_Engineering,
		Profession2:      proto.Profession_Blacksmithing,
		GearSet:          core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "holy_shock"),
		ItemSwapSet:      core.GetItemSwapGearSet("../../../ui/specs/paladin/retribution/gear_sets", "holy_shock_item_swap"),
		Talents:          HolyShockTalents,
		Rotation:         core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "holy_shock"),
		ItemFilter: core.ItemFilter{
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeAxe,
				proto.WeaponType_WeaponTypeMace,
				proto.WeaponType_WeaponTypeOffHand,
				proto.WeaponType_WeaponTypePolearm,
				proto.WeaponType_WeaponTypeShield,
				proto.WeaponType_WeaponTypeSword,
			},
			ArmorType: proto.ArmorType_ArmorTypePlate,
			HandTypes: []proto.HandType{
				proto.HandType_HandTypeMainHand,
				proto.HandType_HandTypeOneHand,
				proto.HandType_HandTypeTwoHand,
			},
			RangedWeaponTypes: []proto.RangedWeaponType{
				proto.RangedWeaponType_RangedWeaponTypeLibram,
			},
		},
	}
}

func retributionSuite() core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:            proto.Class_ClassPaladin,
		Race:             proto.Race_RaceUndead,
		OtherRaces:       []proto.Race{proto.Race_RaceHuman},
		SpecOptions:      core.SpecOptionsCombo{Label: "Default", SpecOptions: DefaultSpecOptions},
		StartingDistance: 25,
		Consumables:      DefaultConsumables,
		Profession1:      proto.Profession_Engineering,
		Profession2:      proto.Profession_Blacksmithing,
		GearSet:          core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default"),
		Talents:          UnsealedWrathTalents,
		Rotation:         core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default"),
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

var DefaultSpecOptions = &proto.Player_RetributionPaladin{
	RetributionPaladin: &proto.RetributionPaladin{
		Options: &proto.RetributionPaladin_Options{
			ClassOptions: &proto.PaladinOptions{},
		},
	},
}

var HolyShockTalents = "05025210221021--0523330110130033"
var UnsealedWrathTalents = "050250002-501-05233011201301351"

var DefaultConsumables = &proto.ConsumesSpec{
	ConjuredId: 12662, // Demonic Rune
	FlaskId:    13512, // Flask of Supreme Power
	FoodId:     13933, // Lobster Stew
	ElixirIds: []int32{
		21546,  // Elixir of Holy Power
		250338, // Elixir of Sages
		13452,  // Elixir of The Mongoose
		250337, // Elixir of The Owl
		250351, // Elixir of The Grizzly
		250329, // Elixir of The Phalanx
		13454,  // Greater Arcane Elixir
		250341, // Greater Mageblood Elixir
		12460,  // Juju Might
	},
	ExtraIds:       []int32{8410, 10646, 12217, 18269}, // R.O.I.D.S, Goblin Sapper Charge, Dragonbreath Chili, Gordok Green Grog
	MhImbueId:      3824,                               // Shadow Oil
	OffensivePotId: 250937,                             // Major Spellblasting Potion
	ResourcePotId:  13444,                              // Major Mana Potion
}

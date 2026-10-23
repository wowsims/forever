package mage

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterMage()
	common.RegisterAllEffects()
}

func TestArcane(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{arcaneSuite()}))
}

func TestFFB(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{ffbSuite()}))
}

func TestFire(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{fireSuite()}))
}

func TestFrost(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{frostSuite()}))
}

func BenchmarkMage(b *testing.B) {
	b.Run("Arcane", func(b *testing.B) { core.CharacterBenchmark(b, arcaneSuite()) })
	b.Run("FFB", func(b *testing.B) { core.CharacterBenchmark(b, ffbSuite()) })
	b.Run("Fire", func(b *testing.B) { core.CharacterBenchmark(b, fireSuite()) })
	b.Run("Frost", func(b *testing.B) { core.CharacterBenchmark(b, frostSuite()) })
}

func arcaneSuite() core.CharacterSuiteConfig {
	return mageSuite(ArcaneTalents, "arcaneFrostFill")
}

func ffbSuite() core.CharacterSuiteConfig {
	return mageSuite(FFBTalents, "ffb")
}

func fireSuite() core.CharacterSuiteConfig {
	return mageSuite(FireTalents, "deepFire")
}

func frostSuite() core.CharacterSuiteConfig {
	return mageSuite(FrostTalents, "frost")
}

func mageSuite(talents string, apl string) core.CharacterSuiteConfig {
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassMage,
		Race:       proto.Race_RaceHighOrderSkyborne,
		OtherRaces: []proto.Race{proto.Race_RaceGnome},
		SpecOptions: core.SpecOptionsCombo{Label: "Default", SpecOptions: &proto.Player_Mage{
			Mage: &proto.Mage{
				Options: &proto.Mage_Options{
					ClassOptions: &proto.MageOptions{
						DefaultMageArmor: proto.MageArmor_MageArmorMageArmor,
					},
				},
			},
		}},
		Consumables: DefaultConsumables,
		Profession1: proto.Profession_Engineering,
		Profession2: proto.Profession_Tailoring,
		GearSet:     core.GetGearSet("../../ui/specs/mage/dps/gear_sets", "p1Arcane"),
		Talents:     talents,
		Rotation:    core.GetAplRotation("../../ui/specs/mage/dps/apls", apl),
		ItemFilter: core.ItemFilter{
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeOffHand,
				proto.WeaponType_WeaponTypeStaff,
			},
			ArmorType: proto.ArmorType_ArmorTypeCloth,
			RangedWeaponTypes: []proto.RangedWeaponType{
				proto.RangedWeaponType_RangedWeaponTypeWand,
			},
			EnchantBlacklist: []int32{2673, 3225, 3273},
		},
	}
}

var ArcaneTalents = "0552205003100311531-03-0533"
var FFBTalents = "-2355000013011503-2035010311013002"
var FireTalents = "05022050001001-23550000130135031-003"
var FrostTalents = "05022050011003--053500030100331251"

var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId: 13512, // Flask of Supreme Power
	FoodId:  13931, // Nightfin Soup
	ElixirIds: []int32{
		13454,  // Greater Arcane Elixir
		13447,  // Elixir of the Sages
		250348, // Elixir of the Whale
		250341, // Greater Mageblood Elixir
	},
	MhImbueId:       20749,  // Brilliant Wizard Oil
	MhWeaponImbueId: 277502, // Scroll of Imbue Precision
	ResourcePotId:   13444,  // Major Mana Potion
}

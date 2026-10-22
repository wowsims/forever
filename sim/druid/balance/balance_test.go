package balance

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterBalanceDruid()
	common.RegisterAllEffects()
}

func TestBalance(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_BalanceDruid{
				BalanceDruid: &proto.BalanceDruid{
					Options: &proto.BalanceDruid_Options{
						ClassOptions: &proto.DruidOptions{},
					},
				},
			}},
			GearSet: core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p1"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "prebis"),
			},
			Talents:          DefaultTalents,
			Rotation:         core.GetAplRotation("../../../ui/specs/druid/balance/apls", "default"),
			Consumables:      DefaultConsumables,
			Profession1:      proto.Profession_Enchanting,
			Profession2:      proto.Profession_Tailoring,
			StartingDistance: 20,
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
			EPReferenceStat: proto.Stat_StatSpellDamage,
			StatsToWeigh: []proto.Stat{
				proto.Stat_StatIntellect,
				proto.Stat_StatSpellDamage,
				proto.Stat_StatArcaneDamage,
				proto.Stat_StatNatureDamage,
				proto.Stat_StatSpellPiercing,
				proto.Stat_StatHitRating,
				proto.Stat_StatCritRating,
				proto.Stat_StatHasteRating,
				proto.Stat_StatMP5,
				proto.Stat_StatSpirit,
			},
		},
	}))
}

// The Moonkin preset on the Balance page.
const DefaultTalents = "5232220115501351--505003"

const DefaultArmorType = proto.ArmorType_ArmorTypeLeather

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeOffHand,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeIdol,
}

// The Balance page's consumes: Flask of Supreme Power, Greater Arcane Elixir, Elixir of Nature Power,
// Elixir of the Owl, Greater Mageblood Elixir, Lobster Stew, Brilliant Wizard Oil, Major Spellblasting
// Potion, Major Mana Potion, Demonic Rune and SAF-T Clever Dynamite.
var DefaultConsumables = &proto.ConsumesSpec{
	FlaskId:        13512,
	ElixirIds:      []int32{13454, 250343, 250337, 250341},
	FoodId:         13933,
	MhImbueId:      20749,
	OffensivePotId: 250937,
	ResourcePotId:  13444,
	ConjuredId:     12662,
	ExplosiveId:    260814,
}

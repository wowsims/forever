package protection

import (
	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"

	"testing"
)

func init() {
	RegisterProtectionWarrior()
	common.RegisterAllEffects()
}

func TestProtectionWarrior(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{protectionWarriorSuite()}))
}

func BenchmarkProtectionWarrior(b *testing.B) {
	core.CharacterBenchmark(b, protectionWarriorSuite())
}

func protectionWarriorSuite() core.CharacterSuiteConfig {
	preset := core.GetAplRotation("../../../ui/specs/warrior/protection/apls", "default")
	return core.CharacterSuiteConfig{
		Class:      proto.Class_ClassWarrior,
		Race:       proto.Race_RaceOrc,
		OtherRaces: []proto.Race{proto.Race_RaceHuman},
		GearSet:    core.GetGearSet("../../../ui/specs/warrior/protection/gear_sets", "prebis_balanced"),
		OtherGearSets: []core.GearSetCombo{
			core.GetGearSet("../../../ui/specs/warrior/protection/gear_sets", "prebis_offensive"),
			core.GetGearSet("../../../ui/specs/warrior/protection/gear_sets", "p1_bis"),
		},
		Talents:          DefaultProtectionTalents,
		OtherTalentSets:  []core.TalentsCombo{{Label: "Offensive", Talents: OffensiveProtectionTalents}},
		Consumables:      DefaultConsumables,
		SpecOptions:      core.SpecOptionsCombo{Label: "Protection", SpecOptions: DefaultOptions},
		StartingDistance: 0,
		Profession1:      proto.Profession_Engineering,
		Profession2:      proto.Profession_Blacksmithing,

		Rotation: preset,
		OtherRotations: []core.RotationCombo{
			core.RotationWithSpellIds("default_book_ranks", preset, map[int32]int32{11567: 25286, 11551: 25289, 11601: 25288}),
		},

		IndividualBuffs: core.FullTankIndividualBuffs,

		IsTank:          true,
		InFrontOfTarget: true,

		ItemFilter: core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypePlate,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeFist,
				proto.WeaponType_WeaponTypeMace,
				proto.WeaponType_WeaponTypeSword,
				proto.WeaponType_WeaponTypeAxe,
				proto.WeaponType_WeaponTypeShield,
			},
			HandTypes: []proto.HandType{
				proto.HandType_HandTypeMainHand,
				proto.HandType_HandTypeOffHand,
				proto.HandType_HandTypeOneHand,
			},
		},
	}
}

var DefaultOptions = &proto.Player_ProtectionWarrior{
	ProtectionWarrior: &proto.ProtectionWarrior{
		Options: &proto.ProtectionWarrior_Options{
			ClassOptions: &proto.WarriorOptions{
				StartingRage:   0,
				UseBattleShout: true,
				DefaultStance:  proto.WarriorStance_WarriorStanceDefensive,
				QueueDelay:     250,
				StanceSnapshot: true,
				HasBsT2:        true,
			},
		},
	},
}

var DefaultProtectionTalents = "05-05050005-150531120300001351"
var OffensiveProtectionTalents = "3-0505000501-050032120331021351"

var DefaultConsumables = &proto.ConsumesSpec{
	OffensivePotId: 250943, // Major Frenzy Potion
	FoodId:         20452,  // Smoked Desert Dumplings
	ExplosiveId:    260814, // SAF-T Clever Dynamite
	MhImbueId:      3241,   // Heavy Weightstone
	ConjuredId:     9421,   // Major Healthstone
	// Elixir of the Mongoose, Gift of Arthas, Elixir of Greater Fortitude, Elixir of Greater Defense
	ElixirIds: []int32{13452, 9088, 250335, 13445},
	// Goblin Sapper Charge, Rumsey Rum, Juju Power, Juju Might, Bogling Root, R.O.I.D.S.
	ExtraIds: []int32{10646, 17048, 12451, 12460, 5206, 8410},
}

package protection

import (
	"testing"

	"github.com/wowsims/forever/sim/common" // imported to get item effects included.
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
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

// A talents string with more points in Shield Specialization than it takes builds at its top rank,
// so the digit a build saved before the tree was reworked registers the talent instead of panicking.
func TestShieldSpecializationOverItsRanksBuildsAtTheTopRank(t *testing.T) {
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet,
		TalentsString: "050050002-05302103213301531-502", // the default talents with Shield Specialization at 3
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, DefaultOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	if pally.Talents.ShieldSpecialization != 2 {
		t.Fatalf("Shield Specialization at 3 points fills the talent at %d, want its top rank 2", pally.Talents.ShieldSpecialization)
	}
	if pally.GetAura("Shield Specialization"+pally.Label) == nil {
		t.Fatal("Shield Specialization's block trigger is not registered")
	}
}

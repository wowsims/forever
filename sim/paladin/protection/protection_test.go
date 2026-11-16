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

// Windfury Totem's strike, landed or avoided, spends Reckoning's stored charges as a main-hand swing
// does but swings none of them; the next main-hand swing swings each stored charge.
func TestWindfuryTotemsStrikeSpendsReckoningWithNoSwing(t *testing.T) {
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet,
		TalentsString: DefaultProtectionTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, DefaultOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{WindfuryTotem: true}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pally.AutoAttacks.CancelAutoSwing(sim)
	target := pally.CurrentTarget

	const reckoning = 1326084 // the stored charges
	stored := pally.GetAuraByID(core.ActionID{SpellID: reckoning})
	strike := pally.GetSpell(core.ActionID{SpellID: 10610})
	copied := pally.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: reckoning})
	if pally.Talents.Reckoning == 0 || stored == nil || strike == nil || copied == nil {
		t.Fatal("the paladin has no Reckoning, no stored charges, no Windfury Totem strike or no Reckoning swing")
	}
	settle := func() {
		for end := sim.CurrentTime + 2*core.SpellBatchWindow; sim.CurrentTime < end && !sim.Step(); {
		}
	}
	swings := func() int32 { return copied.SpellMetrics[target.UnitIndex].Casts }

	for _, outcome := range []core.HitOutcome{core.OutcomeHit, core.OutcomeDodge} {
		stored.Activate(sim)
		stored.SetStacks(sim, 2)
		before := swings()
		pally.OnSpellHitDealt(sim, strike, &core.SpellResult{Target: target, Outcome: outcome, Damage: 100})
		settle()
		if stored.IsActive() || swings() != before {
			t.Errorf("a strike with outcome %v left the charges up: %v, and swung %d of them, want spent and none swung", outcome, stored.IsActive(), swings()-before)
		}
	}

	stored.Activate(sim)
	stored.SetStacks(sim, 2)
	before := swings()
	pally.OnSpellHitDealt(sim, pally.AutoAttacks.MHAuto(), &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	settle()
	if stored.IsActive() || swings()-before != 2 {
		t.Errorf("a main-hand swing swung %d stored charges, want 2", swings()-before)
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

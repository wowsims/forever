package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// Conviction 5/5 gives the two-handed axe's attacks 5% crit and Hammer of Wrath, a ranged attack,
// none.
func TestConvictionReachesTheMeleeWeaponsAttacks(t *testing.T) {
	player := &proto.Player{
		Name:          "Conviction",
		Race:          proto.Race_RaceUndead,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: TwistOfLightTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()

	if pal.Talents.Conviction != 5 {
		t.Fatalf("the talents take Conviction %d/5, want 5/5", pal.Talents.Conviction)
	}
	if crit := pal.AutoAttacks.MHAuto().BonusCritPercent; crit != 5 {
		t.Errorf("main-hand auto crit = %0.2f%%, want 5%%", crit)
	}
	hammer := pal.GetSpell(core.ActionID{SpellID: paladin.HammerOfWrathRankMap.Highest().ID})
	if hammer == nil {
		t.Fatal("Hammer of Wrath is not registered")
	}
	if hammer.BonusCritPercent != 0 {
		t.Errorf("Hammer of Wrath crit = %0.2f%%, want 0%%", hammer.BonusCritPercent)
	}
}

package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// Casts Seal of Righteousness and one Judgement under the talents, and returns the seal's base cost,
// what the seal cost to cast, and the mana Sanctified Judgement returned.
func judgeSealOfRighteousness(t *testing.T, talents string) (baseCost, paid, refund float64) {
	t.Helper()
	player := &proto.Player{
		Name:          "Sanctified Judgement",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: talents,
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

	if pal.Talents.SanctifiedJudgement != 3 {
		t.Fatalf("the talents take Sanctified Judgement %d/3, want 3", pal.Talents.SanctifiedJudgement)
	}
	if pal.Talents.Benediction != 5 {
		t.Fatalf("the talents take Benediction %d/5, want 5", pal.Talents.Benediction)
	}

	target := pal.CurrentTarget
	seal := pal.GetSpell(core.ActionID{SpellID: paladin.SealOfRighteousnessRankMap.Highest().ID})
	if seal == nil {
		t.Fatal("Seal of Righteousness is not registered")
	}
	if !seal.Cast(sim, target) {
		t.Fatal("Seal of Righteousness did not cast")
	}

	before := pal.CurrentMana()
	if !pal.Judgement.Cast(sim, target) {
		t.Fatal("Judgement did not cast")
	}
	refund = pal.CurrentMana() - before + pal.Judgement.CurCast.Cost
	return float64(seal.Cost.BaseCost), seal.CurCast.Cost, refund
}

// Sanctified Judgement at 3/3 returns 60% of the judged seal's base mana cost on every Judgement,
// and Benediction's 10% cut on the seal is added to the refund instead of taken from it: a 90 mana
// seal costs 81 and returns 59.4 in game.
func TestSanctifiedJudgementReturnsSealMana(t *testing.T) {
	baseCost, paid, refund := judgeSealOfRighteousness(t, UnsealedWrathTalents)
	if math.Abs(paid-baseCost*0.9) > 1e-6 {
		t.Fatalf("Benediction 5/5 leaves the seal costing %.3f of %.0f, want 90%%", paid, baseCost)
	}
	if want := baseCost * 1.1 * 0.6; math.Abs(refund-want) > 1e-6 {
		t.Errorf("Judgement returned %.3f mana, want %.3f (seal base cost %.0f)", refund, want, baseCost)
	}
}

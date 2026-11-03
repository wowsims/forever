package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// The Unsealed Wrath build without its Conviction points.
var noConvictionTalents = "050250002-5-05203301101301351"

func convictionPaladin(t *testing.T, talents string) *paladin.Paladin {
	t.Helper()
	player := &proto.Player{
		Name:          "Conviction",
		Race:          proto.Race_RaceUndead,
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
	return sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
}

// Conviction 3/3 gives 6% crit to every spell and attack while the two-handed axe is equipped:
// the auto attacks, Holy Strike and Hammer of Wrath alike.
func TestConvictionGivesEverySpellAndAttackCrit(t *testing.T) {
	pal, base := convictionPaladin(t, UnsealedWrathTalents), convictionPaladin(t, noConvictionTalents)
	if pal.Talents.Conviction != 3 || base.Talents.Conviction != 0 {
		t.Fatalf("the talents take Conviction %d/3 and %d/3, want 3/3 and none", pal.Talents.Conviction, base.Talents.Conviction)
	}

	for _, spell := range []struct {
		name string
		id   core.ActionID
	}{
		{"the main-hand auto attack", pal.AutoAttacks.MHAuto().ActionID},
		{"Holy Strike", core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID}},
		{"Hammer of Wrath", core.ActionID{SpellID: paladin.HammerOfWrathRankMap.Highest().ID}},
	} {
		with, without := pal.GetSpell(spell.id), base.GetSpell(spell.id)
		if with == nil || without == nil {
			t.Fatalf("%s is not registered", spell.name)
		}
		got := with.PhysicalCritChance(pal.AttackTables[pal.CurrentTarget.UnitIndex]) -
			without.PhysicalCritChance(base.AttackTables[base.CurrentTarget.UnitIndex])
		if with.DefenseType == core.DefenseTypeMagic {
			got = with.SpellCritChance(pal.CurrentTarget) - without.SpellCritChance(base.CurrentTarget)
		}
		if math.Abs(got-0.06) > 1e-9 {
			t.Errorf("Conviction gives %s %.4f crit, want 0.06", spell.name, got)
		}
	}
}

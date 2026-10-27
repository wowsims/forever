package protection

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// In game Iron Creed's damage reduction comes from making a Holy Strike, the way Reckoning's extra
// attack comes from making a swing: a strike that misses, is dodged or is parried grants it too.
func TestIronCreedFromAnyHolyStrike(t *testing.T) {
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/protection/gear_sets", "default").GearSet,
		TalentsString: DefaultProtectionTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, DefaultOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	target := sim.Encounter.ActiveTargetUnits[0]
	if pally.Talents.IronCreed == 0 {
		t.Fatal("the default protection talents have no points in Iron Creed")
	}

	ironCreed := pally.GetAuraByID(core.ActionID{SpellID: 1311033})
	holyStrike := pally.GetSpell(core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID})
	if ironCreed == nil || holyStrike == nil {
		t.Fatal("Iron Creed's buff or Holy Strike is not registered")
	}
	pally.RighteousFuryAura.Activate(sim)

	avoided := func() int32 {
		metrics := holyStrike.SpellMetrics[target.UnitIndex]
		return metrics.Misses + metrics.Dodges + metrics.Parries
	}

	// Strikes until one is avoided, checking that every strike before it, and that one, grants the
	// buff. A proc's handler runs one spell batch window after the hit, so each strike is followed
	// by stepping past that window.
	for range 500 {
		ironCreed.Deactivate(sim)
		before := avoided()
		holyStrike.SkipCastAndApplyEffects(sim, target)
		for at := sim.CurrentTime; sim.CurrentTime <= at+core.SpellBatchWindow; {
			sim.Step()
		}
		if !ironCreed.IsActive() {
			t.Fatalf("a Holy Strike with Righteous Fury up did not grant Iron Creed (avoided: %v)", avoided() > before)
		}
		if avoided() > before {
			return
		}
	}
	t.Fatal("no Holy Strike out of 500 was avoided")
}

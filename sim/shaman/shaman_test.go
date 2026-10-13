package shaman_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/enhancement"
)

func init() {
	enhancement.RegisterEnhancementShaman()
	common.RegisterAllEffects()
}

// The Enhancement suite's talents, which take Elemental Devastation.
const enhancementTalents = "050533102-053030031005112251"

// Fire Nova's hits are a spell the cast triggers, and the row lacks Not a Proc, so Elemental
// Devastation, which cannot proc from procs, never hears their crits. It hears an Earth Shock's.
func TestFireNovaHitsAreProcs(t *testing.T) {
	player := &proto.Player{
		Name:          "Fire Nova",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassShaman,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: enhancementTalents,
		Spec:          &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
		Rotation:      &proto.APLRotation{},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	target := sham.CurrentTarget
	devastation := sham.GetAuraByID(core.ActionID{SpellID: 30165})
	novaHits := sham.GetSpell(core.ActionID{SpellID: sham.FireNova.ActionID.SpellID, Tag: 1})
	earthShock := sham.EarthShock[len(sham.EarthShock)-1]
	if devastation == nil || novaHits == nil {
		t.Fatalf("Elemental Devastation %v, Fire Nova's hits %v; want both registered", devastation, novaHits)
	}

	devastatedByCrit := func(cast *core.Spell, hits *core.Spell) bool {
		t.Helper()
		devastation.Deactivate(sim)
		ended := false
		casts := 0
		for casts < 1000 && !ended {
			crits := hits.SpellMetrics[target.UnitIndex].Crits
			cast.SkipCastAndApplyEffects(sim, target)
			casts++
			for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !ended; {
				ended = sim.Step()
			}
			if hits.SpellMetrics[target.UnitIndex].Crits > crits {
				return devastation.IsActive()
			}
		}
		t.Fatalf("%d casts of %v never crit", casts, cast.ActionID)
		return false
	}

	if devastatedByCrit(sham.FireNova, novaHits) {
		t.Error("a Fire Nova crit procced Elemental Devastation")
	}
	if !devastatedByCrit(earthShock, earthShock) {
		t.Error("an Earth Shock crit did not proc Elemental Devastation, so the check above proves nothing")
	}
}

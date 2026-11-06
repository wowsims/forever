//go:build with_db

package paladin_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
	"github.com/wowsims/forever/sim/paladin/retribution"
)

func init() {
	retribution.RegisterRetributionPaladin()
	common.RegisterAllEffects()
}

// The ladder's first Seal of Righteousness, 21084, deals its hits as 25741. In game the level 1
// starter seal 20154, which is not on the ladder, deals them as 25742: a log from levels 1 to 10
// shows 145 hits of 25742 while 20154 was up and 641 of 25741 while 21084 was.
func TestSealOfRighteousnessRankOneHitsAs25741(t *testing.T) {
	player := &proto.Player{
		Name:      "Seal of Righteousness",
		Race:      proto.Race_RaceDwarf,
		Class:     proto.Class_ClassPaladin,
		Equipment: core.GetGearSet("../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		// An empty rotation, so nothing swaps the seal out.
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()

	rank := paladin.SealOfRighteousnessRankMap.Rank(1)
	seal := pal.GetAuraByID(core.ActionID{SpellID: rank.ID})
	if seal == nil {
		t.Fatalf("seal %d is not registered", rank.ID)
	}
	seal.Activate(sim)
	// A seal hit lands a spell batch after the swing.
	for i := 0; i < 20; i++ {
		pal.AutoAttacks.MHAuto().SkipCastAndApplyEffects(sim, pal.CurrentTarget)
		for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
	}

	hitsBy := map[int32]int32{}
	for _, spell := range pal.Spellbook {
		if spell.Matches(paladin.SpellMaskSealOfRighteousnessProc) {
			metrics := spell.SpellMetrics[pal.CurrentTarget.UnitIndex]
			if landed := metrics.Hits + metrics.Crits; landed > 0 {
				hitsBy[spell.ActionID.SpellID] += landed
			}
		}
	}
	if len(hitsBy) != 1 || hitsBy[25741] == 0 {
		t.Errorf("seal %d hits as %v (spell id: hits), want spell 25741 alone", rank.ID, hitsBy)
	}
}

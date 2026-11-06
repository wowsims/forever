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

// The ladder's first Seal of Righteousness deals its hits as its own spell, not as the level 1
// starter seal's.
func TestSealOfRighteousnessRankOneHitsAsItsOwnSpell(t *testing.T) {
	const want = 25741 // 21084's damage spell, "Rank 1.5" in the game files
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
	if len(hitsBy) != 1 || hitsBy[want] == 0 {
		t.Errorf("seal %d hits as %v (spell id: hits), want spell %d alone", rank.ID, hitsBy, want)
	}
}

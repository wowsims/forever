package holy

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// Holy Light and Flash of Light make 0.25 threat a point healed in game, half of what other heals
// make, before the paladin's own threat modifiers.
func TestHolyLightAndFlashOfLightMakeQuarterHealingThreat(t *testing.T) {
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet,
		TalentsString: StandardTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, PlayerOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].GetCharacter()
	target := sim.Encounter.ActiveTargetUnits[0]

	for name, ranks := range map[string]spelldata.Ladder{
		"Holy Light":     paladin.HolyLightRankMap,
		"Flash of Light": paladin.FlashOfLightRankMap,
	} {
		heal := pally.GetSpell(core.ActionID{SpellID: ranks.Highest().ID})
		if heal == nil {
			t.Fatalf("%s is not registered", name)
		}

		// Down to 1 health, so the whole heal lands and none of it is overhealing.
		pally.RemoveHealth(sim, pally.CurrentHealth()-1)
		before := pally.CurrentHealth()
		heal.SkipCastAndApplyEffects(sim, &pally.Unit)
		healed := pally.CurrentHealth() - before
		if healed <= 0 {
			t.Fatalf("%s restored %v health", name, healed)
		}

		want := healed * 0.25 * pally.PseudoStats.ThreatMultiplier * pally.PseudoStats.SchoolThreatMultiplier[heal.SchoolIndex]
		if got := heal.SpellMetrics[target.UnitIndex].TotalThreat; !core.WithinToleranceFloat64(want, got, 0.001) {
			t.Errorf("%s restoring %v health made %v threat, want %v", name, healed, got, want)
		}
	}
}

package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// Seal of Light heals on a landed melee hit at the row's procs per minute on the main hand's speed.
func TestSealOfLightHealsAtTheRowsRate(t *testing.T) {
	player := &proto.Player{
		Name:          "Seal of Light",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "twist_of_light").GearSet,
		TalentsString: TwistOfLightTalents,
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
	pal.AutoAttacks.CancelAutoSwing(sim)

	rank := paladin.SealOfLightRankMap.Highest()
	seal := pal.GetAuraByID(core.ActionID{SpellID: rank.ID})
	if seal == nil {
		t.Fatal("Seal of Light is not registered")
	}
	seal.Activate(sim)
	strike := pal.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{OtherID: 1},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})

	// The heal lands a spell batch after the hit, and the sim stops stepping at its duration, so the
	// strikes end before it.
	hits, heals := 0, 0
	for sim.CurrentTime < sim.Duration-core.SpellBatchWindow*4 {
		pal.RemoveHealth(sim, pal.CurrentHealth()/2)
		before := pal.CurrentHealth()
		strike.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
		seal.Refresh(sim)
		for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
		hits++
		if pal.CurrentHealth() > before {
			heals++
		}
	}

	want := float64(spelldata.MustFind(rank.ID).RPPM) * pal.AutoAttacks.MH().SwingSpeed / 60
	got := float64(heals) / float64(hits)
	if tolerance := 5 * math.Sqrt(want*(1-want)/float64(hits)); math.Abs(got-want) > tolerance {
		t.Errorf("the seal healed on %d of %d hits (%.3f), want %.3f within %.3f", heals, hits, got, want, tolerance)
	}
}

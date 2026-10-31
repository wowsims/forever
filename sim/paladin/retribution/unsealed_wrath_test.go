package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// A paladin with Unsealed Wrath and no rotation or auto attacks, a swing that always lands, and a
// count of the Seal of Righteousness hits cast so far.
func unsealedWrathPaladin(t *testing.T) (sim *core.Simulation, pal *paladin.Paladin, swing func(), righteousnessHits func() int32) {
	t.Helper()
	player := &proto.Player{
		Name:          "Unsealed Wrath",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: UnsealedWrathTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim = core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal = sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pal.AutoAttacks.CancelAutoSwing(sim)
	if !pal.Talents.UnsealedWrath {
		t.Fatal("the talents do not take Unsealed Wrath")
	}

	righteousnessHits = func() int32 {
		var casts int32
		for _, spell := range pal.Spellbook {
			if spell.Matches(paladin.SpellMaskSealOfRighteousnessProc) {
				casts += spell.SpellMetrics[pal.CurrentTarget.UnitIndex].Casts
			}
		}
		return casts
	}

	strike := pal.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{OtherID: 1},
		SpellSchool:      core.SpellSchoolPhysical,
		ProcMask:         core.ProcMaskMeleeMHAuto,
		Flags:            core.SpellFlagMeleeMetrics,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})
	// Each swing steps the sim past the batch window the seal effects land in.
	swing = func() {
		strike.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
		for end := sim.CurrentTime + 2*core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
		}
	}
	return sim, pal, swing, righteousnessHits
}

func castSeal(t *testing.T, sim *core.Simulation, pal *paladin.Paladin, id int32) {
	t.Helper()
	spell := pal.GetSpell(core.ActionID{SpellID: id})
	if spell == nil {
		t.Fatalf("seal %d is not registered", id)
	}
	spell.SkipCastAndApplyEffects(sim, pal.CurrentTarget)
}

// Each swing under a seal stores one copy of it, up to the Echo's 3 stacks.
func TestUnsealedWrathStoresUpToThreeCopies(t *testing.T) {
	sim, pal, swing, _ := unsealedWrathPaladin(t)
	castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)

	echo := pal.GetAuraByID(core.ActionID{SpellID: 1311704})
	if echo == nil {
		t.Fatal("Echo of Righteousness is not registered")
	}
	for i, want := range []int32{1, 2, 3, 3} {
		swing()
		if got := echo.GetStacks(); got != want {
			t.Fatalf("after swing %d Echo of Righteousness has %d stacks, want %d", i+1, got, want)
		}
	}
}

// A hit that lands under another seal spends every stored copy, each one a Seal of Righteousness hit,
// and stores a copy of the new seal.
func TestUnsealedWrathSpendsTheStoredCopiesUnderAnotherSeal(t *testing.T) {
	sim, pal, swing, righteousnessHits := unsealedWrathPaladin(t)
	castSeal(t, sim, pal, paladin.SealOfRighteousnessRankMap.Highest().ID)
	for range 3 {
		swing()
	}

	castSeal(t, sim, pal, paladin.SealOfFuryRankMap.Highest().ID)
	before := righteousnessHits()
	swing()

	if got := righteousnessHits() - before; got != 3 {
		t.Errorf("the hit under Seal of Fury replayed %d Seal of Righteousness hits, want 3", got)
	}
	if pal.GetAuraByID(core.ActionID{SpellID: 1311704}).IsActive() {
		t.Error("Echo of Righteousness stays up after its copies are spent")
	}
	if got := pal.GetAuraByID(core.ActionID{SpellID: 1311701}).GetStacks(); got != 1 {
		t.Errorf("Echo of Fury has %d stacks, want 1", got)
	}
}

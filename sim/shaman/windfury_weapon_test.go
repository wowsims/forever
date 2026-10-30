//go:build with_db

package shaman_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/shaman"
)

// Windfury Weapon's strikes are white in name only: Flametongue Totem, which hears auto attacks
// alone, procs off every one that lands, yet they roll the special attack table and never glance.
func TestWindfuryWeaponStrikesProcFlametongueTotemAndNeverGlance(t *testing.T) {
	player := windfuryShaman(&proto.APLRotation{})
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(player, &proto.PartyBuffs{},
			&proto.RaidBuffs{FlametongueTotem: true}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: 20, RandomSeed: 1},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %s", result.Error.Message)
	}

	// A proc logs no casts, so each count is the outcomes it logged.
	strikeTag := spelldata.MustFind(16362).Refs()[0].ID
	flametongueTotem := spelldata.MustFind(16387).ID
	var mainHand, strikes, flametongue proto.TargetedActionMetrics
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		var into *proto.TargetedActionMetrics
		switch {
		case action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack && action.Id.Tag == int32(proto.ItemSlot_ItemSlotMainHand):
			into = &mainHand
		case action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack && action.Id.Tag == strikeTag:
			into = &strikes
		case action.Id.GetSpellId() == flametongueTotem && action.Id.Tag == 1:
			into = &flametongue
		default:
			continue
		}
		for _, target := range action.Targets {
			into.Hits += target.Hits + target.Blocks
			into.Crits += target.Crits + target.BlockedCrits
			into.Glances += target.Glances + target.GlanceBlocks
			into.Misses += target.Misses + target.Dodges + target.Parries
		}
	}
	landed := func(m *proto.TargetedActionMetrics) int32 { return m.Hits + m.Crits + m.Glances }

	if landed(&strikes) == 0 {
		t.Fatal("no Windfury Weapon strike landed over the run")
	}
	if strikes.Glances != 0 {
		t.Errorf("Windfury Weapon's strikes glanced %d times, want none", strikes.Glances)
	}
	if mainHand.Glances == 0 {
		t.Error("the main hand's swings never glanced, so the run cannot tell the strikes from them")
	}
	if got, want := landed(&flametongue)+flametongue.Misses, landed(&mainHand)+landed(&strikes); got != want {
		t.Errorf("Flametongue Totem procced %d times, want one for each of the %d swings and %d Windfury Weapon strikes that landed",
			got, landed(&mainHand), landed(&strikes))
	}
}

// Windfury Weapon's strikes spend no Flurry charge, though they are white for every proc: a main-hand
// swing spends one.
func TestWindfuryWeaponStrikesSpendNoFlurryCharge(t *testing.T) {
	player := windfuryShaman(&proto.APLRotation{})
	player.TalentsString = enhancementTalents
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	target := sham.CurrentTarget
	flurry := sham.GetAuraByID(core.ActionID{SpellID: spelldata.MustFind(16257).ID})
	strike := sham.GetSpell(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: spelldata.MustFind(16362).Refs()[0].ID})
	if flurry == nil || strike == nil {
		t.Fatalf("Flurry %v, Windfury Weapon's strike %v; want both registered", flurry, strike)
	}
	flurry.Activate(sim)
	flurry.SetStacks(sim, flurry.MaxStacks)

	for range 50 {
		strike.SkipCastAndApplyEffects(sim, target)
		if got := flurry.GetStacks(); got != flurry.MaxStacks {
			t.Fatalf("a Windfury Weapon strike left Flurry at %d charges, want %d", got, flurry.MaxStacks)
		}
	}

	swing := sham.AutoAttacks.MHAuto()
	for range 50 {
		swing.SkipCastAndApplyEffects(sim, target)
		if flurry.GetStacks() < flurry.MaxStacks {
			return
		}
	}
	t.Error("50 main-hand swings never spent a Flurry charge, so the check above proves nothing")
}

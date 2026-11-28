//go:build with_db

package shaman_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/shaman"
)

// The spells each Windfury Weapon hit and each Windfury Totem hit report under.
var (
	windfuryWeaponStrike = core.ActionID{SpellID: spelldata.MustFind(439440).ID}
	windfuryTotemStrike  = core.ActionID{SpellID: spelldata.MustFind(10610).ID}
)

// Windfury Weapon hits are main-hand abilities: Flametongue Totem never procs off a Windfury hit,
// and the Windfury hits never glance.
func TestWindfuryWeaponStrikesNeverProcFlametongueTotemOrGlance(t *testing.T) {
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
	flametongueTotem := spelldata.MustFind(16387).ID
	var mainHand, strikes, flametongue proto.TargetedActionMetrics
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		var into *proto.TargetedActionMetrics
		switch {
		case action.Id.GetOtherId() == proto.OtherAction_OtherActionAttack && action.Id.Tag == int32(proto.ItemSlot_ItemSlotMainHand):
			into = &mainHand
		case action.Id.GetSpellId() == windfuryWeaponStrike.SpellID:
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
		t.Fatal("no Windfury Weapon hit landed over the run")
	}
	if strikes.Glances != 0 {
		t.Errorf("Windfury Weapon hits glanced %d times, want none", strikes.Glances)
	}
	if mainHand.Glances == 0 {
		t.Error("the main hand's swings never glanced, so the run cannot tell the Windfury hits from them")
	}
	procs := landed(&flametongue) + flametongue.Misses
	if procs == 0 {
		t.Fatal("Flametongue Totem never procced over the run")
	}
	if want := landed(&mainHand); procs != want {
		t.Errorf("Flametongue Totem procced %d times, want one for each of the %d swings that landed and none for the %d Windfury Weapon hits that landed",
			procs, want, landed(&strikes))
	}
}

// A main-hand swing uses a Flurry charge whatever its outcome, and so does a Windfury Weapon hit.
func TestFlurrySpendsOnSwingsAndWindfuryStrikes(t *testing.T) {
	player := windfuryShaman(&proto.APLRotation{})
	player.TalentsString = enhancementTalents
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{WindfuryTotem: true}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	target := sham.CurrentTarget
	flurry := sham.GetAuraByID(core.ActionID{SpellID: spelldata.MustFind(16257).ID})
	trigger := sham.GetAura("Flurry - Trigger")
	swing := sham.AutoAttacks.MHAuto()
	strike := sham.GetSpell(windfuryWeaponStrike)
	totemStrike := sham.GetSpell(windfuryTotemStrike)
	if flurry == nil || trigger == nil || strike == nil || totemStrike == nil || sham.Stormstrike == nil {
		t.Fatalf("Flurry %v, its trigger %v, Windfury Weapon hit %v, Windfury Totem hit %v, Stormstrike %v; want all registered",
			flurry, trigger, strike, totemStrike, sham.Stormstrike)
	}
	hit := func(spell *core.Spell, outcome core.HitOutcome) {
		trigger.OnSpellHitDealt(trigger, sim, spell, &core.SpellResult{Target: target, Outcome: outcome, Damage: 100})
	}

	// Each case starts a second after the last, past the cooldown between charges used, and hits a
	// moment after the full charges are set.
	expect := func(name string, spell *core.Spell, outcome core.HitOutcome, spends bool) {
		t.Helper()
		sim.CurrentTime += time.Second
		flurry.Activate(sim)
		flurry.SetStacks(sim, flurry.MaxStacks)
		sim.CurrentTime += time.Millisecond
		hit(spell, outcome)
		if spent := flurry.GetStacks() < flurry.MaxStacks; spent != spends {
			t.Errorf("%s: spent a charge %v, want %v", name, spent, spends)
		}
	}
	expect("a dodged swing", swing, core.OutcomeDodge, true)
	expect("a Windfury Weapon hit", strike, core.OutcomeHit, true)
	expect("a Windfury Totem hit", totemStrike, core.OutcomeHit, false)
	expect("a Stormstrike", sham.Stormstrike, core.OutcomeHit, false)

	sim.CurrentTime += time.Second
	hit(swing, core.OutcomeCrit)
	hit(strike, core.OutcomeHit)
	if got, want := flurry.GetStacks(), flurry.MaxStacks-1; got != want {
		t.Errorf("a Windfury Weapon hit in the instant of a swing's crit left %d charges, want %d", got, want)
	}

	sim.CurrentTime += time.Second
	hit(strike, core.OutcomeCrit)
	hit(strike, core.OutcomeHit)
	if got := flurry.GetStacks(); got != flurry.MaxStacks {
		t.Errorf("a Windfury Weapon hit in the instant of a Windfury Weapon hit's crit left %d charges, want %d", got, flurry.MaxStacks)
	}
}

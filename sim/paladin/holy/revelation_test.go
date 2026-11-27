package holy

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/paladin"
)

// Enchant Weapon - Revelation 8217 gives a paladin buff 1323419, whose crit modifier and charge name
// Holy Light by its class flags (family 10) and not Lay on Hands. Holy Light with the buff up takes
// the crit and spends the charge; Lay on Hands leaves the charge up.
// Steps until the clock is past end, with a wake of its own just after it.
func stepPast(sim *core.Simulation, end time.Duration) {
	sim.AddPendingAction(&core.PendingAction{NextActionAt: end + 1, OnAction: func(*core.Simulation) {}})
	for sim.CurrentTime <= end && !sim.Step() {
	}
}

func TestRevelationReachesHolyLight(t *testing.T) {
	gear := core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet
	gear.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 8217
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     gear,
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

	buff := pally.GetAuraByID(core.ActionID{SpellID: 1323419})
	if buff == nil {
		t.Fatal("no Revelation buff 1323419 on the paladin")
	}
	holyLight := pally.GetSpell(core.ActionID{SpellID: paladin.HolyLightRankMap.Highest().ID})
	layOnHands := pally.GetSpell(core.ActionID{SpellID: paladin.LayOnHandsRankMap.Highest().ID})
	if holyLight == nil || layOnHands == nil {
		t.Fatal("Holy Light or Lay on Hands is not registered")
	}

	baseCrit := holyLight.HealingCritChance()
	buff.Activate(sim)
	buff.SetStacks(sim, buff.MaxStacks)
	if got := holyLight.HealingCritChance(); got <= baseCrit {
		t.Errorf("Holy Light's crit chance is %v with Revelation up and %v without", got, baseCrit)
	}

	// The charge goes as the cast completes, and a proc's handler runs one spell batch window after the
	// heal, so each cast is stepped through to its end and past that window.
	castAndStep := func(spell *core.Spell) {
		for !pally.GCD.IsReady(sim) {
			sim.Step()
		}
		if !spell.Cast(sim, &pally.Unit) {
			t.Fatalf("%s could not be cast", spell.ActionID)
		}
		stepPast(sim, sim.CurrentTime+spell.CurCast.CastTime+core.SpellBatchWindow)
	}

	castAndStep(holyLight)
	if buff.IsActive() {
		t.Error("Holy Light did not spend Revelation's charge")
	}

	// Lay on Hands takes all the paladin's mana, so it goes last.
	buff.Activate(sim)
	buff.SetStacks(sim, buff.MaxStacks)
	castAndStep(layOnHands)
	if !buff.IsActive() {
		t.Error("Lay on Hands, which the buff's mask does not name, spent its charge")
	}
}

// Judgement of Light's heal rolls nothing for Revelation on the paladin it heals, and Seal of Light's
// heal rolls it (EV-0149).
func TestRevelationHearsSealOfLightsHealAndNotJudgementOfLights(t *testing.T) {
	gear := core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", "default").GearSet
	gear.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 8217
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassPaladin,
		Race:          proto.Race_RaceHuman,
		Equipment:     gear,
		TalentsString: StandardTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}, PlayerOptions)
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{JudgementOfLight: true}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pally := sim.Raid.Parties[0].Players[0].GetCharacter()

	buff := pally.GetAuraByID(core.ActionID{SpellID: 1323419})
	judgementHeal := pally.GetSpell(core.ActionID{SpellID: buffs.JudgementOfLightMaxRank.HealID})
	sealHeal := pally.GetSpell(core.ActionID{SpellID: paladin.SealOfLightRankMap.Highest().Refs()[0].ID})
	if buff == nil || judgementHeal == nil || sealHeal == nil {
		t.Fatal("the Revelation buff, Judgement of Light's heal or Seal of Light's heal is not registered")
	}

	// A proc's handler runs one spell batch window after the heal.
	procs := func(heal *core.Spell) bool {
		for range 300 {
			heal.Cast(sim, &pally.Unit)
			stepPast(sim, sim.CurrentTime+core.SpellBatchWindow)
			if buff.IsActive() {
				return true
			}
		}
		return false
	}
	if procs(judgementHeal) {
		t.Error("Judgement of Light's heal procced Revelation")
	}
	if !procs(sealHeal) {
		t.Error("Seal of Light's heal procced no Revelation in 300 heals")
	}
}

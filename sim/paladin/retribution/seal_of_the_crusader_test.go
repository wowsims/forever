package retribution

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/paladin"
)

// Seal of the Crusader moves the weapon speed a rate on the weapon's speed reads by the factor its
// swings speed up, and gives it back when the seal drops.
func TestSealOfTheCrusaderMovesTheWeaponSpeed(t *testing.T) {
	player := &proto.Player{
		Name:          "Seal of the Crusader",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: UnsealedWrathTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "default").Rotation,
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()

	seal := pal.GetSpell(core.ActionID{SpellID: paladin.SealOfTheCrusaderRankMap.Highest().ID})
	if seal == nil {
		t.Fatal("Seal of the Crusader is not registered")
	}
	speed := spelldata.MustFind(seal.ActionID.SpellID).EffectN(2)
	if speed.Aura != dbcenums.A_MOD_ATTACKSPEED {
		t.Fatalf("effect 2 of the seal is %v, want A_MOD_ATTACKSPEED", speed.Aura)
	}
	want := 1 + speed.Percent()

	meleeBefore := pal.PseudoStats.MeleeSpeedMultiplier
	if !seal.Cast(sim, pal.CurrentTarget) {
		t.Fatal("Seal of the Crusader did not cast")
	}
	if got := pal.PseudoStats.WeaponSpeedMultiplier; math.Abs(got-want) > 1e-9 {
		t.Errorf("under the seal the weapon speed multiplier is %v, want the row's %v", got, want)
	}
	if got := pal.PseudoStats.MeleeSpeedMultiplier / meleeBefore; math.Abs(got-want) > 1e-9 {
		t.Errorf("under the seal the swings speed up by %v, want the row's %v", got, want)
	}

	pal.GetAuraByID(seal.ActionID).Deactivate(sim)
	if got := pal.PseudoStats.WeaponSpeedMultiplier; math.Abs(got-1) > 1e-9 {
		t.Errorf("after the seal drops the weapon speed multiplier is %v, want 1", got)
	}
}

var noSacredArbiterTalents = "050250002-501-05233011200301351" // The Unsealed Wrath build without Sacred Arbiter, whose own trigger lets Holy Strike refresh the judgement debuffs.

// A landed white swing refreshes Judgement of the Crusader on its target, and so does Windfury
// Totem's strike. Holy Strike, Seal of Command's hit and Judgement of Command leave it running
// down (EV-0117).
func TestJudgementDebuffRefreshesOnWhiteSwingsAndWindfuryTotemsStrike(t *testing.T) {
	player := &proto.Player{
		Name:          "Judgement Refresh",
		Race:          proto.Race_RaceHuman,
		Class:         proto.Class_ClassPaladin,
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "default").GearSet,
		TalentsString: noSacredArbiterTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{WindfuryTotem: true}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	pal := sim.Raid.Parties[0].Players[0].(paladin.PaladinAgent).GetPaladin()
	pal.AutoAttacks.CancelAutoSwing(sim)
	if pal.Talents.SacredArbiter || !pal.Talents.SealOfCommand {
		t.Fatal("the talents take Sacred Arbiter or lack Seal of Command")
	}
	target := pal.CurrentTarget
	byMask := func(mask int64) *core.Spell {
		for _, spell := range pal.Spellbook {
			if spell.Matches(mask) && spell.ActionID.Tag == 0 {
				return spell
			}
		}
		return nil
	}

	castSeal(t, sim, pal, paladin.SealOfTheCrusaderRankMap.Highest().ID)
	judgement := byMask(paladin.SpellMaskJudgementOfTheCrusader)
	if judgement == nil {
		t.Fatal("Judgement of the Crusader is not registered")
	}
	judgement.SkipCastAndApplyEffects(sim, target)
	var debuff *core.Aura
	for _, auras := range pal.JudgementAuras {
		if aura := auras.Get(target); aura.IsActive() {
			debuff = aura
		}
	}
	if debuff == nil {
		t.Fatal("the judgement left no debuff on the target")
	}

	for _, hit := range []struct {
		name      string
		spell     *core.Spell
		refreshes bool
	}{
		{"a white swing", pal.AutoAttacks.MHAuto(), true},
		{"Windfury Totem's strike", pal.GetSpell(core.ActionID{SpellID: windfuryTotemStrike}), true},
		{"Holy Strike", pal.GetSpell(core.ActionID{SpellID: paladin.HolyStrikeRankMap.Highest().ID}), false},
		{"Seal of Command's hit", byMask(paladin.SpellMaskSealOfCommandProc), false},
		{"Judgement of Command", byMask(paladin.SpellMaskJudgementOfCommand), false},
	} {
		if hit.spell == nil {
			t.Fatalf("%s is not registered", hit.name)
		}
		passUnsealedWrathCooldown(sim)
		before := debuff.ExpiresAt()
		pal.OnSpellHitDealt(sim, hit.spell, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
		if got := debuff.ExpiresAt() > before; got != hit.refreshes {
			t.Errorf("%s landing refreshed the judgement debuff: %v, want %v", hit.name, got, hit.refreshes)
		}
	}
}

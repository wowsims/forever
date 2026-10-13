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
		Equipment:     core.GetGearSet("../../../ui/specs/paladin/retribution/gear_sets", "twist_of_light").GearSet,
		TalentsString: TwistOfLightTalents,
		Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
			Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}},
		}},
		Rotation: core.GetAplRotation("../../../ui/specs/paladin/retribution/apls", "twist_of_light").Rotation,
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

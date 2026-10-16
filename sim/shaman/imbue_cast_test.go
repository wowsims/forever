//go:build with_db

package shaman_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/shaman"
)

// Arcanite Reaper, a two-hand axe.
const imbueTestMainHand = 12784

// A shaman with Arcanite Reaper in the main hand and Windfury Weapon on it from the options.
func windfuryShaman(rotation *proto.APLRotation) *proto.Player {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = imbueTestMainHand
	return &proto.Player{
		Name:      "Imbue",
		Race:      proto.Race_RaceOrc,
		Class:     proto.Class_ClassShaman,
		Equipment: &proto.EquipmentSpec{Items: items},
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: &proto.EnhancementShaman_Options{
			ClassOptions: &proto.ShamanOptions{ImbueMh: proto.ShamanImbue_WindfuryWeapon},
		}}},
		Rotation: rotation,
	}
}

// Casting an imbue replaces the one the options put on the main hand, with the cast's own aura,
// mana and GCD, and the next iteration starts from the options again.
func TestImbueCastReplacesTheMainHandImbue(t *testing.T) {
	player := windfuryShaman(&proto.APLRotation{})
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()

	sham := sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
	windfury := spelldata.MustFind(16362)
	rockbiter := spelldata.MustFind(16316)
	windfuryAura := sham.GetAuraByID(core.ActionID{SpellID: windfury.ID})
	rockbiterAura := sham.GetAuraByID(core.ActionID{SpellID: rockbiter.ID})
	rockbiterCast := sham.GetSpell(core.ActionID{SpellID: rockbiter.ID})
	if windfuryAura == nil || rockbiterAura == nil || rockbiterCast == nil {
		t.Fatalf("Windfury Weapon's aura %v, Rockbiter Weapon's aura %v and cast %v; want all registered", windfuryAura, rockbiterAura, rockbiterCast)
	}
	if !windfuryAura.IsActive() || rockbiterAura.IsActive() {
		t.Fatalf("before the cast Windfury is up %t and Rockbiter %t, want the options' Windfury alone", windfuryAura.IsActive(), rockbiterAura.IsActive())
	}

	attackPower := sham.GetStat(stats.AttackPower)
	mana := sham.CurrentMana()
	if !rockbiterCast.Cast(sim, sham.CurrentTarget) {
		t.Fatal("Rockbiter Weapon did not cast")
	}
	if windfuryAura.IsActive() || !rockbiterAura.IsActive() {
		t.Errorf("after the cast Windfury is up %t and Rockbiter %t, want Rockbiter alone", windfuryAura.IsActive(), rockbiterAura.IsActive())
	}
	if got, want := sham.GetStat(stats.AttackPower)-attackPower, spelldata.MustFind(16313).EffectN(1).Average(); got != want {
		t.Errorf("Rockbiter Weapon added %.1f attack power, want its passive's %.1f", got, want)
	}
	if got, want := mana-sham.CurrentMana(), rockbiterCast.Cost.GetCurrentCost(); got != want || got == 0 {
		t.Errorf("the cast spent %.1f mana, want its cost of %.1f", got, want)
	}
	if sham.GCD.IsReady(sim) {
		t.Error("the cast did not start the GCD")
	}

	sim.Cleanup()
	sim.Reset()
	if !windfuryAura.IsActive() || rockbiterAura.IsActive() {
		t.Errorf("after a reset Windfury is up %t and Rockbiter %t, want the options' Windfury again", windfuryAura.IsActive(), rockbiterAura.IsActive())
	}
}

// A rotation that casts Rockbiter Weapon while its aura is down casts it once each iteration: the
// aura is the cast's buff, so the rotation sees the imbue it put on the main hand.
func TestRotationCastsAnImbueOnce(t *testing.T) {
	rockbiter := spelldata.MustFind(16316).ID
	rockbiterID := &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: rockbiter}}
	player := windfuryShaman(&proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: []*proto.APLListItem{{Action: &proto.APLAction{
		Condition: &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: &proto.APLValue{
			Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{AuraId: rockbiterID}},
		}}}},
		Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: rockbiterID}},
	}}}})
	const iterations = 3
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{Iterations: iterations, RandomSeed: 1},
	})
	if result.Error != nil {
		t.Fatalf("the sim failed: %s", result.Error.Message)
	}

	var casts int32
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() == rockbiter {
			for _, target := range action.Targets {
				casts += target.Casts
			}
		}
	}
	if casts != iterations {
		t.Errorf("Rockbiter Weapon was cast %d times over %d iterations, want once each", casts, iterations)
	}
}

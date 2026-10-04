//go:build with_db

package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// A druid in Cat Form swinging the paw the feral constructors build, without the form spells and
// talents the class has not ported yet.
type pawDruid struct {
	*Druid
}

func (paw *pawDruid) Initialize()   {}
func (paw *pawDruid) ApplyTalents() {}

func init() {
	common.RegisterAllEffects()
	core.RegisterAgentFactory(
		proto.Player_FeralCatDruid{},
		proto.Spec_SpecFeralCatDruid,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			paw := &pawDruid{Druid: New(character, Cat, SelfBuffs{}, options.TalentsString)}
			paw.EnableAutoAttacks(paw, core.AutoAttackOptions{MainHand: paw.GetCatWeapon(), AutoSwingMelee: true})
			return paw
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_FeralCatDruid)
		},
	)
}

func catWithMainHand(t *testing.T, mainHand int32, consumes *proto.ConsumesSpec, partyBuffs *proto.PartyBuffs) *core.Character {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = mainHand
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: consumes,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: partyBuffs,
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Dense Weightstone 12643 puts enchant 1703's 8 weapon damage on Monk's Staff 866 (60-90, 2.4 s).
// The cat's paw swings every second for the staff's damage per second, the stone's 8 included once:
// (60 + 8) / 2.4 to (90 + 8) / 2.4.
func TestCatPawTakesTheWeaponstoneOnce(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.PartyBuffs{})

	paw := character.AutoAttacks.MH()
	if wantMin, wantMax := 68/2.4, 98/2.4; !approxEqual(paw.BaseDamageMin, wantMin) || !approxEqual(paw.BaseDamageMax, wantMax) {
		t.Errorf("the paw deals %.3f-%.3f, want %.3f-%.3f", paw.BaseDamageMin, paw.BaseDamageMax, wantMin, wantMax)
	}
}

// A form swap rebuilds the paw from the main hand, as the druid's item swap callback does; the rebuilt
// paw keeps the stone's 8, once.
func TestCatPawKeepsTheWeaponstoneWhenRebuilt(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.PartyBuffs{})
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	for range 2 {
		character.AutoAttacks.SetMH(cat.GetCatWeapon())
	}
	paw := character.AutoAttacks.MH()
	if wantMin, wantMax := 68/2.4, 98/2.4; !approxEqual(paw.BaseDamageMin, wantMin) || !approxEqual(paw.BaseDamageMax, wantMax) {
		t.Errorf("the rebuilt paw deals %.3f-%.3f, want %.3f-%.3f", paw.BaseDamageMin, paw.BaseDamageMax, wantMin, wantMax)
	}
}

// Dire Bear Form swings every 2.5 seconds for the staff's damage per second, the stone's 8 included
// once, through shifts to the paw and back: (60 + 8) / 2.4 * 2.5 to (90 + 8) / 2.4 * 2.5.
func TestBearWeaponTakesTheWeaponstoneOnce(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.PartyBuffs{})
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	for range 2 {
		character.AutoAttacks.SetMH(cat.GetBearWeapon())
		character.AutoAttacks.SetMH(cat.GetCatWeapon())
	}
	character.AutoAttacks.SetMH(cat.GetBearWeapon())
	bear := character.AutoAttacks.MH()
	if wantMin, wantMax := 68/2.4*2.5, 98/2.4*2.5; !approxEqual(bear.BaseDamageMin, wantMin) || !approxEqual(bear.BaseDamageMax, wantMax) || bear.SwingSpeed != 2.5 {
		t.Errorf("the bear weapon deals %.3f-%.3f at %.1f s, want %.3f-%.3f at 2.5 s",
			bear.BaseDamageMin, bear.BaseDamageMax, bear.SwingSpeed, wantMin, wantMax)
	}
}

// Windfury Totem's party aura reaches the cat like any other melee: its paw hears the Windfury
// procs and keeps the stone: (60 + 8) / 2.4.
func TestCatHearsWindfuryTotemAndKeepsItsMainHandImbue(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.PartyBuffs{WindfuryTotem: true})

	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant; got != 1703 {
		t.Errorf("the main hand carries temporary enchant %d, want Dense Weightstone's 1703", got)
	}
	if paw := character.AutoAttacks.MH(); !approxEqual(paw.BaseDamageMin, 68/2.4) {
		t.Errorf("the paw deals %.3f at least, want %.3f", paw.BaseDamageMin, 68/2.4)
	}
	if character.GetAura("Windfury Totem Trigger") == nil {
		t.Error("the cat registers no Windfury Totem proc")
	}
}

// Flametongue Totem's party aura reaches the cat the same way: its paw strikes totem fire and keeps
// the stone.
func TestCatHearsFlametongueTotemAndKeepsItsMainHandImbue(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.PartyBuffs{FlametongueTotem: true})

	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant; got != 1703 {
		t.Errorf("the main hand carries temporary enchant %d, want Dense Weightstone's 1703", got)
	}
	if character.GetAura("Flametongue Totem Trigger") == nil {
		t.Error("the cat registers no Flametongue Totem proc")
	}
	if character.GetSpell(core.ActionID{SpellID: 16389, Tag: -1}) == nil {
		t.Error("the cat registers no Flametongue Totem strike")
	}
}

func approxEqual(a, b float64) bool {
	return a-b < 1e-9 && b-a < 1e-9
}

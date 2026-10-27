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
// talents.
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

func catWithMainHand(t *testing.T, mainHand int32, consumes *proto.ConsumesSpec, raidBuffs *proto.RaidBuffs) *core.Character {
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
			Buffs: &proto.PartyBuffs{},
		}}, Buffs: raidBuffs},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Dense Weightstone 12643 puts enchant 1703's 8 weapon damage on Monk's Staff 866 (60-90, 2.4 s).
// The cat's paw uses the staff's table DPS and the form's variance; the stone adds 8 once.
func TestCatPawTakesTheWeaponstoneOnce(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{})

	paw := character.AutoAttacks.MH()
	if wantMin, wantMax := 60/2.4+8, 38.0+8; !approxEqual(paw.BaseDamageMin, wantMin) || !approxEqual(paw.BaseDamageMax, wantMax) {
		t.Errorf("the paw deals %.3f-%.3f, want %.3f-%.3f", paw.BaseDamageMin, paw.BaseDamageMax, wantMin, wantMax)
	}
}

// A form swap rebuilds the paw from the main hand, as the druid's item swap callback does; the rebuilt
// paw keeps the stone's 8, once.
func TestCatPawKeepsTheWeaponstoneWhenRebuilt(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{})
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	for range 2 {
		character.AutoAttacks.SetMH(cat.GetCatWeapon())
	}
	paw := character.AutoAttacks.MH()
	if wantMin, wantMax := 60/2.4+8, 38.0+8; !approxEqual(paw.BaseDamageMin, wantMin) || !approxEqual(paw.BaseDamageMax, wantMax) {
		t.Errorf("the rebuilt paw deals %.3f-%.3f, want %.3f-%.3f", paw.BaseDamageMin, paw.BaseDamageMax, wantMin, wantMax)
	}
}

// Bear Form uses the staff's table DPS and the form's variance at 2.5 seconds.
// The stone adds 8 once, through shifts to the paw and back.
func TestBearWeaponTakesTheWeaponstoneOnce(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{})
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	for range 2 {
		character.AutoAttacks.SetMH(cat.GetBearWeapon())
		character.AutoAttacks.SetMH(cat.GetCatWeapon())
	}
	character.AutoAttacks.SetMH(cat.GetBearWeapon())
	bear := character.AutoAttacks.MH()
	if wantMin, wantMax := 62.0+8, 94.0+8; !approxEqual(bear.BaseDamageMin, wantMin) || !approxEqual(bear.BaseDamageMax, wantMax) || bear.SwingSpeed != 2.5 {
		t.Errorf("the bear weapon deals %.3f-%.3f at %.1f s, want %.3f-%.3f at 2.5 s",
			bear.BaseDamageMin, bear.BaseDamageMax, bear.SwingSpeed, wantMin, wantMax)
	}
}

// Windfury Totem's party aura reaches the cat like any other melee: its paw hears the Windfury
// procs and keeps the stone: 60 / 2.4 + 8.
func TestCatHearsWindfuryTotemAndKeepsItsMainHandImbue(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{WindfuryTotem: true})

	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant; got != 1703 {
		t.Errorf("the main hand carries temporary enchant %d, want Dense Weightstone's 1703", got)
	}
	if paw := character.AutoAttacks.MH(); !approxEqual(paw.BaseDamageMin, 60/2.4+8) {
		t.Errorf("the paw deals %.3f at least, want %.3f", paw.BaseDamageMin, 60/2.4+8)
	}
	if character.GetAura("Windfury Totem Trigger") == nil {
		t.Error("the cat registers no Windfury Totem proc")
	}
}

// Flametongue Totem's party aura reaches the cat the same way: its paw strikes totem fire and keeps
// the stone.
func TestCatHearsFlametongueTotemAndKeepsItsMainHandImbue(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{FlametongueTotem: true})

	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant; got != 1703 {
		t.Errorf("the main hand carries temporary enchant %d, want Dense Weightstone's 1703", got)
	}
	if character.GetAura("Flametongue Totem Trigger") == nil {
		t.Error("the cat registers no Flametongue Totem proc")
	}
	if character.GetSpell(core.ActionID{SpellID: 16387, Tag: 1}) == nil {
		t.Error("the cat registers no Flametongue Totem strike")
	}
}

func approxEqual(a, b float64) bool {
	return a-b < 1e-9 && b-a < 1e-9
}

func TestFeralEnchantsAndStonesAreIndependentOfWeaponSpeed(t *testing.T) {
	for _, itemID := range []int32{866, 19323} {
		character := catWithMainHand(t, itemID, &proto.ConsumesSpec{MhImbueId: 12643}, &proto.RaidBuffs{})
		druid := character.Env.Raid.Parties[0].Players[0].(*pawDruid)
		item := character.GetMHWeapon()
		stone := item.TempEnchantWeaponDamage
		item.TempEnchantWeaponDamage = 0
		plain := []core.Weapon{druid.GetCatWeapon(), druid.GetBearWeapon()}
		item.TempEnchantWeaponDamage = stone
		item.Enchant = *core.GetEnchantByEffectID(963)
		bonus := item.Enchant.WeaponDamage + item.TempEnchantWeaponDamage
		if bonus == 0 {
			t.Fatal("weapon has no flat damage to test")
		}
		for i, paw := range []core.Weapon{druid.GetCatWeapon(), druid.GetBearWeapon()} {
			baseMin, baseMax := plain[i].BaseDamageMin, plain[i].BaseDamageMax
			if !approxEqual(paw.BaseDamageMin, baseMin+bonus) || !approxEqual(paw.BaseDamageMax, baseMax+bonus) {
				t.Errorf("item %d at %g s: paw %g-%g, want %g-%g", itemID, paw.SwingSpeed, paw.BaseDamageMin, paw.BaseDamageMax, baseMin+bonus, baseMax+bonus)
			}
		}
		weapon := druid.WeaponFromMainHand()
		if !approxEqual(weapon.BaseDamageMin, item.WeaponDamageMin+bonus) || !approxEqual(weapon.BaseDamageMax, item.WeaponDamageMax+bonus) {
			t.Errorf("item %d: caster weapon changed its flat damage", itemID)
		}
	}
}

// Base paw endpoints recovered from in-game measurements, before AP and flat weapon damage.
func TestFeralPawDamageEndpoints(t *testing.T) {
	for _, test := range []struct {
		itemID                           int32
		catMin, catMax, bearMin, bearMax float64
	}{
		{0, 1, 1, 1, 1},
		{5279, 8, 12, 20, 30}, // Harpy Skinner
	} {
		character := catWithMainHand(t, test.itemID, &proto.ConsumesSpec{}, &proto.RaidBuffs{})
		druid := character.Env.Raid.Parties[0].Players[0].(*pawDruid)
		cat, bear := druid.GetCatWeapon(), druid.GetBearWeapon()
		if cat.BaseDamageMin != test.catMin || cat.BaseDamageMax != test.catMax ||
			bear.BaseDamageMin != test.bearMin || bear.BaseDamageMax != test.bearMax {
			t.Errorf("item %d: cat %g-%g and bear %g-%g, want %g-%g and %g-%g",
				test.itemID, cat.BaseDamageMin, cat.BaseDamageMax, bear.BaseDamageMin, bear.BaseDamageMax,
				test.catMin, test.catMax, test.bearMin, test.bearMax)
		}
	}
}

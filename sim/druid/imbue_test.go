//go:build with_db

package druid

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
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

func catWithMainHand(t *testing.T, mainHand int32, consumes *proto.ConsumesSpec, raidBuffs *proto.RaidBuffs) (*core.Character, *core.Simulation) {
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
	return sim.Raid.Parties[0].Players[0].GetCharacter(), sim
}

// Dense Weightstone 12643 puts enchant 1703's 8 weapon damage on Monk's Staff 866 (60-90, 2.4 s).
// The cat's paw swings its own 43-66 every second, the stone's 8 added once: 51-74.
func TestCatPawTakesTheWeaponstoneOnce(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character, _ := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{})

	paw := character.AutoAttacks.MH()
	if wantMin, wantMax := 43.0+8, 66.0+8; !approxEqual(paw.BaseDamageMin, wantMin) || !approxEqual(paw.BaseDamageMax, wantMax) {
		t.Errorf("the paw deals %.3f-%.3f, want %.3f-%.3f", paw.BaseDamageMin, paw.BaseDamageMax, wantMin, wantMax)
	}
}

// A form swap rebuilds the paw from the main hand, as the druid's item swap callback does; the rebuilt
// paw keeps the stone's 8, once.
func TestCatPawKeepsTheWeaponstoneWhenRebuilt(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character, _ := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{})
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	for range 2 {
		character.AutoAttacks.SetMH(cat.GetCatWeapon())
	}
	paw := character.AutoAttacks.MH()
	if wantMin, wantMax := 43.0+8, 66.0+8; !approxEqual(paw.BaseDamageMin, wantMin) || !approxEqual(paw.BaseDamageMax, wantMax) {
		t.Errorf("the rebuilt paw deals %.3f-%.3f, want %.3f-%.3f", paw.BaseDamageMin, paw.BaseDamageMax, wantMin, wantMax)
	}
}

// Dire Bear Form swings the paw's own 109-165 every 2.5 seconds, the stone's 8 added once, through
// shifts to the paw and back: 117-173.
func TestBearWeaponTakesTheWeaponstoneOnce(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character, _ := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{})
	cat := character.Env.Raid.Parties[0].Players[0].(*pawDruid)

	for range 2 {
		character.AutoAttacks.SetMH(cat.GetBearWeapon())
		character.AutoAttacks.SetMH(cat.GetCatWeapon())
	}
	character.AutoAttacks.SetMH(cat.GetBearWeapon())
	bear := character.AutoAttacks.MH()
	if wantMin, wantMax := 109.0+8, 165.0+8; !approxEqual(bear.BaseDamageMin, wantMin) || !approxEqual(bear.BaseDamageMax, wantMax) || bear.SwingSpeed != 2.5 {
		t.Errorf("the bear weapon deals %.3f-%.3f at %.1f s, want %.3f-%.3f at 2.5 s",
			bear.BaseDamageMin, bear.BaseDamageMax, bear.SwingSpeed, wantMin, wantMax)
	}
}

// Windfury Totem's party aura reaches the cat like any other melee: its paw hears the Windfury
// procs and keeps the stone: 60 / 2.4 + 8.
func TestCatHearsWindfuryTotemAndKeepsItsMainHandImbue(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character, _ := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{WindfuryTotem: true})

	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant; got != 1703 {
		t.Errorf("the main hand carries temporary enchant %d, want Dense Weightstone's 1703", got)
	}
	if paw := character.AutoAttacks.MH(); !approxEqual(paw.BaseDamageMin, 43.0+8) {
		t.Errorf("the paw deals %.3f at least, want %.3f", paw.BaseDamageMin, 43.0+8)
	}
	if character.GetAura("Windfury Totem Trigger") == nil {
		t.Error("the cat registers no Windfury Totem proc")
	}
}

// Flametongue Totem's party aura reaches the cat the same way: its paw strikes totem fire and keeps
// the stone.
func TestCatHearsFlametongueTotemAndKeepsItsMainHandImbue(t *testing.T) {
	const monksStaff, denseWeightstone = 866, 12643
	character, _ := catWithMainHand(t, monksStaff, &proto.ConsumesSpec{MhImbueId: denseWeightstone}, &proto.RaidBuffs{FlametongueTotem: true})

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

// Whatever main hand is equipped, both paws keep their own ranges and take the flat damage whole.
func TestFeralEnchantsAndStonesAreIndependentOfWeaponSpeed(t *testing.T) {
	for _, itemID := range []int32{866, 19323} {
		character, _ := catWithMainHand(t, itemID, &proto.ConsumesSpec{MhImbueId: 12643}, &proto.RaidBuffs{})
		druid := character.Env.Raid.Parties[0].Players[0].(*pawDruid)
		item := character.GetMHWeapon()
		item.Enchant = *core.GetEnchantByEffectID(963)
		bonus := item.Enchant.WeaponDamage + item.TempEnchantWeaponDamage
		if bonus == 0 {
			t.Fatal("weapon has no flat damage to test")
		}
		for _, paw := range []core.Weapon{druid.GetCatWeapon(), druid.GetBearWeapon()} {
			baseMin, baseMax := catPawDamageMin, catPawDamageMax
			if paw.SwingSpeed == 2.5 {
				baseMin, baseMax = bearPawDamageMin, bearPawDamageMax
			}
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

// A weapon's damage per second reaches the feral forms only above the paw's own 54.8, as Feral
// Attack Power. Monk's Staff 866 (60-90, 2.4 s) is 31.25 damage per second and grants nothing;
// the same staff raised to 80 grants (80 - 54.8) * 14 while the paw keeps its own range.
func TestFeralWeaponDPSAboveThePawFloorBecomesFeralAttackPower(t *testing.T) {
	character, sim := catWithMainHand(t, 866, &proto.ConsumesSpec{}, &proto.RaidBuffs{})
	druid := character.Env.Raid.Parties[0].Players[0].(*pawDruid)
	if got := druid.GetStat(stats.FeralAttackPower); got != 0 {
		t.Fatalf("Monk's Staff grants %v feral attack power, want 0", got)
	}

	item := character.GetMHWeapon()
	item.WeaponDamageMin = 80 * item.SwingSpeed
	item.WeaponDamageMax = 80 * item.SwingSpeed
	druid.applyFeralWeaponStats(sim)
	if want := (80 - pawDPSFloor) * core.DefaultAttackPowerPerDPS; !approxEqual(druid.GetStat(stats.FeralAttackPower), want) {
		t.Errorf("80 damage per second grants %v feral attack power, want %v", druid.GetStat(stats.FeralAttackPower), want)
	}
	if paw := druid.GetCatWeapon(); !approxEqual(paw.BaseDamageMin, catPawDamageMin) || !approxEqual(paw.BaseDamageMax, catPawDamageMax) {
		t.Errorf("the paw took the weapon's range: %g-%g, want %g-%g", paw.BaseDamageMin, paw.BaseDamageMax, catPawDamageMin, catPawDamageMax)
	}
}

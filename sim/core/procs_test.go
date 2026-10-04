package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// A fixed chance bound to a weapon enchant rolls on the hits of the hand carrying the enchant, and
// moves with it when an item swap hands the enchanted weapon to the other hand.
func TestEnchantProcChanceFollowsTheEnchantedWeapon(t *testing.T) {
	const enchantedID, plainID, enchantID int32 = 990501, 990502, 990503

	oneHander := func(id int32) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Sword",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     proto.WeaponType_WeaponTypeSword,
			HandType:       proto.HandType_HandTypeOneHand,
			WeaponSpeed:    2.6,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}},
		}
	}
	addToDatabase(&proto.SimDatabase{
		Items:    []*proto.SimItem{oneHander(enchantedID), oneHander(plainID)},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Fiery Blaze"}},
	})

	hands := func(mainHand, offHand *proto.ItemSpec) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = mainHand
		items[proto.ItemSlot_ItemSlotOffHand] = offHand
		return items
	}
	enchanted := &proto.ItemSpec{Id: enchantedID, Enchant: enchantID}
	plain := &proto.ItemSpec{Id: plainID}

	agent := &FakeAgent{Character: NewCharacter(&Party{}, 0, &proto.Player{
		Name:           "Swap Tester",
		Class:          proto.Class_ClassWarrior,
		Race:           proto.Race_RaceHuman,
		Spec:           &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment:      &proto.EquipmentSpec{Items: hands(enchanted, plain)},
		EnableItemSwap: true,
		ItemSwap:       &proto.ItemSwap{Items: hands(plain, enchanted)},
	})}
	character := &agent.Character
	character.ItemSwap.initialize(character)
	character.EnableAutoAttacks(agent, AutoAttackOptions{
		MainHand:       character.WeaponFromMainHand(),
		OffHand:        character.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})

	dpm := character.NewDynamicLegacyProcForEnchant(enchantID, 0, 0.15)

	check := func(when string, mainHand, offHand float64) {
		t.Helper()
		if got, _ := dpm.chanceFor(ProcMaskMeleeMHAuto, nil); got != mainHand {
			t.Errorf("%s: main hand chance = %v, want %v", when, got, mainHand)
		}
		if got, _ := dpm.chanceFor(ProcMaskMeleeOHAuto, nil); got != offHand {
			t.Errorf("%s: off hand chance = %v, want %v", when, got, offHand)
		}
	}
	check("enchant on the main hand", 0.15, 0)

	// SwapItems' per-slot step. The rest of it moves stats and activates auras, which needs a running
	// sim, and a prepull swap exchanges the items without one.
	for _, slot := range character.ItemSwap.slots {
		character.ItemSwap.swapItem(nil, slot, true, false)
		for _, onSwap := range character.ItemSwap.onItemSwapCallbacks[slot] {
			onSwap(nil, slot)
		}
	}
	if character.OffHand().Enchant.EffectID != enchantID {
		t.Fatal("the swap did not move the enchanted weapon to the off hand")
	}
	check("enchant swapped to the off hand", 0, 0.15)
}

// A procs-per-minute rate bound to a weapon enchant heard on melee and spell damage: a melee hit rolls
// only on a hand carrying the enchant, at that weapon's speed, and a spell hit never rolls.
func TestEnchantPPMFollowsTheEnchantedWeapon(t *testing.T) {
	const slowID, fastID, enchantID int32 = 990701, 990702, 990703
	const slow, fast, ppm = 2.6, 1.8, 2.0

	weapon := func(id int32, speed float64) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Sword",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     proto.WeaponType_WeaponTypeSword,
			HandType:       proto.HandType_HandTypeOneHand,
			WeaponSpeed:    speed,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}},
		}
	}
	addToDatabase(&proto.SimDatabase{
		Items:    []*proto.SimItem{weapon(slowID, slow), weapon(fastID, fast)},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Rate Enchant", Type: proto.ItemType_ItemTypeWeapon}},
	})

	spec := func(id int32, enchanted bool) *proto.ItemSpec {
		if enchanted {
			return &proto.ItemSpec{Id: id, Enchant: enchantID}
		}
		return &proto.ItemSpec{Id: id}
	}
	hands := func(mainHand, offHand *proto.ItemSpec) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = mainHand
		items[proto.ItemSlot_ItemSlotOffHand] = offHand
		return items
	}
	chance := func(speed float64) float64 { return speed * (ppm / 60) }

	type chances struct{ mainHand, offHand, spell float64 }
	for _, tc := range []struct {
		name        string
		equipped    []*proto.ItemSpec
		swapped     []*proto.ItemSpec
		before      chances
		afterSwap   chances
		swapSummary string
	}{
		{
			name:        "main hand only",
			equipped:    hands(spec(slowID, true), spec(fastID, false)),
			swapped:     hands(spec(fastID, false), spec(slowID, true)),
			before:      chances{chance(slow), 0, 0},
			afterSwap:   chances{0, chance(slow), 0},
			swapSummary: "the enchanted weapon swapped to the off hand",
		},
		{
			name:        "off hand only",
			equipped:    hands(spec(slowID, false), spec(fastID, true)),
			swapped:     hands(spec(fastID, true), spec(slowID, false)),
			before:      chances{0, chance(fast), 0},
			afterSwap:   chances{chance(fast), 0, 0},
			swapSummary: "the enchanted weapon swapped to the main hand",
		},
		{
			name:        "both hands",
			equipped:    hands(spec(slowID, true), spec(fastID, true)),
			swapped:     hands(spec(slowID, false), spec(fastID, false)),
			before:      chances{chance(slow), chance(fast), 0},
			afterSwap:   chances{0, 0, 0},
			swapSummary: "both swapped for unenchanted weapons",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			agent := &FakeAgent{Character: NewCharacter(&Party{}, 0, &proto.Player{
				Name:           "Rate Tester",
				Class:          proto.Class_ClassWarrior,
				Race:           proto.Race_RaceHuman,
				Spec:           &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
				Equipment:      &proto.EquipmentSpec{Items: tc.equipped},
				EnableItemSwap: true,
				ItemSwap:       &proto.ItemSwap{Items: tc.swapped},
			})}
			character := &agent.Character
			character.ItemSwap.initialize(character)
			character.EnableAutoAttacks(agent, AutoAttackOptions{
				MainHand:       character.WeaponFromMainHand(),
				OffHand:        character.WeaponFromOffHand(),
				AutoSwingMelee: true,
			})

			dpm := character.NewDynamicLegacyProcForEnchantWithMask(enchantID, ppm, ProcMaskMelee|ProcMaskSpellDamage)

			check := func(when string, want chances) {
				t.Helper()
				for _, hit := range []struct {
					name string
					mask ProcMask
					want float64
				}{
					{"main hand", ProcMaskMeleeMHAuto, want.mainHand},
					{"off hand", ProcMaskMeleeOHAuto, want.offHand},
					{"spell", ProcMaskSpellDamage, want.spell},
				} {
					if got, _ := dpm.chanceFor(hit.mask, nil); got != hit.want {
						t.Errorf("%s: %s chance = %v, want %v", when, hit.name, got, hit.want)
					}
				}
			}
			check("as equipped", tc.before)

			// SwapItems' per-slot step. A prepull swap exchanges the items without a sim and leaves
			// the weapons to the rest of it, so they are set here as a swap in combat sets them.
			for _, slot := range character.ItemSwap.slots {
				character.ItemSwap.swapItem(nil, slot, true, false)
			}
			character.AutoAttacks.SetMH(character.WeaponFromMainHand())
			character.AutoAttacks.SetOH(character.WeaponFromOffHand())
			for _, slot := range character.ItemSwap.slots {
				for _, onSwap := range character.ItemSwap.onItemSwapCallbacks[slot] {
					onSwap(nil, slot)
				}
			}
			check(tc.swapSummary, tc.afterSwap)
		})
	}
}

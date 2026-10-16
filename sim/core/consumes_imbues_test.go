package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestTemporaryEnchantTakesTheWeaponsItsMasksName(t *testing.T) {
	sharpeningStone := Consumable{WeaponSubclassMask: 33219}    // Dense Sharpening Stone 12404: 16138
	weightstone := Consumable{WeaponSubclassMask: 9264}         // Dense Weightstone 12643: 16622
	wizardOil := Consumable{WeaponInventoryTypeMask: 6430720}   // Brilliant Wizard Oil 20749: 25122
	shieldCoating := Consumable{WeaponInventoryTypeMask: 16384} // Conductive Shield Coating 228980: 463868

	sword := Item{ID: 1, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	twoHandMace := Item{ID: 2, WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeTwoHand}
	dagger := Item{ID: 3, WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOffHand}
	shield := Item{ID: 4, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand}

	for _, tc := range []struct {
		name       string
		consumable Consumable
		item       Item
		want       bool
	}{
		{"sharpening stone on a sword", sharpeningStone, sword, true},
		{"sharpening stone on a dagger", sharpeningStone, dagger, true},
		{"sharpening stone on a mace", sharpeningStone, twoHandMace, false},
		{"weightstone on a mace", weightstone, twoHandMace, true},
		{"weightstone on a sword", weightstone, sword, false},
		{"oil on an off-hand dagger", wizardOil, dagger, true},
		{"oil on a two-hand mace", wizardOil, twoHandMace, true},
		{"oil on a shield", wizardOil, shield, false},
		{"coating on a shield", shieldCoating, shield, true},
		{"coating on a sword", shieldCoating, sword, false},
		{"stone on an empty hand", sharpeningStone, Item{}, false},
	} {
		if got := tc.consumable.TakesItem(proto.ItemSlot_ItemSlotMainHand, tc.item); got != tc.want {
			t.Errorf("%s: takes = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestTemporaryEnchantGoesOnTheHandsTheSpecNames(t *testing.T) {
	const swordID, maceID, stoneID, stoneEnchantID = 990301, 990302, 990303, 990304

	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{
			{Id: swordID, Name: "Test Sword", Type: proto.ItemType_ItemTypeWeapon,
				WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6,
				ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 100, WeaponDamageMax: 150}}},
			{Id: maceID, Name: "Test Mace", Type: proto.ItemType_ItemTypeWeapon,
				WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6,
				ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 100, WeaponDamageMax: 150}}},
		},
		Consumables: []*proto.Consumable{{
			Id: stoneID, Name: "Test Sharpening Stone", Type: proto.ConsumableType_ConsumableTypeImbue,
			WeaponSubclassMask: 33219, TemporaryEnchant: &proto.SimEnchant{EffectId: stoneEnchantID, WeaponDamage: 8},
		}},
	})

	items := make([]*proto.ItemSpec, NumItemSlots)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = swordID
	items[proto.ItemSlot_ItemSlotOffHand].Id = maceID
	character := NewCharacter(&Party{buffs: &proto.PartyBuffs{}}, 0, &proto.Player{
		Class:       proto.Class_ClassWarrior,
		Race:        proto.Race_RaceHuman,
		Spec:        &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{MhImbueId: stoneID, OhImbueId: stoneID},
	})
	if got := character.Equipment[proto.ItemSlot_ItemSlotMainHand].TempEnchant; got != stoneEnchantID {
		t.Errorf("main-hand sword carries enchant %d, want %d", got, stoneEnchantID)
	}
	if got := character.Equipment[proto.ItemSlot_ItemSlotOffHand].TempEnchant; got != 0 {
		t.Errorf("off-hand mace carries enchant %d, which a sharpening stone does not take", got)
	}
	if mh := character.WeaponFromMainHand(); mh.BaseDamageMin != 108 || mh.BaseDamageMax != 158 {
		t.Errorf("main-hand sword deals %v-%v, want 108-158 with the stone's 8", mh.BaseDamageMin, mh.BaseDamageMax)
	}
	if oh := character.WeaponFromOffHand(); oh.BaseDamageMin != 100 || oh.BaseDamageMax != 150 {
		t.Errorf("off-hand mace deals %v-%v, want its own 100-150", oh.BaseDamageMin, oh.BaseDamageMax)
	}
}

// The imbues the spec names for the item swap set go on that set's weapons where their masks take
// them, and the main set's weapons stay bare. The swap's stat offset carries the enchants' stats.
func TestSwapImbueGoesOnTheSwapWeaponsItTakes(t *testing.T) {
	const swordID, maceID, stoneID, stoneEnchantID, firestoneID, firestoneEnchantID = 990321, 990322, 990323, 990324, 990325, 990326

	weapon := func(id int32, weaponType proto.WeaponType) *proto.SimItem {
		return &proto.SimItem{Id: id, Name: "Test Weapon", Type: proto.ItemType_ItemTypeWeapon,
			WeaponType: weaponType, HandType: proto.HandType_HandTypeOneHand, WeaponSpeed: 2.6,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 100, WeaponDamageMax: 150}}}
	}
	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{weapon(swordID, proto.WeaponType_WeaponTypeSword), weapon(maceID, proto.WeaponType_WeaponTypeMace)},
		Consumables: []*proto.Consumable{
			{Id: stoneID, Name: "Test Sharpening Stone", Type: proto.ConsumableType_ConsumableTypeImbue,
				WeaponSubclassMask: 33219, TemporaryEnchant: &proto.SimEnchant{EffectId: stoneEnchantID, WeaponDamage: 8}},
			{Id: firestoneID, Name: "Test Firestone", Type: proto.ConsumableType_ConsumableTypeWeaponImbue,
				TemporaryEnchant: &proto.SimEnchant{EffectId: firestoneEnchantID, Stats: stats.Stats{stats.FireDamage: 14}.ToProtoArray()}},
		},
	})

	hands := func(mainHand, offHand int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, NumItemSlots)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand].Id = mainHand
		items[proto.ItemSlot_ItemSlotOffHand].Id = offHand
		return items
	}
	character := NewCharacter(&Party{}, 0, &proto.Player{
		Class:          proto.Class_ClassWarrior,
		Race:           proto.Race_RaceHuman,
		Spec:           &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
		Equipment:      &proto.EquipmentSpec{Items: hands(swordID, maceID)},
		EnableItemSwap: true,
		ItemSwap:       &proto.ItemSwap{Items: hands(maceID, swordID)},
		Consumables: &proto.ConsumesSpec{
			SwapMhImbueId:       stoneID,
			SwapOhImbueId:       stoneID,
			SwapMhWeaponImbueId: firestoneID,
		},
	})

	stone := GetConsumableByID(stoneID).TemporaryEnchant
	firestone := GetConsumableByID(firestoneID).TemporaryEnchant
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		if item := character.Equipment[slot]; item.TempEnchant != 0 || item.WeaponImbue != 0 || item.TempEnchantStats != nil {
			t.Errorf("main set %s carries enchant %d and imbue %d, want neither", slot, item.TempEnchant, item.WeaponImbue)
		}
	}
	for name, set := range map[string]Equipment{"swap set": character.ItemSwap.swapEquip, "unequipped items": character.ItemSwap.unEquippedItems} {
		mace, sword := set[proto.ItemSlot_ItemSlotMainHand], set[proto.ItemSlot_ItemSlotOffHand]
		if mace.TempEnchant != 0 || mace.WeaponImbue != firestone.EffectID {
			t.Errorf("%s: main-hand mace carries enchant %d and imbue %d, want no stone and the firestone's %d", name, mace.TempEnchant, mace.WeaponImbue, firestone.EffectID)
		}
		if sword.TempEnchant != stone.EffectID || sword.TempEnchantWeaponDamage != stone.WeaponDamage {
			t.Errorf("%s: off-hand sword carries enchant %d with %v weapon damage, want the stone's %d with %v", name, sword.TempEnchant, sword.TempEnchantWeaponDamage, stone.EffectID, stone.WeaponDamage)
		}
	}
	if got := character.ItemSwap.equipmentStats.weaponSlots; got != firestone.Stats {
		t.Errorf("the swap's weapon stat offset is %v, want the firestone's %v", got, firestone.Stats)
	}
}

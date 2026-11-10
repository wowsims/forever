package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// An AllItems test weapon takes the hand a player wears it in, whatever fills the hands, and keeps
// the replaced weapon's enchant where that enchant fits it.
func TestEquipTestItemTakesTheWornSlot(t *testing.T) {
	weapon := func(id int32, hand proto.HandType) Item {
		return Item{ID: id, Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: hand}
	}
	weaponEnchant := Enchant{EffectID: 7, Type: proto.ItemType_ItemTypeWeapon}
	twoHandEnchant := Enchant{EffectID: 8, Type: proto.ItemType_ItemTypeWeapon, EnchantType: proto.EnchantType_EnchantTypeTwoHand}

	var dualWielder Equipment
	*dualWielder.MainHand() = weapon(1, proto.HandType_HandTypeOneHand)
	dualWielder.MainHand().Enchant = weaponEnchant
	*dualWielder.OffHand() = weapon(2, proto.HandType_HandTypeOneHand)

	var twoHander Equipment
	*twoHander.MainHand() = weapon(5, proto.HandType_HandTypeTwoHand)
	twoHander.MainHand().Enchant = twoHandEnchant

	for name, tc := range map[string]struct {
		base            Equipment
		item            Item
		mainHand        int32
		mainHandEnchant int32
		offHand         int32
	}{
		"a one-hander replaces a dual wielder's main hand and keeps its enchant": {dualWielder, weapon(10, proto.HandType_HandTypeOneHand), 10, 7, 2},
		"a two-hander takes a dual wielder's both hands and the weapon enchant":  {dualWielder, weapon(12, proto.HandType_HandTypeTwoHand), 12, 7, 0},
		"an off-hand weapon replaces the off hand":                               {dualWielder, weapon(13, proto.HandType_HandTypeOffHand), 1, 7, 13},
		"a ring leaves the hands":                                                {dualWielder, Item{ID: 15, Type: proto.ItemType_ItemTypeFinger}, 1, 7, 2},
		"a one-hander replaces a two-hander and drops its two-hand enchant":      {twoHander, weapon(10, proto.HandType_HandTypeOneHand), 10, 0, 0},
		"a two-hander replaces a two-hander and keeps its enchant":               {twoHander, weapon(12, proto.HandType_HandTypeTwoHand), 12, 8, 0},
	} {
		equipment := tc.base
		equipTestItem(&equipment, tc.item)
		if got := equipment.MainHand(); got.ID != tc.mainHand || got.Enchant.EffectID != tc.mainHandEnchant {
			t.Errorf("%s: main hand holds %d with enchant %d, want %d with %d", name, got.ID, got.Enchant.EffectID, tc.mainHand, tc.mainHandEnchant)
		}
		if got := equipment.OffHand().ID; got != tc.offHand {
			t.Errorf("%s: off hand holds %d, want %d", name, got, tc.offHand)
		}
	}
}

// The AllItems filter admits a weapon only where the class can wear it: the weapon type, two hands
// where the type allows them, and the ranged type; a filter with no class keeps every weapon.
func TestItemFilterKeepsOutTheWeaponsTheClassCannotWear(t *testing.T) {
	weapon := func(weaponType proto.WeaponType, hand proto.HandType) Item {
		return Item{Type: proto.ItemType_ItemTypeWeapon, WeaponType: weaponType, HandType: hand}
	}
	ranged := func(rangedType proto.RangedWeaponType) Item {
		return Item{Type: proto.ItemType_ItemTypeRanged, RangedWeaponType: rangedType}
	}
	for name, tc := range map[string]struct {
		class proto.Class
		item  Item
		want  bool
	}{
		"a mage's one-hand sword":    {proto.Class_ClassMage, weapon(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand), true},
		"a mage's two-hand sword":    {proto.Class_ClassMage, weapon(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeTwoHand), false},
		"a mage's staff":             {proto.Class_ClassMage, weapon(proto.WeaponType_WeaponTypeStaff, proto.HandType_HandTypeTwoHand), true},
		"a hunter's shield":          {proto.Class_ClassHunter, weapon(proto.WeaponType_WeaponTypeShield, proto.HandType_HandTypeOffHand), false},
		"a hunter's thrown weapon":   {proto.Class_ClassHunter, ranged(proto.RangedWeaponType_RangedWeaponTypeThrown), false},
		"a hunter's bow":             {proto.Class_ClassHunter, ranged(proto.RangedWeaponType_RangedWeaponTypeBow), true},
		"no class, a two-hand sword": {proto.Class_ClassUnknown, weapon(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeTwoHand), true},
	} {
		filter := ItemFilter{Class: tc.class}
		if got := filter.Matches(tc.item, true); got != tc.want {
			t.Errorf("%s: the filter answers %v, want %v", name, got, tc.want)
		}
	}
}

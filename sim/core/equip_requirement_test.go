package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

func weaponItem(weaponType proto.WeaponType, handType proto.HandType) Item {
	return Item{ID: 1, WeaponType: weaponType, HandType: handType}
}

func rangedItem(rangedType proto.RangedWeaponType) Item {
	return Item{ID: 1, RangedWeaponType: rangedType}
}

var equipTestItems = []struct {
	name     string
	item     Item
	subclass int32
}{
	{"one-hand axe", weaponItem(proto.WeaponType_WeaponTypeAxe, proto.HandType_HandTypeOneHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE},
	{"two-hand axe", weaponItem(proto.WeaponType_WeaponTypeAxe, proto.HandType_HandTypeTwoHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE},
	{"one-hand mace", weaponItem(proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeMainHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H},
	{"two-hand mace", weaponItem(proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeTwoHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H},
	{"one-hand sword", weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H},
	{"off-hand sword", weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOffHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H},
	{"two-hand sword", weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeTwoHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H},
	{"polearm", weaponItem(proto.WeaponType_WeaponTypePolearm, proto.HandType_HandTypeTwoHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM},
	{"staff", weaponItem(proto.WeaponType_WeaponTypeStaff, proto.HandType_HandTypeTwoHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF},
	{"fist weapon", weaponItem(proto.WeaponType_WeaponTypeFist, proto.HandType_HandTypeOneHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST},
	{"dagger", weaponItem(proto.WeaponType_WeaponTypeDagger, proto.HandType_HandTypeOneHand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS},
	{"bow", rangedItem(proto.RangedWeaponType_RangedWeaponTypeBow), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW},
	{"gun", rangedItem(proto.RangedWeaponType_RangedWeaponTypeGun), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS},
	{"crossbow", rangedItem(proto.RangedWeaponType_RangedWeaponTypeCrossbow), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW},
	{"thrown", rangedItem(proto.RangedWeaponType_RangedWeaponTypeThrown), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_THROWN},
	{"wand", rangedItem(proto.RangedWeaponType_RangedWeaponTypeWand), dbcenums.ITEM_SUBCLASS_BIT_WEAPON_WAND},
	{"shield", weaponItem(proto.WeaponType_WeaponTypeShield, proto.HandType_HandTypeOffHand), 0},
	{"held in off hand", weaponItem(proto.WeaponType_WeaponTypeOffHand, proto.HandType_HandTypeOffHand), 0},
	{"libram", rangedItem(proto.RangedWeaponType_RangedWeaponTypeLibram), 0},
	{"empty slot", Item{}, 0},
}

func TestWeaponSubclassBit(t *testing.T) {
	for _, test := range equipTestItems {
		if got := test.item.WeaponSubclassBit(); got != test.subclass {
			t.Errorf("%s: WeaponSubclassBit = %#x, want %#x", test.name, got, test.subclass)
		}
	}
}

func TestInvTypeBit(t *testing.T) {
	mh, oh, ranged := proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand, proto.ItemSlot_ItemSlotRanged
	oneHand := weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand)

	tests := []struct {
		name string
		slot proto.ItemSlot
		item Item
		want int32
	}{
		{"one-hand in the main hand", mh, oneHand, dbcenums.INVTYPE_BIT_WEAPON | dbcenums.INVTYPE_BIT_WEAPONMAINHAND},
		{"one-hand in the off hand", oh, oneHand, dbcenums.INVTYPE_BIT_WEAPON | dbcenums.INVTYPE_BIT_WEAPONOFFHAND},
		{"main-hand only", mh, weaponItem(proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeMainHand), dbcenums.INVTYPE_BIT_WEAPONMAINHAND},
		{"off-hand only", oh, weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOffHand), dbcenums.INVTYPE_BIT_WEAPONOFFHAND},
		{"two-hand", mh, weaponItem(proto.WeaponType_WeaponTypeStaff, proto.HandType_HandTypeTwoHand), dbcenums.INVTYPE_BIT_2HWEAPON},
		{"shield", oh, weaponItem(proto.WeaponType_WeaponTypeShield, proto.HandType_HandTypeOffHand), dbcenums.INVTYPE_BIT_SHIELD},
		{"held in off hand", oh, weaponItem(proto.WeaponType_WeaponTypeOffHand, proto.HandType_HandTypeOffHand), dbcenums.INVTYPE_BIT_HOLDABLE},
		{"bow", ranged, rangedItem(proto.RangedWeaponType_RangedWeaponTypeBow), dbcenums.INVTYPE_BIT_RANGED},
		{"thrown", ranged, rangedItem(proto.RangedWeaponType_RangedWeaponTypeThrown), dbcenums.INVTYPE_BIT_THROWN},
		{"gun", ranged, rangedItem(proto.RangedWeaponType_RangedWeaponTypeGun), dbcenums.INVTYPE_BIT_RANGEDRIGHT},
		{"crossbow", ranged, rangedItem(proto.RangedWeaponType_RangedWeaponTypeCrossbow), dbcenums.INVTYPE_BIT_RANGEDRIGHT},
		{"wand", ranged, rangedItem(proto.RangedWeaponType_RangedWeaponTypeWand), dbcenums.INVTYPE_BIT_RANGEDRIGHT},
		{"empty slot", mh, Item{}, 0},
	}

	for _, test := range tests {
		if got := invTypeBit(test.slot, &test.item); got != test.want {
			t.Errorf("%s: invTypeBit = %#x, want %#x", test.name, got, test.want)
		}
	}
	if invTypeBit(oh, &oneHand)&dbcenums.INVTYPE_BIT_WEAPONOFFHAND == 0 || invTypeBit(mh, &oneHand)&dbcenums.INVTYPE_BIT_WEAPONOFFHAND != 0 {
		t.Error("the off-hand inventory type 22 should name a one-hand weapon in the off hand and not in the main hand")
	}
}

// Every weapon subclass mask a stat aura in the spell store names, decoded bit by bit, and the items
// each one takes. Spear, exotic, misc and fishing-pole bits name subclasses no item carries, so the
// two melee masks take the same items.
func TestStoreSubclassMasks(t *testing.T) {
	melee := []string{"one-hand axe", "two-hand axe", "one-hand mace", "two-hand mace", "one-hand sword", "off-hand sword", "two-hand sword", "polearm", "staff", "fist weapon", "dagger"}

	tests := []struct {
		mask  int32
		bits  int32
		items []string
	}{
		{173555, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SPEAR, melee},
		{42483, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_2H |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_POLEARM | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS, melee},
		{41105, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_MACE_1H |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS,
			[]string{"one-hand axe", "one-hand mace", "one-hand sword", "off-hand sword", "fist weapon", "dagger"}},
		{262156, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW, []string{"bow", "gun", "crossbow"}},
		{327692, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS |
			dbcenums.ITEM_SUBCLASS_BIT_WEAPON_THROWN | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW,
			[]string{"bow", "gun", "crossbow", "thrown"}},
		{32768, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS, []string{"dagger"}},
		{8192, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_FIST, []string{"fist weapon"}},
	}

	for _, test := range tests {
		if test.bits != test.mask {
			t.Errorf("mask %d decodes to %d", test.mask, test.bits)
		}
		var got []string
		for _, item := range equipTestItems {
			if item.item.WeaponSubclassBit()&test.mask != 0 {
				got = append(got, item.name)
			}
		}
		slices.Sort(got)
		slices.Sort(test.items)
		if !slices.Equal(got, test.items) {
			t.Errorf("mask %d takes %v, want %v", test.mask, got, test.items)
		}
	}
}

func TestWeaponSubclassMask(t *testing.T) {
	tests := []struct {
		types []proto.WeaponType
		want  int32
	}{
		{[]proto.WeaponType{proto.WeaponType_WeaponTypeSword}, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_1H | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_SWORD_2H},
		{[]proto.WeaponType{proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeDagger}, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_DAGGERS},
		{[]proto.WeaponType{proto.WeaponType_WeaponTypeStaff}, dbcenums.ITEM_SUBCLASS_BIT_WEAPON_STAFF},
		{[]proto.WeaponType{proto.WeaponType_WeaponTypeShield, proto.WeaponType_WeaponTypeOffHand}, 0},
	}
	for _, test := range tests {
		if got := WeaponSubclassMask(test.types...); got != test.want {
			t.Errorf("WeaponSubclassMask(%v) = %#x, want %#x", test.types, got, test.want)
		}
	}
}

func TestEquipRequirementEquipped(t *testing.T) {
	sword := WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)
	oneHandSword := weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeOneHand)
	twoHandSword := weaponItem(proto.WeaponType_WeaponTypeSword, proto.HandType_HandTypeTwoHand)
	dagger := weaponItem(proto.WeaponType_WeaponTypeDagger, proto.HandType_HandTypeOneHand)
	bow := rangedItem(proto.RangedWeaponType_RangedWeaponTypeBow)

	tests := []struct {
		name                      string
		unitType                  UnitType
		mainHand, offHand, ranged Item
		req                       EquipRequirement
		want                      bool
	}{
		{"two-handed sword", PlayerUnit, twoHandSword, Item{}, bow, EquipRequirement{SubclassMask: sword}, true},
		{"sword in the off hand", PlayerUnit, dagger, oneHandSword, bow, EquipRequirement{SubclassMask: sword}, true},
		{"no sword", PlayerUnit, dagger, dagger, bow, EquipRequirement{SubclassMask: sword}, false},
		{"bow in the ranged slot", PlayerUnit, dagger, dagger, bow, EquipRequirement{SubclassMask: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING}, true},
		{"off hand named, sword in the main hand", PlayerUnit, oneHandSword, dagger, bow,
			EquipRequirement{SubclassMask: sword, Hand: ProcMaskMeleeOH}, false},
		{"pet", PetUnit, oneHandSword, dagger, bow, EquipRequirement{SubclassMask: sword}, false},
	}
	for _, test := range tests {
		character := &Character{Unit: Unit{Type: test.unitType}}
		character.Equipment[proto.ItemSlot_ItemSlotMainHand] = test.mainHand
		character.Equipment[proto.ItemSlot_ItemSlotOffHand] = test.offHand
		character.Equipment[proto.ItemSlot_ItemSlotRanged] = test.ranged
		if got := test.req.Equipped(character); got != test.want {
			t.Errorf("%s: Equipped = %v, want %v", test.name, got, test.want)
		}
	}
}

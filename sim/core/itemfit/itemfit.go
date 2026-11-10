// Package itemfit holds the rules for which slots an item goes in and which enchant rows fit an item. The sim's gear
// checks and the SIML editor both read them, so the package imports only proto.
package itemfit

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
)

// What an item's slots and its enchants depend on.
type Item struct {
	Type       proto.ItemType
	HandType   proto.HandType
	WeaponType proto.WeaponType
	RangedType proto.RangedWeaponType
}

// What an enchant row's slots and the items it fits depend on.
type Enchant struct {
	Type        proto.ItemType
	EnchantType proto.EnchantType
	ExtraTypes  []proto.ItemType // Further item types the row goes on, as an armor kit's.
}

// The slots of each item type but the weapon, whose slots its hand type gives. gen_db -gen=go-to-ts writes it to the
// UI as ITEM_TYPE_TO_SLOTS, read by getEligibleItemSlots in ui/sim/proto/items.ts.
var TypeSlots = map[proto.ItemType][]proto.ItemSlot{
	proto.ItemType_ItemTypeUnknown:  {},
	proto.ItemType_ItemTypeHead:     {proto.ItemSlot_ItemSlotHead},
	proto.ItemType_ItemTypeNeck:     {proto.ItemSlot_ItemSlotNeck},
	proto.ItemType_ItemTypeShoulder: {proto.ItemSlot_ItemSlotShoulder},
	proto.ItemType_ItemTypeBack:     {proto.ItemSlot_ItemSlotBack},
	proto.ItemType_ItemTypeChest:    {proto.ItemSlot_ItemSlotChest},
	proto.ItemType_ItemTypeWrist:    {proto.ItemSlot_ItemSlotWrist},
	proto.ItemType_ItemTypeHands:    {proto.ItemSlot_ItemSlotHands},
	proto.ItemType_ItemTypeWaist:    {proto.ItemSlot_ItemSlotWaist},
	proto.ItemType_ItemTypeLegs:     {proto.ItemSlot_ItemSlotLegs},
	proto.ItemType_ItemTypeFeet:     {proto.ItemSlot_ItemSlotFeet},
	proto.ItemType_ItemTypeFinger:   {proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2},
	proto.ItemType_ItemTypeTrinket:  {proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2},
	proto.ItemType_ItemTypeRanged:   {proto.ItemSlot_ItemSlotRanged},
}

// The slots an item goes in.
func ItemSlots(item Item) []proto.ItemSlot {
	if slots, ok := TypeSlots[item.Type]; ok {
		return slots
	}
	if item.Type == proto.ItemType_ItemTypeWeapon {
		switch item.HandType {
		case proto.HandType_HandTypeTwoHand, proto.HandType_HandTypeMainHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
		case proto.HandType_HandTypeOffHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}
		case proto.HandType_HandTypeOneHand:
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
		}
	}
	return nil
}

// The slots an enchant row's item types reach; a weapon enchant reaches both hands.
func EnchantSlots(enchant Enchant) []proto.ItemSlot {
	types := append([]proto.ItemType{enchant.Type}, enchant.ExtraTypes...)
	slots := make([]proto.ItemSlot, 0, len(types)*2)
	for _, itemType := range types {
		if typeSlots, ok := TypeSlots[itemType]; ok {
			slots = append(slots, typeSlots...)
			continue
		}
		if itemType == proto.ItemType_ItemTypeWeapon {
			slots = append(slots, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
		}
	}
	return slots
}

// Whether an enchant row fits an item.
func EnchantFits(enchant Enchant, item Item) bool {
	switch {
	case enchant.EnchantType == proto.EnchantType_EnchantTypeTwoHand && item.HandType != proto.HandType_HandTypeTwoHand,
		enchant.EnchantType == proto.EnchantType_EnchantTypeStaff && item.WeaponType != proto.WeaponType_WeaponTypeStaff,
		(enchant.EnchantType == proto.EnchantType_EnchantTypeShield) != (item.WeaponType == proto.WeaponType_WeaponTypeShield),
		(enchant.EnchantType == proto.EnchantType_EnchantTypeOffHand) != (item.WeaponType == proto.WeaponType_WeaponTypeOffHand):
		return false
	}
	ranged := item.RangedType
	if enchant.Type == proto.ItemType_ItemTypeRanged {
		// Bows, crossbows and guns: the ranged weapons a ranged enchant fits.
		if ranged != proto.RangedWeaponType_RangedWeaponTypeBow && ranged != proto.RangedWeaponType_RangedWeaponTypeCrossbow &&
			ranged != proto.RangedWeaponType_RangedWeaponTypeGun {
			return false
		}
	} else if ranged != proto.RangedWeaponType_RangedWeaponTypeUnknown && ranged != proto.RangedWeaponType_RangedWeaponTypeWand {
		return false
	}
	slots := ItemSlots(item)
	return slices.ContainsFunc(EnchantSlots(enchant), func(s proto.ItemSlot) bool { return slices.Contains(slots, s) })
}

package core

import (
	"github.com/wowsims/forever/sim/core/proto"
)

const (
	DenseSharpeningStoneID int32 = 12404
	DenseWeightstoneID     int32 = 12643
)

// The weapon types a sharpening stone fits, and those a weightstone fits.
var (
	SharpWeaponTypes = []proto.WeaponType{proto.WeaponType_WeaponTypeAxe, proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypePolearm, proto.WeaponType_WeaponTypeSword}
	BluntWeaponTypes = []proto.WeaponType{proto.WeaponType_WeaponTypeFist, proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff}
)

// Puts the temporary enchant each imbue the spec names on its hand, where the enchant takes that
// hand's weapon. It runs before the weapons are built, and a weapon built from a hand's item takes
// the enchant's flat weapon damage. The consumable's registered effect then applies the rest of the
// enchant to the hands carrying it.
func applyTemporaryEnchants(character *Character, consumables *proto.ConsumesSpec) {
	character.applyTemporaryEnchant(consumables.MhImbueId, proto.ItemSlot_ItemSlotMainHand)
	character.applyTemporaryEnchant(consumables.OhImbueId, proto.ItemSlot_ItemSlotOffHand)
	character.applyTemporaryEnchant(consumables.MhWeaponImbueId, proto.ItemSlot_ItemSlotMainHand)
	character.applyTemporaryEnchant(consumables.OhWeaponImbueId, proto.ItemSlot_ItemSlotOffHand)
}

func (character *Character) isFeralDruid() bool {
	return character.Spec == proto.Spec_SpecFeralCatDruid || character.Spec == proto.Spec_SpecFeralBearDruid
}

func (character *Character) applyTemporaryEnchant(itemID int32, slot proto.ItemSlot) {
	consumable := GetConsumableByID(itemID)
	if consumable.TemporaryEnchant == nil || !consumable.TakesItem(slot, character.Equipment[slot]) {
		return
	}
	item := &character.Equipment[slot]
	if consumable.Type == proto.ConsumableType_ConsumableTypeWeaponImbue {
		item.WeaponImbue = consumable.TemporaryEnchant.EffectID
		return
	}
	item.TempEnchant = consumable.TemporaryEnchant.EffectID
	item.TempEnchantWeaponDamage = consumable.TemporaryEnchant.WeaponDamage
}

// Whether the temporary enchant's SpellEquippedItems masks take the item in this slot.
func (consumable Consumable) TakesItem(slot proto.ItemSlot, item Item) bool {
	if item.ID == 0 {
		return false
	}
	return (consumable.WeaponSubclassMask == 0 || consumable.WeaponSubclassMask&item.WeaponSubclassBit() != 0) &&
		(consumable.WeaponInventoryTypeMask == 0 || consumable.WeaponInventoryTypeMask&invTypeBit(slot, &item) != 0)
}

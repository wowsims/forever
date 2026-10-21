package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
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
// the enchant's flat weapon damage; the enchant's stats count among the item's equipment stats and
// its speed percents among the item's speed auras, so both leave with the weapon on an item swap. The
// consumable's registered effect then applies the rest of the enchant to the hands carrying it.
//
// The item swap set takes the same path with the swap imbue ids and the swap weapons.
func applyTemporaryEnchants(mhImbueID, ohImbueID, mhWeaponImbueID, ohWeaponImbueID int32, equipment *Equipment) {
	applyTemporaryEnchant(mhImbueID, proto.ItemSlot_ItemSlotMainHand, equipment)
	applyTemporaryEnchant(ohImbueID, proto.ItemSlot_ItemSlotOffHand, equipment)
	applyTemporaryEnchant(mhWeaponImbueID, proto.ItemSlot_ItemSlotMainHand, equipment)
	applyTemporaryEnchant(ohWeaponImbueID, proto.ItemSlot_ItemSlotOffHand, equipment)
}

func (character *Character) isFeralDruid() bool {
	return character.Spec == proto.Spec_SpecFeralCatDruid || character.Spec == proto.Spec_SpecFeralBearDruid
}

func applyTemporaryEnchant(itemID int32, slot proto.ItemSlot, equipment *Equipment) {
	consumable := GetConsumableByID(itemID)
	if consumable.TemporaryEnchant == nil || !consumable.TakesItem(slot, equipment[slot]) {
		return
	}
	item := &equipment[slot]
	enchant := consumable.TemporaryEnchant
	sum := enchant.Stats.Add(stats.FromPseudoStatsProto(enchant.PseudoStats))
	if item.TempEnchantStats != nil {
		sum = sum.Add(*item.TempEnchantStats)
	}
	item.TempEnchantStats = &sum
	if consumable.Type == proto.ConsumableType_ConsumableTypeWeaponImbue {
		item.WeaponImbue = enchant.EffectID
		item.WeaponImbuePseudoStats = enchant.PseudoStats
		return
	}
	item.TempEnchant = enchant.EffectID
	item.TempEnchantWeaponDamage = enchant.WeaponDamage
	item.TempEnchantPseudoStats = enchant.PseudoStats
}

// Whether the temporary enchant's SpellEquippedItems masks take the item in this slot.
func (consumable Consumable) TakesItem(slot proto.ItemSlot, item Item) bool {
	if item.ID == 0 {
		return false
	}
	return (consumable.WeaponSubclassMask == 0 || consumable.WeaponSubclassMask&item.WeaponSubclassBit() != 0) &&
		(consumable.WeaponInventoryTypeMask == 0 || consumable.WeaponInventoryTypeMask&invTypeBit(slot, &item) != 0)
}

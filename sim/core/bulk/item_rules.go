package bulk

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func replaceItem(existing core.Item, option bulkSimCandidateOption) core.Item {
	itemSpec := existing.ToItemSpecProto()
	itemSpec.Id = option.spec.GetId()
	itemSpec.RandomSuffix = option.spec.GetRandomSuffix()

	if option.enchant != 0 {
		itemSpec.Enchant = option.enchant
	} else if !enchantAppliesToItem(itemSpec.GetEnchant(), option.item) {
		itemSpec.Enchant = 0
	}

	return core.NewItem(core.ItemSpec{
		ID:           itemSpec.GetId(),
		RandomSuffix: itemSpec.GetRandomSuffix(),
		Enchant:      itemSpec.GetEnchant(),
	})
}

func createSelectedItem(option bulkSimCandidateOption) core.Item {
	enchant := option.spec.GetEnchant()
	if option.enchant != 0 {
		enchant = option.enchant
	}
	return core.NewItem(core.ItemSpec{
		ID:           option.spec.GetId(),
		RandomSuffix: option.spec.GetRandomSuffix(),
		Enchant:      enchant,
	})
}

func enchantAppliesToItem(effectID int32, item core.Item) bool {
	if effectID == 0 {
		return false
	}
	return core.AnyEnchantRow(effectID, func(enchant core.Enchant) bool {
		return enchantRowAppliesToItem(enchant, item)
	})
}

func enchantRowAppliesToItem(enchant core.Enchant, item core.Item) bool {
	if enchant.EnchantType == proto.EnchantType_EnchantTypeTwoHand && item.HandType != proto.HandType_HandTypeTwoHand {
		return false
	}
	if enchant.EnchantType == proto.EnchantType_EnchantTypeStaff && item.WeaponType != proto.WeaponType_WeaponTypeStaff {
		return false
	}
	if (enchant.EnchantType == proto.EnchantType_EnchantTypeShield) != (item.WeaponType == proto.WeaponType_WeaponTypeShield) {
		return false
	}
	if (enchant.EnchantType == proto.EnchantType_EnchantTypeOffHand) != (item.WeaponType == proto.WeaponType_WeaponTypeOffHand) {
		return false
	}
	if enchant.Type == proto.ItemType_ItemTypeRanged {
		if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeBow && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeCrossbow && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeGun {
			return false
		}
	} else if item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown && item.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeWand {
		return false
	}
	return core.CheckSliceOverlap(getEligibleEnchantSlots(enchant), core.EligibleSlotsForItem(&item))
}

func getEligibleEnchantSlots(enchant core.Enchant) []proto.ItemSlot {
	types := append([]proto.ItemType{enchant.Type}, enchant.ExtraTypes...)
	slots := make([]proto.ItemSlot, 0, len(types)*2)
	for _, itemType := range types {
		if typeSlots, ok := core.ItemTypeToSlotsMap[itemType]; ok {
			slots = append(slots, typeSlots...)
			continue
		}
		if itemType == proto.ItemType_ItemTypeWeapon {
			slots = append(slots, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
		}
	}
	return slots
}

func canEquipItem(item core.Item, playerClass proto.Class, playerSpec proto.Spec, slot proto.ItemSlot) bool {
	if item.Type == proto.ItemType_ItemTypeFinger || item.Type == proto.ItemType_ItemTypeTrinket {
		return true
	}
	if item.Type == proto.ItemType_ItemTypeWeapon {
		eligibleWeaponTypes := core.ClassWeaponTypeCapabilities[playerClass]
		eligibleWeaponType, ok := eligibleWeaponTypes[item.WeaponType]
		if !ok {
			return false
		}
		if (item.HandType == proto.HandType_HandTypeOffHand || (item.HandType == proto.HandType_HandTypeOneHand && slot == proto.ItemSlot_ItemSlotOffHand)) && item.WeaponType != proto.WeaponType_WeaponTypeShield && item.WeaponType != proto.WeaponType_WeaponTypeOffHand && !core.SpecCanDualWieldCapabilities[playerSpec] {
			return false
		}
		if item.HandType == proto.HandType_HandTypeTwoHand && !eligibleWeaponType.CanUseTwoHand {
			return false
		}
		if item.HandType == proto.HandType_HandTypeTwoHand && slot == proto.ItemSlot_ItemSlotOffHand {
			return false
		}
		return true
	}
	if item.Type == proto.ItemType_ItemTypeRanged {
		return slices.Contains(core.ClassRangedWeaponTypeCapabilities[playerClass], item.RangedWeaponType)
	}
	classArmorTypes := core.ClassArmorTypeCapabilities[playerClass]
	if len(classArmorTypes) == 0 {
		return false
	}
	maxArmorType := classArmorTypes[0]
	return maxArmorType >= item.ArmorType
}

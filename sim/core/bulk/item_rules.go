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
	} else if !core.EnchantAppliesToItem(itemSpec.GetEnchant(), option.item) {
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

// Unique-Equipped: two copies of a unique item, or two items of one limit category, cannot be worn together.
func excludesEachOther(item core.Item, other core.Item) bool {
	return (item.Unique && item.ID == other.ID) || (item.LimitCategory != 0 && item.LimitCategory == other.LimitCategory)
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

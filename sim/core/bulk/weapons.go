package bulk

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Swaps the Dense sharpening stone and weightstone to match the weapon, 0 when neither fits it;
// any other imbue passes through.
func adjustWeaponImbueID(imbueID int32, weapon *proto.ItemSpec) int32 {
	if imbueID != core.DenseSharpeningStoneID && imbueID != core.DenseWeightstoneID {
		return imbueID
	}
	if weapon == nil || weapon.Id == 0 {
		return 0
	}
	item := core.GetItemByID(weapon.Id)
	if item == nil {
		return 0
	}
	if slices.Contains(core.SharpWeaponTypes, item.WeaponType) {
		return core.DenseSharpeningStoneID
	}
	if slices.Contains(core.BluntWeaponTypes, item.WeaponType) {
		return core.DenseWeightstoneID
	}
	return 0
}

func adjustCandidateImbues(player *proto.Player) {
	consumables := player.GetConsumables()
	if consumables == nil {
		return
	}
	if consumables.MhImbueId == 0 && consumables.OhImbueId == 0 {
		return
	}
	items := player.GetEquipment().GetItems()
	var mhWeapon, ohWeapon *proto.ItemSpec
	if int(proto.ItemSlot_ItemSlotMainHand) < len(items) {
		mhWeapon = items[proto.ItemSlot_ItemSlotMainHand]
	}
	if int(proto.ItemSlot_ItemSlotOffHand) < len(items) {
		ohWeapon = items[proto.ItemSlot_ItemSlotOffHand]
	}
	consumables.MhImbueId = adjustWeaponImbueID(consumables.MhImbueId, mhWeapon)
	consumables.OhImbueId = adjustWeaponImbueID(consumables.OhImbueId, ohWeapon)
}

func (generator *bulkSimCandidateGenerator) getAllWeaponCombos() [][2]*bulkSimCandidateOption {
	allWeaponCombos := make([][2]*bulkSimCandidateOption, 0)
	all2HWeapons := make([]bulkSimCandidateOption, 0)
	for _, bulkSlot := range []BulkSimItemSlot{BulkSimItemSlotMainHand, BulkSimItemSlotHandWeapon} {
		for _, option := range generator.selectedByBulkSlot[bulkSlot] {
			if option.item.HandType == proto.HandType_HandTypeTwoHand {
				all2HWeapons = append(all2HWeapons, option)
			}
		}
	}

	for i := range all2HWeapons {
		allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&all2HWeapons[i], nil})
	}

	mhOptions := generator.selectedByBulkSlot[BulkSimItemSlotMainHand]
	ohOptions := generator.selectedByBulkSlot[BulkSimItemSlotOffHand]
	if len(mhOptions) > 0 {
		for i := range mhOptions {
			if optionsContainEquivalent(all2HWeapons, mhOptions[i]) {
				continue
			}
			if len(ohOptions) > 0 {
				for j := range ohOptions {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&mhOptions[i], &ohOptions[j]})
				}
			} else {
				allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&mhOptions[i], nil})
			}
		}
	} else if len(ohOptions) > 0 {
		for i := range ohOptions {
			allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{nil, &ohOptions[i]})
		}
	}

	oneHandOptions := generator.selectedByBulkSlot[BulkSimItemSlotHandWeapon]
	if len(oneHandOptions) > 0 {
		unique := make([]bulkSimCandidateOption, 0, len(oneHandOptions))
		for _, option := range oneHandOptions {
			if optionsContainEquivalent(all2HWeapons, option) {
				continue
			}
			if optionsContainEquivalent(unique, option) {
				continue
			}
			unique = append(unique, option)
		}

		for i := range unique {
			iCanMH := unique[i].item.HandType != proto.HandType_HandTypeOffHand
			iCanOH := unique[i].item.HandType != proto.HandType_HandTypeMainHand
			// Only wield the same 1H weapon in both hands when two copies exist and the weapon
			// itself allows it (not unique, no limit category).
			twoCopies := generator.hasTwoCopies(unique[i], unique[i].item)
			if twoCopies && iCanMH && iCanOH {
				allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[i], &unique[i]})
			}
			for j := i + 1; j < len(unique); j++ {
				if !twoCopies && candidateOptionsShareItem(unique[i], unique[j]) {
					continue
				}
				jCanMH := unique[j].item.HandType != proto.HandType_HandTypeOffHand
				jCanOH := unique[j].item.HandType != proto.HandType_HandTypeMainHand
				if iCanMH && jCanOH {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[i], &unique[j]})
				}
				if jCanMH && iCanOH {
					allWeaponCombos = append(allWeaponCombos, [2]*bulkSimCandidateOption{&unique[j], &unique[i]})
				}
			}
		}
	}

	filteredCombos := make([][2]*bulkSimCandidateOption, 0, len(allWeaponCombos))
	for _, combo := range allWeaponCombos {
		if weaponHandsAllowed(combo[0], combo[1]) && generator.weaponComboMatchesSettings(combo[0], combo[1]) {
			filteredCombos = append(filteredCombos, combo)
		}
	}

	return filteredCombos
}

func (generator *bulkSimCandidateGenerator) matchesWeaponTypeFilter(option *bulkSimCandidateOption, slot proto.ItemSlot) bool {
	filter := generator.weaponTypeFilters[slot]
	if len(filter) == 0 {
		return true
	}
	if option == nil {
		return false
	}
	return option.item.WeaponType > proto.WeaponType_WeaponTypeUnknown && slices.Contains(filter, option.item.WeaponType)
}

func (generator *bulkSimCandidateGenerator) weaponComboMatchesSettings(mhItem *bulkSimCandidateOption, ohItem *bulkSimCandidateOption) bool {
	frozenWeaponItem := generator.frozenWeaponItem
	if generator.frozenWeaponSlot == proto.ItemSlot_ItemSlotMainHand && frozenWeaponItem != nil && !candidateOptionEqualsItemPtr(mhItem, frozenWeaponItem) {
		return false
	}
	if generator.frozenWeaponSlot == proto.ItemSlot_ItemSlotOffHand && frozenWeaponItem != nil && !candidateOptionEqualsItemPtr(ohItem, frozenWeaponItem) {
		return false
	}
	return generator.matchesWeaponTypeFilter(mhItem, proto.ItemSlot_ItemSlotMainHand) && generator.matchesWeaponTypeFilter(ohItem, proto.ItemSlot_ItemSlotOffHand)
}

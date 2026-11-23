package bulk

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// The selected items worn in no slot, in request order. A copy of a worn item, under any enchant,
// is no new item.
func (generator *bulkSimCandidateGenerator) newItemOptions() ([]bulkSimCandidateOption, error) {
	worn := make(map[itemSpecCacheKey]bool, core.NumItemSlots)
	for _, item := range generator.baseEquipment {
		if item.ID != 0 {
			worn[buildItemSpecKey(item.ToItemSpecProto())] = true
		}
	}

	var options []bulkSimCandidateOption
	for _, selectedItem := range generator.settings.GetItems() {
		if selectedItem == nil || selectedItem.GetId() == 0 || worn[buildItemSpecKey(selectedItem)] {
			continue
		}
		option, err := newSelectedItemOption(selectedItem)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, nil
}

// The other slot of a ring, trinket or hand pair.
func pairedSlot(slot proto.ItemSlot) proto.ItemSlot {
	pair := BulkSimItemSlotToItemSlotPairs[getBulkItemSlotFromSlot(slot, true)]
	if pair[0] == slot {
		return pair[1]
	}
	return pair[0]
}

// One gear set per placement of a new item: each slot it fits, crossed with that slot's pool and, for a
// hand, with the other hand's partners. Everything else stays as worn, and a set that repeats the baseline
// or an earlier set is dropped.
func (generator *bulkSimCandidateGenerator) singleChangeGearSets() ([]core.EquipmentSpec, error) {
	options, err := generator.newItemOptions()
	if err != nil {
		return nil, err
	}
	pools := generator.playerEnchantPools()
	base := generator.baseGearSpec()

	var sets []core.EquipmentSpec
	seen := map[core.EquipmentSpec]bool{base: true}
	add := func(gear core.EquipmentSpec) {
		if !seen[gear] {
			seen[gear] = true
			sets = append(sets, gear)
		}
	}

	for _, option := range options {
		for _, slot := range core.EligibleSlotsForItem(&option.item) {
			if !canEquipItem(option.item, generator.playerClass, generator.playerSpec, slot) {
				continue
			}
			// With dual wield off each hand reads its own row, and either finger the Finger1 row.
			for _, variant := range enchantVariants(option, pools.poolFor(getBulkItemSlotFromSlot(slot, false))) {
				gear := base
				gear[slot] = generator.slotSpec(slot, generator.resolveItem(slot, variant))
				switch slot {
				case proto.ItemSlot_ItemSlotMainHand:
					for _, partner := range generator.offHandPartners(base, option.item) {
						gear[proto.ItemSlot_ItemSlotOffHand] = partner
						add(gear)
					}
					continue
				case proto.ItemSlot_ItemSlotOffHand:
					for _, partner := range generator.mainHandPartners(base, option.item) {
						gear[proto.ItemSlot_ItemSlotMainHand] = partner
						add(gear)
					}
					continue
				case proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2, proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2:
					if excludesEachOther(option.item, generator.baseEquipment[pairedSlot(slot)]) {
						continue
					}
				}
				add(gear)
			}
		}
	}
	return sets, nil
}

// What goes in the off hand beside a new main-hand item. A two-hander leaves it empty.
func (generator *bulkSimCandidateGenerator) offHandPartners(base core.EquipmentSpec, newItem core.Item) []core.ItemSpec {
	if newItem.HandType == proto.HandType_HandTypeTwoHand {
		return []core.ItemSpec{{}}
	}
	return generator.handPartners(base, newItem, proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand)
}

// What goes in the main hand beside a new off-hand item. An off-hand item under a worn two-hander
// has none: it would need the two-hander taken off and nothing in the main hand.
func (generator *bulkSimCandidateGenerator) mainHandPartners(base core.EquipmentSpec, newItem core.Item) []core.ItemSpec {
	if generator.baseEquipment[proto.ItemSlot_ItemSlotMainHand].HandType == proto.HandType_HandTypeTwoHand {
		return nil
	}
	return generator.handPartners(base, newItem, proto.ItemSlot_ItemSlotOffHand, proto.ItemSlot_ItemSlotMainHand)
}

// The other hand as worn, then for a dual wielder the one-hander worn in the new item's hand moved
// across. A moved weapon keeps its own enchant: only the new item takes the slot's carry-over.
func (generator *bulkSimCandidateGenerator) handPartners(base core.EquipmentSpec, newItem core.Item, placedSlot proto.ItemSlot, partnerSlot proto.ItemSlot) []core.ItemSpec {
	var partners []core.ItemSpec
	if stays := base[partnerSlot]; stays.ID == 0 || !excludesEachOther(newItem, generator.baseEquipment[partnerSlot]) {
		partners = append(partners, stays)
	}
	moved := generator.baseEquipment[placedSlot]
	if core.SpecCanDualWieldCapabilities[generator.playerSpec] && moved.ID != 0 && moved.HandType == proto.HandType_HandTypeOneHand &&
		canEquipItem(moved, generator.playerClass, generator.playerSpec, partnerSlot) && !excludesEachOther(newItem, moved) {
		partners = append(partners, base[placedSlot])
	}
	return partners
}

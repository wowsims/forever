package bulk

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// enchantPools holds the pooled effect ids per item slot, deduped in request order.
type enchantPools map[proto.ItemSlot][]int32

func newEnchantPools(settings *proto.BulkSettings) enchantPools {
	pools := make(enchantPools)
	for _, row := range settings.GetEnchants() {
		for _, effectID := range row.GetEffectIds() {
			if !slices.Contains(pools[row.GetSlot()], effectID) {
				pools[row.GetSlot()] = append(pools[row.GetSlot()], effectID)
			}
		}
	}
	return pools
}

// poolFor reads the ring pool from the Finger1 row, which serves both fingers. Trinkets take no
// enchants and have no pool.
func (pools enchantPools) poolFor(bulkSlot BulkSimItemSlot) []int32 {
	if bulkSlot == BulkSimItemSlotFinger {
		return pools[proto.ItemSlot_ItemSlotFinger1]
	}
	if slot, ok := BulkSimItemSlotToSingleItemSlot[bulkSlot]; ok {
		return pools[slot]
	}
	return nil
}

func fittingEnchants(pool []int32, item core.Item) []int32 {
	return core.FilterSlice(pool, func(effectID int32) bool {
		return enchantAppliesToItem(effectID, item)
	})
}

func withEnchant(option bulkSimCandidateOption, effectID int32, hand weaponHand) bulkSimCandidateOption {
	option.enchant = effectID
	option.hand = hand
	return option
}

func enchantVariants(option bulkSimCandidateOption, pool []int32) []bulkSimCandidateOption {
	fitting := fittingEnchants(pool, option.item)
	if len(fitting) == 0 {
		return []bulkSimCandidateOption{option}
	}
	variants := make([]bulkSimCandidateOption, 0, len(fitting))
	for _, effectID := range fitting {
		variants = append(variants, withEnchant(option, effectID, eitherHand))
	}
	return variants
}

// handWeaponEnchantVariants expands a weapon a dual-wielder shares between both hands. Each
// variant is limited to the hand(s) whose pool holds its enchant; a hand whose pool fits nothing
// gets the unexpanded option, which keeps the carry-over.
func handWeaponEnchantVariants(option bulkSimCandidateOption, mainHandPool []int32, offHandPool []int32) []bulkSimCandidateOption {
	canMainHand := option.item.HandType != proto.HandType_HandTypeOffHand
	canOffHand := option.item.HandType != proto.HandType_HandTypeMainHand && option.item.HandType != proto.HandType_HandTypeTwoHand
	var mainHandFits, offHandFits []int32
	if canMainHand {
		mainHandFits = fittingEnchants(mainHandPool, option.item)
	}
	if canOffHand {
		offHandFits = fittingEnchants(offHandPool, option.item)
	}
	if len(mainHandFits) == 0 && len(offHandFits) == 0 {
		return []bulkSimCandidateOption{option}
	}

	variants := make([]bulkSimCandidateOption, 0, len(mainHandFits)+len(offHandFits)+1)
	for _, effectID := range mainHandFits {
		hand := mainHandOnly
		if slices.Contains(offHandFits, effectID) {
			hand = eitherHand
		}
		variants = append(variants, withEnchant(option, effectID, hand))
	}
	for _, effectID := range offHandFits {
		if !slices.Contains(mainHandFits, effectID) {
			variants = append(variants, withEnchant(option, effectID, offHandOnly))
		}
	}
	if canMainHand && len(mainHandFits) == 0 {
		option.hand = mainHandOnly
		variants = append(variants, option)
	}
	if canOffHand && len(offHandFits) == 0 {
		option.hand = offHandOnly
		variants = append(variants, option)
	}
	return variants
}

// initEnchantVariants expands every option in a bulk slot with a pool. The frozen weapon keeps
// its own enchant; frozen rings and trinkets are paired from frozenItems, not from these options.
func (generator *bulkSimCandidateGenerator) initEnchantVariants() {
	pools := newEnchantPools(generator.settings)
	if len(pools) == 0 {
		return
	}
	mainHandPool := pools[proto.ItemSlot_ItemSlotMainHand]
	offHandPool := pools[proto.ItemSlot_ItemSlotOffHand]

	for bulkSlot, options := range generator.selectedByBulkSlot {
		expanded := make([]bulkSimCandidateOption, 0, len(options))
		for _, option := range options {
			switch {
			case generator.isFrozenWeapon(option):
				expanded = append(expanded, option)
			case bulkSlot == BulkSimItemSlotHandWeapon:
				expanded = append(expanded, handWeaponEnchantVariants(option, mainHandPool, offHandPool)...)
			default:
				expanded = append(expanded, enchantVariants(option, pools.poolFor(bulkSlot))...)
			}
		}
		generator.selectedByBulkSlot[bulkSlot] = expanded
	}
}

func (generator *bulkSimCandidateGenerator) isFrozenWeapon(option bulkSimCandidateOption) bool {
	return generator.frozenWeaponItem != nil && candidateOptionEqualsItem(option, *generator.frozenWeaponItem)
}

func weaponHandsAllowed(mainHand *bulkSimCandidateOption, offHand *bulkSimCandidateOption) bool {
	return (mainHand == nil || mainHand.hand != offHandOnly) && (offHand == nil || offHand.hand != mainHandOnly)
}

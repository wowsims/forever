import { type ItemSlot, ItemSpec } from '@generated/proto/common';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, getBulkItemSlotFromSlot } from '@sim/bulk/utils';

// The rings, the trinkets and a dual wielder's weapons: two slots whose items can change places.
export const getSwappableItemSlotPair = (slot: ItemSlot, canDualWield: boolean): [ItemSlot, ItemSlot] | undefined =>
	BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(getBulkItemSlotFromSlot(slot, canDualWield));

export const itemSpecPairsEqualUnordered = (
	resultItems: readonly ItemSpec[],
	originalItems: readonly ItemSpec[],
	[slot1, slot2]: [ItemSlot, ItemSlot],
): boolean =>
	(ItemSpec.equals(resultItems[slot1], originalItems[slot1]) && ItemSpec.equals(resultItems[slot2], originalItems[slot2])) ||
	(ItemSpec.equals(resultItems[slot1], originalItems[slot2]) && ItemSpec.equals(resultItems[slot2], originalItems[slot1]));

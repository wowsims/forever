import type { GearData } from '@features/gear/types';
import { type ItemSlot, ItemSpec } from '@generated/proto/common';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, BULK_SIM_ITEM_SLOT_TO_SINGLE_ITEM_SLOT, type BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState } from '@sim/settings/bulk_settings';
import { Emitter } from '@sim/state/events';

import { addBulkItem, updateBulkItem } from './items';
import { pickerEntryAt } from './picker_groups';

/**
 * The selector modal's seam for a batch entry: equipping writes back into the batch instead of onto
 * the player, and the item is read out of the group each time, so a modal left open on an entry
 * that has just been replaced shows the replacement.
 */
export const createBulkGearData = (player: Player<any>, bulkSlot: BulkSimItemSlot, index: number): GearData => {
	const changeEvent = new Emitter<void>();
	return {
		equipItem: (newItem: EquippedItem | null) => {
			if (!newItem) return;
			updateBulkItem(player, index, newItem.asSpec());
			changeEvent.emit();
		},
		getEquippedItem: () => pickerEntryAt(bulkState(player).pickerGroups.get(bulkSlot) ?? [], index)?.item ?? null,
		subscribe: onChange => changeEvent.on(onChange),
	};
};

/** The item slot whose list a group offers: the first slot of a paired group, the main hand for a dual-wielder's weapons. */
export const bulkGroupItemSlot = (bulkSlot: BulkSimItemSlot): ItemSlot =>
	BULK_SIM_ITEM_SLOT_TO_SINGLE_ITEM_SLOT.get(bulkSlot) ?? BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(bulkSlot)![0];

/**
 * The selector modal's seam for adding to a group: a pick adds the item to the batch instead of
 * replacing anything, so nothing reads as equipped and each pick is one more entry. The spec is the
 * bare id, as the list's batch toggle writes it, so the toggle's in-batch match finds it.
 */
export const createBulkAddGearData = (player: Player<any>): GearData => ({
	equipItem: (newItem: EquippedItem | null) => {
		if (newItem) addBulkItem(player, ItemSpec.create({ id: newItem.id }));
	},
	getEquippedItem: () => null,
	subscribe: () => () => undefined,
	addOnly: true,
});

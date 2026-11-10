import { type ConsumesSpec, ItemSlot, type Spec } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { batch } from '@sim/state/batch';

type ImbueField = 'mhImbueId' | 'ohImbueId' | 'mhWeaponImbueId' | 'ohWeaponImbueId';
type SwapImbueField = 'swapMhImbueId' | 'swapOhImbueId' | 'swapMhWeaponImbueId' | 'swapOhWeaponImbueId';

const IMBUE_PAIRS: Partial<Record<ItemSlot, ReadonlyArray<[ImbueField, SwapImbueField]>>> = {
	[ItemSlot.ItemSlotMainHand]: [
		['mhImbueId', 'swapMhImbueId'],
		['mhWeaponImbueId', 'swapMhWeaponImbueId'],
	],
	[ItemSlot.ItemSlotOffHand]: [
		['ohImbueId', 'swapOhImbueId'],
		['ohWeaponImbueId', 'swapOhWeaponImbueId'],
	],
};

/** Exchanges the equipped item and the swap item in each of `itemSlots`, with a weapon's imbues, in one batch so the writes land as a single notification. */
export const swapWithGear = <SpecType extends Spec>(player: Player<SpecType>, itemSlots: ReadonlyArray<ItemSlot>): void => {
	let newGear = player.getGear();
	let newSwapGear = player.itemSwapSettings.getGear();
	const newConsumes: ConsumesSpec = player.getConsumes();

	for (const slot of itemSlots) {
		const gearItem = player.getGear().getEquippedItem(slot);
		const swapItem = player.itemSwapSettings.getGear().getEquippedItem(slot);
		newGear = newGear.withEquippedItem(slot, swapItem);
		newSwapGear = newSwapGear.withEquippedItem(slot, gearItem);
		for (const [field, swapField] of IMBUE_PAIRS[slot] ?? []) {
			[newConsumes[field], newConsumes[swapField]] = [newConsumes[swapField], newConsumes[field]];
		}
	}

	batch(() => {
		player.setGear(newGear);
		player.itemSwapSettings.setGear(newSwapGear);
		player.setConsumes(newConsumes);
	});
};

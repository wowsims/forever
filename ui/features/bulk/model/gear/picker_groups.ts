// `@sim/bulk/utils` reaches `@i18n/entity_mapping`, which reads `BulkSimItemSlot` at module
// scope, so importing the enum through it is a cycle. `constants_auto_gen` depends on the generated
// protos and nothing else.
import { HandType, type ItemSlot } from '@generated/proto/common';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import type { BulkPickerEntry } from '@sim/bulk/types';
import type { EquippedItem } from '@sim/proto/equipped_item';

// Whether both of this bulk slot's physical slots can hold the same item at once, which is what
// makes a same-item combo (two identical rings, one weapon in each hand) a valid input. Mirrors
// canStackTwoCopies in the backend: accepting a pair it refuses inflates the combination count
// the Simulate button is gated on and generates duplicate candidates.
const canStackTwoCopies = (bulkSlot: BulkSimItemSlot, item: EquippedItem): boolean => {
	if (item._item.unique || item._item.limitCategory !== 0) return false;
	switch (bulkSlot) {
		case BulkSimItemSlot.ItemSlotFinger:
		case BulkSimItemSlot.ItemSlotTrinket:
			return true;
		case BulkSimItemSlot.ItemSlotHandWeapon:
			return item._item.handType === HandType.HandTypeOneHand;
		default:
			return false;
	}
};

// True when the slot already holds as many copies of this exact item as can be worn.
// Items sharing a limit category are NOT rejected: only one of them can be worn at a time, but
// listing several is how you compare them (Evil Eye of Galakras ships six ilvl tiers, all
// category 326), and the candidate generator drops the conflicting pairings itself.
const isDuplicateOfExisting = (bulkSlot: BulkSimItemSlot, entries: readonly BulkPickerEntry[], item: EquippedItem): boolean => {
	const maxCopies = canStackTwoCopies(bulkSlot, item) ? 2 : 1;
	return entries.filter(entry => entry.item.id === item.id).length >= maxCopies;
};

/**
 * The group's batch indices made redundant by `equippedItem` being worn.
 *
 * An equipped item is already part of every candidate, so a batch entry for the same item adds
 * nothing and renders as a phantom duplicate. Mirrors the backend, which drops the user-added
 * copy in initSelectedItems — except where equipped + added is what makes a same-item-in-both-
 * slots combo possible.
 */
export const redundantAddedIndices = (bulkSlot: BulkSimItemSlot, entries: readonly BulkPickerEntry[], equippedItem: EquippedItem): number[] =>
	canStackTwoCopies(bulkSlot, equippedItem) ? [] : entries.filter(entry => entry.index >= 0 && entry.item.id === equippedItem.id).map(entry => entry.index);

export const pickerEntryAt = (entries: readonly BulkPickerEntry[], index: number): BulkPickerEntry | null =>
	entries.find(entry => entry.index === index) ?? null;

export const removePickerEntry = (entries: readonly BulkPickerEntry[], index: number): readonly BulkPickerEntry[] =>
	entries.filter(entry => entry.index !== index);

/**
 * One slot group's entries with `item` at `index`, in render order, or `'duplicate'`.
 *
 * Equipped entries are unshifted and user entries pushed: the equipped pair ends up reversed
 * against slot order, and the batch's own items follow in index order. Adding an equipped entry
 * also drops the batch entries it makes redundant.
 */
export const addPickerEntry = (
	bulkSlot: BulkSimItemSlot,
	entries: readonly BulkPickerEntry[],
	index: number,
	item: EquippedItem,
): readonly BulkPickerEntry[] | 'duplicate' => {
	// Equipped entries (index < 0) report what is worn rather than offering a choice, so they
	// always render — the guard must never hide one. They evict redundant batch entries instead.
	if (index < 0) {
		const redundant = redundantAddedIndices(bulkSlot, entries, item);
		return [{ index, item }, ...removePickerEntry(entries, index).filter(entry => !redundant.includes(entry.index))];
	}

	if (isDuplicateOfExisting(bulkSlot, entries, item)) return 'duplicate';

	return [...removePickerEntry(entries, index), { index, item }];
};

/** Null when nothing is at `index`, so the caller can raise its own notice. */
export const updatePickerEntry = (entries: readonly BulkPickerEntry[], index: number, item: EquippedItem): readonly BulkPickerEntry[] | null =>
	pickerEntryAt(entries, index) ? entries.map(entry => (entry.index === index ? { index, item } : entry)) : null;

/**
 * Which of the group's two physical slots holds `frozenItem`, or null.
 *
 * Identity first, then value: a frozen item is captured off the gear it was frozen from, and an
 * equivalent piece moved into the other slot should still read as frozen.
 */
export const frozenItemSlot = (
	gear: { getEquippedItem: (slot: ItemSlot) => EquippedItem | null },
	slots: readonly ItemSlot[] | undefined,
	frozenItem: EquippedItem | null | undefined,
): ItemSlot | null => {
	if (!frozenItem || !slots) return null;
	return slots.find(slot => gear.getEquippedItem(slot) === frozenItem) ?? slots.find(slot => gear.getEquippedItem(slot)?.equals(frozenItem)) ?? null;
};

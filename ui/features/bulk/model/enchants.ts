import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkEnchantPreset, BulkEnchantPresets } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeEnchantPresets } from '@sim/settings/bulk_settings';

export type BulkEnchantPresetNameProblem = 'empty' | 'duplicate';

// Finger1 stands for both fingers. Main Hand holds the weapon enchants, which go to both hands; Off Hand
// holds the shield and held-in-off-hand enchants.
export const BULK_ENCHANT_SLOTS: readonly ItemSlot[] = [
	ItemSlot.ItemSlotHead,
	ItemSlot.ItemSlotShoulder,
	ItemSlot.ItemSlotBack,
	ItemSlot.ItemSlotChest,
	ItemSlot.ItemSlotWrist,
	ItemSlot.ItemSlotHands,
	ItemSlot.ItemSlotLegs,
	ItemSlot.ItemSlotFeet,
	ItemSlot.ItemSlotFinger1,
	ItemSlot.ItemSlotMainHand,
	ItemSlot.ItemSlotOffHand,
	ItemSlot.ItemSlotRanged,
];

// Sorted, so the request that seeds the batch's RNG does not depend on preset order or the order enchants were picked in.
export const bulkEnchantPool = (player: Player<any>): BulkEnchantSlotOptions[] => {
	const checked = bulkState(player).enchantPresets.filter(preset => preset.checked);

	const sources = (slot: ItemSlot) => (slot === ItemSlot.ItemSlotOffHand ? [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand] : [slot]);

	return BULK_ENCHANT_SLOTS.flatMap(slot => {
		const wanted = new Set(
			checked.flatMap(preset => preset.slots.filter(options => sources(slot).includes(options.slot)).flatMap(options => options.effectIds)),
		);
		if (!wanted.size) return [];
		const usable = new Set(player.getEnchants(slot).map(enchant => enchant.effectId));
		const effectIds = [...wanted].filter(id => usable.has(id)).sort((a, b) => a - b);
		return effectIds.length ? [BulkEnchantSlotOptions.create({ slot, effectIds })] : [];
	});
};

export const validateBulkEnchantPresetName = (
	presets: readonly BulkEnchantPreset[],
	name: string,
	previousName?: string,
): BulkEnchantPresetNameProblem | null => {
	const trimmed = name.trim();
	if (!trimmed) return 'empty';
	if (presets.some(preset => preset.name === trimmed && preset.name !== previousName)) return 'duplicate';
	return null;
};

const writeBulkEnchantPresets = (player: Player<any>, enchantPresets: BulkEnchantPreset[], poolChanged: boolean) => {
	patchBulkState(player, { enchantPresets }, poolChanged ? ['settings'] : []);
	storeEnchantPresets(player, BulkEnchantPresets.create({ presets: enchantPresets }));
};

export interface BulkEnchantPresetDraft {
	name: string;
	slots: ReadonlyArray<{ slot: ItemSlot; effectIds: readonly number[] }>;
}

export const saveBulkEnchantPreset = (player: Player<any>, preset: BulkEnchantPresetDraft, previousName?: string) => {
	const { enchantPresets } = bulkState(player);
	const replaced = previousName === undefined ? undefined : enchantPresets.find(existing => existing.name === previousName);
	const saved = BulkEnchantPreset.create({
		name: preset.name.trim(),
		slots: preset.slots
			.filter(({ effectIds }) => effectIds.length)
			.map(({ slot, effectIds }) => BulkEnchantSlotOptions.create({ slot, effectIds: [...effectIds] })),
		checked: replaced?.checked ?? false,
	});

	writeBulkEnchantPresets(
		player,
		replaced ? enchantPresets.map(existing => (existing === replaced ? saved : existing)) : [...enchantPresets, saved],
		saved.checked,
	);
};

export const deleteBulkEnchantPreset = (player: Player<any>, name: string) => {
	const { enchantPresets } = bulkState(player);
	const deleted = enchantPresets.find(preset => preset.name === name);
	if (!deleted) return;

	writeBulkEnchantPresets(
		player,
		enchantPresets.filter(preset => preset !== deleted),
		deleted.checked,
	);
};

export const setBulkEnchantPresetChecked = (player: Player<any>, name: string, checked: boolean) => {
	const { enchantPresets } = bulkState(player);
	const target = enchantPresets.find(preset => preset.name === name);
	if (!target || target.checked === checked) return;

	writeBulkEnchantPresets(
		player,
		enchantPresets.map(preset => (preset === target ? { ...preset, checked } : preset)),
		true,
	);
};

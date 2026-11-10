import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkEnchantPreset, BulkEnchantPresets, BulkEnchantPresetSelection } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeBulkBuildEnchantPresets, storeEnchantPresets } from '@sim/settings/bulk_settings';

export type EnchantPresetNameProblem = 'empty' | 'duplicate';

// Finger1 stands for both fingers. Main Hand holds the weapon enchants, which go to both hands; Off Hand
// holds the shield and held-in-off-hand enchants.
export const ENCHANT_PRESET_SLOTS: readonly ItemSlot[] = [
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

export const enchantPresetPoolSlot = (slot: ItemSlot): ItemSlot => (slot === ItemSlot.ItemSlotFinger2 ? ItemSlot.ItemSlotFinger1 : slot);

// The effect ids the presets hold for an item slot, each once and sorted. The Off Hand also takes the Main Hand's, which
// go to both hands.
export const presetSlotEffectIds = (presets: readonly BulkEnchantPreset[], slot: ItemSlot): number[] => {
	const poolSlot = enchantPresetPoolSlot(slot);
	const sources = poolSlot === ItemSlot.ItemSlotOffHand ? [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand] : [poolSlot];
	const effectIds = presets.flatMap(preset => preset.slots.filter(options => sources.includes(options.slot)).flatMap(options => options.effectIds));
	return [...new Set(effectIds)].sort((a, b) => a - b);
};

// The presets' effect ids for the slot that the player can use there.
const enchantPresetSlotIds = (player: Player<any>, presets: readonly BulkEnchantPreset[], slot: ItemSlot): number[] => {
	const wanted = presetSlotEffectIds(presets, slot);
	if (!wanted.length) return [];
	const usable = new Set(player.getEnchants(slot).map(enchant => enchant.effectId));
	return wanted.filter(id => usable.has(id));
};

// Sorted, so the request that seeds the batch's RNG does not depend on preset order or the order enchants were picked in.
export const enchantPresetPool = (player: Player<any>): BulkEnchantSlotOptions[] => {
	const checked = bulkState(player).enchantPresets.filter(preset => preset.checked);

	return ENCHANT_PRESET_SLOTS.flatMap(slot => {
		const effectIds = enchantPresetSlotIds(player, checked, slot);
		return effectIds.length ? [BulkEnchantSlotOptions.create({ slot, effectIds })] : [];
	});
};

export const validateEnchantPresetName = (presets: readonly BulkEnchantPreset[], name: string, previousName?: string): EnchantPresetNameProblem | null => {
	const trimmed = name.trim();
	if (!trimmed) return 'empty';
	if (presets.some(preset => preset.name === trimmed && preset.name !== previousName)) return 'duplicate';
	return null;
};

const writeEnchantPresets = (player: Player<any>, enchantPresets: BulkEnchantPreset[], poolChanged: boolean) => {
	patchBulkState(player, { enchantPresets }, poolChanged ? ['settings'] : []);
	storeEnchantPresets(player, BulkEnchantPresets.create({ presets: enchantPresets }));
};

// The Builds pick holds preset names, so a rename carries a picked preset over and a delete drops it.
const writeBuildEnchantPicks = (player: Player<any>, buildEnchantPresets: string[]) => {
	patchBulkState(player, { buildEnchantPresets });
	storeBulkBuildEnchantPresets(player, BulkEnchantPresetSelection.create({ names: buildEnchantPresets }));
};

export interface EnchantPresetDraft {
	name: string;
	slots: ReadonlyArray<{ slot: ItemSlot; effectIds: readonly number[] }>;
}

export const saveEnchantPreset = (player: Player<any>, preset: EnchantPresetDraft, previousName?: string) => {
	const { enchantPresets, buildEnchantPresets } = bulkState(player);
	const replaced = previousName === undefined ? undefined : enchantPresets.find(existing => existing.name === previousName);
	const saved = BulkEnchantPreset.create({
		name: preset.name.trim(),
		slots: preset.slots
			.filter(({ effectIds }) => effectIds.length)
			.map(({ slot, effectIds }) => BulkEnchantSlotOptions.create({ slot, effectIds: [...effectIds] })),
		checked: replaced?.checked ?? false,
	});

	writeEnchantPresets(
		player,
		replaced ? enchantPresets.map(existing => (existing === replaced ? saved : existing)) : [...enchantPresets, saved],
		saved.checked,
	);
	if (replaced && replaced.name !== saved.name && buildEnchantPresets.includes(replaced.name)) {
		writeBuildEnchantPicks(
			player,
			buildEnchantPresets.map(picked => (picked === replaced.name ? saved.name : picked)),
		);
	}
};

export const deleteEnchantPreset = (player: Player<any>, name: string) => {
	const { enchantPresets, buildEnchantPresets } = bulkState(player);
	const deleted = enchantPresets.find(preset => preset.name === name);
	if (!deleted) return;

	writeEnchantPresets(
		player,
		enchantPresets.filter(preset => preset !== deleted),
		deleted.checked,
	);
	if (buildEnchantPresets.includes(name)) {
		writeBuildEnchantPicks(
			player,
			buildEnchantPresets.filter(picked => picked !== name),
		);
	}
};

export const setEnchantPresetChecked = (player: Player<any>, name: string, checked: boolean) => {
	const { enchantPresets } = bulkState(player);
	const target = enchantPresets.find(preset => preset.name === name);
	if (!target || target.checked === checked) return;

	writeEnchantPresets(
		player,
		enchantPresets.map(preset => (preset === target ? { ...preset, checked } : preset)),
		true,
	);
};

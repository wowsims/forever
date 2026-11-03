import { type BulkEnchantPreset, BulkEnchantPresetSelection } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { bulkState, patchBulkState, storeBulkBuildEnchantPresets } from '@sim/settings/bulk_settings';

import { checkedRefsAfter, type SavedRef } from './saved_refs';

interface BulkEnchantPickEntry {
	/** Enchant presets are the user's own, so `preset` is always false. */
	ref: SavedRef;
	name: string;
	preset: BulkEnchantPreset;
}

export interface BulkEnchantPick {
	name: string;
	preset: BulkEnchantPreset;
}

const filledSlots = (preset: BulkEnchantPreset) => preset.slots.filter(options => options.effectIds.length > 0);

// The order a preset lists its slots and enchants in changes nothing it applies.
const enchantsKey = (preset: BulkEnchantPreset) =>
	JSON.stringify(
		filledSlots(preset)
			.map(({ slot, effectIds }) => [slot, [...effectIds].sort((a, b) => a - b)] as const)
			.sort(([a], [b]) => a - b),
	);

const refOf = (name: string): SavedRef => ({ name, preset: false });

export const bulkEnchantPickEntries = (presets: readonly BulkEnchantPreset[]): BulkEnchantPickEntry[] =>
	presets.map(preset => ({ ref: refOf(preset.name), name: preset.name, preset }));

// The rows the Enchant presets list shows, in list order. One with no enchants is left out; a checked one holding the
// same enchants as an earlier checked row is flagged.
export const bulkEnchantPickList = (entries: readonly BulkEnchantPickEntry[], checked: readonly string[]) => {
	const picked = new Set<string>();
	return entries
		.filter(entry => filledSlots(entry.preset).length > 0)
		.map(entry => {
			const isChecked = checked.includes(entry.name);
			const enchants = enchantsKey(entry.preset);
			const duplicate = isChecked && picked.has(enchants);
			if (isChecked) picked.add(enchants);
			return { ...entry, checked: isChecked, duplicate };
		});
};

// Every checked row the list neither leaves out nor flags.
export const bulkEnchantPickRuns = (entries: readonly BulkEnchantPickEntry[], checked: readonly string[]): BulkEnchantPick[] =>
	bulkEnchantPickList(entries, checked)
		.filter(item => item.checked && !item.duplicate)
		.map(({ name, preset }) => ({ name, preset }));

// Stored apart from `BulkEnchantPreset.checked`, the Gear mode's pick, so no `settings` bump. Names go in list order, and
// one no longer pickable goes.
export const setBulkEnchantPickChecked = (player: Player<any>, name: string, checked: boolean, pickable: readonly string[]) => {
	const stored = bulkState(player).buildEnchantPresets;
	const buildEnchantPresets = checkedRefsAfter(stored.map(refOf), refOf(name), checked, pickable.map(refOf)).map(ref => ref.name);
	patchBulkState(player, { buildEnchantPresets });
	storeBulkBuildEnchantPresets(player, BulkEnchantPresetSelection.create({ names: buildEnchantPresets }));
};

export const formatEnchantPickSummary = (preset: BulkEnchantPreset): string =>
	i18n.t('bulk_tab.settings.enchant_picks.slots', { count: filledSlots(preset).length });

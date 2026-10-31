import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkEnchantPreset } from '@generated/proto/ui';
import { bulkState, loadStoredBulkBuildEnchantPresets, patchBulkState } from '@sim/settings/bulk_settings';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { bulkEnchantPickEntries, bulkEnchantPickList, bulkEnchantPickRuns, formatEnchantPickSummary, setBulkEnchantPickChecked } from './enchant_picks';

vi.mock('@i18n/config', () => ({
	default: { t: (key: string, options?: { count?: number }) => (options?.count === undefined ? key : `${key}:${options.count}`) },
}));

const preset = (name: string, slots: Array<[ItemSlot, number[]]>, checked = false) =>
	BulkEnchantPreset.create({ name, slots: slots.map(([slot, effectIds]) => BulkEnchantSlotOptions.create({ slot, effectIds })), checked });

const RAID = preset('Raid', [
	[ItemSlot.ItemSlotHead, [2545, 2543]],
	[ItemSlot.ItemSlotChest, [1891]],
]);
const UNDEAD = preset('Undead', [[ItemSlot.ItemSlotMainHand, [1898]]]);
const EMPTY = preset('Empty', [[ItemSlot.ItemSlotBack, []]]);

describe('bulkEnchantPickList', () => {
	it('lists the enchant presets in their own order, leaving out one with no enchants', () => {
		const entries = bulkEnchantPickEntries([UNDEAD, EMPTY, RAID]);

		expect(entries.map(e => [e.name, e.ref.preset])).toEqual([
			['Undead', false],
			['Empty', false],
			['Raid', false],
		]);
		expect(bulkEnchantPickList(entries, []).map(item => item.name)).toEqual(['Undead', 'Raid']);
	});

	it('flags a checked preset holding the same enchants as an earlier checked one', () => {
		const copy = preset('Raid copy', [
			[ItemSlot.ItemSlotChest, [1891]],
			[ItemSlot.ItemSlotHead, [2543, 2545]],
			[ItemSlot.ItemSlotBack, []],
		]);
		const entries = bulkEnchantPickEntries([RAID, copy, UNDEAD]);

		expect(bulkEnchantPickList(entries, ['Raid copy', 'Raid']).map(item => [item.name, item.checked, item.duplicate])).toEqual([
			['Raid', true, false],
			['Raid copy', true, true],
			['Undead', false, false],
		]);
		expect(bulkEnchantPickRuns(entries, ['Raid', 'Raid copy']).map(run => run.name)).toEqual(['Raid']);
	});
});

describe('bulkEnchantPickRuns', () => {
	it('runs the checked presets alone, with no own-enchants entry', () => {
		const entries = bulkEnchantPickEntries([RAID, EMPTY, UNDEAD]);

		expect(bulkEnchantPickRuns(entries, [])).toEqual([]);
		expect(bulkEnchantPickRuns(entries, ['Undead', 'Empty', 'Raid'])).toEqual([
			{ name: 'Raid', preset: RAID },
			{ name: 'Undead', preset: UNDEAD },
		]);
	});
});

describe('setBulkEnchantPickChecked', () => {
	it('stores the picks by name without touching the Gear-mode checked flag or the settings counter', () => {
		const player = makeBulkPlayer();
		patchBulkState(player, {
			enchantPresets: [preset('Raid', [[ItemSlot.ItemSlotHead, [2545]]]), preset('Undead', [[ItemSlot.ItemSlotMainHand, [1898]]], true)],
		});
		const { settings } = bulkState(player).v;

		setBulkEnchantPickChecked(player, 'Raid', true, ['Raid', 'Undead']);

		expect(bulkState(player).enchantPresets.map(p => p.checked)).toEqual([false, true]);
		expect(bulkState(player).v.settings).toBe(settings);
		expect(bulkState(player).buildEnchantPresets).toEqual(['Raid']);
		expect(loadStoredBulkBuildEnchantPresets(player)!.names).toEqual(['Raid']);
	});

	it('stores the names in list order and drops one no longer pickable', () => {
		const player = makeBulkPlayer();

		setBulkEnchantPickChecked(player, 'Undead', true, ['Raid', 'Undead', 'Old']);
		setBulkEnchantPickChecked(player, 'Old', true, ['Raid', 'Undead', 'Old']);
		setBulkEnchantPickChecked(player, 'Raid', true, ['Raid', 'Undead', 'Old']);

		expect(bulkState(player).buildEnchantPresets).toEqual(['Raid', 'Undead', 'Old']);

		setBulkEnchantPickChecked(player, 'Undead', false, ['Raid', 'Undead']);

		expect(loadStoredBulkBuildEnchantPresets(player)!.names).toEqual(['Raid']);
	});
});

describe('formatEnchantPickSummary', () => {
	it('counts the slots the preset enchants', () => {
		expect(
			formatEnchantPickSummary(
				preset('Raid', [
					[ItemSlot.ItemSlotHead, [2545]],
					[ItemSlot.ItemSlotChest, [1891]],
					[ItemSlot.ItemSlotBack, []],
				]),
			),
		).toBe('bulk_tab.settings.enchant_picks.slots:2');
	});
});

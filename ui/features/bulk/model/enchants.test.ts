import { BulkEnchantSlotOptions, BulkSettings } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkEnchantPreset, UIEnchant } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { bulkState, loadStoredEnchantPresets, seedBulkSettings } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore } from '@sim/state/sim_store';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { bulkEnchantPool, deleteBulkEnchantPreset, saveBulkEnchantPreset, setBulkEnchantPresetChecked, validateBulkEnchantPresetName } from './enchants';
import { refreshBulkCombinations } from './run';

const AGILITY = 10;
const STAMINA = 11;
const STRENGTH = 12;
const CRUSADER = 20;
const SHIELD_SPIKE = 21;

const USABLE = new Map<ItemSlot, number[]>([
	[ItemSlot.ItemSlotBack, [AGILITY, STAMINA, STRENGTH]],
	[ItemSlot.ItemSlotMainHand, [CRUSADER]],
	[ItemSlot.ItemSlotOffHand, [CRUSADER, SHIELD_SPIKE]],
]);

const preset = (name: string, slots: Array<[ItemSlot, number[]]>) => ({
	name,
	slots: slots.map(([slot, effectIds]) => BulkEnchantSlotOptions.create({ slot, effectIds })),
});

let player: Player<any>;
let stored: Map<string, string>;

beforeEach(() => {
	stored = new Map();
	player = {
		storeKey: 0,
		sim: {
			store: createSimStore(),
			env: {
				storage: {
					getItem: (key: string) => stored.get(key) ?? null,
					setItem: (key: string, value: string) => stored.set(key, value),
					removeItem: (key: string) => stored.delete(key),
				},
			},
		},
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		getGear: () => ({ getEquippedItem: () => null }),
		getEnchants: (slot: ItemSlot) => (USABLE.get(slot) ?? []).map(effectId => UIEnchant.create({ effectId })),
	} as unknown as Player<any>;
	seedBulkSettings(player);
});

const saveChecked = (...presets: Array<ReturnType<typeof preset>>) =>
	presets.forEach(saved => {
		saveBulkEnchantPreset(player, saved);
		setBulkEnchantPresetChecked(player, saved.name, true);
	});

const settingsVersion = () => bulkState(player).v.settings;

describe('bulkEnchantPool', () => {
	it('unions the checked presets per slot, each enchant once', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [STAMINA, AGILITY]]]), preset('Beasts', [[ItemSlot.ItemSlotBack, [AGILITY, STRENGTH]]]));

		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY, STAMINA, STRENGTH] })]);
	});

	it('ignores a preset that is not checked', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		saveBulkEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotMainHand, [CRUSADER]]]));

		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })]);
	});

	it('sends the weapon enchants for both hands and the off-hand enchants for the off hand only', () => {
		saveChecked(
			preset('Raid', [
				[ItemSlot.ItemSlotMainHand, [CRUSADER]],
				[ItemSlot.ItemSlotOffHand, [SHIELD_SPIKE]],
			]),
		);

		expect(bulkEnchantPool(player)).toEqual([
			BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotMainHand, effectIds: [CRUSADER] }),
			BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotOffHand, effectIds: [CRUSADER, SHIELD_SPIKE] }),
		]);
	});

	it('drops an enchant the player cannot use in that slot, and a slot left empty by it', () => {
		saveChecked(
			preset('Outdoor', [
				[ItemSlot.ItemSlotBack, [AGILITY, 99]],
				[ItemSlot.ItemSlotHead, [AGILITY]],
			]),
		);

		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })]);
	});
});

describe('enchant preset writers', () => {
	it('keeps a renamed preset checked, under its new name', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		saveBulkEnchantPreset(player, preset('Scourge', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead');

		expect(bulkState(player).enchantPresets.map(saved => [saved.name, saved.checked])).toEqual([['Scourge', true]]);
		expect(bulkEnchantPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [STAMINA] })]);
	});

	it('drops a deleted preset from the pool', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		deleteBulkEnchantPreset(player, 'Undead');

		expect(bulkState(player).enchantPresets).toEqual([]);
		expect(bulkEnchantPool(player)).toEqual([]);
	});

	it('trims the name and drops the slots with no enchant', () => {
		saveBulkEnchantPreset(player, {
			name: ' Undead ',
			slots: [
				BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotHead }),
				BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }),
			],
		});

		expect(bulkState(player).enchantPresets).toEqual([
			BulkEnchantPreset.create({ name: 'Undead', slots: [BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })] }),
		]);
	});

	it('stores an unchecked preset without recounting the combinations', () => {
		const before = settingsVersion();

		saveBulkEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		saveBulkEnchantPreset(player, preset('Scourge', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead');
		deleteBulkEnchantPreset(player, 'Scourge');
		saveBulkEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		expect(settingsVersion()).toBe(before);
		expect(loadStoredEnchantPresets(player)?.presets.map(saved => saved.name)).toEqual(['Beasts']);
	});

	it('recounts the combinations when the pool can change: a tick, or saving or deleting a checked preset', () => {
		saveBulkEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		const steps = [
			() => setBulkEnchantPresetChecked(player, 'Undead', true),
			() => saveBulkEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead'),
			() => deleteBulkEnchantPreset(player, 'Undead'),
		];
		for (const step of steps) {
			const before = settingsVersion();
			step();
			expect(settingsVersion()).toBe(before + 1);
		}
		expect(loadStoredEnchantPresets(player)?.presets).toEqual([]);
	});

	it('keeps the edit when the storage refuses to write', () => {
		player.sim.env.storage.setItem = () => {
			throw new Error('storage unavailable');
		};

		saveBulkEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		expect(bulkState(player).enchantPresets.map(saved => saved.name)).toEqual(['Undead']);
	});
});

describe('bulk requests', () => {
	it('carry the enchant pool', async () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		const getBulkCombinationCount = vi.fn(async (_settings: BulkSettings) => ({ combinations: 2, iterations: 1 }));

		await refreshBulkCombinations({
			player,
			sim: { getBulkCombinationCount },
			handleCrash: (e: unknown) => Promise.reject(e),
		} as unknown as IndividualSimHost<any>);

		expect(getBulkCombinationCount.mock.calls[0][0].enchants).toEqual([
			BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }),
		]);
	});
});

describe('validateBulkEnchantPresetName', () => {
	const presets = [BulkEnchantPreset.create({ name: 'Undead' }), BulkEnchantPreset.create({ name: 'Beasts' })];

	it.each([
		['an empty name', '', undefined, 'empty'],
		['a blank name', '   ', undefined, 'empty'],
		['a taken name', ' Beasts ', undefined, 'duplicate'],
		['another preset name while editing', 'Beasts', 'Undead', 'duplicate'],
		['the edited preset its own name', 'Undead', 'Undead', null],
		['a new name', 'Outdoor', undefined, null],
	])('answers %s', (_, name, previousName, expected) => {
		expect(validateBulkEnchantPresetName(presets, name, previousName)).toBe(expected);
	});
});

import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkEnchantPreset, UIEnchant } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { bulkState, loadStoredBulkBuildEnchantPresets, loadStoredEnchantPresets, patchBulkState } from '@sim/settings/bulk_settings';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { beforeEach, describe, expect, it } from 'vitest';

import { deleteEnchantPreset, enchantPresetPool, saveEnchantPreset, setEnchantPresetChecked, validateEnchantPresetName } from './enchant_presets';

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

beforeEach(() => {
	player = makeBulkPlayer({ getEnchants: (slot: ItemSlot) => (USABLE.get(slot) ?? []).map(effectId => UIEnchant.create({ effectId })) });
});

const saveChecked = (...presets: Array<ReturnType<typeof preset>>) =>
	presets.forEach(saved => {
		saveEnchantPreset(player, saved);
		setEnchantPresetChecked(player, saved.name, true);
	});

const settingsVersion = () => bulkState(player).v.settings;

describe('enchantPresetPool', () => {
	it('unions the checked presets per slot, each enchant once', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [STAMINA, AGILITY]]]), preset('Beasts', [[ItemSlot.ItemSlotBack, [AGILITY, STRENGTH]]]));

		expect(enchantPresetPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY, STAMINA, STRENGTH] })]);
	});

	it('ignores a preset that is not checked', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		saveEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotMainHand, [CRUSADER]]]));

		expect(enchantPresetPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })]);
	});

	it('sends the weapon enchants for both hands and the off-hand enchants for the off hand only', () => {
		saveChecked(
			preset('Raid', [
				[ItemSlot.ItemSlotMainHand, [CRUSADER]],
				[ItemSlot.ItemSlotOffHand, [SHIELD_SPIKE]],
			]),
		);

		expect(enchantPresetPool(player)).toEqual([
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

		expect(enchantPresetPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] })]);
	});
});

describe('enchant preset writers', () => {
	it('keeps a renamed preset checked, under its new name', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		saveEnchantPreset(player, preset('Scourge', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead');

		expect(bulkState(player).enchantPresets.map(saved => [saved.name, saved.checked])).toEqual([['Scourge', true]]);
		expect(enchantPresetPool(player)).toEqual([BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [STAMINA] })]);
	});

	it('keeps the Builds pick of a renamed preset under its new name, and drops a deleted one’s for good', () => {
		saveEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		saveEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotBack, [STAMINA]]]));
		patchBulkState(player, { buildEnchantPresets: ['Undead', 'Beasts'] });
		const buildPicks = () => [bulkState(player).buildEnchantPresets, loadStoredBulkBuildEnchantPresets(player)?.names];

		saveEnchantPreset(player, preset('Scourge', [[ItemSlot.ItemSlotBack, [AGILITY]]]), 'Undead');
		expect(buildPicks()).toEqual([
			['Scourge', 'Beasts'],
			['Scourge', 'Beasts'],
		]);

		deleteEnchantPreset(player, 'Beasts');
		saveEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotBack, [STAMINA]]]));
		expect(buildPicks()).toEqual([['Scourge'], ['Scourge']]);
	});

	it('drops a deleted preset from the pool', () => {
		saveChecked(preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		deleteEnchantPreset(player, 'Undead');

		expect(bulkState(player).enchantPresets).toEqual([]);
		expect(enchantPresetPool(player)).toEqual([]);
	});

	it('trims the name and drops the slots with no enchant', () => {
		saveEnchantPreset(player, {
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

		saveEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));
		saveEnchantPreset(player, preset('Scourge', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead');
		deleteEnchantPreset(player, 'Scourge');
		saveEnchantPreset(player, preset('Beasts', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		expect(settingsVersion()).toBe(before);
		expect(loadStoredEnchantPresets(player)?.presets.map(saved => saved.name)).toEqual(['Beasts']);
	});

	it('recounts the combinations when the pool can change: a tick, or saving or deleting a checked preset', () => {
		saveEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		const steps = [
			() => setEnchantPresetChecked(player, 'Undead', true),
			() => saveEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [STAMINA]]]), 'Undead'),
			() => deleteEnchantPreset(player, 'Undead'),
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

		saveEnchantPreset(player, preset('Undead', [[ItemSlot.ItemSlotBack, [AGILITY]]]));

		expect(bulkState(player).enchantPresets.map(saved => saved.name)).toEqual(['Undead']);
	});
});

describe('validateEnchantPresetName', () => {
	const presets = [BulkEnchantPreset.create({ name: 'Undead' }), BulkEnchantPreset.create({ name: 'Beasts' })];

	it.each([
		['an empty name', '', undefined, 'empty'],
		['a blank name', '   ', undefined, 'empty'],
		['a taken name', ' Beasts ', undefined, 'duplicate'],
		['another preset name while editing', 'Beasts', 'Undead', 'duplicate'],
		['the edited preset its own name', 'Undead', 'Undead', null],
		['a new name', 'Outdoor', undefined, null],
	])('answers %s', (_, name, previousName, expected) => {
		expect(validateEnchantPresetName(presets, name, previousName)).toBe(expected);
	});
});

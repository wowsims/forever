import { EquipmentSpec, ItemSlot, ItemSpec, Stat } from '@generated/proto/common';
import { BulkGearSetRef, SavedGearSet } from '@generated/proto/ui';
import type { PresetGear } from '@sim/presets/types';
import type { Gear } from '@sim/proto/gear';
import { Stats } from '@sim/proto/stats';
import type { IndividualSimHost } from '@sim/sim_host';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { compareGearSet } from '@testing/storybook/fixtures/bulk-compare';
import { describe, expect, it, vi } from 'vitest';

import { bulkGearSetEntries, bulkGearSetName, pickableBulkGearSets } from './gear_sets';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const SAVED_KEY = 'saved-gear';

// The helm id stands for the whole set: the stubbed database turns a spec into the set wearing its first item's id.
const spec = (id: number) => EquipmentSpec.create({ items: [ItemSpec.create({ id })] });

const helmId = (gear: Gear) => gear.getEquippedItem(ItemSlot.ItemSlotHead)?.item.id;

const strength = (value: number) => new Stats().withStat(Stat.StatStrength, value);

const entry = (name: string, preset: boolean, group = '') => ({
	ref: BulkGearSetRef.create({ name, preset, group }),
	name,
	gear: compareGearSet(2),
	bonusStats: new Stats(),
});

const gearPlayer = () => makeBulkPlayer({ sim: { db: { lookupEquipmentSpec: (equipment: EquipmentSpec) => compareGearSet(equipment.items[0]?.id ?? 0) } } });

describe('bulkGearSetEntries', () => {
	it('lists the presets whose condition passes for the player, then the saved sets, in panel order', () => {
		const player = gearPlayer();
		const presets: PresetGear[] = [
			{ name: 'Pre-raid', gear: spec(2) },
			{ name: 'Not for this player', gear: spec(3), enableWhen: () => false },
			{ name: 'Raid', gear: spec(4), enableWhen: candidate => candidate === player },
		];

		const entries = bulkGearSetEntries(presets, [{ name: 'Mine', data: SavedGearSet.create({ gear: spec(5) }) }], player);

		expect(entries.map(e => [e.name, e.ref.preset, helmId(e.gear)])).toEqual([
			['Pre-raid', true, 2],
			['Raid', true, 4],
			['Mine', false, 5],
		]);
	});

	it('gives a preset no bonus stats and a saved set the bonus stats it was saved with', () => {
		const entries = bulkGearSetEntries(
			[{ name: 'Raid', gear: spec(2) }],
			[
				{ name: 'Buffed', data: SavedGearSet.create({ gear: spec(3), bonusStatsStats: strength(10).toProto() }) },
				{ name: 'Plain', data: SavedGearSet.create({ gear: spec(4) }) },
			],
			gearPlayer(),
		);

		expect(entries.map(e => e.bonusStats.getStat(Stat.StatStrength))).toEqual([0, 10, 0]);
		expect(entries[0].bonusStats.equals(new Stats())).toBe(true);
	});
});

describe('pickableBulkGearSets', () => {
	it('reads the sets the Saved Gear panel stores, after the spec presets', () => {
		const player = gearPlayer();
		player.sim.env.storage.setItem(
			SAVED_KEY,
			JSON.stringify({ Mine: SavedGearSet.toJson(SavedGearSet.create({ gear: spec(5), bonusStatsStats: strength(10).toProto() })) }),
		);
		const host = {
			player,
			sim: player.sim,
			individualConfig: { presets: { gear: [{ name: 'Raid', gear: spec(2) }] } },
			getSavedGearStorageKey: () => SAVED_KEY,
		} as unknown as IndividualSimHost<any>;

		expect(pickableBulkGearSets(host).map(e => [e.name, helmId(e.gear), e.bonusStats.getStat(Stat.StatStrength)])).toEqual([
			['Raid', 2, 0],
			['Mine', 5, 10],
		]);
	});

	it('lists a stored set that still carries the removed bonus stats field, as the Saved Gear panel does', () => {
		const player = gearPlayer();
		const old = { ...(SavedGearSet.toJson(SavedGearSet.create({ gear: spec(5) })) as object), bonusStats: [10] };
		player.sim.env.storage.setItem(SAVED_KEY, JSON.stringify({ Old: old }));
		const host = {
			player,
			sim: player.sim,
			individualConfig: { presets: { gear: [] } },
			getSavedGearStorageKey: () => SAVED_KEY,
		} as unknown as IndividualSimHost<any>;

		expect(pickableBulkGearSets(host).map(e => [e.name, helmId(e.gear)])).toEqual([['Old', 5]]);
	});
});

describe('bulkGearSetName', () => {
	it('names a grouped preset by its group and name, and an ungrouped preset or a saved set by its name', () => {
		expect([entry('Pre-BiS', true, 'Fury'), entry('Raid', true), entry('Mine', false)].map(bulkGearSetName)).toEqual([
			'bulk_tab.results.grouped_gear_set',
			'Raid',
			'Mine',
		]);
	});
});

import { ArmorType } from '@generated/proto/common';
import { BulkEncounterRef, BulkTopResults, DatabaseFilters } from '@generated/proto/ui';
import { memoryEnv } from '@testing/sim/memory_env';
import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import { PlayerSpecs } from '../player/specs';
import { specStorageKey } from '../state/storage_keys';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '../wasm/bulk_sim/constants_auto_gen';
import {
	loadStoredBatchEncounter,
	loadStoredBulkDropsCombinedGain,
	loadStoredBulkDropsExcluded,
	loadStoredBulkDropsFilters,
	loadStoredBulkDropsOrder,
	loadStoredBulkDropsSources,
	loadStoredBulkDropsView,
	loadStoredBulkMode,
	loadStoredBulkTopResults,
	storeBatchEncounter,
	storeBulkDropsCombinedGain,
	storeBulkDropsExcluded,
	storeBulkDropsFilters,
	storeBulkDropsOrder,
	storeBulkDropsSources,
	storeBulkDropsView,
	storeBulkMode,
	storeBulkTopResults,
} from './bulk_settings';

const playerOn = () => ({ sim: { env: memoryEnv() }, getPlayerSpec: () => PlayerSpecs.DpsWarrior }) as unknown as Player<any>;

describe('loadStoredBulkTopResults', () => {
	it('loads a stored negative top results as its mode’s default', () => {
		const player = playerOn();

		storeBulkTopResults(player, BulkTopResults.create({ gear: -3, builds: -1 }));

		expect(loadStoredBulkTopResults(player)).toEqual(BulkTopResults.create({ gear: BULK_SIM_DEFAULT_TOP_RESULTS, builds: 0 }));
	});
});

describe('loadStoredBatchEncounter', () => {
	it('loads nothing stored as null and a blob that does not parse as off', () => {
		const player = playerOn();
		expect(loadStoredBatchEncounter(player)).toBeNull();

		player.sim.env.storage.setItem(specStorageKey(PlayerSpecs.DpsWarrior, 'bulk-batch-encounter.v1'), '{');

		expect(loadStoredBatchEncounter(player)).toEqual(BulkEncounterRef.create());
	});

	it('loads the stored pick back', () => {
		const player = playerOn();
		storeBatchEncounter(player, BulkEncounterRef.create({ name: 'Council' }));
		expect(loadStoredBatchEncounter(player)).toEqual(BulkEncounterRef.create({ name: 'Council' }));
	});
});

const setRaw = (player: Player<any>, key: string, value: string) => player.sim.env.storage.setItem(specStorageKey(PlayerSpecs.DpsWarrior, key), value);

describe('Drops storage', () => {
	it('loads a stored Drops mode and drops an unknown one', () => {
		const player = playerOn();
		storeBulkMode(player, 'drops');
		expect(loadStoredBulkMode(player)).toBe('drops');

		setRaw(player, 'bulk-mode.v1', 'raid');

		expect(loadStoredBulkMode(player)).toBeNull();
	});

	it('round-trips the Drops sources, excluded ids and view', () => {
		const player = playerOn();
		expect(loadStoredBulkDropsSources(player)).toBeNull();

		storeBulkDropsSources(player, ['drop:1584:9019', 'quest:other']);
		storeBulkDropsExcluded(player, [3, 17]);
		storeBulkDropsView(player, 'items');

		expect(loadStoredBulkDropsSources(player)).toEqual(['drop:1584:9019', 'quest:other']);
		expect(loadStoredBulkDropsExcluded(player)).toEqual([3, 17]);
		expect(loadStoredBulkDropsView(player)).toBe('items');
	});

	it('loads Drops sources that do not parse, or hold a non-string, as nothing stored', () => {
		const player = playerOn();
		setRaw(player, 'bulk-drops-sources.v1', '{');
		expect(loadStoredBulkDropsSources(player)).toBeNull();

		setRaw(player, 'bulk-drops-sources.v1', '[1]');

		expect(loadStoredBulkDropsSources(player)).toBeNull();
	});

	it('loads excluded ids that hold a non-integer as nothing stored', () => {
		const player = playerOn();
		setRaw(player, 'bulk-drops-excluded.v1', '["a"]');
		expect(loadStoredBulkDropsExcluded(player)).toBeNull();

		setRaw(player, 'bulk-drops-excluded.v1', '[1.5]');

		expect(loadStoredBulkDropsExcluded(player)).toBeNull();
	});

	it('round-trips the Drops filters per spec, and loads a blob that does not parse as nothing stored', () => {
		const player = playerOn();
		expect(loadStoredBulkDropsFilters(player)).toBeNull();
		const filters = DatabaseFilters.create({ armorTypes: [ArmorType.ArmorTypePlate], minIlvl: 58, oneHandedWeapons: true });

		storeBulkDropsFilters(player, filters);

		expect(loadStoredBulkDropsFilters(player)).toEqual(filters);
		expect(player.sim.env.storage.getItem(specStorageKey(PlayerSpecs.DpsWarrior, 'bulk-drops-filters.v1'))).not.toBeNull();

		setRaw(player, 'bulk-drops-filters.v1', '{');

		expect(loadStoredBulkDropsFilters(player)).toBeNull();
	});

	it('loads an unknown Drops view as nothing stored', () => {
		const player = playerOn();
		setRaw(player, 'bulk-drops-view.v1', 'cards');
		expect(loadStoredBulkDropsView(player)).toBeNull();
	});

	it('loads the Bosses order it stored, and an unknown one as nothing stored', () => {
		const player = playerOn();
		expect(loadStoredBulkDropsOrder(player)).toBeNull();
		storeBulkDropsOrder(player, 'zone');
		expect(loadStoredBulkDropsOrder(player)).toBe('zone');

		setRaw(player, 'bulk-drops-order.v1', 'boss');

		expect(loadStoredBulkDropsOrder(player)).toBeNull();
	});

	it('loads the Combined gain tick it stored, and anything else as nothing stored', () => {
		const player = playerOn();
		storeBulkDropsCombinedGain(player, false);
		expect(loadStoredBulkDropsCombinedGain(player)).toBe(false);

		setRaw(player, 'bulk-drops-combined-gain.v1', '1');

		expect(loadStoredBulkDropsCombinedGain(player)).toBeNull();
	});
});

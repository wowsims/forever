import { BulkEncounterRef, BulkTopResults } from '@generated/proto/ui';
import { memoryEnv } from '@testing/sim/memory_env';
import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import { PlayerSpecs } from '../player/specs';
import { specStorageKey } from '../state/storage_keys';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '../wasm/bulk_sim/constants_auto_gen';
import { loadStoredBatchEncounter, loadStoredBulkTopResults, storeBatchEncounter, storeBulkTopResults } from './bulk_settings';

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

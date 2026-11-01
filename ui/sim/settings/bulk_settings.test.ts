import { BulkTopResults } from '@generated/proto/ui';
import { memoryEnv } from '@testing/sim/memory_env';
import { describe, expect, it } from 'vitest';

import type { Player } from '../player/player';
import { PlayerSpecs } from '../player/specs';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '../wasm/bulk_sim/constants_auto_gen';
import { loadStoredBulkTopResults, storeBulkTopResults } from './bulk_settings';

const playerOn = () => ({ sim: { env: memoryEnv() }, getPlayerSpec: () => PlayerSpecs.DpsWarrior }) as unknown as Player<any>;

describe('loadStoredBulkTopResults', () => {
	it('loads a stored negative top results as its mode’s default', () => {
		const player = playerOn();

		storeBulkTopResults(player, BulkTopResults.create({ gear: -3, builds: -1 }));

		expect(loadStoredBulkTopResults(player)).toEqual(BulkTopResults.create({ gear: BULK_SIM_DEFAULT_TOP_RESULTS, builds: 0 }));
	});
});

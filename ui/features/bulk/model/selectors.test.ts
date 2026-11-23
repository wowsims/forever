import { bulkState } from '@sim/settings/bulk_settings';
import type { BulkSlice } from '@sim/state/sim_store';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';
import { estimateBulkSimIterations } from '@sim/wasm/bulk_sim/estimate';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it } from 'vitest';

import { bulkDropsIterations, canRunDropsBatch } from './selectors';

const slice = (fields: Partial<BulkSlice> = {}): BulkSlice => ({ ...bulkState(makeBulkPlayer()), combinations: 3, ...fields });

describe('Drops run gate', () => {
	it('estimates the iterations at the default top results, and none while the count reports an error', () => {
		expect(bulkDropsIterations(slice({ combinations: 40 }))).toBe(estimateBulkSimIterations(40, BULK_SIM_DEFAULT_TOP_RESULTS));
		expect(bulkDropsIterations(slice({ combinations: 40, combinationsError: 'no candidate' }))).toBe(0);
	});

	it('runs one candidate and more, with an item picked, inside both limits', () => {
		expect(canRunDropsBatch(slice({ combinations: 1 }), 1, 10, Infinity)).toBe(true);
		expect(canRunDropsBatch(slice({ combinations: 10 }), 3, 10, Infinity)).toBe(true);
	});

	it.each([
		['with no item picked', slice(), 0, Infinity],
		['when every pick is worn', slice({ combinations: 0 }), 2, Infinity],
		['while the count is pending', slice({ combinationsPending: true }), 2, Infinity],
		['while a batch runs', slice({ isRunning: true }), 2, Infinity],
		['over the combinations limit', slice({ combinations: 11 }), 2, Infinity],
		['over the iterations limit', slice(), 2, estimateBulkSimIterations(3, BULK_SIM_DEFAULT_TOP_RESULTS) - 1],
	])('stays off %s', (_case, state, picked, iterationsLimit) => {
		expect(canRunDropsBatch(state, picked, 10, iterationsLimit)).toBe(false);
	});
});

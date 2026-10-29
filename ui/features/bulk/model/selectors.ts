import type { BulkSlice } from '@sim/state/sim_store';
import { estimateBulkSimIterations } from '@sim/wasm/bulk_sim/estimate';

import type { BulkBuildsCount } from './builds';

/** The Gear run's minimum iterations at its top results (0 for All); 0 while the count reports an error. */
export const bulkGearIterations = (slice: BulkSlice, topResults: number): number =>
	slice.combinationsError ? 0 : estimateBulkSimIterations(slice.combinations, topResults || slice.combinations);

export const canRunBatch = (slice: BulkSlice, combinationsLimit: number, iterationsLimit: number, topResults: number): boolean =>
	!slice.combinationsPending &&
	!slice.isRunning &&
	slice.combinations > 1 &&
	slice.combinations <= combinationsLimit &&
	bulkGearIterations(slice, topResults) <= iterationsLimit;

// The requests run one at a time, so the iterations limit is per request.
export const canRunBuildsBatch = (isRunning: boolean, count: BulkBuildsCount, combinationsLimit: number, iterationsLimit: number): boolean =>
	!isRunning && count.builds > 1 && count.largestRequestCandidates <= combinationsLimit && count.largestRequestIterations <= iterationsLimit;

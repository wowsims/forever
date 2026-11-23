import type { BulkSlice } from '@sim/state/sim_store';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';
import { estimateBulkSimIterations } from '@sim/wasm/bulk_sim/estimate';

import type { BulkBuildsCount } from './builds/builds';

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

/** The Drops run's minimum iterations; 0 while the count reports an error. */
export const bulkDropsIterations = (slice: BulkSlice): number =>
	slice.combinationsError ? 0 : estimateBulkSimIterations(slice.combinations, BULK_SIM_DEFAULT_TOP_RESULTS);

// The Drops count leaves out the current gear, so one candidate is enough to run. Every pick worn counts 0.
export const canRunDropsBatch = (slice: BulkSlice, picked: number, combinationsLimit: number, iterationsLimit: number): boolean =>
	!slice.combinationsPending &&
	!slice.isRunning &&
	picked > 0 &&
	slice.combinations >= 1 &&
	slice.combinations <= combinationsLimit &&
	bulkDropsIterations(slice) <= iterationsLimit;

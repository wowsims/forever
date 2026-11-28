import type { BulkSlice } from '@sim/state/sim_store';
import { BULK_SIM_DEFAULT_TOP_RESULTS } from '@sim/wasm/bulk_sim/constants_auto_gen';
import { estimateBulkSimIterations } from '@sim/wasm/bulk_sim/estimate';

import type { BulkBuildsCount } from './builds/builds';
import { bulkCombinationsLimit, bulkIterationsLimit } from './run/limits';

/**
 * The run button's limits. The iterations gate needs power tools, which hold the Top results control that lowers it.
 * Without them Gear and Drops gate on the combinations limit alone, and past the iterations limit the run card only warns.
 */
export const bulkRunLimits = (isNative: boolean | undefined, powerTools: boolean): { combinations: number; iterations: number } => ({
	combinations: bulkCombinationsLimit(isNative),
	iterations: powerTools ? bulkIterationsLimit(isNative) : Number.POSITIVE_INFINITY,
});

/** The Gear run's minimum iterations at its top results (0 for All); 0 while the count reports an error. */
export const bulkGearIterations = (slice: BulkSlice, topResults: number): number =>
	slice.combinationsError ? 0 : estimateBulkSimIterations(slice.combinations, topResults || slice.combinations);

/** `minCombinations` is 2 for Gear, whose count includes the current gear, and 1 for Drops, whose count leaves it out. */
export const canRunBatch = (slice: BulkSlice, combinationsLimit: number, iterationsLimit: number, topResults: number, minCombinations = 2): boolean =>
	!slice.combinationsPending &&
	!slice.isRunning &&
	slice.combinations >= minCombinations &&
	slice.combinations <= combinationsLimit &&
	bulkGearIterations(slice, topResults) <= iterationsLimit;

// The requests run one at a time, so the iterations limit is per request.
export const canRunBuildsBatch = (isRunning: boolean, count: BulkBuildsCount, combinationsLimit: number, iterationsLimit: number): boolean =>
	!isRunning && count.builds > 1 && count.largestRequestCandidates <= combinationsLimit && count.largestRequestIterations <= iterationsLimit;

/** The Drops run's minimum iterations, at the default top results; 0 while the count reports an error. */
export const bulkDropsIterations = (slice: BulkSlice): number => bulkGearIterations(slice, BULK_SIM_DEFAULT_TOP_RESULTS);

// One candidate is enough to run. Every pick worn counts 0.
export const canRunDropsBatch = (slice: BulkSlice, picked: number, combinationsLimit: number, iterationsLimit: number): boolean =>
	picked > 0 && canRunBatch(slice, combinationsLimit, iterationsLimit, BULK_SIM_DEFAULT_TOP_RESULTS, 1);

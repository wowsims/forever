import { BulkSimStage } from '@generated/proto/api';
import type { TopGearResult } from '@sim/bulk/types';
import { Z_95, zTest } from '@sim/utils/math';

/** A paired error > 0 and a paired difference other than 0: a decoded response holds 0 for a difference the backend never set. */
export const hasPairedDiffToBaseline = (row: TopGearResult): row is TopGearResult & { pairedErrorToBaseline: number; pairedDpsDiffToBaseline: number } =>
	!!row.pairedErrorToBaseline && !!row.pairedDpsDiffToBaseline;

/** Culled before the finalist set: fewer iterations than the finalists, so no ± and no tight bracket. */
export const isCulled = (row: TopGearResult): boolean => row.culledStage !== undefined && row.culledStage !== BulkSimStage.BulkSimStageUnknown;

/** The row's own iteration count, `fallback` when its metrics carry none. */
export const rowIterations = (row: TopGearResult, fallback: number): number => row.dpsMetrics.aggregatorData?.n ?? fallback;

/** How far a row sits from the baseline in errors, and whether that tells them apart: its pairedDpsDiffToBaseline over its
 * pairedErrorToBaseline when it has both, else the unpaired zTest at each row's `rowIterations`, `iterations` for one with
 * none. The pairing can be against an earlier stage's baseline than `baseline`, so a paired row never reads the shown
 * baseline's mean. */
export const baselineSignificance = (row: TopGearResult, baseline: TopGearResult, iterations: number): { z: number; isDiff: boolean } => {
	if (hasPairedDiffToBaseline(row)) {
		const gap = Math.abs(row.pairedDpsDiffToBaseline);
		return { z: gap / row.pairedErrorToBaseline, isDiff: gap > Z_95 * row.pairedErrorToBaseline };
	}
	const n = (result: TopGearResult) => rowIterations(result, iterations);
	return zTest(n(row), row.dpsMetrics.avg, row.dpsMetrics.stdev, n(baseline), baseline.dpsMetrics.avg, baseline.dpsMetrics.stdev);
};

/** Whether `baselineSignificance` cannot tell the row apart from the baseline. */
export const isTiedWithBaseline = (row: TopGearResult, baseline: TopGearResult, iterations: number): boolean =>
	!baselineSignificance(row, baseline, iterations).isDiff;

/**
 * Whether two adjacent results are indistinguishable, which decides where the results table draws a
 * "tied" bracket.
 *
 * Two errors are available and they answer different questions. The backend returns a *paired*
 * error — the error of the difference between two runs that shared their random seeds, which is far
 * tighter than either run's own — but only for the pairs it actually compared: each result against
 * the baseline, and each result against the one ranked immediately below it. Anywhere else, or
 * where the backend returned nothing, this falls back to an unpaired two-sample z-test on the two
 * means, which is the honest answer for two runs that share no seed.
 *
 * `backendRank` is what makes "immediately below" mean the backend's order rather than the display
 * order: the two differ, because the baseline row is spliced in and everything is re-sorted by DPS.
 *
 * A culled row ties with the baseline only.
 */
const pairTied = (upper: TopGearResult, lower: TopGearResult, baseline: TopGearResult | null, iterations: number): boolean => {
	if (upper === baseline) return isTiedWithBaseline(lower, baseline, iterations);
	if (lower === baseline) return isTiedWithBaseline(upper, baseline, iterations);
	if (isCulled(upper) || isCulled(lower)) return false;
	if (upper.backendRank !== undefined && lower.backendRank === upper.backendRank + 1 && upper.pairedErrorToNextResult) {
		return Math.abs(upper.dpsMetrics.avg - lower.dpsMetrics.avg) <= Z_95 * upper.pairedErrorToNextResult;
	}
	// Under cull_to_baseline a non-finalist ran fewer iterations than the run's final count.
	const n = (result: TopGearResult) => rowIterations(result, iterations);
	return !zTest(n(upper), upper.dpsMetrics.avg, upper.dpsMetrics.stdev, n(lower), lower.dpsMetrics.avg, lower.dpsMetrics.stdev).isDiff;
};

/**
 * The results split into runs of adjacent rows that cannot be told apart. Every result appears in
 * exactly one chain and the order is preserved, so a chain of one is an ordinary row and a chain of
 * more than one is a bracket. Adjacency is deliberately transitive-by-neighbour rather than
 * all-pairs: A ties B and B ties C puts all three in one chain even where A and C would separate,
 * which is what "the table cannot order these" means for a ranked list.
 */
export const buildTieChains = <T extends TopGearResult>(results: readonly T[], baseline: T | null, iterations: number): T[][] => {
	const chains: T[][] = [];
	for (const result of results) {
		const currentChain = chains[chains.length - 1];
		const previousResult = currentChain?.[currentChain.length - 1];
		if (previousResult && pairTied(previousResult, result, baseline, iterations)) {
			currentChain.push(result);
		} else {
			chains.push([result]);
		}
	}
	return chains;
};

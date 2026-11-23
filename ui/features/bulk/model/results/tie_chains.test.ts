import { BulkSimStage, DistributionMetrics } from '@generated/proto/api';
import { TopGearResult } from '@sim/bulk/types';
import { describe, expect, it } from 'vitest';

import { buildTieChains, isTiedWithBaseline } from './tie_chains';

const result = (avg: number, stdev: number, overrides: Partial<TopGearResult> = {}): TopGearResult =>
	({ gear: null, dpsMetrics: { avg, stdev } as DistributionMetrics, ...overrides }) as TopGearResult;

const shape = (chains: TopGearResult[][]) => chains.map(chain => chain.map(entry => entry.dpsMetrics.avg));

describe('buildTieChains', () => {
	it('keeps every result, in order, exactly once', () => {
		const results = [result(300, 1), result(200, 1), result(100, 1)];
		expect(shape(buildTieChains(results, null, 100))).toEqual([[300], [200], [100]]);
	});

	it('groups neighbours the unpaired z-test cannot separate', () => {
		// stdev 500 over 100 iterations is an error of 50 apiece, so a 10 DPS gap is noise and a
		// 1000 DPS gap is not.
		const results = [result(1010, 500), result(1000, 500), result(0, 500)];
		expect(shape(buildTieChains(results, null, 100))).toEqual([[1010, 1000], [0]]);
	});

	it('takes the tighter paired error against the baseline when the backend supplied one', () => {
		// Unpaired, these two are indistinguishable; the paired error of 1 says the difference is
		// real, and the pair involving the baseline is exactly where it applies.
		const baseline = result(1000, 500);
		const candidate = result(1010, 500, { backendRank: 0, pairedErrorToBaseline: 1, pairedDpsDiffToBaseline: 10 });
		expect(shape(buildTieChains([candidate, baseline], baseline, 100))).toEqual([[1010], [1000]]);
	});

	it('uses pairedErrorToNextResult only for results the backend ranked adjacently', () => {
		const adjacent = [result(1010, 500, { backendRank: 0, pairedErrorToNextResult: 1 }), result(1000, 500, { backendRank: 1 })];
		expect(shape(buildTieChains(adjacent, null, 100))).toEqual([[1010], [1000]]);

		// Same numbers, but the backend ranked them two apart — the paired error does not describe
		// this pair, so the unpaired test answers and calls them tied.
		const separated = [result(1010, 500, { backendRank: 0, pairedErrorToNextResult: 1 }), result(1000, 500, { backendRank: 2 })];
		expect(shape(buildTieChains(separated, null, 100))).toEqual([[1010, 1000]]);
	});

	it('chains transitively through neighbours', () => {
		// A ties B and B ties C, while A and C are far enough apart to separate on their own.
		const results = [result(200, 500), result(150, 500), result(100, 500)];
		expect(shape(buildTieChains(results, null, 100))).toEqual([[200, 150, 100]]);
	});

	it('brackets a culled row with no neighbour but the baseline, and with the baseline only inside its paired error', () => {
		// Unpaired at 100 iterations every gap here is noise, so only the culled rule keeps the rows apart.
		const culled = (avg: number, pairedErrorToBaseline: number) =>
			result(avg, 500, { culledStage: BulkSimStage.BulkSimStageLow, pairedErrorToBaseline, pairedDpsDiffToBaseline: avg - 1000 });
		const baseline = result(1000, 500);
		const top = result(1010, 500, { backendRank: 0, pairedErrorToBaseline: 100, pairedDpsDiffToBaseline: 10 });
		const results = [top, culled(1005, 100), culled(1002, 100), baseline, culled(995, 1)];
		expect(shape(buildTieChains(results, baseline, 100))).toEqual([[1010], [1005], [1002, 1000], [995]]);
	});

	it('judges a paired row by its paired difference to the baseline it ran with, not by the two means', () => {
		// The rows' means sit against the refined baseline at 1000; their paired differences are against an earlier stage's.
		const baseline = result(1000, 500);
		const closeMeanClearlyBelow = result(1001, 500, { pairedErrorToBaseline: 1, pairedDpsDiffToBaseline: -10 });
		const farMeanLevel = result(980, 500, { pairedErrorToBaseline: 1, pairedDpsDiffToBaseline: -0.5 });
		expect(isTiedWithBaseline(closeMeanClearlyBelow, baseline)).toBe(false);
		expect(isTiedWithBaseline(farMeanLevel, baseline)).toBe(true);
	});

	it('judges a row with no paired error against the baseline at each row’s own iteration count', () => {
		// 30 DPS at stdev 500: z ≈ 9.5 at 50,000 iterations each, ≈ 1.9 at 1,000 against 50,000.
		const row = result(1030, 500, {
			pairedErrorToBaseline: 0,
			dpsMetrics: { avg: 1030, stdev: 500, aggregatorData: { n: 1000, sumSq: 0 } } as DistributionMetrics,
		});
		const baseline = result(1000, 500, { dpsMetrics: { avg: 1000, stdev: 500, aggregatorData: { n: 50000, sumSq: 0 } } as DistributionMetrics });
		const separated = result(1030, 500, { dpsMetrics: { avg: 1030, stdev: 500, aggregatorData: { n: 50000, sumSq: 0 } } as DistributionMetrics });
		expect(isTiedWithBaseline(row, baseline)).toBe(true);
		expect(isTiedWithBaseline(separated, baseline)).toBe(false);
	});

	it('is empty for no results', () => {
		expect(buildTieChains([], null, 100)).toEqual([]);
	});
});

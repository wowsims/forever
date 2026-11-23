import { BulkGearResult, BulkSimStage, DistributionMetrics } from '@generated/proto/api';
import { EquipmentSpec } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { appendBulkSimNonFinalists, bulkSimCandidateResultToProto } from './merge';
import type { ConcurrentBulkSimCandidateResult } from './types';

const HIGH_BASELINE_VALUES = [1000, 1010, 990, 1000];
const NOISE = [1, -1, 2, -2];

const candidateResult = (index: number, allValues: number[]): ConcurrentBulkSimCandidateResult => ({
	candidate: { index, gear: EquipmentSpec.create() },
	dpsMetrics: DistributionMetrics.create({ avg: allValues.reduce((sum, value) => sum + value, 0) / allValues.length, stdev: 10, allValues }),
});

// Candidate i sits 50 - 10i above the high-stage baseline, plus (i + 1) x 1, -1, 2, -2: its offset to the baseline, and to
// candidate i + 1, varies per iteration, so both pair with an error above zero.
const highValues = (index: number) => HIGH_BASELINE_VALUES.map((value, valueIdx) => value + 50 - 10 * index + (index + 1) * NOISE[valueIdx]);

describe('appendBulkSimNonFinalists', () => {
	it('appends the non-finalists by mean, each paired with the high-stage baseline', () => {
		const highBaseline = candidateResult(-1, HIGH_BASELINE_VALUES);
		const highResults = [4, 0, 6, 1, 3, 5, 2].map(index => candidateResult(index, highValues(index)));
		// The finalists and the baseline they were paired with carry eight values, so neither pairs with a high-stage row.
		const finalists = [0, 3].map(index => candidateResult(index, highValues(index).concat(highValues(index))));
		const refinedBaseline = candidateResult(-1, HIGH_BASELINE_VALUES.concat(HIGH_BASELINE_VALUES));
		const finalistRows = finalists.map(finalist =>
			BulkGearResult.create({ ...bulkSimCandidateResultToProto(finalist), pairedErrorToBaseline: refinedBaseline.dpsMetrics!.stdev }),
		);

		const rows = appendBulkSimNonFinalists(finalistRows.slice(), finalists, highResults, highBaseline);
		expect(rows.slice(0, 2)).toEqual(finalistRows);
		expect(rows.slice(2).map(row => row.candidateIndex)).toEqual([1, 2, 4, 5, 6]);
		expect(rows.slice(2).every(row => row.pairedErrorToBaseline > 0)).toBe(true);
		expect(rows.slice(2, -1).every(row => row.pairedErrorToNextResult > 0)).toBe(true);
		expect(rows[rows.length - 1].pairedErrorToNextResult).toBe(0);
		expect(rows.slice(2).every(row => row.culledStage === BulkSimStage.BulkSimStageUnknown && row.dpsMetrics!.allValues.length === 0)).toBe(true);
		expect(highResults.every(result => result.dpsMetrics!.allValues.length === 4)).toBe(true);
	});
});

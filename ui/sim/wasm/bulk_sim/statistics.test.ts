import { BulkSimRequest, BulkSimStage, DistributionMetrics } from '@generated/proto/api';
import { EquipmentSpec } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { BULK_SIM_DEFAULT_TOP_RESULTS, BULK_SIM_STAGE_CONFIGS } from './constants_auto_gen';
import { getBulkSimStageMaxSurvivors } from './stage';
import { cullBulkSimStage, selectBulkSimSurvivors } from './statistics';
import type { ConcurrentBulkSimCandidateResult, ConcurrentBulkSimStageConfig } from './types';

const MEDIUM: ConcurrentBulkSimStageConfig = {
	stage: BulkSimStage.BulkSimStageMedium,
	minIterations: 1000,
	targetErrorPct: 0.2,
	minSurvivors: 5,
	maxSurvivors: 25,
	cullingCoefficient: 1.35,
};

const ITERATIONS = 3;

// Candidate i at 1000 - 10i DPS on every iteration, so any two pair with no error and every gap is decisive.
const results = (count: number): ConcurrentBulkSimCandidateResult[] =>
	Array.from({ length: count }, (_, index) => {
		const avg = 1000 - 10 * index;
		return {
			candidate: { index, gear: EquipmentSpec.create() },
			dpsMetrics: DistributionMetrics.create({ avg, stdev: 0, allValues: Array(ITERATIONS).fill(avg) }),
		};
	});

const BASELINE: ConcurrentBulkSimCandidateResult = {
	candidate: { index: -1, gear: EquipmentSpec.create() },
	dpsMetrics: DistributionMetrics.create({ avg: 500, stdev: 0 }),
};

const indexesOf = (candidates: Array<{ index: number }>) => candidates.map(candidate => candidate.index).sort((a, b) => a - b);

const survivorIndexes = (count: number, topResults: number) =>
	indexesOf(selectBulkSimSurvivors(results(count), BASELINE, ITERATIONS, MEDIUM, count, topResults));

// Thirty unpaired candidates at stdev 10 over 100 iterations, so each interval's half-width is 10 / 10 x 1.35 x
// sqrt(log10 30) = 1.64. The leader sits at 1000, the next four at 990 down to 960, the sixth at 958 and the rest at 900
// and below.
const UNPAIRED_ITERATIONS = 100;
const unpaired = (): ConcurrentBulkSimCandidateResult[] =>
	[1000, 990, 980, 970, 960, 958, ...Array.from({ length: 24 }, (_, index) => 900 - index)].map((avg, index) => ({
		candidate: { index, gear: EquipmentSpec.create() },
		dpsMetrics: DistributionMetrics.create({ avg, stdev: 10 }),
	}));
const unpairedSurvivors = (topResults: number) => indexesOf(selectBulkSimSurvivors(unpaired(), BASELINE, UNPAIRED_ITERATIONS, MEDIUM, 30, topResults));

// 120 unpaired candidates at stdev 10 over 100 iterations, so each interval's half-width is 10 / 10 x 1.35 x
// sqrt(log10 120) = 1.95. The leader sits at 1000 and the rest at 900 - 0.03 x index, the last at 896.43.
const LOW = BULK_SIM_STAGE_CONFIGS.find(config => config.stage === BulkSimStage.BulkSimStageLow)!;
const lowStageSurvivorCount = (topResults: number) => {
	const candidates = Array.from({ length: 120 }, (_, index) => ({
		candidate: { index, gear: EquipmentSpec.create() },
		dpsMetrics: DistributionMetrics.create({ avg: index ? 900 - 0.03 * index : 1000, stdev: 10 }),
	}));
	return selectBulkSimSurvivors(candidates, BASELINE, UNPAIRED_ITERATIONS, LOW, candidates.length, topResults).length;
};

describe('selectBulkSimSurvivors', () => {
	it('keeps the stage minimum and culls everything behind the leader at the default top results', () => {
		expect(survivorIndexes(50, 5)).toEqual([0, 1, 2, 3, 4]);
	});

	it('culls against the leader at the default top results', () => {
		// 958 + 1.64 falls short of the leader's 1000 - 1.64, so the sixth candidate goes, though it would reach the fifth.
		expect(unpairedSurvivors(5)).toEqual([0, 1, 2, 3, 4]);
	});

	it('culls the low stage against the tenth best at top 10', () => {
		// The tenth sits at 899.73, so every candidate's 896.43 + 1.95 or more reaches its 899.73 - 1.95.
		expect(lowStageSurvivorCount(10)).toBe(120);
	});

	it('culls the low stage against the leader at the default top results', () => {
		// No candidate behind the leader reaches its 1000 - 1.95, so only the 20 the low stage keeps by mean survive.
		expect(lowStageSurvivorCount(5)).toBe(LOW.minSurvivors);
	});

	it('culls against the last shown candidate above the default top results', () => {
		// Six kept by mean; the rest, at 900 + 1.64 and below, fall short of the sixth's 958 - 1.64.
		expect(unpairedSurvivors(6)).toEqual([0, 1, 2, 3, 4, 5]);
	});

	it('keeps at least the top results when they pass the stage cap', () => {
		expect(survivorIndexes(50, 30)).toEqual(Array.from({ length: 30 }, (_, index) => index));
	});

	it('keeps every candidate while the raised cap covers them', () => {
		expect(survivorIndexes(28, 30)).toHaveLength(28);
	});
});

const range = (from: number, to: number) => Array.from({ length: to - from }, (_, index) => from + index);

// Thirty candidates paired with a baseline at 990, 1010, 990, 1010. 0 to 9 sit 10 down to 1 above it and 10 level with
// it, each on a constant offset, so they pair with no error. 11 trails by 1 with a paired error of 5, inside 1.96 x 5.
// 12 has no per-iteration values to pair. 13 trails by 20 with a paired error of 5, and 14 to 29 trail by 1 with none.
const CULL_BASELINE_VALUES = [990, 1010, 990, 1010];
const CULL_BASELINE: ConcurrentBulkSimCandidateResult = {
	candidate: { index: -1, gear: EquipmentSpec.create() },
	dpsMetrics: DistributionMetrics.create({ avg: 1000, stdev: 10, allValues: CULL_BASELINE_VALUES }),
};
const cullOffsets = (index: number): number[] => {
	if (index < 10) return Array(4).fill(10 - index);
	if (index === 10) return [0, 0, 0, 0];
	if (index === 11) return [-11, 9, -11, 9];
	if (index === 13) return [-30, -10, -30, -10];
	return [-1, -1, -1, -1];
};
const cullResults = (): ConcurrentBulkSimCandidateResult[] =>
	range(0, 30).map(index => {
		const candidate = { index, gear: EquipmentSpec.create() };
		if (index === 12) return { candidate, dpsMetrics: DistributionMetrics.create({ avg: 900, stdev: 10 }) };
		const offsets = cullOffsets(index);
		const avg = 1000 + offsets.reduce((sum, offset) => sum + offset, 0) / offsets.length;
		const allValues = CULL_BASELINE_VALUES.map((value, valueIdx) => value + offsets[valueIdx]);
		return { candidate, dpsMetrics: DistributionMetrics.create({ avg, stdev: 10, allValues }) };
	});

describe('cullBulkSimStage', () => {
	it('keeps every candidate not clearly below the baseline', () => {
		const results = cullResults();
		const { survivors, culled } = cullBulkSimStage(
			BulkSimRequest.create({ cullToBaseline: true }),
			results,
			CULL_BASELINE,
			4,
			MEDIUM,
			30,
			BULK_SIM_DEFAULT_TOP_RESULTS,
		);
		expect(indexesOf(survivors)).toEqual(range(0, 13));
		expect(culled.map(row => row.candidateIndex)).toEqual(range(13, 30));
		expect(culled.every(row => row.culledStage === BulkSimStage.BulkSimStageMedium && row.dpsMetrics!.allValues.length === 0)).toBe(true);
		expect(culled.every(row => row.dpsMetrics!.avg === results[row.candidateIndex].dpsMetrics!.avg)).toBe(true);
		expect(culled[0].pairedErrorToBaseline).toBeCloseTo(5, 9);
		expect(culled.slice(1).every(row => row.pairedErrorToBaseline === 0)).toBe(true);
		expect(results[13].dpsMetrics!.allValues).toHaveLength(4);
	});

	it('culls every candidate when all sit clearly below the baseline', () => {
		const below = range(0, 30).map(index => {
			const candidate = { index, gear: EquipmentSpec.create() };
			const allValues = CULL_BASELINE_VALUES.map(value => value - 20);
			return { candidate, dpsMetrics: DistributionMetrics.create({ avg: 980, stdev: 10, allValues }) };
		});
		const { survivors, culled } = cullBulkSimStage(
			BulkSimRequest.create({ cullToBaseline: true }),
			below,
			CULL_BASELINE,
			4,
			MEDIUM,
			30,
			BULK_SIM_DEFAULT_TOP_RESULTS,
		);
		expect(survivors).toEqual([]);
		expect(culled.map(row => row.candidateIndex)).toEqual(range(0, 30));
		expect(culled.every(row => row.culledStage === BulkSimStage.BulkSimStageMedium)).toBe(true);
	});

	it('culls against the leader without the flag', () => {
		const { survivors, culled } = cullBulkSimStage(BulkSimRequest.create(), cullResults(), CULL_BASELINE, 4, MEDIUM, 30, BULK_SIM_DEFAULT_TOP_RESULTS);
		expect(culled).toEqual([]);
		const indexes = indexesOf(survivors);
		expect(range(0, 5).every(index => indexes.includes(index))).toBe(true);
		expect(range(5, 11).some(index => indexes.includes(index))).toBe(false);
	});

	// 1,200 unpaired candidates at 1000 - 0.001 x index, stdev 10 over 4 iterations, so each interval overlaps the
	// leader's, and 300 paired ones 50 below a baseline at 890, 910, 890, 910. The low stage's cap for the 1,200 above
	// the cut is 110 (100 x sqrt(1.2)), so the soft cap keeps the 220 best.
	it('falls back to the top-results cull above the cut', () => {
		const baselineValues = [890, 910, 890, 910];
		const baseline: ConcurrentBulkSimCandidateResult = {
			candidate: { index: -1, gear: EquipmentSpec.create() },
			dpsMetrics: DistributionMetrics.create({ avg: 900, stdev: 10, allValues: baselineValues }),
		};
		const results: ConcurrentBulkSimCandidateResult[] = range(0, 1500).map(index => ({
			candidate: { index, gear: EquipmentSpec.create() },
			dpsMetrics:
				index < 1200
					? DistributionMetrics.create({ avg: 1000 - 0.001 * index, stdev: 10 })
					: DistributionMetrics.create({ avg: 850, stdev: 10, allValues: baselineValues.map(value => value - 50) }),
		}));

		const { survivors, culled } = cullBulkSimStage(
			BulkSimRequest.create({ cullToBaseline: true }),
			results,
			baseline,
			4,
			LOW,
			1500,
			BULK_SIM_DEFAULT_TOP_RESULTS,
		);
		// The stage's input count, 1,500, would give 2 x 123 = 246.
		expect(2 * getBulkSimStageMaxSurvivors(LOW, 1200, BULK_SIM_DEFAULT_TOP_RESULTS)!).toBe(220);
		expect(indexesOf(survivors)).toEqual(range(0, 220));
		expect(culled.map(row => row.candidateIndex)).toEqual(range(220, 1500));
		expect(culled[0].dpsMetrics!.avg).toBeGreaterThan(baseline.dpsMetrics!.avg);
	});
});

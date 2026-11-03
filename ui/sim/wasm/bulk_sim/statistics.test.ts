import { BulkSimStage, DistributionMetrics } from '@generated/proto/api';
import { EquipmentSpec } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { BULK_SIM_STAGE_CONFIGS } from './constants_auto_gen';
import { selectBulkSimSurvivors } from './statistics';
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

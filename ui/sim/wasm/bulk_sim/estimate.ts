import { BulkSimStage } from '@generated/proto/api';

import { BULK_SIM_MIN_COMBINATIONS } from './constants_auto_gen';
import { bulkSimStageConfigs, getBulkSimStageMaxSurvivors, shouldRunBulkSimStage } from './stage';

export const bulkSimHighStageFloor = (): number => bulkSimStageConfigs[bulkSimStageConfigs.length - 1]!.minIterations ?? 0;

const bulkSimSingleStageIterations = (candidateCount: number): number => bulkSimHighStageFloor() * (candidateCount + 1);

// A run that shows every candidate sims the high stage alone, as culling cannot change what it shows. So does one under
// BULK_SIM_MIN_COMBINATIONS candidates, and one the high stage alone sims in fewer iterations.
export const shouldUseSingleStageBulkSim = (candidateCount: number, topResults: number): boolean => {
	if (candidateCount < BULK_SIM_MIN_COMBINATIONS || topResults >= candidateCount) {
		return true;
	}

	return getBulkSimOptimisationIterationsUpperBound(candidateCount, topResults) >= bulkSimSingleStageIterations(candidateCount);
};

const getBulkSimOptimisationIterationsUpperBound = (candidateCount: number, topResults: number): number => {
	let remainingCandidates = candidateCount;
	let iterations = 0;

	for (const config of bulkSimStageConfigs) {
		if (config.stage === BulkSimStage.BulkSimStageHigh) {
			break;
		}
		if (!shouldRunBulkSimStage(config, remainingCandidates, topResults)) {
			continue;
		}

		iterations += (config.minIterations ?? 0) * (remainingCandidates + 1);
		remainingCandidates = Math.min(remainingCandidates, getBulkSimStageMaxSurvivors(config, remainingCandidates, topResults) ?? remainingCandidates);
	}

	return iterations + bulkSimHighStageFloor() * (remainingCandidates + 1);
};

// What a batch of candidateCount gear sets sims at each stage's minimum, the baseline included, with every culling stage
// keeping its most survivors. The error target can raise a stage above its minimum, and the adaptive passes and the finalist
// stage add iterations on top, so a run usually sims more.
export const estimateBulkSimIterations = (candidateCount: number, topResults: number): number =>
	shouldUseSingleStageBulkSim(candidateCount, topResults)
		? bulkSimSingleStageIterations(candidateCount)
		: getBulkSimOptimisationIterationsUpperBound(candidateCount, topResults);

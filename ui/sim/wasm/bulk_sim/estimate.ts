import { BulkSimStage } from '@generated/proto/api';

import { BULK_SIM_MIN_COMBINATIONS } from './constants_auto_gen';
import { bulkSimStageConfigs, getBulkSimStageMaxSurvivors, shouldRunBulkSimStage } from './stage';

export const bulkSimHighStageFloor = (): number => bulkSimStageConfigs[bulkSimStageConfigs.length - 1]!.minIterations ?? 0;

const bulkSimSingleStageIterations = (candidateCount: number): number => bulkSimHighStageFloor() * (candidateCount + 1);

export const shouldUseSingleStageBulkSim = (candidateCount: number): boolean => {
	if (candidateCount < BULK_SIM_MIN_COMBINATIONS) {
		return true;
	}

	return getBulkSimOptimisationIterationsUpperBound(candidateCount) >= bulkSimSingleStageIterations(candidateCount);
};

const getBulkSimOptimisationIterationsUpperBound = (candidateCount: number): number => {
	let remainingCandidates = candidateCount;
	let iterations = 0;

	for (const config of bulkSimStageConfigs) {
		if (config.stage === BulkSimStage.BulkSimStageHigh) {
			break;
		}
		if (!shouldRunBulkSimStage(config, remainingCandidates)) {
			continue;
		}

		iterations += (config.minIterations ?? 0) * (remainingCandidates + 1);
		remainingCandidates = Math.min(remainingCandidates, getBulkSimStageMaxSurvivors(config, remainingCandidates) ?? remainingCandidates);
	}

	return iterations + bulkSimHighStageFloor() * (remainingCandidates + 1);
};

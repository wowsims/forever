package bulk

import (
	"github.com/wowsims/forever/sim/core/proto"
)

func bulkSimHighStageFloor() int32 {
	return BulkSimStageConfigs[len(BulkSimStageConfigs)-1].MinIterations
}

func bulkSimSingleStageIterations(candidateCount int) int64 {
	return int64(bulkSimHighStageFloor()) * int64(candidateCount+1)
}

// A run that shows every candidate sims the high stage alone, as culling cannot change what it shows. So does one under
// BulkSimMinCombinations candidates, and one the high stage alone sims in fewer iterations.
func shouldUseSingleStageBulkSim(candidateCount int, topResults int) bool {
	if candidateCount < BulkSimMinCombinations || topResults >= candidateCount {
		return true
	}

	return getBulkSimOptimisationIterationsUpperBound(candidateCount, topResults) >= bulkSimSingleStageIterations(candidateCount)
}

func getBulkSimOptimisationIterationsUpperBound(candidateCount int, topResults int) int64 {
	remainingCandidates := candidateCount
	var iterations int64

	for _, stageConfig := range BulkSimStageConfigs {
		if stageConfig.Stage == proto.BulkSimStage_BulkSimStageHigh {
			break
		}
		if !shouldRunBulkSimStage(stageConfig, remainingCandidates, topResults) {
			continue
		}

		iterations += int64(stageConfig.MinIterations) * int64(remainingCandidates+1)
		remainingCandidates = min(remainingCandidates, getBulkSimStageMaxSurvivors(stageConfig, remainingCandidates, topResults))
	}

	return iterations + int64(bulkSimHighStageFloor())*int64(remainingCandidates+1)
}

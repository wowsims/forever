package bulk

import (
	"log"
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func mergeBulkSimCandidateResults(result *BulkSimCandidateResult, additionalResult *BulkSimCandidateResult) *BulkSimCandidateResult {
	// Nothing carried over: the additional run is the whole result.
	if result == nil {
		return additionalResult
	}
	if result.Error != nil {
		return result
	}
	if additionalResult == nil || additionalResult.Error != nil {
		return additionalResult
	}

	return &BulkSimCandidateResult{
		Candidate:  result.Candidate,
		DpsMetrics: mergeBulkSimDistributionMetrics(result.DpsMetrics, additionalResult.DpsMetrics),
	}
}

// Combines metrics from two independent sim runs for the same gear set.
// AggregatorData carries the sample count and sum of squares needed to
// recompute the weighted mean/stdev after adaptive extra iterations are
// appended.
func mergeBulkSimDistributionMetrics(metrics *proto.DistributionMetrics, additionalMetrics *proto.DistributionMetrics) *proto.DistributionMetrics {
	if metrics == nil {
		return additionalMetrics
	}
	if additionalMetrics == nil {
		return metrics
	}

	metricsAggregator := bulkSimDistributionMetricsAggregatorData(metrics)
	additionalAggregator := bulkSimDistributionMetricsAggregatorData(additionalMetrics)
	totalN := metricsAggregator.N + additionalAggregator.N
	if totalN <= 0 {
		return googleProto.Clone(metrics).(*proto.DistributionMetrics)
	}

	combinedMetrics := &proto.DistributionMetrics{
		Min:            math.MaxFloat64,
		MinSeed:        math.MaxInt64,
		Hist:           make(map[int32]int32),
		AllValues:      make([]float64, 0),
		AggregatorData: &proto.AggregatorData{},
	}
	core.CombineDistributionMetrics(combinedMetrics, metrics, false, float64(metricsAggregator.N)/float64(totalN))
	core.CombineDistributionMetrics(combinedMetrics, additionalMetrics, true, float64(additionalAggregator.N)/float64(totalN))
	return combinedMetrics
}

func bulkSimDistributionMetricsAggregatorData(metrics *proto.DistributionMetrics) *proto.AggregatorData {
	if metrics.AggregatorData != nil && metrics.AggregatorData.N > 0 {
		return metrics.AggregatorData
	}
	// Fabricating N=1 makes a merge silently mis-weight the two runs, so surface it
	// instead of letting a wrong average through unnoticed.
	log.Printf("[Bulk Sim] WARNING: distribution metrics are missing aggregator data; merged mean/stdev will be weighted as a single sample")
	n := int32(1)
	return &proto.AggregatorData{
		N:     n,
		SumSq: (metrics.Stdev*metrics.Stdev + metrics.Avg*metrics.Avg) * float64(n),
	}
}

func hasBulkSimStageError(baseline *BulkSimCandidateResult, results []*BulkSimCandidateResult) bool {
	if baseline != nil && baseline.Error != nil {
		return true
	}
	for _, result := range results {
		if result != nil && result.Error != nil {
			return true
		}
	}
	return false
}

func bulkSimResultsToCandidates(results []*BulkSimCandidateResult) []BulkSimCandidate {
	return core.MapSlice(results, func(result *BulkSimCandidateResult) BulkSimCandidate {
		return result.Candidate
	})
}

func bulkSimCandidateResultToProto(result *BulkSimCandidateResult) *proto.BulkGearResult {
	if result == nil {
		return nil
	}
	dpsMetrics := result.DpsMetrics
	if dpsMetrics != nil && len(dpsMetrics.AllValues) > 0 {
		allValues := dpsMetrics.AllValues
		dpsMetrics.AllValues = nil
		clone := googleProto.Clone(dpsMetrics).(*proto.DistributionMetrics)
		dpsMetrics.AllValues = allValues
		dpsMetrics = clone
	}
	return &proto.BulkGearResult{
		CandidateIndex: result.Candidate.Index,
		Gear:           result.Candidate.Gear,
		DpsMetrics:     dpsMetrics,
	}
}

// Appends every high-stage result outside the finalists at its high-stage numbers, by mean, high to low. Each row pairs
// with the high stage's baseline: the finalist stage's baseline carries more values and no longer pairs with it.
func appendBulkSimNonFinalists(rows []*proto.BulkGearResult, finalists, highResults []*BulkSimCandidateResult, highBaseline *BulkSimCandidateResult) []*proto.BulkGearResult {
	isFinalist := make(map[int32]bool, len(finalists))
	for _, finalist := range finalists {
		isFinalist[finalist.Candidate.Index] = true
	}
	nonFinalists := make([]*BulkSimCandidateResult, 0, len(highResults))
	for _, result := range highResults {
		if result != nil && result.DpsMetrics != nil && !isFinalist[result.Candidate.Index] {
			nonFinalists = append(nonFinalists, result)
		}
	}
	sortBulkSimResultsByDps(nonFinalists)

	for idx, result := range nonFinalists {
		row := bulkSimCandidateResultToProto(result)
		if idx+1 < len(nonFinalists) {
			if pairedError, ok := bulkSimPairedDpsError(nonFinalists[idx+1].DpsMetrics, result.DpsMetrics); ok {
				row.PairedErrorToNextResult = pairedError
			}
		}
		setBulkSimPairedToBaseline(row, result.DpsMetrics, highBaseline.DpsMetrics)
		rows = append(rows, row)
	}
	return rows
}

func cleanBulkSimDpsMetrics(metrics *proto.DistributionMetrics) *proto.DistributionMetrics {
	if metrics == nil {
		return nil
	}
	// AllValues is kept: selectBulkSimSurvivors needs the per-iteration values. It is
	// stripped again in bulkSimCandidateResultToProto, so it never reaches the frontend.
	clone := googleProto.Clone(metrics).(*proto.DistributionMetrics)
	clone.Hist = nil
	return clone
}

func bulkSimAbortedError() *proto.ErrorOutcome {
	return &proto.ErrorOutcome{Type: proto.ErrorOutcomeType_ErrorOutcomeAborted, Message: "Bulk Sim Aborted"}
}

package bulk

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/testagent"
)

func init() {
	testagent.RegisterAsDpsWarrior()
}

func bulkSimTestRequest(simIterations int32, candidateCount int) *proto.BulkSimRequest {
	request := &proto.BulkSimRequest{
		BaseRequest: &proto.RaidSimRequest{
			SimOptions: &proto.SimOptions{Iterations: simIterations, RandomSeed: 101, IsTest: true},
			Raid: core.SinglePlayerRaidProto(&proto.Player{
				Name:      "Warrior",
				Class:     proto.Class_ClassWarrior,
				Race:      proto.Race_RaceOrc,
				Equipment: &proto.EquipmentSpec{},
				Spec:      &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}},
				Buffs:     &proto.IndividualBuffs{},
				Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
				// Out of melee range, so every iteration deals no damage and no stage raises its iterations to meet the error target.
				DistanceFromTarget: 100,
			}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter: &proto.Encounter{Duration: 10, Targets: []*proto.Target{core.NewDefaultTarget()}},
		},
	}
	for idx := range candidateCount {
		request.Candidates = append(request.Candidates, &proto.BulkGearCandidate{Index: int32(idx), Gear: &proto.EquipmentSpec{}})
	}
	return request
}

func simmedIterations(result *proto.BulkGearResult) int32 {
	return result.DpsMetrics.AggregatorData.N
}

func TestEstimateBulkSimIterationsCountsTheBaselineAtTheStageMinimums(t *testing.T) {
	cases := []struct {
		name       string
		candidates int
		want       int64
	}{
		{"no candidates sims the baseline alone", 0, 10000},
		{"16 candidates and the baseline in the high stage", 16, 17 * 10000},
		{"27 candidates sim fewer iterations in the high stage alone than culling through the medium stage", 27, 28 * 10000},
		{"28 candidates sim fewer iterations culling through the medium stage than in the high stage alone", 28, 29*1000 + 26*10000},
		{"100 candidates cull through the medium stage to 25", 100, 101*1000 + 26*10000},
		{"1000 candidates cull through the low stage to 100 and the medium stage to 25", 1000, 1001*100 + 101*1000 + 26*10000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := estimateBulkSimIterations(tc.candidates); got != tc.want {
				t.Fatalf("estimate for %d candidates = %d, want %d", tc.candidates, got, tc.want)
			}
		})
	}
}

func TestBulkSimRunsAtTheStageMinimumsNotTheSimIterations(t *testing.T) {
	const simIterations = 12500
	cases := []struct {
		name       string
		candidates int
	}{
		{"a run with no candidates sims the baseline at the high stage minimum", 0},
		{"a run with candidates sims them and the baseline at the high stage minimum", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := runBulkSim(bulkSimTestRequest(simIterations, tc.candidates), nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatalf("bulk sim failed: %s", result.Error.Message)
			}
			if got := simmedIterations(result.Baseline); got != 10000 {
				t.Fatalf("baseline simmed %d iterations, want 10000", got)
			}
			if len(result.TopResults) != tc.candidates {
				t.Fatalf("got %d top results, want %d", len(result.TopResults), tc.candidates)
			}
			for _, topResult := range result.TopResults {
				if got := simmedIterations(topResult); got != 10000 {
					t.Fatalf("candidate simmed %d iterations, want 10000", got)
				}
			}
		})
	}
}

func TestGetBulkSimStageMaxSurvivorsScalesLowStage(t *testing.T) {
	lowStageConfig := BulkSimStageConfig{
		Stage:        proto.BulkSimStage_BulkSimStageLow,
		MaxSurvivors: 100,
	}

	testCases := []struct {
		name           string
		candidateCount int
		want           int
	}{
		{name: "below reference", candidateCount: 863, want: 100},
		{name: "at reference", candidateCount: 1000, want: 100},
		{name: "large candidate set", candidateCount: 13000, want: 361},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := getBulkSimStageMaxSurvivors(lowStageConfig, testCase.candidateCount); got != testCase.want {
				t.Fatalf("max survivors for %d candidates = %d, want %d", testCase.candidateCount, got, testCase.want)
			}
		})
	}
}

func TestGetBulkSimStageMaxSurvivorsScalesMediumStage(t *testing.T) {
	mediumStageConfig := BulkSimStageConfig{
		Stage:        proto.BulkSimStage_BulkSimStageMedium,
		MaxSurvivors: 25,
	}

	testCases := []struct {
		name           string
		candidateCount int
		want           int
	}{
		{name: "below reference", candidateCount: 50, want: 25},
		{name: "at reference", candidateCount: 100, want: 25},
		{name: "large low-stage output", candidateCount: 722, want: 68},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := getBulkSimStageMaxSurvivors(mediumStageConfig, testCase.candidateCount); got != testCase.want {
				t.Fatalf("max survivors for %d candidates = %d, want %d", testCase.candidateCount, got, testCase.want)
			}
		})
	}
}

func TestGetBulkSimStageMaxSurvivorsKeepsHighStageUncapped(t *testing.T) {
	highStageConfig := BulkSimStageConfig{
		Stage:        proto.BulkSimStage_BulkSimStageHigh,
		MaxSurvivors: 0,
	}

	if got := getBulkSimStageMaxSurvivors(highStageConfig, 13000); got != 0 {
		t.Fatalf("high max survivors = %d, want uncapped", got)
	}
}

func TestMergeBulkSimDistributionMetrics(t *testing.T) {
	metrics := newBulkSimTestDistributionMetrics([]float64{8, 12})
	metrics.MaxSeed = 12
	metrics.MinSeed = 8

	additionalMetrics := newBulkSimTestDistributionMetrics([]float64{16, 20, 24})
	additionalMetrics.MaxSeed = 24
	additionalMetrics.MinSeed = 16

	merged := mergeBulkSimDistributionMetrics(metrics, additionalMetrics)

	assertFloatEqual(t, "avg", merged.Avg, 16)
	assertFloatEqual(t, "stdev", merged.Stdev, math.Sqrt(32))
	if merged.AggregatorData.N != 5 {
		t.Fatalf("expected 5 merged samples, got %d", merged.AggregatorData.N)
	}
	assertFloatEqual(t, "sumSq", merged.AggregatorData.SumSq, 1440)
	assertFloatEqual(t, "max", merged.Max, 24)
	assertFloatEqual(t, "min", merged.Min, 8)
	if merged.MaxSeed != 24 {
		t.Fatalf("expected max seed 24, got %d", merged.MaxSeed)
	}
	if merged.MinSeed != 8 {
		t.Fatalf("expected min seed 8, got %d", merged.MinSeed)
	}
}

func newBulkSimTestDistributionMetrics(values []float64) *proto.DistributionMetrics {
	metrics := &proto.DistributionMetrics{
		Min:            math.MaxFloat64,
		AggregatorData: &proto.AggregatorData{N: int32(len(values))},
	}
	for idx, value := range values {
		metrics.Avg += value
		metrics.AggregatorData.SumSq += value * value
		if value > metrics.Max {
			metrics.Max = value
			metrics.MaxSeed = int64(value)
		}
		if value < metrics.Min {
			metrics.Min = value
			metrics.MinSeed = int64(value)
		}
		if idx == len(values)-1 {
			metrics.Avg /= float64(len(values))
		}
	}
	metrics.Stdev = math.Sqrt(metrics.AggregatorData.SumSq/float64(len(values)) - metrics.Avg*metrics.Avg)
	return metrics
}

func assertFloatEqual(t *testing.T, name string, actual float64, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("expected %s %.12f, got %.12f", name, expected, actual)
	}
}

package bulk

import (
	"math"
	"slices"
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

func TestShouldUseSingleStageBulkSimTakesTheCheaperPlanAtTheStageMinimums(t *testing.T) {
	cases := []struct {
		name       string
		candidates int
		topResults int
		single     bool
		staged     int64 // What the culling plan sims at the stage minimums, the baseline included; 0 leaves it unchecked.
	}{
		{"16 candidates go straight to the high stage", 16, 5, true, 0},
		{"27 candidates sim fewer iterations in the high stage alone than culling through the medium stage", 27, 5, true, 28*1000 + 26*10000},
		{"28 candidates sim fewer iterations culling through the medium stage than in the high stage alone", 28, 5, false, 29*1000 + 26*10000},
		{"100 candidates cull through the medium stage to 25", 100, 5, false, 101*1000 + 26*10000},
		{"1000 candidates cull through the low stage to 100 and the medium stage to 25", 1000, 5, false, 1001*100 + 101*1000 + 26*10000},
		{"a top 30 of 100 candidates keeps 30 through the medium stage", 100, 30, false, 101*1000 + 31*10000},
		{"28 candidates all shown sim the high stage alone", 28, 28, true, 0},
		{"1000 candidates all shown sim the high stage alone", 1000, 1000, true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldUseSingleStageBulkSim(tc.candidates, tc.topResults); got != tc.single {
				t.Fatalf("single stage for %d candidates, top %d = %v, want %v", tc.candidates, tc.topResults, got, tc.single)
			}
			if tc.staged == 0 {
				return
			}
			if got := getBulkSimOptimisationIterationsUpperBound(tc.candidates, tc.topResults); got != tc.staged {
				t.Fatalf("culling plan for %d candidates, top %d = %d, want %d", tc.candidates, tc.topResults, got, tc.staged)
			}
		})
	}
}

func bulkSimStagesRun(result *proto.BulkSimResult) []proto.BulkSimStage {
	stages := make([]proto.BulkSimStage, 0, len(result.StageMetrics))
	for _, metrics := range result.StageMetrics {
		stages = append(stages, metrics.Stage)
	}
	return stages
}

// Swaps in a ladder of the default one's shape at a hundredth of its iterations: a medium stage culling to 25 at 10
// iterations and a high stage at 100, neither raised by its error target.
func useCheapBulkSimStages(t *testing.T) {
	saved := BulkSimStageConfigs
	BulkSimStageConfigs = []BulkSimStageConfig{
		{Stage: proto.BulkSimStage_BulkSimStageMedium, MinIterations: 10, TargetErrorPct: 100, MinSurvivors: 5, MaxSurvivors: 25, CullingCoefficient: BulkSimCullingCoefficient},
		{Stage: proto.BulkSimStage_BulkSimStageHigh, MinIterations: 100, TargetErrorPct: 100, UseConcurrentSim: true},
	}
	t.Cleanup(func() { BulkSimStageConfigs = saved })
}

// The top-results rule is only visible on a count the culling plan sims more cheaply than the high stage alone.
func requireCullingPlanAtTheDefaultTopResults(t *testing.T, candidates int) {
	t.Helper()
	if shouldUseSingleStageBulkSim(candidates, BulkSimDefaultTopResults) {
		t.Fatalf("%d candidates take the high stage alone at the default top results, so the test cannot tell the top-results rule apart", candidates)
	}
}

func TestBulkSimShowingEveryCandidateRunsTheHighStageAlone(t *testing.T) {
	useCheapBulkSimStages(t)
	const candidates = 28
	requireCullingPlanAtTheDefaultTopResults(t, candidates)
	request := bulkSimTestRequest(12500, candidates)
	request.TopResults = candidates
	result := runBulkSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("bulk sim failed: %s", result.Error.Message)
	}
	if stages := bulkSimStagesRun(result); len(stages) != 1 || stages[0] != proto.BulkSimStage_BulkSimStageHigh {
		t.Fatalf("ran stages %v, want the high stage alone", stages)
	}
	if len(result.TopResults) != candidates {
		t.Fatalf("got %d top results, want every one of the %d candidates", len(result.TopResults), candidates)
	}
	for _, topResult := range result.TopResults {
		if got, want := simmedIterations(topResult), bulkSimHighStageFloor(); got != want {
			t.Fatalf("candidate simmed %d iterations, want %d", got, want)
		}
	}
}

func TestBulkSimAskedForAllResultsRanksEveryCandidateInTheHighStageAlone(t *testing.T) {
	useCheapBulkSimStages(t)
	const candidates = 28
	requireCullingPlanAtTheDefaultTopResults(t, candidates)
	request := bulkSimTestRequest(12500, candidates)
	request.AllResults = true
	request.TopResults = 3
	result := runBulkSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("bulk sim failed: %s", result.Error.Message)
	}
	if stages := bulkSimStagesRun(result); len(stages) != 1 || stages[0] != proto.BulkSimStage_BulkSimStageHigh {
		t.Fatalf("ran stages %v, want the high stage alone", stages)
	}
	if len(result.TopResults) != candidates {
		t.Fatalf("got %d top results, want every one of the %d candidates", len(result.TopResults), candidates)
	}
}

func TestBulkSimShowingTheDefaultTopResultsRunsTheCullingStageFirst(t *testing.T) {
	useCheapBulkSimStages(t)
	const candidates = 28
	requireCullingPlanAtTheDefaultTopResults(t, candidates)
	result := runBulkSim(bulkSimTestRequest(12500, candidates), nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("bulk sim failed: %s", result.Error.Message)
	}
	stages := bulkSimStagesRun(result)
	if len(stages) < 2 || stages[0] != proto.BulkSimStage_BulkSimStageMedium || stages[1] != proto.BulkSimStage_BulkSimStageHigh {
		t.Fatalf("ran stages %v, want the medium stage and then the high stage", stages)
	}
	if len(result.TopResults) != BulkSimDefaultTopResults {
		t.Fatalf("got %d top results, want %d", len(result.TopResults), BulkSimDefaultTopResults)
	}
}

func TestBulkSimKeepsTheTopResultsThroughTheCullingStage(t *testing.T) {
	useCheapBulkSimStages(t)
	const candidates, topResults = 100, 30
	request := bulkSimTestRequest(12500, 0)
	// In melee range and unarmed, so each candidate's chest, worth 20 attack power more than the last, spreads the DPS.
	request.BaseRequest.Raid.Parties[0].Players[0].DistanceFromTarget = 0
	for idx := range candidates {
		id := int32(930000 + idx)
		core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{
			Id:   id,
			Type: proto.ItemType_ItemTypeChest,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{
				0: {Stats: map[int32]float64{int32(proto.Stat_StatAttackPower): float64(20 * (idx + 1))}},
			},
		}}})
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotChest+1)
		for slot := range items {
			items[slot] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: id}
		request.Candidates = append(request.Candidates, &proto.BulkGearCandidate{Index: int32(idx), Gear: &proto.EquipmentSpec{Items: items}})
	}
	request.TopResults = topResults
	result := runBulkSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("bulk sim failed: %s", result.Error.Message)
	}
	if stages := bulkSimStagesRun(result); len(stages) < 2 || stages[0] != proto.BulkSimStage_BulkSimStageMedium {
		t.Fatalf("ran stages %v, want the medium stage first", stages)
	}
	if survivors := result.StageMetrics[0].Survivors; survivors < topResults {
		t.Fatalf("the medium stage kept %d candidates, want at least %d", survivors, topResults)
	}
	if len(result.TopResults) != topResults {
		t.Fatalf("got %d top results, want %d", len(result.TopResults), topResults)
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
			if got := getBulkSimStageMaxSurvivors(lowStageConfig, testCase.candidateCount, BulkSimDefaultTopResults); got != testCase.want {
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
			if got := getBulkSimStageMaxSurvivors(mediumStageConfig, testCase.candidateCount, BulkSimDefaultTopResults); got != testCase.want {
				t.Fatalf("max survivors for %d candidates = %d, want %d", testCase.candidateCount, got, testCase.want)
			}
		})
	}
}

func TestGetBulkSimStageMaxSurvivorsKeepsAtLeastTheTopResults(t *testing.T) {
	mediumStageConfig := BulkSimStageConfig{Stage: proto.BulkSimStage_BulkSimStageMedium, MaxSurvivors: 25}

	if got := getBulkSimStageMaxSurvivors(mediumStageConfig, 50, 30); got != 30 {
		t.Fatalf("max survivors for a top 30 of 50 candidates = %d, want 30", got)
	}
	if got := getBulkSimStageMaxSurvivors(mediumStageConfig, 722, 30); got != 68 {
		t.Fatalf("max survivors for a top 30 of 722 candidates = %d, want the scaled 68", got)
	}
}

// Thirty unpaired candidates at stdev 10 over 100 iterations, so each interval's half-width is 10 / 10 x 1.35 x
// sqrt(log10 30) = 1.64. The leader sits at 1000, the next four at 990 down to 960, the sixth at 958 and the rest at 900
// and below; the baseline trails them all.
func unpairedBulkSimSurvivors(topResults int) []int32 {
	averages := []float64{1000, 990, 980, 970, 960, 958}
	for idx := range 24 {
		averages = append(averages, float64(900-idx))
	}
	results := make([]*BulkSimCandidateResult, 0, len(averages))
	for idx, avg := range averages {
		results = append(results, &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: int32(idx)}, DpsMetrics: &proto.DistributionMetrics{Avg: avg, Stdev: 10}})
	}
	baseline := &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: -1}, DpsMetrics: &proto.DistributionMetrics{Avg: 500}}
	medium := BulkSimStageConfig{Stage: proto.BulkSimStage_BulkSimStageMedium, MinSurvivors: 5, MaxSurvivors: 25, CullingCoefficient: BulkSimCullingCoefficient}
	indexes := []int32{}
	for _, candidate := range selectBulkSimSurvivors(results, baseline, 100, medium, len(averages), topResults) {
		indexes = append(indexes, candidate.Index)
	}
	slices.Sort(indexes)
	return indexes
}

func TestSelectBulkSimSurvivorsCullsAgainstTheLeaderAtTheDefaultTopResults(t *testing.T) {
	// 958 + 1.64 falls short of the leader's 1000 - 1.64, so the sixth candidate goes, though it would reach the fifth.
	if got, want := unpairedBulkSimSurvivors(BulkSimDefaultTopResults), []int32{0, 1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("survivors at the default top results = %v, want %v", got, want)
	}
}

func TestSelectBulkSimSurvivorsCullsAgainstTheLastShownCandidateAtALargerCap(t *testing.T) {
	// Six kept by mean; the rest, at 900 + 1.64 and below, fall short of the sixth's 958 - 1.64.
	if got, want := unpairedBulkSimSurvivors(6), []int32{0, 1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Fatalf("survivors at top 6 = %v, want %v", got, want)
	}
}

// 120 unpaired candidates at stdev 10 over 100 iterations, so each interval's half-width is 10 / 10 x 1.35 x
// sqrt(log10 120) = 1.95. The leader sits at 1000 and the rest at 900 - 0.03 x index, the last at 896.43.
func lowStageBulkSimSurvivorCount(t *testing.T, topResults int) int {
	t.Helper()
	low := BulkSimStageConfigs[0]
	if low.Stage != proto.BulkSimStage_BulkSimStageLow || low.MinSurvivors <= topResults {
		t.Fatalf("the first stage config is %v keeping %d by mean, want the low stage keeping more than %d", low.Stage, low.MinSurvivors, topResults)
	}
	results := make([]*BulkSimCandidateResult, 0, 120)
	for idx := range 120 {
		avg := 900 - 0.03*float64(idx)
		if idx == 0 {
			avg = 1000
		}
		results = append(results, &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: int32(idx)}, DpsMetrics: &proto.DistributionMetrics{Avg: avg, Stdev: 10}})
	}
	baseline := &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: -1}, DpsMetrics: &proto.DistributionMetrics{Avg: 500}}
	return len(selectBulkSimSurvivors(results, baseline, 100, low, len(results), topResults))
}

func TestSelectBulkSimSurvivorsCullsTheLowStageAgainstTheTenthBestAtTopTen(t *testing.T) {
	// The tenth sits at 899.73, so every candidate's 896.43 + 1.95 or more reaches its 899.73 - 1.95.
	if got := lowStageBulkSimSurvivorCount(t, 10); got != 120 {
		t.Fatalf("the low stage kept %d of 120 candidates at top 10, want every one", got)
	}
}

func TestSelectBulkSimSurvivorsCullsTheLowStageAgainstTheLeaderAtTheDefaultTopResults(t *testing.T) {
	// No candidate behind the leader reaches its 1000 - 1.95, so only the 20 the low stage keeps by mean survive.
	if got := lowStageBulkSimSurvivorCount(t, BulkSimDefaultTopResults); got != BulkSimStageConfigs[0].MinSurvivors {
		t.Fatalf("the low stage kept %d of 120 candidates at the default top results, want its %d by mean", got, BulkSimStageConfigs[0].MinSurvivors)
	}
}

func TestEnsureBulkSimCandidatesGeneratedLeavesARequestWithoutBulkSettingsAlone(t *testing.T) {
	request := bulkSimTestRequest(100, 0)
	if err := EnsureBulkSimCandidatesGenerated(request); err != nil {
		t.Fatalf("candidate generation failed: %s", err)
	}
	if len(request.Candidates) != 0 {
		t.Fatalf("generated %d candidates for a request without bulk settings, want none", len(request.Candidates))
	}
}

func TestGetBulkSimStageMaxSurvivorsKeepsHighStageUncapped(t *testing.T) {
	highStageConfig := BulkSimStageConfig{
		Stage:        proto.BulkSimStage_BulkSimStageHigh,
		MaxSurvivors: 0,
	}

	if got := getBulkSimStageMaxSurvivors(highStageConfig, 13000, BulkSimDefaultTopResults); got != 0 {
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

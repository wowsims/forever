package bulk

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/testagent"
	googleProto "google.golang.org/protobuf/proto"
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

// Thirty candidates paired with a baseline at 990, 1010, 990, 1010. 0 to 9 sit 10 down to 1 above it and 10 level with
// it, each on a constant offset, so they pair with no error. 11 trails by 1 with a paired error of 5, inside 1.96 x 5.
// 12 has no per-iteration values to pair. 13 trails by 20 with a paired error of 5, and 14 to 29 trail by 1 with none.
func baselineCullFixture() ([]*BulkSimCandidateResult, *BulkSimCandidateResult) {
	baselineValues := []float64{990, 1010, 990, 1010}
	baseline := &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: -1}, DpsMetrics: &proto.DistributionMetrics{Avg: 1000, Stdev: 10, AllValues: baselineValues}}
	results := make([]*BulkSimCandidateResult, 0, 30)
	for idx := range 30 {
		var offsets []float64
		switch {
		case idx < 10:
			offset := float64(10 - idx)
			offsets = []float64{offset, offset, offset, offset}
		case idx == 10:
			offsets = []float64{0, 0, 0, 0}
		case idx == 11:
			offsets = []float64{-11, 9, -11, 9}
		case idx == 12:
			results = append(results, &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: int32(idx)}, DpsMetrics: &proto.DistributionMetrics{Avg: 900, Stdev: 10}})
			continue
		case idx == 13:
			offsets = []float64{-30, -10, -30, -10}
		default:
			offsets = []float64{-1, -1, -1, -1}
		}
		values := make([]float64, len(baselineValues))
		var sum float64
		for valueIdx, value := range baselineValues {
			values[valueIdx] = value + offsets[valueIdx]
			sum += offsets[valueIdx]
		}
		metrics := &proto.DistributionMetrics{Avg: 1000 + sum/float64(len(offsets)), Stdev: 10, AllValues: values}
		results = append(results, &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: int32(idx)}, DpsMetrics: metrics})
	}
	return results, baseline
}

var mediumBulkSimStageConfig = BulkSimStageConfig{Stage: proto.BulkSimStage_BulkSimStageMedium, MinSurvivors: 5, MaxSurvivors: 25, CullingCoefficient: BulkSimCullingCoefficient}

func sortedBulkSimCandidateIndexes(candidates []BulkSimCandidate) []int32 {
	indexes := make([]int32, 0, len(candidates))
	for _, candidate := range candidates {
		indexes = append(indexes, candidate.Index)
	}
	slices.Sort(indexes)
	return indexes
}

func bulkSimGearResultIndexes(rows []*proto.BulkGearResult) []int32 {
	indexes := make([]int32, 0, len(rows))
	for _, row := range rows {
		indexes = append(indexes, row.CandidateIndex)
	}
	return indexes
}

func bulkSimIndexRange(from, to int32) []int32 {
	indexes := make([]int32, 0, to-from)
	for idx := from; idx < to; idx++ {
		indexes = append(indexes, idx)
	}
	return indexes
}

func TestCullBulkSimStageKeepsEveryCandidateNotClearlyBelowTheBaseline(t *testing.T) {
	results, baseline := baselineCullFixture()
	survivors, culled := cullBulkSimStage(&proto.BulkSimRequest{CullToBaseline: true}, results, baseline, 4, mediumBulkSimStageConfig, len(results), BulkSimDefaultTopResults)

	if got, want := sortedBulkSimCandidateIndexes(survivors), bulkSimIndexRange(0, 13); !slices.Equal(got, want) {
		t.Fatalf("survivors = %v, want %v", got, want)
	}
	if got, want := bulkSimGearResultIndexes(culled), bulkSimIndexRange(13, 30); !slices.Equal(got, want) {
		t.Fatalf("culled rows = %v, want %v", got, want)
	}
	for _, row := range culled {
		if row.CulledStage != proto.BulkSimStage_BulkSimStageMedium {
			t.Fatalf("culled row %d has stage %v, want the medium stage", row.CandidateIndex, row.CulledStage)
		}
		if len(row.DpsMetrics.AllValues) != 0 {
			t.Fatalf("culled row %d carries %d per-iteration values, want none", row.CandidateIndex, len(row.DpsMetrics.AllValues))
		}
		if got, want := row.DpsMetrics.Avg, results[row.CandidateIndex].DpsMetrics.Avg; got != want {
			t.Fatalf("culled row %d has mean %v, want the stage's %v", row.CandidateIndex, got, want)
		}
		wantPairedError := 0.0
		if row.CandidateIndex == 13 {
			wantPairedError = 5
		}
		if math.Abs(row.PairedErrorToBaseline-wantPairedError) > 1e-9 {
			t.Fatalf("culled row %d has a paired error to the baseline of %v, want %v", row.CandidateIndex, row.PairedErrorToBaseline, wantPairedError)
		}
	}
	if got := len(results[13].DpsMetrics.AllValues); got != 4 {
		t.Fatalf("the stage's result for candidate 13 keeps %d per-iteration values, want its 4", got)
	}
}

func TestCullBulkSimStageGivesEachCulledRowItsMeanPairedDifferenceToTheStageBaseline(t *testing.T) {
	results, baseline := baselineCullFixture()
	_, culled := cullBulkSimStage(&proto.BulkSimRequest{CullToBaseline: true}, results, baseline, 4, mediumBulkSimStageConfig, len(results), BulkSimDefaultTopResults)

	for _, row := range culled {
		// 13 trails each iteration by 30, 10, 30, 10, the rest by 1 each.
		want := -1.0
		if row.CandidateIndex == 13 {
			want = -20
		}
		if math.Abs(row.PairedDpsDiffToBaseline-want) > 1e-9 {
			t.Fatalf("culled row %d has a paired difference to the baseline of %v, want %v", row.CandidateIndex, row.PairedDpsDiffToBaseline, want)
		}
	}
}

// Each non-finalist sits on fixed per-iteration offsets to the high stage's baseline, so its paired difference is their
// mean: 2 for candidate 1 and -8 for candidate 2.
func TestAppendBulkSimNonFinalistsPairsEachRowWithTheHighStageBaseline(t *testing.T) {
	highBaselineValues := []float64{1000, 1010, 990, 1000}
	result := func(index int32, offsets []float64) *BulkSimCandidateResult {
		values := make([]float64, len(highBaselineValues))
		var sum float64
		for idx, value := range highBaselineValues {
			values[idx] = value + offsets[idx]
			sum += values[idx]
		}
		return &BulkSimCandidateResult{
			Candidate:  BulkSimCandidate{Index: index},
			DpsMetrics: &proto.DistributionMetrics{Avg: sum / float64(len(values)), Stdev: 10, AllValues: values},
		}
	}
	highBaseline := result(-1, []float64{0, 0, 0, 0})
	finalist := result(0, []float64{50, 50, 50, 50})
	highResults := []*BulkSimCandidateResult{finalist, result(1, []float64{6, -2, 4, 0}), result(2, []float64{-12, -4, -8, -8})}

	rows := appendBulkSimNonFinalists(nil, []*BulkSimCandidateResult{finalist}, highResults, highBaseline)

	if got, want := bulkSimGearResultIndexes(rows), []int32{1, 2}; !slices.Equal(got, want) {
		t.Fatalf("appended rows %v, want %v", got, want)
	}
	for idx, want := range []float64{2, -8} {
		if got := rows[idx].PairedDpsDiffToBaseline; math.Abs(got-want) > 1e-9 {
			t.Fatalf("non-finalist %d has a paired difference to the baseline of %v, want %v", rows[idx].CandidateIndex, got, want)
		}
		if rows[idx].PairedErrorToBaseline <= 0 {
			t.Fatalf("non-finalist %d has no paired error to the baseline", rows[idx].CandidateIndex)
		}
	}
}

func TestCullBulkSimStageWithoutTheFlagCullsAgainstTheLeader(t *testing.T) {
	results, baseline := baselineCullFixture()
	survivors, culled := cullBulkSimStage(&proto.BulkSimRequest{}, results, baseline, 4, mediumBulkSimStageConfig, len(results), BulkSimDefaultTopResults)

	if culled != nil {
		t.Fatalf("culled %d rows without the flag, want none", len(culled))
	}
	indexes := sortedBulkSimCandidateIndexes(survivors)
	for idx := range int32(5) {
		if !slices.Contains(indexes, idx) {
			t.Fatalf("survivors %v miss candidate %d, one of the five best", indexes, idx)
		}
	}
	for idx := int32(5); idx <= 10; idx++ {
		if slices.Contains(indexes, idx) {
			t.Fatalf("survivors %v keep candidate %d, behind the leader with no paired error", indexes, idx)
		}
	}
}

// 1,200 unpaired candidates at 1000 - 0.001 x index, stdev 10 over 4 iterations, so each interval overlaps the leader's,
// and 300 paired ones 50 below a baseline at 890, 910, 890, 910. The low stage's cap for the 1,200 above the cut is
// 110 (100 x sqrt(1.2)), so the soft cap keeps the 220 best.
func TestCullBulkSimStageFallsBackToTheTopResultsCullAboveTheCut(t *testing.T) {
	low := BulkSimStageConfigs[0]
	if low.Stage != proto.BulkSimStage_BulkSimStageLow {
		t.Fatalf("the first stage config is %v, want the low stage", low.Stage)
	}
	baselineValues := []float64{890, 910, 890, 910}
	baseline := &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: -1}, DpsMetrics: &proto.DistributionMetrics{Avg: 900, Stdev: 10, AllValues: baselineValues}}
	results := make([]*BulkSimCandidateResult, 0, 1500)
	for idx := range 1200 {
		results = append(results, &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: int32(idx)}, DpsMetrics: &proto.DistributionMetrics{Avg: 1000 - 0.001*float64(idx), Stdev: 10}})
	}
	for idx := 1200; idx < 1500; idx++ {
		values := make([]float64, len(baselineValues))
		for valueIdx, value := range baselineValues {
			values[valueIdx] = value - 50
		}
		results = append(results, &BulkSimCandidateResult{Candidate: BulkSimCandidate{Index: int32(idx)}, DpsMetrics: &proto.DistributionMetrics{Avg: 850, Stdev: 10, AllValues: values}})
	}

	survivors, culled := cullBulkSimStage(&proto.BulkSimRequest{CullToBaseline: true}, results, baseline, 4, low, len(results), BulkSimDefaultTopResults)

	// The stage's input count, 1,500, would give 2 x 123 = 246.
	if want := 2 * getBulkSimStageMaxSurvivors(low, 1200, BulkSimDefaultTopResults); len(survivors) != want || want != 220 {
		t.Fatalf("kept %d survivors, want the soft cap of the 1,200 above the cut, 220 (computed %d)", len(survivors), want)
	}
	if got, want := sortedBulkSimCandidateIndexes(survivors), bulkSimIndexRange(0, 220); !slices.Equal(got, want) {
		t.Fatalf("survivors = %v, want %v", got, want)
	}
	if got, want := bulkSimGearResultIndexes(culled), bulkSimIndexRange(220, 1500); !slices.Equal(got, want) {
		t.Fatalf("culled rows run %d long, want candidates 220 to 1499 in order", len(got))
	}
	if culled[0].DpsMetrics.Avg <= baseline.DpsMetrics.Avg {
		t.Fatalf("the first culled row's mean %v is not above the baseline's %v", culled[0].DpsMetrics.Avg, baseline.DpsMetrics.Avg)
	}
}

func TestResolveBulkSimTopResults(t *testing.T) {
	cases := []struct {
		name    string
		request *proto.BulkSimRequest
		want    int
	}{
		{"no top results asked for gives the default", &proto.BulkSimRequest{}, 5},
		{"top results asked for are kept", &proto.BulkSimRequest{TopResults: 7}, 7},
		{"all results shows every candidate", &proto.BulkSimRequest{AllResults: true}, 40},
		{"cull to baseline ignores all results", &proto.BulkSimRequest{CullToBaseline: true, AllResults: true}, 5},
		{"cull to baseline ignores top results", &proto.BulkSimRequest{CullToBaseline: true, TopResults: 40}, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveBulkSimTopResults(tc.request, 40); got != tc.want {
				t.Fatalf("top results for 40 candidates = %d, want %d", got, tc.want)
			}
		})
	}
}

func chestBulkSimGear(attackPower int) *proto.EquipmentSpec {
	return statChestBulkSimGear(int32(931000+attackPower), map[int32]float64{int32(proto.Stat_StatAttackPower): float64(attackPower)})
}

func statChestBulkSimGear(id int32, stats map[int32]float64) *proto.EquipmentSpec {
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{
		Id:             id,
		Type:           proto.ItemType_ItemTypeChest,
		ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {Stats: stats}},
	}}})
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotChest+1)
	for slot := range items {
		items[slot] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: id}
	return &proto.EquipmentSpec{Items: items}
}

// A warrior in melee range wearing only a chest of the given attack power, so DPS rises with it. Each chest's id is
// 931000 + its attack power (AddToDatabase keeps the first row per id), so every test asking for one gets the same row.
func chestBulkSimRequest(baselineAttackPower int, candidateAttackPowers []int) *proto.BulkSimRequest {
	request := bulkSimTestRequest(12500, 0)
	player := request.BaseRequest.Raid.Parties[0].Players[0]
	player.DistanceFromTarget = 0
	player.Equipment = chestBulkSimGear(baselineAttackPower)
	for idx, attackPower := range candidateAttackPowers {
		request.Candidates = append(request.Candidates, &proto.BulkGearCandidate{Index: int32(idx), Gear: chestBulkSimGear(attackPower)})
	}
	return request
}

func bulkSimAttackPowers(from, to, step int) []int {
	attackPowers := make([]int, 0, (to-from)/step+1)
	for attackPower := from; attackPower <= to; attackPower += step {
		attackPowers = append(attackPowers, attackPower)
	}
	return attackPowers
}

// Candidate idx has 20 x (idx + 1) attack power, 20 to 800, and the baseline 670, between candidates 32 and 33.
func cullToBaselineRequest() *proto.BulkSimRequest {
	request := chestBulkSimRequest(670, bulkSimAttackPowers(20, 800, 20))
	request.CullToBaseline = true
	return request
}

func runCullToBaselineBulkSim(t *testing.T, request *proto.BulkSimRequest) *proto.BulkSimResult {
	t.Helper()
	result := runBulkSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("bulk sim failed: %s", result.Error.Message)
	}
	return result
}

func TestBulkSimCullToBaselineReturnsEveryCandidate(t *testing.T) {
	useCheapBulkSimStages(t)
	requireCullingPlanAtTheDefaultTopResults(t, 40)
	result := runCullToBaselineBulkSim(t, cullToBaselineRequest())

	if stages := bulkSimStagesRun(result); stages[0] != proto.BulkSimStage_BulkSimStageMedium {
		t.Fatalf("ran stages %v, want the medium stage first", stages)
	}
	medium := result.StageMetrics[0]
	if got := len(result.TopResults); got <= BulkSimDefaultTopResults || got > 25 || got != int(medium.Survivors) {
		t.Fatalf("got %d top results, want the medium stage's %d survivors, more than %d and at most 25", got, medium.Survivors, BulkSimDefaultTopResults)
	}
	for idx, row := range result.TopResults {
		if row.CulledStage != proto.BulkSimStage_BulkSimStageUnknown {
			t.Fatalf("top result %d has culled stage %v, want none", row.CandidateIndex, row.CulledStage)
		}
		// The finalists pair with the refined baseline the result returns.
		if want := row.DpsMetrics.Avg - result.Baseline.DpsMetrics.Avg; idx < BulkSimDefaultTopResults && math.Abs(row.PairedDpsDiffToBaseline-want) > 1e-6 {
			t.Fatalf("finalist %d has a paired difference of %v to the baseline, want %v", row.CandidateIndex, row.PairedDpsDiffToBaseline, want)
		}
	}
	for _, row := range result.CulledResults {
		if row.CulledStage != proto.BulkSimStage_BulkSimStageMedium {
			t.Fatalf("culled row %d has stage %v, want the medium stage", row.CandidateIndex, row.CulledStage)
		}
		if len(row.DpsMetrics.AllValues) != 0 {
			t.Fatalf("culled row %d carries %d per-iteration values, want none", row.CandidateIndex, len(row.DpsMetrics.AllValues))
		}
		if got := simmedIterations(row); got != medium.Iterations {
			t.Fatalf("culled row %d simmed %d iterations, want the medium stage's %d", row.CandidateIndex, got, medium.Iterations)
		}
		if gap := medium.BaselineAvgDps - row.DpsMetrics.Avg; gap <= BulkSimZ95*row.PairedErrorToBaseline {
			t.Fatalf("culled row %d trails the baseline by %v, not more than %v x its paired error %v", row.CandidateIndex, gap, BulkSimZ95, row.PairedErrorToBaseline)
		}
		if want := row.DpsMetrics.Avg - medium.BaselineAvgDps; math.Abs(row.PairedDpsDiffToBaseline-want) > 1e-6 {
			t.Fatalf("culled row %d has a paired difference of %v to the baseline, want %v against the medium stage's", row.CandidateIndex, row.PairedDpsDiffToBaseline, want)
		}
	}

	topIndexes := bulkSimGearResultIndexes(result.TopResults)
	culledIndexes := bulkSimGearResultIndexes(result.CulledResults)
	allIndexes := slices.Concat(topIndexes, culledIndexes)
	slices.Sort(allIndexes)
	if want := bulkSimIndexRange(0, 40); !slices.Equal(allIndexes, want) {
		t.Fatalf("top and culled rows together hold candidates %v, want each of 0 to 39 once", allIndexes)
	}
	for idx := int32(33); idx < 40; idx++ {
		if !slices.Contains(topIndexes, idx) {
			t.Fatalf("top results %v miss candidate %d, above the baseline", topIndexes, idx)
		}
	}
	for idx := range int32(10) {
		if !slices.Contains(culledIndexes, idx) {
			t.Fatalf("culled rows %v miss candidate %d, far below the baseline", culledIndexes, idx)
		}
	}
}

func TestBulkSimCullToBaselineReadsNeitherTopResultsNorAllResults(t *testing.T) {
	useCheapBulkSimStages(t)
	request := cullToBaselineRequest()
	request.AllResults = true
	request.TopResults = 40
	result := runCullToBaselineBulkSim(t, request)

	if stages := bulkSimStagesRun(result); stages[0] != proto.BulkSimStage_BulkSimStageMedium {
		t.Fatalf("ran stages %v, want the medium stage first, not the high stage alone", stages)
	}
	if got := len(result.TopResults) + len(result.CulledResults); got != 40 {
		t.Fatalf("got %d top and culled rows, want all 40 candidates", got)
	}
}

func TestBulkSimCullToBaselineWithEveryCandidateBelowTheBaselineStillSimsTheBaseline(t *testing.T) {
	useCheapBulkSimStages(t)
	requireCullingPlanAtTheDefaultTopResults(t, 28)
	request := chestBulkSimRequest(2000, bulkSimAttackPowers(20, 560, 20))
	request.CullToBaseline = true
	result := runCullToBaselineBulkSim(t, request)

	if len(result.TopResults) != 0 {
		t.Fatalf("got %d top results, want none", len(result.TopResults))
	}
	if len(result.CulledResults) != 28 {
		t.Fatalf("got %d culled rows, want all 28 candidates", len(result.CulledResults))
	}
	if stages := bulkSimStagesRun(result); stages[len(stages)-1] != proto.BulkSimStage_BulkSimStageHigh {
		t.Fatalf("ran stages %v, want the high stage last", stages)
	}
	if got, want := simmedIterations(result.Baseline), bulkSimHighStageFloor(); got != want {
		t.Fatalf("baseline simmed %d iterations, want the high stage floor %d", got, want)
	}
}

func TestBulkSimCullToBaselineReturnsASmallRunWhollyInTheTopResults(t *testing.T) {
	useCheapBulkSimStages(t)
	request := chestBulkSimRequest(670, bulkSimAttackPowers(600, 780, 20))
	request.CullToBaseline = true
	result := runCullToBaselineBulkSim(t, request)

	stages := bulkSimStagesRun(result)
	if stages[0] != proto.BulkSimStage_BulkSimStageHigh || len(stages) > 2 || len(stages) == 2 && stages[1] != proto.BulkSimStage_BulkSimStageFinalist {
		t.Fatalf("ran stages %v, want the high stage alone, then the finalist stage if it ran", stages)
	}
	if len(result.TopResults) != 10 {
		t.Fatalf("got %d top results, want all 10 candidates", len(result.TopResults))
	}
	if len(result.CulledResults) != 0 {
		t.Fatalf("got %d culled rows, want none", len(result.CulledResults))
	}
	if !slices.ContainsFunc(result.TopResults, func(row *proto.BulkGearResult) bool { return row.DpsMetrics.Avg < result.Baseline.DpsMetrics.Avg }) {
		t.Fatalf("no top result is below the baseline's %v, want the losses kept", result.Baseline.DpsMetrics.Avg)
	}
}

// Twelve chests of 700 attack power and 10 x (idx + 1) crit rating, ids 932000 + their crit rating. Crit rating changes
// which swings crit, so neighbours pair with an error wide enough that the high stage leaves the top five unresolved.
func TestBulkSimCullToBaselineRefinesOnlyTheDefaultTopResults(t *testing.T) {
	useCheapBulkSimStages(t)
	request := chestBulkSimRequest(670, nil)
	for idx := range 12 {
		critRating := 10 * (idx + 1)
		gear := statChestBulkSimGear(int32(932000+critRating), map[int32]float64{int32(proto.Stat_StatAttackPower): 700, int32(proto.Stat_StatCritRating): float64(critRating)})
		request.Candidates = append(request.Candidates, &proto.BulkGearCandidate{Index: int32(idx), Gear: gear})
	}
	request.CullToBaseline = true
	result := runCullToBaselineBulkSim(t, request)

	stages := bulkSimStagesRun(result)
	if len(stages) != 2 || stages[0] != proto.BulkSimStage_BulkSimStageHigh || stages[1] != proto.BulkSimStage_BulkSimStageFinalist {
		t.Fatalf("ran stages %v, want the high stage and then the finalist stage", stages)
	}
	if got := result.StageMetrics[1].InputGearSets; got != BulkSimDefaultTopResults {
		t.Fatalf("the finalist stage refined %d gear sets, want the default %d", got, BulkSimDefaultTopResults)
	}
	if len(result.TopResults) != 12 {
		t.Fatalf("got %d top results, want all 12 candidates", len(result.TopResults))
	}
	for _, row := range result.TopResults[BulkSimDefaultTopResults:] {
		if row.PairedErrorToBaseline <= 0 {
			t.Fatalf("top result %d has a paired error to the baseline of %v, want it paired with the high stage's baseline", row.CandidateIndex, row.PairedErrorToBaseline)
		}
		if got, want := simmedIterations(row), bulkSimHighStageFloor(); got != want {
			t.Fatalf("top result %d simmed %d iterations, want the high stage's %d", row.CandidateIndex, got, want)
		}
	}
}

func TestBulkSimCullToBaselineIsDeterministic(t *testing.T) {
	useCheapBulkSimStages(t)
	request := cullToBaselineRequest()
	first := runCullToBaselineBulkSim(t, googleProto.Clone(request).(*proto.BulkSimRequest))
	second := runCullToBaselineBulkSim(t, googleProto.Clone(request).(*proto.BulkSimRequest))

	if !googleProto.Equal(first.Baseline, second.Baseline) {
		t.Fatalf("the two runs' baselines differ")
	}
	for _, rows := range [][2][]*proto.BulkGearResult{{first.TopResults, second.TopResults}, {first.CulledResults, second.CulledResults}} {
		if len(rows[0]) != len(rows[1]) {
			t.Fatalf("the two runs return %d and %d rows", len(rows[0]), len(rows[1]))
		}
		for idx := range rows[0] {
			if !googleProto.Equal(rows[0][idx], rows[1][idx]) {
				t.Fatalf("row %d differs: candidate %d against candidate %d", idx, rows[0][idx].CandidateIndex, rows[1][idx].CandidateIndex)
			}
		}
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

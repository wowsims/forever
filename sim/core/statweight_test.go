package core

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

// A distribution over values, with the aggregate data a sim result carries.
func statWeightDistribution(values []float64) *proto.DistributionMetrics {
	var aggr aggregator
	for _, value := range values {
		aggr.add(value)
	}
	mean, stdev := aggr.meanAndStdDev()
	return &proto.DistributionMetrics{Avg: mean, Stdev: stdev, AllValues: values, AggregatorData: &proto.AggregatorData{N: int32(aggr.n), SumSq: aggr.sumSq}}
}

func statWeightMetrics(dps float64) *proto.UnitMetrics {
	distribution := func(v float64) *proto.DistributionMetrics {
		return statWeightDistribution([]float64{v, v})
	}
	return &proto.UnitMetrics{Dps: distribution(dps), Hps: distribution(0), Threat: distribution(0), Dtps: distribution(0), Tmi: distribution(0)}
}

func statWeightResult(dps float64) *proto.RaidSimResult {
	return &proto.RaidSimResult{RaidMetrics: &proto.RaidMetrics{Parties: []*proto.PartyMetrics{{Players: []*proto.UnitMetrics{statWeightMetrics(dps)}}}}}
}

// A stat's runs below and above the baseline; low is nil for a stat simulated above it only.
type statWeightStatRuns struct {
	statData  *proto.StatWeightsStatData
	low, high *proto.RaidSimResult
}

func isHitRatingRuns(runs statWeightStatRuns) bool {
	return runs.statData.UnitStat == int32(hitRatingUnitStat)
}

// The calc request for stat runs over the baseline's first iterations.
func pairedStatWeightCalc(requests *proto.StatWeightRequestsData, base *proto.RaidSimResult, runs []statWeightStatRuns) *proto.StatWeightsCalcRequest {
	baseline := statWeightValuesOf(base)
	calc := &proto.StatWeightsCalcRequest{
		EpReferenceStat:    requests.EpReferenceStat,
		HpsReferenceStat:   requests.HpsReferenceStat,
		ProbeIterations:    int32(len(baseline[0])),
		BaselineIterations: int32(len(baseline[0])),
		BaseChanceOfDeath:  base.RaidMetrics.Parties[0].Players[0].ChanceOfDeath,
	}
	for _, statRuns := range runs {
		pairs := &proto.StatWeightsStatPairs{StatData: statRuns.statData, High: pairStatWeightRun(baseline, statRuns.high, 0)}
		if statRuns.low != nil {
			pairs.Low = pairStatWeightRun(baseline, statRuns.low, 0)
		}
		calc.StatPairs = append(calc.StatPairs, pairs)
	}
	return calc
}

// Adds a later pass to calc, each stat's runs continuing its pairs where they stop. baseline holds the
// baseline's values over every pass, and base, when set, is the baseline's run in this pass.
func addStatWeightPass(calc *proto.StatWeightsCalcRequest, baseline statWeightRunValues, base *proto.RaidSimResult, runs []statWeightStatRuns) {
	if base != nil {
		calc.BaselineIterations += int32(len(statWeightValuesOf(base)[0]))
	}
	for _, statRuns := range runs {
		pairs := calc.StatPairs[slices.IndexFunc(calc.StatPairs, func(pairs *proto.StatWeightsStatPairs) bool {
			return pairs.StatData.UnitStat == statRuns.statData.UnitStat
		})]
		offset := int(pairs.High.Dps.N)
		if statRuns.low != nil {
			addStatWeightPairs(pairs.Low, pairStatWeightRun(baseline, statRuns.low, offset))
		}
		addStatWeightPairs(pairs.High, pairStatWeightRun(baseline, statRuns.high, offset))
	}
}

const syntheticBaselineDps = 1000.0

// Builds the stat weight requests and answers each sub-sim with the DPS a point of its stat is worth
// in dpsPerPoint, then computes the weights.
func runSyntheticStatWeights(t *testing.T, statsToWeigh []proto.Stat, pseudoStatsToWeigh []proto.PseudoStat, dpsPerPoint map[stats.UnitStat]float64) (*proto.StatWeightRequestsData, *proto.StatWeightsResult) {
	t.Helper()
	requests := buildStatWeightRequests(&proto.StatWeightsRequest{
		Player:             &proto.Player{},
		Encounter:          &proto.Encounter{},
		SimOptions:         &proto.SimOptions{Iterations: 10, RandomSeed: 1},
		StatsToWeigh:       statsToWeigh,
		PseudoStatsToWeigh: pseudoStatsToWeigh,
		EpReferenceStat:    proto.Stat_StatAgility,
		HpsReferenceStat:   proto.Stat_StatAgility,
	})

	var runs []statWeightStatRuns
	for _, request := range requests.StatSimRequests {
		stat := stats.UnitStatFromIdx(int(request.StatData.UnitStat))
		statRuns := statWeightStatRuns{
			statData: request.StatData,
			high:     statWeightResult(syntheticBaselineDps + dpsPerPoint[stat]*request.StatData.ModHigh),
		}
		if request.RequestLow != nil {
			statRuns.low = statWeightResult(syntheticBaselineDps + dpsPerPoint[stat]*request.StatData.ModLow)
		}
		runs = append(runs, statRuns)
	}

	return requests, computeStatWeights(pairedStatWeightCalc(requests, statWeightResult(syntheticBaselineDps), runs))
}

func protoUnitStat(unitStats *proto.UnitStats, stat stats.UnitStat) float64 {
	if stat.IsStat() {
		return unitStats.Stats[stat.StatIdx()]
	}
	return unitStats.PseudoStats[stat.PseudoStatIdx()]
}

func simulatesStat(requests *proto.StatWeightRequestsData, stat stats.UnitStat) bool {
	for _, request := range requests.StatSimRequests {
		if request.StatData.UnitStat == int32(stat) {
			return true
		}
	}
	return false
}

var (
	agilityUnitStat   = stats.UnitStatFromStat(stats.Agility)
	hitRatingUnitStat = stats.UnitStatFromStat(stats.HitRating)
	meleeHitUnitStat  = stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent)
	rangedHitUnitStat = stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent)
)

// A spec weighing melee and ranged hit percent, or a hunter weighing ranged hit percent alone, still
// simulates hit rating, and its weight is that sim's answer. Each percent sub-sim measures its own
// channel, so the rating is simulated directly to value all of its channels at once.
func TestStatWeightsSimulateHitRatingBesideItsPercents(t *testing.T) {
	for _, c := range []struct {
		name        string
		percents    []proto.PseudoStat
		dpsPerPoint map[stats.UnitStat]float64
		want        map[string]float64
	}{
		{"melee and ranged hit", []proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent},
			map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3, meleeHitUnitStat: 25, rangedHitUnitStat: 10},
			map[string]float64{"hit rating weight": 3, "hit rating EP": 3, "melee hit percent weight": 25, "ranged hit percent weight": 10}},
		{"ranged hit alone", []proto.PseudoStat{proto.PseudoStat_PseudoStatRangedHitPercent},
			map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 2, rangedHitUnitStat: 18},
			map[string]float64{"hit rating weight": 2}},
	} {
		t.Run(c.name, func(t *testing.T) {
			requests, result := runSyntheticStatWeights(t, []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating}, c.percents, c.dpsPerPoint)

			if !simulatesStat(requests, hitRatingUnitStat) {
				t.Fatal("no sub-sim for hit rating")
			}
			got := map[string]float64{
				"hit rating weight":         protoUnitStat(result.Dps.Weights, hitRatingUnitStat),
				"hit rating EP":             protoUnitStat(result.Dps.EpValues, hitRatingUnitStat),
				"melee hit percent weight":  protoUnitStat(result.Dps.Weights, meleeHitUnitStat),
				"ranged hit percent weight": protoUnitStat(result.Dps.Weights, rangedHitUnitStat),
			}
			for name, want := range c.want {
				if math.Abs(got[name]-want) > 1e-9 {
					t.Errorf("%s = %v, want %v", name, got[name], want)
				}
			}
		})
	}
}

// A percent the spec weighs without its rating leaves the rating unweighed.
func TestStatWeightsLeaveAnUnrequestedRatingAtZero(t *testing.T) {
	requests, result := runSyntheticStatWeights(t,
		[]proto.Stat{proto.Stat_StatAgility},
		[]proto.PseudoStat{proto.PseudoStat_PseudoStatMeleeHitPercent},
		map[stats.UnitStat]float64{agilityUnitStat: 1, meleeHitUnitStat: 25})

	if simulatesStat(requests, hitRatingUnitStat) {
		t.Error("a sub-sim for hit rating, which the request does not weigh")
	}
	if got := protoUnitStat(result.Dps.Weights, hitRatingUnitStat); got != 0 {
		t.Errorf("hit rating weight = %v, want 0", got)
	}
}

// A stat simulated below and above its baseline pools both sides' samples, and hit rating has only the
// side above.
func TestStatWeightsReportSampleCounts(t *testing.T) {
	_, result := runSyntheticStatWeights(t,
		[]proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		nil,
		map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3})

	perSide := float64(len(statWeightMetrics(syntheticBaselineDps).Dps.AllValues))
	if got := protoUnitStat(result.SampleCounts, agilityUnitStat); got != 2*perSide {
		t.Errorf("agility samples = %v, want %v", got, 2*perSide)
	}
	if got := protoUnitStat(result.SampleCounts, hitRatingUnitStat); got != perSide {
		t.Errorf("hit rating samples = %v, want %v", got, perSide)
	}
}

func statWeightRun(dps []float64) *proto.RaidSimResult {
	return statWeightRunWithHps(dps, make([]float64, len(dps)))
}

func statWeightRunWithHps(dps, hps []float64) *proto.RaidSimResult {
	zeros := make([]float64, len(dps))
	return &proto.RaidSimResult{RaidMetrics: &proto.RaidMetrics{Parties: []*proto.PartyMetrics{{Players: []*proto.UnitMetrics{{
		Dps:    statWeightDistribution(dps),
		Hps:    statWeightDistribution(hps),
		Threat: statWeightDistribution(zeros),
		Dtps:   statWeightDistribution(zeros),
		Tmi:    statWeightDistribution(zeros),
	}}}}}}
}

// A run pairs each of its iterations with the baseline's at the same seed, from the offset it
// continues at, and keeps its own chance of death.
func TestStatWeightPairsARunWithTheBaselineFromItsOffset(t *testing.T) {
	baseline := statWeightValuesOf(statWeightRunWithHps([]float64{1000, 1010, 990, 1005, 1020}, []float64{1, 2, 3, 4, 5}))
	run := statWeightRunWithHps([]float64{995, 1013}, []float64{6, 3})
	setChanceOfDeath(run, 0.25)

	got := pairStatWeightRun(baseline, run, 2)
	want := &proto.StatWeightsSidePairs{
		Dps:           &proto.StatWeightsPairedSums{N: 2, Sum: 5 + 8, SumSq: 5*5 + 8*8},
		Hps:           &proto.StatWeightsPairedSums{N: 2, Sum: 3 - 1, SumSq: 3*3 + 1*1},
		Tps:           &proto.StatWeightsPairedSums{N: 2},
		Dtps:          &proto.StatWeightsPairedSums{N: 2},
		Tmi:           &proto.StatWeightsPairedSums{N: 2},
		ChanceOfDeath: 0.25,
	}
	if !googleProto.Equal(got, want) {
		t.Errorf("pairs = %v, want %v", got, want)
	}
}

// The pairs of two passes add up to the pairs of one run over both, and keep the first pass's chance
// of death.
func TestStatWeightPairsFromTwoPassesAddUpToOneWindow(t *testing.T) {
	baseline := statWeightValuesOf(statWeightRunWithHps([]float64{1000, 1010, 990, 1005, 1020, 998}, []float64{1, 2, 3, 4, 5, 6}))
	dps, hps := []float64{1003, 1004, 991, 1011, 1017, 1002}, []float64{2, 2, 5, 1, 7, 6}
	long := statWeightRunWithHps(dps, hps)
	first, second := statWeightRunWithHps(dps[:4], hps[:4]), statWeightRunWithHps(dps[4:], hps[4:])
	setChanceOfDeath(long, 0.1)
	setChanceOfDeath(first, 0.1)
	setChanceOfDeath(second, 0.9)

	got := pairStatWeightRun(baseline, first, 0)
	addStatWeightPairs(got, pairStatWeightRun(baseline, second, 4))
	if want := pairStatWeightRun(baseline, long, 0); !googleProto.Equal(got, want) {
		t.Errorf("pairs of two passes = %v, want %v", got, want)
	}
}

var noisyDpsPerPoint = map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3}

// Agility, the reference, simulated below and above the baseline, and hit rating above it only.
func noisyStatWeightRequests() *proto.StatWeightRequestsData {
	return buildStatWeightRequests(&proto.StatWeightsRequest{
		Player:           &proto.Player{},
		Encounter:        &proto.Encounter{},
		SimOptions:       &proto.SimOptions{Iterations: 2, RandomSeed: 1},
		StatsToWeigh:     []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		EpReferenceStat:  proto.Stat_StatAgility,
		HpsReferenceStat: proto.Stat_StatAgility,
	})
}

// The baseline and every stat's runs over iterations [from, to). Each iteration has a DPS shared by
// all runs, and each run adds its stat's worth and noise of its own, so the paired differences vary.
func noisyStatWeightResults(requests *proto.StatWeightRequestsData, from, to int) (*proto.RaidSimResult, []statWeightStatRuns) {
	values := func(run int, shift float64) []float64 {
		var dps []float64
		for i := from; i < to; i++ {
			dps = append(dps, syntheticBaselineDps+float64((i*7)%13)+shift+float64(((i+1)*(run+3)*5)%7))
		}
		return dps
	}

	var runs []statWeightStatRuns
	for j, request := range requests.StatSimRequests {
		worth := noisyDpsPerPoint[stats.UnitStatFromIdx(int(request.StatData.UnitStat))]
		statRuns := statWeightStatRuns{
			statData: request.StatData,
			high:     statWeightRun(values(2*j+2, worth*request.StatData.ModHigh)),
		}
		if request.RequestLow != nil {
			statRuns.low = statWeightRun(values(2*j+1, worth*request.StatData.ModLow))
		}
		runs = append(runs, statRuns)
	}
	return statWeightRun(values(0, 0)), runs
}

func noisyStatWeightCalc(perSide int) *proto.StatWeightsCalcRequest {
	requests := noisyStatWeightRequests()
	base, runs := noisyStatWeightResults(requests, 0, perSide)
	return pairedStatWeightCalc(requests, base, runs)
}

// A stat whose run above the baseline repeats it has hit a hard cap and gets no weight. One whose
// differences only cancel out is weighed, as is one that moves threat or DTPS alone.
func TestStatWeightsSkipAStatWhoseRunRepeatsTheBaseline(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	base, runs := noisyStatWeightResults(requests, 0, perSide)
	baseDps := base.RaidMetrics.Parties[0].Players[0].Dps.AllValues
	hit := slices.IndexFunc(runs, isHitRatingRuns)

	runs[hit].high = statWeightRun(slices.Clone(baseDps))
	capped := computeStatWeights(pairedStatWeightCalc(requests, base, runs))
	if got := protoUnitStat(capped.SampleCounts, hitRatingUnitStat); got != 0 {
		t.Errorf("hard-capped hit rating has %v samples, want none", got)
	}
	if got := protoUnitStat(capped.Dps.Weights, hitRatingUnitStat); got != 0 {
		t.Errorf("hard-capped hit rating weight = %v, want 0", got)
	}

	cancelling := slices.Clone(baseDps)
	for i := range cancelling {
		cancelling[i] += float64(1 - 2*(i%2))
	}
	runs[hit].high = statWeightRun(cancelling)
	weighed := computeStatWeights(pairedStatWeightCalc(requests, base, runs))
	if got := protoUnitStat(weighed.SampleCounts, hitRatingUnitStat); got != perSide {
		t.Errorf("hit rating whose differences cancel out has %v samples, want %d", got, perSide)
	}

	for _, metric := range []string{"threat", "DTPS"} {
		moved := make([]float64, perSide)
		for i := range moved {
			moved[i] = float64(i % 3)
		}
		runs[hit].high = statWeightRun(slices.Clone(baseDps))
		player := runs[hit].high.RaidMetrics.Parties[0].Players[0]
		if metric == "threat" {
			player.Threat = statWeightDistribution(moved)
		} else {
			player.Dtps = statWeightDistribution(moved)
		}
		weighed := computeStatWeights(pairedStatWeightCalc(requests, base, runs))
		if got := protoUnitStat(weighed.SampleCounts, hitRatingUnitStat); got != perSide {
			t.Errorf("hit rating that moves only %s has %v samples, want %d", metric, got, perSide)
		}
	}
}

// A target of 0 or below, or one that is not a number, plans nothing.
func TestStatWeightTopUpPlanNeedsATargetAboveZero(t *testing.T) {
	calc := noisyStatWeightCalc(20)
	if plan := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: &proto.StatWeightsTarget{EpCi: 1e-3}}); len(plan.Stats) == 0 {
		t.Fatal("a target above 0 plans nothing, want a top-up")
	}
	for _, target := range []float64{0, -0.01, math.NaN()} {
		if plan := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: &proto.StatWeightsTarget{EpCi: target}}); len(plan.Stats) != 0 {
			t.Errorf("target %v plans %v, want nothing", target, plannedIterations(plan))
		}
	}
}

// The progress counts stay in the int32 range and keep their ratio past it.
func TestStatWeightProgressCountsFitInt32(t *testing.T) {
	if done, total := statWeightProgressCounts(250, 1000); done != 250 || total != 1000 {
		t.Errorf("progress 250/1000 = %v/%v, want it as is", done, total)
	}
	total := 3 * int64(math.MaxInt32)
	done, scaledTotal := statWeightProgressCounts(total/3, total)
	if scaledTotal != math.MaxInt32 || done <= 0 || math.Abs(float64(done)/float64(scaledTotal)-1.0/3) > 1e-6 {
		t.Errorf("progress %v/%v = %v/%v, want a third of MaxInt32", total/3, total, done, scaledTotal)
	}
}

func shownEpCi(result *proto.StatWeightsResult, stat stats.UnitStat) float64 {
	return StatWeightZ90 * protoUnitStat(result.Dps.EpValuesStdev, stat) / math.Sqrt(protoUnitStat(result.SampleCounts, stat))
}

func plannedIterations(plan *proto.StatWeightsTopUpPlan) map[stats.UnitStat]int32 {
	planned := make(map[stats.UnitStat]int32)
	for _, stat := range plan.Stats {
		planned[stats.UnitStatFromIdx(int(stat.UnitStat))] = stat.Iterations
	}
	return planned
}

// The plan sizes each stat for the target from its EP spread and pooled sides, caps it at a multiple
// of the first baseline run, runs the baseline as far as the furthest stat, and leaves out every stat
// already at or under the target.
func TestStatWeightTopUpPlanSizesEachStatForTheTarget(t *testing.T) {
	const perSide = 20
	calc := noisyStatWeightCalc(perSide)
	result := computeStatWeights(calc)
	agilityCi, hitCi := shownEpCi(result, agilityUnitStat), shownEpCi(result, hitRatingUnitStat)
	if agilityCi == 0 || hitCi == 0 || agilityCi == hitCi {
		t.Fatalf("shown ± agility %v, hit rating %v: want two different spreads", agilityCi, hitCi)
	}
	sized := func(stat stats.UnitStat, target, sides float64) int32 {
		return int32(math.Ceil(math.Pow(StatWeightZ90*protoUnitStat(result.Dps.EpValuesStdev, stat)/target, 2) / sides))
	}
	const uncapped = math.MaxInt32 / perSide
	plan := func(target float64, multiplier int32) *proto.StatWeightsTopUpPlan {
		return planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: &proto.StatWeightsTarget{EpCi: target, MaxIterationMultiplier: multiplier}})
	}

	t.Run("sized by the formula", func(t *testing.T) {
		target := min(agilityCi, hitCi) / 4
		got := plan(target, uncapped)
		want := map[stats.UnitStat]int32{agilityUnitStat: sized(agilityUnitStat, target, 2), hitRatingUnitStat: sized(hitRatingUnitStat, target, 1)}
		for stat, iterations := range want {
			if plannedIterations(got)[stat] != iterations {
				t.Errorf("stat %d planned %d iterations per side, want %d", stat, plannedIterations(got)[stat], iterations)
			}
		}
		if got.BaselineIterations != max(want[agilityUnitStat], want[hitRatingUnitStat]) {
			t.Errorf("baseline planned %d iterations, want the furthest stat's", got.BaselineIterations)
		}
	})

	t.Run("capped", func(t *testing.T) {
		const multiplier = 2
		got := plan(min(agilityCi, hitCi)/4, multiplier)
		for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
			if plannedIterations(got)[stat] != multiplier*perSide {
				t.Errorf("stat %d planned %d iterations per side, want the cap", stat, plannedIterations(got)[stat])
			}
		}
		if got.BaselineIterations != multiplier*perSide {
			t.Errorf("baseline planned %d iterations, want the cap", got.BaselineIterations)
		}
	})

	t.Run("capped by the first baseline run, the baseline counted over all of them", func(t *testing.T) {
		const multiplier = 3
		topped := googleProto.Clone(calc).(*proto.StatWeightsCalcRequest)
		topped.BaselineIterations += perSide
		got := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: topped, Target: &proto.StatWeightsTarget{EpCi: min(agilityCi, hitCi) / 4, MaxIterationMultiplier: multiplier}})
		for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
			if plannedIterations(got)[stat] != multiplier*perSide {
				t.Errorf("stat %d planned %d iterations per side, want the cap", stat, plannedIterations(got)[stat])
			}
		}
		done := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: topped, Target: &proto.StatWeightsTarget{EpCi: max(agilityCi, hitCi), MaxIterationMultiplier: multiplier}})
		if done.BaselineIterations != 2*perSide {
			t.Errorf("baseline holds %d iterations, want both runs' %d", done.BaselineIterations, 2*perSide)
		}
	})

	t.Run("uncapped when unset", func(t *testing.T) {
		target := min(agilityCi, hitCi) / 100
		got := plan(target, 0)
		want := map[stats.UnitStat]int32{agilityUnitStat: sized(agilityUnitStat, target, 2), hitRatingUnitStat: sized(hitRatingUnitStat, target, 1)}
		for stat, iterations := range want {
			if plannedIterations(got)[stat] != iterations {
				t.Errorf("stat %d planned %d iterations per side, want %d", stat, plannedIterations(got)[stat], iterations)
			}
		}
		if got.BaselineIterations != max(want[agilityUnitStat], want[hitRatingUnitStat]) {
			t.Errorf("baseline planned %d iterations, want the furthest stat's", got.BaselineIterations)
		}
	})

	t.Run("a stat at its target is left out", func(t *testing.T) {
		got := plannedIterations(plan((agilityCi+hitCi)/2, uncapped))
		wider, narrower := agilityUnitStat, hitRatingUnitStat
		if hitCi > agilityCi {
			wider, narrower = narrower, wider
		}
		if _, ok := got[narrower]; ok {
			t.Errorf("stat %d is under the target and still planned", narrower)
		}
		if _, ok := got[wider]; !ok {
			t.Errorf("stat %d is over the target and not planned", wider)
		}
	})

	t.Run("done once every stat is at its target", func(t *testing.T) {
		got := plan(max(agilityCi, hitCi), uncapped)
		if len(got.Stats) != 0 || got.BaselineIterations != perSide {
			t.Errorf("plan = %v, want no stats and the baseline's own count", got)
		}
	})

	t.Run("done once the passes have run", func(t *testing.T) {
		target := &proto.StatWeightsTarget{EpCi: min(agilityCi, hitCi) / 4, MaxIterationMultiplier: uncapped}
		after := func(passes int32) *proto.StatWeightsTopUpPlan {
			return planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: target, Passes: passes})
		}
		if got := after(StatWeightMaxTopUpPasses - 1); len(got.Stats) == 0 {
			t.Errorf("plan after %d passes = %v, want stats", StatWeightMaxTopUpPasses-1, got)
		}
		if got := after(StatWeightMaxTopUpPasses); len(got.Stats) != 0 {
			t.Errorf("plan after %d passes = %v, want no stats", StatWeightMaxTopUpPasses, got)
		}
	})
}

// A stat that stopped early pairs with the baseline's first iterations, so it weighs as if every run
// had stopped where it did; a later pass adds up to the weights one longer run gives.
func TestStatWeightsPairAStatThatStoppedEarlyWithTheBaselinePrefix(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	firstBase, firstRuns := noisyStatWeightResults(requests, 0, perSide)
	secondBase, secondRuns := noisyStatWeightResults(requests, perSide, 2*perSide)
	longBase, longRuns := noisyStatWeightResults(requests, 0, 2*perSide)

	// Hit rating keeps only its first iterations; agility and the baseline run on.
	longRuns[slices.IndexFunc(longRuns, isHitRatingRuns)] = firstRuns[slices.IndexFunc(firstRuns, isHitRatingRuns)]
	early := computeStatWeights(pairedStatWeightCalc(requests, longBase, longRuns))
	trimmed := computeStatWeights(pairedStatWeightCalc(requests, firstBase, firstRuns))
	twoPassCalc := pairedStatWeightCalc(requests, firstBase, firstRuns)
	addStatWeightPass(twoPassCalc, statWeightValuesOf(longBase), secondBase, slices.DeleteFunc(secondRuns, isHitRatingRuns))
	twoPass := computeStatWeights(twoPassCalc)

	same := func(name string, got, want *proto.StatWeightsResult, stat stats.UnitStat) {
		t.Helper()
		for field, pair := range map[string][2]*proto.UnitStats{
			"weight":  {got.Dps.Weights, want.Dps.Weights},
			"stdev":   {got.Dps.WeightsStdev, want.Dps.WeightsStdev},
			"samples": {got.SampleCounts, want.SampleCounts},
		} {
			if g, w := protoUnitStat(pair[0], stat), protoUnitStat(pair[1], stat); math.Abs(g-w) > 1e-9 {
				t.Errorf("%s stat %d %s = %v, want %v", name, stat, field, g, w)
			}
		}
	}
	if protoUnitStat(early.SampleCounts, hitRatingUnitStat) != perSide {
		t.Fatalf("hit rating samples = %v, want its own iterations", protoUnitStat(early.SampleCounts, hitRatingUnitStat))
	}
	same("stopped early", early, trimmed, hitRatingUnitStat)
	for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
		same("two passes", twoPass, early, stat)
	}
}

var attackPowerUnitStat = stats.UnitStatFromStat(stats.AttackPower)

// The weapon tester weighing attack power, its reference, in a test sim. It auto-shoots, so it holds
// a bow: an empty ranged slot deals NaN damage.
func weaponTesterStatWeightsRequest(iterations int32) *proto.StatWeightsRequest {
	return &proto.StatWeightsRequest{
		Player:           weaponTestRaid(proto.Race_RaceHuman, WeaponTestSlots(WeaponTestSwordID, 0, WeaponTestBowID), nil).Parties[0].Players[0],
		Encounter:        MakeSingleTargetEncounter(0),
		SimOptions:       &proto.SimOptions{Iterations: iterations, RandomSeed: 1, IsTest: true},
		StatsToWeigh:     []proto.Stat{proto.Stat_StatAttackPower},
		EpReferenceStat:  proto.Stat_StatAttackPower,
		HpsReferenceStat: proto.Stat_StatAttackPower,
	}
}

// A run continued at the seed after its last iteration, through the concurrent split and with labeled
// rands, repeats the iterations one longer run has at those seeds, so its pairs add up to the same weights.
func TestStatWeightTopUpRepeatsOneLongerRun(t *testing.T) {
	const probe, topUp = 7, 5
	requests := buildStatWeightRequests(weaponTesterStatWeightsRequest(2 * (probe + topUp)))
	run := func(simRequest *proto.RaidSimRequest, done, iterations int32) *proto.RaidSimResult {
		continued := googleProto.Clone(simRequest).(*proto.RaidSimRequest)
		continued.SimOptions.Iterations = iterations
		continued.SimOptions.RandomSeed += int64(done)
		return runSimConcurrent(continued, nil, simsignals.CreateSignals())
	}
	statRequest := requests.StatSimRequests[0]
	statRuns := func(done, iterations int32) []statWeightStatRuns {
		return []statWeightStatRuns{{
			statData: statRequest.StatData,
			low:      run(statRequest.RequestLow, done, iterations),
			high:     run(statRequest.RequestHigh, done, iterations),
		}}
	}

	longBase := run(requests.BaseRequest, 0, probe+topUp)
	long := pairedStatWeightCalc(requests, longBase, statRuns(0, probe+topUp))

	probeBase, topUpBase := run(requests.BaseRequest, 0, probe), run(requests.BaseRequest, probe, topUp)
	baseline := statWeightValuesOf(probeBase)
	for metric, values := range statWeightValuesOf(topUpBase) {
		baseline[metric] = append(baseline[metric], values...)
	}
	for metric, values := range statWeightValuesOf(longBase) {
		if !slices.Equal(baseline[metric], values) {
			t.Errorf("baseline metric %d per iteration = %v, want %v", metric, baseline[metric], values)
		}
	}
	continued := pairedStatWeightCalc(requests, probeBase, statRuns(0, probe))
	addStatWeightPass(continued, baseline, topUpBase, statRuns(probe, topUp))

	near := func(got, want float64) bool { return math.Abs(got-want) <= 1e-9*math.Abs(want) }
	for name, sides := range map[string][2]*proto.StatWeightsSidePairs{
		"below": {continued.StatPairs[0].Low, long.StatPairs[0].Low},
		"above": {continued.StatPairs[0].High, long.StatPairs[0].High},
	} {
		for metric, pair := range [][2]*proto.StatWeightsPairedSums{{sides[0].Dps, sides[1].Dps}, {sides[0].Hps, sides[1].Hps}, {sides[0].Tps, sides[1].Tps}, {sides[0].Dtps, sides[1].Dtps}, {sides[0].Tmi, sides[1].Tmi}} {
			if pair[0].N != pair[1].N || !near(pair[0].Sum, pair[1].Sum) || !near(pair[0].SumSq, pair[1].SumSq) {
				t.Errorf("attack power %s metric %d pairs = %v, want %v", name, metric, pair[0], pair[1])
			}
		}
	}

	got, want := computeStatWeights(continued), computeStatWeights(long)
	if protoUnitStat(want.Dps.Weights, attackPowerUnitStat) == 0 {
		t.Fatal("attack power has no weight, so the runs cannot be told apart")
	}
	for field, pair := range map[string][2]*proto.UnitStats{
		"weight":  {got.Dps.Weights, want.Dps.Weights},
		"stdev":   {got.Dps.WeightsStdev, want.Dps.WeightsStdev},
		"samples": {got.SampleCounts, want.SampleCounts},
	} {
		if g, w := protoUnitStat(pair[0], attackPowerUnitStat), protoUnitStat(pair[1], attackPowerUnitStat); !near(g, w) {
			t.Errorf("attack power %s = %v, want %v", field, g, w)
		}
	}
}

// With a target ±, the runs go on until every DPS EP shows a ± at or under it or its runs reach the
// cap, and a run that needs more iterations gets them.
func TestStatWeightsTargetModeRunsToTheTargetOrTheCap(t *testing.T) {
	const probePerSide, multiplier = StatWeightTargetProbeIterations / 2, 4
	weigh := func(target float64) *proto.StatWeightsResult {
		request := weaponTesterStatWeightsRequest(StatWeightTargetProbeIterations)
		request.StatsToWeigh = []proto.Stat{proto.Stat_StatAttackPower, proto.Stat_StatHitRating}
		if target > 0 {
			request.Target = &proto.StatWeightsTarget{EpCi: target, MaxIterationMultiplier: multiplier}
		}
		return runStatWeights(request, nil, simsignals.CreateSignals())
	}
	weighed := []stats.UnitStat{attackPowerUnitStat, hitRatingUnitStat}

	probe := weigh(0)
	target := math.Inf(1)
	for _, stat := range weighed {
		target = min(target, shownEpCi(probe, stat)/2)
	}
	if target == 0 {
		t.Fatal("a stat has no spread, so there is nothing to top up")
	}

	result := weigh(target)
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	for _, stat := range weighed {
		sides := map[stats.UnitStat]float64{attackPowerUnitStat: 2, hitRatingUnitStat: 1}[stat]
		samples := protoUnitStat(result.SampleCounts, stat)
		if samples <= protoUnitStat(probe.SampleCounts, stat) {
			t.Errorf("stat %d kept its probe's %v samples", stat, samples)
		}
		if ci := shownEpCi(result, stat); ci > target && samples != sides*probePerSide*multiplier {
			t.Errorf("stat %d shows ± %v over the target %v with %v samples, under the cap", stat, ci, target, samples)
		}
	}
}

var (
	intellectUnitStat  = stats.UnitStatFromStat(stats.Intellect)
	healingDpsPerPoint = map[stats.UnitStat]float64{agilityUnitStat: 1, hitRatingUnitStat: 3}
	healingHpsPerPoint = map[stats.UnitStat]float64{agilityUnitStat: 0.5, hitRatingUnitStat: 0.2, intellectUnitStat: 2}
)

// Agility and hit rating weighed with agility as the DPS reference and hpsReferenceStat as the HPS one.
func healingStatWeightsRequest(hpsReferenceStat proto.Stat) *proto.StatWeightsRequest {
	return &proto.StatWeightsRequest{
		Player:           &proto.Player{},
		Encounter:        &proto.Encounter{},
		SimOptions:       &proto.SimOptions{Iterations: 2, RandomSeed: 1},
		StatsToWeigh:     []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		EpReferenceStat:  proto.Stat_StatAgility,
		HpsReferenceStat: hpsReferenceStat,
	}
}

// The runs over perSide iterations with intellect as the HPS reference. Every run's DPS differs from
// the baseline's by its stat's worth alone, so each DPS ± is 0, and each run's HPS carries noise of
// its own.
func healingStatWeightCalc(perSide int) *proto.StatWeightsCalcRequest {
	requests := buildStatWeightRequests(healingStatWeightsRequest(proto.Stat_StatIntellect))
	run := func(index int, stat stats.UnitStat, mod float64) *proto.RaidSimResult {
		var dps, hps []float64
		for i := range perSide {
			dps = append(dps, syntheticBaselineDps+float64((i*7)%13)+healingDpsPerPoint[stat]*mod)
			hps = append(hps, syntheticBaselineDps+float64((i*5)%11)+healingHpsPerPoint[stat]*mod+float64(((i+1)*(index+3)*5)%7))
		}
		return statWeightRunWithHps(dps, hps)
	}

	var runs []statWeightStatRuns
	for j, request := range requests.StatSimRequests {
		stat := stats.UnitStatFromIdx(int(request.StatData.UnitStat))
		statRuns := statWeightStatRuns{statData: request.StatData, high: run(2*j+2, stat, request.StatData.ModHigh)}
		if request.RequestLow != nil {
			statRuns.low = run(2*j+1, stat, request.StatData.ModLow)
		}
		runs = append(runs, statRuns)
	}
	return pairedStatWeightCalc(requests, run(0, 0, 0), runs)
}

// The HPS reference stat is simulated, below and above its baseline like the DPS one, and one equal to
// the DPS reference stat adds no runs.
func TestStatWeightsSimulateTheHpsReferenceStat(t *testing.T) {
	set := buildStatWeightRequests(healingStatWeightsRequest(proto.Stat_StatIntellect))
	index := slices.IndexFunc(set.StatSimRequests, func(request *proto.StatWeightsStatRequestData) bool {
		return request.StatData.UnitStat == int32(intellectUnitStat)
	})
	if index < 0 {
		t.Fatal("no sub-sim for intellect, the HPS reference stat")
	}
	if statData := set.StatSimRequests[index].StatData; statData.ModLow >= 0 || statData.ModHigh <= 0 {
		t.Errorf("intellect modifiers %v and %v, want one below and one above", statData.ModLow, statData.ModHigh)
	}

	same := buildStatWeightRequests(healingStatWeightsRequest(proto.Stat_StatAgility))
	if len(same.StatSimRequests) != len(set.StatSimRequests)-1 || simulatesStat(same, intellectUnitStat) {
		t.Errorf("HPS reference equal to the DPS one simulates %d stats, want the weighed ones alone", len(same.StatSimRequests))
	}
	if same.HpsReferenceStat != proto.Stat_StatAgility {
		t.Errorf("HPS reference = %v, want agility", same.HpsReferenceStat)
	}
}

// HPS EP divides by the HPS reference stat's HPS weight.
func TestStatWeightsDivideHpsEpByTheHpsReference(t *testing.T) {
	calc := healingStatWeightCalc(20)
	for _, ref := range []stats.UnitStat{intellectUnitStat, agilityUnitStat} {
		calc.HpsReferenceStat = proto.Stat(ref)
		result := computeStatWeights(calc)
		want := math.Abs(protoUnitStat(result.Hps.Weights, hitRatingUnitStat) / protoUnitStat(result.Hps.Weights, ref))
		if got := protoUnitStat(result.Hps.EpValues, hitRatingUnitStat); want == 0 || math.Abs(got-want) > 1e-12 {
			t.Errorf("hit rating HPS EP against stat %d = %v, want %v", ref, got, want)
		}
	}
}

// A stat whose DPS ± is under the target and whose HPS ± is over it is planned only when HPS is
// targeted, and its HPS ± divides by the HPS reference stat's weight.
func TestStatWeightTopUpPlanSizesTargetedHpsByTheHpsReference(t *testing.T) {
	const perSide = 20
	calc := healingStatWeightCalc(perSide)
	result := computeStatWeights(calc)
	samples := protoUnitStat(result.SampleCounts, agilityUnitStat)
	hpsEpStdev := func(stat stats.UnitStat, ref stats.UnitStat) float64 {
		return protoUnitStat(result.Hps.WeightsStdev, stat) / math.Abs(protoUnitStat(result.Hps.Weights, ref))
	}
	agilityHpsCi := StatWeightZ90 * hpsEpStdev(agilityUnitStat, intellectUnitStat) / math.Sqrt(samples)
	if agilityHpsCi == 0 || !(shownEpCi(result, agilityUnitStat) < agilityHpsCi/4) {
		t.Fatalf("agility shows DPS ± %v and HPS ± %v: want the DPS one well under the HPS one", shownEpCi(result, agilityUnitStat), agilityHpsCi)
	}
	target := agilityHpsCi / 4
	plan := func(metrics []proto.StatWeightsTargetMetric) map[stats.UnitStat]int32 {
		return plannedIterations(planStatWeightTopUp(&proto.StatWeightsTopUpRequest{
			Calc:   calc,
			Target: &proto.StatWeightsTarget{EpCi: target, MaxIterationMultiplier: math.MaxInt32 / perSide, Metrics: metrics},
		}))
	}

	if got := plan(nil); len(got) != 0 {
		t.Errorf("DPS alone planned %v, want nothing: every DPS ± is 0", got)
	}

	got := plan([]proto.StatWeightsTargetMetric{proto.StatWeightsTargetMetric_StatWeightsTargetMetricDps, proto.StatWeightsTargetMetric_StatWeightsTargetMetricHps})
	sized := func(ref stats.UnitStat) int32 {
		return int32(math.Ceil(math.Pow(StatWeightZ90*hpsEpStdev(agilityUnitStat, ref)/target, 2) / 2))
	}
	if sized(intellectUnitStat) == sized(agilityUnitStat) {
		t.Fatal("intellect and agility have the same HPS weight, so the references cannot be told apart")
	}
	if got[agilityUnitStat] != sized(intellectUnitStat) {
		t.Errorf("agility planned %d iterations per side, want %d from the HPS reference's weight", got[agilityUnitStat], sized(intellectUnitStat))
	}
}

func setChanceOfDeath(run *proto.RaidSimResult, chance float64) {
	if run != nil {
		run.RaidMetrics.Parties[0].Players[0].ChanceOfDeath = chance
	}
}

// The chance of death weights compare the first runs alone, so runs topped up past them, and a
// baseline topped up further than a stat, leave them as they were.
func TestStatWeightsReadTheChanceOfDeathFromTheFirstRuns(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	probeBase, probeRuns := noisyStatWeightResults(requests, 0, perSide)
	setChanceOfDeath(probeBase, 0.1)
	for _, statRuns := range probeRuns {
		setChanceOfDeath(statRuns.high, 0.3)
		setChanceOfDeath(statRuns.low, 0.05)
	}
	calc := pairedStatWeightCalc(requests, probeBase, probeRuns)
	probeOnly := computeStatWeights(calc)
	hitModHigh := probeRuns[slices.IndexFunc(probeRuns, isHitRatingRuns)].statData.ModHigh
	if got, want := protoUnitStat(probeOnly.PDeath.Weights, hitRatingUnitStat), (0.3-0.1)/hitModHigh; math.Abs(got-want) > 1e-12 {
		t.Fatalf("hit rating chance of death weight = %v, want %v", got, want)
	}

	longBase, _ := noisyStatWeightResults(requests, 0, 2*perSide)
	topUpBase, topUpRuns := noisyStatWeightResults(requests, perSide, 2*perSide)
	setChanceOfDeath(topUpBase, 0.9)
	topUpRuns = slices.DeleteFunc(topUpRuns, isHitRatingRuns)
	for _, statRuns := range topUpRuns {
		setChanceOfDeath(statRuns.high, 0.9)
		setChanceOfDeath(statRuns.low, 0.9)
	}
	addStatWeightPass(calc, statWeightValuesOf(longBase), topUpBase, topUpRuns)

	toppedUp := computeStatWeights(calc)
	for _, stat := range []stats.UnitStat{agilityUnitStat, hitRatingUnitStat} {
		if got, want := protoUnitStat(toppedUp.PDeath.Weights, stat), protoUnitStat(probeOnly.PDeath.Weights, stat); got != want {
			t.Errorf("stat %d chance of death weight after top-ups = %v, want the first runs' %v", stat, got, want)
		}
	}
}

func TestStatWeightRequestsRunAtLeastOneIterationPerSide(t *testing.T) {
	requests := buildStatWeightRequests(&proto.StatWeightsRequest{
		Player:           &proto.Player{},
		Encounter:        &proto.Encounter{},
		SimOptions:       &proto.SimOptions{Iterations: 1, RandomSeed: 1},
		StatsToWeigh:     []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
		EpReferenceStat:  proto.Stat_StatAgility,
		HpsReferenceStat: proto.Stat_StatAgility,
	})
	if got := requests.BaseRequest.SimOptions.Iterations; got != 1 {
		t.Errorf("baseline runs %d iterations, want 1", got)
	}
	for _, request := range requests.StatSimRequests {
		for _, simRequest := range []*proto.RaidSimRequest{request.RequestLow, request.RequestHigh} {
			if simRequest != nil && simRequest.SimOptions.Iterations != 1 {
				t.Errorf("stat %d runs %d iterations, want 1", request.StatData.UnitStat, simRequest.SimOptions.Iterations)
			}
		}
	}
}

// A request with a target probes at a fixed count whatever its iteration count, and one without runs
// its count.
func TestStatWeightRequestsProbeATargetAtAFixedCount(t *testing.T) {
	for _, target := range []*proto.StatWeightsTarget{nil, {EpCi: 0.01}} {
		for _, iterations := range []int32{10, 100 * StatWeightTargetProbeIterations} {
			requests := buildStatWeightRequests(&proto.StatWeightsRequest{
				Player:           &proto.Player{},
				Encounter:        &proto.Encounter{},
				SimOptions:       &proto.SimOptions{Iterations: iterations, RandomSeed: 1},
				StatsToWeigh:     []proto.Stat{proto.Stat_StatAgility, proto.Stat_StatHitRating},
				EpReferenceStat:  proto.Stat_StatAgility,
				HpsReferenceStat: proto.Stat_StatAgility,
				Target:           target,
			})
			want := iterations / 2
			if target != nil {
				want = StatWeightTargetProbeIterations / 2
			}
			if got := requests.BaseRequest.SimOptions.Iterations; got != want {
				t.Errorf("target %v, %d requested: baseline runs %d iterations, want %d", target, iterations, got, want)
			}
			for _, request := range requests.StatSimRequests {
				for _, simRequest := range []*proto.RaidSimRequest{request.RequestLow, request.RequestHigh} {
					if simRequest != nil && simRequest.SimOptions.Iterations != want {
						t.Errorf("target %v, %d requested: stat %d runs %d iterations, want %d", target, iterations, request.StatData.UnitStat, simRequest.SimOptions.Iterations, want)
					}
				}
			}
		}
	}
}

// Pairs over more iterations than the baseline holds have nothing to pair with, so the weights are
// refused and the plan asks for nothing more.
func TestStatWeightsRefuseAStatRunLongerThanTheBaseline(t *testing.T) {
	const perSide = 12
	requests := noisyStatWeightRequests()
	longBase, longRuns := noisyStatWeightResults(requests, 0, 2*perSide)
	calc := pairedStatWeightCalc(requests, longBase, longRuns)
	calc.ProbeIterations, calc.BaselineIterations = perSide, perSide

	if result := computeStatWeights(calc); result.Error == nil {
		t.Error("weights computed for stat runs longer than the baseline, want an error")
	}
	if plan := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: &proto.StatWeightsTarget{EpCi: 1e-9}}); len(plan.Stats) != 0 {
		t.Errorf("plan = %v, want no stats", plan)
	}
}

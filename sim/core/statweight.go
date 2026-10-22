package core

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	googleProto "google.golang.org/protobuf/proto"
)

const DTPSReferenceStat = stats.Armor

type UnitStats struct {
	Stats       stats.Stats
	PseudoStats []float64
}

func NewUnitStats() UnitStats {
	return UnitStats{
		PseudoStats: make([]float64, stats.PseudoStatsLen),
	}
}
func (s *UnitStats) AddStat(stat stats.UnitStat, value float64) {
	if stat.IsStat() {
		s.Stats[stat.StatIdx()] += value
	} else {
		s.PseudoStats[stat.PseudoStatIdx()] += value
	}
}
func (s *UnitStats) Get(stat stats.UnitStat) float64 {
	if stat.IsStat() {
		return s.Stats[stat.StatIdx()]
	} else {
		return s.PseudoStats[stat.PseudoStatIdx()]
	}
}

func (s *UnitStats) ToProto() *proto.UnitStats {
	return &proto.UnitStats{
		Stats:       s.Stats.ToProtoArray(),
		PseudoStats: s.PseudoStats,
		ApiVersion:  GetCurrentProtoVersion(),
	}
}

type StatWeightValues struct {
	Weights       UnitStats
	WeightsStdev  UnitStats
	EpValues      UnitStats
	EpValuesStdev UnitStats
}

func NewStatWeightValues() StatWeightValues {
	return StatWeightValues{
		Weights:       NewUnitStats(),
		WeightsStdev:  NewUnitStats(),
		EpValues:      NewUnitStats(),
		EpValuesStdev: NewUnitStats(),
	}
}

func (swv *StatWeightValues) ToProto() *proto.StatWeightValues {
	return &proto.StatWeightValues{
		Weights:       swv.Weights.ToProto(),
		WeightsStdev:  swv.WeightsStdev.ToProto(),
		EpValues:      swv.EpValues.ToProto(),
		EpValuesStdev: swv.EpValuesStdev.ToProto(),
	}
}

type StatWeightsResult struct {
	Dps    StatWeightValues
	Hps    StatWeightValues
	Tps    StatWeightValues
	Dtps   StatWeightValues
	Tmi    StatWeightValues
	PDeath StatWeightValues

	SampleCounts UnitStats
}

func NewStatWeightsResult() *StatWeightsResult {
	return &StatWeightsResult{
		Dps:    NewStatWeightValues(),
		Hps:    NewStatWeightValues(),
		Tps:    NewStatWeightValues(),
		Dtps:   NewStatWeightValues(),
		Tmi:    NewStatWeightValues(),
		PDeath: NewStatWeightValues(),

		SampleCounts: NewUnitStats(),
	}
}

func (swr *StatWeightsResult) ToProto() *proto.StatWeightsResult {
	return &proto.StatWeightsResult{
		Dps:    swr.Dps.ToProto(),
		Hps:    swr.Hps.ToProto(),
		Tps:    swr.Tps.ToProto(),
		Dtps:   swr.Dtps.ToProto(),
		Tmi:    swr.Tmi.ToProto(),
		PDeath: swr.PDeath.ToProto(),

		SampleCounts: swr.SampleCounts.ToProto(),
	}
}

func buildStatWeightRequests(swr *proto.StatWeightsRequest) *proto.StatWeightRequestsData {
	if swr.Player.BonusStats == nil {
		swr.Player.BonusStats = &proto.UnitStats{}
	}
	if swr.Player.BonusStats.Stats == nil {
		swr.Player.BonusStats.Stats = make([]float64, stats.ProtoStatsLen)
	}
	if swr.Player.BonusStats.PseudoStats == nil {
		swr.Player.BonusStats.PseudoStats = make([]float64, stats.PseudoStatsLen)
	}

	raidProto := SinglePlayerRaidProto(swr.Player, swr.PartyBuffs, swr.RaidBuffs, swr.Debuffs)
	if swr.Raid != nil && len(swr.Raid.Parties) > 0 && swr.Raid.Parties[0] != nil && len(swr.Raid.Parties[0].Players) > 0 {
		raidProto = googleProto.Clone(swr.Raid).(*proto.Raid)
		raidProto.Parties[0].Players[0] = swr.Player
		raidProto.Buffs = swr.RaidBuffs
		raidProto.Debuffs = swr.Debuffs
		raidProto.Parties[0].Buffs = swr.PartyBuffs
	}
	raidProto.Tanks = swr.Tanks

	swr.SimOptions.SaveAllValues = true

	if swr.Target != nil {
		swr.SimOptions.Iterations = StatWeightTargetProbeIterations
	}

	// Cut in half since we're doing above and below separately.
	// This number needs to be the same for the baseline sim too, so that RNG lines up perfectly.
	// At least 1, since a sim split across workers fails on 0.
	swr.SimOptions.Iterations = max(swr.SimOptions.Iterations/2, 1)

	// Make sure an RNG seed is always set because it gives more consistent results.
	// When there is no user-supplied seed it needs to be a randomly-selected seed
	// though, so that run-run differences still exist.
	if swr.SimOptions.RandomSeed == 0 {
		swr.SimOptions.RandomSeed = time.Now().UnixNano()
	}

	// Reduce variance even more by using test-level RNG controls.
	swr.SimOptions.UseLabeledRands = true

	hpsReferenceStat := swr.HpsReferenceStat

	swBaseResponse := &proto.StatWeightRequestsData{
		BaseRequest: &proto.RaidSimRequest{
			Raid:       raidProto,
			Encounter:  swr.Encounter,
			SimOptions: swr.SimOptions,
		},
		EpReferenceStat:  swr.EpReferenceStat,
		HpsReferenceStat: hpsReferenceStat,
		StatSimRequests:  []*proto.StatWeightsStatRequestData{},
	}

	// Do half the iterations with a positive, and half with a negative value for better accuracy.
	const defaultStatMod = 10.0
	statModsLow := make([]float64, stats.UnitStatsLen)
	statModsHigh := make([]float64, stats.UnitStatsLen)

	// Make sure the reference stats are included.
	for _, referenceStat := range []proto.Stat{swr.EpReferenceStat, hpsReferenceStat} {
		statModsLow[referenceStat] = -defaultStatMod
		statModsHigh[referenceStat] = defaultStatMod
	}

	statsToWeigh := stats.ProtoArrayToStatsList(swr.StatsToWeigh)
	// An Armor weight is per point of armor from items, which the armor modifiers scale, so armor
	// they leave alone is weighed beside it.
	if slices.Contains(statsToWeigh, stats.Armor) && !slices.Contains(statsToWeigh, stats.ArmorNoMods) {
		statsToWeigh = append(statsToWeigh, stats.ArmorNoMods)
	}
	for _, s := range statsToWeigh {
		stat := stats.UnitStatFromStat(s)
		statMod := defaultStatMod
		if stat.EqualsStat(stats.Armor) || stat.EqualsStat(stats.BonusArmor) || stat.EqualsStat(stats.ArmorNoMods) || stat.EqualsStat(stats.ArmorPenetration) {
			statMod = defaultStatMod * 10
		}
		statModsHigh[stat] = statMod
		statModsLow[stat] = -statMod

		if stat.EqualsStat(stats.HitRating) {
			statModsLow[stats.HitRating] = 0
		} else if stat.EqualsStat(stats.ExpertiseRating) {
			statModsLow[stats.ExpertiseRating] = 0
		}
	}

	for _, s := range swr.PseudoStatsToWeigh {
		stat := stats.UnitStatFromPseudoStat(s)
		statName := proto.PseudoStat_name[int32(s)]

		// Scale down the stat increment depending on the type of PseudoStat
		statMod := defaultStatMod

		if stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent) ||
			stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent) ||
			stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatSpellHitPercent) {
			statMod /= HitRatingPerHitPercent
		} else if stat.EqualsPseudoStat(proto.PseudoStat_PseudoStatExpertisePercent) {
			statMod /= ExpertiseRatingPerExpertisePercent
		} else if strings.Contains(statName, "SchoolHit") {
			statMod /= HitRatingPerHitPercent
		} else if strings.Contains(statName, "Crit") {
			statMod /= CritRatingPerCritPercent
		} else if strings.Contains(statName, "Dps") {
			statMod *= 0.5
		} else {
			panic(fmt.Sprintf("Unsupported PseudoStat in stat weights request: %s", statName))
		}

		statModsHigh[stat] = statMod
		statModsLow[stat] = -statMod
	}

	for i := range statModsLow {
		stat := stats.UnitStatFromIdx(i)
		if statModsLow[stat] == 0 && statModsHigh[stat] == 0 {
			continue
		}

		var lowSimRequest *proto.RaidSimRequest
		if statModsLow[stat] != 0 {
			lowSimRequest = googleProto.Clone(swBaseResponse.BaseRequest).(*proto.RaidSimRequest)
			stat.AddToStatsProto(lowSimRequest.Raid.Parties[0].Players[0].BonusStats, statModsLow[stat])
		}

		highSimRequest := googleProto.Clone(swBaseResponse.BaseRequest).(*proto.RaidSimRequest)
		stat.AddToStatsProto(highSimRequest.Raid.Parties[0].Players[0].BonusStats, statModsHigh[stat])

		swBaseResponse.StatSimRequests = append(swBaseResponse.StatSimRequests, &proto.StatWeightsStatRequestData{
			StatData: &proto.StatWeightsStatData{
				UnitStat: int32(stat),
				ModLow:   statModsLow[stat],
				ModHigh:  statModsHigh[stat],
			},
			RequestLow:  lowSimRequest,
			RequestHigh: highSimRequest,
		})
	}

	return swBaseResponse
}

// gen_db -gen=go-to-ts writes these to the UI as STAT_WEIGHT_*, which the stat sweep's Target ± also runs by.
const (
	StatWeightZ90                   = 1.645 // The two-sided 90% interval, the ± shown behind each EP.
	StatWeightMaxTopUpPasses        = 3     // The first top-up, which sizes from the probe alone, and two that resize from the runs so far.
	StatWeightTargetProbeIterations = 1000  // The probe's iterations with a target, both sides together; the request's own count plays no part.
)

// The first player's per-iteration DPS, HPS, threat, DTPS and TMI of a run.
type statWeightRunValues [5][]float64

func statWeightValuesOf(run *proto.RaidSimResult) statWeightRunValues {
	player := run.RaidMetrics.Parties[0].Players[0]
	return statWeightRunValues{player.Dps.AllValues, player.Hps.AllValues, player.Threat.AllValues, player.Dtps.AllValues, player.Tmi.AllValues}
}

// The side's sums in statWeightRunValues order.
func statWeightSideSums(side *proto.StatWeightsSidePairs) [5]**proto.StatWeightsPairedSums {
	return [5]**proto.StatWeightsPairedSums{&side.Dps, &side.Hps, &side.Tps, &side.Dtps, &side.Tmi}
}

func statWeightPairedSums(aggr *aggregator) *proto.StatWeightsPairedSums {
	return &proto.StatWeightsPairedSums{N: int32(aggr.n), Sum: aggr.sum, SumSq: aggr.sumSq}
}

// Pairs each iteration of a stat run with the baseline's iteration offset+i, which ran at the same seed.
func pairStatWeightRun(baseline statWeightRunValues, run *proto.RaidSimResult, offset int) *proto.StatWeightsSidePairs {
	side := &proto.StatWeightsSidePairs{ChanceOfDeath: run.RaidMetrics.Parties[0].Players[0].ChanceOfDeath}
	values := statWeightValuesOf(run)
	for metric, sums := range statWeightSideSums(side) {
		var aggr aggregator
		for i, value := range values[metric] {
			aggr.add(value - baseline[metric][offset+i])
		}
		*sums = statWeightPairedSums(&aggr)
	}
	return side
}

// Adds a later pass's pairs to a side's. The chance of death stays the first run's: one number per run
// cannot pair by iteration, and the first runs of the baseline and of every stat share their iterations.
func addStatWeightPairs(side *proto.StatWeightsSidePairs, pass *proto.StatWeightsSidePairs) {
	passSums := statWeightSideSums(pass)
	for metric, sums := range statWeightSideSums(side) {
		aggr, passAggr := statWeightAggregator(*sums), statWeightAggregator(*passSums[metric])
		*sums = statWeightPairedSums(aggr.merge(&passAggr))
	}
}

func statWeightAggregator(sums *proto.StatWeightsPairedSums) aggregator {
	return aggregator{n: int(sums.N), sum: sums.Sum, sumSq: sums.SumSq}
}

// Whether every paired difference of the side is 0, as for a stat whose runs repeat the baseline.
func matchesBaseline(side *proto.StatWeightsSidePairs) bool {
	for _, sums := range statWeightSideSums(side) {
		if (*sums).SumSq != 0 {
			return false
		}
	}
	return true
}

func computeStatWeights(swcr *proto.StatWeightsCalcRequest) *proto.StatWeightsResult {
	result, err := calcStatWeights(swcr)
	if err != nil {
		return &proto.StatWeightsResult{Error: err}
	}
	return result.ToProto()
}

func calcStatWeights(swcr *proto.StatWeightsCalcRequest) (*StatWeightsResult, *proto.ErrorOutcome) {
	haveRefStat := false
	for _, statPairs := range swcr.StatPairs {
		if statPairs.StatData.UnitStat == int32(swcr.EpReferenceStat) {
			haveRefStat = true
			break
		}
	}
	if !haveRefStat {
		return nil, &proto.ErrorOutcome{Message: "No result for reference stat exists!"}
	}

	result := NewStatWeightsResult()

	for _, statPairs := range swcr.StatPairs {
		stat := stats.UnitStatFromIdx(int(statPairs.StatData.UnitStat))

		low, high := statPairs.Low, statPairs.High
		hasLowMeasurement := low != nil
		if high.Dps.N > swcr.BaselineIterations || low.GetDps().GetN() > swcr.BaselineIterations {
			return nil, &proto.ErrorOutcome{Message: "A stat run has more iterations than the baseline!"}
		}

		modLow, modHigh := statPairs.StatData.ModLow, statPairs.StatData.ModHigh

		// Check for hard caps. Hard caps will have results identical to the baseline because RNG is fixed.
		// When we find a hard-capped stat, just skip it (will return 0).
		if matchesBaseline(high) {
			continue
		}

		calcWeightResults := func(lowSums *proto.StatWeightsPairedSums, highSums *proto.StatWeightsPairedSums, weightResults *StatWeightValues) int {
			var lo aggregator
			if hasLowMeasurement {
				lo = statWeightAggregator(lowSums)
				lo.scale(1 / modLow)
			}

			hi := statWeightAggregator(highSums)
			hi.scale(1 / modHigh)

			var aggr = &hi
			if hasLowMeasurement {
				aggr = aggr.merge(&lo)
			}

			mean, stdev := aggr.meanAndStdDev()

			weightResults.Weights.AddStat(stat, mean)
			weightResults.WeightsStdev.AddStat(stat, stdev)
			return aggr.n
		}

		samples := calcWeightResults(low.GetDps(), high.Dps, &result.Dps)
		result.SampleCounts.AddStat(stat, float64(samples))
		calcWeightResults(low.GetHps(), high.Hps, &result.Hps)
		calcWeightResults(low.GetTps(), high.Tps, &result.Tps)
		calcWeightResults(low.GetDtps(), high.Dtps, &result.Dtps)
		calcWeightResults(low.GetTmi(), high.Tmi, &result.Tmi)

		meanLow := (low.GetChanceOfDeath() - swcr.BaseChanceOfDeath) / modLow
		meanHigh := (high.ChanceOfDeath - swcr.BaseChanceOfDeath) / modHigh

		pDeathAvg := meanHigh
		if hasLowMeasurement {
			pDeathAvg += meanLow
			pDeathAvg /= 2
		}

		result.PDeath.Weights.AddStat(stat, pDeathAvg)
		result.PDeath.WeightsStdev.AddStat(stat, 0)
	}

	referenceStat := stats.Stat(swcr.EpReferenceStat)

	// Compute EP results.
	for _, statData := range swcr.StatPairs {
		stat := stats.UnitStatFromIdx(int(statData.StatData.UnitStat))

		calcEpResults := func(weightResults *StatWeightValues, refStat stats.Stat) {
			if weightResults.Weights.Stats[refStat] == 0 {
				return
			}
			mean := weightResults.Weights.Get(stat) / weightResults.Weights.Stats[refStat]
			stdev := weightResults.WeightsStdev.Get(stat) / math.Abs(weightResults.Weights.Stats[refStat])
			weightResults.EpValues.AddStat(stat, math.Abs(mean))
			weightResults.EpValuesStdev.AddStat(stat, math.Abs(stdev))
		}

		calcEpResults(&result.Dps, referenceStat)
		calcEpResults(&result.Hps, stats.Stat(swcr.HpsReferenceStat))
		calcEpResults(&result.Tps, referenceStat)
		calcEpResults(&result.Dtps, DTPSReferenceStat)
		calcEpResults(&result.Tmi, DTPSReferenceStat)
		calcEpResults(&result.PDeath, DTPSReferenceStat)
	}

	return result, nil
}

// Per-side iteration totals that bring each stat's shown EP ± down to the target in every targeted
// metric, sized by the metric that needs the most. A metric is skipped for a stat whose weight or
// reference weight in it is 0, a stat at or under the target in all of them is left out, and the
// baseline runs as far as the furthest stat. The sides divide the total because the low and high
// runs pool their samples. A set multiplier caps each run at that multiple of the first pass's
// iterations, and the plan is empty once StatWeightMaxTopUpPasses passes have run or without a
// target above 0.
func planStatWeightTopUp(request *proto.StatWeightsTopUpRequest) *proto.StatWeightsTopUpPlan {
	target := request.Target
	calc := request.Calc
	plan := &proto.StatWeightsTopUpPlan{
		BaselineIterations: calc.BaselineIterations,
	}
	if target == nil || !(target.EpCi > 0) || request.Passes >= StatWeightMaxTopUpPasses {
		return plan
	}
	result, err := calcStatWeights(calc)
	if err != nil {
		return plan
	}

	// Without a multiplier the int32 range still bounds the count, so the conversion below stays defined.
	maxIterationsPerSide := float64(math.MaxInt32)
	if target.MaxIterationMultiplier > 0 {
		maxIterationsPerSide = min(maxIterationsPerSide, float64(calc.ProbeIterations)*float64(target.MaxIterationMultiplier))
	}
	targetMetrics := target.Metrics
	if len(targetMetrics) == 0 {
		targetMetrics = []proto.StatWeightsTargetMetric{proto.StatWeightsTargetMetric_StatWeightsTargetMetricDps}
	}
	metricValues := func(metric proto.StatWeightsTargetMetric) (*StatWeightValues, stats.Stat) {
		switch metric {
		case proto.StatWeightsTargetMetric_StatWeightsTargetMetricHps:
			return &result.Hps, stats.Stat(calc.HpsReferenceStat)
		case proto.StatWeightsTargetMetric_StatWeightsTargetMetricTps:
			return &result.Tps, stats.Stat(calc.EpReferenceStat)
		default:
			return &result.Dps, stats.Stat(calc.EpReferenceStat)
		}
	}

	for _, statPairs := range calc.StatPairs {
		stat := stats.UnitStatFromIdx(int(statPairs.StatData.UnitStat))
		samples := result.SampleCounts.Get(stat)
		epStdev := 0.0
		for _, metric := range targetMetrics {
			values, refStat := metricValues(metric)
			refWeight := values.Weights.Stats[refStat]
			if values.Weights.Get(stat) == 0 || refWeight == 0 {
				continue
			}
			metricEpStdev := values.WeightsStdev.Get(stat) / math.Abs(refWeight)
			if StatWeightZ90*metricEpStdev/math.Sqrt(samples) > target.EpCi {
				epStdev = max(epStdev, metricEpStdev)
			}
		}
		if epStdev == 0 {
			continue
		}

		sides := 1.0
		if statPairs.Low != nil {
			sides = 2
		}
		current := statPairs.High.Dps.N
		needed := int32(math.Min(math.Ceil(math.Pow(StatWeightZ90*epStdev/target.EpCi, 2)/sides), maxIterationsPerSide))
		if needed <= current {
			continue
		}

		plan.Stats = append(plan.Stats, &proto.StatWeightsStatIterations{UnitStat: statPairs.StatData.UnitStat, Iterations: needed})
		plan.BaselineIterations = max(plan.BaselineIterations, needed)
	}
	return plan
}

// The iteration counts a progress message carries as int32, both scaled by one factor once the total
// passes the int32 range, which a plan of up to MaxInt32 iterations per side reaches over two stats.
func statWeightProgressCounts(done int64, total int64) (int32, int32) {
	if total <= math.MaxInt32 {
		return int32(done), int32(total)
	}
	return int32(float64(done) * math.MaxInt32 / float64(total)), math.MaxInt32
}

// Run stat weight sims and compute weights.
func runStatWeights(request *proto.StatWeightsRequest, progress chan *proto.ProgressMetrics, signals simsignals.Signals) *proto.StatWeightsResult {
	requestData := buildStatWeightRequests(request)

	var iterationsTotal int64 = int64(requestData.BaseRequest.SimOptions.Iterations)
	var iterationsDone int64 = 0
	var simsTotal int32 = 1
	var simsCompleted int32 = 0

	for _, reqData := range requestData.StatSimRequests {
		if reqData.RequestLow != nil {
			iterationsTotal += int64(reqData.RequestLow.SimOptions.Iterations)
			simsTotal += 1
		}
		iterationsTotal += int64(reqData.RequestHigh.SimOptions.Iterations)
		simsTotal += 1
	}

	waitForResult := func(srcProgressChannel chan *proto.ProgressMetrics) *proto.RaidSimResult {
		var lastCompleted int32 = 0
		for metrics := range srcProgressChannel {
			iterationsDone += int64(metrics.CompletedIterations - lastCompleted)
			lastCompleted = metrics.CompletedIterations

			if progress != nil {
				completedIterations, totalIterations := statWeightProgressCounts(iterationsDone, iterationsTotal)
				progress <- &proto.ProgressMetrics{
					TotalIterations:     totalIterations,
					CompletedIterations: completedIterations,
					CompletedSims:       simsCompleted,
					TotalSims:           simsTotal,
				}
			}

			if metrics.FinalRaidResult != nil {
				simsCompleted++
				return metrics.FinalRaidResult
			}
		}
		return nil
	}

	simFunc := runSimConcurrent
	// Don't use go threads in wasm, it just adds more overhead and makes the worker more unresponsive.
	if IsRunningInWasm() {
		simFunc = RunSim
	}

	baseProgress := make(chan *proto.ProgressMetrics, 100)
	go simFunc(requestData.BaseRequest, baseProgress, signals)
	baselineResult := waitForResult(baseProgress)
	if baselineResult.Error != nil {
		return &proto.StatWeightsResult{Error: baselineResult.Error}
	}

	// Only the baseline keeps its per-iteration values, for the stat runs of later passes to pair with.
	baselineValues := statWeightValuesOf(baselineResult)
	calc := &proto.StatWeightsCalcRequest{
		EpReferenceStat:    requestData.EpReferenceStat,
		HpsReferenceStat:   requestData.HpsReferenceStat,
		ProbeIterations:    int32(len(baselineValues[0])),
		BaselineIterations: int32(len(baselineValues[0])),
		BaseChanceOfDeath:  baselineResult.RaidMetrics.Parties[0].Players[0].ChanceOfDeath,
	}

	// Iteration i runs at seed RandomSeed+i, so a run continued at the seed after its last iteration
	// repeats what one longer run would have simulated.
	continueRun := func(simRequest *proto.RaidSimRequest, done int32, iterations int32) *proto.RaidSimResult {
		simOptions := googleProto.Clone(simRequest.SimOptions).(*proto.SimOptions)
		simOptions.Iterations = iterations
		simOptions.RandomSeed += int64(done)
		continued := &proto.RaidSimRequest{
			RequestId:  simRequest.RequestId,
			Raid:       simRequest.Raid,
			Encounter:  simRequest.Encounter,
			SimOptions: simOptions,
			Type:       simRequest.Type,
		}
		continuedProgress := make(chan *proto.ProgressMetrics, 100)
		go simFunc(continued, continuedProgress, signals)
		return waitForResult(continuedProgress)
	}

	type statWeightStatRun struct {
		reqData *proto.StatWeightsStatRequestData
		pairs   *proto.StatWeightsStatPairs
	}
	statRuns := make([]statWeightStatRun, 0, len(requestData.StatSimRequests))
	statRunOf := func(unitStat int32) statWeightStatRun {
		return statRuns[slices.IndexFunc(statRuns, func(run statWeightStatRun) bool { return run.reqData.StatData.UnitStat == unitStat })]
	}

	pairSide := func(simRequest *proto.RaidSimRequest, side **proto.StatWeightsSidePairs, done int32, iterations int32) *proto.ErrorOutcome {
		result := continueRun(simRequest, done, iterations)
		if result.Error != nil {
			return result.Error
		}
		pass := pairStatWeightRun(baselineValues, result, int(done))
		if *side == nil {
			*side = pass
		} else {
			addStatWeightPairs(*side, pass)
		}
		return nil
	}
	runStat := func(run statWeightStatRun, done int32, iterations int32) *proto.ErrorOutcome {
		if run.reqData.RequestLow != nil {
			if err := pairSide(run.reqData.RequestLow, &run.pairs.Low, done, iterations); err != nil {
				return err
			}
		}
		return pairSide(run.reqData.RequestHigh, &run.pairs.High, done, iterations)
	}

	for _, reqData := range requestData.StatSimRequests {
		run := statWeightStatRun{reqData: reqData, pairs: &proto.StatWeightsStatPairs{StatData: reqData.StatData}}
		if err := runStat(run, 0, reqData.RequestHigh.SimOptions.Iterations); err != nil {
			return &proto.StatWeightsResult{Error: err}
		}
		statRuns = append(statRuns, run)
		calc.StatPairs = append(calc.StatPairs, run.pairs)
	}

	if request.Target == nil {
		return computeStatWeights(calc)
	}

	for passes := int32(0); ; passes++ {
		plan := planStatWeightTopUp(&proto.StatWeightsTopUpRequest{Calc: calc, Target: request.Target, Passes: passes})
		if len(plan.Stats) == 0 {
			break
		}

		baselineDelta := plan.BaselineIterations - calc.BaselineIterations
		if baselineDelta > 0 {
			iterationsTotal += int64(baselineDelta)
			simsTotal++
		}
		for _, planned := range plan.Stats {
			run := statRunOf(planned.UnitStat)
			runs := TernaryInt32(run.reqData.RequestLow != nil, 2, 1)
			iterationsTotal += int64(runs) * int64(planned.Iterations-run.pairs.High.Dps.N)
			simsTotal += runs
		}

		if baselineDelta > 0 {
			baseRes := continueRun(requestData.BaseRequest, calc.BaselineIterations, baselineDelta)
			if baseRes.Error != nil {
				return &proto.StatWeightsResult{Error: baseRes.Error}
			}
			for metric, values := range statWeightValuesOf(baseRes) {
				baselineValues[metric] = append(baselineValues[metric], values...)
			}
			calc.BaselineIterations = int32(len(baselineValues[0]))
		}
		for _, planned := range plan.Stats {
			run := statRunOf(planned.UnitStat)
			done := run.pairs.High.Dps.N
			if err := runStat(run, done, planned.Iterations-done); err != nil {
				return &proto.StatWeightsResult{Error: err}
			}
		}
	}

	return computeStatWeights(calc)
}

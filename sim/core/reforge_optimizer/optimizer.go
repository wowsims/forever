package reforgeoptimizer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

var reforgeOptimizeRequestID atomic.Uint64

func OptimizeAsync(request *proto.ReforgeOptimizeRequest, signals simsignals.Signals) *proto.ReforgeOptimizeResult {
	requestID := reforgeOptimizeRequestID.Add(1)
	startedAt := time.Now()
	debug := request.GetDebug()
	if debug {
		log.Printf("[reforgeOptimize:%d] started debug=%t", requestID, debug)
		logRequestInput(requestID, request)
	}

	if request.Raid == nil || len(request.Raid.Parties) == 0 || len(request.Raid.Parties[0].Players) == 0 {
		log.Printf("[reforgeOptimize:%d] failed after %s: missing player", requestID, time.Since(startedAt))
		return optimizeError("Reforge optimizer requires a raid with player 0.")
	}
	if request.Raid.Parties[0].Players[0].Equipment == nil {
		log.Printf("[reforgeOptimize:%d] failed after %s: missing baseline gear", requestID, time.Since(startedAt))
		return optimizeError("Reforge optimizer requires baseline gear.")
	}
	if signals.Abort.IsTriggered() {
		return optimizeAborted()
	}

	solveStartedAt := time.Now()
	optimizer, err := newReforgeOptimizer(request, signals)
	var optimizedGear *proto.EquipmentSpec
	var score float64
	var passes int
	if err == nil {
		optimizedGear, score, passes, err = optimizer.optimizeReforges()
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("[reforgeOptimize:%d] aborted solving after %s", requestID, time.Since(startedAt))
			return optimizeAborted()
		}
		gear := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment()
		gearJSON, _ := protojson.Marshal(gear)
		log.Printf("[reforgeOptimize:%d] HiGHS failed after %s: %s gear=%s", requestID, time.Since(solveStartedAt), err.Error(), gearJSON)
		return optimizeError(fmt.Sprintf("HiGHS reforge optimizer failed: %s", err.Error()))
	}
	if debug {
		log.Printf("[reforgeOptimize:%d] HiGHS solved in %s score=%.3f", requestID, time.Since(solveStartedAt), score)
	}
	if signals.Abort.IsTriggered() {
		log.Printf("[reforgeOptimize:%d] aborted after solving in %s", requestID, time.Since(startedAt))
		return optimizeAborted()
	}
	log.Printf("[Reforge Optimizer] Reforge optimization completed requestID=%d total=%s score=%.3f", requestID, time.Since(startedAt), score)

	return &proto.ReforgeOptimizeResult{
		OptimizedGear: optimizedGear,
		Score:         score,
		PassesDone:    int32(passes),
	}
}

type reforgeOptimizer struct {
	request  *proto.ReforgeOptimizeRequest
	settings *proto.ReforgeSettings
	player   *proto.Player
	signals  simsignals.Signals

	frozenSlots    map[proto.ItemSlot]bool
	undershootCaps core.UnitStats

	// The baseline's build-phase stat dependency manager (ComputeStatDependencies). It resolves every
	// stat conversion the sim models, and resolveStatDelta reads it for each LP variable's cap-space
	// coefficients (the FULL dependency graph), apart from the EP-calibrated objective coefficients
	// applyReforgeStat produces.
	statDeps *stats.StatDependencyManager
	sheet    core.SheetAvoidance

	candidates []enchantCandidate
	baseGear   *proto.EquipmentSpec
	baseStats  core.UnitStats
}

// Builds the optimizer context from a clone of the request, whose raid becomes the baseline: the
// enchant candidates of the input gear, then the baseline's base stats and stat dependency manager.
// With no candidates there is nothing to solve, so the baseline's environment is not built.
func newReforgeOptimizer(request *proto.ReforgeOptimizeRequest, signals simsignals.Signals) (*reforgeOptimizer, error) {
	request = googleProto.Clone(request).(*proto.ReforgeOptimizeRequest)
	player := request.Raid.Parties[0].Players[0]
	if player.Database != nil {
		core.AddToDatabase(player.Database)
	}

	o := &reforgeOptimizer{
		request:  request,
		settings: request.GetSettings(),
		player:   player,
		signals:  signals,

		frozenSlots:    frozenItemSlots(request.GetSettings()),
		undershootCaps: protoToCoreUnitStats(request.GetUndershootCaps()),
		baseGear:       player.Equipment,
	}
	o.candidates = o.enchantCandidates(o.baseGear)
	if len(o.candidates) == 0 {
		return o, nil
	}

	// The baseline's environment build yields both its FinalStats and its finalized
	// StatDependencyManager.
	baseResult, baseSDM, sheet := computeReforgeStatsAndDeps(&proto.ComputeStatsRequest{Raid: request.Raid})
	if baseResult.ErrorResult != "" {
		return nil, errors.New(baseResult.ErrorResult)
	}
	if signals.Abort.IsTriggered() {
		return nil, context.Canceled
	}
	o.statDeps = baseSDM
	o.sheet = sheet
	o.baseStats = protoToCoreUnitStats(baseResult.RaidStats.Parties[0].Players[0].FinalStats)
	return o, nil
}

// Picks one static enchant per listed slot. The baseline is the input gear with those slots' static
// enchants cleared, so each candidate's coefficients are its full contribution. The score is the EP
// of the picks over that baseline, without the tie-break, not their gain over the input gear.
func (o *reforgeOptimizer) optimizeReforges() (*proto.EquipmentSpec, float64, int, error) {
	if len(o.candidates) == 0 {
		return o.baseGear, 0, 0, nil
	}

	reforgeCaps := computeStatCapsDelta(o.baseStats, protoToCoreUnitStats(o.settings.GetStatCaps()))
	var softCapConfigs []*proto.StatCapConfig
	if o.settings.GetUseSoftCapBreakpoints() {
		softCapConfigs = o.request.GetSoftCaps()
	}
	reforgeSoftCaps := computeReforgeSoftCaps(o.baseStats, softCapConfigs)
	weights := checkWeights(protoToCoreUnitStats(o.request.GetPreCapEpWeights()), reforgeCaps, reforgeSoftCaps)

	variables, constraints := o.buildEnchantModel(weights)
	selected, score, passes, err := o.solveModel(weights, reforgeCaps, reforgeSoftCaps, variables, constraints, optimizerTimeout.Seconds())
	if err != nil {
		return nil, 0, 0, err
	}
	return o.applyEnchantSolution(selected), score, passes, nil
}

func computeReforgeStatsAndDeps(request *proto.ComputeStatsRequest) (*proto.ComputeStatsResult, *stats.StatDependencyManager, core.SheetAvoidance) {
	request.SkipRotation = true
	return core.ComputeStatsAndDeps(request)
}

func optimizeError(message string) *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{Message: message},
	}
}

func optimizeAborted() *proto.ReforgeOptimizeResult {
	return &proto.ReforgeOptimizeResult{
		Error: &proto.ErrorOutcome{
			Type:    proto.ErrorOutcomeType_ErrorOutcomeAborted,
			Message: "Reforge optimization aborted.",
		},
	}
}

const (
	optimizerTimeout = 30 * time.Second
)

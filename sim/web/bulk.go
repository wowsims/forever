package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/wowsims/forever/sim/core/bulk"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Runs entirely on its own goroutine: candidate generation alone can take minutes for a
// large selection, and the HTTP handler has to hand the client its progress id first or
// there is nothing to poll and nothing to cancel.
func BulkSimAsync(request *proto.BulkSimRequest, progress chan *proto.ProgressMetrics, requestId string) {
	go runBulkSimAsync(request, progress, requestId)
}

// Terminal outcome: publishes the result and closes the progress channel.
func finishBulkSim(progress chan *proto.ProgressMetrics, stage proto.BulkSimStage, result *proto.BulkSimResult) {
	progress <- &proto.ProgressMetrics{BulkStage: stage, FinalBulkSimResult: result}
	close(progress)
}

func runBulkSimAsync(request *proto.BulkSimRequest, progress chan *proto.ProgressMetrics, requestId string) {
	defer func() {
		if err := recover(); err != nil {
			errStr := fmt.Sprint(err) + "\nStack Trace:\n" + string(debug.Stack())
			log.Printf("[ERROR] Bulk sim panicked: %s", errStr)
			finishBulkSim(progress, proto.BulkSimStage_BulkSimStageError, &proto.BulkSimResult{
				Error: &proto.ErrorOutcome{Message: errStr},
			})
		}
	}()

	// Registered before generation, not after: generation can take minutes, and the
	// client already holds its progress id, so an abort arriving in that window has to
	// land somewhere. bulk.BulkSimAsync registers the same id itself, so the id is
	// handed back before delegating to it.
	signals, err := simsignals.RegisterWithId(requestId)
	if err != nil {
		finishBulkSim(progress, proto.BulkSimStage_BulkSimStageError, &proto.BulkSimResult{
			Error: &proto.ErrorOutcome{Message: "Couldn't register for signal API: " + err.Error()},
		})
		return
	}
	registered := true
	unregister := func() {
		if registered {
			simsignals.UnregisterId(requestId)
			registered = false
		}
	}
	defer unregister()

	generatesCandidates := request.GetBulkSettings() != nil && len(request.GetCandidates()) == 0
	generationStartedAt := time.Now()
	if err := bulk.EnsureBulkSimCandidatesGenerated(request); err != nil {
		finishBulkSim(progress, proto.BulkSimStage_BulkSimStageError, &proto.BulkSimResult{
			Error: &proto.ErrorOutcome{Message: err.Error()},
		})
		return
	}
	if generatesCandidates {
		log.Printf("[Bulk Sim] Candidate generation completed total=%s candidates=%d", time.Since(generationStartedAt), len(request.GetCandidates()))
	}
	if signals.Abort.IsTriggered() {
		log.Printf("[Bulk Sim] Cancelled during candidate generation")
		finishBulkSim(progress, proto.BulkSimStage_BulkSimStageError, &proto.BulkSimResult{
			Error: &proto.ErrorOutcome{Type: proto.ErrorOutcomeType_ErrorOutcomeAborted},
		})
		return
	}
	unregister()
	bulk.BulkSimAsync(request, progress, requestId)
}

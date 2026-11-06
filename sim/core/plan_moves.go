package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A move of the plan with this iteration's start times.
type planMove struct {
	config        *proto.PlanMove
	skipped       bool
	occurrences   planOccurrences
	delayed       []PendingAction // Per player: the walk that waits for a hardcast.
	delayedQueued []bool
}

// Builds the moves' occurrence buffers and actions once, with the environment.
func (plan *encounterPlan) registerMoves(env *Environment) {
	players := env.Raid.AllPlayerUnits
	plan.walkEnds = make([]time.Duration, len(players))
	for index, config := range plan.config.Moves {
		move := &planMove{config: config}
		if config.Yards <= 0 && config.Seconds <= 0 {
			move.skipped = true
			plan.constructNotes = append(plan.constructNotes, fmt.Sprintf("move %d walks no yards and no seconds; skipped.", index+1))
		}
		move.occurrences = newPlanOccurrences(maxOccurrences(config.At, config.Repeat), ActionPriorityDOT+1, func(sim *Simulation) { plan.runMove(sim, move) })
		move.delayed = make([]PendingAction, len(players))
		move.delayedQueued = make([]bool, len(players))
		for i, player := range players {
			move.delayed[i] = PendingAction{
				Priority: ActionPriorityPrePull + 1,
				OnAction: func(sim *Simulation) {
					move.delayedQueued[i] = false
					plan.walk(sim, move, i, player)
				},
			}
		}
		plan.moves = append(plan.moves, move)
	}
}

// Schedules the iteration's moves: each one a timed walk of the sim's player, yards over its run speed or the
// move's seconds when yards is 0, that waits for a running hardcast, with its repeats rolled through
// rollOccurrences. Keeps nextMoveAt and nextMove on the next move not yet started.
func (plan *encounterPlan) scheduleMoves(sim *Simulation) {
	clear(plan.walkEnds)
	for _, move := range plan.moves {
		clear(move.delayedQueued)
		if !move.skipped {
			move.occurrences.schedule(sim, plan, move.config.At, move.config.Repeat)
		}
	}
	plan.updateMoveLookahead(sim)
}

// Walks every player of the raid for the move, and queues the move's next occurrence. A player whose
// hardcast is still queued from an earlier occurrence takes that walk only.
func (plan *encounterPlan) runMove(sim *Simulation, move *planMove) {
	move.occurrences.advance(sim)
	plan.updateMoveLookahead(sim)

	for i, player := range sim.Raid.AllPlayerUnits {
		if player.Hardcast.Expires > sim.CurrentTime && !player.Hardcast.CanMove {
			if !move.delayedQueued[i] {
				move.delayedQueued[i] = true
				move.delayed[i].NextActionAt = player.Hardcast.Expires
				sim.AddPendingAction(&move.delayed[i])
			}
		} else {
			plan.walk(sim, move, i, player)
		}
	}
}

func (plan *encounterPlan) walk(sim *Simulation, move *planMove, playerIndex int, player *Unit) {
	duration := planMoveDuration(move.config, player)
	plan.walkEnds[playerIndex] = max(plan.walkEnds[playerIndex], sim.CurrentTime+duration)
	player.MoveDurationByHand(duration, sim)
}

// Whether a move of the plan walks the raid's player of the index now, from the walk's start, after any hardcast it
// waits for, to its end.
func (plan *encounterPlan) walking(sim *Simulation, playerIndex int) bool {
	return sim.CurrentTime < plan.walkEnds[playerIndex]
}

func planMoveDuration(config *proto.PlanMove, player *Unit) time.Duration {
	if config.Yards > 0 {
		return DurationFromSeconds(config.Yards / player.GetMovementSpeed())
	}
	return DurationFromSeconds(config.Seconds)
}

// Points nextMoveAt and nextMove at the earliest move not yet started.
func (plan *encounterPlan) updateMoveLookahead(sim *Simulation) {
	plan.nextMoveAt = NeverExpires
	plan.nextMove = nil
	if len(sim.Raid.AllPlayerUnits) == 0 {
		return
	}
	for _, move := range plan.moves {
		if at := move.occurrences.nextAt(); at < plan.nextMoveAt {
			plan.nextMoveAt = at
			plan.nextMove = move
		}
	}
}

package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A move of the design with this iteration's start times.
type designMove struct {
	config        *proto.DesignMove
	skipped       bool
	occurrences   designOccurrences
	delayed       []PendingAction // Per player: the walk that waits for a hardcast.
	delayedQueued []bool
}

// Builds the moves' occurrence buffers and actions once, with the environment.
func (design *encounterDesign) registerMoves(env *Environment) {
	players := env.Raid.AllPlayerUnits
	for index, config := range design.config.Moves {
		move := &designMove{config: config}
		if config.Yards <= 0 && config.Seconds <= 0 {
			move.skipped = true
			design.constructNotes = append(design.constructNotes, fmt.Sprintf("move %d walks no yards and no seconds; skipped.", index+1))
		}
		move.occurrences = newDesignOccurrences(maxOccurrences(config.At, config.Repeat), ActionPriorityDOT+1, func(sim *Simulation) { design.runMove(sim, move) })
		move.delayed = make([]PendingAction, len(players))
		move.delayedQueued = make([]bool, len(players))
		for i, player := range players {
			move.delayed[i] = PendingAction{
				Priority: ActionPriorityPrePull + 1,
				OnAction: func(sim *Simulation) {
					move.delayedQueued[i] = false
					player.MoveDurationByHand(designMoveDuration(move.config, player), sim)
				},
			}
		}
		design.moves = append(design.moves, move)
	}
}

// Schedules the iteration's moves: each one a timed walk of the sim's player, yards over its run speed or the
// move's seconds when yards is 0, that waits for a running hardcast, with its repeats rolled through
// rollOccurrences. Keeps nextMoveAt and nextMove on the next move not yet started.
func (design *encounterDesign) scheduleMoves(sim *Simulation) {
	for _, move := range design.moves {
		clear(move.delayedQueued)
		if !move.skipped {
			move.occurrences.schedule(sim, design, move.config.At, move.config.Repeat)
		}
	}
	design.updateMoveLookahead(sim)
}

// Walks every player of the raid for the move, and queues the move's next occurrence. A player whose
// hardcast is still queued from an earlier occurrence takes that walk only.
func (design *encounterDesign) runMove(sim *Simulation, move *designMove) {
	move.occurrences.advance(sim)
	design.updateMoveLookahead(sim)

	for i, player := range sim.Raid.AllPlayerUnits {
		if player.Hardcast.Expires > sim.CurrentTime && !player.Hardcast.CanMove {
			if !move.delayedQueued[i] {
				move.delayedQueued[i] = true
				move.delayed[i].NextActionAt = player.Hardcast.Expires
				sim.AddPendingAction(&move.delayed[i])
			}
		} else {
			player.MoveDurationByHand(designMoveDuration(move.config, player), sim)
		}
	}
}

func designMoveDuration(config *proto.DesignMove, player *Unit) time.Duration {
	if config.Yards > 0 {
		return DurationFromSeconds(config.Yards / player.GetMovementSpeed())
	}
	return DurationFromSeconds(config.Seconds)
}

// Points nextMoveAt and nextMove at the earliest move not yet started.
func (design *encounterDesign) updateMoveLookahead(sim *Simulation) {
	design.nextMoveAt = NeverExpires
	design.nextMove = nil
	if len(sim.Raid.AllPlayerUnits) == 0 {
		return
	}
	for _, move := range design.moves {
		if at := move.occurrences.nextAt(); at < design.nextMoveAt {
			design.nextMoveAt = at
			design.nextMove = move
		}
	}
}

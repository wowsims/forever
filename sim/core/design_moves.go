package core

// Schedules the iteration's moves: each one a timed walk of the sim's player, yards over its run speed or the
// move's seconds when yards is 0, that waits for a running hardcast, with its repeats rolled through
// rollOccurrences. Keeps nextMoveAt and nextMoveDuration on the next move not yet started. The runner calls it
// only for a design with moves.
func (design *encounterDesign) scheduleMoves(_ *Simulation) {
	panic("not implemented")
}

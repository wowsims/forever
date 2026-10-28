package core

// Schedules the iteration's damage amp windows: a caster buff, an environment buff, or a target debuff on the
// design's chosen targets and the adds that spawn inside the window, each raising damage by the window's
// percent from its start to its end, with its repeats rolled through rollOccurrences. Keeps amps[type] on
// whether a window of the type runs and when the next one starts. The runner calls it only for a design with
// damage amps.
func (design *encounterDesign) scheduleAmps(_ *Simulation) {
	panic("not implemented")
}

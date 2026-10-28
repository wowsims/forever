package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

const (
	designMaxOccurrences = 1000 // Occurrences of one repeating event per iteration; a repeat until the fight ends stops there.
	designMaxPoolTargets = 64   // Targets one wave allocates; a wave whose count alone is more is skipped.
)

const numDamageAmpTypes = int(proto.DamageAmpType_DamageAmpTypeTargetDebuff) + 1

// Runs an encounter design's timeline each iteration: the boss's health, the targets' downtimes, the add waves
// and the lanes in design_moves.go, design_amps.go and design_boss_abilities.go. Every time on the timeline is
// a fraction of the rolled fight duration, and every jitter is rolled at reset from the sim's design rand, in
// lane order (downtimes, boss health, waves, moves, amps, boss abilities), index order inside a lane and
// occurrence order inside an event, so the timeline does not move with gear. A sim without a design has none.
type encounterDesign struct {
	config              *proto.EncounterDesign
	numEncounterTargets int      // The encounter's own targets, which come before every wave's pool.
	constructNotes      []string // What construction dropped from the design; logged at each reset, since no log exists before.

	bossKnots []HealthKnot
	downtimes []*designDowntime
	waves     []*designWave
	targets   []designTargetState // Indexed by Target.Index, over every target of the encounter.

	// Lookahead the movement and damage amp lanes keep current for the APL.
	nextMoveAt       time.Duration // The start of the next move not yet started, or NeverExpires.
	nextMoveDuration time.Duration // How long that move lasts, 0 without one.
	amps             [numDamageAmpTypes]designAmpLookahead
	lanes            designLanes // The movement and damage amp lanes' state, built with the environment.
}

type designAmpLookahead struct {
	active      bool          // Whether a window of the type is running.
	nextStartAt time.Duration // The start of the next window of the type not yet started, or NeverExpires.
}

// A target's downtimes are a union: it is down while any of them runs, and comes back when the last one ends,
// only if it was active when the first one took it down.
type designTargetState struct {
	wave            *designWave   // The wave whose pool holds the target, nil for the encounter's own targets.
	nextActiveAt    time.Duration // While downtimes hold the target: the latest end among them, or NeverExpires when it does not come back.
	activeDowntimes int
	latestEnd       time.Duration // The latest end among the running downtimes, NeverExpires for one that lasts the fight.
	backAfterDown   bool          // Whether the last downtime's end brings the target back.
	diesAt          time.Duration // The despawn of the add's current life.
	deathAction     PendingAction
	deathQueued     bool
}

type designDowntime struct {
	config *proto.DesignDowntime
	target *Target // nil when the index is not one of the encounter's own targets.
	start  time.Duration
	end    time.Duration
	down   bool // Whether the downtime has started this iteration.
	action PendingAction
}

const (
	downtimeEndPriority   = ActionPriorityDOT
	downtimeStartPriority = ActionPriorityDOT + 1 // A downtime that starts as another ends keeps its target down.
)

// An add wave, whose pool holds the most adds it has alive at once. Each spawn takes the free pool targets with
// the lowest index; a template TankIndex makes Enable pull that tank onto every add spawned. NextActiveAt and
// DiesAt of an idle pool target report the wave's next spawn, the same for every idle target of the wave.
type designWave struct {
	design    *encounterDesign
	index     int
	config    *proto.AddWave
	pool      []*Target
	count     int
	knots     [4]HealthKnot // An add's health over its life, from the wave's execute proportions.
	fractions []float64     // This iteration's spawns as fractions of the fight, in time order.
	spawns    []time.Duration
	next      int // The spawn still to come.
	capped    bool
	action    PendingAction
}

// The design of the encounter, with every wave's targets appended to the encounter, or nil when the design
// has no event in any lane. Only NewEncounter may call it, after its own targets.
func newEncounterDesign(encounter *Encounter, config *proto.EncounterDesign) *encounterDesign {
	if config == nil || (len(config.Downtimes) == 0 && len(config.BossHealth) == 0 && len(config.AddWaves) == 0 &&
		len(config.Moves) == 0 && len(config.DamageAmps) == 0 && len(config.BossAbilities) == 0) {
		return nil
	}

	design := &encounterDesign{
		config:              config,
		numEncounterTargets: len(encounter.AllTargets),
	}

	design.bossKnots = design.bossHealthKnots(config.BossHealth)

	for _, config := range config.Downtimes {
		downtime := &designDowntime{config: config}
		if index := int(config.TargetIndex); index >= 0 && index < design.numEncounterTargets {
			downtime.target = encounter.AllTargets[index]
		}
		downtime.action = PendingAction{
			OnAction: func(sim *Simulation) { sim.Encounter.design.runDowntime(sim, downtime) },
		}
		design.downtimes = append(design.downtimes, downtime)
	}

	for index, config := range config.AddWaves {
		wave := &designWave{design: design, index: index, config: config, count: int(config.Count)}
		p20 := Clamp(config.ExecuteProportion_20, 0, 1)
		p35 := max(Clamp(config.ExecuteProportion_35, 0, 1), p20)
		wave.knots = [4]HealthKnot{{At: 0, Percent: 100}, {At: 1 - p35, Percent: 35}, {At: 1 - p20, Percent: 20}, {At: 1, Percent: 0}}
		occurrences := maxOccurrences(config.GetSpawn(), config.Repeat)
		wave.fractions = make([]float64, 0, occurrences)
		wave.spawns = make([]time.Duration, 0, occurrences)
		wave.action = PendingAction{
			Priority: ActionPriorityDOT,
			OnAction: wave.spawn,
		}
		if skip := wave.skipReason(); skip != "" {
			design.constructNotes = append(design.constructNotes, fmt.Sprintf("wave %s %s; skipped.", wave.name(), skip))
		} else {
			poolSize := wave.count * maxConcurrentOccurrences(config, occurrences)
			if poolSize > designMaxPoolTargets {
				poolSize = designMaxPoolTargets
				wave.capped = true
			}
			wave.pool = encounter.allocatePoolTargets(config.Target, poolSize)
		}
		design.waves = append(design.waves, wave)
	}

	design.targets = make([]designTargetState, len(encounter.AllTargets))
	for _, wave := range design.waves {
		for _, target := range wave.pool {
			state := &design.targets[target.Index]
			state.wave = wave
			state.deathAction = PendingAction{
				Priority: ActionPriorityDOT + 1, // A death frees its target before a spawn at the same time.
				OnAction: func(sim *Simulation) { wave.despawn(sim, target) },
			}
		}
	}

	return design
}

// The boss's curve from the keyframes in time order, each one's time clamped into the fight, its health into
// 0 to 100 and to at most the keyframe's before it, since health only falls. The boss starts at full health
// and dies at the fight's end: a first keyframe after the pull gets (0, 100) before it, and a last one before
// the end still above 0 gets (1, 0) after it. Past the most knots a curve holds, the first ones that fit are
// kept.
func (design *encounterDesign) bossHealthKnots(keyframes []*proto.HealthKeyframe) []HealthKnot {
	if len(keyframes) == 0 {
		return nil
	}
	knots := make([]HealthKnot, 0, len(keyframes)+2)
	clamped := false
	for _, keyframe := range keyframes {
		knot := HealthKnot{At: Clamp(keyframe.At, 0, 1), Percent: Clamp(keyframe.HealthPercent, 0, 100)}
		clamped = clamped || knot.At != keyframe.At || knot.Percent != keyframe.HealthPercent
		knots = append(knots, knot)
	}
	slices.SortStableFunc(knots, func(a, b HealthKnot) int {
		switch {
		case a.At < b.At:
			return -1
		case a.At > b.At:
			return 1
		}
		return 0
	})
	for i := 1; i < len(knots); i++ {
		if knots[i].Percent > knots[i-1].Percent {
			knots[i].Percent = knots[i-1].Percent
			clamped = true
		}
	}
	if clamped {
		design.constructNotes = append(design.constructNotes,
			"boss health keyframes outside the fight, outside 0 to 100 or rising are clamped.")
	}
	if knots[0].At > 0 {
		knots = slices.Insert(knots, 0, HealthKnot{At: 0, Percent: 100})
	}
	if last := knots[len(knots)-1]; last.At < 1 && last.Percent > 0 {
		knots = append(knots, HealthKnot{At: 1, Percent: 0})
	}
	if len(knots) > maxHealthKnots {
		design.constructNotes = append(design.constructNotes, fmt.Sprintf(
			"the boss health curve holds %d knots, %d given; the first %d are kept.", maxHealthKnots, len(knots), maxHealthKnots))
		knots = knots[:maxHealthKnots]
	}
	return knots
}

// The most occurrences an event with the repeat has in one iteration: the first and count repeats, or with a
// count of 0 every repeat whose time before its jitter is inside the fight at the earliest first time.
func maxOccurrences(start *proto.DesignTime, repeat *proto.DesignRepeat) int {
	if repeat.GetInterval() <= 0 {
		return 1
	}
	if repeat.Count > 0 {
		return min(int(repeat.Count)+1, designMaxOccurrences)
	}
	repeats := (1 - (start.GetAt() - max(start.GetJitter(), 0))) / repeat.Interval
	return int(Clamp(repeats, 0, designMaxOccurrences-1)) + 1
}

// The most lives of a wave alive at once, with the repeat's jitter at its worst: a life lasts the lifetime,
// and two spawns are at least the interval less twice the repeat jitter apart.
func maxConcurrentOccurrences(config *proto.AddWave, occurrences int) int {
	if occurrences <= 1 {
		return 1
	}
	spacing := config.Repeat.Interval - 2*max(config.Repeat.Jitter, 0)
	if spacing <= 0 {
		return occurrences
	}
	alive := (max(config.Lifetime, 0) + 2*max(config.GetSpawn().GetJitter(), 0)) / spacing
	return min(int(alive)+1, occurrences)
}

func designRandSeed(rseed int64) int64 {
	return makeTestRandSeed(rseed, "Encounter Design")
}

// A fraction of the fight moved by up to jitter either way, drawn from the design rand; no draw without jitter.
func (design *encounterDesign) roll(sim *Simulation, at float64, jitter float64) float64 {
	if jitter <= 0 {
		return at
	}
	return at + jitter*(2*sim.designRand.NextFloat64()-1)
}

// The time of a fraction of the fight, clamped into the fight.
func designTime(sim *Simulation, fraction float64) time.Duration {
	return time.Duration(Clamp(fraction, 0, 1) * float64(sim.Duration))
}

// Rolls the occurrences of an event first at start: repeat k lands at the first occurrence plus k intervals,
// moved by its own jitter. Count repeats follow the first, or with a count of 0 every repeat whose time before
// its jitter is inside the fight. The fractions go into dst[:0] in time order, up to dst's capacity; size dst
// with maxOccurrences.
func (design *encounterDesign) rollOccurrences(sim *Simulation, start *proto.DesignTime, repeat *proto.DesignRepeat, dst []float64) []float64 {
	first := design.roll(sim, start.GetAt(), start.GetJitter())
	dst = append(dst[:0], first)
	if interval := repeat.GetInterval(); interval > 0 {
		for k := 1; len(dst) < cap(dst); k++ {
			at := first + float64(k)*interval
			if (repeat.Count > 0 && k > int(repeat.Count)) || (repeat.Count <= 0 && at > 1) {
				break
			}
			dst = append(dst, design.roll(sim, at, repeat.Jitter))
		}
	}
	slices.Sort(dst)
	return dst
}

// Sets the encounter back to the design's start, after the targets reset and before the raid does: every
// encounter target enabled unless it starts disabled, every wave's pool disabled at full health, the boss on
// its health curve, and each lane's events rolled and scheduled. Events at the pull run once the pull has
// started, so a target gone at 0 takes the raid's swings with it.
func (design *encounterDesign) reset(sim *Simulation) {
	if design == nil {
		return
	}
	encounter := &sim.Encounter
	if sim.Log != nil {
		for _, note := range design.constructNotes {
			sim.Log("Encounter design: %s", note)
		}
	}

	for _, target := range encounter.AllTargets {
		state := &design.targets[target.Index]
		state.nextActiveAt = NeverExpires
		state.activeDowntimes = 0
		state.latestEnd = 0
		state.backAfterDown = false
		state.diesAt = NeverExpires
		state.deathQueued = false

		config := encounter.targetConfigs[target.Index]
		target.enabled = state.wave == nil && (config == nil || !config.DisabledAtStart)
		target.CurrentTarget = nil
		if target.enabled {
			target.CurrentTarget = target.defaultTarget
		}
		if state.wave != nil {
			target.SetHealthCurve(sim, 0, 0, HealthKnot{At: 0, Percent: 100})
		}
	}
	encounter.resetActiveTargetOrder()

	if len(design.bossKnots) > 0 {
		encounter.AllTargets[0].SetHealthCurve(sim, 0, sim.Duration, design.bossKnots...)
	}

	for index, downtime := range design.downtimes {
		downtime.down = false
		if downtime.target == nil {
			if sim.Log != nil {
				sim.Log("Encounter design: downtime %d names target %d, which the encounter does not have; skipped.", index+1, downtime.config.TargetIndex)
			}
			continue
		}
		downtime.start = designTime(sim, design.roll(sim, downtime.config.GetStart().GetAt(), downtime.config.GetStart().GetJitter()))
		downtime.end = designTime(sim, design.roll(sim, downtime.config.GetEnd().GetAt(), downtime.config.GetEnd().GetJitter()))
		if downtime.start >= sim.Duration || downtime.end <= downtime.start {
			continue
		}
		downtime.action.NextActionAt = downtime.start
		downtime.action.Priority = downtimeStartPriority
		sim.AddPendingAction(&downtime.action)
	}

	for _, wave := range design.waves {
		wave.next = 0
		wave.spawns = wave.spawns[:0]
		if wave.pool == nil {
			continue
		}
		if wave.capped && sim.Log != nil {
			sim.Log("Encounter design: wave %s holds at most %d adds at once.", wave.name(), len(wave.pool))
		}
		wave.fractions = design.rollOccurrences(sim, wave.config.GetSpawn(), wave.config.Repeat, wave.fractions)
		for _, fraction := range wave.fractions {
			if at := designTime(sim, fraction); at < sim.Duration {
				wave.spawns = append(wave.spawns, at)
			}
		}
		if len(wave.spawns) > 0 {
			wave.action.NextActionAt = wave.spawns[0]
			sim.AddPendingAction(&wave.action)
		}
	}

	design.nextMoveAt = NeverExpires
	design.nextMoveDuration = 0
	for i := range design.amps {
		design.amps[i] = designAmpLookahead{nextStartAt: NeverExpires}
	}
	if len(design.config.Moves) > 0 {
		design.scheduleMoves(sim)
	}
	if len(design.config.DamageAmps) > 0 {
		design.scheduleAmps(sim)
	}
	if len(design.config.BossAbilities) > 0 {
		design.scheduleBossAbilities(sim)
	}
}

// Takes the downtime's target away at its start, clearing its debuffs when the downtime says so even with the
// target already down, and at its end, once no other downtime holds the target, brings it back if it was active
// when the first of them started. An end at or after the fight's end never comes.
func (design *encounterDesign) runDowntime(sim *Simulation, downtime *designDowntime) {
	target := downtime.target
	state := &design.targets[target.Index]
	if downtime.down {
		downtime.down = false
		state.activeDowntimes--
		back := state.activeDowntimes == 0 && state.backAfterDown
		if sim.Log != nil {
			target.Log(sim, "Downtime ends (target back: %t).", back)
		}
		if state.activeDowntimes > 0 {
			return
		}
		state.nextActiveAt = NeverExpires
		state.latestEnd = 0
		if back {
			target.Enable(sim)
			design.reapplyAmps(sim, target)
		}
		return
	}

	downtime.down = true
	if state.activeDowntimes == 0 {
		state.backAfterDown = target.IsEnabled()
	}
	state.activeDowntimes++
	end := NeverExpires
	if downtime.end < sim.Duration {
		end = downtime.end
		downtime.action.NextActionAt = end
		downtime.action.Priority = downtimeEndPriority
		sim.AddPendingAction(&downtime.action)
	}
	state.latestEnd = max(state.latestEnd, end)
	state.nextActiveAt = NeverExpires
	if state.backAfterDown {
		state.nextActiveAt = state.latestEnd
	}

	if sim.Log != nil {
		target.Log(sim, "Downtime starts (debuffs cleared: %t).", downtime.config.ClearDebuffs)
	}
	if target.IsEnabled() {
		target.Disable(sim, downtime.config.ClearDebuffs)
	} else if downtime.config.ClearDebuffs {
		target.auraTracker.expireAll(sim)
	}
}

// Why the wave spawns nothing, or "" when it spawns.
func (wave *designWave) skipReason() string {
	switch {
	case wave.config.Target == nil:
		return "has no target"
	case wave.count <= 0:
		return "spawns no adds"
	case wave.count > designMaxPoolTargets:
		return fmt.Sprintf("spawns %d adds at once, more than the %d a wave holds", wave.count, designMaxPoolTargets)
	case wave.config.Lifetime <= 0:
		return "has no lifetime"
	}
	return ""
}

func (wave *designWave) name() string {
	if wave.config.Name != "" {
		return wave.config.Name
	}
	return fmt.Sprintf("%d", wave.index+1)
}

func (wave *designWave) lifetime(sim *Simulation) time.Duration {
	return time.Duration(max(wave.config.Lifetime, 0) * float64(sim.Duration))
}

// The next spawn of the wave this iteration, or NeverExpires.
func (wave *designWave) nextSpawnAt() time.Duration {
	if wave.next < len(wave.spawns) {
		return wave.spawns[wave.next]
	}
	return NeverExpires
}

// Brings count free adds of the pool in, each on the wave's health curve over its lifetime and dying at its
// end when that comes before the fight ends. With fewer than count free, the spawn is skipped.
func (wave *designWave) spawn(sim *Simulation) {
	wave.next++
	if wave.next < len(wave.spawns) {
		wave.action.NextActionAt = wave.spawns[wave.next]
		sim.AddPendingAction(&wave.action)
	}

	free := 0
	for _, target := range wave.pool {
		if !target.IsEnabled() {
			free++
		}
	}
	if free < wave.count {
		if sim.Log != nil {
			sim.Log("Encounter design: wave %s has %d of %d adds free; spawn skipped.", wave.name(), free, wave.count)
		}
		return
	}

	lifetime := wave.lifetime(sim)
	diesAt := sim.CurrentTime + lifetime
	spawned := 0
	for _, target := range wave.pool {
		if spawned == wave.count {
			break
		}
		if target.IsEnabled() {
			continue
		}
		spawned++

		state := &wave.design.targets[target.Index]
		state.diesAt = diesAt
		if state.deathQueued {
			state.deathAction.Cancel(sim)
			state.deathQueued = false
		}
		if diesAt < sim.Duration {
			state.deathAction.NextActionAt = diesAt
			state.deathQueued = true
			sim.AddPendingAction(&state.deathAction)
		}

		target.SetHealthCurve(sim, sim.CurrentTime, lifetime, wave.knots[:]...)
		target.Enable(sim)
		wave.design.reapplyAmps(sim, target)
	}
	if sim.Log != nil {
		sim.Log("Encounter design: wave %s spawned %d adds.", wave.name(), wave.count)
	}
}

// Kills the add at the end of its life, with the wave's kill share, and leaves it idle at full health.
func (wave *designWave) despawn(sim *Simulation, target *Target) {
	state := &wave.design.targets[target.Index]
	state.deathQueued = false
	state.diesAt = NeverExpires
	if sim.Log != nil {
		target.Log(sim, "Encounter design: despawns from wave %s.", wave.name())
	}
	target.Die(sim, wave.config.KillShare)
	target.SetHealthCurve(sim, 0, 0, HealthKnot{At: 0, Percent: 100})
}

// The current time while the target is active; else the time a downtime or its wave brings it in next, or
// NeverExpires.
func (target *Target) NextActiveAt(sim *Simulation) time.Duration {
	if target.IsEnabled() {
		return sim.CurrentTime
	}
	design := sim.Encounter.design
	if design == nil {
		return NeverExpires
	}
	state := &design.targets[target.Index]
	if state.wave != nil {
		return state.wave.nextSpawnAt()
	}
	return state.nextActiveAt
}

// When the target's current life ends: an add's despawn, the fight's end for the encounter's own targets and
// for an add that outlives it. A target that is not active dies after the life its next activation starts,
// or at NeverExpires with none.
func (target *Target) DiesAt(sim *Simulation) time.Duration {
	if design := sim.Encounter.design; design != nil {
		if state := &design.targets[target.Index]; state.wave != nil {
			if target.IsEnabled() {
				return min(state.diesAt, sim.Duration)
			}
			if next := state.wave.nextSpawnAt(); next != NeverExpires {
				return min(next+state.wave.lifetime(sim), sim.Duration)
			}
			return NeverExpires
		}
	}
	if target.NextActiveAt(sim) != NeverExpires {
		return sim.Duration
	}
	return NeverExpires
}

// The active targets other than the boss, target 0.
func (encounter *Encounter) NumActiveAdds() int {
	if encounter.AllTargets[0].IsEnabled() {
		return len(encounter.ActiveTargets) - 1
	}
	return len(encounter.ActiveTargets)
}

// The start of the design's next move not yet started, or NeverExpires.
func (encounter *Encounter) NextMoveAt(_ *Simulation) time.Duration {
	if encounter.design == nil {
		return NeverExpires
	}
	return encounter.design.nextMoveAt
}

// How long the design's next move not yet started lasts, or 0.
func (encounter *Encounter) NextMoveDuration(_ *Simulation) time.Duration {
	if encounter.design == nil {
		return 0
	}
	return encounter.design.nextMoveDuration
}

// Whether a damage amp window of the type is running.
func (encounter *Encounter) AmpActive(_ *Simulation, ampType proto.DamageAmpType) bool {
	return encounter.design != nil && encounter.design.amps[ampType].active
}

// The current time while a damage amp window of the type runs; else the start of the next one, or NeverExpires.
func (encounter *Encounter) NextAmpAt(sim *Simulation, ampType proto.DamageAmpType) time.Duration {
	if encounter.design == nil {
		return NeverExpires
	}
	amp := &encounter.design.amps[ampType]
	if amp.active {
		return sim.CurrentTime
	}
	return amp.nextStartAt
}

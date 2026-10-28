package core

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// The lanes' state, built once with the environment.
type designLanes struct {
	moves       []*designMove
	amps        []*designAmp
	ampAuras    []*designAmpAura
	targetAuras []*designAmpAura // Indexed by Target.Index, nil for a target no target debuff names.
}

// An amp window starting or ending.
type designAmpEvent struct {
	at    time.Duration
	start bool
}

// An amp aura with the open windows' percent it carries, in hundredths of a percent. The stacks are set from
// the sum, never moved by a delta, so an aura a downtime or a death took off comes back with the sum.
type designAmpAura struct {
	aura *Aura
	sum  int32
}

// Sets the aura to the sum: active with the sum as its stacks, or off at 0.
func (amp *designAmpAura) apply(sim *Simulation) {
	switch {
	case amp.sum > 0:
		if !amp.aura.IsActive() {
			amp.aura.Activate(sim)
		}
		amp.aura.SetStacks(sim, amp.sum)
	case amp.aura.IsActive():
		amp.aura.SetStacks(sim, 0)
	}
}

// A damage amp of the design. A window adds percent to the sum of every aura of the amp's type and its end
// takes it off again, so overlapping windows add up. The stacks count hundredths of a percent, which keeps a
// fractional percent exact; the multiplier follows the stacks. The auras are shared by every amp of the type.
type designAmp struct {
	config     *proto.DesignDamageAmp
	skipped    bool
	auras      []*designAmpAura
	hundredths int32 // The window's percent in hundredths of a percent.
	fractions  []float64
	starts     []time.Duration // This iteration's window starts in time order.
	events     []designAmpEvent
	nextEvent  int
	nextStart  int // The window start still to come.
	open       int // Windows running.
	action     PendingAction
}

var designAmpLabels = [numDamageAmpTypes]string{
	proto.DamageAmpType_DamageAmpTypeCasterBuff:      "Damage Done % (Caster Buff)",
	proto.DamageAmpType_DamageAmpTypeEnvironmentBuff: "Damage Done % (Environment Buff)",
	proto.DamageAmpType_DamageAmpTypeTargetDebuff:    "Damage Taken % (Target Debuff)",
}

// Builds the design's lanes once the raid exists: the moves' buffers and each amp's auras and buffers.
func (design *encounterDesign) registerLanes(env *Environment) {
	if design == nil {
		return
	}
	design.registerMoves(env)

	design.lanes.targetAuras = make([]*designAmpAura, len(env.Encounter.AllTargets))
	var casterAuras [numDamageAmpTypes][]*designAmpAura
	for index, config := range design.config.DamageAmps {
		ampType := config.Type
		amp := &designAmp{config: config, hundredths: int32(math.Round(config.Percent * 100))}
		design.lanes.amps = append(design.lanes.amps, amp)
		switch {
		case int(ampType) < 0 || int(ampType) >= numDamageAmpTypes:
			amp.skipped = true
			design.constructNotes = append(design.constructNotes, fmt.Sprintf("damage amp %d has an unknown type; skipped.", index+1))
			continue
		case amp.hundredths <= 0:
			amp.skipped = true
			design.constructNotes = append(design.constructNotes, fmt.Sprintf("damage amp %d has no percent above 0; skipped.", index+1))
			continue
		}

		if ampType == proto.DamageAmpType_DamageAmpTypeTargetDebuff {
			for _, target := range design.ampTargets(env, config.Targets) {
				if design.lanes.targetAuras[target.Index] == nil {
					design.lanes.targetAuras[target.Index] = design.newAmpAura(registerDesignAmpTargetAura(target))
				}
				amp.auras = append(amp.auras, design.lanes.targetAuras[target.Index])
			}
			if len(amp.auras) == 0 {
				amp.skipped = true
				design.constructNotes = append(design.constructNotes, fmt.Sprintf("damage amp %d names no target the encounter has; skipped.", index+1))
				continue
			}
		} else {
			if casterAuras[ampType] == nil {
				for _, playerUnit := range env.Raid.AllPlayerUnits {
					if agent := env.Raid.GetPlayerFromUnit(playerUnit); agent != nil {
						casterAuras[ampType] = append(casterAuras[ampType], design.newAmpAura(registerDesignAmpCasterAura(agent.GetCharacter(), ampType)))
					}
				}
			}
			amp.auras = casterAuras[ampType]
		}

		occurrences := maxOccurrences(config.Start, config.Repeat)
		amp.fractions = make([]float64, 0, occurrences)
		amp.starts = make([]time.Duration, 0, occurrences)
		amp.events = make([]designAmpEvent, 0, 2*occurrences)
		amp.action = PendingAction{
			Priority: ActionPriorityDOT,
			OnAction: func(sim *Simulation) { design.runAmp(sim, amp) },
		}
	}
}

func (design *encounterDesign) newAmpAura(aura *Aura) *designAmpAura {
	ampAura := &designAmpAura{aura: aura}
	design.lanes.ampAuras = append(design.lanes.ampAuras, ampAura)
	return ampAura
}

// The targets a target debuff reaches: the encounter's own targets it names and every pool target of the
// waves it names, or the boss with neither.
func (design *encounterDesign) ampTargets(env *Environment, targets *proto.DesignTargets) []*Target {
	var resolved []*Target
	for _, index := range targets.GetTargetIndices() {
		if index >= 0 && int(index) < design.numEncounterTargets {
			resolved = append(resolved, env.Encounter.AllTargets[index])
		}
	}
	for _, index := range targets.GetWaveIndices() {
		if index >= 0 && int(index) < len(design.waves) {
			resolved = append(resolved, design.waves[index].pool...)
		}
	}
	if len(targets.GetTargetIndices()) == 0 && len(targets.GetWaveIndices()) == 0 {
		resolved = append(resolved, env.Encounter.AllTargets[0])
	}
	slices.SortFunc(resolved, func(a, b *Target) int { return int(a.Index - b.Index) })
	return slices.Compact(resolved)
}

func registerDesignAmpCasterAura(character *Character, ampType proto.DamageAmpType) *Aura {
	return character.GetOrRegisterAura(Aura{
		Label:     designAmpLabels[ampType],
		ActionID:  ActionID{OtherID: proto.OtherAction_OtherActionDamageAmplifier}.WithTag(int32(ampType)),
		Duration:  NeverExpires,
		MaxStacks: math.MaxInt32,
		OnStacksChange: func(_ *Aura, _ *Simulation, oldStacks int32, newStacks int32) {
			amount := (1 + float64(newStacks)/10000) / (1 + float64(oldStacks)/10000)
			character.PseudoStats.DamageDealtMultiplier *= amount
			for _, pet := range character.Pets {
				if !pet.isGuardian {
					pet.PseudoStats.DamageDealtMultiplier *= amount
				}
			}
		},
	})
}

func registerDesignAmpTargetAura(target *Target) *Aura {
	ampType := proto.DamageAmpType_DamageAmpTypeTargetDebuff
	return target.GetOrRegisterAura(Aura{
		Label:     designAmpLabels[ampType],
		ActionID:  ActionID{OtherID: proto.OtherAction_OtherActionDamageAmplifier}.WithTag(int32(ampType)),
		Duration:  NeverExpires,
		MaxStacks: math.MaxInt32,
		OnStacksChange: func(_ *Aura, _ *Simulation, oldStacks int32, newStacks int32) {
			target.PseudoStats.DamageTakenMultiplier *= (1 + float64(newStacks)/10000) / (1 + float64(oldStacks)/10000)
		},
	})
}

// Puts the target's open target debuff windows back on it, once a spawn or the end of a downtime enabled it.
func (design *encounterDesign) reapplyAmps(sim *Simulation, target *Target) {
	if ampAura := design.lanes.targetAuras[target.Index]; ampAura != nil {
		ampAura.apply(sim)
	}
}

// Schedules the iteration's damage amp windows: a caster buff, an environment buff, or a target debuff on the
// design's chosen targets and the adds that spawn inside the window, each raising damage by the window's
// percent from its start to its end, with its repeats rolled through rollOccurrences. Keeps amps[type] on
// whether a window of the type runs and when the next one starts. The runner calls it only for a design with
// damage amps.
//
// A window's length is the end's time less the start's, moved by the end's jitter per window. A target debuff
// sits on every pool target of its waves for the whole window, which the adds cannot be hit outside of.
func (design *encounterDesign) scheduleAmps(sim *Simulation) {
	for _, ampAura := range design.lanes.ampAuras {
		ampAura.sum = 0
	}
	for _, amp := range design.lanes.amps {
		amp.starts = amp.starts[:0]
		amp.events = amp.events[:0]
		amp.nextEvent = 0
		amp.nextStart = 0
		amp.open = 0
		if amp.skipped {
			continue
		}
		length := amp.config.End.GetAt() - amp.config.Start.GetAt()
		amp.fractions = design.rollOccurrences(sim, amp.config.Start, amp.config.Repeat, amp.fractions)
		for _, fraction := range amp.fractions {
			start := designTime(sim, fraction)
			end := designTime(sim, fraction+design.roll(sim, length, amp.config.End.GetJitter()))
			if start >= sim.Duration || end <= start {
				continue
			}
			amp.starts = append(amp.starts, start)
			amp.events = append(amp.events, designAmpEvent{at: start, start: true})
			if end < sim.Duration {
				amp.events = append(amp.events, designAmpEvent{at: end})
			}
		}
		// A window that starts as another ends is up before the other leaves, so the aura never fades between them.
		slices.SortStableFunc(amp.events, func(a, b designAmpEvent) int {
			switch {
			case a.at < b.at:
				return -1
			case a.at > b.at:
				return 1
			case a.start == b.start:
				return 0
			case a.start:
				return -1
			}
			return 1
		})
		if len(amp.events) > 0 {
			amp.action.NextActionAt = amp.events[0].at
			sim.AddPendingAction(&amp.action)
		}
	}
	design.updateAmpLookahead()
}

// Applies the amp's next window start or end, and queues the event after it.
func (design *encounterDesign) runAmp(sim *Simulation, amp *designAmp) {
	event := amp.events[amp.nextEvent]
	amp.nextEvent++
	if amp.nextEvent < len(amp.events) {
		amp.action.NextActionAt = amp.events[amp.nextEvent].at
		sim.AddPendingAction(&amp.action)
	}

	delta := -amp.hundredths
	if event.start {
		delta = amp.hundredths
		amp.open++
		amp.nextStart++
	} else {
		amp.open--
	}
	for _, ampAura := range amp.auras {
		ampAura.sum += delta
		ampAura.apply(sim)
	}
	if sim.Log != nil {
		sim.Log("Encounter design: damage amp %s of %v%% %s.", designAmpLabels[amp.config.Type], amp.config.Percent, Ternary(event.start, "starts", "ends"))
	}
	design.updateAmpLookahead()
}

// Points amps[type] at whether a window of the type runs and the earliest window start of the type to come.
func (design *encounterDesign) updateAmpLookahead() {
	for i := range design.amps {
		design.amps[i] = designAmpLookahead{nextStartAt: NeverExpires}
	}
	for _, amp := range design.lanes.amps {
		if amp.skipped {
			continue
		}
		lookahead := &design.amps[amp.config.Type]
		lookahead.active = lookahead.active || amp.open > 0
		if amp.nextStart < len(amp.starts) {
			lookahead.nextStartAt = min(lookahead.nextStartAt, amp.starts[amp.nextStart])
		}
	}
}

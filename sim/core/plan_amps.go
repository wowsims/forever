package core

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// An amp window starting or ending.
type planAmpEvent struct {
	at    time.Duration
	start bool
}

// An amp aura with the open windows' percent it carries, in hundredths of a percent. A target's survives the
// target's downtimes and deaths, as expireAll skips a NeverExpires aura.
type planAmpAura struct {
	aura *Aura
	sum  int32
}

// Sets the aura to the sum: active with the sum as its stacks, or off at 0.
func (amp *planAmpAura) apply(sim *Simulation) {
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

// A damage amp of the plan. A window adds percent to the sum of every aura of the amp's type and its end
// takes it off again, so overlapping windows add up. The stacks count hundredths of a percent, which keeps a
// fractional percent exact; the multiplier follows the stacks. The auras are shared by every amp of the type.
type planAmp struct {
	config     *proto.PlanDamageAmp
	skipped    bool
	auras      []*planAmpAura
	hundredths int32 // The window's percent in hundredths of a percent.
	fractions  []float64
	starts     []time.Duration // This iteration's window starts in time order.
	events     []planAmpEvent
	nextEvent  int
	nextStart  int // The window start still to come.
	open       int // Windows running.
	action     PendingAction
}

var planAmpLabels = [numDamageAmpTypes]string{
	proto.DamageAmpType_DamageAmpTypeCasterBuff:      "Damage Done % (Caster Buff)",
	proto.DamageAmpType_DamageAmpTypeEnvironmentBuff: "Damage Done % (Environment Buff)",
	proto.DamageAmpType_DamageAmpTypeTargetDebuff:    "Damage Taken % (Target Debuff)",
}

// Builds each amp's auras and buffers once, with the environment.
func (plan *encounterPlan) registerAmps(env *Environment) {
	targetAuras := make([]*planAmpAura, len(env.Encounter.AllTargets)) // Indexed by Target.Index, nil for a target no target debuff names.
	var casterAuras [numDamageAmpTypes][]*planAmpAura
	for index, config := range plan.config.DamageAmps {
		ampType := config.Type
		amp := &planAmp{config: config, hundredths: int32(math.Round(config.Percent * 100))}
		plan.damageAmps = append(plan.damageAmps, amp)
		switch {
		case int(ampType) < 0 || int(ampType) >= numDamageAmpTypes:
			amp.skipped = true
			plan.constructNotes = append(plan.constructNotes, fmt.Sprintf("damage amp %d has an unknown type; skipped.", index+1))
			continue
		case amp.hundredths <= 0:
			amp.skipped = true
			plan.constructNotes = append(plan.constructNotes, fmt.Sprintf("damage amp %d has no percent above 0; skipped.", index+1))
			continue
		}

		if ampType == proto.DamageAmpType_DamageAmpTypeTargetDebuff {
			for _, target := range plan.ampTargets(env, config.Targets) {
				if targetAuras[target.Index] == nil {
					targetAuras[target.Index] = plan.newAmpAura(registerPlanAmpTargetAura(target))
				}
				amp.auras = append(amp.auras, targetAuras[target.Index])
			}
			if len(amp.auras) == 0 {
				amp.skipped = true
				plan.constructNotes = append(plan.constructNotes, fmt.Sprintf("damage amp %d names no target the encounter has; skipped.", index+1))
				continue
			}
		} else {
			if casterAuras[ampType] == nil {
				for _, playerUnit := range env.Raid.AllPlayerUnits {
					if agent := env.Raid.GetPlayerFromUnit(playerUnit); agent != nil {
						casterAuras[ampType] = append(casterAuras[ampType], plan.newAmpAura(registerPlanAmpCasterAura(agent.GetCharacter(), ampType)))
					}
				}
			}
			amp.auras = casterAuras[ampType]
		}

		occurrences := maxOccurrences(config.Start, config.Repeat)
		amp.fractions = make([]float64, 0, occurrences)
		amp.starts = make([]time.Duration, 0, occurrences)
		amp.events = make([]planAmpEvent, 0, 2*occurrences)
		amp.action = PendingAction{
			Priority: ActionPriorityDOT,
			OnAction: func(sim *Simulation) { plan.runAmp(sim, amp) },
		}
	}
}

func (plan *encounterPlan) newAmpAura(aura *Aura) *planAmpAura {
	ampAura := &planAmpAura{aura: aura}
	plan.ampAuras = append(plan.ampAuras, ampAura)
	return ampAura
}

// The targets a target debuff reaches: the encounter's own targets it names and every pool target of the
// waves it names, or the boss with neither.
func (plan *encounterPlan) ampTargets(env *Environment, targets *proto.PlanTargets) []*Target {
	var resolved []*Target
	for _, index := range targets.GetTargetIndices() {
		if index >= 0 && int(index) < plan.numEncounterTargets {
			resolved = append(resolved, env.Encounter.AllTargets[index])
		}
	}
	for _, index := range targets.GetWaveIndices() {
		if index >= 0 && int(index) < len(plan.waves) {
			resolved = append(resolved, plan.waves[index].pool...)
		}
	}
	if len(targets.GetTargetIndices()) == 0 && len(targets.GetWaveIndices()) == 0 {
		resolved = append(resolved, env.Encounter.AllTargets[0])
	}
	slices.SortFunc(resolved, func(a, b *Target) int { return int(a.Index - b.Index) })
	return slices.Compact(resolved)
}

// The aura of an amp type. A change of its stacks hands multiply the ratio of the new sum to the old.
func planAmpAuraConfig(ampType proto.DamageAmpType, multiply func(ratio float64)) Aura {
	return Aura{
		Label:     planAmpLabels[ampType],
		ActionID:  ActionID{OtherID: proto.OtherAction_OtherActionDamageAmplifier}.WithTag(int32(ampType)),
		Duration:  NeverExpires,
		MaxStacks: math.MaxInt32,
		OnStacksChange: func(_ *Aura, _ *Simulation, oldStacks int32, newStacks int32) {
			multiply((1 + float64(newStacks)/10000) / (1 + float64(oldStacks)/10000))
		},
	}
}

func registerPlanAmpCasterAura(character *Character, ampType proto.DamageAmpType) *Aura {
	return character.GetOrRegisterAura(planAmpAuraConfig(ampType, func(ratio float64) {
		character.PseudoStats.DamageDealtMultiplier *= ratio
		for _, pet := range character.Pets {
			if !pet.isGuardian {
				pet.PseudoStats.DamageDealtMultiplier *= ratio
			}
		}
	}))
}

func registerPlanAmpTargetAura(target *Target) *Aura {
	return target.GetOrRegisterAura(planAmpAuraConfig(proto.DamageAmpType_DamageAmpTypeTargetDebuff, func(ratio float64) {
		target.PseudoStats.DamageTakenMultiplier *= ratio
	}))
}

// Schedules the iteration's damage amp windows: a caster buff, an environment buff, or a target debuff on the
// plan's chosen targets and the adds that spawn inside the window, each raising damage by the window's
// percent from its start to its end, with its repeats rolled through rollOccurrences. Keeps amps[type] on
// whether a window of the type runs and when the next one starts.
//
// A window's length is the end's time less the start's, moved by the end's jitter per window. A target debuff
// sits on every pool target of its waves for the whole window, which the adds cannot be hit outside of.
func (plan *encounterPlan) scheduleAmps(sim *Simulation) {
	for _, ampAura := range plan.ampAuras {
		ampAura.sum = 0
	}
	for _, amp := range plan.damageAmps {
		amp.starts = amp.starts[:0]
		amp.events = amp.events[:0]
		amp.nextEvent = 0
		amp.nextStart = 0
		amp.open = 0
		if amp.skipped {
			continue
		}
		length := amp.config.End.GetAt() - amp.config.Start.GetAt()
		amp.fractions = plan.rollOccurrences(sim, amp.config.Start, amp.config.Repeat, amp.fractions)
		for _, fraction := range amp.fractions {
			start := planTime(sim, fraction)
			end := planTime(sim, fraction+plan.roll(sim, length, amp.config.End.GetJitter()))
			if start >= sim.Duration || end <= start {
				continue
			}
			amp.starts = append(amp.starts, start)
			amp.events = append(amp.events, planAmpEvent{at: start, start: true})
			if end < sim.Duration {
				amp.events = append(amp.events, planAmpEvent{at: end})
			}
		}
		// A window that starts as another ends is up before the other leaves, so the aura never fades between them.
		slices.SortStableFunc(amp.events, func(a, b planAmpEvent) int {
			return cmp.Or(cmp.Compare(a.at, b.at), cmp.Compare(Ternary(a.start, 0, 1), Ternary(b.start, 0, 1)))
		})
		if len(amp.events) > 0 {
			amp.action.NextActionAt = amp.events[0].at
			sim.AddPendingAction(&amp.action)
		}
	}
	plan.updateAmpLookahead()
}

// Applies the amp's next window start or end, and queues the event after it.
func (plan *encounterPlan) runAmp(sim *Simulation, amp *planAmp) {
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
		sim.Log("Encounter plan: damage amp %s of %v%% %s.", planAmpLabels[amp.config.Type], amp.config.Percent, Ternary(event.start, "starts", "ends"))
	}
	plan.updateAmpLookahead()
}

// Points amps[type] at whether a window of the type runs and the earliest window start of the type to come.
func (plan *encounterPlan) updateAmpLookahead() {
	for i := range plan.amps {
		plan.amps[i] = planAmpLookahead{nextStartAt: NeverExpires}
	}
	for _, amp := range plan.damageAmps {
		if amp.skipped {
			continue
		}
		lookahead := &plan.amps[amp.config.Type]
		lookahead.active = lookahead.active || amp.open > 0
		if amp.nextStart < len(amp.starts) {
			lookahead.nextStartAt = min(lookahead.nextStartAt, amp.starts[amp.nextStart])
		}
	}
}

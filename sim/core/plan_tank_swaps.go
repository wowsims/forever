package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// A tank swap of the plan. While one of its windows runs, the target's single-target attacks land on its other
// tank: the melee swing, a Strike it takes and a Thrash's extra attacks off it, and a Spell Hit cast then. Area Hits
// and Pulses hit everyone as before. The swing keeps its timer through a window, so a swing due inside it goes to
// the other tank, a Strike it takes is spent there, and the first one after it lands on the running timer. The end
// gives every repeat's window its length, as a Thrash's does, and the swap holds through the target's downtimes.
// Both edges of a window run before every other plan event at their time, so a Spell Hit on either edge finds the
// window running from its start up to its end; a swing due on an edge runs before it.
const planTankSwapPriority = planStartPriority + 1

type planTankSwap struct {
	index  int
	config *proto.PlanTankSwap
	target *Target // nil when the index is not one of the encounter's own targets.
	skip   bool    // Set when construction or the tanks dropped the swap, with a construct note saying why.

	occurrences planOccurrences // This iteration's window starts.
	ends        []time.Duration
	nextEnd     int
	endAction   PendingAction
}

func (plan *encounterPlan) newTankSwaps(encounter *Encounter) {
	for index, config := range plan.config.TankSwaps {
		swap := &planTankSwap{index: index, config: config}
		plan.tankSwaps = append(plan.tankSwaps, swap)
		if targetIndex := int(config.TargetIndex); targetIndex >= 0 && targetIndex < plan.numEncounterTargets {
			swap.target = encounter.AllTargets[targetIndex]
		}
		if skip := swap.skipReason(); skip != "" {
			swap.skip = true
			plan.constructNotes = append(plan.constructNotes, fmt.Sprintf("tank swap %s %s; skipped.", swap.name(), skip))
			continue
		}

		occurrences := maxOccurrences(config.GetStart(), config.Repeat)
		swap.occurrences = newPlanOccurrences(occurrences, planTankSwapPriority, swap.start)
		swap.ends = make([]time.Duration, 0, occurrences)
		swap.endAction = PendingAction{Priority: planTankSwapPriority, OnAction: swap.end}
	}
}

// Why the swap does nothing, or "" when it acts.
func (swap *planTankSwap) skipReason() string {
	if swap.target == nil {
		return fmt.Sprintf("names target %d, which the encounter does not have", swap.config.TargetIndex)
	}
	if swap.config.GetEnd().GetAt() <= swap.config.GetStart().GetAt() {
		return "has a window that ends by its start"
	}
	return ""
}

func (swap *planTankSwap) name() string {
	if swap.config.Name != "" {
		return swap.config.Name
	}
	return fmt.Sprintf("%d", swap.index+1)
}

// Sets the two tanks of each swapped target, once the targets have their tanks and the players their tank slots: the
// unit the target attacks, and the first player in a tank slot that is not that unit, nil for a tank the sim does
// not model. A target no tank slot tanks, or with neither tank in the sim, has nothing to swap.
func (plan *encounterPlan) setupTankSwaps(env *Environment, numTankSlots int) {
	if plan == nil {
		return
	}
	for _, swap := range plan.tankSwaps {
		if swap.skip {
			continue
		}
		if tankIndex := env.Encounter.targetConfigs[swap.target.Index].GetTankIndex(); tankIndex < 0 || int(tankIndex) >= numTankSlots {
			swap.skip = true
			plan.constructNotes = append(plan.constructNotes, fmt.Sprintf(
				"tank swap %s has target %d, which no tank slot tanks; skipped.", swap.name(), swap.config.TargetIndex))
			continue
		}
		state := &plan.targets[swap.target.Index]
		if !state.swapsTanks {
			state.swapTanks[0] = swap.target.CurrentTarget
			for _, unit := range env.Raid.AllPlayerUnits {
				if agent := env.Raid.GetPlayerFromUnit(unit); agent != nil && agent.GetCharacter().isTank && unit != state.swapTanks[0] {
					state.swapTanks[1] = unit
					break
				}
			}
			state.swapsTanks = state.swapTanks[0] != nil || state.swapTanks[1] != nil
		}
		if !state.swapsTanks {
			swap.skip = true
			plan.constructNotes = append(plan.constructNotes, fmt.Sprintf(
				"tank swap %s has target %d swap between two tanks outside the sim; skipped.", swap.name(), swap.config.TargetIndex))
		}
	}
}

// Whether a tank swap brings a target onto the unit, which then tanks it.
func (plan *encounterPlan) swapsOnto(unit *Unit) bool {
	if plan == nil {
		return false
	}
	for _, swap := range plan.tankSwaps {
		if !swap.skip && plan.targets[swap.target.Index].swapTanks[1] == unit {
			return true
		}
	}
	return false
}

// The unit the target's single-target attacks land on, and whether a tank swap decides it: its tank, or its other
// tank while a window runs, nil for a tank the sim does not model, on which the attack is spent. Without a swap the
// caller keeps its own rule.
func (plan *encounterPlan) tankSwapVictim(target *Unit) (*Unit, bool) {
	if plan == nil || target.Type != EnemyUnit {
		return nil, false
	}
	state := &plan.targets[target.Index]
	if !state.swapsTanks {
		return nil, false
	}
	if state.swapWindows > 0 {
		return state.swapTanks[1], true
	}
	return state.swapTanks[0], true
}

// Whether the target swings with no target of its own: its tank is outside the sim and a swap brings it onto a tank
// inside it, so its swing runs from the pull and a window's first swing lands on the running timer.
func (plan *encounterPlan) swingsUntanked(target *Unit) bool {
	if plan == nil || target.Type != EnemyUnit {
		return false
	}
	state := &plan.targets[target.Index]
	return state.swapTanks[0] == nil && state.swapTanks[1] != nil
}

// Schedules the iteration's tank swap windows, with their repeats rolled through rollOccurrences and one roll of
// the end giving every window its length.
func (plan *encounterPlan) scheduleTankSwaps(sim *Simulation) {
	for _, swap := range plan.tankSwaps {
		swap.nextEnd = 0
		swap.ends = swap.ends[:0]
		if swap.skip {
			continue
		}

		config := swap.config
		occurrences := &swap.occurrences
		occurrences.roll(sim, plan, config.GetStart(), config.Repeat)
		end := plan.roll(sim, config.GetEnd().GetAt(), config.GetEnd().GetJitter())
		length := end - occurrences.fractions[0]
		if length <= 0 {
			if sim.Log != nil {
				sim.Log("Encounter plan: tank swap %s rolled a window that ends by its start; skipped this iteration.", swap.name())
			}
			continue
		}
		for _, fraction := range occurrences.fractions {
			if at := planTime(sim, fraction); at < sim.Duration {
				occurrences.times = append(occurrences.times, at)
				swap.ends = append(swap.ends, planTime(sim, fraction+length))
			}
		}
		occurrences.queueFirst(sim)
	}
}

// Opens the swap's next window, whether its target is active or not.
func (swap *planTankSwap) start(sim *Simulation) {
	occurrence := swap.occurrences.next
	swap.occurrences.advance(sim)
	if end := swap.ends[occurrence]; end < sim.Duration && swap.nextEnd == occurrence {
		swap.endAction.NextActionAt = end
		sim.AddPendingAction(&swap.endAction)
	}
	sim.Encounter.plan.targets[swap.target.Index].swapWindows++
	if sim.Log != nil {
		swap.target.Log(sim, "Encounter plan: tank swap %s starts.", swap.name())
	}
}

// Ends the swap's next window; the target's attacks go back to its tank once no other window on it runs.
func (swap *planTankSwap) end(sim *Simulation) {
	swap.nextEnd++
	if swap.nextEnd < swap.occurrences.next && swap.ends[swap.nextEnd] < sim.Duration {
		swap.endAction.NextActionAt = swap.ends[swap.nextEnd]
		sim.AddPendingAction(&swap.endAction)
	}
	sim.Encounter.plan.targets[swap.target.Index].swapWindows--
	if sim.Log != nil {
		swap.target.Log(sim, "Encounter plan: tank swap %s ends.", swap.name())
	}
}

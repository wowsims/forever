package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

type APLActionSequence struct {
	defaultAPLActionImpl
	unit       *Unit
	name       string
	subactions []*APLAction
	curIdx     int
}

func (rot *APLRotation) newActionSequence(config *proto.APLActionSequence) APLActionImpl {
	subactions := MapSlice(config.Actions, func(action *proto.APLAction) *APLAction {
		return rot.newAPLAction(action)
	})
	subactions = FilterSlice(subactions, func(action *APLAction) bool { return action != nil })
	if len(subactions) == 0 {
		return nil
	}

	return &APLActionSequence{
		unit:       rot.unit,
		name:       config.Name,
		subactions: subactions,
	}
}
func (action *APLActionSequence) GetInnerActions() []*APLAction {
	return Flatten(MapSlice(action.subactions, func(action *APLAction) []*APLAction { return action.GetAllActions() }))
}
func (action *APLActionSequence) Finalize(rot *APLRotation) {
	for _, subaction := range action.subactions {
		subaction.impl.Finalize(rot)
	}
}
func (action *APLActionSequence) PostFinalize(rot *APLRotation) {
	for _, subaction := range action.subactions {
		subaction.impl.PostFinalize(rot)
	}
}
func (action *APLActionSequence) Reset(*Simulation) {
	action.curIdx = 0
}
func (action *APLActionSequence) IsReady(sim *Simulation) bool {
	action.unit.Rotation.inSequence = true
	isReady := (action.curIdx < len(action.subactions)) && action.subactions[action.curIdx].IsReady(sim)
	action.unit.Rotation.inSequence = false
	return isReady
}
func (action *APLActionSequence) Execute(sim *Simulation) {
	action.unit.Rotation.inSequence = true
	action.subactions[action.curIdx].Execute(sim)

	if action.unit.CanQueueSpell(sim) {
		// Only advance to the next step in the sequence if we actually cast a spell rather than simply queueing one up.
		action.curIdx++
	} else {
		// If we did queue up a spell, then modify the queue action to advance the sequence when it fires.
		queueAction := action.unit.QueuedSpell.queueAction
		oldFunc := queueAction.OnAction
		queueAction.OnAction = func(sim *Simulation) {
			oldFunc(sim)
			action.curIdx++
			queueAction.OnAction = oldFunc
		}
		// Behind a cast bar the rotation still looks at its end, where major cooldowns can go.
		if !action.unit.QueuedSpell.behindCast {
			action.unit.SetRotationTimer(sim, queueAction.NextActionAt+time.Duration(1))
		}
	}

	action.unit.Rotation.inSequence = false
}
func (action *APLActionSequence) String() string {
	return "Sequence(" + strings.Join(MapSlice(action.subactions, func(subaction *APLAction) string { return fmt.Sprintf("(%s)", subaction) }), "+") + ")"
}

type APLActionResetSequence struct {
	defaultAPLActionImpl
	name     string
	sequence *APLActionSequence
}

func (rot *APLRotation) newActionResetSequence(config *proto.APLActionResetSequence) APLActionImpl {
	if config.SequenceName == "" {
		rot.ValidationMessage(proto.LogLevel_Warning, "Reset Sequence must provide a sequence name")
		return nil
	}
	return &APLActionResetSequence{
		name: config.SequenceName,
	}
}
func (action *APLActionResetSequence) Finalize(rot *APLRotation) {
	for _, otherAction := range rot.allAPLActions() {
		if sequence, ok := otherAction.impl.(*APLActionSequence); ok && sequence.name == action.name {
			action.sequence = sequence
			return
		}
	}
	rot.ValidationMessage(proto.LogLevel_Warning, "No sequence with name: '%s'", action.name)
}
func (action *APLActionResetSequence) IsReady(sim *Simulation) bool {
	return action.sequence != nil && action.sequence.curIdx != 0
}
func (action *APLActionResetSequence) Execute(sim *Simulation) {
	action.sequence.curIdx = 0
}
func (action *APLActionResetSequence) String() string {
	return fmt.Sprintf("Reset Sequence(name = '%s')", action.name)
}

type APLActionStrictSequence struct {
	defaultAPLActionImpl
	unit       *Unit
	subactions []*APLAction
	curIdx     int

	subactionSpells []*Spell

	// The queue action currently hooked to advance this sequence, and the callback it replaced.
	hookedQueueAction *PendingAction
	hookedOnAction    func(*Simulation)

	gate              APLValue // A Strict Multidot's conditions; the sequence ends when it is false as its next cast can go out.
	yieldsToCooldowns bool     // A Strict Multidot's: while its step waits behind a cast bar, major cooldowns go at the bar's end.
}

func (rot *APLRotation) newActionStrictSequence(config *proto.APLActionStrictSequence) APLActionImpl {
	subactions := MapSlice(config.Actions, func(action *proto.APLAction) *APLAction {
		return rot.newAPLAction(action)
	})
	subactions = FilterSlice(subactions, func(action *APLAction) bool { return action != nil })
	if len(subactions) == 0 {
		return nil
	}

	return &APLActionStrictSequence{
		unit:       rot.unit,
		subactions: subactions,
	}
}
func (action *APLActionStrictSequence) GetInnerActions() []*APLAction {
	return Flatten(MapSlice(action.subactions, func(action *APLAction) []*APLAction { return action.GetAllActions() }))
}
func (action *APLActionStrictSequence) Finalize(rot *APLRotation) {
	for _, subaction := range action.subactions {
		subaction.impl.Finalize(rot)
		action.subactionSpells = append(action.subactionSpells, subaction.GetAllSpells()...)
	}
}
func (action *APLActionStrictSequence) PostFinalize(rot *APLRotation) {
	for _, subaction := range action.subactions {
		subaction.impl.PostFinalize(rot)
	}
}
func (action *APLActionStrictSequence) Reset(*Simulation) {
	action.curIdx = 0
	action.unhookQueueAction()
	action.unit.Rotation.inSequence = false
}
func (action *APLActionStrictSequence) IsReady(sim *Simulation) bool {
	action.unit.Rotation.inSequence = true

	if (action.unit.GCD.TimeToReady(sim) > MaxSpellQueueWindow) && (len(action.subactionSpells) > 0) {
		action.unit.Rotation.inSequence = false
		return false
	}
	if !action.subactions[0].IsReady(sim) {
		action.unit.Rotation.inSequence = false
		return false
	}

	return true
}
func (action *APLActionStrictSequence) Execute(sim *Simulation) {
	action.unit.Rotation.inSequence = true
	action.unit.Rotation.pushControllingAction(action)
}
func (action *APLActionStrictSequence) relinquishControl() {
	action.curIdx = 0
	action.unhookQueueAction()
	action.unit.Rotation.inSequence = false
	action.unit.Rotation.popControllingAction(action)
}
func (action *APLActionStrictSequence) advanceSequence() {
	action.curIdx++
	if action.curIdx == len(action.subactions) {
		action.relinquishControl()
	}
}

// The subaction the sequence is at, without moving it on.
func (action *APLActionStrictSequence) current() *APLAction {
	return action.subactions[action.curIdx]
}
func (action *APLActionStrictSequence) GetNextAction(sim *Simulation) *APLAction {
	if nextAction := action.current(); nextAction.IsReady(sim) {
		castReady := action.unit.GCD.IsReady(sim)
		switch cast := nextAction.impl.(type) {
		case *APLActionCastSpell:
			castReady = cast.spell.CanCast(sim, cast.target.Get())
		case *APLActionCastFriendlySpell:
			castReady = cast.spell.CanCast(sim, cast.target.Get())
		case *APLActionChannelSpell:
			castReady = cast.spell.CanCast(sim, cast.target.Get())
		}
		if action.curIdx > 0 && action.gate != nil {
			// A queued cast fires without the gate read again, so a gated cast waits until it can go out.
			if !castReady {
				return nil
			}
			if !action.gate.GetBool(sim) {
				action.relinquishControl()
				return action.unit.Rotation.getNextAction(sim)
			}
		}
		if castReady {
			action.advanceSequence()
		} else if _, ok := nextAction.impl.(*APLActionWait); ok {
			action.advanceSequence()
		} else if _, ok := nextAction.impl.(*APLActionItemSwap); ok {
			action.advanceSequence()
		} else {
			// Do not advance the sequence; the action will be queued if it can be
		}

		return nextAction
	} else if !action.unit.CanQueueSpell(sim) && action.unit.QueuedSpell != nil {
		// The spell was already queued at this timestep; modify it to advance the sequence when it executes
		// and return nil to wait for the GCD to become ready.
		queueAction := action.unit.QueuedSpell.queueAction
		action.hookQueueAction(queueAction)
		if !action.yieldsToCooldowns || !action.unit.QueuedSpell.behindCast {
			action.unit.SetRotationTimer(sim, queueAction.NextActionAt+time.Duration(1))
		}

		return nil
	} else if action.yieldsToCooldowns && action.hookedQueueAction != nil && action.unit.castQueuedBehindBar() {
		// The step queued behind the bar goes off in the cast-end pass; ahead of it only a major cooldown,
		// the rotation's first ready action, may go.
		rot := action.unit.Rotation
		for _, priorityAction := range rot.priorityList {
			if priorityAction.IsReady(sim) {
				return Ternary(rot.isMajorCooldown(sim, priorityAction), priorityAction, nil)
			}
		}
		return nil
	} else if action.unit.GCD.TimeToReady(sim) <= MaxSpellQueueWindow {
		// If the GCD is ready when the next subaction isn't, it means the sequence is bad
		// so reset and exit the sequence.
		action.relinquishControl()
		return action.unit.Rotation.getNextAction(sim)
	} else {
		// Return nil to wait for the GCD to become ready.
		return nil
	}
}

// hookQueueAction makes the queued spell's action advance this sequence when it fires.
// Must stay idempotent: the rotation can be re-evaluated within a timestep, and a second
// wrapper would advance twice and strand a stale hook on the queue action, which is reused
// for later unrelated spells. The wrapper unhooks itself first, so any callback that manages
// OnAction gets the last word; the guard keeps it off another action's hook.
func (action *APLActionStrictSequence) hookQueueAction(queueAction *PendingAction) {
	if action.hookedQueueAction == queueAction {
		return
	}
	action.unhookQueueAction()

	oldFunc := queueAction.OnAction
	action.hookedQueueAction = queueAction
	action.hookedOnAction = oldFunc
	queueAction.OnAction = func(sim *Simulation) {
		if action.hookedQueueAction == queueAction {
			action.unhookQueueAction()
		}
		oldFunc(sim)
		action.advanceSequence()
	}
}
func (action *APLActionStrictSequence) unhookQueueAction() {
	if action.hookedQueueAction == nil {
		return
	}
	action.hookedQueueAction.OnAction = action.hookedOnAction
	action.hookedQueueAction = nil
	action.hookedOnAction = nil
}
func (action *APLActionStrictSequence) String() string {
	return "Strict Sequence(" + strings.Join(MapSlice(action.subactions, func(subaction *APLAction) string { return fmt.Sprintf("(%s)", subaction) }), "+") + ")"
}

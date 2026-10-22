package core

import (
	"time"
)

type QueuedSpell struct {
	spell       *Spell
	target      *Unit
	queueAction *PendingAction
	castFn      func(*Simulation)

	channelAction *APLActionChannelSpell // The channel action that queued the spell, which its cast carries when it fires.
	behindCast    bool                   // Queued under a cast bar that holds the rotation, as the cast that follows it.

	// Stores the time at which QueueSpell() was called to ensure that only one spell can be queued up per timestep
	QueueInitiatedAt time.Duration
}

// Models the use of /cqs macros to change which spell should be cast at the last minute
func (qs *QueuedSpell) Cancel(sim *Simulation) {
	if (qs.queueAction != nil) && !qs.queueAction.consumed {
		qs.queueAction.Cancel(sim)
		qs.queueAction = nil
		qs.behindCast = false
	}

	qs.channelAction = nil
	qs.QueueInitiatedAt = -NeverExpires
}

func (unit *Unit) CancelQueuedSpell(sim *Simulation) {
	if unit.QueuedSpell != nil {
		unit.QueuedSpell.Cancel(sim)
	}
}

func (qs *QueuedSpell) InitiateQueue(sim *Simulation, spell *Spell, target *Unit, executeAt time.Duration) {
	qs.Cancel(sim)
	qs.spell = spell
	qs.target = target

	if qs.queueAction == nil {
		if qs.castFn == nil {
			qs.castFn = func(sim *Simulation) {
				qs.spell.castForRotation(sim, qs.target, qs.channelAction)
				// The rotation goes on from a queued instant at once, as it does after one it casts itself, so
				// an action off the GCD, such as a move, does not wait for the next GCD.
				if unit := qs.spell.Unit; unit.Hardcast.Expires <= sim.CurrentTime {
					unit.SetRotationTimer(sim, sim.CurrentTime)
				}
			}
		}
		pa := &PendingAction{
			NextActionAt: executeAt,
			Priority:     ActionPriorityGCD,
			OnAction:     qs.castFn,
		}
		qs.queueAction = pa
	} else {
		qs.queueAction.cancelled = false
		qs.queueAction.NextActionAt = executeAt
	}

	sim.AddPendingAction(qs.queueAction)
	qs.QueueInitiatedAt = sim.CurrentTime
}

func (unit *Unit) QueueSpell(sim *Simulation, spell *Spell, target *Unit, queueAt time.Duration) {
	if unit.QueuedSpell == nil {
		qs := &QueuedSpell{}
		unit.QueuedSpell = qs
	}

	fireAt := queueAt + time.Duration(1) // 1ns artificial delay guarantees last-second cancellation if desired
	unit.QueuedSpell.InitiateQueue(sim, spell, target, fireAt)

	if sim.Log != nil {
		unit.Log(sim, "Queueing up %s to cast at %s.", spell.ActionID, fireAt)
	}
}

// Enforce only one queued spell per timestep
func (unit *Unit) CanQueueSpell(sim *Simulation) bool {
	return (unit.QueuedSpell == nil) || (unit.QueuedSpell.QueueInitiatedAt != sim.CurrentTime)
}

// Returns whether the spell could be queued by the player at the current time using the
// game's spell queueing functionality. Assumes the maximum spell queue window of 400ms
// that the game allows.
func (spell *Spell) CanQueue(sim *Simulation, target *Unit) bool {
	if !spell.canStart(sim, target) {
		return false
	}

	// Apply SQW leniency to any pending hardcasts
	if spell.Flags.Matches(SpellFlagChanneled) && spell.Unit.Hardcast.Expires > sim.CurrentTime+MaxSpellQueueWindow {
		return false
	}

	// Apply SQW leniency to GCD timer
	if spell.DefaultCast.GCD > 0 && spell.Unit.GCD.TimeToReady(sim) > MaxSpellQueueWindow {
		return false
	}

	// Spells that are within one SQW of coming off cooldown can also be queued
	if MaxTimeToReady(spell.CD.Timer, spell.SharedCD.Timer, sim) > MaxSpellQueueWindow {
		return false
	}

	// A spell off the GCD goes once a cast bar that holds the rotation ends, ahead of the cast queued behind it.
	if spell.DefaultCast.EffectiveTime() == 0 && spell.Unit.castBarHoldsRotation(sim) {
		return false
	}

	return spell.meetsCost(sim, false)
}

// Helper function for APL checks to prevent infinite loops
func (spell *Spell) CanCastOrQueue(sim *Simulation, target *Unit) bool {
	return spell.Unit.CanQueueSpell(sim) && !spell.waitsForQueuedCast(sim) && spell.CanQueue(sim, target)
}

// Once a cast bar ends, the cast queued behind it is the rotation's next one: until it goes off, no other
// spell that starts the GCD or a cast bar is cast or queued. A major cooldown still goes, in its place.
func (spell *Spell) waitsForQueuedCast(sim *Simulation) bool {
	return spell.Unit.castQueuedBehindBar() && spell.Unit.Hardcast.Expires <= sim.CurrentTime &&
		spell.DefaultCast.EffectiveTime() > 0 && !spell.Flags.Matches(SpellFlagMCD)
}

// Whether a cast queued under a cast bar has yet to go off.
func (unit *Unit) castQueuedBehindBar() bool {
	qs := unit.QueuedSpell
	return qs != nil && qs.behindCast && !qs.queueAction.IsConsumed()
}

// Whether a cast queued under a cast bar went off at this instant, on its own or in the cast-end look.
func (unit *Unit) castQueuedBehindBarWentOff(sim *Simulation) bool {
	qs := unit.QueuedSpell
	return qs != nil && qs.behindCast && qs.queueAction.IsConsumed() && qs.queueAction != nil &&
		(qs.queueAction.NextActionAt == sim.CurrentTime || qs.queueAction.NextActionAt == sim.CurrentTime+1)
}

// Whether the cast queued behind a cast bar that has ended is due to go off.
func (unit *Unit) castQueuedBehindBarDue(sim *Simulation) bool {
	return unit.castQueuedBehindBar() && unit.Hardcast.Expires <= sim.CurrentTime && unit.QueuedSpell.queueAction.NextActionAt <= sim.CurrentTime+1
}

// Casts the spell queued behind a cast bar that has ended, as its pending action would.
func (unit *Unit) fireQueuedCastBehindBar(sim *Simulation) {
	queueAction := unit.QueuedSpell.queueAction
	queueAction.Cancel(sim)
	queueAction.consumed = true
	queueAction.OnAction(sim)
}

func (spell *Spell) CastOrQueue(sim *Simulation, target *Unit) {
	spell.castOrQueue(sim, target, nil)
}

func (spell *Spell) castOrQueue(sim *Simulation, target *Unit, channelAction *APLActionChannelSpell) {
	if spell.CanCast(sim, target) {
		spell.castForRotation(sim, target, channelAction)
	} else if spell.CanQueue(sim, target) {
		// Determine which timer the spell is waiting on
		queueTime := max(spell.Unit.Hardcast.Expires, BothTimersReadyAt(spell.CD.Timer, spell.SharedCD.Timer))

		if (spell.DefaultCast.GCD > 0) || spell.Flags.Matches(SpellFlagMCD) {
			queueTime = max(queueTime, spell.Unit.GCD.ReadyAt())
		}

		// Schedule the cast to go off without delay
		spell.Unit.QueueSpell(sim, spell, target, queueTime)
		spell.Unit.QueuedSpell.channelAction = channelAction
		spell.Unit.QueuedSpell.behindCast = spell.Unit.castBarHoldsRotation(sim)
	} else {
		// Fallback to make sure there is always log output
		spell.castForRotation(sim, target, channelAction)
	}
}

type rotationCast struct {
	spell         *Spell                 // Until its effects apply or its hardcast ends, completed or not.
	channelAction *APLActionChannelSpell // The channel action that cast it, whose interruptIf and allowRecast the channel it starts takes.
}

// A cast by the rotation or a major cooldown: once it completes it restarts the wand's timer. The record
// outlives the call only while the cast is still running.
func (spell *Spell) castForRotation(sim *Simulation, target *Unit, channelAction *APLActionChannelSpell) {
	unit := spell.Unit
	unit.rotationCast = rotationCast{spell: spell, channelAction: channelAction}
	if !spell.Cast(sim, target) || unit.Hardcast.Expires <= sim.CurrentTime {
		unit.endRotationCast()
	}
}

func (unit *Unit) endRotationCast() {
	unit.rotationCast = rotationCast{}
}

// Hands over the channel action of the cast in flight to a channel starting on spell, the action's own
// or its RelatedDotSpell; nil for any other channel.
func (unit *Unit) takeChannelAction(spell *Spell) *APLActionChannelSpell {
	action := unit.rotationCast.channelAction
	if action == nil || (spell != action.spell && spell != action.spell.RelatedDotSpell) {
		return nil
	}
	unit.rotationCast.channelAction = nil
	return action
}

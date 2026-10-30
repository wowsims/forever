package core

import (
	"fmt"
	"time"
)

// Note that this is only used when the hardcast and GCD actions happen at different times.
func (unit *Unit) newHardcastAction(sim *Simulation) {
	if unit.Metrics.isTanking {
		// While casting, the players dodge, parry and block gets reduced to 0
		unit.HardcastAvoidanceAura.Activate(sim)
	}

	if (unit.hardcastAction != nil) && !unit.hardcastAction.consumed {
		unit.hardcastAction.Cancel(sim)
	}

	if unit.hardcastAction == nil {
		pa := &PendingAction{
			NextActionAt: unit.Hardcast.Expires,
			Priority:     ActionPriorityGCD,
			OnAction:     unit.completeDueHardcast,
		}
		unit.hardcastAction = pa
	} else {
		unit.hardcastAction.cancelled = false
		unit.hardcastAction.NextActionAt = unit.Hardcast.Expires
	}

	sim.AddPendingAction(unit.hardcastAction)
}

func (unit *Unit) completeDueHardcast(sim *Simulation) {
	if hc := &unit.Hardcast; hc.Expires != startingCDTime && hc.Expires <= sim.CurrentTime {
		unit.pickAtCastEnd(sim)
		hc.Expires = startingCDTime
		if hc.OnComplete != nil {
			hc.OnComplete(sim, hc.Target)
		}
		unit.endRotationCast()
	}
}

// Whether the unit's hardcast is still running: until its end and, at the end, while the rotation picks
// what follows it.
func (unit *Unit) hardcastRuns(sim *Simulation) bool {
	return unit.Hardcast.Expires > sim.CurrentTime || unit.pickingAtCastEnd
}

// At the end of a hardcast that holds the rotation, the rotation due now picks its next cast before the
// cast lands, so without its outcome, and queues it for after the cast, dropping any cast queued after
// the hardcast earlier. The pick goes off in the rotation's pass at the cast end, a major cooldown ahead
// of it; what the rotation would do after it waits for that pass. An interactive sim asks for input instead.
func (unit *Unit) pickAtCastEnd(sim *Simulation) {
	rot := unit.Rotation
	if rot == nil || rot.inLoop || unit.rotationAction == nil || sim.Options.Interactive ||
		(sim.CurrentTime < 0 && !unit.engagedBeforePull) || !unit.RotationTimer.IsReady(sim) {
		return
	}
	unit.pickingAtCastEnd = true
	if unit.hardcastHoldsRotation(sim) {
		if unit.spellQueuedAfterHardcast() {
			unit.CancelQueuedSpell(sim)
		}
		// A rotation look an action sets off here, such as on a rage gain, would complete the cast inside
		// the pick.
		rot.inLoop = true
		unit.UpdatePosition(sim, false)
		for i := 0; !unit.spellQueuedAfterHardcast(); i++ {
			action := rot.getNextAction(sim)
			if action == nil {
				break
			}
			if i > 1000 {
				panic(fmt.Sprintf("[USER_ERROR] Infinite loop detected, current action:\n%s", action))
			}
			action.Execute(sim)
		}
		rot.inLoop = false
	}
	unit.pickingAtCastEnd = false
}

func (unit *Unit) NextGCDAt() time.Duration {
	return unit.GCD.ReadyAt()
}

func (unit *Unit) NextRotationActionAt() time.Duration {
	return unit.RotationTimer.ReadyAt()
}

func (unit *Unit) SetGCDTimer(sim *Simulation, gcdReadyAt time.Duration) {
	if unit.rotationAction == nil {
		return
	}

	unit.GCD.Set(gcdReadyAt)

	// A queued spell that lands during a Wait keeps the Wait's earlier wake, so the rotation can
	// still act off the GCD, such as a weaver walking out of melee.
	if waitEnd := unit.NextRotationActionAt(); waitEnd > sim.CurrentTime && waitEnd < gcdReadyAt &&
		unit.Rotation.HasBlockingControllingAction() {
		return
	}
	unit.SetRotationTimer(sim, gcdReadyAt)
}

// Whether a player's hardcast holds the rotation, the GCD ending with it, so that the rotation picks what
// follows it at the cast end. After a cast the GCD outlasts, the rotation looks when the GCD ends; after a
// channel, or a pet's or an enemy's cast, it looks once the cast has landed.
func (unit *Unit) hardcastHoldsRotation(sim *Simulation) bool {
	hc := &unit.Hardcast
	return unit.Type == PlayerUnit && unit.hardcastRuns(sim) && !hc.IsChanneled && unit.GCD.ReadyAt() <= hc.Expires
}

func (unit *Unit) SetRotationTimer(sim *Simulation, rotationReadyAt time.Duration) {
	if unit.rotationAction == nil {
		return
	}

	unit.RotationTimer.Set(rotationReadyAt)

	if !unit.rotationAction.consumed {
		unit.rotationAction.Cancel(sim)
	}

	unit.rotationAction.cancelled = false
	unit.rotationAction.NextActionAt = rotationReadyAt
	sim.AddPendingAction(unit.rotationAction)
}

// Call this when reacting to events that occur before the next scheduled rotation action
func (unit *Unit) ReactToEvent(sim *Simulation, randomizeReactionTime bool, addReactionTime bool) {
	// A stat computation runs no rotation, and an aura gained on reset can still add rage (Furor).
	if unit.Rotation == nil {
		return
	}
	// If the next rotation action was already scheduled for this timestep then execute it now
	unit.Rotation.DoNextAction(sim)

	// A blocking controlling action (Wait / WaitUntil) has set the rotation timer
	// to when it wants to wake; don't shrink it based on reaction time.
	if unit.Rotation.HasBlockingControllingAction() {
		return
	}

	// Otherwise schedule an evaluation based on reaction time
	newEvaluationTime := sim.CurrentTime
	if addReactionTime {
		newEvaluationTime += unit.ReactionTime
	}

	if randomizeReactionTime {
		newEvaluationTime = sim.CurrentTime + DurationFromSeconds(sim.RandomFloat("Reaction Time")*2*unit.ReactionTime.Seconds())
	}

	if unit.NextRotationActionAt() > newEvaluationTime {
		unit.SetRotationTimer(sim, newEvaluationTime)
	}
}

// Call this to stop the GCD loop for a unit.
// This is mostly used for pets that get summoned / expire.
func (unit *Unit) CancelGCDTimer(sim *Simulation) {
	if unit.rotationAction == nil {
		return
	}

	unit.rotationAction.Cancel(sim)
}

// Interrupt cancels any cast, channel or queued spell in progress, e.g. when the
// unit gets crowd controlled. Does not touch the GCD; callers that also want to
// lock the rotation should call ExtendGCDUntil afterwards.
func (unit *Unit) Interrupt(sim *Simulation) {
	if unit.IsChanneling() {
		dot := unit.ChanneledDot
		dot.stopTicking(sim)
		dot.Deactivate(sim)
	}

	if unit.Hardcast.Expires > sim.CurrentTime {
		unit.CancelHardcast(sim)
	}

	unit.CancelQueuedSpell(sim)
}

func (unit *Unit) CancelHardcast(sim *Simulation) {
	unit.HardcastAvoidanceAura.Deactivate(sim)
	unit.Hardcast.Expires = startingCDTime
	unit.AutoAttacks.releaseMeleeHeldByCast(sim)
	unit.endRotationCast()
	unit.SetGCDTimer(sim, sim.CurrentTime+unit.ReactionTime)
}

func (unit *Unit) WaitUntil(sim *Simulation, readyTime time.Duration) {
	if readyTime < sim.CurrentTime {
		panic(unit.Label + ": cannot wait negative time")
	}
	unit.SetRotationTimer(sim, readyTime)
	if sim.Log != nil && readyTime > sim.CurrentTime {
		unit.Log(sim, "Pausing rotation for %s due to resources / CDs.", readyTime-sim.CurrentTime)
	}
}

func (unit *Unit) ExtendGCDUntil(sim *Simulation, readyTime time.Duration) {
	if readyTime < sim.CurrentTime {
		panic(unit.Label + ": cannot wait negative time")
	}
	unit.SetGCDTimer(sim, readyTime)
	if sim.Log != nil && readyTime > sim.CurrentTime {
		unit.Log(sim, "Extending GCD for %s due to rotation / CDs.", readyTime-sim.CurrentTime)
	}
}

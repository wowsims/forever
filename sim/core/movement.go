package core

import (
	"math"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

type MovementAction struct {
	PendingAction
	srcPosition  float64        // starting position
	moveDistance float64        // distance moved, optional
	startTime    time.Duration  // starting time of the movement
	speed        float64        // theoretical movement speed, can be 0
	meleeEdge    *PendingAction // wakes the unit where the walk crosses melee range
}

func (action *MovementAction) GetCurrentPosition(sim *Simulation) float64 {
	return action.srcPosition + float64(sim.CurrentTime-action.startTime)*action.speed/float64(time.Second)
}

func (unit *Unit) initMovement() {
	unit.moveAura = unit.GetOrRegisterAura(Aura{
		Label:     "Movement",
		ActionID:  ActionID{OtherID: proto.OtherAction_OtherActionMove},
		Duration:  NeverExpires,
		MaxStacks: 100,

		OnGain: func(aura *Aura, sim *Simulation) {
			unit.Moving = true
			sim.movingUnits = append(sim.movingUnits, unit)
			unit.AutoAttacks.cancelRangedWindup(sim)
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			unit.Moving = false
			if i := slices.Index(sim.movingUnits, unit); i != -1 {
				sim.movingUnits = slices.Delete(sim.movingUnits, i, i+1)
			}
			unit.movementAction = nil
			unit.AutoAttacks.resumeHeldRangedSwing(sim)
		},
	})

	unit.moveSpell = unit.GetOrRegisterSpell(SpellConfig{
		ActionID: ActionID{OtherID: proto.OtherAction_OtherActionMove},
		Flags:    SpellFlagMeleeMetrics,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			unit.moveAura.Activate(sim)
			unit.moveAura.SetStacks(sim, max(int32(unit.DistanceFromTarget), 1))
		},
	})
}

func (unit *Unit) MoveTo(moveRange float64, sim *Simulation) {
	if moveRange == unit.DistanceFromTarget {
		return
	}

	unit.UpdatePosition(sim, false)
	moveDistance := moveRange - unit.DistanceFromTarget
	timeToMove := time.Duration(math.Abs(moveDistance)/unit.GetMovementSpeed()*1000) * time.Millisecond
	registerMovementAction(unit, sim, unit.GetMovementSpeed()*TernaryFloat64(moveDistance < 0, -1., 1.), sim.CurrentTime+timeToMove, moveDistance)
}

func (unit *Unit) MoveDuration(duration time.Duration, sim *Simulation) {
	if duration == 0 {
		return
	}

	unit.UpdatePosition(sim, false)
	registerMovementAction(unit, sim, 0., sim.CurrentTime+duration, 0)
}

func (unit *Unit) UpdatePosition(sim *Simulation, isFinal bool) {
	if !unit.Moving {
		return
	}

	oldDist := unit.DistanceFromTarget
	if isFinal && unit.movementAction.moveDistance != 0 {
		unit.DistanceFromTarget = unit.movementAction.srcPosition + unit.movementAction.moveDistance
	} else {
		unit.DistanceFromTarget = unit.movementAction.GetCurrentPosition(sim)
	}
	if oldDist == unit.DistanceFromTarget {
		return
	}

	unit.OnMovement(sim, unit.DistanceFromTarget, MovementUpdate)

	// update auto attack state
	if unit.AutoAttacks.mh.enabled != unit.AutoAttacks.mh.IsInRange() {
		if unit.AutoAttacks.mh.IsInRange() {
			unit.AutoAttacks.cancelMeleeWeaveWakeup(sim)
			unit.AutoAttacks.EnableMeleeSwing(sim)
		} else {
			unit.AutoAttacks.CancelMeleeSwing(sim)
			unit.AutoAttacks.scheduleMeleeWeaveWakeup(sim)
		}
	}

	unit.AutoAttacks.UpdateRangedRange(sim)

	yards := max(int32(unit.DistanceFromTarget), 1) // never set to 0 yards as we deactivate the aura
	if yards != unit.moveAura.GetStacks() {
		unit.moveAura.SetStacks(sim, yards)
	}
}

func (unit *Unit) FinalizeMovement(sim *Simulation) {
	if !unit.Moving {
		return
	}

	unit.UpdatePosition(sim, true)
	unit.moveAura.Deactivate(sim)

	unit.OnMovement(sim, unit.DistanceFromTarget, MovementEnd)
}

func registerMovementAction(unit *Unit, sim *Simulation, speed float64, endTime time.Duration, moveDistance float64) {
	if unit.movementAction != nil {
		unit.movementAction.Cancel(sim)
		if unit.movementAction.meleeEdge != nil {
			unit.movementAction.meleeEdge.Cancel(sim)
		}
	} else {
		unit.moveSpell.Cast(sim, unit.CurrentTarget)
	}

	movementAction := MovementAction{
		startTime:    sim.CurrentTime,
		speed:        speed,
		srcPosition:  unit.DistanceFromTarget,
		moveDistance: moveDistance,
	}

	movementAction.NextActionAt = endTime
	movementAction.OnAction = func(sim *Simulation) {
		unit.FinalizeMovement(sim)
	}

	unit.OnMovement(sim, unit.DistanceFromTarget, MovementStart)
	unit.movementAction = &movementAction
	sim.AddPendingAction(&movementAction.PendingAction)

	// Every event sees the current position, but a melee swing has to start or stop where the walk
	// crosses melee range even when nothing else happens there, so the walk wakes the unit on the line.
	// A ranged auto needs no wake, as it never fires on the move.
	edge, src, dest := unit.AutoAttacks.mh.MaxRange, movementAction.srcPosition, movementAction.srcPosition+moveDistance
	if unit.AutoAttacks.AutoSwingMelee && speed != 0 && edge != dest && (edge-src)*(edge-dest) <= 0 {
		movementAction.meleeEdge = &PendingAction{
			// A hair past the line, as standing on it is still in range.
			NextActionAt: sim.CurrentTime + DurationFromSeconds((edge-src)/speed) + time.Microsecond,
			OnAction:     func(sim *Simulation) {}, // advance moves the unit there
		}
		sim.AddPendingAction(movementAction.meleeEdge)
	}
}

type MovementUpdateType byte

const (
	MovementStart = iota
	MovementUpdate
	MovementEnd
)

type MovementCallback func(sim *Simulation, position float64, kind MovementUpdateType)

func (unit *Unit) RegisterMovementCallback(callback MovementCallback) {
	unit.movementCallbacks = append(unit.movementCallbacks, callback)
}

func (unit *Unit) OnMovement(sim *Simulation, position float64, kind MovementUpdateType) {
	for _, movementCallback := range unit.movementCallbacks {
		movementCallback(sim, position, kind)
	}
}

func (unit *Unit) MultiplyMovementSpeed(sim *Simulation, amount float64) {
	oldMultiplier := unit.PseudoStats.MovementSpeedMultiplier
	oldSpeed := unit.GetMovementSpeed()
	unit.PseudoStats.MovementSpeedMultiplier *= amount
	if sim.Log != nil {
		unit.Log(sim, "[DEBUG] Movement speed changed from %.2f (%.2f%%) to %.2f (%.2f%%)", oldSpeed, (oldMultiplier-1)*100.0, unit.GetMovementSpeed(), (unit.PseudoStats.MovementSpeedMultiplier-1)*100.0)
	}

	// A walk under way keeps its destination and covers the rest at the new speed.
	if unit.movementAction != nil && unit.movementAction.speed != 0 {
		dest := unit.movementAction.srcPosition + unit.movementAction.moveDistance
		unit.MoveTo(dest, sim)
	}
}

// Returns the units current movement speed in yards / second
func (unit *Unit) GetMovementSpeed() float64 {
	if unit.Type == PlayerUnit {
		return 7. * unit.PseudoStats.MovementSpeedMultiplier
	}

	return 8. * unit.PseudoStats.MovementSpeedMultiplier
}

func (unit *Unit) NewPassiveMovementSpeedAura(label string, actionID ActionID, multiplier float64) *Aura {
	aura := MakePermanent(unit.GetOrRegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
	}))

	aura.NewPassiveMovementSpeedEffect(multiplier)

	return aura
}

func (aura *Aura) NewPassiveMovementSpeedEffect(multiplier float64) *ExclusiveEffect {
	return aura.newExclusiveMovementSpeedEffect("PassiveMovementSpeed", multiplier)
}

func (aura *Aura) NewActiveMovementSpeedEffect(multiplier float64) *ExclusiveEffect {
	return aura.newExclusiveMovementSpeedEffect("ActiveMovementSpeed", multiplier)
}

// Only the highest speed in a category applies, and the auras under it stay up: one that is outbid
// takes over again when the higher one ends.
func (aura *Aura) newExclusiveMovementSpeedEffect(categoryKey string, multiplier float64) *ExclusiveEffect {
	return aura.NewExclusiveEffect(categoryKey, false, ExclusiveEffect{
		Priority: multiplier,
		OnGain: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyMovementSpeed(sim, 1+multiplier)
		},
		OnExpire: func(ee *ExclusiveEffect, sim *Simulation) {
			ee.Aura.Unit.MultiplyMovementSpeed(sim, 1.0/(1+multiplier))
		},
	})
}

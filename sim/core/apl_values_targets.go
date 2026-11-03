package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// The enemy target behind a unit, or nil for a unit that is not one.
func enemyTargetOf(unit *Unit) *Target {
	if unit == nil || unit.Type != EnemyUnit {
		return nil
	}
	return unit.Env.GetTargetByIndex(unit.Index)
}

// The time left until an absolute time, NeverExpires for none and 0 for one that has passed.
func timeUntil(sim *Simulation, at time.Duration) time.Duration {
	if at == NeverExpires {
		return NeverExpires
	}
	return max(at-sim.CurrentTime, 0)
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
		return state.wave.occurrences.nextAt()
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
			if next := state.wave.occurrences.nextAt(); next != NeverExpires {
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

// Resolves the reference to an enemy target of the encounter, with a warning for a unit that is not one.
func (rot *APLRotation) getDesignTargetUnit(ref *proto.UnitReference, uuid *proto.UUID) (UnitReference, bool) {
	unit := rot.GetTargetUnit(ref)
	resolvedUnit := unit.Get()
	if resolvedUnit == nil {
		return unit, false
	}
	if enemyTargetOf(resolvedUnit) == nil {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s is not an enemy target", resolvedUnit.Label)
		return unit, false
	}
	return unit, true
}

type APLValueTargetIsActive struct {
	DefaultAPLValueImpl
	target UnitReference
}

func (rot *APLRotation) newValueTargetIsActive(config *proto.APLValueTargetIsActive, uuid *proto.UUID) APLValue {
	unit, ok := rot.getDesignTargetUnit(config.TargetUnit, uuid)
	if !ok {
		return nil
	}
	return &APLValueTargetIsActive{target: unit}
}
func (value *APLValueTargetIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueTargetIsActive) GetBool(sim *Simulation) bool {
	target := enemyTargetOf(value.target.Get())
	return target != nil && target.IsEnabled()
}
func (value *APLValueTargetIsActive) String() string {
	return fmt.Sprintf("Target Is Active(%s)", value.target.Get().Label)
}

type APLValueNumActiveAdds struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueNumActiveAdds(_ *proto.APLValueNumActiveAdds, _ *proto.UUID) APLValue {
	return &APLValueNumActiveAdds{}
}
func (value *APLValueNumActiveAdds) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueNumActiveAdds) GetInt(sim *Simulation) int32 {
	return int32(sim.Encounter.NumActiveAdds())
}
func (value *APLValueNumActiveAdds) String() string {
	return "Num Active Adds"
}

type APLValueTimeUntilTargetActive struct {
	DefaultAPLValueImpl
	target UnitReference
}

func (rot *APLRotation) newValueTimeUntilTargetActive(config *proto.APLValueTimeUntilTargetActive, uuid *proto.UUID) APLValue {
	unit, ok := rot.getDesignTargetUnit(config.TargetUnit, uuid)
	if !ok {
		return nil
	}
	return &APLValueTimeUntilTargetActive{target: unit}
}
func (value *APLValueTimeUntilTargetActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTimeUntilTargetActive) GetDuration(sim *Simulation) time.Duration {
	target := enemyTargetOf(value.target.Get())
	if target == nil {
		return NeverExpires
	}
	return timeUntil(sim, target.NextActiveAt(sim))
}
func (value *APLValueTimeUntilTargetActive) String() string {
	return fmt.Sprintf("Time Until Target Active(%s)", value.target.Get().Label)
}

type APLValueTimeUntilTargetDies struct {
	DefaultAPLValueImpl
	target UnitReference
}

func (rot *APLRotation) newValueTimeUntilTargetDies(config *proto.APLValueTimeUntilTargetDies, uuid *proto.UUID) APLValue {
	unit, ok := rot.getDesignTargetUnit(config.TargetUnit, uuid)
	if !ok {
		return nil
	}
	return &APLValueTimeUntilTargetDies{target: unit}
}
func (value *APLValueTimeUntilTargetDies) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTimeUntilTargetDies) GetDuration(sim *Simulation) time.Duration {
	target := enemyTargetOf(value.target.Get())
	if target == nil {
		return NeverExpires
	}
	return timeUntil(sim, target.DiesAt(sim))
}
func (value *APLValueTimeUntilTargetDies) String() string {
	return fmt.Sprintf("Time Until Target Dies(%s)", value.target.Get().Label)
}

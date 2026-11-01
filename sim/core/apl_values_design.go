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
	return unit.Env.Encounter.AllTargets[unit.Index]
}

// The time left until an absolute time, NeverExpires for none and 0 for one that has passed.
func timeUntil(sim *Simulation, at time.Duration) time.Duration {
	if at == NeverExpires {
		return NeverExpires
	}
	return max(at-sim.CurrentTime, 0)
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

func (rot *APLRotation) validAmpType(ampType proto.DamageAmpType, uuid *proto.UUID) bool {
	if ampType < 0 || int(ampType) >= numDamageAmpTypes {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Unknown damage amp type %d", int32(ampType))
		return false
	}
	return true
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
	return enemyTargetOf(value.target.Get()) != nil && value.target.Get().IsEnabled()
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

type APLValueTimeUntilNextMove struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueTimeUntilNextMove(_ *proto.APLValueTimeUntilNextMove, _ *proto.UUID) APLValue {
	return &APLValueTimeUntilNextMove{}
}
func (value *APLValueTimeUntilNextMove) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTimeUntilNextMove) GetDuration(sim *Simulation) time.Duration {
	return timeUntil(sim, sim.Encounter.NextMoveAt(sim))
}
func (value *APLValueTimeUntilNextMove) String() string {
	return "Time Until Next Move"
}

type APLValueNextMoveDuration struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueNextMoveDuration(_ *proto.APLValueNextMoveDuration, _ *proto.UUID) APLValue {
	return &APLValueNextMoveDuration{}
}
func (value *APLValueNextMoveDuration) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueNextMoveDuration) GetDuration(sim *Simulation) time.Duration {
	return sim.Encounter.NextMoveDuration(sim)
}
func (value *APLValueNextMoveDuration) String() string {
	return "Next Move Duration"
}

type APLValueAmpIsActive struct {
	DefaultAPLValueImpl
	ampType proto.DamageAmpType
}

func (rot *APLRotation) newValueAmpIsActive(config *proto.APLValueAmpIsActive, uuid *proto.UUID) APLValue {
	if !rot.validAmpType(config.AmpType, uuid) {
		return nil
	}
	return &APLValueAmpIsActive{ampType: config.AmpType}
}
func (value *APLValueAmpIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueAmpIsActive) GetBool(sim *Simulation) bool {
	return sim.Encounter.AmpActive(sim, value.ampType)
}
func (value *APLValueAmpIsActive) String() string {
	return fmt.Sprintf("Amp Is Active(%s)", value.ampType)
}

type APLValueTimeUntilAmp struct {
	DefaultAPLValueImpl
	ampType proto.DamageAmpType
}

func (rot *APLRotation) newValueTimeUntilAmp(config *proto.APLValueTimeUntilAmp, uuid *proto.UUID) APLValue {
	if !rot.validAmpType(config.AmpType, uuid) {
		return nil
	}
	return &APLValueTimeUntilAmp{ampType: config.AmpType}
}
func (value *APLValueTimeUntilAmp) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTimeUntilAmp) GetDuration(sim *Simulation) time.Duration {
	return timeUntil(sim, sim.Encounter.NextAmpAt(sim, value.ampType))
}
func (value *APLValueTimeUntilAmp) String() string {
	return fmt.Sprintf("Time Until Amp(%s)", value.ampType)
}

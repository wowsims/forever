package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

type APLValueCurrentTime struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueCurrentTime(config *proto.APLValueCurrentTime, _ *proto.UUID) APLValue {
	return &APLValueCurrentTime{}
}
func (value *APLValueCurrentTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueCurrentTime) GetDuration(sim *Simulation) time.Duration {
	return sim.CurrentTime
}
func (value *APLValueCurrentTime) String() string {
	return "Current Time"
}

type APLValueCurrentTimePercent struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueCurrentTimePercent(config *proto.APLValueCurrentTimePercent, _ *proto.UUID) APLValue {
	return &APLValueCurrentTimePercent{}
}
func (value *APLValueCurrentTimePercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentTimePercent) GetFloat(sim *Simulation) float64 {
	return sim.CurrentTime.Seconds() / sim.Duration.Seconds()
}
func (value *APLValueCurrentTimePercent) String() string {
	return fmt.Sprintf("Current Time %%")
}

type APLValueRemainingTime struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueRemainingTime(config *proto.APLValueRemainingTime, _ *proto.UUID) APLValue {
	return &APLValueRemainingTime{}
}
func (value *APLValueRemainingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueRemainingTime) GetDuration(sim *Simulation) time.Duration {
	return sim.GetRemainingDuration()
}
func (value *APLValueRemainingTime) String() string {
	return "Remaining Time"
}

type APLValueRemainingTimePercent struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueRemainingTimePercent(config *proto.APLValueRemainingTimePercent, _ *proto.UUID) APLValue {
	return &APLValueRemainingTimePercent{}
}
func (value *APLValueRemainingTimePercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueRemainingTimePercent) GetFloat(sim *Simulation) float64 {
	return sim.GetRemainingDurationPercent()
}
func (value *APLValueRemainingTimePercent) String() string {
	return fmt.Sprintf("Remaining Time %%")
}

type APLValueNumberTargets struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueNumberTargets(config *proto.APLValueNumberTargets, _ *proto.UUID) APLValue {
	return &APLValueNumberTargets{}
}
func (value *APLValueNumberTargets) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueNumberTargets) GetInt(sim *Simulation) int32 {
	return sim.ActiveTargetCount()
}
func (value *APLValueNumberTargets) GetFloat(sim *Simulation) float64 {
	return float64(sim.ActiveTargetCount())
}
func (value *APLValueNumberTargets) String() string {
	return "Num Active Targets"
}

type APLValueIsExecutePhase struct {
	DefaultAPLValueImpl
	target    UnitReference
	threshold proto.APLValueIsExecutePhase_ExecutePhaseThreshold
}

func (rot *APLRotation) newValueIsExecutePhase(config *proto.APLValueIsExecutePhase, uuid *proto.UUID) APLValue {
	switch config.Threshold {
	case proto.APLValueIsExecutePhase_Unknown:
		return nil
	case proto.APLValueIsExecutePhase_E20, proto.APLValueIsExecutePhase_E35:
	default:
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Unknown execute phase threshold")
		return nil
	}
	target := rot.GetTargetUnit(config.TargetUnit)
	if target.Get() == nil {
		return nil
	}
	return &APLValueIsExecutePhase{
		target:    target,
		threshold: config.Threshold,
	}
}
func (value *APLValueIsExecutePhase) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueIsExecutePhase) GetBool(sim *Simulation) bool {
	if value.threshold == proto.APLValueIsExecutePhase_E20 {
		return value.target.Get().IsExecutePhase20()
	}
	return value.target.Get().IsExecutePhase35()
}
func (value *APLValueIsExecutePhase) String() string {
	return "Is Execute Phase"
}

type APLValueEncounterInArea struct {
	DefaultAPLValueImpl
	areaType proto.AreaType
}

func (rot *APLRotation) newValueEncounterInArea(config *proto.APLValueEncounterInArea, uuid *proto.UUID) APLValue {
	if config.AreaType == proto.AreaType_AreaTypeUnknown {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Unknown area type")
		return nil
	}
	return &APLValueEncounterInArea{
		areaType: config.AreaType,
	}
}
func (value *APLValueEncounterInArea) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueEncounterInArea) GetBool(sim *Simulation) bool {
	return sim.Encounter.InArea(value.areaType)
}
func (value *APLValueEncounterInArea) String() string {
	return fmt.Sprintf("In Area(%s)", value.areaType)
}

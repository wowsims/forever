package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// An aura as the unit's rotation reads it.
type auraReading struct {
	active    bool
	remaining time.Duration
	stacks    int32
	landing   bool // Read as the unit's running hardcast leaves it when it lands.
}

func (unit *Unit) readAura(sim *Simulation, aura *Aura) auraReading {
	if aura == nil {
		return auraReading{}
	}
	return unit.readLanding(sim, aura, func() time.Duration { return aura.Duration })
}

// A dot as its caster's rotation reads it, landing with the duration an application now would give it.
func readDot(sim *Simulation, dot *Dot) auraReading {
	if dot == nil {
		return auraReading{}
	}
	return dot.Spell.Unit.readLanding(sim, dot.Aura, dot.freshDuration)
}

// While the unit hardcasts the aura's own spell on the aura's unit, the rotation reads the aura as if the
// cast had landed at the cast end: up for duration from then and, for a stacking aura, with one more
// stack up to its maximum, or one. Any other aura, a proc the cast triggers among them, reads as it is.
func (unit *Unit) readLanding(sim *Simulation, aura *Aura, duration func() time.Duration) auraReading {
	hc := &unit.Hardcast
	if !unit.hardcastRuns(sim) || hc.IsChanneled || hc.Target != aura.Unit || hc.ActionID.IsEmptyAction() ||
		!hc.ActionID.SameActionIgnoreTag(aura.ActionID) {
		return auraReading{active: aura.IsActive(), remaining: aura.RemainingDuration(sim), stacks: aura.GetStacks()}
	}
	reading := auraReading{active: true, remaining: NeverExpires, stacks: aura.GetStacks(), landing: true}
	if d := duration(); d != NeverExpires {
		reading.remaining = hc.Expires + d - sim.CurrentTime
	}
	if aura.MaxStacks > 0 {
		reading.stacks = TernaryInt32(aura.IsActive(), min(aura.GetStacks()+1, aura.MaxStacks), 1)
	}
	return reading
}

func (rot *APLRotation) newValueAuraIsKnown(config *proto.APLValueAuraIsKnown, uuid *proto.UUID) APLValue {
	aura := rot.GetAPLAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	return rot.newValueConst(&proto.APLValueConst{
		Val: Ternary(aura.Get() != nil, "true", "false"),
	}, uuid)
}

type APLValueAuraIsActive struct {
	DefaultAPLValueImpl
	unit                *Unit
	aura                AuraReference
	reactionTime        time.Duration
	includeReactionTime bool
}

func (rot *APLRotation) newValueAuraIsActive(config *proto.APLValueAuraIsActive, _ *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}
	return &APLValueAuraIsActive{
		unit:                rot.unit,
		aura:                aura,
		reactionTime:        rot.unit.ReactionTime,
		includeReactionTime: config.IncludeReactionTime,
	}
}
func (value *APLValueAuraIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueAuraIsActive) GetBool(sim *Simulation) bool {
	aura := value.aura.Get()
	if reading := value.unit.readAura(sim, aura); reading.landing {
		return reading.active
	}
	if value.includeReactionTime {
		return aura.IsActive() && aura.TimeActive(sim) >= value.reactionTime
	}
	return aura.IsActive()
}
func (value *APLValueAuraIsActive) String() string {
	return fmt.Sprintf("Aura Active(%s)", value.aura.String())
}

type APLValueAuraIsInactive struct {
	DefaultAPLValueImpl
	unit                *Unit
	aura                AuraReference
	reactionTime        time.Duration
	includeReactionTime bool
}

func (rot *APLRotation) newValueAuraIsInactive(config *proto.APLValueAuraIsInactive, _ *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}

	return &APLValueAuraIsInactive{
		unit:                rot.unit,
		aura:                aura,
		reactionTime:        rot.unit.ReactionTime,
		includeReactionTime: config.IncludeReactionTime,
	}
}
func (value *APLValueAuraIsInactive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueAuraIsInactive) GetBool(sim *Simulation) bool {
	aura := value.aura.Get()
	if reading := value.unit.readAura(sim, aura); reading.landing {
		return !reading.active
	}
	if value.includeReactionTime {
		return !aura.IsActive() && aura.TimeInactive(sim) >= value.reactionTime
	}
	return !aura.IsActive()
}
func (value *APLValueAuraIsInactive) String() string {
	return fmt.Sprintf("Aura Inactive(%s)", value.aura.String())
}

type APLValueAuraRemainingTime struct {
	DefaultAPLValueImpl
	unit *Unit
	aura AuraReference
}

func (rot *APLRotation) newValueAuraRemainingTime(config *proto.APLValueAuraRemainingTime, _ *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}
	return &APLValueAuraRemainingTime{
		unit: rot.unit,
		aura: aura,
	}
}
func (value *APLValueAuraRemainingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAuraRemainingTime) GetDuration(sim *Simulation) time.Duration {
	return value.unit.readAura(sim, value.aura.Get()).remaining
}
func (value *APLValueAuraRemainingTime) String() string {
	return fmt.Sprintf("Aura Remaining Time(%s)", value.aura.String())
}

type APLValueAuraNumStacks struct {
	DefaultAPLValueImpl
	unit                *Unit
	aura                AuraReference
	reactionTime        time.Duration
	includeReactionTime bool

	stackHistories []auraStackHistory // Indexed by UnitIndex, one per unit the aura reference can resolve to.
}

type auraStackHistory struct {
	stackUpdateTime time.Duration
	stacks          int32
	previousStacks  int32
}

func (rot *APLRotation) newValueAuraNumStacks(config *proto.APLValueAuraNumStacks, uuid *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	resolvedAura := aura.Get()
	if resolvedAura == nil {
		return nil
	}
	if resolvedAura.MaxStacks == 0 {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "%s is not a stackable aura", ProtoToActionID(config.AuraId))
		return nil
	}

	value := &APLValueAuraNumStacks{
		unit:                rot.unit,
		aura:                aura,
		reactionTime:        rot.unit.ReactionTime,
		includeReactionTime: config.IncludeReactionTime,
		stackHistories:      make([]auraStackHistory, len(rot.unit.Env.AllUnits)),
	}

	auras := aura.allTargetAuras
	if aura.fixedAura != nil {
		auras = AuraArray{aura.fixedAura}
	}
	for _, unitAura := range auras {
		unitAura.ApplyOnStacksChange(func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
			history := &value.stackHistories[aura.Unit.UnitIndex]
			if sim.CurrentTime-history.stackUpdateTime >= value.reactionTime {
				history.previousStacks = oldStacks
			}
			history.stackUpdateTime = sim.CurrentTime
			history.stacks = newStacks
		}).ApplyOnReset(func(aura *Aura, sim *Simulation) {
			history := &value.stackHistories[aura.Unit.UnitIndex]
			history.stackUpdateTime = NeverExpires
			history.previousStacks = aura.GetStacks()
			history.stacks = aura.GetStacks()
		})
	}

	return value
}
func (value *APLValueAuraNumStacks) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueAuraNumStacks) GetInt(sim *Simulation) int32 {
	aura := value.aura.Get()
	if aura == nil {
		return 0
	}
	if reading := value.unit.readAura(sim, aura); reading.landing {
		return reading.stacks
	}
	history := &value.stackHistories[aura.Unit.UnitIndex]
	if value.includeReactionTime {
		return TernaryInt32(sim.CurrentTime-history.stackUpdateTime >= value.reactionTime, history.stacks, history.previousStacks)
	}
	return history.stacks
}
func (value *APLValueAuraNumStacks) String() string {
	return fmt.Sprintf("Aura Num Stacks(%s)", value.aura.String())
}

type APLValueAuraInternalCooldown struct {
	DefaultAPLValueImpl
	aura AuraReference
}

func (rot *APLRotation) newValueAuraInternalCooldown(config *proto.APLValueAuraInternalCooldown, _ *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLICDAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}
	return &APLValueAuraInternalCooldown{
		aura: aura,
	}
}
func (value *APLValueAuraInternalCooldown) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAuraInternalCooldown) GetDuration(sim *Simulation) time.Duration {
	return value.aura.Get().Icd.TimeToReady(sim)
}
func (value *APLValueAuraInternalCooldown) String() string {
	return fmt.Sprintf("Aura Remaining ICD(%s)", value.aura.String())
}

type APLValueAuraICDIsReady struct {
	DefaultAPLValueImpl
	aura                AuraReference
	reactionTime        time.Duration
	includeReactionTime bool
}

func (rot *APLRotation) newValueAuraICDIsReady(config *proto.APLValueAuraICDIsReady, _ *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLICDAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}
	return &APLValueAuraICDIsReady{
		aura:                aura,
		reactionTime:        rot.unit.ReactionTime,
		includeReactionTime: config.IncludeReactionTime,
	}
}
func (value *APLValueAuraICDIsReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueAuraICDIsReady) GetBool(sim *Simulation) bool {
	aura := value.aura.Get()
	if value.includeReactionTime {
		return aura.Icd.IsReady(sim) || (aura.IsActive() && aura.TimeActive(sim) < value.reactionTime)
	}
	return aura.Icd.IsReady(sim)
}
func (value *APLValueAuraICDIsReady) String() string {
	return fmt.Sprintf("Aura ICD Is Ready(%s)", value.aura.String())
}

type APLValueAuraShouldRefresh struct {
	DefaultAPLValueImpl
	aura       AuraReference
	maxOverlap APLValue
}

func (rot *APLRotation) newValueAuraShouldRefresh(config *proto.APLValueAuraShouldRefresh, uuid *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLAura(rot.GetTargetUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}

	maxOverlap := rot.coerceTo(rot.newAPLValue(config.MaxOverlap), proto.APLValueType_ValueTypeDuration)
	if maxOverlap == nil {
		maxOverlap = rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, uuid)
	}

	return &APLValueAuraShouldRefresh{
		aura:       aura,
		maxOverlap: maxOverlap,
	}
}
func (value *APLValueAuraShouldRefresh) GetInnerValues() []APLValue {
	return []APLValue{value.maxOverlap}
}
func (value *APLValueAuraShouldRefresh) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueAuraShouldRefresh) GetBool(sim *Simulation) bool {
	return value.aura.Get().ShouldRefreshExclusiveEffects(sim, value.maxOverlap.GetDuration(sim))
}
func (value *APLValueAuraShouldRefresh) String() string {
	return fmt.Sprintf("Should Refresh Aura(%s)", value.aura.String())
}

type APLValueAuraDuration struct {
	DefaultAPLValueImpl
	aura AuraReference
}

func (rot *APLRotation) newValueAuraDuration(config *proto.APLValueAuraDuration, _ *proto.UUID) APLValue {
	if config.AuraId == nil {
		return nil
	}
	aura := rot.GetAPLAura(rot.GetSourceUnit(config.SourceUnit), config.AuraId)
	if aura.Get() == nil {
		return nil
	}
	return &APLValueAuraDuration{
		aura: aura,
	}
}
func (value *APLValueAuraDuration) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueAuraDuration) GetDuration(sim *Simulation) time.Duration {
	return value.aura.Get().Duration
}
func (value *APLValueAuraDuration) String() string {
	return fmt.Sprintf("Aura Duration(%s)", value.aura.String())
}

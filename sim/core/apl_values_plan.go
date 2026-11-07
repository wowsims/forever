package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func (rot *APLRotation) validAmpType(ampType proto.DamageAmpType, uuid *proto.UUID) bool {
	if ampType < 0 || int(ampType) >= numDamageAmpTypes {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Unknown damage amp type %d", int32(ampType))
		return false
	}
	return true
}

// The boss abilities of a name that the four boss ability values read. With no unit set, those on any unit; with
// one set, those whose target it is, a fixed target picked when the value is built and a current, previous or next
// target at each read. An Environment Pulse or Area Hit comes from no target, so a set unit never matches one; a
// Boss one is its target's. No value reads an ability construction skipped (its construct note says why): it is
// never active, never next, never casts and is not known. Its name still counts as one the plan has, so it draws no
// unknown-name warning.
type planAbilityFilter struct {
	abilities  []*planBossAbility   // Every read's, with no unit or a fixed target.
	byTarget   [][]*planBossAbility // With a current, previous or next target, the abilities on each unit by unit index.
	kept       bool                 // Whether any ability was kept.
	unit       UnitReference        // The set unit; the zero reference for none.
	named      bool                 // Whether the plan has an ability of the name on the unit, skipped or not kept.
	targetless bool                 // Whether the plan has an ability of the name and each one is an Environment Pulse or Area Hit.
}

// The filter of the name and unit with the abilities keep accepts, every one for a nil keep. Nil for a unit that
// resolves to no enemy target when the value is built, which getUnit or getPlanTargetUnit warns about. No name
// names none of them.
func (rot *APLRotation) planAbilityFilter(name string, targetRef *proto.UnitReference, keep func(*planBossAbility) bool, uuid *proto.UUID) *planAbilityFilter {
	filter := &planAbilityFilter{}
	targeted := targetRef != nil && targetRef.Type != proto.UnitReference_Unknown
	var fixed *Target
	if targeted {
		unit, ok := rot.getPlanTargetUnit(targetRef, uuid)
		if !ok {
			return nil
		}
		filter.unit = unit
		if unit.fixedUnit != nil {
			fixed = enemyTargetOf(unit.fixedUnit)
		} else {
			filter.byTarget = make([][]*planBossAbility, len(rot.unit.Env.AllUnits))
		}
	}
	plan := rot.unit.Env.Encounter.plan
	if plan == nil || name == "" {
		return filter
	}
	found, onTarget := false, false
	for _, ability := range plan.bossAbilities {
		if ability.config.Name != name {
			continue
		}
		found = true
		if !ability.targetless() {
			onTarget = true
		}
		if (targeted && ability.targetless()) || (fixed != nil && ability.target != fixed) {
			continue
		}
		filter.named = true
		if ability.skip || (keep != nil && !keep(ability)) {
			continue
		}
		filter.kept = true
		if filter.byTarget != nil {
			filter.byTarget[ability.target.UnitIndex] = append(filter.byTarget[ability.target.UnitIndex], ability)
		} else {
			filter.abilities = append(filter.abilities, ability)
		}
	}
	filter.targetless = found && !onTarget
	return filter
}

// The filter of the name and unit, or nil with a warning when the plan has no ability of the name on the unit.
func (rot *APLRotation) planBossAbilities(name string, targetRef *proto.UnitReference, keep func(*planBossAbility) bool, uuid *proto.UUID) *planAbilityFilter {
	filter := rot.planAbilityFilter(name, targetRef, keep, uuid)
	if filter == nil || filter.named {
		return filter
	}
	switch {
	case filter.unit == (UnitReference{}):
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No boss ability named %q", name)
	case filter.targetless:
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Boss ability %q is only an Environment Pulse or Area Hit, which comes from no target and never matches a set unit", name)
	default:
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No boss ability named %q on %s", name, planUnitName(filter.unit))
	}
	return nil
}

// The unit as a warning names it: the reference for a current, previous or next target, else its label.
func planUnitName(unit UnitReference) string {
	switch unit.Type {
	case proto.UnitReference_CurrentTarget:
		return "the Current Target"
	case proto.UnitReference_PreviousTarget:
		return "the Previous Target"
	case proto.UnitReference_NextTarget:
		return "the Next Target"
	}
	return unit.Get().Label
}

// Whether no read finds an ability.
func (filter *planAbilityFilter) empty() bool {
	return !filter.kept
}

// The abilities one read takes: with a current, previous or next target, those on the target it resolves to now.
func (filter *planAbilityFilter) read() []*planBossAbility {
	if filter.byTarget == nil {
		return filter.abilities
	}
	unit := filter.unit.Get()
	if unit == nil {
		return nil
	}
	return filter.byTarget[unit.UnitIndex]
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

type APLValueBossAbilityIsKnown struct {
	DefaultAPLValueImpl
	name   string
	filter *planAbilityFilter
}

// With no unit or a fixed target, folds to a constant when the rotation is built, so a line it guards prunes. With a
// current, previous or next target it reads at each evaluation, and folds to false when no target has an ability of
// the name. A name the plan lacks reads false with no warning, since absence is what the value tests.
func (rot *APLRotation) newValueBossAbilityIsKnown(config *proto.APLValueBossAbilityIsKnown, uuid *proto.UUID) APLValue {
	filter := rot.planAbilityFilter(config.AbilityName, config.TargetUnit, nil, uuid)
	if filter != nil && filter.byTarget != nil && filter.kept {
		return &APLValueBossAbilityIsKnown{name: config.AbilityName, filter: filter}
	}
	known := filter != nil && filter.kept
	return rot.newValueConst(&proto.APLValueConst{Val: Ternary(known, "true", "false")}, uuid)
}
func (value *APLValueBossAbilityIsKnown) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossAbilityIsKnown) GetBool(_ *Simulation) bool {
	return len(value.filter.read()) > 0
}
func (value *APLValueBossAbilityIsKnown) String() string {
	return fmt.Sprintf("Boss Ability Is Known(%s)", value.name)
}

type APLValueBossAbilityIsActive struct {
	DefaultAPLValueImpl
	name   string
	filter *planAbilityFilter
}

func (rot *APLRotation) newValueBossAbilityIsActive(config *proto.APLValueBossAbilityIsActive, uuid *proto.UUID) APLValue {
	filter := rot.planBossAbilities(config.AbilityName, config.TargetUnit, nil, uuid)
	if filter == nil {
		return nil
	}
	return &APLValueBossAbilityIsActive{name: config.AbilityName, filter: filter}
}
func (value *APLValueBossAbilityIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossAbilityIsActive) GetBool(_ *Simulation) bool {
	for _, ability := range value.filter.read() {
		if ability.open > 0 {
			return true
		}
	}
	return false
}
func (value *APLValueBossAbilityIsActive) String() string {
	return fmt.Sprintf("Boss Ability Is Active(%s)", value.name)
}

type APLValueTimeUntilBossAbility struct {
	DefaultAPLValueImpl
	name   string
	filter *planAbilityFilter
}

func (rot *APLRotation) newValueTimeUntilBossAbility(config *proto.APLValueTimeUntilBossAbility, uuid *proto.UUID) APLValue {
	filter := rot.planBossAbilities(config.AbilityName, config.TargetUnit, nil, uuid)
	if filter == nil {
		return nil
	}
	return &APLValueTimeUntilBossAbility{name: config.AbilityName, filter: filter}
}
func (value *APLValueTimeUntilBossAbility) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueTimeUntilBossAbility) GetDuration(sim *Simulation) time.Duration {
	next := NeverExpires
	for _, ability := range value.filter.read() {
		next = min(next, ability.nextAt(sim))
	}
	return timeUntil(sim, next)
}
func (value *APLValueTimeUntilBossAbility) String() string {
	return fmt.Sprintf("Time Until Boss Ability(%s)", value.name)
}

type APLValueBossAbilityCastTimeRemaining struct {
	DefaultAPLValueImpl
	name   string
	filter *planAbilityFilter
}

// Whether the ability is a Spell Hit, or a Boss Pulse or Area Hit, with a hardcast. One with no cast time casts
// instantly.
func planAbilityHardcasts(ability *planBossAbility) bool {
	casts := ability.config.Kind == proto.PlanBossAbility_SpellHit || (ability.isArea() && !ability.targetless())
	return casts && ability.spell.DefaultCast.CastTime > 0
}

// Reads the abilities of the filter that hardcast. A name with none of them, only the other kinds, ones with no cast
// time or ones construction skipped, reads a constant 0ms with a warning, so a line on it folds instead of losing its
// condition.
func (rot *APLRotation) newValueBossAbilityCastTimeRemaining(config *proto.APLValueBossAbilityCastTimeRemaining, uuid *proto.UUID) APLValue {
	filter := rot.planBossAbilities(config.AbilityName, config.TargetUnit, planAbilityHardcasts, uuid)
	if filter == nil {
		return nil
	}
	if filter.empty() {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Boss ability %q has no Spell Hit or Boss Pulse or Area Hit with a cast time, so its cast time remaining is always 0", config.AbilityName)
		return rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, uuid)
	}
	return &APLValueBossAbilityCastTimeRemaining{name: config.AbilityName, filter: filter}
}
func (value *APLValueBossAbilityCastTimeRemaining) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueBossAbilityCastTimeRemaining) GetDuration(sim *Simulation) time.Duration {
	soonest := NeverExpires
	for _, ability := range value.filter.read() {
		hc := &ability.spell.Unit.Hardcast
		if left := hc.timeLeft(sim.CurrentTime); left > 0 && hc.ActionID.SameAction(ability.spell.ActionID) {
			soonest = min(soonest, left)
		}
	}
	if soonest == NeverExpires {
		return 0
	}
	return soonest
}
func (value *APLValueBossAbilityCastTimeRemaining) String() string {
	return fmt.Sprintf("Boss Ability Cast Time Remaining(%s)", value.name)
}

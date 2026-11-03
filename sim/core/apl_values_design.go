package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

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

// The boss abilities a name-based value reads: those of the name on any unit with no target set, else those whose
// target is that enemy target. A pulse or an area hit comes from no target, so a set target leaves it out. A fixed
// target filters when the value is built, a current, previous or next target at each read (onTarget).
type designAbilityFilter struct {
	abilities []*designBossAbility
	target    UnitReference // The current, previous or next target each read filters on; Type Unknown for none.
}

// The filter of the name and target, or nil, with a warning, for a target that is not an enemy target. No name
// names none of them.
func (rot *APLRotation) designAbilityFilter(name string, targetRef *proto.UnitReference, uuid *proto.UUID) *designAbilityFilter {
	filter := &designAbilityFilter{}
	targeted := targetRef != nil && targetRef.Type != proto.UnitReference_Unknown
	var fixed *Target
	if targeted {
		unit, ok := rot.getDesignTargetUnit(targetRef, uuid)
		if !ok {
			return nil
		}
		if unit.fixedUnit != nil {
			fixed = enemyTargetOf(unit.fixedUnit)
		} else {
			filter.target = unit
		}
	}
	if design := rot.unit.Env.Encounter.design; design != nil && name != "" {
		for _, ability := range design.bossAbilities {
			if ability.config.Name == name && !(targeted && ability.isArea()) && (fixed == nil || ability.target == fixed) {
				filter.abilities = append(filter.abilities, ability)
			}
		}
	}
	return filter
}

// The filter of the name and target, or nil, with a warning, when no ability matches.
func (rot *APLRotation) designBossAbilities(name string, targetRef *proto.UnitReference, uuid *proto.UUID) *designAbilityFilter {
	filter := rot.designAbilityFilter(name, targetRef, uuid)
	if filter == nil {
		return nil
	}
	if len(filter.abilities) == 0 {
		if filter.target.Type != proto.UnitReference_Unknown {
			rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No boss ability named %q on a target", name)
		} else if targetRef != nil && targetRef.Type != proto.UnitReference_Unknown {
			rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No boss ability named %q on %s", name, rot.GetTargetUnit(targetRef).Get().Label)
		} else {
			rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "No boss ability named %q", name)
		}
		return nil
	}
	return filter
}

// Whether the ability is on the filter's current, previous or next target, or the filter reads every unit.
func (filter *designAbilityFilter) onTarget(ability *designBossAbility) bool {
	if filter.target.Type == proto.UnitReference_Unknown {
		return true
	}
	target := enemyTargetOf(filter.target.Get())
	return target != nil && ability.target == target
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

// Folds to a constant when the rotation is built, a current, previous or next target resolved then. A name the
// design lacks reads false with no warning, since absence is what the value tests.
func (rot *APLRotation) newValueBossAbilityIsKnown(config *proto.APLValueBossAbilityIsKnown, uuid *proto.UUID) APLValue {
	filter := rot.designAbilityFilter(config.AbilityName, config.TargetUnit, uuid)
	known := filter != nil && slices.ContainsFunc(filter.abilities, filter.onTarget)
	return rot.newValueConst(&proto.APLValueConst{Val: Ternary(known, "true", "false")}, uuid)
}

type APLValueBossAbilityIsActive struct {
	DefaultAPLValueImpl
	name   string
	filter *designAbilityFilter
}

func (rot *APLRotation) newValueBossAbilityIsActive(config *proto.APLValueBossAbilityIsActive, uuid *proto.UUID) APLValue {
	filter := rot.designBossAbilities(config.AbilityName, config.TargetUnit, uuid)
	if filter == nil {
		return nil
	}
	return &APLValueBossAbilityIsActive{name: config.AbilityName, filter: filter}
}
func (value *APLValueBossAbilityIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossAbilityIsActive) GetBool(_ *Simulation) bool {
	for _, ability := range value.filter.abilities {
		if ability.open > 0 && value.filter.onTarget(ability) {
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
	filter *designAbilityFilter
}

func (rot *APLRotation) newValueTimeUntilBossAbility(config *proto.APLValueTimeUntilBossAbility, uuid *proto.UUID) APLValue {
	filter := rot.designBossAbilities(config.AbilityName, config.TargetUnit, uuid)
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
	for _, ability := range value.filter.abilities {
		if value.filter.onTarget(ability) {
			next = min(next, ability.nextAt(sim))
		}
	}
	return timeUntil(sim, next)
}
func (value *APLValueTimeUntilBossAbility) String() string {
	return fmt.Sprintf("Time Until Boss Ability(%s)", value.name)
}

type APLValueBossAbilityCastTimeRemaining struct {
	DefaultAPLValueImpl
	name   string
	filter *designAbilityFilter
}

// Reads the spell hits of the filter. A name of abilities that never cast, the other kinds or a spell hit construction
// skipped, reads a constant 0ms with a warning, so a line on it folds instead of losing its condition.
func (rot *APLRotation) newValueBossAbilityCastTimeRemaining(config *proto.APLValueBossAbilityCastTimeRemaining, uuid *proto.UUID) APLValue {
	filter := rot.designBossAbilities(config.AbilityName, config.TargetUnit, uuid)
	if filter == nil {
		return nil
	}
	filter.abilities = slices.DeleteFunc(filter.abilities, func(ability *designBossAbility) bool {
		return ability.config.Kind != proto.DesignBossAbility_SpellHit || ability.spell == nil
	})
	if len(filter.abilities) == 0 {
		rot.ValidationMessageByUUID(uuid, proto.LogLevel_Warning, "Boss ability %q has no Spell Hit that casts, so its cast time remaining is always 0", config.AbilityName)
		return rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, uuid)
	}
	return &APLValueBossAbilityCastTimeRemaining{name: config.AbilityName, filter: filter}
}
func (value *APLValueBossAbilityCastTimeRemaining) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueBossAbilityCastTimeRemaining) GetDuration(sim *Simulation) time.Duration {
	soonest := NeverExpires
	for _, ability := range value.filter.abilities {
		hc := &ability.spell.Unit.Hardcast
		if hc.Expires > sim.CurrentTime && hc.ActionID.SameAction(ability.spell.ActionID) && value.filter.onTarget(ability) {
			soonest = min(soonest, hc.Expires-sim.CurrentTime)
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

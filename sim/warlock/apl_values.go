package warlock

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func (warlock *Warlock) NewAPLValue(rot *core.APLRotation, config *proto.APLValue) core.APLValue {
	switch config.Value.(type) {
	case *proto.APLValue_WarlockIsAssignedCurse:
		return warlock.newValueWarlockIsAssignedCurse(rot, config.GetWarlockIsAssignedCurse())
	case *proto.APLValue_WarlockActivePet:
		return warlock.newValueWarlockActivePet(rot, config.GetWarlockActivePet())
	case *proto.APLValue_WarlockAssignedCurseIsActive:
		return warlock.newValueWarlockAssignedCurseIsActive(rot, config.GetWarlockAssignedCurseIsActive())
	case *proto.APLValue_WarlockBaneIsActiveOnAllTargets:
		return &APLValueWarlockBaneIsActiveOnAllTargets{warlock: warlock}
	case *proto.APLValue_WarlockAssignedCurseIsActiveOnAllTargets:
		return &APLValueWarlockAssignedCurseIsActiveOnAllTargets{warlock: warlock, spell: warlock.GetAssignedCurse()}
	default:
		return nil
	}
}

type APLValueWarlockBaneIsActiveOnAllTargets struct {
	core.DefaultAPLValueImpl
	warlock *Warlock
}

func (x *APLValueWarlockBaneIsActiveOnAllTargets) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockBaneIsActiveOnAllTargets) GetBool(_ *core.Simulation) bool {
	return x.warlock.onAllTargets(x.warlock.Banes)
}

func (x *APLValueWarlockBaneIsActiveOnAllTargets) String() string {
	return "Bane Is Active On All Targets()"
}

type APLValueWarlockAssignedCurseIsActiveOnAllTargets struct {
	core.DefaultAPLValueImpl
	warlock *Warlock
	spell   *core.Spell
}

func (x *APLValueWarlockAssignedCurseIsActiveOnAllTargets) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockAssignedCurseIsActiveOnAllTargets) GetBool(_ *core.Simulation) bool {
	return x.spell != nil && x.warlock.onAllTargets([]*core.Spell{x.spell})
}

func (x *APLValueWarlockAssignedCurseIsActiveOnAllTargets) String() string {
	if x.spell == nil {
		return "Assigned Curse Is Active On All Targets (none)"
	}
	return fmt.Sprintf("Assigned Curse Is Active On All Targets (%s)", x.spell.ActionID)
}

// Whether every target that is up holds one of the spells, as its dot or its debuff. Banes and curses are
// instant, so none is ever the running hardcast, and each reads as it is.
func (warlock *Warlock) onAllTargets(spells []*core.Spell) bool {
	for _, target := range warlock.Env.Encounter.AllTargetUnits {
		if target.IsEnabled() && !anyActiveOn(spells, target) {
			return false
		}
	}
	return true
}

func anyActiveOn(spells []*core.Spell, target *core.Unit) bool {
	for _, spell := range spells {
		if dot := spell.Dot(target); dot != nil && dot.IsActive() {
			return true
		}
		for _, auras := range spell.RelatedAuraArrays {
			if aura := auras.Get(target); aura != nil && aura.IsActive() {
				return true
			}
		}
	}
	return false
}

type APLValueWarlockAssignedCurseIsActive struct {
	core.DefaultAPLValueImpl
	target core.UnitReference
	spell  *core.Spell
}

func (warlock *Warlock) newValueWarlockAssignedCurseIsActive(rot *core.APLRotation, config *proto.APLValueWarlockAssignedCurseIsActive) core.APLValue {
	target := rot.GetTargetUnit(config.TargetUnit)

	if target.Get() == nil {
		return nil
	}
	return &APLValueWarlockAssignedCurseIsActive{
		target: target,
		spell:  warlock.GetAssignedCurse(),
	}
}

func (x *APLValueWarlockAssignedCurseIsActive) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockAssignedCurseIsActive) GetBool(sim *core.Simulation) bool {
	if x.spell == nil {
		return false
	}
	target := x.target.Get()

	// This warlock's own copy of the curse. The warlock cast it, so it counts as up the moment it lands, with
	// no reaction time.
	for _, auras := range x.spell.RelatedAuraArrays {
		if aura := auras.Get(target); aura.IsActive() {
			return true
		}
	}
	return false
}

func (x *APLValueWarlockAssignedCurseIsActive) String() string {
	if x.spell == nil {
		return "Is Assigned Curse Active (none)"
	}
	return fmt.Sprintf("Is Assigned Curse Active (%s)", x.spell.ActionID)
}

type APLValueWarlockIsAssignedCurse struct {
	core.DefaultAPLValueImpl
	warlock         *Warlock
	isAssignedCurse bool
}

func (x *APLValueWarlockIsAssignedCurse) GetSpellFromAction(sim *core.Simulation) *core.Spell {
	return x.warlock.GetAssignedCurse()
}

func (warlock *Warlock) newValueWarlockIsAssignedCurse(_ *core.APLRotation, config *proto.APLValueWarlockIsAssignedCurse) core.APLValue {
	isAssignedCurse := config.CurseType == warlock.Options.CurseOptions

	return &APLValueWarlockIsAssignedCurse{
		warlock:         warlock,
		isAssignedCurse: isAssignedCurse,
	}
}

func (x *APLValueWarlockIsAssignedCurse) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockIsAssignedCurse) GetBool(sim *core.Simulation) bool {
	return x.isAssignedCurse
}

func (x *APLValueWarlockIsAssignedCurse) String() string {
	return fmt.Sprintf("Is Assigned Curse (%t)", x.isAssignedCurse)
}

type APLValueWarlockActivePet struct {
	core.DefaultAPLValueImpl
	warlock *Warlock
	summon  proto.WarlockOptions_Summon
}

func (warlock *Warlock) newValueWarlockActivePet(_ *core.APLRotation, config *proto.APLValueWarlockActivePet) core.APLValue {
	return &APLValueWarlockActivePet{
		warlock: warlock,
		summon:  config.Summon,
	}
}

func (x *APLValueWarlockActivePet) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockActivePet) GetBool(_ *core.Simulation) bool {
	pet := x.warlock.ActivePet
	if pet == nil || !pet.IsEnabled() {
		return x.summon == proto.WarlockOptions_NoSummon
	}
	return pet.summon == x.summon
}

func (x *APLValueWarlockActivePet) String() string {
	return fmt.Sprintf("Active Pet Is (%s)", x.summon)
}

func (warlock *Warlock) NewAPLAction(rot *core.APLRotation, config *proto.APLAction) core.APLActionImpl {
	switch config.Action.(type) {
	case *proto.APLAction_CastWarlockAssignedCurse:
		return warlock.newActionWarlockAssignedCurseAction(rot, config.GetCastWarlockAssignedCurse())
	default:
		return nil
	}
}

type APLActionCastWarlockAssignedCurse struct {
	warlock *Warlock
	target  core.UnitReference
}

func (x *APLActionCastWarlockAssignedCurse) GetInnerActions() []*core.APLAction { return nil }
func (x *APLActionCastWarlockAssignedCurse) GetAPLValues() []core.APLValue      { return nil }
func (x *APLActionCastWarlockAssignedCurse) Finalize(*core.APLRotation)         {}
func (x *APLActionCastWarlockAssignedCurse) PostFinalize(*core.APLRotation)     {}
func (x *APLActionCastWarlockAssignedCurse) GetNextAction(*core.Simulation) *core.APLAction {
	return nil
}
func (x *APLActionCastWarlockAssignedCurse) ReResolveVariableRefs(*core.APLRotation, map[string]*proto.APLValue) {
}

func (x *APLActionCastWarlockAssignedCurse) GetSpellFromAction(sim *core.Simulation) *core.Spell {
	return x.warlock.GetAssignedCurse()
}

func (warlock *Warlock) newActionWarlockAssignedCurseAction(rot *core.APLRotation, config *proto.APLActionCastWarlockAssignedCurse) core.APLActionImpl {
	target := rot.GetTargetUnit(config.Target)
	if target.Get() == nil {
		return nil
	}
	return &APLActionCastWarlockAssignedCurse{
		warlock: warlock,
		target:  target,
	}
}

func (x *APLActionCastWarlockAssignedCurse) Execute(sim *core.Simulation) {
	x.GetSpellFromAction(sim).CastOrQueue(sim, x.target.Get())
}

func (x *APLActionCastWarlockAssignedCurse) IsReady(sim *core.Simulation) bool {
	spell := x.GetSpellFromAction(sim)
	return spell != nil && spell.CanCastOrQueue(sim, x.target.Get())
}

func (x *APLActionCastWarlockAssignedCurse) Reset(*core.Simulation) {}

func (x *APLActionCastWarlockAssignedCurse) String() string {
	curse := x.warlock.GetAssignedCurse()
	if curse == nil {
		return "Cast Assigned Curse(none)"
	}
	return fmt.Sprintf("Cast Assigned Curse(%s)", curse.ActionID)
}

// Nil under NoCurse. Elements is the default: any value it does not name falls back to it.
func (warlock *Warlock) GetAssignedCurse() *core.Spell {
	switch warlock.Options.CurseOptions {
	case proto.WarlockOptions_NoCurse:
		return nil
	case proto.WarlockOptions_Recklessness:
		return warlock.CurseOfRecklessness
	case proto.WarlockOptions_Weakness:
		return warlock.CurseOfWeakness
	}
	return warlock.CurseOfElements
}

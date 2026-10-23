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
		return &APLValueWarlockActiveOnAllTargets{warlock: warlock, spells: warlock.Banes, label: "Bane Is Active On All Targets"}
	case *proto.APLValue_WarlockAssignedCurseIsActiveOnAllTargets:
		curse := warlock.GetAssignedCurse()
		if curse == nil {
			rot.ValidationMessage(proto.LogLevel_Warning, "Assigned Curse Is Active On All Targets: no curse is assigned")
			return nil
		}
		return &APLValueWarlockActiveOnAllTargets{
			warlock: warlock,
			spells:  []*core.Spell{curse},
			label:   fmt.Sprintf("Assigned Curse Is Active On All Targets (%s)", curse.ActionID),
		}
	default:
		return nil
	}
}

// True while every target that is up holds one of the spells.
type APLValueWarlockActiveOnAllTargets struct {
	core.DefaultAPLValueImpl
	warlock *Warlock
	spells  []*core.Spell
	label   string
}

func (x *APLValueWarlockActiveOnAllTargets) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (x *APLValueWarlockActiveOnAllTargets) GetBool(_ *core.Simulation) bool {
	for _, target := range x.warlock.Env.Encounter.AllTargetUnits {
		if target.IsEnabled() && !anyActiveOn(x.spells, target) {
			return false
		}
	}
	return true
}

func (x *APLValueWarlockActiveOnAllTargets) String() string {
	return x.label
}

func anyActiveOn(spells []*core.Spell, target *core.Unit) bool {
	for _, spell := range spells {
		if activeOn(spell, target) {
			return true
		}
	}
	return false
}

// Whether the spell's dot or one of its debuffs is up on the target, read as it is: Banes and curses are
// instant, so none is ever the running hardcast that the rotation reads as landed.
func activeOn(spell *core.Spell, target *core.Unit) bool {
	if dot := spell.Dot(target); dot != nil && dot.IsActive() {
		return true
	}
	for _, auras := range spell.RelatedAuraArrays {
		if aura := auras.Get(target); aura != nil && aura.IsActive() {
			return true
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
	// This warlock's own copy of the curse. The warlock cast it, so it counts as up the moment it lands, with
	// no reaction time.
	return x.spell != nil && activeOn(x.spell, x.target.Get())
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

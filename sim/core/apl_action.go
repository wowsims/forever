package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

type APLAction struct {
	condition APLValue
	impl      APLActionImpl
	spell     *Spell // The spell a cast action casts or queues, whose queue check runs before the condition.
}

func (action *APLAction) Finalize(rot *APLRotation) {
	// Defensive check to prevent nil pointer dereference
	if action == nil || action.impl == nil {
		return
	}

	action.impl.Finalize(rot)
	for _, value := range action.GetAllAPLValues() {
		value.Finalize(rot)
	}
}

func (action *APLAction) IsReady(sim *Simulation) bool {
	return (action.condition == nil || action.condition.GetBool(sim)) && action.impl.IsReady(sim)
}

// Whether a queued spell holds back the action's cast, which is then not ready whatever its condition says.
func (action *APLAction) heldByQueuedSpell(sim *Simulation) bool {
	return action.spell != nil && action.spell.heldByQueuedSpell(sim)
}

func (action *APLAction) Execute(sim *Simulation) {
	action.impl.Execute(sim)
}

// Returns this Action, along with all inner Actions.
func (action *APLAction) GetAllActions() []*APLAction {
	// Defensive check to prevent nil pointer dereference
	if action == nil || action.impl == nil {
		return nil
	}

	actions := action.impl.GetInnerActions()
	actions = append(actions, action)
	return actions
}

// Returns all APLValues used by this action and all of its inner Actions.
func (action *APLAction) GetAllAPLValues() []APLValue {
	var values []APLValue
	for _, a := range action.GetAllActions() {
		// Defensive check to prevent nil pointer dereference
		if a == nil || a.impl == nil {
			continue
		}

		unprocessed := a.impl.GetAPLValues()
		if a.condition != nil {
			unprocessed = append(unprocessed, a.condition)
		}

		for len(unprocessed) > 0 {
			next := unprocessed[len(unprocessed)-1]
			unprocessed = unprocessed[:len(unprocessed)-1]
			if next == nil {
				continue
			}
			values = append(values, next)
			unprocessed = append(unprocessed, next.GetInnerValues()...)
		}
	}
	return FilterSlice(values, func(val APLValue) bool { return val != nil })
}

func (action *APLAction) GetAllSpells() []*Spell {
	var spells []*Spell
	for _, a := range action.GetAllActions() {
		if impl, ok := a.impl.(*APLActionCastSpell); ok {
			spells = append(spells, impl.spell)
		} else if impl, ok := a.impl.(*APLActionMultidot); ok {
			spells = append(spells, impl.spell)
		} else if impl, ok := a.impl.(*APLActionStrictMultidot); ok {
			spells = append(spells, impl.spell)
		} else if impl, ok := a.impl.(*APLActionMultishield); ok {
			spells = append(spells, impl.spell)
		}
	}
	return spells
}

func (action *APLAction) String() string {
	if action.condition == nil {
		return fmt.Sprintf("ACTION = %s", action.impl)
	} else {
		return fmt.Sprintf("IF = %s\nACTION = %s", action.condition, action.impl)
	}
}

type APLActionImpl interface {
	// Returns all inner APL Actions.
	GetInnerActions() []*APLAction

	// Returns all APLValues used by this Action (but not by inner Actions).
	GetAPLValues() []APLValue

	// Performs optional post-processing.
	Finalize(*APLRotation)
	PostFinalize(*APLRotation)

	// Invoked before each sim iteration.
	Reset(*Simulation)

	// Whether this action is available to be used right now.
	IsReady(*Simulation) bool

	// Performs the action.
	Execute(*Simulation)

	// Called only while this action is controlling the rotation.
	GetNextAction(sim *Simulation) *APLAction

	// Swaps each value this action holds in its own fields for what replace gives for it.
	ReplaceAPLValues(replace func(APLValue) APLValue)

	// Pretty-print string for debugging.
	String() string
}

// Provides empty implementations for the action impl interface functions.
type defaultAPLActionImpl struct {
}

func (impl defaultAPLActionImpl) GetInnerActions() []*APLAction            { return nil }
func (impl defaultAPLActionImpl) GetAPLValues() []APLValue                 { return nil }
func (impl defaultAPLActionImpl) Finalize(*APLRotation)                    {}
func (impl defaultAPLActionImpl) PostFinalize(*APLRotation)                {}
func (impl defaultAPLActionImpl) Reset(*Simulation)                        {}
func (impl defaultAPLActionImpl) GetNextAction(*Simulation) *APLAction     { return nil }
func (impl defaultAPLActionImpl) ReplaceAPLValues(func(APLValue) APLValue) {}

func (rot *APLRotation) newAPLAction(config *proto.APLAction) *APLAction {
	if config == nil {
		return nil
	}

	impl := rot.newAPLActionImpl(config)
	if impl == nil {
		// Don't create an action with a nil implementation
		return nil
	}

	condition := rot.coerceTo(rot.newAPLValue(config.Condition), proto.APLValueType_ValueTypeBool)
	if config.Condition != nil && config.Condition.Value != nil && condition == nil {
		rot.ValidationMessage(proto.LogLevel_Warning, "Condition could not be resolved and will be treated as always true. Add an 'Aura Known' check if this action should be skipped when the aura is absent.")
	}

	// If condition is const false, this action can never fire — prune it.
	// Track the impl so its spells are still removed from MCD auto-casting.
	if condition != nil {
		if constVal, ok := condition.(*APLValueConst); ok && constVal.valType == proto.APLValueType_ValueTypeBool {
			if !constVal.boolVal {
				rot.prunedActions = append(rot.prunedActions, impl)
				return nil
			}
			// Const true condition is equivalent to no condition.
			condition = nil
		}
	}

	return newAPLActionEntry(condition, impl)
}

// Builds a rotation entry and hands it to the actions that read their own entry's condition.
func newAPLActionEntry(condition APLValue, impl APLActionImpl) *APLAction {
	action := &APLAction{
		condition: condition,
		impl:      impl,
		spell:     castSpellOf(impl),
	}
	switch impl := impl.(type) {
	case *APLActionStrictMultidot:
		impl.entry = action
	case *APLActionGroupReference:
		impl.entry = action
	case *APLActionMove:
		impl.entry = action
	}
	return action
}

func removeFromMajorCooldowns(action APLActionImpl, character *Character) {
	if spell := castSpellOf(action); spell != nil {
		character.removeInitialMajorCooldown(spell.ActionID)
	}
}

func (rot *APLRotation) newAPLActionImpl(config *proto.APLAction) APLActionImpl {
	if config == nil {
		fmt.Println("newAPLActionImpl called with nil config")
		return nil
	}

	customAction := rot.unit.Env.GetAgentFromUnit(rot.unit).NewAPLAction(rot, config)
	if customAction != nil {
		return customAction
	}

	switch config.Action.(type) {
	// Casting
	case *proto.APLAction_CastSpell:
		return rot.newActionCastSpell(config.GetCastSpell())
	case *proto.APLAction_CastFriendlySpell:
		return rot.newActionCastFriendlySpell(config.GetCastFriendlySpell())
	case *proto.APLAction_ChannelSpell:
		return rot.newActionChannelSpell(config.GetChannelSpell())
	case *proto.APLAction_Multidot:
		return rot.newActionMultidot(config.GetMultidot())
	case *proto.APLAction_StrictMultidot:
		return rot.newActionStrictMultidot(config.GetStrictMultidot())
	case *proto.APLAction_Multishield:
		return rot.newActionMultishield(config.GetMultishield())
	case *proto.APLAction_CastAllStatBuffCooldowns:
		return rot.newActionCastAllStatBuffCooldowns(config.GetCastAllStatBuffCooldowns())
	case *proto.APLAction_AutocastOtherCooldowns:
		return rot.newActionAutocastOtherCooldowns(config.GetAutocastOtherCooldowns())

	// Timing
	case *proto.APLAction_Wait:
		return rot.newActionWait(config.GetWait())
	case *proto.APLAction_WaitUntil:
		return rot.newActionWaitUntil(config.GetWaitUntil())
	case *proto.APLAction_Schedule:
		return rot.newActionSchedule(config.GetSchedule())

	// Sequences
	case *proto.APLAction_Sequence:
		return rot.newActionSequence(config.GetSequence())
	case *proto.APLAction_ResetSequence:
		return rot.newActionResetSequence(config.GetResetSequence())
	case *proto.APLAction_StrictSequence:
		return rot.newActionStrictSequence(config.GetStrictSequence())

	// Misc
	case *proto.APLAction_ChangeTarget:
		return rot.newActionChangeTarget(config.GetChangeTarget())
	case *proto.APLAction_ActivateAura:
		return rot.newActionActivateAura(config.GetActivateAura())
	case *proto.APLAction_ActivateAuraWithStacks:
		return rot.newActionActivateAuraWithStacks(config.GetActivateAuraWithStacks())
	case *proto.APLAction_ActivateAllStatBuffProcAuras:
		return rot.newActionActivateAllStatBuffProcAuras(config.GetActivateAllStatBuffProcAuras())
	case *proto.APLAction_CancelAura:
		return rot.newActionCancelAura(config.GetCancelAura())
	case *proto.APLAction_TriggerIcd:
		return rot.newActionTriggerICD(config.GetTriggerIcd())
	case *proto.APLAction_ItemSwap:
		return rot.newActionItemSwap(config.GetItemSwap())
	case *proto.APLAction_Move:
		return rot.newActionMove(config.GetMove())
	case *proto.APLAction_MoveDuration:
		return rot.newActionMoveDuration(config.GetMoveDuration())
	case *proto.APLAction_PetAttack:
		return rot.newActionPetAttack(config.GetPetAttack())
	case *proto.APLAction_StartAttack:
		return rot.newActionStartAttack(config.GetStartAttack())
	case *proto.APLAction_CustomRotation:
		return rot.newActionCustomRotation(config.GetCustomRotation())
	case *proto.APLAction_GroupReference:
		return rot.newActionGroupReference(config.GetGroupReference())

	default:
		return nil
	}
}

// Default implementation of Agent.NewAPLAction so each spec doesn't need this boilerplate.
func (unit *Unit) NewAPLAction(rot *APLRotation, config *proto.APLAction) APLActionImpl {
	return nil
}

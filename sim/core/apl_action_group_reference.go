package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
)

type APLActionGroupReference struct {
	defaultAPLActionImpl
	groupName string
	variables map[string]*proto.APLValue
	group     *APLGroup
	matched   bool
	rot       *APLRotation
	entry     *APLAction // The entry holding this reference, whose condition a Strict Multidot chain in the group re-checks.
}

func (rot *APLRotation) newActionGroupReference(config *proto.APLActionGroupReference) APLActionImpl {
	if config == nil {
		return nil
	}

	// Don't create group references with empty names
	if config.GroupName == "" {
		fmt.Println("Skipping group reference creation: GroupName is empty")
		return nil
	}

	vars := make(map[string]*proto.APLValue)
	for _, v := range config.Variables {
		vars[v.Name] = v.Value
	}

	return &APLActionGroupReference{
		groupName: config.GroupName,
		variables: vars,
		rot:       rot,
	}
}

func (action *APLActionGroupReference) GetInnerActions() []*APLAction {
	if action.group == nil || len(action.group.actions) == 0 {
		return nil
	}

	var actions []*APLAction
	for _, groupAction := range action.group.actions {
		if groupAction == nil {
			continue
		}
		actions = append(actions, groupAction.GetAllActions()...)
	}
	return actions
}

func (action *APLActionGroupReference) GetAPLValues() []APLValue {
	if action.group == nil {
		return nil
	}

	var values []APLValue
	for _, groupAction := range action.group.actions {
		// Defensive check to prevent nil pointer dereference
		if groupAction == nil {
			continue
		}
		values = append(values, groupAction.GetAllAPLValues()...)
	}
	return values
}

func (action *APLActionGroupReference) Finalize(rot *APLRotation) {

	// Skip finalization if groupName is empty
	if action.groupName == "" {
		fmt.Println("Skipping finalization: groupName is empty")
		return
	}

	// Find the referenced group
	for _, group := range rot.groups {
		if (group.name == action.groupName) && ((group.referencedBy == nil) || (group.referencedBy == action)) {
			action.group = group
			group.referencedBy = action
			break
		}
	}

	if action.group == nil {
		rot.ValidationMessage(proto.LogLevel_Error, "Group reference '%s' not found", action.groupName)
		return
	}

	// Every action of the group, those inside its sequences included, whose conditions and values may hold
	// placeholders.
	var actions []*APLAction
	for _, groupAction := range action.group.actions {
		actions = append(actions, groupAction.GetAllActions()...)
	}

	// Scan for VariablePlaceholder values in the group
	placeholders := map[string]struct{}{}
	for _, groupAction := range actions {
		// Check condition for placeholders
		if groupAction.condition != nil {
			action.scanForPlaceholders(groupAction.condition, placeholders)
		}
		// Check action implementation for placeholders
		action.scanActionForPlaceholders(groupAction.impl, placeholders)
	}

	// Check that all placeholders are set and provide detailed error messages
	hasUnfilledPlaceholders := false
	for name := range placeholders {
		if _, ok := action.variables[name]; !ok {
			rot.ValidationMessage(proto.LogLevel_Error, "Group '%s' requires variable placeholder '%s' to be filled with a variable", action.groupName, name)
			hasUnfilledPlaceholders = true
		}
	}

	// If there are unfilled placeholders, don't continue with finalization
	if hasUnfilledPlaceholders {
		return
	}

	// Replace placeholder values with provided variables
	for _, groupAction := range actions {
		groupAction.condition = action.replacePlaceholders(groupAction.condition, action.variables, rot)
		groupAction.impl.ReplaceAPLValues(func(value APLValue) APLValue {
			return action.replacePlaceholders(value, action.variables, rot)
		})
	}

	// Add all provided variables to the group's variables map
	// This allows normal variable references to resolve to the provided values
	for k, v := range action.variables {
		action.group.variables[k] = v
	}

	// Re-resolve variable references in group actions with the updated group variables
	for _, groupAction := range action.group.actions {
		// Re-resolve the condition if it contains variable references
		if groupAction.condition != nil {
			groupAction.condition = rot.reResolveVariableRefs(groupAction.condition, action.group.variables)
		}
	}

	// Finalize all actions in the group
	for _, groupAction := range action.group.actions {
		groupAction.Finalize(rot)
	}
}

func (action *APLActionGroupReference) Reset(sim *Simulation) {
	// No need to reset inner actions manually - the main APL rotation handles that
}

func (action *APLActionGroupReference) IsReady(sim *Simulation) bool {
	if action.group == nil {
		return false
	}

	// Check if any action in the group is ready
	held := action.rot.unit.queuedSpellHoldsCasts(sim)
	for _, groupAction := range action.group.actions {
		if !(held && groupAction.heldByQueuedSpell(sim)) && groupAction.IsReady(sim) {
			return true
		}
	}
	return false
}

func (action *APLActionGroupReference) Execute(sim *Simulation) {
	if action.group == nil {
		return
	}
	condition := action.entry.condition
	if condition != nil {
		action.rot.groupConditions = append(action.rot.groupConditions, condition)
	}

	// Execute the first ready action in the group
	for _, groupAction := range action.group.actions {
		if groupAction.IsReady(sim) {
			groupAction.Execute(sim)
			break
		}
	}

	if condition != nil {
		action.rot.groupConditions = action.rot.groupConditions[:len(action.rot.groupConditions)-1]
	}
}

func (action *APLActionGroupReference) String() string {
	return fmt.Sprintf("Group Reference: %s", action.groupName)
}

// Helper methods for placeholder handling

func (action *APLActionGroupReference) scanForPlaceholders(value APLValue, placeholders map[string]struct{}) {
	if value == nil {
		return
	}

	// Check if this value is a placeholder
	if placeholder, ok := value.(*APLValueVariablePlaceholder); ok {
		placeholders[placeholder.name] = struct{}{}
		return
	}

	// Recursively scan inner values
	for _, innerValue := range value.GetInnerValues() {
		action.scanForPlaceholders(innerValue, placeholders)
	}
}

func (action *APLActionGroupReference) scanActionForPlaceholders(actionImpl APLActionImpl, placeholders map[string]struct{}) {
	if actionImpl == nil {
		return
	}

	// Get all APL values from the action and scan them
	for _, value := range actionImpl.GetAPLValues() {
		action.scanForPlaceholders(value, placeholders)
	}
}

func (action *APLActionGroupReference) replacePlaceholders(value APLValue, variables map[string]*proto.APLValue, rot *APLRotation) APLValue {
	if value == nil {
		return nil
	}

	// Check if this value is a placeholder
	if placeholder, ok := value.(*APLValueVariablePlaceholder); ok {
		if replacement, ok := variables[placeholder.name]; ok {
			// Create a new value from the replacement. One that is empty or does not resolve keeps the
			// placeholder, which reads zero, so what holds it still has a value.
			if filled := rot.newAPLValue(replacement); filled != nil {
				return filled
			}
			rot.ValidationMessage(proto.LogLevel_Warning, "Group '%s' variable '%s' could not be read, its placeholder reads as zero", action.groupName, placeholder.name)
		}
		return value // Keep original if no replacement found
	}

	// For composite values, recursively replace placeholders in inner values
	innerValues := value.GetInnerValues()
	if len(innerValues) > 0 {
		// Check if any inner values were replaced
		anyReplaced := false
		newInnerValues := make([]APLValue, len(innerValues))
		for i, innerValue := range innerValues {
			newInnerValues[i] = action.replacePlaceholders(innerValue, variables, rot)
			if newInnerValues[i] != innerValue {
				anyReplaced = true
			}
		}

		// If any inner values were replaced, we need to create a new value object. A Math one is checked as its
		// constructor checks it, and one that fails keeps the placeholders, which read zero.
		if anyReplaced {
			// Handle specific value types that need special treatment
			switch v := value.(type) {
			case *APLValueCompare:
				// Re-coerce the types after replacement to ensure compatibility
				lhs, rhs := rot.coerceToSameType(newInnerValues[0], newInnerValues[1])
				return &APLValueCompare{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					op:                  v.op,
					lhs:                 lhs,
					rhs:                 rhs,
				}
			case *APLValueMath:
				// Re-coerce the types after replacement to ensure compatibility
				lhs, rhs := newInnerValues[0], newInnerValues[1]
				if v.op == proto.APLValueMath_OpAdd || v.op == proto.APLValueMath_OpSub {
					lhs, rhs = rot.coerceToSameType(lhs, rhs)
				}
				if problem := mathTypeProblem(v.op, lhs, rhs); problem != "" {
					rot.ValidationMessage(proto.LogLevel_Warning, "%s", problem)
					return value
				}
				return &APLValueMath{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					op:                  v.op,
					lhs:                 lhs,
					rhs:                 rhs,
				}
			case *APLValueAnd:
				return &APLValueAnd{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					vals:                newInnerValues,
				}
			case *APLValueOr:
				return &APLValueOr{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					vals:                newInnerValues,
				}
			case *APLValueNot:
				return &APLValueNot{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					val:                 newInnerValues[0],
				}
			case *APLValueMax:
				return &APLValueMax{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					vals:                rot.coerceAllToSameType(newInnerValues),
				}
			case *APLValueMin:
				return &APLValueMin{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					vals:                rot.coerceAllToSameType(newInnerValues),
				}
			case *APLValueCoerced:
				// A fill the wrapper cannot read as its type keeps the wrapper around the placeholder, which
				// reads zero.
				if !canCoerce(newInnerValues[0].Type(), v.valueType) {
					rot.warnCannotCoerce(newInnerValues[0].Type(), v.valueType)
					return value
				}
				return &APLValueCoerced{
					DefaultAPLValueImpl: v.DefaultAPLValueImpl,
					valueType:           v.valueType,
					inner:               newInnerValues[0],
				}
			default:
				// For other value types, we can't easily create new instances
				// so we'll return the original value
				return value
			}
		}
	}

	return value
}

type APLValueActionGroupUsed struct {
	DefaultAPLValueImpl
	name   string
	isUsed bool
}

func (rot *APLRotation) newValueActionGroupUsed(config *proto.APLValueActionGroupUsed, _ *proto.UUID) APLValue {
	isUsed := false

	return &APLValueActionGroupUsed{
		name:   config.Name,
		isUsed: isUsed,
	}
}
func (value *APLValueActionGroupUsed) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueActionGroupUsed) GetBool(sim *Simulation) bool {
	return value.isUsed
}
func (value *APLValueActionGroupUsed) String() string {
	return fmt.Sprintf("Action Group Is Used(%s)", value.name)
}
func (value *APLValueActionGroupUsed) Finalize(rotation *APLRotation) {
	for _, group := range rotation.groups {
		if group.name == value.name && (group.referencedBy != nil) {
			value.isUsed = true
			break
		}
	}
}

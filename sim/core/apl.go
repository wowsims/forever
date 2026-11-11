package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type APLRotation struct {
	unit           *Unit
	prepullActions []*APLAction
	priorityList   []*APLAction
	groups         []*APLGroup
	valueVariables []*APLValueVariable
	resetActions   []*APLAction

	canMoveCloserKnown bool // canMoveCloser's answer is set once, as it depends only on the rotation.
	canMoveCloserValue bool

	hasMoveKnown bool // hasMove's answer is set once, as it depends only on the rotation.
	hasMoveValue bool

	majorCooldownActionsKnown bool         // majorCooldownActions is set once, as it depends only on the rotation.
	majorCooldownActions      []*APLAction // The priority list's actions that can be a major cooldown going ahead of the spell queued after a hardcast.

	// Shared variable caches: all APLValueVariableRef instances for the same
	// variable name share the same cache (but each has its own expression tree).
	variableCaches map[string]*variableCache
	evalGeneration uint32

	// Values orphaned by compile-time constant folding (e.g. And/Or short-circuit).
	// These still need Finalize() called on them.
	orphanedValues []APLValue

	// Actions pruned by compile-time constant folding (const false condition).
	// Their spells still need to be removed from MCD auto-casting.
	prunedActions []APLActionImpl

	// The moves that look again at their condition and range while their walk is under way.
	recheckedMoves []*APLActionMove

	// Action currently controlling this rotation (only used for certain actions, such as StrictSequence).
	controllingActions []APLActionImpl

	// The conditions of the Group References whose group is executing; a Strict Multidot started there re-checks them.
	groupConditions []APLValue

	// Value that should evaluate to 'true' if the current channel is to be interrupted.
	// Will be nil when there is no active channel.
	interruptChannelIf APLValue

	// If true, can recast channel when interrupted.
	allowChannelRecastOnInterrupt bool

	// Used inside of actions/value to determine whether they will occur during the prepull or regular rotation.
	parsingPrepull bool

	// Used to avoid recursive APL loops.
	inLoop bool

	// Set while a sequence asks whether its step is ready or runs it: major cooldowns skip their GCD
	// wait, and a Change Target to the current target is a step that does nothing.
	inSequence bool

	// The path of the entry whose action is running, kept only while the sim logs; empty outside one.
	runningEntryPath string

	// Validation warnings that occur during proto parsing.
	// We return these back to the user for display in the UI.
	curValidations          []*proto.APLValidation
	prepullValidations      [][]*proto.APLValidation
	priorityListValidations [][]*proto.APLValidation
	groupListValidations    [][][]*proto.APLValidation
	uuidValidations         map[*proto.UUID][]*proto.APLValidation

	// Maps indices in filtered sim lists to indices in configs.
	prepullIdxMap      []int
	prepullDoAt        []time.Duration // When each of prepullActions runs.
	priorityListIdxMap []int
	groupListIdxMap    [][]int
}

// Called with every action a sim that RunSimWithAPLHook runs executes: a row, a group reference's row and then the
// row it ran from the group, and the actions inside a row. Tools that read a rotation's choices pass one, and find
// the rows among the actions with APLRotation.RowPaths.
type APLActionHook func(sim *Simulation, action *APLAction)

type APLGroup struct {
	name         string
	actions      []*APLAction
	variables    map[string]*proto.APLValue
	referencedBy *APLActionGroupReference
	configIdx    int // The group's index in the config's groups, shared by its copies.
}

type APLValueVariable struct {
	name  string
	value *proto.APLValue
}

func (rot *APLRotation) ValidationMessage(log_level proto.LogLevel, message string, vals ...any) {
	formatted_message := fmt.Sprintf(message, vals...)
	rot.curValidations = append(rot.curValidations, &proto.APLValidation{
		LogLevel:   log_level,
		Validation: formatted_message,
	})
}

func (rot *APLRotation) ValidationMessageByUUID(uuid *proto.UUID, log_level proto.LogLevel, message string, vals ...any) {
	if uuid != nil {
		formatted_message := fmt.Sprintf(message, vals...)
		rot.uuidValidations[uuid] = append(rot.uuidValidations[uuid], &proto.APLValidation{
			LogLevel:   log_level,
			Validation: formatted_message,
		})
	}
}

// Invokes the fn function, and attributes all warnings generated during its invocation
// to the provided warningsList.
func (rot *APLRotation) doAndRecordWarnings(warningsList *[]*proto.APLValidation, isPrepull bool, fn func()) {
	rot.parsingPrepull = isPrepull
	fn()
	if warningsList != nil {
		*warningsList = append(*warningsList, rot.curValidations...)
	}
	rot.curValidations = nil
	rot.parsingPrepull = false
}

func (unit *Unit) newCustomRotation() *APLRotation {
	return unit.newAPLRotation(&proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{
			{
				Action: &proto.APLAction{
					Action: &proto.APLAction_CustomRotation{},
				},
			},
		},
	})
}

// Read before the encounter starts, with a nil sim. A value that reads the encounter (the current time, a
// unit's health) panics then, and the time is unknown; the panic comes back as err.
func prepullTime(value APLValue) (doAt time.Duration, err any) {
	defer func() { err = recover() }()
	return value.GetDuration(nil), nil
}

func (unit *Unit) newAPLRotation(config *proto.APLRotation) *APLRotation {
	if config == nil {
		return nil
	}

	groupsConfig := config.Groups
	rotation := &APLRotation{
		unit:                    unit,
		variableCaches:          make(map[string]*variableCache),
		evalGeneration:          1, // Start at 1 so zero-value generation fields never match
		prepullValidations:      make([][]*proto.APLValidation, len(config.PrepullActions)),
		priorityListValidations: make([][]*proto.APLValidation, len(config.PriorityList)),
		groupListValidations:    make([][][]*proto.APLValidation, len(groupsConfig)),
		uuidValidations:         make(map[*proto.UUID][]*proto.APLValidation),
	}

	// Parse value variables FIRST, before any actions that might reference them
	for _, condVar := range config.ValueVariables {
		rotation.valueVariables = append(rotation.valueVariables, &APLValueVariable{
			name:  condVar.Name,
			value: condVar.Value,
		})
	}

	// Parse prepull actions
	for i, prepullItem := range config.PrepullActions {
		prepullIdx := i // Save to local variable for correct lambda capture behavior
		rotation.doAndRecordWarnings(&rotation.prepullValidations[prepullIdx], true, func() {
			if !prepullItem.Hide {
				doAtVal := rotation.coerceTo(rotation.newAPLValue(prepullItem.DoAtValue), proto.APLValueType_ValueTypeDuration)
				if doAtVal == nil {
					rotation.ValidationMessage(proto.LogLevel_Warning, "'Do At' needs a time, ignoring this Prepull Action")
				} else {
					doAt, err := prepullTime(doAtVal)
					if err != nil {
						rotation.ValidationMessage(proto.LogLevel_Warning, "'Do At' cannot be read before the pull (%v), ignoring this Prepull Action", err)
					} else if doAt > 0 {
						rotation.ValidationMessage(proto.LogLevel_Warning, "Invalid time for 'Do At', ignoring this Prepull Action")
					} else {
						action := rotation.newAPLAction(prepullItem.Action)
						if action != nil {
							rotation.prepullActions = append(rotation.prepullActions, action)
							rotation.prepullIdxMap = append(rotation.prepullIdxMap, prepullIdx)
							rotation.prepullDoAt = append(rotation.prepullDoAt, doAt)
							unit.RegisterPrepullAction(doAt, func(sim *Simulation) {
								// A prepull action at a cast's last instant runs before the cast's own
								// completion; the cast completes first.
								unit.completeDueHardcast(sim)

								// Warnings for prepull cast failure are detected by running a fake prepull,
								// so this action.Execute needs to record warnings.
								rotation.doAndRecordWarnings(&rotation.prepullValidations[prepullIdx], true, func() {
									if action.condition == nil || action.condition.GetBool(sim) {
										action.Execute(sim)
									}
								})
							})
						}
					}
				}
			}
		})
	}

	// Parse priority list
	for i, aplItem := range config.PriorityList {
		rotation.doAndRecordWarnings(&rotation.priorityListValidations[i], false, func() {
			if !aplItem.Hide {
				action := rotation.newAPLAction(aplItem.Action)
				if action != nil {
					rotation.priorityList = append(rotation.priorityList, action)
					rotation.priorityListIdxMap = append(rotation.priorityListIdxMap, i)
				}
			}
		})
	}

	// Parse groups
	for groupIdx := 0; groupIdx < len(groupsConfig); groupIdx++ {
		groupConfig := groupsConfig[groupIdx]
		group := &APLGroup{
			name:      groupConfig.Name,
			variables: make(map[string]*proto.APLValue),
			configIdx: slices.Index(config.Groups, groupConfig),
		}

		for _, varConfig := range groupConfig.Variables {
			group.variables[varConfig.Name] = varConfig.Value
		}

		if rotation.groupListValidations[groupIdx] == nil {
			rotation.groupListValidations[groupIdx] = make([][]*proto.APLValidation, len(groupConfig.Actions))
			rotation.groupListIdxMap = append(rotation.groupListIdxMap, []int{})
		}

		// Parse actions in the group
		for i, aplItem := range groupConfig.Actions {
			rotation.doAndRecordWarnings(&rotation.groupListValidations[groupIdx][i], false, func() {
				if !aplItem.Hide {
					// Don't pass group.variables here - placeholders should remain as placeholders
					// until the group is actually referenced with specific variable values
					action := rotation.newAPLActionWithGroupVars(aplItem.Action, nil)
					if action != nil {
						group.actions = append(group.actions, action)
						rotation.groupListIdxMap[groupIdx] = append(rotation.groupListIdxMap[groupIdx], i)
					}
				}
			})
		}

		rotation.groups = append(rotation.groups, group)

		// Each reference runs its own copy of the group: the group is duplicated for every reference past the first
		// in the prepull and priority lists, and for every reference in another group's actions. A reference inside
		// a sequence counts too.
		foundReference := false

		for _, entry := range slices.Concat(rotation.prepullActions, rotation.priorityList) {
			for _, action := range entry.GetAllActions() {
				if groupReferenceAction, ok := action.impl.(*APLActionGroupReference); ok {
					if (groupReferenceAction.groupName == group.name) && !groupReferenceAction.matched {
						if foundReference {
							groupsConfig = append(groupsConfig, groupConfig)
							rotation.groupListValidations = append(rotation.groupListValidations, nil)
						}

						foundReference = true
						groupReferenceAction.matched = true
						group.referencedBy = groupReferenceAction
					}
				}
			}
		}

		for _, entry := range group.actions {
			for _, action := range entry.GetAllActions() {
				if groupReferenceAction, ok := action.impl.(*APLActionGroupReference); ok {
					for _, groupConfig := range groupsConfig {
						if (groupReferenceAction.groupName == groupConfig.Name) && !groupReferenceAction.matched {
							groupsConfig = append(groupsConfig, groupConfig)
							rotation.groupListValidations = append(rotation.groupListValidations, nil)
							groupReferenceAction.matched = true
						}
					}
				}
			}
		}
	}

	// Finalize
	for i, action := range rotation.prepullActions {
		rotation.doAndRecordWarnings(&rotation.prepullValidations[rotation.prepullIdxMap[i]], true, func() {
			action.Finalize(rotation)
		})
	}

	for groupIdx, group := range rotation.groups {
		for actionIdx, action := range group.actions {
			rotation.doAndRecordWarnings(&rotation.groupListValidations[groupIdx][rotation.groupListIdxMap[groupIdx][actionIdx]], false, func() {
				action.Finalize(rotation)
			})
		}
	}

	for i, action := range rotation.priorityList {
		rotation.doAndRecordWarnings(&rotation.priorityListValidations[rotation.priorityListIdxMap[i]], false, func() {
			action.Finalize(rotation)
		})
	}

	// Finalize orphaned values from compile-time constant folding.
	for _, value := range rotation.orphanedValues {
		value.Finalize(rotation)
	}

	agent := unit.Env.GetAgentFromUnit(unit)
	if agent != nil {
		character := agent.GetCharacter()

		// Remove MCDs that are referenced by APL actions (including pruned ones),
		// so that the Autocast Other Cooldowns action does not include them.
		for _, action := range rotation.allAPLActions() {
			removeFromMajorCooldowns(action.impl, character)
		}
		for _, impl := range rotation.prunedActions {
			removeFromMajorCooldowns(impl, character)
		}

		rotation.warnOnPrepullSwapWithoutMain()
	}

	rotation.assignEntryPaths(config)

	return rotation
}

func (rot *APLRotation) getStats() *proto.APLStats {
	// Perform one final round of validation after post-finalize effects.
	for i, action := range rot.prepullActions {
		rot.doAndRecordWarnings(&rot.prepullValidations[rot.prepullIdxMap[i]], true, func() {
			rot.postFinalize(action)
		})
	}

	// Parse groups
	for groupIdx, group := range rot.groups {
		// Parse actions in the group
		for actionIdx, action := range group.actions {
			rot.doAndRecordWarnings(&rot.groupListValidations[groupIdx][rot.groupListIdxMap[groupIdx][actionIdx]], false, func() {
				rot.postFinalize(action)
			})
		}
	}

	for i, action := range rot.priorityList {
		rot.doAndRecordWarnings(&rot.priorityListValidations[rot.priorityListIdxMap[i]], false, func() {
			rot.postFinalize(action)
		})
	}

	uuidValidationsArr := make([]*proto.UUIDValidations, len(rot.uuidValidations))
	i := 0
	for uuid, validations := range rot.uuidValidations {
		uuidValidationsArr[i] = &proto.UUIDValidations{
			Uuid:        uuid,
			Validations: validations,
		}
		i++
	}

	return &proto.APLStats{
		PrepullActions: MapSlice(rot.prepullValidations, func(validations []*proto.APLValidation) *proto.APLActionStats {
			return &proto.APLActionStats{Validations: validations}
		}),
		PriorityList: MapSlice(rot.priorityListValidations, func(validations []*proto.APLValidation) *proto.APLActionStats {
			return &proto.APLActionStats{Validations: validations}
		}),
		Groups: MapSlice(rot.groupListValidations, func(validations [][]*proto.APLValidation) *proto.APLGroupStats {
			return &proto.APLGroupStats{Actions: MapSlice(validations, func(validations []*proto.APLValidation) *proto.APLActionStats {
				return &proto.APLActionStats{Validations: validations}
			})}
		}),
		UuidValidations: uuidValidationsArr,
	}
}

func getAllActionsSafe(action *APLAction) []*APLAction {
	// Check if action is nil before calling GetAllActions
	if action == nil {
		return []*APLAction{}
	}
	return action.GetAllActions()
}

func (rot *APLRotation) allAPLActions() []*APLAction {
	if rot == nil || rot.priorityList == nil {
		return []*APLAction{}
	}

	actions := Flatten(MapSlice(rot.priorityList, getAllActionsSafe))

	//Grab rot.groups actions as well
	if rot.groups != nil {
		for _, group := range rot.groups {
			groupActions := Flatten(MapSlice(group.actions, getAllActionsSafe))
			actions = append(actions, groupActions...)
		}
	}

	return actions
}

// Whether a Move sits anywhere in the rotation.
func (rot *APLRotation) hasMove() bool {
	if !rot.hasMoveKnown {
		rot.hasMoveKnown = true
		rot.hasMoveValue = slices.ContainsFunc(rot.allAPLActions(), func(action *APLAction) bool {
			_, isMove := action.impl.(*APLActionMove)
			return isMove
		})
	}
	return rot.hasMoveValue
}

// A set swapped in before the pull stays on into the fight until an Item Swap: Main runs. A prepull swap away from
// the main set warns when no Item Swap: Main follows it in the prepull or sits in the rotation.
func (rot *APLRotation) warnOnPrepullSwapWithoutMain() {
	swapsTo := func(action *APLAction, main bool) bool {
		return slices.ContainsFunc(getAllActionsSafe(action), func(inner *APLAction) bool {
			swap, ok := inner.impl.(*APLActionItemSwap)
			return ok && (swap.swapSet == proto.APLActionItemSwap_Main) == main
		})
	}

	// Prepull actions run in time order, and in list order at the same time.
	runsNoEarlier := func(j, i int) bool {
		return rot.prepullDoAt[j] > rot.prepullDoAt[i] || (rot.prepullDoAt[j] == rot.prepullDoAt[i] && j >= i)
	}
	lastMain := -1 // The Main swap that runs last.
	var swapsAway []int
	for i, action := range rot.prepullActions {
		if swapsTo(action, true) && (lastMain < 0 || runsNoEarlier(i, lastMain)) {
			lastMain = i
		}
		if swapsTo(action, false) {
			swapsAway = append(swapsAway, i)
		}
	}
	// Walked from the priority list, so a group only counts where a shown line references it.
	if len(swapsAway) == 0 || slices.ContainsFunc(rot.priorityList, func(action *APLAction) bool { return swapsTo(action, true) }) {
		return
	}

	for _, i := range swapsAway {
		if lastMain >= 0 && runsNoEarlier(lastMain, i) {
			continue
		}
		rot.doAndRecordWarnings(&rot.prepullValidations[rot.prepullIdxMap[i]], true, func() {
			rot.ValidationMessage(proto.LogLevel_Warning, "Nothing swaps back to the main set, so the swapped set stays on for the whole fight. Add an Item Swap: Main to the rotation.")
		})
	}
}

// Returns all action objects from the prepull as an unstructured list. Used for easily finding specific actions.
func (rot *APLRotation) allPrepullActions() []*APLAction {
	return Flatten(MapSlice(rot.prepullActions, func(action *APLAction) []*APLAction { return action.GetAllActions() }))
}

func (rot *APLRotation) reset(sim *Simulation) {
	rot.controllingActions = rot.controllingActions[:0]
	rot.inLoop = false
	rot.runningEntryPath = ""
	rot.interruptChannelIf = nil
	rot.allowChannelRecastOnInterrupt = false
	rot.evalGeneration++ // Invalidate any variable caches from previous iteration or initialization
	if rot.resetActions == nil {
		rot.resetActions = rot.allAPLActions()
	}
	for _, action := range rot.resetActions {
		action.impl.Reset(sim)
	}
}

// We intentionally try to mimic the behavior of simc APL to avoid confusion
// and leverage the community's existing familiarity.
// https://github.com/simulationcraft/simc/wiki/ActionLists
func (apl *APLRotation) DoNextAction(sim *Simulation) {
	if sim.CurrentTime < 0 && !apl.unit.engagedBeforePull {
		return
	}

	if apl.inLoop {
		return
	}
	apl.unit.drawReactionTime(sim)

	// An event at a cast's last instant, such as a rage tick, wakes the rotation before the cast's
	// own completion runs; the cast completes first.
	apl.unit.completeDueHardcast(sim)

	if apl.unit.IsChanneling() {
		dot := apl.unit.ChanneledDot
		// All ticks consumed but aura not yet expired (lazy expiry). Deactivate now so
		// the caster doesn't idle between GCD-ready and the lazy aura expiry time.
		if dot.remainingTicks == 0 {
			dot.Deactivate(sim)
			return
		}
		// Increment evalGeneration so variable refs re-evaluate with current game state,
		// not the cached values from the previous getNextAction call.
		apl.evalGeneration++
		// Also evaluate interruptIf when the GCD fires, not only on each tick.
		if apl.shouldInterruptChannel(sim) {
			clipDelay := dot.getChannelClipDelay()
			dot.stopTicking(sim)
			dot.Deactivate(sim)
			if dot.Spell.Unit.GCD.IsReady(sim) {
				apl.unit.WaitUntil(sim, sim.CurrentTime+clipDelay)
			}
		}
		return
	}

	if !apl.unit.RotationTimer.IsReady(sim) {
		return
	}

	i := 0
	apl.inLoop = true

	apl.unit.UpdatePosition(sim, false)
	if len(apl.recheckedMoves) > 0 {
		apl.evalGeneration++
		for _, move := range apl.recheckedMoves {
			move.recheck(sim)
		}
	}
	free := apl.unit.GCD.IsReady(sim) && apl.unit.Hardcast.Expires <= sim.CurrentTime
	apl.unit.costBlocked = false
	for nextAction := apl.firstAction(sim); ; i, nextAction = i+1, apl.getNextAction(sim) {
		if i > 1000 {
			panic(fmt.Sprintf("[USER_ERROR] Infinite loop detected, current action:\n%s", nextAction))
		}

		// The spell queued after the hardcast that just ended goes off first in this pass, a major cooldown
		// alone ahead of it; what the rotation puts after the GCD it starts, such as an explosive, still goes now.
		if apl.unit.queuedSpellDue(sim) && (nextAction == nil || !apl.isMajorCooldown(sim, nextAction)) {
			apl.unit.castDueQueuedSpell(sim)
		} else if nextAction != nil {
			nextAction.Execute(sim)
		} else {
			break
		}
		// Shoot has no GCD to end the pass; any later cast would cut it.
		if apl.unit.Wand.wanding() {
			i++
			break
		}
	}
	apl.inLoop = false

	// OOM from the first pass that skips a spell for its mana cost to the first that skips none, counting
	// only passes the unit starts free to act; a pass during the GCD or a cast sees the costs of few spells.
	if free {
		if apl.unit.costBlocked {
			apl.unit.StartOOMEvent(sim)
		} else {
			apl.unit.EndOOMEvent(sim)
		}
	}

	// A look while a hardcast holds the rotation, at its end with the next spell queued after it, or as that
	// spell goes off, waits on that spell, so it logs no idle pass and no pause.
	hardcastHoldsRotation := apl.unit.hardcastHoldsRotation(sim)
	waitsOnCast := hardcastHoldsRotation || apl.unit.spellQueuedAfterHardcast() || apl.unit.queuedSpellWentOff(sim)

	if sim.Log != nil && i == 0 && !waitsOnCast {
		apl.unit.Log(sim, "No available actions!")
	}

	// Schedule the next rotation evaluation based on either the GCD or reaction time
	if apl.unit.RotationTimer.IsReady(sim) {
		nextEvaluation := sim.CurrentTime + apl.unit.ReactionTime

		// A rotation that moves keeps looking through the GCD, so a walk timed to a swing timer starts
		// when its condition turns true, as a player watching the timer starts it.
		if !apl.unit.Moving && !apl.hasMove() {
			nextEvaluation = max(nextEvaluation, apl.unit.NextGCDAt())
		}

		// Under a hardcast that holds the rotation, the next look is at its end, where the pick is made.
		if hardcastHoldsRotation {
			nextEvaluation = apl.unit.Hardcast.Expires
		}

		if waitsOnCast {
			apl.unit.SetRotationTimer(sim, nextEvaluation)
		} else {
			apl.unit.WaitUntil(sim, nextEvaluation)
		}
	}
}

func (apl *APLRotation) getNextAction(sim *Simulation) *APLAction {
	apl.evalGeneration++

	if len(apl.controllingActions) != 0 {
		return apl.controllingActions[len(apl.controllingActions)-1].GetNextAction(sim)
	}

	held := apl.unit.queuedSpellHoldsCasts(sim)
	for _, action := range apl.priorityList {
		if !(held && action.heldByQueuedSpell(sim)) && action.IsReady(sim) {
			return action
		}
	}

	return nil
}

// The pass's first action. With the spell queued after a hardcast due and no major cooldown ready to go
// ahead of it, there is none, and the spell goes off without a look at the priority list.
func (apl *APLRotation) firstAction(sim *Simulation) *APLAction {
	if apl.unit.queuedSpellDue(sim) && len(apl.controllingActions) == 0 {
		apl.evalGeneration++ // The variable refs in the major cooldowns' conditions read the state after the cast landed.
		if !apl.majorCooldownMayGo(sim) {
			return nil
		}
	}
	return apl.getNextAction(sim)
}

// Whether an action of the priority list that can be a major cooldown is ready.
func (apl *APLRotation) majorCooldownMayGo(sim *Simulation) bool {
	if !apl.majorCooldownActionsKnown {
		apl.majorCooldownActionsKnown = true
		apl.majorCooldownActions = FilterSlice(apl.priorityList, canBeMajorCooldown)
	}
	// A Strict Sequence asked whether it is ready marks the rotation as in a sequence. Nothing asked here
	// runs, so the mark is put back.
	inSequence := apl.inSequence
	mayGo := slices.ContainsFunc(apl.majorCooldownActions, func(action *APLAction) bool { return action.IsReady(sim) })
	apl.inSequence = inSequence
	return mayGo
}

// Whether isMajorCooldown can accept the action: an autocast, a cast of a major cooldown, or a group holding
// one of these. The two must agree, or a major cooldown misses its turn ahead of the spell queued after a
// hardcast.
func canBeMajorCooldown(action *APLAction) bool {
	switch impl := action.impl.(type) {
	case *APLActionAutocastOtherCooldowns:
		return true
	case *APLActionCastSpell:
		return impl.spell.Flags.Matches(SpellFlagMCD)
	case *APLActionCastFriendlySpell:
		return impl.spell.Flags.Matches(SpellFlagMCD)
	case *APLActionChannelSpell:
		return impl.spell.Flags.Matches(SpellFlagMCD)
	case *APLActionGroupReference:
		return impl.group != nil && slices.ContainsFunc(impl.group.actions, canBeMajorCooldown)
	}
	return false
}

// A major cooldown, or a group whose first ready action is one. canBeMajorCooldown must accept every action
// this accepts.
func (apl *APLRotation) isMajorCooldown(sim *Simulation, action *APLAction) bool {
	switch impl := action.impl.(type) {
	case *APLActionAutocastOtherCooldowns:
		return true
	case *APLActionCastSpell:
		return impl.spell.Flags.Matches(SpellFlagMCD)
	case *APLActionCastFriendlySpell:
		return impl.spell.Flags.Matches(SpellFlagMCD)
	case *APLActionChannelSpell:
		return impl.spell.Flags.Matches(SpellFlagMCD)
	case *APLActionGroupReference:
		if impl.group == nil {
			return false
		}
		for _, groupAction := range impl.group.actions {
			if groupAction.IsReady(sim) {
				return apl.isMajorCooldown(sim, groupAction)
			}
		}
	}
	return false
}

// HasBlockingControllingAction reports whether the rotation is currently parked
// in a controlling action that has set the rotation timer to a specific wake
// time and must not be cut short by reaction-time-based shrinks (Wait, WaitUntil).
// Cooperative controllers like StrictSequence return false.
func (apl *APLRotation) HasBlockingControllingAction() bool {
	if apl == nil || len(apl.controllingActions) == 0 {
		return false
	}
	switch apl.controllingActions[len(apl.controllingActions)-1].(type) {
	case *APLActionWait, *APLActionWaitUntil:
		return true
	}
	return false
}

func (apl *APLRotation) pushControllingAction(ca APLActionImpl) {
	apl.controllingActions = append(apl.controllingActions, ca)
}

func (apl *APLRotation) popControllingAction(ca APLActionImpl) {
	if len(apl.controllingActions) == 0 || apl.controllingActions[len(apl.controllingActions)-1] != ca {
		panic("Wrong APL controllingAction in pop()")
	}
	apl.controllingActions = apl.controllingActions[:len(apl.controllingActions)-1]
}

// Whether the first action the rotation would run casts the channel's spell. Read inside withChannelEnded.
func (apl *APLRotation) firstActionRecasts(sim *Simulation, channeled *Spell) bool {
	if step, controlled := apl.controlledStep(); controlled {
		return step != nil && step.IsReady(sim) && step.casts(channeled)
	}
	_, recasts := apl.firstActionIn(sim, apl.priorityList, channeled)
	return recasts
}

// The current step of the strict sequence that controls the rotation, read without moving the sequence
// on; nil under any other controlling action, which runs on its own. controlled says one controls it.
func (apl *APLRotation) controlledStep() (step *APLAction, controlled bool) {
	n := len(apl.controllingActions)
	if n == 0 {
		return nil, false
	}
	if sequence, ok := apl.controllingActions[n-1].(*APLActionStrictSequence); ok {
		return sequence.current(), true
	}
	return nil, true
}

// Walks actions as the rotation would, a referenced group's actions in the reference's place and a
// sequence's current step in the sequence's, to the first one that would run: found says there is one,
// recasts that it casts the channel's spell. A cast or channel is read with CanCast rather than
// CanCastOrQueue, so a spell queued for the GCD's end does not end the channel before then; it fires
// after the channel's next tick. Autocast Other Cooldowns fires its cooldowns beside the rotation and is
// passed over.
func (apl *APLRotation) firstActionIn(sim *Simulation, actions []*APLAction, channeled *Spell) (found, recasts bool) {
	for _, action := range actions {
		if action.condition != nil && !action.condition.GetBool(sim) {
			continue
		}

		var inner []*APLAction
		var canCast bool
		switch impl := action.impl.(type) {
		case *APLActionAutocastOtherCooldowns:
			continue
		case *APLActionGroupReference:
			if impl.group == nil {
				continue
			}
			inner = impl.group.actions
		case *APLActionSequence:
			if impl.curIdx == len(impl.subactions) {
				continue
			}
			inner = impl.subactions[impl.curIdx : impl.curIdx+1]
		case *APLActionStrictSequence:
			if !impl.IsReady(sim) {
				continue
			}
			inner = impl.subactions[impl.curIdx : impl.curIdx+1]
		case *APLActionCastSpell:
			canCast = impl.spell.CanCast(sim, impl.target.Get()) && !impl.spell.majorCooldownWaits(sim)
		case *APLActionCastFriendlySpell:
			canCast = impl.spell.CanCast(sim, impl.target.Get()) && !impl.spell.majorCooldownWaits(sim)
		case *APLActionChannelSpell:
			canCast = impl.spell.CanCast(sim, impl.target.Get())
		default:
			canCast = action.impl.IsReady(sim)
		}

		if inner != nil {
			if found, recasts := apl.firstActionIn(sim, inner, channeled); found {
				return true, recasts
			}
			continue
		}
		if action.casts(channeled) {
			return true, canCast
		}
		if canCast {
			return true, false
		}
	}

	return false, false
}

func (apl *APLRotation) shouldInterruptChannel(sim *Simulation) bool {
	channeledDot := apl.unit.ChanneledDot

	if channeledDot == nil || !channeledDot.ChannelCanBeInterrupted(sim) {
		return false
	}

	var recasts bool
	apl.withChannelEnded(func() {
		recasts = apl.firstActionRecasts(sim, channeledDot.Spell)
	})

	if recasts {
		return apl.allowChannelRecastOnInterrupt
	}

	return true
}

// Runs evaluate as if the channel had ended, with no ChanneledDot and the channel's aura read as down,
// which it restores after. A variable read while the channel was up reads again. A sequence asked
// whether it is ready marks the rotation as in a sequence; nothing asked here runs, so the mark is put
// back.
func (apl *APLRotation) withChannelEnded(evaluate func()) {
	dot, inSequence := apl.unit.ChanneledDot, apl.inSequence
	apl.unit.ChanneledDot = nil
	apl.unit.endedChannel = dot
	apl.evalGeneration++
	evaluate()
	apl.unit.endedChannel = nil
	apl.unit.ChanneledDot = dot
	apl.inSequence = inSequence
}

// A starting channel takes the condition of the channel action that cast it. It takes the action's
// interruptIf only if it starts inside the cast, with no travel delay before its dot applies.
func (apl *APLRotation) adoptChannelAction(dot *Dot) {
	if action := apl.unit.takeChannelAction(dot.Spell); action != nil {
		apl.interruptChannelIf = action.interruptIf
		apl.allowChannelRecastOnInterrupt = action.allowRecast
	}
}

func APLRotationFromJsonString(jsonString string) *proto.APLRotation {
	apl := &proto.APLRotation{}
	data := []byte(jsonString)
	if err := protojson.Unmarshal(data, apl); err != nil {
		panic(err)
	}
	return apl
}

// Add newAPLActionWithGroupVars to propagate groupVars to action condition and impl
func (rot *APLRotation) newAPLActionWithGroupVars(config *proto.APLAction, groupVars map[string]*proto.APLValue) *APLAction {
	if config == nil {
		return nil
	}

	condition := rot.coerceTo(rot.newAPLValueWithContext(config.Condition, groupVars), proto.APLValueType_ValueTypeBool)
	if config.Condition != nil && config.Condition.Value != nil && condition == nil {
		rot.ValidationMessage(proto.LogLevel_Warning, "Condition could not be resolved and will be treated as always true. Add an 'Aura Known' check if this action should be skipped when the aura is absent.")
	}

	impl := rot.newAPLActionImpl(config)
	if impl == nil {
		return nil
	}
	return newAPLActionEntry(condition, impl)
}

// Re-resolve variable references in an APLValue with updated group variables
func (rot *APLRotation) reResolveVariableRefs(value APLValue, groupVars map[string]*proto.APLValue) APLValue {
	if value == nil {
		return nil
	}

	// Check if this is a variable reference that needs re-resolving
	if varRef, ok := value.(*APLValueVariableRef); ok {
		// Re-resolve the variable reference with the updated group variables
		if groupVars != nil {
			if val, ok := groupVars[varRef.name]; ok {
				resolved := rot.newAPLValue(val)
				if resolved != nil {
					// Group override: update resolved value and use a private cache
					// (not shared with global variable refs)
					varRef.resolved = resolved
					varRef.cache = &variableCache{}
					return varRef
				}
			}
		}

		// Fall back to global value variables
		for _, condVar := range rot.valueVariables {
			if condVar.name == varRef.name {
				resolved := rot.newAPLValue(condVar.value)
				if resolved != nil {
					varRef.resolved = resolved
					return varRef
				}
			}
		}
	}

	// For other value types, recursively re-resolve their inner values
	innerValues := value.GetInnerValues()
	if len(innerValues) > 0 {
		for i, innerValue := range innerValues {
			innerValues[i] = rot.reResolveVariableRefs(innerValue, groupVars)
		}
	}

	return value
}

// orphanValue collects a value and all its descendants into orphanedValues
// so they still get Finalize() called even though they've been removed from
// the active APL tree by compile-time constant folding.
func (rot *APLRotation) orphanValue(value APLValue) {
	if value == nil {
		return
	}
	rot.orphanedValues = append(rot.orphanedValues, value)
	for _, inner := range value.GetInnerValues() {
		rot.orphanValue(inner)
	}
}

// The rows of the rotation by their JSON paths in the APL. A group referenced twice is parsed twice, and both
// copies carry the path of the one in the APL.
func (rot *APLRotation) RowPaths() map[*APLAction]string {
	paths := map[*APLAction]string{}
	for i, action := range rot.prepullActions {
		paths[action] = PrepullRowPath(rot.prepullIdxMap[i])
	}
	for i, action := range rot.priorityList {
		paths[action] = PriorityRowPath(rot.priorityListIdxMap[i])
	}
	for g, group := range rot.groups {
		inAPL := slices.IndexFunc(rot.groups, func(other *APLGroup) bool { return other.name == group.name })
		for i, action := range group.actions {
			paths[action] = GroupRowPath(inAPL, rot.groupListIdxMap[g][i])
		}
	}
	return paths
}

// The JSON path of a row in the APL.
func PrepullRowPath(i int) string  { return fmt.Sprintf("prepullActions/%d", i) }
func PriorityRowPath(i int) string { return fmt.Sprintf("priorityList/%d", i) }
func GroupRowPath(group, i int) string {
	return fmt.Sprintf("groups/%d/actions/%d", group, i)
}

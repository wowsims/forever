package core

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type APLActionCastSpell struct {
	defaultAPLActionImpl
	spell  *Spell
	target UnitReference
}

func (rot *APLRotation) newActionCastSpell(config *proto.APLActionCastSpell) APLActionImpl {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	target := rot.GetTargetUnit(config.Target)
	if target.Get() == nil {
		return nil
	}
	return &APLActionCastSpell{
		spell:  spell,
		target: target,
	}
}
func (action *APLActionCastSpell) IsReady(sim *Simulation) bool {
	return action.spell.CanCastOrQueue(sim, action.target.Get()) &&
		(!action.spell.Flags.Matches(SpellFlagMCD) ||
			action.spell.Flags.Matches(SpellFlagReactive) ||
			action.spell.Unit.GCD.IsReady(sim) ||
			action.spell.Unit.Rotation.inSequence)
}
func (action *APLActionCastSpell) Execute(sim *Simulation) {
	action.spell.CastOrQueue(sim, action.target.Get())
}
func (action *APLActionCastSpell) String() string {
	return fmt.Sprintf("Cast Spell(%s)", action.spell.ActionID)
}

// Runs the action's own post-finalize step, then warns for each spell the action and its inner
// actions cast from out of reach. The actions a group reference expands to are skipped: the
// group's own actions warn.
func (rot *APLRotation) postFinalize(action *APLAction) {
	action.impl.PostFinalize(rot)
	all := action.GetAllActions()
	inGroup := make(map[*APLAction]bool)
	for _, inner := range all {
		if ref, ok := inner.impl.(*APLActionGroupReference); ok {
			for _, grouped := range ref.GetInnerActions() {
				inGroup[grouped] = true
			}
		}
	}
	for _, inner := range all {
		if spell := castSpellOf(inner); spell != nil && !inGroup[inner] {
			rot.warnIfNeverInRange(spell)
		}
	}
}

func castSpellOf(action *APLAction) *Spell {
	switch impl := action.impl.(type) {
	case *APLActionCastSpell:
		return impl.spell
	case *APLActionCastFriendlySpell:
		return impl.spell
	case *APLActionChannelSpell:
		return impl.spell
	case *APLActionMultidot:
		return impl.spell
	case *APLActionStrictMultidot:
		return impl.spell
	case *APLActionMultishield:
		return impl.spell
	}
	return nil
}

// The range check refuses a cast from farther than the spell's MaxRange, so a spell that does not
// reach the distance the character starts at is never cast unless the rotation moves it closer. A
// range mod that is not active yet may still lengthen the reach, so the spell gets the benefit of it.
func (rot *APLRotation) warnIfNeverInRange(spell *Spell) {
	reach := spell.EffectiveMaxRange()
	start := rot.unit.StartDistanceFromTarget
	if reach == 0 || start <= reach || rot.canMoveCloser() || hasInactiveRangeMod(spell) {
		return
	}
	rot.ValidationMessage(proto.LogLevel_Warning, "%s reaches %.0f yd and the character starts %.0f yd from the target; nothing in the rotation moves it closer, so it is never cast.", spell.ActionID, reach, start)
}

// Whether a move, a custom rotation or a cast that carries the caster to its target sits anywhere in
// the rotation or the pre-pull. A Move Duration moves on the spot and does not count.
func (rot *APLRotation) canMoveCloser() bool {
	if !rot.canMoveCloserKnown {
		rot.canMoveCloserKnown = true
		rot.canMoveCloserValue = anyMovesCloser(rot.allAPLActions()) || anyMovesCloser(rot.allPrepullActions())
	}
	return rot.canMoveCloserValue
}

func anyMovesCloser(actions []*APLAction) bool {
	for _, action := range actions {
		switch impl := action.impl.(type) {
		case *APLActionMove, *APLActionCustomRotation:
			return true
		case *APLActionCastSpell:
			if impl.spell.Flags.Matches(SpellFlagMovesToTarget) {
				return true
			}
		}
	}
	return false
}

func hasInactiveRangeMod(spell *Spell) bool {
	for _, mod := range spell.Unit.spellMods {
		if !mod.IsActive && (mod.Kind == SpellMod_Range_Flat || mod.Kind == SpellMod_Range_Pct) && slices.Contains(mod.AffectedSpells, spell) {
			return true
		}
	}
	return false
}

type APLActionCastFriendlySpell struct {
	defaultAPLActionImpl
	spell  *Spell
	target UnitReference
}

func (rot *APLRotation) newActionCastFriendlySpell(config *proto.APLActionCastFriendlySpell) APLActionImpl {
	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	target := rot.GetTargetUnit(config.Target)
	if target.Get() == nil {
		return nil
	}
	return &APLActionCastFriendlySpell{
		spell:  spell,
		target: target,
	}
}
func (action *APLActionCastFriendlySpell) IsReady(sim *Simulation) bool {
	return action.spell.CanCastOrQueue(sim, action.target.Get()) && (!action.spell.Flags.Matches(SpellFlagMCD) || action.spell.Flags.Matches(SpellFlagReactive) || action.spell.Unit.GCD.IsReady(sim) || action.spell.Unit.Rotation.inSequence)
}
func (action *APLActionCastFriendlySpell) Execute(sim *Simulation) {
	action.spell.CastOrQueue(sim, action.target.Get())
}
func (action *APLActionCastFriendlySpell) String() string {
	return fmt.Sprintf("Cast Friendly Spell(%s)", action.spell.ActionID)
}

type APLActionChannelSpell struct {
	defaultAPLActionImpl
	spell       *Spell
	target      UnitReference
	interruptIf APLValue
	allowRecast bool
}

func (rot *APLRotation) newActionChannelSpell(config *proto.APLActionChannelSpell) APLActionImpl {
	interruptIf := rot.coerceTo(rot.newAPLValue(config.InterruptIf), proto.APLValueType_ValueTypeBool)
	if interruptIf == nil {
		return rot.newActionCastSpell(&proto.APLActionCastSpell{
			SpellId: config.SpellId,
			Target:  config.Target,
		})
	}

	spell := rot.GetAPLSpell(config.SpellId)
	if spell == nil {
		return nil
	}
	if !spell.Flags.Matches(SpellFlagChanneled) {
		return nil
	}

	target := rot.GetTargetUnit(config.Target)
	if target.Get() == nil {
		return nil
	}

	return &APLActionChannelSpell{
		spell:       spell,
		target:      target,
		interruptIf: interruptIf,
		allowRecast: config.AllowRecast,
	}
}
func (action *APLActionChannelSpell) GetAPLValues() []APLValue {
	return []APLValue{action.interruptIf}
}
func (action *APLActionChannelSpell) IsReady(sim *Simulation) bool {
	return action.spell.CanCastOrQueue(sim, action.target.Get())
}
func (action *APLActionChannelSpell) Execute(sim *Simulation) {
	action.spell.castOrQueue(sim, action.target.Get(), action)
}
func (action *APLActionChannelSpell) String() string {
	return fmt.Sprintf("Channel Spell(%s, interruptIf=%s)", action.spell.ActionID, action.interruptIf)
}

type APLActionMultidot struct {
	defaultAPLActionImpl
	spell      *Spell
	maxDots    int32
	maxOverlap APLValue

	nextTarget *Unit
}

func (rot *APLRotation) newActionMultidot(config *proto.APLActionMultidot) APLActionImpl {
	unit := rot.unit

	spell := rot.GetAPLMultidotSpell(config.SpellId)
	if spell == nil {
		return nil
	}

	maxOverlap := rot.coerceTo(rot.newAPLValue(config.MaxOverlap), proto.APLValueType_ValueTypeDuration)
	if maxOverlap == nil {
		maxOverlap = rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, &proto.UUID{Value: ""})
	}

	maxDots := config.MaxDots
	maxTargets := unit.Env.TotalTargetCount()
	if spell.Flags.Matches(SpellFlagHelpful) {
		maxTargets = int32(len(unit.Env.Raid.AllPlayerUnits))
	}
	if maxTargets < maxDots {
		rot.ValidationMessage(proto.LogLevel_Warning, "Encounter only has %d targets. Using that for Max Dots instead of %d", maxTargets, maxDots)
		maxDots = maxTargets
	}

	return &APLActionMultidot{
		spell:      spell,
		maxDots:    maxDots,
		maxOverlap: maxOverlap,
	}
}
func (action *APLActionMultidot) GetAPLValues() []APLValue {
	return []APLValue{action.maxOverlap}
}
func (action *APLActionMultidot) Reset(*Simulation) {
	action.nextTarget = nil
}
func (action *APLActionMultidot) IsReady(sim *Simulation) bool {
	maxOverlap := action.maxOverlap.GetDuration(sim)

	if action.spell.Flags.Matches(SpellFlagHelpful) {
		for i := int32(0); i < action.maxDots; i++ {
			target := sim.Raid.AllPlayerUnits[i]
			dot := action.spell.Dot(target)
			if (!dot.IsActive() || dot.RemainingDuration(sim) < maxOverlap) && action.spell.CanCastOrQueue(sim, target) {
				action.nextTarget = target
				return true
			}
		}
	} else {
		for i := int32(0); i < action.maxDots; i++ {
			target := sim.Encounter.AllTargetUnits[i]
			dot := action.spell.Dot(target)
			if (!dot.IsActive() || dot.RemainingDuration(sim) < maxOverlap) && action.spell.CanCastOrQueue(sim, target) {
				action.nextTarget = target
				return true
			}
		}
	}
	return false
}
func (action *APLActionMultidot) Execute(sim *Simulation) {
	action.spell.CastOrQueue(sim, action.nextTarget)
}
func (action *APLActionMultidot) String() string {
	return fmt.Sprintf("Multidot(%s)", action.spell.ActionID)
}

type APLActionStrictMultidot struct {
	defaultAPLActionImpl
	unit       *Unit
	spell      *Spell
	maxDots    int32
	maxOverlap APLValue

	isHeal       bool
	targets      []*Unit
	actions      []*APLAction
	readyActions []*APLAction
	entry        *APLAction // The entry holding this action, whose condition the chain re-checks before each cast after the first.
}

func (rot *APLRotation) newActionStrictMultidot(config *proto.APLActionStrictMultidot) APLActionImpl {
	unit := rot.unit

	spell := rot.GetAPLMultidotSpell(config.SpellId)
	if spell == nil {
		return nil
	}

	maxOverlap := rot.coerceTo(rot.newAPLValue(config.MaxOverlap), proto.APLValueType_ValueTypeDuration)
	if maxOverlap == nil {
		maxOverlap = rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, &proto.UUID{Value: ""})
	}

	isHeal := spell.Flags.Matches(SpellFlagHelpful)
	maxDots := config.MaxDots
	numTargets := int32(0)
	var targets = []*Unit{}
	if isHeal {
		targets = unit.Env.Raid.AllPlayerUnits
		numTargets = int32(len(targets))
	} else {
		numTargets = unit.Env.TotalTargetCount()
		targets = unit.Env.Encounter.AllTargetUnits
	}
	if numTargets < maxDots {
		rot.ValidationMessage(proto.LogLevel_Warning, "Encounter only has %d targets. Using that for Max Dots instead of %d", numTargets, maxDots)
		maxDots = numTargets
	}

	actions := []*APLAction{}
	for i := int32(0); i < maxDots; i++ {
		if isHeal {
			actions = append(actions, &APLAction{
				impl: rot.newActionCastSpell(&proto.APLActionCastSpell{
					SpellId: config.SpellId,
					Target: &proto.UnitReference{
						Type:  proto.UnitReference_Player,
						Index: targets[i].Index,
					},
				}),
			})
		} else {
			actions = append(actions, &APLAction{
				impl: rot.newActionCastSpell(&proto.APLActionCastSpell{
					SpellId: config.SpellId,
					Target: &proto.UnitReference{
						Type:  proto.UnitReference_Target,
						Index: targets[i].Index,
					},
				}),
			})
		}
	}

	return &APLActionStrictMultidot{
		unit:       rot.unit,
		spell:      spell,
		isHeal:     isHeal,
		maxDots:    maxDots,
		maxOverlap: maxOverlap,
		actions:    actions,
		targets:    targets,
	}
}
func (action *APLActionStrictMultidot) GetAPLValues() []APLValue {
	return []APLValue{action.maxOverlap}
}
func (action *APLActionStrictMultidot) Reset(*Simulation) {
	action.readyActions = []*APLAction{}
}
func (action *APLActionStrictMultidot) IsReady(sim *Simulation) bool {
	maxOverlap := action.maxOverlap.GetDuration(sim)
	action.readyActions = []*APLAction{}

	var previousTarget *Unit
	for i := int32(0); i < action.maxDots; i++ {
		target := action.targets[i]
		dot := action.spell.Dot(target)
		if (!dot.IsActive() || dot.RemainingDuration(sim) < maxOverlap) && action.spell.CanCastOrQueue(sim, target) {
			action.readyActions = append(action.readyActions, action.actions[i])
			previousTarget = target
		} else if previousTarget != nil {
			previousDot := action.spell.Dot(previousTarget)
			// Take the previous Dot + Cast Time into account for the overlap check
			if (!previousDot.IsActive() || dot.RemainingDuration(sim)-action.spell.EffectiveCastTime() <= previousDot.RemainingDuration(sim)) && action.spell.CanCastOrQueue(sim, target) {
				action.readyActions = append(action.readyActions, action.actions[i])
				previousTarget = target
			}
		}
	}

	return len(action.readyActions) > 0
}
func (action *APLActionStrictMultidot) Execute(sim *Simulation) {
	groupConditions := action.unit.Rotation.groupConditions
	var gate APLValue
	if len(groupConditions) > 0 || action.entry.condition != nil {
		vals := make([]APLValue, 0, len(groupConditions)+1)
		vals = append(vals, groupConditions...)
		if action.entry.condition != nil {
			vals = append(vals, action.entry.condition)
		}
		gate = &APLValueAnd{vals: vals}
	}
	sequence := APLActionStrictSequence{
		unit:       action.unit,
		subactions: action.readyActions,
		gate:       gate,
	}
	sequence.Execute(sim)
}
func (action *APLActionStrictMultidot) String() string {
	return fmt.Sprintf("StrictMultidot(%s)", action.spell.ActionID)
}

type APLActionMultishield struct {
	defaultAPLActionImpl
	spell      *Spell
	maxShields int32
	maxOverlap APLValue

	nextTarget *Unit
}

func (rot *APLRotation) newActionMultishield(config *proto.APLActionMultishield) APLActionImpl {
	unit := rot.unit

	spell := rot.GetAPLMultishieldSpell(config.SpellId)
	if spell == nil {
		return nil
	}

	maxOverlap := rot.coerceTo(rot.newAPLValue(config.MaxOverlap), proto.APLValueType_ValueTypeDuration)
	if maxOverlap == nil {
		maxOverlap = rot.newValueConst(&proto.APLValueConst{Val: "0ms"}, &proto.UUID{Value: ""})
	}

	maxShields := config.MaxShields
	numTargets := int32(len(unit.Env.Raid.AllPlayerUnits))
	if numTargets < maxShields {
		rot.ValidationMessage(proto.LogLevel_Warning, "Encounter only has %d targets. Using that for Max Shields instead of %d", numTargets, maxShields)
		maxShields = numTargets
	}

	return &APLActionMultishield{
		spell:      spell,
		maxShields: maxShields,
		maxOverlap: maxOverlap,
	}
}
func (action *APLActionMultishield) GetAPLValues() []APLValue {
	return []APLValue{action.maxOverlap}
}
func (action *APLActionMultishield) Reset(*Simulation) {
	action.nextTarget = nil
}
func (action *APLActionMultishield) IsReady(sim *Simulation) bool {
	maxOverlap := action.maxOverlap.GetDuration(sim)

	for i := int32(0); i < action.maxShields; i++ {
		target := sim.Raid.AllPlayerUnits[i]
		shield := action.spell.Shield(target)
		if (!shield.IsActive() || shield.RemainingDuration(sim) < maxOverlap) && action.spell.CanCastOrQueue(sim, target) {
			action.nextTarget = target
			return true
		}
	}
	return false
}
func (action *APLActionMultishield) Execute(sim *Simulation) {
	action.spell.CastOrQueue(sim, action.nextTarget)
}
func (action *APLActionMultishield) String() string {
	return fmt.Sprintf("Multishield(%s)", action.spell.ActionID)
}

type APLActionCastAllStatBuffCooldowns struct {
	defaultAPLActionImpl
	character *Character

	statTypesToMatch []stats.Stat

	allSubactions   []*APLActionCastSpell
	readySubactions []*APLActionCastSpell
}

func (rot *APLRotation) newActionCastAllStatBuffCooldowns(config *proto.APLActionCastAllStatBuffCooldowns) APLActionImpl {
	unit := rot.unit
	actionImpl := &APLActionCastAllStatBuffCooldowns{
		character:        unit.Env.Raid.GetPlayerFromUnit(unit).GetCharacter(),
		statTypesToMatch: stats.IntTupleToStatsList(config.StatType1, config.StatType2, config.StatType3),
	}

	unit.Env.RegisterPostFinalizeEffect(func() {
		// This needs to happen after the rotation is finalized so that
		// all manually casted MCDs are removed from the cooldown list
		// before it is filtered for the desired buff types.
		actionImpl.processMajorCooldowns()
	})

	return actionImpl
}
func (action *APLActionCastAllStatBuffCooldowns) processMajorCooldowns() {
	matchingSpells := action.character.GetMatchingStatBuffSpells(action.statTypesToMatch)

	action.allSubactions = MapSlice(matchingSpells, func(buffSpell *Spell) *APLActionCastSpell {
		return &APLActionCastSpell{
			spell:  buffSpell,
			target: action.character.Rotation.GetTargetUnit(nil),
		}
	})

	action.character.Env.RegisterPostFinalizeEffect(func() {
		// We again need a delayed evaluation here so that other instances of
		// this action within the rotation can assemble their spell lists
		// before the filtered spells are irreversibly removed from the MCD
		// manager.
		for _, buffSpell := range matchingSpells {
			action.character.removeInitialMajorCooldown(buffSpell.ActionID)
		}
	})
}
func (action *APLActionCastAllStatBuffCooldowns) getEquippedSubActions() []*APLActionCastSpell {
	return FilterSlice(action.allSubactions, func(subAction *APLActionCastSpell) bool {
		return !subAction.spell.Flags.Matches(SpellFlagSwapped)
	})
}
func (action *APLActionCastAllStatBuffCooldowns) IsReady(sim *Simulation) bool {
	allEquippedSubactions := action.getEquippedSubActions()
	if len(allEquippedSubactions) == 0 {
		return false
	}

	action.readySubactions = FilterSlice(allEquippedSubactions, func(subAction *APLActionCastSpell) bool {
		return subAction.IsReady(sim)
	})

	return Ternary(action.character.Rotation.inSequence, len(action.readySubactions) == len(allEquippedSubactions), len(action.readySubactions) > 0)
}
func (action *APLActionCastAllStatBuffCooldowns) Execute(sim *Simulation) {
	actionSetToUse := action.readySubactions

	if sim.CurrentTime < 0 {
		actionSetToUse = action.getEquippedSubActions()
	}

	for _, subaction := range actionSetToUse {
		subaction.Execute(sim)
	}
}
func (action *APLActionCastAllStatBuffCooldowns) String() string {
	return fmt.Sprintf("CastAllBuffCooldownsFor(%s)", StringFromStatTypes(action.statTypesToMatch))
}
func (action *APLActionCastAllStatBuffCooldowns) PostFinalize(rot *APLRotation) {
	if len(action.allSubactions) == 0 {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s will not cast any spells! There are either no major cooldowns buffing the specified stat type(s), or all of them are manually cast in the APL.", action)
	} else {
		actionIDs := MapSlice(action.allSubactions, func(subaction *APLActionCastSpell) ActionID {
			return subaction.spell.ActionID
		})

		rot.ValidationMessage(proto.LogLevel_Information, "%s will cast the following spells: %s", action, StringFromActionIDs(actionIDs))
	}
}

type APLActionAutocastOtherCooldowns struct {
	defaultAPLActionImpl
	character *Character

	nextReadyMCD *MajorCooldown
}

func (rot *APLRotation) newActionAutocastOtherCooldowns(config *proto.APLActionAutocastOtherCooldowns) APLActionImpl {
	unit := rot.unit
	return &APLActionAutocastOtherCooldowns{
		character: unit.Env.Raid.GetPlayerFromUnit(unit).GetCharacter(),
	}
}
func (action *APLActionAutocastOtherCooldowns) Reset(*Simulation) {
	action.nextReadyMCD = nil
}
func (action *APLActionAutocastOtherCooldowns) IsReady(sim *Simulation) bool {
	action.nextReadyMCD = action.character.getFirstReadyMCD(sim)

	// Explicitly check for GCD because MCDs are usually desired to be cast immediately
	// before the next spell, rather than immediately after the previous spell. This is
	// true even for MCDs which do not require the GCD. The one exception to this rule
	// is Engineering explosives, which should instead be cast right *after* incurring
	// a GCD, since they are off-GCD but have small cast times.
	return (action.nextReadyMCD != nil) && ((action.character.GCD.IsReady(sim) != action.nextReadyMCD.Type.Matches(CooldownTypeExplosive)) || action.nextReadyMCD.Spell.Flags.Matches(SpellFlagReactive))
}
func (action *APLActionAutocastOtherCooldowns) Execute(sim *Simulation) {
	action.nextReadyMCD.tryActivateHelper(sim, action.character)
	action.character.UpdateMajorCooldowns()
}
func (action *APLActionAutocastOtherCooldowns) String() string {
	return "Autocast Other Cooldowns"
}
func (action *APLActionAutocastOtherCooldowns) PostFinalize(rot *APLRotation) {
	if len(action.character.initialMajorCooldowns) == 0 {
		rot.ValidationMessage(proto.LogLevel_Warning, "%s will not cast any spells! There are either no major cooldowns configured for this character, or all of them are manually cast in the APL.", action)
	} else {
		actionIDs := MapSlice(action.character.initialMajorCooldowns, func(mcd MajorCooldown) ActionID {
			return mcd.Spell.ActionID
		})

		rot.ValidationMessage(proto.LogLevel_Information, "%s will cast the following spells: %s", action, StringFromActionIDs(actionIDs))
	}
}

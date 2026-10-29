// Functions for creating common types of auras.
package core

import (
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/stats"
)

type AuraCallback uint16

func (c AuraCallback) Matches(other AuraCallback) bool {
	return (c & other) != 0
}

const (
	CallbackEmpty AuraCallback = 0

	CallbackOnSpellHitDealt AuraCallback = 1 << iota
	CallbackOnSpellHitTaken
	CallbackOnPeriodicDamageDealt
	CallbackOnHealDealt
	CallbackOnPeriodicHealDealt
	CallbackOnCastComplete
	CallbackOnApplyEffects
	CallbackOnPeriodicDamageTaken
	CallbackOnCastResult
	CallbackOnTargetDies // Any enemy's death. The handler and ExtraCondition get a nil spell and a result whose Target is the dying target; a deferred handler runs after the target is disabled.
	CallbackOnKill       // A death the raid gets the kill for. The handler and ExtraCondition get a nil spell and a result whose Target is the dying target; a deferred handler runs after the target is disabled.

	CallbackLast
)

type ProcHandler func(sim *Simulation, spell *Spell, result *SpellResult)
type ProcExtraCondition func(sim *Simulation, spell *Spell, result *SpellResult) bool

type ProcTrigger struct {
	Name               string
	ActionID           ActionID
	MetricsActionID    ActionID
	Duration           time.Duration
	Callback           AuraCallback
	ProcMask           ProcMask
	ProcMaskExclude    ProcMask
	SpellFlags         SpellFlag
	SpellFlagsExclude  SpellFlag
	Outcome            HitOutcome
	RequireDamageDealt bool

	// The trigger must be a spell with a damage component: a hit that deals damage, or the landed
	// cast of a spell with a DoT, whose application deals none. A DoT counts whether or not its
	// ticks deal damage, and a hit absorbed to 0 from a spell without one does not count.
	RequireDamageComponent bool

	ProcChance          float64
	DPM                 *DynamicProcManager
	ICD                 time.Duration
	Handler             ProcHandler
	TriggerImmediately  bool       // If false (default), the handler will be called one spell batch window later for improved realism.
	ClassSpellMask      int64      // The sim's own custom spell masks, not from the game files.
	ClassFlags          ClassFlags // The spells this trigger fires on (EffectSpellClassMask in the game files); with ClassSpellMask, both must match.
	ProcSpellClassFlags ClassFlags // The proccing spell's class flags, which proc chance talents match (e.g. Improved Poisons on Instant Poison).
	Caster              *Unit      // Whose proc chance mods apply, when not the aura's unit (e.g. Judgement of Light on the target).
	ClassSpellsOnly     bool       // Corresponds to the Only Proc From Class Abilities flag, e.g. https://www.wowhead.com/forever/spell=32106/lesser-spell-blasting
	ExtraCondition      ProcExtraCondition

	// The Can Proc From Procs attribute (Attributes[3] 0x4000000): the listener also fires on hits
	// from spells flagged SpellFlagProc. Without it a proc's hits are invisible to this listener,
	// which is how the game treats every talent, set bonus and Equip effect by default.
	CanProcFromProcs bool
	// A weapon proc: a "Chance on hit" item effect (ItemEffect.TriggerType 2) or a combat enchant
	// (SpellItemEnchantment.Effect 1, e.g. Mongoose, Crusader, poisons, imbues).
	// The game casts these off every melee or ranged hit that lacks SpellFlagSuppressWeaponProcs,
	// whether or not that hit was itself a proc, so SpellFlagProc and CanProcFromProcs are ignored.
	// A weapon proc's own strike is such a hit, unless an off-hand hit cast it.
	// Which hand's hits qualify is still the DPM's or ProcMask's job.
	IsWeaponProc bool
}

// The game's rule for whether a hit may reach a listener at all, before the proc flags are
// matched. Weapon procs (Player::CastItemCombatSpell) hear only melee and ranged hits, and of those
// skip the spells with Suppress Weapon Procs; a hit with no melee or ranged mask, such as the damage
// of a magic item proc, never casts them. Aura procs (Aura::IsProcTriggeredOnEvent) skip triggered
// spells that lack Not a Proc, unless the listener carries Can Proc From Procs. Auto attacks, extra
// attacks included, carry neither flag and so are always eligible.
func (config *ProcTrigger) canProcFrom(spell *Spell) bool {
	if config.IsWeaponProc {
		return spell.ProcMask.Matches(ProcMaskMeleeOrRanged) && !spell.Flags.Matches(SpellFlagSuppressWeaponProcs)
	}
	return config.CanProcFromProcs || !spell.Flags.Matches(SpellFlagProc)
}

// The spell-side checks every callback path applies, in one place so the paths cannot drift:
// eligibility, the flag filters, the class spell filters and the hit-kind mask.
func (config *ProcTrigger) matchesSpell(spell *Spell) bool {
	if !config.canProcFrom(spell) {
		return false
	}
	if config.SpellFlags != SpellFlagNone && !spell.Flags.Matches(config.SpellFlags) {
		return false
	}
	if config.SpellFlagsExclude != SpellFlagNone && spell.Flags.Matches(config.SpellFlagsExclude) {
		return false
	}
	if config.ClassSpellMask > 0 && config.ClassSpellMask&spell.ClassSpellMask == 0 {
		return false
	}
	if !config.ClassFlags.Reaches(spell) {
		return false
	}
	if config.ClassSpellsOnly && spell.ClassSpellMask == 0 && spell.ClassFlags.IsZero() {
		return false
	}
	if config.ProcMaskExclude != ProcMaskUnknown && spell.ProcMask.Matches(config.ProcMaskExclude) {
		return false
	}
	if config.ProcMask != ProcMaskUnknown && !spell.ProcMask.Matches(config.ProcMask) {
		return false
	}
	return true
}

// Whether the trigger's rolls let a hit of this proc mask through. The stated chance draws unless it
// is 1 or more; a manager then rolls its own chance for the mask, and a mask none of its masks hears
// fails without a draw. The unit's proc chance mods raise the manager's chance where there is one and
// the stated chance otherwise. The cast and apply-effects listeners roll no manager.
func (config *ProcTrigger) rolls(sim *Simulation, unit *Unit, dpm *DynamicProcManager, procMask ProcMask) bool {
	if config.Caster != nil {
		unit = config.Caster
	}

	stated := config.ProcChance
	if dpm == nil {
		stated = unit.ModifiedProcChance(config.ProcSpellClassFlags, stated)
	}
	if stated < 1 && sim.RandomFloat(config.Name) > stated {
		return false
	}
	if dpm == nil {
		return true
	}

	chance, heard := dpm.chanceFor(procMask, sim)
	return heard && sim.Proc(unit.ModifiedProcChance(config.ProcSpellClassFlags, chance), config.Name)
}

type deferredProc struct {
	pa        PendingAction
	spell     *Spell
	result    SpellResult
	hasResult bool
	cleanUp   func(*Simulation)
}

func (procAura *Aura) AttachProcTriggerCallback(unit *Unit, config ProcTrigger) {
	var icd Cooldown
	if config.ICD != 0 {
		icd = Cooldown{
			Timer:    unit.NewTimer(),
			Duration: config.ICD,
		}
		procAura.Icd = &icd
	}
	icdOwner := unit
	if config.Caster != nil {
		icdOwner = config.Caster
	}
	startICD := func(sim *Simulation) {
		if config.ICD == 0 {
			return
		}
		icd.Set(sim.CurrentTime + icdOwner.ModifiedProcCooldown(config.ProcSpellClassFlags, config.ICD))
	}

	var dpm *DynamicProcManager
	if config.DPM != nil {
		dpm = config.DPM
		procAura.Dpm = config.DPM
	}

	handler := config.Handler
	var freeDeferred []*deferredProc

	applyHandler := func(sim *Simulation, spell *Spell, result *SpellResult) {
		if config.TriggerImmediately {
			handler(sim, spell, result)
			return
		}

		// Deferred-proc records (PendingAction + result snapshot) come from a
		// per-trigger free list and return to it when they fire or are cleaned up.
		var dp *deferredProc
		if n := len(freeDeferred); n > 0 {
			dp = freeDeferred[n-1]
			freeDeferred = freeDeferred[:n-1]
		} else {
			dp = &deferredProc{}
			release := func() {
				dp.spell = nil
				freeDeferred = append(freeDeferred, dp)
			}
			dp.pa.OnAction = func(sim *Simulation) {
				res := &dp.result
				if !dp.hasResult {
					res = nil
				}
				handler(sim, dp.spell, res)
				release()
			}
			dp.cleanUp = func(sim *Simulation) { release() }
		}
		dp.spell = spell
		dp.hasResult = result != nil
		if result != nil {
			// Snapshot the result: the original may be reused before the handler runs.
			dp.result = *result
			dp.result.inUse = true
		}
		dp.pa.NextActionAt = sim.CurrentTime + SpellBatchWindow
		dp.pa.Priority = ActionPriorityDOT
		dp.pa.CleanUp = dp.cleanUp
		sim.AddPendingAction(&dp.pa)
	}

	callback := func(aura *Aura, sim *Simulation, spell *Spell, result *SpellResult) {
		if !config.matchesSpell(spell) {
			return
		}
		if config.Outcome != OutcomeEmpty {
			if !result.Outcome.Matches(config.Outcome) {
				return
			}
		}
		if config.RequireDamageDealt && result.Damage == 0 {
			return
		}
		if config.RequireDamageComponent && result.Damage == 0 && !spell.hasDot(result.Target) {
			return
		}
		if icd.Duration != 0 && !icd.IsReady(sim) {
			return
		}
		if config.ExtraCondition != nil && !config.ExtraCondition(sim, spell, result) {
			return
		}
		if !config.rolls(sim, unit, dpm, spell.ProcMask) {
			return
		}

		startICD(sim)
		applyHandler(sim, spell, result)
	}

	if config.ProcChance == 0 {
		config.ProcChance = 1
	}

	if config.Callback.Matches(CallbackOnSpellHitDealt) {
		procAura.OnSpellHitDealt = callback
	}
	if config.Callback.Matches(CallbackOnSpellHitTaken) {
		procAura.OnSpellHitTaken = callback
	}
	if config.Callback.Matches(CallbackOnPeriodicDamageDealt) {
		procAura.OnPeriodicDamageDealt = callback
	}
	if config.Callback.Matches(CallbackOnHealDealt) {
		procAura.OnHealDealt = callback
	}
	if config.Callback.Matches(CallbackOnPeriodicHealDealt) {
		procAura.OnPeriodicHealDealt = callback
	}
	if config.Callback.Matches(CallbackOnCastComplete) {
		procAura.OnCastComplete = func(aura *Aura, sim *Simulation, spell *Spell) {
			if !config.matchesSpell(spell) {
				return
			}
			if icd.Duration != 0 && !icd.IsReady(sim) {
				return
			}
			// No result exists for a cast, so a condition here can only look at the spell.
			if config.ExtraCondition != nil && !config.ExtraCondition(sim, spell, nil) {
				return
			}
			if !config.rolls(sim, unit, nil, ProcMaskUnknown) {
				return
			}

			startICD(sim)
			applyHandler(sim, spell, nil)
		}
	}
	if config.Callback.Matches(CallbackOnApplyEffects) {
		procAura.OnApplyEffects = func(aura *Aura, sim *Simulation, target *Unit, spell *Spell) {
			if !config.matchesSpell(spell) {
				return
			}
			if icd.Duration != 0 && !icd.IsReady(sim) {
				return
			}
			emptyResult := spell.NewResult(target)
			defer spell.DisposeResult(emptyResult)
			if config.ExtraCondition != nil && !config.ExtraCondition(sim, spell, emptyResult) {
				return
			}
			if !config.rolls(sim, unit, nil, ProcMaskUnknown) {
				return
			}

			startICD(sim)
			applyHandler(sim, spell, emptyResult)
		}
	}
	if config.Callback.Matches(CallbackOnPeriodicDamageTaken) {
		procAura.OnPeriodicDamageTaken = callback
	}
	if config.Callback.Matches(CallbackOnCastResult) {
		procAura.OnCastResult = callback
	}

	// A death has no spell and deals no damage, so only the ICD, the condition and the chance gate it.
	// The handler gets a nil spell and a result of its own whose Target is the dead unit.
	onDeath := func(aura *Aura, sim *Simulation, target *Unit) {
		if icd.Duration != 0 && !icd.IsReady(sim) {
			return
		}
		deathResult := SpellResult{Target: target}
		if config.ExtraCondition != nil && !config.ExtraCondition(sim, nil, &deathResult) {
			return
		}
		if !config.rolls(sim, unit, nil, ProcMaskUnknown) {
			return
		}

		startICD(sim)
		applyHandler(sim, nil, &deathResult)
	}
	if config.Callback.Matches(CallbackOnTargetDies) {
		procAura.OnTargetDies = onDeath
	}
	if config.Callback.Matches(CallbackOnKill) {
		procAura.OnKill = onDeath
	}
}

func (unit *Unit) MakeProcTriggerAura(config ProcTrigger) *Aura {
	aura := Aura{
		Label:           config.Name,
		ActionIDForProc: config.ActionID,
		Duration:        config.Duration,
		ActionID:        config.MetricsActionID,
	}
	if config.Duration == 0 {
		aura.Duration = NeverExpires
		aura.OnReset = func(aura *Aura, sim *Simulation) {
			aura.Activate(sim)
		}
	}

	aura.AttachProcTriggerCallback(unit, config)

	return unit.GetOrRegisterAura(aura)
}

type CustomStatBuffProcCondition func(sim *Simulation, aura *Aura) bool

// Analog to an Aura "sub-class" that additionally links the Aura to one or more
// Stats. Used within APL snapshotting wrappers.
type StatBuffAura struct {
	*Aura

	// All Stat types that are buffed (before dependencies) when this Aura
	// is activated.
	BuffedStatTypes []stats.Stat

	// Any special conditions (beyond standard ICD checks etc.) that must be
	// satisfied before this Aura can be activated.
	CustomProcCondition CustomStatBuffProcCondition

	// Whether the aura is currently swapped (in another item set) out or not.
	IsSwapped bool
}

func (aura *StatBuffAura) BuffsMatchingStat(statTypesToMatch []stats.Stat) bool {
	if aura == nil {
		return false
	}

	return CheckSliceOverlap(aura.BuffedStatTypes, statTypesToMatch)
}

func (aura *StatBuffAura) CanProc(sim *Simulation) bool {
	return !aura.IsSwapped && ((aura.CustomProcCondition == nil) || aura.CustomProcCondition(sim, aura.Aura))
}

// The stats a buff to any of makes a survival cooldown.
var SurvivalStats = []stats.Stat{stats.Armor, stats.ArmorNoMods, stats.BlockPercent, stats.BlockValue, stats.DodgeRating, stats.ParryRating,
	stats.DodgePercent, stats.ParryPercent, stats.Health, stats.ArcaneResistance, stats.FireResistance, stats.FrostResistance,
	stats.NatureResistance, stats.ShadowResistance}

// The stats a buff to only these makes a mana cooldown.
var ManaStats = []stats.Stat{stats.MP5, stats.Spirit}

// A buff to any defensive stat or resistance, or to Stamina alone, is a survival cooldown, one to
// mana regen or spirit alone a mana cooldown, one to healing power alone a healing cooldown, and any
// other a DPS cooldown.
func CooldownTypeOf(buffed []stats.Stat) CooldownType {
	only := func(set ...stats.Stat) bool {
		return len(buffed) > 0 && !slices.ContainsFunc(buffed, func(stat stats.Stat) bool {
			return !slices.Contains(set, stat)
		})
	}

	switch {
	case CheckSliceOverlap(buffed, SurvivalStats), only(stats.Stamina):
		return CooldownTypeSurvival
	case only(ManaStats...):
		return CooldownTypeMana
	case only(stats.HealingPower):
		return CooldownTypeHealing
	}
	return CooldownTypeDPS
}

func (aura *StatBuffAura) InferCDType() CooldownType {
	return CooldownTypeOf(aura.BuffedStatTypes)
}

type StackingStatAura struct {
	Aura          Aura
	BonusPerStack stats.Stats
}

func MakeStackingAura(character *Character, config StackingStatAura) *StatBuffAura {
	bonusPerStack := config.BonusPerStack
	config.Aura.OnStacksChange = func(aura *Aura, sim *Simulation, oldStacks int32, newStacks int32) {
		character.AddStatsDynamic(sim, bonusPerStack.Multiply(float64(newStacks-oldStacks)))
	}
	return &StatBuffAura{
		Aura:            character.GetOrRegisterAura(config.Aura),
		BuffedStatTypes: bonusPerStack.GetBuffedStatTypes(),
	}
}

// Returns the same Aura for chaining.
func MakePermanent(aura *Aura) *Aura {
	aura.Duration = NeverExpires
	if aura.OnReset == nil {
		aura.OnReset = func(aura *Aura, sim *Simulation) {
			aura.Duration = NeverExpires
			aura.Activate(sim)
		}
	} else {
		oldOnReset := aura.OnReset
		aura.OnReset = func(aura *Aura, sim *Simulation) {
			aura.Duration = NeverExpires
			oldOnReset(aura, sim)
			aura.Activate(sim)
		}
	}
	return aura
}

func BlockPrepull(aura *Aura) *Aura {
	return aura.ApplyOnEncounterStart(func(aura *Aura, sim *Simulation) {
		aura.Deactivate(sim)
	})
}

type TemporaryStatBuffWithStacksConfig struct {
	StackingAuraLabel    string
	StackingAuraActionID ActionID
	AuraLabel            string
	ActionID             ActionID
	BonusPerStack        stats.Stats
	MaxStacks            int32
	TimePerStack         time.Duration
	Duration             time.Duration
	TickImmediately      bool
	// Set when stacks come from an event rather than a timer. The window aura is still built and
	// still bounds the stacking aura, but nothing ticks - the caller adds stacks from its own
	// trigger. The stacking aura is then given no duration of its own and is dropped when the
	// window ends, which is what a child aura with no duration in the game files does.
	StacksFromEvent bool
}

func (character *Character) NewTemporaryStatBuffWithStacks(config TemporaryStatBuffWithStacksConfig) (*StatBuffAura, *Aura) {
	stackingAura := MakeStackingAura(character, StackingStatAura{
		Aura: Aura{
			Label:     Ternary(config.StackingAuraLabel != "", config.StackingAuraLabel, config.AuraLabel),
			ActionID:  Ternary(!config.StackingAuraActionID.IsEmptyAction(), config.StackingAuraActionID, config.ActionID),
			Duration:  Ternary(config.StacksFromEvent, NeverExpires, config.Duration),
			MaxStacks: config.MaxStacks,
		},
		BonusPerStack: config.BonusPerStack,
	})

	if config.TimePerStack > 0 || config.StacksFromEvent {
		aura := character.RegisterAura(Aura{
			Label:    config.AuraLabel,
			ActionID: config.ActionID,
			Duration: config.Duration,
			OnGain: func(aura *Aura, sim *Simulation) {
				stackingAura.Activate(sim)

				if config.TimePerStack <= 0 {
					return
				}

				StartPeriodicAction(sim, PeriodicActionOptions{
					Period:          config.TimePerStack,
					NumTicks:        int(config.MaxStacks),
					TickImmediately: config.TickImmediately,
					OnAction: func(sim *Simulation) {
						if stackingAura.IsActive() {
							stackingAura.AddStack(sim)
						}
					},
				})
			},
			OnExpire: func(aura *Aura, sim *Simulation) {
				// The window owns the stacking aura's lifetime in the event case, because there
				// the child has no duration of its own to expire on.
				if config.StacksFromEvent {
					stackingAura.Deactivate(sim)
				}
			},
		})
		return stackingAura, aura
	}

	return stackingAura, nil

}

// Helper for the common case of making an aura that adds stats.
func (character *Character) NewTemporaryStatsAura(auraLabel string, actionID ActionID, tempStats stats.Stats, duration time.Duration) *StatBuffAura {
	return character.NewTemporaryStatsAuraWrapped(auraLabel, actionID, tempStats, duration, nil)
}

// Alternative that allows modifying the Aura config.
func (character *Character) NewTemporaryStatsAuraWrapped(auraLabel string, actionID ActionID, buffs stats.Stats, duration time.Duration, modConfig func(*Aura)) *StatBuffAura {
	// If one of the stat bonuses is a health bonus, then set up healing metrics for the associated
	// heal, since all temporary max health bonuses also instantaneously heal the player.
	var healthMetrics *ResourceMetrics
	amountHealed := buffs[stats.Health]
	includesHealthBuff := amountHealed > 0

	buffedStatTypes := buffs.GetBuffedStatTypes()

	if includesHealthBuff {
		healthMetrics = character.NewHealthMetrics(actionID)
		buffs[stats.Health] = 0
	}

	config := Aura{
		Label:    auraLabel,
		ActionID: actionID,
		Duration: duration,
		OnGain: func(aura *Aura, sim *Simulation) {
			if sim.Log != nil {
				character.Log(sim, "Gained %s from %s.", buffs.FlatString(), actionID)
			}
			character.AddStatsDynamic(sim, buffs)

			if includesHealthBuff {
				character.UpdateMaxHealth(sim, amountHealed, healthMetrics)
			}

			for i := range character.OnTemporaryStatsChanges {
				character.OnTemporaryStatsChanges[i](sim, aura, buffs)
			}
		},
		OnExpire: func(aura *Aura, sim *Simulation) {
			if sim.Log != nil {
				character.Log(sim, "Lost %s from fading %s.", buffs.FlatString(), actionID)
			}
			character.AddStatsDynamic(sim, buffs.Invert())

			if includesHealthBuff {
				character.UpdateMaxHealth(sim, -amountHealed, healthMetrics)
			}

			for i := range character.OnTemporaryStatsChanges {
				character.OnTemporaryStatsChanges[i](sim, aura, buffs.Invert())
			}
		},
	}

	if modConfig != nil {
		modConfig(&config)
	}

	return &StatBuffAura{
		Aura:            character.GetOrRegisterAura(config),
		BuffedStatTypes: buffedStatTypes,
	}
}

// A stat an aura multiplies and the factor it multiplies it by.
type StatMultiplier struct {
	Stat       stats.Stat
	Multiplier float64
}

// A buff that multiplies stats through dynamic stat dependencies, turned on and off with the aura. The
// temporary stats listeners hear the stats the dependencies add on gain and remove on expire, measured
// at that moment.
func (character *Character) NewTemporaryStatMultiplierAura(config Aura, multipliers []StatMultiplier) *StatBuffAura {
	deps := make([]*stats.StatDependency, len(multipliers))
	buffed := make([]stats.Stat, len(multipliers))
	for i, m := range multipliers {
		deps[i] = character.NewDynamicMultiplyStat(m.Stat, m.Multiplier)
		buffed[i] = m.Stat
	}

	toggle := func(aura *Aura, sim *Simulation, set func(*Simulation, *stats.StatDependency)) {
		if len(character.OnTemporaryStatsChanges) == 0 {
			for _, dep := range deps {
				set(sim, dep)
			}
			return
		}

		before := character.GetStats()
		for _, dep := range deps {
			set(sim, dep)
		}

		change := character.GetStats().Subtract(before)
		for _, onChange := range character.OnTemporaryStatsChanges {
			onChange(sim, aura, change)
		}
	}
	config.OnGain = func(aura *Aura, sim *Simulation) {
		toggle(aura, sim, character.EnableBuildPhaseStatDep)
	}
	config.OnExpire = func(aura *Aura, sim *Simulation) {
		toggle(aura, sim, character.DisableBuildPhaseStatDep)
	}

	return &StatBuffAura{Aura: character.GetOrRegisterAura(config), BuffedStatTypes: buffed}
}

// Creates a new ProcTriggerAura that is dependent on a parent Aura being active
// This should only be used if the dependent Aura is:
// 1. On the a different Unit than parent Aura is registered to (usually the Character)
// 2. You need to register multiple dependent Aura's for the same Unit
func (parentAura *Aura) MakeDependentProcTriggerAura(unit *Unit, config ProcTrigger) *Aura {
	oldExtraCondition := config.ExtraCondition
	config.ExtraCondition = func(sim *Simulation, spell *Spell, result *SpellResult) bool {
		return parentAura.IsActive() && ((oldExtraCondition == nil) || oldExtraCondition(sim, spell, result))
	}

	aura := unit.MakeProcTriggerAura(config)

	return aura
}

// Attaches a ProcTrigger to a parent Aura
// Preffered use-case.
// For non standard use-cases see: MakeDependentProcTriggerAura
// Returns parent aura for chaining
func (parentAura *Aura) AttachProcTrigger(config ProcTrigger) *Aura {
	parentAura.AttachProcTriggerCallback(parentAura.Unit, config)

	return parentAura
}

// One charge per cast, however many targets it hits: first is true for a cast's first hit only. A
// cast's hits share its spell id and time, so a tagged off-hand strike counts with its main-hand one.
// reset forgets the last cast.
func OncePerCast() (first func(sim *Simulation, spell *Spell) bool, reset func()) {
	var recorded bool
	var lastID int32
	var lastAt time.Duration
	first = func(sim *Simulation, spell *Spell) bool {
		if recorded && spell.ActionID.SpellID == lastID && sim.CurrentTime == lastAt {
			return false
		}
		recorded, lastID, lastAt = true, spell.ActionID.SpellID, sim.CurrentTime
		return true
	}
	reset = func() {
		recorded = false
	}
	return first, reset
}

// Attaches a SpellMod to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachSpellMod(spellModConfig SpellModConfig) *Aura {
	parentAuraDep := parentAura.Unit.AddDynamicMod(spellModConfig)

	parentAura.ApplyOnGain(func(_ *Aura, _ *Simulation) {
		parentAuraDep.Activate()
	})

	parentAura.ApplyOnExpire(func(_ *Aura, _ *Simulation) {
		parentAuraDep.Deactivate()
	})

	return parentAura
}

// Attaches a StatDependency to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachStatDependency(statDep *stats.StatDependency) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.EnableBuildPhaseStatDep(sim, statDep)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.DisableBuildPhaseStatDep(sim, statDep)
	})

	if parentAura.IsActive() {
		parentAura.Unit.StatDependencyManager.EnableDynamicStatDep(statDep)
	}

	return parentAura
}

// Adds Stats to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachStatsBuff(stats stats.Stats) *Aura {
	parentAura.ApplyOnGain(func(aura *Aura, sim *Simulation) {
		aura.Unit.AddStatsDynamic(sim, stats)
	})

	parentAura.ApplyOnExpire(func(aura *Aura, sim *Simulation) {
		aura.Unit.AddStatsDynamic(sim, stats.Invert())
	})

	if parentAura.IsActive() {
		parentAura.Unit.AddStats(stats)
	}

	return parentAura
}

// Adds a Stat to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachStatBuff(stat stats.Stat, value float64) *Aura {
	statsToAdd := stats.Stats{}
	statsToAdd[stat] = value
	parentAura.AttachStatsBuff(statsToAdd)

	return parentAura
}

// Attaches a multiplicative PseudoStat buff to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachMultiplicativePseudoStatBuff(fieldPointer *float64, multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, _ *Simulation) {
		*fieldPointer *= multiplier
	})

	parentAura.ApplyOnExpire(func(_ *Aura, _ *Simulation) {
		*fieldPointer /= multiplier
	})

	if parentAura.IsActive() {
		*fieldPointer *= multiplier
	}

	return parentAura
}

// Attaches an additive PseudoStat buff to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachAdditivePseudoStatBuff(fieldPointer *float64, bonus float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, _ *Simulation) {
		*fieldPointer += bonus
	})

	parentAura.ApplyOnExpire(func(_ *Aura, _ *Simulation) {
		*fieldPointer -= bonus
	})

	if parentAura.IsActive() {
		*fieldPointer += bonus
	}

	return parentAura
}

// Attaches a ReducedCritTakenPercent buff that updates the base and recalculates totals.
// Returns parent aura for chaining.
func (parentAura *Aura) AttachReducedCritTakenPercentBuff(bonus float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, _ *Simulation) {
		parentAura.Unit.AddReducedCritTakenPercent(bonus)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, _ *Simulation) {
		parentAura.Unit.AddReducedCritTakenPercent(-bonus)
	})

	if parentAura.IsActive() {
		parentAura.Unit.AddReducedCritTakenPercent(bonus)
	}

	return parentAura
}

func (parentAura *Aura) AttachMultiplyCastSpeed(multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyCastSpeed(sim, multiplier)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyCastSpeed(sim, 1/multiplier)
	})

	return parentAura
}

func (parentAura *Aura) AttachMultiplyMeleeSpeed(multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyMeleeSpeed(sim, multiplier)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyMeleeSpeed(sim, 1/multiplier)
	})

	return parentAura
}

func (parentAura *Aura) AttachMultiplyWeaponSpeed(multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyWeaponSpeed(sim, multiplier)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyWeaponSpeed(sim, 1/multiplier)
	})

	return parentAura
}

func (parentAura *Aura) AttachMultiplyAttackSpeed(multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyAttackSpeed(sim, multiplier)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyAttackSpeed(sim, 1/multiplier)
	})

	return parentAura
}

func (parentAura *Aura) AttachMultiplyRangedSpeed(multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyRangedSpeed(sim, multiplier)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyRangedSpeed(sim, 1/multiplier)
	})

	return parentAura
}

func (parentAura *Aura) AttachMultiplyRangedHaste(multiplier float64) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyRangedHaste(sim, multiplier)
	})

	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		parentAura.Unit.MultiplyRangedHaste(sim, 1/multiplier)
	})

	return parentAura
}

// Attaches a Damage Done By Caster buff to a parent Aura
// Returns parent aura for chaining
func (parentAura *Aura) AttachDDBC(index int, maxIndex int, attackTables *[]*AttackTable, handler DynamicDamageDoneByCaster) *Aura {
	parentAura.ApplyOnGain(func(_ *Aura, _ *Simulation) {
		EnableDamageDoneByCaster(index, maxIndex, (*attackTables)[parentAura.Unit.UnitIndex], handler)
	})

	parentAura.ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		DisableDamageDoneByCaster(index, (*attackTables)[parentAura.Unit.UnitIndex])
	})

	if parentAura.IsActive() {
		EnableDamageDoneByCaster(index, maxIndex, (*attackTables)[parentAura.Unit.UnitIndex], handler)
	}

	return parentAura
}

func (parentAura *Aura) AttachPeriodicAction(config PeriodicActionOptions) *Aura {
	var pa *PendingAction
	tickImmediately := config.TickImmediately
	config.TickImmediately = false
	parentAura.ApplyOnGain(func(_ *Aura, sim *Simulation) {
		if pa == nil {
			pa = NewPeriodicAction(sim, config)
		}
		pa.cancelled = false
		pa.NextActionAt = sim.CurrentTime + TernaryDuration(tickImmediately, 0, config.Period)
		sim.AddPendingAction(pa)
	})
	parentAura.ApplyOnExpire(func(_ *Aura, sim *Simulation) {
		pa.Cancel(sim)
	})
	if parentAura.IsActive() {
		panic("Can't attach a periodic action to an active aura.")
	}
	return parentAura
}

type ShieldStrengthCalculator func(unit *Unit) float64
type OnDamageAbsorbedCallback func(sim *Simulation, aura *DamageAbsorptionAura, result *SpellResult, absorbedDamage float64)
type ShieldShouldApplyCondition func(sim *Simulation, spell *Spell, result *SpellResult, isPeriodic bool) bool

type AbsorptionAuraConfig struct {
	Aura                     Aura
	DamageMultiplier         float64
	MaxAbsorbPerHit          float64
	ShieldStrengthCalculator ShieldStrengthCalculator
	OnDamageAbsorbed         OnDamageAbsorbedCallback
	ShouldApplyToResult      ShieldShouldApplyCondition
}

type DamageAbsorptionAura struct {
	*Aura
	freshShieldStrengthCalculator ShieldStrengthCalculator

	ShieldStrength   float64
	OnDamageAbsorbed []OnDamageAbsorbedCallback
}

func (aura *DamageAbsorptionAura) Activate(sim *Simulation) {
	aura.Aura.Activate(sim)
	aura.ShieldStrength = aura.freshShieldStrengthCalculator(aura.Unit)
	stacks := max(1, int32(aura.ShieldStrength))
	aura.Aura.MaxStacks = stacks
	aura.Aura.SetStacks(sim, stacks)
}

func (unit *Unit) NewDamageAbsorptionAura(config AbsorptionAuraConfig) *DamageAbsorptionAura {
	if config.Aura.Duration == 0 {
		config.Aura.Duration = NeverExpires
	}

	aura := &DamageAbsorptionAura{
		Aura:                          unit.RegisterAura(config.Aura),
		freshShieldStrengthCalculator: config.ShieldStrengthCalculator,
	}

	if config.OnDamageAbsorbed != nil {
		aura.OnDamageAbsorbed = append(aura.OnDamageAbsorbed, config.OnDamageAbsorbed)
	}

	if config.DamageMultiplier == 0 {
		config.DamageMultiplier = 1
	}

	aura.ApplyOnExpire(func(_ *Aura, _ *Simulation) {
		aura.ShieldStrength = 0
	})

	extraSpellCheck := func(sim *Simulation, spell *Spell, result *SpellResult, isPeriodic bool) bool {
		return ((config.ShouldApplyToResult == nil) || config.ShouldApplyToResult(sim, spell, result, isPeriodic))
	}

	unit.AddDynamicDamageTakenModifier(func(sim *Simulation, spell *Spell, result *SpellResult, isPeriodic bool) {
		if aura.Aura.IsActive() && (result.Damage > 0) && extraSpellCheck(sim, spell, result, isPeriodic) {
			absorbedDamage := min(aura.ShieldStrength, result.Damage*config.DamageMultiplier)
			if config.MaxAbsorbPerHit > 0 {
				absorbedDamage = min(config.MaxAbsorbPerHit, absorbedDamage)
			}

			result.Damage -= absorbedDamage
			aura.ShieldStrength -= absorbedDamage

			if sim.Log != nil {
				unit.Log(sim, "%s absorbed %.1f damage, new shield strength: %.1f", aura.Label, absorbedDamage, aura.ShieldStrength)
			}

			for _, callback := range aura.OnDamageAbsorbed {
				callback(sim, aura, result, absorbedDamage)
			}

			if aura.ShieldStrength <= 0 {
				aura.Deactivate(sim)
				return
			}

			aura.Aura.SetStacks(sim, int32(aura.ShieldStrength))
		}
	})

	return aura
}

func (aura *DamageAbsorptionAura) AttachOnDamageAbsorbed(callback OnDamageAbsorbedCallback) *DamageAbsorptionAura {
	aura.OnDamageAbsorbed = append(aura.OnDamageAbsorbed, callback)

	return aura
}

type DamageAbsorptionAuraArray []*DamageAbsorptionAura

func (auras DamageAbsorptionAuraArray) Get(target *Unit) *DamageAbsorptionAura {
	if auras == nil {
		return nil
	}
	return auras[target.UnitIndex]
}

func (auras DamageAbsorptionAuraArray) IsEmpty() bool {
	for _, aura := range auras {
		if aura != nil {
			return false
		}
	}

	return true
}

func (auras DamageAbsorptionAuraArray) FindLabel() string {
	for _, aura := range auras {
		if aura != nil {
			return aura.Label
		}
	}

	panic("No valid damage absorption auras in array!")
}

func (caster *Unit) NewAllyDamageAbsorptionAuraArray(makeAura func(*Unit) *DamageAbsorptionAura) DamageAbsorptionAuraArray {
	auras := make([]*DamageAbsorptionAura, len(caster.Env.AllUnits))
	for _, target := range caster.Env.AllUnits {
		if target.Type != EnemyUnit {
			auras[target.UnitIndex] = makeAura(target)
		}
	}
	return auras
}

// TODO: Manual review needed -- the run speed and the callers' overshoot are the sim's movement model.
// The arrival callback runs only on reaching the target, never when the aura runs out or the iteration is torn down.
func MakeDashAura(unit *Unit, label string, actionID ActionID, duration time.Duration, onArrive func(sim *Simulation)) *Aura {
	aura := unit.RegisterAura(Aura{
		Label:    label,
		ActionID: actionID,
		Duration: duration,
		OnGain: func(_ *Aura, sim *Simulation) {
			unit.MultiplyMovementSpeed(sim, 3.0)
		},
		OnExpire: func(_ *Aura, sim *Simulation) {
			unit.MultiplyMovementSpeed(sim, 1.0/3.0)
		},
	})

	unit.RegisterMovementCallback(func(sim *Simulation, _ float64, kind MovementUpdateType) {
		if kind == MovementEnd && aura.IsActive() {
			aura.Deactivate(sim)
			if onArrive != nil {
				onArrive(sim)
			}
		}
	})

	return aura
}

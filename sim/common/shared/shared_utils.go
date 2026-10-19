package shared

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

type ProcStatBonusEffect struct {
	Name               string
	ItemID             int32
	EnchantID          int32
	MaxStacks          int32
	Callback           core.AuraCallback
	ProcMask           core.ProcMask
	Outcome            core.HitOutcome
	RequireDamageDealt bool
	ClassSpellsOnly    bool
	// See core.ProcTrigger.
	RequireDamageComponent bool
	// The listener's Can Proc From Procs attribute. See core.ProcTrigger.
	CanProcFromProcs bool
	// A "Chance on hit" item effect or a combat enchant. See core.ProcTrigger. The generator
	// reads it from the trigger type; a hand-written effect states it.
	IsWeaponProc bool
	// Carried through to the trigger. The generator sets SpellFlagSuppressWeaponProcs here for the
	// two auras marked Aura Is Weapon Proc.
	SpellFlagsExclude core.SpellFlag

	// What adds a stack while a stacking trinket's window is open. Derived from the container
	// spell's own proc flags, which are not the ones that open the window.
	// For example: Blackened Naaru Sliver opens on a melee hit
	// and then stacks on every attack for the next 20s.
	StackCallback core.AuraCallback
	StackProcMask core.ProcMask
	StackOutcome  core.HitOutcome

	// Any other custom proc conditions not covered by the above fields.
	CustomProcCondition core.CustomStatBuffProcCondition
}

type DamageEffect struct {
	SpellID          int32
	School           core.SpellSchool
	DefenseType      core.DefenseType // From SpellCategories. Left unset, it is inferred from School and IsMelee.
	MinDmg           float64
	MaxDmg           float64
	BonusCoefficient float64
	IsMelee          bool
	ProcMask         core.ProcMask
	Outcome          OutcomeType
	Flags            core.SpellFlag
	// Set when the game files bar the damage spell from critting, which the sim has no way to know:
	// it is a spell attribute, so only the database generator can see it.
	CannotCrit bool
}

type ExtraSpellInfo struct {
	Spell   *core.Spell
	Trigger func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult)
}

type ItemVariant struct {
	ItemID   int32
	ItemName string
}

type CustomProcHandler func(sim *core.Simulation, procAura *core.StatBuffAura)

// The DefenseType a proc's damage rolls with. A stated one (the generator reads it from the
// SpellCategories table) wins. Otherwise the school decides: anything non-physical rolls against the
// spell tables, and IsMelee stays honoured for a caller that means melee damage without saying so
// through the school.
func damageDefenseType(defenseType core.DefenseType, school core.SpellSchool, isMelee bool) core.DefenseType {
	if defenseType != core.DefenseTypeNone {
		return defenseType
	}

	if isMelee || school.Matches(core.SpellSchoolPhysical) {
		return core.DefenseTypeMelee
	}

	return core.DefenseTypeMagic
}

// The outcome a proc's damage rolls when the caller states none: the hit table of its DefenseType,
// in the no-crit variant when the game files bar the spell from critting.
func damageOutcome(defenseType core.DefenseType, cannotCrit bool, outcome OutcomeType) OutcomeType {
	if outcome != OutcomeDefault {
		return outcome
	}

	switch defenseType {
	case core.DefenseTypeMelee:
		if cannotCrit {
			return OutcomeMeleeNoCrit
		}

		return OutcomeMeleeCanCrit
	case core.DefenseTypeRanged:
		if cannotCrit {
			return OutcomeRangedNoCrit
		}

		return OutcomeRangedCanCrit
	}

	if cannotCrit {
		return OutcomeSpellNoCrit
	}

	return OutcomeSpellCanCrit
}

func NewProcStatBonusEffectWithDamageProc(config ProcStatBonusEffect, damage DamageEffect) {
	procMask := core.ProcMaskEmpty
	if damage.ProcMask != core.ProcMaskUnknown {
		procMask = damage.ProcMask
	}

	factory_StatBonusEffect(config, func(agent core.Agent) ExtraSpellInfo {
		character := agent.GetCharacter()

		defenseType := damageDefenseType(damage.DefenseType, damage.School, damage.IsMelee)
		procSpell := character.RegisterSpell(core.SpellConfig{
			ActionID:                 core.ActionID{SpellID: damage.SpellID},
			SpellSchool:              damage.School,
			DefenseType:              defenseType,
			ProcMask:                 procMask,
			Flags:                    core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,
			DamageMultiplier:         1,
			DamageMultiplierAdditive: 1,
			ThreatMultiplier:         1,
			BonusCoefficient:         damage.BonusCoefficient,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(damage.MinDmg, damage.MaxDmg), GetOutcome(spell, damageOutcome(defenseType, damage.CannotCrit, damage.Outcome)))
			},
		})

		return ExtraSpellInfo{
			Spell: procSpell,
			Trigger: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				procSpell.Cast(sim, result.Target)
			},
		}
	})
}

func factory_StatBonusEffect(config ProcStatBonusEffect, extraSpell func(agent core.Agent) ExtraSpellInfo) {
	// Ignore empty dummy implementations
	if config.Callback == core.CallbackEmpty {
		return
	}

	source := config.effectSource()

	// Soft fail to allow for overrides for bad effects
	if source.isAlreadyImplemented() {
		return
	}

	triggerActionID := source.actionID()

	source.registerEffect(func(agent core.Agent) {
		character := agent.GetCharacter()
		eligibleSlots := source.eligibleSlots(character)

		procEffects := source.procEffects()
		if len(procEffects) == 0 {
			panic(fmt.Sprintf("Error getting proc effects for item/enchant %v", source.id))
		}

		for _, effect := range procEffects {
			proc := effect.GetProc()

			// windowAura is set only for the stacking trinkets, where the trigger opens a window
			// that accumulates a separate stat aura. The handler then activates the window rather
			// than the stat aura, so a re-proc restarts the window instead of refreshing a duration
			// the game does not refresh when a stack lands.
			procAura, windowAura := buildProcAura(character, config, effect)

			dpm := procDPM(character, config, source, proc)

			procAura.CustomProcCondition = config.CustomProcCondition

			var procSpell ExtraSpellInfo
			if extraSpell != nil {
				procSpell = extraSpell(agent)
			}

			triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
				ActionID:           triggerActionID,
				Name:               config.Name,
				Callback:           config.Callback,
				ProcMask:           config.ProcMask,
				SpellFlagsExclude:  config.SpellFlagsExclude,
				Outcome:            config.Outcome,
				RequireDamageDealt: config.RequireDamageDealt,
				ClassSpellsOnly:    config.ClassSpellsOnly,
				CanProcFromProcs:   config.CanProcFromProcs,
				IsWeaponProc:       config.IsWeaponProc,
				ProcChance:         proc.GetProcChance(),
				DPM:                dpm,
				ICD:                time.Millisecond * time.Duration(proc.IcdMs),
				Handler:            procHandler(config, effect, procAura, windowAura, procSpell),
				ActivatesAura:      core.Ternary(windowAura != nil, windowAura, procAura.Aura),

				RequireDamageComponent: config.RequireDamageComponent,
			})

			attachStackTrigger(character, config, effect, procAura, windowAura)

			// Carried on the stacking path too. Nothing in the stacking machinery reads this -
			// CanProc consults only IsSwapped and CustomProcCondition - so it gates nothing. What
			// it feeds is GetMatchingItemProcAuras, which drops any aura whose Icd is nil, and with
			// it every ICD-aware APL value. Hand-written trinkets all set it, so leaving it nil
			// would hide the generated ones from those APLs.
			if proc.IcdMs != 0 {
				procAura.Icd = triggerAura.Icd
			}

			source.registerProc(character, triggerAura, eligibleSlots)
			source.registerWeaponEnchantBuff(character, procAura)
			character.AddStatProcBuff(source.id, procAura, source.isEnchant, eligibleSlots)
		}
	})
}

// What the proc does when it fires. When a custom condition refuses, the ICD is rolled back so the
// next opportunity still counts instead of the effect being locked out by a proc that never
// happened.
func procHandler(config ProcStatBonusEffect, effect *proto.ItemEffect, procAura *core.StatBuffAura, windowAura *core.Aura, procSpell ExtraSpellInfo) func(*core.Simulation, *core.Spell, *core.SpellResult) {
	return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		// A custom condition gates the body rather than replacing it. Written as its own branch it
		// skipped the window, the stack accumulation and the extra spell, so any override set on an
		// item whose database entry resolves a stacking aura would open nothing.
		if config.CustomProcCondition != nil && !procAura.CanProc(sim) {
			if procAura.Icd != nil && procAura.Icd.Duration != 0 {
				procAura.Icd.Reset()
			}

			return
		}

		// Activating the window and not the stat aura is what makes a re-proc restart the window
		// instead of refreshing stacks the game would not refresh.
		if windowAura != nil {
			windowAura.Activate(sim)
		} else {
			procAura.Activate(sim)
			if effect.MaxCumulativeStacks > 0 {
				procAura.AddStack(sim)
			}
		}

		if procSpell.Spell != nil {
			procSpell.Trigger(sim, spell, result)
		}
	}
}

// The three shapes a proc buff comes in. Only the first returns a second aura: there the trigger
// opens a window and the stat aura inside it accumulates, so the caller has two things to wire.
func buildProcAura(character *core.Character, config ProcStatBonusEffect, effect *proto.ItemEffect) (*core.StatBuffAura, *core.Aura) {
	label := config.Name + " Proc"
	action := core.ActionID{SpellID: effect.BuffId}
	duration := time.Millisecond * time.Duration(effect.EffectDurationMs)

	if stackingAura := effect.StackingAura; stackingAura != nil {
		return character.NewTemporaryStatBuffWithStacks(core.TemporaryStatBuffWithStacksConfig{
			AuraLabel:            label,
			ActionID:             action,
			Duration:             duration,
			MaxStacks:            stackingAura.MaxCumulativeStacks,
			BonusPerStack:        stats.FromProtoMap(stackingAura.GetScalingOptions()[int32(0)].GetStats()),
			StackingAuraActionID: core.ActionID{SpellID: stackingAura.BuffId},
			StackingAuraLabel:    config.Name + " Stacks",
			TimePerStack:         time.Millisecond * time.Duration(effect.GetStackPeriodMs()),
			TickImmediately:      true,
			StacksFromEvent:      effect.GetStackProc() != nil,
		})
	}

	if effect.MaxCumulativeStacks > 0 {
		return core.MakeStackingAura(character, core.StackingStatAura{
			Aura: core.Aura{
				Label:     label,
				ActionID:  action,
				Duration:  duration,
				MaxStacks: effect.MaxCumulativeStacks,
			},
			BonusPerStack: stats.FromProtoMap(effect.GetScalingOptions()[int32(0)].GetStats()),
		}), nil
	}

	return character.NewTemporaryStatsAura(label, action, stats.FromProtoMap(effect.GetScalingOptions()[int32(0)].GetStats()), duration), nil
}

func procDPM(character *core.Character, config ProcStatBonusEffect, source effectSource, proc *proto.ProcEffect) *core.DynamicProcManager {
	return dpmForMask(character, source, proc.GetPpm(), config.ProcMask)
}

func dpmForMask(character *core.Character, source effectSource, ppm float64, mask core.ProcMask) *core.DynamicProcManager {
	if ppm <= 0 {
		return nil
	}

	if mask != core.ProcMaskUnknown {
		// A procs-per-minute enchant rolls on weapon hits only: spells and heals never proc it.
		if source.isEnchant {
			mask &= core.ProcMaskMeleeOrRanged
		}
		if source.enchantPlacement() == enchantOnWeapon {
			return character.NewDynamicLegacyProcForEnchantWithMask(source.id, ppm, mask)
		}
		return character.NewLegacyPPMManager(ppm, mask)
	}

	// With no mask of its own the rate has to be read off whatever the effect sits on.
	if source.isEnchant {
		return character.NewDynamicLegacyProcForEnchant(source.id, ppm, 0)
	}

	return character.NewDynamicLegacyProcForWeapon(source.id, ppm, 0)
}

// Event-driven stacks come from their own trigger: the container's proc flags decide what counts,
// and it only does anything while the window is open. A timer-driven stacking aura fills itself and
// needs none of this, so this is a no-op for everything else.
func attachStackTrigger(character *core.Character, config ProcStatBonusEffect, effect *proto.ItemEffect, statAura *core.StatBuffAura, windowAura *core.Aura) {
	stackProc := effect.GetStackProc()
	if stackProc == nil || windowAura == nil || config.StackCallback == core.CallbackEmpty {
		return
	}

	// Attached to the window rather than registered as its own aura: it is then only live while
	// the window is open, needs no active check, and cannot outlive the item the way a permanent
	// trigger would across an item swap.
	// The container spell's proc-ness rules are not read separately; the stack trigger follows the
	// opener's. No stacking item in the database differs between the two.
	windowAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
		Name:              config.Name + " Stack Trigger",
		Callback:          config.StackCallback,
		ProcMask:          config.StackProcMask,
		Outcome:           config.StackOutcome,
		SpellFlagsExclude: config.SpellFlagsExclude,
		CanProcFromProcs:  config.CanProcFromProcs,
		IsWeaponProc:      config.IsWeaponProc,
		ProcChance:        stackProc.GetProcChance(),
		DPM:               dpmForMask(character, config.effectSource(), stackProc.GetPpm(), config.StackProcMask),
		ICD:               time.Millisecond * time.Duration(stackProc.IcdMs),
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if !statAura.IsActive() {
				return
			}
			statAura.AddStack(sim)
		},
	})
}

// Registers the same effect once per item that carries it.
func forEachVariant(config ProcStatBonusEffect, variants []ItemVariant, register func(config ProcStatBonusEffect)) {
	for _, variant := range variants {
		config.Name = variant.ItemName
		config.ItemID = variant.ItemID
		register(config)
	}
}

func NewProcStatBonusEffectWithVariants(config ProcStatBonusEffect, variants []ItemVariant) {
	forEachVariant(config, variants, NewProcStatBonusEffect)
}

func NewProcStatBonusEffect(config ProcStatBonusEffect) {
	factory_StatBonusEffect(config, nil)
}

///////////////////////////////////////////////////////////////////////////
//							Procs read from the spell data
///////////////////////////////////////////////////////////////////////////

// An item or enchant proc as the game files' own rows state it. What the listener hears, how often it
// fires and what its internal cooldown is come from the spell carrying the proc; how long the buff
// lasts and how it stacks come from the spell it applies; only the stats come from the item's
// effect entry, which is where the sim's item level scaling lives.
type SpellDataProc struct {
	// Filled per variant by NewSpellDataProc. An enchant, which has no variants, states its own.
	Name   string
	ItemID int32
	// Set for an enchant, which registers through the enchant registry rather than the item one.
	EnchantID int32

	// The spell the item effect names: the one carrying the proc flags, the chance and the ICD.
	TriggerSpellID int32
	// The buff the proc applies, where the game files make it a spell of its own. Zero where the
	// trigger's own row is the buff.
	BuffSpellID int32

	// A "Chance on hit" item effect or a combat enchant. The game casts those off the hit itself
	// without consulting a proc mask, so the row states no listener and the shape is stated here.
	IsWeaponProc bool
}

// Registers the same effect once per item that carries it, the way the stat-bonus constructors do it.
func NewSpellDataProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataProc)
}

// An item or enchant proc whose "buff" is a damage spell: the game files apply no aura at all, it
// casts a spell that deals damage. BuffSpellID names that spell.
func NewSpellDataDamageProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataDamageProc)
}

// An item or enchant proc whose "buff" heals the wearer: it casts an E_HEAL_PCT or E_HEAL spell, or
// one applying an A_PERIODIC_HEAL aura to the wearer. BuffSpellID names that spell.
func NewSpellDataHealProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataHealProc)
}

func forEachSpellDataVariant(cfg SpellDataProc, variants []ItemVariant, register func(SpellDataProc)) {
	if len(variants) == 0 {
		register(cfg)
		return
	}

	for _, variant := range variants {
		cfg.Name = variant.ItemName
		cfg.ItemID = variant.ItemID
		register(cfg)
	}
}

// An aura the row applies to an enemy is NewSpellDataDebuffProc's to build, never a buff on the
// wearer, so the effect is left unregistered rather than handed to the wrong unit.
func registerSpellDataProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, (*spelldata.Spell).AppliesAnAuraToAnEnemy, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, buff *spelldata.Spell) {
		applySpellDataProc(agent, cfg, source, trigger, buff)
	})
}

// The effect on one character: the buff it applies, the listener that applies it, and the
// registrations that let an item swap and an APL find both.
func applySpellDataProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, buff *spelldata.Spell) {
	character := agent.GetCharacter()
	eligibleSlots := source.eligibleSlots(character)

	effect := source.procEntry(cfg, trigger)
	if effect == nil {
		panic(fmt.Sprintf("Error getting proc effects for item/enchant %v", source.id))
	}

	procAura := spellDataProcAura(character, cfg, trigger, buff, effect)

	listener := spellDataProcListener(character, cfg, source, trigger)
	listener.Handler = spellDataProcHandler(buff, procAura)
	listener.ActivatesAura = procAura.Aura

	// The same fallback the database layer applies: Bulwark of Azzinoth's armor buff sits in a
	// spell category with a 60s recovery while its trigger states nothing. Only a buff that is a
	// spell of its own counts - where the two are one spell the recovery is that spell's own cast
	// throttle rather than a gate on re-applying the buff.
	if listener.ICD == 0 && buff.ID != trigger.ID {
		listener.ICD = buff.CategoryCooldown()
	}
	triggerAura := source.registerTrigger(character, listener)

	attachChargeSpender(character, cfg, trigger, buff, procAura)

	// See factory_StatBonusEffect: this is what keeps the ICD-aware APL values from dropping the
	// effect, not a gate on the proc. It is assigned after the charge spender and whether or not
	// there is one, since the lockout an APL asks after is the proc's rather than the one between
	// two charges, which attaching the spender would otherwise leave here.
	procAura.Icd = triggerAura.Icd

	source.registerWeaponEnchantBuff(character, procAura)
	character.AddStatProcBuff(source.id, procAura, source.isEnchant, eligibleSlots)
}

// The listener the trigger's row describes, without the handler, plus what the row cannot state: the
// item's own name and action, which are what the sim keys the rolls and the metrics by, and the rate
// the item's effect entry states.
func spellDataProcListener(character *core.Character, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell) core.ProcTrigger {
	config := spelldata.ProcTrigger(character, trigger, nil, spelldata.ItemProcChance(trigger),
		weaponProcShape(cfg), spellDataProcRate(source, trigger, source.procEntry(cfg, trigger).GetProc()),
		statedWeaponProcChance(cfg.IsWeaponProc, source))
	config.Name = cfg.Name
	config.ActionID = source.actionID()

	return config
}

// A weapon proc's listener, which no row states; anything else keeps the one its row decodes to.
func weaponProcShape(cfg SpellDataProc) spelldata.ProcOpt {
	if !cfg.IsWeaponProc {
		return func(*core.Character, *core.ProcTrigger) {}
	}
	return spelldata.WeaponProc()
}

// The rate for a proc the game files state none for. Procs per minute are not in the game files' spell
// data - no row carries a SpellProcsPerMinuteID - so an item's reaches the sim through its effect
// entry, and the store carries one only where an override put it there. Either way the manager is
// built here rather than by the resolver, since only the effect knows which weapon or enchant slot
// the rate has to be measured against.
func spellDataProcRate(source effectSource, row *spelldata.Spell, proc *proto.ProcEffect) spelldata.ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		ppm := proc.GetPpm()
		if ppm == 0 {
			ppm = float64(row.RPPM)
		}

		dpm := dpmForMask(character, source, ppm, trigger.ProcMask)
		if dpm == nil {
			return
		}

		// The callback rolls the chance first and the manager only after it, so a chance left in
		// place next to a manager would gate the rate twice.
		trigger.ProcChance = 0
		trigger.DPM = dpm
	}
}

// A weapon proc's flat chance, rolled on the hits of the weapon carrying the item or the enchant only:
// the weapon shape hears every hit, so a chance left on the trigger would also roll on the other
// hand's. The manager rolls the same chance on the hands that carry it and follows them through an
// item swap. A rate measured in procs per minute already has its manager, and keeps it.
func statedWeaponProcChance(isWeaponProc bool, source effectSource) spelldata.ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		if !isWeaponProc || trigger.DPM != nil || trigger.ProcChance == 0 {
			return
		}

		chance := trigger.ProcChance
		trigger.ProcChance = 0
		if source.isEnchant {
			trigger.DPM = character.NewDynamicLegacyProcForEnchant(source.id, 0, chance)
		} else {
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(source.id, 0, chance)
		}
	}
}

// What the proc does when it fires: it applies the buff, adds a stack where the buff accumulates
// them, and hands a buff the game spends by charges its full count to spend.
func spellDataProcHandler(buff *spelldata.Spell, procAura *core.StatBuffAura) core.ProcHandler {
	return func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		procAura.Activate(sim)

		switch {
		case buff.MaxStack > 0:
			procAura.AddStack(sim)
		case buff.ProcCharges > 0:
			procAura.SetStacks(sim, int32(buff.ProcCharges))
		}
	}
}

// The buff the proc applies. Which shape it takes is the game files' to say, and the two counts it
// keeps in one field are not the same thing: a CumulativeAura count is stacks that each add their
// own stats, a ProcCharges count is one buff at full stats that the game spends by uses.
func spellDataProcAura(character *core.Character, cfg SpellDataProc, trigger *spelldata.Spell, buff *spelldata.Spell, effect *proto.ItemEffect) *core.StatBuffAura {
	// A trinket whose trigger opens a window and whose stats accumulate on a second aura inside it
	// resolves no stats on the aura the trigger applies, so building one here would grant nothing at
	// all. That shape needs the window machinery in factory_StatBonusEffect.
	if effect.GetStackingAura() != nil {
		panic(fmt.Sprintf("%s (%d): a proc whose stats live on an accumulating aura needs the stacking constructor", cfg.Name, cfg.ItemID))
	}

	aura := spelldata.AuraConfig(buff, spelldata.Label(cfg.Name+" Proc"))
	aura.Duration = procBuffDuration(cfg, trigger, buff)

	buffStats := stats.FromProtoMap(effect.GetScalingOptions()[int32(0)].GetStats())

	// The game files state the count on whichever of the two rows carries the aura, and the item
	// effect entry takes the higher of them, so this does too: Idol of the Huntress keeps its 200
	// on the trigger while the buff it applies states none.
	if stacks := max(buff.MaxStack, trigger.MaxStack); stacks > 0 {
		aura.MaxStacks = int32(stacks)
		return core.MakeStackingAura(character, core.StackingStatAura{Aura: aura, BonusPerStack: buffStats})
	}

	// Charges are the buff's own: they count how many times *it* acts before it drops. The trigger's,
	// where it has any, count how many times the trigger fires, which is a different thing and not
	// this aura's business.
	aura.MaxStacks = int32(buff.ProcCharges)

	return character.NewTemporaryStatsAuraWrapped(aura.Label, aura.ActionID, buffStats, aura.Duration, func(config *core.Aura) {
		config.MaxStacks = aura.MaxStacks
	})
}

// How long the buff lasts. The game files leave it off the buff's own row on a fair few procs and
// state it on the trigger instead, which is the fallback the item effect entry applies as well. An
// aura of no duration is one core refuses to activate mid-fight, so a pair of rows that states none
// anywhere is refused here, where the message can name what to look at.
func procBuffDuration(cfg SpellDataProc, trigger *spelldata.Spell, buff *spelldata.Spell) time.Duration {
	if buff.DurationMs != 0 {
		return buff.Duration()
	}

	if trigger.DurationMs != 0 {
		return trigger.Duration()
	}

	panic(fmt.Sprintf("%s (%d): neither the proc's spell %d nor its buff %d states a duration for the aura it applies",
		cfg.Name, cfg.effectSource().id, trigger.ID, buff.ID))
}

// What spends a charge. The buff's own row states which hits do - Lightning Shield's three charges
// go to the melee hits it answers - so it is read as a listener of its own and attached to the buff,
// where it is live only while the buff is up. A buff whose row states no listener or no rate keeps
// its charges unspent and runs out its duration instead.
//
// Only a buff that is a spell of its own can say what spends a charge. Where the buff is the trigger,
// the flags on the row are the ones that granted the buff, so a spender built from them would take a
// charge back on the very hit that handed them over and the count would never run down.
func attachChargeSpender(character *core.Character, cfg SpellDataProc, trigger *spelldata.Spell, buff *spelldata.Spell, procAura *core.StatBuffAura) {
	if buff.ID == trigger.ID || buff.ProcCharges <= 0 || buff.MaxStack > 0 || !statesATrigger(buff) {
		return
	}

	spelldata.AttachChargeSpender(character, procAura.Aura, buff, cfg.Name+" Charge")
}

// Whether a trigger can be built from this row at all: it has to name hits the sim hears and a rate
// that resolves to something. The conditions are the ones spelldata.ProcTrigger panics on, since the
// point of asking is to not reach that panic for a row nobody stated a rate for.
func statesATrigger(s *spelldata.Spell) bool {
	decoded := core.DecodeProcTypeMask(s.ProcFlags, s.ProcHint)
	if decoded.Callback == core.CallbackEmpty {
		return false
	}

	// A procs-per-minute rate is measured against the mask, which an empty one cannot do.
	if s.RPPM > 0 {
		return decoded.ProcMask != core.ProcMaskUnknown
	}
	return s.StatedChance() != 0
}

func registerSpellDataDamageProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, nil, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, damage *spelldata.Spell) {
		applySpellDataDamageProc(agent, cfg, source, trigger, damage)
	})
}

// The effect on one character: the spell the proc casts, and the listener that casts it.
func applySpellDataDamageProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, damage *spelldata.Spell) {
	character := agent.GetCharacter()
	config := spellDataProcListener(character, cfg, source, trigger)
	attachProcDamage(character, &config, spellDataProcDamageSpell(character, damage, true))
	source.registerTrigger(character, config)
}

// The spell the proc casts, as its row states it: school, defense type, spell power share, travel
// time and the amount it rolls. What the row cannot state is that it is a proc's spell: out of the
// rotation and not a cast of its own. Its hits reach the damage listeners as a proc's do, so a
// listener with Can Proc From Procs hears them (Vengeance off a Lightning Strike crit).
//
// Used from an item instead, a row that states Always Hit rolls no miss, and crits the way damage to
// its own caster does (casterDamageOutcome): Dragonfire Potion's 1252554 always lands.
func spellDataProcDamageSpell(character *core.Character, damage *spelldata.Spell, asProc bool) core.SpellConfig {
	config := spelldata.SpellConfig(&character.Unit, damage, damageShape(damage), castBy(asProc))

	defenseType := damageDefenseType(config.DefenseType, config.SpellSchool, false)
	config.DefenseType = defenseType
	outcome := damageOutcome(defenseType, damage.CannotCrit(), OutcomeDefault)
	application := damageOutcome(defenseType, true, OutcomeDefault)
	alwaysHits := !asProc && damage.AlwaysHits()
	effect := damage.DamageEffect()
	config.DirectVariance = effect.DirectVariance()
	// A split row rolls once for every target it reaches (calcMultiTargetDamage), not once a hit.
	if damage.SplitsDamage {
		config.DirectVariance = 0
	}

	// Bound once per spell: binding the outcomes per cast would allocate.
	var boundSpell *core.Spell
	var hitOutcome, applicationOutcome core.OutcomeApplier
	bind := func(spell *core.Spell) {
		if boundSpell == spell {
			return
		}
		boundSpell = spell
		if alwaysHits {
			hitOutcome, applicationOutcome = casterDamageOutcome(spell, damage.CannotCrit()), spell.OutcomeAlwaysHit
			return
		}
		hitOutcome, applicationOutcome = GetOutcome(spell, outcome), GetOutcome(spell, application)
	}

	single := make(core.SpellResultSlice, 1)
	debuff := debuffOnLanding(character, damage)
	periodic := damage.PeriodicDamageEffect()
	if periodic == spelldata.NilEffect {
		multiTarget := effect.HitsAnArea() || effect.ChainTargets > 1
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			bind(spell)
			if multiTarget {
				dealOnArrival(sim, spell, target, calcMultiTargetDamage(sim, spell, target, damage, effect, hitOutcome), debuff)
				return
			}
			single[0] = spell.CalcDamage(sim, target, effect.DirectAmount(sim), hitOutcome)
			dealOneOnArrival(sim, spell, target, single, debuff)
		}
		return config
	}

	// The damage over time lands only with the hit that applies it, which is its one hit roll: the
	// direct hit where the row deals one, else a roll of its own on the hit table of the spell's
	// defense type, without the crit an application cannot deal. The hit and the damage over time
	// each reach every active enemy where their own effect covers an area (the row's MaxTargets of
	// them from the target out, where it states one), else the target alone. An enemy the damage over
	// time reaches and the hit does not rolls its own application.
	config.Dot = spelldata.DotConfig(damage, periodic)
	areaHit, areaDot := effect.HitsAnArea(), periodic.HitsAnArea()
	after := applyDotIfLanded(areaDot)
	if debuff != nil {
		applyDot := after
		after = func(sim *core.Simulation, spell *core.Spell, target *core.Unit, results core.SpellResultSlice) {
			applyDot(sim, spell, target, results)
			debuff(sim, spell, target, results)
		}
	}
	var rolls core.SpellResultSlice
	if areaDot && !areaHit {
		rolls = make(core.SpellResultSlice, 0, len(character.Env.AllUnits))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		bind(spell)
		switch {
		case areaHit:
			dealOnArrival(sim, spell, target, calcMultiTargetDamage(sim, spell, target, damage, effect, hitOutcome), after)
			return
		case areaDot:
			rolls = rolls[:0]
			if effect != spelldata.NilEffect {
				rolls = append(rolls, spell.CalcDamage(sim, target, effect.DirectAmount(sim), hitOutcome))
			}
			reached := sim.Environment.ActiveTargetCount()
			if damage.MaxTargets > 0 {
				reached = min(reached, int32(damage.MaxTargets))
			}
			enemy := target
			if !enemy.IsEnabled() {
				enemy = sim.Environment.NextActiveTargetUnit(enemy)
			}
			for range reached {
				if effect == spelldata.NilEffect || enemy != target {
					rolls = append(rolls, spell.CalcOutcome(sim, enemy, applicationOutcome))
				}
				enemy = sim.Environment.NextActiveTargetUnit(enemy)
			}
			dealOnArrival(sim, spell, target, rolls, after)
			return
		case effect == spelldata.NilEffect:
			single[0] = spell.CalcOutcome(sim, target, applicationOutcome)
		default:
			single[0] = spell.CalcDamage(sim, target, effect.DirectAmount(sim), hitOutcome)
		}
		dealOneOnArrival(sim, spell, target, single, after)
	}

	return config
}

type afterDealt func(sim *core.Simulation, spell *core.Spell, target *core.Unit, results core.SpellResultSlice)

// Deals the results the cast calculated, then runs after where it is set: at once, or, where the
// spell has a travel time, once it has landed. The flight carries a copy of the results, since the
// spell's next cast calculates into the same ones.
func dealOnArrival(sim *core.Simulation, spell *core.Spell, target *core.Unit, results core.SpellResultSlice, after afterDealt) {
	if spell.TravelTime() == 0 {
		dealResults(sim, spell, target, results, after)
		return
	}
	results = slices.Clone(results)
	for _, result := range results {
		spell.DealCastResult(sim, result)
	}
	suppress := spell.Flags & core.SpellFlagSuppressWeaponProcs
	spell.WaitTravelTime(sim, func(sim *core.Simulation) {
		withFlags(spell, suppress, func() { dealResults(sim, spell, target, results, after) })
	})
}

// The same for the one result the cast calculated into single, which the spell's next cast calculates
// into again: the flight carries the result itself, and the cast's Suppress Weapon Procs with it.
func dealOneOnArrival(sim *core.Simulation, spell *core.Spell, target *core.Unit, single core.SpellResultSlice, after afterDealt) {
	result := single[0]
	if spell.TravelTime() == 0 {
		dealOne(sim, spell, target, single, result, after)
		return
	}
	spell.DealCastResult(sim, result)
	suppress := spell.Flags & core.SpellFlagSuppressWeaponProcs
	spell.WaitTravelTime(sim, func(sim *core.Simulation) {
		withFlags(spell, suppress, func() { dealOne(sim, spell, target, single, result, after) })
	})
}

// A cast during the flight, or one the landing hit asks for from inside DealDamage, calculates into
// single, so the result goes back into it before after reads it.
func dealOne(sim *core.Simulation, spell *core.Spell, target *core.Unit, single core.SpellResultSlice, result *core.SpellResult, after afterDealt) {
	spell.DealDamage(sim, result)
	if after != nil {
		single[0] = result
		after(sim, spell, target, single)
	}
}

func dealResults(sim *core.Simulation, spell *core.Spell, target *core.Unit, results core.SpellResultSlice, after afterDealt) {
	for _, result := range results {
		spell.DealDamage(sim, result)
	}
	if after != nil {
		after(sim, spell, target, results)
	}
}

// The damage over time goes on with the roll that carries it: over an area, on every enemy whose roll
// landed and that is still active when it lands; else on the target alone, where its roll landed.
func applyDotIfLanded(area bool) afterDealt {
	return func(sim *core.Simulation, spell *core.Spell, target *core.Unit, results core.SpellResultSlice) {
		for _, result := range results {
			if result.Landed() && (area && result.Target.IsEnabled() || !area && result.Target == target) {
				spell.Dot(result.Target).Apply(sim)
			}
		}
	}
}

// The damage of an effect that reaches more than its target, dealt as a batch: every enemy in
// its area, or the row's MaxTargets of them from the target out; or a chain of ChainTargets keeping
// ChainAmp of the damage at each jump. A split row rolls once and divides the roll evenly among the
// targets it reaches. Otherwise each target rolls its own, and an uncapped area takes the encounter's
// AoE cap the way an explosive does.
func calcMultiTargetDamage(sim *core.Simulation, spell *core.Spell, target *core.Unit, row *spelldata.Spell, effect *spelldata.Effect, outcome core.OutcomeApplier) core.SpellResultSlice {
	roll := func(sim *core.Simulation, _ *core.Spell) float64 {
		return effect.DirectAmount(sim)
	}

	if !effect.HitsAnArea() {
		keep := 1.0
		return spell.CalcCleaveDamageWithVariance(sim, target, int32(effect.ChainTargets), outcome, func(sim *core.Simulation, spell *core.Spell) float64 {
			damage := roll(sim, spell) * keep
			keep *= float64(effect.ChainAmp)
			return damage
		})
	}

	capped := row.MaxTargets > 0
	if row.SplitsDamage {
		reached := sim.Environment.ActiveTargetCount()
		if capped {
			reached = min(reached, int32(row.MaxTargets))
		}
		share := effect.Roll(sim) / float64(reached)
		if capped {
			return spell.CalcCleaveDamage(sim, target, int32(row.MaxTargets), share, outcome)
		}
		return spell.CalcAoeDamage(sim, share, outcome)
	}

	if capped {
		return spell.CalcCleaveDamageWithVariance(sim, target, int32(row.MaxTargets), outcome, roll)
	}
	return spell.CalcAoeDamageWithVariance(sim, outcome, func(sim *core.Simulation, spell *core.Spell) float64 {
		return roll(sim, spell) * sim.Encounter.AOECapMultiplier()
	})
}

// Which multipliers and metrics bucket the damage belongs in. The row's defense type decides, since
// that is what picks its hit table: a physical proc is a melee one whatever cast it in.
func damageShape(damage *spelldata.Spell) spelldata.SpellOpt {
	if damageDefenseType(damage.DefenseTypeCore(), damage.SpellSchool(), false) == core.DefenseTypeMelee {
		return spelldata.Melee(core.ProcMaskEmpty)
	}

	return spelldata.Magic(core.ProcMaskEmpty)
}

// What a proc's damage lands on, for both damage constructors. A cast the spell's own hit asks for
// while the spell is still casting, a weapon's strike rolling its own weapon again, goes one spell
// batch later: inline it would recurse for as long as the rolls succeed, without end where the
// chance cannot fail. A melee strike an off-hand hit cast carries Suppress Weapon Procs for that cast
// (weaponProcStrikeMask).
func procDamageHandler(character *core.Character, damageSpell *core.Spell, callback core.AuraCallback) core.ProcHandler {
	meleeStrike := damageSpell.ProcMask.Matches(core.ProcMaskMeleeMHSpecial)
	casting := false
	var cast func(sim *core.Simulation, target *core.Unit, flags core.SpellFlag)
	cast = func(sim *core.Simulation, target *core.Unit, flags core.SpellFlag) {
		if casting {
			pa := sim.GetConsumedPendingActionFromPool()
			pa.NextActionAt = sim.CurrentTime + core.SpellBatchWindow
			pa.Priority = core.ActionPriorityDOT
			pa.OnAction = func(sim *core.Simulation) { cast(sim, target, flags) }
			sim.AddPendingAction(pa)
			return
		}
		casting = true
		withFlags(damageSpell, flags, func() { damageSpell.Cast(sim, target) })
		casting = false
	}
	return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		flags := core.SpellFlagNone
		if meleeStrike && spell.ProcMask.Matches(core.ProcMaskMeleeOH) && !spell.ProcMask.Matches(core.ProcMaskMeleeMH) {
			flags = core.SpellFlagSuppressWeaponProcs
		}
		cast(sim, procDamageTarget(character, callback, spell, result), flags)
	}
}

// Runs deal with the flags added to the spell's own, and takes off again the ones it did not carry.
func withFlags(spell *core.Spell, flags core.SpellFlag, deal func()) {
	added := flags &^ spell.Flags
	spell.Flags |= added
	deal()
	spell.Flags &^= added
}

// The trigger casts the spell on the hit that caused the proc, after its options have set the callback
// the target depends on. The mask is weaponProcStrikeMask's.
func attachProcDamage(character *core.Character, trigger *core.ProcTrigger, cfg core.SpellConfig) *core.Spell {
	cfg.ProcMask = weaponProcStrikeMask(cfg.DefenseType, cfg.Flags, trigger.IsWeaponProc)
	spell := character.RegisterSpell(cfg)
	trigger.Handler = procDamageHandler(character, spell, trigger.Callback)
	trigger.TriggerImmediately = true
	return spell
}

// The mask a proc's damage spell casts with. A weapon proc's spell of melee defense is a weapon's
// strike, a main-hand special as a triggered melee spell is; one of ranged defense is a ranged
// special. The weapon procs, combat enchants and oils that hear those hits hear it, its own weapon
// included, and a listener on auto attacks alone does not. A melee strike an off-hand hit cast is
// heard by no weapon proc, combat enchant or oil on either hand: that cast carries Suppress Weapon
// Procs, through its flight where it has one, and its damage and bleed are those of any other cast.
// Only a spell flagged as a proc is a strike: any aura on melee specials would hear one without the
// flag, so it keeps no mask, as any other proc's damage does.
func weaponProcStrikeMask(defenseType core.DefenseType, flags core.SpellFlag, isWeaponProc bool) core.ProcMask {
	if !isWeaponProc || !flags.Matches(core.SpellFlagProc) {
		return core.ProcMaskEmpty
	}
	switch defenseType {
	case core.DefenseTypeMelee:
		return core.ProcMaskMeleeMHSpecial
	case core.DefenseTypeRanged:
		return core.ProcMaskRangedSpecial
	}
	return core.ProcMaskEmpty
}

// Which unit the proc answers. What the result names depends on the callback, so the callback
// decides whether it may be read at all.
func procDamageTarget(character *core.Character, callback core.AuraCallback, spell *core.Spell, result *core.SpellResult) *core.Unit {
	target := character.CurrentTarget

	switch {
	case callback.Matches(core.CallbackOnSpellHitTaken):
		// Here result.Target is the wearer - core dispatches hit-taken through
		// result.Target.OnSpellHitTaken - so the retaliation goes to the attacker instead of
		// into the wearer's own health. This is the shield spike shape.
		if spell != nil && spell.Unit != nil {
			target = spell.Unit
		}

	case callback.Matches(core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt):
		// Land the extra damage on whatever was hit, not on the primary target - unless that is
		// the wearer. A sapper charge is a hit the character deals to itself, and "chance on hit
		// to deal damage" means the enemy it is fighting, not its own health.
		if result != nil && result.Target != nil && result.Target != &character.Unit {
			target = result.Target
		}

	default:
		// The heal callbacks, cast complete and apply effects carry either no result or one
		// whose target is an ally, so nothing there can name what to damage and the current
		// target stands. Reading result.Target regardless is what would have a heal-triggered
		// damage proc hit the healed ally.
	}

	return target
}

func registerSpellDataHealProc(cfg SpellDataProc) {
	registerSpellDataSelfProc(cfg, func(character *core.Character, heal *spelldata.Spell) core.SpellConfig {
		return spellDataProcHealSpell(character, heal, true)
	})
}

// An item or enchant proc whose spell shields the wearer with an A_SCHOOL_ABSORB aura. BuffSpellID
// names that spell.
func NewSpellDataAbsorbProc(cfg SpellDataProc, variants []ItemVariant) {
	forEachSpellDataVariant(cfg, variants, registerSpellDataAbsorbProc)
}

func registerSpellDataAbsorbProc(cfg SpellDataProc) {
	sourceID := cfg.effectSource().id
	registerSpellDataSelfProc(cfg, func(character *core.Character, absorb *spelldata.Spell) core.SpellConfig {
		return spellDataAbsorbSpell(character, absorb, sourceID, true)
	})
}

// A proc that casts BuffSpellID's spell on the wearer, as spellConfig builds it.
func registerSpellDataSelfProc(cfg SpellDataProc, spellConfig func(*core.Character, *spelldata.Spell) core.SpellConfig) {
	registerSpellDataRowProc(cfg, nil, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, row *spelldata.Spell) {
		applySpellDataSelfProc(agent, cfg, source, trigger, spellConfig(agent.GetCharacter(), row))
	})
}

// A proc read from its trigger's row and the row it applies, BuffSpellID's or the trigger's own. It is
// left unregistered where the item or enchant is implemented by hand, where refuses (if set) refuses
// the row, and where the trigger's row names no callback: a listener with no callback never fires,
// and the row says so before any character exists.
func registerSpellDataRowProc(cfg SpellDataProc, refuses func(row *spelldata.Spell) bool,
	apply func(agent core.Agent, source effectSource, trigger *spelldata.Spell, row *spelldata.Spell)) {
	source := cfg.effectSource()

	// Soft fail to allow for overrides for bad effects
	if source.isAlreadyImplemented() {
		return
	}

	trigger := spelldata.MustFind(cfg.TriggerSpellID)
	row := trigger
	if cfg.BuffSpellID != 0 {
		row = spelldata.MustFind(cfg.BuffSpellID)
	}

	if refuses != nil && refuses(row) {
		return
	}

	if !cfg.IsWeaponProc && decodedCallback(trigger) == core.CallbackEmpty {
		return
	}

	source.registerEffect(func(agent core.Agent) {
		apply(agent, source, trigger, row)
	})
}

func applySpellDataSelfProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, spellConfig core.SpellConfig) {
	character := agent.GetCharacter()
	selfSpell := character.RegisterSpell(spellConfig)

	config := spellDataProcListener(character, cfg, source, trigger)
	config.Handler = func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		selfSpell.Cast(sim, &character.Unit)
	}
	config.TriggerImmediately = true

	source.registerTrigger(character, config)
}

// The heal the proc casts, as its row states it: a share of the target's maximum health or an amount
// the effect rolls, the spell power share the row states, and a crit unless the row rules one out;
// or a heal over time where the row's heal is a periodic aura. It goes through the healing path so it is measured as healing. Like the damage shape it is a
// proc's spell, out of the rotation and not a cast of its own.
func spellDataProcHealSpell(character *core.Character, heal *spelldata.Spell, asProc bool) core.SpellConfig {
	config := spelldata.SpellConfig(&character.Unit, heal, spelldata.Magic(core.ProcMaskSpellHealing), castBy(asProc))
	// A heal crits for the magic multiplier whatever the row files it under: 1248759 states no
	// defense type at all.
	config.DefenseType = core.DefenseTypeMagic

	effect := heal.ProcHealEffect()
	if effect.Aura == dbcenums.A_PERIODIC_HEAL {
		return spellDataProcHotSpell(character, heal, effect, config)
	}

	config.DirectVariance = effect.DirectVariance()
	amount := func(sim *core.Simulation, target *core.Unit) float64 {
		if effect.Type == dbcenums.E_HEAL_PCT {
			return target.MaxHealth() * effect.Percent()
		}
		return effect.DirectAmount(sim)
	}

	cannotCrit := heal.CannotCrit()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		outcome := spell.OutcomeHealingCrit
		if cannotCrit {
			outcome = spell.OutcomeHealing
		}
		spell.CalcAndDealHealing(sim, target, amount(sim, target), outcome)
	}

	return config
}

// A heal over time on the wearer: the amount the effect rolls every period for the row's duration,
// on the spell power share of the ticking effect. A tick crits only where the row states Periodic
// Can Crit and does not rule crits out. A second proc while it runs starts it over.
func spellDataProcHotSpell(character *core.Character, heal *spelldata.Spell, effect *spelldata.Effect, config core.SpellConfig) core.SpellConfig {
	canCrit := heal.PeriodicCanCrit() && !heal.CannotCrit()
	config.Hot = spelldata.DotConfig(heal, effect, spelldata.Label(heal.Name+" HoT"))
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		outcome := dot.OutcomeTick
		if canCrit {
			outcome = dot.Spell.OutcomeTickHealingCrit
		}
		dot.Spell.CalcAndDealPeriodicHealing(sim, target, effect.Roll(sim), outcome)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}

	return config
}

func decodedCallback(s *spelldata.Spell) core.AuraCallback {
	return core.DecodeProcTypeMask(s.ProcFlags, s.ProcHint).Callback
}

func (cfg SpellDataProc) effectSource() effectSource {
	if cfg.EnchantID != 0 {
		return effectSource{id: cfg.EnchantID, isEnchant: true}
	}

	return effectSource{id: cfg.ItemID}
}

func NewSimpleStatActive(itemID int32) {
	// Soft fail to allow for overrides for bad effects
	if core.HasItemEffect(itemID) {
		return
	}

	core.NewItemEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()

		for _, itemEffect := range onUseEffectsFor(itemID) {
			spellConfig := core.SpellConfig{
				ActionID: core.ActionID{ItemID: itemID},
				Cast:     OnUseCast(character, itemEffect),
			}

			buffStats, buffDuration := onUseStatBuff(character, itemEffect)
			core.RegisterTemporaryStatsOnUseCD(character, itemEffect.BuffName, buffStats, buffDuration, spellConfig)
		}
	})
}

// An on-use item whose spell deals damage to its target: the direct damage its row rolls, the damage
// over time it applies, or both.
func NewSpellDataDamageOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeDPS, spellDataOnUseDamageSpell)
}

// The proc's damage spell, cast as a spell of its own rather than a proc's.
func spellDataOnUseDamageSpell(character *core.Character, damage *spelldata.Spell) core.SpellConfig {
	return spellDataProcDamageSpell(character, damage, false)
}

// An on-use item whose spell keeps an A_PERIODIC_TRIGGER_SPELL aura on the wearer for the row's
// duration. Every period the aura casts the row it triggers, whose damage reaches every enemy in its
// area or cone: Cloak of Fire's 18364 casts 18374, 25 Fire to the enemies around the wearer, every 5 s
// for 15 s. The label carries the item, since two items may apply the same row.
func NewSpellDataPeriodicDamageOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeDPS, func(character *core.Character, row *spelldata.Spell) core.SpellConfig {
		periodic := row.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL)
		damage := character.RegisterSpell(spellDataProcDamageSpell(character, spelldata.MustFind(periodic.TriggerID), true))

		config := spelldata.SpellConfig(&character.Unit, row, castBy(false))
		config.ProcMask = core.ProcMaskEmpty
		config.Dot = spelldata.DotConfig(row, periodic, spelldata.Label(fmt.Sprintf("%s %d", row.Name, itemID)))
		config.Dot.SelfOnly = true
		config.Dot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
			damage.Cast(sim, character.CurrentTarget)
		}
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.SelfHot().Apply(sim)
		}
		return config
	})
}

// An on-use item whose spell heals the wearer, at once or over time.
func NewSpellDataHealOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeSurvival, func(character *core.Character, heal *spelldata.Spell) core.SpellConfig {
		return spellDataProcHealSpell(character, heal, false)
	})
}

// An on-use item whose spell shields the wearer with an A_SCHOOL_ABSORB aura. The stats the item's
// effect entry states beside the absorb are the shield's for as long as it holds: in game a used-up
// absorb removes its whole aura, so Blazing Emblem's 33 Fire resistance goes with its 465 Fire absorb.
func NewSpellDataAbsorbOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeSurvival, func(character *core.Character, absorb *spelldata.Spell) core.SpellConfig {
		config := spellDataAbsorbSpell(character, absorb, itemID, false)
		for _, itemEffect := range onUseEffectsFor(itemID) {
			if itemEffect.BuffId != absorb.ID {
				continue
			}
			if buffStats, _ := onUseStatBuff(character, itemEffect); buffStats != (stats.Stats{}) {
				config.RelatedSelfBuff.AttachStatsBuff(buffStats)
			}
		}
		return config
	})
}

// The shield the row applies to the caster: the amount its absorb effect rolls, taken off the damage
// of the schools the effect's Misc masks, for the row's duration. A second cast replaces the shield
// left, and a shield used up before it expires deals the damage AbsorbBreakDamage names. The label
// carries the item or enchant, since two of them may apply the same row.
func spellDataAbsorbSpell(character *core.Character, absorb *spelldata.Spell, sourceID int32, asProc bool) core.SpellConfig {
	effect := absorb.AbsorbEffect()
	schools := core.SpellSchool(effect.Misc)

	var amount float64
	shield := character.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: spelldata.AuraConfig(absorb, spelldata.Label(fmt.Sprintf("%s %d", absorb.Name, sourceID))),
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return amount
		},
		ShouldApplyToResult: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult, _ bool) bool {
			return spell.SpellSchool.Matches(schools)
		},
	})

	if id, ok := AbsorbBreakDamage(absorb); ok {
		damage := character.RegisterSpell(spellDataProcDamageSpell(character, spelldata.MustFind(id), true))
		shield.AttachOnDamageAbsorbed(func(sim *core.Simulation, aura *core.DamageAbsorptionAura, _ *core.SpellResult, _ float64) {
			if aura.ShieldStrength <= 0 {
				damage.Cast(sim, character.CurrentTarget)
			}
		})
	}

	config := spelldata.SpellConfig(&character.Unit, absorb, castBy(asProc))
	config.ProcMask = core.ProcMaskEmpty
	config.RelatedSelfBuff = shield.Aura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		amount = effect.Roll(sim)
		shield.Activate(sim)
	}

	return config
}

// Absorbs whose tooltip states damage dealt when the shield is gone before its duration ends, keyed by
// the absorb with the spell dealing it. The spell store carries no tooltip text and no row field states
// the condition, so each is listed: Adaptive Combat Assistant's 1291097 deals 1291099's Nature damage
// to the enemies in melee range "if this shield expires before" its duration.
var absorbBreakDamage = map[int32]int32{
	1291097: 1291099,
}

// The spell dealing the damage the absorb's shield deals when it breaks early, where one is listed.
func AbsorbBreakDamage(absorb *spelldata.Spell) (int32, bool) {
	id, ok := absorbBreakDamage[absorb.ID]
	return id, ok
}

// The spells dealing damage that an absorb's row triggers or its tooltip references. The shield deals
// only the one AbsorbBreakDamage names; any other is damage the absorb shape does not model.
func AbsorbLinkedDamage(absorb *spelldata.Spell) []int32 {
	linked := slices.Clone(absorb.RefIDs)
	for _, e := range absorb.Effects {
		if e.TriggerID != 0 {
			linked = append(linked, e.TriggerID)
		}
	}

	var damage []int32
	for _, id := range linked {
		s := spelldata.Find(id)
		if id != absorb.ID && s != spelldata.Nil && !slices.Contains(damage, id) &&
			(s.DamageEffect() != spelldata.NilEffect || s.PeriodicDamageEffect() != spelldata.NilEffect) {
			damage = append(damage, id)
		}
	}
	return damage
}

// An on-use item whose spell raises the wearer's melee, ranged or cast speed. Every speed the row
// states is on one aura, up for the row's duration.
func NewSpellDataSpeedOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeDPS, spellDataOnUseSpeedSpell)
}

func spellDataOnUseSpeedSpell(character *core.Character, row *spelldata.Spell) core.SpellConfig {
	aura := character.RegisterAura(spelldata.AuraConfig(row)).AttachHastePseudoStats(row.SpeedPseudoStats())

	config := spelldata.SpellConfig(&character.Unit, row)
	config.ProcMask = core.ProcMaskEmpty
	config.RelatedSelfBuff = aura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	return config
}

func castBy(asProc bool) spelldata.SpellOpt {
	if asProc {
		return procSpell
	}
	return itemUseSpell
}

// The game casts a proc's spell off the hit that caused it. It spends neither the player's global
// cooldown nor the resource bar the row prices the spell at, both of which belong to casting it from
// the bar, it has no cast time to spend either, and it is a proc unless the row says it is not.
func procSpell(config *core.SpellConfig, row *spelldata.Spell) {
	spelldata.Proc()(config, row)
	if row.IsAProc() {
		config.Flags |= core.SpellFlagProc
	}
}

func itemUseSpell(config *core.SpellConfig, _ *spelldata.Spell) {
	config.Flags &^= core.SpellFlagAPL | core.SpellFlagPassiveSpell
	config.ManaCost = core.ManaCostOptions{}
	config.RageCost = core.RageCostOptions{}
	config.EnergyCost = core.EnergyCostOptions{}
	config.FocusCost = core.FocusCostOptions{}
}

// The spell a proc of the same row would cast, used from the item instead: it is the item's action,
// counts its casts, and runs on the item's cooldowns rather than the row's, with the cast time and
// global cooldown the row states. The player casts it, so it is not a proc: its hits and heals reach
// the listeners and its cast completes like any other.
func registerSpellDataOnUse(itemID int32, cdType core.CooldownType, spellConfig func(*core.Character, *spelldata.Spell) core.SpellConfig) {
	registerSpellDataOnUseCooldown(itemID, func(character *core.Character, row *spelldata.Spell) (core.SpellConfig, core.MajorCooldown, bool) {
		return spellConfig(character, row), core.MajorCooldown{Type: cdType}, true
	})
}

// The same, where the row decides the cooldown the manager files the spell under and when it uses it,
// or that the character has nothing to use it for.
func registerSpellDataOnUseCooldown(itemID int32, onUse func(*core.Character, *spelldata.Spell) (core.SpellConfig, core.MajorCooldown, bool)) {
	// Soft fail to allow for overrides for bad effects
	if core.HasItemEffect(itemID) {
		return
	}

	core.NewItemEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()

		for _, itemEffect := range onUseEffectsFor(itemID) {
			row := spelldata.MustFind(itemEffect.BuffId)
			config, cooldown, ok := onUse(character, row)
			if !ok {
				continue
			}
			config.ActionID = core.ActionID{ItemID: itemID}

			itemCast := OnUseCast(character, itemEffect)
			config.Cast = spelldata.Cast(row)
			config.Cast.CD, config.Cast.SharedCD = itemCast.CD, itemCast.SharedCD

			cooldown.Spell = character.RegisterSpell(config)
			character.AddMajorCooldown(cooldown)
		}
	})
}

func onUseEffectsFor(itemID int32) []*proto.ItemEffect {
	onUseEffects := core.FilterSlice(itemEffectsFor(itemID), func(effect *proto.ItemEffect) bool {
		return effect.GetOnUse() != nil
	})
	if len(onUseEffects) == 0 {
		panic(fmt.Sprintf("No active effects found for item with ID: %d!", itemID))
	}
	return onUseEffects
}

// The item effect's own cooldown and the category cooldown it shares, whatever the spell's row states.
func OnUseCast(character *core.Character, itemEffect *proto.ItemEffect) core.CastConfig {
	cast := core.CastConfig{SharedCD: sharedCooldown(character, itemEffect)}
	if cooldownMs := itemEffect.GetOnUse().CooldownMs; cooldownMs > 0 {
		cast.CD = core.Cooldown{
			Timer:    character.NewTimer(),
			Duration: time.Duration(cooldownMs) * time.Millisecond,
		}
	}
	return cast
}

// The on-use buff's stats and duration, scaled by its area bonus.
func onUseStatBuff(character *core.Character, itemEffect *proto.ItemEffect) (stats.Stats, time.Duration) {
	amount, duration := spelldata.Find(itemEffect.BuffId).AreaBonus(&character.Env.Encounter)
	buffStats := stats.FromProtoMap(itemEffect.GetScalingOptions()[int32(0)].GetStats()).Multiply(amount)
	return buffStats, time.Duration(float64(itemEffect.EffectDurationMs)*duration) * time.Millisecond
}

type StackingStatBonusCD struct {
	Name               string
	ID                 int32
	CD                 time.Duration
	Callback           core.AuraCallback
	ProcMask           core.ProcMask
	SpellFlags         core.SpellFlag
	Outcome            core.HitOutcome
	RequireDamageDealt bool
	// See core.ProcTrigger.
	RequireDamageComponent bool
	// See core.ProcTrigger. The generator reads them off the stack proc's aura, which is always an
	// Equip aura, so there is no weapon-proc counterpart here.
	CanProcFromProcs  bool
	SpellFlagsExclude core.SpellFlag

	// The stacks will only be granted as long as the trinket is active
	TrinketLimitsDuration bool
}

// Where the stacks actually live. A database-resolved stacking trinket keeps the window and the
// stacks in two auras, so the count, the per-stack stats and the stat aura's identity all come from
// the nested one rather than from the effect itself, and the window is then always what bounds it -
// whatever the config asked for. A flat trinket states all of it on the effect.
type stackingStats struct {
	actionID      core.ActionID
	maxStacks     int32
	perStack      map[int32]float64
	windowBounded bool
}

func resolveStackingStats(effect *proto.ItemEffect, effectActionID core.ActionID, trinketLimitsDuration bool) stackingStats {
	if stackingAura := effect.StackingAura; stackingAura != nil {
		return stackingStats{
			actionID:      core.ActionID{SpellID: stackingAura.BuffId},
			maxStacks:     stackingAura.MaxCumulativeStacks,
			perStack:      stackingAura.GetScalingOptions()[int32(0)].GetStats(),
			windowBounded: true,
		}
	}

	return stackingStats{
		actionID:      effectActionID,
		maxStacks:     effect.MaxCumulativeStacks,
		perStack:      effect.GetScalingOptions()[int32(0)].GetStats(),
		windowBounded: trinketLimitsDuration,
	}
}

// The aura the on-use itself applies. Effects that name no buff spell fall back to the item.
func stackingAuraID(effect *proto.ItemEffect, itemID int32) core.ActionID {
	if auraID := (core.ActionID{SpellID: effect.BuffId}); !auraID.IsEmptyAction() {
		return auraID
	}

	return core.ActionID{ItemID: itemID}
}

// The aura pair a stacking on-use drives. Where the window bounds the stacks the stat aura is given
// no duration of its own and a second aura ends it on expiry; otherwise the stat aura is its own
// window and both returns are the same object, which is what the caller's identity check keys off.
func buildStackingCDAuras(character *core.Character, config StackingStatBonusCD, effect *proto.ItemEffect, stacks stackingStats) (*core.StatBuffAura, *core.Aura) {
	auraDuration := time.Millisecond * time.Duration(effect.EffectDurationMs)

	statAura := core.MakeStackingAura(character, core.StackingStatAura{
		Aura: core.Aura{
			Label:     config.Name + " Proc",
			ActionID:  stacks.actionID,
			Duration:  core.TernaryDuration(stacks.windowBounded, core.NeverExpires, auraDuration),
			MaxStacks: stacks.maxStacks,
		},
		BonusPerStack: stats.FromProtoMap(stacks.perStack),
	})

	if !stacks.windowBounded {
		return statAura, statAura.Aura
	}

	return statAura, character.RegisterAura(core.Aura{
		Label:    fmt.Sprintf("%s Limit Aura %s", config.Name, effect.BuffName),
		ActionID: stackingAuraID(effect, config.ID),
		Duration: auraDuration,
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			statAura.Deactivate(sim)
		},
	})
}

// What moves the stack count while the window is open. Attached to the window so it is live only
// then, and a decaying trinket spends a stack per event where the rest gain one.
func attachStackingCDTrigger(character *core.Character, config StackingStatBonusCD, effect *proto.ItemEffect, statAura *core.StatBuffAura, windowAura *core.Aura) {
	// Rate and lockout come off the stack proc, the same way the proc-item sibling reads them. Taken
	// from the config instead they were always zero - the generated on-use call states neither - and
	// a zero chance is normalised to 1, so a stack proc with a real chance or an ICD would have
	// stacked on every qualifying event with no cooldown. The getters are nil-safe: an item effect
	// that carries no stack proc keeps a plain always-on trigger.
	stackProc := effect.GetStackProc()

	var stackDPM *core.DynamicProcManager
	if stackProc != nil {
		stackDPM = dpmForMask(character, effectSource{id: config.ID}, stackProc.GetPpm(), config.ProcMask)
	}

	windowAura.AttachProcTriggerCallback(&character.Unit, core.ProcTrigger{
		Name:               config.Name,
		Callback:           config.Callback,
		ProcMask:           config.ProcMask,
		SpellFlags:         config.SpellFlags,
		SpellFlagsExclude:  config.SpellFlagsExclude,
		CanProcFromProcs:   config.CanProcFromProcs,
		Outcome:            config.Outcome,
		RequireDamageDealt: config.RequireDamageDealt,
		ProcChance:         core.TernaryFloat64(stackDPM == nil, stackProc.GetProcChance(), 0),
		ICD:                time.Millisecond * time.Duration(stackProc.GetIcdMs()),
		DPM:                stackDPM,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			if !statAura.IsActive() {
				return
			}

			if effect.StacksDecay {
				statAura.RemoveStack(sim)
			} else {
				statAura.AddStack(sim)
			}
		},

		RequireDamageComponent: config.RequireDamageComponent,
	})
}

// Creates a new stacking stats bonus aura based on the configuration. If Bonus is not given, the ItemEffect of the item will be used
// to determine the correct values.
func NewStackingStatBonusCD(config StackingStatBonusCD) {
	core.NewItemEffect(config.ID, func(agent core.Agent) {
		character := agent.GetCharacter()
		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(config.ID)

		for _, itemEffect := range itemEffectsFor(config.ID) {
			stacks := resolveStackingStats(itemEffect, stackingAuraID(itemEffect, config.ID), config.TrinketLimitsDuration)
			statAura, procAura := buildStackingCDAuras(character, config, itemEffect, stacks)

			attachStackingCDTrigger(character, config, itemEffect, statAura, procAura)

			spell := character.RegisterSpell(core.SpellConfig{
				ActionID: core.ActionID{ItemID: config.ID},
				Flags:    core.SpellFlagNoOnCastComplete,

				Cast: core.CastConfig{
					CD: core.Cooldown{
						Timer:    character.NewTimer(),
						Duration: config.CD,
					},
					SharedCD: sharedCooldown(character, itemEffect),
				},

				ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
					statAura.Activate(sim)
					if procAura != statAura.Aura {
						procAura.Activate(sim)
					}
					if itemEffect.StacksDecay {
						statAura.SetStacks(sim, stacks.maxStacks)
					}
				},

				RelatedSelfBuff: statAura.Aura,
			})

			character.AddMajorCooldown(core.MajorCooldown{
				Spell:    spell,
				Type:     core.CooldownTypeDPS,
				BuffAura: statAura,
			})

			// The stat aura and not the window: the stacks are what an APL keys off. This is the
			// registry behind the "Item Stat Proc Check" value and the "Activate All Stat Buff Proc
			// Auras" action, so without it an APL asking after Insight of the Qiraji stacks matches
			// nothing. Registering the major cooldown alone leaves plain use-on-cooldown output
			// unchanged, which is why no fixture records the difference.
			character.AddStatProcBuff(config.ID, statAura, false, eligibleSlots)
		}
	})
}

func NewStackingStatBonusEffectWithVariants(config ProcStatBonusEffect, variants []ItemVariant) {
	forEachVariant(config, variants, func(config ProcStatBonusEffect) {
		factory_StatBonusEffect(config, nil)
	})
}

// func NewStackingStatBonusEffect(config StackingStatBonusEffect) {
// 	// Ignore empty dummy implementations
// 	if config.Callback == core.CallbackEmpty {
// 		return
// 	}

// 	if core.HasItemEffect(config.ItemID) {
// 		return
// 	}

// 	core.NewItemEffect(config.ItemID, func(agent core.Agent) {
// 		character := agent.GetCharacter()
// 		eligibleSlots := character.ItemSwap.EligibleSlotsForItem(config.ItemID)
// 		item := core.GetItemByID(config.ItemID)

// 		for _, itemEffect := range item.ItemEffects {

// 			var procEffect *proto.ItemEffect
// 			if itemEffect != nil {
// 				if itemEffect.GetProc() != nil {
// 					procEffect = itemEffect
// 				}
// 			}

// 			if procEffect == nil {
// 				err, _ := fmt.Printf("Error getting proc effect for item/enchant %v", config.ItemID)
// 				panic(err)
// 			}

// 			proc := procEffect.GetProc()
// 			procAction := core.ActionID{SpellID: procEffect.BuffId}
// 			procAura := core.MakeStackingAura(character, core.StackingStatAura{
// 				Aura: core.Aura{
// 					Label:     config.Name + " Proc",
// 					ActionID:  procAction,
// 					Duration:  time.Millisecond * time.Duration(procEffect.EffectDurationMs),
// 					MaxStacks: config.MaxStacks,
// 				},
// 				BonusPerStack: stats.FromProtoMap(procEffect.ScalingOptions[int32(0)].Stats),
// 			})

// 			var dpm *core.DynamicProcManager
// 			if proc.GetPpm() > 0 {
// 				if config.ProcMask == core.ProcMaskUnknown {
// 					dpm = character.NewDynamicLegacyProcForEnchant(config.ItemID, proc.GetPpm(), 0)
// 				} else {
// 					dpm = character.NewLegacyPPMManager(proc.GetPpm(), config.ProcMask)
// 				}
// 			}

// 			triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
// 				ActionID:           core.ActionID{ItemID: config.ItemID},
// 				Name:               config.Name,
// 				Callback:           config.Callback,
// 				ProcMask:           config.ProcMask,
// 				SpellFlags:         config.SpellFlags,
// 				Outcome:            config.Outcome,
// 				RequireDamageDealt: config.RequireDamageDealt,
// 				ProcChance:         proc.GetProcChance(),
// 				DPM:                dpm,
// 				ICD:                time.Millisecond * time.Duration(proc.IcdMs),
// 				Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
// 					procAura.Activate(sim)
// 					procAura.AddStack(sim)
// 				},
// 			})

// 			procAura.Icd = triggerAura.Icd
// 			character.AddStatProcBuff(config.ItemID, procAura, false, eligibleSlots)
// 			character.ItemSwap.RegisterProcWithSlots(config.ItemID, triggerAura, eligibleSlots)

// 		}
// 	})
// }

type OutcomeType uint64

const (
	OutcomeDefault                  = 0
	OutcomeMeleeCanCrit OutcomeType = iota
	OutcomeMeleeNoCrit
	OutcomeMeleeNoBlockDodgeParry
	OutcomeMeleeNoBlockDodgeParryCrit
	OutcomeSpellCanCrit
	OutcomeSpellNoCrit
	OutcomeSpellNoMissCanCrit
	OutcomeRangedCanCrit
	OutcomeRangedNoCrit
	OutcomeAlwaysHit
)

type ProcDamageEffect struct {
	ItemID     int32
	SpellID    int32
	EnchantID  int32
	Trigger    core.ProcTrigger
	TriggerDPM func(*core.Character) *core.DynamicProcManager
	School     core.SpellSchool
	// From SpellCategories. Left unset, it is inferred from School and IsMelee.
	DefenseType      core.DefenseType
	MinDmg           float64
	MaxDmg           float64
	BonusCoefficient float64
	IsMelee          bool
	Flags            core.SpellFlag
	Outcome          OutcomeType
	// Set when the game files bar the damage spell from critting, which the sim has no way to know:
	// it is a spell attribute, so only the database generator can see it.
	CannotCrit bool
}

func GetOutcome(spell *core.Spell, outcome OutcomeType) core.OutcomeApplier {
	switch outcome {
	case OutcomeMeleeCanCrit:
		return spell.OutcomeMeleeSpecialHitAndCrit
	case OutcomeMeleeNoCrit:
		return spell.OutcomeMeleeSpecialHit
	case OutcomeMeleeNoBlockDodgeParry:
		return spell.OutcomeMeleeSpecialNoBlockDodgeParry
	case OutcomeMeleeNoBlockDodgeParryCrit:
		return spell.OutcomeMeleeSpecialNoBlockDodgeParryNoCrit
	case OutcomeSpellCanCrit:
		return spell.OutcomeMagicHitAndCrit
	case OutcomeSpellNoMissCanCrit:
		return spell.OutcomeMagicCrit
	case OutcomeSpellNoCrit:
		return spell.OutcomeMagicHit
	case OutcomeRangedCanCrit:
		return spell.OutcomeRangedHitAndCrit
	case OutcomeRangedNoCrit:
		return spell.OutcomeRangedHit
	case OutcomeAlwaysHit:
		return spell.OutcomeAlwaysHit
	default:
		return spell.OutcomeMagicHitAndCrit
	}
}

func NewProcDamageEffect(config ProcDamageEffect) {
	isEnchant := config.EnchantID != 0

	var effectFn func(id int32, effect core.ApplyEffect)
	var effectID int32
	var triggerActionID core.ActionID

	if isEnchant {
		effectID = config.EnchantID
		effectFn = core.NewEnchantEffect
		triggerActionID = core.ActionID{SpellID: config.SpellID}
	} else {
		effectID = config.ItemID
		effectFn = core.NewItemEffect
		triggerActionID = core.ActionID{ItemID: config.ItemID}
	}

	effectFn(effectID, func(agent core.Agent) {
		character := agent.GetCharacter()

		minDmg := config.MinDmg
		maxDmg := config.MaxDmg

		defenseType := damageDefenseType(config.DefenseType, config.School, config.IsMelee)

		// Per-character copy. config is captured once at registration and this body runs for
		// every character the effect applies to, so filling the trigger in place would hand
		// the second character the first one's DPM - a proc manager bound to another unit,
		// carrying its proc timing.
		triggerConfig := config.Trigger

		if core.ActionID.IsEmptyAction(triggerConfig.ActionID) {
			triggerConfig.ActionID = triggerActionID
		}

		if config.TriggerDPM != nil {
			triggerConfig.DPM = config.TriggerDPM(character)
		}

		attachProcDamage(character, &triggerConfig, core.SpellConfig{
			ActionID:    core.ActionID{SpellID: config.SpellID},
			SpellSchool: config.School,
			DefenseType: defenseType,
			Flags:       config.Flags,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			BonusCoefficient: config.BonusCoefficient,

			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.CalcAndDealDamage(sim, target, sim.Roll(minDmg, maxDmg), GetOutcome(spell, damageOutcome(defenseType, config.CannotCrit, config.Outcome)))
			},
		})

		triggerAura := character.MakeProcTriggerAura(triggerConfig)

		if isEnchant {
			character.ItemSwap.RegisterEnchantProc(effectID, triggerAura)
		} else {
			character.ItemSwap.RegisterProc(effectID, triggerAura)
		}
	})
}

///////////////////////////////////////////////////////////////////////////
//							Item and enchant plumbing
///////////////////////////////////////////////////////////////////////////

// Which of the two registries an effect belongs to, item or enchant, and its ID within it. Those
// are the only things the two differ in; every helper above treats them identically. Resolving it
// once keeps the same isEnchant branch from being written out at each of the six places that would
// otherwise need it.
type effectSource struct {
	id        int32
	isEnchant bool
}

func (config ProcStatBonusEffect) effectSource() effectSource {
	if config.EnchantID != 0 {
		return effectSource{id: config.EnchantID, isEnchant: true}
	}

	return effectSource{id: config.ItemID}
}

func (s effectSource) registerEffect(apply core.ApplyEffect) {
	if s.isEnchant {
		core.NewEnchantEffect(s.id, apply)
	} else {
		core.NewItemEffect(s.id, apply)
	}
}

// Whether a hand-written effect already covers this. That is the soft fail letting an override win
// over the generated registration, and it is why deleting one hands the generated version back.
func (s effectSource) isAlreadyImplemented() bool {
	if s.isEnchant {
		return core.HasEnchantEffect(s.id)
	}

	return core.HasItemEffect(s.id)
}

func (s effectSource) actionID() core.ActionID {
	if s.isEnchant {
		return core.ActionID{SpellID: s.id}
	}

	return core.ActionID{ItemID: s.id}
}

func (s effectSource) eligibleSlots(character *core.Character) []proto.ItemSlot {
	if s.isEnchant {
		return character.ItemSwap.EligibleSlotsForEffect(s.id)
	}

	return character.ItemSwap.EligibleSlotsForItem(s.id)
}

func (s effectSource) isEquipped(character *core.Character, slots []proto.ItemSlot) bool {
	if s.isEnchant {
		return character.HasEnchantEquipped(s.id, slots)
	}

	return character.HasItemEquipped(s.id, slots)
}

// The melee hands an enchant sits on, as proc masks, or 0 for an item and for an enchant in any other
// slot.
func (s effectSource) weaponHands(character *core.Character, slots []proto.ItemSlot) core.ProcMask {
	if !s.isEnchant {
		return 0
	}
	var hands core.ProcMask
	for _, slot := range slots {
		switch slot {
		case proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand:
			hands |= character.GetProcMaskForWeaponSlot(slot)
		default:
			return 0
		}
	}
	return hands
}

// The proc-carrying effect this item or enchant declares for the proc: the one keyed by the buff it
// applies, or by the trigger where that is the buff. Nil where it declares none.
func (s effectSource) procEntry(cfg SpellDataProc, trigger *spelldata.Spell) *proto.ItemEffect {
	effects := s.procEffects()
	if effect := effects[cfg.BuffSpellID]; effect != nil {
		return effect
	}
	return effects[trigger.ID]
}

// The proc-carrying effects this item or enchant declares, keyed by the aura each one applies.
func (s effectSource) procEffects() map[int32]*proto.ItemEffect {
	var declared []*proto.ItemEffect
	if s.isEnchant {
		declared = core.GetEnchantByEffectID(s.id).EnchantEffects
	} else if item := core.GetItemByID(s.id); item != nil {
		declared = item.ItemEffects
	}

	procEffects := make(map[int32]*proto.ItemEffect)
	for _, effect := range declared {
		if effect.GetProc() != nil {
			procEffects[effect.BuffId] = effect
		}
	}

	return procEffects
}

// A weapon enchant's buff drops when the weapon carrying it is swapped out, and a shield or
// held-in-off-hand enchant's when its item leaves the off hand. Any other enchant's runs out its
// duration: AddStatProcBuff only flips IsSwapped, which gates the next proc.
func (s effectSource) registerWeaponEnchantBuff(character *core.Character, procAura *core.StatBuffAura) {
	switch s.enchantPlacement() {
	case enchantOnWeapon:
		character.ItemSwap.RegisterWeaponEnchantBuff(procAura.Aura, s.id)
	case enchantInOffHand:
		character.ItemSwap.RegisterEnchantBuffWithSlots(procAura.Aura, s.id, []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand})
	}
}

type enchantPlacement byte

const (
	enchantElsewhere enchantPlacement = iota
	enchantOnWeapon
	enchantInOffHand
)

// A shield or held-in-off-hand enchant shares the weapon type but sits on no weapon.
func (s effectSource) enchantPlacement() enchantPlacement {
	if !s.isEnchant {
		return enchantElsewhere
	}

	ench := core.GetEnchantByEffectID(s.id)
	if ench == nil {
		return enchantElsewhere
	}

	switch ench.Type {
	case proto.ItemType_ItemTypeRanged:
		return enchantOnWeapon
	case proto.ItemType_ItemTypeWeapon:
		if ench.EnchantType == proto.EnchantType_EnchantTypeShield || ench.EnchantType == proto.EnchantType_EnchantTypeOffHand {
			return enchantInOffHand
		}
		return enchantOnWeapon
	}
	return enchantElsewhere
}

func (s effectSource) registerTrigger(character *core.Character, config core.ProcTrigger) *core.Aura {
	triggerAura := character.MakeProcTriggerAura(config)
	s.registerProc(character, triggerAura, s.eligibleSlots(character))
	return triggerAura
}

func (s effectSource) registerProc(character *core.Character, triggerAura *core.Aura, slots []proto.ItemSlot) {
	if s.isEnchant {
		character.ItemSwap.RegisterEnchantProcWithSlots(s.id, triggerAura, slots)
	} else {
		character.ItemSwap.RegisterProcWithSlots(s.id, triggerAura, slots)
	}
}

// The effects an on-use helper works from. A generated registration naming an item with no effect
// data is a database bug rather than a runtime case, so both callers fail loudly and identically.
func itemEffectsFor(itemID int32) []*proto.ItemEffect {
	item := core.GetItemByID(itemID)
	if item == nil {
		panic(fmt.Sprintf("No item with ID: %d", itemID))
	}

	if len(item.ItemEffects) == 0 {
		panic(fmt.Sprintf("No effects data for item with ID: %d", itemID))
	}

	return item.ItemEffects
}

// Share a cooldown only when the effect says it belongs to a category. One with no category shares
// nothing, and putting it on a generic trinket timer would gate it against unrelated items.
func sharedCooldown(character *core.Character, effect *proto.ItemEffect) core.Cooldown {
	onUse := effect.GetOnUse()
	if onUse == nil || onUse.CategoryId <= 0 {
		return core.Cooldown{}
	}

	duration := time.Millisecond * time.Duration(onUse.CategoryCooldownMs)
	if duration <= 0 {
		duration = time.Millisecond * time.Duration(effect.EffectDurationMs)
	}

	return core.Cooldown{
		Timer:    character.GetOrInitSpellCategoryTimer(onUse.CategoryId),
		Duration: duration,
	}
}

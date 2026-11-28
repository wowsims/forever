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
	// A "Chance on hit" item effect or a combat enchant (core.ProcTrigger); the generator reads it
	// from the trigger type.
	IsWeaponProc bool
	// The generator sets SpellFlagSuppressWeaponProcs here when the listener's aura is marked Aura Is
	// Weapon Proc.
	SpellFlagsExclude core.SpellFlag

	// What adds a stack while a stacking trinket's window is open: the container spell's own proc
	// flags, which are not the ones that open the window.
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

// The DefenseType a proc's damage rolls with. A stated one comes from the SpellCategories table, and
// IsMelee is for a caller that means melee damage without saying so through the school.
func damageDefenseType(defenseType core.DefenseType, school core.SpellSchool, isMelee bool) core.DefenseType {
	if defenseType != core.DefenseTypeNone {
		return defenseType
	}

	if isMelee || school.Matches(core.SpellSchoolPhysical) {
		return core.DefenseTypeMelee
	}

	return core.DefenseTypeMagic
}

// The outcome a proc's damage rolls when the caller states none: the hit table of its DefenseType.
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
				target := result.Target
				if spell == nil {
					target = procDamageTarget(character, config.Callback, spell, result)
				}
				if target != nil {
					procSpell.Cast(sim, target)
				}
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

			// windowAura is set only where the trigger opens a window that accumulates a separate stat aura.
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

				RequireDamageComponent: config.RequireDamageComponent,
			})

			attachStackTrigger(character, config, effect, procAura, windowAura)

			// This gates nothing: GetMatchingItemProcAuras drops an aura whose Icd is nil, and with it
			// every ICD-aware APL value.
			if proc.IcdMs != 0 {
				procAura.Icd = triggerAura.Icd
			}

			source.registerProc(character, triggerAura, eligibleSlots)
			source.registerWeaponEnchantBuff(character, procAura)
			character.AddStatProcBuff(source.id, procAura, source.isEnchant, eligibleSlots)
		}
	})
}

// What the proc does when it fires. A custom condition that refuses rolls the ICD back, since no proc
// happened.
func procHandler(config ProcStatBonusEffect, effect *proto.ItemEffect, procAura *core.StatBuffAura, windowAura *core.Aura, procSpell ExtraSpellInfo) func(*core.Simulation, *core.Spell, *core.SpellResult) {
	return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		// A custom condition gates the whole body, the window, the stack and the extra spell included.
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

// The three shapes a proc buff comes in; only the window shape returns a second aura.
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

// Event-driven stacks come from their own trigger, on the container's proc flags; a timer-driven
// stacking aura needs none.
func attachStackTrigger(character *core.Character, config ProcStatBonusEffect, effect *proto.ItemEffect, statAura *core.StatBuffAura, windowAura *core.Aura) {
	stackProc := effect.GetStackProc()
	if stackProc == nil || windowAura == nil || config.StackCallback == core.CallbackEmpty {
		return
	}

	// The stack trigger follows the opener's proc-ness rules; no stacking item in the database
	// differs between the two.
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

// An item or enchant proc as the game files' own rows state it: the listener, rate and ICD from the
// spell carrying the proc, the buff's duration and stacks from the spell it applies, and only the
// stats from the item's effect entry, where the sim's item level scaling lives.
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

	// A rate option in place of the one the item's entry and the trigger's row give, for a rate priced on
	// something neither states. It may reshape the listener.
	Rate spelldata.ProcOpt
}

// What an item adds to the spell its NewSpellDataDamageProc proc casts where the rows leave it to a
// script, by item id, given the spell's config before it registers: Pendulum of Doom's execute bonus.
// Hand-written code fills it from init, so the item's effect stays generated.
var DamageProcExtras = map[int32]func(character *core.Character, config *core.SpellConfig){}

// What an item changes on the listener its generated proc builds, by item id, before the handler is
// set: Felstriker's buff goes up in the instant of the hit that procs it. Hand-written code fills it
// from init, so the item's effect stays generated.
var ProcTriggerExtras = map[int32]func(trigger *core.ProcTrigger){}

// The same for an enchant's generated proc, by enchant id, given the wearer: Insight rolls once per cast.
var EnchantProcTriggerExtras = map[int32]func(character *core.Character, trigger *core.ProcTrigger){}

// How much longer than its rows state the buff of an item's generated proc lasts, by item id:
// Felstriker's buff counts its 3 s from the spell batch its proc would land in.
var ProcBuffDurationExtras = map[int32]time.Duration{}

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

// An aura the row applies to an enemy is NewSpellDataDebuffProc's to build, so it is left unregistered.
func registerSpellDataProc(cfg SpellDataProc) {
	registerSpellDataRowProc(cfg, (*spelldata.Spell).AppliesAnAuraToAnEnemy, func(agent core.Agent, source effectSource, trigger *spelldata.Spell, buff *spelldata.Spell) {
		applySpellDataProc(agent, cfg, source, trigger, buff)
	})
}

// The effect on one character: the buff it applies and the listener that applies it.
func applySpellDataProc(agent core.Agent, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell, buff *spelldata.Spell) {
	character := agent.GetCharacter()
	eligibleSlots := source.eligibleSlots(character)

	effect := source.procEntry(cfg, trigger)
	if effect == nil {
		panic(fmt.Sprintf("Error getting proc effects for item/enchant %v", source.id))
	}

	procAura := spellDataProcAura(character, cfg, trigger, buff, effect)

	listener := spellDataProcListener(character, cfg, source, trigger)
	listener.Handler = withProcEnergize(character, buff, spellDataProcHandler(buff, procAura))

	// The fallback the database layer applies. Where buff and trigger are one spell, the category
	// recovery is that spell's own cast throttle, so it is no gate on re-applying the buff.
	if listener.ICD == 0 && buff.ID != trigger.ID {
		listener.ICD = buff.CategoryCooldown()
	}
	triggerAura := source.registerTrigger(character, listener)

	attachChargeSpender(character, cfg, trigger, buff, procAura)

	// Keeps the ICD-aware APL values from dropping the effect (see factory_StatBonusEffect). Set after
	// the charge spender, which would leave the lockout between two charges here.
	procAura.Icd = triggerAura.Icd

	source.registerWeaponEnchantBuff(character, procAura)
	character.AddStatProcBuff(source.id, procAura, source.isEnchant, eligibleSlots)
}

// The listener the trigger's row describes, without the handler, plus the item's own name and action,
// which key the rolls and the metrics, and the rate the item's effect entry states.
func spellDataProcListener(character *core.Character, cfg SpellDataProc, source effectSource, trigger *spelldata.Spell) core.ProcTrigger {
	rate := cfg.Rate
	if rate == nil {
		rate = spellDataProcRate(source, trigger, source.procEntry(cfg, trigger).GetProc())
	}
	config := spelldata.ProcTrigger(character, trigger, nil, spelldata.ItemProcChance(trigger),
		weaponProcShape(cfg), rate, statedWeaponProcChance(cfg.IsWeaponProc, source))
	config.Name = cfg.Name
	config.ActionID = source.actionID()
	if extra, ok := ProcTriggerExtras[source.id]; ok && !source.isEnchant {
		extra(&config)
	}
	if extra, ok := EnchantProcTriggerExtras[source.id]; ok && source.isEnchant {
		extra(character, &config)
	}

	return config
}

// A weapon proc's listener, which no row states; anything else keeps the one its row decodes to.
func weaponProcShape(cfg SpellDataProc) spelldata.ProcOpt {
	if !cfg.IsWeaponProc {
		return func(*core.Character, *core.ProcTrigger) {}
	}
	return spelldata.WeaponProc()
}

// The rate for a proc the game files state none for. No row carries a SpellProcsPerMinuteID, so an
// item's rate comes from its effect entry, or from the store where an override put one. Only the
// effect knows which weapon or enchant slot the rate has to be measured against.
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
// hand's.
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

// What the proc does when it fires.
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

// The buff the proc applies. Its one count field holds either of two things: CumulativeAura stacks
// that each add their own stats, or the ProcCharges of one buff at full stats that the game spends by uses.
func spellDataProcAura(character *core.Character, cfg SpellDataProc, trigger *spelldata.Spell, buff *spelldata.Spell, effect *proto.ItemEffect) *core.StatBuffAura {
	// A window trinket resolves no stats on the aura its trigger applies, so it needs the window
	// machinery in factory_StatBonusEffect.
	if effect.GetStackingAura() != nil {
		panic(fmt.Sprintf("%s (%d): a proc whose stats live on an accumulating aura needs the stacking constructor", cfg.Name, cfg.ItemID))
	}

	aura := spelldata.AuraConfig(buff, spelldata.Label(cfg.Name+" Proc"))
	aura.Duration = procBuffDuration(cfg, trigger, buff)

	buffStats := stats.FromProtoMap(effect.GetScalingOptions()[int32(0)].GetStats())

	// The count sits on whichever row carries the aura, and the item effect entry takes the higher, as
	// this does: Idol of the Huntress keeps its 200 on the trigger.
	if stacks := max(buff.MaxStack, trigger.MaxStack); stacks > 0 {
		// A stacking buff's bonus against creature types, and its speed, would need adding per stack.
		if slices.ContainsFunc(buff.Effects, func(e spelldata.Effect) bool { return len(spelldata.VersusStats(e.Aura)) > 0 }) {
			panic(fmt.Sprintf("%s (%d): a stacking proc grants stats against creature types, which only a single buff carries", cfg.Name, cfg.ItemID))
		}
		if len(buff.SpeedEffects()) > 0 {
			panic(fmt.Sprintf("%s (%d): a stacking proc raises a speed, which only a single buff carries", cfg.Name, cfg.ItemID))
		}
		aura.MaxStacks = int32(stacks)
		return core.MakeStackingAura(character, core.StackingStatAura{Aura: aura, BonusPerStack: buffStats})
	}

	// Charges are the buff's own; the trigger's, where it has any, count the trigger's fires and are
	// not this aura's.
	aura.MaxStacks = int32(buff.ProcCharges)

	procAura := character.NewTemporaryStatsAuraWrapped(aura.Label, aura.ActionID, buffStats, aura.Duration, func(config *core.Aura) {
		config.MaxStacks = aura.MaxStacks
	})
	attachVersusStats(procAura.Aura, buff)
	// The item effect entry carries the buff's stats alone; its speed auras come from the row
	// (Ragehammer's Enrage 12686 states attack power and melee attack speed).
	procAura.AttachHastePseudoStats(buff.SpeedPseudoStats())
	return procAura
}

// The stats the buff grants against some creature types, which the item effect entry leaves out:
// Draconic Infused Emblem's 1318930 adds its spell damage a second time against Dragonkin.
func attachVersusStats(aura *core.Aura, buff *spelldata.Spell) {
	for _, effect := range buff.Effects {
		if bonus := spelldata.VersusBonus(effect.Aura, effect.Average()); bonus != (stats.Stats{}) {
			aura.AttachMobTypeBonusStats(bonus, spelldata.MobTypes(effect.Misc)...)
		}
	}
}

// How long the buff lasts: the buff's row, else the trigger's, the fallback the item effect entry
// applies as well. Core refuses to activate an aura of no duration mid-fight, so rows that state none
// are refused here, where the message can name what to look at.
func procBuffDuration(cfg SpellDataProc, trigger *spelldata.Spell, buff *spelldata.Spell) time.Duration {
	var extra time.Duration
	if source := cfg.effectSource(); !source.isEnchant {
		extra = ProcBuffDurationExtras[source.id]
	}

	if buff.DurationMs != 0 {
		return buff.Duration() + extra
	}

	if trigger.DurationMs != 0 {
		return trigger.Duration() + extra
	}

	panic(fmt.Sprintf("%s (%d): neither the proc's spell %d nor its buff %d states a duration for the aura it applies",
		cfg.Name, cfg.effectSource().id, trigger.ID, buff.ID))
}

// What spends a charge: the hits the buff's own row states. A buff whose row states no listener or no
// rate keeps its charges and runs out its duration. Where the buff is the trigger, the row's flags are
// the ones that granted it, so a spender built from them would take a charge on the hit that handed
// them over.
func attachChargeSpender(character *core.Character, cfg SpellDataProc, trigger *spelldata.Spell, buff *spelldata.Spell, procAura *core.StatBuffAura) {
	if buff.ID == trigger.ID || buff.ProcCharges <= 0 || buff.MaxStack > 0 || !statesATrigger(buff) {
		return
	}

	spelldata.AttachChargeSpender(character, procAura.Aura, buff, cfg.Name+" Charge")
}

// Whether a trigger can be built from this row, on the conditions spelldata.ProcTrigger panics on.
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
	// A listener with no ICD waits out its spell's own cooldown (EV-0145). A row that deals its damage
	// as the listener's own aura casts nothing.
	if config.ICD == 0 && damage.ProcDamageEffect().Aura != dbcenums.A_PROC_TRIGGER_DAMAGE {
		config.ICD = damage.Cooldown()
	}
	damageConfig := spellDataProcDamageSpell(character, damage, true)
	damageConfig.VersusBonus = ItemVersusBonus(character, source.id, proto.VersusEffectKind_VersusEffectKindDamage, trigger)
	if extra, ok := DamageProcExtras[source.id]; ok && !source.isEnchant {
		extra(character, &damageConfig)
	}
	attachProcDamage(character, &config, damageConfig)
	config.Handler = withProcEnergize(character, damage, config.Handler)
	source.registerTrigger(character, config)
}

// The handler followed by the row's own gain into the wearer's mana, rage or energy, where the row
// states one on its caster: it comes with every proc (EV-0108).
func withProcEnergize(character *core.Character, row *spelldata.Spell, handler core.ProcHandler) core.ProcHandler {
	gain := row.EnergizeEffect()
	if gain.Target[0] != dbcenums.TARGET_UNIT_CASTER {
		return handler
	}
	restore, ok := effectGain(character, gain, core.ActionID{SpellID: row.ID})
	if !ok {
		return handler
	}
	return func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		handler(sim, spell, result)
		restore(sim)
	}
}

// The spell the proc casts, as its row states it, and a proc's spell: out of the rotation and not a
// cast of its own, its hits reaching the damage listeners as a proc's do. A row that states Always Hit
// rolls no miss, dodge or parry, and crits the way damage to its own caster does (EV-0164).
func spellDataProcDamageSpell(character *core.Character, damage *spelldata.Spell, asProc bool) core.SpellConfig {
	config := spelldata.SpellConfig(&character.Unit, damage, damageShape(damage), castBy(asProc))

	defenseType := damageDefenseType(config.DefenseType, config.SpellSchool, false)
	config.DefenseType = defenseType
	outcome := damageOutcome(defenseType, damage.CannotCrit(), OutcomeDefault)
	application := damageOutcome(defenseType, true, OutcomeDefault)
	alwaysHits := damage.AlwaysHits()
	effect := damage.ProcDamageEffect()
	// A proc's hit on the magic defense type is a spell hit, so a listener on spell damage that hears
	// procs hears it. A row with no defense type of its own rolls no such listener (EV-0149, EV-0161).
	if asProc && damage.DefenseTypeCore() == core.DefenseTypeMagic {
		config.ProcMask = core.ProcMaskSpellDamage
	}
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
	onLanding := bothAfter(debuffOnLanding(character, damage), leechOnLanding(character, damage, effect))
	periodic := damage.PeriodicDamageEffect()
	if periodic == spelldata.NilEffect {
		multiTarget := effect.HitsAnArea() || effect.ChainTargets > 1
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			bind(spell)
			if multiTarget {
				dealOnArrival(sim, spell, target, calcMultiTargetDamage(sim, spell, target, damage, effect, hitOutcome), onLanding)
				return
			}
			single[0] = spell.CalcDamage(sim, target, effect.DirectAmount(sim), hitOutcome)
			dealOneOnArrival(sim, spell, target, single, onLanding)
		}
		return config
	}

	// The damage over time lands only with the hit that applies it: the direct hit where the row deals
	// one, else a no-crit roll on the spell's defense type. The hit and the damage over time each reach
	// an area where their own effect covers one, and an enemy only the damage over time reaches rolls
	// its own application.
	config.Dot = spelldata.DotConfig(damage, periodic)
	areaHit, areaDot := effect.HitsAnArea(), periodic.HitsAnArea()
	after := bothAfter(applyDotIfLanded(areaDot), onLanding)
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

// Runs first, then second; either may be nil.
func bothAfter(first, second afterDealt) afterDealt {
	if first == nil {
		return second
	}
	if second == nil {
		return first
	}
	return func(sim *core.Simulation, spell *core.Spell, target *core.Unit, results core.SpellResultSlice) {
		first(sim, spell, target, results)
		second(sim, spell, target, results)
	}
}

// What an E_HEALTH_LEECH gives its caster for the damage the hits dealt, after armor or resistance.
// Nil where the effect is no leech.
func leechOnLanding(character *core.Character, damage *spelldata.Spell, effect *spelldata.Effect) afterDealt {
	if effect.Type != dbcenums.E_HEALTH_LEECH {
		return nil
	}
	healthMetrics := character.NewHealthMetrics(core.ActionID{SpellID: damage.ID})
	return func(sim *core.Simulation, _ *core.Spell, _ *core.Unit, results core.SpellResultSlice) {
		dealt := 0.0
		for _, result := range results {
			dealt += result.Damage
		}
		GainLeechHealth(sim, character, dealt, effect, healthMetrics)
	}
}

// Health for the damage a leech dealt, an E_HEALTH_LEECH's or an A_PERIODIC_LEECH's: its Amplitude
// share, times the caster's healing taken multiplier.
func GainLeechHealth(sim *core.Simulation, character *core.Character, dealt float64, leech *spelldata.Effect, healthMetrics *core.ResourceMetrics) {
	if dealt <= 0 {
		return
	}
	character.GainHealth(sim, dealt*float64(leech.Amplitude)*character.PseudoStats.HealingTakenMultiplier, healthMetrics)
}

// Deals the results the cast calculated, then runs after, once the spell has landed. The flight
// carries a copy of the results, since the spell's next cast calculates into the same ones.
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

// The damage of an effect that reaches more than its target, without the enemies outside the
// creature types the row names, which still take their rolls and count toward a split and the AoE cap.
func calcMultiTargetDamage(sim *core.Simulation, spell *core.Spell, target *core.Unit, row *spelldata.Spell, effect *spelldata.Effect, outcome core.OutcomeApplier) core.SpellResultSlice {
	results := calcMultiTargetDamageOnAll(sim, spell, target, row, effect, outcome)
	if row.TargetCreatureType == 0 {
		return results
	}
	mobTypes := spelldata.MobTypes(row.TargetCreatureType)
	return slices.DeleteFunc(results, func(result *core.SpellResult) bool {
		if slices.Contains(mobTypes, result.Target.MobType) {
			return false
		}
		spell.DisposeResult(result)
		return true
	})
}

// The damage of an effect that reaches more than its target, dealt as a batch. A split row rolls once
// and divides the roll evenly among the targets it reaches. Otherwise each target rolls its own, and an
// uncapped area takes the encounter's AoE cap.
func calcMultiTargetDamageOnAll(sim *core.Simulation, spell *core.Spell, target *core.Unit, row *spelldata.Spell, effect *spelldata.Effect, outcome core.OutcomeApplier) core.SpellResultSlice {
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

// What a proc's damage lands on, for both damage constructors. A melee strike rolls weapon procs only
// where a main-hand white swing of the weapon cast it (R-0040, EV-0145). A strike asked for while the
// first is still casting goes one spell batch later.
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
		target := procDamageTarget(character, callback, spell, result)
		if target == nil {
			return
		}
		flags := core.SpellFlagNone
		if meleeStrike && (spell == nil || !spell.ProcMask.Matches(core.ProcMaskMeleeMHAuto) || character.PseudoStats.FeralCombatEnabled) {
			flags = core.SpellFlagSuppressWeaponProcs
		}
		cast(sim, target, flags)
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
// the target depends on.
func attachProcDamage(character *core.Character, trigger *core.ProcTrigger, cfg core.SpellConfig) *core.Spell {
	if strike := weaponProcStrikeMask(cfg.DefenseType, cfg.Flags, trigger.IsWeaponProc); strike != core.ProcMaskEmpty || cfg.ProcMask == core.ProcMaskUnknown {
		cfg.ProcMask = strike
	}
	spell := character.RegisterSpell(cfg)
	trigger.Handler = procDamageHandler(character, spell, trigger.Callback)
	trigger.TriggerImmediately = true
	return spell
}

// The mask a proc's damage spell casts with. A weapon proc's spell of melee defense is a weapon's
// strike, a main-hand special as a triggered melee spell is; one of ranged defense is a ranged
// special. The weapon procs, combat enchants and oils that hear those hits hear it, except where
// procDamageHandler's cast carries Suppress Weapon Procs. Only a spell flagged as a proc is a
// strike: any aura on melee specials would hear one without the flag, so it keeps no mask.
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

// Which unit the proc answers, nil where no enemy is left to answer. What the result names depends on
// the callback, so the callback decides whether it may be read at all.
func procDamageTarget(character *core.Character, callback core.AuraCallback, spell *core.Spell, result *core.SpellResult) *core.Unit {
	target := character.CurrentTarget

	switch {
	case spell == nil && callback.Matches(core.CallbackOnKill|core.CallbackOnTargetDies):
		// The dying enemy the result names takes nothing more.
		dying := result.Target
		if target != dying && target.IsEnabled() {
			return target
		}
		if next := character.Env.NextActiveTarget(dying); next != nil && &next.Unit != dying {
			return &next.Unit
		}
		return nil

	case callback.Matches(core.CallbackOnSpellHitTaken):
		// Here result.Target is the wearer (core dispatches hit-taken through
		// result.Target.OnSpellHitTaken), so the retaliation goes to the attacker: the shield spike shape.
		if spell != nil && spell.Unit != nil {
			target = spell.Unit
		}

	case callback.Matches(core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt):
		// The extra damage goes on whatever was hit, unless that is the wearer: a sapper charge is a
		// hit the character deals to itself.
		if result != nil && result.Target != nil && result.Target != &character.Unit {
			target = result.Target
		}

	default:
		// The heal callbacks, cast complete and apply effects carry either no result or one
		// whose target is an ally, so the current target stands.
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

// A proc read from its trigger's row and the row it applies, BuffSpellID's or the trigger's own. A
// listener with no callback never fires, and the row says so before any character exists.
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

// The heal the proc casts, as its row states it, or a heal over time where the row's heal is a
// periodic aura.
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

// A heal over time on the wearer. A second proc while it runs starts it over.
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
	source := effectSource{id: itemID}
	// Soft fail to allow for overrides for bad effects
	if source.isAlreadyImplemented() {
		return
	}

	source.registerEffect(func(agent core.Agent) {
		character := agent.GetCharacter()

		for _, itemEffect := range onUseEffectsFor(itemID) {
			spellConfig := core.SpellConfig{
				ActionID: core.ActionID{ItemID: itemID},
				Cast:     OnUseCast(character, itemEffect),
			}

			buffStats, buffDuration := onUseStatBuff(character, itemEffect)
			buff := core.RegisterTemporaryStatsOnUseCD(character, itemEffect.BuffName, buffStats, buffDuration, spellConfig)
			attachCarriedProc(character, buff.Aura, spelldata.Find(itemEffect.BuffId))
		}
	})
}

// The proc a stat buff's row carries beside its stats, live while the buff is up: Aegis of
// Preservation's 23780 heals the wearer on each melee or ranged hit taken.
func attachCarriedProc(character *core.Character, buff *core.Aura, row *spelldata.Spell) {
	if len(CarriedProcUnsupported(row)) > 0 {
		return
	}
	heal := character.RegisterSpell(spellDataProcHealSpell(character, row.FiredSpell(), true))
	trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		heal.Cast(sim, &character.Unit)
	}, spelldata.ItemProcChance(row))
	trigger.Name = buff.Label + " Proc"
	// Cast with the hit, the way every generated proc that casts a spell on the wearer is.
	trigger.TriggerImmediately = true
	buff.AttachProcTrigger(trigger)
}

// Why the proc a stat buff carries cannot be built, or nil where it can.
func CarriedProcUnsupported(row *spelldata.Spell) []string {
	fired := row.FiredSpell()
	heal := fired.ProcHealEffect()
	switch {
	case fired == spelldata.Nil:
		return []string{fmt.Sprintf("%d fires no spell the store carries", row.ID)}
	case heal == spelldata.NilEffect:
		return []string{fmt.Sprintf("%d fires %d, which heals no one", row.ID, fired.ID)}
	case heal.Target[0] != dbcenums.TARGET_UNIT_CASTER:
		return []string{fmt.Sprintf("%d's heal lands on implicit target %d, not the wearer", fired.ID, heal.Target[0])}
	case heal.Aura == dbcenums.A_PERIODIC_HEAL && (heal.PeriodMs <= 0 || fired.DurationMs <= 0):
		return []string{fmt.Sprintf("%d's heal over time states no period or no duration to tick over", fired.ID)}
	case row.ProcCharges > 0:
		return []string{fmt.Sprintf("%d states %d charges, which the stat aura does not spend", row.ID, row.ProcCharges)}
	}
	return spelldata.ItemProcUnsupportedWithRate(row, false, 0, 0)
}

// An on-use item whose spell deals damage to its target: the direct damage its row rolls, the damage
// over time it applies, or both.
func NewSpellDataDamageOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeDPS, func(character *core.Character, damage *spelldata.Spell) core.SpellConfig {
		config := spellDataOnUseDamageSpell(character, damage)
		config.VersusBonus = ItemVersusBonus(character, itemID, proto.VersusEffectKind_VersusEffectKindDamage, damage)
		return config
	})
}

// The proc's damage spell, cast as a spell of its own rather than a proc's.
func spellDataOnUseDamageSpell(character *core.Character, damage *spelldata.Spell) core.SpellConfig {
	return spellDataProcDamageSpell(character, damage, false)
}

// An on-use item whose spell keeps an A_PERIODIC_TRIGGER_SPELL aura on the wearer, which casts the row
// it triggers every period. The label carries the item, since two items may apply the same row.
func NewSpellDataPeriodicDamageOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeDPS, func(character *core.Character, row *spelldata.Spell) core.SpellConfig {
		periodic := row.FirstAura(dbcenums.A_PERIODIC_TRIGGER_SPELL)
		damageConfig := spellDataProcDamageSpell(character, spelldata.MustFind(periodic.TriggerID), true)
		damageConfig.VersusBonus = ItemVersusBonus(character, itemID, proto.VersusEffectKind_VersusEffectKindDamage, row)
		damage := character.RegisterSpell(damageConfig)

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

// An on-use item whose spell heals the wearer, at once or over time, and with the heal restores the
// resource its row's energize effect states: Major Recombobulator's 23064 restores mana too.
func NewSpellDataHealOnUse(itemID int32) {
	registerSpellDataOnUse(itemID, core.CooldownTypeSurvival, func(character *core.Character, heal *spelldata.Spell) core.SpellConfig {
		config := spellDataProcHealSpell(character, heal, false)
		restore, ok := effectGain(character, heal.EnergizeEffect(), core.ActionID{ItemID: itemID})
		if !ok {
			return config
		}
		applyEffects := config.ApplyEffects
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			applyEffects(sim, target, spell)
			restore(sim)
		}
		return config
	})
}

// An on-use item whose spell shields the wearer with an A_SCHOOL_ABSORB aura. The stats the item's
// effect entry states beside the absorb go with the shield (EV-0100).
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
		attachShieldEnergize(character, config.RelatedSelfBuff, absorb, itemID)
		return config
	})
}

// The resource the absorb's row restores every period while its shield holds: Mark of the Dragon
// Lord's 17252 restores mana every period until the shield is used up or runs out.
func attachShieldEnergize(character *core.Character, shield *core.Aura, absorb *spelldata.Spell, itemID int32) {
	gain := absorb.FirstAura(dbcenums.A_PERIODIC_ENERGIZE)
	restore, ok := effectGain(character, gain, core.ActionID{ItemID: itemID})
	if !ok {
		return
	}
	var ticks *core.PendingAction
	shield.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		ticks = core.StartPeriodicAction(sim, core.PeriodicActionOptions{Period: gain.Period(), OnAction: restore})
	})
	shield.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		ticks.Cancel(sim)
	})
}

// The shield the row applies to the caster. A second cast replaces the shield left. The label carries
// the item or enchant, since two of them may apply the same row.
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
// the absorb with the spell dealing it. No row field states the condition, so each is listed: Adaptive
// Combat Assistant's 1291097 deals 1291099's damage "if this shield expires before" its duration.
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

// The game casts a proc's spell off the hit that caused it: no global cooldown, cost or cast time, and
// a proc unless the row says it is not.
func procSpell(config *core.SpellConfig, row *spelldata.Spell) {
	spelldata.Proc()(config, row)
	spelldata.Triggered()(config, row)
}

func itemUseSpell(config *core.SpellConfig, _ *spelldata.Spell) {
	config.Flags &^= core.SpellFlagAPL | core.SpellFlagPassiveSpell
	config.ManaCost = core.ManaCostOptions{}
	config.RageCost = core.RageCostOptions{}
	config.EnergyCost = core.EnergyCostOptions{}
	config.FocusCost = core.FocusCostOptions{}
}

// The spell a proc of the same row would cast, used from the item instead, on the item's cooldowns
// with the cast time and global cooldown the row states. The player casts it, so it is not a proc.
func registerSpellDataOnUse(itemID int32, cdType core.CooldownType, spellConfig func(*core.Character, *spelldata.Spell) core.SpellConfig) {
	registerSpellDataOnUseCooldown(itemID, func(character *core.Character, row *spelldata.Spell) (core.SpellConfig, core.MajorCooldown, bool) {
		return spellConfig(character, row), core.MajorCooldown{Type: cdType}, true
	})
}

// The same, where the row decides the cooldown the manager files the spell under and when it uses it,
// or that the character has nothing to use it for.
func registerSpellDataOnUseCooldown(itemID int32, onUse func(*core.Character, *spelldata.Spell) (core.SpellConfig, core.MajorCooldown, bool)) {
	source := effectSource{id: itemID}
	// Soft fail to allow for overrides for bad effects
	if source.isAlreadyImplemented() {
		return
	}

	source.registerEffect(func(agent core.Agent) {
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

// Where the stacks live. A database-resolved stacking trinket keeps them on the nested aura, always
// bounded by the window whatever the config asks; a flat trinket states all of it on the effect.
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

// The aura pair a stacking on-use drives. Where the window does not bound the stacks, both returns
// are the same object, which is what the caller's identity check keys off.
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

// What moves the stack count while the window is open.
func attachStackingCDTrigger(character *core.Character, config StackingStatBonusCD, effect *proto.ItemEffect, statAura *core.StatBuffAura, windowAura *core.Aura) {
	// Rate and lockout come off the stack proc, since the generated on-use call states neither in the
	// config. The getters are nil-safe: an item effect with no stack proc keeps an always-on trigger.
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

// A stacking stats bonus aura, its values read from the item's effects.
func NewStackingStatBonusCD(config StackingStatBonusCD) {
	effectSource{id: config.ID}.registerEffect(func(agent core.Agent) {
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
			// Auras" action.
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

		// Per-character copy: config is shared by every character the effect applies to, so filling
		// the trigger in place would hand the next character this one's DPM.
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

// Which of the two registries an effect belongs to, item or enchant, and its ID within it.
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

// The generated files' effects by item, composed into the one registration core takes per item. The
// dispatcher reads the map when a character is built, after every registration, so a captured slice
// would miss the later ones.
var generatedItemEffects = map[int32][]core.ApplyEffect{}

var generating bool

// Marks the registrations that run until the returned function is called as the generated files'. A
// class file's call to the same constructor is a hand-written effect: it registers with core on its own.
func Generated() func() {
	generating = true
	return func() { generating = false }
}

func (s effectSource) registerEffect(apply core.ApplyEffect) {
	switch {
	case s.isEnchant:
		core.NewEnchantEffect(s.id, apply)
	case !generating:
		core.NewItemEffect(s.id, apply)
	default:
		first := len(generatedItemEffects[s.id]) == 0
		generatedItemEffects[s.id] = append(generatedItemEffects[s.id], apply)
		if first {
			core.NewItemEffect(s.id, func(agent core.Agent) {
				for _, apply := range generatedItemEffects[s.id] {
					apply(agent)
				}
			})
		}
	}
}

// Whether the item carries more than one generated effect. Core folds a second registration under a
// label into the first, so a composed item's equip aura takes a label of its own. Two listeners on one
// item would still fold, which the generator refuses.
func (s effectSource) composed() bool {
	return !s.isEnchant && len(generatedItemEffects[s.id]) > 1
}

// Whether a hand-written effect already covers this, the soft fail letting an override win over the
// generated registration. A hand-written effect registers with core on import, before the generated
// ones, so an id the generated effects hold is not one.
func (s effectSource) isAlreadyImplemented() bool {
	if s.isEnchant {
		return core.HasEnchantEffect(s.id)
	}

	return core.HasItemEffect(s.id) && len(generatedItemEffects[s.id]) == 0
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

// The effects an on-use helper works from. An item with no effect data is a database bug, so it panics.
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

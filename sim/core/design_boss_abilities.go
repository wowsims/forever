package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// Each ability's spells carry this plus the ability's index as their tag, which no spell id reaches, so an
// extra attack's tag names no spell.
const designBossAbilityTag int32 = 2_000_000

const (
	designSpellHitSpellID int32 = 18282 // Dummy Spell
	designStrikeSpellID   int32 = 15284 // Cleave
)

const (
	bossAbilityEndPriority   = ActionPriorityDOT
	bossAbilityStartPriority = ActionPriorityDOT + 1 // A window that starts as another ends keeps its effect up.
)

// A boss ability of the design. A spell hit or a strike happens at each occurrence that begins by the end, so
// the end cuts its repeats. For a thrash or swing timing window the end gives only the window's length, from
// the start to the end, and each repeat opens a window of that length wherever it falls.
type designBossAbility struct {
	index  int
	config *proto.DesignBossAbility
	target *Target // nil when the index is not one of the encounter's own targets.
	skip   bool    // Set when construction dropped the ability, with a construct note saying why.

	spell   *Spell // The spell hit's or the strike's.
	aura    *Aura  // The thrash or swing timing window's, permanent while a window runs.
	striker *designStriker

	fractions []float64
	starts    []time.Duration
	ends      []time.Duration
	applied   []bool // Whether each window took effect, which a disabled target's does not.
	nextStart int
	nextEnd   int
	running   int // The windows running at once.
	armed     bool

	startAction PendingAction
	endAction   PendingAction
}

// The strikes of one target, which replace its next main-hand swing in index order while any is armed.
type designStriker struct {
	target    *Target
	strikes   []*designBossAbility
	installed bool
	replace   ReplaceMHSwing
	previous  ReplaceMHSwing // The target's own replacement, back on the swing once no strike is armed.
}

// Registers the spells and auras of the design's boss abilities on their targets. The environment calls it
// once the raid is initialized, since a target's swing exists only then.
func (design *encounterDesign) registerBossAbilities(encounter *Encounter) {
	if design == nil || len(design.config.BossAbilities) == 0 {
		return
	}
	strikers := make(map[*Target]*designStriker)
	for index, config := range design.config.BossAbilities {
		ability := &designBossAbility{index: index, config: config}
		if targetIndex := int(config.TargetIndex); targetIndex >= 0 && targetIndex < design.numEncounterTargets {
			ability.target = encounter.AllTargets[targetIndex]
		}
		design.bossAbilities = append(design.bossAbilities, ability)
		if ability.target == nil {
			continue
		}
		if skip := ability.skipReason(); skip != "" {
			ability.skip = true
			design.constructNotes = append(design.constructNotes, fmt.Sprintf("boss ability %d %s; skipped.", index+1, skip))
			continue
		}

		occurrences := maxOccurrences(config.GetStart(), config.Repeat)
		ability.fractions = make([]float64, 0, occurrences)
		ability.starts = make([]time.Duration, 0, occurrences)
		ability.ends = make([]time.Duration, 0, occurrences)
		ability.applied = make([]bool, 0, occurrences)
		ability.startAction = PendingAction{Priority: bossAbilityStartPriority, OnAction: ability.start}
		ability.endAction = PendingAction{Priority: bossAbilityEndPriority, OnAction: ability.end}

		switch config.Kind {
		case proto.DesignBossAbility_SpellHit:
			ability.registerSpellHit()
		case proto.DesignBossAbility_Strike:
			ability.registerStrike()
			striker := strikers[ability.target]
			if striker == nil {
				striker = &designStriker{target: ability.target}
				striker.replace = striker.replaceSwing
				strikers[ability.target] = striker
			}
			striker.strikes = append(striker.strikes, ability)
			ability.striker = striker
		case proto.DesignBossAbility_Thrash:
			ability.registerThrash()
		case proto.DesignBossAbility_SwingTiming:
			ability.registerSwingTiming()
		}
	}
}

// Why the ability does nothing, or "" when it acts.
func (ability *designBossAbility) skipReason() string {
	config := ability.config
	if config.Kind != proto.DesignBossAbility_SpellHit && !ability.target.AutoAttacks.AutoSwingMelee {
		return "is a melee ability of a target that does not swing"
	}
	if ability.isWindow() && config.GetEnd().GetAt() <= config.GetStart().GetAt() {
		return "has a window that ends by its start"
	}
	switch config.Kind {
	case proto.DesignBossAbility_Strike:
		if config.WeaponPercent <= 0 {
			return "deals no weapon damage"
		}
	case proto.DesignBossAbility_Thrash:
		if config.ProcChance <= 0 || config.ExtraAttacks <= 0 {
			return "grants no extra attacks"
		}
	case proto.DesignBossAbility_SwingTiming:
		if config.SwingSpeedMultiplier < 0 {
			return "has a negative swing speed multiplier"
		}
	}
	return ""
}

func (ability *designBossAbility) tag() int32 {
	return designBossAbilityTag + int32(ability.index)
}

func (ability *designBossAbility) isWindow() bool {
	return ability.config.Kind == proto.DesignBossAbility_Thrash || ability.config.Kind == proto.DesignBossAbility_SwingTiming
}

// A hardcast magic spell, binary, whose damage is rolled between the ability's minimum and maximum.
func (ability *designBossAbility) registerSpellHit() {
	config := ability.config
	label := fmt.Sprintf("Encounter Design Spell Hit %d", ability.index+1)
	ability.spell = ability.target.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: designSpellHitSpellID, Tag: ability.tag()},
		SpellSchool: SpellSchoolFromProto(config.School),
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       SpellFlagBinary,

		Cast: CastConfig{
			DefaultCast: Cast{CastTime: DurationFromSeconds(max(config.CastTime, 0))},
			IgnoreHaste: true,
		},

		DamageMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := sim.RollWithLabel(config.MinDamage, config.MaxDamage, label)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
		},
	})
}

// A main-hand strike that deals the ability's percent of the weapon's damage.
func (ability *designBossAbility) registerStrike() {
	ability.spell = ability.target.RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: designStrikeSpellID, Tag: ability.tag()},
		SpellSchool:      SpellSchoolPhysical,
		DefenseType:      DefenseTypeMelee,
		ProcMask:         ProcMaskMeleeMHSpecial,
		Flags:            SpellFlagMeleeMetrics,
		DamageMultiplier: ability.config.WeaponPercent / 100,

		Cast: CastConfig{IgnoreHaste: true},

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := spell.Unit.AutoAttacks.MH().EnemyWeaponDamage(sim, spell.MeleeAttackPower(target), spell.Unit.PseudoStats.DamageSpread)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeEnemyMeleeWhite)
		},
	})
}

// While a window runs, a landed white swing grants the ability's extra attacks at its chance, once per ICD.
func (ability *designBossAbility) registerThrash() {
	config := ability.config
	unit := &ability.target.Unit
	trigger := ProcTrigger{
		Name:       fmt.Sprintf("Encounter Design Thrash %d", ability.index+1),
		Duration:   NeverExpires,
		Callback:   CallbackOnSpellHitDealt,
		ProcMask:   ProcMaskMeleeWhiteHit,
		Outcome:    OutcomeLanded,
		ProcChance: config.ProcChance / 100,
		ICD:        DurationFromSeconds(max(config.Icd, 0)),
	}
	onInit := unit.ExtraAttackProc(&trigger, ability.tag(), int(config.ExtraAttacks))
	ability.aura = unit.MakeProcTriggerAura(trigger).ApplyOnInit(onInit)
}

// While a window runs, the target's melee speed is multiplied by the ability's multiplier, or its swing is
// stopped for a multiplier of 0 and starts again at the window's end while the target is active.
func (ability *designBossAbility) registerSwingTiming() {
	multiplier := ability.config.SwingSpeedMultiplier
	target := ability.target
	aura := Aura{
		Label:    fmt.Sprintf("Encounter Design Swing Timing %d", ability.index+1),
		Duration: NeverExpires,
	}
	if multiplier == 0 {
		aura.OnGain = func(_ *Aura, sim *Simulation) {
			target.AutoAttacks.CancelAutoSwing(sim)
		}
		aura.OnExpire = func(_ *Aura, sim *Simulation) {
			if target.IsEnabled() && sim.CurrentTime < sim.Duration {
				target.AutoAttacks.EnableAutoSwing(sim)
			}
		}
	} else {
		aura.OnGain = func(_ *Aura, sim *Simulation) {
			target.MultiplyMeleeSpeed(sim, multiplier)
		}
		aura.OnExpire = func(_ *Aura, sim *Simulation) {
			target.MultiplyMeleeSpeed(sim, 1/multiplier)
		}
	}
	ability.aura = target.RegisterAura(aura)
}

// The lowest-index armed strike, which the swing becomes, disarmed; the target's own replacement comes back
// on the swing once no strike is armed.
func (striker *designStriker) replaceSwing(_ *Simulation, mhSwingSpell *Spell) *Spell {
	spell := mhSwingSpell
	armed := false
	for _, strike := range striker.strikes {
		if !strike.armed {
			continue
		}
		if spell == mhSwingSpell {
			strike.armed = false
			spell = strike.spell
			continue
		}
		armed = true
	}
	if !armed {
		striker.uninstall()
	}
	return spell
}

func (striker *designStriker) arm(strike *designBossAbility) {
	strike.armed = true
	if !striker.installed {
		striker.previous = striker.target.AutoAttacks.mh.replaceSwing
		striker.target.AutoAttacks.SetReplaceMHSwing(striker.replace)
		striker.installed = true
	}
}

func (striker *designStriker) uninstall() {
	if !striker.installed {
		return
	}
	striker.target.AutoAttacks.SetReplaceMHSwing(striker.previous)
	striker.previous = nil
	striker.installed = false
}

// Puts back what bringing the target in undoes: a running pause stops the swing Enable has started again.
// Thrash and speed windows hold through it, their auras being permanent, and an armed strike stays armed.
func (design *encounterDesign) reapplyBossAbilities(sim *Simulation, target *Target) {
	for _, ability := range design.bossAbilities {
		if ability.target == target && ability.running > 0 &&
			ability.config.Kind == proto.DesignBossAbility_SwingTiming && ability.config.SwingSpeedMultiplier == 0 {
			target.AutoAttacks.CancelAutoSwing(sim)
			return
		}
	}
}

// Schedules the iteration's boss abilities on their target: a spell hit with its hardcast, a strike that
// replaces the next swing, a thrash window of extra attacks, or a swing timing window that scales or pauses
// the swing, with their repeats rolled through rollOccurrences. The runner calls it only for a design with
// boss abilities.
func (design *encounterDesign) scheduleBossAbilities(sim *Simulation) {
	for _, ability := range design.bossAbilities {
		ability.nextStart = 0
		ability.nextEnd = 0
		ability.running = 0
		ability.armed = false
		ability.starts = ability.starts[:0]
		ability.ends = ability.ends[:0]
		ability.applied = ability.applied[:0]
		if ability.striker != nil {
			ability.striker.uninstall()
		}
		if ability.target == nil {
			if sim.Log != nil {
				sim.Log("Encounter design: boss ability %d names target %d, which the encounter does not have; skipped.", ability.index+1, ability.config.TargetIndex)
			}
			continue
		}
		if ability.skip {
			continue
		}

		config := ability.config
		ability.fractions = design.rollOccurrences(sim, config.GetStart(), config.Repeat, ability.fractions)
		end := design.roll(sim, config.GetEnd().GetAt(), config.GetEnd().GetJitter())
		if ability.isWindow() {
			length := end - ability.fractions[0]
			if length <= 0 {
				if sim.Log != nil {
					sim.Log("Encounter design: boss ability %d rolled a window that ends by its start; skipped this iteration.", ability.index+1)
				}
				continue
			}
			for _, fraction := range ability.fractions {
				if at := designTime(sim, fraction); at < sim.Duration {
					ability.starts = append(ability.starts, at)
					ability.ends = append(ability.ends, designTime(sim, fraction+length))
				}
			}
		} else {
			for _, fraction := range ability.fractions {
				if at := designTime(sim, fraction); fraction <= end && at < sim.Duration {
					ability.starts = append(ability.starts, at)
				}
			}
		}
		if len(ability.starts) > 0 {
			ability.startAction.NextActionAt = ability.starts[0]
			sim.AddPendingAction(&ability.startAction)
		}
	}
}

// Runs the ability's next occurrence, which does nothing while its target is not active.
func (ability *designBossAbility) start(sim *Simulation) {
	occurrence := ability.nextStart
	ability.nextStart++
	if ability.nextStart < len(ability.starts) {
		ability.startAction.NextActionAt = ability.starts[ability.nextStart]
		sim.AddPendingAction(&ability.startAction)
	}

	target := ability.target
	enabled := target.IsEnabled()
	if ability.isWindow() {
		ability.applied = append(ability.applied, enabled)
		if end := ability.ends[occurrence]; end < sim.Duration && ability.nextEnd == occurrence {
			ability.endAction.NextActionAt = end
			sim.AddPendingAction(&ability.endAction)
		}
		if !enabled {
			return
		}
		ability.running++
		if ability.running == 1 {
			ability.aura.Activate(sim)
		}
		return
	}

	if !enabled {
		return
	}
	switch ability.config.Kind {
	case proto.DesignBossAbility_SpellHit:
		victim := target.CurrentTarget
		if victim == nil {
			// An individual sim without a tank: the boss casts at the sim's player.
			victim = sim.Raid.AllPlayerUnits[0]
		}
		ability.spell.Cast(sim, victim)
	case proto.DesignBossAbility_Strike:
		if sim.Log != nil {
			target.Log(sim, "Encounter design: boss ability %d replaces the next swing.", ability.index+1)
		}
		ability.striker.arm(ability)
	}
}

// Ends the ability's next window, the effect going once no other window of it runs.
func (ability *designBossAbility) end(sim *Simulation) {
	occurrence := ability.nextEnd
	ability.nextEnd++
	if ability.nextEnd < ability.nextStart && ability.ends[ability.nextEnd] < sim.Duration {
		ability.endAction.NextActionAt = ability.ends[ability.nextEnd]
		sim.AddPendingAction(&ability.endAction)
	}

	if !ability.applied[occurrence] {
		return
	}
	ability.running--
	if ability.running == 0 {
		ability.aura.Deactivate(sim)
	}
}

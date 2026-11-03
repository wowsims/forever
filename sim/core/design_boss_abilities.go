package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

// Each ability's spells carry this plus the ability's index as their tag, which no spell id reaches, so an
// extra attack's tag names no spell. ENCOUNTER_DESIGN_ABILITY_TAG in ui/sim/proto/action_id/index.ts mirrors it.
const designBossAbilityTag int32 = 2_000_000

const (
	designSpellHitSpellID int32 = 18282 // Dummy Spell
	designStrikeSpellID   int32 = 15284 // Cleave
)

const designMaxTicks = 10_000 // Ticks of one pulse per iteration; past it the pulse ticks no more.

// A boss ability of the design. A spell hit, a strike or an area hit happens at each occurrence that begins by
// the end, so the end cuts its repeats. For a thrash, swing timing or pulse window the end gives only the
// window's length, from the start to the end, and each repeat opens a window of that length wherever it falls.
type designBossAbility struct {
	index  int
	config *proto.DesignBossAbility
	target *Target // nil when the index is not one of the encounter's own targets; target 0 for a pulse or an area hit.
	skip   bool    // Set when construction dropped the ability, with a construct note saying why.

	spell   *Spell // The spell hit's, the strike's, the pulse's or the area hit's.
	aura    *Aura  // The thrash or swing timing window's, permanent while a window runs.
	striker *designStriker

	occurrences designOccurrences // This iteration's starts.
	ends        []time.Duration
	applied     []bool // Whether each window took effect, which a disabled target's does not, nor a pulse's, which has no aura.
	nextEnd     int
	running     int // The windows running at once.
	open        int // The windows open on the timeline, the target active or not.
	armed       bool
	ticks       designOccurrences // A pulse's ticks this iteration, over all its windows in time order.

	endAction PendingAction
}

// The strikes of one target, which replace its next main-hand swing in index order while any is armed.
type designStriker struct {
	target    *Target
	strikes   []*designBossAbility
	installed bool
	replace   ReplaceMHSwing
	previous  ReplaceMHSwing // The target's own replacement, back on the swing once no strike is armed.
}

// Registers the spells and auras of the design's boss abilities on their targets.
func (design *encounterDesign) registerBossAbilities(encounter *Encounter) {
	strikers := make(map[*Target]*designStriker)
	for index, config := range design.config.BossAbilities {
		ability := &designBossAbility{index: index, config: config}
		if ability.isArea() {
			ability.target = encounter.AllTargets[0]
		} else if targetIndex := int(config.TargetIndex); targetIndex >= 0 && targetIndex < design.numEncounterTargets {
			ability.target = encounter.AllTargets[targetIndex]
		}
		design.bossAbilities = append(design.bossAbilities, ability)
		if skip := ability.skipReason(); skip != "" {
			ability.skip = true
			design.constructNotes = append(design.constructNotes, fmt.Sprintf("boss ability %d %s; skipped.", index+1, skip))
			continue
		}

		occurrences := maxOccurrences(config.GetStart(), config.Repeat)
		ability.occurrences = newDesignOccurrences(occurrences, designStartPriority, ability.start)
		ability.ends = make([]time.Duration, 0, occurrences)
		ability.applied = make([]bool, 0, occurrences)
		ability.endAction = PendingAction{Priority: designEndPriority, OnAction: ability.end}

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
		case proto.DesignBossAbility_Pulse:
			ability.registerArea(design)
			ticks := maxPulseTicks(config, occurrences, encounter.Duration+encounter.DurationVariation)
			if ticks > designMaxTicks {
				ticks = designMaxTicks
				design.constructNotes = append(design.constructNotes, fmt.Sprintf("boss ability %d ticks at most %d times a fight.", index+1, designMaxTicks))
			}
			ability.ticks = designOccurrences{
				times:  make([]time.Duration, 0, ticks),
				action: PendingAction{Priority: designEndPriority, OnAction: ability.tick},
			}
		case proto.DesignBossAbility_AreaHit:
			ability.registerArea(design)
			// After a move that starts at the same time, as a pulse's tick is, so an avoidable hit misses its walk.
			ability.occurrences.action.Priority = designEndPriority
		}
	}
}

// The most ticks a pulse has in one iteration of the longest fight: each of its windows at its longest, with
// the first tick at the window's start.
func maxPulseTicks(config *proto.DesignBossAbility, occurrences int, maxDuration time.Duration) int {
	length := (config.GetEnd().GetAt() + max(config.GetEnd().GetJitter(), 0)) - (config.GetStart().GetAt() - max(config.GetStart().GetJitter(), 0))
	perWindow := Clamp(length, 0, 1) * maxDuration.Seconds() / config.TickInterval
	return occurrences * (int(min(perWindow, designMaxTicks)) + 1)
}

// Why the ability does nothing, or "" when it acts.
func (ability *designBossAbility) skipReason() string {
	config := ability.config
	if ability.target == nil {
		return fmt.Sprintf("names target %d, which the encounter does not have", config.TargetIndex)
	}
	if config.Kind != proto.DesignBossAbility_SpellHit && !ability.isArea() && !ability.target.AutoAttacks.AutoSwingMelee {
		return "is a melee ability of a target that does not swing"
	}
	if ability.isWindow() && config.GetEnd().GetAt() <= config.GetStart().GetAt() {
		return "has a window that ends by its start"
	}
	if ability.isArea() && config.MaxDamage <= 0 {
		return "deals no damage"
	}
	switch config.Kind {
	case proto.DesignBossAbility_Pulse:
		if config.TickInterval <= 0 {
			return "ticks at no interval"
		}
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
	switch ability.config.Kind {
	case proto.DesignBossAbility_Thrash, proto.DesignBossAbility_SwingTiming, proto.DesignBossAbility_Pulse:
		return true
	}
	return false
}

func (ability *designBossAbility) isArea() bool {
	return ability.config.Kind == proto.DesignBossAbility_Pulse || ability.config.Kind == proto.DesignBossAbility_AreaHit
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

// A pulse's tick or an area hit: damage rolled between the ability's minimum and maximum for every player and
// every active pet, skipping a player a design move walks when the ability is avoidable. It belongs to target 0,
// for its level and its metrics, and is applied with no cast, so it lands while target 0 is down or dead; the
// boss's own damage modifiers leave it alone. A resistible one rolls the spell hit and the partial resist of a
// spell of its school, armor for a physical one; any other lands in full. Damage taken modifiers and absorbs
// apply to both, and the damage taken callbacks hear it as a boss spell's.
func (ability *designBossAbility) registerArea(design *encounterDesign) {
	config := ability.config
	label := fmt.Sprintf("Encounter Design Area Damage %d", ability.index+1)
	flags := SpellFlagIgnoreAttackerModifiers | SpellFlagNoOnCastComplete
	if !config.Resistible {
		flags |= SpellFlagIgnoreResists
	}

	ability.spell = ability.target.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: designSpellHitSpellID, Tag: ability.tag()},
		SpellSchool: SpellSchoolFromProto(config.School),
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskEmpty,
		Flags:       flags,

		DamageMultiplier: 1,

		ApplyEffects: func(sim *Simulation, _ *Unit, spell *Spell) {
			outcome := spell.OutcomeAlwaysHit
			if config.Resistible {
				outcome = spell.OutcomeMagicHit
			}
			hit := func(unit *Unit) {
				baseDamage := sim.RollWithLabel(config.MinDamage, config.MaxDamage, label)
				spell.CalcAndDealDamage(sim, unit, baseDamage, outcome)
			}
			for i, player := range sim.Raid.AllPlayerUnits {
				if !player.IsEnabled() {
					continue
				}
				if config.Avoidable && design.walking(sim, i) {
					if sim.Log != nil {
						player.Log(sim, "Encounter design: avoids boss ability %d, moving.", ability.index+1)
					}
					continue
				}
				hit(player)
			}
			for _, unit := range sim.Raid.AllUnits {
				if unit.Type == PetUnit && unit.IsEnabled() {
					hit(unit)
				}
			}
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
			if target.IsEnabled() && sim.CurrentTime < sim.Duration && !sim.Encounter.design.swingPaused(target) {
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
	if design.swingPaused(target) {
		target.AutoAttacks.CancelAutoSwing(sim)
	}
}

// Whether a swing timing window with a multiplier of 0 runs on the target, of any ability.
func (design *encounterDesign) swingPaused(target *Target) bool {
	for _, ability := range design.bossAbilities {
		if ability.target == target && ability.running > 0 &&
			ability.config.Kind == proto.DesignBossAbility_SwingTiming && ability.config.SwingSpeedMultiplier == 0 {
			return true
		}
	}
	return false
}

// Schedules the iteration's boss abilities on their target: a spell hit with its hardcast, a strike that
// replaces the next swing, a thrash window of extra attacks, a swing timing window that scales or pauses
// the swing, a pulse window of area ticks or an area hit, with their repeats rolled through rollOccurrences.
func (design *encounterDesign) scheduleBossAbilities(sim *Simulation) {
	for _, ability := range design.bossAbilities {
		ability.nextEnd = 0
		ability.running = 0
		ability.open = 0
		ability.armed = false
		ability.ends = ability.ends[:0]
		ability.applied = ability.applied[:0]
		ability.ticks.next = 0
		ability.ticks.times = ability.ticks.times[:0]
		if ability.striker != nil {
			ability.striker.uninstall()
		}
		if ability.skip {
			continue
		}

		config := ability.config
		occurrences := &ability.occurrences
		occurrences.roll(sim, design, config.GetStart(), config.Repeat)
		end := design.roll(sim, config.GetEnd().GetAt(), config.GetEnd().GetJitter())
		if ability.isWindow() {
			length := end - occurrences.fractions[0]
			if length <= 0 {
				if sim.Log != nil {
					sim.Log("Encounter design: boss ability %d rolled a window that ends by its start; skipped this iteration.", ability.index+1)
				}
				continue
			}
			for _, fraction := range occurrences.fractions {
				if at := designTime(sim, fraction); at < sim.Duration {
					occurrences.times = append(occurrences.times, at)
					ability.ends = append(ability.ends, designTime(sim, fraction+length))
				}
			}
		} else {
			for _, fraction := range occurrences.fractions {
				if at := designTime(sim, fraction); fraction <= end && at < sim.Duration {
					occurrences.times = append(occurrences.times, at)
				}
			}
		}
		occurrences.queueFirst(sim)
		if config.Kind == proto.DesignBossAbility_Pulse {
			ability.rollTicks(sim)
		}
	}
}

// Rolls the pulse's ticks window by window in occurrence order and queues the first: one at the window's
// start, then tick k at k intervals after it, moved by its own jitter from the design rand. A tick moved out of
// its window is dropped.
func (ability *designBossAbility) rollTicks(sim *Simulation) {
	ticks := &ability.ticks
	interval := DurationFromSeconds(ability.config.TickInterval)
	jitter := ability.config.TickJitter
	for i, start := range ability.occurrences.times {
		end := ability.ends[i]
		for k := time.Duration(0); start+k*interval < end && len(ticks.times) < cap(ticks.times); k++ {
			at := start + k*interval
			if k > 0 && jitter > 0 {
				at += DurationFromSeconds(jitter * (2*sim.designRand.NextFloat64() - 1))
			}
			if at >= start && at < end {
				ticks.times = append(ticks.times, at)
			}
		}
	}
	slices.Sort(ticks.times)
	ticks.queueFirst(sim)
}

func (ability *designBossAbility) tick(sim *Simulation) {
	ability.ticks.advance(sim)
	ability.spell.SkipCastAndApplyEffects(sim, sim.Raid.AllPlayerUnits[0])
}

// Runs the ability's next occurrence, which does nothing while its target is not active, except for a pulse's
// window and an area hit, which come from no target.
func (ability *designBossAbility) start(sim *Simulation) {
	occurrence := ability.occurrences.next
	ability.occurrences.advance(sim)

	target := ability.target
	enabled := target.IsEnabled()
	if ability.isWindow() {
		ability.open++
		applied := enabled && ability.aura != nil
		ability.applied = append(ability.applied, applied)
		if end := ability.ends[occurrence]; end < sim.Duration && ability.nextEnd == occurrence {
			ability.endAction.NextActionAt = end
			sim.AddPendingAction(&ability.endAction)
		}
		if !applied {
			return
		}
		ability.running++
		if ability.running == 1 {
			ability.aura.Activate(sim)
		}
		return
	}

	if ability.config.Kind == proto.DesignBossAbility_AreaHit {
		ability.spell.SkipCastAndApplyEffects(sim, sim.Raid.AllPlayerUnits[0])
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
	if ability.nextEnd < ability.occurrences.next && ability.ends[ability.nextEnd] < sim.Duration {
		ability.endAction.NextActionAt = ability.ends[ability.nextEnd]
		sim.AddPendingAction(&ability.endAction)
	}

	ability.open--
	if !ability.applied[occurrence] {
		return
	}
	ability.running--
	if ability.running == 0 {
		ability.aura.Deactivate(sim)
	}
}

// The current time while a window of the ability is open; else the start of its next window or occurrence, or
// NeverExpires.
func (ability *designBossAbility) nextAt(sim *Simulation) time.Duration {
	if ability.open > 0 {
		return sim.CurrentTime
	}
	return ability.occurrences.nextAt()
}

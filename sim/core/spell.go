package core

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type OnApplyEffects func(aura *Aura, sim *Simulation, target *Unit, spell *Spell)
type ApplySpellResults func(sim *Simulation, target *Unit, spell *Spell)
type ExpectedDamageCalculator func(sim *Simulation, target *Unit, spell *Spell, useSnapshot bool) *SpellResult
type CanCastCondition func(sim *Simulation, target *Unit) bool

type SpellConfig struct {
	// See definition of Spell (below) for comments on these.
	ActionID
	SpellSchool    SpellSchool
	DefenseType    DefenseType
	ProcMask       ProcMask
	Flags          SpellFlag
	BaseCost       float64
	MetricSplits   int
	ClassSpellMask int64
	ClassFlags     ClassFlags
	Rank           int32

	MissileSpeed       float64
	MissileMinDuration time.Duration

	ManaCost   ManaCostOptions
	EnergyCost EnergyCostOptions
	RageCost   RageCostOptions
	FocusCost  FocusCostOptions

	Cast               CastConfig
	ExtraCastCondition CanCastCondition
	CastRequirement    CastRequirement

	// Optional range constraints. If supplied, these are used to modify the ExtraCastCondition above to additionally check for DistanceFromTarget.
	MinRange     float64
	MaxRange     float64
	Charges      int // The maximum amount of charges this spell can have
	RechargeTime time.Duration

	BonusHitPercent       float64
	BonusCritPercent      float64
	BonusSpellDamage      float64
	BonusExpertisePercent float64

	DamageMultiplier         float64
	DamageMultiplierAdditive float64
	CritMultiplierAdditive   float64 // Additive extra crit damage %

	BonusBaseDamage       float64 // Certain items can increase the base damage of a spell e.g. https://www.wowhead.com/forever/item=28248/totem-of-the-void
	BonusCoefficient      float64 // EffectBonusCoefficient in the game files' SpellEffect table, "SP mod" on Wowhead (not necessarily shown there even if > 0)
	DirectVariance        float64 // EffectVariance of a spell-school hit or a heal: the direct amount, power included, lands within ±DirectVariance/2 of itself.
	TakenBonusCoefficient float64 // The coefficient the target's flat school bonus is taken at by a spell that sets no BonusCoefficient because its spell power goes through a weapon percent.

	ThreatMultiplier float64

	FlatThreatBonus     float64
	ThreatAPCoefficient float64

	// Performs the actions of this spell.
	ApplyEffects ApplySpellResults

	// Optional field. Calculates expected average damage.
	ExpectedInitialDamage ExpectedDamageCalculator
	ExpectedTickDamage    ExpectedDamageCalculator

	Dot    DotConfig
	Hot    DotConfig
	Shield ShieldConfig

	RelatedAuraArrays LabeledAuraArrays
	RelatedDotSpell   *Spell
	RelatedSelfBuff   *Aura

	Source ActionSource // Left unset, registration fills in the phase it runs in.
}

type Spell struct {
	// ID for this spell.
	ActionID
	//Rank of the spell.
	Rank int32

	// The unit who will perform this spell.
	Unit *Unit

	// Fire, Frost, Shadow, etc.
	SpellSchool SpellSchool
	SchoolIndex stats.SchoolIndex

	// Decides the base crit multiplier (1.5 for Magic, 2.0 for Melee and Ranged) and which
	// SpellMods with a DefenseType filter apply. Taken from the game files' SpellCategories table.
	// A spell left at DefenseTypeNone panics if it ever tries to crit.
	DefenseType DefenseType

	// Controls which effects can proc from this spell.
	ProcMask ProcMask

	// Flags
	Flags SpellFlag

	// The specific class spell id
	// should be a unique bit
	ClassSpellMask int64

	// The game files' SpellClassOptions: the family and mask bit its talents name this spell by.
	ClassFlags ClassFlags

	// Speed in yards/second. Spell missile speeds can be found in the game data.
	// Example: https://wow.tools/dbc/?dbc=spellmisc&build=3.4.0.44996
	MissileSpeed float64
	// The shortest the missile flies, however close the target: a spell with no speed flies this long.
	MissileMinDuration time.Duration

	ResourceMetrics *ResourceMetrics
	healthMetrics   []*ResourceMetrics

	Cost               *SpellCost // Cost for the spell.
	DefaultCast        Cast       // Default cast parameters with all static effects applied.
	CD                 Cooldown
	SharedCD           Cooldown
	IgnoreHaste        bool
	ExtraCastCondition CanCastCondition

	CastRequirement    CastRequirement
	hasCastRequirement bool
	casterAuras        []*Aura
	excludeCasterAuras []*Aura
	equipment          *Equipment

	// Optional range constraints. If supplied, these are used to modify the ExtraCastCondition above to additionally check for DistanceFromTarget.
	MinRange     float64
	MaxRange     float64
	MaxCharges   int // Maximum amount of charges the spell can have
	charges      int // Current amount of charges the spell has
	RechargeTime time.Duration

	rechargeTimer *PendingAction // used for the recharge timer

	castTimeFn func(spell *Spell) time.Duration // allows to override CastTime()

	// Performs a cast of this spell.
	castFn CastSuccessFunc

	hitRanges      []*hitRanges
	splitHitRanges [][]*hitRanges

	SpellMetrics      []SpellMetrics
	splitSpellMetrics [][]SpellMetrics // Used to split metrics by some condition.
	casts             int              // Sum of casts on all targets, for efficient CPM calculation
	lastAppliedAt     time.Duration    // When the spell last applied its effects; -NeverExpires if not yet this iteration.

	// Performs the actions of this spell.
	ApplyEffects ApplySpellResults

	// Optional field. Calculates expected average damage.
	expectedInitialDamageInternal ExpectedDamageCalculator
	expectedTickDamageInternal    ExpectedDamageCalculator
	expecting                     bool // While an expected-damage call runs: DirectVariance does not roll.

	// The current or most recent cast data.
	CurCast Cast

	BonusHitPercent          float64
	BonusCritPercent         float64
	BonusSpellDamage         float64
	BonusExpertisePercent    float64
	CastTimeMultiplier       float64
	CdMultiplier             float64       // The product of the nonzero cooldown multiplier mods; CooldownMultiplier reads it.
	cdMultiplierZeroes       int32         // Active cooldown multiplier mods of 0 or below, which remove the cooldown and cannot be divided back out.
	cdFlatDeficit            time.Duration // How far the flat cooldown mods take the cooldown below 0, where CD.Duration stops.
	GCDMultiplier            float64       // Scales DefaultCast.GCD, after the flat GCD mods and before haste and the GCD minimum; 1 = unmodified. Fed by SpellMod_GlobalCooldown_Pct.
	unscaledGCD              time.Duration // DefaultCast.GCD is rewritten from this on every GCD mod, so a direct write there is lost.
	RangeMultiplier          float64       // Scales MaxRange, after the flat range mods; 1 = unmodified. Fed by SpellMod_Range_Pct.
	DamageMultiplier         float64
	DamageMultiplierAdditive float64
	// Added to DamageMultiplierAdditive on direct hits and direct heals only. Unlike that bucket
	// this one is 0 when nothing feeds it.
	DirectDamageMultiplierAdditive float64
	DirectDamageMultiplier         float64 // Multiplies direct hits and direct heals only, 1 = unmodified. Fed by SpellMod_DirectDamageDone_Pct.
	CritMultiplierPct              float64 // Multiplies the base crit multiplier, 1 = unmodified. Fed by SpellMod_CritMultiplier_Pct.
	CritMultiplierAdditive         float64 // Additive critical damage bonus

	BonusBaseDamage       float64 // Certain items can increase the base damage of a spell e.g. https://www.wowhead.com/forever/item=28248/totem-of-the-void
	BonusCoefficient      float64 // EffectBonusCoefficient in the game files' SpellEffect table, "SP mod" on Wowhead (not necessarily shown there even if > 0)
	DirectVariance        float64 // EffectVariance of a spell-school hit or a heal: the direct amount, power included, lands within ±DirectVariance/2 of itself.
	TakenBonusCoefficient float64 // The coefficient the target's flat school bonus is taken at by a spell that sets no BonusCoefficient because its spell power goes through a weapon percent.

	// Multiplier for all threat generated by this effect.
	ThreatMultiplier float64

	FlatThreatBonus     float64 // Adds a fixed amount of threat to this spell, before multipliers.
	ThreatAPCoefficient float64 // Adds this share of the caster's attack power to the flat threat bonus, before multipliers.

	resultCache SpellResultCache
	resultSlice SpellResultSlice

	dots   DotArray
	aoeDot *Dot

	shields    ShieldArray
	selfShield *Shield

	// Per-target auras that are related to this spell, usually buffs or debuffs applied by the spell.
	RelatedAuraArrays LabeledAuraArrays
	RelatedDotSpell   *Spell
	RelatedSelfBuff   *Aura

	Source ActionSource
}

func (unit *Unit) OnSpellRegistered(handler SpellRegisteredHandler) {
	for _, spell := range unit.Spellbook {
		handler(spell)
	}
	unit.spellRegistrationHandlers = append(unit.spellRegistrationHandlers, handler)
}

// Registers a new spell to the unit. Returns the newly created spell.
func (unit *Unit) RegisterSpell(config SpellConfig) *Spell {
	if len(unit.Spellbook) > 1000 {
		panic(fmt.Sprintf("Over 1000 registered spells when registering %s! There is probably a spell being registered every iteration.", config.ActionID))
	}

	// Default the other damage multiplier to 1 if only one or the other is set.
	if config.DamageMultiplier != 0 && config.DamageMultiplierAdditive == 0 {
		config.DamageMultiplierAdditive = 1
	} else if config.DamageMultiplierAdditive != 0 && config.DamageMultiplier == 0 {
		config.DamageMultiplier = 1
	}

	if (config.DamageMultiplier != 0 || config.ThreatMultiplier != 0) && config.ProcMask == ProcMaskUnknown {
		panic("ProcMask for spell " + config.ActionID.String() + " not set")
	}

	if (config.DamageMultiplier != 0 || config.ThreatMultiplier != 0) && config.SpellSchool == SpellSchoolNone {
		panic("SpellSchool for spell " + config.ActionID.String() + " not set")
	}

	if config.Cast.CD.Timer != nil && config.Cast.CD.Duration == 0 {
		panic("Cast.CD w/o Duration specified for spell " + config.ActionID.String())
	}

	if config.Cast.SharedCD.Timer != nil && config.Cast.SharedCD.Duration == 0 {
		panic("Cast.SharedCD w/o Duration specified for spell " + config.ActionID.String())
	}

	if config.Charges > 0 && config.RechargeTime == 0 {
		panic("Spell has charges but no recharge time.")
	}

	if config.Cast.CastTime == nil {
		config.Cast.CastTime = func(spell *Spell) time.Duration {
			return spell.Unit.ApplyCastSpeedForSpell(spell.DefaultCast.CastTime, spell)
		}
	}

	if config.Source == SourceUnset {
		config.Source = unit.currentSource()
	}

	spell := &Spell{
		ActionID:       config.ActionID,
		Rank:           config.Rank,
		Unit:           unit,
		SpellSchool:    config.SpellSchool,
		DefenseType:    config.DefenseType,
		ProcMask:       config.ProcMask,
		Flags:          config.Flags,
		ClassSpellMask: config.ClassSpellMask,
		ClassFlags:     config.ClassFlags,

		MissileSpeed:       config.MissileSpeed,
		MissileMinDuration: config.MissileMinDuration,

		DefaultCast:        config.Cast.DefaultCast,
		CD:                 config.Cast.CD,
		SharedCD:           config.Cast.SharedCD,
		IgnoreHaste:        config.Cast.IgnoreHaste,
		ExtraCastCondition: config.ExtraCastCondition,
		CastRequirement:    config.CastRequirement,
		hasCastRequirement: config.CastRequirement != CastRequirement{},

		castTimeFn: config.Cast.CastTime,

		ApplyEffects: config.ApplyEffects,

		expectedInitialDamageInternal: config.ExpectedInitialDamage,
		expectedTickDamageInternal:    config.ExpectedTickDamage,

		BonusHitPercent:          config.BonusHitPercent,
		BonusCritPercent:         config.BonusCritPercent,
		BonusSpellDamage:         config.BonusSpellDamage,
		BonusExpertisePercent:    config.BonusExpertisePercent,
		CastTimeMultiplier:       1,
		CdMultiplier:             1,
		GCDMultiplier:            1,
		RangeMultiplier:          1,
		unscaledGCD:              config.Cast.DefaultCast.GCD,
		DamageMultiplier:         config.DamageMultiplier,
		DamageMultiplierAdditive: config.DamageMultiplierAdditive,
		DirectDamageMultiplier:   1,
		CritMultiplierPct:        1,
		CritMultiplierAdditive:   config.CritMultiplierAdditive,

		BonusBaseDamage:       config.BonusBaseDamage,
		BonusCoefficient:      config.BonusCoefficient,
		DirectVariance:        config.DirectVariance,
		TakenBonusCoefficient: config.TakenBonusCoefficient,

		ThreatMultiplier:    config.ThreatMultiplier,
		FlatThreatBonus:     config.FlatThreatBonus,
		ThreatAPCoefficient: config.ThreatAPCoefficient,

		splitSpellMetrics: make([][]SpellMetrics, max(1, config.MetricSplits)),
		splitHitRanges:    make([][]*hitRanges, max(1, config.MetricSplits)),

		RelatedAuraArrays: config.RelatedAuraArrays,
		RelatedDotSpell:   config.RelatedDotSpell,
		RelatedSelfBuff:   config.RelatedSelfBuff,

		Source: config.Source,

		charges:      config.Charges,
		MaxCharges:   config.Charges,
		RechargeTime: config.RechargeTime,

		resultCache: make(SpellResultCache, 1),
		resultSlice: make(SpellResultSlice, 0, 1),
	}

	spell.SchoolIndex = spell.SpellSchool.MultiplierSchool().SchoolIndex()

	if config.ManaCost.BaseCostPercent != 0 || config.ManaCost.FlatCost != 0 {
		spell.Cost = newManaCost(spell, config.ManaCost)
	} else if config.EnergyCost.Cost != 0 {
		spell.Cost = newEnergyCost(spell, config.EnergyCost, &unit.energyBar)
	} else if config.RageCost.Cost != 0 {
		spell.Cost = newRageCost(spell, config.RageCost)
	} else if config.FocusCost.Cost != 0 {
		spell.Cost = newFocusCost(spell, config.FocusCost)
	}

	// Create timer to track recharge time if none given
	if spell.CD.Timer == nil && spell.RechargeTime > 0 {
		spell.CD.Timer = spell.Unit.NewTimer()
	}

	spell.createDots(config.Dot, false)
	spell.createDots(config.Hot, true)
	spell.createShields(config.Shield)

	if spell.Cost != nil {
		spell.DefaultCast.Cost = float64(spell.Cost.BaseCost)
	}

	var emptyCast Cast

	if spell.DefaultCast == emptyCast && spell.Cost != nil {
		panic("Empty DefaultCast with a cost for spell " + config.ActionID.String())
	}

	if spell.DefaultCast.GCD <= 0 && spell.DefaultCast.CastTime == 0 {
		config.Cast.IgnoreHaste = true
	}

	if spell.DefaultCast == emptyCast {
		if config.ExtraCastCondition == nil && config.Cast.CD.Timer == nil && config.Cast.SharedCD.Timer == nil && !spell.hasCastRequirement {
			spell.castFn = spell.makeCastFuncAutosOrProcs()
		} else {
			spell.castFn = spell.makeCastFuncSimple()
		}
	} else {
		spell.castFn = spell.makeCastFunc(config.Cast)
	}

	if spell.ApplyEffects == nil {
		spell.ApplyEffects = func(*Simulation, *Unit, *Spell) {}
	}

	// Apply range constraints if requested. This is done after generating the castFn
	// for performance reasons, so that auto-attacks can be managed separately during
	// movement actions rather than constantly polling range checks.
	if (config.MinRange != 0) || (config.MaxRange != 0) {
		spell.MinRange = config.MinRange
		spell.MaxRange = config.MaxRange
		oldExtraCastCondition := spell.ExtraCastCondition
		spell.ExtraCastCondition = func(sim *Simulation, target *Unit) bool {
			if ((spell.MinRange != 0) && (spell.Unit.DistanceFromTarget < spell.MinRange)) || ((spell.MaxRange != 0) && (spell.Unit.DistanceFromTarget > spell.EffectiveMaxRange())) {
				/*if sim.Log != nil {
					sim.Log("Cannot cast spell %s, out of range!", spell.ActionID)
				}*/

				return false
			}

			return (oldExtraCastCondition == nil) || oldExtraCastCondition(sim, target)
		}
	}

	unit.Spellbook = append(unit.Spellbook, spell)

	for _, handler := range unit.spellRegistrationHandlers {
		handler(spell)
	}

	if unit.Env != nil && unit.Env.IsFinalized() {
		spell.finalize()
	}

	return spell
}

// Returns the first registered spell with the given ID, or nil if there are none.
func (unit *Unit) GetSpell(actionID ActionID) *Spell {
	for _, spell := range unit.Spellbook {
		if spell.ActionID.SameAction(actionID) {
			return spell
		}
	}
	return nil
}

// Retrieves an existing spell with the same ID as the config uses, or registers it if there is none.
func (unit *Unit) GetOrRegisterSpell(config SpellConfig) *Spell {
	registered := unit.GetSpell(config.ActionID)
	if registered == nil {
		return unit.RegisterSpell(config)
	} else {
		return registered
	}
}

func (spell *Spell) Dot(target *Unit) *Dot {
	if spell.dots == nil {
		if spell.RelatedDotSpell != nil {
			return spell.RelatedDotSpell.Dot(target)
		}
		return nil
	}
	return spell.dots.Get(target)
}

// The distance the range check allows: MaxRange after the range mods.
func (spell *Spell) EffectiveMaxRange() float64 {
	return spell.MaxRange * spell.RangeMultiplier
}

func (spell *Spell) CurDot() *Dot {
	if spell.dots == nil {
		if spell.RelatedDotSpell != nil {
			return spell.RelatedDotSpell.CurDot()
		}
		return nil
	}
	return spell.dots.Get(spell.Unit.CurrentTarget)
}
func (spell *Spell) AOEDot() *Dot {
	if spell.aoeDot == nil && spell.RelatedDotSpell != nil {
		return spell.RelatedDotSpell.AOEDot()
	}
	return spell.aoeDot
}

func (spell *Spell) hasDot(target *Unit) bool {
	return spell.Dot(target) != nil || spell.aoeDot != nil
}
func (spell *Spell) Hot(target *Unit) *Dot {
	if spell.dots == nil {
		return nil
	}
	return spell.dots.Get(target)
}
func (spell *Spell) CurHot() *Dot {
	if spell.dots == nil {
		return nil
	}
	return spell.dots.Get(spell.Unit.CurrentTarget)
}
func (spell *Spell) AOEHot() *Dot {
	return spell.aoeDot
}
func (spell *Spell) SelfHot() *Dot {
	return spell.aoeDot
}
func (spell *Spell) Shield(target *Unit) *Shield {
	if spell.shields == nil {
		return nil
	}
	return spell.shields.Get(target)
}
func (spell *Spell) SelfShield() *Shield {
	return spell.selfShield
}

func (spell *Spell) ApplyAllDots(sim *Simulation) {
	for _, target := range sim.Encounter.ActiveTargetUnits {
		spell.Dot(target).Apply(sim)
	}
}

func (spell *Spell) TickAllDotsOnce(sim *Simulation) {
	for _, target := range sim.Encounter.ActiveTargetUnits {
		spell.Dot(target).TickOnce(sim)
	}
}

func (spell *Spell) AnyDotsActive(sim *Simulation) bool {
	for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
		if spell.Dot(aoeTarget).IsActive() {
			return true
		}
	}

	return false
}

// Metrics for the current iteration
func (spell *Spell) CurDamagePerCast() float64 {
	if spell.SpellMetrics[0].Casts == 0 {
		return 0
	} else {
		casts := int32(0)
		damage := 0.0
		for _, opponent := range spell.Unit.GetAllOpponents() {
			casts += spell.SpellMetrics[opponent.UnitIndex].Casts
			damage += spell.SpellMetrics[opponent.UnitIndex].TotalDamage
		}
		return damage / float64(casts)
	}
}

// Current casts per minute
func (spell *Spell) CurCPM(sim *Simulation) float64 {
	if sim.CurrentTime <= 0 {
		return 0
	}
	casts := float64(spell.casts)
	minutes := float64(sim.CurrentTime) / float64(time.Minute)
	return casts / minutes
}

func (spell *Spell) finalize() {
	if len(spell.splitSpellMetrics) > 1 && spell.ActionID.Tag != 0 {
		panic(spell.ActionID.String() + " has split metrics and a non-zero tag, can only have one!")
	}
	for i := range spell.splitSpellMetrics {
		spell.splitSpellMetrics[i] = make([]SpellMetrics, len(spell.Unit.Env.AllUnits))
	}
	spell.SpellMetrics = spell.splitSpellMetrics[0]
	for i := range spell.splitHitRanges {
		spell.splitHitRanges[i] = make([]*hitRanges, len(spell.Unit.Env.AllUnits))
	}
	spell.hitRanges = spell.splitHitRanges[0]

	if spell.hasCastRequirement {
		spell.resolveCasterAuras()
		spell.equipment = spell.Unit.playerEquipment()
	}

	// Set the "static" "default" cost here
	if spell.Cost != nil {
		spell.DefaultCast.Cost = spell.Cost.GetCurrentCost()
	}
}

func (spell *Spell) reset(sim *Simulation) {
	for i := range spell.splitSpellMetrics {
		for j := range spell.SpellMetrics {
			spell.splitSpellMetrics[i][j] = SpellMetrics{}
		}
	}
	spell.casts = 0
	spell.lastAppliedAt = -NeverExpires
	if spell.rechargeTimer != nil {
		spell.rechargeTimer.Cancel(sim)
		spell.rechargeTimer = nil
		spell.charges = spell.MaxCharges
	}
}

func (spell *Spell) SetMetricsSplit(splitIdx int32) {
	spell.SpellMetrics = spell.splitSpellMetrics[splitIdx]
	spell.hitRanges = spell.splitHitRanges[splitIdx]
	spell.ActionID.Tag = splitIdx

	// Also set the tag on any dots to have them line up in the timeline
	if spell.dots != nil {
		for _, dot := range spell.dots {
			if dot != nil && dot.ActionID.SameActionIgnoreTag(spell.ActionID) {
				dot.ActionID.Tag = splitIdx
			}
		}
	}
	if spell.aoeDot != nil && spell.aoeDot.ActionID.SameActionIgnoreTag(spell.ActionID) {
		spell.aoeDot.ActionID.Tag = splitIdx
	}
}

func (spell *Spell) GetMetricSplitCount() int {
	return len(spell.splitSpellMetrics)
}

func (spell *Spell) doneIteration() {
	if spell.Flags.Matches(SpellFlagNoMetrics) {
		return
	}

	if len(spell.splitSpellMetrics) == 1 {
		spell.Unit.Metrics.addSpellMetrics(spell, spell.ActionID, spell.SpellMetrics, spell.hitRanges)
	} else {
		for i, spellMetrics := range spell.splitSpellMetrics {
			spell.Unit.Metrics.addSpellMetrics(spell, spell.ActionID.WithTag(int32(i)), spellMetrics, spell.splitHitRanges[i])
		}
	}
}

func (spell *Spell) rangesToRecord(target *Unit, value float64) *hitRanges {
	if value <= 0 || spell.Flags.Matches(SpellFlagNoMetrics) {
		return nil
	}
	hr := spell.hitRanges[target.UnitIndex]
	if hr == nil {
		hr = &hitRanges{}
		spell.hitRanges[target.UnitIndex] = hr
	}
	return hr
}

func (spell *Spell) recordHit(target *Unit, kind proto.HitKind, value float64) {
	if hr := spell.rangesToRecord(target, value); hr != nil {
		hr.kinds[kind].add(value)
	}
}

// Credits an absorbed or applied amount to the spell's shielding total and its shield range.
func (spell *Spell) RecordShielding(target *Unit, amount float64) {
	spell.SpellMetrics[target.UnitIndex].TotalShielding += amount
	spell.recordHit(target, proto.HitKind_HitKindShield, amount)
}

// Feeds the named range in this spell's details, for values the hit kinds do not cover.
func (spell *Spell) RecordCustomMetric(target *Unit, key string, value float64) {
	if hr := spell.rangesToRecord(target, value); hr != nil {
		hr.customRange(key).add(value)
	}
}

func (spell *Spell) HealthMetrics(target *Unit) *ResourceMetrics {
	if spell.healthMetrics == nil {
		spell.healthMetrics = make([]*ResourceMetrics, len(spell.Unit.AttackTables))
	}
	if spell.healthMetrics[target.UnitIndex] == nil {
		spell.healthMetrics[target.UnitIndex] = target.NewHealthMetrics(spell.ActionID)
	}
	return spell.healthMetrics[target.UnitIndex]
}

// The multiplier a cast applies to the spell's cooldown: 0 while a mod removes the cooldown.
func (spell *Spell) CooldownMultiplier() float64 {
	if spell.cdMultiplierZeroes > 0 {
		return 0
	}
	return spell.CdMultiplier
}

func (spell *Spell) ReadyAt() time.Duration {
	return BothTimersReadyAt(spell.CD.Timer, spell.SharedCD.Timer)
}

func (spell *Spell) IsReady(sim *Simulation) bool {
	if spell == nil {
		return false
	}
	return BothTimersReady(spell.CD.Timer, spell.SharedCD.Timer, sim)
}

func (spell *Spell) TimeToReady(sim *Simulation) time.Duration {
	return MaxTimeToReady(spell.CD.Timer, spell.SharedCD.Timer, sim)
}

// Returns whether a call to Cast() would be successful, without actually
// doing a cast.
func (spell *Spell) CanCast(sim *Simulation, target *Unit) bool {
	if !spell.canStart(sim, target) {
		return false
	}

	// While casting or channeling, no other action is possible
	if spell.Unit.Hardcast.Expires > sim.CurrentTime {
		//if sim.Log != nil {
		//	sim.Log("Cant cast because already casting/channeling")
		//}
		return false
	}

	if ((spell.DefaultCast.GCD > 0) || (spell.Flags.Matches(SpellFlagMCD) && spell.Unit.Rotation.inSequence)) && !spell.Unit.GCD.IsReady(sim) {
		//if sim.Log != nil {
		//	sim.Log("Cant cast because of GCD")
		//}
		return false
	}

	if !BothTimersReady(spell.CD.Timer, spell.SharedCD.Timer, sim) {
		//if sim.Log != nil {
		//	sim.Log("Cant cast because of CDs")
		//}
		return false
	}

	return spell.meetsCost(sim, false)
}

// The checks casting and queueing share ahead of their timer checks. While moving only instant casts
// are possible.
func (spell *Spell) canStart(sim *Simulation, target *Unit) bool {
	return spell != nil && spell.CanCompleteCast(sim, target, false) &&
		!spell.Flags.Matches(SpellFlagSwapped) &&
		(spell.Flags&SpellFlagCanCastWhileMoving != 0 || spell.DefaultCast.CastTime == 0 || !spell.Unit.Moving) &&
		(spell.MaxCharges == 0 || spell.charges > 0)
}

// The cost check runs after the other checks of its caller. A mana cost failure on a spell castable now
// marks the rotation pass as waiting on mana; one during the GCD, a cast or the spell's cooldown, of an
// off-GCD spell or inside the spell queue window, does not.
func (spell *Spell) meetsCost(sim *Simulation, logCastFailure bool) bool {
	if spell.Cost == nil || spell.Cost.MeetsRequirement(sim, spell) {
		return true
	}
	unit := spell.Unit
	if _, ok := spell.Cost.ResourceCostImpl.(*ManaCost); ok &&
		unit.GCD.IsReady(sim) && unit.Hardcast.Expires <= sim.CurrentTime && spell.IsReady(sim) {
		unit.costBlocked = true
	}
	if logCastFailure {
		return spell.castFailureHelper(sim, spell.Cost.CostFailureReason(sim, spell))
	}
	return false
}

// Returns whether the spell being cast can be completed, its cost aside: the unit and target may no longer
// qualify by the end of the cast.
func (spell *Spell) CanCompleteCast(sim *Simulation, target *Unit, logCastFailure bool) bool {
	if !spell.Unit.IsEnabled() {
		if logCastFailure {
			return spell.castFailureHelper(sim, "unit is disabled")
		}
		return false
	}

	if spell.Unit.PseudoStats.Incapacitated && !spell.Flags.Matches(SpellFlagCastWhileIncapacitated) {
		if logCastFailure {
			return spell.castFailureHelper(sim, "unit is incapacitated")
		}
		return false
	}

	if target == nil {
		if logCastFailure {
			return spell.castFailureHelper(sim, "target is not set")
		}
		return false
	}

	if !target.IsEnabled() {
		if logCastFailure {
			return spell.castFailureHelper(sim, "target is disabled")
		}
		return false
	}

	if spell.hasCastRequirement {
		if reason, _ := spell.castRequirementFailure(sim, target); reason != "" {
			if logCastFailure {
				return spell.castFailureHelper(sim, reason)
			}
			return false
		}
	}

	if spell.ExtraCastCondition != nil && !spell.ExtraCastCondition(sim, target) {
		//if sim.Log != nil {
		//	sim.Log("Cant cast because of extra condition")
		//}
		if logCastFailure {
			return spell.castFailureHelper(sim, "extra spell condition")
		}
		return false
	}

	return true
}

func (spell *Spell) Cast(sim *Simulation, target *Unit) bool {
	if spell.DefaultCast.EffectiveTime() > 0 {
		spell.Unit.CancelQueuedSpell(sim)
	}
	if target == nil {
		target = spell.Unit.CurrentTarget
	}
	return spell.castFn(sim, target)
}

func (spell *Spell) CastOnAllOtherTargets(sim *Simulation, mainTarget *Unit) {
	for _, target := range sim.Encounter.ActiveTargetUnits {
		if target != mainTarget {
			spell.Cast(sim, target)
		}
	}
}

// Skips the actual cast and applies spell effects immediately.
func (spell *Spell) SkipCastAndApplyEffects(sim *Simulation, target *Unit) {
	if sim.Log != nil && !spell.Flags.Matches(SpellFlagNoLogs) {
		spell.Unit.Log(sim, "Casting %s (Cost = %0.03f, Cast Time = %s)",
			spell.ActionID, spell.DefaultCast.Cost, time.Duration(0))
		spell.Unit.Log(sim, "Completed cast %s", spell.ActionID)
	}
	spell.applyEffects(sim, target)
}

func (spell *Spell) applyEffects(sim *Simulation, target *Unit) {
	spell.SpellMetrics[target.UnitIndex].Casts++
	spell.casts++
	spell.lastAppliedAt = sim.CurrentTime

	// Not sure if we want to split this flag into its own?
	// Both are used to optimize away unneccesery calls and 99%
	// of the time are gonna be used together. For now just in one
	if !spell.Flags.Matches(SpellFlagNoOnCastComplete) {
		spell.Unit.OnApplyEffects(sim, target, spell)
	}

	spell.ApplyEffects(sim, target, spell)

	if spell == spell.Unit.rotationCast.spell {
		spell.Unit.endRotationCast()
		spell.Unit.Wand.castCompleted(sim, spell)
	}
}

func (spell *Spell) ApplyAOEThreatIgnoreMultipliers(threatAmount float64) {
	for _, target := range spell.Unit.Env.GetActiveTargetUnits() {
		spell.SpellMetrics[target.UnitIndex].TotalThreat += threatAmount
	}
}

// Threat no single enemy draws, such as a heal's or a resource gain's, split between the targets in
// combat at the time.
func (spell *Spell) applySplitThreat(sim *Simulation, threat float64) {
	if targets := len(sim.Encounter.ActiveTargetUnits); targets > 0 {
		spell.ApplyAOEThreatIgnoreMultipliers(threat / float64(targets))
	}
}
func (spell *Spell) ApplyAOEThreat(threatAmount float64) {
	spell.ApplyAOEThreatIgnoreMultipliers(threatAmount * spell.unitThreatMultiplier())
}

func (spell *Spell) unitThreatMultiplier() float64 {
	return spell.Unit.PseudoStats.ThreatMultiplier * spell.Unit.PseudoStats.SchoolThreatMultiplier[spell.SchoolIndex]
}

func (spell *Spell) finalizeExpectedDamage(result *SpellResult) {
	result.inUse = false
}
func (spell *Spell) expectedDamage(sim *Simulation, target *Unit, calculator ExpectedDamageCalculator, useSnapshot bool) float64 {
	spell.expecting = true
	result := calculator(sim, target, spell, useSnapshot)
	spell.expecting = false
	spell.finalizeExpectedDamage(result)
	return result.Damage
}
func (spell *Spell) ExpectedInitialDamage(sim *Simulation, target *Unit) float64 {
	return spell.expectedDamage(sim, target, spell.expectedInitialDamageInternal, false)
}
func (spell *Spell) ExpectedTickDamage(sim *Simulation, target *Unit) float64 {
	return spell.expectedDamage(sim, target, spell.expectedTickDamageInternal, false)
}
func (spell *Spell) ExpectedTickDamageFromCurrentSnapshot(sim *Simulation, target *Unit) float64 {
	return spell.expectedDamage(sim, target, spell.expectedTickDamageInternal, true)
}

// The damage multiplier of a critical strike. The base comes from the DefenseType: spells crit
// for 150%, melee and ranged attacks for 200%. Multiplicative modifiers (the slaying
// talents on the attack table, SpellMod_CritMultiplier_Pct) scale that base; additive
// ones (Ruin, Impale, SpellMod_CritMultiplier_Flat) scale only the bonus part above 100%.
// https://web.archive.org/web/20081014064638/http://elitistjerks.com/f31/t12595-relentless_earthstorm_diamond_-_melee_only/p4/
func (spell *Spell) CritDamageMultiplier(at *AttackTable) float64 {
	var base float64
	switch spell.DefenseType {
	case DefenseTypeNone:
		panic(fmt.Sprintf("CritDamageMultiplier() called for %s which has no DefenseType", spell.ActionID))
	case DefenseTypeMagic:
		base = 1.5
	default:
		base = 2.0
	}
	// Heals roll their crit without an attack table.
	tableMultiplier := 1.0
	if at != nil {
		tableMultiplier = at.CritMultiplier
	}
	return (base*spell.CritMultiplierPct*spell.Unit.PseudoStats.CritDamageMultiplier*tableMultiplier-1)*(spell.CritMultiplierAdditive+1) + 1
}

// Time until either the cast is finished or GCD is ready again, whichever is longer
func (spell *Spell) EffectiveCastTime() time.Duration {
	// TODO: this is wrong for spells like shadowfury, that have a GCD of less than 1s
	return max(spell.Unit.SpellGCD(),
		spell.Unit.ApplyCastSpeedForSpell(spell.DefaultCast.EffectiveTime(), spell))
}

// Time until the cast is finished (ignoring GCD)
func (spell *Spell) CastTime() time.Duration {
	return spell.castTimeFn(spell)
}

func (spell *Spell) TravelTime() time.Duration {
	if spell.MissileSpeed == 0 {
		return spell.MissileMinDuration
	}
	return max(spell.MissileMinDuration, time.Duration(float64(time.Second)*spell.Unit.DistanceFromTarget/spell.MissileSpeed))
}

// Returns true if the given mask matches the spell mask
func (spell *Spell) Matches(mask int64) bool {
	return spell.ClassSpellMask&mask > 0
}

// Returns true if the given class flags name this spell
func (spell *Spell) MatchesFlags(f ClassFlags) bool {
	return f.Matches(spell.ClassFlags)
}

// Whether a modifier or trigger restricted to these class flags reaches the spell: zero flags mean
// every spell.
func (f ClassFlags) Reaches(spell *Spell) bool {
	return f.IsZero() || spell.MatchesFlags(f)
}

// Handles computing the cost of spells and checking whether the Unit
// meets them.
type ResourceCostImpl interface {
	// Whether the Unit associated with the spell meets the resource cost
	// requirements to cast the spell.
	MeetsRequirement(*Simulation, *Spell) bool

	// Returns a message for when the cast fails due to lack of resources.
	CostFailureReason(*Simulation, *Spell) string

	// Subtracts the resources used from a cast from the Unit.
	SpendCost(*Simulation, *Spell)

	// Space for handling refund mechanics. Not all spells provide refunds.
	IssueRefund(*Simulation, *Spell)
}

type SpellCost struct {
	BaseCost                int32   // The base power cost before all modifiers.
	FlatModifier            int32   // Flat value added to base cost before pct mods
	PercentModifier         float64 // Multiplier for cost, as of MoP a float
	AdditivePercentModifier float64 // Additive pct cost bucket: defaults to 1, additive pct mods sum into it
	percentModifierZeroes   int32   // Active percent cost mods of -100% or lower; the cost is 0 while any is active.
	spell                   *Spell
	ResourceCostImpl
}

func (sc *SpellCost) ApplyCostModifiers(cost int32) float64 {
	if sc.percentModifierZeroes > 0 {
		return 0
	}
	spell := sc.spell
	cost = max(0, cost+sc.FlatModifier)
	cost = max(0, cost*spell.Unit.PseudoStats.SpellCostPercentModifier/100)
	return max(0, float64(cost)*sc.PercentModifier*sc.AdditivePercentModifier)
}

// Get power cost after all modifiers.
func (sc *SpellCost) GetCurrentCost() float64 {
	return sc.ApplyCostModifiers(sc.BaseCost)
}

func (spell *Spell) IssueRefund(sim *Simulation) {
	spell.Cost.IssueRefund(sim, spell)
}

func (spell *Spell) ConsumeCharge(sim *Simulation) {
	if spell.MaxCharges == 0 {
		return
	}

	if spell.charges == 0 {
		panic("Trying to consume charge, but spell has no charges left.")
	}

	spell.charges--

	// No recharge in progress yet, start timer
	if spell.rechargeTimer == nil {
		spell.scheduleRechargeAction(sim)
	}
}

func (spell *Spell) scheduleRechargeAction(sim *Simulation) {
	spell.rechargeTimer = &PendingAction{
		NextActionAt: sim.CurrentTime + spell.RechargeTime,
		Priority:     ActionPriorityAuto,
		OnAction: func(sim *Simulation) {
			spell.RefreshCharge(sim)
			spell.rechargeTimer = nil
			if spell.charges < spell.MaxCharges {
				spell.scheduleRechargeAction(sim)
			}
		},
	}

	sim.AddPendingAction(spell.rechargeTimer)
}

func (spell *Spell) GetNumCharges() int {
	return spell.charges
}

// Calculates the time until the next charge is available.
// Will return 0 if no recharge is in progress
func (spell *Spell) NextChargeIn(sim *Simulation) time.Duration {
	if spell.rechargeTimer == nil {
		return 0
	}

	return spell.rechargeTimer.NextActionAt - sim.CurrentTime
}

// Refreshes a charge of the spell
// Can be called if the spell has max charges
func (spell *Spell) RefreshCharge(sim *Simulation) {
	if spell.charges == spell.MaxCharges {
		return
	}

	spell.charges++
}

package core

import (
	"slices"
	"strconv"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type Encounter struct {
	Duration          time.Duration
	DurationVariation time.Duration
	AllTargets        []*Target
	ActiveTargets     []*Target
	AllTargetUnits    []*Unit
	ActiveTargetUnits []*Unit

	targetConfigs   []*proto.Target // The config of each target in AllTargets, nil for the stand-in target of an encounter without any.
	strandedSwings  []strandedSwing // Raid units whose auto attacks stopped because no target was left to switch to.
	strandedScratch []strandedSwing // The stranded swings being brought back, kept so bringing them back allocates nothing.
	resetting       bool            // While the environment resets, from its first target to its raid.
	design          *encounterDesign

	ExecuteProportion_20 float64
	ExecuteProportion_35 float64

	// Every kind of area the fight is in, deduplicated.
	AreaTypes []proto.AreaType

	// Value to multiply by, for damage spells which are subject to the aoe cap.
	aoeCapMultiplier float64
}

func areaTypeSet(areaTypes []proto.AreaType) []proto.AreaType {
	var set []proto.AreaType
	for _, areaType := range areaTypes {
		if areaType != proto.AreaType_AreaTypeUnknown && !slices.Contains(set, areaType) {
			set = append(set, areaType)
		}
	}
	return set
}

func (encounter *Encounter) InArea(areaType proto.AreaType) bool {
	return slices.Contains(encounter.AreaTypes, areaType)
}

func NewEncounter(options *proto.Encounter) Encounter {
	options.ExecuteProportion_35 = max(options.ExecuteProportion_35, options.ExecuteProportion_20)
	totalTargetCount := max(len(options.Targets), 1)

	encounter := Encounter{
		Duration:             DurationFromSeconds(options.Duration),
		DurationVariation:    DurationFromSeconds(options.DurationVariation),
		ExecuteProportion_20: max(options.ExecuteProportion_20, 0),
		ExecuteProportion_35: max(options.ExecuteProportion_35, 0),
		AreaTypes:            areaTypeSet(options.AreaTypes),
		AllTargets:           make([]*Target, 0, totalTargetCount),
		ActiveTargets:        make([]*Target, 0, totalTargetCount),
		AllTargetUnits:       make([]*Unit, 0, totalTargetCount),
		ActiveTargetUnits:    make([]*Unit, 0, totalTargetCount),
		targetConfigs:        make([]*proto.Target, 0, totalTargetCount),
	}

	for targetIndex, targetOptions := range options.Targets {
		target := NewTarget(targetOptions, int32(targetIndex))
		encounter.AllTargets = append(encounter.AllTargets, target)
		encounter.AllTargetUnits = append(encounter.AllTargetUnits, &target.Unit)
		encounter.targetConfigs = append(encounter.targetConfigs, targetOptions)

		if target.IsEnabled() {
			encounter.ActiveTargets = append(encounter.ActiveTargets, target)
			encounter.ActiveTargetUnits = append(encounter.ActiveTargetUnits, &target.Unit)
		}
	}

	if len(encounter.AllTargets) == 0 {
		// Add a dummy target. The only case where targets aren't specified is when
		// computing character stats, and targets won't matter there.
		target := NewTarget(&proto.Target{}, 0)
		encounter.AllTargets = append(encounter.AllTargets, target)
		encounter.ActiveTargets = append(encounter.ActiveTargets, target)
		encounter.AllTargetUnits = append(encounter.AllTargetUnits, &target.Unit)
		encounter.ActiveTargetUnits = append(encounter.ActiveTargetUnits, &target.Unit)
		encounter.targetConfigs = append(encounter.targetConfigs, nil)
	}

	encounter.design = newEncounterDesign(&encounter, options.Design)
	encounter.updateAOECapMultiplier()

	return encounter
}

// n disabled targets built from template, without the AI of a preset with the template's id, appended
// to the encounter. Only NewEncounter may call it: the construct phase lists the encounter's units and
// points the raid at them right after.
func (encounter *Encounter) allocatePoolTargets(template *proto.Target, n int) []*Target {
	pool := make([]*Target, 0, n)
	for range n {
		target := newTarget(template, int32(len(encounter.AllTargets)))
		target.enabled = false
		encounter.AllTargets = append(encounter.AllTargets, target)
		encounter.AllTargetUnits = append(encounter.AllTargetUnits, &target.Unit)
		encounter.targetConfigs = append(encounter.targetConfigs, template)
		pool = append(pool, target)
	}
	return pool
}

// Target 0 when it is enabled, else the enabled target with the lowest index, else nil.
func (encounter *Encounter) preferredActiveTarget() *Target {
	for _, target := range encounter.AllTargets {
		if target.IsEnabled() {
			return target
		}
	}
	return nil
}

// Rebuilds the active target lists in index order, which removing a target scrambles.
func (encounter *Encounter) resetActiveTargetOrder() {
	encounter.ActiveTargets = encounter.ActiveTargets[:0]
	encounter.ActiveTargetUnits = encounter.ActiveTargetUnits[:0]
	for _, target := range encounter.AllTargets {
		if target.IsEnabled() {
			encounter.ActiveTargets = append(encounter.ActiveTargets, target)
			encounter.ActiveTargetUnits = append(encounter.ActiveTargetUnits, &target.Unit)
		}
	}
	encounter.updateAOECapMultiplier()
}

func (encounter *Encounter) AOECapMultiplier() float64 {
	return encounter.aoeCapMultiplier
}

// With no active target the division gives +Inf, which the min keeps at 1.
func (encounter *Encounter) updateAOECapMultiplier() {
	encounter.aoeCapMultiplier = min(20/float64(len(encounter.ActiveTargets)), 1)
}

func (encounter *Encounter) addActiveTarget(target *Target) {
	if !slices.Contains(encounter.AllTargets, target) {
		panic("Target was not defined during the construction phase of the encounter!")
	}

	if slices.Contains(encounter.ActiveTargets, target) {
		panic("Target is already present in active target list!")
	}

	encounter.ActiveTargets = append(encounter.ActiveTargets, target)
	encounter.ActiveTargetUnits = append(encounter.ActiveTargetUnits, &target.Unit)
	encounter.updateAOECapMultiplier()
}

func (encounter *Encounter) removeInactiveTarget(target *Target) {
	if idx := slices.Index(encounter.ActiveTargets, target); idx != -1 {
		encounter.ActiveTargets = removeBySwappingToBack(encounter.ActiveTargets, idx)
		encounter.ActiveTargetUnits = removeBySwappingToBack(encounter.ActiveTargetUnits, idx)
	} else {
		panic("Target is not present in active target list!")
	}

	encounter.updateAOECapMultiplier()
}

func (encounter *Encounter) doneIteration(sim *Simulation) {
	for _, target := range encounter.AllTargets {
		target.doneIteration(sim)
	}
}

func (encounter *Encounter) GetMetricsProto() *proto.EncounterMetrics {
	metrics := &proto.EncounterMetrics{
		Targets: make([]*proto.UnitMetrics, len(encounter.AllTargets)),
	}

	for idx, target := range encounter.AllTargets {
		metrics.Targets[idx] = target.GetMetricsProto()
	}

	return metrics
}

// Target is an enemy/boss that can be the target of player attacks/spells.
type Target struct {
	Unit

	AI TargetAI

	health healthCurve
}

func NewTarget(options *proto.Target, targetIndex int32) *Target {
	target := newTarget(options, targetIndex)

	preset := GetPresetTargetWithID(options.Id)
	if preset != nil && preset.AI != nil {
		target.AI = preset.AI()
	}

	return target
}

func newTarget(options *proto.Target, targetIndex int32) *Target {
	unitStats := stats.Stats{}
	if options.Stats != nil {
		unitStats = stats.FromProtoArray(options.Stats)
	}

	target := &Target{
		Unit: Unit{
			Type:           EnemyUnit,
			Index:          targetIndex,
			Label:          "Target " + strconv.Itoa(int(targetIndex)+1),
			Level:          options.Level,
			MobType:        options.MobType,
			CreatureFamily: options.CreatureFamily,
			auraTracker:    newAuraTracker(),
			stats:          unitStats,
			PseudoStats:    stats.NewPseudoStats(),
			Metrics:        NewUnitMetrics(),

			StatDependencyManager: stats.NewStatDependencyManager(),
			ReactionTime:          BossGCD,
			enabled:               !options.DisabledAtStart,
		},
	}
	defaultRaidBossLevel := int32(CharacterLevel + 3)
	target.GCD = target.NewTimer()
	target.RotationTimer = target.NewTimer()
	if target.Level == 0 {
		target.Level = defaultRaidBossLevel
	}

	// Default Crit chance for NPCs depends only on their level relative to the level of their
	// player target. If there is a need to model a custom Crit % for a Target, this can be
	// accomplished by specifying the extra (or reduced) Crit within the CritRating field of the
	// proto stats array. Any specified CritRating will be applied as an OFFSET to the default
	// level-based values, rather than as a replacement.
	target.stats[stats.PhysicalCritPercent] += UnitLevelFloat64(target.Level-CharacterLevel, 5.0, 5.2, 5.4, 5.6)
	target.addUniversalStatDependencies()

	if target.Level == 63 && options.SuppressDodge {
		// Sunwell boss Dodge Suppression. -20% dodge and -5% miss chance.
		target.PseudoStats.DodgeReduction += 0.2
		target.PseudoStats.IncreasedMissChance -= 0.05
	}

	target.PseudoStats.CanBlock = true
	target.PseudoStats.CanParry = true
	target.PseudoStats.CanCrush = options.CanCrush
	target.PseudoStats.ParryHaste = options.ParryHaste
	target.PseudoStats.InFrontOfTarget = true
	target.PseudoStats.DamageSpread = options.DamageSpread

	return target
}

func (target *Target) Reset(sim *Simulation) {
	target.Unit.reset(sim, nil)
	target.CurrentTarget = target.defaultTarget

	if !target.IsEnabled() && (target.CurrentTarget != nil) {
		target.retargetOwnTarget()
	}

	target.SetGCDTimer(sim, 0)

	if target.AI != nil {
		target.AI.Reset(sim)
	}
}

func (target *Target) Enable(sim *Simulation) {
	if target.defaultTarget != nil {
		target.defaultTarget.CurrentTarget = &target.Unit
		target.CurrentTarget = target.defaultTarget
	}

	target.AutoAttacks.EnableAutoSwing(sim)

	// Randomize GCD and swing timings to prevent fake APL-Haste couplings.
	target.ExtendGCDUntil(sim, sim.CurrentTime+DurationFromSeconds(sim.RandomFloat("Specials Timing")*BossGCD.Seconds()))
	target.AutoAttacks.RandomizeMeleeTiming(sim)

	if !target.IsEnabled() {
		target.enabled = true
		sim.Encounter.addActiveTarget(target)
		sim.Encounter.returnStrandedRaid(sim)
	}
}

func (sim *Simulation) EnableTargetUnit(targetUnit *Unit) {
	if targetUnit.Type != EnemyUnit {
		panic("Unit is not an enemy target!")
	}

	sim.Encounter.AllTargets[targetUnit.Index].Enable(sim)
}

func (target *Target) Disable(sim *Simulation, expireAuras bool) {
	if !target.IsEnabled() {
		return
	}

	// A hardcast still running would land at its end, after the target has gone. The interrupt comes
	// before the GCD stops, since it sets the GCD again.
	target.Interrupt(sim)
	target.CancelGCDTimer(sim)
	target.AutoAttacks.CancelAutoSwing(sim)

	target.enabled = false
	sim.Encounter.removeInactiveTarget(target)

	if expireAuras {
		target.auraTracker.expireAll(sim)
	}

	if target.CurrentTarget != nil {
		target.retargetOwnTarget()
	}

	sim.Encounter.moveRaidOffTarget(sim, target)
}

// Points the unit this target attacks at the next active target, or leaves it on this one when none
// is left, and clears this target's own target.
func (target *Target) retargetOwnTarget() {
	if next := target.NextActiveTarget(); next != nil {
		target.CurrentTarget.CurrentTarget = &next.Unit
	}
	target.CurrentTarget = nil
}

// A swing a raid unit had running when its last target went away.
type strandedSwing struct {
	unit   *Unit
	melee  bool
	ranged bool
}

// Moves every raid unit on the disabled target to the preferred active target and has its rotation look
// again, since what it waited on is gone. With none left, the unit
// keeps the disabled target, so its target is never nil, and its auto attacks stop until a target comes.
// While the environment resets, the raid is left alone: its own reset points it at its first target.
func (encounter *Encounter) moveRaidOffTarget(sim *Simulation, target *Target) {
	if encounter.resetting {
		return
	}
	preferred := encounter.preferredActiveTarget()
	for _, unit := range sim.Raid.AllUnits {
		if unit.CurrentTarget != &target.Unit {
			continue
		}
		if preferred != nil {
			unit.CurrentTarget = &preferred.Unit
			wakeRotation(sim, unit)
			continue
		}
		aa := &unit.AutoAttacks
		swing := strandedSwing{unit: unit, melee: aa.mh.enabled || aa.oh.enabled, ranged: aa.ranged.enabled}
		if swing.melee || swing.ranged {
			encounter.strandedSwings = append(encounter.strandedSwings, swing)
			aa.CancelAutoSwing(sim)
		}
	}
}

// Brings back the raid units a disabled target stranded once a target is active again: a unit on a
// disabled enemy moves to the preferred active target, the auto attacks moveRaidOffTarget stopped start
// again, and each of these units looks at its rotation again.
func (encounter *Encounter) returnStrandedRaid(sim *Simulation) {
	if encounter.resetting {
		return
	}
	// A rotation woken here may take the new target away again, which records fresh swings, or bring one
	// back, which runs this again. The swings move into the scratch slice, which this call holds until it
	// is done, so a nested call works on its own copy and the slices are reused across iterations.
	swings := append(encounter.strandedScratch[:0], encounter.strandedSwings...)
	encounter.strandedScratch = nil
	encounter.strandedSwings = encounter.strandedSwings[:0]
	preferred := encounter.preferredActiveTarget()
	for _, unit := range sim.Raid.AllUnits {
		if unit.CurrentTarget == nil || unit.CurrentTarget.Type != EnemyUnit || unit.CurrentTarget.IsEnabled() {
			continue
		}
		unit.CurrentTarget = &preferred.Unit
		if !hasStrandedSwing(swings, unit) {
			wakeRotation(sim, unit)
		}
	}
	for _, swing := range swings {
		if !swing.unit.IsEnabled() {
			continue
		}
		if swing.melee {
			swing.unit.AutoAttacks.EnableMeleeSwing(sim)
		}
		if swing.ranged {
			swing.unit.AutoAttacks.EnableRangedSwing(sim)
		}
		wakeRotation(sim, swing.unit)
	}
	encounter.strandedScratch = swings[:0]
}

func hasStrandedSwing(swings []strandedSwing, unit *Unit) bool {
	for _, swing := range swings {
		if swing.unit == unit {
			return true
		}
	}
	return false
}

func wakeRotation(sim *Simulation, unit *Unit) {
	if unit.IsEnabled() && unit.Rotation != nil && !sim.Options.Interactive {
		unit.ReactToEvent(sim, false, true)
	}
}

func (sim *Simulation) DisableTargetUnit(targetUnit *Unit, expireAuras bool) {
	if targetUnit.Type != EnemyUnit {
		panic("Unit is not an enemy target!")
	}

	sim.Encounter.AllTargets[targetUnit.Index].Disable(sim, expireAuras)
}

// The first enabled target after this one in index order, wrapping around and ending on this one; nil
// when no target is enabled.
func (target *Target) NextActiveTarget() *Target {
	count := target.Env.TotalTargetCount()
	for step := int32(1); step <= count; step++ {
		if next := target.Env.GetTargetByIndex((target.Index + step) % count); next.IsEnabled() {
			return next
		}
	}
	return nil
}

// The first enabled target before this one in index order, wrapping around and ending on this one; nil
// when no target is enabled.
func (target *Target) PreviousActiveTarget() *Target {
	count := target.Env.TotalTargetCount()
	for step := int32(1); step <= count; step++ {
		if prev := target.Env.GetTargetByIndex((target.Index - step + count) % count); prev.IsEnabled() {
			return prev
		}
	}
	return nil
}

func (target *Target) GetMetricsProto() *proto.UnitMetrics {
	metrics := target.Metrics.ToProto()
	metrics.Name = target.Label
	metrics.UnitIndex = target.UnitIndex
	metrics.Auras = target.auraTracker.GetMetricsProto()
	return metrics
}

type DynamicDamageDoneByCaster func(sim *Simulation, spell *Spell, attackTable *AttackTable) float64
type DynamicThreatDoneByCaster DynamicDamageDoneByCaster

// The crit chance, in percentage points, a target-side aura adds to one of its caster's spells.
type DynamicCritChanceByCaster func(spell *Spell) float64

// Holds cached values for outcome/damage calculations, for a specific attacker+defender pair.
// These are updated dynamically when attacker or defender stats change.
type AttackTable struct {
	Attacker *Unit
	Defender *Unit

	BaseSpellMissChance float64
	BaseCrushChance     float64

	AuraCritSuppression float64 // The most of the attacker's crit beyond base and Agility crit that a defender 3 or more levels above takes off.

	MH     AttackTableHand
	OH     AttackTableHand
	Ranged AttackTableHand

	// Raised only by A_MOD_CRIT_PERCENT_VERSUS effects, such as effect 1 of Humanoid Slaying 19151 and Monster
	// Slaying 24294. A_MOD_DAMAGE_DONE_VERSUS effects, Beast Slaying 20557 among them, raise damage alone.
	CritMultiplier float64

	DamageDealtMultiplier       float64 // attacker buff, applied in applyAttackerModifiers()
	DamageTakenMultiplier       float64 // defender debuff, applied in applyTargetModifiers()
	HealingDealtMultiplier      float64
	IgnoreArmor                 bool    // Ignore defender's armor for specifically this attacker's attacks
	ArmorIgnoreFactor           float64 // Percentage of armor to ignore for this attacker's attacks
	BonusSpellCritPercent       float64 // Analagous to Defender.PseudoStats.BonusSpellCritPercentTaken, but only for this attacker specifically
	RangedDamageTakenMultiplier float64
	// This is for "Apply Aura: Mod Damage Done By Caster" effects.
	// If set, the damage taken multiplier is multiplied by the callbacks result.
	DamageDoneByCasterMultiplier DynamicDamageDoneByCaster

	MobTypeBonusStats map[proto.MobType]stats.Stats // Bonus stats against mobs of the current type.

	// When you need more then 1 active, default to using the above one
	// Used with EnableDamageDoneByCaster/DisableDamageDoneByCaster
	DamageDoneByCasterExtraMultiplier []DynamicDamageDoneByCaster

	// The crit chance target-side auras add to this attacker's spells alone, indexed the way
	// DamageDoneByCasterExtraMultiplier is.
	// Used with EnableCritChanceByCaster/DisableCritChanceByCaster
	CritChanceByCasterBonus []DynamicCritChanceByCaster
}

func NewAttackTable(attacker *Unit, defender *Unit) *AttackTable {
	table := &AttackTable{
		Attacker: attacker,
		Defender: defender,

		CritMultiplier:              1,
		DamageDealtMultiplier:       1,
		DamageTakenMultiplier:       1,
		RangedDamageTakenMultiplier: 1,
		HealingDealtMultiplier:      1,

		MobTypeBonusStats: make(map[proto.MobType]stats.Stats),
	}

	if defender.Type == EnemyUnit {
		levelDiff := combatLevelDiff(attacker, defender)

		table.BaseSpellMissChance = spellMissChance(levelDiff)
		// Era logs show no aura crit suppression against targets 1 or 2 levels above.
		if levelDiff >= 3 {
			table.AuraCritSuppression = 0.018
		}

		table.setWeapons(nil, nil, nil)
	} else {
		levelDiff := combatLevelDiff(defender, attacker)
		levelDelta := -0.002 * float64(levelDiff)

		table.BaseSpellMissChance = spellMissChance(-levelDiff)
		table.BaseCrushChance = TernaryFloat64(levelDiff >= 3, 0.15, 0)

		table.MH = AttackTableHand{
			BaseMissChance:      0.05 + levelDelta,
			BaseDodgeChance:     levelDelta,
			BaseParryChance:     0.05 + levelDelta,
			BaseBlockChance:     0.05 + levelDelta,
			BaseSpellMissChance: table.BaseSpellMissChance,
		}
		table.OH = table.MH
		table.Ranged = table.MH
	}

	return table
}

func combatLevelDiff(attacker *Unit, defender *Unit) int32 {
	return min(defender.Level-attacker.Level, 3)
}

// The level-based spell miss for the defender's level minus the attacker's: 4% at equal level, one point
// per level either side, 0% at -4 and below, and 17% at +3 and above.
func spellMissChance(levelDiff int32) float64 {
	if levelDiff >= 3 {
		return 0.17
	}
	return max(0, 0.04+0.01*float64(levelDiff))
}

// The parts of an attack table that depend on the weapon skill of the attacking hand.
type AttackTableHand struct {
	BaseMissChance       float64
	BaseDodgeChance      float64
	BaseParryChance      float64
	BaseBlockChance      float64
	BaseGlanceChance     float64
	GlanceMultiplierMin  float64
	GlanceMultiplierMax  float64
	MeleeCritSuppression float64
	BaseSpellMissChance  float64 // The spell table's miss with the level difference taken from the hand's skill.
}

func newAttackTableHand(attacker *Unit, defender *Unit, weapon *Item) AttackTableHand {
	levelDiff := combatLevelDiff(attacker, defender)
	bonus := GetWeaponSkill(attacker, weapon)
	skillDiff := float64(levelDiff*5) - bonus
	// Parry, glance chance and crit suppression count skill only up to the level-based maximum.
	baseSkillDiff := float64(levelDiff*5) - min(bonus, 0)
	// The spell table's level difference, capped at +3, less skill above the level-based maximum / 5,
	// truncated toward zero, so 5 points of skill short of the maximum count as one more target level and
	// 5 above it as one less.
	skillLevelDiff := int32(float64(levelDiff) - bonus/5)

	hand := AttackTableHand{
		BaseDodgeChance:     0.05 + skillDiff*0.001,
		BaseBlockChance:     min(0.05, 0.05+skillDiff*0.001),
		BaseGlanceChance:    Clamp(0.1+baseSkillDiff*0.02, 0, 1),
		GlanceMultiplierMin: Clamp(1.3-0.05*skillDiff, 0.01, 0.91),
		GlanceMultiplierMax: Clamp(1.2-0.03*skillDiff, 0.2, 0.99),
		BaseSpellMissChance: spellMissChance(skillLevelDiff),
	}
	if skillDiff > 10 {
		hand.BaseMissChance = 0.05 + skillDiff*0.002
	} else {
		hand.BaseMissChance = 0.05 + skillDiff*0.001
	}
	if baseSkillDiff > 10 {
		hand.BaseParryChance = 0.05 + baseSkillDiff*0.006
	} else {
		hand.BaseParryChance = 0.05 + baseSkillDiff*0.001
	}
	if baseSkillDiff > 0 {
		hand.MeleeCritSuppression = baseSkillDiff * 0.002
	} else {
		hand.MeleeCritSuppression = baseSkillDiff * 0.0004
	}
	return hand
}

// Recomputes the hand entries against an enemy defender from the weapons in each hand.
func (at *AttackTable) setWeapons(mh *Item, oh *Item, ranged *Item) {
	at.MH = newAttackTableHand(at.Attacker, at.Defender, mh)
	at.OH = newAttackTableHand(at.Attacker, at.Defender, oh)
	at.Ranged = newAttackTableHand(at.Attacker, at.Defender, ranged)
}

// Hand returns the entry of the hand the spell attacks with.
func (at *AttackTable) Hand(spell *Spell) *AttackTableHand {
	if spell.ProcMask.Matches(ProcMaskRanged) {
		return &at.Ranged
	}
	if spell.ProcMask.Matches(ProcMaskMeleeOH) && !spell.ProcMask.Matches(ProcMaskMeleeMH) {
		return &at.OH
	}
	return &at.MH
}

// Puts a handler in slot index of one of an attack table's caster slices, which is made maxIndex long
// on first use.
func setCasterSlot[T any](slots *[]T, index int, maxIndex int, handler T) {
	if *slots == nil {
		*slots = make([]T, maxIndex)
	}
	(*slots)[index] = handler
}

func EnableDamageDoneByCaster(index int, maxIndex int, attackTable *AttackTable, handler DynamicDamageDoneByCaster) {
	setCasterSlot(&attackTable.DamageDoneByCasterExtraMultiplier, index, maxIndex, handler)
}

func DisableDamageDoneByCaster(index int, attackTable *AttackTable) {
	attackTable.DamageDoneByCasterExtraMultiplier[index] = nil
}

func EnableCritChanceByCaster(index int, maxIndex int, attackTable *AttackTable, handler DynamicCritChanceByCaster) {
	setCasterSlot(&attackTable.CritChanceByCasterBonus, index, maxIndex, handler)
}

func DisableCritChanceByCaster(index int, attackTable *AttackTable) {
	attackTable.CritChanceByCasterBonus[index] = nil
}

// The percentage points of crit chance the handlers add to the spell.
func (attackTable *AttackTable) critChanceByCaster(spell *Spell) float64 {
	bonus := 0.0
	for _, handler := range attackTable.CritChanceByCasterBonus {
		if handler != nil {
			bonus += handler(spell)
		}
	}
	return bonus
}

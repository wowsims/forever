package core

import (
	"math"
	"time"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type UnitType int
type SpellRegisteredHandler func(spell *Spell)
type OnMasteryStatChanged func(sim *Simulation, oldMasteryRating float64, newMasteryRating float64)
type OnSpeedChanged func(oldSpeed float64, newSpeed float64)
type OnTemporaryStatsChange func(sim *Simulation, buffAura *Aura, statsChangeWithoutDeps stats.Stats)

const (
	PlayerUnit UnitType = iota
	EnemyUnit
	PetUnit
)

type PowerBarType int

const (
	ManaBar PowerBarType = iota
	EnergyBar
	RageBar
	FocusBar
)

type DynamicDamageTakenModifier func(sim *Simulation, spell *Spell, result *SpellResult, isPeriodic bool)
type DynamicHealingTakenModifier func(sim *Simulation, spell *Spell, result *SpellResult)

type GetSpellDamageValue func(spell *Spell, target *Unit) float64
type GetAttackPowerValue func(spell *Spell, target *Unit) float64

// Unit is an abstraction of a Character/Boss/Pet/etc, containing functionality
// shared by all of them.
type Unit struct {
	Type UnitType

	// Index of this unit with its group.
	//  For Players, this is the 0-indexed raid index (0-24).
	//  For Enemies, this is its enemy index.
	//  For Pets, this is the same as the owner's index.
	Index int32

	// Unique index of this unit among all units in the environment.
	// This is used as the index for attack tables.
	UnitIndex int32

	// Unique label for logging.
	Label string

	Level int32 // Level of Unit, e.g. Bosses are 83.

	MobType        proto.MobType
	CreatureFamily proto.CreatureFamily

	// Amount of time it takes for the human agent to react to in-game events.
	// Used by certain APL values and actions.
	ReactionTime time.Duration

	// Amount of time following a post-GCD channel tick, to when the next action can be performed.
	ChannelClipDelay time.Duration

	// How far this unit is from its target(s). Measured in yards, this is used
	// for calculating spell travel time for certain spells.
	StartDistanceFromTarget float64
	DistanceFromTarget      float64
	Moving                  bool
	movementCallbacks       []MovementCallback
	moveAura                *Aura
	moveSpell               *Spell
	movementAction          *MovementAction

	// Environment in which this Unit exists. This will be nil until after the
	// construction phase.
	Env *Environment

	// Whether this unit is able to perform actions.
	enabled bool

	// A pet its owner set on the target before the pull: its rotation and swings run before 0.
	engagedBeforePull bool

	executePhase int32 // 35 or 20 for the execute phase a target is in, 0 above them and for players and pets.

	// Stats this Unit will have at the very start of each Sim iteration.
	// Includes all equipment / buffs / permanent effects but not temporary
	// effects from items / abilities.
	initialStats            stats.Stats
	initialStatsWithoutDeps stats.Stats

	initialPseudoStats stats.PseudoStats

	// Cast speed without any temporary effects.
	initialCastSpeed float64

	// Melee swing speed without any temporary effects.
	initialMeleeSwingSpeed float64

	// Ranged swing speed without any temporary effects.
	initialRangedSwingSpeed float64

	basePhysicalCritPercent float64 // The base physical crit of the class or pet, which no defender suppresses.
	critPerAgility          float64 // The physical crit per point of Agility, which no defender suppresses.

	// Provides aura tracking behavior.
	auraTracker

	// Current stats, including temporary effects but not dependencies.
	statsWithoutDeps stats.Stats

	// Current stats, including temporary effects and dependencies.
	stats stats.Stats

	// Provides stat dependency management behavior.
	stats.StatDependencyManager

	PseudoStats stats.PseudoStats

	// The form the unit is in, by the game files' numbering; the class writes it.
	ShapeshiftForm dbcenums.ShapeshiftForm
	// Leaves the current form, for a cast the form refuses but no form allows. Nil: none.
	AutoUnshift func(sim *Simulation)

	currentPowerBar PowerBarType
	healthBar
	manaBar
	rageBar
	energyBar
	focusBar

	// All spells that can be cast by this unit.
	Spellbook                 []*Spell
	spellRegistrationHandlers []SpellRegisteredHandler

	spellMods []*SpellMod
	equipMods []*SpellMod

	// The spell mods that raise the chance of the unit's proc triggers, which read them when they roll.
	procChanceMods []*SpellMod
	// A pet's owner, whose proc chance mods the pet's triggers count with the pet's own.
	procChanceOwner *Unit

	// Pets owned by this Unit.
	PetAgents []PetAgent

	DynamicStatsPets       []*Pet
	DynamicMeleeSpeedPets  []*Pet
	DynamicRangedSpeedPets []*Pet
	DynamicCastSpeedPets   []*Pet
	RegenInheritancePets   []*Pet

	// AutoAttacks is the manager for auto attack swings.
	// Must be enabled to use, with "EnableAutoAttacks()".
	AutoAttacks AutoAttacks
	Wand        *Wand // Set by EnableWand for a unit that can shoot a wand.

	Rotation *APLRotation

	// Statistics describing the results of the sim.
	Metrics UnitMetrics

	cdTimers []*Timer

	categoryTimers map[int32]*Timer

	dotRankGroups map[string]*dotRankGroup

	AttackTables                 []*AttackTable
	DynamicDamageTakenModifiers  []DynamicDamageTakenModifier
	DynamicHealingTakenModifiers []DynamicHealingTakenModifier
	Blockhandler                 func(sim *Simulation, spell *Spell, result *SpellResult)

	GCD *Timer

	// Separate from GCD timer to support spell queueing and off-GCD actions
	RotationTimer *Timer

	// Used for applying the effect of a hardcast spell when casting finishes.
	//  For channeled spells, only Expires is set.
	// No more than one cast may be active at any given time.
	Hardcast         Hardcast
	pickingAtCastEnd bool // The rotation is picking what follows the hardcast at its end, before it lands.
	// Aura that reduces dodge/parry/block to 0 while casting.
	HardcastAvoidanceAura *Aura

	// Auras granting immunity to a crowd control kind
	ccImmunityAuras map[string][]*Aura

	// Rotation-related PendingActions.
	rotationAction *PendingAction
	hardcastAction *PendingAction

	// Cached mana regen per second.
	manaRegenWhileCasting    float64
	manaRegenWhileNotCasting float64

	CastSpeed           float64
	meleeAttackSpeed    float64
	rangedAttackSpeed   float64
	meleeAndRangedHaste float64

	CurrentTarget   *Unit
	defaultTarget   *Unit
	SecondaryTarget *Unit // An NPC's second tank assignment; counts that tank as tanking for chance of death.

	// The currently-channeled DOT spell, otherwise nil.
	ChanneledDot       *Dot
	channelEndHandlers []func(sim *Simulation, dot *Dot)

	// Data about the most recently queued spell, otherwise nil.
	QueuedSpell *QueuedSpell

	// The cast the rotation or a major cooldown has in flight; procs never set it.
	rotationCast rotationCast

	// Used for reacting to mastery stat changes
	OnMasteryStatChanged []OnMasteryStatChanged

	// Used for reacting to cast speed changes
	OnCastSpeedChanged []OnSpeedChanged

	// Used for reacting to melee attack speed changes
	OnMeleeAttackSpeedChanged []OnSpeedChanged

	// Used for reacting to ranged attack speed changes
	OnRangedAttackSpeedChanged []OnSpeedChanged

	// Used for reacting to melee and ranged haste changes
	OnMeleeAndRangedHasteChanged []OnSpeedChanged

	// Used for reacting to transient stat changes (e.g. for caching snapshotting calculations)
	OnTemporaryStatsChanges []OnTemporaryStatsChange

	GetSpellDamageValue GetSpellDamageValue

	GetAttackPowerValue GetAttackPowerValue

	SpellsInFlight map[*Spell]int32
}

func (unit *Unit) getSpellDamageValueImpl(spell *Spell, _ *Unit) float64 {
	return spell.Unit.GetStat(stats.SpellDamage) + spell.BonusSpellDamage + spell.SpellSchoolBonusDamage()
}

func (unit *Unit) getAttackPowerValueImpl(spell *Spell, target *Unit) float64 {
	return unit.GetStat(stats.AttackPower) + target.PseudoStats.BonusAttackPower + spell.Unit.AttackTables[target.UnitIndex].MobTypeBonusStats[target.MobType][stats.AttackPower]
}

// Units can be disabled for several reasons:
//  1. Downtime for temporary pets (e.g. Water Elemental)
//  2. Enemy units in various phases (not yet implemented)
//  3. Dead units (not yet implemented)
func (unit *Unit) IsEnabled() bool {
	return unit.enabled
}

func (unit *Unit) IsActive() bool {
	return unit.IsEnabled() && unit.CurrentHealthPercent() > 0
}

func (unit *Unit) IsOpponent(other *Unit) bool {
	return (unit.Type == EnemyUnit) != (other.Type == EnemyUnit)
}

func (unit *Unit) GetAllOpponents() []*Unit {
	if unit.Type == EnemyUnit {
		return unit.Env.Raid.AllUnits
	} else {
		return unit.Env.Encounter.AllTargetUnits
	}
}

func (unit *Unit) LogLabel() string {
	return "[" + unit.Label + "]"
}

func (unit *Unit) Log(sim *Simulation, message string, vals ...interface{}) {
	sim.Log(unit.LogLabel()+" "+message, vals...)
}

func (unit *Unit) GetInitialStat(stat stats.Stat) float64 {
	return unit.initialStats[stat]
}
func (unit *Unit) GetStats() stats.Stats {
	return unit.stats
}
func (unit *Unit) GetStatsWithoutDeps() stats.Stats {
	return unit.statsWithoutDeps
}

// Given an array of Stat types, return the Stat whose value is largest for this
// Unit.
func (unit *Unit) GetHighestStatType(statTypeOptions []stats.Stat) stats.Stat {
	return unit.stats.GetHighestStatType(statTypeOptions)
}
func (unit *Unit) GetStat(stat stats.Stat) float64 {
	return unit.stats[stat]
}
func (unit *Unit) AddStats(stat stats.Stats) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, use AddStatsDynamic instead!")
	}
	unit.stats = unit.stats.Add(stat)
}
func (unit *Unit) AddStat(stat stats.Stat, amount float64) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, use AddStatDynamic instead!")
	}
	unit.stats[stat] += amount
}

func (unit *Unit) AddDynamicDamageTakenModifier(ddtm DynamicDamageTakenModifier) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add dynamic damage taken modifier!")
	}
	unit.DynamicDamageTakenModifiers = append(unit.DynamicDamageTakenModifiers, ddtm)
}

func (unit *Unit) AddDynamicHealingTakenModifier(dhtm DynamicHealingTakenModifier) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add dynamic healing taken modifier!")
	}
	unit.DynamicHealingTakenModifiers = append(unit.DynamicHealingTakenModifiers, dhtm)
}

// AddReducedCritTakenPercent adds to the base crit reduction and recalculates the total.
// Use this for dynamic changes during simulation (e.g., auras).
func (unit *Unit) AddReducedCritTakenPercent(amount float64) {
	unit.PseudoStats.BaseReducedCritTakenPercent += amount
	unit.updateReducedCritTakenPercent()
}

func (unit *Unit) AddOnMasteryStatChanged(omsc OnMasteryStatChanged) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add on mastery stat changed callback!")
	}
	unit.OnMasteryStatChanged = append(unit.OnMasteryStatChanged, omsc)
}

func (unit *Unit) AddOnCastSpeedChanged(ocsc OnSpeedChanged) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add on casting speed changed callback!")
	}
	unit.OnCastSpeedChanged = append(unit.OnCastSpeedChanged, ocsc)
}

func (unit *Unit) AddOnMeleeAttackSpeedChanged(ocsc OnSpeedChanged) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add on melee attack speed changed callback!")
	}
	unit.OnMeleeAttackSpeedChanged = append(unit.OnMeleeAttackSpeedChanged, ocsc)
}

func (unit *Unit) AddOnRangedAttackSpeedChanged(ocsc OnSpeedChanged) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add on ranged attack speed changed callback!")
	}
	unit.OnRangedAttackSpeedChanged = append(unit.OnRangedAttackSpeedChanged, ocsc)
}

func (unit *Unit) AddOnMeleeAndRangedHasteChanged(ocsc OnSpeedChanged) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add on melee and ranged haste changed callback!")
	}
	unit.OnMeleeAndRangedHasteChanged = append(unit.OnMeleeAndRangedHasteChanged, ocsc)
}

func (unit *Unit) AddOnTemporaryStatsChange(otsc OnTemporaryStatsChange) {
	if unit.Env != nil && unit.Env.IsFinalized() {
		panic("Already finalized, cannot add on temporary stats change callback!")
	}
	unit.OnTemporaryStatsChanges = append(unit.OnTemporaryStatsChanges, otsc)
}

func (unit *Unit) AddStatsDynamic(sim *Simulation, bonus stats.Stats) {
	unit.addStatsDynamic(sim, &bonus)
}

func (unit *Unit) addStatsDynamic(sim *Simulation, bonus *stats.Stats) {
	if unit.Env == nil {
		panic("Environment not constructed.")
	} else if !unit.Env.IsFinalized() && !unit.Env.MeasuringStats {
		panic("Not finalized, use AddStats instead!")
	}

	unit.statsWithoutDeps.AddInplace(bonus)

	if !unit.Env.MeasuringStats || unit.Env.State == Finalized {
		// Recompute the full total rather than applying dependencies to the
		// delta: floored stats are only path-independent as floor(newTotal) -
		// floor(oldTotal), while flooring per-delta would drift on every
		// aura gain/expire cycle.
		old := &unit.stats
		*bonus = unit.statsWithoutDeps
		unit.StatDependencyManager.ApplyStatDependenciesInPlace(bonus)
		bonus.FloorGameStatsInPlace()
		for k := range bonus {
			bonus[k], old[k] = bonus[k]-old[k], bonus[k]
		}
	} else {
		unit.stats.AddInplace(bonus)
	}

	if sim.Log != nil {
		unit.Log(sim, "Dynamic stat change: %s", bonus.FlatString())
	}

	unit.processDynamicBonus(sim, bonus)
}

func (unit *Unit) AddStatDynamic(sim *Simulation, stat stats.Stat, amount float64) {
	var bonus stats.Stats
	bonus[stat] = amount
	unit.addStatsDynamic(sim, &bonus)
}

func (unit *Unit) processDynamicBonus(sim *Simulation, bonus *stats.Stats) {
	if bonus[stats.MP5] != 0 || bonus[stats.Intellect] != 0 || bonus[stats.Spirit] != 0 {
		unit.UpdateManaRegenRates(sim)
	}
	if bonus[stats.Mana] != 0 && unit.HasManaBar() {
		if unit.CurrentMana() > unit.MaxMana() {
			unit.currentMana = unit.MaxMana()
		}
	}
	if bonus[stats.HasteRating] != 0 {
		unit.updateAttackSpeed()
		unit.updateMeleeAndRangedHaste()
		unit.AutoAttacks.UpdateSwingTimers(sim)
		unit.updateCastSpeed()
	}
	if bonus[stats.ReducedCritTakenPercent] != 0 {
		unit.updateReducedCritTakenPercent()
	}

	// Higher performance than calling TriggerDelayedPetInheritance()
	for _, pet := range unit.DynamicStatsPets {
		if !pet.enabled {
			continue
		}

		if pet.statInheritanceAction.consumed || (pet.statInheritanceAction.NextActionAt == 0) {
			numHeartbeats := (sim.CurrentTime - unit.Env.heartbeatOffset) / PetUpdateInterval
			pet.statInheritanceAction.NextActionAt = PetUpdateInterval*(numHeartbeats+1) + unit.Env.heartbeatOffset
			sim.AddPendingAction(pet.statInheritanceAction)
		}
	}
}

func (unit *Unit) EnableDynamicStatDep(sim *Simulation, dep *stats.StatDependency) {
	if unit.StatDependencyManager.EnableDynamicStatDep(dep) {
		oldStats := unit.stats
		unit.stats = unit.ApplyStatDependencies(unit.statsWithoutDeps).FloorGameStats()
		d := unit.stats.Subtract(oldStats)
		unit.processDynamicBonus(sim, &d)

		if sim.Log != nil {
			unit.Log(sim, "Dynamic dep enabled (%s): %s", dep.String(), unit.stats.Subtract(oldStats).FlatString())
		}
	}
}
func (unit *Unit) DisableDynamicStatDep(sim *Simulation, dep *stats.StatDependency) {
	if unit.StatDependencyManager.DisableDynamicStatDep(dep) {
		oldStats := unit.stats
		unit.stats = unit.ApplyStatDependencies(unit.statsWithoutDeps).FloorGameStats()
		d := unit.stats.Subtract(oldStats)
		unit.processDynamicBonus(sim, &d)

		if sim.Log != nil {
			unit.Log(sim, "Dynamic dep disabled (%s): %s", dep.String(), unit.stats.Subtract(oldStats).FlatString())
		}
	}
}

func (unit *Unit) UpdateDynamicStatDep(sim *Simulation, dep *stats.StatDependency, newAmount float64) {
	dep.UpdateValue(newAmount)

	if unit.Env.IsFinalized() {
		oldStats := unit.stats
		unit.stats = unit.ApplyStatDependencies(unit.statsWithoutDeps).FloorGameStats()
		statsChange := unit.stats.Subtract(oldStats)
		unit.processDynamicBonus(sim, &statsChange)

		if sim.Log != nil {
			unit.Log(sim, "Dynamic dep updated (%s): %s", dep.String(), statsChange.FlatString())
		}
	}
}

func (unit *Unit) EnableBuildPhaseStatDep(sim *Simulation, dep *stats.StatDependency) {
	if unit.Env.MeasuringStats && unit.Env.State != Finalized {
		unit.StatDependencyManager.EnableDynamicStatDep(dep)
	} else {
		unit.EnableDynamicStatDep(sim, dep)
	}
}

func (unit *Unit) DisableBuildPhaseStatDep(sim *Simulation, dep *stats.StatDependency) {
	if unit.Env.MeasuringStats && unit.Env.State != Finalized {
		unit.StatDependencyManager.DisableDynamicStatDep(dep)
	} else {
		unit.DisableDynamicStatDep(sim, dep)
	}
}

// Returns whether the indicates stat is currently modified by a temporary bonus.
func (unit *Unit) HasTemporaryBonusForStat(stat stats.Stat) bool {
	return unit.initialStats[stat] != unit.stats[stat]
}

// Returns if spell casting has any temporary increases active.
func (unit *Unit) HasTemporarySpellCastSpeedIncrease() bool {
	return unit.CastSpeed != unit.initialCastSpeed
}

// Returns if melee swings have any temporary increases active.
func (unit *Unit) HasTemporaryMeleeSwingSpeedIncrease() bool {
	return unit.TotalMeleeHasteMultiplier() != unit.initialMeleeSwingSpeed
}

// Returns if ranged swings have any temporary increases active.
func (unit *Unit) HasTemporaryRangedSwingSpeedIncrease() bool {
	return unit.TotalRangedHasteMultiplier() != unit.initialRangedSwingSpeed
}

func (unit *Unit) InitialCastSpeed() float64 {
	return unit.initialCastSpeed
}

func (unit *Unit) IsChanneling() bool {
	return unit.ChanneledDot != nil
}

// Runs the handler with the ended channel each time one of the unit's channels ends, run out or cut short, after its last tick
// and after the rotation's wake-up is set, which the handler may move.
func (unit *Unit) OnChannelEnd(handler func(sim *Simulation, dot *Dot)) {
	unit.channelEndHandlers = append(unit.channelEndHandlers, handler)
}

func (unit *Unit) SpellGCD() time.Duration {
	return max(GCDMin, unit.ApplyCastSpeed(GCDDefault))
}

func (unit *Unit) hasteRatingMultiplier() float64 {
	return 1 + unit.stats[stats.HasteRating]/(HasteRatingPerHastePercent*100)
}

func (unit *Unit) TotalSpellHasteMultiplier() float64 {
	return unit.PseudoStats.CastSpeedMultiplier * unit.hasteRatingMultiplier()
}

func (unit *Unit) updateCastSpeed() {
	oldCastSpeed := unit.CastSpeed
	unit.CastSpeed = 1 / unit.TotalSpellHasteMultiplier()
	newCastSpeed := unit.CastSpeed

	for i := range unit.OnCastSpeedChanged {
		unit.OnCastSpeedChanged[i](oldCastSpeed, newCastSpeed)
	}
}
func (unit *Unit) MultiplyCastSpeed(sim *Simulation, amount float64) {
	unit.PseudoStats.CastSpeedMultiplier *= amount

	unit.Env.TriggerDelayedPetInheritance(sim, unit.DynamicCastSpeedPets, func(sim *Simulation, pet *Pet) {
		pet.dynamicCastSpeedInheritance(sim, amount)
	})

	unit.updateCastSpeed()
}

func (unit *Unit) ApplyCastSpeed(dur time.Duration) time.Duration {
	return time.Duration(float64(dur) * unit.CastSpeed)
}
func (unit *Unit) ApplyCastSpeedForSpell(dur time.Duration, spell *Spell) time.Duration {
	return max(0, time.Duration(float64(dur)*unit.CastSpeed*max(0, spell.CastTimeMultiplier)))
}

// ApplyRangedSpeed applies ranged haste to a duration, for ranged abilities that should be affected by ranged haste
func (unit *Unit) ApplyRangedSpeed(dur time.Duration) time.Duration {
	return time.Duration(float64(dur) / unit.TotalRangedHasteMultiplier())
}

func (unit *Unit) ApplyRealHaste(dur time.Duration) time.Duration {
	return time.Duration(float64(dur) / unit.TotalRealHasteMultiplier())
}

func (unit *Unit) ApplyRealRangedHaste(dur time.Duration) time.Duration {
	return time.Duration(float64(dur) / unit.TotalRealRangedHasteMultiplier())
}
func (unit *Unit) TotalMeleeHasteMultiplier() float64 {
	return unit.PseudoStats.AttackSpeedMultiplier * unit.PseudoStats.MeleeSpeedMultiplier * unit.hasteRatingMultiplier()
}

// Returns the melee haste multiplier only including equip haste and real haste modifiers like lust
// Same value for ranged and melee
func (unit *Unit) TotalRealHasteMultiplier() float64 {
	return unit.PseudoStats.AttackSpeedMultiplier * unit.hasteRatingMultiplier()
}

func (unit *Unit) TotalRealRangedHasteMultiplier() float64 {
	return unit.PseudoStats.AttackSpeedMultiplier * unit.PseudoStats.RangedHasteMultiplier * unit.hasteRatingMultiplier()
}

func (unit *Unit) Armor() float64 {
	return unit.PseudoStats.ArmorMultiplier * unit.stats[stats.Armor]
}

func (unit *Unit) BlockDamageReduction() float64 {
	return max(0, unit.stats[stats.Strength]*unit.PseudoStats.BlockValuePerStrength-1) + unit.stats[stats.BlockValue]*unit.PseudoStats.BlockValueMultiplier
}

func (unit *Unit) TotalRangedHasteMultiplier() float64 {
	return unit.PseudoStats.AttackSpeedMultiplier * unit.PseudoStats.RangedSpeedMultiplier * unit.hasteRatingMultiplier()
}

func (unit *Unit) updateMeleeAttackSpeed() {
	oldMeleeAttackSpeed := unit.meleeAttackSpeed
	unit.meleeAttackSpeed = unit.TotalMeleeHasteMultiplier()

	for i := range unit.OnMeleeAttackSpeedChanged {
		unit.OnMeleeAttackSpeedChanged[i](oldMeleeAttackSpeed, unit.meleeAttackSpeed)
	}
}

// MultiplyMeleeSpeed will alter the attack speed multiplier and change swing speed of all autoattack swings in progress.
func (unit *Unit) MultiplyMeleeSpeed(sim *Simulation, amount float64) {
	unit.PseudoStats.MeleeSpeedMultiplier *= amount
	unit.updateMeleeAttackSpeed()

	if len(unit.DynamicMeleeSpeedPets) > 0 {
		unit.Env.TriggerDelayedPetInheritance(sim, unit.DynamicMeleeSpeedPets, func(sim *Simulation, pet *Pet) {
			pet.dynamicMeleeSpeedInheritance(sim, amount)
		})
	}

	unit.AutoAttacks.UpdateSwingTimers(sim)
}

// For an attack speed aura that swings the weapon as a faster, lighter one: the swings speed up as
// melee speed does, and WeaponSpeedMultiplier carries the factor for the procs that read the weapon's
// speed.
func (unit *Unit) MultiplyWeaponSpeed(sim *Simulation, amount float64) {
	unit.PseudoStats.WeaponSpeedMultiplier *= amount
	unit.MultiplyMeleeSpeed(sim, amount)
}

func (unit *Unit) updateRangedAttackSpeed() {
	oldRangedAttackSpeed := unit.rangedAttackSpeed
	unit.rangedAttackSpeed = unit.TotalRangedHasteMultiplier()

	for i := range unit.OnRangedAttackSpeedChanged {
		unit.OnRangedAttackSpeedChanged[i](oldRangedAttackSpeed, unit.rangedAttackSpeed)
	}
}

// Used for "Ranged attack speed" effects like Steady Focus and Serpent's Swiftness
func (unit *Unit) MultiplyRangedSpeed(sim *Simulation, amount float64) {
	unit.PseudoStats.RangedSpeedMultiplier *= amount
	unit.updateRangedAttackSpeed()

	// The pets that take the owner's ranged speed take each change at once, not at the next pet
	// heartbeat.
	for _, pet := range unit.DynamicRangedSpeedPets {
		pet.dynamicRangedSpeedInheritance(sim, amount)
	}

	unit.AutoAttacks.UpdateSwingTimers(sim)
}

// Used for "Ranged haste" effects that modify both attack speed and focus regen, like Rapid Fire and Focus Fire
func (unit *Unit) MultiplyRangedHaste(sim *Simulation, amount float64) {
	unit.PseudoStats.RangedHasteMultiplier *= amount
	unit.MultiplyRangedSpeed(sim, amount)
	unit.MultiplyResourceRegenSpeed(sim, amount)
}

func (unit *Unit) updateAttackSpeed() {
	oldMeleeAttackSpeed := unit.meleeAttackSpeed
	oldRangedAttackSpeed := unit.rangedAttackSpeed

	unit.meleeAttackSpeed = unit.TotalMeleeHasteMultiplier()
	unit.rangedAttackSpeed = unit.TotalRangedHasteMultiplier()

	for i := range unit.OnMeleeAttackSpeedChanged {
		unit.OnMeleeAttackSpeedChanged[i](oldMeleeAttackSpeed, unit.meleeAttackSpeed)
	}

	for i := range unit.OnRangedAttackSpeedChanged {
		unit.OnRangedAttackSpeedChanged[i](oldRangedAttackSpeed, unit.rangedAttackSpeed)
	}
}

func (unit *Unit) updateMeleeAndRangedHaste() {
	oldMeleeAndRangedHaste := unit.meleeAndRangedHaste

	unit.meleeAndRangedHaste = unit.TotalRealHasteMultiplier()

	for i := range unit.OnMeleeAndRangedHasteChanged {
		unit.OnMeleeAndRangedHasteChanged[i](oldMeleeAndRangedHaste, unit.meleeAndRangedHaste)
	}
}

// Helper for when true haste effects are multiplied.
// Seems to also always impact the regen rate
func (unit *Unit) MultiplyAttackSpeed(sim *Simulation, amount float64) {
	unit.PseudoStats.AttackSpeedMultiplier *= amount
	// unit.MultiplyResourceRegenSpeed(sim, amount)
	unit.updateAttackSpeed()
	unit.updateMeleeAndRangedHaste()

	if len(unit.DynamicMeleeSpeedPets) > 0 {
		unit.Env.TriggerDelayedPetInheritance(sim, unit.DynamicMeleeSpeedPets, func(sim *Simulation, pet *Pet) {
			pet.dynamicMeleeSpeedInheritance(sim, amount)
		})
	}

	unit.AutoAttacks.UpdateSwingTimers(sim)
}

// Helper for multiplying resource generation speed
func (unit *Unit) MultiplyResourceRegenSpeed(sim *Simulation, amount float64) {
	if unit.HasEnergyBar() {
		unit.MultiplyEnergyRegenSpeed(sim, amount)
	}

	unit.Env.TriggerDelayedPetInheritance(sim, unit.RegenInheritancePets, func(sim *Simulation, pet *Pet) {
		pet.MultiplyResourceRegenSpeed(sim, amount)
	})
}

func (unit *Unit) SetCurrentPowerBar(bar PowerBarType) {
	unit.currentPowerBar = bar
}

func (unit *Unit) GetCurrentPowerBar() PowerBarType {
	return unit.currentPowerBar
}

// A rating that converts into a percent stat at RatingPerPercent rating per percent point. Where Step
// is non-zero the rating counts in whole steps of that size only.
type RatingConversion struct {
	Rating           stats.Stat
	Percent          stats.Stat
	RatingPerPercent float64
	Step             float64
}

// The melee, ranged and spell percents one rating raises together, as the PseudoStats the character sheet
// shows, at RatingPerPercent rating per percent point. Where HasPercentStats is set the rating also
// converts into the back-end PhysicalPercent and SpellPercent; haste has no back-end percent Stat.
type RatingChannels struct {
	Rating                        stats.Stat
	Melee, Ranged, Spell          proto.PseudoStat
	PhysicalPercent, SpellPercent stats.Stat
	HasPercentStats               bool
	RatingPerPercent              float64
}

func (channels RatingChannels) Channels() []proto.PseudoStat {
	return []proto.PseudoStat{channels.Melee, channels.Ranged, channels.Spell}
}

// The channels of the hit, crit and haste ratings. RATING_CHANNELS in ui/sim/proto/stats.ts is the UI's
// copy of the same channels; the two are kept in sync by hand.
var RatingChannelTable = []RatingChannels{
	{
		Rating:           stats.HitRating,
		Melee:            proto.PseudoStat_PseudoStatMeleeHitPercent,
		Ranged:           proto.PseudoStat_PseudoStatRangedHitPercent,
		Spell:            proto.PseudoStat_PseudoStatSpellHitPercent,
		PhysicalPercent:  stats.PhysicalHitPercent,
		SpellPercent:     stats.SpellHitPercent,
		HasPercentStats:  true,
		RatingPerPercent: HitRatingPerHitPercent,
	},
	{
		Rating:           stats.CritRating,
		Melee:            proto.PseudoStat_PseudoStatMeleeCritPercent,
		Ranged:           proto.PseudoStat_PseudoStatRangedCritPercent,
		Spell:            proto.PseudoStat_PseudoStatSpellCritPercent,
		PhysicalPercent:  stats.PhysicalCritPercent,
		SpellPercent:     stats.SpellCritPercent,
		HasPercentStats:  true,
		RatingPerPercent: CritRatingPerCritPercent,
	},
	{
		Rating:           stats.HasteRating,
		Melee:            proto.PseudoStat_PseudoStatMeleeHastePercent,
		Ranged:           proto.PseudoStat_PseudoStatRangedHastePercent,
		Spell:            proto.PseudoStat_PseudoStatSpellHastePercent,
		RatingPerPercent: HasteRatingPerHastePercent,
	},
}

// The channels of a rating, and whether RatingChannelTable lists it.
func RatingChannelsOf(rating stats.Stat) (RatingChannels, bool) {
	for _, channels := range RatingChannelTable {
		if channels.Rating == rating {
			return channels, true
		}
	}
	return RatingChannels{}, false
}

// The pseudo stat each magic school's hit percent travels as. None and Physical have none and hold
// the zero value, PseudoStatMainHandDps, so read from SchoolIndexArcane on.
var SchoolHitPseudoStats = [stats.SchoolLen]proto.PseudoStat{
	stats.SchoolIndexArcane: proto.PseudoStat_PseudoStatSchoolHitPercentArcane,
	stats.SchoolIndexFire:   proto.PseudoStat_PseudoStatSchoolHitPercentFire,
	stats.SchoolIndexFrost:  proto.PseudoStat_PseudoStatSchoolHitPercentFrost,
	stats.SchoolIndexHoly:   proto.PseudoStat_PseudoStatSchoolHitPercentHoly,
	stats.SchoolIndexNature: proto.PseudoStat_PseudoStatSchoolHitPercentNature,
	stats.SchoolIndexShadow: proto.PseudoStat_PseudoStatSchoolHitPercentShadow,
}

// The melee and spell percent of each channel table rating with back-end percent Stats. The ranged Stat
// holds only what exceeds the melee percent, so no rating converts into it.
func channelConversions() []RatingConversion {
	var conversions []RatingConversion
	for _, channels := range RatingChannelTable {
		if channels.HasPercentStats {
			conversions = append(conversions,
				RatingConversion{channels.Rating, channels.PhysicalPercent, channels.RatingPerPercent, 0},
				RatingConversion{channels.Rating, channels.SpellPercent, channels.RatingPerPercent, 0})
		}
	}
	return conversions
}

// The rating conversions every unit has, in the order the reforge optimizer tries a rating's percents.
var RatingConversions = append(channelConversions(), []RatingConversion{
	{stats.ExpertiseRating, stats.ExpertisePercent, ExpertiseRatingPerExpertisePercent, 0},
	{stats.DodgeRating, stats.DodgePercent, DodgeRatingPerDodgePercent, 0},
	{stats.ParryRating, stats.ParryPercent, ParryRatingPerParryPercent, 0},
	{stats.BlockRating, stats.BlockPercent, BlockRatingPerBlockPercent, 0},
	{stats.DefenseRating, stats.ReducedCritTakenPercent, DefenseRatingPerAvoidancePercent, DefenseRatingPerDefenseLevel},
	{stats.DefenseRating, stats.DodgePercent, DefenseRatingPerAvoidancePercent, 0},
	{stats.DefenseRating, stats.ParryPercent, DefenseRatingPerAvoidancePercent, 0},
	{stats.DefenseRating, stats.BlockPercent, DefenseRatingPerAvoidancePercent, 0},
}...)

func AddRatingConversions(sdm *stats.StatDependencyManager) {
	for _, conversion := range RatingConversions {
		if conversion.Step != 0 {
			sdm.AddFlooredStatDependency(conversion.Rating, conversion.Percent, conversion.Step, 1/conversion.RatingPerPercent)
		} else {
			sdm.AddStatDependency(conversion.Rating, conversion.Percent, 1/conversion.RatingPerPercent)
		}
	}
}

// Stat dependencies that apply both to players/pets (represented as Character
// structs) and to NPCs (represented as Target structs).
func (unit *Unit) addUniversalStatDependencies() {
	AddRatingConversions(&unit.StatDependencyManager)
	// A_MOD_ARMOR_NO_MODS armor, which no percent armor modifier scales.
	unit.AddUnmultipliedStatDependency(stats.ArmorNoMods, stats.Armor)
}

func (unit *Unit) finalize() {
	if unit.Env.IsFinalized() {
		panic("Unit already finalized!")
	}

	// Make sure we don't accidentally set initial stats instead of stats.
	if !unit.initialStats.Equals(stats.Stats{}) {
		panic("Initial stats may not be set before finalized: " + unit.initialStats.String())
	}

	if unit.ReactionTime == 0 {
		panic("Unset unit.ReactionTime")
	}

	unit.defaultTarget = unit.CurrentTarget
	unit.applyParryHaste()
	unit.updateCastSpeed()
	unit.updateAttackSpeed()
	unit.updateMeleeAndRangedHaste()
	unit.initMovement()

	// All stats added up to this point are part of the 'initial' stats.
	unit.initialStatsWithoutDeps = unit.stats
	unit.initialCastSpeed = unit.CastSpeed
	unit.initialMeleeSwingSpeed = unit.TotalMeleeHasteMultiplier()
	unit.initialRangedSwingSpeed = unit.TotalRangedHasteMultiplier()

	unit.StatDependencyManager.FinalizeStatDeps()
	unit.initialStats = unit.ApplyStatDependencies(unit.initialStatsWithoutDeps).FloorGameStats()
	unit.statsWithoutDeps = unit.initialStatsWithoutDeps
	unit.stats = unit.initialStats

	unit.updateReducedCritTakenPercent()
	unit.initialPseudoStats = unit.PseudoStats

	unit.AutoAttacks.finalize()

	if unit.GetSpellDamageValue == nil {
		unit.GetSpellDamageValue = unit.getSpellDamageValueImpl
	}

	if unit.GetAttackPowerValue == nil {
		unit.GetAttackPowerValue = unit.getAttackPowerValueImpl
	}

	for _, spell := range unit.Spellbook {
		spell.finalize()
	}
}

func (unit *Unit) reset(sim *Simulation, _ Agent) {
	if unit.Type != EnemyUnit {
		unit.enabled = true
	}

	unit.resetCDs(sim)
	unit.Hardcast.Expires = startingCDTime
	unit.pickingAtCastEnd = false
	unit.ChanneledDot = nil
	unit.QueuedSpell = nil
	unit.endRotationCast()
	unit.DistanceFromTarget = unit.StartDistanceFromTarget
	unit.engagedBeforePull = false
	unit.Metrics.reset()
	unit.ResetStatDeps()
	unit.statsWithoutDeps = unit.initialStatsWithoutDeps
	unit.stats = unit.initialStats
	unit.PseudoStats = unit.initialPseudoStats
	unit.auraTracker.reset(sim)

	for _, spell := range unit.Spellbook {
		spell.reset(sim)
	}

	unit.manaBar.reset()
	unit.focusBar.reset(sim)
	unit.healthBar.reset(sim)
	unit.UpdateManaRegenRates(sim)

	unit.energyBar.reset(sim)
	unit.rageBar.reset(sim)

	unit.AutoAttacks.reset(sim)

	if unit.Rotation != nil {
		unit.Rotation.reset(sim)
	}

	unit.DynamicStatsPets = unit.DynamicStatsPets[:0]
	unit.DynamicMeleeSpeedPets = unit.DynamicMeleeSpeedPets[:0]
	unit.DynamicRangedSpeedPets = unit.DynamicRangedSpeedPets[:0]
	clear(unit.SpellsInFlight)

	if unit.Type != PetUnit {
		sim.addTracker(&unit.auraTracker)
	}
}

func (unit *Unit) onEncounterStart(sim *Simulation) {
	if agent := unit.Env.GetAgentFromUnit(unit); agent != nil {
		agent.OnEncounterStart(sim)
	}

	// Reduce Haste fakepoints arising from overly precise swing timings relative to the GCD timer.
	if unit.AutoAttacks.AutoSwingMelee && (unit.Type != EnemyUnit) && unit.AutoAttacks.RandomMeleeOffset {
		unit.AutoAttacks.RandomizeMeleeTiming(sim)
	}

	unit.OnEncounterStart(sim)
}

func (unit *Unit) startPull(sim *Simulation) {
	unit.AutoAttacks.startPull(sim)

	if unit.Type == PlayerUnit {
		unit.SetGCDTimer(sim, max(0, unit.GCD.ReadyAt()))
	}
}

func (unit *Unit) doneIteration(sim *Simulation) {
	unit.Hardcast = Hardcast{}

	unit.manaBar.doneIteration(sim)

	unit.auraTracker.doneIteration(sim)
	for _, spell := range unit.Spellbook {
		spell.doneIteration()
	}
}

func (unit *Unit) GetSpellsMatchingSchool(school SpellSchool) []*Spell {
	var spells []*Spell
	for _, spell := range unit.Spellbook {
		if spell.SpellSchool.Matches(school) {
			spells = append(spells, spell)
		}
	}
	return spells
}

func (unit *Unit) SpellInFlight(spell *Spell) bool {
	return unit.SpellsInFlight[spell] > 0
}

func (unit *Unit) SpellInFlightByID(spellID int32) bool {
	spell := unit.GetSpell(ActionID{SpellID: spellID})
	return unit.SpellInFlight(spell)
}

func (unit *Unit) GetUnit(ref *proto.UnitReference) *Unit {
	return unit.Env.GetUnit(ref, unit)
}

func (unit *Unit) GetMetadata() *proto.UnitMetadata {
	return unit.metadata(unit)
}

// casters are the units whose spells can cast this unit's auras, which decides an aura's category.
func (unit *Unit) metadata(casters ...*Unit) *proto.UnitMetadata {
	metadata := &proto.UnitMetadata{
		Name: unit.Label,
	}

	// The rank rides on the id so the UI can group a spell's ranks; the tooltips the UI fetches do not
	// always state one. An aura takes the rank of the spell sharing its id.
	ranks := map[ActionID]int32{}
	rankedID := func(actionID ActionID) *proto.ActionID {
		id := actionID.ToProto()
		id.Rank = ranks[actionID]
		return id
	}
	for _, spell := range unit.Spellbook {
		if spell.Rank > 0 {
			ranks[spell.ActionID] = spell.Rank
		}
	}

	metadata.Spells = MapSlice(unit.Spellbook, func(spell *Spell) *proto.SpellStats {
		return &proto.SpellStats{
			Id: rankedID(spell.ActionID),

			IsCastable:         spell.Flags.Matches(SpellFlagAPL),
			IsChanneled:        spell.Flags.Matches(SpellFlagChanneled),
			IsMajorCooldown:    spell.Flags.Matches(SpellFlagMCD),
			HasDot:             spell.dots != nil || spell.aoeDot != nil || (spell.RelatedDotSpell != nil && (spell.RelatedDotSpell.dots != nil || spell.RelatedDotSpell.aoeDot != nil)),
			HasShield:          spell.shields != nil || spell.selfShield != nil,
			PrepullOnly:        spell.Flags.Matches(SpellFlagPrepullOnly),
			EncounterOnly:      spell.Flags.Matches(SpellFlagEncounterOnly),
			HasCastTime:        spell.DefaultCast.CastTime > 0,
			IsFriendly:         spell.Flags.Matches(SpellFlagHelpful),
			IsPotion:           spell.Flags.Matches(SpellFlagPotion),
			IsConjured:         spell.Flags.Matches(SpellFlagConjured),
			HasExpectedTick:    spell.expectedTickDamageInternal != nil,
			HasExpectedInitial: spell.expectedInitialDamageInternal != nil,
			HasMissileSpeed:    spell.MissileSpeed > 0.0 || spell.MissileMinDuration > 0,
			HasRanks:           spell.Rank > 0,
			Category:           spell.category(),
			MenuGroup:          spell.MenuGroup,
		}
	})

	aplAuras := FilterSlice(unit.auras, func(aura *Aura) bool {
		return !aura.ActionID.IsEmptyAction()
	})
	metadata.Auras = MapSlice(aplAuras, func(aura *Aura) *proto.AuraStats {
		return &proto.AuraStats{
			Id:                 rankedID(aura.ActionID),
			MaxStacks:          aura.MaxStacks,
			HasIcd:             aura.Icd != nil,
			HasExclusiveEffect: len(aura.ExclusiveEffects) > 0,
			Category:           auraCategory(aura, casters),
			MenuGroup:          aura.MenuGroup,
		}
	})

	return metadata
}

func (unit *Unit) ExecuteCustomRotation(sim *Simulation) {
	panic("Unimplemented ExecuteCustomRotation")
}

func (unit *Unit) GetDodgeFromRating() float64 {
	return unit.stats[stats.DodgePercent] / 100
}
func (unit *Unit) GetParryFromRating() float64 {
	return unit.stats[stats.ParryPercent] / 100
}
func (unit *Unit) GetBlockFromRating() float64 {
	return unit.stats[stats.BlockPercent] / 100
}

func (unit *Unit) GetTotalDodgeChanceAsDefender(spell *Spell, atkTable *AttackTable) float64 {
	if spell.Flags.Matches(SpellFlagNoActiveDefense) {
		return 0
	}
	chance := unit.PseudoStats.BaseDodgeChance +
		atkTable.Hand(spell).BaseDodgeChance +
		unit.GetDodgeFromRating() -
		spell.DodgeParrySuppression() -
		spell.Unit.PseudoStats.DodgeReduction
	return math.Max(chance, 0.0)
}

func (unit *Unit) GetTotalParryChanceAsDefender(spell *Spell, atkTable *AttackTable) float64 {
	if spell.Flags.Matches(SpellFlagNoActiveDefense) {
		return 0
	}
	chance := unit.PseudoStats.BaseParryChance +
		atkTable.Hand(spell).BaseParryChance +
		unit.GetParryFromRating() -
		spell.DodgeParrySuppression()
	return math.Max(chance, 0.0)
}

func (unit *Unit) GetTotalChanceToBeMissedAsDefender(spell *Spell, atkTable *AttackTable) float64 {
	// ReducedPhysicalHitTakenChance is stored in percent points, BaseMissChance as a fraction.
	defense := math.Floor(unit.stats[stats.DefenseRating] / DefenseRatingPerDefenseLevel)
	chance := atkTable.Hand(spell).BaseMissChance +
		unit.PseudoStats.ReducedPhysicalHitTakenChance/100 +
		defense*MissDodgeParryBlockCritChancePerDefense/100
	return math.Max(chance, 0.0)
}

func (unit *Unit) GetTotalBlockChanceAsDefender(spell *Spell, atkTable *AttackTable) float64 {
	if spell.Flags.Matches(SpellFlagNoActiveDefense) {
		return 0
	}
	chance := unit.PseudoStats.BaseBlockChance +
		atkTable.Hand(spell).BaseBlockChance +
		unit.GetBlockFromRating()
	return math.Max(chance, 0.0)
}

func (unit *Unit) updateReducedCritTakenPercent() {
	unit.PseudoStats.ReducedCritTakenPercent = unit.PseudoStats.BaseReducedCritTakenPercent +
		unit.stats[stats.ReducedCritTakenPercent]/100
}

func (unit *Unit) GetTotalAvoidanceChance(spell *Spell, atkTable *AttackTable) float64 {
	miss := unit.GetTotalChanceToBeMissedAsDefender(spell, atkTable)
	dodge := unit.GetTotalDodgeChanceAsDefender(spell, atkTable)
	parry := unit.GetTotalParryChanceAsDefender(spell, atkTable)
	block := unit.GetTotalBlockChanceAsDefender(spell, atkTable)
	return miss + dodge + parry + block
}

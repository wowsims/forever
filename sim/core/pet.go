package core

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Extension of Agent interface, for Pets.
type PetAgent interface {
	Agent

	// The Pet controlled by this PetAgent.
	GetPet() *Pet
}

type OnPetEnable func(sim *Simulation)
type OnPetDisable func(sim *Simulation)

type PetStatInheritance func(ownerStats stats.Stats) stats.Stats
type PetSpeedInheritance func(sim *Simulation, ownerSpeedMultiplier float64)

type PetConfig struct {
	Name                             string
	Owner                            *Character
	BaseStats                        stats.Stats
	StatInheritance                  PetStatInheritance
	EnabledOnStart                   bool
	IsGuardian                       bool
	IsDynamic                        bool
	HasDynamicMeleeSpeedInheritance  bool
	HasDynamicRangedSpeedInheritance bool // The owner's ranged speed, as the pet's melee speed.
	HasDynamicCastSpeedInheritance   bool
	HasResourceRegenInheritance      bool
	StartsAtOwnerDistance            bool
}

// Pet is an extension of Character, for any entity created by a player that can
// take actions on its own.
type Pet struct {
	Character

	Owner *Character

	isGuardian     bool
	enabledOnStart bool

	OnPetEnable  OnPetEnable
	OnPetDisable OnPetDisable

	enableEffects  []OnPetEnable
	disableEffects []OnPetDisable

	// How the owner's stats map to the pet's, what the pet holds from them now, and for a dynamic
	// pet the refresh queued for the next pet heartbeat.
	statInheritance       PetStatInheritance
	inheritedStats        stats.Stats
	statInheritanceAction *PendingAction

	// A pet inherits its owner's melee, ranged and cast speed rather than
	// having the haste aura applied to it.
	dynamicMeleeSpeedInheritance   PetSpeedInheritance
	inheritedMeleeSpeedMultiplier  float64
	dynamicRangedSpeedInheritance  PetSpeedInheritance
	inheritedRangedSpeedMultiplier float64
	dynamicCastSpeedInheritance    PetSpeedInheritance
	inheritedCastSpeedMultiplier   float64

	// If true the pet will automatically inherit the owner's stats
	isDynamic bool
	// If true the pet will automatically inherit the owner's melee speed
	hasDynamicMeleeSpeedInheritance bool
	// If true the owner's ranged speed becomes the pet's melee speed
	hasDynamicRangedSpeedInheritance bool
	// If true the pet will automatically inherit the owner's cast speed
	hasDynamicCastSpeedInheritance bool
	// If true the pet will automatically inherit the owner's regen speed multiplier
	hasResourceRegenInheritance bool

	isReset bool

	// Some pets expire after a certain duration. This is the pending action that disables
	// the pet on expiration.
	timeoutAction *PendingAction

	// Examples:
	// DK Raise Dead is doing its whole RP thing by climbing out of the ground before attacking.
	// Monk clones Rush towards targets before attacking.
	startAttackDelay time.Duration
}

func NewPet(config PetConfig) Pet {
	distanceFromTarget := TernaryFloat64(config.StartsAtOwnerDistance, config.Owner.StartDistanceFromTarget, MaxMeleeRange)

	pet := Pet{
		Character: Character{
			Unit: Unit{
				Type:        PetUnit,
				Index:       config.Owner.Party.Raid.getNextPetIndex(),
				Label:       fmt.Sprintf("%s - %s", config.Owner.Label, config.Name),
				Level:       CharacterLevel,
				PseudoStats: stats.NewPseudoStats(),
				auraTracker: newAuraTracker(),
				Metrics:     NewUnitMetrics(),

				StatDependencyManager: stats.NewStatDependencyManager(),

				ReactionTime:    config.Owner.ReactionTime,
				procChanceOwner: &config.Owner.Unit,

				StartDistanceFromTarget: distanceFromTarget,
				DistanceFromTarget:      distanceFromTarget,
			},
			Name:       config.Name,
			Party:      config.Owner.Party,
			PartyIndex: config.Owner.PartyIndex,
			baseStats:  config.BaseStats,
		},
		Owner:                            config.Owner,
		statInheritance:                  makeStatInheritanceFunc(config.StatInheritance),
		isDynamic:                        config.IsDynamic,
		hasDynamicMeleeSpeedInheritance:  config.HasDynamicMeleeSpeedInheritance && config.IsDynamic,
		inheritedMeleeSpeedMultiplier:    1,
		hasDynamicRangedSpeedInheritance: config.HasDynamicRangedSpeedInheritance && config.IsDynamic,
		inheritedRangedSpeedMultiplier:   1,
		hasDynamicCastSpeedInheritance:   config.HasDynamicCastSpeedInheritance && config.IsDynamic,
		inheritedCastSpeedMultiplier:     1,
		hasResourceRegenInheritance:      config.HasResourceRegenInheritance,
		enabledOnStart:                   config.EnabledOnStart,
		isGuardian:                       config.IsGuardian,
	}

	pet.GCD = pet.NewTimer()
	pet.RotationTimer = pet.NewTimer()

	pet.AddStats(config.BaseStats)
	pet.addUniversalStatDependencies()
	pet.PseudoStats.InFrontOfTarget = config.Owner.PseudoStats.InFrontOfTarget

	// Pre-allocate timeout action since it cannot be pooled.
	pet.timeoutAction = &PendingAction{}

	return pet
}

func (pet *Pet) Initialize() {
	if pet.hasResourceRegenInheritance {
		pet.enableResourceRegenInheritance()
	}
}

func makeStatInheritanceFunc(statInheritance PetStatInheritance) PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		inheritedStats := statInheritance(ownerStats)

		// TODO Apparently fire ele doesnt inherit anything, is it the same for other pets ?

		return inheritedStats
	}
}

func (pet *Pet) inheritOwnerStats(sim *Simulation) {
	pet.inheritedStats = pet.statInheritance(pet.Owner.GetStats())
	pet.AddStatsDynamic(sim, pet.inheritedStats)
}

// Snapshots the owner's stats again: a dynamic pet at the pet heartbeat after its owner's stats
// change, any other pet on its own schedule.
func (pet *Pet) RefreshInheritedStats(sim *Simulation) {
	inheritedStats := pet.statInheritance(pet.Owner.GetStats())
	if inheritedStats == pet.inheritedStats {
		return
	}
	pet.AddStatsDynamic(sim, inheritedStats.Subtract(pet.inheritedStats))
	pet.inheritedStats = inheritedStats
}

func (pet *Pet) enableDynamicStats(sim *Simulation) {
	if !pet.isDynamic {
		return
	}
	if slices.Contains(pet.Owner.DynamicStatsPets, pet) {
		panic("Pet already present in dynamic stats pet list!")
	}

	pet.inheritOwnerStats(sim)
	pet.Owner.DynamicStatsPets = append(pet.Owner.DynamicStatsPets, pet)
	pet.statInheritanceAction = &PendingAction{Priority: ActionPriorityDOT, OnAction: pet.RefreshInheritedStats}
}

func (pet *Pet) resetInheritedOwnerStats(sim *Simulation) {
	pet.AddStatsDynamic(sim, pet.inheritedStats.Invert())
	pet.inheritedStats = stats.Stats{}
}

func (pet *Pet) resetDynamicStats(sim *Simulation) {
	if !pet.isDynamic {
		return
	}

	if idx := slices.Index(pet.Owner.DynamicStatsPets, pet); idx != -1 {
		pet.Owner.DynamicStatsPets = removeBySwappingToBack(pet.Owner.DynamicStatsPets, idx)
	} else {
		panic("Pet not present in dynamic stats pet list!")
	}

	// Drops a refresh still queued for the dismissed pet.
	pet.statInheritanceAction.Cancel(sim)
	pet.resetInheritedOwnerStats(sim)
}

func (pet *Pet) reset(sim *Simulation, agent PetAgent) {
	if pet.isReset {
		return
	}
	pet.isReset = true

	pet.Character.reset(sim, agent)

	pet.CancelGCDTimer(sim)
	pet.AutoAttacks.CancelAutoSwing(sim)

	pet.enabled = false
	if pet.enabledOnStart {
		pet.Enable(sim, agent)
	}
}
func (pet *Pet) doneIteration(sim *Simulation) {
	pet.Character.doneIteration(sim)
	pet.Disable(sim)
	pet.isReset = false
}

func (pet *Pet) IsGuardian() bool {
	return pet.isGuardian
}

func (pet *Pet) EnabledOnStart() bool {
	return pet.enabledOnStart
}

// Sets the pet on the target before the pull: its rotation runs from now on, and it swings once in reach.
func (pet *Pet) AttackBeforePull(sim *Simulation) {
	if sim.CurrentTime >= 0 || pet.engagedBeforePull {
		return
	}
	pet.engagedBeforePull = true
	// A pet summoned before the pull has its first swing queued at 0
	pet.AutoAttacks.CancelMeleeSwing(sim)
	pet.AutoAttacks.mh.swingAt = sim.CurrentTime
	pet.AutoAttacks.EnableMeleeSwing(sim)
	pet.SetGCDTimer(sim, sim.CurrentTime)
}

// Runs enableEffect on every summon, after OnPetEnable, for a caller that does not own the pet.
func (pet *Pet) RegisterEnableEffect(enableEffect OnPetEnable) {
	pet.enableEffects = append(pet.enableEffects, enableEffect)
}

// Runs disableEffect on every dismiss, after OnPetDisable. The dismiss expires only the auras that
// expire on their own, so this is where a permanent aura that should not outlast the pet comes off.
func (pet *Pet) RegisterDisableEffect(disableEffect OnPetDisable) {
	pet.disableEffects = append(pet.disableEffects, disableEffect)
}

// petAgent should be the PetAgent which embeds this Pet.
func (pet *Pet) Enable(sim *Simulation, petAgent PetAgent) {
	if pet.enabled {
		if sim.Log != nil {
			pet.Log(sim, "Pet already summoned")
		}
		return
	}

	// In case of Pre-pull guardian summoning we need to reset
	// TODO: Check if this has side effects
	if !pet.isReset {
		pet.reset(sim, petAgent)
	}

	if pet.isDynamic {
		pet.enableDynamicStats(sim)
	} else {
		pet.inheritOwnerStats(sim)
	}
	//reset current mana after applying stats
	pet.manaBar.reset()
	pet.manaBar.lastManaTick = sim.CurrentTime

	//reset current health after applying stats
	pet.healthBar.reset(sim)

	// Call onEnable callbacks before enabling auto swing
	// to not have to reorder PAs multiple times
	pet.enabled = true

	if pet.OnPetEnable != nil {
		pet.OnPetEnable(sim)
	}
	for _, enableEffect := range pet.enableEffects {
		enableEffect(sim)
	}

	if pet.hasDynamicMeleeSpeedInheritance {
		pet.enableDynamicMeleeSpeed(sim)
	}

	if pet.hasDynamicRangedSpeedInheritance {
		pet.enableDynamicRangedSpeed(sim)
	}

	if pet.hasDynamicCastSpeedInheritance {
		pet.enableDynamicCastSpeed(sim)
	}

	pet.SetGCDTimer(sim, max(0, sim.CurrentTime+pet.startAttackDelay, sim.CurrentTime))
	pet.AutoAttacks.EnableAutoSwing(sim)

	if pet.startAttackDelay > 0 {
		pet.AutoAttacks.StopMeleeUntil(sim, max(SpellBatchWindow, sim.CurrentTime+pet.startAttackDelay)-pet.AutoAttacks.MainhandSwingSpeed())
	}

	if sim.Log != nil {
		pet.Log(sim, "Pet stats: %s", pet.GetStats().FlatString())
		pet.Log(sim, "Pet inherited stats: %s", pet.ApplyStatDependencies(pet.inheritedStats).FlatString())
		pet.Log(sim, "Pet summoned")
	}

	sim.addTracker(&pet.auraTracker)

	if pet.HasFocusBar() {
		// make sure to reset it to refresh focus
		pet.focusBar.reset(sim)
		pet.focusBar.enable(sim, sim.CurrentTime)
	}

	if pet.HasEnergyBar() {
		// make sure to reset it to refresh energy
		pet.energyBar.reset(sim)
		pet.energyBar.enable(sim, sim.CurrentTime)
		if pet.hasResourceRegenInheritance {
			pet.energyBar.energyRegenMultiplier *= pet.Owner.PseudoStats.AttackSpeedMultiplier
		}
	}
}

func (pet *Pet) EnableWithStartAttackDelay(sim *Simulation, petAgent PetAgent, startAttackDelay time.Duration) {
	pet.startAttackDelay = startAttackDelay
	pet.Enable(sim, petAgent)
}

// Helper for enabling a pet that will expire after a certain duration.
func (pet *Pet) EnableWithTimeout(sim *Simulation, petAgent PetAgent, petDuration time.Duration) {
	pet.Enable(sim, petAgent)
	pet.SetTimeoutAction(sim, petDuration)
}

func (pet *Pet) SetTimeoutAction(sim *Simulation, duration time.Duration) {
	if !pet.timeoutAction.consumed {
		pet.timeoutAction.Cancel(sim)
	}

	pet.timeoutAction.cancelled = false
	pet.timeoutAction.NextActionAt = sim.CurrentTime + duration
	pet.timeoutAction.OnAction = pet.Disable
	sim.AddPendingAction(pet.timeoutAction)
}

func (pet *Pet) SetStartAttackDelay(startAttackDelay time.Duration) {
	pet.startAttackDelay = startAttackDelay
}

func (pet *Pet) enableDynamicMeleeSpeed(sim *Simulation) {
	if slices.Contains(pet.Owner.DynamicMeleeSpeedPets, pet) {
		panic("Pet already present in dynamic melee speed pet list!")
	}

	if math.Abs(pet.inheritedMeleeSpeedMultiplier-1) > 1e-14 {
		panic(fmt.Sprintf("Pet melee speed multiplier was not reset properly! Current inherited value = %.17f", pet.inheritedMeleeSpeedMultiplier))
	}

	pet.dynamicMeleeSpeedInheritance = func(sim *Simulation, ownerSpeedMultiplier float64) {
		pet.inheritedMeleeSpeedMultiplier *= ownerSpeedMultiplier
		pet.MultiplyMeleeSpeed(sim, ownerSpeedMultiplier)
	}

	pet.dynamicMeleeSpeedInheritance(sim, pet.Owner.PseudoStats.MeleeSpeedMultiplier)
	pet.dynamicMeleeSpeedInheritance(sim, pet.Owner.PseudoStats.AttackSpeedMultiplier)
	pet.Owner.DynamicMeleeSpeedPets = append(pet.Owner.DynamicMeleeSpeedPets, pet)
}

func (pet *Pet) resetDynamicMeleeSpeed(sim *Simulation) {
	if pet.dynamicMeleeSpeedInheritance == nil {
		return
	}

	if idx := slices.Index(pet.Owner.DynamicMeleeSpeedPets, pet); idx != -1 {
		pet.Owner.DynamicMeleeSpeedPets = removeBySwappingToBack(pet.Owner.DynamicMeleeSpeedPets, idx)
	} else {
		panic("Pet not present in dynamic melee speed pet list!")
	}

	pet.dynamicMeleeSpeedInheritance(sim, 1/pet.inheritedMeleeSpeedMultiplier)
	pet.dynamicMeleeSpeedInheritance = nil
}

func (pet *Pet) enableDynamicRangedSpeed(sim *Simulation) {
	if slices.Contains(pet.Owner.DynamicRangedSpeedPets, pet) {
		panic("Pet already present in dynamic ranged speed pet list!")
	}

	if math.Abs(pet.inheritedRangedSpeedMultiplier-1) > 1e-14 {
		panic(fmt.Sprintf("Pet ranged speed multiplier was not reset properly! Current inherited value = %.17f", pet.inheritedRangedSpeedMultiplier))
	}

	pet.dynamicRangedSpeedInheritance = func(sim *Simulation, ownerSpeedMultiplier float64) {
		pet.inheritedRangedSpeedMultiplier *= ownerSpeedMultiplier
		pet.MultiplyMeleeSpeed(sim, ownerSpeedMultiplier)
	}

	pet.dynamicRangedSpeedInheritance(sim, pet.Owner.PseudoStats.RangedSpeedMultiplier)
	pet.Owner.DynamicRangedSpeedPets = append(pet.Owner.DynamicRangedSpeedPets, pet)
}

func (pet *Pet) resetDynamicRangedSpeed(sim *Simulation) {
	if pet.dynamicRangedSpeedInheritance == nil {
		return
	}

	if idx := slices.Index(pet.Owner.DynamicRangedSpeedPets, pet); idx != -1 {
		pet.Owner.DynamicRangedSpeedPets = removeBySwappingToBack(pet.Owner.DynamicRangedSpeedPets, idx)
	} else {
		panic("Pet not present in dynamic ranged speed pet list!")
	}

	pet.dynamicRangedSpeedInheritance(sim, 1/pet.inheritedRangedSpeedMultiplier)
	pet.dynamicRangedSpeedInheritance = nil
}

func (pet *Pet) enableDynamicCastSpeed(sim *Simulation) {
	if slices.Contains(pet.Owner.DynamicCastSpeedPets, pet) {
		panic("Pet already present in dynamic cast speed pet list!")
	}

	if math.Abs(pet.inheritedCastSpeedMultiplier-1) > 1e-14 {
		panic(fmt.Sprintf("Pet cast speed multiplier was not reset properly! Current inherited value = %.17f", pet.inheritedCastSpeedMultiplier))
	}

	pet.dynamicCastSpeedInheritance = func(sim *Simulation, ownerSpeedMultiplier float64) {
		pet.inheritedCastSpeedMultiplier *= ownerSpeedMultiplier
		pet.MultiplyCastSpeed(sim, ownerSpeedMultiplier)
	}

	pet.dynamicCastSpeedInheritance(sim, pet.Owner.PseudoStats.CastSpeedMultiplier)
	pet.Owner.DynamicCastSpeedPets = append(pet.Owner.DynamicCastSpeedPets, pet)
}

func (pet *Pet) resetDynamicCastSpeed(sim *Simulation) {
	if pet.dynamicCastSpeedInheritance == nil {
		return
	}

	if idx := slices.Index(pet.Owner.DynamicCastSpeedPets, pet); idx != -1 {
		pet.Owner.DynamicCastSpeedPets = removeBySwappingToBack(pet.Owner.DynamicCastSpeedPets, idx)
	} else {
		panic("Pet not present in dynamic cast speed pet list!")
	}

	pet.dynamicCastSpeedInheritance(sim, 1/pet.inheritedCastSpeedMultiplier)
	pet.dynamicCastSpeedInheritance = nil
}

func (pet *Pet) enableResourceRegenInheritance() {
	if !slices.Contains(pet.Owner.RegenInheritancePets, pet) {
		pet.Owner.RegenInheritancePets = append(pet.Owner.RegenInheritancePets, pet)
	}
}

func (pet *Pet) Disable(sim *Simulation) {
	if !pet.enabled {
		if sim.Log != nil {
			pet.Log(sim, "No pet summoned")
		}
		return
	}

	if pet.isDynamic {
		pet.resetDynamicStats(sim)
	} else {
		pet.resetInheritedOwnerStats(sim)
	}
	pet.resetDynamicMeleeSpeed(sim)
	pet.resetDynamicRangedSpeed(sim)
	pet.resetDynamicCastSpeed(sim)
	pet.CancelGCDTimer(sim)
	pet.focusBar.disable(sim)
	pet.energyBar.disable(sim)
	pet.AutoAttacks.CancelAutoSwing(sim)
	pet.enabled = false

	// If a pet is immediately re-summoned it might try to use GCD, so we need to clear it.
	pet.Hardcast = Hardcast{}

	if !pet.timeoutAction.consumed {
		pet.timeoutAction.Cancel(sim)
	}

	if pet.OnPetDisable != nil {
		pet.OnPetDisable(sim)
	}
	for _, disableEffect := range pet.disableEffects {
		disableEffect(sim)
	}

	pet.auraTracker.expireAll(sim)

	sim.removeTracker(&pet.auraTracker)

	if sim.Log != nil {
		pet.Log(sim, "Pet dismissed")
		pet.Log(sim, pet.GetStats().FlatString())
	}
}
func (pet *Pet) ChangeStatInheritance(statInheritance PetStatInheritance) {
	pet.statInheritance = makeStatInheritanceFunc(statInheritance)
}

func (pet *Pet) GetInheritedStats() stats.Stats {
	return pet.inheritedStats
}

func (pet *Pet) DisableOnStart() {
	pet.enabledOnStart = false
}

// Default implementations for some Agent functions which most Pets don't need.
func (pet *Pet) GetCharacter() *Character {
	return &pet.Character
}
func (pet *Pet) AddRaidBuffs(_ *proto.RaidBuffs)   {}
func (pet *Pet) AddPartyBuffs(_ *proto.PartyBuffs) {}
func (pet *Pet) ApplyTalents()                     {}
func (pet *Pet) OnGCDReady(_ *Simulation)          {}

func (env *Environment) TriggerDelayedPetInheritance(sim *Simulation, dynamicPets []*Pet, inheritanceFunc func(*Simulation, *Pet)) {
	for _, pet := range dynamicPets {
		if !pet.IsEnabled() {
			continue
		}

		numHeartbeats := (sim.CurrentTime - env.heartbeatOffset) / PetUpdateInterval
		nextHeartbeat := PetUpdateInterval*(numHeartbeats+1) + env.heartbeatOffset

		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = nextHeartbeat
		pa.Priority = ActionPriorityDOT

		pa.OnAction = func(sim *Simulation) {
			if pet.enabled {
				inheritanceFunc(sim, pet)
			}
		}

		sim.AddPendingAction(pa)
	}
}

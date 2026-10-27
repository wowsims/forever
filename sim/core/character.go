package core

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protowire"
	googleproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type CharacterBuildPhase uint8

func (cbp CharacterBuildPhase) Matches(other CharacterBuildPhase) bool {
	return (cbp & other) != 0
}

const (
	CharacterBuildPhaseNone CharacterBuildPhase = 0
	CharacterBuildPhaseBase CharacterBuildPhase = 1 << iota
	CharacterBuildPhaseGear
	CharacterBuildPhaseTalents
	CharacterBuildPhaseBuffs
	CharacterBuildPhaseConsumes
)

const CharacterBuildPhaseAll = CharacterBuildPhaseBase | CharacterBuildPhaseGear | CharacterBuildPhaseTalents | CharacterBuildPhaseBuffs | CharacterBuildPhaseConsumes

type RotationTransformation func(raid *proto.Raid, rotation *proto.APLRotation)

// Character is a data structure to hold all the shared values that all
// class logic shares.
// All players have stats, equipment, auras, etc
type Character struct {
	Unit

	Name  string // Different from Label, needed for returned results.
	Race  proto.Race
	Class proto.Class
	Spec  proto.Spec

	// Current gear.
	Equipment

	// Stat buff auras associated with any proc effects in the Character's equippable items
	ItemProcBuffs []*StatBuffAura

	//Item Swap Handler
	ItemSwap ItemSwap

	// Consumables this Character will be using.
	Consumables *proto.ConsumesSpec

	// Base stats for this Character.
	baseStats stats.Stats

	// Bonus stats for this Character, specified in the UI and/or EP
	// calculator
	bonusStats     stats.Stats
	bonusMHDps     float64
	bonusOHDps     float64
	bonusRangedDps float64

	professions    [2]proto.Profession
	professionRank [2]int32 // 0 is the maximum.

	// Used for effects like "Increased Armor Value from Items"
	*EquipScalingManager

	// Provides major cooldown management behavior.
	majorCooldownManager

	// Listed in any of the raid's tank slots.
	isTank bool

	// Up reference to this Character's Party.
	Party *Party

	// This character's index within its party [0-4].
	PartyIndex int

	Pets []*Pet // cached in AddPet, for advance()

	talents []*proto.ActionID // Every talent with points in it, at its talent_spell_id and held rank.

	// Used for manually modifying a rotation before it's constructed
	rotationTransformations []RotationTransformation
}

func NewCharacter(party *Party, partyIndex int, player *proto.Player) Character {
	if player.Database != nil {
		addToDatabase(player.Database)
	}

	character := Character{
		Unit: Unit{
			Type:        PlayerUnit,
			Index:       int32(party.Index*5 + partyIndex),
			Level:       CharacterLevel,
			auraTracker: newAuraTracker(),
			PseudoStats: stats.NewPseudoStats(),
			Metrics:     NewUnitMetrics(),

			StatDependencyManager: stats.NewStatDependencyManager(),

			ReactionTime:            time.Duration(max(player.ReactionTimeMs, 10)) * time.Millisecond,
			ChannelClipDelay:        max(0, time.Duration(player.ChannelClipDelayMs)*time.Millisecond),
			StartDistanceFromTarget: player.DistanceFromTarget,
			DistanceFromTarget:      player.DistanceFromTarget,
		},

		Name:  player.Name,
		Race:  player.Race,
		Class: player.Class,
		Spec:  PlayerProtoToSpec(player),

		Equipment: ProtoToEquipment(player.Equipment).inArea(party.areaTypes()),

		professions: [2]proto.Profession{
			player.Profession1,
			player.Profession2,
		},
		professionRank: [2]int32{
			player.Prof1Rank,
			player.Prof2Rank,
		},

		Party:      party,
		PartyIndex: partyIndex,

		majorCooldownManager: newMajorCooldownManager(player.Cooldowns),
	}
	character.GCD = character.NewTimer()
	character.RotationTimer = character.NewTimer()

	character.Label = fmt.Sprintf("%s (#%d)", character.Name, character.Index+1)

	character.Consumables = &proto.ConsumesSpec{}
	if player.Consumables != nil {
		character.Consumables = player.Consumables
	}
	consumables := character.Consumables
	applyTemporaryEnchants(consumables.MhImbueId, consumables.OhImbueId, consumables.MhWeaponImbueId, consumables.OhWeaponImbueId, &character.Equipment)

	character.baseStats = BaseStats[BaseStatsKey{Race: character.Race, Class: character.Class}]

	character.AddStats(character.baseStats)
	character.AddStat(stats.Health, 20-10*20)
	character.addUniversalStatDependencies()
	character.AddStatDependency(stats.Agility, stats.DodgePercent, dodgePerAgility(character.Class))
	character.PseudoStats.BaseDodgeChance += ClassBaseDodge[character.Class]
	if _, ok := ClassWeaponTypeCapabilities[character.Class][proto.WeaponType_WeaponTypeShield]; ok {
		character.PseudoStats.BlockValuePerStrength = 0.05
	}

	if player.BonusStats != nil {
		if player.BonusStats.Stats != nil {
			character.bonusStats = stats.FromUnitStatsProto(player.BonusStats)
		}
		if player.BonusStats.PseudoStats != nil {
			ps := player.BonusStats.PseudoStats
			character.bonusMHDps = ps[proto.PseudoStat_PseudoStatMainHandDps]
			character.bonusOHDps = ps[proto.PseudoStat_PseudoStatOffHandDps]
			character.bonusRangedDps = ps[proto.PseudoStat_PseudoStatRangedDps]
			character.PseudoStats.BaseReducedCritTakenPercent = ps[proto.PseudoStat_PseudoStatReducedCritTakenPercent] / 100
			for school := stats.SchoolIndexArcane; school < stats.SchoolLen; school++ {
				character.PseudoStats.SchoolBonusHitChance[school] = ps[SchoolHitPseudoStats[school]]
			}
			character.PseudoStats.BonusMHDps += character.bonusMHDps
			character.PseudoStats.BonusOHDps += character.bonusOHDps
			character.PseudoStats.BonusRangedDps += character.bonusRangedDps
		}
	}

	if weapon := character.OffHand(); weapon.ID != 0 {
		if weapon.WeaponType == proto.WeaponType_WeaponTypeShield {
			character.PseudoStats.CanBlock = true
		}
	}
	character.PseudoStats.InFrontOfTarget = player.InFrontOfTarget
	character.applyWeaponSkills(player.WeaponSkills)

	if player.EnableItemSwap && player.ItemSwap != nil {
		character.enableItemSwap(player.ItemSwap)
	}

	character.EquipScalingManager = character.NewEquipScalingManager()

	return character
}

type EquipScalingManager struct {
	itemStatMultipliers map[stats.Stat]float64
	cachedEquipStats    stats.Stats
	equipStatsApplied   bool
	equipCacheValid     bool
}

func (character *Character) NewEquipScalingManager() *EquipScalingManager {
	return &EquipScalingManager{
		itemStatMultipliers: make(map[stats.Stat]float64),
		cachedEquipStats:    character.Equipment.Stats(character.Spec).Add(character.bonusStats),
		equipCacheValid:     true,
	}
}

func (character *Character) AddDynamicEquipStats(sim *Simulation, equipStats stats.Stats) {
	character.AddStatsDynamic(sim, equipStats.ApplyMultipliers(character.itemStatMultipliers))
	character.equipCacheValid = false
}

func (character *Character) applyEquipScalingInternal(stat stats.Stat, multiplier float64) float64 {
	character.updateCachedEquipStats()
	oldMultiplier, exists := character.itemStatMultipliers[stat]

	if !exists {
		oldMultiplier = 1.0
	}

	newMultiplier := oldMultiplier * multiplier
	character.itemStatMultipliers[stat] = newMultiplier

	return character.cachedEquipStats[stat] * (newMultiplier - oldMultiplier)
}

func (character *Character) ApplyEquipScaling(stat stats.Stat, multiplier float64) {
	statDiff := character.applyEquipScalingInternal(stat, multiplier)
	// Equipment stats already applied, so need to manually at the bonus to
	// the character now to ensure correct values
	if character.equipStatsApplied {
		character.AddStat(stat, statDiff)
	}
}

func (character *Character) ApplyDynamicEquipScaling(sim *Simulation, stat stats.Stat, multiplier float64) {
	if character.Env.MeasuringStats && (character.Env.State != Finalized) {
		character.ApplyEquipScaling(stat, multiplier)
	} else {
		statDiff := character.applyEquipScalingInternal(stat, multiplier)
		character.AddStatDynamic(sim, stat, statDiff)
	}
}

func (character *Character) RemoveEquipScaling(stat stats.Stat, multiplier float64) {
	character.ApplyEquipScaling(stat, 1/multiplier)
}

func (character *Character) RemoveDynamicEquipScaling(sim *Simulation, stat stats.Stat, multiplier float64) {
	character.ApplyDynamicEquipScaling(sim, stat, 1/multiplier)
}

func (character *Character) updateCachedEquipStats() {
	if !character.equipCacheValid {
		character.cachedEquipStats = character.Equipment.Stats(character.Spec).Add(character.bonusStats)
		character.equipCacheValid = true
	}
}

func (character *Character) EquipStats() stats.Stats {
	character.updateCachedEquipStats()
	return character.cachedEquipStats.ApplyMultipliers(character.itemStatMultipliers)
}

func (character *Character) applyEquipment() {
	if character.EquipScalingManager == nil {
		return
	}

	if character.equipStatsApplied {
		panic("Equipment stats already applied to character!")
	}
	character.AddStats(character.EquipStats())
	character.equipStatsApplied = true
}

func (character *Character) addUniversalStatDependencies() {
	character.Unit.addUniversalStatDependencies()
	character.AddStatDependency(stats.Stamina, stats.Health, 10)
	character.AddStatDependency(stats.Agility, stats.Armor, 2)
}

// Returns a partially-filled PlayerStats proto for use in the CharacterStats api call.
func (character *Character) applyAllEffects(agent Agent, raidBuffs *proto.RaidBuffs, partyBuffs *proto.PartyBuffs, individualBuffs *proto.IndividualBuffs) *proto.PlayerStats {
	playerStats := &proto.PlayerStats{}

	measureStats := func() *proto.UnitStats {
		baseStats := character.GetStats()
		character.stats = character.SortAndApplyStatDependencies(character.stats).FloorGameStats()
		measuredStatsProto := &proto.UnitStats{
			Stats:       character.GetStats().ToProtoArray(),
			PseudoStats: character.GetPseudoStatsProto(),
			ApiVersion:  GetCurrentProtoVersion(),
		}
		character.stats = baseStats
		return measuredStatsProto
	}

	character.WithSource(SourceRacial, func() { applyRaceEffects(agent) })
	character.WithSource(SourceProfession, character.applyProfessionEffects)
	character.applyBuildPhaseAuras(CharacterBuildPhaseBase)
	playerStats.BaseStats = measureStats()

	character.applyEquipment()
	character.WithSource(SourceItem, func() {
		character.applyItemEffects(agent)
		character.applyItemSetBonusEffects(agent)
	})
	character.applyBuildPhaseAuras(CharacterBuildPhaseGear)
	playerStats.GearStats = measureStats()

	character.WithSource(SourceTalent, agent.ApplyTalents)
	character.applyBuildPhaseAuras(CharacterBuildPhaseTalents)
	playerStats.TalentsStats = measureStats()

	character.WithSource(SourceExternalBuff, func() {
		registeredBuffs().ApplyBuffs(character, raidBuffs, partyBuffs, individualBuffs)
	})
	character.applyBuildPhaseAuras(CharacterBuildPhaseBuffs)
	playerStats.BuffsStats = measureStats()

	character.WithSource(SourceConsumable, func() { applyConsumeEffects(agent) })
	character.applyBuildPhaseAuras(CharacterBuildPhaseConsumes)
	playerStats.ConsumesStats = measureStats()
	character.clearBuildPhaseAuras(CharacterBuildPhaseAll)

	return playerStats
}
func (character *Character) applyBuildPhaseAuras(phase CharacterBuildPhase) {
	sim := Simulation{}
	character.Env.MeasuringStats = true
	for _, aura := range character.auras {
		if aura.BuildPhase.Matches(phase) {
			aura.Activate(&sim)
		}
	}
	character.Env.MeasuringStats = false
}
func (character *Character) clearBuildPhaseAuras(phase CharacterBuildPhase) {
	sim := Simulation{}
	character.Env.MeasuringStats = true
	for _, aura := range character.auras {
		if aura.BuildPhase.Matches(phase) {
			aura.Deactivate(&sim)
		}
	}
	character.Env.MeasuringStats = false
}

// Apply effects from all equipped core.
func (character *Character) applyItemEffects(agent Agent) {
	registeredItemEffects := make(map[int32]bool)
	registeredItemEnchantEffects := make(map[int32]bool)

	character.Equipment.applyItemEffects(agent, registeredItemEffects, registeredItemEnchantEffects)

	if character.ItemSwap.IsEnabled() {
		character.ItemSwap.unEquippedItems.applyItemEffects(agent, registeredItemEffects, registeredItemEnchantEffects)
	}

	character.registerEquipSpeedAuras()
}

func (character *Character) AddPet(pet PetAgent) {
	if character.Env != nil {
		panic("Pets must be added during construction!")
	}

	character.PetAgents = append(character.PetAgents, pet)
	character.Pets = append(character.Pets, pet.GetPet())
}

func (character *Character) GetBaseStats() stats.Stats {
	return character.baseStats
}

func (character *Character) AddRaidBuffs(_ *proto.RaidBuffs) {
}
func (character *Character) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {
}

func (character *Character) initialize(agent Agent) {
	character.majorCooldownManager.initialize(character)
	character.ItemSwap.initialize(character)
	character.RegisterWeaponChangeCallback(func(_ *Simulation) {
		character.refilterEquipMods()
	})

	character.rotationAction = &PendingAction{
		Priority: ActionPriorityGCD,
		OnAction: func(sim *Simulation) {
			character.completeDueHardcast(sim)

			if sim.CurrentTime < 0 && !character.engagedBeforePull {
				return
			}

			if sim.Options.Interactive {
				if character.GCD.IsReady(sim) {
					sim.NeedsInput = true
				}
				return
			}

			if character.Rotation != nil {
				character.Rotation.DoNextAction(sim)
				return
			}
		},
	}
}

func (character *Character) Finalize() {
	if character.Env.IsFinalized() {
		return
	}

	character.PseudoStats.ParryHaste = character.PseudoStats.CanParry

	if character.Unit.Metrics.isTanking {
		character.HardcastAvoidanceAura = character.RegisterAura(Aura{
			Label:    "Reduced avoidance",
			Tag:      ReducedAvoidanceAuraTag,
			Duration: NeverExpires,
			// Stunned is shared with the stun kind, so both derive it rather
			// than writing it, otherwise whichever ends first clears it.
			OnGain: func(aura *Aura, _ *Simulation) {
				aura.Unit.refreshIncapacitateState()
			},
			OnExpire: func(aura *Aura, _ *Simulation) {
				aura.Unit.refreshIncapacitateState()
			},
		})

		character.MakeProcTriggerAura(ProcTrigger{
			Name:               "Pushback trigger",
			Callback:           CallbackOnSpellHitTaken,
			Outcome:            OutcomeLanded,
			RequireDamageDealt: true,

			ExtraCondition: func(sim *Simulation, spell *Spell, result *SpellResult) bool {
				return character.Hardcast.Expires > sim.CurrentTime &&
					(character.Hardcast.IsChanneled || character.Hardcast.Pushback) &&
					// Dots will not trigger pushback
					!(spell.dots != nil || spell.aoeDot != nil || (spell.RelatedDotSpell != nil && (spell.RelatedDotSpell.dots != nil || spell.RelatedDotSpell.aoeDot != nil)))
			},

			Handler: func(sim *Simulation, spell *Spell, result *SpellResult) {
				if !sim.Proc(character.PseudoStats.PushbackChance, "Pushback") {
					return
				}

				if character.Hardcast.IsChanneled {
					// Channeled spells will lose 25% of their total duration
					pushback := character.Hardcast.CastTime / 4
					character.Hardcast.Expires = max(sim.CurrentTime, character.Hardcast.Expires-pushback)

					if sim.Log != nil {
						character.Log(sim, "%s pushed back %s while channeling", character.Hardcast.ActionID, pushback)
					}
				} else {
					pushback := character.Hardcast.pushBack(sim.CurrentTime)
					if pushback <= 0 {
						return
					}

					if sim.Log != nil {
						character.Log(sim, "%s pushed back %s while casting", character.Hardcast.ActionID, pushback)
					}
				}

				// Re-schedule the cast at the new Expires time
				character.newHardcastAction(sim)
			},
		})
	}

	character.Unit.finalize()
	character.basePhysicalCritPercent = character.baseStats[stats.PhysicalCritPercent]
	character.critPerAgility = character.GetStatDependencyCoeff(stats.Agility, stats.PhysicalCritPercent)

	character.majorCooldownManager.finalize()
}

func (character *Character) FillPlayerStats(playerStats *proto.PlayerStats) {
	if playerStats == nil {
		return
	}

	character.applyBuildPhaseAuras(CharacterBuildPhaseAll)
	playerStats.FinalStats = &proto.UnitStats{
		Stats:       character.GetStats().ToProtoArray(),
		PseudoStats: character.GetPseudoStatsProto(),
		ApiVersion:  GetCurrentProtoVersion(),
	}
	playerStats.WeaponSkills = character.weaponSkillsProto()

	character.clearBuildPhaseAuras(CharacterBuildPhaseAll)
	playerStats.Sets = character.GetActiveSetBonusNames()

	playerStats.Metadata = character.GetMetadata()
	playerStats.Metadata.Talents = character.talents
	for _, pet := range character.Pets {
		playerStats.Pets = append(playerStats.Pets, &proto.PetStats{
			Metadata: pet.GetMetadata(),
		})
	}

	if character.Rotation != nil {
		playerStats.RotationStats = character.Rotation.getStats()
	}
}

func (character *Character) reset(sim *Simulation, agent Agent) {
	character.Unit.reset(sim, agent)
	character.majorCooldownManager.reset(sim)
	character.CurrentTarget = character.defaultTarget

	agent.Reset(sim)

	character.ItemSwap.reset(sim)

	for _, petAgent := range character.PetAgents {
		petAgent.GetPet().reset(sim, petAgent)
	}
}

func (character *Character) HasProfession(prof proto.Profession) bool {
	return prof == character.professions[0] || prof == character.professions[1]
}

// Whether the character has prof at rank or higher. ProfessionUnknown is always met.
func (character *Character) MeetsProfession(prof proto.Profession, rank int32) bool {
	if prof == proto.Profession_ProfessionUnknown {
		return true
	}
	for i, p := range character.professions {
		if p == prof && (character.professionRank[i] == 0 || character.professionRank[i] >= rank) {
			return true
		}
	}
	return false
}

func (character *Character) HasTrinketEquipped(itemID int32) bool {
	return character.Trinket1().ID == itemID ||
		character.Trinket2().ID == itemID
}

func (character *Character) HasRingEquipped(itemID int32) bool {
	return character.Finger1().ID == itemID || character.Finger2().ID == itemID
}

// Returns the MH weapon if one is equipped, and null otherwise.
func (character *Character) GetMHWeapon() *Item {
	weapon := character.MainHand()
	if weapon.ID == 0 {
		return nil
	}
	return weapon
}
func (character *Character) HasMHWeapon() bool {
	return character.GetMHWeapon() != nil
}

// Returns the OH weapon if one is equipped, and null otherwise. Note that
// shields / Held-in-off-hand items are NOT counted as weapons in this function.
func (character *Character) GetOHWeapon() *Item {
	weapon := character.OffHand()
	if weapon.ID == 0 ||
		weapon.WeaponType == proto.WeaponType_WeaponTypeShield ||
		weapon.WeaponType == proto.WeaponType_WeaponTypeOffHand {
		return nil
	} else {
		return weapon
	}
}
func (character *Character) HasOHWeapon() bool {
	return character.GetOHWeapon() != nil
}
func (character *Character) HasOH() bool {
	return character.OffHand().ID != 0
}

// Returns the ranged weapon if one is equipped, and null otherwise.
func (character *Character) GetRangedWeapon() *Item {
	weapon := character.Ranged()
	if weapon.ID == 0 ||
		weapon.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeIdol ||
		weapon.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeLibram ||
		weapon.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeTotem ||
		weapon.RangedWeaponType == proto.RangedWeaponType_RangedWeaponTypeSigil {
		return nil
	} else {
		return weapon
	}
}

func (character *Character) HasRangedWeapon() bool {
	return character.Ranged() != nil
}

func (character *Character) GetDynamicProcMaskForWeaponEnchant(effectID int32) *ProcMask {
	return character.getDynamicProcMaskPointer(func() ProcMask {
		return character.getCurrentProcMaskForWeaponEnchant(effectID)
	})
}

func (character *Character) GetDynamicProcMaskForWeaponTempEnchant(effectID int32) *ProcMask {
	return character.getDynamicProcMaskPointer(func() ProcMask {
		return character.getCurrentProcMaskForWeaponTempEnchant(effectID)
	})
}

func (character *Character) getDynamicProcMaskPointer(procMaskFn func() ProcMask) *ProcMask {
	procMask := procMaskFn()

	character.RegisterItemSwapCallback(AllWeaponSlots(), func(sim *Simulation, slot proto.ItemSlot) {
		procMask = procMaskFn()
	})

	return &procMask
}

func (character *Character) getCurrentProcMaskForWeaponEnchant(effectID int32) ProcMask {
	return character.getCurrentProcMaskFor(func(weapon *Item) bool {
		return weapon.Enchant.EffectID == effectID
	})
}

func (character *Character) getCurrentProcMaskForWeaponTempEnchant(effectID int32) ProcMask {
	return character.getCurrentProcMaskFor(func(weapon *Item) bool {
		return weapon.HasTempEnchant(effectID)
	})
}

func (character *Character) GetDynamicProcMaskForWeaponEffect(itemID int32) *ProcMask {
	return character.getDynamicProcMaskPointer(func() ProcMask {
		return character.getCurrentProcMaskForWeaponEffect(itemID)
	})
}

func (character *Character) getCurrentProcMaskForWeaponEffect(itemID int32) ProcMask {
	return character.getCurrentProcMaskFor(func(weapon *Item) bool {
		return weapon.ID == itemID
	})
}

func (character *Character) GetDynamicProcMaskForTypes(weaponTypes ...proto.WeaponType) *ProcMask {
	return character.getDynamicProcMaskPointer(func() ProcMask {
		return character.getCurrentProcMaskFor(func(weapon *Item) bool {
			return weapon != nil && slices.Contains(weaponTypes, weapon.WeaponType)
		})
	})
}

func (character *Character) GetProcMaskForTypes(weaponTypes ...proto.WeaponType) ProcMask {
	return character.getCurrentProcMaskFor(func(weapon *Item) bool {
		return weapon != nil && slices.Contains(weaponTypes, weapon.WeaponType)
	})
}

func (character *Character) GetDynamicProcMaskForTypesAndHand(twohand bool, weaponTypes ...proto.WeaponType) *ProcMask {
	return character.getDynamicProcMaskPointer(func() ProcMask {
		return character.getCurrentProcMaskFor(func(weapon *Item) bool {
			return weapon != nil && (weapon.HandType == proto.HandType_HandTypeTwoHand) == twohand && slices.Contains(weaponTypes, weapon.WeaponType)
		})
	})
}

func (character *Character) GetProcMaskForTypesAndHand(twohand bool, weaponTypes ...proto.WeaponType) ProcMask {
	return character.getCurrentProcMaskFor(func(weapon *Item) bool {
		return weapon != nil && (weapon.HandType == proto.HandType_HandTypeTwoHand) == twohand && slices.Contains(weaponTypes, weapon.WeaponType)
	})
}

func (character *Character) getCurrentProcMaskFor(pred func(item *Item) bool) ProcMask {
	mask := ProcMaskUnknown

	if character == nil {
		return mask
	}

	if pred(character.Ranged()) {
		mask |= ProcMaskRanged
	}
	if pred(character.MainHand()) {
		mask |= ProcMaskMeleeMH
	}
	if pred(character.OffHand()) {
		mask |= ProcMaskMeleeOH
	}
	return mask
}

func (character *Character) GetProcMaskForWeaponSlot(slot proto.ItemSlot) ProcMask {
	switch slot {
	case proto.ItemSlot_ItemSlotRanged:
		return ProcMaskRanged
	case proto.ItemSlot_ItemSlotMainHand:
		return ProcMaskMeleeMH
	case proto.ItemSlot_ItemSlotOffHand:
		return ProcMaskMeleeOH
	}

	return ProcMaskEmpty
}

func (character *Character) doneIteration(sim *Simulation) {
	character.ItemSwap.doneIteration(sim)

	// Need to do pets first, so we can add their results to the owners.
	for _, pet := range character.Pets {
		pet.doneIteration(sim)
		character.Metrics.AddFinalPetMetrics(&pet.Metrics)
	}

	character.Unit.doneIteration(sim)
}

func (character *Character) GetMetricsProto() *proto.UnitMetrics {
	metrics := character.Metrics.ToProto()
	metrics.Name = character.Name
	metrics.UnitIndex = character.UnitIndex
	metrics.Auras = character.auraTracker.GetMetricsProto()

	metrics.Pets = make([]*proto.UnitMetrics, len(character.Pets))
	for i, pet := range character.Pets {
		metrics.Pets[i] = pet.GetMetricsProto()
	}

	return metrics
}

func (character *Character) GetDefensiveTrinketCD() *Timer {
	return character.GetOrInitSpellCategoryTimer(1190)
}
func (character *Character) GetCombatConsumableCD() *Timer {
	return character.GetOrInitSpellCategoryTimer(1153)
}
func (character *Character) GetOffensiveTrinketCD() *Timer {
	return character.GetOrInitSpellCategoryTimer(1141)
}
func (character *Character) GetConjuredCD() *Timer {
	return character.GetOrInitSpellCategoryTimer(30)
}

func (character *Character) AddStatProcBuff(effectID int32, procAura *StatBuffAura, isEnchant bool, eligibleSlots []proto.ItemSlot) {
	hasEquippedCheck := Ternary(isEnchant, character.Equipment.containsEnchantInSlots, character.Equipment.containsItemInSlots)

	procAura.IsSwapped = !hasEquippedCheck(effectID, eligibleSlots)
	character.ItemProcBuffs = append(character.ItemProcBuffs, procAura)

	character.RegisterItemSwapCallback(eligibleSlots, func(sim *Simulation, slot proto.ItemSlot) {
		procAura.IsSwapped = !hasEquippedCheck(effectID, eligibleSlots)
	})
}

func (character *Character) GetMatchingItemProcAuras(statTypesToMatch []stats.Stat, minIcd time.Duration) []*StatBuffAura {
	includeIcdFilter := (minIcd > 0)
	return FilterSlice(character.ItemProcBuffs, func(aura *StatBuffAura) bool {
		return aura.BuffsMatchingStat(statTypesToMatch) && (!includeIcdFilter || ((aura.Icd != nil) && (aura.Icd.Duration > minIcd)))
	})
}

// Uses proto reflection to set fields in a talents proto (e.g. MageTalents,
// WarriorTalents) based on a talentsStr. The character keeps every talent with points in it, which
// TalentRank and the unit metadata read.
func (character *Character) FillTalentsProto(data protoreflect.Message, talentsStr string, treeSizes [3]int) {
	treeStrs := strings.Split(talentsStr, "-")
	fieldDescriptors := data.Descriptor().Fields()
	character.talents = nil

	var offset int
	for treeIdx, treeStr := range treeStrs {
		for talentIdx, talentValStr := range treeStr {
			talentVal, _ := strconv.Atoi(string(talentValStr))
			talentOffset := offset + talentIdx + 1
			fd := fieldDescriptors.ByNumber(protowire.Number(talentOffset))
			if fd == nil {
				panic(fmt.Sprintf("Couldn't find proto field for talent #%d, full string: %s", talentOffset, talentsStr))
			}
			rank := int32(talentVal)
			if fd.Kind() == protoreflect.BoolKind {
				rank = TernaryInt32(talentVal == 1, 1, 0)
				data.Set(fd, protoreflect.ValueOfBool(rank == 1))
			} else { // Int32Kind
				data.Set(fd, protoreflect.ValueOfInt32(rank))
			}
			if spellID := googleproto.GetExtension(fd.Options(), proto.E_TalentSpellId).(int32); spellID != 0 && rank > 0 {
				id := ActionID{SpellID: spellID}.ToProto()
				id.Rank = rank
				character.talents = append(character.talents, id)
			}
		}
		offset += treeSizes[treeIdx]
	}
}

// The points held in the talent whose field carries this talent_spell_id, or 0.
func (character *Character) TalentRank(spellID int32) int32 {
	for _, talent := range character.talents {
		if talent.GetSpellId() == spellID {
			return talent.Rank
		}
	}
	return 0
}

func (character *Character) RegisterRotationTransformation(transformation RotationTransformation) {
	character.rotationTransformations = append(character.rotationTransformations, transformation)
}

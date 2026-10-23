package stats

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

type Stats [SimStatsLen]float64

type Stat byte

// Use internal representation instead of proto.Stat so we can add functions
// and use 'byte' as the data type.
//
// This needs to stay synced with proto.Stat: it is okay for SimStatsLen to
// exceed ProtoStatsLen, but the shared indices between the two must match 1:1 .
const (
	Strength Stat = iota
	Agility
	Stamina
	Intellect
	HealingPower
	SpellDamage
	ArcaneDamage
	FireDamage
	FrostDamage
	HolyDamage
	NatureDamage
	ShadowDamage
	SpellPiercing
	Spirit
	AttackPower
	RangedAttackPower
	FeralAttackPower
	HitRating
	CritRating
	HasteRating
	ArmorPenetration
	ExpertiseRating
	DefenseRating
	BlockRating
	BlockValue
	DodgeRating
	ParryRating
	Armor
	BonusArmor
	Health
	Mana
	MP5
	ArcaneResistance
	FireResistance
	FrostResistance
	NatureResistance
	ShadowResistance
	PhysicalDamage
	ArmorNoMods
	HP5
	// end of Stat enum in proto/common.proto

	// The remaining stats below are stored as PseudoStats rather than as
	// Stats in UnitStats proto messages, since they are not required in the
	// database files. However, it is valuable to keep these as proper Stats
	// in the back-end, since they are used in various stat dependencies.
	// The units for all 11 of these are percentages (between 0 and 100).
	PhysicalHitPercent
	SpellHitPercent
	PhysicalCritPercent
	SpellCritPercent
	BlockPercent
	RangedHitPercent
	RangedCritPercent
	DodgePercent
	ParryPercent
	ReducedCritTakenPercent
	ExpertisePercent
	// DO NOT add new stats here without discussing it first; new stats come
	// with a performance penalty.

	SimStatsLen
)

var ProtoStatsLen = len(proto.Stat_name)
var PseudoStatsLen = len(proto.PseudoStat_name)
var UnitStatsLen = ProtoStatsLen + PseudoStatsLen

type SchoolIndex byte

const (
	SchoolIndexNone     SchoolIndex = 0
	SchoolIndexPhysical SchoolIndex = iota
	SchoolIndexArcane
	SchoolIndexFire
	SchoolIndexFrost
	SchoolIndexHoly
	SchoolIndexNature
	SchoolIndexShadow

	SchoolLen
)

func NewSchoolFloatArray() [SchoolLen]float64 {
	return [SchoolLen]float64{
		1, 1, 1, 1, 1, 1, 1, 1,
	}
}

func ProtoArrayToStatsList(protoStats []proto.Stat) []Stat {
	stats := make([]Stat, len(protoStats))
	for i, v := range protoStats {
		stats[i] = Stat(v)
	}
	return stats
}

func IntTupleToStatsList(statType1 int32, statType2 int32, statType3 int32) []Stat {
	statTypes := make([]Stat, 0, 3)

	for _, statIdx := range []int32{statType1, statType2, statType3} {
		if statIdx >= 0 {
			statTypes = append(statTypes, Stat(statIdx))
		}
	}

	return statTypes
}

func (s Stat) StatName() string {
	switch s {
	case Strength:
		return "Strength"
	case Agility:
		return "Agility"
	case Stamina:
		return "Stamina"
	case Intellect:
		return "Intellect"
	case Spirit:
		return "Spirit"
	case SpellPiercing:
		return "SpellPiercing"
	case HitRating:
		return "HitRating"
	case CritRating:
		return "CritRating"
	case HasteRating:
		return "HasteRating"
	case ExpertiseRating:
		return "ExpertiseRating"
	case ArmorPenetration:
		return "ArmorPenetration"
	case DodgeRating:
		return "DodgeRating"
	case ParryRating:
		return "ParryRating"
	case AttackPower:
		return "AttackPower"
	case RangedAttackPower:
		return "RangedAttackPower"
	case FeralAttackPower:
		return "FeralAttackPower"
	case HealingPower:
		return "HealingPower"
	case SpellDamage:
		return "SpellDamage"
	case ArcaneDamage:
		return "ArcaneDamage"
	case FireDamage:
		return "FireDamage"
	case FrostDamage:
		return "FrostDamage"
	case HolyDamage:
		return "HolyDamage"
	case NatureDamage:
		return "NatureDamage"
	case ShadowDamage:
		return "ShadowDamage"
	case Armor:
		return "Armor"
	case BonusArmor:
		return "BonusArmor"
	case ArmorNoMods:
		return "ArmorNoMods"
	case Health:
		return "Health"
	case HP5:
		return "HP5"
	case Mana:
		return "Mana"
	case MP5:
		return "MP5"
	case PhysicalHitPercent:
		return "PhysicalHitPercent"
	case SpellHitPercent:
		return "SpellHitPercent"
	case PhysicalCritPercent:
		return "PhysicalCritPercent"
	case SpellCritPercent:
		return "SpellCritPercent"
	case BlockPercent:
		return "BlockPercent"
	case RangedHitPercent:
		return "RangedHitPercent"
	case RangedCritPercent:
		return "RangedCritPercent"
	case DodgePercent:
		return "DodgePercent"
	case ParryPercent:
		return "ParryPercent"
	case ReducedCritTakenPercent:
		return "ReducedCritTakenPercent"
	case ExpertisePercent:
		return "ExpertisePercent"
	case DefenseRating:
		return "DefenseRating"
	case BlockRating:
		return "BlockRating"
	case BlockValue:
		return "BlockValue"
	case ArcaneResistance:
		return "ArcaneResistance"
	case FireResistance:
		return "FireResistance"
	case FrostResistance:
		return "FrostResistance"
	case NatureResistance:
		return "NatureResistance"
	case ShadowResistance:
		return "ShadowResistance"
	case PhysicalDamage:
		return "PhysicalDamage"
	}

	return "none"
}

func FromProtoArray(values []float64) Stats {
	// SimStatsLen can be larger than ProtoStatsLen, but the built-in copy
	// function will only import the shared indices between the two.
	var stats Stats
	copy(stats[:], values)
	return stats
}

// Runs FromProtoArray() on the stats array embedded in the UnitStats message, but additionally imports any
// PseudoStats that we want to model as proper Stats in the back-end. This allows us to include only essential
// basic properties in database stats arrays, while still letting the back-end Stat enum include derived
// properties when it is computationally convenient to do so (such as for automatically applying stat
// dependencies). Make sure to update this function if you add any back-end Stat entries that are modeled as
// PseudoStats in the front-end.
func FromUnitStatsProto(unitStatsMessage *proto.UnitStats) Stats {
	return FromProtoArray(unitStatsMessage.Stats).Add(FromPseudoStatsProto(unitStatsMessage.PseudoStats))
}

// The percent PseudoStats an item or enchant states, as the back-end Stats that model them. Ranged hit and
// crit are totals that include the melee share. Dodge, parry and block are read as chance added to the
// base; GetPseudoStatsProto writes them as the total chance.
func FromPseudoStatsProto(pseudoStats []float64) Stats {
	var simStats Stats
	for _, pair := range PercentPseudoStats {
		simStats[pair.Stat] = PseudoStatValue(pseudoStats, pair.PseudoStat)
		if pair.MeleeShare != nil {
			simStats[pair.Stat] -= PseudoStatValue(pseudoStats, pair.MeleeShare.PseudoStat)
		}
	}
	return simStats
}

// The weights a UnitStats message carries, per point of whatever each stat counts: a percent's weight
// is copied as it is, so Block% stays per percent and ranged hit and crit keep their own weight.
func WeightsFromUnitStatsProto(weights *proto.UnitStats) Stats {
	simStats := FromProtoArray(weights.Stats)
	for _, pair := range PercentPseudoStats {
		simStats[pair.Stat] = PseudoStatValue(weights.PseudoStats, pair.PseudoStat)
	}
	return simStats
}

// A percent the character sheet shows as a PseudoStat and the back end models as a Stat. A ranged total
// on the sheet includes its MeleeShare, which the ranged Stat leaves out.
type PercentPseudoStat struct {
	Stat       Stat
	PseudoStat proto.PseudoStat
	MeleeShare *PercentPseudoStat
}

// The hit and crit percents, named so each ranged one can point at its melee share.
var (
	meleeHitPseudoStat   = PercentPseudoStat{PhysicalHitPercent, proto.PseudoStat_PseudoStatMeleeHitPercent, nil}
	rangedHitPseudoStat  = PercentPseudoStat{RangedHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent, &meleeHitPseudoStat}
	spellHitPseudoStat   = PercentPseudoStat{SpellHitPercent, proto.PseudoStat_PseudoStatSpellHitPercent, nil}
	meleeCritPseudoStat  = PercentPseudoStat{PhysicalCritPercent, proto.PseudoStat_PseudoStatMeleeCritPercent, nil}
	rangedCritPseudoStat = PercentPseudoStat{RangedCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent, &meleeCritPseudoStat}
	spellCritPseudoStat  = PercentPseudoStat{SpellCritPercent, proto.PseudoStat_PseudoStatSpellCritPercent, nil}
)

var PercentPseudoStats = []PercentPseudoStat{
	meleeHitPseudoStat,
	spellHitPseudoStat,
	meleeCritPseudoStat,
	spellCritPseudoStat,
	{BlockPercent, proto.PseudoStat_PseudoStatBlockPercent, nil},
	rangedHitPseudoStat,
	rangedCritPseudoStat,
	{DodgePercent, proto.PseudoStat_PseudoStatDodgePercent, nil},
	{ParryPercent, proto.PseudoStat_PseudoStatParryPercent, nil},
	{ExpertisePercent, proto.PseudoStat_PseudoStatExpertisePercent, nil},
}

func PseudoStatValue(pseudoStats []float64, pseudoStat proto.PseudoStat) float64 {
	if int(pseudoStat) < len(pseudoStats) {
		return pseudoStats[pseudoStat]
	}
	return 0
}

// Adds two Stats together, returning the new Stats.
func (stats Stats) Add(other Stats) Stats {
	for k := range stats {
		stats[k] += other[k]
	}
	return stats
}

// Adds another to Stats to this, in-place. For performance, only.
func (stats *Stats) AddInplace(other *Stats) {
	for k := range stats {
		stats[k] += other[k]
	}
}

// Subtracts another Stats from this one, returning the new Stats.
func (stats Stats) Subtract(other Stats) Stats {
	for k := range stats {
		stats[k] -= other[k]
	}
	return stats
}

func (stats Stats) Invert() Stats {
	for k, v := range stats {
		stats[k] = -v
	}
	return stats
}

// Rounds all stat values down to the nearest integer, returning the new Stats.
func (stats Stats) Floor() Stats {
	for k, v := range stats {
		stats[k] = math.Floor(v)
	}
	return stats
}

// Stats the game stores as rounded-down integers on the unit: the five
// attributes. The game floors these after resolving percentage multipliers
// (e.g. Blessing of Kings), and derived values consume the floored integers
// (health from floored Stamina, dodge from floored Agility). Verified against
// live character sheets.
//
// Unlike attributes, combat ratings must NOT be floored here: the sim uses
// rating stats as mixed accumulators that include fractional conversions from
// talents and racials (e.g. dodge from Agility stored as DodgeRating), and TBC has
// no rating multipliers, so real rating totals are already integers.
var flooredGameStats = []Stat{
	Strength, Agility, Stamina, Intellect, Spirit,
}

var isFlooredGameStat = func() [SimStatsLen]bool {
	var m [SimStatsLen]bool
	for _, s := range flooredGameStats {
		m[s] = true
	}
	return m
}()

// Rounds attributes down to integers, matching how the game stores them.
// Must be applied after stat dependencies are resolved.
func (stats Stats) FloorGameStats() Stats {
	for _, k := range flooredGameStats {
		stats[k] = math.Floor(stats[k])
	}
	return stats
}

// FloorGameStatsInPlace is FloorGameStats without copying the array.
func (stats *Stats) FloorGameStatsInPlace() {
	for _, k := range flooredGameStats {
		stats[k] = math.Floor(stats[k])
	}
}

func (stats Stats) Multiply(multiplier float64) Stats {
	for k := range stats {
		stats[k] *= multiplier
	}
	return stats
}

// Multiplies two Stats together by multiplying the values of corresponding
// stats, like a dot product operation.
func (stats Stats) DotProduct(other Stats) Stats {
	for k := range stats {
		stats[k] *= other[k]
	}
	return stats
}

// Higher performance variant of the above.
func (stats Stats) ApplyMultipliers(multipliers map[Stat]float64) Stats {
	for k, v := range multipliers {
		stats[k] *= v
	}
	return stats
}

func (stats Stats) Equals(other Stats) bool {
	return stats == other
}

func (stats Stats) EqualsWithTolerance(other Stats, tolerance float64) bool {
	for k, v := range stats {
		if v < other[k]-tolerance || v > other[k]+tolerance {
			return false
		}
	}
	return true
}

// Given an array of Stat types, return the Stat whose value is largest within
// this Stats array.
func (stats Stats) GetHighestStatType(statTypeOptions []Stat) Stat {
	if len(statTypeOptions) < 1 {
		panic("Must supply at least one Stat type option!")
	}

	var highestStatType Stat
	var highestStatValue float64

	for idx, statType := range statTypeOptions {
		if (idx == 0) || (stats[statType] > highestStatValue) {
			highestStatType = statType
			highestStatValue = stats[statType]
		}
	}

	return highestStatType
}

// Returns all Stat types with positive representation in this Stats array.
func (stats Stats) GetBuffedStatTypes() []Stat {
	buffedStatTypes := make([]Stat, 0, SimStatsLen)

	for statIdx, statValue := range stats {
		if statValue > 0 {
			buffedStatTypes = append(buffedStatTypes, Stat(statIdx))
		}
	}

	return buffedStatTypes
}

func (stats Stats) String() string {
	var sb strings.Builder
	sb.WriteString("\n{\n")

	for statIdx, statValue := range stats {
		if statValue == 0 {
			continue
		}
		if name := Stat(statIdx).StatName(); name != "none" {
			_, _ = fmt.Fprintf(&sb, "\t%s: %0.3f,\n", name, statValue)
		}
	}

	sb.WriteString("\n}")
	return sb.String()
}

// Like String() but without the newlines.
func (stats Stats) FlatString() string {
	var sb strings.Builder
	sb.WriteString("{")

	for statIdx, statValue := range stats {
		if statValue == 0 {
			continue
		}
		if name := Stat(statIdx).StatName(); name != "none" {
			_, _ = fmt.Fprintf(&sb, "\"%s\": %0.3f,", name, statValue)
		}
	}

	sb.WriteString("}")
	return sb.String()
}

func (stats Stats) ToProtoArray() []float64 {
	// SimStatsLen can be larger than ProtoStatsLen, so export only the
	// shared indices between the two.
	return stats[:ProtoStatsLen]
}
func (stats Stats) ToProtoMap() map[int32]float64 {
	m := make(map[int32]float64, ProtoStatsLen)
	for i := 0; i < int(ProtoStatsLen); i++ {
		if stats[i] != 0 {
			m[int32(i)] = stats[i]
		}
	}
	return m
}

func FromProtoMap(m map[int32]float64) Stats {
	var stats Stats
	for k, v := range m {
		stats[k] = v
	}
	return stats
}

type PseudoStats struct {
	///////////////////////////////////////////////////
	// Effects that apply when this unit is the attacker.
	///////////////////////////////////////////////////

	SpellCostPercentModifier int32 // Multiplies spell cost.

	CastSpeedMultiplier   float64
	MeleeSpeedMultiplier  float64
	RangedSpeedMultiplier float64
	RangedHasteMultiplier float64
	AttackSpeedMultiplier float64 // Used for real haste effects that modify resoruce regen and are used for RPPM effects
	WeaponSpeedMultiplier float64 // Seal of the Crusader's faster, lighter swings, which a rate on the weapon's speed divides by

	FiveSecondRuleRefreshTime time.Duration // last time a spell was cast
	SpiritRegenRateCasting    float64       // percentage of spirit regen allowed during casting

	SpiritRegenMultiplier float64 // Multiplier on spirit portion of mana regen.

	// If true, allows block/parry.
	InFrontOfTarget bool

	BonusMHDps     float64
	BonusOHDps     float64
	BonusRangedDps float64

	DisableDWMissPenalty bool // Used by Heroic Strike and Cleave

	IncreasedMissChance float64 // Insect Swarm and Scorpid Sting
	DodgeReduction      float64 // Used by Warrior talent 'Weapon Mastery' and SWP boss auras.

	WeaponSkills       [proto.WeaponSkill_WeaponSkillWands + 1]float64 // Bonus skill over the level-based maximum per weapon skill, which a skill below it makes negative.
	FeralCombatSkill   float64
	FeralCombatEnabled bool // While set, FeralCombatSkill replaces the skill of the equipped weapon.

	ThreatMultiplier       float64            // Modulates the threat generated. Affected by things like salv.
	SchoolThreatMultiplier [SchoolLen]float64 // The threat of one spell school, on top of ThreatMultiplier.

	DamageDealtMultiplier          float64            // All damage
	SchoolDamageDealtMultiplier    [SchoolLen]float64 // For specific spell schools (arcane, fire, shadow, etc).
	SchoolSpellPiercing            [SchoolLen]float64 // Target resistance ignored by spells of one school, on top of SpellPiercing.
	DotDamageMultiplierAdditive    float64            // All periodic damage
	HealingDealtMultiplier         float64            // All non-shield healing
	PeriodicHealingDealtMultiplier float64            // All periodic healing (on top of HealingDealtMultiplier)
	CritDamageMultiplier           float64            // All multiplicative crit damage

	BonusRangedAttackPower float64 // Hunter's mark
	BonusAttackPower       float64 // Also Hunter's Mark

	// Important when unit is attacker or target
	BlockValueMultiplier  float64
	BlockValuePerStrength float64

	// Only used for NPCs, governs variance in enemy auto-attack damage
	DamageSpread float64

	///////////////////////////////////////////////////
	// Crowd control effects. See sim/core/incapacitate.go.
	///////////////////////////////////////////////////

	Incapacitated bool // Under crowd control (e.g. a boss's Fear); cannot cast or queue any spell. Procs still fire.
	Stunned       bool // Prevents blocks, dodges, and parries. Read while this unit is the target.
	FearImmune    bool // Fear effects cannot be applied to this unit. (ie. Berserker Rage)
	StunImmune    bool // Stun effects cannot be applied to this unit.

	// Scale how long a Fear/Stun lasts on this unit, e.g. the Warrior talent Iron Will.
	FearDurationMultiplier float64
	StunDurationMultiplier float64

	///////////////////////////////////////////////////
	// Effects that apply when this unit is the target.
	///////////////////////////////////////////////////

	CanBlock bool
	CanParry bool
	CanCrush bool

	ParryHaste bool

	// Avoidance % not affected by Diminishing Returns, represented as
	// probabilities (between 0 and 1).
	BaseDodgeChance float64
	BaseParryChance float64
	BaseBlockChance float64

	BaseReducedCritTakenPercent float64 // Base crit reduction from talents/auras (before the Defense contribution).
	ReducedCritTakenPercent     float64 // Total crit reduction including the Defense contribution.

	BonusHealingTaken          float64 // Talisman of Troll Divinity
	BonusSpellCritPercentTaken float64 // Imp Shadow Bolt / Imp Scorch / Winter's Chill debuff
	BonusPhysicalDamageTaken   float64 // Hemo, Gift of Arthas, etc
	BonusSpellDamageTaken      float64 // Amp Magic
	BonusRangedDamageTaken     float64 // Physical ranged attacks only: Windwall

	DamageTakenMultiplier       float64            // All damage
	SchoolDamageTakenMultiplier [SchoolLen]float64 // For specific spell schools (arcane, fire, shadow, etc.)
	SchoolBonusSpellDamage      [SchoolLen]float64 // Flat school damage the target takes extra, ex: Judgement of the Crusader; added at the spell's coefficient after the caster's done multipliers, before the outcome
	SchoolBonusHitChance        [SchoolLen]float64 // The bonus stats' school hit, on every spell of the school with a ClassSpellMask or class flags; a talent's school hit is a class-flags mod with ShowAsSchoolHit

	DiseaseDamageTakenMultiplier          float64
	PeriodicPhysicalDamageTakenMultiplier float64
	MeleeDamageTakenMultiplier            float64 // Melee attacks of any school

	ArmorMultiplier float64 // Major/minor/special multiplicative armor modifiers

	ReducedPhysicalHitTakenChance float64
	ReducedArcaneHitTakenChance   float64
	ReducedFireHitTakenChance     float64
	ReducedFrostHitTakenChance    float64
	ReducedNatureHitTakenChance   float64
	ReducedShadowHitTakenChance   float64

	HealingTakenMultiplier         float64 // All healing sources including self-healing
	ExternalHealingTakenMultiplier float64 // Modulates the output of the individual tank sim healing model
	MovementSpeedMultiplier        float64 // Multiplier for movement speed, default to 1. Player base movement 7 yards/s. All effects affecting movements are multipliers.
	SelfHealingMultiplier          float64 // Healing from spells and abilities, only-self
	PushbackChance                 float64 // Chance of being pushed back by a spell cast, defaults to 1.
}

func NewPseudoStats() PseudoStats {
	return PseudoStats{
		SpellCostPercentModifier: 100,

		CastSpeedMultiplier:   1,
		MeleeSpeedMultiplier:  1,
		RangedSpeedMultiplier: 1,
		RangedHasteMultiplier: 1,
		AttackSpeedMultiplier: 1,
		WeaponSpeedMultiplier: 1,
		SpiritRegenMultiplier: 1,

		ThreatMultiplier:       1,
		SchoolThreatMultiplier: NewSchoolFloatArray(),

		DamageDealtMultiplier:          1,
		SchoolDamageDealtMultiplier:    NewSchoolFloatArray(),
		DotDamageMultiplierAdditive:    1,
		HealingDealtMultiplier:         1,
		PeriodicHealingDealtMultiplier: 1,
		CritDamageMultiplier:           1,

		BlockValueMultiplier:        1,
		BaseReducedCritTakenPercent: 0,
		ReducedCritTakenPercent:     0,

		DamageSpread: 0.3333,

		// Target effects.
		DamageTakenMultiplier:       1,
		SchoolDamageTakenMultiplier: NewSchoolFloatArray(),

		DiseaseDamageTakenMultiplier:          1,
		PeriodicPhysicalDamageTakenMultiplier: 1,
		MeleeDamageTakenMultiplier:            1,

		ArmorMultiplier: 1,

		HealingTakenMultiplier:         1,
		ExternalHealingTakenMultiplier: 1,
		MovementSpeedMultiplier:        1,
		PushbackChance:                 1,
		FearDurationMultiplier:         1,
		StunDurationMultiplier:         1,
	}
}

type UnitStat int

func (s UnitStat) IsStat() bool               { return int(s) < int(ProtoStatsLen) }
func (s UnitStat) IsPseudoStat() bool         { return !s.IsStat() }
func (s UnitStat) EqualsStat(other Stat) bool { return s.IsStat() && (s.StatIdx() == int(other)) }
func (s UnitStat) EqualsPseudoStat(other proto.PseudoStat) bool {
	return s.IsPseudoStat() && (s.PseudoStatIdx() == int(other))
}
func (s UnitStat) StatIdx() int {
	if !s.IsStat() {
		panic("Is a pseudo stat")
	}
	return int(s)
}
func (s UnitStat) PseudoStatIdx() int {
	if s.IsStat() {
		panic("Is a regular stat")
	}
	return int(s) - int(ProtoStatsLen)
}
func (s UnitStat) AddToStatsProto(p *proto.UnitStats, value float64) {
	if s.IsStat() {
		p.Stats[s.StatIdx()] += value
	} else {
		p.PseudoStats[s.PseudoStatIdx()] += value
	}
}

func UnitStatFromIdx(s int) UnitStat   { return UnitStat(s) }
func UnitStatFromStat(s Stat) UnitStat { return UnitStat(s) }
func UnitStatFromPseudoStat(s proto.PseudoStat) UnitStat {
	return UnitStat(int(s) + int(ProtoStatsLen))
}

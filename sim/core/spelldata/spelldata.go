// Package spelldata is the game files' spell rows as the sim reads them: one Spell per spell ID, with
// its effects and power costs, and accessors that turn the game files' units into core's.
//
// Values are stored in the units the game files state them in - a percentage is the integer 16,
// rage is on a 0-1000 bar, times are milliseconds - and the conversion happens in the accessors, so
// a row always matches what the DBC says.
//
// Every accessor answers on the zero value, so Nil and NilEffect are safe to chain: an unknown spell
// reads as zeroes rather than crashing the registration that asked for it.
//
// sim/core must never import this package: tools/database imports core, and a cycle there would stop
// the generator that writes this package's data.
package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// One spell row. Every field is the game files' column, in the game files' units.
type Spell struct {
	ID   int32
	Name string

	// Spell.NameSubtext_lang: "Rank 4" on a ranked spell, empty on everything else.
	Rank string

	// SpellMisc.SchoolMask, which core.SpellSchool states in the same bits.
	School core.SpellSchool

	// SpellMisc.Speed: how fast the missile flies, in yards per second, or its flight time in seconds
	// where the row states Fixed Travel Time (FixedTravelTime). Zero, with no MinDuration, hits on cast.
	Speed float32

	// SpellMisc.MinDuration: the shortest the missile flies, in seconds, however close the target.
	MinDuration float32

	// SpellMisc.Attributes_0..16, read through HasAttr and the named accessors in attributes.go.
	Attr [17]uint32

	// SpellLevels: the level the rank is taught at, the level its per-level scaling counts from, and
	// the level that scaling stops at. A spell with no SpellLevels row carries SpellLevel and BaseLevel
	// 60 and MaxLevel 0, which is what the generator writes for "no level scaling".
	SpellLevel, BaseLevel, MaxLevel int16

	// SpellCastTimes.Base through SpellMisc.CastingTimeIndex.
	CastTimeMs int32

	// SpellDuration.Duration through SpellMisc.DurationIndex. -1 is the game files' permanent aura and
	// is kept as it is; Duration() turns it into core.NeverExpires.
	DurationMs int32

	// SpellRange.RangeMin_0 / RangeMax_0. A nonzero minimum is the dead zone on a charge.
	MinRange, MaxRange float32

	// SpellCooldowns.RecoveryTime, CategoryRecoveryTime and StartRecoveryTime. The category cooldown
	// is the shared one Fire Blast and Cone of Cold run off.
	CooldownMs, CategoryCooldownMs, GCDMs int32

	// SpellCategories.Category, StartRecoveryCategory and ChargeCategory.
	Category, StartRecoveryCategory, ChargeCategory int16

	// SpellCategories.DefenseType, DispelType, Mechanic and PreventionType. The game files count
	// defense types none, magic, melee, ranged, which is the order core.DefenseType is declared in.
	DefenseType    core.DefenseType
	DispelType     uint8
	Mechanic       dbcenums.Mechanic
	PreventionType uint8

	// SpellAuraOptions.CumulativeAura: how high the aura stacks. Zero is one application.
	MaxStack int16

	// SpellAuraOptions.ProcChance as the game files state it, which is not always a roll: 100 and 101
	// both read as "fires on its own condition". It is the chance only where ProcChanceSource is
	// ProcChanceColumn; under any other source the column means nothing and the source says where
	// the rate is.
	ProcChance uint8

	// SpellAuraOptions.ProcCharges: how many times the aura acts before it drops. Zero is unlimited.
	ProcCharges int16

	// SpellAuraOptions.ProcTypeMask_0/_1, decoded by the dbcenums.PROC_FLAG_ names.
	ProcFlags [2]uint32

	// SpellAuraOptions.ProcCategoryRecovery: the internal cooldown between two procs, in ms.
	ICDMs int32

	// Procs per minute. The game files carry none - SpellProcsPerMinuteID is 0 on every row - so
	// this is zero unless an override fills it.
	RPPM float32

	// The threat an ability adds beyond its damage, where the tooltip states threat without a
	// number: "causes a high amount of threat". The amounts the game files do state are an E_THREAT
	// effect, so this is zero unless an override fills it.
	FlatThreat float32

	// An area bonus from an override: AreaGroup ids (any counts), amount and duration factors.
	AreaBonusGroups                        []int32
	AreaMultiplier, AreaDurationMultiplier float32

	// SpellClassOptions.SpellClassSet and SpellClassMask_0..3: the set of spells a talent effect
	// naming this family reaches.
	ClassFlags core.ClassFlags

	// SpellInterrupts.InterruptFlags, read through PushedBack.
	InterruptFlags uint32

	// SpellInterrupts.AuraInterruptFlags_0/_1 and ChannelInterruptFlags_0/_1.
	AuraInterrupt, ChannelInterrupt [2]uint32

	// SpellShapeshift.ShapeshiftMask_0 | ShapeshiftMask_1 << 32: the forms the spell is usable in.
	StanceMask uint64

	// SpellShapeshift.ShapeshiftExclude_0 | ShapeshiftExclude_1 << 32: the forms the spell is barred from.
	StanceExclude uint64

	// SpellAuraRestrictions.CasterAuraSpell: the spell id whose aura the caster must carry to cast this one.
	CasterAura int32

	// SpellAuraRestrictions.ExcludeCasterAuraSpell: the spell id whose aura bars the caster from casting this one.
	ExcludeCasterAura int32

	// SpellTargetRestrictions.MaxTargets for an area effect. Zero is unlimited.
	MaxTargets int16

	// SpellTargetRestrictions.TargetCreatureType: the creature types the spell can hit, as a mask of
	// 1 << (CreatureType - 1), so 4 is demons and 256 mechanicals. Zero is any.
	TargetCreatureType int32

	// SpellCastingRequirements.RequiredAreasID. Zero is anywhere.
	RequiredAreas int32

	// SpellEquippedItems: the item class, subclass mask and inventory-type mask the spell requires,
	// which is how a weapon-specific proc states the weapons it fires from.
	EquipClass                  int8
	EquipSubclass, EquipInvType int32

	// SpellLabel.LabelID: the labels a label-keyed modifier aura addresses this spell through.
	Labels []int16

	// The spell IDs the tooltip references ($12966n, $26364s1), in the order they appear. Refs()
	// resolves them.
	RefIDs []int32

	// How the tooltip words the proc, baked in at generation: it is what says whether ProcChance is a
	// roll at all.
	ProcChanceSource ProcChanceSource

	// The effect whose value is the roll, when ProcChanceSource is ProcChanceEffectN, as the
	// position EffectN counts by.
	ProcChanceEffect int8

	// What the tooltip states about the trigger that the proc mask cannot, baked in at generation.
	ProcHint core.ProcHint

	// Whether one roll of the damage is divided evenly among the targets hit, baked in at generation.
	// No column states it; the tooltip reads "split between", "divided up evenly" or "$s1 total ...
	// damage".
	SplitsDamage bool

	Effects  []Effect
	Powers   []Power
	Reagents []Reagent
}

// Where the proc's chance is stated, since the ProcChance column alone cannot tell a roll from a
// condition.
type ProcChanceSource uint8

const (
	// ProcChance is the roll: the tooltip renders it as "$h%".
	ProcChanceColumn ProcChanceSource = iota

	// The value of effect ProcChanceEffect is the roll: the tooltip renders it as "$mN%".
	ProcChanceEffectN

	// The column reads 100 or 101 and the tooltip's trigger clause states no chance at all, so the
	// aura fires whenever its condition is met.
	ProcChanceAlways

	// The game files state no chance anywhere, or state 100 or 101 beside a trigger clause saying the
	// effect only sometimes happens, so the rate has to come from an override into RPPM.
	ProcChancePPM
)

// One SpellEffect row.
type Effect struct {
	ID, SpellID int32

	// SpellEffect.EffectIndex, the game files' own numbering. EffectN indexes by position instead.
	Index uint8

	Type dbcenums.SpellEffectType
	Aura dbcenums.EffectAuraType

	// EffectBasePointsF as the game files state it, which is a percentage as an integer and rage on a
	// 0-1000 bar. Average() adds the per-level scaling on top.
	BasePoints float64

	// EffectRealPointsPerLevel: what the effect gains for each level between the spell's BaseLevel and
	// the caster's.
	PPL float64

	// The owning spell's SpellLevels, copied onto the effect at generation: Average() prices the
	// per-level gain off them, and reading them here rather than looking the owner up in the store
	// keeps the answer the same for an effect whose spell the store does not carry.
	BaseLevel, MaxLevel int16

	// EffectVariance, the full width of the roll: a spell-school hit or a heal lands within
	// Variance/2 of its whole amount, power included (DirectVariance); a physical hit rolls
	// |BasePoints| * Variance either side of the base points and any other effect half that, plus
	// the per-level gain (Min, Max). Zero on an effect that does not roll.
	Variance float64

	// EffectBonusCoefficient and BonusCoefficientFromAP: the share of spell power and attack power the
	// effect adds.
	SPCoef float64
	APCoef float64

	// EffectPvpMultiplier.
	PvpMult float32

	// EffectAmplitude and EffectAuraPeriod, the tick interval in ms.
	Amplitude float32
	PeriodMs  int32

	// SpellRadius through EffectRadiusIndex_0.
	RadiusMin, RadiusMax float32

	// EffectMiscValue_0/_1. What they mean is the aura's business: a stat for A_MOD_STAT, a stat
	// mask in the second for A_MOD_TOTAL_STAT_PERCENTAGE (TotalStatPercentStats reads it), a school
	// mask for A_MOD_DAMAGE_DONE, a modifier op for A_ADD_PCT_MODIFIER.
	Misc, Misc2 int32

	// EffectSpellClassMask_0..3 with the owning spell's family: the spells this effect modifies.
	ClassFlags core.ClassFlags

	// EffectTriggerSpell: the spell this effect fires. Trigger() resolves it.
	TriggerID int32

	// EffectChainTargets and EffectChainAmplitude: how far a chain jumps and what each jump keeps.
	ChainTargets int16
	ChainAmp     float32

	// EffectMechanic.
	Mechanic dbcenums.Mechanic

	// EffectPointsPerResource: what the amount gains per point of the resource spent.
	PointsPerResource float32

	// ImplicitTarget_0/_1.
	Target [2]dbcenums.ImplicitTarget

	// EffectAttributes.
	Attributes int32
}

// One SpellPower row: what the spell costs and out of which bar.
type Power struct {
	// SpellPower.PowerType. Rage is on a 0-1000 bar, so PowerCost divides it; the rest are stated in
	// whole points.
	Type dbcenums.PowerType

	// ManaCost and ManaCostPerLevel, in the bar's own units.
	Cost, CostPerLevel int32

	// PowerCostPct: the share of the caster's base the cast takes instead of a flat amount.
	CostPct float32

	// ManaPerSecond, for a channel that drains while it runs.
	PerSecond int32
}

// One filled slot of the SpellReagents row: an item the cast consumes, and how many.
type Reagent struct {
	Item  int32
	Count int16
}

// What an unknown spell and an out-of-range effect answer with: every accessor reads zero off them,
// so a caller can chain through a spell the store does not carry. Both are shared by every caller
// that lands on them, so neither may be written through.
var Nil = &Spell{}
var NilEffect = &Effect{}

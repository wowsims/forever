package reforgeoptimizer

import (
	"math"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Reads weights, caps and sheet stats alike. The percent pseudo stats stay in PseudoStats in the
// sheet's percent, the unit GetPseudoStatsProto writes: a cap is compared there and a weight is
// read there, so a Block% weight or cap stays per percent and ranged hit and crit stay totals.
func protoToCoreUnitStats(protoStats *proto.UnitStats) core.UnitStats {
	if protoStats == nil {
		return core.NewUnitStats()
	}
	return core.UnitStats{
		Stats:       stats.FromProtoArray(protoStats.Stats),
		PseudoStats: slices.Clone(protoStats.PseudoStats),
	}
}

func getUnitStat(unitStats core.UnitStats, unitStat stats.UnitStat) float64 {
	if unitStat.IsStat() {
		return unitStats.Stats[unitStat.StatIdx()]
	}
	pseudoStatIdx := int(unitStat.PseudoStatIdx())
	if pseudoStatIdx >= len(unitStats.PseudoStats) {
		return 0
	}
	return unitStats.PseudoStats[pseudoStatIdx]
}

func setUnitStat(unitStats core.UnitStats, unitStat stats.UnitStat, value float64) core.UnitStats {
	writeUnitStat(&unitStats, unitStat, value)
	return unitStats
}

// Sets the value in place, so a hot loop does not copy the Stats array on every write.
func writeUnitStat(unitStats *core.UnitStats, unitStat stats.UnitStat, value float64) {
	if unitStat.IsStat() {
		unitStats.Stats[unitStat.StatIdx()] = value
		return
	}
	pseudoStatIdx := int(unitStat.PseudoStatIdx())
	for len(unitStats.PseudoStats) <= pseudoStatIdx {
		unitStats.PseudoStats = append(unitStats.PseudoStats, 0)
	}
	unitStats.PseudoStats[pseudoStatIdx] = value
}

func isEmptyUnitStats(unitStats core.UnitStats) bool {
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		if unitStats.Stats[statIdx] != 0 {
			return false
		}
	}
	for _, value := range unitStats.PseudoStats {
		if value != 0 {
			return false
		}
	}
	return true
}

// Maps each haste% pseudo-stat to its speed multiplier pseudo-stat for the analytical haste delta
// calculation.
// Δhaste% = speedMult × ΔHasteRating / HasteRatingPerHastePercent
var hasteSpeedMultipliers = map[proto.PseudoStat]proto.PseudoStat{
	proto.PseudoStat_PseudoStatMeleeHastePercent:  proto.PseudoStat_PseudoStatMeleeSpeedMultiplier,
	proto.PseudoStat_PseudoStatRangedHastePercent: proto.PseudoStat_PseudoStatRangedSpeedMultiplier,
	proto.PseudoStat_PseudoStatSpellHastePercent:  proto.PseudoStat_PseudoStatCastSpeedMultiplier,
}

// Applies the character's stat dependency graph to delta, resolving conversions such as
// HitRating→Hit%, CritRating→Crit%, Agility→PhysicalCritPercent.
// It also mirrors the resolved Stats values back to their corresponding PseudoStats
// so that LP constraint evaluation (which reads PseudoStats for hit/crit/haste caps)
// sees the correct contribution. A rating counted in whole steps moves its percent by
// the steps it adds on top of baseStats.
//
// Haste% is multiplicative with a speed multiplier that is not captured by the dep
// manager. We read it from baseStats.PseudoStats (populated by GetPseudoStatsProto):
//
//	Δhaste% = speedMult × ΔHasteRating / HasteRatingPerHastePercent
func resolveStatDelta(sdm *stats.StatDependencyManager, baseStats core.UnitStats, delta core.UnitStats) core.UnitStats {
	if isEmptyUnitStats(delta) {
		return delta
	}
	delta.Stats = sdm.ApplyStatDependencies(delta.Stats)
	for _, conversion := range core.RatingConversions {
		rating := delta.Stats[conversion.Rating]
		if conversion.Step == 0 || rating == 0 {
			continue
		}
		base := baseStats.Stats[conversion.Rating]
		counted := math.Floor(rating / conversion.Step)
		whole := math.Floor((base+rating)/conversion.Step) - math.Floor(base/conversion.Step)
		delta.Stats[conversion.Percent] += (whole - counted) * conversion.Step / conversion.RatingPerPercent
	}

	// Mirror dual-stored stats from Stats (updated by SDM — e.g. HitRating→Hit%,
	// CritRating→Crit%, Agility→PhysicalCritPercent) back to their PseudoStat indices.
	for _, pair := range stats.PercentPseudoStats {
		value := delta.Stats[pair.Stat]
		if pair.MeleeShare != nil {
			value = delta.Stats[pair.MeleeShare.Stat] + value
		}
		writeUnitStat(&delta, stats.UnitStatFromPseudoStat(pair.PseudoStat), value)
	}
	for percent, offHand := range offHandPercents {
		writeUnitStat(&delta, stats.UnitStatFromPseudoStat(offHand), delta.Stats[percent])
	}
	spellHitDelta := delta.Stats[stats.SpellHitPercent]
	for _, schoolHitPS := range schoolHitPercents {
		writeUnitStat(&delta, stats.UnitStatFromPseudoStat(schoolHitPS), spellHitDelta)
	}
	writeUnitStat(&delta, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatReducedCritTakenPercent),
		delta.Stats[stats.ReducedCritTakenPercent])

	// Haste% pseudo-stats: read speed multipliers from baseStats.PseudoStats, which
	// GetPseudoStatsProto populates as MeleeSpeedMultiplier×AttackSpeedMultiplier etc.
	if hasteRatingDelta := delta.Stats[stats.HasteRating]; hasteRatingDelta != 0 {
		for hastePS, speedPS := range hasteSpeedMultipliers {
			speedMult := getUnitStat(baseStats, stats.UnitStatFromPseudoStat(speedPS))
			writeUnitStat(&delta, stats.UnitStatFromPseudoStat(hastePS), speedMult*hasteRatingDelta/core.HasteRatingPerHastePercent)
		}
	}

	return delta
}

// Invokes fn for every stat and pseudo-stat index in the vector.
func eachUnitStat(vec core.UnitStats, fn func(unitStat stats.UnitStat, value float64)) {
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		fn(stats.UnitStatFromStat(stats.Stat(statIdx)), vec.Stats[statIdx])
	}
	for pseudoStatIdx := 0; pseudoStatIdx < int(stats.PseudoStatsLen); pseudoStatIdx++ {
		value := 0.0
		if pseudoStatIdx < len(vec.PseudoStats) {
			value = vec.PseudoStats[pseudoStatIdx]
		}
		fn(stats.UnitStatFromPseudoStat(proto.PseudoStat(pseudoStatIdx)), value)
	}
}

var schoolHitPercents = core.SchoolHitPseudoStats[stats.SchoolIndexArcane:]

// The off hand's percent moves with the physical percent. The per-attack ratings on the sheet are left
// out: a rating's children are counted in percent.
var offHandPercents = map[stats.Stat]proto.PseudoStat{
	stats.PhysicalHitPercent:  proto.PseudoStat_PseudoStatOffHandHitPercent,
	stats.PhysicalCritPercent: proto.PseudoStat_PseudoStatOffHandCritPercent,
}

// The pseudo-stats the character sheet shows a back-end percent stat in: its own, the ranged total
// it is the melee share of, the off hand's for physical hit and crit, and for spell hit the
// per-school percentages.
func sheetPseudoStats(percent stats.Stat) []proto.PseudoStat {
	if percent == stats.ReducedCritTakenPercent {
		return []proto.PseudoStat{proto.PseudoStat_PseudoStatReducedCritTakenPercent}
	}
	var pseudoStats []proto.PseudoStat
	for _, pair := range stats.PercentPseudoStats {
		if pair.Stat == percent || (pair.MeleeShare != nil && pair.MeleeShare.Stat == percent) {
			pseudoStats = append(pseudoStats, pair.PseudoStat)
		}
	}
	if offHand, ok := offHandPercents[percent]; ok {
		pseudoStats = append(pseudoStats, offHand)
	}
	if percent == stats.SpellHitPercent {
		pseudoStats = append(pseudoStats, schoolHitPercents...)
	}
	return pseudoStats
}

// One sheet percent a rating moves through a core rating conversion.
type ratingPercent struct {
	rating           stats.Stat
	percent          proto.PseudoStat
	ratingPerPercent float64
	root             bool // The rating is the percent's root stat, as UnitStat.getRootStat gives it.
}

// Every core rating conversion's sheet percents, in the order core.RatingConversions lists them.
var ratingPercents = buildRatingPercents()

// Defense is no percent's root: dodge, parry and block have their own ratings, and reduced crit
// taken has none. The per-school hit percents have no root either.
func buildRatingPercents() []ratingPercent {
	var lookup []ratingPercent
	for _, conversion := range core.RatingConversions {
		for _, percent := range sheetPseudoStats(conversion.Percent) {
			lookup = append(lookup, ratingPercent{
				rating:           conversion.Rating,
				percent:          percent,
				ratingPerPercent: conversion.RatingPerPercent,
				root:             conversion.Rating != stats.DefenseRating && !slices.Contains(schoolHitPercents, percent),
			})
		}
	}
	return lookup
}

// The percent pseudo-stats a rating moves, from the core rating conversions in the order
// core.RatingConversions lists them, then the channels without a back-end percent Stat.
func childPseudoStats(parent stats.Stat) []proto.PseudoStat {
	var children []proto.PseudoStat
	for _, entry := range ratingPercents {
		if entry.rating == parent {
			children = append(children, entry.percent)
		}
	}
	if channels, ok := core.RatingChannelsOf(parent); ok && !channels.HasPercentStats {
		children = append(children, channels.Channels()...)
	}
	return children
}

func ratingPerPseudoStatPercent(pseudoStat proto.PseudoStat, parent stats.Stat) float64 {
	for _, entry := range ratingPercents {
		if entry.rating == parent && entry.percent == pseudoStat {
			return entry.ratingPerPercent
		}
	}
	if channels, ok := core.RatingChannelsOf(parent); ok && slices.Contains(channels.Channels(), pseudoStat) {
		return channels.RatingPerPercent
	}
	return 1
}

// The remaining room from the current sheet value to cap. A gap of exactly 0 is returned as 1e-12
// so the cap still registers as configured (a zero entry means "no cap").
func computeGapToCap(baseStats core.UnitStats, unitStat stats.UnitStat, cap float64) float64 {
	statDelta := cap - getUnitStat(baseStats, unitStat)
	if statDelta == 0 {
		return 1e-12
	}
	return statDelta
}

// The per-unit-stat gap-to-cap, but only for caps whose configured value is > 0 (a cap of 0 maps
// to 0 and means "no cap").
func computeStatCapsDelta(baseStats core.UnitStats, statCaps core.UnitStats) core.UnitStats {
	result := core.NewUnitStats()
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		if cap := statCaps.Stats[statIdx]; cap > 0 {
			unitStat := stats.UnitStatFromStat(stats.Stat(statIdx))
			result = setUnitStat(result, unitStat, computeGapToCap(baseStats, unitStat, cap))
		}
	}
	for pseudoStatIdx := 0; pseudoStatIdx < int(stats.PseudoStatsLen); pseudoStatIdx++ {
		cap := 0.0
		if pseudoStatIdx < len(statCaps.PseudoStats) {
			cap = statCaps.PseudoStats[pseudoStatIdx]
		}
		if cap > 0 {
			unitStat := stats.UnitStatFromPseudoStat(proto.PseudoStat(pseudoStatIdx))
			result = setUnitStat(result, unitStat, computeGapToCap(baseStats, unitStat, cap))
		}
	}
	return result
}

func unitStatFromUIStat(uiStat *proto.UIStat) (stats.UnitStat, bool) {
	if uiStat == nil {
		return 0, false
	}
	switch unitStat := uiStat.UnitStat.(type) {
	case *proto.UIStat_Stat:
		return stats.UnitStatFromStat(stats.Stat(unitStat.Stat)), true
	case *proto.UIStat_PseudoStat:
		return stats.UnitStatFromPseudoStat(unitStat.PseudoStat), true
	default:
		return 0, false
	}
}

// ---------------------------------------------------------------------------
// LP coefficient-key scheme
// ---------------------------------------------------------------------------
//
// Coefficients are keyed by the proto enum NAME (e.g. "StatHitRating",
// "PseudoStatSpellHitPercent", "ItemSlotHead") so checkCaps can recover the stat from a key.
// Slot keys ("ItemSlot...") and the "score" key are not stat names and parse to (_, false).

func statCoeffKey(stat proto.Stat) string {
	return proto.Stat_name[int32(stat)]
}

func pseudoStatCoeffKey(pseudoStat proto.PseudoStat) string {
	return proto.PseudoStat_name[int32(pseudoStat)]
}

// Recovers the stat from a coefficient key, or (_, false) for non-stat keys. PseudoStat and Stat
// names occupy disjoint namespaces so lookup order is irrelevant.
func unitStatFromCoeffKey(key string) (stats.UnitStat, bool) {
	if value, ok := proto.PseudoStat_value[key]; ok {
		return stats.UnitStatFromPseudoStat(proto.PseudoStat(value)), true
	}
	if value, ok := proto.Stat_value[key]; ok {
		return stats.UnitStatFromStat(stats.Stat(value)), true
	}
	return 0, false
}

// The coefficient/constraint key for a unit stat: its proto enum name.
func coeffKeyForUnitStat(unitStat stats.UnitStat) string {
	if unitStat.IsStat() {
		return statCoeffKey(proto.Stat(unitStat.StatIdx()))
	}
	return pseudoStatCoeffKey(proto.PseudoStat(unitStat.PseudoStatIdx()))
}

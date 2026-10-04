package reforgeoptimizer

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// reforgeSoftCap represents a soft cap or threshold expressed in reforge-relative (gap-to-cap)
// breakpoints. breakpoints/postCapEPs are mutated in place across solver passes as breakpoints
// are consumed.
type reforgeSoftCap struct {
	unitStat    stats.UnitStat
	breakpoints []float64
	postCapEPs  []float64
	capType     proto.StatCapType
}

// ---------------------------------------------------------------------------
// LP-path cap computation (mirrors the reference solver).
// ---------------------------------------------------------------------------

// computeReforgeSoftCaps converts each configured soft cap's absolute breakpoints into
// gap-to-cap deltas. For TypeThreshold caps the breakpoints are reversed (largest gap evaluated
// first) and every post-cap EP is set to the FIRST configured value, which is interpreted as the
// residual stat value just after passing a threshold discontinuity.
func computeReforgeSoftCaps(baseStats core.UnitStats, configs []*proto.StatCapConfig) []*reforgeSoftCap {
	result := make([]*reforgeSoftCap, 0, len(configs))
	for _, config := range configs {
		unitStat, ok := unitStatFromUIStat(config.GetUnitStat())
		if !ok {
			continue
		}

		weights := slices.Clone(config.GetPostCap_EPs())
		relativeBreakpoints := make([]float64, 0, len(config.GetBreakpoints()))
		for _, breakpoint := range config.GetBreakpoints() {
			relativeBreakpoints = append(relativeBreakpoints, computeGapToCap(baseStats, unitStat, breakpoint))
		}

		if config.GetCapType() == proto.StatCapType_TypeThreshold {
			slices.Reverse(relativeBreakpoints)
			first := 0.0
			if len(weights) > 0 {
				first = weights[0]
			}
			weights = make([]float64, len(relativeBreakpoints))
			for i := range weights {
				weights[i] = first
			}
		}

		result = append(result, &reforgeSoftCap{
			unitStat:    unitStat,
			breakpoints: relativeBreakpoints,
			capType:     config.GetCapType(),
			postCapEPs:  weights,
		})
	}
	return result
}

var offenseRatings = []stats.Stat{stats.HitRating, stats.CritRating, stats.HasteRating, stats.ExpertiseRating}

// avoidanceRatings share percent children: defense moves dodge, parry and block alongside their own
// ratings.
var avoidanceRatings = []stats.Stat{stats.DefenseRating, stats.DodgeRating, stats.ParryRating, stats.BlockRating}

// checkWeights moves each weighted rating's EP onto the first of its percent children that carries
// a cap, as EP per percent, so the solver values the rating in the units its caps use. Without a
// capped child the rating keeps its EP; a rating without EP is valued through its children.
// A hit, crit, haste or expertise rating with EP drops its children's weights first. An avoidance
// rating is zeroed instead when any of its children already carries a weight; defense runs first,
// so its EP landing on Dodge% zeroes dodge rating.
func checkWeights(weights core.UnitStats, reforgeCaps core.UnitStats, reforgeSoftCaps []*reforgeSoftCap) core.UnitStats {
	validated := weights
	for _, parent := range slices.Concat(offenseRatings, avoidanceRatings) {
		if validated.Stats[parent] == 0 {
			continue
		}
		children := childPseudoStats(parent)
		if len(children) == 0 {
			continue
		}

		if slices.Contains(avoidanceRatings, parent) {
			if slices.ContainsFunc(children, func(child proto.PseudoStat) bool {
				return getUnitStat(validated, stats.UnitStatFromPseudoStat(child)) != 0
			}) {
				validated.Stats[parent] = 0
				continue
			}
		} else {
			for _, child := range children {
				validated = setUnitStat(validated, stats.UnitStatFromPseudoStat(child), 0)
			}
		}

		for _, child := range children {
			if !pseudoStatHasCap(child, reforgeCaps, reforgeSoftCaps) {
				continue
			}
			if ratingPerPercent := ratingPerPseudoStatPercent(child, parent); ratingPerPercent != 0 {
				childUnitStat := stats.UnitStatFromPseudoStat(child)
				rescaled := validated.Stats[parent] * ratingPerPercent
				validated = setUnitStat(validated, childUnitStat, getUnitStat(validated, childUnitStat)+rescaled)
				validated.Stats[parent] = 0
			}
			break
		}
	}
	return validated
}

// ---------------------------------------------------------------------------
// Cap-detection helpers. reforgeCaps is a gap-to-cap vector (the remaining rating room before
// each hard cap).
// ---------------------------------------------------------------------------

func pseudoStatHasCap(pseudoStat proto.PseudoStat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
	if getUnitStat(reforgeCaps, unitStat) != 0 {
		return true
	}
	for _, softCap := range softCaps {
		if softCap.unitStat == unitStat {
			return true
		}
	}
	return false
}

func statHasCap(stat stats.Stat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromStat(stat)
	if getUnitStat(reforgeCaps, unitStat) != 0 {
		return true
	}
	for _, softCap := range softCaps {
		if softCap.unitStat == unitStat {
			return true
		}
	}
	return false
}

func pseudoStatIsCapped(pseudoStat proto.PseudoStat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromPseudoStat(pseudoStat)
	return getUnitStat(reforgeCaps, unitStat) < 0
}

func statIsCapped(stat stats.Stat, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	unitStat := stats.UnitStatFromStat(stat)
	return getUnitStat(reforgeCaps, unitStat) < 0
}

// includesStatWithCap reports whether any key in the coefficient map names a stat that has a
// configured cap.
func includesStatWithCap(coeffs map[string]float64, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	for key := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			if unitStat.IsPseudoStat() {
				if pseudoStatHasCap(proto.PseudoStat(unitStat.PseudoStatIdx()), reforgeCaps, softCaps) {
					return true
				}
			} else if statHasCap(stats.Stat(unitStat.StatIdx()), reforgeCaps, softCaps) {
				return true
			}
		}
	}
	return false
}

// includesCappedStat reports whether any key in the coefficient map names a stat that is already
// capped.
func includesCappedStat(coeffs map[string]float64, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) bool {
	for key := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			if unitStat.IsPseudoStat() {
				if pseudoStatIsCapped(proto.PseudoStat(unitStat.PseudoStatIdx()), reforgeCaps, softCaps) {
					return true
				}
			} else if statIsCapped(stats.Stat(unitStat.StatIdx()), reforgeCaps, softCaps) {
				return true
			}
		}
	}
	return false
}

// getCappedStatKeys returns the coefficient keys whose stat has a configured cap (hard or soft).
func getCappedStatKeys(coeffs map[string]float64, reforgeCaps core.UnitStats, softCaps []*reforgeSoftCap) []string {
	var keys []string
	for key := range coeffs {
		unitStat, ok := unitStatFromCoeffKey(key)
		if !ok {
			continue
		}
		if unitStat.IsPseudoStat() {
			if pseudoStatHasCap(proto.PseudoStat(unitStat.PseudoStatIdx()), reforgeCaps, softCaps) {
				keys = append(keys, key)
			}
		} else if statHasCap(stats.Stat(unitStat.StatIdx()), reforgeCaps, softCaps) {
			keys = append(keys, key)
		}
	}
	return keys
}

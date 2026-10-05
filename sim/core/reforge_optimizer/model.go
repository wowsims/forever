package reforgeoptimizer

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// model.go holds the coefficient layer of the optimizer's LP model: how a stat delta becomes a
// variable's objective and cap coefficients, and how a variable is scored. It is a port of the
// reference reforger's model layer: applyReforgeStat and its variable scoring.
//
// Coefficient space: the reference keeps ONE coefficient map per variable — the same map feeds both
// the objective score and the cap constraints, so a stat only counts toward a cap when
// applyReforgeStat happened to resolve it onto that cap's pseudo-stat (which it does only for a root
// stat carrying no EP). This port splits the two spaces:
//   - objByName keeps the EP-calibrated applyReforgeStat output, so the objective stays exactly as
//     the reference calibrates it.
//   - byName holds the CAP coefficients, built by pushing the variable's RAW stat delta through the
//     full stat dependency graph (resolveCapCoeffs), plus the structural/constraint keys. Cap rows
//     and checkCaps therefore also count the dependencies applyReforgeStat never models at all —
//     Intellect -> SpellCrit%, Agility -> Crit%/Dodge% — instead of silently under-counting them.
//
// The racial multipliers (Human Spirit x1.1, Gnome Intellect x1.05) are registered in the stat
// dependency manager, so the cap space resolves the raw delta and must NOT pre-apply them; only the
// objective space applies them by hand (see applyReforgeStat).

// scoreCoeffKey is the objective key holding each variable's EP score.
const scoreCoeffKey = "score"

// computeCoeffScore sums weight*value over the coefficient keys that name a stat or pseudo-stat.
// Structural keys (slot keys and "score" itself) are not stat names and contribute nothing.
func computeCoeffScore(coeffs map[string]float64, weights core.UnitStats) float64 {
	score := 0.0
	for key, value := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			score += getUnitStat(weights, unitStat) * value
		}
	}
	return score
}

// updateReforgeScores recomputes every variable's "score" coefficient against the given weights
// and returns the updated variable set. The weights are expected to already carry post-cap values
// for any stat constrained to its cap in an earlier pass.
//
// The score is computed from the OBJECTIVE coefficients and written into the cap-space map, which is
// what buildLPObjective reads — so the objective stays exactly as calibrated while the constraint
// rows keep their stat-dependency-resolved coefficients.
func (o *reforgeOptimizer) updateReforgeScores(variables *lpVariables, weights core.UnitStats) *lpVariables {
	updated := newLPVariables()
	variables.each(func(name string, coeffs map[string]float64) {
		out := make(map[string]float64, len(coeffs)+1)
		for key, value := range coeffs {
			out[key] = value
		}
		objCoeffs := variables.getObj(name)
		out[scoreCoeffKey] = computeCoeffScore(objCoeffs, weights)
		updated.set(name, out)
		// Carry the objective coefficients forward so the cap-refinement recursion (which re-invokes
		// updateReforgeScores on this returned value) can re-score against the tightened weights
		// instead of collapsing every score to zero.
		updated.setObj(name, objCoeffs)
	})
	return updated
}

// scoreCoeffMap scores a standalone coefficient map, mirroring the reference's trick of running a
// single-entry variable map through updateReforgeScores.
func scoreCoeffMap(coeffs map[string]float64, weights core.UnitStats) map[string]float64 {
	scored := make(map[string]float64, len(coeffs)+1)
	for key, value := range coeffs {
		scored[key] = value
	}
	scored[scoreCoeffKey] = computeCoeffScore(coeffs, weights)
	return scored
}

// applyReforgeStat adds a stat amount to a coefficient map, applying the racial stat modifiers and
// resolving the stat onto its child pseudo-stats when the root stat itself carries no EP.
func (o *reforgeOptimizer) applyReforgeStat(coeffs map[string]float64, stat stats.Stat, amount float64, preCapEPs core.UnitStats) {
	race := o.player.GetRace()
	if stat == stats.Spirit && race == proto.Race_RaceHuman {
		amount *= 1.1
	}
	if stat == stats.Intellect && race == proto.Race_RaceGnome {
		amount *= 1.05
	}

	// If the pre-cap EP for the root stat is non-zero, apply the root stat directly and don't look
	// for any children.
	if preCapEPs.Stats[stat] != 0 {
		coeffs[statCoeffKey(proto.Stat(stat))] += amount
		return
	}

	for _, child := range childPseudoStats(stat) {
		if getUnitStat(preCapEPs, stats.UnitStatFromPseudoStat(child)) == 0 {
			continue
		}
		// ratingPerPseudoStatPercent encodes the per-parent conversion (including Defense, which
		// parents Dodge, Parry, Block and ReducedCritTakenPercent), so dividing by it reproduces the
		// reference's convertStatToChildPseudoStat exactly.
		if ratingPerPercent := ratingPerPseudoStatPercent(child, stat); ratingPerPercent != 0 {
			coeffs[pseudoStatCoeffKey(child)] += amount / ratingPerPercent
		}
	}
}

// resolveCapCoeffs builds a variable's CAP-space coefficient map from a raw stat delta by resolving
// it through the full stat dependency graph and keying every nonzero resolved stat/pseudo-stat by its
// coefficient-key name. These feed the LP cap constraint rows and checkCaps, so every dependency
// (Intellect -> SpellCrit%, Agility -> PhysicalCrit%/Dodge%, the haste speed multiplier) counts
// toward the caps. The racial stat multipliers live in the dependency manager, so rawStats must be
// passed through unscaled.
func (o *reforgeOptimizer) resolveCapCoeffs(rawDelta stats.Stats) map[string]float64 {
	resolved := resolveStatDelta(o.statDeps, o.baseStats, rawUnitStatsFromStats(rawDelta))
	o.sheet.Gate(resolved.PseudoStats)
	coeffs := map[string]float64{}
	eachUnitStat(resolved, func(unitStat stats.UnitStat, value float64) {
		if value != 0 {
			coeffs[coeffKeyForUnitStat(unitStat)] = value
		}
	})
	return coeffs
}

// epRelevantStats returns the stats this spec is willing to optimize for. The request carries the
// spec's configured list, which deliberately includes stats weighted at 0 (Stamina, spell
// penetration) so options carrying them stay selectable — that is why it cannot be derived from the
// EP weights.
//
// When the request has no list, fall back to deriving it: a stat counts when it, or any of its
// child pseudo-stats, carries a non-zero pre-cap EP. Either way attack power and ranged attack power
// imply each other, matching the reference.
func epRelevantStats(preCapEPs core.UnitStats, settings *proto.ReforgeSettings) map[stats.Stat]bool {
	result := make(map[stats.Stat]bool)
	if configured := settings.GetEpStats(); len(configured) > 0 {
		for _, stat := range configured {
			result[stats.Stat(stat)] = true
		}
	} else {
		for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
			stat := stats.Stat(statIdx)
			if preCapEPs.Stats[statIdx] != 0 {
				result[stat] = true
				continue
			}
			for _, child := range childPseudoStats(stat) {
				if getUnitStat(preCapEPs, stats.UnitStatFromPseudoStat(child)) != 0 {
					result[stat] = true
					break
				}
			}
		}
	}
	if result[stats.AttackPower] && !result[stats.RangedAttackPower] {
		result[stats.RangedAttackPower] = true
	} else if result[stats.RangedAttackPower] && !result[stats.AttackPower] {
		result[stats.AttackPower] = true
	}
	return result
}

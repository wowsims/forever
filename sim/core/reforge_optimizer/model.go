package reforgeoptimizer

import (
	"maps"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Holds the coefficient layer of the optimizer's LP model: how a stat delta becomes a variable's
// objective and cap coefficients, and how a variable is scored. It is a port of the reference
// reforger's model layer: applyReforgeStat and its variable scoring.
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

// Holds each variable's EP score, which the reported score sums.
const scoreCoeffKey = "score"

// Holds the score plus the variable's tie-break, which the LP maximizes.
const objectiveCoeffKey = "objective"

// The objective coefficient holding a variable's tie-break bonus.
const tieBreakCoeffKey = "tieBreak"

// Sums weight*value over the coefficient keys that name a stat or pseudo-stat. Structural keys
// (slot keys, "score", "objective" and "tieBreak") are not stat names and contribute nothing.
func computeCoeffScore(coeffs map[string]float64, weights core.UnitStats) float64 {
	score := 0.0
	for key, value := range coeffs {
		if unitStat, ok := unitStatFromCoeffKey(key); ok {
			score += getUnitStat(weights, unitStat) * value
		}
	}
	return score
}

// Recomputes every variable's "score" and "objective" coefficients against the
// given weights and returns the updated variable set. The weights are expected to already carry
// post-cap values for any stat constrained to its cap in an earlier pass.
//
// Both are computed from the OBJECTIVE coefficients and written into the cap-space map, which is
// what buildLPObjective reads — so the objective stays exactly as calibrated while the constraint
// rows keep their stat-dependency-resolved coefficients.
func (o *reforgeOptimizer) updateReforgeScores(variables *lpVariables, weights core.UnitStats) *lpVariables {
	updated := newLPVariables()
	variables.each(func(name string, coeffs map[string]float64) {
		out := make(map[string]float64, len(coeffs)+2)
		maps.Copy(out, coeffs)
		objCoeffs := variables.getObj(name)
		out[scoreCoeffKey] = computeCoeffScore(objCoeffs, weights)
		out[objectiveCoeffKey] = out[scoreCoeffKey] + objCoeffs[tieBreakCoeffKey]
		updated.set(name, out)
		// Carry the objective coefficients forward so the cap-refinement recursion (which re-invokes
		// updateReforgeScores on this returned value) can re-score against the tightened weights
		// instead of collapsing every score to zero.
		updated.setObj(name, objCoeffs)
	})
	return updated
}

// Adds a stat amount to a coefficient map, applying the racial stat modifiers and resolving the
// stat onto its child pseudo-stats when the root stat itself carries no EP.
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

// Builds a variable's CAP-space coefficient map from a raw stat delta by resolving
// it through the full stat dependency graph and keying every nonzero resolved stat/pseudo-stat by its
// coefficient-key name. These feed the LP cap constraint rows and checkCaps, so every dependency
// (Intellect -> SpellCrit%, Agility -> PhysicalCrit%/Dodge%, the haste speed multiplier) counts
// toward the caps. The racial stat multipliers live in the dependency manager, so the delta must be
// passed through unscaled. Percents go in delta.Stats as well as delta.PseudoStats, since
// resolveStatDelta writes every percent pseudo-stat from its Stat.
func (o *reforgeOptimizer) resolveCapCoeffs(delta core.UnitStats) map[string]float64 {
	resolved := resolveStatDelta(o.statDeps, o.baseStats, delta)
	o.sheet.Gate(resolved.PseudoStats)
	coeffs := map[string]float64{}
	eachUnitStat(resolved, func(unitStat stats.UnitStat, value float64) {
		if value != 0 {
			coeffs[coeffKeyForUnitStat(unitStat)] = value
		}
	})
	return coeffs
}

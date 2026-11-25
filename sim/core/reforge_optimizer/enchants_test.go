package reforgeoptimizer

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The EP the coefficients give at the weights, as the solve's objective adds them up.
func coeffsEP(coeffs map[string]float64, weights core.UnitStats) float64 {
	ep := 0.0
	for key, value := range coeffs {
		switch key {
		case statCoeffKey(proto.Stat_StatHitRating):
			ep += value * weights.Stats[stats.HitRating]
		case pseudoStatCoeffKey(proto.PseudoStat_PseudoStatRangedHitPercent):
			ep += value * getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent))
		}
	}
	return ep
}

func TestApplyPseudoStatsValuesRangedHitByClass(t *testing.T) {
	ratingWeights := core.NewUnitStats()
	ratingWeights.Stats[stats.HitRating] = 5
	rangedHitWeights := core.NewUnitStats()
	rangedHitWeights.Stats[stats.HitRating] = 5
	rangedHitWeights = setUnitStat(rangedHitWeights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent), 2)

	// 2523 Biznicks 247x128 Accurascope's 3% ranged hit, and a 1% hit that reaches melee and ranged attacks.
	scope := pseudoStatsProto(map[proto.PseudoStat]float64{proto.PseudoStat_PseudoStatRangedHitPercent: 3}).PseudoStats
	physicalHit := pseudoStatsProto(map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeHitPercent:  1,
		proto.PseudoStat_PseudoStatRangedHitPercent: 1,
	}).PseudoStats

	for _, test := range []struct {
		name        string
		pseudoStats []float64
		weights     core.UnitStats
		class       proto.Class
		want        float64
	}{
		{"scope on a warrior", scope, ratingWeights, proto.Class_ClassWarrior, 0},
		{"scope on a shaman", scope, ratingWeights, proto.Class_ClassShaman, 0},
		{"scope on a paladin", scope, ratingWeights, proto.Class_ClassPaladin, 0},
		{"scope on a hunter", scope, ratingWeights, proto.Class_ClassHunter, 3 * core.HitRatingPerHitPercent * 5},
		{"scope on a warrior with a ranged hit weight", scope, rangedHitWeights, proto.Class_ClassWarrior, 3 * 2},
		{"scope on a paladin with a ranged hit weight", scope, rangedHitWeights, proto.Class_ClassPaladin, 3 * 2},
		{"melee and ranged hit on a warrior", physicalHit, ratingWeights, proto.Class_ClassWarrior, core.HitRatingPerHitPercent * 5},
	} {
		coeffs := map[string]float64{}
		applyPseudoStats(coeffs, test.pseudoStats, test.weights, test.class)
		if got := coeffsEP(coeffs, test.weights); math.Abs(got-test.want) > 1e-9 {
			t.Errorf("%s: EP %v, want %v", test.name, got, test.want)
		}
	}
}

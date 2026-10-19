package reforgeoptimizer

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func pseudoStatsProto(values map[proto.PseudoStat]float64) *proto.UnitStats {
	pseudoStats := make([]float64, stats.PseudoStatsLen)
	for pseudoStat, value := range values {
		pseudoStats[pseudoStat] = value
	}
	return &proto.UnitStats{Stats: make([]float64, stats.ProtoStatsLen), PseudoStats: pseudoStats}
}

func rawUnitStatsFromStats(statValues stats.Stats) core.UnitStats {
	unitStats := core.NewUnitStats()
	for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
		if statValues[statIdx] != 0 {
			unitStats.Stats[statIdx] = statValues[statIdx]
		}
	}
	return unitStats
}

func ratingWeight(stat stats.Stat, ep float64) core.UnitStats {
	weights := core.NewUnitStats()
	weights.Stats[stat] = ep
	return weights
}

func TestReforgeWeightsKeepPercentWeightsPerPercent(t *testing.T) {
	weights := protoToCoreUnitStats(pseudoStatsProto(map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatBlockPercent:     5,
		proto.PseudoStat_PseudoStatMeleeHitPercent:  2,
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
	}))

	if got := computeCoeffScore(map[string]float64{pseudoStatCoeffKey(proto.PseudoStat_PseudoStatBlockPercent): 1}, weights); got != 5 {
		t.Errorf("one Block%% scores %v, want the 5 per percent the EP values state", got)
	}
	if got := getUnitStat(weights, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent)); got != 3 {
		t.Errorf("ranged hit weight %v, want 3", got)
	}
	for statIdx := int(stats.ProtoStatsLen); statIdx < int(stats.SimStatsLen); statIdx++ {
		if weights.Stats[statIdx] != 0 {
			t.Errorf("%s carries %v; the optimizer reads percent weights from PseudoStats only", stats.Stat(statIdx).StatName(), weights.Stats[statIdx])
		}
	}
}

func TestReforgeHitCapMatchesTheBackEndHitPercent(t *testing.T) {
	sdm := stats.NewStatDependencyManager()
	sdm.AddStatDependency(stats.HitRating, stats.PhysicalHitPercent, 1/core.HitRatingPerHitPercent)
	sdm.FinalizeStatDeps()

	baseStats := protoToCoreUnitStats(pseudoStatsProto(map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeHitPercent: 7,
		proto.PseudoStat_PseudoStatBlockPercent:    25,
	}))
	caps := computeStatCapsDelta(baseStats, protoToCoreUnitStats(pseudoStatsProto(map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeHitPercent: 9,
		proto.PseudoStat_PseudoStatBlockPercent:    30,
	})))

	meleeHit := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent)
	if gap := getUnitStat(caps, meleeHit); gap != 2 {
		t.Fatalf("9%% hit cap leaves a gap of %v over 7%% hit, want 2", gap)
	}
	if gap := getUnitStat(caps, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatBlockPercent)); gap != 5 {
		t.Errorf("30%% block cap leaves a gap of %v over 25%% block, want 5", gap)
	}

	delta := core.NewUnitStats()
	delta.Stats[stats.HitRating] = 2 * core.HitRatingPerHitPercent
	resolved := resolveStatDelta(&sdm, baseStats, delta)
	if got := getUnitStat(resolved, meleeHit); math.Abs(got-2) > 1e-9 {
		t.Errorf("rating for 2%% hit fills %v of the hit cap gap, want 2", got)
	}
}

func TestReforgeBlockDeltaLandsInPercent(t *testing.T) {
	sdm := stats.NewStatDependencyManager()
	sdm.AddStatDependency(stats.Strength, stats.BlockPercent, 0.5)
	sdm.FinalizeStatDeps()

	delta := core.NewUnitStats()
	delta.Stats[stats.Strength] = 10
	resolved := resolveStatDelta(&sdm, core.NewUnitStats(), delta)
	if got := resolved.Stats[stats.BlockPercent]; math.Abs(got-5) > 1e-12 {
		t.Fatalf("back-end BlockPercent delta %v, want 5 percent", got)
	}
	if got := getUnitStat(resolved, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatBlockPercent)); math.Abs(got-5) > 1e-9 {
		t.Errorf("Block%% delta %v, want 5 percent", got)
	}
}

func TestReforgeDodgeAndParryRatingReachTheirPercentCaps(t *testing.T) {
	sdm := stats.NewStatDependencyManager()
	sdm.AddStatDependency(stats.DodgeRating, stats.DodgePercent, 1/core.DodgeRatingPerDodgePercent)
	sdm.AddStatDependency(stats.ParryRating, stats.ParryPercent, 1/core.ParryRatingPerParryPercent)
	sdm.FinalizeStatDeps()

	delta := core.NewUnitStats()
	delta.Stats[stats.DodgeRating] = 2 * core.DodgeRatingPerDodgePercent
	delta.Stats[stats.ParryRating] = 3 * core.ParryRatingPerParryPercent
	resolved := resolveStatDelta(&sdm, core.NewUnitStats(), delta)
	if got := getUnitStat(resolved, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatDodgePercent)); math.Abs(got-2) > 1e-9 {
		t.Errorf("rating for 2%% dodge moves Dodge%% by %v, want 2", got)
	}
	if got := getUnitStat(resolved, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatParryPercent)); math.Abs(got-3) > 1e-9 {
		t.Errorf("rating for 3%% parry moves Parry%% by %v, want 3", got)
	}
}

func TestReforgeDefenseReachesTheCritTakenCap(t *testing.T) {
	sdm := stats.NewStatDependencyManager()
	core.AddRatingConversions(&sdm)
	sdm.FinalizeStatDeps()

	delta := core.NewUnitStats()
	delta.Stats[stats.DefenseRating] = 25 * core.DefenseRatingPerDefenseLevel
	resolved := resolveStatDelta(&sdm, core.NewUnitStats(), delta)
	if got := getUnitStat(resolved, stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatReducedCritTakenPercent)); math.Abs(got-1) > 1e-9 {
		t.Errorf("25 defense moves crit taken by %v, want 1", got)
	}
}

func TestReforgeCritTakenCountsWholeDefense(t *testing.T) {
	sdm := stats.NewStatDependencyManager()
	core.AddRatingConversions(&sdm)
	sdm.FinalizeStatDeps()

	critTaken := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatReducedCritTakenPercent)
	for _, test := range []struct {
		base, delta, want float64
	}{
		{10, 2, 2 / core.DefenseRatingPerAvoidancePercent},
		{10.5, 0.5, 1 / core.DefenseRatingPerAvoidancePercent},
		{10, 0.5, 0},
		{10.5, 2.4, 2 / core.DefenseRatingPerAvoidancePercent},
	} {
		baseStats := core.NewUnitStats()
		baseStats.Stats[stats.DefenseRating] = test.base * core.DefenseRatingPerDefenseLevel
		delta := core.NewUnitStats()
		delta.Stats[stats.DefenseRating] = test.delta * core.DefenseRatingPerDefenseLevel
		resolved := resolveStatDelta(&sdm, baseStats, delta)
		if got := getUnitStat(resolved, critTaken); math.Abs(got-test.want) > 1e-9 {
			t.Errorf("%v defense on %v moves crit taken by %v, want %v", test.delta, test.base, got, test.want)
		}
	}
}

func TestReforgeAvoidanceRatingWeightMovesOntoItsCappedPercent(t *testing.T) {
	for _, test := range []struct {
		rating           stats.Stat
		percent          proto.PseudoStat
		ratingPerPercent float64
	}{
		{stats.DodgeRating, proto.PseudoStat_PseudoStatDodgePercent, core.DodgeRatingPerDodgePercent},
		{stats.ParryRating, proto.PseudoStat_PseudoStatParryPercent, core.ParryRatingPerParryPercent},
		{stats.BlockRating, proto.PseudoStat_PseudoStatBlockPercent, core.BlockRatingPerBlockPercent},
	} {
		percent := stats.UnitStatFromPseudoStat(test.percent)
		caps := setUnitStat(core.NewUnitStats(), percent, 2)

		validated := checkWeights(ratingWeight(test.rating, 1), caps, nil)
		if validated.Stats[test.rating] != 0 {
			t.Errorf("%s keeps weight %v under a %s cap, want 0", test.rating.StatName(), validated.Stats[test.rating], test.percent)
		}
		if got := getUnitStat(validated, percent); got != test.ratingPerPercent {
			t.Errorf("%s weighs %v per percent, want %v", test.percent, got, test.ratingPerPercent)
		}

		coeffs := map[string]float64{}
		o := &reforgeOptimizer{player: &proto.Player{}}
		o.applyReforgeStat(coeffs, test.rating, 3*test.ratingPerPercent, validated)
		if got := coeffs[pseudoStatCoeffKey(test.percent)]; math.Abs(got-3) > 1e-9 {
			t.Errorf("the rating for 3%% counts %v toward %s, want 3", got, test.percent)
		}
	}
}

func TestReforgeDefenseWeightMovesOntoItsFirstCappedPercent(t *testing.T) {
	critTaken := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatReducedCritTakenPercent)
	dodge := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatDodgePercent)

	validated := checkWeights(ratingWeight(stats.DefenseRating, 1), setUnitStat(setUnitStat(core.NewUnitStats(), critTaken, 2), dodge, 2), nil)
	if got := getUnitStat(validated, critTaken); got != core.DefenseRatingPerAvoidancePercent {
		t.Errorf("crit taken weighs %v per percent, want the 25 defense one percent costs", got)
	}
	if got := getUnitStat(validated, dodge); got != 0 {
		t.Errorf("Dodge%% takes weight %v from defense under a crit taken cap, want 0", got)
	}

	validated = checkWeights(ratingWeight(stats.DefenseRating, 1), setUnitStat(core.NewUnitStats(), dodge, 2), nil)
	if got := getUnitStat(validated, dodge); got != core.DefenseRatingPerAvoidancePercent {
		t.Errorf("Dodge%% weighs %v per percent under a dodge cap alone, want the 25 defense one percent costs", got)
	}
	if validated.Stats[stats.DefenseRating] != 0 {
		t.Errorf("defense keeps weight %v, want 0", validated.Stats[stats.DefenseRating])
	}
}

func TestReforgeRatingKeepsItsWeightOverItsPercentWeights(t *testing.T) {
	rangedHit := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedHitPercent)
	rangedCrit := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatRangedCritPercent)
	weights := ratingWeight(stats.HitRating, 2)
	weights.Stats[stats.CritRating] = 1.5
	weights = setUnitStat(weights, rangedHit, 50)
	weights = setUnitStat(weights, rangedCrit, 40)
	softCaps := []*reforgeSoftCap{{
		unitStat:    rangedHit,
		breakpoints: []float64{3},
		postCapEPs:  []float64{0},
		capType:     proto.StatCapType_TypeSoftCap,
	}}
	validated := checkWeights(weights, core.NewUnitStats(), softCaps)
	o := &reforgeOptimizer{player: &proto.Player{}, undershootCaps: core.NewUnitStats()}
	hitCoeffs := map[string]float64{}
	o.applyReforgeStat(hitCoeffs, stats.HitRating, 10, validated)

	t.Run("checkWeights", func(t *testing.T) {
		if got := getUnitStat(validated, rangedHit); got != 2*core.HitRatingPerHitPercent {
			t.Errorf("ranged hit weighs %v per percent, want the hit rating's 2 per rating", got)
		}
		for _, child := range childPseudoStats(stats.HitRating) {
			if child != proto.PseudoStat_PseudoStatRangedHitPercent && getUnitStat(validated, stats.UnitStatFromPseudoStat(child)) != 0 {
				t.Errorf("%s keeps a weight beside the capped ranged hit", child)
			}
		}
		if validated.Stats[stats.CritRating] != 1.5 || getUnitStat(validated, rangedCrit) != 0 {
			t.Errorf("uncapped crit weighs %v per rating and %v per ranged crit percent, want 1.5 and 0",
				validated.Stats[stats.CritRating], getUnitStat(validated, rangedCrit))
		}
	})

	t.Run("score", func(t *testing.T) {
		if got := computeCoeffScore(hitCoeffs, validated); math.Abs(got-20) > 1e-9 {
			t.Errorf("10 hit rating scores %v, want 20", got)
		}
		critCoeffs := map[string]float64{}
		o.applyReforgeStat(critCoeffs, stats.CritRating, 10, validated)
		if got := computeCoeffScore(critCoeffs, validated); math.Abs(got-15) > 1e-9 {
			t.Errorf("10 crit rating scores %v, want 15", got)
		}
	})

	t.Run("checkCaps", func(t *testing.T) {
		variables := newLPVariables()
		variables.set("hit", map[string]float64{coeffKeyForUnitStat(rangedHit): 4})
		variables.setObj("hit", hitCoeffs)
		exceeded, constraints, capped, _ := o.checkCaps(lpSolution{variables: []string{"hit"}}, core.NewUnitStats(), softCaps, variables, newLPConstraints(), validated)
		if !exceeded {
			t.Fatal("ranged hit past the soft cap gap leaves the cap unenforced")
		}
		if got := constraints.byName[coeffKeyForUnitStat(rangedHit)]; got != greaterEq(3) {
			t.Errorf("ranged hit row %+v, want at least the 3%% gap", got)
		}
		if got := computeCoeffScore(hitCoeffs, capped); got != 0 {
			t.Errorf("10 hit rating past the soft cap scores %v, want the post-cap 0", got)
		}
	})
}

func TestReforgeRatingWithoutPercentWeightsIsUnchanged(t *testing.T) {
	meleeHit := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent)

	validated := checkWeights(ratingWeight(stats.HitRating, 2), setUnitStat(core.NewUnitStats(), meleeHit, 2), nil)
	if validated.Stats[stats.HitRating] != 0 || getUnitStat(validated, meleeHit) != 2*core.HitRatingPerHitPercent {
		t.Errorf("hit under a melee hit cap weighs %v per rating and %v per percent, want 0 and %v",
			validated.Stats[stats.HitRating], getUnitStat(validated, meleeHit), 2*core.HitRatingPerHitPercent)
	}

	validated = checkWeights(ratingWeight(stats.HitRating, 2), core.NewUnitStats(), nil)
	if validated.Stats[stats.HitRating] != 2 || getUnitStat(validated, meleeHit) != 0 {
		t.Errorf("uncapped hit weighs %v per rating and %v per percent, want 2 and 0",
			validated.Stats[stats.HitRating], getUnitStat(validated, meleeHit))
	}
}

func TestReforgeUnweightedRatingKeepsItsPercentWeight(t *testing.T) {
	meleeHit := stats.UnitStatFromPseudoStat(proto.PseudoStat_PseudoStatMeleeHitPercent)
	validated := checkWeights(setUnitStat(core.NewUnitStats(), meleeHit, 30), setUnitStat(core.NewUnitStats(), meleeHit, 2), nil)
	if validated.Stats[stats.HitRating] != 0 || getUnitStat(validated, meleeHit) != 30 {
		t.Errorf("an unweighted hit rating leaves melee hit at %v per percent, want 30", getUnitStat(validated, meleeHit))
	}
}

func TestReforgeCapSpaceSkipsBlockAndParryTheCharacterLacks(t *testing.T) {
	sdm := stats.NewStatDependencyManager()
	sdm.AddStatDependency(stats.DodgeRating, stats.DodgePercent, 1/core.DodgeRatingPerDodgePercent)
	sdm.AddStatDependency(stats.ParryRating, stats.ParryPercent, 1/core.ParryRatingPerParryPercent)
	sdm.AddStatDependency(stats.BlockRating, stats.BlockPercent, 1/core.BlockRatingPerBlockPercent)
	sdm.FinalizeStatDeps()

	var delta stats.Stats
	delta[stats.DodgeRating] = core.DodgeRatingPerDodgePercent
	delta[stats.ParryRating] = core.ParryRatingPerParryPercent
	delta[stats.BlockRating] = core.BlockRatingPerBlockPercent

	for _, sheet := range []core.SheetAvoidance{{CanBlock: true, CanParry: true}, {CanParry: true}, {CanBlock: true}, {}} {
		o := &reforgeOptimizer{statDeps: &sdm, baseStats: core.NewUnitStats(), sheet: sheet}
		coeffs := o.resolveCapCoeffs(rawUnitStatsFromStats(delta))
		for _, want := range []struct {
			pseudoStat proto.PseudoStat
			counts     bool
		}{
			{proto.PseudoStat_PseudoStatDodgePercent, true},
			{proto.PseudoStat_PseudoStatBlockPercent, sheet.CanBlock},
			{proto.PseudoStat_PseudoStatParryPercent, sheet.CanParry},
		} {
			_, counts := coeffs[pseudoStatCoeffKey(want.pseudoStat)]
			if counts != want.counts {
				t.Errorf("can block %v, parry %v: %s in the cap space is %v", sheet.CanBlock, sheet.CanParry, want.pseudoStat, counts)
			}
		}
	}
}

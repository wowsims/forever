package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestOneRatingFeedsPhysicalAndSpellPercents(t *testing.T) {
	unit := &Unit{StatDependencyManager: stats.NewStatDependencyManager()}
	unit.addUniversalStatDependencies()
	unit.FinalizeStatDeps()
	unit.stats = unit.ApplyStatDependencies(stats.Stats{
		stats.HitRating:  HitRatingPerHitPercent,
		stats.CritRating: CritRatingPerCritPercent,
	})

	for _, percent := range []stats.Stat{
		stats.PhysicalHitPercent,
		stats.SpellHitPercent,
		stats.PhysicalCritPercent,
		stats.SpellCritPercent,
	} {
		if got := unit.GetStat(percent); math.Abs(got-1) > 1e-12 {
			t.Errorf("%s from one percent of rating: got %v, want 1", percent.StatName(), got)
		}
	}
}

func TestHasteRatingMovesEverySpeed(t *testing.T) {
	unit := &Unit{PseudoStats: stats.NewPseudoStats()}
	unit.stats[stats.HasteRating] = 10 * HasteRatingPerHastePercent

	for _, want := range []struct {
		name string
		got  float64
	}{
		{"melee", unit.TotalMeleeHasteMultiplier()},
		{"ranged", unit.TotalRangedHasteMultiplier()},
		{"real", unit.TotalRealHasteMultiplier()},
		{"real ranged", unit.TotalRealRangedHasteMultiplier()},
		{"spell", unit.TotalSpellHasteMultiplier()},
	} {
		if math.Abs(want.got-1.1) > 1e-12 {
			t.Errorf("%s haste multiplier from 10%% haste rating: got %v, want 1.1", want.name, want.got)
		}
	}
}

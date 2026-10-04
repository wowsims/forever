package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

func TestDefenseRaisesChanceToBeMissed(t *testing.T) {
	table := &AttackTable{MH: AttackTableHand{BaseMissChance: 0.056}}
	spell := &Spell{ProcMask: ProcMaskMeleeMHAuto}
	missedWith := func(defenseRating float64) float64 {
		unit := &Unit{PseudoStats: stats.NewPseudoStats()}
		unit.stats[stats.DefenseRating] = defenseRating
		return unit.GetTotalChanceToBeMissedAsDefender(spell, table)
	}

	if got := missedWith(0); math.Abs(got-0.056) > 1e-12 {
		t.Errorf("no defense: missed %v, want the table's 0.056", got)
	}
	if got := missedWith(25.9 * DefenseRatingPerDefenseLevel); math.Abs(got-(0.056+25*0.0004)) > 1e-12 {
		t.Errorf("25.9 defense: missed %v, want 0.04%% for each of the 25 whole points on top", got)
	}
}

func TestBlockValueFromStrengthSitsOutsideTheMultiplier(t *testing.T) {
	for _, tc := range []struct {
		name                             string
		strength, blockValue, multiplier float64
		want                             float64
	}{
		{"10 strength", 10, 0, 1, 0},
		{"120 strength", 120, 0, 1, 5},
		{"10 strength and a 30 block value shield", 10, 30, 1, 30},
		{"120 strength and a 30 block value shield", 120, 30, 1, 35},
		{"120 strength and a 30 block value shield at 1.3 times block value", 120, 30, 1.3, 5 + 39},
	} {
		unit := &Unit{PseudoStats: stats.NewPseudoStats()}
		unit.PseudoStats.BlockValuePerStrength = 0.05
		unit.PseudoStats.BlockValueMultiplier = tc.multiplier
		unit.stats[stats.Strength] = tc.strength
		unit.stats[stats.BlockValue] = tc.blockValue
		if got := unit.BlockDamageReduction(); math.Abs(got-tc.want) > 1e-12 {
			t.Errorf("%s: block value %v, want %v", tc.name, got, tc.want)
		}
	}
}

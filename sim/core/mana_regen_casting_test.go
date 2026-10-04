package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// The shares of spirit regen that continue while casting add up, and past all of it they give all of
// it: Innervate's 100% on top of a 30% talent is the regen the unit has while not casting.
func TestSpiritRegenWhileCastingStopsAtAllOfIt(t *testing.T) {
	unit := &Unit{PseudoStats: stats.NewPseudoStats()}
	unit.manaRegenMultiplier = 1
	unit.spiritRegenBase, unit.spiritRegenPerSpirit = 7.5, 1.0/10
	unit.stats[stats.Spirit] = 100

	full := unit.ManaRegenPerSecondWhileNotCasting()
	spirit := unit.SpiritManaRegenPerSecond()

	for _, c := range []struct {
		share float64
		want  float64
	}{
		{0, 0},
		{0.3, spirit * 0.3},
		{1, full},
		{1.3, full},
	} {
		unit.PseudoStats.SpiritRegenRateCasting = c.share
		if got := unit.ManaRegenPerSecondWhileCasting(); got != c.want {
			t.Errorf("casting share %v: %v mana per second while casting, want %v", c.share, got, c.want)
		}
	}
}

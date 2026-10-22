//go:build with_db

package feralcat

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

// Excess Energy holds back, for each pending refresh in time order, the cost that the 10 energy a
// second until then does not cover. Rip runs out in 1s and Rake in 3s: Rip floats its cost less 10,
// Rake its cost less the 20 of the next two seconds.
func TestExcessEnergyHoldsBackThePendingRefreshes(t *testing.T) {
	excess := func(comboPoints int32, fightLeft time.Duration) (float64, *FeralDruid) {
		sim, cat := newCat(t, "", nil)
		target := cat.CurrentTarget
		rip, rake := cat.Rip.Dot(target), cat.Rake.Dot(target)
		rip.Apply(sim)
		// Rake starts so that it runs out 2s after Rip.
		sim.CurrentTime = rip.ExpiresAt() + 2*time.Second - rake.BaseDuration()
		rake.Apply(sim)
		sim.CurrentTime = rip.ExpiresAt() - time.Second
		sim.Duration = sim.CurrentTime + fightLeft
		cat.AddComboPoints(sim, comboPoints, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
		cat.AddEnergy(sim, 40, cat.NewEnergyMetrics(core.ActionID{SpellID: 2}))
		value := cat.NewAPLValue(nil, &proto.APLValue{Value: &proto.APLValue_DruidExcessEnergy{DruidExcessEnergy: &proto.APLValueDruidExcessEnergy{}}})
		return value.GetFloat(sim), cat
	}

	got, cat := excess(5, time.Minute)
	ripCost, rakeCost := cat.Rip.DefaultCast.Cost, cat.Rake.DefaultCast.Cost
	if want := 40 - (ripCost - 10) - (rakeCost - 20); math.Abs(got-want) > 1e-9 {
		t.Errorf("both refreshes pending: excess %g, want 40 - (%g - 10) - (%g - 20) = %g", got, ripCost, rakeCost, want)
	}
	// Without 5 combo points no Rip is coming, so only Rake floats, against 30 energy of regen.
	if got, _ := excess(4, time.Minute); math.Abs(got-(40-max(0, rakeCost-30))) > 1e-9 {
		t.Errorf("4 combo points: excess %g, want %g", got, 40-max(0, rakeCost-30))
	}
	// With 5s of fight left neither bleed would be refreshed.
	if got, _ := excess(5, 5*time.Second); got != 40 {
		t.Errorf("fight ending: excess %g, want all 40", got)
	}
}

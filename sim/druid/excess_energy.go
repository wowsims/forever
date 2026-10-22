package druid

import (
	"cmp"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
)

// A Rip or Rake refresh the cat will have to pay for: when the bleed runs out, and what the new cast
// costs.
type pendingRefresh struct {
	at   time.Duration
	cost float64
}

// The energy the cat can spend now without delaying a coming Rip or Rake refresh.
//
// A refresh counts when its bleed is up and the fight lasts long enough after it runs out for a new one
// to be worth casting: Rip at 5 combo points with 10s of fight left, Rake with a full Rake left. Taking
// the refreshes in time order, the energy regenerated before each one pays for it; what regeneration
// cannot cover has to be kept from now. The result is the current energy minus what has to be kept.
func (druid *Druid) excessEnergy(sim *core.Simulation) float64 {
	target := druid.CurrentTarget
	remaining := sim.GetRemainingDuration()
	refreshes := make([]pendingRefresh, 0, 2)
	if druid.Rip != nil && druid.ComboPoints() >= 5 {
		if dot := druid.Rip.Dot(target); dot.IsActive() && dot.RemainingDuration(sim) <= remaining-10*time.Second {
			refreshes = append(refreshes, pendingRefresh{dot.ExpiresAt(), druid.Rip.DefaultCast.Cost})
		}
	}
	if druid.Rake != nil {
		if dot := druid.Rake.Dot(target); dot.IsActive() && dot.RemainingDuration(sim) <= remaining-dot.BaseDuration() {
			refreshes = append(refreshes, pendingRefresh{dot.ExpiresAt(), druid.Rake.DefaultCast.Cost})
		}
	}
	slices.SortStableFunc(refreshes, func(a, b pendingRefresh) int { return cmp.Compare(a.at, b.at) })

	regen := druid.EnergyPerTick * druid.CurrentEnergyRegenMultiplier() / druid.EnergyTickDuration.Seconds()
	// The first refresh after Shifting Power is ready costs Shifting Power's energy less. When Shifting
	// Power gives more than that refresh costs, the cost goes below zero and the energy left over goes to
	// the next refresh.
	shiftEnergy := 0.0
	shiftReadyAt := core.NeverExpires
	if druid.ShiftingPower != nil && druid.Talents.ShiftingPower {
		shiftEnergy = spellData.ShiftingPower.Highest().EnergizeEffect().Average() + druid.WolfsheadEnergyBonus
		shiftReadyAt = druid.ShiftingPower.ReadyAt()
	}
	keep := 0.0
	// The time up to which the regeneration has already paid for earlier refreshes.
	regenUsedUntil := sim.CurrentTime
	for _, refresh := range refreshes {
		if shiftEnergy > 0 && shiftReadyAt <= refresh.at {
			refresh.cost -= shiftEnergy
			shiftEnergy = 0
		}
		// No regeneration in Bear and Dire Bear Form, so the whole cost has to be kept.
		if regen <= 0 {
			keep += max(0, refresh.cost)
			continue
		}
		if gain := (refresh.at - regenUsedUntil).Seconds() * regen; gain < refresh.cost {
			// Regeneration falls short: keep the difference, and all of it up to this refresh is used.
			keep += refresh.cost - gain
			regenUsedUntil = refresh.at
		} else {
			// Regeneration covers it: only the time it takes to pay is used. A cost below zero gives time
			// back, so the next refresh gets that much more regeneration.
			regenUsedUntil += core.DurationFromSeconds(refresh.cost / regen)
		}
	}
	return druid.CurrentEnergy() - keep
}

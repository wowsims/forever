package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// In game each energy Shifting Power gives threatens for 0.5, without the form's threat multiplier.
const shiftingPowerThreatPerEnergy = 0.5

func (druid *Druid) registerShiftingPowerSpell() {
	rank := spellData.ShiftingPower.Highest()
	energy := rank.EnergizeEffect().Average()
	metrics := druid.NewEnergyMetrics(core.ActionID{SpellID: rank.ID})
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	// The October 8 hotfixes: "Shifting Power can no longer be used while already at Full Energy."
	// The rows state no such rule.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.Talents.ShiftingPower && druid.CurrentEnergy() < druid.MaximumEnergy()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		before := druid.CurrentEnergy()
		druid.AddEnergy(sim, energy+druid.WolfsheadEnergyBonus, metrics)
		spell.ApplySplitThreat(sim, (druid.CurrentEnergy()-before)*shiftingPowerThreatPerEnergy)
	}
	druid.ShiftingPower = druid.RegisterSpell(config)
}

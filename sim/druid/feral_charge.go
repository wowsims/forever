package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerFeralChargeSpell() {
	rank := spellData.FeralCharge.Highest()
	root := spellData.FeralChargeTriggered.Highest()
	roots := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(root))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.Talents.FeralCharge
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		// TODO: Ingame research needed -- the row is a charge that triggers a bear-only root and
		// interrupt. It does not state a leap behind the target, and movement is not modelled.
		if !druid.InForm(Bear) {
			return
		}
		roots.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = roots.ToMap()
	druid.FeralCharge = druid.RegisterSpell(config)
}

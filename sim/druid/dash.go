package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerDashSpell() {
	rank := spellData.Dash.Highest()
	druid.DashAura = druid.RegisterAura(spelldata.AuraConfig(rank))
	// Dash and Feral Swiftness use the strongest movement bonus.
	movement := druid.DashAura.NewPassiveMovementSpeedEffect(rank.EffectN(1).Percent())
	druid.CatFormAura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		movement.Deactivate(sim)
	})
	druid.CatFormAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		if druid.DashAura.IsActive() {
			movement.Activate(sim)
		}
	})
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.DashAura.Activate(sim)
	}
	config.RelatedSelfBuff = druid.DashAura
	druid.Dash = druid.RegisterSpell(config)
}

package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerGrowlSpell() {
	druid.Growl = druid.registerBearUtility(spellData.Growl.Highest())
}

// A taunt: castable from an APL, with no effect the sim models.
func (druid *Druid) registerBearUtility(rank *spelldata.Spell) *DruidSpell {
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(_ *core.Simulation, _ *core.Unit, _ *core.Spell) {}
	return druid.RegisterSpell(config)
}

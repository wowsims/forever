package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerThornsSpell() {
	rank := spellData.Thorns.Highest()
	aura := buffs.NewSpellDataThornsAuraWithSpellPower(&druid.Unit, true, 0)
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	config.RelatedSelfBuff = aura
	druid.Thorns = druid.RegisterSpell(config)
}

package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerBarkskin() {
	rank := spellData.Barkskin.Highest()
	aura := druid.RegisterAura(spelldata.AuraConfig(rank))
	spelldata.ParseEffects(&druid.Character, aura, rank)
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	config.RelatedSelfBuff = aura
	druid.Barkskin = druid.RegisterSpell(config)
}

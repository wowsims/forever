package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var thornsRank = spellData.Thorns.Highest()

// Self-cast Thorns. The druid's own copy of the generated shield deals the game files' 22 nature
// damage; the raid buff registers the external copy before Initialize runs, and the two auras bid in
// ThornsCategory, which holds one of them at a time.
func (druid *Druid) registerThornsSpell() {
	thornsAura := druid.GetAura("Thorns (Player)")
	if thornsAura == nil {
		thornsAura = buffs.NewSpellDataThornsAura(&druid.Unit, true, 0)
	}

	config := spelldata.SpellConfig(&druid.Unit, thornsRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = DruidSpellThorns
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		thornsAura.Activate(sim)
	}
	config.RelatedSelfBuff = thornsAura

	druid.RegisterSpell(config)
}

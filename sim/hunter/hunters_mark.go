package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var huntersMarkRank = spellData.HuntersMark.Highest()

func (hunter *Hunter) registerHuntersMark() {
	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataHuntersMarkAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&hunter.Unit, huntersMarkRank, spelldata.Flags(core.SpellFlagAPL))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		auras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = auras.ToMap()

	hunter.RegisterSpell(config)
}

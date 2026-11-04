package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ShadowWordDeathRankMap = spellData.ShadowWordDeath

func (priest *Priest) registerShadowWordDeathSpell(_ int32, rank *spelldata.Spell) {
	backlash := rank.EffectN(3).Percent()

	config := priest.directDamageSpellConfig(rank, PriestSpellShadowWordDeath)
	dealDamage := config.ApplyEffects
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		dealDamage(sim, target, spell)
		priest.SpendHealth(sim, priest.MaxHealth()*backlash, spell.HealthMetrics(&priest.Unit))
	}

	priest.RegisterSpell(config)
}

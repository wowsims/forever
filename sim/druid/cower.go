package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerCowerSpell() {
	rank := spellData.Cower.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskEmpty))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, 0, spell.OutcomeMeleeSpecialHit)
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Cower = druid.RegisterSpell(config)
}

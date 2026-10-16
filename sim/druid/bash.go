package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The row rolls a melee hit and refunds on a miss; the stun lands only on a hit.
func (druid *Druid) registerBashSpell() {
	rank := spellData.Bash.Highest()
	stuns := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name)))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			stuns.Get(target).Activate(sim)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	config.RelatedAuraArrays = stuns.ToMap()
	druid.Bash = druid.RegisterSpell(config)
}

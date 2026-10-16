package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The 3% attack power share is an override on the row: the game files state none.
func (druid *Druid) registerSwipeBearSpell() {
	rank := spellData.Swipe.Highest()
	damageEffect := rank.DamageEffect()
	base := damageEffect.Average()
	targets := int32(damageEffect.ChainTargets)
	if targets < 1 {
		targets = int32(rank.MaxTargets)
	}
	var results core.SpellResultSlice
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		cleave(sim, target, spell, targets, &results, func(target *core.Unit) *core.SpellResult {
			damage := base + damageEffect.APCoeff()*spell.MeleeAttackPower(target)
			return spell.CalcDamage(sim, target, damage*druid.rendAndTearMultiplier(target), spell.OutcomeMeleeSpecialHitAndCrit)
		})
	}
	druid.Swipe = druid.RegisterSpell(config)
}

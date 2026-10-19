package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var interceptRank = spellData.Intercept.Highest()

// The damage sits on the stun the charge triggers, not on Intercept itself.
var interceptStunDamage = interceptRank.EffectN(2).Trigger().DamageEffect().Average()

func (warrior *Warrior) registerIntercept() {
	actionID := core.ActionID{SpellID: interceptRank.ID}
	chargeMinRange := float64(interceptRank.MinRange)

	var spell *core.Spell
	var interceptTarget *core.Unit

	config := spelldata.SpellConfig(&warrior.Unit, interceptRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMovesToTarget))

	aura := warrior.registerDashAura("Intercept", actionID, config.Cast.CD.Duration, func(sim *core.Simulation) {
		spell.CalcAndDealDamage(sim, interceptTarget, interceptStunDamage, spell.OutcomeAlwaysHit)
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		interceptTarget = target
		aura.Duration = spell.CD.Duration
		aura.Activate(sim)
		warrior.MoveTo(chargeMinRange-3.5, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
	}

	spell = warrior.RegisterSpell(config)
}

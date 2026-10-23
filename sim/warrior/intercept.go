package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var interceptRank = spellData.Intercept.Highest()

// The damage sits on the stun the charge triggers, not on Intercept itself.
var interceptStun = interceptRank.EffectN(2).Trigger()
var interceptStunDamage = interceptStun.DamageEffect().Average()

func (warrior *Warrior) registerIntercept() {
	actionID := core.ActionID{SpellID: interceptRank.ID}
	chargeMinRange := float64(interceptRank.MinRange)

	var interceptTarget *core.Unit

	// Physical school with the magic defense type in the game files, like Thunder Clap: it rolls on the
	// spell hit table and crits on spell crit chance for 1.5x. Most physical warrior damage rows state a
	// spell power share of 1, so the share is not read.
	stunConfig := spelldata.SpellConfig(&warrior.Unit, interceptStun, spelldata.Proc())
	stunConfig.ProcMask = core.ProcMaskRangedSpecial
	stunConfig.DamageMultiplier = 1
	stunConfig.ThreatMultiplier = 1
	stun := warrior.RegisterSpell(stunConfig)

	config := spelldata.SpellConfig(&warrior.Unit, interceptRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMovesToTarget))

	aura := warrior.registerDashAura("Intercept", actionID, config.Cast.CD.Duration, func(sim *core.Simulation) {
		// The dash also ends when the iteration is torn down, which runs with the clock at the
		// fight's end; the stun does not land then.
		if sim.CurrentTime == sim.Duration {
			return
		}
		// Targets at raid boss level are immune to the stun and its damage; no game files row states it.
		if interceptTarget.Level >= core.CharacterLevel+3 {
			return
		}
		stun.CalcAndDealDamage(sim, interceptTarget, interceptStunDamage, stun.OutcomeMagicHitAndCrit)
	})

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		interceptTarget = target
		aura.Duration = spell.CD.Duration
		aura.Activate(sim)
		warrior.MoveTo(chargeMinRange-3.5, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
	}

	warrior.RegisterSpell(config)
}

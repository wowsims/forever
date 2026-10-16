package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// E_HEALTH_LEECH: the caster heals for all the damage dealt. The fear (effect 2) does not affect sims, and
// the 2 min cooldown is the row's category cooldown.
func (warlock *Warlock) registerDeathCoil() {
	rankedSpell := spellData.DeathCoil.ByID(17926)
	deathCoilHit := rankedSpell.EffectN(1)

	healthMetrics := warlock.NewHealthMetrics(core.ActionID{SpellID: rankedSpell.ID})

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellDeathCoil
	// SpellConfig reads the spell power share off a damage or heal effect, and E_HEALTH_LEECH is neither.
	config.BonusCoefficient = deathCoilHit.Coeff()

	// Roll the damage and to hit, then land it after the travel time and heal for what it dealt
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, deathCoilHit.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealCastResult(sim, result)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
			heal := result.Damage * warlock.PseudoStats.SelfHealingMultiplier * warlock.PseudoStats.HealingTakenMultiplier
			warlock.GainHealth(sim, heal, healthMetrics)
		})
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, deathCoilHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	warlock.RegisterSpell(config)
}

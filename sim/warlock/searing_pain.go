package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerSearingPain() {
	rankedSpell := spellData.SearingPain.ByID(17923)
	searingPainHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellSearingPain
	// The tooltip's "Causes a high amount of threat" states no number and the row has no threat effect.
	config.ThreatMultiplier = 2 // TODO: Measure Searing Pain's threat in game.

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, searingPainHit.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealDamage(sim, result)
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, searingPainHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	warlock.SearingPain = warlock.RegisterSpell(config)
}

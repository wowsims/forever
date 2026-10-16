package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Incinerate talent: effect 2 is the bonus against a target afflicted by Immolate.
// TODO: Assumed any Immolate counts, not only this warlock's; confirm in-game.
func (warlock *Warlock) registerIncinerate() {
	rankedSpell := spellData.Incinerate.ByID(1293813)
	incinerateHit := rankedSpell.DamageEffect()
	immolateMult := 1 + rankedSpell.EffectN(2).Percent()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellIncinerate

	// Roll the damage and to hit, raised if any Immolate burns on the target, then land it after travel
	// TODO: Confirm in game that the 25% raises the whole hit, spell power included.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, incinerateHit.Average(), spell.OutcomeMagicHitAndCrit)
		if target.HasActiveAuraWithTag(immolateTag) {
			scaleResult(sim, spell, result, immolateMult)
		}
		spell.DealDamageAfterTravel(sim, result)
	}

	// Calculate expected damage from the row's average hit, raised if any Immolate burns on the target
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		result := spell.CalcDamage(sim, target, incinerateHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
		if target.HasActiveAuraWithTag(immolateTag) {
			scaleResult(sim, spell, result, immolateMult)
		}
		return result
	}

	warlock.Incinerate = warlock.RegisterSpell(config)
}

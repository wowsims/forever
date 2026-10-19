package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Shadowburn talent. The 15 sec cooldown is the row's category cooldown, and the Soul Shard
// each cast takes the row's reagent; the one a kill returns is not modelled.
func (warlock *Warlock) registerShadowBurn() {
	rankedSpell := spellData.Shadowburn.ByID(18871)
	shadowburnHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellShadowBurn

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, shadowburnHit.Average(), spell.OutcomeMagicHitAndCrit)
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, shadowburnHit.Average(), spell.OutcomeExpectedMagicCrit)
	}

	warlock.useReagents(&config, rankedSpell)
	warlock.Shadowburn = warlock.RegisterSpell(config)
}

package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerHolyFireSpell(_ int32, rank *spelldata.Spell) {
	holyFireHit := rank.DamageEffect()
	// TODO: Verify rank 5 (15265) in game. Its tick states 13 where the ranks around it state 7 and 10
	// (4/5/6/7/13/10/13/15 by rank, the same 0.05 spell power share on each).
	config := priest.periodicDamageSpellConfig(rank, PriestSpellHolyFire)
	priest.attachPowerInLightMarker(&config)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, holyFireHit.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, holyFireHit.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}

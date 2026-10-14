package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerBindingHeal(_ int32, rank *spelldata.Spell) {
	primaryHeal := rank.EffectN(1)
	secondaryHeal := rank.EffectN(2)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellBindingHeal
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		target = priest.healTarget(target)
		priest.withRenewedHope(sim, target, spell, func() {
			spell.CalcAndDealHealing(sim, target, primaryHeal.Average(), spell.OutcomeHealingCrit)
			spell.CalcAndDealHealing(sim, &priest.Unit, secondaryHeal.Average(), spell.OutcomeHealingCrit)
		})
	}

	priest.RegisterSpell(config)
}

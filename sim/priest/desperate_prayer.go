package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerDesperatePrayer(_ int32, rank *spelldata.Spell) {
	heal := rank.HealEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, &priest.Unit, heal.Average(), spell.OutcomeHealingCrit)
	}

	priest.RegisterSpell(config)
}

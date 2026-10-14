package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerRenew(_ int32, rank *spelldata.Spell) {
	healTick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellHealing))
	config.ClassSpellMask = PriestSpellRenew

	config.Hot = spelldata.DotConfig(rank, healTick, spelldata.RankLabel(rank, &priest.Unit))

	config.Hot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicHealing(sim, target, healTick.Average(), dot.Spell.OutcomeTickHealingCrit)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Hot(priest.healTarget(target)).Apply(sim)
	}

	priest.RegisterSpell(config)
}

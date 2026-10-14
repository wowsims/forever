package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var DevouringPlagueRankMap = spellData.DevouringPlague

func (priest *Priest) registerDevouringPlagueSpell(_ int32, rank *spelldata.Spell) {
	actionID := core.ActionID{SpellID: rank.ID}
	tick := rank.PeriodicEffect()

	healthMetrics := priest.NewHealthMetrics(actionID.WithTag(1))

	config := priest.periodicDamageSpellConfig(rank, PriestSpellDevouringPlague)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.Roll(sim), rank.TickOutcome(dot))

		priest.GainHealth(sim, result.Damage, healthMetrics)
	}

	priest.RegisterSpell(config)
}

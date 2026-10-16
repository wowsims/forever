package priest

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerMindFlaySpell(_ int32, rank *spelldata.Spell) {
	config := priest.periodicDamageSpellConfig(rank, PriestSpellMindFlay)

	if priest.Talents.ImprovedMindFlay > 0 {
		// Improved Mind Flay: each tick deals X% more than the one before it, so the nth tick of the
		// channel is worth (1 + X%)^(n-1) of the first.
		ramp := spellData.ImprovedMindFlay.EffectAt(1).MultiplierAt(priest.Talents.ImprovedMindFlay)
		onTick := config.Dot.OnTick
		config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			multiplier := dot.PeriodicDamageMultiplier
			dot.PeriodicDamageMultiplier *= math.Pow(ramp, float64(dot.TickCount()-1))
			onTick(sim, target, dot)
			dot.PeriodicDamageMultiplier = multiplier
		}
	}

	priest.RegisterSpell(config)
}

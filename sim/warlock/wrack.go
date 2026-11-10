package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// While the channel is on the target, this warlock's Corruption, Siphon Life and Bane of Agony there deal their
// periodic damage faster by the percent of the row's dummy effect: they keep their duration and tick every
// period / (1 + percent) for as long as the channel runs, so the channel adds ticks. The normal period returns
// when the channel ends or is interrupted. The time left to the tick already scheduled scales with the period at
// either end. An effect applied while the channel runs keeps its normal period, since casting it ends the channel.
// TODO: Ingame test needed: whether a tick already scheduled when the channel starts or ends is re-timed, and
// whether Bane of Agony's added ticks take the ramp of the third they land in.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	speedup := 1 + rankedSpell.Effect(dbcenums.A_DUMMY, 0).Percent()

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	config.Dot.Aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			dot.SetTickPeriod(sim, time.Duration(float64(dot.CalcTickPeriod())/speedup).Round(time.Millisecond))
		}
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			dot.SetTickPeriod(sim, dot.CalcTickPeriod())
		}
	}

	warlock.Wrack = warlock.RegisterSpell(config)
}

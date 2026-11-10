package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// While the channel is on the target, this warlock's Corruption, Siphon Life and Bane of Agony there tick every
// period / (1 + the row's dummy percent), keeping their expiry (core.Dot.SetTickPeriod), so the channel adds
// ticks. The normal period returns when the channel ends. An effect applied while the channel runs keeps its
// normal period, since casting it ends the channel.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	speedup := 1 + rankedSpell.Effect(dbcenums.A_DUMMY, 0).Percent()

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	setPeriods := func(sim *core.Simulation, target *core.Unit, divisor float64) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(target)
			dot.SetTickPeriod(sim, time.Duration(float64(dot.CalcTickPeriod())/divisor).Round(time.Millisecond))
		}
	}
	config.Dot.Aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		setPeriods(sim, aura.Unit, speedup)
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		setPeriods(sim, aura.Unit, 1)
	}

	warlock.Wrack = warlock.RegisterSpell(config)
}

// Records on the dot's spell the part of a tick's damage Wrack contributed: the damage times the share of the
// tick's progress the channel added (core.Dot.SpedUpShare), which the tick checks is above 0 so a dot Wrack never
// sped up records nothing. The faster ticks reach Bane of Agony's stronger thirds sooner; that counts for nothing
// extra, as the share applies to the damage the tick dealt.
func recordWrackContribution(target *core.Unit, dot *core.Dot, damage float64, share float64) {
	dot.Spell.RecordCustomMetric(target, "wrack", damage*share)
}

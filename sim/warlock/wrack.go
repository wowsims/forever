package warlock

import (
	"slices"
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

	// Each tick of the effects the channel speeds up records on its spell, under the key wrack, the part of its
	// damage Wrack contributed: the damage times the share of the tick's progress the channel added
	// (core.Dot.SpedUpShare), above 0 only for a tick it sped up. A tick before the pull records nothing, as its
	// damage is not counted either. The faster ticks reach Bane of Agony's stronger thirds sooner; that counts
	// for nothing extra, as the share applies to the damage the tick dealt.
	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Wrack Contribution",
		OnPeriodicDamageDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if sim.CurrentTime < 0 || !slices.Contains(warlock.afflictionDots, spell) {
				return
			}
			if share := spell.Dot(result.Target).SpedUpShare(sim); share > 0 {
				spell.RecordCustomMetric(result.Target, "wrack", result.Damage*share)
			}
		},
	}))
}

package warlock

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// While the channel is on the target, this warlock's Corruption, Siphon Life and Bane of Agony there tick every
// period / (1 + the row's dummy percent), at their application plus whole periods (core.Dot.SetTickPeriod). At half
// the period each normal tick stays, and the half points between them tick too: each one from the instant the
// channel starts up to the instant it ends, both included. A channel that starts on a half point ticks it at once,
// and a recast at the instant the channel ends ticks that point once. Every tick is full and the expiry stays;
// starting or ending the channel moves no tick. An effect applied while the channel runs keeps its normal period,
// since casting it ends the channel.
// TODO: Ingame test needed: combat logs time a channel's start and end with up to 0.2 s of jitter, so they leave a
// channel starting or ending exactly on a half point unmeasured. Cast Wrack so it lands within 0.05 s before and
// after Siphon Life's half point at 1.5 s, and end it within 0.05 s before and after one, many times each, and count
// the ticks at that half point.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	speedup := 1 + rankedSpell.Effect(dbcenums.A_DUMMY, 0).Percent()

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	// The listener below takes a tick on the effect's application plus whole normal periods for a normal one, which
	// holds while the channel's period divides the normal one. The effects register after this talent, so the check
	// runs as the channel starts: a row that breaks it panics on the first channel.
	fastPeriod := func(dot *core.Dot) time.Duration {
		normal := dot.CalcTickPeriod()
		fast := time.Duration(float64(normal) / speedup).Round(time.Millisecond)
		if normal%fast != 0 {
			panic(fmt.Sprintf("Wrack ticks %s every %v, which does not divide its period of %v", dot.Spell.ActionID, fast, normal))
		}
		return fast
	}
	config.Dot.Aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			dot.SetTickPeriod(sim, fastPeriod(dot))
		}
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			dot.SetTickPeriod(sim, dot.CalcTickPeriod())
		}
	}

	warlock.Wrack = warlock.RegisterSpell(config)

	// A tick of these effects off their normal grid, on a half point only the channel's period gives, is wholly
	// Wrack's: it records its damage on its spell under the key wrack. A tick on the normal grid records nothing,
	// nor does one before the pull, as its damage is not counted either.
	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Wrack Contribution",
		OnPeriodicDamageDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if sim.CurrentTime < 0 || !slices.Contains(warlock.afflictionDots, spell) {
				return
			}
			if dot := spell.Dot(result.Target); (sim.CurrentTime-dot.StartedAt())%dot.CalcTickPeriod() != 0 {
				spell.RecordCustomMetric(result.Target, "wrack", result.Damage)
			}
		},
	}))
}

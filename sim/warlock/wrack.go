package warlock

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// While the channel is on the target, this warlock's Corruption, Siphon Life and Bane of Agony there tick every
// period / (1 + the row's dummy percent). At half the period each normal tick stays, and the half points between
// them tick too: each one from the instant the channel starts up to the instant it ends, both included. A channel
// that starts on a half point ticks it at once, and a recast at the instant the channel ends ticks that point once.
// An effect applied while the channel runs keeps its normal period, since casting it ends the channel.
// The start is measured (EV-0085).
// TODO: Ingame test needed: the end is not measured (EV-0085). The logs' 0.05 s steps cannot split a channel
// ending close to a half point. End Wrack within 0.05 s before and after a half point,
// many times each, and count the ticks there.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	speedup := 1 + rankedSpell.Effect(dbcenums.A_DUMMY, 0).Percent()

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	// The listener below holds while the channel's period divides the normal one. The effects register after this
	// talent, so the check runs as the channel starts.
	fastPeriod := func(dot *core.Dot) time.Duration {
		normal := dot.CalcTickPeriod()
		fast := time.Duration(float64(normal) / speedup).Round(time.Millisecond)
		if normal%fast != 0 {
			panic(fmt.Sprintf("Wrack ticks %s every %v, which does not divide its period of %v", dot.Spell.ActionID, fast, normal))
		}
		return fast
	}
	// When each effect was last re-timed, cleared each iteration as the clock starts over.
	retimedAt := map[*core.Dot]time.Duration{}
	warlock.RegisterResetEffect(func(_ *core.Simulation) {
		clear(retimedAt)
	})
	config.Dot.Aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			retimeOnGrid(sim, dot, fastPeriod(dot), retimedAt)
		}
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			retimeOnGrid(sim, dot, dot.CalcTickPeriod(), retimedAt)
		}
	}

	warlock.Wrack = warlock.RegisterSpell(config)

	// A tick of these effects off their normal grid is wholly Wrack's. One before the pull records nothing, as its
	// damage is not counted either.
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

// Ticks a running effect every period from now on, at its expiry less whole periods, so its last tick lands on the
// expiry in full. RestoreState drops the queued tick, so a tick due now deals here, once: also one whose action the
// sim has already taken off the queue. A point only the new period has ticks here at once, unless the effect was
// re-timed at this instant before, so a channel recast at the instant one ends does not tick that point again. A
// point already passed does not tick. The call must not come from inside the effect's own tick, where the tick
// queues the next one a second time.
func retimeOnGrid(sim *core.Simulation, dot *core.Dot, period time.Duration, retimedAt map[*core.Dot]time.Duration) {
	if !dot.IsActive() || period == dot.TickPeriod() {
		return
	}
	state := dot.SaveState(sim)
	untilExpiry := state.RemainingDuration
	lastAt, retimed := retimedAt[dot]
	retimedAt[dot] = sim.CurrentTime
	newPointNow := untilExpiry%period == 0 && untilExpiry%state.TickPeriod != 0 && !(retimed && lastAt == sim.CurrentTime)
	tickNow := state.NextTickIn == 0 || newPointNow

	state.TickPeriod = period
	state.TicksRemaining = int32((untilExpiry + period - 1) / period)
	state.NextTickIn = untilExpiry - time.Duration(state.TicksRemaining-1)*period
	if tickNow {
		// Counted before it deals, as the effect's own tick does; the next tick is already after now, so an
		// expiry the tick brings about finds none due.
		state.TicksDone++
	}
	dot.RestoreState(state, sim)
	if tickNow {
		dot.TickOnce(sim)
	}
}

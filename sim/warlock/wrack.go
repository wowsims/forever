package warlock

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// While the channel is on the target, this warlock's Corruption, Siphon Life and Bane of Agony there tick every
// period / (1 + the row's dummy percent), on the grid of their expiry less whole periods. At half the period each
// normal tick stays, and the half points between them tick too: each one from the instant the channel starts up to
// the instant it ends, the start included and the end left out. A channel that starts on a half point ticks it at
// once, and a recast at the instant a channel ends ticks that point once. A half point ticks after everything else
// due at its instant, so a channel ending or starting there is settled first. Every tick is full, and the effect's
// own ticks and expiry stay as they are. An effect applied while the channel runs keeps its normal period, since
// casting it ends the channel. In game a half point ticks when the channel is up at its instant, and a 6 s channel
// lined up on the grid ticks one of the two half points at its edges, as one off the grid does (EV-0085).
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	speedup := 1 + rankedSpell.Effect(dbcenums.A_DUMMY, 0).Percent()

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	// The half points are the points of the channel's period that are not the effect's own ticks, which holds while
	// the channel's period divides the normal one. The effects register after this talent, so the check runs as
	// the channel starts: a row that breaks it panics on the first channel.
	fastPeriod := func(dot *core.Dot) time.Duration {
		normal := dot.TickPeriod()
		fast := time.Duration(float64(normal) / speedup).Round(time.Millisecond)
		if normal%fast != 0 {
			panic(fmt.Sprintf("Wrack ticks %s every %v, which does not divide its period of %v", dot.Spell.ActionID, fast, normal))
		}
		return fast
	}
	// Each effect's next half point under the channel, cleared each iteration as the clock starts over.
	pending := map[*core.Dot]*core.PendingAction{}
	warlock.RegisterResetEffect(func(_ *core.Simulation) {
		clear(pending)
	})
	// Queues the effect's half point that lies points whole channel periods before its expiry, or the next one
	// before it when that point is one of the effect's own ticks.
	var queueHalfPoint func(sim *core.Simulation, channel *core.Aura, dot *core.Dot, points time.Duration)
	queueHalfPoint = func(sim *core.Simulation, channel *core.Aura, dot *core.Dot, points time.Duration) {
		delete(pending, dot)
		fast, expiresAt := fastPeriod(dot), dot.ExpiresAt()
		if (points*fast)%dot.TickPeriod() == 0 {
			points--
		}
		if points <= 0 {
			return
		}
		pa := &core.PendingAction{NextActionAt: expiresAt - points*fast, Priority: core.ActionPriorityLow}
		pa.OnAction = func(sim *core.Simulation) {
			delete(pending, dot)
			if !channel.IsActive() || !dot.IsActive() || dot.ExpiresAt() != expiresAt {
				return
			}
			dot.TickOnce(sim)
			queueHalfPoint(sim, channel, dot, points-1)
		}
		pending[dot] = pa
		sim.AddPendingAction(pa)
	}
	config.Dot.Aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			if dot := spell.Dot(aura.Unit); dot.IsActive() {
				queueHalfPoint(sim, aura, dot, (dot.ExpiresAt()-sim.CurrentTime)/fastPeriod(dot))
			}
		}
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		for _, spell := range warlock.afflictionDots {
			dot := spell.Dot(aura.Unit)
			if pa := pending[dot]; pa != nil {
				pa.Cancel(sim)
				delete(pending, dot)
			}
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

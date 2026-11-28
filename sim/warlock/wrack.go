package warlock

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// While the channel is on the target, the half points of this warlock's Corruption, Siphon Life and Bane of Agony
// there tick too, in full, from the channel's start up to its end. In game a channel lined up on the grid ticks one
// of its two edge points, the start's and the end's about as often (EV-0085); here it is always the start's. A half
// point runs at the lowest priority, so a channel starting, ending or recast at its instant is settled first.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	speedup := 1 + rankedSpell.Effect(dbcenums.A_DUMMY, 0).Percent()

	config := warlock.soulSiphonChannelConfig(rankedSpell, wrackTick, nil)
	config.ClassSpellMask = WarlockSpellWrack

	// The half period under the channel must divide the effect's own. The effects register after this talent, so the
	// check runs when a channel first starts over each effect.
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
	var halfPointDot *core.Dot // the effect whose half point is ticking
	warlock.RegisterResetEffect(func(_ *core.Simulation) {
		clear(pending)
	})
	// Queues the effect's half point that lies points half periods before its expiry, or the one after it when that
	// point is one of the effect's own ticks.
	var queueHalfPoint func(sim *core.Simulation, channel *core.Aura, dot *core.Dot, points time.Duration)
	queueHalfPoint = func(sim *core.Simulation, channel *core.Aura, dot *core.Dot, points time.Duration) {
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
			halfPointDot = dot
			dot.TickOnce(sim)
			halfPointDot = nil
			if channel.IsActive() && dot.IsActive() {
				queueHalfPoint(sim, channel, dot, points-1)
			}
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

	// A half point's tick records its damage on its spell under the key wrack; none before the pull.
	core.MakePermanent(warlock.RegisterAura(core.Aura{
		Label: "Wrack Contribution",
		OnPeriodicDamageDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if dot := halfPointDot; dot != nil && sim.CurrentTime >= 0 && spell == dot.Spell && result.Target == dot.Unit {
				spell.RecordCustomMetric(result.Target, "wrack", result.Damage)
			}
		},
	}))
}

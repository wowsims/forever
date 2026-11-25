package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Rank 1 hits for 5 at half the mana of the highest rank, a cheap main-hand special for the on-hit
// effects alone.
func (hunter *Hunter) registerWingClip() {
	hunter.registerWingClipRank(spellData.WingClip.Rank(1))
	hunter.registerWingClipRank(spellData.WingClip.Highest())
}

// Effect 1 is the snare, which a boss ignores; effect 2 is the flat hit. As a main-hand special it can
// proc Windfury Totem and other on-hit effects.
func (hunter *Hunter) registerWingClipRank(rank *spelldata.Spell) {
	hit := rank.EffectN(2)
	config := spelldata.SpellConfig(&hunter.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.VersusStats = []stats.Stat{stats.PhysicalDamage}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, hit.Roll(sim), spell.OutcomeMeleeSpecialHitAndCrit)
	}

	hunter.RegisterSpell(config)
}

package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 1 is a cheap main-hand special for the on-hit effects alone.
func (hunter *Hunter) registerWingClip() {
	hunter.registerWingClipRank(spellData.WingClip.Rank(1))
	hunter.registerWingClipRank(spellData.WingClip.Highest())
}

// Effect 1 is the snare, which a boss ignores; effect 2 is the flat hit.
func (hunter *Hunter) registerWingClipRank(rank *spelldata.Spell) {
	hit := rank.EffectN(2)
	config := spelldata.SpellConfig(&hunter.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, hit.Roll(sim), spell.OutcomeMeleeSpecialHitAndCrit)
	}

	hunter.RegisterSpell(config)
}

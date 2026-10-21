package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var wingClipRank = spellData.WingClip.Highest()

// Effect 1 is the snare, which a boss ignores; effect 2 is the flat hit. As a main-hand special it can
// proc Windfury Totem and other on-hit effects.
var wingClipHit = wingClipRank.EffectN(2)

func (hunter *Hunter) registerWingClip() {
	config := spelldata.SpellConfig(&hunter.Unit, wingClipRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, wingClipHit.Roll(sim), spell.OutcomeMeleeSpecialHitAndCrit)
	}

	hunter.RegisterSpell(config)
}

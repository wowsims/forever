package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneShotRank = spellData.ArcaneShot.Highest()
var arcaneShotBaseDamage = arcaneShotRank.DamageEffect().Average()

// Attack power share in EV-0068.
const arcaneShotAPShare = 0.15

func (hunter *Hunter) registerArcaneShot() {
	config := spelldata.SpellConfig(&hunter.Unit, arcaneShotRank, spelldata.Melee(core.ProcMaskRangedSpecial))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := arcaneShotBaseDamage + arcaneShotAPShare*spell.RangedAttackPower(target)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

		spell.DealDamageAfterTravel(sim, result)
	}

	hunter.RegisterSpell(config)
}

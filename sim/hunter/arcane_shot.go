package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneShotRank = spellData.ArcaneShot.Highest()
var arcaneShotBaseDamage = arcaneShotRank.DamageEffect().Average()

// TODO: In-game testing required. The row states no attack power share; the Forever notes say the
// shot scales with 15% of ranged attack power.
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

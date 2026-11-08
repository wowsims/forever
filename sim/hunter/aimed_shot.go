package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var aimedShotRank = spellData.AimedShot.Highest()

func (hunter *Hunter) registerAimedShot() {
	config := spelldata.SpellConfig(&hunter.Unit, aimedShotRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	hunter.hasteRangedCast(&config)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := aimedShotRank.WeaponStrike(sim, spell, hunter.RangedNormalizedWeaponDamage(sim, spell.RangedAttackPower(target)), normalizedShotBonus)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

		spell.DealDamageAfterTravel(sim, result)
	}

	hunter.RegisterSpell(config)
}

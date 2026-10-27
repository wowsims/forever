package hunter

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var multiShotRank = spellData.MultiShot.Highest()
var multiShotTargets = int32(multiShotRank.DamageEffect().ChainTargets)

func (hunter *Hunter) registerMultiShot() {
	config := spelldata.SpellConfig(&hunter.Unit, multiShotRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	// In game Multi-Shot has a 0.5 s cast, which the game files leave out.
	config.Cast.DefaultCast.CastTime = 500 * time.Millisecond
	hunter.hasteRangedCast(&config)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := hunter.RangedNormalizedWeaponDamage(sim, spell.RangedAttackPower(target))
		for _, result := range spell.CalcCleaveDamage(sim, target, multiShotTargets, baseDamage, spell.OutcomeRangedHitAndCrit) {
			spell.DealCastResult(sim, result)
		}

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealBatchedAoeDamage(sim)
		})
	}

	hunter.RegisterSpell(config)
}

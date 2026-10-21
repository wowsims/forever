package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.ConeOfCold lists.
func (mage *Mage) registerConeOfColdSpell() {
	mage.registerConeOfColdRank(spellData.ConeOfCold.Len(), spellData.ConeOfCold.Highest())
}

// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerConeOfColdRank(_ int32, rank *spelldata.Spell) {
	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellConeOfCold

	damage := rank.DamageEffect()

	// Winter's Chill raises the coefficient against each target by that target's own stacks, which
	// this hit then spends, so each target is calculated on its own.
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			bonusCoefficient := mage.wintersChillCoefficient(aoeTarget)

			spell.BonusCoefficient += bonusCoefficient
			result := spell.CalcAndDealDamage(sim, aoeTarget, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
			spell.BonusCoefficient -= bonusCoefficient

			if result.Landed() {
				mage.consumeWintersChill(sim, aoeTarget)
			}
		}
	}

	mage.RegisterSpell(config)
}

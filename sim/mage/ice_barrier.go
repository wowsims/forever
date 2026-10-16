package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.IceBarrier lists.
func (mage *Mage) registerIceBarrierSpell() {
	mage.registerIceBarrierRank(spellData.IceBarrier.Len(), spellData.IceBarrier.Highest())
}

// Ice Barrier
// https://www.wowhead.com/forever/spell=13033
//
// "Instantly shields you, absorbing $s1 damage. Lasts 1 min. While the shield holds, your spellcasts
// will not be interrupted or delayed from taking damage." The row states no spell power coefficient,
// and the sim models no pushback, so only the absorb is.
func (mage *Mage) registerIceBarrierRank(_ int32, rank *spelldata.Spell) {
	absorb := rank.EffectN(1).Average()

	shield := mage.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &mage.Unit)),
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return absorb
		},
	})

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellIceBarrier
	config.RelatedSelfBuff = shield.Aura

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		shield.Activate(sim)
	}

	mage.RegisterSpell(config)
}

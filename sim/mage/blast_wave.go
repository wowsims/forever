package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.BlastWave lists.
func (mage *Mage) registerBlastWaveSpell() {
	mage.registerBlastWaveRank(spellData.BlastWave.Len(), spellData.BlastWave.Highest())
}

// The daze (effect 2) does nothing to a boss.
func (mage *Mage) registerBlastWaveRank(_ int32, rank *spelldata.Spell) {
	// Binary: a spell with a non-damage effect (the daze) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellBlastWave

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}

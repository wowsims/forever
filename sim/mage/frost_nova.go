package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered.
func (mage *Mage) registerFrostNovaSpell() {
	mage.registerFrostNovaRank(spellData.FrostNova.Len(), spellData.FrostNova.Highest())
}

// Frost Nova
// https://www.wowhead.com/forever/spell=10230
//
// The root does nothing to a boss the sim fights, so only the damage is modelled.
func (mage *Mage) registerFrostNovaRank(_ int32, rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFrostNova

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}

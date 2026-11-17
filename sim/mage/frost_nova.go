package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.FrostNova lists.
func (mage *Mage) registerFrostNovaSpell() {
	mage.registerFrostNovaRank(spellData.FrostNova.Len(), spellData.FrostNova.Highest())
}

// Frost Nova
// https://www.wowhead.com/forever/spell=10230
//
// Frost damage to every enemy around the caster, and a root. The root does nothing to a boss the
// sim fights, so only the damage is modelled.
func (mage *Mage) registerFrostNovaRank(_ int32, rank *spelldata.Spell) {
	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFrostNova

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, rollDamage(rank.DamageEffect()))
	}

	mage.RegisterSpell(config)
}

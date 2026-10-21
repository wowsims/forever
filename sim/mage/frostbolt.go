package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.Frostbolt lists.
func (mage *Mage) registerFrostboltSpell() {
	//Rank 10 is last trainer taught spell
	mage.registerFrostboltRank(spellData.Frostbolt.Len(), spellData.Frostbolt.Rank(10))
}

func (mage *Mage) registerFrostboltRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostbolt

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	mage.RegisterSpell(config)
}

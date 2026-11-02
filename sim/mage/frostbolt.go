package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 10 is the last rank the trainer teaches; rank 11 comes from the Tome of Frostbolt XI book.
var frostboltTrainerRank = spellData.Frostbolt.ByID(10181)
var frostboltBookRank = spellData.Frostbolt.ByID(25304)

func (mage *Mage) registerFrostboltSpell() {
	mage.registerFrostboltRank(frostboltTrainerRank)
	mage.registerFrostboltRank(frostboltBookRank)
}

func (mage *Mage) registerFrostboltRank(rank *spelldata.Spell) {
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

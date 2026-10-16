package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.FrostBlast lists.
func (mage *Mage) registerFrostBlastSpell() {
	mage.registerFrostBlastRank(spellData.FrostBlast.Len(), spellData.FrostBlast.Highest())
}

// Frost Blast
// https://www.wowhead.com/forever/spell=1326055
//
// "Quickly launch a weak blast of frost at the enemy, causing $s2 Frost damage and slowing movement
// speed by $s1% for $d." The row carries Frostbolt's class flags, so every talent that names
// Frostbolt by its mask reaches it, and the Chill mask makes it a Chill for Frostbite and Fingers of
// Frost. The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerFrostBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
	}

	mage.RegisterSpell(config)
}

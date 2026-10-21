package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.IceLance lists.
func (mage *Mage) registerIceLanceSpell() {
	mage.registerIceLanceRank(spellData.IceLance.Len(), spellData.IceLance.Highest())
}

// Ice Lance
// https://www.wowhead.com/forever/spell=1240047
//
// Frost damage to one enemy, raised by effect 2's percentage ("Deals $s2% increased damage to Frozen
// targets", 200) while the target is frozen.
//
// The damage effect states no spell power coefficient; the store carries one as an override
// (tools/database/overrides). Effects 3 (20) and 4 (5) are server-side scripts the tooltip does not
// name, and are not modelled.
func (mage *Mage) registerIceLanceRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()
	frozenMultiplier := 1 + rank.EffectN(2).Percent()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellIceLance

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		multiplier := 1.0
		if mage.isFrozen(target) {
			multiplier = frozenMultiplier
		}

		spell.DamageMultiplier *= multiplier
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DamageMultiplier /= multiplier

		spell.DealDamageAfterTravel(sim, result)
	}

	mage.RegisterSpell(config)
}

package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.ArcaneBlast lists.
func (mage *Mage) registerArcaneBlastSpell() {
	mage.registerArcaneBlastRank(spellData.ArcaneBlast.Len(), spellData.ArcaneBlast.Highest())
}

func (mage *Mage) registerArcaneBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneBlast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			// The charges last from the first one, so a new stack leaves the duration alone.
			if !mage.ArcaneChargesAura.IsActive() {
				mage.ArcaneChargesAura.Activate(sim)
			}
			mage.ArcaneChargesAura.AddStack(sim)
		}
	}

	mage.RegisterSpell(config)
}

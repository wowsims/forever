package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered.
func (mage *Mage) registerArcaneBlastSpell() {
	mage.registerArcaneBlastRank(spellData.ArcaneBlast.Len(), spellData.ArcaneBlast.Highest())
}

func (mage *Mage) registerArcaneBlastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellArcaneBlast

	// The cast adds its Arcane Charge before its damage. Charge timing in EV-0190.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if !mage.ArcaneChargesAura.IsActive() {
			mage.ArcaneChargesAura.Activate(sim)
		}
		mage.ArcaneChargesAura.AddStack(sim)

		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	mage.RegisterSpell(config)
}

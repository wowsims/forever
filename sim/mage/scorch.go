package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerScorchSpell() {
	procChance := spellData.ImprovedScorch.EffectAt(1).FractionAt(mage.Talents.ImprovedScorch)

	if mage.Talents.ImprovedScorch > 0 {
		mage.ImprovedScorchAuras = mage.newFireVulnerabilityAuras()
	}

	// Only the highest rank is registered.
	mage.registerScorchRank(spellData.Scorch.Highest(), procChance)
}

func (mage *Mage) registerScorchRank(rank *spelldata.Spell, procChance float64) {
	damage := rank.DamageEffect()

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellScorch

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() && sim.Proc(procChance, "Improved Scorch") {
			aura := mage.ImprovedScorchAuras.Get(target)
			aura.Activate(sim)
			aura.AddStack(sim)
		}
	}

	config.RelatedAuraArrays = mage.ImprovedScorchAuras.ToMap()

	mage.RegisterSpell(config)
}

// Fire Vulnerability (22959), Improved Scorch's debuff: "Increases Fire damage taken from the Mage by
// 3% per stack". It scales only the Fire damage of the mage whose Scorch applied it: each mage keeps
// its own stack on the target.
func (mage *Mage) newFireVulnerabilityAuras() core.AuraArray {
	debuff := spellData.ImprovedScorchTriggered.Highest()

	return mage.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.GetOrRegisterAura(spelldata.AuraConfig(debuff, spelldata.Label(debuff.Name+mage.Label)))
		spelldata.ParseEffects(nil, aura, debuff, spelldata.FromCaster(&mage.Unit, DDBC_FireVulnerability, DDBC_Total))
		return aura
	})
}

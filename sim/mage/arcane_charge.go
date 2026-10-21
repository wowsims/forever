package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Arcane Charges, the buff Arcane Blast stacks (400573): "you gain 15% of your Intellect as Spell
// Damage and the mana cost of Arcane Blast is increased by 175%". Up to 4 stacks, lasting 15 sec
// from the first; nothing spends them.
var arcaneChargesRank = spellData.ArcaneBlastTriggered.Highest()

// The spell damage per charge, as a share of Intellect (effect 1, every magic school).
var arcaneChargeIntellectShare = arcaneChargesRank.Effect(dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT, 126).Percent()

func (mage *Mage) registerArcaneCharges() {
	// The parse skips effect 1 on a stacking aura, since a stat dependency carries no per-stack
	// value, so the dependency here takes the stacks' share itself.
	spellDamageDep := mage.NewDynamicStatDependency(stats.Intellect, stats.SpellDamage, 0)

	config := spelldata.AuraConfig(arcaneChargesRank)

	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		mage.EnableDynamicStatDep(sim, spellDamageDep)
	}

	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		mage.DisableDynamicStatDep(sim, spellDamageDep)
	}

	config.OnStacksChange = func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
		mage.UpdateDynamicStatDep(sim, spellDamageDep, arcaneChargeIntellectShare*float64(newStacks))
	}

	mage.ArcaneChargesAura = core.BlockPrepull(mage.RegisterAura(config))

	// Arcane Blast's cost, per stack, on the row's own mask.
	spelldata.ParseEffects(&mage.Character, mage.ArcaneChargesAura, arcaneChargesRank, spelldata.Effects(2))
}

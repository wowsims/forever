package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var lightningBoltRanks = castableRanks(spellData.LightningBolt)

func (shaman *Shaman) registerLightningBoltSpell() {
	shaman.registerWithOverloads(lightningBoltRanks, lightningBoltOverloadRanks, shaman.newLightningBoltSpellConfig)
}

// The cast, with the overload it may trigger; the overload itself, from its own row, where overload is nil.
func (shaman *Shaman) newLightningBoltSpellConfig(rank *spelldata.Spell, overload *core.Spell) core.SpellConfig {
	config := shaman.newElectricSpellConfig(rank, overload == nil)
	config.ClassSpellMask = SpellMaskLightningBolt
	if overload == nil {
		config.ClassSpellMask = SpellMaskLightningBoltOverload
	}

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealCastResult(sim, result)

		// The bolt lands before its overload is priced, so a debuff charge the bolt spends does
		// not also raise the overload. DealDamage disposes the result, so Landed is read first.
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			landed := result.Landed()
			spell.DealDamage(sim, result)
			if overload != nil && landed && sim.Proc(shaman.GetOverloadChance(), "Lightning Overload") {
				overload.Cast(sim, target)
			}
		})
	}

	return config
}

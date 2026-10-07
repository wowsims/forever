package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (shaman *Shaman) registerLightningBoltSpell() {
	spellData.LightningBolt.Each(func(_ int32, rank *spelldata.Spell) {
		overload := shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(rank, nil))
		shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(rank, overload))
	})
}

// The cast, with the overload it may trigger; the overload itself, where overload is nil.
func (shaman *Shaman) newLightningBoltSpellConfig(rank *spelldata.Spell, overload *core.Spell) core.SpellConfig {
	config := shaman.newElectricSpellConfig(rank, overload == nil)
	config.ClassSpellMask = SpellMaskLightningBolt
	if overload == nil {
		config.ClassSpellMask = SpellMaskLightningBoltOverload
	}

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if overload != nil && result.Landed() && sim.Proc(shaman.GetOverloadChance(), "Lightning Overload") {
				overload.Cast(sim, target)
			}
			spell.DealDamage(sim, result)
		})
	}

	return config
}

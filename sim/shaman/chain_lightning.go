package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (shaman *Shaman) registerChainLightningSpell() {
	spellData.ChainLightning.Each(func(_ int32, rank *spelldata.Spell) {
		overload := shaman.RegisterSpell(shaman.newChainLightningOverloadConfig(rank))
		shaman.RegisterSpell(shaman.newChainLightningSpellConfig(rank, overload))
	})
}

func (shaman *Shaman) newChainLightningSpellConfig(rank *spelldata.Spell, overload *core.Spell) core.SpellConfig {
	config := shaman.newElectricSpellConfig(rank, false)
	config.ClassSpellMask = SpellMaskChainLightning

	// The effect states the targets the bolt reaches and what each jump keeps.
	damage := rank.DamageEffect()
	maxHits := int32(damage.ChainTargets)
	bounce := float64(damage.ChainAmp)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		numHits := min(maxHits, shaman.Env.ActiveTargetCount())
		results := make([]*core.SpellResult, numHits)

		// Every hit is rolled before any lands, so a proc off the first target cannot raise the
		// damage of the next.
		curTarget := target
		for hitIndex := range numHits {
			results[hitIndex] = spell.CalcDamage(sim, curTarget, damage.Average(), spell.OutcomeMagicHitAndCrit)
			curTarget = sim.Environment.NextActiveTargetUnit(curTarget)
			spell.DamageMultiplier *= bounce
		}

		// Each target hit rolls its own overload, at a third of Lightning Bolt's chance, and the
		// overload strikes that target alone.
		//
		// TODO: Ingame test needed: the tooltip of 408438 states one chance for both spells.
		for _, result := range results {
			if result.Landed() && sim.Proc(shaman.GetOverloadChance()/3, "Lightning Overload") {
				overload.Cast(sim, result.Target)
			}
			spell.DealDamage(sim, result)
			spell.DamageMultiplier /= bounce
		}
	}

	return config
}

func (shaman *Shaman) newChainLightningOverloadConfig(rank *spelldata.Spell) core.SpellConfig {
	config := shaman.newElectricSpellConfig(rank, true)
	config.ClassSpellMask = SpellMaskChainLightningOverload

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}

	return config
}

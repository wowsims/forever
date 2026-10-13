package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var chainLightningRanks = castableRanks(spellData.ChainLightning)

func (shaman *Shaman) registerChainLightningSpell() {
	shaman.registerWithOverloads(chainLightningRanks, chainLightningOverloadRanks, shaman.newChainLightningSpellConfig)
}

// The cast, with the overload it may trigger; the overload itself, from its own row, where overload is
// nil. Both jump alike, so an overload reaches as many targets as the cast, starting on the target that
// triggered it.
func (shaman *Shaman) newChainLightningSpellConfig(rank *spelldata.Spell, overload *core.Spell) core.SpellConfig {
	config := shaman.newElectricSpellConfig(rank, overload == nil)
	config.ClassSpellMask = SpellMaskChainLightning
	if overload == nil {
		config.ClassSpellMask = SpellMaskChainLightningOverload
	}

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

		// Each target the cast hits rolls its own overload, at a third of Lightning Bolt's chance. An
		// overload rolls none of its own. DealDamage disposes the result, so Landed is read first.
		//
		// TODO: Ingame test needed: the tooltip of 408438 states one chance for both spells.
		var overloadTargets []*core.Unit
		for _, result := range results {
			landed, hitTarget := result.Landed(), result.Target
			spell.DealDamage(sim, result)
			spell.DamageMultiplier /= bounce
			if overload != nil && landed && sim.Proc(shaman.GetOverloadChance()/3, "Lightning Overload") {
				overloadTargets = append(overloadTargets, hitTarget)
			}
		}

		// The overloads are priced once every hit of the cast has landed, so a debuff charge one of
		// those hits spends, such as Stormstrike's, does not also raise an overload that jumps there.
		for _, overloadTarget := range overloadTargets {
			overload.Cast(sim, overloadTarget)
		}
	}

	return config
}

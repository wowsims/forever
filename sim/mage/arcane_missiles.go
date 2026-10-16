package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.ArcaneMissiles lists.
func (mage *Mage) registerArcaneMissilesSpell() {
	mage.registerArcaneMissilesRank(spellData.ArcaneMissiles.Len(), spellData.ArcaneMissiles.Highest())
}

// The channel's periodic effect names the missile it fires each tick, a rank of
// spellData.ArcaneMissilesTriggered.
func (mage *Mage) registerArcaneMissilesRank(_ int32, rank *spelldata.Spell) {
	missile := spelldata.MustFind(rank.PeriodicEffect().TriggerID)
	missileDamage := missile.DamageEffect()

	tickConfig := spelldata.SpellConfig(&mage.Unit, missile, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL))
	tickConfig.ClassSpellMask = MageSpellArcaneMissilesTick
	// Each missile rolls when the tick fires it and lands after flying at the row's speed.
	tickConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, missileDamage.Average(), spell.OutcomeTickMagicHitAndCrit)
		spell.DealDamageAfterTravel(sim, result)
		spell.SpellMetrics[target.UnitIndex].Casts--
	}

	tickSpell := mage.RegisterSpell(tickConfig)

	spellConfig := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagHelpful))
	spellConfig.ClassSpellMask = MageSpellArcaneMissilesCast

	dotConfig := spelldata.DotConfig(rank, rank.PeriodicEffect())
	dotConfig.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		tickSpell.Cast(sim, target)
	}

	spellConfig.Dot = dotConfig

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The channel never misses; each missile rolls its own hit.
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)

		spell.Dot(target).Apply(sim)

		tickSpell.SpellMetrics[target.UnitIndex].Casts++
		tickSpell.SpellMetrics[target.UnitIndex].Hits++
	}

	mage.RegisterSpell(spellConfig)
}

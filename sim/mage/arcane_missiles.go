package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 7 is the last rank the trainer teaches; rank 8 comes from the Tome of Arcane Missiles VIII book.
var arcaneMissilesTrainerRank = spellData.ArcaneMissiles.ByID(10212)
var arcaneMissilesBookRank = spellData.ArcaneMissiles.ByID(25345)

func (mage *Mage) registerArcaneMissilesSpell() {
	mage.registerArcaneMissilesRank(arcaneMissilesTrainerRank)
	mage.registerArcaneMissilesRank(arcaneMissilesBookRank)
}

// The channel's periodic effect names the missile it fires each tick, a rank of
// spellData.ArcaneMissilesTriggered.
func (mage *Mage) registerArcaneMissilesRank(rank *spelldata.Spell) {
	missile := spelldata.MustFind(rank.PeriodicEffect().TriggerID)
	missileDamage := missile.DamageEffect()

	tickConfig := spelldata.SpellConfig(&mage.Unit, missile, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL))
	tickConfig.ClassSpellMask = MageSpellArcaneMissilesTick
	// Each missile rolls when the tick fires it and lands after flying at the row's speed.
	tickConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, missileDamage.Average(), spell.OutcomeMagicHitAndCrit)
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
	}

	mage.RegisterSpell(spellConfig)
}

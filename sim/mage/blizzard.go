package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.Blizzard lists.
func (mage *Mage) registerBlizzardSpell() {
	mage.registerBlizzardRank(spellData.Blizzard.Len(), spellData.Blizzard.Highest())
}

// Each tick casts the same rank of spellData.BlizzardTriggered on every enemy in the area, and rolls
// a crit unless the tick's row bars it. The cast's row names no tick spell, only an area trigger, so
// the two pair by rank.
func (mage *Mage) registerBlizzardRank(rank int32, castRank *spelldata.Spell) {
	tickRank := spellData.BlizzardTriggered.Rank(rank)
	cannotCrit := tickRank.CannotCrit()

	tickConfig := spelldata.SpellConfig(&mage.Unit, tickRank, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	if tickRank.IsAProc() {
		tickConfig.Flags |= core.SpellFlagProc
	}
	tickConfig.ClassSpellMask = MageSpellBlizzard
	tickConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		outcome := spell.OutcomeMagicHitAndCrit
		if cannotCrit {
			outcome = spell.OutcomeMagicHit
		}
		spell.CalcAndDealAoeDamageWithVariance(sim, outcome, rollDamage(tickRank.DamageEffect()))
	}
	tickSpell := mage.RegisterSpell(tickConfig)

	// The cast's dummy (effect 1) lands on every enemy in the area as the channel starts: a hit with no
	// damage, which a listener on the enemy hears and a listener that needs damage does not. It reports
	// under the cast's id with tag 1 and records no metrics or log lines. Arcane Concentration rolls at
	// the cast itself, in registerArcaneConcentration (EV-0116).
	landing := mage.RegisterSpell(spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Proc(), spelldata.Tag(1), spelldata.Flags(core.SpellFlagNoMetrics|core.SpellFlagNoLogs)))

	config := spelldata.SpellConfig(&mage.Unit, castRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellBlizzard
	config.Dot = spelldata.DotConfig(castRank, castRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0))
	config.Dot.IsAOE = true
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
		tickSpell.Cast(sim, target)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			landing.CalcAndDealOutcome(sim, target, landing.OutcomeAlwaysHit)
		}
		spell.AOEDot().Apply(sim)
	}

	mage.RegisterSpell(config)
}

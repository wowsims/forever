package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A periodic leech like Drain Life's: each tick heals the caster for what it deals. The row does not mark
// Periodic Can Crit, so the ticks do not crit.
func (warlock *Warlock) registerSiphonLifeSpell() {
	rankedSpell := spellData.SiphonLife.ByID(18881)
	siphonLifeTick := rankedSpell.PeriodicEffect()

	healthMetrics := warlock.NewHealthMetrics(core.ActionID{SpellID: rankedSpell.ID})

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellSiphonLife

	// Tick the row's average, rolling only the crit the row allows, and heal for the damage dealt
	config.Dot = spelldata.DotConfig(rankedSpell, siphonLifeTick)
	config.Dot.Aura.Tag = "Affliction"
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, siphonLifeTick.Average(), rankedSpell.TickOutcome(dot))

		heal := result.Damage * warlock.PseudoStats.SelfHealingMultiplier * warlock.PseudoStats.HealingTakenMultiplier
		warlock.GainHealth(sim, heal, healthMetrics)
	}

	// Roll to hit and apply as dot
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage from the row's average tick
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcPeriodicDamage(sim, target, siphonLifeTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	warlock.SiphonLife = warlock.RegisterSpell(config)
}

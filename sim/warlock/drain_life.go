package warlock

import (
	"github.com/wowsims/forever/sim/core"
)

func (warlock *Warlock) registerDrainLife() {
	rankedSpell := spellData.DrainLife.ByID(11700)
	drainLifeTick := rankedSpell.PeriodicEffect()

	healthMetrics := warlock.NewHealthMetrics(core.ActionID{SpellID: rankedSpell.ID})

	// Heal for the damage each tick dealt
	healOnTick := func(sim *core.Simulation, result *core.SpellResult) {
		heal := result.Damage * warlock.PseudoStats.SelfHealingMultiplier * warlock.PseudoStats.HealingTakenMultiplier
		warlock.GainHealth(sim, heal, healthMetrics)
	}

	config := warlock.soulSiphonChannelConfig(rankedSpell, drainLifeTick, healOnTick)
	config.ClassSpellMask = WarlockSpellDrainLife

	warlock.DrainLife = warlock.RegisterSpell(config)
}

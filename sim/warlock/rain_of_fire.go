package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A channel over an area: every 2 sec its periodic dummy (effect 3) fires the wave (1282385) at every enemy
// in it, each wave rolling hit and crit on each target. Both rows carry Rain of Fire's class flags, so the
// talents naming it reach the waves.
func (warlock *Warlock) registerRainOfFire() {
	rankedSpell := spellData.RainOfFire.ByID(11678)
	channelTick := rankedSpell.Effect(dbcenums.A_PERIODIC_DUMMY, 0)

	waveRow := rankedSpell.Refs()[0]
	waveHit := waveRow.DamageEffect()

	// Fired by the channel; hits every enemy in the area
	waveConfig := spelldata.SpellConfig(&warlock.Unit, waveRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	if waveRow.IsAProc() {
		waveConfig.Flags |= core.SpellFlagProc
	}
	waveConfig.ClassSpellMask = WarlockSpellRainOfFire
	waveConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealDamage(sim, target, waveHit.Average(), spell.OutcomeMagicHitAndCrit)
		}
	}
	wave := warlock.RegisterSpell(waveConfig)

	// The cast's dummy (effect 1) lands on every enemy in the area as the channel starts: a hit with no
	// damage, which a listener on the enemy hears and a listener that needs damage does not, so
	// Judgement of Wisdom on each enemy there rolls once. It reports under the cast's id with tag 1 and
	// records no metrics or log lines (EV-0116).
	landing := warlock.RegisterSpell(spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage),
		spelldata.Proc(), spelldata.Tag(1), spelldata.Flags(core.SpellFlagNoMetrics|core.SpellFlagNoLogs)))

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellRainOfFire

	config.Dot = spelldata.DotConfig(rankedSpell, channelTick)
	config.Dot.IsAOE = true
	config.Dot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
		wave.Cast(sim, warlock.CurrentTarget)
	}

	// Channel over the area
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			landing.CalcAndDealOutcome(sim, target, landing.OutcomeAlwaysHit)
		}
		spell.AOEDot().Apply(sim)
	}

	warlock.RainOfFire = warlock.RegisterSpell(config)
}

package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// A channel on the caster: each second it fires Hellfire Effect at every enemy nearby, and burns the caster
// for the same, spell power included. An area effect that ticks rolls hit on every tick, and crit unless
// Hellfire Effect's row bars it.
func (warlock *Warlock) registerHellfire() {
	rankedSpell := spellData.Hellfire.ByID(11684)
	channelTick := rankedSpell.PeriodicEffect()
	selfBurn := rankedSpell.EffectN(2)

	effectRow := rankedSpell.EffectN(1).Trigger()
	effectHit := effectRow.DamageEffect()
	cannotCrit := effectRow.CannotCrit()

	// Fired by the channel; hits every enemy in range
	effectConfig := spelldata.SpellConfig(&warlock.Unit, effectRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	effectConfig.ClassSpellMask = WarlockSpellHellfire
	effectConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		outcome := spell.OutcomeMagicHitAndCrit
		if cannotCrit {
			outcome = spell.OutcomeMagicHit
		}
		for _, target := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealDamage(sim, target, effectHit.Average(), outcome)
		}
	}
	hellfireEffect := warlock.RegisterSpell(effectConfig)

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellHellfire

	// TODO: The burn on the caster is taken as health lost, not as Fire damage its resistance could reduce.
	config.Dot = spelldata.DotConfig(rankedSpell, channelTick)
	config.Dot.SelfOnly = true
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		hellfireEffect.Cast(sim, target)
		spellPower := warlock.GetStat(stats.SpellDamage) + warlock.GetStat(stats.FireDamage)
		warlock.RemoveHealth(sim, selfBurn.Average()+selfBurn.Coeff()*spellPower)
	}

	// Channel on the caster
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}

	warlock.Hellfire = warlock.RegisterSpell(config)
}

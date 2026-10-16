package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerCorruption() {
	rankedSpell := spellData.Corruption.ByID(25311)
	corruptionTick := rankedSpell.PeriodicEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellCorruption

	// Tick the row's average, rolling only the crit: the hit was rolled when the dot landed
	config.Dot = spelldata.DotConfig(rankedSpell, corruptionTick)
	config.Dot.Aura.Tag = "Affliction"
	config.Dot.OnTick = spelldata.PeriodicDamageTick(corruptionTick, rankedSpell.TickOutcome)

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
		return spell.CalcPeriodicDamage(sim, target, corruptionTick.Average(), spell.OutcomeExpectedMagicAlwaysHit)
	}

	warlock.Corruption = warlock.RegisterSpell(config)
}

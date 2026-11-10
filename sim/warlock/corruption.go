package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 6 is the last rank the trainer teaches; rank 7 comes from the Grimoire of Corruption VII book.
func (warlock *Warlock) registerCorruption() {
	warlock.Corruption = warlock.registerCorruptionRank(spellData.Corruption.ByID(11672))
	warlock.registerCorruptionRank(spellData.Corruption.ByID(25311))
}

func (warlock *Warlock) registerCorruptionRank(rankedSpell *spelldata.Spell) *core.Spell {
	corruptionTick := rankedSpell.PeriodicEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellCorruption

	// Tick the row's average, rolling only the crit: the hit was rolled when the dot landed
	config.Dot = spelldata.DotConfig(rankedSpell, corruptionTick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcAndDealPeriodicDamage(sim, target, corruptionTick.Roll(sim), rankedSpell.TickOutcome(dot))
		if share := dot.SpedUpShare(sim); share > 0 {
			recordWrackContribution(target, dot, result.Damage, share)
		}
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
		return spell.CalcPeriodicDamage(sim, target, corruptionTick.Average(), rankedSpell.ExpectedMagicTickOutcome(spell))
	}

	spell := warlock.RegisterSpell(config)
	warlock.afflictionDots = append(warlock.afflictionDots, spell)
	return spell
}

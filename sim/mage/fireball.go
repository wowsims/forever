package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 11 is the last rank the trainer teaches; rank 12 comes from the Tome of Fireball XII book.
var fireballTrainerRank = spellData.Fireball.ByID(10151)
var fireballBookRank = spellData.Fireball.ByID(25306)

func (mage *Mage) registerFireballSpell() {
	mage.registerFireballRank(fireballTrainerRank)
	mage.registerFireballRank(fireballBookRank)
}

func (mage *Mage) registerFireballRank(rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	// The DoT is its own spell (MageSpellFireballDot). Its ranks share one rank group, so a mage keeps one
	// burn on a target and the higher rank holds it.
	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFireballDot
	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	dotConfig.Dot.OnTick = spelldata.PeriodicDamageTick(tick, rank.TickOutcome)
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	fireballDot := mage.RegisterSpell(dotConfig)

	spellConfig := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	spellConfig.ClassSpellMask = MageSpellFireball

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealCastResult(sim, result)
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				fireballDot.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(spellConfig).RelatedDotSpell = fireballDot
}

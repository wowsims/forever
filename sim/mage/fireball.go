package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered: the mage casts the rank configured here, not every rank
// spellData.Fireball lists.
func (mage *Mage) registerFireballSpell() {
	//Rank 11 is last trainer taught spell
	mage.registerFireballRank(spellData.Fireball.Len(), spellData.Fireball.Rank(11))
}

func (mage *Mage) registerFireballRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	// The DoT is its own spell (MageSpellFireballDot).
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

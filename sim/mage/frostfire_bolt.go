package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Only the highest rank is registered.
func (mage *Mage) registerFrostfireBoltSpell() {
	mage.registerFrostfireBoltRank(spellData.FrostfireBolt.Len(), spellData.FrostfireBolt.Highest())
}

// Tick resistance in EV-0092.
// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerFrostfireBoltRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1),
		spelldata.Flags(core.SpellFlagFullResistanceTicks))
	dotConfig.ClassSpellMask = MageSpellFrostfireBoltDot
	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	dotConfig.Dot.OnTick = spelldata.PeriodicDamageTick(tick, rank.TickOutcome)
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFrostfireBolt

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		spell.DealCastResult(sim, result)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				dotSpell.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(config).RelatedDotSpell = dotSpell
}

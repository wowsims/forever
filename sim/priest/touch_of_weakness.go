package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerTouchOfWeakness(_ int32, rank *spelldata.Spell) {
	// The buff's trigger is a dummy; the tooltip names the rank's strike, which deals the damage and
	// weakens the attacker.
	strikeRank := rank.Refs()[0]
	strikeDamage := strikeRank.DamageEffect()

	weakened := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(spelldata.AuraConfig(strikeRank, spelldata.RankLabel(rank, &priest.Unit)))
		spelldata.ParseEffects(nil, aura, strikeRank, spelldata.Level(priest.Level))
		return aura
	})

	strikeConfig := spelldata.SpellConfig(&priest.Unit, strikeRank,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc())
	strikeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, strikeDamage.Average(), spell.OutcomeMagicHitAndCrit)
		if result.Landed() {
			weakened.Get(target).Activate(sim)
		}
	}
	strikeConfig.RelatedAuraArrays = weakened.ToMap()
	strike := priest.RegisterSpell(strikeConfig)

	aura := priest.RegisterAura(spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &priest.Unit)))
	aura.NewExclusiveEffect(rank.Name, true, core.ExclusiveEffect{})
	aura.AttachProcTrigger(spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			strike.Cast(sim, spell.Unit)
			aura.Deactivate(sim)
		}))

	priest.registerSelfBuffSpell(rank, aura)
}

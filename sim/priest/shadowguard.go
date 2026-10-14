package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerShadowguard(_ int32, rank *spelldata.Spell) {
	// The buff's trigger is a dummy and its own effect states a larger amount (116 at rank 6) than the
	// tooltip, which prices the strike off the rank's triggered row (96). The tooltip's row is dealt.
	strikeRank := rank.Refs()[0]
	strikeDamage := strikeRank.DamageEffect()

	strikeConfig := spelldata.SpellConfig(&priest.Unit, strikeRank,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc())
	strikeConfig.ThreatMultiplier = 0 // The tooltip: this damage causes no threat.
	strikeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, strikeDamage.Average(), spell.OutcomeMagicHit)
		spell.DealDamageAfterTravel(sim, result)
	}
	strike := priest.RegisterSpell(strikeConfig)

	aura := priest.RegisterAura(spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &priest.Unit)))
	aura.NewExclusiveEffect(rank.Name, true, core.ExclusiveEffect{})
	aura.AttachProcTrigger(spelldata.ProcTrigger(&priest.Character, rank,
		func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			strike.Cast(sim, spell.Unit)
			aura.RemoveStack(sim)
		}))

	priest.registerSelfBuffSpell(rank, aura)
}

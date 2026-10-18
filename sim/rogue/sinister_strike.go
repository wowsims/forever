package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var sinisterStrikeRank = spellData.SinisterStrike.Highest()

func (rogue *Rogue) registerSinisterStrikeSpell() {
	sinisterStrikeBaseDamage := sinisterStrikeRank.DamageEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, sinisterStrikeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellSinisterStrike

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := sinisterStrikeBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(sinisterStrikeRank.ProcEnergizeEffect().Roll(sim)), spell.ComboPointMetrics())

		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.SinisterStrike = rogue.RegisterSpell(config)
}

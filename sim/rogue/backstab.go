package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 8 is the highest trainable. Rank 9 exists from a AQ book; acquisition TBD
var backstabRank = spellData.Backstab.Rank(8)

func (rogue *Rogue) registerBackstabSpell() {
	backstabBaseDamage := backstabRank.DamageEffect().Average()
	comboPointGen := backstabRank.EnergizeEffect().Average()
	config := spelldata.SpellConfig(&rogue.Unit, backstabRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Flags |= SpellFlagBuilder | SpellFlagColdBlooded
	config.ClassSpellMask = RogueSpellBackstab
	config.DamageMultiplier = backstabRank.EffectN(2).Percent()
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !rogue.PseudoStats.InFrontOfTarget && rogue.HasDagger(core.MainHand)
	}

	// The row lacks No Active Defense, so the target can dodge it; from behind it is never parried or blocked.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := backstabBaseDamage +
			spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))

		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if result.Landed() {
			rogue.AddComboPoints(sim, int32(comboPointGen), spell.ComboPointMetrics())
			chanceForBonus := spellData.PuncturingWounds.EffectAt(2).FractionAt(rogue.Talents.PuncturingWounds)
			if sim.Proc(chanceForBonus, "Puncturing Wounds") {
				rogue.AddComboPoints(sim, 1, rogue.puncturingWoundsComboPointMetrics)
			}
		} else {
			spell.IssueRefund(sim)
		}
	}

	rogue.Backstab = rogue.RegisterSpell(config)
}
